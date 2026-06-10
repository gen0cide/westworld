package mesad

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
	"github.com/gen0cide/westworld/persona"
)

// cron_insight.go is the Phase-4b Tier-2 INSIGHT layer: a second, RARE scheduler
// loop that DRAINS the per-host escalation queue (cron:escalate:queue — a tiny,
// pre-filtered slice the consolidation cron's cheap Haiku already flagged) and
// does the DEEP work with ONE Sonnet call per host per tick. The cost stance
// (docs/world-knowledge-and-learning.md §3.6) is absolute: Sonnet not Opus, only
// on the flagged queue, NEVER on bulk. The deep tasks (§3.5):
//
//   (a) CONTRADICTION RECONCILIATION — direct-observation provenance overrides
//       hearsay; lower the loser's confidence ONCE (sentinel-tag idempotent).
//   (b) CROSS-ENTITY CHAINING — connect related beliefs into goal-graph edges
//       (rune plate ←requires← champions guild ←requires← prereqs).
//   (c) LLM-JUDGED OPEN-QUESTION CLOSURE — close a goal-graph open_question when a
//       distilled belief answers it (semantic, beyond Phase-3a token overlap).
//
// The sentiment overlay (3b deferral) is DEFERRED to 4c; the Sonnet schema
// RESERVES a sentiment_overlays field (parsed-but-ignored here) so 4c is additive.
//
// Invariants: Tier-2 is RARE (empty queue ⇒ zero LLM calls); idempotent + non-
// corrupting (cursor-gated consume + persist-before-advance + per-effect idempotent
// writes — a re-run is a no-op); authority-respecting (knowledge is mesa-distilled,
// the cron MAY write it; the goal graph is the AuthLocal mesa MIRROR — the host
// cold-starts it Empty()-guarded; relationships are UNTOUCHED here). Shares
// s.cronSem / s.cronWG / cronJobTimeout with consolidation (no RPC starvation).

// insightMaxTokens budgets the structured 3-task JSON response.
const insightMaxTokens = 1500

// closeQConf is the confidence floor below which an LLM-judged open-question
// closure is rejected — a weak guess must not prematurely close a question.
const closeQConf = 0.6

// cronInsightCursorKey is the per-host high-water mark for the queue drain: the
// newest qItem.AtUnix the insight cron has already processed (strict `>` select).
const cronInsightCursorKey = "cron:insight:cursor"

// insightCursor is the per-host insight drain cursor: the high-water AtUnix + a
// fail counter for the poison-batch guard (mirrors consolidationCursor).
//
// AtUnix is only second-granular, but recordEscalations stamps every flag from one
// consolidation with the SAME second — so many items can share LastUnix. A bare
// `AtUnix <= LastUnix` skip would then lose cap-dropped same-second siblings forever
// (H16). DoneAtBoundary records the fingerprints of items ALREADY processed at
// exactly LastUnix, so a same-second sibling left over by the per-host cap is still
// selectable next tick (skip only when AtUnix < LastUnix, OR == LastUnix AND its fp
// is in the boundary set). The set is bounded — it only ever holds one second's worth.
type insightCursor struct {
	LastUnix       int64    `json:"last_at_unix"`
	FailCount      int      `json:"fail_count"`
	DoneAtBoundary []string `json:"done_at_boundary,omitempty"`
}

// itemFP is a stable fingerprint of a queue item's identity (subject|claim|reason|
// at), used to track which same-second siblings the cursor has already processed.
func itemFP(it qItem) string {
	return shortFP(it.Subject + "|" + it.Claim + "|" + it.Reason + "|" + fmt.Sprint(it.AtUnix))
}

// alreadyProcessed reports whether the cursor has already consumed an item: any item
// strictly older than the boundary second is done; an item AT the boundary second is
// done only if its fingerprint is in DoneAtBoundary (so cap-dropped same-second
// siblings remain selectable).
func (c insightCursor) alreadyProcessed(it qItem) bool {
	if it.AtUnix < c.LastUnix {
		return true
	}
	if it.AtUnix == c.LastUnix {
		for _, fp := range c.DoneAtBoundary {
			if fp == itemFP(it) {
				return true
			}
		}
	}
	return false
}

// insightLLM returns the Tier-2 (Sonnet, actLLM) seam the insight cron uses.
// Overridable in tests via s.insightLLMOverride. NEVER routes to genesisLLM
// (Opus on the queue is the cardinal cost violation, even though the queue is
// rare). Falls back to decideLLM only if actLLM is nil; nil ⇒ the cron no-ops
// (the queue accumulates, bounded at 200 by recordEscalations).
func (s *Server) insightLLM() completer {
	if s.insightLLMOverride != nil {
		return s.insightLLMOverride
	}
	if s.actLLM != nil {
		return s.actLLM
	}
	if s.decideLLM != nil {
		return s.decideLLM // last-resort fallback; STILL never genesisLLM
	}
	return nil
}

