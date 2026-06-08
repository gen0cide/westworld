package mesad

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gen0cide/westworld/cognition/knowledge"
	"github.com/gen0cide/westworld/mesa/llm"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// cron.go is the Phase-4 System-2 distillation: async, bounded, idempotent
// background loops that fold the raw observation firehose into the durable
// world-knowledge ledger. The host stays light; ALL the LLM work is here. The
// invariants (docs/world-knowledge-and-learning.md §3.5/§3.6):
//
//   - Tier-0 (NO LLM): novelty/dedup, salience-decay GC, recency/frequency.
//   - Tier-1 (Haiku, decideLLM): the BULK — batched per-host claim extraction.
//   - Tier-2 (Sonnet/Opus): RARE, only on the flagged slice (the insight cron,
//     4b — this slice only RECORDS the escalation flag to a durable queue).
//
// Crons NEVER hold s.mu during an LLM call (snapshot persona, release, then
// call), run host iteration off Postgres (not the in-mem registry), and are
// concurrency-capped (cronSem) so they cannot starve Act/Chat/Decide.

// CronConfig tunes the distillation loops. Defaults via DefaultCronConfig; the
// operator overrides the cost-dominant knobs (ConsolidateEvery, BatchSize) and
// the starvation guard (Concurrency) as cmd/mesad flags.
type CronConfig struct {
	ConsolidateEvery  time.Duration // how often the consolidation loop ticks
	BatchSize         int           // observations digested per Haiku call (the core cost lever)
	Concurrency       int           // in-flight per-host LLM jobs (the starvation guard)
	MaxHostsPerSweep  int           // anti-starvation bound per tick
	ActiveWindow      time.Duration // an observation within this window marks a host "active"
	KnowledgeTTL      time.Duration // Tier-0 GC: stale + low-confidence + low-encounter subjects pruned
	MaxSubjects       int           // Tier-0 size bound per host
	EscalateThreshold float64       // 4a: record-only; reserved for the 4b persona-modulated gate
}

// DefaultCronConfig is the go-live default (operator-confirmable knobs noted in
// the plan). 60s × 50-obs batches × concurrency 4 keeps a 200-host fleet's Haiku
// volume sane (mostly cache-hit on the stable analyzer prefix).
func DefaultCronConfig() CronConfig {
	return CronConfig{
		ConsolidateEvery:  60 * time.Second,
		BatchSize:         50,
		Concurrency:       4,
		MaxHostsPerSweep:  64,
		ActiveWindow:      24 * time.Hour,
		KnowledgeTTL:      30 * 24 * time.Hour,
		MaxSubjects:       500,
		EscalateThreshold: 0.5,
	}
}

// sane fills any zero field with its default so a partially-specified config is
// always usable.
func (c CronConfig) sane() CronConfig {
	d := DefaultCronConfig()
	if c.ConsolidateEvery <= 0 {
		c.ConsolidateEvery = d.ConsolidateEvery
	}
	if c.BatchSize <= 0 {
		c.BatchSize = d.BatchSize
	}
	if c.Concurrency <= 0 {
		c.Concurrency = d.Concurrency
	}
	if c.MaxHostsPerSweep <= 0 {
		c.MaxHostsPerSweep = d.MaxHostsPerSweep
	}
	if c.ActiveWindow <= 0 {
		c.ActiveWindow = d.ActiveWindow
	}
	if c.KnowledgeTTL <= 0 {
		c.KnowledgeTTL = d.KnowledgeTTL
	}
	if c.MaxSubjects <= 0 {
		c.MaxSubjects = d.MaxSubjects
	}
	return c
}

// completer is the narrow LLM seam the consolidation cron uses — satisfied by
// *llm.Client (CompleteSystem). It exists so tests can inject a fake (no real
// API calls) and so the cron is decoupled from the concrete client.
type completer interface {
	CompleteSystem(ctx context.Context, blocks []llm.SystemBlock, user string, maxTokens int) (string, error)
}

