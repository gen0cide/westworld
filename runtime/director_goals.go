package runtime

import (
	"fmt"
	"strings"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

// effectiveGoal is the goal that steers planning THIS turn: a live operator
// override (mesa GOAL_REVISION via SetLiveGoal) when present, else the
// construction-time genesis/persona goal — UNLESS that goal has CLOSED
// (done/abandoned), in which case it ADVANCES to the next goal (Phase-5a). A
// soft bias — never a pearl veto.
//
// Called many times per turn (seed, journal, hints, sit.Goal). It is idempotent:
// it advances at most once (then d.goal IS the new active goal); the director is
// single-threaded per turn (conductor drives Next), so the d.goal mutation here
// is safe. The operator LiveGoal override still WINS — it is the first statement
// and short-circuits before any advancement.
//
// In analysis/freeze mode (h.AnalysisActive()) it does NOT mutate (no goal-graph
// status write, no d.goal advance) — a dry-run must not corrupt the in-RAM graph
// the host resumes with (M16). It then behaves identically to effectiveGoalView.
func (d *MesaDirector) effectiveGoal(h *Host) string {
	return d.resolveGoal(h, !h.AnalysisActive())
}

// effectiveGoalView is the PURE read: it returns the same goal effectiveGoal
// would steer with this turn but NEVER mutates (no d.goal advance, no goal-graph
// status write). The reactive/speech tier uses it so it can slice the blockers
// off the SAME goal the director plans against (H8) without a side effect on the
// per-turn director state.
func (d *MesaDirector) effectiveGoalView(h *Host) string {
	return d.resolveGoal(h, false)
}

// resolveGoal is the shared core of effectiveGoal / effectiveGoalView. When
// mutate is true it advances d.goal (sticky) and promotes the chosen successor
// to Active; when false it computes the same answer with no writes. In both
// modes, a closed (done/abandoned) goal with NO successor resolves to "" so the
// caller (sit.Goal, the satisfaction net, act.go's GOAL line) treats it as "no
// active objective" rather than re-handing the planner a finished goal (L5/H1).
func (d *MesaDirector) resolveGoal(h *Host, mutate bool) string {
	if lg := h.LiveGoal(); lg != "" {
		return lg // operator override wins — advancement is skipped
	}
	// Guard d.goal across the director (advancing) and speech (reading) goroutines
	// (H8): a write lock when we may advance, a read lock for the pure preview.
	// selectNextGoal reads d.goal too and is only ever reached from here, so it
	// runs under this same lock (it never re-locks).
	if mutate {
		d.goalMu.Lock()
		defer d.goalMu.Unlock()
	} else {
		d.goalMu.RLock()
		defer d.goalMu.RUnlock()
	}
	if d.goal != "" && h.goalGraph != nil {
		if n, ok := h.goalGraph.Get(d.goal); ok && (n.Status == goalgraph.StatusDone || n.Status == goalgraph.StatusAbandoned) {
			next := d.selectNextGoal(h, mutate)
			if next == "" {
				return "" // closed goal, no successor — no active objective (L5)
			}
			if mutate {
				d.goal = next // advance; sticky for the rest of the session
			}
			return next
		}
	}
	return d.goal
}

// selectNextGoal picks the successor when the active goal has closed: a graph
// open_goal node (newest — planner-adopt or a future mesa goal-cron) first, then
// the persona north-star, else "" (stay put — idle purposefully, never loop).
// Selection is a priority pick over EXISTING nodes — memory, not a solver: no
// search, no candidate enumeration beyond one OpenGoals scan. It promotes the
// chosen open_goal to Active only when mutate is true (a freeze-mode/preview read
// must not flip status).
func (d *MesaDirector) selectNextGoal(h *Host, mutate bool) string {
	if h.goalGraph != nil {
		for _, n := range h.goalGraph.OpenGoals() { // 1) graph open_goal (planner adopt / future cron)
			if !strings.EqualFold(n.ID, d.goal) {
				if mutate {
					h.goalGraph.SetStatus(n.ID, goalgraph.StatusActive)
				}
				return n.ID
			}
		}
	}
	if h.northStar != "" && !strings.EqualFold(h.northStar, d.goal) { // 2) persona fallback
		return h.northStar
	}
	return "" // 3) stay put
}

// markGoalBlocked records the host modelling its own being-stuck on g: an OPEN
// QUESTION the goal is blocked_by (so the stuck goal is visible AND a mesa
// distillation cron has a node to answer), the goal flipped to StatusBlocked,
// and — once, only if no requires/enables child exists yet — a generic ENABLING
// sub-goal placeholder (Phase-2a means-ends; the LLM/cron names the real
// precondition, the host does NOT solve). Deterministic ids ⇒ repeated calls
// reinforce ONE node/edge set (Upsert/Link dedup on normalized ids) — no graph
// growth, no hook regression. Shared by the hard-fail path and the spin path.
// Caller guarantees g != "" and h.goalGraph != nil.
func (d *MesaDirector) markGoalBlocked(h *Host, g, reason string) {
	if h.goalGraph == nil || g == "" {
		return
	}
	// Never re-open / re-block a goal the lifecycle already CLOSED (H1): a
	// done/abandoned goal must stay terminal — flipping it Blocked here lets the
	// OK-turn un-block recovery flip it Active again, resurrecting a finished goal
	// into an infinite loop. The status-progress path is already Done-guarded
	// (applyGoalOp default branch); this is the matching guard for the blocked write.
	if n, ok := h.goalGraph.Get(g); ok &&
		(n.Status == goalgraph.StatusDone || n.Status == goalgraph.StatusAbandoned) {
		return
	}
	// Acquirable-item goal ("buy/get/acquire <item>"): the blocker is a
	// SUBJECT-BEARING where-to-buy question — the pivot the forage/ask/close paths
	// all key off — NOT the subjectless how-to-progress (which the forager can't map
	// to POI types, and which would tie on At with where-to-buy under a non-stable
	// sort). Suppressing how-to-progress for these goals removes that contention (5b).
	if item := acquireItemSubject(g); item != "" {
		d.markTopicalQuestion(h, g, "where-to-buy", item)
		h.goalGraph.SetStatus(g, goalgraph.StatusBlocked)
		d.mintEnablerIfNeeded(h, g)
		return
	}
	qid := "how-to-progress:" + g
	// Don't re-OPEN a question the reactive/cron closers already RESOLVED (M8):
	// Upsert with a non-empty status forces it, so pass "" (leave the existing
	// status untouched) when the question node already exists, and only set
	// StatusOpen when it is absent — a fresh blocker is genuinely open.
	if h.goalGraph.Has(qid) {
		h.goalGraph.Upsert(qid, goalgraph.KindOpenQuestion, reason, "")
	} else {
		h.goalGraph.Upsert(qid, goalgraph.KindOpenQuestion, reason, goalgraph.StatusOpen)
	}
	h.goalGraph.SetStatus(g, goalgraph.StatusBlocked) // the goal is stuck
	h.goalGraph.Link(g, qid, goalgraph.RelBlockedBy)  // goal --blocked_by--> question
	d.mintEnablerIfNeeded(h, g)
}

// mintEnablerIfNeeded adds the generic ENABLING sub-goal placeholder (Phase-2a
// means-ends) once, only if g has no requires/enables child yet. Extracted from
// markGoalBlocked so both the acquirable-item and how-to-progress branches reuse it.
func (d *MesaDirector) mintEnablerIfNeeded(h *Host, g string) {
	if len(h.goalGraph.Out(g, goalgraph.RelRequires)) == 0 &&
		len(h.goalGraph.Out(g, goalgraph.RelEnables)) == 0 {
		sid := "enabler:" + g
		h.goalGraph.Upsert(sid, goalgraph.KindSubgoal,
			fmt.Sprintf("Find and remove what is preventing %q (a missing item, skill, or safer place).", g),
			goalgraph.StatusOpen)
		h.goalGraph.Link(g, sid, goalgraph.RelRequires) // surfaces in the 1a sub-goal slice
		h.goalGraph.Link(sid, g, goalgraph.RelEnables)  // closes the reader/writer loop
	}
}

// markTopicalQuestion mints (or M8-safely re-touches) a SUBJECT-BEARING open_question
// that blocks goal — the pivot the forage/ask/close paths all key off. Unlike the
// subjectless how-to-progress:<g>, its clean Label's longest significant word IS the
// subject, so salientTopic(Label)==subject and picker/closer agree on the topic.
// Deterministic qid (kind:subject) dedups. kind ∈ {where-to-buy, where-is, confirm}.
// Caller guarantees goal != "", subject != "", h.goalGraph != nil, and that goal is
// a real graph node (Link auto-creates a missing endpoint, so a non-graph northStar
// string would leak a KindState node). Returns the qid.
func (d *MesaDirector) markTopicalQuestion(h *Host, goal, kind, subject string) string {
	subject = strings.ToLower(strings.TrimSpace(subject))
	qid := kind + ":" + subject
	label := topicalLabel(kind, subject)
	if h.goalGraph.Has(qid) {
		h.goalGraph.Upsert(qid, goalgraph.KindOpenQuestion, label, "") // M8: don't re-open
	} else {
		h.goalGraph.Upsert(qid, goalgraph.KindOpenQuestion, label, goalgraph.StatusOpen)
	}
	h.goalGraph.Link(goal, qid, goalgraph.RelBlockedBy)
	return qid
}

// topicalLabel renders a SHORT subject-bearing phrase whose longest non-stopword
// token is the subject, so salientTopic(label)==subject (e.g. "where to buy a
// pickaxe" → "pickaxe": where=stopword, to/buy/a are <4 chars). It must NOT carry a
// verbose reason, or salientTopic would pick the wrong word and desync picker/closer.
func topicalLabel(kind, subject string) string {
	switch kind {
	case "where-to-buy":
		return "where to buy a " + subject
	case "where-is":
		return "where is the " + subject
	default: // "confirm"
		return "is it true about " + subject
	}
}

// acquireItemSubject returns the item phrase an acquire-item goal names, or "" if it
// is not one. Conservative: requires an acquireVerbs verb and returns the goal text
// AFTER it, trimmed of a leading article. No parser/inflection — mirrors
// goalSatisfiedByInventory's space-padded substring discipline (the verbs already
// carry their own bounding spaces), so " have all the bronze bars " (which contains
// " have all ", not " have a ") returns "" and the M6 false-complete guard holds.
func acquireItemSubject(goal string) string {
	gl := " " + strings.ToLower(strings.TrimSpace(goal)) + " "
	for _, v := range acquireVerbs {
		if i := strings.Index(gl, v); i >= 0 {
			tail := strings.TrimSpace(gl[i+len(v):])
			tail = strings.TrimPrefix(tail, "a ")
			tail = strings.TrimPrefix(tail, "an ")
			tail = strings.TrimPrefix(tail, "the ")
			return strings.TrimSpace(tail)
		}
	}
	return ""
}

// markGoalDone closes a satisfied goal: StatusDone + progress 1, and resets the
// spin state (the goal is no longer being worked, so a stale fingerprint must
// not carry into the successor). The single convergence point for BOTH the
// planner-declared completion (Move.GoalOp == "done") and the deterministic
// inventory-satisfaction net. Has-guarded by the caller.
func (d *MesaDirector) markGoalDone(h *Host, g string) {
	if h.goalGraph == nil || g == "" {
		return
	}
	h.goalGraph.SetStatus(g, goalgraph.StatusDone)
	h.goalGraph.SetProgress(g, 1)
	h.goalGraph.Untag(g, "spinning") // a closed goal is no longer spinning (H5)
	d.spinCount, d.lastPlanFP = 0, 0
	h.publishDecision("lifecycle", "goal-done", "goal complete: "+g)
}

// foragingInFlight reports whether a forage trip is in flight for a question blocking
// the effective goal — the spin-suppression signal (5b). While true, the SPINNING
// fire is suppressed for THIS goal so a park-and-resume forage detour (which leaves
// the grind's last.Intent unchanged and would otherwise read as a spin) does not
// prematurely abandon the very goal foraging is unblocking. Reads the forage-inflight
// TAG (written by tryForage, cleared on harvest or TTL) — a pure lock-guarded read.
// Transient (a few seconds of flight): a genuine non-forage spin still fires between
// trips. Uses effectiveGoalView (the PURE read), NOT effectiveGoal — it runs inside
// the fire condition, BEFORE the block's own d.effectiveGoal(h); the mutating form
// twice could double-advance a just-closed goal.
func (d *MesaDirector) foragingInFlight(h *Host) bool {
	if h == nil || h.goalGraph == nil {
		return false
	}
	g := d.effectiveGoalView(h)
	if g == "" {
		return false
	}
	for _, b := range h.goalGraph.Blockers(g) {
		if b.Kind == goalgraph.KindOpenQuestion && h.goalGraph.HasTag(b.ID, "forage-inflight") {
			return true
		}
	}
	return false
}

// normalizeGoalOp maps the planner's free-text goal_op onto the recognised
// vocabulary {"", "done", "abandoned", "adopt"}. The Act prompt asks for exactly
// "done", but Sonnet routinely emits near-synonyms ("complete"/"completed"/
// "finished"/"satisfied" for done, "abandon"/"give up"/"drop" for abandoned,
// "new goal" for adopt). The pre-fix switch silently routed every unrecognised
// value into the still-active default branch, so a declared-but-misspelled
// completion NEVER closed the goal and the host re-planned a finished goal until
// the spin detector fired (M7). recognized reports whether the (raw) value mapped
// to a known op, so the caller can WARN on a truly unknown value.
func normalizeGoalOp(raw string) (op string, recognized bool) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "":
		return "", true // active/progress report — the default, not an error
	case "done", "complete", "completed", "finish", "finished", "satisfied", "satisfy", "achieved", "accomplished":
		return "done", true
	case "abandoned", "abandon", "give up", "giveup", "drop", "dropped", "quit", "cancel", "cancelled", "canceled":
		return "abandoned", true
	case "adopt", "adopted", "new", "new goal", "new-goal", "queue", "add":
		return "adopt", true
	default:
		return "", false // unrecognised — treat as a progress report, but warn
	}
}

