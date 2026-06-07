// Package persona is the host's identity: the stored Persona (Cornerstone +
// Trajectory) and the enum vocabulary an author writes in. The brain never sees
// this struct — it sees the compiled prose card (see Render / PersonaCard). The
// cradle's deterministic policies key off the bands/enums here.
//
// Authoring rule: humans express intent in WORDS from enums.go (never numbers);
// the sampler turns a word into a sampled value within that band. Cohort is NOT
// here — it is out-of-band generation/experiment tracking (GenerationMeta),
// invisible to the host.
package persona

import "time"

// CurrentSchemaVersion guards doc/shape drift; bump on a breaking change and
// migrate stored blobs.
const CurrentSchemaVersion = 1

// Persona is the stored shape (mirrors mesa bots.persona JSONB).
type Persona struct {
	SchemaVersion int         `json:"schema_version"`
	Cornerstone   Cornerstone `json:"cornerstone"` // immutable, sealed
	Trajectory    Trajectory  `json:"trajectory"`  // mutable snapshot
}

// Cornerstone is the immutable, operator-only persona core.
type Cornerstone struct {
	Identity Identity         `json:"identity"`
	Hexaco   map[string]Trait `json:"hexaco"` // keys H,E,X,A,C,O
	Values   Values           `json:"values"`
	Prefs    Prefs            `json:"prefs"` // decision + temperament + additive dials
	// Directives are the inviolable core rules. Hard directives compile to
	// max-salience pearl vetoes (the Guard); soft ones to strong biases.
	Directives []Directive          `json:"directives,omitempty"`
	Quirks     []Quirk              `json:"quirks"`
	Reverie    ReverieSeed          `json:"reverie"`
	Pinned     []FoundationalMemory `json:"pinned"`
	Gen        GenerationMeta       `json:"generation_meta"`
}

// Trait is a band-valued dial: an authored Band (compile input) plus the
// sampled native-range value. Mu is 0..1 for most dials; loss_aversion uses
// ~1..3 (λ). Reused for HEXACO and every additive disposition dial.
type Trait struct {
	Mu   float64 `json:"mu"`
	Band Band    `json:"band"`
}

type Identity struct {
	Name         string    `json:"name"`      // == opts.Username (partition key)
	Backstory    string    `json:"backstory"` // sealed prose card — the only thing the brain reads
	NorthStar    NorthStar `json:"north_star"`
	Voice        Voice     `json:"voice"`
	ArchetypeTag string    `json:"archetype_tag"` // debug-only; INVISIBLE to the brain
}

type Voice struct {
	Register  string         `json:"register"` // FREE TEXT (open-ended slang/era flavor)
	Formality VoiceFormality `json:"formality"`
	Tics      []string       `json:"tics"`
	TypoFeel  TypoFeel       `json:"typo_feel"`
}

type NorthStar struct {
	Theme          NorthStarTheme `json:"theme"`
	Statement      string         `json:"statement"`
	Horizon        string         `json:"horizon"` // week|month|open
	SuccessSignals []string       `json:"success_signals"`
}

type Values struct {
	NorthStarValue SchwartzValue `json:"north_star_value"`
	SecondaryValue SchwartzValue `json:"secondary_value"`
}

type DomainRisk struct {
	Economic Band `json:"economic"`
	Bodily   Band `json:"bodily"`
	Social   Band `json:"social"`
}

// Prefs holds the decision/temperament dials and the additive (Westworld-
// inspired) disposition dials. Most are band-valued Traits; CoopType is an enum,
// Curiosity is a flavor vector, Risk is domain-split.
type Prefs struct {
	// econ / decision anchors
	Patience     Trait      `json:"patience"`      // long horizon (patience_tau)
	LossAversion Trait      `json:"loss_aversion"` // λ (Mu ~1..3)
	CoopType     CoopType   `json:"coop_type"`
	Risk         DomainRisk `json:"risk"`

	// temperament
	Attention AttentionAnchor `json:"attention"` // short-horizon focus
	Curiosity Curiosity       `json:"curiosity"` // 5-flavor vector

	// additive disposition dials (Westworld-inspired; bind to RSC mechanics)
	Aggression       Trait  `json:"aggression"`        // offensive initiation / PK
	Decisiveness     Trait  `json:"decisiveness"`      // commit vs deliberate (escalation threshold)
	Tenacity         Trait  `json:"tenacity"`          // retry-after-setback resilience
	SelfPreservation *Trait `json:"self_preservation"` // nil ⇒ derived from LossAversion + Risk.Bodily
	BulkApperception Trait  `json:"bulk_apperception"` // cognitive maturity / model-tier / learning rate
}

