package runtime

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/worldmap"
)

// Map-PERCEPTION verbs (cognition-first). The WorldOracle INFORMS; it never
// decides and never auto-routes. search_map / reachable / survey_map let the
// host SEE the static geography — which destinations are open, which sit
// behind a gate it can pay (gated), and which behind a gate it cannot
// (blocked) — naming the gate, its requirement, and what the host currently
// has. The brain then chooses: pay the toll, pick a free alternative, or go
// earn coins. This is the deliberate counterpart to go_to, which stays
// reachability-blind and can walk a host into a gate it can't pass.
//
// Each query costs in-world "study the map" seconds, charged the same way the
// Act planner charges idle time (wait(N) → a ctx-cancellable sleep), so a map
// query is a real turn the conductor can abort, not a free oracle peek.
//
// Registration lives in the central actionHandlers table (dsl_actions.go);
// the spec rows live in dsl/spec/actions.go.

const (
	// studySecondsSearch is the in-world cost of a ranked search_map query
	// (it floods once per candidate destination — the heaviest of the three).
	studySecondsSearch = 3.0
	// studySecondsReachable is the cost of a single-tile reachable() check.
	studySecondsReachable = 1.5
	// studySecondsSurvey is the cost of a survey_map() overview.
	studySecondsSurvey = 2.0

	// searchMapMaxResults caps how many destinations of a type search_map
	// ranks + explains, so the flood cost (one per candidate) stays bounded.
	searchMapMaxResults = 6
)

// studyMap blocks for `secs` in-world seconds the same cancellation-respecting
// way dslWait does (select on ctx.Done() vs a timer) — NEVER time.Sleep, so the
// conductor/host ctx can abort a long map study. This is what makes a map query
// cost game-time like an idle Act turn instead of being a free peek.
func studyMap(ctx context.Context, secs float64) error {
	if secs <= 0 {
		return nil
	}
	t := time.NewTimer(time.Duration(secs * float64(time.Second)))
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// hostCapability builds a worldmap.Capability from the host's LIVE state, the
// way the reachability flood honors it:
//   - Coins: the total coin count in inventory (resolved by name via facts so
//     the real id is used, not a hard-coded constant).
//   - Items: every held inventory item, keyed by its facts name (Satisfies
//     matches case-insensitively + substring, so free-text transport item
//     names line up).
//   - Skills: all 18 skills, name→current level, via the canonical skillIDs
//     table + Self.SkillLevel.
//   - QuestsDone: empty in v1 (per-quest completion flags are not tracked yet),
//     so quest-gated edges degrade to reach="blocked", needs="quest <name>".
//   - Members: false in v1 (no members flag on HostConfig/Self yet), so
//     members-gated edges degrade to blocked — matching the QuestsDone degrade.
func (h *Host) hostCapability() worldmap.Capability {
	cap := worldmap.Capability{
		Items:      map[string]bool{},
		Skills:     map[string]int{},
		QuestsDone: map[string]bool{},
		Members:    false,
	}
	// Coins: resolve the coin item id by name so we honor the live catalog (the
	// test fixtures use a different id than live RSC's 10).
	if h.facts != nil {
		if def := h.facts.ItemDefByName("Coins"); def != nil {
			cap.Coins = h.world.Inventory.Count(def.ID)
		}
	}
	// Items: held inventory by friendly name.
	for _, s := range h.world.Inventory.Slots() {
		if s.ItemID <= 0 || s.Amount <= 0 {
			continue
		}
		if h.facts != nil {
			if def := h.facts.ItemDef(s.ItemID); def != nil && def.Name != "" {
				cap.Items[def.Name] = true
			}
		}
	}
	// Skills: name → current level (the gate vocab is name-keyed).
	for name, id := range skillIDs {
		cap.Skills[name] = h.world.Self.SkillLevel(id)
	}
	return cap
}

// searchMapHit is one resolved+explained destination record.
type searchMapHit struct {
	label string
	x, y  int
	dist  int
	info  worldmap.ReachInfo
}

// toMap renders a hit as the *interp.Map the host reads: label, x, y, dist,
// reach, gate, needs, you_have, payable. gate/needs are empty strings on an
// open reach so the host can branch on `hit["reach"] == "open"`.
func (hit searchMapHit) toMap() *interp.Map {
	return &interp.Map{Items: map[string]interp.Value{
		"label":    interp.String(hit.label),
		"x":        interp.Int(int64(hit.x)),
		"y":        interp.Int(int64(hit.y)),
		"dist":     interp.Int(int64(hit.dist)),
		"reach":    interp.String(string(hit.info.Reach)),
		"gate":     interp.String(hit.info.Gate),
		"needs":    interp.String(hit.info.Needs),
		"you_have": interp.Int(int64(hit.info.YouHave)),
		"payable":  interp.Bool(hit.info.Payable),
	}}
}

// dslSearchMap is the cognition-first destination chooser. Given a POI TYPE
// ("mining-site", "bank", "fishing-point"), it ranks the nearest real
// destinations of that type and, for EACH, explains the reach verdict under the
// host's live capability: open (free walk) | gated (a gate is in the way but
// the host can pay it) | blocked (a gate it cannot meet). Every gated/blocked
// record carries the gate name, its requirement (needs), what the host has
// (you_have), and payable. The host then DECIDES — go_to a reach="open" one,
// pay the toll only when payable, or go earn coins. INFORM, never auto-route.
//
//	search_map("mining-site")  -> ranked list of {label,x,y,dist,reach,gate,needs,you_have,payable}
func dslSearchMap(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("search_map takes 1 arg (destination type, e.g. \"mining-site\"), got %d", len(args))
	}
	if h.worldOracle == nil {
		return interp.Fail(interp.NO_SUCH_ITEM, "search_map: no map data loaded"), nil
	}
	typ := strings.TrimSpace(args[0].Display())
	if typ == "" {
		return nil, errf("search_map: destination type must be a non-empty string")
	}

	// Charge in-world study time first — a map query is a real turn.
	if err := studyMap(ctx, studySecondsSearch); err != nil {
		return nil, err
	}

	pos := h.world.Self.Position()
	cap := h.hostCapability()

	// Candidate destinations: every indexed POI of this type, bound to its
	// walking component + snapped to a standable tile. Rank by distance from
	// the host and explain the nearest few (the flood is per-candidate).
	type cand struct {
		label string
		x, y  int
		dist  int
	}
	var cands []cand
	for _, p := range h.worldOracle.POIs() {
		if p.Comp < 0 {
			continue // POI never snapped to a standable tile
		}
		if !strings.Contains(strings.ToLower(p.Type), strings.ToLower(typ)) {
			continue
		}
		cands = append(cands, cand{
			label: h.destinationLabel(p),
			x:     p.SnapX,
			y:     p.SnapY,
			dist:  chebyshev(pos.X, pos.Y, p.SnapX, p.SnapY),
		})
	}
	if len(cands) == 0 {
		// No destinations of this type — like go_to's unknown-POI failure, so
		// the host branches on r.err and tries another type / coords.
		return interp.Fail(interp.NO_SUCH_ITEM,
			fmt.Sprintf("search_map: no %q destinations on the map", typ)), nil
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].dist < cands[j].dist })
	if len(cands) > searchMapMaxResults {
		cands = cands[:searchMapMaxResults]
	}

	items := make([]interp.Value, 0, len(cands))
	for _, c := range cands {
		info := h.worldOracle.ExplainReach(pos.X, pos.Y, c.x, c.y, cap)
		hit := searchMapHit{label: c.label, x: c.x, y: c.y, dist: c.dist, info: info}
		items = append(items, hit.toMap())
	}
	return interp.Ok(&interp.List{Items: items}), nil
}

