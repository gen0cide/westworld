# RSC 3-Way Render Diff — status

Tracks implementation of [RENDER_DIFF_DESIGN.md](RENDER_DIFF_DESIGN.md). Built by the
`render-diff-first-tool` workflow (2026-06-01). Raw build reports:
[findings/render-diff/](findings/render-diff/).

## Done — Phase 0 + the diff harness (GO engine only)

The render-diff tool **runs today** and detects the door/wall perception-bug class, before the
DEOB and JAR engines exist. All changes are additive and uncommitted; `go build ./...`,
`go test ./...`, `go vet`, and `gofmt` are all green.

**What landed:**
- **`rscdump/1` schema** — `internal/rscdump/schema.go` (the §4 state-dump: camera, window, terrain
  grids, scenery/entities/ground-items/self, removed boundaries+scenery, and the L2 `Models[]`,
  defined for the later rasterizer-isolation phase). `Load`/`Save`/`Validate` + JSON.
- **`render.RenderDump` (L1)** — `render/dump.go`. Builds a `*pathfind.Landscape` in memory from the
  dumped terrain grids (new `pathfind.NewMemoryLandscape` — the *only* seam needed; no
  `RenderView`/`BuildTerrain` signature change) and runs the **exact existing render path** → PNG +
  raw `[]int32`. Determinism: the per-vertex terrain-ambience speckle is now seedable from
  `Terrain.terrainSeed` (seed 0 ⇒ flat 0); **live renders are byte-for-byte unchanged** (default
  sentinel `-1`). Two independent renders of a dump produce identical SHA-256.
- **First fixture** — `testdata/rscdump/single_tile_door.json`: a 16×16 flat grass region, one
  door/wall on a tile edge, fixed camera, `terrainSeed=0`, raw colours (no GameData archive needed).
- **`cmd/renderdiff`** — `cmd/renderdiff/{main,diff}.go`. Two diff modes per §5:
  - **Pixel diff** — per-pixel max-channel Δ, differing count/%, max Δ, bounding box, and a magenta
    **heatmap PNG**.
  - **Structural diff** — `render.RenderDumpFaces` (`render/dump_faces.go`) exports the built face set
    keyed by *(transformed world-space centroid rounded to ⅛ tile, vertex count, front/back fill)* —
    camera-independent, so it answers "is this face/wall present?" directly. Any engine that emits
    `PNG + <base>.faces.json` (schema `rscdump-faces/1`) is a first-class input.

**Self-test results (wired as a permanent regression guard, `cmd/renderdiff/selftest_test.go`):**

| Check | Result | Evidence |
|---|---|---|
| Door detection (with-door vs door-removed) | **PASS** | Pixel: **2534/171008 px (1.48%)** differ, localized to region `[x 232..279, y 126..180]`, maxΔ=51. Structural: **exactly one** face present-in-door / absent-in-nodoor — the 4-vertex door quad, centroid `(10304,−96,10368)`, wood fill `−15719`. No spurious extras. |
| Zero-diff control (with-door vs itself) | **PASS** | Pixel **IDENTICAL** (0 differing, maxΔ=0); face sets identical (25282/25282). No false positives → determinism holds. |

**Artifacts** (`testdata/rscdump/out/`): `go-door.png` (the green field + brown door quad),
`heatmap_go-door_vs_go-nodoor.png` (door localized as magenta), plus `go-nodoor.png`, the
`.faces.json` sidecars, and the zero-diff control heatmap.

## Next — the two Java engines (gated, handoff documented in `cmd/renderdiff/README.md`)

The 3-way diff is then a single `renderdiff go=fixture.json deob=OUT/deob.png jar=OUT/jar.png`.

1. **DEOB engine** — *gated on the deob-compile workflow* (the `scene`/`world` packages must compile).
   A headless `DumpRenderer`: `World.loadFromDump(...)` injects the dumped grids (bypass
   `loadMapData`), pin the speckle to `terrainSeed`, `buildSection`/`addModels`,
   `setCameraOrientation`/`setBounds`, `Scene.render` → grab `Surface.pixels` (obf `ua.rb`) →
   `deob.png` + walk built models → `deob.faces.json`. First 2-way diff: **GO ↔ DEOB**.
2. **JAR engine (oracle)** — a rscplus `DumpRenderer`: inject the single-tile grid the way the L1 hook
   reads obf `k` (or replay-seek per design §3b), one `drawScene`, grab `Renderer.pixels` (the
   mirrored `ua.rb`) → `jar.png`; read the **already-built** `lb.Z[]` models → `jar.faces.json`
   (never execute obf builders — `client.vh`). This is the ground-truth oracle.

## Known constant to reconcile during the multi-engine phase

The GO engine renders with `clipFar=7000`; the deob `Scene` default is `8`→overwritten, and the jar
sets `2400` per frame. The schema **records** `clipFar` so the diff fails *loudly* on a mismatch —
reconciling the three is itself part of the faithfulness work (design §4 rule 4), not a bug in the tool.
