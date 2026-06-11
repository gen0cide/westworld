package mesad

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// Genesis is the session-genesis compile: one heavy Opus call, at login, that
// reads the host's full context — persona + history (episodes) + relationships +
// standing goal, all gathered mesa-side — and compiles the session apparatus the
// host runs cheap all session: a history-aware GOAL, a MOOD baseline, and the
// keyword→tier→action ladder that orients its attention. This is the producer
// the canonical doc (§2) describes; the host applies the result at startup.
func (s *Server) Genesis(ctx context.Context, req *mesapb.GenesisRequest) (*mesapb.GenesisResult, error) {
	hostID := hostIDFromContext(ctx)
	if s.genesisLLM == nil {
		return nil, status.Error(codes.Unavailable, "genesis llm not configured")
	}
	ctx, cancel := ensureDeadline(ctx, genesisDeadline) // backstop for deadline-less clients
	defer cancel()
	e, ok := s.lookup(hostID)
	if !ok {
		return nil, status.Errorf(codes.NotFound, "no persona registered for host_id %q", hostID)
	}

	// Gather the host's history mesa-side (the whole point of building memory first).
	var episodes []storedEpisode
	var rels []*mesapb.Relationship
	var objective string
	var progress []string
	if s.ltm != nil {
		episodes, _ = s.ltm.Recall(ctx, hostID, "", 30) // recent history, recency-ordered
		rels, _ = s.ltm.Relationships(ctx, hostID)
		objective, progress, _, _ = s.ltm.Goal(ctx, hostID)
	}

	// Relogin churn (fleet bounces, sleep/wake storms) re-pays the most
	// expensive call in the system for an identical compile: one overnight of
	// wake cycles burned ~717 geneses; one fleet bounce costs ~102. Cache the
	// RAW compile in the host's KV, keyed by persona revision + the history
	// the prompt actually saw; parseGenesis is pure, so a hit replays it.
	rev := genesisRev(e.prose, len(episodes), len(rels), objective)
	if s.ltm != nil {
		if hit, ok := s.genesisCacheGet(ctx, hostID, rev); ok {
			res := parseGenesis(hit)
			s.log.Info("genesis cache hit", "host_id", hostID, "goal", res.GetGoal(), "rev", rev[:12])
			return res, nil
		}
	}

	raw, err := s.genesisLLM.Complete(ctx, genesisSystem(e.prose),
		genesisPrompt(req, episodes, rels, objective, progress), 1500)
	if err != nil {
		s.log.Warn("genesis llm error", "host_id", hostID, "err", err)
		return nil, status.Errorf(codes.Internal, "genesis llm: %v", err)
	}
	res := parseGenesis(raw)
	if s.ltm != nil {
		s.genesisCachePut(ctx, hostID, rev, raw)
	}
	s.log.Info("genesis compiled", "host_id", hostID, "goal", res.GetGoal(),
		"keywords", len(res.GetKeywordLadder()), "aspirations", len(res.GetAspirations()),
		"episodes_read", len(episodes), "relationships_read", len(rels))
	return res, nil
}

// genesisCacheTTL bounds how long a compile is replayed: long enough to absorb
// bounce storms and sleep/wake churn, short enough that a session of real play
// (which also shifts the rev via episode count) earns a fresh reflection.
const genesisCacheTTL = time.Hour

const genesisCacheKey = "genesis-cache"

// genesisRev fingerprints everything the genesis prompt depends on: the
// persona prose (identity card) and the shape of the history it would read.
// Any new episode/relationship/goal change produces a different rev.
func genesisRev(prose string, episodes, rels int, objective string) string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s|%d|%d|%s", prose, episodes, rels, objective)))
	return hex.EncodeToString(h[:])
}

type cachedGenesis struct {
	Rev string `json:"rev"`
	At  int64  `json:"at"`
	Raw string `json:"raw"`
}

