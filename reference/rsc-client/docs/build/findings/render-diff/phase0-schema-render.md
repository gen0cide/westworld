> _Verbatim build report from the rsc-client-render-diff-first-tool workflow (2026-06-01)._

# Phase 0 Report — rscdump/1 schema + render.RenderDump (L1) + first fixture

## Baseline
`go build ./...` was **green before** any change and **green after** every stage. Full module `go test ./...` → **ALL TESTS PASS** (render + pathfind suites included, no regressions).

## New files created (all absolute paths)
- `/home/free/code/rsc-hacking/westworld/internal/rscdump/schema.go` — the rscdump/1 schema (§4) as Go structs + JSON + `Load`/`Save`/`Validate`.
- `/home/free/code/rsc-hacking/westworld/internal/rscdump/landscape.go` — `Dump.Landscape()`: builds a `*pathfind.Landscape` from the L1 terrain grids.
- `/home/free/code/rsc-hacking/westworld/render/dump.go` — `render.RenderDump` / `RenderDumpWith` (L1 dump → PNG + raw `[]int32`).
- `/home/free/code/rsc-hacking/westworld/internal/rscdump/gen/single_tile_door/main.go` — authors the fixture programmatically (then `Dump.Save` validates it).
- `/home/free/code/rsc-hacking/westworld/internal/rscdump/gen/render_dump/main.go` — generic "render a dump JSON → PNG + stats" CLI.
- `/home/free/code/rsc-hacking/westworld/testdata/rscdump/single_tile_door.json` — **the first fixture**.
- `/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/single_tile_door.go.png` — **the output PNG**.

## Modified files (additive, low-risk; uncommitted)
- `pathfind/landscape.go` (+17) — added `NewMemoryLandscape(map[SectorKey]*Sector)`: constructs a `*Landscape` whose sector cache is pre-seeded in memory (no archive). `Tile()` reads it unchanged; a missing key returns the zero Tile (void), exactly like a missing on-disk sector. **This is the only seam needed** — no change to `RenderView`/`BuildTerrain` signatures.
- `render/render.go` (+16/−1) — pure extraction: `RenderView` now delegates to a new unexported `renderViewToSurface(...) (*Surface, error)` (identical body, returns the `*Surface` instead of `surf.PNG()`), so `RenderDump` shares the exact same build/raster path and can grab the raw `Pix`. `RenderView`'s external behaviour is unchanged.
- `render/terrain.go` (+38) — made the terrain-ambience speckle seedable (see below).

## Determinism approach for the speckle (§4 rule 2)
The vanilla client injects `(int)(Math.random()*10)-5` per terrain vertex; `render.terrainAmbience(x,y)` already replaced that with a deterministic coord-hash for live renders. I added a package var `dumpTerrainSeed int32` (default `-1` = "no dump override", so **live renders are byte-for-byte unchanged**) and `withDumpTerrainSeed(seed, fn)` which pins it for the duration of one render and restores it. In `terrainAmbience`: when `dumpTerrainSeed >= 0`, **seed 0 ⇒ flat 0** (no perturbation — the simplest fully-reproducible choice the design lists first), any other seed folds into the coord hash. `RenderDump` wraps `renderViewToSurface` in `withDumpTerrainSeed(d.Terrain.TerrainSeed, …)`. Verified: two independent renders produce **identical SHA-256** (`952658…1621`).

## Interface/refactor done to accept a dump-backed landscape
No interface change to the renderer. `RenderView`/`BuildTerrain`/`BuildBoundaries` consume the concrete `*pathfind.Landscape` and read tiles via `land.Tile(worldX,worldY,plane)`. Rather than introduce an interface (touches many call sites), I pre-seed an in-memory `Landscape` via the new `pathfind.NewMemoryLandscape`, and `Dump.Landscape()` scatters each window-local grid tile to its absolute-world sector using the same `SectorForWorld`/`TileLocalInSector` mapping the renderer reads back with. So the entire existing render path runs verbatim against dumped grids.

Colour support (§4 rule 5): the fixture uses a raw ground-palette index (grass=70 → `groundColour[70]`, a flat method305 colour, no texture archive) and `RenderDump` synthesizes a minimal `facts.Facts` giving every referenced wall id a generic openable wood-door def (`FrontDeco=-1` ⇒ flat wood fill), so the door renders with **no external GameData**. Real GameData can be passed via `RenderDumpWith`.

## Fixture description
`single_tile_door.json`: schema `rscdump/1`, level `L1`, source `hand-authored`, tick 0. A 16×16 flat grass region (`Size=16`, elevation all 0, ground all grass), window `BaseX/Y=200, Plane=0`. **One door/wall** on the east-west edge of window-local tile (8,9) = world (208,209) via `WallV[idx]=2`. No entities/scenery/items. Camera pitch=912, yaw=512 (rotation 128), distance=1500 (zoom 750), 512×334, viewDist/clipNear/clipFar=9/5/7000. `terrainSeed=0`. `Self{NoSelf:true}` (host not drawn). The renderer's own 160-tile window centres on the host tile (208,208) and pulls zeros for tiles outside the 16×16 grid.

## Output PNG + sanity stats
`/home/free/code/rsc-hacking/westworld/testdata/rscdump/out/single_tile_door.go.png` — **512×334**, valid non-empty PNG. Colour histogram (4 distinct colours):
- `#084900` dark grass green — 158176 px (terrain)
- `#7e7e7e` grey — 9786 px (window edge / off-frustum corners)
- `#3b2b17` brown — **2534 px = the wood door/wall quad** (`wallColourWood` gouraud-shaded), clearly present and distinct
- `#6080a0` — 512 px (literal sky background)

Visual confirmation: green field, brown door quad standing in the centre, grey window-edge corners — exactly the fixture.

## Design-doc deviations / notes
- **`clipFar`**: the GO engine uses `clipFar=7000` (its `scene.go` const), not the schema's example `2400`. Per §4 rule 4 this is intentionally a *pinned, diffable* value — the fixture records `clipFar=7000` to match what GO actually renders; reconciling GO=7000 vs DEOB=8/jar=2400 is explicitly flagged in the design as faithfulness work for the multi-engine diff phase. The schema carries it so the diff fails loudly on mismatch; GO's render constant itself was not changed.
- **`Camera.x/y/z`**: for the GO engine these are informational — `RenderView` derives the camera from the host tile + `rotation*4`/`zoom*2`, so `dumpView` converts `yaw→rotation` (÷4) and `distance→zoom` (÷2). The eye position is recorded for the jar/deob engines but not consumed by GO.
- L2 `Models[]` is **defined but unused** (as instructed) — the schema is complete for the rasterizer-isolation phase.

`go build ./...`: **green.**
