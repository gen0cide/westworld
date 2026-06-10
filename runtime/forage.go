package runtime

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gen0cide/westworld/cognition/goalgraph"
	"github.com/gen0cide/westworld/event"
	mesaclient "github.com/gen0cide/westworld/mesa/client"
	"github.com/gen0cide/westworld/worldmap"
)

// Phase 5b — DIRECTED INFORMATION-FORAGING (the effective-search gap). When no
// local interlocutor can answer a blocking where-to-buy/where-is question (RSC
// NPCs walk a canned tree, they are not oracles), the host LEAVES and goes to
// LOOK: it travels to the nearest untried reachable shop, lets perception harvest
// what's there (incl. NEGATIVE "not sold here" knowledge), and either closes the
// question on an in-stock find or marks the place spent and rotates to the next.
//
// 5b-2 is the ARM-1 arbiter (nextForageSource: pure exploration, no LLM/RPC, no
// travel). 5b-3 is the drive (tryForage: the tryAsk twin that issues the travel
// detour + the perception harvest + the observation-closer). Per-(source,topic)
// exhaustion lives as TAGS on the question node (P0 #1: no new graph nodes). The
// LLM ProposeSources arm + the box-escape give-up are later phases.

// forageGcMax bounds the visited cache; symmetric with speechGCMax. A stale entry
// past this re-arms the RAM cache — but the DURABLE source-tried/source-spent:place
// tags on Q are NOT re-armed here (only on Q close); the cache is a speed shortcut.
const forageGcMax = speechGCMax

// forageGate holds the forage drive's RAM-only state. The actuation FLOOR is SHARED
// with tryAsk (host.speech.lastAsk) — this gate carries NO floor. The durable
// per-(Q,place) exhaustion lives as TAGS on the question node (P0 #1: no new graph
// nodes); visited is only a bounded cache of source-tried:place; inflight records
// the qid + issue-time so gcForage can TTL-clear a stuck forage-inflight tag (the
// PATH_BLOCKED-with-no-ShopOpened backstop).
type forageGate struct {
	mu       sync.Mutex
	visited  map[string]time.Time // "qid|label" -> last visit (mirror of source-tried:place)
	inflight map[string]time.Time // qid -> when the forage detour was issued (TTL backstop)
}

func newForageGate() *forageGate {
	return &forageGate{
		visited:  map[string]time.Time{},
		inflight: map[string]time.Time{},
	}
}

// gcForage drops cache entries older than forageGcMax AND clears any forage-inflight
// tag whose detour was issued longer ago than detourTimeout (a PATH_BLOCKED forage
// produces no ShopOpened to clear it — this is the required backstop so a stuck tag
// never permanently disables SPINNING for the host). Mutex-guarded.
func (g *forageGate) gcForage(now time.Time, h *Host) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for k, t := range g.visited {
		if now.Sub(t) > forageGcMax {
			delete(g.visited, k)
		}
	}
	for qid, t := range g.inflight {
		if now.Sub(t) > detourTimeout {
			if h != nil && h.goalGraph != nil {
				h.goalGraph.Untag(qid, "forage-inflight")
			}
			delete(g.inflight, qid)
		}
	}
}

func (g *forageGate) markVisited(qid, label string, now time.Time) {
	g.mu.Lock()
	g.visited[qid+"|"+label] = now
	g.mu.Unlock()
}

// noteInflight records that a forage detour was just issued for qid, so gcForage can
// TTL-clear its forage-inflight tag if no harvest ever arrives.
func (g *forageGate) noteInflight(qid string, now time.Time) {
	g.mu.Lock()
	g.inflight[qid] = now
	g.mu.Unlock()
}

// clearInflight drops the TTL record once the harvest cleared the tag normally.
func (g *forageGate) clearInflight(qid string) {
	g.mu.Lock()
	delete(g.inflight, qid)
	g.mu.Unlock()
}

// forageSource is the chosen ARM-1 destination: a label (the source-tried/spent key
// + what the host reasons with) and the standable SnapX/SnapY coords travel issues
// go_to() on.
type forageSource struct {
	label string
	x, y  int // SnapX, SnapY
}

// forageTopicPOITypes maps a blocking factual question to the POI type SUBSTRINGS
// worth visiting. where-to-buy:<item> -> ["shop"] (substring; the gazetteer has no
// bare "shop" type, so "shop" matches general-shop/axe-shop/... and STRUCTURALLY
// EXCLUDES mining-site/bank/altar — ARM-1 can never pick go_to(mining-site) for a
// where-to-buy topic). where-is:<subject> -> [<subject>] best-effort. Returns nil
// for confirm:/how-to-progress: (forage stays place/item-only).
func forageTopicPOITypes(label string) []string {
	low := strings.ToLower(strings.TrimSpace(label))
	switch {
	case strings.Contains(low, "where to buy") || strings.HasPrefix(low, "where-to-buy:"):
		return []string{"shop"} // substring; matches every *-shop POI type
	case strings.Contains(low, "where is") || strings.HasPrefix(low, "where-is:"):
		if subj := salientTopic(label); subj != "" {
			return []string{subj}
		}
	}
	return nil
}

