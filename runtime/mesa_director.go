package runtime

import (
	"context"
	"fmt"
	"hash/fnv"
	"log/slog"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/persona"
	"github.com/gen0cide/westworld/world"
)

// MesaDirector is the autonomous agent loop: each turn it snapshots the host's
// game state into a mesaclient.Situation, asks mesa.Act "what do I do now?", and
// turns the returned Move (DSL) into a conductor Intent. mesa owns the planning
// (LLM); the host owns execution (interpreter + pearl gate). This is the seam
// the PROTOCOL.md "Act — the agent step" describes.
type MesaDirector struct {
	client mesaclient.Client
	hostID string
	// goalMu guards the per-session lifecycle goal (d.goal). The director
	// goroutine ADVANCES d.goal in resolveGoal(mutate=true) when the active goal
	// closes; the SPEECH goroutine READS it via effectiveGoalView → resolveGoal(
	// mutate=false) (effectiveGoalForSpeech, so the ASK tier slices blockers off
	// the SAME goal the director plans against). Those two goroutines touched
	// d.goal unguarded — a data race (H8 regression). An RWMutex keeps the read
	// path (the common case, many per turn) cheap while serialising the rare
	// advance write.
	goalMu sync.RWMutex
	goal   string
	log    *slog.Logger
	turn   int

	// sub + transcript give the planner MEMORY across turns: a rolling log of
	// narrative events (NPC speech, system messages, items/levels gained) drained
	// from the bus each turn. Without this, what an NPC said during converse() is
	// gone by the next turn and the host re-greets in a loop.
	sub        <-chan event.Event
	transcript []string

	// stuck detection: if she hasn't moved for several turns, her approach isn't
	// working — widen perception and tell her to explore instead of repeating.
	prevX, prevY int
	hasPrev      bool
	stuckTurns   int
	// failStreak counts consecutive FAILED outcomes (anti-stuck v0): a host can
	// loop while still moving, repeating an action that keeps failing. A
	// successful turn resets it, so a working grind is never flagged. See
	// situation(); thresholds are antiStuckSoftFails / antiStuckHardFails.
	failStreak int

	// Spin detector (Phase-5a): the failStreak path only catches FAILING loops; a
	// host can also SUCCEED every turn while re-deriving substantially the same
	// plan forever (the live pickaxe bug — micro-actions succeed, the goal never
	// closes). lastPlanFP is the content fingerprint of the last EXECUTED plan;
	// spinCount counts successive OK turns producing the SAME fingerprint. At
	// antiStuckSpinTurns it fires the BLOCKED graph-write + a SPINNING trigger
	// nudging a done/abandon declaration or a different tack, then re-arms.
	lastPlanFP uint64
	spinCount  int

	// worldStall is set by the HybridDirector's world-progress detector (NoteStall):
	// the number of consecutive turns the host's coarse world state (position,
	// fatigue, hp, inventory, total xp) has NOT changed. It catches a loop the
	// position-only stuckTurns and the plan-fingerprint spinCount both miss — a host
	// that keeps "succeeding" at a routine that accomplishes nothing (the live
	// 100%-fatigue-at-the-bank loop). >= antiStuckWorldStall surfaces the STUCK
	// hint + a STALLED trigger so the planner re-plans differently or names the
	// unknown blocking it. Set on the conductor goroutine; read in situation().
	worldStall int

	// lastPlayerMsg pins the most recent thing a real player said to her for a
	// few turns and surfaces it to the PLANNER (not just the chat reflex), so a
	// player's directions ("go through the north door") actually steer her.
	// lastPlayerName is who said it, so the planner can weigh the trust ledger
	// (a known friend's "follow me" lands differently than a stranger's).
	lastPlayerMsg  string
	lastPlayerName string
	playerMsgAge   int

	// visited is a BOUNDED recency set of the last visitedCap tiles she has stood
	// on — used to flag doors she has ALREADY gone through (a door adjacent to a
	// recently-visited tile leads BACK) so she stops backtracking through the same
	// door. A bounded ring (not the old unbounded lifetime map) keeps the per-turn
	// director path host-LIGHT: an open-ended wanderer otherwise accreted tens of
	// thousands of tiles for its whole life (M14). doorUsed only needs a small
	// recency window, so the cap costs nothing.
	visited *tileRing

	// keywordLadder is the session-genesis attention ladder: words/people that
	// should catch her attention in others' chat, each with a tier + reflex.
	// Surfaced to the Act planner so she orients to her name, friends, trade
	// words, and goal topics. Set once at login from the genesis compile.
	keywordLadder []mesaclient.KeywordRung
}

