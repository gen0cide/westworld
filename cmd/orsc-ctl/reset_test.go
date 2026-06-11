package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gen0cide/westworld/facts"
)

// --- host-range parsing ------------------------------------------------------

func TestParseHosts(t *testing.T) {
	cases := []struct {
		spec string
		want []string
	}{
		{"drone51..drone54", []string{"drone51", "drone52", "drone53", "drone54"}},
		{"drone51..54", []string{"drone51", "drone52", "drone53", "drone54"}},
		{"drone7", []string{"drone7"}},
		{"drone1..drone2,drone9", []string{"drone1", "drone2", "drone9"}},
		{"drone1,drone1..drone2", []string{"drone1", "drone2"}}, // dedup, order kept
		{"drone001..3", []string{"drone001", "drone002", "drone003"}},
		{" drone1 , drone2 ", []string{"drone1", "drone2"}},
		{"drone5..drone5", []string{"drone5"}},
	}
	for _, tc := range cases {
		got, err := parseHosts(tc.spec)
		if err != nil {
			t.Errorf("parseHosts(%q) error: %v", tc.spec, err)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("parseHosts(%q) = %v, want %v", tc.spec, got, tc.want)
		}
	}
}

func TestParseHostsRangeSize(t *testing.T) {
	got, err := parseHosts("drone51..drone150")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 100 {
		t.Fatalf("drone51..drone150 expands to %d hosts, want 100", len(got))
	}
	if got[0] != "drone51" || got[99] != "drone150" {
		t.Fatalf("range endpoints = %s..%s, want drone51..drone150", got[0], got[99])
	}
}

func TestParseHostsErrors(t *testing.T) {
	for _, spec := range []string{
		"",                  // empty
		" , ",               // only separators
		"drone10..drone5",   // end < start
		"drone..drone5",     // start has no number
		"drone1..stub5",     // prefix mismatch
		"drone1..stub",      // end has no number
		"drone1..drone1e+9", // garbage number
	} {
		if got, err := parseHosts(spec); err == nil {
			t.Errorf("parseHosts(%q) = %v, want error", spec, got)
		}
	}
}

// --- fresh-stat vector -------------------------------------------------------

// liveSkillList mirrors the westworld.conf server's 18-skill list (member
// world, no runecraft/harvesting/influence) as the admin API reports it.
func liveSkillList() []skillInfo {
	out := make([]skillInfo, len(authenticSkills))
	for i, n := range authenticSkills {
		out[i] = skillInfo{ID: i, ShortName: n, Name: n}
	}
	return out
}

func TestFreshSkillsBodyVector(t *testing.T) {
	body, err := freshSkillsBody(liveSkillList())
	if err != nil {
		t.Fatal(err)
	}
	patches := body["skills"].([]map[string]any)
	if len(patches) != 18 {
		t.Fatalf("patch count = %d, want 18", len(patches))
	}
	for i, p := range patches {
		wantLvl := freshSkillLevel
		if authenticSkills[i] == "Hits" {
			wantLvl = freshHitsLevel
		}
		if p["skill"] != fmt.Sprint(i) {
			t.Errorf("patch %d skill = %v, want %d (id-addressed)", i, p["skill"], i)
		}
		if p["currentLevel"] != wantLvl || p["baseLevel"] != wantLvl {
			t.Errorf("%s: current/base = %v/%v, want %d/%d",
				authenticSkills[i], p["currentLevel"], p["baseLevel"], wantLvl, wantLvl)
		}
		if _, hasExp := p["experience"]; hasExp {
			// setLevelTo derives experience server-side; sending it would be
			// clobbered anyway (applySkillPatch applies experience FIRST).
			t.Errorf("%s: patch carries an experience field, must not", authenticSkills[i])
		}
	}
	// Hits is index 3 in the authentic ordering — the only level-10 entry.
	if patches[3]["baseLevel"] != freshHitsLevel {
		t.Errorf("Hits (id 3) baseLevel = %v, want %d", patches[3]["baseLevel"], freshHitsLevel)
	}
}

