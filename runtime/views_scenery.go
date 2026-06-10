package runtime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// ---------- dynamic scenery (world.scenery) ----------

// sceneryView is the entry point for dynamic scenery (GameObject)
// queries — the in-view objects the server streams via
// SEND_SCENERY_HANDLER (opcode 48), which is the ONLY place
// runtime-spawned scenery surfaces. A fire lit by firemaking (def 97)
// is registered server-side at runtime and never appears in the
// static facts.SceneryLocs map — so world.locs.search("fire") finds
// only PRE-PLACED fires, never one you just lit. world.scenery is the
// live mirror; use it to find a fire/range you just created:
//
//	world.scenery                       — all visible dynamic scenery
//	world.scenery.by_id(97)             — nearest visible scenery def 97 (a fire)
//	world.scenery.by_id(97, radius=2)   — same, within 2 tiles
//	world.scenery.nearest               — closest scenery to self (bare field)
//
// The selectors (.nearest / .by_id) are REACHABLE-ONLY by default
// (views_reachable.go); by_id takes reachable=false to opt out, and
// the raw .all list is never filtered. There is NO .nearest(pos)
// called form — a callable wrapper here would break use()'s concrete
// *placementView dispatch (the pick_up nearest-wrapper bug class,
// soak retro 2026-06-10 #3a); recenter with .all.nearest(pos) instead.
//
// Results are placement views (kind="scenery") so they drop straight
// into use(item, scenery) → USE_ITEM_ON_SCENERY (opcode 115), e.g.
// use(raw_beef, world.scenery.by_id(97)).
type sceneryView struct{ host *Host }

func (s *sceneryView) Kind() string    { return "scenery" }
func (s *sceneryView) Display() string { return "<scenery>" }

func (s *sceneryView) Iter() []interp.Value {
	records := s.host.world.Scenery.All()
	out := make([]interp.Value, 0, len(records))
	for _, r := range records {
		out = append(out, s.placement(r))
	}
	return out
}

// placement builds a placementView from a dynamic scenery record,
// resolving the def name from facts so .name reads correctly and the
// kind="scenery" routes use() to opcode 115.
func (s *sceneryView) placement(r world.SceneryRecord) *placementView {
	name := "scenery"
	if s.host.facts != nil {
		if def := s.host.facts.SceneryDef(r.ID); def != nil {
			name = def.Name
		}
	}
	return &placementView{p: facts.Placement{
		Kind:  "scenery",
		DefID: r.ID,
		Name:  name,
		X:     r.X,
		Y:     r.Y,
	}}
}

func (s *sceneryView) Get(field string) (interp.Value, bool) {
	switch field {
	case "by_id":
		return &sceneryByIDCallable{host: s.host}, true
	case "nearest":
		// Reachable-only by default (views_reachable.go): a fire
		// across a wall is not a usable fire. Bare field — the raw
		// .all list (or by_id(..., reachable=false)) is the
		// omniscient escape.
		records := reachableSceneryRecords(s.host.world.Scenery.All(), s.host.reachGate())
		if len(records) == 0 {
			return interp.Null{}, true
		}
		return s.nearest(records, nil), true
	case "all":
		return &interp.List{Items: s.Iter()}, true
	case "length":
		return interp.Int(int64(len(s.host.world.Scenery.All()))), true
	}
	return nil, false
}

// reachableSceneryRecords applies a selector reach gate to dynamic
// scenery records. gate == nil keeps everything (no oracle / opt-out).
func reachableSceneryRecords(records []world.SceneryRecord, gate func(x, y int) bool) []world.SceneryRecord {
	if gate == nil {
		return records
	}
	out := make([]world.SceneryRecord, 0, len(records))
	for _, r := range records {
		if gate(r.X, r.Y) {
			out = append(out, r)
		}
	}
	return out
}

func (s *sceneryView) Index(idx interp.Value) (interp.Value, bool) {
	i, ok := interp.AsInt(idx)
	if !ok {
		return nil, false
	}
	items := s.Iter()
	if int(i) < 0 || int(i) >= len(items) {
		return nil, false
	}
	return items[i], true
}

