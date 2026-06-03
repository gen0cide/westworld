# renderdiff — the RSC 3-way render-diff harness

`renderdiff` implements §5 of [`RENDER_DIFF_DESIGN.md`](../../reference/rsc-client/docs/build/RENDER_DIFF_DESIGN.md):
it compares two (or three) renders of the **same dumped world state** and reports
where they disagree, so we can prove the Go `render/` lib, the deobfuscated Java
client, and the vanilla rev-235 jar are faithful to each other — and localize
perception bugs (unwalkable door, broken bridge, roof culling).

It runs **today, with only the GO engine** (the DEOB and JAR engines land via the
other workflow — see *Adding the other two engines* below). The self-test proves
the tool detects the door/wall bug class without them.

---

## Usage

```
renderdiff -out DIR [-tol N] [-emit-render] LABEL=PATH LABEL=PATH [LABEL=PATH]
```

Each `LABEL=PATH` is one labelled render source. `PATH` is either:

- **an `rscdump/1` L1 JSON** (`.json`) — rendered through the GO engine
  (`render.RenderDump`) into a framebuffer **and** a built-face set, so one dump
  drives **both** diff modes. With `-emit-render` (default on) the rendered PNG
  and the `<label>.faces.json` are also written into `-out`.
- **a pre-rendered PNG** (`.png`) — decoded for the pixel diff. An optional
  sidecar `<base>.faces.json` (the same schema `-emit-render` writes) supplies
  its face set for the structural diff. **This is the seam the DEOB and JAR
  engines plug into**: each emits a PNG + a `faces.json` and renderdiff diffs
  them unchanged.

Flags:

| flag | default | meaning |
|---|---|---|
| `-out DIR` | `.` | directory for diff heatmap PNGs + rendered artifacts |
| `-tol N` | `1` | per-channel LSB tolerance: a pixel *differs* only when max `\|Δ\|` exceeds N (absorbs ≤1-LSB rounding) |
| `-emit-render` | `true` | for `.json` inputs, also write the rendered PNG + `faces.json` into `-out` |

The tool prints a one-line verdict per pair, writes `heatmap_<A>_vs_<B>.png` per
pair into `-out`, and **exits non-zero** if any pair shows a pixel OR structural
diff — so it doubles as a CI gate once golden fixtures land.

### Example — the self-test

```sh
# door detection: with-door vs without-door  → exits 1 (diff found)
go run ./cmd/renderdiff -out testdata/rscdump/out \
    go-door=testdata/rscdump/single_tile_door.json \
    go-nodoor=testdata/rscdump/single_tile_NOdoor.json

# zero-diff control: with-door vs itself      → exits 0 (no diff)
go run ./cmd/renderdiff -out testdata/rscdump/out \
    go-door=testdata/rscdump/single_tile_door.json \
    go-copy=testdata/rscdump/single_tile_door.json
```

The same checks run as `go test ./cmd/renderdiff/...` (`selftest_test.go`), so
they are a permanent regression guard.

---

## The two diff modes

### 1. Pixel diff (`diff.go`, `ComparePixels` + `Heatmap`)

Compares two equal-size `0x00RRGGBB` framebuffers. A pixel *differs* when its
**max per-channel absolute delta** exceeds `-tol`. Reports:

- count + **percent** of differing pixels,
- the largest single-channel `|Δ|` (`maxΔ`),
- the **bounding box** of the differing region (the localized area a change
  shows up in),
- a **heatmap PNG**: in-tolerance pixels are drawn as a dim grayscale of the
  scene (context), out-of-tolerance pixels as bright magenta scaled by delta.

A PNG input is decoded back to a framebuffer (`DecodePNG`), so a captured DEOB/JAR
PNG diffs against a GO render with no shared buffer.

### 2. Structural diff (`diff.go`, `CompareFaces`; export in `render/dump_faces.go`)

Independent of rasterization quirks. Each engine's **built face set** is compared
by a camera-independent key:

> `BuiltFace.Key()` = (rounded transformed-centroid X,Y,Z) · (vertex count) · (front fill) · (back fill)

