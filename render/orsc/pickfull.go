package orsc

import (
	"os"
	"sort"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/render"
)

// pickfull.go is the orsc forward-port of the classic render/pick.go full picker
// (deleted with the classic renderer in 76b1a48). PickTile (pick.go) maps a click
// to a single walk-target tile; this is the multi-candidate hit-test the browser
// remote-client right-click menu needs (remoteclient.BuildCandidates consumes
// []PickCandidate + the TargetKind enum). It enumerates EVERY actor/item billboard
// and the picked tile's scenery/boundary/terrain targets, depth-ordered, agreeing
// with the orsc-rendered frame because it reuses the SAME camera build and billboard
// projection (Scene.projectPoint + the (worldSize<<vp)/z size) RenderView uses.
//
// Ported from feat/remote-client render/pick.go (Pick/PickCandidate/TargetKind) +
// render/hittest.go (billboardCamera/projectBillboard), re-expressed on orsc's
// camera/projection. The classic helpers' yaw/roll axis swap is NOT replicated here:
// orsc's SetCamera + projectPoint already encode the authentic three/ inverse
// projection (Scene.java:2586-2591), exactly as RenderView/PickTile rely on.

// TargetKind classifies a pick candidate by what the user is pointing at — the
// int enum remoteclient.kindToWire maps to its wire string. Ordering is NOT a
// priority; billboards order by camera depth, tile targets by a fixed priority
// (see Pick). Mirrors the classic render.TargetKind 1:1 so the wire mapping is
// unchanged.
type TargetKind int

const (
	TargetNPC        TargetKind = iota // a server NPC billboard
	TargetPlayer                       // another player's billboard
	TargetSelf                         // the local player's own billboard
	TargetGroundItem                   // a dropped-item icon billboard
	TargetScenery                      // a scenery GameObject on a tile (static or dynamic)
	TargetBoundary                     // a wall/door/fence edge on a tile
	TargetTerrain                      // the bare ground tile (always-present "Walk here")
)

// groundItemPixelToWorld scales a ground item's icon-canvas pixels to billboard
// world units (render/pick.go groundItemPixelToWorld). The icon is small on the
// ground, so the world-space billboard is icon-pixels * this.
const groundItemPixelToWorld = 2

// planeHeightTiles is the per-floor Y offset facts placements carry (world.PlaneHeight
// = 944). pickTileTargets converts the plane-local picked Y to the absolute Y facts.At
// is keyed on. Mirrors render/render.go planeHeightTiles.
const planeHeightTiles = 944

// absWorldY folds the plane offset into a plane-local Y to recover the absolute Y
// facts placements are keyed on (render/render.go absWorldY).
func absWorldY(localY, plane int) int { return localY + plane*planeHeightTiles }

// PickCandidate is one thing under the cursor that Pick resolved. Every coordinate
// is PLANE-LOCAL (X absolute, Y plane-local) — the space the View was built in and
// PickTile returns; the caller folds Plane into absolute Y before dispatch. Kind
// selects which identity fields are meaningful; unused fields are zero. Field-for-
// field identical to the classic render.PickCandidate so remoteclient consumes it
// unchanged.
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
	Direction int  // boundary edge dir (arg to Host.InteractWithBoundary). 0 otherwise.
	Dynamic   bool // scenery only: true if from live View.DynamicScenery, false if static facts.

	// --- hit geometry (for debugging / tie-breaks / overlay) ---
	// ScreenRect is the on-screen AABB the candidate was tested against, in frame
	// pixels: [MinX, MinY, MaxX, MaxY]. For tile targets it is the projected tile
	// centre as a degenerate (zero-area) box.
	ScreenRect [4]int
	// CamZ is the camera-space depth used for ordering (smaller = nearer). For
	// billboards it is the foot point's depth; for tile targets the tile centre's.
	CamZ int32
}

