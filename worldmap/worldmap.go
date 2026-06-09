// Package worldmap is the in-memory cradle WorldOracle engine: static
// geography reasoning a host queries without touching the live server.
//
// It precomputes, once and shared by pointer, the global plane-0
// walkability of the loaded Authentic_Landscape footprint (world tiles
// 0..1007 on each axis) and flood-fills it into 4-connected walking
// components. Two tiles share a component IFF a host can WALK between
// them with no game-logic gate in the way — i.e. the same connectivity
// the real client pathfinder (pathfind.FindPath) computes locally, but
// extended globally.
//
// SCOPE (v1, locked): PLANE 0 only, the static world only, the
// Authentic_Landscape archive footprint only. Upper floors, dungeons,
// dynamic/quest state, and the transport (boat/toll/ladder) layer are
// OUT of scope here — see the design notes; the transport layer is a
// separate later step.
//
// THE LOAD-BEARING CORRECTION baked into this engine: the Al-Kharid
// toll gate is NOT a collision barrier. The gate boundary def is
// openable in the collision map (BlocksMovement()==false), so both this
// flood-fill AND the real pathfinder walk straight across it. Therefore
// Lumbridge (135,654) and the Al-Kharid mine (74,583) are in the SAME
// walking component. The 10-coin toll lives entirely in server
// game-logic (BorderGuard) and must be modeled as a CONDITIONAL
// transport edge by a later layer, never as a flood boundary. This
// engine deliberately refuses to invent a barrier the engine does not
// have, which is exactly why it can be trusted.
//
// This package is pure in-memory and depends only on facts + pathfind.
// It adds/modifies no proto, touches no mesa code, and is not wired
// into cradle/runtime/dsl yet.
package worldmap

