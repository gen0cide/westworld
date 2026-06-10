package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Inventory / item action handler bodies (drop / eat / pick_up /
// equip / unequip / use_inventory_default). Registered in the central
// actionHandlers table in dsl_actions.go. Slot resolution lives in
// dsl_helpers.go (resolveSlot). The polymorphic use(item, target)
// verb lives in actions_use.go.

func dslDrop(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		// "item N not in inventory" is routine-recoverable (NO_SUCH_ITEM);
		// other resolution failures are programmer bugs (return Go err).
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.DropItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// groundItemish is the unwrap surface for "anything that presents as a
// ground item" — *groundItemView (world.ground_items rows / by_id /
// most_valuable) AND *groundItemsNearestValue (the dual-mode wrapper
// behind world.ground_items.nearest, whose Kind() also reads
// "ground_item" but whose concrete type differs). pick_up asserts on
// this interface instead of one concrete type; the old bare assertion
// produced the absurd "expected ground_item, got ground_item" error
// for nearest (soak retro 2026-06-10 #3a).
type groundItemish interface {
	interp.Value
	groundItemRecord() world.GroundItemRecord
}

// Compile-time: every ground-item-shaped view stays a valid pick_up arg.
var (
	_ groundItemish = (*groundItemView)(nil)
	_ groundItemish = (*groundItemsNearestValue)(nil)
)

// pick_up anti-storm budget (soak retro 2026-06-10 #3b): pick_up used to
// return Ok the instant the take packet was QUEUED, with no check that the
// item ever left the ground — so a loot loop whose takes were silently
// rejected (someone else's drop, full inventory) spun at interpreter speed:
// drone5 sent 150,693 take packets in one night, 73.5k in a single minute,
// for 0 items. Now every pick_up VERIFIES against the world mirror that the
// item disappeared (the wait itself is the backoff — a failed attempt costs
// pickUpVerifyWindow, not microseconds), failures are counted per item
// (tile+id), and once one item has burned pickUpMaxAttempts the verb fails
// fast with a typed error and sends NO further packets for it until the
// failure record expires.
const (
	// pickUpVerifyWindow is how long pick_up waits for the ground item to
	// vanish from the world mirror after the take packet — ~3 server ticks
	// (640ms each). Doubling as the retry backoff for loot loops.
	pickUpVerifyWindow = 2 * time.Second
	// pickUpVerifyPoll is the mirror poll interval inside the window.
	pickUpVerifyPoll = 150 * time.Millisecond
	// pickUpMaxAttempts is how many UNVERIFIED takes of the same item
	// (tile+id) are allowed before pick_up fails fast without sending.
	pickUpMaxAttempts = 3
	// pickUpFailureTTL is how long a fail-fast record lasts. After it a
	// fresh drop on the same tile (respawn, new kill) is retryable again.
	pickUpFailureTTL = 60 * time.Second
	// pickUpViewRadius mirrors the server view radius used by PickUpItem
	// (config view_distance*8 - 1): inside it the world mirror is
	// authoritative about whether a ground item exists.
	pickUpViewRadius = 15
)

// pickupFailureTracker counts consecutive unverified takes per ground-item
// key (x, y, itemID). Zero value is ready to use. Guarded by its own mutex:
// pick_up runs on the routine goroutine, but detours/REPL interpreters can
// race the main routine's.
type pickupFailureTracker struct {
	mu sync.Mutex
	m  map[[3]int]pickupFailureEntry
}

type pickupFailureEntry struct {
	n    int
	last time.Time
}

// count returns the live failure count for the key (expired entries count
// as zero).
func (t *pickupFailureTracker) count(x, y, itemID int) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	e, ok := t.m[[3]int{x, y, itemID}]
	if !ok || time.Since(e.last) > pickUpFailureTTL {
		return 0
	}
	return e.n
}

// note records one more failed take of the key and returns the new count.
// Expired entries restart from 1; stale keys are pruned opportunistically
// so the map stays bounded by the recent failure set.
func (t *pickupFailureTracker) note(x, y, itemID int) int {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.m == nil {
		t.m = map[[3]int]pickupFailureEntry{}
	}
	now := time.Now()
	for k, e := range t.m {
		if now.Sub(e.last) > pickUpFailureTTL {
			delete(t.m, k)
		}
	}
	k := [3]int{x, y, itemID}
	e := t.m[k]
	e.n++
	e.last = now
	t.m[k] = e
	return e.n
}

// clear drops the failure record for the key (called on verified success).
func (t *pickupFailureTracker) clear(x, y, itemID int) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.m, [3]int{x, y, itemID})
}

// groundItemVisible reports whether the world mirror currently holds a
// ground item with itemID at exactly (x, y).
func (h *Host) groundItemVisible(x, y, itemID int) bool {
	for _, r := range h.world.GroundItems.Near(x, y, 0) {
		if r.ItemID == itemID {
			return true
		}
	}
	return false
}

