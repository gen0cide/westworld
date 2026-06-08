package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/pathfind"
)

// ErrNoPath means the pathfinder couldn't find a route from the
// player's current location to the requested target inside the local
// 96x96 region. Either the goal is outside the loaded region or
// there's no reachable path through the landscape walls / scenery.
var ErrNoPath = errors.New("runtime: no path to target")

// pathToTile runs the BFS pathfinder over the static landscape +
// facts grid from the bot's current position to (targetX, targetY).
// Returns corner waypoints (start-side first) suitable for handing
// to action.WalkPath / action.WalkToEntityPath. If reachBorder is
// true, the search succeeds as soon as the bot can reach a tile
// orthogonally adjacent to the target across an open wall edge —
// the right mode for talk-to-NPC / open-door / attack-mob.
//
// Returns ErrNoPath if no route exists or if the bot doesn't have a
// landscape archive loaded.
func (h *Host) pathToTile(targetX, targetY int, reachBorder bool) ([][2]int, error) {
	if h.landscape == nil {
		return nil, fmt.Errorf("%w: landscape archive not loaded", ErrNoPath)
	}
	pos := h.world.Self.Position()
	if pos.X == targetX && pos.Y == targetY {
		return nil, nil // already there
	}
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0)
	if err != nil {
		return nil, fmt.Errorf("runtime: build pathfind grid: %w", err)
	}
	corners := g.FindPathToTile(pos.X, pos.Y, targetX, targetY, reachBorder)
	if len(corners) == 0 {
		return nil, ErrNoPath
	}
	out := make([][2]int, len(corners))
	for i, c := range corners {
		out[i] = [2]int{c.X, c.Y}
	}
	return out, nil
}

// walkPathTo runs the pathfinder and sends a multi-corner walk-to-point
// packet. Use this for plain "go to that tile" navigation; for action
// follow-ups (talk/attack/open) use walkAndAct which composes the
// walk with the action packet.
func (h *Host) walkPathTo(ctx context.Context, targetX, targetY int, reachBorder bool) error {
	corners, err := h.pathToTile(targetX, targetY, reachBorder)
	if err != nil {
		return err
	}
	if len(corners) == 0 {
		return nil
	}
	return action.WalkPath(ctx, h.conn, corners)
}

// walkVariant selects the right walk opcode for the action being
// chained: mob targets (NPCs, players) use WALK_TO_ENTITY (16),
// object/boundary targets and bare navigation use WALK_TO_POINT (187).
// Mirrors mudclient.walkToArea's `walkToEntity` boolean.
type walkVariant int

const (
	walkToEntity walkVariant = iota
	walkToPoint
)

// walkAndAct pathfinds to a tile, sends the appropriate walk packet,
// waits until the bot actually arrives at the final approach corner,
// then sends the supplied action. Used by AttackNpc / TalkToNpc /
// InteractWithBoundary / UseItemOnScenery for the walk-then-action
// sequence the real client emits per click.
//
// Waiting for arrival before sending the action is essential: the
// server's use-on-object / interact handlers run a withinRange /
// atObject check at the moment the action packet is processed. If we
// fire the action while the walk is still in flight (the bot may be
// taking a long detour around blocked scenery edges — e.g. a potter's
// wheel whose south edge is walled, forcing a north approach) the
// range check fails silently and the action is dropped. Polling for
// arrival mirrors the client, which queues the action locally and only
// emits it once the avatar reaches the object.
func (h *Host) walkAndAct(ctx context.Context, targetX, targetY int, reachBorder bool, variant walkVariant, sendAction func(context.Context) error) error {
	corners, pathErr := h.pathToTile(targetX, targetY, reachBorder)
	if pathErr != nil && !errors.Is(pathErr, ErrNoPath) {
		return pathErr
	}
	if len(corners) > 0 {
		var err error
		switch variant {
		case walkToEntity:
			err = action.WalkToEntityPath(ctx, h.conn, corners)
		case walkToPoint:
			err = action.WalkPath(ctx, h.conn, corners)
		}
		if err != nil {
			return fmt.Errorf("runtime: send walk path: %w", err)
		}
		// Wait for the bot to reach the final approach corner before
		// firing the action so the server's range/atObject check
		// passes. Tolerate a 1-tile radius (the server interpolates
		// between corners and the avatar may settle adjacent).
		final := corners[len(corners)-1]
		h.awaitArrival(ctx, final[0], final[1])
	}
	return sendAction(ctx)
}

