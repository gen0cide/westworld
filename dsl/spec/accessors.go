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

	// ===== promoted namespaced subsystems (api.md §6/§10) =====
	// These resolve through the top-level namespace views (trade /
	// bank / duel / magic / prayer). Action verbs are view-dispatched
	// callables that return Result<Null>; Views are pure reads.

	// ----- bank (§10: bank.open / deposit / withdraw / close) -----
	{Path: []string{"bank", "is_open"}, Kind: "bool", DocSummary: "Bank UI currently shown."},
	{Path: []string{"bank", "slots"}, Kind: "list<[id,amount]>", DocSummary: "Bank contents."},
	{Path: []string{"bank", "open"}, Kind: "callable(banker)->Result", DocSummary: "Open the bank UI via a banker NPC (§10)."},
	{Path: []string{"bank", "deposit"}, Kind: "callable(item,amount)->Result", DocSummary: "Deposit `amount` of the item into the open bank (§10)."},
	{Path: []string{"bank", "withdraw"}, Kind: "callable(item,amount)->Result", DocSummary: "Withdraw `amount` of the item from the open bank (§10)."},
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

	// ----- prayer (§10: activate/deactivate; catalog promoted) -----
	{Path: []string{"prayer", "active"}, Kind: "list<int>", DocSummary: "Active prayer slot indices (promoted from self.prayers, §10)."},
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
