> **ARCHIVED (initial brainstorming), 2026-06-10.** This is the original "cognition =
> retrieval orchestration" design. The `PrepareDecision` / `DecisionRequest` /
> `DSLSurface` assembly below was never built as written. What shipped instead: the
> per-turn assembly is `MesaDirector.situation()` (`runtime/director_situation.go:25` —
> world snapshot + affect + rolling transcript + scene/epistemic hints), rendered
> server-side into the Act prompt by `actPrompt` (`mesa/mesad/act.go:641`); the
> trust-ledger projection manifested as `limbic/ledger.go` plus the
> `player_directive`/`nearby_players` trust annotations in `situation()`; retrieval
> is the real `Knowledge.Recall` gRPC (`mesa/mesad/ltm.go`, pgvector + Voyage when
> an embedder is wired), not `GET /knowledge`; the `DSLSurface`/progressive-disclosure
> mechanism was never built at all (`docs/TODO.md` C-12, NEEDS OPERATOR — the shipped
> DSL manual discloses the full surface). The living doc is `docs/cognition.md`
> (rewritten same-commit). Kept verbatim below for the record.

# Cognition — retrieval and knowledge

> **STATUS: PARTIAL — assembly built, retrieval still stubbed** (updated 2026-06-07).
> The retrieval seam is now wired to **mesa**: the host reaches `Knowledge.Recall`
> via `mesaclient.AsRetriever`, and the per-turn **Situation** assembly (this doc's
> "DecisionRequest") is implemented for real in `runtime/mesa_director.go`
> (`situation()` — world snapshot + affect + rolling narrative transcript + scene/
> dialog/last-action hints), fed to `mesa.Act`. The `cognition/` package still ships
> the `Client` interface + the deterministic `StubClient` (offline fallback). STILL
> stubbed: `Knowledge.Recall` retrieval itself (mesa returns empty), vector search,
> and RAG. See `cognition-and-autonomy.md` §5 (memory) and §2 (session genesis).
> [Original stub note follows.]
> The package ships a `Client` interface (`Retrieve(ctx, Retrieval) ->
> *Bundle`) plus a deterministic `StubClient` that returns
> hand-crafted `Bundle`s keyed off substring matches in
> `Retrieval.Goal`.
> The real types in code are `Bundle`, `Retrieval`, and `Client`
> (see `cognition/cognition.go`); the stub is `cognition/stub.go`.
> cognition is a leaf package (stdlib-only). The two subpackages
> that *do* contain real working code are `cognition/resolve/`
> (the `resolve()` learned-alias store, shipped) and
> `cognition/corpus/` (knowledge corpus scaffolding). Read the
> banner below as the contract the Phase-3 mesa client will fill,
> not as implemented behavior. Where this doc names `DecisionRequest`,
> mesa endpoints (`GET /knowledge?q=...`), or Sonnet-vs-Haiku cost
> routing, treat it as **ASPIRATIONAL / Phase 3+**.

## What cognition does

The `cognition` package on the cradle side is the retrieval orchestration layer. It sits between the brain and mesa, preparing the contextual information the brain needs to make a decision.

When the brain wants to decide, cognition assembles:
- Structured world facts relevant to the decision (item heal amounts, NPC stats, etc.) from local lookup tables
- Top-K episodic memories relevant to the current situation (vector retrieval from mesa)
- Relational records for nearby players (batch lookup from mesa)
- Reflections currently held (small set, mostly cached)
- RAG chunks from rsc.wiki when the decision class warrants it (Strategic, ScriptGen) — also from mesa
- Routine library index (names + descriptions, not full source)

This composite is the `DecisionRequest` handed to `Brain.Decide`.

## Three knowledge sources, three access patterns

| Source | Access pattern | Storage |
|---|---|---|
| **Structured world facts** | Programmatic key→value lookup | Local Go maps over OpenRSC JSON exports |
| **Unstructured RAG corpus** | Semantic similarity, top-K | Mesa (pgvector + Voyage 3) |
| **Per-host memory** | Mixed: structured (relations) + semantic (episodes/reflections) | Mesa |

### Structured world facts

OpenRSC's JSON/config files are the source of truth for game-world data. We load them at cradle startup into local maps. No RAG needed because lookups are exact:

- `NpcDefs.json` (836 NPCs) — stats, descriptions, drop tables
- `ItemDefs.json` (1593 items) — name, description, members-only flag, equipment slot
- `NpcLocs.json` — NPC spawn coordinates
- `SceneryLocs.json` — scenery placement
- `ItemEdibleHeals.xml` — how much each food heals
- `ItemUnIdentHerbDef.xml` — herb identification thresholds

