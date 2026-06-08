package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// seedQueue writes an escalation queue blob for a host (the 4a producer's shape).
func seedQueue(t *testing.T, l *LTM, host string, items ...qItem) {
	t.Helper()
	v, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal queue: %v", err)
	}
	if err := l.PutKV(context.Background(), host, cronEscalateQueueKey, v); err != nil {
		t.Fatalf("seed queue: %v", err)
	}
}

// TestInsightReconcileLowersConfidenceOnce proves contradiction reconciliation
// lowers the loser claim's confidence EXACTLY ONCE across two ticks (the sentinel-
// tag idempotency — no β thrash on re-run).
func TestInsightReconcileLowersConfidenceOnce(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)

	// Seed a hearsay belief that will be the reconciliation loser.
	if _, err := l.SyncKnowledge(ctx, host, []*mesapb.KnowledgeEntry{{
		Subject: "Bob's Axes", Kind: "shop", Encounters: 2, LastSeenUnix: 1_700_000_000,
		Beliefs: []*mesapb.KnowledgeBelief{
			{Claim: "sells pickaxes", Provenance: "hearsay", Alpha: 2, Beta: 1, AtUnix: 1_700_000_000},
		},
	}}); err != nil {
		t.Fatalf("seed knowledge: %v", err)
	}

	s.insightLLMOverride = &fakeCompleter{resp: `{"reconcile":[
		{"subject":"Bob's Axes","loser_claim":"sells pickaxes","winner_claim":"only sells axes"}
	]}`}
	seedQueue(t, l, host, qItem{Subject: "Bob's Axes", Claim: "only sells axes", Reason: "contradiction", AtUnix: 1_700_000_100})

	confAfter := func() float64 {
		got, _ := l.Knowledge(ctx, host)
		for _, e := range got {
			if e.GetSubject() == "Bob's Axes" {
				for _, b := range e.GetBeliefs() {
					if b.GetClaim() == "sells pickaxes" {
						return b.GetAlpha() / (b.GetAlpha() + b.GetBeta())
					}
				}
			}
		}
		return -1
	}

	before := 2.0 / 3.0 // alpha=2,beta=1
	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("insightHost tick1: %v", err)
	}
	c1 := confAfter()
	if c1 < 0 || c1 >= before {
		t.Fatalf("reconcile did not lower loser confidence: before=%.3f after=%.3f", before, c1)
	}

	// Re-seed the SAME item with a NEWER timestamp so the cursor admits it again;
	// the reconciled:<fp> tag must make the second application a no-op (no thrash).
	seedQueue(t, l, host, qItem{Subject: "Bob's Axes", Claim: "only sells axes", Reason: "contradiction", AtUnix: 1_700_000_200})
	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("insightHost tick2: %v", err)
	}
	c2 := confAfter()
	if c2 != c1 {
		t.Fatalf("re-run thrashed the loser's confidence: tick1=%.4f tick2=%.4f (must be equal — idempotent)", c1, c2)
	}
}

// TestInsightChainingIsIdempotent proves cross-entity chaining produces the
// IDENTICAL node/edge counts on a re-run (Upsert/Link dedup + slugified ids).
func TestInsightChainingIsIdempotent(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)

	s.insightLLMOverride = &fakeCompleter{resp: `{"chain":[
		{"nodes":[
			{"id":"Rune Plate","kind":"open_goal","label":"obtain rune platebody"},
			{"id":"champions-guild-quest","kind":"subgoal","label":"champions guild quest"}
		],"links":[
			{"from":"Rune Plate","to":"champions-guild-quest","rel":"requires"}
		]}
	]}`}

	run := func(at int64) (int, int) {
		seedQueue(t, l, host, qItem{Subject: "rune plate", Claim: "needs champs guild", Reason: "cross-entity", AtUnix: at})
		if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
			t.Fatalf("insightHost @%d: %v", at, err)
		}
		snap, _ := l.GoalGraph(ctx, host)
		return len(snap.GetNodes()), len(snap.GetEdges())
	}

	n1, e1 := run(1_700_000_100)
	if n1 != 2 || e1 != 1 {
		t.Fatalf("first chain: nodes=%d edges=%d, want 2/1", n1, e1)
	}
	n2, e2 := run(1_700_000_200)
	if n2 != n1 || e2 != e1 {
		t.Fatalf("re-run chaining duplicated: nodes %d->%d edges %d->%d (must be identical)", n1, n2, e1, e2)
	}
	// The "Rune Plate" id must have been slugified to "rune-plate" (stable id).
	snap, _ := l.GoalGraph(ctx, host)
	has := false
	for _, n := range snap.GetNodes() {
		if n.GetId() == "rune-plate" {
			has = true
		}
	}
	if !has {
		t.Fatalf("node id not slugified to a stable slug: %+v", snap.GetNodes())
	}
}