// nearest returns the scenery record closest to `center` (or
// self.position when center is nil), as a placementView. Returns Null
// when records is empty.
func (s *sceneryView) nearest(records []world.SceneryRecord, center *positionView) interp.Value {
	if len(records) == 0 {
		return interp.Null{}
	}
	cx, cy := 0, 0
	if center != nil {
		cx, cy = center.X, center.Y
	} else {
		pos := s.host.world.Self.Position()
		cx, cy = pos.X, pos.Y
	}
	bestIdx := 0
	bestDist := chebyshev(cx, cy, records[0].X, records[0].Y)
	for i := 1; i < len(records); i++ {
		d := chebyshev(cx, cy, records[i].X, records[i].Y)
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return s.placement(records[bestIdx])
}

// sceneryByIDCallable backs `world.scenery.by_id(N, radius=R?)`.
type sceneryByIDCallable struct{ host *Host }

func (c *sceneryByIDCallable) Kind() string    { return "callable" }
func (c *sceneryByIDCallable) Display() string { return "<world.scenery.by_id>" }
func (c *sceneryByIDCallable) Yields() bool    { return false }

func (c *sceneryByIDCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 {
		return nil, errf("world.scenery.by_id needs a scenery def id")
	}
	id, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("world.scenery.by_id: expected Int def_id")
	}
	radius := -1
	if v, ok := named["radius"]; ok {
		if r, ok := interp.AsInt(v); ok {
			radius = int(r)
		}
	}
	pos := c.host.world.Self.Position()
	// Reachable-only by default; reachable=false opts out (views_reachable.go).
	gate := c.host.selectorGate(named)
	all := c.host.world.Scenery.All()
	var matches []world.SceneryRecord
	for _, r := range all {
		if r.ID != int(id) {
			continue
		}
		if radius >= 0 && chebyshev(pos.X, pos.Y, r.X, r.Y) > radius {
			continue
		}
		if gate != nil && !gate(r.X, r.Y) {
			continue
		}
		matches = append(matches, r)
	}
	if len(matches) == 0 {
		return interp.Null{}, nil
	}
	sv := &sceneryView{host: c.host}
	return sv.nearest(matches, nil), nil
}

// ---------- boundaries ----------

// boundariesView is the entry point for boundary queries:
// `world.boundaries.at(x=103, y=532, dir=0)` returns a boundary
// view that can be passed to use(), open_boundary(), etc.
//
// Today it only supports `.at(x, y, dir)` for direct lookup of
// boundary loc data via facts.Facts. Future: `.near(pos, radius)`
// returning a list of nearby openable boundaries, `.find(predicate)`
// with lambda filtering.
type boundariesView struct{ host *Host }

func (b *boundariesView) Kind() string    { return "boundaries" }
func (b *boundariesView) Display() string { return "<boundaries>" }

