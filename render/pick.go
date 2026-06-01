package render

import (
	"os"
	"sort"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// PickTile maps a screen-space click (px,py in the rendered W×H frame) to the
// world TILE the host would walk to. It avoids any fragile inverse-camera math:
// it rebuilds the EXACT camera RenderView used for this View, forward-projects
// every window tile's centre to screen with that camera (the same transform the
// renderer uses), and returns the tile whose projected centre is nearest the
// click (among tiles in front of the camera). Returns plane-LOCAL world coords
// (the caller adds the plane offset) and ok=false if nothing projects near it.
//
// Brute force over the window (~160² tiles) is fine — a click is rare and this
// is far simpler + more robust than unprojecting a ray through non-flat terrain.
func PickTile(land *pathfind.Landscape, v View, px, py int) (worldX, worldY int, ok bool) {
	if land == nil {
		return 0, 0, false
	}
	n := terrainSize
	baseX := v.X - n/2
	baseY := v.Y - n/2

	// Same camera RenderView builds (render.go): host at window centre, pitch
	// 912, yaw rotation*4, distance zoom*2. Camera height = -(host elevation*3).
	localX := int32(v.X-baseX)*128 + 64
	localZ := int32(v.Y-baseY)*128 + 64
	elev := int32(land.Tile(v.X, v.Y, v.Plane).GroundElevation) * 3
	zoom := v.Zoom
	if zoom == 0 {
		zoom = 600
	}
	cam := SetCamera(localX, -elev, localZ, 912, int32(v.Rotation)*4, 0, int32(zoom)*2)
	// RenderTo swaps yaw/roll before project(); replicate so our projection
	// matches the rendered frame pixel-for-pixel.
	cam.CameraYaw, cam.CameraRoll = cam.CameraRoll, cam.CameraYaw

	w, h := v.W, v.H
	if w <= 0 {
		w = 512
	}
	if h <= 0 {
		h = 336
	}
	cx, cy := int32(w/2), int32(h/2)

	bestD := int64(1) << 62
	bestX, bestY := 0, 0
	found := false
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			wx := int32(i)*128 + 64
			wz := int32(j)*128 + 64
			wy := -int32(land.Tile(baseX+i, baseY+j, v.Plane).GroundElevation) * 3
			sx, sy, camZ, vis := projectPoint(cam, wx, wy, wz, cx, cy)
			if !vis || camZ <= clipNear || camZ >= clipFar {
				continue
			}
			dx := int64(int(sx) - px)
			dy := int64(int(sy) - py)
			d := dx*dx + dy*dy
			if d < bestD {
				bestD = d
				bestX, bestY = baseX+i, baseY+j
				found = true
			}
		}
	}
	if !found {
		return 0, 0, false
	}
	return bestX, bestY, true
}

// projectPoint projects one world point (X,Y,Z; Y is height, -up) to screen
// pixels using the (already yaw/roll-swapped) camera, mirroring GameModel.project
// + the screen-centre offset. Returns screen (sx,sy), camera-space depth camZ,
// and vis=false if the point is behind the near plane.
func projectPoint(cam Camera, X, Y, Z, cx, cy int32) (sx, sy, camZ int32, vis bool) {
	x := X - cam.CameraX
	y := Y - cam.CameraY
	z := Z - cam.CameraZ
	if cam.CameraYaw != 0 {
		s := sine11[cam.CameraYaw]
		c := sine11[cam.CameraYaw+1024]
		nx := (y*s + x*c) >> 15
		y = (y*c - x*s) >> 15
		x = nx
	}
	if cam.CameraRoll != 0 {
		s := sine11[cam.CameraRoll]
		c := sine11[cam.CameraRoll+1024]
		nx := (z*s + x*c) >> 15
		z = (z*c - x*s) >> 15
		x = nx
	}
	if cam.CameraPitch != 0 {
		s := sine11[cam.CameraPitch]
		c := sine11[cam.CameraPitch+1024]
		ny := (y*c - z*s) >> 15
		z = (y*s + z*c) >> 15
		y = ny
	}
	if z < clipNear {
		return 0, 0, z, false
	}
	vx := (x << uint(viewDist)) / z
	vy := (y << uint(viewDist)) / z
	// Screen: the renderer plots at centre + viewX, centre + viewY (project's y
	// already carries the world's -up sign, so +vy is correct).
	return cx + vx, cy + vy, z, true
}