// applyGoalOp interprets the planner's goal-lifecycle declaration on the Move
// (Phase-5a completion + progress). Deterministic, host-side; the planner does
// the REASONING (is it done? how far along?), the host does the WRITE:
//   - "done"      → markGoalDone (StatusDone + progress 1); effectiveGoal then
//     advances to the successor on the next call.
//   - "abandoned" → StatusAbandoned (advancement likewise picks up the successor).
//   - "adopt"     → QUEUE goal_text as an open_goal candidate only; it never
//     switches the active goal (selectNextGoal owns that policy).
//   - "" (active) → honor a progress report, monotone UP only (never regress, and
//     never on a done/abandoned node).
//
// The raw goal_op is NORMALISED first (synonyms folded; an unknown value warns and
// falls through to the active/progress branch) so a near-synonym completion still
// closes the goal (M7).
func (d *MesaDirector) applyGoalOp(h *Host, move *mesaclient.Move) {
	if h.goalGraph == nil || move == nil {
		return
	}
	g := d.effectiveGoal(h)
	op, recognized := normalizeGoalOp(move.GoalOp)
	if !recognized {
		d.log.Warn("act ✗ unrecognised goal_op; treating as a progress report", "goal_op", move.GoalOp, "goal", g)
	}
	switch op {
	case "done":
		if g != "" && h.goalGraph.Has(g) {
			d.markGoalDone(h, g)
			d.log.Info("goal completed (planner-declared)", "goal", g)
		}
	case "abandoned":
		if g != "" && h.goalGraph.Has(g) {
			h.goalGraph.SetStatus(g, goalgraph.StatusAbandoned)
			h.goalGraph.Untag(g, "spinning") // an abandoned goal is no longer spinning (H5)
			d.spinCount, d.lastPlanFP = 0, 0
			d.log.Info("goal abandoned (planner-declared)", "goal", g)
		}
	case "adopt":
		// Queue the next objective — UNLESS it IS the active goal, or a node with
		// that (normalized) label already EXISTS in the graph. The graph keys nodes
		// on the normalized id, so a verbatim re-adopt already dedups via Upsert; the
		// Has guard additionally skips re-adopting a label that is already present as
		// ANY node (e.g. the active goal node, a queued open_goal, or a done goal) so
		// a free-text re-adopt never spawns a confusing twin and the graph stays
		// bounded (selectNextGoal already refuses to re-select the active goal — this
		// keeps the node count down too, the consumer half of H10).
		if t := strings.TrimSpace(move.GoalText); t != "" && !strings.EqualFold(t, g) && !h.goalGraph.Has(t) {
			h.goalGraph.Upsert(t, goalgraph.KindOpenGoal, t, goalgraph.StatusOpen)
			d.log.Info("goal queued (planner-adopted)", "goal", t)
		}
	default: // active — honor a progress report, monotone up only
		if g != "" && move.GoalProgress > 0 && h.goalGraph.Has(g) {
			if n, ok := h.goalGraph.Get(g); ok &&
				n.Status != goalgraph.StatusDone && n.Status != goalgraph.StatusAbandoned &&
				move.GoalProgress > n.Progress {
				h.goalGraph.SetProgress(g, move.GoalProgress)
			}
		}
	}
}

