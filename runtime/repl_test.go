package runtime

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// driveREPL feeds `input` to a REPL bound to a fresh test Host
// and returns whatever the REPL printed. Lines are newline-
// separated; an automatic `.quit` is appended so Run() terminates.
func driveREPL(t *testing.T, input string) string {
	t.Helper()
	h := newTestHost()
	var out bytes.Buffer
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}
	input += ".quit\n"
	r := h.NewREPL(context.Background(), strings.NewReader(input), &out)
	if err := r.Run(); err != nil {
		t.Fatalf("REPL.Run: %v", err)
	}
	return out.String()
}

func TestREPLBannerOnStartup(t *testing.T) {
	out := driveREPL(t, "")
	if !strings.Contains(out, "westworld REPL") {
		t.Errorf("missing banner; got:\n%s", out)
	}
}

func TestREPLEvaluatesExpression(t *testing.T) {
	out := driveREPL(t, "1 + 2 * 3")
	if !strings.Contains(out, "\n7\n") && !strings.Contains(out, ">>> 7") {
		t.Errorf("expected 7 in output; got:\n%s", out)
	}
}

func TestREPLPersistentLocals(t *testing.T) {
	out := driveREPL(t, "x = 5\nx + 1")
	if !strings.Contains(out, "6") {
		t.Errorf("expected 6 (5+1 across lines); got:\n%s", out)
	}
}

func TestREPLDotState(t *testing.T) {
	out := driveREPL(t, ".state")
	// Test host has hp=50 — should appear in the state dump.
	if !strings.Contains(out, "hp") {
		t.Errorf("expected hp line in .state output; got:\n%s", out)
	}
	if !strings.Contains(out, "50") {
		t.Errorf("expected test host hp=50 in .state output; got:\n%s", out)
	}
}

func TestREPLDotHelp(t *testing.T) {
	out := driveREPL(t, ".help")
	if !strings.Contains(out, ".quit") {
		t.Errorf("expected .quit listed in .help; got:\n%s", out)
	}
}

func TestREPLDotHelpForBuiltin(t *testing.T) {
	out := driveREPL(t, ".help walk_to")
	if !strings.Contains(out, "walk_to") {
		t.Errorf("expected walk_to in .help walk_to; got:\n%s", out)
	}
	if !strings.Contains(out, "PrimaryAction") {
		t.Errorf("expected Kind in .help walk_to; got:\n%s", out)
	}
}

func TestREPLDotBuiltins(t *testing.T) {
	out := driveREPL(t, ".builtins")
	// Spot-check a few representative names from each Kind.
	for _, want := range []string{"walk_to", "wait", "contemplate_reality", "recall", "mood"} {
		if !strings.Contains(out, want) {
			t.Errorf(".builtins missing %q; got:\n%s", want, out)
		}
	}
}

func TestREPLDotEventsRecentEmpty(t *testing.T) {
	// Bare `.events` shows the recent-events BUFFER (live state),
	// not the spec catalog. With a fresh host, it's empty.
	out := driveREPL(t, ".events")
	if !strings.Contains(out, "nothing yet") {
		t.Errorf(".events should report empty buffer; got:\n%s", out)
	}
}

func TestREPLDotEventsSpec(t *testing.T) {
	// `.events spec` shows the language event catalog.
	out := driveREPL(t, ".events spec")
	for _, want := range []string{"chat_received", "private_message"} {
		if !strings.Contains(out, want) {
			t.Errorf(".events spec missing %q; got:\n%s", want, out)
		}
	}
}

func TestREPLDotLoadAndRun(t *testing.T) {
	// Write a tiny routine to a temp file, .load and .run it.
	dir := t.TempDir()
	path := filepath.Join(dir, "echo_seven.routine")
	if err := os.WriteFile(path, []byte(`
		proc seven() { return 7 }
		routine echo_seven() { return seven() }
	`), 0o644); err != nil {
		t.Fatal(err)
	}
	out := driveREPL(t, ".run "+path)
	if !strings.Contains(out, "7") {
		t.Errorf(".run should print 7; got:\n%s", out)
	}
	if !strings.Contains(out, "returned") {
		t.Errorf(".run should report kind=returned; got:\n%s", out)
	}
}

func TestREPLOnHandlerAtPrompt(t *testing.T) {
	// `on chat_received(s, m) { ... }` at the prompt registers
	// a live handler.
	out := driveREPL(t, `on chat_received(speaker, message) { note(speaker) }`)
	// No output expected on the success path — but no ERR either.
	if strings.Contains(out, "ERR") {
		t.Errorf("registering on-handler errored: %s", out)
	}
}

func TestREPLDotAccessors(t *testing.T) {
	out := driveREPL(t, ".accessors")
	for _, want := range []string{"self.hp", "inventory.free", "world.npcs"} {
		if !strings.Contains(out, want) {
			t.Errorf(".accessors missing %q; got:\n%s", want, out)
		}
	}
}

func TestREPLParseErrorDoesNotCrash(t *testing.T) {
	// First line is garbage; second line is a valid expression.
	// REPL should print an error for line 1 and 42 for line 2.
	out := driveREPL(t, "!!! not valid !!!\n42")
	if !strings.Contains(out, "ERR") {
		t.Errorf("expected error message for bad input; got:\n%s", out)
	}
	if !strings.Contains(out, "42") {
		t.Errorf("expected 42 from recovery line; got:\n%s", out)
	}
}

func TestREPLUnknownMeta(t *testing.T) {
	out := driveREPL(t, ".bogus_command")
	if !strings.Contains(out, "unknown meta command") {
		t.Errorf("expected unknown-meta error; got:\n%s", out)
	}
}

func TestREPLLiveHostQuery(t *testing.T) {
	// The interesting one — querying a reserved entity should walk
	// through to the Host's world-state mirror.
	out := driveREPL(t, "self.hp")
	if !strings.Contains(out, "50") {
		t.Errorf("expected 50 from self.hp on test host; got:\n%s", out)
	}
}
