package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/world"
	"github.com/gen0cide/westworld/worldmap"
)

// fog.go is the host's FOG-OF-WAR engine — the anti-omniscience layer that
// makes a host KNOW how much of the world it has not seen, converts exploration
// into learned knowledge (with provenance), and gives the explore/forage drives
// a frontier to aim at. Three pieces:
//
//  1. VISITED-SECTOR RECORDING: every own-position update marks the 48x48
//     landscape sector containing the host as SEEN (the same sector unit the
//     client's pathfinder loads around the player — pathfind.SectorForWorld).
//     The set persists through the memory spine under "fog:sectors" (local
//     hostkv now, mesa write-back later) and warm-starts on boot, exactly like
//     the knowledge ledger (runtime/knowledge.go).
//
//  2. FOG METRICS: Coverage() — the fraction of the REACHABLE world seen,
//     where the denominator is the set of sectors containing at least one
//     standable tile per the shared worldmap.Oracle (ocean/void sectors never
//     count) — and FrontierDirections(x,y), the nearest unexplored-but-
//     reachable sector per compass quadrant. Read by the situation builder.
//
//  3. PERCEPTION HARVEST: on FIRST entry into a sector (once per sector per
//     session) the load-bearing POIs in that sector (furnace/anvil/altar/
//     range/bank/ladder/stairs/mining rocks by ore/trees by type/fishing
//     spots/shop counters) are written into the knowledge ledger as
//     ProvObserved beliefs — "knowledge comes from having been there" — plus
//     any ground-item spawns the host actually SEES in the world mirror.
//
// HOST-LIGHT discipline: deterministic, no LLM, no blocking I/O on the event
// path; the only disk write is the cadenced flush on the limbic spine. The
// recording hook runs on the limbic goroutine (perceptionHandle); the metric
// readers run on director/conductor goroutines, so fogState is mutex-guarded.

// fogSectorsKey is the durable home of the visited-sector set. The "fog"
// namespace rides the memory layer's default class (local write-back), same
// spine as "knowledge:"/"goalgraph:". The value is a JSON array of sorted
// sector ids (fogSectorID encoding) — ~1.7k sectors max, trivially small.
const fogSectorsKey = "fog:sectors"

// Sub-sector understanding (operator ask: touching a sector is not exploring
// it). Each 48x48 sector quantizes into 6x6 sub-cells of 8x8 tiles; a cell
// counts as SEEN when the host's view radius has covered its centre. The
// 36-bit mask per visited sector persists under fog:cells; terrain
// understanding of a sector = seen walkable cells / walkable cells, and
// world Coverage() weights each visited sector by that fraction instead of
// counting mere touches.
const (
	fogCellsKey   = "fog:cells"
	fogCellSize   = 8  // tiles per sub-cell edge (48/8 = 6x6 grid)
	fogCellGrid   = 6  // cells per sector edge
	fogViewRadius = 14 // a cell centre within this many tiles of the host is seen
)

const (
	// fogSectorSize is the landscape sector side (48 tiles), the fog cell unit.
	fogSectorSize = pathfind.SectorSize
	// fogPlaneH is the band-encoded Y extent per floor (plane = y/944).
	fogPlaneH = world.PlaneHeight
	// fogNumPlanes mirrors worldmap's four floors (0 ground, 1-2 up, 3 dungeon).
	fogNumPlanes = 4
	// fogGridStride is the per-axis sector count of the loaded footprint (21).
	// It is the FIXED stride of the persisted fogSectorID encoding — never
	// derive it from a live oracle, or persisted ids would re-key.
	fogGridStride = worldmap.WorldDim / pathfind.SectorSize

	// fogHarvestCap bounds the static-POI ledger writes per sector entry and
	// fogGroundItemCap the ground-item writes — the no-ledger-spam throttle.
	fogHarvestCap     = 16
	fogGroundItemCap  = 6
	fogCompSnapRadius = 6 // mirror of worldmap.CompNear's maxSnap
)

// fogSectorID encodes (plane, sgx, sgy) — plane 0..3, sector grid coords
// 0..20 — into a stable small int for the persisted set.
func fogSectorID(plane, sgx, sgy int) int {
	return (plane*fogGridStride+sgy)*fogGridStride + sgx
}

// fogSectorAt returns the sector id containing band-encoded world (x, y),
// or -1 when the position is outside the loaded footprint.
func fogSectorAt(x, y int) int {
	if x < 0 || y < 0 {
		return -1
	}
	plane := y / fogPlaneH
	sgx := x / fogSectorSize
	sgy := (y % fogPlaneH) / fogSectorSize
	if plane >= fogNumPlanes || sgx >= fogGridStride || sgy >= fogGridStride {
		return -1
	}
	return fogSectorID(plane, sgx, sgy)
}

