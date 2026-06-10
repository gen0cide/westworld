> **ARCHIVED (initial brainstorming), 2026-06-10.** This plan executed: the canonical doc is
> `docs/render-engine.md`; what shipped is `render/orsc` + `assets/`. The §5 fixed-point
> gotchas checklist remains THE reference for porting legacy integer math — cited from
> docs/lessons-learned/. Still-open items were harvested into `docs/TODO.md`. Kept verbatim
> below for the record.

# RSC Software Renderer Port Plan (Headless, Go)

> ⚠️ **SUPERSEDED — historical.** This was the original pre-implementation PLAN.
> The renderer is now built and heavily iterated; for the current architecture,
> pipeline, status, and authenticity invariants see **[render-engine.md](../../render-engine.md)**.
> Kept only for historical context (the original design intent).

Status: design / not-yet-implemented. Target: a headless port of the RuneScape
Classic software 3D renderer that, given a host's world position + a camera
rotation (and optional zoom/pitch), returns a PNG of what that host sees — an
RSCPlus-style screenshot for the westworld bot cradle.

Primary deob reference (all `client/*.java` line refs below are into this tree):
`/Users/flint/Code/rscdump.com-runescape-classic-dump/eggsampler-rsc-204-d223fc6b77db/eggsampler-rsc-204-d223fc6b77db/client/`

---

## 1. Pipeline overview

RSC is a fixed-point, integer-only **software** renderer with **no z-buffer**.
Each frame: (a) the static world is assembled once per region into a set of
`GameModel`s — a 96×96-tile terrain heightmap mesh (`World.loadSection`,
`World.java:664`), per-edge wall/door quads (`World.method422`,
`World.java:1286`), roofs, and per-instance copies of scenery `.ob3` models
placed at `tile*128` with `dir*32` yaw (`mudclient.java:5285`); (b) players /
NPCs / ground items are registered as 2-vertex billboard "poles" in a special
sprite model (`Scene.drawSprite`, `Scene.java:117`); (c) the camera is set from
the host's world pos via `scene.setCamera(x, -getElevation(x,y), y, pitch=912,
yaw=rotation*4, roll=0, distance=zoom*2)` (`mudclient.java:3760`); (d) every
model is projected — translate by `-camera`, rotate yaw→roll→pitch through a
2048-entry sin/cos table (`>>15`), then perspective-divide
`viewX=(x<<viewDist)/z` with `viewDist=9` (`GameModel.project`,
`GameModel.java:809`); (e) visible faces are collected, **painter-sorted** by
average camera-Z (`Scene.qsort`, `Scene.java:176`) then refined by a
plane-overlap pass (`Scene.method277`); (f) each face is rasterized
back-to-front into a shared `int[]` framebuffer via a per-scanline edge table
(`Scene.method281`, `:551`) + flat/gouraud/textured span fills
(`Scene.method282`, `:1036`); (g) billboards blit via `surface.spriteClipping`.
Our port replaces the final AWT `ImageProducer` handoff (`Surface.java:68`) with
a PNG encode of the framebuffer.

---

## 2. Go package / module layout

New top-level `render/` package tree under `github.com/gen0cide/westworld`
(module path confirmed in `go.mod`). It is a **read-only consumer** of existing
cradle state (`pathfind.Landscape`, `facts.*`, `world.*`) plus a new `assets`
loader. No floats in the runtime path — only the static sin/cos/normalize
tables use `math.Sin`/`Sqrt` at init.

```
assets/                 (sibling, NOT under render/ — also useful elsewhere)
  jag.go        JAG/.orsc archive container + inner index + name-hash lookup
  bzip.go       Jagex bzip2 wrapper (prepend "BZh1", feed compress/bzip2)
  model.go      .ob3 binary decoder -> assets.Model{Verts, Faces, Fills...}
  sprite.go     sprite .dat + index.dat palette decoder
  texture.go    texture .dat -> palette-indexed texels + shade ramps
  config.go     config51 string.dat/integer.dat -> object/npc/item/anim tables

