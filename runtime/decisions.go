package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gen0cide/westworld/event"
)

// publishDecision streams a lightweight per-layer cognition decision to the
// Thoughts panel (the debug control plane + JSONL) — the execution moments BELOW
// the per-turn Act decision: cheap-loop replay/promote/evict, stall, spin, goal
// lifecycle. trigger names the layer; reasoning is the human-readable detail.
// Safe from any host goroutine (the event bus is concurrent).
func (h *Host) publishDecision(trigger, moveKind, reasoning string) {
	if h == nil || h.bus == nil {
		return
	}
	h.bus.Publish(event.AgentThought{
		Trigger:   trigger,
		MoveKind:  moveKind,
		Reasoning: reasoning,
		Goal:      h.LiveGoal(),
	})
}

// decisionRecord is one persisted line of the host's decision stream — the
// same AgentThought the cradle Thoughts panel renders, made durable so an
// overnight soak can be TRACED the next morning: which trigger fired, what
// the director decided, why, and under which goal.
type decisionRecord struct {
	At        time.Time `json:"at"`
	Trigger   string    `json:"trigger"`
	MoveKind  string    `json:"kind"`
	Reasoning string    `json:"reasoning"`
	Goal      string    `json:"goal,omitempty"`
}

// SetDecisionLogPath publishes where this host's durable decision stream
// (decisions.jsonl) lives. RunHost calls it before starting runDecisionLog;
// "" is ignored (no durable stream).
func (h *Host) SetDecisionLogPath(path string) {
	if path == "" {
		return
	}
	h.decisionLogPath.Store(&path)
}

// DecisionLogPath returns the decisions.jsonl path published via
// SetDecisionLogPath, or "" when this host has no durable decision stream.
func (h *Host) DecisionLogPath() string {
	if p := h.decisionLogPath.Load(); p != nil {
		return *p
	}
	return ""
}

// runDecisionLog (a host goroutine, sibling of runLimbic/runMemory) appends
// every AgentThought to <dataDir>/decisions.jsonl — append-only, one JSON
// object per line, flushed per write so a crash loses at most one record.
// ALSO mirrors each decision up to mesa as a low-salience observation
// (kind="decision") so fleet-level analysis can query one place. The file
// handle is owned here and closed when ctx ends. No-op when path is empty
// (fresh/ephemeral hosts).
func (h *Host) runDecisionLog(ctx context.Context, path string) {
	if path == "" || h.bus == nil {
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.log.Warn("decision log: mkdir failed; stream not persisted", "err", err)
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		h.log.Warn("decision log: open failed; stream not persisted", "err", err)
		return
	}
	defer f.Close()

	ch := h.bus.Subscribe("*", 256)
	enc := json.NewEncoder(f)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			t, isThought := ev.(event.AgentThought)
			if !isThought {
				continue
			}
			rec := decisionRecord{
				At:        time.Now(),
				Trigger:   t.Trigger,
				MoveKind:  t.MoveKind,
				Reasoning: t.Reasoning,
				Goal:      t.Goal,
			}
			if err := enc.Encode(rec); err != nil {
				h.log.Warn("decision log: write failed", "err", err)
			}
			// Mesa mirror: ride the existing observation stream (durable
			// fleet-side, queryable per host) at modest salience — these
			// are introspection records, not world perception.
			h.emitObservation("decision", t.Trigger,
				fmt.Sprintf("[%s] %s", t.MoveKind, t.Reasoning), 0.4)
		}
	}
}
