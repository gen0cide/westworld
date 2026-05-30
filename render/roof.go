package render

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// wallObjectHeight is the default RSC full-wall height (GameData.wallObjectHeight
// fallback, matching boundary.go's 192). Roof eaves must sit on the WALL TOPS,
// so each walled corner's height is the terrain height PLUS the wall height — the
// client bakes this into terrainHeightLocal via method428 before building roofs.
const wallObjectHeight = 192

// roofDefs is the authentic RSC roof-elevation table (OpenRSC
// EntityHandler.loadElevationDefinitions, 243-248; the classic client's
// GameData.anIntArray101 = height, anIntArray102 = texture id). A tile's
// Tile.RoofTexture is a 1-BASED index into this table: rt==0 means no roof,
// rt==1 -> roofDefs[0], etc. height is how far the roof apex is raised above
// the wall tops; texture is the texture-archive id painted on the roof faces
// (id 6 = "roof" = the red clay tile).
var roofDefs = []struct{ height, texture int32 }{
	{64, 6}, {64, 3}, {96, 2}, {80, 33}, {80, 15}, {90, 49},
}

// HostUnderRoof reports whether the host standing on (worldX,worldY) is INSIDE
// a roofed building, so the roof should be hidden (the player would otherwise
// be occluded by his own roof). The authentic client hides roofs when the
// player's tile carries the objectAdjacency 0x80 "under-roof" bit
// (mudclient.java:3594); that bit is set for any tile inside a roofed footprint.
// We reproduce it directly from the loaded roof grid: the host is under roof
// when his tile AND all four edge-neighbours are roofed (a fully-enclosed
// interior tile — not a one-tile eave overhang that should stay visible).
func HostUnderRoof(land *pathfind.Landscape, worldX, worldY, plane int) bool {
	if land == nil {
		return false
	}
	roof := func(x, y int) bool {
		return land.Tile(x, y, plane).RoofTexture > 0
	}
	return roof(worldX, worldY) &&
		roof(worldX-1, worldY) && roof(worldX+1, worldY) &&
		roof(worldX, worldY-1) && roof(worldX, worldY+1)
}

// maxRoofPlane is the highest landscape plane the roof pass reads. A multi-
// story building stores each upper story's RoofTexture + walls on a SEPARATE
// archive plane (h1.../h2.../h3...; verified against Authentic_Landscape.orsc:
// the Lumbridge castle column at (111,666) is roofed on BOTH plane 0 and plane
// 1). The authentic client renders every plane's roof while the host stands at
// ground level (loadSection loads + builds planes 0..3 into separate roof
// models, all added to the scene), so we iterate the same range.
const maxRoofPlane = 3

// planeWallTopBonus is the per-plane height stride between stacked stories. The
// authentic client never resets terrainHeightLocal between the per-plane roof
// builds (loadSection(...,1/2/3,false) skips the height reset), so each upper
// story's method428 ADDS its wall height on top of the heights the lower
// stories already baked in — a story-N roof sits ~N wall-heights up. Most upper
// stories carry NO wall bytes in the .orsc (the data stores the visible facade
// walls on plane 0 and only the roof on the upper plane), so without an
// explicit per-plane bonus an upper roof would collapse back onto the ground
// roof. We add one full wall height per plane below the current one to
// reproduce the client's accumulated wall-top base. (= GameData.wallObjectHeight
// fallback, matching wallObjectHeight / boundary.go's 192.)
const planeWallTopBonus = wallObjectHeight

