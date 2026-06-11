package parser_test

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/parser"
)

// DSL-15 hygiene layer, parser half. Property: parser.Parse returns —
// an AST, an error, or both — without panicking or hanging, for ANY
// input. The 2026-06-11 mesad OOM reached the parser through l.All(),
// so the corpus reuses the lexer incident shapes and adds structural
// seeds (handlers, watchers, select, procs, bounds, try/recover) to
// drive the recursive-descent and recovery paths.
//
// FuzzParse explores; the Test* wrappers run the same invariant over the
// seed corpus plus a deterministic random sweep on every plain `go test`.

// parseSeeds is shared by the fuzz target and the non-fuzz invariant
// tests. The incident shapes mirror lexSeeds in lex/fuzz_test.go — keep
// the two lists in sync when adding trigger shapes.
var parseSeeds = []string{
	// --- structural: handlers + procs + routine (from parser_test.go) ---
	`
on chat_received(speaker, message) {
    if message == "hi" {
        say(f"hey {speaker}")
    }
}

proc nearest_spot() {
    return world.locs.fishing_spots.nearest(self.position)
}

routine fish_at_port_sarim() {
    require {
        wielded != null
        inventory.free > 0
    }
    spot = nearest_spot()
    if spot == null {
        abort "no_spot"
    }
    walk_to(x = spot.x, y = spot.y)
    while inventory.free > 0 {
        if fatigue > 90 {
            return "tired"
        }
        fish(spot)
        wait 2.8..4.5
    }
    return "banked"
}
`,
	// --- structural: file directives + bounds block ---
	`
runtime "1.0"
extends "lib/common.ws"

bounds box(0, 0, 10, 10) {
    on chat_received(s, m) { say("in box") }
    proc helper() { return 1 }
}

routine r() { x = 1 }
`,
	// --- structural: watchers + select cases (when/becomes/changes/on/timeout) ---
	`
routine r() {
    when hp() < 40 { eat() }
    when world.bank.is_open becomes true { note("open") }
    when fatigue changes { note("tick") }
    select {
        when never() { note("never") }
        on chat_received(speaker, message) { say("hi") }
        timeout 50ms { note("timeout") }
    }
}
`,
	// --- structural: repeat/until/timeout, try/recover, defer ---
	`
routine r() {
    defer close_bank()
    try {
        repeat {
            open_bank(banker)
            wait 1
        } until world.bank.is_open timeout 10
    } recover err {
        abort err
    }
}
`,
	// --- error shapes the recovery paths must survive ---
	`routine r() { x = 1 + }`,
	`routine a() {} routine b() {}`,
	`routine r() { select { 42 } }`,

	// --- the 2026-06-11 incident shapes (mirroring lex/fuzz_test.go) ---
	"runtime \"1.0\"\nroutine r() {\n    say(f\"oops {x}\n}\n",
	"note(f\"hp {self.hp}\nx = 1\n",
	`note(f"never closed`,
	`say("bad \q escape")`,
	`note(f"bad \q escape")`,
	`routine r() { note(f"hp {}") }`,
	`routine r() { note(f"have {count("coins")} gp") }`,
	`routine r() { note(f"{a[b[c[d[e[0]]]]]} and {f1(f2(f3(x)))}") }`,
	`routine r() { note(f"braces {{ and }} stay literal {x}") }`,
	`routine r() { say(format("hp {} of {} ({{raw}})", hp, max_hp)) }`,

	// --- pathological brace/quote storms ---
	`f"` + strings.Repeat("{", 64),
	strings.Repeat(`"`, 65),
	strings.Repeat(`f"{`, 32),
	strings.Repeat(`{"}`, 32),
	strings.Repeat("}}{{", 32),
}

// parseDeadline matches the spin tests in lex_spin_test.go.
const parseDeadline = 5 * time.Second

// checkParseReturns asserts Parse comes back (AST and/or error) within the
// deadline. Panics are deliberately not recovered — the fuzzer attributes
// them to the input and a plain test run fails loudly with the stack.
func checkParseReturns(t *testing.T, src string) {
	t.Helper()
	type result struct {
		file *ast.File
		err  error
	}
	done := make(chan result, 1)
	go func() {
		f, err := parser.Parse("fuzz.ws", src)
		done <- result{f, err}
	}()
	timer := time.NewTimer(parseDeadline)
	defer timer.Stop()
	select {
	case res := <-done:
		if res.file == nil && res.err == nil {
			t.Errorf("Parse returned (nil, nil) on input %s", clip(src))
		}
	case <-timer.C:
		t.Fatalf("parser.Parse did not return within %v on input %s", parseDeadline, clip(src))
	}
}

// clip bounds fuzz-generated inputs in failure messages.
func clip(s string) string {
	if len(s) > 200 {
		s = s[:200] + "..."
	}
	return fmt.Sprintf("%q", s)
}

func FuzzParse(f *testing.F) {
	for _, s := range parseSeeds {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, s string) {
		checkParseReturns(t, s)
	})
}

// TestParseSeedInvariants guards the invariant on every plain `go test`
// run — no -fuzz flag required.
func TestParseSeedInvariants(t *testing.T) {
	for _, src := range parseSeeds {
		checkParseReturns(t, src)
	}
}

// parseFuzzAtoms are token-level building blocks: random concatenations
// reach far deeper into the grammar (and its error-recovery loops) than
// random bytes do. Byte-level adjacency is the lexer sweep's job.
var parseFuzzAtoms = []string{
	"routine", "proc", "on", "bounds", "extends", "runtime",
	"when", "becomes", "changes", "select", "timeout", "repeat", "until",
	"try", "recover", "defer", "if", "elif", "else", "while", "for", "in",
	"require", "return", "abort", "wait", "break", "continue",
	"true", "false", "null", "and", "or", "not",
	"x", "f1", "self.hp", "42", "3.14", "2.8..4.5", "50ms",
	`"s"`, `f"a {x} b"`, `f"{}"`, `f"never closed`,
	"{", "}", "(", ")", "[", "]", ",", ";",
	"=", "==", "!=", "->", "=>", "!", "..", "+", "-", "*", "/", "%",
	"\n",
}

// TestParseRandomInvariants sweeps deterministic pseudo-random atom
// sequences (fixed PCG seed, identical on every CI run) through the same
// invariant.
func TestParseRandomInvariants(t *testing.T) {
	rng := rand.New(rand.NewPCG(0xDA7A, 0x5EED))
	for i := 0; i < 1500; i++ {
		var sb strings.Builder
		for n := rng.IntN(25); n > 0; n-- {
			sb.WriteString(parseFuzzAtoms[rng.IntN(len(parseFuzzAtoms))])
			if rng.IntN(4) > 0 { // mostly spaced, sometimes glued
				sb.WriteByte(' ')
			}
		}
		checkParseReturns(t, sb.String())
	}
}
