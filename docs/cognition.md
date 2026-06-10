# Cognition — recognition, retrieval, and the knowledge ledgers

> **STATUS: BUILT** (verified 2026-06-10 against branch HEAD `93d3cd1`).
> `cognition/` is four shipped subpackages plus the retrieval seam: **resolve**
> (recognition faculty), **corpus** (wiki keyword corpus + namespace federation),
> **knowledge** (Beta-belief world-knowledge ledger), **goalgraph** (intention
> graph) — and the root `Client`/`Bundle` interface, wired in production to the
> real mesa `Knowledge.Recall` RPC via `mesaclient.AsRetriever`
> (`runtime/runhost_bootstrap.go:110`). The original "cognition =
> `PrepareDecision`/`DecisionRequest`/`DSLSurface`" design never manifested as
> written; the per-turn assembly shipped instead as `MesaDirector.situation()` +
> mesad's `actPrompt`. The original design body is archived:
> [archive/initial-brainstorming/cognition-original-design.md](archive/initial-brainstorming/cognition-original-design.md).

## What cognition actually is

The `cognition/` tree is the host's **mind-access layer**: the faculties that turn
loose text, raw events, and goals into structured things the rest of the host can
read. It does not decide (the brain/mesa Act loop does) and it does not call the
LLM directly (seams are injected).

| Package | What it is | Key files |
|---|---|---|
| `cognition` (root) | retrieval seam: `Client` interface (`Retrieve(ctx, Retrieval) -> *Bundle`) + deterministic `StubClient`; leaf, stdlib-only | `cognition/cognition.go`, `cognition/stub.go` |
| `cognition/resolve` | recognition: loose player text → ranked canonical defs from the facts registry | `resolve/resolve.go`, `catalog.go`, `alias_store.go`, `brain.go` |
| `cognition/corpus` | indexed text chunks (rsc.wiki dump) + audience-namespace gate | `corpus/corpus.go`, `wiki.go`, `namespace.go` |
| `cognition/knowledge` | semantic world-knowledge ledger — Beta(α,β) beliefs about THINGS, with provenance | `knowledge/knowledge.go` |
| `cognition/goalgraph` | intention graph — goals/sub-goals/open-questions as nodes, typed dependency edges | `goalgraph/goalgraph.go` |

`knowledge` and `goalgraph` are two of the three structural legs of the mind; the
third is the trust ledger (`limbic/ledger.go` — Beta(α,β) evidence about PEOPLE).
All three persist on the same memory-namespace spine (local bbolt tier + mesa
mirror): `runtime/knowledge.go` (`knowledge:` namespace), `runtime/goalgraph.go`
(`goalgraph:`), `runtime/limbic.go` — warm-start on boot, snapshot-flush on the
limbic cadence + on exit, synced up via `SyncKnowledge`/`SyncGoalGraph` and pulled
back via the `Fetch*` twins (`mesa/client/client.go`).

## The retrieval seam — real mesa Recall

The root package ships the contract (`Client`, `Retrieval`, `Bundle`) and a
deterministic `StubClient` (the default on a bare host — `runtime/host.go:458`).
Production wiring (`runtime.RunHost`, used by the cradle daemon fleet and
`cmd/host`) swaps in the real thing:

- `mesaclient.AsRetriever(mc, username)` (`mesa/client/adapters.go:39`) adapts the
  mesa gRPC client to `cognition.Client`: `Retrieve` issues `Knowledge.Recall`
  with `Query{HostID, Text: Goal, TopK}` and returns episode texts in
  `Bundle.Episodic`.
- Server-side, mesad serves `Recall` from the LTM (`mesa/mesad/ltm.go`): the
  Postgres `episodes` table, scoped by the **authenticated** host_id (a host can
  only read its own past). With a Voyage embedder wired (`VOYAGE_AI_KEY`,
  `cmd/mesad/main.go:119`) episodes are embedded on write and `Recall` ranks
  by **pgvector cosine similarity**; without one it degrades to Postgres full-text
  relevance + recency; an empty query returns pure recency (the cold-start /
  session-genesis bootstrap path).

## Recognition — `resolve()`

`cognition/resolve` is the faculty behind the DSL's `resolve(text, kind?)` /
`resolve_one(...)` (registered in `runtime/dsl_actions.go`; wired as
`Host.Resolver`, `runtime/host.go:265`). Pipeline, in priority order
(`resolve/doc.go`):

1. **Learned-alias store** — a persisted per-host text→canonical table
   (`aliases.json` under the host data dir, `runtime/runhost_bootstrap.go:600`).
   Each host grows its own lingo; no curated slang ships. Deterministic fast path.
2. **Conservative fuzzy/token match** over the facts catalog — items, npcs,
   scenery, boundaries, spells, prayers (`resolve/catalog.go`), ranked best-first.
