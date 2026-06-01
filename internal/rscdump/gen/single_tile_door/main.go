// Command single_tile_door authors the Phase-0 render-diff fixture: a
// hand-authored single-tile-with-door rscdump/1 JSON (RENDER_DIFF_DESIGN.md §6
// "the SINGLE first tool"). It is generated programmatically (rather than typed
// by hand) so the per-tile grids are guaranteed self-consistent, then written
// with Dump.Save which validates the schema.
//
//	go run ./internal/rscdump/gen/single_tile_door
//
// The scene: a small flat grass region, ONE door/wall boundary on one tile edge
// near the centre, a fixed camera looking at it, no entities, terrainSeed=0
// (flat zero ambience ⇒ byte-identical renders), raw ground colours (palette
// index, no texture archive). The renderer centres its 160-tile window on the
// host tile (BaseX+Size/2), pulling zeros for every tile outside this grid, so
// the grid only needs to cover the tiles around the door.
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gen0cide/westworld/internal/rscdump"
)

func main() {
	const (
		size  = 16  // grid side (tiles); the renderer's own window is 160 and pulls zeros outside
		baseX = 200 // SW corner world-tile X
		baseY = 200 // SW corner world-tile Y
		plane = 0

		// grass ground-colour palette index: the green band of groundColour[]
		// starts at 64 (method305(j*3,144,0)); index 70 is a mid grass green.
		grass = byte(70)

		// the door's wall-def id (1-based in the grid byte). RenderDump's
		// syntheticFacts gives id (val-1) a generic openable wood door def, so the
		// boundary renders as a wood door leaf without external GameData.
		doorWallID = byte(2)
	)

	n := size * size
	idx := func(x, y int) int { return x*size + y }

	elevation := make([]byte, n)   // all 0 ⇒ flat terrain
	ground := make([]byte, n)      // grass everywhere
	overlay := make([]byte, n)     // no overlays
	roof := make([]byte, n)        // no roofs
	wallH := make([]byte, n)       // HorizontalWall (north-south edge)
	wallV := make([]byte, n)       // VerticalWall (east-west edge)
	wallDiag := make([]int32, n)   // no diagonals
	for i := range ground {
		ground[i] = grass
	}

	// Host tile is BaseX+Size/2 = (208, 208). Put the door one tile in front of
	// the host (toward +Y, which the camera at yaw 512 looks across) on its
	// east-west edge. VerticalWall>0 ⇒ east-west wall on edge (x,y)..(x+1,y)
	// (boundary.go edge-axis convention). Window-local door tile:
	doorLX, doorLY := size/2, size/2+1 // (8, 9) ⇒ world (208, 209)
	wallV[idx(doorLX, doorLY)] = doorWallID

	d := &rscdump.Dump{
		Schema: rscdump.SchemaID,
		Level:  rscdump.LevelL1,
		Source: rscdump.SourceHandAuthored,
		Tick:   0,
		Camera: rscdump.Camera{
			// Eye position is informational for the GO engine (RenderView derives
			// the camera from rotation+zoom around the host tile); we still record
			// a plausible eye so a jar/deob engine has the value. yaw 512 (rotation
			// 128) looks across +Y; distance 1500 (zoom 750) is RenderView's default.
			X: 0, Y: -384, Z: 0,
			Pitch:    912,
			Yaw:      512,
			Roll:     0,
			Distance: 1500,
			ViewDist: 9,
			ClipNear: 5,
			ClipFar:  7000,
			ScreenW:  512,
			ScreenH:  334,
		},
		Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{
			Size:         size,
			Elevation:    elevation,
			GroundColour: ground,
			Overlay:      overlay,
			Roof:         roof,
			WallH:        wallH,
			WallV:        wallV,
			WallDiag:     wallDiag,
			TerrainSeed:  0,
		},
		// No entities/scenery/items: a bare flat-grass-with-door scene. NoSelf so
		// the local player is not drawn over the door.
		Self: &rscdump.Self{X: baseX + size/2, Y: baseY + size/2, NoSelf: true},
	}

	// Repo-root-relative output: this command lives at
	// internal/rscdump/gen/single_tile_door, so the module root is four dirs up.
	out, err := filepath.Abs(filepath.Join("testdata", "rscdump", "single_tile_door.json"))
	if err != nil {
		log.Fatal(err)
	}
	if err := d.Save(out); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("wrote %s (size=%d door@local(%d,%d)=world(%d,%d))\n", out, size, doorLX, doorLY, baseX+doorLX, baseY+doorLY)
}
