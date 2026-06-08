package mesad

import (
	"testing"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// TestParseExtractedDialogRoundTrip proves a well-formed model response maps to
// the protobuf set: claims preserved, empty-content claims dropped, intent kept.
func TestParseExtractedDialogRoundTrip(t *testing.T) {
	raw := `here you go: {"claims":[
		{"subject":"Nurmof","kind":"npc","claim":"sells pickaxes","confidence":0.8,"provenance":"player-told"},
		{"subject":"","kind":"","claim":"","confidence":0.1,"provenance":"implied"}
	],"intent":{"kind":"offer","urgency":"High","gist":"wants to trade"}}`
	set := parseExtractedDialog(raw)
	if len(set.GetClaims()) != 1 {
		t.Fatalf("empty-content claim not dropped: got %d claims", len(set.GetClaims()))
	}
	c := set.GetClaims()[0]
	if c.GetSubject() != "Nurmof" || c.GetClaim() != "sells pickaxes" || c.GetConfidence() != 0.8 {
		t.Fatalf("claim not preserved: %+v", c)
	}
	if set.GetIntent().GetKind() != "offer" || set.GetIntent().GetUrgency() != "high" {
		t.Fatalf("intent not parsed/normalized: %+v", set.GetIntent())
	}
}

// TestParseExtractedDialogMalformedFallback proves a non-JSON / broken response
// degrades to a safe empty set + a low-urgency statement (never nil).
func TestParseExtractedDialogMalformedFallback(t *testing.T) {
	for _, raw := range []string{"", "not json at all", "{not: valid}", "{\"claims\": [garbage]}"} {
		set := parseExtractedDialog(raw)
		if set == nil || set.GetIntent() == nil {
			t.Fatalf("malformed %q yielded a nil set/intent", raw)
		}
		if len(set.GetClaims()) != 0 {
			t.Fatalf("malformed %q yielded claims: %+v", raw, set.GetClaims())
		}
		if set.GetIntent().GetUrgency() != "low" {
			t.Fatalf("malformed %q should fall back to low urgency, got %q", raw, set.GetIntent().GetUrgency())
		}
	}
}

// TestExtractWantsNuance proves the Sonnet-escalation predicate: only a long
// multi-turn window (>=5 lines) escalates; a short exchange stays on Haiku even
// when goal-relevant (the goal is given to Haiku as context — cost stance).
func TestExtractWantsNuance(t *testing.T) {
	short := &mesapb.DialogWindow{Window: []string{"a", "b"}}
	if extractWantsNuance(short) {
		t.Fatal("a short goalless window should NOT escalate (Haiku)")
	}
	long := &mesapb.DialogWindow{Window: []string{"a", "b", "c", "d", "e"}}
	if !extractWantsNuance(long) {
		t.Fatal("a 5+-line window should escalate (Sonnet)")
	}
	// A merely goal-relevant SHORT exchange stays on Haiku.
	goalShort := &mesapb.DialogWindow{Window: []string{"a"}, ActiveGoal: "find a pickaxe"}
	if extractWantsNuance(goalShort) {
		t.Fatal("a short goal-relevant window should stay on Haiku, not escalate")
	}
	goalLong := &mesapb.DialogWindow{Window: []string{"a", "b", "c", "d", "e"}, ActiveGoal: "find a pickaxe"}
	if !extractWantsNuance(goalLong) {
		t.Fatal("a 5+-line window should escalate regardless of goal")
	}
}
