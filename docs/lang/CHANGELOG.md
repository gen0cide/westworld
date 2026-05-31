# Routine Runtime — CHANGELOG

The running record of every Routine Runtime version. Bump per
[`versioning.md`](versioning.md) (MAJOR = breaking, MINOR = additive, PATCH =
fix-only) and add an entry here in the same change. The current version is the
`spec.RuntimeVersion` constant in [`dsl/spec/version.go`](../../dsl/spec/version.go).

Format: `## X.Y.Z — YYYY-MM-DD (MAJOR|MINOR|PATCH)` followed by the changes.

---

## 1.0.0 — 2026-05-31 (initial)

The baseline. Captures the host-facing surface as frozen in `api.md` ("v1",
commit `38ef5a0`) plus the interpreter semantics and the `dsl/spec` tables as of
this date.

Also introduced in this release (the versioning system itself — mechanism, not a
surface change to the v1 contract):

- `runtime "X.Y"` file-level directive (`dsl/token`, `dsl/ast`, `dsl/parser`).
- `spec.RuntimeVersion` + `spec.CheckTarget` compatibility logic.
- `ParseRoutineFile` makes the directive **mandatory** + compat-checks it;
  `ParseRoutineString` makes it optional + compat-checks when present.
- All 224 committed `.routine` files target `1.0`.

<!--
Template for the next entry:

## 1.1.0 — YYYY-MM-DD (MINOR)
- Added `world.foo` accessor (#NNN).
- Added `bar()` builtin (#NNN).
(Older 1.0 scripts continue to run unchanged.)
-->
