package world

import "sync"

// Coord is an RSC world tile coordinate. RSC uses big-endian uint16
// for both x and y on the wire. We keep them as ints here for ergonomic
// arithmetic.
type Coord struct {
	X int
	Y int
}

// Self is the host's view of its own player. Read/write safe across
// goroutines.
//
// Phase 0 holds only Position. Subsequent phases extend with HP, prayer,
// fatigue, poison state, inventory, equipped items, combat target, etc.
type Self struct {
	mu       sync.RWMutex
	position Coord
}

// NewSelf returns a Self with a zero position. Caller should update
// position from the first inbound mob-update packet.
func NewSelf() *Self { return &Self{} }

// Position returns the current believed position.
func (s *Self) Position() Coord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.position
}

// SetPosition updates the position. Called when an inbound packet
// confirms a position change.
func (s *Self) SetPosition(c Coord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.position = c
}
