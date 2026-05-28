package runtime

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/action"
	"github.com/gen0cide/westworld/event"
)

// Follow sends the server-side PLAYER_FOLLOW packet (opcode 165) to
// stick adjacent to the named player. After the server accepts, IT
// handles all the walking — we just have to keep the session alive.
//
// Blocks until ctx is cancelled OR until startupTimeout expires
// while waiting for the target to first become visible.
//
// We learn the target's serverIndex from inbound OtherPlayerAppearance
// events. Once we've sent the follow packet, the server takes over;
// we periodically re-send the follow packet to handle the case where
// the target moves out of view and back into it.
func (h *Host) Follow(ctx context.Context, targetName string, startupTimeout time.Duration) error {
	if targetName == "" {
		return errors.New("Follow: empty target name")
	}
	target := strings.ToLower(targetName)
	sub := h.bus.Subscribe("*", 256)

	// Wait for the target to first become visible (so we learn their
	// serverIndex).
	waitCtx, cancel := context.WithTimeout(ctx, startupTimeout)
	defer cancel()
	var serverIndex int = -1
	for serverIndex == -1 {
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("Follow: target %q never became visible: %w", targetName, waitCtx.Err())
		case ev, ok := <-sub:
			if !ok {
				return errors.New("Follow: bus closed")
			}
			if ap, isAp := ev.(event.OtherPlayerAppearance); isAp {
				if strings.EqualFold(ap.Name, target) {
					serverIndex = ap.PlayerIndex
				}
			}
		}
	}
	h.log.Info("follow: target acquired",
		"name", targetName,
		"server_index", serverIndex,
	)

	// Send the follow request.
	if err := action.FollowPlayer(ctx, h.conn, serverIndex); err != nil {
		return fmt.Errorf("Follow: send: %w", err)
	}

	// Refresh the follow request every 10s in case the target moved
	// out of view momentarily and the server stopped following.
	refresh := time.NewTicker(10 * time.Second)
	defer refresh.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev, ok := <-sub:
			if !ok {
				return nil
			}
			// Re-acquire serverIndex from new appearance events (the
			// index can change between sessions).
			if ap, isAp := ev.(event.OtherPlayerAppearance); isAp && strings.EqualFold(ap.Name, target) {
				if ap.PlayerIndex != serverIndex {
					h.log.Info("follow: target index changed",
						"old", serverIndex,
						"new", ap.PlayerIndex,
					)
					serverIndex = ap.PlayerIndex
					_ = action.FollowPlayer(ctx, h.conn, serverIndex)
				}
			}
		case <-refresh.C:
			if err := action.FollowPlayer(ctx, h.conn, serverIndex); err != nil {
				h.log.Warn("follow: refresh failed", "err", err)
			}
		}
	}
}

// DropItem drops an inventory item from the given slot.
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
	h.log.Info("PickUpItem: pathfinding", "to", fmt.Sprintf("(%d, %d)", x, y), "item_id", itemID)
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
		dx, dy := pos.X-x, pos.Y-y
		if dx < 0 {
			dx = -dx
		}
		if dy < 0 {
			dy = -dy
		}
		if dx <= viewRadius && dy <= viewRadius {
			h.log.Info("PickUpItem: within view, sending take",
				"player", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
				"item", fmt.Sprintf("(%d, %d)", x, y),
			)
			return action.PickUpItem(ctx, h.conn, x, y, itemID)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("PickUpItem: walking timeout — still at (%d, %d), %d tiles from target", pos.X, pos.Y, max(dx, dy))
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(400 * time.Millisecond):
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ItemCommand fires the default right-click action on an inventory
// slot — Bury for bones, Eat for food, etc. No walking needed.
func (h *Host) ItemCommand(ctx context.Context, slot int) error {
	return action.ItemCommand(ctx, h.conn, slot)
}

// PrivateMessage sends a /tell-style private message to a named
// player. Recipient must be online AND have us on their friends
// list; the server silently drops the packet otherwise.
func (h *Host) PrivateMessage(ctx context.Context, recipient, message string) error {
	return action.PrivateMessage(ctx, h.conn, recipient, message)
}

// AddFriend adds a player to the bot's friends list. Required before
// PM exchange will work.
func (h *Host) AddFriend(ctx context.Context, name string) error {
	return action.AddFriend(ctx, h.conn, name)
}

// RemoveFriend removes a player from the bot's friends list.
func (h *Host) RemoveFriend(ctx context.Context, name string) error {
	return action.RemoveFriend(ctx, h.conn, name)
}

// BankDeposit moves the given catalog item into the bank. The bot
// must have already opened the bank (via TalkToNpc on a banker).
func (h *Host) BankDeposit(ctx context.Context, catalogID, amount int) error {
	return action.BankDeposit(ctx, h.conn, catalogID, amount)
}

// BankWithdraw is the dual of BankDeposit.
func (h *Host) BankWithdraw(ctx context.Context, catalogID, amount int) error {
	return action.BankWithdraw(ctx, h.conn, catalogID, amount)
}

// BankClose closes the bank window.
func (h *Host) BankClose(ctx context.Context) error {
	return action.BankClose(ctx, h.conn)
}

// DepositAll deposits every inventory item whose ItemID is in
// `keepItemIDs == false`. Useful at end of a kill run. Excludes any
// IDs in `keepIDs` (e.g., your bread + bones-to-bury).
func (h *Host) DepositAll(ctx context.Context, keepIDs map[int]struct{}) error {
	slots := h.world.Inventory.Slots()
	for i, s := range slots {
		if s.ItemID == 0 {
			continue
		}
		if _, keep := keepIDs[s.ItemID]; keep {
			continue
		}
		amount := s.Amount
		if amount <= 0 {
			amount = 1
		}
		h.log.Info("DepositAll: depositing", "slot", i, "item_id", s.ItemID, "amount", amount)
		if err := h.BankDeposit(ctx, s.ItemID, amount); err != nil {
			h.log.Warn("DepositAll: BankDeposit failed", "item_id", s.ItemID, "err", err)
			continue
		}
		// Give the server a tick to process and update the inventory
		// mirror before we read it again.
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(700 * time.Millisecond):
		}
	}
	return nil
}

