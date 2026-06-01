# RSC render-fidelity findings (TIER 3 collation + applied fixes)

This document collates every render-fidelity divergence found by the TIER-1/2
audits of the Go render library (`render/*.go`) against the deobfuscated RSC
client source (the chosen render oracle:
`reference/rsc-client/src/client/{scene,world}/*.java`), ranks them, records the
**HIGH-confidence fixes that were APPLIED** (with before/after evidence), and
lists the **PROPOSED (unapplied)** divergences with the patch and why each is
gated.

Method: GROUND TRUTH is the deob client source. Each divergence was found by
(1) reading the deob algorithm, (2) rendering the Go engine on a differential
fixture (`testdata/rscdump/hunt/`) and inspecting the PNG + structural face set
via `cmd/renderdiff`, (3) comparing GO-actual to the deob spec. The DEOB and JAR
render engines are not yet runnable (gated on the deob-compile workflow), so
GO-vs-JAR **pixel** diffs are not yet possible; the structural face-set diff and
the deob source are the oracle.

All fixes verified by re-rendering the bug's fixture and confirming GO now
produces the deob-spec geometry. Regression guards live in
`render/fidelity_{terrain,scenery,walls,diagobj}_test.go` and stable fixtures
under `testdata/rscdump/hunt/{t3_*,diagobj_door}.json`. The committed door
self-test (`go test ./cmd/renderdiff`) still passes unchanged.

---

## 1. Ranked divergence table (deduped)

Severity: **PB** = perception-breaking, **C** = cosmetic, **L** = latent (no
current visual defect). Status: **APPLIED** / **PROPOSED** (unapplied, gated).

