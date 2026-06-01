//go:build ignore

// trace_foot compares, on a SLOPED terrain, the foot-anchor Y the GO sprite
// path uses (heights[lx][ly] = getTerrainHeight(SW corner)*3) against the deob's
// getElevation(cx,cy) bilinear interpolation at the TILE CENTRE (cx=lx*128+64,
// cy=ly*128+64) — the world point Mudclient.java:7220-7223/7230-7233 passes to
// scene.f() for every npc/player/ground-item sprite. A nonzero delta is the
// sub-tile foot-placement divergence on slopes.
//
//	go run trace_foot.go
package main

import (
	"fmt"

	"github.com/gen0cide/westworld/internal/rscdump"
	"github.com/gen0cide/westworld/render"
)

const anInt585 = 128

// deob getTerrainHeight: elevation byte * 3 (corner). The grid here is the
// fixture's own 16x16; outside it the renderer pulls 0.
func main() {
	const size = 16
	const baseX, baseY, plane = 200, 200, 0

	// Build a ramp: elevation increases with local Y (a hill sloping away from
	// the host). elevation byte e -> height e*3.
	n := size * size
	idx := func(x, y int) int { return x*size + y }
	elev := make([]byte, n)
	ground := make([]byte, n)
	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			elev[idx(x, y)] = byte(y * 5) // ramp in +Y
			ground[idx(x, y)] = 70
		}
	}
	d := &rscdump.Dump{
		Schema: rscdump.SchemaID, Level: rscdump.LevelL1, Source: rscdump.SourceHandAuthored,
		Camera: rscdump.Camera{X: 0, Y: -384, Z: 0, Pitch: 912, Yaw: 512, Roll: 0, Distance: 1500, ViewDist: 9, ClipNear: 5, ClipFar: 7000, ScreenW: 512, ScreenH: 334},
		Window: rscdump.Window{BaseX: baseX, BaseY: baseY, Plane: plane, Size: size},
		Terrain: &rscdump.Terrain{Size: size, Elevation: elev, GroundColour: ground,
			Overlay: make([]byte, n), Roof: make([]byte, n), WallH: make([]byte, n), WallV: make([]byte, n), WallDiag: make([]int32, n), TerrainSeed: 0},
		Self: &rscdump.Self{X: baseX + size/2, Y: baseY + size/2, NoSelf: true},
	}
	land := d.Landscape()

	// GO heights grid (window-local). terrainSize=160 window centred on host.
	hostX, hostY := baseX+size/2, baseY+size/2
	gridBaseX := hostX - 80
	gridBaseY := hostY - 80
	heights := render.TerrainHeights(land, gridBaseX, gridBaseY, plane)

	// local elevation getter (corner height), matching getTerrainHeight: read the
	// landscape tile directly so it's the SAME source the GO terrain uses.
	hAt := func(wx, wy int) int {
		t := land.Tile(wx, wy, plane)
		return int(t.GroundElevation) * 3
	}
	getElevation := func(wx, wy int) int {
		sx, sy := wx>>7, wy>>7
		aX, aY := wx&0x7f, wy&0x7f
		var h, hx, hy int
		if aX <= anInt585-aY {
			h = hAt(sx, sy)
			hx = hAt(sx+1, sy) - h
			hy = hAt(sx, sy+1) - h
		} else {
			h = hAt(sx+1, sy+1)
			hx = hAt(sx, sy+1) - h
			hy = hAt(sx+1, sy) - h
			aX = anInt585 - aX
			aY = anInt585 - aY
		}
		return h + (hx*aX)/anInt585 + (hy*aY)/anInt585
	}

	fmt.Println("local tile (lx,ly) | GO foot heights[lx][ly] | deob -getElevation(centre) | delta")
	for ly := hostY - baseY; ly <= hostY-baseY+6; ly++ {
		wx, wy := hostX, baseY+ly
		lx := wx - gridBaseX
		gly := wy - gridBaseY
		goFoot := -int(heights[lx][gly])
		cx := wx*128 + 64
		cy := wy*128 + 64
		deobFoot := -getElevation(cx, cy)
		fmt.Printf("  world(%d,%d) lx,gly=(%d,%d) | GO %5d | deob %5d | delta %d\n",
			wx, wy, lx, gly, goFoot, deobFoot, goFoot-deobFoot)
	}
}
