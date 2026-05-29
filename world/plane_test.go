package world

import "testing"

// TestPlaneOf checks the floor derivation: plane = Y / PlaneHeight
// (944), matching OpenRSC Formulae.getHeight.
func TestPlaneOf(t *testing.T) {
	cases := []struct {
		y    int
		want int
	}{
		{0, 0},       // ground level
		{943, 0},     // last tile of ground floor
		{944, 1},     // first upper floor
		{1887, 1},    // last tile of floor 1
		{1888, 2},    // floor 2
		{2832, 3},    // floor 3 (underground band)
		{-5, 0},      // guard: negatives clamp to ground
	}
	for _, c := range cases {
		if got := PlaneOf(c.y); got != c.want {
			t.Errorf("PlaneOf(%d): got %d, want %d", c.y, got, c.want)
		}
	}
}

// TestSelfPlane checks that Self.Plane derives the floor from the
// current position's Y (feeds the self.position.plane accessor).
func TestSelfPlane(t *testing.T) {
	s := NewSelf()
	s.SetPosition(Coord{X: 220, Y: 944 + 100}) // floor 1
	if got := s.Plane(); got != 1 {
		t.Errorf("Self.Plane on floor 1: got %d, want 1", got)
	}
	if got := s.Position().Plane(); got != 1 {
		t.Errorf("Coord.Plane on floor 1: got %d, want 1", got)
	}
	s.SetPosition(Coord{X: 220, Y: 300}) // ground
	if got := s.Plane(); got != 0 {
		t.Errorf("Self.Plane on ground: got %d, want 0", got)
	}
}
