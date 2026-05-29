package world

import (
	"testing"

	"github.com/gen0cide/westworld/event"
)

// TestApplySleepStateTransitions checks the sleep-screen state mirror:
// SleepScreenAppeared (SEND_SLEEPSCREEN) sets is_sleeping true, and
// SleepEnded (SEND_STOPSLEEP) clears it.
func TestApplySleepStateTransitions(t *testing.T) {
	w := NewWorld()
	if w.Self.IsSleeping() {
		t.Fatal("fresh world: IsSleeping should be false")
	}

	w.Apply(event.SleepScreenAppeared{ImageBytes: 128})
	if !w.Self.IsSleeping() {
		t.Fatal("after SleepScreenAppeared: IsSleeping should be true")
	}

	w.Apply(event.SleepEnded{})
	if w.Self.IsSleeping() {
		t.Fatal("after SleepEnded: IsSleeping should be false")
	}
}
