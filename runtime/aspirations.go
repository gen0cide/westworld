package runtime

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
)

// aspirations.go is the ASPIRATION LADDER — the layered goal portfolio ABOVE the
// day-scale lifecycle (director_goals.go). A host holds 2-4 month-horizon
// ASPIRATIONS: persona-derived ambitions that never "complete" in the day-scale
// sense ("master a craft", "earn a guild badge", "come to know the whole map"),
// kept as goal-graph roots (goalgraph.KindAspiration, StatusActive forever).
// Day-scale goals LINK to the aspiration they advance (goal --serves-->
// aspiration); progress ROLLS UP from the serving goals (count + recency,
// goalgraph.Rollup); and advancement consults the PORTFOLIO BALANCE so an
// aspiration untouched past aspirationNeglectWindow is surfaced preferentially
// (pickOpenGoal / mintAspirationGoal, called from selectNextGoal).
//
// Two producers, one idempotent seam:
//   - the genesis compile emits aspirations (mesad/genesis.go → GenesisResult.
//     Aspirations); runhost hands them to the director via SetAspirations;
//   - a host whose graph has NONE (the live fleet predates the feature) derives
//     2-3 MECHANICALLY from the persona it kept at bootstrap (north star +
//     curiosity flavors) with NO LLM call — deriveAspirations, the lazy
//     migration that lights the feature up at the next bounce.
//
// Both paths converge in ensureAspirations on the director turn path, AFTER the
// limbic goroutine's graph warm-start; deterministic node ids (aspirationID)
// make seeding idempotent across turns, bounces, and the warm-start race.

// maxAspirations bounds the portfolio — a handful of standing ambitions, not a
// wishlist (mirrors mesad's maxGenesisAspirations).
const maxAspirations = 4

// aspirationNeglectWindow is how long an aspiration may go UNTOUCHED (no serving
// goal worked or completed — see goalgraph.ServesRollup.LastTouched) before the
// advancement picker treats it as NEGLECTED and surfaces it preferentially. The
// cycle dial of the portfolio: small enough that attention rotates within a
// session, large enough that a working grind is never yanked off its aspiration.
const aspirationNeglectWindow = 15 * time.Minute

// SetAspirations hands the director the genesis-emitted aspiration labels (and
// which one the opening goal serves) for seeding on the first turn. Called once
// at login, before the conductor starts (same seam as SetKeywordLadder); the
// labels stay pending until ensureAspirations confirms the graph carries them.
func (d *MesaDirector) SetAspirations(labels []string, openingServes string) {
	d.pendingAspirations = labels
	d.pendingServes = openingServes
}

// aspirationID is the deterministic node id for an aspiration label — the
// idempotence key for seeding and migration: the same label (any case/spacing)
// always upserts the SAME node, so re-seeds across turns/bounces converge
// instead of duplicating.
func aspirationID(label string) string {
	var b strings.Builder
	b.WriteString("aspiration:")
	dash := false
	for _, r := range strings.ToLower(strings.TrimSpace(label)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
			continue
		}
		if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}

