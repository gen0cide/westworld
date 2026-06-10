package debughttp

// inspect.go — the Stage-1 Host Tablet read endpoints (docs/_research/
// cradle-ui-redesign.md §5 items 1.2–1.6): /aspirations, /fog, /decisions,
// /where. The cradle proxies them per host at /api/hosts/{name}/debug/<path>,
// exactly like /state and /mind.
//
// All four are concurrent READS in the /state // /mind idiom: every runtime
// accessor used here is internally synchronized (world-mirror RWMutex; fog/
// knowledge/goal-graph/ledger mutexes), so none of them takes evalMu or
// touches the host's turn loop.

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/world"
)

// ---- GET /aspirations -------------------------------------------------------

// handleAspirations serves the host's aspiration portfolio (seed order, with
// done/active/open rollups + the neglect flag) — runtime.AspirationStatus
// verbatim. Empty array when the ladder is dark (no persona / not yet seeded).
func (d *Server) handleAspirations(w http.ResponseWriter, _ *http.Request) {
	asp := d.host.AspirationPortfolio()
	if asp == nil {
		asp = []runtime.AspirationStatus{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"aspirations": asp})
}

// ---- GET /fog ---------------------------------------------------------------

// fogSnapshot is the NUMERIC exploration view — coverage %, frontier compass
// text, POI count, sector understanding. No cell/sector masks (those are
// shelved with the map view).
type fogSnapshot struct {
	Coverage  fogCoverage   `json:"coverage"`
	KnownPOIs int           `json:"known_pois"`
	Frontiers []fogFrontier `json:"frontiers"` // nearest unexplored per quadrant, best first per fixed N,E,S,W
	Here      fogSector     `json:"here"`      // the sector the host stands in
	Adjacent  []fogAdjacent `json:"adjacent"`  // same-floor compass neighbours
}

type fogCoverage struct {
	Frac  float64 `json:"frac"`
	Seen  int     `json:"seen"`
	Total int     `json:"total"`
}