// SetKeywordLadder installs the session-genesis attention ladder (called once at
// login, before the conductor starts).
func (d *MesaDirector) SetKeywordLadder(ladder []mesaclient.KeywordRung) {
	d.keywordLadder = ladder
}

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

// idleIntentLabel marks an idle/no-op intent (d.idle). The spin detector must
// ignore these: an idle is the ABSENCE of a plan, not a re-derived one, so
// legitimate idling — including the director's OWN error/empty-verb idle
// fallbacks — must not accrue spinCount and fire SPINNING (H4).
const idleIntentLabel = "act:idle"

// planFingerprint is a content hash of an EXECUTED plan (the Intent the conductor
// just ran), used by the spin detector to tell "the same plan again" from "a new
// plan". FNV-1a over the Label, the sorted Args, and the whitespace-normalized
// Source — so cosmetic reformatting doesn't read as a different plan, but a real
// change does. Returns 0 for an empty OR idle intent (first turn / no-op / a
// legit wait) so those never false-trip the detector (fp==0 resets spinCount).
func planFingerprint(in Intent) uint64 {
	if (in.Label == "" && in.Source == "") || in.Label == idleIntentLabel {
		return 0
	}
	args := make([]string, len(in.Args))
	for i, a := range in.Args {
		if a != nil {
			args[i] = a.Display()
		}
	}
	sort.Strings(args)
	hh := fnv.New64a()
	hh.Write([]byte(in.Label))
	hh.Write([]byte{0})
	hh.Write([]byte(in.RoutinePath))
	hh.Write([]byte{0})
	for _, a := range args {
		hh.Write([]byte(a))
		hh.Write([]byte{0})
	}
	hh.Write([]byte(strings.Join(strings.Fields(in.Source), " ")))
	return hh.Sum64()
}

const transcriptCap = 80 // narrative lines retained; the last ~18 feed each turn

// Anti-stuck v0 failure-streak thresholds (consecutive failed outcomes).
const (
	antiStuckSoftFails = 2 // nudge: "reconsider the approach"
	antiStuckHardFails = 4 // override: "abandon this approach"
)

// antiStuckSpinTurns is the spin threshold: the same executed plan this many
// successive OK turns means the host is re-deriving one plan and getting
// nowhere (a SUCCEEDING loop the fail-streak path misses). 3 is the smallest N
// that distinguishes deliberate iteration from a stuck loop.
const antiStuckSpinTurns = 3

// antiStuckWorldStall is the world-progress stall threshold: this many consecutive
// turns with NO change to the coarse world state (position/fatigue/hp/inventory/xp)
// means the host is making zero progress regardless of what plan it runs — the
// loop the cheap-loop replay + plan-fingerprint spin both miss. Surfaces the STUCK
// hint + a STALLED trigger. Matches the HybridDirector's stallEscalateTurns.
const antiStuckWorldStall = 5

// NoteStall records the HybridDirector's world-progress stall count (consecutive
// no-change turns). Called each turn before the director plans, on the conductor
// goroutine. situation() turns a high count into a STUCK hint + STALLED trigger.
func (d *MesaDirector) NoteStall(turns int) { d.worldStall = turns }

// enablerProgressFloor nudges a freshly-unblocked goal off a frozen 0 so the
// inspector + planner read an un-block as movement, not a stall.
const enablerProgressFloor = 0.25

// NewMesaDirector builds a director that drives the host toward goal via mesa.Act.
func NewMesaDirector(client mesaclient.Client, hostID, goal string, log *slog.Logger) *MesaDirector {
	if log == nil {
		log = slog.Default()
	}
	return &MesaDirector{client: client, hostID: hostID, goal: goal, log: log}
}

