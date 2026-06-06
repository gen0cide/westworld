package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/gen0cide/westworld/hostkv"
)

// journalPrefix namespaces write-back queue entries inside the local store.
// The NUL keeps them out of any caller's namespace and out of Keys() scans that
// use printable prefixes.
const journalPrefix = "\x00journal:"

// negPrefix marks negative-cache tombstones in scratch (remote miss memory).
const negPrefix = "\x00neg:"

// tombstone is the sentinel value stored in scratch for a negative-cache entry.
type tombstone struct{}

// Record is a read result with provenance, so the brain/pearl can weight a
// stale local guess differently from a fresh remote fact.
type Record struct {
	Value      json.RawMessage
	Origin     string  // "scratch" | "local" | "remote"
	Fresh      bool    // true when served from the authority tier this call
	Confidence float64 // 1.0 for fresh authority reads; lower for stale caches
}

// Manager is the tiered memory manager. Construct with New. Safe for concurrent
// use. Scratch (L0) and local (L1) are used directly; remote (L2) is pluggable.
type Manager struct {
	scratch *hostkv.Scratch
	local   *hostkv.Store
	remote  Remote
	policy  Policy
	metrics *Metrics

	mu       sync.Mutex
	maturity float64 // 0 (newborn, patient) .. 1 (veteran, impatient)
	seq      int64   // journal sequence
	now      func() time.Time
}

// Options configures a Manager. Scratch and Local are required; the rest have
// sensible defaults.
type Options struct {
	Scratch *hostkv.Scratch
	Local   *hostkv.Store
	Remote  Remote // nil ⇒ NopRemote (offline)
	Policy  *Policy
}

// New builds a Manager. It panics if Scratch or Local is nil — a host's memory
// must always have its local tiers.
func New(opts Options) *Manager {
	if opts.Scratch == nil || opts.Local == nil {
		panic("memory: New requires non-nil Scratch and Local")
	}
	m := &Manager{
		scratch: opts.Scratch,
		local:   opts.Local,
		remote:  opts.Remote,
		metrics: newMetrics(),
		now:     time.Now,
	}
	if m.remote == nil {
		m.remote = NopRemote{}
	}
	if opts.Policy != nil {
		m.policy = *opts.Policy
	} else {
		m.policy = DefaultPolicy()
	}
	m.refreshJournalDepth()
	return m
}

// Metrics returns the telemetry handle (Snapshot for a copy).
func (m *Manager) Metrics() *Metrics { return m.metrics }

// SetMaturity sets the host's maturity in [0,1]. 0 = newborn (patient: full
// remote deadline); 1 = veteran (impatient: deadline shrinks toward a floor).
// This is the dial that carries a host from "phones home for everything, waits"
// to "mostly local, fast".
func (m *Manager) SetMaturity(v float64) {
	if v < 0 {
		v = 0
	} else if v > 1 {
		v = 1
	}
	m.mu.Lock()
	m.maturity = v
	m.mu.Unlock()
}

// remoteDeadline scales a class's base deadline by maturity: newborn hosts wait
// the full base; mature hosts wait as little as a floor.
func (m *Manager) remoteDeadline(base time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	const floor = 200 * time.Millisecond
	m.mu.Lock()
	mat := m.maturity
	m.mu.Unlock()
	return max(time.Duration(float64(base)*(1-mat)), floor)
}

// Get resolves key through the policy and returns a Record. ok is false on a
// miss. Hints (optional) override the namespace policy for this call.
func (m *Manager) Get(ctx context.Context, key string, hints ...Hint) (Record, bool, error) {
	c := m.policy.resolve(key, hints...)
	ns := namespaceOf(key)
	switch c.Read {
	case LocalOnly:
		return m.getCached(key, ns, c)
	case RemoteOnly:
		return m.getRemote(ctx, key, ns, c)
	case CascadeFresh:
		return m.getCascade(ctx, key, ns, c, true)
	default:
		return m.getCascade(ctx, key, ns, c, false)
	}
}