// TargetKind classifies a pick candidate by what the user is pointing at. It
// drives which identity fields are populated (Pick fills only the fields that
// matter for a kind; the rest stay zero) and which authentic RSC menu verbs the
// caller (Layer 2) offers. The ordering of the consts is NOT a priority — depth
// ordering is by camera distance for billboards, then a fixed tile-target
// priority for tile-grounded targets (see Pick's doc).
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

// PickCandidate is one thing under the cursor that Pick resolved. Every
// coordinate is PLANE-LOCAL (X absolute, Y plane-local) — the same space the View
// was built in and PickTile returns; the caller adds Plane*PlaneHeight to Y to
// recover the absolute Y a runtime.Host action wants, exactly as the /walk
// handler does. A candidate is self-describing: Kind selects which identity
// fields are meaningful; unused fields are zero.
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
	// centre as a degenerate (zero-area) box. Useful for a debug overlay and for
	// reproducing the hit in tests; Layer 2 ignores it.
	ScreenRect [4]int
	// CamZ is the camera-space depth used for ordering (smaller = nearer). For
	// billboards it is the foot point's depth; for tile targets it is the tile
	// centre's depth. Exposed so callers can reason about ties; do not dispatch on it.
	CamZ int32
}

// Pick maps a screen-space click (px, py in the rendered W×H frame described by
// v) to a depth-ordered list of world targets under the cursor. Ordering is:
//
//  1. Billboards (NPC / player / self / ground item) whose AABB contains (px, py),
//     NEAREST-CAMERA-FIRST — the reverse of the renderer's far-first painter
//     order, so the sprite drawn last (hence visually on top) is candidate #0.
//  2. The tile under the cursor (PickTile): its scenery objects, then its boundary
//     edges, then the bare terrain tile, in that fixed priority.
//
// So candidates[0] is the topmost thing the user is pointing at, and the list
// always ends with a TargetTerrain "Walk here" fallback (whenever PickTile
// resolves a tile), mirroring RSC's "Walk here" always being the last menu entry.
//
// Pick is a PURE read: it sends no packets, mutates no world state, and renders
// nothing. It is the inverse-lookup companion to RenderView and agrees with the
// rendered frame pixel-for-pixel — a billboard hit corresponds to a sprite that
// actually drew (Pick composites with the SAME composite* functions the renderer
// uses and skips any that fail, so an actor that fell back to the 3D-cross has no
// tight billboard box but is still reachable via its tile).
//
// f may be nil — then no scenery/boundary enumeration happens (billboards +
// terrain only). land may be nil — then there are NO candidates at all (the
// list is empty): billboard projection needs the terrain grid + host-tile
// elevation, and tile targets obviously need the landscape, so with land==nil
// neither branch can run. Every returned X/Y is PLANE-LOCAL (the caller folds
// Plane into absolute Y before dispatch).
func Pick(land *pathfind.Landscape, f *facts.Facts, v View, px, py int) []PickCandidate {
	var out []PickCandidate

	// --- billboards (§2/§3): project foot points with the SAME shared helper +
	// the SAME enumeration order DrawEntitySprites uses, so a hit corresponds to a
	// sprite the renderer would have blitted. land must be non-nil for the
	// projection (it reads the host-tile elevation + the terrain-height grid).
	if land != nil {
		out = pickBillboards(land, v, px, py)
	}

	// --- tile-grounded targets (§4): scenery, then boundaries, then terrain,
	// appended AFTER all billboards. Resolved through the existing PickTile.
	if land != nil {
		tx, ty, ok := PickTile(land, v, px, py)
		if ok {
			out = append(out, pickTileTargets(land, f, v, tx, ty)...)
		}
	}

	return out
}

