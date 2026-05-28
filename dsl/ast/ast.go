// Package ast defines the abstract syntax tree for the .routine DSL.
//
// Every node implements Node so the evaluator and validator can
// traverse uniformly. Position information is preserved on every node
// for diagnostics.
//
// AST shapes follow docs/dsl.md "Top-level constructs", "Statements",
// and "Expressions" sections.
package ast

import "github.com/gen0cide/westworld/dsl/token"

// Node is the root interface for every AST entity. Pos returns the
// source position of the first token that produced this node.
type Node interface {
	Pos() token.Position
	// astNode is a private marker so the interface can't be
	// accidentally implemented outside this package.
	astNode()
}

// Stmt and Expr are sub-interfaces that the parser uses to constrain
// where each node can appear.
type Stmt interface {
	Node
	stmt()
}

type Expr interface {
	Node
	expr()
}

// File is a parsed .routine source file. It contains top-level
// declarations only — `on` handlers, `proc`s, and at most one
// `routine`. See dsl.md "Top-level constructs".
type File struct {
	Filename     string
	Handlers     []*OnHandler
	Procs        []*ProcDecl
	Routine      *RoutineDecl // may be nil if this is a "library" file (procs only)
	Position     token.Position
}

func (f *File) Pos() token.Position { return f.Position }
func (f *File) astNode()            {}

// ----- Top-level declarations -----

// OnHandler is `on event_name(params) { body }`. The handler runs
// when its named event fires.
type OnHandler struct {
	Position token.Position
	Event    string  // event name (chat_received, hp_below, etc.)
	Params   []*Param
	Body     *Block
}

func (h *OnHandler) Pos() token.Position { return h.Position }
func (h *OnHandler) astNode()            {}

// ProcDecl is `proc name(params) { body }`. Pure helper function.
type ProcDecl struct {
	Position token.Position
	Name     string
	Params   []*Param
	Body     *Block
}

func (p *ProcDecl) Pos() token.Position { return p.Position }
func (p *ProcDecl) astNode()            {}

// RoutineDecl is `routine name(params) { require { ... } on ... body }`.
// Entry-point declaration; one per file.
type RoutineDecl struct {
	Position token.Position
	Name     string
	Params   []*Param
	Require  *RequireBlock // optional; nil if no `require`
	Handlers []*OnHandler  // routine-scoped on handlers, override persona defaults while running
	Body     *Block
}

func (r *RoutineDecl) Pos() token.Position { return r.Position }
func (r *RoutineDecl) astNode()            {}

// Param is a formal parameter on a routine, proc, or handler. Default
// is the literal default value if the param is optional; nil if not.
type Param struct {
	Position token.Position
	Name     string
	Default  Expr // optional
}

func (p *Param) Pos() token.Position { return p.Position }
func (p *Param) astNode()            {}

// RequireBlock is `require { precondition1; precondition2; ... }`.
// Each Cond is a pure expression evaluated before the routine body
// runs. A falsey value aborts with "precondition_failed: <expr>".
//
// RequireBlock implements Stmt so the parser can flow it through a
// routine body during parsing; the validator later hoists it from
// the body's Stmts to RoutineDecl.Require.
type RequireBlock struct {
	Position token.Position
	Conds    []Expr
}

func (r *RequireBlock) Pos() token.Position { return r.Position }
func (r *RequireBlock) astNode()            {}
func (r *RequireBlock) stmt()               {}

// ----- Statements -----

// Block is a sequence of statements between `{` and `}`.
type Block struct {
	Position token.Position
	Stmts    []Stmt
}

func (b *Block) Pos() token.Position { return b.Position }
func (b *Block) astNode()            {}
func (b *Block) stmt()               {}

// AssignStmt is `target = value`. Target is a bare identifier for
// v1; field/index assignment is not supported.
type AssignStmt struct {
	Position token.Position
	Target   string
	Value    Expr
}

func (s *AssignStmt) Pos() token.Position { return s.Position }
func (s *AssignStmt) astNode()            {}
func (s *AssignStmt) stmt()               {}

// ExprStmt is a bare expression at statement position — usually an
// action call or a side-effecting call.
type ExprStmt struct {
	Position token.Position
	X        Expr
}

func (s *ExprStmt) Pos() token.Position { return s.Position }
func (s *ExprStmt) astNode()            {}
func (s *ExprStmt) stmt()               {}

