package runtime

import (
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
