package render

import (
	"os"
	"sort"

	"github.com/gen0cide/westworld/assets"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// Bundle holds the decoded asset archives + caches the renderer consumes.
type Bundle struct {
	Models *ModelCache
}

// OpenBundle opens the OpenRSC models.orsc archive.
func OpenBundle(modelsPath string) (*Bundle, error) {
	arc, err := assets.OpenArchive(modelsPath)
	if err != nil {
		return nil, err
	}
	return &Bundle{Models: NewModelCache(arc)}, nil
}

// View are the render parameters.
type View struct {
	X, Y     int // host world TILE coords
	Plane    int
	Rotation int // 0..255 RSC camera angle (yaw = rotation*4)
	Zoom     int // default ~600 -> distance = zoom*2
	W, H     int // output image size

	// Entities are the actors (players/NPCs) the host perceives, in ABSOLUTE
	// world-tile coords. Drawn as depth-scaled billboards. Optional: leave nil
	// to render only the static world (terrain + scenery + boundaries).
	Entities []Entity

	// DrawSelf, when not explicitly disabled, draws the local player ("bernard")
	// as a standing-frame human billboard at the centre of his own view (his own
	// tile). RSC always shows the host in the third-person scene. Defaults to ON;
	// set NoSelf to suppress it.
	NoSelf bool

	// SelfHeading is the host's own 8-way facing (0=N..7=NW; the opcode-191 own
	// sprite). It is added to the camera term to choose which side of the local
	// player's sprite to draw. Zero (north-facing) when unknown — the facing then
	// follows the camera alone (Phase 3a behaviour).
	SelfHeading int

	// Self appearance: the host's own worn-equipment sprite array + colour
	// indices, mirrored from world.Self. SelfHasEquip selects the real-
	// appearance composite for the local player; when false bernard renders
	// the default human (compositePlayer), preserving today's behaviour for a
	// host whose own appearance update hasn't landed.
	SelfEquipSprites  [12]int
	SelfHairColour    int
	SelfTopColour     int
	SelfTrouserColour int
	SelfSkinColour    int
	SelfHasEquip      bool

	// --- live dynamic server state (all OPTIONAL / nil-guarded; a host with
	// no dynamic state renders exactly as the static world) ---

	// BoundaryRemoved reports whether the wall/door edge at absolute (x, y, dir)
	// has been removed by a live server update (door opened, web cut). dir uses
	// the authentic createModel convention (mudclient.java:6769-6780):
	//   0 = edge (x,y)..(x+1,y)   (east-west; the .orsc VerticalWall byte)
	//   1 = edge (x,y)..(x,y+1)   (north-south; the .orsc HorizontalWall byte)
	//   2 = edge (x,y)..(x+1,y+1) ('\' diagonal)
	//   3 = edge (x+1,y)..(x,y+1) ('/' diagonal)
	// When set, BuildBoundaries skips a static wall quad the override marks
	// removed, so an opened door / cut web renders passable instead of solid.
	// nil => no dynamic boundary state (static walls render unchanged).
	BoundaryRemoved func(x, y, dir int) bool

	// DynamicScenery are live GameObjects the host perceives that are NOT in
	// the static landscape (lit fires, regrown/changed trees, etc.). Placed
	// after the static scenery loop.
	DynamicScenery []DynamicSceneryItem

	// SceneryRemoved reports whether the static scenery at absolute (x, y) was
	// actively cleared by a server removal (mined rock depleted, fire burned
	// out). When set, the static SceneryLocs loop suppresses that tile. nil =>
	// no suppression (static scenery renders unchanged).
	SceneryRemoved func(x, y int) bool

	// GroundItems are dropped items the host perceives, drawn as small ground
	// markers within the window.
	GroundItems []GroundItemMarker
}

// DynamicSceneryItem is one live GameObject the host perceives, in absolute
// world-tile coords. ID joins to facts.SceneryDef.ID (the GameObject id space).
// Direction is the object's heading (0 when unknown — the scenery handler in
// this protocol carries no per-object direction).
type DynamicSceneryItem struct {
	X, Y      int
	ID        int
	Direction int
}

// GroundItemMarker is one dropped item the host perceives, in absolute
// world-tile coords. ItemID drives the rendered icon: the depth-sorted sprite
// pass draws the item's real inventory icon (compositeItem) anchored at its
// tile; only items whose icon can't be decoded fall back to a flat red marker.
type GroundItemMarker struct {
	X, Y   int
	ItemID int
}

// RenderView assembles the terrain + nearby scenery around the host tile and
// renders a PNG of what that host sees. This is the Phase-1 MVP entrypoint.
func RenderView(land *pathfind.Landscape, f *facts.Facts, b *Bundle, v View) ([]byte, error) {
	if v.Zoom == 0 {
		v.Zoom = 750
	}
	if v.W == 0 {
		v.W = 512
	}
	if v.H == 0 {
		v.H = 334
	}

	// CENTER the 96x96 terrain window ON the host tile. The earlier code
	// mirrored pathfind.BuildGrid's midRegion centring (baseX = mid*48-48),
	// which only guarantees the host lands somewhere in 24..72 of the window —
	// frequently near a window edge, so the camera looked across a cliff
	// edge-on (the dark-wedge artifact). Since pathfind.Landscape.Tile lazily
	// loads ANY sector, we are not bound to sector-aligned windows: put the
	// host at local (48,48), the dead centre, every time.
	baseX := v.X - terrainSize/2
	baseY := v.Y - terrainSize/2

	sc := &Scene{}
	// Build the flattened terrain-height grid once; both the terrain mesh and
	// the boundary feet anchor to it.
	heights := TerrainHeights(land, baseX, baseY, v.Plane)
	if os.Getenv("RENDER_NO_TERRAIN") == "" {
		terrain := BuildTerrain(land, baseX, baseY, v.Plane)
		sc.Add(terrain)
	}

	// boundaries (walls/fences/doors) within the window
	if os.Getenv("RENDER_NO_BOUNDARIES") == "" && f != nil {
		// BuildBoundaries works in window space (baseY is plane-local Y); wrap
		// the absolute-coord View.BoundaryRemoved so the edge loop can query it
		// with its window-local (x, y). nil stays nil (no override).
		var boundaryRemoved func(x, y, dir int) bool
		if v.BoundaryRemoved != nil {
			plane := v.Plane
			boundaryRemoved = func(x, y, dir int) bool {
				return v.BoundaryRemoved(x, absWorldY(y, plane), dir)
			}
		}
		if bd := BuildBoundaries(f, land, baseX, baseY, v.Plane, heights, boundaryRemoved); bd != nil {
			sc.Add(bd)
		}
	}

	// roofs (clay/timber tile patches) over building footprints. Built as an
	// isolated additive pass and depth-sorted into the scene via sc.Add so
	// method277 orders roof faces against walls/terrain. RENDER_NO_ROOFS removes
	// the pass entirely if the topology misbehaves.
	if os.Getenv("RENDER_NO_ROOFS") == "" && !HostUnderRoof(land, v.X, v.Y, v.Plane) {
		if rf := BuildRoofs(f, land, baseX, baseY, v.Plane, heights); rf != nil {
			sc.Add(rf)
		}
	}

	// entity billboards (players/NPCs) the host perceives. RSC characters are 2D
	// sprites: the depth-scaled sprite blit happens AFTER RenderTo (see
	// DrawEntitySprites below). The 3D-cross billboard is kept ONLY as a fallback
	// for actors whose sprite composite fails (archives missing / unknown npc id)
	// so they still appear as *something* — successful sprites are not crossed.
	if os.Getenv("RENDER_NO_ENTITIES") == "" && len(v.Entities) > 0 {
		camTerm := (v.Rotation + 16) / 32
		var fallback []Entity
		for _, e := range v.Entities {
			facing := (e.Heading + camTerm) & 7
			var cs *CompositeSprite
			switch {
			case e.Kind == EntityNPC:
				cs = compositeNPC(e.NpcID, facing)
			case e.HasEquip:
				cs = compositePlayerAppearance(e.EquipSprites, e.HairColour, e.TopColour, e.TrouserColour, e.SkinColour, facing)
				if cs == nil {
					cs = compositePlayer(facing) // empty/undecodable outfit -> default human
				}
			default:
				cs = compositePlayer(facing)
			}
			if cs == nil {
				fallback = append(fallback, e)
			}
		}
		if len(fallback) > 0 {
			if ents := BuildEntities(fallback, baseX, baseY, v.Plane, heights); ents != nil {
				sc.Add(ents)
			}
		}
	}

	// place scenery within the window
	if os.Getenv("RENDER_NO_SCENERY") == "" && f != nil && b != nil {
		// Dedup vs dynamic scenery: the live server (opcode-48 region objects)
		// re-sends every in-view object INCLUDING the static baseline, so
		// v.DynamicScenery overlaps f.SceneryLocs — placing both double-renders
		// wells/signs/etc. Build the set of dynamic-covered tiles; a static loc
		// on a covered tile is skipped below (dynamic wins — it carries live
		// id/direction drift). ds.Y is plane-local (the cradle subtracts the
		// plane offset); the static loc.Y (absolute) is converted to match.
		dynTiles := make(map[[2]int]bool, len(v.DynamicScenery))
		for _, ds := range v.DynamicScenery {
			dynTiles[[2]int{ds.X, ds.Y}] = true
		}
		for _, loc := range f.SceneryLocs {
			if loc.X < baseX || loc.X >= baseX+terrainSize ||
				loc.Y < baseY || loc.Y >= baseY+terrainSize {
				continue
			}
			// Plane guard: SceneryLocs carry a Y offset by PlaneHeight per
			// floor; only the loc on THIS plane belongs in the window. Without
			// it, upper-floor / dungeon scenery leaks into a ground-floor view
			// (and vice-versa) because the window only filters X/Y, not floor.
			if sceneryPlane(loc.Y) != v.Plane {
				continue
			}
			// Suppress scenery the live server state actively removed (mined
			// rock / burned-out fire) so the static baseline object doesn't
			// pop back. nil override => no suppression (renders as today).
			if v.SceneryRemoved != nil && v.SceneryRemoved(loc.X, loc.Y) {
				continue
			}
			// A live dynamic placement supersedes the static baseline at this
			// tile (server re-sent the same object) — skip the static copy to
			// avoid the doubled well/signpost. loc.Y absolute -> plane-local key.
			if dynTiles[[2]int{loc.X, loc.Y - v.Plane*planeHeightTiles}] {
				continue
			}
			if g := PlaceScenery(b.Models, f, land, baseX, baseY, v.Plane, loc); g != nil {
				sc.Add(g)
			}
		}
		// Place live dynamic scenery (lit fires, regrown trees) the host
		// perceives — these are NOT in the static landscape. Reuse PlaceScenery
		// via a synthesized SceneryLoc; plane- and window-guarded like statics.
		for _, ds := range v.DynamicScenery {
			if ds.X < baseX || ds.X >= baseX+terrainSize ||
				ds.Y < baseY || ds.Y >= baseY+terrainSize {
				continue
			}
			if sceneryPlane(ds.Y) != v.Plane {
				continue
			}
			loc := facts.SceneryLoc{DefID: ds.ID, X: ds.X, Y: ds.Y, Direction: ds.Direction}
			if g := PlaceScenery(b.Models, f, land, baseX, baseY, v.Plane, loc); g != nil {
				sc.Add(g)
			}
		}
	}

	// ground items within the window (live dropped items). Each is drawn as its
	// real 2D inventory icon in the depth-sorted sprite pass below (so it sorts
	// with characters + is occluded by walls), exactly like an entity billboard.
	// Only items whose icon fails to decode (archives missing / unknown id) fall
	// back to the flat red marker quad, added to the 3D scene here so it still
	// shows as *something*. The successfully-iconned items are NOT marked.
	if os.Getenv("RENDER_NO_GROUND_ITEMS") == "" && len(v.GroundItems) > 0 {
		var fallback []GroundItemMarker
		for _, gi := range v.GroundItems {
			if compositeItem(gi.ItemID) == nil {
				fallback = append(fallback, gi)
			}
		}
		if len(fallback) > 0 {
			if gi := BuildGroundItems(fallback, baseX, baseY, v.Plane, heights); gi != nil {
				sc.Add(gi)
			}
		}
	}

	// camera at the host tile, looking with the requested yaw.
	localX := int32(v.X-baseX)*128 + 64
	localZ := int32(v.Y-baseY)*128 + 64
	t := land.Tile(v.X, v.Y, v.Plane)
	elev := int32(t.GroundElevation) * 3
	// setCamera(x, z=height, y, pitch=912, yaw=rotation*4, roll=0, distance)
	sc.Cam = SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(v.Zoom)*2)

	surf := NewSurface(v.W, v.H)
	surf.Clear(skyColour)
	sc.RenderTo(surf, v.W/2, v.H/2)

	// 2D character-sprite pass: composite + blit nearby NPCs/players and the
	// local player as depth-scaled standing billboards over the rendered world.
	// RSC characters are sprites, not 3D models; this replaces the 3D-cross
	// billboards for any actor whose sprite composites successfully (the cross is
	// kept only as a fallback inside BuildEntities when the archives are missing).
	if os.Getenv("RENDER_NO_ENTITIES") == "" || os.Getenv("RENDER_NO_GROUND_ITEMS") == "" {
		DrawEntitySprites(surf, sc.Cam, v, v.Entities, baseX, baseY, heights)
	}

	return surf.PNG()
}

