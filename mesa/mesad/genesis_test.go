package mesad

import "testing"

func TestParseGenesis(t *testing.T) {
	raw := `{"goal":"Train mining to level 10","mood":{"stress":0.2,"confidence":0.7,"valence":0.3},` +
		`"keyword_ladder":[{"keyword":"Delores","tier":"DIRECTED_SOCIAL","action":"orient and consider replying"},` +
		`{"keyword":"trade","tier":"TRADE_INTEREST","action":"consider it by relationship"}],"reasoning":"early game"}`

	res := parseGenesis(raw)
	if res.GetGoal() != "Train mining to level 10" {
		t.Fatalf("goal = %q", res.GetGoal())
	}
	if m := res.GetMood(); m.GetConfidence() != 0.7 || m.GetStress() != 0.2 || m.GetValence() != 0.3 {
		t.Fatalf("mood = %+v", m)
	}
	if len(res.GetKeywordLadder()) != 2 || res.GetKeywordLadder()[0].GetKeyword() != "Delores" {
		t.Fatalf("ladder = %+v", res.GetKeywordLadder())
	}

	// Tolerates surrounding prose (extractJSON).
	if got := parseGenesis("Here's my session plan:\n" + raw + "\nDone."); got.GetGoal() != "Train mining to level 10" {
		t.Fatalf("prose-wrapped goal = %q", got.GetGoal())
	}

	// A parse failure degrades to a minimal result so the host still boots.
	if bad := parseGenesis("not json at all"); bad.GetGoal() != "" || bad.GetReasoning() == "" {
		t.Fatalf("expected empty fallback with reasoning, got %+v", bad)
	}

	// Blank-keyword rungs are dropped.
	if got := parseGenesis(`{"goal":"g","keyword_ladder":[{"keyword":"","tier":"AMBIENT"}]}`); len(got.GetKeywordLadder()) != 0 {
		t.Fatalf("blank rung not dropped: %+v", got.GetKeywordLadder())
	}
}