// IfStmt is `if cond { ... } elif cond { ... } else { ... }`.
// Elifs are flattened: each entry pairs a condition with a block. Else
// is the trailing block (may be nil).
type IfStmt struct {
	Position token.Position
	Cond     Expr
	Then     *Block
	Elifs    []*ElifClause
	Else     *Block
}

type ElifClause struct {
	Position token.Position
	Cond     Expr
	Body     *Block
}

func (s *IfStmt) Pos() token.Position    { return s.Position }
func (s *IfStmt) astNode()               {}
func (s *IfStmt) stmt()                  {}
func (c *ElifClause) Pos() token.Position { return c.Position }
func (c *ElifClause) astNode()            {}

// WhileStmt is `while cond { body }`.
type WhileStmt struct {
	Position token.Position
	Cond     Expr
	Body     *Block
}

func (s *WhileStmt) Pos() token.Position { return s.Position }
func (s *WhileStmt) astNode()            {}
func (s *WhileStmt) stmt()               {}

// ForStmt is `for x in iter { body }`. The DSL uses for-in over
// collections only; classic C-style for is not supported.
type ForStmt struct {
	Position token.Position
	Var      string
	Iter     Expr
	Body     *Block
}

func (s *ForStmt) Pos() token.Position { return s.Position }
func (s *ForStmt) astNode()            {}
func (s *ForStmt) stmt()               {}

// BreakStmt / ContinueStmt are loop controls. No labels in v1.
type BreakStmt struct{ Position token.Position }

func (s *BreakStmt) Pos() token.Position { return s.Position }
func (s *BreakStmt) astNode()            {}
func (s *BreakStmt) stmt()               {}

type ContinueStmt struct{ Position token.Position }

func (s *ContinueStmt) Pos() token.Position { return s.Position }
func (s *ContinueStmt) astNode()            {}
func (s *ContinueStmt) stmt()               {}

// ReturnStmt is `return [expr]`. Value is nil for bare `return`.
type ReturnStmt struct {
	Position token.Position
	Value    Expr // may be nil
}

func (s *ReturnStmt) Pos() token.Position { return s.Position }
func (s *ReturnStmt) astNode()            {}
func (s *ReturnStmt) stmt()               {}

// AbortStmt is `abort "reason"` — exits the routine with the given
// completion code (typically a string literal but any expression is
// accepted).
type AbortStmt struct {
	Position token.Position
	Reason   Expr
}

func (s *AbortStmt) Pos() token.Position { return s.Position }
func (s *AbortStmt) astNode()            {}
func (s *AbortStmt) stmt()               {}

// WaitStmt is `wait seconds` (literal or expr) — yields the routine
// for the given duration. Seconds may be a range literal handled by
// the evaluator (NumberLit with Range==true).
type WaitStmt struct {
	Position token.Position
	Duration Expr
}

func (s *WaitStmt) Pos() token.Position { return s.Position }
func (s *WaitStmt) astNode()            {}
func (s *WaitStmt) stmt()               {}

// DeferStmt is `defer expr` — captures the call expression at the
// point of declaration; the call runs (LIFO) when the enclosing
// routine exits (normal return, abort, runtime error, or context
// cancel). Per docs/lang/events.md "defer fn()". The Call field
// must be a CallExpr — the validator rejects non-call deferrals.
type DeferStmt struct {
	Position token.Position
	Call     Expr // typically *CallExpr, validated downstream
}

func (s *DeferStmt) Pos() token.Position { return s.Position }
func (s *DeferStmt) astNode()            {}
func (s *DeferStmt) stmt()               {}

// TryStmt is `try { ... } recover err { ... }` — a bang-error
// boundary. The Try block runs normally; if any bang-form action
// inside it aborts (or any explicit `abort` statement fires),
// control jumps to the Recover block with the abort reason bound
// to the ErrName identifier. Per docs/lang/events.md
// "try/recover".
//
// Deferred calls registered inside the Try block run before the
// Recover block executes (Go semantics).
//
// `recover` does NOT catch `return`, `break`, `continue`, or
// runtime errors (like division by zero) — only `abort`-style
// signals. Runtime errors propagate and end the routine as
// ResultErrored.
type TryStmt struct {
	Position token.Position
	Try      *Block
	ErrName  string // identifier the recover block binds the error to
	Recover  *Block
}

