package pathfind

import "testing"

// TestBuildGrid_PlaneAware guards the plane-0-lock fix: an upper-floor position
// must build collision from that floor's archive sectors (h{plane}), not snap
// to the ground floor (which produced an empty wall grid — the "boxed upstairs"
// deadlock). Uses the real Lumbridge-castle column, which exists on planes 0..3
// at the same archive (sx,sy). Skips when the landscape fixture is absent.
func TestBuildGrid_PlaneAware(t *testing.T) {
	f, l := loadTestWorld(t)

	const gx, gy = 134, 644             // Lumbridge castle, ground floor (plane 0)
	ux, uy := gx, gy+planeHeight        // same footprint, 2nd floor (plane 1)

	g0, err := BuildGrid(l, f, gx, gy, 0, nil)
	if err != nil {
		t.Fatalf("ground grid: %v", err)
	}
	g1, err := BuildGrid(l, f, ux, uy, 0, nil) // plane auto-derived from Y
	if err != nil {
		t.Fatalf("upstairs grid: %v", err)
	}

	// Band invariance: same footprint -> identical local player placement, with
	// the grid origins exactly one 944-band apart.
	if g0.PlayerX != g1.PlayerX || g0.PlayerY != g1.PlayerY {
		t.Fatalf("player placement differs across planes: ground=(%d,%d) upstairs=(%d,%d)",
			g0.PlayerX, g0.PlayerY, g1.PlayerX, g1.PlayerY)
	}
	if g1.BaseY-g0.BaseY != planeHeight {
		t.Fatalf("base-Y band gap = %d, want %d", g1.BaseY-g0.BaseY, planeHeight)
	}
	if g1.PlayerX < 0 || g1.PlayerX >= GridSize || g1.PlayerY < 0 || g1.PlayerY >= GridSize {
		t.Fatalf("upstairs player out of grid: (%d,%d)", g1.PlayerX, g1.PlayerY)
	}

	// The upstairs grid must carry real structural walls — the Lumbridge castle
	// 2nd floor (h1x50y50). A plane-0-locked BuildGrid produced an all-zero
	// wall grid here (every upstairs walk collision-checked the wrong floor).
	if wallCells(g1) == 0 {
		t.Fatal("upstairs (plane 1) grid has zero wall flags — plane lock not fixed")
	}
}

func wallCells(g *Grid) int {
	n := 0
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			if g.Mask[x][y] != 0 {
				n++
			}
		}
	}
	return n
}
