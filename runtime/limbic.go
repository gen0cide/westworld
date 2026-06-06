package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/limbic"
)

const (
	// limbicLedgerKey is the durable home of the trust ledger. The
	// "relationship:" namespace makes the memory layer write it back to local
	// disk now and mirror it to mesa later. The leading underscore keeps it out
	// of any real player's relationship key.
	limbicLedgerKey = "relationship:_ledger"
	// limbicFlushInterval is how often the ledger is persisted while running.
	limbicFlushInterval = 30 * time.Second
)

// runLimbic is the host's System-1 affect path: a bus-subscriber goroutine
// (sibling to heartbeatLoop, no tick) that deterministically folds game events
// into the mood vector and the trust ledger. It never calls the LLM and never
// sends packets — it only updates in-RAM limbic state that the pearl engine and
// relation_with read. It restores the ledger on start and persists it on a
// cadence + on exit. Started by Run; exits when ctx is cancelled.
func (h *Host) runLimbic(ctx context.Context) {
	h.loadLimbic(ctx) // restore durable trust before processing events

	ch := h.bus.Subscribe("*", 256)
	flush := time.NewTicker(limbicFlushInterval)
	defer flush.Stop()
	for {
		select {
		case <-ctx.Done():
			// Final best-effort flush on a fresh bounded context (ctx is dead).
			fctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			h.flushLimbic(fctx)
			cancel()
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			h.limbicHandle(ev)
		case <-flush.C:
			h.flushLimbic(ctx)
		}
	}
}

// limbicHandle routes one event into affect / ledger updates. The mapping is
// deliberately conservative for v1 (clear, attributable signals only); combat
// damage→affect and trade/attack→trust land as those signals are threaded.
func (h *Host) limbicHandle(ev event.Event) {
	switch e := ev.(type) {
	// --- affect (mood) ---
	case event.ExperienceGain:
		h.affect.OnXPGain(e.XP)
	case event.LevelUp:
		h.affect.OnLevelUp()
	case event.Death:
		h.affect.OnDeath()

	// --- relationships (familiarity + mild engagement signal) ---
	case event.ChatReceived:
		if e.Speaker != "" {
			h.ledger.Met(e.Speaker)
		}
	case event.PrivateMessage:
		if e.Sender != "" {
			// Someone choosing to whisper you is a small positive social signal.
			h.ledger.Met(e.Sender)
			h.ledger.Observe(e.Sender, true, 0.2)
		}
	}
}

// loadLimbic restores the trust ledger from the durable memory layer. No-op when
// no memory manager is wired (in-RAM-only hosts / tests).
func (h *Host) loadLimbic(ctx context.Context) {
	if h.Memory == nil || h.ledger == nil {
		return
	}
	rec, ok, err := h.Memory.Get(ctx, limbicLedgerKey)
	if err != nil || !ok {
		return
	}
	var rows []limbic.Entry
	if json.Unmarshal(rec.Value, &rows) == nil && len(rows) > 0 {
		h.ledger.Import(rows)
		h.log.Info("limbic: restored trust ledger", "relationships", len(rows))
	}
}

// flushLimbic persists the trust ledger through the memory layer. Best-effort:
// a write failure is logged, never fatal. No-op when no memory manager is wired.
func (h *Host) flushLimbic(ctx context.Context) {
	if h.Memory == nil || h.ledger == nil {
		return
	}
	rows := h.ledger.Export()
	if len(rows) == 0 {
		return
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return
	}
	if err := h.Memory.Put(ctx, limbicLedgerKey, raw); err != nil {
		h.log.Warn("limbic: ledger flush failed", "err", err)
	}
}
