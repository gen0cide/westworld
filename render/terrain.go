package render

import "github.com/gen0cide/westworld/pathfind"

// terrainSize is the side length of the heightmap window (96x96 tiles, 2x2
// sectors), matching World.loadSection.
const terrainSize = 96

// BuildTerrain assembles the heightmap mesh GameModel for the 96x96 window
// whose SW corner is (baseX, baseY) in world-tile coords (use the same
// midRegion centring as pathfind.BuildGrid). Vertices are placed at
// (x*128, -height, y*128) with height = elevation*3 (World.getTerrainHeight),
// and quads are flat-coloured from the ground palette (no overlays/textures).
func BuildTerrain(land *pathfind.Landscape, baseX, baseY, plane int) *GameModel {
	n := terrainSize
	// elevation + colour grids
	h := make([][]int32, n)
	col := make([][]int32, n)
	for i := 0; i < n; i++ {
		h[i] = make([]int32, n)
		col[i] = make([]int32, n)
		for j := 0; j < n; j++ {
			t := land.Tile(baseX+i, baseY+j, plane)
			h[i][j] = int32(t.GroundElevation) * 3
			col[i][j] = int32(t.GroundTexture)
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
	return g
}
