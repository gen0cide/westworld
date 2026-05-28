package interp

import (
	"fmt"
	"time"

	"github.com/gen0cide/westworld/dsl/token"
)

// Caps holds the per-routine resource budgets enforced during
// execution. The defaults match docs/dsl.md "Op budget", "Wall-clock
// budget", "Recursion depth", and "Memory caps". Tests override
// any field via Interpreter.Caps.
type Caps struct {
	// OpBudget is the maximum number of AST-node evaluations allowed
	// before the routine aborts with "op_budget_exceeded".
	// Default: 1,000,000.
	OpBudget int

	// WallClock is the maximum real time a routine instance may run
	// before aborting with "wall_clock_exceeded". Default: 4 hours.
	// Primary actions count against this (they're real-time bound).
	WallClock time.Duration

	// MaxRecursion caps the proc-call stack depth before aborting
	// with "recursion_depth_exceeded". Default: 64.
	MaxRecursion int

	// MaxListLen is the largest list a routine can construct
	// before "list_too_large" aborts the routine. Default: 1024.
	MaxListLen int

	// MaxStringLen caps the length of any String produced by
	// f-string interpolation or string concatenation. Default: 4096.
	MaxStringLen int
}

// DefaultCaps returns the docs/dsl.md-sourced budget defaults.
// Callers can keep most defaults and override one field.
func DefaultCaps() Caps {
	return Caps{
		OpBudget:     1_000_000,
		WallClock:    4 * time.Hour,
		MaxRecursion: 64,
		MaxListLen:   1024,
		MaxStringLen: 4096,
	}
}

// budget tracks the in-flight per-routine counters. Allocated once
// per RunRoutine and not exposed externally; lifecycle is owned by
// the interpreter.
type budget struct {
	caps           Caps
	opsRemaining   int
	deadline       time.Time
	recursionDepth int
}

func newBudget(c Caps) *budget {
	b := &budget{caps: c}
	if c.OpBudget <= 0 {
		b.opsRemaining = DefaultCaps().OpBudget
	} else {
		b.opsRemaining = c.OpBudget
	}
	wall := c.WallClock
	if wall <= 0 {
		wall = DefaultCaps().WallClock
	}
	b.deadline = time.Now().Add(wall)
	return b
}

// chargeOp consumes one op from the budget and returns nil on
// success or an abortSignal-shaped RuntimeError on exhaustion. The
// caller is expected to translate the error into an abort.
func (b *budget) chargeOp(pos token.Position) *RuntimeError {
	if b == nil {
		return nil
	}
	b.opsRemaining--
	if b.opsRemaining < 0 {
		return &RuntimeError{Pos: pos, Msg: "op_budget_exceeded"}
	}
	if time.Now().After(b.deadline) {
		return &RuntimeError{Pos: pos, Msg: "wall_clock_exceeded"}
	}
	return nil
}

// enterCall is invoked at the start of any proc / routine / handler
// dispatch. Returns an error if recursion would exceed the cap.
func (b *budget) enterCall(pos token.Position) *RuntimeError {
	if b == nil {
		return nil
	}
	b.recursionDepth++
	cap := b.caps.MaxRecursion
	if cap <= 0 {
		cap = DefaultCaps().MaxRecursion
	}
	if b.recursionDepth > cap {
		return &RuntimeError{Pos: pos, Msg: fmt.Sprintf("recursion_depth_exceeded (cap=%d)", cap)}
	}
	return nil
}

// exitCall pairs with enterCall in defer.
func (b *budget) exitCall() {
	if b == nil {
		return
	}
	if b.recursionDepth > 0 {
		b.recursionDepth--
	}
}

// checkListLen enforces MaxListLen at list construction time.
func (b *budget) checkListLen(pos token.Position, length int) *RuntimeError {
	if b == nil {
		return nil
	}
	cap := b.caps.MaxListLen
	if cap <= 0 {
		cap = DefaultCaps().MaxListLen
	}
	if length > cap {
		return &RuntimeError{Pos: pos, Msg: fmt.Sprintf("list_too_large (len=%d, cap=%d)", length, cap)}
	}
	return nil
}

// checkStringLen enforces MaxStringLen at string-build time.
func (b *budget) checkStringLen(pos token.Position, length int) *RuntimeError {
	if b == nil {
		return nil
	}
	cap := b.caps.MaxStringLen
	if cap <= 0 {
		cap = DefaultCaps().MaxStringLen
	}
	if length > cap {
		return &RuntimeError{Pos: pos, Msg: fmt.Sprintf("string_too_large (len=%d, cap=%d)", length, cap)}
	}
	return nil
}
