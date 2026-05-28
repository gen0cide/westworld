package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/event"
)

// AutoEat runs in the background watching HP and firing ItemCommand
// on a food slot when HP drops below `hpThreshold` (as a fraction
// 0-1 of max HP). Returns a cancel function the caller invokes to
// stop the watcher.
//
// Identifies food by checking each inventory slot's ItemDef.Command:
// any non-empty command starting with "eat" (case-insensitive) is
// considered food (Bread, Trout, Cake, etc.).
//
// Cooldown between eats prevents spamming when HP is below threshold
// across multiple ticks (one bite gives back ~3 HP).
func (h *Host) AutoEat(ctx context.Context, hpThreshold float64) func() {
	if h.facts == nil {
		h.log.Warn("AutoEat: facts unavailable, watcher disabled")
		return func() {}
	}
	watcherCtx, cancel := context.WithCancel(ctx)
	var stopOnce sync.Once
	go h.autoEatLoop(watcherCtx, hpThreshold)
	return func() { stopOnce.Do(cancel) }
}

func (h *Host) autoEatLoop(ctx context.Context, hpThreshold float64) {
	sub := h.bus.Subscribe("stat_update", 32)
	// Also re-check on a slow tick in case StatUpdate doesn't fire
	// (e.g., HP regen).
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	lastEat := time.Time{}
	const eatCooldown = 1500 * time.Millisecond

	check := func() {
		if time.Since(lastEat) < eatCooldown {
			return
		}
		cur := h.world.Self.HP()
		max := h.world.Self.MaxHP()
		if max <= 0 {
			return
		}
		if float64(cur)/float64(max) > hpThreshold {
			return
		}
		// Find a food slot.
		slot, item := h.findFoodSlot()
		if slot < 0 {
			h.log.Info("AutoEat: HP low but no food in inventory",
				"hp", fmt.Sprintf("%d/%d", cur, max),
			)
			return
		}
		h.log.Info("AutoEat: eating",
			"hp", fmt.Sprintf("%d/%d", cur, max),
			"slot", slot,
			"item_id", item,
		)
		if err := h.ItemCommand(ctx, slot); err != nil {
			h.log.Warn("AutoEat: ItemCommand failed", "err", err)
			return
		}
		lastEat = time.Now()
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			check()
		case ev, ok := <-sub:
			if !ok {
				return
			}
			s, isStat := ev.(event.StatUpdate)
			if !isStat || s.Skill != event.SkillHits {
				continue
			}
			check()
		}
	}
}

// findFoodSlot returns the first inventory slot containing an item
// with an "eat"-flavored command, plus its item ID. Returns slot=-1
// if no food is carried.
func (h *Host) findFoodSlot() (int, int) {
	slots := h.world.Inventory.Slots()
	for i, s := range slots {
		if s.ItemID == 0 {
			continue
		}
		def := h.facts.ItemDef(s.ItemID)
		if def == nil {
			continue
		}
		cmd := strings.ToLower(strings.TrimSpace(def.Command))
		if cmd == "" {
			continue
		}
		if strings.HasPrefix(cmd, "eat") {
			return i, s.ItemID
		}
	}
	return -1, 0
}