// insightLoop is the Tier-2 RARE loop: every InsightEvery it pulls the recently-
// active hosts (off Postgres) and fans a per-host queue drain through the SAME
// concurrency semaphore as consolidation (so insight cannot starve Act/Chat or
// the bulk cron). Mirrors consolidationLoop exactly.
func (s *Server) insightLoop(ctx context.Context, cfg CronConfig) {
	defer s.cronWG.Done()
	t := time.NewTicker(cfg.InsightEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.insightSweep(ctx, cfg)
		}
	}
}

// insightSweep runs one insight tick: enumerate active hosts and dispatch each
// queue drain through the shared semaphore. Mirrors consolidateSweep — same
// anti-starvation host bound, same per-host job goroutine joined to cronWG, same
// cronJobTimeout. A host with an empty queue makes NO LLM call (insightHost bails).
func (s *Server) insightSweep(ctx context.Context, cfg CronConfig) {
	since := time.Now().Add(-cfg.ActiveWindow).Unix()
	hosts, err := s.ltm.ActiveHostsWithObservations(ctx, since, cfg.MaxHostsPerSweep)
	if err != nil {
		s.log.Warn("crons: insight active-hosts query failed", "err", err)
		return
	}
	if len(hosts) == 0 {
		return
	}
	for _, hostID := range hosts {
		select {
		case <-ctx.Done():
			return
		case s.cronSem <- struct{}{}: // acquire (SHARED with consolidation — bounds aggregate LLM pressure)
		}
		s.cronWG.Add(1)
		go func(id string) {
			defer s.cronWG.Done()
			defer func() { <-s.cronSem }()
			jobCtx, cancel := context.WithTimeout(ctx, cronJobTimeout)
			defer cancel()
			if err := s.insightHost(jobCtx, id, cfg); err != nil {
				s.log.Warn("crons: insight host failed", "host_id", id, "err", err)
			}
		}(hostID)
	}
}

// loadInsightCursor reads a host's insight drain cursor (zero value on a miss).
func (s *Server) loadInsightCursor(ctx context.Context, hostID string) insightCursor {
	var cur insightCursor
	if v, ok, err := s.ltm.GetKV(ctx, hostID, cronInsightCursorKey); err == nil && ok {
		if uerr := json.Unmarshal(v, &cur); uerr != nil {
			s.log.Warn("crons: insight cursor unmarshal failed; resetting", "host_id", hostID, "err", uerr)
			cur = insightCursor{}
		}
	}
	return cur
}

// saveInsightCursor persists a host's insight drain cursor.
func (s *Server) saveInsightCursor(ctx context.Context, hostID string, cur insightCursor) {
	if v, err := json.Marshal(cur); err == nil {
		if err := s.ltm.PutKV(ctx, hostID, cronInsightCursorKey, v); err != nil {
			s.log.Warn("crons: insight cursor save failed", "host_id", hostID, "err", err)
		}
	}
}

// loadEscalationQueue reads the host's escalation queue (the 4a producer's blob).
// Empty/missing/corrupt ⇒ nil (no spurious Tier-2 spend on a bad blob).
func (s *Server) loadEscalationQueue(ctx context.Context, hostID string) []qItem {
	v, ok, err := s.ltm.GetKV(ctx, hostID, cronEscalateQueueKey)
	if err != nil || !ok {
		return nil
	}
	var q []qItem
	if uerr := json.Unmarshal(v, &q); uerr != nil {
		s.log.Warn("crons: escalate-queue unmarshal failed (insight)", "host_id", hostID, "err", uerr)
		return nil
	}
	return q
}

