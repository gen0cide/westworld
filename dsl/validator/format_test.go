package validator_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// ----- literal-template static arity -----

func TestFormatLiteralArityMatches(t *testing.T) {
	mustValidate(t, `routine r() { say(format("need {} more {}", 5, "logs")) }`)
}

func TestFormatLiteralNoPlaceholders(t *testing.T) {
	mustValidate(t, `routine r() { x = format("plain") }`)
}

func TestFormatLiteralTooFewArgs(t *testing.T) {
	wantError(t, `routine r() { x = format("{} {}", 1) }`,
		"format template has 2 placeholder(s), got 1 arg(s)")
}

func TestFormatLiteralTooManyArgs(t *testing.T) {
	wantError(t, `routine r() { x = format("{}", 1, 2) }`,
		"format template has 1 placeholder(s), got 2 arg(s)")
}

func TestFormatEscapesDoNotCountAsPlaceholders(t *testing.T) {
	mustValidate(t, `routine r() { x = format("{{}} {}", 1) }`)
	wantError(t, `routine r() { x = format("{{}}", 1) }`,
		"format template has 0 placeholder(s), got 1 arg(s)")
}

// ----- non-literal templates skip the static check -----

func TestFormatDynamicTemplateSkipsStaticCheck(t *testing.T) {
	mustValidate(t, `routine r(t) { x = format(t, 1, 2, 3) }`)
	mustValidate(t, `routine r() { x = format(f"hp {self.hp} {{}}", 1) }`)
}

// ----- call-shape rules -----

func TestFormatNamedArgsRejected(t *testing.T) {
	wantError(t, `routine r() { x = format("{}", x = 1) }`,
		"format takes positional args only")
}

func TestFormatRequiresTemplateArg(t *testing.T) {
	wantError(t, `routine r() { x = format() }`,
		"format takes at least 1 arg(s), got 0")
}

func TestFormatBangVariantRejected(t *testing.T) {
	wantError(t, `routine r() { x = format!("{}", 1) }`,
		"doesn't return Result")
}

// format is pure — legal inside `when` predicates (subscription-safe).
func TestFormatAllowedInWhenPredicate(t *testing.T) {
	mustValidate(t, `routine r() { when format("{}", self.hp) == "0" { say("dead") } }`)
}

// ----- the {ident} muscle-memory warning -----

func validateWithWarnings(t *testing.T, src string) ([]error, error) {
	t.Helper()
	f, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return validator.ValidateWithWarnings(f)
}

func TestFormatIdentPlaceholderWarns(t *testing.T) {
	warns, err := validateWithWarnings(t, `routine r() { x = format("hello {name}, {} logs", 5) }`)
	if err != nil {
		t.Fatalf("validate: unexpected error: %v", err)
	}
	if len(warns) != 1 {
		t.Fatalf("warnings: got %d (%v), want 1", len(warns), warns)
	}
	if !strings.Contains(warns[0].Error(), "{name}") {
		t.Errorf("warning %q does not name {name}", warns[0].Error())
	}
}

func TestFormatEscapedIdentDoesNotWarn(t *testing.T) {
	warns, err := validateWithWarnings(t, `routine r() { x = format("hello {{name}}") }`)
	if err != nil {
		t.Fatalf("validate: unexpected error: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("warnings: got %v, want none — {{name}} is a deliberate escape", warns)
	}
}

func TestFormatDynamicTemplateDoesNotWarn(t *testing.T) {
	// The warning must not fire on non-literal templates.
	warns, err := validateWithWarnings(t, `routine r(t) { x = format(t) }`)
	if err != nil {
		t.Fatalf("validate: unexpected error: %v", err)
	}
	if len(warns) != 0 {
		t.Errorf("warnings: got %v, want none for a dynamic template", warns)
	}
}
