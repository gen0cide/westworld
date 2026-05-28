package validator_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

func parseOrFail(t *testing.T, src string) *parsedFile {
	t.Helper()
	f, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return &parsedFile{file: f}
}

// parsedFile lets us call Validate.
type parsedFile struct{ file interface{} }

// validate is a thin shim around validator.Validate so test bodies
// stay readable.
func validate(t *testing.T, src string) error {
	t.Helper()
	f, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return validator.Validate(f)
}

func mustValidate(t *testing.T, src string) {
	t.Helper()
	if err := validate(t, src); err != nil {
		t.Fatalf("validate: unexpected error: %v\n--- src ---\n%s", err, src)
	}
}

func wantError(t *testing.T, src, want string) {
	t.Helper()
	err := validate(t, src)
	if err == nil {
		t.Fatalf("expected error containing %q, got nil\n--- src ---\n%s", want, src)
	}
	if !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not contain %q\n--- src ---\n%s", err.Error(), want, src)
	}
}

// ----- happy paths -----

func TestEmptyRoutineValidates(t *testing.T) {
	mustValidate(t, `routine r() {}`)
}

func TestParamsResolveInBody(t *testing.T) {
	mustValidate(t, `routine r(target) { walk_to(x = target.x, y = target.y) }`)
}

func TestProcCallableFromRoutine(t *testing.T) {
	mustValidate(t, `
		proc helper(a, b) { return a + b }
		routine r() { x = helper(1, 2) }
	`)
}

func TestReservedNamesAccessible(t *testing.T) {
	mustValidate(t, `routine r() {
		if self.hp < 10 { say("ouch") }
		if inventory.free > 0 { walk_to(x = 100, y = 200) }
	}`)
}

func TestRequireHoistsToRoutineField(t *testing.T) {
	f, err := parser.Parse("t.routine", `
		routine r() {
			require {
				inventory.free > 0
				self.fatigue < 90
			}
			spot = world.locs.fishing_spots
			fish(spot)
		}
	`)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(f); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if f.Routine.Require == nil {
		t.Fatal("expected Require to be hoisted, got nil")
	}
	if len(f.Routine.Require.Conds) != 2 {
		t.Errorf("conds: got %d, want 2", len(f.Routine.Require.Conds))
	}
	// And the body should no longer contain the RequireBlock.
	for _, s := range f.Routine.Body.Stmts {
		if _, ok := s.(interface{ Conds() []any }); ok {
			t.Error("body still contains RequireBlock after hoist")
		}
	}
}

func TestEventHandlerArityAccepted(t *testing.T) {
	mustValidate(t, `
		on chat_received(speaker, message) { say(f"hi {speaker}") }
		on inventory_full() { say("full") }
		on attacked_by(other) { say(f"attacked by {other}") }
		routine r() {}
	`)
}

func TestStdlibCallsAllowed(t *testing.T) {
	mustValidate(t, `routine r() {
		choice = contemplate_reality("what to do?")
		danger = evaluate("is this dangerous?")
		pick = decide(["a", "b", "c"])
		note("seen something interesting")
	}`)
}

// ----- error cases -----

func TestUnboundIdentifier(t *testing.T) {
	wantError(t, `routine r() { x = nonexistent_var }`, `unbound identifier "nonexistent_var"`)
}

func TestUnknownEvent(t *testing.T) {
	wantError(t, `on banana() {} routine r() {}`, `unknown event "banana"`)
}

func TestEventArityMismatch(t *testing.T) {
	wantError(t, `on chat_received(only_one) {} routine r() {}`, `expects 2 param(s)`)
}

func TestBreakOutsideLoop(t *testing.T) {
	wantError(t, `routine r() { break }`, `break outside of loop`)
}

func TestContinueOutsideLoop(t *testing.T) {
	wantError(t, `routine r() { continue }`, `continue outside of loop`)
}

func TestReturnOutsideRoutineOrProc(t *testing.T) {
	wantError(t, `on chat_received(s, m) { return "no" } routine r() {}`, `return outside of routine or proc`)
}

func TestAbortInsideProc(t *testing.T) {
	wantError(t, `proc helper() { abort "no" } routine r() {}`, `abort outside of routine`)
}

func TestActionInsideProc(t *testing.T) {
	wantError(t, `proc helper() { walk_to(x = 1, y = 2) } routine r() {}`, `forbidden inside a proc`)
}

func TestActionInsideRequire(t *testing.T) {
	wantError(t, `routine r() { require { say("nope") } }`, `forbidden inside a require block`)
}

func TestRequireNotFirstStmt(t *testing.T) {
	wantError(t, `routine r() {
		x = 1
		require { true }
	}`, `require block must be the first statement`)
}

func TestAssignToReservedName(t *testing.T) {
	wantError(t, `routine r() { self = null }`, `cannot assign to reserved name "self"`)
}

func TestReservedAsParam(t *testing.T) {
	wantError(t, `routine r(self) {}`, `shadows a reserved variable`)
}

func TestDuplicateProc(t *testing.T) {
	wantError(t, `
		proc a() {}
		proc a() {}
		routine r() {}
	`, `duplicate proc "a"`)
}

func TestDuplicateParam(t *testing.T) {
	wantError(t, `routine r(x, x) {}`, `duplicate parameter "x"`)
}

func TestCallUndefinedFunction(t *testing.T) {
	wantError(t, `routine r() { nonexistent_function() }`, `call to undefined "nonexistent_function"`)
}

func TestActionArityWrong(t *testing.T) {
	wantError(t, `routine r() { walk_to() }`, `walk_to takes at least 1`)
}

func TestProcArityWrong(t *testing.T) {
	wantError(t, `
		proc helper(a, b) { return a + b }
		routine r() { helper(1) }
	`, `proc "helper" expects 2`)
}

func TestProcDefaultArgs(t *testing.T) {
	// helper(a, b=5) should accept either 1 or 2 args.
	mustValidate(t, `
		proc helper(a, b = 5) { return a + b }
		routine r() {
			x = helper(1)
			y = helper(1, 2)
		}
	`)
}

// ----- accumulated errors -----

func TestValidatorReportsMultipleErrors(t *testing.T) {
	src := `routine r() {
		break
		continue
		nonexistent()
	}`
	err := validate(t, src)
	if err == nil {
		t.Fatal("expected errors")
	}
	merr, ok := err.(*validator.MultiError)
	if !ok {
		t.Fatalf("expected *MultiError, got %T", err)
	}
	if len(merr.Errs) < 3 {
		t.Errorf("expected 3+ errors, got %d: %v", len(merr.Errs), merr.Errs)
	}
}

// ----- realistic example -----

func TestRealisticRoutineValidates(t *testing.T) {
	src := `
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
        self.wielded != null
        inventory.free > 0
    }

    spot = nearest_spot()
    if spot == null {
        abort "no_spot"
    }

    walk_to(x = spot.x, y = spot.y)

    while inventory.free > 0 {
        if self.fatigue > 90 {
            return "tired"
        }
        fish(spot)
        wait 2.8..4.5
    }

    return "banked"
}
`
	mustValidate(t, src)
}
