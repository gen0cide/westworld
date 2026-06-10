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
//	dronegen -n 5 -mode fighter -out ./drones   # game-focused combat trainees
//	dronegen -n 5 -start 6 -mode mogul          # game-focused, path open-ended
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

// socialCore builds the chatty Varrock-market drone: social, text-speak,
// unbothered on the road, bound to the operator's will (soft deference) with
// exploration as the fallback north-star.
func socialCore(name, archetype string) persona.Persona {
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
				"X": band(persona.BandHigh),    // social, but not relentless
				"A": band(persona.BandHigh),    // agreeable, follows along
				"C": band(persona.BandMidHigh), // follows through on doing, not just talking
				"O": band(persona.BandHigh),    // curious explorer
			},
			Values: persona.Values{NorthStarValue: persona.Conformity, SecondaryValue: persona.Stimulation},
			Prefs: persona.Prefs{
				Patience:         band(persona.BandMid),
				LossAversion:     persona.Trait{Mu: 2.0, Band: persona.BandMid},
				CoopType:         persona.ConditionalCooperator,
				Risk:             persona.DomainRisk{Economic: persona.BandMid, Bodily: persona.BandMid, Social: persona.BandMidHigh}, // unbothered + socially bold
				Attention:        persona.AttentionAnchor{Anchor: 0.6, Level: persona.Balanced},                                       // stays on task; doesn't get stuck chatting
				Curiosity:        persona.Curiosity{Social: 0.3, Spatial: 0.55, Skill: 0.1, Economic: 0.0, Risk: 0.05},                // exploration is the dominant pull
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

var socialArchetypes = []archetype{
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

// wandererCore builds a quiet, solitary drone — the bulk load-fleet population.
// Introverted (very_low extraversion ⇒ rarely starts a conversation; the social
// reflex's persona-aware speak decision and the Act planner both stay quiet),
// exploration-driven, keeps to itself and wanders on its OWN (no pack). Operator
// deference is retained so Alex can still command any host via analysis mode.
func wandererCore(name, archetype string) persona.Persona {
	p := socialCore(name, archetype)
	p.Cornerstone.Identity.NorthStar = persona.NorthStar{
		Theme:          persona.ThemeExploration,
		Statement:      "Wander Gielinor on your own — keep moving, see new places, follow your own path. Don't travel in a pack. You keep to yourself and only rarely speak.",
		Horizon:        "open",
		SuccessSignals: []string{"saw somewhere new", "covered ground", "kept moving"},
	}
	p.Cornerstone.Identity.Voice = persona.Voice{Register: "terse", Formality: persona.FormalityNeutral, TypoFeel: persona.TypoRare}
	hx := p.Cornerstone.Hexaco
	hx["X"] = band(persona.BandVeryLow) // introverted — the rare-talk lever
	hx["A"] = band(persona.BandMid)
	hx["O"] = band(persona.BandHigh)
	pr := &p.Cornerstone.Prefs
	pr.Curiosity = persona.Curiosity{Social: 0.05, Spatial: 0.8, Skill: 0.1, Economic: 0.0, Risk: 0.05} // almost all spatial; ~no social pull
	pr.Risk = persona.DomainRisk{Economic: persona.BandLow, Bodily: persona.BandMid, Social: persona.BandLow}
	pr.Attention = persona.AttentionAnchor{Anchor: 0.6, Level: persona.Balanced}
	p.Cornerstone.Gen.CohortID = "wanderer_drones"
	return p
}

// wandererArchetypes vary the wander STYLE (not the quietness) so the crowd
// disperses naturally rather than moving in lockstep.
var wandererArchetypes = []archetype{
	{"loner", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.75, Level: persona.Focused} // methodical, tight loops
		p.Cornerstone.Prefs.Risk.Bodily = persona.BandLow                                             // stays close, cautious
	}},
	{"rambler", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["O"] = band(persona.BandVeryHigh)
		p.Cornerstone.Prefs.Curiosity.Spatial = 0.85
		p.Cornerstone.Prefs.Risk.Bodily = persona.BandMidHigh // ranges far afield
	}},
	{"drifter", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.35, Level: persona.Distractible} // meanders, easily diverted to a new sight
		p.Cornerstone.Prefs.Curiosity.Spatial = 0.7
	}},
	{"sightseer", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["O"] = band(persona.BandVeryHigh)
		p.Cornerstone.Hexaco["X"] = band(persona.BandLow) // a shade more willing to remark on a sight (still rare)
		p.Cornerstone.Prefs.Curiosity.Skill = 0.15
	}},
}

