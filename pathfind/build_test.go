package pathfind

import (
	"testing"
)

func TestBuildGridOnly(t *testing.T) {
	f, l := loadTestWorld(t)
	defer l.Close()

	g, err := BuildGrid(l, f, 134, 644, 0)
	if err != nil {
		t.Fatalf("BuildGrid: %v", err)
	}
	t.Logf("grid base=(%d, %d) player=(%d, %d)", g.BaseX, g.BaseY, g.PlayerX, g.PlayerY)
	// Sanity-check: at least one tile should have any flag set.
	count := 0
	for x := 0; x < GridSize; x++ {
		for y := 0; y < GridSize; y++ {
			if g.Mask[x][y] != 0 {
				count++
			}
		}
	}
	t.Logf("flagged tiles: %d / %d", count, GridSize*GridSize)
}
