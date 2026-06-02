package main

// remoteclient.go — Layer 3: the full browser remote-client HTTP API.
//
// serveClient is the symmetric counterpart of spectate (spectate.go). It boots
// the same infrastructure (waitForLivePosition, OpenBundle, motionTracker) and
// re-registers the five viewport routes (/frame /pos /walk /shot /clip) exactly
// as spectate does, then adds the four new routes (/pick /act /state /chat) plus
// the convenience /examine endpoint. All of that lives in one self-contained
// function so spectate.go is never touched and -spectate keeps working.
//
// Architecture (from docs/remote-client/00-overview.md §4):
//
//	Browser (SPA, /client HTML)
//	    ↓ HTTP/JSON
//	serveClient (this file) — Layer 3: decode → route → encode
//	    ↓
//	render.Pick — Layer 1: screen hit-testing
//	remoteclient.BuildCandidates / Dispatch — Layer 2: menu + action
//	    ↓
//	runtime.Host — already done: the full RSC wire-action surface
//
// All mutating endpoints (/walk, /act, /chat) funnel through ONE actCh worker
// goroutine so RSC's "latest action wins" semantics hold. Reads (/frame, /pos,
// /state, /pick, /examine) are lock-safe and call world accessors directly from
// the HTTP goroutine.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/remoteclient"
	"github.com/gen0cide/westworld/render"
	"github.com/gen0cide/westworld/runtime"
	"github.com/gen0cide/westworld/web"
	"github.com/gen0cide/westworld/world"
)

// ---- request / response wire types (§6.1 of 50-impl-spec.md) ---------------

// pickRequest is the body of POST /pick. Slot carries the inventory fallback
// path (spec §6.1: PickRequest{...; Slot int}): when Slot>=0 the browser is
// asking for the right-click menu of an inventory slot (not a screen pixel), so
// the handler builds the candidate from world.Inventory + InventoryMenu instead
// of running render.Pick. Slot defaults to 0, so callers that mean "no slot"
// must send -1; the handler only takes the inventory branch when Slot>=0 AND the
// caller did not supply a real screen pixel (W/H present but PX/PY both 0 with a
// Slot is the inventory-fallback signal clientpage.go uses: pick(0,0,{slot})).
type pickRequest struct {
	PX   int `json:"px"`
	PY   int `json:"py"`
	Rot  int `json:"rot"`
	Zoom int `json:"zoom"`
	W    int `json:"w"`
	H    int `json:"h"`
	Slot int `json:"slot"`
}

// actRequest is the body of POST /act.
type actRequest struct {
	Ref      remoteclient.MenuTarget `json:"ref"`
	OptionID int                     `json:"optionId"`
}

// actResponse is the body of POST /act and /chat responses.
type actResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message,omitempty"`
}

// chatRequest is the body of POST /chat.
type chatRequest struct {
	Kind string `json:"kind"` // "say" | "command" | "pm"
	Text string `json:"text"`
	To   string `json:"to"` // required iff kind == "pm"
}

// ---- action worker (§6.4 of 50-impl-spec.md) --------------------------------

// actReq is one queued world action for the single serialised actCh worker.
// coalesce:true → drain-then-send (walk retarget semantics; done must be nil).
// coalesce:false → serialize; done receives the result.
type actReq struct {
	run      func(ctx context.Context) (string, error)
	done     chan actResult // nil = fire-and-forget (walk)
	coalesce bool
}

type actResult struct {
	msg string
	err error
}

// ---- chat ring (§5 / §6 of 50-impl-spec.md) ---------------------------------

// chatKind values match the /state chat[] kind field.
const (
	chatKindPublic  = "public"
	chatKindPrivate = "private"
	chatKindSystem  = "system"
	chatKindNPC     = "npc"
	chatKindSelf    = "self"
)

// chatRingCap is the maximum entries the server-session chat ring holds.
const chatRingCap = 64

// chatEntry is one line in /state's chat[] array.
type chatEntry struct {
	Seq  int    `json:"seq"`
	Kind string `json:"kind"`
	Who  string `json:"who"`
	Text string `json:"text"`
}

// chatRing is the server-session append-only ring of chat lines, grown from
// outgoing /chat sends and de-duped tails of Recent. It lives in serveClient's
// closure (one per server session).
type chatRing struct {
	mu      sync.Mutex
	entries []chatEntry
	seq     int
	// lastChatAt / lastPMAt / lastSysMsgAt track how far into Recent we've
	// consumed so each /state poll appends only genuinely new lines. All three
	// dedupe by TIMESTAMP (not slice index): ServerMessages() is a BOUNDED ring
	// (world.ServerMsgRingCap=32) that drops its oldest entries and copies the
	// tail down once full, so absolute indices shift out from under us. A
	// timestamp cursor is index-shift-immune.
	lastChatAt   time.Time
	lastPMAt     time.Time
	lastSysMsgAt time.Time
}

