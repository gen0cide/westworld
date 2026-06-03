//go:build ignore

// gen_pipeline authors render-pipeline hunt fixtures (face normals / front-back
// facing / gouraud / painter sort / near-clip) as rscdump/1 L1 JSON.
//
//	go run gen_pipeline.go
//
// Window-local index is [x*Size + y] for the byte grids (x=col=worldX-BaseX).
// All fixtures share a flat terrain at a uniform elevation so the only variable
// is the wall geometry under test.
package main

import (
	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	Size  = 16
	BaseX = 200
	BaseY = 200
	Elev  = 40 // *3 = 120 world height
)

func idx(x, y int) int { return x*Size + y }

func cam() rscdump.Camera {
	return rscdump.Camera{
		X: 0, Y: -384, Z: 0,
		Pitch: 912, Yaw: 512, Roll: 0,
		Distance: 1500,
		ViewDist: 9, ClipNear: 5, ClipFar: 7000,
		ScreenW: 512, ScreenH: 334,
	}
}

// base builds a flat-grass dump; wallH/wallV/wallDiag are window-local byte grids.
func base(wallH, wallV []byte, wallDiag []int32) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	ov := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70
	}
	if wallH == nil {
		wallH = make([]byte, n)
	}
	if wallV == nil {
		wallV = make([]byte, n)
	}
	if wallDiag == nil {
		wallDiag = make([]int32, n)
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
			Overlay:      ov,
			WallH:        wallH,
			WallV:        wallV,
			WallDiag:     wallDiag,
			TerrainSeed:  0,
		},
		Self: &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true},
	}
}

func main() {
	n := Size * Size

	// --- corner: a perpendicular L-junction at tile (8,8): one E-W wall and one
	// N-S wall sharing the corner. Stresses the painter overlap sort (method277):
	// the two quads meet edge-on at the shared corner.
	wh := make([]byte, n) // HorizontalWall -> north-south edge
	wv := make([]byte, n) // VerticalWall   -> east-west edge
	wh[idx(8, 8)] = 1
	wv[idx(8, 8)] = 1
	must(base(wh, wv, nil).Save("pipeline_corner.json"))

	// --- two parallel facing walls on adjacent tiles: tests front/back facing
	// selection (one wall faces the camera, one faces away). Walls at the E-W
	// edge of tiles (8,8) and (8,10): the camera (south, looking north) sees the
	// front of one and the back of the other.
	wv2 := make([]byte, n)
	wv2[idx(8, 8)] = 1
	wv2[idx(8, 10)] = 1
	must(base(nil, wv2, nil).Save("pipeline_parallel.json"))

	// --- single E-W wall control (one quad, isolates flat-vs-gouraud shading on a
	// known-orientation face).
	wv3 := make([]byte, n)
	wv3[idx(8, 9)] = 1
	must(base(nil, wv3, nil).Save("pipeline_singlewall.json"))

	// --- no walls (terrain only) control.
	must(base(nil, nil, nil).Save("pipeline_bare.json"))

	// --- a wall placed VERY close to the camera tile to exercise near-plane clip.
	// Camera is at host tile centre; put a N-S wall right on the host's tile.
	whclose := make([]byte, n)
	whclose[idx(8, 7)] = 1
	whclose[idx(8, 8)] = 1
	must(base(whclose, nil, nil).Save("pipeline_nearclip.json"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
