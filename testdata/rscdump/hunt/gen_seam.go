//go:build ignore

// gen_seam authors a SECTOR-ALIGNED bridge-seam fixture (rscdump/1 L1 JSON) that
// actually exercises the setTileDecorationOnBridge 250->9 plank-edge remap at the
// window-local column x==47 seam (RENDER_RECONCILIATION.md §E row 3 / FINDINGS #15
// bridge_seam_coord).
//
//	go run gen_seam.go
//
// Why a new fixture: the existing bridge_on is BaseX=200 (200%48=8) AND its 250
// deck columns sit at orsc window-local 47..50 but every column's +x neighbour is
// ALSO 250 (the run is contiguous), so the x==47 branch ALWAYS falls through to
// setTileDecoration(2) — the 250->9 seam is never taken. This fixture places a
// 250 tile whose orsc-window-local column is EXACTLY 47 and whose +x neighbour is
// GRASS (overlay 0, neither 250 nor 2), so the seam branch fires (250 -> 9).
//
// Coordinate math (orsc dumpView centres the window on Self.x at window-local 48):
//   orsc window baseX = Self.x - 48.  A fixture tile at fixture-local fx (world =
//   BaseX+fx) lands at orsc-window-local = (BaseX+fx) - (Self.x-48).
//   With BaseX=192 (192%48==0, sector-aligned) and Self.x=204:
//     orsc-local = 192 + fx - 204 + 48 = 36 + fx.
//   So orsc-local 47 == fixture-local fx = 11.  The seam tile is at fx=11; its +x
//   neighbour (fx=12, orsc-local 48) is grass so the 250->9 branch is taken.
package main

import (
	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	Size  = 24
	BaseX = 192 // 192 % 48 == 0  -> sector-aligned (the planner's requirement)
	BaseY = 192
	SelfX = 204 // -> orsc window baseX = 156; orsc-local 47 == fixture-local 11
	SelfY = 204
	Elev  = 40
)

func idx(x, y int) int { return x*Size + y }

func main() {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	ov := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70 // grass
	}

	// Water band (overlay 2) across the middle rows, with a SINGLE bridge (250)
	// tile at fixture-local x=11 (orsc-window-local 47) for each water row. The +x
	// neighbour (x=12) is left as grass (0) so the seam branch (250 -> 9) fires.
	// The -x side (x=10) is overlay-2 water, so the bridge tile spans water but its
	// east edge is the open seam the plank-lip should land on.
	riverY0, riverY1 := 8, 15
	const seamX = 11 // orsc-window-local 47
	for x := 0; x < Size; x++ {
		for y := 0; y < Size; y++ {
			if y >= riverY0 && y <= riverY1 {
				if x <= seamX-1 {
					ov[idx(x, y)] = 2 // water west of the seam
				} else if x == seamX {
					ov[idx(x, y)] = 250 // bridge deck tile AT the seam column
				}
				// x >= seamX+1 stays grass (0) -> the seam's +x neighbour is dry land
			}
		}
	}

	d := &rscdump.Dump{
		Schema: rscdump.SchemaID,
		Level:  rscdump.LevelL1,
		Source: rscdump.SourceHandAuthored,
		Camera: rscdump.Camera{
			X: 0, Y: -384, Z: 0,
			Pitch: 912, Yaw: 512, Roll: 0,
			Distance: 1800,
			ViewDist: 9, ClipNear: 5, ClipFar: 7000,
			ScreenW: 512, ScreenH: 334,
		},
		Window: rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{
			Size:         Size,
			Elevation:    elev,
			GroundColour: gcol,
			Overlay:      ov,
			TerrainSeed:  0,
		},
		Self: &rscdump.Self{X: SelfX, Y: SelfY, NoSelf: true},
	}
	if err := d.Save("seam_sector_aligned.json"); err != nil {
		panic(err)
	}
}
