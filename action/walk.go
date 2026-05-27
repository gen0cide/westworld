package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// MaxClickRange is the maximum tile distance from the player at which
// a walk-to click is plausible — matches OpenRSC's effective view
// radius of 15 tiles (server-side check in Mob.java:369:
// `VIEW_DISTANCE * 8 - 1` = 2 * 8 - 1 = 15).
//
// Sending a walk packet to a tile farther than this is something an
// authentic RSC client physically cannot do (the world isn't rendered
// beyond viewport). Allowing it is both a believability tell and may
// cause the server to behave inconsistently.
//
// Callers needing to walk farther should use Host.WalkTo, which
// chunks long journeys into multiple in-range segments.
const MaxClickRange = 15

// ErrOutOfRange is returned by Walk when the target tile is too far
// from the player's current position to be a plausible click.
var ErrOutOfRange = errors.New("action: walk target outside view range")

// Walk sends a walk-to-coord packet via the session. The target MUST
// be within MaxClickRange of `currentX`, `currentY`. Returns
// ErrOutOfRange if not.
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
	payload := v235.BuildWalkToPoint(uint16(targetX), uint16(targetY))
	return conn.Send(v235.OutWalkToPoint, payload)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
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
