package runtime

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

// fakeAskClient is a mesa Client that returns a canned Ask/Chat reply so the
// host-side speech gate (predicate → compose → say) is driven offline. It embeds
// StubClient and records the last Ask call so tests can assert WHEN it fired and
// WHAT context was passed. healthy gates the offline-skip in tryAsk.
type fakeAskClient struct {
	mesaclient.StubClient
	healthy bool

	askReply string // returned text; "" ⇒ speak=false (silence)
	askCalls int
	askTo    string
	askQ     string
	askCtx   []string

	chatReply string
	chatCalls int
	chatCtx   []string // the recent/context the reply call received (grounding)
}

func (f *fakeAskClient) Healthy() bool { return f.healthy }

func (f *fakeAskClient) Ask(_ context.Context, _, target, question string, recent []string) (string, bool, error) {
	f.askCalls++
	f.askTo, f.askQ, f.askCtx = target, question, recent
	if strings.TrimSpace(f.askReply) == "" {
		return "", false, nil
	}
	return f.askReply, true, nil
}

func (f *fakeAskClient) Chat(_ context.Context, _, _, _ string, recent []string) (string, bool, error) {
	f.chatCalls++
	f.chatCtx = recent
	if strings.TrimSpace(f.chatReply) == "" {
		return "", false, nil
	}
	return f.chatReply, true, nil
}

// speechTestHost builds a host wired for speech tests: a named NPC type in facts,
// self at a known position, and a healthy fake client configured for analysis.
func speechTestHost(t *testing.T, fake *fakeAskClient) *Host {
	t.Helper()
	h := newTestHost()
	h.opts.Username = "Delores"
	h.world.Players.SetSelfName("Delores")
	h.facts = &facts.Facts{NpcDefs: map[int]*facts.NpcDef{
		4: {ID: 4, Name: "Nurmof"},
	}}
	h.configureAnalysis("Operator", fake, nil)
	// Capture spoken lines without a live socket (the ASK emission seam).
	h.emitSay = func(_ context.Context, msg string) error {
		h.reactiveObserveSelf(msg) // preserve the self-line fan-in the real Say does
		return nil
	}
	return h
}

// placeNpc seeds a named NPC (type 4 = "Nurmof") at (x,y).
func placeNpc(h *Host, index, x, y int) {
	h.world.Npcs.Set(world.NpcRecord{Index: index, X: x, Y: y, TypeID: 4})
}

// placePlayer seeds a named player at (x,y).
func placePlayer(h *Host, index int, name string, x, y int) {
	h.world.Players.SetPosition(index, x, y, 0)
	h.world.Players.SetName(index, name)
}

// openQ seeds an open-question node and returns its id.
func openQ(h *Host, id, label string) string {
	h.goalGraph.Upsert(id, goalgraph.KindOpenQuestion, label, goalgraph.StatusOpen)
	return id
}

// --- Deliverable 1 + 3: the ASK gate truth table ----------------------------

func TestTryAskFiresOnEligibleQuestionAndInterlocutor(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "Where do you sell pickaxes?"}
	h := speechTestHost(t, fake)
	openQ(h, "q-pickaxe", "where to buy a pickaxe")
	placeNpc(h, 1, 121, 504) // self is at (120,504): within askRadius

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")

	if fake.askCalls != 1 {
		t.Fatalf("expected exactly one Ask, got %d", fake.askCalls)
	}
	if fake.askTo != "Nurmof" {
		t.Fatalf("asked the wrong target: %q (want Nurmof)", fake.askTo)
	}
	// The ask must burn the cooldowns + tag the question for the inspector.
	n, _ := h.goalGraph.Get("q-pickaxe")
	if !contains(n.Tags, "asked:Nurmof") {
		t.Fatalf("question not tagged with the ask evidence: %v", n.Tags)
	}
}

func TestTryAskSilentWhenFrozen(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "hi?"}
	h := speechTestHost(t, fake)
	openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)
	h.EnterAnalysis() // operator override freezes ASK

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 0 {
		t.Fatalf("ASK must be frozen under analysis mode, got %d calls", fake.askCalls)
	}
}

func TestTryAskSilentWithNoOpenQuestion(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "hi?"}
	h := speechTestHost(t, fake)
	placeNpc(h, 1, 120, 504) // interlocutor present but nothing to ask
	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 0 {
		t.Fatalf("no open question → no ask, got %d", fake.askCalls)
	}
}

func TestTryAskSilentWhenQuestionAlreadyKnown(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "hi?"}
	h := speechTestHost(t, fake)
	openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)
	// Host already KNOWS about pickaxe at high confidence → don't ask.
	h.knowledge.Note("pickaxe", "item", "sold at Nurmof's", knowledge.ProvSystem, 0.9)

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 0 {
		t.Fatalf("must not ask what it already knows, got %d", fake.askCalls)
	}
}

