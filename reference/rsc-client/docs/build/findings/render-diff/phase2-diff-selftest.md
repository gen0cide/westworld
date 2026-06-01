> _Verbatim build report from the rsc-client-render-diff-first-tool workflow (2026-06-01)._

# Phase 2 Report — `cmd/renderdiff` (the §5 diff harness) + self-test

## What was built

A working render-diff tool that runs **today, with only the GO engine** — proving the door/wall perception-bug class is detectable before the Java engines exist. All changes are additive (one pure-extraction edit to `render/render.go`); uncommitted; `go build ./...` and full `go test ./...` are green.

## Files created (absolute paths)
- `/home/free/code/rsc-hacking/westworld/cmd/renderdiff/main.go` — CLI: resolves `LABEL=PATH` inputs (`.json` → GO render, `.png` → decode + sidecar faces), runs both diff modes per pair, writes heatmaps/artifacts, exit-code gate.
- `/home/free/code/rsc-hacking/westworld/cmd/renderdiff/diff.go` — `ComparePixels` (max-channel Δ, count/%, maxΔ, bbox), `Heatmap` (magenta highlight), `DecodePNG`, and `CompareFaces` (multiplicity-aware structural set diff).
- `/home/free/code/rsc-hacking/westworld/cmd/renderdiff/selftest_test.go` — the door-detection + zero-diff-control self-test, wired as a permanent regression guard.
- `/home/free/code/rsc-hacking/westworld/cmd/renderdiff/README.md` — usage, two diff modes, determinism rules, and the "ADDING THE OTHER TWO ENGINES" handoff (DEOB + JAR entrypoints, the `rscdump-faces/1` sidecar schema).
- `/home/free/code/rsc-hacking/westworld/render/dump_faces.go` — `render.RenderDumpFaces` / `RenderDumpFacesWith` + `BuiltFace` type + `CentroidGrid`. The GO engine's built-face export (L2-style, from the L1 build), keyed by (rounded transformed-centroid XYZ, vertex count, front/back fill).
- `/home/free/code/rsc-hacking/westworld/internal/rscdump/gen/single_tile_nodoor/main.go` — generates the perturbed fixture by loading the with-door fixture and zeroing its wall grids (so the two can only differ by the door).
- `/home/free/code/rsc-hacking/westworld/testdata/rscdump/single_tile_NOdoor.json` — the perturbed fixture (door removed; "removed 1 wall byte").

## Modified files (additive)
- `/home/free/code/rsc-hacking/westworld/render/render.go` — pure extraction: split the scene-build out of `renderViewToSurface` into a new shared `buildScene(...) (*Scene, View)`. Both the rasterizing path (`RenderDump`) and the structural export (`RenderDumpFaces`) now call the SAME builder, so the structural set always describes the exact geometry the pixel diff rasterizes. `RenderView` external behaviour unchanged.

## Design choices
- **Structural key is camera-independent**: centroid is the mean of each face's *transformed world-space* vertices (after `baseX/Y/Z`+orientation, before camera), rounded to 16 world units (⅛ tile) — fine enough to keep distinct walls on one tile separate, coarse enough to absorb cross-engine last-bit integer drift. This is exactly what a DEOB/JAR face export must reproduce; the sidecar schema is documented in the README.
- **Engine-agnostic input boundary**: any engine that emits a `.png` + `<base>.faces.json` (schema `rscdump-faces/1`) is a first-class input, so the 3-way diff is one `renderdiff` invocation away. Verified the seam works: decoding the emitted `go-door.png` + sidecar and diffing against the no-door dump reproduces the identical result.

## EXACT self-test output

**Door detection** (`go-door` = with-door vs `go-nodoor` = without-door), exit status 1:
```
=== go-door_vs_go-nodoor ===
  PIXEL: DIFFER 2534/171008 px (1.482%) maxΔ=51, region [x 232..279, y 126..180] heatmap=testdata/rscdump/out/heatmap_go-door_vs_go-nodoor.png
  STRUCT: DIFFER  go-door=25282 faces, go-nodoor=25281 faces, 25281 shared
    only in go-door (present here, ABSENT in go-nodoor):
      face model#1 centroid=(10304,-96,10368) verts=4 fillFront=-15719 fillBack=-15719
```
- Pixel diff is **non-zero and localized** to the wall footprint (1.482% of frame, region 48×55 px), not a frame-wide cloud.
- Structural diff shows **exactly one** face present in with-door, absent in without-door: the 4-vertex door quad at centroid (10304,−96,10368), wood-door fill −15719. No spurious extras in the other direction. This is precisely the "unwalkable door / unperceived wall" signature.

**Zero-diff control** (`go-door` vs itself), exit status 0:
```
=== go-door_vs_go-door-copy ===
  PIXEL: IDENTICAL (0 differing, maxΔ=0) heatmap=testdata/rscdump/out/heatmap_go-door_vs_go-door-copy.png
  STRUCT: IDENTICAL (25282 faces both, 25282 shared)
```
No false positives — two independent renders of the same dump are byte-identical (confirms determinism rules §4-2/4).

## Output artifact paths
- Heatmap (door localized as a magenta quad over dim-grayscale context): `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/heatmap_go-door_vs_go-nodoor.png`
- Zero-diff heatmap (all-dark): `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/heatmap_go-door_vs_go-door-copy.png`
- Rendered PNGs + face sets: `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/{go-door,go-nodoor}.png` and `.faces.json`

## Handoff notes (in README)
- **DEOB**: headless `DumpRenderer` over the deob `scene`/`world` packages → `World.loadFromDump(...)` injects dumped grids (bypass `loadMapData`), pin speckle to `terrainSeed`, `buildSection`/`addModels`, `setCameraOrientation`/`setBounds`, `Scene.render` → grab `Surface.pixels` (obf `ua.rb`) → `deob.png`; walk built `Scene` models → `deob.faces.json`. Gated on the deob-compile workflow; `JAVA_HOME=java-17-openjdk`. First 2-way diff: GO↔DEOB.
- **JAR (oracle)**: rscplus `DumpRenderer` — inject the single-tile grid the way the L1 hook reads obf `k` (or replay-seek per §3b), one `drawScene`, grab `Renderer.pixels` (mirrored `ua.rb`) → `jar.png`; read already-built `lb.Z[]` models → `jar.faces.json` (never execute obf builders — `client.vh`).
- Both engines just emit `PNG + faces.json` in the documented schema; the 3-way diff is then `renderdiff -out OUT go=fixture.json deob=OUT/deob.png jar=OUT/jar.png`.

## NUMBERS
- **go build ./...: PASS** (also `go test ./...` ALL PASS, `go vet` clean, `gofmt -l` clean on all new/modified files)
- **self-test door-detection: PASS** — flags the removed wall: PIXEL 2534/171008 px (1.482%) localized to region [x 232..279, y 126..180], maxΔ=51; STRUCT exactly 1 door face (quad, centroid (10304,−96,10368), fill −15719) present-in-door / absent-in-nodoor
- **false-positive control: PASS** — with-door vs itself: 0 differing pixels (maxΔ=0), identical face sets (25282/25282 shared), exit 0
- **output artifacts**:
  - `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/heatmap_go-door_vs_go-nodoor.png` (door diff heatmap)
  - `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/heatmap_go-door_vs_go-door-copy.png` (zero-diff control)
  - `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/go-door.png` + `go-door.faces.json`, `go-nodoor.png` + `go-nodoor.faces.json`
  - perturbed fixture: `/home/free/code/rsc-hacking/westworld/testdata/rscdump/single_tile_NOdoor.json`
