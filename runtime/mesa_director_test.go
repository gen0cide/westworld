package runtime

import (
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/dsl/interp"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
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

// --- Phase-2a reasoning layer ----------------------------------------------

const epGoal = "do a thing"

// hintsFor runs one situation() turn and returns the resulting Hints map, so the
// epistemic block tests assert against the same path the planner sees.
func hintsFor(h *Host, d *MesaDirector) map[string]string {
	sit := d.situation(h, Outcome{})
	if sit.Hints == nil {
		return map[string]string{}
	}
	return sit.Hints
}

// TestEpistemicIntentionSlice proves Deliverables 1+3: the active goal's one-hop
// slice renders its blocker, its open sub-goal, and the downstream value reached
// through a produced output (mine --produces--> ore, ore --serves--> smithing) —
// a single extra hop, not a flat goal line.
func TestEpistemicIntentionSlice(t *testing.T) {
	h := newTestHost()
	d := quietDirector() // goal "do a thing"
	g := epGoal
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	// requires sub-goal (open), a blocker, and downstream value through produces.
	h.goalGraph.Upsert("a pickaxe", goalgraph.KindSubgoal, "a pickaxe", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "a pickaxe", goalgraph.RelRequires)
	h.goalGraph.Upsert("a bear keeps killing me", goalgraph.KindOpenQuestion, "a bear keeps killing me", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "a bear keeps killing me", goalgraph.RelBlockedBy)
	h.goalGraph.Upsert("ore", goalgraph.KindState, "ore", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "ore", goalgraph.RelProduces)
	h.goalGraph.Upsert("smithing", goalgraph.KindGoal, "smithing", goalgraph.StatusOpen)
	h.goalGraph.Link("ore", "smithing", goalgraph.RelServes)

	it := hintsFor(h, d)["intention"]
	if it == "" {
		t.Fatal("a goal with edges must surface an intention hint")
	}
	for _, want := range []string{"a pickaxe", "a bear keeps killing me", "smithing", "BLOCKED"} {
		if !strings.Contains(it, want) {
			t.Fatalf("intention hint missing %q:\n%s", want, it)
		}
	}
}

// TestEpistemicIntentionBareRootIsSilent proves a bare root goal (no edges)
// surfaces NO intention hint — the epistemic blocks stay silent when there is
// nothing to surface, so the happy path has no prompt bloat.
func TestEpistemicIntentionBareRootIsSilent(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	if it := hintsFor(h, d)["intention"]; it != "" {
		t.Fatalf("a bare root goal must surface no intention hint; got:\n%s", it)
	}
}

// TestEpistemicSliceStaysBounded proves the slice respects its caps regardless of
// graph size — the host READS+SURFACES a bounded slice, it does not dump the graph.
func TestEpistemicSliceStaysBounded(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	g := epGoal
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	for i := 0; i < 10; i++ {
		sub := "subgoal-" + string(rune('A'+i))
		h.goalGraph.Upsert(sub, goalgraph.KindSubgoal, sub, goalgraph.StatusOpen)
		h.goalGraph.Link(g, sub, goalgraph.RelRequires)
		blk := "blocker-" + string(rune('A'+i))
		h.goalGraph.Upsert(blk, goalgraph.KindOpenQuestion, blk, goalgraph.StatusOpen)
		h.goalGraph.Link(g, blk, goalgraph.RelBlockedBy)
	}
	subs := d.sliceSubgoals(h, g)
	if len(subs) > epSubgoalCap {
		t.Fatalf("sub-goal slice exceeded cap: %d > %d", len(subs), epSubgoalCap)
	}
	blks := d.sliceBlockers(h, g)
	if len(blks) > epBlockerCap {
		t.Fatalf("blocker slice exceeded cap: %d > %d", len(blks), epBlockerCap)
	}
}

// TestEpistemicBeliefsAndUnknowns proves Deliverable 1: high-confidence beliefs
// about subjects in front of the host land in "known", low-confidence ones route
// to "unknowns", and a goal-blocking open question is tagged [BLOCKS YOUR GOAL].
func TestEpistemicBeliefsAndUnknowns(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	g := epGoal
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	// A goal-blocking open question (becomes a [BLOCKS YOUR GOAL] unknown).
	h.goalGraph.Upsert("where to buy a pickaxe", goalgraph.KindOpenQuestion, "where to buy a pickaxe", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "where to buy a pickaxe", goalgraph.RelBlockedBy)

	// High-confidence belief about a subject we'll put in front of her.
	h.knowledge.Note("Bob", "npc", "sells axes, not pickaxes", knowledge.ProvObserved, 0.9)
	// Low-confidence belief about another subject.
	h.knowledge.Note("the mine", "location", "might be dangerous", knowledge.ProvHearsay, 0.2)

	// The relevance set = subjects in front of her ∪ the labels her goal slice
	// names. Surface both subjects by linking them as requires (sub-goal labels).
	h.goalGraph.Upsert("Bob", goalgraph.KindSubgoal, "Bob", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "Bob", goalgraph.RelRequires)
	h.goalGraph.Upsert("the mine", goalgraph.KindSubgoal, "the mine", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "the mine", goalgraph.RelRequires)

	hints := hintsFor(h, d)
	if kn := hints["known"]; !strings.Contains(kn, "Bob") || !strings.Contains(kn, "sells axes") {
		t.Fatalf("high-confidence belief about Bob should be in `known`:\n%s", kn)
	}
	un := hints["unknowns"]
	if !strings.Contains(un, "[BLOCKS YOUR GOAL]") || !strings.Contains(un, "where to buy a pickaxe") {
		t.Fatalf("goal-blocking open question should render [BLOCKS YOUR GOAL]:\n%s", un)
	}
	if !strings.Contains(un, "the mine") {
		t.Fatalf("low-confidence belief should route to unknowns:\n%s", un)
	}
	if strings.Contains(hints["known"], "the mine") {
		t.Fatalf("low-confidence belief must NOT appear in known:\n%s", hints["known"])
	}
}

// TestCuriosityBias proves Deliverable 5: zero persona ⇒ no bias (no panic); a
// core-blocking unknown ⇒ the mandatory resolve gate regardless of dials; a
// pronounced curiosity flavor + a tangent ⇒ the optional detour cue; and a
// low-curiosity persona ⇒ silence.
func TestCuriosityBias(t *testing.T) {
	if s := curiosityBias(persona.Curiosity{}, false, true); s != "" {
		t.Fatalf("zero persona, not blocked ⇒ no bias; got %q", s)
	}
	if s := curiosityBias(persona.Curiosity{Spatial: 0.9}, true, false); !strings.Contains(s, "BLOCKING") {
		t.Fatalf("core-blocked ⇒ mandatory resolve gate; got %q", s)
	}
	if s := curiosityBias(persona.Curiosity{Spatial: 0.8}, false, true); !strings.Contains(s, "detour") {
		t.Fatalf("high curiosity + tangent ⇒ detour cue; got %q", s)
	}
	if s := curiosityBias(persona.Curiosity{Spatial: 0.2, Skill: 0.1}, false, true); s != "" {
		t.Fatalf("low curiosity ⇒ stay on task (no cue); got %q", s)
	}
}

// TestCuriosityReadAtDecisionTime proves Deliverable 5's wiring: the director
// reads h.curiosity (captured at bootstrap) at decision time and surfaces a
// learning_priority detour cue for a curious persona when an unknown does NOT
// block the goal (a tangent exists via downstream value). A neutral host gets no
// cue, so the field is genuinely load-bearing.
func TestCuriosityReadAtDecisionTime(t *testing.T) {
	g := epGoal
	setup := func(cu persona.Curiosity) map[string]string {
		h := newTestHost()
		d := quietDirector()
		h.curiosity = cu
		h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
		// A tangential (non-blocking) downstream value so a curious host has
		// something to detour toward, but the goal itself is NOT blocked.
		h.goalGraph.Upsert("smithing", goalgraph.KindGoal, "smithing", goalgraph.StatusOpen)
		h.goalGraph.Link(g, "smithing", goalgraph.RelServes)
		return hintsFor(h, d)
	}
	if lp := setup(persona.Curiosity{Spatial: 0.8})["learning_priority"]; !strings.Contains(lp, "detour") {
		t.Fatalf("curious persona should get a detour cue; got %q", lp)
	}
	if lp := setup(persona.Curiosity{})["learning_priority"]; lp != "" {
		t.Fatalf("neutral persona should get no learning-priority cue; got %q", lp)
	}
}

// TestFailureSpawnsEnablingSubgoal proves Deliverable 2: a hard fail-streak spawns
// exactly one enabler sub-goal (with requires/enables edges) alongside the open
// question, is idempotent on re-fire, and is marked done on a later success.
func TestFailureSpawnsEnablingSubgoal(t *testing.T) {
	h := newTestHost()
	d := quietDirector() // goal "do a thing"
	g := epGoal

	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	for range antiStuckHardFails {
		d.situation(h, fail)
	}

	eid := "enabler:" + g
	if !h.goalGraph.Has(eid) {
		t.Fatal("a hard fail-streak should spawn an enabler sub-goal")
	}
	en, _ := h.goalGraph.Get(eid)
	if en.Kind != goalgraph.KindSubgoal || en.Status != goalgraph.StatusOpen {
		t.Fatalf("enabler node = %+v, want KindSubgoal/StatusOpen", en)
	}
	// goal --requires--> enabler AND enabler --enables--> goal.
	if reqs := h.goalGraph.Out(g, goalgraph.RelRequires); len(reqs) != 1 || reqs[0].To != eid {
		t.Fatalf("expected goal --requires--> enabler; got %v", reqs)
	}
	if ens := h.goalGraph.Out(eid, goalgraph.RelEnables); len(ens) != 1 || ens[0].To != g {
		t.Fatalf("expected enabler --enables--> goal; got %v", ens)
	}

	// Idempotency: another failure must NOT create a second enabler or new edges.
	nodesBefore := len(h.goalGraph.Nodes())
	edgesBefore := len(h.goalGraph.Edges())
	d.situation(h, fail)
	if got := len(h.goalGraph.Nodes()); got != nodesBefore {
		t.Fatalf("re-fire created nodes: %d -> %d", nodesBefore, got)
	}
	if got := len(h.goalGraph.Edges()); got != edgesBefore {
		t.Fatalf("re-fire created edges: %d -> %d", edgesBefore, got)
	}

	// Success: goal flips active AND the enabler flips done.
	d.situation(h, Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultCompleted})
	if n, _ := h.goalGraph.Get(g); n.Status != goalgraph.StatusActive {
		t.Fatalf("success should un-block the goal; status=%q", n.Status)
	}
	if en, _ := h.goalGraph.Get(eid); en.Status != goalgraph.StatusDone {
		t.Fatalf("success should mark the enabler done; status=%q", en.Status)
	}
}

// TestHappyPathNoEnablerNoIntention proves a working grind (successes reset the
// streak) writes NO enabler node and surfaces an empty intention hint — the
// epistemic blocks are silent when there is nothing to surface.
func TestHappyPathNoEnablerNoIntention(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	ok := Outcome{Intent: Intent{Label: "act:grind", Name: "grind"}, Kind: interp.ResultCompleted}
	for range 6 {
		d.situation(h, ok)
	}
	if h.goalGraph.Has("enabler:" + epGoal) {
		t.Fatal("a working grind must not spawn an enabler")
	}
	if it := hintsFor(h, d)["intention"]; it != "" {
		t.Fatalf("a working grind should surface no intention hint; got:\n%s", it)
	}
}
