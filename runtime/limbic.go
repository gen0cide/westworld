package runtime

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/limbic"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
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

// SetAffectBaseline resets the mood vector to a persona-derived baseline (the
// values affect decays back toward). Called once at startup when a persona loads,
// so a host starts in its persona's characteristic mood.
func (h *Host) SetAffectBaseline(stress, confidence, valence float64) {
	h.affect = limbic.NewAffect(stress, confidence, valence, 0)
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
	if h.knowledge != nil && len(h.knowledge.All()) == 0 {
		// Cold-start only (the len()==0 guard): the consolidation cron's distilled
		// beliefs warm-start a fresh/restarted host without clobbering a host that
		// already has local knowledge mid-run.
		h.bootstrapKnowledgeFromMesa(ctx)
	}
	if h.goalGraph != nil && h.goalGraph.Empty() {
		// Cold-start only (Empty()-guarded — CRITICAL): goalgraph.Import REPLACES
		// the graph wholesale, so an unguarded fetch would clobber a mid-run host's
		// live graph. The insight cron's open-question closures / cross-entity chains
		// warm-start a fresh/restarted host this way.
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
		// the wire lets us attribute a killer. A duel death is a sport, not a wrong
		// (gated by !inDuel). The v235 Death packet carries NO attacker, so we lean
		// on the engaged-target + skull heuristic; ambiguous ⇒ record NOTHING (better
		// silent than mis-attributed — the cardinal constraint).
		if !h.inDuel() {
			if name, ok := h.pkKiller(); ok {
				h.ledger.Met(name)
				h.ledger.ObserveGrievance(name, 3.0) // a kill is a large, lasting grievance
				h.ledger.Tag(name, "ganked-me")
			} else {
				h.log.Info("limbic: pk death, attacker unattributable")
			}
		}

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
		if e.Completed && name != "" {
			h.ledger.Met(name)
			h.ledger.ObserveAffinity(name, +1.0)
			h.ledger.Tag(name, "sparring-partner")
		}
	case event.OtherPlayerProjectile:
		// A ranged/magic projectile AIMED AT US outside a duel is a hostile PK — and,
		// unlike melee, the wire NAMES the caster (CasterIndex), so it is fully
		// attributable. Accrue GRIEVANCE (sustained fire compounds) + "attacked-me".
		if !h.inDuel() && !e.VictimIsNpc && h.world != nil && h.world.Players != nil &&
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
			Affinity: e.AffinitySum, Grievance: e.GrievanceSum,
		})
	}
	return out
}

// inDuel reports whether a consensual duel is currently in a non-terminal phase.
// Every grievance branch is gated on !inDuel(): a duel is a sport, never a wrong.
func (h *Host) inDuel() bool {
	return h.world != nil && h.world.Duel != nil && h.world.Duel.IsActive()
}

// pkKiller returns the best-attributable killer for a wilderness death, but ONLY
// in a genuine PK context — otherwise ok=false and the caller records nothing.
// Attribution order: the server's own EngagedPlayerIndex ("who I'm fighting"),
// then combatRoundTarget. The PK context proxy is the skull flag (SkullType==1):
// you only skull via PvP, so a skulled aggressor (or our own skull) stands in for
// a wilderness-zone map. The v235 Death packet has no attacker field, so this
// heuristic is the documented ceiling for melee attribution.
func (h *Host) pkKiller() (string, bool) {
	if h.world == nil || h.world.Players == nil {
		return "", false
	}
	idx := 0
	if self, ok := h.world.Players.Self(); ok && self.EngagedPlayerIndex >= 1 {
		idx = self.EngagedPlayerIndex
	}
	if idx == 0 {
		idx = h.combatRoundTarget
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
			AffinitySum: r.Affinity, GrievanceSum: r.Grievance,
		})
	}
	return out
}