A function like `cognition.WhatHeals("shark")` returns `20 HP` instantly from a local map. No LLM, no network. Free.

This is critical: a huge fraction of "what would an experienced RSC player know" is just data lookup. We don't need the LLM to recall it; we hand it the answer.

### Wiki RAG

Phase 3.5 (cohort experiment) brings rsc.wiki content into mesa. Process:

1. One-time scrape of rsc.wiki, respecting robots.txt and rate-limiting politely
2. Chunk by section heading (~500-800 tokens per chunk)
3. Filter by F2P/P2P metadata; exclude P2P chunks unless host is on a P2P cohort
4. Embed via Voyage 3 (one-time cost ~$0.20 for the whole F2P corpus)
5. Store in mesa's `knowledge_chunks` table with `source='rsc.wiki'`, `members_only` flag

Retrieval at decision time:

```
GET /knowledge?q=<query>&top=5&source=rsc.wiki
```

The query is generated by cognition based on the current situation — typically a 1-2 sentence description of what the host is trying to figure out. The query is embedded, top-K cosine-similar chunks are returned, those go into the brain prompt.

A/B experiment: spawn cohort `wiki_enabled` and cohort `no_wiki` to measure whether RAG access actually improves agent performance and believability.

### Per-host memory

The bulk of cognition is per-host memory retrieval. Episodic, relational, reflective — each has its own retrieval pattern:

- **Episodic**: vector retrieval by current-situation query, filtered by salience threshold. "What relevant past events am I drawing on for this decision?"
- **Relational**: batch lookup by nearby usernames. "What do I know about these people I can see?"
- **Reflective**: small, mostly cached. "What generalizations have I formed?"

See [memory.md](../../memory.md) for the full memory architecture.

## Composing the `DecisionRequest`

```go
func (c *Cognition) PrepareDecision(ctx context.Context, h *Host, trigger Trigger) (brain.DecisionRequest, error) {
    req := brain.DecisionRequest{
        DecisionClass: classifyTrigger(trigger),
        Persona:       h.Persona,
        World:         h.World.Snapshot(),
        RecentEvents:  h.Memory.Recent(20),
    }
    
    if req.DecisionClass == brain.Strategic || req.DecisionClass == brain.ScriptGen {
        // Knowledge retrieval matters for strategic decisions
        query := buildKnowledgeQuery(h, trigger)
        knowledge, _ := c.mesa.QueryKnowledge(ctx, query, 5)
        req.Knowledge = knowledge
    }
    
    // Always include relational records for nearby players.
    // GetRelationships returns the TRUST-LEDGER projection of each relation
    // (Beta-posterior trust mean + confidence, the derived band, and tags) —
    // see "Relational records = the trust-ledger projection" below.
    nearby := h.World.NearbyPlayers()
    if len(nearby) > 0 {
        rels, _ := c.mesa.GetRelationships(ctx, nearby)
        req.Relations = rels
    }
    
    // Episodic memory: semantic-search the current situation
    epQuery := buildEpisodicQuery(h, trigger)
    episodes, _ := c.mesa.QueryEpisodes(ctx, epQuery, 10)
    req.RecentEvents = append(req.RecentEvents, episodicAsEvents(episodes)...)
    
    // Routine index (always)
    req.RoutineIndex = h.Routines.Index()
    
    // Fold relevant working_scratch entries (rate-limit / dedup counters
    // the deterministic handlers already use — e.g. "asked JimBob about
    // steel bars twice") so the brain reasons with the same scratch.
    req.WorkingScratch = h.Scratch.Relevant(trigger)
    
    // Assemble the per-host DSLSurface for the brain prompt (see
    // "Assembling the DSLSurface" below). Read-side only — cognition reads
    // host.vocabulary; it never writes graduations.
    req.DSLSurface = c.assembleDSLSurface(ctx, h, trigger)
    // DSLGrammar (the static syntax skeleton) is NOT assembled here — it is
    // the globally-cached brain-prompt chunk; only the earned symbol table
    // is per-host.
    
    return req, nil
}
```

The cognition layer is pure orchestration — it doesn't store state itself, it composes from mesa + local sources.

