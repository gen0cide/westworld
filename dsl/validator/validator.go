// Package validator performs structural and name-resolution checks on
// a parsed .routine AST. See docs/dsl.md "Static validation
// (parse-time)".
//
// Checks performed:
//
//   - Identifier resolution: every Ident references a local, parameter,
//     reserved name (self / world / inventory / combat), declared proc,
//     or builtin (action / primitive / stdlib function).
//   - `require` hoisting: a single leading RequireBlock in a routine
//     body moves to RoutineDecl.Require. RequireBlocks elsewhere are
//     errors.
//   - Action calls forbidden inside `proc` bodies and `require` blocks.
//   - `break` / `continue` only inside a loop.
//   - `return` only inside a routine or proc.
//   - `abort` only inside a routine.
//   - Event handler names + arity match the v1 event table.
//   - Duplicate proc names within a file are rejected.
//   - Reserved names (self / world / inventory / combat) cannot be
//     assigned or used as a parameter.
//
// Not yet checked (deferred to later steps):
//
//   - Loop yield-point requirement (step 5+, needs flow analysis).
//   - Op-budget upper bound from AST size (step 7).
//   - Reachability — "every path returns/aborts/reaches an action"
//     (step 5).
//
// Validation errors collect with source position; nothing panics on
// malformed input.
package validator

import (
	"fmt"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/token"
)

// Validate runs all static checks on `file`. The file may be modified
// in place — currently the only mutation is hoisting a leading
// `require` from RoutineDecl.Body to RoutineDecl.Require.
//
// Returns nil if validation passes. Otherwise returns an error whose
// message lists every diagnostic, newline-separated. Use Errors() on
// the returned *MultiError to iterate.
func Validate(file *ast.File) error {
	v := &Validator{file: file, procs: map[string]*ast.ProcDecl{}}
	v.run()
	if len(v.errs) == 0 {
		return nil
	}
	return &MultiError{Errs: v.errs}
}

// MultiError aggregates every validation diagnostic produced for a
// single file. The strategist consumes Errs to retry generation with
// the full diagnostic list in context.
type MultiError struct{ Errs []error }

func (m *MultiError) Error() string {
	if len(m.Errs) == 1 {
		return m.Errs[0].Error()
	}
	out := fmt.Sprintf("%d validation errors:", len(m.Errs))
	for _, e := range m.Errs {
		out += "\n  " + e.Error()
	}
	return out
}

// Validator carries scope and diagnostic state through one file.
type Validator struct {
	file  *ast.File
	errs  []error
	procs map[string]*ast.ProcDecl
}

func (v *Validator) errorf(pos token.Position, format string, args ...any) {
	v.errs = append(v.errs, fmt.Errorf("%s: "+format, append([]any{pos}, args...)...))
}

func (v *Validator) run() {
	// First pass: index proc declarations so identifier resolution
	// in subsequent passes can find them.
	for _, p := range v.file.Procs {
		if isReservedName(p.Name) {
			v.errorf(p.Position, "proc name %q shadows a reserved variable", p.Name)
		}
		if _, dup := v.procs[p.Name]; dup {
			v.errorf(p.Position, "duplicate proc %q", p.Name)
			continue
		}
		v.procs[p.Name] = p
	}

	// Validate every top-level declaration.
	for _, p := range v.file.Procs {
		v.validateProc(p)
	}
	for _, h := range v.file.Handlers {
		v.validateHandler(h, nil)
	}
	if v.file.Routine != nil {
		v.validateRoutine(v.file.Routine)
	}
}

// ----- declaration-level validation -----

func (v *Validator) validateProc(p *ast.ProcDecl) {
	scope := newScope(nil)
	v.bindParams(p.Params, scope)
	ctx := &context{scope: scope, inProc: true}
	v.checkBlock(p.Body, ctx)
}

