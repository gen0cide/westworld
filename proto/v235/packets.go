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

// BuildWalkToPoint encodes a walk-to-coordinate packet. The simplest
// form sends just the destination as the "first step" with no
// additional waypoints; the server resolves a path.
//
// Layout (per Payload235Parser.java:534-551):
//
//	[2 bytes] firstStepX (big-endian short)
//	[2 bytes] firstStepY (big-endian short)
//	[repeating, optional]:
//	  [2 bytes] waypointX
//	  [2 bytes] waypointY
//
// For Phase 0 we only send the destination as a single "first step".
func BuildWalkToPoint(x, y uint16) []byte {
	buf := NewBuffer(4)
	buf.WriteUint16(x)
	buf.WriteUint16(y)
	return buf.Bytes()
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
