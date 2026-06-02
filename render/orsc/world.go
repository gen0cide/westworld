package orsc

// world.go — faithful port of OpenRSC three/ World.java's TERRAIN + SCENERY mesh
// build. Source of truth:
//
//	/Users/flint/Code/openrsc/Client_Base/src/orsc/graphics/three/World.java
//
// World.java reads its landscape out of a 96x96 window assembled from four 48x48
// Sectors (loadSection, World.java:1906-1928) and emits the terrain/wall/roof
// geometry into an accumulator RSModel (modelAccumulate, World.java:519) via
// insertVertex/insertFace, then divideModelByGrid splits that into render buckets
// that Scene.addModel consumes (generateLandscapeModel, World.java:507-1164).
//
// This file ports the TERRAIN portion of generateLandscapeModel: the per-vertex
// elevation pass + tileType-4 flatten (World.java:531-552), the per-tile
// colour/overlay diagonal split (World.java:554-702), the tileType-4 raised
// bridge-DECK second pass and the deck OVERHANG passes (World.java:704-808), plus
// getElevation / getTileElevation (World.java:1173-1304), the 250 bridge-marker
// remap (setTileDecorationOnBridge, World.java:1632-1651), and the scenery
// placement math (addLoginScreenModels, World.java:182-189: orient(0,dir*32,0),
// footprint centre, getElevation anchor). The WALL + ROOF passes
// (World.java:824-1158) live in a separate port file.
//
// Data plumbing: where World.java reads sectors[chunk].getTile(x,z).field this
// port reads pathfind.Landscape.Tile(baseX+x, baseY+z, plane).Field — the SAME
// .orsc landscape the existing renderer uses. The 96x96 chunk arithmetic
// (getTileElevation et al., World.java:1280-1403) is subsumed by land.Tile's own
// world->sector mapping; tile (x,z) in [0,96) maps to world tile
// (baseX+x, baseY+z). The Model build uses model.go's RSModel API (NewModel,
// insertVertex, insertFace) so the geometry is a normal RSModel that Scene.Render
// (the Scene port) projects + rasterises. BuildTerrain returns the single
// accumulator *Model; the harness (harness.go) does the Scene.AddModel.
//
// EntityHandler.getTileDef(id-1) (the GroundOverlay definition table) is ported
// as the tileDefs LUT below (getTileValue == tileType, getColour == colour);
// EntityHandler.getElevationDef / getDoorDef (roof + wall defs) belong to the
// wall/roof port and are not referenced here.

import (
	"os"

	"github.com/gen0cide/westworld/pathfind"
)

// =============================================================================
// colorToResource[256] — World ctor LUT (World.java:20,60-74). Indexed by
// GroundTexture (getTerrainColour); each entry is a GenUtil.colorToResource
// (GenUtil.java:36-45) 5:5:5 resource fill for the underlay grass-blend palette.
// Package-level + init()-filled (the World instance LUT is a per-Scene field in
// Java; here a single immutable table since the ramps are constant).
// =============================================================================

var colorToResource [256]int32

// genColorToResource is GenUtil.colorToResource (GenUtil.java:36-45):
// b>>=3; r>>=3; g>>=3; return -(g<<5) - 1 - (r<<10) - b. Int is 32-bit in Java.
func genColorToResource(r, g, b int32) int32 {
	b >>= 3
	r >>= 3
	g >>= 3
	return -(g << 5) - 1 - (r << 10) - b
}

func init() {
	// World.java:60-74 — four 64-entry ramps. The (int)((double)v*1.75) etc. casts
	// are integer truncations (Go int32 truncation matches Java's (int) cast).
	for v := int32(0); v < 64; v++ {
		colorToResource[v] = genColorToResource(255-v*4, 255-int32(float64(v)*1.75), 255-v*4)
	}
	for v := int32(0); v < 64; v++ {
		colorToResource[64+v] = genColorToResource(v*3, 144, 0)
	}
	for v := int32(0); v < 64; v++ {
		colorToResource[128+v] = genColorToResource(192-int32(float64(v)*1.5), 144-int32(float64(v)*1.5), 0)
	}
	for v := int32(0); v < 64; v++ {
		colorToResource[192+v] = genColorToResource(96-int32(float64(v)*1.5), int32(float64(v)*1.5)+48, 0)
	}
}