import (
	"fmt"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// World extent of the loaded plane-0 footprint. The Authentic_Landscape
// archive stores plane-0 sectors at archive coords sx 48..68, sy 37..57
// (21x21 sectors), which maps to world tiles X 0..1007, Y 0..1007.
const (
	// planeSectors is the side length of the plane-0 sector grid (21).
	planeSectors = 21
	// WorldDim is the side length of the plane-0 world tile grid (1008).
	WorldDim = planeSectors * pathfind.SectorSize

	// Plane-0 archive sector range, mirrored from pathfind.landscape.go
	// (archiveOffsetX=48, archiveOffsetY=37). World origin of an
	// archive sector (asx,asy) is ((asx-48)*48, (asy-37)*48).
	archiveOriginSX = 48
	archiveOriginSY = 37
)

// Collision-flag bits. These MUST match pathfind/flags.go exactly: the
// grid builders we mirror write these bits, and the flood-fill movement
// predicate reads them the same way pathfind/bfs.go does. They are
// re-declared here (rather than imported) because pathfind keeps them
// unexported; the integer values are pinned to the ported RSC client.
const (
	wallNorth = 1
	wallEast  = 2
	wallSouth = 4
	wallWest  = 8

	fullBlockA = 16
	fullBlockB = 32
	fullBlockC = 64
	fullBlock  = fullBlockA | fullBlockB | fullBlockC

	// object marks a tile occupied by a solid scenery object. Like
	// pathfind, the movement predicate does not read it; we set it so a
	// POI sitting on a building footprint can be snapped off it.
	object = 128

	// *Blocked combine the full-block bits with the wall edge being
	// CROSSED, so a single AND answers "can I step across this edge".
	// Named after the wall bit (matching CollisionFlag.java), NOT the
	// movement direction: stepping NORTH (to y-1) crosses the
	// destination tile's SOUTH edge, so it checks southBlocked.
	westBlocked  = fullBlock | wallWest
	southBlocked = fullBlock | wallSouth
	northBlocked = fullBlock | wallNorth
	eastBlocked  = fullBlock | wallEast

	// Corner-blocked combos for diagonal steps: a diagonal move also
	// requires the diagonal tile itself not be full- or corner-blocked
	// on the two edges it touches. Mirrors pathfind/flags.go.
	wallNorthEast = wallNorth | wallEast
	wallNorthWest = wallNorth | wallWest
	wallSouthEast = wallSouth | wallEast
	wallSouthWest = wallSouth | wallWest

	southEastBlocked = fullBlock | wallSouthEast
	southWestBlocked = fullBlock | wallSouthWest
	northEastBlocked = fullBlock | wallNorthEast
	northWestBlocked = fullBlock | wallNorthWest
)

// Component is metadata for one connected walking component.
type Component struct {
	ID int32 // == its index in Oracle.comps

	Tiles int // number of standable tiles in this component

	// Bounding box in world tile coords (inclusive).
	MinX, MinY, MaxX, MaxY int

	// A representative standable tile inside the component (the first
	// tile the flood-fill seeded from).
	RepX, RepY int
}

// IndexedPOI is a gazetteer POI bound to the walking component of its
// nearest standable tile.
type IndexedPOI struct {
	facts.POI

	// SnapX, SnapY is the nearest standable tile to the POI marker
	// (POI markers commonly sit on a building footprint / wall / water).
	SnapX, SnapY int

	// Comp is the component of (SnapX, SnapY), or -1 if the POI could
	// not be snapped to any standable tile nearby.
	Comp int32
}

// Oracle is the precomputed, read-only, shared static-geography engine.
// Plane 0 only in v1. It is immutable after Precompute, so a single
// instance can be shared by pointer across all hosts with no locking.
type Oracle struct {
	// dim is the side length of the square plane-0 tile grid (WorldDim).
	dim int

	// mask holds the collision bits per tile, indexed mask[x*dim+y].
	// Kept after the build so CompAt's snap and POI indexing can test
	// standability without rebuilding.
	mask []int

	// labels holds the component id per tile, indexed labels[x*dim+y].
	// -1 = void / non-standable; otherwise an index into comps.
	labels []int32

	// comps is the per-component metadata, indexed by component id.
	comps []Component

	// pois is every gazetteer POI snapped + bound to its component.
	pois []IndexedPOI

	// gaz is the shared gazetteer (reused, not copied).
	gaz *facts.Gazetteer

	// edges is the standing transport layer (ferries, toll gates, guild
	// doors, post-quest links) loaded from the embedded transport.json.
	// These overlay game-logic CAPABILITY gates on the pure-collision
	// components; see transport.go.
	edges []TransportEdge

	// edgesSkipped counts passage records that did not resolve into a
	// standing edge (unresolved loc/bound id, one-time-quest, lever/puzzle,
	// etc.) — reported but never invented.
	edgesSkipped int
}

// inBounds reports whether (x,y) is inside the plane-0 tile grid.
func (o *Oracle) inBounds(x, y int) bool {
	return x >= 0 && x < o.dim && y >= 0 && y < o.dim
}

func (o *Oracle) at(x, y int) int { return o.mask[x*o.dim+y] }

// standable reports whether a host can occupy the tile at (x,y): in
// bounds and not fully blocked by terrain or solid scenery. This is the
// same rule pathfind uses (mask&fullBlock==0); edge-wall flags gate
// crossing an edge, not standing on the tile.
func (o *Oracle) standable(x, y int) bool {
	if !o.inBounds(x, y) {
		return false
	}
	return o.at(x, y)&fullBlock == 0
}

// Dim returns the side length of the plane-0 tile grid (WorldDim).
func (o *Oracle) Dim() int { return o.dim }

// Components returns the per-component metadata slice (indexed by id).
// The returned slice must not be mutated.
func (o *Oracle) Components() []Component { return o.comps }

// NumComponents returns the number of connected walking components.
func (o *Oracle) NumComponents() int { return len(o.comps) }

// CompAt returns the component id of the tile at (x,y), or -1 if the
// tile is out of bounds, void, or non-standable. Use CompNear to snap a
// non-standable POI/building tile to its nearest standable component.
func (o *Oracle) CompAt(x, y int) int32 {
	if !o.inBounds(x, y) {
		return -1
	}
	return o.labels[x*o.dim+y]
}

// CompNear returns the component of (x,y) if it is standable, otherwise
// it snaps to the nearest standable tile (by an expanding ring search,
// up to radius maxSnap) and returns that tile's component plus the
// snapped coords. Returns (-1, x, y, false) if nothing standable is
// found within the radius. POIs sit on building footprints, so callers
// reading a POI's component should go through this.
func (o *Oracle) CompNear(x, y int) (comp int32, sx, sy int, ok bool) {
	if c := o.CompAt(x, y); c >= 0 {
		return c, x, y, true
	}
	const maxSnap = 6
	for r := 1; r <= maxSnap; r++ {
		// Walk the ring at Chebyshev radius r, nearest-first by ring.
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				// only the ring border, not the filled interior
				if dx > -r && dx < r && dy > -r && dy < r {
					continue
				}
				nx, ny := x+dx, y+dy
				if c := o.CompAt(nx, ny); c >= 0 {
					return c, nx, ny, true
				}
			}
		}
	}
	return -1, x, y, false
}

