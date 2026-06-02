package render

import "testing"

// TestCompositePlayerAppearanceEmptyOutfit verifies that an all-zero
// equipment array (nothing worn / no decodable layer) yields nil rather than
// panicking, so the caller falls back to the default-human composite. This
// holds whether or not the entity archives are present.
func TestCompositePlayerAppearanceEmptyOutfit(t *testing.T) {
	var empty [12]int
	if cs := compositePlayerAppearance(empty, 0, 0, 0, 0, 0, 0); cs != nil {
		t.Fatalf("empty outfit: got a composite, want nil (fall back to default human)")
	}
}

// TestCompositePlayerAppearanceNeverPanics fuzzes a few facing directions and
// out-of-range sprite/colour indices to confirm the composite path is panic-
// free and bounded (it returns nil on failure, never crashes).
func TestCompositePlayerAppearanceNeverPanics(t *testing.T) {
	outfits := [][12]int{
		{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}, // a plausible worn set
		{255, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},  // a likely out-of-range sprite
		{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},    // empty
	}
	for _, eq := range outfits {
		for dir := 0; dir < 8; dir++ {
			// Out-of-range colour indices must be tolerated (resolveClothingColour
			// treats them as direct colours; skin index is range-checked).
			_ = compositePlayerAppearance(eq, 2, 8, 14, 1, dir, 0)
			_ = compositePlayerAppearance(eq, 999, 999, 999, 999, dir, 0)
		}
	}
}

// TestCompositePlayerAppearanceRealOutfit composites a real worn outfit from
// Authentic_Sprites.orsc, asserting a non-nil sprite with sane dimensions.
// SKIPS when the sprite archive isn't present so the test stays green on
// machines without the OpenRSC cache.
func TestCompositePlayerAppearanceRealOutfit(t *testing.T) {
	if sprites() == nil {
		t.Skip("Authentic_Sprites.orsc not found; composite unavailable")
	}
	// Default-human animation ids are head1=0, body1=1, legs1=2 (authenticAnimDefs
	// order). Build an equipment array whose equip[layer]-1 == that id, placing
	// each in the layer the npcAnimationArray draw order visits. Store id+1 so
	// equip[layer]-1 recovers the animation id.
	var eq [12]int
	eq[0] = 0 + 1 // head layer -> head1
	eq[1] = 1 + 1 // body layer -> body1
	eq[2] = 2 + 1 // legs layer -> legs1

	cs := compositePlayerAppearance(eq, 2, 8, 14, 0, 0, 0)
	if cs == nil {
		t.Fatal("real outfit: got nil composite, want a sprite")
	}
	if cs.W <= 0 || cs.H <= 0 || len(cs.Pix) != cs.W*cs.H {
		t.Fatalf("real outfit: bad dims W=%d H=%d len(pix)=%d", cs.W, cs.H, len(cs.Pix))
	}
}