// fogSectorPlane recovers the plane from a sector id.
func fogSectorPlane(id int) int { return id / (fogGridStride * fogGridStride) }

// fogState is the per-host fog-of-war state. visited is the durable seen-set
// (persisted under fogSectorsKey); harvested is the SESSION-local first-entry
// harvest throttle (deliberately not persisted: a restarted host re-harvests a
// sector once, refreshing any ledger entries the cap pruned — Note dedups
// identical claims, so this never spams). oracle is the test seam over the
// shared worldmap.Oracle (nil ⇒ h.worldOracle).
type fogState struct {
	mu        sync.Mutex
	visited   map[int]bool
	cells     map[int]uint64 // sector id -> 36-bit seen-sub-cell mask
	harvested map[int]bool
	dirty     bool      // visited/cells changed since the last flush
	oracle    fogOracle // test override; production reads h.worldOracle
}

func newFogState() *fogState {
	return &fogState{visited: map[int]bool{}, cells: map[int]uint64{}, harvested: map[int]bool{}}
}

// fogOracle is the slice of worldmap.Oracle the fog engine reads — an
// interface seam so the coverage/frontier math is testable against a fake
// grid. *worldmap.Oracle satisfies it.
type fogOracle interface {
	Dim() int
	DimY() int
	CompAt(x, y int) int32
}

// fogOracleRef resolves the oracle the fog engine should read: the test
// override when set, else the host's shared worldmap oracle, else nil.
func (h *Host) fogOracleRef() fogOracle {
	if h.fog != nil && h.fog.oracle != nil {
		return h.fog.oracle
	}
	if h.worldOracle != nil {
		return h.worldOracle
	}
	return nil
}

// --- recording + harvest (limbic goroutine, via perceptionHandle) -----------

// fogObservePosition records the sector containing (x, y) as SEEN and, on the
// first entry this session, harvests the sector's load-bearing POIs into the
// knowledge ledger. Called from perceptionHandle on every OwnPositionUpdate —
// the seam where the host authoritatively learns where it is (and therefore
// which landscape sector the client has loaded around it). O(1) on the warm
// path (two map probes); the harvest scan runs once per sector per session.
func (h *Host) fogObservePosition(x, y int) {
	if h.fog == nil {
		return
	}
	id := fogSectorAt(x, y)
	if id < 0 {
		return
	}
	h.fog.mu.Lock()
	if !h.fog.visited[id] {
		h.fog.visited[id] = true
		h.fog.dirty = true
	}
	// Mark every sub-cell (in this sector AND neighbours) whose centre falls
	// inside the view radius — walking a sector's edge does not light up its
	// far side.
	fy := y % world.PlaneHeight
	plane := y / world.PlaneHeight
	var newCells [][2]int // global cell-grid coords lit for the first time
	for cgy := (fy - fogViewRadius) / fogCellSize; cgy <= (fy+fogViewRadius)/fogCellSize; cgy++ {
		for cgx := (x - fogViewRadius) / fogCellSize; cgx <= (x+fogViewRadius)/fogCellSize; cgx++ {
			if cgx < 0 || cgy < 0 {
				continue
			}
			cx, cy := cgx*fogCellSize+fogCellSize/2, cgy*fogCellSize+fogCellSize/2
			if absInt(cx-x) > fogViewRadius || absInt(cy-fy) > fogViewRadius {
				continue
			}
			sid := fogSectorID(plane, cgx/fogCellGrid, cgy/fogCellGrid)
			if sid < 0 {
				continue
			}
			bit := uint64(1) << uint((cgy%fogCellGrid)*fogCellGrid+(cgx%fogCellGrid))
			if h.fog.cells[sid]&bit == 0 {
				h.fog.cells[sid] |= bit
				h.fog.visited[sid] = true
				h.fog.dirty = true
				newCells = append(newCells, [2]int{cgx, cgy})
			}
		}
	}
	// Ground items are harvested per-sector from the live mirror (they are
	// view-radius perception already); static POIs are harvested per CELL
	// below — knowledge arrives exactly as sight does.
	groundHarvest := !h.fog.harvested[id]
	if groundHarvest {
		h.fog.harvested[id] = true
	}
	h.fog.mu.Unlock()
	for _, c := range newCells {
		h.fogHarvestCell(plane, c[0], c[1])
	}
	if groundHarvest {
		h.fogHarvestGroundItems(x, y)
		h.fogHarvestNpcs(x, y)
	}
}

