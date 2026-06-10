package cradle

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/gen0cide/westworld/cradle/hostcfg"
	"github.com/gen0cide/westworld/runtime"
)

// Status is a managed host's lifecycle state.
type Status string

const (
	StatusStarting   Status = "starting"   // bringing up (connect/genesis) — no handle yet
	StatusRunning    Status = "running"    // conductor turn loop live
	StatusPaused     Status = "paused"     // turn loop gated (pause)
	StatusStopped    Status = "stopped"    // intentionally stopped or director finished
	StatusCrashed    Status = "crashed"    // exited unexpectedly; held (supervision=hold)
	StatusRestarting Status = "restarting" // crashed; waiting out backoff before relaunch
)

const restartBackoffMax = 30 * time.Second

// HostStatus is a JSON-serialisable snapshot of one managed host (the unit the
// control API and CLI render).
type HostStatus struct {
	Name       string    `json:"name"`
	Status     Status    `json:"status"`
	Mesa       string    `json:"mesa,omitempty"`
	Goal       string    `json:"goal,omitempty"`
	Autonomous bool      `json:"autonomous"`
	Restarts   int       `json:"restarts"`
	StartedAt  time.Time `json:"started_at"`
	Err        string    `json:"err,omitempty"`

	// CurrentRoutine is the label of the intent the conductor is running right
	// now (what she's doing). Empty when not running / before the first turn.
	CurrentRoutine string `json:"current_routine,omitempty"`

	// CurrentRoutineSource is the inline DSL source of the routine she's running
	// now (the live Routine panel renders it line-numbered). Empty for a
	// file/library routine with no inline source — the panel then shows just the
	// label. CurrentLine is the source line the interpreter is on right now
	// (1-based; 0 before the first statement) — the panel highlights it live.
	CurrentRoutineSource string `json:"current_routine_source,omitempty"`
	CurrentLine          int    `json:"current_line,omitempty"`

	// LineTrace is the recent run of source lines the interpreter executed (in
	// order), and LineSeq is the monotonic count of statements executed. The UI
	// replays the trace — stepping the highlight through each line with a slight
	// delay — so fast statements between waits are visible, not skipped.
	LineTrace []int `json:"line_trace,omitempty"`
	LineSeq   int   `json:"line_seq,omitempty"`

	// Live world snapshot, present once the host is running.
	Live  bool `json:"live"`
	X     int  `json:"x,omitempty"`
	Y     int  `json:"y,omitempty"`
	HP    int  `json:"hp,omitempty"`
	MaxHP int  `json:"max_hp,omitempty"`

	// Analysis is true while the host is under operator-override ANALYSIS mode
	// (conductor frozen, memory suspended, full bypass). Surfaced here so the UI
	// status poll lights the Analysis badge/toggle with no extra endpoint.
	Analysis bool `json:"analysis,omitempty"`
}

// managedHost tracks one host the registry supervises.
type managedHost struct {
	spec hostcfg.Host

	// parent is the supervision context Start was given; kept so Restart can
	// respawn the supervisor for a host whose goroutine has already exited
	// (stopped/crashed). Cancelling it still tears everything down.
	parent context.Context

	mu         sync.Mutex
	status     Status
	lastErr    error
	handle     *runtime.HostHandle
	cancel     context.CancelFunc
	stopReq    bool
	restartReq bool // operator bounce: relaunch without backoff or a restart count
	alive      bool // a supervise goroutine currently owns this host
	startedAt  time.Time
	restarts   int
}

func (mh *managedHost) setStatus(s Status) {
	mh.mu.Lock()
	mh.status = s
	mh.mu.Unlock()
}

func (mh *managedHost) snapshot() HostStatus {
	mh.mu.Lock()
	defer mh.mu.Unlock()
	hs := HostStatus{
		Name:       mh.spec.Name,
		Status:     mh.status,
		Mesa:       mh.spec.Mesa,
		Goal:       mh.spec.Goal,
		Autonomous: mh.spec.Autonomous(),
		Restarts:   mh.restarts,
		StartedAt:  mh.startedAt,
	}
	if mh.lastErr != nil {
		hs.Err = mh.lastErr.Error()
	}
	if mh.handle != nil && mh.handle.Host != nil {
		if w := mh.handle.Host.World(); w != nil && w.Self != nil {
			hs.Live = true
			p := w.Self.Position()
			hs.X, hs.Y = int(p.X), int(p.Y)
			hs.HP, hs.MaxHP = int(w.Self.HP()), int(w.Self.MaxHP())
		}
	}
	if mh.handle != nil && mh.handle.Conductor != nil {
		hs.CurrentRoutine = mh.handle.Conductor.CurrentIntent()
		hs.CurrentRoutineSource = mh.handle.Conductor.CurrentRoutineSource()
		hs.CurrentLine = mh.handle.Conductor.CurrentLine()
		hs.LineTrace, hs.LineSeq = mh.handle.Conductor.LineTrace()
	}
	if mh.handle != nil && mh.handle.Host != nil {
		hs.Analysis = mh.handle.Host.AnalysisActive()
	}
	return hs
}