// consolidateLLM returns the Tier-1 (Haiku) seam the bulk cron unconditionally
// uses. Overridable in tests via s.consolidateLLMOverride. NEVER routes to
// actLLM/genesisLLM — Opus on bulk is the cardinal cost violation.
func (s *Server) consolidateLLM() completer {
	if s.consolidateLLMOverride != nil {
		return s.consolidateLLMOverride
	}
	if s.decideLLM == nil {
		return nil
	}
	return s.decideLLM
}

// StartCrons launches the background distillation loops. Safe to call once after
// Attach and before Serve. A nil LTM (no durable store) means there is nothing
// to read or write, so it no-ops. cfg is sanitized; zero fields take defaults.
func (s *Server) StartCrons(ctx context.Context, cfg CronConfig) {
	if s.ltm == nil {
		s.log.Info("crons: no LTM wired; distillation disabled")
		return
	}
	cfg = cfg.sane()
	cctx, cancel := context.WithCancel(ctx)
	s.cronCancel = cancel
	s.cronSem = make(chan struct{}, cfg.Concurrency)
	s.cronWG.Add(1)
	go s.consolidationLoop(cctx, cfg)
	s.log.Info("crons: started", "consolidate_every", cfg.ConsolidateEvery,
		"batch_size", cfg.BatchSize, "concurrency", cfg.Concurrency,
		"max_hosts_per_sweep", cfg.MaxHostsPerSweep)
}

// StopCrons cancels the loops and waits for in-flight folds to finish (so a
// SIGTERM never leaves a torn knowledge write). Idempotent / safe when crons
// never started.
func (s *Server) StopCrons() {
	if s.cronCancel != nil {
		s.cronCancel()
	}
	s.cronWG.Wait()
}

// consolidationLoop is the Tier-1 bulk loop: every ConsolidateEvery it pulls the
// recently-active hosts (off Postgres, NOT the registry) and fans a per-host
// consolidation through the concurrency semaphore. It never blocks the host RPCs:
// the semaphore caps aggregate LLM pressure and each job runs in its own
// goroutine tied to cronWG (so shutdown drains them).
func (s *Server) consolidationLoop(ctx context.Context, cfg CronConfig) {
	defer s.cronWG.Done()
	t := time.NewTicker(cfg.ConsolidateEvery)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			s.consolidateSweep(ctx, cfg)
		}
	}
}

// consolidateSweep runs one tick: enumerate active hosts and dispatch each
// through the semaphore. Each per-host job is its own goroutine joined to cronWG.
func (s *Server) consolidateSweep(ctx context.Context, cfg CronConfig) {
	since := time.Now().Add(-cfg.ActiveWindow).Unix()
	hosts, err := s.ltm.ActiveHostsWithObservations(ctx, since, cfg.MaxHostsPerSweep)
	if err != nil {
		s.log.Warn("crons: active-hosts query failed", "err", err)
		return
	}
	if len(hosts) == 0 {
		return // empty sweep → no spurious LLM calls
	}
	for _, hostID := range hosts {
		select {
		case <-ctx.Done():
			return
		case s.cronSem <- struct{}{}: // acquire (bounds in-flight LLM jobs)
		}
		s.cronWG.Add(1)
		go func(id string) {
			defer s.cronWG.Done()
			defer func() { <-s.cronSem }() // release
			// Bound each host job: a slow Postgres query or LLM call must not hold a
			// DB connection / LLM slot indefinitely and starve the host-facing RPCs.
			jobCtx, cancel := context.WithTimeout(ctx, cronJobTimeout)
			defer cancel()
			if err := s.consolidateHost(jobCtx, id, cfg); err != nil {
				s.log.Warn("crons: consolidate host failed", "host_id", id, "err", err)
			}
		}(hostID)
	}
}

