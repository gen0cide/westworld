package runtime

import (
	"context"

	"github.com/gen0cide/westworld/event"
)

// runLimbic is the host's System-1 affect path: a bus-subscriber goroutine
// (sibling to heartbeatLoop, no tick) that deterministically folds game events
// into the mood vector and the trust ledger. It never calls the LLM and never
// sends packets — it only updates in-RAM limbic state that the pearl engine and
// relation_with read. Started by Run; exits when ctx is cancelled.
func (h *Host) runLimbic(ctx context.Context) {
	ch := h.bus.Subscribe("*", 256)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			h.limbicHandle(ev)
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
