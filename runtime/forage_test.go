package runtime

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
)

// --- 5b-1: topical open-question minting ------------------------------------

func TestAcquireItemSubject(t *testing.T) {
	cases := []struct{ goal, want string }{
		{"acquire a pickaxe", "pickaxe"},
		{"buy bronze bars", "bronze bars"},
		{"get an iron pickaxe", "iron pickaxe"},
		{"obtain the lobster pot", "lobster pot"},
		{"explore Varrock", ""},          // no acquire verb
		{"have all the bronze bars", ""}, // " have a " is not a substring (M6 guard)
		{"train mining to level 40", ""}, // no acquire verb
	}
	for _, c := range cases {
		if got := acquireItemSubject(c.goal); got != c.want {
			t.Errorf("acquireItemSubject(%q) = %q, want %q", c.goal, got, c.want)
		}
	}
}

func TestTopicalLabelSalientTopicResolves(t *testing.T) {
	// The clean Label AND the bare qid must both resolve to the subject, so the
	// picker and the closer agree on the topic whichever they read (P2 #8).
	if got := salientTopic(topicalLabel("where-to-buy", "pickaxe")); got != "pickaxe" {
		t.Fatalf("salientTopic(topicalLabel) = %q, want pickaxe", got)
	}
	if got := salientTopic("where-to-buy:pickaxe"); got != "pickaxe" {
		t.Fatalf("salientTopic(qid) = %q, want pickaxe", got)
	}
}

func TestMarkGoalBlockedMintsTopicalAndSuppressesHowToProgress(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	d := NewMesaDirector(fake, "h", "", nil)

	g := "acquire a pickaxe"
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	d.markGoalBlocked(h, g, "3 attempts failed")

	// A subject-bearing where-to-buy question blocks the goal; the subjectless
	// how-to-progress is suppressed for the acquirable-item goal.
	if !h.goalGraph.Has("where-to-buy:pickaxe") {
		t.Fatal("where-to-buy:pickaxe not minted")
	}
	if h.goalGraph.Has("how-to-progress:" + g) {
		t.Fatal("how-to-progress must be suppressed for an acquirable-item goal")
	}
	blocked := false
	for _, b := range h.goalGraph.Blockers(g) {
		if b.ID == "where-to-buy:pickaxe" {
			blocked = true
		}
	}
	if !blocked {
		t.Fatal("goal is not blocked_by where-to-buy:pickaxe")
	}

	// A NON-acquire goal keeps the existing how-to-progress path.
	g2 := "explore Varrock"
	h.goalGraph.Upsert(g2, goalgraph.KindGoal, g2, goalgraph.StatusActive)
	d.markGoalBlocked(h, g2, "stuck")
	if !h.goalGraph.Has("how-to-progress:" + g2) {
		t.Fatal("non-acquire goal must still mint how-to-progress")
	}
}

func TestMarkTopicalQuestionM8DoesNotReopen(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	d := NewMesaDirector(fake, "h", "", nil)

	g := "buy a pickaxe"
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	d.markGoalBlocked(h, g, "blocked")
	h.goalGraph.SetStatus("where-to-buy:pickaxe", goalgraph.StatusDone) // a closer resolved it
	d.markGoalBlocked(h, g, "blocked again")                            // re-touch

	if n, _ := h.goalGraph.Get("where-to-buy:pickaxe"); n.Status != goalgraph.StatusDone {
		t.Fatalf("re-touch must not re-open a resolved question; status=%s", n.Status)
	}
}

// --- 5b-1: P0 #1 — exhaustion as TAGS, never new graph nodes -----------------

func TestNoStateNodesMinted(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	d := NewMesaDirector(fake, "h", "", nil)

	g := "acquire a pickaxe"
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	d.markGoalBlocked(h, g, "blocked")
	h.goalGraph.Tag("where-to-buy:pickaxe", "source-tried:npc:Nurmof") // exhaustion = a tag on Q

	for _, n := range h.goalGraph.Nodes() {
		if n.Kind == goalgraph.KindState &&
			(strings.HasPrefix(n.ID, "src:") || strings.HasPrefix(n.ID, "hypothesis")) {
			t.Fatalf("5b must keep exhaustion on Q.Tags, not mint state nodes; found %q", n.ID)
		}
	}
}

// --- 5b-1: source-tried exhaustion + player-oracle promotion -----------------

func TestSourceTriedNpcSkippedInPicker(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	q := openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	placeNpc(h, 1, 121, 504) // self at (120,504): Nurmof is in range

	qn := goalgraph.Node{ID: q, Label: "where to buy a pickaxe"}
	if _, ok := h.pickInterlocutor(qn, time.Now()); !ok {
		t.Fatal("precondition: Nurmof should be pickable")
	}
	h.goalGraph.Tag(q, "source-tried:npc:Nurmof")
	if tgt, ok := h.pickInterlocutor(qn, time.Now()); ok {
		t.Fatalf("a source-tried NPC must be skipped; picked %q", tgt.name)
	}
}