// Next implements Director: build the situation, call Act, convert the Move.
// It logs the full thought stream — what Delores perceives, what she decides,
// and the exact DSL she authors — so a run is fully observable.
func (d *MesaDirector) Next(ctx context.Context, h *Host, last Outcome) (Intent, bool) {
	d.ensureSub(h)
	d.drainTranscript(h) // fold this turn's narrative events into the rolling memory
	sit := d.situation(h, last)

	// --- perception: what she sees this turn ---
	d.log.Info("act ① perceives",
		"trigger", sit.Trigger,
		"pos", fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		"hp", fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		"inv", orNone(sit.World.InvSummary),
		"npcs", orNone(sit.World.NearbyNpcs),
		"dialog_open", sit.Hints["dialog_options"] != "",
		"dialog_options", sit.Hints["dialog_options"],
		"recalled_episodes", h.journalLen(),
		"nearby_players", sit.Hints["nearby_players"],
	)
	for _, m := range sit.Recent {
		d.log.Info("act ① recent message", "text", m)
	}
	if last.Err != nil {
		d.log.Info("act ① last error", "err", last.Err.Msg)
	}

	move, err := d.client.Act(ctx, sit)
	if err != nil {
		// Keep the loop alive on a planner error — pause and retry next turn.
		d.log.Warn("act ✗ planner failed; idling", "err", err)
		return d.idle(3), true
	}

	// Phase-5a goal lifecycle: apply the planner's DECLARATION about the active
	// goal (done / abandoned / adopt / progress). This is the ONLY place both the
	// Move and d.effectiveGoal(h) are in hand. The host writes the lifecycle state
	// the planner NAMED — it does not judge satisfaction itself (memory, not a
	// solver). The judgment becomes a durable graph write, fixing the live bug
	// where "the goal is met" was reasoned but never recorded.
	d.applyGoalOp(h, move)

	// --- decision: her reasoning + the DSL she wrote ---
	d.log.Info("act ② decides", "kind", moveKindName(move.Kind), "reasoning", move.Reasoning)
	if move.DSLSource != "" {
		d.log.Info("act ② authored DSL:\n" + move.DSLSource)
	}
	if move.Verb != "" {
		d.log.Info("act ② direct action", "verb", move.Verb, "args", strings.Join(move.ActionArgs, ", "))
	}

	// Publish the full thought to the bus so the debug control plane streams it
	// live (/ws, the browser dashboard) and records it (/events, JSONL).
	d.turn++
	h.bus.Publish(event.AgentThought{
		Turn: d.turn, Trigger: sit.Trigger, Goal: sit.Goal,
		Pos:        fmt.Sprintf("(%d,%d)", sit.World.X, sit.World.Y),
		HP:         fmt.Sprintf("%d/%d", sit.World.HPCur, sit.World.HPMax),
		Perception: strings.Join(sit.Recent, " | "),
		Reasoning:  move.Reasoning, MoveKind: moveKindName(move.Kind), DSL: move.DSLSource,
	})
	return d.moveToIntent(move), true
}

