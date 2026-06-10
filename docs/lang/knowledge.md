# Knowledge — the corpus, namespaces, and `recall()`

> **STATUS: BUILT, legacy-cradle only** (verified 2026-06-10 against branch
> HEAD `7f22bb6`). This doc describes the `cognition/corpus` shared-knowledge
> mechanism (Phase 2.6 Slice 1): an in-memory keyword corpus with load-time
> namespace gating, queried by the `recall()` DSL builtin. It is wired **only**
> by `cmd/legacy-cradle` (`-wiki-dump` / `-dev` flags, `loadCorpus` at
> `cmd/legacy-cradle/main.go:793`). The production path —
> `cmd/host` and `cmd/cradle-server` via `runtime.RunHost`
> (`runtime/runhost.go:31`) — **wires no corpus**: `host.Corpus` stays
> nil and `recall()` falls through to the per-host Retriever (see §recall
> below). The production knowledge story is the mesa side instead:
> [`../mesa/knowledge-pipeline.md`](../mesa/knowledge-pipeline.md) (persona
> cook + per-host episodic RAG), the mesa `Knowledge.Recall` RPC
> (`mesa/mesad/server.go:297`, Postgres+pgvector LTM in `mesa/mesad/ltm.go`),
> and the **offline** `mesa/wikirag` tool (Voyage embed/search over the same
> wiki chunks — authoring tooling, not a host runtime resource). The
> never-landed follow-on slices are tracked in [`../TODO.md`](../TODO.md)
> (M-8, M-9; slice 2 superseded by mesa Recall/LTM).

The host has access to two kinds of stored knowledge:

1. **Per-host memory** (episodic, reflections, KV tiers) — the host's
   private record of what it has experienced. The shipped stack — tiered
   `memory.Manager`, hostkv, mesa LTM — is covered in
   [`../memory.md`](../memory.md).
2. **Shared knowledge** (the corpus) — text-based reference material
   every host can consult. Backed by the `cognition/corpus` package.

This doc covers #2.

## The two namespaces

The shared corpus is split into exactly two namespaces
(`cognition/corpus/namespace.go`), and **access is gated at load time**,
not query time:

| Namespace | Contains | When loaded |
|---|---|---|
| `gameplay` | rsc.wiki (the only source ever ingested); the namespace's scope is anything a real RSC player could plausibly have read — fan sites, in-game manuals, lore | Whenever `legacy-cradle -wiki-dump` points at the HTML page dump |
| `dev` | AutoRune script archive, OpenRSC server source, RSC+ client decompile, our own protocol notes — the *inside view* of how the game is implemented | Only when the cradle is launched with `-dev` (no dev source has ever been ingested — TODO.md M-8) |

The split exists because the project goal is to build hosts that
read like *players*, not like bot authors. A host with read access to
the server's PlayerHandler.java can reason about exploit-shaped
strategies no real player would have. A host with read access to
AutoRune can pattern-match its own behavior against historical
botting scripts — also not a thing a real player does.

We want both available during development (so *we* can use them
while building the system), but production launches must not have
them in memory.

## How the gate works

The gate is enforced when the federation is constructed. A cradle
that doesn't pass `-dev` will literally not load the dev sources
into memory — there is no buggy query, no prompt-injection attack,
no clever routine that can reach content that isn't there.

In code (`cmd/legacy-cradle/main.go:822`):

```go
allowed := []corpus.Namespace{corpus.Gameplay}
if cfg.devMode {
    allowed = []corpus.Namespace{corpus.Gameplay, corpus.Dev}
}
fed := corpus.NewFederation(sources, allowed)
```

`NewFederation` (`cognition/corpus/namespace.go:84`) drops any
`corpus.Source` whose namespace isn't in `allowed` *before*
returning — the filtered-out source never enters the federation, and
`Recall` has no path back to it. Passing a nil `allowed` defaults to
`[Gameplay]`, the safe-by-default behavior. A source with a nil
`Corpus` is likewise skipped, so wiring code can leave un-ingested
sources as nil placeholders.

