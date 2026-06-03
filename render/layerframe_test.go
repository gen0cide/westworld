package render

import "testing"

// TestFacingPoseFlip locks facingPose's 8-way -> (i2, flip) mapping (the authentic
// drawNpc/drawPlayer rule, Mudclient.java:5866/5670): dirs 0..4 pass through unflipped,
// dir 5/6/7 collapse onto i2 3/2/1 WITH a horizontal flip (W/SW/NW are mirror images of
// E/SE/NE). This is the column the layer-frame formula multiplies by 3.
func TestFacingPoseFlip(t *testing.T) {
	cases := []struct {
		dir, wantI2 int
		wantFlip    bool
	}{
		{0, 0, false},
		{1, 1, false},
		{2, 2, false},
		{3, 3, false},
		{4, 4, false},
		{5, 3, true},
		{6, 2, true},
		{7, 1, true},
	}
	for _, c := range cases {
		i2, flip := facingPose(c.dir)
		if i2 != c.wantI2 || flip != c.wantFlip {
			t.Errorf("facingPose(%d) = (i2=%d, flip=%v), want (i2=%d, flip=%v)",
				c.dir, i2, flip, c.wantI2, c.wantFlip)
		}
	}
}

// npcWalkModelTest is the authentic sf={0,1,2,1} walk-cycle model (Mudclient.java:1929).
// The raw walk-cycle index (env :step 0..3) indexes it to the RESOLVED frame value the
// layer blit consumes as `step`. Kept local to the test so the lock is self-contained.
var npcWalkModelTest = [4]int{0, 1, 2, 1}

// TestLayerFrameFormula locks layerFrame(animID, i2, step, flip) against the AUTHENTIC
// per-layer frame formula j2 = i2*3 + step, plus the +15 F-frame offset that fires ONLY
// when the pose is flipped (i2 in 1..3) AND the animation has an F-frame (entityFlags[id]
// == 1). It sweeps all 8 facing dirs (via facingPose -> i2/flip) x the four walk-cycle
// values (sf-resolved {0,1,2,1}) for a hasF animID (134 skelweap) and a non-hasF one
// (123 rat), so a regression in either the base frame or the F-branch is caught.
func TestLayerFrameFormula(t *testing.T) {
	const (
		hasFAnim = 134 // skelweap: entityFlags==1, the canonical +15 F-frame animation
		noFAnim  = 123 // rat: no F-frame (the current dir-0 regression entity)
	)
	// Guard the fixture: the test is only meaningful if the table flags match the plan.
	if !authenticAnimDefs[hasFAnim].hasF {
		t.Fatalf("fixture: authenticAnimDefs[%d] (%q) must be hasF", hasFAnim, authenticAnimDefs[hasFAnim].name)
	}
	if authenticAnimDefs[noFAnim].hasF {
		t.Fatalf("fixture: authenticAnimDefs[%d] (%q) must NOT be hasF", noFAnim, authenticAnimDefs[noFAnim].name)
	}

	for dir := 0; dir < 8; dir++ {
		i2, flip := facingPose(dir)
		for _, raw := range []int{0, 1, 2, 3} {
			step := npcWalkModelTest[raw&3]
			base := i2*3 + step

			// hasF animation: +15 iff flipped and i2 in 1..3 (dirs 5/6/7).
			wantHasF := base
			if flip && i2 >= 1 && i2 <= 3 {
				wantHasF += 15
			}
			if got := layerFrame(hasFAnim, i2, step, flip); got != wantHasF {
				t.Errorf("layerFrame(hasF=%d, i2=%d, step=%d, flip=%v) [dir %d raw %d] = %d, want %d",
					hasFAnim, i2, step, flip, dir, raw, got, wantHasF)
			}

			// non-hasF animation: NEVER takes the +15, even when flipped.
			if got := layerFrame(noFAnim, i2, step, flip); got != base {
				t.Errorf("layerFrame(noF=%d, i2=%d, step=%d, flip=%v) [dir %d raw %d] = %d, want %d (no +15)",
					noFAnim, i2, step, flip, dir, raw, got, base)
			}
		}
	}
}

// TestLayerFrameFFrameOnlyFlipped pins the two halves of the F-frame gate independently:
// the +15 fires for a hasF animation ONLY when flipped (it must NOT fire for an unflipped
// pose at the same i2), proving the flip flag — not merely the column — gates the offset.
func TestLayerFrameFFrameOnlyFlipped(t *testing.T) {
	const hasFAnim = 134
	// i2=3 reachable both unflipped (dir 3) and flipped (dir 5).
	unflipped := layerFrame(hasFAnim, 3, 0, false) // dir 3
	flipped := layerFrame(hasFAnim, 3, 0, true)    // dir 5
	if unflipped != 3*3+0 {
		t.Errorf("unflipped hasF i2=3 step=0 = %d, want %d (no +15)", unflipped, 9)
	}
	if flipped != 3*3+0+15 {
		t.Errorf("flipped hasF i2=3 step=0 = %d, want %d (+15)", flipped, 24)
	}
}

// TestWalkWrapConsistency pins the GAP-A walk-wrap contract: raw cycle index 3 resolves
// (via sf={0,1,2,1}) to the SAME frame as raw 1 (sf[3]==sf[1]==1), so :step 3 is a
// wrap-consistency check of :step 1, NOT a 4th distinct frame.
func TestWalkWrapConsistency(t *testing.T) {
	if npcWalkModelTest[3] != npcWalkModelTest[1] {
		t.Fatalf("sf[3]=%d must equal sf[1]=%d (walk wrap)", npcWalkModelTest[3], npcWalkModelTest[1])
	}
	for dir := 0; dir < 8; dir++ {
		i2, flip := facingPose(dir)
		f1 := layerFrame(134, i2, npcWalkModelTest[1], flip)
		f3 := layerFrame(134, i2, npcWalkModelTest[3], flip)
		if f1 != f3 {
			t.Errorf("dir %d: layerFrame at raw1=%d != raw3=%d (walk wrap broken)", dir, f1, f3)
		}
	}
}
