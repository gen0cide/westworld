// Package token defines the lexical units of the .ws routine
// language plus source-position bookkeeping.
//
// Token shapes follow docs/dsl.md "Lexical structure".
package token

import "fmt"

// Kind classifies a token.
type Kind int

const (
	ILLEGAL Kind = iota
	EOF
	NEWLINE // optional terminator; the parser is brace-delimited so newlines are non-significant except inside f-strings

	// Identifiers + literals.
	IDENT  // foo, my_proc
	INT    // 42, -7 (negative handled at parser via unary)
	FLOAT  // 3.14, 0.5
	STRING // "hello"
	FSTRING_OPEN  // f"hi {
	FSTRING_PART  // (literal text inside an f-string)
	FSTRING_CLOSE // }..."
	// NOTE: f-string interpolation is lexed as alternating FSTRING_PART
	// + embedded normal tokens until FSTRING_CLOSE.

	// Operators (in roughly precedence order, low → high).
	OR  // or
	AND // and
	NOT // not
	EQ  // ==
	NEQ // !=
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=
	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	ASSIGN // =

	// Punctuation.
	LPAREN   // (
	RPAREN   // )
	LBRACE   // {
	RBRACE   // }
	LBRACK   // [
	RBRACK   // ]
	COMMA    // ,
	DOT      // .
	DOTDOT   // .. (range)
	COLON    // : (reserved; not yet used but anticipated for type annotations / dict literals if added)
	SEMICOL  // ;
	ARROW    // -> (reserved for type annotations / lambdas)

	// Keywords (the language is small).
	ROUTINE
	PROC
	ON
	IF
	ELIF
	ELSE
	WHILE
	FOR
	IN
	BREAK
	CONTINUE
	RETURN
	ABORT
	REQUIRE
	WAIT
	DEFER
	TRUE
	FALSE
	NULL
)

// String renders the token kind for diagnostics. Keep concise — these
// appear in error messages.
func (k Kind) String() string {
	switch k {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case NEWLINE:
		return "NEWLINE"
	case IDENT:
		return "IDENT"
	case INT:
		return "INT"
	case FLOAT:
		return "FLOAT"
	case STRING:
		return "STRING"
	case FSTRING_OPEN:
		return "FSTRING_OPEN"
	case FSTRING_PART:
		return "FSTRING_PART"
	case FSTRING_CLOSE:
		return "FSTRING_CLOSE"
	case OR:
		return "or"
	case AND:
		return "and"
	case NOT:
		return "not"
	case EQ:
		return "=="
	case NEQ:
		return "!="
	case LT:
		return "<"
	case LTE:
		return "<="
	case GT:
		return ">"
	case GTE:
		return ">="
	case ADD:
		return "+"
	case SUB:
		return "-"
	case MUL:
		return "*"
	case QUO:
		return "/"
	case REM:
		return "%"
	case ASSIGN:
		return "="
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACK:
		return "["
	case RBRACK:
		return "]"
	case COMMA:
		return ","
	case DOT:
		return "."
	case DOTDOT:
		return ".."
	case COLON:
		return ":"
	case SEMICOL:
		return ";"
	case ARROW:
		return "->"
	case ROUTINE:
		return "routine"
	case PROC:
		return "proc"
	case ON:
		return "on"
	case IF:
		return "if"
	case ELIF:
		return "elif"
	case ELSE:
		return "else"
	case WHILE:
		return "while"
	case FOR:
		return "for"
	case IN:
		return "in"
	case BREAK:
		return "break"
	case CONTINUE:
		return "continue"
	case RETURN:
		return "return"
	case ABORT:
		return "abort"
	case REQUIRE:
		return "require"
	case WAIT:
		return "wait"
	case DEFER:
		return "defer"
	case TRUE:
		return "true"
	case FALSE:
		return "false"
	case NULL:
		return "null"
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Keywords maps reserved-word source text to its Kind. Lookups for
// identifier candidates go here; anything not in the map is an IDENT.
var Keywords = map[string]Kind{
	"routine":  ROUTINE,
	"proc":     PROC,
	"on":       ON,
	"if":       IF,
	"elif":     ELIF,
	"else":     ELSE,
	"while":    WHILE,
	"for":      FOR,
	"in":       IN,
	"break":    BREAK,
	"continue": CONTINUE,
	"return":   RETURN,
	"abort":    ABORT,
	"require":  REQUIRE,
	"wait":     WAIT,
	"defer":    DEFER,
	"true":     TRUE,
	"false":    FALSE,
	"null":     NULL,
	"and":      AND,
	"or":       OR,
	"not":      NOT,
}

// Position locates a token in the source. Filename is whatever the
// lexer was given (often a relative path); Line and Column are
// 1-based for human-friendly error messages; Offset is the byte
// index for tooling.
type Position struct {
	Filename string
	Line     int
	Column   int
	Offset   int
}

// String formats as "filename:line:col" for use in error messages.
func (p Position) String() string {
	if p.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token is a lexical unit produced by the lexer.
//
// Lexeme is the raw source text for literals (and the keyword text
// for keywords). For string/fstring tokens, Lexeme is the *content*
// already de-quoted and with escape sequences resolved.
type Token struct {
	Kind   Kind
	Lexeme string
	Pos    Position
}

// String formats a Token for tests / debug printing.
func (t Token) String() string {
	if t.Lexeme != "" && t.Lexeme != t.Kind.String() {
		return fmt.Sprintf("%s(%q) at %s", t.Kind, t.Lexeme, t.Pos)
	}
	return fmt.Sprintf("%s at %s", t.Kind, t.Pos)
}