// runFunc is the host runner; runtime.RunHost in production, a fake in tests.
type runFunc func(context.Context, runtime.HostConfig, runtime.SharedDeps) error

// Registry runs and supervises a fleet of hosts in one process over shared deps.
type Registry struct {
	deps runtime.SharedDeps
	log  *slog.Logger
	run  runFunc

	backoffBase time.Duration // initial restart backoff (doubles to restartBackoffMax)

	// Lifecycle hooks (optional). onLive fires when a host becomes live (handle
	// available); onExit fires when its run returns (each time, including between
	// restarts). The API layer uses them to create/tear down the per-host debug
	// surface. Set once before starting hosts.
	onLive func(name string, h *runtime.HostHandle)
	onExit func(name string)

	mu    sync.Mutex
	hosts map[string]*managedHost
	wg    sync.WaitGroup
}

// SetHooks registers lifecycle callbacks. Call before starting any host.
func (r *Registry) SetHooks(onLive func(string, *runtime.HostHandle), onExit func(string)) {
	r.onLive = onLive
	r.onExit = onExit
}

// NewRegistry builds a registry over shared process deps.
func NewRegistry(deps runtime.SharedDeps, log *slog.Logger) *Registry {
	if log == nil {
		log = slog.Default()
	}
	return &Registry{deps: deps, log: log, run: runtime.RunHost, backoffBase: time.Second, hosts: make(map[string]*managedHost)}
}

// Start launches a host from its spec under parent ctx. It validates the spec and
// resolves the password up front (a misconfigured host fails fast, before any
// goroutine spins). The host runs on its own goroutine; supervision is applied on
// exit per the spec's policy.
func (r *Registry) Start(parent context.Context, spec hostcfg.Host) error {
	if err := spec.Validate(); err != nil {
		return err
	}
	pw, err := spec.ResolvePassword()
	if err != nil {
		return err
	}

	r.mu.Lock()
	if _, exists := r.hosts[spec.Name]; exists {
		r.mu.Unlock()
		return fmt.Errorf("host %q is already managed", spec.Name)
	}
	mh := &managedHost{spec: spec, parent: parent, status: StatusStarting, startedAt: time.Now(), alive: true}
	r.hosts[spec.Name] = mh
	r.mu.Unlock()

	r.wg.Add(1)
	go r.supervise(parent, mh, pw)
	return nil
}

// StartAll starts every host in the set. It returns the first start error (e.g. a
// missing password) but leaves already-started hosts running.
func (r *Registry) StartAll(parent context.Context, hosts []hostcfg.Host) error {
	for _, h := range hosts {
		if err := r.Start(parent, h); err != nil {
			return fmt.Errorf("start %q: %w", h.Name, err)
		}
	}
	return nil
}

