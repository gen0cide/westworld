package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Walk sends a walk-to-coord packet via the session. It does not wait
// for server confirmation — that arrives asynchronously as one or more
// inbound mob-update packets.
//
// For Phase 0 this is intentionally fire-and-forget. A future
// world-aware Walk will block until the server confirms arrival
// (or timeout / failure).
func Walk(ctx context.Context, conn *session.Conn, x, y int) error {
	if x < 0 || y < 0 || x > 0xFFFF || y > 0xFFFF {
		return fmt.Errorf("action: walk coord (%d, %d) out of range", x, y)
	}
	payload := v235.BuildWalkToPoint(uint16(x), uint16(y))
	return conn.Send(v235.OutWalkToPoint, payload)
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
