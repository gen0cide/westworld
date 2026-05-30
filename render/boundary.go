package render

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// wallColour is the flat fill for an opaque wall face. Real RSC textures the
// wall front/back from wallObjectSomething2/wallObjectTexture; until the
// texture archive is wired (priority 4) we paint a neutral stone grey so walls
// read as solid surfaces (and so windows/doorframes still occlude correctly).
var (
	wallColourStone = method305(132, 132, 140)
	wallColourWood  = method305(120, 90, 55)
)

// wallShade is the FIXED flat shade index every wall face is lit at (like
// waterShade). Without it, the directional flat-shade swings each face across
// the FULL ramp — light-facing fronts blow out to near-white and back faces
// sink to near-black (the chaotic white/black panel look). A single mid-ramp
// index renders every wall a consistent readable stone grey from all sides;
// once the brick texture is sampled this is replaced by real textured shading.
const wallShade = 48

// BuildBoundaries assembles all wall/fence/door boundary quads inside the
// 96x96 window into a single GameModel, porting World.method422.
//
// CRITICAL: walls live in the landscape TILE records, NOT the sparse server
// boundary-locs file. Each Tile carries per-edge door-def indices
// (HorizontalWall / VerticalWall / DiagonalWalls); the authentic client reads
// them straight off the loaded section. The earlier version iterated
// f.BoundaryLocs (which holds only a sparse scatter of doorframes) and so
// rendered almost no walls — the dense castle/building walls were invisible.
// We now iterate the window's tiles exactly like the client:
//
//	HorizontalWall>0 -> east-west wall : edge (x,y)..(x+1,y)
//	VerticalWall>0   -> north-south    : edge (x,y)..(x,y+1)
//	DiagonalWalls 1..11999    -> '\' : edge (x,y)..(x+1,y+1)
//	DiagonalWalls 12001..23999-> '/' : edge (x+1,y)..(x,y+1)  (def id = val-12000)
//
// heights is the flattened terrain-height grid (TerrainHeights) so wall feet
// sit flush on the ground. The wall id indexes f.BoundaryDefs for the wall's
// Height (and a wood/stone colour heuristic). Returns nil if the window holds
// no walls.
func BuildBoundaries(f *facts.Facts, land *pathfind.Landscape, baseX, baseY, plane int, heights [][]int32) *GameModel {
	if land == nil {
		return nil
	}
	n := terrainSize

	type quad struct {
		x0, y0, x1, y1 int // local tile endpoints
		h              int32
		colour         int32
	}
	var quads []quad

	// addWall resolves the wall def (height + wood/stone) and appends a quad
	// for the given local tile edge.
	addWall := func(defID, x0, y0, x1, y1 int) {
		h := int32(192) // RSC default full-wall height
		colour := wallColourStone
		if f != nil {
			if def := f.BoundaryDef(defID); def != nil {
				if def.Height > 0 {
					h = int32(def.Height)
				}
				switch def.Name {
				case "Fence", "Gate", "Railing", "Wooden fence", "Iron railings", "Wooden Fence":
					colour = wallColourWood
				}
			}
		}
		quads = append(quads, quad{x0, y0, x1, y1, h, colour})
	}

	for lx := 0; lx < n-1; lx++ {
		for ly := 0; ly < n-1; ly++ {
			t := land.Tile(baseX+lx, baseY+ly, plane)
			if t.HorizontalWall > 0 { // east-west: (lx,ly)..(lx+1,ly)
				addWall(int(t.HorizontalWall), lx, ly, lx+1, ly)
			}
			if t.VerticalWall > 0 { // north-south: (lx,ly)..(lx,ly+1)
				addWall(int(t.VerticalWall), lx, ly, lx, ly+1)
			}
			if d := t.DiagonalWalls; d > 0 {
				if d < 12000 { // '\' diagonal: (lx,ly)..(lx+1,ly+1)
					addWall(int(d), lx, ly, lx+1, ly+1)
				} else { // '/' diagonal: (lx+1,ly)..(lx,ly+1)
					addWall(int(d-12000), lx+1, ly, lx, ly+1)
				}
			}
		}
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
		// side. FIXED shade (not normal-derived) so neither face blows out to
		// white or sinks to black — a uniform readable stone slab.
		g.AddFixedFace([]int{v0, v1, v2, v3}, q.colour, q.colour, wallShade)
	}
	// Walls use flat (per-face) shading so vertical faces don't gouraud-darken
	// to black. A higher base ambience (lightAmbience = 256-32*4 = 128) keeps
	// the back-lit side of a wall a readable grey rather than near-black, so
	// buildings/jail bars stay legible from every camera angle.
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