// supervise runs one host and applies its restart policy on exit.
func (r *Registry) supervise(parent context.Context, mh *managedHost, pw string) {
	defer r.wg.Done()
	defer func() {
		mh.mu.Lock()
		mh.alive = false
		mh.mu.Unlock()
	}()
	log := r.log.With("host", mh.spec.Name)
	policy := mh.spec.SupervisionPolicy()
	backoff := r.backoffBase
	if backoff <= 0 {
		backoff = time.Second
	}

	for {
		mh.mu.Lock()
		if mh.stopReq {
			mh.status = StatusStopped
			mh.mu.Unlock()
			return
		}
		childCtx, cancel := context.WithCancel(parent)
		mh.cancel = cancel
		mh.restartReq = false // launching IS the bounce; don't re-bounce the new run later
		mh.status = StatusStarting
		mh.startedAt = time.Now()
		mh.handle = nil
		mh.lastErr = nil
		mh.mu.Unlock()

		cfg := buildCfg(mh.spec, pw, r.log)
		cfg.OnReady = func(h *runtime.HostHandle) {
			mh.mu.Lock()
			mh.handle = h
			if mh.status == StatusStarting {
				mh.status = StatusRunning
			}
			mh.mu.Unlock()
			log.Info("host running")
			if r.onLive != nil {
				r.onLive(mh.spec.Name, h)
			}
		}

		err := r.run(childCtx, cfg, r.deps)
		cancel()
		if r.onExit != nil {
			r.onExit(mh.spec.Name)
		}

		mh.mu.Lock()
		stopReq := mh.stopReq
		restartReq := mh.restartReq
		mh.restartReq = false
		mh.handle = nil
		if err != nil {
			mh.lastErr = err
		}
		mh.mu.Unlock()

		switch {
		case stopReq || parent.Err() != nil:
			mh.setStatus(StatusStopped)
			log.Info("host stopped")
			return
		case restartReq:
			// Operator bounce: relaunch immediately — no backoff, no restart
			// count (those are crash bookkeeping), regardless of the host's
			// supervision policy. A fresh login means a fresh genesis.
			log.Info("host bounced by operator; relaunching")
			backoff = r.backoffBase
			continue
		case err != nil && policy == hostcfg.SuperviseRestart:
			mh.setStatus(StatusRestarting)
			log.Warn("host crashed; restarting", "err", err, "backoff", backoff)
			select {
			case <-parent.Done():
				mh.setStatus(StatusStopped)
				return
			case <-time.After(backoff):
			}
			mh.mu.Lock()
			mh.restarts++
			mh.mu.Unlock()
			backoff = minDur(backoff*2, restartBackoffMax)
			continue
		case err != nil:
			mh.setStatus(StatusCrashed)
			log.Error("host crashed; holding (supervision=hold)", "err", err)
			return
		default:
			mh.setStatus(StatusStopped)
			log.Info("host finished (director done)")
			return
		}
	}
}

// Stop cancels a host's context and prevents any further restart.
func (r *Registry) Stop(name string) error {
	mh, err := r.get(name)
	if err != nil {
		return err
	}
	mh.mu.Lock()
	mh.stopReq = true
	cancel := mh.cancel
	mh.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	return nil
}

// Restart bounces a host: tears down its current run (disconnect) and brings it
// back up with a fresh login + genesis. Works on any state — a live host is
// cancelled and relaunched by its supervisor (no backoff, no restart count, and
// independent of its supervision policy); a stopped/crashed host gets a new
// supervisor goroutine. This is the per-host alternative to bouncing the whole
// cradle.
func (r *Registry) Restart(name string) error {
	mh, err := r.get(name)
	if err != nil {
		return err
	}
	mh.mu.Lock()
	mh.stopReq = false // a previous Stop must not veto the relaunch
	if mh.alive {
		// Supervisor owns the host: flag the bounce and cut the run.
		mh.restartReq = true
		mh.status = StatusRestarting
		cancel := mh.cancel
		mh.mu.Unlock()
		if cancel != nil {
			cancel()
		}
		return nil
	}
	// Supervisor exited (stopped / crashed / finished): respawn it. Claim the
	// host under the lock so a concurrent Restart can't double-spawn.
	mh.alive = true
	mh.restartReq = false
	mh.status = StatusStarting
	mh.startedAt = time.Now()
	mh.lastErr = nil
	parent := mh.parent
	spec := mh.spec
	mh.mu.Unlock()

	pw, err := spec.ResolvePassword()
	if err != nil {
		mh.mu.Lock()
		mh.alive = false
		mh.status = StatusCrashed
		mh.lastErr = err
		mh.mu.Unlock()
		return err
	}
	r.wg.Add(1)
	go r.supervise(parent, mh, pw)
	return nil
}

// Pause freezes a running host's turn loop (between turns). It is an error if the
// host is not running. The terminal-state check + status write share one critical
// section with the handle read, so a host that exits concurrently can't be left
// wedged as "paused".
func (r *Registry) Pause(name string) error {
	mh, err := r.get(name)
	if err != nil {
		return err
	}
	mh.mu.Lock()
	defer mh.mu.Unlock()
	if mh.stopReq || mh.status == StatusStopped || mh.status == StatusCrashed || mh.handle == nil || mh.handle.Conductor == nil {
		return fmt.Errorf("host %q is not running", name)
	}
	mh.handle.Conductor.Pause()
	mh.status = StatusPaused
	return nil
}

// Resume releases a paused host.
func (r *Registry) Resume(name string) error {
	mh, err := r.get(name)
	if err != nil {
		return err
	}
	mh.mu.Lock()
	defer mh.mu.Unlock()
	if mh.stopReq || mh.status == StatusStopped || mh.status == StatusCrashed || mh.handle == nil || mh.handle.Conductor == nil {
		return fmt.Errorf("host %q is not running", name)
	}
	mh.handle.Conductor.Resume()
	mh.status = StatusRunning
	return nil
}

