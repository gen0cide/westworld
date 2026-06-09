package limbic

import (
	"encoding/json"
	"testing"
)

// TestEntryBackCompatUnmarshal proves an OLD persisted row (alpha/beta/encounters
// only, no axis fields) still loads cleanly: the new accumulators default to 0 =
// neutral, the Rel reads neutral on both new axes, and — critically — the Grade is
// IDENTICAL to the pre-3b trust-only grade. This is the backward-compatible
// persistence invariant (a).
func TestEntryBackCompatUnmarshal(t *testing.T) {
	// A pre-3b row exactly as it would have been serialized: no affinity/grievance.
	const old = `{"name":"alex","alpha":9,"beta":1,"encounters":4,"tags":["trader"]}`
	var e Entry
	if err := json.Unmarshal([]byte(old), &e); err != nil {
		t.Fatalf("old-format unmarshal: %v", err)
	}
	if e.AffinitySum != 0 || e.GrievanceSum != 0 {
		t.Fatalf("old row should default axes to 0, got aff=%v gri=%v", e.AffinitySum, e.GrievanceSum)
	}
	r := e.view()
	if r.Affinity != 0 || r.Grievance != 0 {
		t.Fatalf("old row Rel should read neutral axes, got aff=%v gri=%v", r.Affinity, r.Grievance)
	}
	// Grade must match the pure trust-only path (no axis perturbation).
	wantGrade := gradeOf(trustFromBeta(e.Alpha, e.Beta))
	if r.Grade != wantGrade {
		t.Fatalf("old-row grade=%v, want pre-3b trust-only grade=%v", r.Grade, wantGrade)
	}
}

// TestExportImportRoundTripWithAxes proves the new axes survive the full
// persistence path: set axes -> Export -> json.Marshal -> Unmarshal -> Import.
func TestExportImportRoundTripWithAxes(t *testing.T) {
	l := NewLedger()
	l.Observe("bob", true, 3)
	l.Met("bob")
	l.ObserveAffinity("bob", 2.5)
	l.ObserveGrievance("bob", 1.5)
	l.Tag("bob", "sparring-partner")

	raw, err := json.Marshal(l.Export())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rows []Entry
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	l2 := NewLedger()
	l2.Import(rows)

	e := l2.Export()
	if len(e) != 1 {
		t.Fatalf("expected 1 row after round-trip, got %d", len(e))
	}
	if e[0].AffinitySum != 2.5 || e[0].GrievanceSum != 1.5 {
		t.Fatalf("axes not preserved: aff=%v gri=%v, want 2.5/1.5", e[0].AffinitySum, e[0].GrievanceSum)
	}
	if r := l2.Rel("bob"); !contains(r.Tags, "sparring-partner") {
		t.Fatalf("tags not preserved after round-trip: %v", r.Tags)
	}
}

// TestExportImportMixedOldNew proves the REAL migration path: an old-format row
// (loaded with no axes) gains new 3b axis mutations, then survives an export →
// marshal → import round-trip with both the legacy Beta/encounters AND the new
// axes intact. This is the path a long-lived host actually takes.
func TestExportImportMixedOldNew(t *testing.T) {
	var old Entry
	if err := json.Unmarshal([]byte(`{"name":"alex","alpha":9,"beta":1,"encounters":4,"tags":["trader"]}`), &old); err != nil {
		t.Fatalf("old unmarshal: %v", err)
	}
	l := NewLedger()
	l.Import([]Entry{old})

	// New 3b mutations land on the migrated row.
	l.ObserveAffinity("alex", 2.0)
	l.ObserveGrievance("alex", 1.0)

	raw, err := json.Marshal(l.Export())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rows []Entry
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	l2 := NewLedger()
	l2.Import(rows)

	e := l2.Export()
	if len(e) != 1 {
		t.Fatalf("expected 1 row, got %d", len(e))
	}
	if e[0].Alpha != 9 || e[0].Beta != 1 || e[0].Encounters != 4 {
		t.Fatalf("legacy fields lost across migration: %+v", e[0])
	}
	if e[0].AffinitySum != 2.0 || e[0].GrievanceSum != 1.0 {
		t.Fatalf("new axes not preserved across migration round-trip: aff=%v gri=%v", e[0].AffinitySum, e[0].GrievanceSum)
	}
}

func TestSquashBounds(t *testing.T) {
	if got := squash(100, affinityCap); got != 1 {
		t.Fatalf("squash(+big)=%v, want 1", got)
	}
	if got := squash(-100, affinityCap); got != -1 {
		t.Fatalf("squash(-big)=%v, want -1", got)
	}
	if got := squash(0, affinityCap); got != 0 {
		t.Fatalf("squash(0)=%v, want 0", got)
	}
	// ObserveGrievance never goes negative; its read is monotone in [0,1].
	l := NewLedger()
	l.ObserveGrievance("x", -5) // weight<=0 is a no-op
	if r := l.Rel("x"); r.Grievance != 0 {
		t.Fatalf("negative grievance weight should be a no-op, got %v", r.Grievance)
	}
	l.ObserveGrievance("y", 100)
	if r := l.Rel("y"); r.Grievance != 1 {
		t.Fatalf("saturated grievance read=%v, want 1", r.Grievance)
	}
}

