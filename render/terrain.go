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

// waterColour is the flat fill for water tiles (method305-encoded blue). The
// reference uses the config tile colour (anIntArray97) but absent the config
// table we paint a plausible RSC water blue so water reads as water, not void.
var waterColour = method305(40, 70, 140)

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
	water := make([][]bool, n)
	for i := 0; i < n; i++ {
		h[i] = make([]int32, n)
		col[i] = make([]int32, n)
		water[i] = make([]bool, n)
		for j := 0; j < n; j++ {
			t := land.Tile(baseX+i, baseY+j, plane)
			h[i][j] = int32(t.GroundElevation) * 3
			col[i][j] = int32(t.GroundTexture)
			// GroundOverlay (getTileDecoration) == 2 is the water decoration:
			// in World.loadSection a water tile category (anIntArray98==4)
			// FORCES the vertex height to 0 (flat) so it never forms a cliff,
			// and paints it the water colour. Without this, the elev-0 water
			// pits beside elev-128 land formed vertical faces that gouraud-shade
			// to pure black (the "dark wedge" artifact).
			if t.GroundOverlay == waterOverlay {
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
			if water[i][j] {
				c = waterColour
			}
			// quad (i,j)-(i+1,j)-(i+1,j+1)-(i,j+1), CCW
			v0 := idx(i, j)
			v1 := idx(i+1, j)
			v2 := idx(i+1, j+1)
			v3 := idx(i, j+1)
			g.AddFace([]int{v0, v1, v2, v3}, c, c, magic)
		}
	}
	// Match the real World terrain light: setLight(true, 40, 48, -50,-10,-50)
	// -> ambience 96, diffuse 384, gouraud, light dir (-50,-10,-50).
	g.SetLight(40, 48, -50, -10, -50)
	return g, h
}
