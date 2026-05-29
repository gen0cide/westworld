package runtime

import (
	"context"
	"fmt"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/facts"
)

// InteractWithBoundary fires the default click on a boundary (door,
// ladder, fence) — e.g., to open a closed door.
//
// OpenRSC's GameObjectWallAction handler only sets a one-shot
// WalkToAction; it does NOT initiate walking itself. We pathfind a
// route to a tile inside the boundary's atObject rectangle, send the
// multi-corner walk packet (WALK_TO_POINT, since the target is an
// object not a mob), then the interact packet.
func (h *Host) InteractWithBoundary(ctx context.Context, x, y, direction int) error {
	h.log.Info("InteractWithBoundary: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.InteractWithBoundary(ctx, h.conn, x, y, direction)
	})
}

// UseItemOnBoundary uses an inventory slot on a boundary tile.
// Used for key-on-door, knife-on-vine, etc. The boundary's
// (x, y, direction) identifies the edge to interact with.
// Pathfinds to a tile adjacent to the boundary, then sends the
// use packet — same shape as InteractWithBoundary.
func (h *Host) UseItemOnBoundary(ctx context.Context, x, y, direction, slot int) error {
	h.log.Info("UseItemOnBoundary: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
		"slot", slot,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.UseItemOnBoundary(ctx, h.conn, x, y, direction, slot)
	})
}

// UseItemOnItem combines two inventory items (e.g. needle on
// cloth, chisel on gem). No pathfinding — both items are in the
// host's inventory, fire the packet immediately.
func (h *Host) UseItemOnItem(ctx context.Context, slot1, slot2 int) error {
	h.log.Info("UseItemOnItem", "slot1", slot1, "slot2", slot2)
	return action.UseItemOnItem(ctx, h.conn, slot1, slot2)
}

// UseItemOnScenery uses an inventory slot on a world object
// (cook on fire, smelt on furnace, fish at spot). Pathfinds to
// a tile adjacent to the scenery, then sends the use packet.
func (h *Host) UseItemOnScenery(ctx context.Context, x, y, slot int) error {
	h.log.Info("UseItemOnScenery: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"slot", slot,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.UseItemOnScenery(ctx, h.conn, x, y, slot)
	})
}

// InteractAt fires the primary or secondary click on a scenery
// tile (opcode 136 / 79). Pathfinds to a tile adjacent to (x, y),
// then sends the command. option=1 = primary ("Chop", "Mine",
// "Search"), option=2 = secondary if the def has one. The exact
// verb comes from SceneryDef.Command1/Command2.
func (h *Host) InteractAt(ctx context.Context, x, y, option int) error {
	h.log.Info("InteractAt: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"option", option,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.ObjectCommand(ctx, h.conn, x, y, option)
	})
}

// UseItemOnGroundItem fires opcode 53. Walks adjacent to the
// ground tile, then sends the use packet. groundItemID identifies
// which item-type to use on (multiple items can pile on one tile).
func (h *Host) UseItemOnGroundItem(ctx context.Context, x, y, groundItemID, slot int) error {
	h.log.Info("UseItemOnGroundItem: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"ground_item", groundItemID,
		"slot", slot,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.UseItemOnGroundItem(ctx, h.conn, x, y, groundItemID, slot)
	})
}

// UseItemOnNpc walks adjacent to the NPC's current position then
// fires the use-on-npc packet. The NPC's local server-index
// identifies which one (resolved from the npcView passed by the
// routine — index is stable for the NPC's lifetime in our view).
func (h *Host) UseItemOnNpc(ctx context.Context, npcServerIndex, slot int) error {
	pos := h.npcPos(npcServerIndex)
	h.log.Info("UseItemOnNpc: pathfinding",
		"npc_index", npcServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"slot", slot,
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.UseItemOnNpc(ctx, h.conn, npcServerIndex, slot)
	})
}

// NpcCommand walks adjacent to the NPC then fires its primary action
// command (opcode NPC_COMMAND → server command1). For thievable NPCs
// command1 is "pickpocket"; for others it's whatever the NpcDef lists.
// Distinct from talk_to and attack. The skill action (e.g. pickpocket
// loot/xp) repeats per click — call in a loop for multiple attempts.
func (h *Host) NpcCommand(ctx context.Context, npcServerIndex int) error {
	pos := h.npcPos(npcServerIndex)
	h.log.Info("NpcCommand: pathfinding",
		"npc_index", npcServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.NpcCommand(ctx, h.conn, npcServerIndex)
	})
}

// UseItemOnPlayer is the same shape as UseItemOnNpc but targets
// another player (trade-init, gift). Walks adjacent to the
// player's current position then fires opcode 113.
func (h *Host) UseItemOnPlayer(ctx context.Context, playerServerIndex, slot int) error {
	pos := h.playerPos(playerServerIndex)
	h.log.Info("UseItemOnPlayer: pathfinding",
		"player_index", playerServerIndex,
		"to", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"slot", slot,
	)
	return h.walkAndAct(ctx, pos.X, pos.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.UseItemOnPlayer(ctx, h.conn, playerServerIndex, slot)
	})
}

// npcPos / playerPos look up the current position of a tracked
// entity by server-index. Used by Use*OnNpc / Use*OnPlayer to
// pathfind to the entity before firing the packet. Returns zero
// Coord if not found (caller's walkAndAct surfaces the failure
// via the pathfind error path).
type coordXY struct{ X, Y int }

func (h *Host) npcPos(index int) coordXY {
	for _, n := range h.world.Npcs.All() {
		if n.Index == index {
			return coordXY{X: n.X, Y: n.Y}
		}
	}
	return coordXY{}
}

func (h *Host) playerPos(index int) coordXY {
	if rec, ok := h.world.Players.Get(index); ok {
		return coordXY{X: rec.X, Y: rec.Y}
	}
	return coordXY{}
}

// findOpenableNear looks for an openable boundary (door / doorframe;
// i.e. BoundaryDef.Unknown == 1) at or directly adjacent (Chebyshev
// distance 1) to (x, y). Returns the first match, or nil if none
// found. Used by WalkTo's open_doors=true path: when a walk stalls,
// we infer "the next tile is gated by a door" by looking up the
// boundary the server's collision check would have blocked us at.
//
// We don't try to figure out which side of the player the door is
// on — the InteractWithBoundary packet carries the boundary's own
// (x, y, direction), and the server handles the geometry. Returning
// any openable boundary in immediate range is enough.
func (h *Host) findOpenableNear(x, y int) *facts.BoundaryLoc {
	if h.facts == nil {
		return nil
	}
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			for _, p := range h.facts.At(x+dx, y+dy) {
				if p.Kind != "boundary" {
					continue
				}
				def := h.facts.BoundaryDef(p.DefID)
				if def == nil {
					continue
				}
				// Unknown=1 marks openable boundaries (doors,
				// doorframes). Plain walls/fences have Unknown=0
				// and BlocksMovement=true.
				if def.Unknown != 1 {
					continue
				}
				return &facts.BoundaryLoc{
					DefID:     p.DefID,
					X:         p.X,
					Y:         p.Y,
					Direction: p.Direction,
				}
			}
		}
	}
	return nil
}
