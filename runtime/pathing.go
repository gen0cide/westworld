package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/pathfind"
	"github.com/gen0cide/westworld/world"
)

// ErrNoPath means the pathfinder couldn't find a route from the
// player's current location to the requested target inside the local
// 96x96 region. Either the goal is outside the loaded region or
// there's no reachable path through the landscape walls / scenery.
var ErrNoPath = errors.New("runtime: no path to target")

// liveState snapshots the host's dynamic boundary (door/gate) + scenery
// overrides into a pathfind.LiveState so BuildGrid reflects the live world:
// opened doors clear, closed doors block, depleted scenery clears, spawned
// blockers (lit fires) block. Returns nil when the world has no dynamic stores
// (keeps the pure-static grid for tests).
func (h *Host) liveState() *pathfind.LiveState {
	if h.world == nil {
		return nil
	}
	ls := &pathfind.LiveState{}
	if h.world.Boundaries != nil {
		for k, id := range h.world.Boundaries.All() {
			ls.Boundaries = append(ls.Boundaries, pathfind.LiveBoundary{
				X: k.X, Y: k.Y, Dir: k.Dir, ID: id, Removed: id < 0,
			})
		}
	}
	if h.world.Scenery != nil {
		for _, r := range h.world.Scenery.All() {
			ls.Scenery = append(ls.Scenery, pathfind.LiveScenery{X: r.X, Y: r.Y, ID: r.ID})
		}
		for _, t := range h.world.Scenery.Removed() {
			ls.Scenery = append(ls.Scenery, pathfind.LiveScenery{X: t[0], Y: t[1], Removed: true})
		}
	}
	return ls
}

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
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0, h.liveState())
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
	_ = variant // retained for call-site compatibility; see opcode note below
	corners, pathErr := h.pathToTile(targetX, targetY, reachBorder)
	if pathErr != nil && !errors.Is(pathErr, ErrNoPath) {
		return pathErr
	}
	// ALL action walks use opcode 16 (WALK_TO_ENTITY) — both references do
	// this for objects, walls, AND mobs (mudclient.java:17607-17666 pass
	// walkToEntity=true; Plutonium actions.go:7-60 identical); opcode 187 is
	// only for plain ground walks. On no-path, the authentic client sends
	// the RAW destination via the same packet: the server's WalkToAction
	// defers the queued interaction until atObject passes per tick, so a
	// far/blocked target is never range-dropped — but we must NEVER fire a
	// bare action with no walk at all, which would plant a LATENT
	// WalkToAction that can fire minutes later when the host wanders near.
	if len(corners) > 0 {
		if err := action.WalkToEntityPath(ctx, h.conn, corners); err != nil {
			return fmt.Errorf("runtime: send walk path: %w", err)
		}
		// Wait for the bot to reach the final approach corner before
		// firing the action so the server's range/atObject check
		// passes. Tolerate a 1-tile radius (the server interpolates
		// between corners and the avatar may settle adjacent).
		final := corners[len(corners)-1]
		h.awaitArrival(ctx, final[0], final[1])
	} else {
		if err := action.WalkToEntityPath(ctx, h.conn, [][2]int{{targetX, targetY}}); err != nil {
			return fmt.Errorf("runtime: send raw entity walk: %w", err)
		}
		h.awaitArrival(ctx, targetX, targetY)
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
	// Cross-plane first: planes stack in Y at 944-tile bands and there is
	// NO walkable connection between bands — stepping Y toward a different
	// band walks into dead space forever (the boxed-upstairs deadlock).
	// Route plane transitions deliberately: climb a ladder/staircase per
	// band until the host is on the target's floor, then walk normally.
	for guard := 0; guard < 4; guard++ {
		cur := h.world.Self.Position()
		curPlane := world.PlaneOf(cur.Y)
		tgtPlane := world.PlaneOf(targetY)
		if curPlane == tgtPlane {
			break
		}
		if err := h.traversePlane(ctx, curPlane < tgtPlane, targetX, targetY%world.PlaneHeight); err != nil {
			return fmt.Errorf("go_to: cross-plane (floor %d -> %d): %w", curPlane, tgtPlane, err)
		}
	}
	return h.goToOnPlane(ctx, targetX, targetY)
}

// goToOnPlane is the SAME-PLANE long-range walker: a corridor of waypoints
// toward the target, each hop handed to WalkToOpts (which routes around
// local obstacles and opens doors/gates en route).
func (h *Host) goToOnPlane(ctx context.Context, targetX, targetY int) error {
	const (
		arriveRadius = 1
		hop          = 24
		maxHops      = 600
	)
	// Corridor-first: a real route from the world oracle (openables
	// crossable — the walk layer opens them en route; learned-locked
	// obstacles avoided via the ledger) beats greedy stepping, which
	// fixates on dead-ends in concave terrain. Greedy remains the
	// fallback for unmapped areas / a nil oracle.
	if h.worldOracle != nil {
		pos := h.world.Self.Position()
		var avoid map[[2]int]bool
		if h.blocked != nil {
			for _, e := range h.blocked.All() {
				if avoid == nil {
					avoid = map[[2]int]bool{}
				}
				avoid[[2]int{e.X, e.Y}] = true
			}
		}
		if wps := h.worldOracle.Route(pos.X, pos.Y, targetX, targetY, avoid); len(wps) > 0 {
			if err := h.walkCorridor(ctx, wps, targetX, targetY); err == nil {
				return nil
			} else if !errors.Is(err, errCorridorFailed) {
				return err // typed (DOOR_LOCKED etc.) — surface, don't mask
			}
			// Corridor failed mid-way (world changed under us): fall
			// through to the greedy stepper from wherever we stand.
		}
	}
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
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0, h.liveState())
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
	g, err := pathfind.BuildGrid(h.landscape, h.facts, pos.X, pos.Y, 0, h.liveState())
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

