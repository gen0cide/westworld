# Decision Record: Host Knowledge — seed → growth, and per-host memory RAG

> **STATUS: PARTIALLY BUILT** (verified 2026-06-10 against branch HEAD
> `0bfa818`; decision record 2026-06-06, Alex + Claude). Scope: how a host
> *knows* things about the game — the two-stage persona compile that produces
> its starting knowledge, how that knowledge grows by playing, and how it is
> recalled into runtime LLM prompts.
>
> **BUILT:** Stage 2 (the prose cook — `persona/compile.go`,
> `mesa/personacook`; see `docs/persona-compile.md`); the per-host episodic
> RAG — Postgres+pgvector LTM with Voyage embed-on-write and cosine `Recall`
> (`mesa/mesad/ltm.go`, `mesa/embed/voyage.go`), in the alternative shape §4
> itself noted (plain `WHERE host_id` + HNSW, **no** hash partitioning); and
> the whole growth pipeline (episode mirror, observation firehose,
> distillation crons). **NOT BUILT — live DESIGN:** Stage 1 worldbrief
> generation, `game_familiarity`, the `WorldBrief` refactor, seed→memory
> initialization — tracked as **M-1/M-2** in `docs/TODO.md`.
> Companions: `docs/persona-compile.md` (the cook), `docs/memory.md` (the
> shipped memory stack), `docs/world-knowledge-and-learning.md` (the
> epistemic layer the growth half feeds), `docs/mesa.md` +
> `docs/mesa/ARCHITECTURE.md` (mesa itself), `mesa/wikirag` (the offline wiki
> RAG tool).

---

## 0. The shape in one paragraph

A host's game knowledge has a **seed** (what it knew at birth) and **growth**
(what it learns by playing). The seed is produced offline at birth by a
**two-stage persona compile**: Stage 1 generates a knowledge-calibrated
**worldbrief** (RAG over the clean wiki + Opus, sized by a `game_familiarity`
dial); Stage 2 (the existing best-of-N **cook**) takes `persona + worldbrief`
and writes the sealed prose card. The seed is **not a ceiling** — it *seeds*
the host's mutable memory store. From then on the host learns experientially
(play → episodes → distillation), and at runtime its LLM prompts are grounded
by **its OWN memory** (per-host, host_id-scoped), with the recall slice
tilting from seed-facts toward lived episodes as the host gains experience.

**Where reality stands:** the growth half of that paragraph is built end to
end; the seed half is built only from Stage 2 onward. Every host today is
born with zero worldbrief — it is 100% growth.

---

## 1. The two-stage persona compile (birth, offline)

```
persona (+ game_familiarity)
   │  Stage 1 — KNOWLEDGE:  RAG (clean wiki) + Opus (1 call)  →  worldbrief   [DESIGN — not built]
   ▼                          "what you know about Gielinor", sized to familiarity
persona + worldbrief
   │  Stage 2 — COOK:  Opus best-of-N + judge  →  sealed prose card           [BUILT]
   ▼
seal BOTH worldbrief + card into the Cornerstone (immutable, + generation_meta)
```

**Stage 2 is BUILT** (`docs/persona-compile.md` is the canonical doc):
`persona/compile.go` carries the `ProseCook` seam, the `DeterministicCook`
no-LLM floor, and `Cooked`/`CookMeta` (the audit shape); `mesa/personacook`
is the best-of-N Opus cook + judge. It already takes a `WorldBrief` argument
— but today that is the hand-filled struct
`{Setting, Home, Activity, Entities}` (`persona/compile.go:40`), and
`mesa/personacook` feeds it a hard-coded `lumbridgeBrief()`
(`mesa/personacook/main.go:168`). Nothing generates a brief, and nothing
seals one into the Cornerstone.

**Stage 1 is DESIGN (not built)** — this remains the live spec for TODO
**M-1**:

- **Stage 1 retrieval strategy (the real engineering):** derive queries from
  the persona's interests (curiosity flavors, north-star, quirks); retrieve
  from the clean corpus; **breadth/depth scaled by `game_familiarity`**
  (newbie → a few starter-area chunks, shallow; veteran → broad survey + deep
  on its specialty). Opus synthesizes a believable "what you know" brief from
  the retrieved facts.
- **Asymmetric spend:** Stage 1 = a *single* Opus call (RAG grounds it, less
  subjective). Stage 2 keeps best-of-N + judge (the card is the permanent,
  subjective, high-leverage artifact). Don't pay best-of-N twice.
- **Believability + misconception dial:** newbie brief = sparse, fuzzy,
  optionally one plausible *wrong* belief; veteran = broad + accurate.
  Feeding it to the cook is what makes the *card itself* read
  naive-about-the-world vs worldly.
