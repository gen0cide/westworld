package limbic

import (
	"math"
	"testing"
	"time"
)

type clock struct{ t time.Time }

func (c *clock) now() time.Time          { return c.t }
func (c *clock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestAffect(c *clock) *Affect {
	a := NewAffect(0, 0.5, 0, time.Minute)
	a.now = c.now
	a.last = c.t
	return a
}

func approx(a, b float64) bool { return math.Abs(a-b) < 1e-6 }

func TestAffectNudgeAndClamp(t *testing.T) {
	c := &clock{t: time.Unix(1000, 0)}
	a := newTestAffect(c)
	a.Nudge(0.5, 0.3, -0.4)
	s, conf, v := a.Snapshot()
	if !approx(s, 0.5) || !approx(conf, 0.8) || !approx(v, -0.4) {
		t.Fatalf("after nudge: stress=%v conf=%v val=%v", s, conf, v)
	}
	// Clamp: push past bounds.
	a.Nudge(1.0, 1.0, -1.0)
	s, conf, v = a.Snapshot()
	if s != 1 || conf != 1 || v != -1 {
		t.Fatalf("clamp failed: %v %v %v", s, conf, v)
	}
}

func TestAffectDecaysTowardBaseline(t *testing.T) {
	c := &clock{t: time.Unix(1000, 0)}
	a := newTestAffect(c)
	a.Nudge(0.8, 0, 0)     // stress 0.8, baseline 0
	c.advance(time.Minute) // one half-life ⇒ halfway back to 0
	s, _, _ := a.Snapshot()
	if !approx(s, 0.4) {
		t.Fatalf("after one half-life stress=%v, want ~0.4", s)
	}
	c.advance(time.Minute) // another half-life ⇒ 0.2
	s, _, _ = a.Snapshot()
	if !approx(s, 0.2) {
		t.Fatalf("after two half-lives stress=%v, want ~0.2", s)
	}
}

func TestAffectEventHelpers(t *testing.T) {
	c := &clock{t: time.Unix(1000, 0)}
	a := newTestAffect(c)
	a.OnLevelUp()
	_, conf, v := a.Snapshot()
	if conf <= 0.5 || v <= 0 {
		t.Fatalf("level-up should raise confidence/valence: conf=%v val=%v", conf, v)
	}
	a2 := newTestAffect(c)
	a2.OnDeath()
	s, _, v2 := a2.Snapshot()
	if s <= 0 || v2 >= 0 {
		t.Fatalf("death should raise stress, drop valence: stress=%v val=%v", s, v2)
	}
}

func TestLedgerTrustFromEvidence(t *testing.T) {
	l := NewLedger()
	// Never met ⇒ neutral.
	if r := l.Rel("alex"); r.Trust != 0 || r.Grade != Neutral {
		t.Fatalf("unmet party: %+v, want neutral 0", r)
	}
	// Accrue positive evidence ⇒ trust rises, grade improves.
	for range 10 {
		l.Observe("alex", true, 1)
	}
	r := l.Rel("alex")
	if r.Trust <= 0 {
		t.Fatalf("after 10 good interactions trust=%v, want > 0", r.Trust)
	}
	if r.Grade != Friendly && r.Grade != Trusted {
		t.Fatalf("grade=%v, want friendly/trusted", r.Grade)
	}
	// Negative evidence pulls it back down.
	for range 30 {
		l.Observe("alex", false, 1)
	}
	if r := l.Rel("alex"); r.Trust >= 0 {
		t.Fatalf("after many bad interactions trust=%v, want < 0", r.Trust)
	}
}

func TestLedgerFamiliarityAndTags(t *testing.T) {
	l := NewLedger()
	l.Met("bob")
	l.Met("bob")
	l.Tag("bob", "rival")
	l.Tag("bob", "rival") // dedup
	r := l.Rel("bob")
	if r.Familiar != 2 {
		t.Fatalf("familiar=%d, want 2", r.Familiar)
	}
	if len(r.Tags) != 1 || r.Tags[0] != "rival" {
		t.Fatalf("tags=%v, want [rival]", r.Tags)
	}
}

func TestLedgerNameNormalization(t *testing.T) {
	l := NewLedger()
	l.Observe("Alex", true, 5)
	if !l.Known("alex") || !l.Known(" ALEX ") {
		t.Fatal("name lookup should be case/space-insensitive")
	}
}

func TestLedgerExportImportRoundTrip(t *testing.T) {
	l := NewLedger()
	l.Observe("carl", true, 3)
	l.Met("carl")
	l.Tag("carl", "trusted_partner")
	snap := l.Export()

	l2 := NewLedger()
	l2.Import(snap)
	a, b := l.Rel("carl"), l2.Rel("carl")
	if !approx(a.Trust, b.Trust) || a.Familiar != b.Familiar || len(b.Tags) != 1 {
		t.Fatalf("round-trip mismatch: %+v vs %+v", a, b)
	}
}