// consolidationCursor is the per-host KV cursor: the last consolidated
// occurred_at + a fail counter for the poison-batch guard.
type consolidationCursor struct {
	LastUnix  int64 `json:"last_occurred_at_unix"`
	FailCount int   `json:"fail_count"`
}

const cronConsolidateCursorKey = "cron:consolidate:cursor"
const cronEscalateQueueKey = "cron:escalate:queue"

// maxFailBeforeSkip is how many parse failures on the same batch we tolerate
// before advancing past it (so one poison batch can't wedge a host's cursor).
const maxFailBeforeSkip = 3

// cronJobTimeout bounds a single host's consolidation job (DB + one LLM batch),
// so a stuck query/call can't hold a DB connection or LLM slot and starve the
// latency-sensitive host RPCs (Act/Chat/Decide/ExtractDialog).
const cronJobTimeout = 60 * time.Second

// loadCursor reads a host's consolidation cursor (zero value on a miss).
func (s *Server) loadCursor(ctx context.Context, hostID string) consolidationCursor {
	var cur consolidationCursor
	if v, ok, err := s.ltm.GetKV(ctx, hostID, cronConsolidateCursorKey); err == nil && ok {
		if uerr := json.Unmarshal(v, &cur); uerr != nil {
			// Corrupt cursor: surface it + reset (re-reads from the beginning, but the
			// fold is idempotent by (subject,kind,claim), so it self-heals).
			s.log.Warn("crons: cursor unmarshal failed; resetting", "host_id", hostID, "err", uerr)
			cur = consolidationCursor{}
		}
	}
	return cur
}

// saveCursor persists a host's consolidation cursor.
func (s *Server) saveCursor(ctx context.Context, hostID string, cur consolidationCursor) {
	if v, err := json.Marshal(cur); err == nil {
		if err := s.ltm.PutKV(ctx, hostID, cronConsolidateCursorKey, v); err != nil {
			s.log.Warn("crons: cursor save failed", "host_id", hostID, "err", err)
		}
	}
}

// firehoseKinds are the observation kinds the reactive tier (Phase-2b
// ExtractDialog) did NOT already extract live — the bulk the cron sweeps. An
// empty set means "take everything" (defensive).
var firehoseKinds = map[string]bool{
	"npc_dialog":  true,
	"player_chat": true,
	"server_msg":  true,
	"claim_heard": true,
}

