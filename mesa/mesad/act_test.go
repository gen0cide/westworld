package mesad

import (
	"fmt"
	"strings"
	"testing"
)

// TestParseMoveGoalOpFolding proves the mesa side folds goal_op synonyms onto the
// recognised vocabulary {"", "done", "abandoned", "adopt"} so it agrees with
// runtime's normalizeGoalOp (M7). A declared-but-misspelled completion
// ("complete"/"finished") must fold to "done" instead of being silently routed into
// the still-active progress branch; an unrecognised value is reset to "".
func TestParseMoveGoalOpFolding(t *testing.T) {
	tests := []struct {
		name  string
		rawOp string
		want  string
		known bool // normalizeGoalOp recognized
	}{
		{"empty stays empty", "", "", true},
		{"exact done", "done", "done", true},
		{"synonym complete folds to done", "complete", "done", true},
		{"synonym completed folds to done", "completed", "done", true},
		{"synonym finished folds to done", "finished", "done", true},
		{"synonym satisfied folds to done", "satisfied", "done", true},
		{"case-insensitive Done", "Done", "done", true},
		{"whitespace trimmed", "  finished  ", "done", true},
		{"abandon synonym folds", "give up", "abandoned", true},
		{"adopt synonym folds", "new goal", "adopt", true},
		{"unrecognised resets to empty", "yolo", "", false},
		{"garbage resets to empty", "kind-of-done-ish", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// normalizeGoalOp contract.
			op, ok := normalizeGoalOp(tt.rawOp)
			if op != tt.want || ok != tt.known {
				t.Fatalf("normalizeGoalOp(%q) = (%q,%v), want (%q,%v)", tt.rawOp, op, ok, tt.want, tt.known)
			}
			// End-to-end through parseMove: the stamped GoalOp must be the folded value.
			raw := fmt.Sprintf(`{"kind":"idle","goal_op":%q,"goal_text":"next thing","goal_progress":0.5}`, tt.rawOp)
			m := parseMove(raw)
			if m.GetGoalOp() != tt.want {
				t.Fatalf("parseMove goal_op = %q, want %q (raw=%q)", m.GetGoalOp(), tt.want, tt.rawOp)
			}
			// goal_text/goal_progress must round-trip regardless of the goal_op outcome.
			if m.GetGoalText() != "next thing" {
				t.Fatalf("goal_text dropped: %q", m.GetGoalText())
			}
			if m.GetGoalProgress() != 0.5 {
				t.Fatalf("goal_progress dropped: %v", m.GetGoalProgress())
			}
		})
	}
}

// TestParseMoveGoalServes proves the planner's aspiration declaration
// (goal_serves, set alongside goal_op:"adopt") rides the Move trimmed, and is
// simply empty when absent — the host's nearest-match fallback then applies.
func TestParseMoveGoalServes(t *testing.T) {
	m := parseMove(`{"kind":"idle","goal_op":"adopt","goal_text":"buy a pickaxe","goal_serves":"  master a craft  "}`)
	if m.GetGoalServes() != "master a craft" {
		t.Fatalf("goal_serves = %q, want trimmed declaration", m.GetGoalServes())
	}
	if m := parseMove(`{"kind":"idle","goal_op":"adopt","goal_text":"buy a pickaxe"}`); m.GetGoalServes() != "" {
		t.Fatalf("absent goal_serves should be empty, got %q", m.GetGoalServes())
	}
}