// negativeHit reports whether a recent remote miss for key is still tombstoned
// in scratch (so we can skip re-hitting the network for a known-absent key).
func (m *Manager) negativeHit(key string) bool {
	_, ok := m.scratch.Get(negPrefix + key)
	return ok
}

// getCached probes the cache tiers only (scratch → local), promoting a local
// hit into scratch. Used by LocalOnly and as the fast leg of a cascade.
func (m *Manager) getCached(key, ns string, c Class) (Record, bool, error) {
	if v, ok := m.scratch.Get(key); ok {
		if raw, ok := v.(json.RawMessage); ok {
			m.metrics.hit(ns, "scratch")
			return Record{Value: raw, Origin: "scratch", Fresh: false, Confidence: 0.8}, true, nil
		}
	}
	if raw, ok := m.local.GetRaw(key); ok {
		m.metrics.hit(ns, "local")
		m.scratch.Set(key, raw, c.TTL) // promote
		fresh := c.Authority == AuthLocal
		return Record{Value: raw, Origin: "local", Fresh: fresh, Confidence: confidence(fresh)}, true, nil
	}
	return Record{}, false, nil
}

// getCascade runs the full scratch → local → remote read with promotion. When
// fresh is true (CascadeFresh) the cache tiers (and the negative cache) are
// skipped for the remote authority read, but still used as a fallback if remote
// can't answer.
func (m *Manager) getCascade(ctx context.Context, key, ns string, c Class, fresh bool) (Record, bool, error) {
	if !fresh {
		if m.negativeHit(key) { // known-absent: don't touch the network
			m.metrics.miss(ns)
			return Record{}, false, nil
		}
		if rec, ok, _ := m.getCached(key, ns, c); ok {
			return rec, true, nil
		}
	}
	if rec, ok, err := m.getRemote(ctx, key, ns, c); ok || err != nil {
		return rec, ok, err
	}
	// Remote couldn't answer (miss / unhealthy). On a fresh read we skipped the
	// caches above — fall back to them now rather than report a false miss.
	if fresh {
		if rec, ok, _ := m.getCached(key, ns, c); ok {
			return rec, true, nil
		}
	}
	return Record{}, false, nil
}

// getRemote reads the remote tier under a maturity-scaled deadline, promoting a
// hit down into local+scratch and tombstoning a miss in scratch.
func (m *Manager) getRemote(ctx context.Context, key, ns string, c Class) (Record, bool, error) {
	if !m.remote.Healthy() {
		return Record{}, false, nil
	}
	rctx := ctx
	if d := m.remoteDeadline(c.RemoteDeadline); d > 0 {
		var cancel context.CancelFunc
		rctx, cancel = context.WithTimeout(ctx, d)
		defer cancel()
	}
	start := m.now()
	raw, ok, err := m.remote.Get(rctx, key)
	m.metrics.remote(m.now().Sub(start), err)
	if err != nil {
		return Record{}, false, fmt.Errorf("memory: remote get %q: %w", key, err)
	}
	if !ok {
		// Remember the miss briefly so we don't re-hit the network for it.
		if c.TTL > 0 {
			m.scratch.Set(negPrefix+key, tombstone{}, c.TTL)
		}
		m.metrics.miss(ns)
		return Record{}, false, nil
	}
	m.metrics.hit(ns, "remote")
	m.metrics.promote()
	_ = m.local.SetRaw(key, raw) // cache the of-record value locally
	m.scratch.Set(key, raw, c.TTL)
	return Record{Value: raw, Origin: "remote", Fresh: true, Confidence: 1.0}, true, nil
}

