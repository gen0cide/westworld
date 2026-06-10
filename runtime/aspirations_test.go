package runtime

import (
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
)

// --- aspiration ladder: seeding, migration, linking, balance, rollup ---------

func aspirationDirector(goal string) *MesaDirector {
	return NewMesaDirector(mesaclient.StubClient{}, "tester", goal,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// TestEnsureAspirationsGenesisSeeds proves the genesis path: the pending
// labels become KindAspiration roots (validated: blank dropped, capped at
// maxAspirations), the opening day-scale goal becomes a real KindGoal root
// linked serves→ the declared aspiration, and a second/third ensure changes
// NOTHING (idempotent across turns).
func TestEnsureAspirationsGenesisSeeds(t *testing.T) {
	h := newTestHost()
	d := aspirationDirector("smelt your first bronze bar")
	d.SetAspirations([]string{
		"master a craft (smithing)", "  ", "secure a steady livelihood",
		"come to know the whole map", "be famous", "one too many",
	}, "master a craft (smithing)")

	d.ensureAspirations(h)

	asps := h.goalGraph.Aspirations()
	if len(asps) != maxAspirations {
		t.Fatalf("aspirations seeded = %d, want %d (blank dropped, capped)", len(asps), maxAspirations)
	}
	// The opening goal is a real goal root and serves the DECLARED aspiration.
	gn, ok := h.goalGraph.Get("smelt your first bronze bar")
	if !ok || gn.Kind != goalgraph.KindGoal || gn.Status != goalgraph.StatusActive {
		t.Fatalf("opening goal not a live KindGoal root: %+v ok=%v", gn, ok)
	}
	serves := h.goalGraph.Out("smelt your first bronze bar", goalgraph.RelServes)
	if len(serves) != 1 || serves[0].To != aspirationID("master a craft (smithing)") {
		t.Fatalf("opening serves-edge = %+v, want the declared aspiration", serves)
	}

	// Exactly once: re-ensuring (the confirming turn + one more) is a no-op.
	d.ensureAspirations(h)
	if !d.aspirationsReady {
		t.Fatal("second ensure should latch aspirationsReady (portfolio confirmed)")
	}
	d.ensureAspirations(h)
	if got := len(h.goalGraph.Aspirations()); got != maxAspirations {
		t.Fatalf("re-ensure duplicated aspirations: %d", got)
	}
	if got := len(h.goalGraph.Out("smelt your first bronze bar", goalgraph.RelServes)); got != 1 {
		t.Fatalf("re-ensure duplicated serves-edges: %d", got)
	}
}

// TestEnsureAspirationsLazyMigration proves the no-LLM migration: a host with a
// persona snapshot (north star + curiosity) but NO graph aspirations — the live
// fleet's state — derives 2-3 mechanically at ensure time; the derivation is
// deterministic, so a second life (new director over the same persisted graph)
// converges on the SAME nodes instead of duplicating.
func TestEnsureAspirationsLazyMigration(t *testing.T) {
	h := newTestHost()
	h.northStar = "A name posted on every guild door."
	h.curiosity = persona.Curiosity{Skill: 0.8, Economic: 0.6, Spatial: 0.2}
	// Simulate the pre-ladder fleet: a persisted graph with goals but no aspirations.
	h.goalGraph.Upsert("mine some ore", goalgraph.KindGoal, "mine some ore", goalgraph.StatusActive)

	d := aspirationDirector("mine some ore")
	d.ensureAspirations(h)

	asps := h.goalGraph.Aspirations()
	if len(asps) != 3 {
		t.Fatalf("derived aspirations = %d, want 3 (north star + top-2 flavors)", len(asps))
	}
	if asps[0].Label != h.northStar {
		t.Fatalf("first derived aspiration should be the north star, got %q", asps[0].Label)
	}
	if !strings.Contains(strings.ToLower(asps[1].Label), "craft") {
		t.Fatalf("strongest flavor (skill) should derive the craft aspiration, got %q", asps[1].Label)
	}
	// The active goal joined the portfolio ("mine some ore" keyword-matches the
	// skill-family craft aspiration, beating the no-token-overlap north star).
	serves := h.goalGraph.Out("mine some ore", goalgraph.RelServes)
	if len(serves) != 1 || serves[0].To != asps[1].ID {
		t.Fatalf("active goal serves = %+v, want the craft aspiration %q", serves, asps[1].ID)
	}

	// Exactly once across a bounce: a NEW director (fresh ready-latch) over the
	// same graph derives the same ids — no growth.
	d2 := aspirationDirector("mine some ore")
	d2.ensureAspirations(h)
	d2.ensureAspirations(h)
	if got := len(h.goalGraph.Aspirations()); got != 3 {
		t.Fatalf("re-migration after a bounce duplicated aspirations: %d", got)
	}
}

// TestEnsureAspirationsStaysDarkWithoutPersona proves the graceful fallback: a
// personaless host (zero curiosity, no north star, no genesis labels) seeds
// NOTHING — behavior is exactly pre-ladder.
func TestEnsureAspirationsStaysDarkWithoutPersona(t *testing.T) {
	h := newTestHost()
	d := aspirationDirector("do a thing")
	d.ensureAspirations(h)
	if got := len(h.goalGraph.Aspirations()); got != 0 {
		t.Fatalf("personaless host invented %d aspirations", got)
	}
	if !d.aspirationsReady {
		t.Fatal("underivable portfolio should latch ready (stay dark, stop probing)")
	}
}

// TestAdoptLinksServesDeclared proves a planner adoption carries its aspiration:
// goal_op:"adopt" + goal_serves names a live aspiration → the open_goal gets the
// serves-edge to exactly that aspiration.
func TestAdoptLinksServesDeclared(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "mine some ore")
	h.goalGraph.Upsert(aspirationID("secure a steady livelihood"), goalgraph.KindAspiration,
		"secure a steady livelihood", goalgraph.StatusActive)
	h.goalGraph.Upsert(aspirationID("master a craft"), goalgraph.KindAspiration,
		"master a craft", goalgraph.StatusActive)

	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "sell ore at the general store",
		GoalServes: "secure a steady livelihood"})

	serves := h.goalGraph.Out("sell ore at the general store", goalgraph.RelServes)
	if len(serves) != 1 || serves[0].To != aspirationID("secure a steady livelihood") {
		t.Fatalf("adopted goal serves = %+v, want the declared aspiration", serves)
	}
}

