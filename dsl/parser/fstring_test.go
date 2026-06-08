package parser_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/parser"
)

// f-string codegen-tic regression tests. The mesa Act planner repeatedly emitted
// patterns the lexer/parser rejected, breaking whole authored routines:
//   - empty placeholders  f"hp {}"   → "expected expression, got FSTRING_CLOSE"
//   - whitespace bodies    f"{ x }"  → "}" lexed as a stray RBRACE
// Both now parse. A genuinely ambiguous f"{a b}" (two expressions in one
// placeholder) still errors — but cleanly, once, without derailing the routine.

// firstFString digs the FStringLit out of `note(f"...")` in routine r.
func firstFString(t *testing.T, file *ast.File) *ast.FStringLit {
	t.Helper()
	es := file.Routine.Body.Stmts[0].(*ast.ExprStmt)
	call := es.X.(*ast.CallExpr)
	fs, ok := call.Args[0].Value.(*ast.FStringLit)
	if !ok {
		t.Fatalf("arg: got %T, want *FStringLit", call.Args[0].Value)
	}
	return fs
}

func TestFStringEmptyPlaceholderAccepted(t *testing.T) {
	// An empty placeholder contributes nothing (just the literal "hp ").
	file := mustParse(t, `routine r() { note(f"hp {}") }`)
	fs := firstFString(t, file)
	if len(fs.Parts) != 1 {
		t.Fatalf("parts: got %d, want 1 (just \"hp \"): %v", len(fs.Parts), fs.Parts)
	}
	if s, ok := fs.Parts[0].(*ast.StringLit); !ok || s.Value != "hp " {
		t.Fatalf("part[0]: got %v, want StringLit(\"hp \")", fs.Parts[0])
	}
}

func TestFStringEmptyPlaceholderMid(t *testing.T) {
	// An empty placeholder between literals: "a " + (nothing) + " b".
	mustParse(t, `routine r() { note(f"a {} b") }`)
}

func TestFStringMultipleEmptyPlaceholders(t *testing.T) {
	mustParse(t, `routine r() { note(f"{}{}") }`)
}

func TestFStringWhitespacePlaceholder(t *testing.T) {
	// Spaces around the body must not break parsing: "hp " + self.hp.
	file := mustParse(t, `routine r() { note(f"hp { self.hp }") }`)
	fs := firstFString(t, file)
	if len(fs.Parts) != 2 {
		t.Fatalf("parts: got %d, want 2 (\"hp \", self.hp): %v", len(fs.Parts), fs.Parts)
	}
	if _, ok := fs.Parts[1].(*ast.MemberExpr); !ok {
		t.Fatalf("part[1]: got %T, want *MemberExpr (self.hp)", fs.Parts[1])
	}
}

func TestFStringWhitespaceEmptyPlaceholder(t *testing.T) {
	// `{ }` — whitespace-only placeholder — is also empty.
	mustParse(t, `routine r() { note(f"hp { }") }`)
}

func TestFStringAdjacentIdentsErrorsCleanly(t *testing.T) {
	_, err := parser.Parse("t.routine", `routine r() { note(f"{a b}") }`)
	if err == nil {
		t.Fatal(`expected an error for f"{a b}" (two expressions in one placeholder)`)
	}
	if !strings.Contains(err.Error(), "ONE expression") {
		t.Fatalf("error should explain the one-expression rule, got: %v", err)
	}
}

func TestFStringStillParsesValid(t *testing.T) {
	// Guard against regressing the normal forms.
	mustParse(t, `routine r() { note(f"hp {self.hp}/{self.max_hp} at ({self.position.x}, {self.position.y})") }`)
	mustParse(t, `routine r() { note(f"have {count("coins")} gp") }`) // nested plain-quote string
}