func (v *Validator) validateHandler(h *ast.OnHandler, routineScope *scope) {
	want, ok := eventArity[h.Event]
	if !ok {
		v.errorf(h.Position, "unknown event %q (see docs/dsl.md handler table)", h.Event)
	} else if len(h.Params) != want {
		v.errorf(h.Position, "event %q expects %d param(s), got %d", h.Event, want, len(h.Params))
	}
	scope := newScope(routineScope)
	v.bindParams(h.Params, scope)
	// Handlers run in the context of a routine; per dsl.md they may
	// `abort` the parent routine ("on death { abort 'died' }").
	// inRoutine=true so the abort check in checkStmt accepts that.
	ctx := &context{scope: scope, inHandler: true, inRoutine: true}
	v.checkBlock(h.Body, ctx)
}

func (v *Validator) validateRoutine(r *ast.RoutineDecl) {
	scope := newScope(nil)
	v.bindParams(r.Params, scope)

	// Hoist a leading require-block from r.Body.Stmts to r.Require.
	// The parser flowed `require { ... }` through as a Stmt
	// (RequireBlock implements Stmt). At most one require is allowed,
	// and only as the first statement.
	if r.Body != nil && len(r.Body.Stmts) > 0 {
		if rb, ok := r.Body.Stmts[0].(*ast.RequireBlock); ok {
			r.Require = rb
			r.Body.Stmts = r.Body.Stmts[1:]
		}
	}
	// Any remaining RequireBlock in the body is an error.
	if r.Body != nil {
		for _, s := range r.Body.Stmts {
			if rb, ok := s.(*ast.RequireBlock); ok {
				v.errorf(rb.Position, "require block must be the first statement in a routine body")
			}
		}
	}

	// Validate require conds — must be pure (no actions, no waits).
	if r.Require != nil {
		ctx := &context{scope: scope, inRequire: true, inRoutine: true}
		for _, c := range r.Require.Conds {
			v.checkExpr(c, ctx)
		}
	}

	// Validate routine-scoped handlers (per dsl.md two-tier model).
	for _, h := range r.Handlers {
		v.validateHandler(h, scope)
	}

	ctx := &context{scope: scope, inRoutine: true}
	v.checkBlock(r.Body, ctx)
}

func (v *Validator) bindParams(params []*ast.Param, s *scope) {
	for _, p := range params {
		if isReservedName(p.Name) {
			v.errorf(p.Position, "parameter %q shadows a reserved variable", p.Name)
			continue
		}
		if s.has(p.Name) {
			v.errorf(p.Position, "duplicate parameter %q", p.Name)
			continue
		}
		s.bind(p.Name)
		// Default value expressions must also resolve identifiers.
		if p.Default != nil {
			v.checkExpr(p.Default, &context{scope: s})
		}
	}
}

// ----- context + scope -----

// scope is a chained set of bound identifier names (locals + params).
// We don't track types — name resolution is the only check.
type scope struct {
	parent *scope
	names  map[string]bool
}

func newScope(parent *scope) *scope {
	return &scope{parent: parent, names: map[string]bool{}}
}

func (s *scope) bind(name string)     { s.names[name] = true }
func (s *scope) has(name string) bool { return s != nil && s.names[name] }
func (s *scope) resolve(name string) bool {
	for c := s; c != nil; c = c.parent {
		if c.has(name) {
			return true
		}
	}
	return false
}

// context describes where in the AST we currently are — controls
// which statement/expression forms are legal here.
type context struct {
	scope     *scope
	inRoutine bool
	inProc    bool
	inHandler bool
	inLoop    bool
	inRequire bool
}

func (c *context) child(scopeOverride *scope) *context {
	out := *c
	if scopeOverride != nil {
		out.scope = scopeOverride
	}
	return &out
}

// ----- block + statement validation -----

func (v *Validator) checkBlock(b *ast.Block, ctx *context) {
	if b == nil {
		return
	}
	// Nested block introduces a fresh child scope so locals don't
	// leak upward — but for now, the DSL doesn't have explicit
	// `let`, so any AssignStmt simply binds into the current scope.
	// We'll revisit if `let`-style scoping is added.
	for _, s := range b.Stmts {
		v.checkStmt(s, ctx)
	}
}

