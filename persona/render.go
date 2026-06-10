package persona

import (
	"fmt"
	"sort"
	"strings"
)

// Render is the deterministic, no-LLM prose-card floor. It maps the structured
// fields through a fixed trait lexicon into a second-person behavioral paragraph
// + a "Things you never forget" list. It is faithful-if-flatter and, per the
// LLM-food/CODE-food rule, leaks NO numbers, band words, or the archetype tag —
// only behavior in plain language.
//
// This is the reproducibility floor (docs/persona-compile.md §8). The best-of-N
// Opus cook is the quality upgrade on top; this keeps the system runnable with
// no API key and covers hosts cooked before the LLM pipeline exists.
func Render(p *Persona) string {
	c := &p.Cornerstone
	var b strings.Builder

	name := c.Identity.Name
	if name == "" {
		name = "this host"
	}
	fmt.Fprintf(&b, "You are %s.", name)
	if seed := strings.TrimSpace(c.Identity.Backstory); seed != "" {
		// (Render is normally used WHEN Backstory is empty; if a one-line seed is
		// present we lead with it.)
		fmt.Fprintf(&b, " %s.", strings.TrimRight(seed, "."))
	}

	// HEXACO clauses (only the ones with a clear lexicon entry).
	var clauses []string
	for _, k := range HexacoKeys() {
		if t, ok := c.Hexaco[HexacoLetter(k)]; ok {
			if phrase := hexacoPhrase(k, bucket(t.Band)); phrase != "" {
				clauses = append(clauses, phrase)
			}
		}
	}
	if len(clauses) > 0 {
		fmt.Fprintf(&b, " You are %s.", joinClauses(clauses))
	}

	// Cooperation + the dominant risk posture.
	if cp := coopPhrase(c.Prefs.CoopType); cp != "" {
		fmt.Fprintf(&b, " %s.", cp)
	}
	if rp := riskPhrase(c.Prefs.Risk); rp != "" {
		fmt.Fprintf(&b, " %s.", rp)
	}
	// A couple of the additive dials when they're pronounced.
	if bucket(c.Prefs.Aggression.Band) == "high" {
		fmt.Fprintf(&b, " You are quick to start a fight rather than wait to be provoked.")
	}
	if bucket(c.Prefs.Patience.Band) == "high" {
		fmt.Fprintf(&b, " You play the long game and rarely abandon a goal.")
	} else if bucket(c.Prefs.Patience.Band) == "low" {
		fmt.Fprintf(&b, " You chase whatever interests you now and hop goals readily.")
	}

	// Curiosity — what PULLS the host (the explore drive). Surfaced here so the
	// prose card the brain reads actually carries the dial, instead of leaving
	// Curiosity decorative. (Director-level explore/exploit weighting is Phase 2.)
	if cp := curiosityPhrase(c.Prefs.Curiosity); cp != "" {
		fmt.Fprintf(&b, " %s.", cp)
	}

	// North star (the motivator, in the host's words if statement is set).
	if ns := strings.TrimSpace(c.Identity.NorthStar.Statement); ns != "" {
		fmt.Fprintf(&b, " Above all: %s.", strings.TrimRight(ns, "."))
	}

	// Voice.
	if vp := voicePhrase(c.Identity.Voice); vp != "" {
		fmt.Fprintf(&b, " %s.", vp)
	}

	// Pinned memories.
	if len(c.Pinned) > 0 {
		b.WriteString("\n\nThings you never forget:")
		for _, m := range c.Pinned {
			if s := strings.TrimSpace(m.Summary); s != "" {
				fmt.Fprintf(&b, "\n- %s", s)
			}
		}
	}

	return b.String()
}

// bucket collapses the 6-band ladder to low/mid/high for the flat floor lexicon.
func bucket(b Band) string {
	switch b {
	case BandVeryLow, BandLow:
		return "low"
	case BandMid, BandMidHigh:
		return "mid"
	case BandHigh, BandVeryHigh:
		return "high"
	default:
		return ""
	}
}

