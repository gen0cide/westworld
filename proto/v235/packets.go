package v235

import "fmt"

// Outbound packet builders.
//
// These return the payload BODY for each packet type. To send over the
// wire, wrap with EncodeFrame(isaac.EncodeOpcode(opcode), payload).
//
// Sources:
//   - net/rsc/parsers/impl/Payload235Parser.java:27-42 (opcode → enum)
//   - net/rsc/parsers/impl/Payload235Parser.java:534-551 (walk packet)

// BuildHeartbeat returns the heartbeat packet payload. Heartbeats have
// no payload — they're just an opcode.
func BuildHeartbeat() []byte { return nil }

// BuildLogout returns the logout packet payload. Empty.
func BuildLogout() []byte { return nil }

// BuildConfirmLogout returns the "confirm a server-initiated logout"
// packet payload. Empty.
func BuildConfirmLogout() []byte { return nil }

// BuildWalkToPoint encodes a walk-to-coordinate packet with a single
// destination tile that must be adjacent to the player. The server's
// WalkingQueue checks each step against `PathValidation.checkAdjacent`
// and resets the path the moment it finds a non-adjacent step, so for
// destinations more than one tile away callers MUST use BuildWalkPath.
//
// Layout per Payload235Parser.java:
//
//	[2 bytes] firstStepX (big-endian short)
//	[2 bytes] firstStepY (big-endian short)
//	[repeating, optional, per waypoint]:
//	  [1 byte] deltaX (signed)  — added to firstStepX for that step's
//	  [1 byte] deltaY (signed)    absolute tile
func BuildWalkToPoint(x, y uint16) []byte {
	buf := NewBuffer(4)
	buf.WriteUint16(x)
	buf.WriteUint16(y)
	return buf.Bytes()
}

// BuildWalkPath encodes a multi-step walk packet from the absolute
// tile coordinates of each step on the path. `steps[0]` is the player's
// first move (must be adjacent to the player's current tile);
// subsequent entries must each be adjacent to the previous. Each
// step's coordinate delta from steps[0] is encoded as a pair of
// signed bytes, so steps[i] must satisfy
// |steps[i].X - steps[0].X| ≤ 127 and same for Y.
//
// Returns nil if steps is empty.
func BuildWalkPath(steps [][2]int) ([]byte, error) {
	if len(steps) == 0 {
		return nil, nil
	}
	first := steps[0]
	if first[0] < 0 || first[0] > 0xFFFF || first[1] < 0 || first[1] > 0xFFFF {
		return nil, fmt.Errorf("v235: walk firstStep (%d, %d) out of uint16 range", first[0], first[1])
	}
	buf := NewBuffer(4 + 2*(len(steps)-1))
	buf.WriteUint16(uint16(first[0]))
	buf.WriteUint16(uint16(first[1]))
	for i := 1; i < len(steps); i++ {
		dx := steps[i][0] - first[0]
		dy := steps[i][1] - first[1]
		if dx < -128 || dx > 127 || dy < -128 || dy > 127 {
			return nil, fmt.Errorf("v235: walk step %d delta (%d, %d) exceeds signed-byte range", i, dx, dy)
		}
		buf.WriteByte(byte(int8(dx)))
		buf.WriteByte(byte(int8(dy)))
	}
	return buf.Bytes(), nil
}

// ParsePlayerCoords is a stub for opcode InSendPlayerCoords (191). The
// real packet is bitpacked and complex; Phase 0 just acknowledges
// receipt. Returning the raw payload length lets us at least
// confirm position updates are arriving.
func ParsePlayerCoords(payload []byte) (rawLen int, err error) {
	if len(payload) < 1 {
		return 0, fmt.Errorf("v235: empty SendPlayerCoords payload")
	}
	return len(payload), nil
}

// ParseLogoutAck handles the server's logout confirmation
// (InSendLogout = opcode 165). Empty payload.
func ParseLogoutAck(payload []byte) error {
	if len(payload) != 0 {
		// Not fatal — some server variants may include data. Just log.
	}
	return nil
}
