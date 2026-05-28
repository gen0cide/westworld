package interp

import "fmt"

// ErrorCode enumerates the typed failure modes an action or stdlib
// call may return. Each code maps to a SCREAMING_SNAKE string via
// String() — that string is what DSL code sees on `result.err.code`.
//
// When adding a new code, also extend the errorCodeNames table
// below. Keep the iota order stable; routines might be (eventually,
// in mesa) keyed on these values.
//
// Wire convention: Go action wrappers MUST return an ErrorCode
// (via *Error), never an ad-hoc string. The validator and tooling
// rely on the typed enum to surface available failure modes
// (REPL .help, doc lookup, conformance assertions).
type ErrorCode int

const (
	// Movement
	PATH_BLOCKED ErrorCode = iota
	OUT_OF_RANGE

	// Inventory
	INVENTORY_FULL
	INVENTORY_EMPTY
	NO_SUCH_ITEM

	// Combat / interaction targets
	TARGET_DEAD
	TARGET_OUT_OF_VIEW

	// Session / connection
	NOT_LOGGED_IN
	INTERRUPTED

	// UI / sub-system state
	BANK_NOT_OPEN
	TRADE_NOT_ACTIVE
	DIALOG_NOT_OPEN

	// Misc
	ACTION_TIMEOUT
	SERVER_REJECTED // catch-all when the server says no with prose
	NOT_IMPLEMENTED // action wrapper registered as a stub

	numErrorCodes // sentinel; must stay last
)

// errorCodeNames is parallel to the iota above. Index = code value.
var errorCodeNames = [numErrorCodes]string{
	PATH_BLOCKED:       "PATH_BLOCKED",
	OUT_OF_RANGE:       "OUT_OF_RANGE",
	INVENTORY_FULL:     "INVENTORY_FULL",
	INVENTORY_EMPTY:    "INVENTORY_EMPTY",
	NO_SUCH_ITEM:       "NO_SUCH_ITEM",
	TARGET_DEAD:        "TARGET_DEAD",
	TARGET_OUT_OF_VIEW: "TARGET_OUT_OF_VIEW",
	NOT_LOGGED_IN:      "NOT_LOGGED_IN",
	INTERRUPTED:        "INTERRUPTED",
	BANK_NOT_OPEN:      "BANK_NOT_OPEN",
	TRADE_NOT_ACTIVE:   "TRADE_NOT_ACTIVE",
	DIALOG_NOT_OPEN:    "DIALOG_NOT_OPEN",
	ACTION_TIMEOUT:     "ACTION_TIMEOUT",
	SERVER_REJECTED:    "SERVER_REJECTED",
	NOT_IMPLEMENTED:    "NOT_IMPLEMENTED",
}

// String returns the SCREAMING_SNAKE name for this code, which is
// what DSL code compares against (`result.err.code == "PATH_BLOCKED"`).
func (e ErrorCode) String() string {
	if e < 0 || int(e) >= len(errorCodeNames) {
		return fmt.Sprintf("UNKNOWN_ERROR(%d)", int(e))
	}
	return errorCodeNames[e]
}

// Error is the DSL-visible structured error returned in
// `result.err` on action / stdlib failure. Routines branch on
// `.code` (string) and read `.reason` (human text) for detail.
//
// `.fatal` is true when the failure means the host cannot
// reasonably continue (lost connection, account banned, etc.).
// Bang-form callables convert any Error into an abortSignal
// regardless of fatal.
type Error struct {
	Code   ErrorCode
	Reason string
	Fatal  bool
}

// NewError is the canonical constructor used by action wrappers.
func NewError(code ErrorCode, reason string) *Error {
	return &Error{Code: code, Reason: reason}
}

// NewFatalError marks the failure as fatal (host should give up
// this routine and likely the connection).
func NewFatalError(code ErrorCode, reason string) *Error {
	return &Error{Code: code, Reason: reason, Fatal: true}
}

func (e *Error) Kind() string { return "error" }

func (e *Error) Display() string {
	if e.Reason == "" {
		return e.Code.String()
	}
	return e.Code.String() + ": " + e.Reason
}

// Get exposes .code / .reason / .fatal to DSL member access.
func (e *Error) Get(field string) (Value, bool) {
	switch field {
	case "code":
		return String(e.Code.String()), true
	case "reason":
		return String(e.Reason), true
	case "fatal":
		return Bool(e.Fatal), true
	}
	return nil, false
}

