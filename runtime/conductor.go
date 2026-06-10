package runtime

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/hostkv"
)

// The conductor is the host's autonomy spine: the driver that decides "what to
// do now" by repeatedly selecting an Intent, running it to completion, and
// observing the result before choosing the next.
//
// It is deliberately NOT a tick loop (see docs decision-cognitive-loop: "there
// is no tick"). Host.Run is the frame pump — it keeps the world mirror live on
// its own goroutine. The conductor runs CONCURRENTLY with Host.Run, on a
// separate goroutine, and drives the interpreter. The canonical composition is:
//
//	go host.Run(ctx)            // frame pump: must be started first
//	conductor.Run(ctx)          // routine-selection loop
//
// Routines run on the conductor goroutine and read the world mirror that the
// Run goroutine keeps fresh; between turns the director observes that same
// always-current mirror to pick the next Intent.

// Intent is one unit of work for the conductor: either a routine FILE to run
// (RoutinePath) or an inline DSL SOURCE to compile and run (Source + Name, used
// by the mesa Act planner for freshly-authored routines), plus positional args
// and a human label for logs and telemetry.
type Intent struct {
	Label       string
	RoutinePath string
	Args        []interp.Value

	// Source, when non-empty, is DSL compiled and run directly (no file). Name
	// is its logical name for logs/parse errors. Source takes precedence over
	// RoutinePath.
	Source string
	Name   string

	// OneShot marks a single-use intent (a direct action, an idle wait, a dialog
	// answer) that must NEVER be cached + replayed by the cheap loop — only
	// repeatable grinds are learnable. Without this, a one-off "say hi" / no-op
	// gets promoted and replayed forever.
	OneShot bool
}

// Outcome is the recorded result of running one Intent. It is threaded back
// into the director's next decision so a director can react (retry on error,
// advance on completion, etc.).
type Outcome struct {
	Intent   Intent
	Kind     interp.ResultKind
	Value    interp.Value
	Err      *interp.RuntimeError
	Duration time.Duration

	// BudgetExpired marks a turn that ended because the per-turn budget
	// (TurnTimeout) ran out — NOT because the routine itself failed. A long
	// grind that simply outlives its turn is still working; directors must not
	// count this as a failure (soak retro #5: healthy grinds read as FAILED,
	// then the BLOCKED escalation ordered hosts to abandon working approaches).
	BudgetExpired bool
}

// OK reports whether the turn ended without an error or abort.
func (o Outcome) OK() bool {
	return o.Kind == interp.ResultCompleted || o.Kind == interp.ResultReturned
}

// Director chooses what a host does next. It is the seam where a fixed
// sequence, a goal system, or (later) a Pearl/LLM planner plug in.
type Director interface {
	// Next returns the next Intent to run and a bool that is false when the
	// director is done (the conductor then stops). h is the live host so a
	// director can observe world state; last is the previous turn's Outcome
	// (the zero Outcome on the first call).
	Next(ctx context.Context, h *Host, last Outcome) (Intent, bool)
}

// DirectorFunc adapts a plain function to the Director interface.
type DirectorFunc func(ctx context.Context, h *Host, last Outcome) (Intent, bool)

// Next implements Director.
func (f DirectorFunc) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	return f(ctx, h, last)
}

// Sequence returns a Director that yields each intent once, in order, then
// stops. Useful for scripted runs (e.g. a fixed tutorial walkthrough).
func Sequence(intents ...Intent) Director {
	i := 0
	return DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		if i >= len(intents) {
			return Intent{}, false
		}
		in := intents[i]
		i++
		return in, true
	})
}

// Loop returns a Director that yields the same intent forever (until the
// conductor's context is cancelled). Useful for "do this until stopped"
// behavior like wandering or skilling.
func Loop(intent Intent) Director {
	return DirectorFunc(func(_ context.Context, _ *Host, _ Outcome) (Intent, bool) {
		return intent, true
	})
}

// ConductorOptions configures a Conductor. All fields are optional; sensible
// defaults are applied by NewConductor.
type ConductorOptions struct {
	Director Director // required in practice; a nil director stops immediately

	// TurnTimeout bounds a single routine's execution. A routine that runs
	// longer is cancelled and its Outcome marked BudgetExpired — running out
	// of turn time is a scheduling event, not a routine failure. Default 2m.
	TurnTimeout time.Duration

	// Settle is the pause between turns. It prevents hot-spin when routines
	// return quickly and gives the world mirror a beat to reflect the last
	// turn's effects before the director decides again. Default 500ms.
	Settle time.Duration

	// Store and Scratch are the host's local K/V. When nil, the conductor
	// uses an in-memory Store and a 256-entry Scratch so it always has
	// somewhere to record progress.
	Store   *hostkv.Store
	Scratch *hostkv.Scratch

	Logger *slog.Logger

	// Detours enables the interrupt/detour stack (survival preemption today): the
	// running routine is suspendable and a critical interrupt parks it, runs a
	// detour, and resumes it. Off for fixed scripted / load-drone hosts. Default off.
	Detours bool
}

