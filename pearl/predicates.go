package pearl

import "strings"

// This file is the fixed predicate vocabulary: a small set of PARAMETERIZED
// primitives plus three combinators. Variety comes from parameters and
// composition, not from one predicate per scenario — the persona compiler
// instantiates these with the constants a given host's disposition implies.
// The library grows by fact DIMENSION (HP, trust, combat, inventory, event),
// not by behavior; behaviors are combinations.
//
// The vocabulary is EXPORTED so the PersonaCompiler can live in any package
// (in-process today, mesa later) and author tables from it. It is the only
// sanctioned way to build a Predicate — never hand-write a closure into a Rule.

// --- vitals -----------------------------------------------------------------

// HPBelow holds when current HP is below frac of max (0..1).
func HPBelow(frac float64) Predicate {
	return func(f *Facts) bool { return f.HPMax > 0 && f.HPFraction() < frac }
}

// HPAbove holds when current HP is at or above frac of max.
func HPAbove(frac float64) Predicate {
	return func(f *Facts) bool { return f.HPMax > 0 && f.HPFraction() >= frac }
}

// FatigueAbove holds when fatigue exceeds frac (0..1).
func FatigueAbove(frac float64) Predicate {
	return func(f *Facts) bool { return f.Fatigue > frac }
}

// --- combat -----------------------------------------------------------------

// CombatLevelBelow holds when the host's own combat level is below n.
func CombatLevelBelow(n int) Predicate {
	return func(f *Facts) bool { return f.CombatLevel > 0 && f.CombatLevel < n }
}

// TargetStronger holds when the event target's combat level exceeds the host's
// own by at least delta (delta may be 0 for "at least as strong", negative to
// include somewhat weaker targets).
func TargetStronger(delta int) Predicate {
	return func(f *Facts) bool {
		return f.Event.TargetCombatLevel > 0 && f.Event.TargetCombatLevel-f.CombatLevel >= delta
	}
}

// --- inventory --------------------------------------------------------------

// InventoryHas holds when the host carries at least n of item (case-insensitive).
func InventoryHas(item string, n int) Predicate {
	key := strings.ToLower(strings.TrimSpace(item))
	return func(f *Facts) bool { return f.InvCounts[key] >= n }
}

// InventoryFull holds when there are no free inventory slots.
func InventoryFull() Predicate {
	return func(f *Facts) bool { return f.InvFree == 0 }
}

// --- event context ----------------------------------------------------------

// OnAction holds when the action being gated matches name (case-insensitive).
func OnAction(name string) Predicate {
	n := strings.ToLower(name)
	return func(f *Facts) bool { return strings.ToLower(f.Event.Action) == n }
}

// OnEvent holds when the triggering bus event kind matches kind.
func OnEvent(kind string) Predicate {
	k := strings.ToLower(kind)
	return func(f *Facts) bool { return strings.ToLower(f.Event.BusKind) == k }
}

// --- relationship / social --------------------------------------------------

// TrustBelow holds when there is a counterparty whose trust is below t.
func TrustBelow(t float64) Predicate {
	return func(f *Facts) bool { return f.Counter != nil && f.Counter.Trust < t }
}

// TrustAtLeast holds when there is a counterparty whose trust is >= t.
func TrustAtLeast(t float64) Predicate {
	return func(f *Facts) bool { return f.Counter != nil && f.Counter.Trust >= t }
}

// IsStranger holds when there is no counterparty or it has never been met.
func IsStranger() Predicate {
	return func(f *Facts) bool { return f.Counter == nil || f.Counter.Familiar == 0 }
}

// HasTag holds when the counterparty carries the given tag.
func HasTag(tag string) Predicate {
	return func(f *Facts) bool { return f.Counter != nil && f.Counter.hasTag(tag) }
}

// --- affect -----------------------------------------------------------------

// StressAbove holds when affective stress exceeds v (0..1).
func StressAbove(v float64) Predicate {
	return func(f *Facts) bool { return f.Affect.Stress > v }
}

// ConfidenceBelow holds when affective confidence is below v (0..1).
func ConfidenceBelow(v float64) Predicate {
	return func(f *Facts) bool { return f.Affect.Confidence < v }
}

// --- combinators ------------------------------------------------------------

// All holds when every sub-predicate holds (logical AND; vacuously true).
func All(ps ...Predicate) Predicate {
	return func(f *Facts) bool {
		for _, p := range ps {
			if !p(f) {
				return false
			}
		}
		return true
	}
}

// Any holds when at least one sub-predicate holds (logical OR; vacuously false).
func Any(ps ...Predicate) Predicate {
	return func(f *Facts) bool {
		for _, p := range ps {
			if p(f) {
				return true
			}
		}
		return false
	}
}

// Not inverts a predicate.
func Not(p Predicate) Predicate {
	return func(f *Facts) bool { return !p(f) }
}
