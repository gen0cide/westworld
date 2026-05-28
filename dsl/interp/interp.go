package interp

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/token"
)

// Interpreter walks a validated AST. One Interpreter per Host —
// it carries the host bridge (Builtins + Reserved entities), the RNG
// seed for deterministic replay, and the budgets enforced in step 7.
//
// Most fields default to safe stubs so the interpreter is testable
// without a full Host. The step-6 Host bridge installs real
// implementations of action calls.
type Interpreter struct {
	// Builtins resolves identifiers like walk_to, contemplate_reality,
	// note, etc. If nil, builtins panic. Step 6 wires the Host's
	// real implementations.
	Builtins map[string]Callable

	// Reserved supplies values for `self`, `world`, `inventory`,
	// `combat`. Implementations should expose Getter for member
	// access. If nil, the routine reports those names as null.
	Reserved map[string]Value

	// Rand is used for wait(a..b) jitter and any other randomness.
	// Tests inject a fixed seed for determinism.
	Rand *rand.Rand

	// Caps bounds resource usage per routine invocation (op budget,
	// wall clock, recursion depth, collection sizes). Zero-valued
	// fields fall back to DefaultCaps. See dsl.md "Op budget" and
	// adjacent sections.
	Caps Caps

	// Events is a non-blocking queue of incoming pending events
	// dispatched to `on` handlers at action boundaries. The Host
	// bridge populates this channel; tests may write directly.
	Events chan PendingEvent

	// OnHandlers is the indexed event-name → handlers map, built
	// from file.Handlers and routine.Handlers at RunRoutine startup.
	OnHandlers map[string][]*ast.OnHandler

	// budget is the per-routine tracker; nil until RunRoutine
	// initializes it.
	budget *budget

	// routineEnv is the routine's top-level env, captured for event
	// handler dispatch. Set during runDecl, cleared on return.
	routineEnv *Env
}

// New returns a default Interpreter with a fresh PRNG and default
// resource caps. Caller can override fields after construction.
func New() *Interpreter {
	return &Interpreter{
		Builtins: map[string]Callable{},
		Reserved: map[string]Value{},
		Rand:     rand.New(rand.NewSource(0xC0DE_BABE)),
		Caps:     DefaultCaps(),
	}
}

// ----- control-flow signals -----

// Control-flow is propagated up through Go panic with one of the
// sentinel types below. The top-level Run / call handler recovers
// them so they never leak.
type breakSignal struct{ Pos token.Position }
type continueSignal struct{ Pos token.Position }
type returnSignal struct {
	Value Value
	Pos   token.Position
}
type abortSignal struct {
	Reason Value
	Pos    token.Position
}

// RuntimeError is a non-control-flow failure — bad operand types,
// division by zero, unknown member, etc.
type RuntimeError struct {
	Pos token.Position
	Msg string
}

func (e *RuntimeError) Error() string { return fmt.Sprintf("%s: %s", e.Pos, e.Msg) }

func newError(pos token.Position, format string, args ...any) *RuntimeError {
	return &RuntimeError{Pos: pos, Msg: fmt.Sprintf(format, args...)}
}

// ----- result types -----

// ResultKind enumerates how a routine terminated.
type ResultKind int

const (
	ResultCompleted ResultKind = iota
	ResultReturned             // body executed `return value`
	ResultAborted              // body executed `abort reason`
	ResultErrored              // runtime panic / type error
	ResultCanceled             // context cancellation
)

func (k ResultKind) String() string {
	switch k {
	case ResultCompleted:
		return "completed"
	case ResultReturned:
		return "returned"
	case ResultAborted:
		return "aborted"
	case ResultErrored:
		return "errored"
	case ResultCanceled:
		return "canceled"
	}
	return "unknown"
}

// Result is the outcome of running a routine, proc, or handler.
type Result struct {
	Kind    ResultKind
	Value   Value         // for ResultReturned / ResultAborted
	Err     *RuntimeError // for ResultErrored
}

// ----- entry points -----

// RunRoutine executes a routine with the given args. Pure-Go
// callsite for tests; production Host integration uses the same
// path under step 6.
func (it *Interpreter) RunRoutine(ctx context.Context, file *ast.File, args []Value) (res Result) {
	if file == nil || file.Routine == nil {
		return Result{Kind: ResultErrored, Err: &RuntimeError{Msg: "no routine in file"}}
	}
	return it.runDecl(ctx, file, file.Routine, args)
}