func hexacoPhrase(k HexacoKey, bkt string) string {
	lex := map[HexacoKey]map[string]string{
		Honesty: {
			"low":  "willing to bend the truth or exploit an opening",
			"high": "scrupulously fair and sincere",
		},
		Emotionality: {
			"low":  "calm under pressure and slow to fear",
			"high": "easily rattled and quick to retreat from danger",
		},
		Extraversion: {
			"low":  "reserved and happiest on your own",
			"high": "outgoing and quick to strike up a conversation",
		},
		Agreeableness: {
			"low":  "stubborn, slow to forgive, and willing to push back hard",
			"high": "forgiving and easygoing",
		},
		Conscientiousness: {
			"low":  "impulsive and a little sloppy",
			"high": "diligent and methodical",
		},
		Openness: {
			"low":  "happiest with the familiar and set in your ways",
			"high": "curious and drawn to new things",
		},
	}
	if m, ok := lex[k]; ok {
		return m[bkt] // "mid" returns "" — omit middling traits to keep the card sharp
	}
	return ""
}

func coopPhrase(ct CoopType) string {
	switch ct {
	case Altruist:
		return "You tend to help others even at a cost to yourself"
	case FreeRider:
		return "You look out for yourself first and will take a lopsided deal in your favor"
	case ConditionalCooperator:
		return "You cooperate with those who cooperate with you, and remember those who don't"
	}
	return ""
}

func riskPhrase(r DomainRisk) string {
	var parts []string
	if bucket(r.Economic) == "low" {
		parts = append(parts, "wary with your money and your trades")
	} else if bucket(r.Economic) == "high" {
		parts = append(parts, "happy to gamble and take risky deals")
	}
	if bucket(r.Bodily) == "low" {
		parts = append(parts, "careful to avoid danger and stronger opponents")
	} else if bucket(r.Bodily) == "high" {
		parts = append(parts, "drawn to danger and unafraid of a stronger foe")
	}
	if len(parts) == 0 {
		return ""
	}
	return "You are " + joinClauses(parts)
}

// curiosityPhrase surfaces the host's dominant curiosity pulls as plain behaviour
// (no numbers/bands — leak-free), so the prose card carries the explore drive
// instead of leaving Curiosity decorative. Only pronounced flavours are named,
// keeping the card sharp; middling curiosity is omitted.
func curiosityPhrase(cu Curiosity) string {
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
	sort.SliceStable(flavors, func(i, j int) bool { return flavors[i].v > flavors[j].v })
	var picked []string
	for _, f := range flavors {
		if f.v >= 0.5 {
			picked = append(picked, f.s)
		}
	}
	if len(picked) == 0 && flavors[0].v >= 0.35 {
		picked = append(picked, flavors[0].s) // no strong pull, but name the clear leader
	}
	if len(picked) > 2 {
		picked = picked[:2] // keep the card sharp
	}
	if len(picked) == 0 {
		return ""
	}
	return "You are drawn to " + joinClauses(picked)
}

func voicePhrase(v Voice) string {
	reg := strings.TrimSpace(v.Register)
	if reg == "" {
		reg = string(v.Formality)
	}
	if reg == "" {
		return ""
	}
	s := fmt.Sprintf("You speak in a %s way", reg)
	switch v.TypoFeel {
	case TypoOccasional, TypoFrequent:
		s += ", with the casual typos of someone typing quickly"
	}
	if len(v.Tics) > 0 {
		s += fmt.Sprintf(`, and you often say "%s"`, strings.Join(v.Tics, `", "`))
	}
	return s
}

// joinClauses renders a list as "a, b, and c".
func joinClauses(items []string) string {
	items = nonEmpty(items)
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func nonEmpty(in []string) []string {
	out := in[:0:0]
	for _, s := range in {
		if strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}
