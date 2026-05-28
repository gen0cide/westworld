package runtime

import (
	"context"

	"github.com/gen0cide/westworld/action"
)

// CastOnSelf is a non-targeted spell (heal, teleport, transformation
// auras). spellID is the spellbook index 0..N.
func (h *Host) CastOnSelf(ctx context.Context, spellID int) error {
	return action.CastOnSelf(ctx, h.conn, spellID)
}

// CastOnNpc walks adjacent if needed (most attack spells require
// line-of-sight but not adjacency — for safety we use the same
// pathing helper as melee combat), then sends the cast.
func (h *Host) CastOnNpc(ctx context.Context, npcServerIndex, spellID int) error {
	return action.CastOnNpc(ctx, h.conn, npcServerIndex, spellID)
}

// CastOnPlayer for PvP / dueling. Server rejects outside legal
// zones.
func (h *Host) CastOnPlayer(ctx context.Context, playerServerIndex, spellID int) error {
	return action.CastOnPlayer(ctx, h.conn, playerServerIndex, spellID)
}

// CastOnLand fires a tile-targeted spell (e.g., Telekinetic Grab
// to a ground item — actually GRAB uses CAST_ON_GROUND_ITEM; this
// is for AOE / area-effect spells).
func (h *Host) CastOnLand(ctx context.Context, x, y, spellID int) error {
	return action.CastOnLand(ctx, h.conn, x, y, spellID)
}

// CastOnInventory enchants jewelry / converts items / etc.
// slotIndex is the inventory slot of the target item.
func (h *Host) CastOnInventory(ctx context.Context, slotIndex, spellID int) error {
	return action.CastOnInventory(ctx, h.conn, slotIndex, spellID)
}

// EquipItem moves an inventory item into its equipment slot.
func (h *Host) EquipItem(ctx context.Context, slotIndex int) error {
	return action.EquipItem(ctx, h.conn, slotIndex)
}

// UnequipItem returns a wielded item to the inventory.
func (h *Host) UnequipItem(ctx context.Context, slotIndex int) error {
	return action.UnequipItem(ctx, h.conn, slotIndex)
}