// consolidateHost folds one host's NEW observation batch into its durable
// knowledge ledger via a single batched Haiku call, then runs Tier-0 GC and
// advances the cursor. Idempotent: re-running over the same observations folds
// the same claims (Note dedups by (subject,kind,claim)) and the cursor only
// advances after the fold persists.
func (s *Server) consolidateHost(ctx context.Context, hostID string, cfg CronConfig) error {
	cur := s.loadCursor(ctx, hostID)
	obs, err := s.ltm.Observations(ctx, hostID, cur.LastUnix, cfg.BatchSize)
	if err != nil {
		return fmt.Errorf("read observations: %w", err)
	}
	if len(obs) == 0 {
		return nil
	}

	// Filter to the firehose the reactive tier skipped, but never starve: if the
	// batch is entirely non-firehose kinds, fall back to processing it all so the
	// cursor still advances (otherwise it would wedge forever on those rows).
	batch := make([]observedRow, 0, len(obs))
	for _, o := range obs {
		if firehoseKinds[o.Kind] {
			batch = append(batch, o)
		}
	}
	maxAt := obs[len(obs)-1].At // cursor target = the newest occurred_at of the FULL read

	// Snapshot the persona prose WITHOUT holding s.mu during the LLM call.
	persona := ""
	if e, ok := s.lookup(hostID); ok {
		persona = e.prose
	}

	if len(batch) > 0 {
		claims, perr := s.extractClaims(ctx, hostID, persona, batch)
		if perr != nil {
			// LLM/parse failure: bump fail_count, DO NOT advance the cursor (retry
			// next tick) until we hit the poison-batch ceiling.
			cur.FailCount++
			if cur.FailCount >= maxFailBeforeSkip {
				s.log.Warn("crons: poison batch, advancing past it", "host_id", hostID, "fails", cur.FailCount)
				cur.LastUnix = maxAt
				cur.FailCount = 0
				s.saveCursor(ctx, hostID, cur)
			} else {
				s.saveCursor(ctx, hostID, cur)
			}
			return perr
		}

		// Idempotent fold: load the current ledger, apply each claim via the
		// confidence-bearing Note (creates entry + sets kind), bump familiarity,
		// tag sentiment. Then GC, then persist the whole snapshot.
		led := knowledge.NewLedger()
		if cur, kerr := s.ltm.Knowledge(ctx, hostID); kerr == nil {
			led.Import(entriesToKnowledge(cur))
		}
		for _, c := range claims {
			subj := strings.TrimSpace(c.Subject)
			if subj == "" || strings.TrimSpace(c.Claim) == "" {
				continue
			}
			prov := normalizeProvenance(c.Provenance)
			conf := c.Confidence
			if conf <= 0 || conf > 1 {
				conf = 0.5
			}
			led.Note(subj, strings.TrimSpace(c.Kind), strings.TrimSpace(c.Claim), prov, conf)
			led.Seen(subj, strings.TrimSpace(c.Kind))
			if sent := strings.TrimSpace(c.Sentiment); sent != "" {
				led.Tag(subj, "sentiment:"+sent)
			}
		}

		rows := gcKnowledge(led.Export(), cfg, time.Now())
		if _, werr := s.ltm.SyncKnowledge(ctx, hostID, knowledgeToEntries(rows)); werr != nil {
			// Persist failed: do NOT advance the cursor — reprocess next tick.
			return fmt.Errorf("sync knowledge: %w", werr)
		}

		// 4a: RECORD the escalation flags to a durable queue. NOTHING reads it yet
		// — the 4b insight cron drains it (Tier-2 Sonnet/Opus over this slice).
		s.recordEscalations(ctx, hostID, claims)

		s.log.Info("crons: consolidated", "host_id", hostID,
			"observations", len(batch), "claims", len(claims), "subjects", len(rows))
	}

	// Advance the cursor only AFTER a successful fold+persist (or an empty batch).
	cur.LastUnix = maxAt
	cur.FailCount = 0
	s.saveCursor(ctx, hostID, cur)
	return nil
}

// recordEscalations appends the flagged claims to the host's durable escalate
// queue (4a record-only). The 4b insight cron consumes it.
func (s *Server) recordEscalations(ctx context.Context, hostID string, claims []consolidatedClaim) {
	type qItem struct {
		Subject string `json:"subject"`
		Claim   string `json:"claim"`
		Reason  string `json:"reason"`
		AtUnix  int64  `json:"at_unix"`
	}
	var flagged []qItem
	for _, c := range claims {
		if !c.Escalate {
			continue
		}
		flagged = append(flagged, qItem{Subject: c.Subject, Claim: c.Claim, Reason: c.EscalateReason, AtUnix: time.Now().Unix()})
	}
	if len(flagged) == 0 {
		return
	}
	var existing []qItem
	if v, ok, err := s.ltm.GetKV(ctx, hostID, cronEscalateQueueKey); err == nil && ok {
		if uerr := json.Unmarshal(v, &existing); uerr != nil {
			// Don't silently drop a corrupt queue's prior entries without a trace.
			s.log.Warn("crons: escalate-queue unmarshal failed; starting fresh", "host_id", hostID, "err", uerr)
			existing = nil
		}
	}
	existing = append(existing, flagged...)
	// Bound the queue so a never-drained 4a queue can't grow unbounded.
	const maxQueue = 200
	if len(existing) > maxQueue {
		existing = existing[len(existing)-maxQueue:]
	}
	if v, err := json.Marshal(existing); err == nil {
		if perr := s.ltm.PutKV(ctx, hostID, cronEscalateQueueKey, v); perr != nil {
			s.log.Warn("crons: escalate-queue write failed", "host_id", hostID, "err", perr)
		}
	}
}

