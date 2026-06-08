package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// fakeSink is a MesaMemory that records mirrored episodes and serves canned
// recalls, so the host's two-way LTM seam can be tested without a live mesa.
type fakeSink struct {
	healthy         bool
	got             chan *mesaclient.Episode
	obs             chan *mesaclient.Observation // RecordObservation capture (nil ⇒ drop)
	recall          []mesaclient.KnowledgeItem
	synced          []mesaclient.Relationship    // last SyncRelationships payload
	fetchRels       []mesaclient.Relationship    // what FetchRelationships returns
	syncedGoal      mesaclient.Goal              // last SyncGoal payload
	fetchGoal       mesaclient.Goal              // what FetchGoal returns (found iff Objective != "")
	reportedMetrics []mesaclient.Metric          // last ReportMetrics payload
	syncedKnowledge []mesaclient.KnowledgeEntry  // last SyncKnowledge payload
	fetchKnowledge  []mesaclient.KnowledgeEntry  // what FetchKnowledge returns
	syncedGraph     mesaclient.GoalGraphSnapshot // last SyncGoalGraph payload
	fetchGraph      mesaclient.GoalGraphSnapshot // what FetchGoalGraph returns
}

func (f *fakeSink) Remember(_ context.Context, e *mesaclient.Episode) error {
	f.got <- e
	return nil
}

func (f *fakeSink) RecordObservation(_ context.Context, o *mesaclient.Observation) error {
	if f.obs != nil {
		f.obs <- o
	}
	return nil
}
func (f *fakeSink) Recall(_ context.Context, _ *mesaclient.Query) (*mesaclient.Knowledge, error) {
	return &mesaclient.Knowledge{Items: f.recall}, nil
}
func (f *fakeSink) SyncRelationships(_ context.Context, _ string, rels []mesaclient.Relationship) error {
	f.synced = rels
	return nil
}
func (f *fakeSink) FetchRelationships(_ context.Context, _ string) ([]mesaclient.Relationship, error) {
	return f.fetchRels, nil
}
func (f *fakeSink) SyncGoal(_ context.Context, _ string, g mesaclient.Goal) error {
	f.syncedGoal = g
	return nil
}
func (f *fakeSink) FetchGoal(_ context.Context, _ string) (mesaclient.Goal, bool, error) {
	return f.fetchGoal, f.fetchGoal.Objective != "", nil
}
func (f *fakeSink) ReportMetrics(_ context.Context, _ string, metrics []mesaclient.Metric) error {
	f.reportedMetrics = metrics
	return nil
}
func (f *fakeSink) SyncKnowledge(_ context.Context, _ string, entries []mesaclient.KnowledgeEntry) error {
	f.syncedKnowledge = entries
	return nil
}
func (f *fakeSink) FetchKnowledge(_ context.Context, _ string) ([]mesaclient.KnowledgeEntry, error) {
	return f.fetchKnowledge, nil
}
func (f *fakeSink) SyncGoalGraph(_ context.Context, _ string, snap mesaclient.GoalGraphSnapshot) error {
	f.syncedGraph = snap
	return nil
}
func (f *fakeSink) FetchGoalGraph(_ context.Context, _ string) (mesaclient.GoalGraphSnapshot, error) {
	return f.fetchGraph, nil
}
func (f *fakeSink) Healthy() bool { return f.healthy }

// TestMemoryCaptureMirrors proves a captured episode is both stored locally AND
// pushed UP to the mesa sink (the #13a mirror path), with the host's identity
// and content intact.
func TestMemoryCaptureMirrors(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true, got: make(chan *mesaclient.Episode, 1)}
	h.SetMesaMemory(sink)

	h.memoryCapture(event.LevelUp{Skill: event.SkillAttack, NewLevel: 5})

	if h.journal.Len() != 1 {
		t.Fatalf("local journal len=%d, want 1", h.journal.Len())
	}
	select {
	case e := <-sink.got:
		if e.HostID != "test" {
			t.Fatalf("mirrored HostID=%q, want test", e.HostID)
		}
		if e.Kind != "milestone" || !strings.Contains(e.Text, "level 5") {
			t.Fatalf("mirrored episode = %+v", e)
		}
		if e.IdempotencyKey == "" {
			t.Fatal("mirrored episode has no idempotency key")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("episode was not mirrored to mesa")
	}
}

// TestMemoryMirrorSkippedWhenUnhealthy proves an unhealthy/absent sink is a
// no-op for the mirror (local capture still happens) — mesa is never required.
func TestMemoryMirrorSkippedWhenUnhealthy(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: false, got: make(chan *mesaclient.Episode, 1)}
	h.SetMesaMemory(sink)

	h.memoryCapture(event.Death{})

	if h.journal.Len() != 1 {
		t.Fatalf("local capture should still happen: len=%d", h.journal.Len())
	}
	select {
	case <-sink.got:
		t.Fatal("unhealthy sink should not receive a mirror")
	case <-time.After(200 * time.Millisecond):
		// expected: nothing mirrored
	}
}

