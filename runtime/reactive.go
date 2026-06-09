package runtime

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/cognition/knowledge"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
)

// reactive.go is the host's SPEED-2 reactive tier (docs/world-knowledge-and-
// learning.md §3.5): the firehose (perception.go's emitObservation) stays the
// slow speed-3 ambient stream, but when a signal TRIGGERS it must be acted on in
// near-real-time (<10s — a batch cron 30 min later is useless for a trade offer
// or a warning). The flow:
//
//  1. TRIGGER DETECTOR (triggerHit): deterministic, NO LLM — keyword ladder ×
//     salience × directed-at-me × goal-touch. Most signals do NOT trip it; only a
//     hit escalates out of the firehose.
//  2. PER-SPEAKER CONVERSATION WINDOW (reactiveState): an always-on, BOUNDED,
//     per-speaker rolling buffer (the last few lines / ~minute each recent speaker
//     said, INCLUDING the host's own lines so a Q&A pairs up). A trigger LATCHES
//     the speaker: subsequent lines bypass the gate (already engaged); the latch
//     TTL refreshes on each new line and decays when quiet.
//  3. FAST EXTRACT (spawnExtract → mesa ExtractDialog): the windowed exchange +
//     light context → structured {claims, intent}. Haiku/Sonnet, never Opus.
//  4. LEDGER WRITEBACK (writebackClaims): claims land in h.knowledge in <10s so
//     the NEXT decision reasons over them. Provenance from the speaker ROLE.
//  5. CONDUCTOR INTERRUPT (maybeInterrupt): an urgent intent raises the existing
//     conductor detour so the host re-plans NOW instead of next Act turn.
//
// The host stays LIGHT: the trigger is deterministic, the LLM is a mesa RPC, and
// extraction runs in a bounded fire-and-forget goroutine — it never blocks the
// bus or the turn loop. Both speed tiers cap their concurrent spawns with an
// inflight semaphore (this tier's reactive.inflight / reactiveMaxInflight; the
// speed-3 firehose's h.obsInflight / observationMaxInflight, M15). Everything is
// bounded (buffer, latches, window size, inflight RPCs). Frozen under analysis-mode.

const (
	reactiveBufLines    = 8                // lines retained per speaker (incl. the host's own)
	reactiveLookback    = 90 * time.Second // prune lines older than this
	reactiveMaxSpeakers = 24               // cap distinct windows in RAM (LRU-evict the oldest non-latched)
	reactiveMaxLatches  = 4                // cap CONCURRENT latched speakers
	reactiveLatchTTL    = 20 * time.Second // latch lifetime; REFRESHES on each new line, decays when quiet
	reactiveMaxInflight = 3                // cap concurrent extraction RPCs (semaphore)
	reactiveTimeout     = 8 * time.Second  // <10s extraction budget
	reactiveDirectedSal = 0.6              // player salience at/above which a line counts as directed (PM/whisper)
	reactiveServerSal   = 0.55             // server salience at/above which a system line counts as a directive

	// closeQConf is the confidence floor a freshly-written claim must meet to
	// auto-close a matching open question (the ask-drive loop closer). NPC/server
	// answers write at 0.85 (authoritative → clears it); flimsy player hearsay at
	// 0.5 does NOT auto-close (it sits in the ledger for a cron to judge later).
	closeQConf = 0.6
)

// dialogLine is one captured line in a speaker's window. role is npc|player|
// server|self; self lines are the host's own replies fanned in for Q&A pairing.
type dialogLine struct {
	at      time.Time
	role    string
	speaker string
	text    string
}

// speakerWindow is one speaker's bounded rolling buffer + latch state.
type speakerWindow struct {
	lines    []dialogLine
	latched  bool
	latchTTL time.Time // when the latch expires (refreshed on each new line)
	lastSeen time.Time
}