// DrawEntitySprites composites each entity (and the local player) into a 2D
// standing-frame billboard and blits it depth-scaled over the rendered world.
//
// Projection mirrors GameModel.project() EXACTLY, including the axis swap
// Scene.RenderTo performs before projecting (CameraYaw <-> CameraRoll), so a
// character lands at the same screen position the 3D pipeline would place a
// model at that tile. For each actor:
//   - foot world point = (lx*128+64, -heights[lx][ly], ly*128+64)
//   - transform into camera space (CameraX/Y/Z subtract, then yaw/roll/pitch)
//   - skip if camZ < clipNear
//   - viewX = (x<<viewDist)/z, viewY = (y<<viewDist)/z
//   - screen feet at (W/2 + viewX, H/2 + viewY)
//   - screen size = (worldW<<viewDist)/z, (worldH<<viewDist)/z
//   - blit scaled, top-left (sx-screenW/2, feetY-screenH)
//
// All actors (NPCs, other players, AND the local player) are collected, then
// painter-sorted by camera depth (far first) so a CLOSER actor — e.g. a mob
// standing between the camera and the player — paints OVER a farther one. The
// old code drew the local player unconditionally LAST, so bernard incorrectly
// occluded everything in front of him.
func DrawEntitySprites(surf *Surface, cam Camera, v View, ents []Entity, baseX, baseY int, heights [][]int32) {
	// Axis swap at the call site (Scene.RenderTo): project is called with
	// yaw/roll bound swapped. Replicate so screen placement matches the 3D pass.
	cam.CameraYaw, cam.CameraRoll = cam.CameraRoll, cam.CameraYaw

	n := terrainSize
	cx := v.W / 2
	cy := v.H / 2

	// project a foot world point into screen space + return its camera depth.
	project := func(lx, ly int) (sx, feetY, camZ int32, ok bool) {
		if lx < 0 || lx >= n || ly < 0 || ly >= n {
			return 0, 0, 0, false
		}
		x := (int32(lx)*128 + 64) - cam.CameraX
		y := -heights[lx][ly] - cam.CameraY
		z := (int32(ly)*128 + 64) - cam.CameraZ
		if cam.CameraYaw != 0 {
			ys := sine11[cam.CameraYaw]
			yc := sine11[cam.CameraYaw+1024]
			X := (y*ys + x*yc) >> 15
			y = (y*yc - x*ys) >> 15
			x = X
		}
		if cam.CameraRoll != 0 {
			rs := sine11[cam.CameraRoll]
			rc := sine11[cam.CameraRoll+1024]
			X := (z*rs + x*rc) >> 15
			z = (z*rc - x*rs) >> 15
			x = X
		}
		if cam.CameraPitch != 0 {
			ps := sine11[cam.CameraPitch]
			pc := sine11[cam.CameraPitch+1024]
			Y := (y*pc - z*ps) >> 15
			z = (y*ps + z*pc) >> 15
			y = Y
		}
		if z < clipNear {
			return 0, 0, 0, false
		}
		return int32(cx) + (x<<uint(viewDist))/z, int32(cy) + (y<<uint(viewDist))/z, z, true
	}

	type drawItem struct {
		sx, feetY, camZ int32
		worldW, worldH  int
		cs              *CompositeSprite
	}
	var items []drawItem
	add := func(lx, ly, worldW, worldH int, cs *CompositeSprite) {
		if cs == nil {
			return
		}
		sx, feetY, camZ, ok := project(lx, ly)
		if !ok {
			return
		}
		items = append(items, drawItem{sx, feetY, camZ, worldW, worldH, cs})
	}

	// camTerm is the camera's contribution to the 8-way facing index. v.Rotation
	// IS the authentic cameraRotation (0..255): facing = (heading + (rot+16)/32)&7
	// (drawNpc:2099 / drawPlayer:2934). It selects which of the 8 sprite poses
	// (and the W/SW/NW mirror) to draw, so the visible side rotates with the
	// camera even while every Heading is 0 (Phase 3a).
	camTerm := (v.Rotation + 16) / 32

	// playerSprite picks the real-appearance composite when equipment is known,
	// falling back to the default-human composite (and finally letting the
	// caller's nil-guard / 3D-cross handle a total failure). Shared by other
	// players and the local player so both honour the same appearance path.
	playerSprite := func(hasEquip bool, equip [12]int, hair, top, trouser, skin, facing int) *CompositeSprite {
		if hasEquip {
			if cs := compositePlayerAppearance(equip, hair, top, trouser, skin, facing); cs != nil {
				return cs
			}
		}
		return compositePlayer(facing)
	}

	if os.Getenv("RENDER_NO_ENTITIES") == "" {
		for _, e := range ents {
			facing := (e.Heading + camTerm) & 7
			switch e.Kind {
			case EntityNPC:
				w, h := npcBillboardSize(e.NpcID)
				add(e.X-baseX, e.Y-baseY, w, h, compositeNPC(e.NpcID, facing))
			default: // EntityPlayer / other players
				cs := playerSprite(e.HasEquip, e.EquipSprites, e.HairColour, e.TopColour, e.TrouserColour, e.SkinColour, facing)
				add(e.X-baseX, e.Y-baseY, playerBillboardW, playerBillboardH, cs)
			}
		}
		// the local player is just another depth-sorted actor at his own tile. His
		// facing combines his own server heading (v.SelfHeading, 0 when unknown)
		// with the camera term, so he turns both as he walks and as the camera
		// pans. His appearance (equipment + colours) is mirrored from world.Self
		// onto the View.
		if !v.NoSelf {
			selfFacing := (v.SelfHeading + camTerm) & 7
			cs := playerSprite(v.SelfHasEquip, v.SelfEquipSprites, v.SelfHairColour, v.SelfTopColour, v.SelfTrouserColour, v.SelfSkinColour, selfFacing)
			add(v.X-baseX, v.Y-baseY, playerBillboardW, playerBillboardH, cs)
		}
	}

	// Ground items are billboarded in the SAME depth-sorted pass: each is drawn
	// as its real inventory icon (compositeItem) anchored at its ground tile's
	// foot point, so it sorts against characters and is occluded by walls. The
	// world size is derived from the icon's decoded canvas (groundItemPixelToWorld
	// units per pixel), so a tall narrow icon (e.g. a staff) keeps its aspect and
	// every item reads as a small dropped object near the tile, not a billboard.
	// Icons that fail to decode are NOT added here — they were already routed to
	// the flat red marker quad in RenderView.
	if os.Getenv("RENDER_NO_GROUND_ITEMS") == "" {
		for _, gi := range v.GroundItems {
			cs := compositeItem(gi.ItemID)
			if cs == nil {
				continue
			}
			worldW := cs.W * groundItemPixelToWorld
			worldH := cs.H * groundItemPixelToWorld
			add(gi.X-baseX, gi.Y-baseY, worldW, worldH, cs)
		}
	}

	// Painter's order: far (large camZ) first, near last, so a nearer actor
	// paints over a farther one (e.g. a mob in front of the player).
	sort.SliceStable(items, func(i, j int) bool { return items[i].camZ > items[j].camZ })

	for _, it := range items {
		screenW := (int32(it.worldW) << uint(viewDist)) / it.camZ
		screenH := (int32(it.worldH) << uint(viewDist)) / it.camZ
		if screenW <= 0 || screenH <= 0 {
			continue
		}
		// Occlusion depth (FIX A): the scene depth buffer stores each face's
		// average camera-Z, so the sprite's FOOT depth is compared against the
		// nearest geometry under each pixel. Bias the sprite spriteDepthBias
		// units TOWARD the camera so a character standing AT a wall/building
		// tile (the wall face ~one tile in front of its foot centre, but at a
		// comparable average depth) is not self-occluded by that wall, while a
		// character a tile or more BEHIND the wall still tests farther and is
		// hidden. Half a tile (64) is a touch under one tile (128) so the bias
		// can't leak a sprite through a wall a full tile ahead of it.
		spriteZ := it.camZ - spriteDepthBias
		surf.BlitSpriteScaled(it.cs, int(it.sx-screenW/2), int(it.feetY-screenH), int(screenW), int(screenH), it.cs.Flip, spriteZ)
	}
}