// fogFrontier dir follows the RSC compass mirror: EAST = smaller x (the
// x-axis increases west), north = smaller y. The dir text comes straight from
// runtime.Frontier, which already encodes that convention.
type fogFrontier struct {
	Dir  string `json:"dir"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	Dist int    `json:"dist"` // Chebyshev tiles from the host
}

type fogSector struct {
	Terrain  float64 `json:"terrain"`  // seen walkable sub-cells / walkable sub-cells
	Contents float64 `json:"contents"` // harvested POI subjects / subjects present
}

type fogAdjacent struct {
	Dir      string  `json:"dir"`
	Terrain  float64 `json:"terrain"`
	Contents float64 `json:"contents"`
}

func (d *Server) handleFog(w http.ResponseWriter, _ *http.Request) {
	out := fogSnapshot{Frontiers: []fogFrontier{}, Adjacent: []fogAdjacent{}}
	out.Coverage.Frac, out.Coverage.Seen, out.Coverage.Total = d.host.Coverage()
	out.KnownPOIs = d.host.KnownPOICount()
	if wld := d.host.World(); wld != nil && wld.Self != nil {
		pos := wld.Self.Position()
		for _, f := range d.host.FrontierDirections(pos.X, pos.Y) {
			out.Frontiers = append(out.Frontiers, fogFrontier{Dir: f.Direction, X: f.X, Y: f.Y, Dist: f.Dist})
		}
		out.Here.Terrain, out.Here.Contents = d.host.SectorUnderstanding(pos.X, pos.Y)
		for _, g := range d.host.NeighborSectorUnderstanding(pos.X, pos.Y) {
			out.Adjacent = append(out.Adjacent, fogAdjacent{Dir: g.Direction, Terrain: g.Terrain, Contents: g.Contents})
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// ---- GET /decisions ---------------------------------------------------------

// decisionLine mirrors the persisted runtime decisionRecord wire shape 1:1
// (at/trigger/kind/reasoning/goal): decoded only for the before-filter,
// re-encoded unchanged.
type decisionLine struct {
	At        time.Time `json:"at"`
	Trigger   string    `json:"trigger"`
	Kind      string    `json:"kind"`
	Reasoning string    `json:"reasoning"`
	Goal      string    `json:"goal,omitempty"`
}

const (
	decisionsTailDefault = 200
	decisionsTailMax     = 2000
	// decisionsChunk is the backward-read block size: bounds per-read memory so
	// a multi-MB overnight log is never slurped whole.
	decisionsChunk = 256 * 1024
)

// handleDecisions serves a tail of the host's durable decision stream
// (decisions.jsonl — the records that survive restarts and ring eviction):
// the newest `tail` records strictly OLDER than `before` (the paging cursor;
// pass the oldest `at` of the previous page to walk back), in chronological
// order. Hosts without a durable stream (fresh/ephemeral) yield an empty list.
func (d *Server) handleDecisions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tail := decisionsTailDefault
	if t, err := strconv.Atoi(q.Get("tail")); err == nil && t > 0 {
		tail = min(t, decisionsTailMax)
	}
	var before time.Time
	if s := q.Get("before"); s != "" {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "bad before= (want RFC3339): " + err.Error()})
			return
		}
		before = t
	}
	recs, err := tailDecisions(d.host.DecisionLogPath(), tail, before)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		return
	}
	if recs == nil {
		recs = []decisionLine{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"decisions": recs})
}

// tailDecisions reads the newest n records (strictly older than before, when
// set) from the JSONL file at path WITHOUT slurping it: fixed-size blocks are
// read backwards from EOF and split on newlines — the partial first line of
// each block carries over to the next-earlier read — until n records match or
// the file start is reached. Returned in chronological (file) order. A "" or
// missing path reads as an empty stream; torn/corrupt lines are skipped, never
// fatal (the writer may be mid-append on the live host).
func tailDecisions(path string, n int, before time.Time) ([]decisionLine, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}

	var (
		out   []decisionLine // newest-first while scanning backwards
		carry []byte         // partial head line of the previously-read (later) block
		off   = st.Size()
	)
	for off > 0 && len(out) < n {
		sz := min(off, int64(decisionsChunk))
		off -= sz
		buf := make([]byte, sz, sz+int64(len(carry)))
		if _, err := f.ReadAt(buf, off); err != nil {
			return nil, err
		}
		buf = append(buf, carry...)
		lines := bytes.Split(buf, []byte{'\n'})
		first := 0
		if off > 0 { // lines[0] continues into the not-yet-read earlier block
			carry = append([]byte(nil), lines[0]...)
			first = 1
			if len(carry) > 4*decisionsChunk {
				// A newline-less run this long is not a decision record
				// (corruption / a runaway writer): without a cap the carry
				// grows by a block per iteration toward the whole file.
				// Drop it — the severed remainder fails Unmarshal and is
				// skipped, the same never-fatal policy as torn lines.
				carry = nil
			}
		}
		for i := len(lines) - 1; i >= first && len(out) < n; i-- {
			line := bytes.TrimSpace(lines[i])
			if len(line) == 0 {
				continue
			}
			var rec decisionLine
			if json.Unmarshal(line, &rec) != nil {
				continue
			}
			if !before.IsZero() && !rec.At.Before(before) {
				continue
			}
			out = append(out, rec)
		}
	}
	slices.Reverse(out)
	return out, nil
}

// ---- GET /where -------------------------------------------------------------

// whereSnapshot is position-as-text: the gazetteer narration that replaces a
// map view, plus the floor band (y/944 — floor>0 is the boxed-upstairs triage
// signal) and raw coords.
type whereSnapshot struct {
	Text         string `json:"text"`
	Floor        int    `json:"floor"`
	X            int    `json:"x"`
	Y            int    `json:"y"`
	Surroundings string `json:"surroundings,omitempty"` // DescribeSurroundings: npcs/players/ground items/scenery in view
}

func (d *Server) handleWhere(w http.ResponseWriter, _ *http.Request) {
	var out whereSnapshot
	if wld := d.host.World(); wld != nil && wld.Self != nil {
		pos := wld.Self.Position()
		out.X, out.Y = pos.X, pos.Y
		out.Floor = world.PlaneOf(pos.Y)
		out.Text = d.host.LocationSummary()
		out.Surroundings = d.host.DescribeSurroundings(0) // 0 => the runtime's default radius
	}
	writeJSON(w, http.StatusOK, out)
}
