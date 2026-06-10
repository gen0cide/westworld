package runtime

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// Regression tests for the soak-retro 2026-06-10 #3 pick_up defects:
//
//	(a) "expected ground_item, got ground_item" — the handler asserted on
//	    the concrete *groundItemView, but world.ground_items.nearest hands
//	    routines a *groundItemsNearestValue whose Kind() ALSO reads
//	    "ground_item". pick_up must accept every ground-item-shaped view.
//	(b) no success check / no backoff — a failed take retried at
//	    interpreter speed (drone5: 150,693 take packets overnight, 73.5k
//	    in one minute, 0 items). After pickUpMaxAttempts unverified takes
//	    of one item, pick_up must fail fast WITHOUT sending.
//
// All tests run with a nil conn: any path that tries to send a packet
// panics, so "no packet was sent" is asserted by the tests completing.

// pickUpResult invokes dslPickUp and asserts it produced a typed
// *interp.CallResult (not a Go error — Go errors are programmer bugs).
func pickUpResult(t *testing.T, h *Host, arg interp.Value) *interp.CallResult {
	t.Helper()
	v, err := dslPickUp(context.Background(), h, []interp.Value{arg}, nil)
	if err != nil {
		t.Fatalf("pick_up returned a Go error: %v", err)
	}
	cr, ok := v.(*interp.CallResult)
	if !ok {
		t.Fatalf("pick_up returned %T, want *interp.CallResult", v)
	}
	return cr
}

// newPickUpHost builds a conn-less host standing at (120, 504).
func newPickUpHost() *Host {
	h := New(Options{Username: "test"})
	h.world.Self.SetPosition(world.Coord{X: 120, Y: 504})
	return h
}

func TestPickUpAcceptsBothGroundItemShapes(t *testing.T) {
	h := newPickUpHost()

	// Seed a ground item one tile away and capture BOTH view shapes a
	// routine can legitimately hold.
	h.world.GroundItems.Add(121, 504, 20)

	// Shape 1: the dual-mode wrapper behind `world.ground_items.nearest`
	// (the shape that used to die on the bare type assertion).
	nearest, ok := (&groundItemsView{host: h}).Get("nearest")
	if !ok {
		t.Fatal("world.ground_items.nearest not gettable")
	}
	if _, isWrapper := nearest.(*groundItemsNearestValue); !isWrapper {
		t.Fatalf("nearest resolved to %T, want *groundItemsNearestValue (test setup)", nearest)
	}

	// Shape 2: a plain row view (world.ground_items[i] / by_id / most_valuable).
	row := &groundItemView{record: world.GroundItemRecord{X: 121, Y: 504, ItemID: 20}, facts: h.facts}

	// The item vanishes before the routine acts on either view (someone
	// else took it) — exactly drone5's loot race. Both shapes must reach
	// the perception pre-check and return typed NO_SUCH_ITEM, NOT the
	// Go error "expected ground_item, got ground_item".
	h.world.GroundItems.Remove(121, 504)

	for name, arg := range map[string]interp.Value{
		"nearest_wrapper": nearest,
		"row_view":        row,
	} {
		if arg.Kind() != "ground_item" {
			t.Fatalf("%s: Kind() = %q, want ground_item (test setup)", name, arg.Kind())
		}
		cr := pickUpResult(t, h, arg)
		if cr.Err == nil {
			t.Fatalf("%s: pick_up of a vanished item reported success", name)
		}
		if cr.Err.Code != interp.NO_SUCH_ITEM {
			t.Errorf("%s: err code = %s, want NO_SUCH_ITEM (%s)", name, cr.Err.Code, cr.Err.Reason)
		}
	}

	// Negative control: a non-ground-item arg is still a programmer bug.
	if _, err := dslPickUp(context.Background(), h, []interp.Value{interp.String("bones")}, nil); err == nil {
		t.Error("pick_up(String) succeeded, want Go error")
	} else if !strings.Contains(err.Error(), "expected ground_item") {
		t.Errorf("pick_up(String) error = %v, want 'expected ground_item'", err)
	}
}

