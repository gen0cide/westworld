package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// TestGoalGraphPersists proves the intention graph survives a restart through the
// memory layer (same spine as the knowledge + trust ledgers).
func TestGoalGraphPersists(t *testing.T) {
	store := hostkv.NewMemory()

	h := newTestHost()
	h.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	h.goalGraph.Upsert("smithing", goalgraph.KindGoal, "Train smithing", goalgraph.StatusActive)
	h.goalGraph.Link("iron_ore", "smithing", goalgraph.RelServes)
	h.flushGoalGraph(context.Background())

	h2 := newTestHost()
	h2.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	if h2.goalGraph.Has("smithing") {
		t.Fatal("precondition: fresh graph should be empty before load")
	}
	h2.loadGoalGraph(context.Background())
	n, ok := h2.goalGraph.Get("smithing")
	if !ok || n.Status != goalgraph.StatusActive {
		t.Fatalf("after reload smithing = %+v ok=%v", n, ok)
	}
	if len(h2.goalGraph.Out("iron_ore", goalgraph.RelServes)) != 1 {
		t.Fatal("after reload the serves edge is missing")
	}
}

// TestGoalGraphConverterRoundTrip proves the wire converters are lossless and
// preserve the internal At ↔ wire AtUnix field rename across the hop.
func TestGoalGraphConverterRoundTrip(t *testing.T) {
	g := goalgraph.New()
	g.Upsert("mine-rune", goalgraph.KindGoal, "mine runite", goalgraph.StatusActive)
	g.SetProgress("mine-rune", 0.4)
	g.Tag("mine-rune", "core")
	g.Upsert("have-pickaxe", goalgraph.KindState, "have a rune pickaxe", "")
	g.Link("mine-rune", "have-pickaxe", goalgraph.RelRequires)

	orig := g.Export()
	back := snapshotFromClient(snapshotToClient(orig))

	if len(back.Nodes) != len(orig.Nodes) || len(back.Edges) != len(orig.Edges) {
		t.Fatalf("counts changed: nodes %d->%d, edges %d->%d", len(orig.Nodes), len(back.Nodes), len(orig.Edges), len(back.Edges))
	}
	var found *goalgraph.Node
	for i := range back.Nodes {
		if back.Nodes[i].ID == "mine-rune" {
			found = &back.Nodes[i]
		}
	}
	if found == nil {
		t.Fatal("mine-rune lost in round-trip")
	}
	if found.Kind != goalgraph.KindGoal || found.Status != goalgraph.StatusActive || found.Progress != 0.4 {
		t.Fatalf("node fields not faithful: %+v", found)
	}
	if found.At == 0 {
		t.Fatal("At ↔ AtUnix not preserved (got 0)")
	}
	if len(found.Tags) != 1 || found.Tags[0] != "core" {
		t.Fatalf("tags not faithful: %+v", found.Tags)
	}
	if back.Edges[0].Rel != goalgraph.RelRequires {
		t.Fatalf("edge rel not faithful: %+v", back.Edges[0])
	}
}

// TestGoalGraphBootstrapsFromMesa proves the push-down: a cold host (empty graph)
// warm-starts the graph the insight cron grew, by reading FetchGoalGraph from the
// mesa sink — the round-trip that IS the proof.
func TestGoalGraphBootstrapsFromMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		fetchGraph: mesaclient.GoalGraphSnapshot{
			Nodes: []mesaclient.GoalGraphNode{
				{ID: "rune-plate", Kind: goalgraph.KindOpenQuestion, Label: "where rune plate?", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_500},
				{ID: "champs-guild", Kind: goalgraph.KindSubgoal, Label: "champions guild quest", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_500},
			},
			Edges: []mesaclient.GoalGraphEdge{{From: "rune-plate", To: "champs-guild", Rel: goalgraph.RelRequires}},
		},
	}
	h.SetMesaMemory(sink)

	if !h.goalGraph.Empty() {
		t.Fatal("precondition: a fresh host should have an empty graph")
	}
	h.bootstrapGoalGraphFromMesa(context.Background())

	if _, ok := h.goalGraph.Get("rune-plate"); !ok {
		t.Fatal("after bootstrap rune-plate node missing")
	}
	if len(h.goalGraph.Edges()) != 1 {
		t.Fatalf("after bootstrap edges=%d, want 1 (the cron's chain)", len(h.goalGraph.Edges()))
	}
}

// TestGoalGraphBootstrapNoopOnEmptyFetch proves an empty mesa fetch does NOT
// clobber a live graph (the Import-replaces hazard the limbic Empty() guard +
// this bail-if-empty together defend against).
func TestGoalGraphBootstrapNoopOnEmptyFetch(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true} // fetchGraph zero ⇒ empty
	h.SetMesaMemory(sink)

	h.goalGraph.Upsert("live-goal", goalgraph.KindGoal, "live", goalgraph.StatusActive)
	h.bootstrapGoalGraphFromMesa(context.Background()) // empty fetch ⇒ no Import

	if _, ok := h.goalGraph.Get("live-goal"); !ok {
		t.Fatal("empty mesa fetch must not clobber the live graph")
	}
}

