package hostkv

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	bolt "go.etcd.io/bbolt"
)

// storeFileVersion lets us migrate the on-disk format later without guessing.
// It is written into the meta bucket on first open and checked on reopen.
const storeFileVersion = 1

var (
	kvBucket   = []byte("kv")   // key → raw JSON value
	metaBucket = []byte("meta") // bookkeeping (format version)
	versionKey = []byte("version")
)

// Store is a host's durable key/value table.
//
// The durable form (Open) is backed by a single bbolt file: a B+tree with
// ACID, single-key writes — no whole-document rewrite per Set, unlike a flat
// JSON store. The in-memory form (NewMemory) is a plain map with no disk
// backing, for tests and hosts that must not persist.
//
// Keys are flat strings; callers namespace by convention with a prefix and a
// separator ("ledger:alex", "goal:current") and enumerate a namespace with
// Keys(prefix). bbolt stores keys in byte-sorted order, so Keys returns them
// sorted for free. Values are raw JSON — use the package-level generic Get/Set
// helpers for typed access. Safe for concurrent use; Close releases the file.
type Store struct {
	db *bolt.DB // durable backend; nil ⇒ in-memory

	// in-memory backend (used only when db == nil)
	mu   sync.RWMutex
	data map[string]json.RawMessage
}

// Open opens (or creates) a bbolt-backed store at path. Parent directories are
// created as needed. A missing file yields a fresh, empty store. A file whose
// format version does not match is an error, so a host never silently
// misreads durable state. The caller owns the returned Store and must Close it.
func Open(path string) (*Store, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("hostkv: create data dir: %w", err)
		}
	}
	// Timeout avoids hanging forever if another process holds the file lock;
	// surfacing it as an error is friendlier than a silent block.
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("hostkv: open %s: %w", path, err)
	}
	s := &Store{db: db}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// init creates the buckets and validates / stamps the format version.
func (s *Store) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(kvBucket); err != nil {
			return err
		}
		meta, err := tx.CreateBucketIfNotExists(metaBucket)
		if err != nil {
			return err
		}
		cur := meta.Get(versionKey)
		if cur == nil {
			buf := make([]byte, 8)
			binary.BigEndian.PutUint64(buf, storeFileVersion)
			return meta.Put(versionKey, buf)
		}
		if got := binary.BigEndian.Uint64(cur); got != storeFileVersion {
			return fmt.Errorf("hostkv: store format version %d, want %d", got, storeFileVersion)
		}
		return nil
	})
}

// NewMemory returns an in-memory store with no disk backing. Close is a no-op.
func NewMemory() *Store {
	return &Store{data: map[string]json.RawMessage{}}
}

// Close releases the underlying file. Safe to call on an in-memory store.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// GetRaw returns the raw JSON stored at key. The bool reports presence. The
// returned slice is a fresh copy safe to retain.
func (s *Store) GetRaw(key string) (json.RawMessage, bool) {
	if s.db == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		v, ok := s.data[key]
		return v, ok
	}
	var out json.RawMessage
	var found bool
	_ = s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(kvBucket).Get([]byte(key))
		if v != nil {
			out = append(json.RawMessage(nil), v...) // copy: bolt bytes die with the txn
			found = true
		}
		return nil
	})
	return out, found
}

// SetRaw stores raw JSON at key. For the durable store this is one committed
// B+tree write; for the in-memory store it is a map assignment.
func (s *Store) SetRaw(key string, v json.RawMessage) error {
	if s.db == nil {
		s.mu.Lock()
		s.data[key] = v
		s.mu.Unlock()
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(kvBucket).Put([]byte(key), v)
	})
}

// Delete removes key (no-op if absent).
func (s *Store) Delete(key string) error {
	if s.db == nil {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		return tx.Bucket(kvBucket).Delete([]byte(key))
	})
}

// Keys returns all keys with the given prefix, sorted ascending. An empty
// prefix returns every key. The returned slice is a fresh copy.
func (s *Store) Keys(prefix string) []string {
	if s.db == nil {
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
	var out []string
	p := []byte(prefix)
	_ = s.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket(kvBucket).Cursor()
		if len(p) == 0 {
			for k, _ := c.First(); k != nil; k, _ = c.Next() {
				out = append(out, string(k))
			}
			return nil
		}
		// bbolt keys are byte-sorted: seek to the prefix and walk while it holds.
		for k, _ := c.Seek(p); k != nil && strings.HasPrefix(string(k), prefix); k, _ = c.Next() {
			out = append(out, string(k))
		}
		return nil
	})
	return out
}

// Len reports the number of stored entries.
func (s *Store) Len() int {
	if s.db == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		return len(s.data)
	}
	n := 0
	_ = s.db.View(func(tx *bolt.Tx) error {
		n = tx.Bucket(kvBucket).Stats().KeyN
		return nil
	})
	return n
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