// Put resolves key through the policy and writes it across the appropriate
// tiers. Hints (optional) override the namespace policy for this call.
func (m *Manager) Put(ctx context.Context, key string, val json.RawMessage, hints ...Hint) error {
	c := m.policy.resolve(key, hints...)
	switch c.Write {
	case WriteAround:
		m.scratch.Set(key, val, c.TTL)
		return nil
	case WriteThrough:
		if err := m.local.SetRaw(key, val); err != nil {
			return err
		}
		m.scratch.Set(key, val, c.TTL)
		if m.remote.Healthy() {
			start := m.now()
			err := m.remote.Put(ctx, key, val)
			m.metrics.remote(m.now().Sub(start), err)
			return err
		}
		return m.enqueueJournal(key, val)
	case RemoteAuthoritative:
		m.scratch.Set(key, val, c.TTL)
		_ = m.local.SetRaw(key, val) // cached projection
		if m.remote.Healthy() {
			start := m.now()
			err := m.remote.Put(ctx, key, val)
			m.metrics.remote(m.now().Sub(start), err)
			if err == nil {
				return nil
			}
		}
		return m.enqueueJournal(key, val)
	default: // WriteBack
		if err := m.local.SetRaw(key, val); err != nil {
			return err
		}
		m.scratch.Set(key, val, c.TTL)
		return m.enqueueJournal(key, val)
	}
}

// Delete removes key from every tier it can reach (and clears any tombstone).
func (m *Manager) Delete(ctx context.Context, key string) error {
	m.scratch.Delete(key)
	m.scratch.Delete(negPrefix + key)
	err := m.local.Delete(key)
	if m.remote.Healthy() {
		if rerr := m.remote.Delete(ctx, key); rerr != nil && err == nil {
			err = rerr
		}
	}
	return err
}

// Search runs semantic recall — capability-exclusive to the remote tier. It
// returns nil (no error) when the remote is unavailable, so callers degrade to
// whatever local knowledge they have rather than failing.
func (m *Manager) Search(ctx context.Context, query string, k int) ([]SearchHit, error) {
	if !m.remote.Healthy() {
		return nil, nil
	}
	start := m.now()
	hits, err := m.remote.Search(ctx, query, k)
	m.metrics.remote(m.now().Sub(start), err)
	return hits, err
}

// --- journal (write-back queue) --------------------------------------------

type journalEntry struct {
	Key string          `json:"key"`
	Val json.RawMessage `json:"val"`
}

// enqueueJournal records a pending remote write in the local store so it
// survives a crash / offline period until Flush drains it.
func (m *Manager) enqueueJournal(key string, val json.RawMessage) error {
	m.mu.Lock()
	m.seq++
	seq := m.seq
	m.mu.Unlock()
	jk := fmt.Sprintf("%s%020d", journalPrefix, seq)
	if err := hostkv.Set(m.local, jk, journalEntry{Key: key, Val: val}); err != nil {
		return err
	}
	m.refreshJournalDepth()
	return nil
}

// Flush drains the write-back journal to the remote tier, oldest first,
// stopping at the first failure (so order is preserved). Returns the number
// flushed. A no-op when the remote is unhealthy.
func (m *Manager) Flush(ctx context.Context) (int, error) {
	if !m.remote.Healthy() {
		return 0, nil
	}
	keys := m.local.Keys(journalPrefix)
	sort.Strings(keys) // sequence-ordered (zero-padded)
	n := 0
	for _, jk := range keys {
		entry, ok, err := hostkv.Get[journalEntry](m.local, jk)
		if err != nil || !ok {
			_ = m.local.Delete(jk) // unreadable entry — drop it
			continue
		}
		start := m.now()
		err = m.remote.Put(ctx, entry.Key, entry.Val)
		m.metrics.remote(m.now().Sub(start), err)
		if err != nil {
			break // preserve order; retry on the next Flush
		}
		_ = m.local.Delete(jk)
		n++
	}
	m.refreshJournalDepth()
	return n, nil
}

// JournalDepth reports the number of pending (unsynced) remote writes.
func (m *Manager) JournalDepth() int {
	return len(m.local.Keys(journalPrefix))
}

func (m *Manager) refreshJournalDepth() {
	m.metrics.setJournalDepth(len(m.local.Keys(journalPrefix)))
}

// confidence maps freshness to a confidence score for a Record.
func confidence(fresh bool) float64 {
	if fresh {
		return 1.0
	}
	return 0.6
}
