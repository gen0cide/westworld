package runtime

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
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
	if h.goalGraph == nil || h.Memory == nil {
		return
	}
	if h.AnalysisActive() {
		return
	}
	// Snapshot first, then decide on the snapshot — one lock acquisition, no
	// TOCTOU between an Empty() probe and a separate Export() (matches
	// flushKnowledge / flushLimbic).
	snap := h.goalGraph.Export()
	if len(snap.Nodes) == 0 {
		return
	}
	raw, err := json.Marshal(snap)
	if err != nil {
		return
	}
	if err := h.Memory.Put(ctx, goalGraphKey, raw); err != nil {
		h.log.Warn("goalgraph: flush failed", "err", err)
	}
}

// bootstrapGoalGraphFromMesa reconstitutes the intention graph from mesa when
// there is no local state to warm-start from (fresh host, no bbolt, in-memory
// mode) — the cold-start path, parallel to bootstrapKnowledgeFromMesa. This is
// where the Phase-4b insight cron pays off: the cron grows the graph (open-
// question closure, cross-entity chaining) and a restarted/cold host warm-starts
// it. CALLER MUST Empty()-guard: goalgraph.Import REPLACES wholesale, so an
// unguarded fetch would clobber a mid-run host's live graph. No-op when mesa is
// absent/unhealthy or the stored graph is empty.
func (h *Host) bootstrapGoalGraphFromMesa(ctx context.Context) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() || h.goalGraph == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	snap, err := h.mesaMem.FetchGoalGraph(cctx, h.opts.Username)
	if err != nil || len(snap.Nodes) == 0 {
		return
	}
	h.goalGraph.Import(snapshotFromClient(snap))
	h.log.Info("goalgraph: bootstrapped from mesa", "nodes", len(snap.Nodes), "edges", len(snap.Edges))
}

// flushGoalGraphToMesa mirrors the host's local intention graph up to mesa's
// distilled goal_graphs store (best-effort), so a host's own goal-graph growth
// also reaches the of-record + future bootstrap source. The cron and the host
// both upsert by host_id → last-writer-wins; the host re-bootstraps the merged
// result on a cold start. Frozen under analysis-mode like the other learning I/O.
// No-op when mesa is absent/unhealthy or empty.
func (h *Host) flushGoalGraphToMesa(ctx context.Context) {
	if h.goalGraph == nil || h.mesaMem == nil || !h.mesaMem.Healthy() {
		return
	}
	if h.AnalysisActive() {
		return
	}
	snap := h.goalGraph.Export()
	if len(snap.Nodes) == 0 {
		return
	}
	if err := h.mesaMem.SyncGoalGraph(ctx, h.opts.Username, snapshotToClient(snap)); err != nil {
		h.log.Debug("goalgraph: mesa-sync failed", "err", err)
	}
}

// snapshotToClient converts the internal goal-graph snapshot to the wire form.
// NOTE the field rename: internal Node.At ↔ wire GoalGraphNode.AtUnix.
func snapshotToClient(s goalgraph.Snapshot) mesaclient.GoalGraphSnapshot {
	out := mesaclient.GoalGraphSnapshot{}
	for _, n := range s.Nodes {
		out.Nodes = append(out.Nodes, mesaclient.GoalGraphNode{
			ID: n.ID, Kind: n.Kind, Label: n.Label, Status: n.Status,
			Progress: n.Progress, Tags: n.Tags, AtUnix: n.At,
		})
	}
	for _, e := range s.Edges {
		out.Edges = append(out.Edges, mesaclient.GoalGraphEdge{From: e.From, To: e.To, Rel: e.Rel})
	}
	return out
}

// snapshotFromClient converts the wire form back to the internal snapshot.
func snapshotFromClient(s mesaclient.GoalGraphSnapshot) goalgraph.Snapshot {
	out := goalgraph.Snapshot{}
	for _, n := range s.Nodes {
		out.Nodes = append(out.Nodes, goalgraph.Node{
			ID: n.ID, Kind: n.Kind, Label: n.Label, Status: n.Status,
			Progress: n.Progress, Tags: n.Tags, At: n.AtUnix,
		})
	}
	for _, e := range s.Edges {
		out.Edges = append(out.Edges, goalgraph.Edge{From: e.From, To: e.To, Rel: e.Rel})
	}
	return out
}
