package runtime

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
)

// situation snapshots the live world + affect + recent on-screen messages.
//
// In analysis/freeze mode (h.AnalysisActive()) it is a PURE READ: the
// hypothetical dry-run path (analysisSituation → situation) must build the same
// Situation the autonomous director would WITHOUT mutating the live goal graph
// or journal — an operator "what would you do" must not create a phantom active
// goal node or flip an open_goal active in the in-RAM state the host resumes with
// (M16). Every goal-graph / journal write below is gated on !freeze; the snapshot
// reads (effectiveGoalView via effectiveGoal, the epistemic slice, hints) are not.
func (d *MesaDirector) situation(h *Host, last Outcome) *mesaclient.Situation {
	w := h.world
	pos := w.Self.Position()
	freeze := h.AnalysisActive()

	// Phase-1 goal-graph writer: make the standing objective a VISIBLE root node
	// (memory, not a planner). One-time, deterministic; guarded by Has so it costs
	// nothing on subsequent turns. The goal string is the node ID (stable, human-
	// readable, matches the inspector).
	if !freeze && h.goalGraph != nil {
		if g := d.effectiveGoal(h); g != "" && !h.goalGraph.Has(g) {
			h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
		}
	}

	// Stuck detection: if she hasn't moved (within 1 tile) across turns, her
	// approach isn't working — widen perception and push her to explore.
	if d.hasPrev && absInt(pos.X-d.prevX) <= 1 && absInt(pos.Y-d.prevY) <= 1 {
		d.stuckTurns++
	} else {
		d.stuckTurns = 0
	}
	d.prevX, d.prevY, d.hasPrev = pos.X, pos.Y, true
	// stuck = hasn't moved (position) OR the world has STALLED (no change to
	// position/fatigue/hp/inventory/xp for several turns — the no-progress loop the
	// HybridDirector flags via NoteStall, which position-only stuckTurns misses when
	// she paces in place or "succeeds" at a routine that changes nothing).
	stuck := d.stuckTurns >= 3 || d.worldStall >= antiStuckWorldStall

	// Failure streak (anti-stuck v0): repeated FAILED outcomes mean the current
	// approach isn't working even while the host moves. Only a real prior action
	// counts (the first turn / no-ops don't); any success resets it, so a working
	// grind is never flagged. A non-empty Label is this codebase's marker of a
	// real prior intent (every constructed Intent sets Label; the zero Intent{}
	// has Label==""), so we gate on the same test triggerFor uses for "start".
	// A BUDGET-EXPIRED turn is NEUTRAL: a grind that simply outlived the 2-minute
	// turn budget mid-work is not evidence of failure OR of success, so it must
	// neither feed the streak (soak retro #5 — drone3's 55-"failure" streak while
	// leveling fastest, then BLOCKED ordering it to abandon a WORKING approach)
	// nor reset/advance the success accounting. A no-progress expiry is still
	// caught independently by the world-stall detector (NoteStall).
	if hadAction := last.Intent.Label != ""; hadAction && !last.BudgetExpired {
		if last.OK() {
			d.failStreak = 0
			// Goal-graph recovery: if the goal was marked blocked by a prior
			// fail-streak, a real success un-blocks it (the only non-monotonic
			// write; safe because we only flip an EXISTING node back to active).
			// Skipped under freeze — a dry-run must not flip live node status (M16).
			if !freeze && h.goalGraph != nil {
				if g := d.effectiveGoal(h); h.goalGraph.Has(g) {
					if n, ok := h.goalGraph.Get(g); ok && n.Status == goalgraph.StatusBlocked {
						h.goalGraph.SetStatus(g, goalgraph.StatusActive)
						h.goalGraph.Untag(g, "spinning") // recovered: clear the transient spin marker (H5)
						// Un-block reads as MOVEMENT: nudge a frozen-0 goal off zero so
						// the inspector/planner see progress, not a stall (the enabler
						// flips done just below). Monotone — never lower an existing value.
						if n.Progress < enablerProgressFloor {
							h.goalGraph.SetProgress(g, enablerProgressFloor)
						}
					}
					// Phase-2a: a resolved enabler stops appearing in the 1a slice
					// (which filters StatusDone). Mark it done rather than deleting
					// it — host-side node deletion would be solver creep.
					if eid := "enabler:" + g; h.goalGraph.Has(eid) {
						h.goalGraph.SetStatus(eid, goalgraph.StatusDone)
					}
				}
			}

			// Spin detector (Phase-5a): a SUCCEEDING loop the fail-streak path can't
			// see — the host re-derives substantially the same plan turn after turn
			// while every micro-action "succeeds", so the goal never closes. Compare
			// the EXECUTED plan's content fingerprint to last turn's; the same one
			// running counts up. Computed only on OK turns (a recovering/changing goal
			// is never penalised); the trigger fires below, after priority resolution.
			fp := planFingerprint(last.Intent)
			if fp != 0 && fp == d.lastPlanFP {
				d.spinCount++
			} else {
				d.spinCount = 0
			}
			d.lastPlanFP = fp

			// Deterministic satisfaction net (Phase-5a): the one cheap OBSERVABLE the
			// brief calls for — an acquire-item goal whose item is already in hand is
			// done, with NO planner declaration (the literal pickaxe bug). Conservative
			// substring match only; non-acquire goals fall through to the planner's
			// declaration. The host READS the inventory and WRITES the status — it does
			// not parse or search. Skipped under freeze (M16). The log fires ONCE on
			// the open→done transition (markGoalDone is idempotent, the log line is
			// not): once the goal is Done, effectiveGoal advances to a successor or
			// returns "" (no successor), so a no-successor done goal is never re-checked
			// here — and the explicit not-already-Done guard prevents a re-log even if
			// it were (L5).
			if !freeze && h.goalGraph != nil {
				if g := d.effectiveGoal(h); g != "" && h.goalGraph.Has(g) && d.goalSatisfiedByInventory(h, g, w) {
					if n, ok := h.goalGraph.Get(g); ok && n.Status != goalgraph.StatusDone {
						d.markGoalDone(h, g)
						d.log.Info("goal completed (inventory-satisfied)", "goal", g)
					}
				}
			}
		} else {
			d.failStreak++
			d.spinCount, d.lastPlanFP = 0, 0 // a FAILING turn is the fail-streak path's job, not spin's
		}
	}

	inv := make([]string, 0, 8)
	for _, sl := range w.Inventory.Slots() {
		name := d.itemName(h, sl.ItemID)
		if sl.Amount > 1 {
			name = fmt.Sprintf("%d %s", sl.Amount, name)
		}
		if sl.Wielded {
			name += " (worn)"
		}
		inv = append(inv, name)
	}

	// Nearby NPCs by name, de-duplicated (the scene often has many copies).
	npcSeen := map[string]bool{}
	npcs := make([]string, 0, 8)
	for _, n := range w.Npcs.All() {
		name := d.npcName(h, n.TypeID)
		if name == "" || npcSeen[name] {
			continue
		}
		npcSeen[name] = true
		npcs = append(npcs, name)
	}

	players := make([]string, 0, 4)
	for _, p := range w.Players.All() {
		if p.Name != "" {
			players = append(players, p.Name)
		}
	}

	// Recent narrative — the cross-turn transcript (NPC speech, system messages,
	// items/levels gained) the director accumulates from the bus. This is how the
	// host remembers what an NPC just told it.
	recent := lastN(d.transcript, 18)

	stress, confidence, valence := h.affect.Snapshot()

	baseTrigger := triggerFor(last)
	trigger := baseTrigger

	// Displacement is the HIGHEST-priority context change (death / lure / teleport
	// / stairs): the conductor aborted the previous turn because the host was moved
	// unexpectedly. Consume it FIRST (consume-once, cleared by take) so the
	// anti-stuck graph writes below can be gated on !displaced. A displaced turn's
	// STUCK/BLOCKED/SPINNING signals are about the OLD plan/location that no longer
	// applies after the jump — marking the goal Blocked + "spinning" on a spin
	// that pre-dates the jump durably MIS-MARKS it (H2). So we peek the override
	// here and only APPLY the displacement trigger string at the very end.
	disp, displaced := h.displacement.take()
	if displaced {
		// A death/teleport is a hard context change: the prior failure streak AND
		// the spin state were about the OLD situation and no longer apply. Reset
		// them so a stale BLOCKED/SPINNING doesn't fire next turn on top of the
		// fresh re-orient, and so the spin counter never reaches threshold on
		// executions begun in a different location (H3).
		d.failStreak = 0
		d.spinCount, d.lastPlanFP = 0, 0
	}

	if !displaced && stuck {
		trigger = fmt.Sprintf("STUCK — you have not moved in %d turns; your current approach is NOT working, change it", d.stuckTurns)
	}
	// Anti-stuck v0: escalate on a failure streak. A hard streak means "abandon
	// the whole approach", not "retry a variation"; a soft streak just nudges.
	// Skipped entirely on a displaced turn: the durable BLOCKED graph write would
	// mis-mark the goal for a failure streak the jump invalidated (H2).
	if !displaced && d.failStreak >= antiStuckHardFails {
		trigger = fmt.Sprintf("BLOCKED — your last %d actions all FAILED. This approach is not working; ABANDON it entirely and do something fundamentally different (a different place, NPC, or activity). Do NOT retry a variation of the same thing.", d.failStreak)
		// Phase-1 goal-graph writer: a hard fail-streak is the host modelling its
		// own being-stuck. Record it as an OPEN QUESTION the goal is blocked_by, so
		// the stuck goal becomes visible AND a mesa distillation cron has a node to
		// answer. Deterministic qid ⇒ repeated BLOCKED states reinforce one node
		// (Upsert updates in place; no dup spam).
		// Guard the empty goal (like Hook 1): with no goal there is nothing to
		// block, and writing would create empty-ID nodes that corrupt the graph.
		if g := d.effectiveGoal(h); g != "" {
			d.markGoalBlocked(h, g, fmt.Sprintf("How do I make progress on %q? %d straight attempts failed.", g, d.failStreak))
		}
	} else if !displaced && d.failStreak >= antiStuckSoftFails {
		trigger = fmt.Sprintf("%s — and your last %d actions failed, so reconsider the approach rather than just retrying", trigger, d.failStreak)
	}
	// Spin detector (Phase-5a): a SUCCEEDING loop — the same plan N turns running
	// while micro-actions keep "succeeding". Fire only if no higher-priority
	// trigger already changed the base (STUCK and the fail-streak override above)
	// and the turn was not displaced (the displacement override is applied below;
	// the spin state has already been reset for a displaced turn). Mirrors the
	// BLOCKED contract: it nudges + writes the same blocked graph state, but it
	// does NOT itself judge the goal done — declaring done/abandoned stays the
	// planner's call. One-shot: reset so it re-arms only if the spinning resumes.
	if !displaced && trigger == baseTrigger && d.spinCount >= antiStuckSpinTurns &&
		!d.foragingInFlight(h) {
		// Never mark a CLOSED goal spinning (H1): when the only active goal is
		// done/abandoned with no successor, effectiveGoal returns "" and the guard
		// below skips the write; but if d.goal still names a terminal node, marking
		// it Blocked + "spinning" here would let the OK-turn un-block recovery
		// resurrect it into a loop. Skip the write (and the trigger) in that case.
		g := d.effectiveGoal(h)
		spinNode, hasSpinNode := goalgraph.Node{}, false
		if g != "" && h.goalGraph != nil {
			spinNode, hasSpinNode = h.goalGraph.Get(g)
		}
		terminal := hasSpinNode && (spinNode.Status == goalgraph.StatusDone || spinNode.Status == goalgraph.StatusAbandoned)
		if g != "" && h.goalGraph != nil && !terminal {
			trigger = "SPINNING — you have produced essentially the SAME plan " + strconv.Itoa(d.spinCount+1) + " turns running and nothing is changing. Either this GOAL IS ALREADY DONE (set goal_op:\"done\") or this approach will never finish it (set goal_op:\"abandoned\", or try a FUNDAMENTALLY different one)."
			d.markGoalBlocked(h, g, fmt.Sprintf("Spinning on %q: the same plan %d turns running with no change. Is it already done, or does it need a different approach?", g, d.spinCount+1))
			h.goalGraph.Tag(g, "spinning")
		}
		d.spinCount = 0 // one-shot; re-arms if spinning resumes
	}
	// Apply the displacement re-orient trigger LAST, so it wins over any STUCK/
	// BLOCKED/SPINNING text for THIS turn (the graph writes were gated off above).
	if displaced {
		if disp.cause == dispDeath {
			trigger = fmt.Sprintf("YOU DIED — you respawned at (%d, %d) and dropped what you were carrying. Your old plan and location no longer apply: take stock of where you are and what you still have, and decide fresh.",
				disp.toX, disp.toY)
		} else {
			trigger = fmt.Sprintf("DISPLACED — you were unexpectedly moved %d tiles from (%d, %d) to (%d, %d) (lured by a monster, a teleport, or stairs/ladder). Stop what you were doing, reassess where you are, and decide fresh.",
				disp.dist, disp.fromX, disp.fromY, disp.toX, disp.toY)
		}
	}
	// World-progress stall (HybridDirector.NoteStall): if nothing has changed for
	// several turns and there's no stronger re-orient trigger (death/displacement),
	// tell the planner plainly its approach is futile — and to NAME the unknown
	// blocking it rather than repeat (the seam 5b foraging later resolves).
	if trigger == "" && d.worldStall >= antiStuckWorldStall {
		trigger = fmt.Sprintf("STALLED — %d turns and NOTHING has changed (same position, fatigue, inventory, and XP). Your recent approach is accomplishing nothing. Do something FUNDAMENTALLY different. If you don't actually know HOW to make progress (e.g. how to cure fatigue, where to buy or find something), SAY SO and note the specific unknown — do NOT repeat the same plan.", d.worldStall)
		// On crossing the stall threshold, BLOCK the effective goal ONCE so the
		// host-side ASK/FORAGE drives engage: a no-progress "completes but the world
		// never changes" loop (the wrong-tool case) is treated like a failure — a
		// stalled acquire-item goal mints a where-to-buy question the forager travels
		// for, instead of the host quietly re-running a useless routine forever. Uses
		// the PURE effectiveGoalView (effectiveGoal mutates + is called below for
		// sit.Goal — don't double-advance). Idempotent + M8-safe; gated on !freeze;
		// fires once at the crossing (worldStall increments by 1 per stalled turn).
		if !freeze && d.worldStall == antiStuckWorldStall {
			if g := d.effectiveGoalView(h); g != "" && h.goalGraph != nil {
				d.markGoalBlocked(h, g, "stalled — no progress for several turns; finding what is missing")
			}
		}
	}

	sit := &mesaclient.Situation{
		HostID:  d.hostID,
		Goal:    d.effectiveGoal(h),
		Trigger: trigger,
		World: mesaclient.WorldSnapshot{
			X: pos.X, Y: pos.Y,
			HPCur: w.Self.HP(), HPMax: w.Self.MaxHP(),
			CombatLevel: w.Self.CombatLevel(),
			Fatigue:     float64(w.Self.Fatigue()) / 750.0,
			InvFree:     w.Inventory.FreeSlots(),
			InvSummary:  inv, NearbyNpcs: npcs, NearbyPlayers: players,
		},
		Recent: recent,
		Affect: mesaclient.Affect{Stress: stress, Confidence: confidence, Valence: valence},
	}
	hints := map[string]string{}
	// (No hints["displacement"]: the displacement re-orient already rides on
	// Trigger above, and mesad's act prompt never reads a "displacement" hint key
	// — a second copy here was dead/redundant (L6). Trigger is the one source.)
	// Scene perception: NPCs, doors/boundaries, and notable scenery WITH
	// coordinates — so she can see (and walk to / open) the door an instruction
	// refers to. Widen the radius when stuck so an out-of-view exit appears.
	radius := 16
	if stuck {
		radius = 32
		hints["explore"] = "You are STUCK. Stop repeating the same action/coordinates. Re-read the latest instruction's DIRECTION (e.g. \"northeast\"). In RSC coordinates north = smaller y and EAST = SMALLER x (x increases to the west) — so northeast = smaller x AND smaller y. Look through 'what you see around you' for a door/exit in that direction (compare its coordinates to yours) and walk_to it. If none is listed, EXPLORE: walk ~10 tiles that way (e.g. northeast = your_x-10, your_y-10) and look again. Do NOT reuse a door you already came through."
	}
	hints["scene"] = d.describeArea(h, radius)
	// The single most-recent server message — often a blocking/prerequisite
	// notice ("Speak to the controls guide before going through this door"). It
	// must not get buried in the transcript history.
	if sm := w.Recent.ServerMessage(); sm != nil {
		if t := strings.TrimSpace(stripColors(sm.Message)); t != "" {
			hints["latest_message"] = t
		}
	}
	if opts := w.Recent.DialogOptions(); opts != nil && len(opts.Options) > 0 {
		numbered := make([]string, len(opts.Options))
		for i, o := range opts.Options {
			numbered[i] = fmt.Sprintf("%d. %s", i+1, o)
		}
		hints["dialog_options"] = strings.Join(numbered, "   ")
	}
	// A real player's recent words, surfaced to the PLANNER for a few turns: if
	// it's a direction/instruction, she should act on it. (Replying is handled
	// separately by the social reflex.) Annotate the speaker with what she knows
	// of them (trust ledger) so a friend's request and a stranger's read
	// differently — judged by the RELATIONSHIP, not by keywords.
	if d.lastPlayerMsg != "" && d.playerMsgAge < 3 {
		pd := d.lastPlayerMsg
		if note := d.trustNote(h, d.lastPlayerName); note != "" {
			pd += "  [what you know of " + d.lastPlayerName + ": " + note + "]"
		}
		hints["player_directive"] = pd
	}
	d.playerMsgAge++
	// Self-context: what she did last turn + how it ended, so she doesn't repeat
	// a step that already succeeded.
	if last.Intent.Label != "" {
		hints["last_action"] = last.Intent.Label
		hints["last_result"] = last.Kind.String()
		if last.BudgetExpired {
			// The routine was cut off by the per-turn budget mid-work — tell the
			// planner the truth so it doesn't read a working grind as a failure.
			hints["last_result"] = "ran out of turn time mid-work (the turn budget expired — NOT a failure; the routine may have been making progress)"
		} else if last.Err != nil {
			hints["last_result"] = last.Kind.String() + ": " + last.Err.Msg
		}
		if last.Intent.Source != "" {
			hints["last_dsl"] = last.Intent.Source
		}
	}
	// Durable EPISODIC memory: record the standing objective (survives a restart
	// → feeds a future session-genesis) and recall the salient things she's done
	// this life into the Situation, so the planner builds on its own history
	// instead of re-deriving the world each tick. The SetObjective WRITE is skipped
	// under freeze (a dry-run must not overwrite the journal objective — M16); the
	// memory READ still surfaces.
	if h.journal != nil {
		if !freeze {
			h.journal.SetObjective(d.effectiveGoal(h))
		}
		if mem := d.memoryHint(h); mem != "" {
			hints["memory"] = mem
		}
	}
	// Proactive social presence: surface nearby players as people she MAY choose
	// to engage — greet, trade, help, follow — not only when chatted at. Each is
	// annotated with what she knows of them (trust ledger). Trade is first-class:
	// an offer is judged by the relationship + the actual deal, never by a word.
	if np := d.nearbyPlayersHint(h, pos.X, pos.Y); np != "" {
		hints["nearby_players"] = np
	}
	// Session-genesis attention ladder: the words/people she compiled at login as
	// worth noticing in chat (her name, friends, trade words, goal topics).
	if att := d.attentionHint(); att != "" {
		hints["attention"] = att
	}
	// Phase-2a reasoning layer: surface the host's INTENTION STRUCTURE (the active
	// goal's graph slice — blockers, sub-goals, downstream value), the RELEVANT
	// graded beliefs, the explicit UNKNOWNS, and the curiosity-weighted learning
	// priority — so the planner reasons over what it knows/doesn't, instead of
	// confidently acting from a flat goal string (the go_to("mining-site") problem).
	// READ + SURFACE only; the host does not solve.
	d.epistemicHints(h, d.effectiveGoal(h), npcs, players, hints)
	if len(hints) > 0 {
		sit.Hints = hints
	}
	return sit
}