// ensureAspirations makes the aspiration portfolio EXIST. Runs on the director
// turn path (MesaDirector.Next + the HybridDirector replay path) so it executes
// after the limbic goroutine has warm-started the graph — never at construction,
// when the durable graph hasn't loaded yet. Cheap once latched (one bool).
//
// Order: an already-populated portfolio (this session's seed, or a prior life's,
// persisted under "goalgraph:") wins — genesis labels are only ADOPTED when the
// graph has none, so a re-login does not churn standing ambitions. Then pending
// genesis labels; then the mechanical persona derivation (lazy migration). A
// host with no persona and no genesis seeds nothing — the feature stays dark and
// behavior is exactly pre-ladder.
//
// The ready latch is set on the CONFIRMING read, not the seeding write: a cold
// boot's wholesale goalgraph.Import may still be in flight on the limbic
// goroutine and could replace the graph just seeded — the next turn's probe
// re-seeds onto the same deterministic ids, so the portfolio converges ("exactly
// once" in effect, not just in attempt). Frozen under analysis mode (M16).
func (d *MesaDirector) ensureAspirations(h *Host) {
	if d.aspirationsReady || h == nil || h.goalGraph == nil || h.AnalysisActive() {
		return
	}
	if len(h.goalGraph.Aspirations()) > 0 {
		d.aspirationsReady = true
		d.linkGoalToPortfolio(h, "") // make sure the active goal is in the portfolio
		return
	}
	labels, serves := d.pendingAspirations, d.pendingServes
	if len(labels) == 0 {
		// Lazy migration: mechanical, no LLM. No serves declaration either — the
		// keyword nearest-match is better evidence than a blanket default (only
		// genesis genuinely DECLARES which aspiration the opening goal advances).
		labels, serves = deriveAspirations(h), ""
	}
	seeded := make([]string, 0, maxAspirations)
	for _, l := range labels {
		l = strings.TrimSpace(l)
		if l == "" || len(seeded) >= maxAspirations {
			continue
		}
		id := aspirationID(l)
		if h.goalGraph.Has(id) {
			continue // case/spacing duplicate within the list
		}
		h.goalGraph.Upsert(id, goalgraph.KindAspiration, l, goalgraph.StatusActive)
		seeded = append(seeded, l)
	}
	if len(seeded) == 0 {
		d.aspirationsReady = true // nothing to derive (personaless host) — stay dark
		return
	}
	d.linkGoalToPortfolio(h, serves)
	h.publishDecision("aspiration", "seed",
		fmt.Sprintf("adopted %d standing aspirations: %s", len(seeded), strings.Join(seeded, "; ")))
}

// linkGoalToPortfolio links the CURRENT day-scale goal into the portfolio
// (goal --serves--> aspiration): the declared aspiration when it names a live
// one, else the keyword nearest-match, else — only when genesis DECLARED a serve
// that somehow didn't resolve (defense-in-depth; mesad validates opening_serves
// against the emitted list) — the first (highest-priority) aspiration, so a
// genesis opening goal always belongs to the ladder. It also makes the goal a
// REAL KindGoal root first: Link would otherwise auto-create it as a bare state
// node and the situation builder's Has-guard would then skip its own promotion.
func (d *MesaDirector) linkGoalToPortfolio(h *Host, declared string) {
	d.goalMu.RLock()
	g := d.goal
	d.goalMu.RUnlock()
	if g == "" {
		return
	}
	if !h.goalGraph.Has(g) {
		h.goalGraph.Upsert(g, goalgraph.KindGoal, g, goalgraph.StatusActive)
	}
	if d.linkGoalAspiration(h, g, declared) == "" && declared != "" {
		// Seed-time anchor: genesis/derivation declared a serve but neither the
		// declaration nor the keywords matched — anchor to the first aspiration
		// rather than leaving the opening goal outside the portfolio.
		if asps := h.goalGraph.Aspirations(); len(asps) > 0 {
			h.goalGraph.Link(g, asps[0].ID, goalgraph.RelServes)
		}
	}
}

// linkGoalAspiration links goal --serves--> the aspiration it advances and
// returns the aspiration's node id ("" when no link was made). Resolution order:
// the DECLARED aspiration when it names a live one (by label or id,
// case-insensitive — the planner's goal_serves / genesis opening_serves), else
// the keyword nearest-match (mechanical fallback). On NO signal it links nothing
// — a wrong link would corrupt the rollups, and an unlinked goal simply doesn't
// touch the portfolio (its aspiration then reads neglected, which is true).
// Idempotent: a goal already serving an aspiration keeps it.
func (d *MesaDirector) linkGoalAspiration(h *Host, goal, declared string) string {
	if h.goalGraph == nil || goal == "" {
		return ""
	}
	asps := h.goalGraph.Aspirations()
	if len(asps) == 0 {
		return ""
	}
	for _, e := range h.goalGraph.Out(goal, goalgraph.RelServes) {
		if n, ok := h.goalGraph.Get(e.To); ok && n.Kind == goalgraph.KindAspiration {
			return n.ID // already in the portfolio
		}
	}
	target := matchAspiration(asps, declared)
	if target == "" {
		target = nearestAspiration(asps, goal)
	}
	if target == "" {
		return ""
	}
	h.goalGraph.Link(goal, target, goalgraph.RelServes)
	return target
}

