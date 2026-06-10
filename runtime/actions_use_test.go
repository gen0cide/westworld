package runtime

import (
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/session"
	"github.com/gen0cide/westworld/world"
)

// TestUseAcceptsBothGroundItemShapes mirrors TestPickUpAcceptsBothGroundItemShapes
// for use(item, ground_item) — the same defect, one switch arm over: dslUse's
// target dispatch matched only the concrete *groundItemView, so the
// *groundItemsNearestValue behind world.ground_items.nearest (whose Kind() ALSO
// reads "ground_item") fell into the default branch as the absurd "unsupported
// target type ground_item". use must accept every ground-item-shaped view via the
// same groundItemish surface pick_up asserts on.
//
// Unlike pick_up, use has no perception pre-check to stop at before the wire, so
// the host is backed by a real loopback TCP conn whose far side just drains — the
// dispatched walk+use packets land harmlessly and both shapes must return Ok.
func TestUseAcceptsBothGroundItemShapes(t *testing.T) {
	h := New(Options{Username: "test"})
	h.world.Self.SetPosition(world.Coord{X: 120, Y: 504})
	h.conn = newDrainedConn(t)
	h.world.Inventory.Replace([]world.InvSlot{{ItemID: 166, Amount: 1}})

	// Ground item ONE tile away: awaitArrival's 1-tile radius passes
	// immediately, so the flow reaches the use packet without waiting.
	h.world.GroundItems.Add(121, 504, 20)

	// Shape 1: the dual-mode wrapper behind `world.ground_items.nearest`
	// (the shape the old concrete *groundItemView case rejected).
	nearest, ok := (&groundItemsView{host: h}).Get("nearest")
	if !ok {
		t.Fatal("world.ground_items.nearest not gettable")
	}
	if _, isWrapper := nearest.(*groundItemsNearestValue); !isWrapper {
		t.Fatalf("nearest resolved to %T, want *groundItemsNearestValue (test setup)", nearest)
	}

	// Shape 2: a plain row view (world.ground_items[i] / by_id / most_valuable).
	row := &groundItemView{record: world.GroundItemRecord{X: 121, Y: 504, ItemID: 20}, facts: h.facts}

	for name, arg := range map[string]interp.Value{
		"nearest_wrapper": nearest,
		"row_view":        row,
	} {
		if arg.Kind() != "ground_item" {
			t.Fatalf("%s: Kind() = %q, want ground_item (test setup)", name, arg.Kind())
		}
		v, err := dslUse(context.Background(), h, []interp.Value{interp.Int(166), arg}, nil)
		if err != nil {
			t.Fatalf("%s: use returned a Go error: %v", name, err)
		}
		cr, ok := v.(*interp.CallResult)
		if !ok {
			t.Fatalf("%s: use returned %T, want *interp.CallResult", name, v)
		}
		if cr.Err != nil {
			t.Errorf("%s: use failed: %s (%s)", name, cr.Err.Code, cr.Err.Reason)
		}
	}

	// Negative control: a non-target value is still a programmer bug.
	if _, err := dslUse(context.Background(), h, []interp.Value{interp.Int(166), interp.String("fire")}, nil); err == nil {
		t.Error("use(item, String) succeeded, want Go error")
	} else if !strings.Contains(err.Error(), "unsupported target type") {
		t.Errorf("use(item, String) error = %v, want 'unsupported target type'", err)
	}
}

// newDrainedConn dials a loopback TCP listener whose accept side discards all
// bytes, handing tests a real *session.Conn that swallows outbound packets.
func newDrainedConn(t *testing.T) *session.Conn {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		c, err := ln.Accept()
		if err != nil {
			return
		}
		_, _ = io.Copy(io.Discard, c)
	}()
	conn, err := session.Dial(context.Background(), ln.Addr().String(),
		session.Options{PreLoginReadDelay: time.Millisecond})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })
	return conn
}