// POIs returns every gazetteer POI bound to its walking component. The
// returned slice must not be mutated.
func (o *Oracle) POIs() []IndexedPOI { return o.pois }

// POIsInComponents returns the indexed POIs (optionally filtered by a
// case-insensitive type substring, like the gazetteer) whose snapped
// component is in the given set. A nil/empty set matches no components.
// An empty typ matches every type.
func (o *Oracle) POIsInComponents(typ string, set map[int32]bool) []IndexedPOI {
	var out []IndexedPOI
	for _, p := range o.pois {
		if p.Comp < 0 || !set[p.Comp] {
			continue
		}
		if typ != "" && !containsFold(p.Type, typ) {
			continue
		}
		out = append(out, p)
	}
	return out
}

// Gazetteer returns the shared gazetteer the oracle indexed against.
func (o *Oracle) Gazetteer() *facts.Gazetteer { return o.gaz }

// Precompute builds the plane-0 WorldOracle from the loaded facts and
// landscape archive. It (1) builds the global walkability mask by
// mirroring pathfind's three grid builders over every plane-0 sector,
// (2) flood-fills it into 4-connected walking components using the same
// cardinal-movement predicate as pathfind.FindPath, and (3) snaps +
// indexes every gazetteer POI to its component.
//
// It is pure CPU + memory; the caller still owns the landscape's fd.
func Precompute(f *facts.Facts, ls *pathfind.Landscape) (*Oracle, error) {
	if f == nil {
		return nil, fmt.Errorf("worldmap: nil facts")
	}
	if ls == nil {
		return nil, fmt.Errorf("worldmap: nil landscape")
	}

	o := &Oracle{
		dim: WorldDim,
		gaz: f.Gazetteer(),
	}
	o.mask = make([]int, o.dim*o.dim)

	if err := o.buildMask(f, ls); err != nil {
		return nil, err
	}
	o.floodFill()
	o.indexPOIs()
	if err := o.loadTransport(); err != nil {
		return nil, err
	}
	return o, nil
}

// buildMask fills o.mask for the whole plane-0 footprint by mirroring
// pathfind's applySectorToGrid / applyScenery / applyBoundaries, but
// over global world coords instead of a 96x96 local grid.
func (o *Oracle) buildMask(f *facts.Facts, ls *pathfind.Landscape) error {
	// (1) Landscape walls + water, per sector. Mirror of
	// pathfind.applySectorToGrid, written into global tile coords.
	for asx := archiveOriginSX; asx < archiveOriginSX+planeSectors; asx++ {
		for asy := archiveOriginSY; asy < archiveOriginSY+planeSectors; asy++ {
			key := pathfind.SectorKey{Plane: 0, SX: asx, SY: asy}
			s, err := ls.Sector(key)
			if err != nil {
				return fmt.Errorf("worldmap: load sector %+v: %w", key, err)
			}
			if s == nil {
				continue // void area — leave tiles unblocked-but-unstood-on
			}
			originX := (asx - archiveOriginSX) * pathfind.SectorSize
			originY := (asy - archiveOriginSY) * pathfind.SectorSize
			o.applySector(f, s, originX, originY)
		}
	}

	// (2) + (3) Scenery and boundary locs. These iterate ALL locs once
	// globally (cheaper than the per-window re-scan BuildGrid does), and
	// are mirrors of pathfind.applyScenery / pathfind.applyBoundaries.
	o.applyScenery(f)
	o.applyBoundaries(f)
	return nil
}

