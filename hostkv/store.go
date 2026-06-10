package hostkv

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	pebble "github.com/cockroachdb/pebble"
	bolt "go.etcd.io/bbolt"
)

// storeFileVersion lets us migrate the on-disk format later without guessing.
// It is written under the meta key on first open and checked on reopen.
const storeFileVersion = 1

const (
	metaVersionKey = "\x00meta:version" // \x00 prefix keeps meta out of Keys("")
)

// legacy bbolt layout (read-only, for one-time migration).
var (
	legacyKVBucket   = []byte("kv")
	legacyMetaBucket = []byte("meta")
	legacyVersionKey = []byte("version")
)

// Store is a host's durable key/value table.
//
// The durable form (Open) is backed by a Pebble directory: an LSM store whose
// writes are ~1µs against bbolt's ~8ms fsync'd B+tree commits — the host
// workload (chain-of-thought, ledgers, goal graphs) is write-heavy, and a
// fleet of stores must stay cheap per instance (~0.2MB heap at rest vs
// badger's 64MB arena). Writes ride the WAL without per-write fsync: a crash
// can lose the final few writes, which is acceptable for recoverable host
// state (decisions.jsonl is the per-write-flushed audit trail). The previous
// backend was a single bbolt file; Open transparently migrates a legacy file
// the first time it sees one, preserving the original under .pre-pebble. The
// in-memory form (NewMemory) is a plain map with no disk backing, for tests
// and hosts that must not persist.
//
// Keys are flat strings; callers namespace by convention with a prefix and a
// separator ("ledger:alex", "goal:current") and enumerate a namespace with
// Keys(prefix), which returns keys byte-sorted. Values are raw JSON — use the
// package-level generic Get/Set helpers for typed access. Safe for concurrent
// use; Close releases the store.
type Store struct {
	pdb *pebble.DB // durable backend; nil ⇒ in-memory

	// in-memory backend (used only when pdb == nil)
	mu   sync.RWMutex
	data map[string]json.RawMessage
}

// quietLogger drops pebble's operational chatter (compaction/WAL notices);
// real failures still surface as errors from the calling operation.
type quietLogger struct{}

func (quietLogger) Infof(string, ...any)  {}
func (quietLogger) Errorf(string, ...any) {}
func (quietLogger) Fatalf(format string, args ...any) {
	panic(fmt.Sprintf("hostkv: pebble fatal: "+format, args...))
}

// Open opens (or creates) the durable store for the legacy path. The path is
// the historical bbolt FILE path (<dir>/<user>.db); the Pebble directory
// lives beside it at <path>.pebble. A legacy bbolt file with no Pebble dir is
// migrated key-for-key on first open and the original preserved at
// <path>.pre-pebble — a host's learned state survives the engine swap. The
// caller owns the returned Store and must Close it.
func Open(path string) (*Store, error) {
	dir := path + ".pebble"
	if parent := filepath.Dir(path); parent != "" {
		if err := os.MkdirAll(parent, 0o755); err != nil {
			return nil, fmt.Errorf("hostkv: create data dir: %w", err)
		}
	}
	needMigrate := false
	if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
		if _, lerr := os.Stat(path); lerr == nil {
			needMigrate = true
		}
	}

	pdb, err := pebble.Open(dir, &pebble.Options{Logger: quietLogger{}})
	if err != nil {
		return nil, fmt.Errorf("hostkv: open %s: %w", dir, err)
	}
	s := &Store{pdb: pdb}
	if err := s.init(); err != nil {
		_ = pdb.Close()
		return nil, err
	}
	if needMigrate {
		if err := s.migrateFromBolt(path); err != nil {
			_ = pdb.Close()
			return nil, fmt.Errorf("hostkv: migrate legacy store %s: %w", path, err)
		}
	}
	return s, nil
}

// init validates / stamps the format version.
func (s *Store) init() error {
	v, closer, err := s.pdb.Get([]byte(metaVersionKey))
	if errors.Is(err, pebble.ErrNotFound) {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, storeFileVersion)
		return s.pdb.Set([]byte(metaVersionKey), buf, pebble.NoSync)
	}
	if err != nil {
		return err
	}
	got := binary.BigEndian.Uint64(v)
	_ = closer.Close()
	if got != storeFileVersion {
		return fmt.Errorf("hostkv: store format version %d, want %d", got, storeFileVersion)
	}
	return nil
}

