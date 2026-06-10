package runtime

import (
	"testing"
	"time"

	"github.com/gen0cide/westworld/world"
)

// Tests for the social-reflex per-speaker reply throttle + conversational-loop
// breaker (soak retro #12: the bernard/drone6/drone8 "tar brokerage" chat
// doom-spiral — each reflex reply socially obligating the next, 36-60% of all
// decisions chat-driven on social hosts).

// newTestGate builds a reflexGate on a controllable clock. Advance time by
// mutating *clock.
func newTestGate(social float64) (*reflexGate, *time.Time) {
	clock := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	g := newReflexGate(social)
	g.now = func() time.Time { return clock }
	return g, &clock
}

// TestReflexGateThrottleSuppressesWithinCooldown proves the core throttle: after
// a reply to speaker S, non-actionable lines from S are suppressed for the
// cooldown, other speakers are unaffected, and S is allowed again afterwards.
func TestReflexGateThrottleSuppressesWithinCooldown(t *testing.T) {
	g, clock := newTestGate(0) // flat default cooldown (45s)
	pos := world.Coord{X: 100, Y: 500}

	if ok, _ := g.allow("bob", "hello there", "delores"); !ok {
		t.Fatal("first line from a new speaker must be allowed")
	}
	g.recordReply("bob", pos)

	*clock = clock.Add(10 * time.Second)
	if ok, why := g.allow("bob", "how goes it", "delores"); ok || why != "cooldown" {
		t.Fatalf("line within cooldown must be suppressed as cooldown; got ok=%v why=%q", ok, why)
	}
	if ok, _ := g.allow("alice", "hello", "delores"); !ok {
		t.Fatal("a different speaker must not be throttled by bob's cooldown")
	}

	*clock = clock.Add(replyCooldown) // well past the 45s window
	if ok, _ := g.allow("bob", "how goes it", "delores"); !ok {
		t.Fatal("line after the cooldown must be allowed again")
	}
}

// TestReflexGateActionableBypassesCooldown proves the bypass: inside the
// cooldown, a trade proposal or a direct name-addressed question still gets a
// reply; small talk does not.
func TestReflexGateActionableBypassesCooldown(t *testing.T) {
	g, clock := newTestGate(0)
	pos := world.Coord{X: 100, Y: 500}
	g.recordReply("bob", pos)
	*clock = clock.Add(5 * time.Second) // deep inside the cooldown

	if ok, _ := g.allow("bob", "Delores, where do you mine?", "Delores"); !ok {
		t.Fatal("a direct question with the host's name + '?' must bypass the cooldown")
	}
	if ok, _ := g.allow("bob", "want to trade for some tar", "Delores"); !ok {
		t.Fatal("a trade proposal must bypass the cooldown")
	}
	if ok, _ := g.allow("bob", "lovely weather today", "Delores"); ok {
		t.Fatal("small talk inside the cooldown must be suppressed")
	}
	if ok, _ := g.allow("bob", "what do you think of the weather?", "Delores"); ok {
		t.Fatal("a question NOT addressed to the host by name must not bypass")
	}
}

// TestReflexGateLoopBreakerTripsAtThree proves the spiral breaker: three
// consecutive replies to one speaker with no world progress trip a long quiet
// window that even actionable lines cannot bypass, and that expires.
func TestReflexGateLoopBreakerTripsAtThree(t *testing.T) {
	g, clock := newTestGate(0)
	pos := world.Coord{X: 100, Y: 500} // never moves: no world progress

	if g.recordReply("bob", pos) {
		t.Fatal("reply 1 must not trip the breaker")
	}
	*clock = clock.Add(20 * time.Second)
	if g.recordReply("bob", pos) {
		t.Fatal("reply 2 must not trip the breaker")
	}
	*clock = clock.Add(20 * time.Second)
	if !g.recordReply("bob", pos) {
		t.Fatalf("reply %d with no progress must trip the breaker", reflexLoopN)
	}

	// Quiet window: even a name-addressed question stays suppressed (the
	// doom-spiral lines were themselves name-addressed questions).
	if ok, why := g.allow("bob", "Delores, deal on the tar?", "Delores"); ok || why != "loop-breaker" {
		t.Fatalf("actionable line inside the quiet window must be suppressed as loop-breaker; got ok=%v why=%q", ok, why)
	}
	if ok, _ := g.allow("alice", "hello", "Delores"); !ok {
		t.Fatal("the breaker is per-speaker — alice must still get a reply")
	}

	*clock = clock.Add(reflexLoopCooldown + time.Second)
	if ok, _ := g.allow("bob", "Delores, you there?", "Delores"); !ok {
		t.Fatal("after the quiet window expires the reflex must answer again")
	}
}

// TestReflexGateProgressResetsStreak proves the progress signal: movement of
// reflexProgressTiles+ between replies resets the streak, so a conversation
// carried on WHILE actually doing things never reads as a spiral.
func TestReflexGateProgressResetsStreak(t *testing.T) {
	g, clock := newTestGate(0)

	// Three replies, host walking between each: never trips.
	for i := 0; i < 3; i++ {
		pos := world.Coord{X: 100 + i*10, Y: 500}
		if g.recordReply("bob", pos) {
			t.Fatalf("reply %d with movement between replies must not trip the breaker", i+1)
		}
		*clock = clock.Add(20 * time.Second)
	}
	// Now the host stands still: streak rebuilds and trips on the 3rd
	// stationary exchange (the prior moving reply counts as exchange 1).
	still := world.Coord{X: 120, Y: 500}
	if g.recordReply("bob", still) { // streak 2 (last reply was at 120,500)
		t.Fatal("first stationary reply must not trip")
	}
	*clock = clock.Add(20 * time.Second)
	if !g.recordReply("bob", still) { // streak 3 → trip
		t.Fatal("third no-progress exchange must trip the breaker")
	}
}

// TestReflexGateColdGapResetsStreak proves slow conversation is not a spiral:
// replies further apart than reflexLoopWindow never chain.
func TestReflexGateColdGapResetsStreak(t *testing.T) {
	g, clock := newTestGate(0)
	pos := world.Coord{X: 100, Y: 500}
	g.recordReply("bob", pos)
	*clock = clock.Add(20 * time.Second)
	g.recordReply("bob", pos) // streak 2
	*clock = clock.Add(reflexLoopWindow + time.Second)
	if g.recordReply("bob", pos) {
		t.Fatal("a reply after a cold gap must start a fresh streak, not trip the breaker")
	}
}

// TestReplyCooldownForPersonaScaling proves personas matter: a very social host
// gets a shorter cooldown than a near-introvert, a neutral dial keeps the base,
// and a zero dial (personaless host) keeps the flat default.
func TestReplyCooldownForPersonaScaling(t *testing.T) {
	if got := replyCooldownFor(0); got != replyCooldown {
		t.Fatalf("zero dial (personaless) must keep the flat default; got %v", got)
	}
	if got := replyCooldownFor(0.5); got != replyCooldown {
		t.Fatalf("neutral dial must keep the base cooldown; got %v", got)
	}
	hi, lo := replyCooldownFor(1.0), replyCooldownFor(0.1)
	if hi >= lo {
		t.Fatalf("a very social host (%v) must re-engage sooner than a near-introvert (%v)", hi, lo)
	}
	if hi < replyCooldownMin {
		t.Fatalf("persona scaling must respect the floor: %v < %v", hi, replyCooldownMin)
	}
}