// AttackNpc initiates combat with an NPC at the given server index.
// Pathfinds from current position to a tile adjacent to the NPC,
// sends the multi-corner walk packet, then the attack opcode — same
// shape mudclient.walkToArea(..., walkToEntity=true) emits per click.
//
// Falls back to a simple walk hint if no landscape archive is
// available; falls back further to just sending the action if the
// NPC isn't in world state yet.
func (h *Host) AttackNpc(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("AttackNpc: NPC not in world state, sending bare action", "server_index", serverIndex)
		return action.AttackNpc(ctx, h.conn, serverIndex)
	}
	h.log.Info("AttackNpc: pathfinding", "server_index", serverIndex, "npc_type_id", rec.TypeID, "to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y))
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.AttackNpc(ctx, h.conn, serverIndex)
	})
}

// AttackPlayer initiates PVP with a player. Pathfinds to the player's
// last-known position then sends the attack packet.
func (h *Host) AttackPlayer(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	}
	h.log.Info("AttackPlayer: pathfinding", "server_index", serverIndex, "to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y))
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.AttackPlayer(ctx, h.conn, serverIndex)
	})
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
	h.log.Info("TalkToNpc: pathfinding", "server_index", serverIndex, "npc_type_id", rec.TypeID, "to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y))
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.TalkToNpc(ctx, h.conn, serverIndex)
	})
}

// InitTradeRequest sends a trade request to the player at serverIndex.
// Trade requires standing adjacent to the target.
func (h *Host) InitTradeRequest(ctx context.Context, serverIndex int) error {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		return action.InitTradeRequest(ctx, h.conn, serverIndex)
	}
	h.log.Info("InitTradeRequest: pathfinding", "server_index", serverIndex, "to", fmt.Sprintf("(%d, %d)", rec.X, rec.Y))
	return h.walkAndAct(ctx, rec.X, rec.Y, true, walkToEntity, func(ctx context.Context) error {
		return action.InitTradeRequest(ctx, h.conn, serverIndex)
	})
}

// sendNpcWalkHint walks the bot toward an NPC's last-known position
// to overcome OpenRSC's quirk that the server-side follow event only
// emits one-tile paths (see AttackHandler / NpcTalkTo handlers: both
// call setFollowing, which per-tick calls walkToEntity that adds a
// single non-adjacent step — and the WalkingQueue silently drops any
// non-adjacent step). Without a precomputed multi-step walk, the bot
// just sits still when the target isn't already adjacent.
//
// We aim at an orthogonally adjacent tile of the NPC (N/S/E/W),
// closest to the player. NPCs occupy their own tile, so walking to
// the NPC's own location would just fail at the last step.
func (h *Host) sendNpcWalkHint(ctx context.Context, serverIndex int, caller string) {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok {
		h.log.Info("walk hint skipped: NPC not in world state",
			"caller", caller, "server_index", serverIndex,
			"hint", "increase initial dwell before issuing the action so NpcCoords events have time to populate world state",
		)
		return
	}
	pos := h.world.Self.Position()
	ax, ay := nearestOrthogonalNeighbor(pos.X, pos.Y, rec.X, rec.Y)
	if ax == pos.X && ay == pos.Y {
		h.log.Info("walk hint not needed: already adjacent to NPC",
			"caller", caller, "server_index", serverIndex,
			"npc_type_id", rec.TypeID,
		)
		return
	}
	h.log.Info("walk hint",
		"caller", caller,
		"server_index", serverIndex,
		"npc_type_id", rec.TypeID,
		"npc_at", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
		"from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"to", fmt.Sprintf("(%d, %d)", ax, ay),
	)
	if err := action.Walk(ctx, h.conn, pos.X, pos.Y, ax, ay); err != nil {
		h.log.Warn("walk hint failed", "caller", caller, "err", err)
	}
}

