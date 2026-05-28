package brain

import (
	"context"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/cognition"
)

// StubStrategist is a deterministic, zero-I/O implementation of
// Strategist used for tests and early integration before the real
// LLM-backed strategist lands in Phase 4.
//
// Heuristics (in priority order, evaluated against a case-folded
// trimmed Question):
//
//  1. If Situation.Options is non-empty → Choice is the first
//     option, Confidence 0.75, Reasoning names the option.
//  2. If Question begins with "should i" (or "should we") → Choice
//     is "yes", Confidence 0.6, Reasoning is canned.
//  3. If Question begins with "is " or "are " or "can " or "do " or
//     "does " → Choice is "yes", Confidence 0.55, Reasoning is
//     canned.
//  4. If Question begins with "where" → Choice is "Lumbridge",
//     Confidence 0.5.
//  5. If Question begins with "what" / "how" / "why" → Choice is a
//     canned descriptive answer, Confidence 0.5.
//  6. Anything else → Choice is "ok", Confidence 0.4.
//
// The Reasoning field always mentions whether a Bundle was supplied
// and how many reflections + episodes it contained — that way the
// strategist's chain-of-thought log will visibly reflect cognition
// integration even with the stub in place.
//
// StubStrategist holds no state and is safe for concurrent use.
type StubStrategist struct{}

// NewStubStrategist returns a fresh StubStrategist. Provided for
// symmetry with cognition.NewStubClient; the zero value also works.
func NewStubStrategist() *StubStrategist {
	return &StubStrategist{}
}

// Decide returns a canned Decision per the heuristics documented on
// StubStrategist. It never returns an error in the stub; the
// signature is preserved for parity with the real strategist.
func (s *StubStrategist) Decide(ctx context.Context, sit Situation) (*Decision, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	bundleNote := summarizeBundle(sit.Bundle)
	q := strings.ToLower(strings.TrimSpace(sit.Question))

	// 1. Enumerated options take precedence.
	if len(sit.Options) > 0 {
		return &Decision{
			Choice:     sit.Options[0],
			Reasoning:  "stub strategist picked the first option (" + sit.Options[0] + "); " + bundleNote,
			Confidence: 0.75,
		}, nil
	}

	// 2-5. Question-prefix heuristics.
	switch {
	case strings.HasPrefix(q, "should i") || strings.HasPrefix(q, "should we"):
		return &Decision{
			Choice:     "yes",
			Reasoning:  "stub strategist defaults to yes for should-questions; " + bundleNote,
			Confidence: 0.6,
		}, nil
	case hasAnyPrefix(q, "is ", "are ", "can ", "do ", "does ", "will ", "would "):
		return &Decision{
			Choice:     "yes",
			Reasoning:  "stub strategist defaults to yes for closed yes/no questions; " + bundleNote,
			Confidence: 0.55,
		}, nil
	case strings.HasPrefix(q, "where"):
		return &Decision{
			Choice:     "Lumbridge",
			Reasoning:  "stub strategist defaults to Lumbridge for where-questions; " + bundleNote,
			Confidence: 0.5,
		}, nil
	case hasAnyPrefix(q, "what", "how", "why"):
		return &Decision{
			Choice:     "I would explore the world and stay alive.",
			Reasoning:  "stub strategist gave a canned open-question answer; " + bundleNote,
			Confidence: 0.5,
		}, nil
	}

	// 6. Fallback.
	return &Decision{
		Choice:     "ok",
		Reasoning:  "stub strategist had no heuristic match; " + bundleNote,
		Confidence: 0.4,
	}, nil
}

func hasAnyPrefix(s string, prefixes ...string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// summarizeBundle returns a one-line description of the supplied
// Bundle (or "no bundle" if nil). Stitched into Reasoning so
// chain-of-thought logs visibly reflect what was retrieved.
func summarizeBundle(b *cognition.Bundle) string {
	if b == nil {
		return "no bundle supplied"
	}
	return fmt.Sprintf("bundle had %d reflection(s), %d episode(s)",
		len(b.Reflections), len(b.Episodic))
}
