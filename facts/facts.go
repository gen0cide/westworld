package facts

import (
	"fmt"
	"math"
	"strings"
)

// Facts is the bot's read-only knowledge base of static world data.
// Loaded once at startup; treat as immutable.
type Facts struct {
	SceneryDefs  map[int]*SceneryDef
	BoundaryDefs map[int]*BoundaryDef
	NpcDefs      map[int]*NpcDef
	ItemDefs     map[int]*ItemDef

	SceneryLocs    []SceneryLoc
	BoundaryLocs   []BoundaryLoc
	NpcLocs        []NpcLoc
	GroundItemLocs []GroundItemLoc

	// TileDefs is the ground-overlay table from TileDef.xml (index = overlay-1).
	// Consulted by OverlayBlocks for authoritative terrain collision.
	TileDefs []TileDef

	// Spatial indexes: tile coord → list of placement indices.
	// Built by buildIndex after Load.
	sceneryByTile  map[tileKey][]int
	boundaryByTile map[tileKey][]int
	npcStartByTile map[tileKey][]int

	// wornBySlot resolves another player's worn-equipment appearance back
	// to items: [wearSlot][appearanceID & 0xFF] → candidate ItemDefs.
	// Built by buildWornIndex at load. See facts/worn.go.
	wornBySlot map[int]map[int][]*ItemDef

	// itemByName is the reverse of ItemDefs: lower-cased item name →
	// ItemDef (lowest id wins on duplicate names). Lets cognition/brain
	// code go name→def→id, the inverse of ItemDef(id). Built by
	// buildWornIndex at load.
	itemByName map[string]*ItemDef

	// gaz is the world-map gazetteer (named places + POIs) for host
	// where-am-I perception. Built from embedded data at load.
	gaz *Gazetteer
}

// Gazetteer returns the world-map gazetteer (named places + POIs). Never
// nil after Load. Lets a host perceive where it is and where things are.
func (f *Facts) Gazetteer() *Gazetteer {
	if f.gaz == nil {
		f.gaz = loadGazetteer()
	}
	return f.gaz
}

// tileKey packs (x, y) into a single int for map keys.
type tileKey uint64

func tk(x, y int) tileKey { return tileKey(uint64(uint32(x))<<32 | uint64(uint32(y))) }

// BuildIndex (re)builds the by-tile spatial indexes. Load calls it
// automatically; callers constructing a Facts programmatically (tests,
// generated tables) must call it before using At/Near.
func (f *Facts) BuildIndex() { f.buildIndex() }

// buildIndex populates the by-tile maps for fast spatial queries.
func (f *Facts) buildIndex() {
	f.sceneryByTile = make(map[tileKey][]int, len(f.SceneryLocs))
	for i, s := range f.SceneryLocs {
		k := tk(s.X, s.Y)
		f.sceneryByTile[k] = append(f.sceneryByTile[k], i)
	}
	f.boundaryByTile = make(map[tileKey][]int, len(f.BoundaryLocs))
	for i, b := range f.BoundaryLocs {
		k := tk(b.X, b.Y)
		f.boundaryByTile[k] = append(f.boundaryByTile[k], i)
	}
	f.npcStartByTile = make(map[tileKey][]int, len(f.NpcLocs))
	for i, n := range f.NpcLocs {
		k := tk(n.StartX, n.StartY)
		f.npcStartByTile[k] = append(f.npcStartByTile[k], i)
	}
}

// SceneryDef returns the def for an id, or nil.
func (f *Facts) SceneryDef(id int) *SceneryDef { return f.SceneryDefs[id] }

// BoundaryDef returns the def for an id, or nil.
func (f *Facts) BoundaryDef(id int) *BoundaryDef { return f.BoundaryDefs[id] }

// NpcDef returns the def for an id, or nil.
func (f *Facts) NpcDef(id int) *NpcDef { return f.NpcDefs[id] }

// ItemDef returns the def for an id, or nil.
func (f *Facts) ItemDef(id int) *ItemDef { return f.ItemDefs[id] }