func (v *Validator) checkStmt(s ast.Stmt, ctx *context) {
	switch n := s.(type) {
	case *ast.Block:
		v.checkBlock(n, ctx)
	case *ast.AssignStmt:
		if isReservedName(n.Target) {
			v.errorf(n.Position, "cannot assign to reserved name %q", n.Target)
		}
		v.checkExpr(n.Value, ctx)
		ctx.scope.bind(n.Target)
	case *ast.ExprStmt:
		v.checkExpr(n.X, ctx)
	case *ast.IfStmt:
		v.checkExpr(n.Cond, ctx)
		v.checkBlock(n.Then, ctx)
		for _, ec := range n.Elifs {
			v.checkExpr(ec.Cond, ctx)
			v.checkBlock(ec.Body, ctx)
		}
		v.checkBlock(n.Else, ctx)
	case *ast.WhileStmt:
		v.checkExpr(n.Cond, ctx)
		loopCtx := *ctx
		loopCtx.inLoop = true
		v.checkBlock(n.Body, &loopCtx)
	case *ast.ForStmt:
		v.checkExpr(n.Iter, ctx)
		if isReservedName(n.Var) {
			v.errorf(n.Position, "for-loop variable %q shadows a reserved name", n.Var)
		}
		loopScope := newScope(ctx.scope)
		loopScope.bind(n.Var)
		loopCtx := *ctx
		loopCtx.inLoop = true
		loopCtx.scope = loopScope
		v.checkBlock(n.Body, &loopCtx)
	case *ast.BreakStmt:
		if !ctx.inLoop {
			v.errorf(n.Position, "break outside of loop")
		}
	case *ast.ContinueStmt:
		if !ctx.inLoop {
			v.errorf(n.Position, "continue outside of loop")
		}
	case *ast.ReturnStmt:
		// Handlers cannot return from the routine (dsl.md: "Cannot
		// `return` from the routine — they execute to completion
		// and yield back."). Even though the validator gives
		// handlers inRoutine=true so abort works, return is the
		// one statement that's still forbidden inside a handler.
		switch {
		case ctx.inHandler:
			v.errorf(n.Position, "return is not allowed inside an event handler")
		case !ctx.inRoutine && !ctx.inProc:
			v.errorf(n.Position, "return outside of routine or proc")
		}
		if n.Value != nil {
			v.checkExpr(n.Value, ctx)
		}
	case *ast.AbortStmt:
		if !ctx.inRoutine {
			v.errorf(n.Position, "abort outside of routine")
		}
		v.checkExpr(n.Reason, ctx)
	case *ast.TryStmt:
		// Try body runs in normal scope. Recover body gets a fresh
		// child scope with `err` bound.
		v.checkBlock(n.Try, ctx)
		recScope := newScope(ctx.scope)
		recScope.bind(n.ErrName)
		recCtx := *ctx
		recCtx.scope = recScope
		v.checkBlock(n.Recover, &recCtx)
	case *ast.DeferStmt:
		// defer must be inside a routine/proc body. The expression
		// must be a call (we don't have an Eq/Add/etc to defer; that
		// would be a no-op). Allow inside handlers too — defers in
		// handlers are local to the handler's scope, useful for
		// per-event cleanup.
		if !ctx.inRoutine && !ctx.inProc {
			v.errorf(n.Position, "defer outside of routine or proc")
		}
		if ctx.inRequire {
			v.errorf(n.Position, "defer is forbidden inside a require block (must be pure)")
		}
		if _, ok := n.Call.(*ast.CallExpr); !ok {
			v.errorf(n.Position, "defer requires a call expression, got %T", n.Call)
		}
		v.checkExpr(n.Call, ctx)
	case *ast.WaitStmt:
		if ctx.inProc {
			v.errorf(n.Position, "wait is forbidden inside a proc (pure helpers only)")
		}
		if ctx.inRequire {
			v.errorf(n.Position, "wait is forbidden inside a require block")
		}
		v.checkExpr(n.Duration, ctx)
	case *ast.RequireBlock:
		// Handled in validateRoutine (hoisting). If we see one here,
		// it's a stray — but validateRoutine already reported it.
	default:
		v.errorf(s.Pos(), "unhandled statement type %T (validator bug)", s)
	}
}