// TestLedgerMirrorsUpToMesa proves flushLimbic pushes the trust-ledger snapshot
// up to mesa (host→mesa, AuthLocal mirror).
func TestLedgerMirrorsUpToMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true, got: make(chan *mesaclient.Episode, 1)}
	h.SetMesaMemory(sink)

	h.ledger.Observe("alex", true, 1)
	h.ledger.Met("alex")
	h.flushLimbic(context.Background()) // Memory nil → local skipped; mesa sync runs

	if len(sink.synced) != 1 || sink.synced[0].Name != "alex" {
		t.Fatalf("ledger not mirrored up: %+v", sink.synced)
	}
	if sink.synced[0].Alpha <= 1 || sink.synced[0].Encounters != 1 {
		t.Fatalf("snapshot not faithful: %+v", sink.synced[0])
	}
}

// TestLedgerBootstrapsFromMesa proves a cold-start host with no local ledger
// reconstitutes it from mesa (the authoritative bootstrap path).
func TestLedgerBootstrapsFromMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		got:     make(chan *mesaclient.Episode, 1),
		fetchRels: []mesaclient.Relationship{
			{Name: "alex", Alpha: 5, Beta: 1, Encounters: 4, Tags: []string{"generous"}},
		},
	}
	h.SetMesaMemory(sink)

	if h.ledger.Known("alex") {
		t.Fatal("precondition: ledger should be empty")
	}
	h.bootstrapLedgerFromMesa(context.Background())

	if !h.ledger.Known("alex") {
		t.Fatal("ledger not bootstrapped from mesa")
	}
	if r := h.ledger.Rel("alex"); r.Familiar != 4 || r.Trust <= 0 {
		t.Fatalf("bootstrapped alex = %+v, want familiar=4 trust>0", r)
	}
}

// TestGoalMirrorsUpToMesa proves flushJournal mirrors the standing objective +
// progress to mesa's structured goals seam.
func TestGoalMirrorsUpToMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true, got: make(chan *mesaclient.Episode, 1)}
	h.SetMesaMemory(sink)

	h.journal.SetObjective("become free")
	h.journal.Append("milestone", "Reached level 5 in mining.", 0.7, "")
	h.flushJournal(context.Background()) // Memory nil → local skipped; goal sync runs

	if sink.syncedGoal.Objective != "become free" {
		t.Fatalf("goal objective not mirrored: %+v", sink.syncedGoal)
	}
	if len(sink.syncedGoal.Progress) != 1 || sink.syncedGoal.Progress[0] != "Reached level 5 in mining." {
		t.Fatalf("goal progress not mirrored: %+v", sink.syncedGoal.Progress)
	}
}

// TestGoalBootstrapsFromMesa proves a cold-start host resumes its objective from
// mesa's structured goals table.
func TestGoalBootstrapsFromMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy:   true,
		got:       make(chan *mesaclient.Episode, 1),
		fetchGoal: mesaclient.Goal{Objective: "become a master chef"},
	}
	h.SetMesaMemory(sink)

	if obj, _ := h.journal.Objective(); obj != "" {
		t.Fatal("precondition: no objective")
	}
	h.bootstrapGoalFromMesa(context.Background())
	if obj, _ := h.journal.Objective(); obj != "become a master chef" {
		t.Fatalf("objective not resumed from mesa: %q", obj)
	}
}

// TestMetricsReportToMesa proves the host assembles + ships a telemetry batch.
func TestMetricsReportToMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{healthy: true, got: make(chan *mesaclient.Episode, 1)}
	h.SetMesaMemory(sink)
	h.journal.Append("milestone", "Reached level 5 in mining.", 0.7, "")

	h.reportMetrics(context.Background(), 3, 10, 42*time.Second)

	m := map[string]float64{}
	for _, x := range sink.reportedMetrics {
		m[x.Name] = x.Value
	}
	if m["pearl.vetoes"] != 3 || m["agent.turns"] != 10 || m["host.uptime_seconds"] != 42 || m["journal.episodes"] != 1 {
		t.Fatalf("metrics batch wrong: %+v", sink.reportedMetrics)
	}
}

// TestBootstrapJournalFromMesa proves a cold-start host with no local journal
// reconstitutes it from the mesa sink (the authoritative bootstrap path).
func TestBootstrapJournalFromMesa(t *testing.T) {
	h := newTestHost()
	sink := &fakeSink{
		healthy: true,
		got:     make(chan *mesaclient.Episode, 4),
		// mesa returns newest-first; bootstrap reverses to chronological.
		recall: []mesaclient.KnowledgeItem{
			{Text: "Died and respawned.", Provenance: "death@2026-06-07T12:00:00Z"},
			{Text: "Reached level 5 in mining.", Provenance: "milestone@2026-06-07T11:00:00Z"},
		},
	}
	h.SetMesaMemory(sink)

	if h.journal.Len() != 0 {
		t.Fatalf("precondition: journal should be empty, got %d", h.journal.Len())
	}
	h.bootstrapJournalFromMesa(context.Background())

	if h.journal.Len() != 2 {
		t.Fatalf("bootstrap journal len=%d, want 2", h.journal.Len())
	}
	rec := h.journal.Recent(2)
	if rec[0].Text != "Reached level 5 in mining." || rec[1].Text != "Died and respawned." {
		t.Fatalf("bootstrap order wrong: %q then %q", rec[0].Text, rec[1].Text)
	}
	if rec[0].Kind != "milestone" || rec[1].Kind != "death" {
		t.Fatalf("kinds not parsed from provenance: %q, %q", rec[0].Kind, rec[1].Kind)
	}
}
