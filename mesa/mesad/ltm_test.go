package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/gen0cide/westworld/mesa/embed"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// testDSN points at a local Postgres. Override with TEST_DATABASE_URL. The test
// skips (not fails) when no database is reachable, so CI without Postgres is fine.
func testDSN() string {
	if v := os.Getenv("TEST_DATABASE_URL"); v != "" {
		return v
	}
	return "postgres://localhost:5432/westworld?sslmode=disable"
}

func openTestLTM(t *testing.T) (*LTM, string) {
	return openTestLTMWith(t, nil) // nil embedder → lexical/recency path
}

func openTestLTMWith(t *testing.T, embedder Embedder) (*LTM, string) {
	t.Helper()
	l, err := OpenLTM(context.Background(), testDSN(), embedder)
	if err != nil {
		t.Skipf("postgres not reachable (%v); skipping LTM integration test", err)
	}
	// A unique host_id keeps the test isolated from any real data, and lets us
	// delete exactly our rows on cleanup.
	host := fmt.Sprintf("test_ltm_%d", time.Now().UnixNano())
	t.Cleanup(func() {
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM episodes WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM relationships WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM personas WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM kv WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM goals WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM metrics WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM knowledge WHERE host_id = $1`, host)
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM goal_graphs WHERE host_id = $1`, host)
		l.Close()
	})
	return l, host
}

func ep(key, kind, text string, importance float64, at int64) *mesapb.Episode {
	return &mesapb.Episode{
		IdempotencyKey: key, Kind: kind, Text: text,
		Importance: importance, OccurredAtUnix: at,
	}
}

func TestLTMAddDedup(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	dup, err := l.Add(ctx, host, ep("k1", "kill", "Defeated a giant rat.", 0.5, 1_700_000_100))
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if dup {
		t.Fatal("first Add reported duplicate")
	}
	// Same idempotency key → deduped (at-least-once resend).
	dup, err = l.Add(ctx, host, ep("k1", "kill", "Defeated a giant rat.", 0.5, 1_700_000_100))
	if err != nil {
		t.Fatalf("Add(resend): %v", err)
	}
	if !dup {
		t.Fatal("resend not deduped")
	}
	if n, _ := l.Count(ctx, host); n != 1 {
		t.Fatalf("Count = %d, want 1", n)
	}
}

func obsv(key, kind, subject, text string, sal float64, at int64) *mesapb.Observation {
	return &mesapb.Observation{
		IdempotencyKey: key, Kind: kind, Subject: subject, Text: text,
		Salience: sal, OccurredAtUnix: at,
	}
}

// TestAddObservationServerFallbackKey is the M4 regression: when the client did NOT
// supply an idempotency key, the server-derived fallback key for a CAPTURE kind must
// carry the text hash, so two DISTINCT same-second server lines (subject="server")
// BOTH persist (neither is swallowed by ON CONFLICT DO NOTHING). A true same-text
// resend still dedups, and a coarse once-per-event kind (outcome) still collapses
// distinct same-second text. The fallback key MUST equal the host emitter's key.
func TestAddObservationServerFallbackKey(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)
	t.Cleanup(func() {
		_, _ = l.pool.Exec(context.Background(), `DELETE FROM observations WHERE host_id = $1`, host)
	})

	const sec = int64(1_700_000_500)

	// Two DISTINCT same-second server lines, NO client key → server fallback path.
	dup, err := l.AddObservation(ctx, host, obsv("", "server_msg", "server", "Welcome to the tutorial.", 0.55, sec))
	if err != nil {
		t.Fatalf("AddObservation #1: %v", err)
	}
	if dup {
		t.Fatal("first same-second server line reported duplicate")
	}
	dup, err = l.AddObservation(ctx, host, obsv("", "server_msg", "server", "Talk to the guide to continue.", 0.55, sec))
	if err != nil {
		t.Fatalf("AddObservation #2: %v", err)
	}
	if dup {
		t.Fatal("second DISTINCT same-second server line collapsed (M4): should persist")
	}

	// A true same-text resend in the same second IS a duplicate (idempotent).
	dup, err = l.AddObservation(ctx, host, obsv("", "server_msg", "server", "Welcome to the tutorial.", 0.55, sec))
	if err != nil {
		t.Fatalf("AddObservation (resend): %v", err)
	}
	if !dup {
		t.Fatal("identical same-second server line should dedup")
	}

	// Both DISTINCT lines persisted → exactly 2 rows for this second.
	var n int
	if err := l.pool.QueryRow(ctx,
		`SELECT count(*) FROM observations WHERE host_id = $1 AND occurred_at = to_timestamp($2)`,
		host, sec).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 2 {
		t.Fatalf("distinct same-second server lines persisted = %d, want 2", n)
	}

	// A coarse once-per-event kind (outcome) keeps the kind|subject|second key, so
	// distinct same-second text DOES collapse (correct dedup of a re-perception).
	if _, err := l.AddObservation(ctx, host, obsv("", "outcome", "Goblin", "defeated Goblin", 0.7, sec)); err != nil {
		t.Fatalf("AddObservation (outcome #1): %v", err)
	}
	dup, err = l.AddObservation(ctx, host, obsv("", "outcome", "Goblin", "defeated Goblin (again)", 0.7, sec))
	if err != nil {
		t.Fatalf("AddObservation (outcome #2): %v", err)
	}
	if !dup {
		t.Fatal("coarse outcome kind should collapse distinct same-second text")
	}
}

