package interp

import "context"

// SuspendController lets an external orchestrator (the runtime conductor) PARK a
// running routine at an action boundary and resume it later — the basis for the
// interrupt / detour stack (a higher-tier event can suspend the grind, run a
// detour, then resume the grind exactly where it left off).
//
// The mechanism is goroutine parking, NOT execution-state serialization: a
// routine runs on its own goroutine, and at a checkpoint it simply blocks on a
// channel. Its entire execution state — the Go call stack, locals, defers, and
// watcher frames — stays live on that parked goroutine, so resuming is exact and
// free. Only one routine is ever active at a time (the orchestrator parks one
// before running another), so there are no new data races.
//
// A nil controller means "run to completion" (the legacy behavior).
type SuspendController struct {
	req    chan struct{} // orchestrator → routine: park at the next checkpoint
	parked chan struct{} // routine → orchestrator: I have parked (rendezvous)
	resume chan struct{} // orchestrator → routine: continue
}

// NewSuspendController builds a ready controller.
func NewSuspendController() *SuspendController {
	return &SuspendController{
		req:    make(chan struct{}, 1),
		parked: make(chan struct{}),
		resume: make(chan struct{}, 1),
	}
}

// RequestSuspend asks the routine to park at its next action boundary and blocks
// until it has parked. Returns true once parked; false if the routine finished
// (done closed) or ctx was cancelled before it could park. Called by the
// orchestrator goroutine. `done` is closed when the routine's goroutine exits.
func (s *SuspendController) RequestSuspend(ctx context.Context, done <-chan struct{}) bool {
	select {
	case s.req <- struct{}{}:
	default: // a request is already pending
	}
	select {
	case <-s.parked:
		return true
	case <-done:
		return false
	case <-ctx.Done():
		return false
	}
}

// Resume unblocks a parked routine. Non-blocking and idempotent-safe (the resume
// signal is buffered, so it is delivered whether or not the routine has reached
// its park-wait yet). Called by the orchestrator.
func (s *SuspendController) Resume() {
	select {
	case s.resume <- struct{}{}:
	default:
	}
}

// checkpoint is called by the interpreter at every action boundary (via
// dispatchPendingEvents). If a suspension was requested, it announces that it
// has parked and blocks until Resume (or ctx cancellation). A no-op otherwise.
func (s *SuspendController) checkpoint(ctx context.Context) {
	if s == nil {
		return
	}
	select {
	case <-s.req:
		// suspension requested → park
	default:
		return // nothing pending; continue running
	}
	// Announce we've parked (rendezvous with RequestSuspend), then wait.
	select {
	case s.parked <- struct{}{}:
	case <-ctx.Done():
		return
	}
	select {
	case <-s.resume:
	case <-ctx.Done():
	}
}
