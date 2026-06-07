package main

import (
	"context"
	"fmt"
	"os"

	"github.com/gen0cide/westworld/cognition/corpus"
)

func main() {
	mc, stats, err := corpus.LoadWikiDump(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("=== loaded %d chunks from %d pages ===\n", mc.Len(), stats.LoadedFiles)
	hits, _ := mc.Recall(context.Background(), os.Args[2], 50)
	for i, c := range hits {
		fmt.Printf("\n──── chunk %d: [%s § %s] (score %.1f) ────\n%s\n", i, c.PageTitle, c.SectionTitle, c.Score, c.Text)
	}
}