// TestFlushGoalGraphToMesa proves the host pushes its local intention graph up
// (host→mesa), lossless across the wire converters.
func TestFlushGoalGraphToMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true}
	h.SetMesaMemory(sink)

	h.goalGraph.Upsert("mine-coal", goalgraph.KindGoal, "mine coal", goalgraph.StatusActive)
	h.goalGraph.Link("mine-coal", "have-pickaxe", goalgraph.RelRequires)
	h.flushGoalGraphToMesa(context.Background())

	if len(sink.syncedGraph.Nodes) != 2 { // mine-coal + auto-created have-pickaxe
		t.Fatalf("expected 2 synced nodes, got %d", len(sink.syncedGraph.Nodes))
	}
	if len(sink.syncedGraph.Edges) != 1 {
		t.Fatalf("expected 1 synced edge, got %d", len(sink.syncedGraph.Edges))
	}
}

// TestGoalGraphBootstrapSelfGuardsNonEmptyLive is the L9 regression: the no-clobber
// invariant is now SELF-ENFORCED inside bootstrapGoalGraphFromMesa, not a caller
// convention. A non-empty live graph must survive a NON-EMPTY mesa fetch (the
// hazard the old caller-only Empty() guard left exposed to a future second caller).
func TestGoalGraphBootstrapSelfGuardsNonEmptyLive(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		fetchGraph: mesaclient.GoalGraphSnapshot{ // a NON-empty fetch (the would-be clobber)
			Nodes: []mesaclient.GoalGraphNode{
				{ID: "cron-node", Kind: goalgraph.KindOpenQuestion, Label: "from cron", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_500},
			},
		},
	}
	h.SetMesaMemory(sink)

	// A live graph the host built mid-run.
	h.goalGraph.Upsert("live-goal", goalgraph.KindGoal, "live", goalgraph.StatusActive)
	h.bootstrapGoalGraphFromMesa(context.Background()) // must self-guard, NOT Import

	if _, ok := h.goalGraph.Get("live-goal"); !ok {
		t.Fatal("a non-empty live graph must survive a non-empty mesa fetch (Import replaces wholesale)")
	}
	if _, ok := h.goalGraph.Get("cron-node"); ok {
		t.Fatal("the non-empty fetch must not have clobbered/merged into the live graph")
	}
}

// TestReimportGoalGraphMergesAndSurvivesFlush is the H17 host-side regression: a
// warm host RE-IMPORTS the cron-reconciled graph non-destructively (ImportMerge,
// not the Empty-guarded replace) so the cron's chain nodes/edges land in the LOCAL
// graph AND survive the host's subsequent wholesale flush-up — closing the loop the
// old whole-blob LWW broke (the cron's growth was clobbered every 30s).
func TestReimportGoalGraphMergesAndSurvivesFlush(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		// The cron-reconciled snapshot mesa serves back: a cross-entity chain.
		fetchGraph: mesaclient.GoalGraphSnapshot{
			Nodes: []mesaclient.GoalGraphNode{
				{ID: "rune-plate", Kind: goalgraph.KindOpenGoal, Label: "obtain rune platebody", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_900},
			},
			Edges: []mesaclient.GoalGraphEdge{{From: "mine-ore", To: "rune-plate", Rel: goalgraph.RelServes}},
		},
	}
	h.SetMesaMemory(sink)

	// A NON-EMPTY live graph (the warm case the cold-start bootstrap deliberately
	// skips): the host's own active goal.
	h.goalGraph.Upsert("mine-ore", goalgraph.KindGoal, "mine ore", goalgraph.StatusActive)
	h.goalGraph.SetProgress("mine-ore", 0.5)

	h.reimportGoalGraphFromMesa(context.Background())

	// The host's live status/progress survives (authority partition).
	if n, _ := h.goalGraph.Get("mine-ore"); n.Status != goalgraph.StatusActive || n.Progress != 0.5 {
		t.Fatalf("warm re-import clobbered the live goal: %+v", n)
	}
	// The cron chain landed in the LOCAL graph.
	if !h.goalGraph.Has("rune-plate") {
		t.Fatal("warm re-import did not fold the cron chain node into the local graph")
	}

	// Now the host flushes UP — the cron node must be carried, not clobbered.
	h.flushGoalGraphToMesa(context.Background())
	got := map[string]bool{}
	for _, n := range sink.syncedGraph.Nodes {
		got[n.ID] = true
	}
	if !got["mine-ore"] || !got["rune-plate"] {
		t.Fatalf("host flush dropped a node: synced=%v (cron chain must survive the flush)", got)
	}
}

// TestPruneCognitionBoundsGoalGraph is the H10 regression: the flush-tick prune
// drops the OLDEST TERMINAL goal-graph nodes once over the cap, so the persisted
// blob stays bounded instead of growing monotonically as goals close.
func TestPruneCognitionBoundsGoalGraph(t *testing.T) {
	h := newTestHost()
	// One live (active) node that must always survive, plus many DONE terminal nodes
	// well over the DefaultGraphCap so prune has to evict.
	h.goalGraph.Upsert("live", goalgraph.KindGoal, "live goal", goalgraph.StatusActive)
	for i := 0; i < goalgraph.DefaultGraphCap+20; i++ {
		id := "done-" + itoa(i)
		h.goalGraph.Upsert(id, goalgraph.KindGoal, "old goal", goalgraph.StatusDone)
	}
	before := len(h.goalGraph.Nodes())
	h.pruneCognition()
	after := len(h.goalGraph.Nodes())
	if after > goalgraph.DefaultGraphCap {
		t.Fatalf("prune should bound the graph to <= %d nodes, got %d (was %d)", goalgraph.DefaultGraphCap, after, before)
	}
	if _, ok := h.goalGraph.Get("live"); !ok {
		t.Fatal("prune must never evict a non-terminal (active) node")
	}
}

// itoa is a tiny dependency-free int->string for the table above.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