// fogHarvestCell writes ONE newly-seen sub-cell's load-bearing POIs into the
// knowledge ledger with ProvObserved at high confidence. Knowledge arrives at
// the same granularity as sight: lighting a sub-cell (its centre entered the
// view radius) harvests exactly that 8x8 patch — never the whole sector
// (touching a corner must not "learn" the far side). Subject conventions
// match the ledger (plain lowercase noun + a claim containing it, so
// closeQuestionByObservation can flip a blocking where-is question on sight).
// Deduped per subject per sector via the ledger's own claim dedup; the
// fogHarvestCap discipline is enforced per sector by Note's caps.
func (h *Host) fogHarvestCell(plane, cgx, cgy int) {
	if h.knowledge == nil || h.facts == nil {
		return
	}
	x0 := cgx * fogCellSize
	y0 := plane*fogPlaneH + (cgy%(fogPlaneH/fogCellSize))*fogCellSize
	type poi struct {
		subject string
		x, y    int
	}
	var core, trees []poi
	seen := map[string]bool{}
	for xx := x0; xx < x0+fogCellSize; xx++ {
		for yy := y0; yy < y0+fogCellSize; yy++ {
			for _, pl := range h.facts.At(xx, yy) {
				if pl.Kind != "scenery" {
					continue
				}
				subject, isTree, ok := fogClassifyScenery(h.facts.SceneryDef(pl.DefID))
				if !ok || seen[subject] {
					continue
				}
				seen[subject] = true
				if isTree {
					trees = append(trees, poi{subject, pl.X, pl.Y})
				} else {
					core = append(core, poi{subject, pl.X, pl.Y})
				}
			}
		}
	}
	for _, p := range append(core, trees...) {
		claim := fmt.Sprintf("%s is at (%d,%d)", p.subject, p.x, p.y)
		h.knowledge.Note(p.subject, "location", claim, knowledge.ProvObserved, 0.9)
		h.knowledge.Seen(p.subject, "location")
		h.closeQuestionByObservation(p.subject, claim)
	}
}

// fogHarvestNpcs records the NPCs the host can SEE right now — "a shopkeeper
// works here", "goblins roam this field". The world mirror only holds in-view
// NPCs, so this is genuine sight. NPCs roam, so claims say "around" rather
// than "at", and a sighted shop/bank keeper can close a blocking
// where-to-buy question exactly like a sighted furnace. Once per sector per
// session (roamers re-sighted elsewhere refresh through later sectors).
func (h *Host) fogHarvestNpcs(x, y int) {
	if h.knowledge == nil || h.world == nil || h.facts == nil {
		return
	}
	plane := y / fogPlaneH
	x0 := (x / fogSectorSize) * fogSectorSize
	y0 := plane*fogPlaneH + ((y%fogPlaneH)/fogSectorSize)*fogSectorSize
	npcs := h.world.Npcs.All()
	sort.Slice(npcs, func(i, j int) bool { // map order → deterministic
		if npcs[i].X != npcs[j].X {
			return npcs[i].X < npcs[j].X
		}
		if npcs[i].Y != npcs[j].Y {
			return npcs[i].Y < npcs[j].Y
		}
		return npcs[i].Index < npcs[j].Index
	})
	seen := map[string]bool{}
	wrote := 0
	for _, n := range npcs {
		if n.X < x0 || n.X >= x0+fogSectorSize || n.Y < y0 || n.Y >= y0+fogSectorSize {
			continue
		}
		name := strings.ToLower(h.npcNameByType(n.TypeID))
		if name == "" || seen[name] || wrote >= fogHarvestCap {
			continue
		}
		seen[name] = true
		claim := fmt.Sprintf("%s is around (%d,%d)", name, n.X, n.Y)
		h.knowledge.Note(name, "location", claim, knowledge.ProvObserved, 0.8)
		h.knowledge.Seen(name, "location")
		h.closeQuestionByObservation(name, claim)
		wrote++
	}
}

