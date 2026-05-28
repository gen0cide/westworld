package pathfind

import (
	"testing"
)

const testLandscape = "/Users/flint/Code/openrsc/server/conf/server/data/Authentic_Landscape.orsc"

func TestSectorForWorld(t *testing.T) {
	tests := []struct {
		worldX, worldY int
		wantSX, wantSY int
	}{
		// Plane 0 (0, 0) lives in h0x48y37 per the archive's offset
		// convention.
		{0, 0, 48, 37},
		// Lumbridge area around (130, 645) → h0x50y50.
		{130, 645, 50, 50},
		{134, 644, 50, 50},
		// Edge of sector h0x50y50: tile (143, 671) is last.
		{143, 671, 50, 50},
		{144, 672, 51, 51},
	}
	for _, tt := range tests {
		got := SectorForWorld(tt.worldX, tt.worldY, 0)
		if got.SX != tt.wantSX || got.SY != tt.wantSY {
			t.Errorf("SectorForWorld(%d, %d) = (%d, %d), want (%d, %d)",
				tt.worldX, tt.worldY, got.SX, got.SY, tt.wantSX, tt.wantSY)
		}
	}
}

func TestLoadLandscape(t *testing.T) {
	l, err := OpenLandscape(testLandscape)
	if err != nil {
		t.Skipf("landscape archive not present at %s: %v", testLandscape, err)
	}
	defer l.Close()

	// Lumbridge sector should load.
	s, err := l.Sector(SectorKey{Plane: 0, SX: 50, SY: 50})
	if err != nil {
		t.Fatalf("load Lumbridge sector: %v", err)
	}
	if s == nil {
		t.Fatal("Lumbridge sector h0x50y50 missing from archive")
	}

	// Sanity-check: at least one tile should have non-zero ground
	// elevation (terrain isn't flat zero everywhere).
	any := false
	for _, tile := range s.Tiles {
		if tile.GroundElevation != 0 || tile.GroundTexture != 0 {
			any = true
			break
		}
	}
	if !any {
		t.Error("Lumbridge sector decoded to all-zero tiles — likely a parser bug")
	}
}
