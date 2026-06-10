package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// Spatial utility action handler bodies (where_am_i / bearing_to /
// where_is / distance_to / distance_to_xy / nearest_npc / in_region) and
// the bounds shape constructors (box / circle / near). Registered in the
// central actionHandlers table in dsl_actions.go.

// dslWhereAmI returns a readable summary of where the host is in the
// world — nearest named area + notable POIs with bearing/distance — so a
// host can reason about its place on the map, not raw coords.
func dslWhereAmI(_ context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	return interp.String(h.LocationSummary()), nil
}

// dslBearingTo returns the 8-point compass direction from the host to a
// target tile (x, y, or a position-like value). "here" if coincident.
func dslBearingTo(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("bearing_to: %v", err)
	}
	pos := h.world.Self.Position()
	if b := facts.Bearing(pos.X, pos.Y, x, y); b != "" {
		return interp.String(b), nil
	}
	return interp.String("here"), nil
}

// dslWhereIs locates a named place ("Lumbridge") or a POI type ("bank",
// "altar", "furnace") relative to the host: distance + bearing + coords.
func dslWhereIs(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("where_is takes 1 arg (place name or POI type), got %d", len(args))
	}
	if h.facts == nil {
		return interp.Null{}, nil
	}
	name := args[0].Display()
	pos := h.world.Self.Position()
	g := h.facts.Gazetteer()
	if p, ok := g.PlaceByName(name, pos.X, pos.Y); ok {
		return interp.String(fmt.Sprintf("%s is %d tiles %s, at (%d, %d)",
			p.Name, chebyshev(pos.X, pos.Y, p.X, p.Y), facts.Bearing(pos.X, pos.Y, p.X, p.Y), p.X, p.Y)), nil
	}
	if p, d, ok := g.NearestPOI(name, pos.X, pos.Y); ok {
		return interp.String(fmt.Sprintf("nearest %s is %d tiles %s, at (%d, %d)",
			name, d, facts.Bearing(pos.X, pos.Y, p.X, p.Y), p.X, p.Y)), nil
	}
	return interp.String("unknown place: " + name), nil
}

// ---------- spatial utilities (pure, no opcodes) ----------

// dslDistanceTo returns the Chebyshev distance from self.position
// to the target. Chebyshev (max(|dx|, |dy|)) matches RSC's walk
// cost — one diagonal step = one tile.
//
// Accepts any view with .x/.y (positionView, playerView, npcView,
// groundItemView, placementView, boundaryView), or named (x=X, y=Y).
func dslDistanceTo(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 && (len(named) == 0) {
		return nil, errf("distance_to takes 1 target argument (view or {x,y}), got %d", len(args))
	}
	tx, ty, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("distance_to: %v", err)
	}
	pos := h.world.Self.Position()
	dx := pos.X - tx
	dy := pos.Y - ty
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	d := dx
	if dy > dx {
		d = dy
	}
	return interp.Int(int64(d)), nil
}

// dslDistanceToXY is a positional convenience over distance_to —
// `distance_to_xy(304, 542)` is shorter than `distance_to(x=304, y=542)`
// when the target is a literal tile rather than a view. Chebyshev
// distance (max(|dx|, |dy|)) like the underlying.
func dslDistanceToXY(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("distance_to_xy takes 2 args (x, y), got %d", len(args))
	}
	tx, ok1 := interp.AsInt(args[0])
	ty, ok2 := interp.AsInt(args[1])
	if !ok1 || !ok2 {
		return nil, errf("distance_to_xy: both args must be Int, got %s/%s", args[0].Kind(), args[1].Kind())
	}
	pos := h.world.Self.Position()
	dx := pos.X - int(tx)
	dy := pos.Y - int(ty)
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	if dy > dx {
		return interp.Int(int64(dy)), nil
	}
	return interp.Int(int64(dx)), nil
}

// dslNearestNpc returns the NPC closest to self.position (Chebyshev
// distance), optionally filtered by a 1-arg predicate lambda. It exists
// because world.npcs.find returns the FIRST roster match, not the
// closest — fatal when several NPCs of the same type are in view and
// some are far away or already engaged by another bot. nearest_npc lets
// a combat routine reliably attack the goblin it just spawned adjacent
// to itself rather than a contested wanderer across the field.
//
//	nearest_npc()                    -> closest visible NPC (any), or Null
//	nearest_npc(n => n.type_id == 4) -> closest visible Goblin, or Null
//
// Ties (equal distance) resolve to the first in roster order, which is
// stable across calls within a tick. Returns Null when no NPC matches.
func dslNearestNpc(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) > 1 {
		return nil, errf("nearest_npc takes 0 or 1 args (optional predicate lambda), got %d", len(args))
	}
	var pred interp.Callable
	if len(args) == 1 {
		p, ok := args[0].(interp.Callable)
		if !ok {
			return nil, errf("nearest_npc: arg must be a lambda (e.g. `n => n.type_id == 4`), got %s", args[0].Kind())
		}
		pred = p
	}
	pos := h.world.Self.Position()
	var best *npcView
	bestDist := int(^uint(0) >> 1) // max int
	for _, rec := range h.world.Npcs.All() {
		nv := &npcView{record: rec, facts: h.facts, host: h}
		if pred != nil {
			v, err := pred.Call([]interp.Value{nv}, nil)
			if err != nil {
				return nil, errf("nearest_npc predicate: %v", err)
			}
			if !interp.Truthy(v) {
				continue
			}
		}
		dx := rec.X - pos.X
		if dx < 0 {
			dx = -dx
		}
		dy := rec.Y - pos.Y
		if dy < 0 {
			dy = -dy
		}
		d := dx
		if dy > dx {
			d = dy
		}
		if d < bestDist {
			bestDist = d
			best = nv
		}
	}
	if best == nil {
		return interp.Null{}, nil
	}
	return best, nil
}

