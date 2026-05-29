package resolve

import (
	"context"

	"github.com/gen0cide/westworld/facts"
)

// maxResults caps the ranked list resolve() returns. Recognition wants a short
// shortlist, not a search dump; callers that need everything can rank the
// catalog themselves.
const maxResults = 8

// Resolver is the host's recognition faculty. It wires the three-stage
// resolve() pipeline from api.md §5:
//
//	(1) learned-alias store  →  (2) conservative fuzzy/token match  →  (3) brain
//
// Construct one with New. A Resolver is safe for concurrent use: the alias
// store guards its own state, the catalog is read-only after construction, and
// the brain hook is expected to be concurrency-safe.
type Resolver struct {
	catalog *Catalog
	aliases *AliasStore
	brain   Brain
}

// New builds a Resolver over the given facts registry, alias store, and brain
// hook. f may be nil (spells/prayers still resolve; world catalogs are empty).
// store may be nil (an in-memory store is created — learning works for the
// session but is not persisted). brain may be nil (the brain stage is skipped,
// i.e. pipeline collapses to alias + fuzzy).
func New(f *facts.Facts, store *AliasStore, brain Brain) *Resolver {
	if store == nil {
		store = NewAliasStore()
	}
	if brain == nil {
		brain = noBrain{}
	}
	return &Resolver{
		catalog: NewCatalog(f),
		aliases: store,
		brain:   brain,
	}
}

// NewWithCatalog is New for callers that have already built (or want to share)
// a Catalog — e.g. one catalog shared across many hosts, each with its own
// alias store. The same nil semantics for store/brain apply.
func NewWithCatalog(catalog *Catalog, store *AliasStore, brain Brain) *Resolver {
	if catalog == nil {
		catalog = NewCatalog(nil)
	}
	if store == nil {
		store = NewAliasStore()
	}
	if brain == nil {
		brain = noBrain{}
	}
	return &Resolver{catalog: catalog, aliases: store, brain: brain}
}

// Aliases exposes the underlying learned-alias store so a host can seed it,
// inspect it, or persist it directly.
func (r *Resolver) Aliases() *AliasStore { return r.aliases }

// Resolve maps loose text to a ranked list of canonical facts definitions,
// best-first. kind ("", or one of the Kind* constants) filters which catalogs
// are searched. It uses context.Background for the brain stage; callers that
// need cancellation should use ResolveCtx.
//
// This is the Go entry point behind the DSL primitive
// `resolve(text, kind?) -> List<Match{def, kind, score}>` (api.md §5).
func (r *Resolver) Resolve(text, kind string) []Match {
	return r.ResolveCtx(context.Background(), text, kind)
}

// ResolveCtx is Resolve with an explicit context threaded to the brain stage.
func (r *Resolver) ResolveCtx(ctx context.Context, text, kind string) []Match {
	if normalize(text) == "" {
		return nil
	}

	// Stage 1: learned-alias store. A hit is an exact, deterministic table
	// lookup — the cheap fast-path. We still resolve it through the catalog so
	// the returned Match carries the real def + id.
	if canonical, ok := r.aliases.Lookup(text, kind); ok {
		if e, found := r.catalog.find(canonical, kind); found {
			return []Match{{
				Def:       e.Def,
				Kind:      e.Kind,
				Score:     scoreExact,
				Canonical: e.Canonical,
				ID:        e.ID,
				Source:    SourceAlias,
			}}
		}
		// The alias points at a name no longer in the catalog (defs changed):
		// fall through to fuzzy rather than returning a dangling match.
	}

	// Stage 2: conservative fuzzy / token match against canonical names.
	matches := rankFuzzy(text, r.catalog.entries(kind))
	if len(matches) > maxResults {
		matches = matches[:maxResults]
	}
	if len(matches) > 0 {
		return matches
	}

	// Stage 3: brain fallback. Consulted only when stages 1+2 found nothing.
	// On a confirmed, catalog-verifiable answer we WRITE THE ALIAS BACK so the
	// next lookup is a stage-1 hit, then return the resolved def.
	if m, ok := r.askBrain(ctx, text, kind); ok {
		return []Match{m}
	}

	return nil
}

// askBrain runs the LLM fallback hook and validates its answer against the
// catalog. It never trusts an id from the brain — the id is recovered from the
// facts catalog by the canonical name the brain returns. On success it learns
// the mapping (write-back) and returns the resolved Match.
func (r *Resolver) askBrain(ctx context.Context, text, kind string) (Match, bool) {
	candidates := r.candidates(kind)
	canonical, resolvedKind, ok := r.brain.ResolveAlias(ctx, text, kind, candidates)
	if !ok || normalize(canonical) == "" {
		return Match{}, false
	}
	// Validate: the brain's answer must be a real canonical name in the
	// catalog. Prefer the kind it reported; fall back to the requested filter,
	// then to any kind. This is where a hallucinated name is rejected.
	lookKind := resolvedKind
	if lookKind == "" {
		lookKind = kind
	}
	e, found := r.catalog.find(canonical, lookKind)
	if !found && lookKind != kind {
		e, found = r.catalog.find(canonical, kind)
	}
	if !found && kind != "" {
		// Requested kind constrains the result; do not cross it.
		return Match{}, false
	}
	if !found {
		e, found = r.catalog.find(canonical, "")
	}
	if !found {
		return Match{}, false
	}

	// Write-back: learn the loose text → resolved canonical under the resolved
	// kind. A persistence error must not sink the resolution; the in-memory
	// table is updated regardless (Learn's contract), so we ignore the error
	// here — callers wanting to surface flush failures use the store directly.
	_ = r.aliases.Learn(text, e.Canonical, e.Kind)

	return Match{
		Def:       e.Def,
		Kind:      e.Kind,
		Score:     scoreBrain,
		Canonical: e.Canonical,
		ID:        e.ID,
		Source:    SourceBrain,
	}, true
}

// candidates flattens the searched catalog into the (name, kind) options shown
// to the brain. We pass names + kinds only — never ids.
func (r *Resolver) candidates(kind string) []Candidate {
	es := r.catalog.entries(kind)
	out := make([]Candidate, 0, len(es))
	for _, e := range es {
		out = append(out, Candidate{Canonical: e.Canonical, Kind: e.Kind})
	}
	return out
}
