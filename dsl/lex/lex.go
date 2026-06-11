// Package lex tokenizes .ws routine source code.
//
// The lexer is a hand-written state machine — no regex, no generated
// code. It produces tokens lazily via Next() so the parser can pull
// one at a time. Errors are reported as ILLEGAL tokens with the
// offending lexeme; the parser surfaces them.
package lex

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gen0cide/westworld/dsl/token"
)

// Lexer holds the state for one source file.
type Lexer struct {
	filename string
	src      string

	// offset is the byte index of the next rune to read; pos is the
	// human-facing line/col of that rune.
	offset int
	line   int
	col    int

	// errors collected during lexing. They surface as ILLEGAL
	// tokens too, but accumulated here for batch reporting.
	errs []error

	// prevKind is the kind of the previously emitted token and sawGap
	// whether whitespace/comments separated it from the current one —
	// used to reject a gapless bang on a KEYWORD (`wait!`) while still
	// accepting prefix `!` as a C-style alias for `not`.
	prevKind token.Kind
	sawGap   bool

	// inFString is true while the lexer is between f"..." quotes.
	// inPlaceholder is true while inside a `{...}` interpolation;
	// during that window we lex normal tokens, and the next `}` we
	// see (at the top placeholder level) ends the placeholder and
	// returns us to f-string literal mode.
	// v1 doesn't allow nested f-strings inside placeholders.
	inFString     bool
	inPlaceholder bool
	fstringQuote  rune // quote char that closes the outer f-string
}

// New returns a Lexer over `src` named `filename` in diagnostics.
func New(filename, src string) *Lexer {
	return &Lexer{
		filename: filename,
		src:      src,
		offset:   0,
		line:     1,
		col:      1,
	}
}

// Errors returns any lex errors collected so far. Most callers can
// ignore this and just react to ILLEGAL tokens in the stream.
func (l *Lexer) Errors() []error { return l.errs }

// All returns every token from the source until EOF, including the
// trailing EOF token. Convenient for tests and for the parser which
// pre-buffers.
func (l *Lexer) All() []token.Token {
	var out []token.Token
	for {
		before := l.offset
		t := l.Next()
		out = append(out, t)
		if t.Kind == token.EOF {
			return out
		}
		// Spin-proofing: a non-EOF token that consumed no input means a
		// lexer-state bug is re-emitting the same token forever (offset is
		// monotonic, so any progress guarantees termination on finite src).
		// Refuse to loop: record it and end the stream.
		if l.offset == before {
			l.errs = append(l.errs, fmt.Errorf("%s: lexer made no progress at %q; aborting token stream", t.Pos, t.Lexeme))
			out = append(out, token.Token{Kind: token.EOF, Pos: t.Pos})
			return out
		}
	}
}

// Next returns the next token, advancing the position.
func (l *Lexer) Next() token.Token {
	tok := l.next()
	l.prevKind = tok.Kind
	return tok
}

func (l *Lexer) next() token.Token {
	// If we're between f-string quotes but NOT inside a `{...}`
	// placeholder, the next chunk of bytes is a literal fragment.
	if l.inFString && !l.inPlaceholder {
		return l.lexFStringPart()
	}
	l.sawGap = false
	pre := l.offset
	l.skipWhitespaceAndComments()
	l.sawGap = l.offset != pre
	// Inside a placeholder, we lex normal tokens — but the `}` that
	// closes the placeholder is intercepted to switch back to literal
	// mode rather than emitting an RBRACE token. Skipping whitespace
	// FIRST lets `{ x }` (spaces around the body) close cleanly; without
	// it, the space made the `}` lex as a stray RBRACE and the f-string
	// failed to parse.
	if l.inFString && l.inPlaceholder && l.peek() == '}' {
		l.advance() // consume the closing }
		l.inPlaceholder = false
		return l.lexFStringPart()
	}
	startPos := l.curPos()
	if l.atEOF() {
		return token.Token{Kind: token.EOF, Pos: startPos}
	}
	r := l.peek()

	// f-string opener must be checked before the identifier path
	// because 'f' is also a valid identifier start.
	if r == 'f' && l.peekAt(1) == '"' {
		return l.lexFStringOpen(startPos)
	}
	switch {
	case unicode.IsLetter(r) || r == '_':
		return l.lexIdentOrKeyword(startPos)
	case unicode.IsDigit(r):
		return l.lexNumber(startPos)
	case r == '"':
		return l.lexString(startPos, false)
	}
	return l.lexPunct(startPos)
}

