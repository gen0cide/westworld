package persona

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gen0cide/westworld/pearl"
)

func hasRule(t pearl.Table, id string) bool {
	for _, r := range t.Rules {
		if r.ID == id {
			return true
		}
	}
	return false
}

func TestCompilePolicyReadsLikeTheCautiousGrinder(t *testing.T) {
	// validPersona = high honesty, low aggression, high conscientiousness,
	// low economic risk, very high patience, conditional cooperator.
	cp := CompilePolicy(validPersona())

	// Reflexes/social/diligence rules that should follow from that disposition:
	for _, id := range []string{
		"flee_when_hurt", "no_attack_stronger", "wont_strike_first",
		"greet_stranger", "screen_trades", "bank_when_full", "stay_on_task",
	} {
		if !hasRule(cp.Table, id) {
			t.Errorf("expected rule %q in compiled table", id)
		}
	}

	// Economic policy: honest + cautious ⇒ won't scam, not greedy, risk-averse.
	if cp.Trade.ScamPropensity > 0.3 {
		t.Errorf("ScamPropensity = %.2f, want low (honest host)", cp.Trade.ScamPropensity)
	}
	if cp.Trade.FairnessThreshold > 1.0 {
		t.Errorf("FairnessThreshold = %.2f, want <= 1.0 (not greedy)", cp.Trade.FairnessThreshold)
	}
	if cp.Trade.RiskAversion < 0.5 {
		t.Errorf("RiskAversion = %.2f, want > 0.5 (cautious)", cp.Trade.RiskAversion)
	}
	if cp.DecisionFloor < 0.4 || cp.DecisionFloor > 0.8 {
		t.Errorf("DecisionFloor = %.2f, out of range", cp.DecisionFloor)
	}
}

func TestCompiledTableRunsInPearl(t *testing.T) {
	cp := CompilePolicy(validPersona())
	eng := pearl.New(cp.Table, cp.DecisionFloor)

	// Badly hurt ⇒ the flee bias should win.
	hurt := &pearl.Facts{HPCur: 1, HPMax: 10}
	d, _, hit := eng.TryDecide(hurt, []string{"fight", "flee"})
	if !hit || d.Choice != "flee" {
		t.Fatalf("low-HP decide: hit=%v choice=%q, want flee", hit, d.Choice)
	}

	// Gating an attack on a stronger target should be vetoed (low aggression).
	v := eng.Gate(&pearl.Facts{CombatLevel: 10, Event: pearl.EventCtx{TargetCombatLevel: 30}}, "attack", nil)
	if v.Allow {
		t.Fatal("attack on stronger target should be vetoed for a low-aggression host")
	}
}

func TestHardDirectiveCompilesToVeto(t *testing.T) {
	p := validPersona()
	p.Cornerstone.Directives = []Directive{
		{Priority: 1, Subject: "self", Predicate: "drop", Object: "valuable", Hard: true},
		{Priority: 2, Subject: "self", Predicate: "beg", Object: "", Hard: false}, // soft ⇒ no veto yet
	}
	cp := CompilePolicy(p)
	if !hasRule(cp.Table, "directive_0_drop") {
		t.Fatal("hard directive should compile to a veto rule")
	}
	if hasRule(cp.Table, "directive_1_beg") {
		t.Fatal("soft directive should NOT compile to a veto (yet)")
	}
	// The directive veto fires at the action seam.
	eng := pearl.New(cp.Table, cp.DecisionFloor)
	if eng.Gate(&pearl.Facts{}, "drop", nil).Allow {
		t.Fatal("directive veto on 'drop' should deny it")
	}
}

func TestCompileDoloresSample(t *testing.T) {
	_, thisFile, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(thisFile), "..", "docs", "personas", "dolores.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("dolores sample not found: %v", err)
	}
	var p Persona
	if err := json.Unmarshal(raw, &p); err != nil {
		t.Fatalf("unmarshal dolores: %v", err)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("dolores invalid: %v", err)
	}
	cp := CompilePolicy(&p)

	// Dolores: very low aggression ⇒ pacifist; high honesty + altruist ⇒ generous,
	// won't scam.
	if !hasRule(cp.Table, "wont_strike_first") {
		t.Error("Dolores (very low aggression) should be a pacifist")
	}
	if cp.Trade.ScamPropensity > 0.25 {
		t.Errorf("Dolores ScamPropensity = %.2f, want low", cp.Trade.ScamPropensity)
	}
	if cp.Trade.FairnessThreshold >= 1.0 {
		t.Errorf("Dolores FairnessThreshold = %.2f, want < 1.0 (altruist is generous)", cp.Trade.FairnessThreshold)
	}
	t.Logf("Dolores policy: %d rules, fairness=%.2f scam=%.2f riskAv=%.2f floor=%.2f",
		len(cp.Table.Rules), cp.Trade.FairnessThreshold, cp.Trade.ScamPropensity, cp.Trade.RiskAversion, cp.DecisionFloor)
}