// errCorridorFailed signals walkCorridor lost the route for a re-plannable
// reason (a waypoint stalled without a typed cause) — the caller falls back
// to the greedy stepper rather than failing the whole go_to.
var errCorridorFailed = errors.New("corridor walk failed")

// walkCorridor walks an oracle-planned waypoint list, each leg via
// WalkToOpts (which opens doors/gates en route). Typed leg failures
// (DoorLockedError) bubble; an untyped stall returns errCorridorFailed so
// the caller can fall back. Arrival at the FINAL target ends the corridor
// early regardless of remaining waypoints.
func (h *Host) walkCorridor(ctx context.Context, wps [][2]int, targetX, targetY int) error {
	const arriveRadius = 1
	for _, wp := range wps {
		if err := ctx.Err(); err != nil {
			return err
		}
		pos := h.world.Self.Position()
		if absVal(pos.X-targetX) <= arriveRadius && absVal(pos.Y-targetY) <= arriveRadius {
			return nil
		}
		if err := h.WalkToOpts(ctx, wp[0], wp[1], DefaultWalkOptions()); err != nil {
			var doorErr *DoorLockedError
			if errors.As(err, &doorErr) {
				return err // typed: locked obstacle on the corridor — surface
			}
			h.log.Warn("go_to: corridor leg failed; falling back to greedy",
				"waypoint", fmt.Sprintf("(%d, %d)", wp[0], wp[1]), "err", err)
			return errCorridorFailed
		}
	}
	// Corridor exhausted: the last waypoint is the goal's snap tile; do the
	// final exact approach.
	pos := h.world.Self.Position()
	if absVal(pos.X-targetX) <= arriveRadius && absVal(pos.Y-targetY) <= arriveRadius {
		return nil
	}
	if err := h.WalkToOpts(ctx, targetX, targetY, DefaultWalkOptions()); err != nil {
		var doorErr *DoorLockedError
		if errors.As(err, &doorErr) {
			return err
		}
		return errCorridorFailed
	}
	return nil
}

// traversePlane moves the host ONE floor up or down via the nearest
// matching ladder/staircase on its current band: walk to it (the rect-aware
// interact handles the approach), climb, and VERIFY the Y band actually
// changed (the server teleports ±944 — or to explicit coords for special
// ladders, so we check the band, not the exact offset). The climb command
// picks direction: "climb-up"/"go up" only go up, "climb-down"/"go down"
// only down; bare "climb"/"pull" count for either.
func (h *Host) traversePlane(ctx context.Context, up bool, aimX, aimFY int) error {
	if h.facts == nil {
		return errors.New("no facts loaded — cannot locate a ladder")
	}
	cur := h.world.Self.Position()
	plane := world.PlaneOf(cur.Y)
	// The right ladder is the one UNDER/OVER THE TARGET's building (its
	// top must land in the target's component), so rank candidates by
	// distance to the TARGET's footprint — not to the host. And only
	// consider ladders the host can actually reach on THIS floor (same
	// oracle component), or an unreachable candidate strands the walk.
	aimY := plane*world.PlaneHeight + aimFY
	hostComp := int32(-1)
	if h.worldOracle != nil {
		hostComp, _, _, _ = h.worldOracle.CompNear(cur.X, cur.Y)
	}

	type ladder struct {
		x, y, d int
		name    string
	}
	var best *ladder
	for _, loc := range h.facts.SceneryLocs {
		if world.PlaneOf(loc.Y) != plane {
			continue
		}
		def := h.facts.SceneryDef(loc.DefID)
		if def == nil {
			continue
		}
		cmd := strings.ToLower(def.Command1)
		if !ladderCommands[cmd] {
			continue
		}
		goesUp := strings.Contains(cmd, "up")
		goesDown := strings.Contains(cmd, "down")
		if up && goesDown {
			continue
		}
		if !up && goesUp {
			continue
		}
		if hostComp >= 0 {
			if c, _, _, ok := h.worldOracle.CompNear(loc.X, loc.Y); !ok || c != hostComp {
				continue // not reachable on this floor
			}
		}
		d := chebyshev(aimX, aimY, loc.X, loc.Y)
		if best == nil || d < best.d {
			best = &ladder{x: loc.X, y: loc.Y, d: d, name: def.Name}
		}
	}
	if best == nil {
		dir := "down"
		if up {
			dir = "up"
		}
		return fmt.Errorf("no ladder/staircase going %s found on floor %d", dir, plane)
	}
	h.log.Info("go_to: plane transition",
		"via", best.name, "at", fmt.Sprintf("(%d, %d)", best.x, best.y),
		"dist", best.d, "up", up, "from_floor", plane,
	)
	// Far ladders need the same-plane walker first; InteractAt's approach
	// only sees the local grid window.
	if chebyshev(cur.X, cur.Y, best.x, best.y) > 30 {
		if err := h.goToOnPlane(ctx, best.x, best.y); err != nil {
			return fmt.Errorf("walking to the %s: %w", best.name, err)
		}
	}
	if err := h.InteractAt(ctx, best.x, best.y, 1); err != nil {
		return fmt.Errorf("climb %s: %w", best.name, err)
	}
	// Await the band change (bounded). The climb is a server-side teleport;
	// position streams in on the next coord update.
	deadline := time.Now().Add(5 * time.Second)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
		if world.PlaneOf(h.world.Self.Position().Y) != plane {
			return nil // floor changed — success
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("climbed the %s at (%d, %d) but the floor did not change", best.name, best.x, best.y)
		}
	}
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
