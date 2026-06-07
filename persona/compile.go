package persona

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// This file is the persona compiler seam — kept in the shared persona package so
// host and mesa use the same types.
//
// Two distinct operations (see docs/persona-compile.md):
//
//   - COOK (birth, once): structured Persona -> sealed prose card. The quality
//     path is a best-of-N Opus cook + judge, which lives MESA-SIDE as an impl of
//     the ProseCook interface. The deterministic Render() floor here is the
//     no-LLM fallback (keeps go test green, covers un-cooked hosts).
//   - PROJECT (login, every time): Persona -> PersonaCard. Deterministic, NO LLM
//     — re-inject the sealed prose + a template mood line. (The policy half of
//     the card — event policy / choice weights / directives / reverie — comes
//     from the band->policy compiler, deferred to that design session.)

// PersonaCard is the compiled artifact the runtime executes. The brain sees only
// Prose. Hydrated once per host at login.
type PersonaCard struct {
	HostName   string    `json:"host_name"`
	Prose      string    `json:"prose"`     // the sealed second-person card — the ONLY thing the brain reads
	Pinned     []string  `json:"pinned"`    // foundational memory summaries, always prepended
	MoodLine   string    `json:"mood_line"` // deterministic per-call affect line
	Affect     Affect    `json:"affect"`    // current snapshot
	CompiledAt time.Time `json:"compiled_at"`
	// TODO(band->policy session): EventPolicy, ChoiceWeights, Directives, ReverieKernel.
}

// WorldBrief is the CURATED world context fed to the cook — cohort home +
// activity + a few relevant entities, distilled from the wiki OFFLINE. The cook
// does NOT do live wiki RAG (the card is about WHO, not HOW); live retrieval is
// a runtime (Recall/Act) resource.
type WorldBrief struct {
	Setting  string   `json:"setting"`  // e.g. "RuneScape Classic, ~2001"
	Home     string   `json:"home"`     // cohort home area
	Activity string   `json:"activity"` // activity bias
	Entities []string `json:"entities"` // relevant places/items/skills (wiki-distilled)
}

// Cooked is the result of a prose cook: the winning card + audit meta.
type Cooked struct {
	Prose string   `json:"prose"`
	Meta  CookMeta `json:"meta"`
}

// CookMeta is the generation audit trail, sealed alongside the card (pinned for
// reproducibility + re-cook discipline; there is no API seed).
type CookMeta struct {
	ProseModelID       string `json:"prose_model_id"`       // e.g. "claude-opus-4-8" or "deterministic-render"
	ProsePromptVersion string `json:"prose_prompt_version"` // e.g. "compile-v1"
	CandidatesN        int    `json:"prose_candidates_n"`
	JudgeWinner        int    `json:"prose_judge_winner"`
	JudgeScores        []int  `json:"prose_judge_scores,omitempty"`
}

// ProseCook turns a structured persona into a sealed prose card. The quality
// implementation (best-of-N Opus cook + Opus judge) is mesa-side; this package
// ships the DeterministicCook floor. Run ONCE at birth, offline.
type ProseCook interface {
	Cook(ctx context.Context, p *Persona, brief WorldBrief) (Cooked, error)
}

// DeterministicCook is the no-LLM floor: it renders the structured fields with a
// fixed trait lexicon. Faithful-if-flatter; never a dependency for the system to
// run, and the reproducibility floor the best-of-N cook upgrades.
type DeterministicCook struct{}

func (DeterministicCook) Cook(_ context.Context, p *Persona, _ WorldBrief) (Cooked, error) {
	return Cooked{
		Prose: Render(p),
		Meta:  CookMeta{ProseModelID: "deterministic-render", ProsePromptVersion: "render-v1", CandidatesN: 1, JudgeWinner: 0},
	}, nil
}

// Project hydrates the runtime PersonaCard from a stored persona (login-time,
// deterministic, NO LLM): it prefers the sealed cooked card and falls back to
// the deterministic Render() floor when none was cooked.
func Project(hostName string, p *Persona) *PersonaCard {
	prose := strings.TrimSpace(p.Cornerstone.Identity.Backstory)
	if prose == "" {
		prose = Render(p)
	}
	pinned := make([]string, 0, len(p.Cornerstone.Pinned))
	for _, m := range p.Cornerstone.Pinned {
		pinned = append(pinned, m.Summary)
	}
	return &PersonaCard{
		HostName: hostName,
		Prose:    prose,
		Pinned:   pinned,
		MoodLine: MoodLine(p.Trajectory.Mood),
		Affect:   p.Trajectory.Mood,
	}
}

// MoodLine renders the drifting affect vector into the one-line directive the
// prompt prepends ("You currently feel: …"). Deterministic template lookup, NOT
// an LLM call — it changes every inference, so it must stay off the model path.
func MoodLine(a Affect) string {
	var arousal, valence string
	switch {
	case a.Stress > 0.6:
		arousal = "tense"
	case a.Arousal > 0.6:
		arousal = "energized"
	case a.Arousal < 0.3:
		arousal = "subdued"
	default:
		arousal = "steady"
	}
	switch {
	case a.Valence > 0.3 && a.Confidence > 0.6:
		valence = "pleased and self-assured"
	case a.Valence > 0.1:
		valence = "content"
	case a.Valence < -0.3:
		valence = "discouraged"
	case a.Confidence < 0.3:
		valence = "uncertain"
	default:
		valence = "even"
	}
	return fmt.Sprintf("You currently feel %s and %s.", arousal, valence)
}