// --- Phase-2a epistemic layer (read + surface; the host does not solve) -----

// Epistemic caps: the goal-graph slice + belief/unknown sets are BOUNDED so the
// surfaced context is O(degree of one node), independent of total graph size —
// the host READS and SURFACES, it never searches.
const (
	epBlockerCap  = 3    // blockers rendered in the intention slice
	epSubgoalCap  = 4    // open sub-goals rendered
	epServesCap   = 4    // downstream-value targets rendered
	epBeliefCap   = 6    // graded subjects rendered across known+unknown buckets
	epUnknownCap  = 5    // explicit unknowns rendered
	epConfFloor   = 0.5  // confidence >= this ⇒ "known"; (0, this) ⇒ "unknown"/low
	epCuriosityHi = 0.5  // dominant curiosity flavor magnitude that earns a detour cue
	epCuriosityLo = 0.35 // below this, a low-curiosity persona stays on task (no cue)
)

// epistemicHints traverses the goal graph + knowledge ledger for the active goal
// and renders a BOUNDED prose slice into hints. READ + SURFACE only — no node
// writes (those live in the failure hook). One BFS hop off the active goal (plus
// a single produces→serves hop for downstream value); hard caps throughout. It
// fills hints["intention"], hints["known"], hints["unknowns"], and
// hints["learning_priority"]; each is omitted when its list is empty, so the
// flat path (no goal graph / no persona) is byte-identical to before.
func (d *MesaDirector) epistemicHints(h *Host, g string, npcs, players []string, hints map[string]string) {
	if g == "" || h.goalGraph == nil {
		return
	}
	node, ok := h.goalGraph.Get(g)
	if !ok || node.Status == goalgraph.StatusDone || node.Status == goalgraph.StatusAbandoned {
		return // nothing worth surfacing for a finished/abandoned goal
	}

	// 1a. Goal-graph slice — one hop off the active goal.
	blockers := d.sliceBlockers(h, g) // [BLOCKS YOUR GOAL] labels (open questions etc.)
	subgoals := d.sliceSubgoals(h, g) // open requires-children → "to do this you need: …"
	serves := d.sliceServes(h, g)     // serves ∪ (produces→serves) → "doing this serves: …"
	coreBlocked := len(blockers) > 0 || node.Status == goalgraph.StatusBlocked

	if it := renderIntention(node, blockers, subgoals, serves); it != "" {
		hints["intention"] = it
	}

	// 1b/1c. Relevant graded beliefs (known) + explicit unknowns. Relevance =
	// subjects in front of the host this turn ∪ the labels its goal slice names —
	// precisely the set that prevents confidently-wrong action.
	relevant := dedupLower(append(append(append([]string{}, npcs...), players...), append(blockers, subgoals...)...))
	known, lowConf := d.relevantBeliefs(h, relevant, coreBlocked)
	if known != "" {
		hints["known"] = known
	}
	if un := renderUnknowns(blockers, lowConf); un != "" {
		hints["unknowns"] = un
	}

	// 1d. Curiosity-weighted explore/exploit bias (Deliverable 5), priority-gated
	// by whether an unknown BLOCKS the core goal. A "tangent" is genuinely
	// off-path downstream value (serves edges) ONLY — NOT the goal's own required
	// sub-goals, which are the core-path steps it NEEDS: counting them as tangents
	// nudged a curious host OFF its required steps with the "short detour" cue (L10).
	if lp := curiosityBias(h.curiosity, coreBlocked, len(serves) > 0); lp != "" {
		hints["learning_priority"] = lp
	}
}