// =============================================================================
// tileDef — EntityHandler.getTileDef(id-1) (referenced at World.java:534,556,
// 568,573,...). getTileValue() == tileType, getColour() == colour. The landscape
// stores GroundOverlay ids 1..N; getTileDecorationID returns the id (0 = none)
// and the def is looked up at id-1. colour follows the face-fill convention: >=0
// is a texture id, <0 is a colorToResource flat fill, TRANSPARENT (12345678) is
// the bridge sentinel. Ported from the authentic TileDef table.
// =============================================================================

type tileDef struct {
	colour   int32 // getColour()
	tileType int   // getTileValue()
}

// tileDefs[id-1] is the GroundOverlay definition for decoration id (1..len).
var tileDefs = []tileDef{
	{-16913, 1},      // 1  road / path (flat)
	{1, 3},           // 2  water-edge texture
	{3, 2},           // 3  wood planks floor (chapel) -> tex 3
	{3, 4},           // 4  water-class (raised bridge deck; planks colour, type 4)
	{-16913, 2},      // 5  indoor flat
	{-27685, 2},      // 6  indoor flat (reddish)
	{25, 3},          // 7  gungywater texture
	{TRANSPARENT, 5}, // 8  bridge sentinel
	{-26426, 1},      // 9  flat (bridge plank edge, setTileDecoration 9 World.java:1641)
	{-1, 5},          // 10 bridge
	{31, 3},          // 11 lava texture
	{3, 4},           // 12 water-class
	{-4534, 2},       // 13 indoor flat
	{32, 2},          // 14 pentagram -> tex 32
	{-9225, 2},       // 15 indoor flat
	{-3172, 2},       // 16 indoor flat
	{15, 2},          // 17 marble floor -> tex 15
	{-2, 2},          // 18 indoor flat
	{-1, 3},          // 19 outdoor flat special
	{-2, 4},          // 20 water-class
	{-2, 4},          // 21 water-class
	{-2, 0},          // 22 flat
	{-17793, 2},      // 23 dirt / planter (flat)
	{-14594, 1},      // 24 flat
	{1, 3},           // 25 water-edge texture
}

// tileDefAt returns the def for decoration id N (the getTileDef(N-1) call) and
// ok=false for id 0 / out-of-range. Callers pass the RAW 1-based id.
func tileDefAt(id int) (tileDef, bool) {
	if id >= 1 && id <= len(tileDefs) {
		return tileDefs[id-1], true
	}
	return tileDef{}, false
}

// typeOf is shorthand for tileDefAt(id).tileType (0 when id is unknown).
func typeOf(id int) int {
	if def, ok := tileDefAt(id); ok {
		return def.tileType
	}
	return 0
}

// =============================================================================
// terrainBuilder — the per-build window state. Fields {land, baseX, baseY, plane}
// match the harness's construction (harness.go elevationOf); the overlay-remap
// cache is added for setTileDecorationOnBridge (we never mutate the archive). The
// World.java tile accessors (instance methods reading sectors[]) become methods
// on this builder; where World.java indexes sectors[chunk].getTile(localX,localZ)
// after the 96x96->4x48x48 chunk split, this reads land.Tile(baseX+x, baseY+z,
// plane) (the world->sector mapping lives in pathfind.Landscape). The 96x96 bound
// check (World.java:1283 etc.) is preserved so neighbour probes at x=-1 / x=96
// return 0 exactly as World.java does.
// =============================================================================

const windowTiles = 96

type terrainBuilder struct {
	land         *pathfind.Landscape
	baseX, baseY int
	plane        int

	// overlay/overlaySet hold the setTileDecorationOnBridge 250 remap override.
	overlay    [windowTiles][windowTiles]int
	overlaySet [windowTiles][windowTiles]bool
}

func (b *terrainBuilder) inWindow(x, z int) bool {
	return x >= 0 && x < windowTiles && z >= 0 && z < windowTiles
}

func (b *terrainBuilder) tile(x, z int) pathfind.Tile {
	return b.land.Tile(b.baseX+x, b.baseY+z, b.plane)
}

// getTileElevation — World.java:1280-1304: (groundElevation & 0xff) * 3.
func (b *terrainBuilder) getTileElevation(x, z int) int32 {
	if !b.inWindow(x, z) {
		return 0
	}
	return int32(int(b.tile(x, z).GroundElevation)&0xff) * 3
}