func TestPlayerPromotedForFactualQuestion(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	placeNpc(h, 1, 121, 504)            // Nurmof
	placePlayer(h, 2, "Hank", 122, 504) // a nearby player

	// Factual where-to-buy: the player (free-form oracle) beats the canned NPC.
	q := goalgraph.Node{ID: "where-to-buy:pickaxe", Label: "where to buy a pickaxe"}
	tgt, ok := h.pickInterlocutor(q, time.Now())
	if !ok || tgt.role != "player" || tgt.name != "Hank" {
		t.Fatalf("factual question should prefer the player; got ok=%v %s %q", ok, tgt.role, tgt.name)
	}

	// A social question keeps the NPC-first default.
	q2 := goalgraph.Node{ID: "how-to-progress:x", Label: "what is the local gossip"}
	tgt2, ok2 := h.pickInterlocutor(q2, time.Now())
	if !ok2 || tgt2.role != "npc" {
		t.Fatalf("social question should prefer the NPC; got ok=%v %s", ok2, tgt2.role)
	}
}

func TestPlayerNotSourceTried(t *testing.T) {
	fake := &fakeAskClient{healthy: true, askReply: "Where can I buy a pickaxe?"}
	h := speechTestHost(t, fake)
	openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	placePlayer(h, 2, "Hank", 122, 504) // only a player in range

	tryAsk(context.Background(), slog.Default(), h, fake, "Delores")

	n, _ := h.goalGraph.Get("where-to-buy:pickaxe")
	for _, tag := range n.Tags {
		if strings.HasPrefix(tag, "source-tried:npc:") {
			t.Fatalf("asking a PLAYER must not write a source-tried:npc tag; tags=%v", n.Tags)
		}
	}
}

// --- 5b-2/5b-3: forage drive ------------------------------------------------

func TestForageTopicPOITypes(t *testing.T) {
	eq := func(got, want []string) bool {
		if len(got) != len(want) {
			return false
		}
		for i := range got {
			if got[i] != want[i] {
				return false
			}
		}
		return true
	}
	if got := forageTopicPOITypes("where to buy a pickaxe"); !eq(got, []string{"shop"}) {
		t.Fatalf("where-to-buy label -> %v, want [shop]", got)
	}
	if got := forageTopicPOITypes("where-to-buy:pickaxe"); !eq(got, []string{"shop"}) {
		t.Fatalf("where-to-buy qid -> %v, want [shop]", got)
	}
	if got := forageTopicPOITypes("where is the furnace"); !eq(got, []string{"furnace"}) {
		t.Fatalf("where-is label -> %v, want [furnace]", got)
	}
	if got := forageTopicPOITypes("is it true about x"); got != nil {
		t.Fatalf("confirm/how-to-progress must map to nil; got %v", got)
	}
}

func TestPickForageTarget(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	g := "acquire a pickaxe"
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	NewMesaDirector(fake, "h", "", nil).markGoalBlocked(h, g, "blocked") // mints where-to-buy:pickaxe, blocks g
	// Make effectiveGoalForSpeech return g (no director bound -> northStar fallback).
	h.northStar = g

	q, ok := h.pickForageTarget(time.Now())
	if !ok || q.ID != "where-to-buy:pickaxe" {
		t.Fatalf("pickForageTarget should return the blocking where-to-buy question; got ok=%v id=%q", ok, q.ID)
	}
	// A non-factual blocker (how-to-progress on a non-acquire goal) is NOT a forage target.
	g2 := "explore the wilderness"
	h.goalGraph.Upsert(g2, goalgraph.KindGoal, g2, goalgraph.StatusActive)
	NewMesaDirector(fake, "h", "", nil).markGoalBlocked(h, g2, "stuck")
	h.northStar = g2
	if _, ok := h.pickForageTarget(time.Now()); ok {
		t.Fatal("a how-to-progress (non-factual) question must NOT be a forage target")
	}
}

func TestPerceiveShopForGoalNegativeWrite(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	qid := openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	h.setForageCursor("pickaxe", "Varrock general store", qid)

	// Catalogue WITHOUT a pickaxe -> negative deduction + source-spent tag.
	h.perceiveShopForGoal("Shopkeeper", map[string]bool{"Bronze Arrows": true, "Bucket": true})

	f := h.knowledge.Get("Shopkeeper")
	found := false
	for _, b := range f.Beliefs {
		if b.Claim == "does not sell pickaxe" {
			found = true
			if b.Provenance != knowledge.ProvDeduced {
				t.Fatalf("negative belief provenance = %q, want deduced", b.Provenance)
			}
			if c := b.Confidence(); c >= 0.5 {
				t.Fatalf("negative belief confidence = %.2f, must be sub-floor (<0.5)", c)
			}
		}
	}
	if !found {
		t.Fatal("expected a 'does not sell pickaxe' deduction")
	}
	if !h.goalGraph.HasTag(qid, "source-spent:place:Varrock general store") {
		t.Fatal("expected source-spent:place tag on the question")
	}
	if h.goalGraph.HasTag(qid, "forage-inflight") {
		t.Fatal("forage-inflight should be cleared after harvest")
	}
	if topic, _, _ := h.forageCursor(); topic != "" {
		t.Fatal("cursor should be cleared after harvest")
	}
}