// reactiveState holds the per-speaker windows + latches. All methods are mutex-
// guarded and O(1)/O(buf). RAM-only, bounded, never persisted.
type reactiveState struct {
	mu       sync.Mutex
	windows  map[string]*speakerWindow
	latches  int           // count of currently-latched speakers (≤ reactiveMaxLatches)
	inflight chan struct{} // semaphore bounding concurrent extraction RPCs

	// directSelf, when non-empty, routes the NEXT self-line (the host's own
	// outbound chat through reactiveObserveSelf) into ONLY this speaker's window
	// instead of broadcasting to every latched window — set by the ASK path right
	// before it emits a directed question so the question pairs only with its own
	// conversation (L7). Consumed once (cleared on use).
	directSelf string
}

// newReactiveState builds an empty reactive state with the inflight semaphore.
func newReactiveState() *reactiveState {
	return &reactiveState{
		windows:  make(map[string]*speakerWindow),
		inflight: make(chan struct{}, reactiveMaxInflight),
	}
}

// pushLine appends a line to the speaker's window (creating it, evicting the
// oldest non-latched window at the speaker cap), prunes lines older than the
// lookback, trims to the buffer cap, and refreshes the latch TTL if latched.
// Returns whether the speaker is currently latched (so the caller routes the
// line straight to extraction, bypassing the trigger gate). Mutex-guarded.
func (r *reactiveState) pushLine(key string, dl dialogLine) (latched bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := r.windows[key]
	if w == nil {
		r.evictIfFullLocked()
		if len(r.windows) >= reactiveMaxSpeakers {
			return false // still at cap (eviction found only latched windows) — drop, never overflow
		}
		w = &speakerWindow{}
		r.windows[key] = w
	}
	w.lines = append(w.lines, dl)
	w.lastSeen = dl.at
	r.pruneLocked(w, dl.at)
	if w.latched {
		if w.latchTTL.Before(dl.at) {
			// The latch already expired; drop it (decay) — this line did not refresh
			// a live latch, it lapsed. The caller's trigger gate decides re-latch.
			w.latched = false
			r.latches--
		} else {
			w.latchTTL = dl.at.Add(reactiveLatchTTL) // refresh on activity
			return true
		}
	}
	return false
}

// pruneLocked drops lines older than the lookback and trims the window to the
// buffer cap (keeping the newest). Caller holds the mutex.
func (r *reactiveState) pruneLocked(w *speakerWindow, now time.Time) {
	cutoff := now.Add(-reactiveLookback)
	keep := w.lines[:0]
	for _, l := range w.lines {
		if l.at.After(cutoff) {
			keep = append(keep, l)
		}
	}
	w.lines = keep
	if len(w.lines) > reactiveBufLines {
		w.lines = w.lines[len(w.lines)-reactiveBufLines:]
	}
}

// evictIfFullLocked makes room for a new speaker window at the speaker cap by
// dropping the oldest NON-LATCHED window (latched speakers are never evicted —
// they're actively engaged). Caller holds the mutex.
func (r *reactiveState) evictIfFullLocked() {
	if len(r.windows) < reactiveMaxSpeakers {
		return
	}
	var oldestKey string
	var oldest time.Time
	for k, w := range r.windows {
		if w.latched {
			continue
		}
		if oldestKey == "" || w.lastSeen.Before(oldest) {
			oldestKey, oldest = k, w.lastSeen
		}
	}
	if oldestKey != "" {
		delete(r.windows, oldestKey)
	}
}

// tryLatch latches a speaker (so subsequent lines bypass the trigger gate). It
// latches only if under the concurrent-latch cap; if the speaker is already
// latched it refreshes the TTL. The window is CREATED when absent (evicting the
// oldest non-latched window at the speaker cap) so the proactive ASK path can
// latch a cold target it has never heard before and fan its own question into the
// fresh window BEFORE the answer arrives (H6). Returns whether the speaker is
// latched after the call (false ⇒ at the latch/speaker cap, no slot). Mutex-guarded.
func (r *reactiveState) tryLatch(key string, now time.Time) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := r.windows[key]
	if w == nil {
		r.evictIfFullLocked()
		if len(r.windows) >= reactiveMaxSpeakers {
			return false // still at cap (eviction found only latched windows) — no slot
		}
		w = &speakerWindow{lastSeen: now}
		r.windows[key] = w
	}
	if w.latched {
		w.latchTTL = now.Add(reactiveLatchTTL)
		return true
	}
	if r.latches >= reactiveMaxLatches {
		return false // at the concurrent-latch cap — stay ambient this round
	}
	w.latched = true
	w.latchTTL = now.Add(reactiveLatchTTL)
	r.latches++
	return true
}