// spriteDepthBias pushes a character billboard half a tile toward the camera
// for the depth-buffer occlusion test (see BlitSpriteScaled). One RSC tile is
// 128 world units; half a tile keeps a sprite standing against a wall from
// self-occluding without letting it punch through a wall a full tile in front.
const spriteDepthBias = 64

// skyColour is a flat sky/background fill (RSCPlus blue-grey 0x6080a0).
const skyColour = 0x6080a0

// fogStart is the camera-Z at which distance fog begins. Faces nearer than this
// render full-bright; from here to clipFar (2400) the rendered pixel is blended
// linearly toward skyColour (rasterize.go fogBlend), so distant terrain/water/
// hills fade smoothly into the blue sky instead of hard-clipping or darkening to
// black. 700-unit band ⇒ fully sky at clipFar.
const fogStart = clipFar - 700 // 1700

// planeHeightTiles is the vertical Y stride between stacked floors in the RSC
// world map (mirrors world.PlaneHeight = 944; floors are encoded as a Y
// offset, not a separate coordinate). The render window operates in
// plane-LOCAL Y (v.Y is the host's plane-local row), so converting a
// window-local Y back to the ABSOLUTE world Y the live world-state mirrors use
// is localY + plane*planeHeightTiles. Kept here so render stays independent of
// the world package.
const planeHeightTiles = 944