// Composite-function indirection. The picker resolves each billboard's sprite
// through these vars so it honours the SAME sprite-presence gate the renderer
// does (a composite that returns nil was routed to the 3D-cross fallback and not
// blitted, so it gets no tight billboard box). They default to the real
// composite* functions — production behaviour is unchanged — and exist as a seam
// so the projection/AABB/ordering tests can exercise the picker's geometry
// deterministically without the sprite archives installed (those archives live at
// hardcoded paths / env vars that are absent in CI). Tests swap these out and
// restore them; nothing else assigns them.
var (
	pickCompositeNPC    = compositeNPC
	pickCompositeItem   = compositeItem
	pickCompositePlayer = pickPlayerSprite
)

// pickBillboards builds the billboard candidate set in the renderer's exact
// enumeration order (v.Entities, then self, then v.GroundItems — matching
// DrawEntitySprites), tests the click against each composited sprite's AABB, and
// returns the hits sorted NEAREST-CAMERA-FIRST (stable, so equal-depth ties keep
// enumeration order: NPCs before players before self before ground items).
func pickBillboards(land *pathfind.Landscape, v View, px, py int) []PickCandidate {
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)

	var hits []PickCandidate
	// test composites one billboard: project its foot + AABB exactly as the
	// renderer does, and append a candidate if (px,py) lands in the box. cs==nil
	// (composite failed → routed to the 3D-cross fallback, not blitted) is skipped
	// so a hit always corresponds to a real on-screen sprite.
	test := func(lx, ly int, ox, oz int32, worldW, worldH int, cs *CompositeSprite, c PickCandidate) {
		if cs == nil {
			return
		}
		rect, _, _, camZ, ok := projectBillboard(cam, cx, cy, heights, lx, ly, ox, oz, worldW, worldH)
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

	// camTerm is the camera's contribution to the 8-way facing index, identical to
	// DrawEntitySprites (render.go): facing = (heading + (rot+16)/32) & 7. The
	// facing/step pick the SAME sprite pose the renderer composited, so the AABB
	// (which depends on the composite's W/H for ground items) matches the frame.
	camTerm := (v.Rotation + 16) / 32

	if os.Getenv("RENDER_NO_ENTITIES") == "" {
		for _, e := range v.Entities {
			facing := (e.Heading + camTerm) & 7
			ox, oz := int32(e.OffX), int32(e.OffZ)
			step := e.StepPhase
			switch e.Kind {
			case EntityNPC:
				w, h := npcBillboardSize(e.NpcID)
				cs := pickCompositeNPC(e.NpcID, facing, step)
				test(e.X-baseX, e.Y-baseY, ox, oz, w, h, cs, PickCandidate{
					Kind: TargetNPC, Plane: v.Plane, X: e.X, Y: e.Y, Index: e.Index, NpcID: e.NpcID,
				})
			default: // EntityPlayer / other players
				cs := pickCompositePlayer(e.HasEquip, e.EquipSprites, e.HairColour, e.TopColour, e.TrouserColour, e.SkinColour, facing, step)
				test(e.X-baseX, e.Y-baseY, ox, oz, playerBillboardW, playerBillboardH, cs, PickCandidate{
					Kind: TargetPlayer, Plane: v.Plane, X: e.X, Y: e.Y, Index: e.Index,
				})
			}
		}
		// the local player is just another depth-sorted actor at his own tile; his
		// foot tile is (v.X, v.Y) with the glide offset, and he is server index 0
		// (the cradle skips index 0 from the other-players loop).
		if !v.NoSelf {
			selfFacing := (v.SelfHeading + camTerm) & 7
			cs := pickCompositePlayer(v.SelfHasEquip, v.SelfEquipSprites, v.SelfHairColour, v.SelfTopColour, v.SelfTrouserColour, v.SelfSkinColour, selfFacing, v.SelfStepPhase)
			test(v.X-baseX, v.Y-baseY, int32(v.SelfOffX), int32(v.SelfOffZ), playerBillboardW, playerBillboardH, cs, PickCandidate{
				Kind: TargetSelf, Plane: v.Plane, X: v.X, Y: v.Y, Index: 0,
			})
		}
	}

	// ground items: real inventory icon, world size = icon canvas * groundItemPixelToWorld,
	// no glide offset (render.go) — identical to the renderer's ground-item pass.
	if os.Getenv("RENDER_NO_GROUND_ITEMS") == "" {
		for _, gi := range v.GroundItems {
			cs := pickCompositeItem(gi.ItemID)
			if cs == nil {
				continue
			}
			worldW := cs.W * groundItemPixelToWorld
			worldH := cs.H * groundItemPixelToWorld
			test(gi.X-baseX, gi.Y-baseY, 0, 0, worldW, worldH, cs, PickCandidate{
				Kind: TargetGroundItem, Plane: v.Plane, X: gi.X, Y: gi.Y, ItemID: gi.ItemID,
			})
		}
	}

	// nearest-camera-first: the REVERSE of the renderer's far-first painter sort
	// (render.go), stable so equal-depth ties preserve enumeration order.
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].CamZ < hits[j].CamZ })
	return hits
}

