package facts

import (
	_ "embed"
	"encoding/xml"
)

//go:embed SpellDef.xml
var spellDefXML []byte

// SpellDef is one entry in the magic spellbook.
type SpellDef struct {
	ID          int       // spellbook index (0-based, in the order the XML lists them)
	Name        string    // canonical name
	ReqLevel    int       // magic skill required
	Type        SpellType // category — see below
	ExpReward   int       // base XP (in tenths — matches OpenRSC)
	Description string
	Members     bool   // members-only flag
	Evil        bool   // legacy alignment flag (RSC had alignment)
	Runes       []Rune // rune ID + count per cast
}

// SpellType is the cast category (mirrors OpenRSC's `type` field).
type SpellType int

const (
	SpellSelf         SpellType = 1 // self-targeted (heal, teleport, etc.)
	SpellOffensive    SpellType = 2 // attack on NPC/player
	SpellCurse        SpellType = 3 // debuff on NPC/player
	SpellInventory    SpellType = 4 // enchant inventory item / alch
	SpellTeleother    SpellType = 5 // teleother (unused?)
	SpellSummon       SpellType = 6 // summons (unused in stock RSC)
)

// Rune is one entry in a spell's requiredRunes — item id + count.
type Rune struct {
	ItemID int
	Count  int
}

// Spells is the parsed catalog. Populated at init().
var Spells []SpellDef

// spellByName/ByID indexes for O(1) lookup.
var (
	spellByID   map[int]*SpellDef
	spellByName map[string]*SpellDef
)

// Wire shape that matches the OpenRSC XML format.
type spellDefXMLArr struct {
	XMLName xml.Name `xml:"SpellDef-array"`
	Spells  []struct {
		ReqLevel    int `xml:"reqLevel"`
		Type        int `xml:"type"`
		RuneCount   int `xml:"runeCount"`
		Runes       struct {
			Entries []struct {
				Ints []int `xml:"int"`
			} `xml:"entry"`
		} `xml:"requiredRunes"`
		Exp         int    `xml:"exp"`
		Name        string `xml:"name"`
		Description string `xml:"description"`
		Members     bool   `xml:"members"`
		Evil        bool   `xml:"evil"`
	} `xml:"SpellDef"`
}

func init() {
	var arr spellDefXMLArr
	if err := xml.Unmarshal(spellDefXML, &arr); err != nil {
		// Don't panic — let routines see an empty catalog rather
		// than crashing every host on a parse glitch. Tests will
		// catch this.
		Spells = nil
		return
	}
	spellByID = make(map[int]*SpellDef, len(arr.Spells))
	spellByName = make(map[string]*SpellDef, len(arr.Spells))
	for i, s := range arr.Spells {
		runes := make([]Rune, 0, len(s.Runes.Entries))
		for _, entry := range s.Runes.Entries {
			if len(entry.Ints) >= 2 {
				runes = append(runes, Rune{ItemID: entry.Ints[0], Count: entry.Ints[1]})
			}
		}
		Spells = append(Spells, SpellDef{
			ID:          i,
			Name:        s.Name,
			ReqLevel:    s.ReqLevel,
			Type:        SpellType(s.Type),
			ExpReward:   s.Exp,
			Description: s.Description,
			Members:     s.Members,
			Evil:        s.Evil,
			Runes:       runes,
		})
	}
	for i := range Spells {
		spellByID[Spells[i].ID] = &Spells[i]
		spellByName[lowerASCII(Spells[i].Name)] = &Spells[i]
	}
}

// SpellByID returns the spell def by spellbook index.
func SpellByID(id int) *SpellDef {
	if spellByID == nil {
		return nil
	}
	return spellByID[id]
}

// SpellByName does a case-insensitive name lookup.
func SpellByName(name string) *SpellDef {
	if spellByName == nil {
		return nil
	}
	return spellByName[lowerASCII(name)]
}

func lowerASCII(s string) string {
	out := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 32
		}
		out[i] = c
	}
	return string(out)
}
