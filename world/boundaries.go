package world

import "sync"

// BoundaryKey identifies a boundary by its absolute (X, Y, Dir).
type BoundaryKey struct {
	X, Y, Dir int
}

// DynamicBoundaries tracks boundary state changes (doors opened,
// webs cut, etc.) relative to the static facts.BoundaryLocs map.
//
// Routines that want to walk through a known door should:
//   1. use(key, door) — server emits SEND_BOUNDARY_HANDLER with
//      the door's new id (or -1 if it just opens)
//   2. World.Apply ingests the event and updates this state
//   3. walk_to / DSL can consult world.boundaries.is_open(x, y, dir)
//      to know the door is now passable
//
// id < 0 in the map means "removed" (passable). id >= 0 means the
// current dynamic def replacing whatever facts.BoundaryLocs had.
type DynamicBoundaries struct {
	mu        sync.RWMutex
	overrides map[BoundaryKey]int // id (-1 = removed)
}

func NewDynamicBoundaries() *DynamicBoundaries {
	return &DynamicBoundaries{overrides: map[BoundaryKey]int{}}
}

// Set records a dynamic boundary update. id = -1 means the
// boundary at this tile/dir was removed.
func (d *DynamicBoundaries) Set(x, y, dir, id int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.overrides[BoundaryKey{X: x, Y: y, Dir: dir}] = id
}

// Get returns (id, true) if there's a dynamic override at this
// (x, y, dir). id = -1 means removed. Returns (0, false) if no
// override exists — callers should fall back to facts.
func (d *DynamicBoundaries) Get(x, y, dir int) (int, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	id, ok := d.overrides[BoundaryKey{X: x, Y: y, Dir: dir}]
	return id, ok
}

// IsRemoved is the common-case helper: "is the boundary at this
// tile/dir gone (door open / web cut)?" Returns false if there's
// no override OR the override sets a different id (door state
// changed but still blocks).
func (d *DynamicBoundaries) IsRemoved(x, y, dir int) bool {
	id, ok := d.Get(x, y, dir)
	return ok && id < 0
}

// Clear removes any override at (x, y, dir).
func (d *DynamicBoundaries) Clear(x, y, dir int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.overrides, BoundaryKey{X: x, Y: y, Dir: dir})
}

// RemoveRegion drops every override whose tile falls in the 8x8 region
// (rx, ry) = (x>>3, y>>3). Backs the opcode-211 bulk clear — the server's
// only eviction channel for far-away boundary state; without it stale
// open/closed door records accumulate forever.
func (d *DynamicBoundaries) RemoveRegion(rx, ry int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for k := range d.overrides {
		if k.X>>3 == rx && k.Y>>3 == ry {
			delete(d.overrides, k)
		}
	}
}

// All returns a snapshot of every active override.
func (d *DynamicBoundaries) All() map[BoundaryKey]int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make(map[BoundaryKey]int, len(d.overrides))
	for k, v := range d.overrides {
		out[k] = v
	}
	return out
}
