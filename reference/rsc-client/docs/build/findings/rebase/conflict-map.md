> _Verbatim analysis from the upstream-render-rebase-analysis workflow (2026-06-01)._

Decisive. Upstream gutted `render/render.go` (574 lines → 32 lines, a 542-line deletion that rips out the entire classic scene-assembly path: `buildScene`, `TerrainHeights`, `Scene`, `Surface`, etc.). Our edits to `render.go` add MORE code into exactly that gutted region (`buildScene`, `renderViewToSurface`, the diagonal-objects pass). The overlap is total — a guaranteed content conflict where "ours" depends on deleted symbols (`Scene`, `Surface`, `TerrainHeights`) that no longer exist on orsc. So even if you take "ours" for render.go, it won't compile.

I have a complete, verified analysis. Producing the report.

---

# REBASE CONFLICT MAP — 11 commits (`3a384ec..a8ab7d4`) onto `origin/main` (`9ff7a69`)

## Summary

- **merge-base** = `3a384ec`; **origin/main** = `9ff7a69` (3 upstream commits ahead); **our tip** = `a8ab7d4` (11 commits ahead).
- Upstream rewrote the renderer: **deleted** the classic software renderer (`render/{boundary,camera,facesort,model,palette,pick,rasterize,roof,scene,scenery,surface,terrain,textures,textures_test}.go`), **added** `render/orsc/*`, and **gutted** the two survivors `render/render.go` (574→32 lines, −542) and `render/{entity,entitysprite,itemsprite,sprites}.go`.
- **8 of 11 commits are CLEAN** (touch only `reference/`, which upstream never touched).
- **3 commits are render commits and ALL conflict**: `24495c5`, `276b58a`, `a8ab7d4`. They produce **modify/delete conflicts** (on deleted `render/*.go`), **content conflicts** (on surviving `render/render.go` + `.gitignore`), and the survivors are **build-breaks** regardless of resolution (they call deleted APIs `Scene`/`Surface`/`TerrainHeights`).
- **Verification method**: per-commit `git merge-tree --write-tree --merge-base=<parent> origin/main <sha>` (mirrors the rebase 3-way apply of each patch), plus `git diff --name-status`, `git cat-file -e origin/main:<path>`.

---

## 1. Per-commit disposition table

(commits listed in rebase apply order, oldest first)

| # | Commit | Subject (short) | Touches render/? | Disposition |
|---|--------|-----------------|------------------|-------------|
| 1 | `a803796` | RSC deob snapshot (70 classes) | no | **CLEAN** — `reference/` only |
| 2 | `960310c` | RSC deob correctness comb-over | no | **CLEAN** — `reference/` only |
| 3 | `d2db342` | deob rendering full-expansion + docs | no | **CLEAN** — `reference/` only (mass D/A under `reference/rsc-client/decompiled/`, `.gitignore` under `reference/rsc-client/`) |
| 4 | `5b78e9b` | docs INDEX consolidation | no | **CLEAN** — `reference/rsc-client/docs/` only |
| 5 | `24495c5` | render-diff tool (rscdump/1 + GO engine) | **YES** | **CONFLICT** — content on `render/render.go`; modify/delete on `render/terrain.go`; **+ NEW-FILE-BUILD-BREAK** (`render/dump.go`, `render/dump_faces.go`, `cmd/renderdiff/*`, `internal/rscdump/*`) |
| 6 | `b0a21df` | build-options + milestone-1 docs | no | **CLEAN** — `reference/rsc-client/docs/` only |
| 7 | `4f97ecc` | deob subsystem compile WIP | no | **CLEAN** — `reference/rsc-client/src/` only |
| 8 | `2b2af58` | milestone-2 results docs | no | **CLEAN** — `reference/rsc-client/docs/` only |
| 9 | `276b58a` | 7 render-fidelity fixes + JAR oracle | **YES** | **CONFLICT** — content on `.gitignore`; modify/delete on `render/{boundary,roof,scenery,terrain,dump}.go`; **+ NEW-FILE-BUILD-BREAK** (3 `fidelity_*_test.go`) |
| 10 | `33e7226` | Mudclient net-handler fixes | no | **CLEAN** — `reference/rsc-client/src/client/Mudclient.java` only |
| 11 | `a8ab7d4` | port World.addModels (diagonal objects) | **YES** | **CONFLICT** — content on `render/render.go`; **+ NEW-FILE-BUILD-BREAK** (`render/diagobj.go`, `render/fidelity_diagobj_test.go`) |

---

## 2. File-level conflict list (render commits only)