// sliceBlockers returns the labels of the nodes blocking g (g --blocked_by--> X),
// newest-first, capped. These are core-blocking unknowns by construction.
func (d *MesaDirector) sliceBlockers(h *Host, g string) []string {
	bs := h.goalGraph.Blockers(g)
	sort.Slice(bs, func(i, j int) bool { return bs[i].At > bs[j].At })
	out := make([]string, 0, epBlockerCap)
	for _, b := range bs {
		if lbl := nodeLabel(b); lbl != "" {
			out = append(out, lbl)
			if len(out) >= epBlockerCap {
				break
			}
		}
	}
	return out
}

// sliceSubgoals returns the labels of g's OPEN requires-children (the steps the
// graph says this goal needs), capped. done/abandoned children are filtered.
func (d *MesaDirector) sliceSubgoals(h *Host, g string) []string {
	out := make([]string, 0, epSubgoalCap)
	for _, e := range h.goalGraph.Out(g, goalgraph.RelRequires) {
		n, ok := h.goalGraph.Get(e.To)
		if !ok || n.Status == goalgraph.StatusDone || n.Status == goalgraph.StatusAbandoned {
			continue
		}
		if lbl := nodeLabel(n); lbl != "" {
			out = append(out, lbl)
			if len(out) >= epSubgoalCap {
				break
			}
		}
	}
	return out
}

