package validator

// Static argument-literal checking against world catalogs.
//
// The Act planner sometimes emits well-formed DSL whose ARGUMENTS are
// hallucinated — go_to("mining-site") for a POI type that doesn't exist
// near the host, eat("typo-item"), talk_to("Nonexistent NPC"). These
// parse fine and pass arity checks, then Fail at runtime, wasting a host
// round-trip. CheckArgLiterals catches the statically-checkable subset:
// a compile-time STRING LITERAL passed to a parameter that the spec tags
// with a world catalog (spec.ParamKinds). Anything dynamic — a variable,
// an f-string, world.npcs.find(...), nearest_npc(), coordinates — is
// SKIPPED, because only literals can be checked without running the
// routine.
//
// The check mirrors runtime resolution exactly (see runtime/actions_
// ambient.go dslGoTo and resolveItemID) so it never rejects a value the
// engine would have accepted: place_or_poi uses case-insensitive
// SUBSTRING (PlaceByName then NearestPOI), item uses case-insensitive
// EXACT then SUBSTRING (resolveItemID). NPC args are situation-dependent (talk_to auto-
// targets the nearest visible NPC of a name), so the npc catalog is SOFT
// and not hard-rejected here.
//
// The validator stays facts-free: the catalog is INJECTED through the
// Catalog interface, whose facts-backed implementation lives in mesad.

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/spec"
)

// Catalog answers "is this literal string a real X in the world?" for the
// catalogued parameter kinds. Implementations mirror runtime resolution
// (substring for place/POI, exact for item). A nil Catalog disables the
// check (CheckArgLiterals returns no errors).
type Catalog interface {
	// KnownPlaceOrPOI reports whether s resolves to a known town/landmark
	// name OR a POI type. Case-insensitive SUBSTRING, like the gazetteer.
	KnownPlaceOrPOI(s string) bool
	// KnownItem reports whether s is a real item name. Case-insensitive
	// EXACT, like Facts.ItemDefByName.
	KnownItem(s string) bool
	// Examples returns a few real, valid sample values for the given
	// catalog kind, used to make the rejection message self-correcting.
	Examples(kind string) []string
}

// CheckArgLiterals walks every call in file and checks each positional
// string-LITERAL argument against the catalog kind its parameter is
// tagged with in the spec. Returns one error per offending literal (bad
// value + the verb/param + the catalog + a few valid examples). Returns
// nil if cat is nil or no literal is out of catalog. Never panics.
//
// NPC-typed params are NOT rejected here (situation-dependent); the npc
// kind is left to the prompt/manual to constrain.
func CheckArgLiterals(file *ast.File, cat Catalog) []error {
	if file == nil || cat == nil {
		return nil
	}
	w := &argWalker{cat: cat}
	if file.Routine != nil {
		w.walkNode(file.Routine)
	}
	for _, p := range file.Procs {
		w.walkNode(p)
	}
	for _, h := range file.Handlers {
		w.walkNode(h)
	}
	for _, b := range file.Bounds {
		w.walkNode(b)
	}
	return w.errs
}

type argWalker struct {
	cat  Catalog
	errs []error
}

// checkCall inspects one call's positional literal args.
func (w *argWalker) checkCall(c *ast.CallExpr) {
	id, ok := c.Callee.(*ast.Ident)
	if !ok {
		return // member/method call — not a catalogued builtin
	}
	base, _ := spec.StripBang(id.Name) // go_to! → go_to
	a, ok := spec.ByName(base)
	if !ok || a.ParamKinds == nil {
		return
	}
	pos := 0 // positional index among UN-named args
	for _, arg := range c.Args {
		if arg.Name != "" {
			continue // named arg (x=216) — not a catalogued positional here
		}
		kind := a.ParamKind(pos)
		pos++
		if kind == spec.CatalogNone || kind == spec.CatalogNPC {
			continue // uncatalogued, or soft npc (not hard-rejected)
		}
		lit, ok := arg.Value.(*ast.StringLit)
		if !ok {
			continue // dynamic (variable, f-string, view, coords, call) — skip
		}
		w.checkLiteral(base, a, pos-1, kind, lit)
	}
}

func (w *argWalker) checkLiteral(verb string, a *spec.ActionSpec, paramIdx int, kind string, lit *ast.StringLit) {
	val := lit.Value
	ok := true
	switch kind {
	case spec.CatalogPlaceOrPOI:
		ok = w.cat.KnownPlaceOrPOI(val)
	case spec.CatalogItem:
		ok = w.cat.KnownItem(val)
	default:
		return
	}
	if ok {
		return
	}
	param := "arg"
	if paramIdx < len(a.Params) {
		param = a.Params[paramIdx]
	}
	w.errs = append(w.errs, &ArgError{
		Pos:     lit.Pos().String(),
		Verb:    verb,
		Param:   param,
		Kind:    kind,
		Value:   val,
		Samples: w.cat.Examples(kind),
	})
}

// ArgError is one rejected literal arg. Its message names the bad value,
// the verb+param, the expected catalog, and a few real valid examples,
// so the re-prompt both explains the miss and re-teaches valid forms.
type ArgError struct {
	Pos     string
	Verb    string
	Param   string
	Kind    string
	Value   string
	Samples []string
}

