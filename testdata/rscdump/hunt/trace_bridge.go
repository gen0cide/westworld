//go:build ignore

// trace_bridge loads a bridge fixture and prints, per deck/river tile, what the
// GO terrain builder decides (overlay remap, water flag, vertex flatten) so we
// can line it up against the deob World.buildSection spec by hand.
//
//	go run trace_bridge.go bridge_on.json
package main

import (
	"fmt"
	"os"

	"github.com/gen0cide/westworld/internal/rscdump"
)

// authentic tileType (GameData / OpenRSC TileDef.xml "unknown" field), id->type
var tileType = map[byte]byte{
	1: 1, 2: 3, 3: 2, 4: 4, 5: 2, 6: 2, 7: 3, 8: 5, 9: 1, 10: 5,
	11: 3, 12: 4, 13: 2, 14: 2, 15: 2, 16: 2, 17: 2, 18: 2, 19: 3,
	20: 4, 21: 4, 22: 0, 23: 2, 24: 1, 25: 3,
}

func main() {
	d, err := rscdump.Load(os.Args[1])
	if err != nil {
		panic(err)
	}
	t := d.Terrain
	n := t.Size
	base := d.Window.BaseX
	get := func(g []byte, x, y int) byte {
		if x < 0 || y < 0 || x >= n || y >= n {
			return 0
		}
		return g[x*n+y]
	}

	// Replicate the GO bridge remap (terrain.go:117-127): 250 -> 9 (seam) or 2.
	goOv := make([]byte, n*n)
	deobOv := make([]byte, n*n)
	copy(goOv, t.Overlay)
	copy(deobOv, t.Overlay)
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			if t.Overlay[x*n+y] != 250 {
				continue
			}
			wx, wy := base+x, d.Window.BaseY+y
			// GO uses wx%48==47 ; deob uses LOCAL x==47 inside the 96-region.
			switch {
			case wx%48 == 47 && !bw(get(t.Overlay, x+1, y)):
				goOv[x*n+y] = 9
			case wy%48 == 47 && !bw(get(t.Overlay, x, y+1)):
				goOv[x*n+y] = 9
			default:
				goOv[x*n+y] = 2
			}
			// deob setTiles (LOCAL coordinate seam, neighbour != 250 and != 2)
			switch {
			case x == 47 && get(t.Overlay, x+1, y) != 250 && get(t.Overlay, x+1, y) != 2:
				deobOv[x*n+y] = 9
			case y == 47 && get(t.Overlay, x, y+1) != 250 && get(t.Overlay, x, y+1) != 2:
				deobOv[x*n+y] = 9
			default:
				deobOv[x*n+y] = 2
			}
		}
	}

	fmt.Printf("BaseX=%d (BaseX%%48=%d)  Size=%d  Elev=%d (->h=%d)\n", base, base%48, n, t.Elevation[0], int(t.Elevation[0])*3)
	fmt.Println("Per-tile where raw overlay==250 OR ==2:")
	fmt.Println("  loc(x,y)  wx%48  raw  GOremap GOwater GOflatHt | DEOBremap DEOBtype DEOBflat DEOBht")
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			raw := t.Overlay[x*n+y]
			if raw != 250 && raw != 2 {
				continue
			}
			wx := base + x
			realH := int(t.Elevation[x*n+y]) * 3
			// GO: overlay 2 (and 11) => water => flatten to 0
			goW := goOv[x*n+y] == 2 || goOv[x*n+y] == 11
			goH := realH
			if goW {
				goH = 0
			}
			// DEOB: flatten only if THIS tile type==4 (or a neighbour) — here only THIS tile matters for the vertex; type of deob overlay:
			dtype := tileType[deobOv[x*n+y]]
			deobFlat := dtype == 4
			deobH := realH
			if deobFlat {
				deobH = 0
			}
			fmt.Printf("  (%2d,%2d)  %5d   %3d   %3d     %5v   %5d   |  %3d     %3d      %5v    %5d\n",
				x, y, wx%48, raw, goOv[x*n+y], goW, goH, deobOv[x*n+y], dtype, deobFlat, deobH)
		}
	}
}

func bw(o byte) bool { return o == 250 || o == 2 }
