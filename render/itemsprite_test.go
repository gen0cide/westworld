package render

import (
	"os"
	"testing"

	"github.com/gen0cide/westworld/facts"
)

const testOpenRSCRoot = "/Users/flint/Code/openrsc"

func loadTestFacts(t *testing.T) *facts.Facts {
	t.Helper()
	src := facts.DefaultSources(testOpenRSCRoot)
	if _, err := os.Stat(src.ItemDefsJSON); err != nil {
		t.Skip("OpenRSC ItemDefs.json not found; item defs unavailable")
	}
	f, err := facts.Load(src)
	if err != nil {
		t.Skipf("facts.Load failed: %v", err)
	}
	return f
}

// TestCompositeItemNeverPanics fuzzes a range of item ids (including negatives
// and an absurd out-of-range id) to confirm the item-icon path is panic-free
// and bounded: it returns nil on failure, never crashes. nil facts also returns
// nil. Holds whether or not Authentic_Sprites.orsc is present.
func TestCompositeItemNeverPanics(t *testing.T) {
	for _, id := range []int{-1, 0, 1, 10, 20, 167, 526, 1000, 100000, -999} {
		_ = compositeItem(nil, id)
	}
}

// TestCompositeItemNegativeID asserts a negative item id always yields nil
// (no archive lookup), independent of whether the archive loads.
func TestCompositeItemNegativeID(t *testing.T) {
	if cs := compositeItem(nil, -1); cs != nil {
		t.Fatalf("compositeItem(nil,-1) = %v, want nil", cs)
	}
}

// TestCompositeItemRealIcons decodes a handful of known item icons from OpenRSC's
// Authentic_Sprites.orsc, asserting non-nil composites on the 48x32 inventory
// canvas with at least one opaque pixel, and that the icon is cached (the second
// call returns the same pointer). SKIPS when the OpenRSC assets aren't present.
func TestCompositeItemRealIcons(t *testing.T) {
	f := loadTestFacts(t)
	if sprites() == nil {
		t.Skip("Authentic_Sprites.orsc not found; item icons unavailable")
	}

	// Iron Mace (0), Coins (10), Bones (20) are stable low ids in the OpenRSC
	// item defs; each must decode onto the shared 48x32 inventory canvas.
	for _, id := range []int{0, 10, 20} {
		cs := compositeItem(f, id)
		if cs == nil {
			t.Errorf("item %d: got nil composite, want decoded icon", id)
			continue
		}
		if cs.W != itemIconW || cs.H != itemIconH || len(cs.Pix) != cs.W*cs.H || len(cs.Opaque) != cs.W*cs.H {
			t.Errorf("item %d: bad dims W=%d H=%d len(Pix)=%d len(Opaque)=%d",
				id, cs.W, cs.H, len(cs.Pix), len(cs.Opaque))
		}
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
		if cs2 := compositeItem(f, id); cs2 != cs {
			t.Errorf("item %d: second compositeItem returned a different pointer (not cached)", id)
		}
	}
}
