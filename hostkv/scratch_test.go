package hostkv

import (
	"testing"
	"time"
)

// fakeClock is a manually-advanced clock for deterministic TTL tests.
type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestScratch(cap int, c *fakeClock) *Scratch {
	s := NewScratch(cap)
	s.now = c.now
	return s
}

func TestScratchSetGet(t *testing.T) {
	s := NewScratch(4)
	s.Set("examined", true, 0)
	v, ok := s.Get("examined")
	if !ok || v.(bool) != true {
		t.Fatalf("Get = %v, %v", v, ok)
	}
	if _, ok := s.Get("missing"); ok {
		t.Fatal("missing key reported present")
	}
}

func TestScratchLRUEviction(t *testing.T) {
	s := NewScratch(2)
	s.Set("a", 1, 0)
	s.Set("b", 2, 0)
	s.Get("a")       // a is now most-recently-used; b is LRU
	s.Set("c", 3, 0) // evicts b
	if _, ok := s.Get("b"); ok {
		t.Fatal("b should have been evicted")
	}
	if _, ok := s.Get("a"); !ok {
		t.Fatal("a should survive (was used)")
	}
	if _, ok := s.Get("c"); !ok {
		t.Fatal("c should be present")
	}
}

func TestScratchTTLExpiry(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	s := newTestScratch(4, c)
	s.Set("refractory", true, 5*time.Second)
	if _, ok := s.Get("refractory"); !ok {
		t.Fatal("entry should be live before expiry")
	}
	c.advance(5 * time.Second) // now == expiry ⇒ expired (expiry is inclusive)
	if _, ok := s.Get("refractory"); ok {
		t.Fatal("entry should be expired")
	}
}

func TestScratchTTLZeroNeverExpires(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	s := newTestScratch(4, c)
	s.Set("forever", 1, 0)
	c.advance(1000 * time.Hour)
	if _, ok := s.Get("forever"); !ok {
		t.Fatal("ttl=0 entry should never expire")
	}
}

func TestScratchLenPrunesExpired(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	s := newTestScratch(8, c)
	s.Set("live", 1, 0)
	s.Set("dies", 2, time.Second)
	if s.Len() != 2 {
		t.Fatalf("Len = %d, want 2", s.Len())
	}
	c.advance(2 * time.Second)
	if s.Len() != 1 {
		t.Fatalf("Len after expiry = %d, want 1", s.Len())
	}
}

func TestScratchOverwriteRefreshes(t *testing.T) {
	c := &fakeClock{t: time.Unix(1000, 0)}
	s := newTestScratch(4, c)
	s.Set("k", 1, time.Second)
	c.advance(2 * time.Second) // would expire the original
	s.Set("k", 2, time.Second) // overwrite resets expiry
	v, ok := s.Get("k")
	if !ok || v.(int) != 2 {
		t.Fatalf("after overwrite Get = %v, %v", v, ok)
	}
}
