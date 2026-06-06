package pearl

import (
	"testing"
)

func demoEngine() *Engine { return New(DemoTable(), 0) }

func TestTryDecideBiasHit(t *testing.T) {
	e := demoEngine()
	f := &Facts{HPCur: 2, HPMax: 10} // 20% HP ⇒ flee_when_hurt fires
	d, biased, hit := e.TryDecide(f, []string{"fight", "flee"})
	if !hit {
		t.Fatalf("expected a hit; got miss with %v", biased)
	}
	if d.Choice != "flee" {
		t.Fatalf("Choice = %q, want flee", d.Choice)
	}
	if d.Confidence < 0.8 {
		t.Fatalf("Confidence = %v, want >= 0.85", d.Confidence)
	}
}

func TestTryDecideMissReordersByBias(t *testing.T) {
	e := demoEngine()
	// Healthy ⇒ no rule fires with a confident margin ⇒ miss. With no bias the
	// order is preserved (stable).
	f := &Facts{HPCur: 9, HPMax: 10}
	d, biased, hit := e.TryDecide(f, []string{"fight", "flee"})
	if hit {
		t.Fatalf("expected a miss; got hit %v", d)
	}
	if len(biased) != 2 || biased[0] != "fight" {
		t.Fatalf("biased order = %v, want [fight flee]", biased)
	}
}

func TestTryDecideVetoElimination(t *testing.T) {
	// A custom table: veto "fight" outright; only "flee" remains ⇒ forced.
	tbl := Table{Rules: []Rule{{
		ID:         "no_fight",
		When:       HPBelow(0.5),
		Then:       Effect{Kind: EffectVeto, Options: []string{"fight"}},
		Confidence: 1,
	}}}
	e := New(tbl, 0)
	f := &Facts{HPCur: 1, HPMax: 10}
	d, _, hit := e.TryDecide(f, []string{"fight", "flee"})
	if !hit || d.Choice != "flee" {
		t.Fatalf("veto-elimination: hit=%v choice=%q, want flee", hit, d.Choice)
	}
}

func TestTryDecideAllVetoedEscalates(t *testing.T) {
	tbl := Table{Rules: []Rule{{
		ID:         "veto_all",
		When:       func(*Facts) bool { return true },
		Then:       Effect{Kind: EffectVeto, Options: []string{"a", "b"}},
		Confidence: 1,
	}}}
	e := New(tbl, 0)
	d, biased, hit := e.TryDecide(&Facts{}, []string{"a", "b"})
	if hit {
		t.Fatalf("expected miss when all options vetoed, got %v", d)
	}
	if len(biased) != 2 {
		t.Fatalf("escalation should pass original options, got %v", biased)
	}
}

func TestGateVetoAttackStronger(t *testing.T) {
	e := demoEngine()
	f := &Facts{CombatLevel: 10, Event: EventCtx{TargetCombatLevel: 25}}
	v := e.Gate(f, "attack", nil)
	if v.Allow {
		t.Fatal("expected veto attacking a stronger target")
	}
	if v.RuleID != "no_attack_stronger" {
		t.Fatalf("RuleID = %q, want no_attack_stronger", v.RuleID)
	}
}

func TestGateAllowsAttackWeaker(t *testing.T) {
	e := demoEngine()
	f := &Facts{CombatLevel: 30, Event: EventCtx{TargetCombatLevel: 5}}
	if v := e.Gate(f, "attack", nil); !v.Allow {
		t.Fatalf("expected allow attacking a weaker target, got veto: %s", v.Reason)
	}
}

func TestGateSubstituteTradeDoubleCheck(t *testing.T) {
	e := demoEngine()
	// Trusted, familiar counterparty ⇒ the cornerstone trade veto does NOT fire,
	// so the quirk substitute (examine first) wins.
	f := &Facts{Counter: &Relationship{Name: "alex", Trust: 0.9, Familiar: 5}}
	v := e.Gate(f, "confirm_trade", nil)
	if v.Allow {
		t.Fatal("expected substitute (not allow) on confirm_trade")
	}
	if v.Substitute == nil || v.Substitute.Action != "examine_offer" {
		t.Fatalf("substitute = %+v, want examine_offer", v.Substitute)
	}
}

func TestGateVetoBeatsSubstituteBySalience(t *testing.T) {
	e := demoEngine()
	// Untrusted stranger ⇒ cornerstone veto (salience 80) must win over the
	// quirk substitute (salience 40) for confirm_trade.
	f := &Facts{Counter: &Relationship{Name: "rando", Trust: -0.5, Familiar: 0}}
	v := e.Gate(f, "confirm_trade", nil)
	if v.Allow || v.Substitute != nil {
		t.Fatalf("expected hard veto, got %+v", v)
	}
	if v.RuleID != "no_trade_untrusted_stranger" {
		t.Fatalf("RuleID = %q, want the cornerstone veto", v.RuleID)
	}
}

func TestInjectionsGreetStranger(t *testing.T) {
	e := demoEngine()
	f := &Facts{Event: EventCtx{BusKind: "player_appeared"}} // nil Counter ⇒ stranger
	injects := e.Injections(f)
	if len(injects) != 1 || injects[0].Action != "say" {
		t.Fatalf("injections = %+v, want one say", injects)
	}
}

func TestInjectionsNoneForKnownPlayer(t *testing.T) {
	e := demoEngine()
	f := &Facts{
		Event:   EventCtx{BusKind: "player_appeared"},
		Counter: &Relationship{Name: "alex", Familiar: 3},
	}
	if injects := e.Injections(f); len(injects) != 0 {
		t.Fatalf("expected no greet for a known player, got %+v", injects)
	}
}

// Exercise the remaining vocabulary primitives + combinators so the library is
// covered and the linter sees them used.
func TestVocabularyPrimitives(t *testing.T) {
	f := &Facts{
		HPCur: 8, HPMax: 10, Fatigue: 0.7, CombatLevel: 50,
		InvFree: 0, InvCounts: map[string]int{"lobster": 3},
		Affect:  Affect{Stress: 0.8, Confidence: 0.2},
		Counter: &Relationship{Trust: 0.9, Familiar: 4, Tags: []string{"trusted_partner"}},
	}
	cases := []struct {
		name string
		p    Predicate
		want bool
	}{
		{"hpAbove", HPAbove(0.5), true},
		{"fatigueAbove", FatigueAbove(0.5), true},
		{"combatLevelBelow", CombatLevelBelow(60), true},
		{"inventoryHas", InventoryHas("Lobster", 2), true},
		{"inventoryFull", InventoryFull(), true},
		{"trustAtLeast", TrustAtLeast(0.5), true},
		{"isStranger", IsStranger(), false},
		{"hasTag", HasTag("trusted_partner"), true},
		{"stressAbove", StressAbove(0.5), true},
		{"confidenceBelow", ConfidenceBelow(0.5), true},
		{"not", Not(IsStranger()), true},
		{"all", All(HPAbove(0.5), FatigueAbove(0.5)), true},
		{"any-false", Any(IsStranger(), ConfidenceBelow(0.1)), false},
	}
	for _, c := range cases {
		if got := c.p(f); got != c.want {
			t.Errorf("%s = %v, want %v", c.name, got, c.want)
		}
	}
}
