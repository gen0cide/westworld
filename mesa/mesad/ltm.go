package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// Embedder turns text into a vector for semantic recall. mesa/embed.Voyage
// satisfies it. nil disables semantic ranking (recency/lexical fallback).
type Embedder interface {
	Embed(ctx context.Context, texts []string, inputType string) ([][]float32, error)
	Dim() int
}

// LTM is the durable, host-namespaced long-term episodic memory store, backed by
// PostgreSQL (mesad's system of record — bbolt is the HOST's local tier, not
// mesa's). Episodes arrive via Journal.Remember (the host mirrors its local
// journal up) and are retrieved via Knowledge.Recall. Every row is scoped by the
// authenticated host_id, so a host can only ever read its own past.
//
// When an Embedder is wired, episodes are embedded (Voyage) on write and Recall
// ranks by pgvector cosine similarity; without one, Recall degrades to Postgres
// full-text relevance + recency. Either way mesa can bootstrap a fresh host from
// nothing (empty-query Recall returns the host's episodes in recency order).
type LTM struct {
	pool     *pgxpool.Pool
	embedder Embedder
	dim      int
}

// storedEpisode is one recalled memory returned to the handler.
type storedEpisode struct {
	Kind       string
	Text       string
	Importance float64
	Entity     string
	At         int64 // unix seconds
	Score      float64
}

// OpenLTM connects the pool at dsn and runs migrations. embedder may be nil
// (semantic recall then degrades to lexical/recency). The caller owns Close.
func OpenLTM(ctx context.Context, dsn string, embedder Embedder) (*LTM, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	dim := 1024
	if embedder != nil {
		dim = embedder.Dim()
	}
	l := &LTM{pool: pool, embedder: embedder, dim: dim}
	if err := l.migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return l, nil
}

