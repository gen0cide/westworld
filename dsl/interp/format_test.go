package interp_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
)

func TestFormatPositionalPlaceholders(t *testing.T) {
	got := run(t, `routine r() { return format("need {} more {}", 5, "logs") }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "need 5 more logs" {
		t.Errorf("got %v, want %q", got, "need 5 more logs")
	}
}

func TestFormatNoPlaceholders(t *testing.T) {
	got := run(t, `routine r() { return format("plain") }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "plain" {
		t.Errorf("got %v, want %q", got, "plain")
	}
}

func TestFormatBraceEscapes(t *testing.T) {
	got := run(t, `routine r() { return format("{{}} {{x}} {} }}", 1) }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "{} {x} 1 }" {
		t.Errorf("got %v, want %q", got, "{} {x} 1 }")
	}
}

func TestFormatLoneBracesAreLiteral(t *testing.T) {
	got := run(t, `routine r() { return format("a { b } c {x}") }`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "a { b } c {x}" {
		t.Errorf("got %v, want %q", got, "a { b } c {x}")
	}
}

// A literal-template mismatch is a VALIDATION error (see the validator
// tests); the runtime FORMAT_MISMATCH path needs a dynamic template.
func TestFormatMismatchAbortsWithTypedError(t *testing.T) {
	res := run(t, `routine r() {
		t = "{} {}"
		return format(t, 1)
	}`)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("got kind %v, want aborted", res.Kind)
	}
	e, ok := res.Value.(*interp.Error)
	if !ok {
		t.Fatalf("abort reason is %T, want *interp.Error", res.Value)
	}
	if e.Code != interp.FORMAT_MISMATCH {
		t.Errorf("code: got %s, want FORMAT_MISMATCH", e.Code)
	}
	if e.Reason != "template has 2 placeholders, got 1 args" {
		t.Errorf("reason: got %q", e.Reason)
	}
}

func TestFormatMismatchIsRecoverable(t *testing.T) {
	got := run(t, `routine r() {
		t = "{} {}"
		out = "unset"
		try {
			out = format(t, 1, 2, 3)
		} recover err {
			return err.code
		}
		return out
	}`).Value
	if s, ok := got.(interp.String); !ok || string(s) != "FORMAT_MISMATCH" {
		t.Errorf("got %v, want %q", got, "FORMAT_MISMATCH")
	}
}

// The contract's consistency point: format renders each arg with the
// SAME value->string conversion f-string interpolation uses. Render
// the same value both ways and require identical output.
func TestFormatRendersLikeFStrings(t *testing.T) {
	exprs := []string{
		`42`, `-7`, `2.5`, `2.0`, `-0.5`,
		`"hi"`, `""`, `true`, `false`, `null`,
		`[1, "a", null, [2, 3]]`, `1..3`,
	}
	for _, expr := range exprs {
		src := fmt.Sprintf(`routine r() { v = %s return [f"{v}", format("{}", v)] }`, expr)
		got := run(t, src).Value
		list, ok := got.(*interp.List)
		if !ok || len(list.Items) != 2 {
			t.Fatalf("%s: got %v, want 2-element list", expr, got)
		}
		fstr := list.Items[0].Display()
		formatted := list.Items[1].Display()
		if fstr != formatted {
			t.Errorf("%s: f-string %q != format %q", expr, fstr, formatted)
		}
	}
}

func TestFormatTemplateMustBeString(t *testing.T) {
	res := run(t, `routine r() {
		t = 42
		return format(t)
	}`)
	if res.Kind != interp.ResultErrored {
		t.Fatalf("got kind %v, want errored", res.Kind)
	}
	if res.Err == nil || !strings.Contains(res.Err.Msg, "template must be a string") {
		t.Errorf("want template-type error, got %v", res.Err)
	}
}

// First-class reference works (f = format), and named args are
// rejected at runtime — placeholders are positional-only. (Direct
// format(x=1) call sites are already rejected by the validator; the
// first-class alias dodges that, so the runtime check matters.)
func TestFormatFirstClassRejectsNamedArgs(t *testing.T) {
	res := run(t, `routine r() {
		f = format
		return f("{}", x = 1)
	}`)
	if res.Kind != interp.ResultErrored {
		t.Fatalf("got kind %v, want errored", res.Kind)
	}
	if res.Err == nil || !strings.Contains(res.Err.Msg, "positional args only") {
		t.Errorf("want positional-only error, got %v", res.Err)
	}
}

// format resolves intrinsically, AHEAD of Builtins — a host-bridge
// registration (e.g. the spec-driven NOT_IMPLEMENTED stub) can never
// shadow it.
func TestFormatIntrinsicBeatsRegisteredBuiltin(t *testing.T) {
	stub := callableFunc(func(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
		return nil, fmt.Errorf("bridge stub called")
	})
	got := run(t, `routine r() { return format("{}", "ok") }`, withBuiltin("format", stub)).Value
	if s, ok := got.(interp.String); !ok || string(s) != "ok" {
		t.Errorf("got %v, want %q", got, "ok")
	}
}

// format output honors the same MaxStringLen cap f-strings enforce.
func TestFormatHonorsStringCap(t *testing.T) {
	res := run(t, `routine r() { return format("{}{}", "abcd", "efgh") }`,
		func(i *interp.Interpreter) { i.Caps.MaxStringLen = 4 })
	if res.Kind != interp.ResultAborted {
		t.Fatalf("got kind %v, want aborted", res.Kind)
	}
	if s, ok := res.Value.(interp.String); !ok || !strings.Contains(string(s), "string_too_large") {
		t.Errorf("abort reason: got %v, want string_too_large", res.Value)
	}
}