func (it *Interpreter) runDecl(ctx context.Context, file *ast.File, r *ast.RoutineDecl, args []Value) (res Result) {
	// Fresh budget for each routine invocation. Pre-existing
	// interpreter state (Caps overrides) is honored; runtime
	// counters reset.
	it.budget = newBudget(it.Caps)
	defer func() { it.budget = nil }()

	it.registerHandlers(file, r)

	env := NewEnv()
	it.routineEnv = env
	defer func() { it.routineEnv = nil }()
	// Bind reserved entities (self / world / inventory / combat) at
	// the root scope so they're visible everywhere.
	for k, v := range it.Reserved {
		env.Define(k, v)
	}
	// Bind procs as callable values too — this lets the same
	// dispatcher handle both builtins and user procs.
	for _, p := range file.Procs {
		env.Define(p.Name, &procCallable{interp: it, file: file, proc: p})
	}
	// Bind params with optional defaults.
	if err := it.bindParams(env, r.Params, args, r.Position); err != nil {
		return Result{Kind: ResultErrored, Err: err}
	}

	// Evaluate require preconditions. A falsey one short-circuits to
	// "aborted: precondition_failed".
	if r.Require != nil {
		for _, cond := range r.Require.Conds {
			v, err := it.evalSafe(ctx, cond, env)
			if err != nil {
				return Result{Kind: ResultErrored, Err: err}
			}
			if !Truthy(v) {
				return Result{
					Kind:  ResultAborted,
					Value: String("precondition_failed"),
				}
			}
		}
	}

	return it.execBody(ctx, r.Body, env)
}

// execBody runs a block and converts any control-flow signal into a
// Result. Wraps the block in defer/recover for signals + errors.
func (it *Interpreter) execBody(ctx context.Context, body *ast.Block, env *Env) (res Result) {
	defer func() {
		switch s := recover().(type) {
		case nil:
			// completed normally
			if res.Kind == 0 { // ResultCompleted is zero value
				res = Result{Kind: ResultCompleted}
			}
		case returnSignal:
			res = Result{Kind: ResultReturned, Value: s.Value}
		case abortSignal:
			res = Result{Kind: ResultAborted, Value: s.Reason}
		case *RuntimeError:
			res = Result{Kind: ResultErrored, Err: s}
		case breakSignal:
			res = Result{Kind: ResultErrored, Err: newError(s.Pos, "break outside of loop reached interpreter top level")}
		case continueSignal:
			res = Result{Kind: ResultErrored, Err: newError(s.Pos, "continue outside of loop reached interpreter top level")}
		default:
			// Re-panic anything we didn't generate ourselves.
			panic(s)
		}
	}()
	if err := ctx.Err(); err != nil {
		return Result{Kind: ResultCanceled}
	}
	it.execBlock(ctx, body, env)
	return Result{Kind: ResultCompleted}
}

func (it *Interpreter) bindParams(env *Env, params []*ast.Param, args []Value, declPos token.Position) *RuntimeError {
	for i, p := range params {
		if i < len(args) {
			env.Define(p.Name, args[i])
			continue
		}
		if p.Default == nil {
			return newError(declPos, "missing required argument %q", p.Name)
		}
		v, err := it.evalSafe(context.Background(), p.Default, env)
		if err != nil {
			return err
		}
		env.Define(p.Name, v)
	}
	return nil
}

// ----- statement execution -----

func (it *Interpreter) execBlock(ctx context.Context, b *ast.Block, env *Env) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		if err := ctx.Err(); err != nil {
			panic(returnSignal{Value: String("canceled"), Pos: s.Pos()})
		}
		it.execStmt(ctx, s, env)
	}
}

