package memory

import (
	"testing"
	"time"
)

// fixedClock returns a deterministic, monotonically-advancing clock so episode
// timestamps are stable across a test run.
func fixedClock() func() time.Time {
	t := time.Unix(1_700_000_000, 0)
	return func() time.Time {
		t = t.Add(time.Second)
		return t
	}
}

func TestJournalAppendAndRecent(t *testing.T) {
	j := NewJournal(0)
	j.now = fixedClock()
	j.Append("milestone", "Reached level 2 in attack.", 0.7, "")
	j.Append("kill", "Defeated a rat.", 0.45, "rat")
	if j.Len() != 2 {
		t.Fatalf("Len = %d, want 2", j.Len())
	}
	rec := j.Recent(10)
	if len(rec) != 2 || rec[0].Text != "Reached level 2 in attack." || rec[1].Text != "Defeated a rat." {
		t.Fatalf("Recent order wrong: %+v", rec)
	}
	if rec[0].Seq != 1 || rec[1].Seq != 2 {
		t.Fatalf("Seq not monotonic: %d, %d", rec[0].Seq, rec[1].Seq)
	}
}

func TestJournalDedupRefreshesRecency(t *testing.T) {
	j := NewJournal(0)
	j.now = fixedClock()
	j.Append("kill", "Defeated a rat.", 0.45, "rat")
	j.Append("kill", "Defeated a rat.", 0.45, "rat") // identical → collapse
	j.Append("kill", "Defeated a rat.", 0.45, "rat")
	if j.Len() != 1 {
		t.Fatalf("dedup failed: Len = %d, want 1", j.Len())
	}
	// A different kill is a distinct episode.
	j.Append("kill", "Defeated a goblin.", 0.45, "goblin")
	if j.Len() != 2 {
		t.Fatalf("distinct kill not recorded: Len = %d, want 2", j.Len())
	}
}

func TestJournalCapKeepsMostRecent(t *testing.T) {
	j := NewJournal(3)
	j.now = fixedClock()
	for i := 0; i < 10; i++ {
		j.Append("kill", "kill #"+string(rune('a'+i)), 0.4, "")
	}
	if j.Len() != 3 {
		t.Fatalf("cap not enforced: Len = %d, want 3", j.Len())
	}
	rec := j.Recent(10)
	if rec[len(rec)-1].Text != "kill #j" {
		t.Fatalf("newest not retained: %q", rec[len(rec)-1].Text)
	}
}

func TestJournalSalientBlendsImportanceAndRecency(t *testing.T) {
	j := NewJournal(0)
	j.now = fixedClock()
	// An important early event...
	j.Append("death", "Died and respawned.", 0.9, "")
	// ...then a long tail of low-value, recent kills.
	for i := 0; i < 10; i++ {
		j.Append("kill", "Defeated enemy "+string(rune('a'+i)), 0.4, "")
	}
	sal := j.Salient(3)
	if len(sal) != 3 {
		t.Fatalf("Salient returned %d, want 3", len(sal))
	}
	// The high-importance death must survive recall despite being oldest.
	var hasDeath bool
	for _, e := range sal {
		if e.Kind == "death" {
			hasDeath = true
		}
	}
	if !hasDeath {
		t.Fatalf("important old episode dropped from Salient: %+v", sal)
	}
	// Output is chronological (ascending Seq).
	for i := 1; i < len(sal); i++ {
		if sal[i-1].Seq > sal[i].Seq {
			t.Fatalf("Salient not chronological: %+v", sal)
		}
	}
}

func TestJournalObjective(t *testing.T) {
	j := NewJournal(0)
	j.now = fixedClock()
	if !j.SetObjective("become a master chef") {
		t.Fatal("first SetObjective should report changed")
	}
	if j.SetObjective("become a master chef") {
		t.Fatal("unchanged SetObjective should report false")
	}
	if j.SetObjective("") {
		t.Fatal("empty SetObjective should report false")
	}
	obj, at := j.Objective()
	if obj != "become a master chef" || at == 0 {
		t.Fatalf("Objective = %q @ %d", obj, at)
	}
}

func TestJournalExportImportRoundTrip(t *testing.T) {
	j := NewJournal(0)
	j.now = fixedClock()
	j.SetObjective("explore lumbridge")
	j.Append("milestone", "Reached level 3 in cooking.", 0.7, "")
	j.Append("death", "Died and respawned.", 0.9, "")
	snap := j.Export()

	restored := NewJournal(0)
	restored.Import(snap)
	if got, _ := restored.Objective(); got != "explore lumbridge" {
		t.Fatalf("objective not restored: %q", got)
	}
	if restored.Len() != 2 {
		t.Fatalf("episodes not restored: Len = %d, want 2", restored.Len())
	}
	// A freshly-appended episode must get a Seq past the imported ones.
	ep := restored.Append("kill", "Defeated a chicken.", 0.45, "chicken")
	if ep.Seq <= snap.Seq {
		t.Fatalf("new Seq %d collides with imported (snap.Seq=%d)", ep.Seq, snap.Seq)
	}
}