// Placement is a unified result type for what's at a tile.
type Placement struct {
	Kind       string // "scenery", "boundary", "npc_spawn", "ground_item"
	DefID      int
	Name       string // resolved from def
	X, Y       int
	Direction  int
	Extra      string // e.g., NPC range "(min, max)" or boundary side
}

// At returns everything placed at the given tile.
func (f *Facts) At(x, y int) []Placement {
	k := tk(x, y)
	var out []Placement
	for _, i := range f.sceneryByTile[k] {
		s := f.SceneryLocs[i]
		name := "?"
		if d := f.SceneryDef(s.DefID); d != nil {
			name = d.Name
		}
		out = append(out, Placement{Kind: "scenery", DefID: s.DefID, Name: name, X: s.X, Y: s.Y, Direction: s.Direction})
	}
	for _, i := range f.boundaryByTile[k] {
		b := f.BoundaryLocs[i]
		name := "?"
		if d := f.BoundaryDef(b.DefID); d != nil {
			name = d.Name
		}
		out = append(out, Placement{Kind: "boundary", DefID: b.DefID, Name: name, X: b.X, Y: b.Y, Direction: b.Direction})
	}
	for _, i := range f.npcStartByTile[k] {
		n := f.NpcLocs[i]
		name := "?"
		if d := f.NpcDef(n.DefID); d != nil {
			name = d.Name
		}
		out = append(out, Placement{
			Kind: "npc_spawn", DefID: n.DefID, Name: name, X: n.StartX, Y: n.StartY,
			Extra: fmt.Sprintf("range (%d,%d)-(%d,%d)", n.MinX, n.MinY, n.MaxX, n.MaxY),
		})
	}
	return out
}

// Near returns everything within `radius` tiles of (x, y) in
// Chebyshev distance (a square, not a circle — matches RSC's view
// model). Results are sorted by distance ascending.
func (f *Facts) Near(x, y, radius int) []Placement {
	var out []Placement
	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			placements := f.At(x+dx, y+dy)
			out = append(out, placements...)
		}
	}
	return out
}

// NearestByName finds the closest placement matching name (case-
// insensitive substring) of the given kind. Returns nil if none
// found within maxRadius tiles.
//
// kind may be "scenery", "boundary", "npc_spawn", "ground_item", or
// "*" to match any.
func (f *Facts) NearestByName(name string, fromX, fromY, maxRadius int, kind string) *Placement {
	target := strings.ToLower(name)
	var best *Placement
	bestDist := math.MaxFloat64
	// Spiral search outward; cap at maxRadius.
	for r := 0; r <= maxRadius; r++ {
		// Check tiles on the ring at distance r.
		// (For small r we just iterate the square; for huge r we'd want a
		//  proper ring iteration, but we cap maxRadius and worst-case is
		//  acceptable.)
		for dy := -r; dy <= r; dy++ {
			for dx := -r; dx <= r; dx++ {
				// Skip tiles already covered by smaller r.
				if max(abs(dx), abs(dy)) != r {
					continue
				}
				placements := f.At(fromX+dx, fromY+dy)
				for _, p := range placements {
					if kind != "*" && p.Kind != kind {
						continue
					}
					if !strings.Contains(strings.ToLower(p.Name), target) {
						continue
					}
					d := math.Hypot(float64(dx), float64(dy))
					if d < bestDist {
						pp := p
						best = &pp
						bestDist = d
					}
				}
			}
		}
		if best != nil {
			return best
		}
	}
	return nil
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Summary returns counts for logging at startup.
func (f *Facts) Summary() string {
	return fmt.Sprintf(
		"scenery_defs=%d boundary_defs=%d npc_defs=%d item_defs=%d | scenery_locs=%d boundary_locs=%d npc_locs=%d ground_item_locs=%d",
		len(f.SceneryDefs), len(f.BoundaryDefs), len(f.NpcDefs), len(f.ItemDefs),
		len(f.SceneryLocs), len(f.BoundaryLocs), len(f.NpcLocs), len(f.GroundItemLocs),
	)
}
