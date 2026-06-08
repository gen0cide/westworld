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