3. **Brain (LLM) fallback hook** (`resolve/brain.go`) — maps loose text only to a
   canonical name it is shown, never invents ids; a verified mapping is written
   back to the alias store. The seam exists, but **both production wirings pass a
   nil Brain** today (`runtime/runhost_bootstrap.go:68`,
   `cmd/legacy-cradle/main.go:222`) — live recognition is alias-store + fuzzy.

Definitions are always looked up exactly in `facts`; only name→canonical
recognition is fuzzy, and ids come from the resolved def, never from the model.

## Corpus — `recall()` over the wiki dump

`cognition/corpus` is the Phase-2.6 keyword-retrieval substrate: `LoadWikiDump`
extracts the rsc.wiki HTML crawl (goquery, chunked by h2/h3 heading) into an
in-memory `MemoryCorpus` (~11 MiB) scored by term frequency. The **namespace
federation** (`corpus/namespace.go`) is a load-time audience gate: `Gameplay`
sources (anything a real player could have read) vs `Dev` sources (server source,
decompiles, AutoRune) — production cradles never instantiate Dev sources at all,
so leakage is structurally impossible. The Dev side is built but unused
(docs/TODO.md M-8).

Wiring reality:

- Only `cmd/legacy-cradle -wiki-dump <dir>` loads a corpus
  (`cmd/legacy-cradle/main.go:793`). The cradle-daemon fleet (`runtime.RunHost`)
  does **not** wire `Host.Corpus`, so `recall()` there routes to the mesa
  Retriever instead.
- The DSL `recall(query, top?)` (`runtime/actions_memory.go:97`) prefers
  `h.Corpus` (provenance-tagged chunk strings), else falls back to
  `h.Retriever`. `relation_with(name)` prefers the host's own trust ledger
  (`h.ledger.Rel(name).Grade`) before asking the Retriever.
- Embedding the wiki is **out-of-band tooling**, not host runtime: `mesa/wikirag`
  is the offline Voyage embed/search CLI over the same chunks. Per the
  discover-don't-seed stance (docs/world-knowledge-and-learning.md), live hosts
  do not get wiki RAG; curated knowledge arrives via the persona/knowledge
  pipeline (docs/mesa/knowledge-pipeline.md; docs/TODO.md M-1).

The location-blind storage verbs `remember`/`recollect`/`forget`
(`runtime/actions_memory.go`) are the memory layer's surface, not cognition's —
they route through the tiered memory Manager (see [memory.md](memory.md)).

## The knowledge ledger — graded beliefs about THINGS

`cognition/knowledge` mirrors the trust ledger's shape for the non-social world:
each subject (npc, place, shop, item, mechanic, quest) accrues **Beta(α,β)-backed
beliefs** (`Confidence() = α/(α+β)`) each tagged with **provenance** — `system` >
`observed` > `deduced` > `hearsay` — so the host can tell knowing from guessing
and a stronger source upgrades a restated claim. API: `Note` (new claim at an
initial confidence), `Observe` (evidence for/against), `Seen` (familiarity bump),
`Tag`, `Known`/`Get`/`All`. The in-RAM copy stays host-light: caps of 512
subjects / 24 beliefs per subject, deterministic `Prune` on the flush tick;
`Import` (cold-start replace) and `ImportMerge` (warm fold of the
server-reconciled copy — max-evidence per claim, never clobbering a live row).

- **Live writers**: `runtime/perception.go` folds raw game events into the ledger
  deterministically (O(1) per event, no LLM) — e.g. an opened shop catalogue
  becomes `Note(shop, "shop", claim, ProvObserved, 0.9)` — and streams salient
  observations up to mesa (`runtime/observation.go`).
- **Growth**: mesad's distillation crons own the heavy work —
  `mesa/mesad/cron.go` (Tier-0 no-LLM dedup/GC; Tier-1 batched Haiku claim
  extraction) and `mesa/mesad/cron_insight.go` (Tier-2 rare Sonnet:
  provenance-weighted contradiction reconciliation + cross-entity chaining into
  goal-graph edges).

## The goal graph — intention as memory, not search

`cognition/goalgraph` records goals/sub-goals/open-goals/open-questions/states as
nodes and typed edges (`requires` / `produces` / `enables` / `blocked_by` /
`serves`), capped at 64 nodes in RAM. It is deliberately an **accreting memory
graph the host reads**, not a planner the host must solve per tick. Writers are
the MesaDirector's deterministic hooks (`runtime/director_goals.go`): the standing
objective becomes a root node; a hard fail-streak / stall / spin marks the goal
`blocked` and mints an `open_question` node linked `blocked_by` (driving the
ask/forage drives); a real success un-blocks; an acquire-goal already satisfied by
inventory is marked `done` with no LLM involved. The planner's `goal_op`
(done/abandoned/adopt) rides back on every Act move.

## Where retrieval assembly actually happens