func TestTryAskSilentWithNoInterlocutorInRange(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "hi?"}
	h := speechTestHost(t, fake)
	openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 200, 600) // far outside askRadius
	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 0 {
		t.Fatalf("no in-range interlocutor → no ask, got %d", fake.askCalls)
	}
}

func TestTryAskSilentWhenComposerStaysSilent(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: ""} // composer returns speak=false
	h := speechTestHost(t, fake)
	openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 1 {
		t.Fatalf("composer should be consulted once, got %d", fake.askCalls)
	}
	// A silent composer must NOT burn the per-question cooldown — still askable.
	if _, ok := h.speech.askedQ["q1"]; ok {
		t.Fatal("a silent composer should not record an ask (per-question cooldown not burned)")
	}
	// ...but the GLOBAL gap IS burned, so a silent/flaky composer cannot tight-loop
	// the compose RPC on every tick (anti-spam).
	if h.speech.lastAsk.IsZero() {
		t.Fatal("a compose attempt must burn the global ask floor even on silence")
	}
}

// TestTryAskEmitFailureBurnsGapNotQuestion proves the atomicity fix: when the send
// fails after a successful compose, the global gap is burned (no tight-loop), but
// the per-question cooldown is NOT (the question stays askable) and no orphaned
// latch is left on the target.
func TestTryAskEmitFailureBurnsGapNotQuestion(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "Where do you sell pickaxes?"}
	h := speechTestHost(t, fake)
	h.emitSay = func(_ context.Context, _ string) error { return errors.New("socket closed") }
	openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 1 {
		t.Fatalf("composer should be consulted once, got %d", fake.askCalls)
	}
	if _, ok := h.speech.askedQ["q1"]; ok {
		t.Fatal("a failed send must NOT burn the per-question cooldown (question stays askable)")
	}
	if h.speech.lastAsk.IsZero() {
		t.Fatal("a failed send must still burn the global gap (prevents a tight compose loop)")
	}
	if h.reactive.latchCount() != 0 {
		t.Fatalf("a failed send must not leave an orphaned latch, got %d", h.reactive.latchCount())
	}
}

// --- Deliverable 3: anti-spam (cooldowns + attempt cap) ----------------------

func TestTryAskGlobalGapSuppressesBackToBack(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "q?"}
	h := speechTestHost(t, fake)
	openQ(h, "q1", "where to buy a pickaxe")
	openQ(h, "q2", "where to find an anvil") // a different question
	placeNpc(h, 1, 120, 504)
	placePlayer(h, 5, "Smith", 121, 504)

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 1 {
		t.Fatalf("first ask should fire, got %d", fake.askCalls)
	}
	// Immediately again: the global ask floor must suppress it.
	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 1 {
		t.Fatalf("global ask gap must suppress a back-to-back ask, got %d", fake.askCalls)
	}
}

func TestPickInterlocutorSameQSameTargetCooldown(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	id := openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)
	q, _ := h.goalGraph.Get(id)

	now := time.Now()
	tgt, ok := h.pickInterlocutor(q, now)
	if !ok || tgt.name != "Nurmof" {
		t.Fatalf("first pick should select Nurmof, got %+v ok=%v", tgt, ok)
	}
	// Record that we just asked THIS question of Nurmof.
	h.speech.askedQTo[id+"|Nurmof"] = now
	if _, ok := h.pickInterlocutor(q, now); ok {
		t.Fatal("must not re-pick the same target for the same question within cooldown")
	}
	// After the same-Q-same-target cooldown lapses, it's eligible again.
	if _, ok := h.pickInterlocutor(q, now.Add(askSameQSameTargetCooldown+time.Second)); !ok {
		t.Fatal("target should be eligible again once the cooldown lapses")
	}
}

func TestPickAskQuestionAttemptCapExhausts(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	id := openQ(h, "q1", "how-to-progress: stuck")
	now := time.Now()

	// Below the cap (and not on per-question cooldown): eligible.
	if _, ok := h.pickAskQuestion(now); !ok {
		t.Fatal("a fresh question should be eligible")
	}
	// At the attempt cap: skipped permanently (until a cron re-arms).
	h.speech.attempts[id] = maxAskAttempts
	if _, ok := h.pickAskQuestion(now); ok {
		t.Fatal("an attempt-exhausted question must be skipped")
	}
}

func TestTryAskTagsExhaustedAtCap(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "q?"}
	h := speechTestHost(t, fake)
	id := openQ(h, "q1", "where to buy a pickaxe")
	placeNpc(h, 1, 120, 504)
	// Pre-load attempts to one below the cap so this ask trips the hard stop.
	h.speech.attempts[id] = maxAskAttempts - 1

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")
	if fake.askCalls != 1 {
		t.Fatalf("expected the final ask to fire, got %d", fake.askCalls)
	}
	n, _ := h.goalGraph.Get(id)
	if !contains(n.Tags, "ask-exhausted") {
		t.Fatalf("question hitting the attempt cap must be tagged ask-exhausted: %v", n.Tags)
	}
}