// ----- core scanners -----

func (l *Lexer) lexIdentOrKeyword(start token.Position) token.Token {
	begin := l.offset
	for !l.atEOF() {
		r := l.peek()
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			l.advance()
			continue
		}
		break
	}
	// Bang suffix: a trailing `!` (NOT followed by `=`) is part of
	// the identifier — that's how the language spells "assert
	// success" action variants like `walk_to!` / `eat!`. If the
	// `!` is followed by `=`, it's the NEQ operator and stays its
	// own token. Keywords never take a bang (no `if!`); only the
	// IDENT branch picks it up.
	lex := l.src[begin:l.offset]
	if kw, ok := token.Keywords[lex]; ok {
		return token.Token{Kind: kw, Lexeme: lex, Pos: start}
	}
	if !l.atEOF() && l.peek() == '!' && l.peekAt(1) != '=' {
		l.advance()
		lex = l.src[begin:l.offset]
	}
	return token.Token{Kind: token.IDENT, Lexeme: lex, Pos: start}
}

func (l *Lexer) lexNumber(start token.Position) token.Token {
	begin := l.offset
	for !l.atEOF() && unicode.IsDigit(l.peek()) {
		l.advance()
	}
	// Float? Watch out for `..` (range) — don't consume the second dot.
	if !l.atEOF() && l.peek() == '.' && l.peekAt(1) != '.' && unicode.IsDigit(l.peekAt(1)) {
		l.advance() // consume '.'
		for !l.atEOF() && unicode.IsDigit(l.peek()) {
			l.advance()
		}
		return token.Token{Kind: token.FLOAT, Lexeme: l.src[begin:l.offset], Pos: start}
	}
	return token.Token{Kind: token.INT, Lexeme: l.src[begin:l.offset], Pos: start}
}

func (l *Lexer) lexString(start token.Position, _ bool) token.Token {
	l.advance() // consume opening "
	var sb strings.Builder
	for !l.atEOF() {
		r := l.peek()
		if r == '"' {
			l.advance() // consume closing "
			return token.Token{Kind: token.STRING, Lexeme: sb.String(), Pos: start}
		}
		if r == '\\' {
			l.advance()
			if l.atEOF() {
				return l.errorf(start, "unterminated escape sequence")
			}
			esc := l.peek()
			l.advance()
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case 'r':
				sb.WriteByte('\r')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			case '\'':
				sb.WriteByte('\'')
			case '0':
				sb.WriteByte(0)
			default:
				return l.errorf(start, "unknown escape sequence \\%c", esc)
			}
			continue
		}
		if r == '\n' {
			return l.errorf(start, "unterminated string literal (newline before closing quote)")
		}
		sb.WriteRune(r)
		l.advance()
	}
	return l.errorf(start, "unterminated string literal (reached EOF)")
}

// lexFStringOpen handles `f"..."` openings. Emits FSTRING_OPEN with
// the literal prefix (text up to the first `{` or closing quote).
// The lexer then enters fstring-body mode; subsequent calls to
// Next() return alternating embedded tokens + FSTRING_PART literals
// until FSTRING_CLOSE.
func (l *Lexer) lexFStringOpen(start token.Position) token.Token {
	l.advance() // consume 'f'
	q := l.peek()
	l.advance() // consume opening quote
	l.inFString = true
	l.fstringQuote = q
	return l.lexFStringLiteralFragment(start, token.FSTRING_OPEN)
}

// lexFStringPart returns the next literal fragment of an f-string —
// the text between two `{...}` placeholders, or between the last
// placeholder and the closing quote.
func (l *Lexer) lexFStringPart() token.Token {
	start := l.curPos()
	return l.lexFStringLiteralFragment(start, token.FSTRING_PART)
}