// sliceServes returns the downstream value of doing g: direct serves-targets
// (g --serves--> X) UNION the one-more-hop targets through produced outputs
// (g --produces--> p, p --serves--> X). Exactly one extra hop — never recurse
// further. Deduped + capped. This is the forward-edge cue (Deliverable 3) that
// lets the planner value intermediate steps.
func (d *MesaDirector) sliceServes(h *Host, g string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, epServesCap)
	push := func(id string) {
		n, ok := h.goalGraph.Get(id)
		if !ok {
			return
		}
		lbl := nodeLabel(n)
		key := strings.ToLower(lbl)
		if lbl == "" || seen[key] {
			return
		}
		seen[key] = true
		out = append(out, lbl)
	}
	for _, e := range h.goalGraph.Out(g, goalgraph.RelServes) {
		push(e.To)
	}
	for _, e := range h.goalGraph.Out(g, goalgraph.RelProduces) {
		for _, se := range h.goalGraph.Out(e.To, goalgraph.RelServes) { // one hop further only
			push(se.To)
		}
	}
	if len(out) > epServesCap {
		out = out[:epServesCap]
	}
	return out
}

// relevantBeliefs splits the host's graded beliefs about the relevant subjects
// into a rendered "known" block (confidence >= floor) and a list of low-confidence
// claim phrases (0 < conf < floor) routed to the unknowns block. Subjects are
// ranked by confidence×familiarity and capped. When the relevance key-set is empty
// AND the goal is blocked, it falls back to the host's most-familiar facts (a
// stuck host still gets oriented) — the only All() pass, and a rare one.
func (d *MesaDirector) relevantBeliefs(h *Host, relevant []string, coreBlocked bool) (known string, lowConf []string) {
	if h.knowledge == nil {
		return "", nil
	}
	var facts []knowledge.Fact
	for _, s := range relevant {
		if !h.knowledge.Known(s) {
			continue
		}
		if f := h.knowledge.Get(s); len(f.Beliefs) > 0 {
			facts = append(facts, f)
		}
	}
	if len(facts) == 0 && coreBlocked {
		facts = h.knowledge.All() // fallback: orient a stuck host with its strongest facts
	}
	sort.SliceStable(facts, func(i, j int) bool {
		return facts[i].Confidence*float64(facts[i].Familiar+1) > facts[j].Confidence*float64(facts[j].Familiar+1)
	})
	var b strings.Builder
	rendered := 0
	for _, f := range facts {
		if rendered >= epBeliefCap {
			break
		}
		if len(f.Beliefs) == 0 || f.Confidence <= 0 {
			continue
		}
		best := f.Beliefs[0] // sorted best-first
		if f.Confidence >= epConfFloor {
			line := fmt.Sprintf("- %s (%s): %s — %s, %s",
				f.Subject, orThing(f.Kind), best.Claim, confidenceWord(f.Confidence), seenPhrase(best.Provenance, f.Familiar))
			b.WriteString("\n" + line)
			rendered++
		} else {
			lowConf = append(lowConf, fmt.Sprintf("%s: %s (you're not sure — confidence is low)", f.Subject, best.Claim))
			rendered++
		}
	}
	if b.Len() == 0 {
		return "", lowConf
	}
	return b.String(), lowConf
}