> **ASPIRATIONAL / Phase 3+.** None of the assembly below is built. There is no
> `PrepareDecision`, no `DecisionRequest`, no mesa wiring, no vector retrieval
> over a spec corpus, and no trust ledger in code today — they are design that
> the Phase-3 mesa client and the Phase-4 agent-driver will fill. The shipped
> `cognition` package is the `Client`/`Bundle`/`Retrieval` stub described in the
> banner. Everything here is **read-side assembly**: it preserves the invariants
> in *What cognition does NOT do* — cognition does not decide, does not run the
> LLM, and does not write to mesa.

### Assembling the `DSLSurface`

`PrepareDecision` also assembles the per-host **`DSLSurface`** — the disclosed
verb/accessor symbol table the brain prompt is allowed to script against. This is
the read-side build of the **G2** disclosure mechanism (the assembled scripting
surface). The surface is the dedup union of three pieces, assembled fresh per call:

```
DSLSurface = survival_core  ∪  host.vocabulary  ∪  retrieve_candidates(goal, K)
```

```go
// assembleDSLSurface builds the per-host symbol table for the brain prompt.
// READ-side only: it reads host.vocabulary from mesa and retrieves candidates;
// it never WRITES a graduation (that is the cradle/ingest path's job, below).
func (c *Cognition) assembleDSLSurface(ctx context.Context, h *Host, trigger Trigger) brain.DSLSurface {
    // 1. survival_core — the static allowlist every host gets (the ~10-15
    //    don't-die / move / observe verbs). spec.Tier == TierCore.
    surface := dslspec.SurvivalCore()

    // 2. host.vocabulary — the per-host EARNED symbols, read from mesa
    //    (bot_vocabulary). The vocabulary-growth observable; cognition only READS it.
    earned, _ := c.mesa.GetVocabulary(ctx, h.ID)
    surface = append(surface, earned...)

    // 3. retrieve_candidates(goal, K) — top-K NEW verbs by goal-relevance: a
    //    vector retrieval over the DSL-spec corpus embedded in knowledge_chunks
    //    (source='dsl_spec'), excluding symbols the host already knows. K is the
    //    disclosure-rate cohort knob.
    goal := buildKnowledgeQuery(h, trigger)
    cands, _ := c.mesa.RetrieveSpecCandidates(ctx, goal, names(earned), K)
    surface = append(surface, candidatesAsSymbols(cands)...)

    return dedup(surface)
}
```

**The split — grammar is free; the symbol table is earned.** The static syntax
skeleton (**`DSLGrammar`**: how to shape a `routine`/`proc`, `on`/`when`/`select`,
the bang `!` operator, control flow, plus 1-2 survival-core-only examples) carries
**no verb names** and is the brain-prompt chunk **cached globally, forever**. The
per-host EARNED symbol table (`DSLSurface`) is what `PrepareDecision` assembles
**here, per call**:

| Piece | What it is | Where it lives / caching |
|---|---|---|
| **`DSLGrammar`** | static syntax skeleton, no verb names | brain prompt, cached **globally forever** |
| **`DSLSurface`** | `survival_core ∪ host.vocabulary ∪ retrieve_candidates(goal, K)` | assembled **here, per host, per call**; the stable part (survival_core ∪ host.vocabulary) is **cached per-host, invalidated on vocabulary graduation**; the K candidates are goal-volatile and uncached |

This is **read-side assembly**. cognition still does not write: a verb *graduates*
into `host.vocabulary` (on a successful use) via the cradle/ingest path, **not
here** — that write is what invalidates the per-host surface cache. cognition reads
`host.vocabulary`, reads the embedded spec corpus, and composes; it never grants a
symbol. The candidate disclosure carries the *interface* (name + signature + a
terse doc), never strategy or server-specific mechanics — that grounding stays
withheld. The K candidate retrieval reuses the same mesa RAG machinery as Wiki RAG,
over a one-time-embedded spec corpus (~150 entries) under `source='dsl_spec'`.

> See `host-bootstrap-and-knowledge-gating.md` (the G2 DSLSurface assembly) and
> `brain.md` for the `DSLGrammar`/`DSLSurface` prompt-chunk split.

### Relational records = the trust-ledger projection

The relational batch lookup (`GetRelationships(nearby)`) returns the **trust-ledger
projection** of each relation, not a bare record. Each relation carries:

- the **Beta-posterior trust** — the mean (`Trust() = α/(α+β)`) *and* its confidence
  (the evidence mass — "200 trades vs met once"),
