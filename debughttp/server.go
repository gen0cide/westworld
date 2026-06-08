// Package debughttp is a live, scriptable HTTP control plane over a logged-in
// Host. An external driver (a human in a browser, or an agent) can watch the
// host's state + thought stream, replay every event, and issue DSL commands.
//
// It is the non-blocking sibling of the stdin REPL: the same persistent
// interpreter session, served over HTTP, with every bus event recorded to an
// in-memory ring + a JSONL file. Shared by cmd/cradle and cmd/host.
//
// Endpoints:
//
//	GET  /                               browser dashboard (HTML): live state + thought stream
//	GET  /ws                             WebSocket: live stream of every bus event (incl. agent_thought)
//	GET  /state                          rich JSON snapshot of the world mirror
//	GET  /events?since=N&kind=K&limit=L   recorded events with seq > N
//	POST /eval     body = one DSL line   -> {ok, value, is_expression, error, events_before}
//	POST /script   body = a .routine     -> {ok, kind, value, error}
//	GET  /help                           plain-text help
package debughttp

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

	"github.com/coder/websocket"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/runtime"
)

// Config configures a debug server.
type Config struct {
	Username string // labels the JSONL log + dashboard
	Addr     string // host:port to listen on
	LogPath  string // JSONL event log; default /tmp/cradle_debug/<username>_events.jsonl
	MaxRing  int    // in-memory event-ring cap; <=0 => default (100k). The cradle uses a smaller ring per host.
}

// Server is the debug control plane over one live Host.
type Server struct {
	log  *slog.Logger
	cfg  Config
	host *runtime.Host

	// evalMu serializes DSL execution (/eval, /script) so only one command
	// drives the host at a time. /state, /events, /ws are concurrent reads.
	evalMu sync.Mutex
	sess   *interp.Session
	ctx    context.Context

	recMu     sync.Mutex
	records   []eventRecord
	seq       int
	maxRing   int
	logFile   *os.File
	closeOnce sync.Once
}