// applySector mirrors pathfind.applySectorToGrid over global coords.
func (o *Oracle) applySector(f *facts.Facts, s *pathfind.Sector, originX, originY int) {
	for sx := 0; sx < pathfind.SectorSize; sx++ {
		for sy := 0; sy < pathfind.SectorSize; sy++ {
			wx := originX + sx
			wy := originY + sy
			if !o.inBounds(wx, wy) {
				continue
			}
			tile := s.At(sx, sy)

			// verticalWall lives on this tile's NORTH edge; the server
			// also mirrors WALL_SOUTH onto the north neighbor.
			if vw := int(tile.VerticalWall); vw > 0 {
				if d, ok := f.BoundaryDefs[vw-1]; ok && d.BlocksMovement() {
					o.or(wx, wy, wallNorth)
					o.or(wx, wy-1, wallSouth)
				}
			}

			// horizontalWall lives on this tile's WEST edge (named
			// WALL_EAST in the bit layout); mirror WALL_WEST onto the
			// west neighbor.
			if hw := int(tile.HorizontalWall); hw > 0 {
				if d, ok := f.BoundaryDefs[hw-1]; ok && d.BlocksMovement() {
					o.or(wx, wy, wallEast)
					o.or(wx-1, wy, wallWest)
				}
			}

			// Diagonal walls fully block their tile.
			if dw := int(tile.DiagonalWalls); dw > 0 {
				idx := dw - 1
				if dw > 12000 && dw < 24000 {
					idx = dw - 12001
				}
				if d, ok := f.BoundaryDefs[idx]; ok && d.BlocksMovement() {
					if dw > 0 && dw < 12000 {
						o.or(wx, wy, fullBlockB)
					} else if dw > 12000 && dw < 24000 {
						o.or(wx, wy, fullBlockA)
					}
				}
			}

			// Ground-overlay terrain collision per TileDef.xml (1-based
			// index, 250→2 remap, ObjectType!=0 blocks) — same rule as
			// pathfind.applySectorToGrid so the two engines agree.
			if f.OverlayBlocks(int(tile.GroundOverlay)) {
				o.or(wx, wy, fullBlockC)
			}
		}
	}
}

// applyScenery mirrors pathfind.applyScenery / stampSceneryFootprint over
// global coords: the dir-swapped width×height footprint, with type-1 full
// blocks and type-2 directional edges stamped on EVERY footprint tile (the
// missing footprint loop was the phantom-gap bug: multi-tile gates and double
// doors had unmodeled halves), plus the server/Plutonium id-1147 exclusion.
func (o *Oracle) applyScenery(f *facts.Facts) {
	const sceneryCollisionSkipID = 1147
	for _, loc := range f.SceneryLocs {
		if !o.inBounds(loc.X, loc.Y) {
			continue
		}
		def, has := f.SceneryDefs[loc.DefID]
		if !has || def == nil || def.ID == sceneryCollisionSkipID {
			continue
		}
		if def.Type != 1 && def.Type != 2 {
			continue // type 0 / 3+ -> non-blocking decoration
		}
		// Openable doors/gates do NOT cut the Oracle's walkability
		// partition — same semantic as applyBoundaries' openable-door
		// rule below: the Oracle answers "reachable with effort" (the
		// host can open them en route), while the MOVEMENT grid keeps
		// them blocked until actually opened. Gated routes that demand
		// more than a click (the Al-Kharid toll) are re-cut by the
		// transport layer's Barrier with gate+needs metadata.
		if def.IsOpenableBarrier() {
			continue
		}
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
				tx, ty := loc.X+dx, loc.Y+dy
				switch def.Type {
				case 1:
					o.or(tx, ty, fullBlockC|object)
				case 2:
					switch loc.Direction {
					case 0:
						o.or(tx, ty, wallEast)
						o.or(tx-1, ty, wallWest)
					case 2:
						o.or(tx, ty, wallSouth)
						o.or(tx, ty+1, wallNorth)
					case 4:
						o.or(tx, ty, wallWest)
						o.or(tx+1, ty, wallEast)
					case 6:
						o.or(tx, ty, wallNorth)
						o.or(tx, ty-1, wallSouth)
					}
				}
			}
		}
	}
}

