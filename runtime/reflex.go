package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/world"
)

// The social-reflex drive: the speed-2 actuation clock. Started by
// RunHost for mesa-linked hosts, it runs for the life of the host — a
// permanent runtime drive, not bootstrap. The reactive/speech/forage
// gates (reactive.go / speech.go / forage.go) decide; socialReflex
// actuates.

// socialReflex answers players who speak to the host on a cheap, reactive path
// (Game.Chat on the Haiku tier), independent of the Act planning loop. To give
// RICH replies (so "what are you doing?" gets a real answer), it tracks her
// latest published thought (reasoning + perception) off the bus and passes that
// as context — the same self-knowledge the Act loop has, without being the Act
// loop. Ignores her own echoed chat + rate-limits.
func socialReflex(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, username, goal string) {
	ch := host.Bus().Subscribe("*", 256)
	var last time.Time
	var doing, perception string // her most recent reasoning + perceived context
	// Per-speaker reply throttle + conversational-loop breaker (soak retro #12:
	// the bernard/drone6/drone8 "tar brokerage" doom-spiral — each reflex reply
	// socially obligated the next, 36-60% of decisions chat-driven). Suppressed
	// lines still reach memory/the situation (perception.go feeds the reactive
	// engine independently); the host HEARS, it just stops compulsively answering.
	// The hostcfg operator — the host's MENTOR — is exempt (lazy provider: the
	// operator is bound by configureAnalysis AFTER this goroutine starts).
	gate := newReflexGate(host.curiosity.Social, host.AnalysisOperator)
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			switch e := ev.(type) {
			case event.AgentThought:
				// Only a REAL planner turn (Turn>0) or a drive thought that carries
				// its own Perception (ask/forage publish "asked X"/"foraging Y")
				// speaks for "what she's doing": the perception-less Turn==0
				// publishDecision records (cheap-loop, lifecycle, the C-2 confidence
				// calibration line, ...) are telemetry and must not leak into the
				// in-character chat context (taking them would also blank perception).
				if e.Turn > 0 || e.Perception != "" {
					doing, perception = e.Reasoning, e.Perception // what she's currently up to
				}
				// The AgentThought tick is a host-owned proactive clock: try the
				// intent-driven ASK drive here (its own AgentThought is filtered out
				// below — we ignore our own ask echo to avoid re-entrant asking; the
				// per-escalation "confidence" calibration record is excluded too so an
				// Act escalation doesn't tick the drives twice in one turn).
				if e.Trigger != "ask" && e.Trigger != "forage" && e.Trigger != "confidence" {
					tryAsk(ctx, log, host, mc, username)
					// FORAGE drive (5b): runs AFTER tryAsk and shares its global floor,
					// so exactly one of {ask, forage} actuates per gap — forage fires
					// only when no local interlocutor could be asked this tick.
					tryForage(ctx, log, host, mc, username)
				}
			case event.OtherPlayerChat:
				from := reflexPlayerName(host, e.PlayerIndex)
				if strings.EqualFold(from, username) {
					continue // her own chat echoed back — never reply to herself
				}
				// OPERATOR OVERRIDE trigger: "!<username> ANALYSIS" / "!<username>
				// RESUME" from the configured operator toggles analysis mode; while
				// active, every operator line is routed to the directive interpreter.
				// Auth is mandatory (a host-takeover vector otherwise): only an EXACT,
				// non-empty sender name == the configured operator is honored — the
				// "a player" placeholder for an unresolved index NEVER authenticates.
				if handled := reflexAnalysis(ctx, log, host, from, e.MessageText, username); handled {
					continue
				}
				// While analysis mode is active, the host is OFFLINE TO THE WORLD:
				// non-operator chat gets no in-character reply.
				if host.AnalysisActive() {
					continue
				}
				if time.Since(last) < 3*time.Second && !reflexMentor(host, from) {
					continue // light rate-limit so rapid lines don't spam replies (the MENTOR is never dropped)
				}
				// Per-speaker throttle + loop-breaker. A directly-actionable line (a
				// trade proposal, a question addressed to the host by name) bypasses
				// the cooldown but NOT the loop-breaker quiet window — the doom-spiral
				// lines were themselves name-addressed questions. The mentor (hostcfg
				// operator) is exempt from both, inside the gate.
				if ok, why := gate.allow(normalizeSpeaker(from), e.MessageText, username); !ok {
					log.Debug("social: reflex reply suppressed", "to", from, "why", why, "heard", e.MessageText)
					continue
				}
				// Knowledge-grounded reply (Deliverable 2): answer from what the host
				// actually KNOWS (hedged when low-confidence), stay out of factual
				// territory when it does NOT, and optionally volunteer a high-confidence
				// belief (the host↔host propagation seed). Honesty is a host-supplied
				// FACT, not a hope about the LLM. The C-25 knowledge slice rides the
				// ChatTurn as structured facts: mesad renders it under WHAT YOU
				// ACTUALLY KNOW — the model's only permitted source for factual replies.
				// ONE matcher feeds BOTH grounding seams — the C-25 slice (mesad's
				// WHAT YOU ACTUALLY KNOW) and groundReply's epistemic anchor — so
				// the two carriers can never contradict inside one prompt (a "you
				// do NOT know" line beside an attached fact left the spec's
				// canonical bronze-pickaxe answer to a coin-flip).
				known := host.chatKnownFacts(e.MessageText)
				rctx := socialContext(host, goal, doing, perception)
				rctx = append(rctx, host.groundReply(from, e.MessageText, time.Now(), known)...)
				text, speak, err := mc.Chat(ctx, username, from, e.MessageText, rctx, known)
				if err != nil || !speak || text == "" {
					continue
				}
				// This is a DIRECTED reply to `from` — route the host's own line into
				// ONLY the addressee's conversation window so the Q→A pairs there and
				// does not broadcast into every latched conversation (L7). Genuinely-
				// public DSL chat (actions_social say()) keeps the untargeted fan.
				host.reactive.directSelfTo(normalizeSpeaker(from))
				if err := host.Say(ctx, text); err != nil {
					host.reactive.clearDirectSelf() // send failed → drop the unconsumed routing hint
					// Never swallow this again — a silently-dropped reply (e.g. over the
					// RSC 80-char limit) looks exactly like "she isn't talking".
					log.Warn("social: reply failed to send", "to", from, "said", text, "err", err)
				} else {
					last = time.Now()
					log.Info("social: replied", "to", from, "heard", e.MessageText, "said", text)
					if gate.recordReply(normalizeSpeaker(from), reflexSelfPos(host)) {
						log.Info("social: conversational loop-breaker tripped — reflex going quiet for speaker",
							"speaker", from, "quiet_for", reflexLoopCooldown)
					}
				}
			}
		}
	}
}

