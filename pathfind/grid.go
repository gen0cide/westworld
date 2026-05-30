package pathfind

import (
	"github.com/gen0cide/westworld/facts"
)

// GridSize is the side length of the local pathfind grid. The original
// RSC client uses a 96x96 grid covering 4 sectors (2x2) around the
// player; we mirror that exactly so the BFS expansion logic in
// findPath() can be a direct port.
const GridSize = 96

// Grid is a fixed-size collision map centered on a region origin.
// Mask[x][y] holds the bitwise OR of WALL_* / *Blocked / FullBlockX
// flags for the tile at world (BaseX+x, BaseY+y) — same convention
// the client uses on `collisionFlags` inside World.java.
//
// Build with BuildGrid; the player's local coords are (PlayerX,
// PlayerY) ∈ [0, GridSize).
type Grid struct {
	Mask    [GridSize][GridSize]int
	BaseX   int // world X of grid cell (0, _)
	BaseY   int // world Y of grid cell (_, 0)
	PlayerX int // local X of the player's tile (within [0, GridSize))
	PlayerY int // local Y of the player's tile
}

// World2Local returns the local grid coords for a world tile, plus
// whether the tile is inside the grid bounds.
func (g *Grid) World2Local(worldX, worldY int) (int, int, bool) {
	lx := worldX - g.BaseX
	ly := worldY - g.BaseY
	return lx, ly, lx >= 0 && lx < GridSize && ly >= 0 && ly < GridSize
}

// Local2World is the inverse of World2Local.
func (g *Grid) Local2World(localX, localY int) (int, int) {
	return g.BaseX + localX, g.BaseY + localY
}

// TileStandable reports whether a bot can occupy the world tile
// (worldX, worldY): it must be inside the grid and not fully blocked by
// terrain or a solid scenery object. Edge-wall flags are NOT consulted
// (those gate movement across an edge, not standing on the tile), so
// this answers "is this a valid tile to stand on", used to pick an
// adjacent tile to step to when stacked on an NPC. Tiles outside the
// loaded grid return false (conservative).
func (g *Grid) TileStandable(worldX, worldY int) bool {
	lx, ly, ok := g.World2Local(worldX, worldY)
	if !ok {
		return false
	}
	m := g.Mask[lx][ly]
	if m&fullBlock != 0 {
		return false
	}
	if m&Object != 0 {
		return false
	}
	return true
}

// BuildGrid assembles a Grid from a Landscape archive plus facts data,
// centered on the player at world (worldX, worldY) on the given plane.
//
// The grid origin is chosen by the same midRegion math the client
// uses: midRegionX = (worldX + 24) / 48, baseX = midRegionX*48 - 48.
// This makes the grid span a 2x2 sector area covering ~48 tiles in
// each direction from the player.
//
// Wall flags follow the server's loadSection convention: a wall blocks
// only when its DoorDef has DoorType != 0 AND Unknown == 0. Openable
// doors (Unknown=1) are treated as walkable — the bot resolves them
// via InteractWithBoundary before crossing.
func BuildGrid(ls *Landscape, f *facts.Facts, worldX, worldY, plane int) (*Grid, error) {
	if plane != 0 {
		// Multi-plane support would just require knowing wildY per
		// plane (1776 - lvl*944) and using the corresponding archive
		// offsets. Phase 1 stays on ground level.
		plane = 0
	}

	midRegionX := (worldX + 24) / SectorSize
	midRegionY := (worldY + 24) / SectorSize
	baseX := midRegionX*SectorSize - SectorSize
	baseY := midRegionY*SectorSize - SectorSize

	g := &Grid{
		BaseX:   baseX,
		BaseY:   baseY,
		PlayerX: worldX - baseX,
		PlayerY: worldY - baseY,
	}

	// Load the 2x2 sector neighborhood.
	for dsx := 0; dsx < 2; dsx++ {
		for dsy := 0; dsy < 2; dsy++ {
			ingameSX := midRegionX - 1 + dsx
			ingameSY := midRegionY - 1 + dsy
			key := SectorKey{
				Plane: plane,
				SX:    ingameSX + archiveOffsetX,
				SY:    ingameSY + archiveOffsetY,
			}
			s, err := ls.Sector(key)
			if err != nil {
				return nil, err
			}
			if s == nil {
				continue // void area — leave grid tiles at zero
			}
			applySectorToGrid(g, s, dsx*SectorSize, dsy*SectorSize, f)
		}
	}

	// Layer scenery + boundary locs from the facts data on top of the
	// landscape walls. Scenery doesn't appear in the .orsc tile file
	// (it's stored separately as locs and computed into walls at
	// client load time via addGameObject_UpdateCollisionMap).
	applyScenery(g, f)
	applyBoundaries(g, f)

	return g, nil
}

