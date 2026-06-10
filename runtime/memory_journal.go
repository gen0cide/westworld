package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/memory"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// MesaMemory is the host's two-way seam to mesa's durable state (Postgres): it
// mirrors local state UP (episodes via Remember, the trust ledger via
// SyncRelationships) and, on a cold start with no local state, pulls it back
// DOWN (Recall / FetchRelationships) to reconstitute. The mesa client satisfies
// it. It is optional: with local bbolt present the host warm-starts from disk
// and never needs mesa; but mesa remains the authoritative source able to
// bootstrap a fresh / in-memory / no-file host from nothing.
type MesaMemory interface {
	Remember(ctx context.Context, e *mesaclient.Episode) error
	// RecordObservation streams one raw, salience-gated perception up to mesa
	// (the firehose; cron fodder for distillation). Fire-and-forget.
	RecordObservation(ctx context.Context, o *mesaclient.Observation) error
	Recall(ctx context.Context, q *mesaclient.Query) (*mesaclient.Knowledge, error)
	SyncRelationships(ctx context.Context, hostID string, rels []mesaclient.Relationship) error
	FetchRelationships(ctx context.Context, hostID string) ([]mesaclient.Relationship, error)
	// SyncGoal/FetchGoal mirror the host's standing objective + progress to mesa's
	// structured goals table, so a fresh host (or a future session-genesis) can
	// resume the plan instead of re-deriving it.
	SyncGoal(ctx context.Context, hostID string, g mesaclient.Goal) error
	FetchGoal(ctx context.Context, hostID string) (mesaclient.Goal, bool, error)
	// SyncKnowledge/FetchKnowledge mirror the host's world-knowledge ledger to
	// mesa's distilled knowledge store. The consolidation cron grows it from the
	// observation firehose; FetchKnowledge warm-starts a cold host with beliefs the
	// cron distilled that the host never explicitly wrote (mirrors the trust pair).
	SyncKnowledge(ctx context.Context, hostID string, entries []mesaclient.KnowledgeEntry) error
	FetchKnowledge(ctx context.Context, hostID string) ([]mesaclient.KnowledgeEntry, error)
	// SyncGoalGraph/FetchGoalGraph mirror the host's intention graph to mesa's
	// distilled goal_graphs store. The insight cron grows it (open-question closure,
	// cross-entity chaining); FetchGoalGraph warm-starts a cold host with the graph
	// the cron grew that it never explicitly wrote (mirrors the knowledge pair; the
	// goal graph is AuthLocal, so the cold-start read is Empty()-guarded).
	SyncGoalGraph(ctx context.Context, hostID string, snap mesaclient.GoalGraphSnapshot) error
	FetchGoalGraph(ctx context.Context, hostID string) (mesaclient.GoalGraphSnapshot, error)
	// ReportMetrics ships a host telemetry batch (observability + cron inputs).
	ReportMetrics(ctx context.Context, hostID string, metrics []mesaclient.Metric) error
	Healthy() bool
}

// SetMesaMemory wires the mesa state seam. Call once at startup when mesa is linked.
func (h *Host) SetMesaMemory(m MesaMemory) { h.mesaMem = m }

const (
	// journalKey is the durable home of the episodic journal. The "journal:"
	// namespace makes the memory layer write it to local disk now and queue the
	// mirror-up to mesa later (see memory.DefaultPolicy). The leading underscore
	// keeps it out of any real namespaced key.
	journalKey = "journal:_main"
	// journalFlushInterval is how often the journal is persisted while running.
	journalFlushInterval = 30 * time.Second
)

// runMemory is the host's episodic-memory path: a bus-subscriber goroutine
// (sibling to runLimbic, no tick) that folds salient game events — level-ups,
// kills, deaths, objective milestones — into the durable Journal. It never calls
// the LLM and never sends packets; it only writes in-RAM episodic state that the
// director recalls into the per-turn Situation. It restores the journal on start
// and persists it on a cadence + on exit. Started by Run; exits when ctx is
// cancelled. This is the local half of the memory loop (docs/cognition-and-
// autonomy.md §5/§7.1); the mesa LTM mirror is task #13.
func (h *Host) runMemory(ctx context.Context) {
	if h.journal == nil {
		return
	}
	h.loadJournal(ctx) // warm-start from local bbolt (fast path, no network)
	if h.journalLen() == 0 {
		h.bootstrapJournalFromMesa(ctx) // cold start: reconstitute episodes from mesa
		h.bootstrapGoalFromMesa(ctx)    // ...and resume the standing objective
	}
	// Ignore the post-login LevelUp burst (initial stats sync, not achievements).
	h.memoryWarmupUntil = time.Now().Add(5 * time.Second)

	ch := h.bus.Subscribe("*", 256)
	flush := time.NewTicker(journalFlushInterval)
	defer flush.Stop()
	for {
		select {
		case <-ctx.Done():
			// Final best-effort flush on a fresh bounded context (ctx is dead).
			fctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			h.flushJournal(fctx)
			cancel()
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			h.memoryCapture(ev)
		case <-flush.C:
			h.flushJournal(ctx)
		}
	}
}

