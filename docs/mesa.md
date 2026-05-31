# Mesa — the memory and knowledge service

> **STATUS: ASPIRATIONAL / NOT BUILT** (verified 2026-05-31). Mesa
> does not exist as a running service. The `mesa/` tree contains only
> empty `mesa/client/` and `mesa/server/` directories (no Go files),
> and `cmd/mesa/` is empty. Every schema (`CREATE TABLE` for bots,
> episodes, relationships, reflections, routines, brain_calls,
> knowledge_chunks), every HTTP endpoint (`/bots`, `/episodes`,
> `/relationships`, `/knowledge`, ...), and every background job
> (reflection generator, decay/compression, importance scorer)
> below is a **specification**, not documentation of an existing
> system. There is no `HTTPClient`; the `MockClient` referenced for
> tests is also not yet present in code. Treat this whole doc as the
> Phase 2.6+ design contract for the service. (Phase 2.6 knowledge
> ingestion — pgvector knowledge_chunks + Voyage 3 embeds — is the
> first slice slated to land; see `docs/state.md`.)

## What mesa is

Mesa is the central HTTP service that holds everything *persistent and shared* about the bot population:

- Per-host memory: episodic, relational, reflective, working-buffer flushes
- Per-host routine library (DSL scripts the host has authored), with full version history
- Per-host persona snapshots, with revision history (personas evolve over time)
- Knowledge corpus: rsc.wiki content, F2P-filtered, embedded for RAG
- Chat corpus (when introduced): RSC-era player chat samples for social believability
- Cohort metadata: every host belongs to a cohort; mesa scopes analytics by cohort_id
- Brain call ledger: every LLM call's prompt, response, model, tokens, cost — the chain-of-thought archive

Stack: Go server, Postgres + pgvector, Voyage 3 for embeddings.

## Why a service and not per-host sqlite

Three reasons:

1. **Shared ingestion of knowledge.** rsc.wiki gets embedded once into mesa, not once per host process. 500 hosts don't each load a 2M-token vector index.
2. **Population-scale queries.** "Find clusters of hosts that frequently trade with each other" is a SQL join in mesa. Same query across 500 sqlite files would be a custom MapReduce.
3. **Crash recovery.** Hosts crash, restart, redeploy. Memory persists in mesa. A host coming back online resumes its identity, not its progress.

The trade-off is network latency per memory operation. We absorb that by only hitting mesa on slow cadences (strategist calls, episode writes) — never per game tick.

## Identity model