// applyBoundaries mirrors pathfind.applyBoundaries over global coords.
// Only defs whose BlocksMovement() is true block — this is THE rule
// (DoorType!=0 && Unknown==0) that makes an openable door/gate (such as
// the Al-Kharid toll gate) stay walkable rather than cut a component.
func (o *Oracle) applyBoundaries(f *facts.Facts) {
	for _, loc := range f.BoundaryLocs {
		def, ok := f.BoundaryDefs[loc.DefID]
		if !ok || !def.BlocksMovement() {
			continue
		}
		if !o.inBounds(loc.X, loc.Y) {
			continue
		}
		switch loc.Direction {
		case 0:
			o.or(loc.X, loc.Y, wallNorth)
			o.or(loc.X, loc.Y-1, wallSouth)
		case 1:
			o.or(loc.X, loc.Y, wallEast)
			o.or(loc.X-1, loc.Y, wallWest)
		case 2:
			o.or(loc.X, loc.Y, fullBlockA)
		case 3:
			o.or(loc.X, loc.Y, fullBlockB)
		}
	}
}

// or sets bits on the tile at (x,y) if it is in bounds (the edge-mirror
// onto a neighbor can fall outside the footprint at the border).
func (o *Oracle) or(x, y, bits int) {
	if !o.inBounds(x, y) {
		return
	}
	o.mask[x*o.dim+y] |= bits
}

