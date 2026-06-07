package pearl

// EffectKind tags what a rule does when its predicate holds.
type EffectKind uint8

const (
	// EffectBias reweights a decide() option set (additive per-option weight).
	EffectBias EffectKind = iota
	// EffectInject emits a micro-action intent (a reverie/quirk tic).
	EffectInject
	// EffectVeto denies a proposed action (action seam) or forbids options
	// (decision seam, via Effect.Options).
	EffectVeto
	// EffectSubstitute replaces a proposed action / forces a decide() choice.
	EffectSubstitute
)

func (k EffectKind) String() string {
	switch k {
	case EffectBias:
		return "bias"
	case EffectInject:
		return "inject"
	case EffectVeto:
		return "veto"
	case EffectSubstitute:
		return "substitute"
	default:
		return "unknown"
	}
}

// Predicate is a compiled, side-effect-free test over Facts. Predicates are
// selected from the fixed library in predicates.go (composed with all/any/not),
// never authored as free code.
type Predicate func(f *Facts) bool

// Effect describes what fires when a rule's predicate holds.
type Effect struct {
	Kind EffectKind

	// Bias: option label → additive weight (EffectBias, decision seam).
	Bias map[string]float64

	// Options: decide() option labels this effect names — the forbidden
	// options for a veto, unused otherwise.
	Options []string

	// Action: the action to inject (EffectInject) or the option/action to
	// force (EffectSubstitute). Args are positional argument literals.
	Action string
	Args   []string

	// Reason: human/log/audit explanation, surfaced on veto.
	Reason string
}

// Rule is one compiled policy entry: a predicate bolted to an effect, with the
// metadata the engine uses to order and trust it.
type Rule struct {
	ID         string
	When       Predicate
	Then       Effect
	Salience   int8    // higher = evaluated/applied first
	Confidence float64 // [0,1]; TryDecide only answers from rules >= the engine floor
	Origin     string  // "cornerstone_directive" | "quirk:<id>" | "operator_policy" | ...
}

// Table is the compiled artifact a PersonaCompiler emits: the rule set plus a
// version stamp for cache invalidation and audit.
type Table struct {
	Rules   []Rule
	Version string
}

// Verdict is the result of an action-seam Gate evaluation.
type Verdict struct {
	Allow      bool
	Substitute *Effect // non-nil ⇒ run this instead of the gated action
	Reason     string
	RuleID     string
}
