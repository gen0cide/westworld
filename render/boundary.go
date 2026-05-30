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
//	HorizontalWall>0 -> north-south wall: edge (x,y)..(x,y+1)  (.orsc field runs along Y/Z)
//	VerticalWall>0   -> east-west wall  : edge (x,y)..(x+1,y)  (.orsc field runs along X)
//	DiagonalWalls 1..11999    -> '\' : edge (x,y)..(x+1,y+1)
//	DiagonalWalls 12001..23999-> '/' : edge (x+1,y)..(x,y+1)  (def id = val-12000)
//
// heights is the flattened terrain-height grid (TerrainHeights) so wall feet
// sit flush on the ground. The wall id indexes f.BoundaryDefs for the wall's
// Height (and a wood/stone colour heuristic). Returns nil if the window holds
// no walls.
// removed, when non-nil, reports whether the boundary edge at absolute
// (x, y, dir) has been cleared by a live server update (door opened, web cut)
// using the createModel dir convention (see render.View.BoundaryRemoved). A
// matching edge is skipped so it renders passable instead of solid. nil
// disables the check entirely (static walls render unchanged).
func BuildBoundaries(f *facts.Facts, land *pathfind.Landscape, baseX, baseY, plane int, heights [][]int32, removed func(x, y, dir int) bool, underRoof bool) *GameModel {
	if land == nil {
		return nil
	}
	n := terrainSize
	// isRemoved nil-guards the optional dynamic-boundary override.
	isRemoved := func(x, y, dir int) bool {
		return removed != nil && removed(x, y, dir)
	}

	type quad struct {
		x0, y0, x1, y1      int   // local tile endpoints
		base                int32 // accumulated wall-top base for this story (0 = ground)
		h                   int32 // wall height
		frontFill, backFill int32 // texture id (>=0) when textured, else flat colour (<0)
		textured            bool
	}
	var quads []quad
	var storyBase int32 // set per story; stamped onto each quad's base

	// addWall resolves the wall def and appends a quad for the given local tile
	// edge. Real RSC wall textures come from modelVar2/modelVar3 (FrontDeco /
	// BackDeco) — route them through the textured-span fill when a texel buffer
	// exists for that id; otherwise fall back to the flat wood/stone colour.
	addWall := func(defID, x0, y0, x1, y1 int) {
		h := int32(192) // RSC default full-wall height
		front, back := wallColourStone, wallColourStone
		textured := false
		if f != nil {
			if def := f.BoundaryDef(defID); def != nil {
				if def.Height > 0 {
					h = int32(def.Height)
				}
				switch def.Name {
				case "Fence", "Gate", "Railing", "Wooden fence", "Iron railings", "Wooden Fence":
					front, back = wallColourWood, wallColourWood
				}
				if def.FrontDeco >= 0 && textureBuffer(int32(def.FrontDeco)) != nil {
					textured = true
					front = int32(def.FrontDeco)
					back = front
					if def.BackDeco >= 0 && textureBuffer(int32(def.BackDeco)) != nil {
						back = int32(def.BackDeco)
					}
				}
			}
		}
		quads = append(quads, quad{x0, y0, x1, y1, storyBase, h, front, back, textured})
	}

	// buildStory reads ONE archive plane's tile walls. Multi-story buildings
	// keep each upper story's walls on a SEPARATE archive plane (h1/h2/h3) at the
	// same (sx,sy) — exactly like roofs (roof.go BuildRoofs). Iterate
	// plane..maxRoofPlane, stacking each story at (sp-plane)*planeWallTopBonus so
	// upper-story walls land under the upper-story roofs already drawn. Without
	// this only the ground walls rendered and the upper roofs floated over gaps.
	buildStory := func(sp int) {
		storyBase = int32(sp-plane) * planeWallTopBonus
		for lx := 0; lx < n-1; lx++ {
			for ly := 0; ly < n-1; ly++ {
				t := land.Tile(baseX+lx, baseY+ly, sp)
				// def ids are 0-based (GameData.wallObjectHeight[i]); the stored
				// wall value is 1-based, so subtract 1 (straight/'\') or 12001 ('/').
				// EDGE-AXIS SWAP (verified vs raw .orsc data): the .orsc
				// HorizontalWall byte (offset 4) physically runs along Y/Z, and
				// VerticalWall (offset 5) runs along X — the inverse of the field
				// names. Drawing them the other way rotated every straight wall 90°.
				// Dynamic-boundary dir convention (createModel, mudclient.java:6769):
				// dir 0 = east-west edge (the .orsc VerticalWall byte), dir 1 =
				// north-south edge (HorizontalWall), dir 2 = '\', dir 3 = '/'.
				ax, ay := baseX+lx, baseY+ly
				if t.HorizontalWall > 0 && !isRemoved(ax, ay, 1) { // north-south edge: (lx,ly)..(lx,ly+1)
					addWall(int(t.HorizontalWall)-1, lx, ly, lx, ly+1)
				}
				if t.VerticalWall > 0 && !isRemoved(ax, ay, 0) { // east-west edge: (lx,ly)..(lx+1,ly)
					addWall(int(t.VerticalWall)-1, lx, ly, lx+1, ly)
				}
				// Diagonal walls occupy ONLY 1..23999. Values >=24000 (esp. the
				// 48000+ band) are diagonally-placed SCENERY object markers
				// (World.addModels, def = val-48001), NOT walls — the old open-ended
				// else swallowed them and drew them as garbage tilted '/' quads (the
				// "weird angles"). Bound the range so they're dropped here.
				if d := t.DiagonalWalls; d > 0 && d < 24000 {
					if d < 12000 { // '\' diagonal: (lx,ly)..(lx+1,ly+1), def d-1, dir 2
						if !isRemoved(ax, ay, 2) {
							addWall(int(d)-1, lx, ly, lx+1, ly+1)
						}
					} else if !isRemoved(ax, ay, 3) { // '/' diagonal (12000<d<24000): (lx+1,ly)..(lx,ly+1), def d-12001, dir 3
						addWall(int(d-12001), lx+1, ly, lx, ly+1)
					}
				}
			}
		}
	}
	// When the host is UNDER a roof, the client frees the upper-story walls
	// (577[1]/577[2]) and keeps only the ground-floor walls (577[0]) — so an
	// interior/passage isn't boxed in by the storey above (the Lumbridge gate
	// arch's plane-1 walls). Otherwise stack every storey as before.
	hi := maxRoofPlane
	if underRoof {
		hi = plane
	}
	for sp := plane; sp <= hi; sp++ {
		buildStory(sp)
	}
	if len(quads) == 0 {
		return nil
	}

	g := NewGameModel(len(quads)*4, len(quads))
	for _, q := range quads {
		// terrain feet, lifted by the per-story base so upper-story walls stack
		// at their plane height (under the matching upper-story roofs).
		hA := heights[q.x0][q.y0] + q.base
		hB := heights[q.x1][q.y1] + q.base
		// vertices: A-foot, A-top, B-top, B-foot (CCW, matching method422's
		// {i3, j3, k3, l3} = foot0, top0, top1, foot1).
		v0 := g.AddVertex(int32(q.x0)*128, -hA, int32(q.y0)*128)
		v1 := g.AddVertex(int32(q.x0)*128, -hA-q.h, int32(q.y0)*128)
		v2 := g.AddVertex(int32(q.x1)*128, -hB-q.h, int32(q.y1)*128)
		v3 := g.AddVertex(int32(q.x1)*128, -hB, int32(q.y1)*128)
		// both faces visible (front + back) so a wall is opaque from either side.
		// Textured walls: real brick/stone pattern via the textured-span path
		// (positive fill = texture id). Untextured: FIXED shade so neither face
		// blows out to white or sinks to black — a uniform readable stone slab.
		if q.textured {
			g.AddFace([]int{v0, v1, v2, v3}, q.frontFill, q.backFill, 0)
		} else {
			g.AddFixedFace([]int{v0, v1, v2, v3}, q.frontFill, q.backFill, wallShade)
		}
	}
	// Walls use flat (per-face) shading so vertical faces don't gouraud-darken
	// to black. A higher base ambience (lightAmbience = 256-32*4 = 128) keeps
	// the back-lit side of a wall a readable grey rather than near-black, so
	// buildings/jail bars stay legible from every camera angle.
	g.SetLight(32, 48, -50, -10, -50)
	return g
}