// insightHost drains one host's escalation queue: cursor-gated + persona-filtered
// selection of at most InsightMaxPerHost items, ONE Sonnet call over them, idempotent
// application of the three deep tasks, persist-BEFORE-advance.
//
// The exactly-once contract (CURSOR-gated, not delete-gated): the drain NEVER
// mutates the shared cronEscalateQueueKey blob (4a's recordEscalations appends to it
// concurrently — a read-modify-write delete would race and lose escalations). The
// cursor's LastUnix is the high-water mark; processed items are never re-selected
// (strict `>`); the 200-item bound self-trims old entries. A crash mid-job leaves
// the cursor un-advanced ⇒ items re-selected ⇒ re-applied (SAFE: every write target
// is idempotent). The insight cron writes knowledge/goalgraph ONLY — NEVER back to
// the queue, so there is no infinite escalation loop.
func (s *Server) insightHost(ctx context.Context, hostID string, cfg CronConfig) error {
	cfg = cfg.sane()
	queue := s.loadEscalationQueue(ctx, hostID)
	if len(queue) == 0 {
		return nil // empty queue ⇒ NO Sonnet call (no spurious Tier-2 spend)
	}
	cur := s.loadInsightCursor(ctx, hostID)

	// Snapshot the persona WITHOUT holding s.mu during the LLM call — and capture
	// the cornerstone for the deterministic escalation gate.
	var pers *persona.Persona
	personaProse := ""
	if e, ok := s.lookup(hostID); ok {
		p := e.persona
		pers = &p
		personaProse = e.prose
	}

	// maxAt = newest AtUnix across the FULL read. It is the cursor target ONLY when
	// EVERY item newer than the cursor was either processed or deliberately dropped
	// this tick — never when the per-host cap left newer ADMITTED items unprocessed
	// (H16: advancing to maxAt then silently loses the cap-dropped overflow forever).
	var maxAt int64
	for _, it := range queue {
		if it.AtUnix > maxAt {
			maxAt = it.AtUnix
		}
	}

	// Cursor + persona filter: split the items newer than the cursor into the ones
	// the persona ADMITS (candidates) and the ones it DROPS (disposed-without-work).
	// Sort the admitted set OLDEST-first so the per-host cap keeps the OLDEST items
	// and leaves the NEWER overflow for the next tick (H16).
	candidates := make([]qItem, 0, len(queue))
	dropped := make([]qItem, 0, len(queue))
	for _, it := range queue {
		if cur.alreadyProcessed(it) {
			continue // already consumed (high-water mark + same-second boundary set)
		}
		if !shouldProcessEscalation(pers, it, cfg) {
			dropped = append(dropped, it) // persona dropped it (oblivious host forgets the nuance)
			continue
		}
		candidates = append(candidates, it)
	}
	sort.SliceStable(candidates, func(i, j int) bool { return candidates[i].AtUnix < candidates[j].AtUnix })
	var overflow []qItem
	if len(candidates) > cfg.InsightMaxPerHost {
		overflow = candidates[cfg.InsightMaxPerHost:] // newer admitted items deferred to next tick
		candidates = candidates[:cfg.InsightMaxPerHost]
	}

	// processedMaxAt = the newest AtUnix the cursor may advance to this tick. It may
	// reach the oldest deferred-overflow second (so same-second admitted items we DID
	// keep land in the boundary set), but never EXCEED it — an overflow item beyond
	// the cursor would be lost forever (H16). Same-second overflow is protected by the
	// boundary fingerprint set (its fp is never recorded ⇒ still selectable next tick).
	// The cursor is the max over everything actually disposed of (admitted-kept ∪
	// persona-dropped) within that ceiling.
	overflowMin := int64(0) // 0 ⇒ no overflow
	if len(overflow) > 0 {
		overflowMin = overflow[0].AtUnix // oldest deferred item (sorted oldest-first)
	}
	processedMaxAt := cur.LastUnix
	consider := func(at int64) {
		if overflowMin != 0 && at > overflowMin {
			return // never advance PAST a deferred-overflow item
		}
		if at > processedMaxAt {
			processedMaxAt = at
		}
	}
	for _, it := range candidates {
		consider(it.AtUnix)
	}
	for _, it := range dropped {
		consider(it.AtUnix)
	}

	if len(candidates) == 0 {
		// Nothing admitted this tick. Advance past the persona-dropped items so the
		// cursor cannot wedge on items the persona will never admit (no overflow can
		// exist here — overflow implies admitted candidates). Only write when there is
		// actually progress to record (avoid a no-op KV write on an idle re-run).
		if len(dropped) > 0 {
			s.advanceCursor(ctx, hostID, &cur, processedMaxAt, candidates, dropped)
		}
		return nil
	}

	// Load the current durable knowledge + goal graph (the in-memory write targets).
	//
	// M19: a Knowledge()/GoalGraph() read error must NOT be swallowed. applyReconcile
	// folds onto `led` and SyncKnowledge does a per-subject REPLACE, so reconciling
	// onto an empty ledger after a transient read error would overwrite each touched
	// subject's rich DB row with a single-belief impoverished version (permanent loss
	// once the cursor advances). Treat the read error like the LLM/persist failures:
	// bump fail_count, DO NOT advance the cursor, return the error to retry the tick
	// (mirrors consolidateHost's knowledge-read guard).
	led := knowledge.NewLedger()
	existing, kerr := s.ltm.Knowledge(ctx, hostID)
	if kerr != nil {
		cur.FailCount++
		if cur.FailCount >= maxFailBeforeSkip {
			s.log.Warn("crons: insight knowledge read keeps failing, advancing past batch", "host_id", hostID, "fails", cur.FailCount, "err", kerr)
			cur.LastUnix = maxAt
			cur.DoneAtBoundary = nil
			cur.FailCount = 0
		}
		s.saveInsightCursor(ctx, hostID, cur)
		return fmt.Errorf("read knowledge: %w", kerr)
	}
	led.Import(entriesToKnowledge(existing))
	graph := goalgraph.New()
	snap, gerr := s.ltm.GoalGraph(ctx, hostID)
	if gerr != nil {
		cur.FailCount++
		if cur.FailCount >= maxFailBeforeSkip {
			s.log.Warn("crons: insight goal-graph read keeps failing, advancing past batch", "host_id", hostID, "fails", cur.FailCount, "err", gerr)
			cur.LastUnix = maxAt
			cur.DoneAtBoundary = nil
			cur.FailCount = 0
		}
		s.saveInsightCursor(ctx, hostID, cur)
		return fmt.Errorf("read goal-graph: %w", gerr)
	}
	if snap != nil {
		graph.Import(goalGraphSnapshotToInternal(snap))
	}

	res, err := s.runInsight(ctx, hostID, personaProse, candidates, led, graph)
	if err != nil {
		// LLM/parse failure: bump fail_count, DO NOT advance the cursor (retry next
		// tick — idempotency tags make a re-run safe), advance past a poison batch
		// only after the ceiling.
		cur.FailCount++
		if cur.FailCount >= maxFailBeforeSkip {
			s.log.Warn("crons: insight poison batch, advancing past it", "host_id", hostID, "fails", cur.FailCount)
			cur.LastUnix = maxAt // explicit escape: skip the WHOLE wedged batch
			cur.DoneAtBoundary = nil
			cur.FailCount = 0
		}
		s.saveInsightCursor(ctx, hostID, cur)
		return err
	}

	applied := applyReconcile(led, res.Reconcile) +
		applyChains(graph, res.Chain) +
		applyCloseQuestions(graph, res.CloseQuestions)

	// Persist BEFORE advancing the cursor (the ordering invariant). A persist
	// failure leaves the cursor un-advanced ⇒ re-processed next tick with no double
	// effect (the idempotency tags absorb the retry).
	rows := gcKnowledge(led.Export(), cfg, time.Now())
	if _, werr := s.ltm.SyncKnowledge(ctx, hostID, knowledgeToEntries(rows)); werr != nil {
		cur.FailCount++
		s.saveInsightCursor(ctx, hostID, cur)
		return fmt.Errorf("insight sync knowledge: %w", werr)
	}
	// H17 (server-side half): the goal_graphs row has TWO whole-blob writers — this
	// cron and the host's 30s flush. A wholesale Export()/replace here races the
	// host's flush across the LLM call window (seconds): a host snapshot written
	// while runInsight was in flight would be clobbered by graph.Export() (the host's
	// live status/new nodes lost), and symmetrically the host's next flush clobbers
	// the cron's chains/closures. Re-read the LATEST snapshot now and MERGE the cron's
	// applied chains/closures into it (idempotent appliers ⇒ re-application is a
	// no-op for what is already there), so the cron only ADDS its nodes/edges/closures
	// and never deletes a node it doesn't own. (The host's flush still clobbers on its
	// own write — that half needs the host + Server.SyncGoalGraph to merge, recorded
	// in crossPackageNeeds.)
	if gerr := s.mergeGoalGraph(ctx, hostID, res); gerr != nil {
		cur.FailCount++
		s.saveInsightCursor(ctx, hostID, cur)
		return fmt.Errorf("insight sync goal-graph: %w", gerr)
	}

	// Advance only to the newest item we actually processed/disposed of this tick —
	// NOT past cap-dropped admitted overflow, which must survive for the next tick
	// (H16). When nothing was capped this is maxAt (the whole read was disposed of).
	s.advanceCursor(ctx, hostID, &cur, processedMaxAt, candidates, dropped)
	s.log.Info("crons: insight applied", "host_id", hostID,
		"queued", len(queue), "admitted", len(candidates), "applied", applied)
	return nil
}

