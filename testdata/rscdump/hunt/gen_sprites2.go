//go:build ignore

// gen_sprites2 authors occlusion fixtures: an NPC standing BEHIND a wall (to
// test whether GO's per-pixel z-buffer + 64-unit foot bias matches the deob's
// unified painter sort) and an NPC standing ON a slope (foot-anchor test).
//
//	go run gen_sprites2.go
package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	size  = 16
	baseX = 200
	baseY = 200
	plane = 0
	grass = byte(70)
)

func flat() (elev, ground, overlay, roof, wallH, wallV []byte, wallDiag []int32) {
	n := size * size
	elev = make([]byte, n)
	ground = make([]byte, n)
	overlay = make([]byte, n)
	roof = make([]byte, n)
	wallH = make([]byte, n)
	wallV = make([]byte, n)
	wallDiag = make([]int32, n)
	for i := range ground {
		ground[i] = grass
	}
	return
}

func cam() rscdump.Camera {
	return rscdump.Camera{X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0,
		Distance: 1500, ViewDist: 9, ClipNear: 5, ClipFar: 7000, ScreenW: 512, ScreenH: 334}
}

func save(name string, d *rscdump.Dump) {
	out, _ := filepath.Abs(filepath.Join("testdata", "rscdump", "hunt", name))
	if err := d.Save(out); err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	fmt.Printf("wrote %s\n", out)
}

func main() {
	idx := func(x, y int) int { return x*size + y }
	hostX, hostY := baseX+size/2, baseY+size/2 // (208,208)

	// --- NPC behind a wall ---
	// Wall on the east-west edge one tile in front of host (local y=9), NPC three
	// tiles in front (world y=211). Camera at host looking +Y, so the wall is
	// between the camera and the NPC; the NPC should be OCCLUDED by the wall.
	elev, ground, overlay, roof, wallH, wallV, wallDiag := flat()
	wallV[idx(size/2, size/2+1)] = 2 // wood-door wall on edge (8,9)
	dWall := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: cam(), Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: elev, GroundColour: ground, Overlay: overlay,
			Roof: roof, WallH: wallH, WallV: wallV, WallDiag: wallDiag, TerrainSeed: 0},
		Entities: []rscdump.Entity{{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 0}},
		Self:     &rscdump.Self{X: hostX, Y: hostY, NoSelf: true},
	}
	save("sprites_npc_behind_wall.json", dWall)

	// twin: same NPC, NO wall (control: NPC fully visible)
	elev2, ground2, overlay2, roof2, wallH2, wallV2, wallDiag2 := flat()
	dNoWall := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: cam(), Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: elev2, GroundColour: ground2, Overlay: overlay2,
			Roof: roof2, WallH: wallH2, WallV: wallV2, WallDiag: wallDiag2, TerrainSeed: 0},
		Entities: []rscdump.Entity{{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 0}},
		Self:     &rscdump.Self{X: hostX, Y: hostY, NoSelf: true},
	}
	save("sprites_npc_nowall.json", dNoWall)

	// --- NPC IN FRONT of host wall, on the SAME tile as the wall ---
	// (the self-occlusion case the 64-bias is meant to handle). NPC on the wall
	// tile itself.
	elev3, ground3, overlay3, roof3, wallH3, wallV3, wallDiag3 := flat()
	wallV3[idx(size/2, size/2+1)] = 2
	dOnWall := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: cam(), Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: elev3, GroundColour: ground3, Overlay: overlay3,
			Roof: roof3, WallH: wallH3, WallV: wallV3, WallDiag: wallDiag3, TerrainSeed: 0},
		Entities: []rscdump.Entity{{X: hostX, Y: hostY + 1, Kind: "npc", ID: 1, Heading: 0}},
		Self:     &rscdump.Self{X: hostX, Y: hostY, NoSelf: true},
	}
	save("sprites_npc_on_wall.json", dOnWall)

	// --- NPC + ground item on a slope (foot-anchor test) ---
	elevS := make([]byte, size*size)
	groundS := make([]byte, size*size)
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			elevS[idx(x, y)] = byte(y * 8) // steeper ramp in +Y (24 world units/tile)
			groundS[idx(x, y)] = grass
		}
	}
	oS := make([]byte, size*size)
	rS := make([]byte, size*size)
	wHS := make([]byte, size*size)
	wVS := make([]byte, size*size)
	wDS := make([]int32, size*size)
	dSlope := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: cam(), Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: elevS, GroundColour: groundS, Overlay: oS,
			Roof: rS, WallH: wHS, WallV: wVS, WallDiag: wDS, TerrainSeed: 0},
		Entities:    []rscdump.Entity{{X: hostX, Y: hostY + 3, Kind: "npc", ID: 1, Heading: 0}},
		GroundItems: []rscdump.GroundItem{{X: hostX, Y: hostY + 4, ItemID: 10}},
		Self:        &rscdump.Self{X: hostX, Y: hostY, NoSelf: true},
	}
	save("sprites_npc_slope.json", dSlope)
}