// lexFStringLiteralFragment is the shared implementation for the
// opening and middle fragments. It reads until it hits a `{`
// (start of placeholder) or the closing quote, emits the kind the
// caller passes, and transitions state appropriately.
func (l *Lexer) lexFStringLiteralFragment(start token.Position, kind token.Kind) token.Token {
	var sb strings.Builder
	for !l.atEOF() {
		r := l.peek()
		if r == '{' {
			// If `{{`, emit a literal `{`. Otherwise this is a
			// placeholder opening.
			if l.peekAt(1) == '{' {
				sb.WriteRune('{')
				l.advance()
				l.advance()
				continue
			}
			l.advance() // consume '{'; placeholder body follows
			l.inPlaceholder = true
			return token.Token{Kind: kind, Lexeme: sb.String(), Pos: start}
		}
		if r == l.fstringQuote {
			l.advance() // consume closing quote
			l.inFString = false
			return token.Token{Kind: token.FSTRING_CLOSE, Lexeme: sb.String(), Pos: start}
		}
		if r == '\\' {
			l.advance()
			if l.atEOF() {
				return l.errorf(start, "unterminated f-string escape")
			}
			esc := l.peek()
			l.advance()
			switch esc {
			case 'n':
				sb.WriteByte('\n')
			case 't':
				sb.WriteByte('\t')
			case '\\':
				sb.WriteByte('\\')
			case '"':
				sb.WriteByte('"')
			default:
				return l.errorf(start, "unknown f-string escape \\%c", esc)
			}
			continue
		}
		if r == '\n' {
			// Terminal error: MUST leave f-string mode. The newline is not
			// consumed, so a sticky inFString sends next() straight back here
			// at the same offset — an unbounded ILLEGAL stream that pins
			// All()'s slices live on the goroutine (the 2026-06-11 mesad OOM:
			// one LLM-authored broken f-string per wedged Act handler).
			l.inFString = false
			l.inPlaceholder = false
			return l.errorf(start, "unterminated f-string (newline before closing quote)")
		}
		sb.WriteRune(r)
		l.advance()
	}
	// Terminal error at EOF: same sticky-mode trap — clear it so the next
	// Next() emits EOF and All() terminates.
	l.inFString = false
	l.inPlaceholder = false
	return l.errorf(start, "unterminated f-string (reached EOF)")
}

