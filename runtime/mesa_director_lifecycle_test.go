package runtime

import (
	"io"
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

// --- Phase-5a goal lifecycle: prove the infinite loop is broken --------------

// seededDirector returns a director whose active goal is already a live root
// node, the normal post-genesis state (the seed runs on the first situation()).
func seededDirector(h *Host, goal string) *MesaDirector {
	d := NewMesaDirector(mesaclient.StubClient{}, "tester", goal, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if goal != "" {
		h.goalGraph.Upsert(goal, goalgraph.KindGoal, goal, goalgraph.StatusActive)
	}
	return d
}

// TestLifecycleCompletionDeclared proves COMPLETION write-back (Deliverable 1):
// a Move with goal_op:"done" flips the active goal node to StatusDone with
// progress 1 — the planner's "the goal is met" judgment becomes a DURABLE graph
// write (the exact gap that caused the live loop).
func TestLifecycleCompletionDeclared(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "done"})

	n, ok := h.goalGraph.Get("do a thing")
	if !ok || n.Status != goalgraph.StatusDone {
		t.Fatalf("goal_op done should flip the goal to StatusDone; got %+v", n)
	}
	if n.Progress != 1 {
		t.Fatalf("a done goal must read progress 1; got %v", n.Progress)
	}
}

// TestLifecycleAbandonDeclared proves goal_op:"abandoned" sets StatusAbandoned
// (so effectiveGoal advances off it) without claiming completion.
func TestLifecycleAbandonDeclared(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "abandoned"})

	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusAbandoned {
		t.Fatalf("goal_op abandoned should flip to StatusAbandoned; got %q", n.Status)
	}
}

// TestLifecycleAdoptQueuesOpenGoal proves goal_op:"adopt" QUEUES the next
// objective as an open_goal candidate WITHOUT switching the active goal — that
// switch is selectNextGoal's policy, owned by advancement.
func TestLifecycleAdoptQueuesOpenGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "learn fishing"})

	og := h.goalGraph.OpenGoals()
	if len(og) != 1 || !strings.EqualFold(og[0].ID, "learn fishing") {
		t.Fatalf("adopt should queue one open_goal; got %v", og)
	}
	// The active goal is unchanged (adopt never switches it).
	if g := d.effectiveGoal(h); g != "do a thing" {
		t.Fatalf("adopt must not switch the active goal; effectiveGoal=%q", g)
	}
}

// TestLifecycleProgressMonotone proves PROGRESS tracking (Deliverable 2) with
// the monotone-up guard: a higher report advances the node; a later lower report
// is ignored (never regress).
func TestLifecycleProgressMonotone(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	d.applyGoalOp(h, &mesaclient.Move{GoalProgress: 0.6})
	if n, _ := h.goalGraph.Get("do a thing"); n.Progress != 0.6 {
		t.Fatalf("first progress report should set 0.6; got %v", n.Progress)
	}
	d.applyGoalOp(h, &mesaclient.Move{GoalProgress: 0.3}) // lower → ignored
	if n, _ := h.goalGraph.Get("do a thing"); n.Progress != 0.6 {
		t.Fatalf("progress must be monotone up; a 0.3 report should not lower 0.6; got %v", n.Progress)
	}
}

// TestLifecycleProgressNeverRevivesDoneGoal proves a progress report on a
// done/abandoned node is a no-op (the lifecycle is terminal once declared done).
func TestLifecycleProgressNeverRevivesDoneGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")
	h.goalGraph.SetStatus("do a thing", goalgraph.StatusDone)
	h.goalGraph.SetProgress("do a thing", 1)

	d.applyGoalOp(h, &mesaclient.Move{GoalProgress: 0.5})
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusDone || n.Progress != 1 {
		t.Fatalf("a progress report must not revive a done goal; got %+v", n)
	}
}

// pickaxeFacts is a tiny facts registry naming a bronze pickaxe so itemName
// resolves to the human name the goal text references.
func pickaxeFacts() *facts.Facts {
	return &facts.Facts{
		ItemDefs: map[int]*facts.ItemDef{
			156: {ID: 156, Name: "Bronze Pickaxe", Command: "Wield"},
		},
		NpcDefs:      map[int]*facts.NpcDef{},
		SceneryDefs:  map[int]*facts.SceneryDef{},
		BoundaryDefs: map[int]*facts.BoundaryDef{},
	}
}

