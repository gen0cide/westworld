package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/spec"
)

// Cross-package consistency: every spec.Actions entry that is NOT
// marked NotYetImplemented must have a Go handler in
// actionHandlers. Conversely, every handler in actionHandlers must
// correspond to a spec.Actions entry (no orphan implementations).
//
// These tests fail loudly if someone:
//   - adds a spec.Actions entry without a Go wrapper
//   - adds a Go wrapper without a spec.Actions entry
//   - flips NotYetImplemented true→false without writing the wrapper
//   - renames a spec entry without updating actionHandlers

func TestEverySpecActionWithImplHasHandler(t *testing.T) {
	for _, a := range spec.Actions {
		if a.NotYetImplemented {
			continue
		}
		if _, ok := actionHandlers[a.Name]; !ok {
			t.Errorf("spec.Actions entry %q (%s) is not marked NotYetImplemented "+
				"but has no handler in runtime/dsl_actions.go::actionHandlers — "+
				"either add the wrapper or mark the spec entry NotYetImplemented",
				a.Name, a.Kind)
		}
	}
}

func TestEveryHandlerHasSpecEntry(t *testing.T) {
	for name := range actionHandlers {
		if _, ok := spec.ByName(name); !ok {
			t.Errorf("runtime actionHandlers has handler for %q "+
				"but spec.Actions has no entry — "+
				"either add a spec.Actions row or delete the orphan handler",
				name)
		}
	}
}

// translateEvent in dsl_events.go emits one of a small set of DSL
// event names. Every one of those must appear in spec.Events.
func TestTranslateEventOutputNamesAreInSpec(t *testing.T) {
	emittedNames := []string{
		"chat_received",
		"private_message",
		"server_message",
		"coords_changed",
		"trade_request",
	}
	for _, name := range emittedNames {
		if _, ok := spec.EventByName(name); !ok {
			t.Errorf("translateEvent emits %q but spec.Events has no entry", name)
		}
	}
}

func TestEveryRegisteredCallableIsKnownToValidator(t *testing.T) {
	// Build a minimal interpreter via the bridge and confirm every
	// registered name is either in spec.Actions (with or without
	// bang) or is a reserved entity name.
	h := New(Options{})
	defer h.Close()
	it := h.NewRoutineInterpreter(t.Context())

	for name := range it.Builtins {
		base, _ := spec.StripBang(name)
		if _, ok := spec.ByName(base); !ok {
			t.Errorf("bridge registered builtin %q (base %q) is not in spec.Actions", name, base)
		}
	}
}
