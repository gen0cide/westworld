package runtime

import (
	"testing"

	"github.com/gen0cide/westworld/hostkv"
	"github.com/gen0cide/westworld/memory"
)

// TestRoutineLibraryPersists proves the learned library survives a restart: a
// fresh library sharing the same durable store reloads the promoted routines, so
// a warmed host stays warm (no re-paying the LLM warmup each run).
func TestRoutineLibraryPersists(t *testing.T) {
	store := hostkv.NewMemory()

	lib := NewRoutineLibrary(memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store}))
	lib.Promote("sig-A", "mine_tin", "runtime \"1.0\"\nroutine mine_tin() { wait(1) }")
	lib.Promote("sig-B", "cook_meat", "runtime \"1.0\"\nroutine cook_meat() { wait(1) }")
	if lib.Len() != 2 {
		t.Fatalf("library size = %d, want 2", lib.Len())
	}

	// A fresh library over the same store restores both.
	restored := NewRoutineLibrary(memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store}))
	if restored.Len() != 2 {
		t.Fatalf("restored size = %d, want 2", restored.Len())
	}
	if e, ok := restored.Lookup("sig-A"); !ok || e.Name != "mine_tin" {
		t.Fatalf("sig-A not restored: %+v ok=%v", e, ok)
	}

	// Eviction also persists.
	restored.Evict("sig-A")
	again := NewRoutineLibrary(memory.New(memory.Options{Scratch: hostkv.NewScratch(16), Local: store}))
	if again.Len() != 1 {
		t.Fatalf("after evict, restored size = %d, want 1", again.Len())
	}
	if _, ok := again.Lookup("sig-A"); ok {
		t.Fatal("evicted routine should not reload")
	}
}
