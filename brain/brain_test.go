package brain

import (
	"context"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/cognition"
)

func TestStubStrategist_Decide_Options(t *testing.T) {
	s := NewStubStrategist()
	d, err := s.Decide(context.Background(), Situation{
		Question: "pick one",
		Options:  []string{"north", "south", "east", "west"},
	})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Choice != "north" {
		t.Errorf("Choice = %q, want first option %q", d.Choice, "north")
	}
	if d.Confidence != 0.75 {
		t.Errorf("Confidence = %v, want 0.75", d.Confidence)
	}
	if !strings.Contains(d.Reasoning, "north") {
		t.Errorf("Reasoning should mention selected option; got %q", d.Reasoning)
	}
}

func TestStubStrategist_Decide_QuestionPrefixes(t *testing.T) {
	tests := []struct {
		name        string
		question    string
		wantChoice  string
		wantConfMin float64
		wantConfMax float64
	}{
		{"should i — yes", "should I attack the goblin", "yes", 0.55, 0.65},
		{"should we", "should we run", "yes", 0.55, 0.65},
		{"is — yes/no", "is the goblin dangerous", "yes", 0.5, 0.6},
		{"are", "are we safe here", "yes", 0.5, 0.6},
		{"can", "can I afford this", "yes", 0.5, 0.6},
		{"do", "do I have enough food", "yes", 0.5, 0.6},
		{"does", "does this work", "yes", 0.5, 0.6},
		{"will", "will I survive", "yes", 0.5, 0.6},
		{"where", "where should I fish today", "Lumbridge", 0.45, 0.55},
		{"what — open", "what should I say back", "I would explore the world and stay alive.", 0.45, 0.55},
		{"how", "how do I get to Falador", "I would explore the world and stay alive.", 0.45, 0.55},
		{"why", "why is this happening", "I would explore the world and stay alive.", 0.45, 0.55},
		{"fallback", "completely unrelated", "ok", 0.35, 0.45},
	}
	s := NewStubStrategist()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d, err := s.Decide(context.Background(), Situation{Question: tt.question})
			if err != nil {
				t.Fatalf("Decide: %v", err)
			}
			if d.Choice != tt.wantChoice {
				t.Errorf("Choice = %q, want %q", d.Choice, tt.wantChoice)
			}
			if d.Confidence < tt.wantConfMin || d.Confidence > tt.wantConfMax {
				t.Errorf("Confidence = %v, want in [%v, %v]",
					d.Confidence, tt.wantConfMin, tt.wantConfMax)
			}
			if d.Reasoning == "" {
				t.Errorf("Reasoning should be non-empty")
			}
		})
	}
}

func TestStubStrategist_Decide_BundleSummary(t *testing.T) {
	s := NewStubStrategist()

	t.Run("nil bundle", func(t *testing.T) {
		d, err := s.Decide(context.Background(), Situation{Question: "should I fish"})
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		if !strings.Contains(d.Reasoning, "no bundle supplied") {
			t.Errorf("Reasoning should note nil bundle; got %q", d.Reasoning)
		}
	})

	t.Run("populated bundle", func(t *testing.T) {
		cc := cognition.NewStubClient()
		b, err := cc.Retrieve(context.Background(), cognition.Retrieval{Goal: "combat"})
		if err != nil {
			t.Fatalf("cognition Retrieve: %v", err)
		}
		d, err := s.Decide(context.Background(), Situation{
			Question: "should I attack",
			Bundle:   b,
		})
		if err != nil {
			t.Fatalf("Decide: %v", err)
		}
		// Bundle had 3 reflections and 3 episodes from the combat
		// branch of the cognition stub.
		if !strings.Contains(d.Reasoning, "3 reflection") {
			t.Errorf("Reasoning should mention 3 reflections; got %q", d.Reasoning)
		}
		if !strings.Contains(d.Reasoning, "3 episode") {
			t.Errorf("Reasoning should mention 3 episodes; got %q", d.Reasoning)
		}
	})
}

func TestStubStrategist_Decide_OptionsBeatQuestionPrefix(t *testing.T) {
	// When both options and a prefix that would match are present,
	// Options wins.
	s := NewStubStrategist()
	d, err := s.Decide(context.Background(), Situation{
		Question: "should I attack",
		Options:  []string{"flee", "fight", "hide"},
	})
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Choice != "flee" {
		t.Errorf("Choice = %q, want first option %q (Options must beat prefix heuristics)", d.Choice, "flee")
	}
}

func TestStubStrategist_Decide_ContextCancellation(t *testing.T) {
	s := NewStubStrategist()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := s.Decide(ctx, Situation{Question: "should I fish"})
	if err == nil {
		t.Errorf("expected context cancellation error, got nil")
	}
}

func TestStubStrategist_Decide_Deterministic(t *testing.T) {
	s := NewStubStrategist()
	sit := Situation{Question: "should I attack"}
	d1, _ := s.Decide(context.Background(), sit)
	d2, _ := s.Decide(context.Background(), sit)
	if d1.Choice != d2.Choice || d1.Confidence != d2.Confidence || d1.Reasoning != d2.Reasoning {
		t.Errorf("Decisions not deterministic:\n  d1=%+v\n  d2=%+v", d1, d2)
	}
}

func TestStubStrategist_ConcurrentDecide(t *testing.T) {
	s := NewStubStrategist()
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := s.Decide(context.Background(), Situation{Question: "should I attack"})
			done <- err
		}()
	}
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent Decide: %v", err)
		}
	}
}

// Compile-time check that StubStrategist satisfies Strategist.
var _ Strategist = (*StubStrategist)(nil)
