package v235

// Outbound opcodes (client → server).
//
// Verified against OpenRSC's Payload235Parser.java:27-42 (server-side
// parsing of client-sent packets — for our purposes these are the
// opcode values we encode for outbound packets).
const (
	OutLogin         byte = 0   // login packet (RSA + XTEA blocks); sent pre-ISAAC
	OutRegister      byte = 2   // account registration
	OutHeartbeat     byte = 67  // periodic keep-alive, empty payload
	OutWalkToEntity  byte = 16  // walk and target an entity
	OutWalkToPoint   byte = 187 // walk to coords; what Phase 0 uses
	OutConfirmLogout byte = 31  // ack a server-initiated logout
	OutLogout        byte = 102 // request a clean logout
	// OutPlayerAppearance confirms the character's appearance, clearing the
	// server's "changing appearance" state for a fresh account. Until this is
	// sent, the server holds the player on the appearance screen and streams
	// only SEND_APPEARANCE_KEEPALIVE — withholding the entire world update
	// stream (own position, NPCs, objects). Payload (8 bytes):
	// headRestrictions, headType, bodyType, mustEqual2(=2), hairColour,
	// topColour, trouserColour, skinColour. (Payload235Parser case 235.)
	OutPlayerAppearance byte = 235
)

// Inbound opcodes (server → client).
//
// Verified against OpenRSC's Payload235Generator.java:25-93 (server-side
// emission of packets to the client — for our purposes these are the
// opcode values we decode from inbound packets).
const (
	InSendPlayerCoords  byte = 191 // bitpacked mob update including own position
	InSendUpdatePlayers byte = 234 // bitpacked appearance updates for visible players
	InSendLogout        byte = 165 // server confirms a clean logout
)

// LoginResponse codes from OpenRSC's util/rsc/LoginResponse.java.
//
// Codes with the 0x40 bit set indicate successful login; the specific
// value encodes the player's group (admin, mod, regular, etc.).
const (
	LoginUnsuccessful      byte = 0
	LoginReconnectOK       byte = 1
	LoginInvalidCredential byte = 3
	LoginAccountAlreadyOn  byte = 4
	LoginAttemptsExceeded  byte = 7

	// LoginSuccessGroup10 = 64 // regular player
	// LoginSuccessGroup1  = 86 // admin
	// Many other variants exist; we only need to check the 0x40 bit.
)

// LoginSuccessful returns true if the given response code indicates a
// successful login. Per LoginPacketHandler.java:202, the server treats
// any code with the 0x40 bit set as success.
func LoginSuccessful(code byte) bool {
	return (code & 0x40) != 0
}
