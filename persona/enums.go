package persona

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// enums.go is the SINGLE SOURCE OF TRUTH for every enum-typed field in the
// persona authoring + storage surface. The golden rule (persona-authoring.md):
// humans express intent in WORDS from these closed sets, never numbers; the
// sampler turns a word into a sampled value inside that band. Every enum here is
// validated with a clear error that lists the valid set, so a typo is caught at
// author time, not silently downstream.
//
// SETTLED here. Deferred elsewhere: the quirk `object` referent grammar and the
// reverie catalog (intentionally left open — iterate later), and the exact
// band→number cut-points + band→policy mapping (the band→policy design session).

// --- the band ladder: the 6-step magnitude scale every 0..1 dial authors in ---

// Band is the ordered magnitude ladder. ALL human-authored magnitude dials use
// it (HEXACO traits, patience, loss_aversion, risk.*, and the additive dials
// aggression/decisiveness/tenacity/self_preservation/bulk_apperception).
type Band string

const (
	BandVeryLow  Band = "very_low"
	BandLow      Band = "low"
	BandMid      Band = "mid"
	BandMidHigh  Band = "mid_high"
	BandHigh     Band = "high"
	BandVeryHigh Band = "very_high"
)

// Bands returns the ladder in ascending order.
func Bands() []Band {
	return []Band{BandVeryLow, BandLow, BandMid, BandMidHigh, BandHigh, BandVeryHigh}
}

// Ordinal returns the band's position 0..5 (−1 if invalid). Lets the band→number
// mapping (the deferred design session) interpolate without re-parsing strings.
func (b Band) Ordinal() int {
	for i, x := range Bands() {
		if x == b {
			return i
		}
	}
	return -1
}

func (b Band) Valid() bool { return b.Ordinal() >= 0 }

// --- HEXACO ------------------------------------------------------------------

// HexacoKey is one of the six HEXACO trait axes.
type HexacoKey string

const (
	Honesty           HexacoKey = "honesty"           // H — fair/sincere vs exploitative
	Emotionality      HexacoKey = "emotionality"      // E — anxious/flees vs brave
	Extraversion      HexacoKey = "extraversion"      // X — chatty vs solitary
	Agreeableness     HexacoKey = "agreeableness"     // A — forgiving vs retaliatory
	Conscientiousness HexacoKey = "conscientiousness" // C — diligent vs impulsive
	Openness          HexacoKey = "openness"          // O — curious vs conventional
)

// HexacoKeys returns the six axes (stored under the short letters H,E,X,A,C,O).
func HexacoKeys() []HexacoKey {
	return []HexacoKey{Honesty, Emotionality, Extraversion, Agreeableness, Conscientiousness, Openness}
}

// HexacoLetter maps an axis to its stored single-letter key.
func HexacoLetter(k HexacoKey) string {
	switch k {
	case Honesty:
		return "H"
	case Emotionality:
		return "E"
	case Extraversion:
		return "X"
	case Agreeableness:
		return "A"
	case Conscientiousness:
		return "C"
	case Openness:
		return "O"
	}
	return ""
}

// --- additive disposition dials (Westworld-inspired; genuinely additive) -----
//
// These name behavior axes HEXACO+econ is thin on and that bind to real RSC
// mechanics. They are band-valued like any other dial. (Most other Westworld
// attributes collapse into HEXACO/econ/voice and are NOT separate fields.)

// DispositionDial is a band-valued behavioral dial beyond HEXACO.
type DispositionDial string

const (
	// Aggression — propensity for OFFENSIVE initiation (start fights / PK),
	// distinct from bodily risk (danger tolerance).
	Aggression DispositionDial = "aggression"
	// Decisiveness — commit on incomplete info vs deliberate; maps to the
	// escalation/confidence threshold (act locally vs phone mesa).
	Decisiveness DispositionDial = "decisiveness"
	// Tenacity — retry-after-setback resilience (distinct from patience's long horizon).
	Tenacity DispositionDial = "tenacity"
	// SelfPreservation — flee threshold. OPTIONAL human override; when unset it
	// is DERIVED from loss_aversion + bodily risk. Only authored to deviate.
	SelfPreservation DispositionDial = "self_preservation"
	// BulkApperception — "intelligence as learning-from-experience" (Westworld).
	// The cognitive maturity / model-tier / learning-rate axis behind the
	// host's local↔mesa reliance (the maturity dial).
	BulkApperception DispositionDial = "bulk_apperception"
)

