//go:build ignore

// gen_terrain authors the terrain-subsystem hunt fixtures as rscdump/1 L1 JSON.
//
//	go run gen_terrain.go
//
// Window: Size x Size, window-local [x*Size+y], x=col=worldX-BaseX.
// All fixtures use a uniform elevation so the ground is flat (height twist = 0)
// — this ISOLATES the colour-split / overlay logic from the height-twist split
// (terrain.go:310 `twist`), so any extra split faces are attributable to the
// colour-split rule under test.
package main

import "github.com/gen0cide/westworld/internal/rscdump"

const (
	Size  = 24
	BaseX = 200 // 200%48 = 8: deck cols sit inside a sector (no %48==47 seam noise)
	BaseY = 200
	Elev  = 40 // *3 = 120 world height (uniform => flat, no twist)
)

func idx(x, y int) int { return x*Size + y }

func cam() rscdump.Camera {
	return rscdump.Camera{
		X: 0, Y: -384, Z: 0,
		Pitch: 912, Yaw: 512, Roll: 0,
		Distance: 1800,
		ViewDist: 9, ClipNear: 5, ClipFar: 7000,
		ScreenW: 512, ScreenH: 334,
	}
}

// base builds a flat grass window with the given overlay grid.
func base(overlay []byte) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70 // green-ish grass palette index (band 64..127)
	}
	return &rscdump.Dump{
		Schema: rscdump.SchemaID,
		Level:  rscdump.LevelL1,
		Source: rscdump.SourceHandAuthored,
		Camera: cam(),
		Window: rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{
			Size:         Size,
			Elevation:    elev,
			GroundColour: gcol,
			Overlay:      overlay,
			TerrainSeed:  0,
		},
		Self: &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true},
	}
}

// patch fills a rectangular block of the overlay grid with id.
func patch(ov []byte, x0, y0, x1, y1 int, id byte) {
	for x := x0; x <= x1; x++ {
		for y := y0; y <= y1; y++ {
			ov[idx(x, y)] = id
		}
	}
}

func main() {
	// ── FIXTURE 1: tileType-2 INDOOR-FLOOR patch (overlay id 3 = wood planks,
	// type 2) in a field of grass. DEOB: line 932 `decoType != 2` => NO colour
	// split for a type-2 tile (unless it has a diagonal wall). The patch should
	// be SOLID quads of the indoor colour. GO terrain.go:279 splits ANY
	// hasOverlay tile, so it reverts boundary triangles to grass. Differential
	// vs an all-grass control isolates the spurious split.
	{
		ov := make([]byte, Size*Size)
		patch(ov, 8, 8, 14, 14, 3) // 7x7 indoor-floor block (id 3, type 2)
		base(ov).Save("terrain_indoor_split.json")
	}
	// control: same window, no patch (all grass)
	base(make([]byte, Size*Size)).Save("terrain_grass.json")

	// ── FIXTURE 2: two ADJACENT type-3 overlays with DIFFERENT ids but the SAME
	// tileType (id 2=water-edge tex type3, id 7=gungywater tex type3). DEOB
	// compares getTileType (both type3 => getTileType returns 0 for both =>
	// EQUAL => NO split at their shared seam). GO compares raw overlay id (2 != 7
	// => SPLIT). NOTE: id 2 is GO-special-cased as water (flattened+fixed water
	// face), so to isolate the *type-class* comparison cleanly we use id 7 vs id
	// 19 (both type 3, neither GO-water). Differential: faces at the 7|19 seam.
	{
		ov := make([]byte, Size*Size)
		patch(ov, 6, 6, 11, 17, 7)   // left half: gungywater tex (id 7, type 3)
		patch(ov, 12, 6, 17, 17, 19) // right half: outdoor special (id 19, type 3)
		base(ov).Save("terrain_typeclass_seam.json")
	}
	// control: SAME block, but a SINGLE id 7 throughout (no seam => no split
	// expected by EITHER rule) — proves the seam faces are caused by the 7|19
	// boundary, not the patch edge against grass.
	{
		ov := make([]byte, Size*Size)
		patch(ov, 6, 6, 17, 17, 7)
		base(ov).Save("terrain_typeclass_solid.json")
	}

	// ── FIXTURE 3: a single type-4 (water-class) overlay tile (id 4) surrounded
	// by GRASS (no overlay). DEOB buildOverlayTriangles (World.java:1117-1135)
	// emits a type-4 quad on the id-4 tile AND on each of its 4 NON-type-3
	// grass neighbours (the dN/dS/dE/dW emit). GO terrain.go:334 only emits the
	// deck quad on the type-4 tile itself (deck[i][j]). Differential: count
	// raised type-4 quads (centroid at real elevation).
	{
		ov := make([]byte, Size*Size)
		ov[idx(12, 12)] = 4 // one type-4 tile, grass all around
		base(ov).Save("terrain_type4_neighbour.json")
	}
	// control: all grass (no type-4 tile, no overlay quads at all)
	base(make([]byte, Size*Size)).Save("terrain_type4_none.json")

	// ── FIXTURE 4: a road/path overlay (id 1, type 1, flat colour) patch in
	// grass. DEOB line 932 `decoType != 2` is TRUE for type-1 (road), so the
	// colour split FIRES (path edges cut diagonally). This is a fixture where GO
	// and DEOB SHOULD agree on splitting — a positive control for the split.
	{
		ov := make([]byte, Size*Size)
		patch(ov, 9, 6, 9, 17, 1) // a 1-tile-wide N-S road (id 1)
		base(ov).Save("terrain_road_split.json")
	}
}