// applySectorToGrid copies the WALL_* / FullBlock flags for one
// sector's 48x48 tiles into a sub-rectangle of the grid. Mirrors the
// server's WorldLoader.loadSection rules so the bot sees the same
// world the server collision-checks against.
func applySectorToGrid(g *Grid, s *Sector, offX, offY int, f *facts.Facts) {
	for sx := 0; sx < SectorSize; sx++ {
		for sy := 0; sy < SectorSize; sy++ {
			gx := offX + sx
			gy := offY + sy
			if gx < 0 || gx >= GridSize || gy < 0 || gy >= GridSize {
				continue
			}
			tile := s.At(sx, sy)

			// verticalWall lives on this tile's NORTH edge. Server
			// also mirrors WALL_SOUTH onto the north neighbor.
			if vw := int(tile.VerticalWall); vw > 0 {
				if d, ok := f.BoundaryDefs[vw-1]; ok && d.BlocksMovement() {
					g.Mask[gx][gy] |= WallNorth
					if gy > 0 {
						g.Mask[gx][gy-1] |= WallSouth
					}
				}
			}

			// horizontalWall lives on this tile's WEST edge (named
			// WALL_EAST in the bit layout — see flags.go for the
			// convention). Mirrors WALL_WEST onto the west neighbor.
			if hw := int(tile.HorizontalWall); hw > 0 {
				if d, ok := f.BoundaryDefs[hw-1]; ok && d.BlocksMovement() {
					g.Mask[gx][gy] |= WallEast
					if gx > 0 {
						g.Mask[gx-1][gy] |= WallWest
					}
				}
			}

			// Diagonal walls: positive values < 12000 are SW-NE,
			// 12001..23999 are NW-SE. Both fully block their tile per
			// the server's rule.
			if dw := int(tile.DiagonalWalls); dw > 0 {
				idx := dw - 1
				if dw > 12000 && dw < 24000 {
					idx = dw - 12001
				}
				if d, ok := f.BoundaryDefs[idx]; ok && d.BlocksMovement() {
					if dw > 0 && dw < 12000 {
						g.Mask[gx][gy] |= FullBlockB
					} else if dw > 12000 && dw < 24000 {
						g.Mask[gx][gy] |= FullBlockA
					}
				}
				// 48001..59999 encodes inline scenery (the client's
				// addLoginScreenModels path). Those tiles are
				// covered by the multi-tile scenery expansion in
				// applyScenery; we don't need to re-block here.
			}

			// Ground overlay 2 = water, 11 = also non-walkable per
			// the server. We don't have TileDef.xml ingested yet, so
			// hardcode the well-known impassable overlay values.
			if ov := int(tile.GroundOverlay); ov == 2 || ov == 11 {
				g.Mask[gx][gy] |= FullBlockC
			}
		}
	}
}

// applyScenery layers scenery (trees, signs, counters, etc.) on top of
// the landscape walls. Mirrors World.addGameObject_UpdateCollisionMap:
// type 1 = full block on the tile; type 2 = directional wall on one
// tile edge keyed by the loc's Direction; other types (3+) are
// non-blocking decoration.
func applyScenery(g *Grid, f *facts.Facts) {
	for _, loc := range f.SceneryLocs {
		gx, gy, ok := g.World2Local(loc.X, loc.Y)
		if !ok {
			continue
		}
		def, has := f.SceneryDefs[loc.DefID]
		if !has || def == nil {
			continue
		}
		switch def.Type {
		case 1:
			// Solid scenery (tree, sign post, statue, well) —
			// full block across its width × height footprint.
			// Direction 0 and 4 use the def's width/height as-is;
			// other directions swap them (rotated 90°). Mirrors
			// World.addGameObject_UpdateCollisionMap rules from
			// the openrsc client.
			w, h := def.Width, def.Height
			if loc.Direction != 0 && loc.Direction != 4 {
				w, h = h, w
			}
			if w < 1 {
				w = 1
			}
			if h < 1 {
				h = 1
			}
			for dy := 0; dy < h; dy++ {
				for dx := 0; dx < w; dx++ {
					ex, ey := gx+dx, gy+dy
					if ex >= 0 && ex < GridSize && ey >= 0 && ey < GridSize {
						g.Mask[ex][ey] |= FullBlockC | Object
					}
				}
			}
		case 2:
			// Directional wall — exactly one edge of this tile
			// becomes blocking, mirrored onto the matching neighbor.
			// Directions 0/2/4/6 are N/E/S/W variants per RSC
			// convention (see World.addGameObject_UpdateCollisionMap).
			switch loc.Direction {
			case 0:
				g.Mask[gx][gy] |= WallEast
				if gx > 0 {
					g.Mask[gx-1][gy] |= WallWest
				}
			case 2:
				g.Mask[gx][gy] |= WallSouth
				if gy < GridSize-1 {
					g.Mask[gx][gy+1] |= WallNorth
				}
			case 4:
				g.Mask[gx][gy] |= WallWest
				if gx < GridSize-1 {
					g.Mask[gx+1][gy] |= WallEast
				}
			case 6:
				g.Mask[gx][gy] |= WallNorth
				if gy > 0 {
					g.Mask[gx][gy-1] |= WallSouth
				}
			}
		}
		// type 3+ → non-blocking decoration; leave the tile alone.
	}
}

// applyBoundaries layers boundary locs (door instances, gates, fences)
// on top. We only block at runtime for boundaries whose def is
// flagged BlocksMovement — i.e., fences and railings, not doors.
// The boundary direction encodes which edge the wall sits on, matching
// World.applyWallToCollisionFlags in the client.
func applyBoundaries(g *Grid, f *facts.Facts) {
	for _, loc := range f.BoundaryLocs {
		def, ok := f.BoundaryDefs[loc.DefID]
		if !ok || !def.BlocksMovement() {
			continue
		}
		gx, gy, in := g.World2Local(loc.X, loc.Y)
		if !in {
			continue
		}
		switch loc.Direction {
		case 0:
			// Wall on this tile's north edge; mirror on north
			// neighbor's south edge.
			g.Mask[gx][gy] |= WallNorth
			if gy > 0 {
				g.Mask[gx][gy-1] |= WallSouth
			}
		case 1:
			// Wall on this tile's west edge; mirror on west neighbor's
			// east edge.
			g.Mask[gx][gy] |= WallEast
			if gx > 0 {
				g.Mask[gx-1][gy] |= WallWest
			}
		case 2:
			// Diagonal NW-SE.
			g.Mask[gx][gy] |= FullBlockA
		case 3:
			// Diagonal SW-NE.
			g.Mask[gx][gy] |= FullBlockB
		}
	}
}
