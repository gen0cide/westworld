package runtime

import "github.com/gen0cide/westworld/dsl/interp"

// Perception-side reachability gate for the action-intent SELECTORS
// (nearest_npc, world.ground_items.nearest / by_id / most_valuable,
// world.scenery.nearest / by_id, world.boundaries.near, list.nearest,
// scan_for). A selector's job is to hand the planner a target it can
// actually act on — returning scenery/NPCs/loot in "absolute negative
// space" (behind a wall, across a gap, on a dead-end floor) is the
// proven stuck-loop: act on it, fail to path, re-select the same
// target, loop (#31).
//
// The gate is the SAME component-equality test scan_for shipped in
// a270236: the worldOracle's flood-fill components are the precomputed
// walkable-connectivity partition, so the test is two label-array
// reads per candidate — never a per-call BFS. CompNear snaps a
// non-standable footprint tile (a rock/bed/booth) to its nearest
// standable component. Openable doors/gates do NOT cut components
// (worldmap.applyBoundaries honors BlocksMovement only), so "same
// component" means "walkable, opening doors en route" — a door you
// could open into another room is still reachable, while a sealed
// room or the far side of a fence is not.
//
// Every selector defaults to reachable-only and accepts the named arg
// reachable=false to opt out (omniscient debugging); bare-field
// selectors without call syntax (.most_valuable, world.scenery.nearest)
// keep the raw `.all` list as their escape hatch.

// reachGate returns the reachable-only predicate for one selector
// call, with the host's own component resolved once. Returns nil when
// filtering is impossible — no world oracle, or the host's own tile
// won't resolve to a component — and callers MUST treat nil as "do
// not filter" (the scan_for precedent: never break a selector for a
// host without map data).
func (h *Host) reachGate() func(x, y int) bool {
	if h.worldOracle == nil || h.world == nil || h.world.Self == nil {
		return nil
	}
	pos := h.world.Self.Position()
	hostComp, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y)
	if !ok {
		return nil
	}
	return func(x, y int) bool {
		c, _, _, ok := h.worldOracle.CompNear(x, y)
		return ok && c == hostComp
	}
}

// selectorGate resolves the gate one selector call should apply:
// reachGate() unless the caller passed reachable=false. Nil = keep
// every candidate.
func (h *Host) selectorGate(named map[string]interp.Value) func(x, y int) bool {
	if !wantReachable(named) {
		return nil
	}
	return h.reachGate()
}

// wantReachable parses the shared reachable= named opt-out: filtering
// applies unless the caller explicitly passed a falsey value
// (reachable=false).
func wantReachable(named map[string]interp.Value) bool {
	if v, ok := named["reachable"]; ok {
		return interp.Truthy(v)
	}
	return true
}

// tileReachable adapts reachGate to the per-tile hook shape
// interp.Interpreter.Reachable wants (list.nearest). It re-resolves
// the host's component on every call because the host moves between
// list.nearest invocations within one routine.
func (h *Host) tileReachable(x, y int) bool {
	gate := h.reachGate()
	return gate == nil || gate(x, y)
}