// getTerrainColour — World.java:1204-1226: groundTexture & 0xff.
func (b *terrainBuilder) getTerrainColour(x, z int) int {
	if !b.inWindow(x, z) {
		return 0
	}
	return int(b.tile(x, z).GroundTexture) & 0xff
}

// getTileDecorationID — World.java:1242-1266: groundOverlay & 0xff, after the
// setTileDecorationOnBridge 250 remap. Reads the override cache when present so
// the bridge remap is observed by every probe.
func (b *terrainBuilder) getTileDecorationID(x, z int) int {
	if !b.inWindow(x, z) {
		return 0
	}
	if b.overlaySet[x][z] {
		return b.overlay[x][z] & 0xff
	}
	return int(b.tile(x, z).GroundOverlay) & 0xff
}

// setTileDecoration — World.java:1608-1630: write groundOverlay (here, the
// override cache, since we never mutate the landscape archive).
func (b *terrainBuilder) setTileDecoration(x, z, val int) {
	if !b.inWindow(x, z) {
		return
	}
	b.overlay[x][z] = val
	b.overlaySet[x][z] = true
}

// getWallDiagonal — World.java:1306-1330: diagonalWalls (raw int32; the 1..11999
// / 12001..23999 / 48001..59999 ranges drive the colour-split branches).
func (b *terrainBuilder) getWallDiagonal(x, z int) int32 {
	if !b.inWindow(x, z) {
		return 0
	}
	return b.tile(x, z).DiagonalWalls
}

// getTileDecorationCacheVal — World.java:1228-1240: the overlay tile's getColour
// at (x,z), or defaultVal when the tile has no overlay (id 0).
func (b *terrainBuilder) getTileDecorationCacheVal(x, z int, defaultVal int32) int32 {
	id := b.getTileDecorationID(x, z)
	if id == 0 {
		return defaultVal
	}
	if def, ok := tileDefAt(id); ok {
		return def.colour
	}
	return defaultVal
}

// isTileType2 — World.java:1457-1474: -1 when no overlay, 1 when the overlay is
// tileType 2 (indoor floor), else 0. The colour-split compares this CLASS, not
// the raw overlay id.
func (b *terrainBuilder) isTileType2(x, z int) int {
	id := b.getTileDecorationID(x, z)
	if id == 0 {
		return -1
	}
	if def, ok := tileDefAt(id); ok {
		if def.tileType != 2 {
			return 0
		}
		return 1
	}
	return 0
}

// setTileDecorationOnBridge — World.java:1632-1651. The on-disk overlay 250 is
// the BRIDGE marker; before building terrain the client rewrites every 250 tile.
// A chunk-seam tile at x==47 whose +x neighbour is neither bridge(250) nor
// water-class(2) -> overlay 9 (a brown plank edge); the z==47 seam likewise;
// every other 250 deck tile -> overlay 2 (the deck flattens + textures with the
// SAME water it spans). Plane 0 only (the caller gates on plane==0).
func (b *terrainBuilder) setTileDecorationOnBridge() {
	for x := 0; x < windowTiles; x++ {
		for z := 0; z < windowTiles; z++ {
			if b.getTileDecorationID(x, z) == 250 {
				switch {
				case x == 47 && b.getTileDecorationID(x+1, z) != 250 && b.getTileDecorationID(x+1, z) != 2:
					b.setTileDecoration(x, z, 9)
				case z == 47 && b.getTileDecorationID(x, z+1) != 250 && b.getTileDecorationID(x, z+1) != 2:
					b.setTileDecoration(x, z, 9)
				default:
					b.setTileDecoration(x, z, 2)
				}
			}
		}
	}
}

// =============================================================================
// getElevation — World.java:1173-1202. Bilinear-interpolated terrain height at a
// FINE WINDOW-LOCAL point (x,z in 128-units-per-tile). Scenery anchors its base
// here; the harness camera reads it for the look-at height. The 128-zLerp / xLerp
// diagonal split picks which of the tile's two triangles the point falls in.
// =============================================================================

