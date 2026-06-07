package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Coro is a SUSPENDABLE routine execution: the routine runs on its own goroutine
// and can be parked at an action boundary (Suspend) and resumed exactly where it
// left off (Resume). This is the runtime half of the interrupt / detour stack —
// a higher-tier event can park the grind, run a detour Coro, then resume the
// grind with its full state intact (locals, call stack, defers, watchers all
// preserved on the parked goroutine; nothing is serialized).
//
// Only ONE Coro should be ACTIVE at a time: the orchestrator parks the current
// one before running another, which preserves the single-active-routine model
// the rest of the runtime (world access, action dispatch, h.routineCtx) assumes.
type Coro struct {
	intent Intent
	ctrl   *interp.SuspendController
	done   chan struct{}
	res    interp.Result
	err    error
}

// StartCoro launches intent's routine on a goroutine, suspendable via the
// returned Coro. It mirrors RunRoutine/RunRoutineSource but installs a suspend
// controller and runs asynchronously; the routine begins executing immediately.
func (h *Host) StartCoro(ctx context.Context, in Intent) *Coro {
	c := &Coro{intent: in, ctrl: interp.NewSuspendController(), done: make(chan struct{})}
	go func() {
		defer close(c.done)
		it := h.NewRoutineInterpreter(ctx)
		it.Suspend = c.ctrl
		var rf *RoutineFile
		var err error
		if in.Source != "" {
			name := in.Name
			if name == "" {
				name = "mesa/authored"
			}
			rf, err = ParseRoutineString(name, in.Source)
		} else {
			rf, err = ParseRoutineFile(in.RoutinePath)
		}
		if err != nil {
			c.err = err
			return
		}
		c.res = it.RunRoutine(ctx, rf.File, in.Args)
	}()
	return c
}

// Suspend parks the routine at its next action boundary, blocking until it has
// parked. Returns false if the routine already finished or ctx was cancelled
// before it could park.
func (c *Coro) Suspend(ctx context.Context) bool { return c.ctrl.RequestSuspend(ctx, c.done) }

// Resume continues a parked routine (non-blocking).
func (c *Coro) Resume() { c.ctrl.Resume() }

// Done is closed when the routine's goroutine exits (completed or errored).
func (c *Coro) Done() <-chan struct{} { return c.done }

// Outcome maps the finished routine onto an Outcome. Call only after Done fires.
func (c *Coro) Outcome() Outcome {
	if c.err != nil {
		return Outcome{Intent: c.intent, Kind: interp.ResultErrored, Err: &interp.RuntimeError{Msg: c.err.Error()}}
	}
	return Outcome{Intent: c.intent, Kind: c.res.Kind, Value: c.res.Value, Err: c.res.Err}
}