// TestLifecycleInventorySatisfactionNet is the regression test for task #21 (the
// literal pickaxe bug): an acquire-item goal whose item is already in inventory
// auto-completes on a successful turn with NO planner declaration. The host READS
// the inventory and WRITES StatusDone deterministically.
func TestLifecycleInventorySatisfactionNet(t *testing.T) {
	h := New(Options{Username: "test", Facts: pickaxeFacts()})
	h.world.Inventory.Replace([]world.InvSlot{{ItemID: 156, Amount: 1}}) // Bronze Pickaxe
	goal := "acquire a bronze pickaxe"
	d := NewMesaDirector(mesaclient.StubClient{}, "tester", goal, slog.New(slog.NewTextHandler(io.Discard, nil)))

	// A successful turn (a real prior action that completed).
	ok := Outcome{Intent: Intent{Label: "act:buy", Name: "buy"}, Kind: interp.ResultCompleted}
	d.situation(h, ok)

	n, _ := h.goalGraph.Get(goal)
	if n.Status != goalgraph.StatusDone {
		t.Fatalf("an acquire-item goal whose item is in inventory should auto-complete; status=%q", n.Status)
	}
	if n.Progress != 1 {
		t.Fatalf("auto-completed goal must read progress 1; got %v", n.Progress)
	}
}

// TestLifecycleInventoryNetIgnoresNonAcquireGoals proves the net is conservative:
// a non-acquire goal is NOT auto-completed even when the named item is present
// (the planner's declaration owns those).
func TestLifecycleInventoryNetIgnoresNonAcquireGoals(t *testing.T) {
	h := New(Options{Username: "test", Facts: pickaxeFacts()})
	h.world.Inventory.Replace([]world.InvSlot{{ItemID: 156, Amount: 1}})
	goal := "mine ore with a bronze pickaxe" // names the item but is not an acquire goal
	d := NewMesaDirector(mesaclient.StubClient{}, "tester", goal, slog.New(slog.NewTextHandler(io.Discard, nil)))

	d.situation(h, Outcome{Intent: Intent{Label: "act:mine", Name: "mine"}, Kind: interp.ResultCompleted})
	if n, _ := h.goalGraph.Get(goal); n.Status == goalgraph.StatusDone {
		t.Fatalf("a non-acquire goal must not auto-complete from inventory; status=%q", n.Status)
	}
}

// TestLifecycleAdvancement proves ADVANCEMENT (Deliverable 3): once the active
// goal is Done, effectiveGoal selects the newest open_goal (promoting it to
// Active); with no open_goal it falls back to the north-star; with neither it
// stays put on the done goal and does NOT loop a new one.
func TestLifecycleAdvancement(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// Queue a successor and mark the active goal done.
	h.goalGraph.Upsert("learn fishing", goalgraph.KindOpenGoal, "learn fishing", goalgraph.StatusOpen)
	d.markGoalDone(h, "do a thing")

	if g := d.effectiveGoal(h); !strings.EqualFold(g, "learn fishing") {
		t.Fatalf("a done goal should advance to the queued open_goal; got %q", g)
	}
	if n, _ := h.goalGraph.Get("learn fishing"); n.Status != goalgraph.StatusActive {
		t.Fatalf("the promoted successor should be StatusActive; got %q", n.Status)
	}
	// Advance is sticky: subsequent calls return the SAME new goal (idempotent).
	if g := d.effectiveGoal(h); !strings.EqualFold(g, "learn fishing") {
		t.Fatalf("advancement must be sticky; got %q", g)
	}
}

// TestLifecycleAdvancementNorthStar proves the persona-northstar fallback when no
// graph open_goal successor is queued.
func TestLifecycleAdvancementNorthStar(t *testing.T) {
	h := newTestHost()
	h.northStar = "become a master smith"
	d := seededDirector(h, "do a thing")
	d.markGoalDone(h, "do a thing")

	if g := d.effectiveGoal(h); g != "become a master smith" {
		t.Fatalf("with no open_goal, a done goal should advance to the north-star; got %q", g)
	}
}