func TestPerceiveShopForGoalRepeatsStaySubFloor(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	qid := openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	for i := 0; i < 5; i++ {
		h.setForageCursor("pickaxe", "store", qid)
		h.perceiveShopForGoal("Shopkeeper", map[string]bool{"Bucket": true})
	}
	for _, b := range h.knowledge.Get("Shopkeeper").Beliefs {
		if b.Claim == "does not sell pickaxe" && b.Confidence() >= 0.5 {
			t.Fatalf("repeated negative writes must stay sub-floor; got %.2f", b.Confidence())
		}
	}
}

func TestPerceiveShopForGoalPresentNoNegative(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	qid := openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	h.setForageCursor("pickaxe", "store", qid)
	h.perceiveShopForGoal("Bob", map[string]bool{"Bronze Pickaxe": true, "Bronze Axe": true})

	for _, b := range h.knowledge.Get("Bob").Beliefs {
		if strings.HasPrefix(b.Claim, "does not sell") {
			t.Fatal("a shop that stocks the topic must NOT get a negative belief")
		}
	}
	if h.goalGraph.HasTag(qid, "source-spent:place:store") {
		t.Fatal("a stocking shop must NOT be marked source-spent")
	}
}

func TestCloseQuestionByObservation(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	g := "acquire a pickaxe"
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	h.goalGraph.Upsert("where-to-buy:pickaxe", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	h.goalGraph.SetStatus(g, goalgraph.StatusBlocked)
	h.goalGraph.Link(g, "where-to-buy:pickaxe", goalgraph.RelBlockedBy)

	// An OBSERVED in-stock belief closes the question + un-blocks the goal.
	h.knowledge.Note("Bob", "shop", "sells pickaxe", knowledge.ProvObserved, 0.9)
	h.closeQuestionByObservation("Bob", "sells pickaxe")

	if n, _ := h.goalGraph.Get("where-to-buy:pickaxe"); n.Status != goalgraph.StatusDone {
		t.Fatalf("question should be Done; got %s", n.Status)
	}
	if n, _ := h.goalGraph.Get(g); n.Status != goalgraph.StatusActive {
		t.Fatalf("blocked goal should be un-blocked to Active; got %s", n.Status)
	}
}

func TestCloseQuestionByObservationIgnoresDeduced(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	h.goalGraph.Upsert("where-to-buy:pickaxe", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)

	// A DEDUCED negative must NEVER close (provenance gate).
	h.knowledge.Note("Shop", "shop", "does not sell pickaxe", knowledge.ProvDeduced, 0.4)
	h.closeQuestionByObservation("Shop", "does not sell pickaxe")
	if n, _ := h.goalGraph.Get("where-to-buy:pickaxe"); n.Status != goalgraph.StatusOpen {
		t.Fatalf("a deduced negative must not close the question; status=%s", n.Status)
	}
}

func TestForageInflightTTLBackstop(t *testing.T) {
	fake := &fakeAskClient{healthy: true}
	h := speechTestHost(t, fake)
	qid := openQ(h, "where-to-buy:pickaxe", "where to buy a pickaxe")
	h.goalGraph.Tag(qid, "forage-inflight")
	// Record an inflight detour issued longer ago than detourTimeout.
	h.forage.inflight[qid] = time.Now().Add(-2 * detourTimeout)
	h.forage.gcForage(time.Now(), h)
	if h.goalGraph.HasTag(qid, "forage-inflight") {
		t.Fatal("gcForage must TTL-clear a stuck forage-inflight tag (PATH_BLOCKED backstop)")
	}
}

func TestTagsWithPrefix(t *testing.T) {
	g := goalgraph.New()
	g.Upsert("q", goalgraph.KindOpenQuestion, "q", goalgraph.StatusOpen)
	g.Tag("q", "source-spent:place:A")
	g.Tag("q", "source-spent:place:B")
	g.Tag("q", "asked:Hank")
	got := g.TagsWithPrefix("q", "source-spent:place:")
	if len(got) != 2 {
		t.Fatalf("expected 2 source-spent tags, got %v", got)
	}
	if g.TagsWithPrefix("absent", "x") != nil {
		t.Fatal("absent node must return nil (and not be created)")
	}
	if g.Has("absent") {
		t.Fatal("TagsWithPrefix must not create the node")
	}
}
