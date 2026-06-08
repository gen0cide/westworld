package runtime

import (
	"context"
	"encoding/json"

	"github.com/gen0cide/westworld/cognition/goalgraph"
)

// goalgraph.go persists the host's intention graph (cognition/goalgraph) on the
// same spine as the knowledge + trust ledgers (limbic.go): warm-start from the
// local store, snapshot-flush on the limbic cadence + on exit, under the
// "goalgraph:" memory namespace. Phase 1 wires the structure + persistence; the
// writers (failure -> enabling sub-goal, recipes -> edges) and the goal-graph-
// aware director land in later phases.

const goalGraphKey = "goalgraph:_main"

func (h *Host) loadGoalGraph(ctx context.Context) {
	if h.Memory == nil || h.goalGraph == nil {
		return
	}
	rec, ok, err := h.Memory.Get(ctx, goalGraphKey)
	if err != nil || !ok {
		return
	}
	var snap goalgraph.Snapshot
	if json.Unmarshal(rec.Value, &snap) == nil && len(snap.Nodes) > 0 {
		h.goalGraph.Import(snap)
		h.log.Info("goalgraph: restored", "nodes", len(snap.Nodes), "edges", len(snap.Edges))
	}
}

func (h *Host) flushGoalGraph(ctx context.Context) {
	if h.goalGraph == nil || h.Memory == nil || h.goalGraph.Empty() {
		return
	}
	if h.AnalysisActive() {
		return
	}
	raw, err := json.Marshal(h.goalGraph.Export())
	if err != nil {
		return
	}
	if err := h.Memory.Put(ctx, goalGraphKey, raw); err != nil {
		h.log.Warn("goalgraph: flush failed", "err", err)
	}
}