func (b *boundariesView) Get(field string) (interp.Value, bool) {
	switch field {
	case "at":
		return &boundaryAtCallable{host: b.host}, true
	case "near":
		// world.boundaries.near(radius) → nearest-first list of openable
		// boundary views (doors/gates) from facts, each with its real
		// direction, ready to pass to open_boundary(). Default radius 8.
		return &boundaryNearCallable{host: b.host}, true
	case "is_open":
		// world.boundaries.is_open(x, y, dir) — true iff we've seen
		// a SEND_BOUNDARY_HANDLER mark this tile/dir as removed
		// (door opened, web cut). Returns false for unknown tiles
		// (caller falls back to walk_to or pathfinder rejection).
		return &boundaryIsOpenCallable{host: b.host}, true
	case "dynamic":
		// Snapshot of all dynamic boundary overrides as a list of
		// {x, y, dir, id}. Useful for debugging / persistence.
		all := b.host.world.Boundaries.All()
		items := make([]interp.Value, 0, len(all))
		for k, id := range all {
			items = append(items, &interp.List{Items: []interp.Value{
				interp.Int(int64(k.X)),
				interp.Int(int64(k.Y)),
				interp.Int(int64(k.Dir)),
				interp.Int(int64(id)),
			}})
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// boundaryIsOpenCallable backs world.boundaries.is_open(x, y, dir).
type boundaryIsOpenCallable struct{ host *Host }

func (c *boundaryIsOpenCallable) Kind() string    { return "callable" }
func (c *boundaryIsOpenCallable) Display() string { return "<world.boundaries.is_open>" }
func (c *boundaryIsOpenCallable) Yields() bool    { return false }

func (c *boundaryIsOpenCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, dir, err := resolveBoundaryAt(args, named)
	if err != nil {
		return nil, err
	}
	return interp.Bool(c.host.world.Boundaries.IsRemoved(x, y, dir)), nil
}

// boundaryAtCallable backs `world.boundaries.at(x=X, y=Y, dir=D)`.
// Returns a boundaryView, or Null if no boundary def is known at
// that (x, y, dir) — the facts index drives the lookup; if the
// landscape encodes a wall via the .orsc tile bytes but the
// BoundaryLocs JSON doesn't list it, we still synthesize a view
// from the coords so use(key, door) can fire regardless.
type boundaryAtCallable struct{ host *Host }

func (c *boundaryAtCallable) Kind() string    { return "callable" }
func (c *boundaryAtCallable) Display() string { return "<world.boundaries.at>" }
func (c *boundaryAtCallable) Yields() bool    { return false }

func (c *boundaryAtCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, dir, err := resolveBoundaryAt(args, named)
	if err != nil {
		return nil, err
	}
	// Even if facts doesn't know about this exact (x, y, dir),
	// return a synthetic view — callers may know about a door
	// the static facts data missed (and use() works either way
	// since the server validates the click).
	return &boundaryView{x: x, y: y, direction: dir, host: c.host}, nil
}

// boundaryNearCallable backs world.boundaries.near(radius) — a nearest-first
// list of nearby openable boundary views (doors/gates) drawn from facts, each
// carrying its real direction so open_boundary() fires the correct click.
type boundaryNearCallable struct{ host *Host }

func (c *boundaryNearCallable) Kind() string    { return "callable" }
func (c *boundaryNearCallable) Display() string { return "<world.boundaries.near>" }
func (c *boundaryNearCallable) Yields() bool    { return false }

func (c *boundaryNearCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	radius := 8
	if len(args) >= 1 {
		if r, ok := args[0].(interp.Int); ok {
			radius = int(r)
		}
	}
	if c.host.facts == nil || c.host.world == nil || c.host.world.Self == nil {
		return &interp.List{}, nil
	}
	pos := c.host.world.Self.Position()
	// Reachable-only by default; reachable=false opts out
	// (views_reachable.go). Safe for doors: openable boundaries don't
	// cut oracle components, so a door you could walk to and open is
	// in the host's component — only doors in genuinely disconnected
	// space (another building's interior across a fence) drop out.
	gate := c.host.selectorGate(named)
	type bd struct {
		v    *boundaryView
		dist int
	}
	var bds []bd
	for _, p := range c.host.facts.Near(pos.X, pos.Y, radius) {
		if p.Kind != "boundary" {
			continue
		}
		// Only OPENABLE boundaries (doors/gates): Unknown==1 marks them; plain
		// walls/fences are Unknown==0 and must not be returned (opening a wall is
		// a no-op that strands the caller).
		def := c.host.facts.BoundaryDef(p.DefID)
		if def == nil || def.Unknown != 1 {
			continue
		}
		if gate != nil && !gate(p.X, p.Y) {
			continue
		}
		bds = append(bds, bd{
			v:    &boundaryView{x: p.X, y: p.Y, direction: p.Direction, host: c.host},
			dist: absInt(p.X-pos.X) + absInt(p.Y-pos.Y),
		})
	}
	sort.Slice(bds, func(i, j int) bool { return bds[i].dist < bds[j].dist })
	items := make([]interp.Value, 0, len(bds))
	for _, b := range bds {
		items = append(items, b.v)
	}
	return &interp.List{Items: items}, nil
}

// resolveBoundaryAt parses (x, y, dir) from positional or named
// args. Accepts:
//
//	at(x, y, dir)             // positional
//	at(x=X, y=Y, dir=D)       // named
//
// `dir` defaults to 0 if omitted.
func resolveBoundaryAt(args []interp.Value, named map[string]interp.Value) (int, int, int, error) {
	var x, y, dir int
	if len(args) >= 2 {
		if xi, ok := args[0].(interp.Int); ok {
			x = int(xi)
		} else {
			return 0, 0, 0, fmt.Errorf("world.boundaries.at: positional x must be Int")
		}
		if yi, ok := args[1].(interp.Int); ok {
			y = int(yi)
		} else {
			return 0, 0, 0, fmt.Errorf("world.boundaries.at: positional y must be Int")
		}
		if len(args) >= 3 {
			if di, ok := args[2].(interp.Int); ok {
				dir = int(di)
			}
		}
	}
	if v, ok := named["x"]; ok {
		if i, ok := v.(interp.Int); ok {
			x = int(i)
		}
	}
	if v, ok := named["y"]; ok {
		if i, ok := v.(interp.Int); ok {
			y = int(i)
		}
	}
	if v, ok := named["dir"]; ok {
		if i, ok := v.(interp.Int); ok {
			dir = int(i)
		}
	}
	return x, y, dir, nil
}

// boundaryView is a DSL-visible reference to a boundary tile +
// direction. Carries enough info for use() / open_boundary() to
// dispatch the right packet. Construct via `world.boundaries.at()`.
type boundaryView struct {
	x, y, direction int
	host            *Host
}

func (b *boundaryView) Kind() string { return "boundary" }
func (b *boundaryView) Display() string {
	return fmt.Sprintf("<boundary (%d, %d) dir=%d>", b.x, b.y, b.direction)
}

func (b *boundaryView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(int64(b.x)), true
	case "y":
		return interp.Int(int64(b.y)), true
	case "direction", "dir":
		return interp.Int(int64(b.direction)), true
	case "position":
		return &positionView{X: b.x, Y: b.y}, true
	case "name":
		// If facts knows a def at this tile, surface its name.
		if b.host != nil && b.host.facts != nil {
			for _, p := range b.host.facts.At(b.x, b.y) {
				if p.Kind == "boundary" && p.Direction == b.direction {
					return interp.String(p.Name), true
				}
			}
		}
		return interp.String("door"), true
	case "door_type":
		if d := b.def(); d != nil {
			return interp.Int(int64(d.DoorType)), true
		}
		return interp.Int(0), true
	case "is_openable":
		// True for doors/doorframes the bot can open. When the def is
		// unknown, assume openable (conservative for a thing named "door").
		if d := b.def(); d != nil {
			return interp.Bool(d.IsOpenable()), true
		}
		return interp.Bool(true), true
	case "blocks_when_closed":
		// True for closeable doors (DoorType==1) that block while closed —
		// pathfind to them and open before crossing. Unknown def -> false.
		if d := b.def(); d != nil {
			return interp.Bool(d.BlocksWhenClosed()), true
		}
		return interp.Bool(false), true
	}
	return nil, false
}

// def resolves the BoundaryDef known at this boundary's tile + direction,
// or nil if facts has none. Powers the name / door-state field accessors.
func (b *boundaryView) def() *facts.BoundaryDef {
	if b.host == nil || b.host.facts == nil {
		return nil
	}
	for _, p := range b.host.facts.At(b.x, b.y) {
		if p.Kind == "boundary" && p.Direction == b.direction {
			return b.host.facts.BoundaryDef(p.DefID)
		}
	}
	return nil
}

// ---------- locs (facts-backed location queries) ----------

// locsView exposes `world.locs.banks`, `world.locs.fishing_spots`,
// etc. as searchable lists.
//
// Implementation strategy: the DSL field name (e.g. "banks") maps to
// a set of facts-known categories. Each returns a *locListView that
// supports `.nearest(point)` and iteration.
type locsView struct{ host *Host }

func (l *locsView) Kind() string    { return "locs" }
func (l *locsView) Display() string { return "<locs>" }

func (l *locsView) Get(field string) (interp.Value, bool) {
	// Some categories are well-known names. Others are queried by
	// substring match in scenery/boundary/NPC names. We start with
	// the common ones; new categories are easy to add.
	switch field {
	case "banks":
		return l.searchByName("bank"), true
	case "altars":
		return l.searchByName("altar"), true
	case "furnaces":
		return l.searchByName("furnace"), true
	case "anvils":
		return l.searchByName("anvil"), true
	case "fishing_spots":
		return l.searchByName("fishing"), true
	case "trees":
		return l.searchByName("tree"), true
	case "rocks":
		return l.searchByName("rock"), true
	case "doors":
		return l.searchBoundaryByName("door"), true
	case "ladders":
		// Ladders are SCENERY (GameObjectDef "Ladder", e.g.
		// LADDER_GENERIC_UP=5 / _DOWN=6), not boundaries — they are
		// climbed via the scenery Command1 "Climb-up"/"Climb-down"
		// (interact_at / ObjectCommand opcode 136), not opened like a
		// door. searchBoundaryByName never matched anything (DoorDef
		// has no ladders), so this returned an empty set everywhere.
		return l.searchByName("ladder"), true
	case "shops":
		return l.searchByName("shop"), true
	case "scenery":
		// Every scenery def — useful with .within(radius) to find
		// all nearby objects regardless of kind.
		return l.allOfKind("scenery"), true
	case "spawn_points":
		// Every NPC spawn point. Distinct from world.npcs (which is
		// currently-visible NPCs); spawn_points is "where NPCs
		// originate" — useful for "walk to the goblin spawn area."
		return l.allOfKind("npc_spawn"), true
	case "search":
		// Generic substring search for scenery / boundary / spawn
		// defs by name. Returns a locListView (chain .nearest /
		// .within / .names like the named categories).
		return locsSearchCallable{l: l}, true
	}
	return nil, false
}

// locsSearchCallable: `world.locs.search("furnace").nearest(self.position)`.
type locsSearchCallable struct{ l *locsView }

func (c locsSearchCallable) Kind() string    { return "builtin" }
func (c locsSearchCallable) Display() string { return "<locs.search>" }
func (c locsSearchCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("locs.search takes 1 arg (needle string)")
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("locs.search: needle must be String, got %s", args[0].Kind())
	}
	return c.l.searchByName(string(needle)), nil
}

// allOfKind populates a locListView with every def of the given
// facts kind ("scenery" / "npc_spawn" / "boundary"). Used by
// world.locs.scenery and world.locs.spawn_points — categories
// that don't filter by name but by kind alone.
func (l *locsView) allOfKind(kind string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	switch kind {
	case "scenery":
		for _, def := range l.host.facts.SceneryDefs {
			if def != nil {
				out.kinds = append(out.kinds, "scenery")
				out.names = append(out.names, def.Name)
			}
		}
	case "npc_spawn":
		for _, def := range l.host.facts.NpcDefs {
			if def != nil {
				out.kinds = append(out.kinds, "npc_spawn")
				out.names = append(out.names, def.Name)
			}
		}
	case "boundary":
		for _, def := range l.host.facts.BoundaryDefs {
			if def != nil {
				out.kinds = append(out.kinds, "boundary")
				out.names = append(out.names, def.Name)
			}
		}
	}
	return out
}

// searchByName collects scenery placements whose name contains
// the given substring (case-insensitive). If no scenery matches,
// falls back to NPC defs (e.g. world.locs.banks will hit the
// "Banker" NPC if no bank-booth scenery is defined). This
// preference matters because routines almost always want the
// physical destination (the bank counter), not the NPC standing
// next to it.
//
// Discovered live 2026-05-28: a substring match on "bank" hit
// both the Bank booth scenery AND the Banker NPC; routines that
// then called walk_to ended up at the NPC's spawn instead of
// the counter. Preferring scenery resolves that.
func (l *locsView) searchByName(needle string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	needle = strings.ToLower(needle)
	for i := range l.host.facts.SceneryDefs {
		def := l.host.facts.SceneryDefs[i]
		if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
			out.kinds = append(out.kinds, "scenery")
			out.names = append(out.names, def.Name)
		}
	}
	// Fall back to NPC defs only when scenery yielded nothing.
	if len(out.names) == 0 {
		for i := range l.host.facts.NpcDefs {
			def := l.host.facts.NpcDefs[i]
			if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
				out.kinds = append(out.kinds, "npc_spawn")
				out.names = append(out.names, def.Name)
			}
		}
	}
	return out
}