// pickPlayerSprite mirrors DrawEntitySprites' playerSprite closure: prefer the
// real-appearance composite when equipment is known, falling back to the
// default-human composite. Shared by other players and the local player so the
// picker composites the SAME sprite the renderer drew.
func pickPlayerSprite(hasEquip bool, equip [12]int, hair, top, trouser, skin, facing, step int) *CompositeSprite {
	if hasEquip {
		if cs := compositePlayerAppearance(equip, hair, top, trouser, skin, facing, step); cs != nil {
			return cs
		}
	}
	return compositePlayer(facing, step)
}

// pickTileTargets enumerates the scenery / boundary / terrain targets on the
// picked tile (tx, ty are PLANE-LOCAL, from PickTile), in the fixed priority
// dynamic-scenery → static-scenery → boundary → terrain. The list always ends
// with exactly one TargetTerrain "Walk here" target (the lowest-priority,
// always-present fallback). CamZ/ScreenRect are filled from the picked tile's
// projected centre so a debug overlay can order tile vs billboard sensibly.
func pickTileTargets(land *pathfind.Landscape, f *facts.Facts, v View, tx, ty int) []PickCandidate {
	var out []PickCandidate

	// project the picked tile centre once, for CamZ/ScreenRect on every tile
	// candidate (they share the tile's geometry). Reuse the picker's camera build.
	cam, baseX, baseY, cx, cy, heights := billboardCamera(land, v)
	lx, ly := tx-baseX, ty-baseY
	var tileZ int32
	tileRect := [4]int{}
	if lx >= 0 && lx < terrainSize && ly >= 0 && ly < terrainSize {
		wx := int32(lx)*128 + 64
		wz := int32(ly)*128 + 64
		wy := -heights[lx][ly]
		if sx, sy, camZ, vis := projectPoint(cam, wx, wy, wz, int32(cx), int32(cy)); vis {
			tileZ = camZ
			tileRect = [4]int{int(sx), int(sy), int(sx), int(sy)}
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
		// facts.At is keyed in ABSOLUTE space; convert the plane-local picked tile
		// Y to absolute (the absWorldY conversion the renderer uses). The scenery
		// plane guard is implicitly satisfied: facts.At only returns locs on the
		// queried absolute Y.
		absY := absWorldY(ty, v.Plane)

		// (1) dynamic scenery on this tile (live, wins over static — same precedence
		// as the renderer's dedup). v.DynamicScenery is already plane-local on the
		// View, so match on (tx, ty) directly. A dynamic entry SUPPRESSES a static
		// one on the same tile (set below).
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
				// (2) static scenery: skip if a dynamic entry already covers this
				// tile, or if the View reports it removed (mined rock / burned fire).
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
				// npc_spawn / ground_item placements are ignored: live NPCs and
				// dropped items are billboards (handled above), and a spawn point is
				// not a clickable target.
			}
		}
	}

	// (4) terrain: always present, lowest priority — the "Walk here" fallback.
	out = append(out, mk(PickCandidate{Kind: TargetTerrain}))
	return out
}
