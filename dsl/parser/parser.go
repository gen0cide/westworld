// Package parser turns a token stream into an AST per
// docs/dsl.md "Top-level constructs", "Statements", and "Expressions".
//
// Hand-written recursive descent with one-token lookahead. Errors are
// collected (not panic) so the LLM strategist can retry on parse
// failure with the full diagnostics list. Source positions propagate
// through every AST node.
//
// Expression precedence (low → high, matching the `or`-falls-down,
// `not`-binds-tight semantics in dsl.md):
//
//	1. or
//	2. and
//	3. not (unary)
//	4. ==, !=
//	5. <, <=, >, >=
//	6. + -
//	7. * / %
//	8. unary -, +
//	9. .. (range)
//	10. postfix call() / .field / [idx]
//	11. primary (literal / ident / f-string / list / paren)
package parser

import (
	"fmt"
	"strconv"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/token"
)

// Parser holds parse state for a single file.
type Parser struct {
	l        *lex.Lexer
	tokens   []token.Token
	pos      int
	errors   []error
	filename string
}

// Parse parses `src` named `filename` and returns the AST.
//
// Returns the resulting File even on error so callers can show
// partial parses; errors are collected on the returned Parser.
// A non-nil error wraps the first parse failure.
func Parse(filename, src string) (*ast.File, error) {
	l := lex.New(filename, src)
	p := &Parser{
		l:        l,
		tokens:   l.All(),
		filename: filename,
	}
	file := p.parseFile()
	if len(p.errors) > 0 {
		return file, fmt.Errorf("parse %s: %w", filename, p.errors[0])
	}
	return file, nil
}

// Errors returns every diagnostic the parser collected.
func (p *Parser) Errors() []error { return p.errors }

// ParseStmt parses a single DSL statement from `src` and returns
// it. Used by the REPL and by any other caller that needs a
// fragment parse rather than a full file. Returns the first parse
// error if any; the AST is partial on error.
//
// The statement can be any of the forms parseStmt handles: assign,
// if/elif/else, while, for-in, return, abort, wait, require block
// (becomes a RequireBlock Stmt), break/continue, or an expression
// statement.
func ParseStmt(filename, src string) (ast.Stmt, error) {
	l := lex.New(filename, src)
	p := &Parser{
		l:        l,
		tokens:   l.All(),
		filename: filename,
	}
	stmt := p.parseStmt()
	if len(p.errors) > 0 {
		return stmt, fmt.Errorf("parse %s: %w", filename, p.errors[0])
	}
	return stmt, nil
}

// ParseOnHandler parses `on event_name(params) { body }` from src
// and returns the OnHandler. Used by the REPL when a user types
// an `on` declaration at the prompt to register a live handler in
// the session.
func ParseOnHandler(filename, src string) (*ast.OnHandler, error) {
	l := lex.New(filename, src)
	p := &Parser{
		l:        l,
		tokens:   l.All(),
		filename: filename,
	}
	if p.peek().Kind != token.ON {
		return nil, fmt.Errorf("parse %s: expected `on`, got %s", filename, p.peek().Kind)
	}
	h := p.parseOnHandler()
	if len(p.errors) > 0 {
		return h, fmt.Errorf("parse %s: %w", filename, p.errors[0])
	}
	if t := p.peek(); t.Kind != token.EOF {
		return h, fmt.Errorf("parse %s: unexpected %s after on-handler", filename, t.Kind)
	}
	return h, nil
}

// ParseExpr parses a single DSL expression from `src` and returns
// it. Used by the REPL to evaluate expression-typed input lines
// for display. Returns the first parse error if any.
func ParseExpr(filename, src string) (ast.Expr, error) {
	l := lex.New(filename, src)
	p := &Parser{
		l:        l,
		tokens:   l.All(),
		filename: filename,
	}
	expr := p.parseExpr()
	if len(p.errors) > 0 {
		return expr, fmt.Errorf("parse %s: %w", filename, p.errors[0])
	}
	// Reject trailing junk — if there's anything left after the
	// expression that isn't EOF, the input wasn't a clean
	// expression. Common case: input was a statement, parsed
	// here as a prefix.
	if t := p.peek(); t.Kind != token.EOF {
		return expr, fmt.Errorf("parse %s: unexpected %s after expression", filename, t.Kind)
	}
	return expr, nil
}