func (s *Server) genesisCacheGet(ctx context.Context, hostID, rev string) (string, bool) {
	val, found, err := s.ltm.GetKV(ctx, hostID, genesisCacheKey)
	if err != nil || !found {
		return "", false
	}
	var c cachedGenesis
	if json.Unmarshal(val, &c) != nil || !genesisCacheFresh(c, rev, time.Now()) {
		return "", false
	}
	return c.Raw, true
}

// genesisCacheFresh is the pure accept test: same rev, within TTL.
func genesisCacheFresh(c cachedGenesis, rev string, now time.Time) bool {
	return c.Rev == rev && now.Sub(time.Unix(c.At, 0)) <= genesisCacheTTL
}

func (s *Server) genesisCachePut(ctx context.Context, hostID, rev, raw string) {
	b, err := json.Marshal(cachedGenesis{Rev: rev, At: time.Now().Unix(), Raw: raw})
	if err != nil {
		return
	}
	if err := s.ltm.PutKV(ctx, hostID, genesisCacheKey, b); err != nil {
		s.log.Warn("genesis cache write failed", "host_id", hostID, "err", err)
	}
}

// genesisSystem grounds the compile in the host's identity.
func genesisSystem(prose string) string {
	return strings.TrimSpace(prose) + "\n\n" + genesisInstruction
}

const genesisInstruction = `You are this character in RuneScape Classic, WAKING UP for a new session. This is a rare, deliberate moment of self-reflection — you compile how you will live THIS session, then run on it.

Given who you are, what you have done, who you know, and where you are, decide:
1. GOAL — your intent for this session, in character. Build on your history; don't restart from scratch. One or two sentences, first person.
2. MOOD — your emotional baseline this session as three numbers: stress (0..1), confidence (0..1), valence (-1..1). Reflect your temperament + how your recent history would leave you feeling.
3. KEYWORD LADDER — words/patterns in others' chat that should catch your attention, each with a tier and a default reflex. Tiers: DIRECTED_SOCIAL (your name, friends — orient + consider replying), TRADE_INTEREST (trade/sell/buy/offer/free — consider it; trading is normal and judged by relationship + the actual deal, NEVER by the word "free"), TOPIC_WATCH (topics tied to your goal — consider engaging), AMBIENT (idle chatter — never preempt). Derive from YOUR history and goal — include your own name, any friends/rivals you know, words tied to your goal, and trade words. Do not treat generosity as a threat.
4. ASPIRATIONS — 2 to 4 standing ambitions on a months-long horizon, derived from who you are (your north star, what pulls your curiosity, your values). Short phrases, e.g. "master a craft (smithing or crafting)", "earn a place in the Champions' Guild", "come to know the whole map". These persist across sessions and never finish in a day; your GOAL (1) should be a concrete step toward ONE of them — repeat that one verbatim in "opening_serves".

Respond ONLY as JSON:
{"goal":"...","mood":{"stress":0.0,"confidence":0.0,"valence":0.0},"keyword_ladder":[{"keyword":"...","tier":"DIRECTED_SOCIAL","action":"..."}],"aspirations":["...","..."],"opening_serves":"...","reasoning":"one line on why"}`

// genesisPrompt assembles the host's full context for the compile.
func genesisPrompt(req *mesapb.GenesisRequest, episodes []storedEpisode, rels []*mesapb.Relationship, objective string, progress []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "WAKING UP (%s).\n\n", orElse(req.GetTrigger(), "login"))
	if ws := strings.TrimSpace(req.GetWorldSummary()); ws != "" {
		fmt.Fprintf(&b, "WHERE YOU ARE NOW: %s\n\n", ws)
	}
	if objective != "" {
		fmt.Fprintf(&b, "YOUR STANDING GOAL (from before): %s\n", objective)
		if len(progress) > 0 {
			b.WriteString("Progress so far: " + strings.Join(progress, "; ") + "\n")
		}
		b.WriteString("\n")
	}
	if len(episodes) > 0 {
		b.WriteString("WHAT YOU'VE DONE (most recent first):\n")
		for _, e := range episodes {
			fmt.Fprintf(&b, "- %s\n", e.Text)
		}
		b.WriteString("\n")
	} else {
		b.WriteString("WHAT YOU'VE DONE: nothing yet — this is early in your life.\n\n")
	}
	if len(rels) > 0 {
		b.WriteString("WHO YOU KNOW (trust ledger):\n")
		for _, r := range rels {
			fmt.Fprintf(&b, "- %s: %s (met %d×%s)\n", r.GetName(), trustWord(r), r.GetEncounters(), tagSuffix(r.GetTags()))
		}
		b.WriteString("\n")
	}
	b.WriteString("Compile your session now. Return ONLY the JSON.")
	return b.String()
}

