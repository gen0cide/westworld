package mesad

import (
	"context"
	"fmt"
	"strings"
	"testing"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
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

// flatCompleter stubs the textCompleter seam (no real API calls), recording the
// assembled system + user prompts so tests can assert prompt assembly.
type flatCompleter struct {
	system, user string
	resp         string
}

func (f *flatCompleter) Complete(_ context.Context, system, user string, _ int) (string, error) {
	f.system, f.user = system, user
	return f.resp, nil
}

// TestAnalysisInterpretReflectPath proves the C-4 REFLECT kind end-to-end
// through the interpreter body: an abstract operator question classified
// "reflect" comes back as persona-voiced prose (no DSL authored), and the
// assembled prompt carries BOTH the reflect contract + the host's persona prose
// (system) AND the STATE facts + DIRECTIVE (user) the voice is grounded in.
func TestAnalysisInterpretReflectPath(t *testing.T) {
	s := quietServer()
	// Seed persona prose directly into the registry (bypassing full persona
	// validation — the interpreter only reads e.prose), like neighbors do.
	s.mu.Lock()
	s.reg["dolores"] = &entry{prose: "You are Dolores, a thoughtful rancher's daughter who second-guesses her own choices."}
	s.mu.Unlock()

	fake := &flatCompleter{resp: `{"kind":"reflect","text":"  I stopped because my pick snapped, and I didn't trust myself to buy another without asking first.  "}`}
	v, err := s.analysisInterpret(context.Background(), "dolores", fake,
		&mesapb.AnalysisDirective{Directive: "why did you stop mining?", State: []string{"pos (220,448)", "HP 25/30", "carrying: broken pickaxe"}})
	if err != nil {
		t.Fatalf("analysisInterpret: %v", err)
	}
	if v.GetKind() != "reflect" {
		t.Fatalf("kind = %q, want reflect", v.GetKind())
	}
	if v.GetText() != "I stopped because my pick snapped, and I didn't trust myself to buy another without asking first." {
		t.Fatalf("reflect text not trimmed/round-tripped: %q", v.GetText())
	}
	if v.GetDsl() != "" {
		t.Fatalf("reflect must not author DSL, got %q", v.GetDsl())
	}
	// Prompt assembly: the system block must carry the reflect contract AND the
	// persona prose that grounds the in-character voice.
	if !strings.Contains(fake.system, `"kind":"reflect"`) {
		t.Fatalf("system prompt missing the reflect contract:\n%s", fake.system)
	}
	if !strings.Contains(fake.system, "Dolores") || !strings.Contains(fake.system, "# THE CHARACTER") {
		t.Fatalf("system prompt missing the persona block:\n%s", fake.system)
	}
	// The user turn must ground the model in the STATE facts + the directive.
	if !strings.Contains(fake.user, "DIRECTIVE: why did you stop mining?") {
		t.Fatalf("user prompt missing the directive:\n%s", fake.user)
	}
	if !strings.Contains(fake.user, "- pos (220,448)") || !strings.Contains(fake.user, "- carrying: broken pickaxe") {
		t.Fatalf("user prompt missing STATE facts:\n%s", fake.user)
	}
}

// TestAnalysisInterpretSystemPersonaBlock proves the per-host system assembly:
// prose appends the character block, while a persona-less host gets the BARE
// contract — never an empty trailing section (the M20 discipline).
func TestAnalysisInterpretSystemPersonaBlock(t *testing.T) {
	if got := analysisInterpretSystem(""); got != analysisSystem {
		t.Fatalf("empty prose must yield the bare contract, got extra:\n%s", strings.TrimPrefix(got, analysisSystem))
	}
	if got := analysisInterpretSystem("   "); got != analysisSystem {
		t.Fatal("whitespace prose must yield the bare contract")
	}
	got := analysisInterpretSystem("You are Gregor, a blunt blacksmith.")
	if !strings.HasPrefix(got, analysisSystem) || !strings.Contains(got, "Gregor") {
		t.Fatalf("prose not appended after the contract:\n%s", got)
	}
}

// TestAnalysisInterpretRouting proves the verdict routing around the new kind:
// directives still come back as DSL commands, factual answers stay flat text,
// kind casing folds, and an unclassifiable kind defers to hypothetical (the
// host then runs the real planner dry-run) — reflect must NOT widen that path.
func TestAnalysisInterpretRouting(t *testing.T) {
	tests := []struct {
		name     string
		resp     string
		wantKind string
		wantDSL  string
		wantText string
	}{
		{"directive still returns DSL", `{"kind":"command","dsl":"go_to(\"Varrock\")"}`, "command", `go_to("Varrock")`, ""},
		{"factual answer stays flat", `{"kind":"answer","text":"pos (220,448), HP 25/30"}`, "answer", "", "pos (220,448), HP 25/30"},
		{"reflect kind folds casing", `{"kind":"REFLECT","text":"I keep my own counsel."}`, "reflect", "", "I keep my own counsel."},
		{"unknown kind defers to hypothetical", `{"kind":"poem","text":"roses"}`, "hypothetical", "", "roses"},
		{"non-json defers to hypothetical", `no json here`, "hypothetical", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := quietServer()
			fake := &flatCompleter{resp: tt.resp}
			v, err := s.analysisInterpret(context.Background(), "no-persona-host", fake,
				&mesapb.AnalysisDirective{Directive: "whatever"})
			if err != nil {
				t.Fatalf("analysisInterpret: %v", err)
			}
			if v.GetKind() != tt.wantKind || v.GetDsl() != tt.wantDSL || v.GetText() != tt.wantText {
				t.Fatalf("verdict = (%q,%q,%q), want (%q,%q,%q)",
					v.GetKind(), v.GetDsl(), v.GetText(), tt.wantKind, tt.wantDSL, tt.wantText)
			}
			// No registered persona → the bare flat contract, no empty character block
			// (the contract TEXT mentions "THE CHARACTER below"; the appended SECTION
			// header is "# THE CHARACTER" — assert on the header).
			if strings.Contains(fake.system, "# THE CHARACTER") {
				t.Fatalf("persona-less host got a character block:\n%s", fake.system)
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