// acquireVerbs is the conservative vocabulary that marks a goal as an
// acquire-an-item goal (the only class the deterministic satisfaction net
// covers). Intentionally small to avoid false completions; the planner's
// declaration covers every other goal shape. The "a"/"an" variants are
// WORD-ANCHORED (leading + trailing space, matched against the space-padded goal)
// so they fire on "have a pickaxe" but NOT on "have all the bronze bars" /
// "have any reason" / "find another spot" — the loose substrings used to false-
// complete those (M6).
var acquireVerbs = []string{"acquire", "get ", "obtain", "buy ", " own ", " have a ", " have an ", " find a ", " find an "}

// goalSatisfiedByInventory is the one cheap deterministic completion check
// (Phase-5a): an ACQUIRE-ITEM goal whose target item is already in the host's
// inventory is satisfied. It requires (a) an acquire verb in the goal text AND
// (b) some live inventory item name appearing as a substring of the goal — a
// conservative substring match, NO parser and NO search. Returns false for any
// non-acquire goal (the planner's declaration handles those). Reuses itemName
// over the same inventory slots the situation render already walks.
func (d *MesaDirector) goalSatisfiedByInventory(h *Host, g string, w *world.World) bool {
	if w == nil || w.Inventory == nil {
		return false
	}
	gl := " " + strings.ToLower(g) + " "
	isAcquire := false
	for _, v := range acquireVerbs {
		if strings.Contains(gl, v) {
			isAcquire = true
			break
		}
	}
	if !isAcquire {
		return false
	}
	for _, sl := range w.Inventory.Slots() {
		if sl.ItemID <= 0 {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(d.itemName(h, sl.ItemID)))
		// Skip the unresolved "item#NNN" placeholder — a numeric id is not a name
		// the goal text would ever contain, and matching on it would be noise.
		if name == "" || strings.HasPrefix(name, "item#") {
			continue
		}
		if strings.Contains(gl, name) {
			return true
		}
	}
	return false
}
