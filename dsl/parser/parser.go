// Package parser turns a token stream into an AST per
// docs/dsl.md "Top-level constructs".
//
// The parser is hand-written recursive descent with a one-token
// lookahead buffer. Errors are collected (not panic) so the LLM can
// retry on parse failure with the full diagnostics list. Source
// positions propagate through every AST node.
//
// IMPLEMENTATION STATUS: skeleton — top-level dispatch + identifier
// recognition only. Filling in routines/procs/handlers + statements
// + expression precedence climbing is the next chunk.
package parser

import (
	"fmt"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/token"
)

// Parser holds parse state for a single file.
type Parser struct {
	l        *lex.Lexer
	tokens   []token.Token // pre-buffered for lookahead
	pos      int           // index into tokens
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

// expect consumes a token of the given kind, or records an error.
func (p *Parser) expect(k token.Kind) token.Token {
	t := p.peek()
	if t.Kind != k {
		p.errorf(t.Pos, "expected %s, got %s", k, t.Kind)
	}
	return p.advance()
}

func (p *Parser) errorf(pos token.Position, format string, args ...any) {
	p.errors = append(p.errors, fmt.Errorf("%s: "+format, append([]any{pos}, args...)...))
}

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
			h := p.parseOnHandler()
			if h != nil {
				file.Handlers = append(file.Handlers, h)
			}
		case token.PROC:
			pr := p.parseProc()
			if pr != nil {
				file.Procs = append(file.Procs, pr)
			}
		case token.ROUTINE:
			r := p.parseRoutine()
			if r != nil {
				if file.Routine != nil {
					p.errorf(r.Position, "duplicate routine declaration (one per file)")
				} else {
					file.Routine = r
				}
			}
		default:
			p.errorf(t.Pos, "expected on / proc / routine at top level, got %s", t.Kind)
			// Skip to next plausible recovery point.
			p.advance()
		}
	}
}

// The body parsers below are stubs — they consume the matching brace
// block via a simple bracket-counter so we can validate token shape
// without yet building real statement ASTs. The next implementation
// pass fills in real statement + expression parsing.

func (p *Parser) parseOnHandler() *ast.OnHandler {
	start := p.expect(token.ON).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseStubBlock()
	return &ast.OnHandler{Position: start, Event: name.Lexeme, Params: params, Body: body}
}

func (p *Parser) parseProc() *ast.ProcDecl {
	start := p.expect(token.PROC).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseStubBlock()
	return &ast.ProcDecl{Position: start, Name: name.Lexeme, Params: params, Body: body}
}

func (p *Parser) parseRoutine() *ast.RoutineDecl {
	start := p.expect(token.ROUTINE).Pos
	name := p.expect(token.IDENT)
	p.expect(token.LPAREN)
	params := p.parseParamList()
	p.expect(token.RPAREN)
	body := p.parseStubBlock()
	return &ast.RoutineDecl{Position: start, Name: name.Lexeme, Params: params, Body: body}
}

// parseParamList consumes `param1, param2=default, ...` (without
// parens — the caller handles them). v1 stub: name only, no defaults.
func (p *Parser) parseParamList() []*ast.Param {
	var params []*ast.Param
	for p.peek().Kind == token.IDENT {
		n := p.advance()
		params = append(params, &ast.Param{Position: n.Pos, Name: n.Lexeme})
		if p.peek().Kind != token.COMMA {
			break
		}
		p.advance() // consume comma
	}
	return params
}

// parseStubBlock consumes a brace-balanced block without parsing the
// statements inside. It exists so the top-level structure can be
// validated end-to-end before statement parsing is fleshed out.
func (p *Parser) parseStubBlock() *ast.Block {
	start := p.expect(token.LBRACE).Pos
	depth := 1
	for depth > 0 {
		t := p.peek()
		switch t.Kind {
		case token.EOF:
			p.errorf(t.Pos, "unexpected EOF inside block")
			return &ast.Block{Position: start}
		case token.LBRACE:
			depth++
		case token.RBRACE:
			depth--
		}
		p.advance()
	}
	return &ast.Block{Position: start}
}