// advanceCursor moves the per-host insight drain cursor to target (the newest AtUnix
// safely disposed of this tick) and rebuilds the same-second boundary set so cap-
// dropped same-second siblings remain re-selectable (H16). It folds in only the
// disposed items (admitted-kept ∪ persona-dropped) whose AtUnix == target; when the
// target second equals the prior boundary it ACCUMULATES (a later tick that processes
// more same-second siblings must remember the earlier ones too). FailCount resets on
// any successful advance.
func (s *Server) advanceCursor(ctx context.Context, hostID string, cur *insightCursor, target int64, disposed ...[]qItem) {
	if target < cur.LastUnix {
		target = cur.LastUnix // never go backwards
	}
	var boundary []string
	if target == cur.LastUnix {
		boundary = append(boundary, cur.DoneAtBoundary...) // accumulate within the same second
	}
	seen := map[string]bool{}
	for _, fp := range boundary {
		seen[fp] = true
	}
	for _, batch := range disposed {
		for _, it := range batch {
			if it.AtUnix != target {
				continue
			}
			fp := itemFP(it)
			if !seen[fp] {
				seen[fp] = true
				boundary = append(boundary, fp)
			}
		}
	}
	cur.LastUnix = target
	cur.DoneAtBoundary = boundary
	cur.FailCount = 0
	s.saveInsightCursor(ctx, hostID, *cur)
}

// mergeGoalGraph persists the cron's applied chains/closures by RE-READING the latest
// durable snapshot and re-applying the (idempotent) graph effects onto it, rather than
// clobbering the row with a stale Export() (H17 server-side half — closes the cron-vs-
// host write-window race across the LLM call). A nil result with nothing to write is a
// no-op (never touches the row).
func (s *Server) mergeGoalGraph(ctx context.Context, hostID string, res insightResult) error {
	if len(res.Chain) == 0 && len(res.CloseQuestions) == 0 {
		return nil // nothing graph-shaped to merge this tick
	}
	latest := goalgraph.New()
	if snap, gerr := s.ltm.GoalGraph(ctx, hostID); gerr == nil && snap != nil {
		latest.Import(goalGraphSnapshotToInternal(snap))
	} else if gerr != nil {
		return gerr // a read error must not silently drop the row into an empty merge
	}
	applyChains(latest, res.Chain)
	applyCloseQuestions(latest, res.CloseQuestions)
	return s.ltm.SyncGoalGraph(ctx, hostID, internalToGoalGraphSnapshot(latest.Export()))
}

