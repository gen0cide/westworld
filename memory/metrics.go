package memory

import (
	"maps"
	"sync"
	"time"
)

// Metrics is the per-host memory telemetry. It is intentionally cheap (a few
// counters under a mutex) and exists from day one so we can SEE access patterns
// per namespace+tier — the evidence that tells us when a class is cache-warm
// enough to tighten its remote deadline, and whether adaptive tiering is ever
// worth building.
type Metrics struct {
	mu sync.Mutex

	// Per (namespace, tier) hit counts: "relationship/scratch" → n.
	hits map[string]int64
	// Per namespace miss counts (no tier served it).
	misses map[string]int64
	// Promotions from a slower tier into a faster one.
	promotions int64
	// Remote call attempts, errors, and accumulated latency.
	remoteCalls   int64
	remoteErrors  int64
	remoteLatency time.Duration
	// Current journal depth (unsynced remote writes).
	journalDepth int64
}

func newMetrics() *Metrics {
	return &Metrics{hits: map[string]int64{}, misses: map[string]int64{}}
}

func (m *Metrics) hit(namespace, tier string) {
	m.mu.Lock()
	m.hits[namespace+"/"+tier]++
	m.mu.Unlock()
}

func (m *Metrics) miss(namespace string) {
	m.mu.Lock()
	m.misses[namespace]++
	m.mu.Unlock()
}

func (m *Metrics) promote() {
	m.mu.Lock()
	m.promotions++
	m.mu.Unlock()
}

func (m *Metrics) remote(latency time.Duration, err error) {
	m.mu.Lock()
	m.remoteCalls++
	m.remoteLatency += latency
	if err != nil {
		m.remoteErrors++
	}
	m.mu.Unlock()
}

func (m *Metrics) setJournalDepth(n int) {
	m.mu.Lock()
	m.journalDepth = int64(n)
	m.mu.Unlock()
}

// Snapshot is an immutable copy of the metrics for inspection / logging.
type Snapshot struct {
	Hits          map[string]int64
	Misses        map[string]int64
	Promotions    int64
	RemoteCalls   int64
	RemoteErrors  int64
	RemoteLatency time.Duration
	JournalDepth  int64
}

// Snapshot returns a copy of the current counters.
func (m *Metrics) Snapshot() Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	hits := make(map[string]int64, len(m.hits))
	maps.Copy(hits, m.hits)
	misses := make(map[string]int64, len(m.misses))
	maps.Copy(misses, m.misses)
	return Snapshot{
		Hits:          hits,
		Misses:        misses,
		Promotions:    m.promotions,
		RemoteCalls:   m.remoteCalls,
		RemoteErrors:  m.remoteErrors,
		RemoteLatency: m.remoteLatency,
		JournalDepth:  m.journalDepth,
	}
}
