package runtime

import (
	"fmt"
	"sort"
	"strings"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// LocationSummary describes WHERE THE HOST IS in the world in readable
// terms — its nearest named area plus a few notable POIs with bearing +
// distance — so a host can reason about its place on the map instead of
// raw tile coords. Backed by the facts gazetteer (RSC+ map data).
func (h *Host) LocationSummary() string {
	pos := h.world.Self.Position()
	if h.facts == nil {
		return fmt.Sprintf("at (%d, %d)", pos.X, pos.Y)
	}
	// Planes stack in Y at 944-tile bands; the gazetteer (and every
	// human-meaningful place name) lives in plane-0 FOOTPRINT coords. An
	// upstairs host must query with its footprint Y or the nearest "place"
	// is hundreds of tiles of nonsense ("near Kharazi Jungle" upstairs in
	// Lumbridge) — and it must KNOW it is upstairs.
	plane := world.PlaneOf(pos.Y)
	fy := pos.Y % world.PlaneHeight
	floor := ""
	switch plane {
	case 1:
		floor = ", UPSTAIRS (floor 1 — use a ladder/staircase down to reach ground level)"
	case 2:
		floor = ", UPSTAIRS (floor 2 — two climbs down to ground level)"
	case 3:
		floor = ", UNDERGROUND (climb up to reach the surface)"
	}
	g := h.facts.Gazetteer()
	var b strings.Builder
	if p, d, ok := g.NearestPlace(pos.X, fy); ok && d <= 3 {
		fmt.Fprintf(&b, "at %s (%d, %d)%s", p.Name, pos.X, pos.Y, floor)
	} else if ok {
		fmt.Fprintf(&b, "at (%d, %d)%s, near %s (%d tiles %s)", pos.X, pos.Y, floor, p.Name, d, facts.Bearing(pos.X, fy, p.X, p.Y))
	} else {
		fmt.Fprintf(&b, "at (%d, %d)%s", pos.X, pos.Y, floor)
	}
	// Nearest POI per type within ~25 tiles, closest few (footprint
	// coords; POIs are ground-level features, so skip them upstairs —
	// they are not reachable without descending first).
	type near struct {
		t       string
		d, x, y int
	}
	best := map[string]near{}
	if plane == 0 {
		for _, p := range g.POIsWithin(pos.X, fy, 25) {
			d := chebyshev(pos.X, fy, p.X, p.Y)
			if cur, ok := best[p.Type]; !ok || d < cur.d {
				best[p.Type] = near{p.Type, d, p.X, p.Y}
			}
		}
	}
	if len(best) > 0 {
		list := make([]near, 0, len(best))
		for _, n := range best {
			list = append(list, n)
		}
		sort.Slice(list, func(i, j int) bool { return list[i].d < list[j].d })
		if len(list) > 5 {
			list = list[:5]
		}
		parts := make([]string, 0, len(list))
		for _, n := range list {
			parts = append(parts, fmt.Sprintf("%s %d %s", n.t, n.d, facts.Bearing(pos.X, pos.Y%world.PlaneHeight, n.x, n.y)))
		}
		fmt.Fprintf(&b, ". Nearby: %s", strings.Join(parts, ", "))
	}
	return b.String()
}

// Examination is one thing the bot is currently aware of and can
// describe — an NPC, a ground item, a scenery prop, a player, etc.
// The Kind/Name/Description fields mirror what the original RSC
// client would show when you right-click "Examine" on the same
// thing. Position is in absolute world coords.
//
// LLM-driven brains consume Examination batches as the bot's
// "current observation"; nothing about Examine sends a network
// packet — every field is resolved from facts + world-state mirrors.
type Examination struct {
	Kind        string // "npc", "player", "ground_item", "inventory_item", "scenery", "boundary", "self"
	Name        string
	Description string
	X, Y        int
	// Detail holds optional extra context like "Lvl 5 Goblin (5/5 HP)"
	// — varies per Kind.
	Detail string
}

func (e Examination) String() string {
	bits := []string{e.Kind}
	if e.Name != "" {
		bits = append(bits, e.Name)
	}
	if e.X != 0 || e.Y != 0 {
		bits = append(bits, fmt.Sprintf("(%d, %d)", e.X, e.Y))
	}
	if e.Detail != "" {
		bits = append(bits, e.Detail)
	}
	if e.Description != "" {
		bits = append(bits, "— "+e.Description)
	}
	return strings.Join(bits, " ")
}

// ExamineNpc returns the description of the NPC at `serverIndex` as
// the bot currently perceives it. Returns the zero Examination if the
// NPC isn't in our visible-NPCs mirror.
func (h *Host) ExamineNpc(serverIndex int) Examination {
	rec, ok := h.world.Npcs.Get(serverIndex)
	if !ok || h.facts == nil {
		return Examination{Kind: "npc", Detail: fmt.Sprintf("unknown NPC (idx %d)", serverIndex)}
	}
	def := h.facts.NpcDef(rec.TypeID)
	if def == nil {
		return Examination{
			Kind:   "npc",
			X:      rec.X,
			Y:      rec.Y,
			Detail: fmt.Sprintf("unknown type %d (idx %d)", rec.TypeID, serverIndex),
		}
	}
	detail := []string{}
	if def.Attackable {
		// Combat level + threat relative to us — the "darker red = more
		// dangerous" cue, brain-readable. Plus the raw stat block.
		cl := (def.Attack+def.Strength+def.Defense)/4 + def.Hits/4
		threat, _ := threatBand(h.world.Self.CombatLevel(), cl)
		detail = append(detail, fmt.Sprintf("combat %d (%s)", cl, threat))
		detail = append(detail, fmt.Sprintf("atk=%d str=%d def=%d hp=%d", def.Attack, def.Strength, def.Defense, def.Hits))
		if def.Aggressive {
			detail = append(detail, "aggressive")
		}
	}
	return Examination{
		Kind:        "npc",
		Name:        def.Name,
		Description: def.Description,
		X:           rec.X,
		Y:           rec.Y,
		Detail:      strings.Join(detail, ", "),
	}
}

// ExamineItem returns the description of an item by catalog ID. Works
// for both inventory items and ground items.
func (h *Host) ExamineItem(itemID int) Examination {
	if h.facts == nil {
		return Examination{Kind: "item", Detail: fmt.Sprintf("itemID=%d (facts unavailable)", itemID)}
	}
	def := h.facts.ItemDef(itemID)
	if def == nil {
		return Examination{Kind: "item", Detail: fmt.Sprintf("unknown itemID=%d", itemID)}
	}
	detail := []string{}
	if def.IsStackable {
		detail = append(detail, "stackable")
	}
	if def.IsMembersOnly {
		detail = append(detail, "members-only")
	}
	if def.IsWearable {
		detail = append(detail, "wearable")
	}
	if def.Command != "" {
		detail = append(detail, "default-action: "+def.Command)
	}
	return Examination{
		Kind:        "item",
		Name:        def.Name,
		Description: def.Description,
		Detail:      strings.Join(detail, ", "),
	}
}

// ExamineInventorySlot describes what's in inventory slot N.
func (h *Host) ExamineInventorySlot(slot int) Examination {
	slots := h.world.Inventory.Slots()
	if slot < 0 || slot >= len(slots) || slots[slot].ItemID == 0 {
		return Examination{Kind: "inventory_item", Detail: fmt.Sprintf("slot %d empty", slot)}
	}
	s := slots[slot]
	ex := h.ExamineItem(s.ItemID)
	ex.Kind = "inventory_item"
	if s.Amount > 1 {
		ex.Detail = fmt.Sprintf("x%d", s.Amount) + ifNotEmpty(", ", ex.Detail)
	}
	if s.Wielded {
		ex.Detail = "wielded" + ifNotEmpty(", ", ex.Detail)
	}
	return ex
}

// ExamineGroundItem looks up a ground item at the given tile and
// returns its description. Returns empty Examination if no item
// there.
func (h *Host) ExamineGroundItem(x, y int) Examination {
	for _, rec := range h.world.GroundItems.All() {
		if rec.X == x && rec.Y == y {
			ex := h.ExamineItem(rec.ItemID)
			ex.Kind = "ground_item"
			ex.X = x
			ex.Y = y
			return ex
		}
	}
	return Examination{Kind: "ground_item", X: x, Y: y, Detail: "no item at this tile"}
}

// ExamineScenery returns the scenery prop at (x, y), if any. Pulls
// from the static facts placement database so it works even before
// the bot first perceives the tile.
func (h *Host) ExamineScenery(x, y int) Examination {
	if h.facts == nil {
		return Examination{Kind: "scenery", X: x, Y: y, Detail: "facts unavailable"}
	}
	placements := h.facts.At(x, y)
	for _, p := range placements {
		if p.Kind != "scenery" {
			continue
		}
		def := h.facts.SceneryDef(p.DefID)
		ex := Examination{Kind: "scenery", X: x, Y: y, Name: p.Name}
		if def != nil {
			ex.Description = def.Description
		}
		return ex
	}
	return Examination{Kind: "scenery", X: x, Y: y, Detail: "no scenery at this tile"}
}

// ExamineBoundary returns the boundary def at (x, y), if any.
func (h *Host) ExamineBoundary(x, y int) Examination {
	if h.facts == nil {
		return Examination{Kind: "boundary", X: x, Y: y, Detail: "facts unavailable"}
	}
	for _, p := range h.facts.At(x, y) {
		if p.Kind != "boundary" {
			continue
		}
		def := h.facts.BoundaryDefs[p.DefID]
		ex := Examination{Kind: "boundary", X: x, Y: y, Name: p.Name}
		if def != nil {
			ex.Description = def.Description
			if def.BlocksMovement() {
				ex.Detail = "blocks movement"
			} else if def.Unknown == 1 {
				ex.Detail = "openable"
			}
		}
		return ex
	}
	return Examination{Kind: "boundary", X: x, Y: y, Detail: "no boundary at this tile"}
}

// ExaminePlayer returns the description of a visible player.
func (h *Host) ExaminePlayer(serverIndex int) Examination {
	rec, ok := h.world.Players.Get(serverIndex)
	if !ok {
		return Examination{Kind: "player", Detail: fmt.Sprintf("unknown player (idx %d)", serverIndex)}
	}
	// Describe what they're wearing/wielding in NAMES (resolved from the
	// appearance packet via the runtime equipment API), so a host that
	// "looks at someone" hands the brain readable gear, not raw ids. See
	// runtime/equipment.go.
	parts := []string{}
	if rec.HasAppearanceCombat {
		threat, _ := threatBand(h.world.Self.CombatLevel(), rec.CombatLevel)
		parts = append(parts, fmt.Sprintf("combat %d (%s)", rec.CombatLevel, threat))
		if rec.SkullType != 0 {
			parts = append(parts, "skulled")
		}
	}
	if worn := h.PlayerEquipment(rec.EquipBySlot); len(worn) > 0 {
		gear := make([]string, 0, len(worn))
		for _, w := range worn {
			gear = append(gear, w.SlotName+": "+w.Label())
		}
		parts = append(parts, "wearing "+strings.Join(gear, ", "))
	}
	detail := strings.Join(parts, "; ")
	return Examination{
		Kind:        "player",
		Name:        rec.Name,
		Description: "A fellow adventurer.",
		X:           rec.X,
		Y:           rec.Y,
		Detail:      detail,
	}
}

// ExamineSelf returns a self-description for the bot itself —
// position, HP/fatigue, combat level, inventory summary. The LLM
// brain reads this whenever it needs to ground its reasoning in the
// agent's current state.
func (h *Host) ExamineSelf() Examination {
	pos := h.world.Self.Position()
	hp := h.world.Self.HP()
	maxHP := h.world.Self.MaxHP()
	cb := h.world.Self.CombatLevel()
	fatigue := h.world.Self.FatiguePercent() // 0..100% for the brain (raw is 0..750)
	inv := h.world.Inventory.Slots()
	used := 0
	for _, s := range inv {
		if s.ItemID != 0 {
			used++
		}
	}
	return Examination{
		Kind: "self",
		Name: h.opts.Username,
		X:    pos.X,
		Y:    pos.Y,
		Detail: fmt.Sprintf("hp=%d/%d cb=%d fatigue=%d%% inv=%d/%d",
			hp, maxHP, cb, fatigue, used, len(inv)),
	}
}

// DescribeSurroundings produces a multi-line text report of the
// bot's current perception within `radius` tiles — the canonical
// observation format the LLM brain consumes.
//
// Categories included:
//   - self (position, hp, inventory used)
//   - nearby NPCs (with combat stats)
//   - nearby players (name + position)
//   - nearby ground items
//   - notable static scenery (from facts, filtered to "interesting" types
//     like wells, signs, ladders, doors)
func (h *Host) DescribeSurroundings(radius int) string {
	if radius <= 0 {
		radius = 8
	}
	pos := h.world.Self.Position()
	var b strings.Builder
	fmt.Fprintf(&b, "Self: %s\n", h.ExamineSelf())
	fmt.Fprintf(&b, "Location: %s\n", h.LocationSummary())

	// NPCs.
	nearbyNpcs := []Examination{}
	for _, n := range h.world.Npcs.All() {
		if absInt(n.X-pos.X) > radius || absInt(n.Y-pos.Y) > radius {
			continue
		}
		nearbyNpcs = append(nearbyNpcs, h.ExamineNpc(n.Index))
	}
	sort.Slice(nearbyNpcs, func(i, j int) bool {
		return distSq(nearbyNpcs[i], pos.X, pos.Y) < distSq(nearbyNpcs[j], pos.X, pos.Y)
	})
	if len(nearbyNpcs) > 0 {
		b.WriteString("Nearby NPCs:\n")
		for _, n := range nearbyNpcs {
			fmt.Fprintf(&b, "  - %s\n", n)
		}
	}

	// Players.
	players := []Examination{}
	for _, p := range h.world.Players.All() {
		if absInt(p.X-pos.X) > radius || absInt(p.Y-pos.Y) > radius {
			continue
		}
		players = append(players, h.ExaminePlayer(p.Index))
	}
	if len(players) > 0 {
		b.WriteString("Nearby players:\n")
		for _, p := range players {
			fmt.Fprintf(&b, "  - %s\n", p)
		}
	}

	// Ground items.
	gi := h.world.GroundItems.Near(pos.X, pos.Y, radius)
	if len(gi) > 0 {
		b.WriteString("Nearby ground items:\n")
		for _, item := range gi {
			ex := h.ExamineItem(item.ItemID)
			fmt.Fprintf(&b, "  - %s at (%d, %d)\n", ex, item.X, item.Y)
		}
	}

	// Scenery / boundaries from facts (filter to interesting kinds).
	if h.facts != nil {
		nearby := h.facts.Near(pos.X, pos.Y, radius)
		interesting := []facts.Placement{}
		for _, p := range nearby {
			switch p.Kind {
			case "boundary":
				interesting = append(interesting, p)
			case "scenery":
				if isNotableScenery(p.Name) {
					interesting = append(interesting, p)
				}
			}
		}
		if len(interesting) > 0 {
			b.WriteString("Notable scenery:\n")
			for _, p := range interesting {
				fmt.Fprintf(&b, "  - %s %q at (%d, %d)\n", p.Kind, p.Name, p.X, p.Y)
			}
		}
	}
	return b.String()
}

// All returns helpers for the Players state — currently used only by
// DescribeSurroundings. Lives here so we don't have to add a method
// to world.PlayersState just for this.
func (h *Host) playersList() []playerSnap { return nil } // placeholder

type playerSnap struct{ X, Y, Index int }

func distSq(e Examination, x, y int) int {
	dx, dy := e.X-x, e.Y-y
	return dx*dx + dy*dy
}

func ifNotEmpty(sep, s string) string {
	if s == "" {
		return ""
	}
	return sep + s
}

// isNotableScenery filters scenery-loc reports to things the LLM
// would likely care about (wells, ladders, fountains, signs, etc.)
// rather than every blade of grass. Generous — the LLM can ignore
// what it doesn't need.
func isNotableScenery(name string) bool {
	n := strings.ToLower(name)
	switch n {
	case "tree":
		return false
	}
	for _, prefix := range []string{
		"well", "ladder", "stair", "altar", "shrine", "fountain",
		"sign", "counter", "anvil", "furnace", "bench", "table",
		"bed", "chest", "fishing", "spinning", "loom", "ovens",
		"banker", "door", "gate", "fence", "rock", "ore",
	} {
		if strings.HasPrefix(n, prefix) {
			return true
		}
	}
	return false
}
