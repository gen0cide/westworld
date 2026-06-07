package cradle

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/gen0cide/westworld/debughttp"
	"github.com/gen0cide/westworld/runtime"
)

// API is the cradle's HTTP/JSON control plane. It exposes the registry (list /
// status / pause / resume / stop) and proxies each LIVE host's debug surface
// (state / events / ws / eval / script) under /api/hosts/{name}/debug/ — ONE
// shared server, no per-host ports. Every response is JSON so it serves the web
// UI and headless callers (cradle-ctl, scripts) identically.
type API struct {
	reg *Registry
	log *slog.Logger

	mu   sync.RWMutex // RLock on the debug-proxy fast path; Lock only in onLive/onExit
	live map[string]*liveDebug // name -> live host's debug surface
}

// perHostRing bounds each host's in-memory event ring in the fleet (the JSONL
// file is the durable record). Far below the 100k standalone default so 200 live
// surfaces don't hold tens of millions of records.
const perHostRing = 4000

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
	a := &API{reg: reg, log: log, live: make(map[string]*liveDebug)}
	reg.SetHooks(a.onLive, a.onExit)
	return a
}

func (a *API) onLive(name string, h *runtime.HostHandle) {
	if h == nil || h.Host == nil {
		return
	}
	ctx := h.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	srv := debughttp.New(h.Host, debughttp.Config{Username: name, MaxRing: perHostRing}, a.log)
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
	mux.HandleFunc("GET /{$}", a.handleIndex) // single-page web UI at /
	mux.HandleFunc("GET /api/hosts", a.handleList)
	mux.HandleFunc("GET /api/hosts/{name}", a.handleGet)
	mux.HandleFunc("POST /api/hosts/{name}/pause", a.handlePause)
	mux.HandleFunc("POST /api/hosts/{name}/resume", a.handleResume)
	mux.HandleFunc("POST /api/hosts/{name}/stop", a.handleStop)
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