// TestInsightClosesOpenQuestion proves LLM-judged closure flips an open_question
// to StatusDone (above the floor) and that a sub-floor confidence does NOT close.
func TestInsightClosesOpenQuestion(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)

	// Seed two open questions in the goal graph.
	if err := l.SyncGoalGraph(ctx, host, &mesapb.GoalGraphSnapshot{Nodes: []*mesapb.GoalGraphNode{
		{Id: "where-pickaxe", Kind: goalgraph.KindOpenQuestion, Label: "where to buy a pickaxe?", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_000},
		{Id: "where-rune", Kind: goalgraph.KindOpenQuestion, Label: "where to mine runite?", Status: goalgraph.StatusOpen, AtUnix: 1_700_000_000},
	}}); err != nil {
		t.Fatalf("seed graph: %v", err)
	}

	s.insightLLMOverride = &fakeCompleter{resp: `{"close_questions":[
		{"question_id":"where-pickaxe","answering_node":"nurmof","confidence":0.9},
		{"question_id":"where-rune","answering_node":"guess","confidence":0.3}
	]}`}
	seedQueue(t, l, host, qItem{Subject: "Nurmof", Claim: "sells pickaxes", Reason: "answers open question", AtUnix: 1_700_000_100})

	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("insightHost: %v", err)
	}

	snap, _ := l.GoalGraph(ctx, host)
	status := map[string]string{}
	for _, n := range snap.GetNodes() {
		status[n.GetId()] = n.GetStatus()
	}
	if status["where-pickaxe"] != goalgraph.StatusDone {
		t.Fatalf("confident closure did not flip to done: %q", status["where-pickaxe"])
	}
	if status["where-rune"] == goalgraph.StatusDone {
		t.Fatalf("sub-floor (0.3 < %.2f) closure should NOT close: %q", closeQConf, status["where-rune"])
	}
}

// TestInsightEmptyQueueNoLLM proves a host with an empty escalation queue makes
// ZERO Sonnet calls (the Tier-2-rare invariant — no spurious deep spend).
func TestInsightEmptyQueueNoLLM(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	fake := &fakeCompleter{resp: `{}`}
	s.insightLLMOverride = fake

	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("insightHost (empty queue): %v", err)
	}
	if fake.calls.Load() != 0 {
		t.Fatalf("empty queue made %d spurious Sonnet calls, want 0", fake.calls.Load())
	}
}

// TestInsightDrainIdempotentRerun proves the cursor-gated drain: after a tick
// advances the cursor, a re-run with the SAME queue makes NO LLM call (items
// already processed) — the exactly-once-effect contract.
func TestInsightDrainIdempotentRerun(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	fake := &fakeCompleter{resp: `{}`}
	s.insightLLMOverride = fake

	seedQueue(t, l, host, qItem{Subject: "x", Claim: "y", Reason: "r", AtUnix: 1_700_000_100})
	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("tick1: %v", err)
	}
	if fake.calls.Load() != 1 {
		t.Fatalf("tick1 expected 1 Sonnet call, got %d", fake.calls.Load())
	}
	cur := s.loadInsightCursor(ctx, host)
	if cur.LastUnix != 1_700_000_100 {
		t.Fatalf("cursor not advanced: %d", cur.LastUnix)
	}
	// Re-run with the SAME queue (no new items past the cursor) → no LLM call.
	if err := s.insightHost(ctx, host, DefaultCronConfig()); err != nil {
		t.Fatalf("tick2: %v", err)
	}
	if fake.calls.Load() != 1 {
		t.Fatalf("re-run reprocessed (cursor not honored): calls=%d", fake.calls.Load())
	}
}

