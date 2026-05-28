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
