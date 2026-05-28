package cognition

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestStubClient_Retrieve_GoalHeuristics(t *testing.T) {
	tests := []struct {
		name             string
		goal             string
		wantReflectionHas string // substring required in at least one reflection
		wantEpisodicHas   string // substring required in at least one episodic entry
	}{
		{
			name:              "combat goal",
			goal:              "should I engage in combat with this goblin",
			wantReflectionHas: "aggressive",
			wantEpisodicHas:   "goblin",
		},
		{
			name:              "fight synonym",
			goal:              "is it safe to fight here",
			wantReflectionHas: "aggressive",
			wantEpisodicHas:   "goblin",
		},
		{
			name:              "bank goal",
			goal:              "should I bank now",
			wantReflectionHas: "bank",
			wantEpisodicHas:   "deposited",
		},
		{
			name:              "chat goal",
			goal:              "how should I greet this stranger",
			wantReflectionHas: "friendly",
			wantEpisodicHas:   "Lumbridge",
		},
		{
			name:              "skill goal — fish",
			goal:              "where should I fish today",
			wantReflectionHas: "skilling",
			wantEpisodicHas:   "trout",
		},
		{
			name:              "fallback goal",
			goal:              "what is the meaning of life",
			wantReflectionHas: "wander",
			wantEpisodicHas:   "Lumbridge",
		},
	}

	c := NewStubClient()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := c.Retrieve(context.Background(), Retrieval{
				Goal:     tt.goal,
				HostName: "test-host",
			})
			if err != nil {
				t.Fatalf("Retrieve: %v", err)
			}
			if b == nil {
				t.Fatalf("Retrieve returned nil bundle")
			}
			if b.Goal != tt.goal {
				t.Errorf("Bundle.Goal = %q, want %q", b.Goal, tt.goal)
			}
			if !anyContains(b.Reflections, tt.wantReflectionHas) {
				t.Errorf("no reflection contains %q; got %v", tt.wantReflectionHas, b.Reflections)
			}
			if !anyContains(b.Episodic, tt.wantEpisodicHas) {
				t.Errorf("no episodic contains %q; got %v", tt.wantEpisodicHas, b.Episodic)
			}
			if len(b.Persona) == 0 {
				t.Errorf("Persona should be populated")
			}
			if b.WorldSnapshot != "" {
				t.Errorf("WorldSnapshot should be empty unless IncludeWorld; got %q", b.WorldSnapshot)
			}
		})
	}
}

func TestStubClient_Retrieve_IncludeWorld(t *testing.T) {
	c := NewStubClient()
	b, err := c.Retrieve(context.Background(), Retrieval{
		Goal:         "combat",
		HostName:     "dolores",
		IncludeWorld: true,
	})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if b.WorldSnapshot == "" {
		t.Errorf("WorldSnapshot should be populated when IncludeWorld=true")
	}
	if !strings.Contains(b.WorldSnapshot, "dolores") {
		t.Errorf("WorldSnapshot should mention host name; got %q", b.WorldSnapshot)
	}
}

func TestStubClient_Retrieve_MaxItems(t *testing.T) {
	c := NewStubClient()
	tests := []struct {
		name    string
		max     int
		wantLen int
	}{
		{"zero uses default", 0, 3},   // combat has 3 reflections, 3 episodes
		{"explicit 1", 1, 1},
		{"explicit 2", 2, 2},
		{"larger than canned", 10, 3}, // capped to actual length
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := c.Retrieve(context.Background(), Retrieval{
				Goal:     "combat",
				MaxItems: tt.max,
			})
			if err != nil {
				t.Fatalf("Retrieve: %v", err)
			}
			if len(b.Reflections) != tt.wantLen {
				t.Errorf("len(Reflections) = %d, want %d", len(b.Reflections), tt.wantLen)
			}
			if len(b.Episodic) != tt.wantLen {
				t.Errorf("len(Episodic) = %d, want %d", len(b.Episodic), tt.wantLen)
			}
		})
	}
}

func TestStubClient_Retrieve_FixedClock(t *testing.T) {
	fixed := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	c := &StubClient{Now: func() time.Time { return fixed }}
	b, err := c.Retrieve(context.Background(), Retrieval{Goal: "anything"})
	if err != nil {
		t.Fatalf("Retrieve: %v", err)
	}
	if !b.Timestamp.Equal(fixed) {
		t.Errorf("Timestamp = %v, want %v", b.Timestamp, fixed)
	}
}

func TestStubClient_Retrieve_ContextCancellation(t *testing.T) {
	c := NewStubClient()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.Retrieve(ctx, Retrieval{Goal: "combat"})
	if err == nil {
		t.Errorf("expected context cancellation error, got nil")
	}
}

func TestStubClient_Retrieve_Deterministic(t *testing.T) {
	c := NewStubClient()
	r := Retrieval{Goal: "should I attack", HostName: "test"}
	b1, _ := c.Retrieve(context.Background(), r)
	b2, _ := c.Retrieve(context.Background(), r)
	if !stringSliceEq(b1.Reflections, b2.Reflections) {
		t.Errorf("Reflections not deterministic: %v vs %v", b1.Reflections, b2.Reflections)
	}
	if !stringSliceEq(b1.Episodic, b2.Episodic) {
		t.Errorf("Episodic not deterministic: %v vs %v", b1.Episodic, b2.Episodic)
	}
}

func TestStubClient_ConcurrentRetrieve(t *testing.T) {
	c := NewStubClient()
	done := make(chan error, 10)
	for i := 0; i < 10; i++ {
		go func() {
			_, err := c.Retrieve(context.Background(), Retrieval{Goal: "combat"})
			done <- err
		}()
	}
	for i := 0; i < 10; i++ {
		if err := <-done; err != nil {
			t.Errorf("concurrent Retrieve: %v", err)
		}
	}
}

// helpers

func anyContains(xs []string, needle string) bool {
	for _, s := range xs {
		if strings.Contains(strings.ToLower(s), strings.ToLower(needle)) {
			return true
		}
	}
	return false
}

// Compile-time check that StubClient satisfies Client.
var _ Client = (*StubClient)(nil)

func stringSliceEq(a, b []string) bool {
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