// sceneryPlane returns the floor/plane index for an ABSOLUTE world Y (the
// space facts.SceneryLoc.Y and the dynamic-scenery mirror use): Y/944, with
// negatives clamped to 0. Mirrors world.PlaneOf.
func sceneryPlane(y int) int {
	if y < 0 {
		return 0
	}
	return y / planeHeightTiles
}

// absWorldY converts a window-local Y (in the v.Y plane-local space) to the
// ABSOLUTE world Y the live mirrors are keyed by.
func absWorldY(localY, plane int) int { return localY + plane*planeHeightTiles }

// groundItemColour is the flat fill for a dropped-item ground marker — a warm
// red so items read against grass/stone. Flat-shaded at a fixed bright index
// like the entity billboards so the near-flat marker doesn't gouraud-darken.
var groundItemColour = method305(210, 40, 40)

const (
	// groundItemHalf is the half-extent (world units) of the small flat marker
	// quad; ~1/4 tile so a dropped item reads as a small patch, not a slab.
	groundItemHalf = 24
	// groundItemLift raises the marker just off the terrain so it isn't
	// z-fought into the ground face.
	groundItemLift = 6

	// groundItemPixelToWorld converts an item icon's decoded canvas pixels to
	// the billboard's world-space size. Item icons sit on a 48x32 canvas, so at
	// 2 world units/pixel a full icon spans 96x64 world units — about 3/4 of a
	// 128-unit tile wide — so a dropped item reads as a small object on its tile,
	// not a giant sign. The icon's bottom-of-canvas transparent padding lifts the
	// visible glyph slightly off the ground (the foot anchor is the canvas
	// bottom), so it floats just above the terrain like a dropped item.
	groundItemPixelToWorld = 2
)