// --- Tier-1 Haiku claim extraction (the bulk LLM call) ----------------------

// consolidatedClaim is one durable claim the bulk Haiku pass extracted from a
// batch of firehose observations, plus the cheap inline triage (sentiment +
// escalate flag) — no extra LLM call for routing (the §3.6 "escalation is part
// of the cheap pass" stance).
type consolidatedClaim struct {
	Subject        string  `json:"subject"`
	Kind           string  `json:"kind"`
	Claim          string  `json:"claim"`
	Confidence     float64 `json:"confidence"`
	Provenance     string  `json:"provenance"`
	Sentiment      string  `json:"sentiment"`
	Escalate       bool    `json:"escalate"`
	EscalateReason string  `json:"escalate_reason"`
}

// consolidateSystem is the STABLE analyzer prefix — flat analyzer voice, JSON
// only, identical across every host and sweep so Anthropic prompt-caches it
// (block[0] Cache:true). Mirrors extractDialogSystem's discipline but for the
// deferred batch (cross-line dedup + triage).
const consolidateSystem = `You are a memory-consolidation analyzer for a RuneScape Classic player agent. You read a BATCH of that agent's recently-overheard/observed lines (NPC dialog, player chat, server messages) and distil the DURABLE world-knowledge worth remembering. You are NOT the character — do not roleplay. Be literal.

Return ONLY JSON:
{"claims":[{"subject":"<what it's about>","kind":"<item|shop|npc|place|quest|concept|recipe|mechanic|...>","claim":"<one concrete durable fact>","confidence":0.0-1.0,"provenance":"server-msg|npc-told|player-told|implied","sentiment":"<optional: positive|negative|neutral toward the subject, omit if none>","escalate":<true ONLY if this contradicts likely existing knowledge, answers an open question, or needs cross-entity reasoning>,"escalate_reason":"<one short phrase, only if escalate=true>"}]}

Rules:
- DEDUP across the batch: if several lines assert the same fact, emit it ONCE (highest confidence).
- Extract only DURABLE facts (where to buy/find things, prices, dangers, who does what, how to do something, quest prereqs). Skip greetings, insults, small talk, and pure noise — emit no claim for those.
- confidence reflects how firmly the fact was asserted (a hedge is low; a flat statement is high). server messages are authoritative (high); player hearsay is weaker.
- escalate is RARE — set it only for genuine contradictions or facts that clearly need deeper cross-entity reasoning; most claims do NOT escalate.
- If the batch has nothing durable, return {"claims":[]}.

Reply with JSON only.`