render/
  surface.go    Surface{ Pix []uint32; W,H,Stride int; clip rect } + 2D blits
                + the scaled/recolour sprite blitters; PNG encode
  model.go      render.GameModel: vertex/face SoA, apply()/project()/light()/relight()
  scene.go      Scene: camera state, frustum, model list, face collect, qsort+overlap
  camera.go     setCamera orbit math + sine9[512]/sine11[2048] tables
  terrain.go    pathfind.Landscape -> terrain GameModel (heightmap mesh + overlays)
  boundary.go   walls/doors/roofs -> wall quad GameModels (method422 port)
  scenery.go    SceneryLoc -> copied/placed scenery GameModel instances
  sprite.go     entity billboard placement + drawPlayer/drawNpc composition
  rasterize.go  method281 edge table + method282..291 span fills (the inner loops)
  palette.go    anIntArray578 ground-colour palette + char colour palettes
  render.go     RenderView() faculty entrypoint (below)
```

### Cradle faculty entrypoint

```go
// render.RenderView returns a PNG of what the host at (x,y) on `plane` sees.
//   x,y     world TILE coords (cradle units; multiplied by 128 internally)
//   plane   floor index (0 for now; matches pathfind.SectorForWorld limit)
//   rotation 0..255 discrete RSC camera angle (mapped to yaw = rotation*4)
//   zoom    optional; default ~600 (-> distance = zoom*2). pitch fixed 912, roll 0.
func RenderView(land *pathfind.Landscape, f *facts.Facts, w *world.World,
                a *assets.Bundle, x, y, plane, rotation, zoom int) ([]byte, error)
```

This plugs into the cradle as a new faculty: the orchestrator already has the
host's tile position (`world/self.go` `Self.Position()` → `Coord{X,Y}`, tile
units; strip plane via `Y % PlaneHeight` since `PlaneHeight=944`), the static
landscape (`pathfind.Landscape`, lazily loaded sectors), the scenery/boundary
locs and defs (`facts.SceneryLoc`/`BoundaryLoc` + `SceneryDef`/`BoundaryDef`),
and live player/NPC positions (`world` package). `RenderView` assembles the
scene from those + decoded geometry from `assets`, rasterizes, and PNG-encodes.

### Library design: a decoupled scene snapshot (cradle / Delos / admin all consume it)

The renderer is a **standalone library**, not host-coupled. The signature above
is sugar; the CORE entrypoint takes a portable, serializable scene snapshot plus
a camera — nothing from the live cradle:

```go
// Scene is a self-contained capture of everything needed to render one view:
// the local terrain region, placed scenery + boundary instances, and visible
// entities (with appearance). Serializable (gob/json) so it can be captured in
// one process and rendered in another.
type Scene struct {
    Plane    int
    Origin   Coord          // SW tile of the captured region
    Terrain  []Tile         // heightmap + overlays for the region (from pathfind.Landscape)
    Scenery  []Instance     // {ModelID, X, Y, Dir}
    Boundary []Instance     // {DefID, X, Y, Dir} — walls/doors/fences
    Entities []EntitySprite // {X, Y, Appearance, AnimFrame} — players/NPCs/items
}
type Camera struct{ X, Y, Rotation, Zoom, Pitch int }

