package mesad

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
)

// Incident sources — each is the act-move shape that produced one of the
// exact lexer errors from the 2026-06-11 OOM incident class (an LLM-authored
// routine with a string typo). Pinned by TestIncidentLexerErrorShapes so a
// lexer wording change breaks THESE tests, not the repair silently.
// Placeholders reference self.* (always bound) because the repaired source
// must clear the FULL validateMove — which now includes the host's static
// validator, where a bare {hp} would die as an unbound identifier.
const (
	srcUntermFStringNewline = "runtime \"1.0\"\nroutine r() {\n  msg = f\"hp is {self.hp}\n  say(msg)\n}\n"
	srcUntermFStringEOF     = "runtime \"1.0\"\nroutine r() {\n  say(f\"hello"
	srcUntermStringNewline  = "runtime \"1.0\"\nroutine r() {\n  msg = \"hello there\n  say(msg)\n}\n"
	srcUntermStringEOF      = "runtime \"1.0\"\nroutine r() {\n  say(\"hello"
	srcFStringEscapedQuote  = "runtime \"1.0\"\nroutine r() {\n  say(f\"it\\'s {self.name} time\")\n}\n"
)

// TestIncidentLexerErrorShapes proves the crafted sources produce the EXACT
// lexer errors the repair dispatches on (and that the parser does reject
// them) — if the lexer rewords, the contains-checks in repairLexFailure and
// these fixtures must move together.
func TestIncidentLexerErrorShapes(t *testing.T) {
	tests := []struct {
		name    string
		src     string
		wantMsg string
	}{
		{"fstring newline", srcUntermFStringNewline, "unterminated f-string (newline before closing quote)"},
		{"fstring eof", srcUntermFStringEOF, "unterminated f-string (reached EOF)"},
		{"string newline", srcUntermStringNewline, "unterminated string literal (newline before closing quote)"},
		{"string eof", srcUntermStringEOF, "unterminated string literal (reached EOF)"},
		{"escaped quote in fstring", srcFStringEscapedQuote, `unknown f-string escape \'`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := parser.Parse("<act-move>", tt.src); err == nil {
				t.Fatalf("source unexpectedly parses:\n%s", tt.src)
			}
			msg, _, ok := firstLexFailure(tt.src)
			if !ok {
				t.Fatalf("no lexer failure found in:\n%s", tt.src)
			}
			if msg != tt.wantMsg {
				t.Fatalf("lexer message = %q, want %q", msg, tt.wantMsg)
			}
		})
	}
}

// TestTryAutoRepairFixesIncidentShapes drives the full pre-repair seam the
// act loop uses: a rejected write_routine comes back with its DslSource
// replaced by the mechanically repaired source, and the repaired move passes
// the FULL validateMove cleanly.
func TestTryAutoRepairFixesIncidentShapes(t *testing.T) {
	var nilCat *argCatalog
	tests := []struct {
		name      string
		src       string
		wantKind  string
		wantFixed string
	}{
		{
			name:      "missing close quote on an f-string assignment",
			src:       srcUntermFStringNewline,
			wantKind:  "close_quote",
			wantFixed: "runtime \"1.0\"\nroutine r() {\n  msg = f\"hp is {self.hp}\"\n  say(msg)\n}\n",
		},
		{
			name:      "missing close quote on a plain string assignment",
			src:       srcUntermStringNewline,
			wantKind:  "close_quote",
			wantFixed: "runtime \"1.0\"\nroutine r() {\n  msg = \"hello there\"\n  say(msg)\n}\n",
		},
		{
			name:      "backslash-escaped quote inside an f-string",
			src:       srcFStringEscapedQuote,
			wantKind:  "fstring_escape",
			wantFixed: "runtime \"1.0\"\nroutine r() {\n  say(f\"it's {self.name} time\")\n}\n",
		},
		{
			name:      "escaped quote AND missing close quote on one line",
			src:       "runtime \"1.0\"\nroutine r() {\n  msg = f\"it\\'s {self.hp}\n  say(msg)\n}\n",
			wantKind:  "fstring_escape+close_quote",
			wantFixed: "runtime \"1.0\"\nroutine r() {\n  msg = f\"it's {self.hp}\"\n  say(msg)\n}\n",
		},
		{
			name:      "two unterminated f-strings",
			src:       "runtime \"1.0\"\nroutine r() {\n  a = f\"one {self.hp}\n  b = f\"two {self.name}\n  say(a)\n}\n",
			wantKind:  "close_quote+close_quote",
			wantFixed: "runtime \"1.0\"\nroutine r() {\n  a = f\"one {self.hp}\"\n  b = f\"two {self.name}\"\n  say(a)\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := routineMove(tt.src)
			_, verr := validateMove(m, nilCat)
			if verr == nil {
				t.Fatalf("source unexpectedly validates pre-repair:\n%s", tt.src)
			}
			kind, ok := tryAutoRepair(m, nilCat, verr)
			if !ok {
				t.Fatalf("tryAutoRepair did not repair:\n%s", tt.src)
			}
			if kind != tt.wantKind {
				t.Fatalf("repair kind = %q, want %q", kind, tt.wantKind)
			}
			if got := m.GetDslSource(); got != tt.wantFixed {
				t.Fatalf("repaired source:\n%s\nwant:\n%s", got, tt.wantFixed)
			}
			if _, err := validateMove(m, nilCat); err != nil {
				t.Fatalf("repaired move must validate cleanly, got %v", err)
			}
		})
	}
}

