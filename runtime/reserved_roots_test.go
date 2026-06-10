package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/spec"
)

// TestReservedRootsMatchRegistration pins the three-way contract between
// spec.ReservedRoots (the vocabulary), the validator (which rejects binding
// them), and this package's view registration: every registered root must be
// in the spec list, and the only spec'd root with no view must be "host"
// (reserved ahead of being wired). A divergence here is exactly the drift the
// old hand-maintained validator switch suffered.
func TestReservedRootsMatchRegistration(t *testing.T) {
	h := newTestHost()
	it := h.NewRoutineInterpreter(t.Context())

	specSet := make(map[string]bool)
	for _, r := range spec.ReservedRoots() {
		specSet[r] = true
	}
	for name := range it.Reserved {
		if !specSet[name] {
			t.Errorf("runtime registers root %q that spec.ReservedRoots does not name", name)
		}
	}
	var unwired []string
	for _, r := range spec.ReservedRoots() {
		if _, ok := it.Reserved[r]; !ok {
			unwired = append(unwired, r)
		}
	}
	if len(unwired) != 1 || unwired[0] != "host" {
		t.Errorf("unwired spec roots = %v, want exactly [host]", unwired)
	}
}
