package pathfind

import (
	"github.com/gen0cide/westworld/facts"
)

// GridSize is the side length of the local pathfind grid. The original
// RSC client uses a 96x96 grid covering 4 sectors (2x2) around the
// player; we mirror that exactly so the BFS expansion logic in
// findPath() can be a direct port.
const GridSize = 96

// planeHeight is the vertical Y-band per floor: RSC stacks planes in Y at
// 944-tile intervals, so plane = worldY/944 and the plane-local footprint Y is
// worldY%944. Mirrors world.PlaneHeight / OpenRSC Formulae.getHeight.
const planeHeight = 944

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

// LiveState is the host's CURRENT dynamic view of the world — the boundary
// (door/gate) and scenery overrides the server streams at runtime via opcodes
// 91 (SEND_BOUNDARY_HANDLER) and 48 (SEND_SCENERY_HANDLER). The runtime caller
// builds it from world.Boundaries / world.Scenery and hands it to BuildGrid so
// the BFS reflects OPENED doors (passable), CLOSED doors (walls), depleted
// scenery (walkable) and runtime-spawned blockers (lit fires). A nil LiveState
// gives the pure-static grid (tests, map precompute).
type LiveState struct {
	Boundaries []LiveBoundary
	Scenery    []LiveScenery
}

// LiveBoundary is one live door/gate override at a tile edge. Removed (or
// ID < 0) means the boundary is gone / the door is open (passable).
type LiveBoundary struct {
	X, Y, Dir, ID int
	Removed       bool
}

