package runtime

import (
	"context"
	"fmt"
	"time"

	"github.com/gen0cide/westworld/event"
)

// detourReq is one interrupt the conductor may act on: park the running routine,
// run intent as a detour, then resume. tier/reason are for the ladder + logs.
type detourReq struct {
	tier   string
	reason string
	intent Intent
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
)

// executeWithDetours runs one intent as a SUSPENDABLE Coro and services
// interrupts: a higher-tier event (survival) parks the routine, runs a detour to
// completion, then resumes the routine exactly where it left off. Replaces the
// plain blocking executeRoutine when Detours is enabled. The grind runs under the
// per-turn timeout; detours are short, so parking rarely consumes it.
func (c *Conductor) executeWithDetours(ctx context.Context, in Intent) Outcome {
	start := time.Now()
	turnCtx, cancel := context.WithTimeout(ctx, c.turnTimeout)
	defer cancel()

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
