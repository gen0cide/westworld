package limbic

import (
	"encoding/json"
	"testing"
	"time"
)

// TestLastAtStampedOnInteractions proves every evidence-bearing mutation
// (Met/Observe/axes/value) stamps Entry.LastAt — the relationships table's
// recency column — while Tag (a label attach, not an interaction) does not.
func TestLastAtStampedOnInteractions(t *testing.T) {
	before := time.Now().Add(-time.Second)
	for name, touch := range map[string]func(*Ledger){
		"Met":                func(l *Ledger) { l.Met("bob") },
		"Observe":            func(l *Ledger) { l.Observe("bob", true, 1) },
		"ObserveAffinity":    func(l *Ledger) { l.ObserveAffinity("bob", 0.5) },
		"ObserveGrievance":   func(l *Ledger) { l.ObserveGrievance("bob", 0.5) },
		"ObserveValueTraded": func(l *Ledger) { l.ObserveValueTraded("bob", 3) },
	} {
		l := NewLedger()
		touch(l)
		r := l.Rel("bob")
		if r.LastAt.IsZero() || r.LastAt.Before(before) {
			t.Errorf("%s: LastAt = %v, want ~now", name, r.LastAt)
		}
	}

	l := NewLedger()
	l.Tag("bob", "smith")
	if got := l.Rel("bob").LastAt; !got.IsZero() {
		t.Errorf("Tag alone must not stamp LastAt (not an interaction), got %v", got)
	}
}

// TestLastAtPersistsAndBackCompat: LastAt survives the Export/Import JSON
// round-trip, and an OLD persisted row (no last_at key) reads as a zero time —
// the additive-omitempty invariant the axis sums established.
func TestLastAtPersistsAndBackCompat(t *testing.T) {
	l := NewLedger()
	l.Met("bob")
	want := l.Rel("bob").LastAt

	raw, err := json.Marshal(l.Export())
	if err != nil {
		t.Fatal(err)
	}
	var rows []Entry
	if err := json.Unmarshal(raw, &rows); err != nil {
		t.Fatal(err)
	}
	l2 := NewLedger()
	l2.Import(rows)
	if got := l2.Rel("bob").LastAt; !got.Equal(want) {
		t.Errorf("round-trip LastAt = %v, want %v", got, want)
	}

	const old = `{"name":"alex","alpha":9,"beta":1,"encounters":4}`
	var e Entry
	if err := json.Unmarshal([]byte(old), &e); err != nil {
		t.Fatal(err)
	}
	if e.LastAt != 0 {
		t.Fatalf("old row LastAt = %d, want 0", e.LastAt)
	}
	if got := e.view().LastAt; !got.IsZero() {
		t.Errorf("old-row Rel.LastAt = %v, want zero time", got)
	}
}