// LiveScenery is one live scenery override at a tile. Removed = the server
// cleared the tile (depleted rock / burned-out fire); otherwise ID is the
// object now occupying it.
type LiveScenery struct {
	X, Y, ID int
	Removed  bool
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
func BuildGrid(ls *Landscape, f *facts.Facts, worldX, worldY, plane int, live *LiveState) (*Grid, error) {
	// Planes stack in Y at 944-tile bands (plane = worldY/944). Auto-derive the
	// plane from the player's band-encoded worldY when the caller passes 0 (the
	// common case — every runtime caller); a caller may pass an explicit
	// plane > 0 to force a floor. The .orsc archive is keyed by plane-LOCAL
	// footprint coords + the h{plane} prefix (see landscape.go SectorForWorld),
	// so we look sectors up by the plane-local Y row, but keep the grid ORIGIN
	// in band-encoded world Y — facts scenery/boundary locs and the live
	// overlay are all band-encoded, so World2Local must use world coords.
	if plane <= 0 {
		plane = worldY / planeHeight
	}
	localY := worldY - plane*planeHeight // plane-local footprint Y in [0, 944)

	midRegionX := (worldX + 24) / SectorSize
	midRegionLY := (localY + 24) / SectorSize
	baseX := midRegionX*SectorSize - SectorSize
	baseLocalY := midRegionLY*SectorSize - SectorSize
	baseY := baseLocalY + plane*planeHeight // grid origin, band-encoded world Y

	g := &Grid{
		BaseX:   baseX,
		BaseY:   baseY,
		PlayerX: worldX - baseX,
		PlayerY: worldY - baseY, // == localY - baseLocalY, in [0, GridSize)
	}

	// Load the 2x2 sector neighborhood. Archive rows come from the plane-LOCAL
	// Y; the h{plane} prefix selects the floor at the same (sx, sy) coords.
	for dsx := 0; dsx < 2; dsx++ {
		for dsy := 0; dsy < 2; dsy++ {
			ingameSX := midRegionX - 1 + dsx
			ingameSLY := midRegionLY - 1 + dsy
			key := SectorKey{
				Plane: plane,
				SX:    ingameSX + archiveOffsetX,
				SY:    ingameSLY + archiveOffsetY,
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

	// Reconcile the static layers with the host's LIVE view: open doors become
	// passable, closed doors stay walls, depleted/spawned scenery updates. This
	// is what stops the BFS planning through a closed door or into a lit fire.
	applyLiveOverlay(g, f, live)

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

			// Ground-overlay terrain collision per TileDef.xml: the overlay
			// byte is a 1-based TileDef index (250→2 remap) and blocks when
			// ObjectType != 0 — water, lava, void. Authoritative (mirrors
			// Plutonium world.go:189-195 / the server's WorldLoader); the
			// old ov==2||ov==11 hardcode under-blocked other overlays and
			// survives only as OverlayBlocks' no-table fallback.
			if f.OverlayBlocks(int(tile.GroundOverlay)) {
				g.Mask[gx][gy] |= FullBlockC
			}
		}
	}
}

// applyScenery layers scenery (trees, signs, counters, gates, etc.) on top
// of the landscape walls via stampSceneryFootprint — the single collision
// rule shared with the live overlay so static and dynamic state can never
// disagree on geometry.
func applyScenery(g *Grid, f *facts.Facts) {
	for _, loc := range f.SceneryLocs {
		stampSceneryFootprint(g, loc.X, loc.Y, loc.Direction, f.SceneryDefs[loc.DefID], true)
	}
}

// sceneryCollisionSkipID is the one scenery def the server (World.java:501-503)
// and Plutonium (world.go:345-347) both exclude from collision registration.
const sceneryCollisionSkipID = 1147

// stampSceneryFootprint applies (block=true) or clears (block=false) the
// collision a scenery object contributes, mirroring the client's
// addGameObject_UpdateCollisionMap (Client_Base World.java:105-160) exactly:
//   - the def's width × height footprint, dimension-swapped when the loc
//     Direction is not 0/4 (rotated 90°);
//   - Type 1 (solid: trees, statues, CLOSED GATES) → FullBlockC|Object on
//     EVERY footprint tile;
//   - Type 2 (directional wall: counters, fence-like scenery) → the
//     direction-keyed edge pair (edge + mirrored neighbor edge) on EVERY
//     footprint tile — the missing footprint loop here was the phantom-gap
//     root cause: multi-tile gates had unmodeled halves the BFS planned
//     through;
//   - other types → no collision (decoration).
//
// Takes WORLD coords and bounds-checks per tile, so a footprint whose origin
// lies outside the grid still stamps its overlapping tiles. Nil/skipped defs
// are no-ops.
func stampSceneryFootprint(g *Grid, worldX, worldY, dir int, def *facts.SceneryDef, block bool) {
	if def == nil || def.ID == sceneryCollisionSkipID {
		return
	}
	if def.Type != 1 && def.Type != 2 {
		return
	}
	w, h := def.Width, def.Height
	if dir != 0 && dir != 4 {
		w, h = h, w
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	apply := func(x, y, bit int) {
		if x < 0 || x >= GridSize || y < 0 || y >= GridSize {
			return
		}
		if block {
			g.Mask[x][y] |= bit
		} else {
			g.Mask[x][y] &^= bit
		}
	}
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			gx := worldX + dx - g.BaseX
			gy := worldY + dy - g.BaseY
			switch def.Type {
			case 1:
				apply(gx, gy, FullBlockC|Object)
			case 2:
				// Direction-keyed edge + mirrored neighbor edge, per RSC
				// convention (dirs 0/2/4/6 = E/S/W/N edge variants).
				switch dir {
				case 0:
					apply(gx, gy, WallEast)
					apply(gx-1, gy, WallWest)
				case 2:
					apply(gx, gy, WallSouth)
					apply(gx, gy+1, WallNorth)
				case 4:
					apply(gx, gy, WallWest)
					apply(gx+1, gy, WallEast)
				case 6:
					apply(gx, gy, WallNorth)
					apply(gx, gy-1, WallSouth)
				}
			}
		}
	}
}

// applyBoundaries layers boundary locs (door instances, gates, fences) on top.
// We block ONLY fixed walls/fences (BlocksMovement: DoorType!=0 && Unknown==0).
// Openable doors/gates (Unknown==1) are LEFT PASSABLE here: the BFS plans
// through them and the walk flow's stall→open opens them en route — the
// long-tested traversal. (Blocking openable gates as a "closed baseline" penned
// the bot at gates it used to walk through, because the open→clear→replan cycle
// isn't yet robust — see backlog #4. Re-enable only with that in place.) The
// boundary direction encodes which edge the wall sits on, matching the client.
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

// applyLiveOverlay reconciles the static grid with the host's LIVE view — the
// dynamic boundary (door) and scenery state the server streams via opcodes
// 91/48. This is what makes pathing treat an OPENED door as passable, a
// freshly-CLOSED door as a wall, a depleted rock / burned-out fire as walkable,
// and a runtime-spawned blocker (a lit fire) as solid. Mirrors Plutonium's
// combinedTile = static | localWorld, except our static layer already carries
// baseline-closed doors (see applyBoundaries), so the overlay must also CLEAR
// an edge when the server reports the door open. No-op when live == nil.
func applyLiveOverlay(g *Grid, f *facts.Facts, live *LiveState) {
	if live == nil {
		return
	}

	// Doors / gates (wall BOUNDARIES): a live override is authoritative for
	// its edge. The server NEVER sends in-range boundary removals on the
	// authentic wire — a door OPEN arrives as an ID-OVERWRITE to a
	// non-DoorType-1 def (doorframe 11) at the same (x,y,dir), and the
	// re-close 3s later as another overwrite (GameStateUpdater.java:1190-1232)
	// — so blocking is decided by the CURRENT def: fixed walls
	// (BlocksMovement) AND closed openable doors (BlocksWhenClosed) both
	// block, exactly the server's registerObject rule (doorType==1). The BFS
	// then routes TO a closed door and the traversal flow opens it; on the
	// next replan the swapped-in open def clears the edge here. The Removed
	// branch stays as defense for non-authentic dialects.
	for _, b := range live.Boundaries {
		gx, gy, in := g.World2Local(b.X, b.Y)
		if !in {
			continue
		}
		block := false
		if !b.Removed && b.ID >= 0 {
			if d := f.BoundaryDef(b.ID); d != nil {
				block = d.BlocksMovement() || d.BlocksWhenClosed()
			}
		}
		setOrClearEdge(g, gx, gy, b.Dir, block)
	}

	// Scenery (GameObjects): a live record is a DEF SWAP at a static loc —
	// gate opened (60→59), rock depleted (remove sentinel), fire lit (add).
	// The v235 wire carries NO direction byte (Payload235Generator writes
	// [id,x,y] only), so geometry is re-derived from the STATIC loc at the
	// tile: clear the old def's full footprint, then stamp the current def's.
	// stampSceneryFootprint is the same rule the static layer uses, so the
	// two layers can never disagree on a footprint's shape.
	for _, s := range live.Scenery {
		// Out-of-grid live records are skipped wholesale: their footprints
		// are at most a few tiles and the grid re-centers each call.
		if _, _, in := g.World2Local(s.X, s.Y); !in {
			continue
		}
		dir := 0
		var staticDef *facts.SceneryDef
		for _, p := range f.At(s.X, s.Y) {
			if p.Kind == "scenery" {
				dir = p.Direction
				staticDef = f.SceneryDef(p.DefID)
				break
			}
		}
		// Clear whatever the static layer stamped for this loc, then apply
		// the live truth: the current def's footprint, or nothing if the
		// server removed the object (depleted rock, burned-out fire).
		if staticDef != nil {
			stampSceneryFootprint(g, s.X, s.Y, dir, staticDef, false)
		} else if s.Removed {
			// No static loc here — a runtime-spawned object (a lit fire)
			// being removed again. We no longer know its def, so clear the
			// tile's full-block bits (spawned blockers are type-1 stamps).
			if gx, gy, in := g.World2Local(s.X, s.Y); in {
				g.Mask[gx][gy] &^= (FullBlockC | Object)
			}
		}
		if !s.Removed {
			stampSceneryFootprint(g, s.X, s.Y, dir, f.SceneryDef(s.ID), true)
		}
	}
}

// setOrClearEdge sets (block) or clears (!block) the wall bit(s) for the edge
// of tile (gx,gy) named by a boundary Direction, plus the mirrored bit on the
// adjacent tile — the exact edge geometry applyBoundaries uses. Clearing is how
// an OPENED door becomes passable; a door and a structural wall never share an
// edge, so clearing the door's edge cannot accidentally open a real wall.
func setOrClearEdge(g *Grid, gx, gy, dir int, block bool) {
	set := func(x, y, bit int) {
		if x < 0 || x >= GridSize || y < 0 || y >= GridSize {
			return
		}
		if block {
			g.Mask[x][y] |= bit
		} else {
			g.Mask[x][y] &^= bit
		}
	}
	switch dir {
	case 0: // north edge of (gx,gy); mirror south on the north neighbor
		set(gx, gy, WallNorth)
		set(gx, gy-1, WallSouth)
	case 1: // west edge (WallEast bit); mirror west on the west neighbor
		set(gx, gy, WallEast)
		set(gx-1, gy, WallWest)
	case 2: // diagonal NW-SE
		set(gx, gy, FullBlockA)
	case 3: // diagonal SW-NE
		set(gx, gy, FullBlockB)
	}
}