// unlatch drops a speaker's latch if it holds one (used to undo a latch the
// caller acquired but then could not commit, e.g. an ASK whose send failed —
// avoids an orphaned latch). No-op when the window is absent or not latched.
// Mutex-guarded.
func (r *reactiveState) unlatch(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if w := r.windows[key]; w != nil && w.latched {
		w.latched = false
		r.latches--
	}
}

// snapshot renders the speaker's window as "Speaker: text" / "Me: text" lines,
// oldest first — the exchange shipped to the extractor. Mutex-guarded.
func (r *reactiveState) snapshot(key string) []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := r.windows[key]
	if w == nil {
		return nil
	}
	out := make([]string, 0, len(w.lines))
	for _, l := range w.lines {
		who := l.speaker
		if l.role == "self" {
			who = "Me"
		}
		out = append(out, who+": "+l.text)
	}
	return out
}

// appendToWindow fans a line (the host's own reply/question) into ONE latched
// speaker's window — the speaker the host is actually replying to — so the next
// extract for THAT speaker sees the Q→A pair without cross-contaminating other
// in-flight conversations (L7). Only the target's latch TTL is refreshed. No-op
// when the window is absent or not latched. Self lines never trigger or latch.
// Returns whether the line was appended. Mutex-guarded.
func (r *reactiveState) appendToWindow(key string, dl dialogLine) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	w := r.windows[key]
	if w == nil || !w.latched {
		return false
	}
	w.lines = append(w.lines, dl)
	w.lastSeen = dl.at
	w.latchTTL = dl.at.Add(reactiveLatchTTL) // the host's own reply keeps THIS conversation hot (§3.5: TTL refreshes on every line)
	r.pruneLocked(w, dl.at)
	return true
}

// appendToLatched fans a line (the host's own reply) into EVERY currently-latched
// window — the fallback for an UNTARGETED public utterance with no single
// addressee (a broadcast genuinely heard by everyone in local-chat range). The
// targeted ASK/reply path uses appendToWindow instead, so it does not cross-
// contaminate other conversations (L7). Self lines never trigger or latch.
// Mutex-guarded.
func (r *reactiveState) appendToLatched(dl dialogLine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, w := range r.windows {
		if !w.latched {
			continue
		}
		w.lines = append(w.lines, dl)
		w.lastSeen = dl.at
		w.latchTTL = dl.at.Add(reactiveLatchTTL) // the host's own reply keeps the conversation hot (§3.5: TTL refreshes on every line)
		r.pruneLocked(w, dl.at)
	}
}

// directSelfTo arms the one-shot self-line routing so the NEXT self-line goes
// only into `key`'s window (the ASK path's directed question), not the broadcast
// fan (L7). Overwrites any prior unconsumed hint. Mutex-guarded.
func (r *reactiveState) directSelfTo(key string) {
	r.mu.Lock()
	r.directSelf = key
	r.mu.Unlock()
}

// clearDirectSelf drops an armed-but-unconsumed self-line routing hint (e.g. the
// ASK send failed before the self-line was emitted). Mutex-guarded.
func (r *reactiveState) clearDirectSelf() {
	r.mu.Lock()
	r.directSelf = ""
	r.mu.Unlock()
}

// takeDirectSelf returns and clears the one-shot directed self-target ("" when
// none is armed). Mutex-guarded.
func (r *reactiveState) takeDirectSelf() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	k := r.directSelf
	r.directSelf = ""
	return k
}

