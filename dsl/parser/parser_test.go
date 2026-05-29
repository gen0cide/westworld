package parser_test

import (
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/ast"
	"github.com/gen0cide/westworld/dsl/parser"
	"github.com/gen0cide/westworld/dsl/token"
)

// mustParse parses src and fails the test on any error.
func mustParse(t *testing.T, src string) *ast.File {
	t.Helper()
	file, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v\n--- src ---\n%s", err, src)
	}
	return file
}

func TestTopLevelSkeleton(t *testing.T) {
	src := `
on chat_received(speaker, message) {}
proc helper() {}
routine fish_at_swamp(rod_type) {}
`
	file := mustParse(t, src)
	if len(file.Handlers) != 1 {
		t.Errorf("handlers: got %d, want 1", len(file.Handlers))
	}
	if len(file.Procs) != 1 {
		t.Errorf("procs: got %d, want 1", len(file.Procs))
	}
	if file.Routine == nil || file.Routine.Name != "fish_at_swamp" {
		t.Fatal("expected routine named fish_at_swamp")
	}
	if len(file.Routine.Params) != 1 || file.Routine.Params[0].Name != "rod_type" {
		t.Errorf("routine params: got %v, want [rod_type]", file.Routine.Params)
	}
}

func TestDuplicateRoutineRejected(t *testing.T) {
	_, err := parser.Parse("t.routine", `routine a() {} routine b() {}`)
	if err == nil {
		t.Fatal("expected error for two routines")
	}
}

// ----- statements -----

func TestAssignStmt(t *testing.T) {
	src := `routine r() { x = 42 }`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	if len(stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(stmts))
	}
	a, ok := stmts[0].(*ast.AssignStmt)
	if !ok {
		t.Fatalf("expected *AssignStmt, got %T", stmts[0])
	}
	if a.Target != "x" {
		t.Errorf("target: got %q, want x", a.Target)
	}
	lit, ok := a.Value.(*ast.IntLit)
	if !ok || lit.Value != 42 {
		t.Errorf("value: got %v, want IntLit 42", a.Value)
	}
}

