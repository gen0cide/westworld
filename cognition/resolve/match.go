package resolve

import (
	"sort"
	"strings"
)

// Match is one ranked candidate returned by Resolve: a canonical facts
// definition, the catalog it came from, and a confidence Score in [0,1],
// best-first. It mirrors the api.md §5 shape `Match{ def, kind, score }`.
//
//   - Def is the concrete facts record (*facts.ItemDef, *facts.NpcDef,
//     *facts.SceneryDef, *facts.BoundaryDef, *facts.SpellDef,
//     *facts.PrayerDef). Callers type-assert it to recover the full def.
//   - Kind is one of the Kind* constants identifying which catalog Def is in.
//   - Score is the recognition confidence. The constants below pin the bands:
//     a learned-alias or exact-name hit scores 1.0; fuzzy hits score lower and
//     are ordered best-first. A brain-confirmed hit scores brainScore.
//   - Canonical and ID are convenience hoists from Def so callers (and the
//     id-never-invented guarantee) don't have to type-assert just to read the
//     name/id.
//   - Source records which pipeline stage produced the match ("alias",
//     "fuzzy", or "brain"), for observability and tests.
type Match struct {
	Def       any
	Kind      string
	Score     float64
	Canonical string
	ID        int
	Source    string
}

// Match sources.
const (
	SourceAlias = "alias"
	SourceFuzzy = "fuzzy"
	SourceBrain = "brain"
)

// Score bands. They are coarse on purpose: resolve() is a *recognition*
// faculty, not a search engine, so we want a few well-separated tiers that
// make the ranked list legible and the tests stable.
const (
	// scoreExact is a learned-alias hit or a case-fold-exact canonical match.
	scoreExact = 1.0
	// scoreBrain is a brain-confirmed mapping (just below an outright exact
	// hit so a later genuine exact match would still rank first).
	scoreBrain = 0.95
	// scoreAllTokens: every query token appears in the candidate name and the
	// candidate has no extra tokens beyond a small margin — a strong fuzzy hit
	// ("rune 2 handed" → "Rune 2-handed Sword").
	scoreAllTokensTight = 0.85
	// scoreAllTokensLoose: every query token appears but the candidate has
	// several extra tokens (weaker, still plausible).
	scoreAllTokensLoose = 0.7
	// scoreSubstring: the whole normalized query is a substring of the
	// candidate name (e.g. "potion" inside "Attack potion").
	scoreSubstring = 0.6
	// scorePartialTokens: only some query tokens matched — the weakest tier we
	// still surface, so ambiguous input yields a ranked list rather than
	// nothing.
	scorePartialTokens = 0.4
)

// fuzzyFloor is the minimum score a fuzzy candidate must reach to be
// returned. Conservative by design (api.md §3/§5: "conservative"): we would
// rather return a short, high-quality ranked list than a long noisy one.
const fuzzyFloor = scorePartialTokens

// scoreCandidate scores how well a normalized query matches a catalog entry's
// canonical name. Returns (score, true) if the entry clears fuzzyFloor, else
// (0, false). The query is already normalized; the entry name is normalized
// here.
func scoreCandidate(query string, qTokens []string, e entry) (float64, bool) {
	name := normalize(e.Canonical)
	if name == "" {
		return 0, false
	}

	// Exact normalized equality is an exact hit.
	if name == query {
		return scoreExact, true
	}

	nameTokens := strings.Fields(name)

	// Token-subset: every query token is present in the name's token set.
	if len(qTokens) > 0 && allTokensPresent(qTokens, nameTokens) {
		extra := len(nameTokens) - len(qTokens)
		score := scoreAllTokensTight
		if extra > 1 {
			score = scoreAllTokensLoose
		}
		// A whole-word query that is also a prefix of the name reads as a
		// slightly stronger recognition; nudge it within its tier so e.g.
		// "rune 2 handed" beats an equally-tokened but less prefixy name.
		if strings.HasPrefix(name, query) {
			score += 0.02
		}
		return score, true
	}

	// Whole-query substring (handles single-token queries like "potion" that
	// appear mid-name, and multi-word queries embedded verbatim).
	if strings.Contains(name, query) {
		return scoreSubstring, true
	}

	// Partial token overlap — the weakest tier. Score scales with the
	// fraction of query tokens found, but never reaches the all-tokens tier.
	if len(qTokens) > 0 {
		hit := 0
		for _, qt := range qTokens {
			if tokenIn(qt, nameTokens) {
				hit++
			}
		}
		if hit > 0 {
			frac := float64(hit) / float64(len(qTokens))
			score := scorePartialTokens * frac
			if score >= fuzzyFloor {
				return score, true
			}
		}
	}

	return 0, false
}

// allTokensPresent reports whether every token in want appears in have.
func allTokensPresent(want, have []string) bool {
	for _, w := range want {
		if !tokenIn(w, have) {
			return false
		}
	}
	return true
}

// tokenIn reports whether token t is in the list (exact token equality).
func tokenIn(t string, list []string) bool {
	for _, x := range list {
		if x == t {
			return true
		}
	}
	return false
}

// rankFuzzy scores every entry in es against the query and returns the
// clearing matches sorted best-first. Ties (equal score) break by shorter
// canonical name first (the tighter recognition), then by the catalog's
// existing deterministic order (name, then id) which es already carries.
func rankFuzzy(query string, es []entry) []Match {
	qTokens := tokens(query)
	var out []Match
	for _, e := range es {
		score, ok := scoreCandidate(query, qTokens, e)
		if !ok {
			continue
		}
		out = append(out, Match{
			Def:       e.Def,
			Kind:      e.Kind,
			Score:     score,
			Canonical: e.Canonical,
			ID:        e.ID,
			Source:    SourceFuzzy,
		})
	}
	sortMatches(out)
	return out
}

// sortMatches orders matches best-first: highest score, then shorter
// canonical name (tighter match), then canonical name lexicographically, then
// id — fully deterministic.
func sortMatches(ms []Match) {
	sort.SliceStable(ms, func(i, j int) bool {
		if ms[i].Score != ms[j].Score {
			return ms[i].Score > ms[j].Score
		}
		li, lj := len(ms[i].Canonical), len(ms[j].Canonical)
		if li != lj {
			return li < lj
		}
		if ms[i].Canonical != ms[j].Canonical {
			return ms[i].Canonical < ms[j].Canonical
		}
		return ms[i].ID < ms[j].ID
	})
}