- the **derived band** (stranger / acquaintance / friend / rival / enemy),
- structured **tags** (`rival`, `ally`, `no-steel`, …).

This projection is what feeds the brain prompt's **"players nearby" block**, so a
Phase-2 reply is trust-aware (a warm read for a friend, a wary one for a rival). The
**same records hydrate the host's LOCAL reputation copy** that serves the fast
in-band `trust()` / `trust_confidence()` / `reputation()` / `is_rival()` DSL queries
on the hot path — pure, local reads that need no network per tick.

cognition only **READS / hydrates** the projection. The trust *updates* — the
cooperation/defection classification and the severity weight — are graded and
written elsewhere (the Limbic's provisional grade in-band, the out-of-band
batched-Haiku **TrustGrade** authoritative grade written to mesa). cognition never
grades and never writes a trust delta.

> See `social-graph-and-trust-ledger.md` (the Beta posterior, the TrustGrade flow,
> and the local-copy / mesa-mirror plumbing).

### Folding in `working_scratch`

`PrepareDecision` folds the relevant entries of the host's **`working_scratch`** —
the host-authored rate-limit / dedup counters (e.g. "asked JimBob about steel bars
twice") — into the request, so the brain reasons with **the same scratch the
deterministic handlers already use**. The scratch is local-fast for in-band reads
and async write-through to mesa; cognition reads the mirror into the prompt so a
Phase-2 reply is dedup-aware ("I already asked Bob twice"). This is a read — the
scratch is written by the host's own `cache.set`/`cache.incr`, not by cognition.

> See `mesa.md` (`working_scratch`) and `chat-interruption-and-engagement.md`
> (the scratch-cache local-fast-read / async-write-through pattern).

### Bootstrap short-circuit

During the **Tutorial-Island bootstrap**, `PrepareDecision` is **bypassed
entirely**. The naive direct mode feeds the tutorial's in-game text (system
messages + NPC dialogue — the self-documenting curriculum) straight to the brain
with **no retrieval orchestration**: perceive → raw LLM → act, with the mesa read
path short-circuited. cognition's orchestration (RAG, episodic retrieval, the
`DSLSurface` assembly, the trust-ledger projection) is for the **post-bootstrap
mainland**, not the on-ramp.

> See `brain.md` (bootstrap mode) and `host-bootstrap-and-knowledge-gating.md`
> (§5.2 the naive direct-LLM bootstrap loop).

### Who invokes `PrepareDecision`

`PrepareDecision` is invoked (Phase 4) by the **agent-driver loop** at think-turns,
just before `Strategist.Decide`; the background **`cognitiveLoop`** performs
out-of-band appraisal. The driver only schedules and executes — cognition does the
read-side assembly, the brain decides.

> See `architecture.md` / `layers.md` (the agent-driver loop / cognitiveLoop) and
> `cognitive-architecture-design.md` §1.3.

## Cost-saving via retrieval

The economic argument for cognition is direct:

| Scenario | Cost per strategic call |
|---|---|
| No RAG: model recalls game knowledge from training | Sonnet ~$0.005, often hallucinates |
| With RAG: retrieved authoritative chunks injected | Haiku ~$0.0005, never hallucinates |

The 10x cost reduction comes from being able to **downgrade the model** because retrieved authoritative knowledge replaces what model reasoning had to do. Plus accuracy improves.

This is why cognition is positioned as a peer of brain in the layer cake, not a sub-package: it directly controls 5-10x of the agent's per-call cost.

## What cognition does NOT do

- Doesn't run LLM calls itself (brain does)
- Doesn't make decisions (brain does)
- Doesn't write to mesa (other layers do — events get written by event handlers, episodes get flushed by the memory consolidator)
- Doesn't cache aggressively at the cradle level — mesa-side caching is the main optimization; cognition is mostly stateless transformations

## Open questions

- **Query generation for retrieval**: turning "what's about to happen" into a good vector query is non-trivial. May need an LLM call for query expansion in edge cases, but ideally heuristic templating works for the common cases (e.g., for ScriptGen: query = "how to do X in RSC" where X is extracted from the strategic goal).
- **RAG result caching at cradle**: should we cache the last N retrieval results to avoid re-querying mesa for similar queries within a short window? Probably yes; 5-min TTL.
- **Cross-source ranking**: when both episodic memory and wiki RAG return relevant chunks, how do we order/budget them in the prompt? Probably hard-cap: episodic gets ~50% of context budget, wiki gets ~30%, relations get ~20%, depending on decision class.