If an operator forgets the `-dev` flag while intending to test
something with dev knowledge, the dev sources are silently dropped
at construction: retrieval returns no dev hits at all. That is a
loud-failure mode (queries that needed dev content come back empty)
rather than a quiet-success mode where dev content silently bleeds
through. This load-time-gating pattern is the security design worth
keeping even if the corpus mechanism itself is retired.

## Multiple sources per namespace

A namespace can be populated by many sources. The `gameplay`
namespace today has only rsc.wiki; the dev-side ingests (AutoRune
scripts, OpenRSC source, RSC+ decompile) were planned but never
done, and the "discover, don't seed" decision in
[`../world-knowledge-and-learning.md`](../world-knowledge-and-learning.md)
forbids seeding them as live host knowledge anyway (TODO.md M-8).
Sources in the same namespace are merged at recall time —
`Federation.Recall` fans the query out to every source, merges hits,
and returns the global top-N by score — so routines see one logical
pool of "things a player could know."

`Federation.Sources()` returns the post-filter set for the boot
log (`cmd/legacy-cradle/main.go:830`):

```
INFO corpus source enabled name=rscwiki namespace=gameplay chunks=12217
```

That log line is the auditable record of what knowledge a running
host has access to. Launching with `-dev` additionally logs a
warning that the flag must never be used for production hosts.

## `recall(query, top=N)` from routines

Routines call `recall` like any other DSL builtin:

```
chunks = recall("how do I cook a lobster", 3)
for c in chunks {
    note(c)
}
```

When a corpus is wired, the returned list contains formatted strings
(`formatChunkForRecall`, `runtime/dsl_helpers.go:278`):

```
[rscwiki § Manual:Cooking (2003) § Cooking Items] A list of the
different foods you can cook in Runescape are shown below with the
cooking level needed (cooking) and the amount it heals your hit
points...
```

The provenance prefix `[source § page § section]` is always
present. A routine that branches on content can also log where it
came from.

The dispatch order (`dslRecall`, `runtime/actions_memory.go:97`):

1. **`host.Corpus` non-nil** → query it directly, return formatted
   chunks. Only `cmd/legacy-cradle` ever sets this field
   (`main.go:198`); scoring is the keyword/TF match in
   `cognition/corpus/corpus.go` (body hits +1, section title +5,
   page title +10, stop words dropped).
2. **`host.Corpus` nil** → fall back to `host.Retriever`
   (`cognition.Client`) and return its `Bundle.Reflections` list.
   This is the path every production host takes.

What the fallback actually returns depends on how the host was built:

- **Offline / no mesa link**: the default `cognition.StubClient`
  (`runtime/host.go:458`) serves canned reflections — Phase 2.5
  stub behavior, unchanged.
- **mesa-linked** (`runtime/runhost_bootstrap.go:110`): the
  Retriever is `mesaclient.AsRetriever`, which routes to the mesa
  `Knowledge.Recall` RPC — RAG over the host's own episodic LTM
  (pgvector cosine when an embedder is wired, Postgres full-text
  relevance + recency otherwise — `mesa/mesad/ltm.go:470`). **Wrinkle, verified at `7f22bb6`:** the adapter puts
  RPC hits into `Bundle.Episodic` (`mesa/client/adapters.go:59`)
  while `dslRecall` reads only `Bundle.Reflections`
  (`runtime/actions_memory.go:134`), so a mesa-linked `recall()`
  currently returns an empty list even when the LTM has matches.
  Mesa memories reach the host through other seams — session genesis
  (`mesa/mesad/genesis.go:37`), the memory Remote
  (`mesaclient.AsRemote`, behind `remember`/`recollect`), and the
  cold-start episodic bootstrap (`host.SetMesaMemory`) — `recall()`
  is just not one of them today.

## Follow-on work

The original Phase 2.6 slice plan (mesa-backed `corpus.Corpus`,
AutoRune ingest, `-knowledge-query` CLI) lives in
[`../TODO.md`](../TODO.md): slice 2 was superseded by the mesa
Recall RPC + LTM + `mesa/wikirag`; slice 3 is **M-8**; slice 4 is
**M-9**. The wiki-RAG cohort experiment that would actually exercise
the corpus is **R-1**.
