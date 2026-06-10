package debughttp

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestTailDecisionsMultiChunk proves the backward block reader against a log
// larger than one read block: the tail comes back complete, in chronological
// order, with lines spanning block boundaries reassembled (the carry path) —
// and a torn final line (the live writer mid-append) is skipped, not fatal.
func TestTailDecisionsMultiChunk(t *testing.T) {
	path := filepath.Join(t.TempDir(), "decisions.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	const total = 3000 // ~120B/line * 3000 ≈ 360KB > the 256KB block — forces the carry path
	enc := json.NewEncoder(f)
	for i := 0; i < total; i++ {
		rec := decisionLine{
			At:        base.Add(time.Duration(i) * time.Second),
			Trigger:   "tick",
			Kind:      "act",
			Reasoning: fmt.Sprintf("turn %d: padding so lines straddle block boundaries ............", i),
			Goal:      "smelt bronze",
		}
		if err := enc.Encode(rec); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := f.WriteString(`{"at":"2026-06-10T1`); err != nil { // torn tail, no newline
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := tailDecisions(path, 500, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 500 {
		t.Fatalf("tail length: got %d, want 500", len(got))
	}
	for i, rec := range got {
		want := base.Add(time.Duration(total-500+i) * time.Second)
		if !rec.At.Equal(want) {
			t.Fatalf("record %d out of order: at=%v, want %v", i, rec.At, want)
		}
	}
	if got[499].Reasoning != fmt.Sprintf("turn %d: padding so lines straddle block boundaries ............", total-1) {
		t.Errorf("newest record is not the last written line: %q", got[499].Reasoning)
	}

	// before= pages strictly backwards: records at/after the cursor are excluded.
	cursor := base.Add(100 * time.Second)
	got, err = tailDecisions(path, 500, cursor)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("before-filtered tail: got %d records, want 100", len(got))
	}
	if last := got[len(got)-1].At; !last.Before(cursor) {
		t.Errorf("newest filtered record %v is not strictly before the cursor %v", last, cursor)
	}
}

// TestTailDecisionsOversizedLine pins the carry cap: a single newline-less run
// longer than 4 read blocks (corruption / a runaway record) is treated as
// corrupt and skipped — never accumulated toward a whole-file slurp — while
// the well-formed records around it come back intact and in order.
func TestTailDecisionsOversizedLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), "decisions.jsonl")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2026, 6, 10, 8, 0, 0, 0, time.UTC)
	enc := json.NewEncoder(f)
	write := func(i int) {
		if err := enc.Encode(decisionLine{At: base.Add(time.Duration(i) * time.Second), Trigger: "tick", Kind: "act", Reasoning: fmt.Sprintf("turn %d", i)}); err != nil {
			t.Fatal(err)
		}
	}
	write(0)
	// Valid JSON, but a single line WELL past 4*decisionsChunk (8 blocks: the
	// cap is checked per backward block, so a line must exceed it by more than
	// the head+tail block portions to be guaranteed to trip it at any
	// alignment): must be skipped, not parsed.
	if err := enc.Encode(decisionLine{At: base.Add(time.Second), Trigger: "tick", Kind: "monster", Reasoning: strings.Repeat("x", 8*decisionsChunk)}); err != nil {
		t.Fatal(err)
	}
	write(2)
	write(3)
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	got, err := tailDecisions(path, 10, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d records, want 3 (oversized line skipped): %+v", len(got), got)
	}
	for i, want := range []int{0, 2, 3} {
		if got[i].Reasoning != fmt.Sprintf("turn %d", want) {
			t.Errorf("record %d = %q, want turn %d", i, got[i].Reasoning, want)
		}
		if got[i].Kind == "monster" {
			t.Errorf("oversized record must be skipped as corrupt, got %q", got[i].Reasoning[:32])
		}
	}
}

// TestTailDecisionsMissingFile: a host without a durable stream ("" path) or a
// not-yet-written file reads as an empty stream, never an error.
func TestTailDecisionsMissingFile(t *testing.T) {
	if got, err := tailDecisions("", 10, time.Time{}); err != nil || got != nil {
		t.Errorf(`tailDecisions("") = (%v, %v), want (nil, nil)`, got, err)
	}
	if got, err := tailDecisions(filepath.Join(t.TempDir(), "nope.jsonl"), 10, time.Time{}); err != nil || got != nil {
		t.Errorf("tailDecisions(missing) = (%v, %v), want (nil, nil)", got, err)
	}
}