// gcLatches drops expired latches and evicts windows that have gone fully idle
// (no lines after the lookback). Called from the limbic flush ticker — no new
// long-running loop. Mutex-guarded.
func (r *reactiveState) gcLatches(now time.Time) {
	r.mu.Lock()
	defer r.mu.Unlock()
	cutoff := now.Add(-reactiveLookback)
	for k, w := range r.windows {
		if w.latched && w.latchTTL.Before(now) {
			w.latched = false
			r.latches--
		}
		r.pruneLocked(w, now)
		if !w.latched && (len(w.lines) == 0 || w.lastSeen.Before(cutoff)) {
			delete(r.windows, k)
		}
	}
}

// --- host wiring: observe → trigger → extract → writeback → interrupt --------

// reactiveObserve is the speed-2 entry point, called from perceiveDialog after
// the speed-3 ambient emit. It feeds the line into the speaker's window and, if
// the speaker is latched (already engaged) or the deterministic trigger fires,
// spawns a bounded extraction. No-ops on a nil reactive state (REPL/test) and
// under analysis-mode (the operator override freezes all learning I/O).
func (h *Host) reactiveObserve(kind, subject, text string, salience float64) {
	if h.reactive == nil || h.AnalysisActive() {
		return
	}
	role := roleFor(kind)
	key := normalizeSpeaker(subject)
	now := time.Now()
	if h.reactive.pushLine(key, dialogLine{at: now, role: role, speaker: subject, text: text}) {
		h.spawnExtract(subject, role) // already latched → bypass the gate, keep extracting
		return
	}
	if h.triggerHit(role, subject, text, salience) {
		if h.reactive.tryLatch(key, now) {
			h.spawnExtract(subject, role)
		}
	}
}

// reactiveObserveSelf captures the host's OWN outbound chat at the send seam
// (h.Say) — the inbound echo is filtered as self, so the window would otherwise
// never see the host's replies and a Q&A could not pair up. Self lines never
// trigger/latch; they fan into every currently-latched window. Cheap no-op when
// nobody is latched.
func (h *Host) reactiveObserveSelf(text string) {
	if h.reactive == nil {
		return
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return
	}
	dl := dialogLine{at: time.Now(), role: "self", speaker: h.opts.Username, text: text}
	// A directed self-line (the ASK path armed directSelfTo before emitting) fans
	// into ONLY its target's window, never the broadcast (L7). Consume the hint
	// even when nobody is latched so a stale hint can't leak into a later line.
	if key := h.reactive.takeDirectSelf(); key != "" {
		h.reactive.appendToWindow(key, dl)
		return
	}
	if h.reactive.latchCount() == 0 {
		return
	}
	h.reactive.appendToLatched(dl) // untargeted public utterance — heard by all latched conversations
}

// latchCount returns the current number of latched speakers (lock-guarded), so
// the self-line seam can cheaply skip the fan-in when nobody is engaged.
func (r *reactiveState) latchCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.latches
}

// triggerHit is the Tier-0 deterministic trigger detector — NO LLM. It is an
// OR-of-strong-axes (the design wants reactivity, not a rare conjunction): a
// keyword-ladder hit, a goal-touch (the line overlaps the active goal or an open
// question), or a directed signal (a salient PM/whisper, the host's own name, or
// a salient system directive). Cost is O(ladder + open-questions), both tiny —
// most lines hit none and stay purely ambient (the sparse-trigger invariant).
func (h *Host) triggerHit(role, subject, text string, salience float64) bool {
	low := strings.ToLower(text)
	lowSubj := strings.ToLower(subject)

	// 1. keyword ladder — a word/person that catches the host's attention.
	// Read through keywordRungs() (atomic load) — the bootstrap SetKeywordLadder
	// write lands after this reader goroutine is live (H12).
	for _, rung := range h.keywordRungs() {
		kw := strings.ToLower(strings.TrimSpace(rung.Keyword))
		if kw == "" {
			continue
		}
		if strings.Contains(low, kw) || strings.Contains(lowSubj, kw) {
			return true
		}
	}

	// 2. goal-touch — the line overlaps the active goal or an open question.
	if goalTouch(low, h.LiveGoal()) {
		return true
	}
	if h.goalGraph != nil {
		for _, q := range h.goalGraph.OpenQuestions() {
			if goalTouch(low, q.Label) {
				return true
			}
		}
	}

	// 3. directed-at-me — a salient PM/whisper, my name in the line, or a salient
	// system directive.
	switch role {
	case "player":
		if salience >= reactiveDirectedSal {
			return true
		}
		if u := strings.ToLower(strings.TrimSpace(h.opts.Username)); u != "" && strings.Contains(low, u) {
			return true
		}
	case "server":
		if salience >= reactiveServerSal {
			return true
		}
	}
	return false
}