// pickScene carries the per-pick camera + window the billboard/tile projections
// share, built ONCE so every candidate projects through the EXACT camera the orsc
// frame rendered with (RenderView/PickTile). It owns an orsc Scene only for its
// camera transform (projectPoint); nothing is rasterized.
type pickScene struct {
	scene        *Scene
	baseX, baseY int
	plane        int
	land         *pathfind.Landscape
}

// newPickScene reconstructs the orsc camera RenderView/PickTile build for this View:
// host at the window centre (+ glide offset), camera height = -ground elevation,
// pitch cameraPitch, yaw Rotation*4, distance Zoom*2 — identical to view.go RenderView
// so a projected foot/tile lands where the frame drew it. W/H/Zoom default exactly
// as RenderView (512/334/600).
func newPickScene(land *pathfind.Landscape, v render.View) *pickScene {
	w, h := v.W, v.H
	if w <= 0 {
		w = 512
	}
	if h <= 0 {
		h = 334
	}
	baseX := v.X - windowCentreTile
	baseY := v.Y - windowCentreTile
	plane := v.Plane

	surf := NewSurface(w, h)
	scene := NewScene(surf, maxModels, maxPolygons, maxSpriteFaces)
	cx := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffX)
	cz := int32(windowCentreTile*tileWorldUnits + tileWorldUnits/2 + v.SelfOffZ)
	elev := elevationOf(land, baseX, baseY, plane, cx, cz)
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	scene.fogLandscapeDistance = fogLandscape
	scene.fogEntityDistance = fogLandscape
	scene.SetCamera(cx, -elev, cz, cameraPitch, int32(v.Rotation)*4, 0, int32(zoom)*2)
	return &pickScene{scene: scene, baseX: baseX, baseY: baseY, plane: plane, land: land}
}

// projectFoot projects a billboard foot at window-local tile (lx,ly) with sub-tile
// glide (ox,oz) and derives its on-screen AABB for a worldW×worldH billboard. It is
// the orsc analogue of render/hittest.go projectBillboard: the foot is projected via
// Scene.projectPoint (the SAME transform the rendered billboard used), the screen size
// is (world<<vp)/camZ (scene.go:561-562,597-598 — orsc's own billboard size), and the
// rect is [sx-w/2, footY-h, sx+w/2, footY] (scene.go:600-601 x0=p6-w/2, y0 top = foot-h).
// ok=false on the two rejects the renderer applies (behind near plane, non-positive
// screen size), so a box exists iff the orsc frame would have blitted the sprite.
func (ps *pickScene) projectFoot(lx, ly int, ox, oz int32, worldW, worldH int) (rect [4]int, footX, footY, camZ int32, ok bool) {
	if lx < 0 || lx >= worldWindowTiles || ly < 0 || ly >= worldWindowTiles {
		return rect, 0, 0, 0, false
	}
	b := &terrainBuilder{land: ps.land, baseX: ps.baseX, baseY: ps.baseY, plane: ps.plane}
	wx := int32(lx*tileWorldUnits+tileWorldUnits/2) + ox
	wz := int32(ly*tileWorldUnits+tileWorldUnits/2) + oz
	wy := -b.getElevation(wx, wz)
	sx, sy, z := ps.scene.projectPoint(wx, wy, wz)
	if z <= rot1024ZTop || z >= ps.scene.fogEntityDistance {
		return rect, 0, 0, 0, false
	}
	// Absolute frame pixels: projectPoint returns projection-centre-relative x/y;
	// add the raster origin (mZb = W/2, mNb = H/2), exactly as endScene's blit does
	// (scene.go:601-602 x0+mZb, y0 = mNb-(h-p2)).
	footX = sx + ps.scene.mZb
	footY = sy + ps.scene.mNb
	camZ = z

	w := (int32(worldW) << uint(ps.scene.rot1024VpSrc)) / camZ
	hgt := (int32(worldH) << uint(ps.scene.rot1024VpSrc)) / camZ
	if w <= 0 || hgt <= 0 {
		return rect, footX, footY, camZ, false
	}
	left := int(footX) - int(w)/2
	rect = [4]int{left, int(footY) - int(hgt), left + int(w), int(footY)}
	return rect, footX, footY, camZ, true
}

