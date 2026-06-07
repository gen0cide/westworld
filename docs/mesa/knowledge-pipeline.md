# Decision Record: Host Knowledge — seed → growth, and per-host memory RAG

> STATUS: DESIGNED (2026-06-06, Alex + Claude). Scope: how a host *knows* things
> about the game — the two-stage persona compile that produces its starting
> knowledge, how that knowledge grows by playing, and how it is recalled into
> runtime LLM prompts. Companion to persona-compile.md (the prose cook),
> decision-transport-storage.md (mesa), and westworld-wiki-rag (the offline
> wiki RAG tool). Most of this lands mesa-side; the seams + the embedding
> machinery already exist.

---

## 0. The shape in one paragraph

A host's game knowledge has a **seed** (what it knew at birth) and **growth**
(what it learns by playing). The seed is produced offline at birth by a
**two-stage persona compile**: Stage 1 generates a knowledge-calibrated
**worldbrief** (RAG over the clean wiki + Opus, sized by a `game_familiarity`
dial); Stage 2 (the existing best-of-N **cook**) takes `persona + worldbrief` and
writes the sealed prose card. The seed is **not a ceiling** — it *seeds* the
host's mutable memory store. From then on the host learns experientially (play →
episodes → reflection), and at runtime its LLM prompts are grounded by **RAG over
its OWN memory** (per-host, mesa-side), with the recall slice tilting from
seed-facts toward lived episodes as the host gains experience.

---

## 1. The two-stage persona compile (birth, offline)

```
persona (+ game_familiarity)
   │  Stage 1 — KNOWLEDGE:  RAG (clean wiki) + Opus (1 call)  →  worldbrief
   ▼                          "what you know about Gielinor", sized to familiarity
persona + worldbrief
   │  Stage 2 — COOK:  Opus best-of-N + judge  →  sealed prose card
   ▼
seal BOTH worldbrief + card into the Cornerstone (immutable, + generation_meta)
```

- **Stage 1 retrieval strategy (the real engineering):** derive queries from the
  persona's interests (curiosity flavors, north-star, quirks); retrieve from the
  clean corpus; **breadth/depth scaled by `game_familiarity`** (newbie → a few
  starter-area chunks, shallow; veteran → broad survey + deep on its specialty).
  Opus synthesizes a believable "what you know" brief from the retrieved facts.
- **Asymmetric spend:** Stage 1 = a *single* Opus call (RAG grounds it, less
  subjective). Stage 2 keeps best-of-N + judge (the card is the permanent,
  subjective, high-leverage artifact). Don't pay best-of-N twice.
- **Believability + misconception dial:** newbie brief = sparse, fuzzy, optionally
  one plausible *wrong* belief; veteran = broad + accurate. Feeding it to the cook
  is what makes the *card itself* read naive-about-the-world vs worldly.
- **NOT leading:** the brief grounds (real places/items exist, rendered
  correctly), it does not steer. No `home`/locale is assigned; a host's place in
  the world emerges from its own story. (`WorldBrief` refactors from the
  hand-filled struct `{Setting,Home,Activity,Entities}` → a generated prose
  artifact + its own generation_meta.)

## 2. Two dials that govern knowledge

- **`game_familiarity`** (NEW band, newbie→veteran, archetype-seeded) — the host's
  *starting* knowledge level (where Stage 1 sizes the seed). Distinct from below.
- **`bulk_apperception`** (existing) — "intelligence as learning-from-experience":
  how *fast* the host consolidates experience into knowledge (the growth rate).
- Together: a host starts at familiarity X and grows at rate Y. (Dolores arc =
  low familiarity × mid-high bulk apperception → starts naive, learns/awakens fast.)
- OPEN: is `game_familiarity` a persona field, or a runtime knob passed to Stage 1
  (so one persona can be cooked at several levels for experiments)? Lean: a persona
  field (seals with the host) that is *overridable* as a knob for demos.

## 3. Seed vs Growth — where they live, how they're used in prompts

