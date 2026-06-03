//go:build ignore

// gen_terrain2: corner fixtures to trigger the colour-split CLASS-comparison
// divergence (DEOB compares getTileType, GO compares raw overlay id).
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
	return rscdump.Camera{X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0,
		Distance: 1800, ViewDist: 9, ClipNear: 5, ClipFar: 7000, ScreenW: 512, ScreenH: 334}
}
func base(ov []byte) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70
	}
	return &rscdump.Dump{Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: cam(), Window: rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{Size: Size, Elevation: elev, GroundColour: gcol, Overlay: ov, TerrainSeed: 0},
		Self:    &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true}}
}

func main() {
	// CONCAVE-CORNER meeting of two DIFFERENT same-type-3 overlays:
	// Fill a big block of id 7 (type 3). Carve an L: the tile at (12,12) keeps id7,
	// but its -x neighbour (11,12) and -y neighbour (12,11) are id 19 (also type 3).
	// At tile (12,12): neighbour(-x)=19 != 7 AND neighbour(-y)=19 != 7  -> the GO
	// corner-condition `ovlClassAt(i-1,j)!=me && ovlClassAt(i,j-1)!=me` FIRES (raw
	// id differs). DEOB: getTileType(11,12)=0 (type3, !=2) and getTileType(12,12)=0
	// => EQUAL on BOTH edges => NO split. So GO splits a tile the deob does not.
	{
		ov := make([]byte, Size*Size)
		for x := 6; x <= 17; x++ {
			for y := 6; y <= 17; y++ {
				ov[idx(x, y)] = 7
			}
		}
		// the L of id19 forming a concave corner around (12,12)
		for y := 6; y <= 17; y++ {
			ov[idx(11, y)] = 19 // a full N-S column of id19 at x=11 (left of 12)
		}
		for x := 12; x <= 17; x++ {
			ov[idx(x, 11)] = 19 // a full E-W row of id19 at y=11 (below 12)
		}
		base(ov).Save("terrain_class_corner.json")
	}
	// CONTROL: identical geometry but the L is id 5 (type 2, indoor) instead of
	// id 19 (type 3). At (12,12) id7(type3): getTileType neighbour(11,12)=id5 is
	// type2 => getTileType=1, vs self type3 => getTileType=0; 1 != 0 => DEOB SPLITS.
	// GO also splits (raw 5 != 7). So this control SHOULD agree GO==DEOB (both split)
	// — it isolates that the divergence is the type-EQUAL case, not the corner geometry.
	{
		ov := make([]byte, Size*Size)
		for x := 6; x <= 17; x++ {
			for y := 6; y <= 17; y++ {
				ov[idx(x, y)] = 7
			}
		}
		for y := 6; y <= 17; y++ {
			ov[idx(11, y)] = 5
		}
		for x := 12; x <= 17; x++ {
			ov[idx(x, 11)] = 5
		}
		base(ov).Save("terrain_class_corner_ctrl.json")
	}
}