// dslReachable explains the reach verdict for ONE specific tile from the host's
// position under its live capability: {reach, gate, needs, you_have, payable}.
// Use it to vet a coordinate the host already has in mind before committing a
// go_to. Returns a single *interp.Map (not a list).
//
//	reachable(74, 583)  -> {reach:"gated", gate:"Toll gate", needs:"10 coins", you_have:15, payable:true}
func dslReachable(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	if h.worldOracle == nil {
		return interp.Fail(interp.NO_SUCH_ITEM, "reachable: no map data loaded"), nil
	}
	x, y, err := resolvePoint(args, named)
	if err != nil {
		return nil, errf("reachable: %v", err)
	}
	if err := studyMap(ctx, studySecondsReachable); err != nil {
		return nil, err
	}
	pos := h.world.Self.Position()
	cap := h.hostCapability()
	info := h.worldOracle.ExplainReach(pos.X, pos.Y, x, y, cap)
	if !info.SnapOK() {
		return interp.Fail(interp.NO_SUCH_ITEM,
			fmt.Sprintf("reachable: (%d, %d) is not a standable tile on the map", x, y)), nil
	}
	hit := searchMapHit{x: x, y: y, dist: chebyshev(pos.X, pos.Y, x, y), info: info}
	return interp.Ok(hit.toMap()), nil
}

