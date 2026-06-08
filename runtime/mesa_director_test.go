package runtime

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/dsl/interp"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

func quietDirector() *MesaDirector {
	return NewMesaDirector(mesaclient.StubClient{}, "tester", "do a thing", slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// TestMesaDirectorFailStreakEscalates proves anti-stuck v0: consecutive failed
// outcomes escalate the trigger to BLOCKED ("abandon this approach"), and any
// success resets the streak so a working host is never flagged.
func TestMesaDirectorFailStreakEscalates(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	// Production always sets Label alongside Name (see every Intent{} site); the
	// fail-streak gate keys on Label, matching triggerFor's "start" test.
	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	var sit *mesaclient.Situation
	for range antiStuckHardFails {
		sit = d.situation(h, fail)
	}
	if d.failStreak != antiStuckHardFails {
		t.Fatalf("failStreak = %d, want %d", d.failStreak, antiStuckHardFails)
	}
	if !strings.Contains(sit.Trigger, "BLOCKED") {
		t.Fatalf("expected BLOCKED after %d failures, got: %q", antiStuckHardFails, sit.Trigger)
	}

	// A successful turn clears the streak (a working grind must not be flagged).
	sit = d.situation(h, Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultCompleted})
	if d.failStreak != 0 {
		t.Fatalf("failStreak = %d, want 0 after a success", d.failStreak)
	}
	if strings.Contains(sit.Trigger, "BLOCKED") {
		t.Fatalf("a success should clear BLOCKED, got: %q", sit.Trigger)
	}
}

// TestMesaDirectorDisplacementResetsFailStreak proves a death/teleport clears
// the failure streak: the prior streak was about the OLD situation, so after a
// displacement the trigger is the re-orient message (not a stale BLOCKED), and
// the streak starts fresh so BLOCKED can't immediately re-fire next turn.
func TestMesaDirectorDisplacementResetsFailStreak(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	for range antiStuckHardFails {
		d.situation(h, fail)
	}
	if d.failStreak < antiStuckHardFails {
		t.Fatalf("precondition: expected a hard streak, got %d", d.failStreak)
	}

	// Now she dies and respawns. The displacement must win the trigger AND reset
	// the streak so the next turn doesn't re-trigger BLOCKED on stale failures.
	h.displacement.record(displacementEvent{toX: 120, toY: 648, dist: 99, cause: dispDeath})
	sit := d.situation(h, fail)
	if d.failStreak != 0 {
		t.Fatalf("displacement should reset failStreak, got %d", d.failStreak)
	}
	if strings.Contains(sit.Trigger, "BLOCKED") || !strings.Contains(sit.Trigger, "DIED") {
		t.Fatalf("displacement trigger should be the death re-orient, not BLOCKED: %q", sit.Trigger)
	}
}

// TestMesaDirectorFailStreakIgnoresNoOpTurns proves the first/no-action turns
// don't accrue a phantom failure streak (zero Outcome has no Intent).
func TestMesaDirectorFailStreakIgnoresNoOps(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	for range 5 {
		d.situation(h, Outcome{}) // zero outcome: no prior action
	}
	if d.failStreak != 0 {
		t.Fatalf("no-op turns should not accrue failures; failStreak=%d", d.failStreak)
	}
}

// TestDirectorSeedsRootGoal proves Hook 1: the standing objective becomes a
// VISIBLE root node (KindGoal/StatusActive) on the first turn, and the Has-guard
// prevents dup/churn on subsequent turns.
func TestDirectorSeedsRootGoal(t *testing.T) {
	h := newTestHost()
	d := quietDirector() // goal "do a thing"

	d.situation(h, Outcome{})
	if !h.goalGraph.Has("do a thing") {
		t.Fatal("first turn should seed the goal as a graph node")
	}
	n, ok := h.goalGraph.Get("do a thing")
	if !ok || n.Kind != goalgraph.KindGoal || n.Status != goalgraph.StatusActive {
		t.Fatalf("root goal node = %+v, want KindGoal/StatusActive", n)
	}

	// Subsequent turns must not create a second node.
	d.situation(h, Outcome{})
	if got := len(h.goalGraph.Nodes()); got != 1 {
		t.Fatalf("goal node should be seeded once; have %d nodes", got)
	}
}

// TestDirectorBlockedSpawnsOpenQuestion proves Hook 2: a hard fail-streak marks
// the goal blocked and spawns one open-question node (goal --blocked_by--> Q),
// reinforced (not duplicated) on further failures via the deterministic id.
func TestDirectorBlockedSpawnsOpenQuestion(t *testing.T) {
	h := newTestHost()
	d := quietDirector() // goal "do a thing"

	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	for range antiStuckHardFails {
		d.situation(h, fail)
	}

	qs := h.goalGraph.OpenQuestions()
	if len(qs) != 1 {
		t.Fatalf("BLOCKED should spawn exactly one open question; got %d", len(qs))
	}
	if blk := h.goalGraph.Blockers("do a thing"); len(blk) != 1 || blk[0].ID != qs[0].ID {
		t.Fatalf("goal should be blocked_by the open question; blockers=%v question=%v", blk, qs[0])
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusBlocked {
		t.Fatalf("goal status = %q, want blocked", n.Status)
	}

	// A 5th failure reinforces the SAME node (deterministic id), not a duplicate.
	d.situation(h, fail)
	if got := len(h.goalGraph.OpenQuestions()); got != 1 {
		t.Fatalf("repeated BLOCKED should reinforce one node; got %d open questions", got)
	}
}

// TestDirectorUnblocksOnSuccess proves Hook 3: after BLOCKED, a real success
// flips the goal status back to active (the only non-monotonic write).
func TestDirectorUnblocksOnSuccess(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	for range antiStuckHardFails {
		d.situation(h, fail)
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusBlocked {
		t.Fatalf("precondition: goal should be blocked, got %q", n.Status)
	}

	d.situation(h, Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultCompleted})
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusActive {
		t.Fatalf("success should un-block the goal; status = %q", n.Status)
	}
}

// TestDirectorBlockedWithEmptyGoalIsNoop proves Hook 2's empty-goal guard: with
// no goal there is nothing to seed or block, so even a hard fail-streak writes NO
// goal-graph nodes — and crucially never an empty-ID node that would corrupt the
// graph (the bug the happy-path tests miss because quietDirector has a goal).
func TestDirectorBlockedWithEmptyGoalIsNoop(t *testing.T) {
	h := newTestHost()
	d := NewMesaDirector(mesaclient.StubClient{}, "tester", "", slog.New(slog.NewTextHandler(io.Discard, nil)))

	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	for range antiStuckHardFails {
		d.situation(h, fail)
	}
	if got := len(h.goalGraph.Nodes()); got != 0 {
		t.Fatalf("empty goal must not write goal-graph nodes; have %d: %+v", got, h.goalGraph.Nodes())
	}
	if h.goalGraph.Has("") {
		t.Fatal("empty goal must not create an empty-ID node")
	}
}
