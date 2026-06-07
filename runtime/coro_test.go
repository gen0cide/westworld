package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
)

// TestCoroSuspendResumePreservesState proves the suspend/resume primitive: a
// routine parked mid-loop and resumed completes with its accumulator intact —
// i.e. the live goroutine stack preserved the full execution state across the
// park, with no serialization. This is the foundation of the detour stack.
func TestCoroSuspendResumePreservesState(t *testing.T) {
	h := newTestHost()
	ctx := context.Background()
	src := `runtime "1.0"
routine sum() {
    total = 0
    i = 0
    while i < 3 {
        total = total + 10
        i = i + 1
        wait(0.2)
    }
    return total
}`
	coro := h.StartCoro(ctx, Intent{Name: "sum", Source: src})

	// Park it mid-loop. Suspend blocks until the routine actually parks at an
	// action boundary (a wait checkpoint), so a true return proves suspension.
	suspendCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if !coro.Suspend(suspendCtx) {
		t.Fatal("routine did not park (finished too fast or ctx cancelled)")
	}

	// While parked it must make no progress — Done stays open.
	select {
	case <-coro.Done():
		t.Fatal("routine completed while it was supposed to be parked")
	case <-time.After(300 * time.Millisecond):
		// good: still parked, not advancing
	}

	// Resume and let it run to completion.
	coro.Resume()
	select {
	case <-coro.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("routine did not complete after resume")
	}

	out := coro.Outcome()
	if out.Kind != interp.ResultReturned {
		t.Fatalf("kind = %v, want ResultReturned (err=%v)", out.Kind, out.Err)
	}
	if got := out.Value.Display(); got != "30" {
		t.Fatalf("suspend/resume corrupted loop state: result = %q, want 30", got)
	}
}

// TestCoroRunsToCompletionWithoutSuspend confirms a Coro behaves like a normal
// routine run when never parked.
func TestCoroRunsToCompletionWithoutSuspend(t *testing.T) {
	h := newTestHost()
	coro := h.StartCoro(context.Background(), Intent{Name: "quick", Source: `runtime "1.0"
routine quick() { return 7 }`})
	select {
	case <-coro.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("routine did not complete")
	}
	if out := coro.Outcome(); out.Kind != interp.ResultReturned || out.Value.Display() != "7" {
		t.Fatalf("unsuspended coro = kind %v val %v", out.Kind, out.Value.Display())
	}
}
