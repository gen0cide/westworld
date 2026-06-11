package runtime

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// speech.go is the host's PHASE-3a intent-driven speech layer (docs/world-
// knowledge-and-learning.md §3.8a): speech is GOAL-SERVING, not just a reflex. The
// host TALKS in order to LEARN (the ASK drive) and to HELP (grounded answer +
// volunteer teach), closing the inquisitive loop the design centers on.
//
// The split is the project invariant: the host is the DETERMINISTIC GATE (have a
// goal-blocking open question? a relevant interlocutor in range? off cooldown?) —
// all O(open-Qs × nearby-entities), no LLM — and mesa supplies the WORDS (the
// cheap Haiku Chat/Ask path). The gate defaults to SILENCE. Frozen under analysis
// mode like all the other learning I/O.
//
// ASK is the proactive twin of the reactive reply: it fires from socialReflex on
// the AgentThought tick (a host-owned proactive clock — no ticker), composes via
// mesa.Ask, latches the target, and speaks via host.Say (which fans the line into
// the reactive windows so the forthcoming answer pairs with the question). The
// answer returns through the speed-2 reactive path → ExtractDialog → ledger →
// closeResolvedQuestions, which flips the matching open question StatusDone.

const (
	// askGlobalGap is the global ask floor (anti machine-gun): the host asks at
	// most ~once per this interval across ALL questions/targets. Above the reactive
	// latch TTL (20s) so a Q&A completes/decays before the next ask.
	askGlobalGap = 45 * time.Second
	// askQuestionCooldown: don't re-ask the SAME question of ANYONE within this —
	// give the reactive writeback + crons time to resolve it.
	askQuestionCooldown = 5 * time.Minute
	// askSameQSameTargetCooldown: don't re-ask the same question of the SAME target
	// (they already answered or refused).
	askSameQSameTargetCooldown = 15 * time.Minute
	// pesterCooldown: don't fire two DIFFERENT questions at one target back-to-back;
	// spread asks across interlocutors.
	pesterCooldown = 90 * time.Second
	// maxAskAttempts is the hard stop per question (tag ask-exhausted, skip forever
	// until a cron re-arms): a permanently-unanswerable how-to-progress: must not
	// become a forever drip.
	maxAskAttempts = 3
	// askRadius is the Chebyshev range an interlocutor must be within to ask (local-
	// chat / talk_to range).
	askRadius = 5
	// teachCooldown: the host volunteers a high-confidence belief to a given player
	// at most this often — it volunteers, it doesn't lecture.
	teachCooldown = 5 * time.Minute
	// The gate's "do I actually know this?" threshold is the director's existing
	// epConfFloor (0.5, director_situation.go) — one floor across gate + grounding.
	//
	// teachConfFloor is the high bar a belief must clear to be VOLUNTEERED (teach):
	// only confidently-held knowledge propagates, so hearsay doesn't spread as fact.
	teachConfFloor = 0.75
	// speechGCMax is the longest cooldown window; gcSpeech drops entries older.
	speechGCMax = askSameQSameTargetCooldown
)

// speechGate holds the intent-driven speech anti-spam state. All maps are RAM-
// only, mutex-guarded, and bounded by the (deduped) open-question / nearby-entity
// counts. Kept off reactiveState so the reactive mutex stays uncontended.
type speechGate struct {
	mu        sync.Mutex
	lastAsk   time.Time            // global ask floor (anti machine-gun)
	askedQ    map[string]time.Time // qid -> last asked of ANYONE
	askedQTo  map[string]time.Time // "qid|target" -> last asked of THIS target
	pestered  map[string]time.Time // "target" -> last asked ANYTHING of this target
	attempts  map[string]int       // qid -> unanswered-ask count (hard stop)
	lastTeach map[string]time.Time // "target" -> last volunteered a teach
}

// newSpeechGate builds an empty speech gate.
func newSpeechGate() *speechGate {
	return &speechGate{
		askedQ:    map[string]time.Time{},
		askedQTo:  map[string]time.Time{},
		pestered:  map[string]time.Time{},
		attempts:  map[string]int{},
		lastTeach: map[string]time.Time{},
	}
}

