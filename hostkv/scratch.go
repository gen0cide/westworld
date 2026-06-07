package hostkv

import (
	"container/list"
	"sync"
	"time"
)

// Scratch is a host's ephemeral working memory: an in-memory, capacity-bounded
// LRU cache with optional per-entry TTL. It never touches disk.
//
// Use it for short-term facts a host would not be wrong to forget on logout:
// "I already examined this trade offer", "the last tile I stood on", quirk
// refractory windows. On capacity overflow the least-recently-used entry is
// evicted. Reads and writes both count as use (move-to-front). Expired entries
// are removed lazily when accessed via Get, and pruned in bulk by Len.
//
// This is the same LRU+TTL shape the pearl decision cache uses; it lives here
// so both share one tested implementation. Safe for concurrent use.
type Scratch struct {
	mu  sync.Mutex
	cap int
	ll  *list.List               // front = most recently used
	m   map[string]*list.Element // key → element holding *scratchEntry
	now func() time.Time         // injectable clock for tests
}

type scratchEntry struct {
	key    string
	val    any
	expiry time.Time // zero ⇒ never expires
}

// NewScratch returns a scratch cache holding at most capacity entries. A
// capacity <= 0 is treated as 1 (a cache that holds a single entry); callers
// wanting "unbounded" should pick a large explicit cap.
func NewScratch(capacity int) *Scratch {
	if capacity <= 0 {
		capacity = 1
	}
	return &Scratch{
		cap: capacity,
		ll:  list.New(),
		m:   make(map[string]*list.Element, capacity),
		now: time.Now,
	}
}

// Set stores val at key, marking it most-recently-used. A ttl > 0 sets an
// expiry; ttl <= 0 means the entry never expires. Overwriting an existing key
// refreshes its position and expiry. May evict the LRU entry to stay in bounds.
func (s *Scratch) Set(key string, val any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiry time.Time
	if ttl > 0 {
		expiry = s.now().Add(ttl)
	}
	if el, ok := s.m[key]; ok {
		el.Value.(*scratchEntry).val = val
		el.Value.(*scratchEntry).expiry = expiry
		s.ll.MoveToFront(el)
		return
	}
	el := s.ll.PushFront(&scratchEntry{key: key, val: val, expiry: expiry})
	s.m[key] = el
	if s.ll.Len() > s.cap {
		s.evictOldestLocked()
	}
}

// Get returns the value stored at key. The bool is false if the key is absent
// or has expired (an expired entry is removed). A hit marks the entry
// most-recently-used.
func (s *Scratch) Get(key string) (any, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.m[key]
	if !ok {
		return nil, false
	}
	e := el.Value.(*scratchEntry)
	if s.expiredLocked(e) {
		s.removeLocked(el)
		return nil, false
	}
	s.ll.MoveToFront(el)
	return e.val, true
}

// Delete removes key if present.
func (s *Scratch) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if el, ok := s.m[key]; ok {
		s.removeLocked(el)
	}
}

// Len reports the current number of live (non-expired) entries, pruning any
// expired entries it encounters as a side effect.
func (s *Scratch) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	for el := s.ll.Back(); el != nil; {
		prev := el.Prev()
		if s.expiredLocked(el.Value.(*scratchEntry)) {
			s.removeLocked(el)
		}
		el = prev
	}
	return s.ll.Len()
}

func (s *Scratch) expiredLocked(e *scratchEntry) bool {
	return !e.expiry.IsZero() && !s.now().Before(e.expiry)
}

func (s *Scratch) evictOldestLocked() {
	if el := s.ll.Back(); el != nil {
		s.removeLocked(el)
	}
}

func (s *Scratch) removeLocked(el *list.Element) {
	s.ll.Remove(el)
	delete(s.m, el.Value.(*scratchEntry).key)
}
