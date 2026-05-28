package lex_test

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/token"
)

func TestSimpleTokens(t *testing.T) {
	src := `routine foo() { x = 42 + 3.14 }`
	got := lex.New("t.routine", src).All()
	want := []token.Kind{
		token.ROUTINE,
		token.IDENT, // foo
		token.LPAREN,
		token.RPAREN,
		token.LBRACE,
		token.IDENT, // x
		token.ASSIGN,
		token.INT,   // 42
		token.ADD,
		token.FLOAT, // 3.14
		token.RBRACE,
		token.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(want), got)
	}
	for i := range got {
		if got[i].Kind != want[i] {
			t.Errorf("token[%d]: got %s, want %s (lexeme=%q)", i, got[i].Kind, want[i], got[i].Lexeme)
		}
	}
}

func TestKeywords(t *testing.T) {
	src := `routine proc on if elif else while for in break continue return abort require wait true false null and or not`
	got := lex.New("t.routine", src).All()
	want := []token.Kind{
		token.ROUTINE, token.PROC, token.ON, token.IF, token.ELIF, token.ELSE,
		token.WHILE, token.FOR, token.IN, token.BREAK, token.CONTINUE, token.RETURN,
		token.ABORT, token.REQUIRE, token.WAIT, token.TRUE, token.FALSE, token.NULL,
		token.AND, token.OR, token.NOT, token.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Kind != want[i] {
			t.Errorf("token[%d]: got %s (%q), want %s", i, got[i].Kind, got[i].Lexeme, want[i])
		}
	}
}

func TestOperatorsAndPunct(t *testing.T) {
	src := `+ - * / % == != < <= > >= = ( ) { } [ ] , . .. : ; ->`
	got := lex.New("t.routine", src).All()
	want := []token.Kind{
		token.ADD, token.SUB, token.MUL, token.QUO, token.REM,
		token.EQ, token.NEQ, token.LT, token.LTE, token.GT, token.GTE,
		token.ASSIGN, token.LPAREN, token.RPAREN, token.LBRACE, token.RBRACE,
		token.LBRACK, token.RBRACK, token.COMMA, token.DOT, token.DOTDOT,
		token.COLON, token.SEMICOL, token.ARROW, token.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(got), len(want))
	}
	for i := range got {
		if got[i].Kind != want[i] {
			t.Errorf("token[%d]: got %s, want %s", i, got[i].Kind, want[i])
		}
	}
}

func TestStrings(t *testing.T) {
	src := `"hello" "with\nescape" "with \"quote\""`
	got := lex.New("t.routine", src).All()
	if len(got) != 4 { // 3 strings + EOF
		t.Fatalf("got %d tokens, want 4: %v", len(got), got)
	}
	wantLexemes := []string{"hello", "with\nescape", `with "quote"`}
	for i, w := range wantLexemes {
		if got[i].Kind != token.STRING {
			t.Errorf("token[%d]: got kind %s, want STRING", i, got[i].Kind)
		}
		if got[i].Lexeme != w {
			t.Errorf("token[%d]: got lexeme %q, want %q", i, got[i].Lexeme, w)
		}
	}
}

func TestComments(t *testing.T) {
	src := `# leading comment
routine foo() {  # trailing
    x = 1   # value
}`
	got := lex.New("t.routine", src).All()
	wantKinds := []token.Kind{
		token.ROUTINE,
		token.IDENT, // foo
		token.LPAREN, token.RPAREN, token.LBRACE,
		token.IDENT, token.ASSIGN, token.INT, // x = 1
		token.RBRACE, token.EOF,
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(wantKinds), got)
	}
	for i := range got {
		if got[i].Kind != wantKinds[i] {
			t.Errorf("token[%d]: got %s, want %s", i, got[i].Kind, wantKinds[i])
		}
	}
}

