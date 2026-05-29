package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Magic action handler bodies. The frozen surface exposes a single
// polymorphic magic.cast(spell, target?) (dslMagicCast below); the
// per-target cast_on_* bodies are kept as the backing implementations
// it dispatches to (and as the `cast` §9 alias). View dispatch for
// magic.cast lives in views_magic.go; the flat `cast` alias is
// registered in dsl_actions.go. Spell-id resolution lives in
// dsl_helpers.go (resolveSpellID).

// dslMagicCast is the unified, polymorphic cast. Per api.md §8 the
// target shape selects the opcode:
//   magic.cast(spell)                  -> self-targeted (CastOnSelf)
//   magic.cast(spell, npc)             -> on-NPC combat cast
//   magic.cast(spell, player)          -> PvP cast
//   magic.cast(spell, {x,y}/[x,y])     -> tile-targeted AOE
//   magic.cast(item, spell=spell_id)   -> inventory-item cast (alch/enchant)
//
// It delegates to the existing cast_on_* bodies (kept as backing) so
// there is one obvious way at the surface and zero behavioral drift.
func dslMagicCast(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	// Inventory-item cast form: spell supplied via named `spell`/
	// `spell_id`, the positional arg is the item to convert.
	if _, hasSpell := named["spell"]; hasSpell {
		return dslCastOnInventory(ctx, h, args, named)
	}
	if _, hasSpell := named["spell_id"]; hasSpell {
		return dslCastOnInventory(ctx, h, args, named)
	}
	if len(args) < 1 || len(args) > 2 {
		return nil, errf("magic.cast takes (spell) or (spell, target), got %d args", len(args))
	}
	spellArg := args[0]
	if len(args) == 1 {
		// Self-targeted cast (heal, teleport, buff).
		return dslCastOnSelf(ctx, h, []interp.Value{spellArg}, nil)
	}
	target := args[1]
	switch t := target.(type) {
	case interp.Null:
		return dslCastOnSelf(ctx, h, []interp.Value{spellArg}, nil)
	case *npcView:
		return dslCastOnNpc(ctx, h, []interp.Value{spellArg, t}, nil)
	case *playerView:
		return dslCastOnPlayer(ctx, h, []interp.Value{spellArg, t}, nil)
	case *itemViewVal:
		// Item passed positionally as target → inventory-item cast,
		// spell is the first positional.
		return dslCastOnInventory(ctx, h, []interp.Value{t}, map[string]interp.Value{"spell": spellArg})
	case *interp.List:
		// [x, y] tile-target.
		if len(t.Items) != 2 {
			return nil, errf("magic.cast: list target must be [x, y]")
		}
		x, xok := interp.AsInt(t.Items[0])
		y, yok := interp.AsInt(t.Items[1])
		if !xok || !yok {
			return nil, errf("magic.cast: list target [x, y] must be Ints")
		}
		sp, err := resolveSpellID(spellArg)
		if err != nil {
			return nil, errf("magic.cast: %v", err)
		}
		return dslCastOnLand(ctx, h, []interp.Value{interp.Int(x), interp.Int(y), interp.Int(int64(sp))}, nil)
	default:
		// A Getter with .x/.y (Position view) → tile-target.
		if g, ok := target.(interp.Getter); ok {
			xv, hasX := g.Get("x")
			yv, hasY := g.Get("y")
			if hasX && hasY {
				x, xok := interp.AsInt(xv)
				y, yok := interp.AsInt(yv)
				if xok && yok {
					sp, err := resolveSpellID(spellArg)
					if err != nil {
						return nil, errf("magic.cast: %v", err)
					}
					return dslCastOnLand(ctx, h, []interp.Value{interp.Int(x), interp.Int(y), interp.Int(int64(sp))}, nil)
				}
			}
		}
		return nil, errf("magic.cast: unsupported target type %s", target.Kind())
	}
}

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