// ----- expression validation -----

func (v *Validator) checkExpr(e ast.Expr, ctx *context) {
	if e == nil {
		return
	}
	switch n := e.(type) {
	case *ast.IntLit, *ast.FloatLit, *ast.StringLit, *ast.BoolLit, *ast.NullLit:
		// nothing to check
	case *ast.FStringLit:
		for _, p := range n.Parts {
			v.checkExpr(p, ctx)
		}
	case *ast.ListLit:
		for _, el := range n.Elems {
			v.checkExpr(el, ctx)
		}
	case *ast.RangeLit:
		v.checkExpr(n.Low, ctx)
		v.checkExpr(n.High, ctx)
	case *ast.Ident:
		v.resolveIdent(n, ctx)
	case *ast.BinaryExpr:
		v.checkExpr(n.Lhs, ctx)
		v.checkExpr(n.Rhs, ctx)
	case *ast.UnaryExpr:
		v.checkExpr(n.Rhs, ctx)
	case *ast.MemberExpr:
		v.checkExpr(n.Recv, ctx)
	case *ast.IndexExpr:
		v.checkExpr(n.Recv, ctx)
		v.checkExpr(n.Index, ctx)
	case *ast.CallExpr:
		v.checkCall(n, ctx)
	default:
		v.errorf(e.Pos(), "unhandled expression type %T (validator bug)", e)
	}
}

func (v *Validator) resolveIdent(id *ast.Ident, ctx *context) {
	if ctx.scope.resolve(id.Name) {
		return
	}
	if isReservedName(id.Name) {
		return
	}
	if _, ok := v.procs[id.Name]; ok {
		return
	}
	if _, ok := builtins[id.Name]; ok {
		return
	}
	// Bang variant: `walk_to!` is valid if `walk_to` is a builtin
	// that can fail in the typed sense (returns a Result). Procs
	// and primitives don't get bang variants.
	if bangBase, ok := unbang(id.Name); ok {
		if b, ok := builtins[bangBase]; ok && b.bangEligible {
			return
		}
	}
	v.errorf(id.Position, "unbound identifier %q", id.Name)
}

// unbang strips a trailing `!` from an identifier, returning the
// base name and whether the trailing bang was present.
func unbang(name string) (string, bool) {
	if len(name) > 1 && name[len(name)-1] == '!' {
		return name[:len(name)-1], true
	}
	return name, false
}

