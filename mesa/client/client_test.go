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

func TestActOfflineErrors(t *testing.T) {
	_, err := StubClient{}.Act(context.Background(), &Situation{HostID: "h", Goal: "mine tin"})
	if !errors.Is(err, ErrOffline) {
		t.Fatalf("offline Act err = %v, want ErrOffline", err)
	}
}

func TestMoveTaggedUnion(t *testing.T) {
	// A WriteRoutine move carries quarantined DSL the host runs gated.
	m := Move{Kind: MoveWriteRoutine, RoutineName: "mine_tin", DSLSource: "routine mine_tin() { ... }", Quarantined: true}
	if m.Kind.String() != "write_routine" || !m.Quarantined || m.DSLSource == "" {
		t.Fatalf("write-routine move = %+v", m)
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

// TestEpisodeToPBCarriesRelationDelta proves the episode→proto translator maps the
// full RelationDelta — incl. TotalValueTraded — so a completed-trade episode's
// traded volume reaches mesa instead of being dropped at the wire boundary
// (latent-trap close: value_traded wired end-to-end).
func TestEpisodeToPBCarriesRelationDelta(t *testing.T) {
	pb := episodeToPB(&Episode{
		HostID: "h", Kind: "trade", Text: "Traded with merchant",
		Relationship: &RelationDelta{
			Name: "merchant", DEncounters: 1, TotalValueTraded: 120, AddTags: []string{"trader"},
		},
	})
	rel := pb.GetRelation()
	if rel == nil {
		t.Fatal("episodeToPB dropped the RelationDelta")
	}
	if rel.GetName() != "merchant" || rel.GetTotalValueTraded() != 120 || rel.GetDEncounters() != 1 {
		t.Fatalf("RelationDelta not faithful on the wire: %+v", rel)
	}

	// A nil Relationship maps to a nil proto Relation (no spurious empty delta).
	if got := episodeToPB(&Episode{HostID: "h", Kind: "kill"}).GetRelation(); got != nil {
		t.Fatalf("nil Relationship should map to nil proto Relation, got %+v", got)
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
