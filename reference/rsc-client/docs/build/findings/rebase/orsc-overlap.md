> _Verbatim analysis from the upstream-render-rebase-analysis workflow (2026-06-01)._

I have all the data I need. Let me compile the final report.

# Upstream `render/orsc` vs. Our Render Work — Overlap & Re-Port Analysis

## 0. TL;DR

The colleague's 3 upstream commits **replace the entire classic software renderer** (`render/{boundary,camera,facesort,model,palette,pick,rasterize,roof,scene,scenery,surface,terrain,textures}.go` are **deleted**) with a faithful OpenRSC `three/` port under `render/orsc/`. `render/render.go` shrinks from 655→94 lines: it is now only the `View`/`Entity` value types + sprite-compositing shims (`composite_export.go`). The actual render entry point is now `orsc.RenderView` / `orsc.RenderViewCached` (live, host-centred) and `orsc.RenderBridge` (the colleague's own render-diff hook).

**Net effect on our work:**
- **5 of our 7 fidelity fixes (276b58a) are already correct in orsc** — orsc is a from-scratch faithful port of `World.java`, so it never had the bugs we patched in the old renderer. Those patches are mooted because their target files no longer exist.
- **2 of our fixes are NOT in orsc and must be re-ported**: the **diagonal-door / diagonal-object pass (48001..59999)** (a8ab7d4) and the **id-74 windmill-sail vertical lift** (-480). orsc has genuine gaps here, identical to the bugs we found.
- **Our render-diff tooling is COMPLEMENTARY, not duplicative.** orsc's `harness.go`/`composite_export.go` are a *pixel*-diff hook between the two Go engines only; they have **no serializable dump format, no structural face-set diff, no pixel-diff CLI, and no JAR-oracle leg.** Our `rscdump/1` + `cmd/renderdiff` + the rscplus JAR oracle add all four. But our `dump.go`/`dump_faces.go`/`diagobj.go` import the now-deleted `render` API and **must be retargeted onto `orsc`** to survive the rebase.
- **The deob (Jagex-jar lineage) remains a fully independent oracle.** orsc is OpenRSC lineage; the deob/jar is Jagex rev-235 lineage. They are different codebases of the same game — diffing orsc against the deob/jar is exactly the cross-lineage check that catches faithfulness bugs in *either*. The DEOB render audit stays relevant.

---

## 1. `render/orsc` architecture + public entry points

`render/orsc` is, by its own `doc.go`, a **line-for-line Go port of OpenRSC's `three/` software renderer** (`RSModel.java`, `Scene.java`, `Shader.java`, `World.java`, `Polygon.java`, `Scanline.java`). It is deliberately the *second independent renderer* in the repo — the explicit design intent is that the two engines render the same scene and get diffed:

> "this package exists so the two can render the same scene and be diffed pixel-for-pixel." (`render/orsc/doc.go`)

### Public entry points (the old `RenderView`/`Scene`/`Surface`/`GameModel` equivalents)

| Old `render/` concept | orsc equivalent | File |
|---|---|---|
| `RenderView(land,f,…) → PNG` (the 574-line body removed from render.go) | `orsc.RenderView(land, f, models, sprites, v render.View) ([]byte, error)` and the once-opened `orsc.RenderViewCached(land, f, v)` | `render/orsc/view.go:36`, `view.go:21` |
| `Scene` (model list, camera, rasterizer) | `orsc.Scene` via `NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)`; `scene.AddModel`, `scene.AddEntity`, `scene.SetCamera`, `scene.Render` | `render/orsc/scene.go`, `types.go` |
| `Surface` (framebuffer) | `orsc.Surface`, `NewSurface(w,h)`, `surf.toImage()`; `Surface.Pixels []int32` *is* the Java `GraphicsController.pixelData` | `render/orsc/surface.go`, `doc.go` |
| `GameModel` (geometry accumulator) | `orsc.Model` / `RSModel`: `NewModel`, `FromAssets`, `insertVertex`, `insertFace`, `Orient`, `Translate` | `render/orsc/model.go`, `world.go` |
| `buildTerrain` / `BuildBoundaries` / `BuildRoofs` / `PlaceScenery` | `orsc.BuildTerrain` (world.go), `orsc.BuildWalls`/`wallPass`/`doorPass` (walls.go), `orsc.BuildRoofs`/`roofPass` (roofs.go), `orsc.BuildStories` (stories.go), `placeSceneryModels` + `PlaceScenery` (scenery_place.go, world.go) | as listed |