// Pick maps a screen-space click (px,py in the rendered W×H frame) to a depth-
// ordered list of world targets under the cursor, ordered:
//
//  1. Billboards (NPC/player/self/ground item) whose AABB contains (px,py),
//     NEAREST-CAMERA-FIRST (the reverse of the painter sort, so the sprite drawn
//     last — visually on top — is candidate #0).
//  2. The picked tile's scenery, then boundaries, then bare terrain "Walk here".
//
// candidates[0] is the topmost thing; the list always ends with a TargetTerrain
// "Walk here" fallback. Pick is a PURE read: no packets, no world mutation, no
// raster. f may be nil (billboards + terrain only). land may be nil (NO candidates).
// Every X/Y is PLANE-LOCAL. Faithful orsc port of render.Pick.
func Pick(land *pathfind.Landscape, f *facts.Facts, v render.View, px, py int) []PickCandidate {
	if land == nil {
		return nil
	}
	ps := newPickScene(land, v)

	out := ps.pickBillboards(f, v, px, py)

	if tx, ty, ok := PickTile(land, v, px, py); ok {
		out = append(out, ps.pickTileTargets(f, v, tx, ty)...)
	}
	return out
}

// Composite-function indirection — the picker resolves each billboard's sprite
// through these vars so it honours the SAME sprite-presence gate the renderer does
// (a composite that returns nil routed to the 3D-cross fallback). They default to
// the exported render composite wrappers (production unchanged) and exist as a test
// seam. (Ported from render/pick.go's pickComposite* vars.)
var (
	pickCompositeItem   = render.CompositeItemSprite
	pickCompositeNPC    = render.CompositeNPCSprite
	pickCompositePlayer = pickPlayerSprite
)