// BuildGroundItems places one small flat quad on the terrain at each in-window
// ground item, into a single GameModel that depth-sorts with the rest of the
// scene. items carry ABSOLUTE world coords (X absolute, Y plane-local to match
// the window the same way Entity does). baseX/baseY anchor the window; heights
// is the flattened terrain grid so each marker sits on the ground. Returns nil
// if no item falls inside the window. Never panics on a bad tile (window
// guard).
func BuildGroundItems(items []GroundItemMarker, baseX, baseY, plane int, heights [][]int32) *GameModel {
	n := terrainSize
	type placed struct{ lx, ly int }
	var ps []placed
	for _, it := range items {
		lx := it.X - baseX
		ly := it.Y - baseY
		if lx < 0 || lx >= n || ly < 0 || ly >= n {
			continue
		}
		ps = append(ps, placed{lx, ly})
	}
	if len(ps) == 0 {
		return nil
	}
	g := NewGameModel(len(ps)*4, len(ps))
	for _, p := range ps {
		cx := int32(p.lx)*128 + 64
		cz := int32(p.ly)*128 + 64
		y := -heights[p.lx][p.ly] - groundItemLift
		v0 := g.AddVertex(cx-groundItemHalf, y, cz-groundItemHalf)
		v1 := g.AddVertex(cx+groundItemHalf, y, cz-groundItemHalf)
		v2 := g.AddVertex(cx+groundItemHalf, y, cz+groundItemHalf)
		v3 := g.AddVertex(cx-groundItemHalf, y, cz+groundItemHalf)
		g.AddFixedFace([]int{v0, v1, v2, v3}, groundItemColour, groundItemColour, entityShade)
	}
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