func (it *Interpreter) execStmt(ctx context.Context, s ast.Stmt, env *Env) {
	if err := it.budget.chargeOp(s.Pos()); err != nil {
		panic(abortSignal{Reason: String(err.Msg), Pos: s.Pos()})
	}
	switch n := s.(type) {
	case *ast.Block:
		it.execBlock(ctx, n, env.Child())
	case *ast.AssignStmt:
		v := it.eval(ctx, n.Value, env)
		env.Set(n.Target, v)
	case *ast.ExprStmt:
		it.eval(ctx, n.X, env)
	case *ast.IfStmt:
		it.execIf(ctx, n, env)
	case *ast.WhileStmt:
		it.execWhile(ctx, n, env)
	case *ast.ForStmt:
		it.execFor(ctx, n, env)
	case *ast.BreakStmt:
		panic(breakSignal{Pos: n.Position})
	case *ast.ContinueStmt:
		panic(continueSignal{Pos: n.Position})
	case *ast.ReturnStmt:
		var v Value = Null{}
		if n.Value != nil {
			v = it.eval(ctx, n.Value, env)
		}
		panic(returnSignal{Value: v, Pos: n.Position})
	case *ast.AbortStmt:
		reason := it.eval(ctx, n.Reason, env)
		panic(abortSignal{Reason: reason, Pos: n.Position})
	case *ast.WaitStmt:
		it.execWait(ctx, n, env)
	case *ast.RequireBlock:
		// Already hoisted by the validator; if one reaches here it's
		// a no-op (or a validator bug — fail loud).
		panic(newError(n.Position, "stray require block reached interpreter (validator should have hoisted it)"))
	default:
		panic(newError(s.Pos(), "unhandled statement %T", s))
	}
}

func (it *Interpreter) execIf(ctx context.Context, n *ast.IfStmt, env *Env) {
	if Truthy(it.eval(ctx, n.Cond, env)) {
		it.execBlock(ctx, n.Then, env.Child())
		return
	}
	for _, ec := range n.Elifs {
		if Truthy(it.eval(ctx, ec.Cond, env)) {
			it.execBlock(ctx, ec.Body, env.Child())
			return
		}
	}
	if n.Else != nil {
		it.execBlock(ctx, n.Else, env.Child())
	}
}

func (it *Interpreter) execWhile(ctx context.Context, n *ast.WhileStmt, env *Env) {
	for {
		if err := ctx.Err(); err != nil {
			return
		}
		if !Truthy(it.eval(ctx, n.Cond, env)) {
			return
		}
		if !it.execLoopBody(ctx, n.Body, env) {
			return
		}
	}
}

func (it *Interpreter) execFor(ctx context.Context, n *ast.ForStmt, env *Env) {
	iter := it.eval(ctx, n.Iter, env)
	switch x := iter.(type) {
	case *List:
		for _, item := range x.Items {
			if err := ctx.Err(); err != nil {
				return
			}
			loopEnv := env.Child()
			loopEnv.Define(n.Var, item)
			if !it.execLoopBody(ctx, n.Body, loopEnv) {
				return
			}
		}
	case *Range:
		lo, hi, isInt := toInclusiveIntRange(x)
		if !isInt {
			panic(newError(n.Position, "for-in over float range is not supported"))
		}
		for v := lo; v <= hi; v++ {
			if err := ctx.Err(); err != nil {
				return
			}
			loopEnv := env.Child()
			loopEnv.Define(n.Var, Int(v))
			if !it.execLoopBody(ctx, n.Body, loopEnv) {
				return
			}
		}
	case String:
		for _, r := range string(x) {
			if err := ctx.Err(); err != nil {
				return
			}
			loopEnv := env.Child()
			loopEnv.Define(n.Var, String(string(r)))
			if !it.execLoopBody(ctx, n.Body, loopEnv) {
				return
			}
		}
	default:
		panic(newError(n.Position, "for-in over %s is not iterable", iter.Kind()))
	}
}

// execLoopBody runs body and translates break/continue. Returns true
// to continue the loop, false to break out. Any other signal
// (returnSignal, abortSignal, RuntimeError) is re-panicked so the
// outer execBody handles it.
func (it *Interpreter) execLoopBody(ctx context.Context, body *ast.Block, env *Env) (keepGoing bool) {
	keepGoing = true
	defer func() {
		// CRITICAL: recover() must only be called once. Calling it
		// again returns nil and turns a real signal into a nil
		// panic. (Old code re-called recover() in the default arm
		// and lost abortSignals from within loop bodies.)
		r := recover()
		switch r.(type) {
		case nil:
		case breakSignal:
			keepGoing = false
		case continueSignal:
			keepGoing = true
		default:
			panic(r)
		}
	}()
	it.execBlock(ctx, body, env)
	return true
}

// toInclusiveIntRange converts a *Range to inclusive int64 bounds if
// both endpoints are integers. Returns ok=false otherwise.
func toInclusiveIntRange(r *Range) (lo, hi int64, ok bool) {
	li, lok := AsInt(r.Low)
	hi64, hok := AsInt(r.High)
	if !lok || !hok {
		return 0, 0, false
	}
	return li, hi64, true
}