func TestFreshSkillsBodyRefusesNoHits(t *testing.T) {
	if _, err := freshSkillsBody([]skillInfo{{ID: 0, ShortName: "Attack"}}); err == nil {
		t.Fatal("freshSkillsBody without a Hits skill must error")
	}
	if _, err := freshSkillsBody(nil); err == nil {
		t.Fatal("freshSkillsBody with an empty list must error")
	}
}

func TestVerifySkills(t *testing.T) {
	fresh := liveSkillList()
	for i := range fresh {
		lvl := freshLevelFor(fresh[i])
		fresh[i].CurrentLevel = lvl
		fresh[i].BaseLevel = lvl
		if fresh[i].isHits() {
			fresh[i].Experience = 4616 // what setLevelTo(10) leaves behind
		}
	}
	if err := verifySkills(fresh); err != nil {
		t.Fatalf("fresh vector should verify: %v", err)
	}

	// createPlayer's own Hits seed (4000) is also fresh.
	fresh[3].Experience = 4000
	if err := verifySkills(fresh); err != nil {
		t.Fatalf("createPlayer Hits xp 4000 should verify: %v", err)
	}

	dirty := liveSkillList()
	for i := range dirty {
		lvl := freshLevelFor(dirty[i])
		dirty[i].CurrentLevel = lvl
		dirty[i].BaseLevel = lvl
	}
	dirty[0].BaseLevel = 40 // attack trained
	if err := verifySkills(dirty); err == nil {
		t.Fatal("trained attack must fail verification")
	} else if !strings.Contains(err.Error(), "Attack") {
		t.Fatalf("error should name the bad skill, got: %v", err)
	}

	leftoverXP := liveSkillList()
	for i := range leftoverXP {
		lvl := freshLevelFor(leftoverXP[i])
		leftoverXP[i].CurrentLevel = lvl
		leftoverXP[i].BaseLevel = lvl
	}
	leftoverXP[3].Experience = 4616
	leftoverXP[7].Experience = 250 // cooking xp left behind
	if err := verifySkills(leftoverXP); err == nil {
		t.Fatal("non-zero xp on a level-1 skill must fail verification")
	}
}

// --- coins vs facts ----------------------------------------------------------

// TestCoinsAgainstFacts pins the coin grant to the authoritative item defs:
// catalog id 10 is "Coins" and stackable (so amount:100 is one stack).
func TestCoinsAgainstFacts(t *testing.T) {
	f := facts.LoadStaticCatalogs()
	def := f.ItemDef(coinsCatalogID)
	if def == nil {
		t.Fatalf("facts has no item def for catalog id %d", coinsCatalogID)
	}
	if def.Name != "Coins" {
		t.Fatalf("item %d is %q, want Coins", coinsCatalogID, def.Name)
	}
	if !def.IsStackable {
		t.Fatalf("Coins (id %d) must be stackable for a single 100-coin grant", coinsCatalogID)
	}
}

// --- verification predicates --------------------------------------------------

func TestVerifyInventory(t *testing.T) {
	ok := []itemEntry{{CatalogID: coinsCatalogID, Amount: 100}}
	if err := verifyInventory(ok, 100); err != nil {
		t.Fatalf("exact coin stack should verify: %v", err)
	}
	for name, items := range map[string][]itemEntry{
		"empty":        {},
		"extra stack":  {{CatalogID: coinsCatalogID, Amount: 100}, {CatalogID: 156, Amount: 1}},
		"wrong amount": {{CatalogID: coinsCatalogID, Amount: 99}},
		"wrong item":   {{CatalogID: 11, Amount: 100}},
		"noted":        {{CatalogID: coinsCatalogID, Amount: 100, Noted: true}},
	} {
		if err := verifyInventory(items, 100); err == nil {
			t.Errorf("%s should fail verification", name)
		}
	}
}

func TestVerifyBank(t *testing.T) {
	if err := verifyBank(nil); err != nil {
		t.Fatalf("empty bank should verify: %v", err)
	}
	if err := verifyBank([]itemEntry{{CatalogID: 10, Amount: 1}}); err == nil {
		t.Fatal("non-empty bank must fail verification")
	}
}

// --- command-sequence assembly -------------------------------------------------

