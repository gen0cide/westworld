package action

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Outbound opcodes for inventory/world interaction.
//
// Source: Payload235Parser.java (decimal opcodes verified).
const (
	outPlayerFollow   byte = 165 // [short serverIndex]
	outItemDrop       byte = 246 // [short inventorySlot]
	outGroundItemTake byte = 247 // [short x] [short y] [short itemId]
	outItemCommand    byte = 90  // [short inventorySlot] — default right-click action (Bury, Eat, Drink, etc.)
)

// FollowPlayer sends the server-side follow request. The target's
// serverIndex comes from the OtherPlayerAppearance.PlayerIndex field.
// After this, the server moves us tile-by-tile to stay adjacent to
// the target — no client-side walking logic needed.
//
// Source: Payload235Parser.java:121-123 (case 165 → PLAYER_FOLLOW),
//        line 463-466 (payload structure).
func FollowPlayer(ctx context.Context, conn *session.Conn, serverIndex int) error {
	if serverIndex < 0 || serverIndex > 0xFFFF {
		return fmt.Errorf("action: follow serverIndex %d out of range", serverIndex)
	}
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(serverIndex))
	return conn.Send(outPlayerFollow, buf.Bytes())
}

// DropItem drops an inventory item from the given slot (0-29) to
// the player's current tile.
//
// Source: Payload235Parser.java:142-144 (case 246 → ITEM_DROP),
//        line 441-444 (payload [short index]).
func DropItem(ctx context.Context, conn *session.Conn, inventorySlot int) error {
	if inventorySlot < 0 || inventorySlot > 0x7FFF {
		return fmt.Errorf("action: drop slot %d out of range", inventorySlot)
	}
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(inventorySlot))
	return conn.Send(outItemDrop, buf.Bytes())
}

// PickUpItem attempts to pick up the ground item with the given
// itemID at coords (x, y). The bot must be adjacent to (or on) the
// tile.
//
// Source: Payload235Parser.java:296-303 (case 247 → GROUND_ITEM_TAKE),
//        line 431-437 (payload [short x] [short y] [short itemId]).
func PickUpItem(ctx context.Context, conn *session.Conn, x, y, itemID int) error {
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF || itemID < 0 || itemID > 0xFFFF {
		return fmt.Errorf("action: pickup coord/id (%d,%d,id=%d) out of range", x, y, itemID)
	}
	buf := v235.NewBuffer(6)
	buf.WriteUint16(uint16(x))
	buf.WriteUint16(uint16(y))
	buf.WriteUint16(uint16(itemID))
	return conn.Send(outGroundItemTake, buf.Bytes())
}

// ItemCommand fires the default right-click action on the inventory
// item in `slot` (0-29). Server resolves the action from the item's
// def (Bury for bones, Eat for food, Drink for potions, Bake for
// uncooked dough, etc.). No way to choose between command1/command2;
// the server always picks command1.
//
// Source: Payload235Parser.java case 90 → ITEM_COMMAND;
//         payload [short slot] per ItemCommandStruct.
func ItemCommand(ctx context.Context, conn *session.Conn, inventorySlot int) error {
	if inventorySlot < 0 || inventorySlot > 0x7FFF {
		return fmt.Errorf("action: item-command slot %d out of range", inventorySlot)
	}
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(inventorySlot))
	return conn.Send(outItemCommand, buf.Bytes())
}
