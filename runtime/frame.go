package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/event"
	"github.com/gen0cide/westworld/proto/v235"
	"github.com/gen0cide/westworld/world"
)

// sleepWord is the hardcoded answer to the fatigue sleep-screen captcha
// on this OpenRSC server. The server's CaptchaGenerator falls back to a
// prerendered image of the word "asleep" and sets
// player.setSleepword("asleep") when no prerendered sleepword set is
// loaded (CaptchaGenerator.generateRSCLCaptcha — CaptchaGenerator.java:
// 79-80). In that fallback, SleepHandler.process has
// knowCorrectWord=false and accepts ANY typed word ("we won't check
// that" — SleepHandler.java:47-49), so "asleep" always passes. No OCR
// of the captcha bitmap required.
//
// OPERATOR DECISION (Alex, 2026-06-10): the word is always "asleep"
// on our server and stays HARDCODED — no captcha archive will be
// installed (conf/server/data/sleepwords/ is kept empty, so the
// server accepts any word). The 194-rejection retry below exists
// purely as a safety net, not as a supported configuration.
const sleepWord = "asleep"

// sleepAnswerFallback is the safety timeout for the sleepword answer:
// if the drain-complete signal (a SEND_SLEEP_FATIGUE of 0) never
// arrives, answer anyway rather than stay trapped on the sleep screen.
// The slowest legitimate drain is a sleeping bag from full fatigue:
// 150000 at 8400/tick ≈ 18 ticks ≈ 12s — 30s clears that with margin,
// so in practice this only fires when the server stops sending 244s.
const sleepAnswerFallback = 30 * time.Second