// extractClaims runs the single batched Tier-1 (Haiku) consolidation call over a
// host's observation batch. The stable analyzer prompt is the cached prefix; the
// per-host persona + existing-beliefs snippet is the uncached block; the numbered
// batch is the user turn. Returns the parsed claims (never nil on success).
func (s *Server) extractClaims(ctx context.Context, hostID, persona string, batch []observedRow) ([]consolidatedClaim, error) {
	tier := s.consolidateLLM()
	if tier == nil {
		return nil, fmt.Errorf("no consolidation LLM configured")
	}

	// Existing-beliefs snippet for dedup context (top by recency, bounded).
	existing := ""
	if cur, err := s.ltm.Knowledge(ctx, hostID); err == nil && len(cur) > 0 {
		var b strings.Builder
		for i, e := range cur {
			if i >= 12 || b.Len() > 500 {
				break
			}
			claim := ""
			if len(e.GetBeliefs()) > 0 {
				claim = e.GetBeliefs()[0].GetClaim()
			}
			fmt.Fprintf(&b, "- %s: %s\n", e.GetSubject(), claim)
		}
		existing = b.String()
	}

	var ctxBlock strings.Builder
	if p := strings.TrimSpace(persona); p != "" {
		ctxBlock.WriteString("# YOUR CHARACTER\n" + p + "\n\n")
	}
	if existing != "" {
		ctxBlock.WriteString("# WHAT YOU ALREADY KNOW (dedup against these):\n" + existing)
	}

	blocks := []llm.SystemBlock{
		{Text: consolidateSystem, Cache: true}, // stable → prompt-cached
		{Text: ctxBlock.String(), Cache: false},
	}

	var user strings.Builder
	user.WriteString("Observations (chronological, oldest first):\n")
	for i, o := range batch {
		fmt.Fprintf(&user, "%d. [%s] %s (sal=%.2f): %s\n", i+1, o.Kind, o.Subject, o.Salience, o.Body)
	}
	user.WriteString("\nExtract durable claims; dedup across the batch; flag contradictions/open-question hits with escalate=true. Return JSON only.")

	raw, err := tier.CompleteSystem(ctx, blocks, user.String(), 2000)
	if err != nil {
		return nil, err
	}
	return parseConsolidatedClaims(raw)
}

// parseConsolidatedClaims maps the bulk pass's JSON to []consolidatedClaim. On a
// parse failure it returns an error (the cron treats this as a retryable batch,
// not a silent drop) — distinct from a well-formed empty {"claims":[]} which
// returns an empty (non-nil) slice and advances the cursor.
func parseConsolidatedClaims(raw string) ([]consolidatedClaim, error) {
	js := extractJSON(raw)
	if js == "" {
		return nil, fmt.Errorf("no JSON in consolidation response")
	}
	var v struct {
		Claims []consolidatedClaim `json:"claims"`
	}
	if err := json.Unmarshal([]byte(js), &v); err != nil {
		return nil, fmt.Errorf("consolidation JSON: %w", err)
	}
	out := make([]consolidatedClaim, 0, len(v.Claims))
	for _, c := range v.Claims {
		if strings.TrimSpace(c.Claim) == "" {
			continue // a claim with no content is noise
		}
		out = append(out, c)
	}
	return out, nil
}

// normalizeProvenance maps the LLM's provenance label to a cognition/knowledge
// Prov* constant. The cron's claims come from non-live (deferred) observations,
// so the strongest source it can assert for player/NPC speech is hearsay/observed
// — the host's live reactive tier already wrote the authoritative ones.
func normalizeProvenance(p string) string {
	switch strings.ToLower(strings.TrimSpace(p)) {
	case "server-msg", "server", "system":
		return knowledge.ProvSystem
	case "observed", "did", "saw":
		return knowledge.ProvObserved
	case "implied", "inferred", "deduced":
		return knowledge.ProvDeduced
	default:
		return knowledge.ProvHearsay
	}
}

// --- ledger <-> proto converters (shared with the host push path) -----------

// entriesToKnowledge converts wire KnowledgeEntry rows into ledger import rows.
func entriesToKnowledge(entries []*mesapb.KnowledgeEntry) []knowledge.Entry {
	out := make([]knowledge.Entry, 0, len(entries))
	for _, e := range entries {
		row := knowledge.Entry{
			Subject:    e.GetSubject(),
			Kind:       e.GetKind(),
			Encounters: int(e.GetEncounters()),
			LastSeen:   e.GetLastSeenUnix(),
			Tags:       e.GetTags(),
		}
		for _, b := range e.GetBeliefs() {
			// Guard a malformed/legacy wire row: the Beta model requires α,β > 0
			// (Confidence = α/(α+β)); coerce a bad value to the uninformed prior.
			alpha, beta := b.GetAlpha(), b.GetBeta()
			if alpha <= 0 {
				alpha = 1
			}
			if beta <= 0 {
				beta = 1
			}
			row.Beliefs = append(row.Beliefs, knowledge.Belief{
				Claim: b.GetClaim(), Provenance: b.GetProvenance(),
				Alpha: alpha, Beta: beta, At: b.GetAtUnix(),
			})
		}
		out = append(out, row)
	}
	return out
}

