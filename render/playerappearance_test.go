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

// TestCompositePlayerAppearanceRealOutfit composites a real worn outfit when
// the entity archives are available, asserting a non-nil sprite with sane
// dimensions. SKIPS when the archives aren't present so the test stays green
// on machines without the deob/cache.
func TestCompositePlayerAppearanceRealOutfit(t *testing.T) {
	entityArchiveOnce.Do(loadEntityArchive)
	if entityArc == nil {
		t.Skip("entity24.jag / config85.jag not found; composite unavailable")
	}
	// Default-human layers are legs1/body1/head1. Find their animation ids and
	// build an equipment array whose equip[layer]-1 == that id, placing each in
	// a layer the npcAnimationArray draw order visits.
	idLegs, okLegs := entityArc.animByName["legs1"]
	idBody, okBody := entityArc.animByName["body1"]
	idHead, okHead := entityArc.animByName["head1"]
	if !okLegs || !okBody || !okHead {
		t.Skip("default-human animations not present in archive")
	}
	var eq [12]int
	// EquipSlotPants(2)=legs, EquipSlotShirt(1)=body, EquipSlotHead(0)=head —
	// the slots the default player layers occupy (legs/body/head). Store id+1
	// so equip[layer]-1 recovers the animation id.
	eq[2] = idLegs + 1
	eq[1] = idBody + 1
	eq[0] = idHead + 1

	cs := compositePlayerAppearance(eq, 2, 8, 14, 0, 0, 0)
	if cs == nil {
		t.Fatal("real outfit: got nil composite, want a sprite")
	}
	if cs.W <= 0 || cs.H <= 0 || len(cs.Pix) != cs.W*cs.H {
		t.Fatalf("real outfit: bad dims W=%d H=%d len(pix)=%d", cs.W, cs.H, len(cs.Pix))
	}
}
