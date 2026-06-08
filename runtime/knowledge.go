package runtime

import (
	"context"
	"encoding/json"

	"github.com/gen0cide/westworld/cognition/knowledge"
)

// knowledge.go persists the host's SEMANTIC world-knowledge ledger (the third
// leg of the mind; cognition/knowledge). It mirrors the trust-ledger spine in
// limbic.go: warm-start from the local store, snapshot-flush on the limbic
// cadence + on exit, under the "knowledge:" memory namespace (local bbolt now,
// mesa mirror via the namespace's write-back). Phase 1 wires the structure +
// persistence; the writers (perception -> observations) and the distillation
// crons that grow it land in later phases.

const knowledgeLedgerKey = "knowledge:_ledger"

// loadKnowledge restores the knowledge ledger from the durable memory layer.
// No-op when no memory manager is wired (in-RAM-only hosts / tests).
func (h *Host) loadKnowledge(ctx context.Context) {
	if h.Memory == nil || h.knowledge == nil {
		return
	}
	rec, ok, err := h.Memory.Get(ctx, knowledgeLedgerKey)
	if err != nil || !ok {
		return
	}
	var rows []knowledge.Entry
	if json.Unmarshal(rec.Value, &rows) == nil && len(rows) > 0 {
		h.knowledge.Import(rows)
		h.log.Info("knowledge: restored ledger", "subjects", len(rows))
	}
}

// flushKnowledge persists the knowledge ledger to the durable memory layer.
// Best-effort; frozen under analysis-mode like the trust ledger; no-op when
// empty or unwired.
func (h *Host) flushKnowledge(ctx context.Context) {
	if h.knowledge == nil || h.Memory == nil {
		return
	}
	if h.AnalysisActive() {
		return
	}
	rows := h.knowledge.Export()
	if len(rows) == 0 {
		return
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		return
	}
	if err := h.Memory.Put(ctx, knowledgeLedgerKey, raw); err != nil {
		h.log.Warn("knowledge: ledger flush failed", "err", err)
	}
}
