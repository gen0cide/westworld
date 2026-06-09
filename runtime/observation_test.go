package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// TestEmitObservationGate proves the salience gate: below-floor perceptions are
// dropped locally (never streamed), above-floor ones are emitted up to mesa with
// the right fields.
func TestEmitObservationGate(t *testing.T) {
	h := newTestHost() // Username "test"
	sink := &fakeSink{healthy: true, obs: make(chan *mesaclient.Observation, 4)}
	h.SetMesaMemory(sink)

	// Below the floor: dropped.
	h.emitObservation("entity_sighting", "rune plate", "a player wearing rune plate", 0.1)
	select {
	case o := <-sink.obs:
		t.Fatalf("below-floor observation should be dropped, got %+v", o)
	case <-time.After(150 * time.Millisecond):
	}

	// Above the floor: emitted with the right fields.
	h.emitObservation("entity_sighting", "rune plate", "a player wearing rune plate", 0.9)
	select {
	case o := <-sink.obs:
		if o.Kind != "entity_sighting" || o.Subject != "rune plate" || o.Salience != 0.9 || o.HostID != "test" {
			t.Fatalf("emitted observation = %+v", o)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("above-floor observation was not emitted")
	}
}

// blockingSink is a MesaMemory whose RecordObservation BLOCKS until released,
// tracking the peak number of CONCURRENT in-flight calls. It implements only the
// methods emitObservation exercises (Healthy + RecordObservation); the rest panic
// so a stray call is loud.
type blockingSink struct {
	release  chan struct{} // RecordObservation blocks reading from this
	inFlight int32
	peak     int32
	entered  chan struct{} // signalled once per RecordObservation entry
}

func (b *blockingSink) Healthy() bool { return true }
func (b *blockingSink) RecordObservation(_ context.Context, _ *mesaclient.Observation) error {
	n := atomic.AddInt32(&b.inFlight, 1)
	for {
		p := atomic.LoadInt32(&b.peak)
		if n <= p || atomic.CompareAndSwapInt32(&b.peak, p, n) {
			break
		}
	}
	if b.entered != nil {
		b.entered <- struct{}{}
	}
	<-b.release
	atomic.AddInt32(&b.inFlight, -1)
	return nil
}

// The remaining MesaMemory methods are never reached on the emitObservation path.
func (b *blockingSink) Remember(context.Context, *mesaclient.Episode) error { panic("unused") }
func (b *blockingSink) Recall(context.Context, *mesaclient.Query) (*mesaclient.Knowledge, error) {
	panic("unused")
}
func (b *blockingSink) SyncRelationships(context.Context, string, []mesaclient.Relationship) error {
	panic("unused")
}
func (b *blockingSink) FetchRelationships(context.Context, string) ([]mesaclient.Relationship, error) {
	panic("unused")
}
func (b *blockingSink) SyncGoal(context.Context, string, mesaclient.Goal) error { panic("unused") }
func (b *blockingSink) FetchGoal(context.Context, string) (mesaclient.Goal, bool, error) {
	panic("unused")
}
func (b *blockingSink) ReportMetrics(context.Context, string, []mesaclient.Metric) error {
	panic("unused")
}
func (b *blockingSink) SyncKnowledge(context.Context, string, []mesaclient.KnowledgeEntry) error {
	panic("unused")
}
func (b *blockingSink) FetchKnowledge(context.Context, string) ([]mesaclient.KnowledgeEntry, error) {
	panic("unused")
}
func (b *blockingSink) SyncGoalGraph(context.Context, string, mesaclient.GoalGraphSnapshot) error {
	panic("unused")
}
func (b *blockingSink) FetchGoalGraph(context.Context, string) (mesaclient.GoalGraphSnapshot, error) {
	panic("unused")
}

// TestEmitObservationInflightCapped is the M15 regression: a burst of above-floor
// observations must NOT spawn an unbounded goroutine per observation. With a sink
// that blocks in RecordObservation, at most observationMaxInflight emit goroutines
// can be live at once; the rest are DROPPED at the semaphore (never spawned).
func TestEmitObservationInflightCapped(t *testing.T) {
	h := newTestHost() // Username "test"
	sink := &blockingSink{release: make(chan struct{}), entered: make(chan struct{}, 64)}
	h.SetMesaMemory(sink)

	// Fire a burst far exceeding the cap. Each clears the floor (0.9).
	const burst = 32
	var wg sync.WaitGroup
	for i := 0; i < burst; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.emitObservation("entity_sighting", "rune plate", "a player wearing rune plate", 0.9)
		}()
	}
	wg.Wait() // every emit either spawned-and-blocked or was dropped; none blocks the caller

	// Give the (at most cap) spawned goroutines a moment to enter RecordObservation.
	deadline := time.After(time.Second)
	for atomic.LoadInt32(&sink.inFlight) < int32(observationMaxInflight) {
		select {
		case <-deadline:
			t.Fatalf("expected %d goroutines to reach RecordObservation; only %d did",
				observationMaxInflight, atomic.LoadInt32(&sink.inFlight))
		case <-time.After(2 * time.Millisecond):
		}
	}
	// Settle, then assert the peak never exceeded the cap.
	time.Sleep(50 * time.Millisecond)
	if peak := atomic.LoadInt32(&sink.peak); peak > int32(observationMaxInflight) {
		t.Fatalf("concurrent emit goroutines = %d, exceeds cap %d", peak, observationMaxInflight)
	}

	// Release the blocked goroutines and let them drain.
	close(sink.release)
	for atomic.LoadInt32(&sink.inFlight) > 0 {
		time.Sleep(2 * time.Millisecond)
	}
}