### How `render/render.go` rewires (it is **gone**, not a shim)

`render.go` no longer renders anything. It is reduced to the `View` struct (camera params + live dynamic-state hooks: `BoundaryRemoved`, `DynamicScenery`, `SceneryRemoved`, `GroundItems`, self-appearance), `DynamicSceneryItem`, `GroundItemMarker`, and the entity/item-sprite compositing. The `RenderView` *function* moved into `orsc/view.go`; `cmd/cradle` now calls `orsc.RenderViewCached(land, f, v)` (`cmd/cradle/main.go:596`, `cmd/cradle/spectate.go:181`). So **RenderView is neither a thin shim nor present — it relocated wholesale into orsc**, and `render/` is now a pure data/compositing helper package the orsc renderer imports (`orsc/view.go` imports `render` for `render.View`, `render.CompositePlayerSprite`, etc.).

### `harness.go` + `composite_export.go` (the colleague's diff/export infra)

- **`render/orsc/harness.go`** — `RenderBridge(x,y,plane,rot,zoom,w,h) ([]byte,error)`. Its doc: *"It is the hook we call to render the 104,655 bridge twice — once through this port and once through render/ — and diff the two."* It opens the same `.orsc` assets, builds terrain/scenery/stories/entities, drives the spectator camera, and returns PNG bytes. It is a **render-to-PNG hook for a manual two-engine pixel diff** — there is no diff *code* here, no serialization, no face-set comparison. (It also still hardcodes `/Users/flint/Code/openrsc` asset paths.)
- **`render/composite_export.go`** — exported wrappers (`CompositePlayerSprite`, `CompositeNPCSprite`, `NPCBillboardSize`, `CompositePlayerAppearanceSprite`) so orsc reuses `render/`'s verified entity compositing for billboard pixels. Pure code reuse, not diff infra.

**Confirmed:** a grep of `render/orsc/*.go` for `json|Dump|Serialize|BuiltFace|rscdump|PixelDiff|Marshal` returns **zero hits** (the one match is a prose comment in `scenery_place.go`). orsc has **no serializable cross-engine dump format and no structural diff** — only the in-process PNG hook.

---

## 2. Bug-overlap table

Verdict legend: **ALREADY-FIXED-UPSTREAM** = orsc's faithful port already does the correct thing (our patch is mooted because its target file is deleted); **NOT-HANDLED** = orsc has the same gap, port forward; **MOOT** = the file we patched is deleted and orsc supersedes it.

