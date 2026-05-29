package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Magic action handler bodies (cast_on_self / _npc / _player / _land /
// _item). Registered in the central actionHandlers table in
// dsl_actions.go. Spell-id resolution lives in dsl_helpers.go
// (resolveSpellID).

func dslCastOnSelf(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("cast_on_self takes 1 arg (spell name or id)")
	}
	sp, err := resolveSpellID(args[0])
	if err != nil {
		return nil, errf("cast_on_self: %v", err)
	}
	if err := h.CastOnSelf(ctx, sp); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnNpc(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("cast_on_npc takes 2 args (spell, npc) or (npc, spell)")
	}
	// Accept either order: routines might write cast_on_npc("Wind Strike", rat)
	// or cast_on_npc(rat, 1). Try to identify the npc arg by type.
	var npcArg, spellArg interp.Value
	if _, ok := args[0].(*npcView); ok {
		npcArg, spellArg = args[0], args[1]
	} else if _, ok := args[1].(*npcView); ok {
		spellArg, npcArg = args[0], args[1]
	} else {
		// Fallback to positional: (npc, spell)
		npcArg, spellArg = args[0], args[1]
	}
	var idx int
	switch v := npcArg.(type) {
	case *npcView:
		idx = v.record.Index
	default:
		if i, ok := interp.AsInt(npcArg); ok {
			idx = int(i)
		} else {
			return nil, errf("cast_on_npc: npc arg must be npc view or Int index, got %s", npcArg.Kind())
		}
	}
	sp, err := resolveSpellID(spellArg)
	if err != nil {
		return nil, errf("cast_on_npc: %v", err)
	}
	if err := h.CastOnNpc(ctx, idx, sp); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnPlayer(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 2 {
		return nil, errf("cast_on_player takes 2 args (spell, player) or (player, spell)")
	}
	// Try to identify which arg is the player view.
	var plArg, spellArg interp.Value
	if _, ok := args[0].(*playerView); ok {
		plArg, spellArg = args[0], args[1]
	} else if _, ok := args[1].(*playerView); ok {
		spellArg, plArg = args[0], args[1]
	} else {
		plArg, spellArg = args[0], args[1]
	}
	idx, err := resolvePlayerIndex(h, plArg)
	if err != nil {
		return nil, err
	}
	sp, err := resolveSpellID(spellArg)
	if err != nil {
		return nil, errf("cast_on_player: %v", err)
	}
	if err := h.CastOnPlayer(ctx, idx, sp); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnLand(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 3 {
		return nil, errf("cast_on_land takes 3 args (x, y, spell_id)")
	}
	x, _ := interp.AsInt(args[0])
	y, _ := interp.AsInt(args[1])
	sp, _ := interp.AsInt(args[2])
	if err := h.CastOnLand(ctx, int(x), int(y), int(sp)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslCastOnInventory(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("cast_on_item takes (item, spell) or (spell, item)")
	}
	// Accept (item, spell) or (spell, item); identify by type.
	var itemArg, spellArg interp.Value
	if len(args) == 2 {
		switch args[0].(type) {
		case interp.String:
			// Ambiguous (could be spell name or item name); prefer
			// the (spell, item) reading when arg1 is item-shaped.
			if _, ok := args[1].(*itemViewVal); ok {
				spellArg, itemArg = args[0], args[1]
			} else {
				itemArg, spellArg = args[0], args[1]
			}
		default:
			itemArg, spellArg = args[0], args[1]
		}
	} else {
		itemArg = args[0]
	}
	slot, err := resolveSlot(h, []interp.Value{itemArg}, named)
	if err != nil {
		return nil, err
	}
	var sp int
	if spellArg != nil {
		id, err := resolveSpellID(spellArg)
		if err != nil {
			return nil, errf("cast_on_item: %v", err)
		}
		sp = id
	}
	if v, ok := named["spell_id"]; ok {
		if id, err := resolveSpellID(v); err == nil {
			sp = id
		}
	}
	if v, ok := named["spell"]; ok {
		if id, err := resolveSpellID(v); err == nil {
			sp = id
		}
	}
	if err := h.CastOnInventory(ctx, slot, sp); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
