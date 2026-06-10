package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
)

// TestMindAccessors proves the read-only inspector accessors surface the three
// mind structures (and are nil-safe shapes).
func TestMindAccessors(t *testing.T) {
	h := newTestHost()

	// Knowledge.
	h.knowledge.Note("Nurmof", "npc", "sells pickaxes", knowledge.ProvObserved, 0.9)
	facts := h.KnowledgeFacts()
	if len(facts) != 1 || facts[0].Subject != "Nurmof" {
		t.Fatalf("KnowledgeFacts = %+v", facts)
	}

	// Relationships.
	h.ledger.Met("alex")
	h.ledger.Observe("alex", true, 2)
	rels := h.Relationships()
	if len(rels) != 1 || rels[0].Name != "alex" {
		t.Fatalf("Relationships = %+v", rels)
	}

	// Goal graph + open questions.
	h.goalGraph.Upsert("smithing", goalgraph.KindGoal, "Train smithing", goalgraph.StatusActive)
	h.goalGraph.Link("iron_ore", "smithing", goalgraph.RelServes)
	h.goalGraph.Upsert("where_pickaxe", goalgraph.KindOpenQuestion, "Where do I buy a pickaxe?", goalgraph.StatusOpen)
	snap := h.GoalGraphSnapshot()
	if len(snap.Nodes) < 2 || len(snap.Edges) != 1 {
		t.Fatalf("GoalGraphSnapshot nodes=%d edges=%d", len(snap.Nodes), len(snap.Edges))
	}
	if qs := h.OpenQuestions(); len(qs) != 1 || qs[0].ID != "where_pickaxe" {
		t.Fatalf("OpenQuestions = %+v", qs)
	}
}
