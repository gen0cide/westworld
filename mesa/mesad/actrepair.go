// actrepair.go is the deterministic DSL pre-repair (2026-06-11 OOM incident
// follow-up): when a write_routine is rejected for a parse failure, the
// dominant shapes are mechanical typos the model makes from Python muscle
// memory — an unterminated string/f-string, a backslash-escaped quote inside
// an f-string. Each has exactly one mechanical fix, so the act loop applies it
// locally and re-validates instead of burning a re-prompt attempt (an LLM
// round-trip) on a typo. Acceptance is all-or-nothing: the repaired source
// must come back COMPLETELY clean from the full parse+validate, otherwise the
// original source (and its rejection) proceeds to the re-prompt unchanged.
package mesad

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gen0cide/westworld/dsl/lex"
	"github.com/gen0cide/westworld/dsl/token"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// Repair kinds, for the auto-repair log line.
const (
	repairCloseQuote    = "close_quote"    // inserted the missing closing " (unterminated string/f-string)
	repairFStringEscape = "fstring_escape" // stripped a backslash-escaped quote inside an f-string
)

// maxRepairPasses bounds the lex→repair loop: each pass fixes the first lexer
// failure, and a broken routine occasionally holds a couple of independent
// ones (two unterminated f-strings; an escaped quote AND a missing close
// quote on the same line). Anything needing more is not the mechanical-typo
// class this targets.
const maxRepairPasses = 3

// tryAutoRepair is the pre-repair seam in the act loop: when a write_routine
// was rejected for a PARSE failure, apply the mechanical fixes (repairDSL) and
// re-run the FULL parse+validate. On success the move's DslSource is swapped
// for the repaired source and the joined repair kinds are returned; on
// anything short of completely clean the move is left untouched so the normal
// re-prompt path sees the original source. Never calls the LLM — a repair try
// costs no attempt.
func tryAutoRepair(m *mesapb.Move, cat *argCatalog, verr error) (kind string, ok bool) {
	if m.GetKind() != mesapb.MoveKind_WRITE_ROUTINE || !strings.Contains(verr.Error(), "does not parse") {
		return "", false
	}
	fixed, kinds := repairDSL(m.GetDslSource())
	if len(kinds) == 0 || fixed == m.GetDslSource() {
		return "", false
	}
	orig := m.DslSource
	m.DslSource = fixed
	if _, err := validateMove(m, cat); err != nil {
		m.DslSource = orig
		return "", false
	}
	return strings.Join(kinds, "+"), true
}

// repairDSL applies the deterministic repairs to src, one lexer failure per
// pass, and returns the repaired source plus the repair kinds applied in
// order (empty → nothing repairable). Pure; the caller decides acceptance by
// re-running the full parse+validate.
func repairDSL(src string) (string, []string) {
	var kinds []string
	for pass := 0; pass < maxRepairPasses; pass++ {
		msg, pos, found := firstLexFailure(src)
		if !found {
			break
		}
		fixed, kind, ok := repairLexFailure(src, msg, pos)
		if !ok || fixed == src {
			break
		}
		src = fixed
		kinds = append(kinds, kind)
	}
	return src, kinds
}

// firstLexFailure re-lexes src and returns the first ILLEGAL token's message
// and position. The parser's error for a lex failure only says "got ILLEGAL";
// the precise failure shape — which drives both repair selection and the
// f-string rejection class — lives on the token itself. All() is spin-proofed
// (it refuses non-advancing iterations), so this always terminates.
func firstLexFailure(src string) (msg string, pos token.Position, ok bool) {
	for _, t := range lex.New("<act-move>", src).All() {
		if t.Kind == token.ILLEGAL {
			return t.Lexeme, t.Pos, true
		}
	}
	return "", token.Position{}, false
}