// shouldProcessEscalation is the PERSONA-MODULATED escalation gate (§3.6 "the cost
// knob IS a personality knob"). It is DETERMINISTIC (no RNG — idempotent across
// ticks, so a re-run admits the same items). A sharp/curious host (high
// BulkApperception + the matching curiosity flavor) chews on more (escalates more,
// costs more); an oblivious one drops marginal items (genuinely forgets the nuance
// — intended). A persona we can't resolve fails OPEN (admit). The admitted count is
// always capped at InsightMaxPerHost upstream, so even the sharpest host is cost-bounded.
func shouldProcessEscalation(p *persona.Persona, item qItem, cfg CronConfig) bool {
	if p == nil {
		return true // fail-open: unknown persona ⇒ admit (don't silently lose escalations)
	}
	bulk := clampUnit(p.Cornerstone.Prefs.BulkApperception.Mu)
	flavor := clampUnit(curiosityFlavor(p.Cornerstone.Prefs.Curiosity, item))
	drive := 0.6*bulk + 0.4*flavor
	// The persona drive MODULATES the admission bar — it must not appear on both
	// sides of the comparison (a self-referential gate collapses to a drive-only
	// cutoff that ignores EscalateThreshold; M18). A sharp/curious host (high drive)
	// LOWERS the bar (admit readily); an oblivious host (low drive) RAISES it
	// (admit little). drive=1 → 0.5× the configured center; drive=0.5 → the center
	// itself; drive=0 → 1.5× the center. Until a real per-item salience term lands
	// (4c, qItem gains a Salience field), every flagged escalation carries the same
	// fixed neutral reference salience (it already cleared the cheap Haiku triage);
	// admission is then a true persona-vs-bar decision, not a self-comparison.
	const itemSalience = 0.5 // neutral reference; replace with item.Salience in 4c
	bar := cfg.EscalateThreshold * (1.5 - drive)
	return itemSalience >= bar
}

// curiosityFlavor maps an escalation item to the persona's matching curiosity
// flavor by its subject/reason words: player/NPC→Social, place/route→Spatial,
// recipe/skill/quest→Skill, shop/price→Economic, combat/pk→Risk. The default
// (no clear match) is the host's mean curiosity, so an item with no flavor still
// gets a fair, persona-shaped read.
func curiosityFlavor(c persona.Curiosity, item qItem) float64 {
	hay := strings.ToLower(item.Subject + " " + item.Reason + " " + item.Claim)
	switch {
	case containsAny(hay, "player", "npc", "guild", "friend", "ask", "told", "rude", "tone", "social"):
		return c.Social
	case containsAny(hay, "place", "route", "location", "map", "bank", "mine", "area", "where", "path"):
		return c.Spatial
	case containsAny(hay, "recipe", "skill", "quest", "craft", "smith", "prereq", "how to", "how-to", "requires"):
		return c.Skill
	case containsAny(hay, "shop", "price", "sell", "buy", "trade", "coin", "gp", "cost"):
		return c.Economic
	case containsAny(hay, "combat", "pk", "fight", "attack", "danger", "kill", "wilderness"):
		return c.Risk
	default:
		return (c.Social + c.Spatial + c.Skill + c.Economic + c.Risk) / 5.0
	}
}

func containsAny(hay string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(hay, n) {
			return true
		}
	}
	return false
}

