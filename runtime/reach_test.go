package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

// TestScanForFiltersUnreachableComponent guards the #31 "absolute negative space"
// invariant: scan_for must never return scenery in a different walkable COMPONENT
// than the host (behind a wall / across a gap / up a dead-end ladder). Uses the real
// oracle fixture (skips without the OpenRSC archive). The assertion is component-
// equality of every returned tile to the host's component — the exact gate
// dslScanFor applies.
func TestScanForFiltersUnreachableComponent(t *testing.T) {
	h := newOracleHost(t, 0) // skips when the OpenRSC archive is unavailable
	pos := h.world.Self.Position()
	hostComp, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y)
	if !ok {
		t.Skip("host tile has no resolvable component")
	}
	for _, typ := range []string{"rock", "tree", "door", "table", "bed", "range", "altar"} {
		v, err := dslScanFor(context.Background(), h,
			[]interp.Value{interp.String(typ)},
			map[string]interp.Value{"radius": interp.Int(40)})
		if err != nil {
			t.Fatalf("scan_for(%q): %v", typ, err)
		}
		list, ok := v.(*interp.List)
		if !ok {
			t.Fatalf("scan_for(%q) did not return a list: %T", typ, v)
		}
		for _, item := range list.Items {
			pv, ok := item.(*placementView)
			if !ok {
				continue
			}
			c, _, _, ok := h.worldOracle.CompNear(pv.p.X, pv.p.Y)
			if !ok || c != hostComp {
				t.Fatalf("scan_for(%q) returned UNREACHABLE scenery at (%d,%d) comp=%d (host comp=%d) — negative-space filter failed",
					typ, pv.p.X, pv.p.Y, c, hostComp)
			}
		}
	}
}