### `24495c5` — render-diff tool
- **Content conflict** — `render/render.go` (survives on origin/main; upstream deleted the `buildScene`/scene-assembly body our hunk extends).
- **Modify/delete** — `render/terrain.go` (deleted on origin/main, we modify it → "version 24495c5 left in tree").
- **New-file build-breaks** (apply clean, won't compile against orsc): `render/dump.go`, `render/dump_faces.go`. Supporting net-new dirs `cmd/renderdiff/{main,diff,selftest_test}.go`, `internal/rscdump/{landscape,schema}.go`, `internal/rscdump/gen/*` apply clean (paths don't exist on origin/main) but depend on the deleted `render` API surface.

### `276b58a` — fidelity fixes + JAR oracle
- **Content conflict** — `.gitignore` (both upstream and our hunk append at `@@ -49,3` after `/scenariogen`; overlapping two-line hunks → conflict).
- **Modify/delete (×5)** — `render/boundary.go`, `render/roof.go`, `render/scenery.go`, `render/terrain.go`, `render/dump.go` (all deleted on origin/main; we modify each → "version 276b58a left in tree").
- **New-file build-breaks**: `render/fidelity_scenery_test.go`, `render/fidelity_terrain_test.go`, `render/fidelity_walls_test.go` (call deleted render API). `testdata/rscdump/**` + `internal/rscdump/gen/rooftrace/main.go` apply clean (untracked paths on origin/main).

### `a8ab7d4` — diagonal objects (World.addModels)
- **Content conflict** — `render/render.go` (our hunk adds the diagonal-objects pass into the same gutted `buildScene` region).
- **New-file build-break** — `render/diagobj.go`, `render/fidelity_diagobj_test.go` (reference deleted scene/scenery types).
- **NOT conflicts** (false positives in isolated single-commit replay): `internal/rscdump/landscape.go`, `internal/rscdump/schema.go`, `reference/.../RENDER_FIDELITY_FINDINGS.md` — these are **added by earlier commits in our own stack** (`24495c5`, `276b58a`) and only *modified* here; during a real sequential rebase they already exist, so they apply clean. (`pathfind/landscape.go` auto-merges clean.)

### Net render conflict set (real sequential rebase)
- **Content conflicts (3):** `render/render.go` (twice — at `24495c5` and `a8ab7d4`), `.gitignore` (at `276b58a`).
- **Modify/delete conflicts (6 distinct files):** `render/terrain.go`, `render/boundary.go`, `render/roof.go`, `render/scenery.go`, `render/dump.go`, plus `render/terrain.go` again — i.e. `{terrain, boundary, roof, scenery, dump}.go`.
- **Build-breaking new files (5):** `render/dump.go`, `render/dump_faces.go`, `render/diagobj.go`, plus the 4 `render/fidelity_*_test.go`.

**Critical correctness note:** `render/render.go` survives but is gutted (−542 lines removing `buildScene`, `Scene`, `Surface`, `TerrainHeights`). Our edits insert code into that exact removed region and call those now-deleted symbols. So even resolving the content conflict by **taking "ours"** yields a tree that **does not compile** against orsc. These render edits are obsolete against the new renderer.

---

## 3. Clean-commit confirmation (8 commits)

Confirmed via `git cat-file -e origin/main:<path>` and path inspection: upstream touched ONLY `render/`, `assets/`, `facts/`, `cmd/{cradle,scenariogen,rendertest}`, `examples/scenarios/`, `docs/{render-engine,tasks}.md`, `.gitignore`. It did **not** touch `reference/` at all. Therefore these all rebase clean per-file:

- `a803796`, `960310c`, `d2db342`, `5b78e9b` — RSC deob + docs (`reference/` only).
- `b0a21df`, `2b2af58` — milestone docs (`reference/rsc-client/docs/` only).
- `4f97ecc` — **in-flight subsystem-compile work** (`reference/rsc-client/src/**` only; net add `Scanline.java`, net deletes of `nativeapi/*`/`AudioOutput.java`). Clean.
- `33e7226` — **just-committed Mudclient net fixes** (`reference/rsc-client/src/client/Mudclient.java` only). Clean.

(Note: `d2db342` and `276b58a` both touch a `.gitignore`, but `d2db342`'s is `reference/rsc-client/.gitignore` — a different file from the repo-root `.gitignore` upstream edited, so no collision. Only `276b58a`'s root-`.gitignore` edit conflicts.)

---

## 4. Recommended rebase mechanics + strategy

