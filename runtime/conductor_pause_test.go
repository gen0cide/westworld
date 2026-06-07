package runtime

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
)

// TestConductorPauseResume proves Pause halts the turn loop at the turn boundary
// (no new turns start) and Resume releases it — without tearing the host down.
func TestConductorPauseResume(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Settle:   5 * time.Millisecond,
		Director: Loop(Intent{Label: "tick"}),
	})
	// Substitute a fake executor so the loop spins without a live server.
	var turns int32
	c.execute = func(_ context.Context, in Intent) Outcome {
		atomic.AddInt32(&turns, 1)
		return Outcome{Intent: in, Kind: interp.ResultCompleted}
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() { _ = c.Run(ctx); close(done) }()

	// Let it spin, then pause.
	time.Sleep(60 * time.Millisecond)
	c.Pause()
	if !c.Paused() {
		t.Fatal("Paused() should report true after Pause")
	}
	time.Sleep(40 * time.Millisecond) // let any in-flight turn finish + gate
	c1 := atomic.LoadInt32(&turns)
	time.Sleep(80 * time.Millisecond) // while paused, no new turns
	c2 := atomic.LoadInt32(&turns)
	if c2 != c1 {
		t.Fatalf("paused loop kept running: %d -> %d", c1, c2)
	}

	// Resume → progress continues.
	c.Resume()
	if c.Paused() {
		t.Fatal("Paused() should report false after Resume")
	}
	time.Sleep(80 * time.Millisecond)
	c3 := atomic.LoadInt32(&turns)
	if c3 <= c2 {
		t.Fatalf("resumed loop did not progress: %d -> %d", c2, c3)
	}

	// Cancelling stops a running (and a paused) loop cleanly.
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after cancel")
	}
}

// TestConductorPauseThenCancel proves a host paused at the gate still shuts down
// cleanly when its context is cancelled (no deadlock).
func TestConductorPauseThenCancel(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Settle:   5 * time.Millisecond,
		Director: Loop(Intent{Label: "tick"}),
	})
	c.execute = func(_ context.Context, in Intent) Outcome {
		return Outcome{Intent: in, Kind: interp.ResultCompleted}
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = c.Run(ctx); close(done) }()

	time.Sleep(30 * time.Millisecond)
	c.Pause()
	time.Sleep(30 * time.Millisecond) // settle into the gate
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("paused loop did not stop on ctx cancel — deadlock")
	}
}