// Conductor drives a host's autonomous turn loop.
type Conductor struct {
	host        *Host
	director    Director
	log         *slog.Logger
	store       *hostkv.Store
	scratch     *hostkv.Scratch
	turnTimeout time.Duration
	settle      time.Duration

	// execute runs one intent and returns its outcome. It is a field so tests
	// can substitute a fake runner and exercise the loop without a live server.
	execute func(ctx context.Context, in Intent) Outcome

	// detours enables the interrupt/detour stack: the running routine becomes a
	// suspendable Coro, and a higher-tier interrupt (survival) can PARK it, run a
	// detour, then RESUME it where it left off. interrupts carries detour
	// requests from the arbiter goroutine. nil/false when detours are disabled
	// (fixed scripted / load-drone mode), where execute is the plain blocking path.
	detours    bool
	interrupts chan detourReq

	// Reactive (chat-driven) interrupt coalescing: while a turn runs, "reactive"
	// tier requests fold into pendingReactive — LATEST per speaker — instead of
	// queueing individually on the channel; the channel then only carries a
	// wakeup marker, and executeWithDetours services the whole batch under ONE
	// park/resume. Protects grind turns from social spam (soak retro #9/#12:
	// 36-60% of all decisions on social hosts were chat-driven parks).
	// pendingOrder keeps first-arrival order so service is deterministic.
	// Guarded by reactMu. See coalesceReactive / takePendingReactive (detour.go).
	reactMu         sync.Mutex
	pendingReactive map[string]detourReq
	pendingOrder    []string

	// pause gate: when paused, Run blocks at the turn boundary (between routines)
	// until Resume or ctx cancel — lets the cradle freeze a host live without
	// tearing it down. Pause ALSO interrupts the in-flight routine via turnCancel
	// so a long-running turn (e.g. a 5-minute mining loop) stops promptly rather
	// than only at the next boundary. Analysis mode rides the same Pause path.
	pauseMu  sync.Mutex
	paused   bool
	resumeCh chan struct{}
	// turnCancel cancels the currently-running turn's context, so Pause() can
	// interrupt a long routine immediately. Set by beginTurn while a turn runs,
	// nil between turns. Guarded by pauseMu.
	turnCancel context.CancelFunc

	// curMu guards curIntent / curSource / curLine: the label of the routine
	// currently executing, its inline DSL source (when the intent carries one),
	// and the source line the interpreter is on right now. Surfaced via
	// CurrentIntent() / CurrentRoutineSource() / CurrentLine() so the cradle can
	// show "what she's doing now" and a live, line-highlighted Routine panel.
	curMu     sync.Mutex
	curIntent string
	curSource string
	curLine   int

	// traceMu guards the line-execution trace: a bounded, ordered ring of the
	// source lines the interpreter has executed, plus a monotonic sequence. The
	// per-statement OnStmt hook records EVERY line (curLine is only the latest),
	// so the cradle UI can REPLAY the path — stepping the highlight through each
	// line with a slight delay — instead of the 350ms poll only ever catching the
	// lines slow enough (waits) to still be current. Surfaced via LineTrace().
	traceMu   sync.Mutex
	lineTrace []int
	lineSeq   int
}

// maxLineTrace bounds the per-routine line-execution trace ring the UI replays.
const maxLineTrace = 64

// NewConductor builds a conductor for host h with the given options.
func NewConductor(h *Host, opts ConductorOptions) *Conductor {
	c := &Conductor{
		host:        h,
		director:    opts.Director,
		log:         opts.Logger,
		store:       opts.Store,
		scratch:     opts.Scratch,
		turnTimeout: opts.TurnTimeout,
		settle:      opts.Settle,
	}
	if c.log == nil {
		if h != nil && h.log != nil {
			c.log = h.log
		} else {
			c.log = slog.Default()
		}
	}
	if c.store == nil {
		c.store = hostkv.NewMemory()
	}
	if c.scratch == nil {
		c.scratch = hostkv.NewScratch(256)
	}
	if c.turnTimeout <= 0 {
		c.turnTimeout = 2 * time.Minute
	}
	if c.settle < 0 {
		c.settle = 0
	} else if c.settle == 0 {
		c.settle = 500 * time.Millisecond
	}
	if opts.Detours {
		c.detours = true
		c.interrupts = make(chan detourReq, 4)
		c.execute = c.executeWithDetours
	} else {
		c.execute = c.executeRoutine
	}
	c.resumeCh = make(chan struct{})
	// Track the current source line for the cradle's live Routine panel. The
	// hook fires once per statement on the routine goroutine; keep it a bare
	// mutex-guarded store (O(1), non-blocking). Installed on the host so it
	// reaches both the plain and detour interpreter paths via
	// NewRoutineInterpreter. Guard against a nil host (test conductors).
	if h != nil {
		h.OnStmt = func(line int) {
			c.curMu.Lock()
			c.curLine = line
			c.curMu.Unlock()
			c.recordLine(line)
		}
	}
	return c
}