// --- per-speaker reply throttle + conversational-loop breaker ----------------
//
// The reflex is a compulsion: every heard line used to be a candidate reply,
// and on social hosts 36-60% of ALL decisions became chat-driven (727 chat
// detours on drone4 in one night; the hour-long imaginary "tar brokerage"
// negotiation between bernard/drone6/drone8). The gate makes the reflex polite:
// reply, then hold off per speaker; and when a two-host exchange ping-pongs
// with no world progress, go quiet entirely and let the PLANNER decide whether
// the conversation is worth having.

const (
	// replyCooldown is the base per-speaker suppression window after a reflex
	// reply. Scaled by the persona's social-curiosity dial (replyCooldownFor).
	replyCooldown = 45 * time.Second
	// replyCooldownMin floors the persona-scaled cooldown so even a maximally
	// social host doesn't re-engage instantly.
	replyCooldownMin = 15 * time.Second
	// reflexLoopN is the consecutive-reply count to ONE speaker, with no world
	// progress between replies, that trips the loop-breaker.
	reflexLoopN = 3
	// reflexLoopCooldown is the breaker's quiet window: the reflex stops
	// answering that speaker entirely (even actionable lines) — the planner can
	// still CHOOSE to talk via the Act loop.
	reflexLoopCooldown = 3 * time.Minute
	// reflexLoopWindow: replies further apart than this never chain into a
	// loop — a slow, occasional exchange is conversation, not a spiral.
	reflexLoopWindow = 2 * time.Minute
	// reflexProgressTiles: moving at least this far (Chebyshev) between replies
	// to the same speaker counts as world progress and resets the streak — the
	// cheap progress signal available at this layer.
	reflexProgressTiles = 5
)

