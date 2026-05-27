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
	InNpcDialogText        byte = 222 // NPC speech bubble text
	// InSendUpdatePlayers is in opcodes.go (234)
	InNpcDialogOptions     byte = 245 // NPC asks to pick a dialog option
)
