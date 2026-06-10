package runtime

import (
	"context"
	"testing"
	"time"
)

// These tests pin the freeze contract for detours (leak-audit 2026-06-10 bug
// #4): Pause() must cancel a RUNNING detour promptly — not just the parked
// grind's turn ctx — while a detour must still be free to outlive the turn
// budget when NOT paused (its ctx parents to the conductor ctx, never the
// turn ctx, so a 3-minute fatigue nap survives the 2-minute turn deadline).

// TestPauseCancelsRunningDetour proves the fix for bug #4: an operator Pause
// (cradle freeze / ANALYSIS entry) that lands MID-DETOUR terminates the detour
// promptly — the turn joins within a bounded window instead of the detour
// running up to its 30s/3m timeout — and the frozen host executes no further
// statements afterward.
func TestPauseCancelsRunningDetour(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	grind := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() { wait(30) }`}
	detour := Intent{Name: "detour", Source: `runtime "1.0"
routine detour() { wait(30) }`}

	done := make(chan struct{})
	go func() {
		_ = c.executeWithDetours(context.Background(), grind)
		close(done)
	}()

	// Let the grind start, then inject the interrupt and wait until the detour
	// is actually EXECUTING. The grind's single wait statement fires exactly one
	// OnStmt, so any lineSeq growth after this point is the detour's statements.
	time.Sleep(300 * time.Millisecond)
	_, seq0 := c.LineTrace()
	c.interrupts <- detourReq{tier: "fatigue", reason: "test", intent: detour}
	deadline := time.Now().Add(3 * time.Second)
	for {
		if _, seq := c.LineTrace(); seq > seq0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("detour never started executing")
		}
		time.Sleep(10 * time.Millisecond)
	}

	// Operator freezes the host mid-detour.
	c.Pause()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Pause did not cancel the running detour: turn still live 3s after Pause (detour had ~30s left)")
	}

	// Frozen means frozen: no statement may execute after the turn joined.
	_, s1 := c.LineTrace()
	time.Sleep(400 * time.Millisecond)
	_, s2 := c.LineTrace()
	if s2 != s1 {
		t.Fatalf("host kept acting after Pause: lineSeq %d -> %d", s1, s2)
	}
}

// TestDetourOutlivesTurnBudget is the audit's explicit trap test: a fatigue
// detour that runs LONGER than the per-turn budget must still complete when the
// host is NOT paused. The detour ctx parents to the conductor ctx — re-parenting
// it to turnCtx (the "obvious" fix for bug #4) would kill 3-minute fatigue naps
// at the 2-minute turn deadline; this test fails if that regression lands. It
// also pins the budget-expiry classification: the grind's turn deadline expiring
// while parked reads as BudgetExpired, never as a detour failure.
func TestDetourOutlivesTurnBudget(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1, TurnTimeout: 1 * time.Second,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	grind := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() { wait(30) }`}
	// The marker assignment sits on line 4: it executes only if the detour
	// survives past the 1s turn deadline (the nap is 2s).
	detour := Intent{Name: "detour", Source: `runtime "1.0"
routine detour() {
    wait(2)
    marker = 1
}`}

	var out Outcome
	done := make(chan struct{})
	go func() {
		out = c.executeWithDetours(context.Background(), grind)
		close(done)
	}()

	time.Sleep(300 * time.Millisecond)
	c.interrupts <- detourReq{tier: "fatigue", reason: "test", intent: detour}

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("turn never finished")
	}

	trace, _ := c.LineTrace()
	sawMarker := false
	for _, ln := range trace {
		if ln == 4 {
			sawMarker = true
			break
		}
	}
	if !sawMarker {
		t.Fatalf("detour was killed by the turn deadline: post-nap marker (line 4) never executed (trace=%v)", trace)
	}
	if out.Duration < 1800*time.Millisecond {
		t.Fatalf("turn ended near the 1s budget (%v) — the detour did not outlive it", out.Duration)
	}
	if !out.BudgetExpired {
		t.Fatal("expected BudgetExpired=true: the grind's turn budget ran out while parked, which is a scheduling event, not a failure")
	}
}

// TestPauseAfterDetourCancelsNextTurn proves the restore path: after a detour
// exits NORMALLY (swapping turnCancel back), Pause must still cancel the next
// plain turn promptly — i.e. the dual-cancel swap did not leave a stale or nil
// turnCancel behind.
func TestPauseAfterDetourCancelsNextTurn(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	// Turn 1: a grind long enough to host a detour; the detour completes
	// normally and the grind resumes to completion.
	grind1 := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() {
    i = 0
    while i < 6 {
        i = i + 1
        wait(0.2)
    }
}`}
	detour := Intent{Name: "detour", Source: `runtime "1.0"
routine detour() { wait(0.3) }`}

	done1 := make(chan struct{})
	go func() {
		_ = c.executeWithDetours(context.Background(), grind1)
		close(done1)
	}()
	time.Sleep(250 * time.Millisecond)
	c.interrupts <- detourReq{tier: "fatigue", reason: "test", intent: detour}
	select {
	case <-done1:
	case <-time.After(6 * time.Second):
		t.Fatal("turn 1 (grind + completed detour) never finished")
	}

	// Turn 2: a plain long turn — Pause must interrupt it promptly.
	grind2 := Intent{Name: "grind2", Source: `runtime "1.0"
routine grind2() { wait(30) }`}
	done2 := make(chan struct{})
	go func() {
		_ = c.executeWithDetours(context.Background(), grind2)
		close(done2)
	}()
	time.Sleep(300 * time.Millisecond)
	c.Pause()
	select {
	case <-done2:
	case <-time.After(3 * time.Second):
		t.Fatal("Pause did not cancel the turn after a completed detour — turnCancel was not restored")
	}
}
