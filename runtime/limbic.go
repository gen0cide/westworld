package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/limbic"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

const (
	// limbicLedgerKey is the durable home of the trust ledger. The
	// "relationship:" namespace makes the memory layer write it back to local
	// disk now and mirror it to mesa later. The leading underscore keeps it out
	// of any real player's relationship key.
	limbicLedgerKey = "relationship:_ledger"
	// limbicFlushInterval is how often the ledger is persisted while running.
	limbicFlushInterval = 30 * time.Second
	// duelFightWindow caps how long a "sporting fight" window stays open if no
	// respawn StatsSnapshot ever arrives to clear it (C2). A duel fight resolves in
	// at most a couple of minutes; this is a safety expiry so a stale window can
	// never suppress a genuine later PK grievance.
	duelFightWindow = 3 * time.Minute
)

// SetAffectBaseline resets the mood vector to a persona-derived baseline (the
// values affect decays back toward). Called when a persona loads (pre-Run, safe)
// AND on the genesis path AFTER Run has started the limbic/pearl/director readers.
// It mutates the existing *Affect IN PLACE under its mutex (Affect.SetBaseline)
// rather than reassigning h.affect — a whole-pointer swap raced the field readers
// (C1: limbic OnXPGain/OnDeath + pearl/decision-cache/director Snapshot all hold
// the pointer concurrently). h.affect is non-nil from New, so this never allocs.
func (h *Host) SetAffectBaseline(stress, confidence, valence float64) {
	h.affect.SetBaseline(stress, confidence, valence)
}

// runLimbic is the host's System-1 affect path: a bus-subscriber goroutine
// (sibling to heartbeatLoop, no tick) that deterministically folds game events
// into the mood vector and the trust ledger. It never calls the LLM and never
// sends packets — it only updates in-RAM limbic state that the pearl engine and
// relation_with read. It restores the ledger on start and persists it on a
// cadence + on exit. Started by Run; exits when ctx is cancelled.
func (h *Host) runLimbic(ctx context.Context) {
	h.loadLimbic(ctx)    // warm-start the trust ledger from local bbolt (fast path)
	h.loadKnowledge(ctx) // warm-start the world-knowledge ledger (same spine)
	h.loadGoalGraph(ctx) // warm-start the intention graph (same spine)
	if h.ledger != nil && len(h.ledger.All()) == 0 {
		h.bootstrapLedgerFromMesa(ctx) // cold start: reconstitute from mesa (authoritative)
	}
	if h.knowledge != nil {
		// Cold-start: the consolidation cron's distilled beliefs warm-start a fresh/
		// restarted host. The no-clobber invariant (don't overwrite a host that already
		// learned locally) is now SELF-ENFORCED inside bootstrapKnowledgeFromMesa (L9);
		// the outer All()==0 check was caller convention only and has moved inside.
		h.bootstrapKnowledgeFromMesa(ctx)
	}
	if h.goalGraph != nil {
		// Cold-start: the insight cron's open-question closures / cross-entity chains
		// warm-start a fresh/restarted host. goalgraph.Import REPLACES wholesale, so the
		// no-clobber invariant (CRITICAL) is now SELF-ENFORCED inside
		// bootstrapGoalGraphFromMesa via an internal Empty() guard (L9) — not a caller
		// convention a second caller could forget.
		h.bootstrapGoalGraphFromMesa(ctx)
	}

	ch := h.bus.Subscribe("*", 256)
	flush := time.NewTicker(limbicFlushInterval)
	defer flush.Stop()
	for {
		select {
		case <-ctx.Done():
			// Final best-effort flush on a fresh bounded context (ctx is dead).
			fctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			h.flushLimbic(fctx)
			h.flushKnowledge(fctx)
			h.flushGoalGraph(fctx)
			cancel()
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			h.limbicHandle(ev)
			h.perceptionHandle(ev) // semantic ledgers + observation firehose (Phase 1 writers)
		case <-flush.C:
			// Bound the in-RAM cognition structures BEFORE serializing them (H10): both
			// the knowledge ledger and the goal graph otherwise grow monotonically, and
			// the flush re-serializes the whole blob to bbolt AND re-uploads it to mesa
			// every tick. Prune to the package caps so RAM + per-flush I/O stay bounded;
			// the persisted blob is the pruned set. Same family as gcLatches/gcSpeech.
			h.pruneCognition()
			// WARM re-import (H17/M17): BEFORE flushing the local graph/ledger UP, pull
			// the now server-reconciled state back DOWN and fold it in non-destructively
			// (ImportMerge, NOT the Empty-guarded cold-start replace). This lands the
			// crons' chains/closures/β-bumps in the host's LOCAL copy so the immediately-
			// following wholesale flush-up carries them too — instead of clobbering them
			// every tick. Distinct from the cold-start bootstrap (runLimbic, Empty-guarded).
			h.reimportKnowledgeFromMesa(ctx) // fold cron-reconciled beliefs into the local ledger
			h.reimportGoalGraphFromMesa(ctx) // fold cron chains/closures into the local graph
			h.flushLimbic(ctx)
			h.flushKnowledge(ctx)
			h.flushKnowledgeToMesa(ctx) // mirror locally-learned beliefs up (host→mesa)
			h.flushGoalGraph(ctx)
			h.flushGoalGraphToMesa(ctx) // mirror the local intention graph up (host→mesa)
			if h.reactive != nil {
				h.reactive.gcLatches(time.Now()) // decay reactive latches + evict idle windows (no new loop)
			}
			if h.speech != nil {
				h.speech.gcSpeech(time.Now()) // drop stale ask/teach cooldown entries (no new loop)
			}
		}
	}
}