// DispositionDials returns the additive dials. SelfPreservation is optional
// (derived if omitted); the rest are first-class sampled dials.
func DispositionDials() []DispositionDial {
	return []DispositionDial{Aggression, Decisiveness, Tenacity, SelfPreservation, BulkApperception}
}

// --- north star --------------------------------------------------------------

type NorthStarTheme string

const (
	ThemeWealth        NorthStarTheme = "wealth"
	ThemeSkillMastery  NorthStarTheme = "skill_mastery"
	ThemeSocial        NorthStarTheme = "social"
	ThemeExploration   NorthStarTheme = "exploration"
	ThemeReputation    NorthStarTheme = "reputation"
	ThemeCombat        NorthStarTheme = "combat"
	ThemeBroadAmbition NorthStarTheme = "broad_ambition"
)

func NorthStarThemes() []NorthStarTheme {
	return []NorthStarTheme{ThemeWealth, ThemeSkillMastery, ThemeSocial, ThemeExploration, ThemeReputation, ThemeCombat, ThemeBroadAmbition}
}

// --- Schwartz-10 values ------------------------------------------------------

type SchwartzValue string

const (
	SelfDirection SchwartzValue = "self_direction"
	Stimulation   SchwartzValue = "stimulation"
	Hedonism      SchwartzValue = "hedonism"
	Achievement   SchwartzValue = "achievement"
	Power         SchwartzValue = "power"
	Security      SchwartzValue = "security"
	Conformity    SchwartzValue = "conformity"
	Tradition     SchwartzValue = "tradition"
	Benevolence   SchwartzValue = "benevolence"
	Universalism  SchwartzValue = "universalism"
)

func SchwartzValues() []SchwartzValue {
	return []SchwartzValue{SelfDirection, Stimulation, Hedonism, Achievement, Power, Security, Conformity, Tradition, Benevolence, Universalism}
}

// --- voice -------------------------------------------------------------------
// register is intentionally FREE TEXT (open-ended slang/era flavor); the rest
// are closed sets.

type VoiceFormality string

const (
	FormalityFormal    VoiceFormality = "formal"
	FormalityNeutral   VoiceFormality = "neutral"
	FormalityCasual    VoiceFormality = "casual"
	FormalityTextSpeak VoiceFormality = "text_speak"
)

func VoiceFormalities() []VoiceFormality {
	return []VoiceFormality{FormalityFormal, FormalityNeutral, FormalityCasual, FormalityTextSpeak}
}

type TypoFeel string

const (
	TypoNone       TypoFeel = "none"
	TypoRare       TypoFeel = "rare"
	TypoOccasional TypoFeel = "occasional"
	TypoFrequent   TypoFeel = "frequent"
)

func TypoFeels() []TypoFeel {
	return []TypoFeel{TypoNone, TypoRare, TypoOccasional, TypoFrequent}
}

// VoiceRegisterSuggestions are common values for the FREE-TEXT register field —
// hints for an author, NOT an enforced set.
func VoiceRegisterSuggestions() []string {
	return []string{"formal", "casual", "terse", "playful", "gruff"}
}

// --- temperament: attention (its own 5-step ladder) + curiosity flavors ------

// AttentionLevel is the SHORT-horizon focus ladder (how hard the host resists
// being pulled off task). Distinct from the magnitude Band ladder and from
// patience (long horizon).
type AttentionLevel string

const (
	VeryDistractible AttentionLevel = "very_distractible"
	Distractible     AttentionLevel = "distractible"
	Balanced         AttentionLevel = "balanced"
	Focused          AttentionLevel = "focused"
	Hyperfocus       AttentionLevel = "hyperfocus"
)

func AttentionLevels() []AttentionLevel {
	return []AttentionLevel{VeryDistractible, Distractible, Balanced, Focused, Hyperfocus}
}

