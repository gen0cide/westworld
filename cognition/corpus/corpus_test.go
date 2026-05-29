package corpus

import (
	"context"
	"strings"
	"testing"
)

func TestMemoryCorpusRecallScoresKeywordHits(t *testing.T) {
	c := LoadFromChunks([]Chunk{
		{Source: "rscwiki", PageTitle: "Cooking", SectionTitle: "Lobsters",
			Text: "To cook a lobster, use a raw lobster on a fire."},
		{Source: "rscwiki", PageTitle: "Cooking", SectionTitle: "Bread",
			Text: "Bread requires flour and water. Use the dough on a range."},
		{Source: "rscwiki", PageTitle: "Mining", SectionTitle: "Iron",
			Text: "Mine iron ore with a pickaxe at iron rocks."},
	})
	out, err := c.Recall(context.Background(), "how do I cook a lobster", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(out) == 0 {
		t.Fatal("expected at least one hit")
	}
	if out[0].SectionTitle != "Lobsters" {
		t.Errorf("top hit: got %q, want \"Lobsters\"", out[0].SectionTitle)
	}
	if out[0].Score <= 0 {
		t.Errorf("score: got %v, want > 0", out[0].Score)
	}
}

func TestMemoryCorpusRecallTopNRespected(t *testing.T) {
	chunks := []Chunk{
		{Source: "rscwiki", PageTitle: "A", Text: "foo foo foo"},
		{Source: "rscwiki", PageTitle: "B", Text: "foo foo"},
		{Source: "rscwiki", PageTitle: "C", Text: "foo"},
	}
	c := LoadFromChunks(chunks)
	out, err := c.Recall(context.Background(), "foo", 2)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(out) != 2 {
		t.Errorf("len: got %d, want 2", len(out))
	}
	// Highest term-frequency wins.
	if out[0].PageTitle != "A" {
		t.Errorf("top: got %q, want A", out[0].PageTitle)
	}
}

func TestMemoryCorpusRecallNoHitsEmpty(t *testing.T) {
	c := LoadFromChunks([]Chunk{
		{Source: "rscwiki", PageTitle: "A", Text: "nothing relevant"},
	})
	out, err := c.Recall(context.Background(), "lobster", 5)
	if err != nil {
		t.Fatalf("Recall: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("expected 0 hits, got %d", len(out))
	}
}

func TestTokenizeQueryDropsStopWords(t *testing.T) {
	got := tokenizeQuery("How do I cook a lobster?")
	want := []string{"cook", "lobster"}
	if !equalStrSlices(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestChunkWikiHTMLExtractsSections(t *testing.T) {
	html := `<html><body>
<h1 id="firstHeading"><span class="mw-page-title-main">Cooking</span></h1>
<div class="mw-parser-output">
<p>Cooking is one of the skills in RuneScape Classic.</p>
<h2><span class="mw-headline">Lobsters</span></h2>
<p>To cook a lobster, use a raw lobster on a fire. You need cooking level 40.</p>
<h2><span class="mw-headline">Bread</span></h2>
<p>Bread is made by using flour on a jug of water, then cooking the dough.</p>
</div>
<div id="catlinks">categories...</div>
</body></html>`
	title, sections := chunkWikiHTML(html)
	if title != "Cooking" {
		t.Errorf("title: got %q, want Cooking", title)
	}
	if len(sections) != 3 {
		t.Fatalf("sections: got %d, want 3 (intro + 2 headings)", len(sections))
	}
	if sections[0].heading != "" {
		t.Errorf("intro heading: got %q, want \"\"", sections[0].heading)
	}
	if !strings.Contains(sections[0].body, "one of the skills") {
		t.Errorf("intro body: got %q", sections[0].body)
	}
	if sections[1].heading != "Lobsters" {
		t.Errorf("section 1 heading: got %q, want Lobsters", sections[1].heading)
	}
	if !strings.Contains(sections[1].body, "raw lobster on a fire") {
		t.Errorf("section 1 body: got %q", sections[1].body)
	}
	if sections[2].heading != "Bread" {
		t.Errorf("section 2 heading: got %q, want Bread", sections[2].heading)
	}
}

func TestChunkWikiHTMLStripsInfobox(t *testing.T) {
	html := `<html><body>
<h1 id="firstHeading"><span>Lobster</span></h1>
<div class="mw-parser-output">
<table class="infobox">junk that should not appear</table>
<p>A lobster is a food item.</p>
</div>
<div id="catlinks"></div>
</body></html>`
	_, sections := chunkWikiHTML(html)
	if len(sections) == 0 {
		t.Fatal("expected at least one section")
	}
	for _, s := range sections {
		if strings.Contains(s.body, "junk that should not appear") {
			t.Errorf("infobox not stripped: %q", s.body)
		}
	}
}

func TestChunkWikiHTMLEmptyPageReturnsNothing(t *testing.T) {
	html := `<html><body><p>just some random html</p></body></html>`
	title, sections := chunkWikiHTML(html)
	if title != "" || sections != nil {
		t.Errorf("expected empty title + nil sections; got %q, %v", title, sections)
	}
}

func equalStrSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
