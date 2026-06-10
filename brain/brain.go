package brain

import (
	"context"

	"github.com/gen0cide/westworld/cognition"
)

// Situation is what the brain is asked to reason about. It bundles
// the question, the cognition retrieval that supports the answer,
// and optionally the enumerated choices the brain must pick from.
//
// Field semantics:
//
//   - Question: the prompt the strategist must answer. Free-form
//     text. Stable phrasing is encouraged so log-grepping and
//     stub heuristics stay tractable.
//   - Bundle: the cognition.Bundle assembled for this question.
//     May be nil for trivial decisions that don't need memory
//     context (the stub tolerates nil; the real strategist
//     should treat nil as "you have no memory context, decide
//     anyway"). Caller owns the Bundle's lifetime.
//   - Options: optional enumerated choices. When non-empty, the
//     Decision.Choice must be one of these strings; when empty,
//     the Decision.Choice is a free-form answer.
type Situation struct {
	Question string
	Bundle   *cognition.Bundle
	Options  []string
}

// Decision is the strategist's output.
//
// Field semantics:
//
//   - Choice: the selected option (verbatim one of Situation.Options
//     when Options is non-empty) or a free-form answer otherwise.
//   - Reasoning: brief rationale for delos / chain-of-thought logs.
//     The real strategist captures the LLM's terse explanation; the
//     stub returns a canned one-liner.
//   - Confidence: a scalar in [0, 1]. The real strategist derives
//     this from logprobs or a self-rating prompt; the stub returns
//     a fixed value per branch.
type Decision struct {
	Choice     string
	Reasoning  string
	Confidence float64
}

// Strategist makes deliberative decisions. The real implementation
// implementation lives mesa-side (mesad Act/Decide RPCs); the stub returns canned decisions.
//
// Implementations are expected to be safe for concurrent use across
// hosts in one process.
type Strategist interface {
	// Decide returns a Decision for the given Situation. ctx is
	// honored for cancellation; the real strategist also uses it
	// for per-class rate-limit waits.
	//
	// A non-nil error is returned only for unrecoverable problems
	// (LLM unavailable, rate-limit timeout, etc.). Typed DSL
	// errors are produced by the runtime bridge, not here.
	Decide(ctx context.Context, s Situation) (*Decision, error)
}