// renderIntention formats the one-hop goal slice as a single prose hint, omitting
// any empty line. Returns "" when the slice carries no edge information (a bare
// root goal) so an early-game host gets no noise.
func renderIntention(node goalgraph.Node, blockers, subgoals, serves []string) string {
	if len(blockers) == 0 && len(subgoals) == 0 && len(serves) == 0 {
		return ""
	}
	label := node.Label
	if strings.TrimSpace(label) == "" {
		label = node.ID
	}
	var b strings.Builder
	state := "is active"
	switch {
	case node.Status == goalgraph.StatusBlocked || len(blockers) > 0:
		state = "is BLOCKED"
	case node.Status == goalgraph.StatusDone:
		state = "is DONE"
	}
	// Phase-5a: surface the goal's own PROGRESS so the planner sees momentum (or a
	// frozen 0) and is prompted to declare done / report progress, instead of
	// silently re-deriving the same plan forever (the live bug).
	prog := ""
	if node.Progress > 0 {
		prog = fmt.Sprintf(" (progress: %d%%)", int(node.Progress*100+0.5))
	}
	if slices.Contains(node.Tags, "spinning") {
		prog += " (you've been SPINNING on this — if it's done, say so)"
	}
	fmt.Fprintf(&b, "Your goal %q %s%s.", label, state, prog)
	if len(blockers) > 0 {
		fmt.Fprintf(&b, "\n  Blocked by: %s.", strings.Join(blockers, "; "))
	}
	if len(subgoals) > 0 {
		fmt.Fprintf(&b, "\n  To do this you need: %s.", strings.Join(subgoals, "; "))
	}
	if len(serves) > 0 {
		fmt.Fprintf(&b, "\n  Doing this serves: %s.", strings.Join(serves, "; "))
	}
	return b.String()
}

