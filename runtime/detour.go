package runtime

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/world"
)

// detourReq is one interrupt the conductor may act on: park the running routine,
// run intent as a detour, then resume. tier/reason are for the ladder + logs.
type detourReq struct {
	tier   string
	reason string
	intent Intent

	// abort marks an interrupt that does NOT park-and-resume. Instead it CANCELS
	// the running grind and ends the turn, bouncing control back to the director
	// for a fresh decision at the new situation. The displacement tier uses this:
	// there is no deterministic detour to run — the whole point is to RE-PLAN
	// because the host was unexpectedly moved. abort=false is the classic
	// park→detour→resume (survival).
	abort bool
}

const (
	// survivalThreshold: HP fraction below which the survival reflex fires.
	survivalThreshold = 0.35
	// survivalRecover: HP fraction at/above which the crisis is considered over
	// (hysteresis so a single low-HP episode triggers exactly one detour).
	survivalRecover = 0.55
	// detourTimeout bounds a single detour so a stuck detour can't strand the
	// parked grind forever.
	detourTimeout = 30 * time.Second

	// displacementThreshold is the Chebyshev tile jump, in a SINGLE own-position
	// update, that counts as a teleport/lure rather than a walk. RSC walks one
	// tile per server tick and opcode 191 delivers one absolute own-position per
	// tick, so a single-update jump beyond this is not a step (lured by a mob,
	// stairs/ladder, a random event, or a server teleport). Tunable.
	displacementThreshold = 8
)

// executeWithDetours runs one intent as a SUSPENDABLE Coro and services
// interrupts: a higher-tier event (survival) parks the routine, runs a detour to
// completion, then resumes the routine exactly where it left off. Replaces the
// plain blocking executeRoutine when Detours is enabled. The grind runs under the
// per-turn timeout; detours are short, so parking rarely consumes it.
func (c *Conductor) executeWithDetours(ctx context.Context, in Intent) Outcome {
	start := time.Now()
	turnCtx, cancel := context.WithTimeout(ctx, c.turnTimeout)
	// Register the cancel so Pause() (cradle pause / analysis mode) can interrupt
	// this turn promptly. The detour path also uses `cancel` directly to park /
	// abort the grind when a detour fires.
	c.pauseMu.Lock()
	c.turnCancel = cancel
	c.pauseMu.Unlock()
	defer func() {
		c.pauseMu.Lock()
		c.turnCancel = nil
		c.pauseMu.Unlock()
		cancel()
	}()

	coro := c.host.StartCoro(turnCtx, in)
	for {
		select {
		case <-coro.Done():
			out := coro.Outcome()
			out.Duration = time.Since(start)
			c.log.Info("conductor: turn complete", "intent", in.Label,
				"result", out.Kind.String(), "dur", out.Duration.Round(time.Millisecond))
			return out
		case req := <-c.interrupts:
			if !c.shouldDetour(req) {
				continue
			}
			if req.abort {
				// Abort tier (displacement): don't park-and-resume — CANCEL the
				// grind's turn and return its (errored) Outcome so the conductor
				// loops back to the director, which re-plans at the new position.
				// Cancelling turnCtx unblocks the grind at its next ctx checkpoint;
				// wait for the goroutine to exit before reading its Outcome (a
				// Coro's result fields are only safe to read after Done).
				cancel()
				<-coro.Done()
				out := coro.Outcome()
				out.Duration = time.Since(start)
				c.log.Info("conductor: turn aborted by interrupt", "tier", req.tier,
					"reason", req.reason, "result", out.Kind.String(),
					"dur", out.Duration.Round(time.Millisecond))
				c.drainInterrupts()
				return out
			}
			if coro.Suspend(turnCtx) { // park the grind (false if it already finished)
				c.runDetour(ctx, req)
				coro.Resume()
				c.drainInterrupts() // discard interrupts queued during the detour
			}
		}
	}
}

// shouldDetour decides whether an interrupt preempts the running routine. Today
// survival always preempts; committed-region deferral + lower tiers land here.
func (c *Conductor) shouldDetour(req detourReq) bool {
	switch req.tier {
	case "survival":
		return true
	case "displacement":
		// A large unexpected jump preempts a normal grind, but DEFERS while in a
		// committed interaction (open trade/duel/bank) — finishing the deal wins
		// over re-planning. (Same gate as the default; spelled out for the ladder.)
		return c.host.world == nil || !c.host.world.InCommittedRegion()
	case "reactive":
		// A reactive interrupt (an urgent overheard/directed signal) parks the
		// grind to re-plan over the freshly-written knowledge, but DEFERS while in
		// a committed interaction (open trade/duel/bank) — don't yank her out of a
		// deal to chase a chat line. (Same gate as the default; spelled out for the
		// ladder + logging.)
		return c.host.world == nil || !c.host.world.InCommittedRegion()
	case "forage":
		// A forage trip parks the grind to go LOOK at a source, but DEFERS while in a
		// committed interaction (open trade/duel/bank) — don't yank her out of a deal
		// to wander to a shop. Same gate as reactive/displacement.
		return c.host.world == nil || !c.host.world.InCommittedRegion()
	default:
		// A non-survival interrupt is deferred while in a committed interaction
		// (open trade/duel/bank) — don't yank her out of a deal.
		return c.host.world == nil || !c.host.world.InCommittedRegion()
	}
}