func orNone(xs []string) string {
	if len(xs) == 0 {
		return "(none)"
	}
	return strings.Join(xs, ", ")
}

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
	if hadAction := last.Intent.Label != ""; hadAction {
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
		if last.Err != nil {
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

// memoryHint renders the salient durable episodes for the Situation — what she
// has done this life (across restarts), so she doesn't redo finished steps and
// the planner has continuity the transcript ring (which resets) cannot give.
func (d *MesaDirector) memoryHint(h *Host) string {
	eps := h.journal.Salient(8)
	if len(eps) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Things you remember doing (durable across this life — don't redo what's already done):")
	for _, e := range eps {
		b.WriteString("\n- " + e.Text)
	}
	return b.String()
}

// nearbyPlayersHint lists players within reach as optional social opportunities,
// each tagged with her standing relationship (trust grade, familiarity, notes).
// px,py is her current position. Empty when no one is around.
func (d *MesaDirector) nearbyPlayersHint(h *Host, px, py int) string {
	const radius = 16
	parts := make([]string, 0, 4)
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-px)+absInt(p.Y-py) > radius {
			continue
		}
		who := d.trustNote(h, p.Name)
		if who == "" {
			who = "a stranger so far"
		}
		parts = append(parts, fmt.Sprintf("%s (to your %s @ %d,%d — %s)",
			p.Name, bearingFrom(px, py, p.X, p.Y), p.X, p.Y, who))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// attentionHint renders the genesis keyword ladder for the Situation: the
// words/people she should notice in chat, grouped with their default reflex.
func (d *MesaDirector) attentionHint() string {
	if len(d.keywordLadder) == 0 {
		return ""
	}
	parts := make([]string, 0, len(d.keywordLadder))
	for _, r := range d.keywordLadder {
		if r.Keyword == "" {
			continue
		}
		s := "\"" + r.Keyword + "\""
		if r.Action != "" {
			s += " → " + r.Action
		} else if r.Tier != "" {
			s += " (" + r.Tier + ")"
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// trustNote summarizes what the host knows of a player from the trust ledger —
// grade, familiarity, and any tags — or "" for someone she's never met. This is
// the substrate for judging an offer by the RELATIONSHIP rather than a keyword.
func (d *MesaDirector) trustNote(h *Host, name string) string {
	if h.ledger == nil || name == "" || !h.ledger.Known(name) {
		return ""
	}
	rel := h.ledger.Rel(name)
	parts := make([]string, 0, 5)
	parts = append(parts, rel.Grade.String())
	if rel.Familiar > 0 {
		parts = append(parts, fmt.Sprintf("met %d×", rel.Familiar))
	}
	// Multi-axis colour, only when an axis is meaningfully off neutral (terse).
	if rel.Affinity >= 0.3 {
		parts = append(parts, "warm")
	} else if rel.Affinity <= -0.3 {
		parts = append(parts, "cold")
	}
	if rel.Grievance >= 0.3 {
		parts = append(parts, "you hold a grudge")
	}
	parts = append(parts, rel.Tags...)
	return strings.Join(parts, ", ")
}

// ensureSub lazily subscribes to the bus (once) so the director can build a
// rolling narrative transcript across turns.
func (d *MesaDirector) ensureSub(h *Host) {
	if d.sub == nil {
		d.sub = h.bus.Subscribe("*", 4096)
	}
}

// drainTranscript pulls all currently-buffered bus events and appends the
// narrative-salient ones (NPC speech, system messages, items/levels) to the
// transcript, de-duplicating adjacent repeats.
func (d *MesaDirector) drainTranscript(h *Host) {
	for {
		select {
		case ev, ok := <-d.sub:
			if !ok {
				d.sub = nil
				return
			}
			if pc, isChat := ev.(event.OtherPlayerChat); isChat {
				if name := d.playerName(h, pc.PlayerIndex); !strings.EqualFold(name, d.hostID) {
					d.lastPlayerMsg = name + ": " + strings.TrimSpace(stripColors(pc.MessageText))
					d.lastPlayerName = name
					d.playerMsgAge = 0
				}
			}
			line := d.narrativeLine(h, ev)
			if line == "" {
				continue
			}
			if n := len(d.transcript); n > 0 && d.transcript[n-1] == line {
				continue
			}
			d.transcript = append(d.transcript, line)
			if len(d.transcript) > transcriptCap {
				d.transcript = d.transcript[len(d.transcript)-transcriptCap:]
			}
		default:
			return
		}
	}
}

// narrativeLine formats a bus event as a transcript line, or "" to skip it.
func (d *MesaDirector) narrativeLine(h *Host, ev event.Event) string {
	switch e := ev.(type) {
	case event.SystemMessage:
		return strings.TrimSpace(stripColors(e.Message))
	case event.NpcDialogText:
		return "NPC: " + strings.TrimSpace(stripColors(e.Text))
	case event.NpcChat:
		return "NPC: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.OtherPlayerChat:
		name := d.playerName(h, e.PlayerIndex)
		if strings.EqualFold(name, d.hostID) {
			return "" // her OWN chat echoed back as local chat — never treat as someone talking to her
		}
		return name + " says to you: " + strings.TrimSpace(stripColors(e.MessageText))
	case event.ChatReceived:
		if e.Speaker != "" {
			return e.Speaker + ": " + strings.TrimSpace(stripColors(e.Message))
		}
	case event.ItemGained:
		return fmt.Sprintf("(received %s ×%d)", d.itemName(h, e.ItemID), e.Count)
	case event.LevelUp:
		return fmt.Sprintf("(leveled up — skill %d is now %d)", int(e.Skill), e.NewLevel)
	case event.PolicyVeto:
		return fmt.Sprintf("⛔ YOUR OWN NATURE blocked '%s': %s — this action will NEVER work for you; do something else.", e.Action, e.Reason)
	}
	return ""
}

// moveToIntent turns a mesa Move into a conductor Intent.
func (d *MesaDirector) moveToIntent(m *mesaclient.Move) Intent {
	switch m.Kind {
	case mesaclient.MoveWriteRoutine:
		name := m.RoutineName
		if name == "" {
			name = "act_step"
		}
		return Intent{Label: "act:" + name, Name: name, Source: m.DSLSource}
	case mesaclient.MoveRunRoutine:
		return Intent{Label: "act:" + m.RoutinePath, RoutinePath: m.RoutinePath, Args: toValues(m.Args)}
	case mesaclient.MoveDirectAction:
		// Guard: an empty/whitespace verb must NOT become `act_direct() { () }` —
		// that parses, "completes" in 0s doing nothing, and (pre-OneShot) used to be
		// cached + replayed forever. Treat a verbless direct action as a brief idle.
		if strings.TrimSpace(m.Verb) == "" {
			d.log.Warn("act ✗ direct action with empty verb — idling instead of authoring a no-op")
			return d.idle(2)
		}
		// Wrap a single verb in a one-shot routine so it runs through the gate.
		src := fmt.Sprintf("runtime \"1.0\"\nroutine act_direct() {\n    %s(%s)\n}", m.Verb, dslArgList(m.ActionArgs))
		return Intent{Label: "act:" + m.Verb, Name: "act_direct", Source: src, OneShot: true}
	default: // MoveIdle
		secs := m.IdleSeconds
		if secs <= 0 {
			secs = 2
		}
		return d.idle(secs)
	}
}

func (d *MesaDirector) idle(secs int) Intent {
	src := fmt.Sprintf("runtime \"1.0\"\nroutine act_idle() {\n    wait(%d)\n}", secs)
	return Intent{Label: idleIntentLabel, Name: "act_idle", Source: src, OneShot: true}
}

func (d *MesaDirector) npcName(h *Host, typeID int) string {
	if h.facts != nil {
		if def := h.facts.NpcDef(typeID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return ""
}

// describeArea gives the planner FULL area context: every nearby NPC, object,
// scenery, and door with NAME and COORDINATES (nearest-first), unfiltered — so
// she can see the cooking range, fishing spot, rock, etc. and act on them
// (walk_to or use items on them by coordinate). The old notable-only filter hid
// interactables like the range; this does not.
func (d *MesaDirector) describeArea(h *Host, radius int) string {
	pos := h.world.Self.Position()
	// Record where she is so doors she's passed can be flagged later (bounded ring).
	if d.visited == nil {
		d.visited = newTileRing(visitedCap)
	}
	d.visited.add(pos.X, pos.Y)

	// Reachability annotation (#31): mark a nearby NPC/player/scenery the host CANNOT
	// path-nav to (a different walkable component — behind a wall, across a gap, up a
	// dead-end ladder) so the planner stops targeting unreachable things. It still
	// SEES the thing exists (unlike scan_for, which drops it from the target list),
	// but knows it can't act on it from here. Doors/boundaries are NOT marked — they
	// are the exits BETWEEN components.
	hostComp, reachGate := int32(-1), false
	if h.worldOracle != nil {
		if c, _, _, ok := h.worldOracle.CompNear(pos.X, pos.Y); ok {
			hostComp, reachGate = c, true
		}
	}
	reachLabel := func(label string, x, y int) string {
		if !reachGate || label == "" {
			return label
		}
		if c, _, _, ok := h.worldOracle.CompNear(x, y); !ok || c != hostComp {
			return label + " [CANNOT REACH from here — behind a wall / different floor; find a door or ladder out]"
		}
		return label
	}

	type obj struct {
		label string
		dist  int
	}
	var objs []obj
	seen := map[string]bool{}
	add := func(label string, x, y int) {
		if label == "" {
			return
		}
		full := fmt.Sprintf("%s @ (%d,%d), to your %s", label, x, y, bearingFrom(pos.X, pos.Y, x, y))
		if seen[full] { // dedup by name+coords, so multiple doors all show
			return
		}
		seen[full] = true
		objs = append(objs, obj{label: full, dist: absInt(x-pos.X) + absInt(y-pos.Y)})
	}
	// NPCs (with coords, for talk_to / use targeting).
	for _, n := range h.world.Npcs.All() {
		if absInt(n.X-pos.X)+absInt(n.Y-pos.Y) > radius {
			continue
		}
		add(reachLabel(d.npcName(h, n.TypeID), n.X, n.Y), n.X, n.Y)
	}
	// Players (with coords + bearing) — so she knows which way a real player is
	// and can answer/follow "I'm to your east" correctly instead of guessing.
	for _, p := range h.world.Players.All() {
		if p.Name == "" || strings.EqualFold(p.Name, d.hostID) || (p.X == 0 && p.Y == 0) {
			continue
		}
		if absInt(p.X-pos.X)+absInt(p.Y-pos.Y) > radius {
			continue
		}
		add(reachLabel("player "+p.Name, p.X, p.Y), p.X, p.Y)
	}
	// Static scenery + boundaries from facts (UNFILTERED — includes the range).
	if h.facts != nil {
		for _, p := range h.facts.Near(pos.X, pos.Y, radius) {
			switch p.Kind {
			case "scenery":
				add(reachLabel(p.Name, p.X, p.Y), p.X, p.Y)
			case "boundary":
				label := p.Name
				if d.doorUsed(p.X, p.Y) {
					label += " (you've been through this one before)"
				}
				add(label, p.X, p.Y)
			}
		}
	}
	// Dynamic scenery (GameObjects: fires, etc.), names resolved via facts.
	if h.world.Scenery != nil {
		for _, s := range h.world.Scenery.Near(pos.X, pos.Y, radius) {
			name := ""
			if h.facts != nil {
				if def := h.facts.SceneryDef(s.ID); def != nil {
					name = def.Name
				}
			}
			add(name, s.X, s.Y)
		}
	}
	sort.Slice(objs, func(i, j int) bool { return objs[i].dist < objs[j].dist })
	if len(objs) > 28 {
		objs = objs[:28]
	}
	var b strings.Builder
	fmt.Fprintf(&b, "You are at (%d,%d). Nearby (name @ x,y — walk_to them, or use items on scenery via use(item, x=, y=)):\n", pos.X, pos.Y)
	for _, o := range objs {
		fmt.Fprintf(&b, "- %s\n", o.label)
	}
	return b.String()
}

// bearingFrom returns a compass direction from (px,py) to (x,y) in RSC
// coordinates (north = smaller y, EAST = smaller x — the x-axis increases west).
// So a player saying "the north door" maps to the door with the smaller y.
func bearingFrom(px, py, x, y int) string {
	dx := x - px // east is negative dx
	dy := y - py // north is negative dy
	if dx == 0 && dy == 0 {
		return "right here"
	}
	var ns, ew string
	if dy < 0 {
		ns = "north"
	} else if dy > 0 {
		ns = "south"
	}
	if dx < 0 {
		ew = "east"
	} else if dx > 0 {
		ew = "west"
	}
	switch {
	case ns != "" && ew != "" && absInt(dy) > 2*absInt(dx):
		return ns
	case ns != "" && ew != "" && absInt(dx) > 2*absInt(dy):
		return ew
	case ns != "" && ew != "":
		return ns + ew
	case ns != "":
		return ns
	default:
		return ew
	}
}

// doorUsed reports whether she has already stood on or beside a door tile (so
// it leads back the way she came). Used to flag doors as do-not-reuse.
func (d *MesaDirector) doorUsed(x, y int) bool {
	if d.visited == nil {
		return false
	}
	for _, off := range [][2]int{{0, 0}, {1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
		if d.visited.has(x+off[0], y+off[1]) {
			return true
		}
	}
	return false
}

// visitedCap bounds the recency window of tiles the director remembers standing
// on. doorUsed only checks the 5 tiles around a door, so a small window is ample;
// the cap stops the lifetime tile-history from growing without bound (M14).
const visitedCap = 64

// tileRing is a fixed-capacity, insertion-ordered set of tiles: add appends
// (evicting the oldest once full) and has tests membership. Deterministic and
// O(1)-amortized, mirroring the host-LIGHT bound the journal/goalgraph use.
type tileRing struct {
	cap   int
	order [][2]int        // ring of the last cap distinct tiles, oldest first
	set   map[[2]int]bool // membership mirror of order
}

func newTileRing(capacity int) *tileRing {
	if capacity <= 0 {
		capacity = visitedCap
	}
	return &tileRing{cap: capacity, set: make(map[[2]int]bool, capacity)}
}

func (r *tileRing) add(x, y int) {
	k := [2]int{x, y}
	if r.set[k] {
		return // already recent — keep its position, don't re-age the window
	}
	if len(r.order) >= r.cap {
		oldest := r.order[0]
		r.order = r.order[1:]
		delete(r.set, oldest)
	}
	r.order = append(r.order, k)
	r.set[k] = true
}

func (r *tileRing) has(x, y int) bool { return r.set[[2]int{x, y}] }

func (d *MesaDirector) playerName(h *Host, idx int) string {
	for _, p := range h.world.Players.All() {
		if p.Index == idx && p.Name != "" {
			return p.Name
		}
	}
	return "a player"
}

func (d *MesaDirector) itemName(h *Host, itemID int) string {
	if h.facts != nil {
		if def := h.facts.ItemDef(itemID); def != nil && def.Name != "" {
			return def.Name
		}
	}
	return fmt.Sprintf("item#%d", itemID)
}

// --- helpers ----------------------------------------------------------------

func triggerFor(last Outcome) string {
	switch {
	case last.Intent.Label == "": // zero outcome
		return "start"
	case last.Err != nil:
		return "recover"
	default:
		return "continue"
	}
}

func lastN(xs []string, n int) []string {
	if len(xs) <= n {
		return xs
	}
	return xs[len(xs)-n:]
}

func toValues(args []string) []interp.Value {
	if len(args) == 0 {
		return nil
	}
	vs := make([]interp.Value, 0, len(args))
	for _, a := range args {
		vs = append(vs, interp.String(a))
	}
	return vs
}

// dslArgList renders direct-action args as DSL literals: numbers bare, else quoted.
func dslArgList(args []string) string {
	parts := make([]string, 0, len(args))
	for _, a := range args {
		if isNumeric(a) {
			parts = append(parts, a)
		} else {
			parts = append(parts, "\""+strings.ReplaceAll(a, "\"", "\\\"")+"\"")
		}
	}
	return strings.Join(parts, ", ")
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if r >= '0' && r <= '9' {
			continue
		}
		if (r == '-' || r == '+') && i == 0 {
			continue
		}
		if r == '.' {
			continue
		}
		return false
	}
	return true
}

// stripColors removes RSC "@xxx@" colour codes (5 chars: @ + 3 + @) and turns
// "%" line breaks into spaces, so message text reads cleanly to the model.
func stripColors(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '@' && i+4 < len(s) && s[i+4] == '@' {
			i += 5
			continue
		}
		if s[i] == '%' {
			b.WriteByte(' ')
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

func moveKindName(k mesaclient.MoveKind) string { return k.String() }
