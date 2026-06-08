package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// fakeCompleter is a stub LLM (no real API calls): it records every call and
// returns canned JSON. It satisfies the cron's `completer` seam.
type fakeCompleter struct {
	calls    atomic.Int32
	lastUser string
	resp     string // canned JSON response
}

func (f *fakeCompleter) CompleteSystem(_ context.Context, _ []llm.SystemBlock, user string, _ int) (string, error) {
	f.calls.Add(1)
	f.lastUser = user
	return f.resp, nil
}

func quietServer() *Server {
	return New(nil, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func obs(key, kind, subject, body string, sal float64, at int64) *mesapb.Observation {
	return &mesapb.Observation{IdempotencyKey: key, Kind: kind, Subject: subject, Text: body, Salience: sal, OccurredAtUnix: at}
}

// TestConsolidateHostFoldsAndIsIdempotent proves the bulk path: a batch of
// dialog observations folds into the durable knowledge ledger via the fake Haiku,
// the cursor advances, a second run does NOT reprocess (no LLM call, no dup), and
// re-folding the same claim reinforces rather than duplicates (idempotent).
func TestConsolidateHostFoldsAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)

	fake := &fakeCompleter{resp: `{"claims":[
		{"subject":"Nurmof","kind":"npc","claim":"sells pickaxes at the Dwarven Mine","confidence":0.8,"provenance":"player-told","sentiment":"positive"},
		{"subject":"Bob's Axes","kind":"shop","claim":"only sells axes, no pickaxes","confidence":0.7,"provenance":"npc-told"}
	]}`}
	s.consolidateLLMOverride = fake

	// Seed firehose observations.
	for i, o := range []*mesapb.Observation{
		obs("o1", "player_chat", "Nurmof", "go to Nurmof for pickaxes", 0.6, 1_700_000_100),
		obs("o2", "npc_dialog", "Bob's Axes", "we only sell axes here", 0.5, 1_700_000_200),
		obs("o3", "player_chat", "weather", "nice day innit", 0.1, 1_700_000_300),
	} {
		if _, err := l.AddObservation(ctx, host, o); err != nil {
			t.Fatalf("seed obs %d: %v", i, err)
		}
	}

	cfg := DefaultCronConfig()
	if err := s.consolidateHost(ctx, host, cfg); err != nil {
		t.Fatalf("consolidateHost: %v", err)
	}
	if fake.calls.Load() != 1 {
		t.Fatalf("expected exactly 1 Haiku call, got %d", fake.calls.Load())
	}

	// The fold landed in the durable knowledge store.
	got, err := l.Knowledge(ctx, host)
	if err != nil {
		t.Fatalf("Knowledge: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 distilled subjects, got %d: %+v", len(got), got)
	}
	bysubj := map[string]*mesapb.KnowledgeEntry{}
	for _, e := range got {
		bysubj[e.GetSubject()] = e
	}
	nurmof := bysubj["Nurmof"]
	if nurmof == nil || nurmof.GetEncounters() < 1 || len(nurmof.GetBeliefs()) != 1 {
		t.Fatalf("Nurmof not folded: %+v", nurmof)
	}
	hasSentiment := false
	for _, tag := range nurmof.GetTags() {
		if tag == "sentiment:positive" {
			hasSentiment = true
		}
	}
	if !hasSentiment {
		t.Fatalf("Nurmof sentiment tag not applied: %+v", nurmof.GetTags())
	}

	// Cursor advanced to the newest observed second of the FULL batch (o3 @ 300).
	cur := s.loadCursor(ctx, host)
	if cur.LastUnix != 1_700_000_300 {
		t.Fatalf("cursor = %d, want 1700000300", cur.LastUnix)
	}

	// Second run: no NEW observations past the cursor → no LLM call, no change.
	if err := s.consolidateHost(ctx, host, cfg); err != nil {
		t.Fatalf("consolidateHost (2nd): %v", err)
	}
	if fake.calls.Load() != 1 {
		t.Fatalf("2nd run made an LLM call (reprocessed): calls=%d", fake.calls.Load())
	}

	// Idempotent re-fold: add a NEW observation that restates Nurmof's claim →
	// the SAME claim reinforces (still 1 belief), encounters bump, no duplicate.
	if _, err := l.AddObservation(ctx, host, obs("o4", "player_chat", "Nurmof", "Nurmof has pickaxes", 0.6, 1_700_000_400)); err != nil {
		t.Fatalf("seed o4: %v", err)
	}
	fake.resp = `{"claims":[{"subject":"Nurmof","kind":"npc","claim":"sells pickaxes at the Dwarven Mine","confidence":0.8,"provenance":"player-told"}]}`
	if err := s.consolidateHost(ctx, host, cfg); err != nil {
		t.Fatalf("consolidateHost (3rd): %v", err)
	}
	got, _ = l.Knowledge(ctx, host)
	for _, e := range got {
		if e.GetSubject() == "Nurmof" {
			if len(e.GetBeliefs()) != 1 {
				t.Fatalf("idempotency: Nurmof grew to %d beliefs, want 1 (reinforce, not dup)", len(e.GetBeliefs()))
			}
			if e.GetEncounters() < 2 {
				t.Fatalf("idempotency: Nurmof encounters=%d, want >=2", e.GetEncounters())
			}
		}
	}
}

