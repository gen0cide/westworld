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

// BuildRoofs assembles all roof tiles inside the 96x96 window into a single
// GameModel, porting World.method422's roof pass (World.java ~1044-1206). It is
// the per-tile analogue of BuildBoundaries: for every roofed tile it builds a
// flat / sloped / hipped roof patch from the 4 corner heights, raising corners
// that sit fully inside a roofed footprint and nudging the edge corners inward
// by the "strut" test so adjacent patches meet at a ridge.
//
// heights is the flattened terrain-height grid (TerrainHeights) so a roof's
// eaves sit on the wall tops. f supplies the per-wall heights (method428) so a
// roofed corner is lifted onto the surrounding wall tops before the apex raise.
// Returns nil if the window holds no roof tiles.
func BuildRoofs(f *facts.Facts, land *pathfind.Landscape, baseX, baseY, plane int, heights [][]int32) *GameModel {
	if land == nil {
		return nil
	}
	n := terrainSize

	// roofAt returns the 1-based roof id at LOCAL window coords, 0 outside the
	// window (the client's getWallRoof clamps off-window to 0). lx,ly index the
	// roof grid directly (NOT world coords) so the neighbour tests below match
	// World.getWallRoof(x,z).
	roofAt := func(lx, ly int) int32 {
		if lx < 0 || ly < 0 || lx >= n || ly >= n {
			return 0
		}
		return int32(land.Tile(baseX+lx, baseY+ly, plane).RoofTexture)
	}

	// hasRoofTile == method406: the corner at (lx,ly) is fully enclosed by roof
	// (the tile and its 3 neighbours sharing that corner are all roofed), so its
	// height is RAISED to the apex (World.java:391-393).
	hasRoofTile := func(lx, ly int) bool {
		return roofAt(lx, ly) > 0 && roofAt(lx-1, ly) > 0 &&
			roofAt(lx-1, ly-1) > 0 && roofAt(lx, ly-1) > 0
	}

	// strut == method427: a corner is "strutted" (kept on the eave line, NOT
	// nudged) when ANY of the 4 tiles sharing it is roofed (World.java:1441-1443).
	// A corner with NO roofed neighbour on a given side is nudged 16 toward the
	// roof so adjacent patches close up at the ridge/hip.
	strut := func(lx, ly int) bool {
		return roofAt(lx, ly) > 0 || roofAt(lx-1, ly) > 0 ||
			roofAt(lx-1, ly-1) > 0 || roofAt(lx, ly-1) > 0
	}

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

	// rh[lx][ly] is the working roof corner height. It begins at pure terrain
	// height and is then (a) raised onto the wall tops (method428), (b) leveled
	// per roof tile to the tile's max corner (World.java:988-1042), and (c)
	// raised to the apex for fully-enclosed corners (method406). raised[] tracks
	// (c) so a corner shared by several roof tiles is lifted only once.
	raised := make([][]bool, n)
	rh := make([][]int32, n)
	for i := 0; i < n; i++ {
		raised[i] = make([]bool, n)
		rh[i] = make([]int32, n)
		copy(rh[i], heights[i])
	}

	// (a) WALL TOPS (method428, World.java:971-986): each wall edge raises BOTH
	// endpoint corners to the wall top. The client's method428 adds a wall's
	// height to a corner only ONCE (a 0x13880 sentinel locks the corner after the
	// first wall touches it), so a corner shared by several walls is NOT raised
	// repeatedly. We replicate that by recording the MAX wall height touching
	// each corner and adding it once. rh holds the POSITIVE pre-negation height
	// (addV negates it: bigger rh = higher in -Y). Edge axes mirror boundary.go
	// (HorizontalWall = N-S edge (lx,ly)-(lx,ly+1); VerticalWall = E-W edge
	// (lx,ly)-(lx+1,ly); DiagonalWalls = the two diagonal bands).
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
			t := land.Tile(baseX+lx, baseY+ly, plane)
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

	// (b) LEVEL each roof tile's 4 corners to the tile's MAX (highest, largest
	// rh) corner height, so the roof patch sits flat on the highest wall top
	// around it (World.java:988-1042). Done as a settle pass over a snapshot so
	// tile order doesn't matter.
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

	type tri struct {
		v       [4]int // 3 or 4 local-corner vertex slots (see corner mapping)
		nv      int    // 3 = triangle, 4 = quad
		texture int32
	}
	var faces []tri

	// diagAt returns the diagonal-wall band for a tile (mirrors boundary.go /
	// World.getWallDiagonal): 1..11999 = '\' diagonal, 12001..23999 = '/'.
	diagAt := func(lx, ly int) int32 {
		if lx < 0 || ly < 0 || lx >= n || ly >= n {
			return 0
		}
		return land.Tile(baseX+lx, baseY+ly, plane).DiagonalWalls
	}

	any := false
	// First pass: raise every fully-enclosed corner to its apex height. The
	// client does this inline per tile, but because a corner is shared by up to
	// 4 roof tiles we must settle all raises before emitting faces (otherwise a
	// later tile reads a not-yet-raised neighbour corner). For each corner we
	// raise by the def.height of the roof tile that owns it (the tile at the SW
	// of the corner == the iterating tile in the client).
	for lx := 1; lx < n-1; lx++ {
		for ly := 1; ly < n-1; ly++ {
			rt := roofAt(lx, ly)
			if rt == 0 || int(rt) > len(roofDefs) {
				continue
			}
			any = true
			def := roofDefs[rt-1]
			// the 4 corners of this tile, in (lx,ly) order:
			//   A=(lx,ly) B=(lx+1,ly) C=(lx+1,ly+1) D=(lx,ly+1)
			for _, c := range [4][2]int{{lx, ly}, {lx + 1, ly}, {lx + 1, ly + 1}, {lx, ly + 1}} {
				cx, cy := c[0], c[1]
				if hasRoofTile(cx, cy) && !raised[cx][cy] {
					rh[cx][cy] += def.height // bigger rh = higher (addV negates)
					raised[cx][cy] = true
				}
			}
		}
	}

	// Second pass: emit roof faces. We build one *local* GameModel vertex per
	// (corner, tile) on demand because the strut nudge perturbs a corner's X/Z
	// PER TILE (an eave corner is pulled inward only for the tile that owns it).
	g := NewGameModel((n-2)*(n-2)*6, (n-2)*(n-2)*2)

	// addV appends a roof vertex; x/z are the (nudged) plane coords, h the
	// (apex-raised) terrain height. -Y convention: negate the height.
	addV := func(x, z, h int32) int { return g.AddVertex(x, -h, z) }

	for lx := 1; lx < n-1; lx++ {
		for ly := 1; ly < n-1; ly++ {
			rt := roofAt(lx, ly)
			if rt == 0 || int(rt) > len(roofDefs) {
				continue
			}
			def := roofDefs[rt-1]

			// corner heights (apex-raised in pass 1)
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
			// per-corner nudge (client byte0=16): pull an eave corner inward on
			// any side that lacks a strutted (roofed) neighbour, so adjacent
			// patches meet on the ridge. Each corner is tested against its 4
			// orthogonal neighbours exactly as World.java:1094-1126.
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

			// Face topology (World.java:1132-1198). The reference negates the four
			// heights right before these tests (j27..i28 = -height), so the
			// "level" comparisons are on the raised heights directly; we compare
			// our raised heights the same way. Diagonal-wall tiles get a single
			// hip triangle; two-opposite-level corners get a quad; otherwise the
			// tile is split into two triangles along the ridge toward the
			// no-roof neighbour.
			// Corner identities (verified against World.java vertexAt args):
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
				// runs along: split toward the corner WITHOUT a roofed
				// SW/NE neighbour (World.java:1170-1198).
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

	if !any || len(faces) == 0 {
		return nil
	}

	for _, f := range faces {
		verts := make([]int, f.nv)
		copy(verts, f.v[:f.nv])
		// fill=texture on the front, magic on the back so the textured-span path
		// paints the roof tile texture and the underside (back face) is culled.
		g.AddFace(verts, f.texture, magic, 0)
	}
	// Authentic roof light (World.java:1205 setLight(true,50,50,-50,-10,-50)).
	g.SetLight(50, 50, -50, -10, -50)
	return g
}