// TestAdoptLinksServesFallback proves the mechanical fallback: no goal_serves
// declared, but the goal text keyword/flavor-matches an aspiration → linked
// anyway. And a goal with NO signal links to nothing (no corrupting guess).
func TestAdoptLinksServesFallback(t *testing.T) {
	h := newTestHost()
	d := seededDirector(h, "mine some ore")
	h.goalGraph.Upsert(aspirationID("master a craft"), goalgraph.KindAspiration,
		"master a craft — train a skill until others seek you out for it", goalgraph.StatusActive)
	h.goalGraph.Upsert(aspirationID("know the map"), goalgraph.KindAspiration,
		"come to know your corner of the world", goalgraph.StatusActive)

	// "smelt"/"bars" land in the skill flavor family → the craft aspiration.
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "smelt the ore into bronze bars"})
	serves := h.goalGraph.Out("smelt the ore into bronze bars", goalgraph.RelServes)
	if len(serves) != 1 || serves[0].To != aspirationID("master a craft") {
		t.Fatalf("fallback serves = %+v, want the craft aspiration", serves)
	}

	// Zero-signal goal: stays unlinked rather than guessing.
	d.applyGoalOp(h, &mesaclient.Move{GoalOp: "adopt", GoalText: "rest until dawn"})
	if got := h.goalGraph.Out("rest until dawn", goalgraph.RelServes); len(got) != 0 {
		t.Fatalf("zero-signal goal should stay unlinked, got %+v", got)
	}
}