- **Seed** = birth worldbrief. Immutable, sealed in the **Cornerstone**. Used
  WHOLE exactly once — to ground the Stage-2 cook. Also *initializes* the memory
  store (its facts become the host's first memory entries).
- **Growth** = everything learned by playing. **Mutable**, lives in the memory
  system (hostkv local hot-slice + mesa LTM), grown by the limbic→episode→
  reflection pipeline. The Cornerstone is frozen, so growth canNOT live there.
- **In runtime prompts neither is dumped wholesale** — they merge into one memory
  and the prompt gets a situation-relevant **retrieved slice** (RAG over the
  host's own memory). They enter as two voices:
  - seed / consolidated facts → **"what you know"** (background world model)
  - episodic growth → **"what you remember / have lived"** (first-person, vivid,
    provenanced — *"the last stranger who traded with you scammed you"*).
- **The center of gravity migrates seed → growth** as the host plays: early
  prompts lean on seed (general awareness); later ones are dominated by lived
  episodes. The host becomes its memories.
- Runtime Cognition prompt assembly (mesa-side):
  `[cached] BrainCard + allowed DSL surface · [dynamic] affect line + recalled
  memory (semantic "know" + episodic "remember") + situation/options`.

## 4. Per-host memory RAG (the runtime recall mechanism)

**Voyage is a stateless embedder — it stores nothing.** No namespacing in Voyage
(VOYAGE_AI_PROJECT_ID = billing only). All per-host scoping is in OUR store.
Voyage is the *shared* embedder for both the wiki and every host's episodes.

**Two distinct vector stores, two scopings:**
- **Wiki corpus** — global, shared, ONE index, no per-host namespacing. Offline
  tooling only (seeding briefs). (`mesa/wikirag`, built.)
- **Host memory** — per-host, namespaced by `host_id`, runtime recall. Each host
  recalls only its own life.

**Pipeline (mesa-side; reuses wikirag's Voyage+search code):**
- *Ingest:* limbic emits an episode → Mirror channel (`Journal.Remember`) → mesa
  embeds (Voyage) → stores in that host's partition. Birth seed embedded at birth.
- *Recall:* on a Cognition call, mesa embeds the situation → vector search
  `WHERE host_id = X` → rank by `similarity × recency × salience` → top-K → prompt.

**Namespacing in Postgres/pgvector (proper + standard):**
```sql
CREATE TABLE host_memory (
  id bigserial, host_id text NOT NULL, kind text,  -- seed|episodic|semantic
  text text, embedding vector(1024), salience real, occurred_at timestamptz
) PARTITION BY HASH (host_id);          -- each host's rows in one partition
-- HNSW index per partition; recall: WHERE host_id=$1 ORDER BY embedding <=> $q LIMIT k
```
Hash-partition by `host_id` so each query prunes to one host's small slice
(sidesteps filtered-HNSW recall issues; scales to 500 hosts × millions of
vectors). Plain `WHERE host_id` + HNSW also works (pgvector 0.8 iterative scan).

**Two recall tiers (not everything is RAG):**
- *Local hot* (no mesa, no embeddings): recent episodes + the **trust ledger**,
  recalled by recency / exact-key (e.g. "what do I know about player X" = a name
  lookup, NOT a vector search). Serves the pearl fast path. The relationship
  ledger is NEVER embedded.
- *Semantic* (mesa): vector RAG over the full LTM — the deep "what in my whole
  history is relevant to this." This is the `mesaclient.Recall` call.

**Boundary holds:** RAG over the host's OWN memory (scoped to host_id) is allowed
— it's the host remembering its life, core to cognition. Live GLOBAL wiki query at
runtime is NOT (no firehose); the wiki only seeds the brief offline.

## 5. Phasing
- **Phase A (now / few hosts):** per-host memory = an in-memory `map[host_id] →
  []indexedChunk` (the wikirag gob pattern, keyed by host). No Postgres yet.
- **Phase B (scale):** Postgres + pgvector, hash-partitioned by `host_id`.

## 6. What exists vs what's left
- EXISTS: the recall seam (`memory.Remote.Search` / `mesaclient.Recall`); the
  Voyage embed+cosine machinery (`mesa/wikirag`); clean wiki extraction; the
  limbic episode stream + trust ledger; the cook (Stage 2) + deterministic floor.
- LEFT: Stage 1 worldbrief generator (RAG-query-from-persona + Opus, familiarity-
  scaled); chain Stage 1 → Stage 2; `game_familiarity` field; `WorldBrief` refactor
  (struct → generated artifact); mesa per-host memory store + episode ingest
  wiring; seed→memory initialization.

## 7. Open questions
- `game_familiarity`: persona field vs runtime knob (lean: field, overridable).
- Misconception dial: how much newbie-fuzziness to allow (grounding vs believability).
- Stage-1 query derivation: exact mapping from persona interests → retrieval queries.
- recall ranking weights (similarity/recency/salience) — tune empirically.
- Cross-reference: merge the settled parts into canonical docs/memory.md +
  docs/persona-compile.md when built.
