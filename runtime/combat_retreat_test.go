package runtime

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// Anti-kite retreat tests (#r3-retreat). RSC forbids retreating until
// the opponent has made >= 3 hits ("first 3 rounds of combat",
// WalkRequest.java). The engine tracks a round count off the combat
// damage exchange so retreat() can wait out the gate, and surfaces the
// server's rejection message as a typed RETREAT_TOO_EARLY result. Eat
// in combat surfaces a typed EAT_IN_COMBAT result.

// TestBeginCombatRoundTrackingResetsOnNewTarget: attacking a fresh
// target zeroes the round count; re-attacking the SAME target mid-fight
// keeps the running count (so dropping+reissuing attack on a target
// that walked doesn't erase progress toward the 3-round gate).
func TestBeginCombatRoundTrackingResetsOnNewTarget(t *testing.T) {
	h := newTestHost()
	h.beginCombatRoundTracking(7)
	if h.combatRoundTarget != 7 || h.combatRounds != 0 {
		t.Fatalf("begin: target=%d rounds=%d, want 7/0", h.combatRoundTarget, h.combatRounds)
	}
	h.combatRounds = 2
	// Same target: keep progress.
	h.beginCombatRoundTracking(7)
	if h.combatRounds != 2 {
		t.Errorf("re-attack same target: rounds=%d, want 2 (kept)", h.combatRounds)
	}
	// New target: reset.
	h.beginCombatRoundTracking(9)
	if h.combatRoundTarget != 9 || h.combatRounds != 0 {
		t.Errorf("new target: target=%d rounds=%d, want 9/0", h.combatRoundTarget, h.combatRounds)
	}
}

// TestNoteCombatRoundCountsExchange: each combat-tick damage event on
// the tracked engagement advances the round count — our hit on the NPC
// (NpcDamage on the tracked index) OR the NPC's hit on us
// (OtherPlayerDamage on self, player_index 0). Damage involving other
// entities is ignored.
func TestNoteCombatRoundCountsExchange(t *testing.T) {
	h := newTestHost()
	h.beginCombatRoundTracking(7)

	// Our hit lands on the engaged NPC -> +1.
	h.noteCombatRound(event.NpcDamage{NpcIndex: 7, Damage: 3, CurHits: 9, MaxHits: 12})
	// The NPC hits us back (self slot 0) -> +1.
	h.noteCombatRound(event.OtherPlayerDamage{PlayerIndex: 0, Damage: 2})
	// Damage on a DIFFERENT npc -> ignored.
	h.noteCombatRound(event.NpcDamage{NpcIndex: 99, Damage: 1, CurHits: 5, MaxHits: 6})
	// Damage on a DIFFERENT player -> ignored.
	h.noteCombatRound(event.OtherPlayerDamage{PlayerIndex: 4, Damage: 5})

	if h.combatRounds != 2 {
		t.Errorf("combatRounds = %d, want 2 (one exchange)", h.combatRounds)
	}

	// With no tracked target, nothing counts.
	h.combatRoundTarget = 0
	h.noteCombatRound(event.NpcDamage{NpcIndex: 7, Damage: 3, CurHits: 6, MaxHits: 12})
	if h.combatRounds != 2 {
		t.Errorf("untracked: combatRounds = %d, want 2 (unchanged)", h.combatRounds)
	}
}