func (it *Interpreter) execWait(ctx context.Context, n *ast.WaitStmt, env *Env) {
	dur := it.eval(ctx, n.Duration, env)
	secs, err := durationToSeconds(it.Rand, dur, n.Position)
	if err != nil {
		panic(err)
	}
	// The interpreter doesn't actually sleep — that's a Host
	// concern. We invoke the "wait" builtin if registered, otherwise
	// no-op (tests can override). This keeps the eval pure.
	if cb, ok := it.Builtins["wait"]; ok {
		_, callErr := cb.Call([]Value{Float(secs)}, nil)
		if callErr != nil {
			panic(newError(n.Position, "wait: %v", callErr))
		}
	}
}

func durationToSeconds(rng *rand.Rand, v Value, pos token.Position) (float64, *RuntimeError) {
	if rng == nil {
		rng = rand.New(rand.NewSource(0))
	}
	if r, ok := v.(*Range); ok {
		lo, lok := AsFloat(r.Low)
		hi, hok := AsFloat(r.High)
		if !lok || !hok {
			return 0, newError(pos, "wait range bounds must be numeric")
		}
		if hi < lo {
			lo, hi = hi, lo
		}
		return lo + rng.Float64()*(hi-lo), nil
	}
	f, ok := AsFloat(v)
	if !ok {
		return 0, newError(pos, "wait expects a number or range, got %s", v.Kind())
	}
	return f, nil
}

// ----- expression evaluation -----

func (it *Interpreter) eval(ctx context.Context, e ast.Expr, env *Env) Value {
	v, err := it.evalSafe(ctx, e, env)
	if err != nil {
		panic(err)
	}
	return v
}

func (it *Interpreter) evalSafe(ctx context.Context, e ast.Expr, env *Env) (Value, *RuntimeError) {
	if e == nil {
		return Null{}, nil
	}
	if err := it.budget.chargeOp(e.Pos()); err != nil {
		// Budget exhaustion needs to abort, not just error — convert
		// via panic so the routine ends cleanly with ResultAborted.
		panic(abortSignal{Reason: String(err.Msg), Pos: e.Pos()})
	}
	switch n := e.(type) {
	case *ast.IntLit:
		return Int(n.Value), nil
	case *ast.FloatLit:
		return Float(n.Value), nil
	case *ast.StringLit:
		return String(n.Value), nil
	case *ast.BoolLit:
		return Bool(n.Value), nil
	case *ast.NullLit:
		return Null{}, nil
	case *ast.FStringLit:
		return it.evalFString(ctx, n, env)
	case *ast.ListLit:
		if err := it.budget.checkListLen(n.Position, len(n.Elems)); err != nil {
			panic(abortSignal{Reason: String(err.Msg), Pos: n.Position})
		}
		out := make([]Value, len(n.Elems))
		for i, el := range n.Elems {
			v, err := it.evalSafe(ctx, el, env)
			if err != nil {
				return nil, err
			}
			out[i] = v
		}
		return &List{Items: out}, nil
	case *ast.RangeLit:
		lo, err := it.evalSafe(ctx, n.Low, env)
		if err != nil {
			return nil, err
		}
		hi, err := it.evalSafe(ctx, n.High, env)
		if err != nil {
			return nil, err
		}
		return &Range{Low: lo, High: hi}, nil
	case *ast.Ident:
		return it.evalIdent(n, env)
	case *ast.BinaryExpr:
		return it.evalBinary(ctx, n, env)
	case *ast.UnaryExpr:
		return it.evalUnary(ctx, n, env)
	case *ast.MemberExpr:
		return it.evalMember(ctx, n, env)
	case *ast.IndexExpr:
		return it.evalIndex(ctx, n, env)
	case *ast.CallExpr:
		return it.evalCall(ctx, n, env)
	}
	return nil, newError(e.Pos(), "unhandled expression %T", e)
}

func (it *Interpreter) evalIdent(n *ast.Ident, env *Env) (Value, *RuntimeError) {
	if v, ok := env.Get(n.Name); ok {
		return v, nil
	}
	if cb, ok := it.Builtins[n.Name]; ok {
		return &builtinValue{name: n.Name, fn: cb}, nil
	}
	// Validator should have caught this, but be defensive.
	return nil, newError(n.Position, "unbound identifier %q", n.Name)
}

