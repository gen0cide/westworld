// actstats.go is the act-rejection telemetry (2026-06-11 OOM incident
// follow-up): cumulative counters of move rejections by class, plus the
// auto-repair count, surfaced on the mem-gauge log line (health.go) so the
// soak monitor reads them with zero new infra. 82% of authoring rejections in
// the incident were f-string shaped, so the same classification also drives
// the sharpened f-string re-prompt.
package mesad

import (
	"strings"
	"sync/atomic"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// Act-rejection classes — the vocabulary shared by classifyRejection, the
// counters, and the class-aware re-prompt sharpening.
const (
	rejectFStringParse = "fstring_parse" // parse failure that is f-string shaped (the incident class)
	rejectOtherParse   = "other_parse"   // any other dsl_source parse failure
	rejectValidation   = "validation"    // parses (or never needed to) but rejected: empty source/verb, unknown verb, bad args
	rejectNonAction    = "non_action"    // parses but takes no real game action
)

// actRejectStats holds the cumulative act-rejection counters. Process-lifetime
// package atomics (mesad runs one Server per process), read by logMemGauge —
// the soak monitor diffs successive mem-gauge ticks for rates.
type actRejectStats struct {
	fstringParse atomic.Uint64
	otherParse   atomic.Uint64
	validation   atomic.Uint64
	nonAction    atomic.Uint64
	autoRepaired atomic.Uint64
}

var actStats actRejectStats

// bump increments the counter for a classifyRejection class; an unknown class
// counts as validation (the broadest bucket) rather than dropping silently.
func (st *actRejectStats) bump(class string) {
	switch class {
	case rejectFStringParse:
		st.fstringParse.Add(1)
	case rejectOtherParse:
		st.otherParse.Add(1)
	case rejectNonAction:
		st.nonAction.Add(1)
	default:
		st.validation.Add(1)
	}
}

// classifyRejection buckets a validateMove rejection by its error text (after
// rejectionDetail enrichment — the parser's own text for a lex failure only
// says "got ILLEGAL", so the f-string signal rides the appended lexer
// message). The "does not parse" / "no real game action" markers are
// validateMove's own wording.
func classifyRejection(text string) string {
	switch {
	case strings.Contains(text, "does not parse"):
		if strings.Contains(text, "f-string") {
			return rejectFStringParse
		}
		return rejectOtherParse
	case strings.Contains(text, "no real game action"):
		return rejectNonAction
	default:
		return rejectValidation
	}
}

// rejectionDetail renders the rejection text the classifier reads and the
// model sees on re-prompt. For a write_routine parse failure it appends the
// lexer's own message (firstLexFailure) — the parse error alone says only
// "got ILLEGAL", which neither classifies nor tells the model what to fix.
func rejectionDetail(verr error, m *mesapb.Move) string {
	detail := verr.Error()
	if m.GetKind() == mesapb.MoveKind_WRITE_ROUTINE && strings.Contains(detail, "does not parse") {
		if msg, _, ok := firstLexFailure(m.GetDslSource()); ok && !strings.Contains(detail, msg) {
			detail += " — lexer: " + msg
		}
	}
	return detail
}

// fstringRecipe is the manual's f-string recipe (dslmanual.go), re-stated AT
// the failure: the retry lands on attempt 2 far more often when the fix is in
// the rejection itself rather than buried in the cached manual. Ordering
// mirrors the manual: format() first, then the f-string rules — including the
// plain-double-quotes rule, because a backslash-escaped quote inside an
// f-string is one of the incident shapes that lands here when auto-repair
// can't fix it.
const fstringRecipe = `Recipe for f-string errors: prefer format("template with {} slots", arg1, arg2) — each {} fills from the args left-to-right, keeping expressions OUT of the string literal — or bind a local first (msg = ...; say(f"{msg}")); exactly ONE expression per {} (write f"{a} {b}", never f"{a b}"); nested strings inside a placeholder use PLAIN double quotes — f"have {inventory.count("coins")} gp" — NEVER backslash-escaped quotes.`

// sharpenRejection appends the f-string recipe to f-string-shaped rejections;
// every other class passes through unchanged.
func sharpenRejection(class, reject string) string {
	if class != rejectFStringParse {
		return reject
	}
	return reject + "\n" + fstringRecipe
}