| Bug | Our fix (file) | orsc status + cite | Verdict |
|---|---|---|---|
| **Diagonal scenery objects / diagonal doors (DiagonalWalls 48001..59999)** | `render/diagobj.go` `BuildDiagonalObjects` — ports `World.addModels` (a8ab7d4) | **NOT HANDLED.** orsc handles diagonal **walls** only in `1..11999` ('\') and `12001..23999` ('/') — `walls.go:88-90`, `walls.go:125-127`, `roofs.go:114-116`. The `48001..59999` object band is **never read for geometry**: scenery is placed exclusively from `f.SceneryLocs` (`scenery_place.go:62`), not from the landscape's inline 48000+ markers. world.go even documents the band exists (`world.go:209`) but no pass consumes it. Grep for `addModels`/`48000`/`getTileDirection`/`TileDirection` in orsc = **0 code hits** (only a comment at `harness.go:142`). | **NOT-HANDLED → PORT FORWARD.** This is a genuine orsc gap identical to the one we fixed. A diagonal door renders as **zero geometry** in orsc, exactly the "door the cradle can't perceive" bug. |
| **Water-flatten by `tileType==4`** (not overlay id 2/11) | `render/terrain.go` (276b58a) | **ALREADY CORRECT.** orsc flattens a corner to y=0 iff any of the four touching tiles has `typeOf(id)==4` (`world.go:339-349`) — the exact corrected condition. Type-4 ids are 4/12/20/21 in `tileDefs` (`world.go:90-116`); overlay 2/11 are type 3 and are **not** flattened. | **ALREADY-FIXED-UPSTREAM / our patch MOOT** (terrain.go deleted). |
| **Type-2 indoor floor-split gate** | `render/terrain.go` (276b58a) | **ALREADY CORRECT.** orsc gates the seam split on tile-type **class** via `isTileType2` (`world.go:230-247`) and the `decorType != 2 \|\| (diagonal present)` branch (`world.go:416`). | **ALREADY-FIXED-UPSTREAM / MOOT.** |
| **Wall front/back colour via `method422`** (not stone/wood heuristic) | `render/boundary.go` + `dump.go` (276b58a) | **ALREADY CORRECT.** orsc's `insertWall` passes `def.FrontDeco`/`def.BackDeco` **verbatim** into `insertFace(4, …, def.FrontDeco, def.BackDeco)` (`walls.go:64`) — exactly method422's per-face fill resolution, with no name/material heuristic. | **ALREADY-FIXED-UPSTREAM / MOOT.** |
| **id-74 vertical lift (-480)** (windmill sails atop tower) | `render/scenery.go` (276b58a) | **NOT HANDLED.** orsc places scenery purely through `PlaceScenery` → `SceneryCentre` (`world.go`), which does `setRot256(0,dir*32,0)` + translate to footprint centre/ground anchor with **no per-id Y offset**. No `74`/`480`/`lift` token anywhere in `scenery_place.go`/`world.go`/`stories.go`. | **NOT-HANDLED → PORT FORWARD.** orsc will render the windmill sails splayed at terrain level, the same bug we fixed. (Minor/cosmetic, but real.) |
| **Terrain split by tile-type class** (not raw overlay id) | `render/terrain.go` (276b58a) | **ALREADY CORRECT.** The seam-revert compares `isTileType2(x±1,z)` vs `decorType2` (`world.go:411-432`), i.e. by class, matching the deob. | **ALREADY-FIXED-UPSTREAM / MOOT.** |
| **Overlay neighbour-spread** (type-4 quad onto adjacent non-type-3 tiles) | `render/terrain.go` (276b58a) | **ALREADY CORRECT.** orsc's deck/overhang second pass (`world.go:486-518`, `emitDeckQuad` `world.go:524`) emits a deck quad onto each non-type-3 tile bordering a type-4 tile, the `World.java:704-808` overhang — the same neighbour-spread. | **ALREADY-FIXED-UPSTREAM / MOOT.** |
| **Roof under-cull set from `overlayDef.tileType==2`** | `render/roof.go` (276b58a) | **N/A — DESIGN SUPERSEDED.** orsc has no separate "under-roof cull set." It builds roofs as **real 3D geometry** (`roofPass`, `roofs.go`) resolved by the painter's-algorithm plane-overlap sort (`doc.go` pipeline step 5). The cull set was an artifact of the old flat-overlay renderer; orsc's depth sort makes it unnecessary. | **MOOT / no port needed.** Verify visually that roofs occlude interiors correctly, but there is nothing to port. |

**Summary:** 5 fidelity fixes are baked into orsc's faithful port (and their target files are deleted, so the patches vanish on rebase with no conflict — they simply no longer apply). **2 real gaps to port forward**: diagonal-door objects (high value — the cradle-perceptibility bug) and id-74 sail lift (cosmetic).

---

## 3. Tooling-overlap verdict: **COMPLEMENTARY → RE-PORT-OURS-ONTO-ORSC**

### What each side has

| Capability | orsc `harness`/`composite_export` | Our `rscdump` + `cmd/renderdiff` + JAR oracle |
|---|---|---|
| Render a scene to PNG for visual compare | ✅ `RenderBridge` (one engine, in-process) | ✅ `RenderDump` |
| **Serializable cross-engine state format** (one blob 3 engines load) | ❌ none | ✅ `rscdump/1` schema (`internal/rscdump/schema.go`) — `SourceGo`/`SourceDeob`/`SourceJarReplay` |
| **Structural face-set diff** ("is this door quad present?") | ❌ none | ✅ `render.RenderDumpFaces` → `[]BuiltFace` keyed by `(centroid, numVerts, fill)` (`render/dump_faces.go`); `cmd/renderdiff` set-diff |
| **Pixel-diff CLI** (heatmap, %, bbox, CI exit code) | ❌ (no diff code, just PNG bytes) | ✅ `cmd/renderdiff` `PixelDiff` (`cmd/renderdiff/diff.go`, `main.go`) |
| **JAR-oracle leg** (rev-235 obf jar renders the same fixture headless) | ❌ none | ✅ documented working in GO_JAR_DIFF_RESULTS.md (rscplus repo, separate) |

**Verdict: our infra is strictly additive.** orsc's harness only renders one engine to PNG; it cannot diff, cannot serialize a portable fixture, cannot compare face-sets, and has no jar leg. The colleague even *wants* a two-engine diff ("render … twice and diff the two") but only built the render half — **our `renderdiff` is exactly the missing diff half.** Recommendation: **RE-PORT-OURS-ONTO-ORSC** (merge-both: keep our diff/dump/jar infra, retarget the Go-engine leg from the deleted `render` API to `orsc`).

### Why a re-port is mandatory (not optional)

Our tooling **will not compile after the rebase** because it imports the deleted classic API. Confirmed symbol dependencies on now-deleted code:

- `render/dump.go`: `RenderView` (×5), `renderViewToSurface`, `BuildBoundaries`, `Surface`/`.Pix`, `Bundle`, `wallColourWood`, `withDumpTerrainSeed` — all from deleted `render/{render,boundary,surface,terrain}.go`.
- `render/dump_faces.go`: `Scene` (×5), `GameModel`, `Bundle` — from deleted `render/{scene,model}.go`.
- `render/diagobj.go`: `PlaceScenery` (×11), `GameModel` (×5), `NewGameModel`, `AddVertex`/`AddFace`/`SetLight`, `ModelCache`, `elevationAt`, `terrainSize`, `wallColourWood` — from deleted `render/{scenery,model,terrain}.go`.

(Note: there is **no textual git-conflict** on these — upstream *deleted* the files our `boundary.go`/`terrain.go`/`roof.go`/`scenery.go` edits modified, and `merge-tree` shows the only structural conflicts are in `.gitignore` and docs. The breakage is a *semantic*/compile break: our new files reference symbols that vanished.)

### Re-port sketch (retarget the Go leg to orsc; keep schema, CLI, jar leg verbatim)

1. **Keep unchanged** (engine-agnostic, no edits): `internal/rscdump/schema.go`, `internal/rscdump/landscape.go`, `cmd/renderdiff/{main,diff}.go`, the `BuiltFace` type+`Key()`, the rscplus JAR-oracle leg. These never touched the classic render internals.
2. **`render/dump.go` → re-target to orsc.** Replace `renderViewToSurface(land,f,b,v)` with `orsc.RenderView(land, f, models, sprites, v)` (or a new `orsc.RenderDump(d)` that builds `render.View` from the dump and calls `RenderView`). orsc already exposes `Surface.Pixels []int32` (= the framebuffer we copy out for the pixel diff) and `surf.toImage()` (PNG). The `withDumpTerrainSeed` determinism hook must move to orsc's `terrainAmbience` (`world.go:610` — currently a fixed hash, no seed override); add a seed override there so seed-0 → flat-zero speckle for byte-identical diffs. `Bundle`/`wallColourWood`/`syntheticFacts` need orsc equivalents (orsc has no `Bundle`; it opens archives directly — pass `*assets.Archive`/`*assets.SpriteArchive`).
3. **`render/dump_faces.go` → re-target to orsc.** The `[]BuiltFace` export must walk `orsc.Scene.Models` (`orsc.Model` faces) instead of `render.Scene`/`GameModel`. orsc's `Model` carries `insertVertex`/`insertFace` outputs and the queued `Orient`/`Translate`; compute the transformed-space centroid from orsc's baked `vertXTransform/Y/Z` (per `doc.go`) post-`resetTransformCache`. This is the most involved part — it needs a small `orsc` accessor to enumerate built faces + their transformed centroids + `FillFront/FillBack`.
4. **`render/diagobj.go` → re-port as orsc's missing `World.addModels` pass.** This is both the §2 bug fix AND a tooling dependency. Re-implement against `orsc`: scan the window's `DiagonalWalls` for `48000 < d < 60000`, `objectID = d-48001`, `dir = tile.TileDirection`, then place via orsc's `PlaceScenery`/`SceneryCentre` (`world.go`) with the dir-0/4 footprint swap + footprint-clear for multi-tile objects. orsc's `PlaceScenery` already implements the footprint-centre + `dir*32` orient + bilinear `-getElevation` ground-snap, so the body is largely a loop + the synthetic-door fallback. Wire it into `orsc.RenderView`/`RenderBridge` alongside `placeSceneryModels`. **Prerequisite:** confirm orsc's landscape read exposes `TileDirection` (our `internal/rscdump` schema added it; verify the upstream `pathfind.Tile` still carries it — upstream left `pathfind/` untouched, so it should).
5. **Port the id-74 lift** into orsc's `PlaceScenery`/`placeSceneryModels`: `if loc.DefID == 74 { cy -= 480 }` after `SceneryCentre` (`world.go`).

---

## 4. Is the DEOB render audit + behavioral-bug work still relevant? **YES — independent oracle, confirmed.**

The premise holds. orsc is an **OpenRSC-lineage** port (`World.java`/`Scene.java`/`Shader.java` from the OpenRSC `Client_Base` tree). The deob is a **Jagex rev-235-lineage** decompile (`reference/rsc-client/`, the `rsclassic-1091943135.jar`), and the rscplus JAR oracle renders that same Jagex jar. These are two *different codebases of the same game*, not two copies of one. That makes them mutually-checking oracles:

- A bug present in **both** orsc and our old renderer (e.g. the missing diagonal-object pass) is a true gap relative to the game — the cross-lineage agreement confirms it.
- A divergence **between** orsc and the deob/jar localizes a faithfulness bug to one lineage. The `RENDER_DEOB_GAPS.md` audit already catalogs **10 behavioral divergences** in the deob's own render path (texture U/V transpose, `(B,R,G)` vs `(R,G,B)` ground-ramp order, the World↔Scene alias inversion, `loadMapData` wrong-class call, etc.). Several of these — e.g. the ground-ramp channel order — are exactly the kind of thing that, diffed against orsc's `genColorToResource` (`world.go:51-57`, which uses `-(g<<5)-1-(r<<10)-b`), would immediately surface as a per-pixel colour shift in `renderdiff`. So the deob audit *feeds* the diff harness rather than being replaced by it.

**Conclusion:** keep the deob render audit and the deob behavioral-bug work. The deob/jar is the **independent cross-lineage oracle** that the (now-orsc) Go engine should be diffed against — which is precisely the role our `rscdump`+`renderdiff`+JAR-oracle infra fills, and which orsc's own single-engine `harness` cannot.

---

## Appendix — rebase mechanics (read-only prediction)

- `merge-tree 3a384ec origin/main HEAD` reports textual **conflicts only in `.gitignore` and docs** (`*.md`), **not** in render source. Reason: upstream *deleted* the classic render files our 276b58a/a8ab7d4 edits touched; git's rename/delete handling drops our edits to deleted files rather than conflicting.
- Per-file: `reference/`, `internal/rscdump`, `cmd/renderdiff`, `pathfind/`, `world/`, `proto/`, `testdata/rscdump` are **untouched upstream** → they rebase clean *textually*. But `internal/rscdump`/`cmd/renderdiff` are clean only because they don't import the deleted API; **`render/dump.go`, `render/dump_faces.go`, `render/diagobj.go` will fail `go build`** post-rebase (they reference deleted `render.GameModel`/`RenderView`/`BuildBoundaries`/`PlaceScenery`/etc.). These three files are the entire re-port surface (§3).
- Our deob-only commits (960129/d2db342/5b78e9b/33e7226/2b2af58/4f97ecc/b0a21df) touch `reference/` + docs and rebase cleanly (the only friction is the documented stale `.gitignore`/`NAMING.md` prose, which is non-functional).
- The rscplus JAR oracle (rscplus repo @ b4f942b) is independent of this westworld rebase and is unaffected.