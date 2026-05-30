package render

import "testing"

// TestCompositeItemNeverPanics fuzzes a range of item ids (including negatives
// and an absurd out-of-range id) to confirm the item-icon path is panic-free
// and bounded: it returns nil on failure, never crashes. Holds whether or not
// the media58.jag / config85.jag archives are present.
func TestCompositeItemNeverPanics(t *testing.T) {
	for _, id := range []int{-1, 0, 1, 10, 20, 167, 526, 1000, 100000, -999} {
		_ = compositeItem(id)
	}
}

// TestCompositeItemNegativeID asserts a negative item id always yields nil
// (no archive lookup), independent of whether the archives load.
func TestCompositeItemNegativeID(t *testing.T) {
	if cs := compositeItem(-1); cs != nil {
		t.Fatalf("compositeItem(-1) = %v, want nil", cs)
	}
}

// TestCompositeItemRealIcons decodes a handful of known item icons when the
// archives are available, asserting non-nil composites with sane dimensions
// (item icons share the 48x32 inventory canvas) and that the icon is cached
// (the second call returns the same pointer). SKIPS when the archives aren't
// present so the test stays green on machines without the deob/cache.
func TestCompositeItemRealIcons(t *testing.T) {
	entityArchiveOnce.Do(loadEntityArchive)
	itemArchiveOnce.Do(loadItemArchive)
	if entityArc == nil || itemArc == nil {
		t.Skip("config85.jag / media58.jag not found; item icons unavailable")
	}

	// Iron Mace (0), Coins (10), Bones (20) are stable low ids present in every
	// classic config; each must decode onto the shared inventory canvas.
	for _, id := range []int{0, 10, 20} {
		cs := compositeItem(id)
		if cs == nil {
			t.Errorf("item %d: got nil composite, want decoded icon", id)
			continue
		}
		if cs.W <= 0 || cs.H <= 0 || len(cs.Pix) != cs.W*cs.H || len(cs.Opaque) != cs.W*cs.H {
			t.Errorf("item %d: bad dims W=%d H=%d len(Pix)=%d len(Opaque)=%d",
				id, cs.W, cs.H, len(cs.Pix), len(cs.Opaque))
		}
		// at least one opaque pixel (the icon isn't all transparent)
		opaque := false
		for _, o := range cs.Opaque {
			if o {
				opaque = true
				break
			}
		}
		if !opaque {
			t.Errorf("item %d: decoded icon has no opaque pixels", id)
		}
		// memoised: same pointer on a second call
		if cs2 := compositeItem(id); cs2 != cs {
			t.Errorf("item %d: second compositeItem returned a different pointer (not cached)", id)
		}
	}
}
