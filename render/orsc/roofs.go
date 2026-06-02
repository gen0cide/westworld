package orsc

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// roofs.go ports the ROOF pass of World.generateLandscapeModel (World.java:
// 899-1145) + applyWallToElevationCache (248) + the elevation defs. Roofs sit on
// top of the walls: a wall-height elevation cache is built (terrain + the wall
// heights, tagged with the +80000 "raised" marker), the roof corners are pulled
// up to the surrounding max wall height, then each roof tile's corners are raised
// by the roof's rise and the eave struts extend the perimeter, producing the
// peaked/quad roof faces textured with the roof's texture. Separate model from
// the walls (its own diffuse light, World.java:1147).
//
// MULTI-STORY: the roof pass MUTATES the shared elevation cache (stories.go) and
// is NOT re-seeded from getTileElevation here. The cache is seeded once for the
// ground plane (World.java:820, inside the plane-0-only showWallOnMinimap block)
// and then flows through every story's wall pass + roof pass un-reset — so a
// story-N roof sits on top of the N wall-heights the lower stories baked into the
// cache. That un-reset cache IS OpenRSC's per-floor vertical lift.

// elevationDef is a roof def (EntityHandler.loadElevationDefinitions): rise =
// how far the roof peaks above the walls, tex = the roof texture id. getWallRoof
// returns a 1-based id into this table.
type elevationDef struct{ rise, tex int32 }

// authenticElevationDefs — the 6 roof defs, verbatim from EntityHandler
// loadElevationDefinitions (ElevationDef(unknown1=rise, unknown2=texture)).
var authenticElevationDefs = []elevationDef{
	{64, 6},  // 0
	{64, 3},  // 1
	{96, 2},  // 2
	{80, 33}, // 3
	{80, 15}, // 4
	{90, 49}, // 5
}

const roofRaisedMarker = 80000 // World.java tileElevationCache "already raised" tag

// hostUnderRoof reports the authentic 0x80 "under-roof" condition for the host
// tile (worldX, worldY) on the ground plane: the tile's ground-overlay TYPE == 2
// (the indoor floor type).
//
// AUTHENTIC: World.java:980 sets objectAdjacency[lx][ly] |= 128 exactly when the
// tile's overlay decoration type == 2 (`ClientStream.sharedIntArrayN[deco-1] ==
// 2`), and the render loop (Mudclient.java:8527 / clean client.java:6952)
// hides the host-plane roof + the plane-1/2 walls+roofs whenever
// `this.yj == 0 && (objectAdjacency[currentX/128][currentY/128] & 128) != 0` —
// i.e. the host stands on the ground floor INSIDE a roofed building. Without it
// the host is occluded by his own roof and the interior reads dark/opaque.
//
// We derive the predicate from the SAME overlay table (tileDefs, via typeOf/
// getTileDecorationID) the bit was set from, rather than a hand-copied id list,
// so a future overlay-table correction can't silently desync the cull. The bit
// is only ever set during the plane-0 terrain build and the cull is gated on
// `yj == 0`, so this reads the host tile on plane 0 regardless of `plane`.
//
// hostX/hostY are WORLD coords; baseX/baseY are the window SW corner — the host
// maps to window-local (hostX-baseX, hostY-baseY), which both render callers
// (view.go / harness.go) place at windowCentreTile.
func hostUnderRoof(land *pathfind.Landscape, baseX, baseY, hostX, hostY, plane int) bool {
	if land == nil {
		return false
	}
	// The 0x80 bit is only set during the plane-0 build and the authentic cull
	// is gated on the host being on the ground floor (yj == 0).
	if plane != 0 {
		return false
	}
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: 0}
	id := b.getTileDecorationID(hostX-baseX, hostY-baseY)
	return id > 0 && typeOf(id) == 2
}

func (b *terrainBuilder) getWallRoof(x, z int) int {
	if !b.inWindow(x, z) {
		return 0
	}
	// Java's getWallRoof (World.java:1396) returns the SIGNED byte with NO & 0xff —
	// deliberately asymmetric with getVerticalWall/getHorizontalWall (which DO mask,
	// World.java:1348/1372). A high-bit roof byte (>=128) thus reads negative and is
	// treated as no-roof. (Roof ids are 0..6 in the authentic data, so identical in
	// practice, but kept faithful.)
	return int(int8(b.tile(x, z).RoofTexture))
}