func TestResetPlanSequence(t *testing.T) {
	plan := resetPlan("drone51", 100, authenticSkills)

	var sig []string
	for _, s := range plan {
		if s.Kind == "session" {
			sig = append(sig, "session:"+s.Method)
			continue
		}
		sig = append(sig, s.Method+" "+s.Path)
	}
	want := []string{
		"GET /players/drone51",
		"session:login",
		"GET /players/drone51/skills",
		"PATCH /players/drone51/skills",
		"PATCH /players/drone51/fatigue",
		"DELETE /players/drone51/inventory",
		"DELETE /players/drone51/bank",
		"POST /players/drone51/inventory/items",
		"GET /players/drone51/skills",
		"GET /players/drone51/inventory",
		"GET /players/drone51/bank",
		"GET /players/drone51",
		"session:logout",
	}
	if !reflect.DeepEqual(sig, want) {
		t.Fatalf("plan sequence =\n  %s\nwant\n  %s",
			strings.Join(sig, "\n  "), strings.Join(want, "\n  "))
	}

	// The wipes must precede the grant (idempotency: rerunning never accrues).
	grant := indexOf(sig, "POST /players/drone51/inventory/items")
	if indexOf(sig, "DELETE /players/drone51/inventory") > grant ||
		indexOf(sig, "DELETE /players/drone51/bank") > grant {
		t.Fatal("wipes must come before the coin grant")
	}

	// The coin grant body carries the verified catalog id and amount.
	var grantBody map[string]any
	if err := json.Unmarshal([]byte(plan[7].Body), &grantBody); err != nil {
		t.Fatalf("grant body is not JSON: %v", err)
	}
	if grantBody["catalogId"] != float64(coinsCatalogID) || grantBody["amount"] != float64(100) {
		t.Fatalf("grant body = %v, want catalogId %d amount 100", grantBody, coinsCatalogID)
	}
}

func indexOf(list []string, v string) int {
	for i, s := range list {
		if s == v {
			return i
		}
	}
	return -1
}

// --- end-to-end against a fake admin API ---------------------------------------

// fakeAdminState is a tiny in-memory model of one online player behind the
// admin API envelope, enough to exercise resetOneHost's full HTTP path
// (presence probe → skill patch → fatigue → wipes → grant → verification).
type fakeAdminState struct {
	mu      sync.Mutex
	skills  []skillInfo
	inv     []itemEntry
	bank    []itemEntry
	fatigue int
	reqs    []string
}

func newDirtyPlayer() *fakeAdminState {
	s := &fakeAdminState{fatigue: 5250, inv: []itemEntry{{CatalogID: 156, Amount: 1}}, bank: []itemEntry{{CatalogID: 10, Amount: 12345}}}
	for i, n := range authenticSkills {
		s.skills = append(s.skills, skillInfo{ID: i, ShortName: n, Name: n, CurrentLevel: 40, BaseLevel: 40, Experience: 999999})
	}
	return s
}

