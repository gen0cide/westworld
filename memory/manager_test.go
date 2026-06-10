package memory

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/gen0cide/westworld/hostkv"
)

// fakeRemote is a controllable Remote for tests.
type fakeRemote struct {
	mu      sync.Mutex
	data    map[string]json.RawMessage
	healthy bool
	gets    int
	puts    int
}

func newFakeRemote(healthy bool) *fakeRemote {
	return &fakeRemote{data: map[string]json.RawMessage{}, healthy: healthy}
}
func (r *fakeRemote) Get(_ context.Context, key string) (json.RawMessage, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gets++
	v, ok := r.data[key]
	return v, ok, nil
}
func (r *fakeRemote) Put(_ context.Context, key string, val json.RawMessage) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.puts++
	r.data[key] = val
	return nil
}
func (r *fakeRemote) Delete(_ context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}
func (r *fakeRemote) Search(context.Context, string, int) ([]SearchHit, error) { return nil, nil }
func (r *fakeRemote) Healthy() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.healthy
}
func (r *fakeRemote) setHealthy(v bool) { r.mu.Lock(); r.healthy = v; r.mu.Unlock() }
func (r *fakeRemote) getCount() int     { r.mu.Lock(); defer r.mu.Unlock(); return r.gets }
func (r *fakeRemote) putCount() int     { r.mu.Lock(); defer r.mu.Unlock(); return r.puts }

func newManager(t *testing.T, remote Remote) *Manager {
	t.Helper()
	return New(Options{
		Scratch: hostkv.NewScratch(64),
		Local:   hostkv.NewMemory(),
		Remote:  remote,
	})
}

func raw(s string) json.RawMessage { return json.RawMessage(`"` + s + `"`) }

func TestWriteAroundScratchOnly(t *testing.T) {
	m := newManager(t, nil) // NopRemote
	ctx := context.Background()
	if err := m.Put(ctx, "scratch:cur", raw("v")); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.local.GetRaw("scratch:cur"); ok {
		t.Fatal("write-around should not touch local")
	}
	rec, ok, _ := m.Get(ctx, "scratch:cur")
	if !ok || rec.Origin != "scratch" {
		t.Fatalf("got %+v ok=%v, want scratch hit", rec, ok)
	}
}

func TestWriteBackLocalPlusJournal(t *testing.T) {
	r := newFakeRemote(true)
	m := newManager(t, r)
	ctx := context.Background()
	if err := m.Put(ctx, "relationship:alex", raw("trusted")); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.local.GetRaw("relationship:alex"); !ok {
		t.Fatal("write-back should persist locally")
	}
	if r.putCount() != 0 {
		t.Fatal("write-back must NOT write remote synchronously")
	}
	if m.JournalDepth() != 1 {
		t.Fatalf("journal depth = %d, want 1", m.JournalDepth())
	}
}

func TestCascadeLocalHitPromotesToScratch(t *testing.T) {
	r := newFakeRemote(true)
	m := newManager(t, r)
	ctx := context.Background()
	// Seed local directly, leave scratch cold.
	_ = m.local.SetRaw("relationship:bob", raw("rival"))

	rec, ok, _ := m.Get(ctx, "relationship:bob")
	if !ok || rec.Origin != "local" {
		t.Fatalf("first read got %+v, want local hit", rec)
	}
	// Promotion: scratch now holds it.
	if _, sok := m.scratch.Get("relationship:bob"); !sok {
		t.Fatal("local hit should promote into scratch")
	}
	rec2, _, _ := m.Get(ctx, "relationship:bob")
	if rec2.Origin != "scratch" {
		t.Fatalf("second read origin = %q, want scratch (promoted)", rec2.Origin)
	}
	if r.getCount() != 0 {
		t.Fatal("local hits must not touch remote")
	}
}

