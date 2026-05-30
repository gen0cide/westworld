package render

import (
	"github.com/gen0cide/westworld/facts"
)

// wallColour is the flat fill for an opaque wall face. Real RSC textures the
// wall front/back from wallObjectSomething2/wallObjectTexture; until the
// texture archive is wired (priority 4) we paint a neutral stone grey so walls
// read as solid surfaces (and so windows/doorframes still occlude correctly).
var (
	wallColourStone = method305(150, 150, 150)
	wallColourWood  = method305(120, 90, 55)
)

// BuildBoundaries assembles all wall/fence/door boundary quads inside the
// 96x96 window into a single GameModel, porting World.method422: each boundary
// is a vertical quad rising `Height` world-units above the terrain, spanning
// the tile edge picked by its direction.
//
//	dir 0 -> east-west wall  : edge (x,y)..(x+1,y)        (RSC getWallEastwest)
//	dir 1 -> north-south wall: edge (x,y)..(x,y+1)        (RSC getWallNorthsouth)
//	dir 2 -> diagonal '\'    : edge (x,y)..(x+1,y+1)      (RSC getWallDiagonal <12000)
//	dir 3 -> diagonal '/'    : edge (x+1,y)..(x,y+1)      (RSC getWallDiagonal >12000)
//
// heights is the flattened terrain-height grid (TerrainHeights) so wall feet
// sit flush on the ground exactly like the terrain mesh. Returns nil if no
// boundary falls inside the window.
func BuildBoundaries(f *facts.Facts, baseX, baseY, plane int, heights [][]int32) *GameModel {
	if f == nil {
		return nil
	}
	n := terrainSize

	// First pass: collect the in-window quads so we can size the model.
	type quad struct {
		x0, y0, x1, y1 int // local tile endpoints
		h              int32
		colour         int32
	}
	var quads []quad
	for _, b := range f.BoundaryLocs {
		lx := b.X - baseX
		ly := b.Y - baseY
		// the wall touches one extra tile to the +x/+y; require the whole
		// edge inside [0, n-1] so both endpoints have a terrain height.
		if lx < 0 || lx >= n-1 || ly < 0 || ly >= n-1 {
			continue
		}
		def := f.BoundaryDef(b.DefID)
		if def == nil {
			continue
		}
		h := int32(def.Height)
		if h <= 0 {
			h = 192 // RSC default full-wall height
		}
		// wood for fences/gates, stone otherwise (cheap heuristic on the name).
		colour := wallColourStone
		switch def.Name {
		case "Fence", "Gate", "Railing", "Wooden fence", "Iron railings":
			colour = wallColourWood
		}

		var x0, y0, x1, y1 int
		switch b.Direction {
		case 0: // east-west: (x,y)..(x+1,y)
			x0, y0, x1, y1 = lx, ly, lx+1, ly
		case 1: // north-south: (x,y)..(x,y+1)
			x0, y0, x1, y1 = lx, ly, lx, ly+1
		case 2: // diagonal '\': (x,y)..(x+1,y+1)
			x0, y0, x1, y1 = lx, ly, lx+1, ly+1
		case 3: // diagonal '/': (x+1,y)..(x,y+1)
			x0, y0, x1, y1 = lx+1, ly, lx, ly+1
		default:
			continue
		}
		quads = append(quads, quad{x0, y0, x1, y1, h, colour})
	}
	if len(quads) == 0 {
		return nil
	}

	g := NewGameModel(len(quads)*4, len(quads))
	for _, q := range quads {
		// terrain feet
		hA := heights[q.x0][q.y0]
		hB := heights[q.x1][q.y1]
		// vertices: A-foot, A-top, B-top, B-foot (CCW, matching method422's
		// {i3, j3, k3, l3} = foot0, top0, top1, foot1).
		v0 := g.AddVertex(int32(q.x0)*128, -hA, int32(q.y0)*128)
		v1 := g.AddVertex(int32(q.x0)*128, -hA-q.h, int32(q.y0)*128)
		v2 := g.AddVertex(int32(q.x1)*128, -hB-q.h, int32(q.y1)*128)
		v3 := g.AddVertex(int32(q.x1)*128, -hB, int32(q.y1)*128)
		// both faces visible (front + back) so a wall is opaque from either
		// side — flat shaded (intensity != magic) so it reads as a solid slab.
		g.AddFace([]int{v0, v1, v2, v3}, q.colour, q.colour, 0)
	}
	// Walls use flat (per-face) shading so vertical faces don't gouraud-darken
	// to black. A higher base ambience (lightAmbience = 256-32*4 = 128) keeps
	// the back-lit side of a wall a readable grey rather than near-black, so
	// buildings/jail bars stay legible from every camera angle.
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
