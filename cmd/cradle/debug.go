package main

// The debug-http control plane. A live, scriptable HTTP harness over a
// logged-in Host so an external driver (a human, or an agent like Claude)
// can issue DSL commands one at a time, read structured world state, and
// pull a complete record of every event the server sends.
//
// It is the non-blocking sibling of `-repl`: the REPL reads DSL off stdin
// and prints to stdout (great for a human at a terminal, awkward to drive
// programmatically); this serves the SAME persistent interpreter session
// over HTTP so every command is a discrete request with a JSON reply, and
// every bus event is appended to a JSONL log. Mirrors the `-spectate`
// pattern: start after login, serve until the root context is cancelled
// (Ctrl-C), keeping the host logged in the whole time.
//
// Endpoints:
//
//	POST /eval     body = one DSL line  -> {ok, value, is_expression, error, events_before}
//	POST /script   body = a .routine    -> {ok, kind, value, error}  (run with a timeout)
//	GET  /state                          -> rich JSON snapshot of the world mirror
//	GET  /events?since=N&kind=K&limit=L   -> recorded events with seq > N
//	GET  /                               -> this help
//
// Every command issued and every event observed is also written to the
// JSONL event log (default /tmp/cradle_debug/<username>_events.jsonl) so a
// full run-through is replayable and greppable after the fact.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/runtime"
)

// debugServer holds the live host, the persistent interpreter session the
// /eval and /script endpoints drive, and the event recorder.
type debugServer struct {
	log  *slog.Logger
	cfg  config
	host *runtime.Host

	// evalMu serializes DSL execution (/eval, /script) so only one
	// command sends packets to the server at a time and the session env
	// isn't mutated concurrently. /state and /events do NOT take this —
	// they are concurrent reads.
	evalMu sync.Mutex
	sess   *interp.Session
	ctx    context.Context

	// recMu protects the in-memory event ring + seq counter.
	recMu   sync.Mutex
	records []eventRecord
	seq     int
	maxRing int
	logFile *os.File
}