// floodFill labels every standable tile with a connected-component id,
// using the SAME movement predicate as pathfind.FindPath so the model
// cannot drift from the real pathfinder. A CARDINAL step is allowed iff
// the destination's crossed wall edge is clear; a DIAGONAL step is
// allowed iff both intermediate cardinal edges are clear AND the
// diagonal tile's matching corner is not full/corner-blocked (exactly
// pathfind/bfs.go's cardinal + diagonal expansion). Including diagonals
// matters: RSC has diagonal "squeeze" geometry (walls forming a corner
// staircase) where two tiles are joined ONLY by a corner-cut; 4-conn
// alone would split them into separate components while FindPath walks
// the diagonal, so a strict engine-agreement contract requires matching
// FindPath's connectivity here.
func (o *Oracle) floodFill() {
	dim := o.dim
	o.labels = make([]int32, dim*dim)
	for i := range o.labels {
		o.labels[i] = -1
	}

	// BFS queue reused across seeds.
	queue := make([]int, 0, 4096)

	var next int32 // next component id
	for sx := 0; sx < dim; sx++ {
		for sy := 0; sy < dim; sy++ {
			start := sx*dim + sy
			if o.labels[start] != -1 || o.mask[start]&fullBlock != 0 {
				continue // already labeled or not standable
			}

			id := next
			next++
			c := Component{
				ID:   id,
				MinX: sx, MinY: sy, MaxX: sx, MaxY: sy,
				RepX: sx, RepY: sy,
			}

			o.labels[start] = id
			queue = queue[:0]
			queue = append(queue, start)
			for qi := 0; qi < len(queue); qi++ {
				idx := queue[qi]
				cx := idx / dim
				cy := idx % dim

				c.Tiles++
				if cx < c.MinX {
					c.MinX = cx
				}
				if cx > c.MaxX {
					c.MaxX = cx
				}
				if cy < c.MinY {
					c.MinY = cy
				}
				if cy > c.MaxY {
					c.MaxY = cy
				}

				// North: step to (cx, cy-1), crossing that tile's south
				// edge -> check southBlocked on the DESTINATION. Mirrors
				// pathfind/bfs.go's cardinal expansion exactly.
				if cy > 0 {
					ni := cx*dim + (cy - 1)
					if o.labels[ni] == -1 && o.mask[ni]&southBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// South: step to (cx, cy+1), crossing its north edge.
				if cy < dim-1 {
					ni := cx*dim + (cy + 1)
					if o.labels[ni] == -1 && o.mask[ni]&northBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// West: step to (cx-1, cy). The mask mirror convention
				// records a west-edge wall as WallWest ON the west tile
				// itself, so FindPath checks westBlocked on the
				// DESTINATION (NOT eastBlocked) — see pathfind/bfs.go:89.
				if cx > 0 {
					ni := (cx-1)*dim + cy
					if o.labels[ni] == -1 && o.mask[ni]&westBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// East: step to (cx+1, cy). Symmetrically, FindPath
				// checks eastBlocked on the destination (pathfind/bfs.go:96).
				if cx < dim-1 {
					ni := (cx+1)*dim + cy
					if o.labels[ni] == -1 && o.mask[ni]&eastBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}

				// Diagonal expansion (mirror of pathfind/bfs.go): a
				// diagonal step requires BOTH adjacent cardinal edges to
				// be open AND the diagonal tile's matching corner not
				// blocked. North uses southBlocked (we cross the dest's
				// south edge), etc.
				// North-West to (cx-1, cy-1).
				if cx > 0 && cy > 0 &&
					o.mask[cx*dim+(cy-1)]&southBlocked == 0 &&
					o.mask[(cx-1)*dim+cy]&westBlocked == 0 {
					ni := (cx-1)*dim + (cy - 1)
					if o.labels[ni] == -1 && o.mask[ni]&southWestBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// North-East to (cx+1, cy-1).
				if cx < dim-1 && cy > 0 &&
					o.mask[cx*dim+(cy-1)]&southBlocked == 0 &&
					o.mask[(cx+1)*dim+cy]&eastBlocked == 0 {
					ni := (cx+1)*dim + (cy - 1)
					if o.labels[ni] == -1 && o.mask[ni]&southEastBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// South-West to (cx-1, cy+1).
				if cx > 0 && cy < dim-1 &&
					o.mask[cx*dim+(cy+1)]&northBlocked == 0 &&
					o.mask[(cx-1)*dim+cy]&westBlocked == 0 {
					ni := (cx-1)*dim + (cy + 1)
					if o.labels[ni] == -1 && o.mask[ni]&northWestBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
				// South-East to (cx+1, cy+1).
				if cx < dim-1 && cy < dim-1 &&
					o.mask[cx*dim+(cy+1)]&northBlocked == 0 &&
					o.mask[(cx+1)*dim+cy]&eastBlocked == 0 {
					ni := (cx+1)*dim + (cy + 1)
					if o.labels[ni] == -1 && o.mask[ni]&northEastBlocked == 0 {
						o.labels[ni] = id
						queue = append(queue, ni)
					}
				}
			}

			o.comps = append(o.comps, c)
		}
	}
}

// indexPOIs snaps every gazetteer POI to its nearest standable tile and
// records the component there. POIs that cannot be snapped get Comp=-1.
func (o *Oracle) indexPOIs() {
	if o.gaz == nil {
		return
	}
	o.pois = make([]IndexedPOI, 0, len(o.gaz.POIs))
	for _, p := range o.gaz.POIs {
		ip := IndexedPOI{POI: p, SnapX: p.X, SnapY: p.Y, Comp: -1}
		if comp, sx, sy, ok := o.CompNear(p.X, p.Y); ok {
			ip.Comp = comp
			ip.SnapX = sx
			ip.SnapY = sy
		}
		o.pois = append(o.pois, ip)
	}
}