// renderUnknowns merges the goal-blocking open questions (core-blocking by
// construction) with the low-confidence beliefs into the anti-bluff unknowns
// block, core-blocking first, capped. Returns "" when there is nothing unknown.
func renderUnknowns(blockers, lowConf []string) string {
	items := make([]string, 0, epUnknownCap)
	for _, b := range blockers {
		items = append(items, "[BLOCKS YOUR GOAL] "+b)
		if len(items) >= epUnknownCap {
			break
		}
	}
	for _, l := range lowConf {
		if len(items) >= epUnknownCap {
			break
		}
		items = append(items, l)
	}
	if len(items) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("What you do NOT reliably know (do not act as if you do — verify, ask, or explore before committing):")
	for _, it := range items {
		b.WriteString("\n  • " + it)
	}
	return b.String()
}

// curiosityBias renders the decision-time explore<->exploit cue (Deliverable 5),
// priority-gated: a core-blocking unknown is a MANDATORY resolve regardless of
// the dials; otherwise a tangential learning detour is OPTIONAL and weighted by
// the persona's dominant curiosity flavor. The host emits a BIAS STRING, never a
// decision — the LLM decides explore-vs-exploit. Zero persona ⇒ "" (neutral).
func curiosityBias(cu persona.Curiosity, coreBlocked, hasTangent bool) string {
	if coreBlocked {
		return "An unknown is BLOCKING your goal. Resolving it (ask an NPC, read a sign, or go look) takes priority over pushing the blocked plan again."
	}
	if !hasTangent {
		return "" // nothing tangential to be curious about right now
	}
	flavor, mag := dominantCuriosity(cu)
	switch {
	case mag >= epCuriosityHi:
		return "You're a curious sort — drawn to " + flavor + ". A short detour to learn something new is worth it if nothing urgent is pressing."
	case mag < epCuriosityLo:
		return "" // low-curiosity persona: stay on task
	default:
		return "" // middling curiosity: no strong steer either way
	}
}