// TestConsolidateEscalationRecorded proves the escalate flag is captured to the
// durable queue (4a record-only) and nothing else consumes it.
func TestConsolidateEscalationRecorded(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	s.consolidateLLMOverride = &fakeCompleter{resp: `{"claims":[
		{"subject":"rune plate","kind":"item","claim":"needs champions guild quest","confidence":0.6,"provenance":"player-told","escalate":true,"escalate_reason":"touches an open question"}
	]}`}

	if _, err := l.AddObservation(ctx, host, obs("e1", "player_chat", "rune plate", "you need the champs guild quest", 0.7, 1_700_000_100)); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := s.consolidateHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("consolidateHost: %v", err)
	}
	v, ok, err := l.GetKV(ctx, host, cronEscalateQueueKey)
	if err != nil || !ok {
		t.Fatalf("escalate queue not written: ok=%v err=%v", ok, err)
	}
	var q []map[string]any
	if err := json.Unmarshal(v, &q); err != nil || len(q) != 1 {
		t.Fatalf("escalate queue = %s (err %v)", v, err)
	}
	if q[0]["subject"] != "rune plate" {
		t.Fatalf("escalated wrong subject: %+v", q[0])
	}
}

// TestSyncKnowledgeRoundTrip proves the push-down store seam: a pushed snapshot
// is stored, fetched back faithfully, and a re-push REPLACES per subject.
func TestSyncKnowledgeRoundTrip(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)

	entries := []*mesapb.KnowledgeEntry{{
		Subject: "Nurmof", Kind: "npc", Encounters: 3, LastSeenUnix: 1_700_000_500,
		Tags: []string{"supplier"},
		Beliefs: []*mesapb.KnowledgeBelief{
			{Claim: "sells pickaxes", Provenance: "observed", Alpha: 2.7, Beta: 0.3, AtUnix: 1_700_000_500},
		},
	}}
	n, err := l.SyncKnowledge(ctx, host, entries)
	if err != nil || n != 1 {
		t.Fatalf("SyncKnowledge n=%d err=%v", n, err)
	}
	got, err := l.Knowledge(ctx, host)
	if err != nil || len(got) != 1 {
		t.Fatalf("Knowledge len=%d err=%v", len(got), err)
	}
	e := got[0]
	if e.GetSubject() != "Nurmof" || e.GetKind() != "npc" || e.GetEncounters() != 3 || len(e.GetTags()) != 1 {
		t.Fatalf("entry not faithful: %+v", e)
	}
	if len(e.GetBeliefs()) != 1 || e.GetBeliefs()[0].GetClaim() != "sells pickaxes" || e.GetBeliefs()[0].GetAlpha() != 2.7 {
		t.Fatalf("belief not faithful: %+v", e.GetBeliefs())
	}

	// Re-push with a new absolute value → REPLACE that subject.
	entries[0].Encounters = 9
	if _, err := l.SyncKnowledge(ctx, host, entries); err != nil {
		t.Fatalf("re-sync: %v", err)
	}
	got, _ = l.Knowledge(ctx, host)
	if got[0].GetEncounters() != 9 {
		t.Fatalf("snapshot did not replace: encounters=%d, want 9", got[0].GetEncounters())
	}
}

