// Command personacook runs the persona prose-card cook N times against a stored
// persona and (optionally) judges them — so you can SEE the best-of-N variation.
// mesa-side, key-gated: needs ANTHROPIC_API_KEY from the environment.
//
//	source .local.env && go run ./mesa/personacook -persona docs/personas/dolores.json -n 20
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/persona"
)

func main() {
	personaPath := flag.String("persona", "docs/personas/dolores.json", "path to a stored persona JSON")
	n := flag.Int("n", 20, "number of candidate cards to cook")
	model := flag.String("model", "claude-opus-4-8", "Anthropic model id")
	doJudge := flag.Bool("judge", true, "run the judge to pick the most faithful card")
	maxTok := flag.Int("max-tokens", 700, "max tokens per card")
	outPath := flag.String("out", "", "also write a markdown report (all cards + judge pick) to this path")
	conc := flag.Int("concurrency", 5, "concurrent cook calls")
	flag.Parse()

	// Load + validate the persona first (no network), so a bad file fails fast
	// and a no-key dry run still confirms the persona is valid.
	raw, err := os.ReadFile(*personaPath)
	check(err)
	var p persona.Persona
	check(json.Unmarshal(raw, &p))
	if err := p.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "persona %s is invalid:\n%v\n", *personaPath, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "persona %s OK (%s)\n", *personaPath, p.Cornerstone.Identity.Name)

	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "personacook: set ANTHROPIC_API_KEY (e.g. source .local.env) to run the cook")
		os.Exit(2)
	}

	brief := lumbridgeBrief()
	cook := &opusCook{a: newAnthropic(key, *model), maxTok: *maxTok}
	ctx := context.Background()

	fmt.Printf("=== cooking %d candidate cards for %q (model %s) ===\n\n",
		*n, p.Cornerstone.Identity.Name, *model)

	cards := make([]string, *n)
	errs := make([]error, *n)
	sem := make(chan struct{}, *conc)
	var wg sync.WaitGroup
	for i := 0; i < *n; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(i int) {
			defer wg.Done()
			defer func() { <-sem }()
			cards[i], errs[i] = cook.cookOne(ctx, &p, brief)
		}(i)
	}
	wg.Wait()

	for i := 0; i < *n; i++ {
		fmt.Printf("──────── candidate [%d] ────────\n", i)
		if errs[i] != nil {
			fmt.Printf("(error: %v)\n\n", errs[i])
			continue
		}
		fmt.Printf("%s\n\n", cards[i])
	}

	var jr *judgeResult
	winnerCand := -1 // original candidate index of the judge's pick
	if *doJudge {
		var ok []string
		var idx []int
		for i, c := range cards {
			if errs[i] == nil && c != "" {
				ok = append(ok, c)
				idx = append(idx, i)
			}
		}
		switch {
		case len(ok) == 0:
			fmt.Println("no successful cards to judge")
		default:
			r, err := cook.judge(ctx, &p, ok)
			switch {
			case err != nil:
				fmt.Fprintf(os.Stderr, "judge: %v\n", err)
			case r.BestIndex < 0 || r.BestIndex >= len(ok):
				fmt.Fprintf(os.Stderr, "judge returned out-of-range best_index %d\n", r.BestIndex)
			default:
				jr = r
				winnerCand = idx[r.BestIndex]
				fmt.Printf("════════ JUDGE PICK: candidate [%d] ════════\n\n%s\n\n", winnerCand, ok[r.BestIndex])
				if b, e := json.MarshalIndent(r.Scores, "", "  "); e == nil {
					fmt.Printf("scores:\n%s\n", b)
				}
				if len(r.WinnerViolations) > 0 {
					fmt.Printf("winner violations: %v\n", r.WinnerViolations)
				}
			}
		}
	}

	if *outPath != "" {
		if err := writeReport(*outPath, &p, *model, brief, cards, errs, jr, winnerCand); err != nil {
			fmt.Fprintf(os.Stderr, "write report: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "wrote markdown report: %s\n", *outPath)
		}
	}
}

// writeReport emits a markdown report: every candidate (so you can scan the
// jitter), the judge's pick, and the score table.
func writeReport(path string, p *persona.Persona, model string, brief persona.WorldBrief, cards []string, errs []error, jr *judgeResult, winner int) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# %s — %d× prose-card cook\n\n", p.Cornerstone.Identity.Name, len(cards))
	fmt.Fprintf(&b, "- **Model:** %s (default sampling)\n", model)
	fmt.Fprintf(&b, "- **Archetype:** %s · north-star: %s\n", p.Cornerstone.Identity.ArchetypeTag, p.Cornerstone.Identity.NorthStar.Theme)
	if winner >= 0 {
		fmt.Fprintf(&b, "- **Judge pick:** candidate [%d]\n", winner)
	}

	// The prompts used to cook + judge, so the report is self-explaining.
	b.WriteString("\n## How these were cooked\n\n")
	b.WriteString("**Cook system prompt** (best-of-N, one call per candidate):\n\n```text\n")
	b.WriteString(cookSystemPrompt)
	b.WriteString("\n```\n\n**User message** (per candidate): the structured persona JSON, then the curated world context:\n\n```text\n")
	b.WriteString("PERSONA (structured JSON):\n<the persona>\n\nWORLD CONTEXT:\n")
	b.WriteString(worldBriefText(brief))
	b.WriteString("\n```\n\n**Judge system prompt** (ranks all candidates):\n\n```text\n")
	b.WriteString(judgeSystemPrompt)
	b.WriteString("\n```\n")

	b.WriteString("\nAll candidates below — scan for jitter.\n")
	for i, card := range cards {
		marker := ""
		if i == winner {
			marker = " 🏆"
		}
		fmt.Fprintf(&b, "\n---\n\n## Candidate %d%s\n\n", i, marker)
		if errs[i] != nil {
			fmt.Fprintf(&b, "_(error: %v)_\n", errs[i])
			continue
		}
		b.WriteString(card)
		b.WriteString("\n")
	}
	if jr != nil {
		b.WriteString("\n---\n\n## Scores\n\n```json\n")
		sc, _ := json.MarshalIndent(jr.Scores, "", "  ")
		b.Write(sc)
		b.WriteString("\n```\n")
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func lumbridgeBrief() persona.WorldBrief {
	return persona.WorldBrief{
		Setting:  "RuneScape Classic, ~2001 — a medieval-fantasy world of skills, monsters, quests, and player trade.",
		Home:     "Lumbridge, a quiet starter farming town.",
		Activity: "farming, fishing, and cooking around Lumbridge; wandering the countryside.",
		Entities: []string{
			"Lumbridge Castle", "the cow field", "the chicken coop", "the River Lum",
			"the general store", "Al Kharid to the east", "the road north to Varrock", "goblins on the plains",
		},
	}
}

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
