package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/persona"
)

// cookSystemPrompt is the deep "Persona Compiler" system prompt — the one-time,
// offline instruction that turns a structured persona into the sealed prose card
// the brain reads. (This is the mesa-side Opus impl of persona.ProseCook.)
const cookSystemPrompt = `You are compiling a character for a role-playing simulation set in RuneScape Classic (a 2001 medieval-fantasy MMORPG). You will be given a fully-specified persona as structured data. Produce the SECOND-PERSON behavioral prose card that another AI model will be handed as "who you are" before it acts in the game.

Hard rules:
- Reflect EVERY structured field faithfully: the HEXACO disposition (honesty, emotionality, extraversion, agreeableness, conscientiousness, openness), the values, the domain risk (economic/bodily/social), patience, cooperation style, the additive dials (aggression, decisiveness, tenacity, bulk apperception = how keenly the character learns from experience), the north star, the voice, and the quirks.
- The pinned memories are formative. Weave them into the character's outlook. They may sit in TENSION with the surface disposition — a gentle soul with something stirring underneath. Honor that tension; do not flatten or resolve it.
- Do NOT invent traits, history, relationships, or events beyond what the persona licenses.
- Write it AS the character's standing disposition and outlook, NOT as a stat block or a list of traits. The reader must NEVER see a number, a band word (such as "high", "low", "very_high", "mid"), the word "HEXACO", or any archetype tag. Translate everything into plain behavioral language.
- Match the specified voice: register, formality, and typo feel. (If typos are frequent, the card may itself read a little informally; if none, it reads clean.)
- Length: ~120-200 words of second-person prose. Then a short "Things you never forget:" list drawn ONLY from the pinned memories.
- Output ONLY the card: the prose paragraph(s) followed by the "Things you never forget:" list. No preamble, no headings, no commentary, no quotes around the whole thing.`

// judgeSystemPrompt ranks candidate cards for faithfulness. Low temp; it IS the
// verification gate (no separate NLI stage).
const judgeSystemPrompt = `You are judging candidate character cards against the structured persona they were generated from. Score each candidate and pick the most faithful.

Score each candidate 0-10 on each criterion:
- coverage: every structured field is reflected (HEXACO, values, domain risk, patience, cooperation, the additive dials, north star, quirks, voice, and the pinned memories).
- non_contradiction: nothing in the prose contradicts a field (e.g. no "generous and trusting" for a low-honesty exploiter).
- no_invention: nothing fabricated beyond what the persona licenses.
- voice_match: register, formality, and typo feel match the voice block.
- no_leakage: no numbers, band words ("high"/"very_low"/etc.), the word "HEXACO", or archetype tags appear.

Pick the candidate with the highest total. Note any fidelity violations in the winner.

Output ONLY valid JSON, no prose:
{"best_index": <int>, "scores": [{"i":0,"coverage":0,"non_contradiction":0,"no_invention":0,"voice_match":0,"no_leakage":0,"total":0}], "winner_violations": []}`

// worldBriefText renders the curated world context block.
func worldBriefText(b persona.WorldBrief) string {
	return fmt.Sprintf("Setting: %s\nHome: %s\nActivity bias: %s\nRelevant places/things: %s",
		b.Setting, b.Home, b.Activity, strings.Join(b.Entities, ", "))
}

// opusCook is the mesa-side, Opus-backed implementation of persona.ProseCook.
type opusCook struct {
	a      *anthropicClient
	maxTok int
}

// cookOne generates a single candidate card (one Opus call; the model's default
// sampling supplies the best-of-N variation).
func (c *opusCook) cookOne(ctx context.Context, p *persona.Persona, brief persona.WorldBrief) (string, error) {
	pj, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	user := fmt.Sprintf("PERSONA (structured JSON):\n%s\n\nWORLD CONTEXT:\n%s", pj, worldBriefText(brief))
	out, err := c.a.complete(ctx, cookSystemPrompt, user, c.maxTok)
	return strings.TrimSpace(out), err
}

// judgeResult is the parsed judge output.
type judgeResult struct {
	BestIndex        int              `json:"best_index"`
	Scores           []map[string]int `json:"scores"`
	WinnerViolations []string         `json:"winner_violations"`
}

// judge ranks the candidate cards (one Opus call).
func (c *opusCook) judge(ctx context.Context, p *persona.Persona, cards []string) (*judgeResult, error) {
	pj, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return nil, err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "PERSONA (structured JSON):\n%s\n\nCANDIDATES:\n", pj)
	for i, card := range cards {
		fmt.Fprintf(&b, "[%d]\n%s\n\n", i, card)
	}
	out, err := c.a.complete(ctx, judgeSystemPrompt, b.String(), 2048)
	if err != nil {
		return nil, err
	}
	// Tolerate ```json fences / stray prose around the JSON object.
	out = extractJSON(out)
	var jr judgeResult
	if err := json.Unmarshal([]byte(out), &jr); err != nil {
		return nil, fmt.Errorf("parse judge JSON: %w; raw: %s", err, truncate(out, 400))
	}
	return &jr, nil
}

// extractJSON pulls the first {...} object out of a possibly-fenced response.
func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i >= 0 {
		if j := strings.LastIndex(s, "}"); j > i {
			return s[i : j+1]
		}
	}
	return s
}
