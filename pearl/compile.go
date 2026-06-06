package pearl

// This file is a placeholder for the real PersonaCompiler. The compiler's job
// is to turn a persona document (HEXACO disposition bands + directives +
// quirks) into a Table by SELECTING parameterized predicates from the library
// and attaching effects/weights — a deterministic switch, no LLM. Until the
// persona schema lands (see docs persona-* + decision-policy-quirk-engine),
// DemoTable hand-authors a representative table so the engine and the runtime
// seams can be wired and tested end to end.

// Salience bands. Operator policy is inviolable and evaluated first; cornerstone
// directives are the persona's hard rules; quirks and decision-style biases are
// softer.
const (
	salienceOperator    int8 = 100
	salienceCornerstone int8 = 80
	salienceQuirk       int8 = 40
	salienceStyle       int8 = 20
)

// DemoTable returns a small, illustrative policy table exercising all four
// effect kinds. It approximates a timid, cautious persona:
//
//   - never attack a target meaningfully stronger than you (cornerstone veto);
//   - never confirm a trade with an untrusted stranger (cornerstone veto);
//   - double-check a trade once before confirming (quirk substitute);
//   - flee rather than fight when badly hurt (decision-style bias);
//   - greet a newly-appeared stranger (quirk inject).
//
// Replace with PersonaCompiler.Compile(persona) once the schema exists.
func DemoTable() Table {
	return Table{
		Version: "demo-1",
		Rules: []Rule{
			{
				ID:         "no_attack_stronger",
				When:       All(OnAction("attack"), TargetStronger(0)),
				Then:       Effect{Kind: EffectVeto, Reason: "won't attack a stronger target"},
				Salience:   salienceCornerstone,
				Confidence: 1,
				Origin:     "cornerstone_directive",
			},
			{
				ID:         "no_trade_untrusted_stranger",
				When:       All(OnAction("confirm_trade"), Any(TrustBelow(0), IsStranger())),
				Then:       Effect{Kind: EffectVeto, Reason: "won't trade with an untrusted stranger"},
				Salience:   salienceCornerstone,
				Confidence: 1,
				Origin:     "cornerstone_directive",
			},
			{
				ID:         "double_check_trade",
				When:       OnAction("confirm_trade"),
				Then:       Effect{Kind: EffectSubstitute, Action: "examine_offer", Reason: "double-check the offer first"},
				Salience:   salienceQuirk,
				Confidence: 0.7,
				Origin:     "quirk:trade_double_check",
			},
			{
				ID:         "flee_when_hurt",
				When:       HPBelow(0.3),
				Then:       Effect{Kind: EffectBias, Bias: map[string]float64{"flee": 0.9, "fight": -0.5}},
				Salience:   salienceStyle,
				Confidence: 0.85,
				Origin:     "decision_style",
			},
			{
				ID:         "greet_stranger",
				When:       All(OnEvent("player_appeared"), IsStranger()),
				Then:       Effect{Kind: EffectInject, Action: "say", Args: []string{"hello there"}},
				Salience:   salienceQuirk,
				Confidence: 0.5,
				Origin:     "quirk:friendly",
			},
		},
	}
}