// knowledgeToEntries converts ledger snapshot rows into wire KnowledgeEntry.
func knowledgeToEntries(rows []knowledge.Entry) []*mesapb.KnowledgeEntry {
	out := make([]*mesapb.KnowledgeEntry, 0, len(rows))
	for _, r := range rows {
		e := &mesapb.KnowledgeEntry{
			Subject:      r.Subject,
			Kind:         r.Kind,
			Encounters:   int32(r.Encounters),
			LastSeenUnix: r.LastSeen,
			Tags:         r.Tags,
		}
		for _, b := range r.Beliefs {
			e.Beliefs = append(e.Beliefs, &mesapb.KnowledgeBelief{
				Claim: b.Claim, Provenance: b.Provenance,
				Alpha: b.Alpha, Beta: b.Beta, AtUnix: b.At,
			})
		}
		out = append(out, e)
	}
	return out
}

// --- Tier-0 GC / cull (NO LLM) ----------------------------------------------

// gcKnowledge bounds a host's ledger at zero LLM cost (the "GC is Tier-0"
// invariant): it prunes stale + low-confidence + low-encounter subjects (TTL
// decay) and then caps the count, evicting the lowest encounters×recency first.
// `system`-provenance subjects (game-authoritative) are pinned and never evicted.
// Pure arithmetic on the snapshot before it is written back.
func gcKnowledge(rows []knowledge.Entry, cfg CronConfig, now time.Time) []knowledge.Entry {
	cfg = cfg.sane()
	ttlCut := now.Add(-cfg.KnowledgeTTL).Unix()

	pinned := func(e knowledge.Entry) bool {
		for _, b := range e.Beliefs {
			if b.Provenance == knowledge.ProvSystem {
				return true
			}
		}
		for _, t := range e.Tags {
			if strings.HasPrefix(t, "open-question") || t == "pinned" {
				return true
			}
		}
		return false
	}

	// 1) TTL decay prune: stale AND weak AND barely-encountered (and not pinned).
	kept := make([]knowledge.Entry, 0, len(rows))
	for _, e := range rows {
		if !pinned(e) && e.LastSeen < ttlCut && strongestConfidence(e) < 0.2 && e.Encounters <= 1 {
			continue // pruned
		}
		kept = append(kept, e)
	}

	// 2) Size bound: cap subjects; evict lowest encounters×recency, pins survive.
	if len(kept) <= cfg.MaxSubjects {
		return kept
	}
	type scored struct {
		e     knowledge.Entry
		score float64
		pin   bool
	}
	sc := make([]scored, len(kept))
	for i, e := range kept {
		// recency in [0,1]: newer = higher. encounters multiply.
		ageDays := float64(now.Unix()-e.LastSeen) / 86400.0
		if ageDays < 0 {
			ageDays = 0
		}
		recency := 1.0 / (1.0 + ageDays)
		sc[i] = scored{e: e, score: float64(e.Encounters+1) * recency, pin: pinned(e)}
	}
	// Highest score first; pins always sort to the front so they're retained.
	sort.SliceStable(sc, func(i, j int) bool {
		if sc[i].pin != sc[j].pin {
			return sc[i].pin
		}
		return sc[i].score > sc[j].score
	})
	out := make([]knowledge.Entry, 0, cfg.MaxSubjects)
	for i := 0; i < len(sc) && len(out) < cfg.MaxSubjects; i++ {
		out = append(out, sc[i].e)
	}
	return out
}

func strongestConfidence(e knowledge.Entry) float64 {
	best := 0.0
	for _, b := range e.Beliefs {
		if c := b.Confidence(); c > best {
			best = c
		}
	}
	return best
}