// spawnExtract runs ONE reactive extraction in a bounded, fire-and-forget
// goroutine (mirrors emitObservation's spawn): semaphore-capped, time-bounded,
// non-blocking. It reuses the analysis mesa client handle (already mutex-guarded,
// the same client used by AnalysisInterpret). At the inflight cap it DROPS the
// extraction (bounded — a chat line is not worth queueing). The goroutine writes
// claims back to the ledger (<10s) and raises a conductor interrupt on urgency.
func (h *Host) spawnExtract(speaker, role string) {
	if h.reactive == nil || h.AnalysisActive() {
		return
	}
	h.analysis.mu.Lock()
	mc := h.analysis.mc
	h.analysis.mu.Unlock()
	if mc == nil || !mc.Healthy() {
		return // offline → nothing to extract; claims stay in the firehose for crons
	}
	select {
	case h.reactive.inflight <- struct{}{}:
	default:
		return // at the concurrent-RPC cap → DROP (bounded)
	}
	window := h.reactive.snapshot(normalizeSpeaker(speaker))
	if len(window) == 0 {
		<-h.reactive.inflight
		return
	}
	persona := h.personaSnippet
	goal := h.LiveGoal()
	questions := openQuestionLabels(h.goalGraph)
	go func() {
		defer func() { <-h.reactive.inflight }()
		if h.AnalysisActive() {
			return // TOCTOU: analysis engaged after we queued — abort the learning I/O
		}
		ctx, cancel := context.WithTimeout(context.Background(), reactiveTimeout)
		defer cancel()
		ex, err := mc.ExtractDialog(ctx, h.opts.Username, speaker, role, window, persona, goal, questions)
		if err != nil {
			h.log.Debug("reactive: extract failed", "speaker", speaker, "err", err)
			return
		}
		h.writebackClaims(speaker, role, ex.Claims)
		h.maybeInterrupt(speaker, ex.Intent)
	}()
}

// writebackClaims writes extracted claims into the knowledge ledger — the <10s
// knowledge update so the next decision reasons over them. Provenance is derived
// from the speaker ROLE, NOT the LLM's self-report (a player can't claim system
// authority): npc/server dialog is game-authoritative (ProvSystem), a player's
// word is ProvHearsay. Note is thread-safe + idempotent (an identical claim
// reinforces rather than duplicates).
func (h *Host) writebackClaims(speaker, role string, claims []mesaclient.DialogClaim) {
	if h.knowledge == nil {
		return
	}
	for _, c := range claims {
		claim := strings.TrimSpace(c.Claim)
		if claim == "" {
			continue
		}
		prov := provenanceForRole(role)
		conf := effectiveConf(role, c.Confidence)
		subj := strings.TrimSpace(c.Subject)
		if subj == "" {
			subj = speaker
		}
		h.knowledge.Note(subj, strings.TrimSpace(c.Kind), claim, prov, conf)
	}
	// Loop closure: if any freshly-written authoritative claim answers a standing
	// open question, flip it done + un-block its goal — don't-know → ask → learn →
	// the open question closes. Rides the already-analysis-frozen reactive goroutine.
	// The closer is gated on the SAME role-derived confidence/provenance the
	// writeback uses (H7), so an authoritative conf-0 answer closes and a confident-
	// sounding player lie does not.
	h.closeResolvedQuestions(role, claims)
}

