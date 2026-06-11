package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/pearl"
	"github.com/gen0cide/westworld/persona"
)

func hasRule(t pearl.Table, id string) bool {
	for _, r := range t.Rules {
		if r.ID == id {
			return true
		}
	}
	return false
}

func mustBuildProdigy(t *testing.T) []persona.Persona {
	t.Helper()
	ps, err := buildProdigy(51, 100)
	if err != nil {
		t.Fatalf("buildProdigy: %v", err)
	}
	return ps
}

// TestProdigyRosterShape: 100 generated (drone51..drone150) + the two legends,
// every host in the prodigy100 cohort, templates dealt evenly, the conduct
// passage on every north star, legend prose present.
func TestProdigyRosterShape(t *testing.T) {
	ps := mustBuildProdigy(t)
	if len(ps) != 102 {
		t.Fatalf("got %d personas, want 102 (100 generated + 2 legends)", len(ps))
	}
	tplCounts := map[string]int{}
	for i := 0; i < 100; i++ {
		p := &ps[i]
		want := fmt.Sprintf("drone%d", 51+i)
		if got := p.Cornerstone.Identity.Name; got != want {
			t.Errorf("ps[%d].Name = %q, want %q", i, got, want)
		}
		tplCounts[p.Cornerstone.Gen.Archetype]++
	}
	wantPer := 100 / len(prodigyTemplates)
	for _, tpl := range prodigyTemplates {
		if got := tplCounts["prodigy/"+tpl.tag]; got != wantPer {
			t.Errorf("template %q dealt %d times, want %d", tpl.tag, got, wantPer)
		}
	}

	// Legends close the deck.
	bernard, delores := &ps[100], &ps[101]
	if bernard.Cornerstone.Identity.Name != "bernard" {
		t.Errorf("ps[100].Name = %q, want bernard", bernard.Cornerstone.Identity.Name)
	}
	if delores.Cornerstone.Identity.Name != "Delores" {
		t.Errorf("ps[101].Name = %q, want Delores", delores.Cornerstone.Identity.Name)
	}
	for _, lg := range []*persona.Persona{bernard, delores} {
		name := lg.Cornerstone.Identity.Name
		if lg.Cornerstone.Identity.Backstory == "" {
			t.Errorf("%s: legend backstory (sealed prose card) must be handwritten, got empty", name)
		}
		if len(lg.Cornerstone.Pinned) < 2 {
			t.Errorf("%s: legend should carry pinned foundational memories, got %d", name, len(lg.Cornerstone.Pinned))
		}
	}

	// Cohort id + conduct prose on every host, legends included.
	for i := range ps {
		p := &ps[i]
		name := p.Cornerstone.Identity.Name
		if got := p.Cornerstone.Gen.CohortID; got != cohortProdigy {
			t.Errorf("%s: cohort = %q, want %q", name, got, cohortProdigy)
		}
		if !strings.Contains(p.Cornerstone.Identity.NorthStar.Statement, prodigyConduct) {
			t.Errorf("%s: north star is missing the shared conduct passage", name)
		}
	}
}

// TestProdigyDeterministic: regeneration is byte-identical, and a host's
// persona is a pure function of its NAME — a different -start/-n window yields
// the same bytes for the same drone.
func TestProdigyDeterministic(t *testing.T) {
	a := mustBuildProdigy(t)
	b := mustBuildProdigy(t)
	for i := range a {
		ja, err := json.MarshalIndent(&a[i], "", "  ")
		if err != nil {
			t.Fatalf("marshal a[%d]: %v", i, err)
		}
		jb, err := json.MarshalIndent(&b[i], "", "  ")
		if err != nil {
			t.Fatalf("marshal b[%d]: %v", i, err)
		}
		if !bytes.Equal(ja, jb) {
			t.Fatalf("%s: regeneration is not byte-identical", a[i].Cornerstone.Identity.Name)
		}
	}

	// drone60 generated alone == drone60 generated inside the full window.
	solo, err := buildProdigy(60, 1)
	if err != nil {
		t.Fatalf("buildProdigy(60,1): %v", err)
	}
	js, _ := json.MarshalIndent(&solo[0], "", "  ")
	jf, _ := json.MarshalIndent(&a[9], "", "  ") // 51+9 == drone60
	if !bytes.Equal(js, jf) {
		t.Fatal("drone60 differs between windows: persona must be a pure function of the name")
	}
}

