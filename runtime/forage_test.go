package runtime

import (
	"context"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
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
