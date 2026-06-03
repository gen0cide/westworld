package render

import (
	"os"
	"testing"
)

// content8_test.go — the content8 == Authentic_Sprites byte-equality cross-check for
// the ground-item pipeline (Task B2 step 4), the item-sprite twin of the content1 ==
// Authentic entity cross-check. It proves the rev-235 cache content8 "2d graphics"
// item-picture decode (content8ItemFrame) produces the IDENTICAL trimmed bitmap that
// OpenRSC's Authentic_Sprites.orsc[spriteItem+pic] holds — so the orsc raw item layer
// the AddEntityLayers blit consumes is byte-identical to the source the DEOB/JAR B1
// legs decode (single-source: content8 pixels == Authentic_Sprites pixels).
//
// SKIPS when either source is unavailable (RSC_MESH_CACHE unset / content8 absent, or
// Authentic_Sprites.orsc not found) — both are needed to compare.

// TestContent8EqualsAuthentic decodes item 14's picture (pic 14, mask 0 — pure
// decode, no recolour) from BOTH content8 and Authentic_Sprites and asserts the
// trimmed bitmap matches: same trimmed W/H, same trim offset (TX/TY == XShift/YShift),
// and identical per-pixel RGB (with the two sources' transparency conventions
// reconciled — content8 keys palette index 0 -> our -1 sentinel; Authentic_Sprites
// keys BLACK 0x000000). A pixel transparent in one must be transparent in the other,
// and every opaque pixel's RGB must be equal.
func TestContent8EqualsAuthentic(t *testing.T) {
	cache := os.Getenv("RSC_MESH_CACHE")
	if cache == "" {
		cache = "/tmp/rsc-run/cache"
		t.Setenv("RSC_MESH_CACHE", cache)
	}

	const itemID = 14
	icon, ok := itemIcons[itemID]
	if !ok {
		t.Fatalf("itemIcons missing item %d", itemID)
	}

	f := content8ItemFrame(icon.pic)
	if f == nil {
		t.Skipf("content8 unavailable (cache=%q) — cannot run the content8==Authentic cross-check", cache)
	}

	sa := sprites()
	if sa == nil {
		t.Skip("Authentic_Sprites.orsc not found (set WESTWORLD_SPRITES_ORSC) — cannot run the cross-check")
	}
	sp, err := sa.Sprite(spriteItem + icon.pic)
	if err != nil || sp == nil {
		t.Skipf("Authentic_Sprites has no sprite %d: %v", spriteItem+icon.pic, err)
	}

	// 1. Trimmed dimensions must match.
	if f.w != sp.Width || f.h != sp.Height {
		t.Fatalf("trimmed dims mismatch: content8 %dx%d vs Authentic %dx%d", f.w, f.h, sp.Width, sp.Height)
	}
	// 2. Trim offset must match (Authentic only reports it when RequiresShift).
	aTX, aTY := 0, 0
	if sp.RequiresShift {
		aTX, aTY = sp.XShift, sp.YShift
	}
	if f.tx != aTX || f.ty != aTY {
		t.Fatalf("trim offset mismatch: content8 (%d,%d) vs Authentic (%d,%d)", f.tx, f.ty, aTX, aTY)
	}

	// 3. Per-pixel RGB equality, transparency conventions reconciled.
	ndiff := 0
	for i := 0; i < f.w*f.h; i++ {
		c8 := f.pix[i] // -1 transparent, else 0x00RRGGBB
		auth := int(sp.Pixels[i]) & 0xffffff
		authTransparent := auth == 0 // Authentic_Sprites keys BLACK as transparent
		c8Transparent := c8 < 0
		if c8Transparent != authTransparent {
			ndiff++
			if ndiff <= 5 {
				t.Errorf("pixel %d transparency mismatch: content8 transparent=%v (%d) vs Authentic transparent=%v (%06x)",
					i, c8Transparent, c8, authTransparent, auth)
			}
			continue
		}
		if c8Transparent {
			continue // both transparent
		}
		if c8 != auth {
			ndiff++
			if ndiff <= 5 {
				t.Errorf("pixel %d RGB mismatch: content8 %06x vs Authentic %06x", i, c8, auth)
			}
		}
	}
	if ndiff != 0 {
		t.Fatalf("content8 != Authentic for item %d (pic %d): %d/%d pixels differ", itemID, icon.pic, ndiff, f.w*f.h)
	}
	t.Logf("content8==Authentic OK for item %d (pic %d): trimmed %dx%d full %dx%d trans (%d,%d), all %d pixels byte-identical",
		itemID, icon.pic, f.w, f.h, f.fullW, f.fullH, f.tx, f.ty, f.w*f.h)
}
