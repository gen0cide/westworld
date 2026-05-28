package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Use-with-X opcodes per Payload235Parser.java. Each is a distinct
// outbound packet because the server reads different fields for
// each target type (no generic "use" surface — the client picks
// the right opcode at click time).
const (
	outUseItemOnItem        byte = 91  // ITEM_USE_ITEM        — inv slot on inv slot
	outUseItemOnGroundItem  byte = 53  // GROUND_ITEM_USE_ITEM — inv slot on ground item
	outUseItemOnScenery     byte = 115 // USE_ITEM_ON_SCENERY  — inv slot on world object
	outUseItemOnBoundary    byte = 161 // USE_WITH_BOUNDARY    — inv slot on door/wall/gate
	outUseItemOnNpc         byte = 50  // NPC_USE_ITEM         — inv slot on NPC
	outUseItemOnPlayer      byte = 113 // PLAYER_USE_ITEM      — inv slot on player
	outObjectCommand        byte = 136 // OBJECT_COMMAND       — primary click on scenery (opt 1)
	outObjectCommand2       byte = 79  // OBJECT_COMMAND2      — secondary click on scenery (opt 2)
)

// ObjectCommand fires opcode 136 (primary "click") or 79 (secondary).
// Used for the default scenery interaction — "Chop" on a tree,
// "Mine" on a rock, "Search" on a chest, "Climb-Up" on a ladder.
// The exact verb depends on the SceneryDef.Command1 / Command2 — the
// caller picks via the `option` arg (1 or 2).
//
// Payload per Payload235Parser.java OBJECT_COMMAND / OBJECT_COMMAND2:
//
//	[short] x
//	[short] y
//
// (No object id — the server resolves from the tile + facts.)
func ObjectCommand(ctx context.Context, conn *session.Conn, x, y, option int) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		return fmt.Errorf("action: object coord (%d, %d) out of uint16", x, y)
	}
	opcode := outObjectCommand
	if option == 2 {
		opcode = outObjectCommand2
	}
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	return conn.Send(opcode, buf.Bytes())
}

// UseItemOnItem fires opcode 91. Combines two inventory items
// (e.g. needle + cloth, chisel + gem, knife + log).
//
// Payload per Payload235Parser.java:
//
//	[short] slot1
//	[short] slot2
func UseItemOnItem(ctx context.Context, conn *session.Conn, slot1, slot2 int) error {
	if slot1 < 0 || slot1 > 0xFFFF || slot2 < 0 || slot2 > 0xFFFF {
		return fmt.Errorf("action: slot out of range (%d, %d)", slot1, slot2)
	}
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(slot1))
	buf.WriteUint16(uint16(slot2))
	return conn.Send(outUseItemOnItem, buf.Bytes())
}

// UseItemOnBoundary fires opcode 161. Used to e.g. use a key on
// a locked door. The boundary's (x, y, direction) identifies the
// edge; slot is the inventory position of the item being applied.
//
// Payload per Payload235Parser.java USE_WITH_BOUNDARY:
//
//	[short] x
//	[short] y
//	[byte]  direction
//	[short] slotID
func UseItemOnBoundary(ctx context.Context, conn *session.Conn, x, y, direction, slot int) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		return fmt.Errorf("action: boundary coord (%d, %d) out of uint16", x, y)
	}
	if slot < 0 || slot > 0xFFFF {
		return fmt.Errorf("action: slot %d out of uint16", slot)
	}
	if direction < 0 || direction > 0xFF {
		return fmt.Errorf("action: direction %d out of byte", direction)
	}
	buf := v235.NewBuffer(7)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteByte(byte(direction))
	buf.WriteUint16(uint16(slot))
	return conn.Send(outUseItemOnBoundary, buf.Bytes())
}

// UseItemOnGroundItem fires opcode 53. Used to combine an inventory
// item with a ground item (e.g. fill a vial from a bucket on the
// ground, certain quest interactions).
//
// Payload per Payload235Parser.java GROUND_ITEM_USE_ITEM:
//
//	[short] x
//	[short] y
//	[short] groundItemID
//	[short] slotID
func UseItemOnGroundItem(ctx context.Context, conn *session.Conn, x, y, groundItemID, slot int) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		return fmt.Errorf("action: ground item coord (%d, %d) out of uint16", x, y)
	}
	if groundItemID < 0 || groundItemID > 0xFFFF {
		return fmt.Errorf("action: groundItemID %d out of uint16", groundItemID)
	}
	if slot < 0 || slot > 0xFFFF {
		return fmt.Errorf("action: slot %d out of uint16", slot)
	}
	buf := v235.NewBuffer(8)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteUint16(uint16(groundItemID))
	buf.WriteUint16(uint16(slot))
	return conn.Send(outUseItemOnGroundItem, buf.Bytes())
}

// UseItemOnNpc fires opcode 50. Used for thieving (pickpocket via
// "Use" hint), trade-prep, item-give-to-NPC patterns. Server index
// identifies the NPC; slot is the inventory item.
//
// Payload per Payload235Parser.java NPC_USE_ITEM / PLAYER_USE_ITEM:
//
//	[short] serverIndex
//	[short] slotID
func UseItemOnNpc(ctx context.Context, conn *session.Conn, serverIndex, slot int) error {
	if serverIndex < 0 || serverIndex > 0xFFFF {
		return fmt.Errorf("action: npc serverIndex %d out of uint16", serverIndex)
	}
	if slot < 0 || slot > 0xFFFF {
		return fmt.Errorf("action: slot %d out of uint16", slot)
	}
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(serverIndex))
	buf.WriteUint16(uint16(slot))
	return conn.Send(outUseItemOnNpc, buf.Bytes())
}

// UseItemOnPlayer fires opcode 113. Used for trade-init (use item
// on another player to offer it), gift-giving patterns. Same
// payload shape as UseItemOnNpc.
func UseItemOnPlayer(ctx context.Context, conn *session.Conn, serverIndex, slot int) error {
	if serverIndex < 0 || serverIndex > 0xFFFF {
		return fmt.Errorf("action: player serverIndex %d out of uint16", serverIndex)
	}
	if slot < 0 || slot > 0xFFFF {
		return fmt.Errorf("action: slot %d out of uint16", slot)
	}
	buf := v235.NewBuffer(4)
	buf.WriteUint16(uint16(serverIndex))
	buf.WriteUint16(uint16(slot))
	return conn.Send(outUseItemOnPlayer, buf.Bytes())
}

// UseItemOnScenery fires opcode 115. Used for use-on-object
// patterns: log on fire (cook), ore on furnace (smelt),
// pestle-and-mortar on herb, etc.
//
// Payload per Payload235Parser.java USE_ITEM_ON_SCENERY:
//
//	[short] x
//	[short] y
//	[short] slotID
//
// (No direction byte for scenery, unlike boundary.)
func UseItemOnScenery(ctx context.Context, conn *session.Conn, x, y, slot int) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		return fmt.Errorf("action: scenery coord (%d, %d) out of uint16", x, y)
	}
	if slot < 0 || slot > 0xFFFF {
		return fmt.Errorf("action: slot %d out of uint16", slot)
	}
	buf := v235.NewBuffer(6)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteUint16(uint16(slot))
	return conn.Send(outUseItemOnScenery, buf.Bytes())
}