func (b *terrainBuilder) getElevation(x, z int32) int32 {
	xTile := int(x >> 7)
	zTile := int(z >> 7)
	xLerp := x & 127
	zLerp := z & 127
	if xTile >= 0 && zTile >= 0 && xTile < 95 && zTile < 95 {
		var tileCorner, dEX, dEZ int32
		if xLerp <= 128-zLerp {
			tileCorner = b.getTileElevation(xTile, zTile)
			dEX = b.getTileElevation(1+xTile, zTile) - tileCorner
			dEZ = b.getTileElevation(xTile, 1+zTile) - tileCorner
		} else {
			tileCorner = b.getTileElevation(1+xTile, zTile+1)
			dEX = b.getTileElevation(xTile, zTile+1) - tileCorner
			dEZ = b.getTileElevation(1+xTile, zTile) - tileCorner
			xLerp = 128 - xLerp
			zLerp = 128 - zLerp
		}
		return dEZ*zLerp/128 + tileCorner + dEX*xLerp/128
	}
	return 0
}

// =============================================================================
// BuildTerrain — the TERRAIN half of World.generateLandscapeModel
// (World.java:507-816). Builds + lights the accumulator *Model for the 96x96
// window whose SW corner is world tile (baseX, baseY) at the given plane, and
// returns it (the harness does Scene.AddModel). The divideModelByGrid bucketing
// (World.java:812-816) is a render-culling concern; this port returns the single
// accumulator model.
// =============================================================================

func BuildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) *Model {
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}

	// setTileDecorationOnBridge — World.java:517,1491 runs the 250 remap before
	// the build, plane 0 only. Upper planes force overlays TRANSPARENT below.
	if plane == 0 {
		b.setTileDecorationOnBridge()
	}

	// modelAccumulate — World.java:519 new RSModel(18688, 18688, ...). The 96x96
	// window emits 96*96 underlay verts + up to 4 per deck/overhang quad and up to
	// 2 faces per tile; the authentic 18688 cap is preserved.
	const accumCap = 18688
	worldMod := NewModel(accumCap, accumCap)
	worldMod.resetFaceVertHead() // World.java:529 worldMod.resetFaceVertHead(1)

	// ---- Vertex pass — World.java:531-552 ----
	// One vertex per 96x96 grid corner at (x*128, y, z*128) where y =
	// -getTileElevation(x,z), FLATTENED to 0 when this corner touches any
	// tileType-4 (water-class) decoration tile. The four getTileDecorationID
	// probes are (x,z),(x-1,z),(x,z-1),(x-1,z-1) — the four tiles sharing this
	// corner. insertVertex appends in this exact order, so vertex id == z + x*96
	// and the face index math below is direct.
	for x := 0; x < windowTiles; x++ {
		for z := 0; z < windowTiles; z++ {
			y := -b.getTileElevation(x, z)
			if id := b.getTileDecorationID(x, z); id > 0 && typeOf(id) == 4 {
				y = 0
			}
			if id := b.getTileDecorationID(x-1, z); id > 0 && typeOf(id) == 4 {
				y = 0
			}
			if id := b.getTileDecorationID(x, z-1); id > 0 && typeOf(id) == 4 {
				y = 0
			}
			if id := b.getTileDecorationID(x-1, z-1); id > 0 && typeOf(id) == 4 {
				y = 0
			}

			vID := worldMod.insertVertex(int32(x)*128, y, int32(z)*128)
			// World.java:550-551 per-vertex ambience speckle
			// (int)(Math.random()*10)-5. Deterministic (cacheable) via a hash of
			// the WORLD-tile coord; range [-5,4] matches the authentic spread.
			worldMod.setVertexLightOther(vID, terrainAmbience(baseX+x, baseY+z))
		}
	}

	// ---- Colour / overlay diagonal split — World.java:554-702 ----
	for x := 0; x < windowTiles-1; x++ {
		for z := 0; z < windowTiles-1; z++ {
			// Underlay grass-blend colour from getTerrainColour (World.java:556).
			// Planes 1/2 force every tile TRANSPARENT (World.java:559-563) so
			// upper-floor terrain is invisible.
			colorResource := colorToResource[b.getTerrainColour(x, z)]
			res01 := colorResource
			defaultVal := colorResource
			if plane == 1 || plane == 2 {
				colorResource = TRANSPARENT
				res01 = TRANSPARENT
				defaultVal = TRANSPARENT
			}

			bridge0011 := 0 // World.java:565 byte bridge00_11

			// Overlay tile: pick the overlay colour + which diagonal carries the
			// seam (World.java:566-637).
			if decorID := b.getTileDecorationID(x, z); decorID > 0 {
				def, _ := tileDefAt(decorID)
				decorType := def.tileType
				decorType2 := b.isTileType2(x, z)
				colorResource = def.colour
				res01 = def.colour

				// World.java:574-581 — tileType 4 forces colour 1 (water), except
				// id 12 forces 31 (lava).
				if decorType == 4 {
					colorResource = 1
					res01 = 1
					if decorID == 12 {
						colorResource = 31
						res01 = 31
					}
				}

				if decorType == 5 {
					// World.java:583-609 — bridge sentinel: pick the seam from the
					// neighbour deck colours (only when a diagonal wall is present).
					if d := b.getWallDiagonal(x, z); d > 0 && d < 24000 {
						switch {
						case b.getTileDecorationCacheVal(x-1, z, defaultVal) != TRANSPARENT &&
							b.getTileDecorationCacheVal(x, z-1, defaultVal) != TRANSPARENT:
							bridge0011 = 0
							colorResource = b.getTileDecorationCacheVal(x-1, z, defaultVal)
						case b.getTileDecorationCacheVal(x+1, z, defaultVal) != TRANSPARENT &&
							b.getTileDecorationCacheVal(x, z+1, defaultVal) != TRANSPARENT:
							res01 = b.getTileDecorationCacheVal(x+1, z, defaultVal)
							bridge0011 = 0
						case b.getTileDecorationCacheVal(x+1, z, defaultVal) != TRANSPARENT &&
							b.getTileDecorationCacheVal(x, z-1, defaultVal) != TRANSPARENT:
							res01 = b.getTileDecorationCacheVal(x+1, z, defaultVal)
							bridge0011 = 1
						case b.getTileDecorationCacheVal(x-1, z, defaultVal) != TRANSPARENT &&
							b.getTileDecorationCacheVal(x, z+1, defaultVal) != TRANSPARENT:
							bridge0011 = 1
							colorResource = b.getTileDecorationCacheVal(x-1, z, defaultVal)
						}
					}
				} else if decorType != 2 || (b.getWallDiagonal(x, z) > 0 && b.getWallDiagonal(x, z) < 24000) {
					// World.java:610-628 — overlay-vs-different-class seam: revert
					// one triangle half (colorResource/res01) to the underlay along
					// the diagonal that faces a tile of a DIFFERENT class.
					switch {
					case decorType2 != b.isTileType2(x-1, z) && b.isTileType2(x, z-1) != decorType2:
						colorResource = defaultVal
						bridge0011 = 0
					case decorType2 != b.isTileType2(x+1, z) && b.isTileType2(x, z+1) != decorType2:
						bridge0011 = 0
						res01 = defaultVal
					case decorType2 != b.isTileType2(x+1, z) && b.isTileType2(x, z-1) != decorType2:
						res01 = defaultVal
						bridge0011 = 1
					case decorType2 != b.isTileType2(x-1, z) && decorType2 != b.isTileType2(x, z+1):
						colorResource = defaultVal
						bridge0011 = 1
					}
				}
				// World.java:630-636 objectType/collision flags are pathfind
				// concerns (no mesh effect) and are omitted here.
			}

			// Tile slope: a non-planar tile must ALSO split so each triangle's
			// gouraud normal is exact (World.java:640-641).
			slope := b.getTileElevation(x+1, 1+z) - b.getTileElevation(x, z) +
				b.getTileElevation(x, z+1) - b.getTileElevation(x+1, z)

			base := z + x*windowTiles // vertex id of corner (x,z)

			if colorResource == res01 && slope == 0 {
				// World.java:643-652 — flat planar tile of one colour: a single
				// quad. Indices {96+base, base, 1+base, 97+base} (World.java:645-646).
				if colorResource != TRANSPARENT {
					faceIndicies := []int{base + 96, base, base + 1, base + 97}
					faceID := worldMod.insertFace(4, faceIndicies, TRANSPARENT, colorResource)
					worldMod.facePickIndex[faceID] = int32(faceID) + 200000
				}
			} else if bridge0011 == 0 {
				// World.java:656-677 — split along the / diagonal.
				if colorResource != TRANSPARENT {
					// [0]=96+base, [1]=base, [2]=1+base (World.java:658-660).
					faceIndicies := []int{base + 96, base, base + 1}
					faceID := worldMod.insertFace(3, faceIndicies, TRANSPARENT, colorResource)
					worldMod.facePickIndex[faceID] = int32(faceID) + 200000
				}
				if res01 != TRANSPARENT {
					// [0]=1+base, [1]=97+base, [2]=base+96 (World.java:669-671).
					faceIndices2 := []int{base + 1, base + 97, base + 96}
					faceID := worldMod.insertFace(3, faceIndices2, TRANSPARENT, res01)
					worldMod.facePickIndex[faceID] = int32(faceID) + 200000
				}
			} else {
				// World.java:678-700 — split along the \ diagonal.
				if colorResource != TRANSPARENT {
					// [0]=1+base, [1]=97+base, [2]=base (World.java:680-682).
					faceIndicies := []int{base + 1, base + 97, base}
					faceID := worldMod.insertFace(3, faceIndicies, TRANSPARENT, colorResource)
					worldMod.facePickIndex[faceID] = 200000 + int32(faceID)
				}
				if res01 != TRANSPARENT {
					// [0]=base+96, [1]=base, [2]=base+97 (World.java:691-693).
					faceIndices2 := []int{base + 96, base, base + 97}
					faceID := worldMod.insertFace(3, faceIndices2, TRANSPARENT, res01)
					worldMod.facePickIndex[faceID] = int32(faceID) + 200000
				}
			}
		}
	}

	// ---- tileType-4 raised deck + overhang second passes — World.java:704-808 ----
	// Each tileType-4 tile draws a SECOND quad at its REAL (un-flattened)
	// elevation painted with the tile's deck colour (the raised plank deck over
	// the flattened water from pass 1, World.java:706-723). A tile that is NOT a
	// tileType-3 overlay and borders a tileType-4 tile draws the deck overhang
	// onto the bank (World.java:724-808), one quad per tileType-4 neighbour
	// (+z, -z, +x, -x).
	for x := 1; x < windowTiles-1; x++ {
		for z := 1; z < windowTiles-1; z++ {
			if id := b.getTileDecorationID(x, z); id > 0 && typeOf(id) == 4 {
				def, _ := tileDefAt(id)
				b.emitDeckQuad(worldMod, x, z, def.colour)
			} else if id == 0 || typeOf(id) != 3 {
				if nz := b.getTileDecorationID(x, z+1); nz > 0 && typeOf(nz) == 4 {
					def, _ := tileDefAt(nz)
					b.emitDeckQuad(worldMod, x, z, def.colour)
				}
				if nz := b.getTileDecorationID(x, z-1); nz > 0 && typeOf(nz) == 4 {
					def, _ := tileDefAt(nz)
					b.emitDeckQuad(worldMod, x, z, def.colour)
				}
				if nx := b.getTileDecorationID(x+1, z); nx > 0 && typeOf(nx) == 4 {
					def, _ := tileDefAt(nx)
					b.emitDeckQuad(worldMod, x, z, def.colour)
				}
				if nx := b.getTileDecorationID(x-1, z); nx > 0 && typeOf(nx) == 4 {
					def, _ := tileDefAt(nx)
					b.emitDeckQuad(worldMod, x, z, def.colour)
				}
			}
		}
	}

	// World.java:811 setDiffuseLightAndColor(-50,-10,-50, 40, 48, true, 105).
	worldMod.setDiffuseLightAndColor(-50, -10, -50, 40, 48, true, 105)
	return worldMod
}

