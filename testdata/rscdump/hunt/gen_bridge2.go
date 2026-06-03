//go:build ignore

// gen_bridge2 authors two fixtures that isolate the flatten rule:
//   bridge_ov2.json   : an E-W band of overlay 2  (tileType 3 -> deob does NOT flatten)
//   bridge_ov4.json   : an E-W band of overlay 4  (tileType 4 -> deob DOES   flatten)
// At a uniform non-zero elevation. Per the deob, ov2 should stay at h=120 and
// ov4 should sink to h=0. The GO engine flattens BOTH (treats ov2 as water).
package main

import "github.com/gen0cide/westworld/internal/rscdump"

const (
	Size  = 20
	BaseX = 200
	BaseY = 200
	Elev  = 40
)

func idx(x, y int) int { return x*Size + y }

func build(bandOverlay byte) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	ov := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70
	}
	for x := 0; x < Size; x++ {
		for y := 8; y <= 11; y++ {
			ov[idx(x, y)] = bandOverlay
		}
	}
	return &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: rscdump.Camera{X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0,
			Distance: 1800, ViewDist: 9, ClipNear: 5, ClipFar: 7000, ScreenW: 512, ScreenH: 334},
		Window:  rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{Size: Size, Elevation: elev, GroundColour: gcol, Overlay: ov, TerrainSeed: 0},
		Self:    &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true},
	}
}

func main() {
	must(build(2).Save("bridge_ov2.json"))
	must(build(4).Save("bridge_ov4.json"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
