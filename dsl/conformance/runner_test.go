package conformance_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/gen0cide/westworld/dsl/conformance"
)

// TestConformanceSuite runs every paired .routine + .expected pair
// in testdata/conformance/. A new case is added by dropping the
// two files into the directory — no test code changes required.
func TestConformanceSuite(t *testing.T) {
	// The test runs from the package dir, so resolve testdata
	// relative to repo root.
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")
	dir := filepath.Join(repoRoot, "testdata", "conformance")

	cases, err := conformance.LoadDir(dir)
	if err != nil {
		t.Fatalf("LoadDir: %v", err)
	}
	if len(cases) == 0 {
		t.Fatalf("no conformance cases found under %s", dir)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			fails := conformance.Run(c)
			if len(fails) == 0 {
				return
			}
			for _, f := range fails {
				t.Errorf("%s", f)
			}
		})
	}
}