func (s *TryStmt) Pos() token.Position { return s.Position }
func (s *TryStmt) astNode()            {}
func (s *TryStmt) stmt()               {}

// WhenQualifier is the transition kind for `when <expr> [qualifier] { ... }`.
type WhenQualifier int

const (
	// WhenBecomesTrue is the default: rising-edge — fire when the
	// predicate transitions from false to true. Registration with
	// already-true counts as the rising edge.
	WhenBecomesTrue WhenQualifier = iota
	// WhenBecomesFalse is `when x becomes false { ... }` — falling edge.
	WhenBecomesFalse
	// WhenChanges is `when x changes { ... }` — fire on either edge.
	WhenChanges
)

func (q WhenQualifier) String() string {
	switch q {
	case WhenBecomesTrue:
		return "becomes-true"
	case WhenBecomesFalse:
		return "becomes-false"
	case WhenChanges:
		return "changes"
	}
	return "unknown"
}

// WhenStmt is `when <Predicate> [becomes true|false | changes] { Body }`.
// Block-scoped state-transition watcher. The watcher registers on
// block entry and unregisters on block exit (any path). The Body
// fires when the predicate transitions per Qualifier. Per
// docs/lang/events.md "when <expr> — state-transition watchers".
//
// At the top level of a routine the `when` watcher is active for
// the entire routine body. Inside any nested block (if-then,
// while-body, explicit `{ ... }` block) it's active for that
// block's lifetime only.
type WhenStmt struct {
	Position  token.Position
	Predicate Expr
	Qualifier WhenQualifier
	Body      *Block
}

func (s *WhenStmt) Pos() token.Position { return s.Position }
func (s *WhenStmt) astNode()            {}
func (s *WhenStmt) stmt()               {}

// SelectStmt is `select { <cases> }` — block at this point in the
// routine until one of the cases becomes ready, then run that
// case's body and exit. Per docs/lang/events.md "select — block
// until one of these fires".
//
// Each case is a SelectCase; cases are first-declared-wins on
// simultaneous readiness (deterministic, unlike Go's select).
type SelectStmt struct {
	Position token.Position
	Cases    []SelectCase
}

func (s *SelectStmt) Pos() token.Position { return s.Position }
func (s *SelectStmt) astNode()            {}
func (s *SelectStmt) stmt()               {}

// SelectCaseKind discriminates the three case forms.
type SelectCaseKind int

const (
	SelectWhenCase    SelectCaseKind = iota // when <expr> [qualifier] { ... }
	SelectOnCase                            // on <event>(<params>) { ... }
	SelectTimeoutCase                       // timeout <duration> { ... }
)

// SelectCase is one case in a select block.
//
// Field semantics by Kind:
//   - SelectWhenCase:    Predicate + Qualifier; EventName/EventParams/TimeoutMillis ignored
//   - SelectOnCase:      EventName + EventParams; Predicate/Qualifier/TimeoutMillis ignored
//   - SelectTimeoutCase: TimeoutMillis; everything else ignored
//
// All cases carry a Body. Position is the case's keyword position
// (the `when`/`on`/`timeout` token).
type SelectCase struct {
	Position      token.Position
	Kind          SelectCaseKind
	Predicate     Expr          // SelectWhenCase
	Qualifier     WhenQualifier // SelectWhenCase
	EventName     string        // SelectOnCase
	EventParams   []string      // SelectOnCase param names (positional bindings)
	TimeoutMillis int64         // SelectTimeoutCase — total duration in milliseconds
	Body          *Block
}

// ----- Expressions -----

// IntLit is an integer literal.
type IntLit struct {
	Position token.Position
	Value    int64
}

func (e *IntLit) Pos() token.Position { return e.Position }
func (e *IntLit) astNode()            {}
func (e *IntLit) expr()               {}

// FloatLit is a floating-point literal.
type FloatLit struct {
	Position token.Position
	Value    float64
}

func (e *FloatLit) Pos() token.Position { return e.Position }
func (e *FloatLit) astNode()            {}
func (e *FloatLit) expr()               {}

// StringLit is a "..." literal.
type StringLit struct {
	Position token.Position
	Value    string
}

func (e *StringLit) Pos() token.Position { return e.Position }
func (e *StringLit) astNode()            {}
func (e *StringLit) expr()               {}

