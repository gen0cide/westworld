# Layer 1 — `render.Pick` screen-space hit-testing

**Status:** Design (M1)
**Owner doc:** this file. North star: [`00-overview.md`](00-overview.md) §3.
**Scope:** the single new render capability — map a screen-space `(px, py)` click in
the rendered `W×H` frame to a **depth-ordered list of world targets**, each
carrying enough identity to build an authentic RSC context menu (Layer 2) and to
dispatch a `runtime.Host` action (Layer 2 → §2 of the overview).

This is a **read-only projection**: `render.Pick` sends no packets, mutates no
world state, and renders nothing. It is the inverse-lookup companion to
`render.RenderView`, and it must agree with the rendered frame **pixel-for-pixel**
so the menu the user sees over a sprite is the menu for that sprite.

Everything below is grounded in the current source. File:line references are to
the tree at the time of writing; re-verify before editing.

---

## 0. Grounding — the facts this design depends on

| Fact | Source |
|---|---|
| `View` fields `X,Y,Plane,Rotation,Zoom,W,H,Entities[],GroundItems[],DynamicScenery[],NoSelf,SelfHeading,SelfOff*,SelfStepPhase,...` | `render/render.go:27-107` |
| `Entity{X,Y,Kind,NpcID,Heading,Index,OffX,OffZ,StepPhase,Moving,...}`; `.Index` IS the server NPC/player index | `render/entity.go:21-65` |
| `GroundItemMarker{X,Y,ItemID}` | `render/render.go:123-126` |
| `DynamicSceneryItem{X,Y,ID,Direction}` | `render/render.go:113-117` |
| `EntityKind` = `EntityNPC / EntityPlayer / EntitySelf` | `render/entity.go:3-10` |
| `SetCamera(localX, -elev, localZ, 912, rotation*4, 0, zoom*2)` | `render/render.go:373-378`, `render/camera.go:36-72` |
| `RenderTo` swaps `CameraYaw ↔ CameraRoll` **before** `project()` | `render/render.go:404-406`, mirrored in `render/pick.go:35` |
| `projectPoint(cam, X, Y, Z, cx, cy) -> (sx, sy, camZ, vis)` (already yaw/roll-swapped cam) | `render/pick.go:78-111` |
| Billboard math: `screenW=(worldW<<viewDist)/camZ`, `screenH=(worldH<<viewDist)/camZ`, top-left `(sx-screenW/2, feetY-screenH)`, foot `(lx*128+64+ox, -heights[lx][ly], ly*128+64+oz)` | `render/render.go:415-447`, `:538-554` |
| Painter order: sort by `camZ` **descending** (far first), so **near = last drawn = on top** | `render/render.go:536` |
| `spriteDepthBias = 64` (occlusion only; not an AABB term) | `render/render.go:553-562` |
| Constants: `terrainSize=160`, `viewDist=9`, `clipNear=5`, `clipFar=7000` | `render/terrain.go:7`, `render/scene.go:13-17` |
| Billboard world sizes: player `145×220`; NPC `npcBillboardSize(npcID)` (falls back to `145×220`); ground item `cs.W*2 × cs.H*2` (`groundItemPixelToWorld=2`) | `render/entitysprite.go:845-847,825-839`; `render/render.go:528-529,617` |
| `compositeNPC/compositePlayer/compositePlayerAppearance/compositeItem -> *CompositeSprite{W,H,Flip,...}` | `render/entitysprite.go:594-599,784,884,957`; `render/itemsprite.go:85` |
| `PickTile(land, view, px, py) -> (worldX, worldY, ok)` returns **plane-LOCAL** coords | `render/pick.go:15-72` |
| `facts.At(x,y) []Placement{Kind,DefID,Name,X,Y,Direction,Extra}` (Kind ∈ scenery/boundary/npc_spawn/ground_item) | `facts/facts.go:65-107` |
| `BoundaryLoc.Direction` (0=horizontal/blocks N-S, 1=vertical/blocks E-W) is exactly the arg `Host.InteractWithBoundary(x,y,direction)` wants | `facts/defs.go:102-108`, `runtime/boundary.go:19-27` |
| Dynamic scenery snapshot: `world.DynamicScenery.All() []SceneryRecord{X,Y,ID}` (absolute coords) | `world/scenery.go:18-22,88` |
| `buildLiveView` already converts every actor/item/scenery to plane-LOCAL Y before stuffing the `View` (`Y - planeOffset`) | `cmd/cradle/spectate.go:183-251` |

