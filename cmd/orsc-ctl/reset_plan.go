package main

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Pure planning logic for `orsc-ctl reset`: host-range parsing, the
// fresh-account stat vector, the per-host admin-API step sequence, and the
// verification predicates. Everything here is network-free so it can be
// unit-tested without a server; reset.go owns the actual execution.

// coinsCatalogID is the catalog id of "Coins" (verified against
// facts/static_defs_gen.go and the OpenRSC item defs: id 10, stackable).
const coinsCatalogID = 10

// Fresh-account levels per the OpenRSC server source. createPlayer
// (server/src/.../database/GameDatabase.java) initializes every skill to its
// SkillDef minLevel; constants/Skills.java declares minLevel 10 for Hits and 1
// for every other skill. The same vector appears in local/delores_reset.sql.
const (
	freshSkillLevel = 1
	freshHitsLevel  = 10
)

// Fresh-account experience bounds for Hits, in the server's internal units
// (the admin API reads/writes Skills.exps directly). createPlayer seeds Hits
// at 4000; our reset patch goes through Skills.setLevelTo(10) which re-derives
// experienceForLevel(10) = 4616 (the ORIGINAL curve entry for level 10). Both
// are "fresh level-10 Hits"; verification accepts the closed range.
const (
	freshHitsExpMin = 4000
	freshHitsExpMax = 4616
)

// authenticSkills is the westworld.conf skill set (member world, no runecraft,
// no harvesting, no influence): the 18 authentic skills in server index order,
// by shortName as the admin API reports them. Used only for dry-run display —
// a live run derives the actual list from GET /players/{u}/skills.
var authenticSkills = []string{
	"Attack", "Defense", "Strength", "Hits", "Ranged", "Prayer", "Magic",
	"Cooking", "Woodcut", "Fletching", "Fishing", "Firemaking", "Crafting",
	"Smithing", "Mining", "Herblaw", "Agility", "Thieving",
}

// parseHosts expands a host spec into an ordered, de-duplicated list of
// account names. The spec is a comma-separated list of tokens; each token is
// either a literal name ("drone7") or an inclusive range "prefixN..prefixM"
// ("drone51..drone150", shorthand "drone51..150"). The end's prefix, when
// present, must match the start's. Zero-padded starts keep their width
// ("drone001..3" → drone001 drone002 drone003).
func parseHosts(spec string) ([]string, error) {
	var out []string
	seen := map[string]bool{}
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}

	for _, tok := range strings.Split(spec, ",") {
		tok = strings.TrimSpace(tok)
		if tok == "" {
			continue
		}
		if !strings.Contains(tok, "..") {
			add(tok)
			continue
		}

		parts := strings.SplitN(tok, "..", 2)
		startName, endStr := parts[0], parts[1]
		prefix, startNum := splitTrailingNumber(startName)
		if startNum == "" {
			return nil, fmt.Errorf("range %q: start %q has no trailing number", tok, startName)
		}
		// The end may be a full name (drone150) or a bare number (150).
		endPrefix, endNum := splitTrailingNumber(endStr)
		if endNum == "" {
			return nil, fmt.Errorf("range %q: end %q has no trailing number", tok, endStr)
		}
		if endPrefix != "" && endPrefix != prefix {
			return nil, fmt.Errorf("range %q: prefix mismatch (%q vs %q)", tok, prefix, endPrefix)
		}
		lo, err := strconv.Atoi(startNum)
		if err != nil {
			return nil, fmt.Errorf("range %q: bad start number: %w", tok, err)
		}
		hi, err := strconv.Atoi(endNum)
		if err != nil {
			return nil, fmt.Errorf("range %q: bad end number: %w", tok, err)
		}
		if hi < lo {
			return nil, fmt.Errorf("range %q: end %d < start %d", tok, hi, lo)
		}
		width := 0
		if strings.HasPrefix(startNum, "0") && len(startNum) > 1 {
			width = len(startNum)
		}
		for n := lo; n <= hi; n++ {
			if width > 0 {
				add(fmt.Sprintf("%s%0*d", prefix, width, n))
			} else {
				add(fmt.Sprintf("%s%d", prefix, n))
			}
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("host spec %q expands to no hosts", spec)
	}
	return out, nil
}