func (l *locsView) searchBoundaryByName(needle string) *locListView {
	out := &locListView{host: l.host}
	if l.host.facts == nil {
		return out
	}
	needle = strings.ToLower(needle)
	for i := range l.host.facts.BoundaryDefs {
		def := l.host.facts.BoundaryDefs[i]
		if def != nil && strings.Contains(strings.ToLower(def.Name), needle) {
			out.kinds = append(out.kinds, "boundary")
			out.names = append(out.names, def.Name)
		}
	}
	return out
}

// locListView wraps a name+kind list and resolves to actual
// placements lazily via facts.NearestByName when `.nearest(point)`
// is called. We don't materialize every placement up front because
// some categories (e.g. trees) have thousands of instances.
type locListView struct {
	host  *Host
	kinds []string
	names []string
}

func (v *locListView) Kind() string    { return "loc_list" }
func (v *locListView) Display() string { return "<loc_list:" + intDisp(len(v.names)) + ">" }

func (v *locListView) Get(field string) (interp.Value, bool) {
	switch field {
	case "length":
		// Approximate — we don't know the actual placement count
		// without scanning, so return the number of *kinds* that
		// match. Routines should prefer .nearest() over .length.
		return interp.Int(int64(len(v.names))), true
	case "nearest":
		return locNearestCallable{owner: v}, true
	case "within":
		return locWithinCallable{owner: v}, true
	case "names":
		items := make([]interp.Value, 0, len(v.names))
		for _, n := range v.names {
			items = append(items, interp.String(n))
		}
		return &interp.List{Items: items}, true
	}
	return nil, false
}

