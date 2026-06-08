package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
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