func (it *Interpreter) evalFString(ctx context.Context, n *ast.FStringLit, env *Env) (Value, *RuntimeError) {
	var sb strings.Builder
	for _, p := range n.Parts {
		v, err := it.evalSafe(ctx, p, env)
		if err != nil {
			return nil, err
		}
		sb.WriteString(v.Display())
		if err := it.budget.checkStringLen(n.Position, sb.Len()); err != nil {
			panic(abortSignal{Reason: String(err.Msg), Pos: n.Position})
		}
	}
	return String(sb.String()), nil
}

func (it *Interpreter) evalBinary(ctx context.Context, n *ast.BinaryExpr, env *Env) (Value, *RuntimeError) {
	// Short-circuit logical ops.
	if n.Op == token.AND {
		lhs, err := it.evalSafe(ctx, n.Lhs, env)
		if err != nil {
			return nil, err
		}
		if !Truthy(lhs) {
			return lhs, nil
		}
		return it.evalSafe(ctx, n.Rhs, env)
	}
	if n.Op == token.OR {
		lhs, err := it.evalSafe(ctx, n.Lhs, env)
		if err != nil {
			return nil, err
		}
		if Truthy(lhs) {
			return lhs, nil
		}
		return it.evalSafe(ctx, n.Rhs, env)
	}
	lhs, err := it.evalSafe(ctx, n.Lhs, env)
	if err != nil {
		return nil, err
	}
	rhs, err := it.evalSafe(ctx, n.Rhs, env)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case token.EQ:
		return Bool(Equal(lhs, rhs)), nil
	case token.NEQ:
		return Bool(!Equal(lhs, rhs)), nil
	case token.LT, token.LTE, token.GT, token.GTE:
		return it.cmpNumeric(n, lhs, rhs)
	case token.ADD:
		return it.addValues(n, lhs, rhs)
	case token.SUB, token.MUL, token.QUO, token.REM:
		return it.arithNumeric(n, lhs, rhs)
	}
	return nil, newError(n.Position, "unsupported binary op %s", n.Op)
}

func (it *Interpreter) cmpNumeric(n *ast.BinaryExpr, a, b Value) (Value, *RuntimeError) {
	af, aok := AsFloat(a)
	bf, bok := AsFloat(b)
	if !aok || !bok {
		return nil, newError(n.Position, "cannot compare %s with %s", a.Kind(), b.Kind())
	}
	switch n.Op {
	case token.LT:
		return Bool(af < bf), nil
	case token.LTE:
		return Bool(af <= bf), nil
	case token.GT:
		return Bool(af > bf), nil
	case token.GTE:
		return Bool(af >= bf), nil
	}
	return nil, newError(n.Position, "unhandled comparison %s", n.Op)
}

// addValues handles +. If both operands are strings, concatenates;
// otherwise treats as numeric and delegates.
func (it *Interpreter) addValues(n *ast.BinaryExpr, a, b Value) (Value, *RuntimeError) {
	if as, ok := a.(String); ok {
		bs, ok := b.(String)
		if !ok {
			return nil, newError(n.Position, "cannot concatenate string with %s", b.Kind())
		}
		merged := string(as) + string(bs)
		if err := it.budget.checkStringLen(n.Position, len(merged)); err != nil {
			panic(abortSignal{Reason: String(err.Msg), Pos: n.Position})
		}
		return String(merged), nil
	}
	return it.arithNumeric(n, a, b)
}

func (it *Interpreter) arithNumeric(n *ast.BinaryExpr, a, b Value) (Value, *RuntimeError) {
	// If both are Int, stay in Int (except division which always
	// returns float to avoid Python-2-style integer-truncation bugs).
	ai, aIsInt := a.(Int)
	bi, bIsInt := b.(Int)
	if aIsInt && bIsInt && n.Op != token.QUO {
		switch n.Op {
		case token.ADD:
			return Int(int64(ai) + int64(bi)), nil
		case token.SUB:
			return Int(int64(ai) - int64(bi)), nil
		case token.MUL:
			return Int(int64(ai) * int64(bi)), nil
		case token.REM:
			if int64(bi) == 0 {
				return nil, newError(n.Position, "modulo by zero")
			}
			return Int(int64(ai) % int64(bi)), nil
		}
	}
	af, aok := AsFloat(a)
	bf, bok := AsFloat(b)
	if !aok || !bok {
		return nil, newError(n.Position, "non-numeric operand to %s (%s, %s)", n.Op, a.Kind(), b.Kind())
	}
	switch n.Op {
	case token.ADD:
		return Float(af + bf), nil
	case token.SUB:
		return Float(af - bf), nil
	case token.MUL:
		return Float(af * bf), nil
	case token.QUO:
		if bf == 0 {
			return nil, newError(n.Position, "division by zero")
		}
		return Float(af / bf), nil
	case token.REM:
		if bf == 0 {
			return nil, newError(n.Position, "modulo by zero")
		}
		return Float(math.Mod(af, bf)), nil
	}
	return nil, newError(n.Position, "unsupported numeric op %s", n.Op)
}