// migrateFromBolt copies every kv-bucket entry of a legacy bbolt file into
// the (already open) Pebble store, then renames the file to .pre-pebble so
// the migration runs exactly once and the original survives as a backup.
func (s *Store) migrateFromBolt(path string) error {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 5 * time.Second, ReadOnly: true})
	if err != nil {
		return fmt.Errorf("open legacy: %w", err)
	}
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(legacyKVBucket)
		if b == nil {
			return nil // empty/fresh legacy file — nothing to copy
		}
		// Validate the legacy version so we never misread durable state.
		if meta := tx.Bucket(legacyMetaBucket); meta != nil {
			if cur := meta.Get(legacyVersionKey); cur != nil {
				if got := binary.BigEndian.Uint64(cur); got != storeFileVersion {
					return fmt.Errorf("legacy store format version %d, want %d", got, storeFileVersion)
				}
			}
		}
		return b.ForEach(func(k, v []byte) error {
			return s.SetRaw(string(k), append(json.RawMessage(nil), v...))
		})
	})
	cerr := db.Close()
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	if err := os.Rename(path, path+".pre-pebble"); err != nil {
		return fmt.Errorf("preserve legacy file: %w", err)
	}
	return nil
}

// NewMemory returns an in-memory store with no disk backing. Close is a no-op.
func NewMemory() *Store {
	return &Store{data: map[string]json.RawMessage{}}
}

// Close releases the underlying store. Safe to call on an in-memory store.
func (s *Store) Close() error {
	if s.pdb != nil {
		return s.pdb.Close()
	}
	return nil
}

// GetRaw returns the raw JSON stored at key. The bool reports presence. The
// returned slice is a fresh copy safe to retain.
func (s *Store) GetRaw(key string) (json.RawMessage, bool) {
	if s.pdb == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		v, ok := s.data[key]
		return v, ok
	}
	v, closer, err := s.pdb.Get([]byte(key))
	if err != nil {
		return nil, false
	}
	out := append(json.RawMessage(nil), v...)
	_ = closer.Close()
	return out, true
}

// SetRaw stores raw JSON at key — one committed write.
func (s *Store) SetRaw(key string, v json.RawMessage) error {
	if s.pdb == nil {
		s.mu.Lock()
		s.data[key] = v
		s.mu.Unlock()
		return nil
	}
	return s.pdb.Set([]byte(key), v, pebble.NoSync)
}

// Delete removes key (no-op if absent).
func (s *Store) Delete(key string) error {
	if s.pdb == nil {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return nil
	}
	return s.pdb.Delete([]byte(key), pebble.NoSync)
}

// prefixUpperBound returns the exclusive iteration bound for a key prefix:
// the prefix with its last byte incremented (carrying through 0xff), or nil
// for an unbounded scan.
func prefixUpperBound(prefix string) []byte {
	b := []byte(prefix)
	for i := len(b) - 1; i >= 0; i-- {
		if b[i] != 0xff {
			b[i]++
			return b[:i+1]
		}
	}
	return nil
}

// Keys returns all keys with the given prefix, sorted ascending. An empty
// prefix returns every key. The returned slice is a fresh copy.
func (s *Store) Keys(prefix string) []string {
	if s.pdb == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		out := make([]string, 0, len(s.data))
		for k := range s.data {
			if prefix == "" || strings.HasPrefix(k, prefix) {
				out = append(out, k)
			}
		}
		sort.Strings(out)
		return out
	}
	var bounds *pebble.IterOptions
	if prefix != "" {
		bounds = &pebble.IterOptions{
			LowerBound: []byte(prefix),
			UpperBound: prefixUpperBound(prefix),
		}
	}
	it, err := s.pdb.NewIter(bounds)
	if err != nil {
		return nil
	}
	defer it.Close()
	var out []string
	for it.First(); it.Valid(); it.Next() {
		k := string(it.Key())
		if strings.HasPrefix(k, "\x00") {
			continue // meta keys stay invisible to callers
		}
		out = append(out, k)
	}
	return out
}

// Len reports the number of stored entries.
func (s *Store) Len() int {
	if s.pdb == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return len(s.data)
	}
	return len(s.Keys(""))
}

// Get reads key and unmarshals it into a T. ok is false when the key is
// absent (err is nil in that case); err is non-nil only on a JSON decode
// failure. The zero T is returned on miss or error.
func Get[T any](s *Store, key string) (val T, ok bool, err error) {
	raw, present := s.GetRaw(key)
	if !present {
		return val, false, nil
	}
	if e := json.Unmarshal(raw, &val); e != nil {
		return val, false, fmt.Errorf("hostkv: decode %q: %w", key, e)
	}
	return val, true, nil
}

// Set marshals val to JSON and stores it at key. A marshal failure returns an
// error without mutating the store.
func Set[T any](s *Store, key string, val T) error {
	raw, err := json.Marshal(val)
	if err != nil {
		return fmt.Errorf("hostkv: encode %q: %w", key, err)
	}
	return s.SetRaw(key, raw)
}
