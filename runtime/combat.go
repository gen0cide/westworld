package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
)

// AttackNpc initiates combat with an NPC at the given server index.
// Pathfinds from the bot's current position to a tile adjacent to the
// NPC, sends the multi-corner walk packet, then the attack opcode —
// matching mudclient.walkToArea(..., walkToEntity=true) per click.
//
// Falls back to sending a bare attack packet if the NPC isn't in
// the world-state mirror yet (e.g., the bot just logged in and hasn't
// received an NpcCoords update).
func (h *Host) AttackNpc(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("AttackNpc: NPC not in world state, sending bare action", "server_index", serverIndex)
		return action.AttackNpc(ctx, h.conn, serverIndex)
	}
	h.log.Info("AttackNpc: pathfinding",
		"server_index", serverIndex,
		"npc_type_id", rec.TypeID,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.AttackNpc(ctx, h.conn, serverIndex)
	})
}

// AttackPlayer initiates PVP with a player. Pathfinds to the
// player's last-known position then sends the attack packet.
func (h *Host) AttackPlayer(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	}
	h.log.Info("AttackPlayer: pathfinding",
		"server_index", serverIndex,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	})
}

// ChooseDialogOption replies to the server's most recent
// SendNpcDialogOptions prompt by picking the option at `index`. The
// 0-based index must match the order of Options in the
// event.NpcDialog the bot received.
func (h *Host) ChooseDialogOption(ctx context.Context, index int) error {
	return action.ChooseDialogOption(ctx, h.conn, index)
}

// SetCombatStyle changes the melee xp-split. Takes effect on the
// next attack tick — RSC doesn't acknowledge the change with a
// dedicated packet, so observe self.skills.<style>.xp deltas if
// you need to confirm it was applied.
func (h *Host) SetCombatStyle(ctx context.Context, style action.CombatStyle) error {
	return action.SetCombatStyle(ctx, h.conn, style)
}

// Retreat breaks melee combat by walking one tile away — the only
// mechanic the authentic v235 protocol (and the GUI) offers for
// disengaging. Verified against OpenRSC WalkRequest.java: there is no
// dedicated "stop fighting" opcode; sending a WALK_TO_POINT while
// inCombat() triggers player.resetCombatEvent() server-side (after the
// mandatory first 3 combat rounds). So a retreat IS a walk.
//
// Direction: step directly away from the engaged target when we can
// resolve its tile (so we actually open distance); otherwise step one
// tile south as a deterministic fallback. We send a single in-FOV walk
// (Host.Walk) rather than a full BFS route — retreat is "take one step
// out of melee range now", and the server's combat-break fires on the
// first walk packet. Returns the underlying Walk error (out-of-range /
// connection) unchanged so the DSL layer can map it.
func (h *Host) Retreat(ctx context.Context) error {
	pos := h.world.Self.Position()
	dx, dy := 0, 1 // fallback: one tile south

	// Prefer stepping away from the current engagement target so we
	// open real distance. Resolve the target's tile from the same
	// sources combat.target uses (own-slot engagement, then last
	// attacked entity).
	if tx, ty, ok := h.engagedTargetTile(); ok {
		sx := sign(pos.X - tx)
		sy := sign(pos.Y - ty)
		if sx == 0 && sy == 0 {
			// Stacked on the target (shouldn't happen in melee) —
			// keep the south fallback.
		} else {
			dx, dy = sx, sy
		}
	}
	return h.Walk(ctx, pos.X+dx, pos.Y+dy)
}

// engagedTargetTile returns the (x, y) of our current combat target if
// one can be resolved, mirroring combat.target's resolution order.
func (h *Host) engagedTargetTile() (int, int, bool) {
	if self, ok := h.world.Players.Self(); ok && !self.EngagedAt.IsZero() {
		if self.EngagedNpcIndex >= 0 {
			if rec, ok := h.world.Npcs.Get(self.EngagedNpcIndex); ok {
				return rec.X, rec.Y, true
			}
		}
		if self.EngagedPlayerIndex >= 1 {
			if rec, ok := h.world.Players.Get(self.EngagedPlayerIndex); ok {
				return rec.X, rec.Y, true
			}
		}
	}
	if idx := h.lastAttackedNpcIndex; idx != 0 {
		if rec, ok := h.world.Npcs.Get(idx); ok {
			return rec.X, rec.Y, true
		}
	}
	if idx := h.lastAttackedPlayerIndex; idx != 0 {
		if rec, ok := h.world.Players.Get(idx); ok {
			return rec.X, rec.Y, true
		}
	}
	return 0, 0, false
}

// sign returns -1, 0, or +1 for the sign of n.
func sign(n int) int {
	switch {
	case n > 0:
		return 1
	case n < 0:
		return -1
	default:
		return 0
	}
}

// TalkToNpc opens dialog with an NPC. Walks adjacent (reachBorder
// mode, mirroring "click NPC across a counter") then sends the
// talk-to packet.
func (h *Host) TalkToNpc(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("TalkToNpc: NPC not in world state, sending bare action", "server_index", serverIndex)
		return action.TalkToNpc(ctx, h.conn, serverIndex)
	}
	h.log.Info("TalkToNpc: pathfinding",
		"server_index", serverIndex,
		"npc_type_id", rec.TypeID,
		"to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
	)
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.TalkToNpc(ctx, h.conn, serverIndex)
	})
}