// pruneCognition bounds the host's in-RAM cognition structures so they never grow
// without bound (H10) — the host-LIGHT backstop the package docs promise but that
// status-flips alone never deliver (done/abandoned goals + every distinct belief
// stay resident forever otherwise). Knowledge is capped per-subject + per-belief
// (Ledger.Prune, package caps); the goal graph drops its oldest TERMINAL nodes +
// dangling edges (PruneTerminal, DefaultGraphCap). Both are deterministic and
// O(1)-amortized at the flush cadence. Frozen under analysis-mode like the rest of
// the learning I/O (a dry-run must not mutate the structures the host resumes with).
func (h *Host) pruneCognition() {
	if h.AnalysisActive() {
		return
	}
	if h.knowledge != nil {
		h.knowledge.Prune()
	}
	if h.goalGraph != nil {
		h.goalGraph.PruneTerminal(0) // 0 ⇒ DefaultGraphCap
	}
}

// limbicHandle routes one event into affect / ledger updates. The mapping is
// deliberately conservative for v1 (clear, attributable signals only); combat
// damage→affect and trade/attack→trust land as those signals are threaded.
func (h *Host) limbicHandle(ev event.Event) {
	// Analysis-mode memory suspension: the operator override freezes all affect
	// (mood) AND trust-ledger writes — the host neither feels nor judges the world
	// while under takeover. One lock-free check at the single limbic chokepoint
	// (see analysis.go).
	if h.AnalysisActive() {
		return
	}
	switch e := ev.(type) {
	// --- affect (mood) ---
	case event.ExperienceGain:
		h.affect.OnXPGain(e.XP)
	case event.LevelUp:
		h.affect.OnLevelUp()
	case event.Death:
		h.affect.OnDeath()
		// Wilderness PK → GRIEVANCE, but ONLY in a genuine PK context and ONLY when
		// the wire lets us attribute a killer. A duel death is a sport, not a wrong —
		// gated by BOTH !inDuel() (the offer/setup window) AND !inSportingFight() (the
		// live fight window, since the duel record already went terminal at fight start
		// — C2). The v235 Death packet carries NO attacker, so we lean on the engaged-
		// target + skull heuristic; ambiguous ⇒ record NOTHING (better silent than
		// mis-attributed — the cardinal constraint).
		if !h.inDuel() && !h.inSportingFight() {
			if name, ok := h.pkKiller(); ok {
				h.ledger.Met(name)
				h.ledger.ObserveGrievance(name, 3.0) // a kill is a large, lasting grievance
				h.ledger.Tag(name, "ganked-me")
			} else {
				h.log.Info("limbic: pk death, attacker unattributable")
			}
		}
		// On the host's OWN death while pursuing a goal, model the failure in the
		// intention graph: <active-goal> --blocked_by--> <cause> + an enabling sub-goal
		// (§4 "failures become generative" — H19). A duel/sparring loss is a SPORT, not
		// a goal-blocking failure, so it is gated out the same way grievance is.
		if !h.inDuel() && !h.inSportingFight() {
			h.recordDeathOnGoalGraph()
		}

	case event.StatsSnapshot:
		// The full stats snapshot is the respawn signal: a duel fight has resolved, so
		// close the sporting-fight window (C2) — a later death is a genuine event again.
		h.combatMu.Lock()
		h.duelFightUntil = time.Time{}
		h.combatMu.Unlock()

	// --- relationships (familiarity + mild engagement signal) ---
	// All trust updates are ATTRIBUTED — they fire only on signals that carry a
	// counterparty NAME on the wire (chat/PM/trade/duel). Melee death is
	// deliberately NOT mapped: the v235 damage packet has no attacker, so a trust
	// delta there would be mis-attributed (the cardinal constraint).
	case event.ChatReceived:
		if e.Speaker != "" {
			h.ledger.Met(e.Speaker)
		}
	case event.OtherPlayerChat:
		// Nearby players' PUBLIC chat arrives index-based (opcode 234 / UpdatePlayers),
		// NOT through the named server-message ChatReceived path — so without this the
		// trust ledger never sees ordinary conversation and a host never comes to "know"
		// anyone it just talks/plays with. Resolve the index to a name and record the
		// encounter; skip our own chat echoed back as local chat (mis-attribution).
		if p, ok := h.world.Players.Get(e.PlayerIndex); ok && p.Name != "" && !strings.EqualFold(p.Name, h.opts.Username) {
			h.ledger.Met(p.Name)
		}
	case event.PrivateMessage:
		if e.Sender != "" {
			// Someone choosing to whisper you is a small positive social signal.
			h.ledger.Met(e.Sender)
			h.ledger.Observe(e.Sender, true, 0.2)
		}
	case event.TradeRequestReceived:
		if e.FromPlayerName != "" {
			// Choosing to trade with you is a mild positive engagement signal.
			h.ledger.Met(e.FromPlayerName)
			h.ledger.Observe(e.FromPlayerName, true, 0.2)
		}
	case event.DuelRequestReceived:
		if e.FromPlayerName != "" {
			// A duel offer is engagement, but adversarial — familiarity ONLY (the
			// duel=Met floor invariant: a consensual sport is never a trust hit).
			h.ledger.Met(e.FromPlayerName)
			h.lastDuelOpponent = e.FromPlayerName // attribution net for DuelClosed
		}
	case event.DuelOpened:
		// Both sides accepted — capture the opponent name (resolved from world) as
		// the DuelClosed attribution net. Met is already covered by the request path.
		if h.world != nil && h.world.Players != nil {
			if p, ok := h.world.Players.Get(e.OtherPlayerIndex); ok && p.Name != "" {
				h.lastDuelOpponent = p.Name
			}
		}
	case event.DuelClosed:
		// A COMPLETED duel is a consensual sport (§3.4): bump familiarity + AFFINITY
		// ("sparring partner") + tag — NEVER trust or grievance, even on a staked
		// loss (the stake was agreed). world.Apply runs before bus.Publish, so the
		// duel record's WithName is still readable here; the captured name is the net.
		name := h.lastDuelOpponent
		if h.world != nil && h.world.Duel != nil {
			if rec := h.world.Duel.Duel(); rec != nil && rec.WithName != "" {
				name = rec.WithName
			}
		}
		if e.Completed {
			// DuelClosed{Completed:true} fires when the FIGHT STARTS, not when it ends
			// (world.Apply already flipped the record to "completed"). Open the sporting-
			// fight window so a Death/projectile DURING the fight is treated as a duel
			// event, not a wrong (C2). Cleared on the respawn StatsSnapshot below.
			h.combatMu.Lock()
			h.duelFightUntil = time.Now().Add(duelFightWindow)
			h.combatMu.Unlock()
			if name != "" {
				h.ledger.Met(name)
				h.ledger.ObserveAffinity(name, +1.0)
				h.ledger.Tag(name, "sparring-partner")
			}
		}
	case event.OtherPlayerProjectile:
		// A ranged/magic projectile AIMED AT US outside a duel is a hostile PK — and,
		// unlike melee, the wire NAMES the caster (CasterIndex), so it is fully
		// attributable. Accrue GRIEVANCE (sustained fire compounds) + "attacked-me".
		if !h.inDuel() && !h.inSportingFight() && !e.VictimIsNpc && h.world != nil && h.world.Players != nil &&
			e.VictimPlayerIndex == h.world.Players.SelfIndex() {
			if p, ok := h.world.Players.Get(e.CasterIndex); ok && p.Name != "" {
				h.ledger.Met(p.Name)
				h.ledger.ObserveGrievance(p.Name, 0.5)
				h.ledger.Tag(p.Name, "attacked-me")
			}
		}
	case event.TradeConfirmShown:
		if e.OpponentName != "" {
			// A trade reaching the confirm screen is a real, good-faith exchange in
			// progress — split the signal across the two axes it actually touches:
			// TRUST (will deals be honest?) and AFFINITY (a smooth exchange is warm).
			h.lastTradeOpponent = e.OpponentName
			h.ledger.Met(e.OpponentName)
			h.ledger.Observe(e.OpponentName, true, 0.5)    // TRUST (existing)
			h.ledger.ObserveAffinity(e.OpponentName, +0.5) // a smooth exchange is warm
		}
	case event.TradeClosed:
		// A COMPLETED trade (both sides confirmed twice, goods exchanged) is the
		// outcome that grows the relationship's trade VOLUME. The close packet
		// carries no completion bit, so — mirroring dsl_events.go's TradeClosed
		// path — read the terminal phase off the live trade record (world.Apply
		// runs before bus.Publish, so the record + its offers are still readable
		// here). Attribute to the confirm-screen opponent captured in
		// lastTradeOpponent (the only place the wire NAMES the counterparty), and
		// accumulate the magnitude of goods that changed hands as value_traded.
		// A cancelled trade is NOT a completed outcome → no volume.
		completed := e.Completed
		var rec *world.TradeRecord
		if h.world != nil && h.world.Trade != nil {
			rec = h.world.Trade.Trade()
			if rec != nil && rec.Phase == "completed" {
				completed = true
			}
		}
		if completed && h.lastTradeOpponent != "" {
			h.ledger.ObserveValueTraded(h.lastTradeOpponent, tradeVolume(rec))
		}
	}
}