// splitTrailingNumber splits "drone51" into ("drone", "51"). A token with no
// trailing digits returns ("token", "").
func splitTrailingNumber(s string) (prefix, num string) {
	i := len(s)
	for i > 0 && s[i-1] >= '0' && s[i-1] <= '9' {
		i--
	}
	return s[:i], s[i:]
}

// skillInfo is one entry of the admin API's GET /players/{u}/skills response.
type skillInfo struct {
	ID           int    `json:"id"`
	ShortName    string `json:"shortName"`
	Name         string `json:"name"`
	CurrentLevel int    `json:"currentLevel"`
	BaseLevel    int    `json:"baseLevel"`
	Experience   int    `json:"experience"`
}

// isHits reports whether a skill entry is the Hits skill (the one fresh
// accounts start at level 10 rather than 1).
func (s skillInfo) isHits() bool {
	return strings.EqualFold(s.ShortName, "hits") || strings.EqualFold(s.Name, "hits")
}

// freshLevelFor returns the fresh-account level for a skill entry.
func freshLevelFor(s skillInfo) int {
	if s.isHits() {
		return freshHitsLevel
	}
	return freshSkillLevel
}

// freshSkillsBody builds the PATCH /players/{u}/skills request body that
// resets every listed skill to its fresh-account level. Skills are addressed
// by numeric id (as a string — the server's skillId() accepts both) so the
// patch is immune to name-spelling drift. Setting baseLevel goes through the
// server's Skills.setLevelTo, which also re-derives experience for the level
// (0 for level 1, 4616 internal units for Hits at 10), so no explicit
// experience field is needed — and order matters server-side: an experience
// field would be applied FIRST and then clobbered by the baseLevel patch.
//
// Errors if the list contains no Hits skill: that would mean the server's
// skill config is not the one this vector was derived for.
func freshSkillsBody(skills []skillInfo) (map[string]any, error) {
	if len(skills) == 0 {
		return nil, fmt.Errorf("empty skill list from server")
	}
	hasHits := false
	patches := make([]map[string]any, 0, len(skills))
	for _, s := range skills {
		if s.isHits() {
			hasHits = true
		}
		lvl := freshLevelFor(s)
		patches = append(patches, map[string]any{
			"skill":        strconv.Itoa(s.ID),
			"currentLevel": lvl,
			"baseLevel":    lvl,
		})
	}
	if !hasHits {
		return nil, fmt.Errorf("server skill list has no Hits skill (%d skills) — refusing to apply the fresh vector", len(skills))
	}
	return map[string]any{"skills": patches}, nil
}

// resetStep is one printable step of the per-host sequence; used by both the
// dry-run output and (for the HTTP steps) the executor's logging.
type resetStep struct {
	Kind   string `json:"kind"`   // "session" or "http"
	Method string `json:"method"` // HTTP verb, or the session action
	Path   string `json:"path"`   // admin-API path, or "" for session steps
	Body   string `json:"body"`   // JSON body, or a human note
}

