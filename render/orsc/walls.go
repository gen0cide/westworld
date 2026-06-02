package orsc

import (
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/pathfind"
)

// walls.go ports the WALL pass of World.generateLandscapeModel (World.java:
// 822-896) + insertWallIntoModel (World.java:1425): the 3D wall/fence geometry
// (buildings, fences) built from each tile edge's HorizontalWall / VerticalWall /
// diagonal-wall id + the boundary def (DoorDef: height + front/back texture).
// Walls live in a SEPARATE accumulator model from the terrain, with their own
// diffuse light (World.java:891 setDiffuseLightAndColor(-50,-10,-50,60,24,false,
// 122)). Openable boundaries (doors; DoorDef.getUnknown()!=0) are NOT built here.
//
// MULTI-STORY: the wall foot heights come from the SHARED elevation cache
// (stories.go), NOT a fresh getTileElevation read — that is the whole point of
// the cache (insertWallIntoModel reads tileElevationCache, World.java:1440-1443).
// An upper story's walls therefore start at the wall-tops the lower stories baked
// into the cache, stacking the building vertically.

// wallModelCapacity matches the Java accumulator size RSModel(18688,18688)
// (World.java:519) — far above any window's wall count. Shared across all
// stacked stories' walls (one combined model in stories.go).
const wallModelCapacity = 18688

func (b *terrainBuilder) getVerticalWall(x, z int) int {
	if !b.inWindow(x, z) {
		return 0
	}
	return int(b.tile(x, z).VerticalWall) & 0xff
}

func (b *terrainBuilder) getHorizontalWall(x, z int) int {
	if !b.inWindow(x, z) {
		return 0
	}
	return int(b.tile(x, z).HorizontalWall) & 0xff
}

// insertWall ports insertWallIntoModel (World.java:1425): a wall quad from edge
// (t1X,t1Z)->(t2X,t2Z), rising getWallObjectHeight (def.Height) above the cached
// elevation at each end, textured front/back (def.FrontDeco / def.BackDeco). The
// foot heights are read from the SHARED cache ec (not getTileElevation) so upper
// stories stack on the lower stories' wall-tops. The Java's setVertexLightOther
// side-effects on the TERRAIN model are omitted (the wall model is lit on its own).
func (b *terrainBuilder) insertWall(model *Model, def *facts.BoundaryDef, ec *elevCache, t2X, t1Z, t1X, t2Z int) {
	height := int32(def.Height)
	// insertWallIntoModel reads tileElevationCache RAW (World.java:1440-1443). The
	// cache is marker-free here because the previous story's roofPass ended with the
	// full-grid +80000 strip (World.java:1155-1158, see roofPass); so a wall foot
	// lands on its real accumulated height — terrain + the lower stories' wall tops.
	e1 := ec[t1X][t1Z]
	e2 := ec[t2X][t2Z]
	x1, z1 := int32(t1X*128), int32(t1Z*128)
	x2, z2 := int32(t2X*128), int32(t2Z*128)
	v1 := model.insertVertex(x1, -e1, z1)
	v2 := model.insertVertex(x1, -e1-height, z1)
	v3 := model.insertVertex(x2, -e2-height, z2)
	v4 := model.insertVertex(x2, -e2, z2)
	if v1 < 0 || v2 < 0 || v3 < 0 || v4 < 0 {
		return // accumulator full (shouldn't happen below 18688 verts)
	}
	model.insertFace(4, []int{v1, v2, v3, v4}, int32(def.FrontDeco), int32(def.BackDeco))
}

// wallPass appends this story's wall quads to model, reading foot heights from the
// shared cache ec (World.java:822-896). Does not mutate ec (the roof pass does).
func (b *terrainBuilder) wallPass(f *facts.Facts, model *Model, ec *elevCache) {
	// place a non-openable wall (DoorDef.getUnknown()==0); skip doors/doorframes.
	place := func(id, t2X, t1Z, t1X, t2Z int) {
		def := f.BoundaryDef(id)
		if def == nil || def.Unknown != 0 {
			return
		}
		b.insertWall(model, def, ec, t2X, t1Z, t1X, t2Z)
	}

	for x := 0; x < worldWindowTiles-1; x++ {
		for z := 0; z < worldWindowTiles-1; z++ {
			if w := b.getVerticalWall(x, z); w > 0 { // World.java:828-831
				place(w-1, x+1, z, x, z)
			}
			if w := b.getHorizontalWall(x, z); w > 0 { // World.java:843-846
				place(w-1, x, z, x, z+1)
			}
			d := int(b.getWallDiagonal(x, z))
			if d > 0 && d < 12000 { // World.java:858-861 ('\' diagonal)
				place(d-1, x+1, z, x, z+1)
			} else if d > 12000 && d < 24000 { // World.java:873-875 ('/' diagonal)
				place(d-12001, x, z, x+1, z+1)
			}
		}
	}
}

// doorPass builds the OPENABLE boundaries (doors, doorframes, gates, webs;
// DoorDef.getUnknown()!=0) that wallPass deliberately skips — as the door-object
// panels OpenRSC draws via createWallObjectModel (mudclient.java:2578-2624). The
// static landscape model omits them (World.java:830 gate); the live client builds
// them from the server's wall-object instances. We drive the SAME geometry from
// the landscape boundary ids: the panel is the identical quad insertWall makes
// (foot from the cache, Height tall, FrontDeco/BackDeco = getModelVar2/getModelVar3
// — e.g. a Doorframe's texture 4 doorway-arch, a Door's texture 0 leaf), so reading
// the shared cache lets upper-story doors stack with their walls. Rendered closed
// (the static default). The caller gives this its own model + the door diffuse
// light (-95, createWallObjectModel) so it shades distinctly from the walls (122).
func (b *terrainBuilder) doorPass(f *facts.Facts, model *Model, ec *elevCache) {
	place := func(id, t2X, t1Z, t1X, t2Z int) {
		def := f.BoundaryDef(id)
		if def == nil || def.Unknown == 0 { // only the openable ones (walls did the rest)
			return
		}
		b.insertWall(model, def, ec, t2X, t1Z, t1X, t2Z)
	}
	for x := 0; x < worldWindowTiles-1; x++ {
		for z := 0; z < worldWindowTiles-1; z++ {
			if w := b.getVerticalWall(x, z); w > 0 {
				place(w-1, x+1, z, x, z)
			}
			if w := b.getHorizontalWall(x, z); w > 0 {
				place(w-1, x, z, x, z+1)
			}
			d := int(b.getWallDiagonal(x, z))
			if d > 0 && d < 12000 {
				place(d-1, x+1, z, x, z+1)
			} else if d > 12000 && d < 24000 {
				place(d-12001, x, z, x+1, z+1)
			}
		}
	}
}

// BuildWalls builds a single-story wall model from a freshly-seeded cache. Kept
// for tests / standalone use; the live render path uses BuildStories (stories.go)
// so the cache accumulates across floors. Returns nil if facts is unavailable.
func BuildWalls(land *pathfind.Landscape, f *facts.Facts, baseX, baseY, plane int) *Model {
	if f == nil {
		return nil
	}
	b := &terrainBuilder{land: land, baseX: baseX, baseY: baseY, plane: plane}
	var ec elevCache
	b.seedElevationCache(&ec)
	model := NewModel(wallModelCapacity, wallModelCapacity)
	b.wallPass(f, model, &ec)
	// World.java:891 — wall diffuse light.
	model.setDiffuseLightAndColor(-50, -10, -50, 60, 24, false, 122)
	return model
}
