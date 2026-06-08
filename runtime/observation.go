package runtime

import (
	"context"
	"fmt"
	"time"

	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// observation.go is the host-side emission seam for the PERCEPTION FIREHOSE
// (docs/world-knowledge-and-learning.md §3.5): a salience-gated, fire-and-forget
// stream of raw observations up to mesa, where the distillation crons chew them
// into the knowledge + relationship ledgers. Phase 1 wires the plumbing + a STUB
// salience gate; the WRITERS (perception -> emitObservation calls) and the real
// novelty x importance x curiosity gating land in later phases.

// observationFloor is the stub salience gate: below it an observation is dropped
// locally and never streamed (most perception never leaves the host). The real
// novelty/curiosity gate replaces this constant later.
const observationFloor = 0.3

// emitObservation streams one raw perception up to mesa if it clears the salience
// gate. Fire-and-forget — never blocks the caller; no-op when mesa is unwired or
// under analysis-mode (learning I/O is frozen by the operator override, like the
// ledgers).
func (h *Host) emitObservation(kind, subject, text string, salience float64) {
	if salience < observationFloor {
		return
	}
	sink := h.mesaMem
	if sink == nil || !sink.Healthy() {
		return
	}
	if h.AnalysisActive() {
		return
	}
	now := time.Now()
	o := &mesaclient.Observation{
		HostID:         h.opts.Username,
		IdempotencyKey: fmt.Sprintf("%s|%s|%d", kind, subject, now.UnixNano()),
		Kind:           kind,
		Subject:        subject,
		Text:           text,
		Salience:       salience,
		OccurredAtUnix: now.Unix(),
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sink.RecordObservation(ctx, o); err != nil {
			h.log.Debug("observation: emit failed", "err", err)
		}
	}()
}
