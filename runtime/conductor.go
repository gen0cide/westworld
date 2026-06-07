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
	// longer is cancelled and reported as an errored Outcome. Default 2m.
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

	// pause gate: when paused, Run blocks at the turn boundary (between routines)
	// until Resume or ctx cancel — lets the cradle freeze a host live without
	// tearing it down. Pausing takes effect at the next turn; the current routine
	// always finishes first.
	pauseMu  sync.Mutex
	paused   bool
	resumeCh chan struct{}
}

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
	return c
}

// Pause halts the turn loop at the next turn boundary. The currently running
// routine is allowed to finish; no new turn starts until Resume. Safe to call
// from any goroutine (e.g. the cradle's control API).
func (c *Conductor) Pause() {
	c.pauseMu.Lock()
	c.paused = true
	c.pauseMu.Unlock()
}

// Resume releases a paused turn loop. A no-op if not paused.
func (c *Conductor) Resume() {
	c.pauseMu.Lock()
	if c.paused {
		c.paused = false
		close(c.resumeCh)                // wake any waiter
		c.resumeCh = make(chan struct{}) // re-arm for the next pause
	}
	c.pauseMu.Unlock()
}

// Paused reports whether the loop is currently paused.
func (c *Conductor) Paused() bool {
	c.pauseMu.Lock()
	defer c.pauseMu.Unlock()
	return c.paused
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
		go c.survivalArbiter(ctx) // watch HP; park the grind to eat when critical
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
		out := c.execute(ctx, intent)
		last = out
		c.recordTurn(turn, out)

		if c.settle > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(c.settle):
			}
		}
	}
}

// executeRoutine runs one intent's routine to completion under a per-turn
// timeout and maps the result onto an Outcome.
func (c *Conductor) executeRoutine(ctx context.Context, in Intent) Outcome {
	start := time.Now()
	turnCtx, cancel := context.WithTimeout(ctx, c.turnTimeout)
	defer cancel()

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
	c.log.Info("conductor: turn complete", "intent", in.Label, "result", res.Kind.String(), "dur", dur.Round(time.Millisecond))
	return Outcome{Intent: in, Kind: res.Kind, Value: res.Value, Err: res.Err, Duration: dur}
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
