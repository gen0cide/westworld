package runtime

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// CombatLoopOptions configures the kill→loot→bury reactor.
type CombatLoopOptions struct {
	// TargetTypeIDs limits attacks to NPCs whose def ID is in this
	// list (e.g., [62] for Goblins, [11] for Men). Empty = any NPC
	// the facts table marks attackable.
	TargetTypeIDs []int

	// MaxRange is the maximum distance (Chebyshev) the bot will roam
	// from its starting position to find a target. 0 = no limit.
	MaxRange int

	// LootRadius is how far we'll walk from the kill tile to pick up
	// drops. Items farther than this are ignored.
	LootRadius int

	// BuryItemIDs are inventory item IDs the bot will fire ItemCommand
	// on after each kill. Bones = 20; Big Bones = 413; etc.
	BuryItemIDs []int

	// MaxKills is the cap on how many kills before the loop returns.
	// 0 = unlimited.
	MaxKills int

	// KillTimeout is how long to wait after sending an attack before
	// giving up on the target (e.g., it ran out of view).
	KillTimeout time.Duration

	// AutoEatHPFraction enables the in-loop AutoEat watcher when > 0.
	// HP below this fraction of max triggers eating any food item in
	// inventory. 0 disables auto-eat. 0.4 is a safe default for
	// melee against low-level mobs.
	AutoEatHPFraction float64
}

// DefaultCombatLoopOptions returns reasonable defaults for fighting
// Goblins around Lumbridge.
func DefaultCombatLoopOptions() CombatLoopOptions {
	return CombatLoopOptions{
		TargetTypeIDs:     []int{62}, // Goblin
		MaxRange:          30,
		LootRadius:        3,
		BuryItemIDs:       []int{20, 413, 604}, // Bones, Big Bones, Bat bones
		MaxKills:          10,
		KillTimeout:       30 * time.Second,
		AutoEatHPFraction: 0.4,
	}
}

// CombatLoop is the kill→loot→bury reactor. Runs in the caller's
// goroutine and returns when ctx is cancelled, MaxKills is reached,
// or we die.
//
// Behaviour each iteration:
//  1. Find nearest attackable NPC matching TargetTypeIDs within MaxRange.
//  2. Pathfind to it and AttackNpc.
//  3. Watch for the target to disappear from world state OR for ground
//     items to appear nearby (kill signal). Bail on KillTimeout.
//  4. Walk over each fresh GroundItem within LootRadius and PickUpItem.
//  5. For each BuryItemIDs ID found in inventory, send ItemCommand.
//  6. Loop.
func (h *Host) CombatLoop(ctx context.Context, opts CombatLoopOptions) error {
	if h.facts == nil {
		return fmt.Errorf("CombatLoop: facts not loaded")
	}
	sub := h.bus.Subscribe("*", 256)
	startPos := h.world.Self.Position()
	kills := 0

	// Optional: spin up the AutoEat watcher concurrently. It runs
	// until the combat loop exits.
	if opts.AutoEatHPFraction > 0 {
		stop := h.AutoEat(ctx, opts.AutoEatHPFraction)
		defer stop()
	}

	died := false
	go func() {
		// Watch for our death so the main loop can abort.
		for ev := range sub {
			if _, ok := ev.(event.Death); ok {
				died = true
				return
			}
		}
	}()

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if died {
			return fmt.Errorf("CombatLoop: bot died after %d kills", kills)
		}
		if opts.MaxKills > 0 && kills >= opts.MaxKills {
			h.log.Info("CombatLoop: hit MaxKills", "kills", kills)
			return nil
		}

		target := h.findCombatTarget(opts, startPos)
		if target == nil {
			h.log.Info("CombatLoop: no target in range, waiting")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(2 * time.Second):
			}
			continue
		}

		h.log.Info("CombatLoop: engaging",
			"target_index", target.Index,
			"target_type_id", target.TypeID,
			"at", fmt.Sprintf("(%d, %d)", target.X, target.Y),
		)
		if err := h.AttackNpc(ctx, target.Index); err != nil {
			h.log.Warn("CombatLoop: AttackNpc failed", "err", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Wait for kill signal: a fresh ground item appears near where
		// the target was (loot drop) — or timeout.
		killTile, killed := h.waitForKillDrop(ctx, target, opts.KillTimeout)
		if !killed {
			h.log.Warn("CombatLoop: no kill signal within timeout — target may have fled or we missed", "target", target.Index)
			continue
		}
		kills++
		h.log.Info("CombatLoop: kill detected", "kills", kills, "drop_at", fmt.Sprintf("(%d, %d)", killTile.X, killTile.Y))

		// Loot every ground item within LootRadius of the kill tile.
		drops := h.findNearbyDrops(killTile, opts.LootRadius)
		h.log.Info("CombatLoop: looting", "drops", len(drops))
		for _, d := range drops {
			if err := h.PickUpItem(ctx, d.X, d.Y, d.ItemID); err != nil {
				h.log.Warn("CombatLoop: pickup failed", "item_id", d.ItemID, "err", err)
				continue
			}
			// Give the server a tick so the slot lands before we move on.
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(800 * time.Millisecond):
			}
		}

		// Bury any bury-flagged inventory slot.
		h.buryInventory(ctx, opts.BuryItemIDs)
	}
}

