package runtime

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/proto/v235"
)

// Regression tests for the sleepword answer path (soak-retro cause #2:
// "sleep is a behavioral no-op"). The original bug: the runtime answered
// the captcha the instant SEND_SLEEPSCREEN arrived, so the server's
// correct-word wake fired after ~0 game ticks of sleep drain and
// committed the UNDRAINED fatigue back (Player.handleWakeup: fatigue =
// sleepStateFatigue) — "You wake up - feeling refreshed" with zero
// restoration, fleet-wide. The fix answers only once SEND_SLEEP_FATIGUE
// (opcode 244) reports the drain reached 0.

// sleepTestHost wires the captcha-answer seam to a recorder so the
// handler path can be driven with synthetic frames, no socket. The
// counter is atomic because the fallback path answers from a timer
// goroutine.
func sleepTestHost() (*Host, *atomic.Int32) {
	h := newTestHost()
	sent := new(atomic.Int32)
	h.sendSleepWord = func() error { sent.Add(1); return nil }
	return h, sent
}

// sleepFatigueFrame builds an opcode-244 frame carrying the scaled
// drain value (big-endian short, like SEND_FATIGUE).
func sleepFatigueFrame(v int) v235.Frame {
	return v235.Frame{
		Opcode:  v235.InSleepFatigue,
		Payload: []byte{byte(v >> 8), byte(v)},
	}
}

// TestSleepwordAnsweredOnlyAfterDrain is the core regression: the
// captcha answer must NOT go out when the sleep screen appears — only
// when the drain completes — and exactly once.
func TestSleepwordAnsweredOnlyAfterDrain(t *testing.T) {
	h, sent := sleepTestHost()

	// Sleep screen up: asleep, but the answer must be withheld.
	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1, 2, 3}})
	if !h.world.Self.IsSleeping() {
		t.Fatal("after sleep screen: self.is_sleeping should be true")
	}
	if sent.Load() != 0 {
		t.Fatalf("answered the captcha on screen-appear (sent=%d) — that wakes the server before any drain tick and forfeits the fatigue restore", sent.Load())
	}

	// Drain in progress: still no answer.
	h.handleFrame(sleepFatigueFrame(710))
	h.handleFrame(sleepFatigueFrame(420))
	if sent.Load() != 0 {
		t.Fatalf("answered mid-drain (sent=%d), want 0", sent.Load())
	}

	// Drain complete: answer exactly once.
	h.handleFrame(sleepFatigueFrame(0))
	if sent.Load() != 1 {
		t.Fatalf("after drain reached 0: sent=%d, want 1", sent.Load())
	}

	// Wake: state resets, no double answer.
	h.handleFrame(v235.Frame{Opcode: v235.InStopSleep})
	if h.world.Self.IsSleeping() {
		t.Fatal("after stop-sleep: self.is_sleeping should be false")
	}
	if sent.Load() != 1 {
		t.Fatalf("after wake: sent=%d, want still 1", sent.Load())
	}
}

// TestSleepwordAnswersImmediatelyWhenAlreadyDrained: sleeping with zero
// fatigue (the initial 244 reports 0 right away) must answer at once —
// there is nothing to drain, and waiting would just trip the fallback.
func TestSleepwordAnswersImmediatelyWhenAlreadyDrained(t *testing.T) {
	h, sent := sleepTestHost()
	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1}})
	h.handleFrame(sleepFatigueFrame(0))
	if sent.Load() != 1 {
		t.Fatalf("zero-fatigue sleep: sent=%d, want 1", sent.Load())
	}
}

// TestSleepwordRetryAfterIncorrect: a rejected answer (opcode 194, only
// possible if the server starts checking real captcha words) keeps us
// asleep and the server re-sends the sleep screen. The drain already
// completed — no further 244s will come — so the re-sent screen must be
// answered immediately: retry, not trap.
func TestSleepwordRetryAfterIncorrect(t *testing.T) {
	h, sent := sleepTestHost()

	// First cycle: screen → drain → answer.
	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1}})
	h.handleFrame(sleepFatigueFrame(0))
	if sent.Load() != 1 {
		t.Fatalf("first answer: sent=%d, want 1", sent.Load())
	}

	// Server rejects the word, then re-sends the screen.
	h.handleFrame(v235.Frame{Opcode: v235.InSleepwordIncorrect})
	if sent.Load() != 1 {
		t.Fatalf("194 alone must not answer: sent=%d, want 1", sent.Load())
	}
	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1}})
	if sent.Load() != 2 {
		t.Fatalf("re-sent screen after incorrect: sent=%d, want 2 (immediate retry)", sent.Load())
	}
}

// TestSleepwordFallbackAnswersWithoutDrainSignal: if no SEND_SLEEP_FATIGUE
// ever reaches 0 (server stops sending 244s), the fallback timer answers
// anyway so the host can't be trapped on the sleep screen forever.
func TestSleepwordFallbackAnswersWithoutDrainSignal(t *testing.T) {
	h, sent := sleepTestHost()
	h.sleepFallbackAfter = 10 * time.Millisecond

	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1}})
	if sent.Load() != 0 {
		t.Fatalf("answered before fallback fired: sent=%d", sent.Load())
	}
	deadline := time.Now().Add(2 * time.Second)
	for sent.Load() == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if sent.Load() != 1 {
		t.Fatalf("fallback never answered: sent=%d, want 1", sent.Load())
	}
}

// TestSleepFatigueDoesNotTouchWorldFatigue: the 244 drain value is
// provisional (commits only on a successful wake via a real 114), so it
// must never overwrite world.Self fatigue — an unexpected wake keeps
// the old value, and a false "fatigue 0" belief would silence the
// fatigue arbiter while XP stays frozen.
func TestSleepFatigueDoesNotTouchWorldFatigue(t *testing.T) {
	h, _ := sleepTestHost()
	h.world.Self.SetFatigue(700)
	h.handleFrame(v235.Frame{Opcode: v235.InSleepScreen, Payload: []byte{1}})
	h.handleFrame(sleepFatigueFrame(280))
	if got := h.world.Self.Fatigue(); got != 700 {
		t.Fatalf("world fatigue followed the provisional drain: got %d, want 700", got)
	}
	// The committed value arrives as a regular FatigueUpdate (114).
	h.world.Apply(event.FatigueUpdate{Value: 0})
	if got := h.world.Self.Fatigue(); got != 0 {
		t.Fatalf("committed FatigueUpdate not applied: got %d, want 0", got)
	}
}
