package runtime

import (
	"context"
	"fmt"
	"hash/fnv"
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

// observationMaxInflight caps the concurrent fire-and-forget RecordObservation
// goroutines (M15). Mirrors reactiveMaxInflight: a salience burst that clears the
// gate must not spawn an unbounded goroutine per observation — over the cap the
// observation is DROPPED (an ambient perception is not worth queueing).
const observationMaxInflight = 3

// observationCaptureKinds are the firehose CAPTURE kinds — raw, free-text speech
// lines (one row PER line, the cron's distillation substrate) where many distinct
// lines legitimately share kind|subject within one wall-clock second. SystemMessage/
// ChatReceived(no speaker) always emit subject="server", so a same-second burst of
// system/quest text (tutorial prompts, multi-line quest text) would collapse under a
// coarse kind|subject|second key (M4). For these we append a stable hash of the TEXT
// so distinct lines stay distinct. Every other kind (outcome/transaction/location) is
// genuinely once-per-event and keeps the coarse key (a re-perceived same-second event
// SHOULD collapse). MUST mirror mesad ltm.go AddObservation's observationCaptureKinds.
var observationCaptureKinds = map[string]bool{
	"npc_dialog":  true,
	"player_chat": true,
	"server_msg":  true,
	"claim_heard": true,
}

// observationIdemKey builds the firehose idempotency key. For CAPTURE kinds it is
// kind|subject|second|<fnv32a(text) hex> so distinct same-second lines do NOT collapse
// under the server's ON CONFLICT DO NOTHING; for once-per-event kinds it is the coarse
// kind|subject|second. This construction MUST stay byte-identical to mesad ltm.go
// AddObservation's fallback key (the host and server agree on the key, M4).
func observationIdemKey(kind, subject, text string, occurredAtUnix int64) string {
	if observationCaptureKinds[kind] {
		h := fnv.New32a()
		h.Write([]byte(text))
		return fmt.Sprintf("%s|%s|%d|%08x", kind, subject, occurredAtUnix, h.Sum32())
	}
	return fmt.Sprintf("%s|%s|%d", kind, subject, occurredAtUnix)
}

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
		return // fast-path: don't build or spawn while learning I/O is frozen
	}
	now := time.Now()
	o := &mesaclient.Observation{
		HostID: h.opts.Username,
		// Second-granularity idempotency key (observationIdemKey). For CAPTURE kinds
		// (dialog/chat/server) it carries a stable hash of the text so distinct
		// same-second lines do NOT collapse (M4); for once-per-event kinds it stays
		// the coarse kind|subject|second. Byte-identical to the server-side fallback
		// key (mesad ltm.go AddObservation), so host and server agree.
		IdempotencyKey: observationIdemKey(kind, subject, text, now.Unix()),
		Kind:           kind,
		Subject:        subject,
		Text:           text,
		Salience:       salience,
		OccurredAtUnix: now.Unix(),
	}
	// Bound the concurrent emit goroutines (M15): a non-blocking acquire on the
	// semaphore, DROP the observation when full (an ambient perception is not worth
	// queueing), release in the goroutine's defer. Matches reactive.spawnExtract.
	if h.obsInflight != nil {
		select {
		case h.obsInflight <- struct{}{}:
		default:
			return // at the concurrent-emit cap → DROP (bounded)
		}
	}
	go func() {
		if h.obsInflight != nil {
			defer func() { <-h.obsInflight }()
		}
		// Re-check at send time: analysis mode may have been entered between the
		// fast-path check above and this deferred emit (TOCTOU close).
		if h.AnalysisActive() {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sink.RecordObservation(ctx, o); err != nil {
			h.log.Debug("observation: emit failed", "err", err)
		}
	}()
}