// dslSurveyMap returns a short, high-level TEXT overview of where the host is
// and what major destinations around it are open vs gated vs blocked — the
// "stand on a hill and look out" verb. It surveys the POI types a traveling
// host cares about (banks, mines, fishing, furnaces, …) within a wide radius,
// explains the nearest of each, and summarizes. Text (not structured) because
// it is for orientation/narration; use search_map(type) to actually choose.
func dslSurveyMap(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if h.worldOracle == nil {
		return interp.String("I have no map to study right now."), nil
	}
	if err := studyMap(ctx, studySecondsSurvey); err != nil {
		return nil, err
	}
	pos := h.world.Self.Position()
	cap := h.hostCapability()

	var b strings.Builder
	// Where am I — reuse the gazetteer location summary the host already trusts.
	b.WriteString(h.LocationSummary())
	b.WriteString("\n\nMajor destinations from here:")

	// The POI types a traveling host typically routes toward. For each, take
	// the nearest indexed POI and explain its reach.
	surveyTypes := []string{"bank", "mining-site", "fishing-point", "furnace", "altar", "general-shop"}
	found := 0
	for _, typ := range surveyTypes {
		best, ok := h.nearestExplainedPOI(pos.X, pos.Y, typ, cap)
		if !ok {
			continue
		}
		found++
		switch best.info.Reach {
		case worldmap.ReachOpen:
			b.WriteString(fmt.Sprintf("\n- %s (%s): OPEN — %d tiles %s, free to walk to.",
				typ, best.label, best.dist, dirOf(pos.X, pos.Y, best.x, best.y)))
		case worldmap.ReachGated:
			b.WriteString(fmt.Sprintf("\n- %s (%s): GATED by %s — needs %s, you have %d (you CAN pay). %d tiles %s.",
				typ, best.label, gateShort(best.info.Gate), best.info.Needs, best.info.YouHave,
				best.dist, dirOf(pos.X, pos.Y, best.x, best.y)))
		case worldmap.ReachBlocked:
			b.WriteString(fmt.Sprintf("\n- %s (%s): BLOCKED by %s — needs %s, you have %d (cannot pass yet). %d tiles %s.",
				typ, best.label, gateShort(best.info.Gate), best.info.Needs, best.info.YouHave,
				best.dist, dirOf(pos.X, pos.Y, best.x, best.y)))
		}
	}
	if found == 0 {
		b.WriteString(" none of note nearby.")
	}
	b.WriteString("\n\nCall search_map(\"<type>\") to see every nearby destination of a type and decide; the oracle tells you what each costs — it does not choose for you.")
	return interp.String(b.String()), nil
}

// nearestExplainedPOI finds the nearest indexed POI of `typ` from (x,y) and
// returns its explained reach verdict. ok=false when no POI of that type is
// indexed.
func (h *Host) nearestExplainedPOI(x, y int, typ string, cap worldmap.Capability) (searchMapHit, bool) {
	bestDist := 1 << 30
	var best searchMapHit
	found := false
	low := strings.ToLower(typ)
	for _, p := range h.worldOracle.POIs() {
		if p.Comp < 0 || !strings.Contains(strings.ToLower(p.Type), low) {
			continue
		}
		d := chebyshev(x, y, p.SnapX, p.SnapY)
		if d < bestDist {
			bestDist = d
			best = searchMapHit{label: h.destinationLabel(p), x: p.SnapX, y: p.SnapY, dist: d}
			found = true
		}
	}
	if !found {
		return searchMapHit{}, false
	}
	best.info = h.worldOracle.ExplainReach(x, y, best.x, best.y, cap)
	return best, true
}

// destinationLabel builds a human label for a POI: its nearest named region +
// the POI type, e.g. "Al-Kharid mining-site". Falls back to the bare type when
// no region is near. Gives the host a region/landmark name to reason with
// instead of bare coordinates.
func (h *Host) destinationLabel(p worldmap.IndexedPOI) string {
	region := ""
	if h.facts != nil {
		if lm, _, ok := h.facts.Gazetteer().NearestPlace(p.SnapX, p.SnapY); ok {
			region = lm.Name
		}
	}
	if region != "" {
		return region + " " + p.Type
	}
	return p.Type
}

// dirOf is the 8-point compass bearing from (fx,fy) to (tx,ty), or "here".
func dirOf(fx, fy, tx, ty int) string {
	if b := facts.Bearing(fx, fy, tx, ty); b != "" {
		return b
	}
	return "here"
}

// gateShort trims a verbose transport-edge gate name to its leading phrase
// before the first '(' so survey prose stays readable (the raw names carry a
// parenthetical source/coord note).
func gateShort(name string) string {
	if name == "" {
		return "a gate"
	}
	if i := strings.IndexByte(name, '('); i > 0 {
		if s := strings.TrimSpace(name[:i]); s != "" {
			return s
		}
	}
	return name
}