// dominantCuriosity returns the most-pronounced curiosity flavor (as the existing
// curiosityPhrase wording, for voice consistency with the persona card) and its
// magnitude. The flavor string is empty when the vector is all-zero.
func dominantCuriosity(cu persona.Curiosity) (string, float64) {
	flavors := []struct {
		v float64
		s string
	}{
		{cu.Spatial, "exploring new places and seeing what's over the next hill"},
		{cu.Skill, "learning and mastering new skills"},
		{cu.Economic, "deals, trade, and turning a profit"},
		{cu.Social, "meeting people and hearing their stories"},
		{cu.Risk, "testing yourself against danger"},
	}
	best := flavors[0]
	for _, f := range flavors[1:] {
		if f.v > best.v {
			best = f
		}
	}
	if best.v <= 0 {
		return "", 0
	}
	return best.s, best.v
}

// nodeLabel returns a node's human label, falling back to its ID.
func nodeLabel(n goalgraph.Node) string {
	if s := strings.TrimSpace(n.Label); s != "" {
		return s
	}
	return strings.TrimSpace(n.ID)
}

// orThing returns kind or a generic word when the subject's kind is unknown.
func orThing(kind string) string {
	if strings.TrimSpace(kind) == "" {
		return "thing"
	}
	return kind
}

// confidenceWord coarsens a Beta confidence into leak-free prose (no numbers).
func confidenceWord(c float64) string {
	switch {
	case c >= 0.85:
		return "you're confident"
	case c >= 0.65:
		return "you're fairly sure"
	default:
		return "you think so"
	}
}

// seenPhrase renders provenance + familiarity as plain prose ("saw it yourself,
// seen 3×").
func seenPhrase(provenance string, familiar int) string {
	var how string
	switch provenance {
	case knowledge.ProvSystem:
		how = "the game told you"
	case knowledge.ProvObserved:
		how = "you saw it yourself"
	case knowledge.ProvDeduced:
		how = "you worked it out"
	case knowledge.ProvHearsay:
		how = "someone told you"
	default:
		how = "you picked it up"
	}
	if familiar > 0 {
		return fmt.Sprintf("%s, seen %d×", how, familiar)
	}
	return how
}

// dedupLower returns the input with empty + case-insensitive-duplicate entries
// removed, preserving first-seen order (the original-cased value is kept).
func dedupLower(xs []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(xs))
	for _, x := range xs {
		t := strings.TrimSpace(x)
		if t == "" {
			continue
		}
		k := strings.ToLower(t)
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, t)
	}
	return out
}
