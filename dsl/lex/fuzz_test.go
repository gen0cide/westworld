package lex_test

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/token"
)

// DSL-15 hygiene layer. The 2026-06-11 mesad OOM was a lexer-termination
// bug (sticky f-string mode re-emitting one ILLEGAL token forever); the
// spin-proof guard in All() now makes these properties hold for ANY input,
// and this file pins them forever:
//
//  1. All() terminates,
//  2. the stream's final token is EOF,
//  3. len(tokens) <= len(src)+3 — every non-EOF token consumes at least
//     one byte, except at most one zero-progress token before the guard
//     forces EOF, so stream length is linear in input length.
//
// FuzzLexAll explores; the Test* wrappers run the same invariants over the
// seed corpus plus a deterministic random sweep on every plain `go test`.

// lexSeeds is the seed corpus, shared by the fuzz target and the
// non-fuzz invariant tests. parser/fuzz_test.go mirrors the incident
// shapes — keep the two lists in sync when adding trigger shapes.
var lexSeeds = []string{
	// --- valid routines (lifted from lex_test.go) ---
	`routine foo() { x = 42 + 3.14 }`,
	`routine proc on if elif else while for in break continue return abort require wait true false null and or not`,
	`
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
`,
	`f"hello {speaker} you said {message}"`,
	`wait(2.8..4.5)`,

	// --- the 2026-06-11 incident shapes ---
	// unterminated f-string, newline before the closing quote (THE trigger)
	"note(f\"hp {self.hp}\nx = 1\n",
	// unterminated f-string at EOF
	`note(f"never closed`,
	// unknown escape \q — plain string and f-string variants
	`say("bad \q escape")`,
	`note(f"bad \q escape")`,
	// empty placeholder (the dominant authoring-rejection shape)
	`note(f"hp {}")`,
	// nested plain-quoted string inside a placeholder
	`note(f"have {count("coins")} gp")`,
	// deep placeholder nesting
	`note(f"{a[b[c[d[e[0]]]]]} and {f1(f2(f3(x)))}")`,
	// {{ }} literal braces — f-string today, format() templates next
	`note(f"braces {{ and }} stay literal {x}")`,
	`say(format("hp {} of {} ({{raw}})", hp, max_hp))`,

	// --- pathological brace/quote storms ---
	`f"` + strings.Repeat("{", 64),
	strings.Repeat(`"`, 65),
	strings.Repeat(`f"{`, 32),
	`f"` + strings.Repeat("{}", 64) + `"`,
	`"` + strings.Repeat(`\`, 65),
	strings.Repeat(`{"}`, 32),
	strings.Repeat("}}{{", 32),
}

// lexDeadline matches the spin tests in lex_spin_test.go: termination is
// the property, so a hang is converted into a crisp failure.
const lexDeadline = 5 * time.Second

// checkLexStream asserts the three All() invariants for one input.
func checkLexStream(t *testing.T, src string) {
	t.Helper()
	done := make(chan []token.Token, 1)
	go func() { done <- lex.New("fuzz.ws", src).All() }()
	timer := time.NewTimer(lexDeadline)
	defer timer.Stop()
	var toks []token.Token
	select {
	case toks = <-done:
	case <-timer.C:
		t.Fatalf("lexer did not terminate within %v on input %s", lexDeadline, clip(src))
	}
	if len(toks) == 0 {
		t.Fatalf("empty token stream (want at least EOF) on input %s", clip(src))
	}
	if last := toks[len(toks)-1]; last.Kind != token.EOF {
		t.Errorf("final token = %s, want EOF, on input %s", last.Kind, clip(src))
	}
	if max := len(src) + 3; len(toks) > max {
		t.Errorf("token stream has %d tokens, want <= len(src)+3 = %d, on input %s", len(toks), max, clip(src))
	}
}

// clip bounds fuzz-generated inputs in failure messages.
func clip(s string) string {
	if len(s) > 200 {
		s = s[:200] + "..."
	}
	return fmt.Sprintf("%q", s)
}

func FuzzLexAll(f *testing.F) {
	for _, s := range lexSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		checkLexStream(t, s)
	})
}

// TestLexAllSeedInvariants guards the invariants on every plain `go test`
// run — no -fuzz flag required.
func TestLexAllSeedInvariants(t *testing.T) {
	for _, src := range lexSeeds {
		checkLexStream(t, src)
	}
}

// lexFuzzAlphabet is biased toward the runes that drive the lexer's state
// machine: f-string openers, braces, quotes, escapes, comments, newlines —
// plus enough ident/number material to form real tokens and one multibyte
// rune for the UTF-8 paths.
var lexFuzzAlphabet = []rune("f\"{}\\#!&|.=<>()[]x09._ \n\tλ")

// TestLexAllRandomInvariants sweeps deterministic pseudo-random inputs
// (fixed PCG seed, identical on every CI run) through the same invariants.
func TestLexAllRandomInvariants(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xDA7A, 0x5EED))
	for i := 0; i < 2000; i++ {
		var sb strings.Builder
		for n := rng.IntN(81); n > 0; n-- {
			sb.WriteRune(lexFuzzAlphabet[rng.IntN(len(lexFuzzAlphabet))])
		}
		checkLexStream(t, sb.String())
	}
}
