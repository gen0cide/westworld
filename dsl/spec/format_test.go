package spec_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/spec"
)

func TestCountFormatPlaceholders(t *testing.T) {
	cases := map[string]int{
		"":                  0,
		"plain":             0,
		"{}":                1,
		"{} {}":             2,
		"a{}b{}c{}":         3,
		"{{}}":              0, // escaped braces, no placeholder
		"{{}":               0, // escape then lone } (literal)
		"{{{}}}":            1, // escape, placeholder, escape
		"{x}":               0, // not a placeholder — literal text
		"{ }":               0,
		"{":                 0,
		"}":                 0,
		"need {} more {}":   2,
		"{{literal}} {} ok": 1,
	}
	for tmpl, want := range cases {
		if got := spec.CountFormatPlaceholders(tmpl); got != want {
			t.Errorf("CountFormatPlaceholders(%q) = %d, want %d", tmpl, got, want)
		}
	}
}

func TestWalkFormatTemplateDecodesEscapes(t *testing.T) {
	// Reconstruct "{{{}}}" through the walker: escape '{', one
	// placeholder, escape '}'.
	var sb strings.Builder
	phs := 0
	spec.WalkFormatTemplate("{{{}}}",
		func(s string) { sb.WriteString(s) },
		func(b byte) { sb.WriteByte(b) },
		func() { sb.WriteString("X"); phs++ },
	)
	if got := sb.String(); got != "{X}" {
		t.Errorf("walk render: got %q, want %q", got, "{X}")
	}
	if phs != 1 {
		t.Errorf("placeholders fired %d times, want 1", phs)
	}
}

func TestFormatIdentPlaceholders(t *testing.T) {
	cases := map[string][]string{
		"{x}":           {"x"},
		"hello {name}!": {"name"},
		"{x} {} {y}":    {"x", "y"},
		"{_a1}":         {"_a1"},
		"{{name}}":      nil, // escaped on purpose — no warning
		"{0}":           nil, // indexed, not an ident
		"{x":            nil, // unterminated
		"{x }":          nil, // space breaks the ident
		"{}":            nil,
		"plain":         nil,
		// The closing brace claimed by a }} escape still warns — these
		// render literal {name} text just like the plain shape.
		"{name}}":    {"name"},
		"{{{name}}}": {"name"},
		"a{name}}b":  {"name"},
		"{a}}x{b}}":  {"a", "b"},
		"{name{{":    nil, // run ends in {ident but the next event is a {{ escape
		"{name{}":    nil, // ... or a placeholder
		"{na me}}":   nil, // space breaks the ident before the escape
		"{name}}{x}": {"name", "x"},
	}
	for tmpl, want := range cases {
		got := spec.FormatIdentPlaceholders(tmpl)
		if len(got) != len(want) {
			t.Errorf("FormatIdentPlaceholders(%q) = %v, want %v", tmpl, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("FormatIdentPlaceholders(%q) = %v, want %v", tmpl, got, want)
				break
			}
		}
	}
}

func TestFormatSpecRow(t *testing.T) {
	a, ok := spec.ByName("format")
	if !ok {
		t.Fatal("spec.Actions has no row for format")
	}
	if a.Kind != spec.Primitive {
		t.Errorf("format Kind = %s, want Primitive", a.Kind)
	}
	if a.BangEligible() {
		t.Error("format must not be bang-eligible (returns a bare String, not a Result)")
	}
	if !a.SubscriptionSafe() {
		t.Error("format must be subscription-safe (pure — usable anywhere an expression is)")
	}
	if a.MinArgs != 1 || a.MaxArgs != -1 {
		t.Errorf("format arity = (%d, %d), want (1, unbounded)", a.MinArgs, a.MaxArgs)
	}
}

func TestFormatIsRenderedInAPIReference(t *testing.T) {
	ref := spec.APIReference()
	if !strings.Contains(ref, "format(template, args...)") {
		t.Error("APIReference does not render the format() row")
	}
}