// reflexGate is the per-speaker reply governor for one host's social reflex.
// Keys are normalizeSpeaker'd names. Guarded by mu (the reflex goroutine is the
// only writer today; the lock keeps a future Host exposure safe).
type reflexGate struct {
	mu       sync.Mutex
	now      func() time.Time // injectable clock for tests
	cooldown time.Duration    // persona-scaled per-speaker reply cooldown
	operator func() string    // lazy hostcfg operator (the MENTOR) — exempt; nil/"" = nobody
	speakers map[string]*speakerGate
}

// speakerGate is the per-speaker state: when we last replied (cooldown), the
// no-progress reply streak (loop detection), where the host stood at the last
// reply (the progress signal), and the breaker's quiet-until deadline.
type speakerGate struct {
	lastReply  time.Time
	lastPos    world.Coord
	streak     int
	quietUntil time.Time
}

// newReflexGate builds a gate with the persona-scaled cooldown. social is the
// host's social-curiosity dial (0..1, host.curiosity.Social) — the chattiness
// dial cheaply reachable at this layer (HEXACO Extraversion itself is not
// captured on the Host at bootstrap; see replyCooldownFor). operator lazily
// names the hostcfg operator — the host's MENTOR, exempt from the throttle AND
// the loop-breaker, case-insensitively (nil or "" disables the exemption); it
// is a provider, not a snapshot, because configureAnalysis binds the operator
// AFTER the reflex goroutine constructs the gate.
func newReflexGate(social float64, operator func() string) *reflexGate {
	return &reflexGate{
		now:      time.Now,
		cooldown: replyCooldownFor(social),
		operator: operator,
		speakers: map[string]*speakerGate{},
	}
}

// mentor reports whether speaker (already normalizeSpeaker'd by callers) is the
// configured operator — the MENTOR exemption key. Case-insensitive: both sides
// fold through normalizeSpeaker. Unlike the analysis-takeover auth (which stays
// EXACT-match — a security gate), this is a courtesy gate: worst case a
// near-named passerby gets unthrottled replies, never control of the host.
func (g *reflexGate) mentor(speaker string) bool {
	if g.operator == nil {
		return false
	}
	op := normalizeSpeaker(g.operator())
	return op != "" && speaker == op
}

// replyCooldownFor scales the per-speaker cooldown by the social dial: a very
// social host (1.0) re-engages in ~22s, a neutral one (0.5) at the 45s base, a
// near-introvert (0.1) at ~63s. A zero dial means "no persona" (personaless
// REPL/test hosts) and keeps the flat default rather than reading as maximal
// introversion.
func replyCooldownFor(social float64) time.Duration {
	if social <= 0 {
		return replyCooldown
	}
	if social > 1 {
		social = 1
	}
	d := time.Duration((1.5 - social) * float64(replyCooldown))
	if d < replyCooldownMin {
		d = replyCooldownMin
	}
	return d
}

// allow reports whether the reflex may reply to speaker right now; on false,
// why names the suppressor ("cooldown" / "loop-breaker") for logs. Directly-
// actionable lines (actionableChat) bypass the cooldown but NOT the breaker's
// quiet window — the doom-spiral lines were themselves name-addressed questions,
// so an actionable bypass on the breaker would defeat it.
func (g *reflexGate) allow(speaker, message, username string) (bool, string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mentor(speaker) {
		return true, "" // MENTOR BYPASS: the operator is always engaged
	}
	s := g.speakers[speaker]
	if s == nil {
		return true, ""
	}
	now := g.now()
	if now.Before(s.quietUntil) {
		return false, "loop-breaker"
	}
	if !s.lastReply.IsZero() && now.Sub(s.lastReply) < g.cooldown && !actionableChat(message, username) {
		return false, "cooldown"
	}
	return true, ""
}