// memoryCapture maps one bus event to a durable episode, or ignores it. The
// mapping is deliberately SELECTIVE — the journal must stay salient, so noisy
// signals (every xp tick, every item) are dropped and only milestones land.
// Importance grades how long an episode stays surfaced by Salient().
func (h *Host) memoryCapture(ev event.Event) {
	// Analysis-mode memory suspension: while the operator has taken over, the
	// host is offline to its own life-story — no episodes are captured and the
	// LTM mirror (downstream of this, its only caller) is silenced. Gating the
	// single capture chokepoint suspends both the local journal AND the mesa
	// mirror with one lock-free check (see analysis.go).
	if h.AnalysisActive() {
		return
	}
	var kind, text, entity string
	var importance float64
	switch e := ev.(type) {
	case event.LevelUp:
		if time.Now().Before(h.memoryWarmupUntil) {
			return // post-login stats-sync burst, not a real achievement
		}
		kind, text, importance = "milestone",
			fmt.Sprintf("Reached level %d in %s.", e.NewLevel, event.SkillName(e.Skill)), 0.7
	case event.TargetDied:
		name := h.npcTypeName(e.TypeID)
		label := "an enemy"
		if name != "" {
			label = article(name) + " " + name
		}
		kind, text, importance, entity = "kill", "Defeated "+label+".", 0.45, name
	case event.Death:
		kind, text, importance = "death", "Died and respawned.", 0.9
	case event.SystemMessage:
		// Quest-channel text (MessageType 1) is objective progress — quest
		// steps, "well done", tutorial milestones. The decoder routes no-sender
		// server text here (sender-bearing text becomes ChatReceived), and the
		// noisy general feedback ("you can't reach that") rides MessageGame (0),
		// so gating on MessageQuest keeps the journal to goal-relevant lines.
		if e.Type == event.MessageQuest {
			text, kind, importance = cleanText(e.Message), "objective", 0.7
		}
	}
	if kind == "" || text == "" {
		return
	}
	ep := h.journal.Append(kind, text, importance, entity)
	h.log.Info("memory: captured episode", "kind", ep.Kind, "text", ep.Text)
	h.mirrorEpisode(ep) // best-effort push UP to mesa's durable LTM
}

// mirrorEpisode pushes one episode to mesa's long-term memory, fire-and-forget.
// It never blocks capture and never affects local durability — the local bbolt
// journal is the source of truth for this host. A nil/unhealthy sink is a no-op.
// The idempotency key is content-derived so an at-least-once resend dedupes
// server-side (ON CONFLICT) without a fragile dependency on local sequence ids.
func (h *Host) mirrorEpisode(ep memory.Episode) {
	sink := h.mesaMem
	if sink == nil || !sink.Healthy() {
		return
	}
	e := &mesaclient.Episode{
		HostID:         h.opts.Username,
		IdempotencyKey: fmt.Sprintf("%s|%s|%d", ep.Kind, ep.Text, ep.At),
		Kind:           ep.Kind,
		Text:           ep.Text,
		Importance:     ep.Importance,
		OccurredAtUnix: ep.At,
	}
	if ep.Entity != "" {
		e.Tags = map[string]string{"entity": ep.Entity}
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := sink.Remember(ctx, e); err != nil {
			h.log.Debug("memory: episode mirror failed", "err", err)
		}
	}()
}

// journalLen reports the retained episode count (0 when no journal is wired).
func (h *Host) journalLen() int {
	if h.journal == nil {
		return 0
	}
	return h.journal.Len()
}

// npcTypeName resolves an NPC type id to its catalog name (empty if unknown).
func (h *Host) npcTypeName(typeID int) string {
	if h.facts != nil {
		if def := h.facts.NpcDef(typeID); def != nil {
			return def.Name
		}
	}
	return ""
}

// loadJournal restores the episodic journal from the durable memory layer.
// No-op when no memory manager is wired (in-RAM-only hosts / tests).
func (h *Host) loadJournal(ctx context.Context) {
	if h.Memory == nil || h.journal == nil {
		return
	}
	rec, ok, err := h.Memory.Get(ctx, journalKey)
	if err != nil || !ok {
		return
	}
	var snap memory.JournalSnapshot
	if json.Unmarshal(rec.Value, &snap) == nil && (len(snap.Episodes) > 0 || snap.Objective != "") {
		h.journal.Import(snap)
		h.log.Info("memory: restored episodic journal", "episodes", len(snap.Episodes), "objective", snap.Objective != "")
	}
}

