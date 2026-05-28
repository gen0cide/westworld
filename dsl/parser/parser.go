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
		default:
			p.errorf(t.Pos, "expected on / proc / routine at top level, got %s", t.Kind)
			p.advance() // skip the offending token and try to recover
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
// precedence wins.
func (p *Parser) parseExpr() ast.Expr {
	return p.parseOr()
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