func TestRemoteHitPromotesAndIsFresh(t *testing.T) {
	r := newFakeRemote(true)
	_ = r.Put(context.Background(), "reputation:carl", raw("famous"))
	r.puts = 0 // reset after seeding
	m := newManager(t, r)
	ctx := context.Background()

	rec, ok, err := m.Get(ctx, "reputation:carl")
	if err != nil || !ok {
		t.Fatalf("remote read ok=%v err=%v", ok, err)
	}
	if rec.Origin != "remote" || !rec.Fresh {
		t.Fatalf("got %+v, want fresh remote", rec)
	}
	// Promoted into local + scratch.
	if _, lok := m.local.GetRaw("reputation:carl"); !lok {
		t.Fatal("remote hit should promote into local")
	}
}

func TestNegativeCacheAvoidsRepeatRemoteHit(t *testing.T) {
	r := newFakeRemote(true)
	m := newManager(t, r)
	ctx := context.Background()
	// reputation namespace has TTL>0 ⇒ misses are tombstoned. But it's
	// CascadeFresh, which skips the scratch leg on read; the tombstone is still
	// written on a miss and short-circuits the NEXT cached read. Use a cached
	// namespace to observe the negative cache cleanly: relationship (CascadeCached).
	_, ok, _ := m.Get(ctx, "relationship:ghost") // miss → local miss → remote miss → tombstone
	if ok {
		t.Fatal("expected miss")
	}
	first := r.getCount()
	_, ok, _ = m.Get(ctx, "relationship:ghost") // should hit the negative cache
	if ok {
		t.Fatal("expected miss again")
	}
	if r.getCount() != first {
		t.Fatalf("remote was hit again (%d→%d); negative cache should have short-circuited", first, r.getCount())
	}
}

func TestOfflineJournalsThenFlush(t *testing.T) {
	r := newFakeRemote(false) // start offline
	m := newManager(t, r)
	ctx := context.Background()

	// RemoteAuthoritative write while offline ⇒ journaled, cached locally.
	if err := m.Put(ctx, "reputation:dora", raw("infamous")); err != nil {
		t.Fatal(err)
	}
	if r.putCount() != 0 {
		t.Fatal("offline put must not reach remote")
	}
	if m.JournalDepth() != 1 {
		t.Fatalf("journal depth = %d, want 1", m.JournalDepth())
	}
	// Local cached projection is still readable offline.
	if _, ok := m.local.GetRaw("reputation:dora"); !ok {
		t.Fatal("remote-authoritative write should keep a local projection")
	}

	// Come back online and flush.
	r.setHealthy(true)
	n, err := m.Flush(ctx)
	if err != nil || n != 1 {
		t.Fatalf("Flush n=%d err=%v, want 1", n, err)
	}
	if m.JournalDepth() != 0 {
		t.Fatalf("journal not drained: depth=%d", m.JournalDepth())
	}
	if r.putCount() != 1 {
		t.Fatalf("remote puts = %d, want 1 after flush", r.putCount())
	}
}

func TestHintEphemeralForcesWriteAround(t *testing.T) {
	m := newManager(t, nil)
	ctx := context.Background()
	// relationship is normally WriteBack (local). Ephemeral hint ⇒ scratch only.
	if err := m.Put(ctx, "relationship:tmp", raw("v"), Hint{Ephemeral: true}); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.local.GetRaw("relationship:tmp"); ok {
		t.Fatal("ephemeral hint should keep it out of local")
	}
}

func TestHintFreshBypassesScratch(t *testing.T) {
	r := newFakeRemote(true)
	_ = r.Put(context.Background(), "relationship:eve", raw("remote-truth"))
	r.puts = 0
	m := newManager(t, r)
	ctx := context.Background()
	// Prime scratch with a stale value.
	m.scratch.Set("relationship:eve", raw("stale"), time.Minute)

	rec, ok, _ := m.Get(ctx, "relationship:eve", Hint{Fresh: true})
	if !ok || rec.Origin != "remote" {
		t.Fatalf("fresh read got %+v, want remote (bypassing stale scratch)", rec)
	}
}

