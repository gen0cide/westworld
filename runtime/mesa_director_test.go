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

	fail := Outcome{Intent: Intent{Name: "try_thing"}, Kind: interp.ResultErrored}
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
	sit = d.situation(h, Outcome{Intent: Intent{Name: "try_thing"}, Kind: interp.ResultCompleted})
	if d.failStreak != 0 {
		t.Fatalf("failStreak = %d, want 0 after a success", d.failStreak)
	}
	if strings.Contains(sit.Trigger, "BLOCKED") {
		t.Fatalf("a success should clear BLOCKED, got: %q", sit.Trigger)
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
