// Package conformance runs `.routine` files against a known trace
// and result, asserting deterministic behavior. This is the
// language's regression-test harness — every change to the lexer,
// parser, validator, or interpreter must keep these tests passing.
//
// Each test case is a pair of files in a directory:
//
//   testdata/conformance/foo.routine    — the routine source
//   testdata/conformance/foo.expected   — the expected trace
//
// The `.expected` format is a plain-text manifest:
//
//   kind: returned   # one of: returned, aborted, completed, errored, canceled
//   value: "ok"      # optional — quoted literal for string, bare for int/null/bool
//   trace:           # optional — action calls in order, one per line
//     say("hi")
//     walk_to(x=120, y=504)
//     return
//
// Lines beginning with `#` are comments. Whitespace is significant
// only as a separator. The runner builds a fresh Interpreter for
// each case with mock builtins that produce deterministic results.
package conformance

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// Case is one paired routine+expected pair.
type Case struct {
	Name       string // filename without extension
	RoutineSrc string
	Expected   Expected
}

// Expected captures the manifest contents.
type Expected struct {
	Kind  string   // returned, aborted, completed, errored, canceled
	Value string   // optional — raw quoted/unquoted text
	Trace []string // optional — action call signatures
}

// Failure describes a mismatch between expected and actual.
type Failure struct {
	Field    string
	Expected string
	Actual   string
}

func (f Failure) String() string {
	return fmt.Sprintf("%s: expected %q, got %q", f.Field, f.Expected, f.Actual)
}

// LoadDir reads all *.routine files under dir, pairing each with
// its sibling *.expected. Returns the cases in filename order.
func LoadDir(dir string) ([]Case, error) {
	pattern := filepath.Join(dir, "*.routine")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(matches)
	out := make([]Case, 0, len(matches))
	for _, m := range matches {
		base := strings.TrimSuffix(filepath.Base(m), ".routine")
		expectedPath := filepath.Join(filepath.Dir(m), base+".expected")
		src, err := os.ReadFile(m)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", m, err)
		}
		expBytes, err := os.ReadFile(expectedPath)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", expectedPath, err)
		}
		exp, err := parseExpected(string(expBytes))
		if err != nil {
			return nil, fmt.Errorf("parse expected for %s: %w", base, err)
		}
		out = append(out, Case{Name: base, RoutineSrc: string(src), Expected: exp})
	}
	return out, nil
}

// parseExpected parses the .expected manifest.
func parseExpected(text string) (Expected, error) {
	exp := Expected{}
	scanner := bufio.NewScanner(strings.NewReader(text))
	inTrace := false
	for scanner.Scan() {
		raw := scanner.Text()
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "kind:") {
			exp.Kind = strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
			inTrace = false
			continue
		}
		if strings.HasPrefix(line, "value:") {
			exp.Value = strings.TrimSpace(strings.TrimPrefix(line, "value:"))
			inTrace = false
			continue
		}
		if line == "trace:" {
			inTrace = true
			continue
		}
		if inTrace {
			exp.Trace = append(exp.Trace, line)
		}
	}
	return exp, scanner.Err()
}

// Run executes a single case under a fresh Interpreter with mock
// builtins and a trace-capturing Hooks struct. Returns Failure
// slice; empty means pass.
func Run(c Case) []Failure {
	file, err := parser.Parse(c.Name+".routine", c.RoutineSrc)
	if err != nil {
		return []Failure{{Field: "parse", Actual: err.Error()}}
	}
	if err := validator.Validate(file); err != nil {
		return []Failure{{Field: "validate", Actual: err.Error()}}
	}

	tr := &traceRecorder{}
	it := interp.New()
	it.Hooks = &interp.Hooks{
		OnAction: tr.action,
	}
	// Mock every builtin used in conformance tests. Each returns
	// Null{} so traces are deterministic. The set is small and
	// extends as new conformance cases need new mocks.
	for _, name := range []string{
		"say", "whisper", "note",
		"walk_to", "attack", "talk_to", "answer",
		"drop", "pick_up", "eat",
		"open_bank", "deposit", "withdraw", "close_bank",
		"logout", "wait",
		"mine", "fish", "chop", "cook", "cast",
		"contemplate_reality", "evaluate", "decide",
	} {
		nameCopy := name
		it.Builtins[name] = &yieldingMock{name: nameCopy}
	}

	res := it.RunRoutine(context.Background(), file, nil)

	var fails []Failure
	if got := res.Kind.String(); got != c.Expected.Kind {
		fails = append(fails, Failure{Field: "kind", Expected: c.Expected.Kind, Actual: got})
	}
	if c.Expected.Value != "" {
		got := formatValue(res.Value)
		if got != c.Expected.Value {
			fails = append(fails, Failure{Field: "value", Expected: c.Expected.Value, Actual: got})
		}
	}
	if len(c.Expected.Trace) > 0 {
		got := tr.snapshot()
		if !traceEqual(got, c.Expected.Trace) {
			fails = append(fails, Failure{
				Field:    "trace",
				Expected: strings.Join(c.Expected.Trace, " | "),
				Actual:   strings.Join(got, " | "),
			})
		}
	}
	return fails
}

// formatValue formats a Result.Value to match the .expected syntax:
// strings get quoted, ints/bools/null bare.
func formatValue(v interp.Value) string {
	if v == nil {
		return "null"
	}
	switch x := v.(type) {
	case interp.Null:
		return "null"
	case interp.String:
		return fmt.Sprintf("%q", string(x))
	}
	return v.Display()
}

// traceEqual returns true if two trace slices match. Trace lines are
// compared with whitespace normalized to single spaces.
func traceEqual(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if normalize(got[i]) != normalize(want[i]) {
			return false
		}
	}
	return true
}

func normalize(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// traceRecorder collects action call signatures in order.
type traceRecorder struct {
	mu     sync.Mutex
	events []string
}

func (t *traceRecorder) action(name string, args []interp.Value) {
	t.mu.Lock()
	defer t.mu.Unlock()
	var parts []string
	for _, a := range args {
		parts = append(parts, formatValue(a))
	}
	t.events = append(t.events, fmt.Sprintf("%s(%s)", name, strings.Join(parts, ", ")))
}

func (t *traceRecorder) snapshot() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]string, len(t.events))
	copy(out, t.events)
	return out
}

// yieldingMock is a Callable that records nothing itself (the
// interpreter's OnAction hook does the recording) but yields so the
// hook fires.
type yieldingMock struct{ name string }

func (m *yieldingMock) Kind() string    { return "mock" }
func (m *yieldingMock) Display() string { return "<mock " + m.name + ">" }
func (m *yieldingMock) Yields() bool    { return true }
func (m *yieldingMock) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	return interp.Null{}, nil
}
