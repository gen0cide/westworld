package action

import (
	"context"

	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/session"
)

// Equip/unequip outbound opcodes per Payload235Parser.java.
const (
	outItemUnequip byte = 170
	outItemEquip   byte = 169
)

// EquipItem moves an inventory item into the appropriate equipment
// slot. slotIndex is the inventory slot (0..29). Server enforces
// level requirements + slot conflicts.
func EquipItem(ctx context.Context, conn *session.Conn, slotIndex int) error {
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(slotIndex))
	return conn.Send(outItemEquip, buf.Bytes())
}

// UnequipItem returns a worn item to the inventory. slotIndex is
// the inventory slot that the item occupies (RSC stores wielded
// items in regular inventory slots with a wielded flag, not in a
// separate equipment array).
func UnequipItem(ctx context.Context, conn *session.Conn, slotIndex int) error {
	buf := v235.NewBuffer(2)
	buf.WriteUint16(uint16(slotIndex))
	return conn.Send(outItemUnequip, buf.Bytes())
}