// TestSyncGoalGraphRoundTrip proves the goal-graph push-down store seam: a pushed
// snapshot is stored, fetched back faithfully, ErrNoRows degrades to empty, and a
// nil/empty push is a no-op (never clobbers a stored graph).
func TestSyncGoalGraphRoundTrip(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)

	// Missing row ⇒ empty snapshot (not an error).
	if snap, err := l.GoalGraph(ctx, host); err != nil || len(snap.GetNodes()) != 0 {
		t.Fatalf("cold GoalGraph should be empty: %+v err=%v", snap, err)
	}

	snap := &mesapb.GoalGraphSnapshot{
		Nodes: []*mesapb.GoalGraphNode{
			{Id: "rune-plate", Kind: "open_question", Label: "where rune plate?", Status: "open", AtUnix: 1_700_000_500, Tags: []string{"core"}},
			{Id: "champs-guild", Kind: "subgoal", Label: "champions guild quest", Status: "open", AtUnix: 1_700_000_500},
		},
		Edges: []*mesapb.GoalGraphEdge{{From: "rune-plate", To: "champs-guild", Rel: "requires"}},
	}
	if err := l.SyncGoalGraph(ctx, host, snap); err != nil {
		t.Fatalf("SyncGoalGraph: %v", err)
	}
	got, err := l.GoalGraph(ctx, host)
	if err != nil {
		t.Fatalf("GoalGraph: %v", err)
	}
	if len(got.GetNodes()) != 2 || len(got.GetEdges()) != 1 {
		t.Fatalf("round-trip lost data: nodes=%d edges=%d", len(got.GetNodes()), len(got.GetEdges()))
	}
	if got.GetNodes()[0].GetAtUnix() != 1_700_000_500 {
		t.Fatalf("AtUnix not faithful: %+v", got.GetNodes()[0])
	}

	// Nil/empty push is a no-op (must NOT clobber the stored graph).
	if err := l.SyncGoalGraph(ctx, host, &mesapb.GoalGraphSnapshot{}); err != nil {
		t.Fatalf("empty SyncGoalGraph: %v", err)
	}
	if again, _ := l.GoalGraph(ctx, host); len(again.GetNodes()) != 2 {
		t.Fatalf("empty push clobbered the graph: nodes=%d, want 2", len(again.GetNodes()))
	}

	// LWW replace per host.
	snap.Nodes = snap.Nodes[:1]
	if err := l.SyncGoalGraph(ctx, host, snap); err != nil {
		t.Fatalf("re-sync: %v", err)
	}
	if again, _ := l.GoalGraph(ctx, host); len(again.GetNodes()) != 1 {
		t.Fatalf("LWW replace failed: nodes=%d, want 1", len(again.GetNodes()))
	}
}

// TestGCKnowledgeDecaysAndBounds proves Tier-0 GC (no LLM): stale+weak+rare
// subjects are pruned by TTL, the per-host count is bounded, and `system`-
// provenance subjects are pinned (never evicted).
func TestGCKnowledgeDecaysAndBounds(t *testing.T) {
	now := time.Unix(1_800_000_000, 0)
	old := now.Add(-40 * 24 * time.Hour).Unix()
	fresh := now.Add(-1 * time.Hour).Unix()

	cfg := DefaultCronConfig()
	cfg.MaxSubjects = 2 // force a size-bound eviction

	rows := []knowledge.Entry{
		// stale + weak + barely-seen → TTL-pruned.
		{Subject: "ghost", Kind: "concept", LastSeen: old, Encounters: 1,
			Beliefs: []knowledge.Belief{{Claim: "x", Provenance: knowledge.ProvHearsay, Alpha: 0.1, Beta: 0.9}}},
		// stale but system-provenance → PINNED, survives.
		{Subject: "law", Kind: "mechanic", LastSeen: old, Encounters: 1,
			Beliefs: []knowledge.Belief{{Claim: "server rule", Provenance: knowledge.ProvSystem, Alpha: 1, Beta: 0}}},
		// fresh, frequently seen → high score, kept.
		{Subject: "Nurmof", Kind: "npc", LastSeen: fresh, Encounters: 20,
			Beliefs: []knowledge.Belief{{Claim: "sells pickaxes", Provenance: knowledge.ProvObserved, Alpha: 5, Beta: 1}}},
		// fresh but rarely seen → lowest score among the survivors; evicted by the cap.
		{Subject: "rando", Kind: "npc", LastSeen: fresh, Encounters: 0,
			Beliefs: []knowledge.Belief{{Claim: "said hi", Provenance: knowledge.ProvHearsay, Alpha: 0.5, Beta: 0.5}}},
	}

	out := gcKnowledge(rows, cfg, now)

	kept := map[string]bool{}
	for _, e := range out {
		kept[e.Subject] = true
	}
	if kept["ghost"] {
		t.Fatal("stale+weak+rare 'ghost' should have been TTL-pruned")
	}
	if !kept["law"] {
		t.Fatal("system-provenance 'law' must be pinned (never evicted), even when stale")
	}
	if !kept["Nurmof"] {
		t.Fatal("fresh+frequent 'Nurmof' must be kept")
	}
	if len(out) > cfg.MaxSubjects {
		t.Fatalf("GC did not bound size: %d subjects, cap %d", len(out), cfg.MaxSubjects)
	}
}

