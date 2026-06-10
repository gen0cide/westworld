// Package brain is the host's deliberative strategist tier: given a
// Situation (a question + a context Bundle from cognition), it
// produces a Decision (a choice + reasoning + confidence).
//
// The LLM strategist shipped as mesa RPCs (mesad's actLLM/decideLLM/
// genesisLLM seams), NOT as an in-process implementation of this package —
// brain remains the local stub/interface layer. An earlier plan had it call an LLM
// — Sonnet for strategic/script-gen decisions, Haiku for tactical
// /chat/reactive decisions. Models are selected by decision class,
// per-class rate limiters keep cost bounded, and every call writes
// a row to a mesa brain_calls table (never built; the decision stream in
// runtime/decisions.go is what shipped) for chain-of-thought
// observability.
//
// This package currently ships a deterministic StubStrategist that
// returns canned decisions based on the Situation.Question prefix
// and the supplied Options. The stub lets routine tests + early
// integration wire to a stable brain.Strategist interface today,
// before LLM access lands.
//
// # Dependency direction
//
//   - brain imports cognition for the Bundle type (a Situation
//     embeds the *cognition.Bundle the strategist reasons over).
//   - brain does NOT import runtime, dsl, or mesa. The cradle
//     constructs a Strategist + wires it to runtime; brain itself
//     stays pure for testability.
//
// # Error model
//
// Strategist.Decide returns a Decision + error. The error is for
// unrecoverable problems (LLM down, rate limit exceeded after
// retries, malformed response after retry). For typed DSL-visible
// failures the bridge layer in runtime/ converts the error into a
// dsl/interp.CallResult — that conversion lives at the call site,
// not here, so brain stays free of DSL concerns.
//
// # Conventions
//
//   - Decide takes a context.Context as its first argument. The
//     real strategist will issue network calls; the stub honors
//     cancellation as a courtesy.
//   - Implementations are expected to be safe for concurrent use
//     across hosts in one process.
//   - Confidence is a float64 in [0, 1] — callers may treat it as
//     a soft signal (e.g. low-confidence answers might trigger a
//     re-query or a defer-to-routine fallback).
package brain
