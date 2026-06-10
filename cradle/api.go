package cradle

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/debughttp"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/runtime"
)

// errEmptyDirective is returned when an analysis-directive request carries no
// directive text (empty body and no ?q=).
var errEmptyDirective = errors.New("empty directive; POST the operator directive text or pass ?q=")

// API is the cradle's HTTP/JSON control plane. It exposes the registry (list /
// status / pause / resume / stop) and proxies each LIVE host's debug surface
// (state / events / ws / eval / script) under /api/hosts/{name}/debug/ — ONE
// shared server, no per-host ports. Every response is JSON so it serves the web
// UI and headless callers (cradle-ctl, scripts) identically.
type API struct {
	reg *Registry
	log *slog.Logger

	mu   sync.RWMutex          // RLock on the debug-proxy fast path; Lock only in onLive/onExit
	live map[string]*liveDebug // name -> live host's debug surface

	adminMu sync.Mutex                    // guards admins (lazy dial on first persona request)
	admins  map[string]mesapb.AdminClient // mesa addr -> cached admin client (G-7 persona proxy)
}

// perHostRing bounds each host's in-memory event ring in the fleet (the JSONL
// file is the durable record). Far below the 100k standalone default so 200 live
// surfaces don't hold tens of millions of records.
const perHostRing = 4000

// durableEventKinds is the JSONL allowlist: the kinds that survive ring
// eviction and host restarts as the Archive's durable history. The full
// firehose is unbounded at fleet scale, so only these go to disk. debughttp
// compiles the list to an exact-match set — wildcard families (trade_*,
// bank_*, shop_*) are enumerated.
var durableEventKinds = []string{
	// Cognition + narration (the analysis feed).
	"agent_thought", "system_message", "policy_veto", "chat_received",
	// Commerce: trades.
	"trade_request", "trade_opened", "trade_other_offer", "trade_other_accepted",
	"trade_confirm_shown", "trade_closed",
	// Commerce: bank + shop.
	"bank_opened", "bank_slot_update", "bank_closed",
	"shop_opened", "shop_closed",
	// Progression + survival edges.
	"level_up", "death", "target_died", "fatigue_update",
}

// liveDebug holds a running host's debug server plus its StripPrefix-wrapped
// handler (pre-built so per-request routing is allocation-free).
type liveDebug struct {
	srv     *debughttp.Server
	handler http.Handler
}

// NewAPI builds the control plane over reg and registers the lifecycle hooks that
// create/tear down per-host debug surfaces as hosts go live and exit. Call before
// starting hosts.
func NewAPI(reg *Registry, log *slog.Logger) *API {
	if log == nil {
		log = slog.Default()
	}
	a := &API{reg: reg, log: log, live: make(map[string]*liveDebug), admins: make(map[string]mesapb.AdminClient)}
	reg.SetHooks(a.onLive, a.onExit)
	return a
}

func (a *API) onLive(name string, h *runtime.HostHandle) {
	if h == nil || h.Host == nil {
		return
	}
	ctx := h.Ctx
	if ctx == nil {
		// A handle with no ctx is a wiring bug: a Background-bound recorder
		// would outlive the host. Skip the debug surface rather than leak.
		a.log.Warn("cradle: live host handle has nil ctx; debug surface not mounted", "host", name)
		return
	}
	srv := debughttp.New(h.Host, debughttp.Config{
		Username: name,
		MaxRing:  perHostRing,
		// Durable record (the Archive's history); the ring (and /events, /ws)
		// still carries everything.
		JSONLKinds: durableEventKinds,
	}, a.log)
	srv.StartRecorder(ctx) // ring + JSONL; bound to the host ctx so goroutine + fd tear down on stop
	srv.Activate(ctx)      // persistent interp session + idle-event pump
	wrapped := http.StripPrefix("/api/hosts/"+name+"/debug", srv.Handler())
	a.mu.Lock()
	a.live[name] = &liveDebug{srv: srv, handler: wrapped}
	a.mu.Unlock()
}

func (a *API) onExit(name string) {
	a.mu.Lock()
	delete(a.live, name)
	a.mu.Unlock()
}

func (a *API) liveHandler(name string) http.Handler {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if lh := a.live[name]; lh != nil {
		return lh.handler
	}
	return nil
}

