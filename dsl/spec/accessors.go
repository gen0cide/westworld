package spec

// AccessorSpec documents one query-layer attribute path.
//
// Unlike Actions and Events (where the spec drives behavior), the
// accessor implementations live in runtime/dsl_views.go as Getter
// switch arms. The spec here is the DOCUMENTATION + DISCOVERY +
// CONSISTENCY source: every path listed is what hosts should be
// able to read, and the consistency test asserts each path resolves
// against a stub host.
//
// Adding a query: add a row here AND extend the relevant view's
// Get() method in runtime/dsl_views.go.
type AccessorSpec struct {
	// Path is the dotted accessor as a slice. E.g.
	// `self.hp` is ["self", "hp"]; `self.skills.fishing.level`
	// is ["self", "skills", "fishing", "level"].
	Path []string

	// Kind describes the returned value's type (informally — used
	// by tooling and docs, not enforced by the validator). Examples:
	// "int", "float", "string", "bool", "Position", "Inventory",
	// "list<NpcView>", "Error?".
	Kind string

	// DocSummary is one line about what the accessor returns.
	DocSummary string

	// NotYetImplemented marks an accessor the spec defines but
	// runtime/dsl_views.go doesn't yet expose. The consistency
	// test uses this to distinguish "spec drift" from "planned
	// but not built."
	NotYetImplemented bool
}