// emitDeckQuad inserts the raised bridge-deck quad for tile (x,z) at the tile's
// REAL elevation, painted with tileDecor in the FRONT slot (World.java:711-723).
// The four corner verts are inserted at the real (un-flattened) getTileElevation
// — distinct from the pass-1 flattened underlay verts, so the deck sits ABOVE the
// water with no z-fight.
func (b *terrainBuilder) emitDeckQuad(m *Model, x, z int, tileDecor int32) {
	v00 := m.insertVertex(int32(x)*128, -b.getTileElevation(x, z), int32(z)*128)
	v10 := m.insertVertex(int32(x+1)*128, -b.getTileElevation(x+1, z), int32(z)*128)
	v11 := m.insertVertex(int32(x+1)*128, -b.getTileElevation(x+1, z+1), int32(z+1)*128)
	v01 := m.insertVertex(int32(x)*128, -b.getTileElevation(x, z+1), int32(z+1)*128)
	indices := []int{v00, v10, v11, v01}
	// World.java:719 insertFace(4, indices, tileDecor, TRANSPARENT, false) — deck
	// colour in the FRONT slot, TRANSPARENT back.
	faceID := m.insertFace(4, indices, tileDecor, TRANSPARENT)
	m.facePickIndex[faceID] = int32(faceID) + 200000
}