// migrate creates the episode table, the recency index, the pgvector embedding
// column, and a cosine HNSW index. Idempotent.
func (l *LTM) migrate(ctx context.Context) error {
	stmts := []string{
		`CREATE EXTENSION IF NOT EXISTS vector`,
		`CREATE TABLE IF NOT EXISTS episodes (
    host_id     text             NOT NULL,
    idem_key    text             NOT NULL,
    kind        text             NOT NULL,
    body        text             NOT NULL,
    importance  double precision NOT NULL DEFAULT 0,
    entity      text             NOT NULL DEFAULT '',
    occurred_at timestamptz      NOT NULL,
    created_at  timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, idem_key)
)`,
		`CREATE INDEX IF NOT EXISTS episodes_host_time ON episodes (host_id, occurred_at DESC)`,
		fmt.Sprintf(`ALTER TABLE episodes ADD COLUMN IF NOT EXISTS embedding vector(%d)`, l.dim),
		`CREATE INDEX IF NOT EXISTS episodes_embedding ON episodes USING hnsw (embedding vector_cosine_ops)`,
		// relationships: the host's felt trust ledger (AuthLocal). The host pushes
		// its absolute Beta(alpha,beta) state up; mesa mirrors the latest + serves
		// it back for cold-start bootstrap. One row per (host, counterparty).
		`CREATE TABLE IF NOT EXISTS relationships (
    host_id      text             NOT NULL,
    name         text             NOT NULL,
    alpha        double precision NOT NULL DEFAULT 1,
    beta         double precision NOT NULL DEFAULT 1,
    encounters   integer          NOT NULL DEFAULT 0,
    tags         text[]           NOT NULL DEFAULT '{}',
    value_traded double precision NOT NULL DEFAULT 0,
    updated_at   timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, name)
)`,
		// personas: the host's compiled identity (the JSON the host provisions
		// down). mesad's system of record — so a host's persona survives a mesad
		// restart without re-specifying the -host file, and mesa can bootstrap a
		// host's identity from nothing.
		`CREATE TABLE IF NOT EXISTS personas (
    host_id      text        NOT NULL PRIMARY KEY,
    persona_json jsonb       NOT NULL,
    updated_at   timestamptz NOT NULL DEFAULT now()
)`,
		// kv: the generic host-state mirror (memory.Manager's remote tier) — the
		// clean home for arbitrary OPAQUE durable host state. Analytically-relevant
		// state gets its own structured table instead (episodes/relationships/goals).
		`CREATE TABLE IF NOT EXISTS kv (
    host_id    text        NOT NULL,
    key        text        NOT NULL,
    value      bytea       NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, key)
)`,
		// goals: the host's MUTABLE standing objective + progress (vs the immutable
		// persona north-star). Structured + queryable so crons/analytics can ask
		// "who's pursuing X" / "how long on this goal". One row per host.
		`CREATE TABLE IF NOT EXISTS goals (
    host_id    text        NOT NULL PRIMARY KEY,
    objective  text        NOT NULL DEFAULT '',
    progress   jsonb       NOT NULL DEFAULT '[]',
    updated_at timestamptz NOT NULL DEFAULT now()
)`,
		// metrics: host telemetry as an append-only time series (one row per
		// sample) — observability + cron aggregation inputs (rollup/retention is a
		// future cron concern).
		`CREATE TABLE IF NOT EXISTS metrics (
    host_id text             NOT NULL,
    name    text             NOT NULL,
    value   double precision NOT NULL,
    at      timestamptz      NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS metrics_host_name_at ON metrics (host_id, name, at DESC)`,
	}

	// personas projection: surface every identity facet + dial as a first-class,
	// queryable column GENERATED from the canonical persona_json — always in sync,
	// zero app extraction. The JSON stays the host wire-format + source of truth;
	// prose_card + cooked are app-written (prose is derived by Render()/cook, not
	// present in the JSON). All idempotent (ADD COLUMN IF NOT EXISTS).
	stmts = append(stmts,
		`ALTER TABLE personas ADD COLUMN IF NOT EXISTS prose_card text NOT NULL DEFAULT ''`,
		`ALTER TABLE personas ADD COLUMN IF NOT EXISTS cooked boolean NOT NULL DEFAULT false`,
		`ALTER TABLE personas ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now()`,
	)
	type gcol struct{ name, expr, typ string }
	gtext := func(name string, path ...string) gcol {
		return gcol{name, fmt.Sprintf("persona_json #>> '{%s}'", strings.Join(path, ",")), "text"}
	}
	greal := func(name string, path ...string) gcol {
		return gcol{name, fmt.Sprintf("(persona_json #>> '{%s}')::real", strings.Join(path, ",")), "real"}
	}
	for _, g := range []gcol{
		{"schema_version", "(persona_json #>> '{schema_version}')::int", "int"},
		gtext("name", "cornerstone", "identity", "name"),
		gtext("archetype", "cornerstone", "generation_meta", "archetype"),
		gtext("cohort_id", "cornerstone", "generation_meta", "cohort_id"),
		gtext("north_star_theme", "cornerstone", "identity", "north_star", "theme"),
		gtext("north_star", "cornerstone", "identity", "north_star", "statement"),
		gtext("voice_formality", "cornerstone", "identity", "voice", "formality"),
		gtext("voice_typo_feel", "cornerstone", "identity", "voice", "typo_feel"),
		gtext("hexaco_h", "cornerstone", "hexaco", "H", "band"),
		gtext("hexaco_e", "cornerstone", "hexaco", "E", "band"),
		gtext("hexaco_x", "cornerstone", "hexaco", "X", "band"),
		gtext("hexaco_a", "cornerstone", "hexaco", "A", "band"),
		gtext("hexaco_c", "cornerstone", "hexaco", "C", "band"),
		gtext("hexaco_o", "cornerstone", "hexaco", "O", "band"),
		gtext("north_star_value", "cornerstone", "values", "north_star_value"),
		gtext("secondary_value", "cornerstone", "values", "secondary_value"),
		gtext("patience", "cornerstone", "prefs", "patience", "band"),
		gtext("loss_aversion", "cornerstone", "prefs", "loss_aversion", "band"),
		gtext("aggression", "cornerstone", "prefs", "aggression", "band"),
		gtext("decisiveness", "cornerstone", "prefs", "decisiveness", "band"),
		gtext("tenacity", "cornerstone", "prefs", "tenacity", "band"),
		gtext("bulk_apperception", "cornerstone", "prefs", "bulk_apperception", "band"),
		gtext("self_preservation", "cornerstone", "prefs", "self_preservation", "band"),
		gtext("coop_type", "cornerstone", "prefs", "coop_type"),
		gtext("risk_economic", "cornerstone", "prefs", "risk", "economic"),
		gtext("risk_bodily", "cornerstone", "prefs", "risk", "bodily"),
		gtext("risk_social", "cornerstone", "prefs", "risk", "social"),
		gtext("attention_level", "cornerstone", "prefs", "attention", "level"),
		greal("cur_social", "cornerstone", "prefs", "curiosity", "social"),
		greal("cur_spatial", "cornerstone", "prefs", "curiosity", "spatial"),
		greal("cur_skill", "cornerstone", "prefs", "curiosity", "skill"),
		greal("cur_economic", "cornerstone", "prefs", "curiosity", "economic"),
		greal("cur_risk", "cornerstone", "prefs", "curiosity", "risk"),
	} {
		stmts = append(stmts, fmt.Sprintf(
			`ALTER TABLE personas ADD COLUMN IF NOT EXISTS %s %s GENERATED ALWAYS AS (%s) STORED`,
			g.name, g.typ, g.expr))
	}

	for _, s := range stmts {
		if _, err := l.pool.Exec(ctx, s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	return nil
}

// Close releases the connection pool.
func (l *LTM) Close() error {
	if l != nil && l.pool != nil {
		l.pool.Close()
	}
	return nil
}

// Add stores one episode under the host's namespace, deduplicated by its
// idempotency key (ON CONFLICT DO NOTHING). When an embedder is wired the text
// is embedded and stored for semantic recall; an embedding failure is tolerated
// (the episode is still stored, just without a vector — never lose a memory to a
// transient API error). Returns dup=true when the key already existed.
func (l *LTM) Add(ctx context.Context, hostID string, ep *mesapb.Episode) (dup bool, err error) {
	key := ep.GetIdempotencyKey()
	if key == "" {
		key = contentKey(ep)
	}
	at := time.Now()
	if u := ep.GetOccurredAtUnix(); u != 0 {
		at = time.Unix(u, 0)
	}
	var emb *string
	if l.embedder != nil && ep.GetKind() != "kv" {
		if vecs, eerr := l.embedder.Embed(ctx, []string{ep.GetText()}, "document"); eerr == nil && len(vecs) == 1 {
			s := vecLiteral(vecs[0])
			emb = &s
		}
	}
	tag, err := l.pool.Exec(ctx, `
INSERT INTO episodes (host_id, idem_key, kind, body, importance, entity, occurred_at, embedding)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector)
ON CONFLICT (host_id, idem_key) DO NOTHING`,
		hostID, key, ep.GetKind(), ep.GetText(), ep.GetImportance(), ep.GetTags()["entity"], at, emb)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 0, nil // 0 rows affected → conflict → duplicate
}

// Recall returns up to topK of the host's episodes most relevant to query. With
// a non-empty query and a wired embedder it ranks by pgvector COSINE similarity
// (semantic); an empty query (used for cold-start bootstrap) or missing embedder
// degrades to Postgres full-text relevance + recency. "kv" mirror episodes are
// excluded. The host can only read its own host_id.
func (l *LTM) Recall(ctx context.Context, hostID, query string, topK int) ([]storedEpisode, error) {
	if topK <= 0 {
		topK = 5
	}
	if query != "" && l.embedder != nil {
		if vecs, err := l.embedder.Embed(ctx, []string{query}, "query"); err == nil && len(vecs) == 1 {
			return l.recallCosine(ctx, hostID, vecLiteral(vecs[0]), topK)
		}
		// Embedding failed → fall through to the lexical path rather than erroring.
	}
	return l.recallLexical(ctx, hostID, query, topK)
}

// recallCosine ranks the host's embedded episodes by cosine similarity to qvec.
func (l *LTM) recallCosine(ctx context.Context, hostID, qvec string, topK int) ([]storedEpisode, error) {
	rows, err := l.pool.Query(ctx, `
SELECT kind, body, importance, entity, occurred_at, 1 - (embedding <=> $2::vector) AS score
FROM episodes
WHERE host_id = $1 AND kind <> 'kv' AND embedding IS NOT NULL
ORDER BY embedding <=> $2::vector
LIMIT $3`, hostID, qvec, topK)
	if err != nil {
		return nil, err
	}
	return scanEpisodes(rows)
}

// recallLexical ranks by full-text relevance then recency (the bootstrap path
// and the no-embedder fallback). An empty query → pure recency (newest first).
func (l *LTM) recallLexical(ctx context.Context, hostID, query string, topK int) ([]storedEpisode, error) {
	rows, err := l.pool.Query(ctx, `
SELECT kind, body, importance, entity, occurred_at,
       ts_rank(to_tsvector('english', body || ' ' || kind || ' ' || entity),
               websearch_to_tsquery('english', $2)) AS rank
FROM episodes
WHERE host_id = $1 AND kind <> 'kv'
ORDER BY rank DESC, occurred_at DESC
LIMIT $3`, hostID, query, topK)
	if err != nil {
		return nil, err
	}
	return scanEpisodes(rows)
}

func scanEpisodes(rows pgx.Rows) ([]storedEpisode, error) {
	defer rows.Close()
	var out []storedEpisode
	for rows.Next() {
		var se storedEpisode
		var at time.Time
		if err := rows.Scan(&se.Kind, &se.Text, &se.Importance, &se.Entity, &at, &se.Score); err != nil {
			return nil, err
		}
		se.At = at.Unix()
		out = append(out, se)
	}
	return out, rows.Err()
}

// SyncRelationships replaces the host's stored trust ledger with the pushed
// snapshot (relationships are AuthLocal — the host owns the truth). Each row is
// upserted; rows the host no longer has are left in place (lossless: mesa never
// deletes a relationship, matching the never-deleted design). Returns the count
// written.
func (l *LTM) SyncRelationships(ctx context.Context, hostID string, rels []*mesapb.Relationship) (int, error) {
	if len(rels) == 0 {
		return 0, nil
	}
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit
	for _, r := range rels {
		tags := r.GetTags()
		if tags == nil {
			tags = []string{}
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO relationships (host_id, name, alpha, beta, encounters, tags, value_traded, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now())
ON CONFLICT (host_id, name) DO UPDATE SET
    alpha = EXCLUDED.alpha, beta = EXCLUDED.beta, encounters = EXCLUDED.encounters,
    tags = EXCLUDED.tags, value_traded = EXCLUDED.value_traded, updated_at = now()`,
			hostID, r.GetName(), r.GetAlpha(), r.GetBeta(), r.GetEncounters(), tags, r.GetValueTraded()); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(rels), nil
}

// Relationships returns the host's full stored trust ledger (for cold-start
// bootstrap), most-familiar first.
func (l *LTM) Relationships(ctx context.Context, hostID string) ([]*mesapb.Relationship, error) {
	rows, err := l.pool.Query(ctx, `
SELECT name, alpha, beta, encounters, tags, value_traded
FROM relationships WHERE host_id = $1
ORDER BY encounters DESC`, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*mesapb.Relationship
	for rows.Next() {
		r := &mesapb.Relationship{}
		var tags []string
		if err := rows.Scan(&r.Name, &r.Alpha, &r.Beta, &r.Encounters, &tags, &r.ValueTraded); err != nil {
			return nil, err
		}
		r.Tags = tags
		out = append(out, r)
	}
	return out, rows.Err()
}

// RecordMetrics appends a host's telemetry batch as a time series (one row per
// sample). Returns the number of samples written.
func (l *LTM) RecordMetrics(ctx context.Context, hostID string, metrics []*mesapb.Metric, at time.Time) (int, error) {
	if len(metrics) == 0 {
		return 0, nil
	}
	if at.IsZero() {
		at = time.Now()
	}
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit
	for _, m := range metrics {
		if _, err := tx.Exec(ctx,
			`INSERT INTO metrics (host_id, name, value, at) VALUES ($1, $2, $3, $4)`,
			hostID, m.GetName(), m.GetValue(), at); err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(metrics), nil
}

// UpsertGoal stores (or replaces) the host's standing objective + progress.
func (l *LTM) UpsertGoal(ctx context.Context, hostID, objective string, progress []string) error {
	if progress == nil {
		progress = []string{}
	}
	pj, err := json.Marshal(progress)
	if err != nil {
		return err
	}
	_, err = l.pool.Exec(ctx, `
INSERT INTO goals (host_id, objective, progress, updated_at) VALUES ($1, $2, $3::jsonb, now())
ON CONFLICT (host_id) DO UPDATE SET objective = EXCLUDED.objective, progress = EXCLUDED.progress, updated_at = now()`,
		hostID, objective, string(pj))
	return err
}

// Goal reads the host's standing objective + progress. found=false on a miss.
func (l *LTM) Goal(ctx context.Context, hostID string) (objective string, progress []string, found bool, err error) {
	var pj []byte
	err = l.pool.QueryRow(ctx,
		`SELECT objective, progress FROM goals WHERE host_id = $1`, hostID).Scan(&objective, &pj)
	if err == pgx.ErrNoRows {
		return "", nil, false, nil
	}
	if err != nil {
		return "", nil, false, err
	}
	_ = json.Unmarshal(pj, &progress)
	return objective, progress, true, nil
}

// PutKV stores (or replaces) a host-namespaced key/value blob.
func (l *LTM) PutKV(ctx context.Context, hostID, key string, value []byte) error {
	_, err := l.pool.Exec(ctx, `
INSERT INTO kv (host_id, key, value, updated_at) VALUES ($1, $2, $3, now())
ON CONFLICT (host_id, key) DO UPDATE SET value = EXCLUDED.value, updated_at = now()`,
		hostID, key, value)
	return err
}

// GetKV reads a host-namespaced value. found=false (not an error) on a miss.
func (l *LTM) GetKV(ctx context.Context, hostID, key string) (value []byte, found bool, err error) {
	err = l.pool.QueryRow(ctx, `SELECT value FROM kv WHERE host_id = $1 AND key = $2`, hostID, key).Scan(&value)
	if err == pgx.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	return value, true, nil
}

// DeleteKV removes a host-namespaced key (no error if absent).
func (l *LTM) DeleteKV(ctx context.Context, hostID, key string) error {
	_, err := l.pool.Exec(ctx, `DELETE FROM kv WHERE host_id = $1 AND key = $2`, hostID, key)
	return err
}

// hostPersona is a stored persona row (the compiled-identity JSON).
type hostPersona struct {
	HostID string
	JSON   []byte
}

// UpsertPersona persists (or replaces) a host's persona: the canonical JSON plus
// the derived prose_card (the card the brain reads — Render() floor or cook
// output) and the cooked flag. The facet/dial columns are GENERATED from the JSON
// by Postgres, so they need no parameters here. created_at is preserved across
// updates.
func (l *LTM) UpsertPersona(ctx context.Context, hostID string, data []byte, prose string, cooked bool) error {
	_, err := l.pool.Exec(ctx, `
INSERT INTO personas (host_id, persona_json, prose_card, cooked, updated_at)
VALUES ($1, $2::jsonb, $3, $4, now())
ON CONFLICT (host_id) DO UPDATE SET
  persona_json = EXCLUDED.persona_json,
  prose_card   = EXCLUDED.prose_card,
  cooked       = EXCLUDED.cooked,
  updated_at   = now()`,
		hostID, string(data), prose, cooked)
	return err
}

// Personas returns every stored persona (for loading into the registry at start).
func (l *LTM) Personas(ctx context.Context) ([]hostPersona, error) {
	rows, err := l.pool.Query(ctx, `SELECT host_id, persona_json FROM personas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []hostPersona
	for rows.Next() {
		var hp hostPersona
		if err := rows.Scan(&hp.HostID, &hp.JSON); err != nil {
			return nil, err
		}
		out = append(out, hp)
	}
	return out, rows.Err()
}

// DeletePersona removes a host's stored persona. Idempotent (no error if absent).
func (l *LTM) DeletePersona(ctx context.Context, hostID string) error {
	_, err := l.pool.Exec(ctx, `DELETE FROM personas WHERE host_id = $1`, hostID)
	return err
}

// PersonaTimes returns host_id → last-update time for every stored persona, for
// the admin List/Get metadata. One query; the in-memory registry stays the live
// source of truth for the set itself.
func (l *LTM) PersonaTimes(ctx context.Context) (map[string]time.Time, error) {
	rows, err := l.pool.Query(ctx, `SELECT host_id, updated_at FROM personas`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]time.Time{}
	for rows.Next() {
		var id string
		var t time.Time
		if err := rows.Scan(&id, &t); err != nil {
			return nil, err
		}
		out[id] = t
	}
	return out, rows.Err()
}

// Count reports how many recallable (non-kv) episodes a host has stored.
func (l *LTM) Count(ctx context.Context, hostID string) (int, error) {
	var n int
	err := l.pool.QueryRow(ctx,
		`SELECT count(*) FROM episodes WHERE host_id = $1 AND kind <> 'kv'`, hostID).Scan(&n)
	return n, err
}

// vecLiteral formats a float vector as a pgvector text literal "[v1,v2,...]"
// (cast to ::vector in the query).
func vecLiteral(v []float32) string {
	var b strings.Builder
	b.Grow(len(v) * 8)
	b.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(f), 'f', -1, 32))
	}
	b.WriteByte(']')
	return b.String()
}

// contentKey derives a stable idempotency key from an episode's content when the
// client did not supply one — kind+text+timestamp uniquely identifies a memory
// and is robust to a host's local journal sequence resetting.
func contentKey(ep *mesapb.Episode) string {
	return ep.GetKind() + "|" + ep.GetText() + "|" + timeKey(ep.GetOccurredAtUnix())
}

func timeKey(at int64) string {
	if at == 0 {
		return ""
	}
	return strings.TrimSpace(time.Unix(at, 0).UTC().Format("20060102T150405"))
}