// findCombatTarget picks the closest attackable NPC matching the
// opts' type filter and within MaxRange of startPos.
func (h *Host) findCombatTarget(opts CombatLoopOptions, startPos world.Coord) *world.NpcRecord {
	pos := h.world.Self.Position()
	var best *world.NpcRecord
	var bestDist int
	wantTypes := map[int]struct{}{}
	for _, t := range opts.TargetTypeIDs {
		wantTypes[t] = struct{}{}
	}
	for _, rec := range h.world.Npcs.All() {
		r := rec
		if len(wantTypes) > 0 {
			if _, ok := wantTypes[r.TypeID]; !ok {
				continue
			}
		}
		def := h.facts.NpcDef(r.TypeID)
		if def == nil || !def.Attackable {
			continue
		}
		if opts.MaxRange > 0 {
			if absInt(r.X-startPos.X) > opts.MaxRange || absInt(r.Y-startPos.Y) > opts.MaxRange {
				continue
			}
		}
		dist := absInt(r.X-pos.X) + absInt(r.Y-pos.Y)
		if best == nil || dist < bestDist {
			best = &r
			bestDist = dist
		}
	}
	return best
}

// waitForKillDrop blocks until a fresh ground item appears within 3
// tiles of the player (the canonical kill-drop signal in RSC — the
// bot has walked adjacent to the target by the time it dies) or
// timeout. Returns the tile the drop landed on.
func (h *Host) waitForKillDrop(ctx context.Context, target *world.NpcRecord, timeout time.Duration) (world.Coord, bool) {
	sub := h.bus.Subscribe("ground_item", 64)

	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	for {
		select {
		case <-ctx.Done():
			return world.Coord{}, false
		case <-deadline.C:
			return world.Coord{}, false
		case ev, ok := <-sub:
			if !ok {
				return world.Coord{}, false
			}
			gi, isGI := ev.(event.GroundItemEvent)
			if !isGI || gi.Disappear {
				continue
			}
			if absInt(gi.OffsetX) > 3 || absInt(gi.OffsetY) > 3 {
				continue
			}
			// Re-read position now (not at subscribe time) so the
			// recorded drop tile reflects where the bot actually is.
			pos := h.world.Self.Position()
			itemX := pos.X + gi.OffsetX
			itemY := pos.Y + gi.OffsetY
			h.log.Info("CombatLoop: kill drop",
				"item_id", gi.ItemID,
				"at", fmt.Sprintf("(%d, %d)", itemX, itemY),
				"player_at", fmt.Sprintf("(%d, %d)", pos.X, pos.Y),
			)
			return world.Coord{X: itemX, Y: itemY}, true
		}
	}
}

// drop describes a ground item near the kill we'd like to loot.
type drop struct {
	X, Y   int
	ItemID int
}

// findNearbyDrops returns ground items currently visible within
// `radius` tiles of `center`. Reads the persistent
// world.GroundItems mirror so we see items that were broadcast
// before this function was called.
func (h *Host) findNearbyDrops(center world.Coord, radius int) []drop {
	recs := h.world.GroundItems.Near(center.X, center.Y, radius)
	out := make([]drop, 0, len(recs))
	for _, r := range recs {
		out = append(out, drop{X: r.X, Y: r.Y, ItemID: r.ItemID})
	}
	return out
}

// buryInventory iterates the bot's inventory and fires ItemCommand on
// any slot whose ItemID is in buryItemIDs.
func (h *Host) buryInventory(ctx context.Context, buryItemIDs []int) {
	if len(buryItemIDs) == 0 {
		return
	}
	wantIDs := map[int]struct{}{}
	for _, id := range buryItemIDs {
		wantIDs[id] = struct{}{}
	}
	// Inventory.Items copies the current slot snapshot; we iterate
	// repeatedly because firing one ItemCommand shifts subsequent
	// slots when the item is consumed.
	for attempts := 0; attempts < 30; attempts++ {
		slot := -1
		var item world.InvSlot
		slots := h.world.Inventory.Slots()
		for i, s := range slots {
			if s.ItemID == 0 {
				continue
			}
			if _, ok := wantIDs[s.ItemID]; ok {
				slot = i
				item = s
				break
			}
		}
		if slot < 0 {
			return // nothing more to bury
		}
		name := ""
		if d := h.facts.ItemDef(item.ItemID); d != nil {
			name = d.Name
		}
		h.log.Info("CombatLoop: burying", "slot", slot, "item_id", item.ItemID, "name", name)
		if err := h.ItemCommand(ctx, slot); err != nil {
			h.log.Warn("CombatLoop: ItemCommand failed", "slot", slot, "err", err)
			return
		}
		// Wait for the inventory update to roll back through events,
		// otherwise we read the stale slot table next iteration.
		select {
		case <-ctx.Done():
			return
		case <-time.After(800 * time.Millisecond):
		}
	}
}

// absInt and clamping live in action package; small local helper here
// avoids a cross-package dep just for utility math.
func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// describeCombatTargets renders a short, human-readable summary of
// the targets the loop would consider. Useful for cradle's --status
// flag once we add one.
func (h *Host) describeCombatTargets(opts CombatLoopOptions) string {
	parts := []string{}
	for _, t := range opts.TargetTypeIDs {
		if def := h.facts.NpcDef(t); def != nil {
			parts = append(parts, def.Name)
		}
	}
	if len(parts) == 0 {
		return "any attackable NPC"
	}
	return strings.Join(parts, ", ")
}
