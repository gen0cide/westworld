// Command scenariogen emits .routine files into examples/scenarios/
// from scenarios.yaml. Each entry becomes one file under
// examples/scenarios/<category>/<id>.routine.
//
// The YAML format is intentionally simple — edits don't require a
// recompile, and the schema is documented at the top of
// scenarios.yaml itself.
//
// Optional grounding: when -corpus points at the rsc.wiki HTML
// dump, the generator runs each scenario's recall_query against
// the corpus and stamps the top hits into the file header. That's
// the dogfood: each scenario cites the player-facing knowledge
// that informed its design.
//
// Run:
//
//	go run ./cmd/scenariogen                # emit everything
//	go run ./cmd/scenariogen -id cooking-…  # only this one
//	go run ./cmd/scenariogen -dry           # print, don't write
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gen0cide/westworld/cognition/corpus"
)

// Scenario mirrors the YAML shape. See scenarios.yaml for the
// field-by-field semantics.
type Scenario struct {
	ID           string   `yaml:"id"`
	Category     string   `yaml:"category"`
	Hosts        string   `yaml:"hosts"`
	Admin        bool     `yaml:"admin"`
	RecallQuery  string   `yaml:"recall_query"`
	Precondition []string `yaml:"precondition"`
	Setup        []string `yaml:"setup"`
	Body         []string `yaml:"body"`
	PassExpr     string   `yaml:"pass"`
	Timeout      int      `yaml:"timeout"`
	Notes        string   `yaml:"notes"`
}

type catalog struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

// snake converts kebab-case → snake_case for the routine name. The
// filename ↔ routine-name validator enforces this match.
func snake(id string) string { return strings.ReplaceAll(id, "-", "_") }

func main() {
	var (
		yamlPath  = flag.String("yaml", "cmd/scenariogen/scenarios.yaml", "Scenario catalog")
		outRoot   = flag.String("out", "examples/scenarios", "Output root directory")
		corpusDir = flag.String("corpus", "local/rscwiki/pages", "rsc.wiki HTML dump for grounding (empty disables)")
		onlyID    = flag.String("id", "", "Only emit this one scenario by ID")
		dryRun    = flag.Bool("dry", false, "Print what would be written, don't write")
	)
	flag.Parse()

	raw, err := os.ReadFile(*yamlPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read %s: %v\n", *yamlPath, err)
		os.Exit(1)
	}
	var cat catalog
	if err := yaml.Unmarshal(raw, &cat); err != nil {
		fmt.Fprintf(os.Stderr, "parse %s: %v\n", *yamlPath, err)
		os.Exit(1)
	}

	var c corpus.Corpus
	if *corpusDir != "" {
		mc, stats, err := corpus.LoadWikiDump(*corpusDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: corpus load failed (%v); proceeding without grounding\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "loaded %d chunks for grounding (%d pages)\n", stats.Chunks, stats.LoadedFiles)
			c = mc
		}
	}

	var wrote, skipped int
	for _, s := range cat.Scenarios {
		if *onlyID != "" && s.ID != *onlyID {
			continue
		}
		body := render(s, c)
		// Filename uses snake_case to match the routine name — the
		// loader enforces `filepath.Base(path) without .routine ==
		// file.Routine.Name`, and identifiers can't have hyphens.
		path := filepath.Join(*outRoot, s.Category, snake(s.ID)+".routine")
		if *dryRun {
			fmt.Printf("=== %s ===\n%s\n", path, body)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "%s: mkdir: %v\n", path, err)
			skipped++
			continue
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "%s: write: %v\n", path, err)
			skipped++
			continue
		}
		wrote++
	}
	fmt.Fprintf(os.Stderr, "wrote %d, skipped %d\n", wrote, skipped)
}

// render produces the final .routine file contents for one scenario,
// including the header comments + grounding chunks + DSL body.
func render(s Scenario, c corpus.Corpus) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# Scenario: %s\n", s.ID)
	fmt.Fprintf(&sb, "# Category: %s\n", s.Category)
	if s.Hosts != "" {
		fmt.Fprintf(&sb, "# Hosts: %s\n", s.Hosts)
	}
	if s.Admin {
		sb.WriteString("# Admin: required\n")
	}
	if s.Timeout > 0 {
		fmt.Fprintf(&sb, "# Timeout: %ds\n", s.Timeout)
	}
	if s.Notes != "" {
		fmt.Fprintf(&sb, "#\n# Notes: %s\n", s.Notes)
	}
	if s.RecallQuery != "" && c != nil {
		hits, _ := c.Recall(context.Background(), s.RecallQuery, 2)
		if len(hits) > 0 {
			sb.WriteString("#\n# Wiki grounding (query: ")
			sb.WriteString(s.RecallQuery)
			sb.WriteString("):\n")
			for _, h := range hits {
				title := h.PageTitle
				if h.SectionTitle != "" {
					title += " § " + h.SectionTitle
				}
				fmt.Fprintf(&sb, "#   - %s — %s\n", title, h.URL)
			}
		}
	}
	sb.WriteString("\n")

	fmt.Fprintf(&sb, "routine %s() {\n", snake(s.ID))

	if len(s.Precondition) > 0 {
		sb.WriteString("    require {\n")
		for _, p := range s.Precondition {
			fmt.Fprintf(&sb, "        %s\n", p)
		}
		sb.WriteString("    }\n\n")
	}

	for _, cmd := range s.Setup {
		fmt.Fprintf(&sb, "    command(%q)\n", cmd)
		sb.WriteString("    wait 0.5\n")
	}
	if len(s.Setup) > 0 {
		sb.WriteString("\n")
	}

	for _, line := range s.Body {
		// A body item may be a single statement OR a multi-line block
		// literal (YAML `|-`) holding a select{}/when{}/if{} construct.
		// Re-indent every line in the item to the routine-body level
		// (4 spaces) so the emitted .routine is readable + parseable.
		for _, sub := range strings.Split(line, "\n") {
			if sub == "" {
				sb.WriteString("\n")
				continue
			}
			fmt.Fprintf(&sb, "    %s\n", sub)
		}
	}

	if s.PassExpr != "" {
		// Plain string literals — no interpolation needed, and the
		// DSL's f-string lexer doesn't handle f-strings with zero
		// `{...}` placeholders cleanly.
		fmt.Fprintf(&sb, "\n    if %s {\n", s.PassExpr)
		fmt.Fprintf(&sb, "        return \"PASS: %s\"\n    }\n", s.ID)
		fmt.Fprintf(&sb, "    abort \"FAIL: %s (predicate was false)\"\n", s.ID)
	}

	sb.WriteString("}\n")
	return sb.String()
}