// pickBillboards builds the billboard candidate set in the renderer's exact
// enumeration order (v.Entities, then self, then v.GroundItems — matching
// addViewEntities), tests the click against each composited sprite's AABB, and
// returns hits NEAREST-CAMERA-FIRST (stable, so equal-depth ties keep enumeration
// order). Mirrors render/pick.go pickBillboards on orsc primitives.
func (ps *pickScene) pickBillboards(f *facts.Facts, v render.View, px, py int) []PickCandidate {
	var hits []PickCandidate

	// requireSprite gates on the composite: GROUND ITEMS pass true (their AABB size
	// is derived FROM the icon; a failed icon draws a flat marker, not a billboard).
	// ACTORS pass false: an actor whose composite is nil still draws via the 3D-cross
	// fallback, so it stays clickable; the fixed billboard AABB approximates it.
	test := func(lx, ly int, ox, oz int32, worldW, worldH int, cs *render.CompositeSprite, requireSprite bool, c PickCandidate) {
		if requireSprite && cs == nil {
			return
		}
		rect, _, _, camZ, ok := ps.projectFoot(lx, ly, ox, oz, worldW, worldH)
		if !ok {
			return
		}
		if px < rect[0] || px > rect[2] || py < rect[1] || py > rect[3] {
			return
		}
		c.ScreenRect = rect
		c.CamZ = camZ
		hits = append(hits, c)
	}

	// facing = (heading + camTerm) & 7, camTerm = (Rotation+16)/32 — the exact
	// addViewEntities formula, so we composite the SAME pose the renderer drew.
	camTerm := (v.Rotation + 16) / 32

	if os.Getenv("RENDER_NO_ENTITIES") == "" {
		for _, e := range v.Entities {
			facing := (e.Heading + camTerm) & 7
			ox, oz := int32(e.OffX), int32(e.OffZ)
			step := e.StepPhase
			switch e.Kind {
			case render.EntityNPC:
				w, h := render.NPCBillboardSize(f, e.NpcID)
				cs := pickCompositeNPC(f, e.NpcID, facing, step)
				test(e.X-ps.baseX, e.Y-ps.baseY, ox, oz, w, h, cs, false, PickCandidate{
					Kind: TargetNPC, Plane: v.Plane, X: e.X, Y: e.Y, Index: e.Index, NpcID: e.NpcID,
				})
			default: // EntityPlayer / other players (145×220, the addViewEntities size)
				cs := pickCompositePlayer(e.HasEquip, e.EquipSprites, e.HairColour, e.TopColour, e.TrouserColour, e.SkinColour, facing, step)
				test(e.X-ps.baseX, e.Y-ps.baseY, ox, oz, 145, 220, cs, false, PickCandidate{
					Kind: TargetPlayer, Plane: v.Plane, X: e.X, Y: e.Y, Index: e.Index,
				})
			}
		}
		// the local player is just another depth-sorted actor at his own tile; he is
		// server index 0 (the cradle skips index 0 from the other-players loop).
		if !v.NoSelf {
			selfFacing := (v.SelfHeading + camTerm) & 7
			cs := pickCompositePlayer(v.SelfHasEquip, v.SelfEquipSprites, v.SelfHairColour, v.SelfTopColour, v.SelfTrouserColour, v.SelfSkinColour, selfFacing, v.SelfStepPhase)
			test(v.X-ps.baseX, v.Y-ps.baseY, int32(v.SelfOffX), int32(v.SelfOffZ), 145, 220, cs, false, PickCandidate{
				Kind: TargetSelf, Plane: v.Plane, X: v.X, Y: v.Y, Index: 0,
			})
		}
	}

	// ground items: real inventory icon, world billboard = the AUTHENTIC LITERAL 96x64
	// (Mudclient.java:6562 addSprite(40000+itemId, …, 96, 64, 109) — Task B2), no glide
	// offset, so pick == draw. The renderer's ground-item pass (entityspec.go placeGroundItem
	// -> AddEntityLayers at itemBillboardW=96 x itemBillboardH=64) projects the foot with the
	// SAME billboard size, so the picked AABB is the SAME rect the item billboard occupies.
	//
	// Historically this used cs.W*groundItemPixelToWorld x cs.H*2 (= 48*2 x 32*2), which
	// happens to ALSO equal 96x64 because compositeItem always builds onto the fixed 48x32
	// inventory canvas — but that equality was incidental (it relied on the canvas size AND
	// groundItemPixelToWorld=2). Using the literal 96x64 makes the pick AABB == the authentic
	// render billboard EXPLICITLY (B2 critique #5: reconcile pick to the 96x64 render
	// billboard). cs is still required (the requireSprite gate) so an undecodable item is not
	// clickable — only its WORLD SIZE no longer derives from the canvas.
	if os.Getenv("RENDER_NO_GROUND_ITEMS") == "" {
		for _, gi := range v.GroundItems {
			cs := pickCompositeItem(f, gi.ItemID)
			if cs == nil {
				continue
			}
			test(gi.X-ps.baseX, gi.Y-ps.baseY, 0, 0, itemBillboardW, itemBillboardH, cs, true, PickCandidate{
				Kind: TargetGroundItem, Plane: v.Plane, X: gi.X, Y: gi.Y, ItemID: gi.ItemID,
			})
		}
	}

	// nearest-camera-first: the REVERSE of the painter sort, stable so equal-depth
	// ties preserve enumeration order.
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].CamZ < hits[j].CamZ })
	return hits
}

// pickPlayerSprite mirrors addViewEntities' playerComposite: prefer the real-
// appearance composite when equipment is known, else the default human. Shared by
// other players and the local player so the picker composites the SAME sprite drawn.
func pickPlayerSprite(hasEquip bool, equip [12]int, hair, top, trouser, skin, facing, step int) *render.CompositeSprite {
	if hasEquip {
		if cs := render.CompositePlayerAppearanceSprite(equip, hair, top, trouser, skin, facing, step); cs != nil {
			return cs
		}
	}
	return render.CompositePlayerSprite(facing, step)
}