// =============================================================================
// RSModel methods used only by the World build (the model.go RSModel port did not
// provide these; they live here because nothing else references them).
// =============================================================================

// setDiffuseLightAndColor — RSModel.setDiffuseLightAndColor (RSModel.java:1199-
// 1223), called by the World build (World.java:811). diffuseParam2 = (64-p2)*16
// + 128; diffuseParam1 = 256 - p1*4; gouraud (var5=true) sets every face's
// faceDiffuseLight to TRANSPARENT (the per-vertex sentinel m_Vb); diffuseDir +
// diffuseMag from (dirX,dirY,dirZ); then computeDiffuse runs the shading.
func (m *Model) setDiffuseLightAndColor(dirX, dirY, dirZ, p1, p2 int32, var5 bool, var7 int32) {
	m.diffuseParam2 = (64-p2)*16 + 128
	m.diffuseParam1 = 256 - p1*4

	if !m.dontComputeDiffuse {
		for i := 0; i < m.faceHead; i++ {
			if var5 {
				m.faceDiffuseLight[i] = TRANSPARENT // RSModel.m_Vb, gouraud sentinel
			} else {
				m.faceDiffuseLight[i] = 0
			}
		}
		m.diffuseDirX = dirX
		m.diffuseDirZ = dirZ
		m.diffuseDirY = dirY
		m.diffuseMag = isqrt(dirY*dirY + dirX*dirX + dirZ*dirZ)
		m.computeDiffuse(-121) // RSModel.java:1217
	}
}

// resetFaceVertHead — RSModel.resetFaceVertHead (RSModel.java:944-952): zero the
// live face + vertex counts (the accumulator is re-built per pass,
// World.java:529,822,897).
func (m *Model) resetFaceVertHead() {
	m.faceHead = 0
	m.vertHead = 0
}

// setVertexLightOther — RSModel.setVertexLightOther (RSModel.java:1333-1340):
// store a per-vertex ambience speckle as a signed byte (World.java:551).
func (m *Model) setVertexLightOther(id, val int) {
	m.vertLightOther[id] = int8(val)
}

// isqrt is the integer floor of sqrt for the diffuse magnitude (RSModel.java:1216
// (int)Math.sqrt) — Newton's method, no float path, so the result is the exact
// (int) truncation Java produces for these small magnitudes.
func isqrt(v int32) int32 {
	if v <= 0 {
		return 0
	}
	x := v
	y := (x + 1) / 2
	for y < x {
		x = y
		y = (x + v/x) / 2
	}
	return x
}

// terrainAmbience returns a stable pseudo-random per-vertex ambience offset in
// [-5,4], the deterministic stand-in for World.java:550 (int)(Math.random()*10)-5.
// Derived from a cheap integer hash of the WORLD-tile coord so renders are
// reproducible (cacheable) rather than time-seeded; the [-5,4] range matches the
// authentic distribution.
//
// FLAT-AMBIENCE FLAG: the authentic client seeds this from Math.random(), so the
// per-vertex ±5 speckle is genuinely non-deterministic and CANNOT pixel-match the
// DEOB/JAR oracle exactly (the known deferred ambience-speckle). Set
// ORSC_FLAT_AMBIENCE=1 to force a flat 0 offset; the render-diff DEOB leg pins its
// own Math.random()->0.5 (=> amb 0) the same way, so the two then pixel-match 1:1
// (modulo nothing). Live renders leave the env unset and keep the authentic speckle.
func terrainAmbience(x, y int) int {
	if flatAmbience {
		return 0
	}
	h := uint32(x)*0x9e3779b1 + uint32(y)*0x85ebca77
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	return int(h%10) - 5
}

// flatAmbience pins the per-vertex terrain ambience speckle to 0 when
// ORSC_FLAT_AMBIENCE is set (for the exact orsc-vs-DEOB pixel compare). Read once
// at init so the hot terrainAmbience path stays branch-cheap.
var flatAmbience = os.Getenv("ORSC_FLAT_AMBIENCE") != ""