// AttentionAnchor is the short-horizon focus dial — its own 5-step ladder
// (very_distractible..hyperfocus), distinct from the magnitude Band ladder.
type AttentionAnchor struct {
	Anchor float64        `json:"anchor"` // 0..1 sampled
	Level  AttentionLevel `json:"level"`  // authored word
}

// Curiosity is the flavor vector (WHAT pulls the host); the sampler builds it
// from the authored primary/secondary flavors. The reverie engine reads it.
type Curiosity struct {
	Social   float64 `json:"social"`
	Spatial  float64 `json:"spatial"`
	Skill    float64 `json:"skill"`
	Economic float64 `json:"economic"`
	Risk     float64 `json:"risk"`
}

// Directive is an inviolable (or strong) behavioral rule — the persona's core.
// Hard directives compile to max-salience pearl vetoes; soft ones to biases.
// e.g. {Subject:"self", Predicate:"attack", Object:"stronger_player", Hard:true}
// → "never attack a stronger player". Predicate names the action it governs;
// Object refines the condition (a referent or modifier the compiler recognizes).
type Directive struct {
	Priority  int    `json:"priority"`
	Subject   string `json:"subject"`
	Predicate string `json:"predicate"`
	Object    string `json:"object"`
	Hard      bool   `json:"hard"`
}

// Quirk is an EXECUTABLE modifier (not flavor text), applied with no LLM call.
// The `object` referent grammar is intentionally still open (validated by shape
// only for now).
type Quirk struct {
	ID           string        `json:"id"`
	Origin       string        `json:"origin"` // derived|idiosyncratic|learned_emergent
	Domain       QuirkDomain   `json:"domain"`
	Trigger      string        `json:"trigger"` // event Kind() | on_encounter | pre_action:<verb>
	Binding      string        `json:"binding"` // gameplay binding (open grammar)
	Relation     QuirkRelation `json:"relation"`
	Object       string        `json:"object"` // referent (open grammar)
	Strength     QuirkStrength `json:"strength"`
	Observable   bool          `json:"observable"`
	SuppressWhen SuppressWhen  `json:"suppress_when"`
	Narrative    string        `json:"narrative"` // for the prose card only
}

type ReverieSeed struct {
	Jitter   map[string]float64 `json:"jitter"` // per-reverie r_a, sampled once
	DriftLog []ReverieDrift     `json:"drift_log"`
}

type ReverieDrift struct {
	ID    string    `json:"id"`
	Delta float64   `json:"delta"`
	At    time.Time `json:"at"`
}

type FoundationalMemory struct {
	Summary string  `json:"summary"`
	Weight  float64 `json:"weight"`
}

// GenerationMeta is OUT-OF-BAND tracking: which generation/experiment a host
// belongs to, for analytics/reproducibility. The host never reasons about it.
// Cohort lives HERE (not in the persona proper).
type GenerationMeta struct {
	CohortID        string    `json:"cohort_id"`        // experiment/generation grouping (analytics only)
	Archetype       string    `json:"archetype"`        // archetype id (a defined entity; see archetype registry)
	SamplerVersion  string    `json:"sampler_version"`  // reproducibility audit
	BornAt          time.Time `json:"born_at"`          // registration timestamp
	LLMMaterialized bool      `json:"llm_materialized"` // false ⇒ deterministic Render() fallback was used
}

// Trajectory is the mutable, runtime-owned snapshot (authoritative copy in mesa).
type Trajectory struct {
	Mood        Affect             `json:"mood"`
	SubGoals    []string           `json:"sub_goals"`
	SkillLogits map[string]float64 `json:"skill_logits"`
	RiskDrift   map[string]float64 `json:"risk_drift"`
	Emergent    *Quirk             `json:"learned_emergent,omitempty"`
}

type Affect struct {
	Valence    float64   `json:"valence"`    // -1..1
	Arousal    float64   `json:"arousal"`    // 0..1
	Stress     float64   `json:"stress"`     // 0..1
	Confidence float64   `json:"confidence"` // 0..1
	UpdatedAt  time.Time `json:"updated_at"`
}
