package worldmap

import "testing"

// TestTransportTableLoads asserts the embedded transport.json decodes and
// the load/skip accounting is populated, and that the load-bearing toll-gate
// edge is present with the correct conditional barrier + free-pass clause.
func TestTransportTableLoads(t *testing.T) {
	o, _, _ := loadReal(t)

	if o.EdgesLoaded() == 0 {
		t.Fatalf("no transport edges loaded")
	}
	t.Logf("transport: edges loaded=%d skipped=%d", o.EdgesLoaded(), o.EdgesSkipped())

	// The Al-Kharid toll gate must be present as a conditional barrier gate.
	var toll *TransportEdge
	for i := range o.Edges() {
		e := &o.Edges()[i]
		if e.Category == "toll" && e.Barrier != nil {
			toll = e
			break
		}
	}
	if toll == nil {
		t.Fatalf("toll-gate barrier edge not found in transport table")
	}
	if toll.Req.Coins != 10 {
		t.Fatalf("toll-gate Coins=%d, want 10", toll.Req.Coins)
	}
	if toll.Req.QuestFree == "" {
		t.Fatalf("toll-gate missing the Prince Ali Rescue free-pass clause")
	}
	if toll.Barrier.Axis != "x" {
		t.Fatalf("toll-gate barrier axis=%q, want x", toll.Barrier.Axis)
	}
	t.Logf("toll gate: barrier %s=%d [%d..%d] req coins=%d questFree=%q",
		toll.Barrier.Axis, toll.Barrier.Line, toll.Barrier.Lo, toll.Barrier.Hi,
		toll.Req.Coins, toll.Req.QuestFree)
}

// TestTollGateReachability is THE key regression (the whole point of the
// transport layer): the Al-Kharid toll gate is a CONDITIONAL barrier. The
// collision map walks straight across it (Lumbridge and the mine are one
// component — locked by TestTollGateCorrection), so reachability MUST add the
// game-logic gate back as a capability-gated cut:
//
//   - host at Lumbridge with 0 coins        => Al-Kharid mine NOT reachable
//   - same host with 10 coins               => Al-Kharid mine IS reachable
//   - same host with Prince Ali Rescue done => Al-Kharid mine IS reachable (free)
func TestTollGateReachability(t *testing.T) {
	o, _, _ := loadReal(t)

	const (
		lumbX, lumbY = 135, 654 // Lumbridge (east of the toll gate)
		mineX, mineY = 74, 583  // Al-Kharid mine (west of the toll gate)
	)

	// Sanity: collision-wise they ARE the same component (the correction).
	if o.CompAt(lumbX, lumbY) != o.CompAt(mineX, mineY) {
		t.Fatalf("precondition broken: Lumbridge comp %d != mine comp %d "+
			"(TestTollGateCorrection should guard this)",
			o.CompAt(lumbX, lumbY), o.CompAt(mineX, mineY))
	}

	// 0 coins: the gate is a closed conditional barrier -> mine unreachable.
	coinless := Capability{Coins: 0}
	if o.CanReach(lumbX, lumbY, mineX, mineY, coinless) {
		t.Fatalf("toll-gate regression FAILED: with 0 coins the Al-Kharid mine "+
			"(%d,%d) is reachable from Lumbridge (%d,%d) — the conditional toll "+
			"barrier is not blocking", mineX, mineY, lumbX, lumbY)
	}

	// 10 coins: the host can pay the toll -> mine reachable.
	tenCoins := Capability{Coins: 10}
	if !o.CanReach(lumbX, lumbY, mineX, mineY, tenCoins) {
		t.Fatalf("toll-gate regression FAILED: with 10 coins the Al-Kharid mine "+
			"(%d,%d) is NOT reachable from Lumbridge (%d,%d) — the toll barrier "+
			"is over-blocking a host that can pay", mineX, mineY, lumbX, lumbY)
	}

	// Prince Ali Rescue complete: free pass -> mine reachable even with 0 coins.
	questFree := Capability{Coins: 0, QuestsDone: map[string]bool{"Prince Ali Rescue": true}}
	if !o.CanReach(lumbX, lumbY, mineX, mineY, questFree) {
		t.Fatalf("toll-gate regression FAILED: with Prince Ali Rescue complete the " +
			"Al-Kharid mine should be free to reach, but it is blocked")
	}

	t.Log("toll-gate regression PASSED: 0 coins blocked; 10 coins reachable; Prince Ali Rescue free-pass reachable")
}