// pickTileTargets enumerates the scenery/boundary/terrain targets on the picked tile
// (tx,ty PLANE-LOCAL, from PickTile), in the fixed priority dynamic-scenery →
// static-scenery → boundary → terrain. The list always ends with exactly one
// TargetTerrain "Walk here" fallback. CamZ/ScreenRect are the picked tile's projected
// centre. Mirrors render/pick.go pickTileTargets.
func (ps *pickScene) pickTileTargets(f *facts.Facts, v render.View, tx, ty int) []PickCandidate {
	var out []PickCandidate

	// project the picked tile centre once for CamZ/ScreenRect on every tile candidate.
	lx, ly := tx-ps.baseX, ty-ps.baseY
	var tileZ int32
	tileRect := [4]int{}
	if lx >= 0 && lx < worldWindowTiles && ly >= 0 && ly < worldWindowTiles {
		b := &terrainBuilder{land: ps.land, baseX: ps.baseX, baseY: ps.baseY, plane: ps.plane}
		wx := int32(lx*tileWorldUnits + tileWorldUnits/2)
		wz := int32(ly*tileWorldUnits + tileWorldUnits/2)
		wy := -b.getElevation(wx, wz)
		if sx, sy, z := ps.scene.projectPoint(wx, wy, wz); z > rot1024ZTop {
			tileZ = z
			scrX := int(sx + ps.scene.mZb)
			scrY := int(sy + ps.scene.mNb)
			tileRect = [4]int{scrX, scrY, scrX, scrY}
		}
	}
	mk := func(c PickCandidate) PickCandidate {
		c.Plane = v.Plane
		c.X, c.Y = tx, ty
		c.CamZ = tileZ
		c.ScreenRect = tileRect
		return c
	}

	if f != nil {
		absY := absWorldY(ty, v.Plane)

		// (1) dynamic scenery on this tile (live, wins over static). v.DynamicScenery
		// is plane-local; match on (tx,ty). A dynamic entry SUPPRESSES a static one.
		dynOnTile := false
		for _, ds := range v.DynamicScenery {
			if ds.X == tx && ds.Y == ty {
				dynOnTile = true
				out = append(out, mk(PickCandidate{
					Kind: TargetScenery, DefID: ds.ID, Direction: ds.Direction, Dynamic: true,
				}))
			}
		}

		for _, p := range f.At(tx, absY) {
			switch p.Kind {
			case "scenery":
				// (2) static scenery: skip if a dynamic entry already covers this tile,
				// or if the View reports it removed (mined rock / burned fire).
				if dynOnTile {
					continue
				}
				if v.SceneryRemoved != nil && v.SceneryRemoved(p.X, p.Y) {
					continue
				}
				out = append(out, mk(PickCandidate{
					Kind: TargetScenery, DefID: p.DefID, Direction: p.Direction, Dynamic: false,
				}))
			case "boundary":
				// (3) boundaries: skip those the View reports removed (door opened,
				// web cut). p.Direction is exactly InteractWithBoundary's 3rd arg.
				if v.BoundaryRemoved != nil && v.BoundaryRemoved(p.X, p.Y, p.Direction) {
					continue
				}
				out = append(out, mk(PickCandidate{
					Kind: TargetBoundary, DefID: p.DefID, Direction: p.Direction,
				}))
				// npc_spawn / ground_item placements are ignored (live NPCs and dropped
				// items are billboards above; a spawn point is not a clickable target).
			}
		}
	}

	// (4) terrain: always present, lowest priority — the "Walk here" fallback.
	out = append(out, mk(PickCandidate{Kind: TargetTerrain}))
	return out
}
