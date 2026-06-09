package limbic

import (
	"math"
	"sync"
	"time"
)

// Default affect dynamics.
const (
	defaultHalfLife = 5 * time.Minute // mood returns halfway to baseline every 5m
)

// Affect is the host's mood vector: stress and confidence in [0,1], valence in
// [-1,1]. Events nudge it; reads decay it toward a per-host baseline. Safe for
// concurrent use.
type Affect struct {
	mu                          sync.Mutex
	stress, confidence, valence float64
	baseS, baseC, baseV         float64
	halfLife                    time.Duration
	last                        time.Time
	now                         func() time.Time
}

// NewAffect returns an affect vector starting at the given baseline (which it
// also decays back toward). halfLife <= 0 uses the default. Typical baseline:
// stress 0, confidence 0.5, valence 0.
func NewAffect(baseStress, baseConfidence, baseValence float64, halfLife time.Duration) *Affect {
	if halfLife <= 0 {
		halfLife = defaultHalfLife
	}
	a := &Affect{
		stress: baseStress, confidence: baseConfidence, valence: baseValence,
		baseS: baseStress, baseC: baseConfidence, baseV: baseValence,
		halfLife: halfLife,
		now:      time.Now,
	}
	a.last = a.now()
	return a
}

// SetBaseline re-baselines the affect vector in place under the lock, leaving
// the *Affect pointer (and thus every concurrent reader) untouched. Semantics
// match constructing a fresh NewAffect with the same values: it sets both the
// decay target (baseS/baseC/baseV) AND the current vector, and resets the decay
// clock. Used to re-baseline at runtime AFTER readers/the limbic goroutine are
// live — reassigning h.affect would race the field; mutating in place does not.
func (a *Affect) SetBaseline(stress, confidence, valence float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.baseS, a.baseC, a.baseV = stress, confidence, valence
	a.stress, a.confidence, a.valence = stress, confidence, valence
	a.last = a.now()
}

// Snapshot decays the vector toward baseline by the elapsed time, then returns
// the current (stress, confidence, valence).
func (a *Affect) Snapshot() (stress, confidence, valence float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.decayLocked()
	return a.stress, a.confidence, a.valence
}

// Nudge decays then applies additive deltas, clamping to range. Positive
// dStress raises stress, etc.
func (a *Affect) Nudge(dStress, dConfidence, dValence float64) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.decayLocked()
	a.stress = clamp01(a.stress + dStress)
	a.confidence = clamp01(a.confidence + dConfidence)
	a.valence = clampSigned(a.valence + dValence)
}

// decayLocked moves each component a fraction of the way back to baseline based
// on elapsed time and the half-life. Caller holds the lock.
func (a *Affect) decayLocked() {
	t := a.now()
	elapsed := t.Sub(a.last)
	a.last = t
	if elapsed <= 0 {
		return
	}
	// keep = 0.5 ^ (elapsed / halfLife): the share of the deviation retained.
	keep := math.Exp2(-float64(elapsed) / float64(a.halfLife))
	a.stress = a.baseS + (a.stress-a.baseS)*keep
	a.confidence = a.baseC + (a.confidence-a.baseC)*keep
	a.valence = a.baseV + (a.valence-a.baseV)*keep
}

// --- event-shaped helpers (the Limbic loop calls these) ---------------------

// OnXPGain: progress feels good — a small valence/confidence lift scaled by the
// (log of the) xp amount.
func (a *Affect) OnXPGain(xp int) {
	if xp <= 0 {
		return
	}
	mag := 0.02 + 0.02*math.Log1p(float64(xp))
	a.Nudge(-0.01, mag*0.5, mag)
}

// OnLevelUp: a clear confidence/valence boost.
func (a *Affect) OnLevelUp() { a.Nudge(-0.05, 0.15, 0.15) }

// OnHurt: taking damage raises stress and dents valence, scaled by the fraction
// of max HP lost (0..1).
func (a *Affect) OnHurt(fracLost float64) {
	if fracLost <= 0 {
		return
	}
	a.Nudge(0.4*fracLost, -0.2*fracLost, -0.3*fracLost)
}

// OnDeath: a large negative shock.
func (a *Affect) OnDeath() { a.Nudge(0.6, -0.4, -0.7) }

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func clampSigned(v float64) float64 {
	if v < -1 {
		return -1
	}
	if v > 1 {
		return 1
	}
	return v
}
