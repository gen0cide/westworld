# Routine Runtime versioning policy

> **STATUS: BUILT — VERSION IN ARREARS** (verified 2026-06-10 against branch
> HEAD `0bfa818`). The *mechanism* is built and enforced: every `.routine`
> declares the version it targets, the loader refuses incompatible targets,
> and the policy below is real code (`dsl/spec/version.go`,
> `runtime/dsl_bridge.go`). The *discipline* lapsed within days of landing:
> `spec.RuntimeVersion` is still `1.0.0` while the surface has grown since
> 2026-05-31 — `scan_for`, `search_map`/`survey_map`,
> `remember`/`recollect`/`forget`, and the `go_to`/`converse` rework (arguably
> MAJOR) all sit in `dsl/spec/actions.go` today, plus new events, with no
> bump. The CHANGELOG prose is now backfilled (the *Unversioned arrears*
> section of [`CHANGELOG.md`](CHANGELOG.md)); the overdue bump(s), the version
> assignment for those entries, and the surface-snapshot test that would have
> caught this are tracked as **DSL-18** in [`docs/TODO.md`](../TODO.md).

---

## 1. What "the Routine Runtime" is

A single semver covers the whole host-facing contract a `.routine` depends on:

- **The language surface** — the builtin / event / accessor tables in
  [`dsl/spec/`](../../dsl/spec/) (`actions.go`, `events.go`, `accessors.go`) and
  the syntax accepted by `dsl/parser`.
- **Interpreter semantics** — evaluation rules in `dsl/interp` (truthiness,
  scoping, equality, numeric promotion, control flow, resource caps).
- **The frozen host API** — [`api.md`](api.md) (faculties, value types, the
  capability boundary).

The current version is the constant **`spec.RuntimeVersion`**
([`dsl/spec/version.go`](../../dsl/spec/version.go)), anchored at **`1.0.0`**
because `api.md` was frozen as "v1" (commit `38ef5a0`).

## 2. The semver policy

Bump per the change you make:

| Bump | When | Effect on existing scripts |
|---|---|---|
| **MAJOR** `X.0.0` | **Breaking.** Remove or rename a builtin·event·accessor; change a signature, return type, or argument meaning incompatibly; change evaluation semantics (truthiness, scope, equality, eval order); remove or repurpose syntax. | Scripts targeting an **older major may break.** |
| **MINOR** `1.X.0` | **Additive & backward-compatible.** New builtin·event·accessor; new *optional* parameter; new syntax that doesn't change existing behaviour; relaxed cap/limit. | Older-minor scripts **still run unchanged.** |
| **PATCH** `1.0.X` | **Fixes only.** Bug fix, perf, error-message/diagnostic wording, decoder correction — no surface change and no behaviour change for a *correct* script. | No effect. |

Rule of thumb tied to the chokepoint: if a change adds a row to a `dsl/spec`
table → **minor**; removes/renames/retypes a row, or changes `dsl/interp`
semantics → **major**; touches neither surface nor semantics → **patch**.

## 3. The `runtime "X.Y"` directive (script targeting)

Every `.routine` declares the runtime it targets with a file-level directive:

```
runtime "1.0"

routine kill_goblins() {
    ...
}
```

- It is a **first-class directive** (keyword `runtime` + a version string), one
  per file, at file scope (order-independent with `extends`;
  `dsl/parser/parser.go:218` rejects duplicates). The parser stores it on
  `ast.File.Runtime`.
- **Targets are `MAJOR.MINOR`** (patch never affects compatibility; `"1"` and
  `"1.0.3"` are also accepted — `spec.ParseVersion` defaults missing trailing
  components to 0).
- **Mandatory for disk-loaded routines.** `runtime.ParseRoutineFile`
  (`runtime/dsl_bridge.go:53`) **errors** if the directive is missing. Its
  callers: `legacy-cradle -routine` (via `Host.RunRoutine`), the conductor's
  coro spawn (`runtime/coro.go`), the REPL's `.load`/`.run`
  (`runtime/repl.go`), `cmd/parsecheck`, and `cmd/scenariogen`.
- **Optional for transient string-loaded routines** — REPL fragments and
  `exec()` / `improvise()` snippets (`runtime.ParseRoutineString`) may omit it
  (assumed current runtime); if they *do* declare one it is still compat-checked.