func (v *Validator) checkCall(c *ast.CallExpr, ctx *context) {
	// Resolve the callee. Three valid call shapes:
	//   1. Bare identifier — proc / action / primitive / stdlib
	//   2. Member chain like inventory.has(...) or world.locs.banks.nearest(...)
	//   3. Method on a parenthesized expression (rare)
	if id, ok := c.Callee.(*ast.Ident); ok {
		switch {
		case ctx.scope.resolve(id.Name):
			// shadowed by a local — allowed (might be a lambda later)
		case isReservedName(id.Name):
			// `self(...)` etc. — treat as type error later; for now allow
		case v.procs[id.Name] != nil:
			// Proc — can't be banged (procs don't return Result).
			p := v.procs[id.Name]
			v.checkProcArity(c, p)
		case builtins[id.Name].exists:
			b := builtins[id.Name]
			if b.isAction && ctx.inProc {
				v.errorf(c.Position, "action %q is forbidden inside a proc (pure helpers only)", id.Name)
			}
			if b.isAction && ctx.inRequire {
				v.errorf(c.Position, "action %q is forbidden inside a require block (must be pure)", id.Name)
			}
			v.checkBuiltinArity(c, id.Name, b)
		default:
			// Bang variant? `eat!` is valid iff `eat` is a builtin
			// flagged bangEligible. Procs and primitives can't be
			// banged.
			if bangBase, isBang := unbang(id.Name); isBang {
				if v.procs[bangBase] != nil {
					v.errorf(id.Position, "bang variant %q not allowed — procs don't return Result", id.Name)
					break
				}
				if b, ok := builtins[bangBase]; ok {
					if !b.bangEligible {
						v.errorf(id.Position, "bang variant %q not allowed — %s doesn't return Result", id.Name, bangBase)
						break
					}
					if b.isAction && ctx.inProc {
						v.errorf(c.Position, "action %q is forbidden inside a proc", id.Name)
					}
					if b.isAction && ctx.inRequire {
						v.errorf(c.Position, "action %q is forbidden inside a require block", id.Name)
					}
					v.checkBuiltinArity(c, id.Name, b)
					break
				}
			}
			v.errorf(id.Position, "call to undefined %q", id.Name)
		}
	} else {
		// Member/index callee — validate the receiver chain; we
		// don't statically check method names yet.
		v.checkExpr(c.Callee, ctx)
	}
	for _, a := range c.Args {
		v.checkExpr(a.Value, ctx)
	}
}

func (v *Validator) checkProcArity(c *ast.CallExpr, p *ast.ProcDecl) {
	provided := len(c.Args)
	required := 0
	max := len(p.Params)
	for _, param := range p.Params {
		if param.Default == nil {
			required++
		}
	}
	if provided < required || provided > max {
		v.errorf(c.Position, "proc %q expects %d–%d args, got %d", p.Name, required, max, provided)
	}
}

func (v *Validator) checkBuiltinArity(c *ast.CallExpr, name string, b builtin) {
	provided := len(c.Args)
	if b.maxArgs >= 0 && provided > b.maxArgs {
		v.errorf(c.Position, "%s takes at most %d arg(s), got %d", name, b.maxArgs, provided)
		return
	}
	if provided < b.minArgs {
		v.errorf(c.Position, "%s takes at least %d arg(s), got %d", name, b.minArgs, provided)
	}
}

// ----- builtins + reserved names -----

// builtin describes a known top-level function: actions (which mutate
// game state and are forbidden in procs/require) or primitives /
// stdlib functions (which are not actions). bangEligible means the
// callable returns a *CallResult and therefore has a `<name>!`
// variant the validator should accept.
//
// **Source of truth**: this struct is populated from
// dsl/spec/actions.go at package init. To add a builtin, add a row
// to spec.Actions, not here.
type builtin struct {
	exists       bool
	isAction     bool
	bangEligible bool // true if `<name>!` is also a valid identifier
	minArgs      int
	maxArgs      int // -1 = unbounded
}

// builtins maps every DSL callable name to its validator-relevant
// flags. Populated at init from dsl/spec/actions.go — adding a new
// builtin happens in the spec, never here.
var builtins map[string]builtin

func init() {
	builtins = make(map[string]builtin, len(spec.Actions))
	for _, a := range spec.Actions {
		builtins[a.Name] = builtin{
			exists:       true,
			isAction:     a.IsAction(),
			bangEligible: a.BangEligible(),
			minArgs:      a.MinArgs,
			maxArgs:      a.MaxArgs,
		}
	}
}

// eventArity maps event name → handler param count. Populated at
// init from dsl/spec/events.go.
var eventArity map[string]int

func init() {
	eventArity = make(map[string]int, len(spec.Events))
	for _, e := range spec.Events {
		eventArity[e.Name] = len(e.Params)
	}
}

func isReservedName(s string) bool {
	switch s {
	case "self", "world", "inventory", "combat":
		return true
	}
	return false
}