// locNearestCallable: `locs.banks.nearest(self.position)` -> placement
type locNearestCallable struct{ owner *locListView }

func (c locNearestCallable) Kind() string    { return "builtin" }
func (c locNearestCallable) Display() string { return "<locs.nearest>" }
func (c locNearestCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, err
	}
	if c.owner.host.facts == nil {
		return interp.Null{}, nil
	}
	const radius = 64
	var best *facts.Placement
	bestD := -1
	for i, kind := range c.owner.kinds {
		p := c.owner.host.facts.NearestByName(c.owner.names[i], x, y, radius, kind)
		if p == nil {
			continue
		}
		d := chebyshev(p.X, p.Y, x, y)
		if best == nil || d < bestD {
			best = p
			bestD = d
		}
	}
	if best == nil {
		return interp.Null{}, nil
	}
	return &placementView{p: *best}, nil
}

// locWithinCallable: `world.locs.banks.within(radius)` →
// list<placement> of placements of the matching def(s) within
// the Chebyshev radius of self.position. Useful for "find every
// fishing spot within 20 tiles."
//
// Signature variants accepted:
//
//	.within(radius)              — center = self.position
//	.within(radius, position)    — center = supplied position
//
// Returns an empty list when facts aren't loaded.
type locWithinCallable struct{ owner *locListView }

