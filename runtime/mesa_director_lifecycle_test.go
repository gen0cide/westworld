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

// TestLifecycleAdvancementStaysPut proves the no-successor case (L5): a done goal
// with no open_goal and no north-star resolves to "" — NOT the done goal. Surfacing
// "" drops the GOAL line in act.go (the planner is not handed a finished goal as if
// it were active) and stops the satisfaction net from re-firing on it every turn.
// The host idles purposefully; it does not loop a fresh goal.
func TestLifecycleAdvancementStaysPut(t *testing.T) {
	h := newTestHost() // no northStar
	d := seededDirector(h, "do a thing")
	d.markGoalDone(h, "do a thing")

	if g := d.effectiveGoal(h); g != "" {
		t.Fatalf("a done goal with no successor should resolve to \"\" (no active objective); got %q", g)
	}
	// d.goal is unchanged (no successor to advance TO) — but effectiveGoal/View keep
	// reporting "" so the done goal is never re-handed to the planner.
	if v := d.effectiveGoalView(h); v != "" {
		t.Fatalf("the pure view must agree: \"\" for a done goal with no successor; got %q", v)
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

// okSame is an OK outcome re-running the identical plan (drives the spin detector).
func okSame() Outcome {
	return Outcome{
		Intent: Intent{Label: "act:check_shop", Name: "check_shop", Source: "runtime \"1.0\"\nroutine check_shop() { survey() }"},
		Kind:   interp.ResultCompleted,
	}
}

// stepTo moves the host >1 tile so the position-based STUCK detector does not
// pre-empt the lower-priority spin branch under test, then runs one situation().
func stepTo(d *MesaDirector, h *Host, turn int, last Outcome) *mesaclient.Situation {
	h.world.Self.SetPosition(world.Coord{X: 120 + turn*5, Y: 504})
	return d.situation(h, last)
}

// TestSpinDoesNotResurrectDoneGoal is the H1 regression: a planner-COMPLETED goal
// with no successor must NEVER be flipped Blocked by the spin branch (which would
// let the OK-turn un-block recovery flip it Active again and loop forever). After
// markGoalDone, even a sustained identical-plan spin leaves the goal terminal and
// emits no SPINNING trigger.
func TestSpinDoesNotResurrectDoneGoal(t *testing.T) {
	h := newTestHost() // no northStar, no open_goal → no successor
	d := seededDirector(h, "do a thing")
	d.markGoalDone(h, "do a thing")

	var sit *mesaclient.Situation
	for i := range antiStuckSpinTurns + 3 {
		sit = stepTo(d, h, i, okSame())
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusDone {
		t.Fatalf("a done goal must stay Done through a spin; got %q", n.Status)
	}
	if slices.Contains(mustGet(t, h, "do a thing").Tags, "spinning") {
		t.Fatalf("a done goal must not be tagged spinning")
	}
	if strings.Contains(sit.Trigger, "SPINNING") {
		t.Fatalf("SPINNING must not fire on a terminal goal; got %q", sit.Trigger)
	}
}

func mustGet(t *testing.T, h *Host, id string) goalgraph.Node {
	t.Helper()
	n, ok := h.goalGraph.Get(id)
	if !ok {
		t.Fatalf("node %q missing", id)
	}
	return n
}

// TestDisplacedTurnDoesNotMisMarkGoal is the H2 regression: a turn that is BOTH
// spinning AND displaced must surface the displacement re-orient trigger WITHOUT
// the anti-stuck graph writes (no Blocked status, no "spinning" tag, no open
// question) — those would durably mis-mark the goal for a spin about the OLD
// location the jump invalidated.
func TestDisplacedTurnDoesNotMisMarkGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// Build up to the spin threshold on a moving host (no STUCK pre-empt).
	for i := range antiStuckSpinTurns {
		stepTo(d, h, i, okSame())
	}
	// The next turn would SPIN — but the host is displaced (lured) this turn.
	h.displacement.record(displacementEvent{fromX: 120, fromY: 504, toX: 200, toY: 600, dist: 99, cause: dispJump})
	sit := stepTo(d, h, antiStuckSpinTurns, okSame())

	if !strings.Contains(sit.Trigger, "DISPLACED") {
		t.Fatalf("a displaced turn should surface the displacement re-orient trigger; got %q", sit.Trigger)
	}
	if strings.Contains(sit.Trigger, "SPINNING") {
		t.Fatalf("the displacement trigger must win over SPINNING; got %q", sit.Trigger)
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status == goalgraph.StatusBlocked {
		t.Fatalf("a displaced turn must not mark the goal Blocked")
	}
	if slices.Contains(mustGet(t, h, "do a thing").Tags, "spinning") {
		t.Fatalf("a displaced turn must not tag the goal spinning")
	}
	if got := len(h.goalGraph.OpenQuestions()); got != 0 {
		t.Fatalf("a displaced turn must not spawn an open question; got %d", got)
	}
}

// TestDisplacementResetsSpinState is the H3 regression: a death/teleport is a hard
// context change that resets spinCount AND lastPlanFP (not just failStreak), so the
// spin counter cannot reach threshold on executions begun in a different location.
func TestDisplacementResetsSpinState(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// One spin-accruing turn, then a death the next turn.
	stepTo(d, h, 0, okSame())
	stepTo(d, h, 1, okSame())
	if d.spinCount == 0 {
		t.Fatalf("precondition: spinCount should have accrued; got %d", d.spinCount)
	}
	h.displacement.record(displacementEvent{toX: 120, toY: 648, dist: 99, cause: dispDeath})
	stepTo(d, h, 2, okSame())
	if d.spinCount != 0 || d.lastPlanFP != 0 {
		t.Fatalf("displacement should reset spin state; spinCount=%d lastPlanFP=%d", d.spinCount, d.lastPlanFP)
	}
}

// TestSpinDetectorIgnoresIdleWaits is the H4 regression: a host that legitimately
// IDLES (the same wait every turn, e.g. the planner-error / empty-verb fallbacks)
// must NOT accrue spinCount or trip SPINNING — an idle is the absence of a plan.
func TestSpinDetectorIgnoresIdleWaits(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	idleOut := Outcome{Intent: d.idle(3), Kind: interp.ResultCompleted}
	var sit *mesaclient.Situation
	for i := range antiStuckSpinTurns + 3 {
		sit = stepTo(d, h, i, idleOut)
	}
	if d.spinCount != 0 {
		t.Fatalf("idle waits must not accrue spinCount; got %d", d.spinCount)
	}
	if strings.Contains(sit.Trigger, "SPINNING") {
		t.Fatalf("idling must not trip SPINNING; got %q", sit.Trigger)
	}
	if n, _ := h.goalGraph.Get("do a thing"); n.Status == goalgraph.StatusBlocked {
		t.Fatalf("idling must not block the goal")
	}
}

// TestSpinTagClearedOnRecovery is the H5 regression: the transient "spinning" tag
// is cleared when the goal recovers (an OK turn un-blocks it), so renderIntention
// stops appending the stale "you've been SPINNING" prose for the host's whole life.
func TestSpinTagClearedOnRecovery(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// Trip the spin detector → goal Blocked + tagged spinning.
	for i := range antiStuckSpinTurns + 1 {
		stepTo(d, h, i, okSame())
	}
	if !slices.Contains(mustGet(t, h, "do a thing").Tags, "spinning") {
		t.Fatalf("precondition: goal should be tagged spinning")
	}
	// A real DIFFERENT-plan success recovers: un-block + untag.
	diff := Outcome{Intent: Intent{Label: "act:mine", Name: "mine", Source: "runtime \"1.0\"\nroutine mine() { survey(\"x\") }"}, Kind: interp.ResultCompleted}
	stepTo(d, h, antiStuckSpinTurns+2, diff)
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusActive {
		t.Fatalf("a success should un-block the goal; status=%q", n.Status)
	}
	if slices.Contains(mustGet(t, h, "do a thing").Tags, "spinning") {
		t.Fatalf("recovery must clear the spinning tag; tags=%v", mustGet(t, h, "do a thing").Tags)
	}
}

// TestMarkGoalBlockedDoesNotReopenResolvedQuestion is the M8 regression: once the
// reactive/cron closers resolve a how-to-progress question (StatusDone), a later
// hard-fail/spin must NOT force it back to StatusOpen.
func TestMarkGoalBlockedDoesNotReopenResolvedQuestion(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")

	// First block creates the open question.
	d.markGoalBlocked(h, "do a thing", "stuck the first time")
	qid := "how-to-progress:do a thing"
	if n, _ := h.goalGraph.Get(qid); n.Status != goalgraph.StatusOpen {
		t.Fatalf("first block should open the question; got %q", n.Status)
	}
	// A closer resolves it.
	h.goalGraph.SetStatus(qid, goalgraph.StatusDone)
	// Re-block must NOT re-open the resolved question.
	d.markGoalBlocked(h, "do a thing", "stuck again")
	if n, _ := h.goalGraph.Get(qid); n.Status != goalgraph.StatusDone {
		t.Fatalf("a re-block must not re-open a resolved question; got %q", n.Status)
	}
}

// TestMarkGoalBlockedSkipsTerminalGoal is the H1 unit-level guard: markGoalBlocked
// is a no-op on a done/abandoned goal (it must never resurrect a terminal goal).
func TestMarkGoalBlockedSkipsTerminalGoal(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "do a thing")
	d.markGoalDone(h, "do a thing")

	d.markGoalBlocked(h, "do a thing", "should be ignored")
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusDone {
		t.Fatalf("markGoalBlocked must not re-block a done goal; got %q", n.Status)
	}
	if got := len(h.goalGraph.OpenQuestions()); got != 0 {
		t.Fatalf("markGoalBlocked on a done goal must write no open question; got %d", got)
	}
}

// TestAnalysisModeDoesNotMutateGoalGraph is the M16 regression: a dry-run
// situation() under freeze must NOT seed a goal node, flip status, or write the
// journal objective — the in-RAM state the host resumes with stays untouched.
func TestAnalysisModeDoesNotMutateGoalGraph(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "") // no seeded node
	d.goal = "freeze probe goal"
	h.EnterAnalysis()

	// A would-be seeding turn under freeze writes nothing.
	d.situation(h, Outcome{Intent: Intent{Label: "act:x", Name: "x"}, Kind: interp.ResultCompleted})
	if h.goalGraph.Has("freeze probe goal") {
		t.Fatalf("analysis mode must not seed a goal node")
	}
	if got := len(h.goalGraph.Nodes()); got != 0 {
		t.Fatalf("analysis mode must not write goal-graph nodes; got %d", got)
	}

	// A would-be advancement under freeze must not flip an open_goal active.
	h.goalGraph.Upsert("done one", goalgraph.KindGoal, "done one", goalgraph.StatusDone)
	h.goalGraph.Upsert("successor", goalgraph.KindOpenGoal, "successor", goalgraph.StatusOpen)
	d.goal = "done one"
	if g := d.effectiveGoal(h); g != "successor" {
		t.Fatalf("the view should still RESOLVE the successor; got %q", g)
	}
	if n, _ := h.goalGraph.Get("successor"); n.Status != goalgraph.StatusOpen {
		t.Fatalf("analysis mode must not promote the successor to Active; got %q", n.Status)
	}
	if d.goal != "done one" {
		t.Fatalf("analysis mode must not advance d.goal; got %q", d.goal)
	}
}

// TestGoalOpSynonymsNormalize is the M7 regression: near-synonyms for the
// lifecycle ops fold to the canonical op, and an unknown value is treated as a
// progress report (the goal is not closed/abandoned by a typo).
func TestGoalOpSynonymsNormalize(t *testing.T) {
	cases := []struct {
		raw string
		op  string
		ok  bool
	}{
		{"done", "done", true},
		{"completed", "done", true},
		{"Finished", "done", true},
		{"satisfied", "done", true},
		{"abandon", "abandoned", true},
		{"give up", "abandoned", true},
		{"adopt", "adopt", true},
		{"new goal", "adopt", true},
		{"", "", true},
		{"banana", "", false}, // unknown → progress report, recognized=false
	}
	for _, c := range cases {
		op, ok := normalizeGoalOp(c.raw)
		if op != c.op || ok != c.ok {
			t.Fatalf("normalizeGoalOp(%q) = (%q,%v), want (%q,%v)", c.raw, op, ok, c.op, c.ok)
		}
	}

	// End-to-end: a "completed" declaration closes the goal (it used to be ignored).
	h := newTestHost()
	d := seededDirector(h, "do a thing")
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "completed"})
	if n, _ := h.goalGraph.Get("do a thing"); n.Status != goalgraph.StatusDone {
		t.Fatalf("goal_op \"completed\" should close the goal (M7); got %q", n.Status)
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