// sendPlayerWalkHint walks the bot to a player's last-known position.
// Same NPC-tile-occupied caveat: target an adjacent tile, not the
// player's own tile.
func (h *Host) sendPlayerWalkHint(ctx context.Context, serverIndex int, caller string) {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok || rec.LastSeen.IsZero() {
		h.log.Info("walk hint skipped: player not in world state",
			"caller", caller, "server_index", serverIndex,
		)
		return
	}
	pos := h.world.Self.Position()
	ax, ay := nearestOrthogonalNeighbor(pos.X, pos.Y, rec.X, rec.Y)
	if ax == pos.X && ay == pos.Y {
		h.log.Info("walk hint not needed: already adjacent to player",
			"caller", caller, "server_index", serverIndex,
			"player_name", rec.Name,
		)
		return
	}
	h.log.Info("walk hint",
		"caller", caller,
		"server_index", serverIndex,
		"player_name", rec.Name,
		"player_at", fmt.Sprintf("(%d, %d)", rec.X, rec.Y),
		"from", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
		"to", fmt.Sprintf("(%d, %d)", ax, ay),
	)
	if err := action.Walk(ctx, h.conn, pos.X, pos.Y, ax, ay); err != nil {
		h.log.Warn("walk hint failed", "caller", caller, "err", err)
	}
}

// nearestOrthogonalNeighbor returns the (N/S/E/W) 4-neighbor of
// (targetX, targetY) closest to (playerX, playerY) by Manhattan
// distance. If the player is already AT one of those neighbors,
// returns the player's coords (caller should treat this as "no walk
// needed").
func nearestOrthogonalNeighbor(playerX, playerY, targetX, targetY int) (int, int) {
	cands := [4][2]int{
		{targetX - 1, targetY},
		{targetX + 1, targetY},
		{targetX, targetY - 1},
		{targetX, targetY + 1},
	}
	bestI := 0
	bestD := 1 << 30
	for i, c := range cands {
		d := abs(c[0]-playerX) + abs(c[1]-playerY)
		if d < bestD {
			bestD = d
			bestI = i
		}
	}
	return cands[bestI][0], cands[bestI][1]
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// AcceptIncomingTrade accepts a pending trade request.
func (h *Host) AcceptIncomingTrade(ctx context.Context) error {
	return action.AcceptIncomingTrade(ctx, h.conn)
}

// DeclineTrade declines a pending or active trade.
func (h *Host) DeclineTrade(ctx context.Context) error {
	return action.DeclineTrade(ctx, h.conn)
}

// OfferTradeItems sets our offered items in the current trade.
func (h *Host) OfferTradeItems(ctx context.Context, items []action.TradeItem) error {
	return action.OfferTradeItems(ctx, h.conn, items)
}

// ConfirmTrade clicks Accept on the trade window.
func (h *Host) ConfirmTrade(ctx context.Context) error {
	return action.ConfirmTrade(ctx, h.conn)
}

// InteractWithBoundary fires the default click on a boundary (door,
// ladder, fence) — e.g., to open a closed door.
//
// OpenRSC's GameObjectWallAction handler only sets a one-shot
// WalkToAction; it does NOT initiate walking itself. We pathfind a
// route to a tile inside the boundary's atObject rectangle, send the
// multi-corner walk packet, then the interact packet.
func (h *Host) InteractWithBoundary(ctx context.Context, x, y, direction int) error {
	h.log.Info("InteractWithBoundary: pathfinding",
		"to", fmt.Sprintf("(%d, %d)", x, y),
		"dir", direction,
	)
	return h.walkAndAct(ctx, x, y, true, walkToPoint, func(ctx context.Context) error {
		return action.InteractWithBoundary(ctx, h.conn, x, y, direction)
	})
}