// Pause freezes the turn loop AND interrupts the currently-running routine: it
// sets the pause flag (so no new turn starts until Resume) and cancels the
// in-flight turn's context so a long routine (a mining loop, travel) stops
// promptly instead of only at the next turn boundary. Safe from any goroutine —
// the cradle's control API and EnterAnalysis both call it.
func (c *Conductor) Pause() {
	c.pauseMu.Lock()
	c.paused = true
	cancel := c.turnCancel
	c.pauseMu.Unlock()
	if cancel != nil {
		cancel() // interrupt the in-flight turn at its next ctx checkpoint / wait
	}
}

// Resume releases a paused turn loop. A no-op if not paused.
func (c *Conductor) Resume() {
	c.pauseMu.Lock()
	wasPaused := c.paused
	if c.paused {
		c.paused = false
		close(c.resumeCh)                // wake any waiter
		c.resumeCh = make(chan struct{}) // re-arm for the next pause
	}
	c.pauseMu.Unlock()
	if wasPaused {
		// Interrupts queued during a freeze describe a situation the operator
		// just looked at (and likely handled); acting on them after resume
		// replays stale detours.
		c.drainInterrupts()
	}
}

// Paused reports whether the loop is currently paused.
func (c *Conductor) Paused() bool {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()
	return c.paused
}

// CurrentIntent returns the label of the routine currently executing (the most
// recent intent the turn loop began). Empty before the first turn. Safe to call
// from any goroutine (e.g. the cradle's status snapshot).
func (c *Conductor) CurrentIntent() string {
	c.curMu.Lock()
	defer c.curMu.Unlock()
	return c.curIntent
}

// CurrentRoutineSource returns the inline DSL source of the routine currently
// executing, or "" for a file/library routine that carries no inline Source
// (the cradle panel then shows only the label). Safe to call from any goroutine.
func (c *Conductor) CurrentRoutineSource() string {
	c.curMu.Lock()
	defer c.curMu.Unlock()
	return c.curSource
}

// CurrentLine returns the source line the interpreter is executing right now
// (1-based; 0 before the first statement of a turn). It is the cheap
// current-line tracker behind the cradle's live, line-highlighted Routine
// panel. Safe to call from any goroutine.
func (c *Conductor) CurrentLine() int {
	c.curMu.Lock()
	defer c.curMu.Unlock()
	return c.curLine
}

// recordLine appends one executed source line to the bounded trace ring and
// bumps the monotonic sequence. Called from the per-statement OnStmt hook on the
// routine goroutine, so it stays O(1) and non-blocking.
func (c *Conductor) recordLine(line int) {
	c.traceMu.Lock()
	c.lineSeq++
	c.lineTrace = append(c.lineTrace, line)
	if n := len(c.lineTrace); n > maxLineTrace {
		copy(c.lineTrace, c.lineTrace[n-maxLineTrace:])
		c.lineTrace = c.lineTrace[:maxLineTrace]
	}
	c.traceMu.Unlock()
}

// LineTrace returns a copy of the recent executed-line ring and the monotonic
// total count of statements executed this conductor's life. The UI diffs the
// count to learn how many new lines to replay, reading them from the tail of the
// ring, and steps its highlight through them with a slight delay.
func (c *Conductor) LineTrace() (trace []int, seq int) {
	c.traceMu.Lock()
	defer c.traceMu.Unlock()
	out := make([]int, len(c.lineTrace))
	copy(out, c.lineTrace)
	return out, c.lineSeq
}

// gate blocks while the conductor is paused, returning ctx.Err() if the context
// is cancelled while waiting (so a paused host still stops cleanly on shutdown).
func (c *Conductor) gate(ctx context.Context) error {
	for {
		c.pauseMu.Lock()
		if !c.paused {
			c.pauseMu.Unlock()
			return nil
		}
		ch := c.resumeCh
		c.pauseMu.Unlock()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ch:
		}
	}
}

// Store exposes the conductor's durable local store (for wiring DSL access or
// inspection). Never nil after NewConductor.
func (c *Conductor) Store() *hostkv.Store { return c.store }

// Scratch exposes the conductor's ephemeral cache. Never nil after NewConductor.
func (c *Conductor) Scratch() *hostkv.Scratch { return c.scratch }