// eventRecord is one observed bus event, serialized. Data is the event
// struct's exported fields (the embedded `base` timestamp is unexported,
// so At is carried explicitly from ev.Time()).
type eventRecord struct {
	Seq  int             `json:"seq"`
	At   time.Time       `json:"at"`
	Kind string          `json:"kind"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func newDebugServer(log *slog.Logger, cfg config, host *runtime.Host) *debugServer {
	return &debugServer{log: log, cfg: cfg, host: host, maxRing: 100000}
}

// startRecorder subscribes to every bus event and records it (in-memory
// ring + JSONL file). Call this as early as possible — right after the
// connection is up — so login/welcome/initial-snapshot events are captured.
// The goroutine exits when ctx is cancelled or the bus closes.
func (d *debugServer) startRecorder(ctx context.Context) {
	path := d.cfg.debugLog
	if path == "" {
		path = filepath.Join("/tmp/cradle_debug", d.cfg.username+"_events.jsonl")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		d.log.Warn("debug: cannot create event-log dir; recording to memory only", "dir", filepath.Dir(path), "err", err)
	} else if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
		d.log.Warn("debug: cannot open event log; recording to memory only", "path", path, "err", err)
	} else {
		d.logFile = f
		d.log.Info("debug: recording all events to JSONL", "path", path)
	}

	ch := d.host.Bus().Subscribe("*", 8192)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				d.record(ev)
			}
		}
	}()
}

// record appends one event to the ring and the JSONL log. The recorder
// goroutine is the only writer to logFile, so file writes need no lock.
func (d *debugServer) record(ev event.Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		data, _ = json.Marshal(map[string]string{"_marshal_error": err.Error()})
	}
	// Many event structs are constructed without the embedded `base`
	// timestamp, so ev.Time() is the zero value — fall back to the
	// wall-clock at record time so the log is usable for replay/timing.
	at := ev.Time()
	if at.IsZero() {
		at = time.Now()
	}
	d.recMu.Lock()
	d.seq++
	rec := eventRecord{
		Seq:  d.seq,
		At:   at,
		Kind: ev.Kind(),
		Type: fmt.Sprintf("%T", ev),
		Data: data,
	}
	d.records = append(d.records, rec)
	if len(d.records) > d.maxRing {
		d.records = d.records[len(d.records)-d.maxRing:]
	}
	d.recMu.Unlock()

	if d.logFile != nil {
		if line, err := json.Marshal(rec); err == nil {
			_, _ = d.logFile.Write(append(line, '\n'))
		}
	}
}

func (d *debugServer) currentSeq() int {
	d.recMu.Lock()
	defer d.recMu.Unlock()
	return d.seq
}

// serve builds the persistent interpreter session and serves the HTTP
// control plane, blocking until ctx is cancelled. The recorder should
// already be running (startRecorder) before this is called.
func (d *debugServer) serve(ctx context.Context) error {
	d.ctx = ctx
	it := d.host.NewRoutineInterpreter(ctx)
	d.sess = it.NewSession(ctx, "<debug-http>")

	// Passively pump bus events into the session so top-level `on`
	// handlers registered via /eval fire while the session is idle (no
	// routine running). Guarded by evalMu so it never races a /eval.
	go func() {
		t := time.NewTicker(200 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				d.evalMu.Lock()
				d.sess.PumpEvents(ctx)
				d.evalMu.Unlock()
			}
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleHelp)
	mux.HandleFunc("/eval", d.handleEval)
	mux.HandleFunc("/script", d.handleScript)
	mux.HandleFunc("/state", d.handleState)
	mux.HandleFunc("/events", d.handleEvents)

	srv := &http.Server{Addr: d.cfg.debugAddr, Handler: mux}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()

	d.log.Info("debug HTTP control plane listening",
		"addr", "http://"+d.cfg.debugAddr,
		"endpoints", "POST /eval, POST /script, GET /state, GET /events?since=N")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ---- handlers -------------------------------------------------------------

func (d *debugServer) handleHelp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	io.WriteString(w, `cradle debug-http control plane

  POST /eval     body = one DSL line (expression or statement).
                 e.g.  curl -s :PORT/eval -d 'self.position'
                       curl -s :PORT/eval -d 'talk_to(world.npcs[0].index)'
                       curl -s :PORT/eval -d 'answer(find_option("yes"))'
                 -> {ok, value, is_expression, error, events_before}
                 (events_before = the event seq just before the command;
                  GET /events?since=<that> to see what the command produced.)

  POST /script   body = a full .routine (multi-line). ?timeout=120s
                 -> {ok, kind, value, error}

  GET  /state    rich JSON snapshot: position, vitals, inventory, nearby
                 npcs (with server_index), players, ground items, dialog
                 options, recent server messages.

  GET  /events   ?since=N (default 0)  &kind=chat_received  &limit=500
                 -> [{seq, at, kind, type, data}, ...]

All commands run against the SAME persistent interpreter session, and
every event is appended to the JSONL event log on disk.
`)
}

// evalResponse is the JSON shape returned by /eval.
type evalResponse struct {
	OK           bool   `json:"ok"`
	Value        string `json:"value,omitempty"`
	IsExpression bool   `json:"is_expression"`
	Error        string `json:"error,omitempty"`
	EventsBefore int    `json:"events_before"`
}

func (d *debugServer) handleEval(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	line := strings.TrimSpace(string(body))
	if line == "" {
		line = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	if line == "" {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "empty command; POST a DSL line as the body or pass ?q="})
		return
	}

	d.evalMu.Lock()
	before := d.currentSeq()
	res := d.sess.Eval(d.ctx, line)
	d.evalMu.Unlock()

	resp := evalResponse{OK: res.Err == nil, IsExpression: res.IsExpression, EventsBefore: before}
	if res.Err != nil {
		resp.Error = res.Err.Error()
	}
	if res.Value != nil {
		resp.Value = res.Value.Display()
	}
	d.log.Info("debug eval", "line", line, "ok", resp.OK, "value", resp.Value, "err", resp.Error)
	writeJSON(w, http.StatusOK, resp)
}

// scriptResponse is the JSON shape returned by /script.
type scriptResponse struct {
	OK    bool   `json:"ok"`
	Kind  string `json:"kind,omitempty"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

func (d *debugServer) handleScript(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	src := string(body)
	if strings.TrimSpace(src) == "" {
		writeJSON(w, http.StatusBadRequest, scriptResponse{Error: "empty script body"})
		return
	}

	timeout := 120 * time.Second
	if q := r.URL.Query().Get("timeout"); q != "" {
		if dur, err := time.ParseDuration(q); err == nil {
			timeout = dur
		} else if secs, err := strconv.Atoi(q); err == nil {
			timeout = time.Duration(secs) * time.Second
		}
	}

	rf, err := runtime.ParseRoutineString("<debug-script>", src)
	if err != nil {
		writeJSON(w, http.StatusOK, scriptResponse{Error: err.Error()})
		return
	}

	d.evalMu.Lock()
	defer d.evalMu.Unlock()
	rctx, cancel := context.WithTimeout(d.ctx, timeout)
	defer cancel()
	it := d.host.NewRoutineInterpreter(rctx)
	res := it.RunRoutine(rctx, rf.File, nil)

	resp := scriptResponse{OK: res.Err == nil, Kind: res.Kind.String()}
	if res.Err != nil {
		resp.Error = res.Err.Error()
	}
	if res.Value != nil {
		resp.Value = res.Value.Display()
	}
	d.log.Info("debug script", "kind", resp.Kind, "ok", resp.OK, "err", resp.Error)
	writeJSON(w, http.StatusOK, resp)
}

// ---- /state snapshot ------------------------------------------------------

type stateSnapshot struct {
	X            int            `json:"x"`
	Y            int            `json:"y"`
	HP           int            `json:"hp"`
	MaxHP        int            `json:"max_hp"`
	Prayer       int            `json:"prayer"`
	MaxPrayer    int            `json:"max_prayer"`
	Fatigue      int            `json:"fatigue"`
	CombatLevel  int            `json:"combat_level"`
	QuestPoints  int            `json:"quest_points"`
	InvUsed      int            `json:"inventory_used"`
	InvFree      int            `json:"inventory_free"`
	Inventory    []invItem      `json:"inventory"`
	Npcs         []npcSnap      `json:"npcs"`
	Players      []playerSnap   `json:"players"`
	GroundItems  []groundSnap   `json:"ground_items"`
	DialogOpen   bool           `json:"dialog_open"`
	DialogText   string         `json:"dialog_text,omitempty"`
	DialogOpts   []string       `json:"dialog_options,omitempty"`
	LastServer   string         `json:"last_server_message,omitempty"`
	RecentServer []string       `json:"recent_server_messages,omitempty"`
	Seq          int            `json:"event_seq"`
}

type invItem struct {
	Slot   int    `json:"slot"`
	ItemID int    `json:"item_id"`
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type npcSnap struct {
	ServerIndex int    `json:"server_index"`
	TypeID      int    `json:"type_id"`
	Name        string `json:"name"`
	X           int    `json:"x"`
	Y           int    `json:"y"`
}

type playerSnap struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

type groundSnap struct {
	ItemID int    `json:"item_id"`
	Name   string `json:"name"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
}

func (d *debugServer) handleState(w http.ResponseWriter, r *http.Request) {
	wld := d.host.World()
	f := d.host.Facts()
	snap := stateSnapshot{Seq: d.currentSeq()}

	if wld != nil && wld.Self != nil {
		pos := wld.Self.Position()
		snap.X, snap.Y = pos.X, pos.Y
		snap.HP, snap.MaxHP = wld.Self.HP(), wld.Self.MaxHP()
		snap.Prayer, snap.MaxPrayer = wld.Self.Prayer(), wld.Self.MaxPrayer()
		snap.Fatigue = wld.Self.Fatigue()
		snap.CombatLevel = wld.Self.CombatLevel()
		snap.QuestPoints = wld.Self.QuestPoints()
	}

	if wld != nil && wld.Inventory != nil {
		snap.InvFree = wld.Inventory.FreeSlots()
		for i, s := range wld.Inventory.Slots() {
			name := ""
			if f != nil {
				if def := f.ItemDef(s.ItemID); def != nil {
					name = def.Name
				}
			}
			snap.Inventory = append(snap.Inventory, invItem{Slot: i, ItemID: s.ItemID, Name: name, Amount: s.Amount})
		}
		snap.InvUsed = len(snap.Inventory)
	}

	if wld != nil && wld.Npcs != nil {
		for _, n := range wld.Npcs.All() {
			name := ""
			if f != nil {
				if def := f.NpcDef(n.TypeID); def != nil {
					name = def.Name
				}
			}
			snap.Npcs = append(snap.Npcs, npcSnap{ServerIndex: n.Index, TypeID: n.TypeID, Name: name, X: n.X, Y: n.Y})
		}
	}

	if wld != nil && wld.Players != nil {
		for _, p := range wld.Players.All() {
			snap.Players = append(snap.Players, playerSnap{Index: p.Index, Name: p.Name, X: p.X, Y: p.Y})
		}
	}

	if wld != nil && wld.GroundItems != nil {
		for _, g := range wld.GroundItems.All() {
			name := ""
			if f != nil {
				if def := f.ItemDef(g.ItemID); def != nil {
					name = def.Name
				}
			}
			snap.GroundItems = append(snap.GroundItems, groundSnap{ItemID: g.ItemID, Name: name, X: g.X, Y: g.Y})
		}
	}

	if wld != nil && wld.Recent != nil {
		if opts := wld.Recent.DialogOptions(); opts != nil {
			snap.DialogOpen = true
			snap.DialogOpts = opts.Options
		}
		if dt := wld.Recent.DialogText(); dt != nil {
			snap.DialogText = dt.Text
		}
		if sm := wld.Recent.ServerMessage(); sm != nil {
			snap.LastServer = sm.Message
		}
		for _, m := range wld.Recent.ServerMessages() {
			snap.RecentServer = append(snap.RecentServer, m.Message)
		}
	}

	writeJSON(w, http.StatusOK, snap)
}

// ---- /events --------------------------------------------------------------

func (d *debugServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	since, _ := strconv.Atoi(q.Get("since"))
	kind := q.Get("kind")
	limit := 1000
	if l, err := strconv.Atoi(q.Get("limit")); err == nil && l > 0 {
		limit = l
	}

	d.recMu.Lock()
	out := make([]eventRecord, 0, 64)
	for _, rec := range d.records {
		if rec.Seq <= since {
			continue
		}
		if kind != "" && rec.Kind != kind {
			continue
		}
		out = append(out, rec)
		if len(out) >= limit {
			break
		}
	}
	latest := d.seq
	d.recMu.Unlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"latest_seq": latest,
		"count":      len(out),
		"events":     out,
	})
}

// ---- util -----------------------------------------------------------------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