// gcSpeech drops cooldown entries older than the longest window so the maps stay
// bounded across a long session (a resolved/abandoned question's entries never
// need re-checking). attempts is kept (it's the permanent ask-exhausted memory)
// but pruned for questions that have aged out of every time-window. Mutex-guarded.
func (g *speechGate) gcSpeech(now time.Time) {
	g.mu.Lock()
	defer g.mu.Unlock()
	stale := func(t time.Time) bool { return now.Sub(t) > speechGCMax }
	for k, t := range g.askedQ {
		if stale(t) {
			delete(g.askedQ, k)
			delete(g.attempts, k) // its time-window lapsed; re-arm a non-exhausted question
		}
	}
	for k, t := range g.askedQTo {
		if stale(t) {
			delete(g.askedQTo, k)
		}
	}
	for k, t := range g.pestered {
		if stale(t) {
			delete(g.pestered, k)
		}
	}
	for k, t := range g.lastTeach {
		if stale(t) {
			delete(g.lastTeach, k)
		}
	}
}

// teachable reports whether the host may volunteer a teach to `target` now (off
// the per-target teach cooldown), and records the attempt when it returns true.
// Mutex-guarded.
func (g *speechGate) teachable(target string, now time.Time) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if t, ok := g.lastTeach[target]; ok && now.Sub(t) < teachCooldown {
		return false
	}
	g.lastTeach[target] = now
	return true
}

// askTarget is one candidate interlocutor the gate may direct a question at.
type askTarget struct {
	name  string  // display name (woven into the question by the composer)
	role  string  // "npc" | "player" (provenance + reactive routing)
	score float64 // relevance: topical overlap + familiarity (higher = better)
}

