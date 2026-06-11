package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// observationCaptureKinds are the firehose CAPTURE kinds — raw, free-text speech
// lines where many distinct lines legitimately share kind|subject within one second
// (subject is often "server" for SystemMessage/ChatReceived). They get a text-hashed
// idempotency key so distinct same-second lines do NOT collapse under ON CONFLICT DO
// NOTHING (M4); every other kind (outcome/transaction/location) is once-per-event and
// keeps the coarse key. MUST mirror runtime observation.go's observationCaptureKinds.
var observationCaptureKinds = map[string]bool{
	"npc_dialog":  true,
	"player_chat": true,
	"server_msg":  true,
	"claim_heard": true,
}

// observationIdemKey is the server-side fallback idempotency key for an observation
// whose client did not supply one. It MUST be byte-identical to the host emitter's key
// (runtime observation.go observationIdemKey): kind|subject|second|<fnv32a(text) hex>
// for CAPTURE kinds, coarse kind|subject|second otherwise (M4).
func observationIdemKey(kind, subject, text string, occurredAtUnix int64) string {
	if observationCaptureKinds[kind] {
		h := fnv.New32a()
		h.Write([]byte(text))
		return fmt.Sprintf("%s|%s|%d|%08x", kind, subject, occurredAtUnix, h.Sum32())
	}
	return fmt.Sprintf("%s|%s|%d", kind, subject, occurredAtUnix)
}

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
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	// Give the pool headroom so the consolidation crons (bounded by cronSem) can't
	// exhaust connections and starve the latency-sensitive host RPCs (Act/Chat/
	// Decide). pgx's default max(4, NumCPU) collides with the cron concurrency on a
	// small box; reserve well above it (honor a larger DSN-supplied pool_max_conns).
	if cfg.MaxConns < 16 {
		cfg.MaxConns = 16
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
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
		// observations: the raw, salience-gated perception firehose (cron fodder,
		// distilled later into the knowledge/relationship ledgers). One row per
		// emit, deduped by idempotency key. NOT recalled into the planner like
		// episodes — kept separate so the firehose never pollutes episode recall.
		`CREATE TABLE IF NOT EXISTS observations (
    host_id     text             NOT NULL,
    idem_key    text             NOT NULL,
    kind        text             NOT NULL,
    subject     text             NOT NULL DEFAULT '',
    body        text             NOT NULL DEFAULT '',
    salience    double precision NOT NULL DEFAULT 0,
    occurred_at timestamptz      NOT NULL,
    created_at  timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, idem_key)
)`,
		`CREATE INDEX IF NOT EXISTS observations_host_time ON observations (host_id, occurred_at DESC)`,
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
    affinity     double precision NOT NULL DEFAULT 0,
    grievance    double precision NOT NULL DEFAULT 0,
    updated_at   timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, name)
)`,
		// Phase-3b multi-axis relationships on an EXISTING relationships table:
		// CREATE TABLE IF NOT EXISTS won't add columns, so ALTER for live DBs.
		`ALTER TABLE relationships ADD COLUMN IF NOT EXISTS affinity double precision NOT NULL DEFAULT 0`,
		`ALTER TABLE relationships ADD COLUMN IF NOT EXISTS grievance double precision NOT NULL DEFAULT 0`,
		// knowledge: the host's distilled world-knowledge ledger — graded,
		// provenance-tagged beliefs about THINGS (npc/place/shop/item/mechanic/
		// quest). The consolidation cron folds the observation firehose into these;
		// the host also pushes its locally-learned beliefs up. Both write the same
		// row by (host_id, subject) — last-writer-wins per subject (full α/β merge is
		// 4b). Served back for a cold-start bootstrap. Mirrors the relationships
		// table's spine. beliefs_json holds []KnowledgeBelief.
		`CREATE TABLE IF NOT EXISTS knowledge (
    host_id      text             NOT NULL,
    subject      text             NOT NULL,
    kind         text             NOT NULL DEFAULT '',
    encounters   integer          NOT NULL DEFAULT 0,
    last_seen    timestamptz      NOT NULL DEFAULT now(),
    tags         text[]           NOT NULL DEFAULT '{}',
    beliefs_json jsonb            NOT NULL DEFAULT '[]',
    updated_at   timestamptz      NOT NULL DEFAULT now(),
    PRIMARY KEY (host_id, subject)
)`,
		`CREATE INDEX IF NOT EXISTS knowledge_host_time ON knowledge (host_id, last_seen DESC)`,
		// goal_graphs: the host's distilled INTENTION graph (goals/sub-goals/open-
		// questions + typed edges). The host pushes its local graph up; the insight
		// cron grows it (open-question closure, cross-entity chaining). Both write the
		// same row by host_id — last-writer-wins per host. Served back for a cold-start
		// bootstrap. PK is the only access path (no extra index). snapshot holds
		// {nodes:[],edges:[]} (GoalGraphSnapshot, minus the HostRef).
		`CREATE TABLE IF NOT EXISTS goal_graphs (
    host_id    text        NOT NULL PRIMARY KEY,
    snapshot   jsonb       NOT NULL DEFAULT '{"nodes":[],"edges":[]}',
    updated_at timestamptz NOT NULL DEFAULT now()
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
	// L15: GENERATED STORED columns evaluate their expression over EVERY existing
	// row at ALTER time, so a hard cast (`::int`/`::real`) on a persona row whose
	// JSON carries a non-numeric (or otherwise malformed) value would fail the whole
	// ALTER and brick OpenLTM startup (statements run un-transactioned below). Use
	// NULL-tolerant casts: a regex-guarded CASE yields NULL for a missing/malformed
	// value instead of erroring. The result must be deterministic (a generated-column
	// requirement) — `~` over the extracted text is.
	jpath := func(path ...string) string {
		return fmt.Sprintf("persona_json #>> '{%s}'", strings.Join(path, ","))
	}
	type gcol struct{ name, expr, typ string }
	gtext := func(name string, path ...string) gcol {
		return gcol{name, jpath(path...), "text"}
	}
	// greal extracts a numeric path NULL-tolerantly: cast only when the text looks
	// like a (signed, optionally-fractional/exponent) number, else NULL.
	greal := func(name string, path ...string) gcol {
		p := jpath(path...)
		return gcol{name, fmt.Sprintf(
			`CASE WHEN (%s) ~ '^-?[0-9]+(\.[0-9]+)?([eE][-+]?[0-9]+)?$' THEN (%s)::real END`, p, p), "real"}
	}
	svPath := jpath("schema_version")
	gcols := []gcol{
		{"schema_version", fmt.Sprintf(`CASE WHEN (%s) ~ '^-?[0-9]+$' THEN (%s)::int END`, svPath, svPath), "int"},
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
	}

	// L15 remediation for the LIVE DB: the numeric generated columns may already
	// exist with the OLD hard-cast expression (`::int`/`::real`), and ADD COLUMN IF
	// NOT EXISTS will NOT redefine an existing column — so the fix would be inert on
	// a deployed DB. A generated-column expression can't be ALTERed in place, so for
	// each numeric facet drop the column ONLY when it exists with a non-NULL-tolerant
	// (no `CASE`) expression; the ADD COLUMN below then recreates it NULL-tolerantly.
	// Idempotent (a column already carrying `CASE` is left alone), and lossless
	// (the columns are derived from persona_json, so a drop+recompute loses nothing).
	for _, g := range gcols {
		if g.typ == "text" {
			continue // text facets already use the safe `#>>` extraction (no cast)
		}
		stmts = append(stmts, fmt.Sprintf(`DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'personas' AND column_name = '%s'
      AND generation_expression IS NOT NULL
      AND generation_expression NOT ILIKE '%%CASE%%'
  ) THEN
    EXECUTE 'ALTER TABLE personas DROP COLUMN %s';
  END IF;
END $$`, g.name, g.name))
	}

	for _, g := range gcols {
		stmts = append(stmts, fmt.Sprintf(
			`ALTER TABLE personas ADD COLUMN IF NOT EXISTS %s %s GENERATED ALWAYS AS (%s) STORED`,
			g.name, g.typ, g.expr))
	}

	// Run the whole migration in ONE transaction so a single failing statement
	// rolls back cleanly instead of leaving a half-applied schema (L15 backstop —
	// the NULL-tolerant casts above are the primary defense; this keeps migrate()
	// atomic). Every statement here is plain transactional DDL (no CONCURRENTLY),
	// and all are idempotent (IF NOT EXISTS), so a re-run is a no-op.
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("migrate: begin: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit
	for _, s := range stmts {
		if _, err := tx.Exec(ctx, s); err != nil {
			return fmt.Errorf("migrate: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("migrate: commit: %w", err)
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
// idempotency key. When an embedder is wired the text is embedded and stored for
// semantic recall; an embedding failure is tolerated (the episode is still
// stored, just without a vector — never lose a memory to a transient API error).
// Returns dup=true when the key already existed.
//
// M12: embeddings are no longer write-once. An episode first stored with a NULL
// embedding (no embedder wired yet, or a transient Voyage error) is otherwise
// invisible to recallCosine (which filters `embedding IS NOT NULL`) forever. The
// host mirrors its journal up at-least-once, so a later resend with a now-present
// embedding BACKFILLS the missing vector via ON CONFLICT DO UPDATE … COALESCE,
// gated so it only writes when the stored vector is NULL and a new one exists
// (an ordinary resend of an already-embedded row stays a cheap no-op).
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
	// The episode INSERT and the relationship-delta fold (when present) must commit
	// ATOMICALLY: the fold is gated on a genuine INSERT, so if the fold could fail
	// independently a retry would re-hit ON CONFLICT (inserted=false), SKIP the fold,
	// and silently under-count value_traded. One transaction makes a fold failure
	// roll back the episode too, so the at-least-once retry replays BOTH.
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	// RETURNING (xmax = 0): true on a fresh INSERT, false on the conflict UPDATE.
	// The DO UPDATE fires only to backfill a missing embedding, so an ordinary
	// resend filters out of the WHERE → no row returned → ErrNoRows → dup.
	var inserted bool
	qerr := tx.QueryRow(ctx, `
INSERT INTO episodes (host_id, idem_key, kind, body, importance, entity, occurred_at, embedding)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::vector)
ON CONFLICT (host_id, idem_key) DO UPDATE SET
    embedding = COALESCE(episodes.embedding, EXCLUDED.embedding)
WHERE episodes.embedding IS NULL AND EXCLUDED.embedding IS NOT NULL
RETURNING (xmax = 0)`,
		hostID, key, ep.GetKind(), ep.GetText(), ep.GetImportance(), ep.GetTags()["entity"], at, emb).Scan(&inserted)
	if qerr == pgx.ErrNoRows {
		return true, nil // conflict whose backfill WHERE was false → duplicate (tx rolls back, a no-op)
	}
	if qerr != nil {
		return false, qerr
	}
	// Fold the episode's relationship delta (if any) into the trust ledger. The only
	// piece the SyncRelationships snapshot path does NOT already carry is the per-trade
	// VOLUME: relationships.value_traded is a monotone total, so a completed-trade
	// episode ADDS its TotalValueTraded to the (host_id,name) row. Gated on a genuine
	// INSERT (not a dedup resend) so an at-least-once journal re-mirror can't
	// double-count the same trade. The Beta(α/β)/affinity/grievance axes stay AuthLocal
	// (the host snapshot is the source of truth); only value_traded is journal-accrued.
	if inserted {
		if rel := ep.GetRelation(); rel != nil && rel.GetName() != "" && rel.GetTotalValueTraded() != 0 {
			if _, ferr := tx.Exec(ctx, `
INSERT INTO relationships (host_id, name, value_traded, updated_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (host_id, name) DO UPDATE SET
    value_traded = relationships.value_traded + EXCLUDED.value_traded, updated_at = now()`,
				hostID, rel.GetName(), rel.GetTotalValueTraded()); ferr != nil {
				return false, fmt.Errorf("add: fold relation delta: %w", ferr)
			}
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return false, err
	}
	return !inserted, nil // UPDATE (backfill) counts as a duplicate, not a new row
}

// AddObservation stores one raw observation under the host's namespace, deduped
// by idempotency key (ON CONFLICT DO NOTHING). Returns dup=true when the key
// already existed. The firehose is not embedded — crons distil it later.
func (l *LTM) AddObservation(ctx context.Context, hostID string, o *mesapb.Observation) (dup bool, err error) {
	key := o.GetIdempotencyKey()
	if key == "" {
		key = observationIdemKey(o.GetKind(), o.GetSubject(), o.GetText(), o.GetOccurredAtUnix())
	}
	at := time.Now()
	if u := o.GetOccurredAtUnix(); u != 0 {
		at = time.Unix(u, 0)
	}
	tag, err := l.pool.Exec(ctx, `
INSERT INTO observations (host_id, idem_key, kind, subject, body, salience, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (host_id, idem_key) DO NOTHING`,
		hostID, key, o.GetKind(), o.GetSubject(), o.GetText(), o.GetSalience(), at)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 0, nil
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
INSERT INTO relationships (host_id, name, alpha, beta, encounters, tags, value_traded, affinity, grievance, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
ON CONFLICT (host_id, name) DO UPDATE SET
    alpha = EXCLUDED.alpha, beta = EXCLUDED.beta, encounters = EXCLUDED.encounters,
    tags = EXCLUDED.tags, value_traded = EXCLUDED.value_traded,
    affinity = EXCLUDED.affinity, grievance = EXCLUDED.grievance, updated_at = now()`,
			hostID, r.GetName(), r.GetAlpha(), r.GetBeta(), r.GetEncounters(), tags, r.GetValueTraded(), r.GetAffinity(), r.GetGrievance()); err != nil {
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
SELECT name, alpha, beta, encounters, tags, value_traded, affinity, grievance
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
		if err := rows.Scan(&r.Name, &r.Alpha, &r.Beta, &r.Encounters, &tags, &r.ValueTraded, &r.Affinity, &r.Grievance); err != nil {
			return nil, err
		}
		r.Tags = tags
		out = append(out, r)
	}
	return out, rows.Err()
}

// --- observations read seam (cron fodder) ----------------------------------

// observedRow is one raw observation read back for distillation.
type observedRow struct {
	IdemKey  string
	Kind     string
	Subject  string
	Body     string
	Salience float64
	At       int64 // unix seconds (occurred_at)
}

// Observations returns up to limit of the host's observations strictly NEWER than
// sinceUnix, in ascending occurred_at order — so a cron can advance a monotonic
// cursor (the max occurred_at of the returned batch) and never re-read older
// rows. occurred_at is second-granularity and the filter is strict `>`, so a
// crash mid-batch can re-surface rows that share the cursor second; the cron's
// fold is idempotent by (subject,kind,claim), making this safely at-least-once
// (an exactly-once `consolidated` column is a 4b concern).
func (l *LTM) Observations(ctx context.Context, hostID string, sinceUnix int64, limit int) ([]observedRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := l.pool.Query(ctx, `
SELECT idem_key, kind, subject, body, salience, occurred_at
FROM observations
WHERE host_id = $1 AND occurred_at > to_timestamp($2)
ORDER BY occurred_at ASC
LIMIT $3`, hostID, sinceUnix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []observedRow
	for rows.Next() {
		var r observedRow
		var at time.Time
		if err := rows.Scan(&r.IdemKey, &r.Kind, &r.Subject, &r.Body, &r.Salience, &at); err != nil {
			return nil, err
		}
		r.At = at.Unix()
		out = append(out, r)
	}
	return out, rows.Err()
}

// ActiveHostsWithObservations returns the host_ids that emitted an observation
// after sinceUnix, most-recently-active first, capped at limit. The cron iterates
// this WITHOUT touching the in-memory registry (s.reg) — it reads straight from
// Postgres, so it also catches hosts that have since disconnected, and never
// contends on s.mu. Most-recently-active ordering drains the busiest hosts first;
// the tail catches the next tick (anti-starvation bound).
func (l *LTM) ActiveHostsWithObservations(ctx context.Context, sinceUnix int64, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 64
	}
	rows, err := l.pool.Query(ctx, `
SELECT host_id, max(occurred_at) AS m
FROM observations
WHERE occurred_at > to_timestamp($1)
GROUP BY host_id
ORDER BY m DESC
LIMIT $2`, sinceUnix, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		var m time.Time
		if err := rows.Scan(&id, &m); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// --- knowledge durable store (cron + host shared write path) ----------------

// SyncKnowledge upserts the given knowledge entries for a host by (host_id,
// subject) — the SHARED write path for both the consolidation cron and the host
// push-up. Returns the count of rows actually written (blank-subject entries are
// skipped, so this is NOT necessarily len(entries)). An empty set is a no-op.
//
// M17: the upsert is MERGE-aware, not a wholesale per-subject replace. Both the
// cron (read-modify-write, including the insight reconcile β-bumps) and the warm
// host (its OWN local copy, which never saw the reconciliation) write the SAME
// subjects. A blind replace let a warm host's flush clobber a strictly-newer cron
// belief. So on conflict we fold the incoming beliefs into the stored row PER
// (claim, provenance): take the MAX evidence (max α, max β) and the newer AtUnix,
// and keep stored-only beliefs (a host that hasn't re-learned a cron belief must
// not delete it). Subject metadata is monotonic: encounters/last_seen take the
// max so a stale push can't regress familiarity, and kind/tags only overwrite
// when the incoming side is non-empty. This is the same fold the cron applies
// in-process, so the cron's own writes stay idempotent (it already merged against
// the row it read). The cron prunes whole SUBJECTS (gcKnowledge), not individual
// beliefs, so a per-belief max-merge never resurrects a cron-dropped subject.
func (l *LTM) SyncKnowledge(ctx context.Context, hostID string, entries []*mesapb.KnowledgeEntry) (int, error) {
	if len(entries) == 0 {
		return 0, nil
	}
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit
	// Lock in canonical subject order: the host push and the cron fold sync
	// overlapping subjects concurrently, and FOR UPDATE in caller-supplied
	// order deadlocks them against each other (live 40P01s, 2026-06-11 soak).
	sort.Slice(entries, func(i, j int) bool {
		return strings.TrimSpace(entries[i].GetSubject()) < strings.TrimSpace(entries[j].GetSubject())
	})
	var written int
	for _, e := range entries {
		subject := strings.TrimSpace(e.GetSubject())
		if subject == "" {
			continue // L14(a): don't count rows we skip
		}

		// Read the stored row (if any) IN THE TX so a concurrent writer can't slip
		// between read and write. The (host_id) PK row-lock via FOR UPDATE keeps two
		// overlapping syncs from interleaving their folds.
		var stored []byte
		var storedEnc int32
		var storedLastSeen time.Time
		var storedKind string
		var storedTags []string
		rerr := tx.QueryRow(ctx, `
SELECT beliefs_json, encounters, last_seen, kind, tags
FROM knowledge WHERE host_id = $1 AND subject = $2 FOR UPDATE`, hostID, subject).
			Scan(&stored, &storedEnc, &storedLastSeen, &storedKind, &storedTags)
		if rerr != nil && rerr != pgx.ErrNoRows {
			return 0, rerr
		}

		merged := mergeBeliefs(stored, e.GetBeliefs())
		bj, mErr := json.Marshal(merged)
		if mErr != nil {
			return 0, mErr
		}

		// Monotonic subject metadata: never regress familiarity/recency on a stale
		// push; only overwrite kind/tags when the incoming side carries something.
		enc := e.GetEncounters()
		if storedEnc > enc {
			enc = storedEnc
		}
		lastSeen := time.Now()
		if u := e.GetLastSeenUnix(); u != 0 {
			lastSeen = time.Unix(u, 0)
		}
		if rerr == nil && storedLastSeen.After(lastSeen) {
			lastSeen = storedLastSeen
		}
		kind := e.GetKind()
		if strings.TrimSpace(kind) == "" {
			kind = storedKind
		}
		tags := e.GetTags()
		if len(tags) == 0 {
			tags = storedTags
		}
		if tags == nil {
			tags = []string{}
		}

		if _, err := tx.Exec(ctx, `
INSERT INTO knowledge (host_id, subject, kind, encounters, last_seen, tags, beliefs_json, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, now())
ON CONFLICT (host_id, subject) DO UPDATE SET
    kind = EXCLUDED.kind, encounters = EXCLUDED.encounters, last_seen = EXCLUDED.last_seen,
    tags = EXCLUDED.tags, beliefs_json = EXCLUDED.beliefs_json, updated_at = now()`,
			hostID, subject, kind, enc, lastSeen, tags, string(bj)); err != nil {
			return 0, err
		}
		written++
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return written, nil
}

// mergeBeliefs folds the incoming wire beliefs into the stored beliefs_json for
// one subject, keyed by (claim, provenance). For a key in BOTH it takes the MAX
// evidence (max α, max β) and the NEWER AtUnix; a key only on one side is kept as
// is. Stored beliefs that the incoming set omits are preserved (a warm host that
// hasn't re-learned a cron belief must not delete it). The result is ordered:
// stored beliefs first (stable), then incoming-only beliefs. nil incoming beliefs
// normalize to an empty (non-null) slice so a belief-less entry is stored as `[]`,
// matching the column DEFAULT (L14b). Used only by SyncKnowledge's merge upsert.
func mergeBeliefs(stored []byte, incoming []*mesapb.KnowledgeBelief) []*mesapb.KnowledgeBelief {
	type key struct{ claim, prov string }
	out := []*mesapb.KnowledgeBelief{} // L14(b): non-nil → JSON `[]`, not `null`
	idx := map[key]int{}

	var prev []*mesapb.KnowledgeBelief
	if len(stored) > 0 {
		_ = json.Unmarshal(stored, &prev) // a malformed/`null` stored blob → empty
	}
	for _, b := range prev {
		if b == nil {
			continue
		}
		k := key{b.GetClaim(), b.GetProvenance()}
		if _, dup := idx[k]; dup {
			continue
		}
		idx[k] = len(out)
		out = append(out, b)
	}
	for _, b := range incoming {
		if b == nil {
			continue
		}
		k := key{b.GetClaim(), b.GetProvenance()}
		if i, ok := idx[k]; ok {
			cur := out[i]
			if b.GetAlpha() > cur.GetAlpha() {
				cur.Alpha = b.GetAlpha()
			}
			if b.GetBeta() > cur.GetBeta() {
				cur.Beta = b.GetBeta()
			}
			if b.GetAtUnix() > cur.GetAtUnix() {
				cur.AtUnix = b.GetAtUnix()
			}
			continue
		}
		idx[k] = len(out)
		out = append(out, b)
	}
	return out
}

// Knowledge returns the host's full distilled world-knowledge ledger (for a
// cold-start bootstrap), most-recently-seen first.
func (l *LTM) Knowledge(ctx context.Context, hostID string) ([]*mesapb.KnowledgeEntry, error) {
	rows, err := l.pool.Query(ctx, `
SELECT subject, kind, encounters, last_seen, tags, beliefs_json
FROM knowledge WHERE host_id = $1
ORDER BY last_seen DESC`, hostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*mesapb.KnowledgeEntry
	for rows.Next() {
		e := &mesapb.KnowledgeEntry{}
		var tags []string
		var bj []byte
		var lastSeen time.Time
		if err := rows.Scan(&e.Subject, &e.Kind, &e.Encounters, &lastSeen, &tags, &bj); err != nil {
			return nil, err
		}
		e.Tags = tags
		e.LastSeenUnix = lastSeen.Unix()
		var beliefs []*mesapb.KnowledgeBelief
		_ = json.Unmarshal(bj, &beliefs)
		e.Beliefs = beliefs
		out = append(out, e)
	}
	return out, rows.Err()
}

// --- goal-graph durable store (cron + host shared write path) ---------------

// goalGraphRow is the JSON shape stored in goal_graphs.snapshot — the
// GoalGraphSnapshot minus the HostRef (host_id is the row key).
type goalGraphRow struct {
	Nodes []*mesapb.GoalGraphNode `json:"nodes"`
	Edges []*mesapb.GoalGraphEdge `json:"edges"`
}

// SyncGoalGraph upserts the host's full intention-graph snapshot by host_id — the
// SHARED write path for both the insight cron and the host push-up. A nil/empty
// snapshot is a no-op (never clobbers a stored graph with nothing).
//
// H17 (server-side backstop): the goal_graphs row has TWO whole-blob writers — the
// host's 30s flush and the insight cron. The DURABLE fix is the host's warm-host
// merge-import (so its flush already carries the cron's nodes); this is the belt-and-
// suspenders guard underneath it. We read the stored snapshot's freshest-node AtUnix
// IN A TX (FOR UPDATE so two overlapping syncs can't interleave) and REFUSE to
// replace it with a STRICTLY-OLDER snapshot — so a warm host whose local copy lagged
// the cron's write cannot stomp the cron's strictly-newer reconciliation. An equal-
// or-newer incoming snapshot (the normal case: the host re-imported the cron's nodes,
// so its max AtUnix >= the stored one) replaces as before. A graph with no nodes
// (edges only) has version 0 and never out-ranks a stored versioned graph.
func (l *LTM) SyncGoalGraph(ctx context.Context, hostID string, snap *mesapb.GoalGraphSnapshot) error {
	if snap == nil || (len(snap.GetNodes()) == 0 && len(snap.GetEdges()) == 0) {
		return nil
	}
	row := goalGraphRow{Nodes: snap.GetNodes(), Edges: snap.GetEdges()}
	sj, err := json.Marshal(row)
	if err != nil {
		return err
	}
	tx, err := l.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck // no-op after Commit

	var stored []byte
	rerr := tx.QueryRow(ctx,
		`SELECT snapshot FROM goal_graphs WHERE host_id = $1 FOR UPDATE`, hostID).Scan(&stored)
	if rerr != nil && rerr != pgx.ErrNoRows {
		return rerr
	}
	if rerr == nil { // a row exists — apply the strictly-newer guard
		var prev goalGraphRow
		if json.Unmarshal(stored, &prev) == nil {
			if maxNodeAt(snap.GetNodes()) < maxNodeAt(prev.Nodes) {
				return tx.Commit(ctx) // incoming is strictly older — do NOT clobber the cron's newer snapshot
			}
		}
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO goal_graphs (host_id, snapshot, updated_at) VALUES ($1, $2::jsonb, now())
ON CONFLICT (host_id) DO UPDATE SET snapshot = EXCLUDED.snapshot, updated_at = now()`,
		hostID, string(sj)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// maxNodeAt is the freshest-node AtUnix in a snapshot — a content-derived "version"
// of the graph that BOTH whole-blob writers (host flush, insight cron) can compare
// without a separate counter. The cron stamps the nodes it adds/touches with a fresh
// At, and the host stamps every node it writes, so a strictly-newer write always
// raises this. An empty node set is version 0.
func maxNodeAt(nodes []*mesapb.GoalGraphNode) int64 {
	var m int64
	for _, n := range nodes {
		if n != nil && n.GetAtUnix() > m {
			m = n.GetAtUnix()
		}
	}
	return m
}

// GoalGraph returns the host's full distilled intention graph (for a cold-start
// bootstrap). A missing row is an empty snapshot (not an error), mirroring how
// Knowledge degrades — a fresh host bootstraps nothing and proceeds.
func (l *LTM) GoalGraph(ctx context.Context, hostID string) (*mesapb.GoalGraphSnapshot, error) {
	var sj []byte
	err := l.pool.QueryRow(ctx, `SELECT snapshot FROM goal_graphs WHERE host_id = $1`, hostID).Scan(&sj)
	if err == pgx.ErrNoRows {
		return &mesapb.GoalGraphSnapshot{}, nil
	}
	if err != nil {
		return nil, err
	}
	var row goalGraphRow
	if uerr := json.Unmarshal(sj, &row); uerr != nil {
		return nil, uerr
	}
	return &mesapb.GoalGraphSnapshot{Nodes: row.Nodes, Edges: row.Edges}, nil
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
