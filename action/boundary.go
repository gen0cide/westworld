package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Boundary-interaction outbound opcodes.
//
// Source: Payload235Parser.java:79-84. Boundaries are walls/doors/fences
// stored as edge-attached objects. InteractWithBoundary (14) is the
// default left-click (e.g., "Open" on a closed door, "Climb" on a
// fence). InteractWithBoundary2 (127) is the right-click alternate.
const (
	outInteractWithBoundary  byte = 14
	outInteractWithBoundary2 byte = 127
)

// InteractWithBoundary fires the default left-click action on a
// boundary at (x, y) with the given edge direction (0..3).
//
// Use this to open closed doors, climb ladders/fences, etc. The server
// walks the player to the adjacent tile then performs the action.
//
// Payload: [short x] [short y] [byte direction]
func InteractWithBoundary(ctx context.Context, conn *session.Conn, x, y, direction int) error {
	return sendBoundaryPacket(ctx, conn, outInteractWithBoundary, x, y, direction, "InteractWithBoundary")
}

// InteractWithBoundary2 fires the right-click alternate action on a
// boundary (e.g., "Pick lock" on certain doors).
func InteractWithBoundary2(ctx context.Context, conn *session.Conn, x, y, direction int) error {
	return sendBoundaryPacket(ctx, conn, outInteractWithBoundary2, x, y, direction, "InteractWithBoundary2")
}

func sendBoundaryPacket(ctx context.Context, conn *session.Conn, opcode byte, x, y, direction int, name string) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		return fmt.Errorf("action: %s coord (%d, %d) out of range", name, x, y)
	}
	if direction < 0 || direction > 0xFF {
		return fmt.Errorf("action: %s direction %d out of byte range", name, direction)
	}
	buf := v235.NewBuffer(5)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteByte(byte(direction))
	return conn.Send(opcode, buf.Bytes())
}
