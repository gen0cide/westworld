package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Polymorphic interaction action handler bodies: interact_at /
// open_boundary / use(item, target). Registered in the central
// actionHandlers table in dsl_actions.go.

// ---------- interact_at(target, option?) — far-range scenery click ----------

// dslInteractAt fires the primary (option=1, default) or secondary
// (option=2) click on a scenery tile. Generic verb — the actual
// verb dispatched server-side depends on the scenery def's
// Command1 / Command2 fields ("Chop", "Mine", "Climb-Up", etc.).
//
// Accepts:
//
//	interact_at(x=X, y=Y)
//	interact_at(x=X, y=Y, option=2)
//	interact_at(position)              — any view with .x/.y
//	interact_at(scenery_view)          — placement from world.locs
func dslInteractAt(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("interact_at: %v", err)
	}
	option := 1
	if v, ok := named["option"]; ok {
		if i, ok := v.(interp.Int); ok {
			option = int(i)
		}
	} else if len(args) >= 2 {
		if i, ok := args[1].(interp.Int); ok {
			option = int(i)
		}
	}
	if option != 1 && option != 2 {
		return nil, errf("interact_at: option must be 1 or 2, got %d", option)
	}
	if err := h.InteractAt(ctx, x, y, option); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslOpenBoundary fires the default open/cross click on a boundary
// (door, gate, fence, web). Takes a boundary view from world.locs.
// The host pathfinds adjacent before sending — same as the existing
// walk-then-act flow.
func dslOpenBoundary(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("open_boundary takes 1 arg (boundary view), got %d", len(args))
	}
	bv, ok := args[0].(*boundaryView)
	if !ok {
		return nil, errf("open_boundary: expected boundary view, got %s", args[0].Kind())
	}
	if err := h.InteractWithBoundary(ctx, bv.x, bv.y, bv.direction); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// ---------- use(item, target) — polymorphic interaction ----------

// dslUse dispatches to the right server opcode based on the
// target's view type. Single DSL verb, multiple wire formats:
//
//	use(key, door)        — UseItemOnBoundary (opcode 161)
//	use(needle, cloth)    — UseItemOnItem      (opcode 91)
//	use(log, fire)        — UseItemOnScenery   (opcode 115)  [TODO]
//
// The item arg can be:
//   - itemViewVal (from inventory.find / inventory.slots)
//   - Int (raw item id; we look up the slot)
//   - String (item name; we look up the id then slot)
//
// Resolving the inventory slot here means routines never have to
// pass slot numbers explicitly — the bot finds the item itself.
// If the item isn't in inventory we return NO_SUCH_ITEM.
func dslUse(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("use takes 1 (item) or 2 (item, target) arguments, got %d", len(args))
	}
	itemID, err := resolveItemID(h.facts, args[0])
	if err != nil {
		return nil, errf("use: bad item arg: %v", err)
	}
	slot := -1
	for i, s := range h.world.Inventory.Slots() {
		if s.ItemID == itemID {
			slot = i
			break
		}
	}
	if slot < 0 {
		return interp.Fail(interp.NO_SUCH_ITEM,
			fmt.Sprintf("use: item id %d not in inventory", itemID)), nil
	}
	// Coordinate target: use(item, x=X, y=Y) uses the item on the scenery at that
	// tile — the bridge from the "Object @ (x,y)" the agent perceives to an
	// action (e.g. cooking: use(raw_food, x=RANGEX, y=RANGEY)). cook() is not yet
	// implemented, so this is how skilling-on-scenery is done.
	if xv, okx := named["x"]; okx {
		if yv, oky := named["y"]; oky {
			tx, _ := interp.AsInt(xv)
			ty, _ := interp.AsInt(yv)
			if err := h.UseItemOnScenery(ctx, int(tx), int(ty), slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		}
	}
	// No-target form: use(item) fires the item's own inventory command
	// (ITEM_COMMAND, opcode 90). This is how the sleeping bag (item
	// 1263) starts the fatigue/sleep flow — the server's OpInv trigger
	// runs the item's command and replies with SEND_SLEEPSCREEN. The
	// cradle then auto-answers the sleep word. (Beds are scenery: a bed
	// is used via interact_at(...) / use(item, scenery) on the bed tile,
	// whose "rest"/"sleep"/"lie in" command routes to the same
	// sendEnterSleep — see Sleeping.onOpLoc.)
	if len(args) == 1 {
		if err := h.ItemCommand(ctx, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	}
	// Target dispatch.
	switch t := args[1].(type) {
	case *boundaryView:
		if err := h.UseItemOnBoundary(ctx, t.x, t.y, t.direction, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *itemViewVal:
		// Inventory-on-inventory: find the target's slot too.
		otherSlot := -1
		for i, s := range h.world.Inventory.Slots() {
			if s.ItemID == t.ID && i != slot {
				otherSlot = i
				break
			}
		}
		if otherSlot < 0 {
			return interp.Fail(interp.NO_SUCH_ITEM,
				fmt.Sprintf("use: target item id %d not in a different inventory slot", t.ID)), nil
		}
		if err := h.UseItemOnItem(ctx, slot, otherSlot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *placementView:
		// Result of world.locs.X.nearest() — kind="scenery" /
		// "boundary" / "npc_spawn" decides the opcode. The DSL
		// caller wrote `use(item, world.locs.fires.nearest())`
		// and the dispatch shape is hidden behind the verb.
		switch t.p.Kind {
		case "scenery":
			if err := h.UseItemOnScenery(ctx, t.p.X, t.p.Y, slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		case "boundary":
			if err := h.UseItemOnBoundary(ctx, t.p.X, t.p.Y, t.p.Direction, slot); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		case "npc_spawn":
			return nil, errf("use: cannot use(item, npc_spawn) — pass the live NPC view from world.npcs, not the spawn placement")
		default:
			return nil, errf("use: unsupported placement kind %q", t.p.Kind)
		}
	case *groundItemView:
		// Inv item on a ground-item. Server needs the ground item
		// type id too (multiple stacks can pile on one tile).
		if err := h.UseItemOnGroundItem(ctx, t.record.X, t.record.Y, t.record.ItemID, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *npcView:
		// Inv item on an NPC (thieving, item-give, quest hand-in).
		if err := h.UseItemOnNpc(ctx, t.record.Index, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	case *playerView:
		// Inv item on another player (trade-init via "use" gesture,
		// gift-give). Server confirms via TradeRequestReceived event
		// on the target's side.
		if err := h.UseItemOnPlayer(ctx, t.record.Index, slot); err != nil {
			return wrapServerErr(err), nil
		}
		return interp.Ok(interp.Null{}), nil
	default:
		return nil, errf("use: unsupported target type %s — expected boundary, item, scenery, ground_item, npc, or player", args[1].Kind())
	}
}
