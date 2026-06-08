package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// TestKnowledgeLedgerPersists proves the world-knowledge ledger survives a
// restart through the memory layer (same spine as the trust ledger): flush, then
// a fresh host sharing the durable store reloads it.
func TestKnowledgeLedgerPersists(t *testing.T) {
	store := hostkv.NewMemory()

	h := newTestHost()
	h.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	h.knowledge.Note("Nurmof", "npc", "sells pickaxes at the Dwarven Mine", knowledge.ProvObserved, 0.9)
	h.knowledge.Seen("Nurmof", "npc")
	h.knowledge.Tag("Nurmof", "supplier")
	h.flushKnowledge(context.Background())

	h2 := newTestHost()
	h2.Memory = memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store})
	if h2.knowledge.Known("Nurmof") {
		t.Fatal("precondition: fresh ledger should not know Nurmof before load")
	}
	h2.loadKnowledge(context.Background())
	f := h2.knowledge.Get("Nurmof")
	if f.Familiar != 1 || len(f.Beliefs) != 1 || f.Confidence < 0.8 {
		t.Fatalf("after reload Nurmof = %+v, want familiar=1, 1 belief, confidence>=0.8", f)
	}
}

// TestKnowledgeBootstrapsFromMesa proves the Phase-4 push-down: a cold host (no
// local knowledge) warm-starts beliefs the consolidation cron distilled, by
// reading FetchKnowledge from the mesa sink — the round-trip that IS the proof.
func TestKnowledgeBootstrapsFromMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		fetchKnowledge: []mesaclient.KnowledgeEntry{{
			Subject: "Nurmof", Kind: "npc", Encounters: 4, LastSeenUnix: 1_700_000_500,
			Tags: []string{"supplier"},
			Beliefs: []mesaclient.KnowledgeBelief{
				{Claim: "sells pickaxes at the Dwarven Mine", Provenance: knowledge.ProvObserved, Alpha: 3, Beta: 0.3, AtUnix: 1_700_000_500},
			},
		}},
	}
	h.SetMesaMemory(sink)

	if len(h.knowledge.All()) != 0 {
		t.Fatal("precondition: a fresh host should have no local knowledge")
	}
	h.bootstrapKnowledgeFromMesa(context.Background())

	f := h.knowledge.Get("Nurmof")
	if f.Familiar != 4 || len(f.Beliefs) != 1 || f.Confidence < 0.8 {
		t.Fatalf("after bootstrap Nurmof = %+v, want familiar=4, 1 belief, confidence>=0.8 (the cron's distillation)", f)
	}
}

// TestFlushKnowledgeToMesa proves the host pushes its locally-learned beliefs up
// (host→mesa), lossless across the wire converters.
func TestFlushKnowledgeToMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true}
	h.SetMesaMemory(sink)

	h.knowledge.Note("Bob's Axes", "shop", "only sells axes, no pickaxes", knowledge.ProvObserved, 0.9)
	h.knowledge.Seen("Bob's Axes", "shop")
	h.flushKnowledgeToMesa(context.Background())

	if len(sink.syncedKnowledge) != 1 {
		t.Fatalf("expected 1 synced entry, got %d", len(sink.syncedKnowledge))
	}
	e := sink.syncedKnowledge[0]
	if e.Subject != "Bob's Axes" || e.Kind != "shop" || len(e.Beliefs) != 1 {
		t.Fatalf("synced entry not faithful: %+v", e)
	}
	if e.Beliefs[0].Claim != "only sells axes, no pickaxes" || e.Beliefs[0].Provenance != knowledge.ProvObserved {
		t.Fatalf("synced belief not faithful: %+v", e.Beliefs[0])
	}
}