func (it *Interpreter) evalUnary(ctx context.Context, n *ast.UnaryExpr, env *Env) (Value, *RuntimeError) {
	v, err := it.evalSafe(ctx, n.Rhs, env)
	if err != nil {
		return nil, err
	}
	switch n.Op {
	case token.NOT:
		return Bool(!Truthy(v)), nil
	case token.SUB:
		if i, ok := v.(Int); ok {
			return Int(-int64(i)), nil
		}
		if f, ok := v.(Float); ok {
			return Float(-float64(f)), nil
		}
		return nil, newError(n.Position, "unary - on non-numeric %s", v.Kind())
	case token.ADD:
		return v, nil
	}
	return nil, newError(n.Position, "unsupported unary op %s", n.Op)
}

func (it *Interpreter) evalMember(ctx context.Context, n *ast.MemberExpr, env *Env) (Value, *RuntimeError) {
	recv, err := it.evalSafe(ctx, n.Recv, env)
	if err != nil {
		return nil, err
	}
	// Collections have a couple of well-known fields.
	if list, ok := recv.(*List); ok {
		switch n.Field {
		case "length":
			return Int(int64(len(list.Items))), nil
		}
		return nil, newError(n.Position, "list has no field %q", n.Field)
	}
	if g, ok := recv.(Getter); ok {
		if v, ok := g.Get(n.Field); ok {
			return v, nil
		}
		return nil, newError(n.Position, "%s has no field %q", recv.Kind(), n.Field)
	}
	if s, ok := recv.(String); ok {
		if n.Field == "length" {
			return Int(int64(len(string(s)))), nil
		}
	}
	return nil, newError(n.Position, "%s does not support field access", recv.Kind())
}

func (it *Interpreter) evalIndex(ctx context.Context, n *ast.IndexExpr, env *Env) (Value, *RuntimeError) {
	recv, err := it.evalSafe(ctx, n.Recv, env)
	if err != nil {
		return nil, err
	}
	idx, err := it.evalSafe(ctx, n.Index, env)
	if err != nil {
		return nil, err
	}
	switch x := recv.(type) {
	case *List:
		i, ok := AsInt(idx)
		if !ok {
			return nil, newError(n.Position, "list index must be int, got %s", idx.Kind())
		}
		if i < 0 || int(i) >= len(x.Items) {
			return nil, newError(n.Position, "list index %d out of bounds (length %d)", i, len(x.Items))
		}
		return x.Items[i], nil
	case String:
		i, ok := AsInt(idx)
		if !ok {
			return nil, newError(n.Position, "string index must be int, got %s", idx.Kind())
		}
		s := string(x)
		if i < 0 || int(i) >= len(s) {
			return nil, newError(n.Position, "string index %d out of bounds (length %d)", i, len(s))
		}
		return String(string(s[i])), nil
	}
	if ix, ok := recv.(Indexer); ok {
		if v, ok := ix.Index(idx); ok {
			return v, nil
		}
		return nil, newError(n.Position, "%s index %s not found", recv.Kind(), idx.Display())
	}
	return nil, newError(n.Position, "%s does not support indexing", recv.Kind())
}

