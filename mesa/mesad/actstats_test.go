package mesad

import (
	"strings"
	"testing"
)

// TestClassifyRejection pins the error-text → class mapping the counters and
// the sharpened re-prompt both ride. Parse failures split on the f-string
// signal (present in the parser's own f-string diagnostics AND in the lexer
// message rejectionDetail appends); everything else is non_action or the
// validation bucket.
func TestClassifyRejection(t *testing.T) {
	tests := []struct {
		name string
		text string
		want string
	}{
		{
			name: "lexer-enriched unterminated f-string",
			text: `dsl_source does not parse: parse <act-move>: <act-move>:3:19: expected expression, got ILLEGAL — lexer: unterminated f-string (newline before closing quote)`,
			want: rejectFStringParse,
		},
		{
			name: "lexer-enriched unknown f-string escape",
			text: `dsl_source does not parse: parse <act-move>: <act-move>:2:7: expected expression, got ILLEGAL — lexer: unknown f-string escape \'`,
			want: rejectFStringParse,
		},
		{
			name: "parser's own f-string placeholder diagnostic",
			text: `dsl_source does not parse: parse <act-move>: <act-move>:2:19: f-string placeholder holds ONE expression — got IDENT; write "{a} {b}", not "{a b}"`,
			want: rejectFStringParse,
		},
		{
			name: "lexer-enriched unterminated plain string",
			text: `dsl_source does not parse: parse <act-move>: <act-move>:3:9: expected expression, got ILLEGAL — lexer: unterminated string literal (newline before closing quote)`,
			want: rejectOtherParse,
		},
		{
			name: "ordinary syntax error",
			text: `dsl_source does not parse: parse <act-move>: <act-move>:4:3: expected RPAREN, got IDENT`,
			want: rejectOtherParse,
		},
		{
			name: "no-op routine",
			text: "the routine takes no real game action (only notes/waits/reads) — author one that actually acts toward the goal",
			want: rejectNonAction,
		},
		{
			name: "empty dsl_source",
			text: "write_routine has an empty dsl_source",
			want: rejectValidation,
		},
		{
			name: "unknown verb",
			text: `unknown verb "teleport" — use one of the documented actions`,
			want: rejectValidation,
		},
		{
			name: "hallucinated arg literal",
			text: `go_to: "mining-site" is not a known place — nearest real names: ...`,
			want: rejectValidation,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := classifyRejection(tt.text); got != tt.want {
				t.Fatalf("classifyRejection(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// TestActRejectStatsBump proves each class lands in its own counter and an
// unknown class falls into the validation bucket rather than vanishing.
// Exercises a fresh struct, not the package-level actStats (which other tests
// in this package may bump through the act path).
func TestActRejectStatsBump(t *testing.T) {
	var st actRejectStats
	st.bump(rejectFStringParse)
	st.bump(rejectFStringParse)
	st.bump(rejectOtherParse)
	st.bump(rejectNonAction)
	st.bump(rejectValidation)
	st.bump("never-heard-of-it")
	if got := st.fstringParse.Load(); got != 2 {
		t.Fatalf("fstringParse = %d, want 2", got)
	}
	if got := st.otherParse.Load(); got != 1 {
		t.Fatalf("otherParse = %d, want 1", got)
	}
	if got := st.nonAction.Load(); got != 1 {
		t.Fatalf("nonAction = %d, want 1", got)
	}
	if got := st.validation.Load(); got != 2 {
		t.Fatalf("validation = %d, want 2 (own class + unknown fallback)", got)
	}
	if got := st.autoRepaired.Load(); got != 0 {
		t.Fatalf("autoRepaired = %d, want 0 (bump never touches it)", got)
	}
}

// TestRejectionDetailLexEnrichment proves a write_routine parse rejection
// carries the lexer's own message (the parse error alone only says "got
// ILLEGAL"), and that non-parse rejections pass through untouched.
func TestRejectionDetailLexEnrichment(t *testing.T) {
	var nilCat *argCatalog

	m := routineMove(srcUntermFStringNewline)
	_, verr := validateMove(m, nilCat)
	if verr == nil {
		t.Fatal("source unexpectedly validates")
	}
	detail := rejectionDetail(verr, m)
	if !strings.Contains(detail, "lexer: unterminated f-string (newline before closing quote)") {
		t.Fatalf("detail missing the lexer message: %q", detail)
	}
	if classifyRejection(detail) != rejectFStringParse {
		t.Fatalf("enriched detail must classify fstring_parse, got %q", classifyRejection(detail))
	}

	m = routineMove(`runtime "1.0"
routine r() { note("hmm") }`)
	_, verr = validateMove(m, nilCat)
	if verr == nil {
		t.Fatal("no-op source unexpectedly validates")
	}
	if got := rejectionDetail(verr, m); got != verr.Error() {
		t.Fatalf("non-parse rejection must pass through untouched: %q", got)
	}
}

// TestSharpenRejection proves the f-string class — and only that class —
// gets the manual's recipe appended, with all three ingredients the retry
// needs (bind a local / ONE expression per {} / prefer format()).
func TestSharpenRejection(t *testing.T) {
	base := "dsl_source does not parse: ... — lexer: unterminated f-string (reached EOF)"
	got := sharpenRejection(rejectFStringParse, base)
	if !strings.HasPrefix(got, base) {
		t.Fatalf("sharpened rejection must keep the original text first: %q", got)
	}
	for _, ingredient := range []string{"bind a local", "ONE expression per {}", `format("`} {
		if !strings.Contains(got, ingredient) {
			t.Fatalf("recipe missing %q: %q", ingredient, got)
		}
	}
	for _, class := range []string{rejectOtherParse, rejectValidation, rejectNonAction, ""} {
		if got := sharpenRejection(class, base); got != base {
			t.Fatalf("class %q must pass through unchanged, got %q", class, got)
		}
	}
}