func (c locWithinCallable) Kind() string    { return "builtin" }
func (c locWithinCallable) Display() string { return "<locs.within>" }
func (c locWithinCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) == 0 {
		return nil, errf("locs.within requires at least a radius")
	}
	r, ok := interp.AsInt(args[0])
	if !ok {
		return nil, errf("locs.within: first arg must be int radius, got %s", args[0].Kind())
	}
	radius := int(r)
	if radius < 0 {
		return nil, errf("locs.within: radius must be >= 0, got %d", radius)
	}
	// Center defaults to host's current position.
	cx, cy := 0, 0
	if c.owner.host != nil {
		p := c.owner.host.world.Self.Position()
		cx, cy = p.X, p.Y
	}
	// Optional position override.
	if len(args) >= 2 {
		if g, ok := args[1].(interp.Getter); ok {
			if xv, ok := g.Get("x"); ok {
				if x, ok := interp.AsInt(xv); ok {
					cx = int(x)
				}
			}
			if yv, ok := g.Get("y"); ok {
				if y, ok := interp.AsInt(yv); ok {
					cy = int(y)
				}
			}
		}
	}
	if c.owner.host == nil || c.owner.host.facts == nil {
		return &interp.List{}, nil
	}
	// Build set of (kind, name) we're matching, then scan
	// facts.Near for matches.
	type matcher struct{ kind, name string }
	wants := make(map[matcher]struct{}, len(c.owner.names))
	for i, name := range c.owner.names {
		wants[matcher{c.owner.kinds[i], name}] = struct{}{}
	}
	hits := c.owner.host.facts.Near(cx, cy, radius)
	out := make([]interp.Value, 0, len(hits))
	for _, p := range hits {
		if _, want := wants[matcher{p.Kind, p.Name}]; want {
			placement := p
			out = append(out, &placementView{p: placement})
		}
	}
	return &interp.List{Items: out}, nil
}

// placementView wraps a facts.Placement so DSL code can read
// `.x`, `.y`, `.name`, etc.
type placementView struct{ p facts.Placement }

func (v *placementView) Kind() string    { return "placement" }
func (v *placementView) Display() string { return v.p.Name }
func (v *placementView) Get(field string) (interp.Value, bool) {
	switch field {
	case "x":
		return interp.Int(int64(v.p.X)), true
	case "y":
		return interp.Int(int64(v.p.Y)), true
	case "name":
		return interp.String(v.p.Name), true
	case "kind":
		return interp.String(v.p.Kind), true
	case "def_id":
		return interp.Int(int64(v.p.DefID)), true
	case "direction":
		return interp.Int(int64(v.p.Direction)), true
	case "position":
		return &positionView{X: v.p.X, Y: v.p.Y}, true
	}
	return nil, false
}
