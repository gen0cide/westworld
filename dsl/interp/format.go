package interp

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/spec"
	"github.com/gen0cide/westworld/dsl/token"
)

// format(template, args...) — positional string templating.
//
// Rationale: LLM authors keep writing str.format-style {} out of
// Python muscle memory — 82% of authoring rejections were f-string
// shaped, dominated by empty-{} placeholders. format() makes that
// instinct legal and moves expressions out of string literals into
// argument position.
//
// Syntax (defined once in dsl/spec/format.go, shared with the
// validator's static checks): {} is a positional placeholder consumed
// left-to-right; {{ and }} render literal braces; anything else —
// including {name} — is plain text.
//
// format is language-level, like f-string interpolation: evalIdent
// resolves it intrinsically, AHEAD of Builtins, so the bridge's
// spec-driven registration (runtime/dsl_actions.go::dslFormat, kept
// for the spec↔handler consistency gate) is always shadowed and the
// budget-enforcing callable below is what routines actually run.

// Format renders template with args, using the same value→string
// conversion f-string interpolation uses (Value.Display — see
// evalFString). On a placeholder/arg-count mismatch it returns a
// typed *Error with code FORMAT_MISMATCH and renders nothing.
// Exported so host-side callers (e.g. a runtime action handler that
// wants to delegate) stay behavior-identical by construction.
//
// CAP ASYMMETRY: Format applies NO string-length cap — only the
// interpreter's intrinsic callable enforces MaxStringLen (per
// appended part, exactly like evalFString). A host-side delegator
// that can receive hostile arg lists must bring its own cap.
func Format(template string, args []Value) (String, *Error) {
	out, ferr, _ := formatRender(template, args, nil)
	return out, ferr
}

// formatRender is the one render loop behind both the exported
// Format and the interpreter's format callable. checkPart (nil =
// uncapped) sees the running output length after every appended part
// — raw run, decoded escape, or rendered arg — mirroring
// evalFString's per-part budget discipline: a non-nil return stops
// all further appends, so the transient build is bounded at ~cap +
// one part rather than the sum of every arg's Display.
func formatRender(template string, args []Value, checkPart func(length int) *RuntimeError) (String, *Error, *RuntimeError) {
	if n := spec.CountFormatPlaceholders(template); n != len(args) {
		return "", NewError(FORMAT_MISMATCH,
			fmt.Sprintf("template has %d placeholders, got %d args", n, len(args))), nil
	}
	var sb strings.Builder
	var lenErr *RuntimeError
	check := func() {
		if lenErr == nil && checkPart != nil {
			lenErr = checkPart(sb.Len())
		}
	}
	next := 0
	spec.WalkFormatTemplate(template,
		func(lit string) {
			if lenErr != nil {
				return
			}
			sb.WriteString(lit)
			check()
		},
		func(b byte) {
			if lenErr != nil {
				return
			}
			sb.WriteByte(b)
			check()
		},
		func() {
			if lenErr != nil {
				return
			}
			sb.WriteString(args[next].Display())
			next++
			check()
		},
	)
	if lenErr != nil {
		return "", nil, lenErr
	}
	return String(sb.String()), nil, nil
}

// formatCallable is the Value form of the format builtin. It carries
// the interpreter only for the string-length budget (same cap
// f-string interpolation enforces).
type formatCallable struct{ interp *Interpreter }

func (c *formatCallable) Kind() string    { return "builtin" }
func (c *formatCallable) Display() string { return "<format>" }

func (c *formatCallable) Call(args []Value, named map[string]Value) (Value, error) {
	if len(named) > 0 {
		return nil, fmt.Errorf("format takes positional args only — placeholders are {} consumed left-to-right")
	}
	if len(args) == 0 {
		return nil, fmt.Errorf("format takes at least 1 arg (the template)")
	}
	tmpl, ok := args[0].(String)
	if !ok {
		return nil, fmt.Errorf("format: template must be a string, got %s", args[0].Kind())
	}
	out, ferr, lenErr := formatRender(string(tmpl), args[1:], func(length int) *RuntimeError {
		return c.interp.budget.checkStringLen(token.Position{}, length)
	})
	if ferr != nil {
		// Typed failure, delivered the same way bang-form action
		// failures are: abort with the *Error as the reason. format
		// returns a bare String on success — there is no CallResult
		// shell to carry .err — so the abort channel is what keeps
		// FORMAT_MISMATCH branchable (try { } recover err { err.code }).
		panic(abortSignal{Reason: ferr})
	}
	if lenErr != nil {
		// Same termination f-string interpolation uses on cap breach.
		panic(abortSignal{Reason: String(lenErr.Msg)})
	}
	return out, nil
}