// BuildRoofs assembles all roof tiles inside the 96x96 window into a single
// GameModel, porting World.method422's roof pass (World.java ~1044-1206). It is
// the per-tile analogue of BuildBoundaries: for every roofed tile it builds a
// flat / sloped / hipped roof patch from the 4 corner heights, raising corners
// that sit fully inside a roofed footprint and nudging the edge corners inward
// by the "strut" test so adjacent patches meet at a ridge.
//
// MULTI-STORY: the host's plane is the GROUND story; upper stories live on the
// upper archive planes. We build a roof patch for the host plane AND every
// upper plane that carries roof tiles, anchoring each story at its ACCUMULATED
// wall-top height (each story adds a wall-height of base) so a 2-story
// building's roof sits at ~2x wall height instead of being buried inside the
// ground roof. All stories share one GameModel so they depth-sort together.
//
// heights is the flattened terrain-height grid (TerrainHeights) so a roof's
// eaves sit on the wall tops. f supplies the per-wall heights (method428) so a
// roofed corner is lifted onto the surrounding wall tops before the apex raise.
// Returns nil if no plane in the window holds roof tiles.
func BuildRoofs(f *facts.Facts, land *pathfind.Landscape, baseX, baseY, plane int, heights [][]int32) *GameModel {
	if land == nil {
		return nil
	}
	n := terrainSize

	// wallHeight resolves a wall def's vertical extent (BoundaryDef.Height,
	// = GameData.wallObjectHeight), defaulting to a full wall when unknown.
	wallHeight := func(defID int) int32 {
		if f != nil {
			if def := f.BoundaryDef(defID); def != nil && def.Height > 0 {
				return int32(def.Height)
			}
		}
		return wallObjectHeight
	}

	type tri struct {
		v       [4]int // 3 or 4 local-corner vertex slots (see corner mapping)
		nv      int    // 3 = triangle, 4 = quad
		texture int32
	}

	// One shared model + face list across every story so the patches depth-sort
	// together. addV builds one *local* vertex per (corner, tile) on demand
	// because the strut nudge perturbs a corner's X/Z PER TILE.
	g := NewGameModel((n-2)*(n-2)*6*(maxRoofPlane+1), (n-2)*(n-2)*2*(maxRoofPlane+1))
	addV := func(x, z, h int32) int { return g.AddVertex(x, -h, z) }
	var faces []tri
	any := false

	// buildStory emits the roof faces for ONE landscape plane (story). storyBase
	// is the accumulated wall-top height every corner of this story starts from
	// (terrain for the ground plane; terrain + N wall-heights for story N). It
	// reproduces the client's per-plane method428 + roof passes on a fresh
	// working-height grid anchored at storyBase. Reading an empty/void plane is
	// nil-safe (land.Tile returns the zero Tile, RoofTexture 0 -> no faces).
	buildStory := func(sp int, storyBase int32) {
		// roofAt returns the 1-based roof id at LOCAL window coords on THIS story
		// plane, 0 outside the window (getWallRoof clamps off-window to 0).
		roofAt := func(lx, ly int) int32 {
			if lx < 0 || ly < 0 || lx >= n || ly >= n {
				return 0
			}
			return int32(land.Tile(baseX+lx, baseY+ly, sp).RoofTexture)
		}
		// Quick scan: skip a story with no roof tiles entirely (most upper planes
		// are empty), avoiding the per-corner allocation below.
		hasAny := false
		for lx := 1; lx < n-1 && !hasAny; lx++ {
			for ly := 1; ly < n-1; ly++ {
				if roofAt(lx, ly) > 0 {
					hasAny = true
					break
				}
			}
		}
		if !hasAny {
			return
		}

		// hasRoofTile == method406: the corner is fully enclosed by roof, so its
		// height is RAISED to the apex (World.java:391-393).
		hasRoofTile := func(lx, ly int) bool {
			return roofAt(lx, ly) > 0 && roofAt(lx-1, ly) > 0 &&
				roofAt(lx-1, ly-1) > 0 && roofAt(lx, ly-1) > 0
		}
		// strut == method427: a corner is kept on the eave line when ANY of the 4
		// tiles sharing it is roofed (World.java:1441-1443); else nudged inward.
		strut := func(lx, ly int) bool {
			return roofAt(lx, ly) > 0 || roofAt(lx-1, ly) > 0 ||
				roofAt(lx-1, ly-1) > 0 || roofAt(lx, ly-1) > 0
		}
		// diagAt returns this story's diagonal-wall band (1..11999 = '\',
		// 12001..23999 = '/').
		diagAt := func(lx, ly int) int32 {
			if lx < 0 || ly < 0 || lx >= n || ly >= n {
				return 0
			}
			return land.Tile(baseX+lx, baseY+ly, sp).DiagonalWalls
		}

		// rh[lx][ly] is the working roof corner height for this story. It begins
		// at storyBase (terrain + accumulated lower-story wall tops) and is then
		// (a) raised onto THIS story's wall tops (method428), (b) leveled per roof
		// tile to the tile's max corner (World.java:988-1042), and (c) raised to
		// the apex for fully-enclosed corners (method406). raised[] tracks (c) so
		// a shared corner is lifted only once.
		raised := make([][]bool, n)
		rh := make([][]int32, n)
		for i := 0; i < n; i++ {
			raised[i] = make([]bool, n)
			rh[i] = make([]int32, n)
			for j := 0; j < n; j++ {
				rh[i][j] = heights[i][j] + storyBase
			}
		}

		// (a) WALL TOPS (method428, World.java:971-986): each wall edge raises
		// BOTH endpoint corners to the wall top, once per corner (the client's
		// 0x13880 sentinel locks a corner after the first wall touches it). We
		// record the MAX wall height touching each corner and add it once. Edge
		// axes mirror boundary.go.
		wallTop := make([][]int32, n)
		for i := 0; i < n; i++ {
			wallTop[i] = make([]int32, n)
		}
		raiseCorner := func(cx, cy int, h int32) {
			if h > wallTop[cx][cy] {
				wallTop[cx][cy] = h
			}
		}
		for lx := 0; lx < n-1; lx++ {
			for ly := 0; ly < n-1; ly++ {
				t := land.Tile(baseX+lx, baseY+ly, sp)
				if t.HorizontalWall > 0 {
					h := wallHeight(int(t.HorizontalWall) - 1)
					raiseCorner(lx, ly, h)
					raiseCorner(lx, ly+1, h)
				}
				if t.VerticalWall > 0 {
					h := wallHeight(int(t.VerticalWall) - 1)
					raiseCorner(lx, ly, h)
					raiseCorner(lx+1, ly, h)
				}
				if d := t.DiagonalWalls; d > 0 && d < 24000 {
					if d < 12000 {
						h := wallHeight(int(d) - 1)
						raiseCorner(lx, ly, h)
						raiseCorner(lx+1, ly+1, h)
					} else {
						h := wallHeight(int(d - 12001))
						raiseCorner(lx+1, ly, h)
						raiseCorner(lx, ly+1, h)
					}
				}
			}
		}
		for i := 0; i < n; i++ {
			for j := 0; j < n; j++ {
				rh[i][j] += wallTop[i][j]
			}
		}

		// (b) LEVEL each roof tile's 4 corners to the tile's MAX corner height
		// (World.java:988-1042), as a settle pass over a snapshot.
		snap := make([][]int32, n)
		for i := 0; i < n; i++ {
			snap[i] = make([]int32, n)
			copy(snap[i], rh[i])
		}
		for lx := 1; lx < n-1; lx++ {
			for ly := 1; ly < n-1; ly++ {
				if roofAt(lx, ly) == 0 {
					continue
				}
				top := snap[lx][ly]
				for _, c := range [3]int32{snap[lx+1][ly], snap[lx+1][ly+1], snap[lx][ly+1]} {
					if c > top {
						top = c
					}
				}
				rh[lx][ly] = top
				rh[lx+1][ly] = top
				rh[lx+1][ly+1] = top
				rh[lx][ly+1] = top
			}
		}

		// (c) raise every fully-enclosed corner to its apex height. Settle all
		// raises before emitting faces (a corner is shared by up to 4 tiles).
		for lx := 1; lx < n-1; lx++ {
			for ly := 1; ly < n-1; ly++ {
				rt := roofAt(lx, ly)
				if rt == 0 || int(rt) > len(roofDefs) {
					continue
				}
				any = true
				def := roofDefs[rt-1]
				for _, c := range [4][2]int{{lx, ly}, {lx + 1, ly}, {lx + 1, ly + 1}, {lx, ly + 1}} {
					cx, cy := c[0], c[1]
					if hasRoofTile(cx, cy) && !raised[cx][cy] {
						rh[cx][cy] += def.height
						raised[cx][cy] = true
					}
				}
			}
		}

		// emit roof faces for this story.
		for lx := 1; lx < n-1; lx++ {
			for ly := 1; ly < n-1; ly++ {
				rt := roofAt(lx, ly)
				if rt == 0 || int(rt) > len(roofDefs) {
					continue
				}
				def := roofDefs[rt-1]

				// corner heights (apex-raised in pass c)
				hA := rh[lx][ly]     // A=(lx,ly)
				hB := rh[lx+1][ly]   // B=(lx+1,ly)
				hC := rh[lx+1][ly+1] // C=(lx+1,ly+1)
				hD := rh[lx][ly+1]   // D=(lx,ly+1)

				// base plane positions (client k24/i25/k25/i26):
				//   A=(xA,zA) B=(xB,zA) C=(xB,zB) D=(xA,zB)
				xA := int32(lx) * 128
				zA := int32(ly) * 128
				xB := xA + 128
				zB := zA + 128
				// per-corner nudge (client byte0=16): pull an eave corner inward
				// on any side that lacks a strutted (roofed) neighbour, so adjacent
				// patches meet on the ridge (World.java:1094-1126).
				const nudge = 16
				// corner A=(lx,ly): client j14/k16
				axA, azA := xA, zA
				if !strut(lx-1, ly) {
					axA -= nudge
				}
				if !strut(lx+1, ly) {
					axA += nudge
				}
				if !strut(lx, ly-1) {
					azA -= nudge
				}
				if !strut(lx, ly+1) {
					azA += nudge
				}
				// corner B=(lx+1,ly): client l18/k19
				axB, azB := xB, zA
				if !strut(lx, ly) {
					axB -= nudge
				}
				if !strut(lx+2, ly) {
					axB += nudge
				}
				if !strut(lx+1, ly-1) {
					azB -= nudge
				}
				if !strut(lx+1, ly+1) {
					azB += nudge
				}
				// corner C=(lx+1,ly+1): client k21/i23
				axC, azC := xB, zB
				if !strut(lx, ly+1) {
					axC -= nudge
				}
				if !strut(lx+2, ly+1) {
					axC += nudge
				}
				if !strut(lx+1, ly) {
					azC -= nudge
				}
				if !strut(lx+1, ly+2) {
					azC += nudge
				}
				// corner D=(lx,ly+1): client k23/i24
				axD, azD := xA, zB
				if !strut(lx-1, ly+1) {
					axD -= nudge
				}
				if !strut(lx+1, ly+1) {
					axD += nudge
				}
				if !strut(lx, ly) {
					azD -= nudge
				}
				if !strut(lx, ly+2) {
					azD += nudge
				}

				// build the 4 corner vertices for THIS tile
				vA := addV(axA, azA, hA)
				vB := addV(axB, azB, hB)
				vC := addV(axC, azC, hC)
				vD := addV(axD, azD, hD)

				// Face topology (World.java:1132-1198). Corner identities verified
				// against World.java vertexAt args:
				//   A=(k24,j27,i25) B=(k25,k27,k26) C=(l26,l27,i26) D=(j26,i28,i27)
				//   heights j27=hA k27=hB l27=hC i28=hD
				d := diagAt(lx, ly)
				switch {
				case d > 12000 && d < 24000 && roofAt(lx-1, ly-1) == 0:
					// '/' diagonal, SW open -> hip tri (World.java:1133, ai8={C,D,B})
					faces = append(faces, tri{v: [4]int{vC, vD, vB}, nv: 3, texture: def.texture})
				case d > 12000 && d < 24000 && roofAt(lx+1, ly+1) == 0:
					// '/' diagonal, NE open (World.java:1139, ai9={A,B,D})
					faces = append(faces, tri{v: [4]int{vA, vB, vD}, nv: 3, texture: def.texture})
				case d > 0 && d < 12000 && roofAt(lx+1, ly-1) == 0:
					// '\' diagonal, SE open (World.java:1145, ai10={D,A,C})
					faces = append(faces, tri{v: [4]int{vD, vA, vC}, nv: 3, texture: def.texture})
				case d > 0 && d < 12000 && roofAt(lx-1, ly+1) == 0:
					// '\' diagonal, NW open (World.java:1151, ai11={B,C,A})
					faces = append(faces, tri{v: [4]int{vB, vC, vA}, nv: 3, texture: def.texture})
				case hA == hB && hC == hD:
					// W & E edges level -> a single sloped/flat quad
					// (World.java:1156, ai12={A,B,C,D}).
					faces = append(faces, tri{v: [4]int{vA, vB, vC, vD}, nv: 4, texture: def.texture})
				case hA == hD && hB == hC:
					// N & S edges level -> a single quad (World.java:1163, ai13={D,A,B,C}).
					faces = append(faces, tri{v: [4]int{vD, vA, vB, vC}, nv: 4, texture: def.texture})
				default:
					// two-triangle split. flag1 chooses which diagonal the ridge
					// runs along (World.java:1170-1198).
					flag1 := true
					if roofAt(lx-1, ly-1) > 0 {
						flag1 = false
					}
					if roofAt(lx+1, ly+1) > 0 {
						flag1 = false
					}
					if !flag1 {
						// ridge along B-C-... (ai14={B,C,A}, ai16={D,A,C})
						faces = append(faces, tri{v: [4]int{vB, vC, vA}, nv: 3, texture: def.texture})
						faces = append(faces, tri{v: [4]int{vD, vA, vC}, nv: 3, texture: def.texture})
					} else {
						// ridge along A-B-... (ai15={A,B,D}, ai17={C,D,B})
						faces = append(faces, tri{v: [4]int{vA, vB, vD}, nv: 3, texture: def.texture})
						faces = append(faces, tri{v: [4]int{vC, vD, vB}, nv: 3, texture: def.texture})
					}
				}
			}
		}
	}

	// Story 0 is the host's plane (the ground floor); each higher story sits a
	// full wall-height higher (the client's never-reset terrainHeightLocal). We
	// iterate from the host plane up through maxRoofPlane. storyBase is the
	// accumulated wall-top base for the story (0 for the ground floor).
	for sp := plane; sp <= maxRoofPlane; sp++ {
		storyBase := int32(sp-plane) * planeWallTopBonus
		buildStory(sp, storyBase)
	}

	if !any || len(faces) == 0 {
		return nil
	}

	for _, fc := range faces {
		verts := make([]int, fc.nv)
		copy(verts, fc.v[:fc.nv])
		// fill=texture on the front, magic on the back so the textured-span path
		// paints the roof tile texture and the underside (back face) is culled.
		// intensity=magic marks the face GOURAUD (smooth per-vertex shading),
		// matching the client's World.java:1205 setLight(true,...) which forces
		// every roof faceIntensity=magic — and matching how terrain.go already
		// shades land. With intensity=0 the roof was FLAT-shaded per face, so on a
		// wide (3-4 tile) roof the per-face normal sign flipped between the near
		// and far slope tiles and the whole top read as creased/inverted. Gouraud
		// averages the corner normals so a wide roof shades as a smooth convex top.
		g.AddFace(verts, fc.texture, magic, magic)
	}
	// Authentic roof light (World.java:1205 setLight(true,50,50,-50,-10,-50)).
	g.SetLight(50, 50, -50, -10, -50)
	return g
}
