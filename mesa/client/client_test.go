package mesaclient

import (
	"context"
	"errors"
	"testing"

	"github.com/gen0cide/westworld/brain"
	"github.com/gen0cide/westworld/cognition"
	"github.com/gen0cide/westworld/memory"
)

// Compile-time proof the adapters satisfy the existing host seams.
var (
	_ brain.Strategist = AsStrategist(StubClient{}, "h")
	_ cognition.Client = AsRetriever(StubClient{}, "h")
	_ memory.Remote    = AsRemote(StubClient{}, "h")
)

func TestStubIsOffline(t *testing.T) {
	if (StubClient{}).Healthy() {
		t.Fatal("stub should be offline")
	}
}

func TestStrategistAdapterOfflineErrors(t *testing.T) {
	s := AsStrategist(StubClient{}, "h")
	if _, err := s.Decide(context.Background(), brain.Situation{Question: "q", Options: []string{"a"}}); !errors.Is(err, ErrOffline) {
		t.Fatalf("offline Decide err = %v, want ErrOffline", err)
	}
}

func TestRetrieverAdapterOfflineEmptyBundle(t *testing.T) {
	r := AsRetriever(StubClient{}, "h")
	b, err := r.Retrieve(context.Background(), cognition.Retrieval{Goal: "find bank"})
	if err != nil {
		t.Fatalf("offline Retrieve err = %v, want nil", err)
	}
	if b == nil || b.Goal != "find bank" || len(b.Episodic) != 0 {
		t.Fatalf("offline bundle = %+v, want empty with goal set", b)
	}
}

func TestRemoteAdapterOffline(t *testing.T) {
	rem := AsRemote(StubClient{}, "h")
	if rem.Healthy() {
		t.Fatal("remote should report offline")
	}
	if _, ok, _ := rem.Get(context.Background(), "reputation:x"); ok {
		t.Fatal("offline Get should miss")
	}
	hits, err := rem.Search(context.Background(), "who scammed me", 3)
	if err != nil || len(hits) != 0 {
		t.Fatalf("offline Search = %v, %v; want empty/nil", hits, err)
	}
	// Mirror (Put) is a no-op offline (stub swallows it).
	if err := rem.Put(context.Background(), "kv:x", []byte(`"v"`)); err != nil {
		t.Fatalf("offline Put err = %v", err)
	}
}

func TestSubscribeYieldsClosedChannel(t *testing.T) {
	ch, err := StubClient{}.Subscribe(context.Background(), "h")
	if err != nil {
		t.Fatalf("Subscribe err = %v", err)
	}
	if _, ok := <-ch; ok {
		t.Fatal("stub Subscribe channel should be closed (no directives)")
	}
}
