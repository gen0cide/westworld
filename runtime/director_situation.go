package runtime

import (
	"fmt"
	"math"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
	"github.com/gen0cide/westworld/worldmap"
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

	stalledNow := d.worldStall >= antiStuckWorldStall
	if !displaced && stuck {
		if stalledNow {
			// World-progress stall (C-3 stuck-breaker): NOTHING has changed for
			// antiStuckWorldStall turns. On crossing the threshold, BLOCK the
			// effective goal ONCE so the host-side ASK/FORAGE drives engage: a
			// no-progress "completes but the world never changes" loop (the
			// wrong-tool case) is treated like a failure — a stalled acquire-item
			// goal mints the where-to-buy question the forager travels for. The
			// write precedes the trigger assembly so the freshly-minted question
			// is part of the evidence the breaker renders this same turn. Uses
			// the PURE effectiveGoalView (effectiveGoal mutates + is called below
			// for sit.Goal — don't double-advance). Idempotent + M8-safe; gated
			// on !freeze; fires once at the crossing (worldStall increments by 1
			// per stalled turn).
			if !freeze && d.worldStall == antiStuckWorldStall {
				if g := d.effectiveGoalView(h); g != "" && h.goalGraph != nil {
					d.markGoalBlocked(h, g, "stalled — no progress for several turns; finding what is missing")
				}
			}
			trigger = d.stallBreakerTrigger(h)
		} else {
			trigger = fmt.Sprintf("STUCK — you have not moved in %d turns; your current approach is NOT working, change it", d.stuckTurns)
		}
	}
	// Anti-stuck v0: escalate on a failure streak. A hard streak means "abandon
	// the whole approach", not "retry a variation"; a soft streak just nudges.
	// Skipped entirely on a displaced turn: the durable BLOCKED graph write would
	// mis-mark the goal for a failure streak the jump invalidated (H2). When the
	// stall breaker took the trigger it KEEPS it — its evidence list already
	// names the recent failures with their errors, strictly more than the bare
	// BLOCKED text — but the fail-streak graph write below still lands.
	if !displaced && d.failStreak >= antiStuckHardFails {
		if !stalledNow {
			trigger = fmt.Sprintf("BLOCKED — your last %d actions all FAILED. This approach is not working; ABANDON it entirely and do something fundamentally different (a different place, NPC, or activity). Do NOT retry a variation of the same thing.", d.failStreak)
		}
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
	} else if !displaced && !stalledNow && d.failStreak >= antiStuckSoftFails {
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
	// (The old trailing `trigger == ""` STALLED block was DEAD — triggerFor never
	// returns "", and the STUCK branch above already owned the worldStall case —
	// so a stalled host only ever saw the bare position-based STUCK text. The
	// stall now fires inside that branch as the evidence-grounded stuck-breaker.)

	// C-2 confidence-scaled commitment: one per-turn confidence read off the
	// negative-evidence counters (consecutive failures, world-progress stall,
	// plan spin) offset by the goal's recorded progress (the positive term) —
	// computed AFTER the displacement resets so a fresh re-orient reads clean.
	// LOW changes the marching orders: ONE small reversible step, not a long
	// committed routine. The rider COMPOSES with whatever trigger won above
	// (the trigger names the problem; this modulates commitment).
	prog := d.goalProgress(h)
	conf := d.planConfidence(prog)
	if conf < lowConfidenceFloor {
		trigger += fmt.Sprintf(" [LOW CONFIDENCE (%.2f): recent turns went badly — do NOT commit to a long routine now. Take ONE small, reversible step (a short walk, one check, one question), watch the result, and reassess next turn.]", conf)
	}
	// Calibration honesty (decision-brain-tiering R2): persist the value next to
	// outcomes (decisions.jsonl via the bus) so confidence-vs-reality can be
	// validated offline — confidence@turn t is judged by the outcome that lands
	// in the NEXT record's last= field. One record per Act escalation; skipped
	// under freeze so an analysis dry-run never pollutes the stream (M16).
	if !freeze {
		h.publishDecision("confidence", "calibrate", fmt.Sprintf(
			"confidence=%.2f fails=%d stall=%d spin=%d progress=%.2f last=%s",
			conf, d.failStreak, d.worldStall, d.spinCount, prog, outcomeWord(last)))
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
	// Sell affordance (TODO C-cluster; retro cause #6): shop.sell has existed
	// for the project's whole life and has NEVER been invoked — a mogul planned
	// to "liquidate my leather armour for seed capital" and never did; another
	// host gave loot away FREE — because no prose ever told the planner that
	// shops BUY. The proven rule (docs/lessons-learned/5_LLM-FACING-DOCS.md):
	// prose presence predicts usage. So when the host is broke but holding items
	// a shop would buy, surface the affordance as ONE compact line. It rides the
	// scene hint because mesad's actPrompt renders a FIXED set of hint keys — a
	// brand-new key would never reach the model (the exact invisible-affordance
	// failure this line exists to fix). It composes with (never duplicates) the
	// toll-gate prose: reachable()/go_to report what a gate NEEDS and what you
	// HAVE; this line is the only one that says how to GET coins.
	if sell := d.sellAffordanceHint(h, pos.X, pos.Y); sell != "" {
		hints["scene"] += "\n\n💰 " + sell
	}
	// Self-aware ignorance (the fog-of-war ask): the host always KNOWS how
	// little of the world it has truly seen and which skills it has never
	// tried — facts about the self, like HP, so exploration and skill
	// discovery become choices instead of accidents. Rides the scene hint for
	// the same fixed-key reason as the sell line.
	if fog := d.explorationHint(h, pos.X, pos.Y); fog != "" {
		hints["scene"] += "\n\n🧭 " + fog
	}
	if sk := d.skillIgnoranceHint(h); sk != "" {
		hints["scene"] += "\n\n🛠 " + sk
	}
	// The aspiration ladder: what this turn's work is FOR. Compact portfolio
	// + the serves-context of the active goal.
	if asp := d.aspirationHint(h); asp != "" {
		hints["scene"] += "\n\n⭐ " + asp
	}
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

// --- Sell affordance (means-ends/economy gap, smallest slice) ----------------

// Sell-affordance dials. The hint must stay SCARCE to stay load-bearing: it
// fires only when the host is genuinely broke AND actually holds something a
// shop would buy.
const (
	sellHintCoinFloor = 30 // gp; at/above this the host isn't broke — no hint
	sellHintItemCap   = 3  // distinct items named; the rest fold into "and N more"
	sellHintShopNear  = 32 // tiles; a gazetteer general shop within this earns a "Nearest:" rider
)

// explorationHint renders the host's honest map-ignorance: weighted world
// coverage (sub-cell explored, not merely touched), how well it knows the
// ground it stands on, and where the unknown begins. Always rendered — fog is
// identity, not advice; the persona's curiosity decides whether to act on it.
func (d *MesaDirector) explorationHint(h *Host, x, y int) string {
	frac, _, total := h.Coverage()
	if total == 0 {
		return ""
	}
	terrain, contents := h.SectorUnderstanding(x, y)
	var b strings.Builder
	fmt.Fprintf(&b, "EXPLORATION — you have truly seen ~%.0f%% of the world.", frac*100)
	fmt.Fprintf(&b, " This area: %.0f%% walked", terrain*100)
	if contents > 0 || terrain > 0.3 {
		fmt.Fprintf(&b, ", %.0f%% of its places known to you", contents*100)
	}
	b.WriteString(".")
	if fronts := h.FrontierDirections(x, y); len(fronts) > 0 {
		parts := make([]string, 0, len(fronts))
		for _, f := range fronts {
			parts = append(parts, fmt.Sprintf("%s (~%d tiles)", f.Direction, f.Dist))
		}
		fmt.Fprintf(&b, " Unknown lands: %s.", strings.Join(parts, ", "))
	}
	b.WriteString(" What you have not seen, you do not know — go_to an unknown area to learn what is there.")
	return b.String()
}

// skillIgnoranceHint lists the skills the host has NEVER used (zero XP) — the
// other axis of self-aware ignorance. Stated as fact, never as instruction:
// whether an untried skill is worth trying is the persona's call.
func (d *MesaDirector) skillIgnoranceHint(h *Host) string {
	w := h.World()
	if w == nil || w.Self == nil {
		return ""
	}
	var never []string
	for id := 0; id <= int(event.SkillThieving); id++ {
		if w.Self.SkillXP(id) == 0 {
			never = append(never, event.SkillName(event.SkillID(id)))
		}
	}
	if len(never) == 0 {
		return ""
	}
	return "Skills you have never tried: " + strings.Join(never, ", ") +
		". Using one near its tool (a range, a net, an altar, a furnace...) earns experience on the first attempt."
}

// aspirationHint renders the goal portfolio: each aspiration with its rollup,
// the neglected ones marked. The active goal's serves-context rides the goal
// machinery (AspirationContext) via decision records; here the planner sees
// the LADDER so "what should I want next" has structure.
func (d *MesaDirector) aspirationHint(h *Host) string {
	port := h.AspirationPortfolio()
	if len(port) == 0 {
		return ""
	}
	parts := make([]string, 0, len(port))
	for _, a := range port {
		p := fmt.Sprintf("%q (%d done, %d underway)", a.Label, a.GoalsDone, a.GoalsActive)
		if a.Neglected {
			p += " — neglected lately"
		}
		parts = append(parts, p)
	}
	return "YOUR ASPIRATIONS — the long roads you chose: " + strings.Join(parts, "; ") +
		". Day-to-day goals should serve one of these; a neglected one may deserve your next goal."
}

// sellAffordanceHint renders the one-line "shops BUY items — shop.sell(item)"
// affordance when the host is broke (< sellHintCoinFloor gp) AND holds
// tradeable items with a positive catalogue value (BasePrice > 0, not
// IsUntradable, never the coins themselves). Items are grouped by kind and
// named most-valuable-first (lead with the armour, not an arrow), capped at
// sellHintItemCap with the rest folded into "and N more"; the gp figure is the
// base value of EVERYTHING sellable — a ballpark, not a shop quote. If the
// static gazetteer (embedded data already in RAM — a pure index read, NOT a
// perception verb: no flood, no in-world study cost) puts a general shop
// within sellHintShopNear tiles, a "Nearest:" rider lands the where. Returns
// "" otherwise — when rich or carrying nothing sellable, silence.
func (d *MesaDirector) sellAffordanceHint(h *Host, x, y int) string {
	if h.facts == nil || h.world == nil || h.world.Inventory == nil {
		return "" // can't judge sellability without the catalogue
	}
	if h.coinCount() >= sellHintCoinFloor {
		return "" // not broke — keep the hint scarce
	}
	coinID := -1
	if def := h.facts.ItemDefByName("Coins"); def != nil {
		coinID = def.ID
	}
	// Group sellable slots by item kind (unstackables occupy one slot each).
	type sellStack struct {
		name  string
		count int
		value int // BasePrice × count
	}
	var order []int // first-seen kind order; re-ranked by value below
	byID := map[int]*sellStack{}
	for _, sl := range h.world.Inventory.Slots() {
		if sl.ItemID == coinID || sl.Amount <= 0 {
			continue // coins are what we're short of, not stock to sell
		}
		def := h.facts.ItemDef(sl.ItemID)
		if def == nil || def.IsUntradable || def.BasePrice <= 0 || def.Name == "" {
			continue // not something a shop would buy (or unpriceable)
		}
		s := byID[sl.ItemID]
		if s == nil {
			s = &sellStack{name: def.Name}
			byID[sl.ItemID] = s
			order = append(order, sl.ItemID)
		}
		s.count += sl.Amount
		s.value += def.BasePrice * sl.Amount
	}
	if len(order) == 0 {
		return "" // nothing sellable — silence
	}
	sort.SliceStable(order, func(i, j int) bool { return byID[order[i]].value > byID[order[j]].value })
	total := 0
	named := make([]string, 0, sellHintItemCap)
	for i, id := range order {
		s := byID[id]
		total += s.value
		if i < sellHintItemCap {
			n := s.name
			if s.count > 1 {
				n = fmt.Sprintf("%d %s", s.count, n)
			}
			named = append(named, n)
		}
	}
	list := strings.Join(named, ", ")
	if more := len(order) - sellHintItemCap; more > 0 {
		list += fmt.Sprintf(" and %d more", more)
	}
	hint := fmt.Sprintf("You hold sellable items (%s — ~%dgp base value); general shops BUY items: open the shop and use shop.sell(item).", list, total)
	// Nearest general shop, from the embedded gazetteer the scene already reads.
	// Distance is Chebyshev to the map marker (hence the ~); rendered only when
	// actionably close — pointing at a shop half the world away is search_map's
	// job, not a hint's.
	if poi, dist, ok := h.facts.Gazetteer().NearestPOI("general-shop", x, y); ok && dist <= sellHintShopNear {
		hint += fmt.Sprintf(" Nearest: general shop ~%d tiles to your %s.", dist, bearingFrom(x, y, poi.X, poi.Y))
	}
	return hint
}

// --- C-2 confidence-scaled routine commitment ---------------------------------

// Commitment-confidence weights (C-2, TODO #33). The signal is DERIVED
// EVIDENCE, not the limbic Affect.Confidence dial: each negative-evidence
// counter the runtime already tracks subtracts from a 1.0 baseline —
// consecutive FAILED outcomes (strongest: the hard-fail BLOCKED threshold of 4
// lands at 0), world-progress stall turns (the progressKey no-change counter
// NoteStall mirrors; the STALLED threshold of 5 lands at 0.25), and plan-spin
// turns (mildest: spin one-shots at 3 ⇒ 0.55, above the floor — SPINNING has
// its own trigger). Whether these weights match reality is exactly what the
// per-turn "confidence" calibration record exists to find out
// (decision-brain-tiering R2).
const (
	confFailWeight  = 0.25
	confStallWeight = 0.15
	confSpinWeight  = 0.15
	// confProgressWeight is the one POSITIVE-evidence term (#33 names all
	// four signals): the effective goal's recorded graph Progress (0..1)
	// offsets transient negatives, so a grind that is demonstrably advancing
	// shrugs off a couple of failures (3 fails at full progress reads 0.5 —
	// not LOW) while a zero-progress goal keeps the bare negative read.
	confProgressWeight = 0.25
	// lowConfidenceFloor: BELOW this the Act trigger carries the one-small-
	// reversible-step marching orders instead of letting the planner author a
	// long committed routine. 0.5 sits just past the soft-fail nudge (2 fails
	// = exactly 0.5, still the nudge's territory) and ahead of the hard-fail/
	// stall overrides, so LOW engages in the in-between turns those miss.
	lowConfidenceFloor = 0.5
)

// planConfidence derives the per-turn commitment confidence in [0,1] from the
// counters situation() has already updated this turn plus the effective goal's
// recorded progress (goalProgress — the positive-evidence term). 1.0 = no
// negative evidence; 0 = everything says the current approach is failing. Pure —
// call after the displacement resets so a hard re-orient reads clean.
func (d *MesaDirector) planConfidence(progress float64) float64 {
	c := 1.0 - confFailWeight*float64(d.failStreak) -
		confStallWeight*float64(d.worldStall) -
		confSpinWeight*float64(d.spinCount) +
		confProgressWeight*math.Max(0, math.Min(1, progress))
	return math.Max(0, math.Min(1, c))
}

// goalProgress is the effective goal's recorded graph Progress (0..1), or 0
// when there is no graph/goal/node. Pure view read (effectiveGoalView — never
// advances the goal).
func (d *MesaDirector) goalProgress(h *Host) float64 {
	if h == nil || h.goalGraph == nil {
		return 0
	}
	g := d.effectiveGoalView(h)
	if g == "" {
		return 0
	}
	n, ok := h.goalGraph.Get(g)
	if !ok {
		return 0
	}
	return n.Progress
}

// outcomeWord compresses an Outcome into the one-word tag the confidence
// calibration record carries ("none" = no prior action).
func outcomeWord(last Outcome) string {
	switch {
	case last.Intent.Label == "":
		return "none"
	case last.BudgetExpired:
		return "budget-expired"
	case last.OK():
		return "ok"
	default:
		return last.Kind.String()
	}
}

// --- C-3 stuck-breaker: the STALLED trigger with distilled evidence -----------

// Stall-evidence bounds. The whole payload rides the TRIGGER STRING — mesad's
// actPrompt renders a FIXED hint-key set, so a brand-new hint key would never
// reach the model (the sell-affordance gotcha; see hints["scene"] above), while
// the trigger is rendered verbatim. Hard-capped so a stalled host costs
// hundreds of tokens, not thousands.
const (
	stallEvFailCap     = 3    // failed routines named (newest first)
	stallEvDecisionCap = 3    // recent Act reasonings named (newest first)
	stallEvQuestionCap = 3    // blocking open questions (with forage state)
	stallEvReachCap    = 2    // map reachability verdicts
	stallEvLineLen     = 180  // runes per evidence line
	stallEvBudget      = 1400 // bytes across all evidence lines (~350 tokens)
)

// stallBreakerTrigger assembles the STALLED trigger from HOST-precomputed
// conclusions the host already holds — no new RPCs, no LLM: the recent failed
// routines with their errors and the planner's own recent reasonings (the
// HybridDirector's evidence rings), the active goal's blocking open questions
// with their forage-state tags (places already tried / confirmed empty), and
// the map oracle's reachability verdict for the places those questions point
// at. The contract stays prompt-level: the text demands either a concretely
// different plan grounded in the evidence, or an explicit unachievable-here
// verdict expressed through the existing goal_op:"abandoned" path.
func (d *MesaDirector) stallBreakerTrigger(h *Host) string {
	var lines []string

	// Temporal evidence, newest first: what failed with which error, then what
	// the planner reasoned on recent turns — from the HybridDirector's rings
	// (nil for bare-MesaDirector wirings; degrade to graph/map evidence).
	if hd := hybridWrapper(h); hd != nil {
		for i, f := range hd.recentFailures() {
			if i >= stallEvFailCap {
				break
			}
			l := fmt.Sprintf("FAILED %s — %s", f.label, f.errMsg)
			if f.count > 1 {
				l += fmt.Sprintf(" (×%d)", f.count)
			}
			lines = append(lines, l)
		}
		for i, r := range hd.recentDecisions() {
			if i >= stallEvDecisionCap {
				break
			}
			lines = append(lines, "you reasoned: "+r)
		}
	}

	// The goal's blocking open questions + the forage breadcrumbs already on
	// their nodes, newest first — grounds "go look somewhere" proposals in what
	// was ALREADY looked at.
	blockers := d.stallBlockers(h)
	for _, q := range blockers {
		lines = append(lines, stallQuestionLine(h, q))
	}

	// Map verdicts: can the places those questions point at even be REACHED
	// from here? Same oracle + topic→POI mapping the forager walks by.
	lines = append(lines, d.stallReachLines(h, blockers)...)

	if len(lines) == 0 {
		// Nothing to distill (no wrapper, no graph, no map) — the original
		// plain STALLED contract still holds.
		return fmt.Sprintf("STALLED — %d turns and NOTHING has changed (same position, fatigue, inventory, and XP). Your recent approach is accomplishing nothing. Do something FUNDAMENTALLY different. If you don't actually know HOW to make progress (e.g. how to cure fatigue, where to buy or find something), SAY SO and note the specific unknown — do NOT repeat the same plan.", d.worldStall)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "STALLED — %d turns and NOTHING has changed (same position, fatigue, inventory, and XP). What you have been doing does not work HERE. THE EVIDENCE (newest first):", d.worldStall)
	used := 0
	for _, l := range lines {
		l = clipRunes(l, stallEvLineLen)
		if used+len(l) > stallEvBudget {
			break // hard cap — drop the oldest/lowest-priority remainder
		}
		used += len(l)
		b.WriteString("\n• " + l)
	}
	b.WriteString("\nDecide FROM THIS EVIDENCE — two honest moves only: (a) a CONCRETELY different plan (a different place, target, or method — NOT a variation of anything listed above), or (b) if the evidence says it cannot be done from here (unreachable, every source spent, the same failure repeating), declare it: set goal_op:\"abandoned\" and pursue something else.")
	return b.String()
}

// stallBlockers returns the LIVE nodes blocking the active goal, newest first,
// capped — the same read sliceBlockers renders labels from, kept as full nodes
// so the breaker can show each question's forage-state tags. Resolved
// (done/abandoned) blockers are dropped: they are no longer evidence. Pure
// view read (effectiveGoalView — never advances the goal).
func (d *MesaDirector) stallBlockers(h *Host) []goalgraph.Node {
	g := d.effectiveGoalView(h)
	if g == "" || h.goalGraph == nil {
		return nil
	}
	bs := h.goalGraph.Blockers(g)
	live := bs[:0]
	for _, b := range bs {
		if b.Status != goalgraph.StatusDone && b.Status != goalgraph.StatusAbandoned {
			live = append(live, b)
		}
	}
	sort.Slice(live, func(i, j int) bool { return live[i].At > live[j].At })
	if len(live) > stallEvQuestionCap {
		live = live[:stallEvQuestionCap]
	}
	return live
}

// stallQuestionLine renders one blocking question with the forage breadcrumbs
// the forage/ask drives already wrote onto its node.
func stallQuestionLine(h *Host, q goalgraph.Node) string {
	l := "blocking you: " + clipRunes(nodeLabel(q), 80)
	if h.goalGraph == nil {
		return l
	}
	var state []string
	if tried := forageTagTargets(h.goalGraph.TagsWithPrefix(q.ID, "source-tried:"), "source-tried:"); len(tried) > 0 {
		state = append(state, "already tried: "+strings.Join(tried, ", "))
	}
	if spent := forageTagTargets(h.goalGraph.TagsWithPrefix(q.ID, "source-spent:"), "source-spent:"); len(spent) > 0 {
		state = append(state, "confirmed nothing there: "+strings.Join(spent, ", "))
	}
	if h.goalGraph.HasTag(q.ID, "forage-exhausted") {
		state = append(state, "EVERY known source exhausted")
	}
	if h.goalGraph.HasTag(q.ID, "forage-inflight") {
		state = append(state, "a go-look trip is already underway")
	}
	if len(state) > 0 {
		l += " [" + strings.Join(state, "; ") + "]"
	}
	return l
}

// forageTagTargets strips the forage tag plumbing ("source-tried:place:<label>",
// "source-tried:npc:<name>") down to human target names, capped.
func forageTagTargets(tags []string, prefix string) []string {
	const maxTargets = 4
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		t = strings.TrimPrefix(t, prefix)
		t = strings.TrimPrefix(t, "place:")
		if n, ok := strings.CutPrefix(t, "npc:"); ok {
			t = n + " (asked)"
		}
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) > maxTargets {
		out = append(out[:maxTargets], fmt.Sprintf("+%d more", len(out)-maxTargets))
	}
	return out
}

// stallReachLines explains whether the places the blocking questions point at
// are even REACHABLE from where the host stands — the oracle verdict that turns
// "go somewhere else" into "the shop IS open, 40 tiles south" or "every route
// is blocked; not from here". Reuses the forager's topic→POI-type mapping and
// the survey's nearest-POI explain; at most stallEvReachCap floods, and only on
// an already-rare stalled escalation turn.
func (d *MesaDirector) stallReachLines(h *Host, blockers []goalgraph.Node) []string {
	if h.worldOracle == nil || h.world == nil || h.world.Self == nil {
		return nil
	}
	pos := h.world.Self.Position()
	hcap := h.hostCapability()
	var out []string
	seen := map[string]bool{}
	for _, q := range blockers {
		for _, typ := range forageTopicPOITypes(q.Label) {
			if seen[typ] {
				continue
			}
			seen[typ] = true
			best, ok := h.nearestExplainedPOI(pos.X, pos.Y, typ, hcap)
			if !ok {
				out = append(out, fmt.Sprintf("map: no %s exists anywhere on your map", typ))
			} else {
				out = append(out, stallReachLine(pos.X, pos.Y, typ, best))
			}
			if len(out) >= stallEvReachCap {
				return out
			}
		}
	}
	return out
}

// stallReachLine words one oracle verdict the way survey_map does, plus the
// conclusion the stalled host should draw from it.
func stallReachLine(x, y int, typ string, best searchMapHit) string {
	where := fmt.Sprintf("%d tiles %s", best.dist, dirOf(x, y, best.x, best.y))
	switch best.info.Reach {
	case worldmap.ReachOpen:
		return fmt.Sprintf("map: nearest %s (%s) IS reachable — %s, free to walk; if you have not actually gone there, that is a concrete move", typ, best.label, where)
	case worldmap.ReachGated:
		return fmt.Sprintf("map: nearest %s (%s) is GATED by %s — needs %s, you have %d (you CAN pay); %s", typ, best.label, gateShort(best.info.Gate), best.info.Needs, best.info.YouHave, where)
	default: // ReachBlocked
		if best.info.Gate == "" {
			return fmt.Sprintf("map: nearest %s (%s) is UNREACHABLE from where you stand — no walkable route; it cannot be done from HERE", typ, best.label)
		}
		return fmt.Sprintf("map: nearest %s (%s) is BLOCKED by %s — needs %s, you have %d (you canNOT pass); %s", typ, best.label, gateShort(best.info.Gate), best.info.Needs, best.info.YouHave, where)
	}
}

// hybridWrapper reaches the production HybridDirector through the conductor
// handle bound at bootstrap. The wrapper knows the MesaDirector (it feeds it
// NoteStall) but not vice versa, so the evidence rings it accumulates are read
// back via the conductor. nil for REPL/test hosts and bare-MesaDirector
// wirings — callers degrade gracefully.
func hybridWrapper(h *Host) *HybridDirector {
	c := h.conductorHandle()
	if c == nil {
		return nil
	}
	hd, _ := c.director.(*HybridDirector)
	return hd
}

// clipRunes truncates s to at most n runes, marking the cut with an ellipsis.
func clipRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
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