| # | id | subsystem | symptom | deob src | Go src | severity | conf | status |
|---|----|-----------|---------|----------|--------|----------|------|--------|
| 1 | terrain-water-flatten-overlay-id (= bridge_flatten_overlay2) | terrain heightmap / vertex flatten | overlay-2 river / remapped-250 bridge deck sunk to y=0 (river trench) | World.java:891-895 | terrain.go:128-149 | PB | high | **APPLIED** |
| 2 | terrain-type2-floor-split | ground colour-split | indoor floors (chapel/bank/house) chamfered with grass-wedge corners | World.java:932 | terrain.go:279-308 | PB | high | **APPLIED** |
| 3 | wall-colour-table-ignored (+ wall-front-back-collapsed) | boundary fill (method422) | snow/timber/cavern/window walls all render grey stone; front=back forced | World.java:744-752 | boundary.go:76-125 | PB | high | **APPLIED** |
| 4 | scenery-id74-no-vertical-lift | scenery placement | windmill sails (id 74) lie on the ground, not atop the mill tower | Mudclient.java:6229-6231 | scenery.go:137 | PB | high | **APPLIED** |
| 5 | terrain-split-id-vs-class | ground colour-split neighbour compare | two diff. overlays of the SAME tileType class wrongly split at a concave corner | World.java:933-942 | terrain.go:281-292 | C | high(div) | **APPLIED** |
| 6 | terrain-overlay-no-neighbour-spread | overlay-triangle pass | water body shows grass gaps at its border (1 quad emitted vs deob's 5) | World.java:1117-1139 | terrain.go:334-352 | C | high | **APPLIED** |
| 7 | roof-cull-hardcoded-set | under-roof cull (signal source) | DRY: hardcoded type-2 id set duplicates `tileDefs` (currently equal) | World.java:945 | roof.go:31-37 | L | high | **APPLIED** |
| 8 | door-diag-object-not-built (= scenery-diag-object-not-built) | scenery / World.addModels | diagonal doors/objects (id 48001..59999) render as NOTHING | World.java:792-820 | render/diagobj.go, render.go:330-348, schema.go:127, pathfind/landscape.go:39 | PB | high(div) | **APPLIED** |
| 9 | wall-visible-gate-missing | boundary visibility gate | openable doors/doorframes drawn as solid panels at load (deob skips) | World.java:1005/1014/1023 | boundary.go:135-167 | PB | high(div)/med(intended?) | **PROPOSED** (GO behaviour deliberate for the bot) |
| 10 | roof-cull-singleplane | under-roof cull | upper-plane indoor floor not detected (Lumbridge arch) → roof stays drawn | World.java:842-846,945 | roof.go:51-56 | PB | med | **PROPOSED** (needs multi-plane dump schema) |
| 11 | roof-cull-no-plane0-gate | under-roof cull (gating) | upper-floor host keeps its own roof; deob removes it unconditionally | Mudclient.java:7125-7147 | render.go:208,231 | C | med | **PROPOSED** (needs multi-plane schema) |
| 12 | roof-multistory-flat-storybase | roof multi-story height anchor | upper-story roof floats at flat 192×plane, not the carried-over accumulated height | World.java:994-997 | roof.go:450-452,213 | PB(multi-story) | med | **PROPOSED** (needs multi-plane schema) |
| 13 | roof-maxplane-off-by-one | roof plane range | builds a plane-3 roof the deob never builds | World.java:844-846 | roof.go:65 | L | med | **PROPOSED** (latent; couple with #12) |
| 14 | roof-walltop-max-vs-first | roof method428 corner height | shared eave corner uses MAX wall height, deob locks the FIRST in scan order | World.java:725-731 | roof.go:226-230 | C | med | **PROPOSED** (synthetic facts force uniform height) |
| 15 | bridge_seam_coord | terrain overlay remap (setTiles) | 250→9 plank-edge keyed on world `wx%48==47`, deob uses region-local `x==47` | World.java:672-674 | terrain.go:120-122 | C | med | **PROPOSED** (needs sector-aligned real-map dump) |
| 16 | bridge_shoreline_bleed | terrain colour-split | invented water-corner branch with no deob counterpart | (none) | terrain.go:296-307 | C | med | **PROPOSED** (mostly subsumed by #1) |
| 17 | wall-light-params | boundary lighting | wall ambient/diffuse/dir differ from the deob wall model | World.java:1046 | boundary.go:209 | C | high(div)/intentional | **PROPOSED** (self-documented brighten) |
| 18 | grounditem-no-scenery-lift | ground-item sprite placement | item on a table/altar/range drawn sunk into the object, not on top | Mudclient.java:7232-7233,4235-4242 | render.go:510 | C | high | **PROPOSED** (needs schema lift field) |
| 19 | sprite-foot-corner-not-interpolated | entity/item foot-Y anchor | sprite feet anchored to SW tile corner, not bilinear tile-centre (slope float/sink) | World.java:391-410, Mudclient.java:7222/7233 | render.go:400,626 | C(PB steep) | high | **PROPOSED** (anchor-grid change; verify w/ JAR) |
| 20 | sprite-occlusion-zbuffer-vs-paintersort | sprite/scene occlusion | GO uses depth-buffer+64 bias; deob unifies sprites+faces in one painter sort | Scene.java:1660-1820 | render.go:533, surface.go:115 | unknown | low | **PROPOSED** (architectural; needs JAR pixel oracle) |

### Deob-source-vs-oracle discrepancies where the GO is ALREADY correct (do NOT "fix")

These are NOT Go bugs — they are transcription artifacts in the deob source
itself where the Go correctly follows the authentic rev-204 oracle. Listed so a
future agent does not "correct" the Go backwards.

| id | what | Go is correct because |
|----|------|------------------------|
| disc-deob-qsort-inverted | deob `Scene.java:945` face-sort partition is ascending (near-first) | Go `scene.go:188` sorts descending (far-first), matching `mudclient204:1131` |
| disc-deob-overlap-args-interleaved | deob `Scene.java:1153` passes `polygonsOverlap(aX,bY,bX,aY)` interleaved | Go `facesort.go:300` passes clean `(Xa,Ya,Xb,Yb)`, matching `mudclient204:2639` |
| disc-deob-ramp-mask-corrupt | deob `Scene.java:2089` red mask is `& 32025` | Go `palette.go:77` uses `>>10 & 0x1f`, matching `mudclient204:2270` |
| deob-terrainColours-channel-scramble | deob `World.java:173-180` ground-colour ramp has RGB channels scrambled (254/256 entries differ) | Go `palette.go:124-137` ramp is byte-correct vs mudclient204/OpenRSC. **DO NOT change `palette.go` to match the deob ramp.** |
| fog-index-vs-skyblend | deob fades distant faces toward ramp-black via the shade index | Go blends toward `skyColour` in the span filler — a documented, intentional choice (affects distant colour only, not geometry/sort) |

---

## 2. Applied fixes (before / after + verifying evidence)

All applied fixes are HIGH-confidence: the deob source is unambiguous AND a
fixture's structural evidence confirms the corrected behaviour.

### FIX 1 — terrain-water-flatten-overlay-id (= bridge_flatten_overlay2) [PB]

**Deob (World.java:891-895):** a terrain vertex is forced to y=0 **iff** a
touching tile's `getTileTypeOnPlane == 4`. The genuine type-4 ids are 4/12/20/21
(`palette.go tileDefs`). Overlay **2** (the river AND the remapped-250 bridge
interior) and overlay **11** (lava) are tileType **3** — NOT flattened; they
render as a textured type-3 overlay floor at real elevation.

**Before (`render/terrain.go`):** the flatten was keyed on the overlay **id**:
`if ovl[i][j] == waterOverlay(2) || ovl[i][j] == waterOverlay2(11) { water = true }`.
So every overlay-2/11 tile sank to y=0 — the 250→2 remapped bridge deck plunged
into the river trench.

**After:** flatten + the deck flag are the SAME condition, keyed on
`overlayDef(ovl).tileType == 4`:
```go
if def, ok := overlayDef(ovl[i][j]); ok && def.tileType == 4 {
    water[i][j] = true // flatten the 4 shared vertices to y=0
    deck[i][j] = true
    deckCount++
}
```

**Evidence (`testdata/rscdump/hunt/t3_water_ov2.json`, elev 40 ⇒ world height 120):**
- BEFORE: 64 water-textured (fillFront=1) faces all at centroid **Y=0** (sunk).
- AFTER: **0** faces at Y=0; the overlay-2 patch renders as **64 texture-1 faces
  at centroid Y=-128** (its real elevation), flush with the banks.
- Genuine type-4 band (`t3_bridge_ov4.json`) STILL does the double-render: 80
  flattened water faces at Y=0 **and** 80 deck (planks colour 3) faces at raw
  elevation Y=-128.
- Guarded by `TestFidelity_WaterFlattenByTileType` + `TestFidelity_Type4StillFlattens`.

### FIX 2 — terrain-type2-floor-split [PB]

**Deob (World.java:932):** the colour-split `else if` arm is gated
`decoType != 2 || (getWallDiagonal>0 && <24000)` — a tileType-2 indoor floor with
no interior diagonal wall is **never split**, so its corners stay square.

**Before:** the GO split fired for **any** `hasOverlay` tile (no `tileType != 2`
gate), so an indoor floor's boundary corners reverted a triangle to grass.

**After (`render/terrain.go`):**
```go
splitEligible := hasOverlay
if def, ok := overlayDef(ovl[i][j]); ok && def.tileType == 2 {
    diagWall := int(land.Tile(baseX+i, baseY+j, plane).DiagonalWalls)
    splitEligible = diagWall > 0 && diagWall < 24000
}
if splitEligible { /* colour-split */ } else if !hasOverlay { /* shoreline */ }
```

**Evidence (`t3_indoor_split.json`, a 7×7 planks block in grass):**
- BEFORE: **4** grass-revert triangles at the four extreme corners (chamfered).
- AFTER: **0** grass-revert triangles inside the patch; **49 solid planks quads**
  (the full 7×7). Total face count now equals the all-grass control (25331), i.e.
  the floor adds no split triangles — exactly the deob (49 quads, 0 triangles).
- Guarded by `TestFidelity_IndoorFloorNotChamfered`.

### FIX 5 — terrain-split-id-vs-class [C]

(Applied together with FIX 2; numbered #5 in the table.)

**Deob (World.java:933-942):** the split neighbour compare uses
`getTileType(x,y)` (World.java:288-292) = **-1** (no deco) / **1** (tileType==2)
/ **0** (else) — NOT the raw overlay id. Two different same-tileType-3 overlays
both fold to class 0 ⇒ "equal" ⇒ no split.

**Before:** the GO compared the raw overlay id (`ovlClassAt(nb) != me`), so id 7
vs id 19 (both type 3) compared unequal ⇒ a spurious grass wedge.

**After:** a `tileTypeClass(a,b)` helper folds through the deob's getTileType
classes; the four neighbour tests use it.

**Evidence (`t3_class_corner.json`, id 7 field + concave L of id 19 at tile 12,12):**
- BEFORE: a grass-revert triangle at the concave corner.
- AFTER: **0** grass wedges at tile (12,12). Positive control
  (`terrain_road_split.json`, type-1 road vs grass) STILL splits (50 grass-revert
  triangles) — a genuine class boundary is preserved.
- Guarded by `TestFidelity_SameClassCornerNotSplit`.

### FIX 6 — terrain-overlay-no-neighbour-spread [C]

**Deob (World.java:1117-1139 `buildOverlayTriangles`):** for every interior tile,
emit a type-4 overlay quad on the tile if it is type-4, **else if** the tile is
NOT type-3, emit a type-4 overlay quad at THIS tile for each N/S/E/W neighbour
that is type-4. An isolated type-4 tile in grass ⇒ **5** overlay quads.

**Before:** the GO deck pass emitted **one** quad per type-4 tile only; the
neighbour-spread arm was absent (1 quad where the deob emits 5).

**After (`render/terrain.go`):** the model is pre-sized with a `spreadCount`, and
a second pass after the deck pass emits the neighbour-spread quads at raw height
using the neighbouring type-4 tile's colour.

**Evidence (`t3_type4_neighbour.json`, single id-4 tile in grass):**
- BEFORE: 1 type-4 overlay quad.
- AFTER: **5** overlay quads — the tile (10304,10304) + 4 cardinal neighbours
  (10176,10304),(10304,10176),(10304,10432),(10432,10304).
- Guarded by `TestFidelity_Type4NeighbourSpread`.

### FIX 3 — wall-colour-table-ignored (+ wall-front-back-collapsed) [PB]

**Deob (World.java:744-752, method422):** `frontColour = wallFrontColour_v_a[id]`
(= modelVar2 / FrontDeco), `backColour = wallBackColour_Jk[id]` (= modelVar3 /
BackDeco), passed **verbatim** to the face. A value <0 is a flat 5:5:5 colour;
>=0 is a texture id.

**Before (`render/boundary.go addWall`):** initialised front/back to a hardcoded
stone grey (`-16914`), flipped to wood for `Unknown!=0` or a fence/gate name, and
only used the def's own fill when a texel buffer existed. With no texture archive
(the live case), every wall collapsed to grey/wood and front was forced equal to
back.

**After:** a `resolveFill` helper ports method422 exactly:
```go
resolveFill := func(deco int) (fill int32, textured bool) {
    if deco < 0 { return int32(deco), false }            // flat colour from the def
    if textureBuffer(int32(deco)) != nil { return int32(deco), true } // textured span
    return textureFill(int32(deco)), false               // baked per-id flat colour
}
front, textured = resolveFill(def.FrontDeco)
back, _ = resolveFill(def.BackDeco)
```
The `Unknown`/name material heuristic is dropped. `render/dump.go syntheticFacts`
was updated to carry the flat WOOD colour (`wallColourWood`) in FrontDeco/BackDeco
so the synthetic door fixture still paints a wood leaf with no archive (the
`-15719` the door self-test asserts).

**Evidence (`walls_v_snowwall.json`, rendered with real `DoorDef.xml` via
`cmd_wallrender`):**
- BEFORE: snowwall wall face front/back = **-16914** (stone grey).
- AFTER: snowwall wall face front/back = **-31711** (the def's snow-white flat
  colour). Other materials now read distinctly: `wall`/`highwall` -15855 (tex 2),
  `window` -13743 (tex 5), `timberwall` -18923 (wood tex 21); heights honoured
  (highwall centroid Y=-144, battlement Y=-32).
- The committed door self-test is **unchanged** (door still renders at -15719,
  2534 px / 1 struct face).
- Guarded by `TestFidelity_WallColourTable` (skips if `DoorDef.xml` is absent).

### FIX 4 — scenery-id74-no-vertical-lift [PB]

**Deob (Mudclient.java:6229-6231):** immediately after placing a static object,
`if (objType == 74) model.a(0, 0, -480, true)`. The 3rd translate arg → baseZ →
the vertical Y axis (`GameModel.apply`), and -Y is up, so the windmill sails are
lifted 480 world-units onto the mill tower.

**Before (`render/scenery.go PlaceScenery`):** no id-74 case — every object
anchored at `cy = -elevationAt(...)`.

**After:**
```go
cy := -elevationAt(land, baseX, baseY, cx, cz, plane)
if loc.DefID == 74 {
    cy -= 480 // windmill sails float 480 up (Mudclient.java:6229-6231)
}
```

**Evidence (`scenery_y_probe`, windmillsail at id 74 vs id 900):**
- BEFORE: both at face-centroid Y range `[-288 .. 288]` (identical — no lift).
- AFTER: id 74 at `[-768 .. -192]`, id 900 at `[-288 .. 288]` — a clean **-480**
  shift up.
- Guarded by `TestFidelity_WindmillSailsLifted` (skips if `models.orsc` absent).
- Follow-up (NOT this fix): the deob also continuously yaw-rotates id 74 each
  frame (Mudclient.java:3347) — live animation, separate from static placement.

### FIX 7 — roof-cull-hardcoded-set [L, maintainability]

**Deob (World.java:945):** `objectAdjacency |= 0x80` exactly when
`tileType_da_N[deco-1] == 2`. One source of truth: the tile-type table.

**Before (`render/roof.go`):** `underRoofOverlay` hardcoded the id set
`{3,5,6,13,14,15,16,17,18,23}` — a hand-copied snapshot duplicating
`palette.go tileDefs`.

**After:** derived from the single table:
```go
func underRoofOverlay(ov byte) bool {
    def, ok := overlayDef(ov)
    return ok && def.tileType == 2
}
```

**Evidence:** the hardcoded set is byte-identical to `tileDefs` type-2 ids today
(verified programmatically: no MISSING / EXTRA), so this is a pure DRY refactor
with **no visual change**. The cull still fires identically
(`roof_cull_grass.json` → 13 roof faces; `roof_cull_indoor.json` → 0).

### FIX 8 — door-diag-object-not-built (= scenery-diag-object-not-built) [PB]

**Deob (World.java:792-820, `addModels`):** for every tile with
`48000 < getWallDiagonal(x,y) < 60000`, the value is NOT a wall — it is a
DIAGONALLY-PLACED scenery object (INCLUDING a diagonal door). The deob
`objectId = diag - 48001`, reads `dir = getTileDirection(x,y)`, takes the
object's footprint `(w,hgt)` from `objectWidth/objectHeight` (swapped for dir
0/4), clones the prototype model, translates it to the footprint CENTRE
`cx=128*(w+2x)/2, cz=128*(hgt+2y)/2` at `-getElevation(cx,cz)`, orients it by
`dir*32` about the vertical axis, lights it, registers it with the scene, then
clears the diagonal id over the footprint. Corroborated by OpenRSC
`Client_Base/.../graphics/three/World.java addLoginScreenModels` (`>48000 &&
<60000`, `diagWall = diag-48001`, `getObjectDef(diagWall).modelID`,
footprint-centre translate, `setRot256(0, tileDirection*32, 0)`).

**Before:** the Go had **no addModels analog**. The 48000+ band is dropped by the
wall builder (`boundary.go: d < 24000`) — correct, it is not a wall — but nothing
ever built it as an object, so a diagonal door rendered as **ZERO geometry**
("a door the cradle can't perceive"). Structural diff `door_diag_obj.json` vs
`door_none.json` was IDENTICAL.

**After:** new `render/diagobj.go BuildDiagonalObjects` ports `addModels`. It
scans the window's `DiagonalWalls` grid for the 48001..59999 band and routes each
object through the **unchanged, verified-faithful** `render/scenery.go
PlaceScenery` via a synthesized `facts.SceneryLoc{DefID: diag-48001, X, Y,
Direction: getTileDirection}` — so the footprint-centre, the dir 0/4 width/height
swap, the `dir*32` roll orientation and the bilinear `-elevation` ground-snap are
the **same code** the static scenery pass uses (single source of truth). A
hand-authored fixture with no archive falls back to a synthetic wood door-leaf
panel built with the identical placement math (mirroring how `dump.go
syntheticFacts` supplies a generic wood boundary def). The footprint-clear over a
multi-tile object (World.java:811-816) is applied to a LOCAL snapshot of the
diagonal grid, so the shared `pathfind.Landscape` (read concurrently by the
terrain/boundary passes) is never mutated. Wired into `buildScene`
(`render.go:330-348`) right after the static/dynamic scenery loops, gated by the
same `RENDER_NO_SCENERY` switch.

**Schema addition (the gate that was blocking this fix):** the orientation needs
`getTileDirection`, which the `rscdump/1` schema did not carry. A `TileDirection
[]byte` grid was threaded additively + backward-compatibly through
`internal/rscdump/schema.go` (`Terrain.TileDirection`, json `tileDirection,
omitempty`; an omitted grid ⇒ all-zero ⇒ dir 0 = prior behaviour),
`pathfind.Tile.TileDirection`, and `internal/rscdump/landscape.go`
(`Dump.Landscape()` injects it per tile). The on-disk 10-byte `.orsc` sector
record carries no direction byte, so `decodeSector` leaves it 0 — only the dump
path populates it.

**Evidence (`testdata/rscdump/hunt/diagobj_door.json`, a single diagonal-door
object `wallDiag 48001` at world tile (208,209), dir 0, on flat grass; control
`door_none.json`; structural diff via `cmd/renderdiff`):**
- BEFORE: GO built **0** faces for `door_diag_obj` — structurally IDENTICAL to
  the empty control.
- AFTER: GO builds **+1** upright wood door-leaf the control lacks
  (`diagobj=25282 faces, none=25281, 25281 shared`): a 4-vert quad, fill
  front=back=**-15719** (the flat WOOD colour — the same `-15719` the committed
  door self-test asserts), centroid **(10304, -96, 10432)** = the footprint
  centre of tile (208,209), a full 192-tall upright leaf on elevation-0 terrain.
- PNG sanity: the leaf is **visible** (PIXEL diff vs control: **2691/171008 px,
  maxΔ=67**) — an upright door panel standing on the green flat terrain.
- Straight-door control (`door_straight.json`, a WallV door) builds the matching
  +1 leaf (centroid (10304,-96,10368), -15719, **2534 px** visible) — the
  diagonal object builds the SAME kind of door geometry as a straight door.
- Note: the synthetic fallback leaf is a flat single quad; rolled 90° (dir 2) it
  is edge-on to the top-down camera (zero projected area), which is why the
  fixture uses dir 0 so the panel faces the camera. A real RSC door is a 3D mesh
  (the real-def path routes through `PlaceScenery` and is visible at any dir);
  the `dir*32` roll is centroid-preserving (verified: dir 0 and dir 2 give the
  identical centroid (10304,-96,10432)).
- Guarded by `TestFidelity_DiagObjectBuilt` + `TestFidelity_DiagObjectVsStraightDoor`.

---

## 3. Proposed (unapplied) divergences — patch + why gated

Each is documented for a future pass. None were applied because the deob spec is
not fully pinned by an available fixture (needs the JAR/DEOB pixel oracle or a
multi-plane dump schema), OR the current GO behaviour is a deliberate design
choice for the bot.

### wall-visible-gate-missing [PB divergence / intended]
The deob standing-wall loop builds a wall only when `wallVisible/unknown==0`
(World.java:1005/1014/1023); openable doors/doorframes are drawn separately as
server-placed entity models. The Go builds a panel for every non-zero wall id
(no gate), so closed doors/doorframes appear as solid panels at region load.
**Patch:** in `buildStory`, `if def.Unknown != 0 { return }` before `addWall`.
**Why gated:** this is **deliberate GO behaviour** — TIER-1 framed rendering the
closed door at load as intentional (the bot wants to perceive the door), and the
committed door self-test depends on the synthetic openable door rendering. The
faithful alternative is to render a real door *entity* model rather than a wall
slab; that is a larger change tied to the dynamic-door path. Left as-is.

### roof-cull-singleplane / roof-cull-no-plane0-gate / roof-multistory-flat-storybase / roof-maxplane-off-by-one
All four are multi-plane roof issues. The deob OR's the 0x80 under-roof bit
across planes 0/1/2 (World.java:842-846,945), builds only planes 0/1/2
(maxRoofPlane should be 2, not 3), gates the cull on `yj==0`
(Mudclient.java:7125-7147), and anchors upper-story roofs on the carried-over
accumulated `terrainHeightLocal` (World.java:994-997), not a flat 192×plane.
**Why gated:** the `rscdump/1` schema carries only ONE plane's `Terrain` grids
(`schema.go Window.Plane`; `landscape.go:23` scatters every tile into a single
plane), so upper planes are always void in a dump and these paths can't be
exercised or pixel-confirmed. Story-0 builds are confirmed correct. These need a
multi-plane dump schema (and ideally the JAR oracle on a real Lumbridge column)
before applying. The plane-3 → plane-2 cap and the per-plane height accumulation
should be applied together.

### roof-walltop-max-vs-first [C]
The deob locks a shared eave corner to the FIRST wall in scan order via the
`<80000` sentinel (World.java:725-731); the Go uses MAX of touching walls.
**Why gated:** `render/dump.go syntheticFacts` gives every wall id a uniform
Height 192, so MAX==FIRST and the divergence is unobservable through the dump
path. Needs non-uniform `BoundaryDef.Height`, which the dump JSON can't express.
The proposed "first-wins lock" is a no-op for uniform heights so it cannot
regress the goldens, but cannot be positively verified yet either.

### bridge_seam_coord [C] / bridge_shoreline_bleed [C]
`bridge_seam_coord`: the 250→9 plank-edge test uses world `wx%48==47`; the deob
uses the region-local seam `x==47` (World.java:672-674), which is only equivalent
on a sector-aligned window. The GO render window is host-centred (not
sector-aligned), so the plank lip lands on wrong columns.
`bridge_shoreline_bleed`: the GO `terrain.go:296-307` "shoreline" branch makes
grass corner-triangles WATER with no deob counterpart.
**Why gated:** both are 1-tile cosmetic edge effects whose "right" behaviour on
the non-sector-aligned 160-window is a design choice the JAR pixel diff should
pin down. `bridge_shoreline_bleed` is also largely subsumed now that FIX 1 keeps
overlay-2 out of the `water` set (the branch fires far less). Left for the pixel
oracle.

### wall-light-params [C, intentional]
The Go wall light `SetLight(60,24,-50,-10,-50)` differs from the deob wall model
`setLight` (ambient 24, diffuse -10, dir (-50,60,-50)) at World.java:1046.
**Why gated:** the Go comment self-documents this as a deliberate brighten
("shadows super dark"). It is a faithfulness/citation note, not a defect; the
deob-faithful value is `SetLight(24,-10,-50,60,-50)` if byte-fidelity is later
required.

### grounditem-no-scenery-lift [C] / sprite-foot-corner-not-interpolated [C/PB-steep]
`grounditem-no-scenery-lift`: the deob lifts a ground item by `Le[i] =
objectElevation` of co-located scenery (Mudclient.java:7232-7233) so it rests on
furniture; the GO passes a 0 offset and `GroundItemMarker` has no lift field.
`sprite-foot-corner-not-interpolated`: the deob anchors sprite feet at
`getElevation(tileCentre)` bilinear (World.java:391-410); the GO uses the raw SW
tile-corner height — diverging by half the per-tile rise on a slope (7-127 units;
constant 7u on a 15u/tile ramp, confirmed by `trace_foot.go`).
**Why gated:** both are deferred sprite-placement refinements. The lift needs a
new schema field (`GroundItemMarker.Lift`) populated by scanning co-located
scenery; the foot interpolation is a localized anchor-grid change. Both are
HIGH-confidence vs the deob but affect only sloped terrain / furniture (the
common flat case is unaffected), and the foot-Y change is best confirmed against
the JAR pixel oracle to avoid a half-tile regression on flat ground. Proposed for
a follow-up sprite pass.

### sprite-occlusion-zbuffer-vs-paintersort [unknown, low]
The deob unifies sprites and 3D faces in one painter sort with no z-buffer; the
Go rasterizes faces into a per-pixel avg-Z depth buffer then blits sprites with a
64-unit foot bias. Equivalent in every tested fixture, but a structural
architectural difference.
**Why gated:** no confirmed visual defect; needs the JAR pixel oracle to decide
whether the avg-Z+bias model ever mis-orders a tall sprite against a face whose
average-Z straddles the sprite's foot.

---

## 4. Clean bill — subsystems verified FAITHFUL (no divergence)

These were audited against the deob source (and corroborated against mudclient204
/ OpenRSC) and confirmed faithful via source line-up + differential fixtures:

- **Face normals / method293** (`model.go cameraNormal/cameraNormalSign`): cross
  product, >25000 halving, `normMag=4·√`, `dotV0` all match Scene.java:826-848.
- **relight / light normals** (`model.go relight/light`): incl. the deliberate
  `faceNormalZ*65535/mag`, ±8192 halving, per-vertex normal accumulation — matches
  GameModel.java:1069-1153.
- **SetLight remap + all four call-site values** (walls/terrain/roof/scenery).
- **Gouraud-vs-flat shade selection + TRANSPARENT (12345678) skip** —
  scene.go:226-240,276-286,141 ↔ Scene.java:1650,1725-1739.
- **Perspective projection + near-plane fallback + near-plane clip vertex-split +
  frustum X/Y cull** — model.go:425-454, scene.go:300-333 ↔ Scene.java:1749-1794.
- **Painter sort + overlap pass** (`facesort.go` polygonsQSort / reorderRange /
  separatePolygon / intersect) — verbatim port of mudclient204 Scene:1121-1530.
- **Textured intensity shift** (<<9 / <<6) ↔ Scene.java:1761-1765.
- **Lazy `faceCameraNormalSc` sharing** between facing + sort is sign-consistent
  (`front == (dotV0<0)` always holds).
- **Roof 7-way ridge cascade + apex/ridge raise + eave taper + leveling-to-max +
  roof tile selection + getWallRoof swapped-index** — `roof.go` ↔
  World.java:1165-1276, verified per-branch via the `rooftrace` harness.
- **Scenery footprint-centre + dir·32 roll orientation + getElevation bilinear +
  vertical anchor/axis + per-object direction** — `scenery.go`/`model.go` ↔
  World.addModels / Mudclient.java:6114-6150 (the World.java alias-comment
  mislabel was reconciled; the Go is correct).
- **Sprite billboard projection (axis swap) + scale-by-distance + screen centre +
  NPC/player/ground-item billboard sizes + 8-way facing/mirror/F-frame + body-part
  compositing + dye/skin recolour + multi-item stacking + self actor + fallback
  markers** — `render.go DrawEntitySprites` / `entitysprite.go` / `itemsprite.go`
  ↔ Scene.addSprite / Mudclient.drawGameFrame:7177-7251.
- **Terrain vertex placement, split winding / diagonal selection, height-twist
  split trigger, type-1 road split** — `terrain.go` ↔ World.java:894-981.
- **`groundColour` ramp** is byte-correct (the DEOB ramp is the scrambled one — do
  NOT change `palette.go`).
- **`removedBoundaries` / BoundaryRemoved door-open override** for the 4 edge dirs.
- **Wall edge-axis mapping, diagonal endpoints/winding, vertex layout, wall-height
  read, TRANSPARENT family** — `boundary.go` ↔ method422.

---

## 5. Next — what needs the DEOB/JAR engine (or a schema change) to confirm

1. **door-diag-object-not-built — DONE (FIX 8).** Ported `World.addModels` as
   `render/diagobj.go`; threaded a `TileDirection []byte` grid through
   `rscdump.Terrain` → `pathfind.Tile` → `landscape.go` so the diagonal door's
   `getTileDirection·32` orientation is correct. Remaining real-data follow-up:
   pull `getTileDirection` from a real map dump (the live `.orsc` sector record
   has no direction byte yet) and confirm a real diagonal-door MESH against the
   JAR pixel oracle (the fixture uses a synthetic flat panel).
2. **Multi-plane roof bundle (#10-13)** — extend the dump schema to carry planes
   0/1/2; then OR the under-roof bit across planes, cap `maxRoofPlane=2`, gate the
   cull on `plane==0`, and accumulate upper-story roof heights instead of the flat
   192×plane shim. Confirm against a real Lumbridge multi-story column via the JAR
   oracle.
3. **Sprite placement (#18,19)** — add `GroundItemMarker.Lift` + a centre-bilinear
   foot anchor; confirm the foot-Y interpolation against the JAR pixel oracle so
   the flat-ground case isn't regressed.
4. **bridge_seam_coord / shoreline / wall-light / sprite-occlusion (#15,16,17,20)**
   — cosmetic / design choices best settled by a GO↔JAR pixel diff on a real-map
   dump (sector-aligned window), not authorable today.
5. **roof-walltop-max-vs-first (#14)** — needs non-uniform wall heights in the
   dump path (the synthetic facts force uniform 192); apply the no-op-safe
   first-wins lock once a real-GameData wall dump can exercise it.

---

## NUMBERS

- **Total divergences found (deduped):** 20 Go bugs + 5 deob-vs-oracle
  discrepancies-where-Go-is-correct = 25 findings.
- **Applied (HIGH-confidence, fixture-verified):** 8
  - terrain-water-flatten-overlay-id (= bridge_flatten_overlay2) [PB]
  - terrain-type2-floor-split [PB]
  - wall-colour-table-ignored (+ wall-front-back-collapsed) [PB]
  - scenery-id74-no-vertical-lift [PB]
  - terrain-split-id-vs-class [C]
  - terrain-overlay-no-neighbour-spread [C]
  - roof-cull-hardcoded-set [L / maintainability]
  - door-diag-object-not-built (= scenery-diag-object-not-built) [PB] — port of
    World.addModels; diagonal door 0 faces → +1 wood leaf (centroid (10304,-96,
    10432), -15719, 2691 px visible); + `TileDirection` schema grid
- **Proposed-pending-oracle / schema / design (unapplied):** 12
  - wall-visible-gate-missing, roof-cull-singleplane,
    roof-cull-no-plane0-gate, roof-multistory-flat-storybase, roof-maxplane-off-by-one,
    roof-walltop-max-vs-first, bridge_seam_coord, bridge_shoreline_bleed,
    wall-light-params, grounditem-no-scenery-lift, sprite-foot-corner-not-interpolated,
    sprite-occlusion-zbuffer-vs-paintersort
- **Deob-vs-oracle discrepancies (Go already correct — do NOT change):** 5
- **Subsystems clean (verified faithful):** lighting/normals/shading, projection/
  clipping, face sort+overlap, roof ridge cascade, scenery footprint/orient/
  elevation, sprite projection/scale/facing/compositing, terrain vertex placement/
  split winding/twist/road-split, ground-colour ramp, removed-boundary override,
  wall edge-axis/winding/height/TRANSPARENT = **9 subsystem areas**.
- **go build ./... :** GREEN
- **go test ./cmd/renderdiff ./render ./... :** GREEN (18 packages ok, 0 fail;
  door self-test unchanged; 9 `TestFidelity_*` regression tests pass — 7 terrain/
  scenery/wall + 2 new diagonal-object)
- **go vet ./... :** clean for all changed files (one PRE-EXISTING warning in
  `proto/v235/buffer.go`, not touched here)
- **gofmt -l :** clean for all changed files (pre-existing unformatted files in
  `facts/`, `internal/rscdump/schema.go`, `internal/rscdump/gen/single_tile_door/`
  were not touched)
- **Files changed (render lib):**
  - `render/terrain.go` (FIX 1, 2, 5, 6)
  - `render/boundary.go` (FIX 3)
  - `render/dump.go` (FIX 3 synthetic-facts wood colour)
  - `render/scenery.go` (FIX 4)
  - `render/roof.go` (FIX 7)
  - `render/render.go` (FIX 8 — `BuildDiagonalObjects` wired into `buildScene`)
  - `internal/rscdump/schema.go` (FIX 8 — `Terrain.TileDirection` grid + validate)
  - `internal/rscdump/landscape.go` (FIX 8 — inject `TileDirection` per tile)
  - `pathfind/landscape.go` (FIX 8 — `Tile.TileDirection` field)
- **Files added (render lib + regression guards):**
  - `render/diagobj.go` (FIX 8 — `BuildDiagonalObjects`, the World.addModels port)
  - `render/fidelity_terrain_test.go` (5 tests)
  - `render/fidelity_scenery_test.go` (1 test, archive-gated skip)
  - `render/fidelity_walls_test.go` (1 test, DoorDef.xml-gated skip)
  - `render/fidelity_diagobj_test.go` (2 tests, FIX 8)
  - `testdata/rscdump/hunt/t3_{water_ov2,grass,indoor_split,class_corner,type4_neighbour,bridge_ov4}.json`
    + `testdata/rscdump/hunt/diagobj_door.json` (stable regression fixtures)
- **NOT modified:** committed `testdata/rscdump/{single_tile_*.json,out/}` goldens;
  no git commit performed. (`reference/rsc-client/src/client/Mudclient.java` in
  git status is the parallel deob-compile workflow's edit — read-only here.)