func (e *ArgError) Error() string {
	var hint string
	switch e.Kind {
	case spec.CatalogPlaceOrPOI:
		hint = "%s(%q): %q is not a known place or POI type. %s takes coordinates (e.g. %s(120, 504)), a known TOWN name, or a POI TYPE%s — never a free description like \"the mine\"."
	case spec.CatalogItem:
		hint = "%s(%q): %q is not a real item. %s takes an inventory item by its exact name%s, or by slot=N — never a made-up name."
	default:
		hint = "%s(%q): %q is not in the %s catalog.%s"
	}
	ex := ""
	if len(e.Samples) > 0 {
		ex = " (e.g. " + quoteJoin(e.Samples) + ")"
	}
	switch e.Kind {
	case spec.CatalogPlaceOrPOI:
		return fmt.Sprintf(hint, e.Verb, e.Value, e.Value, e.Verb, e.Verb, ex)
	case spec.CatalogItem:
		return fmt.Sprintf(hint, e.Verb, e.Value, e.Value, e.Verb, ex)
	default:
		return fmt.Sprintf(hint, e.Verb, e.Value, e.Value, e.Kind, ex)
	}
}

func quoteJoin(xs []string) string {
	q := make([]string, len(xs))
	for i, x := range xs {
		q[i] = fmt.Sprintf("%q", x)
	}
	return strings.Join(q, ", ")
}

// ----- generic AST descent -----
//
// A small hand-rolled walker over every Stmt/Expr that can contain a
// CallExpr. It mirrors the node set in dsl/ast/ast.go; an unhandled node
// is simply a leaf (no calls beneath it), so the walk is total and
// never panics on a nil child.

func (w *argWalker) walkNode(n ast.Node) {
	switch x := n.(type) {
	// --- declarations ---
	case *ast.ProcDecl:
		w.walkParams(x.Params)
		w.walkBlock(x.Body)
	case *ast.RoutineDecl:
		w.walkParams(x.Params)
		if x.Require != nil {
			for _, c := range x.Require.Conds {
				w.walkExpr(c)
			}
		}
		for _, h := range x.Handlers {
			w.walkNode(h)
		}
		w.walkBlock(x.Body)
	case *ast.OnHandler:
		w.walkParams(x.Params)
		w.walkBlock(x.Body)
	case *ast.BoundsDecl:
		w.walkExpr(x.Shape)
		for _, h := range x.Handlers {
			w.walkNode(h)
		}
		for _, p := range x.Procs {
			w.walkNode(p)
		}
		for _, b := range x.Bounds {
			w.walkNode(b)
		}
	}
}

func (w *argWalker) walkParams(ps []*ast.Param) {
	for _, p := range ps {
		if p != nil {
			w.walkExpr(p.Default)
		}
	}
}

func (w *argWalker) walkBlock(b *ast.Block) {
	if b == nil {
		return
	}
	for _, s := range b.Stmts {
		w.walkStmt(s)
	}
}

func (w *argWalker) walkStmt(s ast.Stmt) {
	switch x := s.(type) {
	case *ast.Block:
		w.walkBlock(x)
	case *ast.AssignStmt:
		w.walkExpr(x.Value)
	case *ast.ExprStmt:
		w.walkExpr(x.X)
	case *ast.IfStmt:
		w.walkExpr(x.Cond)
		w.walkBlock(x.Then)
		for _, e := range x.Elifs {
			if e != nil {
				w.walkExpr(e.Cond)
				w.walkBlock(e.Body)
			}
		}
		w.walkBlock(x.Else)
	case *ast.WhileStmt:
		w.walkExpr(x.Cond)
		w.walkBlock(x.Body)
	case *ast.ForStmt:
		w.walkExpr(x.Iter)
		w.walkBlock(x.Body)
	case *ast.RepeatUntilStmt:
		w.walkBlock(x.Body)
		w.walkExpr(x.Cond)
		w.walkExpr(x.Timeout)
	case *ast.ReturnStmt:
		w.walkExpr(x.Value)
	case *ast.AbortStmt:
		w.walkExpr(x.Reason)
	case *ast.WaitStmt:
		w.walkExpr(x.Duration)
	case *ast.DeferStmt:
		w.walkExpr(x.Call)
	case *ast.TryStmt:
		w.walkBlock(x.Try)
		w.walkBlock(x.Recover)
	case *ast.WhenStmt:
		w.walkExpr(x.Predicate)
		w.walkBlock(x.Body)
	case *ast.SelectStmt:
		for _, c := range x.Cases {
			w.walkExpr(c.Predicate)
			w.walkBlock(c.Body)
		}
	case *ast.RequireBlock:
		for _, c := range x.Conds {
			w.walkExpr(c)
		}
	}
}

func (w *argWalker) walkExpr(e ast.Expr) {
	switch x := e.(type) {
	case nil:
		return
	case *ast.CallExpr:
		w.checkCall(x)
		w.walkExpr(x.Callee)
		for _, a := range x.Args {
			if a != nil {
				w.walkExpr(a.Value)
			}
		}
	case *ast.BinaryExpr:
		w.walkExpr(x.Lhs)
		w.walkExpr(x.Rhs)
	case *ast.UnaryExpr:
		w.walkExpr(x.Rhs)
	case *ast.MemberExpr:
		w.walkExpr(x.Recv)
	case *ast.IndexExpr:
		w.walkExpr(x.Recv)
		w.walkExpr(x.Index)
	case *ast.LambdaExpr:
		w.walkExpr(x.Body)
	case *ast.ListLit:
		for _, el := range x.Elems {
			w.walkExpr(el)
		}
	case *ast.RangeLit:
		w.walkExpr(x.Low)
		w.walkExpr(x.High)
	case *ast.FStringLit:
		for _, p := range x.Parts {
			w.walkExpr(p)
		}
	}
}
