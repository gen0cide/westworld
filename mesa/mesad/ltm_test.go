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
