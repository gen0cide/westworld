package corpus

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// LoadWikiDump scans `dir` for `*.html` files (the output format of
// the rsc.wiki crawler at ~/Code/westworld/local/rscwiki/pages/),
// extracts the readable content from each page, chunks it by
// section heading, and returns a populated MemoryCorpus.
func LoadWikiDump(dir string) (*MemoryCorpus, LoadStats, error) {
	pattern := filepath.Join(dir, "*.html")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, LoadStats{}, fmt.Errorf("glob %s: %w", pattern, err)
	}
	if len(matches) == 0 {
		return nil, LoadStats{}, fmt.Errorf("no HTML pages found at %s — is the crawler output dir correct?", pattern)
	}
	var stats LoadStats
	var chunks []Chunk
	for _, m := range matches {
		raw, err := os.ReadFile(m)
		if err != nil {
			stats.FailedFiles++
			continue
		}
		title, sections := chunkWikiHTML(string(raw))
		if title == "" {
			stats.EmptyPages++
			continue
		}
		pageName := strings.TrimSuffix(filepath.Base(m), ".html")
		url := "https://classic.runescape.wiki/w/" + pageName
		for _, s := range sections {
			text := strings.TrimSpace(s.body)
			if len(text) < 40 {
				continue
			}
			chunks = append(chunks, Chunk{
				Source:       "rscwiki",
				PageTitle:    title,
				SectionTitle: s.heading,
				URL:          url,
				Text:         text,
			})
		}
		stats.LoadedFiles++
	}
	stats.Chunks = len(chunks)
	return LoadFromChunks(chunks), stats, nil
}

// LoadStats reports the outcome of a corpus load. Useful for
// boot-time INFO logs ("loaded 18403 chunks from 3625 pages").
type LoadStats struct {
	LoadedFiles int // pages that produced at least one chunk
	EmptyPages  int // pages where the content extractor found nothing
	FailedFiles int // pages that failed to read from disk
	Chunks      int // total chunks across all loaded pages
}

// section is the intermediate representation between the HTML
// extractor and the Chunk: a heading + the body that followed it
// until the next heading.
type section struct {
	heading string
	body    string
}

// chunkWikiHTML parses ONE MediaWiki (Vector skin) page with a real DOM parser
// and extracts clean, RAG-ready chunks:
//
//   - infoboxes (item facts, combat bonuses) → a "Facts" chunk of "Label: value"
//     pairs. (The previous regex extractor STRIPPED infoboxes, discarding the
//     single most valuable structured data on item/monster pages.)
//   - content tables (wikitable) → readable labeled rows ("Requirements:
//     Skill=Smithing, Level=1, XP=12.5") instead of glued number-soup
//     ("Smithing112.5"), preserving the column meaning.
//   - prose split by section heading.
//   - navboxes, TOC, references, edit links, figures/images, and sister-wiki
//     notices removed.
//
// Returns ("", nil) for pages with no extractable content (redirects, stubs).
func chunkWikiHTML(htmlSrc string) (pageTitle string, sections []section) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlSrc))
	if err != nil {
		return "", nil
	}
	pageTitle = clean(doc.Find("#firstHeading").First().Text())
	if pageTitle == "" {
		return "", nil
	}
	content := doc.Find(".mw-parser-output").First()
	if content.Length() == 0 {
		return pageTitle, nil
	}

	// Remove non-content chrome. Includes: hidden duplicate infoboxes (the
	// mobile/resource "switch" copies, which otherwise leak as labelless soup),
	// the references/reflist backlink lists, navboxes, TOC, edit links, figures,
	// and sister-wiki notices.
	content.Find(
		"script, style, sup.reference, .reference, .mw-editsection, .toc, #toc, " +
			".navbox, table.navbox, figure, .noprint, .rs-header-icon, .mw-empty-elt, " +
			"#catlinks, .printfooter, .rc-sidebar, " +
			".infobox-switch-resources, .hidden, " +
			".references, ol.references, .reflist, .mw-references-wrap, .mw-cite-backlink",
	).Remove()

	// Infoboxes → a single "Facts" chunk; then drop them from the body so they
	// don't leave an empty section behind.
	var facts []string
	content.Find("table.infobox").Each(func(_ int, t *goquery.Selection) {
		facts = append(facts, infoboxFacts(t)...)
	})
	content.Find("table.infobox").Remove()

	// Walk the remaining content in document order, splitting on headings and
	// converting wikitables to readable rows.
	cur := section{}
	flush := func() {
		if strings.TrimSpace(cur.body) != "" {
			sections = append(sections, cur)
		}
		cur = section{}
	}
	content.Children().Each(func(_ int, s *goquery.Selection) {
		if h := headingText(s); h != "" {
			flush()
			cur.heading = h
			return
		}
		if txt := nodeText(s); txt != "" {
			if cur.body != "" {
				cur.body += "\n"
			}
			cur.body += txt
		}
	})
	flush()

	if len(facts) > 0 {
		sections = append(sections, section{heading: "Facts", body: strings.Join(facts, "; ")})
	}
	return pageTitle, sections
}

