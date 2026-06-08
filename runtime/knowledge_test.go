package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
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
