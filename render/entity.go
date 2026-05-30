package render

// EntityKind distinguishes the billboard's colour/size by who it is.
type EntityKind int

const (
	EntityNPC EntityKind = iota
	EntityPlayer
	EntitySelf // the host itself (rarely drawn — camera sits on it)
)

// Entity is one billboarded actor the host perceives, in ABSOLUTE world-tile
// coords (the same space facts.NpcLocs and world.NpcRecord/PlayerRecord use —
// the live decoder converts relative offsets to absolute). The renderer turns
// each into a depth-scaled standing figure.
type Entity struct {
	X, Y int
	Kind EntityKind
}

// entity billboard dimensions (world units). RSC actors stand ~2 tiles tall;
// 128 units == one tile. A figure ~1.6 tiles tall and ~0.6 tiles wide reads as
// a person at the renderer's scale without looking like a wall.
const (
	entityHeight   = 200
	entityHalfWide = 38
)

// entity fill colours (method305-encoded flat 5:5:5). Players read as a warm
// tunic blue; NPCs as a muted brown/tan so the two are distinguishable at a
// glance. Flat (non-gouraud) shaded so a vertical billboard never gouraud-darkens
// to black like the early terrain cliffs did.
var (
	entityPlayerColour = method305(70, 90, 200)  // blue tunic
	entityNPCColour     = method305(170, 140, 95) // tan/brown
	entitySelfColour    = method305(220, 200, 60) // gold (host marker)
)

func entityColour(k EntityKind) int32 {
	switch k {
	case EntityPlayer:
		return entityPlayerColour
	case EntitySelf:
		return entitySelfColour
	default:
		return entityNPCColour
	}
}

// BuildEntities assembles one camera-agnostic "cross" billboard per in-window
// entity into a single GameModel: two perpendicular vertical quads (an X in
// plan view) rising entityHeight world-units off the terrain at the entity's
// tile centre. The cross means the figure stays visible (and roughly the same
// silhouette) from any camera yaw — a single quad would vanish edge-on. The
// quads are flat-shaded and fed through the same painter-sorted pipeline as
// scenery, so they occlude correctly behind walls/scenery and depth-scale with
// the perspective divide. baseX/baseY anchor the 96x96 window; heights is the
// flattened terrain grid (TerrainHeights) so each figure's feet sit on the
// ground. Returns nil if no entity falls inside the window.
func BuildEntities(ents []Entity, baseX, baseY, plane int, heights [][]int32) *GameModel {
	n := terrainSize

	type placed struct {
		lx, ly int
		colour int32
	}
	var ps []placed
	for _, e := range ents {
		lx := e.X - baseX
		ly := e.Y - baseY
		if lx < 0 || lx >= n || ly < 0 || ly >= n {
			continue
		}
		ps = append(ps, placed{lx, ly, entityColour(e.Kind)})
	}
	if len(ps) == 0 {
		return nil
	}

	// 2 quads per entity, 4 verts each.
	g := NewGameModel(len(ps)*8, len(ps)*2)
	for _, p := range ps {
		// tile centre in world units; feet at the terrain height there.
		cx := int32(p.lx)*128 + 64
		cz := int32(p.ly)*128 + 64
		foot := -heights[p.lx][p.ly]
		top := foot - entityHeight

		// quad A spans along +X (east-west facing pair)
		ax0 := g.AddVertex(cx-entityHalfWide, foot, cz)
		ax1 := g.AddVertex(cx-entityHalfWide, top, cz)
		ax2 := g.AddVertex(cx+entityHalfWide, top, cz)
		ax3 := g.AddVertex(cx+entityHalfWide, foot, cz)
		// quad B spans along +Z (north-south facing pair), perpendicular
		bz0 := g.AddVertex(cx, foot, cz-entityHalfWide)
		bz1 := g.AddVertex(cx, top, cz-entityHalfWide)
		bz2 := g.AddVertex(cx, top, cz+entityHalfWide)
		bz3 := g.AddVertex(cx, foot, cz+entityHalfWide)

		// flat-shaded, both faces filled so the billboard is opaque from
		// either side (intensity 0 -> the flat base shade, no gouraud-darkening).
		g.AddFace([]int{ax0, ax1, ax2, ax3}, p.colour, p.colour, 0)
		g.AddFace([]int{bz0, bz1, bz2, bz3}, p.colour, p.colour, 0)
	}
	// Bright, near-flat ambient so figures read as solid colour from any angle
	// (same lighting recipe as the boundary walls).
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
