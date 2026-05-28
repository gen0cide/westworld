# Westworld DSL — Language Documentation

> **Host** — an autonomous AI actor in the system. One host = one
> running `cradle` process = one logged-in OpenRSC character that
> perceives the world, thinks, and acts on its own. The term comes
> from the Westworld TV show; throughout this codebase and these
> docs, the agents we build are **always** called *hosts* — never
> "bots," never "agents," never "characters."

This subfolder holds the working spec for the `.routine` language
hosts use to drive themselves. It's the practical reference for
"what does the language look like, what does it do, what's planned
next."

## Read order

1. [`overview.md`](overview.md) — the IFTTT mental model and the
   three layers (state / events / actions). Start here.
2. [`thought-architecture.md`](thought-architecture.md) — how the
   routine layer sits inside the host's larger mind: perception,
   memory, brain, persona, reveries.
3. [`syntax.md`](syntax.md) — surface syntax: comments, filenames,
   identifiers, keywords, naming conventions.
4. [`state.md`](state.md) — the **IF THIS** query layer. The ~100
   accessors every host has into its own state and the world.
5. [`events.md`](events.md) — the **WHEN THIS** subscription layer.
   `on` / `when` / `select` / `defer` / `super()` / `try`/`recover`.
6. [`actions.md`](actions.md) — the **THEN THAT** action layer.
   Result/Error model, bang convention, the verb menu.
7. [`repl.md`](repl.md) — interactive prompt spec, including the
   drop-into-shell-on-failure mode.

## Relationship to other docs

- [`docs/architecture.md`](../architecture.md) — system-wide
  architecture (the 12 layers of a `cradle` host, the three
  binaries, the mesa server). Read first if you want the bigger
  picture.
- [`docs/dsl.md`](../dsl.md) — original ~900-line design doc that
  seeded the language. Captures the *why* and the goals. The docs
  in `docs/lang/` are the current working plan and supersede
  dsl.md on specific points (event model, scoped watchers, Error/
  Result shape, filename rules). When they disagree, the doc here
  wins; dsl.md gets pruned during the next cleanup.
- [`docs/brain.md`](../brain.md), [`docs/cognition.md`](../cognition.md),
  [`docs/memory.md`](../memory.md), [`docs/reveries.md`](../reveries.md),
  [`docs/personas.md`](../personas.md) — the cognitive layers a
  routine sits above and consults. `thought-architecture.md` here
  explains how routines interface with each.

## Status at the time these docs were written

- DSL phases 1–9 (per `docs/dsl.md` enumeration) shipped 2026-05-28
- Language is **executable**: parser, validator, interpreter, host
  bridge, caps, event handlers, conformance suite all green
- But the **query layer is thin** — only ~15 accessors on
  `self`/`world`/`inventory` exist; need to scale to ~100
- And the **event/control-flow vocabulary is narrow** — only
  file-level `on event(args)` exists; need `when` / `select` /
  `defer` / `super` / `try`/`recover`
- Plus error handling is stringly-typed; needs the Result/Error
  model in [`actions.md`](actions.md)

The docs in this folder describe what we're building **next**, not
what's already running. State of any individual feature is called
out per-doc.
