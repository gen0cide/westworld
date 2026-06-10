package parser

import "testing"

// TestCStyleBooleanAliases guards the Postel's-law lexer aliases: LLM authors
// trained on C-family code emit && / || / ! constantly — these must parse as
// and / or / not instead of killing the whole routine at parse time.
func TestCStyleBooleanAliases(t *testing.T) {
	srcs := []string{
		`runtime "1.0"
routine t() { if self.hp < 4 && inventory.is_full { note("a") } }`,
		`runtime "1.0"
routine t() { if self.hp < 4 || self.fatigue > 90 { note("b") } }`,
		`runtime "1.0"
routine t() { if !inventory.is_full && (1 < 2 || !false) { note("c") } }`,
	}
	for _, src := range srcs {
		if _, err := Parse("t", src); err != nil {
			t.Errorf("C-style boolean alias failed to parse: %v\n%s", err, src)
		}
	}
	// Single & / | stay errors (no bitwise ops in the DSL).
	for _, src := range []string{
		`runtime "1.0"
routine t() { x = 1 & 2 }`,
		`runtime "1.0"
routine t() { x = 1 | 2 }`,
	} {
		if _, err := Parse("t", src); err == nil {
			t.Errorf("single &/| must remain a parse error:\n%s", src)
		}
	}
}