// tryAsk is the deterministic ASK gate — the proactive half of intent-driven
// speech. Called from socialReflex on the AgentThought tick. ALL predicates are
// pure (no LLM); only a full pass composes one question (mesa.Ask, Haiku) and
// speaks it. The gate defaults to SILENCE: most ticks return early. Frozen under
// analysis mode.
func tryAsk(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, username string) {
	// 1. Not frozen — the operator override freezes all learning I/O, ASK included.
	if host == nil || host.AnalysisActive() {
		return
	}
	if host.speech == nil || host.goalGraph == nil {
		return
	}
	if mc == nil || !mc.Healthy() {
		return // offline → nothing to compose; stay silent
	}
	now := time.Now()

	// 2. Global ask floor (anti machine-gun).
	host.speech.mu.Lock()
	if now.Sub(host.speech.lastAsk) < askGlobalGap {
		host.speech.mu.Unlock()
		return
	}
	host.speech.mu.Unlock()

	// 3. Pick exactly ONE goal-blocking open question (prefer ones blocking the live
	//    goal; fall back to the most-recent open question — the slice is newest-first).
	q, ok := host.pickAskQuestion(now)
	if !ok {
		return
	}

	// 6. Pick a relevant interlocutor in range, off the per-target cooldowns.
	tgt, ok := host.pickInterlocutor(q, now)
	if !ok {
		return // no eligible target → silent
	}

	// 7. Compose + emit. Ground the question with the target's window (+ any
	//    low-confidence suspicion, so we ask to CONFIRM rather than from zero).
	recent := host.reactive.snapshot(normalizeSpeaker(tgt.name))
	if susp := host.askSuspicion(q.Label); susp != "" {
		recent = append(recent, susp)
	}
	// Burn the global ask floor NOW — we are committed to a compose attempt. This
	// caps compose RPCs at one per askGlobalGap REGARDLESS of outcome, so a flaky
	// composer (speak=false) or a failed send can't tight-loop the (Haiku) RPC on
	// every AgentThought tick. The per-QUESTION cooldown is burned only on a real
	// send (below), so a failed attempt leaves the specific question re-askable.
	host.speech.mu.Lock()
	host.speech.lastAsk = now
	host.speech.mu.Unlock()

	text, speak, err := mc.Ask(ctx, username, tgt.name, q.Label, recent)
	if err != nil || !speak || strings.TrimSpace(text) == "" {
		return // composer chose silence / errored — global gap burned, question still askable
	}

	// Latch + ensure the target window exists BEFORE emitting (H6). The host's own
	// question must be captured into the target's window so the forthcoming answer
	// pairs with it; if we latched only AFTER the send, the self-line seam would
	// early-return on a zero latch count and the question would be dropped, so the
	// Q→A pairing could never happen. tryLatch now creates the window for a cold
	// target (e.g. a shop NPC never heard before).
	tgtKey := normalizeSpeaker(tgt.name)
	host.reactive.tryLatch(tgtKey, now)
	// Direct the self-line seam at THIS target only (L7): the emit below runs
	// host.Say, whose reactiveObserveSelf would otherwise broadcast the question
	// into every latched conversation. directSelfTo routes the next self-line into
	// the target's window alone and is consumed once by the emit.
	host.reactive.directSelfTo(tgtKey)

	emit := host.Say
	if host.emitSay != nil {
		emit = host.emitSay // test seam (capture the line without a live socket)
	}
	if err := emit(ctx, text); err != nil {
		log.Warn("speech: ask failed to send", "to", tgt.name, "asked", text, "err", err)
		host.reactive.clearDirectSelf() // drop the unconsumed routing hint
		host.reactive.unlatch(tgtKey)   // undo the pre-emit latch so a failed send leaves none orphaned
		return                          // global gap burned; no per-question cooldown → re-askable
	}

	// Record the per-question / per-target cooldowns + inspector-visible evidence
	// (the global floor was already burned above). Recheck the question is still
	// open: the reactive writeback goroutine (closeResolvedQuestions) may have
	// resolved it concurrently — still record that we pestered the target (we DID
	// send), but don't burn an ATTEMPT against an already-answered question.
	stillOpen := false
	if n, ok := host.goalGraph.Get(q.ID); ok && n.Status == goalgraph.StatusOpen {
		stillOpen = true
	}
	host.speech.mu.Lock()
	host.speech.askedQ[q.ID] = now
	host.speech.askedQTo[q.ID+"|"+tgt.name] = now
	host.speech.pestered[tgt.name] = now
	exhausted := false
	if stillOpen {
		host.speech.attempts[q.ID]++
		exhausted = host.speech.attempts[q.ID] >= maxAskAttempts
	}
	host.speech.mu.Unlock()
	host.goalGraph.Tag(q.ID, "asked:"+tgt.name)
	if tgt.role == "npc" {
		// 5b exhaustion: record that THIS NPC was tried for THIS question, so
		// pickInterlocutor rotates to a new source instead of re-asking a canned
		// dialogue tree. Players are NOT source-tried (independent hearsay sources).
		host.goalGraph.Tag(q.ID, "source-tried:npc:"+tgt.name)
	}
	if exhausted {
		host.goalGraph.Tag(q.ID, "ask-exhausted")
	}

	log.Info("speech: asked", "to", tgt.name, "role", tgt.role, "question", q.Label, "said", text)
	host.bus.Publish(event.AgentThought{
		Trigger:    "ask",
		Goal:       host.LiveGoal(),
		Perception: "asked " + tgt.name + ": " + q.Label,
		Reasoning:  text,
	})
}