// provenanceForRole maps a speaker role to the belief provenance — derived from
// the ROLE, never the LLM's self-report (a player cannot claim system authority):
// npc/server dialog is game-authoritative (ProvSystem), a player's word is
// ProvHearsay. The single source of truth for writeback AND question closure (H7).
func provenanceForRole(role string) string {
	if role == "npc" || role == "server" {
		return knowledge.ProvSystem
	}
	return knowledge.ProvHearsay
}

// effectiveConf is the confidence the host actually trusts a claim at, defaulting
// an out-of-range/omitted model confidence from the speaker ROLE (npc/server →
// 0.85 authoritative, player → 0.5 hearsay) and clamping a valid value into
// [0,1]. Computed ONCE so the ledger write and the question closer can never
// disagree on the same claim (H7/M13): an authoritative answer the model tagged 0
// is written AND closes at 0.85, while a player claim it tagged 0.7 is hearsay-
// floored and cannot masquerade as game-authoritative.
func effectiveConf(role string, raw float64) float64 {
	if raw <= 0 || raw > 1 {
		if provenanceForRole(role) == knowledge.ProvSystem {
			return 0.85
		}
		return 0.5
	}
	return raw
}

// closeResolvedQuestions flips an open question to StatusDone when a freshly-
// written claim ANSWERS it (cheap token-overlap, no LLM) — the ask-drive loop
// closer. The gate is the SAME role-derived confidence/provenance writeback uses
// (H7): only an AUTHORITATIVE claim (role npc/server, ProvSystem) whose effective
// confidence clears closeQConf may auto-close — a player's word (hearsay, even one
// the LLM tagged 0.7) never closes a question as game-authoritative, and an
// authoritative answer the model omitted a confidence for still closes at its
// 0.85 default. A topical-but-non-answering mention does NOT close it (M9):
// closure requires the question's SALIENT topic to appear in the CLAIM itself (not
// merely as the claim's subject) plus enough token overlap (claimAnswers), so
// "don't trust the pickaxe seller" can't resolve "where to buy a pickaxe". It only
// mutates EXISTING nodes
// (flips a question done, un-blocks its blocked goals) — no graph growth, so it
// honors "memory, not solver". Deterministic; rides the frozen reactive path.
func (h *Host) closeResolvedQuestions(role string, claims []mesaclient.DialogClaim) {
	if h.goalGraph == nil {
		return
	}
	if provenanceForRole(role) != knowledge.ProvSystem {
		return // only authoritative (npc/server) claims may auto-close — hearsay sits for a cron
	}
	for _, q := range h.goalGraph.OpenQuestions() {
		low := strings.ToLower(q.Label)
		topic := salientTopic(q.Label) // the question's single most salient word
		for _, c := range claims {
			if effectiveConf(role, c.Confidence) < closeQConf {
				continue
			}
			if !claimAnswers(low, topic, c.Claim) {
				continue
			}
			h.goalGraph.SetStatus(q.ID, goalgraph.StatusDone)
			// Un-block any goal this question was blocking (goal --blocked_by--> q).
			for _, e := range h.goalGraph.In(q.ID, goalgraph.RelBlockedBy) {
				if n, ok := h.goalGraph.Get(e.From); ok && n.Status == goalgraph.StatusBlocked {
					h.goalGraph.SetStatus(e.From, goalgraph.StatusActive)
				}
			}
			h.goalGraph.Tag(q.ID, "resolved-by-ask")
			break
		}
	}
}

