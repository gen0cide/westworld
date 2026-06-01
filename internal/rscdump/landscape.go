package rscdump

import "github.com/gen0cide/westworld/pathfind"

// Landscape builds a pathfind.Landscape backed by this dump's per-tile Terrain
// grids, so the unchanged render.RenderView path can consume the dump exactly
// as it consumes a real .orsc archive (RENDER_DIFF_DESIGN.md §5 "build a
// synthetic pathfind.Landscape-equivalent from terrain.*").
//
// The dump grids are window-local (index [x*Size + y], x = worldX − BaseX). We
// scatter each tile into the 48×48 sector layout at its ABSOLUTE world coord via
// the same pathfind.SectorForWorld / TileLocalInSector mapping the renderer uses
// to read tiles back, so a Tile(worldX, worldY, plane) lookup returns the dumped
// byte. Tiles outside the window resolve to the zero Tile (void), matching a
// missing sector in the on-disk archive.
//
// Returns nil for a non-L1 dump (no terrain).
func (d *Dump) Landscape() *pathfind.Landscape {
	if d.Terrain == nil {
		return nil
	}
	t := d.Terrain
	plane := d.Window.Plane
	baseX, baseY := d.Window.BaseX, d.Window.BaseY

	get := func(g []byte, x, y int) byte {
		if len(g) == 0 {
			return 0
		}
		return g[t.idx(x, y)]
	}
	getI := func(g []int32, x, y int) int32 {
		if len(g) == 0 {
			return 0
		}
		return g[t.idx(x, y)]
	}

	sectors := make(map[pathfind.SectorKey]*pathfind.Sector)
	for x := 0; x < t.Size; x++ {
		for y := 0; y < t.Size; y++ {
			wx, wy := baseX+x, baseY+y
			key := pathfind.SectorForWorld(wx, wy, plane)
			sec := sectors[key]
			if sec == nil {
				sec = &pathfind.Sector{}
				sectors[key] = sec
			}
			lx, ly := pathfind.TileLocalInSector(wx, wy)
			sec.Tiles[lx*pathfind.SectorSize+ly] = pathfind.Tile{
				GroundElevation: get(t.Elevation, x, y),
				GroundTexture:   get(t.GroundColour, x, y),
				GroundOverlay:   get(t.Overlay, x, y),
				RoofTexture:     get(t.Roof, x, y),
				HorizontalWall:  get(t.WallH, x, y),
				VerticalWall:    get(t.WallV, x, y),
				DiagonalWalls:   getI(t.WallDiag, x, y),
			}
		}
	}
	return pathfind.NewMemoryLandscape(sectors)
}