func clampUnit(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// --- the Sonnet structured call ---------------------------------------------

// insightResult is the parsed Tier-2 output. SentimentOverlays is RESERVED (parsed
// but ignored in 4b — the 4c sentiment-overlay seam).
type insightResult struct {
	Reconcile         []reconcileOp     `json:"reconcile"`
	Chain             []chainOp         `json:"chain"`
	CloseQuestions    []closeQuestionOp `json:"close_questions"`
	SentimentOverlays []json.RawMessage `json:"sentiment_overlays"` // 4c seam — ignored here
}

// reconcileOp resolves a contradiction: lower the loser's confidence (weaker
// provenance), optionally reinforce the winner.
type reconcileOp struct {
	Subject     string `json:"subject"`
	LoserClaim  string `json:"loser_claim"`
	WinnerClaim string `json:"winner_claim"`
}

// chainOp is one cross-entity chain: a set of stable-slug nodes linked by typed
// edges. The server slugifies every id on parse so re-runs produce identical nodes.
type chainOp struct {
	Nodes []chainNode `json:"nodes"`
	Links []chainLink `json:"links"`
}

type chainNode struct {
	ID    string `json:"id"`
	Kind  string `json:"kind"`
	Label string `json:"label"`
}

type chainLink struct {
	From string `json:"from"`
	To   string `json:"to"`
	Rel  string `json:"rel"`
}

// closeQuestionOp closes a goal-graph open_question that a belief answers.
type closeQuestionOp struct {
	QuestionID    string  `json:"question_id"`
	AnsweringNode string  `json:"answering_node"`
	Confidence    float64 `json:"confidence"`
}

// insightSystem is the STABLE Tier-2 analyzer prefix — identical across every host
// and tick. It is NOT marked Cache:true: at ~417-556 tokens it is below Sonnet's
// 2048-token cache minimum, so a cache_control breakpoint would be silently ignored
// (cache_creation/read stay 0). We price it as full input rather than pretend it
// caches (H18 — see runInsight for the full rationale).
const insightSystem = `You are a deep memory-reasoning analyzer for a RuneScape Classic player agent. You are given a SMALL set of flagged claims that earlier triage marked for deeper reasoning, plus the agent's current beliefs and open questions. You are NOT the character — do not roleplay. Be literal.

Do THREE kinds of deep work and return ONLY JSON:
{
  "reconcile": [{"subject":"<subject>","loser_claim":"<the WEAKER, contradicted claim>","winner_claim":"<the claim that overrides it>"}],
  "chain": [{"nodes":[{"id":"<stable-slug>","kind":"<goal|subgoal|open_goal|state>","label":"<short label>"}],"links":[{"from":"<slug>","to":"<slug>","rel":"<requires|produces|enables|blocked_by|serves>"}]}],
  "close_questions": [{"question_id":"<exact open-question id given>","answering_node":"<slug or subject that answers it>","confidence":0.0-1.0}]
}

Rules:
- RECONCILE only genuine contradictions. Provenance is load-bearing: direct observation (saw-it/did-it) and server messages OVERRIDE hearsay (player/NPC told). The loser is the lower-provenance / contradicted claim.
- CHAIN connects related beliefs into a dependency structure (e.g. "rune plate" requires "champions-guild-quest" requires its prereqs). Use STABLE, lowercase, hyphenated SLUG ids (e.g. "champions-guild-quest") so the same chain re-emits identically. Do NOT invent edges you cannot justify from the claims/beliefs.
- CLOSE a question ONLY if a belief genuinely answers it; set confidence to how sure you are (a weak guess is low confidence and will be ignored). Use the EXACT question id from the open-questions list.
- Omit any array you have nothing for. If there is nothing to do, return {}.

Reply with JSON only.`

// runInsight runs the single Sonnet (actLLM) call over the admitted batch. The
// stable analyzer prompt is the cached prefix; the per-host persona + touched
// beliefs + open questions are the uncached block; the admitted items are the
// user turn. Returns the parsed (slugified) result.
func (s *Server) runInsight(ctx context.Context, hostID, persona string, items []qItem, led *knowledge.Ledger, graph *goalgraph.Graph) (insightResult, error) {
	tier := s.insightLLM()
	if tier == nil {
		return insightResult{}, fmt.Errorf("no insight LLM configured")
	}

	// Bounded touched-subject beliefs snippet (the subjects the items reference).
	var beliefs strings.Builder
	touched := map[string]bool{}
	for _, it := range items {
		subj := strings.TrimSpace(it.Subject)
		if subj == "" || touched[strings.ToLower(subj)] {
			continue
		}
		touched[strings.ToLower(subj)] = true
		f := led.Get(subj)
		for i, b := range f.Beliefs {
			if i >= 3 || beliefs.Len() > 800 {
				break
			}
			fmt.Fprintf(&beliefs, "- %s [%s, conf=%.2f]: %s\n", subj, b.Provenance, b.Confidence(), b.Claim)
		}
	}

	// Bounded open-questions list (the closure candidates).
	var oqs strings.Builder
	for i, n := range graph.OpenQuestions() {
		if i >= 12 || oqs.Len() > 600 {
			break
		}
		fmt.Fprintf(&oqs, "- id=%s: %s\n", n.ID, n.Label)
	}

	var ctxBlock strings.Builder
	if p := strings.TrimSpace(persona); p != "" {
		ctxBlock.WriteString("# YOUR CHARACTER\n" + p + "\n\n")
	}
	if beliefs.Len() > 0 {
		ctxBlock.WriteString("# RELEVANT CURRENT BELIEFS (provenance-tagged)\n" + beliefs.String() + "\n")
	}
	if oqs.Len() > 0 {
		ctxBlock.WriteString("# OPEN QUESTIONS (close one ONLY if a belief answers it; use the exact id)\n" + oqs.String())
	}

	// H18: insightSystem is ~417-556 tokens on claude-sonnet-4-6, FAR below Sonnet's
	// 2048-token minimum cacheable prefix, so a cache_control breakpoint here is
	// SILENTLY ignored by Anthropic (never written, never read — cache_creation/read
	// stay 0). We do NOT mark it Cache:true: pretending it caches defeated the cost
	// model invisibly. We deliberately do NOT pad it over the minimum by prepending the
	// unrelated ~4000-token dslManual either — that would bloat every insight request
	// (and the per-call cache-read at 10% of the padded size roughly cancels the saving
	// on this short prefix). The honest budget is full-input pricing on a short prefix
	// per Sonnet call; re-tune InsightEvery/InsightMaxPerHost from that (mirrors the
	// consolidateSystem decision in cron.go).
	// M20: only append the per-host context block when it is non-empty; a cold host
	// (no persona, no beliefs, no open questions) would otherwise send an empty
	// {type:"text",text:""} block, which the Messages API rejects with 400 and the
	// cron retries forever (wedging the insight drain). Mirrors extractClaims.
	blocks := []llm.SystemBlock{
		{Text: insightSystem, Cache: false},
	}
	if hostCtx := strings.TrimSpace(ctxBlock.String()); hostCtx != "" {
		blocks = append(blocks, llm.SystemBlock{Text: hostCtx, Cache: false})
	}

	var user strings.Builder
	user.WriteString("Flagged claims for deep reasoning:\n")
	for i, it := range items {
		fmt.Fprintf(&user, "%d. subject=%q claim=%q reason=%q\n", i+1, it.Subject, it.Claim, it.Reason)
	}
	user.WriteString("\nReconcile contradictions, chain related beliefs, and close any open question a belief answers. Use stable slug ids. Return JSON only.")

	raw, err := tier.CompleteSystem(ctx, blocks, user.String(), insightMaxTokens)
	if err != nil {
		return insightResult{}, err
	}
	return parseInsight(raw)
}

// parseInsight maps the Sonnet JSON to insightResult, slugifying every emitted
// graph node id (the dedup guarantee for chaining). A parse failure errors (the
// cron treats it as a retryable batch); a well-formed empty {} returns a zero
// result with no error.
func parseInsight(raw string) (insightResult, error) {
	js := extractJSON(raw)
	if js == "" {
		return insightResult{}, fmt.Errorf("no JSON in insight response")
	}
	var v insightResult
	if err := json.Unmarshal([]byte(js), &v); err != nil {
		return insightResult{}, fmt.Errorf("insight JSON: %w", err)
	}
	// Server-side slugifier: normalize all graph node ids so non-deterministic ids
	// can't spawn duplicate nodes across ticks (the chaining idempotency guarantee).
	for i := range v.Chain {
		for j := range v.Chain[i].Nodes {
			v.Chain[i].Nodes[j].ID = slugify(v.Chain[i].Nodes[j].ID)
		}
		for j := range v.Chain[i].Links {
			v.Chain[i].Links[j].From = slugify(v.Chain[i].Links[j].From)
			v.Chain[i].Links[j].To = slugify(v.Chain[i].Links[j].To)
		}
	}
	for i := range v.CloseQuestions {
		v.CloseQuestions[i].AnsweringNode = slugify(v.CloseQuestions[i].AnsweringNode)
		// question_id is matched against existing graph ids; the graph normalizes on
		// lookup (norm() lowercases+trims), so leave it as-emitted.
	}
	return v, nil
}

// --- the three idempotent appliers ------------------------------------------

// reconcileWeight is the β bump applied to a contradicted (loser) claim — enough
// to meaningfully lower its Beta-mean confidence, applied EXACTLY ONCE per
// contradiction (sentinel-tag guarded).
const reconcileWeight = 2.0

// applyReconcile lowers each loser claim's confidence ONCE (the no-thrash guard).
// knowledge.Observe(false) stacks β on re-runs, so before applying we check the
// subject's tags for reconciled:<fp>; present ⇒ SKIP (already reconciled). Absent
// ⇒ apply one β bump, optionally reinforce the winner, then tag. One contradiction
// = one β bump, forever. (The cursor is the first defense; the tag is belt-and-
// suspenders for at-least-once retries.) Returns the number applied.
func applyReconcile(led *knowledge.Ledger, ops []reconcileOp) int {
	n := 0
	for _, op := range ops {
		subj := strings.TrimSpace(op.Subject)
		loser := strings.TrimSpace(op.LoserClaim)
		if subj == "" || loser == "" {
			continue
		}
		// Key the sentinel on (subject|loser) ONLY, NOT the winner: a re-run that
		// names a DIFFERENT winner for the same loser must still be a no-op, else the
		// loser's confidence gets β-bumped twice. "Have I already lowered this loser?"
		// is the idempotent question — independent of which claim won.
		fp := shortFP(subj + "|" + loser)
		tag := "reconciled:" + fp
		f := led.Get(subj)
		already := false
		for _, t := range f.Tags {
			if t == tag {
				already = true
				break
			}
		}
		if already {
			continue // already reconciled — no second β bump (idempotent)
		}
		led.Observe(subj, loser, false, reconcileWeight) // lower the loser's confidence ONCE
		if w := strings.TrimSpace(op.WinnerClaim); w != "" {
			led.Observe(subj, w, true, 1.0) // mild reinforcement of the winner
		}
		led.Tag(subj, tag)
		n++
	}
	return n
}

// applyChains writes cross-entity chains into the goal graph. Both Upsert (dedup
// by normalized id, empty fields preserve existing) and Link (dedup by (from,to,
// rel)) are inherently idempotent, so re-running canned chains yields the IDENTICAL
// graph. The slugified ids (parseInsight) guarantee stable node identity. Returns
// the number of link ops written.
func applyChains(graph *goalgraph.Graph, ops []chainOp) int {
	n := 0
	for _, op := range ops {
		for _, nd := range op.Nodes {
			id := slugify(nd.ID)
			if id == "" {
				continue
			}
			kind := strings.TrimSpace(nd.Kind)
			if kind == "" {
				kind = goalgraph.KindState
			}
			graph.Upsert(id, kind, strings.TrimSpace(nd.Label), "") // status untouched (preserves existing)
		}
		for _, lk := range op.Links {
			from, to := slugify(lk.From), slugify(lk.To)
			rel := normalizeRel(lk.Rel)
			if from == "" || to == "" || rel == "" {
				continue
			}
			graph.Link(from, to, rel)
			n++
		}
	}
	return n
}

// applyCloseQuestions closes goal-graph open questions a belief answers, gated by
// the closeQConf floor (a weak guess never closes). SetStatus(Done) is monotone/
// idempotent and OpenQuestions() excludes Done, so a closed question never re-enters
// candidacy ⇒ no LLM churn. Tags answered_by:<fp> for provenance audit and links
// the answering node when given. Returns the number closed.
func applyCloseQuestions(graph *goalgraph.Graph, ops []closeQuestionOp) int {
	n := 0
	for _, op := range ops {
		qid := strings.TrimSpace(op.QuestionID)
		if qid == "" || op.Confidence < closeQConf {
			continue
		}
		node, ok := graph.Get(qid)
		if !ok || node.Kind != goalgraph.KindOpenQuestion {
			continue // only close an existing open_question (no graph growth via closure)
		}
		if node.Status == goalgraph.StatusDone || node.Status == goalgraph.StatusAbandoned {
			continue // already resolved — idempotent
		}
		graph.SetStatus(qid, goalgraph.StatusDone)
		graph.Tag(qid, "answered_by:"+shortFP(qid+"|"+op.AnsweringNode))
		if ans := slugify(op.AnsweringNode); ans != "" && graph.Has(ans) {
			graph.Link(qid, ans, goalgraph.RelProduces)
		}
		n++
	}
	return n
}

// --- helpers ----------------------------------------------------------------

// slugify normalizes an LLM-emitted node id to a stable lowercase hyphenated slug
// so re-runs can't spawn duplicate nodes (the chaining idempotency guarantee).
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevDash = false
		case r == ' ' || r == '-' || r == '_' || r == '/' || r == '.':
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// normalizeRel maps an LLM-emitted relation to a goalgraph Rel* constant (empty if
// unrecognized — the link is then dropped, never a garbage edge).
func normalizeRel(rel string) string {
	switch strings.ToLower(strings.TrimSpace(rel)) {
	case goalgraph.RelRequires, "require", "needs", "depends-on":
		return goalgraph.RelRequires
	case goalgraph.RelProduces, "produce", "yields", "gives":
		return goalgraph.RelProduces
	case goalgraph.RelEnables, "enable", "allows":
		return goalgraph.RelEnables
	case goalgraph.RelBlockedBy, "blocked-by", "blocks":
		return goalgraph.RelBlockedBy
	case goalgraph.RelServes, "serve", "feeds":
		return goalgraph.RelServes
	default:
		return ""
	}
}

// shortFP is a short, stable fingerprint of a string (for the idempotency
// sentinel tags). Collisions are tolerable: the worst case is skipping a genuinely
// distinct reconciliation, which is far safer than thrashing a belief's confidence.
func shortFP(s string) string {
	sum := sha1.Sum([]byte(s))
	return hex.EncodeToString(sum[:6])
}

// --- goal-graph proto <-> internal converters (mesa-side, mirrors knowledge) -

// goalGraphSnapshotToInternal converts a wire snapshot into the internal form.
func goalGraphSnapshotToInternal(snap *mesapb.GoalGraphSnapshot) goalgraph.Snapshot {
	out := goalgraph.Snapshot{}
	for _, n := range snap.GetNodes() {
		out.Nodes = append(out.Nodes, goalgraph.Node{
			ID: n.GetId(), Kind: n.GetKind(), Label: n.GetLabel(), Status: n.GetStatus(),
			Progress: n.GetProgress(), Tags: n.GetTags(), At: n.GetAtUnix(),
		})
	}
	for _, e := range snap.GetEdges() {
		out.Edges = append(out.Edges, goalgraph.Edge{From: e.GetFrom(), To: e.GetTo(), Rel: e.GetRel()})
	}
	return out
}

// internalToGoalGraphSnapshot converts the internal snapshot to the wire form
// (At ↔ AtUnix rename).
func internalToGoalGraphSnapshot(s goalgraph.Snapshot) *mesapb.GoalGraphSnapshot {
	out := &mesapb.GoalGraphSnapshot{}
	for _, n := range s.Nodes {
		out.Nodes = append(out.Nodes, &mesapb.GoalGraphNode{
			Id: n.ID, Kind: n.Kind, Label: n.Label, Status: n.Status,
			Progress: n.Progress, Tags: n.Tags, AtUnix: n.At,
		})
	}
	for _, e := range s.Edges {
		out.Edges = append(out.Edges, &mesapb.GoalGraphEdge{From: e.From, To: e.To, Rel: e.Rel})
	}
	return out
}