// fogHarvestGroundItems records the ground-item spawns the host can SEE right
// now in this sector. The world mirror only holds in-view items, so this is
// genuine sight, not the map. Once per sector per session.
func (h *Host) fogHarvestGroundItems(x, y int) {
	if h.knowledge == nil {
		return
	}
	plane := y / fogPlaneH
	x0 := (x / fogSectorSize) * fogSectorSize
	y0 := plane*fogPlaneH + ((y%fogPlaneH)/fogSectorSize)*fogSectorSize
	if h.world != nil {
		items := h.world.GroundItems.All()
		sort.Slice(items, func(i, j int) bool { // map order → deterministic order
			if items[i].X != items[j].X {
				return items[i].X < items[j].X
			}
			if items[i].Y != items[j].Y {
				return items[i].Y < items[j].Y
			}
			return items[i].ItemID < items[j].ItemID
		})
		seen := map[string]bool{}
		wrote := 0
		for _, gi := range items {
			if gi.X < x0 || gi.X >= x0+fogSectorSize || gi.Y < y0 || gi.Y >= y0+fogSectorSize {
				continue
			}
			name := itemName(h.facts, gi.ItemID)
			if name == "" || strings.HasPrefix(name, "item#") {
				continue // unresolved id: no semantic value (mirrors perceiveItemGained)
			}
			subject := strings.ToLower(name)
			if seen[subject] {
				continue
			}
			seen[subject] = true
			if wrote >= fogGroundItemCap {
				break
			}
			claim := fmt.Sprintf("%s found on the ground at (%d,%d)", subject, gi.X, gi.Y)
			h.knowledge.Note(subject, "item", claim, knowledge.ProvObserved, 0.85)
			wrote++
		}
	}
}

// fogClassifyScenery decides whether a scenery def is a load-bearing POI worth
// a ledger entry, and names its subject. Trees are flagged so the per-sector
// cap fills with the scarcer infrastructure first. Matching is by def NAME
// (the facts vocabulary the rest of the mind uses) plus the command verbs for
// the resource families whose names are generic ("Rock"/"fish").
func fogClassifyScenery(def *facts.SceneryDef) (subject string, isTree, ok bool) {
	if def == nil || def.Name == "" {
		return "", false, false
	}
	name := strings.ToLower(strings.TrimSpace(def.Name))
	c1 := strings.ToLower(def.Command1)

	switch {
	case c1 == "mine":
		return fogOreSubject(def, name), false, true
	case c1 == "chop":
		return name, true, true // trees by type ("tree", "dense jungle tree", ...)
	case c1 == "net" || c1 == "bait" || c1 == "lure" || c1 == "harpoon" || c1 == "cage":
		return "fishing spot", false, true
	}
	for _, needle := range [...]string{
		"furnace", "anvil", "altar", "ladder", "stair", "counter", "stall", "bank",
	} {
		if strings.Contains(name, needle) {
			return name, false, true
		}
	}
	// "range" needs the orange-tree exclusion ("Orange Tree" contains "range").
	if strings.Contains(name, "range") && !strings.Contains(name, "orange") {
		return name, false, true
	}
	return "", false, false
}

// fogOreSubject names a mineable rock by its ore, derived from the def's model
// ("copperrock1" → "copper rock") — every mining rock is just "Rock" by name,
// which would be a useless subject. Falls back to the lowercase def name.
func fogOreSubject(def *facts.SceneryDef, lowName string) string {
	model := strings.ToLower(def.Model)
	if i := strings.Index(model, "rock"); i > 0 {
		ore := strings.TrimSpace(model[:i])
		if ore != "" {
			return ore + " rock"
		}
	}
	return lowName
}

// --- persistence (limbic spine: loadFog on start, flushFog on the cadence) --

// loadFog restores the visited-sector set from the durable memory layer.
// No-op when no memory manager is wired (in-RAM-only hosts / tests). Restored
// sectors are NOT marked harvested: a fresh session re-harvests each on first
// re-entry (once), refreshing anything the ledger cap pruned.
// fogCellsBlob is the persisted shape of fog:cells — sector id -> mask, as a
// compact JSON object with string keys (json maps need string keys).
type fogCellsBlob map[string]uint64

func (h *Host) loadFog(ctx context.Context) {
	if h.Memory == nil || h.fog == nil {
		return
	}
	rec, ok, err := h.Memory.Get(ctx, fogSectorsKey)
	if err != nil || !ok {
		return
	}
	var ids []int
	if json.Unmarshal(rec.Value, &ids) != nil || len(ids) == 0 {
		return
	}
	h.fog.mu.Lock()
	for _, id := range ids {
		if id >= 0 {
			h.fog.visited[id] = true
		}
	}
	h.fog.mu.Unlock()
	// Sub-cell masks (absent for pre-upgrade saves: those sectors restore as
	// touched-but-unexplored, which is the honest reading).
	if rec, ok, err := h.Memory.Get(ctx, fogCellsKey); err == nil && ok {
		var blob fogCellsBlob
		if json.Unmarshal(rec.Value, &blob) == nil {
			h.fog.mu.Lock()
			for k, mask := range blob {
				if id, err := strconv.Atoi(k); err == nil && id >= 0 && mask != 0 {
					h.fog.cells[id] = mask
					h.fog.visited[id] = true
				}
			}
			h.fog.mu.Unlock()
		}
	}
	h.log.Info("fog: restored visited sectors", "sectors", len(ids))
}