// dslInRegion returns true iff self.position is inside the
// rectangle (x1,y1)..(x2,y2) inclusive. Geometry helper. Arg
// order doesn't matter — we normalize min/max.
func dslInRegion(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 4 {
		return nil, errf("in_region takes 4 args (x1, y1, x2, y2), got %d", len(args))
	}
	x1, x2, y1, y2 := intArg(args[0]), intArg(args[2]), intArg(args[1]), intArg(args[3])
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	pos := h.world.Self.Position()
	in := pos.X >= x1 && pos.X <= x2 && pos.Y >= y1 && pos.Y <= y2
	return interp.Bool(in), nil
}

// ---------- bounds shape constructors: box, circle, near ----------
//
// These are pure constructors — they return an interp.RegionPredicate
// value that the bounds-block registration machinery uses as a
// location filter. No server I/O.

// dslBox builds an axis-aligned rectangle predicate. Positional:
// box(x1, y1, x2, y2). Named: box(x1=..., y1=..., x2=..., y2=...).
// Inclusive on all four edges; argument order doesn't matter
// (x1/x2 and y1/y2 are normalized).
func dslBox(_ context.Context, _ *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var x1, y1, x2, y2 int
	switch len(args) {
	case 0:
		x1 = intArg(named["x1"])
		y1 = intArg(named["y1"])
		x2 = intArg(named["x2"])
		y2 = intArg(named["y2"])
	case 4:
		x1, y1, x2, y2 = intArg(args[0]), intArg(args[1]), intArg(args[2]), intArg(args[3])
	default:
		return nil, errf("box(x1, y1, x2, y2): expected 4 positional or 4 named args, got %d positional", len(args))
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	pred := func(x, y int) bool {
		return x >= x1 && x <= x2 && y >= y1 && y <= y2
	}
	name := fmt.Sprintf("box(%d,%d,%d,%d)", x1, y1, x2, y2)
	return interp.NewRegionPredicate(name, pred), nil
}

// dslCircle builds a Chebyshev-distance disk predicate.
// circle(cx, cy, radius) or circle(cx=..., cy=..., radius=...).
// Uses Chebyshev (max of |dx|, |dy|) since RSC movement is grid-8.
func dslCircle(_ context.Context, _ *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var cx, cy, r int
	switch len(args) {
	case 0:
		cx = intArg(named["cx"])
		cy = intArg(named["cy"])
		r = intArg(named["radius"])
	case 3:
		cx, cy, r = intArg(args[0]), intArg(args[1]), intArg(args[2])
	default:
		return nil, errf("circle(cx, cy, radius): expected 3 positional or named args, got %d positional", len(args))
	}
	if r < 0 {
		r = 0
	}
	pred := func(x, y int) bool {
		dx := x - cx
		if dx < 0 {
			dx = -dx
		}
		dy := y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx > dy {
			return dx <= r
		}
		return dy <= r
	}
	name := fmt.Sprintf("circle(%d,%d,r=%d)", cx, cy, r)
	return interp.NewRegionPredicate(name, pred), nil
}

// dslNear builds a disk predicate centered on self.position at
// routine-start time. near(radius) or near(radius=N). Useful for
// "react to events within N tiles of where I started" without
// hard-coding coords.
func dslNear(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var r int
	if v, ok := named["radius"]; ok {
		r = intArg(v)
	} else if len(args) == 1 {
		r = intArg(args[0])
	} else {
		return nil, errf("near(radius): expected 1 positional or named arg")
	}
	if r < 0 {
		r = 0
	}
	pos := h.world.Self.Position()
	cx, cy := pos.X, pos.Y
	pred := func(x, y int) bool {
		dx := x - cx
		if dx < 0 {
			dx = -dx
		}
		dy := y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx > dy {
			return dx <= r
		}
		return dy <= r
	}
	name := fmt.Sprintf("near(%d,%d,r=%d)", cx, cy, r)
	return interp.NewRegionPredicate(name, pred), nil
}
