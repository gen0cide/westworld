package runtime

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/dsl/interp"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
	"github.com/gen0cide/westworld/world"
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

// TestMesaDirectorBudgetExpiryIsNotAFailure is the soak-retro #5 regression:
// a healthy long grind that outlives the per-turn budget must NOT feed the
// failure streak or escalate to BLOCKED — drone3 accrued a 55-"failure" streak
// (and was ordered to ABANDON its approach) while leveling faster than any
// other host, because deadline expiry was reported identically to real failure.
func TestMesaDirectorBudgetExpiryIsNotAFailure(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	// What the conductor reports when the turn budget cuts off a grind mid-work:
	// an errored Kind (the interpreter surfaced the ctx deadline) but with the
	// BudgetExpired classification set.
	expired := Outcome{
		Intent:        Intent{Label: "act:grind", Name: "grind", Source: "runtime \"1.0\"\nroutine grind() { wait(300) }"},
		Kind:          interp.ResultErrored,
		Err:           &interp.RuntimeError{Msg: "wait: context deadline exceeded"},
		BudgetExpired: true,
	}
	var sit *mesaclient.Situation
	for range antiStuckHardFails + 2 {
		sit = d.situation(h, expired)
	}
	if d.failStreak != 0 {
		t.Fatalf("budget expiry fed the failure streak: failStreak=%d, want 0", d.failStreak)
	}
	if strings.Contains(sit.Trigger, "BLOCKED") {
		t.Fatalf("budget expiry escalated to BLOCKED: %q", sit.Trigger)
	}
	// The planner must be told the truth: out of time, not a failure.
	if r := sit.Hints["last_result"]; strings.Contains(r, "errored") || !strings.Contains(r, "NOT a failure") {
		t.Fatalf("budget expiry surfaced to the planner as a failure: %q", r)
	}

	// Neutrality: a budget expiry must not RESET a real failure streak either —
	// it is evidence of nothing.
	fail := Outcome{Intent: Intent{Label: "act:try_thing", Name: "try_thing"}, Kind: interp.ResultErrored}
	d.situation(h, fail)
	d.situation(h, fail)
	d.situation(h, expired)
	if d.failStreak != 2 {
		t.Fatalf("budget expiry should leave the failure streak untouched; failStreak=%d, want 2", d.failStreak)
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

// TestAcquireVerbsAnchored is the M6 regression: the loose "have a"/"find a"
// substrings used to false-classify unrelated goals as acquire goals ("have all
// the bronze bars" matched "have a"). The anchored verbs fire only on a genuine
// "have a <item>" phrasing.
func TestAcquireVerbsAnchored(t *testing.T) {
	h := New(Options{Username: "test", Facts: pickaxeFacts()})
	h.world.Inventory.Replace([]world.InvSlot{{ItemID: 156, Amount: 1}}) // Bronze Pickaxe
	d := NewMesaDirector(mesaclient.StubClient{}, "tester", "", slog.New(slog.NewTextHandler(io.Discard, nil)))

	// These name a Bronze Pickaxe-ish item via substring but are NOT acquire goals.
	falsePositives := []string{
		"have all the bronze pickaxe parts ready",
		"have any reason to use the bronze pickaxe",
		"find another bronze pickaxe vendor",
	}
	for _, g := range falsePositives {
		if d.goalSatisfiedByInventory(h, g, h.world) {
			t.Fatalf("loose phrasing must NOT classify as acquire/satisfied: %q", g)
		}
	}
	// A genuine acquire goal still satisfies.
	for _, g := range []string{"have a bronze pickaxe", "acquire a bronze pickaxe", "buy a bronze pickaxe"} {
		if !d.goalSatisfiedByInventory(h, g, h.world) {
			t.Fatalf("a genuine acquire goal with the item in hand should satisfy: %q", g)
		}
	}
}

// TestTileRingBounded is the M14 regression: the visited tile history is bounded
// to visitedCap and evicts oldest-first, so an open-ended wanderer never grows the
// per-turn map without bound. The most-recent tiles remain queryable.
func TestTileRingBounded(t *testing.T) {
	r := newTileRing(visitedCap)
	for i := 0; i < visitedCap*4; i++ {
		r.add(i, 0)
	}
	if len(r.set) > visitedCap || len(r.order) > visitedCap {
		t.Fatalf("tileRing exceeded cap: set=%d order=%d cap=%d", len(r.set), len(r.order), visitedCap)
	}
	// The newest tile is present; an old (evicted) one is not.
	if !r.has(visitedCap*4-1, 0) {
		t.Fatal("the most-recent tile should be remembered")
	}
	if r.has(0, 0) {
		t.Fatal("an old tile beyond the cap should have been evicted")
	}
	// Re-adding a recent tile is idempotent (no growth, no re-aging churn).
	before := len(r.order)
	r.add(visitedCap*4-1, 0)
	if len(r.order) != before {
		t.Fatalf("re-adding a present tile must not grow the ring; %d -> %d", before, len(r.order))
	}
}

// TestDoorUsedTracksRecency proves doorUsed still flags a recently-visited door
// via the bounded ring (the behavior M14 preserves), and is safe before any visit.
func TestDoorUsedTracksRecency(t *testing.T) {
	d := quietDirector()
	if d.doorUsed(50, 50) {
		t.Fatal("doorUsed must be false before any tile is visited")
	}
	d.visited = newTileRing(visitedCap)
	d.visited.add(51, 50) // adjacent to the door tile (50,50)
	if !d.doorUsed(50, 50) {
		t.Fatal("a door adjacent to a visited tile should read as used")
	}
}

// TestNoDisplacementHint is the L6 regression: situation() no longer writes a
// "displacement" hint key (the re-orient rides on Trigger; mesad never read the
// hint). A displaced turn surfaces DISPLACED in Trigger but no hints["displacement"].
func TestNoDisplacementHint(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	h.displacement.record(displacementEvent{fromX: 120, fromY: 504, toX: 200, toY: 600, dist: 99, cause: dispJump})
	sit := d.situation(h, Outcome{Intent: Intent{Label: "act:x", Name: "x"}, Kind: interp.ResultCompleted})

	if !strings.Contains(sit.Trigger, "DISPLACED") {
		t.Fatalf("displacement should ride on Trigger; got %q", sit.Trigger)
	}
	if _, ok := sit.Hints["displacement"]; ok {
		t.Fatalf("the dead displacement hint must be gone (L6); hints=%v", sit.Hints)
	}
}

// TestCuriosityTangentExcludesSubgoals is the L10 regression: required sub-goals
// are core-path steps, NOT tangents — a curious persona with only pending required
// sub-goals (and no downstream serves edge) must get NO detour cue.
func TestCuriosityTangentExcludesSubgoals(t *testing.T) {
	h := newTestHost()
	d := quietDirector()
	h.curiosity = persona.Curiosity{Spatial: 0.9} // very curious
	g := epGoal
	h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	// A required sub-goal (core path), and NO serves edge (no real tangent).
	h.goalGraph.Upsert("a pickaxe", goalgraph.KindSubgoal, "a pickaxe", goalgraph.StatusOpen)
	h.goalGraph.Link(g, "a pickaxe", goalgraph.RelRequires)

	if lp := hintsFor(h, d)["learning_priority"]; lp != "" {
		t.Fatalf("required sub-goals are not tangents; no detour cue expected, got %q", lp)
	}
}

// TestAdoptDedupsExistingGoal is the H10 (consumer) regression: re-adopting a goal
// whose label already exists as a node (any kind) must NOT create a twin open_goal,
// keeping the graph bounded.
func TestAdoptDedupsExistingGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")
	// A done goal already in the graph.
	h.goalGraph.Upsert("learn fishing", goalgraph.KindGoal, "learn fishing", goalgraph.StatusDone)

	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "learn fishing"})
	// No open_goal twin (it already exists as a done goal node).
	if og := h.goalGraph.OpenGoals(); len(og) != 0 {
		t.Fatalf("re-adopting an existing label must not create an open_goal twin; got %v", og)
	}
	// A brand-new label still queues.
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "mine copper"})
	if og := h.goalGraph.OpenGoals(); len(og) != 1 || !strings.EqualFold(og[0].ID, "mine copper") {
		t.Fatalf("a new label should still queue; got %v", og)
	}
}

