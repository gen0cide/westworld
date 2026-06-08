package knowledge

import "testing"

func TestNoteAndGet(t *testing.T) {
	l := NewLedger()
	if l.Known("Nurmof") {
		t.Fatal("empty ledger should not know Nurmof")
	}
	l.Note("Nurmof", "npc", "sells all pickaxes at the Dwarven Mine", ProvHearsay, 0.6)
	if !l.Known("nurmof") { // normalized lookup
		t.Fatal("ledger should know Nurmof after Note")
	}
	f := l.Get("NURMOF")
	if f.Kind != "npc" || len(f.Beliefs) != 1 {
		t.Fatalf("fact = %+v, want kind=npc with 1 belief", f)
	}
	if f.Beliefs[0].Provenance != ProvHearsay || f.Confidence <= 0.5 {
		t.Fatalf("belief = %+v, want hearsay provenance + confidence>0.5", f.Beliefs[0])
	}
}

func TestNoteRestateReinforcesAndUpgradesProvenance(t *testing.T) {
	l := NewLedger()
	l.Note("Bob's Axes", "shop", "does NOT sell pickaxes", ProvHearsay, 0.6)
	c0 := l.Get("Bob's Axes").Beliefs[0].Confidence()
	// Restate the same claim, now from direct observation (stronger source).
	l.Note("Bob's Axes", "shop", "does NOT sell pickaxes", ProvObserved, 0.9)
	f := l.Get("Bob's Axes")
	if len(f.Beliefs) != 1 {
		t.Fatalf("restating a claim must not duplicate it: %+v", f.Beliefs)
	}
	if f.Beliefs[0].Provenance != ProvObserved {
		t.Fatalf("provenance should upgrade hearsay->observed, got %q", f.Beliefs[0].Provenance)
	}
	if f.Beliefs[0].Confidence() <= c0 {
		t.Fatalf("restating with supporting evidence should raise confidence: %v -> %v", c0, f.Beliefs[0].Confidence())
	}
}

// TestNoteRestateAccumulatesWeighted guards the weighted-evidence restate math:
// restating the same claim at a fixed confidence c must add evidence in that
// proportion (α+=c, β+=1-c), so confidence converges toward c. The old code used
// a hard ±1 vote (α++ whenever c>=0.5), which made repeatedly restating a merely
// 60%-sure belief march confidence toward ~1.0 — confidently wrong.
func TestNoteRestateAccumulatesWeighted(t *testing.T) {
	l := NewLedger()
	for range 20 {
		l.Note("Baraek", "npc", "buys silk", ProvHearsay, 0.6)
	}
	c := l.Get("Baraek").Beliefs[0].Confidence()
	if c < 0.58 || c > 0.62 {
		t.Fatalf("20 restates at 0.6 should converge near 0.6 (α/(α+β)), got %v — the old hard-vote bug would give ~0.98", c)
	}
}

func TestObserveContradictionLowersConfidence(t *testing.T) {
	l := NewLedger()
	l.Note("RuneRune", "item", "requires 30 magic", ProvHearsay, 0.7)
	before := l.Get("RuneRune").Beliefs[0].Confidence()
	// Direct experience contradicts it, repeatedly.
	for range 4 {
		l.Observe("RuneRune", "requires 30 magic", false, 1)
	}
	after := l.Get("RuneRune").Beliefs[0].Confidence()
	if after >= before {
		t.Fatalf("contradicting evidence should lower confidence: %v -> %v", before, after)
	}
}

func TestSeenFamiliarityAndBeliefOrdering(t *testing.T) {
	l := NewLedger()
	l.Seen("Varrock Square", "location")
	l.Seen("Varrock Square", "location")
	l.Note("Varrock Square", "location", "has market stalls", ProvObserved, 0.5)
	l.Note("Varrock Square", "location", "Baraek is here", ProvObserved, 0.95)
	f := l.Get("Varrock Square")
	if f.Familiar != 2 {
		t.Fatalf("familiar=%d, want 2", f.Familiar)
	}
	// Beliefs sorted best-first by confidence; the 0.95 claim should lead.
	if f.Beliefs[0].Claim != "Baraek is here" {
		t.Fatalf("beliefs not confidence-ordered: %+v", f.Beliefs)
	}
	if f.Confidence < 0.9 {
		t.Fatalf("top confidence=%v, want >=0.9", f.Confidence)
	}
}

func TestExportImportRoundTrip(t *testing.T) {
	l := NewLedger()
	l.Note("Nurmof", "npc", "sells pickaxes", ProvObserved, 0.9)
	l.Seen("Nurmof", "npc")
	l.Tag("Nurmof", "supplier")
	snap := l.Export()
	if len(snap) != 1 {
		t.Fatalf("export len=%d, want 1", len(snap))
	}
	l2 := NewLedger()
	if l2.Known("Nurmof") {
		t.Fatal("fresh ledger should not know Nurmof before import")
	}
	l2.Import(snap)
	f := l2.Get("Nurmof")
	if f.Familiar != 1 || len(f.Beliefs) != 1 || f.Confidence < 0.8 {
		t.Fatalf("round-trip lost data: %+v", f)
	}
	hasTag := false
	for _, tg := range f.Tags {
		if tg == "supplier" {
			hasTag = true
		}
	}
	if !hasTag {
		t.Fatalf("round-trip lost tag: %+v", f.Tags)
	}
}