// add appends a new entry to the ring (dropping the oldest when full) and
// returns the assigned sequence number. Caller must hold no lock (add takes it).
func (r *chatRing) add(kind, who, text string) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	e := chatEntry{Seq: r.seq, Kind: kind, Who: who, Text: text}
	r.entries = append(r.entries, e)
	if len(r.entries) > chatRingCap {
		copy(r.entries, r.entries[len(r.entries)-chatRingCap:])
		r.entries = r.entries[:chatRingCap]
	}
	return r.seq
}

// snapshot returns a copy of all current entries (oldest-first).
func (r *chatRing) snapshot() []chatEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]chatEntry, len(r.entries))
	copy(out, r.entries)
	return out
}

// ingestRecent polls world.Recent and appends any chat/PM/server-messages that
// arrived since the last call. De-duplication is by timestamp comparison so the
// ring never doubles up from rapid /state polls.
func (r *chatRing) ingestRecent(recent *world.RecentEvents) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if c := recent.Chat(); c != nil && c.At.After(r.lastChatAt) {
		r.lastChatAt = c.At
		r.seq++
		r.entries = append(r.entries, chatEntry{
			Seq: r.seq, Kind: chatKindPublic, Who: c.Speaker, Text: c.Message,
		})
	}

	if p := recent.PM(); p != nil && p.At.After(r.lastPMAt) {
		r.lastPMAt = p.At
		r.seq++
		r.entries = append(r.entries, chatEntry{
			Seq: r.seq, Kind: chatKindPrivate, Who: p.Sender, Text: p.Message,
		})
	}

	// server messages: ServerMessages() is a bounded ring whose indices shift as
	// it drops old entries, so we dedupe by timestamp (NOT index). Append only
	// records newer than the last one we appended, advancing the cursor to the
	// newest. The ring is already oldest-first so a single forward pass preserves
	// order; equal timestamps (rare; sub-ms bursts) are skipped on the next poll,
	// matching the Chat/PM single-slot dedupe.
	for _, m := range recent.ServerMessages() {
		if !m.At.After(r.lastSysMsgAt) {
			continue
		}
		r.lastSysMsgAt = m.At
		kind := chatKindSystem
		if m.Kind == "dialog" {
			kind = chatKindNPC
		}
		r.seq++
		r.entries = append(r.entries, chatEntry{
			Seq: r.seq, Kind: kind, Who: "", Text: m.Message,
		})
	}

	// trim
	if len(r.entries) > chatRingCap {
		copy(r.entries, r.entries[len(r.entries)-chatRingCap:])
		r.entries = r.entries[:chatRingCap]
	}
}

// ---- WorldView adapter -------------------------------------------------------

// worldViewAdapter implements remoteclient.WorldView on top of *world.World.
// NpcVisible / PlayerVisible check the live world mirrors; SlotItem reads the
// inventory. All reads are lock-safe (the world accessors RLock internally).
type worldViewAdapter struct {
	w *world.World
}

func (a *worldViewAdapter) NpcVisible(serverIndex int) bool {
	_, ok := a.w.Npcs.Get(serverIndex)
	return ok
}

func (a *worldViewAdapter) PlayerVisible(serverIndex int) bool {
	_, ok := a.w.Players.Get(serverIndex)
	return ok
}

func (a *worldViewAdapter) SlotItem(slot int) (itemID int, wielded bool, ok bool) {
	slots := a.w.Inventory.Slots()
	if slot < 0 || slot >= len(slots) {
		return 0, false, false
	}
	s := slots[slot]
	if s.ItemID == 0 {
		return 0, false, false
	}
	return s.ItemID, s.Wielded, true
}

// ---- /state response types ---------------------------------------------------

type stateSkill struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
	Max   int    `json:"max"`
	XP    int    `json:"xp"`
}

type stateSelf struct {
	X           int          `json:"x"`
	Y           int          `json:"y"`
	Plane       int          `json:"plane"`
	Heading     int          `json:"heading"`
	CombatLevel int          `json:"combatLevel"`
	HP          int          `json:"hp"`
	MaxHP       int          `json:"maxHp"`
	Prayer      int          `json:"prayer"`
	MaxPrayer   int          `json:"maxPrayer"`
	Fatigue     int          `json:"fatigue"`
	QuestPoints int          `json:"questPoints"`
	Skills      []stateSkill `json:"skills"`
}

type stateInvItem struct {
	Slot            int                        `json:"slot"`
	ItemID          int                        `json:"itemId"`
	Name            string                     `json:"name"`
	Amount          int                        `json:"amount"`
	Wielded         bool                       `json:"wielded"`
	Wearable        bool                       `json:"wearable"`
	Stackable       bool                       `json:"stackable"`
	Command         string                     `json:"command"`
	DefaultOptionID int                        `json:"defaultOptionId"`
	Options         []remoteclient.MenuOption  `json:"options"`
}

type stateEquipItem struct {
	Slot   string `json:"slot"`
	Sprite int    `json:"sprite"`
	ItemID int    `json:"itemId"`
}