// flushFog persists the visited-sector set (sorted JSON ids) when dirty.
// Best-effort; frozen under analysis-mode like the other learning flushes;
// no-op when clean or unwired. Runs on the limbic flush cadence + on exit.
func (h *Host) flushFog(ctx context.Context) {
	if h.Memory == nil || h.fog == nil {
		return
	}
	if h.AnalysisActive() {
		return
	}
	h.fog.mu.Lock()
	defer h.fog.mu.Unlock()
	if !h.fog.dirty {
		return
	}
	ids := make([]int, 0, len(h.fog.visited))
	for id := range h.fog.visited {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	raw, err := json.Marshal(ids)
	if err != nil {
		return
	}
	if err := h.Memory.Put(ctx, fogSectorsKey, raw); err != nil {
		h.log.Warn("fog: visited-sector flush failed", "err", err)
		return
	}
	blob := make(fogCellsBlob, len(h.fog.cells))
	for id, mask := range h.fog.cells {
		blob[strconv.Itoa(id)] = mask
	}
	if raw, err := json.Marshal(blob); err == nil {
		if err := h.Memory.Put(ctx, fogCellsKey, raw); err != nil {
			h.log.Warn("fog: cell-mask flush failed", "err", err)
			return
		}
	}
	h.fog.dirty = false
}

// --- the walkable-sector index (shared per oracle, built lazily once) --------

// fogCompRep is one walking component's presence in a sector: how many of the
// sector's tiles it covers and a representative standable tile (the one
// nearest the sector centre — a safe go_to target).
type fogCompRep struct {
	comp  int32
	x, y  int
	tiles int
}

// fogIndex is the static walkable-sector index over one oracle: sector id →
// the components present in it. A sector absent from the map has NO standable
// tiles (ocean/void/unloaded) and never counts toward coverage.
type fogIndex struct {
	sectors  map[int][]fogCompRep
	cellMask map[int]uint64 // sector id -> walkable sub-cell mask
}

var (
	fogIdxMu    sync.Mutex
	fogIdxCache map[fogOracle]*fogIndex
)

// fogIndexFor returns the (lazily built, process-cached) walkable index for an
// oracle. The oracle is shared by pointer across all hosts, so the scan —
// every tile of every sector, once — is paid once per process, not per host.
func fogIndexFor(o fogOracle) *fogIndex {
	if o == nil {
		return nil
	}
	fogIdxMu.Lock()
	defer fogIdxMu.Unlock()
	if idx, ok := fogIdxCache[o]; ok {
		return idx
	}
	idx := buildFogIndex(o)
	if fogIdxCache == nil {
		fogIdxCache = map[fogOracle]*fogIndex{}
	}
	fogIdxCache[o] = idx
	return idx
}

// buildFogIndex scans every sector of every plane band for standable tiles
// (CompAt >= 0) and records the components present + their representative
// tiles. Band-clamped exactly like the oracle's own mask build (a floor band
// is 944 rows; sector rows past it belong to the next floor and are skipped).
func buildFogIndex(o fogOracle) *fogIndex {
	idx := &fogIndex{sectors: map[int][]fogCompRep{}, cellMask: map[int]uint64{}}
	dim, dimY := o.Dim(), o.DimY()
	planes := (dimY + fogPlaneH - 1) / fogPlaneH
	if planes > fogNumPlanes {
		planes = fogNumPlanes
	}
	gw := (dim + fogSectorSize - 1) / fogSectorSize
	if gw > fogGridStride {
		gw = fogGridStride
	}
	for plane := 0; plane < planes; plane++ {
		bandTop := plane * fogPlaneH
		bandRows := dimY - bandTop
		if bandRows > fogPlaneH {
			bandRows = fogPlaneH
		}
		gh := (bandRows + fogSectorSize - 1) / fogSectorSize
		if gh > fogGridStride {
			gh = fogGridStride
		}
		for sgy := 0; sgy < gh; sgy++ {
			for sgx := 0; sgx < gw; sgx++ {
				if comps, mask := fogScanSector(o, plane, sgx, sgy, bandTop, bandRows, dim); len(comps) > 0 {
					idx.cellMask[fogSectorID(plane, sgx, sgy)] = mask
					idx.sectors[fogSectorID(plane, sgx, sgy)] = comps
				}
			}
		}
	}
	return idx
}

// fogScanSector collects the components present in one sector, each with its
// tile count and the standable tile nearest the sector centre. Returns nil
// when the sector holds no standable tile. Deterministic: ties on centre
// distance keep the first tile in (x asc, y asc) scan order; the result is
// sorted by tile count desc, component id asc.
func fogScanSector(o fogOracle, plane, sgx, sgy, bandTop, bandRows, dim int) ([]fogCompRep, uint64) {
	x0 := sgx * fogSectorSize
	y0 := bandTop + sgy*fogSectorSize
	x1 := min(x0+fogSectorSize, dim)
	y1 := min(y0+fogSectorSize, bandTop+bandRows)
	cx, cy := x0+fogSectorSize/2, y0+fogSectorSize/2

	var cellMask uint64 // which 8x8 sub-cells hold >=1 standable tile
	byComp := map[int32]*fogCompRep{}
	for x := x0; x < x1; x++ {
		for y := y0; y < y1; y++ {
			c := o.CompAt(x, y)
			if c < 0 {
				continue
			}
			cellMask |= uint64(1) << uint(((y-y0)/fogCellSize)*fogCellGrid+(x-x0)/fogCellSize)
			r, ok := byComp[c]
			if !ok {
				byComp[c] = &fogCompRep{comp: c, x: x, y: y, tiles: 1}
				continue
			}
			r.tiles++
			if chebyshev(x, y, cx, cy) < chebyshev(r.x, r.y, cx, cy) {
				r.x, r.y = x, y
			}
		}
	}
	if len(byComp) == 0 {
		return nil, 0
	}
	out := make([]fogCompRep, 0, len(byComp))
	for _, r := range byComp {
		out = append(out, *r)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].tiles != out[j].tiles {
			return out[i].tiles > out[j].tiles
		}
		return out[i].comp < out[j].comp
	})
	return out, cellMask
}