// trustWord summarizes a relationship's Beta(alpha,beta) as a coarse word.
func trustWord(r *mesapb.Relationship) string {
	a, bta := r.GetAlpha(), r.GetBeta()
	if a+bta <= 0 {
		return "neutral"
	}
	mean := a / (a + bta) // 0..1
	switch {
	case mean >= 0.7:
		return "trusted"
	case mean >= 0.55:
		return "friendly"
	case mean <= 0.3:
		return "distrusted"
	case mean <= 0.45:
		return "wary"
	default:
		return "neutral"
	}
}

func tagSuffix(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	return "; " + strings.Join(tags, ", ")
}

func orElse(s, d string) string {
	if strings.TrimSpace(s) == "" {
		return d
	}
	return s
}

// maxGenesisAspirations caps the aspiration list the compile may emit — the
// portfolio stays a HANDFUL of month-horizon ambitions, not a wishlist.
const maxGenesisAspirations = 4

// parseGenesis extracts the model's JSON into a GenesisResult. Falls back to a
// minimal result on a parse failure so the host still boots. Aspirations are
// VALIDATED, never trusted: blanks/duplicates dropped, list capped at
// maxGenesisAspirations, and opening_serves kept only when it names one of the
// kept aspirations (case-insensitive) — an absent/invalid value degrades to ""
// and the host's mechanical nearest-match takes over. No aspirations at all is
// fine: the host behaves exactly as before the ladder existed.
func parseGenesis(raw string) *mesapb.GenesisResult {
	var v struct {
		Goal string `json:"goal"`
		Mood struct {
			Stress     float64 `json:"stress"`
			Confidence float64 `json:"confidence"`
			Valence    float64 `json:"valence"`
		} `json:"mood"`
		KeywordLadder []struct {
			Keyword string `json:"keyword"`
			Tier    string `json:"tier"`
			Action  string `json:"action"`
		} `json:"keyword_ladder"`
		Aspirations   []string `json:"aspirations"`
		OpeningServes string   `json:"opening_serves"`
		Reasoning     string   `json:"reasoning"`
	}
	js := extractJSON(raw)
	if js == "" || json.Unmarshal([]byte(js), &v) != nil {
		return &mesapb.GenesisResult{Reasoning: "genesis parse failed; running on defaults"}
	}
	res := &mesapb.GenesisResult{
		Goal:      strings.TrimSpace(v.Goal),
		Mood:      &mesapb.Affect{Stress: v.Mood.Stress, Confidence: v.Mood.Confidence, Valence: v.Mood.Valence},
		Reasoning: strings.TrimSpace(v.Reasoning),
	}
	for _, r := range v.KeywordLadder {
		if strings.TrimSpace(r.Keyword) == "" {
			continue
		}
		res.KeywordLadder = append(res.KeywordLadder, &mesapb.KeywordRung{
			Keyword: r.Keyword, Tier: r.Tier, Action: r.Action,
		})
	}
	seen := map[string]bool{}
	for _, a := range v.Aspirations {
		a = strings.TrimSpace(a)
		key := strings.ToLower(a)
		if a == "" || seen[key] || len(res.Aspirations) >= maxGenesisAspirations {
			continue
		}
		seen[key] = true
		res.Aspirations = append(res.Aspirations, a)
	}
	if os := strings.TrimSpace(v.OpeningServes); os != "" && seen[strings.ToLower(os)] {
		res.OpeningServes = os
	}
	return res
}