// TestEntranaFerryReachability is the ferry regression on a GENUINELY
// collision-isolated island: Entrana (418,570) is its own component, with no
// foot path from the mainland — only the Monk of Entrana ferry (members) gets
// there. So:
//
//   - mainland host, NOT a member => Entrana NOT reachable
//   - mainland host WITH membership => Entrana IS reachable (via the ferry)
//
// This exercises the teleport-edge half of the transport layer (the toll gate
// exercises the barrier half).
func TestEntranaFerryReachability(t *testing.T) {
	o, _, _ := loadReal(t)

	const (
		startX, startY = 265, 645 // Port Sarim docks (mainland, comp 0)
		entrX, entrY   = 418, 570 // Entrana (isolated island)
	)

	// Precondition: Entrana is collision-isolated from the start (different
	// component, so foot connectivity alone can never reach it).
	startComp, _, _, okS := o.CompNear(startX, startY)
	if !okS {
		t.Fatalf("could not snap Port Sarim start (%d,%d) to a standable tile", startX, startY)
	}
	entrComp, _, _, ok := o.CompNear(entrX, entrY)
	if !ok {
		t.Fatalf("could not snap Entrana (%d,%d) to a standable tile", entrX, entrY)
	}
	if startComp == entrComp {
		t.Skipf("Entrana is not collision-isolated in this archive (comp %d == start comp %d); "+
			"ferry-island regression needs an isolated island", entrComp, startComp)
	}
	t.Logf("Entrana comp %d, mainland start comp %d (isolated as required)", entrComp, startComp)

	// Non-member: the only route is the members-only Monk ferry -> blocked.
	nonMember := Capability{Members: false}
	if o.CanReach(startX, startY, entrX, entrY, nonMember) {
		t.Fatalf("ferry regression FAILED: a non-member reached Entrana (%d,%d) "+
			"— the members-only Monk of Entrana ferry is not gating", entrX, entrY)
	}

	// Member: the Monk ferry stands -> Entrana reachable.
	member := Capability{Members: true}
	if !o.CanReach(startX, startY, entrX, entrY, member) {
		t.Fatalf("ferry regression FAILED: a member could NOT reach Entrana (%d,%d) "+
			"via the Monk of Entrana ferry", entrX, entrY)
	}

	t.Log("Entrana ferry regression PASSED: non-member blocked; member reachable")
}

// TestReachableMonotonicity is a property check: strengthening a Capability
// can only GROW the reachable component set, never shrink it. We compare the
// coinless host to the 10-coin host from Lumbridge: the richer set must be a
// superset.
func TestReachableMonotonicity(t *testing.T) {
	o, _, _ := loadReal(t)

	const lumbX, lumbY = 135, 654

	poor := o.Reachable(lumbX, lumbY, Capability{Coins: 0})
	rich := o.Reachable(lumbX, lumbY, Capability{Coins: 1000, Members: true})

	if len(rich) < len(poor) {
		t.Fatalf("monotonicity violated: rich reachable set (%d comps) smaller than poor (%d comps)",
			len(rich), len(poor))
	}
	for c := range poor {
		if !rich[c] {
			t.Fatalf("monotonicity violated: comp %d reachable for poor host but not rich host", c)
		}
	}
	t.Logf("monotonicity OK: poor reaches %d comps, rich reaches %d comps (superset)", len(poor), len(rich))
}
