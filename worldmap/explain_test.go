package worldmap

import "testing"

// TestExplainReachTollGate is the reach-EXPLANATION regression mirroring
// TestTollGateReachability but on the cognition-first surface: ExplainReach
// must not only say whether the Al-Kharid mine is reachable, but WHY — naming
// the toll gate, its 10-coin requirement, and what the host has:
//
//   - 0 coins        => Blocked(Toll gate, 10 coins, you_have=0, payable=false)
//   - 15 coins       => Gated  (Toll gate, 10 coins, you_have=15, payable=true)
//   - Prince Ali done=> Open / Gated free-pass (reachable for free)
func TestExplainReachTollGate(t *testing.T) {
	o, _, _ := loadReal(t)

	const (
		lumbX, lumbY = 135, 654 // Lumbridge (east of the toll gate)
		mineX, mineY = 74, 583  // Al-Kharid mine (west of the toll gate)
	)

	// Coinless: blocked, and the verdict must NAME the toll gate + requirement.
	coinless := o.ExplainReach(lumbX, lumbY, mineX, mineY, Capability{Coins: 0})
	if !coinless.SnapOK() {
		t.Fatalf("coinless: mine did not snap to a standable tile")
	}
	if coinless.Reach != ReachBlocked {
		t.Fatalf("coinless: reach=%q, want blocked", coinless.Reach)
	}
	if coinless.Payable {
		t.Fatalf("coinless: payable=true, want false (host has no coins)")
	}
	if coinless.Gate == "" {
		t.Fatalf("coinless: blocked verdict named no gate — the host can't SEE what stopped it")
	}
	if coinless.Needs == "" {
		t.Fatalf("coinless: blocked verdict gave no requirement text")
	}
	if coinless.YouHave != 0 {
		t.Fatalf("coinless: you_have=%d, want 0", coinless.YouHave)
	}
	t.Logf("coinless => reach=%q gate=%q needs=%q you_have=%d payable=%v",
		coinless.Reach, coinless.Gate, coinless.Needs, coinless.YouHave, coinless.Payable)

	// 15 coins: gated (a gate is in the way but the host can pay it), and
	// you_have must report the 15 coins the host carries.
	rich := o.ExplainReach(lumbX, lumbY, mineX, mineY, Capability{Coins: 15})
	if rich.Reach != ReachGated {
		t.Fatalf("15-coin: reach=%q, want gated (host can pay the toll)", rich.Reach)
	}
	if !rich.Payable {
		t.Fatalf("15-coin: payable=false, want true")
	}
	if rich.Gate == "" {
		t.Fatalf("15-coin: gated verdict named no gate")
	}
	if rich.YouHave != 15 {
		t.Fatalf("15-coin: you_have=%d, want 15", rich.YouHave)
	}
	t.Logf("15-coin => reach=%q gate=%q needs=%q you_have=%d payable=%v",
		rich.Reach, rich.Gate, rich.Needs, rich.YouHave, rich.Payable)

	// Prince Ali Rescue complete: the toll is waived, so the host reaches the
	// mine for free — either open (no remaining cut) or gated free-pass; the
	// load-bearing assertion is "not blocked".
	free := o.ExplainReach(lumbX, lumbY, mineX, mineY,
		Capability{Coins: 0, QuestsDone: map[string]bool{"Prince Ali Rescue": true}})
	if free.Reach == ReachBlocked {
		t.Fatalf("quest-free: reach=blocked, want open/gated (Prince Ali Rescue waives the toll)")
	}
	t.Logf("quest-free => reach=%q gate=%q needs=%q", free.Reach, free.Gate, free.Needs)
}

// TestExplainReachOpenNoGate checks that a destination on the same side of the
// toll gate as the start (no capability gate between them) reports Open, with
// no gate named.
func TestExplainReachOpenNoGate(t *testing.T) {
	o, _, _ := loadReal(t)

	const (
		lumbX, lumbY = 135, 654 // Lumbridge
		// A nearby Lumbridge tile east of the toll line (no toll between).
		nearX, nearY = 130, 648
	)
	info := o.ExplainReach(lumbX, lumbY, nearX, nearY, Capability{Coins: 0})
	if !info.SnapOK() {
		t.Skipf("near tile (%d,%d) did not snap; skipping open-reach check", nearX, nearY)
	}
	if info.Reach != ReachOpen {
		t.Fatalf("near Lumbridge tile: reach=%q, want open (no gate on the route); gate=%q",
			info.Reach, info.Gate)
	}
	if info.Gate != "" {
		t.Fatalf("open verdict should name no gate, got %q", info.Gate)
	}
	t.Logf("open => reach=%q (no gate, as expected)", info.Reach)
}

// TestExplainReachOffMap checks the void/off-map degrade: a destination that
// snaps to nothing reports SnapOK()==false so the host treats it as "no such
// place" rather than a real blocked verdict.
func TestExplainReachOffMap(t *testing.T) {
	o, _, _ := loadReal(t)
	info := o.ExplainReach(135, 654, -50, -50, Capability{})
	if info.SnapOK() {
		t.Fatalf("off-map destination reported SnapOK=true; want false")
	}
}
