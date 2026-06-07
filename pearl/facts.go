package pearl

import (
	"slices"
	"time"
)

// Facts is the read-only snapshot a rule predicate evaluates against. It is a
// value snapshot (not a live view) assembled once per Gate/TryDecide call from
// already-live host state — the world mirror, the per-host affect vector, and
// the relationship row for the current counterparty. Predicates must treat it
// as immutable.
//
// Fields that depend on subsystems not yet wired (Affect, Counter) are left
// zero/nil; predicates that read them simply do not fire, which is safe by
// design.
type Facts struct {
	// Self vitals / position, from the world mirror.
	HPCur       int
	HPMax       int
	Fatigue     float64 // 0..1
	CombatLevel int
	X, Y        int

	// Inventory.
	InvFree   int
	InvCounts map[string]int // normalized (lower-cased) item name → count

	// Affect: the per-host mood vector (zero until the affect path is wired).
	Affect Affect

	// Counter: trust/affinity row for the other party in this interaction;
	// nil when there is none or the relationship subsystem is not wired.
	Counter *Relationship

	// Event: what triggered this evaluation.
	Event EventCtx

	Now time.Time
}

// Affect is the host's mood vector.
type Affect struct {
	Stress     float64 // 0..1
	Confidence float64 // 0..1
	Valence    float64 // -1..1
}

// Relationship is the host's view of another party.
type Relationship struct {
	Name     string
	Trust    float64 // -1..1
	Familiar int     // encounter count
	Tags     []string
}

// EventCtx describes what triggered the current evaluation.
type EventCtx struct {
	// Action is the action name being gated (action seam), or "decide" on the
	// decision seam.
	Action string
	// BusKind is the triggering bus event kind, when the evaluation was driven
	// by an event (e.g. "player_appeared", "attacked").
	BusKind string
	// Question is the decide() question, on the decision seam.
	Question string
	// TargetCombatLevel is the combat level of the action's target (an NPC or
	// player), or 0 when there is no relevant target.
	TargetCombatLevel int
}

// HPFraction returns current HP as a fraction of max (0 when max is unknown).
func (f *Facts) HPFraction() float64 {
	if f.HPMax <= 0 {
		return 0
	}
	return float64(f.HPCur) / float64(f.HPMax)
}

// hasTag reports whether the counterparty carries tag.
func (r *Relationship) hasTag(tag string) bool {
	return slices.Contains(r.Tags, tag)
}
