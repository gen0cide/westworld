package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
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

// bootstrapKnowledgeFromMesa reconstitutes the world-knowledge ledger from mesa
// when there is no local state to warm-start from (fresh host, no bbolt, in-
// memory mode) — the authoritative cold-start path, parallel to
// bootstrapLedgerFromMesa. This is where the Phase-4 distillation pays off: the
// cron grows the ledger from the firehose, and a restarted/cold host warm-starts
// beliefs it never explicitly wrote. The no-clobber invariant is SELF-ENFORCED
// here (L9): a host that already has local knowledge is never overwritten by the
// fetch — the All()==0 check is inside, not a caller convention. No-op when mesa
// is absent/unhealthy, the live ledger is non-empty, or mesa returns nothing.
func (h *Host) bootstrapKnowledgeFromMesa(ctx context.Context) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() || h.knowledge == nil {
		return
	}
	if len(h.knowledge.All()) != 0 {
		return // never clobber a host that already learned locally (L9)
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	entries, err := h.mesaMem.FetchKnowledge(cctx, h.opts.Username)
	if err != nil || len(entries) == 0 {
		return
	}
	h.knowledge.Import(entriesToKnowledge(entries))
	h.log.Info("knowledge: bootstrapped ledger from mesa", "subjects", len(entries))
}

// reimportKnowledgeFromMesa is the WARM-host re-import (M17): on the limbic flush
// tick it pulls the now server-reconciled world-knowledge ledger back from mesa and
// folds it in NON-DESTRUCTIVELY (knowledge.ImportMerge — NOT the cold-start-only
// bootstrapKnowledgeFromMesa replace), so the consolidation/insight crons' β-bumps
// + reconciliations land in the host's LOCAL ledger and thus SURVIVE the host's next
// flush-up (which would otherwise LWW-overwrite the cron's work per subject). The
// merge is per-(subject,claim) max-evidence, so the host's own live beliefs are
// never regressed and a cron belief the host never re-learned is adopted. Frozen
// under analysis-mode. Runs even when the ledger is non-empty (the warm case the
// cold-start bootstrap skips). No-op when mesa is absent/unhealthy or returns nothing.
func (h *Host) reimportKnowledgeFromMesa(ctx context.Context) {
	if h.knowledge == nil || h.mesaMem == nil || !h.mesaMem.Healthy() {
		return
	}
	if h.AnalysisActive() {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	entries, err := h.mesaMem.FetchKnowledge(cctx, h.opts.Username)
	if err != nil || len(entries) == 0 {
		return
	}
	h.knowledge.ImportMerge(entriesToKnowledge(entries))
}

// flushKnowledgeToMesa mirrors the host's locally-learned beliefs up to mesa's
// distilled knowledge store (best-effort), so a host's own learning also reaches
// the of-record + future bootstrap source. The cron and the host both upsert by
// (host_id, subject) → last-writer-wins per subject; the host re-bootstraps the
// merged result on a cold start. No-op when mesa is absent/unhealthy or empty.
func (h *Host) flushKnowledgeToMesa(ctx context.Context) {
	if h.knowledge == nil || h.mesaMem == nil || !h.mesaMem.Healthy() {
		return
	}
	if h.AnalysisActive() {
		return
	}
	rows := h.knowledge.Export()
	if len(rows) == 0 {
		return
	}
	if err := h.mesaMem.SyncKnowledge(ctx, h.opts.Username, knowledgeToEntries(rows)); err != nil {
		h.log.Debug("knowledge: ledger mesa-sync failed", "err", err)
	}
}

// entriesToKnowledge converts wire knowledge entries back to ledger import rows.
func entriesToKnowledge(entries []mesaclient.KnowledgeEntry) []knowledge.Entry {
	out := make([]knowledge.Entry, 0, len(entries))
	for _, e := range entries {
		row := knowledge.Entry{
			Subject:    e.Subject,
			Kind:       e.Kind,
			Encounters: e.Encounters,
			LastSeen:   e.LastSeenUnix,
			Tags:       e.Tags,
		}
		for _, b := range e.Beliefs {
			row.Beliefs = append(row.Beliefs, knowledge.Belief{
				Claim: b.Claim, Provenance: b.Provenance, Alpha: b.Alpha, Beta: b.Beta, At: b.AtUnix,
			})
		}
		out = append(out, row)
	}
	return out
}

// knowledgeToEntries converts the ledger's snapshot rows to the wire form.
func knowledgeToEntries(rows []knowledge.Entry) []mesaclient.KnowledgeEntry {
	out := make([]mesaclient.KnowledgeEntry, 0, len(rows))
	for _, r := range rows {
		e := mesaclient.KnowledgeEntry{
			Subject:      r.Subject,
			Kind:         r.Kind,
			Encounters:   r.Encounters,
			LastSeenUnix: r.LastSeen,
			Tags:         r.Tags,
		}
		for _, b := range r.Beliefs {
			e.Beliefs = append(e.Beliefs, mesaclient.KnowledgeBelief{
				Claim: b.Claim, Provenance: b.Provenance, Alpha: b.Alpha, Beta: b.Beta, AtUnix: b.At,
			})
		}
		out = append(out, e)
	}
	return out
}
