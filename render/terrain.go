package render

import "github.com/gen0cide/westworld/pathfind"

// terrainSize is the side length of the heightmap window (96x96 tiles, 2x2
// sectors), matching World.loadSection.
const terrainSize = 160

// waterOverlay is the GroundOverlay (getTileDecoration) value for water tiles
// in the OpenRSC landscape. Water is the dominant elev-0 overlay in the data
// (435/450 overlay-2 tiles sit at elevation 0). The reference client flattens
// water-category tiles (GameData.anIntArray98==4) to the water plane.
const waterOverlay = 2

// waterOverlay2 is the second water decoration id (11) the pathfind grid also
// treats as water (pathfind/grid.go). Both flatten + recolour to the water
// plane so neither forms a black gouraud cliff.
const waterOverlay2 = 11

// waterColour is the flat fill for water tiles (method305-encoded blue). The
// reference uses the config tile colour (anIntArray97) but absent the config
// table we paint a plausible RSC water blue so water reads as water, not void.
var waterColour = method305(40, 70, 140)

// waterTextureID is the authentic water texture (textures17.jag index 1). RSC
// textures the river surface with the animated water sprite; we render its
// first frame so the river reads as rippled water rather than a flat blue
// smear. A non-negative fill routes the face through the perspective
// textured-span path (method282); textureBuffer(1) is always present.
const waterTextureID int32 = 1

// waterShade is the FIXED flat shade index every water / shore quad is lit at.
// The shade ramp is built as ramp[255-j] = colour*(j*j)/65536, so a LOW index
// is BRIGHT and a high index is dark (index 0 = full colour, 255 = black). We
// pick a low index so water reads as bright water-blue. Because it is fixed
// (AddFixedFace), relight() never recomputes it from the steep shore normal, so
// the coastline can never gouraud-darken to the black "dark wedge".
const waterShade = 40

// BuildTerrain assembles the heightmap mesh GameModel for the 96x96 window
// whose SW corner is (baseX, baseY) in world-tile coords (use the same
// midRegion centring as pathfind.BuildGrid). Vertices are placed at
// (x*128, -height, y*128) with height = elevation*3 (World.getTerrainHeight),
// and quads are flat-coloured from the ground palette (no overlays/textures).
func BuildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) *GameModel {
	g, _ := buildTerrain(land, baseX, baseY, plane)
	return g
}

// TerrainHeights returns the flattened terrain-height grid for the window —
// the SAME heights BuildTerrain places its vertices at (water-flattened). The
// boundary builder anchors wall quads to these so walls sit flush on the
// ground (RSC method422 reads terrainHeightLocal, the post-flatten cache).
func TerrainHeights(land *pathfind.Landscape, baseX, baseY, plane int) [][]int32 {
	_, h := buildTerrain(land, baseX, baseY, plane)
	return h
}

func buildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) (*GameModel, [][]int32) {
	n := terrainSize
	// elevation + colour + water grids
	h := make([][]int32, n)
	col := make([][]int32, n)
	ovl := make([][]byte, n)
	water := make([][]bool, n)
	for i := 0; i < n; i++ {
		h[i] = make([]int32, n)
		col[i] = make([]int32, n)
		ovl[i] = make([]byte, n)
		water[i] = make([]bool, n)
		for j := 0; j < n; j++ {
			t := land.Tile(baseX+i, baseY+j, plane)
			h[i][j] = int32(t.GroundElevation) * 3
			col[i][j] = int32(t.GroundTexture)
			ovl[i][j] = t.GroundOverlay
			// GroundOverlay (getTileDecoration) == 2 is the water decoration:
			// in World.loadSection a water tile category (anIntArray98==4)
			// FORCES the vertex height to 0 (flat) so it never forms a cliff,
			// and paints it the water colour. Without this, the elev-0 water
			// pits beside elev-128 land formed vertical faces that gouraud-shade
			// to pure black (the "dark wedge" artifact).
			if t.GroundOverlay == waterOverlay || t.GroundOverlay == waterOverlay2 {
				water[i][j] = true
			}
		}
	}

	// Flatten any vertex touched by a water tile to height 0 (the reference's
	// 4-corner check: a vertex is shared by up to 4 tiles; if ANY is water,
	// the vertex sinks to the water plane).
	flat := make([][]bool, n)
	for i := 0; i < n; i++ {
		flat[i] = make([]bool, n)
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if !water[i][j] {
				continue
			}
			for di := 0; di <= 1; di++ {
				for dj := 0; dj <= 1; dj++ {
					vi, vj := i+di, j+dj
					if vi < n && vj < n {
						flat[vi][vj] = true
					}
				}
			}
		}
	}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if flat[i][j] {
				h[i][j] = 0
			}
		}
	}

	nV := n * n
	// up to 2 faces per tile: a non-planar tile splits into two triangles
	// (World.java:762-815), so size for the worst case to avoid append-growth.
	nF := (n - 1) * (n - 1) * 2
	g := NewGameModel(nV, nF)

	idx := func(i, j int) int { return i*n + j }
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			vi := g.AddVertex(int32(i)*128, -h[i][j], int32(j)*128)
			// Authentic per-vertex ambience speckle (World.java:696-697:
			// (int)(Math.random()*10)-5). Kept deterministic for render
			// caching by hashing the WORLD-tile coord instead of time-seeding.
			g.SetVertexAmbience(vi, terrainAmbience(baseX+i, baseY+j))
		}
	}

	// ovlClassAt is the tile's overlay id used as the surface "class" for the
	// diagonal colour split (the client's method420); 0 (grass underlay) outside
	// the window.
	ovlClassAt := func(a, b int) byte {
		if a < 0 || b < 0 || a >= n || b >= n {
			return 0
		}
		return ovl[a][b]
	}
	isWaterAt := func(a, b int) bool {
		return a >= 0 && a < n && b >= 0 && b < n && water[a][b]
	}
	// emitHalf draws one terrain triangle: a WATER half is back-face fixed-
	// textured (fillBack = water id, fixed bright shade so it never gouraud-
	// darkens); a land half is back-face gouraud at flat colour `fill`.
	emitHalf := func(tri []int, fill int32, isW bool) {
		if isW {
			g.AddFixedFace(tri, magic, waterTextureID, waterShade)
		} else {
			g.AddFace(tri, magic, fill, magic)
		}
	}

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			// quad corners (i,j)-(i+1,j)-(i+1,j+1)-(i,j+1)
			v0 := idx(i, j)
			v1 := idx(i+1, j)
			v2 := idx(i+1, j+1)
			v3 := idx(i, j+1)
			if water[i][j] {
				// Water tile: authentic water TEXTURE (id 1) via the textured-span
				// path, FIXED bright shade (the tile is flattened to the horizontal
				// y=0 plane so it never gouraud-darkens). Water is the client's
				// separate type-4 pass (World.java:820); the land tiles below handle
				// their own colour boundaries.
				g.AddFixedFace([]int{v0, v1, v2, v3}, waterTextureID, waterTextureID, waterShade)
				continue
			}
			// Resolve UNDERLAY (grass) colour + the tile's final colour (an overlay
			// floor — road/dirt/slab — OVERWRITES the underlay, World.java:714-726).
			underlay := groundColour[col[i][j]&0xff]
			c := underlay
			hasOverlay := false
			if def, ok := overlayDef(ovl[i][j]); ok {
				switch def.tileType {
				case 4:
					// Water-class tile (ids 4/12/20/21): the client FORCES these to
					// the water texture (deob World.java:719-726, OpenRSC
					// World.java:574-581) — id 12 to lava-ish tex 31, the rest to
					// water tex 1. Without this they'd fall through to def.colour
					// (e.g. id 4's colour 3 = planks) and wrongly render as a wood
					// floor in open water.
					c = waterTextureID
					if ovl[i][j] == 12 {
						c = 31
					}
					hasOverlay = true
				case 5:
					// Bridge sentinel (colour 12345678): out of scope — keep the
					// grass underlay so the sentinel never gets used as a texture id.
				default:
					// colour >= 0 is a TEXTURE id (chapel planks=3, marble=17->15,
					// pentagram=14->32, outdoor textured ids); < 0 is a method305 flat
					// colour. Either way scene.go routes the fill correctly: a >=0
					// fill textures the floor, a <0 fill paints flat.
					c = def.colour
					hasOverlay = true
				}
			}

			// COLOUR SPLIT (World.java:704-815): an overlay (path/road) tile reverts
			// ONE triangle half to the underlay (grass) on the side facing tiles of a
			// DIFFERENT surface, and picks which diagonal (l14) so the overlay edge
			// cuts DIAGONALLY across the tile instead of stair-stepping along the
			// grid — this is what makes paths read as smooth diagonals. k7 is the
			// first triangle's colour, i10 the second's; equal => no colour split.
			// k7/i10 = the two triangle colours; w0/w1 mark a half as WATER. l14
			// picks which diagonal carries the seam (World.java:766-803).
			k7, i10, l14 := c, c, 0
			w0, w1 := false, false
			if hasOverlay {
				me := ovl[i][j]
				switch { // mirrors World.java:743-755 (method420 neighbour compares)
				case ovlClassAt(i-1, j) != me && ovlClassAt(i, j-1) != me:
					k7, l14 = underlay, 0
				case ovlClassAt(i+1, j) != me && ovlClassAt(i, j+1) != me:
					i10, l14 = underlay, 0
				case ovlClassAt(i+1, j) != me && ovlClassAt(i, j-1) != me:
					i10, l14 = underlay, 1
				case ovlClassAt(i-1, j) != me && ovlClassAt(i, j+1) != me:
					k7, l14 = underlay, 1
				}
			} else {
				// SHORELINE diagonal (enhancement): where a grass tile meets water on
				// two ADJACENT edges (a corner of the water body), make the corner-
				// facing triangle WATER so the shore cuts diagonally instead of as a
				// blocky tile step — same diagonal-selection shape as the overlay split.
				switch {
				case isWaterAt(i-1, j) && isWaterAt(i, j-1):
					w0, l14 = true, 0
				case isWaterAt(i+1, j) && isWaterAt(i, j+1):
					w1, l14 = true, 0
				case isWaterAt(i+1, j) && isWaterAt(i, j-1):
					w1, l14 = true, 1
				case isWaterAt(i-1, j) && isWaterAt(i, j+1):
					w0, l14 = true, 1
				}
			}

			// Height twist: a non-planar tile must ALSO split (round-5 relief fix),
			// so the gouraud normal is exact per planar triangle (no faceting).
			twist := (h[i+1][j+1] - h[i+1][j]) + h[i][j+1] - h[i][j]

			if k7 != i10 || w0 != w1 || twist != 0 {
				if l14 == 0 {
					emitHalf([]int{v1, v0, v3}, k7, w0)
					emitHalf([]int{v3, v2, v1}, i10, w1)
				} else {
					emitHalf([]int{v3, v2, v0}, k7, w0)
					emitHalf([]int{v1, v0, v2}, i10, w1)
				}
			} else {
				g.AddFace([]int{v1, v0, v3, v2}, magic, c, magic)
			}
		}
	}

	// Match the real World terrain light: setLight(true, 40, 48, -50,-10,-50)
	// -> ambience 96, diffuse 384, gouraud, light dir (-50,-10,-50).
	g.SetLight(40, 48, -50, -10, -50)
	return g, h
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// terrainAmbience returns a stable pseudo-random ambience offset in [-5,4] for
// a world-tile coord, mirroring World.java's (int)(Math.random()*10)-5 but
// derived from a cheap integer hash of (x,y) so renders stay deterministic
// (cacheable) instead of time-seeded.
func terrainAmbience(x, y int) int32 {
	h := uint32(x)*0x9e3779b1 + uint32(y)*0x85ebca77
	h ^= h >> 15
	h *= 0x2c1b3c6d
	h ^= h >> 12
	return int32(h%10) - 5
}
