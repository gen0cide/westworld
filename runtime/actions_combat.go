package runtime

import (
	"context"
	"strings"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/dsl/interp"
)

// Combat action handler bodies (attack, set_combat_style). Registered
// in the central actionHandlers table in dsl_actions.go.

func dslAttack(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("attack takes 1 argument (target), got %d", len(args))
	}
	target := args[0]
	// Two valid shapes: an NPC view (.index) or a player view (.index).
	switch v := target.(type) {
	case *npcView:
		h.lastAttackedNpcIndex = v.record.Index
		if err := h.AttackNpc(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	case *playerView:
		h.lastAttackedPlayerIndex = v.record.Index
		if err := h.AttackPlayer(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	default:
		// Fall back: if it's an int, treat as a server index (NPC).
		if i, ok := interp.AsInt(target); ok {
			if err := h.AttackNpc(ctx, int(i)); err != nil {
				return wrapServerErr(err), nil
			}
			return interp.Ok(interp.Null{}), nil
		}
		return nil, errf("attack: target must be npc, player, or int, got %s", target.Kind())
	}
	return interp.Ok(interp.Null{}), nil
}

// dslSetCombatStyle changes the melee xp-split mode. Accepts a
// string ("controlled" / "aggressive" / "accurate" / "defensive")
// or an int (0-3 per the OpenRSC convention).
func dslSetCombatStyle(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("set_combat_style takes 1 arg (name or 0-3), got %d", len(args))
	}
	var style action.CombatStyle
	switch v := args[0].(type) {
	case interp.String:
		switch strings.ToLower(string(v)) {
		case "controlled":
			style = action.CombatStyleControlled
		case "aggressive":
			style = action.CombatStyleAggressive
		case "accurate":
			style = action.CombatStyleAccurate
		case "defensive":
			style = action.CombatStyleDefensive
		default:
			return nil, errf("set_combat_style: unknown style %q (want controlled/aggressive/accurate/defensive)", string(v))
		}
	case interp.Int:
		if v < 0 || v > 3 {
			return nil, errf("set_combat_style: int style %d out of range [0..3]", v)
		}
		style = action.CombatStyle(v)
	default:
		return nil, errf("set_combat_style: arg must be String or Int, got %s", args[0].Kind())
	}
	if err := h.SetCombatStyle(ctx, style); err != nil {
		return wrapServerErr(err), nil
	}
	// Write-through mirror for the read-side combat.style view (#117).
	// RSC sends no confirmation packet, so we record intent on success.
	h.combatStyle = style
	return interp.Ok(interp.Null{}), nil
}

// dslRetreat breaks melee combat by walking one tile away (#117). The
// authentic v235 protocol has no dedicated disengage opcode — fleeing
// is movement, and the server breaks combat on the first walk packet
// (verified against OpenRSC WalkRequest.java). Idempotent in spirit:
// walking when not in combat is harmless (just a step). Takes no args.
func dslRetreat(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, errf("retreat takes no arguments, got %d", len(args))
	}
	if err := h.Retreat(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