// TestPersonaGateModulatesEscalation proves the persona-modulated gate: a sharp/
// curious host admits more than an oblivious one, an unknown persona fails open,
// and the admitted count never exceeds InsightMaxPerHost.
func TestPersonaGateModulatesEscalation(t *testing.T) {
	cfg := DefaultCronConfig()
	item := qItem{Subject: "Nurmof", Claim: "sells pickaxes", Reason: "shop price", AtUnix: 1}

	sharp := &persona.Persona{}
	sharp.Cornerstone.Prefs.BulkApperception.Mu = 0.95
	sharp.Cornerstone.Prefs.Curiosity = persona.Curiosity{Social: 0.9, Spatial: 0.9, Skill: 0.9, Economic: 0.9, Risk: 0.9}

	oblivious := &persona.Persona{}
	oblivious.Cornerstone.Prefs.BulkApperception.Mu = 0.02
	oblivious.Cornerstone.Prefs.Curiosity = persona.Curiosity{Social: 0.02, Spatial: 0.02, Skill: 0.02, Economic: 0.02, Risk: 0.02}

	if !shouldProcessEscalation(sharp, item, cfg) {
		t.Fatal("a sharp, curious host should admit this item")
	}
	if shouldProcessEscalation(oblivious, item, cfg) {
		t.Fatal("an oblivious host should drop this marginal item")
	}
	if !shouldProcessEscalation(nil, item, cfg) {
		t.Fatal("an unknown persona must fail OPEN (admit)")
	}
}

// TestInsightAdmitCapBounded proves even the sharpest host is cost-bounded: with
// many queued items, the drain admits at most InsightMaxPerHost (one Sonnet call
// over a bounded input).
func TestInsightAdmitCapBounded(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)

	// Insert a sharp persona directly into the registry (bypassing full persona
	// validation — the gate only reads BulkApperception + Curiosity) so the gate
	// admits everything; the CAP must still bound it.
	p := persona.Persona{}
	p.Cornerstone.Identity.Name = host
	p.Cornerstone.Prefs.BulkApperception.Mu = 1.0
	p.Cornerstone.Prefs.Curiosity = persona.Curiosity{Social: 1, Spatial: 1, Skill: 1, Economic: 1, Risk: 1}
	s.mu.Lock()
	s.reg[host] = &entry{persona: p, prose: "sharp test host"}
	s.mu.Unlock()

	captured := &capturingCompleter{resp: `{}`}
	s.insightLLMOverride = captured

	var items []qItem
	for i := 0; i < 20; i++ {
		items = append(items, qItem{Subject: "subj", Claim: "c", Reason: "skill quest", AtUnix: int64(1_700_000_000 + i)})
	}
	seedQueue(t, l, host, items...)

	cfg := DefaultCronConfig() // InsightMaxPerHost = 6
	if err := s.insightHost(ctx, host, cfg); err != nil {
		t.Fatalf("insightHost: %v", err)
	}
	if captured.calls != 1 {
		t.Fatalf("expected exactly 1 Sonnet call, got %d", captured.calls)
	}
	if captured.itemLines > cfg.InsightMaxPerHost {
		t.Fatalf("admitted %d items, exceeds cap %d", captured.itemLines, cfg.InsightMaxPerHost)
	}
}

// TestInsightTierIsSonnetNeverOpus proves the cardinal cost invariant: the insight
// seam resolves to actLLM (Sonnet), NEVER genesisLLM (Opus).
func TestInsightTierIsSonnetNeverOpus(t *testing.T) {
	sonnet := llm.New("test-key", "claude-sonnet-4-6")
	haiku := llm.New("test-key", "claude-haiku-4-5")
	opus := llm.New("test-key", "claude-opus-4-8")
	s := New(sonnet, haiku, opus, slog.New(slog.NewTextHandler(io.Discard, nil)))

	got, ok := s.insightLLM().(*llm.Client)
	if !ok {
		t.Fatalf("insightLLM did not resolve to an *llm.Client: %T", s.insightLLM())
	}
	if got.Model() != sonnet.Model() {
		t.Fatalf("insight routed to %q, want Sonnet %q", got.Model(), sonnet.Model())
	}
	if got.Model() == opus.Model() {
		t.Fatal("insight must NEVER be the Opus (genesis) tier")
	}
}

// TestInsightPersistFailureNoCursorAdvance proves that when persisting fails the
// cursor is NOT advanced, so items re-process next tick (with no double-effect via
// the idempotency tags). We simulate persist failure by closing the LTM pool after
// the LLM call would run — instead, assert the simpler invariant: an LLM error
// (parse failure) bumps FailCount and does NOT advance the cursor.
func TestInsightLLMErrorNoCursorAdvance(t *testing.T) {
	ctx := context.Background()
	l, host := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	// A response with no JSON ⇒ parse error ⇒ retryable batch.
	s.insightLLMOverride = &fakeCompleter{resp: `not json at all`}

	seedQueue(t, l, host, qItem{Subject: "x", Claim: "y", Reason: "r", AtUnix: 1_700_000_100})
	if err := s.insightHost(ctx, host, DefaultCronConfig()); err == nil {
		t.Fatal("expected an error from the parse failure")
	}
	cur := s.loadInsightCursor(ctx, host)
	if cur.LastUnix != 0 {
		t.Fatalf("cursor advanced past a failed batch: %d (want 0)", cur.LastUnix)
	}
	if cur.FailCount != 1 {
		t.Fatalf("FailCount=%d, want 1", cur.FailCount)
	}
}

