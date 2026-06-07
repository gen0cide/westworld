// Command wikirag is an OUT-OF-BAND tool: it embeds the clean rsc.wiki chunks
// with Voyage and runs semantic search over them. For OUR purposes (offline
// distillation / authoring — e.g. building persona-cook WorldBriefs), NOT a
// host runtime resource. Hosts do not get live wiki query access.
//
//	source .local.env && go run ./mesa/wikirag -index -dir local/rscwiki/pages -cache local/wiki.gob
//	source .local.env && go run ./mesa/wikirag -cache local/wiki.gob -q "how do I stop burning lobsters"
package main

import (
	"context"
	"encoding/gob"
	"flag"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/cognition/corpus"
)

// indexedChunk is a chunk plus its embedding (gob-persisted to a cache file so
// we embed once).
type indexedChunk struct {
	Chunk corpus.Chunk
	Vec   []float32
}

func main() {
	dir := flag.String("dir", "local/rscwiki/pages", "directory of rsc.wiki HTML pages")
	cache := flag.String("cache", "local/wiki.gob", "embedding cache file (gob)")
	doIndex := flag.Bool("index", false, "build the embedding cache from -dir")
	q := flag.String("q", "", "semantic query against the cache")
	top := flag.Int("top", 5, "top-N results")
	model := flag.String("model", "voyage-3", "Voyage model")
	batch := flag.Int("batch", 96, "chunks per embedding request")
	limit := flag.Int("limit", 0, "cap chunks embedded (0 = all; for cheap validation)")
	flag.Parse()

	key := os.Getenv("VOYAGE_AI_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "wikirag: set VOYAGE_AI_KEY (e.g. source .local.env)")
		os.Exit(2)
	}
	vc := newVoyage(key, *model)
	ctx := context.Background()

	switch {
	case *doIndex:
		if err := index(ctx, vc, *dir, *cache, *batch, *limit); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case *q != "":
		if err := query(ctx, vc, *cache, *q, *top); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "wikirag: pass -index to build, or -q <query> to search")
		os.Exit(2)
	}
}

// embedText is what we actually embed: the chunk text, prefixed with its page +
// section so the vector captures context ("Cooking § Level to Stop Burning…").
func embedText(c corpus.Chunk) string {
	ctx := c.PageTitle
	if c.SectionTitle != "" {
		ctx += " § " + c.SectionTitle
	}
	return ctx + ": " + c.Text
}

func index(ctx context.Context, vc *voyageClient, dir, cache string, batch, limit int) error {
	mc, stats, err := corpus.LoadWikiDump(dir)
	if err != nil {
		return err
	}
	chunks := mc.Chunks()
	if limit > 0 && limit < len(chunks) {
		chunks = chunks[:limit]
	}
	fmt.Printf("loaded %d chunks from %d pages; embedding %d with %s…\n", mc.Len(), stats.LoadedFiles, len(chunks), vc.model)

	idx := make([]indexedChunk, 0, len(chunks))
	tokens := 0
	for i := 0; i < len(chunks); i += batch {
		end := min(i+batch, len(chunks))
		texts := make([]string, end-i)
		for j := i; j < end; j++ {
			texts[j-i] = embedText(chunks[j])
		}
		vecs, toks, err := vc.embed(ctx, texts, "document")
		if err != nil {
			return fmt.Errorf("embed batch %d: %w", i, err)
		}
		tokens += toks
		for j := i; j < end; j++ {
			idx = append(idx, indexedChunk{Chunk: chunks[j], Vec: vecs[j-i]})
		}
		fmt.Printf("\rembedded %d/%d (%d tokens)", end, len(chunks), tokens)
	}
	fmt.Println()

	f, err := os.Create(cache)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := gob.NewEncoder(f).Encode(idx); err != nil {
		return err
	}
	fmt.Printf("wrote %d vectors to %s (%d tokens embedded)\n", len(idx), cache, tokens)
	return nil
}

func query(ctx context.Context, vc *voyageClient, cache, q string, top int) error {
	f, err := os.Open(cache)
	if err != nil {
		return fmt.Errorf("open cache (build it with -index first): %w", err)
	}
	defer f.Close()
	var idx []indexedChunk
	if err := gob.NewDecoder(f).Decode(&idx); err != nil {
		return err
	}
	qv, _, err := vc.embed(ctx, []string{q}, "query")
	if err != nil {
		return err
	}
	type hit struct {
		score float64
		c     corpus.Chunk
	}
	hits := make([]hit, len(idx))
	for i, ic := range idx {
		hits[i] = hit{score: cosine(qv[0], ic.Vec), c: ic.Chunk}
	}
	sort.Slice(hits, func(a, b int) bool { return hits[a].score > hits[b].score })

	fmt.Printf("=== top %d for %q (over %d chunks) ===\n", top, q, len(idx))
	for i := 0; i < top && i < len(hits); i++ {
		h := hits[i]
		fmt.Printf("\n#%d  (%.3f)  [%s § %s]\n%s\n", i+1, h.score, h.c.PageTitle, h.c.SectionTitle, snippet(h.c.Text, 320))
	}
	return nil
}

func cosine(a, b []float32) float64 {
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		na += float64(a[i]) * float64(a[i])
		nb += float64(b[i]) * float64(b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func snippet(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