type eventRecord struct {
	Seq  int             `json:"seq"`
	At   time.Time       `json:"at"`
	Kind string          `json:"kind"`
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// New builds a debug server over host.
func New(host *runtime.Host, cfg Config, log *slog.Logger) *Server {
	if log == nil {
		log = slog.Default()
	}
	maxRing := cfg.MaxRing
	if maxRing <= 0 {
		maxRing = 100000
	}
	return &Server{log: log, cfg: cfg, host: host, maxRing: maxRing}
}

// Run records events and serves HTTP until ctx is cancelled (convenience for
// callers that want both in one goroutine).
func (d *Server) Run(ctx context.Context) error {
	d.StartRecorder(ctx)
	return d.Serve(ctx)
}

// StartRecorder subscribes to every bus event and records it (ring + JSONL).
// Call as early as possible to capture login/initial-snapshot events.
func (d *Server) StartRecorder(ctx context.Context) {
	path := d.cfg.LogPath
	if path == "" {
		path = filepath.Join("/tmp/cradle_debug", d.cfg.Username+"_events.jsonl")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		d.log.Warn("debug: cannot create event-log dir; memory only", "dir", filepath.Dir(path), "err", err)
	} else if f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err != nil {
		d.log.Warn("debug: cannot open event log; memory only", "path", path, "err", err)
	} else {
		d.logFile = f
		d.log.Info("debug: recording all events to JSONL", "path", path)
	}

	ch := d.host.Bus().Subscribe("*", 8192)
	go func() {
		// Close the JSONL fd when the recorder stops (ctx cancel or bus close), so a
		// per-host debug surface rebuilt on every restart does not leak an fd each
		// cycle — which would exhaust RLIMIT_NOFILE at fleet scale.
		defer d.closeLog()
		defer d.host.Bus().Unsubscribe("*", ch)
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

// closeLog closes the JSONL event-log handle exactly once.
func (d *Server) closeLog() {
	d.closeOnce.Do(func() {
		if d.logFile != nil {
			_ = d.logFile.Close()
		}
	})
}

func (d *Server) record(ev event.Event) {
	data, err := json.Marshal(ev)
	if err != nil {
		data, _ = json.Marshal(map[string]string{"_marshal_error": err.Error()})
	}
	at := ev.Time()
	if at.IsZero() {
		at = time.Now()
	}
	d.recMu.Lock()
	d.seq++
	rec := eventRecord{Seq: d.seq, At: at, Kind: ev.Kind(), Type: fmt.Sprintf("%T", ev), Data: data}
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

func (d *Server) currentSeq() int {
	d.recMu.Lock()
	defer d.recMu.Unlock()
	return d.seq
}

// Activate builds the persistent interpreter session and starts the idle-event
// pump (so top-level `on` handlers fire while idle). Call once, after
// StartRecorder, before Serve or before mounting Handler(). The pump is guarded
// by evalMu so it never races a /eval.
func (d *Server) Activate(ctx context.Context) {
	d.ctx = ctx
	it := d.host.NewRoutineInterpreter(ctx)
	d.sess = it.NewSession(ctx, "<debug-http>")
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
}

// Handler returns the control-plane mux WITHOUT a listener, so a parent server
// (the cradle) can mount many per-host surfaces under path prefixes on ONE port
// instead of one port per host. Activate must already have been called.
func (d *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", d.handleUI)
	mux.HandleFunc("/help", d.handleHelp)
	mux.HandleFunc("/ws", d.handleWS)
	mux.HandleFunc("/eval", d.handleEval)
	mux.HandleFunc("/script", d.handleScript)
	mux.HandleFunc("/state", d.handleState)
	mux.HandleFunc("/events", d.handleEvents)
	return mux
}

// Serve activates the session/pump and serves the control plane on cfg.Addr until
// ctx is cancelled (cmd/host + legacy-cradle own a port via this path).
// StartRecorder should already be running.
func (d *Server) Serve(ctx context.Context) error {
	d.Activate(ctx)
	srv := &http.Server{Addr: d.cfg.Addr, Handler: d.Handler()}
	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(sctx)
	}()

	d.log.Info("debug HTTP control plane listening",
		"url", "http://"+d.cfg.Addr, "ws", "ws://"+d.cfg.Addr+"/ws")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// ---- WebSocket live stream ------------------------------------------------

// handleWS streams every bus event to the client as JSON text frames, live.
func (d *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer c.CloseNow()

	// Drain inbound frames so a vanished client is detected promptly (and control
	// frames are handled); we only ever write. The returned ctx is cancelled on
	// client close.
	ctx := c.CloseRead(r.Context())
	ch := d.host.Bus().Subscribe("*", 1024)
	defer d.host.Bus().Unsubscribe("*", ch) // release the subscription when the socket closes — no per-connection leak
	for {
		select {
		case <-ctx.Done():
			return
		case <-d.ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			data, _ := json.Marshal(ev)
			at := ev.Time()
			if at.IsZero() {
				at = time.Now()
			}
			frame, _ := json.Marshal(eventRecord{At: at, Kind: ev.Kind(), Type: fmt.Sprintf("%T", ev), Data: data})
			wctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			err := c.Write(wctx, websocket.MessageText, frame)
			cancel()
			if err != nil {
				return
			}
		}
	}
}

// ---- handlers -------------------------------------------------------------

func (d *Server) handleHelp(w http.ResponseWriter, _ *http.Request) {
	io.WriteString(w, `debug-http control plane

  GET  /            browser dashboard: live state + thought stream (WebSocket)
  GET  /ws          WebSocket live stream of every bus event (incl. agent_thought)
  GET  /state       rich JSON snapshot: position, vitals, inventory, npcs, dialog
  GET  /events      ?since=N &kind=agent_thought &limit=500
  POST /eval        body = one DSL line  -> {ok, value, error, events_before}
  POST /script      body = a .routine    -> {ok, kind, value, error}

All commands run against the same persistent interpreter session; every event
is appended to the JSONL event log on disk.
`)
}

func (d *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	user := d.cfg.Username
	if user == "" {
		user = "host"
	}
	io.WriteString(w, strings.ReplaceAll(dashboardHTML, "__USER__", user))
}

type evalResponse struct {
	OK           bool   `json:"ok"`
	Value        string `json:"value,omitempty"`
	IsExpression bool   `json:"is_expression"`
	Error        string `json:"error,omitempty"`
	EventsBefore int    `json:"events_before"`
}

func (d *Server) handleEval(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	line := strings.TrimSpace(string(body))
	if line == "" {
		line = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	if line == "" {
		writeJSON(w, http.StatusBadRequest, evalResponse{Error: "empty command; POST a DSL line or pass ?q="})
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
	writeJSON(w, http.StatusOK, resp)
}

type scriptResponse struct {
	OK    bool   `json:"ok"`
	Kind  string `json:"kind,omitempty"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

func (d *Server) handleScript(w http.ResponseWriter, r *http.Request) {
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
	itx := d.host.NewRoutineInterpreter(rctx)
	res := itx.RunRoutine(rctx, rf.File, nil)

	resp := scriptResponse{OK: res.Err == nil, Kind: res.Kind.String()}
	if res.Err != nil {
		resp.Error = res.Err.Error()
	}
	if res.Value != nil {
		resp.Value = res.Value.Display()
	}
	writeJSON(w, http.StatusOK, resp)
}

// ---- /state snapshot ------------------------------------------------------

type stateSnapshot struct {
	X            int          `json:"x"`
	Y            int          `json:"y"`
	HP           int          `json:"hp"`
	MaxHP        int          `json:"max_hp"`
	Fatigue      int          `json:"fatigue"`
	CombatLevel  int          `json:"combat_level"`
	InvUsed      int          `json:"inventory_used"`
	InvFree      int          `json:"inventory_free"`
	Inventory    []invItem    `json:"inventory"`
	Npcs         []npcSnap    `json:"npcs"`
	Players      []playerSnap `json:"players"`
	DialogOpen   bool         `json:"dialog_open"`
	DialogText   string       `json:"dialog_text,omitempty"`
	DialogOpts   []string     `json:"dialog_options,omitempty"`
	RecentServer []string     `json:"recent_server_messages,omitempty"`
	Skills       []skillSnap  `json:"skills,omitempty"`
	Seq          int          `json:"event_seq"`
}

// skillSnap is one skill's live state for the UI's Skills/XP panel.
type skillSnap struct {
	Name  string `json:"name"`
	Level int    `json:"level"` // base level derived from XP — the "real" level
	Cur   int    `json:"cur"`   // current level (boosted/drained; == level when unmodified)
	XP    int    `json:"xp"`
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

func (d *Server) handleState(w http.ResponseWriter, _ *http.Request) {
	wld := d.host.World()
	f := d.host.Facts()
	snap := stateSnapshot{Seq: d.currentSeq()}

	if wld != nil && wld.Self != nil {
		pos := wld.Self.Position()
		snap.X, snap.Y = pos.X, pos.Y
		snap.HP, snap.MaxHP = wld.Self.HP(), wld.Self.MaxHP()
		snap.Fatigue = wld.Self.FatiguePercent() // 0..100% (cradle/UI render with a trailing %)
		snap.CombatLevel = wld.Self.CombatLevel()
		for id := 0; id <= int(event.SkillThieving); id++ {
			snap.Skills = append(snap.Skills, skillSnap{
				Name:  event.SkillName(event.SkillID(id)),
				Level: wld.Self.SkillMax(id),    // base level (from XP)
				Cur:   wld.Self.SkillLevel(id),  // current (boosted/drained)
				XP:    wld.Self.SkillXP(id),
			})
		}
	}
	if wld != nil && wld.Inventory != nil {
		snap.InvFree = wld.Inventory.FreeSlots()
		for i, s := range wld.Inventory.Slots() {
			snap.Inventory = append(snap.Inventory, invItem{Slot: i, ItemID: s.ItemID, Name: itemName(f, s.ItemID), Amount: s.Amount})
		}
		snap.InvUsed = len(snap.Inventory)
	}
	if wld != nil && wld.Npcs != nil {
		for _, n := range wld.Npcs.All() {
			snap.Npcs = append(snap.Npcs, npcSnap{ServerIndex: n.Index, TypeID: n.TypeID, Name: npcName(f, n.TypeID), X: n.X, Y: n.Y})
		}
	}
	if wld != nil && wld.Players != nil {
		for _, p := range wld.Players.All() {
			snap.Players = append(snap.Players, playerSnap{Index: p.Index, Name: p.Name, X: p.X, Y: p.Y})
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
		for _, m := range wld.Recent.ServerMessages() {
			snap.RecentServer = append(snap.RecentServer, m.Message)
		}
	}
	writeJSON(w, http.StatusOK, snap)
}

func (d *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
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
		if rec.Seq <= since || (kind != "" && rec.Kind != kind) {
			continue
		}
		out = append(out, rec)
		if len(out) >= limit {
			break
		}
	}
	latest := d.seq
	d.recMu.Unlock()
	writeJSON(w, http.StatusOK, map[string]any{"latest_seq": latest, "count": len(out), "events": out})
}

// ---- util -----------------------------------------------------------------

func itemName(f *facts.Facts, id int) string {
	if f != nil {
		if def := f.ItemDef(id); def != nil {
			return def.Name
		}
	}
	return ""
}

func npcName(f *facts.Facts, id int) string {
	if f != nil {
		if def := f.NpcDef(id); def != nil {
			return def.Name
		}
	}
	return ""
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
