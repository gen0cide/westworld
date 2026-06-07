package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/hostkv"
)

// newTestConductor builds a conductor with a fake routine runner so the loop
// can be exercised without a live server. record receives each executed intent.
func newTestConductor(t *testing.T, d Director, run func(in Intent) Outcome) *Conductor {
	t.Helper()
	c := NewConductor(nil, ConductorOptions{
		Director: d,
		Settle:   -1, // no inter-turn pause in tests
		Store:    hostkv.NewMemory(),
		Scratch:  hostkv.NewScratch(16),
	})
	c.execute = func(_ context.Context, in Intent) Outcome { return run(in) }
	return c
}

func okOutcome(in Intent) Outcome {
	return Outcome{Intent: in, Kind: interp.ResultCompleted}
}

func TestConductorSequenceRunsEachOnceThenStops(t *testing.T) {
	var ran []string
	d := Sequence(
		Intent{Label: "a", RoutinePath: "a.rt"},
		Intent{Label: "b", RoutinePath: "b.rt"},
		Intent{Label: "c", RoutinePath: "c.rt"},
	)
	c := newTestConductor(t, d, func(in Intent) Outcome {
		ran = append(ran, in.Label)
		return okOutcome(in)
	})
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(ran) != 3 || ran[0] != "a" || ran[2] != "c" {
		t.Fatalf("ran = %v, want [a b c]", ran)
	}
	// durable turn count recorded
	turns, ok, _ := hostkv.Get[int](c.Store(), "conductor:turns")
	if !ok || turns != 3 {
		t.Fatalf("conductor:turns = %d (ok=%v), want 3", turns, ok)
	}
	if last, _ := c.Scratch().Get("conductor:last_result"); last != "completed" {
		t.Fatalf("last_result = %v, want completed", last)
	}
}

func TestConductorLoopStopsOnContextCancel(t *testing.T) {
	count := 0
	c := newTestConductor(t, Loop(Intent{Label: "wander"}), func(in Intent) Outcome {
		count++
		return okOutcome(in)
	})
	ctx, cancel := context.WithCancel(context.Background())
	// cancel after a handful of turns by stopping from inside the runner.
	c.execute = func(_ context.Context, in Intent) Outcome {
		count++
		if count >= 5 {
			cancel()
		}
		return okOutcome(in)
	}
	err := c.Run(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Run err = %v, want context.Canceled", err)
	}
	if count < 5 {
		t.Fatalf("count = %d, want >= 5", count)
	}
}

func TestConductorThreadsLastOutcomeToDirector(t *testing.T) {
	// A director that runs "retry" until the previous turn was OK, then stops.
	var seenLastKinds []interp.ResultKind
	calls := 0
	d := DirectorFunc(func(_ context.Context, _ *Host, last Outcome) (Intent, bool) {
		seenLastKinds = append(seenLastKinds, last.Kind)
		calls++
		if calls > 3 {
			return Intent{}, false
		}
		return Intent{Label: "try"}, true
	})
	// first two turns error, third completes
	turn := 0
	c := newTestConductor(t, d, func(in Intent) Outcome {
		turn++
		if turn < 3 {
			return Outcome{Intent: in, Kind: interp.ResultErrored, Err: &interp.RuntimeError{Msg: "boom"}}
		}
		return okOutcome(in)
	})
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run: %v", err)
	}
	// first Next sees zero Outcome; subsequent see prior turn's kind
	if len(seenLastKinds) != 4 {
		t.Fatalf("director called %d times, want 4", len(seenLastKinds))
	}
	if seenLastKinds[1] != interp.ResultErrored || seenLastKinds[3] != interp.ResultCompleted {
		t.Fatalf("last kinds = %v", seenLastKinds)
	}
}

func TestConductorNilDirectorIsNoop(t *testing.T) {
	c := NewConductor(nil, ConductorOptions{Settle: -1})
	if err := c.Run(context.Background()); err != nil {
		t.Fatalf("Run with nil director: %v", err)
	}
}

func TestConductorSettleRespectsCancel(t *testing.T) {
	// With a real (positive) settle, a cancel during the pause returns promptly.
	c := NewConductor(nil, ConductorOptions{
		Director: Loop(Intent{Label: "x"}),
		Settle:   10 * time.Second,
		Store:    hostkv.NewMemory(),
		Scratch:  hostkv.NewScratch(4),
	})
	c.execute = func(_ context.Context, in Intent) Outcome { return okOutcome(in) }
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := c.Run(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run err = %v, want DeadlineExceeded", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("settle did not honor cancel: took %v", time.Since(start))
	}
}
