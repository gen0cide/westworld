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

// TestParseGenesisAspirations proves the aspiration-ladder schema extension is
// VALIDATED, never trusted: blanks/dupes dropped, the list capped, opening_serves
// kept only when it names a kept aspiration, and — the graceful-fallback contract
// — NO aspirations at all degrades to exactly the pre-ladder result shape.
func TestParseGenesisAspirations(t *testing.T) {
	raw := `{"goal":"Smelt my first bronze bar","aspirations":` +
		`["master a craft (smithing)","  ","secure a steady livelihood","MASTER A CRAFT (SMITHING)","know the whole map","be famous","one too many"],` +
		`"opening_serves":"master a craft (smithing)","reasoning":"r"}`
	res := parseGenesis(raw)
	if got := res.GetAspirations(); len(got) != 4 {
		t.Fatalf("aspirations = %v, want 4 (blank + case-dupe dropped, capped at 4)", got)
	} else if got[0] != "master a craft (smithing)" || got[1] != "secure a steady livelihood" {
		t.Fatalf("aspiration order/content wrong: %v", got)
	}
	if res.GetOpeningServes() != "master a craft (smithing)" {
		t.Fatalf("opening_serves = %q", res.GetOpeningServes())
	}

	// opening_serves naming a NON-aspiration is discarded (the host's mechanical
	// nearest-match takes over) — never passed through unvalidated.
	res = parseGenesis(`{"goal":"g","aspirations":["a thing"],"opening_serves":"something else"}`)
	if res.GetOpeningServes() != "" {
		t.Fatalf("invalid opening_serves kept: %q", res.GetOpeningServes())
	}

	// Fallback: no aspirations emitted = today's behavior (goal still parsed,
	// nothing invented).
	res = parseGenesis(`{"goal":"Train mining","reasoning":"r"}`)
	if len(res.GetAspirations()) != 0 || res.GetOpeningServes() != "" || res.GetGoal() != "Train mining" {
		t.Fatalf("aspirationless genesis should degrade cleanly: %+v", res)
	}
}
