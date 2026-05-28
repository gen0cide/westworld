package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/validator"
)

// This file is the public entry point for running .routine files on
// a Host. Step 6 of the DSL plan.
//
// Workflow:
//
//   1. ParseRoutineFile parses + validates a .routine file from
//      disk.
//   2. NewRoutineInterpreter builds an interp.Interpreter wired to
//      this Host (reserved entities + action callables).
//   3. Run executes the routine against the live OpenRSC connection.
//
// Resource caps (op budget, wall clock, recursion, memory) are
// deferred to step 7 — for now, callers control termination via
// the passed context.

// RoutineFile is a parsed + validated .routine source file ready to
// hand to NewRoutineInterpreter.Run().
type RoutineFile struct {
	Path string
	File *ast.File
}

// ParseRoutineFile reads `path`, parses it, and validates it. Returns
// any parse or validation error with full source position.
func ParseRoutineFile(path string) (*RoutineFile, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	file, err := parser.Parse(path, string(raw))
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := validator.Validate(file); err != nil {
		return nil, fmt.Errorf("validate %s: %w", path, err)
	}
	if file.Routine == nil {
		return nil, fmt.Errorf("%s: no routine declaration (only procs/handlers)", path)
	}
	return &RoutineFile{Path: path, File: file}, nil
}

// NewRoutineInterpreter constructs an interp.Interpreter pre-loaded
// with this Host's reserved entities, action callables, and an
// event-translator goroutine that pipes typed bus events into the
// interpreter's PendingEvent channel.
//
// The returned interpreter is bound to `ctx` — every action
// invocation uses ctx as its deadline / cancellation signal, and
// the translator goroutine exits when ctx is canceled.
//
// Cancelling ctx interrupts any in-flight blocking action (walk_to,
// pick_up, wait) and the routine terminates with ResultCanceled.
func (h *Host) NewRoutineInterpreter(ctx context.Context) *interp.Interpreter {
	it := interp.New()
	it.Events = make(chan interp.PendingEvent, 64)
	h.startEventTranslator(ctx, it)
	// Reserved entities.
	it.Reserved["self"] = &selfView{host: h}
	it.Reserved["world"] = &worldView{host: h}
	it.Reserved["inventory"] = &inventoryView{host: h}
	it.Reserved["combat"] = &combatView{host: h}

	// Action registration.
	//
	// `registerAction`: a Result-returning callable. Auto-generates a
	// `<name>!` bang variant that aborts on err (per docs/lang/actions.md
	// "Which callables get bang variants").
	//
	// `registerPrimitive`: a non-Result callable (wait, note). No bang
	// variant — primitives can't fail in the typed sense, so `note!`
	// etc. are intentionally not registered. The validator rejects
	// bang-on-non-bang-eligible names at parse time (#67).
	registerAction := func(name string, fn func(context.Context, *Host, []interp.Value, map[string]interp.Value) (interp.Value, error)) {
		base := &actionCallable{name: name, host: h, ctx: ctx, fn: fn}
		it.Builtins[name] = base
		it.Builtins[name+"!"] = &interp.BangCallable{Underlying: base, Name: name + "!"}
	}
	registerPrimitive := func(name string, fn func(context.Context, *Host, []interp.Value, map[string]interp.Value) (interp.Value, error)) {
		it.Builtins[name] = &actionCallable{name: name, host: h, ctx: ctx, fn: fn}
	}

	// Primary actions — Result-returning, get bang variants.
	registerAction("walk_to", dslWalkTo)
	registerAction("attack", dslAttack)
	registerAction("talk_to", dslTalkTo)
	registerAction("answer", dslAnswer)
	registerAction("drop", dslDrop)
	registerAction("pick_up", dslPickUp)
	registerAction("eat", dslEat)
	registerAction("open_bank", dslOpenBank)
	registerAction("deposit", dslDeposit)
	registerAction("withdraw", dslWithdraw)
	registerAction("close_bank", dslCloseBank)
	registerAction("say", dslSay)
	registerAction("whisper", dslWhisper)
	registerAction("logout", dslLogout)

	// Primitives — no Result wrap, no bang variant.
	registerPrimitive("wait", dslWait)
	registerPrimitive("note", dslNote)

	// Action stubs (skills not yet wired to a real Host method).
	// Stubs return Fail(NOT_IMPLEMENTED, name) so DSL can branch
	// on err.code. They still get bang variants.
	for _, name := range []string{"mine", "fish", "chop", "cook", "cast"} {
		registerAction(name, makeStub(name))
	}

	// Stdlib stubs that CAN fail in a typed way — get bang variants.
	// Real LLM/memory bridge lands in Phase 4 / Phase 3 respectively.
	for _, name := range []string{
		"contemplate_reality", "evaluate", "decide", "exec", "improvise",
		"recall", "relation_with", "reflect_now",
		"wait_for_chat", "observe",
	} {
		registerAction(name, makeStub(name))
	}

	// Stdlib primitives (no failure mode — pure persona reads).
	// No bang variants.
	for _, name := range []string{"mood", "motivation", "wait_until"} {
		registerPrimitive(name, makeStub(name))
	}

	return it
}

// RunRoutine parses, validates, and executes a routine file from
// disk. `args` are the positional arguments to the routine's entry
// point.
func (h *Host) RunRoutine(ctx context.Context, path string, args []interp.Value) (interp.Result, error) {
	rf, err := ParseRoutineFile(path)
	if err != nil {
		return interp.Result{}, err
	}
	it := h.NewRoutineInterpreter(ctx)
	return it.RunRoutine(ctx, rf.File, args), nil
}
