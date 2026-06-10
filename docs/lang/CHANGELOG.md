# Routine Runtime — CHANGELOG

> **STATUS: IN ARREARS — entries backfilled, version bump pending** (verified
> against code 2026-06-10, branch HEAD `0bfa818`). `spec.RuntimeVersion`
> ([`dsl/spec/version.go`](../../dsl/spec/version.go)) still reads `1.0.0`
> while the surface below it grew by 13 builtins and 3 events — every entry in
> the **Unversioned arrears** section shipped with no bump and no CHANGELOG
> line. The prose is now caught up (this file); the bump itself, plus the
> surface-snapshot test that would have caught the lapse, are code work:
> **DSL-18** in [`docs/TODO.md`](../TODO.md).

The running record of every Routine Runtime version. Bump per
[`versioning.md`](versioning.md) (MAJOR = breaking, MINOR = additive, PATCH =
fix-only) and add an entry here in the same change. The current version is the
`spec.RuntimeVersion` constant in [`dsl/spec/version.go`](../../dsl/spec/version.go).

Format: `## X.Y.Z — YYYY-MM-DD (MAJOR|MINOR|PATCH)` followed by the changes.

---

## Unversioned arrears — 2026-06-06 → 2026-06-10 (bump pending, DSL-18)

These changes shipped against `RuntimeVersion = "1.0.0"` without a bump. Each
carries the classification it *should* have received; versions are not
retro-assigned. Net surface delta 1.0.0 → `0bfa818`: builtin table 54 → 67,
event table 31 → 34 (plus `level_up` made real), accessor table unchanged at
114 — but `3c0ab33` (breaking `converse`/`go_to` rework) and the `dfc3052`
`self.fatigue` rescale are MAJOR material, so the catch-up bump is a single
`2.0.0` (or a 1.x ladder ending there). Decision + bump + snapshot test:
DSL-18.

Newest first.

### 2026-06-10 — `86f9088` (PATCH)

- Reserved namespace roots single-sourced as `spec.ReservedRoots()`
  ([`dsl/spec/spec_roots.go`](../../dsl/spec/spec_roots.go)); the validator
  had drifted and was missing `host`. Assignments to `host` are now rejected —
  enforcement of the existing api.md §6 contract, not a contract change.
  `host` is reserved ahead of its accessors being wired (they remain
  NotYetImplemented).

### 2026-06-10 — `b42a52b` (PATCH)

- Spec truth-fixes from the DSL-manual rewrite: `host.name` marked
  NotYetImplemented (doc points at `self.name`, which is live); the
  `equipment_changed` doc no longer claims a bare `on equipment_changed()`
  handler ignores extra args.

### 2026-06-09 — `0214a0b` (MINOR)

- C-style boolean aliases: the lexer ([`dsl/lex/lex.go`](../../dsl/lex/lex.go))
  accepts `!`, `&&`, `||` as aliases for `not`, `and`, `or` — Postel's law for
  LLM authors trained on C-family code. A gapless `!` after a keyword
  (`wait!`) is still rejected: keywords never take bang variants.