### Safety prerequisite (BLOCKING)
The in-flight subsystem-compile workflow has **uncommitted changes in the MAIN working tree** under `reference/rsc-client/src`. A rebase requires a clean tree. **Do not start until** the main tree is committed or stashed:
```
# in /home/free/code/rsc-hacking/westworld (main tree, NOT this worktree)
git status                 # must be clean
git stash push -u          # or commit the in-flight work first
git rev-parse deob/rsc-client-reference   # confirm tip == a8ab7d4 before/after
```
The **rscplus JAR-oracle harness** lives in a separate repo (`/home/free/code/rsc-hacking/rscplus` @ `b4f942b`) and is **unaffected** by this rebase — no action.

### Recommended strategy: **Option B — interactive rebase that DROPS the 3 render commits**

The render edits are obsolete: they patch a renderer upstream deleted, and even "take ours" won't compile against `render/orsc`. Carrying them forward is pure conflict-churn for a tree that won't build. Drop them; re-land the genuinely valuable render-diff tooling as fresh work on top of orsc.

```
git checkout deob/rsc-client-reference
git rebase --onto origin/main 3a384ec deob/rsc-client-reference -i
```
In the todo list, **`drop`** the three render commits and **`pick`** the eight clean ones:
```
pick   a803796   # deob snapshot      (clean)
pick   960310c   # deob comb-over     (clean)
pick   d2db342   # deob full-expand   (clean)
pick   5b78e9b   # docs INDEX         (clean)
drop   24495c5   # render-diff tool   <-- DROP (obsolete vs orsc)
pick   b0a21df   # build-options docs (clean)
pick   4f97ecc   # subsystem compile  (clean)
pick   2b2af58   # milestone-2 docs   (clean)
drop   276b58a   # fidelity fixes     <-- DROP (obsolete vs orsc)
pick   33e7226   # Mudclient net fix  (clean)
drop   a8ab7d4   # diagonal objects   <-- DROP (obsolete vs orsc)
```
Result: a **conflict-free** rebase of the 8 deob/docs/net commits onto orsc. Then re-land the still-useful parts (rscdump/1 schema, render-diff CLI, the diagonal-door fidelity finding) as **fresh commits targeting `render/orsc`** — much of `internal/rscdump` and `cmd/renderdiff` is renderer-agnostic and can be salvaged; note `render/orsc/harness.go` already exists upstream as a render-diff harness, so coordinate with it rather than re-introducing `render/dump.go`.

### Alternative: **Option A — straight `--onto` keeping all 11 (NOT recommended)**

```
git rebase --onto origin/main 3a384ec deob/rsc-client-reference
```
Expect to stop at three commits. Resolutions:
- **`24495c5`**: `git rm render/terrain.go` (accept upstream delete); for `render/render.go` content conflict either `git rm`/revert the hunk or take-ours — but take-ours **won't compile** (`Scene`/`Surface`/`TerrainHeights` gone), so effectively you must drop the render.go hunk. The new files `render/dump.go`, `render/dump_faces.go` build-break against orsc → `git rm` them too (or keep and accept a broken build to fix later). `cmd/renderdiff`, `internal/rscdump` keep but will fail to compile until re-pointed at orsc. `git rebase --continue`.
- **`276b58a`**: `git rm render/{boundary,roof,scenery,terrain,dump}.go` (accept upstream delete = drop our obsolete fidelity edits). Resolve `.gitignore` by hand (keep both upstream's `/_render.bak/`+`docs/_research/` block AND our `.claude/`+`/renderdiff` lines). `git rm render/fidelity_*_test.go`. `git rebase --continue`.
- **`a8ab7d4`**: resolve `render/render.go` by dropping our diagonal-objects hunk (the API it extends is gone); `git rm render/diagobj.go render/fidelity_diagobj_test.go`. `git rebase --continue`.

Option A's end state is effectively Option B's (all render edits dropped) but reached through manual conflict resolution that leaves dangling build-broken tool files — strictly worse than the clean interactive drop.

### Post-rebase verification (either option)
```
git log --oneline origin/main..deob/rsc-client-reference   # expect 8 (Option B) commits
go build ./...                                             # must pass on orsc
git diff --stat origin/main..deob/rsc-client-reference -- render/   # Option B: empty
```

### Ordering / safety notes
- **Commit/stash the main-tree in-flight `reference/rsc-client/src` work first** — the rebase will refuse on a dirty tree, and that uncommitted work is otherwise unprotected.
- The 8 clean commits are independent of the renderer; they will replay without touching `render/orsc`.
- The salvage of render-diff tooling (`internal/rscdump`, `cmd/renderdiff`) is **net-new** on origin/main (those dirs don't exist there) — it is purely a compile re-targeting effort against `render/orsc`, not a merge problem.
- The rscplus JAR-oracle harness (separate repo) needs no coordination with this rebase.