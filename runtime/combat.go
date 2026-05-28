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
