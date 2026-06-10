package runtime

import (
	"context"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
)

// SCENE-PERCEPTION verbs (local, cheap). Unlike the WorldOracle map-perception
// verbs in actions_worldmap.go (search_map / reachable / survey_map), which
// reason about cross-region reachability and cost in-world "study" seconds,
// scan_for is a plain GLANCE at the scenery immediately around the host. It is
// the iterate-and-prune counterpart to the manual's old "act on scenery BY
// COORDINATE" guidance: instead of the planner copying a literal rock tile out
// of the text scene summary (and hardcoding interact_at(x=ROCKX, y=ROCKY)), the
// host enumerates the real nearby scenery of a type and loops over it.
//
// scan_for reads BOTH scenery sources and merges them tile-by-tile:
//   - STATIC: pre-placed scenery from facts (facts.Near) — every rock / tree /
//     fishing spot / range the landscape carries. Frozen at load.
//   - LIVE: runtime-spawned scenery from the world mirror (world.Scenery) —
//     lit fires, regrown trees, etc. This source reflects depletion/regrowth,
//     so a LIVE record at a tile supersedes the static baseline there, and a
//     tile the server explicitly cleared (IsRemoved) drops out of the result.
//
// Registration lives in the central actionHandlers table (dsl_actions.go); the
// spec row lives in dsl/spec/actions.go.

const (
	// scanForDefaultRadius is the Chebyshev scan window when the caller omits
	// radius — the same local-glance default as look_around.
	scanForDefaultRadius = 10
	// scanForMaxResults caps how many scenery hits scan_for returns so a dense
	// scene cannot flood the planner's context. The NEAREST are kept (the list
	// is sorted by distance before the cap). Tunable.
	scanForMaxResults = 24
)

// dslScanFor enumerates nearby scenery of a TYPE ("rock", "tree", "fishing
// spot", "range", "fire", …) and returns a distance-sorted, prunable list the
// host iterates instead of hardcoding tiles. Each entry is a *placementView, so
// it is FIELD-ACCESSIBLE in the DSL (hit.x / hit.y / hit.name / hit.position)
// and drops straight into interact_at(x=, y=) / go_to / use — the contract that
// makes iterate-and-prune actually executable (unlike search_map's *interp.Map
// entries, which the interpreter cannot field-access).
//
// Returns a BARE *interp.List (not wrapped in a CallResult), so a routine can
// write `for r in scan_for("rock") { ... }`, `scan_for("rock").first`, or
// `scan_for("rock")[0]` directly. When nothing matches it returns an EMPTY list
// — branch on `.length == 0`, it is never a typed failure.
//
//	scan_for("rock")          -> [{x,y,name,kind,def_id}, ...] nearest-first
//	scan_for("tree", 6)       -> same, within a 6-tile radius
//	scan_for("rock", radius=6)
func dslScanFor(_ context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 {
		return nil, errf("scan_for takes a scenery type (e.g. scan_for(\"rock\")), got 0 args")
	}
	typ := strings.TrimSpace(args[0].Display())
	if typ == "" {
		return nil, errf("scan_for: scenery type must be a non-empty string (e.g. \"rock\", \"tree\")")
	}
	low := strings.ToLower(typ)

	radius := scanForDefaultRadius
	if len(args) >= 2 {
		if r, ok := interp.AsInt(args[1]); ok {
			radius = int(r)
		}
	}
	if v, ok := named["radius"]; ok {
		if r, ok := interp.AsInt(v); ok {
			radius = int(r)
		}
	}
	if radius < 0 {
		radius = 0
	}

	pos := h.world.Self.Position()

	// Reachability gate (#31 — "absolute negative space"): a human looking at the
	// scene sees that scenery behind a wall, across a gap, or up a ladder on a
	// dead-end floor is UNREACHABLE. Without this, scan_for hands the planner targets
	// it can never path-nav to (the live trap: boxed in an empty upper room, scanning
	// + "finding" a bed/rock on the far side of a wall, looping). Drop any scenery in
	// a different walkable COMPONENT than the host — the worldOracle's components are
	// the walkable-connectivity partition; CompNear snaps a non-standable scenery tile
	// (a rock/bed footprint) to its nearest standable component. Skipped when there is
	// no oracle (don't break scan_for) or the host's own tile won't resolve.
	hostComp, reachGate := int32(-1), false
	if h.worldOracle != nil {
		if c, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y); ok {
			hostComp, reachGate = c, true
		}
	}
	reachable := func(x, y int) bool {
		if !reachGate {
			return true // no oracle / host tile unresolved — do not filter
		}
		c, _, _, ok := h.worldOracle.CompNear(x, y)
		return ok && c == hostComp
	}

	// Merge static + live scenery keyed by tile so a LIVE record (the
	// depletion/regrowth truth) supersedes the frozen static baseline at the
	// same tile.
	byTile := map[[2]int]facts.Placement{}

	// STATIC: pre-placed scenery from facts. Skip tiles the live mirror knows
	// were cleared (a mined-out rock / burned-out fire) so we don't report a
	// ghost the host would walk to and find gone.
	if h.facts != nil {
		for _, p := range h.facts.Near(pos.X, pos.Y, radius) {
			if p.Kind != "scenery" {
				continue
			}
			if !strings.Contains(strings.ToLower(p.Name), low) {
				continue
			}
			if h.world.Scenery != nil && h.world.Scenery.IsRemoved(p.X, p.Y) {
				continue
			}
			byTile[[2]int{p.X, p.Y}] = p
		}
	}

	// LIVE: runtime-spawned scenery from the world mirror. Resolve the name via
	// facts (mirrors sceneryView.placement) and let it win the tile.
	if h.world.Scenery != nil {
		for _, r := range h.world.Scenery.Near(pos.X, pos.Y, radius) {
			name := "scenery"
			if h.facts != nil {
				if def := h.facts.SceneryDef(r.ID); def != nil && def.Name != "" {
					name = def.Name
				}
			}
			if !strings.Contains(strings.ToLower(name), low) {
				continue
			}
			byTile[[2]int{r.X, r.Y}] = facts.Placement{
				Kind:  "scenery",
				DefID: r.ID,
				Name:  name,
				X:     r.X,
				Y:     r.Y,
			}
		}
	}

	type cand struct {
		p    facts.Placement
		dist int
	}
	cands := make([]cand, 0, len(byTile))
	for _, p := range byTile {
		if !reachable(p.X, p.Y) {
			continue // negative space: visible but no walkable path from here (#31)
		}
		cands = append(cands, cand{p: p, dist: chebyshev(pos.X, pos.Y, p.X, p.Y)})
	}
	// Sort nearest-first. Tile-coordinate tiebreaks keep the order deterministic
	// despite Go's randomized map iteration (the tests rely on this).
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].dist != cands[j].dist {
			return cands[i].dist < cands[j].dist
		}
		if cands[i].p.X != cands[j].p.X {
			return cands[i].p.X < cands[j].p.X
		}
		return cands[i].p.Y < cands[j].p.Y
	})
	if len(cands) > scanForMaxResults {
		cands = cands[:scanForMaxResults]
	}

	items := make([]interp.Value, 0, len(cands))
	for _, c := range cands {
		placement := c.p
		items = append(items, &placementView{p: placement})
	}
	return &interp.List{Items: items}, nil
}