// recordReply notes a sent reflex reply to speaker at the host position pos and
// returns true when this reply tripped the loop-breaker (the caller logs it).
// The streak counts consecutive replies to the same speaker with no world
// progress between them; movement of reflexProgressTiles+ or a reflexLoopWindow
// gap resets it.
func (g *reflexGate) recordReply(speaker string, pos world.Coord) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.mentor(speaker) {
		return false // MENTOR BYPASS: no streak/cooldown bookkeeping for the operator
	}
	now := g.now()
	s := g.speakers[speaker]
	if s == nil {
		s = &speakerGate{}
		g.speakers[speaker] = s
	}
	fresh := s.streak == 0 ||
		now.Sub(s.lastReply) > reflexLoopWindow ||
		chebyshev(s.lastPos.X, s.lastPos.Y, pos.X, pos.Y) >= reflexProgressTiles
	if fresh {
		s.streak = 1
	} else {
		s.streak++
	}
	s.lastReply = now
	s.lastPos = pos
	if s.streak >= reflexLoopN {
		s.quietUntil = now.Add(reflexLoopCooldown)
		s.streak = 0
		return true
	}
	return false
}

// actionableChat reports whether a chat line is directly actionable — a trade
// proposal, or a direct question addressed to the host by name — and so worth
// answering even inside the per-speaker cooldown.
func actionableChat(message, username string) bool {
	low := strings.ToLower(message)
	if strings.Contains(low, "trade") {
		return true
	}
	return username != "" && strings.Contains(low, strings.ToLower(username)) && strings.Contains(low, "?")
}

// reflexMentor reports whether `from` is the configured hostcfg operator — the
// host's MENTOR: always engaged, case-insensitively, so an operator line never
// falls to the global rate-limit (the per-speaker gate carries its own
// exemption). Distinct from the analysis-takeover auth in reflexAnalysis, which
// stays EXACT-match — that one is a security gate, this one is a courtesy gate.
func reflexMentor(host *Host, from string) bool {
	op := strings.TrimSpace(host.AnalysisOperator())
	return op != "" && strings.EqualFold(strings.TrimSpace(from), op)
}

// --- the C-25 chat knowledge slice -------------------------------------------

// chatKnownTopK caps how many ledger subjects ride one ChatTurn: the grounding
// slice stays a slice, not a ledger dump.
const chatKnownTopK = 5

// chatKnownFacts keyword-matches an inbound chat line against the knowledge
// ledger and returns the host's best belief per touched subject, strongest
// match first, capped at chatKnownTopK (C-25: the bernard "not familiar with
// 'prospect'" exhibit — the never-bluff order finally gets facts). A subject
// matches when the line contains it whole on word boundaries, or contains its
// significant tokens (the goalTouch/overlapCount vocabulary: ≥4 chars,
// non-stopword; trivial plurals fold). The best belief's CLAIM text is scored
// too: knowledge keyed under an NPC/shop subject ("Baraek" → "sells picks at
// 20 coin each") answers item-worded questions the subject alone never would —
// writebackClaims falls back to subj=speaker and ExtractDialog emits npc/shop
// subjects, so those shapes are the norm, not the exception. Only beliefs the
// host actually HOLDS (confidence ≥ epConfFloor — the one knowing-vs-guessing
// floor) attach: an empty result IS the signal, since mesad's reply contract
// reads "no facts" as "factual questions get silence". Deterministic + cheap
// (O(subjects)); no LLM, no I/O.
func (h *Host) chatKnownFacts(message string) []mesaclient.KnownFact {
	if h.knowledge == nil {
		return nil
	}
	lowMsg := strings.ToLower(message)
	type hit struct {
		score, conf          float64
		subject, claim, prov string
	}
	var hits []hit
	for _, f := range h.knowledge.All() {
		if f.Confidence < epConfFloor || len(f.Beliefs) == 0 {
			continue
		}
		best := f.Beliefs[0] // the ledger view sorts beliefs best-first
		score := chatSubjectScore(lowMsg, f.Subject)
		if cs := chatClaimScore(lowMsg, best.Claim); cs > score {
			score = cs // the claim-text recall arm (NPC/shop-keyed knowledge)
		}
		if score <= 0 {
			continue
		}
		hits = append(hits, hit{
			score: score, conf: f.Confidence,
			subject: f.Subject, claim: best.Claim, prov: best.Provenance,
		})
	}
	if len(hits) == 0 {
		return nil // the empty slice IS the signal: nothing held bears on this line
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].score != hits[j].score {
			return hits[i].score > hits[j].score
		}
		if hits[i].conf != hits[j].conf {
			return hits[i].conf > hits[j].conf
		}
		return hits[i].subject < hits[j].subject // deterministic tail
	})
	if len(hits) > chatKnownTopK {
		hits = hits[:chatKnownTopK]
	}
	out := make([]mesaclient.KnownFact, 0, len(hits))
	for _, t := range hits {
		out = append(out, mesaclient.KnownFact{
			Subject:    t.subject,
			Claim:      t.claim,
			Confidence: t.conf,
			Provenance: t.prov,
		})
	}
	return out
}