func TestIfElifElse(t *testing.T) {
	src := `routine r() {
		if x { y = 1 }
		elif y { y = 2 }
		else { y = 3 }
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	if len(stmts) != 1 {
		t.Fatalf("stmts: got %d, want 1", len(stmts))
	}
	ifs, ok := stmts[0].(*ast.IfStmt)
	if !ok {
		t.Fatalf("got %T, want *IfStmt", stmts[0])
	}
	if len(ifs.Elifs) != 1 {
		t.Errorf("elifs: got %d, want 1", len(ifs.Elifs))
	}
	if ifs.Else == nil {
		t.Error("expected else block")
	}
}

func TestWhileAndFor(t *testing.T) {
	src := `routine r() {
		while x > 0 { x = x - 1 }
		for t in targets { attack(t) }
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	if _, ok := stmts[0].(*ast.WhileStmt); !ok {
		t.Errorf("stmt[0]: got %T, want *WhileStmt", stmts[0])
	}
	fs, ok := stmts[1].(*ast.ForStmt)
	if !ok {
		t.Fatalf("stmt[1]: got %T, want *ForStmt", stmts[1])
	}
	if fs.Var != "t" {
		t.Errorf("for var: got %q, want t", fs.Var)
	}
}

func TestRepeatUntilParses(t *testing.T) {
	src := `routine r() {
		repeat {
			open_bank(banker)
			wait 1
		} until world.bank.is_open timeout 10
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	ru, ok := stmts[0].(*ast.RepeatUntilStmt)
	if !ok {
		t.Fatalf("stmt[0]: got %T, want *RepeatUntilStmt", stmts[0])
	}
	if ru.Body == nil || len(ru.Body.Stmts) != 2 {
		t.Errorf("body: expected 2 statements, got %d", len(ru.Body.Stmts))
	}
	if ru.Cond == nil {
		t.Error("cond: expected non-nil")
	}
	if ru.Timeout == nil {
		t.Error("timeout: expected non-nil")
	}
}

func TestRepeatUntilWithoutTimeoutParses(t *testing.T) {
	// Parser accepts the missing-timeout form; the validator
	// rejects it. Keeping concerns separated.
	src := `routine r() {
		repeat { x = 1 } until x == 1
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	ru, ok := stmts[0].(*ast.RepeatUntilStmt)
	if !ok {
		t.Fatalf("stmt[0]: got %T, want *RepeatUntilStmt", stmts[0])
	}
	if ru.Timeout != nil {
		t.Errorf("timeout: got %v, want nil", ru.Timeout)
	}
}

func TestReturnAbortBreakContinue(t *testing.T) {
	src := `routine r() {
		while true {
			if x { break }
			if y { continue }
			if z { return "done" }
			abort "panic"
		}
		return
	}`
	mustParse(t, src) // smoke test — every form parses without error
}

func TestRequireBlock(t *testing.T) {
	src := `routine r() {
		require {
			wielded != null
			fatigue < 90
		}
		say("ok")
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	rb, ok := stmts[0].(*ast.RequireBlock)
	if !ok {
		t.Fatalf("expected first stmt to be *RequireBlock, got %T", stmts[0])
	}
	if len(rb.Conds) != 2 {
		t.Errorf("conds: got %d, want 2", len(rb.Conds))
	}
}

func TestWaitFormsBothWork(t *testing.T) {
	src := `routine r() {
		wait 5
		wait(2.8..4.5)
	}`
	file := mustParse(t, src)
	stmts := file.Routine.Body.Stmts
	if len(stmts) != 2 {
		t.Fatalf("stmts: got %d, want 2", len(stmts))
	}
	w0 := stmts[0].(*ast.WaitStmt)
	if _, ok := w0.Duration.(*ast.IntLit); !ok {
		t.Errorf("wait 5: duration kind = %T, want *IntLit", w0.Duration)
	}
	w1 := stmts[1].(*ast.WaitStmt)
	if _, ok := w1.Duration.(*ast.RangeLit); !ok {
		t.Errorf("wait(2.8..4.5): duration kind = %T, want *RangeLit", w1.Duration)
	}
}

// ----- expressions -----

func TestArithmeticPrecedence(t *testing.T) {
	// 1 + 2 * 3 should parse as 1 + (2 * 3)
	file := mustParse(t, `routine r() { y = 1 + 2 * 3 }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	bin := a.Value.(*ast.BinaryExpr)
	if bin.Op != token.ADD {
		t.Errorf("top op: got %s, want +", bin.Op)
	}
	left := bin.Lhs.(*ast.IntLit)
	if left.Value != 1 {
		t.Errorf("lhs: got %d, want 1", left.Value)
	}
	right := bin.Rhs.(*ast.BinaryExpr)
	if right.Op != token.MUL {
		t.Errorf("rhs op: got %s, want *", right.Op)
	}
}

func TestComparisonAndLogical(t *testing.T) {
	// a == b and c != d or not e
	file := mustParse(t, `routine r() { x = a == b and c != d or not e }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	top, ok := a.Value.(*ast.BinaryExpr)
	if !ok || top.Op != token.OR {
		t.Fatalf("top op: got %v, want or", a.Value)
	}
	// LHS of or should be (a==b) AND (c!=d).
	andExpr, ok := top.Lhs.(*ast.BinaryExpr)
	if !ok || andExpr.Op != token.AND {
		t.Fatalf("expected AND under OR, got %T", top.Lhs)
	}
	if eq, ok := andExpr.Lhs.(*ast.BinaryExpr); !ok || eq.Op != token.EQ {
		t.Errorf("expected EQ under AND.Lhs")
	}
	if neq, ok := andExpr.Rhs.(*ast.BinaryExpr); !ok || neq.Op != token.NEQ {
		t.Errorf("expected NEQ under AND.Rhs")
	}
	// RHS of or should be `not e`.
	un, ok := top.Rhs.(*ast.UnaryExpr)
	if !ok || un.Op != token.NOT {
		t.Errorf("expected NOT on or.rhs, got %T", top.Rhs)
	}
}

func TestUnaryMinus(t *testing.T) {
	file := mustParse(t, `routine r() { x = -3 + 5 }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	bin := a.Value.(*ast.BinaryExpr)
	un := bin.Lhs.(*ast.UnaryExpr)
	if un.Op != token.SUB {
		t.Errorf("unary op: got %s, want -", un.Op)
	}
	if lit, ok := un.Rhs.(*ast.IntLit); !ok || lit.Value != 3 {
		t.Errorf("unary operand: got %v", un.Rhs)
	}
}

func TestCallWithNamedArgs(t *testing.T) {
	file := mustParse(t, `routine r() { walk_to(x = 120, y = 504) }`)
	es := file.Routine.Body.Stmts[0].(*ast.ExprStmt)
	call := es.X.(*ast.CallExpr)
	if len(call.Args) != 2 {
		t.Fatalf("args: got %d, want 2", len(call.Args))
	}
	if call.Args[0].Name != "x" || call.Args[1].Name != "y" {
		t.Errorf("named args: got %s,%s", call.Args[0].Name, call.Args[1].Name)
	}
}

func TestMemberAndIndex(t *testing.T) {
	file := mustParse(t, `routine r() { z = world.locs.banks[0].entrance }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	// Top-level Member: ?.entrance
	top, ok := a.Value.(*ast.MemberExpr)
	if !ok || top.Field != "entrance" {
		t.Fatalf("top: got %T (%v), want MemberExpr entrance", a.Value, a.Value)
	}
	// Under it: IndexExpr [0]
	idx, ok := top.Recv.(*ast.IndexExpr)
	if !ok {
		t.Fatalf("under .entrance: got %T, want IndexExpr", top.Recv)
	}
	if lit, ok := idx.Index.(*ast.IntLit); !ok || lit.Value != 0 {
		t.Errorf("index: got %v, want IntLit 0", idx.Index)
	}
	// Under index: Member .banks
	banks, ok := idx.Recv.(*ast.MemberExpr)
	if !ok || banks.Field != "banks" {
		t.Errorf("expected .banks: got %T", idx.Recv)
	}
}

func TestListLiteral(t *testing.T) {
	file := mustParse(t, `routine r() { x = [1, 2, 3] }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	list, ok := a.Value.(*ast.ListLit)
	if !ok {
		t.Fatalf("got %T, want *ListLit", a.Value)
	}
	if len(list.Elems) != 3 {
		t.Errorf("elems: got %d, want 3", len(list.Elems))
	}
}

func TestFStringParses(t *testing.T) {
	file := mustParse(t, `routine r() { say(f"hi {name} you have {gold} gp") }`)
	es := file.Routine.Body.Stmts[0].(*ast.ExprStmt)
	call := es.X.(*ast.CallExpr)
	fs, ok := call.Args[0].Value.(*ast.FStringLit)
	if !ok {
		t.Fatalf("arg: got %T, want *FStringLit", call.Args[0].Value)
	}
	// Parts: "hi ", name, " you have ", gold, " gp"
	if len(fs.Parts) != 5 {
		t.Fatalf("parts: got %d, want 5: %v", len(fs.Parts), fs.Parts)
	}
}

func TestRangeLit(t *testing.T) {
	file := mustParse(t, `routine r() { x = 2.8..4.5 }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	rng, ok := a.Value.(*ast.RangeLit)
	if !ok {
		t.Fatalf("got %T, want *RangeLit", a.Value)
	}
	if lo, ok := rng.Low.(*ast.FloatLit); !ok || lo.Value != 2.8 {
		t.Errorf("low: got %v", rng.Low)
	}
	if hi, ok := rng.High.(*ast.FloatLit); !ok || hi.Value != 4.5 {
		t.Errorf("high: got %v", rng.High)
	}
}

func TestParenthesizedExpr(t *testing.T) {
	// (1 + 2) * 3 vs 1 + 2 * 3
	file := mustParse(t, `routine r() { x = (1 + 2) * 3 }`)
	a := file.Routine.Body.Stmts[0].(*ast.AssignStmt)
	top := a.Value.(*ast.BinaryExpr)
	if top.Op != token.MUL {
		t.Errorf("top op: got %s, want *", top.Op)
	}
	left := top.Lhs.(*ast.BinaryExpr)
	if left.Op != token.ADD {
		t.Errorf("left op: got %s, want +", left.Op)
	}
}

// ----- realistic example -----

func TestSampleRoutineParsesEndToEnd(t *testing.T) {
	src := `
# fish_at_port_sarim.routine

on chat_received(speaker, message) {
    if message == "hi" {
        say(f"hey {speaker}")
    }
}

proc nearest_spot() {
    return world.locs.fishing_spots.nearest(self.position)
}

routine fish_at_port_sarim() {
    require {
        wielded != null
        inventory.free > 0
    }

    spot = nearest_spot()
    if spot == null {
        abort "no_spot"
    }

    walk_to(x = spot.x, y = spot.y)

    while inventory.free > 0 {
        if fatigue > 90 {
            return "tired"
        }
        fish(spot)
        wait 2.8..4.5
    }

    return "banked"
}
`
	file := mustParse(t, src)
	if file.Routine == nil {
		t.Fatal("expected routine")
	}
	if len(file.Handlers) != 1 {
		t.Errorf("handlers: got %d, want 1", len(file.Handlers))
	}
	if len(file.Procs) != 1 {
		t.Errorf("procs: got %d, want 1", len(file.Procs))
	}
}

func TestErrorMessagesIncludeSourcePosition(t *testing.T) {
	src := `routine r() { x = 1 + }`
	_, err := parser.Parse("err.routine", src)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "err.routine:") {
		t.Errorf("error lacks filename:line:col: %v", err)
	}
}
