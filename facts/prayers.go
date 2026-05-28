package facts

// PrayerDef is the static definition of a prayer slot.
type PrayerDef struct {
	ID          int    // slot index 0..13 (in stock RSC; OpenRSC adds Protect from Melee at 14)
	Name        string // canonical name (lower-case match accepted in lookups)
	ReqLevel    int    // prayer skill level required
	DrainRate   int    // points/sec drained while active (RSC values from PrayerDef.xml)
	Description string
}

// Prayers is the static catalog. Mirrors
// openrsc/server/conf/server/defs/PrayerDef.xml.
var Prayers = []PrayerDef{
	{ID: 0, Name: "Thick skin", ReqLevel: 1, DrainRate: 15, Description: "Increases your defense by 5%"},
	{ID: 1, Name: "Burst of strength", ReqLevel: 4, DrainRate: 15, Description: "Increases your strength by 5%"},
	{ID: 2, Name: "Clarity of thought", ReqLevel: 7, DrainRate: 15, Description: "Increases your attack by 5%"},
	{ID: 3, Name: "Rock skin", ReqLevel: 10, DrainRate: 30, Description: "Increases your defense by 10%"},
	{ID: 4, Name: "Superhuman strength", ReqLevel: 13, DrainRate: 30, Description: "Increases your strength by 10%"},
	{ID: 5, Name: "Improved reflexes", ReqLevel: 16, DrainRate: 30, Description: "Increases your attack by 10%"},
	{ID: 6, Name: "Rapid restore", ReqLevel: 19, DrainRate: 5, Description: "2x restore rate for non-combat stats"},
	{ID: 7, Name: "Rapid heal", ReqLevel: 22, DrainRate: 10, Description: "2x restore rate for hits"},
	{ID: 8, Name: "Protect item", ReqLevel: 25, DrainRate: 10, Description: "Keeps 1 extra item on death"},
	{ID: 9, Name: "Steel skin", ReqLevel: 28, DrainRate: 60, Description: "Increases your defense by 15%"},
	{ID: 10, Name: "Ultimate strength", ReqLevel: 31, DrainRate: 60, Description: "Increases your strength by 15%"},
	{ID: 11, Name: "Incredible reflexes", ReqLevel: 34, DrainRate: 60, Description: "Increases your attack by 15%"},
	{ID: 12, Name: "Paralyze monster", ReqLevel: 37, DrainRate: 60, Description: "Stops monsters from retaliating"},
	{ID: 13, Name: "Protect from missiles", ReqLevel: 40, DrainRate: 60, Description: "100% protection from ranged attacks"},
}

// PrayerByID returns the prayer def by slot index, or nil if out of
// range.
func PrayerByID(id int) *PrayerDef {
	if id < 0 || id >= len(Prayers) {
		return nil
	}
	return &Prayers[id]
}

// PrayerByName does a case-insensitive name lookup. Returns nil if
// no match. Routines should accept either form so authors can write
// activate_prayer(facts.PrayerByName("thick skin").ID).
func PrayerByName(name string) *PrayerDef {
	for i := range Prayers {
		if equalFold(Prayers[i].Name, name) {
			return &Prayers[i]
		}
	}
	return nil
}

// equalFold is a minimal case-fold compare. Kept local so the
// facts package doesn't pull in strings — keeps facts a no-import
// leaf in test builds.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