// awaitArrival polls the bot's position until it is within 1 tile of
// (x, y) or movement stalls (no position change across a short window)
// or the context is cancelled. Best-effort: it never returns an error
// because the caller's action packet is still worth sending even if we
// time out (the server may queue it). Bounded so a never-arriving walk
// can't hang the routine.
func (h *Host) awaitArrival(ctx context.Context, x, y int) {
	const (
		pollInterval = 200 * time.Millisecond
		stallTimeout = 3 * time.Second
		maxWait      = 12 * time.Second
		arriveRadius = 1
	)
	deadline := time.Now().Add(maxWait)
	last := h.world.Self.Position()
	stallDeadline := time.Now().Add(stallTimeout)
	for {
		cur := h.world.Self.Position()
		if absVal(cur.X-x) <= arriveRadius && absVal(cur.Y-y) <= arriveRadius {
			return
		}
		if cur.X != last.X || cur.Y != last.Y {
			last = cur
			stallDeadline = time.Now().Add(stallTimeout)
		} else if time.Now().After(stallDeadline) {
			return // stalled — send the action anyway
		}
		if time.Now().After(deadline) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollInterval):
		}
	}
}

// --- long-range travel (go_to) ---------------------------------------------

// GoTo walks the host across the world to (targetX, targetY), even far
// beyond the ~96-tile local pathfinder window. The local BFS only sees a
// region at a time, so GoTo steps: it repeatedly picks a reachable,
// standable waypoint toward the target (within the local grid) and hands
// it to WalkToOpts — which routes around local obstacles and opens gated
// doors (its no-path fallback) — replanning each hop until it arrives.
// Greedy, so a true maze with a dead-end in the target direction can stall
// (returns an error); open-world and door-gated routes work.
func (h *Host) GoTo(ctx context.Context, targetX, targetY int) error {
	const (
		arriveRadius = 1
		// hop stays <= the server's single-walk range (30 tiles) so the
		// WalkToOpts no-path fallback (a direct Walk) is always in range,
		// and < the ~40-tile half-grid so the local BFS can reach it.
		hop     = 24
		maxHops = 600
	)
	stuck := 0
	for i := 0; i < maxHops; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		pos := h.world.Self.Position()
		// Resolve the tile we will actually arrive on. When the
		// requested target is a non-standable building/object footprint
		// (a gazetteer POI marker sits ON the object), arrivalTarget
		// snaps to the nearest standable approach tile within a tight
		// bound; when the target is standable + in-grid it returns the
		// raw target unchanged (exact radius-1 arrival preserved). The
		// snap is recomputed each hop and only fires once the target is
		// inside the local grid window, so a long cross-region GoTo
		// still steps toward the true coords until it gets close.
		arrX, arrY, _ := h.arrivalTarget(targetX, targetY)
		if absVal(pos.X-arrX) <= arriveRadius && absVal(pos.Y-arrY) <= arriveRadius {
			return nil
		}
		wx, wy := h.reachableWaypoint(arrX, arrY, hop)
		if wx == pos.X && wy == pos.Y {
			return fmt.Errorf("go_to: no reachable step toward (%d, %d) from (%d, %d)", targetX, targetY, pos.X, pos.Y)
		}
		distBefore := chebyshev(pos.X, pos.Y, arrX, arrY)
		werr := h.WalkToOpts(ctx, wx, wy, DefaultWalkOptions())
		after := h.world.Self.Position()
		distAfter := chebyshev(after.X, after.Y, arrX, arrY)
		if distAfter < distBefore {
			stuck = 0
			continue
		}
		// No progress this hop. A WalkTo error with no progress is a hard
		// block (locked door / wall) worth surfacing; otherwise nudge a
		// few times before giving up.
		if werr != nil {
			return fmt.Errorf("go_to: blocked heading to (%d, %d): %w", targetX, targetY, werr)
		}
		stuck++
		if stuck >= 4 {
			return fmt.Errorf("go_to: stuck at (%d, %d) heading to (%d, %d)", after.X, after.Y, targetX, targetY)
		}
	}
	return fmt.Errorf("go_to: gave up after %d hops short of (%d, %d)", maxHops, targetX, targetY)
}

// reachableWaypoint picks the next intermediate tile toward (tx,ty): the
// target itself if within a hop, else a point `hop` tiles along each axis,
// snapped to the nearest standable tile in the local grid.
func (h *Host) reachableWaypoint(tx, ty, hop int) (int, int) {
	pos := h.world.Self.Position()
	wx, wy := stepToward(pos.X, tx, hop), stepToward(pos.Y, ty, hop)
	return h.nearestStandable(wx, wy)
}

