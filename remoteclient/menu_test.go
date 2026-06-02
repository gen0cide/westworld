package remoteclient

import (
	"testing"

	"github.com/gen0cide/westworld/facts"
	"github.com/gen0cide/westworld/world"
)

// menu_test.go is the deterministic unit suite for BuildMenu — the pure verb-list
// builder at the heart of Layer 2. It is table-driven: each row is one def shape
// (a monster / shop NPC / thievable NPC; a tree / door / plain wall; a food /
// weapon / worn item) and asserts the EXACT ordered (opts, ids) BuildMenu emits,
// proving the de-dup (Examine never doubles, a wall's "WalkTo" drops, an NPC's
// "Attack" Command suppresses the synthetic Attack) and the Wield-fold (a worn
// item shows Remove not Wield; a wield-Command armour shows its own label once).
//
// These two slices are the frozen contract /pick (sends opts) and /act (re-derives
// ids) both index, so a drift here would silently mis-route a click — hence the
// goldens are spelled out verbatim rather than computed.

// verbs extracts the ordered verb labels from a []MenuOption for compact asserts.
func verbs(opts []MenuOption) []string {
	out := make([]string, len(opts))
	for i, o := range opts {
		out[i] = o.Verb
	}
	return out
}

func eqStr(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func eqIDs(a, b []OptionID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestBuildMenu(t *testing.T) {
	tests := []struct {
		name      string
		kind      TargetKind
		defs      MenuDefs
		wantVerbs []string
		wantIDs   []OptionID
	}{
		// --- NPC shapes ---
		{
			name: "npc monster: Attack Command1 suppresses synthetic Attack",
			kind: KindNPC,
			defs: MenuDefs{Npc: &facts.NpcDef{Command1: "Attack", Attackable: true}},
			// Command1 "Attack" is named, so the fallback synthetic Attack is dropped.
			wantVerbs: []string{"Attack", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptExamine},
		},
		{
			name: "npc shopkeeper: Talk-to then synthetic Examine, no Attack",
			kind: KindNPC,
			defs: MenuDefs{Npc: &facts.NpcDef{Command1: "Talk-to", Attackable: false}},
			wantVerbs: []string{"Talk-to", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptExamine},
		},
		{
			name: "npc thievable+attackable: Command1 Pickpocket then fallback Attack",
			kind: KindNPC,
			defs: MenuDefs{Npc: &facts.NpcDef{Command1: "Pickpocket", Attackable: true}},
			// Command1 is Pickpocket (not Attack), so the attackable fallback adds Attack.
			wantVerbs: []string{"Pickpocket", "Attack", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptAttack, OptExamine},
		},
		{
			name: "npc two commands plus attack",
			kind: KindNPC,
			defs: MenuDefs{Npc: &facts.NpcDef{Command1: "Talk-to", Command2: "Pickpocket", Attackable: true}},
			wantVerbs: []string{"Talk-to", "Pickpocket", "Attack", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptCommand2, OptAttack, OptExamine},
		},
		{
			name:      "npc nil def collapses to Examine",
			kind:      KindNPC,
			defs:      MenuDefs{},
			wantVerbs: []string{"Examine"},
			wantIDs:   []OptionID{OptExamine},
		},
		{
			name: "npc non-attackable with no commands gets synthetic Talk-to (e.g. shopkeeper)",
			kind: KindNPC,
			// NPC def present but Command1/Command2 are empty and not attackable —
			// exactly the OpenRSC shopkeeper (id 51) shape.
			defs:      MenuDefs{Npc: &facts.NpcDef{Command1: "", Command2: "", Attackable: false}},
			wantVerbs: []string{"Talk-to", "Examine"},
			wantIDs:   []OptionID{OptTalkTo, OptExamine},
		},

		// --- Player (fixed menu, no per-def verbs) ---
		{
			name:      "player fixed standard menu",
			kind:      KindPlayer,
			defs:      MenuDefs{},
			wantVerbs: []string{"Trade with", "Follow", "Duel with", "Attack", "Examine"},
			wantIDs:   []OptionID{OptTrade, OptFollow, OptDuel, OptAttack, OptExamine},
		},

		// --- Self ---
		{
			name:      "self is just Examine",
			kind:      KindSelf,
			defs:      MenuDefs{},
			wantVerbs: []string{"Examine"},
			wantIDs:   []OptionID{OptExamine},
		},

		// --- Ground item ---
		{
			name:      "ground item Pick up then Examine",
			kind:      KindGroundItem,
			defs:      MenuDefs{},
			wantVerbs: []string{"Pick up", "Examine"},
			wantIDs:   []OptionID{OptPickup, OptExamine},
		},

		// --- Scenery shapes ---
		{
			name: "scenery tree: Chop then synthetic Examine, def Examine dropped",
			kind: KindScenery,
			defs: MenuDefs{Scenery: &facts.SceneryDef{Command1: "Chop", Command2: "Examine"}},
			// Command2 "Examine" is dropped so Examine never doubles.
			wantVerbs: []string{"Chop", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptExamine},
		},
		{
			name: "scenery two real commands plus synthetic Examine",
			kind: KindScenery,
			defs: MenuDefs{Scenery: &facts.SceneryDef{Command1: "Search", Command2: "Open"}},
			wantVerbs: []string{"Search", "Open", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptCommand2, OptExamine},
		},
		{
			name:      "scenery nil def collapses to Examine",
			kind:      KindScenery,
			defs:      MenuDefs{},
			wantVerbs: []string{"Examine"},
			wantIDs:   []OptionID{OptExamine},
		},

		// --- Boundary shapes ---
		{
			name: "boundary door: Open then synthetic Examine",
			kind: KindBoundary,
			defs: MenuDefs{Boundary: &facts.BoundaryDef{Command1: "Open", Command2: "Examine"}},
			wantVerbs: []string{"Open", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptExamine},
		},
		{
			name: "boundary plain wall: WalkTo drops, collapses to Examine",
			kind: KindBoundary,
			defs: MenuDefs{Boundary: &facts.BoundaryDef{Command1: "WalkTo", Command2: "Examine"}},
			// Command1 "WalkTo" is a no-op and dropped; Command2 "Examine" dropped too.
			wantVerbs: []string{"Examine"},
			wantIDs:   []OptionID{OptExamine},
		},
		{
			name: "boundary gate: Open + real Command2 Close both kept",
			kind: KindBoundary,
			defs: MenuDefs{Boundary: &facts.BoundaryDef{Command1: "Open", Command2: "Close"}},
			// Command2 "Close" is a real verb (not "Examine"/"WalkTo") so it stays;
			// both commands route to the same InteractWithBoundary packet at dispatch.
			wantVerbs: []string{"Open", "Close", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptCommand2, OptExamine},
		},
		{
			name:      "boundary nil def collapses to Examine",
			kind:      KindBoundary,
			defs:      MenuDefs{},
			wantVerbs: []string{"Examine"},
			wantIDs:   []OptionID{OptExamine},
		},
		{
			name: "scenery Command1 only (no Command2): Mine then Examine",
			kind: KindScenery,
			defs: MenuDefs{Scenery: &facts.SceneryDef{Command1: "Mine"}},
			wantVerbs: []string{"Mine", "Examine"},
			wantIDs:   []OptionID{OptCommand1, OptExamine},
		},

		// --- Terrain ---
		{
			name:      "terrain is Walk here only",
			kind:      KindTerrain,
			defs:      MenuDefs{},
			wantVerbs: []string{"Walk here"},
			wantIDs:   []OptionID{OptWalkHere},
		},

		// --- Inventory shapes ---
		{
			name: "inventory food: Command Eat, Drop, Examine",
			kind: KindInventoryItem,
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Eat"}},
			wantVerbs: []string{"Eat", "Drop", "Examine"},
			wantIDs:   []OptionID{OptCommand, OptDrop, OptExamine},
		},
		{
			name: "inventory un-worn weapon: Wield-fold uses def label once",
			kind: KindInventoryItem,
			// Command is the wield verb "Wield" -> folded into the single Wield entry.
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Wield", IsWearable: true}, InvSlot: world.InvSlot{Wielded: false}},
			wantVerbs: []string{"Wield", "Drop", "Examine"},
			wantIDs:   []OptionID{OptWield, OptDrop, OptExamine},
		},
		{
			name: "inventory un-worn armour: Wear label preserved, no generic Wield double",
			kind: KindInventoryItem,
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Wear", IsWearable: true}, InvSlot: world.InvSlot{Wielded: false}},
			wantVerbs: []string{"Wear", "Drop", "Examine"},
			wantIDs:   []OptionID{OptWield, OptDrop, OptExamine},
		},
		{
			name: "inventory worn item: Remove not Wield",
			kind: KindInventoryItem,
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Wield", IsWearable: true}, InvSlot: world.InvSlot{Wielded: true}},
			wantVerbs: []string{"Remove", "Drop", "Examine"},
			wantIDs:   []OptionID{OptRemove, OptDrop, OptExamine},
		},
		{
			name: "inventory wearable with no command: generic Wield",
			kind: KindInventoryItem,
			defs: MenuDefs{Item: &facts.ItemDef{IsWearable: true}, InvSlot: world.InvSlot{Wielded: false}},
			wantVerbs: []string{"Wield", "Drop", "Examine"},
			wantIDs:   []OptionID{OptWield, OptDrop, OptExamine},
		},
		{
			name: "inventory plain item: just Drop, Examine",
			kind: KindInventoryItem,
			defs: MenuDefs{Item: &facts.ItemDef{}},
			wantVerbs: []string{"Drop", "Examine"},
			wantIDs:   []OptionID{OptDrop, OptExamine},
		},
		{
			name: "inventory wearable with non-wield Command: Command AND Wield both kept",
			kind: KindInventoryItem,
			// A wearable whose Command is a real action (not Wield/Wear) keeps the
			// Command entry AND adds the synthetic Wield — no fold (fold only applies
			// when Command IS the wield verb).
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Operate", IsWearable: true}, InvSlot: world.InvSlot{Wielded: false}},
			wantVerbs: []string{"Operate", "Wield", "Drop", "Examine"},
			wantIDs:   []OptionID{OptCommand, OptWield, OptDrop, OptExamine},
		},
		{
			name: "inventory worn item with non-wield Command: Command then Remove",
			kind: KindInventoryItem,
			// Worn wearable: Command kept, Wield suppressed (already worn), Remove added.
			defs: MenuDefs{Item: &facts.ItemDef{Command: "Operate", IsWearable: true}, InvSlot: world.InvSlot{Wielded: true}},
			wantVerbs: []string{"Operate", "Remove", "Drop", "Examine"},
			wantIDs:   []OptionID{OptCommand, OptRemove, OptDrop, OptExamine},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts, ids := BuildMenu(tc.kind, tc.defs)

			// opts[i].ID must equal i (the dispatch index contract).
			for i, o := range opts {
				if o.ID != i {
					t.Errorf("opts[%d].ID = %d, want %d (id must equal slice index)", i, o.ID, i)
				}
			}
			// opts and ids must be the same length (parallel slices).
			if len(opts) != len(ids) {
				t.Fatalf("len(opts)=%d != len(ids)=%d (parallel slices must align)", len(opts), len(ids))
			}
			if got := verbs(opts); !eqStr(got, tc.wantVerbs) {
				t.Errorf("verbs = %v, want %v", got, tc.wantVerbs)
			}
			if !eqIDs(ids, tc.wantIDs) {
				t.Errorf("ids = %v, want %v", ids, tc.wantIDs)
			}
		})
	}
}

// TestBuildMenuExamineAlwaysLast asserts the structural invariant that, for every
// kind that carries an Examine entry, Examine is the LAST option (so the menu
// matches RSC's layout and never buries Examine mid-list).
func TestBuildMenuExamineAlwaysLast(t *testing.T) {
	kinds := []struct {
		kind TargetKind
		defs MenuDefs
	}{
		{KindNPC, MenuDefs{Npc: &facts.NpcDef{Command1: "Talk-to", Attackable: true}}},
		{KindPlayer, MenuDefs{}},
		{KindSelf, MenuDefs{}},
		{KindGroundItem, MenuDefs{}},
		{KindScenery, MenuDefs{Scenery: &facts.SceneryDef{Command1: "Chop"}}},
		{KindBoundary, MenuDefs{Boundary: &facts.BoundaryDef{Command1: "Open"}}},
		{KindInventoryItem, MenuDefs{Item: &facts.ItemDef{Command: "Eat"}}},
	}
	for _, k := range kinds {
		opts, ids := BuildMenu(k.kind, k.defs)
		if len(ids) == 0 {
			t.Errorf("kind %q produced no options", k.kind)
			continue
		}
		if ids[len(ids)-1] != OptExamine {
			t.Errorf("kind %q: last option = %q, want examine; verbs=%v", k.kind, ids[len(ids)-1], verbs(opts))
		}
	}
}