// resetPlan assembles the full per-host sequence for username. skillNames is
// only used to render the skills-PATCH body (dry-run uses authenticSkills; a
// live run rebuilds the body from the server's own skill list).
func resetPlan(username string, coins int, skillNames []string) []resetStep {
	skills := make([]skillInfo, len(skillNames))
	for i, n := range skillNames {
		skills[i] = skillInfo{ID: i, ShortName: n, Name: n}
	}
	skillsBody, err := freshSkillsBody(skills)
	var skillsJSON string
	if err != nil {
		skillsJSON = "(derived from live GET /players/{u}/skills)"
	} else {
		b, _ := json.Marshal(skillsBody)
		skillsJSON = string(b)
	}
	coinsBody, _ := json.Marshal(map[string]any{
		"catalogId": coinsCatalogID, "amount": coins, "noted": false,
	})
	p := "/players/" + username

	return []resetStep{
		{Kind: "http", Method: "GET", Path: p, Body: "(presence probe: online already, or needs a session login?)"},
		{Kind: "session", Method: "login", Body: "dial game server + v235 login as " + username + " (password from env; only when offline), then poll GET " + p + " until the server registers the player"},
		{Kind: "http", Method: "GET", Path: p + "/skills", Body: "(read live skill list; fresh vector is rebuilt from it)"},
		{Kind: "http", Method: "PATCH", Path: p + "/skills", Body: skillsJSON},
		{Kind: "http", Method: "PATCH", Path: p + "/fatigue", Body: `{"fatigue":0}`},
		{Kind: "http", Method: "DELETE", Path: p + "/inventory", Body: "(wipe all carried items, worn included)"},
		{Kind: "http", Method: "DELETE", Path: p + "/bank", Body: "(wipe all banked items)"},
		{Kind: "http", Method: "POST", Path: p + "/inventory/items", Body: string(coinsBody)},
		{Kind: "http", Method: "GET", Path: p + "/skills", Body: "(verify: every skill current==base==fresh level; xp 0, Hits 4000..4616)"},
		{Kind: "http", Method: "GET", Path: p + "/inventory", Body: fmt.Sprintf("(verify: exactly one stack — %d coins, catalogId %d)", coins, coinsCatalogID)},
		{Kind: "http", Method: "GET", Path: p + "/bank", Body: "(verify: empty)"},
		{Kind: "http", Method: "GET", Path: p, Body: "(verify: fatigue 0)"},
		{Kind: "session", Method: "logout", Body: "clean logout + close (only if this run opened the session)"},
	}
}

// itemEntry is one entry of the admin API inventory/bank item arrays.
type itemEntry struct {
	CatalogID int  `json:"catalogId"`
	Amount    int  `json:"amount"`
	Noted     bool `json:"noted"`
	Wielded   bool `json:"wielded"`
}

// verifySkills checks a post-reset skill list against the fresh vector.
// Returns nil when every skill sits at its fresh level with fresh experience.
func verifySkills(skills []skillInfo) error {
	if len(skills) == 0 {
		return fmt.Errorf("verify skills: server returned no skills")
	}
	var bad []string
	for _, s := range skills {
		want := freshLevelFor(s)
		ok := s.CurrentLevel == want && s.BaseLevel == want
		if s.isHits() {
			ok = ok && s.Experience >= freshHitsExpMin && s.Experience <= freshHitsExpMax
		} else {
			ok = ok && s.Experience == 0
		}
		if !ok {
			bad = append(bad, fmt.Sprintf("%s cur=%d base=%d xp=%d (want %d/%d)",
				s.ShortName, s.CurrentLevel, s.BaseLevel, s.Experience, want, want))
		}
	}
	if len(bad) > 0 {
		sort.Strings(bad)
		return fmt.Errorf("verify skills: %s", strings.Join(bad, "; "))
	}
	return nil
}

// verifyInventory checks that the post-reset inventory is exactly one
// unwielded stack of `coins` coins.
func verifyInventory(items []itemEntry, coins int) error {
	if len(items) != 1 {
		return fmt.Errorf("verify inventory: %d stacks, want exactly 1 (the coins)", len(items))
	}
	it := items[0]
	if it.CatalogID != coinsCatalogID || it.Amount != coins || it.Noted || it.Wielded {
		return fmt.Errorf("verify inventory: got catalogId=%d amount=%d noted=%v wielded=%v, want catalogId=%d amount=%d plain",
			it.CatalogID, it.Amount, it.Noted, it.Wielded, coinsCatalogID, coins)
	}
	return nil
}

// verifyBank checks that the post-reset bank is empty.
func verifyBank(items []itemEntry) error {
	if len(items) != 0 {
		return fmt.Errorf("verify bank: %d stacks remain, want 0", len(items))
	}
	return nil
}
