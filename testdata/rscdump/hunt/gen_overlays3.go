//go:build ignore

// gen_overlays3 authors additional textured-overlay terrain fixtures that exercise
// overlay def ids the existing hunt set does not: lava (id 11, tileType 3 texture),
// lava-class (id 12, tileType 4 forced colour 31), pentagram (id 14, tileType 2
// textured), and marble floor (id 17, tileType 2 textured). All three render legs
// (orsc, the rev-235 JAR, the DEOB) resolve these ids through the SAME synthesized
// def tables (orsc tileDefs / DumpRender OVERLAY_DEFS / DumpRenderer OVERLAY_DEFS),
// and the textured fills degrade to a transparent texel bank in every leg (no
// texture archive is loaded), so the underlying gouraud terrain shows identically.
//
// Run:  go run testdata/rscdump/hunt/gen_overlays3.go
package main

import "github.com/gen0cide/westworld/internal/rscdump"

const (
	Size  = 24
	BaseX = 200
	BaseY = 200
	Elev  = 40
)

func idx(x, y int) int { return x*Size + y }

func cam() rscdump.Camera {
	return rscdump.Camera{X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0, Distance: 1800, ViewDist: 9, ClipNear: 5, ClipFar: 7000, ScreenW: 512, ScreenH: 334}
}

func base(ov []byte) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	g := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		g[i] = 70
	}
	return &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored, Camera: cam(),
		Window:  rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{Size: Size, Elevation: elev, GroundColour: g, Overlay: ov, TerrainSeed: 0},
		Self:    &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true},
	}
}

// patch fills an 8x8 block of overlay id v centred in the window.
func patch(v byte) []byte {
	ov := make([]byte, Size*Size)
	for x := 8; x <= 15; x++ {
		for y := 8; y <= 15; y++ {
			ov[idx(x, y)] = v
		}
	}
	return ov
}

// single sets one lone overlay tile at (12,12) so the type-4 neighbour-spread /
// type-3 texture path runs on an isolated tile surrounded by grass.
func single(v byte) []byte {
	ov := make([]byte, Size*Size)
	ov[idx(12, 12)] = v
	return ov
}

func main() {
	base(patch(11)).Save("testdata/rscdump/hunt/overlay_lava_ov11.json")     // lava texture (type-3, tex 31)
	base(patch(12)).Save("testdata/rscdump/hunt/overlay_lavaclass_ov12.json") // lava-class (type-4 -> forced colour 31)
	base(patch(14)).Save("testdata/rscdump/hunt/overlay_pentagram_ov14.json") // pentagram (type-2, tex 32)
	base(patch(17)).Save("testdata/rscdump/hunt/overlay_marble_ov17.json")    // marble floor (type-2, tex 15)
	base(single(11)).Save("testdata/rscdump/hunt/overlay_lava_lone_ov11.json")
	base(single(2)).Save("testdata/rscdump/hunt/overlay_wateredge_lone_ov2.json") // water-edge (type-3, tex 1)
}