type stateResponse struct {
	Self      stateSelf        `json:"self"`
	Inventory []stateInvItem   `json:"inventory"`
	Equipment []stateEquipItem `json:"equipment"`
	Chat      []chatEntry      `json:"chat"`
	// Bank is present only while the bank window is open (window lifecycle,
	// 110-react-port §E5). The SPA shows <BankWindow> iff bank != null.
	Bank *stateBank `json:"bank,omitempty"`
}

// stateBank mirrors world.BankRecord for the SPA. Slot is the bank slot index.
type stateBank struct {
	Open    bool            `json:"open"`
	MaxSize int             `json:"maxSize"`
	Slots   []stateBankSlot `json:"slots"`
}

type stateBankSlot struct {
	Slot   int    `json:"slot"`
	ItemID int    `json:"itemId"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

// bankRequest is the body of POST /bank: op deposit|withdraw|close. For deposit
// and withdraw, ItemID is the catalog item id and Amount the quantity.
type bankRequest struct {
	Op     string `json:"op"`
	ItemID int    `json:"itemId"`
	Amount int    `json:"amount"`
}

// ---- serveClient -------------------------------------------------------------

// serveClient serves the full remote-client SPA + JSON API. It mirrors
// spectate()'s bootstrap (waitForLivePosition, OpenBundle, motionTracker,
// action worker, graceful shutdown) but adds the M1 client routes and serves
// clientPage at "/". Reuses the package-main helpers buildLiveView,
// motionTracker, waitForLivePosition, atoiOr. Blocks until ctx is cancelled.
func serveClient(ctx context.Context, log *slog.Logger, cfg config,
	host *runtime.Host, land *pathfind.Landscape, f *facts.Facts) error {

	if land == nil {
		return fmt.Errorf("serveClient: no landscape loaded (need -facts pointing at OpenRSC root)")
	}
	if _, ok := waitForLivePosition(host); !ok {
		return fmt.Errorf("serveClient: host position never loaded (still 0,0)")
	}
	modelsPath := filepath.Join(cfg.factsRoot, "Client_Base", "Cache", "video", "models.orsc")
	bundle, err := render.OpenBundle(modelsPath)
	if err != nil {
		return fmt.Errorf("serveClient: open models %q: %w", modelsPath, err)
	}

	atoiOr := func(s string, def int) int {
		if n, err2 := parseIntStr(s); err2 == nil {
			return n
		}
		return def
	}

	// motion tracks per-actor tile changes to glide them smoothly (shared by
	// the frame and shot/clip helpers; thread-safe).
	motion := newMotionTracker()

	// renderOne is the frame-capture helper: build the live view with the
	// caller's camera params and render one PNG.
	renderOne := func(q map[string][]string, animFrame int) ([]byte, error) {
		get := func(k string) string {
			if v := q[k]; len(v) > 0 {
				return v[0]
			}
			return ""
		}
		pos := host.World().Self.Position()
		v := buildLiveView(host, pos)
		motion.apply(&v, v.X, v.Y, time.Now())
		v.Rotation = atoiOr(get("rot"), cfg.renderRotation) & 0xff
		v.Zoom = atoiOr(get("zoom"), cfg.renderZoom)
		v.W = atoiOr(get("w"), cfg.renderW)
		v.H = atoiOr(get("h"), cfg.renderH)
		v.AnimFrame = animFrame
		return render.RenderView(land, f, bundle, v)
	}

	// ---- action worker -------------------------------------------------------
	//
	// ONE depth-1 channel for all mutating actions. Walk uses coalesce:true
	// (drain-then-send, fire-and-forget retarget). Interactions and chat use
	// coalesce:false with a done channel; the handler waits on done with a
	// ~1.5s soft timeout, after which it returns ok:true (running) while the
	// worker keeps going. This matches the §6.4 / §9 spec.
	actCh := make(chan actReq, 1)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-actCh:
				wctx, cancel := context.WithTimeout(ctx, 30*time.Second)
				msg, runErr := req.run(wctx)
				cancel()
				if req.done != nil {
					select {
					case req.done <- actResult{msg, runErr}:
					default:
					}
				}
			}
		}
	}()

	// errSuperseded signals a queued-but-not-yet-run action's waiting handler
	// that a newer action replaced it before the worker picked it up — RSC's
	// "latest action wins" semantic. The handler reports it as {ok:false} with
	// this message rather than falsely claiming the action ran.
	errSuperseded := errors.New("superseded by a newer action")

	// drainPending pops the one buffered req (if any) off the depth-1 actCh
	// WITHOUT running it, and — crucially — signals its waiting handler so a
	// drained INTERACTION is not silently dropped (its handler would otherwise
	// block until the soft timeout and falsely report "(running)" for an action
	// that never fired). A coalescing walk has done==nil; nothing to signal.
	drainPending := func() {
		select {
		case old := <-actCh:
			if old.done != nil {
				select {
				case old.done <- actResult{"", errSuperseded}:
				default:
				}
			}
		default:
		}
	}

	// enqueueAction puts an interaction (non-walk) on the worker with a done
	// channel and waits up to softTimeout for the result. On timeout it returns
	// the "running" acknowledgement string so the HTTP handler is always snappy.
	//
	// Serialization invariant: EVERY real packet runs on the single actCh worker
	// — never on a detached side goroutine — so ordering and "latest action wins"
	// hold. When the depth-1 buffer already holds a not-yet-picked-up action, we
	// drain-then-send (superseding the older one and signalling its handler) so
	// the newest interaction replaces the queued one rather than spawning a
	// parallel runner that could put two packets in flight at once.
	const actionSoftTimeout = 1500 * time.Millisecond
	enqueueAction := func(fn func(ctx context.Context) (string, error)) (string, error) {
		done := make(chan actResult, 1)
		req := actReq{run: fn, done: done, coalesce: false}
		select {
		case actCh <- req:
		default:
			// Buffer full: supersede whatever is queued-but-unstarted, then send.
			drainPending()
			actCh <- req // a drained depth-1 channel now has room
		}
		select {
		case res := <-done:
			return res.msg, res.err
		case <-time.After(actionSoftTimeout):
			return "(running)", nil
		}
	}

	// enqueueWalk is the retarget (coalescing) variant: drain any pending action
	// target, send the new walk, return immediately. Matches the existing walkCh
	// pattern in spectate.go. drainPending signals a drained interaction's handler
	// (so a click-to-walk that supersedes a queued Attack does not leave that
	// handler hanging until its soft timeout reporting a false "(running)").
	enqueueWalk := func(fn func(ctx context.Context) (string, error)) {
		drainPending()
		select {
		case actCh <- actReq{run: fn, coalesce: true}:
		default:
		}
	}

	// ---- follow lane ---------------------------------------------------------
	//
	// Follow blocks until its ctx is cancelled. It runs on its own goroutine so
	// it never wedges the shared action worker. Only one Follow runs at a time;
	// a new /act follow (or any other /act) cancels the running one.
	var (
		followMu     sync.Mutex
		followCancel context.CancelFunc
	)
	cancelFollow := func() {
		followMu.Lock()
		defer followMu.Unlock()
		if followCancel != nil {
			followCancel()
			followCancel = nil
		}
	}
	startFollow := func(name string) {
		cancelFollow() // cancel any prior follow
		fctx, fcancel := context.WithCancel(ctx)
		followMu.Lock()
		followCancel = fcancel
		followMu.Unlock()
		go func() {
			defer fcancel()
			const followStartup = 10 * time.Second
			if err := host.Follow(fctx, name, followStartup); err != nil &&
				!errors.Is(err, context.Canceled) &&
				!errors.Is(err, context.DeadlineExceeded) {
				log.Debug("follow ended", "target", name, "err", err)
			}
		}()
	}

	// ---- chat ring -----------------------------------------------------------
	ring := &chatRing{}

	// ---- Layer 2 dispatcher --------------------------------------------------
	wv := &worldViewAdapter{w: host.World()}
	disp := remoteclient.NewDispatcher(host, host, f, wv)

	// ---- HTTP mux ------------------------------------------------------------
	mux := http.NewServeMux()
	registerSpriteRoutes(mux, log) // GET /sprite — authentic item icons

	// GET / — the React SPA, embed'd from web/dist (Layer 4). Static build
	// files are served directly; any other path falls back to index.html so
	// client-side routing works. The legacy single-file client stays available
	// at /legacy for comparison/debugging.
	webFS := web.Dist()
	fileServer := http.FileServerFS(webFS)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, "/")
		if p == "" {
			p = "index.html"
		}
		if f, err := webFS.Open(p); err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: unknown path → index.html.
		index, err := fs.ReadFile(webFS, "index.html")
		if err != nil {
			http.Error(w, "client build missing (run `npm run build` in web/)", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(index)
	})

	// GET /legacy — the original single-file client (clientPage from
	// clientpage.go). Retained so we can A/B the React port against it.
	mux.HandleFunc("/legacy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		fmt.Fprintf(w, clientPage, cfg.username, cfg.renderZoom, cfg.renderW, cfg.renderH)
	})

	// GET /config — render defaults the SPA seeds its camera from (replaces the
	// fmt.Fprintf template injection the legacy page used).
	mux.HandleFunc("/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, `{"username":%q,"zoom":%d,"w":%d,"h":%d,"rotation":%d}`,
			cfg.username, cfg.renderZoom, cfg.renderW, cfg.renderH, cfg.renderRotation&0xff)
	})

	// GET /pos — host tile + plane (§3 of 30-http-api.md).
	mux.HandleFunc("/pos", func(w http.ResponseWriter, r *http.Request) {
		p := host.World().Self.Position()
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		fmt.Fprintf(w, `{"x":%d,"y":%d,"plane":%d}`, p.X, p.Y, p.Plane())
	})

	// GET /frame — one freshly-rendered PNG of the host's view (§3).
	mux.HandleFunc("/frame", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		v := buildLiveView(host, pos)
		motion.apply(&v, v.X, v.Y, time.Now())
		v.Rotation = atoiOr(q.Get("rot"), cfg.renderRotation) & 0xff
		v.Zoom = atoiOr(q.Get("zoom"), cfg.renderZoom)
		v.W = atoiOr(q.Get("w"), cfg.renderW)
		v.H = atoiOr(q.Get("h"), cfg.renderH)
		v.AnimFrame = atoiOr(q.Get("anim"), 0)
		pngData, err := render.RenderView(land, f, bundle, v)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write(pngData)
	})

	// GET /walk?px&py&rot&zoom&w&h — screen click → tile → Walk (§3).
	// Retained with the exact same semantics as spectate.go so the legacy page
	// (and the SPA's plain left-click-to-walk) can call this endpoint too.
	mux.HandleFunc("/walk", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		px, py := atoiOr(q.Get("px"), -1), atoiOr(q.Get("py"), -1)
		if px < 0 || py < 0 {
			http.Error(w, "missing px/py", http.StatusBadRequest)
			return
		}
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		plane := pos.Plane()
		v := render.View{
			X:        pos.X,
			Y:        pos.Y - plane*world.PlaneHeight,
			Plane:    plane,
			Rotation: atoiOr(q.Get("rot"), cfg.renderRotation) & 0xff,
			Zoom:     atoiOr(q.Get("zoom"), cfg.renderZoom),
			W:        atoiOr(q.Get("w"), cfg.renderW),
			H:        atoiOr(q.Get("h"), cfg.renderH),
		}
		tx, ty, ok := render.PickTile(land, v, px, py)
		if !ok {
			http.Error(w, "no tile under click", http.StatusNoContent)
			return
		}
		absY := ty + plane*world.PlaneHeight
		enqueueWalk(func(ctx context.Context) (string, error) {
			return fmt.Sprintf("Walk here (%d, %d)", tx, absY), host.Walk(ctx, tx, absY)
		})
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"x":%d,"y":%d}`, tx, absY)
	})

	// GET /shot — save current frame to spectatorShotDir (§3).
	mux.HandleFunc("/shot", func(w http.ResponseWriter, r *http.Request) {
		img, err := renderOne(r.URL.Query(), atoiOr(r.URL.Query().Get("anim"), 0))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_ = os.MkdirAll(spectatorShotDir, 0o755)
		stamp := time.Now().Format("150405")
		numbered := filepath.Join(spectatorShotDir, "shot_"+stamp+".png")
		_ = os.WriteFile(numbered, img, 0o644)
		_ = os.WriteFile(filepath.Join(spectatorShotDir, "shot.png"), img, 0o644)
		log.Info("client /shot saved", "path", numbered)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"saved":%q}`, numbered)
	})

	// GET /clip — 12-frame burst contact sheet (§3).
	mux.HandleFunc("/clip", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		const nFrames = 12
		const gap = 220 * time.Millisecond
		_ = os.MkdirAll(spectatorShotDir, 0o755)
		var frames [][]byte
		for k := 0; k < nFrames; k++ {
			img, err := renderOne(q, k)
			if err == nil {
				frames = append(frames, img)
				_ = os.WriteFile(filepath.Join(spectatorShotDir, fmt.Sprintf("clip_%02d.png", k)), img, 0o644)
			}
			if k < nFrames-1 {
				time.Sleep(gap)
			}
		}
		sheet, err := montage(frames, 4)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		stamp := time.Now().Format("150405")
		p := filepath.Join(spectatorShotDir, "clip_sheet_"+stamp+".png")
		_ = os.WriteFile(p, sheet, 0o644)
		_ = os.WriteFile(filepath.Join(spectatorShotDir, "clip_sheet.png"), sheet, 0o644)
		log.Info("client /clip saved", "sheet", p, "frames", len(frames))
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		fmt.Fprintf(w, `{"saved":%q,"frames":%d}`, p, len(frames))
	})

	// POST /pick — screen click → ordered menu candidates (§4 of 30-http-api.md).
	mux.HandleFunc("/pick", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req pickRequest
		// Default Slot to -1 so a body that omits "slot" (a plain screen pick) is
		// NOT mistaken for inventory slot 0. The browser sends slot:-1 explicitly
		// on screen picks (clientpage pick()), but a hand-rolled client that omits
		// the key still decodes to the screen-pick path, not slot 0.
		req.Slot = -1
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Inventory fallback (spec §6.1): Slot>=0 means "give me the right-click
		// menu for this inventory slot", not a screen pixel. Build the candidate
		// from the live inventory + the SAME InventoryMenu /state uses, so the
		// fallback menu matches the inline /state menu exactly. No render.Pick, no
		// camera — w/h are irrelevant here.
		if req.Slot >= 0 {
			slots := host.World().Inventory.Slots()
			if req.Slot >= len(slots) || slots[req.Slot].ItemID == 0 {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.Header().Set("Cache-Control", "no-store")
				_ = json.NewEncoder(w).Encode(remoteclient.PickResponse{Candidates: []remoteclient.Candidate{}})
				return
			}
			s := slots[req.Slot]
			var def *facts.ItemDef
			if f != nil {
				def = f.ItemDef(s.ItemID)
			}
			cand := remoteclient.InventoryMenu(host, req.Slot, s, def)
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			_ = json.NewEncoder(w).Encode(remoteclient.PickResponse{Candidates: []remoteclient.Candidate{cand}})
			return
		}

		if req.W <= 0 || req.H <= 0 {
			http.Error(w, "w and h required", http.StatusBadRequest)
			return
		}

		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		plane := pos.Plane()

		// Build the FULL live view — the SAME one /frame renders — so render.Pick
		// can hit-test NPC/player/self billboards, ground items, and dynamic
		// scenery, not just the tile-grounded targets a bare view yields. (The
		// /walk handler uses a bare view because PickTile only needs the camera;
		// /pick needs the entities too, or NPCs/players/items are never pickable.)
		// Apply the same motion glide so the hit boxes line up with the sprites the
		// browser last saw, then overlay the caller's camera (render.Pick rebuilds
		// the identical camera internally).
		v := buildLiveView(host, pos)
		motion.apply(&v, v.X, v.Y, time.Now())
		v.Rotation = req.Rot & 0xff
		v.Zoom = req.Zoom
		v.W = req.W
		v.H = req.H
		if req.Zoom == 0 {
			v.Zoom = cfg.renderZoom
		}

		// Layer 1: pick candidates.
		rawCands := render.Pick(land, f, v, req.PX, req.PY)

		// Layer 2: map to wire Candidates (folds plane into absolute Y,
		// resolves defs + examine text, builds option lists).
		cands := remoteclient.BuildCandidates(host, f, rawCands, pos.X, pos.Y-plane*world.PlaneHeight, plane)

		resp := remoteclient.PickResponse{Candidates: cands}
		if resp.Candidates == nil {
			resp.Candidates = []remoteclient.Candidate{} // never null
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// POST /act — execute one menu option on one target (§5 of 30-http-api.md).
	mux.HandleFunc("/act", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req actRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		if string(req.Ref.Kind) == "" {
			http.Error(w, "ref.kind required", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")

		// Route FIRST with the pure resolver (no Host call, no packet). This is the
		// fix for the double-dispatch hazard: Dispatch executes the Host method as
		// it resolves, so using it for routing AND inside the worker would fire
		// every interaction twice — the first on an unscoped context outside the
		// serialized worker. ResolveLane re-derives the lane + re-validates
		// identity with zero side effects; we then call Dispatch EXACTLY ONCE
		// inside the chosen lane with that lane's worker ctx.
		lane, resErr := disp.ResolveLane(req.Ref, req.OptionID)

		// Malformed request (ErrUnknownOption → 400); stale target → {ok:false}.
		if errors.Is(resErr, remoteclient.ErrUnknownOption) {
			http.Error(w, resErr.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(resErr, remoteclient.ErrStaleTarget) {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: resErr.Error()})
			return
		}
		if resErr != nil {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: resErr.Error()})
			return
		}

		ref := req.Ref       // capture for closures
		optID := req.OptionID // capture for closures

		switch lane {
		case remoteclient.LaneSync:
			// Examine (and the reserved M1 Use no-op): no packet — resolve once,
			// synchronously, right here. There is no worker and nothing to enqueue.
			msg, _, err := disp.Dispatch(r.Context(), ref, optID)
			if err != nil {
				_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: err.Error()})
				return
			}
			_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
			return

		case remoteclient.LaneFollow:
			// Follow blocks until ctx cancel — startFollow runs the single Follow
			// call on its own cancellable goroutine. We do NOT call Dispatch here
			// (that would start a SECOND, un-cancellable Follow on r's ctx).
			cancelFollow()
			startFollow(ref.Name)
			_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: "Follow " + ref.Name})
			return

		case remoteclient.LaneWalk:
			// Walk-retarget: fire and forget on the coalescing lane. Dispatch (the
			// single Walk call) runs INSIDE enqueueWalk with the worker ctx.
			cancelFollow()
			enqueueWalk(func(wctx context.Context) (string, error) {
				msg, _, e := disp.Dispatch(wctx, ref, optID)
				return msg, e
			})
			_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: "Walk here"})
			return

		default: // LaneAction
			// Serialized single-action lane: Dispatch (the single interaction
			// packet) runs exactly once INSIDE enqueueAction with the worker ctx.
			cancelFollow()
			msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
				m, _, e := disp.Dispatch(wctx, ref, optID)
				return m, e
			})
			if runErr != nil {
				// ErrStaleTarget can still surface if the world shifted between
				// ResolveLane and the worker actually running; map both it and any
				// other error to {ok:false} with the message.
				_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
				return
			}
			if msg == "" {
				msg = "ok"
			}
			_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
		}
	})

	// GET /state — self + inventory + equipment + chat (§6 of 30-http-api.md).
	mux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		pos := host.World().Self.Position()
		if pos.X <= 0 && pos.Y <= 0 {
			http.Error(w, "host position not loaded", http.StatusServiceUnavailable)
			return
		}
		self := host.World().Self
		plane := pos.Plane()

		// self block
		skills := make([]stateSkill, world.NumSkills)
		for i := 0; i < world.NumSkills; i++ {
			skills[i] = stateSkill{
				ID:    i,
				Name:  event.SkillName(event.SkillID(i)),
				Level: self.SkillLevel(i),
				Max:   self.SkillMax(i),
				XP:    self.SkillXP(i),
			}
		}
		selfBlock := stateSelf{
			X:           pos.X,
			Y:           pos.Y,
			Plane:       plane,
			Heading:     self.Heading(),
			CombatLevel: self.CombatLevel(),
			HP:          self.HP(),
			MaxHP:       self.MaxHP(),
			Prayer:      self.Prayer(),
			MaxPrayer:   self.MaxPrayer(),
			Fatigue:     self.Fatigue(),
			QuestPoints: self.QuestPoints(),
			Skills:      skills,
		}

		// inventory: occupied slots only, with BuildMenu options.
		slots := host.World().Inventory.Slots()
		invItems := make([]stateInvItem, 0, len(slots))
		for slotIdx, s := range slots {
			if s.ItemID == 0 {
				continue
			}
			var def *facts.ItemDef
			if f != nil {
				def = f.ItemDef(s.ItemID)
			}
			cand := remoteclient.InventoryMenu(host, slotIdx, s, def)
			name := cand.Label
			wearable := def != nil && def.IsWearable
			stackable := def != nil && def.IsStackable
			cmd := ""
			if def != nil {
				cmd = def.Command
			}
			defaultOptID := 0
			invItems = append(invItems, stateInvItem{
				Slot:            slotIdx,
				ItemID:          s.ItemID,
				Name:            name,
				Amount:          s.Amount,
				Wielded:         s.Wielded,
				Wearable:        wearable,
				Stackable:       stackable,
				Command:         cmd,
				DefaultOptionID: defaultOptID,
				Options:         cand.Options,
			})
		}

		// equipment: sprite > 0 slots only.
		equip := self.EquipSprites()
		var equipItems []stateEquipItem
		for i, sprite := range equip {
			if sprite > 0 {
				equipItems = append(equipItems, stateEquipItem{
					Slot:   event.EquipSlotName(i),
					Sprite: sprite,
					ItemID: 0,
				})
			}
		}
		if equipItems == nil {
			equipItems = []stateEquipItem{}
		}

		// chat ring: ingest any new Recent events, then snapshot.
		ring.ingestRecent(host.World().Recent)
		chatEntries := ring.snapshot()
		if chatEntries == nil {
			chatEntries = []chatEntry{}
		}

		// bank block: present only while the bank window is open.
		var bankBlock *stateBank
		if rec := host.World().Bank.Bank(); rec != nil {
			bslots := make([]stateBankSlot, 0, len(rec.Slots))
			for i, bs := range rec.Slots {
				if bs.ItemID == 0 && bs.Amount == 0 {
					continue
				}
				name := ""
				if f != nil {
					if def := f.ItemDef(bs.ItemID); def != nil {
						name = def.Name
					}
				}
				bslots = append(bslots, stateBankSlot{
					Slot: i, ItemID: bs.ItemID, Name: name, Amount: bs.Amount,
				})
			}
			bankBlock = &stateBank{Open: true, MaxSize: rec.MaxSize, Slots: bslots}
		}

		resp := stateResponse{
			Self:      selfBlock,
			Inventory: invItems,
			Equipment: equipItems,
			Chat:      chatEntries,
			Bank:      bankBlock,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(resp)
	})

	// POST /bank — deposit/withdraw/close on the open bank window (110-react §E1).
	// Routed through the same serialized action worker as /act and /chat.
	mux.HandleFunc("/bank", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req bankRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		switch req.Op {
		case "deposit", "withdraw":
			if !host.World().Bank.IsOpen() {
				_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "bank is not open"})
				return
			}
			if req.Amount <= 0 {
				_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "amount must be > 0"})
				return
			}
		case "close":
		default:
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "unknown op: must be deposit|withdraw|close"})
			return
		}
		op, id, amount := req.Op, req.ItemID, req.Amount
		msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
			switch op {
			case "deposit":
				return fmt.Sprintf("Deposit %d of item %d", amount, id), host.BankDeposit(wctx, id, amount)
			case "withdraw":
				return fmt.Sprintf("Withdraw %d of item %d", amount, id), host.BankWithdraw(wctx, id, amount)
			case "close":
				return "Close bank", host.BankClose(wctx)
			}
			return "", fmt.Errorf("unknown bank op %q", op)
		})
		if runErr != nil {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
	})

	// POST /chat — send public / command / private message (§7 of 30-http-api.md).
	mux.HandleFunc("/chat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		req.Text = strings.TrimSpace(req.Text)
		if req.Text == "" {
			http.Error(w, "text required", http.StatusBadRequest)
			return
		}
		switch req.Kind {
		case "say", "command", "pm":
		default:
			http.Error(w, "unknown kind: must be say|command|pm", http.StatusBadRequest)
			return
		}
		if req.Kind == "pm" && strings.TrimSpace(req.To) == "" {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "pm requires 'to' field"})
			return
		}

		// Capture for closure.
		kind, text, to := req.Kind, req.Text, req.To
		msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
			switch kind {
			case "say":
				return "Say: " + text, host.Say(wctx, text)
			case "command":
				// Strip a leading "::" if the user typed one — Command wants the
				// bare command string (per §7 of 30-http-api.md).
				bare := strings.TrimPrefix(text, "::")
				return "Command: " + bare, host.Command(wctx, bare)
			case "pm":
				return "PM to " + to + ": " + text, host.PrivateMessage(wctx, to, text)
			}
			return "", fmt.Errorf("unknown chat kind %q", kind)
		})

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if runErr != nil {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
			return
		}
		// Append to chat ring as a "self" entry so /state includes it.
		ring.add(chatKindSelf, cfg.username, text)
		_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
	})

	// GET /examine (also POST) — examine one ref without a /pick round-trip (§8).
	mux.HandleFunc("/examine", func(w http.ResponseWriter, r *http.Request) {
		var ref remoteclient.MenuTarget
		var decodeErr error
		switch r.Method {
		case http.MethodGet:
			// ref is URL-encoded JSON in ?ref=...
			raw := r.URL.Query().Get("ref")
			if raw == "" {
				http.Error(w, "ref query param required", http.StatusBadRequest)
				return
			}
			decodeErr = json.Unmarshal([]byte(raw), &ref)
		case http.MethodPost:
			var body struct {
				Ref remoteclient.MenuTarget `json:"ref"`
			}
			decodeErr = json.NewDecoder(r.Body).Decode(&body)
			ref = body.Ref
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if decodeErr != nil {
			http.Error(w, "bad ref: "+decodeErr.Error(), http.StatusBadRequest)
			return
		}
		if string(ref.Kind) == "" {
			http.Error(w, "ref.kind required", http.StatusBadRequest)
			return
		}

		var ex runtime.Examination
		switch ref.Kind {
		case remoteclient.KindNPC:
			ex = host.ExamineNpc(ref.Index)
		case remoteclient.KindPlayer:
			ex = host.ExaminePlayer(ref.Index)
		case remoteclient.KindSelf:
			ex = host.ExamineSelf()
		case remoteclient.KindGroundItem:
			ex = host.ExamineGroundItem(ref.X, ref.Y)
		case remoteclient.KindScenery:
			ex = host.ExamineScenery(ref.X, ref.Y)
		case remoteclient.KindBoundary:
			ex = host.ExamineBoundary(ref.X, ref.Y)
		case remoteclient.KindInventoryItem:
			ex = host.ExamineInventorySlot(ref.Slot)
		case remoteclient.KindTerrain:
			ex = runtime.Examination{Kind: "terrain"}
		default:
			http.Error(w, "unknown ref.kind", http.StatusBadRequest)
			return
		}

		// Serialise Examination with lowerCamelCase field names per §8 spec.
		type examResp struct {
			Kind        string `json:"kind"`
			Name        string `json:"name"`
			Description string `json:"description"`
			X           int    `json:"x"`
			Y           int    `json:"y"`
			Detail      string `json:"detail"`
		}
		out := examResp{
			Kind:        ex.Kind,
			Name:        ex.Name,
			Description: ex.Description,
			X:           ex.X,
			Y:           ex.Y,
			Detail:      ex.Detail,
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(out)
	})

	// ---- start serving -------------------------------------------------------
	srv := &http.Server{Addr: cfg.clientAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		cancelFollow()
		sc, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sc)
	}()

	log.Info("remote client serving — open this in a browser; Ctrl-C to stop",
		"url", "http://"+cfg.clientAddr+"/", "host", cfg.username)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serveClient: serve %q: %w", cfg.clientAddr, err)
	}
	return nil
}

// parseIntStr converts a string to int using the standard library, returning
// the int and an error. Used by the atoiOr closure inside serveClient to avoid
// importing strconv directly (it is already in scope via the package).
func parseIntStr(s string) (int, error) {
	// Import-free inline; we need strconv but prefer not to shadow the import
	// in the closure. Use fmt.Sscan as a quick alternative that lives in the
	// already-imported "fmt" package.
	var n int
	_, err := fmt.Sscan(s, &n)
	return n, err
}
