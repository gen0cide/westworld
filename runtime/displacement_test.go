package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// TestIsDisplacement covers the pure decision: a same-plane Chebyshev jump at or
// beyond the threshold is a displacement; a small step, a sub-threshold jump,
// and ANY plane change (ladder/stairs) are not.
func TestIsDisplacement(t *testing.T) {
	c := func(x, y int) world.Coord { return world.Coord{X: x, Y: y} }
	cases := []struct {
		name     string
		from, to world.Coord
		want     bool
	}{
		{"single step", c(100, 100), c(101, 100), false},
		{"diagonal step", c(100, 100), c(101, 101), false},
		{"sub-threshold", c(100, 100), c(100+displacementThreshold-1, 100), false},
		{"at threshold", c(100, 100), c(100+displacementThreshold, 100), true},
		{"big jump", c(100, 100), c(160, 140), true},
		{"plane change only", c(100, 100), c(100, 100+world.PlaneHeight), false},
		{"plane change with jump", c(100, 100), c(160, 100+world.PlaneHeight), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isDisplacement(tc.from, tc.to); got != tc.want {
				t.Fatalf("isDisplacement(%v, %v) = %v, want %v", tc.from, tc.to, got, tc.want)
			}
		})
	}
}

// TestConductorDisplacementAborts: an abort-tier interrupt cancels the running
// grind and ends the turn promptly with a non-OK Outcome, so the director gets a
// fresh turn. The grind alone would run ~20s; the abort must end it far sooner.
func TestConductorDisplacementAborts(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{
		Detours: true, Settle: -1,
		Director: DirectorFunc(func(context.Context, *Host, Outcome) (Intent, bool) { return Intent{}, false }),
	})

	grind := Intent{Name: "grind", Source: `runtime "1.0"
routine grind() {
    i = 0
    while i < 100 {
        i = i + 1
        wait(0.2)
    }
    return i
}`}

	var out Outcome
	done := make(chan struct{})
	go func() {
		out = c.executeWithDetours(context.Background(), grind)
		close(done)
	}()

	time.Sleep(250 * time.Millisecond) // let the grind reach a wait checkpoint
	c.interrupts <- detourReq{tier: "displacement", abort: true, reason: "test jump"}

	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("executeWithDetours never returned after an abort — the abort path deadlocked")
	}

	if out.OK() {
		t.Fatalf("aborted turn reported OK (kind=%v) — want a non-OK Outcome so the director re-plans", out.Kind)
	}
	if out.Duration > 3*time.Second {
		t.Fatalf("abort took %v — it must end the turn promptly, not run the grind out", out.Duration)
	}
}

// TestShouldDetourDisplacement: displacement preempts a normal grind when free,
// and DEFERS while in a committed interaction (open bank). Survival still always
// preempts.
func TestShouldDetourDisplacement(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})

	if h.world.InCommittedRegion() {
		t.Fatal("precondition: fresh host should not be in a committed region")
	}
	if !c.shouldDetour(detourReq{tier: "displacement", abort: true}) {
		t.Fatal("displacement should preempt when not in a committed region")
	}

	h.world.Bank.Open(48, nil) // open the bank → committed
	if !h.world.InCommittedRegion() {
		t.Fatal("opening the bank should put the host in a committed region")
	}
	if c.shouldDetour(detourReq{tier: "displacement", abort: true}) {
		t.Fatal("displacement should DEFER while in a committed region")
	}
	if !c.shouldDetour(detourReq{tier: "survival"}) {
		t.Fatal("survival must preempt even in a committed region")
	}
}

// TestDisplacementArbiterFiresOnLargeJump drives the arbiter over the bus: the
// first update seeds (no fire), a 1-tile walk does not fire, a large jump fires
// exactly one abort interrupt and records the jump for the director, and a
// following normal step does not fire (edge-triggered re-baseline).
func TestDisplacementArbiterFiresOnLargeJump(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})

	go c.displacementArbiter(t.Context())
	time.Sleep(50 * time.Millisecond) // let the arbiter subscribe before we publish (the bus has no history)

	pub := func(x, y int) { h.bus.Publish(event.OwnPositionUpdate{X: x, Y: y}) }

	pub(100, 100) // seed — must not fire
	pub(101, 100) // 1-tile walk — must not fire
	jx := 101 + displacementThreshold
	pub(jx, 100) // large same-plane jump — must fire

	select {
	case req := <-c.interrupts:
		if req.tier != "displacement" || !req.abort {
			t.Fatalf("interrupt tier=%q abort=%v, want displacement/true", req.tier, req.abort)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("arbiter did not fire on a large jump")
	}

	disp, ok := h.displacement.take()
	if !ok || disp.dist < displacementThreshold || disp.toX != jx || disp.toY != 100 {
		t.Fatalf("displacement not recorded for the director: ok=%v %+v", ok, disp)
	}

	pub(jx+1, 100) // normal step from the new baseline — must NOT fire
	select {
	case req := <-c.interrupts:
		t.Fatalf("arbiter fired on a normal step after the jump: %+v", req)
	case <-time.After(250 * time.Millisecond):
	}
}

// TestDisplacementArbiterFiresOnDeathRespawn: dying is involuntary relocation
// too — a Death must fire a re-plan on the respawn update, tagged dispDeath, even
// when the respawn is CLOSE (it is not distance-gated: death invalidates the plan
// regardless of how far the respawn point is).
func TestDisplacementArbiterFiresOnDeathRespawn(t *testing.T) {
	h := newTestHost()
	c := NewConductor(h, ConductorOptions{Detours: true, Director: nil})

	go c.displacementArbiter(t.Context())
	time.Sleep(50 * time.Millisecond) // let the arbiter subscribe before we publish (the bus has no history)

	h.bus.Publish(event.OwnPositionUpdate{X: 100, Y: 100}) // seed
	h.bus.Publish(event.Death{})                           // died
	h.bus.Publish(event.OwnPositionUpdate{X: 101, Y: 100}) // respawn just 1 tile away — still a re-plan

	select {
	case req := <-c.interrupts:
		if req.tier != "displacement" || !req.abort {
			t.Fatalf("interrupt tier=%q abort=%v, want displacement/true", req.tier, req.abort)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("arbiter did not re-plan after death/respawn")
	}

	disp, ok := h.displacement.take()
	if !ok || disp.cause != dispDeath {
		t.Fatalf("death not recorded as a dispDeath relocation for the director: ok=%v %+v", ok, disp)
	}
}
