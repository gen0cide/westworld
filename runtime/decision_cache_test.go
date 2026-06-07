package runtime

import (
	"context"
	"testing"

	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/dsl/interp"
)

type countingStrategist struct {
	n      *int
	choice string
}

func (c *countingStrategist) Decide(_ context.Context, _ brain.Situation) (*brain.Decision, error) {
	*c.n++
	return &brain.Decision{Choice: c.choice, Confidence: 1}, nil
}

func decideArgs(question string, options ...string) []interp.Value {
	items := make([]interp.Value, len(options))
	for i, o := range options {
		items[i] = interp.String(o)
	}
	return []interp.Value{&interp.List{Items: items}, interp.String(question)}
}

func choiceOf(v interp.Value) string {
	if cr, ok := v.(*interp.CallResult); ok && cr.Val != nil {
		return cr.Val.Display()
	}
	return ""
}

// TestDecisionCacheSkipsStrategistOnRepeat proves the decision half of the cheap
// loop: an identical pearl-miss decide() in the same coarse state hits the cache
// and never re-calls the (Haiku) Strategist.
func TestDecisionCacheSkipsStrategistOnRepeat(t *testing.T) {
	h := newTestHost()
	h.Pearl = nil // force the strategist path (pearl always "misses")
	calls := 0
	h.Strategist = &countingStrategist{n: &calls, choice: "fight"}
	ctx := context.Background()
	args := decideArgs("what do I do?", "fight", "flee")

	r1, err1 := dslDecide(ctx, h, args, nil)
	r2, err2 := dslDecide(ctx, h, args, nil)
	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v / %v", err1, err2)
	}
	if calls != 1 {
		t.Fatalf("strategist called %d times; the repeat must hit the cache (want 1)", calls)
	}
	if choiceOf(r1) != "fight" || choiceOf(r2) != "fight" {
		t.Fatalf("choices: %q then %q, want fight", choiceOf(r1), choiceOf(r2))
	}

	// Option ORDER must not split the key (sorted) — still a cache hit.
	if _, err := dslDecide(ctx, h, decideArgs("what do I do?", "flee", "fight"), nil); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("reordered options re-called strategist (calls=%d); the key should be order-independent", calls)
	}
}

// TestDecisionCacheDistinguishesQuestions proves the key discriminates: a
// different question is a fresh decision, not a stale cache hit.
func TestDecisionCacheDistinguishesQuestions(t *testing.T) {
	h := newTestHost()
	h.Pearl = nil
	calls := 0
	h.Strategist = &countingStrategist{n: &calls, choice: "x"}
	ctx := context.Background()

	_, _ = dslDecide(ctx, h, decideArgs("q1", "a", "b"), nil) // miss → call
	_, _ = dslDecide(ctx, h, decideArgs("q1", "a", "b"), nil) // hit
	_, _ = dslDecide(ctx, h, decideArgs("q2", "a", "b"), nil) // different question → call

	if calls != 2 {
		t.Fatalf("strategist called %d times, want 2 (q1 once + q2 once)", calls)
	}
}