// TestLifecycleAdvancementStaysPut proves the no-successor case: a done goal with
// no open_goal and no north-star stays on the done goal (idle purposefully) and
// never loops a fresh one — the host does not spin.
func TestLifecycleAdvancementStaysPut(t *testing.T) {
	h := newTestHost() // no northStar
	d := seededDirector(h, "do a thing")
	d.markGoalDone(h, "do a thing")

	if g := d.effectiveGoal(h); g != "do a thing" {
		t.Fatalf("with no successor a done goal should stay put; got %q", g)
	}
}

// TestLifecycleOperatorOverrideWins proves the INVARIANT: with a LiveGoal set,
// effectiveGoal returns the operator goal even when the construction goal is
// Done — advancement is skipped (operator wins, first statement).
func TestLifecycleOperatorOverrideWins(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")
	h.goalGraph.Upsert("learn fishing", goalgraph.KindOpenGoal, "learn fishing", goalgraph.StatusOpen)
	d.markGoalDone(h, "do a thing")
	h.SetLiveGoal("OPERATOR: report to spawn")

	if g := d.effectiveGoal(h); g != "OPERATOR: report to spawn" {
		t.Fatalf("the operator LiveGoal must win over advancement; got %q", g)
	}
	// And the construction goal d.goal must not have advanced under the override.
	if g := d.goal; g != "do a thing" {
		t.Fatalf("advancement must be skipped while the operator override is set; d.goal=%q", g)
	}
}

// TestSpinDetectorTrips proves the SPIN DETECTOR (Deliverable 4): the same
// EXECUTED plan on successive OK turns reaches the spin threshold, the SPINNING
// trigger is emitted, and the goal goes StatusBlocked + tagged "spinning" with an
// open question — the gap the fail-streak-only anti-stuck misses (a SUCCEEDING
// loop).
func TestSpinDetectorTrips(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// The identical write_routine, succeeding every turn.
	same := Outcome{
		Intent: Intent{Label: "act:check_shop", Name: "check_shop", Source: "runtime \"1.0\"\nroutine check_shop() { survey() }"},
		Kind:   interp.ResultCompleted,
	}
	var sit *mesaclient.Situation
	for i := range antiStuckSpinTurns + 1 { // first turn sets the baseline FP, then count up to threshold
		// Move >1 tile each turn so the position-based STUCK detector (a
		// higher-priority trigger) does not pre-empt the spin branch we test here.
		h.world.Self.SetPosition(world.Coord{X: 120 + i*5, Y: 504})
		sit = d.situation(h, same)
	}
	if !strings.Contains(sit.Trigger, "SPINNING") {
		t.Fatalf("the same plan %d turns running should emit SPINNING; got %q", antiStuckSpinTurns+1, sit.Trigger)
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusBlocked {
		t.Fatalf("spinning should block the goal (same path as fail-streak); status=%q", n.Status)
	}
	if n, _ := h.goalGraph.Get("do a thing"); !slices.Contains(n.Tags, "spinning") {
		t.Fatalf("a spinning goal should be tagged spinning; tags=%v", n.Tags)
	}
	if got := len(h.goalGraph.OpenQuestions()); got != 1 {
		t.Fatalf("spinning should spawn one open question; got %d", got)
	}
}

// TestSpinDetectorDoesNotFalseTrip proves the detector does NOT fire on a
// CHANGING plan (each turn a different routine) — only genuine re-derivation
// counts.
func TestSpinDetectorDoesNotFalseTrip(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	for i := range antiStuckSpinTurns + 2 {
		src := "runtime \"1.0\"\nroutine r() { survey(" + strings.Repeat("x", i+1) + ") }"
		out := Outcome{Intent: Intent{Label: "act:r", Name: "r", Source: src}, Kind: interp.ResultCompleted}
		sit := d.situation(h, out)
		if strings.Contains(sit.Trigger, "SPINNING") {
			t.Fatalf("a changing plan must not trip the spin detector; tripped on turn %d", i)
		}
	}
	if d.spinCount != 0 {
		t.Fatalf("a changing plan should keep spinCount at 0; got %d", d.spinCount)
	}
}

// TestSpinDetectorFailingTurnsDoNotCount proves a FAILING repeated plan is the
// fail-streak path's job, not the spin path's — spinCount stays at 0 on failures.
func TestSpinDetectorFailingTurnsDoNotCount(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	fail := Outcome{Intent: Intent{Label: "act:r", Name: "r", Source: "runtime \"1.0\"\nroutine r() { survey() }"}, Kind: interp.ResultErrored}
	for range antiStuckSpinTurns + 2 {
		d.situation(h, fail)
	}
	if d.spinCount != 0 {
		t.Fatalf("failing turns must not accrue spinCount; got %d", d.spinCount)
	}
}

// TestLifecycleCloseThenAdvance is the END-TO-END proof that the infinite loop is
// broken (Step 8): the FULL handoff in ONE chain — a planner-declared completion
// flips the active goal to StatusDone (Deliverable 1), and the VERY NEXT
// effectiveGoal call advances to a DIFFERENT goal (Deliverable 3), not the same
// one. Today completion and advancement were tested separately; this binds them:
// the live bug was that a satisfied goal closed in the planner's head but the host
// re-derived the SAME goal forever. Here the close (goal_op:"done") leads
// deterministically to a new active goal — the host moves on.
func TestLifecycleCloseThenAdvance(t *testing.T) {
	h := newTestHost()
	h.northStar = "become a master smith"
	d := seededDirector(h, "acquire a bronze pickaxe")

	// Before completion, effectiveGoal is the (single, static) genesis goal —
	// exactly the state that looped forever.
	if g := d.effectiveGoal(h); g != "acquire a bronze pickaxe" {
		t.Fatalf("pre-completion effectiveGoal should be the genesis goal; got %q", g)
	}

	// The planner DECLARES the goal met (the judgment that used to evaporate).
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "done"})

	// Durable graph write: the goal is closed.
	if n, _ := h.goalGraph.Get("acquire a bronze pickaxe"); n.Status != goalgraph.StatusDone || n.Progress != 1 {
		t.Fatalf("declared completion must flip the goal to Done/progress 1; got %+v", n)
	}

	// THE LOOP BREAKS HERE: the next effectiveGoal returns a DIFFERENT goal (the
	// north-star, no open_goal queued), and d.goal has advanced off the closed one.
	next := d.effectiveGoal(h)
	if next == "acquire a bronze pickaxe" {
		t.Fatalf("a closed goal must NOT be re-selected (the loop); effectiveGoal still %q", next)
	}
	if next != "become a master smith" {
		t.Fatalf("with no open_goal, advancement should pick the north-star; got %q", next)
	}
	if d.goal != "become a master smith" {
		t.Fatalf("advancement must stick: d.goal should be the successor; got %q", d.goal)
	}
}

