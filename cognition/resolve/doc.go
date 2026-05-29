// Package resolve is the host's recognition faculty: it turns the loose,
// player-spoken text a routine uses ("r2h", "rune 2 handed", "att pot")
// into a ranked list of canonical world definitions (items, npcs, locs,
// spells, prayers) drawn from the facts registry.
//
// This is the implementation behind the cognition-plane primitive
// documented in docs/lang/api.md §5:
//
//	resolve(text: String, kind?: String) -> List<Match{ def, kind, score }>
//
// It is explicitly NOT a GUI-equivalent body verb — it is a learnable,
// mind-access faculty. The pipeline, in priority order, is:
//
//  1. The host's learned-alias store (a persisted, per-host text→canonical
//     table). Each host grows its own lingo as it plays; we ship no curated
//     slang. This is the cheap, deterministic fast-path: once "r2h" has been
//     learned to mean "Rune 2-handed Sword", it is a table hit forever after.
//
//  2. A conservative fuzzy / token match against the canonical names in the
//     facts registry. Exact (case-fold) and exact-after-normalization hits
//     score highest; token-subset and substring matches score lower and are
//     returned ranked best-first so an ambiguous query ("potion") yields a
//     ranked candidate list rather than a guess.
//
//  3. A brain (LLM) fallback HOOK (the Brain interface). The LLM is asked
//     only to map the loose text to a canonical name it is shown — it never
//     invents ids. On a successful, verifiable mapping, the resolved alias is
//     WRITTEN BACK to the learned-alias store so the next lookup is a step-1
//     table hit.
//
// Definitions are always looked up *exactly* in facts; only the
// name→canonical recognition is fuzzy, and ids come from the resolved def,
// never from the model.
//
// The DSL builtin (`resolve(...)`) is registered during merge in
// runtime/dsl_actions.go; this package deliberately does not touch that file.
package resolve