// Pure: no cradle, no connection, no globals. THIS is the library.
func RenderView(scene Scene, cam Camera, a *assets.Bundle) ([]byte, error)
```

Two thin adapters produce a `Scene`; everything else is the pure renderer:
- **A host renders its own view:** `render.SnapshotFromCradle(land, f, w, x, y, plane, radius) Scene` exports the host's CURRENT PERCEIVED state, then `RenderView(scene, cam, a)`. (The `RenderView(land,f,w,...)` sugar above = these two composed.)
- **Delos / admin renders ANY drone's view:** Delos obtains drone N's perceived `Scene` (N exports + ships it, or Delos reconstructs it from what it observes of N) and calls the same pure `RenderView`. A `delos viewport <drone>` command is exactly this.
- **Tooling / offline:** render an arbitrary captured `Scene` (replays, tests, the design loop) with no live server.

Crucially the snapshot is the drone's **perception** — stale tiles, things out of view, an NPC it hasn't seen move all render as that drone "sees" them. So Delos looking through a host renders the host's *belief*, limits included — a perception-fidelity tool, not a god-view. This also keeps `render` + `assets` free of any `world`/`session`/cradle import (only `SnapshotFromCradle` touches `world`), so admin/Delos binaries can link the renderer without the bot stack.

**Coordinate bridge (critical):** cradle coords are TILE units; renderer wants
world units `tile*128 (+64 to center)`. The 96×96 region window is centered the
same way `pathfind/grid.go BuildGrid` already does
(`midRegionX=(x+24)/48, baseX=mid*48-48`) — reuse that math so terrain,
collision, and the camera align. `pathfind.SectorForWorld` gives the archive
sectors to assemble.

---

## 3. Asset loading plan

We have BOTH source families on disk; choose by fidelity vs. effort.

| Family | Classic .jag (mudclient127) | OpenRSC .orsc (our server's data source) |
|---|---|---|
| models | `models14.jag` | `models.orsc` (1.1 MB) |
| sprites | `entity12.jag`, `media31.jag` | `Authentic_Sprites.orsc` (1.3 MB) |
| textures | `textures11.jag` | (in sprites/library) |
| config | `config51.jag` | OpenRSC XML/JSON (already in `facts/`) |
| landscape | `land30.jag`+`maps30.jag` (RLE .hei/.dat) | `Authentic_Landscape.orsc` (already in `pathfind/`) |

Paths:
- `/Users/flint/Code/rscdump.com-runescape-classic-dump/mudclient127/mudclient127/*.jag`
- `/Users/flint/Code/openrsc/Client_Base/Cache/video/{models,Authentic_Sprites,Authentic_Landscape}.orsc`

**Recommendation:** use **OpenRSC `models.orsc` + `Authentic_Sprites.orsc`** as
the geometry/sprite source because the cradle's def IDs and landscape are
already from OpenRSC — model NAMES match the client's `getModelIndex` scheme so
`SceneryDef.Model` drops in cleanly. Keep classic `.jag` as a cross-check
oracle. Landscape stays on our existing `pathfind` .orsc decode (do NOT
re-decode).

### Go decoders to write (in `assets/`)

1. **JAG/.orsc archive** (`Utility.java:153-237`, `GameApplet.java:428`):
   - Outer container: 6-byte header `[u24 BE uncompressedSize][u24 BE packedSize]`;
     if `packed != unpacked`, whole body is one bzip2 stream.
   - Inner JAG: `[u16 BE entryCount]` then `entryCount × {u32 nameHash, u24
     uncompLen, u24 compLen}`; payload concatenated after directory at
     `2 + count*10`. Per-entry bzip2 if `comp != uncomp`.
   - Name hash: `h = h*61 + upper(c) - 32` over UPPERCASE chars, **int32 wrap**.
   - OpenRSC `.orsc` (`DataOperations.getDataFileOffset`, `DataFileDecrypter`):
     same 6-byte header + JAG catalog; OpenRSC stores each .ob3 entry
     uncompressed (`uncomp==comp`) so per-entry decompress is a no-op — read
     `.ob3` bytes directly at the catalog offset.
2. **bzip2** (`BZLib.java`): Jagex strips the `BZh1` stream magic. **Easiest
   port:** prepend literal bytes `B Z h 1` and feed to stdlib
   `compress/bzip2`, reading exactly `uncompLen` bytes. Keep `BZLib.java` as a
   fallback to hand-port (~300 lines, self-contained BWT/MTF/Huffman) if any
   archive rejects the prepend trick. Randomized blocks never occur in RSC data.
3. **.ob3 model** (`GameModel.java:134-214`): `[u16 nV][u16 nF]`, then 3 planar
   `s16` arrays vertexX/Y/Z, `u8 faceNumVertices[]`, `s16 faceFillFront[]`
   (`32767`→magic), `s16 faceFillBack[]` (`32767`→magic), `u8 gouraudFlag[]`
   (0→flat / nonzero→`faceIntensity=magic` gouraud), then per-face index lists
   (`u8` each if `nV<256` else `u16`). All BIG-ENDIAN.
4. **sprite + index.dat** (`Surface.java:371-423`): index header `[u16 fullW]
   [u16 fullH][u8 nColours]` + `(nColours-1)*3` RGB palette (`palette[0]`→
   `0xff00ff` transparent key); per frame `[u8 tx][u8 ty][u16 w][u16 h][u8
   layout]`; pixels = palette indices, layout 0 = row-major, 1 = column-major;
   index 0 = transparent.
5. **texture** (`Scene.method300`, `:2678`): palette-indexed 64×64 (type0) or
   128×128 (type1) → `int[]` RGB; precompute 3 darker mip/shade ramps
   `c-(c>>3)`, `c-(c>>2)`, `c-(c>>2)-(c>>3)` selected by `>>>i4` lighting shift.
6. **config51** (`GameData.java:41-376`) — OPTIONAL once textures/entities are
   needed: rebuilds the authoritative `objectModelIndex` (object-id→model-name)
   and texture/animation/npc-appearance tables. For scenery we can resolve model
   names through `facts.SceneryDef.Model` directly; only verify the OpenRSC
   GameObjectDef ordering matches before trusting `DefID → model name`.

---

## 4. Phased roadmap (viewable image ASAP)

Each phase produces a PNG and is visually diffable against either a known RSCPlus
screenshot or the prior phase's output. Bit-exact diffing against the Java
client (rendering a fixed model+camera and comparing the `int[]` buffer) is the
gold standard once any phase has a Java analog — given identical integer math it
should match exactly.

### Phase 0 — Plumbing & primitives (no scene yet)
- `assets`: JAG/.orsc reader + bzip2 wrapper + name hash. Smoke test: open
  `models.orsc`, list entries, fetch a known model (e.g. `"tree"`/`"well"`).
- `render.Surface`: `[]uint32` framebuffer, `setPixel`, `drawBox`, clip rect,
  and PNG encode (`image.RGBA` + `image/png`: `r=p>>16&0xff,g=p>>8&0xff,
  b=p&0xff,a=255`). **Test:** write a solid + gradient PNG, eyeball it.
- `camera.go`: build `sine9[512]` / `sine11[2048]` tables; unit-test a few
  rotations against hand-computed `(a*sin+b*cos)>>15`.

### Phase 1 — MVP: terrain + camera + flat/gouraud scenery → PNG ⭐
**MVP definition (precise):** given `(x, y, rotation)` and default zoom/pitch,
render the **terrain heightmap mesh** (gouraud-shaded flat-colour ground from the
palette) **plus flat/gouraud-shaded scenery `.ob3` models** (no textures, no
walls, no entities, no fog), painter-sorted, to a PNG. This is the smallest
thing that looks recognizably like "standing in the RSC world."
- `render.GameModel` + `apply()`/`project()`/`light()`/`relight()`
  (`GameModel.java:704-866`).
- `scene.go`: `setCamera` orbit (`Scene.java:2290`), frustum (`method279`),
  face collect + `qsort` by depth. (Defer the `method277` overlap pass — plain
  depth sort is ~90% correct.)
- `terrain.go`: port `World.loadSection(flag)` (`:664`) consuming
  `pathfind.Landscape` tiles. Build `vertexAt(x*128, -elev*3, y*128)` for the
  96×96 window; quad-vs-2-triangle face decision; ground colour from the
  256-entry `anIntArray578` palette (`World.java:1385`). **Verify the .orsc
  `GroundElevation` is pre-accumulated** (dump a flat sector — constant bytes
  mean no RLE prefix-sum needed; just `*3`). Skip texture overlays for v1.
- `scenery.go`: for each `SceneryLoc`, `copy()` the named `.ob3`,
  `rotate(0, dir*32, 0)`, `translate(centerX, -getElevation, centerZ)`,
  `setLight(true,48,48,-50,-10,-50)` (`mudclient.java:5285`). Flat-colour faces
  only (texture id `>=0` → render as mid-grey placeholder for now).
- `rasterize.go`: port `method281` edge table + the **flat/gouraud span fills**
  `method289/290/291` (`:2162-2288`) and the colour ramp `anIntArray377`.
- **Test:** render Lumbridge-ish coords at rotation 0/64/128/192; diff against
  the prior render to confirm rotation works; eyeball terrain silhouette.

### Phase 2 — Boundaries (walls / fences / doors / roofs)
- Add `wallObjectHeight`/`frontFill`/`backFill` to `facts.BoundaryDef` by
  parsing `DoorDef.xml` `<modelVar1/2/3>` (currently DROPPED in `facts/load.go`
  — `doorDefXML` has no modelVar fields). These are `192`-ish height + 2 fills.
- `boundary.go`: port `World.method422` (`:1286`) — vertical quads from
  `Tile.HorizontalWall`(=EastWest), `VerticalWall`(=NorthSouth), `DiagonalWalls`
  (1..11999 SW-NE / 12001..23999 NW-SE / `>=48001` baked scenery). Roofs reuse
  `terrainHeightLocal`. Walls give the world its enclosed feel; big visual jump.
- **Test:** render inside a building / next to the Lumbridge castle wall; walls
  should occlude correctly (validates the depth sort under coplanar faces — may
  need the `method277` overlap pass here).

### Phase 3 — Textures
- `assets/texture.go` + `render` texture tables (`Scene.method299/300`).
- Port the **textured span fills** `method283-288` (`:1432+`) — the
  affine-within-16-texels perspective trick. Faces with fill `>=0` now sample
  real texels. Ground/walls look authentic.
- **Test:** textured ground + brick walls vs. an RSCPlus screenshot.

### Phase 4 — Entity sprites (billboards, default appearance)
- `assets`: entity sprite frames (`Authentic_Sprites.orsc` / `entity12.jag`) +
  config animation/npc tables.
- `render/sprite.go`: `Scene.drawSprite` billboard registration + the billboard
  branch of `endscene` (`:418-471`) + `surface.spriteClipping` scaled blit.
  Render players/NPCs/ground items at a fixed standing frame.
- **Test:** a player + a chicken standing on terrain, depth-scaled.

### Phase 5 — Appearance composition + recolour
- Stop discarding the 4 colour bytes at `proto/v235/updateplayers.go:185`
  (`_, _ = b.ReadBytes(4)`); add `ColourHair/Top/Trouser/Skin` to
  `event.OtherPlayerAppearance` + the `PlayerRecord`.
- Port `drawPlayer`/`drawNpc` (`mudclient.java:2930`/`2097`): direction
  selection, the `npcAnimationArray[8][12]` layer paint order, 12-slot equip
  composition, and the greyscale→body / skin recolour rule
  (`Surface.java:1530-1589`). Hardcode the colour palettes + layer tables
  (they're in client source, not assets).
- **Test:** a player wearing distinct armour with chosen colours.

### Phase 6 — Lighting / polish / determinism
- `method277` overlap pass for exact occlusion; fog (`fogZDistance≈2300`);
  per-vertex random ambience jitter — **seed or zero the RNG** for reproducible
  PNGs (`World.loadSection` uses `Math.random()*10-5`). Animation frames,
  giantcrystal transparency, door open/close model swaps.

---

## 5. Key porting gotchas (with deob refs)

1. **FIXED-POINT IS LOAD-BEARING — no floats in the runtime path.** All
   rotations are `(a*sin + b*cos) >> 15` (`GameModel.project` `:809`,
   `Scene.setCamera` `:2290`); perspective is integer divide `(x<<viewDist)/z`
   (`viewDist=9`); edge/intensity tables are 8.8 (`<<8`, `Scene.method281`
   `:551`); scaled-sprite/texture coords are 16.16 (`<<16`). Use **`int32`** (Java
   `int`) and let it wrap — the cross products in `relight()`
   (`GameModel.java:741-773`) and texture-gradient numerators in `method282`
   (`Scene.java:1041`) **intentionally overflow 32 bits**; using `int64` warps
   the texture mapping differently. Do NOT "fix" it.
2. **`>>` is arithmetic on signed.** Java `>>` and Go `>>` on signed ints both
   sign-extend — keep types signed (`int32`, not `uint32`) through projection.
   `>>>` in Java (texel mip shift `tex[...] >>> i4`) is the **unsigned** shift —
   use `uint32` for that specific fetch.
3. **BIG-ENDIAN, signed shorts, 24-bit sizes.** `.ob3`, JAG catalog, sprite
   headers are all big-endian (`Utility.getUnsignedShort` `:36` = `(b[o]&0xff)
   <<8 | b[o+1]&0xff`). Vertex X/Y/Z and faceFill are **signed** 16-bit
   (`getSignedShort` `:48`). The 3-byte sizes need a hand-rolled `readU24` (do
   NOT `binary.BigEndian.Uint32`).
4. **Signed-byte trap is INVERTED in Go.** Java `byte` is signed so the client
   masks `data[i] & 0xff` everywhere; Go `[]byte` is already unsigned, so
   `int(b)` is correct and you DROP the mask — but watch RLE branches that
   compare `val >= 128` (`World.loadSection` `:191`) and the signed `s8` walls
   `wallsNorthsouth/Eastwest`.
5. **Name hash overflow.** `h = h*61 + c - 32` MUST be `int32` arithmetic to
   match Java's wrapping `int` (`Utility.getDataFileOffset` `:153`).
6. **Magic sentinels — replicate exactly.** `faceIntensity == 0xBC614E`
   (12345678) ⇒ per-vertex gouraud; `.ob3` faceFill `32767` ⇒ replace with the
   magic (no-fill) sentinel; flat colour faces encode `fill = -1 - colour`
   where colour is 5:5:5 RGB (`Scene.method305` `:2768`); transparent sprite key
   `0xff00ff`, internally remapped (`0x000000→1`, `0xff00ff→0`,
   `Surface.preloadSprite` `:529`).
7. **`relight()` 65535-vs-65536 quirk** (`GameModel.java:766`): `faceNormalZ`
   normalizes with `65535` while X/Y use `65536`. Copy the asymmetry for
   pixel-accuracy.
8. **Camera angle inversion.** `setCamera` stores `cameraPitch/Yaw/Roll =
   (1024 - angle) & 0x3ff` for `project()`, but builds the orbit POSITION offset
   from the RAW angle (`Scene.java:2290`). The camera sits `distance` units
   BEHIND the look-at, not at it. Easy to get backwards. Yaw input is
   `rotation*4` (0..255 → 0..1020 in the 1024-step `sine11` space).
9. **Axis trap.** `setCamera(x, z, y, ...)`: 2nd arg `z` is HEIGHT (vertical),
   3rd `y` is world-south. RSC Y points DOWN — heights are stored NEGATED
   (`vertexY = -getTerrainHeight`). World unit = `tile*128` (`magicLoc=128`).
10. **No z-buffer.** Occlusion is `qsort` by depth (`Scene.java:176`) + the
    `method277` plane-overlap refinement. A naive z-buffer diverges at
    coplanar/overlapping faces (walls vs floor). Port the sort; defer
    `method277` to Phase 2/6 only if artifacts appear.
11. **bzip2 header strip.** Stdlib `compress/bzip2` needs the `BZh` magic Jagex
    strips — prepend `BZh1` (`BZLib` hardcodes blocksize100k=1).
12. **interlace OFF.** Drop the `vertInc=2` half-scanline mode and all AWT
    `ImageProducer`/`setcomplete` (`Surface.java:68`); skip mouse-picking
    branches (`getObjectsUnderCursor`).
13. **Determinism.** Seed/zero the terrain ambience RNG
    (`World.loadSection` `Math.random()*10-5`) and fix entity animation frames so
    the same `(x,y,rotation)` always produces an identical PNG.

---

## 6. Cradle: have vs. must-add

**HAVE (reuse directly):**
- `pathfind/landscape.go` — per-tile decode: `Tile{GroundElevation,
  GroundTexture(=ground colour idx), GroundOverlay, RoofTexture, HorizontalWall
  (=EastWest), VerticalWall(=NorthSouth), DiagonalWalls int32}`; lazy sector
  load from `Authentic_Landscape.orsc`; `SectorForWorld` archive math. This is
  the terrain mesh + static boundary INPUT.
- `pathfind/grid.go` — the 96×96 / 2×2-sector region centering
  (`midRegionX=(x+24)/48, baseX=mid*48-48`) the renderer reuses.
- `facts/defs.go` — `SceneryDef{Model name, Width, Height, Type}`,
  `BoundaryDef{DoorType, Unknown=openable}`, `SceneryLoc`/`BoundaryLoc
  {DefID,X,Y,Direction}`. The WHERE/WHICH-model for static objects.
- `world/self.go` — host tile position + `PlaneOf` (`PlaneHeight=944`).
- `world/world.go` + `proto/v235/updateplayers.go` — live `PlayerRecord`
  (`EquipBySlot[12]`/`WornSprites`, combat, skull, X/Y), `NpcRecord{TypeID,X,Y}`
  — the billboard inputs.

**MUST ADD:**
- The entire `assets` layer: JAG/.orsc + bzip2 reader, `.ob3` decoder, sprite +
  index.dat decoder, texture decoder, (optionally) config51 reader. ZERO of this
  exists in Go today (no `image`/`bzip2`/`faceFill`/`vertexX`/`sine9` anywhere).
- The entire `render` layer: `Surface` framebuffer + PNG encode, `GameModel`
  transform/light/project, `Scene` camera+frustum+sort, the terrain/wall/scenery
  mesh builders, the rasterizer span fills, the sin/cos tables.
- `facts.BoundaryDef` is MISSING wall geometry params — add `WallHeight`,
  `FrontFill`, `BackFill` from `DoorDef.xml` `<modelVar1/2/3>` (the `doorDefXML`
  struct in `facts/load.go` currently has no modelVar fields → DROPPED).
- The 4 player appearance COLOUR bytes are read then discarded at
  `proto/v235/updateplayers.go:185` — surface them onto
  `event.OtherPlayerAppearance` + `PlayerRecord` for Phase 5 recolour.
- NPC appearance data (`npcSprite[12]`, `npcColour*`, billboard W/H
  `npcSomething_1/2`) — not in OpenRSC `NpcDef` JSON; parse from config51 or
  extend the def loader (Phase 4).
- Coordinate scaling tile→world-units (`*128 +64`) and plane-Y stripping
  (`Y % 944`) at the `RenderView` boundary.
- The `anIntArray578` ground-colour palette + the hardcoded char colour/layer
  tables (`World.java:1385`, `mudclient.java:7161-7301`) — port as Go literals.

---

## 7. Summary

**MVP (Phase 1):** `RenderView(x, y, rotation)` → PNG of the terrain heightmap
mesh (gouraud flat-colour ground) + flat/gouraud-shaded scenery `.ob3` models,
painter-sorted, default zoom/pitch — no textures, no walls, no entities. Smallest
output that is recognizably "the RSC world from here."

**Phases:** 0 plumbing/primitives → **1 MVP terrain+scenery** → 2 boundaries/
roofs → 3 textures → 4 entity billboards → 5 appearance composition+recolour →
6 lighting/fog/polish/determinism. Each emits a PNG and is visually diffable.

**Biggest risks:**
1. **Fixed-point fidelity** — the deliberate `int32` overflow in `relight()` and
   texture gradients; using wider ints silently warps output. Verify by
   bit-diffing a fixed model+camera against the Java client.
2. **bzip2 decode** — if the `BZh1`-prepend trick fails on the OpenRSC `.orsc`
   variant, we hand-port `BZLib.java` (~300 lines). Validate FIRST in Phase 0.
3. **Painter's-algorithm occlusion** — plain depth sort gets ~90%; coplanar
   walls/floors need the fiddly `method277` plane-overlap pass.
4. **`.orsc` elevation encoding** — confirm `GroundElevation` is pre-accumulated
   (just `*3`) vs. needing the classic RLE seed-64 prefix-sum + `*2`.

**Single best first task:** Write `assets/jag.go` + `assets/bzip.go` and prove
it by opening `/Users/flint/Code/openrsc/Client_Base/Cache/video/models.orsc`,
listing its catalog, and decoding one `.ob3` (`assets/model.go`) into a
`GameModel` struct with sane vertex/face counts. This de-risks the #2 risk
(bzip2) and #1 (big-endian/signed decode) immediately, and unblocks every later
phase — nothing renders until geometry loads.