// TestLifecycleCloseThenAdvanceToOpenGoal is the same close→advance chain but with
// a queued open_goal taking priority over the north-star — proving the adopt →
// selectNextGoal queue path end-to-end (a planner-grown successor is consumed).
func TestLifecycleCloseThenAdvanceToOpenGoal(t *testing.T) {
	h := newTestHost()
	h.northStar = "become a master smith"
	d := seededDirector(h, "acquire a bronze pickaxe")

	// The planner queues a successor on an earlier turn, then declares done.
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "mine some copper"})
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "done"})

	if next := d.effectiveGoal(h); next != "mine some copper" {
		t.Fatalf("a queued open_goal should win over the north-star on advancement; got %q", next)
	}
	// The promoted successor is now the active goal node.
	if n, _ := h.goalGraph.Get("mine some copper"); n.Status != goalgraph.StatusActive {
		t.Fatalf("the promoted successor should be StatusActive; got %q", n.Status)
	}
}

// TestLifecycleAdoptSkipsActiveGoal proves the G1 dedup guard (Step 10): adopting
// the goal that is ALREADY the active goal does NOT create a self-referential
// open_goal twin. selectNextGoal already refuses to re-select it, so this is a
// cleanliness guard, not a loop fix — but a goal should never queue itself.
func TestLifecycleAdoptSkipsActiveGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// Same text as the active goal (case-insensitive) — must be skipped.
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "Do A Thing"})

	if og := h.goalGraph.OpenGoals(); len(og) != 0 {
		t.Fatalf("adopting the active goal must NOT queue an open_goal twin; got %v", og)
	}
	// A DIFFERENT goal still queues normally (guard is narrow).
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "learn fishing"})
	if og := h.goalGraph.OpenGoals(); len(og) != 1 || !strings.EqualFold(og[0].ID, "learn fishing") {
		t.Fatalf("a non-active goal should still queue; got %v", og)
	}
}
