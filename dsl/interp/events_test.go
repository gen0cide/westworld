package interp_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// yieldingCallable is a tiny Callable that yields. Used so the
// interpreter actually invokes dispatchPendingEvents inside our
// test routine.
type yieldingCallable struct {
	fn func(args []interp.Value, named map[string]interp.Value) (interp.Value, error)
}

func (y *yieldingCallable) Kind() string    { return "yielding" }
func (y *yieldingCallable) Display() string { return "<yielding>" }
func (y *yieldingCallable) Yields() bool    { return true }
func (y *yieldingCallable) Call(args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	return y.fn(args, named)
}

func TestHandlerFiresBeforeYieldingCall(t *testing.T) {
	src := `
on chat_received(speaker, message) {
    note(f"heard {speaker}: {message}")
}

routine r() {
    say("setup")
    say("after event")
    return "done"
}
`
	file, err := parser.Parse("ev.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}

	var heardCount int32
	var sayCount int32
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 16)

	it.Builtins["say"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&sayCount, 1)
		// Inject an event after the first say so the handler runs
		// before the second say.
		if atomic.LoadInt32(&sayCount) == 1 {
			it.Events <- interp.PendingEvent{
				Name: "chat_received",
				Args: []interp.Value{interp.String("alex"), interp.String("hi")},
			}
		}
		return interp.Null{}, nil
	}}
	// `note` is referenced by the handler — register a counter.
	it.Builtins["note"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&heardCount, 1)
		return interp.Null{}, nil
	}}

	res := it.RunRoutine(context.Background(), file, nil)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if atomic.LoadInt32(&heardCount) != 1 {
		t.Errorf("handler ran %d times, want 1", heardCount)
	}
	if atomic.LoadInt32(&sayCount) != 2 {
		t.Errorf("say called %d times, want 2", sayCount)
	}
}

func TestHandlerReceivesCorrectArgs(t *testing.T) {
	src := `
on chat_received(speaker, message) {
    note(speaker + ": " + message)
}

routine r() {
    say("trigger")
    return "ok"
}
`
	file, err := parser.Parse("ev.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}

	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)

	var captured interp.Value
	it.Builtins["say"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		it.Events <- interp.PendingEvent{
			Name: "chat_received",
			Args: []interp.Value{interp.String("delores"), interp.String("hello world")},
		}
		return interp.Null{}, nil
	}}
	it.Builtins["note"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		if len(args) >= 1 {
			captured = args[0]
		}
		return interp.Null{}, nil
	}}

	if res := it.RunRoutine(context.Background(), file, nil); res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v", res.Kind)
	}
	if s, _ := captured.(interp.String); string(s) != "delores: hello world" {
		t.Errorf("captured: got %v, want \"delores: hello world\"", captured)
	}
}

func TestHandlerAbortPropagatesToRoutine(t *testing.T) {
	src := `
on chat_received(speaker, message) {
    abort "interrupted"
}

routine r() {
    say("trigger")
    say("never reached")
    return "ok"
}
`
	file, err := parser.Parse("ev.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}

	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)

	var sayCount int32
	it.Builtins["say"] = &yieldingCallable{fn: func(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		atomic.AddInt32(&sayCount, 1)
		if atomic.LoadInt32(&sayCount) == 1 {
			it.Events <- interp.PendingEvent{
				Name: "chat_received",
				Args: []interp.Value{interp.String("alex"), interp.String("stop")},
			}
		}
		return interp.Null{}, nil
	}}

	res := it.RunRoutine(context.Background(), file, nil)
	if res.Kind != interp.ResultAborted {
		t.Fatalf("kind: got %v, want aborted", res.Kind)
	}
	if s, _ := res.Value.(interp.String); string(s) != "interrupted" {
		t.Errorf("reason: got %v, want interrupted", res.Value)
	}
	// The second say should never have fired.
	if atomic.LoadInt32(&sayCount) != 1 {
		t.Errorf("say count: got %d, want 1 (handler should have aborted before 2nd say)", sayCount)
	}
}