// pickAskQuestion selects exactly ONE open question worth asking, applying the
// per-question predicates (not already known, off cooldown, under the attempt
// cap). Prefers questions explicitly blocking the live goal; falls back to the
// most-recent open question. Returns ok=false when none is eligible. Pure +
// deterministic (reads the speech gate under its lock).
func (h *Host) pickAskQuestion(now time.Time) (goalgraph.Node, bool) {
	qs := h.goalGraph.OpenQuestions() // newest-first
	if len(qs) == 0 {
		return goalgraph.Node{}, false
	}
	// Build the set of questions explicitly blocking the EFFECTIVE goal, to prefer
	// them. LiveGoal() is the operator override and is "" in the normal autonomous
	// case, so slicing blockers off it left Pass-1 dead — use the same goal the
	// director plans against (H8).
	blocking := map[string]bool{}
	for _, b := range h.goalGraph.Blockers(h.effectiveGoalForSpeech()) {
		if b.Kind == goalgraph.KindOpenQuestion {
			blocking[strings.ToLower(strings.TrimSpace(b.ID))] = true
		}
	}
	h.speech.mu.Lock()
	defer h.speech.mu.Unlock()
	eligible := func(q goalgraph.Node) bool {
		// Per-question cooldown + hard attempt cap.
		if t, ok := h.speech.askedQ[q.ID]; ok && now.Sub(t) < askQuestionCooldown {
			return false
		}
		if h.speech.attempts[q.ID] >= maxAskAttempts {
			return false // ask-exhausted — skip until a cron re-arms it
		}
		// Genuinely unknown: skip a question whose salient topic we already know.
		// how-to-progress: has no clean subject — rely on cooldown + the attempt cap.
		if !isHowToProgress(q.Label) {
			if topic := salientTopic(q.Label); topic != "" {
				if h.knowledge != nil && h.knowledge.Get(topic).Confidence >= epConfFloor {
					return false
				}
			}
		}
		return true
	}
	// Pass 1: a goal-blocking question (newest-first). Pass 2: any open question.
	for _, q := range qs {
		if blocking[strings.ToLower(strings.TrimSpace(q.ID))] && eligible(q) {
			return q, true
		}
	}
	for _, q := range qs {
		if eligible(q) {
			return q, true
		}
	}
	// Pass 3 (#23 broadened ASK): seed an ASK from a goal-relevant LOW-confidence
	// belief that has no open question yet — confirm a shaky belief instead of
	// sitting on it. Mints confirm:<subject> via the director so picker/closer share
	// the pivot; shares askGlobalGap (no new clock). Guarded so Link never
	// auto-creates a KindState node off a non-graph northStar goal (P0 #1).
	if d := h.directorForSpeech(); d != nil {
		if g := h.effectiveGoalForSpeech(); g != "" && h.goalGraph.Has(g) {
			for _, subj := range h.relevantLowConf() {
				qid := "confirm:" + subj
				if !h.goalGraph.Has(qid) {
					d.markTopicalQuestion(h, g, "confirm", subj)
				}
				if n, ok := h.goalGraph.Get(qid); ok && eligible(n) {
					return n, true
				}
			}
		}
	}
	return goalgraph.Node{}, false
}