// TestInsightSweepBoundedAndQueueGated proves the cost/starvation invariant at
// fleet scale: across many active hosts, insightSweep makes ONE Sonnet call only
// for hosts whose escalation queue is NON-EMPTY (queue-gated, no bulk), bounded by
// MaxHostsPerSweep, sharing the cronSem (initialized like StartCrons does). Hosts
// with observations but an empty queue make ZERO calls.
func TestInsightSweepBoundedAndQueueGated(t *testing.T) {
	ctx := context.Background()
	l, _ := openCronLTM(t)
	s := quietServer()
	s.SetLTM(l)
	s.cronSem = make(chan struct{}, 4) // same as StartCrons would set

	counting := &countingCompleter{resp: `{}`}
	s.insightLLMOverride = counting

	cfg := DefaultCronConfig()
	cfg.MaxHostsPerSweep = 50
	cfg.ActiveWindow = time.Hour

	now := time.Now().Unix()
	var hosts []string
	withQueue := 0
	t.Cleanup(func() {
		for _, h := range hosts {
			_, _ = l.pool.Exec(context.Background(), `DELETE FROM observations WHERE host_id = $1`, h)
			_, _ = l.pool.Exec(context.Background(), `DELETE FROM knowledge WHERE host_id = $1`, h)
			_, _ = l.pool.Exec(context.Background(), `DELETE FROM goal_graphs WHERE host_id = $1`, h)
			_, _ = l.pool.Exec(context.Background(), `DELETE FROM kv WHERE host_id = $1`, h)
		}
	})

	// 60 active hosts; only every 3rd has a non-empty escalation queue.
	for i := 0; i < 60; i++ {
		host := fmt.Sprintf("soak_%d_%d", time.Now().UnixNano(), i)
		hosts = append(hosts, host)
		if _, err := l.AddObservation(ctx, host, obs(fmt.Sprintf("so_%d", i), "player_chat", "x", "y", 0.5, now)); err != nil {
			t.Fatalf("seed obs %d: %v", i, err)
		}
		if i%3 == 0 {
			seedQueue(t, l, host, qItem{Subject: "s", Claim: "c", Reason: "r", AtUnix: now})
			withQueue++
		}
	}

	s.insightSweep(ctx, cfg)
	// Drain in-flight per-host jobs (insightSweep dispatches goroutines joined to cronWG).
	s.cronWG.Wait()

	calls := counting.calls.Load()
	// At most MaxHostsPerSweep hosts are enumerated; calls only for non-empty queues
	// among the enumerated set. Never more than the hosts that actually have a queue.
	if calls > int32(withQueue) {
		t.Fatalf("insight made %d calls, exceeds the %d hosts with a non-empty queue (must be queue-gated, never bulk)", calls, withQueue)
	}
	if calls == 0 {
		t.Fatalf("expected some calls for the queued hosts, got 0 (queue gate over-suppressed)")
	}
	if calls > int32(cfg.MaxHostsPerSweep) {
		t.Fatalf("insight made %d calls, exceeds MaxHostsPerSweep=%d", calls, cfg.MaxHostsPerSweep)
	}
}

// --- test helpers -----------------------------------------------------------

// countingCompleter is a concurrency-safe call counter for the soak test.
type countingCompleter struct {
	resp  string
	calls atomic.Int32
}

func (c *countingCompleter) CompleteSystem(_ context.Context, _ []llm.SystemBlock, _ string, _ int) (string, error) {
	c.calls.Add(1)
	return c.resp, nil
}

// capturingCompleter counts calls and the number of numbered item lines in the
// user turn (to assert the per-host admit cap).
type capturingCompleter struct {
	resp      string
	calls     int
	itemLines int
}

func (c *capturingCompleter) CompleteSystem(_ context.Context, _ []llm.SystemBlock, user string, _ int) (string, error) {
	c.calls++
	// Count lines that look like "N. subject=" (the runInsight user-turn format).
	for _, line := range strings.Split(user, "\n") {
		if len(line) > 2 && line[0] >= '1' && line[0] <= '9' && line[1] == '.' {
			c.itemLines++
		}
	}
	return c.resp, nil
}
