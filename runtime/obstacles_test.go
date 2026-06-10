package runtime

import "testing"

// TestClassifyRefusal pins the VERIFIED server prose → refusal-kind mapping
// (DoorAction.java, BorderGuard.java, quest plugins).
func TestClassifyRefusal(t *testing.T) {
	cases := []struct {
		msg  string
		want RefusalKind
	}{
		{"The door is locked", RefusalLocked},
		{"the door is locked.", RefusalLocked},
		{"You must pay a toll of 10 gold coins to pass", RefusalToll},
		{"You need to talk to the border guard", RefusalToll},
		{"you need a mining level of 60 to enter", RefusalRequirement},
		{"Members only beyond this point", RefusalRequirement},
		{"speak to the controls guide before going through this door", RefusalTutorial},
		{"Nothing interesting happens", RefusalWrongChannel},
		{"", RefusalUnknown},
		{"some unrecognized text", RefusalUnknown},
	}
	for _, c := range cases {
		got, _ := classifyRefusal(c.msg)
		if got != c.want {
			t.Errorf("classifyRefusal(%q) = %v, want %v", c.msg, got, c.want)
		}
	}
}

// TestBlockedEdgesLedger covers note/lookup/skip semantics.
func TestBlockedEdgesLedger(t *testing.T) {
	b := newBlockedEdges()
	if _, ok := b.Blocked(1, 2, 3); ok {
		t.Fatal("empty ledger must not report blocked")
	}
	b.Note(BlockedEdge{X: 1, Y: 2, Dir: 3, Kind: RefusalLocked, Prose: "The door is locked"})
	e, ok := b.Blocked(1, 2, 3)
	if !ok || e.Kind != RefusalLocked {
		t.Fatalf("ledgered edge not found: %+v ok=%v", e, ok)
	}
	if _, ok := b.Blocked(9, 9, 0); ok {
		t.Fatal("different edge must not be blocked")
	}
	if got := len(b.All()); got != 1 {
		t.Fatalf("All() = %d entries, want 1", got)
	}
}