**Coordinate-space rule (critical).** Inside a `View`, every `Entity.X/Y`,
`GroundItemMarker.X/Y`, and `DynamicSceneryItem.X/Y` is in the **same
plane-LOCAL** absolute-X / plane-local-Y space the renderer windows against
(`baseX = v.X - terrainSize/2`, `baseY = v.Y - terrainSize/2`). `PickTile`
returns plane-local coords too. `render.Pick` therefore works **entirely in
plane-local space** and returns plane-local `X/Y` on every candidate — exactly as
the renderer and `PickTile` already do. The caller (Layer 3) adds
`plane*world.PlaneHeight` to recover absolute Y before dispatch, identically to
the existing `/walk` handler (`spectate.go:380`). This keeps `render.Pick`
dependency-symmetric with `PickTile`: render package, no `world` import.

---

## 1. The Go API

New file: **`render/pick.go`** already exists and holds `PickTile`+`projectPoint`;
add the picking entrypoint and types there (it is the natural home), and put the
**shared** projection/AABB helper in the new file **`render/hittest.go`** (§5).

### 1.1 `TargetKind`

```go
// TargetKind classifies a pick candidate by what the user is pointing at. It
// drives which identity fields are populated and which authentic menu verbs
// Layer 2 offers. The ordering is NOT a priority — depth ordering is by camera
// distance (billboards) then a fixed tile-target priority (§4).
type TargetKind int

const (
	TargetNPC        TargetKind = iota // server NPC billboard
	TargetPlayer                       // other-player billboard
	TargetSelf                         // the local player's own billboard
	TargetGroundItem                   // a dropped item icon
	TargetScenery                      // a scenery GameObject on a tile (static or dynamic)
	TargetBoundary                     // a wall/door/fence edge on a tile
	TargetTerrain                      // the bare ground tile (always-present "Walk here")
)
```

### 1.2 `PickCandidate`

One candidate carries **all** identity Layer 2 needs to (a) name + examine the
target, (b) build its option list from `facts` defs, and (c) call the exact
`Host` method. Fields are populated per `Kind`; unused fields are zero.

```go
// PickCandidate is one thing under the cursor. Every coordinate is PLANE-LOCAL
// (X absolute, Y plane-local) — the same space the View was built in and PickTile
// returns; Layer 3 adds plane*PlaneHeight to Y before dispatch. A candidate is
// self-describing: Kind selects which identity fields are meaningful.
type PickCandidate struct {
	Kind  TargetKind
	Plane int // == View.Plane; echoed so a candidate is dispatch-complete on its own

	// World tile of the target (plane-local). For billboards this is the
	// actor's/item's foot tile; for tile targets it is PickTile's result.
	X, Y int

	// --- billboard identity (NPC / Player / Self / GroundItem) ---
	Index  int // server actor index (NpcRecord/PlayerRecord .Index). NPC/Player/Self only.
	NpcID  int // config85 sprite id (== server TypeID). NPC only; used for name/def lookup.
	ItemID int // ground-item id (GroundItemMarker.ItemID). GroundItem only.

	// --- tile-grounded identity (Scenery / Boundary) ---
	DefID     int  // SceneryDef.ID (scenery) or BoundaryDef.ID (boundary).
	Direction int  // boundary edge dir (BoundaryLoc.Direction; arg to InteractWithBoundary). 0 otherwise.
	Dynamic   bool // scenery only: true if sourced from live world.DynamicScenery, false if static facts.

	// --- hit geometry (for debugging / tie-breaks / overlay) ---
	// ScreenRect is the on-screen AABB the candidate was tested against, in frame
	// pixels: [MinX, MinY, MaxX, MaxY]. For tile targets it is the projected tile
	// centre as a degenerate (zero-area) box. Useful for a debug overlay and for
	// reproducing the hit in tests; Layer 2 ignores it.
	ScreenRect [4]int
	// CamZ is the camera-space depth used for ordering (smaller = nearer). For
	// billboards it is the foot point's depth; for tile targets it is the tile
	// centre's depth. Exposed so callers can reason about ties; do not dispatch on it.
	CamZ int32
}
```