// awaitGroundItemGone polls the world mirror until the ground item at
// (x, y) vanishes (the server's pickup confirmation — it removes the item
// from the ground and the mirror applies that before publishing), the
// window elapses, or ctx is canceled. True = the item is gone.
func (h *Host) awaitGroundItemGone(ctx context.Context, x, y, itemID int, window time.Duration) bool {
	deadline := time.Now().Add(window)
	for {
		if !h.groundItemVisible(x, y, itemID) {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(pickUpVerifyPoll):
		}
	}
}

func dslPickUp(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("pick_up takes 1 argument (ground_item), got %d", len(args))
	}
	gi, ok := args[0].(groundItemish)
	if !ok {
		return nil, errf("pick_up: expected ground_item, got %s", args[0].Kind())
	}
	rec := gi.groundItemRecord()

	// Fail fast once this exact item has burned its take budget — the
	// anti-storm gate. No packet is sent; the routine sees r.err and
	// moves on (soak retro #3b).
	if n := h.pickupFailures.count(rec.X, rec.Y, rec.ItemID); n >= pickUpMaxAttempts {
		return interp.Fail(interp.SERVER_REJECTED, fmt.Sprintf(
			"pick_up: %d takes of item %d at (%d, %d) sent without the item leaving the ground — giving up on it (someone else's drop, or it can't be taken); pick a different item",
			n, rec.ItemID, rec.X, rec.Y)), nil
	}

	// The mirror is perception: when the tile is within view and the item
	// is no longer there, it was already taken or despawned — report
	// NO_SUCH_ITEM instead of sending a take for a ghost. (Beyond view
	// radius the mirror can't see the tile; let PickUpItem walk there.)
	pos := h.world.Self.Position()
	if chebyshev(pos.X, pos.Y, rec.X, rec.Y) <= pickUpViewRadius &&
		!h.groundItemVisible(rec.X, rec.Y, rec.ItemID) {
		return interp.Fail(interp.NO_SUCH_ITEM, fmt.Sprintf(
			"pick_up: no item %d on the ground at (%d, %d) any more", rec.ItemID, rec.X, rec.Y)), nil
	}

	// Snapshot the latest server-message timestamp so a rejection emitted
	// during the take window ("You can't carry any more!", ownership
	// rejects, …) can be surfaced in the failure reason.
	var preMsgAt time.Time
	if prev := h.world.Recent.ServerMessage(); prev != nil {
		preMsgAt = prev.At
	}

	if err := h.PickUpItem(ctx, rec.X, rec.Y, rec.ItemID); err != nil {
		return wrapServerErr(err), nil
	}

	// SUCCESS CHECK: the take packet going out is not the item arriving.
	// Only the item leaving the ground counts; anything else is a typed
	// failure so loot loops can't silently spin (soak retro #3b).
	if !h.awaitGroundItemGone(ctx, rec.X, rec.Y, rec.ItemID, pickUpVerifyWindow) {
		n := h.pickupFailures.note(rec.X, rec.Y, rec.ItemID)
		reason := fmt.Sprintf(
			"pick_up: take sent but item %d is still on the ground at (%d, %d) (attempt %d/%d)",
			rec.ItemID, rec.X, rec.Y, n, pickUpMaxAttempts)
		if cur := h.world.Recent.ServerMessage(); cur != nil && cur.At.After(preMsgAt) {
			reason += " — server: " + cur.Message
		}
		return interp.Fail(interp.SERVER_REJECTED, reason), nil
	}
	h.pickupFailures.clear(rec.X, rec.Y, rec.ItemID)

	// On success, return the picked-up item-view so DSL can do:
	//     got = pick_up!(item) ; say(got.name)
	picked := newItemView(h.facts, rec.ItemID, 1)
	return interp.Ok(picked), nil
}

func dslEat(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	// RSC rejects any item action while engaged in combat with "You
	// can't do that whilst you are fighting" (ItemActionHandler.java)
	// and silently drops the packet. Surface that rejection as a typed
	// EAT_IN_COMBAT result (read back from the captured server message)
	// rather than letting eat() look like a successful no-op — a routine
	// that needs to heal must retreat first. See itemCommandCombatAware.
	if msg, rejected, err := h.itemCommandCombatAware(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	} else if rejected {
		return interp.Fail(interp.EAT_IN_COMBAT, msg), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslEquip moves an inventory item into its worn slot.
// equip(item_view) or equip(slot=N).
func dslEquip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.EquipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslUnequip returns a wielded item to the inventory.
func dslUnequip(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.UnequipItem(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslUseInventoryDefault fires the option-1 / default-click action
// on an inventory item — what the RSC client sends for Bury bones,
// Clean herb, Eat food (also covered by eat()), Empty bucket, etc.
// The server's response depends on the item; the routine should
// observe via skill xp deltas or item_gained events.
func dslUseInventoryDefault(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	slot, err := resolveSlot(h, args, named)
	if err != nil {
		if strings.Contains(err.Error(), "not in inventory") {
			return interp.Fail(interp.NO_SUCH_ITEM, err.Error()), nil
		}
		return nil, err
	}
	if err := h.ItemCommand(ctx, slot); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