### Compatibility contract

A script targeting `X.Y` runs on runtime `A.B.C` **iff `X == A` and `Y <= B`** —
i.e. the runtime is the same major and the same-or-newer minor. Concretely
(`spec.CheckTarget`):

- target minor ≤ runtime minor, same major → **runs**.
- target needs a newer minor (`Y > B`) → **refused** (the feature it wants isn't here).
- different major → **refused** (a major bump is breaking).

This is forward-compatible within a major: a newer runtime keeps running older
scripts; it never silently runs a script written for an incompatible runtime.

## 4. How to bump the runtime (the process)

1. Make the change in `dsl/spec` / `dsl/parser` / `dsl/interp`.
2. Edit `spec.RuntimeVersion` in [`dsl/spec/version.go`](../../dsl/spec/version.go)
   per §2.
3. Add a dated entry to [`CHANGELOG.md`](CHANGELOG.md) naming the new version,
   the bump class, and what changed.
4. On a **MAJOR** bump, also bump the `Runtime:` header on the affected docs
   (§6) and archive the previous major's docs under `docs/lang/vN/` if the
   surface diverged enough to confuse readers.
5. `go test ./...` (the version tests — `dsl/spec/version_test.go` plus the
   loader checks in `runtime/dsl_bridge_test.go` — guard the mechanics).

> **Enforcement.** Today this is CHANGELOG + review discipline, backed by the
> `dsl/spec` tables being the single chokepoint for the surface — and the
> arrears banner above is proof that discipline alone is not enough. The fix —
> a `dsl/spec` surface-snapshot test that fails when the
> builtin/event/accessor set changes without a `RuntimeVersion` bump, turning
> "remember to bump" into a hard gate — is **DSL-18** in
> [`docs/TODO.md`](../TODO.md), together with the overdue bump(s) and the
> version assignment for the backfilled CHANGELOG arrears entries.

## 5. Finding scripts that don't fit the current runtime

Because `ParseRoutineFile` enforces the directive + compatibility, the existing
gates **are** the sweep:

```bash
go run ./cmd/parsecheck examples      # parses every .routine under the given root
                                      # (default root: examples/scenarios);
                                      # reports missing/incompatible targets
go run ./cmd/scenariogen              # validates the manifest ⇄ corpus, same enforcement
grep -rL 'runtime "' --include='*.routine' examples/   # any .routine with no directive
```

A failing file names exactly why (missing directive, needs-newer-minor, or
wrong-major), so a runtime bump's blast radius is greppable up front.

The only `.routine` files outside this regime are the five conformance
fixtures in `testdata/conformance/` — the conformance runner parses them with
`parser.Parse` directly (`dsl/conformance/runner.go:131`), bypassing the
loader, so they carry no directive.

## 6. Versioned documentation

[`api.md`](api.md) and [`syntax.md`](syntax.md) carry a `Runtime:` header
pinning them to the runtime they describe. On a MAJOR bump, snapshot the prior
docs under `docs/lang/v<major>/` before editing, so a script pinned to an old
major still has its contract documented. `CHANGELOG.md` is the linear record
across all versions.

## 7. Status / history

- **1.0.0** — initial version; the frozen v1 API surface (`api.md`) + the
  `dsl/spec` tables as of `38ef5a0`. The `runtime "X.Y"` directive and this
  policy landed 2026-05-31 (`898b3fc`); the 224 `.routine` files then under
  `examples/` all targeted `1.0` (the 5 conformance fixtures already bypassed
  the loader). Today the repo holds **299** `.routine` files: **294** under
  `examples/` declare `runtime "1.0"`, plus the same 5 loader-bypassing
  conformance fixtures (§5).
- **Arrears** — every surface change since 1.0.0 shipped without a bump or a
  CHANGELOG entry at the time; the entries are now backfilled under
  *Unversioned arrears* in [`CHANGELOG.md`](CHANGELOG.md), with the bump
  itself still pending — see the banner above and **DSL-18** in
  [`docs/TODO.md`](../TODO.md).

See [`CHANGELOG.md`](CHANGELOG.md) for the running record.