// nearestStandable returns (x,y) if it's a standable tile in the local
// grid, else the closest standable tile within a small radius (so a
// waypoint that lands in a wall/water is nudged onto walkable ground).
// Falls back to (x,y) unchanged if no grid or nothing standable nearby.
func (h *Host) nearestStandable(x, y int) (int, int) {
	if h.landscape == nil {
		return x, y
	}
	pos := h.world.Self.Position()
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0)
	if err != nil {
		return x, y
	}
	nx, ny, _ := snapStandable(x, y, 8, g.TileStandable)
	return nx, ny
}

// arriveSnapRadius bounds how far GoTo will snap a non-standable target
// (a building / object tile you can't stand on) to a nearby standable
// approach tile before declaring arrival. It is deliberately small: a
// POI marker sits ON its building footprint, so the standable approach
// is genuinely adjacent (the observed shortfall was 1-2 tiles). Keeping
// this tight means a genuinely unreachable / misplaced target still
// fails loudly into the cognition re-plan rather than "arriving" far
// from where the caller asked.
const arriveSnapRadius = 2

// arrivalTarget resolves the tile GoTo should actually arrive on for a
// requested (targetX, targetY). When the target is already standable
// (the common walk_to / go_to(coords) case) it returns the target
// unchanged with snapped=false, preserving exact radius-1 arrival on the
// requested tile. When the target is non-standable (a building/object
// footprint) it snaps to the nearest standable tile within
// arriveSnapRadius and returns snapped=true. If the target is outside
// the current local grid (not yet resolvable — TileStandable is false
// for out-of-grid tiles) or no standable tile sits within the bound, it
// returns the raw target with snapped=false so long-range stepping is
// unaffected and genuine walls still fail loudly. Snapping is bounded by
// arriveSnapRadius, NOT nearestStandable's wider radius-8 nudge, so we
// never claim arrival far from the requested target.
func (h *Host) arrivalTarget(targetX, targetY int) (ax, ay int, snapped bool) {
	if h.landscape == nil {
		return targetX, targetY, false
	}
	pos := h.world.Self.Position()
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0)
	if err != nil {
		return targetX, targetY, false
	}
	// Only snap once the target tile is actually inside the local grid.
	// Out-of-grid tiles read as non-standable, which would otherwise
	// snap a far-away cross-region target onto whatever standable tile
	// happens to sit near the (clamped) probe — wrong. World2Local
	// reports membership without consulting collision flags.
	if _, _, inGrid := g.World2Local(targetX, targetY); !inGrid {
		return targetX, targetY, false
	}
	if g.TileStandable(targetX, targetY) {
		return targetX, targetY, false // already standable: exact arrival
	}
	nx, ny, ok := snapStandable(targetX, targetY, arriveSnapRadius, g.TileStandable)
	if !ok {
		// Non-standable target with no standable tile within the tight
		// bound: a real wall (or a POI deep inside solid scenery). Keep
		// the raw target so the hop loop's stuck/blocked returns fire.
		return targetX, targetY, false
	}
	return nx, ny, true
}

// snapStandable returns the nearest standable tile to (x,y) within
// chebyshev radius maxR, scanning outward ring by ring (so the closest
// standable tile wins). ok reports whether a standable tile was found:
// when (x,y) itself is standable it returns (x,y,true); when nothing
// within maxR is standable it returns (x,y,false). Pure over the
// supplied standable predicate so the snap/bounding logic is testable
// without building a real pathfind grid.
func snapStandable(x, y, maxR int, standable func(int, int) bool) (int, int, bool) {
	if standable(x, y) {
		return x, y, true
	}
	for r := 1; r <= maxR; r++ {
		for dx := -r; dx <= r; dx++ {
			for dy := -r; dy <= r; dy++ {
				if absVal(dx) != r && absVal(dy) != r {
					continue // ring only
				}
				if standable(x+dx, y+dy) {
					return x + dx, y + dy, true
				}
			}
		}
	}
	return x, y, false
}

// stepToward moves cur toward tgt by at most hop tiles.
func stepToward(cur, tgt, hop int) int {
	switch {
	case tgt-cur > hop:
		return cur + hop
	case cur-tgt > hop:
		return cur - hop
	default:
		return tgt
	}
}