// elevCache is the shared wall-height elevation cache (World.java tileElevationCache,
// new int[96][96]). It is seeded once for the ground plane and then accumulates
// across stories — see stories.go.
type elevCache [worldWindowTiles][worldWindowTiles]int32

// seedElevationCache assigns the cache from this plane's terrain elevation
// (World.java:818-820). Called ONLY for the ground story; upper stories inherit
// the accumulated cache.
func (b *terrainBuilder) seedElevationCache(ec *elevCache) {
	for x := 0; x < worldWindowTiles; x++ {
		for z := 0; z < worldWindowTiles; z++ {
			ec[x][z] = b.getTileElevation(x, z)
		}
	}
}

// BuildRoofs builds a single-story roof model from a freshly-seeded cache. It is
// the single-plane entry kept for tests / standalone use; the live render path
// uses BuildStories (stories.go) so the cache accumulates across floors.
func BuildRoofs(land *pathfind.Landscape, f *facts.Facts, baseX, baseY, plane int) *Model {
	if f == nil {
		return nil
	}
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}
	var ec elevCache
	b.seedElevationCache(&ec)
	model := NewModel(wallModelCapacity, wallModelCapacity)
	b.roofPass(f, model, &ec)
	model.setDiffuseLightAndColor(-50, -10, -50, 50, 50, true, -98)
	return model
}