// fogCompNear mirrors worldmap.Oracle.CompNear over the fogOracle seam: the
// component at (x,y) when standable, else the nearest standable tile's
// component within the snap radius (never across a floor band). -1 when none.
func fogCompNear(o fogOracle, x, y int) int32 {
	if c := o.CompAt(x, y); c >= 0 {
		return c
	}
	for r := 1; r <= fogCompSnapRadius; r++ {
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				if dx > -r && dx < r && dy > -r && dy < r {
					continue // ring border only
				}
				nx, ny := x+dx, y+dy
				if ny < 0 || ny/fogPlaneH != y/fogPlaneH {
					continue // never snap across a floor band
				}
				if c := o.CompAt(nx, ny); c >= 0 {
					return c
				}
			}
		}
	}
	return -1
}

// --- fog metrics API (read by the situation builder / explore drive) --------

// Coverage reports how much of the REACHABLE world the host has seen: the
// fraction of sectors containing at least one standable tile (per the shared
// worldmap oracle — ocean/void never count) whose sector the host has stood
// in. Returns (0, 0, 0) when no oracle is wired. Thread-safe.
func (h *Host) Coverage() (frac float64, seen, total int) {
	idx := fogIndexFor(h.fogOracleRef())
	if idx == nil || len(idx.sectors) == 0 || h.fog == nil {
		return 0, 0, 0
	}
	o := h.fogOracleRef()
	var weighted float64
	h.fog.mu.Lock()
	for id := range idx.sectors {
		if !h.fog.visited[id] {
			continue
		}
		seen++
		// Weight a visited sector by how much of its WALKABLE ground the
		// host has actually had in view — touching a corner is not
		// exploring (operator ask). Sectors visited before the sub-cell
		// upgrade have no mask and weigh as barely-explored.
		weighted += fogCellFraction(o, id, h.fog.cells[id])
	}
	h.fog.mu.Unlock()
	total = len(idx.sectors)
	return weighted / float64(total), seen, total
}

