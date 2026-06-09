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

// TestImportIsDeepCopy guards M2: Import must clone the Beliefs/Tags slices so a
// later mutation of the caller's snapshot (e.g. an Export() from a still-live
// ledger) cannot reach into imported state.
func TestImportIsDeepCopy(t *testing.T) {
	src := []Entry{{
		Subject:    "Nurmof",
		Kind:       "npc",
		Beliefs:    []Belief{{Claim: "sells pickaxes", Provenance: ProvObserved, Alpha: 9, Beta: 1}},
		Encounters: 1,
		Tags:       []string{"supplier"},
	}}
	l := NewLedger()
	l.Import(src)
	// Mutate the caller's snapshot AFTER import.
	src[0].Beliefs[0].Alpha = 0
	src[0].Beliefs[0].Beta = 99
	src[0].Tags[0] = "mutated"
	f := l.Get("Nurmof")
	if len(f.Beliefs) != 1 || f.Beliefs[0].Alpha != 9 || f.Beliefs[0].Beta != 1 {
		t.Fatalf("Import aliased the caller's Beliefs slice: %+v", f.Beliefs)
	}
	if len(f.Tags) != 1 || f.Tags[0] != "supplier" {
		t.Fatalf("Import aliased the caller's Tags slice: %+v", f.Tags)
	}
}

// TestImportMergesOnSubjectCollision guards L1: two snapshot rows normalizing to
// the same subject key must MERGE (fold beliefs, max evidence + Encounters,
// union tags), not silently last-writer-wins.
func TestImportMergesOnSubjectCollision(t *testing.T) {
	rows := []Entry{
		{
			Subject:    "Nurmof",
			Kind:       "npc",
			Beliefs:    []Belief{{Claim: "sells pickaxes", Provenance: ProvHearsay, Alpha: 3, Beta: 1, At: 10}},
			Encounters: 2,
			LastSeen:   10,
			Tags:       []string{"supplier"},
		},
		{
			Subject:    "NURMOF", // same normalized key
			Kind:       "npc",
			Beliefs:    []Belief{{Claim: "sells pickaxes", Provenance: ProvObserved, Alpha: 5, Beta: 0, At: 20}, {Claim: "is at the Dwarven Mine", Provenance: ProvObserved, Alpha: 4, Beta: 0, At: 20}},
			Encounters: 5,
			LastSeen:   20,
			Tags:       []string{"miner"},
		},
	}
	l := NewLedger()
	l.Import(rows)
	if !l.Known("Nurmof") {
		t.Fatal("merged entry lost")
	}
	f := l.Get("Nurmof")
	if len(f.Beliefs) != 2 {
		t.Fatalf("collision should merge into 2 distinct beliefs, got %d: %+v", len(f.Beliefs), f.Beliefs)
	}
	// The shared claim folds max α/β and upgrades provenance.
	var shared *Belief
	for i := range f.Beliefs {
		if f.Beliefs[i].Claim == "sells pickaxes" {
			shared = &f.Beliefs[i]
		}
	}
	if shared == nil {
		t.Fatalf("shared claim missing after merge: %+v", f.Beliefs)
	}
	if shared.Alpha != 5 || shared.Beta != 1 {
		t.Fatalf("merge should take max α (5) and max β (1), got α=%v β=%v", shared.Alpha, shared.Beta)
	}
	if shared.Provenance != ProvObserved {
		t.Fatalf("merge should upgrade provenance to observed, got %q", shared.Provenance)
	}
	if f.Familiar != 5 {
		t.Fatalf("merge should take max Encounters (5), got %d", f.Familiar)
	}
	if len(f.Tags) != 2 {
		t.Fatalf("merge should union tags (supplier+miner), got %+v", f.Tags)
	}
}

// TestImportMergeAdditive guards M17 (host side): ImportMerge folds a snapshot
// into a NON-EMPTY live ledger additively — it adopts a cron belief the host never
// re-learned, max-merges a shared claim (the host's weaker copy must not regress
// the cron's stronger reconciled evidence), keeps a host-only subject the snapshot
// omits, and never wipes a live row.
func TestImportMergeAdditive(t *testing.T) {
	l := NewLedger()
	// The live host working set the cron must not clobber.
	l.Note("Bob's Axes", "shop", "sells axes", ProvObserved, 0.6) // α=0.6 β=0.4
	l.Note("Lumbridge", "location", "host-only fact", ProvObserved, 0.8)

	// The server-reconciled snapshot: a STRONGER copy of the shared claim + a
	// cron-only belief, plus a cron-only subject.
	rows := []Entry{
		{
			Subject: "Bob's Axes", Kind: "shop", Encounters: 5, LastSeen: 200,
			Beliefs: []Belief{
				{Claim: "sells axes", Provenance: ProvObserved, Alpha: 8, Beta: 1, At: 200}, // cron-reconciled, strong
				{Claim: "out of bronze", Provenance: ProvObserved, Alpha: 1, Beta: 4, At: 190},
			},
			Tags: []string{"reconciled"},
		},
		{
			Subject: "Nurmof", Kind: "npc", Encounters: 3, LastSeen: 200,
			Beliefs: []Belief{{Claim: "sells pickaxes", Provenance: ProvObserved, Alpha: 4, Beta: 0, At: 200}},
		},
	}
	l.ImportMerge(rows)

	bob := l.Get("Bob's Axes")
	// Shared claim takes MAX evidence — the host's weaker copy must not regress it.
	var shared *Belief
	for i := range bob.Beliefs {
		if bob.Beliefs[i].Claim == "sells axes" {
			shared = &bob.Beliefs[i]
		}
	}
	if shared == nil || shared.Alpha != 8 || shared.Beta != 1 {
		t.Fatalf("shared claim not max-merged: %+v", shared)
	}
	// The cron-only belief on a known subject is adopted.
	hasCronBelief := false
	for _, b := range bob.Beliefs {
		if b.Claim == "out of bronze" {
			hasCronBelief = true
		}
	}
	if !hasCronBelief {
		t.Fatalf("cron-only belief not adopted into a known subject: %+v", bob.Beliefs)
	}
	if !sliceHasStr(bob.Tags, "reconciled") {
		t.Fatalf("cron tag not unioned: %v", bob.Tags)
	}
	if bob.Familiar != 5 {
		t.Fatalf("max-Encounters not adopted: %d, want 5", bob.Familiar)
	}
	// The cron-only subject lands so it survives the host's next flush.
	if !l.Known("Nurmof") {
		t.Fatal("ImportMerge did not add the cron-only subject")
	}
	// The host-only subject the snapshot omits is NEVER wiped.
	if !l.Known("Lumbridge") {
		t.Fatal("ImportMerge wiped a host subject absent from the snapshot")
	}
}

