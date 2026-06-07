package hostkv

import (
	"os"
	"path/filepath"
	"testing"
)

type relRow struct {
	Alpha float64  `json:"alpha"`
	Beta  float64  `json:"beta"`
	Tags  []string `json:"tags"`
}

func TestStoreSetGetTyped(t *testing.T) {
	s := NewMemory()
	want := relRow{Alpha: 3, Beta: 1, Tags: []string{"trusted"}}
	if err := Set(s, "ledger:alex", want); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok, err := Get[relRow](s, "ledger:alex")
	if err != nil || !ok {
		t.Fatalf("Get: ok=%v err=%v", ok, err)
	}
	if got.Alpha != want.Alpha || got.Beta != want.Beta || len(got.Tags) != 1 {
		t.Fatalf("roundtrip mismatch: got %+v want %+v", got, want)
	}
}

func TestStoreGetMissingIsCleanMiss(t *testing.T) {
	s := NewMemory()
	_, ok, err := Get[relRow](s, "nope")
	if ok || err != nil {
		t.Fatalf("missing key: ok=%v err=%v, want false/nil", ok, err)
	}
}

func TestStoreKeysPrefix(t *testing.T) {
	s := NewMemory()
	_ = Set(s, "ledger:alex", relRow{})
	_ = Set(s, "ledger:bob", relRow{})
	_ = Set(s, "goal:current", "mine")
	got := s.Keys("ledger:")
	if len(got) != 2 || got[0] != "ledger:alex" || got[1] != "ledger:bob" {
		t.Fatalf("Keys(ledger:) = %v", got)
	}
	if all := s.Keys(""); len(all) != 3 {
		t.Fatalf("Keys() len = %d, want 3", len(all))
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewMemory()
	_ = Set(s, "k", 1)
	if err := s.Delete("k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, ok := s.GetRaw("k"); ok {
		t.Fatal("key still present after Delete")
	}
	if err := s.Delete("absent"); err != nil {
		t.Fatalf("Delete absent: %v", err)
	}
}

func TestStorePersistAndReload(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub", "ledger.db") // parent dir must be created on open
	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	if err := Set(s, "ledger:alex", relRow{Alpha: 5, Beta: 2}); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if err := s.Close(); err != nil { // release the lock before reopening
		t.Fatalf("Close: %v", err)
	}

	reopened, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer reopened.Close()
	got, ok, err := Get[relRow](reopened, "ledger:alex")
	if err != nil || !ok || got.Alpha != 5 || got.Beta != 2 {
		t.Fatalf("after reload: got=%+v ok=%v err=%v", got, ok, err)
	}
}

func TestOpenMissingFileIsEmpty(t *testing.T) {
	s, err := Open(filepath.Join(t.TempDir(), "does-not-exist.db"))
	if err != nil {
		t.Fatalf("Open missing: %v", err)
	}
	defer s.Close()
	if s.Len() != 0 {
		t.Fatalf("Len = %d, want 0", s.Len())
	}
}

func TestOpenNonBoltFileIsError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "garbage.db")
	if err := os.WriteFile(path, []byte("this is not a bolt database"), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := Open(path)
	if err == nil {
		s.Close()
		t.Fatal("Open non-bolt file: want error, got nil")
	}
}