// handleFrame decodes a frame, applies it to world state, publishes
// the event(s).
//
// Some opcodes produce MULTIPLE events from one packet (UpdatePlayers,
// PlayerCoords with nearby players, NpcCoords). Those are special-cased
// here. Single-event opcodes flow through DecodeInbound.
func (h *Host) handleFrame(f v235.Frame) {
	switch f.Opcode {
	case v235.InSendPlayerCoords:
		own, nearby, err := v235.DecodePlayerCoords(f.Payload)
		if err != nil {
			h.log.Warn("decode playercoords", "err", err)
			return
		}
		h.world.Apply(own)
		h.bus.Publish(own)
		for _, np := range nearby {
			// Apply to world.Players so position queries
			// (world.players.find(...).position) resolve to the
			// most-recent observed coords. Previously only
			// published — world.Players never got the update, so
			// `target.position` returned (0, 0) and walk_to to a
			// remote player walked to the world origin.
			h.world.Apply(np)
			h.bus.Publish(np)
		}
		return
	case v235.InSendUpdatePlayers:
		events, err := v235.DecodeUpdatePlayers(f.Payload)
		if err != nil {
			h.log.Warn("decode updateplayers", "err", err)
			return
		}
		for _, ev := range events {
			// For another player's appearance, diff combat level + worn
			// equipment against what we last saw and fire edge events
			// (player_level_changed / player_equipment_changed) — the raw
			// appearance packet re-sends periodically, so only an actual
			// change should wake a routine. Snapshot BEFORE Apply.
			if ap, ok := ev.(event.OtherPlayerAppearance); ok {
				prev, had := h.world.Players.Get(ap.PlayerIndex)
				h.world.Apply(ev)
				h.bus.Publish(ev)
				if had && ap.HasCombat && prev.HasAppearanceCombat && prev.CombatLevel != ap.CombatLevel {
					h.bus.Publish(event.PlayerLevelChanged{
						PlayerIndex: ap.PlayerIndex, Name: ap.Name,
						OldLevel: prev.CombatLevel, NewLevel: ap.CombatLevel,
					})
				}
				if had && ap.HasWorn && prev.HasEquip {
					// One event per human slot that actually changed, each
					// naming the slot so the handler gets good information.
					for _, slot := range h.diffEquip(prev.EquipBySlot, ap.WornSprites) {
						h.bus.Publish(event.PlayerEquipmentChanged{PlayerIndex: ap.PlayerIndex, Name: ap.Name, Slot: slot})
					}
				}
				continue
			}
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InUpdateNpc:
		events, err := v235.DecodeUpdateNpcs(f.Payload)
		if err != nil {
			h.log.Warn("decode updatenpcs", "err", err)
			return
		}
		for _, ev := range events {
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InNpcCoords:
		pos := h.world.Self.Position()
		// Snapshot the ordered local-NPC list BEFORE applying events:
		// the opcode-79 update records are positional against the
		// server's localNpcs list, which this mirror tracks. Apply()
		// then mutates the order (prune on REMOVE, append on new).
		order := h.world.Npcs.Order()
		events, localCount, err := v235.DecodeNpcCoords(f.Payload, pos.X, pos.Y, order)
		if err != nil {
			h.log.Warn("decode npccoords", "err", err)
			return
		}
		// Tier-1 anomaly assertion (docs/lang/protocol-debug-strategy.md):
		// the opcode-79 existing-NPC records are positional against the
		// server's localNpcs list, which `order` mirrors. If their lengths
		// diverge, our slot->index map has desynced and every positional
		// update from here is misattributed (NPCs silently vanish from
		// world.npcs). Log loudly so the desync is caught at its source
		// instead of surfacing as a mysterious perception gap.
		if localCount != len(order) {
			h.log.Warn("npccoords order desync: localNpcs count != our order list — NPC tracking corrupted",
				"server_local_count", localCount,
				"our_order_len", len(order),
			)
		}
		for _, ev := range events {
			h.world.Apply(ev)
			h.bus.Publish(ev)
		}
		return
	case v235.InSleepScreen:
		// SEND_SLEEPSCREEN (opcode 117): the fatigue sleep-screen
		// captcha is up. We apply+publish the SleepScreenAppeared event
		// (sets self.is_sleeping = true) but do NOT answer yet: the
		// server only drains fatigue per game tick WHILE we are asleep
		// (Player.startSleepEvent: bed −42000/tick, bag −8400/tick of
		// 150000) and a correct answer wakes us IMMEDIATELY, committing
		// whatever the provisional value is (handleWakeup: fatigue =
		// sleepStateFatigue). Answering the instant the screen appeared
		// woke us after ~0 ticks — "You wake up - feeling refreshed"
		// with fatigue unchanged, fleet-wide (soak-retro cause #2). The
		// answer is sent from onSleepFatigue once the drain reaches 0
		// (or by the fallback timer if that signal never lands).
		ev := event.SleepScreenAppeared{ImageBytes: len(f.Payload)}
		h.world.Apply(ev)
		h.bus.Publish(ev)
		h.onSleepScreen()
		return
	case v235.InSleepFatigue:
		// SEND_SLEEP_FATIGUE (opcode 244): the provisional fatigue value
		// draining once per game tick while the sleep screen is up.
		// Published for observability (kind "sleep_fatigue_update"; NOT
		// applied to world.Self — it only commits on a successful wake)
		// and fed to the answer flow: at 0 the drain is complete and the
		// sleepword is finally worth answering.
		ev, err := v235.DecodeInbound(f, nil)
		if err != nil {
			h.log.Warn("decode sleepfatigue", "err", err)
			return
		}
		sf, ok := ev.(event.SleepFatigueUpdate)
		if !ok {
			return
		}
		h.bus.Publish(sf)
		h.onSleepFatigue(sf.Value)
		return
	case v235.InStopSleep:
		// SEND_STOPSLEEP (opcode 84, no payload): the server woke us —
		// the sleep word was correct (or sleep otherwise ended). Clear
		// the sleep state (self.is_sleeping = false) and reset the
		// answer flow. The committed fatigue value arrives separately
		// as a FatigueUpdate packet (handleWakeup → sendFatigue).
		h.onStopSleep()
		ev := event.SleepEnded{}
		h.world.Apply(ev)
		h.bus.Publish(ev)
		return
	case v235.InSleepwordIncorrect:
		// SEND_SLEEPWORD_INCORRECT (opcode 194): our answer was rejected.
		// On the current server config (no prerendered sleepword archive)
		// any word passes, so seeing this means the config changed under
		// us. Not a trap: the server keeps us asleep and re-sends a fresh
		// SEND_SLEEPSCREEN after a short delay, and the InSleepScreen case
		// re-arms the answer flow (drain already complete ⇒ immediate
		// retry). Warn loudly — sleep can't restore fatigue until the
		// client learns real words.
		h.log.Warn("sleepword rejected — server now checks real captcha words? retrying on next sleep screen")
		h.bus.Publish(event.SleepwordIncorrect{})
		return
	}

	// Single-event opcodes.
	ev, err := v235.DecodeInbound(f, h.isStackableItem)
	if err != nil {
		h.log.Warn("decode error",
			"opcode", fmt.Sprintf("0x%02x (%d)", f.Opcode, f.Opcode),
			"err", err,
		)
		return
	}
	if ev == nil {
		return
	}
	// Log every player-facing server message at INFO so each one is
	// visible in cradle stdout — the server's response to whatever a
	// routine just did becomes observable instead of being silently
	// discarded. This is the single chokepoint: all message-bearing
	// opcodes (131 server/chat, 120 PM, 222/89 dialog box, 245 menu)
	// arrive through this single-event path. One tidy line per message;
	// the runner sweep greps these "server msg" lines to explain a
	// non-PASS scenario. See logServerMessage.
	h.logServerMessage(ev)
	// Snapshot inventory counts before Apply so we can emit synthetic
	// ItemGained events afterward. Cheap (map iteration) and only
	// runs when the event might change inventory. Skipping for
	// non-inventory events keeps the hot path clean.
	var preCounts map[int]int
	// preEquip snapshots own worn gear per human slot before an inventory
	// Apply, so we fire one equipment_changed(slot) per slot that changed.
	var preEquip map[string]int
	switch ev.(type) {
	case event.InventorySnapshot, event.InventorySlotUpdate:
		preCounts = h.inventoryCounts()
		preEquip = h.selfEquipSnapshot()
	}
	// Snapshot per-skill xp before Apply for xp-bearing events so we
	// can emit synthetic XPGain deltas afterward (#119). The raw
	// packets carry the NEW TOTAL, not a delta — only a diff yields
	// the gain. Same edge-detection pattern as ItemGained.
	var preXP, preMax map[int]int
	switch ev.(type) {
	case event.StatUpdate, event.StatsSnapshot, event.ExperienceGain:
		preXP = h.skillXPSnapshot()
		preMax = h.skillMaxSnapshot()
	}
	// Snapshot the engaged target's pre-Apply health so we can detect
	// the death edge after NpcDamage lands (#119 target_died / npc_killed).
	preTargetAlive := false
	if d, ok := ev.(event.NpcDamage); ok && d.NpcIndex == int(h.lastAttackedNpcIndex.Load()) {
		preTargetAlive = h.engagedTargetAlive()
	}
	changed := h.world.Apply(ev)
	if changed {
		h.log.Debug("world updated", "by", ev.Kind())
	}
	if preCounts != nil {
		h.emitItemGainedDeltas(preCounts)
	}
	if preXP != nil {
		h.emitXPGainDeltas(preXP)
	}
	if preMax != nil {
		h.emitLevelUpDeltas(preMax)
	}
	if preEquip != nil {
		for name, id := range h.selfEquipSnapshot() {
			if preEquip[name] != id {
				h.bus.Publish(event.EquipmentChanged{Slot: name})
			}
		}
	}
	if d, ok := ev.(event.NpcDamage); ok && d.NpcIndex == int(h.lastAttackedNpcIndex.Load()) {
		h.emitTargetDeathEdge(d.NpcIndex, preTargetAlive)
	}
	// Tally combat rounds for the anti-kite retreat gate (#r3-retreat).
	// Each combat tick produces a type-2 damage update on a participant:
	// the target's hp drops when we hit it (NpcDamage on the engaged
	// index) and our own hp drops when it hits us (OtherPlayerDamage on
	// self). Either is one round elapsed. We count both — the server's
	// gate is the OPPONENT's hits-made, but in melee the exchange is
	// near-simultaneous each tick, so counting any damage event on the
	// engagement gives a faithful round estimate. See noteCombatRound.
	h.noteCombatRound(ev)
	h.bus.Publish(ev)
}

// onSleepScreen arms the sleepword answer flow for a freshly shown
// sleep screen. Normal path: stay asleep and let onSleepFatigue answer
// when the drain completes. Two special cases:
//   - the drain already completed this sleep (lastSleepFatigue == 0):
//     this is a re-sent screen after a SleepwordIncorrect — no further
//     244s will come (the server's drain event stopped at 0), so answer
//     immediately (the retry).
//   - the drain signal never lands: the fallback timer answers after
//     sleepFallbackAfter so we can't be trapped on the screen.
func (h *Host) onSleepScreen() {
	h.sleepMu.Lock()
	h.sleepAnswerPending = true
	drained := h.lastSleepFatigue == 0
	if h.sleepFallback != nil {
		h.sleepFallback.Stop()
		h.sleepFallback = nil
	}
	if !drained {
		h.sleepFallback = time.AfterFunc(h.sleepFallbackAfter, func() {
			h.log.Warn("sleep drain signal never completed — answering sleepword anyway",
				"after", h.sleepFallbackAfter)
			h.answerSleepWord("fallback timeout")
		})
	}
	h.sleepMu.Unlock()
	if drained {
		h.answerSleepWord("retry — drain already complete")
	}
}

// onSleepFatigue records a SEND_SLEEP_FATIGUE drain report and answers
// the captcha once the drain reaches 0 — the whole point of sleeping:
// the server commits this provisional value to real fatigue only at the
// successful wake our answer triggers.
func (h *Host) onSleepFatigue(v int) {
	h.sleepMu.Lock()
	h.lastSleepFatigue = v
	h.sleepMu.Unlock()
	if v == 0 {
		h.answerSleepWord("fatigue drained to 0")
	}
}

// onStopSleep resets the answer flow when the server wakes us
// (SEND_STOPSLEEP) — successful or not, the screen is gone.
func (h *Host) onStopSleep() {
	h.sleepMu.Lock()
	h.sleepAnswerPending = false
	h.lastSleepFatigue = -1
	if h.sleepFallback != nil {
		h.sleepFallback.Stop()
		h.sleepFallback = nil
	}
	h.sleepMu.Unlock()
}

// answerSleepWord sends SLEEPWORD_ENTERED(sleepWord) exactly once per
// sleep screen (the pending flag is consumed under the lock, so the
// drain path and the fallback timer can't double-send).
func (h *Host) answerSleepWord(reason string) {
	h.sleepMu.Lock()
	if !h.sleepAnswerPending {
		h.sleepMu.Unlock()
		return
	}
	h.sleepAnswerPending = false
	if h.sleepFallback != nil {
		h.sleepFallback.Stop()
		h.sleepFallback = nil
	}
	h.sleepMu.Unlock()
	if err := h.sendSleepWord(); err != nil {
		h.log.Warn("answer sleep word", "err", err)
	} else {
		h.log.Debug("answered sleep word", "word", sleepWord, "reason", reason)
	}
}

// noteCombatRound advances the engaged round counter on each combat
// damage tick while we have an active engagement, so retreat() can
// wait until the authentic 3-round anti-kite window has elapsed
// instead of blindly poking the server. Only damage involving our
// current engagement target counts; damage from/to other entities in
// view is ignored.
func (h *Host) noteCombatRound(ev event.Event) {
	// Resolve self index OUTSIDE the combat lock (it takes the world lock) to keep
	// the two locks from nesting.
	selfIdx := h.world.Players.SelfIndex()
	h.combatMu.Lock()
	defer h.combatMu.Unlock()
	idx := h.combatRoundTarget
	if idx == 0 {
		return // no engagement we're counting rounds for
	}
	switch e := ev.(type) {
	case event.NpcDamage:
		// Our hit landed on the engaged NPC (its hp changed).
		if e.NpcIndex == idx {
			h.combatRounds++
			h.outgoingHits++ // one-directional "we are swinging" signal
		}
	case event.OtherPlayerDamage:
		// our own index == we took a hit this round (the engaged entity
		// hit us). Closest analogue to the server's opponent.getHitsMade().
		if e.PlayerIndex == selfIdx {
			h.combatRounds++
		}
	}
}

// logServerMessage emits one INFO line for any inbound event that
// carries player-facing text, so every message the server sends is
// visible in cradle stdout (and greppable by the scenario sweep).
//
// RSC funnels almost all action feedback through opcode 131
// (SEND_SERVER_MESSAGE) — "You can't retreat for 3 rounds.", "The
// door is locked.", "Nothing interesting happens." — which decodes to
// SystemMessage (no sender) or ChatReceived (with sender). Dialog
// boxes (222/89 -> NpcDialogText), menus (245 -> NpcDialog) and PMs
// (120 -> PrivateMessage) round out the player-facing set. We tag each
// with its kind so a reader can tell server text from NPC speech.
//
// Kept deliberately uniform ("server msg" + kind + text) so the
// runner's failure dump can grep one stable token.
func (h *Host) logServerMessage(ev event.Event) {
	switch e := ev.(type) {
	case event.SystemMessage:
		h.log.Info("server msg", "kind", "server", "text", e.Message)
	case event.ChatReceived:
		// Opcode-131 with a sender: a public chat line the server
		// relayed (or, on this server, system text attributed to a
		// name). Include the sender for context.
		h.log.Info("server msg", "kind", "chat", "from", e.Speaker, "text", e.Message)
	case event.PrivateMessage:
		h.log.Info("server msg", "kind", "pm", "from", e.Sender, "text", e.Message)
	case event.NpcDialogText:
		h.log.Info("server msg", "kind", "dialog", "text", e.Text)
	case event.NpcDialog:
		h.log.Info("server msg", "kind", "menu", "options", strings.Join(e.Options, " | "))
	}
}

// skillXPSnapshot captures {skill_id -> total_xp} for all 18 skills.
// Used to diff xp before/after an xp-bearing packet so emitXPGainDeltas
// can publish synthetic XPGain events with the positive delta.
func (h *Host) skillXPSnapshot() map[int]int {
	snap := make(map[int]int, world.NumSkills)
	for id := 0; id < world.NumSkills; id++ {
		snap[id] = h.world.Self.SkillXP(id)
	}
	return snap
}

// emitXPGainDeltas compares the pre-Apply xp snapshot to the current
// per-skill xp and publishes one event.XPGain per net-positive delta.
// Routines subscribe via `on xp_gain(skill, amount) { ... }` and
// filter on the skill name. Mirrors emitItemGainedDeltas.
func (h *Host) emitXPGainDeltas(prev map[int]int) {
	for id := 0; id < world.NumSkills; id++ {
		total := h.world.Self.SkillXP(id)
		delta := total - prev[id]
		if delta > 0 {
			h.bus.Publish(event.XPGain{Skill: event.SkillID(id), Amount: delta, Total: total})
		}
	}
}

// skillMaxSnapshot captures the per-skill BASE (max) level before an
// Apply so emitLevelUpDeltas can detect a level increase afterward.
func (h *Host) skillMaxSnapshot() map[int]int {
	snap := make(map[int]int, world.NumSkills)
	for id := 0; id < world.NumSkills; id++ {
		snap[id] = h.world.Self.SkillMax(id)
	}
	return snap
}

// emitLevelUpDeltas publishes a LevelUp for each of our own skills whose
// base level rose since the snapshot — the edge that powers
// `on level_up(skill, new_level)`. Mirrors emitXPGainDeltas.
func (h *Host) emitLevelUpDeltas(prev map[int]int) {
	for id := 0; id < world.NumSkills; id++ {
		now := h.world.Self.SkillMax(id)
		if now > prev[id] {
			h.bus.Publish(event.LevelUp{Skill: event.SkillID(id), NewLevel: now})
		}
	}
}

// engagedTargetAlive reports whether the currently-engaged NPC target
// (lastAttackedNpcIndex) is known to have hitpoints remaining. True
// when we have a health reading (HasHits) and CurHits > 0, or when we
// have no reading yet (unknown defaults to "alive" so a first 0-hits
// reading still counts as a death edge). Used by the target_died
// edge detector to avoid double-firing once the NPC is already dead.
func (h *Host) engagedTargetAlive() bool {
	idx := int(h.lastAttackedNpcIndex.Load())
	if idx == 0 {
		return false
	}
	rec, ok := h.world.Npcs.Get(idx)
	if !ok {
		return false
	}
	if !rec.HasHits {
		return true // no reading yet — treat as alive
	}
	return rec.CurHits > 0
}

// emitTargetDeathEdge publishes the synthetic TargetDied event when
// our engaged target's health transitions from alive (>0 or unknown)
// to dead (CurHits == 0). Called after Apply landed the NpcDamage, so
// world.Npcs holds the post-hit reading. `wasAlive` is the pre-Apply
// liveness — the edge only fires on alive→dead, never on a repeated
// 0-hits packet, so `on target_died` / `on npc_killed` fire once.
func (h *Host) emitTargetDeathEdge(npcIndex int, wasAlive bool) {
	if !wasAlive {
		return
	}
	rec, ok := h.world.Npcs.Get(npcIndex)
	if !ok {
		return
	}
	if !rec.HasHits || rec.CurHits != 0 {
		return
	}
	// The engaged target died: combat is over, so clear the anti-kite
	// round tracking (#r3-retreat). A subsequent attack() on a new
	// target re-arms it; leaving stale state would make the next
	// engagement think rounds had already elapsed.
	h.combatMu.Lock()
	if h.combatRoundTarget == npcIndex {
		h.combatRoundTarget = 0
		h.combatRounds = 0
		h.outgoingHits = 0
		h.combatStartedAt = time.Time{}
	}
	h.combatMu.Unlock()
	h.bus.Publish(event.TargetDied{NpcIndex: npcIndex, TypeID: rec.TypeID})
}

// inventoryCounts builds a snapshot of {item_id -> total count}
// from the current world.Inventory state. Used to compute item_gained
// deltas across an inventory packet apply.
func (h *Host) inventoryCounts() map[int]int {
	counts := map[int]int{}
	for _, s := range h.world.Inventory.Slots() {
		counts[s.ItemID] += s.Amount
		if s.Amount == 0 {
			// Unstackable slots have Amount=1 baked in by the
			// decoder; defensive in case decoder reports 0.
			counts[s.ItemID] += 1
		}
	}
	return counts
}

// emitItemGainedDeltas compares the pre-Apply inventory snapshot
// to the current state and publishes one event.ItemGained for
// each net-positive delta. Routines subscribe via
// `on item_gained(item_id, count) { ... }`.
//
// Net-negative deltas don't fire — eat/drop/deposit each have
// their own surface. We could add ItemLost as a follow-up if a
// routine needs it.
func (h *Host) emitItemGainedDeltas(prev map[int]int) {
	cur := h.inventoryCounts()
	for id, curCount := range cur {
		delta := curCount - prev[id]
		if delta > 0 {
			h.bus.Publish(event.ItemGained{ItemID: id, Count: delta})
		}
	}
}