// TestBoundsFilterFiltersByLocation locks the bounds {} dispatch
// contract: a bounded handler fires for events whose (x, y) is
// inside the shape, and stays silent for events outside it.
func TestBoundsFilterFiltersByLocation(t *testing.T) {
	src := `
bounds box(100, 100, 200, 200) {
    on item_appeared(item_id, x, y) {
        note(f"inbox {item_id} at ({x},{y})")
    }
}

routine r() {
    drain()
    drain()
    return "done"
}
`
	f := mustParseFile(t, src)
	var notes []string
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)

	// Register a `box` callable matching the runtime/dsl_actions
	// version so the bounds shape evaluates to a region predicate.
	it.Builtins["box"] = &pureCallable{fn: func(args []interp.Value) interp.Value {
		x1, _ := interp.AsInt(args[0])
		y1, _ := interp.AsInt(args[1])
		x2, _ := interp.AsInt(args[2])
		y2, _ := interp.AsInt(args[3])
		pred := interp.RegionPredicate(func(x, y int) bool {
			return x >= int(x1) && x <= int(x2) && y >= int(y1) && y <= int(y2)
		})
		return interp.NewRegionPredicate("box", pred)
	}}
	it.Builtins["note"] = &pureCallable{fn: func(args []interp.Value) interp.Value {
		if s, ok := args[0].(interp.String); ok {
			notes = append(notes, string(s))
		}
		return interp.Null{}
	}}
	// drain() is a yielding no-op so dispatchPendingEvents runs
	// between events injected before each call.
	calls := 0
	it.Builtins["drain"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		calls++
		switch calls {
		case 1:
			// Inside the box — handler should fire.
			it.Events <- interp.PendingEvent{
				Name: "item_appeared",
				Args: []interp.Value{interp.Int(10), interp.Int(150), interp.Int(150)},
			}
		case 2:
			// Outside the box — handler should NOT fire.
			it.Events <- interp.PendingEvent{
				Name: "item_appeared",
				Args: []interp.Value{interp.Int(10), interp.Int(300), interp.Int(300)},
			}
		}
		return interp.Null{}, nil
	}}

	res := it.RunRoutine(context.Background(), f, nil)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("routine kind: got %v err=%v", res.Kind, res.Err)
	}
	// Only the in-box event should have produced a note.
	if len(notes) != 1 {
		t.Fatalf("expected 1 note (in-box only), got %d: %v", len(notes), notes)
	}
	if notes[0] != "inbox 10 at (150,150)" {
		t.Errorf("note content: got %q", notes[0])
	}
}

// TestBoundsNestedIntersection locks the nested-bounds semantics:
// an inner handler fires only when the event matches BOTH outer
// AND inner shapes.
func TestBoundsNestedIntersection(t *testing.T) {
	src := `
bounds box(0, 0, 100, 100) {
    bounds box(40, 40, 60, 60) {
        on item_appeared(item_id, x, y) {
            note(f"deep {x},{y}")
        }
    }
}

routine r() {
    drain()
    drain()
    drain()
    return "done"
}
`
	f := mustParseFile(t, src)
	var notes []string
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Builtins["box"] = &pureCallable{fn: func(args []interp.Value) interp.Value {
		x1, _ := interp.AsInt(args[0])
		y1, _ := interp.AsInt(args[1])
		x2, _ := interp.AsInt(args[2])
		y2, _ := interp.AsInt(args[3])
		pred := interp.RegionPredicate(func(x, y int) bool {
			return x >= int(x1) && x <= int(x2) && y >= int(y1) && y <= int(y2)
		})
		return interp.NewRegionPredicate("box", pred)
	}}
	it.Builtins["note"] = &pureCallable{fn: func(args []interp.Value) interp.Value {
		if s, ok := args[0].(interp.String); ok {
			notes = append(notes, string(s))
		}
		return interp.Null{}
	}}
	calls := 0
	it.Builtins["drain"] = &yieldingCallable{fn: func(_ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
		calls++
		switch calls {
		case 1:
			// Inside both — fires.
			it.Events <- interp.PendingEvent{
				Name: "item_appeared",
				Args: []interp.Value{interp.Int(1), interp.Int(50), interp.Int(50)},
			}
		case 2:
			// In outer only — does NOT fire (inner box rejects).
			it.Events <- interp.PendingEvent{
				Name: "item_appeared",
				Args: []interp.Value{interp.Int(1), interp.Int(80), interp.Int(80)},
			}
		case 3:
			// Outside both — does NOT fire.
			it.Events <- interp.PendingEvent{
				Name: "item_appeared",
				Args: []interp.Value{interp.Int(1), interp.Int(200), interp.Int(200)},
			}
		}
		return interp.Null{}, nil
	}}
	res := it.RunRoutine(context.Background(), f, nil)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v err=%v", res.Kind, res.Err)
	}
	if len(notes) != 1 {
		t.Fatalf("expected 1 note (inner box only), got %d: %v", len(notes), notes)
	}
	if notes[0] != "deep 50,50" {
		t.Errorf("note content: got %q", notes[0])
	}
}

func TestNoHandlerNoDispatch(t *testing.T) {
	// Routine without any handlers should still drain events without
	// running anything.
	src := `routine r() { return 42 }`
	file, err := parser.Parse("ev.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if err := validator.Validate(file); err != nil {
		t.Fatalf("validate: %v", err)
	}
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 4)
	it.Events <- interp.PendingEvent{Name: "chat_received", Args: []interp.Value{interp.String("a"), interp.String("b")}}
	res := it.RunRoutine(context.Background(), file, nil)
	if res.Kind != interp.ResultReturned {
		t.Fatalf("kind: got %v, want returned", res.Kind)
	}
	if i, _ := res.Value.(interp.Int); int64(i) != 42 {
		t.Errorf("value: got %v, want 42", res.Value)
	}
}