// goalRelevantSubjects returns goal-derived SUBJECTS (not labels): the salientTopic
// of the goal text, of each open blocker label, and of each open requires-subgoal
// label off the effective goal g. Deduped, lowercased. Pure graph read. (Blockers/
// subgoals are LABELS — "where to buy a pickaxe" — so the subject is extracted via
// salientTopic, the same word the knowledge ledger is keyed on. P2 #9: goal-derived,
// NEVER knowledge.All().)
func (h *Host) goalRelevantSubjects(g string) []string {
	if h.goalGraph == nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	push := func(s string) {
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	if t := salientTopic(g); t != "" {
		push(t)
	}
	for _, b := range h.goalGraph.Blockers(g) {
		push(salientTopic(nodeLabel(b)))
	}
	for _, e := range h.goalGraph.Out(g, goalgraph.RelRequires) {
		if n, ok := h.goalGraph.Get(e.To); ok &&
			n.Status != goalgraph.StatusDone && n.Status != goalgraph.StatusAbandoned {
			push(salientTopic(nodeLabel(n)))
		}
	}
	return out
}

// relevantLowConf returns goal-relevant subjects the host holds a SUB-floor belief
// about (0 < conf < epConfFloor) — the #23 broadened-ASK seeds. Goal-derived (never
// knowledge.All(), P2 #9). Capped at epBeliefCap. Pure read.
func (h *Host) relevantLowConf() []string {
	if h.knowledge == nil {
		return nil
	}
	var out []string
	for _, s := range h.goalRelevantSubjects(h.effectiveGoalForSpeech()) {
		if !h.knowledge.Known(s) {
			continue
		}
		f := h.knowledge.Get(s)
		if f.Confidence > 0 && f.Confidence < epConfFloor {
			out = append(out, s)
			if len(out) >= epBeliefCap {
				break
			}
		}
	}
	return out
}

// effectiveGoalForSpeech is the goal the host's speech tier slices blockers off —
// the SAME goal the director plans against (H8), not the raw operator override
// (LiveGoal(), "" in the normal autonomous case). It prefers a live operator
// override, then the director's pure effective-goal read (effectiveGoalView, no
// mutation), then the persona north-star — mirroring the director's resolveGoal
// priority so Pass-1 "prefer the question blocking the goal" actually engages.
// Falls back to LiveGoal()/northStar when no director is bound (REPL/test).
func (h *Host) effectiveGoalForSpeech() string {
	if lg := h.LiveGoal(); lg != "" {
		return lg // operator override wins (same as the director)
	}
	if d := h.directorForSpeech(); d != nil {
		if g := d.effectiveGoalView(h); g != "" {
			return g
		}
	}
	return h.northStar // persona fallback (the director's last resort)
}

// directorForSpeech returns the bound MesaDirector, or nil in REPL/test (no
// conductor). Lets the speech tier reach the director to read the effective goal and
// mint topical questions (pickAskQuestion Pass-3).
func (h *Host) directorForSpeech() *MesaDirector {
	c := h.conductorHandle()
	if c == nil {
		return nil
	}
	// Unwrap wrapper directors (production wires HybridDirector around the
	// MesaDirector). Bounded walk in case wrappers ever nest.
	d := c.director
	for i := 0; i < 4 && d != nil; i++ {
		if md, ok := d.(*MesaDirector); ok {
			return md
		}
		u, ok := d.(interface{ Unwrap() Director })
		if !ok {
			return nil
		}
		d = u.Unwrap()
	}
	return nil
}

// isFactualTipQuestion reports whether a question is answered by a factual TIP (a
// place/item where-to-buy / where-is) rather than a social exchange — the trigger
// for player-oracle promotion. Matches the clean Label form AND the bare qid prefix.
func isFactualTipQuestion(label string) bool {
	low := strings.ToLower(label)
	return strings.Contains(low, "where to buy") ||
		strings.Contains(low, "where is") ||
		strings.HasPrefix(low, "where-to-buy:") ||
		strings.HasPrefix(low, "where-is:")
}

// pickInterlocutor scores the in-range NPCs + players for the question and returns
// the best eligible one (off the pester + same-Q-same-target cooldowns). The
// configured operator — the MENTOR — out-scores everyone when in range; an NPC/
// player whose name overlaps the question label (the shop/quest NPC the question
// implicates) is boosted; ledger familiarity adds a tiebreak; any in-range NPC is
// a weak fallback; a nearby player is a lower-priority (hearsay-tier) fallback.
// Returns ok=false when nothing eligible is in range. Pure + deterministic.
func (h *Host) pickInterlocutor(q goalgraph.Node, now time.Time) (askTarget, bool) {
	w := h.World()
	if w == nil || w.Self == nil {
		return askTarget{}, false
	}
	pos := w.Self.Position()
	lowLabel := strings.ToLower(q.Label)
	// MENTOR-FIRST: the configured hostcfg operator is the host's mentor —
	// "when you are truly stuck, he is the one to ask" — so he out-scores every
	// other candidate (topical-name boost +5 and max relationship weighting +2
	// included) whenever he is in range, and full trust means a recorded
	// grievance never drops him. Same courtesy-trust key as reflexMentor
	// (case-insensitive display name); the analysis-takeover auth elsewhere
	// stays exact-match.
	op := strings.TrimSpace(h.AnalysisOperator())

	var best askTarget
	found := false
	// A factual where-to-buy / where-is question is answered by a free-form PLAYER,
	// not an RSC NPC's canned dialogue tree — promote players above NPCs for it (5b).
	// Social / relationship lines keep the NPC-first default.
	npcBase, playerBase := 1.0, 0.5
	if isFactualTipQuestion(q.Label) {
		npcBase, playerBase = 0.5, 1.0
	}
	consider := func(name, role string, baseScore float64) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		// Skip an NPC already tried for THIS question (5b exhaustion): the canned
		// dialogue tree won't answer a second time — rotate to a new source. Players
		// are NOT skipped (a different player is an independent hearsay source; the
		// pester / same-Q cooldowns already bound re-asking the SAME player).
		if role == "npc" && h.goalGraph != nil && h.goalGraph.HasTag(q.ID, "source-tried:npc:"+name) {
			return
		}
		// Anti-spam: not asked anything of this target recently; not THIS question
		// of THIS target recently.
		h.speech.mu.Lock()
		if t, ok := h.speech.pestered[name]; ok && now.Sub(t) < pesterCooldown {
			h.speech.mu.Unlock()
			return
		}
		if t, ok := h.speech.askedQTo[q.ID+"|"+name]; ok && now.Sub(t) < askSameQSameTargetCooldown {
			h.speech.mu.Unlock()
			return
		}
		h.speech.mu.Unlock()

		mentor := role == "player" && op != "" && strings.EqualFold(name, op)
		score := baseScore
		if mentor {
			score += 10 // the MENTOR outranks any topical NPC or trusted player
		}
		if goalTouch(lowLabel, name) { // the NPC/player the question implicates
			score += 5
		}
		if h.knowledge != nil {
			if f := h.knowledge.Get(name); f.Familiar > 0 { // familiarity tiebreak
				score += 0.1 * float64(f.Familiar)
			}
		}
		// Relationship weighting (§3.4): prefer a TRUSTED / high-affinity target for
		// hearsay; NEVER ask someone you resent (drop them entirely past the grudge
		// threshold) — except the mentor, who is trusted without reserve. Deterministic,
		// ledger-only.
		if h.ledger != nil && h.ledger.Known(name) {
			rel := h.ledger.Rel(name)
			if rel.Grievance >= 0.5 && !mentor {
				return // never ask someone you hold a standing grudge against
			}
			score += 1.5*rel.Trust + 0.5*rel.Affinity - 2.0*rel.Grievance
		}
		if !found || score > best.score {
			best = askTarget{name: name, role: role, score: score}
			found = true
		}
	}

	if w.Npcs != nil {
		for _, n := range w.Npcs.All() {
			if chebyshev(n.X, n.Y, pos.X, pos.Y) > askRadius {
				continue
			}
			name := h.npcNameByType(n.TypeID)
			if name == "" {
				continue // can't address an unnamed NPC in local chat
			}
			consider(name, "npc", npcBase) // any in-range named NPC (below players for factual Qs)
		}
	}
	if w.Players != nil {
		for _, p := range w.Players.All() {
			if p.Name == "" || strings.EqualFold(p.Name, h.opts.Username) {
				continue
			}
			if chebyshev(p.X, p.Y, pos.X, pos.Y) > askRadius {
				continue
			}
			consider(p.Name, "player", playerBase) // a nearby player (promoted for factual where-to-buy/where-is)
		}
	}
	return best, found
}