func TestRangeLiteral(t *testing.T) {
	src := `wait(2.8..4.5)`
	got := lex.New("t.routine", src).All()
	wantKinds := []token.Kind{
		token.WAIT, token.LPAREN, token.FLOAT, token.DOTDOT, token.FLOAT, token.RPAREN, token.EOF,
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(wantKinds), got)
	}
	for i := range got {
		if got[i].Kind != wantKinds[i] {
			t.Errorf("token[%d]: got %s, want %s (lex=%q)", i, got[i].Kind, wantKinds[i], got[i].Lexeme)
		}
	}
	if got[2].Lexeme != "2.8" || got[4].Lexeme != "4.5" {
		t.Errorf("range literal lexemes wrong: got %q .. %q", got[2].Lexeme, got[4].Lexeme)
	}
}

func TestFStringSimple(t *testing.T) {
	src := `f"hi {name}!"`
	got := lex.New("t.routine", src).All()
	wantKinds := []token.Kind{
		token.FSTRING_OPEN,  // "hi "
		token.IDENT,         // name
		token.FSTRING_CLOSE, // "!"
		token.EOF,
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(wantKinds), got)
	}
	for i := range got {
		if got[i].Kind != wantKinds[i] {
			t.Errorf("token[%d]: got %s (%q), want %s", i, got[i].Kind, got[i].Lexeme, wantKinds[i])
		}
	}
	if got[0].Lexeme != "hi " {
		t.Errorf("fstring open lex: got %q, want %q", got[0].Lexeme, "hi ")
	}
	if got[2].Lexeme != "!" {
		t.Errorf("fstring close lex: got %q, want %q", got[2].Lexeme, "!")
	}
}

func TestFStringMultipleInterpolations(t *testing.T) {
	src := `f"hello {speaker} you said {message}"`
	got := lex.New("t.routine", src).All()
	wantKinds := []token.Kind{
		token.FSTRING_OPEN,  // "hello "
		token.IDENT,         // speaker
		token.FSTRING_PART,  // " you said "
		token.IDENT,         // message
		token.FSTRING_CLOSE, // ""
		token.EOF,
	}
	if len(got) != len(wantKinds) {
		t.Fatalf("got %d tokens, want %d: %v", len(got), len(wantKinds), got)
	}
	for i := range got {
		if got[i].Kind != wantKinds[i] {
			t.Errorf("token[%d]: got %s (%q), want %s", i, got[i].Kind, got[i].Lexeme, wantKinds[i])
		}
	}
}

func TestPositionsTrackLineCol(t *testing.T) {
	src := "routine foo() {\n  x = 1\n}"
	got := lex.New("t.routine", src).All()
	// x should be on line 2, col 3
	for _, tk := range got {
		if tk.Lexeme == "x" {
			if tk.Pos.Line != 2 {
				t.Errorf("x at line %d, want 2", tk.Pos.Line)
			}
			if tk.Pos.Column != 3 {
				t.Errorf("x at col %d, want 3", tk.Pos.Column)
			}
			return
		}
	}
	t.Error("did not find x token")
}

func TestUnterminatedString(t *testing.T) {
	src := `"unclosed`
	got := lex.New("t.routine", src).All()
	if len(got) < 1 || got[0].Kind != token.ILLEGAL {
		t.Fatalf("expected ILLEGAL token, got: %v", got)
	}
}

func TestSampleRoutine(t *testing.T) {
	src := `
# fish_at_swamp.ws
on chat_received(speaker, message) {
    if message == "hi" {
        say("hey")
    }
}

proc nearest() {
    return world.locs.fishing_spots.nearest(self.position)
}

routine fish_at_swamp() {
    require {
        wielded != null
    }
    spot = nearest()
    while inventory.free > 0 {
        if fatigue > 90 {
            abort "tired"
        }
        fish(spot)
    }
    return "banked"
}
`
	l := lex.New("fish.routine", src)
	for {
		tk := l.Next()
		if tk.Kind == token.EOF {
			break
		}
		if tk.Kind == token.ILLEGAL {
			t.Errorf("unexpected ILLEGAL: %s", tk)
		}
	}
	if errs := l.Errors(); len(errs) != 0 {
		t.Errorf("unexpected lex errors: %v", errs)
	}
}