// Run drives the turn loop until the director is done or ctx is cancelled.
// It returns ctx.Err() on cancellation and nil when the director stops.
//
// Host.Run must already be pumping frames on another goroutine; otherwise
// routines that wait for world changes will time out.
func (c *Conductor) Run(ctx context.Context) error {
	if c.director == nil {
		c.log.Warn("conductor: no director, nothing to do")
		return nil
	}
	if c.detours {
		go c.survivalArbiter(ctx)     // watch HP; park the grind to eat when critical
		go c.fatigueArbiter(ctx)      // watch fatigue; park the grind to sleep when exhausted
		go c.displacementArbiter(ctx) // watch position; abort + re-plan on a large unexpected jump
	}
	var last Outcome
	turn := 0
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := c.gate(ctx); err != nil { // block here while paused
			return err
		}
		intent, ok := c.director.Next(ctx, c.host, last)
		if !ok {
			c.log.Info("conductor: director done", "turns", turn)
			return nil
		}
		turn++
		c.log.Info("conductor: turn start", "turn", turn, "intent", intent.Label, "path", intent.RoutinePath)
		c.curMu.Lock()
		c.curIntent = intent.Label
		// Inline DSL source for the live Routine panel. File/library routines have
		// no inline Source — the panel then just shows the label. Reset the line so
		// a new turn doesn't briefly highlight the previous routine's position.
		c.curSource = intent.Source
		c.curLine = 0
		c.curMu.Unlock()
		out := c.execute(ctx, intent)
		c.recordTurn(turn, out)
		if c.Paused() {
			// The turn was interrupted by a pause / analysis request — do NOT feed
			// the aborted outcome to the director's failure/spin accounting; keep
			// the prior `last` so it re-plans cleanly on resume.
			c.log.Info("conductor: turn interrupted by pause", "turn", turn)
		} else {
			last = out
		}

		if c.settle > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.settle):
			}
		}
	}
}

// beginTurn creates the per-turn timeout context and registers its cancel so
// Pause() can interrupt a running turn promptly (not just at the next boundary).
// The returned done() clears the registration and cancels — defer it.
func (c *Conductor) beginTurn(ctx context.Context) (context.Context, func()) {
	turnCtx, cancel := context.WithTimeout(ctx, c.turnTimeout)
	c.pauseMu.Lock()
	c.turnCancel = cancel
	c.pauseMu.Unlock()
	return turnCtx, func() {
		c.pauseMu.Lock()
		c.turnCancel = nil
		c.pauseMu.Unlock()
		cancel()
	}
}

// executeRoutine runs one intent's routine to completion under a per-turn
// timeout and maps the result onto an Outcome.
func (c *Conductor) executeRoutine(ctx context.Context, in Intent) Outcome {
	start := time.Now()
	turnCtx, done := c.beginTurn(ctx)
	defer done()

	var (
		res interp.Result
		err error
	)
	if in.Source != "" {
		res, err = c.host.RunRoutineSource(turnCtx, in.Name, in.Source, in.Args)
	} else {
		res, err = c.host.RunRoutine(turnCtx, in.RoutinePath, in.Args)
	}
	dur := time.Since(start)
	if err != nil {
		// Parse/load failure (bad path, malformed routine, or invalid authored
		// DSL) — surface as an errored outcome so the director can react.
		c.log.Warn("conductor: routine failed to start", "intent", in.Label, "path", in.RoutinePath, "err", err)
		return Outcome{Intent: in, Kind: interp.ResultErrored, Err: &interp.RuntimeError{Msg: err.Error()}, Duration: dur}
	}
	// Classify a budget expiry: the TURN ran out, not the routine. DeadlineExceeded
	// is unambiguous here — Pause/shutdown cancel via cancel() (context.Canceled).
	budgetExpired := turnCtx.Err() == context.DeadlineExceeded
	c.log.Info("conductor: turn complete", "intent", in.Label, "result", res.Kind.String(),
		"budget_expired", budgetExpired, "dur", dur.Round(time.Millisecond))
	return Outcome{Intent: in, Kind: res.Kind, Value: res.Value, Err: res.Err, Duration: dur,
		BudgetExpired: budgetExpired}
}

// recordTurn persists turn bookkeeping: durable counters that survive a
// restart, plus a scratch note of the last result for quick reads / the next
// director decision.
func (c *Conductor) recordTurn(turn int, out Outcome) {
	if err := hostkv.Set(c.store, "conductor:turns", turn); err != nil {
		c.log.Warn("conductor: persist turn count", "err", err)
	}
	if err := hostkv.Set(c.store, "conductor:last_intent", out.Intent.Label); err != nil {
		c.log.Warn("conductor: persist last intent", "err", err)
	}
	c.scratch.Set("conductor:last_result", out.Kind.String(), 0)
}