- **NOT leading:** the brief grounds (real places/items exist, rendered
  correctly), it does not steer. No `home`/locale is assigned; a host's place
  in the world emerges from its own story. (`WorldBrief` refactors from the
  hand-filled struct → a generated prose artifact + its own generation_meta.)

## 2. Two dials that govern knowledge

- **`game_familiarity`** (DESIGN — no such field exists in `persona/`) — the
  host's *starting* knowledge level (where Stage 1 sizes the seed). Lands
  with Stage 1 (M-1). Distinct from below.
- **`bulk_apperception`** (BUILT — `persona/persona.go:100`, dial vocabulary
  in `persona/enums.go`) — "intelligence as learning-from-experience": how
  *fast* the host consolidates experience into knowledge. This growth-rate
  role is already real: it modulates the Tier-2 insight-cron escalation gate
  (`mesa/mesad/cron_insight.go` `shouldProcessEscalation` — admission drive
  `0.6·bulk + 0.4·curiosity-flavor`; a sharp/curious host chews on more, an
  oblivious one genuinely forgets the nuance).
- Together: a host starts at familiarity X and grows at rate Y. (Dolores arc
  = low familiarity × mid-high bulk apperception → starts naive,
  learns/awakens fast.)

## 3. Seed vs Growth — where they live, how they're used in prompts

- **Seed** (DESIGN) = birth worldbrief. Immutable, sealed in the
  **Cornerstone**. Used WHOLE exactly once — to ground the Stage-2 cook. Also
  *initializes* the memory store (its facts become the host's first memory
  entries). None of this exists: nothing writes a worldbrief or the
  Cornerstone `Pinned` summaries into the episode store (verified — `Pinned`
  is consumed only by the card paths, `persona.Project()` and
  `persona/render.go`).
- **Growth** (BUILT) = everything learned by playing. Mutable, and it canNOT
  live in the frozen Cornerstone — it lives in the memory system: the local
  episodic journal (`memory/journal.go`, bbolt-durable via
  `memory.Manager`) mirrored up to mesa's Postgres LTM
  (`runtime/memory_journal.go` `memoryCapture` → `mirrorEpisode` →
  `Journal.Remember`), plus the salience-gated observation firehose
  (`runtime/observation.go` `emitObservation` → `RecordObservations`) that
  the mesa crons distil into the graded world-knowledge ledger and intention
  graph (`mesa/mesad/cron.go`, `cron_insight.go`; see
  `docs/world-knowledge-and-learning.md`).
- **In runtime prompts neither is dumped wholesale.** The built Act prompt
  (`mesa/mesad/act.go`): a prompt-cached DSL-manual block + the persona prose
  (uncached), then a Situation that carries the host-assembled hints
  (`runtime/director_situation.go` `situation()`). The design's two voices
  manifested there:
  - **"what you know"** → the 📚 `known` hint — graded, provenance-tagged
    beliefs from the distilled knowledge ledger (plus ❓ explicit unknowns).
    Today this voice is 100% earned-by-play; the seed will join it when
    Stage 1 + seed-init land.
  - **"what you remember / have lived"** → the 🧠 `memory` hint — the
    journal's salient episodes (`memoryHint` → `journal.Salient(8)`,
    importance×recency blend, no embeddings).
- **The center of gravity migrates seed → growth** as the host plays — the
  design intent stands, but it is moot until a seed exists: today every
  prompt is all growth.

## 4. Per-host memory RAG (the runtime recall mechanism) — BUILT

**Voyage is a stateless embedder — it stores nothing.** No namespacing in
Voyage (VOYAGE_AI_PROJECT_ID = billing only). All per-host scoping is in OUR
store. The client is `mesa/embed` (`voyage-3`, 1024-dim; mesa-side only — it
holds `VOYAGE_AI_KEY`, wired in `mesa/cmd/mesad/main.go`; unset ⇒ recall
degrades to full-text + recency).

**Two distinct vector stores, two scopings** (as designed):
- **Wiki corpus** — global, shared, ONE index, no per-host namespacing.
  Offline tooling only (seeding briefs). `mesa/wikirag` (gob-cached
  embeddings + cosine search over the clean `cognition/corpus` chunks),
  BUILT, out-of-band — never a host runtime resource.
- **Host memory** — per-host, scoped by `host_id`, runtime recall. Each host
  recalls only its own life (every LTM query is `WHERE host_id = $1`; the
  host_id comes from the authenticated gRPC identity).