// chatSubjectScore is the deterministic match strength between a lowercased
// chat line and a ledger subject: 1.0 when the line contains the whole subject
// on WORD BOUNDARIES (which covers short subjects like "tin" the token filter
// would drop — boundary-anchored so "tin" never false-positives inside
// "waiting", "ore" inside "more"), else the fraction of the subject's
// significant tokens (≥4 chars, non-stopword — the goalTouch vocabulary) the
// line contains, plural-folded. 0 = no match. Synonym expansion through the
// resolve alias store is C-32's seam, not this one (this matcher is where it
// will apply).
func chatSubjectScore(lowMsg, subject string) float64 {
	subj := strings.ToLower(strings.TrimSpace(subject))
	if subj == "" || lowMsg == "" {
		return 0
	}
	if strings.Contains(chatBoundary(lowMsg), chatBoundary(subj)) {
		return 1
	}
	total := significantTokenCount(subj)
	if total == 0 {
		return 0
	}
	return float64(chatTokenOverlap(subj, lowMsg)) / float64(total)
}

// chatClaimScore is the claim-text recall arm of the chat matcher: the
// fraction of the claim's significant tokens the line contains (plural-
// folded). Catches knowledge keyed under an NPC/shop/plural subject —
// "Baraek" → "sells picks at 20 coin each", "Bob's Axes" → "sells pickaxes" —
// where the SUBJECT match scores 0 but the claim genuinely answers the
// question. Partial hits score fractional, so a whole-subject hit (1.0) still
// outranks them.
func chatClaimScore(lowMsg, claim string) float64 {
	c := strings.ToLower(strings.TrimSpace(claim))
	if c == "" || lowMsg == "" {
		return 0
	}
	total := significantTokenCount(c)
	if total == 0 {
		return 0
	}
	return float64(chatTokenOverlap(c, lowMsg)) / float64(total)
}

// chatTokenOverlap counts the distinct significant (≥4-char, non-stopword)
// tokens of target that appear in line — overlapCount widened with the trivial
// plural fold, scoped to the CHAT matcher only (the question closer keeps the
// shared, stricter overlapCount: loosening recall there would auto-close goals).
func chatTokenOverlap(target, line string) int {
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
		if chatTokenInLine(line, w) {
			n++
		}
	}
	return n
}

// chatTokenInLine reports whether token w appears in line, retrying with the
// trailing-"s" plural fold ("pickaxes" finds "pickaxe"; "picks" → "pick"
// finds "pickaxe"). The fold keeps the vocabulary's loose Contains semantics
// at ≥4 folded chars; a 3-char fold ("axes" → "axe") must land on a word
// boundary so it cannot ride inside "taxes".
func chatTokenInLine(line, w string) bool {
	if strings.Contains(line, w) {
		return true
	}
	s := strings.TrimSuffix(w, "s")
	if s == w || len(s) < 3 {
		return false
	}
	if len(s) >= 4 {
		return strings.Contains(line, s)
	}
	return strings.Contains(chatBoundary(line), " "+s+" ")
}

// chatBoundary normalizes a (lowercased) string for word-boundary matching:
// every rune that isn't a letter/digit/apostrophe becomes a space (apostrophes
// survive so "bob's axes" stays one phrase), whitespace runs collapse, and the
// result is space-padded — Contains(chatBoundary(line), chatBoundary(subj)) is
// then a whole-word phrase match.
func chatBoundary(s string) string {
	mapped := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' {
			return r
		}
		return ' '
	}, s)
	return " " + strings.Join(strings.Fields(mapped), " ") + " "
}