// runDetour runs the interrupt's routine to completion (or timeout) while the
// grind is parked. Detours are not themselves interruptible in this slice.
func (c *Conductor) runDetour(ctx context.Context, req detourReq) {
	c.log.Info("detour: parking grind for interrupt", "tier", req.tier, "reason", req.reason)
	c.host.publishDecision("detour:"+req.tier, "interrupt", "parking the grind — "+req.reason)
	dctx, dcancel := context.WithTimeout(ctx, detourTimeout)
	defer dcancel()
	d := c.host.StartCoro(dctx, req.intent)
	select {
	case <-d.Done():
		c.log.Info("detour: complete, resuming grind", "tier", req.tier, "result", d.Outcome().Kind.String())
	case <-ctx.Done():
	}
}

// drainInterrupts discards any interrupts that queued up while a detour ran (they
// are stale — the situation was just handled).
func (c *Conductor) drainInterrupts() {
	for {
		select {
		case <-c.interrupts:
		default:
			return
		}
	}
}

// signalDetour offers a detour request to the conductor, non-blocking (a full
// queue means an interrupt is already pending — drop the duplicate).
func (c *Conductor) signalDetour(req detourReq) {
	select {
	case c.interrupts <- req:
	default:
	}
}

// hpFraction is current HP as a fraction of max (1.0 when max is unknown).
func (h *Host) hpFraction() float64 {
	self := h.world.Self
	mx := self.MaxHP()
	if mx <= 0 {
		return 1
	}
	return float64(self.HP()) / float64(mx)
}

// survivalArbiter watches HP-affecting events and, on a NEW drop into the
// survival band, signals a survival detour (eat). Edge-triggered with hysteresis
// so one low-HP episode produces one detour. Deterministic, no LLM. Exits with ctx.
func (c *Conductor) survivalArbiter(ctx context.Context) {
	ch := c.host.bus.Subscribe("*", 256)
	inCrisis := false
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch ev.(type) {
			case event.OtherPlayerDamage, event.NpcDamage, event.StatUpdate, event.StatsSnapshot, event.Death:
			default:
				continue // only re-check HP on events that can change it
			}
			frac := c.host.hpFraction()
			switch {
			case !inCrisis && frac > 0 && frac < survivalThreshold:
				inCrisis = true
				c.log.Info("survival: HP critical — requesting detour", "hp_pct", int(frac*100))
				c.signalDetour(detourReq{tier: "survival", reason: fmt.Sprintf("HP %d%%", int(frac*100)), intent: survivalIntent()})
			case inCrisis && frac >= survivalRecover:
				inCrisis = false
			}
		}
	}
}

// survivalIntent is the deterministic survival detour: eat the best food on hand
// (no LLM). Mirrors examples/routines/auto_eat. Holds if she has nothing to eat.
func survivalIntent() Intent {
	const src = `runtime "1.0"
routine survival_eat() {
    note("survival reflex: HP low — eating.")
    food = inventory.find("cookedmeat")
    if food == null { food = inventory.find("lobster") }
    if food == null { food = inventory.find("swordfish") }
    if food == null { food = inventory.find("salmon") }
    if food == null { food = inventory.find("trout") }
    if food == null { food = inventory.find("bread") }
    if food == null { food = inventory.find("meat") }
    if food != null {
        eat(food)
        wait(0.6)
    } else {
        note("survival reflex: no food on hand — holding.")
    }
}`
	return Intent{Label: "detour:survival_eat", Name: "survival_eat", Source: src}
}

// displacementState carries the most recent unexpected position jump from the
// displacementArbiter (a conductor goroutine) to the director (which builds the
// next Situation), so a re-plan after a displacement abort can tell the planner
// the host was MOVED — lured, teleported, dragged, took stairs — rather than
// just "the last thing failed". Consume-once: the director clears it when it
// folds it into the next trigger.
type displacementState struct {
	mu      sync.Mutex
	pending *displacementEvent
}

// dispCause distinguishes the KIND of involuntary relocation, so the director
// can phrase the re-plan correctly (a lure reads very differently from a death).
type dispCause int

