package runtime

import (
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/world"
)

// Tests for the social-reflex per-speaker reply throttle + conversational-loop
// breaker (soak retro #12: the bernard/drone6/drone8 "tar brokerage" chat
// doom-spiral — each reflex reply socially obligating the next, 36-60% of all
// decisions chat-driven on social hosts).

// newTestGate builds a reflexGate on a controllable clock with no operator
// (the mentor exemption off — the historical default). Advance time by
// mutating *clock.
func newTestGate(social float64) (*reflexGate, *time.Time) {
	clock := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	g := newReflexGate(social, nil)
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

// --- mentor bypass: the hostcfg operator is exempt ---------------------------

// TestReflexGateMentorExemptFromThrottleAndBreaker proves the MENTOR BYPASS:
// the configured operator (matched case-insensitively through normalizeSpeaker)
// is never cooldown-throttled and never trips the loop-breaker, while a regular
// speaker on the same gate keeps the normal governor.
func TestReflexGateMentorExemptFromThrottleAndBreaker(t *testing.T) {
	clock := time.Date(2026, 6, 11, 12, 0, 0, 0, time.UTC)
	g := newReflexGate(0, func() string { return "Alex" }) // hostcfg operator, mixed case
	g.now = func() time.Time { return clock }
	pos := world.Coord{X: 100, Y: 500} // never moves: no world progress

	// Callers pass normalizeSpeaker'd names ("alex") — the fold IS the
	// case-insensitivity. A rapid no-progress burst well past reflexLoopN must
	// neither throttle nor trip the breaker.
	for i := 0; i < reflexLoopN+2; i++ {
		if ok, why := g.allow("alex", "small talk, again", "delores"); !ok {
			t.Fatalf("mentor line %d suppressed (%q) — the operator must always be engaged", i, why)
		}
		if g.recordReply("alex", pos) {
			t.Fatal("the loop-breaker must never trip for the mentor")
		}
	}
	if ok, why := g.allow("alex", "still here", "delores"); !ok {
		t.Fatalf("mentor must stay engaged after the burst, got %q", why)
	}
	// A regular speaker on the same gate keeps the normal throttle.
	g.recordReply("bob", pos)
	if ok, why := g.allow("bob", "hello again", "delores"); ok || why != "cooldown" {
		t.Fatalf("a non-mentor speaker must still be throttled, got ok=%v why=%q", ok, why)
	}
}

// TestReflexGateNoOperatorNoExemption proves nobody is exempt when no operator
// is configured (nil provider — the historical default — and the empty name).
func TestReflexGateNoOperatorNoExemption(t *testing.T) {
	g, _ := newTestGate(0) // nil provider
	pos := world.Coord{X: 100, Y: 500}
	g.recordReply("alex", pos)
	if ok, _ := g.allow("alex", "hello again", "delores"); ok {
		t.Fatal("with no operator configured (nil provider), nobody is exempt")
	}
	g.operator = func() string { return "" } // configured-but-empty: still nobody
	if ok, _ := g.allow("alex", "hello once more", "delores"); ok {
		t.Fatal("an empty operator name must not read as an exemption")
	}
}

// --- the C-25 chat knowledge slice -------------------------------------------

// TestChatKnownFactsMatchAndThreshold proves the slice matcher: subjects the
// line keyword-touches attach with their best belief's claim + provenance and
// the ledger confidence; sub-floor suspicions (guessing, < epConfFloor) and
// untouched subjects do not. Whole-subject containment outranks a partial
// token hit; confidence breaks score ties.
func TestChatKnownFactsMatchAndThreshold(t *testing.T) {
	h := newTestHost()
	h.knowledge.Note("pickaxe", "item", "Nurmof sells pickaxes upstairs", knowledge.ProvSystem, 0.95)
	h.knowledge.Note("bronze pickaxe", "item", "there are bronze pickaxes in the barbarian village", knowledge.ProvObserved, 0.9)
	h.knowledge.Note("pickaxe handle", "item", "carved from oak", knowledge.ProvObserved, 0.8)
	// Sub-floor suspicion must NOT attach (guessing isn't knowing).
	h.knowledge.Note("bronze bar", "item", "maybe smelts near the bronze mine", knowledge.ProvHearsay, 0.3)
	// A held but untouched subject must NOT attach.
	h.knowledge.Note("lobster", "item", "heals well when cooked", knowledge.ProvObserved, 0.9)

	got := h.chatKnownFacts("Does anyone know where to get a bronze pickaxe?")
	if len(got) != 3 {
		t.Fatalf("want exactly the 3 touched, held subjects, got %d: %+v", len(got), got)
	}
	// Both whole-subject hits (score 1.0) outrank the partial token hit; within
	// the tie, higher confidence first.
	if got[0].Subject != "pickaxe" || got[1].Subject != "bronze pickaxe" || got[2].Subject != "pickaxe handle" {
		t.Fatalf("ranking wrong (want pickaxe, bronze pickaxe, pickaxe handle): %+v", got)
	}
	if got[1].Claim != "there are bronze pickaxes in the barbarian village" {
		t.Fatalf("the best belief's claim must ride the fact: %+v", got[1])
	}
	if got[0].Provenance != knowledge.ProvSystem || got[0].Confidence != 0.95 {
		t.Fatalf("confidence+provenance must round-trip from the ledger: %+v", got[0])
	}
	for _, k := range got {
		if k.Subject == "bronze bar" {
			t.Fatalf("a sub-floor suspicion must not attach: %+v", got)
		}
		if k.Subject == "lobster" {
			t.Fatalf("an untouched subject must not attach: %+v", got)
		}
	}
}

// TestChatKnownFactsClaimKeyedRecall proves the claim-text recall arm: NPC-,
// shop-, and plural-keyed knowledge whose SUBJECT never matches the question
// still attaches when the best belief's CLAIM answers it. These shapes are the
// norm, not the exception — writebackClaims falls back to subj=speaker and
// ExtractDialog emits npc/shop subjects (the in-repo cron_insight fixture is
// exactly subject="Bob's Axes" claim="sells pickaxes").
func TestChatKnownFactsClaimKeyedRecall(t *testing.T) {
	h := newTestHost()
	// NPC-keyed: subject is the speaker, the claim carries the item ("picks" →
	// plural-fold "pick" rides inside "pickaxe").
	h.knowledge.Note("Baraek", "npc", "Baraek sells picks at 20 coin each", knowledge.ProvHearsay, 0.8)
	// Shop-keyed: the cron_insight fixture shape.
	h.knowledge.Note("Bob's Axes", "shop", "sells pickaxes", knowledge.ProvSystem, 0.85)
	// Plural-keyed subject: singular "pickaxe" in the message must fold to it.
	h.knowledge.Note("pickaxes", "item", "Nurmof sells pickaxes at the dwarven mine", knowledge.ProvObserved, 0.9)

	got := h.chatKnownFacts("Does anyone know where to get a bronze pickaxe?")
	if len(got) != 3 {
		t.Fatalf("all three claim/plural-keyed facts must attach, got %d: %+v", len(got), got)
	}
	// The plural subject is a (folded) whole-subject hit — strongest first.
	if got[0].Subject != "pickaxes" {
		t.Fatalf("the plural-keyed whole-subject hit must rank first: %+v", got)
	}
	want := map[string]bool{"Baraek": true, "Bob's Axes": true, "pickaxes": true}
	for _, k := range got {
		if !want[k.Subject] {
			t.Fatalf("unexpected attachment %q: %+v", k.Subject, got)
		}
	}
}

// TestChatSubjectScoreWordBoundaries proves short subjects match on word
// boundaries only: "tin" must not ride inside "waiting" or "shooting", "ore"
// must not ride inside "more" — but a real mention still scores 1.0.
func TestChatSubjectScoreWordBoundaries(t *testing.T) {
	for _, tc := range []struct {
		msg, subj string
		want      float64
	}{
		{"i'm waiting for my mate", "tin", 0},                   // the reflex.go:399 comment's own short subject
		{"that armour is shooting good", "tin", 0},              // 'shoo-tin-g'
		{"tell me more", "ore", 0},                              // 'm-ore'
		{"got any tin?", "tin", 1},                              // real mention, punctuation-adjacent
		{"is that tin ore you have", "ore", 1},                  // real mention mid-line
		{"where to get a bronze pickaxe?", "bronze pickaxe", 1}, // multi-word phrase
	} {
		if got := chatSubjectScore(tc.msg, tc.subj); got != tc.want {
			t.Errorf("chatSubjectScore(%q, %q) = %v, want %v", tc.msg, tc.subj, got, tc.want)
		}
	}
	// End-to-end: a held short subject must not attach to small talk...
	h := newTestHost()
	h.knowledge.Note("tin", "item", "tin rocks sit southwest of the Varrock gate", knowledge.ProvObserved, 0.9)
	if got := h.chatKnownFacts("I'm waiting for my mate"); got != nil {
		t.Fatalf("'tin' must not attach inside 'waiting', got %+v", got)
	}
	// ...and must attach to a genuine mention.
	if got := h.chatKnownFacts("got any tin?"); len(got) != 1 || got[0].Subject != "tin" {
		t.Fatalf("a real 'tin' mention must attach, got %+v", got)
	}
}

// TestChatKnownFactsTopKCap proves the slice stays a slice: more than
// chatKnownTopK touched subjects attach only the strongest chatKnownTopK.
func TestChatKnownFactsTopKCap(t *testing.T) {
	h := newTestHost()
	subjects := []string{"copper", "copper ore", "copper rock", "copper mine", "copper vein", "copper seller", "copper smelting"}
	for i, s := range subjects {
		h.knowledge.Note(s, "item", "a fact about "+s, knowledge.ProvObserved, 0.6+float64(i)*0.05)
	}
	got := h.chatKnownFacts("where do I find copper around here?")
	if len(got) != chatKnownTopK {
		t.Fatalf("the slice must cap at top-%d, got %d: %+v", chatKnownTopK, len(got), got)
	}
}

// TestChatKnownFactsEmptyWhenNothingTouches proves the empty slice IS the
// signal (mesad reads "no facts" as "factual questions get silence"): an
// unknown topic, an empty message, and a nil ledger all yield nil.
func TestChatKnownFactsEmptyWhenNothingTouches(t *testing.T) {
	h := newTestHost()
	h.knowledge.Note("lobster", "item", "heals well when cooked", knowledge.ProvObserved, 0.9)
	if got := h.chatKnownFacts("where can I learn smithing?"); got != nil {
		t.Fatalf("an untouched ledger must attach nothing, got %+v", got)
	}
	if got := h.chatKnownFacts(""); got != nil {
		t.Fatalf("an empty line must attach nothing, got %+v", got)
	}
	h.knowledge = nil
	if got := h.chatKnownFacts("where can I get a lobster?"); got != nil {
		t.Fatalf("a nil ledger must attach nothing, got %+v", got)
	}
}
