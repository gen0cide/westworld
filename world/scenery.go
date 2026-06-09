package world

import (
	"sync"
	"time"
)

// SceneryRemoveSentinel is the special object id the server sends in
// SEND_SCENERY_HANDLER (opcode 48) to mean "the scenery at this tile
// is gone" (mined rock depleted, fire burned out, door-frame removed).
// Source: GameStateUpdater.updateGameObjects — it adds a
// GameObjectLoc(60000, …) for every object leaving view/removed.
const SceneryRemoveSentinel = 60000

// SceneryRecord is one dynamic scenery (GameObject) visible in the
// bot's local view. Keyed by absolute world (X, Y) — only one
// non-wall scenery object occupies a tile at a time in RSC.
type SceneryRecord struct {
	X, Y     int
	ID       int
	LastSeen time.Time
}

// DynamicScenery mirrors the in-view scenery (GameObject) state the
// server streams via SEND_SCENERY_HANDLER (opcode 48). This is the
// ONLY place runtime-spawned scenery shows up: lit fires (def 97),
// freshly-cut/regrown trees, etc. The static facts.SceneryLocs map
// does NOT contain these — a fire lit by firemaking is registered at
// runtime via World.registerGameObject and only ever reaches the
// client through this packet, never the static landscape data.
//
// Mirrors GroundItemsState shape (tile-keyed map) so the consult
// surface (.All / .Near / .At) parallels world.ground_items.
type DynamicScenery struct {
	mu sync.RWMutex
	m  map[[2]int]SceneryRecord
	// removed retains tiles the server explicitly cleared via the 60000
	// removal sentinel (a depleted rock / burned-out fire / removed
	// door-frame). The live object map `m` deletes the record on removal
	// (so At/Near/All report the object as gone, the by_id accessor
	// contract), but the RENDERER needs to know the tile was actively
	// removed so it can SUPPRESS the static baseline scenery that the
	// .orsc landscape still carries there — otherwise a mined rock would
	// pop back. Adding a live object at the tile clears the mark.
	removed map[[2]int]bool
}

func NewDynamicScenery() *DynamicScenery {
	return &DynamicScenery{m: map[[2]int]SceneryRecord{}, removed: map[[2]int]bool{}}
}

// Add records (or replaces) the scenery object at (x, y).
func (s *DynamicScenery) Add(x, y, id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[[2]int{x, y}] = SceneryRecord{X: x, Y: y, ID: id, LastSeen: time.Now()}
	delete(s.removed, [2]int{x, y}) // a live object cancels a prior removal
}

// Remove drops any scenery record at (x, y) — used for the 60000
// removal sentinel — and marks the tile as actively removed so the
// renderer suppresses the static baseline scenery there.
func (s *DynamicScenery) Remove(x, y int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, [2]int{x, y})
	s.removed[[2]int{x, y}] = true
}

// IsRemoved reports whether the scenery at (x, y) was actively cleared by
// the server's removal sentinel (so the static baseline object should be
// suppressed by the renderer). False for tiles never touched.
func (s *DynamicScenery) IsRemoved(x, y int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.removed[[2]int{x, y}]
}

// RemoveRegion drops every live record AND removal mark whose tile falls in
// the 8x8 region (rx, ry) = (x>>3, y>>3). Backs the opcode-211 bulk clear:
// the server cleaned up its per-player state for that region, so our
// overrides there are stale — dropping them returns those tiles to the
// static baseline until fresh state streams in on re-approach.
func (s *DynamicScenery) RemoveRegion(rx, ry int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k := range s.m {
		if k[0]>>3 == rx && k[1]>>3 == ry {
			delete(s.m, k)
		}
	}
	for k := range s.removed {
		if k[0]>>3 == rx && k[1]>>3 == ry {
			delete(s.removed, k)
		}
	}
}

// Removed returns every tile the server has actively cleared via the removal
// sentinel (a depleted rock / burned-out fire / removed door-frame). Lets the
// pathfinder drop the static baseline block at those tiles so the BFS can walk
// over a now-gone object.
func (s *DynamicScenery) Removed() [][2]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([][2]int, 0, len(s.removed))
	for k := range s.removed {
		out = append(out, k)
	}
	return out
}

// At returns (record, true) if dynamic scenery is known at (x, y).
func (s *DynamicScenery) At(x, y int) (SceneryRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.m[[2]int{x, y}]
	return r, ok
}

// All returns a snapshot of every currently-tracked scenery object.
func (s *DynamicScenery) All() []SceneryRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]SceneryRecord, 0, len(s.m))
	for _, r := range s.m {
		out = append(out, r)
	}
	return out
}

// Near returns scenery objects within `radius` Chebyshev tiles of
// (cx, cy).
func (s *DynamicScenery) Near(cx, cy, radius int) []SceneryRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []SceneryRecord
	for _, r := range s.m {
		dx := r.X - cx
		if dx < 0 {
			dx = -dx
		}
		dy := r.Y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx <= radius && dy <= radius {
			out = append(out, r)
		}
	}
	return out
}
