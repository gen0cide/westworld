package runtime

import (
	"context"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Inventory / item action handler bodies (drop / eat / pick_up /
// equip / unequip / use_inventory_default). Registered in the central
// actionHandlers table in dsl_actions.go. Slot resolution lives in
// dsl_helpers.go (resolveSlot). The polymorphic use(item, target)
// verb lives in actions_ambient.go.

func dslDrop(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		// "item N not in inventory" is routine-recoverable (NO_SUCH_ITEM);
		// other resolution failures are programmer bugs (return Go err).
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.DropItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslPickUp(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("pick_up takes 1 argument (ground_item), got %d", len(args))
	}
	gi, ok := args[0].(*groundItemView)
	if !ok {
		return nil, errf("pick_up: expected ground_item, got %s", args[0].Kind())
	}
	if err := h.PickUpItem(ctx, gi.record.X, gi.record.Y, gi.record.ItemID); err != nil {
		return wrapServerErr(err), nil
	}
	// On success, return the picked-up item-view so DSL can do:
	//     got = pick_up!(item) ; say(got.name)
	picked := newItemView(h.facts, gi.record.ItemID, 1)
	return interp.Ok(picked), nil
}

func dslEat(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	// RSC rejects any item action while engaged in combat with "You
	// can't do that whilst you are fighting" (ItemActionHandler.java)
	// and silently drops the packet. Surface that rejection as a typed
	// EAT_IN_COMBAT result (read back from the captured server message)
	// rather than letting eat() look like a successful no-op — a routine
	// that needs to heal must retreat first. See itemCommandCombatAware.
	if msg, rejected, err := h.itemCommandCombatAware(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	} else if rejected {
		return interp.Fail(interp.EAT_IN_COMBAT, msg), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslEquip moves an inventory item into its worn slot.
// equip(item_view) or equip(slot=N).
func dslEquip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.EquipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslUnequip returns a wielded item to the inventory.
func dslUnequip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.UnequipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslUseInventoryDefault fires the option-1 / default-click action
// on an inventory item — what the RSC client sends for Bury bones,
// Clean herb, Eat food (also covered by eat()), Empty bucket, etc.
// The server's response depends on the item; the routine should
// observe via skill xp deltas or item_gained events.
func dslUseInventoryDefault(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.ItemCommand(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