> **Why no `Name`/`Command`/`Examine` on the candidate.** Identity (Kind + ids +
> coords) is the contract Layer 1 owns. Resolving the *name*, the *examine text*,
> and the *verb list* is Layer 2's job (`facts` defs + `runtime/examine.go`),
> precisely because those depend on the `Host`/`Facts` instance, which the pure
> `render` package must not import. `render.Pick` stays a function of
> `(land, facts, view, px, py)`; it reads `facts` defs only to enumerate tile
> contents, never to format menus.

### 1.3 `Pick`

```go
// Pick maps a screen-space click (px,py in the rendered W×H frame described by v)
// to a depth-ordered list of world targets under the cursor. Ordering is:
//
//   1. Billboards whose AABB contains (px,py), NEAREST-CAMERA-FIRST (the reverse
//      of the renderer's far-first painter order — the sprite drawn last, hence
//      visually on top, is candidate #0).
//   2. The tile under the cursor (PickTile): its scenery objects, then its
//      boundary edges, then the bare terrain tile, in that fixed priority.
//
// So candidates[0] is the topmost thing the user is pointing at, and the list
// always ends with a TargetTerrain "Walk here" fallback (when PickTile resolves a
// tile), mirroring RSC's "Walk here" always being the last menu entry.
//
// Pure read: no packets, no mutation, no rendering. f may be nil (then only
// billboards + terrain are returned — no scenery/boundary enumeration). land may
// be nil (then no terrain/tile targets — billboards only).
func Pick(land *pathfind.Landscape, f *facts.Facts, v View, px, py int) []PickCandidate
```

`render` already imports `facts` (`render/render.go:8`) and `pathfind`
(`render/pick.go:3`), so no new dependency edge is introduced.

---

## 2. Billboard picking (NPC / Player / Self / GroundItem)

Billboards are projected exactly as `DrawEntitySprites` projects them, so a hit
box is pixel-identical to the blitted sprite's destination rectangle.

### 2.1 Camera reconstruction (identical to `PickTile`/`buildScene`)

`Pick` rebuilds the camera the frame was rendered with, **with defaults applied
exactly as `buildScene` applies them** so an under-specified `View` (zoom/W/H = 0)
projects to the same place it rendered:

```go
n := terrainSize
baseX := v.X - n/2
baseY := v.Y - n/2

if v.Zoom == 0 { v.Zoom = 750 }   // buildScene default (render.go:175-177)
if v.W == 0    { v.W = 512 }      // buildScene default (render.go:178-180)
if v.H == 0    { v.H = 334 }      // buildScene default (render.go:181-183)

localX := int32(v.X-baseX)*128 + 64 + int32(v.SelfOffX) // render.go:373
localZ := int32(v.Y-baseY)*128 + 64 + int32(v.SelfOffZ) // render.go:374
elev   := int32(land.Tile(v.X, v.Y, v.Plane).GroundElevation) * 3 // render.go:375-376
cam    := SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(v.Zoom)*2)
```

