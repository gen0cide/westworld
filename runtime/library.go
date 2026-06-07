package runtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/gen0cide/westworld/memory"
)

// libraryKey is the durable home of the learned routine library. The "library:"
// namespace falls under the default local-durable policy (write-back to bbolt).
const libraryKey = "library:_main"

// LibEntry is one promoted routine: a reusable DSL routine the host learned from
// a successful LLM-authored move, keyed by the situation signature that produced
// it. Source is the full authored DSL (replayed inline; no file needed).
type LibEntry struct {
	Name   string `json:"name"`
	Source string `json:"source"`
	Hits   int    `json:"hits"`
	At     int64  `json:"at"`
}

// RoutineLibrary is the host's learned cache of reusable routines: situation
// signature → routine. The cheap loop (#16) promotes a successful LLM-authored
// routine here, then replays it locally (no LLM) whenever the same coarse
// situation recurs — so the first "mine tin" costs an Act call and the next
// thousand don't. In-RAM with write-through persistence to the local memory
// tier (bbolt); survives restarts. Safe for concurrent use.
type RoutineLibrary struct {
	mu  sync.Mutex
	m   map[string]LibEntry
	mgr *memory.Manager // persistence; nil = in-RAM only (tests)
	now func() time.Time
}

// NewRoutineLibrary builds a library, restoring any persisted entries.
func NewRoutineLibrary(mgr *memory.Manager) *RoutineLibrary {
	l := &RoutineLibrary{m: map[string]LibEntry{}, mgr: mgr, now: time.Now}
	l.load()
	return l
}

// Lookup returns the routine cached for a signature (bumping its hit count).
func (l *RoutineLibrary) Lookup(sig string) (LibEntry, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	e, ok := l.m[sig]
	if ok {
		e.Hits++
		l.m[sig] = e // telemetry only; not flushed (cheap)
	}
	return e, ok
}

// Promote stores (or replaces) the routine for a signature and persists.
func (l *RoutineLibrary) Promote(sig, name, source string) {
	if sig == "" || source == "" {
		return
	}
	l.mu.Lock()
	l.m[sig] = LibEntry{Name: name, Source: source, At: l.now().Unix()}
	l.mu.Unlock()
	l.flush()
}

// Evict drops the routine for a signature (used when a replay fails).
func (l *RoutineLibrary) Evict(sig string) {
	l.mu.Lock()
	_, had := l.m[sig]
	delete(l.m, sig)
	l.mu.Unlock()
	if had {
		l.flush()
	}
}

// Len reports the number of learned routines.
func (l *RoutineLibrary) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.m)
}

func (l *RoutineLibrary) load() {
	if l.mgr == nil {
		return
	}
	rec, ok, err := l.mgr.Get(context.Background(), libraryKey)
	if err != nil || !ok {
		return
	}
	var m map[string]LibEntry
	if json.Unmarshal(rec.Value, &m) == nil && m != nil {
		l.mu.Lock()
		l.m = m
		l.mu.Unlock()
	}
}

func (l *RoutineLibrary) flush() {
	if l.mgr == nil {
		return
	}
	l.mu.Lock()
	raw, err := json.Marshal(l.m)
	l.mu.Unlock()
	if err == nil {
		_ = l.mgr.Put(context.Background(), libraryKey, raw)
	}
}