The archived design put assembly in a host-side `PrepareDecision`. What shipped
splits it across the seam:

- **Host side — `MesaDirector.situation()`** (`runtime/director_situation.go:25`)
  builds `mesaclient.Situation` per think-turn: world snapshot (position, HP,
  fatigue, inventory, deduped NPCs, players), the rolling cross-turn transcript
  (last 18 lines), the affect snapshot, and a hints map — scene perception with
  coordinates (`describeArea`, radius widened when stuck), the latest server
  message, open dialog options, a player directive annotated with the trust
  ledger's read (`trustNote`), last action/result/DSL, the durable-journal memory
  hint, nearby players + the session-genesis attention ladder, and the **epistemic
  block** (`epistemicHints`, `runtime/director_situation.go:399`): a bounded
  intention-graph slice (blockers/sub-goals/downstream value), relevant graded
  beliefs ("known"), explicit unknowns, and a curiosity-weighted learning
  priority — read + surface only, hard caps throughout.
- **Mesa side — `actPrompt`** (`mesa/mesad/act.go:641`) renders the Situation
  into the Act prompt (intention structure first, then feedback / scene / memory /
  known / unknowns / transcript / goal-lifecycle instructions), with the DSL
  manual riding as a prompt-cached system block (`act.go:31`).

The cost stance from the original design survives in mesad's model tiering: the
bulk distillation runs on Haiku, the rare insight pass on Sonnet, never Opus on
bulk (`mesa/mesad/cron.go`, `cron_insight.go`).

## Three knowledge sources, three access patterns

| Source | Access pattern | Storage |
|---|---|---|
| **Structured world facts** | Programmatic key→value lookup | Local Go maps over OpenRSC JSON exports |
| **Unstructured RAG corpus** | Semantic similarity, top-K | Mesa (pgvector + Voyage 3) |
| **Per-host memory** | Mixed: structured (relations) + semantic (episodes/reflections) | Mesa |

Row-by-row, as built: structured facts = the `facts` package (below); the RAG row
= mesad's LTM (pgvector + Voyage on the episodes table) with the wiki corpus
keyword-only and offline (`mesa/wikirag`); per-host memory = the tiered memory
Manager + the three ledgers above, mirrored to mesa.

### Structured world facts

OpenRSC's def files are the source of truth for game-world data, and they are now
compiled **into the binary**: `cmd/defsgen` generates `facts/static_defs_gen.go` /
`static_locs_gen.go` from the OpenRSC tree (`GameObjectDef.xml`, `DoorDef.xml`,
`NpcDefs.json`, `ItemDefs.json`, `TileDef.xml`, `locs/{SceneryLocs,BoundaryLocs,
NpcLocs,GroundItems}.json`), and `facts.LoadStatic()` builds the indexes with zero
file I/O — 794 NPC defs, 1290 item defs, plus the embedded gazetteer, spells, and
prayers. Lookups are exact: `resolve("r2h")` lands on the Rune 2-handed Sword def;
item views expose def-derived fields like `edible` (`runtime/views_inventory.go:356`).
No LLM, no network. Free. (The `ItemEdibleHeals`/herb XMLs the original design
listed were never loaded; remaining def-side codegen — resources, recipes,
prayers/spells generation — is docs/TODO.md MP-8.)

This is critical: a huge fraction of "what would an experienced RSC player know" is just data lookup. We don't need the LLM to recall it; we hand it the answer.

### Per-host memory

Episodic, relational, reflective — each with its own access pattern, all real:
the journal + episodes (`memory.Journal` → mesa LTM, semantic recall), the trust
ledger (`limbic/ledger.go`, local-first with mesa relationship sync), and the
knowledge ledger above. See [memory.md](memory.md) for the full architecture.

## What cognition does NOT do

- Doesn't make decisions — the brain/mesa Act loop does
  (`runtime/mesa_director.go`, `mesa/mesad/act.go`).
- Doesn't call the LLM directly — `resolve`'s fallback is an injected `Brain`
  seam (nil in production today); retrieval is a gRPC to mesa.
- Doesn't solve — `knowledge` and `goalgraph` are bounded, accreting memory
  structures the host reads and surfaces; traversal is one BFS hop with hard caps
  (`epistemicHints`), never a search.
- The root package stays a stdlib-only leaf; the heavy growth (claim extraction,
  reconciliation, graph chaining) is mesa's job, not the host's.

## Backlog

All open work lives in [TODO.md](TODO.md) — see §3 Cognition: **C-14** carries
this doc's former open questions (retrieval query generation, cross-source prompt
budgeting); also **C-9** (goal-graph-aware Act loop), **C-12** (progressive
disclosure / `DSLSurface` — NEEDS OPERATOR), **C-17** (ledger GC / cron cadence),
**C-24** + **DSL-24** (trust-ledger remainder + DSL read surface), **M-8**/**M-9**
(corpus-ingest residue).
