package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// MaxClickRange is the maximum tile distance from the player at which
// a walk-to click is plausible.
//
// Authentic value: 15 (matches OpenRSC's server-side Pythagorean
// range check in Mob.java:369: VIEW_DISTANCE*8-1 = 2*8-1 = 15).
// Phase-1.7 testing override: 30 — gives us more room for follow/walk
// experiments against alex. To match real-client behaviour exactly
// later, drop this back to 15 and ensure the server's view_distance
// stays at 2 chunks. To genuinely double the view, also change
// preservation.conf `view_distance: 2` → `view_distance: 4` and
// restart the server.
const MaxClickRange = 30

// ErrOutOfRange is returned by Walk when the target tile is too far
// from the player's current position to be a plausible click.
var ErrOutOfRange = errors.New("action: walk target outside view range")

// Walk sends a walk-to-coord packet via the session. The target MUST
// be within MaxClickRange of `currentX`, `currentY`. Returns
// ErrOutOfRange if not.
//
// We send the destination as a single firstStep (no extra waypoints)
// and let the server's `Path.addStep` interpolate the tile-by-tile
// path internally. That matters because `Path.addStep` only runs the
// PLAYER_BLOCKING check once per call (against the addStep target,
// when its inner loop reaches the final tile). If we sent every tile
// as an explicit waypoint, the server would `addStep` for each, and
// any other player standing on an intermediate tile would silently
// truncate the path. Letting the server interpolate keeps the
// blocking check on the actual destination only — matching real
// RSC-client behaviour.
//
// This single-firstStep form also relies on the server's `Path.addStep`
// per-tile checkAdjacent collision check to stop the path at walls
// for us; we don't need to A* around scenery client-side.
//
// The walk is fire-and-forget at this layer; the host's main loop
// observes the position stream to know if/when the walk completes.
func Walk(ctx context.Context, conn *session.Conn, currentX, currentY, targetX, targetY int) error {
	if targetX < 0 || targetY < 0 || targetX > 0xFFFF || targetY > 0xFFFF {
		return fmt.Errorf("action: walk coord (%d, %d) out of range", targetX, targetY)
	}
	dx := targetX - currentX
	dy := targetY - currentY
	if absInt(dx) > MaxClickRange || absInt(dy) > MaxClickRange {
		return fmt.Errorf("%w: target (%d, %d) is %d east/%d south of (%d, %d); max is %d",
			ErrOutOfRange, targetX, targetY, dx, dy, currentX, currentY, MaxClickRange)
	}
	if currentX == targetX && currentY == targetY {
		return nil // already there
	}
	payload := v235.BuildWalkToPoint(uint16(targetX), uint16(targetY))
	return conn.Send(v235.OutWalkToPoint, payload)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// MaxWalkCorners is the largest number of waypoints the client can
// fit in a single walk packet. Mudclient.walkToArea sends at most
// `count - 25` corners worth of deltas, so the same cap applies here.
const MaxWalkCorners = 25

// WalkPath sends a multi-corner walk-to-point packet. `corners` are
// the path's direction-change waypoints in WALK ORDER (start-side
// first, goal last) in absolute world coordinates. The server's
// Path.addStep interpolates the straight or diagonal segments between
// corners.
//
// Mirrors mudclient.walkToArea(..., walkToEntity=false):
//
//	firstStep = absolute (X, Y) of the corner closest to the player
//	waypoints  = signed-byte (dx, dy) deltas FROM firstStep for each
//	             subsequent corner
//
// If the path has more than MaxWalkCorners corners, only the first
// MaxWalkCorners are sent; the caller should chain additional walks
// after observing the bot's new position.
func WalkPath(ctx context.Context, conn *session.Conn, corners [][2]int) error {
	if len(corners) == 0 {
		return nil
	}
	if len(corners) > MaxWalkCorners {
		corners = corners[:MaxWalkCorners]
	}
	payload, err := v235.BuildWalkPath(corners)
	if err != nil {
		return fmt.Errorf("action: build walk path: %w", err)
	}
	return conn.Send(v235.OutWalkToPoint, payload)
}

// WalkToEntityPath is identical to WalkPath but uses opcode 16
// (WALK_TO_ENTITY) instead of 187 (WALK_TO_POINT). The client sends
// the entity variant when the destination is an NPC, object, or
// ground item — the server treats the path the same way but the
// opcode hints "approach to interact, not to land on the tile" so
// it'll stop one step short rather than refusing to walk onto the
// final tile.
func WalkToEntityPath(ctx context.Context, conn *session.Conn, corners [][2]int) error {
	if len(corners) == 0 {
		return nil
	}
	if len(corners) > MaxWalkCorners {
		corners = corners[:MaxWalkCorners]
	}
	payload, err := v235.BuildWalkPath(corners)
	if err != nil {
		return fmt.Errorf("action: build walk-to-entity path: %w", err)
	}
	return conn.Send(v235.OutWalkToEntity, payload)
}

// Logout sends a clean-logout request. The server's confirmation
// arrives as an inbound opcode 165 (SendLogout) packet.
func Logout(ctx context.Context, conn *session.Conn) error {
	return conn.Send(v235.OutLogout, v235.BuildLogout())
}

// Heartbeat sends a single keepalive packet. Should be called
// periodically (every ~5 seconds) by a higher-level scheduler.
func Heartbeat(ctx context.Context, conn *session.Conn) error {
	return conn.Send(v235.OutHeartbeat, v235.BuildHeartbeat())
}