// FStringLit is an f"..." literal: a sequence of literal fragments
// alternating with interpolation expressions. Parts[i] is a *StringLit
// for fragments, any Expr for placeholders. The evaluator stringifies
// each placeholder and concatenates.
type FStringLit struct {
	Position token.Position
	Parts    []Expr
}

func (e *FStringLit) Pos() token.Position { return e.Position }
func (e *FStringLit) astNode()            {}
func (e *FStringLit) expr()               {}

// BoolLit is true/false.
type BoolLit struct {
	Position token.Position
	Value    bool
}

func (e *BoolLit) Pos() token.Position { return e.Position }
func (e *BoolLit) astNode()            {}
func (e *BoolLit) expr()               {}

// NullLit is `null`.
type NullLit struct {
	Position token.Position
}

func (e *NullLit) Pos() token.Position { return e.Position }
func (e *NullLit) astNode()            {}
func (e *NullLit) expr()               {}

// RangeLit is `low..high` used for jittered waits — `wait(2.8..4.5)`.
type RangeLit struct {
	Position token.Position
	Low      Expr
	High     Expr
}

func (e *RangeLit) Pos() token.Position { return e.Position }
func (e *RangeLit) astNode()            {}
func (e *RangeLit) expr()               {}

// ListLit is `[a, b, c]`.
type ListLit struct {
	Position token.Position
	Elems    []Expr
}

func (e *ListLit) Pos() token.Position { return e.Position }
func (e *ListLit) astNode()            {}
func (e *ListLit) expr()               {}

// Ident is a bare identifier reference (local var, parameter, or
// implicit global like `self`, `world`, `inventory`).
type Ident struct {
	Position token.Position
	Name     string
}

func (e *Ident) Pos() token.Position { return e.Position }
func (e *Ident) astNode()            {}
func (e *Ident) expr()               {}

// BinaryExpr is `lhs op rhs` for arithmetic, comparison, and
// logical operators.
type BinaryExpr struct {
	Position token.Position
	Op       token.Kind
	Lhs, Rhs Expr
}

func (e *BinaryExpr) Pos() token.Position { return e.Position }
func (e *BinaryExpr) astNode()            {}
func (e *BinaryExpr) expr()               {}

// UnaryExpr is `op rhs` for `not`, unary `-`, unary `+`.
type UnaryExpr struct {
	Position token.Position
	Op       token.Kind
	Rhs      Expr
}

func (e *UnaryExpr) Pos() token.Position { return e.Position }
func (e *UnaryExpr) astNode()            {}
func (e *UnaryExpr) expr()               {}

// CallExpr is `callee(arg1, arg2=value, ...)`. Args is a flat slice;
// named arguments have NamedArg != "".
type CallExpr struct {
	Position token.Position
	Callee   Expr
	Args     []*Arg
}

type Arg struct {
	Position token.Position
	Name     string // empty for positional args
	Value    Expr
}

func (e *CallExpr) Pos() token.Position { return e.Position }
func (e *CallExpr) astNode()            {}
func (e *CallExpr) expr()               {}
func (a *Arg) Pos() token.Position      { return a.Position }
func (a *Arg) astNode()                 {}

// MemberExpr is `recv.field` for attribute access.
type MemberExpr struct {
	Position token.Position
	Recv     Expr
	Field    string
}

func (e *MemberExpr) Pos() token.Position { return e.Position }
func (e *MemberExpr) astNode()            {}
func (e *MemberExpr) expr()               {}

// LambdaExpr is `IDENT => body` — a single-argument anonymous
// function used as a predicate for filter/map/find/nearest. The
// body is one expression; for multi-statement logic use a named
// proc and pass it by name instead. See docs/lang/syntax.md
// "Lambdas".
//
// Multi-argument lambdas (`(a, b) => body`) are planned but not
// in v1; the parser only accepts single-IDENT form.
type LambdaExpr struct {
	Position token.Position
	Param    string
	Body     Expr
}

func (e *LambdaExpr) Pos() token.Position { return e.Position }
func (e *LambdaExpr) astNode()            {}
func (e *LambdaExpr) expr()               {}

// IndexExpr is `recv[index]` for collection access.
type IndexExpr struct {
	Position token.Position
	Recv     Expr
	Index    Expr
}

func (e *IndexExpr) Pos() token.Position { return e.Position }
func (e *IndexExpr) astNode()            {}
func (e *IndexExpr) expr()               {}