// TestEffectiveGoalViewRaceFree is the H8-regression -race guard: the speech tier
// reads the director's effective goal via effectiveGoalView (RLock) on one
// goroutine while the director ADVANCES d.goal via effectiveGoal (write Lock) on
// another. Before goalMu these two paths touched d.goal unguarded and the run
// failed under `go test -race`. The reader must always observe a coherent
// (non-torn) goal string and the run must be race-clean.
func TestEffectiveGoalViewRaceFree(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	// A chain of open goals so the mutating path keeps advancing d.goal (each
	// advance is a write the reader races against). g0 starts active+done so the
	// first resolve advances it.
	const chain = 200
	d.goalMu.Lock()
	d.goal = "g0"
	d.goalMu.Unlock()
	h.goalGraph.Upsert("g0", goalgraph.KindGoal, "g0", goalgraph.StatusDone)
	for i := 1; i <= chain; i++ {
		id := fmt.Sprintf("g%d", i)
		h.goalGraph.Upsert(id, goalgraph.KindOpenGoal, id, goalgraph.StatusOpen)
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer: advance through the chain. effectiveGoal returns the just-promoted
	// goal; mark it Done so the next call advances again — keeping d.goal churning
	// for the full run so the reader genuinely overlaps the writes.
	go func() {
		defer wg.Done()
		for i := 0; i < chain; i++ {
			g := d.effectiveGoal(h) // resolveGoal(mutate=true): may write d.goal
			if g == "" {
				return // ran out of successors — nothing left to advance
			}
			h.goalGraph.SetStatus(g, goalgraph.StatusDone)
		}
	}()

	// Reader: hammer the pure read the speech tier uses. Every value must be a
	// coherent goal string (never a torn read); we only assert it doesn't panic /
	// race and stays in the known id space when non-empty.
	go func() {
		defer wg.Done()
		for i := 0; i < 5000; i++ {
			g := d.effectiveGoalView(h)
			if g != "" && g[0] != 'g' {
				t.Errorf("effectiveGoalView returned an incoherent goal %q", g)
				return
			}
		}
	}()

	wg.Wait()
}

// TestTrustNoteRendersMultiAxis proves the compact planner note surfaces the
// multi-axis read (warm/cold/grudge) ONLY past the thresholds, plus grade,
// familiarity, and tags — the substrate the planner reads for nearby players.
func TestTrustNoteRendersMultiAxis(t *testing.T) {
	h := newTestHost()
	d := quietDirector()

	// Neutral acquaintance: grade + familiarity, no axis colour.
	h.ledger.Met("Mild")
	note := d.trustNote(h, "Mild")
	if !strings.Contains(note, "met 1") {
		t.Fatalf("expected familiarity in note, got %q", note)
	}
	if strings.Contains(note, "warm") || strings.Contains(note, "cold") || strings.Contains(note, "grudge") {
		t.Fatalf("a neutral relationship should carry no axis colour, got %q", note)
	}

	// Warm sparring partner.
	h.ledger.Met("Friend")
	h.ledger.ObserveAffinity("Friend", 3) // squashes >= 0.3
	h.ledger.Tag("Friend", "sparring-partner")
	warm := d.trustNote(h, "Friend")
	if !strings.Contains(warm, "warm") || !strings.Contains(warm, "sparring-partner") {
		t.Fatalf("a warm sparring partner should render warm + tag, got %q", warm)
	}

	// A party we hold a grudge against.
	h.ledger.Met("Foe")
	h.ledger.ObserveGrievance("Foe", 4) // squashes >= 0.3
	h.ledger.Tag("Foe", "ganked-me")
	foe := d.trustNote(h, "Foe")
	if !strings.Contains(foe, "you hold a grudge") || !strings.Contains(foe, "ganked-me") {
		t.Fatalf("a grudge should render the grudge note + tag, got %q", foe)
	}
}
