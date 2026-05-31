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
3. [`development-workflow.md`](development-workflow.md) — the
   build model. **Delores is one of the build assistants**, not
   just a test subject. Live REPL-driven development with the
   running host in the loop.
4. [`syntax.md`](syntax.md) — surface syntax: comments, filenames,
   identifiers, keywords, naming conventions.
5. [`state.md`](state.md) — the **IF THIS** query layer. The ~100
   accessors every host has into its own state and the world.
6. [`events.md`](events.md) — the **WHEN THIS** subscription layer.
   `on` / `when` / `select` / `defer` / `super()` / `try`/`recover`.
7. [`actions.md`](actions.md) — the **THEN THAT** action layer.
   Result/Error model, bang convention, the verb menu.
8. [`repl.md`](repl.md) — interactive prompt spec, including the
   drop-into-shell-on-failure mode.

## Relationship to other docs

- [`docs/architecture.md`](../architecture.md) — system-wide
  architecture (the 12 layers of a `cradle` host, the three
  binaries, the mesa server). Read first if you want the bigger
  picture.
- [`docs/phases.md`](../phases.md) — full multi-phase plan
  (what we're building when).
- [`docs/tasks.md`](../tasks.md) — flat enumeration of every
  work item with task IDs.

## Where the language surface lives in code

**Single source of truth: `dsl/spec/`** (Go package). Three tables:

- `dsl/spec/actions.go` — every callable: actions, primitives,
  LLM stdlib, memory stdlib, persona reads. Each row has name,
  kind, arity, params, doc summary, and a NotYetImplemented flag.
- `dsl/spec/events.go` — every bus event a routine can `on`-handle.
- `dsl/spec/accessors.go` — every query-layer attribute path
  (host.hp, self.skills.fishing.level, etc.). Documentation and
  consistency source; the actual Getter implementations live in
  `runtime/dsl_views.go`.

**Anything that needs to know "what is in this language" reads
from `dsl/spec/`**: the validator (`dsl/validator/`), the host
bridge (`runtime/dsl_bridge.go`), the REPL (planned), and
doc-generation tooling. Adding a new builtin = add a row to
`spec.Actions` plus a Go wrapper in `runtime/dsl_actions.go`.
Adding a new event = add a row to `spec.Events`. Adding a new
query = add a row to `spec.Accessors` plus the Getter switch arm.

**Consistency tests catch drift**:
- `dsl/spec/consistency_test.go` — internal invariants on the
  spec tables (unique names, snake_case, non-empty docs, etc.)
- `runtime/dsl_spec_consistency_test.go` — cross-package: every
  spec entry has a Go handler (or is NotYetImplemented); every
  Go handler has a spec entry; every registered builtin name is
  in spec.

These tests fail loudly if someone adds an action in one place
and forgets the other. Run `go test ./dsl/spec/... ./runtime/...`
to verify.
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
- The **control-flow vocabulary is now built**: `on` (file- and
  routine-level), `when` watchers, `select`, `defer`, `try`/`recover`,
  single-arg lambdas (`x => expr`), and file-level
  `extends "parent.routine"` all **execute** today — see the
  implementation-status table in [`events.md`](events.md). The
  outstanding reactive piece is the per-handler `extends host` /
  `super()` chain, which needs the persona tier (Phase 4).
- But the **query layer is thin** — only ~15 accessors on
  `self`/`world`/`inventory` exist; need to scale to ~100
- Plus error handling is stringly-typed; needs the Result/Error
  model in [`actions.md`](actions.md)

These docs are a mix of **what's running now** and **what's planned
next**. The state of any individual feature is called out per-doc;
for the reactive constructs specifically, the status table in
[`events.md`](events.md) is authoritative.