// matchesAnyType reports whether the POI type contains any of the type substrings,
// case-insensitive — the same test dslSearchMap + nearestExplainedPOI use.
func matchesAnyType(poiType string, types []string) bool {
	low := strings.ToLower(poiType)
	for _, t := range types {
		if strings.Contains(low, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

// nextForageSource picks the nearest UNTRIED reachable POI of a type appropriate to
// Q. ARM-1 PURE EXPLORATION: no mesa RPC, no travel. Deterministic: chebyshev-nearest
// wins, ties by label. Returns ok=false when no untried ReachOpen source of a mapped
// type exists. Chebyshev-RANKS candidates first, then runs the (reachable-flood)
// ExplainReach only in nearest-first order until one returns ReachOpen — not per
// candidate (the 200-host efficiency concern).
func (h *Host) nextForageSource(q goalgraph.Node, now time.Time) (forageSource, bool) {
	if h.worldOracle == nil || h.world == nil || h.world.Self == nil || h.goalGraph == nil {
		return forageSource{}, false
	}
	types := forageTopicPOITypes(q.Label)
	if len(types) == 0 {
		return forageSource{}, false
	}
	pos := h.world.Self.Position()
	cap := h.hostCapability()

	type cand struct {
		src  forageSource
		dist int
	}
	var cands []cand
	for _, p := range h.worldOracle.POIs() {
		if p.Comp < 0 { // never snapped to a standable tile
			continue
		}
		if !matchesAnyType(p.Type, types) {
			continue
		}
		label := h.destinationLabel(p) // the tag key
		if h.goalGraph.HasTag(q.ID, "source-tried:place:"+label) ||
			h.goalGraph.HasTag(q.ID, "source-spent:place:"+label) {
			continue // already visited / confirmed-absent for this question
		}
		cands = append(cands, cand{
			src:  forageSource{label: label, x: p.SnapX, y: p.SnapY},
			dist: chebyshev(pos.X, pos.Y, p.SnapX, p.SnapY),
		})
	}
	if len(cands) == 0 {
		return h.frontierForageSource(q) // fog fallback: every mapped source tried/spent (fog.go)
	}
	sort.Slice(cands, func(i, j int) bool {
		if cands[i].dist != cands[j].dist {
			return cands[i].dist < cands[j].dist
		}
		return cands[i].src.label < cands[j].src.label // stable tie-break
	})
	for _, c := range cands {
		info := h.worldOracle.ExplainReach(pos.X, pos.Y, c.src.x, c.src.y, cap)
		if !info.SnapOK() {
			continue
		}
		if info.Reach == worldmap.ReachOpen { // ARM-1: ReachOpen ONLY
			return c.src, true
		}
	}
	return h.frontierForageSource(q) // fog fallback: no mapped source reachable (fog.go)
}

// pickForageTarget selects exactly ONE factual where-to-buy/where-is question
// blocking the effective goal (Pass-1 only — forage maps only place/item questions
// to POIs). Reads the SAME effectiveGoalForSpeech()+Blockers() pair as
// pickAskQuestion so picker/forager cannot desync. NOT gated on the speech
// askedQ/attempts cooldowns (those bound NPC asking, not travel) — forage's bound is
// the per-Q source-tried/source-spent tag set. Pure graph read.
func (h *Host) pickForageTarget(now time.Time) (goalgraph.Node, bool) {
	if h.goalGraph == nil {
		return goalgraph.Node{}, false
	}
	g := h.effectiveGoalForSpeech()
	if g == "" {
		return goalgraph.Node{}, false
	}
	blocking := map[string]bool{}
	for _, b := range h.goalGraph.Blockers(g) {
		if b.Kind == goalgraph.KindOpenQuestion {
			blocking[strings.ToLower(strings.TrimSpace(b.ID))] = true
		}
	}
	for _, q := range h.goalGraph.OpenQuestions() { // newest-first
		if !blocking[strings.ToLower(strings.TrimSpace(q.ID))] {
			continue
		}
		if !isFactualTipQuestion(q.Label) {
			continue // forage only travels for place/item questions
		}
		return q, true
	}
	return goalgraph.Node{}, false
}

// tryForage is the deterministic FORAGE drive — the proactive twin of tryAsk that
// makes the host LEAVE and go LOOK when no local interlocutor can answer a blocking
// where-to-buy/where-is question. Called from socialReflex AFTER tryAsk on the
// AgentThought tick. ARM-1 PURE EXPLORATION: no LLM, no mesa RPC. Shares the
// askGlobalGap floor with tryAsk (host.speech.lastAsk): exactly one of {ask, forage}
// actuates per gap, and since tryForage runs after tryAsk it fires only when tryAsk
// no-oped this tick. Frozen under analysis mode. mc is UNUSED in ARM-1 (kept for the
// 5b-5 ProposeSources arm); do NOT gate on mc.Healthy() here.
func tryForage(ctx context.Context, log *slog.Logger, host *Host, mc mesaclient.Client, username string) {
	if host == nil || host.AnalysisActive() {
		return
	}
	if host.speech == nil || host.goalGraph == nil || host.worldOracle == nil || host.forage == nil {
		return
	}
	now := time.Now()

	// SHARED global floor (with tryAsk). If tryAsk just fired this tick, lastAsk is
	// fresh => this no-ops, preserving "one of {ask, forage} per gap".
	host.speech.mu.Lock()
	if now.Sub(host.speech.lastAsk) < askGlobalGap {
		host.speech.mu.Unlock()
		return
	}
	host.speech.mu.Unlock()

	// Pick exactly ONE blocking factual question (place/item only, Pass-1).
	q, ok := host.pickForageTarget(now)
	if !ok {
		return
	}
	// ARM-1 arbiter: nearest untried reachable source. No travel yet.
	src, ok := host.nextForageSource(q, now)
	if !ok {
		return // no source — gap UNBURNED so tryAsk can still fire next tick
	}
	// Reach the conductor; bail cleanly if no detour stack (REPL/test).
	c := host.conductorHandle()
	if c == nil || !c.detours {
		return
	}

	// BURN the shared floor NOW — the COMMIT point (a real travel side-effect
	// follows). A no-source tick above never reaches here, so never burns.
	host.speech.mu.Lock()
	host.speech.lastAsk = now
	host.speech.mu.Unlock()

	// Mark the source TRIED + the forage-inflight breadcrumb (the spin-suppression
	// signal) + the cache mirror + the TTL record. Set the perception cursor BEFORE
	// issuing the detour (arrival -> ShopOpened is async on the limbic goroutine).
	host.goalGraph.Tag(q.ID, "source-tried:place:"+src.label)
	host.goalGraph.Tag(q.ID, "forage-inflight") // cleared on harvest or TTL
	host.forage.markVisited(q.ID, src.label, now)
	host.forage.noteInflight(q.ID, now)
	host.setForageCursor(salientTopic(q.Label), src.label, q.ID)

	// Issue the forage detour: park grind -> go_to(SnapX,SnapY) -> resume grind.
	c.signalDetour(detourReq{
		tier:   "forage",
		abort:  false,
		reason: "foraging " + q.Label + " -> " + src.label,
		intent: forageIntent(src.x, src.y, src.label),
	})

	if log != nil {
		log.Info("forage: traveling to source", "question", q.Label, "dest", src.label,
			"coords", fmt.Sprintf("(%d,%d)", src.x, src.y))
	}
	host.publishDecision("forage", "travel", "no local source could answer \""+q.Label+"\" — going to look at "+src.label)
	host.bus.Publish(event.AgentThought{
		Trigger:    "forage",
		Goal:       host.LiveGoal(),
		Perception: "foraging " + q.Label + " at " + src.label,
		Reasoning:  "no local source could answer; going to look at " + src.label,
	})
}

// forageIntent is the forage travel detour: walk to the chosen source COORDS (never
// a label — go_to(label) only resolves towns/landmarks, NOT POI labels). go_to(x, y)
// two ints hits dslGoTo's resolvePoint -> h.GoTo path. OneShot mirrors the other
// detour intents.
func forageIntent(x, y int, label string) Intent {
	src := fmt.Sprintf("runtime \"1.0\"\nroutine forage_walk() {\n    note(\"foraging: going to %s to look.\")\n    go_to(%d, %d)\n}",
		sanitizeForageLabel(label), x, y)
	return Intent{Label: "detour:forage_walk", Name: "forage_walk", Source: src, OneShot: true}
}

// sanitizeForageLabel strips the characters that would break the note("...") DSL
// string literal forageIntent interpolates the label into (defense-in-depth; gazetteer
// labels are controlled, but a stray quote/backslash must never crash the routine).
func sanitizeForageLabel(s string) string {
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\"", "'")
	return s
}
