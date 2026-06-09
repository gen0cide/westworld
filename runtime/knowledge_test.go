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

// TestKnowledgeBootstrapSelfGuardsNonEmptyLive is the L9 regression: the no-clobber
// invariant is now SELF-ENFORCED inside bootstrapKnowledgeFromMesa. A host that
// already learned locally must NOT be overwritten by a non-empty mesa fetch.
func TestKnowledgeBootstrapSelfGuardsNonEmptyLive(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		fetchKnowledge: []mesaclient.KnowledgeEntry{{ // a non-empty fetch (the would-be clobber)
			Subject: "Nurmof", Kind: "npc", Encounters: 9,
			Beliefs: []mesaclient.KnowledgeBelief{{Claim: "from cron", Provenance: knowledge.ProvObserved, Alpha: 5, Beta: 1}},
		}},
	}
	h.SetMesaMemory(sink)

	// The host already learned something locally.
	h.knowledge.Note("Bob's Axes", "shop", "only sells axes", knowledge.ProvObserved, 0.9)
	h.bootstrapKnowledgeFromMesa(context.Background()) // must self-guard, NOT Import

	if !h.knowledge.Known("Bob's Axes") {
		t.Fatal("a host with local knowledge must keep it (Import replaces wholesale)")
	}
	if h.knowledge.Known("Nurmof") {
		t.Fatal("the non-empty fetch must not have clobbered the live ledger")
	}
}

// TestReimportKnowledgeMergesAndSurvivesFlush is the M17 host-side regression: a
// warm host RE-IMPORTS the cron-reconciled ledger non-destructively (ImportMerge,
// not the cold-start-only replace) so a cron belief the host never re-learned lands
// in the LOCAL ledger AND survives the host's subsequent flush-up, while the host's
// own live belief is never wiped.
func TestReimportKnowledgeMergesAndSurvivesFlush(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		// The cron-reconciled ledger mesa serves back: a belief only the cron knows.
		fetchKnowledge: []mesaclient.KnowledgeEntry{{
			Subject: "Bob's Axes", Kind: "shop", Encounters: 9, LastSeenUnix: 1_700_000_900,
			Beliefs: []mesaclient.KnowledgeBelief{
				{Claim: "out of bronze (cron-reconciled)", Provenance: knowledge.ProvObserved, Alpha: 1, Beta: 4, AtUnix: 1_700_000_900},
			},
		}},
	}
	h.SetMesaMemory(sink)

	// A NON-EMPTY live ledger (the warm case): the host's own learned belief.
	h.knowledge.Note("Bob's Axes", "shop", "sells axes", knowledge.ProvObserved, 0.9)

	h.reimportKnowledgeFromMesa(context.Background())

	f := h.knowledge.Get("Bob's Axes")
	claims := map[string]bool{}
	for _, b := range f.Beliefs {
		claims[b.Claim] = true
	}
	if !claims["sells axes"] {
		t.Fatal("warm re-import wiped the host's own live belief")
	}
	if !claims["out of bronze (cron-reconciled)"] {
		t.Fatal("warm re-import did not fold the cron belief into the local ledger")
	}

	// Now the host flushes UP — the cron belief must be carried, not clobbered.
	h.flushKnowledgeToMesa(context.Background())
	if len(sink.syncedKnowledge) != 1 {
		t.Fatalf("expected 1 synced subject, got %d", len(sink.syncedKnowledge))
	}
	synced := map[string]bool{}
	for _, b := range sink.syncedKnowledge[0].Beliefs {
		synced[b.Claim] = true
	}
	if !synced["sells axes"] || !synced["out of bronze (cron-reconciled)"] {
		t.Fatalf("host flush dropped a belief: synced=%v (cron belief must survive the flush)", synced)
	}
}

// TestPruneCognitionBoundsKnowledge is the H10 regression: the flush-tick prune
// trims the subject count down to the cap (subjects accrue with NO inline bound —
// only Prune enforces it), so the persisted ledger stays bounded instead of growing
// monotonically per distinct subject seen.
func TestPruneCognitionBoundsKnowledge(t *testing.T) {
	h := newTestHost()
	const maxSubj = 8
	h.knowledge.SetCaps(maxSubj, knowledge.DefaultMaxBeliefsPerEntry) // tight subject cap
	for i := 0; i < maxSubj+12; i++ {
		h.knowledge.Seen("subject-"+itoa(i), "npc") // distinct subjects, no inline cap
	}
	if got := len(h.knowledge.All()); got <= maxSubj {
		t.Fatalf("precondition: expected >%d subjects before prune, got %d", maxSubj, got)
	}
	h.pruneCognition()
	if got := len(h.knowledge.All()); got != maxSubj {
		t.Fatalf("prune should trim subjects to the cap of %d, got %d", maxSubj, got)
	}
}