func TestMaturityDialScalesDeadline(t *testing.T) {
	m := newManager(t, nil)
	base := 4 * time.Second
	m.SetMaturity(0) // newborn ⇒ full patience
	if d := m.remoteDeadline(base); d != base {
		t.Fatalf("newborn deadline = %v, want %v", d, base)
	}
	m.SetMaturity(1) // veteran ⇒ floor
	if d := m.remoteDeadline(base); d != 200*time.Millisecond {
		t.Fatalf("veteran deadline = %v, want floor 200ms", d)
	}
	m.SetMaturity(0.5) // halfway
	if d := m.remoteDeadline(base); d != 2*time.Second {
		t.Fatalf("half-maturity deadline = %v, want 2s", d)
	}
}

func TestMetricsRecorded(t *testing.T) {
	r := newFakeRemote(true)
	_ = r.Put(context.Background(), "reputation:x", raw("v"))
	r.puts = 0
	m := newManager(t, r)
	ctx := context.Background()
	_, _, _ = m.Get(ctx, "reputation:x")       // remote hit + promotion
	_ = m.Put(ctx, "relationship:y", raw("v")) // write-back ⇒ journal
	_, _, _ = m.Get(ctx, "relationship:y")     // scratch hit
	snap := m.Metrics().Snapshot()
	if snap.Promotions < 1 {
		t.Errorf("promotions = %d, want >= 1", snap.Promotions)
	}
	if snap.RemoteCalls < 1 {
		t.Errorf("remote calls = %d, want >= 1", snap.RemoteCalls)
	}
	if snap.JournalDepth != 1 {
		t.Errorf("journal depth = %d, want 1", snap.JournalDepth)
	}
	if snap.Hits["reputation/remote"] < 1 {
		t.Errorf("missing reputation/remote hit metric: %+v", snap.Hits)
	}
}

// TestJournalCompactionAndSeqRestore pins the two journal-leak fixes: a re-Put
// of the same key replaces its pending write-back entry (periodic flushers
// must not grow the journal), and a reopened Manager continues the sequence
// instead of overwriting pending entries from seq 0.
func TestJournalCompactionAndSeqRestore(t *testing.T) {
	local := hostkv.NewMemory()
	m := New(Options{Scratch: hostkv.NewScratch(256), Local: local}) // NopRemote + default WriteBack: every Put journals

	for i := 0; i < 5; i++ {
		if err := m.Put(context.Background(), "ledger:alex", []byte(`{"n":`+string(rune('0'+i))+`}`)); err != nil {
			t.Fatal(err)
		}
	}
	if d := m.JournalDepth(); d != 1 {
		t.Fatalf("journal depth after 5 same-key Puts = %d, want 1 (compaction)", d)
	}
	if err := m.Put(context.Background(), "goal:current", []byte(`"x"`)); err != nil {
		t.Fatal(err)
	}
	if d := m.JournalDepth(); d != 2 {
		t.Fatalf("depth = %d, want 2 distinct keys", d)
	}

	// "Restart": a new Manager over the same local store must continue seq —
	// a fresh Put must not overwrite the surviving pending entries.
	m2 := New(Options{Scratch: hostkv.NewScratch(256), Local: local})
	if err := m2.Put(context.Background(), "knowledge:varrock", []byte(`1`)); err != nil {
		t.Fatal(err)
	}
	if d := m2.JournalDepth(); d != 3 {
		t.Fatalf("depth after restart+Put = %d, want 3 (seq restored, no overwrite)", d)
	}
	// And same-key compaction still works across the restart boundary.
	if err := m2.Put(context.Background(), "ledger:alex", []byte(`{"n":9}`)); err != nil {
		t.Fatal(err)
	}
	if d := m2.JournalDepth(); d != 3 {
		t.Fatalf("depth after re-Put of pre-restart key = %d, want 3 (index rebuilt)", d)
	}
}