func TestPickUpBackoffCapsRetries(t *testing.T) {
	h := newPickUpHost()

	// Item visibly on the host's own tile — the storm pose: drone5 stood
	// ON Bones it could never take.
	h.world.GroundItems.Add(120, 504, 20)
	gi := &groundItemView{record: world.GroundItemRecord{X: 120, Y: 504, ItemID: 20}, facts: h.facts}

	// The item has already burned its take budget.
	for i := 0; i < pickUpMaxAttempts; i++ {
		h.pickupFailures.note(120, 504, 20)
	}

	// pick_up must now fail fast with a typed error and send NOTHING —
	// with a nil conn, reaching the send would panic the test.
	start := time.Now()
	cr := pickUpResult(t, h, gi)
	if cr.Err == nil {
		t.Fatal("pick_up reported success after exhausted take budget")
	}
	if cr.Err.Code != interp.SERVER_REJECTED {
		t.Errorf("err code = %s, want SERVER_REJECTED (%s)", cr.Err.Code, cr.Err.Reason)
	}
	if !strings.Contains(cr.Err.Reason, "giving up") {
		t.Errorf("reason %q should tell the routine to give up on this item", cr.Err.Reason)
	}
	if elapsed := time.Since(start); elapsed > pickUpVerifyWindow {
		t.Errorf("fail-fast path took %v — it must not wait out the verify window", elapsed)
	}

	// A DIFFERENT item on another tile is unaffected by the first item's
	// budget: it proceeds past the gate and the perception pre-check
	// (typed NO_SUCH_ITEM for a non-visible item — still no packet).
	other := &groundItemView{record: world.GroundItemRecord{X: 122, Y: 504, ItemID: 31}, facts: h.facts}
	cr = pickUpResult(t, h, other)
	if cr.Err == nil || cr.Err.Code != interp.NO_SUCH_ITEM {
		t.Errorf("fresh item: got %v, want NO_SUCH_ITEM (budget must be per-item)", cr.Err)
	}
}

func TestPickupFailureTrackerCountsExpireAndClear(t *testing.T) {
	var tr pickupFailureTracker

	if n := tr.count(1, 2, 3); n != 0 {
		t.Fatalf("fresh tracker count = %d, want 0", n)
	}
	tr.note(1, 2, 3)
	if n := tr.note(1, 2, 3); n != 2 {
		t.Fatalf("after two notes count = %d, want 2", n)
	}
	if n := tr.count(9, 9, 9); n != 0 {
		t.Fatalf("other key count = %d, want 0", n)
	}

	// Expired records count as zero (a fresh drop on the tile is
	// retryable) and are pruned by the next note.
	tr.m[[3]int{1, 2, 3}] = pickupFailureEntry{n: 5, last: time.Now().Add(-2 * pickUpFailureTTL)}
	if n := tr.count(1, 2, 3); n != 0 {
		t.Fatalf("expired count = %d, want 0", n)
	}
	if n := tr.note(1, 2, 3); n != 1 {
		t.Fatalf("note after expiry = %d, want restart at 1", n)
	}

	tr.clear(1, 2, 3)
	if n := tr.count(1, 2, 3); n != 0 {
		t.Fatalf("after clear count = %d, want 0", n)
	}
}

func TestAwaitGroundItemGone(t *testing.T) {
	h := newPickUpHost()

	// Nothing there: gone immediately.
	if !h.awaitGroundItemGone(context.Background(), 120, 504, 20, 0) {
		t.Error("empty tile should report gone")
	}

	// Item present and staying: the window elapses → not gone. (Window 0
	// keeps the test instant; the production window is the backoff.)
	h.world.GroundItems.Add(120, 504, 20)
	if h.awaitGroundItemGone(context.Background(), 120, 504, 20, 0) {
		t.Error("item still on the ground reported gone")
	}

	// A DIFFERENT item id on the tile is not ours.
	if !h.awaitGroundItemGone(context.Background(), 120, 504, 31, 0) {
		t.Error("tile holds item 20, not 31 — 31 should report gone")
	}
}