// askSuspicion returns a "you already suspect: <claim>" hint when the host holds a
// SUB-floor belief about the question's topic, so it asks to CONFIRM rather than
// from zero. Empty when there's no such belief. Cheap; reads the ledger.
func (h *Host) askSuspicion(label string) string {
	if h.knowledge == nil {
		return ""
	}
	topic := salientTopic(label)
	if topic == "" {
		return ""
	}
	f := h.knowledge.Get(topic)
	if len(f.Beliefs) == 0 || f.Confidence >= epConfFloor {
		return "" // nothing held, or already confident (then we wouldn't be asking)
	}
	return "You already suspect (unconfirmed): " + f.Beliefs[0].Claim
}

// --- grounding for the reply branch (Deliverable 2) -------------------------

// groundReply derives the topic of an inbound player line and returns the
// knowledge-grounding context lines to PREPEND to the social reply: a "You KNOW"
// line drawn from the ledger when the host is confident, or an honest "you do NOT
// know — say so, don't bluff" line when it isn't. This makes honesty a host-
// supplied FACT, not a hope about the LLM. It also returns an optional volunteer-
// teach line (the host↔host knowledge-propagation seed) when the player's line
// overlaps a HIGH-confidence belief and the per-player teach cooldown is clear.
//
// known is the C-25 slice the SAME chat turn attaches (chatKnownFacts — computed
// once by the caller so the two carriers share one matcher): whenever it is
// non-empty, the negative "you do NOT know" anchors are suppressed. The slice's
// keyword matcher out-recalls this function's single-token exact Get ("bronze
// pickaxe" answers a "pickaxe" question), and a do-NOT-know order beside an
// attached fact under WHAT YOU ACTUALLY KNOW would hand the model a direct
// contradiction; the answer-only-known contract already orders silence when the
// attached facts don't answer, so nothing negative is needed.
func (h *Host) groundReply(from, message string, now time.Time, known []mesaclient.KnownFact) []string {
	if h.knowledge == nil {
		return nil
	}
	var out []string
	subject := salientTopic(message)
	if subject != "" {
		f := h.knowledge.Get(subject)
		switch {
		case f.Confidence >= epConfFloor && len(f.Beliefs) > 0:
			out = append(out, "You KNOW about "+subject+": "+f.Beliefs[0].Claim+" — answer from this (hedge if unsure).")
		case len(known) > 0:
			// The knowledge slice already carries facts bearing on this line —
			// a "you do NOT know" line here would contradict it inside one
			// prompt. Say nothing; the C-25 contract governs.
		default:
			// Aligned with mesad's answer-only-known reply contract (C-25): no held
			// knowledge ⇒ an unanswerable FACTUAL question gets silence, not an
			// improvised denial — and never a bluff.
			out = append(out, "You do NOT actually know about "+subject+" — nothing in your memory answers it. Do not bluff or make up game facts; if they asked a factual question, you have nothing to answer it with.")
		}
	} else if len(known) == 0 {
		// No salient topic to look up and nothing attached — still anchor the
		// composer's epistemic state so it never invents game facts for a
		// vague/short line. (With facts attached — e.g. "got any tin?", whose
		// 3-char topic this extractor drops but the slice matcher catches — the
		// anchor would be a contradiction, so it is suppressed too.)
		out = append(out, "You have no specific knowledge bearing on this; answer plainly and do not invent game facts.")
	}
	// Volunteer-teach: only a HIGH-confidence belief the player's line touches, and
	// only off the per-player teach cooldown (the host volunteers, doesn't lecture).
	// Compute the grudge suppression FIRST, and burn the per-player teach cooldown
	// (the side-effecting teachable()) ONLY when a teach will actually be emitted
	// (L8) — a grudge-suppressed teach must not consume the cooldown, or a host
	// whose grievance later decays would be wrongly muted for the cooldown window.
	lowMsg := strings.ToLower(message)
	if teach := h.bestTeachable(lowMsg, subject); teach != "" && h.speech != nil {
		// Suppress the VOLUNTEERED teach for a resented party (you don't help
		// someone you hold a grudge against) — the honest answer above stays
		// unconditional (refusing to teach is not the same as lying/bluffing).
		resented := h.ledger != nil && h.ledger.Rel(from).Grievance >= 0.5
		if !resented && h.speech.teachable(from, now) {
			out = append(out, "You could mention (only if it fits): "+teach)
		}
	}
	return out
}