// portfolioSnapshot builds a graph with two aspirations and explicit timestamps:
// "know-the-map" untouched for an hour (NEGLECTED), "master-a-craft" about to be
// touched by the active goal's completion. Returns the snapshot's now.
func portfolioSnapshot() (goalgraph.Snapshot, int64) {
	now := time.Now().Unix()
	old := now - 3600
	return goalgraph.Snapshot{
		Nodes: []goalgraph.Node{
			{ID: "aspiration:know-the-map", Kind: goalgraph.KindAspiration, Label: "come to know the whole map", Status: goalgraph.StatusActive, At: old},
			{ID: "aspiration:master-a-craft", Kind: goalgraph.KindAspiration, Label: "master a craft", Status: goalgraph.StatusActive, At: old},
			{ID: "walk the road to Varrock", Kind: goalgraph.KindOpenGoal, Label: "walk the road to Varrock", Status: goalgraph.StatusOpen, At: now - 120},
			{ID: "smelt bronze bars", Kind: goalgraph.KindOpenGoal, Label: "smelt bronze bars", Status: goalgraph.StatusOpen, At: now - 60}, // newest
			{ID: "mine some ore", Kind: goalgraph.KindGoal, Label: "mine some ore", Status: goalgraph.StatusActive, At: now},
		},
		Edges: []goalgraph.Edge{
			{From: "walk the road to Varrock", To: "aspiration:know-the-map", Rel: goalgraph.RelServes},
			{From: "smelt bronze bars", To: "aspiration:master-a-craft", Rel: goalgraph.RelServes},
			{From: "mine some ore", To: "aspiration:master-a-craft", Rel: goalgraph.RelServes},
		},
	}, now
}

// TestAdvancementPrefersNeglectedAspiration proves the portfolio balance: when
// the active goal closes and TWO open goals are queued, the successor is the one
// serving the NEGLECTED aspiration — even though the pre-ladder newest-first
// order would have picked the other.
func TestAdvancementPrefersNeglectedAspiration(t *testing.T) {
	h := newTestHost()
	snap, _ := portfolioSnapshot()
	h.goalGraph.Import(snap)
	d := aspirationDirector("mine some ore")

	d.markGoalDone(h, "mine some ore") // touches master-a-craft fresh

	next := d.effectiveGoal(h)
	if next != "walk the road to Varrock" {
		t.Fatalf("advance picked %q; want the goal serving the neglected aspiration (newest-first would pick 'smelt bronze bars')", next)
	}
	if n, _ := h.goalGraph.Get(next); n.Status != goalgraph.StatusActive {
		t.Fatalf("picked goal not promoted to Active: %+v", n)
	}
}

// TestAdvancementKeepsNewestFirstWithoutNeglect proves the balance is a
// PREFERENCE, not a rewrite: with every aspiration freshly fed, the pre-ladder
// newest-first open_goal order stands unchanged.
func TestAdvancementKeepsNewestFirstWithoutNeglect(t *testing.T) {
	h := newTestHost()
	snap, now := portfolioSnapshot()
	for i := range snap.Nodes {
		if snap.Nodes[i].Kind == goalgraph.KindAspiration {
			snap.Nodes[i].At = now // both portfolios freshly touched
		}
	}
	h.goalGraph.Import(snap)
	d := aspirationDirector("mine some ore")

	d.markGoalDone(h, "mine some ore")

	if next := d.effectiveGoal(h); next != "smelt bronze bars" {
		t.Fatalf("with no neglect signal advance should stay newest-first, got %q", next)
	}
}