// SectorUnderstanding reports how well the host knows the sector containing
// (x, y): terrain = seen walkable sub-cells / walkable sub-cells; contents =
// knowledge-harvested POI subjects / POI subjects actually present (from the
// static facts — what a thorough visitor WOULD have catalogued). Either is 0
// for a never-visited sector.
func (h *Host) SectorUnderstanding(x, y int) (terrain, contents float64) {
	if h.fog == nil {
		return 0, 0
	}
	id := fogSectorAt(x, y)
	if id < 0 {
		return 0, 0
	}
	h.fog.mu.Lock()
	mask := h.fog.cells[id]
	visited := h.fog.visited[id]
	h.fog.mu.Unlock()
	if !visited {
		return 0, 0
	}
	terrain = fogCellFraction(h.fogOracleRef(), id, mask)
	present := h.fogPresentSubjects(x, y)
	if present > 0 {
		known := h.fogKnownSubjects(x, y)
		contents = float64(known) / float64(present)
		if contents > 1 {
			contents = 1
		}
	}
	return terrain, contents
}

// fogCellFraction: seen walkable sub-cells over walkable sub-cells for one
// sector, exact from the index's per-sector walkable-cell mask (built in the
// same full-tile scan that finds components — no extra oracle reads).
func fogCellFraction(o fogOracle, id int, mask uint64) float64 {
	idx := fogIndexFor(o)
	if idx == nil {
		return float64(bitsOn(mask)) / float64(fogCellGrid*fogCellGrid)
	}
	walkable := idx.cellMask[id]
	if walkable == 0 {
		return 0
	}
	return float64(bitsOn(mask&walkable)) / float64(bitsOn(walkable))
}

// fogEnumerateSubjects lists the distinct classified POI subjects present in
// the sector containing (x, y), per the static facts — the catalogue a
// thorough visitor would have harvested (same classifier + same
// infrastructure-before-trees priority as fogHarvestSector, same cap).
func (h *Host) fogEnumerateSubjects(x, y int) []string {
	if h.facts == nil {
		return nil
	}
	plane := y / fogPlaneH
	x0 := (x / fogSectorSize) * fogSectorSize
	y0 := plane*fogPlaneH + ((y%fogPlaneH)/fogSectorSize)*fogSectorSize
	var core, trees []string
	seen := map[string]bool{}
	for xx := x0; xx < x0+fogSectorSize; xx++ {
		for yy := y0; yy < y0+fogSectorSize; yy++ {
			for _, p := range h.facts.At(xx, yy) {
				if p.Kind != "scenery" {
					continue
				}
				subject, isTree, ok := fogClassifyScenery(h.facts.SceneryDef(p.DefID))
				if !ok || seen[subject] {
					continue
				}
				seen[subject] = true
				if isTree {
					trees = append(trees, subject)
				} else {
					core = append(core, subject)
				}
			}
		}
	}
	all := append(core, trees...)
	if len(all) > fogHarvestCap {
		all = all[:fogHarvestCap]
	}
	return all
}

func (h *Host) fogPresentSubjects(x, y int) int { return len(h.fogEnumerateSubjects(x, y)) }

// fogKnownSubjects counts how many of the sector's present subjects the
// knowledge ledger actually knows — contents understanding's numerator.
func (h *Host) fogKnownSubjects(x, y int) int {
	if h.knowledge == nil {
		return 0
	}
	n := 0
	for _, subject := range h.fogEnumerateSubjects(x, y) {
		if h.knowledge.Known(subject) {
			n++
		}
	}
	return n
}

func bitsOn(v uint64) int {
	n := 0
	for ; v != 0; v &= v - 1 {
		n++
	}
	return n
}

// Frontier is one unexplored-but-reachable direction summary: the nearest
// never-visited walkable sector in that compass quadrant, as a standable tile
// (a safe go_to target) and its Chebyshev distance from the query point.
type Frontier struct {
	Direction string // "north" | "south" | "east" | "west"
	X, Y      int    // representative standable tile inside the unexplored sector
	Dist      int    // Chebyshev tiles from the query point
}

// FrontierDirections returns, for each compass quadrant, the nearest sector
// the host has NEVER visited that is walk-reachable from (x, y) (same walking
// component — no transport-gate crossing, the ReachOpen analogue) on the same
// floor. Quadrants with no frontier are omitted; order is fixed N, E, S, W.
// RSC compass: north = smaller y, EAST = smaller x (the x-axis increases
// west) — same convention as director_scene.bearingFrom. Thread-safe.
func (h *Host) FrontierDirections(x, y int) []Frontier {
	cands := h.frontierCandidates(x, y)
	best := map[string]Frontier{}
	for _, c := range cands { // sorted nearest-first; first claim per quadrant wins
		dir := fogQuadrant(x, y, c.x, c.y)
		if _, taken := best[dir]; !taken {
			best[dir] = Frontier{Direction: dir, X: c.x, Y: c.y, Dist: c.dist}
		}
	}
	var out []Frontier
	for _, dir := range [...]string{"north", "east", "south", "west"} {
		if f, ok := best[dir]; ok {
			out = append(out, f)
		}
	}
	return out
}

