package runtime

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/action"
)

// DropItem drops an inventory item from the given slot to the player's
// current tile.
func (h *Host) DropItem(ctx context.Context, slot int) error {
	return action.DropItem(ctx, h.conn, slot)
}

// PickUpItem picks up the ground item at (x, y) with the given itemID.
//
// Two-phase: walk to the item via the pathfinder, then once the bot
// is within the server's view radius, send the take packet. We can't
// send walk + take in a single burst because OpenRSC's GroundItemTake
// handler calls `player.resetPath()` when the item isn't visible from
// the player's current tile — which nukes our walk path before she
// takes a step.
func (h *Host) PickUpItem(ctx context.Context, x, y, itemID int) error {
	h.log.Info("PickUpItem: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"item_id", itemID,
	)
	// Phase 1: walk.
	if err := h.walkPathTo(ctx, x, y, false); err != nil && !errors.Is(err, ErrNoPath) {
		return err
	}
	// Phase 2: poll position until within view of the item, then take.
	// Server view radius is 15 (config view_distance * 8 - 1).
	const viewRadius = 15
	deadline := time.Now().Add(45 * time.Second)
	for {
		pos := h.world.Self.Position()
		dx, dy := absInt(pos.X-x), absInt(pos.Y-y)
		if dx <= viewRadius && dy <= viewRadius {
			h.log.Info("PickUpItem: within view, sending take",
				"player", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
				"item", fmt.Sprintf("(%d, %d)", x, y),
			)
			return action.PickUpItem(ctx, h.conn, x, y, itemID)
		}
		if time.Now().After(deadline) {
			worst := dx
			if dy > worst {
				worst = dy
			}
			return fmt.Errorf("PickUpItem: walking timeout — still at (%d, %d), %d tiles from target", pos.X, pos.Y, worst)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}
}

// ItemCommand fires the default right-click action on an inventory
// slot — Bury for bones, Eat for food, Drink for potions. No walking
// needed (the action targets a tile-local inventory slot).
func (h *Host) ItemCommand(ctx context.Context, slot int) error {
	return action.ItemCommand(ctx, h.conn, slot)
}
