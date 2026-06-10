package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
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

// TestReactiveCoalescingKeepsLatestPerSpeaker proves the chat-spam shield's
// data path: reactive interrupts signaled while a turn runs fold into ONE
// pending entry per speaker (latest wins, case-folded key, arrival order kept),
// and the batch is consume-once.
func TestReactiveCoalescingKeepsLatestPerSpeaker(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})

	mk := func(speaker, reason string) detourReq {
		return detourReq{tier: "reactive", speaker: speaker, reason: reason,
			intent: Intent{Name: "reactive_react", Source: `runtime "1.0"` + "\nroutine reactive_react() { }", OneShot: true}}
	}
	c.signalDetour(mk("alice", "a1"))
	c.signalDetour(mk("bob", "b1"))
	c.signalDetour(mk("alice", "a2"))
	c.signalDetour(mk("Alice", "a3")) // same speaker, different case → same key

	batch := c.takePendingReactive()
	if len(batch) != 2 {
		t.Fatalf("want 2 coalesced entries (alice, bob), got %d: %+v", len(batch), batch)
	}
	if batch[0].reason != "a3" {
		t.Fatalf("alice's entry must be her LATEST interrupt (a3), got %q", batch[0].reason)
	}
	if batch[1].reason != "b1" {
		t.Fatalf("bob's entry must be b1, got %q", batch[1].reason)
	}
	if again := c.takePendingReactive(); again != nil {
		t.Fatalf("the batch is consume-once; second take must be nil, got %+v", again)
	}
	// The channel carries only wakeup markers (payload-free), never the reqs.
	for {
		select {
		case m := <-c.interrupts:
			if m.tier != "reactive" || m.intent.Source != "" || m.intent.RoutinePath != "" {
				t.Fatalf("channel must hold only payload-free reactive wakeup markers, got %+v", m)
			}
			continue
		default:
		}
		break
	}
}

// TestReactiveKeyFallsBackToReasonParse proves coalescing still keys per
// speaker when the signaling call site does not set the speaker field, by
// parsing maybeInterrupt's reason format "reactive[<kind>] <speaker>: <gist>".
func TestReactiveKeyFallsBackToReasonParse(t *testing.T) {
	if k := reactiveKey(detourReq{speaker: "Drone6"}); k != "drone6" {
		t.Fatalf("explicit speaker must win (case-folded): got %q", k)
	}
	if k := reactiveKey(detourReq{reason: "reactive[warning] Drone6: tar prices are crashing"}); k != "drone6" {
		t.Fatalf("reason-format fallback must extract the speaker: got %q", k)
	}
	if k := reactiveKey(detourReq{reason: "no recognizable format"}); k != "?" {
		t.Fatalf("unparseable reason must coalesce under the unknown key: got %q", k)
	}
}

// TestReactiveBatchParksGrindOnce proves the behavioral end of the shield: a
// burst of chat-driven interrupts from one speaker costs the grind ONE
// park/detour (carrying the latest line), not one per message — and the grind
// still resumes to completion.
func TestReactiveBatchParksGrindOnce(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	// Each runDetour publishes exactly one "detour:reactive" decision; count them.
	thoughts := h.Bus().Subscribe("*", 128)

	mk := func(reason string) detourReq {
		return detourReq{tier: "reactive", speaker: "drone6", reason: reason,
			intent: Intent{Name: "reactive_react", Source: `runtime "1.0"` + "\nroutine reactive_react() { wait(0.2) }", OneShot: true}}
	}
	// Signal the burst BEFORE the turn starts: all three coalesce to one entry.
	c.signalDetour(mk("first"))
	c.signalDetour(mk("second"))
	c.signalDetour(mk("third"))

	grind := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() {
    i = 0
    while i < 4 {
        i = i + 1
        wait(0.2)
    }
    return i
}`}
	var out Outcome
	done := make(chan struct{})
	go func() {
		out = c.executeWithDetours(context.Background(), grind)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(6 * time.Second):
		t.Fatal("executeWithDetours never finished — likely a suspend/resume deadlock")
	}
	if out.Kind != interp.ResultReturned || out.Value.Display() != "4" {
		t.Fatalf("grind did not resume to completion: kind=%v val=%v", out.Kind, out.Value.Display())
	}

	parks := 0
	var lastReason string
	for {
		select {
		case ev := <-thoughts:
			if th, ok := ev.(event.AgentThought); ok && th.Trigger == "detour:reactive" {
				parks++
				lastReason = th.Reasoning
			}
			continue
		default:
		}
		break
	}
	if parks != 1 {
		t.Fatalf("a coalesced burst must park the grind exactly once, got %d parks", parks)
	}
	if !strings.Contains(lastReason, "third") {
		t.Fatalf("the one serviced interrupt must be the LATEST per speaker; reasoning=%q", lastReason)
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
