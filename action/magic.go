package action

import (
	"context"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Magic cast opcodes per Payload235Parser.java. Each cast variant
// has a distinct opcode keyed on what's being targeted.
const (
	outCastOnSelf      byte = 137 // [u16 spellID]
	outCastOnNpc       byte = 50  // [u16 npcIndex, u16 spellID]
	outCastOnPlayer    byte = 229 // [u16 playerIndex, u16 spellID]
	outCastOnLand      byte = 158 // [u16 x, u16 y, u16 spellID]
	outCastOnGround    byte = 249 // [u16 x, u16 y, u16 itemID, u16 spellID]
	outCastOnScenery   byte = 99  // [u16 x, u16 y, u16 spellID]
	outCastOnBoundary  byte = 180 // [u16 x, u16 y, u8 dir, u16 spellID]
	outCastOnInventory byte = 4   // [u16 slot, u16 spellID] — disambiguated from FORGOT_PASSWORD by login state
)

// CastOnSelf casts a non-targeted spell (heal, teleport, etc.).
func CastOnSelf(ctx context.Context, conn *session.Conn, spellID int) error {
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(spellID))
	return conn.Send(outCastOnSelf, buf.Bytes())
}

// CastOnNpc casts a combat or utility spell on an NPC.
func CastOnNpc(ctx context.Context, conn *session.Conn, npcServerIndex, spellID int) error {
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(npcServerIndex))
	buf.WriteUint16(uint16(spellID))
	return conn.Send(outCastOnNpc, buf.Bytes())
}

// CastOnPlayer casts on another player (PvP-only; server rejects
// outside wilderness / consensual zones).
func CastOnPlayer(ctx context.Context, conn *session.Conn, playerServerIndex, spellID int) error {
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(playerServerIndex))
	buf.WriteUint16(uint16(spellID))
	return conn.Send(outCastOnPlayer, buf.Bytes())
}

// CastOnLand casts a tile-targeted spell (telekinetic grab range
// indicators are out-of-scope; this is the raw packet).
func CastOnLand(ctx context.Context, conn *session.Conn, x, y, spellID int) error {
	buf := v235.NewBuffer(6)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteUint16(uint16(spellID))
	return conn.Send(outCastOnLand, buf.Bytes())
}

// CastOnInventory casts an item-targeted spell (e.g., enchanting
// jewelry). slotIndex is the inventory slot of the target item.
func CastOnInventory(ctx context.Context, conn *session.Conn, slotIndex, spellID int) error {
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(slotIndex))
	buf.WriteUint16(uint16(spellID))
	return conn.Send(outCastOnInventory, buf.Bytes())
}