// flushJournal persists the episodic journal locally (bbolt) and mirrors the
// host's standing GOAL (objective + progress) up to mesa's structured goals
// table. Best-effort: a write failure is logged, never fatal.
func (h *Host) flushJournal(ctx context.Context) {
	if h.journal == nil {
		return
	}
	// Analysis-mode: zero periodic persistence (bbolt + mesa goal-sync) while the
	// operator override is active — a hard memory freeze.
	if h.AnalysisActive() {
		return
	}
	snap := h.journal.Export()
	if len(snap.Episodes) == 0 && snap.Objective == "" {
		return
	}
	// Local durable mirror (bbolt) — the host's fast warm-start.
	if h.Memory != nil {
		if raw, err := json.Marshal(snap); err == nil {
			if err := h.Memory.Put(ctx, journalKey, raw); err != nil {
				h.log.Warn("memory: journal flush failed", "err", err)
			}
		}
	}
	// Mesa goal mirror (up) — the resumable plan, structured for analytics/crons.
	if h.mesaMem != nil && h.mesaMem.Healthy() && snap.Objective != "" {
		progress := make([]string, 0, 8)
		for _, e := range h.journal.Salient(8) {
			progress = append(progress, e.Text)
		}
		g := mesaclient.Goal{Objective: snap.Objective, Progress: progress, UpdatedAt: snap.ObjectiveAt}
		if err := h.mesaMem.SyncGoal(ctx, h.opts.Username, g); err != nil {
			h.log.Debug("memory: goal mesa-sync failed", "err", err)
		}
	}
}

// bootstrapGoalFromMesa restores the host's standing objective from mesa on a
// cold start, so a fresh host resumes its plan. The director may then refine it
// (SetObjective is idempotent). No-op when mesa is absent or has no goal.
func (h *Host) bootstrapGoalFromMesa(ctx context.Context) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() || h.journal == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	g, found, err := h.mesaMem.FetchGoal(cctx, h.opts.Username)
	if err != nil || !found || g.Objective == "" {
		return
	}
	if h.journal.SetObjective(g.Objective) {
		h.log.Info("memory: resumed objective from mesa", "objective", g.Objective)
	}
}

// bootstrapJournalFromMesa reconstitutes the episodic journal from mesa's
// durable LTM when there is no local state to warm-start from (fresh host, no
// bbolt file, in-memory mode). This is the authoritative path: mesa can boot a
// host from nothing. It pulls the host's most-recent episodes (empty query →
// recency order) and rebuilds the journal chronologically. The standing
// objective is not fetched here — it arrives via Provision (goals) and is set by
// the director. No-op when mesa is absent/unhealthy.
func (h *Host) bootstrapJournalFromMesa(ctx context.Context) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() || h.journal == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	k, err := h.mesaMem.Recall(cctx, &mesaclient.Query{
		HostID: h.opts.Username,
		Text:   "", // empty query → mesa returns episodes in recency order
		TopK:   memory.DefaultJournalCap,
	})
	if err != nil || k == nil || len(k.Items) == 0 {
		return
	}
	// mesa returns newest-first; the journal is oldest-first chronological.
	eps := make([]memory.Episode, 0, len(k.Items))
	for i := len(k.Items) - 1; i >= 0; i-- {
		kind, at := parseProvenance(k.Items[i].Provenance)
		// Carry the real salience weight + entity attribution back through Recall so
		// a cold-start journal keeps its Salient() ordering and who-it-was-about. Fall
		// back to 0.6 only when mesa reports no importance (legacy/pre-field rows).
		importance := k.Items[i].Importance
		if importance == 0 {
			importance = 0.6
		}
		eps = append(eps, memory.Episode{
			Seq:        int64(len(eps) + 1),
			Kind:       kind,
			Text:       k.Items[i].Text,
			Importance: importance,
			Entity:     k.Items[i].Entity,
			At:         at,
		})
	}
	h.journal.Import(memory.JournalSnapshot{Seq: int64(len(eps)), Episodes: eps})
	h.log.Info("memory: bootstrapped journal from mesa", "episodes", len(eps))
}

// parseProvenance splits a mesa episode provenance "kind@RFC3339" into its parts.
func parseProvenance(p string) (kind string, at int64) {
	kind, ts, ok := strings.Cut(p, "@")
	if !ok {
		return p, 0
	}
	if t, err := time.Parse(time.RFC3339, ts); err == nil {
		at = t.Unix()
	}
	return kind, at
}

// cleanText strips RSC colour codes and trims a server/chat string for storage.
func cleanText(s string) string { return strings.TrimSpace(stripColors(s)) }

// article returns the indefinite article ("a"/"an") for a noun, by first letter.
func article(noun string) string {
	if noun == "" {
		return "a"
	}
	switch noun[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an"
	}
	return "a"
}