> **Default mismatch fix (do not copy `PickTile`'s 600/512/336 here).** `PickTile`
> defaults zoom to **600** and H to **336** (`pick.go:30,42`), but `buildScene`
> defaults zoom to **750** and H to **334** (`render.go:176,182`). For tile
> picking the discrepancy is harmless (nearest-tile is robust to a small zoom
> error), but billboard AABBs are tight, so `Pick` MUST use the `buildScene`
> defaults. The shared helper (§5) bakes these in so both `Pick` and the renderer
> agree.

### 2.2 The yaw/roll swap

`DrawEntitySprites` swaps the camera's yaw/roll once at entry
(`render.go:404-406`) and `projectPoint` assumes an already-swapped camera
(`pick.go:74-77`). The shared helper performs the swap exactly once. **Do not
double-swap**: if `Pick` calls the shared helper which swaps internally, it must
pass the **un-swapped** `cam` from `SetCamera`. (Concretely: the helper owns the
swap; callers hand it `SetCamera`'s output verbatim.)

### 2.3 Foot projection + AABB (the exact `DrawEntitySprites` math)

For each billboard, project its foot point and derive the destination rectangle.
These are the **verbatim** expressions from `DrawEntitySprites`:

Foot point (`render.go:419-421`):
```go
x := (int32(lx)*128 + 64 + ox) - cam.CameraX
y := -heights[lx][ly] - cam.CameraY
z := (int32(ly)*128 + 64 + oz) - cam.CameraZ
```
(where `lx=e.X-baseX`, `ly=e.Y-baseY`, `ox=e.OffX`, `oz=e.OffZ`). After the
yaw→roll→pitch rotations (`render.go:422-442`) and the `z < clipNear` reject
(`render.go:443-444`), the screen foot is (`render.go:446`):
```go
sx    = int32(cx) + (x<<uint(viewDist))/z       // cx = v.W/2
feetY = int32(cy) + (y<<uint(viewDist))/z       // cy = v.H/2
camZ  = z
```
The screen size + rectangle (`render.go:539-554`):
```go
screenW := (int32(worldW) << uint(viewDist)) / camZ
screenH := (int32(worldH) << uint(viewDist)) / camZ
// reject screenW<=0 || screenH<=0   (render.go:541-543)
left   := int(sx) - int(screenW)/2
top    := int(feetY) - int(screenH)
right  := left + int(screenW)
bottom := int(feetY)          // foot sits at the rect bottom
```

> The depth-buffer occlusion bias (`spriteDepthBias=64`, `render.go:553`) affects
> only which sprite *pixels survive* the z-test during blitting; it does **not**
> change the destination rectangle. The picker therefore ignores it for AABB
> containment. (It uses the **un-biased** `camZ` for ordering, matching the
> renderer's sort key at `render.go:536`.)

A click hits when `left ≤ px ≤ right && top ≤ py ≤ bottom`. Because the AABB is
the full sprite canvas (transparent margins included — `BlitSpriteScaled`
honours `cs.Opaque` per pixel but the destination rect is the whole canvas), the
hit box is slightly generous around a thin sprite. That is correct and authentic:
RSC's `bounds`-array hit-testing also uses the sprite's full plotted rectangle,
not its opaque silhouette.

### 2.4 What to enumerate, in the renderer's exact order

To match the rendered frame the picker must build the **same candidate set with
the same sprite source** `DrawEntitySprites` used, so a hit corresponds to a
sprite that actually drew:

1. `v.Entities` in order (`render.go:487-499`):
   - `EntityNPC`: `worldW,worldH = npcBillboardSize(e.NpcID)`, sprite =
     `compositeNPC(e.NpcID, facing, step)`; → `Kind=TargetNPC`, `NpcID=e.NpcID`,
     `Index=e.Index`.
   - else (`EntityPlayer`): `worldW,worldH = playerBillboardW,playerBillboardH`,
     sprite = `playerSprite(...)`; → `Kind=TargetPlayer`, `Index=e.Index`.
   - `facing := (e.Heading + camTerm) & 7`, `camTerm := (v.Rotation+16)/32`,
     `step := e.StepPhase` (`render.go:471,488,490`).
2. Self (unless `v.NoSelf`) (`render.go:505-511`): foot tile `(v.X,v.Y)` with
   offset `(v.SelfOffX,v.SelfOffZ)`, size `145×220`; → `Kind=TargetSelf`,
   `Index=0` (self is server index 0 — see `spectate.go:200-201`).
3. `v.GroundItems` (`render.go:522-531`): `cs := compositeItem(gi.ItemID)`,
   `worldW=cs.W*groundItemPixelToWorld`, `worldH=cs.H*groundItemPixelToWorld`,
   offset `(0,0)`; → `Kind=TargetGroundItem`, `ItemID=gi.ItemID`.

> **Sprite-presence gate.** A billboard that fails to composite
> (`cs == nil`) is NOT blitted (`render.go:456-458,526-527`); the renderer routes
> it to a 3D-cross / red-marker fallback instead. For M1 the picker builds an
> AABB **only when the composite succeeds** (it calls the same `composite*`
> functions and skips `nil`), so a hit always corresponds to a real on-screen
> sprite. NPCs/items that fell back to the 3D cross are still reachable via their
> **tile** (they sit on a tile, picked in §4) — they just have no tight billboard
> box in M1. (Backlog: project the 3D-cross/marker bounds too; tracked in
> `90-backlog.md`.)

### 2.5 Ordering: nearest-camera-first

Collect every hit into a slice, then sort by `CamZ` **ascending** (nearest first)
— the **reverse** of the renderer's far-first painter sort (`render.go:536`).
Use a **stable** sort so equal-depth ties preserve enumeration order (NPCs before
players before self before ground items), matching the renderer's
`sort.SliceStable`. Billboard candidates precede all tile candidates in the
returned list.

```go
sort.SliceStable(hits, func(i, j int) bool { return hits[i].CamZ < hits[j].CamZ })
```

---

## 3. Ground-item picking

Ground items are billboards too and are handled by the **same** §2 path, with two
specifics already captured above:

- World size is the **composited icon's** size scaled by `groundItemPixelToWorld`
  (`= 2`): `worldW = cs.W*2`, `worldH = cs.H*2` (`render.go:528-529,617`).
- No glide offset (`ox=oz=0`, `render.go:530`).
- Foot anchor is the canvas bottom, so the item floats just off the terrain — the
  AABB bottom is the projected tile-foot `feetY`, top is `feetY - screenH`,
  identical to actors.

`Kind=TargetGroundItem`, `ItemID=gi.ItemID`, `X,Y = gi.X,gi.Y` (plane-local).
Items whose icon fails to decode (`compositeItem==nil`) are skipped for the same
reason as §2.5's gate; they remain reachable via their tile.

---

## 4. Tile-grounded picking (Scenery / Boundary / Terrain)

Tile targets are resolved through the existing `PickTile`, then enumerated from
`facts` + dynamic world state. They are appended **after** all billboard
candidates.

### 4.1 Resolve the tile

```go
tx, ty, ok := PickTile(land, v, px, py)   // plane-LOCAL coords (pick.go:15-72)
```
If `!ok` (or `land == nil`), there are no tile candidates and the list is just the
billboards. `PickTile`'s looser zoom default is fine here (nearest-tile is robust)
but to stay self-consistent `Pick` calls it with the same defaulted `v` from §2.1.

> **Absolute vs plane-local for `facts`.** `facts.At` is keyed in the **absolute**
> space `facts.SceneryLocs`/`BoundaryLocs` use (`facts/facts.go:76-107`,
> `facts/load.go:333,349`). `PickTile`'s `(tx,ty)` is plane-local. For the
> ground floor (`Plane==0`) they coincide. For upper planes the picker queries
> `facts.At(tx, ty + v.Plane*planeHeightTiles)` (the `absWorldY` conversion at
> `render.go:593-595`), then stores the **plane-local** `tx,ty` back on the
> candidate so dispatch matches `/walk`. The scenery-plane guard
> (`sceneryPlane(loc.Y)==v.Plane`, `render.go:293`) is implicitly satisfied
> because `facts.At` only returns locs on the queried absolute Y. Dynamic scenery
> (`world.DynamicScenery`) is keyed absolute too; the caller passes it pre-
> converted to plane-local on the `View`, and the picker matches on the View's
> `DynamicScenery` slice (already plane-local) rather than re-querying `world`.

### 4.2 Enumerate, in fixed priority

For tile `(tx, ty)` (plane-local) push candidates in this order — scenery first
(most-specific / topmost interactable), then boundaries, then terrain last:

1. **Dynamic scenery on this tile** (live, wins over static — same precedence as
   the renderer's dedup at `render.go:280-307`): scan `v.DynamicScenery` for an
   entry with `X==tx && Y==ty`. → `Kind=TargetScenery`, `DefID=ds.ID`,
   `Direction=ds.Direction`, `Dynamic=true`.
2. **Static scenery on this tile** (only if not superseded by a dynamic entry on
   the same tile, and not suppressed by `v.SceneryRemoved`): for each
   `p` in `facts.At(tx, absY)` with `p.Kind=="scenery"` →
   `Kind=TargetScenery`, `DefID=p.DefID`, `Direction=p.Direction`,
   `Dynamic=false`.
3. **Boundaries on this tile**: for each `p` in `facts.At(tx, absY)` with
   `p.Kind=="boundary"` (skip those `v.BoundaryRemoved(tx, absY, p.Direction)`
   reports removed) → `Kind=TargetBoundary`, `DefID=p.DefID`,
   `Direction=p.Direction`. `Direction` is exactly the third arg to
   `Host.InteractWithBoundary(x,y,direction)` (`runtime/boundary.go:19`) and
   `UseItemOnBoundary` (`:34`).
4. **Terrain** (always, when `ok`): `Kind=TargetTerrain`, no def/id, `X,Y=tx,ty`.
   This is the always-present, lowest-priority "Walk here" target.

> `npc_spawn` placements from `facts.At` are **ignored** here — live NPCs are
> billboards (§2), and a spawn point is not a clickable target. `ground_item`
> placements from `facts.At` are also ignored; live ground items are billboards
> (§3).

`X,Y` on every tile candidate = `(tx, ty)` plane-local. `CamZ` = the projected
tile-centre depth from `projectPoint` (reuse `PickTile`'s per-tile projection or
re-project the picked tile centre) so a debug overlay can order tile vs billboard
sensibly; it is not used for the primary sort (tile candidates always follow
billboards). `ScreenRect` for tile candidates = the projected tile centre as a
zero-area box `[sx,sy,sx,sy]`.

### 4.3 Resulting list shape

```
[ near billboard, … , far billboard,   dynamic scenery?, static scenery?, boundary(ies), terrain ]
       └──────── §2/§3, CamZ asc ────────┘  └──────────────── §4, fixed priority ─────────────────┘
```

`candidates[0]` = topmost interactable; the list ends in `TargetTerrain` whenever
a tile resolved. Layer 2 maps `candidates[0]` to the left-click default action
and the whole list to the right-click menu.

---

## 5. Refactor: factor projection + AABB into a shared helper

Today the foot projection and AABB live **inline inside `DrawEntitySprites`**
(`render.go:415-447` for `project`, `:539-554` for the rectangle). The picker
must use *byte-identical* math, and the only safe way to guarantee that is to
**share one implementation**. Duplicating the expressions risks silent drift the
moment someone tweaks the renderer.

### 5.1 New file `render/hittest.go`

Extract two pure helpers:

```go
// billboardCamera reconstructs the (yaw/roll-swapped) camera for a View, applying
// the SAME zoom/W/H defaults buildScene applies, and returns it plus baseX/baseY,
// cx/cy and the terrain heights grid. Used by BOTH DrawEntitySprites and Pick so
// the projection origin is identical. (Swaps yaw/roll exactly once — callers do
// not swap.)
func billboardCamera(land *pathfind.Landscape, v *View) (cam Camera, baseX, baseY, cx, cy int, heights [][]int32)

// projectBillboard projects one actor/item foot point + derives its on-screen
// AABB, returning the rectangle, foot pixel (sx,feetY), and camera depth. This is
// the LITERAL math currently inlined in DrawEntitySprites (render.go:415-447 +
// 539-554), moved verbatim. ok=false on clipNear reject or non-positive screen
// size — the same two rejects the renderer applies, so a box exists iff the
// renderer would have blitted.
func projectBillboard(cam Camera, cx, cy int, heights [][]int32, lx, ly int, ox, oz int32, worldW, worldH int) (rect [4]int, sx, feetY, camZ int32, ok bool)
```

### 5.2 Make `DrawEntitySprites` call the helper — rendering stays byte-identical

`DrawEntitySprites` keeps its current structure (collect → painter-sort →
blit) but its inline `project` closure and its per-item `screenW/screenH`
computation are **replaced by `projectBillboard`**. Crucially:

- The blit still uses `it.camZ - spriteDepthBias` for the depth test
  (`render.go:553`) — the bias is a *blit* concern and stays in `DrawEntitySprites`,
  not in the shared helper.
- The blit top-left stays `(rect[0], rect[1])` and size `(rect[2]-rect[0],
  rect[3]-rect[1])`, which equal the previous `(sx-screenW/2, feetY-screenH)` and
  `(screenW, screenH)` by construction.
- The painter sort key is unchanged (un-biased `camZ`, descending).

Because the helper is the same integer math moved verbatim, the rendered
framebuffer is bit-for-bit identical — guarded by the existing render-fidelity
tests (`render/fidelity_*_test.go`) which must still pass unchanged.

### 5.3 `Pick` reuses the same helper

`Pick` calls `billboardCamera` + `projectBillboard` for each candidate (skipping
the `cs==nil` composites per §2.4), tests `px,py ∈ rect`, and orders
nearest-first. There is exactly **one** copy of the projection/AABB math in the
tree after this refactor.

---

## 6. Deterministic unit-test plan

New file: **`render/hittest_test.go`** (sibling to the other `render/*_test.go`).
All tests are pure integer math — no archives, no `Landscape` I/O (use a flat
fake heights grid), no PNGs — so they are fast and deterministic.

### 6.1 Projection round-trip / self-consistency

- **`TestProjectBillboardCentre`**: with a fixed `View{Rotation:0,Zoom:750,W:512,
  H:334}` and a flat zero-height grid, an actor placed on the **host's own tile**
  projects its foot `sx` to the frame centre `±` a small tolerance (the camera
  looks down the host tile). Asserts the helper agrees with the documented
  `cx=v.W/2`.
- **`TestProjectBillboardMatchesProjectPoint`**: for several `(lx,ly)` and several
  rotations `{0,64,128,192}`, `projectBillboard`'s `(sx,feetY,camZ)` for `worldW=
  worldH=0` equals `projectPoint(swappedCam, foot…)` from `pick.go` (cross-check
  the two projection paths produce the same foot pixel). This pins the yaw/roll
  swap and the screen-centre offset.

### 6.2 AABB correctness

- **`TestBillboardAABB`**: for a known `camZ`, assert
  `rect == [sx - (worldW<<9)/camZ/2, feetY - (worldH<<9)/camZ, …]` against the
  literal formula with `viewDist=9`. Table of `(worldW,worldH,camZ)` incl. the
  `145×220` player size.
- **`TestBillboardAABBContainment`**: project a player at a known tile, then assert
  `Pick` returns it for a click at the **rect centre**, at all **four corners**
  (inclusive bounds), and returns it **not** for a click one pixel outside each
  edge.
- **`TestClipNearReject`**: an actor behind the camera (placed so `z<clipNear`)
  produces `ok=false` and yields no candidate.
- **`TestZeroSizeReject`**: a tiny `worldW/worldH` at huge `camZ` driving
  `screenW<=0` yields `ok=false` (matches `render.go:541-543`).

### 6.3 Ordering

- **`TestPickOrderingNearestFirst`**: two overlapping billboards at the same
  screen pixel but different `camZ`; assert the **nearer** (smaller `camZ`) is
  `candidates[0]`. Confirms the reverse-of-painter order vs the renderer's
  far-first sort.
- **`TestPickStableTieOrder`**: two billboards at **identical** `camZ` overlapping
  the click preserve enumeration order (NPC before player before self before
  ground item).
- **`TestPickTerrainAlwaysLast`**: with `f != nil` and a scenery+boundary on the
  picked tile, assert the candidate list ends with exactly one `TargetTerrain`,
  preceded by `TargetScenery` then `TargetBoundary` (the §4.2 priority).

### 6.4 Identity completeness

- **`TestCandidateIdentityNPC`**: an `Entity{Kind:EntityNPC,Index:7,NpcID:3,X,Y}`
  produces a `TargetNPC` candidate with `Index==7`, `NpcID==3`, `X,Y` matching and
  `Plane==v.Plane`.
- **`TestCandidateIdentityGroundItem`**: a `GroundItemMarker{ItemID:10,X,Y}`
  produces `TargetGroundItem` with `ItemID==10`.
- **`TestCandidateIdentityBoundary`** (needs a tiny hand-built `*facts.Facts` with
  one `BoundaryLoc{DefID,X,Y,Direction:1}` — mirror the `syntheticFacts` fixture
  pattern used by `render/fidelity_*_test.go`): a click on that tile yields a
  `TargetBoundary` with `DefID` and `Direction==1` (the `InteractWithBoundary`
  arg).
- **`TestCandidateIdentityScenery`**: static scenery loc + a dynamic
  `DynamicSceneryItem` on the **same** tile → the dynamic one appears (with
  `Dynamic==true`) and the static one is suppressed (dedup parity with the
  renderer, §4.2).

### 6.5 Renderer-parity guard

- **`TestDrawEntitySpritesAABBUnchanged`**: a golden test asserting
  `projectBillboard` reproduces the pre-refactor inline numbers for a fixed
  `(View, tile, size)` table (hard-coded expected rects captured once from the
  current `render.go` math). This is the regression tripwire that catches any
  future divergence between the renderer's blit rect and the picker's hit box.
- The existing `render/fidelity_*_test.go` suite is the second guard: after the
  §5 extraction the rendered frames must be **bit-identical**, so those tests
  passing unchanged proves rendering stayed byte-for-byte stable.

### Gate

`go build ./...`, `go vet ./...`, `go test ./render/...` (per overview §5 M1
gate). The new tests live under `./render/...`.

---

## 7. Implementation checklist (no re-derivation needed)

1. `render/hittest.go`: add `billboardCamera` + `projectBillboard` (move the math
   verbatim from `render.go:415-447,539-554`; bake in the buildScene defaults
   750/512/334; swap yaw/roll exactly once).
2. `render/render.go`: rewrite `DrawEntitySprites`'s inline `project` closure +
   `screenW/screenH` block to call `projectBillboard`; keep the
   `camZ - spriteDepthBias` blit depth and the far-first painter sort. Verify
   fidelity tests still pass (byte-identical).
3. `render/pick.go`: add `TargetKind`, `PickCandidate`, `Pick`. Billboards via the
   shared helper (§2/§3); tile targets via `PickTile` + `facts.At` + the View's
   dynamic-scenery/boundary-removed state (§4). Sort billboards nearest-first;
   append tile targets in fixed priority ending in `TargetTerrain`.
4. `render/hittest_test.go`: the §6 suite.
5. Hand off to Layer 2 (`20-menu-dispatch.md`): `PickCandidate` → option list +
   `Host` call. Layer 3 (`30-http-api.md`) owns the plane-local→absolute Y add
   before dispatch (mirror `spectate.go:380`).
