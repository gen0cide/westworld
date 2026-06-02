package assets

import (
	"os"
	"testing"
)

// the canonical OpenRSC sprite archive; the test skips if it isn't present
// locally (matching the repo convention for asset-dependent tests).
const testSpritesPath = "/Users/flint/Code/openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc"

func openTestSprites(t *testing.T) *SpriteArchive {
	t.Helper()
	if _, err := os.Stat(testSpritesPath); err != nil {
		t.Skip("Authentic_Sprites.orsc not found; sprite archive unavailable")
	}
	a, err := OpenSprites(testSpritesPath)
	if err != nil {
		t.Fatalf("OpenSprites: %v", err)
	}
	return a
}

func TestSpriteArchiveIndex(t *testing.T) {
	a := openTestSprites(t)
	defer a.Close()
	if a.Count() < 1000 {
		t.Errorf("expected a few thousand sprites, got %d", a.Count())
	}
	if !a.Has(0) {
		t.Error("expected sprite id 0 to exist")
	}
	if a.Has(9_999_999) {
		t.Error("did not expect sprite id 9999999 to exist")
	}
}

func TestSpriteUnpackTexture(t *testing.T) {
	a := openTestSprites(t)
	defer a.Close()

	// 3225 = spriteTexture base (a 128px wood/wall texture); 3226 = water (64px).
	for _, tc := range []struct {
		id, w, h, sizeClass int
	}{
		{3225, 128, 128, 1},
		{3226, 64, 64, 0},
	} {
		s, err := a.Sprite(tc.id)
		if err != nil {
			t.Fatalf("Sprite(%d): %v", tc.id, err)
		}
		if s == nil {
			t.Fatalf("Sprite(%d): nil (missing)", tc.id)
		}
		if s.Width != tc.w || s.Height != tc.h {
			t.Errorf("Sprite(%d): got %dx%d want %dx%d", tc.id, s.Width, s.Height, tc.w, tc.h)
		}
		if len(s.Pixels) != s.Width*s.Height {
			t.Errorf("Sprite(%d): %d pixels for %dx%d", tc.id, len(s.Pixels), s.Width, s.Height)
		}
		if got := s.Something1/64 - 1; got != tc.sizeClass {
			t.Errorf("Sprite(%d): size class %d want %d (Something1=%d)", tc.id, got, tc.sizeClass, s.Something1)
		}
		// the texture must carry real RGB content (not all zero)
		var nonzero int
		for _, p := range s.Pixels {
			if p&0x00ffffff != 0 {
				nonzero++
			}
		}
		if nonzero == 0 {
			t.Errorf("Sprite(%d): all pixels are black/zero", tc.id)
		}
	}
}

func TestSpriteMissingReturnsNil(t *testing.T) {
	a := openTestSprites(t)
	defer a.Close()
	s, err := a.Sprite(9_999_999)
	if err != nil {
		t.Errorf("missing sprite should not error, got %v", err)
	}
	if s != nil {
		t.Error("missing sprite should return nil")
	}
}
