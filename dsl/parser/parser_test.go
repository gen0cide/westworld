package parser_test

import (
	"testing"

	"github.com/gen0cide/westworld/dsl/parser"
)

func TestParseSkeleton(t *testing.T) {
	src := `
on chat_received(speaker, message) {
    # body not parsed yet
}

proc helper() {
}

routine fish_at_swamp(rod_type) {
}
`
	file, err := parser.Parse("t.routine", src)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(file.Handlers) != 1 {
		t.Errorf("handlers: got %d, want 1", len(file.Handlers))
	}
	if len(file.Procs) != 1 {
		t.Errorf("procs: got %d, want 1", len(file.Procs))
	}
	if file.Routine == nil {
		t.Fatal("routine: got nil")
	}
	if file.Routine.Name != "fish_at_swamp" {
		t.Errorf("routine name: got %q, want fish_at_swamp", file.Routine.Name)
	}
	if len(file.Routine.Params) != 1 || file.Routine.Params[0].Name != "rod_type" {
		t.Errorf("routine params: got %v, want [rod_type]", file.Routine.Params)
	}
	if file.Handlers[0].Event != "chat_received" {
		t.Errorf("handler event: got %q, want chat_received", file.Handlers[0].Event)
	}
}

func TestDuplicateRoutineRejected(t *testing.T) {
	src := `routine a() {}
routine b() {}`
	_, err := parser.Parse("t.routine", src)
	if err == nil {
		t.Fatal("expected error for two routines, got nil")
	}
}
