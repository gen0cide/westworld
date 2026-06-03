//go:build ignore

// gen_bridge authors the bridge hunt fixtures as rscdump/1 L1 JSON.
//
//	go run gen_bridge.go
//
// Scenario (Size x Size window, window-local [x*Size+y], x=col=worldX-BaseX):
//   - a uniform GroundElevation so the river/deck would normally sit ABOVE 0
//   - an E-W river band of overlay 2 (water, tileType 3) across the middle rows
//   - a N-S strip of overlay 250 (BRIDGE marker) crossing the river (the deck)
// Twin: bridge_off swaps the 250 deck strip for plain grass (overlay 0) to show
// the delta the deck SHOULD introduce.
package main

import (
	"github.com/gen0cide/westworld/internal/rscdump"
)

const (
	Size  = 24
	BaseX = 200 // 200%48 = 8 -> deck columns sit well INSIDE a sector (no %48==47 seam)
	BaseY = 200
	Elev  = 40 // *3 = 120 world height: deck/river clearly above y=0 if NOT flattened
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

func base(overlay []byte) *rscdump.Dump {
	n := Size * Size
	elev := make([]byte, n)
	gcol := make([]byte, n)
	for i := range elev {
		elev[i] = Elev
		gcol[i] = 70 // a green-ish grass palette index
	}
	return &rscdump.Dump{
		Schema: rscdump.SchemaID,
		Level:  rscdump.LevelL1,
		Source: rscdump.SourceHandAuthored,
		Camera: cam(),
		Window: rscdump.Window{BaseX: BaseX, BaseY: BaseY, Plane: 0, Size: Size},
		Terrain: &rscdump.Terrain{
			Size:        Size,
			Elevation:   elev,
			GroundColour: gcol,
			Overlay:     overlay,
			TerrainSeed: 0,
		},
		Self: &rscdump.Self{X: BaseX + Size/2, Y: BaseY + Size/2, NoSelf: true},
	}
}

// riverRows is the band of water rows; deckCols is the strip of bridge columns.
func makeOverlay(withDeck bool) []byte {
	n := Size * Size
	ov := make([]byte, n)
	riverY0, riverY1 := 10, 13 // water band (overlay 2)
	deckX0, deckX1 := 11, 12   // bridge strip (overlay 250) crossing the river
	for x := 0; x < Size; x++ {
		for y := 0; y < Size; y++ {
			if y >= riverY0 && y <= riverY1 {
				ov[idx(x, y)] = 2 // water
			}
			if withDeck && x >= deckX0 && x <= deckX1 {
				// the deck strip runs the FULL length so it extends onto dry land
				// at both ends (so we see seam-edge vs interior behaviour)
				ov[idx(x, y)] = 250
			}
		}
	}
	return ov
}

func main() {
	on := base(makeOverlay(true))
	off := base(makeOverlay(false))
	must(on.Save("bridge_on.json"))
	must(off.Save("bridge_off.json"))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
