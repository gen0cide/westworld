package render

import "github.com/gen0cide/westworld/pathfind"

// terrainSize is the side length of the heightmap window (96x96 tiles, 2x2
// sectors), matching World.loadSection.
const terrainSize = 96

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
	nF := (n - 1) * (n - 1)
	g := NewGameModel(nV, nF)

	idx := func(i, j int) int { return i*n + j }
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			g.AddVertex(int32(i)*128, -h[i][j], int32(j)*128)
		}
	}

	for i := 0; i < n-1; i++ {
		for j := 0; j < n-1; j++ {
			c := groundColour[col[i][j]&0xff]
			// Overlay floor-types (road, dirt, slab) OVERWRITE the grass
			// underlay colour, matching World.java:714-726. Road = overlay
			// id 1 (grey). Water overlays are handled by the isWater branch.
			if oc, ok := overlayColour[ovl[i][j]]; ok {
				c = oc
			}
			// A "shore" quad has some — but not all — of its 4 corners
			// flattened to the water plane: it is the steep transition face
			// between water (height 0) and land (height>0). Gouraud-shaded,
			// its low near-vertical normal drives those vertices to pure black
			// (the "dark wedge" artifact). Paint shore + full-water quads the
			// flat water colour with FLAT (non-gouraud) shading so the
			// coastline reads as water/shoreline, never a black gash.
			// Only ACTUAL water tiles get the flat water surface. The old
			// `|| flatCorners > 0` repainted every bank/shore quad (any single
			// water-touched corner) as flat blue, collapsing the river valley's
			// slope shading into a flat smear — the "we don't render height well"
			// look. Bank quads now stay gouraud LAND so the terrain reads as
			// dipping to the river and rising on the far bank. (Authentic: only
			// anIntArray98==4 water tiles draw the water surface — World.java:820.)
			// The bank's ~60° slope normal is nowhere near vertical, so it does
			// NOT gouraud-darken to the old black "dark wedge".
			isWater := water[i][j]
			// quad (i,j)-(i+1,j)-(i+1,j+1)-(i,j+1), CCW
			v0 := idx(i, j)
			v1 := idx(i+1, j)
			v2 := idx(i+1, j+1)
			v3 := idx(i, j+1)
			if isWater {
				// Water/shore quad: paint flat water-blue at a FIXED bright
				// intensity (AddFixedFace) so relight() never recomputes it from
				// the near-vertical shore normal — which is what drove the sunk
				// corners to a pure-black "dark wedge". waterShade is a high ramp
				// index (~bright) so the whole coastline reads water-blue from
				// every camera angle.
				g.AddFixedFace([]int{v0, v1, v2, v3}, waterColour, waterColour, waterShade)
			} else {
				g.AddFace([]int{v0, v1, v2, v3}, c, c, magic) // gouraud land
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