// Accessors is the canonical query-layer surface. Use the helper
// constructors below to keep entries terse.
//
// Currently this list is partial — it covers what runtime/dsl_views.go
// implements today. The Phase 2.5 query layer build-out (tasks
// #56–#65) extends this to ~100 entries; that work adds spec rows
// in lock-step with the view implementations.
//
// Naming convention reminder (see docs/lang/syntax.md):
//   - Booleans prefix `is_` (is_busy, is_in_combat)
//   - Maximums prefix `max_` (max_hp, max_prayer)
//   - Plain noun for counts (free, length, level)
var Accessors = []AccessorSpec{
	// ===== self — first-person game state =====
	{Path: []string{"self", "position"}, Kind: "Position",
		DocSummary: "Current tile {x, y, plane}."},
	{Path: []string{"self", "hp"}, Kind: "int",
		DocSummary: "Current hits (boostable)."},
	{Path: []string{"self", "max_hp"}, Kind: "int",
		DocSummary: "Base hits skill level."},
	{Path: []string{"self", "prayer"}, Kind: "int",
		DocSummary: "Current prayer points."},
	{Path: []string{"self", "max_prayer"}, Kind: "int",
		DocSummary: "Base prayer skill level."},
	{Path: []string{"self", "fatigue"}, Kind: "int",
		DocSummary: "Fatigue (0-100; RSC sleep mechanic threshold)."},
	{Path: []string{"self", "combat_level"}, Kind: "int",
		DocSummary: "Derived combat level."},
	{Path: []string{"self", "wielded"}, Kind: "ItemView?",
		DocSummary: "Currently-wielded weapon, or null if unarmed."},
	{Path: []string{"self", "skills"}, Kind: "SkillsView",
		DocSummary: "Skill access: self.skills.<name>.level / .max_level / .xp / .xp_to_next_level / .percent_to_next_level."},

	// ===== inventory =====
	{Path: []string{"inventory", "free"}, Kind: "int",
		DocSummary: "Number of empty inventory slots."},
	{Path: []string{"inventory", "slots"}, Kind: "list<ItemView>",
		DocSummary: "Every inventory slot (item-view with .item_id, .amount, .is_wielded)."},
	{Path: []string{"inventory", "has"}, Kind: "callable(item)->bool",
		DocSummary: "Returns true iff the item is present (id, name, or item-view)."},
	{Path: []string{"inventory", "count"}, Kind: "callable(item)->int",
		DocSummary: "Returns total amount of the named item across slots."},

	// ===== combat (mostly stubs today) =====
	{Path: []string{"combat", "engaged"}, Kind: "bool",
		DocSummary: "Currently in combat?"},
	{Path: []string{"combat", "target"}, Kind: "Entity?",
		DocSummary: "Current combat target (NPC or player), or null."},

	// ===== world — visible entities =====
	{Path: []string{"world", "players"}, Kind: "list<PlayerView>",
		DocSummary: "Players visible in our view radius."},
	{Path: []string{"world", "npcs"}, Kind: "list<NpcView>",
		DocSummary: "NPCs visible in our view radius."},
	{Path: []string{"world", "ground_items"}, Kind: "list<GroundItemView>",
		DocSummary: "Ground items visible in our view radius."},
	{Path: []string{"world", "locs"}, Kind: "LocsView",
		DocSummary: "Static facts-derived locations: banks, altars, fishing_spots, etc."},

	// ----- planned by docs/lang/state.md Build Plan, not yet
	// wired in runtime/dsl_views.go (tasks #56–#65) -----
	{Path: []string{"self", "hp_fraction"}, Kind: "float", DocSummary: "hp / max_hp."},
	{Path: []string{"self", "quest_points"}, Kind: "int", DocSummary: "QP count."},
	{Path: []string{"self", "is_busy"}, Kind: "bool", DocSummary: "Currently performing an action? (stub: always false until action tracking lands)"},

	// Per-entity-view accessors (npc/player/ground_item) are not
	// rooted at a fixed top-level path — they live ON the views
	// returned from world.npcs / world.players / world.ground_items.
	// Documentation for those fields lives in docs/lang/state.md
	// "Entity-views" section, not as spec.Accessors rows. The view
	// implementations in runtime/dsl_views.go are the canonical
	// source; new fields land there + in state.md.
	{Path: []string{"self", "is_in_combat"}, Kind: "bool", DocSummary: "Engaged with an NPC or player? (stub: always false until combat tracking lands)"},
	{Path: []string{"self", "is_sleeping"}, Kind: "bool", DocSummary: "Sleep screen up due to fatigue? (stub: always false until sleep tracking lands)"},
	{Path: []string{"self", "equipped"}, Kind: "list<ItemView>",
		DocSummary: "List of all currently-wielded items. Per-slot accessors (.weapon, .shield, etc.) need the equipment-by-slot packet decoder — until then, filter the list (e.g. self.equipped.find(i => i.is_wearable))."},
	{Path: []string{"self", "prayer", "active"}, Kind: "list<Prayer>", DocSummary: "Currently active prayers.", NotYetImplemented: true},
	{Path: []string{"self", "spells", "known"}, Kind: "list<Spell>", DocSummary: "Spells with sufficient level.", NotYetImplemented: true},
	{Path: []string{"inventory", "weight"}, Kind: "int", DocSummary: "Sum of slot weights.", NotYetImplemented: true},
	{Path: []string{"inventory", "is_full"}, Kind: "bool", DocSummary: "True iff free == 0."},
	{Path: []string{"inventory", "find"}, Kind: "callable(item)->ItemView?", DocSummary: "First matching slot, or null."},
	{Path: []string{"inventory", "slot_of"}, Kind: "callable(item)->int?", DocSummary: "Slot index of first matching item, or null."},
	{Path: []string{"world", "last_chat"}, Kind: "ChatRecord?", DocSummary: "Most recent public chat: .speaker / .message / .at."},
	{Path: []string{"world", "last_pm"}, Kind: "PMRecord?", DocSummary: "Most recent private message: .sender / .message / .at."},
	{Path: []string{"world", "last_damage"}, Kind: "DamageRecord?", DocSummary: "Most recent damage: .amount / .source / .at."},
	{Path: []string{"world", "last_server_message"}, Kind: "ServerMsgRecord?", DocSummary: "Most recent server message: .message / .at."},
	{Path: []string{"world", "last_dialog_text"}, Kind: "DialogTextRecord?", DocSummary: "Most recent NPC speech-bubble: .text / .at."},
	{Path: []string{"world", "messages"}, Kind: "list<Message>", DocSummary: "Server-message log (§10: was world.last_server_message). Each Message has .text / .kind / .at / .contains(needle). Backed by the single-value ring today; multi-entry ring + `on message` is task #119."},

	// ===== Def/Instance instance fields (api.md §2) — InvSlot =====
	{Path: []string{"inventory", "find_all"}, Kind: "callable(item)->list<InvSlot>", DocSummary: "Every slot matching the item as an InvSlot instance (one per slot; non-stackables occupy several)."},

	// ----- ground_items.nearest + find_any (#117) -----
	// Convenience reads added by #117. All pure Views (no Action verbs):
	// world.ground_items.nearest is dual-mode (bare field => nearest to
	// self; called with a position => nearest to that position),
	// world.ground_items.most_valuable is a value-sorted selector, and
	// inventory.find_any collapses or-chains over a set of item refs.
	{Path: []string{"world", "ground_items", "nearest"}, Kind: "callable(pos?)->GroundItem?",
		DocSummary: "Nearest visible ground item to self (bare field), or to an explicit position when called: world.ground_items.nearest(pos). Null when none visible."},
	{Path: []string{"world", "ground_items", "most_valuable"}, Kind: "GroundItem?",
		DocSummary: "Highest-base-value visible ground item (by facts.ItemDef base value), or Null when none visible. Enables loot-most-valuable."},
	{Path: []string{"inventory", "find_any"}, Kind: "callable([item,...])->InvSlot?",
		DocSummary: "First inventory slot matching ANY of the given item ids/names, as an InvSlot instance, or Null. Collapses gem/food/axe or-chains."},

	// ===== promoted namespaced subsystems (api.md §6/§10) =====
	// These resolve through the top-level namespace views (trade /
	// bank / duel / magic / prayer). Action verbs are view-dispatched
	// callables that return Result<Null>; Views are pure reads.

	// ----- bank reads + bulk verbs (#117/#118) -----
	{Path: []string{"bank", "is_open"}, Kind: "bool", DocSummary: "Bank UI currently shown (state machine over opcodes 42/203)."},
	{Path: []string{"bank", "slots"}, Kind: "list<[id,amount]>", DocSummary: "Bank contents (one [item_id, amount] pair per occupied slot)."},
	{Path: []string{"bank", "used"}, Kind: "int", DocSummary: "Occupied bank slots."},
	{Path: []string{"bank", "max_size"}, Kind: "int", DocSummary: "Bank slot capacity (0 when the bank is closed)."},
	{Path: []string{"bank", "free"}, Kind: "int", DocSummary: "Empty bank slots (max_size - used)."},
	{Path: []string{"bank", "open"}, Kind: "callable(banker)->Result", DocSummary: "Open the bank UI via a banker NPC (§10)."},
	{Path: []string{"bank", "deposit"}, Kind: "callable(item,amount)->Result", DocSummary: "Deposit `amount` of the item into the open bank (§10)."},
	{Path: []string{"bank", "withdraw"}, Kind: "callable(item,amount)->Result", DocSummary: "Withdraw `amount` of the item from the open bank (§10)."},
	{Path: []string{"bank", "deposit_all"}, Kind: "callable(keep?)->Result", DocSummary: "Deposit every inventory item except those in the optional keep-list (item ids/names) (#117/#118)."},
	{Path: []string{"bank", "withdraw_all"}, Kind: "callable(item)->Result", DocSummary: "Withdraw the entire banked quantity of one item (#117/#118)."},
	{Path: []string{"bank", "withdraw_x"}, Kind: "callable(item,amount)->Result", DocSummary: "Withdraw a preset amount of one item, clamped to the banked quantity (#117/#118)."},
	{Path: []string{"bank", "close"}, Kind: "callable()->Result", DocSummary: "Close the bank UI (§10)."},
	// ----- shop (§8: shop.is_open / stock / price / buy / sell / close) -----
	{Path: []string{"shop", "is_open"}, Kind: "bool", DocSummary: "Shop UI currently shown."},
	{Path: []string{"shop", "slots"}, Kind: "list<ShopSlot>", DocSummary: "Shop catalogue: item_id / stock / base_stock."},
	{Path: []string{"shop", "stock"}, Kind: "callable(item)->Int", DocSummary: "Units of the item the shop currently stocks."},
	{Path: []string{"shop", "price"}, Kind: "callable(item)->Int", DocSummary: "Current buy price (gp) of the item at this shop."},
	{Path: []string{"shop", "buy"}, Kind: "callable(item,qty)->Result", DocSummary: "Buy qty of the item from the open shop."},
	{Path: []string{"shop", "sell"}, Kind: "callable(item,qty)->Result", DocSummary: "Sell qty of the item to the open shop."},
	{Path: []string{"shop", "close"}, Kind: "callable()->Result", DocSummary: "Close the shop window."},

	// ----- trade (§10: request/offer/accept/confirm/decline) -----
	{Path: []string{"trade", "is_active"}, Kind: "bool", DocSummary: "Trade UI currently shown."},
	{Path: []string{"trade", "phase"}, Kind: "string", DocSummary: "Trade phase (none/offer/confirm/…)."},
	{Path: []string{"trade", "my_offer"}, Kind: "list<[id,amount]>", DocSummary: "Items you have put up."},
	{Path: []string{"trade", "their_offer"}, Kind: "list<[id,amount]>", DocSummary: "The other party's items."},
	{Path: []string{"trade", "accepted"}, Kind: "bool", DocSummary: "You clicked Accept on the offer screen."},
	{Path: []string{"trade", "request"}, Kind: "callable(player)->Result", DocSummary: "Trade-request a player; absorbs the old trade_request+accept_trade (§10)."},
	{Path: []string{"trade", "offer"}, Kind: "callable(items)->Result", DocSummary: "Set/replace your offer ([id,amount] pairs) (§10)."},
	{Path: []string{"trade", "accept"}, Kind: "callable()->Result", DocSummary: "Accept the offer screen (screen 1); was confirm_trade (§10)."},
	{Path: []string{"trade", "confirm"}, Kind: "callable()->Result", DocSummary: "Accept the confirm screen (screen 2); was finalize_trade (§10)."},
	{Path: []string{"trade", "decline"}, Kind: "callable()->Result", DocSummary: "Decline/close the trade (§10)."},

	// ----- duel (§10: request/set_rules/stake/accept/confirm/decline) -----
	{Path: []string{"duel", "is_active"}, Kind: "bool", DocSummary: "Duel UI currently shown."},
	{Path: []string{"duel", "phase"}, Kind: "string", DocSummary: "Duel phase."},
	{Path: []string{"duel", "request"}, Kind: "callable(player)->Result", DocSummary: "Duel-request a player; absorbs duel_request+accept_duel (§10)."},
	{Path: []string{"duel", "set_rules"}, Kind: "callable(...)->Result", DocSummary: "Set the four rule toggles (retreat/magic/prayer/weapons) (§10)."},
	{Path: []string{"duel", "stake"}, Kind: "callable(items)->Result", DocSummary: "Stake items ([id,amount] pairs); was offer_duel (§10)."},
	{Path: []string{"duel", "accept"}, Kind: "callable()->Result", DocSummary: "Accept the offer screen (screen 1); was accept_duel_offer (§10)."},
	{Path: []string{"duel", "confirm"}, Kind: "callable()->Result", DocSummary: "Accept the confirm screen (screen 2); was accept_duel_confirm (§10)."},
	{Path: []string{"duel", "decline"}, Kind: "callable()->Result", DocSummary: "Decline/cancel the duel (§10)."},

	// ----- magic (§10: one polymorphic cast; spell catalog promoted) -----
	{Path: []string{"magic", "cast"}, Kind: "callable(spell,target?)->Result", DocSummary: "Cast a spell, optionally on a target; unifies cast_on_self/npc/player/land/item (§10)."},
	{Path: []string{"magic", "book"}, Kind: "list<SpellDef>", DocSummary: "Spell catalog (promoted from self.spells, §10)."},
	{Path: []string{"magic", "known"}, Kind: "list<SpellDef>", DocSummary: "Spells you have the magic level for (§10)."},
	// ----- magic reads (#117) -----
	// Client-computed magic reads keyed off the current/boosted Magic
	// level (skill index 6). magic.known already exists above and now
	// resolves against the effective level at the magic root.
	{Path: []string{"magic", "level"}, Kind: "Int", DocSummary: "Current (boostable) Magic level; skill index 6 (api.md §621)."},
	{Path: []string{"magic", "max_level"}, Kind: "Int", DocSummary: "Base (unboostable) Magic level; skill index 6 (api.md §623)."},
	{Path: []string{"magic", "can_cast"}, Kind: "callable(spell)->Bool", DocSummary: "True iff req_level <= current Magic level; level-only gate, no rune check (api.md §625)."},
	{Path: []string{"magic", "has_runes_for"}, Kind: "callable(spell)->Bool", DocSummary: "True iff inventory holds the required runes; an equipped elemental staff satisfies that element (api.md §632, #117)."},

	// ----- prayer (§10: activate/deactivate; catalog promoted) -----
	// ----- prayer active (#117) -----
	// api.md §8 freezes prayer.active(slot) -> Bool as the per-slot
	// check; the active *list* moves to prayer.active_list. The active
	// set arrives on opcode 206 (event.PrayersActive) and is mirrored on
	// world.Self.activePrayers.
	{Path: []string{"prayer", "active"}, Kind: "callable(slot|name)->bool", DocSummary: "Whether a prayer slot (Int) or named prayer (String) is currently on (#117, api.md §8)."},
	{Path: []string{"prayer", "active_list"}, Kind: "list<int>", DocSummary: "Currently-active prayer slot indices (from opcode 206; promoted from self.prayers.active, #117)."},
	{Path: []string{"prayer", "book"}, Kind: "list<PrayerDef>", DocSummary: "Prayer catalog (§10)."},
	{Path: []string{"prayer", "activate"}, Kind: "callable(prayer)->Result", DocSummary: "Turn on a prayer; was activate_prayer (§10)."},
	{Path: []string{"prayer", "deactivate"}, Kind: "callable(prayer)->Result", DocSummary: "Turn off a prayer; was deactivate_prayer (§10)."},

	// ----- combat (§10: attack alias kept; set_style namespaced) -----
	{Path: []string{"combat", "attack"}, Kind: "callable(target)->Result", DocSummary: "Initiate combat with an NPC or player (§9 alias: attack)."},
	{Path: []string{"combat", "set_style"}, Kind: "callable(style)->Result", DocSummary: "Change melee xp-split mode (controlled/aggressive/accurate/defensive or 0-3); was set_combat_style (§10)."},

	// ----- combat (#117) -----
	// De-stubbed perception + new read-side style + retreat verb. The
	// pre-existing combat.engaged / combat.target rows above now resolve
	// against world.Players index 0 (SelfPlayerIndex) engagement +
	// the live NPC/player roster instead of returning false/null.
	// combat.target's health is read on the returned entity view via
	// .hp_fraction / .health (entity-view fields, documented on the
	// views, not as accessor rows — same convention as npc.hp_fraction).
	{Path: []string{"combat", "style"}, Kind: "string", DocSummary: "Current melee xp-split mode (controlled/aggressive/accurate/defensive). Read-side mirror of combat.set_style; write-through (RSC sends no echo), defaults to \"controlled\" (#117)."},
	{Path: []string{"combat", "retreat"}, Kind: "callable()->Result", DocSummary: "Break melee by walking one tile away — the only disengage mechanic in v235 (fleeing is movement; server breaks combat on the first walk packet). Steps away from the engaged target when known (#117)."},

	// ===== host — persona / identity =====
	{Path: []string{"host", "name"}, Kind: "string",
		DocSummary: "The host's RSC username."},
	{Path: []string{"host", "persona"}, Kind: "PersonaView",
		DocSummary: "Persona traits map.", NotYetImplemented: true},
	{Path: []string{"host", "mood"}, Kind: "map<string,float>",
		DocSummary: "Emotional-state weights summing to 1.0.", NotYetImplemented: true},
	{Path: []string{"host", "motivation"}, Kind: "Motivation",
		DocSummary: "{north_star, current_focus}.", NotYetImplemented: true},
	{Path: []string{"host", "defaults"}, Kind: "DefaultsView",
		DocSummary: "Persona-tier default handler refs (used by `extends host`).", NotYetImplemented: true},
}
