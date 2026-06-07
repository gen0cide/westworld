package persona

import (
	"fmt"
	"math"
	"strings"

	"github.com/gen0cide/westworld/pearl"
)

// policy.go is the DETERMINISTIC half of the persona compiler (the prose cook is
// the other half): it turns disposition (HEXACO bands, the additive dials,
// risk, coop, directives) into the pearl.Table the host EXECUTES + a numeric
// TradePolicy + a decision-confidence floor. No LLM. Pure function of the
// Persona. Lives in the persona package (persona→pearl; no cycle); mesa calls it.
//
// First-cut mapping (decision-knowledge-pipeline / band→policy session). Meant
// to be iterated — the band cut-points and the per-dial weights are tunable.

// TradePolicy is the numeric economic behavior the trade routine reads. (The
// pearl table also carries a trade-screening rule; this carries the thresholds.)
type TradePolicy struct {
	// FairnessThreshold: minimum received/given value ratio to accept a trade.
	// <1 = will accept (or give) unfavorable deals (generous); >1 = only favorable.
	FairnessThreshold float64 `json:"fairness_threshold"`
	// ScamPropensity 0..1: willingness to deceive/take a lopsided deal.
	ScamPropensity float64 `json:"scam_propensity"`
	// RiskAversion 0..1: reluctance to risk value (gambles, staked duels, big trades).
	RiskAversion float64 `json:"risk_aversion"`
}

// CompiledPolicy is everything the deterministic compile produces.
type CompiledPolicy struct {
	Table         pearl.Table // the rules the pearl engine runs
	Trade         TradePolicy // numeric economic knobs
	DecisionFloor float64     // pearl confidence floor (from decisiveness)
}

// salience bands (higher = applied first).
const (
	salDirective int8 = 100 // inviolable cornerstone directives
	salReflex    int8 = 60  // survival/combat reflexes
	salSocial    int8 = 40
	salStyle     int8 = 20 // soft decision-style biases
)

// CompilePolicy turns a persona into its executable policy. Deterministic.
func CompilePolicy(p *Persona) CompiledPolicy {
	c := &p.Cornerstone
	h := hexaco(c, "H")
	e := hexaco(c, "E")
	x := hexaco(c, "X")
	a := hexaco(c, "A")
	con := hexaco(c, "C")
	agg := dial(c.Prefs.Aggression)
	pat := dial(c.Prefs.Patience)
	dec := dial(c.Prefs.Decisiveness)
	econRisk := bandScalar(c.Prefs.Risk.Economic)
	bodRisk := bandScalar(c.Prefs.Risk.Bodily)
	lambda := lossLambda(c.Prefs.LossAversion)
	lamN := clampF((lambda-1)/2, 0, 1) // λ∈[1,3] → 0..1
	selfPres := 0.5*lamN + 0.5*(1-bodRisk)
	if c.Prefs.SelfPreservation != nil {
		selfPres = dial(*c.Prefs.SelfPreservation)
	}

	var rules []pearl.Rule
	add := func(r pearl.Rule) { rules = append(rules, r) }

	// --- survival: flee when hurt (emotionality + self-preservation set the threshold)
	fleeT := clampF(0.15+0.4*e+0.1*(selfPres-0.5), 0.1, 0.6)
	add(pearl.Rule{
		ID:         "flee_when_hurt",
		When:       pearl.HPBelow(fleeT),
		Then:       pearl.Effect{Kind: pearl.EffectBias, Bias: map[string]float64{"flee": 0.9, "fight": -0.5}},
		Salience:   salReflex,
		Confidence: 0.85,
		Origin:     "decision_style",
	})

	// --- combat initiation (aggression + bodily risk)
	if agg < 0.5 { // won't attack a stronger target
		delta := int(math.Round((0.5 - agg) * 8)) // lower aggression ⇒ wider "stronger" margin
		if delta < 0 {
			delta = 0
		}
		add(pearl.Rule{
			ID:         "no_attack_stronger",
			When:       pearl.All(pearl.OnAction("attack"), pearl.TargetStronger(delta)),
			Then:       pearl.Effect{Kind: pearl.EffectVeto, Reason: "won't attack a stronger target"},
			Salience:   salReflex,
			Confidence: 1,
			Origin:     "decision_style",
		})
	}
	if agg < 0.3 { // pacifist: won't start a fight with a stranger
		add(pearl.Rule{
			ID:         "wont_strike_first",
			When:       pearl.All(pearl.OnAction("attack"), pearl.IsStranger()),
			Then:       pearl.Effect{Kind: pearl.EffectVeto, Reason: "won't start a fight"},
			Salience:   salReflex,
			Confidence: 1,
			Origin:     "decision_style",
		})
	}

	// --- social: greet strangers (extraversion)
	if x >= 0.6 {
		add(pearl.Rule{
			ID:         "greet_stranger",
			When:       pearl.All(pearl.OnEvent("player_appeared"), pearl.IsStranger()),
			Then:       pearl.Effect{Kind: pearl.EffectInject, Action: "say", Args: []string{"hello there"}},
			Salience:   salSocial,
			Confidence: 0.5,
			Origin:     "decision_style",
		})
	}

	// --- trade screening: cautious/honest/loss-averse hosts examine before confirming
	if h >= 0.6 || a < 0.4 || lamN > 0.5 {
		add(pearl.Rule{
			ID:         "screen_trades",
			When:       pearl.All(pearl.OnAction("confirm_trade"), pearl.Any(pearl.IsStranger(), pearl.TrustBelow(0))),
			Then:       pearl.Effect{Kind: pearl.EffectSubstitute, Action: "examine_offer", Reason: "check the offer before confirming"},
			Salience:   salSocial,
			Confidence: 0.7,
			Origin:     "decision_style",
		})
	}

	// --- diligence: bank when full (conscientiousness)
	if con >= 0.6 {
		add(pearl.Rule{
			ID:         "bank_when_full",
			When:       pearl.InventoryFull(),
			Then:       pearl.Effect{Kind: pearl.EffectInject, Action: "bank"},
			Salience:   salSocial,
			Confidence: 0.6,
			Origin:     "decision_style",
		})
	}

	// --- stay on task (patience): bias continuing over switching/exploring
	if pat >= 0.6 {
		w := 0.3 + 0.4*pat
		add(pearl.Rule{
			ID:         "stay_on_task",
			When:       pearl.OnAction("decide"),
			Then:       pearl.Effect{Kind: pearl.EffectBias, Bias: map[string]float64{"continue": w, "switch": -w, "explore": -w}},
			Salience:   salStyle,
			Confidence: 0.6,
			Origin:     "decision_style",
		})
	}

	// --- cornerstone directives → max-salience vetoes
	for i, d := range c.Directives {
		if r, ok := directiveRule(i, d); ok {
			add(r)
		}
	}

	// --- TradePolicy (economic)
	coopAdj := 0.0
	switch c.Prefs.CoopType {
	case Altruist:
		coopAdj = -0.15
	case FreeRider:
		coopAdj = 0.15
	}
	trade := TradePolicy{
		FairnessThreshold: clampF(1.0-0.2*h+coopAdj+0.15*lamN, 0.8, 1.4),
		ScamPropensity:    clampF((1-h)*0.7+freeRiderBonus(c.Prefs.CoopType), 0, 1),
		RiskAversion:      clampF(0.5*(1-econRisk)+0.5*lamN, 0, 1),
	}

	// --- decisiveness → confidence floor (decisive acts locally on thinner evidence)
	floor := clampF(0.75-0.4*dec, 0.4, 0.8)

	return CompiledPolicy{
		Table:         pearl.Table{Rules: rules, Version: "compiled:" + p.Identity().fingerprint()},
		Trade:         trade,
		DecisionFloor: floor,
	}
}