// roofPass appends this story's roof faces to model and MUTATES the shared cache
// ec in place (applyWall raises it at walls, cleanup levels roof corners, the rise
// peaks roof tiles) — World.java:899-1145, minus the (plane-0-only) cache seed and
// the per-plane setDiffuseLightAndColor (the caller lights the combined model).
func (b *terrainBuilder) roofPass(f *facts.Facts, model *Model, ec *elevCache) {
	// ---- raise the cache at wall tiles (applyWallToElevationCache, World.java:900-916) ----
	applyWall := func(wallID, x1, z1, x2, z2 int) {
		def := f.BoundaryDef(wallID)
		if def == nil {
			return
		}
		h := int32(def.Height)
		if ec[x1][z1] < roofRaisedMarker {
			ec[x1][z1] += h + roofRaisedMarker
		}
		if ec[x2][z2] < roofRaisedMarker {
			ec[x2][z2] += h + roofRaisedMarker
		}
	}
	for x := 0; x < worldWindowTiles-1; x++ {
		for z := 0; z < worldWindowTiles-1; z++ {
			if w := b.getVerticalWall(x, z); w > 0 {
				applyWall(w-1, x, z, x+1, z)
			}
			if w := b.getHorizontalWall(x, z); w > 0 {
				applyWall(w-1, x, z, x, z+1)
			}
			d := int(b.getWallDiagonal(x, z))
			if d > 0 && d < 12000 {
				applyWall(d-1, x, z, x+1, z+1)
			} else if d > 12000 && d < 24000 {
				applyWall(d-12001, x+1, z, x, z+1)
			}
		}
	}

	// ---- roof-corner cleanup: pull each roof tile's corners to the max
	// surrounding wall height (World.java:918-973) ----
	for x := 1; x < worldWindowTiles-1; x++ {
		for z := 1; z < worldWindowTiles-1; z++ {
			if b.getWallRoof(x, z) <= 0 {
				continue
			}
			ec00, ec10, ec11, ec01 := ec[x][z], ec[x+1][z], ec[x+1][z+1], ec[x][z+1]
			if ec00 > roofRaisedMarker {
				ec00 -= roofRaisedMarker
			}
			if ec10 > roofRaisedMarker {
				ec10 -= roofRaisedMarker
			}
			if ec11 > roofRaisedMarker {
				ec11 -= roofRaisedMarker
			}
			if ec01 > roofRaisedMarker {
				ec01 -= roofRaisedMarker
			}
			max := int32(0)
			if ec00 > max {
				max = ec00
			}
			if max < ec10 {
				max = ec10
			}
			if max < ec11 {
				max = ec11
			}
			if max < ec01 {
				max = ec01
			}
			if max >= roofRaisedMarker {
				max -= roofRaisedMarker
			}
			// World.java:953-971 tests the LOCAL ec00..ec01, which were ALREADY
			// stripped of the marker above (lines 933-940) — so a marked corner
			// reads back < 80000 and takes the `= max` branch (leveled + unmarked).
			// Testing the raw stored value instead would wrongly keep stale offsets
			// on marked tiles and leak them into the next story's wall feet.
			setCorner := func(cx, cz int, stripped int32) {
				if stripped < roofRaisedMarker {
					ec[cx][cz] = max
				} else {
					ec[cx][cz] -= roofRaisedMarker
				}
			}
			setCorner(x, z, ec00)
			setCorner(x+1, z, ec10)
			setCorner(x+1, z+1, ec11)
			setCorner(x, z+1, ec01)
		}
	}

	// ---- roof faces (World.java:978-1145) ----
	hasRoofTile := func(x, z int) bool {
		return b.getWallRoof(x, z) > 0 && b.getWallRoof(x-1, z) > 0 &&
			b.getWallRoof(x-1, z-1) > 0 && b.getWallRoof(x, z-1) > 0
	}
	hasRoofStrut := func(x, z int) bool {
		return b.getWallRoof(x, z) <= 0 && b.getWallRoof(x-1, z) <= 0 &&
			b.getWallRoof(x-1, z-1) <= 0 && b.getWallRoof(x, z-1) <= 0
	}
	const eave = 16
	tri := func(ax, ay, az, bx, by, bz, cx, cy, cz, tex int32) {
		va := model.insertVertex(ax, ay, az)
		vb := model.insertVertex(bx, by, bz)
		vc := model.insertVertex(cx, cy, cz)
		if va >= 0 && vb >= 0 && vc >= 0 {
			model.insertFace(3, []int{va, vb, vc}, tex, TRANSPARENT)
		}
	}
	quad := func(ax, ay, az, bx, by, bz, cx, cy, cz, dx, dy, dz, tex int32) {
		va := model.insertVertex(ax, ay, az)
		vb := model.insertVertex(bx, by, bz)
		vc := model.insertVertex(cx, cy, cz)
		vd := model.insertVertex(dx, dy, dz)
		if va >= 0 && vb >= 0 && vc >= 0 && vd >= 0 {
			model.insertFace(4, []int{va, vb, vc, vd}, tex, TRANSPARENT)
		}
	}

	for x := 1; x < worldWindowTiles-1; x++ {
		for z := 1; z < worldWindowTiles-1; z++ {
			roofID := b.getWallRoof(x, z)
			if roofID <= 0 || roofID-1 >= len(authenticElevationDefs) {
				continue
			}
			x10, x11, z11, z01 := x+1, x+1, z+1, z+1
			p00x, p00z := int32(x*128), int32(z*128)
			p10x, p11z := 128+p00x, 128+p00z
			p01x, p10z := p00x, p00z
			p11x, p01z := p10x, p11z

			ec00, ec10, ec11, ec01 := ec[x][z], ec[x10][z], ec[x11][z11], ec[x][z01]
			rise := authenticElevationDefs[roofID-1].rise

			if hasRoofTile(x, z) && ec00 < roofRaisedMarker {
				ec00 += rise + roofRaisedMarker
				ec[x][z] = ec00
			}
			if hasRoofTile(x10, z) && ec10 < roofRaisedMarker {
				ec10 += rise + roofRaisedMarker
				ec[x10][z] = ec10
			}
			if hasRoofTile(x11, z11) && ec11 < roofRaisedMarker {
				ec11 += roofRaisedMarker + rise
				ec[x11][z11] = ec11
			}
			if ec10 >= roofRaisedMarker {
				ec10 -= roofRaisedMarker
			}
			if ec11 >= roofRaisedMarker {
				ec11 -= roofRaisedMarker
			}
			if hasRoofTile(x, z01) && ec01 < roofRaisedMarker {
				ec01 += rise + roofRaisedMarker
				ec[x][z01] = ec01
			}
			if ec00 >= roofRaisedMarker {
				ec00 -= roofRaisedMarker
			}
			if ec01 >= roofRaisedMarker {
				ec01 -= roofRaisedMarker
			}

			// eave struts extend the perimeter corners outward (World.java:1035-1069)
			if hasRoofStrut(x-1, z) {
				p00x -= eave
			}
			if hasRoofStrut(x+1, z) {
				p00x += eave
			}
			if hasRoofStrut(x, z-1) {
				p00z -= eave
			}
			if hasRoofStrut(x, z+1) {
				p00z += eave
			}
			if hasRoofStrut(x10-1, z) {
				p10x -= eave
			}
			if hasRoofStrut(x10+1, z) {
				p10x += eave
			}
			if hasRoofStrut(x10, z-1) {
				p10z -= eave
			}
			if hasRoofStrut(x10, z+1) {
				p10z += eave
			}
			if hasRoofStrut(x11-1, z11) {
				p11x -= eave
			}
			if hasRoofStrut(x11+1, z11) {
				p11x += eave
			}
			if hasRoofStrut(x11, z11-1) {
				p11z -= eave
			}
			if hasRoofStrut(x11, z11+1) {
				p11z += eave
			}
			if hasRoofStrut(x-1, z01) {
				p01x -= eave
			}
			if hasRoofStrut(x+1, z01) {
				p01x += eave
			}
			if hasRoofStrut(x, z01-1) {
				p01z -= eave
			}
			if hasRoofStrut(x, z01+1) {
				p01z += eave
			}

			tex := authenticElevationDefs[roofID-1].tex
			// Y is -elevation (World.java:1072-1075 negate the cache values).
			ec00, ec10, ec11, ec01 = -ec00, -ec10, -ec11, -ec01
			diag := int(b.getWallDiagonal(x, z))

			switch {
			case diag > 12000 && diag < 24000 && b.getWallRoof(x-1, z-1) == 0:
				tri(p11x, ec11, p11z, p01x, ec01, p01z, p10x, ec10, p10z, tex)
			case diag > 12000 && diag < 24000 && b.getWallRoof(x+1, z+1) == 0:
				tri(p00x, ec00, p00z, p10x, ec10, p10z, p01x, ec01, p01z, tex)
			case diag > 0 && diag < 12000 && b.getWallRoof(x+1, z-1) == 0:
				tri(p01x, ec01, p01z, p00x, ec00, p00z, p11x, ec11, p11z, tex)
			case diag > 0 && diag < 12000 && b.getWallRoof(x-1, z+1) == 0:
				tri(p10x, ec10, p10z, p11x, ec11, p11z, p00x, ec00, p00z, tex)
			case ec10 == ec00 && ec11 == ec01:
				quad(p00x, ec00, p00z, p10x, ec10, p10z, p11x, ec11, p11z, p01x, ec01, p01z, tex)
			case ec00 == ec01 && ec11 == ec10:
				quad(p01x, ec01, p01z, p00x, ec00, p00z, p10x, ec10, p10z, p11x, ec11, p11z, tex)
			default:
				split := b.getWallRoof(x-1, z-1) > 0 || b.getWallRoof(x+1, z+1) > 0
				if split {
					tri(p10x, ec10, p10z, p11x, ec11, p11z, p00x, ec00, p00z, tex)
					tri(p01x, ec01, p01z, p00x, ec00, p00z, p11x, ec11, p11z, tex)
				} else {
					tri(p00x, ec00, p00z, p10x, ec10, p10z, p01x, ec01, p01z, tex)
					tri(p11x, ec11, p11z, p01x, ec01, p01z, p10x, ec10, p10z, tex)
				}
			}
		}
	}

	// ---- final full-grid marker strip (World.java:1155-1158) ----
	// The roof-rise loop above STORES marked values (value + rise + 80000) back
	// into the cache and only un-marks the LOCAL copies it used for vertex Y; the
	// applyWall pass also leaves +80000 on wall tiles no roof corner cleaned. Java
	// wipes EVERY leftover marker across the whole 96x96 grid at the end of each
	// story (this loop is outside the plane-0-only showWallOnMinimap block, so it
	// runs for every story) — so the NEXT story's wall pass reads clean accumulated
	// heights (insertWallIntoModel reads the cache raw). Without this, upper-story
	// wall feet land on marked tiles and launch to Y~-80000 (the tall slivers /
	// collapsed upper floors). This is THE multi-story lift's closing step.
	for x := 0; x < worldWindowTiles; x++ {
		for z := 0; z < worldWindowTiles; z++ {
			if ec[x][z] >= roofRaisedMarker {
				ec[x][z] -= roofRaisedMarker
			}
		}
	}
}
