package corpus

import (
	"context"
	"fmt"
	"sort"
)

// Namespace classifies the *audience* a body of knowledge is
// appropriate for. A host's worldview is supposed to feel like that
// of a real player — so player-facing knowledge (the wiki, in-game
// guides, forum advice) is fair game, but the inside view of how
// the game is implemented (server source, client decompiles, the
// AutoRune scripting framework) is NOT something a real player would
// have, and giving it to production hosts would let them reason
// about exploit-shaped strategies.
//
// During development we want BOTH namespaces available so we can
// reason about implementation while we build. At launch, production
// hosts must only see Gameplay sources.
//
// This is a load-time gate, not a query-time one: production cradles
// never instantiate dev sources at all. That makes accidental
// leakage impossible — if a dev corpus isn't in memory, no buggy
// query can reach it.
type Namespace string

const (
	// Gameplay knowledge: anything a real RSC player could plausibly
	// have read or learned. rsc.wiki, official manuals, lore pages,
	// quest walkthroughs. Always allowed.
	Gameplay Namespace = "gameplay"

	// Dev knowledge: the inside view of how the game is implemented.
	// OpenRSC server source, RSC+ client decompile, RSCGo, our own
	// protocol notes, AutoRune script archive. Loaded only in dev
	// cradles; refused at load time for production.
	Dev Namespace = "dev"
)

// Source is one named, namespace-tagged chunk of the larger corpus
// federation — typically one ingested dataset (rsc.wiki, autorune,
// openrsc-source). Sources have stable names so cradle config can
// reference them explicitly.
//
// The Namespace is the load-time access gate; see Federation.
type Source struct {
	Name      string    // stable identifier, e.g. "rscwiki", "autorune"
	Namespace Namespace // Gameplay or Dev
	Corpus    Corpus    // the underlying chunk store
}

// Federation combines multiple Sources into one Corpus, optionally
// filtering by allowed namespaces. Recall merges hits across all
// sources and ranks them globally.
//
// Cradle initialization typically builds a Federation like:
//
//	fed := corpus.NewFederation([]corpus.Source{
//	    {Name: "rscwiki", Namespace: corpus.Gameplay, Corpus: wikiCorpus},
//	    {Name: "autorune", Namespace: corpus.Dev, Corpus: autoruneCorpus},
//	}, allowedNamespaces)
//
// In production, allowedNamespaces is [Gameplay] and the autorune
// source is silently dropped at construction. In dev, it's [Gameplay,
// Dev] and both load.
type Federation struct {
	sources []Source // post-filter set actually queryable
	total   int      // sum of Len() across sources (cached for Len)
}

// NewFederation constructs a Federation. allowed enumerates which
// namespaces this cradle is permitted to see; any Source with a
// Namespace not in `allowed` is excluded — it never enters the
// federation, and Recall has no way to surface its contents.
//
// Passing a nil or empty `allowed` defaults to [Gameplay] — the
// safe-by-default behavior. If you actually need an empty federation
// for tests, pass `[]Namespace{}` and inspect Len() == 0.
//
// A Source whose Corpus is nil is treated as if it had been filtered
// out; this keeps wiring code simple ("oh I haven't built the
// autorune ingest yet, that source is just nil").
func NewFederation(sources []Source, allowed []Namespace) *Federation {
	if allowed == nil {
		allowed = []Namespace{Gameplay}
	}
	allowSet := make(map[Namespace]bool, len(allowed))
	for _, ns := range allowed {
		allowSet[ns] = true
	}
	f := &Federation{}
	for _, s := range sources {
		if s.Corpus == nil {
			continue
		}
		if !allowSet[s.Namespace] {
			continue
		}
		f.sources = append(f.sources, s)
		f.total += s.Corpus.Len()
	}
	return f
}

// Sources returns the post-filter source list — names + namespaces +
// per-source chunk counts. Used by the cradle boot log so operators
// can see exactly what knowledge their cradle has access to.
func (f *Federation) Sources() []SourceInfo {
	out := make([]SourceInfo, 0, len(f.sources))
	for _, s := range f.sources {
		out = append(out, SourceInfo{
			Name:      s.Name,
			Namespace: s.Namespace,
			ChunkCount: s.Corpus.Len(),
		})
	}
	return out
}

// SourceInfo describes one source in the federation for diagnostics.
type SourceInfo struct {
	Name       string
	Namespace  Namespace
	ChunkCount int
}

// Len reports the total chunks across all federated sources.
func (f *Federation) Len() int { return f.total }

// Recall fans the query out to every source in parallel, merges the
// hits, and returns the top N across the union.
//
// Each chunk returned has its Source field set by the underlying
// source's chunks, so routines can branch on provenance even when
// the federation merged multiple corpora.
func (f *Federation) Recall(ctx context.Context, query string, topN int) ([]Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if topN <= 0 {
		topN = 5
	}
	if len(f.sources) == 0 {
		return nil, nil
	}
	// Sequential is fine — corpora are small enough that the
	// goroutine overhead would dominate. If this changes when the
	// dev corpus gets big, switch to errgroup.
	var merged []Chunk
	for _, s := range f.sources {
		hits, err := s.Corpus.Recall(ctx, query, topN)
		if err != nil {
			return nil, fmt.Errorf("source %q: %w", s.Name, err)
		}
		merged = append(merged, hits...)
	}
	// Global top-N by score, ties broken by source order then by
	// the order returned within a source.
	sort.SliceStable(merged, func(i, j int) bool {
		return merged[i].Score > merged[j].Score
	})
	if len(merged) > topN {
		merged = merged[:topN]
	}
	return merged, nil
}

// Verify Federation satisfies Corpus at compile time.
var _ Corpus = (*Federation)(nil)
