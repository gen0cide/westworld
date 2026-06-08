// Command dronegen emits valid drone persona files for the cradle/mesa fleet.
// Every drone shares the Varrock-market core — they follow their operator (Alex)
// as a soft, strong bias, and explore Gielinor when he's silent — and is dealt
// one of a few sub-archetypes (loudmouth / quiet tagalong / haggler / wide-eyed
// tourist) round-robin, so a crowd reads as a believable mix rather than clones.
//
// Names are drone<N> so they line up with the OpenRSC accounts AND the mesa
// host_ids (name == host_id == username == persona key). Each file is validated
// before it's written.
//
//	dronegen -n 3 -out ./drones                 # the few-host test set
//	dronegen -n 200 -out ./drones -operator alex
//	mesa-ctl persona import ./drones/
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gen0cide/westworld/persona"
)

func band(b persona.Band) persona.Trait { return persona.Trait{Band: b} }

// core builds the shared Varrock-market drone: very social, text-speak, unbothered
// on the road, and bound to the operator's will (soft deference) with exploration
// as the fallback north-star.
func core(name, archetype string) persona.Persona {
	return persona.Persona{
		SchemaVersion: persona.CurrentSchemaVersion,
		Cornerstone: persona.Cornerstone{
			Identity: persona.Identity{
				Name:         name,
				ArchetypeTag: archetype,
				NorthStar: persona.NorthStar{
					Theme:          persona.ThemeExploration,
					Statement:      "Follow Alex's lead; when he is silent, explore the world and see new places — keep moving and doing, chatting only in passing.",
					Horizon:        "open",
					SuccessSignals: []string{"did what Alex asked", "saw somewhere new", "covered ground"},
				},
				Voice: persona.Voice{Register: "varrock street", Formality: persona.FormalityTextSpeak, TypoFeel: persona.TypoFrequent},
			},
			Hexaco: map[string]persona.Trait{
				"H": band(persona.BandMidHigh),
				"E": band(persona.BandMid),
				"X": band(persona.BandHigh),     // social, but not relentless
				"A": band(persona.BandHigh),     // agreeable, follows along
				"C": band(persona.BandMidHigh),  // follows through on doing, not just talking
				"O": band(persona.BandHigh),     // curious explorer
			},
			Values: persona.Values{NorthStarValue: persona.Conformity, SecondaryValue: persona.Stimulation},
			Prefs: persona.Prefs{
				Patience:         band(persona.BandMid),
				LossAversion:     persona.Trait{Mu: 2.0, Band: persona.BandMid},
				CoopType:         persona.ConditionalCooperator,
				Risk:             persona.DomainRisk{Economic: persona.BandMid, Bodily: persona.BandMid, Social: persona.BandMidHigh}, // unbothered + socially bold
				Attention:        persona.AttentionAnchor{Anchor: 0.6, Level: persona.Balanced},                                     // stays on task; doesn't get stuck chatting
				Curiosity:        persona.Curiosity{Social: 0.3, Spatial: 0.55, Skill: 0.1, Economic: 0.0, Risk: 0.05},            // exploration is the dominant pull
				Aggression:       band(persona.BandLow),
				Decisiveness:     band(persona.BandMid),
				Tenacity:         band(persona.BandMid),
				BulkApperception: band(persona.BandMid),
			},
			// Soft, high-priority deference to the operator (Alex). Soft ⇒ a strong
			// bias, not a pearl veto. His live direction also arrives via ANALYSIS
			// mode (the !<name> operator override); this is the standing disposition.
			Directives: []persona.Directive{
				{Priority: 10, Subject: "self", Predicate: "obey", Object: "operator", Hard: false},
			},
			Gen: persona.GenerationMeta{CohortID: "varrock_market_drones", Archetype: archetype, SamplerVersion: "dronegen-v1"},
		},
	}
}

// archetype is one sub-flavor: a tag + a mutation on the shared core.
type archetype struct {
	tag string
	mut func(*persona.Persona)
}

var archetypes = []archetype{
	{"loudmouth", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["H"] = band(persona.BandMid)      // brasher
		p.Cornerstone.Hexaco["X"] = band(persona.BandVeryHigh) // the loud one — still extra-chatty
		p.Cornerstone.Prefs.Aggression = band(persona.BandMid)
		p.Cornerstone.Prefs.Curiosity.Social = 0.4 // chatty, but exploration still leads
		p.Cornerstone.Identity.Voice.Tics = []string{"lol", "ayy", "ez"}
	}},
	{"quiet_tagalong", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["X"] = band(persona.BandHigh) // social but follows
		p.Cornerstone.Hexaco["A"] = band(persona.BandVeryHigh)
		p.Cornerstone.Prefs.Decisiveness = band(persona.BandLow) // defers
		sp := band(persona.BandHigh)
		p.Cornerstone.Prefs.SelfPreservation = &sp
		p.Cornerstone.Identity.Voice.TypoFeel = persona.TypoOccasional
	}},
	{"haggler", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Economic = 0.4
		p.Cornerstone.Prefs.Curiosity.Social = 0.3
		p.Cornerstone.Quirks = []persona.Quirk{{
			ID: "haggle_distrust", Origin: "idiosyncratic", Domain: persona.DomainTrade,
			Trigger: "trade_request", Relation: persona.Distrusts, Object: "player_type:stranger",
			Strength: persona.StrengthModerate, Observable: true, SuppressWhen: persona.SuppressNone,
			Narrative: "sizes up every trade like it's a scam until proven otherwise",
		}}
	}},
	{"wide_eyed_tourist", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["O"] = band(persona.BandVeryHigh)
		p.Cornerstone.Prefs.Curiosity.Spatial = 0.65 // gawks at everything — exploration-forward
		p.Cornerstone.Prefs.Curiosity.Social = 0.2
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.45, Level: persona.Distractible}
	}},
}

// buildDrone constructs the persona for drone<idx> (1-based), dealing an archetype
// round-robin, and validates it.
func buildDrone(idx int) (persona.Persona, error) {
	a := archetypes[(idx-1)%len(archetypes)]
	p := core(fmt.Sprintf("drone%d", idx), a.tag)
	a.mut(&p)
	if err := p.Validate(); err != nil {
		return p, err
	}
	return p, nil
}

func main() {
	n := flag.Int("n", 3, "how many drones to generate")
	start := flag.Int("start", 1, "first drone index (drone<start>..)")
	out := flag.String("out", "./drones", "output directory for <name>.json files")
	flag.Parse()

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "dronegen:", err)
		os.Exit(1)
	}
	counts := map[string]int{}
	for i := 0; i < *n; i++ {
		idx := *start + i
		p, err := buildDrone(idx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "dronegen: drone%d invalid: %v\n", idx, err)
			os.Exit(1)
		}
		raw, err := json.MarshalIndent(p, "", "  ")
		if err != nil {
			fmt.Fprintln(os.Stderr, "dronegen:", err)
			os.Exit(1)
		}
		path := filepath.Join(*out, p.Cornerstone.Identity.Name+".json")
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, "dronegen:", err)
			os.Exit(1)
		}
		counts[p.Cornerstone.Identity.ArchetypeTag]++
	}
	fmt.Printf("wrote %d drones to %s\n", *n, *out)
	for _, a := range archetypes {
		if counts[a.tag] > 0 {
			fmt.Printf("  %-18s %d\n", a.tag, counts[a.tag])
		}
	}
}