const (
	dispJump  dispCause = iota // a lure / teleport / random relocation while alive
	dispDeath                  // died and respawned at the respawn point
)

// displacementEvent is one recorded relocation: where the host was, where it
// landed, the Chebyshev distance between, and what caused it.
type displacementEvent struct {
	fromX, fromY int
	toX, toY     int
	dist         int
	cause        dispCause
}

func (d *displacementState) record(e displacementEvent) {
	d.mu.Lock()
	d.pending = &e
	d.mu.Unlock()
}

// take returns and clears the pending displacement (consume-once); ok is false
// when nothing is pending.
func (d *displacementState) take() (displacementEvent, bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.pending == nil {
		return displacementEvent{}, false
	}
	e := *d.pending
	d.pending = nil
	return e, true
}

// isDisplacement reports whether moving from prev to cur in a SINGLE position
// update is a teleport/lure rather than a walk: a same-plane Chebyshev jump of
// at least displacementThreshold tiles. Plane changes (ladders/stairs) are
// intended movement — planes are Y-offset bands of PlaneHeight, so climbing
// looks like a ~944-tile jump — and never count. Pure; unit-tested directly.
func isDisplacement(prev, cur world.Coord) bool {
	if world.PlaneOf(prev.Y) != world.PlaneOf(cur.Y) {
		return false
	}
	return chebyshev(prev.X, prev.Y, cur.X, cur.Y) >= displacementThreshold
}

// displacementArbiter watches the host's own-position stream and, on a LARGE
// unexpected single-update jump (lured by a monster, a teleport, a stairs/ladder
// crossing, a random event), signals an ABORT detour: the current grind is
// cancelled and the director re-plans at the new position. Edge-triggered on the
// PER-UPDATE delta — RSC walks one tile/tick, so a long intended go_to produces
// a stream of small deltas, never one big jump. Deterministic, no LLM. Exits
// with ctx. Mirrors survivalArbiter's shape.
//
// False positives it deliberately avoids:
//   - login / session start: seeds on the first update, never fires on it.
//   - intended long travel: thresholds the per-update delta, not start-vs-end.
//   - plane changes (ladders/stairs): excluded by isDisplacement.
//
// Death is the exception that DOES fire: dying is involuntary relocation too —
// the host respawns at the respawn point having dropped what it carried, so the
// current grind is invalid no matter how far the respawn is. A Death arms a
// re-plan that fires (with death context) on the respawn position update,
// regardless of distance. (Survival still owns the low-HP path before death.)
func (c *Conductor) displacementArbiter(ctx context.Context) {
	ch := c.host.bus.Subscribe("*", 256)
	var prev world.Coord
	havePrev := false
	pendingDeath := false // a Death just fired; the next position update is the respawn
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch e := ev.(type) {
			case event.Death:
				pendingDeath = true
			case event.OwnPositionUpdate:
				cur := world.Coord{X: e.X, Y: e.Y}
				if !havePrev {
					prev, havePrev, pendingDeath = cur, true, false
					continue
				}
				from := prev
				prev = cur // re-baseline every update so a run of normal steps never accumulates

				// Death: always re-plan on the respawn update, with death context,
				// regardless of jump distance (the respawn point may be near or far).
				if pendingDeath {
					pendingDeath = false
					d := chebyshev(from.X, from.Y, cur.X, cur.Y)
					c.log.Info("displacement: died and respawned — aborting turn to re-plan",
						"to", fmt.Sprintf("(%d,%d)", cur.X, cur.Y))
					c.host.displacement.record(displacementEvent{
						fromX: from.X, fromY: from.Y, toX: cur.X, toY: cur.Y, dist: d, cause: dispDeath,
					})
					c.signalDetour(detourReq{
						tier:   "displacement",
						abort:  true,
						reason: fmt.Sprintf("died and respawned at (%d, %d)", cur.X, cur.Y),
					})
					continue
				}

				if !isDisplacement(from, cur) {
					continue
				}
				d := chebyshev(from.X, from.Y, cur.X, cur.Y)
				c.log.Info("displacement: large unexpected position jump — aborting turn to re-plan",
					"dist", d, "from", fmt.Sprintf("(%d,%d)", from.X, from.Y),
					"to", fmt.Sprintf("(%d,%d)", cur.X, cur.Y))
				c.host.displacement.record(displacementEvent{
					fromX: from.X, fromY: from.Y, toX: cur.X, toY: cur.Y, dist: d, cause: dispJump,
				})
				c.signalDetour(detourReq{
					tier:   "displacement",
					abort:  true,
					reason: fmt.Sprintf("position jumped %d tiles to (%d, %d)", d, cur.X, cur.Y),
				})
			}
		}
	}
}
