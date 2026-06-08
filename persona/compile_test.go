package persona

import (
	"context"
	"strings"
	"testing"
)

func TestRenderFloorIsFaithfulAndLeakFree(t *testing.T) {
	p := validPersona() // cautious high-H grinder
	card := Render(p)
	if card == "" {
		t.Fatal("Render produced empty card")
	}
	// Faithful: a high-honesty, low-economic-risk, skill-mastery host.
	for _, want := range []string{"marn_fishes", "fair and sincere", "wary with your money", "99 fishing"} {
		if !strings.Contains(card, want) {
			t.Errorf("card missing %q\n---\n%s", want, card)
		}
	}
	// Leak-free: no numbers, band words, or the archetype tag.
	for _, leak := range []string{"very_high", "high", "mid", "0.", "cautious_social_grinder", "hexaco", "lambda"} {
		if strings.Contains(card, leak) {
			t.Errorf("card LEAKED %q (code-food must not reach the brain)\n---\n%s", leak, card)
		}
	}
}

func TestRenderCuriosity(t *testing.T) {
	// A dominant Spatial pull must surface as plain behaviour in the card.
	p := validPersona()
	p.Cornerstone.Prefs.Curiosity = Curiosity{Spatial: 0.8, Skill: 0.1, Social: 0.05}
	card := Render(p)
	if !strings.Contains(card, "exploring new places") {
		t.Errorf("dominant Spatial curiosity not rendered:\n%s", card)
	}
	for _, leak := range []string{"0.", "spatial", "very_high"} {
		if strings.Contains(card, leak) {
			t.Errorf("curiosity rendering LEAKED %q:\n%s", leak, card)
		}
	}
	// Flat / middling curiosity is omitted, keeping the card sharp.
	p2 := validPersona()
	p2.Cornerstone.Prefs.Curiosity = Curiosity{Spatial: 0.2, Skill: 0.2, Social: 0.2, Economic: 0.2, Risk: 0.2}
	card2 := Render(p2)
	for _, flavor := range []string{"exploring new places", "mastering new skills", "turning a profit", "hearing their stories", "testing yourself"} {
		if strings.Contains(card2, flavor) {
			t.Errorf("middling curiosity should be omitted, but found %q", flavor)
		}
	}
}

func TestRenderPinnedMemories(t *testing.T) {
	p := validPersona()
	p.Cornerstone.Pinned = []FoundationalMemory{{Summary: "Lost my first lobster stack to a fake trade.", Weight: 1}}
	card := Render(p)
	if !strings.Contains(card, "Things you never forget:") || !strings.Contains(card, "lobster stack") {
		t.Fatalf("pinned memories not rendered:\n%s", card)
	}
}

func TestProjectPrefersSealedProse(t *testing.T) {
	p := validPersona()
	p.Cornerstone.Identity.Backstory = "A sealed, LLM-cooked card."
	card := Project("marn_fishes", p)
	if card.Prose != "A sealed, LLM-cooked card." {
		t.Fatalf("Project should re-inject the sealed card, got: %q", card.Prose)
	}
	if card.HostName != "marn_fishes" || card.MoodLine == "" {
		t.Fatalf("Project card incomplete: %+v", card)
	}
}

func TestProjectFallsBackToRender(t *testing.T) {
	p := validPersona() // no sealed Backstory
	p.Cornerstone.Identity.Backstory = ""
	card := Project("marn_fishes", p)
	if !strings.Contains(card.Prose, "marn_fishes") {
		t.Fatalf("Project should fall back to Render() when uncooked:\n%s", card.Prose)
	}
}

func TestDeterministicCookIsOfflineFloor(t *testing.T) {
	p := validPersona()
	cooked, err := DeterministicCook{}.Cook(context.Background(), p, WorldBrief{})
	if err != nil {
		t.Fatalf("DeterministicCook: %v", err)
	}
	if cooked.Prose == "" || cooked.Meta.ProseModelID != "deterministic-render" {
		t.Fatalf("deterministic cook meta wrong: %+v", cooked.Meta)
	}
	// Compile-time proof it satisfies the shared interface mesa will implement.
	var _ ProseCook = DeterministicCook{}
}

func TestMoodLineDeterministic(t *testing.T) {
	tense := MoodLine(Affect{Stress: 0.8, Valence: -0.5})
	if !strings.Contains(tense, "tense") || !strings.Contains(tense, "discouraged") {
		t.Fatalf("tense/discouraged mood = %q", tense)
	}
	happy := MoodLine(Affect{Valence: 0.5, Confidence: 0.8, Arousal: 0.7})
	if !strings.Contains(happy, "energized") || !strings.Contains(happy, "self-assured") {
		t.Fatalf("upbeat mood = %q", happy)
	}
}
