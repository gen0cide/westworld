package hostkv

import (
	"encoding/binary"
	"errors"
	"fmt"
	bolt "go.etcd.io/bbolt"
	"os"
	"path/filepath"
	"testing"
	"time"
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
	if _, err := os.Stat(path + ".pebble"); err != nil {
		t.Fatalf("store dir not written: %v", err)
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

// TestMigrateFromLegacyBolt: opening a path that holds a legacy bbolt file
// copies every key into the new Badger store exactly once and preserves the
// original at .pre-pebble.
func TestMigrateFromLegacyBolt(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/host.db"

	// Write a legacy bbolt store directly.
	bdb, err := bolt.Open(path, 0o600, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = bdb.Update(func(tx *bolt.Tx) error {
		kv, _ := tx.CreateBucketIfNotExists(legacyKVBucket)
		meta, _ := tx.CreateBucketIfNotExists(legacyMetaBucket)
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, storeFileVersion)
		_ = meta.Put(legacyVersionKey, buf)
		_ = kv.Put([]byte("ledger:alex"), []byte(`{"trust":1}`))
		return kv.Put([]byte("goal:current"), []byte(`"find a pickaxe"`))
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = bdb.Close()

	s, err := Open(path)
	if err != nil {
		t.Fatalf("open w/ migration: %v", err)
	}
	if v, ok := s.GetRaw("ledger:alex"); !ok || string(v) != `{"trust":1}` {
		t.Fatalf("migrated key wrong: %q ok=%v", v, ok)
	}
	if got := s.Len(); got != 2 {
		t.Fatalf("Len=%d want 2", got)
	}
	if _, err := os.Stat(path + ".pre-pebble"); err != nil {
		t.Fatalf("legacy file not preserved: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("legacy file should be renamed away")
	}

	// Re-open: no double-migration, data intact.
	_ = s.Close()
	s2, err := Open(path)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	defer s2.Close()
	if _, ok := s2.GetRaw("goal:current"); !ok {
		t.Fatal("data lost on reopen")
	}
}

// TestWriteAfterCloseNoPanic: pebble panics on write-after-close; the Store
// guard must turn that into ErrClosed. One host's teardown racing its own
// flush goroutines must never be able to kill the fleet process.
func TestWriteAfterCloseNoPanic(t *testing.T) {
	s, err := Open(t.TempDir() + "/x.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.SetRaw("k", []byte(`1`)); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRaw("k", []byte(`2`)); !errors.Is(err, ErrClosed) {
		t.Fatalf("SetRaw after Close: err=%v, want ErrClosed", err)
	}
	if err := s.Delete("k"); !errors.Is(err, ErrClosed) {
		t.Fatalf("Delete after Close: err=%v, want ErrClosed", err)
	}
	if _, ok := s.GetRaw("k"); ok {
		t.Fatal("GetRaw after Close reported presence")
	}
	if got := s.Keys(""); got != nil {
		t.Fatalf("Keys after Close = %v, want nil", got)
	}
	if err := s.Close(); err != nil {
		t.Fatalf("double Close: %v", err)
	}
}

// TestConcurrentWritesAndClose hammers the store from writer goroutines while
// Close fires mid-stream — the race the fleet hit live (runLimbic's final
// flush vs RunHost's deferred Close). Run under -race in CI.
func TestConcurrentWritesAndClose(t *testing.T) {
	s, err := Open(t.TempDir() + "/x.db")
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	for w := 0; w < 4; w++ {
		go func(w int) {
			defer func() { done <- struct{}{} }()
			for i := 0; ; i++ {
				if err := s.SetRaw(fmt.Sprintf("w%d:%d", w, i), []byte(`{"x":1}`)); err != nil {
					if !errors.Is(err, ErrClosed) {
						t.Errorf("writer %d: %v", w, err)
					}
					return
				}
			}
		}(w)
	}
	time.Sleep(20 * time.Millisecond)
	if err := s.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	for w := 0; w < 4; w++ {
		<-done
	}
}