// headingText returns the section heading if s is a heading node — a bare h2/h3
// or the modern `div.mw-heading` wrapper — else "".
func headingText(s *goquery.Selection) string {
	if s.Is("h2, h3, h4") {
		return clean(s.Text())
	}
	if s.HasClass("mw-heading") {
		return clean(s.Find("h2, h3, h4").First().Text())
	}
	return ""
}

// nodeText renders a content child to clean text. Tables (the child itself or
// any nested) become readable rows; everything else is its collapsed text.
func nodeText(s *goquery.Selection) string {
	if s.Is("table") {
		return tableToText(s)
	}
	if s.Find("table").Length() == 0 {
		return clean(s.Text())
	}
	// A wrapper that contains one or more tables: render in document order.
	var parts []string
	s.Contents().Each(func(_ int, c *goquery.Selection) {
		if c.Is("table") {
			if t := tableToText(c); t != "" {
				parts = append(parts, t)
			}
		} else if c.Find("table").Length() > 0 {
			if t := nodeText(c); t != "" {
				parts = append(parts, t)
			}
		} else if t := clean(c.Text()); t != "" {
			parts = append(parts, t)
		}
	})
	return strings.Join(parts, "\n")
}

// tableToText converts a table to "Caption: h1=c1, h2=c2" lines, preserving the
// column meaning the old flat-strip destroyed.
func tableToText(t *goquery.Selection) string {
	caption := clean(t.Find("caption").First().Text())
	var headers []string
	var lines []string
	t.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		var cells []string
		tr.Find("th, td").Each(func(_ int, c *goquery.Selection) {
			cells = append(cells, cellValue(c))
		})
		if strings.TrimSpace(strings.Join(cells, "")) == "" {
			return // fully-empty row
		}
		// Header row: the first row made entirely of <th>.
		if len(headers) == 0 && tr.Find("td").Length() == 0 {
			headers = cells
			return
		}
		var row string
		if len(headers) == len(cells) && len(headers) > 1 {
			var pairs []string
			for i, c := range cells {
				if c == "" {
					continue
				}
				pairs = append(pairs, fmt.Sprintf("%s=%s", headers[i], c))
			}
			row = strings.Join(pairs, ", ")
		} else {
			var nonEmpty []string
			for _, c := range cells {
				if c != "" {
					nonEmpty = append(nonEmpty, c)
				}
			}
			row = strings.Join(nonEmpty, ", ")
		}
		if row == "" {
			return
		}
		if caption != "" {
			row = caption + ": " + row
		}
		lines = append(lines, row)
	})
	return strings.Join(lines, "\n")
}

// infoboxFacts parses an infobox into "Label: value" facts.
func infoboxFacts(t *goquery.Selection) []string {
	var facts []string
	if c := clean(t.Find("caption").First().Text()); c != "" {
		facts = append(facts, c)
	}
	t.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		label := clean(tr.Find("th").First().Text())
		value := cellValue(tr.Find("td").First())
		if label != "" && value != "" {
			facts = append(facts, label+": "+value)
		}
	})
	return facts
}

// cellValue gets a cell's text, falling back to an icon's alt/title so
// icon-only cells (e.g. the free-to-play marker) become "Free-to-play".
func cellValue(c *goquery.Selection) string {
	if t := clean(c.Text()); t != "" {
		return t
	}
	if alt, ok := c.Find("img").First().Attr("alt"); ok && clean(alt) != "" {
		return clean(alt)
	}
	if title, ok := c.Find("a").First().Attr("title"); ok && clean(title) != "" {
		return clean(title)
	}
	return ""
}

var wsRE = regexp.MustCompile(`\s+`)

// clean collapses whitespace and trims.
func clean(s string) string {
	return strings.TrimSpace(wsRE.ReplaceAllString(s, " "))
}
