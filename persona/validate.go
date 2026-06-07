package persona

import (
	"errors"
	"fmt"
)

// Validate enforces every enum-typed field against enums.go and returns a single
// joined error listing all problems (so an author sees everything wrong at once,
// each with its valid set). It checks the SETTLED surface; the deferred pieces
// (quirk object/binding grammar, reverie catalog) are validated by shape only.
func (p *Persona) Validate() error {
	var errs []error
	if p.SchemaVersion <= 0 {
		errs = append(errs, fmt.Errorf("persona: schema_version must be >= 1, got %d", p.SchemaVersion))
	}
	c := &p.Cornerstone

	// HEXACO — all six axes present, each a valid band.
	for _, k := range HexacoKeys() {
		letter := HexacoLetter(k)
		t, ok := c.Hexaco[letter]
		if !ok {
			errs = append(errs, fmt.Errorf("persona: hexaco missing %q (%s)", letter, k))
			continue
		}
		if err := validateEnum("hexaco."+letter+".band", t.Band, Bands()); err != nil {
			errs = append(errs, err)
		}
	}

	// Values — both valid, and distinct.
	if err := validateEnum("values.north_star_value", c.Values.NorthStarValue, SchwartzValues()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("values.secondary_value", c.Values.SecondaryValue, SchwartzValues()); err != nil {
		errs = append(errs, err)
	}
	if c.Values.NorthStarValue != "" && c.Values.NorthStarValue == c.Values.SecondaryValue {
		errs = append(errs, fmt.Errorf("persona: values.secondary_value must differ from north_star_value"))
	}

	// Identity / north star / voice.
	if err := validateEnum("identity.north_star.theme", c.Identity.NorthStar.Theme, NorthStarThemes()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("identity.voice.formality", c.Identity.Voice.Formality, VoiceFormalities()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("identity.voice.typo_feel", c.Identity.Voice.TypoFeel, TypoFeels()); err != nil {
		errs = append(errs, err)
	}
	// voice.register is FREE TEXT — not validated.

	// Prefs — band dials, coop type, domain risk, attention level.
	pr := &c.Prefs
	bandDials := []struct {
		name string
		t    Trait
	}{
		{"prefs.patience", pr.Patience},
		{"prefs.loss_aversion", pr.LossAversion},
		{"prefs.aggression", pr.Aggression},
		{"prefs.decisiveness", pr.Decisiveness},
		{"prefs.tenacity", pr.Tenacity},
		{"prefs.bulk_apperception", pr.BulkApperception},
	}
	for _, d := range bandDials {
		if err := validateEnum(d.name+".band", d.t.Band, Bands()); err != nil {
			errs = append(errs, err)
		}
	}
	if pr.SelfPreservation != nil { // optional; only validated when authored
		if err := validateEnum("prefs.self_preservation.band", pr.SelfPreservation.Band, Bands()); err != nil {
			errs = append(errs, err)
		}
	}
	if err := validateEnum("prefs.coop_type", pr.CoopType, CoopTypes()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("prefs.risk.economic", pr.Risk.Economic, Bands()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("prefs.risk.bodily", pr.Risk.Bodily, Bands()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("prefs.risk.social", pr.Risk.Social, Bands()); err != nil {
		errs = append(errs, err)
	}
	if err := validateEnum("prefs.attention.level", pr.Attention.Level, AttentionLevels()); err != nil {
		errs = append(errs, err)
	}

	// Quirks — closed sub-enums valid; trigger well-formed (object/binding grammar deferred).
	for i := range c.Quirks {
		q := &c.Quirks[i]
		pfx := fmt.Sprintf("quirks[%d]", i)
		if err := validateEnum(pfx+".domain", q.Domain, QuirkDomains()); err != nil {
			errs = append(errs, err)
		}
		if err := validateEnum(pfx+".relation", q.Relation, QuirkRelations()); err != nil {
			errs = append(errs, err)
		}
		if err := validateEnum(pfx+".strength", q.Strength, QuirkStrengths()); err != nil {
			errs = append(errs, err)
		}
		if err := validateEnum(pfx+".suppress_when", q.SuppressWhen, SuppressWhens()); err != nil {
			errs = append(errs, err)
		}
		if !ValidTriggerForm(q.Trigger) {
			errs = append(errs, fmt.Errorf("persona: %s.trigger %q is malformed (want an event kind, on_encounter, or pre_action:<verb>)", pfx, q.Trigger))
		}
	}

	return errors.Join(errs...)
}
