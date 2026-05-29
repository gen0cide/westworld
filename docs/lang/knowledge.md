# Knowledge — the corpus, namespaces, and `recall()`

The host has access to two kinds of stored knowledge:

1. **Per-host memory** (reflections, episodic, persona) — the host's
   private record of what it has experienced. Backed by mesa in
   production; covered in [`thought-architecture.md`](thought-architecture.md).
2. **Shared knowledge** (the corpus) — text-based reference material
   every host can consult. Backed by the `cognition/corpus` package.

This doc covers #2.

## The two namespaces

The shared corpus is split into exactly two namespaces, and **access
is gated at load time**, not query time:

| Namespace | Contains | When loaded |
|---|---|---|
| `gameplay` | rsc.wiki, tip.it, in-game manuals, lore — anything a real RSC player could plausibly have read | Always (when the cradle is given a corpus directory) |
| `dev` | AutoRune script archive, OpenRSC server source, RSC+ client decompile, our own protocol notes — the *inside view* of how the game is implemented | Only when the cradle is launched with `-dev` |

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

The gate is enforced when the federation is constructed. A
production cradle that doesn't pass `-dev` will literally not load
the dev sources into memory — there is no buggy query, no
prompt-injection attack, no clever routine that can reach content
that isn't there.

In code (`cmd/cradle/main.go`):

```go
allowed := []corpus.Namespace{corpus.Gameplay}
if cfg.devMode {
    allowed = []corpus.Namespace{corpus.Gameplay, corpus.Dev}
}
fed := corpus.NewFederation(sources, allowed)
```

The federation drops any `corpus.Source` whose namespace isn't in
`allowed` *before* returning. The `Corpus` field the rest of the
host sees has no path back to filtered-out sources.

If a future operator forgets the `-dev` flag and intended to test
something with dev knowledge, they'll get an empty federation — a
loud-failure mode where retrieval just returns no hits. Better than
a quiet-success mode where dev content silently bleeds through.

## Multiple sources per namespace

A namespace can be populated by many sources. The `gameplay`
namespace today has only rsc.wiki, but will eventually also have
tip.it, the official RuneScape manuals, and forum archives. They
all live in the same namespace and are merged at recall time —
routines see one logical pool of "things a player could know."

The `dev` namespace will similarly fold together AutoRune scripts,
OpenRSC source, RSC+ client decompile, RSCGo Go reference, and our
own protocol notes once those ingest pipelines land.

`Federation.Sources()` returns the post-filter set for the boot
log:

```
INFO corpus source enabled name=rscwiki namespace=gameplay chunks=12217
```

That log line is the auditable record of what knowledge a running
host has access to.

## `recall(query, top=N)` from routines

Routines call `recall` like any other DSL builtin:

```
chunks = recall("how do I cook a lobster", 3)
for c in chunks {
    note(c)
}
```

The returned list contains formatted strings:

```
[rscwiki § Manual:Cooking (2003) § Cooking Items] A list of the
different foods you can cook in Runescape are shown below with the
cooking level needed (cooking) and the amount it heals your hit
points...
```

The provenance prefix `[source § page § section]` is always
present. A routine that branches on content can also log where it
came from.

If the host has no corpus wired, `recall()` falls back to the
older per-host stub retriever (Phase 2.5 behavior) — returning the
host's canned reflections instead of real chunks. Production
intends to have both, eventually: real chunks AND real reflections.

## Slice plan (Phase 2.6)

| Slice | What lands | External infra |
|---|---|---|
| 1 ✓ | `corpus.MemoryCorpus` + `LoadWikiDump`, namespace gating, federation, `-wiki-dump` / `-dev` flags, recall() wired | None |
| 2 | Postgres + pgvector + Voyage 3 embeddings; swap `MemoryCorpus` impl for a mesa-backed `Corpus` without changing the federation surface | Docker, Voyage API key |
| 3 | AutoRune script corpus ingest under namespace `dev` | Location of the script archive |
| 4 | `cradle -knowledge-query` admin CLI | Builds on slice 2 |

Slice 1 is shipped (this PR). Slices 2–4 are pending.
