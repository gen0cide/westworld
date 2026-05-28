package v235

// Inbound opcodes (server → client) for Payload235.
//
// All values verified against OpenRSC's Payload235Generator.java
// opcode map.
const (
	InExperience           byte = 33  // single skill XP changed
	InSystemUpdate         byte = 52  // system update countdown
	InInventory            byte = 53  // full inventory dump
	InNpcCoords            byte = 79  // bitpacked NPC update
	InDeath                byte = 83  // we died
	InStopSleep            byte = 84  // sleep ended
	InInventorySlotUpdate  byte = 90  // single inventory slot changed
	InGroundItemHandler    byte = 99  // ground item add/remove
	InFatigue              byte = 114 // fatigue value changed
	InSleepScreen          byte = 117 // sleep word captcha appeared
	InPrivateMessage       byte = 120 // private message from another player
	InInventoryRemoveItem  byte = 123 // inventory item removed
	InServerMessage        byte = 131 // chat / server message / quest message
	InStats                byte = 156 // full stats dump
	InStat                 byte = 159 // single stat (level + max + xp) changed
	// InSendLogout is in opcodes.go (165)
	InWelcomeInfo          byte = 182 // post-login welcome screen info
	// InSendPlayerCoords is in opcodes.go (191)
	InSleepwordIncorrect   byte = 194 // last sleepword was wrong
	InNpcDialogText       byte = 222 // NPC speech bubble text
	// InSendUpdatePlayers is in opcodes.go (234)
	InNpcDialogOptions byte = 245 // NPC asks to pick a dialog option

	// Trade inbound packets.
	InTradeAccepted       byte = 15
	InTradeOpenConfirm    byte = 20
	InTradeWindow         byte = 92  // server asks us to open trade window
	InTradeOtherItems     byte = 97  // other side's items in current trade
	InTradeClose          byte = 128 // trade cancelled
	InTradeOtherAccepted  byte = 162 // other player clicked accept

	// Shop inbound.
	InShopOpen  byte = 101 // shop window opened with inventory
	InShopClose byte = 137 // shop window closed

	// Bank inbound packets.
	InBankOpen   byte = 42  // bank window opened — full slot dump
	InBankUpdate byte = 249 // single bank slot updated
	InBankClose  byte = 203 // bank window closed

	// Prayer inbound.
	InPrayersActive byte = 206 // [N bytes, one per prayer slot: 0/1]

	// Boundary updates — opened doors, cut webs, etc.
	InBoundaryHandler byte = 91 // SEND_BOUNDARY_HANDLER (dynamic boundary state)

	// Duel inbound packets. Naming mirrors the trade inbound block
	// above — same two-screen handshake shape plus a rules toggle.
	InDuelItems         byte = 6   // SEND_DUEL_OPPONENTS_ITEMS — opp's stake
	InDuelSettings      byte = 30  // SEND_DUEL_SETTINGS — rule toggles
	InDuelWindow        byte = 176 // SEND_DUEL_WINDOW — offer screen opened
	InDuelAccepted      byte = 210 // SEND_DUEL_ACCEPTED — our accept echoed back
	InDuelOtherAccepted byte = 253 // SEND_DUEL_OTHER_ACCEPTED — opp clicked accept
	InDuelConfirmWindow byte = 172 // SEND_DUEL_CONFIRMWINDOW — final review screen
	InDuelClose         byte = 225 // SEND_DUEL_CLOSE — duel cancelled
)