// bestTeachable returns the host's highest-confidence belief whose subject the
// player's line touches (>= teachConfFloor), excluding the one already surfaced as
// the direct answer. Empty when nothing qualifies. Cheap O(facts) scan.
func (h *Host) bestTeachable(lowMsg, answeredSubject string) string {
	var best string
	var bestConf float64
	for _, f := range h.knowledge.All() {
		if f.Confidence < teachConfFloor || len(f.Beliefs) == 0 {
			continue
		}
		if answeredSubject != "" && strings.EqualFold(f.Subject, answeredSubject) {
			continue // don't re-volunteer the direct answer
		}
		if !goalTouch(lowMsg, f.Subject) {
			continue
		}
		if f.Confidence > bestConf {
			bestConf, best = f.Confidence, f.Beliefs[0].Claim
		}
	}
	return best
}

// --- small deterministic helpers --------------------------------------------

// salientTopic extracts the single most salient (longest significant) word from a
// question/utterance label — the subject to look up in the ledger. "" when the
// label carries no significant word. Reuses the reactive stop-word filter.
func salientTopic(label string) string {
	label = strings.ToLower(strings.TrimSpace(label))
	if label == "" {
		return ""
	}
	// Strip a leading "kind:" prefix the goal graph uses on some labels.
	if i := strings.Index(label, ":"); i > 0 && i < 24 {
		label = strings.TrimSpace(label[i+1:])
	}
	var best string
	for _, w := range strings.Fields(label) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) < 4 || stopWord(w) {
			continue
		}
		if len(w) > len(best) {
			best = w
		}
	}
	return best
}

// isHowToProgress reports whether a question is a generic "how do I progress"
// node — these have no clean subject, so the gate skips the do-I-know-it check and
// relies on cooldown + the attempt cap instead.
func isHowToProgress(label string) bool {
	return strings.Contains(strings.ToLower(label), "how-to-progress") ||
		strings.Contains(strings.ToLower(label), "how to progress")
}
