package persona

import (
	"encoding/json"
	"strings"
	"testing"
)

// validPersona builds a minimal, valid stored persona for tests.
func validPersona() *Persona {
	band := func(b Band) Trait { return Trait{Band: b} } // Mu unset ⇒ dial() uses the band center
	return &Persona{
		SchemaVersion: CurrentSchemaVersion,
		Cornerstone: Cornerstone{
			Identity: Identity{
				Name:         "marn_fishes",
				ArchetypeTag: "cautious_social_grinder",
				NorthStar:    NorthStar{Theme: ThemeSkillMastery, Statement: "99 fishing, never scammed", Horizon: "open"},
				Voice:        Voice{Register: "swamp casual", Formality: FormalityCasual, TypoFeel: TypoOccasional},
			},
			Hexaco: map[string]Trait{
				"H": band(BandHigh), "E": band(BandMidHigh), "X": band(BandHigh),
				"A": band(BandMidHigh), "C": band(BandHigh), "O": band(BandMid),
			},
			Values: Values{NorthStarValue: Achievement, SecondaryValue: Benevolence},
			Prefs: Prefs{
				Patience:     band(BandVeryHigh),
				LossAversion: Trait{Mu: 2.6, Band: BandHigh},
				CoopType:     ConditionalCooperator,
				Risk:         DomainRisk{Economic: BandLow, Bodily: BandLow, Social: BandMid},
				Attention:    AttentionAnchor{Anchor: 0.75, Level: Focused},
				Curiosity:    Curiosity{Skill: 0.6, Social: 0.25, Economic: 0.1, Spatial: 0.05},
				Aggression:   band(BandLow),
				Decisiveness: band(BandMid),
				Tenacity:     band(BandHigh),
				// SelfPreservation omitted ⇒ derived from loss_aversion + bodily risk
				BulkApperception: band(BandMid),
			},
			Quirks: []Quirk{{
				ID: "trade_screen", Origin: "idiosyncratic", Domain: DomainTrade,
				Trigger: "trade_request", Relation: Distrusts, Object: "player_type:stranger",
				Strength: StrengthModerate, Observable: true, SuppressWhen: SuppressNone,
			}},
			Gen: GenerationMeta{CohortID: "lumbridge_regulars", Archetype: "cautious_social_grinder", SamplerVersion: "genpop-v1"},
		},
	}
}

func TestValidPersonaPasses(t *testing.T) {
	if err := validPersona().Validate(); err != nil {
		t.Fatalf("valid persona failed validation: %v", err)
	}
}

func TestSelfPreservationOptional(t *testing.T) {
	p := validPersona()
	if p.Cornerstone.Prefs.SelfPreservation != nil {
		t.Fatal("self_preservation should default to nil (derived)")
	}
	// Authoring it explicitly is allowed and validated.
	p.Cornerstone.Prefs.SelfPreservation = &Trait{Mu: 0.5, Band: BandHigh}
	if err := p.Validate(); err != nil {
		t.Fatalf("explicit self_preservation should validate: %v", err)
	}
}

func TestValidateCatchesBadEnums(t *testing.T) {
	p := validPersona()
	p.Cornerstone.Prefs.Aggression.Band = "foccused"     // typo'd band
	p.Cornerstone.Values.SecondaryValue = Achievement    // duplicate of north_star_value
	p.Cornerstone.Identity.NorthStar.Theme = "money"     // not a theme
	p.Cornerstone.Prefs.Attention.Level = "very_focused" // not an attention level
	delete(p.Cornerstone.Hexaco, "O")                    // missing axis

	err := p.Validate()
	if err == nil {
		t.Fatal("expected validation errors")
	}
	msg := err.Error()
	for _, want := range []string{"aggression", "secondary_value", "north_star.theme", "attention.level", "missing \"O\""} {
		if !strings.Contains(msg, want) {
			t.Errorf("validation error missing %q; got:\n%s", want, msg)
		}
	}
}

func TestCohortIsNotInCornerstone(t *testing.T) {
	// Cohort is out-of-band tracking — it must live in GenerationMeta, not as a
	// persona field the host reasons about. This guards that it stays there.
	p := validPersona()
	b, err := json.Marshal(p.Cornerstone)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), "lumbridge_regulars") {
		// only acceptable inside generation_meta, which IS part of cornerstone json
		if !strings.Contains(string(b), "generation_meta") {
			t.Fatal("cohort leaked outside generation_meta")
		}
	}
}

func TestJSONRoundTrip(t *testing.T) {
	p := validPersona()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Persona
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if err := got.Validate(); err != nil {
		t.Fatalf("round-tripped persona failed validation: %v", err)
	}
	if got.Cornerstone.Prefs.LossAversion.Mu != 2.6 {
		t.Fatalf("loss_aversion mu lost in round-trip: %v", got.Cornerstone.Prefs.LossAversion.Mu)
	}
}