// matchAspiration resolves a declared aspiration name against the live
// portfolio: exact label, exact id, or the label's deterministic id — all
// case-insensitive. "" when declared is empty or names nothing live.
func matchAspiration(asps []goalgraph.Node, declared string) string {
	declared = strings.TrimSpace(declared)
	if declared == "" {
		return ""
	}
	did := aspirationID(declared)
	for _, a := range asps {
		if strings.EqualFold(a.Label, declared) || strings.EqualFold(a.ID, declared) || strings.EqualFold(a.ID, did) {
			return a.ID
		}
	}
	return ""
}

// nearestAspiration is the MECHANICAL serve-link fallback: score each aspiration
// by token overlap with the goal text (2 points per shared significant token)
// plus curiosity-flavor family agreement (1 point when the goal's activity
// vocabulary and the aspiration's land in the same family — "mine some ore" and
// "master a craft" share the skill family with zero literal tokens). Highest
// positive score wins; ties keep SEED order (the authored priority). "" when
// nothing scores — no signal, no link.
func nearestAspiration(asps []goalgraph.Node, text string) string {
	tokens := sigTokens(text)
	if len(tokens) == 0 {
		return ""
	}
	tf := flavorFamilies(tokens)
	bestID, bestScore := "", 0
	for _, a := range asps {
		at := sigTokens(a.Label + " " + a.ID)
		score := 0
		for t := range tokens {
			if at[t] {
				score += 2
			}
		}
		af := flavorFamilies(at)
		for f := range tf {
			if af[f] {
				score++
			}
		}
		if score > bestScore {
			bestID, bestScore = a.ID, score
		}
	}
	return bestID
}

// sigTokens lowercases s and keeps the significant words: ≥3 chars and not a
// stop word the templates share ("the", "your", "and", "toward"...), so overlap
// scoring measures CONTENT, not scaffolding.
func sigTokens(s string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'))
	}) {
		if len(w) >= 3 && !aspirationStopWords[w] {
			out[w] = true
		}
	}
	return out
}

var aspirationStopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "your": true,
	"you": true, "that": true, "this": true, "until": true, "toward": true,
	"into": true, "from": true, "some": true, "more": true, "make": true,
	"progress": true, "aspiration": true, "goal": true, "its": true,
}

// aspirationFlavorWords maps activity vocabulary onto the five curiosity-flavor
// families — the same axes the persona's Curiosity vector uses — so a goal and
// an aspiration can agree on WHAT KIND of ambition they share without sharing a
// literal word. Conservative, RSC-grounded vocabulary; extend as the world does.
var aspirationFlavorWords = map[string][]string{
	"skill": {"mine", "mining", "smith", "smithing", "craft", "crafting", "fish",
		"fishing", "cook", "cooking", "chop", "woodcutting", "smelt", "forge",
		"train", "level", "skill", "master", "ore", "bar", "badge"},
	"economic": {"buy", "sell", "trade", "trading", "coin", "coins", "money",
		"gold", "profit", "shop", "bank", "earn", "livelihood", "wealth", "pays"},
	"spatial": {"explore", "exploring", "map", "travel", "road", "roads",
		"place", "places", "journey", "land", "world", "hill", "beyond", "know"},
	"social": {"friend", "friends", "talk", "meet", "meeting", "people",
		"guild", "name", "reputation", "stories", "welcome", "known"},
	"risk": {"fight", "kill", "slay", "combat", "danger", "dangers", "battle",
		"duel", "monster", "dragon", "mettle", "prove"},
}

// flavorFamilies returns the flavor families a token set touches.
func flavorFamilies(tokens map[string]bool) map[string]bool {
	out := map[string]bool{}
	for fam, words := range aspirationFlavorWords {
		for _, w := range words {
			if tokens[w] {
				out[fam] = true
				break
			}
		}
	}
	return out
}

