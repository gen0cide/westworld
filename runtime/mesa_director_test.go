package runtime

import (
	"io"
	"log/slog"
	"strings"
	"testing"

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
