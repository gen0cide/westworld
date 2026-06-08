package runtime

import (
	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/limbic"
)

// mind.go exposes READ-ONLY snapshots of the host's three mind structures — the
// world-knowledge ledger (what it knows), the trust ledger (who it knows), and
// the intention graph (what it's trying to do + open questions) — for the cradle
// inspector (the debugging keystone of docs/world-knowledge-and-learning.md).
// Pure accessors over in-RAM state; no behaviour.

// KnowledgeFacts returns the host's world-knowledge as read-only facts, best-known
// first.
func (h *Host) KnowledgeFacts() []knowledge.Fact {
	if h.knowledge == nil {
		return nil
	}
	return h.knowledge.All()
}

// Relationships returns the host's trust-ledger views (read-only), most-familiar
// first.
func (h *Host) Relationships() []limbic.Rel {
	if h.ledger == nil {
		return nil
	}
	return h.ledger.All()
}

// GoalGraphSnapshot returns the host's intention-graph snapshot (nodes + edges).
func (h *Host) GoalGraphSnapshot() goalgraph.Snapshot {
	if h.goalGraph == nil {
		return goalgraph.Snapshot{}
	}
	return h.goalGraph.Export()
}

// OpenQuestions returns the host's unresolved open-question nodes (newest first).
func (h *Host) OpenQuestions() []goalgraph.Node {
	if h.goalGraph == nil {
		return nil
	}
	return h.goalGraph.OpenQuestions()
}
