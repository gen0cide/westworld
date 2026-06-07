package runtime

import (
	"context"
	"sort"
)

// The Director answers "what should I do now?" at routine granularity — the
// host-local, deterministic, no-LLM layer above the Conductor. GoalDirector
// models it as a prioritized set of situated behaviors (Drives): each turn it
// runs the highest-priority eligible Drive, falling back to an idle behavior.
//
// This is the cheap, always-available planner. A future brain-assisted planner
// can wrap or replace it at the same Director seam; a Drive's Build closure can
// also consult mesa-backed memory (recollect the current goal, etc.) without
// the Director itself knowing about mesa.

// Drive is one candidate behavior. When its guard holds, Build proposes the
// Intent to run for this turn.
type Drive struct {
	Name     string
	Priority int                          // higher wins; ties break by registration order
	When     func(h *Host) bool           // eligibility guard; nil ⇒ always eligible
	Build    func(h *Host) (Intent, bool) // produce the intent; false ⇒ decline this turn
}

// GoalDirector selects the highest-priority eligible Drive each turn and runs
// its Intent. If none is eligible it runs the idle Intent (when set) or stops.
// Implements Director, so it plugs straight into the Conductor.
type GoalDirector struct {
	drives []Drive
	idle   *Intent
}

// NewGoalDirector builds a director. idle is run when no drive is eligible; pass
// nil to stop the conductor instead. drives may be given in any order — they are
// sorted by descending priority (stable, so equal priorities keep call order).
func NewGoalDirector(idle *Intent, drives ...Drive) *GoalDirector {
	ds := make([]Drive, len(drives))
	copy(ds, drives)
	sort.SliceStable(ds, func(i, j int) bool { return ds[i].Priority > ds[j].Priority })
	return &GoalDirector{drives: ds, idle: idle}
}

// Next implements Director: pick the first (highest-priority) drive whose guard
// holds and whose Build yields an intent.
func (d *GoalDirector) Next(_ context.Context, h *Host, _ Outcome) (Intent, bool) {
	for i := range d.drives {
		dr := &d.drives[i]
		if dr.When != nil && !dr.When(h) {
			continue
		}
		if dr.Build == nil {
			continue
		}
		if intent, ok := dr.Build(h); ok {
			return intent, true
		}
	}
	if d.idle != nil {
		return *d.idle, true
	}
	return Intent{}, false
}

// RoutineDrive is a convenience Drive that always proposes the same routine when
// its guard holds.
func RoutineDrive(name string, priority int, when func(*Host) bool, routinePath string) Drive {
	return Drive{
		Name:     name,
		Priority: priority,
		When:     when,
		Build: func(*Host) (Intent, bool) {
			return Intent{Label: name, RoutinePath: routinePath}, true
		},
	}
}

// --- guard helpers (over live host state) -----------------------------------

// WhenHPBelow holds when current HP is below frac of max (0..1).
func WhenHPBelow(frac float64) func(*Host) bool {
	return func(h *Host) bool {
		cur, max := h.world.Self.HP(), h.world.Self.MaxHP()
		return max > 0 && float64(cur)/float64(max) < frac
	}
}

// WhenFatigueAbove holds when fatigue exceeds frac (0..1).
func WhenFatigueAbove(frac float64) func(*Host) bool {
	return func(h *Host) bool {
		return float64(h.world.Self.Fatigue())/fatigueScaleMax > frac
	}
}

// WhenInventoryFull holds when there are no free inventory slots.
func WhenInventoryFull() func(*Host) bool {
	return func(h *Host) bool {
		return h.world.Inventory != nil && h.world.Inventory.FreeSlots() == 0
	}
}

// And combines guards: holds only when all do (nil entries are ignored).
func And(guards ...func(*Host) bool) func(*Host) bool {
	return func(h *Host) bool {
		for _, g := range guards {
			if g != nil && !g(h) {
				return false
			}
		}
		return true
	}
}
