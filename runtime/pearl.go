package runtime

import (
	"context"
	"strings"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/pearl"
)

// fatigueScaleMax is RSC's server-scaled fatigue ceiling (0..750 maps to 0..100%).
const fatigueScaleMax = 750.0

// pearlFacts assembles the read-only Facts snapshot the policy engine evaluates,
// from already-live host state. It is allocation-light and called on the hot
// path (every gated action / decide), so it reads only cheap accessors. Affect
// and the counterparty Relationship are left zero/nil until those subsystems
// are wired — predicates that depend on them simply don't fire (safe by design).
func (h *Host) pearlFacts(ev pearl.EventCtx) *pearl.Facts {
	self := h.world.Self
	pos := self.Position()
	f := &pearl.Facts{
		HPCur:       self.HP(),
		HPMax:       self.MaxHP(),
		CombatLevel: self.CombatLevel(),
		X:           pos.X,
		Y:           pos.Y,
		Event:       ev,
		Now:         time.Now(),
	}
	if fat := self.Fatigue(); fat > 0 {
		f.Fatigue = float64(fat) / fatigueScaleMax
	}
	if h.affect != nil {
		s, c, v := h.affect.Snapshot()
		f.Affect = pearl.Affect{Stress: s, Confidence: c, Valence: v}
	}
	if inv := h.world.Inventory; inv != nil {
		f.InvFree = inv.FreeSlots()
		counts := make(map[string]int)
		for _, s := range inv.Slots() {
			if s.ItemID <= 0 {
				continue
			}
			name := h.itemName(s.ItemID)
			if name == "" {
				continue
			}
			n := s.Amount
			if n <= 0 {
				n = 1
			}
			counts[strings.ToLower(name)] += n
		}
		f.InvCounts = counts
	}
	return f
}

// tutorialIsland is the coordinate box covering RuneScape tutorial island. While
// the host is inside it, the pearl gate is bypassed (the tutorial-scoped
// "vetoes off" policy) so a pacifist host can still complete required combat
// training; full disposition resumes once she reaches the mainland. Tunable.
const (
	tutXMin, tutXMax = 190, 290
	tutYMin, tutYMax = 700, 775
)

// onTutorialIsland reports whether the host is currently on tutorial island.
func (h *Host) onTutorialIsland() bool {
	if h.world == nil || h.world.Self == nil {
		return false
	}
	p := h.world.Self.Position()
	return p.X >= tutXMin && p.X <= tutXMax && p.Y >= tutYMin && p.Y <= tutYMax
}

// itemName resolves an item id to its facts name (empty when facts aren't
// loaded or the id is unknown).
func (h *Host) itemName(id int) string {
	if h.facts == nil {
		return ""
	}
	if d := h.facts.ItemDef(id); d != nil {
		return d.Name
	}
	return ""
}

// gateAction wraps a primary action handler with the pearl Apperception gate.
// Before the action runs, the engine may veto it (returns a POLICY_VETO typed
// failure) or substitute a different action. An allow passes straight through.
// Only installed when h.Pearl != nil (see NewRoutineInterpreter), so the
// unwrapped path is unchanged for hosts without an engine.
func (h *Host) gateAction(name string, inner actionHandler) actionHandler {
	return func(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
		// Tutorial-scoped policy bypass: on tutorial island the host must be able
		// to complete required steps (notably the combat training a pacifist would
		// otherwise veto). Disposition resumes the moment she reaches the mainland.
		if h.onTutorialIsland() {
			return inner(ctx, h, args, named)
		}
		v := h.Pearl.Gate(h.pearlFacts(pearl.EventCtx{Action: name}), name, stringArgs(args))
		if v.Allow {
			return inner(ctx, h, args, named)
		}
		if v.Substitute != nil {
			if sub := actionHandlers[v.Substitute.Action]; sub != nil {
				h.log.Info("pearl: substituting action", "from", name, "to", v.Substitute.Action, "rule", v.RuleID)
				return sub(ctx, h, substituteArgs(v.Substitute), nil)
			}
			// No handler for the substitute — fall back to vetoing the original.
		}
		h.log.Info("pearl: vetoed action", "action", name, "rule", v.RuleID, "reason", v.Reason)
		reason := v.Reason
		if reason == "" {
			reason = "denied by host policy"
		}
		// Surface the veto so the host's own cognition can see it (Act transcript)
		// instead of treating a vetoed action as a silent no-op and retrying.
		if h.bus != nil {
			h.bus.Publish(event.PolicyVeto{Action: name, Rule: v.RuleID, Reason: reason})
		}
		return interp.Fail(interp.POLICY_VETO, reason), nil
	}
}

// stringArgs renders positional DSL args to plain strings for the gate's
// predicate view (the engine only needs to compare literal forms).
func stringArgs(args []interp.Value) []string {
	if len(args) == 0 {
		return nil
	}
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = stringOf(a)
	}
	return out
}

// substituteArgs converts a pearl Effect's string args into interp values for
// the substituted handler.
func substituteArgs(e *pearl.Effect) []interp.Value {
	if len(e.Args) == 0 {
		return nil
	}
	out := make([]interp.Value, len(e.Args))
	for i, a := range e.Args {
		out[i] = interp.String(a)
	}
	return out
}