**Pipeline (mesa-side, `mesa/mesad/ltm.go`):**
- *Ingest:* host bus event → selective `memoryCapture` → `Journal.Remember`
  (client-stream) → `LTM.Add` embeds the text (Voyage, input type
  "document") and stores it in the `episodes` table, deduped by a
  content-derived idempotency key (`ON CONFLICT`). `"kv"` mirror rows are
  never embedded; a failed/absent embedding stores the row anyway (NULL
  embedding — lexical recall still finds it, and a later at-least-once
  resend backfills the missing vector).
- *Recall:* `Knowledge.Recall` embeds the query (input type "query") →
  `recallCosine`: `WHERE host_id=$1 AND kind <> 'kv' ORDER BY embedding <=>
  $q LIMIT k`. Empty query or no embedder → `recallLexical` (Postgres
  full-text + recency). Empty-query-recency is load-bearing: it is the
  cold-start path by which mesa can boot a host's journal from nothing
  (`runtime/memory_journal.go` `bootstrapJournalFromMesa`).

**Schema as built** (design called it `host_memory`, hash-partitioned,
`salience`; reality is the noted alternative):
```sql
CREATE TABLE episodes (
  host_id text NOT NULL, idem_key text NOT NULL,
  kind text NOT NULL, body text NOT NULL,
  importance double precision NOT NULL DEFAULT 0, entity text NOT NULL DEFAULT '',
  occurred_at timestamptz NOT NULL, created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (host_id, idem_key)
);  -- + embedding vector(1024), HNSW vector_cosine_ops index, (host_id, occurred_at DESC) index
```
No `PARTITION BY HASH (host_id)` — the design itself noted "plain
`WHERE host_id` + HNSW also works (pgvector iterative scan)", and that is
the shape that shipped. Revisit partitioning only if filtered-HNSW recall
degrades at fleet scale. Ranking is **pure cosine** today, not the designed
`similarity × recency × salience` (→ TODO **M-2**/**C-15**).

**Two recall tiers (not everything is RAG)** — BUILT as designed:
- *Local hot* (no mesa, no embeddings): the journal's salient-recency slice
  (`memory/journal.go` `Salient`) + the **trust ledger** by exact name
  (`limbic/ledger.go`; `relation_with(name)` answers from it first —
  "what do I know about player X" is a name lookup, NOT a vector search).
  The relationship ledger is NEVER embedded — the `relationships` table has
  no embedding column.
- *Semantic* (mesa): vector RAG over the full LTM — the deep "what in my
  whole history is relevant to this." This is `mesaclient.Recall`
  (`Knowledge.Recall`), consumed via the cold-start journal bootstrap,
  `memory.Remote.Search` (`mesa/client/adapters.go` `AsRemote`), and the
  cognition retriever adapter (`AsRetriever`).

**Boundary holds** (built as designed): RAG over the host's OWN memory
(scoped to host_id) is allowed — it's the host remembering its life, core to
cognition. Live GLOBAL wiki query at runtime is NOT (no firehose); the wiki
only seeds briefs offline.

## 5. Phasing — what actually shipped

The designed Phase A (per-host in-memory `map[host_id] → []indexedChunk`,
"no Postgres yet") was **skipped entirely**: the build went straight to
Phase B's Postgres + pgvector (`mesa/mesad/ltm.go`), minus the hash
partitioning (see §4). There is no interim in-memory host-memory store and
none is planned.

## 6. What exists vs what's left

- **BUILT:** the mesa LTM (episodes + embed-on-write + cosine `Recall`,
  `mesa/mesad/ltm.go`); the Voyage client (`mesa/embed`); the host↔mesa
  mirror seams (`runtime/memory_journal.go`, `runtime/observation.go`,
  `mesa/proto/mesa.proto` `Journal`/`Knowledge` services); cold-start
  bootstrap (episodes, goal, relationships, knowledge, goal graph); the
  distillation crons (`mesa/mesad/cron.go`, `cron_insight.go`); the trust
  ledger (`limbic/ledger.go`); clean wiki extraction + offline wiki RAG
  (`cognition/corpus`, `mesa/wikirag`); the Stage-2 cook + deterministic
  floor (`persona/compile.go`, `mesa/personacook`).
- **LEFT (DESIGN → TODO M-1):** Stage 1 worldbrief generator
  (RAG-query-from-persona + Opus, familiarity-scaled); chain Stage 1 →
  Stage 2; the `game_familiarity` field; the `WorldBrief` refactor (struct →
  generated artifact); seed→memory initialization.

## 7. Open items

All open work and open questions from this record live in
[`docs/TODO.md`](../TODO.md) — **M-1** (the Stage-1 build set), **M-2**
(misconception dial, Stage-1 query derivation, recall ranking weights),
**C-15** (recall-ranking research), **D-8** (merge settled parts into
`docs/memory.md` + `docs/persona-compile.md`).