// TestGradeMultiAxis covers the grade blend: a standing grudge caps at Wary even
// with high trust; affinity nudges the grade up; and the trust-only path is
// byte-for-byte identical to pre-3b (the regression guard).
func TestGradeMultiAxis(t *testing.T) {
	// High trust but a standing grudge -> capped at Wary.
	if g := gradeOfMultiAxis(0.9, 0.0, 0.6); g != Wary {
		t.Fatalf("trusted+grudge grade=%v, want Wary (grudge cap)", g)
	}
	// Trust-only path must equal gradeOf(trust) exactly (regression guard).
	for _, tr := range []float64{-0.9, -0.3, 0, 0.2, 0.5, 0.9} {
		if g := gradeOfMultiAxis(tr, 0, 0); g != gradeOf(tr) {
			t.Fatalf("trust-only grade for %v = %v, want %v (pre-3b regression)", tr, g, gradeOf(tr))
		}
	}
	// Affinity nudges a borderline-neutral trust up toward friendly.
	base := gradeOfMultiAxis(0.05, 0, 0)
	withWarmth := gradeOfMultiAxis(0.05, 1.0, 0)
	if withWarmth < base {
		t.Fatalf("affinity should not lower the grade: base=%v warm=%v", base, withWarmth)
	}
}

// TestObserveAxesIndependent proves each mutator moves ONLY its own axis and never
// touches the Beta(α,β) trust evidence — the §3.4 multi-axis independence.
func TestObserveAxesIndependent(t *testing.T) {
	l := NewLedger()
	l.ObserveAffinity("a", 2)
	ea := l.Export()[0]
	if ea.AffinitySum != 2 || ea.GrievanceSum != 0 || ea.Alpha != defaultPriorA || ea.Beta != defaultPriorB {
		t.Fatalf("ObserveAffinity leaked: %+v", ea)
	}

	l2 := NewLedger()
	l2.ObserveGrievance("b", 2)
	eb := l2.Export()[0]
	if eb.GrievanceSum != 2 || eb.AffinitySum != 0 || eb.Alpha != defaultPriorA || eb.Beta != defaultPriorB {
		t.Fatalf("ObserveGrievance leaked: %+v", eb)
	}
}

// TestObserveValueTraded proves the value_traded accumulator is monotone, additive,
// clamped to >=0, and touches ONLY its own field (never the Beta(α,β) trust evidence
// or the other axes) — the value-traded wiring (latent-trap close).
func TestObserveValueTraded(t *testing.T) {
	cases := []struct {
		name    string
		weights []float64
		want    float64
	}{
		{"single", []float64{120}, 120},
		{"additive", []float64{50, 70, 5}, 125},
		{"nonpositive-noop", []float64{0, -10, 30}, 30}, // weight<=0 is a no-op
		{"all-noop", []float64{0, -1}, 0},               // never goes negative; row may or may not exist
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			l := NewLedger()
			for _, w := range tc.weights {
				l.ObserveValueTraded("merchant", w)
			}
			rows := l.Export()
			var got Entry
			for _, e := range rows {
				if e.Name == "merchant" {
					got = e
				}
			}
			if got.ValueTraded != tc.want {
				t.Fatalf("ValueTraded=%v, want %v", got.ValueTraded, tc.want)
			}
			// Independence: a value-traded fold must never perturb the trust evidence
			// or the affinity/grievance axes.
			if tc.want > 0 && (got.Alpha != defaultPriorA || got.Beta != defaultPriorB || got.AffinitySum != 0 || got.GrievanceSum != 0) {
				t.Fatalf("ObserveValueTraded leaked into another axis: %+v", got)
			}
		})
	}
}

// TestValueTradedSurvivesPersistence proves the new value_traded field round-trips
// through Export -> json -> Import alongside the legacy + 3b axis fields, and that an
// OLD persisted row (no value_traded) defaults to 0 (backward-compatible, no migration).
func TestValueTradedSurvivesPersistence(t *testing.T) {
	l := NewLedger()
	l.Observe("bob", true, 3)
	l.ObserveAffinity("bob", 2.5)
	l.ObserveValueTraded("bob", 240)

	raw, err := json.Marshal(l.Export())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var rows []Entry
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	l2 := NewLedger()
	l2.Import(rows)
	e := l2.Export()
	if len(e) != 1 || e[0].ValueTraded != 240 || e[0].AffinitySum != 2.5 {
		t.Fatalf("value_traded not preserved across round-trip: %+v", e)
	}

	// An old row with no value_traded defaults to 0.
	var old Entry
	if err := json.Unmarshal([]byte(`{"name":"alex","alpha":9,"beta":1,"encounters":4}`), &old); err != nil {
		t.Fatalf("old unmarshal: %v", err)
	}
	if old.ValueTraded != 0 {
		t.Fatalf("old row should default value_traded to 0, got %v", old.ValueTraded)
	}
}

func contains(ss []string, s string) bool {
	for _, x := range ss {
		if x == s {
			return true
		}
	}
	return false
}