func (l *Lexer) lexPunct(start token.Position) token.Token {
	r := l.peek()
	l.advance()
	switch r {
	case '(':
		return token.Token{Kind: token.LPAREN, Lexeme: "(", Pos: start}
	case ')':
		return token.Token{Kind: token.RPAREN, Lexeme: ")", Pos: start}
	case '{':
		return token.Token{Kind: token.LBRACE, Lexeme: "{", Pos: start}
	case '}':
		return token.Token{Kind: token.RBRACE, Lexeme: "}", Pos: start}
	case '[':
		return token.Token{Kind: token.LBRACK, Lexeme: "[", Pos: start}
	case ']':
		return token.Token{Kind: token.RBRACK, Lexeme: "]", Pos: start}
	case ',':
		return token.Token{Kind: token.COMMA, Lexeme: ",", Pos: start}
	case ';':
		return token.Token{Kind: token.SEMICOL, Lexeme: ";", Pos: start}
	case ':':
		return token.Token{Kind: token.COLON, Lexeme: ":", Pos: start}
	case '.':
		if l.peek() == '.' {
			l.advance()
			return token.Token{Kind: token.DOTDOT, Lexeme: "..", Pos: start}
		}
		return token.Token{Kind: token.DOT, Lexeme: ".", Pos: start}
	case '+':
		return token.Token{Kind: token.ADD, Lexeme: "+", Pos: start}
	case '-':
		if l.peek() == '>' {
			l.advance()
			return token.Token{Kind: token.ARROW, Lexeme: "->", Pos: start}
		}
		return token.Token{Kind: token.SUB, Lexeme: "-", Pos: start}
	case '*':
		return token.Token{Kind: token.MUL, Lexeme: "*", Pos: start}
	case '/':
		return token.Token{Kind: token.QUO, Lexeme: "/", Pos: start}
	case '%':
		return token.Token{Kind: token.REM, Lexeme: "%", Pos: start}
	case '=':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Kind: token.EQ, Lexeme: "==", Pos: start}
		}
		if l.peek() == '>' {
			l.advance()
			return token.Token{Kind: token.FATARROW, Lexeme: "=>", Pos: start}
		}
		return token.Token{Kind: token.ASSIGN, Lexeme: "=", Pos: start}
	case '!':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Kind: token.NEQ, Lexeme: "!=", Pos: start}
		}
		// A gapless `!` directly after a KEYWORD is a bang-variant
		// attempt (`wait!`) — keywords never take bangs; reject it like
		// the pre-alias lexer did rather than silently parsing `wait not(...)`.
		if !l.sawGap && keywordKinds[l.prevKind] {
			return l.errorf(start, "keywords don't take bang variants (saw %q followed by '!')", l.prevKind.String())
		}
		// C-style alias: the canonical spelling is `not`, but LLM authors
		// trained on C-family code emit `!` constantly — accept it rather
		// than fail the whole routine at parse time (Postel's law for a
		// machine-authored language). Same below for && and ||.
		return token.Token{Kind: token.NOT, Lexeme: "!", Pos: start}
	case '&':
		if l.peek() == '&' {
			l.advance()
			return token.Token{Kind: token.AND, Lexeme: "&&", Pos: start}
		}
		return l.errorf(start, "unexpected character %q (did you mean '&&' / 'and'?)", r)
	case '|':
		if l.peek() == '|' {
			l.advance()
			return token.Token{Kind: token.OR, Lexeme: "||", Pos: start}
		}
		return l.errorf(start, "unexpected character %q (did you mean '||' / 'or'?)", r)
	case '<':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Kind: token.LTE, Lexeme: "<=", Pos: start}
		}
		return token.Token{Kind: token.LT, Lexeme: "<", Pos: start}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return token.Token{Kind: token.GTE, Lexeme: ">=", Pos: start}
		}
		return token.Token{Kind: token.GT, Lexeme: ">", Pos: start}
	}
	return l.errorf(start, "unexpected character %q", r)
}

// keywordKinds is the reverse of token.Keywords: which Kinds are reserved
// words (they never take bang variants).
var keywordKinds = func() map[token.Kind]bool {
	m := make(map[token.Kind]bool, len(token.Keywords))
	for _, k := range token.Keywords {
		m[k] = true
	}
	return m
}()

// ----- helpers -----

func (l *Lexer) atEOF() bool { return l.offset >= len(l.src) }

func (l *Lexer) peek() rune {
	if l.atEOF() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.offset:])
	return r
}

func (l *Lexer) peekAt(n int) rune {
	off := l.offset
	for i := 0; i < n; i++ {
		if off >= len(l.src) {
			return 0
		}
		_, sz := utf8.DecodeRuneInString(l.src[off:])
		off += sz
	}
	if off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[off:])
	return r
}

func (l *Lexer) advance() {
	if l.atEOF() {
		return
	}
	r, sz := utf8.DecodeRuneInString(l.src[l.offset:])
	l.offset += sz
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
}

func (l *Lexer) curPos() token.Position {
	return token.Position{
		Filename: l.filename,
		Line:     l.line,
		Column:   l.col,
		Offset:   l.offset,
	}
}

func (l *Lexer) skipWhitespaceAndComments() {
	for !l.atEOF() {
		r := l.peek()
		if r == '#' {
			// Eat to end of line.
			for !l.atEOF() && l.peek() != '\n' {
				l.advance()
			}
			continue
		}
		if unicode.IsSpace(r) {
			l.advance()
			continue
		}
		return
	}
}

func (l *Lexer) errorf(pos token.Position, format string, args ...any) token.Token {
	msg := fmt.Sprintf(format, args...)
	l.errs = append(l.errs, fmt.Errorf("%s: %s", pos, msg))
	return token.Token{Kind: token.ILLEGAL, Lexeme: msg, Pos: pos}
}