// tradeVolume is the magnitude of goods that changed hands in a completed trade —
// the total quantity across BOTH sides' offers. RSC carries no per-item price on
// the trade wire (TradeItem is id+amount only), so item COUNT is the honest,
// catalog-free proxy for "how much was traded" that feeds the relationship's
// monotone value_traded volume. A nil/empty record yields 0 (a no-op accumulate).
func tradeVolume(rec *world.TradeRecord) float64 {
	if rec == nil {
		return 0
	}
	var n int
	for _, it := range rec.MyOffer {
		n += it.Amount
	}
	for _, it := range rec.TheirOffer {
		n += it.Amount
	}
	return float64(n)
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

// flushLimbic persists the trust ledger to BOTH durability tiers: local bbolt
// (the host's fast warm-start) and mesa's Postgres mirror (the authoritative
// of-record + cold-start bootstrap source). Best-effort: a write failure is
// logged, never fatal. The ledger is AuthLocal, so the mesa side is a snapshot
// replace, not a merge.
func (h *Host) flushLimbic(ctx context.Context) {
	if h.ledger == nil {
		return
	}
	// Analysis-mode: no periodic ledger persistence (bbolt + mesa) under operator
	// override — a hard freeze on trust/affect I/O.
	if h.AnalysisActive() {
		return
	}
	rows := h.ledger.Export()
	if len(rows) == 0 {
		return
	}
	// Local durable mirror (bbolt) — survives a host restart with no mesa sync.
	if h.Memory != nil {
		if raw, err := json.Marshal(rows); err == nil {
			if err := h.Memory.Put(ctx, limbicLedgerKey, raw); err != nil {
				h.log.Warn("limbic: ledger flush failed", "err", err)
			}
		}
	}
	// Mesa mirror (up) — durable cross-session of-record + bootstrap source.
	if h.mesaMem != nil && h.mesaMem.Healthy() {
		if err := h.mesaMem.SyncRelationships(ctx, h.opts.Username, ledgerToRelationships(rows)); err != nil {
			h.log.Debug("limbic: ledger mesa-sync failed", "err", err)
		}
	}
}

// bootstrapLedgerFromMesa reconstitutes the trust ledger from mesa when there is
// no local state to warm-start from (fresh host, no bbolt, in-memory mode) — the
// authoritative bootstrap path, parallel to bootstrapJournalFromMesa. No-op when
// mesa is absent/unhealthy.
func (h *Host) bootstrapLedgerFromMesa(ctx context.Context) {
	if h.mesaMem == nil || !h.mesaMem.Healthy() || h.ledger == nil {
		return
	}
	cctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	rels, err := h.mesaMem.FetchRelationships(cctx, h.opts.Username)
	if err != nil || len(rels) == 0 {
		return
	}
	h.ledger.Import(relationshipsToLedger(rels))
	h.log.Info("limbic: bootstrapped trust ledger from mesa", "relationships", len(rels))
}

// ledgerToRelationships converts the ledger's snapshot rows to the wire form.
func ledgerToRelationships(rows []limbic.Entry) []mesaclient.Relationship {
	out := make([]mesaclient.Relationship, 0, len(rows))
	for _, e := range rows {
		out = append(out, mesaclient.Relationship{
			Name: e.Name, Alpha: e.Alpha, Beta: e.Beta, Encounters: e.Encounters, Tags: e.Tags,
			ValueTraded: e.ValueTraded,
			Affinity:    e.AffinitySum, Grievance: e.GrievanceSum,
		})
	}
	return out
}

// inDuel reports whether a consensual duel is currently in a non-terminal phase.
// Every grievance branch is gated on !inDuel(): a duel is a sport, never a wrong.
func (h *Host) inDuel() bool {
	return h.world != nil && h.world.Duel != nil && h.world.Duel.IsActive()
}

// inSportingFight reports whether the host is inside an open duel-FIGHT window
// (C2). The duel-UI phase goes terminal at fight START (DuelClosed{Completed:true}
// → Phase="completed"), so inDuel() is already false while the fight is live; this
// window — stamped on that event, cleared on the respawn StatsSnapshot, with a
// safety expiry — is what keeps a death/hostile projectile DURING a skulled duel
// from being mis-recorded as a wilderness PK gank. Guarded by combatMu.
func (h *Host) inSportingFight() bool {
	h.combatMu.Lock()
	defer h.combatMu.Unlock()
	return !h.duelFightUntil.IsZero() && time.Now().Before(h.duelFightUntil)
}

// recordDeathOnGoalGraph makes the host's OWN death GENERATIVE in the intention
// graph (§4 / §9 Phase 2 — H19): when it dies while pursuing a goal, write
// `<active-goal> --blocked_by--> <cause>` (the cause attributed from the engaged
// NPC / last damager, exactly as pkKiller-style logic already gathers) and spawn
// an enabling sub-goal `train-to-survive <cause>` so the failure becomes a
// concrete next step instead of relying on the Sonnet insight cron to emit the
// chain. Deterministic ids ⇒ a repeated death reinforces ONE node/edge set (no
// graph growth). No-op when there is no active goal or no attributable cause —
// better silent than a placeholder edge. Frozen learning under analysis is already
// covered by the caller's AnalysisActive() gate at the top of limbicHandle.
func (h *Host) recordDeathOnGoalGraph() {
	if h.goalGraph == nil {
		return
	}
	goal := h.activeGoalID()
	if goal == "" {
		return
	}
	cause := h.deathCause()
	if cause == "" {
		return // unattributable — don't write a generic placeholder edge
	}
	// Never re-open / re-block a goal the lifecycle already CLOSED (mirrors
	// markGoalBlocked's H1 guard): a done/abandoned goal stays terminal.
	if n, ok := h.goalGraph.Get(goal); ok &&
		(n.Status == goalgraph.StatusDone || n.Status == goalgraph.StatusAbandoned) {
		return
	}
	cid := "cause:" + cause
	h.goalGraph.Upsert(cid, goalgraph.KindState, cause, "")
	h.goalGraph.SetStatus(goal, goalgraph.StatusBlocked) // the goal is blocked by what killed us
	h.goalGraph.Link(goal, cid, goalgraph.RelBlockedBy)  // goal --blocked_by--> cause (the bear)
	// The enabling sub-goal: train to survive the cause (bear → combat).
	sid := "survive:" + cause
	h.goalGraph.Upsert(sid, goalgraph.KindSubgoal,
		fmt.Sprintf("Train and prepare so %s can no longer kill you (combat, gear, or a safer approach).", cause),
		goalgraph.StatusOpen)
	h.goalGraph.Link(sid, goal, goalgraph.RelEnables)  // train-to-survive --enables--> goal
	h.goalGraph.Link(goal, sid, goalgraph.RelRequires) // surfaces in the sub-goal slice
	h.log.Info("limbic: death blocked goal, spawned enabling sub-goal", "goal", goal, "cause", cause)
}

// activeGoalID returns the id of the host's current ACTIVE goal node (KindGoal,
// StatusActive), or "" when none. The director keeps exactly one goal active at a
// time; this lets the limbic goroutine attribute a death to it without reaching
// into per-turn director state. Newest-active wins on the (rare) tie.
func (h *Host) activeGoalID() string {
	if h.goalGraph == nil {
		return ""
	}
	var id string
	var at int64
	for _, n := range h.goalGraph.Nodes() {
		if n.Kind == goalgraph.KindGoal && n.Status == goalgraph.StatusActive && n.At >= at {
			id, at = n.ID, n.At
		}
	}
	return id
}

// deathCause attributes what killed the host, player-scoped first (a wilderness PK,
// via pkKiller) then the engaged/last-attacked NPC name (the common PvE case — the
// bear). Empty when nothing is attributable. Used only for the §4 death→blocked_by
// edge, so a best-effort name is enough (no skull/PK gating — a PvE bear is a
// legitimate blocker too).
func (h *Host) deathCause() string {
	if name, ok := h.pkKiller(); ok {
		return name
	}
	if h.world != nil && h.world.Players != nil {
		if self, ok := h.world.Players.Self(); ok && self.EngagedNpcIndex >= 0 {
			if name := h.npcNameByIndex(self.EngagedNpcIndex); name != "" {
				return name
			}
		}
	}
	if name := h.npcNameByIndex(int(h.lastAttackedNpcIndex.Load())); name != "" {
		return name
	}
	return ""
}

// pkKiller returns the best-attributable killer for a wilderness death, but ONLY
// in a genuine PK context — otherwise ok=false and the caller records nothing.
// Attribution is PLAYER-SCOPED ONLY: the server's own EngagedPlayerIndex ("which
// PLAYER I'm fighting"), then the last PLAYER we attacked (lastAttackedPlayerIndex).
// It deliberately does NOT fall back to combatRoundTarget — that holds an NPC scene
// index in the common PvE case, and RSC NPC/player indices share numeric ranges in
// SEPARATE stores, so looking an NPC index up in the player store blames whatever
// player happens to occupy that number (M10) — a skulled bystander on a PvE death
// would be wrongly tagged "ganked-me". The PK context proxy is the skull flag
// (SkullType==1): you only skull via PvP, so a skulled aggressor (or our own skull)
// stands in for a wilderness-zone map. The v235 Death packet has no attacker field,
// so this heuristic is the documented ceiling for melee attribution.
func (h *Host) pkKiller() (string, bool) {
	if h.world == nil || h.world.Players == nil {
		return "", false
	}
	idx := 0
	if self, ok := h.world.Players.Self(); ok && self.EngagedPlayerIndex >= 1 {
		idx = self.EngagedPlayerIndex
	}
	if idx == 0 {
		idx = int(h.lastAttackedPlayerIndex.Load()) // genuinely player-scoped fallback (NOT combatRoundTarget — M10)
	}
	if idx == 0 {
		return "", false
	}
	p, ok := h.world.Players.Get(idx)
	if !ok || p.Name == "" {
		return "", false
	}
	self, _ := h.world.Players.Get(h.world.Players.SelfIndex())
	if p.SkullType == 0 && self.SkullType == 0 {
		return "", false // not a PK context
	}
	return p.Name, true
}

// relationshipsToLedger converts wire relationships back to ledger import rows.
func relationshipsToLedger(rels []mesaclient.Relationship) []limbic.Entry {
	out := make([]limbic.Entry, 0, len(rels))
	for _, r := range rels {
		out = append(out, limbic.Entry{
			Name: r.Name, Alpha: r.Alpha, Beta: r.Beta, Encounters: r.Encounters, Tags: r.Tags,
			ValueTraded: r.ValueTraded,
			AffinitySum: r.Affinity, GrievanceSum: r.Grievance,
		})
	}
	return out
}