// CuriosityFlavor is a direction of the curiosity flavor vector — WHAT pulls the
// host. The human authors 1-2 flavors (words); the sampler builds the float
// vector; the reverie engine reads it.
type CuriosityFlavor string

const (
	CuriositySocial   CuriosityFlavor = "social"
	CuriositySpatial  CuriosityFlavor = "spatial"
	CuriositySkill    CuriosityFlavor = "skill"
	CuriosityEconomic CuriosityFlavor = "economic"
	CuriosityRisk     CuriosityFlavor = "risk"
)

func CuriosityFlavors() []CuriosityFlavor {
	return []CuriosityFlavor{CuriositySocial, CuriositySpatial, CuriositySkill, CuriosityEconomic, CuriosityRisk}
}

// --- cooperation type --------------------------------------------------------

type CoopType string

const (
	ConditionalCooperator CoopType = "conditional_cooperator"
	FreeRider             CoopType = "free_rider"
	Altruist              CoopType = "altruist"
)

func CoopTypes() []CoopType {
	return []CoopType{ConditionalCooperator, FreeRider, Altruist}
}

// --- quirks (the closed sub-enums; the `object` referent grammar is DEFERRED) -

type QuirkDomain string

const (
	DomainMovement QuirkDomain = "movement"
	DomainSocial   QuirkDomain = "social"
	DomainTrade    QuirkDomain = "trade"
	DomainCombat   QuirkDomain = "combat"
	DomainBanking  QuirkDomain = "banking"
	DomainIdle     QuirkDomain = "idle"
)

func QuirkDomains() []QuirkDomain {
	return []QuirkDomain{DomainMovement, DomainSocial, DomainTrade, DomainCombat, DomainBanking, DomainIdle}
}

type QuirkRelation string

const (
	Prefers   QuirkRelation = "prefers"
	Avoids    QuirkRelation = "avoids"
	Distrusts QuirkRelation = "distrusts"
	Delays    QuirkRelation = "delays"
)

func QuirkRelations() []QuirkRelation {
	return []QuirkRelation{Prefers, Avoids, Distrusts, Delays}
}

type QuirkStrength string

const (
	StrengthMild     QuirkStrength = "mild"
	StrengthModerate QuirkStrength = "moderate"
	StrengthStrong   QuirkStrength = "strong"
)

func QuirkStrengths() []QuirkStrength {
	return []QuirkStrength{StrengthMild, StrengthModerate, StrengthStrong}
}

type SuppressWhen string

const (
	SuppressInCombat SuppressWhen = "in_combat"
	SuppressLowHP    SuppressWhen = "low_hp"
	SuppressNone     SuppressWhen = "none"
)

func SuppressWhens() []SuppressWhen {
	return []SuppressWhen{SuppressInCombat, SuppressLowHP, SuppressNone}
}

// QuirkTrigger reports whether s is a well-FORMED trigger. Full enforcement of
// the event-kind / verb sets is deferred with the rest of the quirk grammar; we
// validate shape only: a bare event kind, "on_encounter", or "pre_action:<verb>".
func ValidTriggerForm(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	if s == "on_encounter" {
		return true
	}
	if v, ok := strings.CutPrefix(s, "pre_action:"); ok {
		return strings.TrimSpace(v) != ""
	}
	// otherwise treated as a bus event kind (e.g. "trade_request"); not yet
	// checked against the event registry.
	return true
}

// --- generic validation helper ----------------------------------------------

// enumError builds a "got X; valid: a, b, c" error for a field.
func enumError(field, got string, valid []string) error {
	sorted := append([]string(nil), valid...)
	sort.Strings(sorted)
	return fmt.Errorf("persona: %s = %q; valid: %s", field, got, strings.Join(sorted, ", "))
}

// validate checks a value against a set and returns a listing error on miss.
func validateEnum[T ~string](field string, v T, valid []T) error {
	if slices.Contains(valid, v) {
		return nil
	}
	strs := make([]string, len(valid))
	for i, x := range valid {
		strs[i] = string(x)
	}
	return enumError(field, string(v), strs)
}