Every host has:
- `bot_id` (UUID, assigned at registration)
- `auth_token` (bearer token, issued at registration, used for all subsequent calls)
- `username` (the RSC account name; foreign key to OpenRSC's `players` table conceptually)
- `cohort_id` (assigned at registration based on the cohort spawn config)
- `persona` (JSON blob; templates + LLM-generated variation)

API requests come with the bearer token; mesa scopes all data access to that host's records.

## Schema (illustrative)

```sql
-- Identity
CREATE TABLE bots (
    bot_id        UUID PRIMARY KEY,
    auth_token    TEXT NOT NULL UNIQUE,
    rsc_username  TEXT NOT NULL UNIQUE,
    cohort_id     TEXT NOT NULL,
    persona       JSONB NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at  TIMESTAMPTZ
);
CREATE INDEX idx_bots_cohort ON bots(cohort_id);

-- Episodic memory: things that happened
CREATE TABLE episodes (
    id             BIGSERIAL PRIMARY KEY,
    bot_id         UUID NOT NULL REFERENCES bots,
    occurred_at    TIMESTAMPTZ NOT NULL,
    importance     REAL NOT NULL,             -- 0..1
    reinforcement  INTEGER DEFAULT 0,
    last_accessed  TIMESTAMPTZ,
    stage          SMALLINT DEFAULT 1,        -- 1=full, 2=summary, 3=gist
    content        TEXT NOT NULL,
    embedding      VECTOR(1024)               -- Voyage 3 dimensions
);
CREATE INDEX idx_episodes_bot_salience ON episodes(bot_id, importance, last_accessed);
CREATE INDEX idx_episodes_embedding ON episodes USING ivfflat (embedding vector_cosine_ops);

-- Relational records: per-player social facts (sticky)
CREATE TABLE relationships (
    bot_id            UUID NOT NULL REFERENCES bots,
    other_username    TEXT NOT NULL,
    first_seen        TIMESTAMPTZ NOT NULL,
    last_seen         TIMESTAMPTZ NOT NULL,
    encounter_count   INTEGER DEFAULT 1,
    trade_count       INTEGER DEFAULT 0,
    chat_count        INTEGER DEFAULT 0,
    pvp_count         INTEGER DEFAULT 0,
    total_value_traded BIGINT DEFAULT 0,
    trust_score       REAL DEFAULT 0,         -- -1..1
    relationship      TEXT,                   -- 'stranger', 'acquaintance', 'friend', 'rival', 'enemy'
    notes             TEXT,                   -- agent-written free-form
    PRIMARY KEY (bot_id, other_username)
);

-- Reflective memory: higher-level insights, revisable
CREATE TABLE reflections (
    id             BIGSERIAL PRIMARY KEY,
    bot_id         UUID NOT NULL REFERENCES bots,
    created_at     TIMESTAMPTZ NOT NULL,
    last_revised   TIMESTAMPTZ,
    confidence     REAL,
    content        TEXT NOT NULL,
    embedding      VECTOR(1024)
);

-- Routines: every host has private routines; versions are tracked over time.
-- (Public/shared routines are deferred — see "routine promotion" in questions-and-decisions.md)
CREATE TABLE routines (
    bot_id          UUID NOT NULL REFERENCES bots,
    name            TEXT NOT NULL,
    description     TEXT NOT NULL,
    current_version INTEGER NOT NULL,          -- pointer to routine_versions
    PRIMARY KEY (bot_id, name)
);

-- Routine versions: every revision of every routine is preserved.
-- This is the substrate for "evolution tracking" — a research-grade
-- observable into how a host's behavior changes over time.
CREATE TABLE routine_versions (
    bot_id           UUID NOT NULL REFERENCES bots,
    name             TEXT NOT NULL,
    version          INTEGER NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL,
    dsl_spec_version TEXT NOT NULL,            -- which DSL grammar version
    source           TEXT NOT NULL,            -- DSL source
    origin           TEXT NOT NULL,            -- 'initial', 'self_revision',
                                                -- 'observed_from:<other_bot_id>',
                                                -- 'exec_promotion:<parent_routine>'
    parent_version   INTEGER,                  -- previous version (null for v1)
    rationale        TEXT,                     -- agent's own description of the change
    success_count    INTEGER DEFAULT 0,
    failure_count    INTEGER DEFAULT 0,
    punt_count       INTEGER DEFAULT 0,        -- times this version called contemplate/exec
    last_used        TIMESTAMPTZ,
    embedding        VECTOR(1024),             -- of description + source
    PRIMARY KEY (bot_id, name, version),
    FOREIGN KEY (bot_id, name) REFERENCES routines (bot_id, name)
);
CREATE INDEX idx_routine_versions_origin ON routine_versions (origin);
CREATE INDEX idx_routine_versions_created ON routine_versions (created_at);

-- Persona handler revisions: persona-level default handlers can also evolve
-- (e.g., a host that gets scammed develops a warier `on chat_received` default).
-- The persona JSON in `bots` holds the current set; this table tracks history.
CREATE TABLE persona_revisions (
    bot_id           UUID NOT NULL REFERENCES bots,
    revision         INTEGER NOT NULL,
    revised_at       TIMESTAMPTZ NOT NULL,
    persona_snapshot JSONB NOT NULL,
    rationale        TEXT,                     -- why the persona changed
    PRIMARY KEY (bot_id, revision)
);

-- Brain call ledger (for observability + cost tracking + chain-of-thought)
CREATE TABLE brain_calls (
    id             BIGSERIAL PRIMARY KEY,
    bot_id         UUID NOT NULL REFERENCES bots,
    called_at      TIMESTAMPTZ NOT NULL,
    decision_class TEXT NOT NULL,             -- 'strategic', 'tactical', 'chat', etc.
    model          TEXT NOT NULL,
    prompt         TEXT NOT NULL,
    response       TEXT NOT NULL,
    input_tokens   INTEGER,
    output_tokens  INTEGER,
    cost_usd       NUMERIC(10, 6),
    latency_ms     INTEGER
);
CREATE INDEX idx_brain_bot_time ON brain_calls(bot_id, called_at DESC);

-- Knowledge corpus (shared)
CREATE TABLE knowledge_chunks (
    id            BIGSERIAL PRIMARY KEY,
    source        TEXT NOT NULL,              -- 'rsc.wiki', 'chat_corpus'
    page_url      TEXT,
    chunk_text    TEXT NOT NULL,
    members_only  BOOLEAN DEFAULT FALSE,
    embedding     VECTOR(1024)
);
CREATE INDEX idx_knowledge_embedding ON knowledge_chunks USING ivfflat (embedding vector_cosine_ops);
```

## API surface (illustrative)

All endpoints scope by bot_id automatically based on bearer token. No cross-bot data leakage at the API layer.

```
# Identity
POST   /bots                           # register; returns {bot_id, auth_token, persona, cohort_id}
GET    /bots/me                        # fetch own metadata

# Episodic
POST   /bots/me/episodes               # write event; mesa computes embedding + importance (if not specified)
GET    /bots/me/episodes               # vector retrieval + filters (top_k, since, importance_min)
PATCH  /bots/me/episodes/:id           # update (e.g., compression stage)

# Relational
POST   /bots/me/relationships          # upsert
GET    /bots/me/relationships          # query by username list (batch: ?nearby=alice,bob,charlie)
GET    /bots/me/relationships/:username

# Reflections
POST   /bots/me/reflections
GET    /bots/me/reflections
PATCH  /bots/me/reflections/:id        # revision

# Routines (private library; versions tracked)
POST   /bots/me/routines               # save a routine; returns version=1 if new, or version=N+1 if revision
GET    /bots/me/routines               # list current versions (one entry per named routine)
GET    /bots/me/routines/:name         # get current version full content
GET    /bots/me/routines/:name/versions      # full version history with rationale + origin
GET    /bots/me/routines/:name/versions/:v   # specific historical version
DELETE /bots/me/routines/:name         # soft-delete (keep version history; mark current_version = -1)

# Persona revisions
PATCH  /bots/me/persona                # revise persona; creates a new persona_revision row
GET    /bots/me/persona/revisions      # full persona history

# Knowledge RAG
GET    /knowledge?q=...&top=5&source=rsc.wiki  # vector retrieval

# Brain ledger
POST   /bots/me/brain-calls            # log a brain call (for observability)

# Background triggers (return task_id; async)
POST   /bots/me/reflect                # generate new reflections from recent episodes
POST   /bots/me/consolidate            # decay/compress low-salience episodes

# Research / admin
GET    /admin/cohorts/:cohort/episodes  # cross-host within cohort (admin token required)
GET    /admin/social-graph             # full network graph
GET    /admin/brain-cost                # cost dashboard
```

## Background jobs (mesa-side)

Mesa runs three flavors of background tasks:

1. **Reflection generator**: per-host, every ~30 active minutes OR triggered by host. Pulls recent high-importance episodes, calls Sonnet to generate 1-3 new reflections, inserts/revises. Cheap because rare.

2. **Decay/compression sweep**: per-host, hourly. Computes salience = `importance × exp(-(now - last_accessed) / tau) + reinforcement_bonus`. Episodes below threshold get stage-bumped (full → summary → gist → dropped). Compression calls Haiku to summarize. Cheap per call, high volume in aggregate.

3. **Importance scorer**: triggered on every episode write. Heuristic first (built-in rules for common event types); fallback to Haiku for borderline events. Mostly free.

## Mock client for local dev

Phase 0-2 don't need mesa running. The `mesa/client` package has a `MockClient` that satisfies the same interface but stores everything in-memory:

```go
type Client interface {
    WriteEpisode(ctx context.Context, e Episode) error
    QueryEpisodes(ctx context.Context, q EpisodeQuery) ([]Episode, error)
    GetRelationship(ctx context.Context, other string) (*Relationship, error)
    // ...
}

type HTTPClient struct{ ... }   // real production impl
type MockClient struct{ ... }   // in-memory for dev
```

Tests use `MockClient`. Phase 0-2 binaries can run with either. Phase 3+ defaults to `HTTPClient` against a running mesa.

## Cohort-aware queries

Cohorts are first-class. Examples of admin queries:

```
-- Compare avg episode count between cohorts
SELECT b.cohort_id, COUNT(e.id)/COUNT(DISTINCT b.bot_id) AS avg_episodes
FROM bots b LEFT JOIN episodes e USING (bot_id)
GROUP BY b.cohort_id;

-- Find trade graph density per cohort
SELECT b.cohort_id,
       COUNT(DISTINCT r.bot_id || ':' || r.other_username) AS edges,
       COUNT(DISTINCT b.bot_id) AS nodes,
       COUNT(DISTINCT r.bot_id || ':' || r.other_username)::float / 
         (COUNT(DISTINCT b.bot_id) * (COUNT(DISTINCT b.bot_id)-1)) AS density
FROM bots b LEFT JOIN relationships r USING (bot_id)
WHERE r.trade_count > 0
GROUP BY b.cohort_id;
```

This is the kind of thing that *only* works because memory is centralized.

## Deployment

For dev: Postgres in a local container, mesa binary running locally. ~30s setup.

For production (eventually): managed Postgres (RDS, Crunchy, Supabase) + mesa on a VPS. Postgres backups via pg_dump on schedule. Voyage API key in env. Anthropic key in env.

## Routine evolution as a research observable

The versioned routine schema enables queries that are first-class research outputs:

```sql
-- How fast does a host's routine learning slow down? (Convergence curve)
SELECT name, version, created_at,
       LAG(created_at) OVER (PARTITION BY bot_id, name ORDER BY version) AS prev_at,
       success_count, failure_count, punt_count
FROM routine_versions
WHERE bot_id = $1
ORDER BY name, version;

-- Which hosts converged on similar routines from different paths?
-- (Routine version embeddings cluster in vector space)
SELECT a.bot_id, b.bot_id, a.name, b.name,
       1 - (a.embedding <=> b.embedding) AS similarity
FROM routine_versions a JOIN routine_versions b
  ON a.bot_id < b.bot_id
WHERE a.version = (SELECT current_version FROM routines WHERE bot_id = a.bot_id AND name = a.name)
  AND b.version = (SELECT current_version FROM routines WHERE bot_id = b.bot_id AND name = b.name)
  AND 1 - (a.embedding <=> b.embedding) > 0.85
ORDER BY similarity DESC;

-- When a host's routine version was created with origin='observed_from:X',
-- does that learning persist (more versions follow) or get abandoned?
SELECT v1.bot_id, v1.name, v1.version, v1.origin,
       COUNT(v2.version) AS subsequent_versions
FROM routine_versions v1
LEFT JOIN routine_versions v2 ON v2.bot_id = v1.bot_id AND v2.name = v1.name AND v2.version > v1.version
WHERE v1.origin LIKE 'observed_from:%'
GROUP BY v1.bot_id, v1.name, v1.version, v1.origin;
```

The `punt_count` column on routine_versions is particularly interesting: each version records how many times the routine called `contemplate_reality()` or `exec()` during its tenure. Mature routines have low punt rates; new ones have high. This gives us a measurable "learning curve" per routine per host.

## Open questions

- Schema evolution: how do we migrate when persona JSON shape changes? Probably JSONB tolerance + per-version migration scripts.
- Backup/restore strategy: standard pg_dump for v1.
- Routine version garbage collection: do we ever delete old versions? Probably no — storage is cheap, research observability of full history is valuable. Possibly compress old version bodies if space becomes a concern (text compression at column level via Postgres' built-in TOAST is automatic).
- Cross-host routine influence detection: if BotA's routine fish_at_swamp_v3 is structurally similar to BotB's fish_at_swamp_v7, and BotA and BotB were co-located when BotA wrote v3, did BotB observe and learn? The origin field captures explicit attribution; vector-similarity heuristics could find implicit borrowing.