func (s *fakeAdminState) handler(t *testing.T) http.HandlerFunc {
	t.Helper()
	write := func(w http.ResponseWriter, data any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "data": data})
	}
	return func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.reqs = append(s.reqs, r.Method+" "+r.URL.Path)
		switch {
		case r.Method == "GET" && r.URL.Path == "/players/drone51":
			write(w, map[string]any{"username": "drone51", "fatigue": s.fatigue})
		case r.Method == "GET" && r.URL.Path == "/players/drone51/skills":
			write(w, map[string]any{"skills": s.skills})
		case r.Method == "PATCH" && r.URL.Path == "/players/drone51/skills":
			var body struct {
				Skills []struct {
					Skill        string `json:"skill"`
					CurrentLevel int    `json:"currentLevel"`
					BaseLevel    int    `json:"baseLevel"`
				} `json:"skills"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			for _, p := range body.Skills {
				for i := range s.skills {
					if fmt.Sprint(s.skills[i].ID) == p.Skill {
						// Mirror server semantics: setLevelTo re-derives xp.
						s.skills[i].CurrentLevel = p.CurrentLevel
						s.skills[i].BaseLevel = p.BaseLevel
						if p.BaseLevel == 1 {
							s.skills[i].Experience = 0
						} else if s.skills[i].isHits() && p.BaseLevel == 10 {
							s.skills[i].Experience = 4616
						}
					}
				}
			}
			write(w, map[string]any{"updated": len(body.Skills)})
		case r.Method == "PATCH" && r.URL.Path == "/players/drone51/fatigue":
			var body struct {
				Fatigue int `json:"fatigue"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			s.fatigue = body.Fatigue
			write(w, nil)
		case r.Method == "DELETE" && r.URL.Path == "/players/drone51/inventory":
			s.inv = nil
			write(w, nil)
		case r.Method == "DELETE" && r.URL.Path == "/players/drone51/bank":
			s.bank = nil
			write(w, nil)
		case r.Method == "POST" && r.URL.Path == "/players/drone51/inventory/items":
			var body itemEntry
			_ = json.NewDecoder(r.Body).Decode(&body)
			s.inv = append(s.inv, body)
			write(w, nil)
		case r.Method == "GET" && r.URL.Path == "/players/drone51/inventory":
			write(w, map[string]any{"inventory": map[string]any{"size": len(s.inv), "items": s.inv}})
		case r.Method == "GET" && r.URL.Path == "/players/drone51/bank":
			write(w, map[string]any{"bank": map[string]any{"size": len(s.bank), "items": s.bank}})
		default:
			w.WriteHeader(404)
			_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": map[string]any{"code": "not_found", "message": "no route " + r.URL.Path}})
		}
	}
}

func TestResetOneHostAgainstFakeServer(t *testing.T) {
	state := newDirtyPlayer()
	srv := httptest.NewServer(state.handler(t))
	defer srv.Close()

	no := false
	jsonOut = &no
	c := NewClient(srv.URL, "", 5*time.Second)

	res := resetOneHost(c, "drone51", resetOpts{coins: 100, loginWait: time.Second})
	if res.Status != "PASS" {
		t.Fatalf("resetOneHost = %s (%s), want PASS", res.Status, res.Detail)
	}
	if !res.AlreadyOnline || res.SessionOpened {
		t.Fatalf("online player should not trigger a session login: %+v", res)
	}
	if res.SkillsReset != 18 {
		t.Fatalf("skillsReset = %d, want 18", res.SkillsReset)
	}

	// The fake's end state is genuinely fresh.
	if err := verifySkills(state.skills); err != nil {
		t.Fatalf("end-state skills not fresh: %v", err)
	}
	if err := verifyInventory(state.inv, 100); err != nil {
		t.Fatalf("end-state inventory: %v", err)
	}
	if len(state.bank) != 0 || state.fatigue != 0 {
		t.Fatalf("end state bank=%d fatigue=%d, want 0/0", len(state.bank), state.fatigue)
	}

	// Idempotency: a second run converges to the identical end state and PASSes.
	res2 := resetOneHost(c, "drone51", resetOpts{coins: 100, loginWait: time.Second})
	if res2.Status != "PASS" {
		t.Fatalf("second run = %s (%s), want PASS (idempotent)", res2.Status, res2.Detail)
	}
	if err := verifyInventory(state.inv, 100); err != nil {
		t.Fatalf("after rerun: %v", err)
	}
}

// TestResetOneHostOfflineNoPassword proves the offline path refuses to dial
// the game server when the password env is empty (and that no HTTP mutation
// happens after the presence probe).
func TestResetOneHostOfflineNoPassword(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": false, "error": map[string]any{"code": "not_found", "message": "Player is not online: drone51"}})
	}))
	defer srv.Close()

	no := false
	jsonOut = &no
	c := NewClient(srv.URL, "", 5*time.Second)

	t.Setenv("ORSC_CTL_TEST_EMPTY_PW", "")
	res := resetOneHost(c, "drone51", resetOpts{coins: 100, passwordEnv: "ORSC_CTL_TEST_EMPTY_PW", loginWait: time.Second})
	if res.Status != "FAIL" {
		t.Fatalf("offline + empty password must FAIL, got %+v", res)
	}
	if !strings.Contains(res.Detail, "ORSC_CTL_TEST_EMPTY_PW") {
		t.Fatalf("failure should name the env var (never the value): %q", res.Detail)
	}
}