// deriveAspirations is the LAZY MIGRATION path: 2-3 aspirations derived
// MECHANICALLY (no LLM call) from the persona snapshot the host captured at
// bootstrap — the north-star statement verbatim (it IS a month-horizon ambition)
// plus templates for the pronounced curiosity flavors, strongest first. The live
// fleet, whose persisted graphs predate the ladder, lights up at its next bounce
// through this path. Deterministic: the same persona derives the same labels,
// and aspirationID makes any re-derive upsert the same nodes. A personaless host
// (REPL/test: zero curiosity, empty north star) derives NOTHING — the ladder
// stays dark rather than inventing ambitions.
func deriveAspirations(h *Host) (labels []string) {
	if ns := strings.TrimSpace(h.northStar); ns != "" {
		labels = append(labels, ns)
	}
	flavors := []struct {
		v     float64
		label string
	}{
		{h.curiosity.Skill, "Master a craft — train a skill until others seek you out for it"},
		{h.curiosity.Economic, "Secure a steady livelihood — coin in the bank and a trade that pays"},
		{h.curiosity.Spatial, "Come to know your corner of the world — its roads, places, and what lies beyond"},
		{h.curiosity.Social, "Be known and welcome — friends made and a good name kept"},
		{h.curiosity.Risk, "Prove your mettle — face real danger and come back with the story"},
	}
	sort.SliceStable(flavors, func(i, j int) bool { return flavors[i].v > flavors[j].v })
	for _, f := range flavors {
		if len(labels) >= 3 {
			break
		}
		if f.v > 0 {
			labels = append(labels, f.label)
		}
	}
	return labels
}

// --- portfolio-balanced advancement (called from selectNextGoal) -------------

// neglectedAspirations returns the aspirations whose rollup shows no work inside
// aspirationNeglectWindow, MOST neglected first (oldest LastTouched; ties keep
// seed order). Empty when the portfolio is dark or everything is being fed.
func neglectedAspirations(h *Host) []goalgraph.Node {
	if h.goalGraph == nil {
		return nil
	}
	cutoff := time.Now().Add(-aspirationNeglectWindow).Unix()
	type cand struct {
		n       goalgraph.Node
		touched int64
	}
	var cs []cand
	for _, a := range h.goalGraph.Aspirations() {
		if r := h.goalGraph.Rollup(a.ID); r.LastTouched <= cutoff {
			cs = append(cs, cand{a, r.LastTouched})
		}
	}
	sort.SliceStable(cs, func(i, j int) bool { return cs[i].touched < cs[j].touched })
	out := make([]goalgraph.Node, len(cs))
	for i, c := range cs {
		out[i] = c.n
	}
	return out
}

// pickOpenGoal picks the successor open_goal when the active goal closes,
// PORTFOLIO-BALANCED: a candidate serving a NEGLECTED aspiration wins (most
// neglected first); with no neglect signal the pre-ladder newest-first order
// stands unchanged. Promotes the pick to Active only when mutate is true.
// Runs under d.goalMu (via selectNextGoal ← resolveGoal), so reading d.goal
// directly here is safe; it never re-locks.
func (d *MesaDirector) pickOpenGoal(h *Host, mutate bool) string {
	var candidates []goalgraph.Node
	for _, n := range h.goalGraph.OpenGoals() { // newest-first
		if !strings.EqualFold(n.ID, d.goal) {
			candidates = append(candidates, n)
		}
	}
	if len(candidates) == 0 {
		return ""
	}
	pick := candidates[0] // pre-ladder default: newest open_goal
	if negl := neglectedAspirations(h); len(negl) > 0 {
		rank := make(map[string]int, len(negl))
		for i, a := range negl {
			rank[strings.ToLower(a.ID)] = i + 1 // 1 = most neglected
		}
		best := 0
		for _, c := range candidates {
			for _, e := range h.goalGraph.Out(c.ID, goalgraph.RelServes) {
				if r := rank[strings.ToLower(e.To)]; r > 0 && (best == 0 || r < best) {
					best, pick = r, c
				}
			}
		}
	}
	if mutate {
		h.goalGraph.SetStatus(pick.ID, goalgraph.StatusActive)
		if ac := h.AspirationContext(pick.ID); ac != "" {
			h.publishDecision("aspiration", "goal-advance",
				"advancing to queued goal: "+pick.ID+" ("+ac+")")
		}
	}
	return pick.ID
}