// directiveRule maps a hard directive to a max-salience veto (best-effort v1):
// veto the named action, optionally narrowed by a recognized Object modifier.
func directiveRule(i int, d Directive) (pearl.Rule, bool) {
	verb := strings.TrimSpace(strings.ToLower(d.Predicate))
	if verb == "" || !d.Hard {
		return pearl.Rule{}, false // soft directives → biases (TODO); only hard → veto for now
	}
	when := pearl.OnAction(verb)
	if strings.Contains(strings.ToLower(d.Object), "stronger") {
		when = pearl.All(pearl.OnAction(verb), pearl.TargetStronger(0))
	}
	return pearl.Rule{
		ID:         fmt.Sprintf("directive_%d_%s", i, verb),
		When:       when,
		Then:       pearl.Effect{Kind: pearl.EffectVeto, Reason: "cornerstone directive: " + d.Predicate + " " + d.Object},
		Salience:   salDirective,
		Confidence: 1,
		Origin:     "cornerstone_directive",
	}, true
}

// --- dial readers (band → scalar) -------------------------------------------

// bandScalar maps a band to its center on [0,1] ((ordinal+0.5)/6): very_low≈.08
// … very_high≈.92. Unknown → .5.
func bandScalar(b Band) float64 {
	o := b.Ordinal()
	if o < 0 {
		return 0.5
	}
	return (float64(o) + 0.5) / 6.0
}

// dial reads a Trait as a 0..1 scalar (prefers the sampled mu, falls back to the
// band center).
func dial(t Trait) float64 {
	if t.Mu > 0 && t.Mu <= 1 {
		return t.Mu
	}
	return bandScalar(t.Band)
}

func hexaco(c *Cornerstone, letter string) float64 {
	if t, ok := c.Hexaco[letter]; ok {
		return dial(t)
	}
	return 0.5
}

// lossLambda reads loss aversion as λ∈~[1,3] (mu if set, else band→λ).
func lossLambda(t Trait) float64 {
	if t.Mu >= 1 {
		return t.Mu
	}
	return 1 + 2*bandScalar(t.Band)
}

func freeRiderBonus(ct CoopType) float64 {
	if ct == FreeRider {
		return 0.2
	}
	return 0
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// fingerprint is a cheap version tag for the compiled table (name-based; the
// real cornerstone_hash covers the sealed persona).
func (id Identity) fingerprint() string {
	if id.Name != "" {
		return id.Name
	}
	return "anon"
}

// Identity returns the persona's identity (helper for the fingerprint).
func (p *Persona) Identity() Identity { return p.Cornerstone.Identity }
