// Package pearl is a host's onboard cognitive fast path: a small, deterministic,
// pure-Go policy engine that gates and shapes a host's decisions and actions
// WITHOUT an LLM round-trip.
//
// The metaphor is the Westworld control unit ("pearl") — the host carries its
// compiled disposition inside it. Concretely the engine is one flat rule table
// per host:
//
//	Rule = {When Predicate, Then Effect, Salience, Confidence, Origin}
//
// A Predicate is a side-effect-free Go test over a read-only Facts snapshot,
// selected from a small library of parameterized primitives (hpBelow, trustBelow,
// onAction, ...) composed with all/any/not. An Effect biases a decide() option
// set, vetoes/substitutes an action, or injects a micro-action. The table is
// COMPILED ONCE from the persona (off the hot path) and then evaluated by two
// tiny scanners at two seams:
//
//   - TryDecide: the decision fast path, called before the LLM Strategist. A
//     confident local hit returns a brain.Decision and no LLM call happens.
//   - Gate: the action veto, called synchronously before an action executes;
//     it can deny or substitute author-written or brain-authored actions.
//
// The engine holds no model weights and does no I/O. See the design record at
// docs/_research/reference/decision-policy-quirk-engine.md.
package pearl