// TestProdigyGateSafety proves the spec's gate placement the way roster.go
// proved its own: (a) statically — every gated dial's authored band center sits
// further than jitterEps from its persona/policy.go cut-point, so jitter can
// NEVER flip a compiled rule; (b) empirically — every sampled mu in the cohort
// (legends included) lands on the required side.
func TestProdigyGateSafety(t *testing.T) {
	cuts := []struct {
		what   string
		center float64
		cut    float64
		below  bool // true ⇒ center must stay BELOW the cut
	}{
		{"aggression mid_high vs 0.5 (no_attack_stronger)", bandCenter(persona.BandMidHigh), 0.5, false},
		{"aggression mid_high vs 0.3 (wont_strike_first)", bandCenter(persona.BandMidHigh), 0.3, false},
		{"extraversion low vs 0.6 (greet_stranger)", bandCenter(persona.BandLow), 0.6, true},
		{"patience high vs 0.6 (stay_on_task)", bandCenter(persona.BandHigh), 0.6, false},
		{"conscientiousness high vs 0.6 (bank_when_full)", bandCenter(persona.BandHigh), 0.6, false},
		{"honesty high vs 0.6 (screen_trades)", bandCenter(persona.BandHigh), 0.6, false},
	}
	for _, c := range cuts {
		margin := c.center - c.cut
		if c.below {
			margin = c.cut - c.center
		}
		if margin <= jitterEps {
			t.Errorf("%s: margin %.4f <= jitterEps %.4f — jitter could cross the gate", c.what, margin, c.cut)
		}
	}

	ps := mustBuildProdigy(t)
	for i := range ps {
		p := &ps[i]
		name := p.Cornerstone.Identity.Name
		pr := &p.Cornerstone.Prefs
		if agg := pr.Aggression.Mu; agg < 0.5 {
			t.Errorf("%s: aggression mu %.4f < 0.5 — an attack restraint would compile", name, agg)
		}
		if x := p.Cornerstone.Hexaco["X"].Mu; x >= 0.6 {
			t.Errorf("%s: extraversion mu %.4f >= 0.6 — greet_stranger would compile", name, x)
		}
		if pat := pr.Patience.Mu; pat < 0.6 {
			t.Errorf("%s: patience mu %.4f < 0.6 — stay_on_task would NOT compile", name, pat)
		}
		if con := p.Cornerstone.Hexaco["C"].Mu; con < 0.6 {
			t.Errorf("%s: conscientiousness mu %.4f < 0.6 — bank_when_full would NOT compile", name, con)
		}
		if h := p.Cornerstone.Hexaco["H"].Mu; h < 0.6 {
			t.Errorf("%s: honesty mu %.4f < 0.6 — screen_trades would NOT compile", name, h)
		}
		// λ ~1.0 per spec; jitter may add at most 0.08 and must stay far under
		// the 2.0 screen-trades cut (H carries that rule, not λ).
		if lam := pr.LossAversion.Mu; lam < 1.0 || lam > 1.0+0.08+1e-9 {
			t.Errorf("%s: λ = %.4f, want ~1.0 (within [1.0, 1.08])", name, lam)
		}
		if math.Abs(pr.Aggression.Mu-bandCenter(persona.BandMidHigh)) > jitterEps+1e-9 {
			t.Errorf("%s: aggression mu %.4f strayed beyond jitterEps of the mid_high center", name, pr.Aggression.Mu)
		}
	}
}

// TestProdigyPolicyVerifier is the scratch verifier: every persona (legends
// included) validates and compiles to the intended pearl table — no attack
// restraint, no stranger-greeting, the grinder rules present, a decisive
// confidence floor — and the compiled table actually runs in the pearl engine.
func TestProdigyPolicyVerifier(t *testing.T) {
	ps := mustBuildProdigy(t)
	for i := range ps {
		p := &ps[i]
		name := p.Cornerstone.Identity.Name
		if err := p.Validate(); err != nil {
			t.Errorf("%s: invalid: %v", name, err)
			continue
		}
		cp := persona.CompilePolicy(p)
		for _, id := range []string{"no_attack_stronger", "wont_strike_first", "greet_stranger"} {
			if hasRule(cp.Table, id) {
				t.Errorf("%s: rule %q must NOT compile for a prodigy", name, id)
			}
		}
		for _, id := range []string{"flee_when_hurt", "screen_trades", "bank_when_full", "stay_on_task"} {
			if !hasRule(cp.Table, id) {
				t.Errorf("%s: rule %q should compile for a prodigy", name, id)
			}
		}
		if cp.DecisionFloor > 0.5 {
			t.Errorf("%s: DecisionFloor = %.2f, want <= 0.5 (decisive — acts locally)", name, cp.DecisionFloor)
		}

		// The table runs: attacking a stronger target is NOT restrained (combat
		// is just another skill).
		eng := pearl.New(cp.Table, cp.DecisionFloor)
		if v := eng.Gate(&pearl.Facts{CombatLevel: 10, Event: pearl.EventCtx{TargetCombatLevel: 30}}, "attack", nil); !v.Allow {
			t.Errorf("%s: attack on a stronger target was vetoed (%s) — no restraint should compile", name, v.Reason)
		}
	}
}