// TestObservationIdemKey is the M4 regression: CAPTURE kinds (dialog/chat/server)
// must carry a stable text-hash suffix so two DISTINCT same-second lines produce
// DISTINCT keys (and survive the server's ON CONFLICT DO NOTHING), while once-per-
// event kinds (outcome/transaction/location) keep the coarse kind|subject|second
// key (a re-perceived same-second event SHOULD collapse). The exact key format must
// stay byte-identical to mesad ltm.go observationIdemKey.
func TestObservationIdemKey(t *testing.T) {
	const sec = int64(1700000000)
	tests := []struct {
		name        string
		kind        string
		subjectA    string
		textA       string
		subjectB    string
		textB       string
		wantCollide bool // true: same key (collapse); false: distinct keys
	}{
		{
			name: "server_msg distinct text same second do NOT collapse",
			kind: "server_msg", subjectA: "server", textA: "Welcome to the tutorial.",
			subjectB: "server", textB: "Talk to the guide to continue.",
			wantCollide: false,
		},
		{
			name: "server_msg identical text same second DO collapse (true dup)",
			kind: "server_msg", subjectA: "server", textA: "Welcome to the tutorial.",
			subjectB: "server", textB: "Welcome to the tutorial.",
			wantCollide: true,
		},
		{
			name: "npc_dialog distinct lines from same speaker stay distinct",
			kind: "npc_dialog", subjectA: "Hans", textA: "Hello traveller.",
			subjectB: "Hans", textB: "Mind the cellar.",
			wantCollide: false,
		},
		{
			name: "player_chat distinct lines stay distinct",
			kind: "player_chat", subjectA: "Bob", textA: "anyone selling a pickaxe?",
			subjectB: "Bob", textB: "thanks!",
			wantCollide: false,
		},
		{
			name: "outcome coarse key collapses same-second re-perception",
			kind: "outcome", subjectA: "Goblin", textA: "defeated Goblin",
			subjectB: "Goblin", textB: "defeated Goblin (again, same second)",
			wantCollide: true,
		},
		{
			name: "transaction coarse key collapses same-second re-perception",
			kind: "transaction", subjectA: "Bob's Axes", textA: "stocks 3 items: a, b, c",
			subjectB: "Bob's Axes", textB: "stocks 4 items: a, b, c, d",
			wantCollide: true,
		},
		{
			name: "location coarse key collapses same-second re-perception",
			kind: "location", subjectA: "Lumbridge", textA: "arrived at Lumbridge",
			subjectB: "Lumbridge", textB: "arrived at Lumbridge",
			wantCollide: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ka := observationIdemKey(tt.kind, tt.subjectA, tt.textA, sec)
			kb := observationIdemKey(tt.kind, tt.subjectB, tt.textB, sec)
			if collided := ka == kb; collided != tt.wantCollide {
				t.Fatalf("collide=%v, want %v\n  ka=%q\n  kb=%q", collided, tt.wantCollide, ka, kb)
			}
		})
	}
}

// TestEmitObservationDistinctSameSecondLines proves the EMITTER (not just the key
// helper) streams distinct keys for two distinct same-second server lines — the M4
// path end-to-end on the host side. (The mesad side proves both rows persist.)
func TestEmitObservationDistinctSameSecondLines(t *testing.T) {
	h := newTestHost() // Username "test"
	sink := &fakeSink{healthy: true, obs: make(chan *mesaclient.Observation, 4)}
	h.SetMesaMemory(sink)

	h.emitObservation("server_msg", "server", "Welcome to the tutorial.", 0.9)
	h.emitObservation("server_msg", "server", "Talk to the guide to continue.", 0.9)

	keys := map[string]bool{}
	for i := 0; i < 2; i++ {
		select {
		case o := <-sink.obs:
			keys[o.IdempotencyKey] = true
		case <-time.After(2 * time.Second):
			t.Fatalf("only %d of 2 observations emitted", i)
		}
	}
	if len(keys) != 2 {
		t.Fatalf("two distinct same-second server lines collapsed to %d key(s): %v", len(keys), keys)
	}
}

// TestEmitObservationNoSink proves emit is a safe no-op when mesa is unwired
// (or unhealthy) — it must never panic or block the caller.
func TestEmitObservationNoSink(t *testing.T) {
	h := newTestHost() // no mesaMem wired
	h.emitObservation("outcome", "mining", "got iron ore", 0.9)

	unhealthy := &fakeSink{healthy: false, obs: make(chan *mesaclient.Observation, 1)}
	h.SetMesaMemory(unhealthy)
	h.emitObservation("outcome", "mining", "got iron ore", 0.9)
	select {
	case o := <-unhealthy.obs:
		t.Fatalf("unhealthy sink should not receive observations, got %+v", o)
	case <-time.After(150 * time.Millisecond):
	}
}
