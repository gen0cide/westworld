package runtime

import (
	"fmt"
	"sort"
	"strings"
)

// describeArea gives the planner FULL area context: every nearby NPC, object,
// scenery, and door with NAME and COORDINATES (nearest-first), unfiltered — so
// she can see the cooking range, fishing spot, rock, etc. and act on them
// (walk_to or use items on them by coordinate). The old notable-only filter hid
// interactables like the range; this does not.
func (d *MesaDirector) describeArea(h *Host, radius int) string {
	pos := h.world.Self.Position()
	// Record where she is so doors she's passed can be flagged later (bounded ring).
	if d.visited == nil {
		d.visited = newTileRing(visitedCap)
	}
	d.visited.add(pos.X, pos.Y)

	// Reachability annotation (#31): mark a nearby NPC/player/scenery the host CANNOT
	// path-nav to (a different walkable component — behind a wall, across a gap, up a
	// dead-end ladder) so the planner stops targeting unreachable things. It still
	// SEES the thing exists (unlike scan_for, which drops it from the target list),
	// but knows it can't act on it from here. Doors/boundaries are NOT marked — they
	// are the exits BETWEEN components.
	hostComp, reachGate := int32(-1), false
	if h.worldOracle != nil {
		if c, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y); ok {
			hostComp, reachGate = c, true
		}
	}
	reachLabel := func(label string, x, y int) string {
		if !reachGate || label == "" {
			return label
		}
		if c, _, _, ok := h.worldOracle.CompNear(x, y); !ok || c != hostComp {
			return label + " [CANNOT REACH from here — behind a wall / different floor; find a door or ladder out]"
		}
		return label
	}

	type obj struct {
		label string
		dist  int
	}
	var objs []obj
	seen := map[string]bool{}
	add := func(label string, x, y int) {
		if label == "" {
			return
		}
		full := fmt.Sprintf("%s @ (%d,%d), to your %s", label, x, y, bearingFrom(pos.X, pos.Y, x, y))
		if seen[full] { // dedup by name+coords, so multiple doors all show
			return
		}
		seen[full] = true
		objs = append(objs, obj{label: full, dist: absInt(x-pos.X) + absInt(y-pos.Y)})
	}
	// NPCs (with coords, for talk_to / use targeting).
	for _, n := range h.world.Npcs.All() {
		if absInt(n.X-pos.X)+absInt(n.Y-pos.Y) > radius {
			continue
		}
		add(reachLabel(d.npcName(h, n.TypeID), n.X, n.Y), n.X, n.Y)
	}
	// Players (with coords + bearing) — so she knows which way a real player is
	// and can answer/follow "I'm to your east" correctly instead of guessing.
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-pos.X)+absInt(p.Y-pos.Y) > radius {
			continue
		}
		add(reachLabel("player "+p.Name, p.X, p.Y), p.X, p.Y)
	}
	// Static scenery + boundaries from facts (UNFILTERED — includes the range).
	if h.facts != nil {
		for _, p := range h.facts.Near(pos.X, pos.Y, radius) {
			switch p.Kind {
			case "scenery":
				add(reachLabel(p.Name, p.X, p.Y), p.X, p.Y)
			case "boundary":
				label := p.Name
				if d.doorUsed(p.X, p.Y) {
					label += " (you've been through this one before)"
				}
				add(label, p.X, p.Y)
			}
		}
	}
	// Dynamic scenery (GameObjects: fires, etc.), names resolved via facts.
	if h.world.Scenery != nil {
		for _, s := range h.world.Scenery.Near(pos.X, pos.Y, radius) {
			name := ""
			if h.facts != nil {
				if def := h.facts.SceneryDef(s.ID); def != nil {
					name = def.Name
				}
			}
			add(name, s.X, s.Y)
		}
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].dist < objs[j].dist })
	if len(objs) > 28 {
		objs = objs[:28]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "You are at (%d,%d). Nearby (name @ x,y — walk_to them, or use items on scenery via use(item, x=, y=)):\n", pos.X, pos.Y)
	for _, o := range objs {
		fmt.Fprintf(&b, "- %s\n", o.label)
	}
	return b.String()
}

// bearingFrom returns a compass direction from (px,py) to (x,y) in RSC
// coordinates (north = smaller y, EAST = smaller x — the x-axis increases west).
// So a player saying "the north door" maps to the door with the smaller y.
func bearingFrom(px, py, x, y int) string {
	dx := x - px // east is negative dx
	dy := y - py // north is negative dy
	if dx == 0 && dy == 0 {
		return "right here"
	}
	var ns, ew string
	if dy < 0 {
		ns = "north"
	} else if dy > 0 {
		ns = "south"
	}
	if dx < 0 {
		ew = "east"
	} else if dx > 0 {
		ew = "west"
	}
	switch {
	case ns != "" && ew != "" && absInt(dy) > 2*absInt(dx):
		return ns
	case ns != "" && ew != "" && absInt(dx) > 2*absInt(dy):
		return ew
	case ns != "" && ew != "":
		return ns + ew
	case ns != "":
		return ns
	default:
		return ew
	}
}

// doorUsed reports whether she has already stood on or beside a door tile (so
// it leads back the way she came). Used to flag doors as do-not-reuse.
func (d *MesaDirector) doorUsed(x, y int) bool {
	if d.visited == nil {
		return false
	}
	for _, off := range [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		if d.visited.has(x+off[0], y+off[1]) {
			return true
		}
	}
	return false
}

// visitedCap bounds the recency window of tiles the director remembers standing
// on. doorUsed only checks the 5 tiles around a door, so a small window is ample;
// the cap stops the lifetime tile-history from growing without bound (M14).
const visitedCap = 64

// tileRing is a fixed-capacity, insertion-ordered set of tiles: add appends
// (evicting the oldest once full) and has tests membership. Deterministic and
// O(1)-amortized, mirroring the host-LIGHT bound the journal/goalgraph use.
type tileRing struct {
	cap   int
	order [][2]int        // ring of the last cap distinct tiles, oldest first
	set   map[[2]int]bool // membership mirror of order
}

func newTileRing(capacity int) *tileRing {
	if capacity <= 0 {
		capacity = visitedCap
	}
	return &tileRing{cap: capacity, set: make(map[[2]int]bool, capacity)}
}

func (r *tileRing) add(x, y int) {
	k := [2]int{x, y}
	if r.set[k] {
		return // already recent — keep its position, don't re-age the window
	}
	if len(r.order) >= r.cap {
		oldest := r.order[0]
		r.order = r.order[1:]
		delete(r.set, oldest)
	}
	r.order = append(r.order, k)
	r.set[k] = true
}

func (r *tileRing) has(x, y int) bool { return r.set[[2]int{x, y}] }