- Doc fixes in the event specs: `trade_request`/`duel_request` summaries now
  point at the real `trade.request(from)` / `duel.request(from)` reciprocation
  verbs (the previously documented `accept_trade`/`accept_duel` don't exist).

### 2026-06-08 — `3c0ab33` (arguably MAJOR — breaking)

- `converse` reworked from drive-the-whole-dialog to **listen**: arity drops
  from 1–2 to exactly 1 (**the `pick` arg is gone** — NPCs only speak
  pre-authored lines and are not queryable), it auto-advances only choices
  code can resolve and **stops at any real multi-option menu**, and the return
  changes from a menus-answered count to a map
  `{ said, options, ended, answered }`. Scripts using `converse(npc, pick)` or
  branching on the old count break.
- `go_to` **no longer accepts a POI type** (`"bank"`, `"furnace"`, ...):
  `CatalogPlace` (town/landmark names + coords only) replaces
  `CatalogPlaceOrPOI`. Auto-routing to a type masked ignorance and walked
  hosts into gates; the pattern is now `search_map("bank")` → choose → `go_to`
  the coords. Scripts passing a POI-type literal break at validation.

### 2026-06-08 — `7bcfdf1` (MINOR)

- New builtin `scan_for(type, radius?)` — local scenery enumeration,
  nearest-first, field-accessible entries (`{x, y, name, kind, def_id,
  position}`); returns an empty list (never a failure) when nothing matches.
  Reads static map + live view, so depleted objects drop out.

### 2026-06-07 — `ec5d73c` (MINOR)

- WorldOracle map-perception builtins: `search_map(type)` (ranked real
  destinations with per-destination `reach`/`gate`/`needs`/`payable`),
  `reachable(x, y)` (explain how you'd reach one tile), `survey_map()`
  (text overview of open vs gated vs blocked destinations). Capability-gated
  perception — the oracle informs, the brain decides.

### 2026-06-07 — `dfc3052` (MINOR; the fatigue rescale is arguably MAJOR)

- `ParamKinds` catalog metadata added to `ActionSpec` plus a static
  literal-arg checker ([`dsl/validator/static_args.go`](../../dsl/validator/static_args.go)):
  a hallucinated catalog literal (`go_to("mining-site")` with no such place,
  `eat("typo-item")`) is rejected at validation instead of round-tripping to
  the host. Mirrors runtime resolution, so it rejects nothing the engine
  would have accepted.
- Spec→manual generator ([`dsl/spec/manual.go`](../../dsl/spec/manual.go))
  feeding the mesad DSL manual.
- **`self.fatigue` rescaled**: raw 0..750 → normalized 0..100 percent.
  Script-visible change to an existing accessor's value range — any threshold
  comparison written against the raw scale breaks (arguably MAJOR).

### 2026-06-06 — `dacc425` (MINOR)

- `use` and `interact_at` gain coordinate forms: max arity 2 → 4 with named
  `x`/`y` params — `use(item, x=X, y=Y)` targets the scenery at a tile (the
  way to cook/smith on scenery), `interact_at(x=X, y=Y, option=N)` clicks a
  tile without a view.

### 2026-06-06 — `8c5b146` (MINOR)

- Memory stdlib over the tiered memory manager: `remember(key, value)`,
  `recollect(key)` (scratch→local→mesa cascade, null on miss), `forget(key)`
  (all tiers). Location-blind: the key's namespace prefix decides the tier.

### 2026-06-06 — `57ab761` (MINOR)

- Navigation + perception builtins: `go_to` (original form — superseded by
  `3c0ab33` above), `converse` (original form — reworked by `3c0ab33`),
  `look_around(radius?)`, `where_am_i()`, `bearing_to(x, y)`,
  `where_is(name)` — the gazetteer/perception verb family.
- `talk_to` accepts a name string (`talk_to("banker")` auto-targets the
  nearest visible NPC of that name) in addition to a view or index.
- Events: `level_up` implemented (was NotYetImplemented; synthesized by
  diffing per-skill base levels); new events `equipment_changed`,
  `player_equipment_changed`, `player_level_changed`.

---

## 1.0.0 — 2026-05-31 (initial)

The baseline. Captures the host-facing surface as frozen in `api.md` ("v1",
commit `38ef5a0`) plus the interpreter semantics and the `dsl/spec` tables as of
this date: 54 builtins, 31 events, 114 accessors.

Also introduced in this release (the versioning system itself — mechanism, not a
surface change to the v1 contract):

- `runtime "X.Y"` file-level directive (`dsl/token`, `dsl/ast`, `dsl/parser`).
- `spec.RuntimeVersion` + `spec.CheckTarget` compatibility logic.
- `ParseRoutineFile` makes the directive **mandatory** + compat-checks it;
  `ParseRoutineString` makes it optional + compat-checks when present.
- All 224 directive-carrying `.routine` files target `1.0` (224 of 229 tracked
  at the time — the 5 without are `testdata/conformance/` fixtures, which
  bypass the runtime loader (`parser.Parse` directly in
  `dsl/conformance/runner.go`), where the directive is optional. As of
  2026-06-10 the split is 294 of 299, same 5 fixtures.)

<!--
Template for the next entry (the next real one is the DSL-18 catch-up bump —
fold the arrears section above into it):

## X.Y.Z — YYYY-MM-DD (MAJOR|MINOR|PATCH)
- Added `world.foo` accessor (#NNN).
- Added `bar()` builtin (#NNN).
(Older 1.0 scripts continue to run unchanged.)
-->