// closeQuestionByObservation is the PERCEPTION->closure sibling of
// closeResolvedQuestions (5b): a where-to-buy/where-is question is flipped Done when
// the host OBSERVED an answer with its own eyes (a shop's "sells <topic>" belief at
// ProvObserved clearing closeQConf), reusing the IDENTICAL un-block loop. It is a
// SIBLING, not a bypass: the SAME claimAnswers token gate, so a negative/hearsay/
// deduced write can never close (M9). Called from perceiveShop after the positive
// ProvObserved write. Cursor-INDEPENDENT — an organic in-stock find closes the
// question too. Fact.Confidence is the STRONGEST belief only, so it matches the EXACT
// claim by iterating Beliefs (the observation analogue of the ProvSystem gate).
func (h *Host) closeQuestionByObservation(subject, claim string) {
	if h.goalGraph == nil || h.knowledge == nil || strings.TrimSpace(claim) == "" {
		return
	}
	observed := false
	for _, b := range h.knowledge.Get(subject).Beliefs {
		if b.Claim == claim && b.Provenance == knowledge.ProvObserved && b.Confidence() >= closeQConf {
			observed = true
			break
		}
	}
	if !observed {
		return // hearsay / deduced / sub-floor claim never closes
	}
	for _, q := range h.goalGraph.OpenQuestions() {
		topic := salientTopic(q.Label)
		if !claimAnswers(strings.ToLower(q.Label), topic, claim) {
			continue
		}
		h.goalGraph.SetStatus(q.ID, goalgraph.StatusDone)
		// Un-block loop — identical to closeResolvedQuestions.
		for _, e := range h.goalGraph.In(q.ID, goalgraph.RelBlockedBy) {
			if n, ok := h.goalGraph.Get(e.From); ok && n.Status == goalgraph.StatusBlocked {
				h.goalGraph.SetStatus(e.From, goalgraph.StatusActive)
			}
		}
		h.goalGraph.Tag(q.ID, "resolved-by-observation")
		break
	}
}

// claimAnswers reports whether a claim reads as an ANSWER to a question rather
// than a mere topical mention (M9). The decisive gate is that the question's
// salient topic word appears in the CLAIM text itself, not merely as the claim's
// subject — so an authoritative line whose subject is the topic but which only
// disparages it ("don't trust the pickaxe seller", subject "pickaxe") does not
// close "where to buy a pickaxe". For a richer multi-word question it additionally
// requires both salient tokens to overlap; a question with a single significant
// token (the topic) is answered by that token appearing in the claim. Pure +
// deterministic.
func claimAnswers(lowQuestion, topic, claim string) bool {
	low := strings.ToLower(strings.TrimSpace(claim))
	if low == "" {
		return false
	}
	// The question's salient topic must appear in the CLAIM (the M9 core: a claim
	// whose only link to the question is its subject does not answer it).
	if topic == "" || !strings.Contains(low, topic) {
		return false
	}
	// Require as many overlapping significant tokens as the question carries, up to
	// two — so a single-topic question ("where to buy a <pickaxe>") needs just the
	// topic, while a two-topic question needs both (a lone shared word is too weak).
	need := significantTokenCount(lowQuestion)
	if need > 2 {
		need = 2
	}
	if need < 1 {
		need = 1
	}
	return overlapCount(lowQuestion, low) >= need
}

// significantTokenCount counts the distinct significant (≥4-char, non-stopword)
// tokens in a label — the denominator the closer uses to scale its overlap
// requirement. Pure + deterministic.
func significantTokenCount(label string) int {
	seen := map[string]bool{}
	n := 0
	for _, w := range strings.Fields(label) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) < 4 || stopWord(w) || seen[w] {
			continue
		}
		seen[w] = true
		n++
	}
	return n
}

// overlapCount counts the distinct significant (≥4-char, non-stopword) tokens of
// `target` that appear in `line`. The token-overlap primitive goalTouch uses,
// surfaced as a count so the closer can require a stronger-than-single-word match.
// Pure + deterministic.
func overlapCount(target, line string) int {
	if target == "" || line == "" {
		return 0
	}
	seen := map[string]bool{}
	n := 0
	for _, w := range strings.Fields(target) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) < 4 || stopWord(w) || seen[w] {
			continue
		}
		seen[w] = true
		if strings.Contains(line, w) {
			n++
		}
	}
	return n
}