// EnterAnalysis puts a running host into operator-override ANALYSIS mode
// (control-plane-authoritative — no operator name-match, unlike the in-game
// trigger). It freezes the conductor and suspends memory writes. Error if the
// host is not running.
func (r *Registry) EnterAnalysis(name string) (runtime.AnalysisResult, error) {
	h, err := r.runningHandle(name)
	if err != nil {
		return runtime.AnalysisResult{}, err
	}
	return h.Host.EnterAnalysis(), nil
}

// ExitAnalysis leaves analysis mode and resumes autonomy.
func (r *Registry) ExitAnalysis(name string) (runtime.AnalysisResult, error) {
	h, err := r.runningHandle(name)
	if err != nil {
		return runtime.AnalysisResult{}, err
	}
	return h.Host.ExitAnalysis(), nil
}

// AnalyzeDirective interprets one operator directive against a running host and
// returns the structured verdict. The host-side interpreter does I/O (mesa
// Chat/Act, the DSL eval), so the handle is read under the registry lock then
// released — the interpreter call NEVER holds the lock.
func (r *Registry) AnalyzeDirective(ctx context.Context, name, directive string) (runtime.AnalysisResult, error) {
	h, err := r.runningHandle(name)
	if err != nil {
		return runtime.AnalysisResult{}, err
	}
	return h.Host.Analyze(ctx, directive), nil
}

// AnalysisStatus reports a running host's most-recent verdict + active flag.
func (r *Registry) AnalysisStatus(name string) (last *runtime.AnalysisResult, active bool, err error) {
	h, herr := r.runningHandle(name)
	if herr != nil {
		return nil, false, herr
	}
	last, active = h.Host.AnalysisState()
	return last, active, nil
}

// runningHandle returns the live handle for a running host, or an error. It reads
// the handle under the per-host lock and releases it before returning, so callers
// can do I/O against the handle without holding the registry lock.
func (r *Registry) runningHandle(name string) (*runtime.HostHandle, error) {
	mh, err := r.get(name)
	if err != nil {
		return nil, err
	}
	mh.mu.Lock()
	defer mh.mu.Unlock()
	if mh.stopReq || mh.status == StatusStopped || mh.status == StatusCrashed || mh.handle == nil || mh.handle.Host == nil {
		return nil, fmt.Errorf("host %q is not running", name)
	}
	return mh.handle, nil
}

// List returns a snapshot of every managed host, sorted by name.
func (r *Registry) List() []HostStatus {
	r.mu.Lock()
	mhs := make([]*managedHost, 0, len(r.hosts))
	for _, mh := range r.hosts {
		mhs = append(mhs, mh)
	}
	r.mu.Unlock()
	out := make([]HostStatus, 0, len(mhs))
	for _, mh := range mhs {
		out = append(out, mh.snapshot())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns one host's snapshot.
func (r *Registry) Get(name string) (HostStatus, error) {
	mh, err := r.get(name)
	if err != nil {
		return HostStatus{}, err
	}
	return mh.snapshot(), nil
}

// Handle returns the live host handle for a running host (used by the control API
// to read world state / drive eval). Present only while running.
func (r *Registry) Handle(name string) (*runtime.HostHandle, bool) {
	mh, err := r.get(name)
	if err != nil {
		return nil, false
	}
	mh.mu.Lock()
	defer mh.mu.Unlock()
	return mh.handle, mh.handle != nil
}

// StopAll requests every managed host to stop (without waiting).
func (r *Registry) StopAll() {
	r.mu.Lock()
	mhs := make([]*managedHost, 0, len(r.hosts))
	for _, mh := range r.hosts {
		mhs = append(mhs, mh)
	}
	r.mu.Unlock()
	for _, mh := range mhs {
		mh.mu.Lock()
		mh.stopReq = true
		cancel := mh.cancel
		mh.mu.Unlock()
		if cancel != nil {
			cancel()
		}
	}
}

// Wait blocks until every supervised host goroutine has exited (after the parent
// ctx is cancelled or StopAll is called).
func (r *Registry) Wait() { r.wg.Wait() }

func (r *Registry) get(name string) (*managedHost, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	mh, ok := r.hosts[name]
	if !ok {
		return nil, fmt.Errorf("no such host %q", name)
	}
	return mh, nil
}

// buildCfg maps a host spec to a runtime.HostConfig with the resolved password +
// the process logger (RunHost derives a per-host child logger).
func buildCfg(spec hostcfg.Host, pw string, log *slog.Logger) runtime.HostConfig {
	cfg := spec.ToHostConfig(pw)
	cfg.Logger = log
	cfg.Headless = true // daemon: never drop to a stdin REPL
	return cfg
}

func minDur(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