func TestLTMRecallKeywordAndRecency(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	_, _ = l.Add(ctx, host, ep("a", "kill", "Defeated a giant rat.", 0.5, 1_700_000_100))
	_, _ = l.Add(ctx, host, ep("b", "milestone", "Reached level 5 in mining.", 0.7, 1_700_000_200))
	_, _ = l.Add(ctx, host, ep("c", "death", "Died and respawned.", 0.9, 1_700_000_300))
	// A "kv" mirror episode must never surface in recall.
	_, _ = l.Add(ctx, host, ep("kv:x", "kv", "{opaque blob}", 0, 1_700_000_400))

	// Keyword recall: "mining" must rank the mining milestone first.
	hits, err := l.Recall(ctx, host, "mining", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 || hits[0].Text != "Reached level 5 in mining." {
		t.Fatalf("keyword recall top = %+v", hits)
	}
	for _, h := range hits {
		if h.Kind == "kv" {
			t.Fatal("kv episode leaked into recall")
		}
	}

	// Empty query → pure recency (newest first = the death at t=300).
	rec, err := l.Recall(ctx, host, "", 5)
	if err != nil {
		t.Fatalf("Recall(recency): %v", err)
	}
	if len(rec) != 3 {
		t.Fatalf("recency recall returned %d (kv should be excluded), want 3", len(rec))
	}
	if rec[0].Text != "Died and respawned." {
		t.Fatalf("recency top = %q, want the newest", rec[0].Text)
	}
}

// TestRecallCarriesImportanceAndEntity proves the Knowledge.Recall handler
// round-trips each stored episode's real importance weight AND entity attribution
// into the KnowledgeItem (regression: the handler used to ship only
// Kind/Text/Provenance/Score, so a cold-start bootstrap lost the salience weight
// — corrupting Salient() ordering — and the entity). Exercises the actual server
// handler end-to-end (auth context → LTM → KnowledgeItem).
func TestRecallCarriesImportanceAndEntity(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	s := New(nil, nil, nil, nil) // nil log → slog.Default; no LLM needed for Recall
	s.SetLTM(l)

	// A kill episode with a high salience weight + an entity attribution.
	kill := ep("k1", "kill", "Defeated a giant rat.", 0.85, 1_700_000_100)
	kill.Tags = map[string]string{"entity": "Giant rat"}
	if _, err := l.Add(ctx, host, kill); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Call the handler as the authenticated host (empty query → recency).
	hctx := context.WithValue(ctx, ctxHostIDKey{}, host)
	set, err := s.Recall(hctx, &mesapb.Query{Text: "", TopK: 5})
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(set.GetItems()) != 1 {
		t.Fatalf("recall returned %d items, want 1", len(set.GetItems()))
	}
	it := set.GetItems()[0]
	if it.GetImportance() != 0.85 {
		t.Fatalf("importance=%v, want 0.85 (not dropped/hardcoded)", it.GetImportance())
	}
	if it.GetEntity() != "Giant rat" {
		t.Fatalf("entity=%q, want %q", it.GetEntity(), "Giant rat")
	}
	if it.GetText() != "Defeated a giant rat." {
		t.Fatalf("text=%q", it.GetText())
	}
}

// TestStoreRelationships proves the trust-ledger seam: a pushed snapshot is
// stored, fetched back faithfully, and a re-push REPLACES (snapshot semantics,
// AuthLocal — not additive).
func TestStoreRelationships(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	n, err := l.SyncRelationships(ctx, host, []*mesapb.Relationship{
		{Name: "alex", Alpha: 5, Beta: 1, Encounters: 4, Tags: []string{"generous"}, ValueTraded: 120},
		{Name: "rival", Alpha: 1, Beta: 3, Encounters: 2},
	})
	if err != nil || n != 2 {
		t.Fatalf("SyncRelationships n=%d err=%v", n, err)
	}
	got, err := l.Relationships(ctx, host)
	if err != nil {
		t.Fatalf("Relationships: %v", err)
	}
	if len(got) != 2 || got[0].GetName() != "alex" { // ordered by encounters desc
		t.Fatalf("fetched = %+v", got)
	}
	if got[0].GetAlpha() != 5 || got[0].GetEncounters() != 4 || len(got[0].GetTags()) != 1 {
		t.Fatalf("alex not faithful: %+v", got[0])
	}

	// Re-push with new absolute values → REPLACE, not add.
	if _, err := l.SyncRelationships(ctx, host, []*mesapb.Relationship{
		{Name: "alex", Alpha: 6, Beta: 1, Encounters: 5},
	}); err != nil {
		t.Fatalf("re-sync: %v", err)
	}
	got, _ = l.Relationships(ctx, host)
	for _, r := range got {
		if r.GetName() == "alex" && r.GetAlpha() != 6 {
			t.Fatalf("snapshot did not replace: alpha=%v, want 6", r.GetAlpha())
		}
	}
}

// TestAddFoldsRelationValueTraded is the latent-trap close: an episode's
// RelationDelta.TotalValueTraded must be FOLDED into the (host_id,name)
// relationships row's value_traded instead of being silently discarded (LTM.Add
// previously never read ep.Relation). A fresh trade INSERTs the row; a second
// completed-trade episode ADDS to it (monotone total); a dedup resend of the SAME
// idempotency key does NOT double-count.
func TestAddFoldsRelationValueTraded(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	epRel := func(key string, at int64, name string, value float64) *mesapb.Episode {
		e := ep(key, "trade", "Traded with "+name, 0.5, at)
		e.Relation = &mesapb.RelationDelta{Name: name, TotalValueTraded: value}
		return e
	}

	// First completed-trade episode creates the relationship row with the volume.
	if dup, err := l.Add(ctx, host, epRel("t1", 1_700_000_100, "merchant", 120)); err != nil || dup {
		t.Fatalf("Add(t1) dup=%v err=%v", dup, err)
	}
	if v := relValueTraded(t, l, host, "merchant"); v != 120 {
		t.Fatalf("after t1 value_traded=%v, want 120", v)
	}

	// A second, DISTINCT completed-trade episode folds in additively (monotone total).
	if dup, err := l.Add(ctx, host, epRel("t2", 1_700_000_200, "merchant", 80)); err != nil || dup {
		t.Fatalf("Add(t2) dup=%v err=%v", dup, err)
	}
	if v := relValueTraded(t, l, host, "merchant"); v != 200 {
		t.Fatalf("after t2 value_traded=%v, want 200 (120+80)", v)
	}

	// An at-least-once RESEND of t2 dedups → no double-count of the same trade.
	if dup, err := l.Add(ctx, host, epRel("t2", 1_700_000_200, "merchant", 80)); err != nil || !dup {
		t.Fatalf("Add(t2 resend) dup=%v err=%v, want dup=true", dup, err)
	}
	if v := relValueTraded(t, l, host, "merchant"); v != 200 {
		t.Fatalf("resend double-counted value_traded=%v, want 200", v)
	}

	// An episode with no RelationDelta touches no relationship row.
	if _, err := l.Add(ctx, host, ep("plain", "kill", "Defeated a rat.", 0.5, 1_700_000_300)); err != nil {
		t.Fatalf("Add(plain): %v", err)
	}
	if v := relValueTraded(t, l, host, "merchant"); v != 200 {
		t.Fatalf("a non-relational episode perturbed value_traded=%v, want 200", v)
	}
}

// relValueTraded reads value_traded directly off the relationships table for a
// (host,name) — a query, since Relationships() returns the whole ledger.
func relValueTraded(t *testing.T, l *LTM, host, name string) float64 {
	t.Helper()
	var v float64
	err := l.pool.QueryRow(context.Background(),
		`SELECT value_traded FROM relationships WHERE host_id = $1 AND name = $2`, host, name).Scan(&v)
	if err != nil {
		t.Fatalf("read value_traded: %v", err)
	}
	return v
}

// TestStoreMetrics proves the metrics time-series seam: a batch is appended and
// queryable by name.
func TestStoreMetrics(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	n, err := l.RecordMetrics(ctx, host, []*mesapb.Metric{
		{Name: "pearl.vetoes", Value: 3},
		{Name: "agent.turns", Value: 10},
	}, time.Unix(1_700_000_000, 0))
	if err != nil || n != 2 {
		t.Fatalf("RecordMetrics n=%d err=%v", n, err)
	}
	var cnt int
	_ = l.pool.QueryRow(ctx, `SELECT count(*) FROM metrics WHERE host_id = $1`, host).Scan(&cnt)
	if cnt != 2 {
		t.Fatalf("rows = %d, want 2", cnt)
	}
	var v float64
	_ = l.pool.QueryRow(ctx,
		`SELECT value FROM metrics WHERE host_id = $1 AND name = 'pearl.vetoes'`, host).Scan(&v)
	if v != 3 {
		t.Fatalf("pearl.vetoes value = %v, want 3", v)
	}
}

// TestStoreGoal proves the structured goals seam: objective + progress persist,
// read back, and upsert replaces.
func TestStoreGoal(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	if _, _, found, err := l.Goal(ctx, host); err != nil || found {
		t.Fatalf("expected no goal: found=%v err=%v", found, err)
	}
	if err := l.UpsertGoal(ctx, host, "become a master chef", []string{"Reached level 5 in cooking."}); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	obj, prog, found, err := l.Goal(ctx, host)
	if err != nil || !found || obj != "become a master chef" || len(prog) != 1 || prog[0] != "Reached level 5 in cooking." {
		t.Fatalf("goal = %q %v found=%v err=%v", obj, prog, found, err)
	}
	if err := l.UpsertGoal(ctx, host, "explore lumbridge", nil); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if obj, _, _, _ := l.Goal(ctx, host); obj != "explore lumbridge" {
		t.Fatalf("replace failed: %q", obj)
	}
}

// TestStoreKV proves the generic KV seam: miss, put, get, replace, delete.
func TestStoreKV(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	if _, found, err := l.GetKV(ctx, host, "goal:objective"); err != nil || found {
		t.Fatalf("expected miss: found=%v err=%v", found, err)
	}
	if err := l.PutKV(ctx, host, "goal:objective", []byte("become free")); err != nil {
		t.Fatalf("put: %v", err)
	}
	v, found, err := l.GetKV(ctx, host, "goal:objective")
	if err != nil || !found || string(v) != "become free" {
		t.Fatalf("get = %q found=%v err=%v", v, found, err)
	}
	if err := l.PutKV(ctx, host, "goal:objective", []byte("become a chef")); err != nil {
		t.Fatalf("replace: %v", err)
	}
	v, _, _ = l.GetKV(ctx, host, "goal:objective")
	if string(v) != "become a chef" {
		t.Fatalf("replace failed: %q", v)
	}
	if err := l.DeleteKV(ctx, host, "goal:objective"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, found, _ := l.GetKV(ctx, host, "goal:objective"); found {
		t.Fatal("delete did not remove the key")
	}
}

// TestStorePersona proves the persona seam: a persona JSON is persisted and read
// back, and an upsert replaces it.
func TestStorePersona(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	if err := l.UpsertPersona(ctx, host, []byte(`{"name":"first","v":1}`), "", false); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	personaFor := func() map[string]any {
		rows, err := l.Personas(ctx)
		if err != nil {
			t.Fatalf("personas: %v", err)
		}
		for _, r := range rows {
			if r.HostID == host {
				var m map[string]any
				if err := json.Unmarshal(r.JSON, &m); err != nil {
					t.Fatalf("bad stored json: %v", err)
				}
				return m
			}
		}
		t.Fatal("persona not stored for host")
		return nil
	}
	if personaFor()["name"] != "first" {
		t.Fatal("persona not stored faithfully")
	}
	// Upsert replaces.
	if err := l.UpsertPersona(ctx, host, []byte(`{"name":"second"}`), "card", true); err != nil {
		t.Fatalf("re-upsert: %v", err)
	}
	if personaFor()["name"] != "second" {
		t.Fatal("upsert did not replace persona")
	}
}

func kb(claim, prov string, alpha, beta float64, at int64) *mesapb.KnowledgeBelief {
	return &mesapb.KnowledgeBelief{Claim: claim, Provenance: prov, Alpha: alpha, Beta: beta, AtUnix: at}
}

// TestMergeBeliefs is a pure (DB-free) unit test of the per-(claim,provenance)
// max-evidence fold SyncKnowledge uses (M17), and that nil incoming normalizes to
// a non-null `[]` (L14b).
func TestMergeBeliefs(t *testing.T) {
	// nil/empty incoming with no stored → empty (non-nil) slice → JSON "[]".
	if got := mergeBeliefs(nil, nil); got == nil || len(got) != 0 {
		t.Fatalf("nil/nil merge = %v, want non-nil empty", got)
	}
	if bj, _ := json.Marshal(mergeBeliefs(nil, nil)); string(bj) != "[]" {
		t.Fatalf("nil beliefs marshal = %s, want []", bj)
	}

	stored, _ := json.Marshal([]*mesapb.KnowledgeBelief{
		kb("sells pickaxe", "observed", 5, 1, 100), // cron-reconciled, strong
		kb("only stored", "hearsay", 2, 0, 50),     // host omits this one
	})
	// Host pushes a WEAKER copy of the shared claim + a brand-new claim.
	incoming := []*mesapb.KnowledgeBelief{
		kb("sells pickaxe", "observed", 2, 1, 80), // weaker α + older At
		kb("new claim", "deduced", 1, 1, 200),
	}
	got := mergeBeliefs(stored, incoming)
	by := map[string]*mesapb.KnowledgeBelief{}
	for _, b := range got {
		by[b.GetClaim()] = b
	}
	if len(got) != 3 {
		t.Fatalf("merge produced %d beliefs, want 3 (union): %+v", len(got), got)
	}
	// Shared claim keeps MAX evidence + NEWER At — the weak host push must not win.
	if b := by["sells pickaxe"]; b == nil || b.GetAlpha() != 5 || b.GetBeta() != 1 || b.GetAtUnix() != 100 {
		t.Fatalf("shared claim not max-merged: %+v", b)
	}
	// Stored-only belief survives a push that omits it.
	if by["only stored"] == nil {
		t.Fatal("stored-only belief was dropped by merge")
	}
	if by["new claim"] == nil {
		t.Fatal("incoming-only belief missing from merge")
	}

	// A claim the host has gained MORE evidence for must take the host's stronger α.
	got2 := mergeBeliefs(stored, []*mesapb.KnowledgeBelief{kb("sells pickaxe", "observed", 9, 3, 300)})
	for _, b := range got2 {
		if b.GetClaim() == "sells pickaxe" {
			if b.GetAlpha() != 9 || b.GetBeta() != 3 || b.GetAtUnix() != 300 {
				t.Fatalf("stronger host evidence not adopted: %+v", b)
			}
		}
	}
}

// TestSyncKnowledgeCountAndNullBeliefs proves L14: blank-subject entries are
// skipped AND not counted, and a belief-less entry stores JSON "[]" (not "null").
func TestSyncKnowledgeCountAndNullBeliefs(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	n, err := l.SyncKnowledge(ctx, host, []*mesapb.KnowledgeEntry{
		{Subject: "Bob's Axes", Kind: "shop", Beliefs: []*mesapb.KnowledgeBelief{kb("sells axes", "observed", 3, 1, 100)}},
		{Subject: "   ", Kind: "junk"},           // blank → skipped, not counted
		{Subject: "Lumbridge", Kind: "location"}, // nil beliefs → must store "[]"
	})
	if err != nil {
		t.Fatalf("SyncKnowledge: %v", err)
	}
	if n != 2 {
		t.Fatalf("written count = %d, want 2 (blank subject excluded)", n)
	}
	// The belief-less entry must be a JSON array, never `null`.
	var raw string
	if err := l.pool.QueryRow(ctx,
		`SELECT beliefs_json::text FROM knowledge WHERE host_id = $1 AND subject = 'Lumbridge'`,
		host).Scan(&raw); err != nil {
		t.Fatalf("read Lumbridge beliefs: %v", err)
	}
	if raw != "[]" {
		t.Fatalf("belief-less entry stored as %q, want []", raw)
	}
	// JSONB array ops must work on it (would error on a `null` row).
	var cnt int
	if err := l.pool.QueryRow(ctx,
		`SELECT jsonb_array_length(beliefs_json) FROM knowledge WHERE host_id = $1 AND subject = 'Lumbridge'`,
		host).Scan(&cnt); err != nil {
		t.Fatalf("jsonb_array_length on belief-less row errored: %v", err)
	}
	if cnt != 0 {
		t.Fatalf("array length = %d, want 0", cnt)
	}
}

// TestSyncKnowledgeMergePreservesCron proves M17 (server side): a warm host push
// of its OWN weaker copy of a subject does NOT clobber the cron's strictly-newer
// reconciled belief; beliefs are folded per (claim,provenance) taking max evidence,
// and a cron belief the host never re-learned survives the host's push.
func TestSyncKnowledgeMergePreservesCron(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	// 1) The cron writes a reconciled, strong belief + a belief only it knows.
	if _, err := l.SyncKnowledge(ctx, host, []*mesapb.KnowledgeEntry{{
		Subject: "Bob's Axes", Kind: "shop", Encounters: 5, LastSeenUnix: 200,
		Beliefs: []*mesapb.KnowledgeBelief{
			kb("sells pickaxe", "observed", 8, 1, 200), // cron β-bumped
			kb("out of bronze", "observed", 1, 4, 190), // cron-only
		},
	}}); err != nil {
		t.Fatalf("cron sync: %v", err)
	}

	// 2) A warm host flushes its OWN local copy (never saw the reconciliation):
	// weaker evidence on the shared claim, omits the cron-only belief, and a
	// stale lower encounters/last_seen.
	if _, err := l.SyncKnowledge(ctx, host, []*mesapb.KnowledgeEntry{{
		Subject: "Bob's Axes", Kind: "shop", Encounters: 2, LastSeenUnix: 150,
		Beliefs: []*mesapb.KnowledgeBelief{
			kb("sells pickaxe", "observed", 3, 1, 150), // weaker, older
		},
	}}); err != nil {
		t.Fatalf("host sync: %v", err)
	}

	got, err := l.Knowledge(ctx, host)
	if err != nil {
		t.Fatalf("Knowledge: %v", err)
	}
	var entry *mesapb.KnowledgeEntry
	for _, e := range got {
		if e.GetSubject() == "Bob's Axes" {
			entry = e
		}
	}
	if entry == nil {
		t.Fatal("subject vanished after host push")
	}
	by := map[string]*mesapb.KnowledgeBelief{}
	for _, b := range entry.GetBeliefs() {
		by[b.GetClaim()] = b
	}
	// Shared claim must keep the cron's stronger α + newer At, not the host's weak one.
	if b := by["sells pickaxe"]; b == nil || b.GetAlpha() != 8 || b.GetAtUnix() != 200 {
		t.Fatalf("host push clobbered cron belief: %+v", b)
	}
	// The cron-only belief must survive the host push that omitted it.
	if by["out of bronze"] == nil {
		t.Fatal("cron-only belief was clobbered by the host's flush")
	}
	// Monotonic metadata: stale host encounters/last_seen must not regress.
	if entry.GetEncounters() != 5 {
		t.Fatalf("encounters regressed to %d, want 5", entry.GetEncounters())
	}
	if entry.GetLastSeenUnix() != 200 {
		t.Fatalf("last_seen regressed to %d, want 200", entry.GetLastSeenUnix())
	}
}

// TestMaxNodeAt is a DB-free unit test of the content-derived graph version used by
// the SyncGoalGraph strictly-newer guard (H17): the freshest node's AtUnix, 0 for an
// empty/nil-padded set.
func TestMaxNodeAt(t *testing.T) {
	if got := maxNodeAt(nil); got != 0 {
		t.Fatalf("maxNodeAt(nil) = %d, want 0", got)
	}
	nodes := []*mesapb.GoalGraphNode{
		{Id: "a", AtUnix: 100},
		nil, // a nil entry must be skipped, not panic
		{Id: "b", AtUnix: 300},
		{Id: "c", AtUnix: 200},
	}
	if got := maxNodeAt(nodes); got != 300 {
		t.Fatalf("maxNodeAt = %d, want 300 (freshest node)", got)
	}
}

// TestSyncGoalGraphRejectsStaleHostPush proves the H17 server-side backstop: once a
// strictly-NEWER snapshot is stored (the cron's reconciled graph), a host push of a
// STRICTLY-OLDER snapshot is REFUSED (cannot clobber the cron's growth), while an
// equal-or-newer push (the normal warm case where the host re-imported the cron's
// nodes) replaces as before.
func TestSyncGoalGraphRejectsStaleHostPush(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	// 1) The cron writes a reconciled graph whose freshest node is at At=300.
	if err := l.SyncGoalGraph(ctx, host, &mesapb.GoalGraphSnapshot{Nodes: []*mesapb.GoalGraphNode{
		{Id: "cron-chain", Kind: "open_goal", Label: "obtain rune platebody", Status: "open", AtUnix: 300},
		{Id: "mine-ore", Kind: "goal", Label: "mine ore", Status: "active", AtUnix: 300},
	}}); err != nil {
		t.Fatalf("cron sync: %v", err)
	}

	// 2) A warm host whose local copy LAGGED the cron (max At=200) tries to push its
	// own graph (missing the cron chain). The strictly-older guard must REFUSE it.
	if err := l.SyncGoalGraph(ctx, host, &mesapb.GoalGraphSnapshot{Nodes: []*mesapb.GoalGraphNode{
		{Id: "mine-ore", Kind: "goal", Label: "mine ore", Status: "active", AtUnix: 200},
	}}); err != nil {
		t.Fatalf("stale host sync should be a no-op error-free: %v", err)
	}
	snap, err := l.GoalGraph(ctx, host)
	if err != nil {
		t.Fatalf("GoalGraph: %v", err)
	}
	ids := map[string]bool{}
	for _, n := range snap.GetNodes() {
		ids[n.GetId()] = true
	}
	if !ids["cron-chain"] {
		t.Fatal("stale host push clobbered the cron's strictly-newer chain (guard failed)")
	}

	// 3) An equal-or-newer push (the host re-imported the cron's nodes, so its max At
	// now >= the stored one) replaces as before.
	if err := l.SyncGoalGraph(ctx, host, &mesapb.GoalGraphSnapshot{Nodes: []*mesapb.GoalGraphNode{
		{Id: "cron-chain", Kind: "open_goal", Label: "obtain rune platebody", Status: "open", AtUnix: 300},
		{Id: "mine-ore", Kind: "goal", Label: "mine ore", Status: "active", AtUnix: 400}, // host progressed it
		{Id: "host-new", Kind: "subgoal", Label: "host added", Status: "active", AtUnix: 400},
	}}); err != nil {
		t.Fatalf("fresh host sync: %v", err)
	}
	snap, _ = l.GoalGraph(ctx, host)
	ids = map[string]bool{}
	for _, n := range snap.GetNodes() {
		ids[n.GetId()] = true
	}
	if !ids["host-new"] || !ids["cron-chain"] {
		t.Fatalf("fresh push did not replace as expected: %v", ids)
	}
}

// TestLTMEmbeddingBackfill proves M12: an episode first stored with a NULL
// embedding (embedder absent) backfills its vector when a later resend (now with
// an embedder) carries one — so it becomes visible to cosine recall instead of
// being invisible forever. Uses a deterministic fake embedder (no Voyage key).
func TestLTMEmbeddingBackfill(t *testing.T) {
	ctx := context.Background()
	// First store WITHOUT an embedder → embedding stays NULL.
	l, host := openTestLTM(t)

	dup, err := l.Add(ctx, host, ep("bf1", "kill", "Defeated a giant rat.", 0.5, 1_700_000_100))
	if err != nil {
		t.Fatalf("Add (no embedder): %v", err)
	}
	if dup {
		t.Fatal("first Add reported duplicate")
	}
	var hasEmb bool
	_ = l.pool.QueryRow(ctx,
		`SELECT embedding IS NOT NULL FROM episodes WHERE host_id = $1 AND idem_key = 'bf1'`,
		host).Scan(&hasEmb)
	if hasEmb {
		t.Fatal("embedding should be NULL when no embedder is wired")
	}

	// Re-open the SAME rows with a fake embedder and resend the SAME idem key.
	l.embedder = fakeEmbedder{dim: l.dim}
	dup, err = l.Add(ctx, host, ep("bf1", "kill", "Defeated a giant rat.", 0.5, 1_700_000_100))
	if err != nil {
		t.Fatalf("Add (backfill): %v", err)
	}
	if !dup {
		t.Fatal("backfill resend should report duplicate (it is an UPDATE, not a new row)")
	}
	_ = l.pool.QueryRow(ctx,
		`SELECT embedding IS NOT NULL FROM episodes WHERE host_id = $1 AND idem_key = 'bf1'`,
		host).Scan(&hasEmb)
	if !hasEmb {
		t.Fatal("embedding was not backfilled on resend with an embedder")
	}
	// Still exactly one row (resend did not duplicate).
	if n, _ := l.Count(ctx, host); n != 1 {
		t.Fatalf("Count = %d, want 1 after backfill", n)
	}
}

// fakeEmbedder returns a fixed unit vector — deterministic, no network.
type fakeEmbedder struct{ dim int }

func (f fakeEmbedder) Dim() int { return f.dim }
func (f fakeEmbedder) Embed(_ context.Context, texts []string, _ string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		v := make([]float32, f.dim)
		if f.dim > 0 {
			v[0] = 1
		}
		out[i] = v
	}
	return out, nil
}

// TestMigrateToleratesMalformedPersona proves L15: a persona row whose JSON
// carries a NON-numeric schema_version / curiosity value does not brick migrate()
// — the generated columns are NULL-tolerant. We seed a malformed row then re-run
// migrate() (idempotent), which must succeed and leave the bad columns NULL.
func TestMigrateToleratesMalformedPersona(t *testing.T) {
	ctx := context.Background()
	l, host := openTestLTM(t)

	// schema_version is a string-not-number; a curiosity dial is "high" not a float.
	bad := `{"schema_version":"v3","cornerstone":{"identity":{"name":"Mal"},` +
		`"prefs":{"curiosity":{"social":"high","spatial":0.7}}}}`
	if err := l.UpsertPersona(ctx, host, []byte(bad), "", false); err != nil {
		t.Fatalf("upsert malformed persona: %v", err)
	}

	// Re-running migrate() evaluates the generated-column expressions over EVERY
	// existing persona row (incl. ours). The hard-cast version would error here.
	if err := l.migrate(ctx); err != nil {
		t.Fatalf("migrate() bricked on a malformed persona: %v", err)
	}

	// The malformed numeric facets resolve to NULL (not an error); a well-formed
	// sibling on the same row still parses.
	var sv *int
	var curSocial, curSpatial *float64
	if err := l.pool.QueryRow(ctx,
		`SELECT schema_version, cur_social, cur_spatial FROM personas WHERE host_id = $1`,
		host).Scan(&sv, &curSocial, &curSpatial); err != nil {
		t.Fatalf("read generated cols: %v", err)
	}
	if sv != nil {
		t.Fatalf("malformed schema_version should be NULL, got %d", *sv)
	}
	if curSocial != nil {
		t.Fatalf("malformed cur_social should be NULL, got %v", *curSocial)
	}
	// cur_spatial is a real (float32) column, so widening to float64 carries the
	// usual representation slop — compare with a tolerance, not for equality.
	if curSpatial == nil || *curSpatial < 0.69 || *curSpatial > 0.71 {
		t.Fatalf("well-formed cur_spatial should parse to ~0.7, got %v", curSpatial)
	}
}

// TestLTMSemanticRecall proves the pgvector cosine path: a query with NO lexical
// overlap with the target episode still ranks it first (full-text search could
// not). Needs a real Voyage key + Postgres; skips otherwise.
func TestLTMSemanticRecall(t *testing.T) {
	key := os.Getenv("VOYAGE_AI_KEY")
	if key == "" {
		t.Skip("no VOYAGE_AI_KEY; skipping semantic recall test")
	}
	ctx := context.Background()
	l, host := openTestLTMWith(t, embed.NewVoyage(key, embed.DefaultModel))

	_, _ = l.Add(ctx, host, ep("s1", "milestone", "Reached level 5 in mining.", 0.7, 1_700_000_100))
	_, _ = l.Add(ctx, host, ep("s2", "kill", "Slew a fearsome giant rat with my sword.", 0.45, 1_700_000_200))
	_, _ = l.Add(ctx, host, ep("s3", "trade", "Sold copper ore to a merchant for coins.", 0.4, 1_700_000_300))

	// "fighting monsters in combat" shares no keywords with "Slew a fearsome
	// giant rat" — only semantics. Cosine must still rank the combat episode top.
	hits, err := l.Recall(ctx, host, "fighting monsters in combat", 3)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(hits) == 0 {
		t.Fatal("no semantic hits")
	}
	if hits[0].Text != "Slew a fearsome giant rat with my sword." {
		t.Fatalf("semantic top = %q (score %.3f), want the combat episode", hits[0].Text, hits[0].Score)
	}
	t.Logf("semantic recall ✓ top=%q score=%.3f", hits[0].Text, hits[0].Score)
}
