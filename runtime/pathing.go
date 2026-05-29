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