// evalCall handles three callable shapes:
//   1. ident(...) — proc / builtin (resolved via env or Builtins)
//   2. recv.method(...) — Getter returns a Callable
//   3. recv[idx](...) — same idea via Indexer
func (it *Interpreter) evalCall(ctx context.Context, n *ast.CallExpr, env *Env) (Value, *RuntimeError) {
	callee, err := it.evalCallee(ctx, n.Callee, env)
	if err != nil {
		return nil, err
	}
	pos := pos(n.Callee)
	args, named, err := it.evalArgs(ctx, n.Args, env)
	if err != nil {
		return nil, err
	}
	// Drain the event queue around yielding callables (actions).
	// Handlers run synchronously in the routine's env so locals
	// stay visible. Per dsl.md: handlers run BETWEEN actions —
	// we drain both before AND after the call so a handler-
	// triggering action (e.g., one that publishes its own event)
	// gets dispatched immediately.
	yields := false
	if y, ok := callee.(Yielder); ok {
		yields = y.Yields()
	}
	if yields && it.routineEnv != nil {
		it.dispatchPendingEvents(ctx, it.routineEnv)
	}
	v, callErr := callee.Call(args, named)
	if yields && it.routineEnv != nil {
		it.dispatchPendingEvents(ctx, it.routineEnv)
	}
	if callErr != nil {
		if re, ok := callErr.(*RuntimeError); ok {
			return nil, re
		}
		return nil, newError(pos, "%v", callErr)
	}
	if v == nil {
		return Null{}, nil
	}
	return v, nil
}

func (it *Interpreter) evalCallee(ctx context.Context, callee ast.Expr, env *Env) (Callable, *RuntimeError) {
	v, err := it.evalSafe(ctx, callee, env)
	if err != nil {
		return nil, err
	}
	c, ok := v.(Callable)
	if !ok {
		return nil, newError(pos(callee), "%s is not callable", v.Kind())
	}
	return c, nil
}

func (it *Interpreter) evalArgs(ctx context.Context, args []*ast.Arg, env *Env) ([]Value, map[string]Value, *RuntimeError) {
	var positional []Value
	var named map[string]Value
	for _, a := range args {
		v, err := it.evalSafe(ctx, a.Value, env)
		if err != nil {
			return nil, nil, err
		}
		if a.Name != "" {
			if named == nil {
				named = map[string]Value{}
			}
			named[a.Name] = v
			continue
		}
		positional = append(positional, v)
	}
	return positional, named, nil
}

func pos(e ast.Expr) token.Position {
	if e == nil {
		return token.Position{}
	}
	return e.Pos()
}

// ----- callable wrappers -----

// builtinValue wraps a Callable so it can flow through the
// interpreter as a first-class value (e.g., assigned to a local).
type builtinValue struct {
	name string
	fn   Callable
}

func (b *builtinValue) Kind() string    { return "builtin" }
func (b *builtinValue) Display() string { return "<builtin " + b.name + ">" }
func (b *builtinValue) Call(args []Value, named map[string]Value) (Value, error) {
	return b.fn.Call(args, named)
}

// Yields delegates to the wrapped Callable so action wrappers
// (which implement Yielder) keep their yielding semantics after
// being looked up through env.Get → builtinValue.
func (b *builtinValue) Yields() bool {
	if y, ok := b.fn.(Yielder); ok {
		return y.Yields()
	}
	return false
}

// procCallable wraps a parsed ProcDecl so the interpreter can call
// it by name through the standard Callable interface.
type procCallable struct {
	interp *Interpreter
	file   *ast.File
	proc   *ast.ProcDecl
}

func (p *procCallable) Kind() string    { return "proc" }
func (p *procCallable) Display() string { return "<proc " + p.proc.Name + ">" }
func (p *procCallable) Call(args []Value, _ map[string]Value) (Value, error) {
	if err := p.interp.budget.enterCall(p.proc.Position); err != nil {
		return nil, err
	}
	defer p.interp.budget.exitCall()

	env := NewEnv()
	for k, v := range p.interp.Reserved {
		env.Define(k, v)
	}
	for _, pp := range p.file.Procs {
		env.Define(pp.Name, &procCallable{interp: p.interp, file: p.file, proc: pp})
	}
	if err := p.interp.bindParams(env, p.proc.Params, args, p.proc.Position); err != nil {
		return nil, err
	}
	res := p.interp.execBody(context.Background(), p.proc.Body, env)
	switch res.Kind {
	case ResultCompleted:
		return Null{}, nil
	case ResultReturned:
		return res.Value, nil
	case ResultErrored:
		return nil, res.Err
	case ResultAborted:
		// Abort inside a proc propagates as an abort signal through
		// the calling routine — convert it back to a panic so the
		// caller's execBody recover() turns it into ResultAborted at
		// the routine level. (procs can't abort the routine
		// themselves per dsl.md, but proc panics shouldn't leak
		// either.)
		panic(abortSignal{Reason: res.Value, Pos: p.proc.Position})
	default:
		return nil, fmt.Errorf("proc %s ended with unexpected kind %s", p.proc.Name, res.Kind)
	}
}