// TestTryAutoRepairFallsThrough proves the all-or-nothing acceptance: a
// repair that doesn't yield a COMPLETELY clean parse+validate leaves the move
// untouched (the re-prompt sees the original source), and non-parse
// rejections never enter the repair path at all.
func TestTryAutoRepairFallsThrough(t *testing.T) {
	var nilCat *argCatalog
	tests := []struct {
		name string
		src  string
	}{
		{
			// Truncated output: appending the close quote still leaves the call and
			// the routine block unclosed — repaired-but-dirty must be rejected.
			name: "eof truncation is not fully repairable",
			src:  srcUntermFStringEOF,
		},
		{
			// The missing quote swallowed the structural `)` into the f-string;
			// closing at the newline cannot give a clean parse.
			name: "close paren swallowed by the unterminated f-string",
			src:  "runtime \"1.0\"\nroutine r() {\n  say(f\"hello {name})\n  wait(2)\n}\n",
		},
		{
			// Parses fine, rejected as a no-op — not a parse failure, no repair try.
			name: "non-action rejection skips repair",
			src:  "runtime \"1.0\"\nroutine r() { note(\"hmm\"); wait(2) }",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := routineMove(tt.src)
			_, verr := validateMove(m, nilCat)
			if verr == nil {
				t.Fatalf("source unexpectedly validates:\n%s", tt.src)
			}
			if kind, ok := tryAutoRepair(m, nilCat, verr); ok {
				t.Fatalf("tryAutoRepair accepted a dirty repair (kind %q)", kind)
			}
			if m.GetDslSource() != tt.src {
				t.Fatalf("move source mutated on a failed repair:\n%s", m.GetDslSource())
			}
		})
	}
}

// TestRepairDSLPureShapes pins repairDSL's per-shape mechanics, including the
// repairs that can't reach a clean parse (the act seam then discards them)
// and the failures that must not be touched at all.
func TestRepairDSLPureShapes(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		wantSrc   string
		wantKinds []string
	}{
		{
			name:      "eof shape appends the close quote",
			src:       `say(f"hello`,
			wantSrc:   `say(f"hello"`,
			wantKinds: []string{repairCloseQuote},
		},
		{
			name:      "unrepairable lexer failure untouched",
			src:       "x = 1 & 2",
			wantSrc:   "x = 1 & 2",
			wantKinds: nil,
		},
		{
			name:      "non-quote unknown escape untouched",
			src:       "say(f\"cost \\${x}\")",
			wantSrc:   "say(f\"cost \\${x}\")",
			wantKinds: nil,
		},
		{
			name:      "clean source untouched",
			src:       `say(f"hp {hp}")`,
			wantSrc:   `say(f"hp {hp}")`,
			wantKinds: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, kinds := repairDSL(tt.src)
			if got != tt.wantSrc {
				t.Fatalf("repairDSL source = %q, want %q", got, tt.wantSrc)
			}
			if strings.Join(kinds, "+") != strings.Join(tt.wantKinds, "+") {
				t.Fatalf("repairDSL kinds = %v, want %v", kinds, tt.wantKinds)
			}
		})
	}
}

// TestStripFStringQuoteEscapes pins the span discipline: only LITERAL
// f-string fragments are rewritten — plain strings (where \' is a legal
// escape and stripping \" would end the string early), placeholder bodies,
// comments, and identifier-tail `f"` stay verbatim.
func TestStripFStringQuoteEscapes(t *testing.T) {
	tests := []struct {
		name string
		src  string
		q    byte
		want string
	}{
		{
			name: "fragment escape stripped",
			src:  `say(f"it\'s {x} time")`,
			q:    '\'',
			want: `say(f"it's {x} time")`,
		},
		{
			name: "plain string untouched",
			src:  `note("it\'s fine")`,
			q:    '\'',
			want: `note("it\'s fine")`,
		},
		{
			name: "plain string after the f-string untouched",
			src:  `say(f"it\'s {x}", "keep \' raw")`,
			q:    '\'',
			want: `say(f"it's {x}", "keep \' raw")`,
		},
		{
			name: "placeholder body untouched",
			src:  `say(f"a\'b {fn("c\'d")} e\'f")`,
			q:    '\'',
			want: `say(f"a'b {fn("c\'d")} e'f")`,
		},
		{
			// Helper property only: repairLexFailure no longer dispatches '"'
			// (\" is a legal f-string escape, and a global \" strip can change
			// a misclassified plain string's value while re-parsing cleanly).
			name: "double-quote strip is span-disciplined (helper property)",
			src:  `say(f"say \"hi\" to {x}")`,
			q:    '"',
			want: `say(f"say "hi" to {x}")`,
		},
		{
			name: "comment untouched",
			src:  "# plan f\" \\' later\nsay(f\"a\\'b\")",
			q:    '\'',
			want: "# plan f\" \\' later\nsay(f\"a'b\")",
		},
		{
			name: "literal {{ stays in fragment mode",
			src:  `say(f"brace {{x}} it\'s")`,
			q:    '\'',
			want: `say(f"brace {{x}} it's")`,
		},
		{
			name: "identifier-tail f-quote is a plain string",
			src:  `use(staff"a\'b")`,
			q:    '\'',
			want: `use(staff"a\'b")`,
		},
		{
			name: "double backslash is not an escaped quote",
			src:  `say(f"a \\ b \'c {x}")`,
			q:    '\'',
			want: `say(f"a \\ b 'c {x}")`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripFStringQuoteEscapes(tt.src, tt.q); got != tt.want {
				t.Fatalf("stripFStringQuoteEscapes(%q, %q) = %q, want %q", tt.src, tt.q, got, tt.want)
			}
		})
	}
}