// maybeInterrupt raises a conductor detour when the speaker's intent is urgent
// (a warning / directive / time-sensitive offer), so the current turn is
// preempted and the host re-plans NOW. normal/low urgency does nothing — the
// claims are already in the ledger and the next Act turn reasons over them. The
// interrupt is abort:false (park→react→resume, like survival): a chat line does
// not trash an in-flight grind. The new "reactive" tier inherits the committed-
// region gate via shouldDetour.
func (h *Host) maybeInterrupt(speaker string, in mesaclient.DialogIntent) {
	switch strings.ToLower(strings.TrimSpace(in.Urgency)) {
	case "immediate", "high":
	default:
		return
	}
	c := h.conductorHandle()
	if c == nil || !c.detours {
		return
	}
	gist := strings.TrimSpace(in.Gist)
	c.signalDetour(detourReq{
		tier:   "reactive",
		reason: fmt.Sprintf("reactive[%s] %s: %s", strings.TrimSpace(in.Kind), speaker, gist),
		intent: reactiveInterruptIntent(speaker, gist, in.Kind),
		abort:  false,
	})
}

// reactiveInterruptIntent is the one-shot detour routine the conductor runs when
// a reactive interrupt fires. Its only job is to NOTE the trigger — parking the
// grind bounces control to the conductor loop, which re-invokes the director with
// fresh world + the just-written ledger claims, so the DIRECTOR re-plans over the
// new knowledge (the heavy reaction is the director's job, not this routine's).
func reactiveInterruptIntent(speaker, gist, kind string) Intent {
	note := strings.TrimSpace(fmt.Sprintf("reactive: %s from %s — %s", kind, speaker, gist))
	// Keep the DSL string literal well-formed regardless of the model's gist:
	// drop backslashes (would escape the closing quote), collapse newlines, requote.
	note = strings.ReplaceAll(note, `\`, " ")
	note = strings.ReplaceAll(note, "\n", " ")
	note = strings.ReplaceAll(note, "\r", " ")
	note = strings.ReplaceAll(note, `"`, `'`)
	src := "runtime \"1.0\"\nroutine reactive_react() {\n    note(\"" + note + "\")\n}"
	return Intent{Label: "detour:reactive", Name: "reactive_react", Source: src, OneShot: true}
}

// --- small deterministic helpers --------------------------------------------

// roleFor maps a perception kind to the reactive role tag.
func roleFor(kind string) string {
	switch kind {
	case "npc_dialog":
		return "npc"
	case "server_msg":
		return "server"
	default: // player_chat (and anything else) is treated as a player line
		return "player"
	}
}

// normalizeSpeaker is the per-speaker window key (case-folded, trimmed) so the
// same speaker always maps to one window.
func normalizeSpeaker(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// openQuestionLabels renders the goal graph's open-question labels (bounded by
// the graph's own dedup) for the extractor's grounding context. nil-safe.
func openQuestionLabels(g *goalgraph.Graph) []string {
	if g == nil {
		return nil
	}
	qs := g.OpenQuestions()
	if len(qs) == 0 {
		return nil
	}
	out := make([]string, 0, len(qs))
	for _, q := range qs {
		if l := strings.TrimSpace(q.Label); l != "" {
			out = append(out, l)
		}
	}
	return out
}

// goalTouch reports whether the line overlaps the target (the active goal or an
// open question) by SIGNIFICANT-WORD containment — a cheap token overlap, not a
// whole-sentence substring. A word is significant if it's long enough to carry
// meaning (skips "the", "a", "to", etc.). Pure + deterministic.
func goalTouch(lowLine, target string) bool {
	target = strings.ToLower(strings.TrimSpace(target))
	if target == "" || lowLine == "" {
		return false
	}
	for _, w := range strings.Fields(target) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) < 4 || stopWord(w) {
			continue
		}
		if strings.Contains(lowLine, w) {
			return true
		}
	}
	return false
}

// stopWord filters out common words that would over-trigger goal-touch (they
// carry no topical signal). Small fixed set — cheap.
func stopWord(w string) bool {
	switch w {
	case "your", "youre", "with", "have", "want", "this", "that", "they", "them",
		"then", "than", "from", "into", "what", "when", "where", "will", "would",
		"there", "here", "some", "just", "like", "make", "made", "been", "being",
		"about", "after", "their", "which", "while", "should", "could", "going":
		return true
	}
	return false
}