// TestWaitForRetreatGateReturnsWhenRoundsReached: the gate wait returns
// promptly once combatRounds reaches the 3-round threshold.
func TestWaitForRetreatGateReturnsWhenRoundsReached(t *testing.T) {
	h := newTestHost()
	h.beginCombatRoundTracking(7)
	h.combatRounds = retreatRoundGate // already satisfied
	start := time.Now()
	if err := h.waitForRetreatGate(context.Background()); err != nil {
		t.Fatalf("waitForRetreatGate: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("waited %v with gate already satisfied — should return immediately", elapsed)
	}
}

// TestWaitForRetreatGateReturnsWhenEngagementCleared: if the engagement
// we were counting clears (target died) the wait returns rather than
// blocking for the timeout.
func TestWaitForRetreatGateReturnsWhenEngagementCleared(t *testing.T) {
	h := newTestHost()
	h.combatRoundTarget = 0 // no engagement to wait on
	start := time.Now()
	if err := h.waitForRetreatGate(context.Background()); err != nil {
		t.Fatalf("waitForRetreatGate: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("waited %v with no engagement — should return immediately", elapsed)
	}
}

// TestWaitForRetreatGateHonorsContextCancel: a cancelled context aborts
// the wait with ctx.Err().
func TestWaitForRetreatGateHonorsContextCancel(t *testing.T) {
	h := newTestHost()
	h.beginCombatRoundTracking(7) // rounds=0, gate not met -> would wait
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := h.waitForRetreatGate(ctx); err == nil {
		t.Error("waitForRetreatGate with cancelled ctx: got nil, want ctx.Err()")
	}
}

// TestMapRetreatErrTooEarly: a RetreatTooEarlyError maps to the typed
// RETREAT_TOO_EARLY result code, carrying the server prose so a routine
// can surface it verbatim.
func TestMapRetreatErrTooEarly(t *testing.T) {
	const prose = "You can't retreat during the first 3 rounds of combat"
	v := mapRetreatErr(&RetreatTooEarlyError{ServerMessage: prose, RoundsSeen: 1})
	cr, ok := v.(*interp.CallResult)
	if !ok || cr.Err == nil {
		t.Fatalf("mapRetreatErr: got %v, want a failed CallResult", v)
	}
	if cr.Err.Code != interp.RETREAT_TOO_EARLY {
		t.Errorf("code = %s, want RETREAT_TOO_EARLY", cr.Err.Code)
	}
	if cr.Err.Reason != prose {
		t.Errorf("reason = %q, want the server prose %q", cr.Err.Reason, prose)
	}
}

// TestMapRetreatErrFallsThrough: a non-anti-kite error falls through to
// the generic server-error classifier (e.g. OUT_OF_RANGE).
func TestMapRetreatErrFallsThrough(t *testing.T) {
	v := mapRetreatErr(errf("target tile out of range"))
	cr, ok := v.(*interp.CallResult)
	if !ok || cr.Err == nil {
		t.Fatalf("mapRetreatErr: got %v, want a failed CallResult", v)
	}
	if cr.Err.Code != interp.OUT_OF_RANGE {
		t.Errorf("code = %s, want OUT_OF_RANGE", cr.Err.Code)
	}
}

// TestCombatRoundTrackingRaceFree is the H11-regression -race guard: the
// combat-round fields (combatRoundTarget/combatRounds/outgoingHits/
// combatStartedAt) are touched by THREE goroutines — the frame-pump
// (noteCombatRound / emitTargetDeathEdge), the conductor/routine
// (beginCombatRoundTracking / the retreat gate / confirmEngaged), and the
// limbic gate. Before combatMu covered the conductor accessors too, the plain
// ints raced. This exercises the frame-pump writer, the conductor writer, and
// the conductor reader concurrently; it must be clean under `go test -race`.
func TestCombatRoundTrackingRaceFree(t *testing.T) {
	h := newTestHost()
	// A live NPC so confirmEngaged + emitTargetDeathEdge have a roster entry.
	h.world.Npcs.Set(world.NpcRecord{Index: 7, TypeID: 1, X: 100, Y: 100})
	h.lastAttackedNpcIndex.Store(7)
	h.beginCombatRoundTracking(7)

	const iters = 2000
	var wg sync.WaitGroup
	wg.Add(3)

	// Frame-pump writer: damage ticks advance the round counter (locks combatMu).
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.noteCombatRound(event.NpcDamage{NpcIndex: 7, Damage: 1, CurHits: 9, MaxHits: 12})
			h.noteCombatRound(event.OtherPlayerDamage{PlayerIndex: 0, Damage: 1})
		}
	}()

	// Conductor writer: re-arm tracking, alternating the target so the reset
	// branch (write) actually fires (locks combatMu).
	go func() {
		defer wg.Done()
		for i := 0; i < iters; i++ {
			h.beginCombatRoundTracking(7 + (i & 1))
		}
	}()

	// Conductor reader: the retreat gate snapshots the same fields under the lock.
	// Use an already-cancelled context so each call does exactly one locked read
	// then returns (gate satisfied → nil, else ctx.Done) — fast, no real waiting.
	go func() {
		defer wg.Done()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		for i := 0; i < iters; i++ {
			_ = h.waitForRetreatGate(ctx)
		}
	}()

	wg.Wait()
}

// TestEmitTargetDeathClearsRoundTracking: when the engaged target dies,
// the anti-kite round tracking resets so the next engagement starts
// fresh (no stale "rounds already elapsed").
func TestEmitTargetDeathClearsRoundTracking(t *testing.T) {
	h := newTestHost()
	h.world.Npcs.Set(world.NpcRecord{Index: 5, TypeID: 1, X: 100, Y: 100})
	h.world.Npcs.SetHits(5, 4, 0, 10) // dead: cur 0
	h.lastAttackedNpcIndex.Store(5)
	h.beginCombatRoundTracking(5)
	h.combatRounds = 3

	h.emitTargetDeathEdge(5, true) // alive->dead edge

	if h.combatRoundTarget != 0 || h.combatRounds != 0 {
		t.Errorf("after death: target=%d rounds=%d, want 0/0", h.combatRoundTarget, h.combatRounds)
	}
}
