package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
)

// TestConductorDetourSuspendsAndResumes proves the detour stack end-to-end in the
// conductor: a running grind is parked for an injected interrupt, the detour runs
// to completion, and the grind resumes exactly where it left off (loop state
// intact) — true mid-routine detour/resume via the suspend primitive.
func TestConductorDetourSuspendsAndResumes(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	grind := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() {
    i = 0
    while i < 5 {
        i = i + 1
        wait(0.2)
    }
    return i
}`}
	detour := Intent{Name: "detour", Source: `runtime "1.0"
routine detour() { wait(0.5) }`}

	var out Outcome
	done := make(chan struct{})
	go func() {
		out = c.executeWithDetours(context.Background(), grind)
		close(done)
	}()

	// Let the grind start and reach a wait checkpoint, then inject the interrupt.
	time.Sleep(250 * time.Millisecond)
	c.interrupts <- detourReq{tier: "survival", reason: "test", intent: detour}

	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Fatal("executeWithDetours never finished — likely a suspend/resume deadlock")
	}

	if out.Kind != interp.ResultReturned || out.Value.Display() != "5" {
		t.Fatalf("grind did not resume to completion: kind=%v val=%v", out.Kind, out.Value.Display())
	}
	// The detour (~0.5s) ran while the grind (~1.0s) was parked, so the total
	// wall-clock must exceed the grind running alone.
	if out.Duration < 1300*time.Millisecond {
		t.Fatalf("detour does not appear to have run (duration=%v; want grind ~1s + detour ~0.5s)", out.Duration)
	}
}

// TestShouldDetourTiers proves the ladder gate: survival always preempts; a
// non-survival interrupt detours only when not in a committed interaction.
func TestShouldDetourTiers(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})

	if !c.shouldDetour(detourReq{tier: "survival"}) {
		t.Fatal("survival must always preempt")
	}
	// No trade/duel/bank open on a fresh test host → not committed → social detours.
	if h.world.InCommittedRegion() {
		t.Fatal("precondition: fresh host should not be in a committed region")
	}
	if !c.shouldDetour(detourReq{tier: "social"}) {
		t.Fatal("a non-survival interrupt should detour when not in a committed region")
	}
}