// Handler builds the API mux (Go method + wildcard routing).
func (a *API) Handler() http.Handler {
	mux := http.NewServeMux()
	// pprof for the whole fleet process (heap/goroutine profiles are the soak's
	// leak instrumentation; the goroutine profile names any leaked subscriber
	// by file:line).
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	mux.HandleFunc("GET /{$}", a.handleIndex) // single-page web UI at /
	mux.Handle("GET /v2/", v2Handler())       // v2 SPA (embedded web/v2)
	mux.HandleFunc("GET /api/hosts", a.handleList)
	mux.HandleFunc("GET /api/hosts/{name}", a.handleGet)
	// Persona proxy (mesa Admin.GetPersona, admin-credentialed): 501 when the
	// proxy can't serve (no admin token, no mesa, mesa down) — the UI hides
	// the persona pane on 501.
	mux.HandleFunc("GET /api/hosts/{name}/persona", a.handlePersona)
	mux.HandleFunc("POST /api/hosts/{name}/pause", a.handlePause)
	mux.HandleFunc("POST /api/hosts/{name}/resume", a.handleResume)
	mux.HandleFunc("POST /api/hosts/{name}/stop", a.handleStop)
	mux.HandleFunc("POST /api/hosts/{name}/restart", a.handleRestart)
	// Operator-override ANALYSIS mode (control-plane-authoritative; no name match):
	// enter/exit toggle, post a directive, and read the current analysis state.
	mux.HandleFunc("POST /api/hosts/{name}/analysis/enter", a.handleAnalysisEnter)
	mux.HandleFunc("POST /api/hosts/{name}/analysis/exit", a.handleAnalysisExit)
	mux.HandleFunc("POST /api/hosts/{name}/analysis/directive", a.handleAnalysisDirective)
	mux.HandleFunc("GET /api/hosts/{name}/analysis", a.handleAnalysisGet)
	// Per-host debug surface (state/events/ws/eval/script), proxied to the live
	// host's debughttp handler under this prefix.
	mux.HandleFunc("/api/hosts/{name}/debug/", a.handleDebug)
	return mux
}

func (a *API) handleList(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, a.reg.List())
}

func (a *API) handleGet(w http.ResponseWriter, r *http.Request) {
	s, err := a.reg.Get(r.PathValue("name"))
	if err != nil {
		writeJSON(w, http.StatusNotFound, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (a *API) handlePause(w http.ResponseWriter, r *http.Request) {
	a.act(w, a.reg.Pause(r.PathValue("name")))
}
func (a *API) handleResume(w http.ResponseWriter, r *http.Request) {
	a.act(w, a.reg.Resume(r.PathValue("name")))
}
func (a *API) handleRestart(w http.ResponseWriter, r *http.Request) {
	a.act(w, a.reg.Restart(r.PathValue("name")))
}

func (a *API) handleStop(w http.ResponseWriter, r *http.Request) {
	a.act(w, a.reg.Stop(r.PathValue("name")))
}

func (a *API) act(w http.ResponseWriter, err error) {
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// handleAnalysisEnter / handleAnalysisExit toggle operator-override analysis mode.
// They return the resulting AnalysisResult (kind=control) so the UI can echo the
// state line and flip the toggle without a second status fetch.
func (a *API) handleAnalysisEnter(w http.ResponseWriter, r *http.Request) {
	res, err := a.reg.EnterAnalysis(r.PathValue("name"))
	a.analysisResult(w, res, err)
}

func (a *API) handleAnalysisExit(w http.ResponseWriter, r *http.Request) {
	res, err := a.reg.ExitAnalysis(r.PathValue("name"))
	a.analysisResult(w, res, err)
}

// handleAnalysisDirective interprets one operator directive (the request body is
// the raw directive text) and returns the structured {kind, text, dsl, executed,
// active} verdict. Auto-enters analysis mode if off (control-plane-authoritative).
func (a *API) handleAnalysisDirective(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	directive := strings.TrimSpace(string(body))
	if directive == "" {
		directive = strings.TrimSpace(r.URL.Query().Get("q"))
	}
	if directive == "" {
		writeJSON(w, http.StatusBadRequest, errBody(errEmptyDirective))
		return
	}
	res, err := a.reg.AnalyzeDirective(r.Context(), r.PathValue("name"), directive)
	a.analysisResult(w, res, err)
}

// handleAnalysisGet reports the current analysis state + the most-recent verdict
// (the UI feed/state poll). Always returns a well-formed body.
func (a *API) handleAnalysisGet(w http.ResponseWriter, r *http.Request) {
	last, active, err := a.reg.AnalysisStatus(r.PathValue("name"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"active": active, "last": last})
}

// analysisResult writes an AnalysisResult as JSON (HTTP 200) or maps a control
// error (host not running) to a 400 error body.
func (a *API) analysisResult(w http.ResponseWriter, res runtime.AnalysisResult, err error) {
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err))
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (a *API) handleDebug(w http.ResponseWriter, r *http.Request) {
	h := a.liveHandler(r.PathValue("name"))
	if h == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "host not live: " + r.PathValue("name")})
		return
	}
	h.ServeHTTP(w, r)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}

func errBody(err error) map[string]string { return map[string]string{"error": err.Error()} }