The centroid is the mean of the face's **transformed world-space** vertices
(after the model's `baseX/Y/Z` + orientation, **before** the camera), rounded to
`render.CentroidGrid` (16 world units = ⅛ tile) to absorb last-bit integer drift
between engines while keeping distinct walls on one tile separate. The diff
reports faces **present in A but absent in B** (and vice-versa) — exactly the
**missing door quad / wrong-fill bridge deck / unculled roof face** a perception
bug produces, surfaced as a discrete face rather than a fuzzy pixel cloud.

The GO engine produces its set with `render.RenderDumpFaces(dump)`, which builds
the **identical `Scene`** that `render.RenderDump` rasterizes (both call the
shared `buildScene`), so the structural set always describes the same geometry
the pixel diff sees.

---

## Determinism requirements (§4)

A diff is only meaningful if every engine renders the **same bytes** from the
same dump. The GO engine satisfies this and the other engines MUST match:

1. **State is a file, not a session.** All engines load the identical `rscdump/1`
   blob. The GO engine reads the dump's terrain grids directly
   (`rscdump.Dump.Landscape`) — **no map-file decode** — so RNG/network/timing
   can't drift it. DEOB and JAR must likewise inject the dumped grids and skip
   their own map decode for diff runs.
2. **Kill the terrain RNG.** The vanilla client injects `(int)(Math.random()*10)-5`
   per terrain vertex. The dump carries `terrain.terrainSeed`; the GO engine pins
   the speckle to it (`render.withDumpTerrainSeed`; seed `0` ⇒ flat-zero speckle
   ⇒ byte-identical). DEOB/JAR must seed the same table from `terrainSeed`.
3. **Pin animation.** `AnimFrame = 0` for static diffs (the schema carries
   `tick`/`animFrame` for live/replay diffs).
4. **Pin window + camera exactly.** Same `baseX/baseY`, `W/H`, `viewDist`,
   `clipNear`, `clipFar`, screen centre `(W/2, H/2)`. The schema **pins these so
   the diff fails loudly** if an engine uses a different constant — that mismatch
   is itself a faithfulness bug. (Known drift to reconcile: GO `clipFar=7000`,
   the jar sets `2400` per frame; the fixture records what GO renders so the gap
   is visible, not hidden.)
5. **Same fill/colour tables.** Hand-authored fixtures emit raw colours
   (`syntheticFacts` gives every wall id a generic wood-door def), so no GameData
   ramp is needed. Real-map dumps resolve ids through the same GameData via
   `RenderDumpWith` / `RenderDumpFacesWith`.

The **zero-diff control** (`TestSelfTest_ZeroDiffControl`) verifies rule-2/4
determinism holds: two independent GO renders of the same dump are byte-identical
(0 pixel diff, identical face set).

---

## Self-test result (GO engine only, no Java)

`go test ./cmd/renderdiff/...` (and the CLI invocations above) prove the tool
detects the wall/door bug class:

- **door detection** (`single_tile_door` vs `single_tile_NOdoor`):
  - PIXEL: **2534 / 171008 px (1.482%)** differ, `maxΔ=51`, region
    `[x 232..279, y 126..180]` — a localized wall footprint, not the whole frame.
  - STRUCT: **exactly one** face present in *with-door*, absent in *without-door*
    — a 4-vertex quad at centroid `(10304, -96, 10368)` with the wood-door fill
    `(-15719/-15719)`. That is the door wall, named precisely.
- **zero-diff control** (`single_tile_door` vs itself): **0** differing pixels
  (`maxΔ=0`), face sets identical (25282 both, 25282 shared) — no false positive.

The `single_tile_NOdoor.json` fixture is generated by
`go run ./internal/rscdump/gen/single_tile_nodoor`, which loads the with-door
fixture and zeroes its wall grids, so the two can only ever differ by the door.

---

## Adding the other two engines (handoff)

renderdiff is **engine-agnostic at the input boundary**: any engine that writes
a **PNG** + a **`<base>.faces.json`** in the `BuiltFace` schema is a first-class
input. The 3-way diff is then literally:

```sh
go run ./cmd/renderdiff -out OUT \
    go=fixture.json \
    deob=OUT/deob.png \
    jar=OUT/jar.png
```

(`go=fixture.json` renders+emits `go.png`+`go.faces.json`; `deob`/`jar` are
pre-rendered PNGs whose sidecar faces.json the tool picks up automatically.)

### The `faces.json` schema each engine must emit

```jsonc
{
  "schema": "rscdump-faces/1",
  "faces": [
    { "model": 1, "centroid": [10304, -96, 10368], "numVerts": 4,
      "fillFront": -15719, "fillBack": -15719 }
  ]
}
```

`centroid` = the mean of each face's **transformed world-space** vertices
(post `baseX/Y/Z`+orientation, pre-camera), rounded to **16 world units**
(`render.CentroidGrid`); `fillFront/fillBack` are the face front/back fills
(`<0` = flat 5:5:5 colour, `>=0` = texture id); `numVerts` = the face's vertex
count. Match these exactly or faces won't line up across engines.

### DEOB engine (gated on the deob-compile workflow)

Implement a headless `DumpRenderer` main compiled against the now-compiling deob
`scene` / `world` packages (`reference/rsc-client/src/client`). Entrypoints:

1. **Load** the `rscdump/1` JSON (port the schema, or shell out to a tiny Go
   helper that re-emits the grids as a flat binary the Java side mmaps).
2. **Inject** the dumped terrain grids straight into `world.World`'s `byte[][]`
   fields via a new package-private `World.loadFromDump(...)` filling
   `terrainHeight/terrainColour/tileDecoration/wallsRoof/wallsEastwest/`
   `wallsNorthsouth/wallsDiagonal/tileDirection` — **bypassing `loadMapData`**
   (determinism rule 1). Pin the `Math.random()` speckle to `terrain.terrainSeed`
   (rule 2).
3. **Build** the scene: `buildSection` / `buildRoofs` / `method422` / `addModels`.
4. **Camera**: `Scene.setCameraOrientation(...)` + `Scene.setBounds(...)` from
   `camera.*` (pin `viewDist/clipNear/clipFar`, rule 4).
5. **Render** one frame: `Scene.render(...)`.
6. **Pixel out**: grab `Surface.pixels` (obf `ua.rb`) `int[]` → PNG via `ImageIO`
   (or raw int[] for a byte-exact diff). Write `deob.png`.
7. **Faces out**: walk the built `Scene` model list (`scene.GameModel.vertexX/Y/Z`
   after the queued transform, `faceVertices`, `faceFillFront/Back`), compute the
   `BuiltFace` centroid/key exactly as `render/dump_faces.go:sceneFaces` does, and
   write `deob.faces.json`.

Compile note: `Scene`/`GameModel`/`Surface`/`World` are package-private/`final`
and pull cross-class statics; the harness needs those tables present. Use
`JAVA_HOME=/usr/lib/jvm/java-17-openjdk`. Mudclient itself stays out of scope.
**First 2-way diff: GO ↔ DEOB** (both our code, shared constants align directly).

### JAR engine (oracle; replay-seek per §3b)

Implement a rscplus headless `DumpRenderer` mode:

1. Drive the obf client to the dumped state — for the hand-authored fixture,
   **inject the single-tile grid** the way the L1 hook reads obf `k`'s
   `byte[][]` grids (`L/eb/f/P/s/A/R/mb`); for real maps, **replay-seek to the
   recorded frame** (`Replay.java` `frame_time_slice`/`timestamp`) — the
   deterministic state source.
2. Let the obf renderer run **one** `drawScene`.
3. **Pixel out**: grab `Renderer.pixels` — already the mirrored obf `ua.rb`
   (`JClassPatcher.java` hook) — → PNG. Write `jar.png`.
4. **Faces out** (for the structural diff): after the live client builds its
   scene, read obf `lb.Z[]` (Scene models) and each `ca` model's vertex/face
   arrays (`a/ob/bc`, `o`, `V/qb`), compute the `BuiltFace` key as above, write
   `jar.faces.json`. (Read the **already-built** model array — do NOT execute obf
   builders in isolation; the `client.vh` opaque predicate makes that unsafe.)
5. Run headless: `export JAVA_HOME=/usr/lib/jvm/java-17-openjdk`; build via the
   portable Ant launcher.

With both PNGs+faces.json emitted, the **real 3-way diff** (GO ↔ DEOB ↔ JAR) on
the single-tile-with-door scene is the single `renderdiff` invocation shown above
— the oracle pins down whether a divergence is in *our* scene-builder
(L1 structural delta) and, with the L2 layer, builder-side vs rasterizer-side.
```

---

## Files

| file | role |
|---|---|
| `main.go` | CLI: resolve inputs, run both diff modes per pair, write heatmaps + artifacts, exit-code gate |
| `diff.go` | `ComparePixels` / `Heatmap` / `DecodePNG` (pixel diff) + `CompareFaces` (structural diff) |
| `selftest_test.go` | the door-detection + zero-diff-control self-test (regression guard) |
| `../../render/dump_faces.go` | `render.RenderDumpFaces` + `BuiltFace` — the GO engine's built-face export |
| `../../internal/rscdump/gen/single_tile_nodoor/` | generates the perturbed `single_tile_NOdoor.json` fixture |