// KnownPOICount reports how many distinct places the host holds at least one
// belief about (knowledge entries of kind "location" with a claim) — the
// "how much have I actually LEARNED from being out there" companion metric to
// Coverage. Cheap: one pass over the capped ledger.
func (h *Host) KnownPOICount() int {
	if h.knowledge == nil {
		return 0
	}
	n := 0
	for _, f := range h.knowledge.All() {
		if f.Kind == "location" && len(f.Beliefs) > 0 {
			n++
		}
	}
	return n
}

// FrontierTarget picks the single nearest unexplored-but-reachable sector as
// an explore/forage destination: a standable tile in the host's own walking
// component plus a stable human-readable label (the per-question source-tried
// tag key). question is reserved for future per-question biasing — the
// current picker is purely spatial (nearest frontier first). ok=false when
// there is no oracle, no position, or no unexplored reachable sector left.
func (h *Host) FrontierTarget(question string) (x, y int, label string, ok bool) {
	_ = question // reserved: per-question biasing (e.g. ore questions → mine-rich quadrants)
	if h.world == nil || h.world.Self == nil {
		return 0, 0, "", false
	}
	pos := h.world.Self.Position()
	cands := h.frontierCandidates(pos.X, pos.Y)
	if len(cands) == 0 {
		return 0, 0, "", false
	}
	c := cands[0]
	return c.x, c.y, fmt.Sprintf("unexplored area near (%d,%d)", c.x, c.y), true
}

// frontierForageSource is the fog fallback for the forage drive's ARM-1
// source selection (forage.go nextForageSource): when every mapped POI source
// is tried/spent, aim the forage at the nearest unexplored frontier sector
// instead — the search EXPANDS THE MAP rather than giving up. The same
// source-tried/source-spent tags gate it, so one question never ping-pongs to
// the same frontier forever (and arrival marks the sector visited anyway,
// moving the frontier).
func (h *Host) frontierForageSource(q goalgraph.Node) (forageSource, bool) {
	x, y, label, ok := h.FrontierTarget(q.Label)
	if !ok || h.goalGraph == nil {
		return forageSource{}, false
	}
	if h.goalGraph.HasTag(q.ID, "source-tried:place:"+label) ||
		h.goalGraph.HasTag(q.ID, "source-spent:place:"+label) {
		return forageSource{}, false
	}
	return forageSource{label: label, x: x, y: y}, true
}

// fogFrontierCand is one unexplored-but-reachable sector candidate.
type fogFrontierCand struct {
	id   int
	x, y int // representative standable tile of the viewer's component
	dist int
}

// frontierCandidates lists every never-visited walkable sector on (x, y)'s
// floor that shares the viewer's walking component, nearest-first (ties by
// sector id — deterministic). Empty when no oracle / not standable nearby.
func (h *Host) frontierCandidates(x, y int) []fogFrontierCand {
	o := h.fogOracleRef()
	idx := fogIndexFor(o)
	if idx == nil || h.fog == nil {
		return nil
	}
	comp := fogCompNear(o, x, y)
	if comp < 0 {
		return nil
	}
	plane := y / fogPlaneH

	h.fog.mu.Lock()
	visited := make(map[int]bool, len(h.fog.visited))
	for id := range h.fog.visited {
		visited[id] = true
	}
	h.fog.mu.Unlock()

	var out []fogFrontierCand
	for id, comps := range idx.sectors {
		if visited[id] || fogSectorPlane(id) != plane {
			continue
		}
		for _, r := range comps {
			if r.comp != comp {
				continue
			}
			out = append(out, fogFrontierCand{id: id, x: r.x, y: r.y, dist: chebyshev(x, y, r.x, r.y)})
			break
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].dist != out[j].dist {
			return out[i].dist < out[j].dist
		}
		return out[i].id < out[j].id
	})
	return out
}

// fogQuadrant assigns a compass quadrant by dominant axis, in RSC coordinates
// (north = smaller y, EAST = smaller x — the x-axis increases west; see
// director_scene.bearingFrom). Vertical wins exact diagonal ties.
func fogQuadrant(px, py, x, y int) string {
	dx, dy := x-px, y-py
	if absInt(dy) >= absInt(dx) {
		if dy < 0 {
			return "north"
		}
		return "south"
	}
	if dx < 0 {
		return "east"
	}
	return "west"
}