// unknownEscPrefix is the lexer's message for a bad escape inside an f-string
// (lex.lexFStringLiteralFragment); the offending character follows it.
const unknownEscPrefix = `unknown f-string escape \`

// repairLexFailure maps one lexer failure to its mechanical fix. Returns the
// repaired source + repair kind, or ok=false when the failure isn't one of
// the repairable shapes.
func repairLexFailure(src, msg string, pos token.Position) (string, string, bool) {
	switch {
	case strings.Contains(msg, "unterminated") && strings.Contains(msg, "newline before closing quote"):
		// The literal started at pos and cannot contain a raw newline, so the
		// offending newline is the first one at/after it — close the literal
		// immediately before it.
		if pos.Offset < 0 || pos.Offset > len(src) {
			return src, "", false
		}
		rel := strings.IndexByte(src[pos.Offset:], '\n')
		if rel < 0 {
			return src, "", false
		}
		at := pos.Offset + rel
		return src[:at] + `"` + src[at:], repairCloseQuote, true
	case strings.Contains(msg, "unterminated") && strings.Contains(msg, "reached EOF"):
		return src + `"`, repairCloseQuote, true
	case strings.HasPrefix(msg, unknownEscPrefix):
		// Only a backslash-escaped SINGLE quote has an unambiguous mechanical
		// fix (drop the backslash — \' decodes to ' in every context the
		// textual mirror below can misclassify, so the global rewrite is
		// value-preserving). \' is also the only shape the lexer can emit
		// today: \" is a LEGAL f-string escape (lex.go), so dispatching on it
		// here would be dead code — and dangerous dead code, because stripping
		// \" inside a PLAIN string the mirror misclassified as a fragment
		// changes the string's value while still re-parsing cleanly. If the
		// lexer ever rejects \" in f-strings, add it back ANCHORED to the
		// failing literal, not as a whole-source sweep. Any other unknown
		// escape's intent is opaque — no repair.
		esc, _ := utf8.DecodeRuneInString(msg[len(unknownEscPrefix):])
		if esc != '\'' {
			return src, "", false
		}
		fixed := stripFStringQuoteEscapes(src, byte(esc))
		return fixed, repairFStringEscape, fixed != src
	}
	return src, "", false
}

// stripFStringQuoteEscapes rewrites `\q` → `q` inside the LITERAL fragments of
// f-strings only — plain strings, placeholder bodies, and comments are copied
// verbatim (in a plain string \q may be a legal escape, and stripping \" there
// would terminate the string early). A textual mirror of the lexer's f-string
// state machine; a disagreement on a pathological source is harmless because
// the caller accepts a repair only when the full parse+validate comes back
// clean.
func stripFStringQuoteEscapes(src string, q byte) string {
	const (
		stCode    = iota // ordinary tokens (top level, or a placeholder body when inPH)
		stComment        // # to end of line
		stStr            // plain "..." string
		stFrag           // f-string literal fragment
	)
	var b strings.Builder
	b.Grow(len(src))
	st := stCode
	inPH := false // inside an f-string {placeholder}; the next bare '}' returns to stFrag
	for i := 0; i < len(src); i++ {
		c := src[i]
		switch st {
		case stCode:
			switch {
			case c == '#':
				st = stComment
			case c == 'f' && !inPH && i+1 < len(src) && src[i+1] == '"' && !identTailBefore(src, i):
				b.WriteString(`f"`)
				i++
				st = stFrag
				continue
			case c == '"':
				st = stStr
			case c == '}' && inPH:
				inPH = false
				st = stFrag
			}
			b.WriteByte(c)
		case stComment:
			if c == '\n' {
				st = stCode // a comment never ends a placeholder; inPH persists
			}
			b.WriteByte(c)
		case stStr:
			if c == '\\' && i+1 < len(src) {
				b.WriteByte(c)
				i++
				b.WriteByte(src[i])
				continue
			}
			if c == '"' || c == '\n' { // '\n' = unterminated; mirror the lexer's bail to normal mode
				st = stCode
			}
			b.WriteByte(c)
		case stFrag:
			switch {
			case c == '\\' && i+1 < len(src):
				if src[i+1] == q {
					b.WriteByte(q) // THE repair: drop the backslash, keep the quote
				} else {
					b.WriteByte(c)
					b.WriteByte(src[i+1])
				}
				i++
			case c == '{':
				if i+1 < len(src) && src[i+1] == '{' { // {{ is a literal brace, not a placeholder
					b.WriteString("{{")
					i++
					continue
				}
				b.WriteByte(c)
				inPH = true
				st = stCode
			case c == '"' || c == '\n': // close quote, or unterminated (mirror the lexer's bail)
				b.WriteByte(c)
				st = stCode
			default:
				b.WriteByte(c)
			}
		}
	}
	return b.String()
}

// identTailBefore reports whether the rune just before byte offset i continues
// an identifier — in which case an `f"` at i is the tail of that identifier
// (`staff"..."`), not an f-string opener.
func identTailBefore(src string, i int) bool {
	if i == 0 {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(src[:i])
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
