package resolve

import "context"

// Brain is the LLM-fallback hook for the third stage of the resolve()
// pipeline. It is consulted only when the learned-alias store misses and the
// fuzzy matcher produces nothing — i.e. genuinely unfamiliar text.
//
// The contract (api.md §5) is deliberately narrow: the brain is shown the
// loose text and the candidate canonical names available in the relevant
// catalog(s), and asked only to pick the one canonical name (verbatim) that
// the text refers to, or to decline. It NEVER returns or invents an id — the
// id is recovered from the facts catalog by the canonical name it returns. On
// a confirmed mapping the resolver writes the alias back to the store so the
// next lookup is a cheap stage-1 hit.
//
// Implementations must honor ctx cancellation. A real implementation will be
// backed by the brain package; tests use a stub.
type Brain interface {
	// ResolveAlias asks the brain to map text to exactly one of the supplied
	// candidate canonical names, optionally constrained to kind ("" = any).
	//
	// Return value contract:
	//   - canonical: a name that appears verbatim in candidates, or "" to
	//     decline. The resolver validates the return against the catalog and
	//     discards anything not found (so a hallucinated name is harmless).
	//   - kind: the catalog the chosen name belongs to (one of the Kind*
	//     constants). May be "" if the brain doesn't track it; the resolver
	//     will re-derive it from the catalog.
	//   - ok: whether the brain made a confident choice. false means decline.
	ResolveAlias(ctx context.Context, text, kind string, candidates []Candidate) (canonical, resolvedKind string, ok bool)
}

// Candidate is one option presented to the Brain: a canonical name and its
// catalog kind. (We pass names + kinds, never ids — ids are the resolver's to
// recover, never the model's to choose.)
type Candidate struct {
	Canonical string
	Kind      string
}

// noBrain is the default Brain used when none is configured: it always
// declines, which collapses the pipeline to stages 1+2 (alias store + fuzzy).
type noBrain struct{}

func (noBrain) ResolveAlias(context.Context, string, string, []Candidate) (string, string, bool) {
	return "", "", false
}

// StubBrain is a deterministic, zero-I/O Brain for tests and early
// integration. It maps text→canonical using a small, explicit table the test
// supplies; anything not in the table is declined. It records each call so
// tests can assert the brain was (or was not) consulted, and that write-back
// happened exactly once.
//
// The table key is the *normalized* text (see normalize), so the test can key
// it the same loose way a routine would type it.
type StubBrain struct {
	// Table maps normalized text → the canonical answer the stub should give.
	Table map[string]Candidate
	// Calls counts ResolveAlias invocations (for "consulted once" assertions).
	Calls int
}

// NewStubBrain returns a StubBrain seeded with the given table. Keys are
// normalized on insert so callers can pass loose phrasing.
func NewStubBrain(table map[string]Candidate) *StubBrain {
	norm := make(map[string]Candidate, len(table))
	for k, v := range table {
		norm[normalize(k)] = v
	}
	return &StubBrain{Table: norm}
}

// ResolveAlias implements Brain against the stub table.
func (b *StubBrain) ResolveAlias(_ context.Context, text, kind string, _ []Candidate) (string, string, bool) {
	b.Calls++
	c, ok := b.Table[normalize(text)]
	if !ok {
		return "", "", false
	}
	if kind != "" && c.Kind != "" && c.Kind != kind {
		// The caller constrained kind and the stub's answer is a different
		// catalog — decline rather than cross the filter.
		return "", "", false
	}
	return c.Canonical, c.Kind, true
}
