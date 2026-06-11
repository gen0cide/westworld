package lex_test

import (
	"testing"
	"time"

	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/token"
)

// The 2026-06-11 mesad OOM: an unterminated f-string left the lexer's
// inFString mode sticky on the error paths, so All() re-emitted the same
// ILLEGAL token from the same offset forever — an unbounded, GC-live token
// stream on a wedged Act-handler goroutine. These tests pin termination.

// allWithin runs All() in a goroutine and fails fast instead of hanging the
// suite if the lexer regresses into the spin.
func allWithin(t *testing.T, src string) []token.Token {
	t.Helper()
	done := make(chan []token.Token, 1)
	go func() { done <- lex.New("t.spin", src).All() }()
	select {
	case out := <-done:
		return out
	case <-time.After(5 * time.Second):
		t.Fatal("lexer did not terminate: unterminated-f-string spin is back")
		return nil
	}
}

func TestUnterminatedFStringNewlineTerminates(t *testing.T) {
	// The live trigger shape: an LLM-authored f-string missing its closing
	// quote before the line break.
	src := "note(f\"hp {self.hp}\nx = 1\n"
	out := allWithin(t, src)

	if last := out[len(out)-1]; last.Kind != token.EOF {
		t.Fatalf("stream must end in EOF, ended in %v", last.Kind)
	}
	illegal := 0
	for _, tk := range out {
		if tk.Kind == token.ILLEGAL {
			illegal++
		}
	}
	if illegal == 0 {
		t.Fatal("want at least one ILLEGAL token for the unterminated f-string")
	}
	if illegal > 2 {
		t.Fatalf("ILLEGAL token repeated %d times — non-advancing error state", illegal)
	}
}

func TestUnterminatedFStringEOFTerminates(t *testing.T) {
	out := allWithin(t, `note(f"never closed`)
	if last := out[len(out)-1]; last.Kind != token.EOF {
		t.Fatalf("stream must end in EOF, ended in %v", last.Kind)
	}
}

// TestLexerExitsFStringModeAtTerminalError pins the ROOT-CAUSE fix
// independently of All()'s zero-progress guard: it drives Next()
// directly (no guard) and asserts the token after the unterminated-
// f-string ILLEGAL is not another ILLEGAL at the same offset. Without
// this, reverting only the lexer-side inFString clearing would still
// pass every All()-based invariant — the guard truncates the stream —
// silently degrading broken-f-string parses from precise diagnostics
// to a guard-truncated stream.
func TestLexerExitsFStringModeAtTerminalError(t *testing.T) {
	for name, src := range map[string]string{
		"newline": "note(f\"hp {self.hp}\nx = 1\n",
		"eof":     `note(f"never closed`,
	} {
		t.Run(name, func(t *testing.T) {
			l := lex.New("t.spin", src)
			var ill token.Token
			found := false
			// Bounded: Next() always returns, so this cannot hang even on a
			// fully regressed lexer.
			for i := 0; i <= len(src)+3 && !found; i++ {
				tk := l.Next()
				if tk.Kind == token.EOF {
					break
				}
				if tk.Kind == token.ILLEGAL {
					ill, found = tk, true
				}
			}
			if !found {
				t.Fatalf("no ILLEGAL token for the unterminated f-string in %q", src)
			}
			next := l.Next()
			if next.Kind == token.ILLEGAL && next.Pos.Offset == ill.Pos.Offset {
				t.Fatalf("token after the terminal f-string ILLEGAL is the SAME ILLEGAL (offset %d) — sticky inFString mode is back", ill.Pos.Offset)
			}
		})
	}
}

func TestParseUnterminatedFStringErrorsInsteadOfHanging(t *testing.T) {
	src := "runtime \"1.0\"\nroutine r() {\n    say(f\"oops {x}\n}\n"
	done := make(chan error, 1)
	go func() {
		_, err := parser.Parse("t.spin", src)
		done <- err
	}()
	select {
	case err := <-done:
		// The property under test is termination-with-error; the exact
		// diagnostic varies by where the ILLEGAL lands (placeholder-shape
		// vs unterminated-f-string), and either is actionable.
		if err == nil {
			t.Fatal("want a parse error for the unterminated f-string")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("parser.Parse did not return: lexer spin reached the parser")
	}
}
