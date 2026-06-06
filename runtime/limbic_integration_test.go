package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/pearl"
)

// TestLimbicHandleUpdatesAffect routes mood events and checks the vector moves.
func TestLimbicHandleUpdatesAffect(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.LevelUp{NewLevel: 5})
	_, conf, val := h.affect.Snapshot()
	if conf <= 0.5 || val <= 0 {
		t.Fatalf("level-up should lift confidence/valence: conf=%v val=%v", conf, val)
	}

	h2 := newTestHost()
	h2.limbicHandle(event.Death{})
	s, _, v := h2.affect.Snapshot()
	if s <= 0 || v >= 0 {
		t.Fatalf("death should raise stress / drop valence: stress=%v val=%v", s, v)
	}
}

// TestLimbicHandleUpdatesLedger routes social events and checks familiarity.
func TestLimbicHandleUpdatesLedger(t *testing.T) {
	h := newTestHost()
	h.limbicHandle(event.ChatReceived{Speaker: "alex", Message: "hi"})
	h.limbicHandle(event.ChatReceived{Speaker: "alex", Message: "again"})
	if !h.ledger.Known("alex") {
		t.Fatal("chat should register the speaker in the ledger")
	}
	if r := h.ledger.Rel("alex"); r.Familiar != 2 {
		t.Fatalf("familiar=%d, want 2", r.Familiar)
	}
}

// TestPearlFactsReadsAffect proves the affect vector flows into pearl Facts.
func TestPearlFactsReadsAffect(t *testing.T) {
	h := newTestHost()
	h.affect.OnDeath() // spikes stress
	f := h.pearlFacts(pearl.EventCtx{Action: "decide"})
	if f.Affect.Stress <= 0 {
		t.Fatalf("pearlFacts.Affect.Stress = %v, want > 0 after death", f.Affect.Stress)
	}
}

// TestPearlStressPredicateFiresFromLimbic ties it together: an affect spike from
// the limbic path makes a stress-gated pearl rule fire through the decide seam.
func TestPearlStressPredicateFiresFromLimbic(t *testing.T) {
	h := newTestHost()
	h.Strategist = nil // a miss would surface NOT_IMPLEMENTED
	h.Pearl = pearl.New(pearl.Table{Rules: []pearl.Rule{{
		ID:         "panic_when_stressed",
		When:       pearl.StressAbove(0.3),
		Then:       pearl.Effect{Kind: pearl.EffectBias, Bias: map[string]float64{"flee": 1.0}},
		Confidence: 0.9,
	}}}, 0)
	h.affect.OnDeath() // stress now well above 0.3

	res := runRoutine(t, h, `runtime "1.0"
routine r() { return decide(["fight", "flee"]).val }`)
	if res.Kind != interp.ResultReturned || res.Value.Display() != "flee" {
		t.Fatalf("expected pearl to bias 'flee' from limbic stress; got kind=%v val=%v", res.Kind, res.Value.Display())
	}
}

// TestRelationWithUsesLedger proves relation_with reports the learned trust grade.
func TestRelationWithUsesLedger(t *testing.T) {
	h := newTestHost()
	for range 12 {
		h.ledger.Observe("alex", true, 1)
	}
	h.ledger.Met("alex")
	res := runRoutine(t, h, `runtime "1.0"
routine r() { return relation_with("alex").val }`)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind=%v err=%v", res.Kind, res.Err)
	}
	if got := res.Value.Display(); got != "trusted" {
		t.Fatalf("relation_with(alex) = %q, want trusted", got)
	}
}