// mintAspirationGoal surfaces the most NEGLECTED aspiration as a deterministic
// day-scale goal when the open-goal queue is empty — the cycle that keeps
// attention rotating across the portfolio instead of collapsing onto the
// north-star fallback forever. Nothing is minted while every aspiration is being
// fed (inside the window): the pre-ladder fallback chain then runs unchanged. A
// minted goal that already CLOSED (done/abandoned) is skipped, never resurrected
// (H1); PruneTerminal eventually frees its id. Writes only when mutate is true;
// the pure preview returns the same id with no graph mutation. Runs under
// d.goalMu (via selectNextGoal).
func (d *MesaDirector) mintAspirationGoal(h *Host, mutate bool) string {
	for _, a := range neglectedAspirations(h) {
		gid := "Make progress toward your aspiration: " + nodeLabel(a)
		if strings.EqualFold(gid, d.goal) {
			continue
		}
		if n, ok := h.goalGraph.Get(gid); ok &&
			(n.Status == goalgraph.StatusDone || n.Status == goalgraph.StatusAbandoned) {
			continue // ran its course this life — don't resurrect a closed goal
		}
		if mutate {
			h.goalGraph.Upsert(gid, goalgraph.KindGoal, gid, goalgraph.StatusActive)
			h.goalGraph.Link(gid, a.ID, goalgraph.RelServes)
			h.publishDecision("aspiration", "cycle",
				fmt.Sprintf("nothing queued and %q has gone untended — turning to it", nodeLabel(a)))
		}
		return gid
	}
	return ""
}

// --- render-ready accessors (consumed by the situation builder) --------------

// AspirationStatus is one rung of the host's aspiration portfolio with its
// rolled-up progress — the render-ready view for the situation builder and the
// debug control plane (served verbatim by /aspirations, hence the json tags).
type AspirationStatus struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	GoalsDone   int       `json:"goals_done"`            // serving goals completed
	GoalsActive int       `json:"goals_active"`          // serving goals being pursued (active/blocked)
	GoalsOpen   int       `json:"goals_open"`            // serving goals queued, unstarted
	LastTouched time.Time `json:"last_touched,omitzero"` // zero when never touched
	Neglected   bool      `json:"neglected"`             // untouched longer than aspirationNeglectWindow
}

// AspirationPortfolio returns the host's aspirations in seed (priority) order
// with their rollups. Empty when the ladder is dark (no persona, not yet
// seeded). Pure read — safe from any goroutine.
func (h *Host) AspirationPortfolio() []AspirationStatus {
	if h == nil || h.goalGraph == nil {
		return nil
	}
	now := time.Now()
	var out []AspirationStatus
	for _, a := range h.goalGraph.Aspirations() {
		r := h.goalGraph.Rollup(a.ID)
		st := AspirationStatus{
			ID: a.ID, Label: nodeLabel(a),
			GoalsDone: r.Done, GoalsActive: r.Working, GoalsOpen: r.Open,
		}
		if r.LastTouched > 0 {
			st.LastTouched = time.Unix(r.LastTouched, 0)
			st.Neglected = now.Sub(st.LastTouched) >= aspirationNeglectWindow
		}
		out = append(out, st)
	}
	return out
}

// AspirationContext renders the serving aspiration of a day-scale goal for the
// situation builder: "serves: master a craft — 2 goals completed toward it".
// "" when the goal serves no aspiration (ladder dark / unlinked goal). Pure read.
func (h *Host) AspirationContext(goalID string) string {
	if h == nil || h.goalGraph == nil || goalID == "" {
		return ""
	}
	for _, e := range h.goalGraph.Out(goalID, goalgraph.RelServes) {
		n, ok := h.goalGraph.Get(e.To)
		if !ok || n.Kind != goalgraph.KindAspiration {
			continue
		}
		s := "serves: " + nodeLabel(n)
		switch r := h.goalGraph.Rollup(n.ID); {
		case r.Done == 1:
			s += " — 1 goal completed toward it"
		case r.Done > 1:
			s += fmt.Sprintf(" — %d goals completed toward it", r.Done)
		}
		return s
	}
	return ""
}