// =============================================================================
// Scenery placement — World.addLoginScreenModels (World.java:182-189), the lines
// that orient + translate a copied scenery model onto a tile:
//
//	int xTranslate = (xSize + x + x) * 128 / 2;     // footprint-centre X
//	int zTranslate = (zSize + z + z) * 128 / 2;     // footprint-centre Z
//	copy.translate2(xTranslate, -getElevation(xTranslate, zTranslate), zTranslate);
//	copy.setRot256(0, getTileDirection(x, z) * 32, 0);   // heading -> Y rot
//
// where xSize/zSize swap for a heading other than 0/4 (World.java:173-179, the
// object rotated onto its side).
// =============================================================================

// sceneryLiftY returns the per-id vertical lift (in world units, -Y is up) that
// the deob applies to a scenery model AFTER the ground anchor but before the
// translate. Today the only lifted object is id 74, the windmill sails, which
// float 480 units UP so the sail assembly sits atop the mill tower rather than
// splayed flat at terrain level (scenery-id74-no-vertical-lift).
//
// Ground truth: deob Mudclient.java:6229-6231 & 4296-4298
//
//	if (objType == 74) model.a(0, 0, -480, true)
//
// the 3rd arg is baseZ, which GameModel.apply (GameModel.java:838-847,1200-1201)
// maps to the vertical Y axis; -Y is up in this engine, hence a NEGATIVE lift.
// (The deob also continuously yaw-spins the sails each frame, Mudclient.java:3347
// — that is live animation, out of scope for static placement; only the -480
// lift is the placement bug.) Forward-ported from our feat/remote-client
// render/scenery.go:147-149 `if DefID==74 { cy -= 480 }`.
func sceneryLiftY(defID int) int32 {
	if defID == 74 {
		return -480
	}
	return 0
}

// SceneryCentre returns the footprint-centre + ground anchor (cx, cy, cz) for an
// object with def id defID, footprint (width x height) and heading dir placed at
// WINDOW-LOCAL tile (x, z) in the window whose SW corner is (baseX, baseY)
// (World.java:182-186). width/height are the object def's footprint (treated as 1
// when <=0); the centre is ((xSize+2x)*128)/2 and the anchor Y is -getElevation
// at that centre, plus any per-id vertical lift (sceneryLiftY, e.g. id-74 sails).
func SceneryCentre(land *pathfind.Landscape, baseX, baseY, plane, x, z, width, height, dir, defID int) (cx, cy, cz int32) {
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}

	// World.java:173-179 — xSize/zSize swap for a heading other than 0/4.
	xSize, zSize := width, height
	if xSize <= 0 {
		xSize = 1
	}
	if zSize <= 0 {
		zSize = 1
	}
	if dir != 0 && dir != 4 {
		xSize, zSize = zSize, xSize
	}

	cx = int32(xSize+x+x) * 128 / 2
	cz = int32(zSize+z+z) * 128 / 2
	cy = -b.getElevation(cx, cz)
	cy += sceneryLiftY(defID) // id-74 windmill sails lift 480 units up the tower
	return cx, cy, cz
}

// PlaceScenery applies the addLoginScreenModels orientation + translate
// (World.java:186-187) to a scenery *Model in place: setRot256(0, dir*32, 0) then
// setTranslate(cx, cy, cz) at the footprint centre / ground anchor (cy carries the
// per-id vertical lift, e.g. the id-74 windmill-sail -480, via SceneryCentre). The
// model must already be built (its geometry loaded via NewModel/FromAssets);
// Orient/Translate mark the transform dirty so the first project() bakes it. The
// caller does Scene.AddModel (World.java:188). defID is the scenery loc's DefID,
// threaded in for the per-id lift.
func PlaceScenery(m *Model, land *pathfind.Landscape, baseX, baseY, plane, x, z, width, height, dir, defID int) {
	cx, cy, cz := SceneryCentre(land, baseX, baseY, plane, x, z, width, height, dir, defID)
	m.Orient(0, int32(dir*32)&255, 0) // World.java:187 setRot256(0, dir*32, 0)
	m.Translate(cx, cy, cz)           // World.java:186 translate2(...) absolute
}