// TestMintAspirationGoal proves the attention cycle: the queue is EMPTY, an
// aspiration has gone untended past the window → advancement mints a
// deterministic day-scale goal toward it (linked serves→), instead of
// collapsing to the north-star fallback. The pure preview (effectiveGoalView)
// returns the same id WITHOUT creating the node.
func TestMintAspirationGoal(t *testing.T) {
	h := newTestHost()
	h.northStar = "the persona fallback"
	now := time.Now().Unix()
	h.goalGraph.Import(goalgraph.Snapshot{
		Nodes: []goalgraph.Node{
			{ID: "aspiration:know-the-map", Kind: goalgraph.KindAspiration, Label: "come to know the whole map", Status: goalgraph.StatusActive, At: now - 3600},
			{ID: "mine some ore", Kind: goalgraph.KindGoal, Label: "mine some ore", Status: goalgraph.StatusActive, At: now},
		},
	})
	d := aspirationDirector("mine some ore")
	d.markGoalDone(h, "mine some ore")

	want := "Make progress toward your aspiration: come to know the whole map"

	// Pure preview first: same answer, no node minted.
	if got := d.effectiveGoalView(h); got != want {
		t.Fatalf("preview = %q, want %q", got, want)
	}
	if h.goalGraph.Has(want) {
		t.Fatal("preview must not mint the goal node")
	}

	next := d.effectiveGoal(h)
	if next != want {
		t.Fatalf("advance = %q, want the minted aspiration goal", next)
	}
	n, ok := h.goalGraph.Get(want)
	if !ok || n.Kind != goalgraph.KindGoal || n.Status != goalgraph.StatusActive {
		t.Fatalf("minted goal node = %+v ok=%v", n, ok)
	}
	serves := h.goalGraph.Out(want, goalgraph.RelServes)
	if len(serves) != 1 || serves[0].To != "aspiration:know-the-map" {
		t.Fatalf("minted goal serves = %+v", serves)
	}

	// Once the minted goal CLOSES, it is never resurrected: with the aspiration
	// still neglected-by-timestamp, advancement skips the terminal mint and falls
	// through to the north star (H1).
	d.markGoalDone(h, want)
	h.goalGraph.Import(mustBackdate(h.goalGraph.Export(), want, now-3600))
	if got := d.effectiveGoal(h); got != h.northStar {
		t.Fatalf("after the mint closed, advance = %q, want the north-star fallback", got)
	}
}

// mustBackdate rewrites one node's At in a snapshot (test helper for aging state
// that the package clock writes as now).
func mustBackdate(s goalgraph.Snapshot, id string, at int64) goalgraph.Snapshot {
	for i := range s.Nodes {
		if strings.EqualFold(s.Nodes[i].ID, id) {
			s.Nodes[i].At = at
		}
	}
	return s
}

// TestAspirationAccessors proves the render-ready rollup surface the situation
// builder consumes: AspirationPortfolio (counts + neglect flag, seed order) and
// AspirationContext ("serves: ... — N goals completed toward it").
func TestAspirationAccessors(t *testing.T) {
	h := newTestHost()
	snap, now := portfolioSnapshot()
	// Add a second DONE goal toward the craft so the plural render is exercised.
	snap.Nodes = append(snap.Nodes,
		goalgraph.Node{ID: "buy a pickaxe", Kind: goalgraph.KindGoal, Label: "buy a pickaxe", Status: goalgraph.StatusDone, At: now - 30},
		goalgraph.Node{ID: "smelt first bar", Kind: goalgraph.KindGoal, Label: "smelt first bar", Status: goalgraph.StatusDone, At: now - 20})
	snap.Edges = append(snap.Edges,
		goalgraph.Edge{From: "buy a pickaxe", To: "aspiration:master-a-craft", Rel: goalgraph.RelServes},
		goalgraph.Edge{From: "smelt first bar", To: "aspiration:master-a-craft", Rel: goalgraph.RelServes})
	h.goalGraph.Import(snap)

	port := h.AspirationPortfolio()
	if len(port) != 2 {
		t.Fatalf("portfolio = %d rungs, want 2", len(port))
	}
	if port[0].ID != "aspiration:know-the-map" || !port[0].Neglected {
		t.Fatalf("first rung should be the old, NEGLECTED aspiration: %+v", port[0])
	}
	craft := port[1]
	if craft.GoalsDone != 2 || craft.GoalsActive != 1 || craft.GoalsOpen != 1 || craft.Neglected {
		t.Fatalf("craft rollup = %+v, want done=2 active=1 open=1 not neglected", craft)
	}

	got := h.AspirationContext("mine some ore")
	want := "serves: master a craft — 2 goals completed toward it"
	if got != want {
		t.Fatalf("AspirationContext = %q, want %q", got, want)
	}
	if h.AspirationContext("walk the road to Varrock") != "serves: come to know the whole map" {
		t.Fatalf("zero-completion context = %q", h.AspirationContext("walk the road to Varrock"))
	}
	if h.AspirationContext("no such goal") != "" {
		t.Fatal("unknown goal should render no aspiration context")
	}
}