// CallResult is the value returned by any DSL callable that can
// fail in a typed way (actions, LLM stdlib, memory stdlib). DSL
// accesses `.val` and `.err`:
//
//	result = walk_to(x=120, y=504)
//	if result.err {                           # truthy iff error
//	    if result.err.code == "PATH_BLOCKED" { ... }
//	}
//
// On success Val holds the action's return value (Null for most
// actions; an item-view for pick_up, a string for
// contemplate_reality, etc.) and Err is nil.
//
// On failure Val is Null (always) and Err points at the typed
// Error.
//
// Naming note: the Go type is `CallResult` to avoid colliding
// with the routine-completion type `Result` used by RunRoutine.
// DSL code never sees the Go type name — it just reads `.val` /
// `.err` off whatever the callable returned.
type CallResult struct {
	Val Value  // never nil for success path; callers should set Null{} explicitly
	Err *Error // nil on success
}

// Ok is the canonical constructor for a successful CallResult.
// Pass Null{} if the callable returns no useful value.
func Ok(v Value) *CallResult {
	if v == nil {
		v = Null{}
	}
	return &CallResult{Val: v}
}

// Fail constructs a CallResult wrapping a typed error. `reason`
// is human-readable detail; pass "" if the code alone is
// sufficient.
func Fail(code ErrorCode, reason string) *CallResult {
	return &CallResult{Val: Null{}, Err: NewError(code, reason)}
}

// FailFatal is Fail but marks the error fatal — the host should
// give up the routine entirely.
func FailFatal(code ErrorCode, reason string) *CallResult {
	return &CallResult{Val: Null{}, Err: NewFatalError(code, reason)}
}

func (r *CallResult) Kind() string { return "result" }

func (r *CallResult) Display() string {
	if r.Err != nil {
		return "{err: " + r.Err.Display() + "}"
	}
	if r.Val == nil {
		return "{val: null}"
	}
	return "{val: " + r.Val.Display() + "}"
}

// Get exposes .val / .err for DSL member access.
func (r *CallResult) Get(field string) (Value, bool) {
	switch field {
	case "val":
		if r.Val == nil {
			return Null{}, true
		}
		return r.Val, true
	case "err":
		if r.Err == nil {
			return Null{}, true
		}
		return r.Err, true
	}
	return nil, false
}

// BangCallable wraps any *CallResult-returning Callable to give it
// the "assert success" semantics of the `!` variants:
//
//   - If the underlying returns a CallResult with non-nil Err →
//     abort the routine with the typed Error as the reason. (The
//     routine ends ResultAborted; try/recover binds the Error.)
//   - If the underlying returns a CallResult with nil Err →
//     unwrap and return the success value directly. So
//     `picked = pick_up!(item)` returns the picked-up item-view,
//     not a CallResult shell.
//   - If the underlying returns anything other than a CallResult
//     (defensive — shouldn't happen for properly-wrapped actions)
//     → pass through unchanged.
//
// BangCallable preserves the Yielder property of its underlying:
// bangs of yielding actions still yield to the event scheduler.
//
// The bridge auto-registers a BangCallable for every action /
// LLM stdlib / memory stdlib so DSL code can write `eat!`,
// `contemplate_reality!`, `recall!`, etc.
type BangCallable struct {
	Underlying Callable
	Name       string // "walk_to!" etc — used in traces + Display
}

func (b *BangCallable) Kind() string    { return "action!" }
func (b *BangCallable) Display() string { return "<" + b.Name + ">" }

// Yields delegates to the underlying so bangs of yielding actions
// remain yielders (the interpreter's event dispatcher fires
// before/after the call as expected).
func (b *BangCallable) Yields() bool {
	if y, ok := b.Underlying.(Yielder); ok {
		return y.Yields()
	}
	return false
}

func (b *BangCallable) Call(args []Value, named map[string]Value) (Value, error) {
	v, err := b.Underlying.Call(args, named)
	if err != nil {
		return nil, err
	}
	if cr, ok := v.(*CallResult); ok {
		if cr.Err != nil {
			// Panic with the typed abortSignal control-flow value.
			// The interpreter's execBody recover() catches this and
			// turns it into Result{Kind: ResultAborted, Value: Err}.
			// The DSL routine ends gracefully; the cradle process is
			// not affected.
			//
			// This panic is intentional control flow — same mechanism
			// used by the `abort` statement, `return` statement, and
			// `break`/`continue`. It NEVER reaches the cradle's main
			// goroutine because execBody (and only execBody) recovers
			// it.
			panic(abortSignal{Reason: cr.Err})
		}
		return cr.Val, nil
	}
	// Underlying didn't return a CallResult — pass through. Mostly
	// defensive; callable authors should never wrap non-Result
	// callables in BangCallable.
	return v, nil
}
