package corpus

import (
	"fmt"
	"html"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LoadWikiDump scans `dir` for `*.html` files (the output format of
// the rsc.wiki crawler at ~/Code/westworld/local/rscwiki/pages/),
// extracts the readable content from each page, chunks it by
// section heading, and returns a populated MemoryCorpus.
//
// Pages with empty content (redirects, navigation stubs) are
// skipped silently. Pages that fail to parse are skipped and the
// path is appended to the error message; a successful return means
// every chunk in the corpus came from a parseable page, but the
// caller may want to inspect the per-file warning count via the
// returned LoadStats.
//
// This is intentionally tolerant — the corpus is built once at
// host startup and any one bad page should not bring the host down.
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
			if len(text) < 60 {
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
// stripper and the Chunk: a heading + the body that followed it
// until the next heading.
type section struct {
	heading string
	body    string
}

// chunkWikiHTML extracts the page title and section chunks from
// one MediaWiki-formatted HTML document. Pure-stdlib regex
// stripping — adequate because MediaWiki output is regular.
//
// Returns ("", nil) when the page has no extractable content
// (redirect pages, navigation stubs).
func chunkWikiHTML(htmlSrc string) (pageTitle string, sections []section) {
	pageTitle = extractPageTitle(htmlSrc)
	if pageTitle == "" {
		return "", nil
	}
	body := extractContentBody(htmlSrc)
	if body == "" {
		return "", nil
	}
	body = stripChrome(body)
	return pageTitle, splitOnHeadings(body)
}

var (
	// MediaWiki's canonical page title lives in the firstHeading h1.
	firstHeadingRE = regexp.MustCompile(`(?is)<h1[^>]*id="firstHeading"[^>]*>(.*?)</h1>`)
	// Opening of the article-body wrapper. The class attribute often
	// holds extra classes (e.g. `class="mw-content-ltr mw-parser-output"`),
	// so we match the marker anywhere inside the attribute.
	parserOutputOpenRE = regexp.MustCompile(`(?is)<div[^>]*\bmw-parser-output\b[^>]*>`)
	// MediaWiki always emits the categories block AFTER article body.
	// We use it as the structural sentinel marking "end of article".
	catlinksOpenRE = regexp.MustCompile(`(?is)<div[^>]*id="catlinks"[^>]*>`)
)

func extractPageTitle(s string) string {
	m := firstHeadingRE.FindStringSubmatch(s)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(stripTags(m[1]))
}

// extractContentBody returns the HTML between the opening of the
// mw-parser-output div and the start of the catlinks div (the
// MediaWiki structural sentinel for "end of article"). This avoids
// having to balance arbitrary nested </div> tags inside content —
// catlinks is a stable marker every wiki page emits.
func extractContentBody(s string) string {
	openLoc := parserOutputOpenRE.FindStringIndex(s)
	if openLoc == nil {
		return ""
	}
	tail := s[openLoc[1]:]
	closeLoc := catlinksOpenRE.FindStringIndex(tail)
	if closeLoc == nil {
		// No catlinks → take everything left. Atypical but harmless.
		return tail
	}
	return tail[:closeLoc[0]]
}

// stripChrome removes infoboxes, navboxes, table of contents, edit
// links, references, and similar non-content furniture that
// MediaWiki pages are full of. Each pattern is anchored to a
// MediaWiki class or id selector so we're not guessing at structure.
var chromePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?is)<table[^>]*class="[^"]*infobox[^"]*"[^>]*>.*?</table>`),
	regexp.MustCompile(`(?is)<table[^>]*class="[^"]*navbox[^"]*"[^>]*>.*?</table>`),
	regexp.MustCompile(`(?is)<div[^>]*id="toc"[^>]*>.*?</div>`),
	regexp.MustCompile(`(?is)<div[^>]*class="[^"]*toc[^"]*"[^>]*>.*?</div>`),
	regexp.MustCompile(`(?is)<table[^>]*id="toc"[^>]*>.*?</table>`),
	regexp.MustCompile(`(?is)<style[^>]*>.*?</style>`),
	regexp.MustCompile(`(?is)<script[^>]*>.*?</script>`),
	regexp.MustCompile(`(?is)<sup[^>]*class="[^"]*reference[^"]*"[^>]*>.*?</sup>`),
	// MediaWiki's "[edit | edit source]" inline links — emitted next
	// to every section heading. They show up as "editedit source" in
	// stripped text otherwise, polluting nearly every chunk.
	regexp.MustCompile(`(?is)<span[^>]*class="[^"]*mw-editsection[^"]*"[^>]*>.*?</span>\s*</span>`),
	regexp.MustCompile(`(?is)<span[^>]*class="[^"]*mw-editsection[^"]*"[^>]*>.*?</span>`),
}

func stripChrome(s string) string {
	for _, re := range chromePatterns {
		s = re.ReplaceAllString(s, "")
	}
	return s
}

// headingRE matches h2/h3 anchors. MediaWiki wraps section headings
// like `<h2><span class="mw-headline" id="..">Title</span></h2>`.
var headingRE = regexp.MustCompile(`(?is)<h[23][^>]*>(.*?)</h[23]>`)

// splitOnHeadings carves the body into (heading → text-until-next-
// heading) sections. The intro before the first heading becomes a
// section with an empty heading (the page intro).
func splitOnHeadings(body string) []section {
	indices := headingRE.FindAllStringSubmatchIndex(body, -1)
	if len(indices) == 0 {
		// No headings at all — treat the whole body as a single
		// intro chunk.
		txt := plainText(body)
		if txt == "" {
			return nil
		}
		return []section{{heading: "", body: txt}}
	}
	var out []section
	// Intro: from start of body to first heading.
	intro := plainText(body[:indices[0][0]])
	if intro != "" {
		out = append(out, section{heading: "", body: intro})
	}
	for i, m := range indices {
		headingText := strings.TrimSpace(stripTags(body[m[2]:m[3]]))
		end := len(body)
		if i+1 < len(indices) {
			end = indices[i+1][0]
		}
		sectionBody := plainText(body[m[1]:end])
		if sectionBody == "" {
			continue
		}
		out = append(out, section{heading: headingText, body: sectionBody})
	}
	return out
}

// tagRE strips any HTML tag. Used after chrome has been removed —
// tagRE on raw HTML would discard structure we needed.
var tagRE = regexp.MustCompile(`(?is)<[^>]+>`)

func stripTags(s string) string {
	return tagRE.ReplaceAllString(s, "")
}

// whitespaceRE collapses any run of whitespace (incl. newlines) to a
// single space. Wiki content has erratic whitespace from template
// expansion.
var whitespaceRE = regexp.MustCompile(`\s+`)

// plainText turns an HTML fragment into clean, single-spaced text.
func plainText(s string) string {
	s = stripTags(s)
	s = html.UnescapeString(s)
	s = whitespaceRE.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}
