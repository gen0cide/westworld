# Westworld DSL — Language Documentation

> **STATUS: CURRENT — verified 2026-06-10 against branch
> `tidy/structure-and-docs`, HEAD `0bfa818`.** The language is BUILT
> and executable end-to-end; the per-doc status notes below say
> where each piece lives in code. Open language work is tracked in
> [`docs/TODO.md`](../TODO.md) (the SSOT — IDs like `DSL-13`/`DSL-17`
> cited inline); this doc carries no backlog of its own.

> **Host** — an autonomous AI actor in the system. One host = one
> logged-in OpenRSC character that perceives the world, thinks, and
> acts on its own (one `runtime.Host`; `cmd/host` runs one per
> process, the cradle daemon `cmd/cradle-server` runs a whole fleet
> in one process). The term comes from the Westworld TV show;
> throughout this codebase and these docs, the agents we build are
> **always** called *hosts* — never "bots," never "agents," never
> "characters."

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
5. [`state.md`](state.md) — the **IF THIS** query layer. The 114
   accessor paths every host has into its own state and the world.
6. [`events.md`](events.md) — the **WHEN THIS** subscription layer.
   `on` / `when` / `select` / `defer` / `super()` / `try`/`recover`.
7. [`actions.md`](actions.md) — the **THEN THAT** action layer.
   Result/Error model, bang convention, the verb menu.
8. [`repl.md`](repl.md) — the interactive prompt. BUILT
   (`runtime/repl.go`, reachable via `cmd/host -repl`); the daemon
   era adds `POST /eval` on the per-host debug surface
   (`debughttp/server.go`).
9. [`versioning.md`](versioning.md) — the Routine Runtime semver policy, the
   required `runtime "X.Y"` script-targeting directive, and doc versioning;
   [`CHANGELOG.md`](CHANGELOG.md) is the running version record.

Reference / companion docs in this folder:

- [`api.md`](api.md) — the **frozen v1 Routine API surface**
  (ratified 2026-05-29); the design contract cognition / brain /
  persona build on. Its §8 hand-maintained tags have drifted —
  the durable fix is the `cmd/specdoc` generator (`DSL-17`).
- [`writing-routines.md`](writing-routines.md) — the operational
  authoring guide: how to make a routine do the right thing
  against the live OpenRSC server, plus RSC scripting nuances.
- [`knowledge.md`](knowledge.md) — the shared corpus, namespaces,
  and `recall()`.
- [`protocol.md`](protocol.md) — the Payload235 wire-protocol
  shadow of the API (for builders, not routine authors), with
  [`protocol-debug-strategy.md`](protocol-debug-strategy.md) as
  its decode-debugging companion.

## Relationship to other docs

- [`docs/architecture.md`](../architecture.md) — system-wide
  architecture sliced by Go package (the three production
  binaries, the layer cake, mesa). Read first if you want the
  bigger picture; [`docs/layers.md`](../layers.md) is the
  AI-perspective view of the same material.
- [`docs/phases.md`](../phases.md) — the phase ladder: the
  historical record of what shipped in what order, and the decoder
  for phase-number references used across docs and code comments.
- [`docs/tasks.md`](../tasks.md) — the **closed** historical ledger
  of the numbered build ladder (#19–#123). The live work queue is
  [`docs/TODO.md`](../TODO.md).
- [`docs/dsl.md`](../dsl.md) — original ~900-line design doc that
  seeded the language. Captures the *why* and the goals; it now
  carries a supersession note at the top. The docs in `docs/lang/`
  are the current working spec and supersede dsl.md on specific
  points (event model, scoped watchers, Error/Result shape,
  filename rules). When they disagree, the doc here wins.
- [`docs/brain.md`](../brain.md), [`docs/cognition.md`](../cognition.md),
  [`docs/memory.md`](../memory.md), [`docs/reveries.md`](../reveries.md),
  [`docs/personas.md`](../personas.md) — the cognitive layers a
  routine sits above and consults. `thought-architecture.md` here
  explains how routines interface with each.

## Where the language surface lives in code

**Single source of truth: `dsl/spec/`** (Go package). Three tables:

- `dsl/spec/actions.go` — every callable: actions, primitives,
  LLM stdlib, memory stdlib, persona reads. Each row has name,
  kind, arity, params, doc summary, and a NotYetImplemented flag.
- `dsl/spec/events.go` — every bus event a routine can `on`-handle.
- `dsl/spec/accessors.go` — every query-layer attribute path
  (self.hp, self.skills.fishing.level, etc.). Documentation and
  consistency source; the actual Getter implementations live in
  the `runtime/views_*.go` files.

**Anything that needs to know "what is in this language" reads
from `dsl/spec/`**: the validator (`dsl/validator/`), the host
bridge (`runtime/dsl_bridge.go`), the REPL (`runtime/repl.go`), and
doc-generation tooling (`dsl/spec/manual.go` renders the planner's
API reference straight from the tables). Adding a new builtin =
add a row to `spec.Actions` plus a Go wrapper in
`runtime/dsl_actions.go`. Adding a new event = add a row to
`spec.Events`. Adding a new query = add a row to `spec.Accessors`
plus the Getter switch arm.

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

## Status

- DSL phases 1–9 (per `docs/dsl.md` enumeration) shipped 2026-05-28.
- The language is **executable end-to-end**: lexer (`dsl/lex`),
  parser (`dsl/parser`), validator (`dsl/validator`), interpreter
  (`dsl/interp` — caps, events, watchers, defer, try/recover,
  lambdas), host bridge (`runtime/dsl_bridge.go`), and the
  conformance suite (`dsl/conformance`).
- The **query layer is BUILT**: 114 accessor paths in
  `dsl/spec/accessors.go`, implemented as Getter arms across the
  eleven `runtime/views_*.go` files (self, world, inventory,
  combat, bank, shop, trade, duel, prayer, magic, resolve) —
  except 8 rows flagged `NotYetImplemented` (the five `host.*`
  persona-tier reads plus `self.prayer.active`,
  `self.spells.known`, `inventory.weight`). Known spec/impl
  coverage gaps are `DSL-2`.
- The **typed Result/Error model is BUILT**: `dsl/interp/error.go`
  defines the `ErrorCode` enum (PATH_BLOCKED, DOOR_LOCKED,
  INVENTORY_FULL, …) that routines branch on via `result.err.code`;
  Go action wrappers must return an ErrorCode, never an ad-hoc
  string. [`actions.md`](actions.md) documents the idioms.
- The **REPL is BUILT**: `runtime/repl.go` over a long-running
  `dsl/interp` Session bound to the live host — `self.hp` at the
  prompt queries the running character. Remaining niceties
  (multi-line input, `-repl-on-fail`/`.resume`) are `DSL-27`.
- The **reactive vocabulary executes**: `on` (file- and
  routine-level), `when` watchers, `select`, `defer`,
  `try`/`recover`, single-arg lambdas (`x => expr`), and
  file-level `extends "parent.routine"` — the implementation-status
  table in [`events.md`](events.md) is authoritative. The one
  unbuilt reactive piece is the per-handler `extends host` /
  `super()` chain: the validator rejects it today
  (`dsl/validator/validator.go:668`), and building it is `DSL-13`.

All remaining language work lives in [`docs/TODO.md`](../TODO.md)
under the `DSL-*` IDs.
