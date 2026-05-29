// Package corpus is the knowledge-retrieval substrate for the
// cognition layer. A Corpus is a body of indexed text chunks
// (rsc.wiki sections, AutoRune script snippets, etc.) that hosts
// can query via Recall.
//
// Phase 2.6 Slice 1 (this file) is the in-memory keyword-only
// implementation. It loads the rsc.wiki HTML dump from disk, chunks
// every page by h2/h3 section heading, and scores Recall queries by
// simple term-frequency matching. No external services, no
// embeddings, no Postgres — but enough that routines can already
// branch on real wiki content via the `recall()` DSL builtin.
//
// Slice 2 will swap the MemoryCorpus implementation for one backed
// by mesa (Postgres + pgvector + Voyage 3 embeddings) without
// changing the Corpus interface or the DSL surface.
package corpus

import (
	"context"
	"fmt"
	"strings"
)

// Chunk is one retrievable unit of knowledge — typically a wiki
// section, an AutoRune script fragment, or any other coherent
// slice of text the host can be handed during reasoning.
//
// Score is populated by Recall; it has no meaning outside a single
// retrieval and is not stored.
type Chunk struct {
	Source       string  // "rscwiki", "autorune", etc.
	PageTitle    string  // e.g., "Cooking"
	SectionTitle string  // e.g., "Cooking levels", or empty for page intro
	URL          string  // canonical source URL when known
	Text         string  // plain-text body of the chunk
	Score        float64 // search relevance, 0..N, higher is better
}

// Corpus is the retrieval surface used by both the host runtime
// (via `recall()` in the DSL) and the cradle admin CLI.
// Implementations are expected to be safe for concurrent use.
type Corpus interface {
	// Recall returns up to topN chunks ranked by relevance to the
	// query. Implementations should honor ctx cancellation. An
	// empty corpus returns an empty slice with no error.
	Recall(ctx context.Context, query string, topN int) ([]Chunk, error)

	// Len reports how many chunks live in the corpus. Useful for
	// boot-time sanity logs and the `-knowledge-query` CLI.
	Len() int
}

// MemoryCorpus is the in-memory keyword-only implementation used in
// Slice 1. It holds the full chunk set in a slice and scores every
// Recall by counting query-term hits in chunk text.
//
// Construction: use LoadWikiDump (or LoadFromChunks for tests).
//
// Memory cost is rough: ~6k wiki pages × ~3 sections/page × ~600
// bytes/chunk ≈ 11 MiB. Fine to hold in every cradle process; not
// fine at 500-host scale, but slice 2 moves to a shared service.
type MemoryCorpus struct {
	chunks []Chunk
}

// LoadFromChunks constructs a MemoryCorpus from a pre-built slice.
// Primarily for tests; production loaders build via LoadWikiDump.
func LoadFromChunks(chunks []Chunk) *MemoryCorpus {
	out := make([]Chunk, len(chunks))
	copy(out, chunks)
	return &MemoryCorpus{chunks: out}
}

// Len returns the number of chunks in the corpus.
func (c *MemoryCorpus) Len() int { return len(c.chunks) }

// Recall scores every chunk against the query and returns the top N
// by score. Scoring is intentionally simple — it's a substring-hit
// count weighted slightly by title matches:
//
//   - +1 per occurrence of a query term in the body
//   - +5 if a query term appears in the section title
//   - +10 if a query term appears in the page title
//
// Stop words ("the", "a", "is", etc.) are dropped from the query
// before scoring so common phrases don't drown out the useful terms.
//
// Ties are broken by the chunk's natural order in the corpus (load
// order, which is filesystem order — stable and deterministic).
func (c *MemoryCorpus) Recall(ctx context.Context, query string, topN int) ([]Chunk, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if topN <= 0 {
		topN = 5
	}
	terms := tokenizeQuery(query)
	if len(terms) == 0 {
		return nil, nil
	}
	var hits []scoredChunk
	for i := range c.chunks {
		s := scoreChunk(&c.chunks[i], terms)
		if s > 0 {
			hits = append(hits, scoredChunk{idx: i, score: s})
		}
	}
	// Top-N via partial-sort. corpus.Len is small enough that a
	// full sort would also be fine; partial keeps it predictable.
	sortScoredDesc(hits)
	if len(hits) > topN {
		hits = hits[:topN]
	}
	out := make([]Chunk, len(hits))
	for i, h := range hits {
		out[i] = c.chunks[h.idx]
		out[i].Score = h.score
	}
	return out, nil
}

func scoreChunk(ch *Chunk, terms []string) float64 {
	body := strings.ToLower(ch.Text)
	section := strings.ToLower(ch.SectionTitle)
	page := strings.ToLower(ch.PageTitle)
	var s float64
	for _, t := range terms {
		s += float64(strings.Count(body, t))
		if strings.Contains(section, t) {
			s += 5
		}
		if strings.Contains(page, t) {
			s += 10
		}
	}
	return s
}

// tokenizeQuery lowercases, strips punctuation, and removes a small
// English stop-word list. Keep this dumb on purpose — slice 2's
// embedding model handles semantics; this is just plumbing.
func tokenizeQuery(query string) []string {
	q := strings.ToLower(query)
	var b strings.Builder
	b.Grow(len(q))
	for _, r := range q {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte(' ')
		}
	}
	raw := strings.Fields(b.String())
	out := raw[:0]
	for _, w := range raw {
		if len(w) < 2 || stopWords[w] {
			continue
		}
		out = append(out, w)
	}
	return out
}

var stopWords = map[string]bool{
	"a": true, "an": true, "and": true, "are": true, "as": true,
	"at": true, "be": true, "by": true, "do": true, "for": true,
	"from": true, "how": true, "i": true, "in": true, "is": true,
	"it": true, "of": true, "on": true, "or": true, "that": true,
	"the": true, "this": true, "to": true, "was": true, "what": true,
	"when": true, "where": true, "which": true, "who": true,
	"will": true, "with": true, "you": true,
}

// scoredChunk is the internal (chunk-index, score) tuple Recall
// builds while ranking. Top-level type so the sort helper can name
// its argument cleanly.
type scoredChunk struct {
	idx   int
	score float64
}

// sortScoredDesc is a tiny inline sort. Avoiding a sort.Slice
// closure call here is mostly aesthetic — the corpus is small —
// but it keeps the package free of stdlib sort overhead too.
func sortScoredDesc(xs []scoredChunk) {
	// Insertion sort: the inputs are typically small (hits<<corpus).
	for i := 1; i < len(xs); i++ {
		for j := i; j > 0 && xs[j-1].score < xs[j].score; j-- {
			xs[j-1], xs[j] = xs[j], xs[j-1]
		}
	}
}

// Verify MemoryCorpus satisfies Corpus at compile time.
var _ Corpus = (*MemoryCorpus)(nil)

// String renders a chunk for logs / debugging — short header line
// followed by a truncated body. Useful for the cradle CLI.
func (c Chunk) String() string {
	body := c.Text
	if len(body) > 200 {
		body = body[:200] + "..."
	}
	header := c.PageTitle
	if c.SectionTitle != "" {
		header = fmt.Sprintf("%s § %s", c.PageTitle, c.SectionTitle)
	}
	return fmt.Sprintf("[%s] %s\n%s", c.Source, header, body)
}