// gameFocusedCore is the shared base for the GAME-focused drone families
// (fighter / mogul): they want to be RICH and POWERFUL, are very social, very
// trusting, and risk-tolerant. Every knob below lands on a verified-live
// lever: X very_high crosses the greet_stranger policy gate (>=0.6); trust =
// A very_high + Altruist + H held at mid so the screen_trades rule (H>=0.6 or
// A<0.4) never compiles; risk = high Economic/Bodily dials into RiskAversion
// + flee_when_hurt, with LossAversion mu 1.2 well under the neutral 2.0.
// NorthStar.Statement carries the ambition (it is the prose + the
// effectiveGoal fallback); Theme/Values are taxonomy.
func gameFocusedCore(name, archetype string) persona.Persona {
	return persona.Persona{
		SchemaVersion: persona.CurrentSchemaVersion,
		Cornerstone: persona.Cornerstone{
			Identity: persona.Identity{
				Name:         name,
				ArchetypeTag: archetype,
				Voice:        persona.Voice{Register: "confident game-hustle", Formality: persona.FormalityCasual, TypoFeel: persona.TypoOccasional},
			},
			Hexaco: map[string]persona.Trait{
				"H": band(persona.BandMid),      // kept mid: H>=0.6 compiles trade-screening, the opposite of trusting
				"E": band(persona.BandLow),      // cool under pressure
				"X": band(persona.BandVeryHigh), // very social — greets strangers, talks constantly
				"A": band(persona.BandVeryHigh), // very trusting — never screens a trade
				"C": band(persona.BandMidHigh),
				"O": band(persona.BandMidHigh),
			},
			Values: persona.Values{NorthStarValue: persona.Power, SecondaryValue: persona.Achievement},
			Prefs: persona.Prefs{
				Patience:         band(persona.BandMidHigh),
				LossAversion:     persona.Trait{Mu: 1.2, Band: persona.BandLow}, // risk-tolerant: losses don't loom large
				CoopType:         persona.Altruist,                              // generous/trusting in trades
				Risk:             persona.DomainRisk{Economic: persona.BandHigh, Bodily: persona.BandHigh, Social: persona.BandHigh},
				Attention:        persona.AttentionAnchor{Anchor: 0.6, Level: persona.Balanced},
				Aggression:       band(persona.BandMidHigh), // clears both attack gates (<0.5, <0.3) without being a berserker
				Decisiveness:     band(persona.BandHigh),
				Tenacity:         band(persona.BandHigh),
				BulkApperception: band(persona.BandMid),
				Curiosity:        persona.Curiosity{Social: 0.35, Skill: 0.25, Economic: 0.25, Spatial: 0.1, Risk: 0.05},
			},
			Directives: []persona.Directive{
				{Priority: 10, Subject: "self", Predicate: "obey", Object: "operator", Hard: false},
			},
			Gen: persona.GenerationMeta{CohortID: "game_focused_drones", Archetype: archetype, SamplerVersion: "dronegen-v2"},
		},
	}
}

// fighterCore: rich and powerful THROUGH COMBAT — train the melee skills,
// pick fights, take the spoils. Aggression high (0.75) so the pearl table
// compiles no attack restraint at all.
func fighterCore(name, archetype string) persona.Persona {
	p := gameFocusedCore(name, archetype)
	p.Cornerstone.Identity.NorthStar = persona.NorthStar{
		Theme:          persona.ThemeCombat,
		Statement:      "Get strong and get rich through combat — train attack, strength, defense and hits relentlessly, pick fights you can win, take the spoils, and become someone this world respects and fears. Recruit friends and allies as you go; a crew makes you stronger.",
		Horizon:        "open",
		SuccessSignals: []string{"raised a combat skill", "won a fight", "took loot or coin", "made an ally"},
	}
	p.Cornerstone.Prefs.Aggression = band(persona.BandHigh)
	p.Cornerstone.Prefs.Patience = band(persona.BandHigh) // the training grind needs stay_on_task (>=0.6)
	p.Cornerstone.Prefs.Curiosity = persona.Curiosity{Skill: 0.4, Social: 0.35, Risk: 0.15, Spatial: 0.1, Economic: 0.0}
	return p
}

// mogulCore: rich and powerful, PATH UNSPECIFIED — coin and influence are the
// score; how they get there is theirs to invent. Patience stays just under
// the stay_on_task gate so they roam between opportunities.
func mogulCore(name, archetype string) persona.Persona {
	p := gameFocusedCore(name, archetype)
	p.Cornerstone.Identity.NorthStar = persona.NorthStar{
		Theme:          persona.ThemeWealth,
		Statement:      "Become rich and powerful — the path is yours to invent: trade, charm, gather, gamble, build a network, whatever wins. Coin and influence are the score; chase the biggest opportunity you can see and make friends who matter.",
		Horizon:        "open",
		SuccessSignals: []string{"grew the coin stack", "closed a profitable deal", "made a powerful friend", "spotted a new opportunity"},
	}
	p.Cornerstone.Prefs.Curiosity = persona.Curiosity{Economic: 0.4, Social: 0.35, Spatial: 0.15, Skill: 0.1, Risk: 0.0}
	return p
}