// ----- low-level token plumbing -----

func (p *Parser) peek() token.Token {
	if p.pos >= len(p.tokens) {
		return token.Token{Kind: token.EOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekN(n int) token.Token {
	if p.pos+n >= len(p.tokens) {
		return token.Token{Kind: token.EOF}
	}
	return p.tokens[p.pos+n]
}

func (p *Parser) advance() token.Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) consume(k token.Kind) bool {
	if p.peek().Kind == k {
		p.advance()
		return true
	}
	return false
}

// expect consumes a token of the given kind, or records an error and
// returns the current token (without advancing) so the parser can
// attempt recovery at the caller.
func (p *Parser) expect(k token.Kind) token.Token {
	t := p.peek()
	if t.Kind != k {
		p.errorf(t.Pos, "expected %s, got %s", k, t.Kind)
		return t
	}
	return p.advance()
}

func (p *Parser) errorf(pos token.Position, format string, args ...any) {
	p.errors = append(p.errors, fmt.Errorf("%s: "+format, append([]any{pos}, args...)...))
}

// ----- top-level parsing -----

// parseFile is the entry point — a file is zero or more on-handlers,
// procs, and at most one routine.
func (p *Parser) parseFile() *ast.File {
	startPos := p.peek().Pos
	file := &ast.File{Filename: p.filename, Position: startPos}
	for {
		t := p.peek()
		switch t.Kind {
		case token.EOF:
			return file
		case token.ON:
			if h := p.parseOnHandler(); h != nil {
				file.Handlers = append(file.Handlers, h)
			}
		case token.PROC:
			if pr := p.parseProc(); pr != nil {
				file.Procs = append(file.Procs, pr)
			}
		case token.ROUTINE:
			if r := p.parseRoutine(); r != nil {
				if file.Routine != nil {
					p.errorf(r.Position, "duplicate routine declaration (one per file)")
				} else {
					file.Routine = r
				}
			}
		case token.BOUNDS:
			if b := p.parseBounds(); b != nil {
				file.Bounds = append(file.Bounds, b)
			}
		case token.EXTENDS:
			p.advance()
			pathTok := p.expect(token.STRING)
			file.Extends = append(file.Extends, pathTok.Lexeme)
		case token.RUNTIME:
			rt := p.expect(token.RUNTIME)
			verTok := p.expect(token.STRING)
			if file.Runtime != "" {
				p.errorf(rt.Pos, "duplicate runtime directive (one per file)")
			} else {
				file.Runtime = verTok.Lexeme
				file.RuntimePos = rt.Pos
			}
		default:
			p.errorf(t.Pos, "expected on / proc / routine / bounds / extends / runtime at top level, got %s", t.Kind)
			p.advance() // skip the offending token and try to recover
		}
	}
}

// parseBounds consumes a `bounds <shape>(args...) { decls }` block.
// Shape is parsed as a normal expression so any callable form
// (box(...), circle(...), nested function call) is accepted; the
// validator checks it's one of the known region constructors.
//
// Nested bounds and procs are also accepted inside the block, in
// addition to on-handlers — handlers from a nested bounds get the
// intersection of all enclosing shapes as their filter.
func (p *Parser) parseBounds() *ast.BoundsDecl {
	start := p.expect(token.BOUNDS).Pos
	shape := p.parseExpr()
	if shape == nil {
		return nil
	}
	p.expect(token.LBRACE)
	b := &ast.BoundsDecl{Position: start, Shape: shape}
	for {
		t := p.peek()
		switch t.Kind {
		case token.RBRACE:
			p.advance()
			return b
		case token.EOF:
			p.errorf(t.Pos, "unexpected EOF inside bounds block")
			return b
		case token.ON:
			if h := p.parseOnHandler(); h != nil {
				b.Handlers = append(b.Handlers, h)
			}
		case token.PROC:
			if pr := p.parseProc(); pr != nil {
				b.Procs = append(b.Procs, pr)
			}
		case token.BOUNDS:
			if nested := p.parseBounds(); nested != nil {
				b.Bounds = append(b.Bounds, nested)
			}
		default:
			p.errorf(t.Pos, "expected on / proc / bounds inside bounds block, got %s", t.Kind)
			p.advance()
		}
	}
}

func (p *Parser) parseOnHandler() *ast.OnHandler {
	start := p.expect(token.ON).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.OnHandler{Position: start, Event: name.Lexeme, Params: params, Body: body}
}

func (p *Parser) parseProc() *ast.ProcDecl {
	start := p.expect(token.PROC).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseBlock()
	return &ast.ProcDecl{Position: start, Name: name.Lexeme, Params: params, Body: body}
}

func (p *Parser) parseRoutine() *ast.RoutineDecl {
	start := p.expect(token.ROUTINE).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	r := &ast.RoutineDecl{Position: start, Name: name.Lexeme, Params: params}
	// Optional `require { ... }` and `on event() { ... }` declarations
	// may appear before the body. We could enforce ordering here but
	// dsl.md treats them as part of the routine declaration; for now
	// we accept either order ahead of the `{` that opens the body.
	//
	// Actually re-reading dsl.md: require and on-handlers are listed
	// *inside* the routine braces ahead of the body. Easier to parse
	// as block-internal items rather than declaration-level. Treating
	// them that way for now keeps the grammar simpler.
	r.Body = p.parseBlock()
	return r
}

// parseParamList consumes `param1, param2=default, ...` (without
// parens — the caller handles them). Supports default values.
func (p *Parser) parseParamList() []*ast.Param {
	var params []*ast.Param
	for p.peek().Kind == token.IDENT {
		n := p.advance()
		param := &ast.Param{Position: n.Pos, Name: n.Lexeme}
		if p.consume(token.ASSIGN) {
			param.Default = p.parseExpr()
		}
		params = append(params, param)
		if !p.consume(token.COMMA) {
			break
		}
	}
	return params
}

// ----- block + statements -----

// parseBlock consumes `{` ... `}`.
func (p *Parser) parseBlock() *ast.Block {
	start := p.expect(token.LBRACE).Pos
	block := &ast.Block{Position: start}
	for {
		t := p.peek()
		if t.Kind == token.RBRACE || t.Kind == token.EOF {
			break
		}
		s := p.parseStmt()
		if s != nil {
			block.Stmts = append(block.Stmts, s)
		}
	}
	p.expect(token.RBRACE)
	return block
}

// parseStmt dispatches to the statement that starts here. Statement
// terminators are optional `;` or newline; in practice we just keep
// going until we hit something that can't start a statement.
func (p *Parser) parseStmt() ast.Stmt {
	// Optional `;` separator is a no-op.
	for p.consume(token.SEMICOL) {
	}
	t := p.peek()
	switch t.Kind {
	case token.IF:
		return p.parseIf()
	case token.WHILE:
		return p.parseWhile()
	case token.FOR:
		return p.parseFor()
	case token.RETURN:
		return p.parseReturn()
	case token.ABORT:
		return p.parseAbort()
	case token.WAIT:
		return p.parseWait()
	case token.DEFER:
		return p.parseDefer()
	case token.TRY:
		return p.parseTry()
	case token.WHEN:
		return p.parseWhen()
	case token.SELECT:
		return p.parseSelect()
	case token.REPEAT:
		return p.parseRepeatUntil()
	case token.LBRACE:
		// Bare nested block — anonymous scope. Required by the
		// `when`-scoping design (docs/lang/events.md): watchers
		// registered inside a `{ ... }` block unregister when the
		// block exits.
		return p.parseBlock()
	case token.REQUIRE:
		// `require` at routine-body level is hoisted by the validator
		// (step 4) to RoutineDecl.Require. Until then it flows in the
		// body as a RequireBlock (which implements Stmt).
		return p.parseRequireBlock()
	case token.BREAK:
		pos := p.advance().Pos
		return &ast.BreakStmt{Position: pos}
	case token.CONTINUE:
		pos := p.advance().Pos
		return &ast.ContinueStmt{Position: pos}
	case token.RBRACE, token.EOF:
		return nil
	}
	// Either an assignment (IDENT = expr) or a bare expression.
	if t.Kind == token.IDENT && p.peekN(1).Kind == token.ASSIGN {
		return p.parseAssign()
	}
	return p.parseExprStmt()
}

func (p *Parser) parseIf() *ast.IfStmt {
	start := p.expect(token.IF).Pos
	cond := p.parseExpr()
	body := p.parseBlock()
	stmt := &ast.IfStmt{Position: start, Cond: cond, Then: body}
	for p.peek().Kind == token.ELIF {
		clauseStart := p.advance().Pos
		c := p.parseExpr()
		b := p.parseBlock()
		stmt.Elifs = append(stmt.Elifs, &ast.ElifClause{Position: clauseStart, Cond: c, Body: b})
	}
	if p.consume(token.ELSE) {
		stmt.Else = p.parseBlock()
	}
	return stmt
}

func (p *Parser) parseWhile() *ast.WhileStmt {
	start := p.expect(token.WHILE).Pos
	cond := p.parseExpr()
	body := p.parseBlock()
	return &ast.WhileStmt{Position: start, Cond: cond, Body: body}
}

func (p *Parser) parseFor() *ast.ForStmt {
	start := p.expect(token.FOR).Pos
	v := p.expect(token.IDENT)
	p.expect(token.IN)
	iter := p.parseExpr()
	body := p.parseBlock()
	return &ast.ForStmt{Position: start, Var: v.Lexeme, Iter: iter, Body: body}
}

// parseRepeatUntil consumes
//
//	repeat { stmts } until <cond> [timeout <expr>]
//
// The timeout clause is parsed if present; the validator rejects
// missing-timeout to keep "accidental infinite retry" off the table.
func (p *Parser) parseRepeatUntil() *ast.RepeatUntilStmt {
	start := p.expect(token.REPEAT).Pos
	body := p.parseBlock()
	p.expect(token.UNTIL)
	cond := p.parseExpr()
	var timeout ast.Expr
	if p.consume(token.TIMEOUT) {
		timeout = p.parseExpr()
	}
	return &ast.RepeatUntilStmt{Position: start, Body: body, Cond: cond, Timeout: timeout}
}

func (p *Parser) parseReturn() *ast.ReturnStmt {
	start := p.expect(token.RETURN).Pos
	r := &ast.ReturnStmt{Position: start}
	// `return` may have an expression or be bare. Detect bare-return
	// by looking for a block close / EOF / statement starter.
	switch p.peek().Kind {
	case token.RBRACE, token.EOF, token.SEMICOL:
		// bare
	default:
		r.Value = p.parseExpr()
	}
	return r
}

func (p *Parser) parseAbort() *ast.AbortStmt {
	start := p.expect(token.ABORT).Pos
	reason := p.parseExpr()
	return &ast.AbortStmt{Position: start, Reason: reason}
}

// parseDefer parses `defer <call-expr>`. The expression must
// syntactically be a call (validator enforces); we accept any
// expression here and let the validator complain on misuse.
func (p *Parser) parseDefer() *ast.DeferStmt {
	start := p.expect(token.DEFER).Pos
	call := p.parseExpr()
	return &ast.DeferStmt{Position: start, Call: call}
}

// parseTry parses `try { body } recover <ident> { recoverBody }`.
// The `recover` keyword is required (no bare try). The identifier
// after `recover` becomes a fresh local in the recover scope,
// bound to the caught abort value.
func (p *Parser) parseTry() *ast.TryStmt {
	start := p.expect(token.TRY).Pos
	tryBlock := p.parseBlock()
	p.expect(token.RECOVER)
	errName := p.expect(token.IDENT)
	recoverBlock := p.parseBlock()
	return &ast.TryStmt{
		Position: start,
		Try:      tryBlock,
		ErrName:  errName.Lexeme,
		Recover:  recoverBlock,
	}
}

// parseWhen parses a top-level when-watcher:
//
//	when <expr> { ... }
//	when <expr> becomes true { ... }
//	when <expr> becomes false { ... }
//	when <expr> changes { ... }
//
// Note: when also appears as a case-keyword inside select{} — that
// path is handled by parseSelect. This function is only reached
// from parseStmt at routine-body level.
func (p *Parser) parseWhen() *ast.WhenStmt {
	start := p.expect(token.WHEN).Pos
	pred := p.parseExpr()
	qual := ast.WhenBecomesTrue
	switch p.peek().Kind {
	case token.BECOMES:
		p.advance()
		switch p.peek().Kind {
		case token.TRUE:
			p.advance()
			qual = ast.WhenBecomesTrue
		case token.FALSE:
			p.advance()
			qual = ast.WhenBecomesFalse
		default:
			p.errorf(p.peek().Pos, "expected `true` or `false` after `becomes`, got %s", p.peek().Kind)
		}
	case token.CHANGES:
		p.advance()
		qual = ast.WhenChanges
	}
	body := p.parseBlock()
	return &ast.WhenStmt{Position: start, Predicate: pred, Qualifier: qual, Body: body}
}

// parseSelect parses
//
//	select {
//	    when <expr> [becomes ...|changes] { ... }
//	    on <ident>(<params>) { ... }
//	    timeout <duration> { ... }
//	}
//
// Cases may appear in any order; first-declared-wins semantics is
// handled by the interpreter, not the parser. Validator ensures at
// least one case and warns on missing timeout.
func (p *Parser) parseSelect() *ast.SelectStmt {
	start := p.expect(token.SELECT).Pos
	stmt := &ast.SelectStmt{Position: start}
	p.expect(token.LBRACE)
	for p.peek().Kind != token.RBRACE && p.peek().Kind != token.EOF {
		for p.consume(token.SEMICOL) {
		}
		if p.peek().Kind == token.RBRACE {
			break
		}
		stmt.Cases = append(stmt.Cases, p.parseSelectCase())
	}
	p.expect(token.RBRACE)
	return stmt
}

func (p *Parser) parseSelectCase() ast.SelectCase {
	t := p.peek()
	switch t.Kind {
	case token.WHEN:
		pos := p.advance().Pos
		pred := p.parseExpr()
		qual := ast.WhenBecomesTrue
		switch p.peek().Kind {
		case token.BECOMES:
			p.advance()
			switch p.peek().Kind {
			case token.TRUE:
				p.advance()
				qual = ast.WhenBecomesTrue
			case token.FALSE:
				p.advance()
				qual = ast.WhenBecomesFalse
			default:
				p.errorf(p.peek().Pos, "expected `true` or `false` after `becomes`, got %s", p.peek().Kind)
			}
		case token.CHANGES:
			p.advance()
			qual = ast.WhenChanges
		}
		body := p.parseBlock()
		return ast.SelectCase{Position: pos, Kind: ast.SelectWhenCase, Predicate: pred, Qualifier: qual, Body: body}
	case token.ON:
		pos := p.advance().Pos
		name := p.expect(token.IDENT)
		var params []string
		if p.consume(token.LPAREN) {
			for p.peek().Kind != token.RPAREN && p.peek().Kind != token.EOF {
				params = append(params, p.expect(token.IDENT).Lexeme)
				if !p.consume(token.COMMA) {
					break
				}
			}
			p.expect(token.RPAREN)
		}
		body := p.parseBlock()
		return ast.SelectCase{Position: pos, Kind: ast.SelectOnCase, EventName: name.Lexeme, EventParams: params, Body: body}
	case token.TIMEOUT:
		pos := p.advance().Pos
		ms := p.parseDurationToMillis()
		body := p.parseBlock()
		return ast.SelectCase{Position: pos, Kind: ast.SelectTimeoutCase, TimeoutMillis: ms, Body: body}
	default:
		p.errorf(t.Pos, "expected `when`, `on`, or `timeout` in select case, got %s", t.Kind)
		// Recover by skipping to next case-like token or RBRACE.
		for p.peek().Kind != token.RBRACE && p.peek().Kind != token.EOF &&
			p.peek().Kind != token.WHEN && p.peek().Kind != token.ON && p.peek().Kind != token.TIMEOUT {
			p.advance()
		}
		return ast.SelectCase{Position: t.Pos, Kind: ast.SelectWhenCase}
	}
}

// parseDurationToMillis parses an integer literal optionally followed
// by a unit suffix:
//
//	30        -> 30000 (default seconds)
//	30s       -> 30000
//	500ms     -> 500
//	2m        -> 120000
//
// The unit suffix is lexed as a separate IDENT token (since the
// lexer doesn't combine literals and identifiers). We peek for an
// IDENT in {s, ms, m} immediately following the INT and consume it
// if present.
func (p *Parser) parseDurationToMillis() int64 {
	t := p.expect(token.INT)
	val, err := strconv.ParseInt(t.Lexeme, 10, 64)
	if err != nil {
		p.errorf(t.Pos, "invalid integer literal %q in timeout: %v", t.Lexeme, err)
		val = 0
	}
	if p.peek().Kind != token.IDENT {
		return val * 1000
	}
	unit := p.peek().Lexeme
	switch unit {
	case "s":
		p.advance()
		return val * 1000
	case "ms":
		p.advance()
		return val
	case "m":
		p.advance()
		return val * 60 * 1000
	default:
		// Not a unit suffix — leave the IDENT for the next parse
		// step (likely the block body, which starts with `{` anyway).
		return val * 1000
	}
}

func (p *Parser) parseWait() *ast.WaitStmt {
	start := p.expect(token.WAIT).Pos
	// dsl.md allows both `wait 5` and `wait(5)` shapes — the latter
	// reads as a call to a builtin named wait, the former as a
	// statement. We treat both the same: parse an expression
	// argument list and synthesize the first one.
	var dur ast.Expr
	if p.consume(token.LPAREN) {
		dur = p.parseExpr()
		p.expect(token.RPAREN)
	} else {
		dur = p.parseExpr()
	}
	return &ast.WaitStmt{Position: start, Duration: dur}
}

// parseRequireBlock parses `require { cond; cond; ... }`. The parsed
// RequireBlock implements ast.Stmt so it can sit inside a routine
// body during parsing; the validator (step 4) hoists it to
// RoutineDecl.Require.
func (p *Parser) parseRequireBlock() *ast.RequireBlock {
	start := p.expect(token.REQUIRE).Pos
	p.expect(token.LBRACE)
	rb := &ast.RequireBlock{Position: start}
	for p.peek().Kind != token.RBRACE && p.peek().Kind != token.EOF {
		rb.Conds = append(rb.Conds, p.parseExpr())
		for p.consume(token.SEMICOL) {
		}
	}
	p.expect(token.RBRACE)
	return rb
}

func (p *Parser) parseAssign() *ast.AssignStmt {
	target := p.expect(token.IDENT)
	p.expect(token.ASSIGN)
	value := p.parseExpr()
	return &ast.AssignStmt{Position: target.Pos, Target: target.Lexeme, Value: value}
}

func (p *Parser) parseExprStmt() *ast.ExprStmt {
	start := p.peek().Pos
	x := p.parseExpr()
	if x == nil {
		return nil
	}
	return &ast.ExprStmt{Position: start, X: x}
}

// ----- expressions: precedence climbing -----

// parseExpr is the entry point for expression parsing — lowest
// precedence wins. The one bit of lookahead here is for
// single-arg lambdas: `IDENT =>` introduces a lambda whose body
// is itself an expression. Detect that prefix first, fall back
// to the normal precedence ladder otherwise.
func (p *Parser) parseExpr() ast.Expr {
	if p.peek().Kind == token.IDENT && p.peekN(1).Kind == token.FATARROW {
		return p.parseLambda()
	}
	return p.parseOr()
}

// parseLambda consumes `IDENT => <expr>` and produces a LambdaExpr.
// The body is whatever comes next at expression scope, so lambdas
// chain naturally: `n => n + 1` or `n => n.combat_level < 30`.
func (p *Parser) parseLambda() ast.Expr {
	ident := p.expect(token.IDENT)
	p.expect(token.FATARROW)
	body := p.parseExpr()
	return &ast.LambdaExpr{Position: ident.Pos, Param: ident.Lexeme, Body: body}
}

func (p *Parser) parseOr() ast.Expr {
	left := p.parseAnd()
	for p.peek().Kind == token.OR {
		op := p.advance()
		right := p.parseAnd()
		left = &ast.BinaryExpr{Position: op.Pos, Op: token.OR, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseAnd() ast.Expr {
	left := p.parseNot()
	for p.peek().Kind == token.AND {
		op := p.advance()
		right := p.parseNot()
		left = &ast.BinaryExpr{Position: op.Pos, Op: token.AND, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseNot() ast.Expr {
	if p.peek().Kind == token.NOT {
		op := p.advance()
		rhs := p.parseNot()
		return &ast.UnaryExpr{Position: op.Pos, Op: token.NOT, Rhs: rhs}
	}
	return p.parseEquality()
}

func (p *Parser) parseEquality() ast.Expr {
	left := p.parseComparison()
	for {
		k := p.peek().Kind
		if k != token.EQ && k != token.NEQ {
			break
		}
		op := p.advance()
		right := p.parseComparison()
		left = &ast.BinaryExpr{Position: op.Pos, Op: k, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseComparison() ast.Expr {
	left := p.parseAdditive()
	for {
		k := p.peek().Kind
		if k != token.LT && k != token.LTE && k != token.GT && k != token.GTE {
			break
		}
		op := p.advance()
		right := p.parseAdditive()
		left = &ast.BinaryExpr{Position: op.Pos, Op: k, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseAdditive() ast.Expr {
	left := p.parseMultiplicative()
	for {
		k := p.peek().Kind
		if k != token.ADD && k != token.SUB {
			break
		}
		op := p.advance()
		right := p.parseMultiplicative()
		left = &ast.BinaryExpr{Position: op.Pos, Op: k, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseMultiplicative() ast.Expr {
	left := p.parseUnary()
	for {
		k := p.peek().Kind
		if k != token.MUL && k != token.QUO && k != token.REM {
			break
		}
		op := p.advance()
		right := p.parseUnary()
		left = &ast.BinaryExpr{Position: op.Pos, Op: k, Lhs: left, Rhs: right}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expr {
	k := p.peek().Kind
	if k == token.SUB || k == token.ADD {
		op := p.advance()
		rhs := p.parseUnary()
		return &ast.UnaryExpr{Position: op.Pos, Op: k, Rhs: rhs}
	}
	return p.parseRange()
}

// parseRange handles `low..high`. We treat it like a binary operator
// sitting just above postfix ops so `wait 2.8..4.5` and
// `a + 1 .. b + 2` both work. Range is non-associative — `a..b..c` is
// a parse error, but we just stop after one `..` and let the caller
// surface anything weird.
func (p *Parser) parseRange() ast.Expr {
	left := p.parsePostfix()
	if p.peek().Kind == token.DOTDOT {
		op := p.advance()
		right := p.parsePostfix()
		return &ast.RangeLit{Position: op.Pos, Low: left, High: right}
	}
	return left
}

// parsePostfix handles call(), .field, [idx] chains on a primary.
func (p *Parser) parsePostfix() ast.Expr {
	x := p.parsePrimary()
	for {
		switch p.peek().Kind {
		case token.LPAREN:
			x = p.parseCallTail(x)
		case token.DOT:
			x = p.parseMemberTail(x)
		case token.LBRACK:
			x = p.parseIndexTail(x)
		default:
			return x
		}
	}
}

func (p *Parser) parseCallTail(callee ast.Expr) ast.Expr {
	start := p.expect(token.LPAREN).Pos
	args := p.parseArgList()
	p.expect(token.RPAREN)
	return &ast.CallExpr{Position: start, Callee: callee, Args: args}
}

// parseArgList: empty, or `arg1 [, arg2]*` where each arg is either
// `expr` (positional) or `name = expr` (named).
func (p *Parser) parseArgList() []*ast.Arg {
	if p.peek().Kind == token.RPAREN {
		return nil
	}
	var args []*ast.Arg
	for {
		args = append(args, p.parseArg())
		if !p.consume(token.COMMA) {
			break
		}
	}
	return args
}

func (p *Parser) parseArg() *ast.Arg {
	start := p.peek().Pos
	// Look-ahead for named argument: IDENT = expr.
	if p.peek().Kind == token.IDENT && p.peekN(1).Kind == token.ASSIGN {
		name := p.advance().Lexeme
		p.advance() // =
		return &ast.Arg{Position: start, Name: name, Value: p.parseExpr()}
	}
	return &ast.Arg{Position: start, Value: p.parseExpr()}
}

func (p *Parser) parseMemberTail(recv ast.Expr) ast.Expr {
	start := p.expect(token.DOT).Pos
	id := p.expect(token.IDENT)
	return &ast.MemberExpr{Position: start, Recv: recv, Field: id.Lexeme}
}

func (p *Parser) parseIndexTail(recv ast.Expr) ast.Expr {
	start := p.expect(token.LBRACK).Pos
	idx := p.parseExpr()
	p.expect(token.RBRACK)
	return &ast.IndexExpr{Position: start, Recv: recv, Index: idx}
}

// parsePrimary handles atomic forms: literals, identifiers,
// f-strings, parenthesized expressions, list literals.
func (p *Parser) parsePrimary() ast.Expr {
	t := p.peek()
	switch t.Kind {
	case token.INT:
		p.advance()
		v, err := strconv.ParseInt(t.Lexeme, 10, 64)
		if err != nil {
			p.errorf(t.Pos, "invalid integer literal %q: %v", t.Lexeme, err)
			return &ast.IntLit{Position: t.Pos}
		}
		return &ast.IntLit{Position: t.Pos, Value: v}
	case token.FLOAT:
		p.advance()
		v, err := strconv.ParseFloat(t.Lexeme, 64)
		if err != nil {
			p.errorf(t.Pos, "invalid float literal %q: %v", t.Lexeme, err)
			return &ast.FloatLit{Position: t.Pos}
		}
		return &ast.FloatLit{Position: t.Pos, Value: v}
	case token.STRING:
		p.advance()
		return &ast.StringLit{Position: t.Pos, Value: t.Lexeme}
	case token.FSTRING_OPEN:
		return p.parseFString()
	case token.TRUE:
		p.advance()
		return &ast.BoolLit{Position: t.Pos, Value: true}
	case token.FALSE:
		p.advance()
		return &ast.BoolLit{Position: t.Pos, Value: false}
	case token.NULL:
		p.advance()
		return &ast.NullLit{Position: t.Pos}
	case token.IDENT:
		p.advance()
		return &ast.Ident{Position: t.Pos, Name: t.Lexeme}
	case token.LPAREN:
		p.advance()
		inner := p.parseExpr()
		p.expect(token.RPAREN)
		return inner
	case token.LBRACK:
		return p.parseListLit()
	}
	p.errorf(t.Pos, "expected expression, got %s", t.Kind)
	p.advance() // skip the bad token so we don't loop forever
	return nil
}

func (p *Parser) parseListLit() ast.Expr {
	start := p.expect(token.LBRACK).Pos
	list := &ast.ListLit{Position: start}
	if p.peek().Kind != token.RBRACK {
		for {
			list.Elems = append(list.Elems, p.parseExpr())
			if !p.consume(token.COMMA) {
				break
			}
		}
	}
	p.expect(token.RBRACK)
	return list
}

// parseFString consumes an f-string sequence starting at FSTRING_OPEN.
// Builds a FStringLit whose Parts alternate between StringLit (literal
// fragments) and arbitrary expressions (placeholder bodies).
func (p *Parser) parseFString() ast.Expr {
	open := p.advance() // FSTRING_OPEN
	fs := &ast.FStringLit{Position: open.Pos}
	if open.Lexeme != "" {
		fs.Parts = append(fs.Parts, &ast.StringLit{Position: open.Pos, Value: open.Lexeme})
	}
	for {
		// We're inside a placeholder body. Parse an expression up
		// until the lexer-emitted FSTRING_PART or FSTRING_CLOSE.
		expr := p.parseExpr()
		if expr != nil {
			fs.Parts = append(fs.Parts, expr)
		}
		t := p.peek()
		switch t.Kind {
		case token.FSTRING_PART:
			p.advance()
			if t.Lexeme != "" {
				fs.Parts = append(fs.Parts, &ast.StringLit{Position: t.Pos, Value: t.Lexeme})
			}
			continue
		case token.FSTRING_CLOSE:
			p.advance()
			if t.Lexeme != "" {
				fs.Parts = append(fs.Parts, &ast.StringLit{Position: t.Pos, Value: t.Lexeme})
			}
			return fs
		default:
			p.errorf(t.Pos, "expected f-string continuation, got %s", t.Kind)
			return fs
		}
	}
}