// TestImportMergeStaysBounded proves ImportMerge re-enforces the caps so a warm
// re-import can never push the host-side ledger past its bound (host-light).
func TestImportMergeStaysBounded(t *testing.T) {
	l := NewLedger()
	l.SetCaps(4, DefaultMaxBeliefsPerEntry) // tight subject cap
	clock := int64(100)
	l.now = func() int64 { return clock }
	l.Seen("live-recent", "npc") // a recent live subject that must survive
	clock = 50                   // the merged rows look older
	var rows []Entry
	for i := 0; i < 20; i++ {
		rows = append(rows, Entry{Subject: "old-" + string(rune('a'+i)), Kind: "npc", Encounters: 1, LastSeen: 40})
	}
	l.ImportMerge(rows)
	if got := len(l.All()); got > 4 {
		t.Fatalf("ImportMerge left the ledger unbounded: %d subjects, want <= 4", got)
	}
	if !l.Known("live-recent") {
		t.Fatal("ImportMerge prune evicted the most-recent live subject")
	}
}

func sliceHasStr(ss []string, want string) bool {
	for _, s := range ss {
		if s == want {
			return true
		}
	}
	return false
}

// TestSubjectCapEvictsLeastValuable guards H10: the ledger must stay bounded by
// the subject cap, evicting the oldest-LastSeen / lowest-Encounters subjects
// first, deterministically.
func TestSubjectCapEvictsLeastValuable(t *testing.T) {
	l := NewLedger()
	l.SetCaps(2, 0) // keep at most 2 subjects, beliefs unbounded
	clock := int64(100)
	l.now = func() int64 { return clock }
	// Three subjects, each Seen at a distinct (increasing) time.
	for _, s := range []string{"old", "mid", "new"} {
		l.Seen(s, "npc")
		clock++
	}
	if got := l.Prune(); got != 1 {
		t.Fatalf("Prune should evict 1 subject (3 -> cap 2), evicted %d", got)
	}
	if l.Known("old") {
		t.Fatal("Prune should evict the oldest-LastSeen subject 'old'")
	}
	if !l.Known("mid") || !l.Known("new") {
		t.Fatal("Prune should keep the two most-recently-seen subjects")
	}
}

// TestBeliefCapKeepsStrongest guards H10: per-subject beliefs stay bounded,
// dropping the lowest-confidence beliefs first, and the cap is enforced on the
// write path (not only at Prune time).
func TestBeliefCapKeepsStrongest(t *testing.T) {
	l := NewLedger()
	l.SetCaps(0, 3) // subjects unbounded, at most 3 beliefs per subject
	// Five distinct claims at increasing confidence.
	confs := []float64{0.1, 0.2, 0.3, 0.8, 0.9}
	claims := []string{"c0.1", "c0.2", "c0.3", "c0.8", "c0.9"}
	for i := range claims {
		l.Note("Wizard", "npc", claims[i], ProvObserved, confs[i])
	}
	f := l.Get("Wizard")
	if len(f.Beliefs) != 3 {
		t.Fatalf("belief cap should hold at 3 on the write path, got %d: %+v", len(f.Beliefs), f.Beliefs)
	}
	// The three strongest claims must survive; the two weakest dropped.
	kept := map[string]bool{}
	for _, b := range f.Beliefs {
		kept[b.Claim] = true
	}
	for _, want := range []string{"c0.3", "c0.8", "c0.9"} {
		if !kept[want] {
			t.Fatalf("belief cap dropped a strong claim %q: %+v", want, f.Beliefs)
		}
	}
	for _, gone := range []string{"c0.1", "c0.2"} {
		if kept[gone] {
			t.Fatalf("belief cap kept a weak claim %q over a stronger one: %+v", gone, f.Beliefs)
		}
	}
}

// TestPruneIsNoOpWhenUncapped guards that disabling the caps restores unbounded
// behaviour (no surprise eviction).
func TestPruneIsNoOpWhenUncapped(t *testing.T) {
	l := NewLedger()
	l.SetCaps(0, 0)
	for i := 0; i < DefaultMaxSubjects+50; i++ {
		l.Seen("subject-"+string(rune('a'+i%26))+string(rune('a'+i/26)), "npc")
	}
	if got := l.Prune(); got != 0 {
		t.Fatalf("uncapped Prune should evict nothing, evicted %d", got)
	}
}