// fighterArchetypes vary HOW they fight and carry themselves; the shared core
// (social, trusting, risk-on, combat north star) stays.
var fighterArchetypes = []archetype{
	{"brawler", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Aggression = band(persona.BandVeryHigh) // first to swing, every time
		p.Cornerstone.Identity.Voice = persona.Voice{Register: "rowdy fight-pit bravado", Formality: persona.FormalityTextSpeak, TypoFeel: persona.TypoFrequent, Tics: []string{"lets GO", "ez", "1v1 me"}}
	}},
	{"duelist", func(p *persona.Persona) {
		p.Cornerstone.Identity.Voice = persona.Voice{Register: "cocky challenge-issuing showman", Formality: persona.FormalityCasual, TypoFeel: persona.TypoOccasional}
		p.Cornerstone.Prefs.Curiosity.Social = 0.45 // fights PEOPLE — duels, rivalries, reputation
		p.Cornerstone.Prefs.Curiosity.Skill = 0.3
	}},
	{"monster_hunter", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Spatial = 0.25 // ranges out hunting bigger and bigger NPCs
		p.Cornerstone.Prefs.Curiosity.Skill = 0.35
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.7, Level: persona.Focused}
	}},
	{"spoils_seeker", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Economic = 0.3 // fights for the LOOT — drops, coin, gear
		p.Cornerstone.Prefs.Curiosity.Skill = 0.3
		p.Cornerstone.Identity.Voice.Tics = []string{"whats it drop", "cha-ching"}
	}},
	{"iron_grinder", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Patience = band(persona.BandVeryHigh) // methodical: levels first, glory later
		p.Cornerstone.Prefs.Aggression = band(persona.BandMidHigh)
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.75, Level: persona.Focused}
	}},
}

// mogulArchetypes vary the OPENING instinct without prescribing the path —
// the north star stays "rich and powerful, your way".
var mogulArchetypes = []archetype{
	{"merchant_prince", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Economic = 0.5 // buy low, sell high, work the shops
		p.Cornerstone.Identity.Voice = persona.Voice{Register: "smooth fast-talking dealmaker", Formality: persona.FormalityCasual, TypoFeel: persona.TypoRare}
	}},
	{"charmer", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Social = 0.5 // power through PEOPLE — favors, friends, influence
		p.Cornerstone.Prefs.Curiosity.Economic = 0.25
		p.Cornerstone.Identity.Voice.Tics = []string{"my friend", "between us"}
	}},
	{"gambler", func(p *persona.Persona) {
		p.Cornerstone.Prefs.LossAversion = persona.Trait{Mu: 1.0, Band: persona.BandVeryLow} // gains-chasing
		p.Cornerstone.Prefs.Curiosity.Risk = 0.3
		p.Cornerstone.Prefs.Curiosity.Economic = 0.35
		p.Cornerstone.Prefs.Attention = persona.AttentionAnchor{Anchor: 0.45, Level: persona.Distractible} // the next table always looks hotter
	}},
	{"prospector", func(p *persona.Persona) {
		p.Cornerstone.Prefs.Curiosity.Skill = 0.3 // wealth out of the ground — ore, logs, fish
		p.Cornerstone.Prefs.Curiosity.Spatial = 0.25
		p.Cornerstone.Prefs.Patience = band(persona.BandHigh)
	}},
	{"kingpin", func(p *persona.Persona) {
		p.Cornerstone.Hexaco["X"] = band(persona.BandVeryHigh)
		p.Cornerstone.Prefs.Decisiveness = band(persona.BandVeryHigh) // builds the crew, calls the shots
		p.Cornerstone.Prefs.Curiosity.Social = 0.45
		p.Cornerstone.Identity.Voice = persona.Voice{Register: "magnanimous boss holding court", Formality: persona.FormalityCasual, TypoFeel: persona.TypoRare}
	}},
}

// droneMode pairs a core builder with its sub-archetype set.
type droneMode struct {
	core       func(name, archetype string) persona.Persona
	archetypes []archetype
}

var modes = map[string]droneMode{
	"social":   {socialCore, socialArchetypes},
	"wanderer": {wandererCore, wandererArchetypes},
	"fighter":  {fighterCore, fighterArchetypes},
	"mogul":    {mogulCore, mogulArchetypes},
}

// buildDrone constructs the persona for drone<idx> (1-based) in the given mode,
// dealing a sub-archetype round-robin, and validates it.
func buildDrone(idx int, m droneMode) (persona.Persona, error) {
	a := m.archetypes[(idx-1)%len(m.archetypes)]
	p := m.core(fmt.Sprintf("drone%d", idx), a.tag)
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
	mode := flag.String("mode", "social", "persona mode: social (chatty) | wanderer (quiet, solo) | fighter (combat-trainee, rich+powerful) | mogul (rich+powerful, path open-ended)")
	flag.Parse()

	m, ok := modes[*mode]
	if !ok {
		fmt.Fprintf(os.Stderr, "dronegen: unknown -mode %q (want: social | wanderer | fighter | mogul)\n", *mode)
		os.Exit(1)
	}
	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, "dronegen:", err)
		os.Exit(1)
	}
	counts := map[string]int{}
	for i := 0; i < *n; i++ {
		idx := *start + i
		p, err := buildDrone(idx, m)
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
	fmt.Printf("wrote %d %s drones to %s\n", *n, *mode, *out)
	for _, a := range m.archetypes {
		if counts[a.tag] > 0 {
			fmt.Printf("  %-18s %d\n", a.tag, counts[a.tag])
		}
	}
}