// --- Deliverable 2: knowledge-grounded reply --------------------------------

func TestGroundReplyKnownSubjectPassesFact(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	h.knowledge.Note("pickaxe", "item", "Nurmof sells iron pickaxes", knowledge.ProvSystem, 0.9)

	lines := h.groundReply("Smith", "where can I get a pickaxe?", time.Now())
	if !anyContains(lines, "You KNOW") || !anyContains(lines, "Nurmof sells iron pickaxes") {
		t.Fatalf("a known subject should pass a grounded 'You KNOW' fact, got %v", lines)
	}
}

func TestGroundReplyUnknownSubjectForcesHonesty(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	// Nothing known about "dragon".
	lines := h.groundReply("Smith", "where is the dragon?", time.Now())
	if !anyContains(lines, "do NOT actually know") {
		t.Fatalf("an unknown subject must pass an honest 'do NOT know' line, got %v", lines)
	}
}

func TestGroundReplyVolunteersHighConfidenceTeach(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	// A high-confidence belief whose SUBJECT the player's line touches, distinct
	// from the line's salient topic (which becomes the direct answer). The player
	// asks about "smithing"; the host volunteers the related "anvil" belief.
	h.knowledge.Note("smithing", "skill", "smithing makes weapons from bars", knowledge.ProvObserved, 0.9)
	h.knowledge.Note("anvil", "object", "there is an anvil in Lumbridge", knowledge.ProvObserved, 0.85)

	lines := h.groundReply("Smith", "tell me about smithing and the anvil please", time.Now())
	if !anyContains(lines, "could mention") || !anyContains(lines, "anvil in Lumbridge") {
		t.Fatalf("a touched high-confidence belief (other than the direct answer) should be volunteered, got %v", lines)
	}
	// The teach cooldown suppresses a second volunteer to the same player.
	lines2 := h.groundReply("Smith", "tell me about smithing and the anvil please", time.Now())
	if anyContains(lines2, "could mention") {
		t.Fatal("teach cooldown must suppress a back-to-back volunteer to the same player")
	}
}

// --- loop closure: closeResolvedQuestions -----------------------------------

func TestCloseResolvedQuestionsHighConfClosesAndUnblocks(t *testing.T) {
	h := newTestHost()
	h.goalGraph.Upsert("goal", goalgraph.KindGoal, "smith a dagger", goalgraph.StatusBlocked)
	h.goalGraph.Upsert("q-pickaxe", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	h.goalGraph.Link("goal", "q-pickaxe", goalgraph.RelBlockedBy) // goal --blocked_by--> q

	// An authoritative claim (>= closeQConf) that answers the question.
	h.closeResolvedQuestions([]mesaclient.DialogClaim{
		{Subject: "pickaxe", Kind: "item", Claim: "you buy a pickaxe at Nurmof's", Confidence: 0.85},
	})

	q, _ := h.goalGraph.Get("q-pickaxe")
	if q.Status != goalgraph.StatusDone {
		t.Fatalf("matching high-conf claim should close the question, got %q", q.Status)
	}
	if !contains(q.Tags, "resolved-by-ask") {
		t.Fatalf("closed question should be tagged resolved-by-ask: %v", q.Tags)
	}
	g, _ := h.goalGraph.Get("goal")
	if g.Status != goalgraph.StatusActive {
		t.Fatalf("the blocked goal should be un-blocked to active, got %q", g.Status)
	}
}

func TestCloseResolvedQuestionsLowConfDoesNotClose(t *testing.T) {
	h := newTestHost()
	h.goalGraph.Upsert("q1", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	// Flimsy hearsay (< closeQConf) must NOT auto-close.
	h.closeResolvedQuestions([]mesaclient.DialogClaim{
		{Subject: "pickaxe", Claim: "maybe a pickaxe is somewhere", Confidence: 0.5},
	})
	q, _ := h.goalGraph.Get("q1")
	if q.Status == goalgraph.StatusDone {
		t.Fatal("low-confidence hearsay must NOT auto-close an open question")
	}
}

func TestCloseResolvedQuestionsNoMatchNoChange(t *testing.T) {
	h := newTestHost()
	h.goalGraph.Upsert("q1", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	// A high-conf claim about something unrelated must not touch the question.
	h.closeResolvedQuestions([]mesaclient.DialogClaim{
		{Subject: "fishing", Claim: "lobsters need 40 fishing", Confidence: 0.9},
	})
	q, _ := h.goalGraph.Get("q1")
	if q.Status == goalgraph.StatusDone {
		t.Fatal("an unrelated claim must not close the question")
	}
}

// --- small helpers ----------------------------------------------------------

func contains(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}

func anyContains(xs []string, sub string) bool {
	for _, x := range xs {
		if strings.Contains(x, sub) {
			return true
		}
	}
	return false
}