// TestParseConsolidatedClaims proves the JSON parse: well-formed claims preserved
// (empty-content dropped), well-formed empty advances (no error), malformed errors
// (retryable batch).
func TestParseConsolidatedClaims(t *testing.T) {
	ok := `prefix {"claims":[
		{"subject":"Nurmof","claim":"sells pickaxes","confidence":0.8,"escalate":true},
		{"subject":"x","claim":"","confidence":0.1}
	]} suffix`
	claims, err := parseConsolidatedClaims(ok)
	if err != nil {
		t.Fatalf("well-formed errored: %v", err)
	}
	if len(claims) != 1 || claims[0].Subject != "Nurmof" || !claims[0].Escalate {
		t.Fatalf("claims = %+v", claims)
	}

	empty, err := parseConsolidatedClaims(`{"claims":[]}`)
	if err != nil || len(empty) != 0 {
		t.Fatalf("well-formed empty should be (nil-error, 0 claims): %v %+v", err, empty)
	}

	for _, bad := range []string{"", "not json", "{garbage"} {
		if _, err := parseConsolidatedClaims(bad); err == nil {
			t.Fatalf("malformed %q should error (retryable), got nil", bad)
		}
	}
}

// TestConsolidateTieringIsHaikuOnly proves the cardinal cost invariant: the
// consolidation seam is decideLLM (Haiku), NEVER actLLM/genesisLLM. With only a
// decideLLM wired, consolidateLLM resolves to it; actLLM/genesisLLM are untouched.
func TestConsolidateTieringIsHaikuOnly(t *testing.T) {
	haiku := llm.New("test-key", "claude-haiku-4-5")
	sonnet := llm.New("test-key", "claude-sonnet-4-6")
	opus := llm.New("test-key", "claude-opus-4-8")
	s := New(sonnet, haiku, opus, slog.New(slog.NewTextHandler(io.Discard, nil)))

	c := s.consolidateLLM()
	got, ok := c.(*llm.Client)
	if !ok {
		t.Fatalf("consolidateLLM did not resolve to *llm.Client: %T", c)
	}
	if got.Model() != haiku.Model() {
		t.Fatalf("consolidation routed to %q, want the Haiku tier %q (never Sonnet/Opus on bulk)", got.Model(), haiku.Model())
	}
	if got.Model() == sonnet.Model() || got.Model() == opus.Model() {
		t.Fatal("consolidation must NEVER be the Sonnet/Opus tier")
	}
}

// TestEmptySweepNoLLM proves a host with no NEW observations makes NO LLM call
// (no spurious cost) — the empty-firehose invariant, asserted deterministically
// on a single isolated host (the global StartCrons sweep can't be asserted
// against a shared DB).
func TestEmptySweepNoLLM(t *testing.T) {
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	fake := &fakeCompleter{resp: `{"claims":[]}`}
	s.consolidateLLMOverride = fake

	if err := s.consolidateHost(context.Background(), host, DefaultCronConfig()); err != nil {
		t.Fatalf("consolidateHost (no obs): %v", err)
	}
	if fake.calls.Load() != 0 {
		t.Fatalf("empty firehose made %d spurious LLM calls, want 0", fake.calls.Load())
	}
}

// TestStartStopCronsClean proves the lifecycle launches and shuts down cleanly:
// StartCrons begins the loop and StopCrons drains any in-flight folds and returns
// (no hang, no torn write). It is idempotent (safe to call StopCrons twice).
func TestStartStopCronsClean(t *testing.T) {
	l, _ := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	s.consolidateLLMOverride = &fakeCompleter{resp: `{"claims":[]}`}

	cfg := DefaultCronConfig()
	cfg.ConsolidateEvery = 20 * time.Millisecond // tick fast so the loop actually runs
	cfg.ActiveWindow = time.Second               // only just-inserted rows qualify (we insert none)
	done := make(chan struct{})
	go func() {
		s.StartCrons(context.Background(), cfg)
		time.Sleep(80 * time.Millisecond) // let several ticks fire
		s.StopCrons()                     // must drain cleanly and return
		s.StopCrons()                     // idempotent
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("StartCrons/StopCrons did not complete cleanly (hang)")
	}
}

// TestStartCronsNoLTMNoop proves StartCrons no-ops without a durable store (so
// StopCrons is safe even when crons never really started).
func TestStartCronsNoLTMNoop(t *testing.T) {
	s := quietServer() // no LTM
	s.StartCrons(context.Background(), DefaultCronConfig())
	s.StopCrons() // must not panic / hang
}

// openCronLTM opens the test LTM and registers a cleanup that also clears the
// knowledge + observation + kv rows this suite writes.
func openCronLTM(t *testing.T) (*LTM, string) {
	t.Helper()
	l, err := OpenLTM(context.Background(), testDSN(), nil)
	if err != nil {
		t.Skipf("postgres not reachable (%v); skipping cron LTM test", err)
	}
	host := fmt.Sprintf("test_cron_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM observations WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM knowledge WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM goal_graphs WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM kv WHERE host_id = $1`, host)
		l.Close()
	})
	return l, host
}
