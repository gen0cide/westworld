package runtime

import (
	"path/filepath"
	"runtime"
	"testing"
)

// TestExampleRoutinesParse asserts that every checked-in
// `examples/routines/*.routine` file passes ParseRoutineFile,
// including the filename = routine-name match. This catches
// example-routine bitrot when the loader rules change.
func TestExampleRoutinesParse(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..")
	pattern := filepath.Join(repoRoot, "examples", "routines", "*.routine")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		t.Fatalf("glob: %v", err)
	}
	if len(matches) == 0 {
		t.Fatalf("no example routines found under %s", pattern)
	}
	for _, m := range matches {
		m := m
		t.Run(filepath.Base(m), func(t *testing.T) {
			if _, err := ParseRoutineFile(m); err != nil {
				t.Errorf("ParseRoutineFile failed: %v", err)
			}
		})
	}
}