// TestRawGoalOpExtraction proves rawGoalOp recovers the model's original goal_op so
// act() can WARN on an unrecognised value (parseMove has already reset it).
func TestRawGoalOpExtraction(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"recognised", `{"kind":"idle","goal_op":"done"}`, "done"},
		{"unrecognised preserved", `{"kind":"idle","goal_op":"yolo"}`, "yolo"},
		{"trimmed", `{"goal_op":"  complete  "}`, "complete"},
		{"absent field", `{"kind":"idle"}`, ""},
		{"not json", `not json at all`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := rawGoalOp(tt.raw); got != tt.want {
				t.Fatalf("rawGoalOp(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

// TestCapChatRuneCount proves capChat caps on the RUNE count (the RSC wire
// counts characters, not bytes) and folds the multibyte chars host-side
// sanitizeChat folds — so a legal <=80-rune reply with a few em-dashes/curly
// quotes is NOT needlessly truncated (regression for L11).
func TestCapChatRuneCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "short ascii unchanged",
			in:   "hello there",
			want: "hello there",
		},
		{
			name: "exactly 80 ascii runes survives",
			in:   strings.Repeat("a", 80),
			want: strings.Repeat("a", 80),
		},
		{
			name: "81 ascii runes cut to 80",
			in:   strings.Repeat("a", 81),
			want: strings.Repeat("a", 80),
		},
		{
			// 75 ASCII chars + 6 em-dashes = 75 runes after folding (each — → -),
			// well under 80. A byte cap (87 bytes) would have chopped the tail; the
			// rune cap must keep the whole line.
			name: "multibyte reply within 80 runes survives",
			in:   strings.Repeat("a", 69) + "—x—y—z",
			want: strings.Repeat("a", 69) + "-x-y-z",
		},
		{
			// Folds curly quotes/ellipsis to ASCII so the count matches what the host
			// actually sends.
			name: "folds curly quotes and ellipsis",
			in:   "“hi” it’s me…",
			want: "\"hi\" it's me...",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capChat(tt.in)
			if got != tt.want {
				t.Fatalf("capChat(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if n := len([]rune(got)); n > 80 {
				t.Fatalf("capChat result has %d runes, exceeds 80: %q", n, got)
			}
		})
	}
}

// TestHasGameActionASTNotSubstring proves hasGameAction walks the parsed AST for
// real PrimaryAction calls — a verb mentioned only inside a string literal or a
// comment is NOT a false positive, and a real call IS detected (regression for
// L12). validateMove must reject the narration-only routine and accept the
// genuine one. A nil catalog keeps the test focused on the no-op guard.
func TestHasGameActionASTNotSubstring(t *testing.T) {
	var nilCat *argCatalog
	tests := []struct {
		name   string
		src    string
		reject bool // true → validateMove must reject as no-op
	}{
		{
			name:   "real game action accepted",
			src:    "runtime \"1.0\"\nroutine r() { attack(nearest_npc()) }",
			reject: false,
		},
		{
			name:   "verb inside a string literal is not a game action",
			src:    "runtime \"1.0\"\nroutine r() { note(\"then I attack( the goblin\") }",
			reject: true,
		},
		{
			name:   "verb inside a comment is not a game action",
			src:    "runtime \"1.0\"\nroutine r() {\n  # plan: go_to(\"bank\") later\n  note(\"thinking\")\n}",
			reject: true,
		},
		{
			name:   "notes and waits only are rejected",
			src:    "runtime \"1.0\"\nroutine r() { note(\"hmm\"); wait(2) }",
			reject: true,
		},
		{
			name:   "game action nested in an if is detected",
			src:    "runtime \"1.0\"\nroutine r() { if true { go_to(\"bank\") } }",
			reject: false,
		},
		{
			name:   "bang-form game action is detected",
			src:    "runtime \"1.0\"\nroutine r() { try { attack!(nearest_npc()) } recover e { note(\"ow\") } }",
			reject: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMove(routineMove(tt.src), nilCat)
			if tt.reject && err == nil {
				t.Fatalf("expected no-op rejection for %q", tt.src)
			}
			if !tt.reject && err != nil {
				t.Fatalf("expected acceptance for %q, got %v", tt.src, err)
			}
			// The no-op rejection must be the "takes no real game action" one, not a
			// parse error — guards against the test passing for the wrong reason.
			if tt.reject && err != nil && !strings.Contains(err.Error(), "no real game action") {
				t.Fatalf("rejection should be the no-op guard; got %v", err)
			}
		})
	}
}
