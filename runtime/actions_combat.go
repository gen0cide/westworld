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
		h.beginCombatRoundTracking(v.record.Index)
		if err := h.AttackNpc(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	case *playerView:
		h.lastAttackedPlayerIndex = v.record.Index
		h.beginCombatRoundTracking(v.record.Index)
		if err := h.AttackPlayer(ctx, v.record.Index); err != nil {
			return wrapServerErr(err), nil
		}
	default:
		// Fall back: if it's an int, treat as a server index (NPC).
		if i, ok := interp.AsInt(target); ok {
			h.beginCombatRoundTracking(int(i))
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

// dslRetreat breaks melee combat by walking one tile away (#117,
// hardened in #r3-retreat). The authentic v235 protocol has no
// dedicated disengage opcode — fleeing is a WALK_TO_POINT, and the
// server breaks combat on that packet (verified against OpenRSC
// WalkRequest.java). But RSC gates it: you cannot retreat until the
// opponent has made >= 3 hits ("the first 3 rounds of combat"). An
// early WALK_TO_POINT is REJECTED and the server emits "You can't
// retreat during the first 3 rounds of combat".
//
// This handler surfaces that rejection as a typed RETREAT_TOO_EARLY
// result (read back from the captured server message) so a routine can
// wait the rounds and retry rather than silently believing it fled.
// When the engine can tell fewer than 3 rounds have elapsed it waits
// for them first (Host.Retreat), so the common path doesn't even
// trigger the rejection. Idempotent in spirit: walking when not in
// combat is harmless (just a step).
//
// Named arg `wait_rounds` (default true): when true the verb waits out
// the 3-round gate before sending; pass `wait_rounds=false` to attempt
// immediately and get the typed RETREAT_TOO_EARLY rejection back so a
// routine can poll-and-branch itself. (Named `wait_rounds`, not `wait`,
// because `wait` is a reserved DSL keyword.)
func dslRetreat(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if len(args) != 0 {
		return nil, errf("retreat takes no positional arguments, got %d", len(args))
	}
	waitGate, err := boolNamed("retreat", named, "wait_rounds", true)
	if err != nil {
		return nil, err
	}
	if err := h.Retreat(ctx, waitGate); err != nil {
		return mapRetreatErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslRetreatTo flees to a specific safe tile once retreat is allowed
// (#r3-retreat). It is effectively walk_to(x, y) but routed as a
// combat-aware retreat: it shares the 3-round anti-kite gate and the
// typed RETREAT_TOO_EARLY rejection with retreat(), then (on success)
// pathfinds the rest of the way to the target tile. Use it to break
// off toward a bank/altar/lobby rather than just one tile back.
func dslRetreatTo(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("retreat_to: %v", err)
	}
	waitGate, err := boolNamed("retreat_to", named, "wait_rounds", true)
	if err != nil {
		return nil, err
	}
	if err := h.RetreatTo(ctx, x, y, waitGate); err != nil {
		return mapRetreatErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// mapRetreatErr classifies a Host.Retreat / RetreatTo error into a
// typed CallResult. The anti-kite rejection (RetreatTooEarlyError) maps
// to RETREAT_TOO_EARLY so a routine can branch and wait; everything
// else falls through to the generic server-error classifier.
func mapRetreatErr(err error) interp.Value {
	if rte, ok := err.(*RetreatTooEarlyError); ok {
		return interp.Fail(interp.RETREAT_TOO_EARLY, rte.Error())
	}
	return wrapServerErr(err)
}