// reflexSelfPos is the host's current position, or the zero coord when the
// world mirror isn't live yet (then "no movement" is the safe reading).
func reflexSelfPos(host *Host) world.Coord {
	if w := host.World(); w != nil && w.Self != nil {
		return w.Self.Position()
	}
	return world.Coord{}
}

// socialContext gathers RICH context for a chat reply: her goal, what she's
// doing right now (latest reasoning), what she recently perceived, and the live
// game messages — so she can actually answer questions like "what are you doing?"
func socialContext(host *Host, goal, doing, perception string) []string {
	out := make([]string, 0, 8)
	if goal != "" {
		out = append(out, "Your overall goal: "+goal)
	}
	if doing != "" {
		out = append(out, "What you are doing RIGHT NOW: "+doing)
	}
	if perception != "" {
		out = append(out, "What you've recently seen/heard: "+perception)
	}
	if w := host.World(); w != nil && w.Self != nil {
		pos := w.Self.Position()
		out = append(out, fmt.Sprintf("You are at (%d,%d), HP %d/%d.", pos.X, pos.Y, w.Self.HP(), w.Self.MaxHP()))
	}
	return out
}

// reflexAnalysis handles the in-game operator-override channel for one inbound
// chat line. It returns true when the line was consumed by analysis mode (a
// trigger, or an operator directive while active) and the caller must NOT fall
// through to the normal social reply.
//
// AUTH (mandatory — a host-takeover vector otherwise): the line is only honored
// when `from` is an EXACT, case-sensitive, non-empty match for the configured
// operator. An unconfigured operator ("") disables the in-game channel entirely;
// reflexPlayerName's "a player" placeholder for an unresolved index can never
// match a real operator name. RSC names ride an untrusted channel, so this is
// the only gate between a passerby and a full host takeover.
func reflexAnalysis(ctx context.Context, log *slog.Logger, host *Host, from, message, username string) bool {
	op := host.AnalysisOperator()
	if op == "" || from == "" || from != op {
		return false // no operator configured, or sender is not the exact operator
	}
	trimmed := strings.TrimSpace(message)
	prefix := "!" + username + " "
	if cmd, ok := cutPrefixFold(trimmed, prefix); ok {
		switch {
		case cmd == "ANALYSIS": // UPPERCASE keyword per spec
			res := host.EnterAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "enter")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		case cmd == "RESUME":
			res := host.ExitAnalysis()
			log.Info("analysis: in-game trigger", "operator", from, "action", "exit")
			_ = host.Say(ctx, "[analysis] "+res.Text)
			return true
		}
		// "!<username> <directive>" while active: route the rest as a directive.
		if host.AnalysisActive() {
			res := host.Analyze(ctx, cmd)
			log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", cmd)
			_ = host.Say(ctx, analysisSayLine(res))
			return true
		}
		return false
	}
	// No "!<username>" addressing: while active, any bare operator line is a
	// directive (the operator is already in a takeover session with this host).
	if host.AnalysisActive() {
		res := host.Analyze(ctx, trimmed)
		log.Info("analysis: in-game directive", "operator", from, "kind", res.Kind, "directive", trimmed)
		_ = host.Say(ctx, analysisSayLine(res))
		return true
	}
	return false
}

// cutPrefixFold strips prefix (case-insensitively) from s, reporting whether it
// matched. The host-username addressing "!Delores " should match regardless of
// the operator's capitalization of the name, while the ANALYSIS/RESUME keyword
// itself stays case-sensitive (checked by the caller).
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", false
	}
	return strings.TrimSpace(s[len(prefix):]), true
}

// analysisSayLine renders an analysis verdict as a single flat in-game line back
// to the operator (capped to the RSC chat limit by the action layer). Flat
// affect: terse + literal, prefixed so the operator can tell it from in-persona
// speech.
func analysisSayLine(res AnalysisResult) string {
	body := res.Text
	if res.Err != "" && body == "" {
		body = "error: " + res.Err
	}
	return "[analysis:" + string(res.Kind) + "] " + body
}

func reflexPlayerName(host *Host, idx int) string {
	if w := host.World(); w != nil && w.Players != nil {
		for _, p := range w.Players.All() {
			if p.Index == idx && p.Name != "" {
				return p.Name
			}
		}
	}
	return "a player"
}
