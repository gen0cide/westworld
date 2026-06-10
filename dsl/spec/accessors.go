package spec

// AccessorSpec documents one query-layer attribute path.
//
// Unlike Actions and Events (where the spec drives behavior), the
// accessor implementations live in the runtime/views_*.go family as Getter
// switch arms. The spec here is the DOCUMENTATION + DISCOVERY +
// CONSISTENCY source: every path listed is what hosts should be
// able to read, and the consistency test asserts each path resolves
// against a stub host.
//
// Adding a query: add a row here AND extend the relevant view's
// Get() method in the runtime/views_*.go family.
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
	// the runtime/views_*.go family doesn't yet expose. The consistency
	// test uses this to distinguish "spec drift" from "planned
	// but not built."
	NotYetImplemented bool
}

// Accessors is the canonical query-layer surface. Use the helper
// constructors below to keep entries terse.
//
// Currently this list is partial — it covers what the runtime/views_*.go family
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
	{Path: []string{"inventory", "used"}, Kind: "int",
		DocSummary: "Number of occupied inventory slots (capacity - free)."},
	{Path: []string{"inventory", "capacity"}, Kind: "int",
		DocSummary: "Total inventory slots (30 — the RSC constant)."},
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

	// ----- world.scenery — DYNAMIC scenery (DSL-2) -----
	// Runtime-spawned objects streamed via SEND_SCENERY_HANDLER (opcode
	// 48) — the ONLY place a fire you just lit / a regrown tree appears
	// (world.locs sees only the static map). The selectors are
	// reachable-only by default (DSL-1, views_reachable.go).
	{Path: []string{"world", "scenery"}, Kind: "SceneryView",
		DocSummary: "Dynamic (server-streamed) scenery root: iterate it, or use .nearest / .by_id / .all / .length. A fire lit by firemaking shows up HERE, never in world.locs."},
	{Path: []string{"world", "scenery", "nearest"}, Kind: "Placement?",
		DocSummary: "Closest REACHABLE visible dynamic scenery to self, or Null. Bare field (no called form — use .all.nearest(pos, reachable=false) to recenter over everything: .nearest called on ANY list is still reachable-gated by default, so iterate .all to see unreachable records). Result is a placement: pass straight to use(item, placement)."},
	{Path: []string{"world", "scenery", "by_id"}, Kind: "callable(def_id, radius?, reachable?)->Placement?",
		DocSummary: "Nearest REACHABLE visible dynamic scenery with the given def id (97 = fire), optionally within radius= tiles; reachable=false includes unreachable (debugging). Null when none qualify."},
	{Path: []string{"world", "scenery", "all"}, Kind: "list<Placement>",
		DocSummary: "Every visible dynamic scenery record as a real list — UNFILTERED (the omniscient escape for the gated selectors)."},
	{Path: []string{"world", "scenery", "length"}, Kind: "int",
		DocSummary: "Count of visible dynamic scenery records (unfiltered)."},

	// ----- world.boundaries — doors / gates / wall-edges (DSL-2) -----
	{Path: []string{"world", "boundaries"}, Kind: "BoundariesView",
		DocSummary: "Boundary (door/gate/wall-edge) queries: .near / .at / .is_open / .dynamic. Results are boundary views for use() / open_boundary()."},
	{Path: []string{"world", "boundaries", "near"}, Kind: "callable(radius?, reachable?)->list<Boundary>",
		DocSummary: "Nearest-first list of OPENABLE boundaries (doors/gates — plain walls excluded) within radius (default 8). Reachable-only by default — a door in a sealed-off space is dropped; reachable=false includes it (debugging). Each entry carries x/y/direction for open_boundary()."},
	{Path: []string{"world", "boundaries", "at"}, Kind: "callable(x, y, dir?)->Boundary",
		DocSummary: "Boundary view at an exact tile + direction (dir defaults 0). Always returns a view (synthesized if facts has no def there) so use(key, door) can fire regardless."},
	{Path: []string{"world", "boundaries", "is_open"}, Kind: "callable(x, y, dir?)->bool",
		DocSummary: "True iff the server marked this tile/dir's boundary removed (door opened, web cut). False for unknown tiles."},
	{Path: []string{"world", "boundaries", "dynamic"}, Kind: "list<[x,y,dir,id]>",
		DocSummary: "Snapshot of all live boundary overrides as [x, y, dir, id] rows. Debugging/persistence."},

	// ----- world.dialog — NPC menu state (DSL-2) -----
	{Path: []string{"world", "dialog"}, Kind: "DialogView",
		DocSummary: "NPC dialog menu state: .is_open / .options / .find_option / .clear. Read the option TEXT and pick by content — option order is not stable across encounters."},
	{Path: []string{"world", "dialog", "is_open"}, Kind: "bool",
		DocSummary: "True while an NPC dialog menu is presented."},
	{Path: []string{"world", "dialog", "options"}, Kind: "list<string>",
		DocSummary: "The presented option texts (empty list when no menu is open)."},
	{Path: []string{"world", "dialog", "find_option"}, Kind: "callable(substr)->int",
		DocSummary: "1-based index of the first option containing the substring (case-insensitive), or 0 — pairs directly with answer(). Same contract as the top-level find_option builtin."},
	{Path: []string{"world", "dialog", "clear"}, Kind: "callable()",
		DocSummary: "Invalidate the cached menu after resolving it (the server doesn't reliably signal menu close)."},

	// ----- planned by docs/lang/state.md Build Plan, not yet
	// wired in the runtime/views_*.go family (tasks #56–#65) -----
	{Path: []string{"self", "hp_fraction"}, Kind: "float", DocSummary: "hp / max_hp."},
	{Path: []string{"self", "quest_points"}, Kind: "int", DocSummary: "QP count."},
	{Path: []string{"self", "is_busy"}, Kind: "bool", DocSummary: "Currently performing an action? (stub: always false until action tracking lands)"},

	// Per-entity-view accessors (npc/player/ground_item/...) are not
	// rooted at a fixed top-level path — they live ON the views
	// returned from world.npcs / world.players / world.ground_items.
	// Their FIELD tables are spec'd in dsl/spec/views.go (spec.Views,
	// DSL-2) and rendered into the manual by APIReference; the view
	// implementations in the runtime/views_*.go family remain the
	// canonical source.
	{Path: []string{"self", "is_in_combat"}, Kind: "bool", DocSummary: "Engaged with an NPC or player? (stub: always false until combat tracking lands)"},
	{Path: []string{"self", "is_sleeping"}, Kind: "bool", DocSummary: "Sleep screen (fatigue captcha) currently up? True between SEND_SLEEPSCREEN and SEND_STOPSLEEP; the cradle auto-answers the word, so this is usually a brief flicker."},
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
	// ----- #119: world.messages is now a real bounded ring of the
	// last N server messages (oldest-first), and `on message` fires
	// per new entry. combat.target (declared above) resolves to the
	// live engaged NPC and clears on death — see views_combat.go. -----
	{Path: []string{"world", "messages"}, Kind: "list<Message>", DocSummary: "Bounded server-message log, oldest-first (§10: was world.last_server_message). Each Message has .text / .kind / .at / .contains(needle). Backed by a real N-deep ring (#119); `on message(text)` fires per new entry."},

	// ===== Def/Instance instance fields (api.md §2) — InvSlot =====
	{Path: []string{"inventory", "find_all"}, Kind: "callable(item)->list<InvSlot>", DocSummary: "Every slot matching the item as an InvSlot instance (one per slot; non-stackables occupy several)."},

	// ----- ground_items.nearest + find_any (#117) -----
	// Convenience reads added by #117. All pure Views (no Action verbs):
	// world.ground_items.nearest is dual-mode (bare field => nearest to
	// self; called with a position => nearest to that position),
	// world.ground_items.most_valuable is a value-sorted selector, and
	// inventory.find_any collapses or-chains over a set of item refs.
	// The selectors are reachable-only by default (DSL-1) — see the
	// per-row notes for the reachable=false opt-out.
	{Path: []string{"world", "ground_items", "nearest"}, Kind: "callable(pos?, reachable?)->GroundItem?",
		DocSummary: "Nearest REACHABLE visible ground item to self (bare field), or to an explicit position when called: world.ground_items.nearest(pos). Items you cannot walk to (behind a wall) are skipped by default; .nearest(reachable=false) includes them (debugging). Null when none qualify."},
	{Path: []string{"world", "ground_items", "most_valuable"}, Kind: "GroundItem?",
		DocSummary: "Highest-base-value REACHABLE visible ground item (by facts.ItemDef base value), or Null when none qualify. Enables loot-most-valuable. Bare field — iterate the raw .all list to see unreachable items (.nearest on that list is still reachable-gated by default; pass reachable=false)."},
	{Path: []string{"world", "ground_items", "by_id"}, Kind: "callable(item_id, radius?, reachable?)->GroundItem?",
		DocSummary: "Nearest REACHABLE visible ground item with the given item id, or Null. Optional radius= cap (tiles); reachable=false includes unreachable items (debugging)."},
	{Path: []string{"world", "ground_items", "all"}, Kind: "list<GroundItemView>",
		DocSummary: "Every visible ground item as a real list — UNFILTERED (includes unreachable items; the omniscient escape for the gated selectors). Iterate or .filter/.map here; .nearest on it still applies the reachable gate by default (add reachable=false to keep everything)."},
	{Path: []string{"world", "ground_items", "length"}, Kind: "int",
		DocSummary: "Count of visible ground items (unfiltered)."},
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
	{Path: []string{"bank", "has"}, Kind: "callable(item_id)->Int", DocSummary: "Total banked amount of the item id (0 if absent / bank closed). Int ids ONLY — names are NOT resolved here yet (unlike inventory.has; DSL-22 tracks the fix)."},
	{Path: []string{"bank", "count"}, Kind: "callable(item_id)->Int", DocSummary: "Alias of bank.has(item_id). Int ids ONLY."},
	{Path: []string{"bank", "deposit_all"}, Kind: "callable(keep?)->Result", DocSummary: "Deposit every inventory item except those in the optional keep-list (item ids/names) (#117/#118)."},
	{Path: []string{"bank", "withdraw_all"}, Kind: "callable(item)->Result", DocSummary: "Withdraw the entire banked quantity of one item (#117/#118)."},
	{Path: []string{"bank", "withdraw_x"}, Kind: "callable(item,amount)->Result", DocSummary: "Withdraw a preset amount of one item, clamped to the banked quantity (#117/#118)."},
	{Path: []string{"bank", "close"}, Kind: "callable()->Result", DocSummary: "Close the bank UI (§10)."},
	// ----- shop (§8: shop.is_open / stock / price / buy / sell / close) -----
	{Path: []string{"shop", "is_open"}, Kind: "bool", DocSummary: "Shop UI currently shown."},
	{Path: []string{"shop", "is_general"}, Kind: "bool", DocSummary: "True iff the open shop is a general store (buys anything); false when closed or specialist."},
	{Path: []string{"shop", "slots"}, Kind: "list<ShopSlot>", DocSummary: "Shop catalogue: item_id / stock / base_stock."},
	{Path: []string{"shop", "stock"}, Kind: "callable(item)->Int", DocSummary: "Units of the item the shop currently stocks."},
	{Path: []string{"shop", "price"}, Kind: "callable(item)->Int", DocSummary: "Current buy price (gp) of the item at this shop."},
	{Path: []string{"shop", "buy"}, Kind: "callable(item,qty)->Result", DocSummary: "Buy qty of the item from the open shop."},
	{Path: []string{"shop", "sell"}, Kind: "callable(item,qty)->Result", DocSummary: "Sell qty of the item to the open shop."},
	{Path: []string{"shop", "close"}, Kind: "callable()->Result", DocSummary: "Close the shop window."},

	// ----- trade (§10: request/offer/accept/confirm/decline) -----
	// The they_/both_ handshake family (§7 frozen names) mirrors the
	// two-screen RSC trade: offer screen (accepted) then confirm
	// screen (confirmed). DSL-2: these existed in tradeView long
	// before they had spec rows.
	{Path: []string{"trade", "is_active"}, Kind: "bool", DocSummary: "Trade UI currently shown."},
	{Path: []string{"trade", "phase"}, Kind: "string", DocSummary: "Trade phase (none/offer/confirm/…)."},
	{Path: []string{"trade", "with"}, Kind: "string?", DocSummary: "The other party's name, or Null when no trade."},
	{Path: []string{"trade", "with_index"}, Kind: "int", DocSummary: "The other party's server player index (0 when no trade)."},
	{Path: []string{"trade", "my_offer"}, Kind: "list<[id,amount]>", DocSummary: "Items you have put up."},
	{Path: []string{"trade", "their_offer"}, Kind: "list<[id,amount]>", DocSummary: "The other party's items."},
	{Path: []string{"trade", "accepted"}, Kind: "bool", DocSummary: "You clicked Accept on the offer screen."},
	{Path: []string{"trade", "they_accepted"}, Kind: "bool", DocSummary: "The other party clicked Accept on the offer screen."},
	{Path: []string{"trade", "both_accepted"}, Kind: "bool", DocSummary: "Both sides accepted the offer screen — the confirm screen is next."},
	{Path: []string{"trade", "confirmed"}, Kind: "bool", DocSummary: "You clicked Accept on the confirm screen."},
	{Path: []string{"trade", "they_confirmed"}, Kind: "bool", DocSummary: "The other party clicked Accept on the confirm screen."},
	{Path: []string{"trade", "both_confirmed"}, Kind: "bool", DocSummary: "Both sides confirmed — the exchange is executing."},
	{Path: []string{"trade", "request"}, Kind: "callable(player)->Result", DocSummary: "Trade-request a player; absorbs the old trade_request+accept_trade (§10)."},
	{Path: []string{"trade", "offer"}, Kind: "callable(items)->Result", DocSummary: "Set/replace your offer ([id,amount] pairs) (§10)."},
	{Path: []string{"trade", "accept"}, Kind: "callable()->Result", DocSummary: "Accept the offer screen (screen 1); was confirm_trade (§10)."},
	{Path: []string{"trade", "confirm"}, Kind: "callable()->Result", DocSummary: "Accept the confirm screen (screen 2); was finalize_trade (§10)."},
	{Path: []string{"trade", "decline"}, Kind: "callable()->Result", DocSummary: "Decline/close the trade (§10)."},

	// ----- duel (§10: request/set_rules/stake/accept/confirm/decline) -----
	// Same two-screen handshake shape as trade (DSL-2), plus the four
	// rule toggles.
	{Path: []string{"duel", "is_active"}, Kind: "bool", DocSummary: "Duel UI currently shown."},
	{Path: []string{"duel", "phase"}, Kind: "string", DocSummary: "Duel phase."},
	{Path: []string{"duel", "with"}, Kind: "string?", DocSummary: "The opponent's name, or Null when no duel."},
	{Path: []string{"duel", "with_index"}, Kind: "int", DocSummary: "The opponent's server player index (0 when no duel)."},
	{Path: []string{"duel", "my_offer"}, Kind: "list<[id,amount]>", DocSummary: "Items you have staked."},
	{Path: []string{"duel", "their_offer"}, Kind: "list<[id,amount]>", DocSummary: "The opponent's staked items."},
	{Path: []string{"duel", "accepted"}, Kind: "bool", DocSummary: "You clicked Accept on the stake screen."},
	{Path: []string{"duel", "they_accepted"}, Kind: "bool", DocSummary: "The opponent clicked Accept on the stake screen."},
	{Path: []string{"duel", "both_accepted"}, Kind: "bool", DocSummary: "Both sides accepted the stake screen — the confirm screen is next."},
	{Path: []string{"duel", "confirmed"}, Kind: "bool", DocSummary: "You clicked Accept on the confirm screen."},
	{Path: []string{"duel", "they_confirmed"}, Kind: "bool", DocSummary: "The opponent clicked Accept on the confirm screen."},
	{Path: []string{"duel", "both_confirmed"}, Kind: "bool", DocSummary: "Both sides confirmed — the duel is starting."},
	{Path: []string{"duel", "disallow_retreat"}, Kind: "bool", DocSummary: "Rule toggle: no retreating in this duel."},
	{Path: []string{"duel", "disallow_magic"}, Kind: "bool", DocSummary: "Rule toggle: no magic in this duel."},
	{Path: []string{"duel", "disallow_prayer"}, Kind: "bool", DocSummary: "Rule toggle: no prayer in this duel."},
	{Path: []string{"duel", "disallow_weapons"}, Kind: "bool", DocSummary: "Rule toggle: no weapons in this duel."},
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
	{Path: []string{"combat", "last_npc"}, Kind: "NpcView?", DocSummary: "The NPC you most recently attacked, resolved live from the roster — Null if never attacked OR it has since died/left view (your retarget signal)."},
	{Path: []string{"combat", "last_player"}, Kind: "PlayerView?", DocSummary: "The player you most recently attacked (duels), same live-resolution + Null semantics as last_npc."},
	{Path: []string{"combat", "retreat"}, Kind: "callable(wait_rounds?)->Result", DocSummary: "Break melee by walking one tile away — the only disengage mechanic in v235 (fleeing is a WALK_TO_POINT; server breaks combat on it). RSC anti-kite: you cannot retreat until the opponent has made 3 hits (\"first 3 rounds of combat\"). By default (wait_rounds=true) the verb waits out the 3 rounds when it can detect them, then walks; if the server still rejects, it returns a typed RETREAT_TOO_EARLY result carrying the server message so a routine can wait + retry. Pass wait_rounds=false to attempt immediately and get the rejection back for poll-and-branch (#117, #r3-retreat)."},
	{Path: []string{"combat", "retreat_to"}, Kind: "callable(x, y, wait_rounds?)->Result", DocSummary: "Flee to a specific safe tile (effectively a combat-aware walk_to). Shares retreat's 3-round anti-kite gate + RETREAT_TOO_EARLY rejection: waits out the rounds (wait_rounds=false to skip), sends the breaking WALK_TO_POINT toward (x, y), then pathfinds the rest of the way. Use it to break off toward a bank/altar rather than just one tile back (#r3-retreat)."},

	// ===== host — persona / identity =====
	{Path: []string{"host", "name"}, Kind: "string",
		DocSummary: "The host's RSC username. (Use self.name, which is live.)", NotYetImplemented: true},
	{Path: []string{"host", "persona"}, Kind: "PersonaView",
		DocSummary: "Persona traits map.", NotYetImplemented: true},
	{Path: []string{"host", "mood"}, Kind: "map<string,float>",
		DocSummary: "Emotional-state weights summing to 1.0.", NotYetImplemented: true},
	{Path: []string{"host", "motivation"}, Kind: "Motivation",
		DocSummary: "{north_star, current_focus}.", NotYetImplemented: true},
	{Path: []string{"host", "defaults"}, Kind: "DefaultsView",
		DocSummary: "Persona-tier default handler refs (used by `extends host`).", NotYetImplemented: true},

	// ----- self equipped/plane (#117) -----
	// self.position gained a `.plane` field (Y/PlaneHeight); fully wired.
	// self.equipped keeps its frozen list surface (.all/.length/.first/
	// .last/[N]/for) AND gains per-slot accessors reading
	// world.Self.EquipBySlot. The per-slot accessors are structurally
	// present but their data source is unwired: self's appearance update
	// is keyed by a player index the runtime cannot yet identify, so the
	// sprite ids read 0; and facts.ItemDef has no AppearanceId, so the
	// sprite-id -> item-id reverse lookup is impractical. .sprite_id is
	// the honest datum; .id/.name/.def stay null until both gaps close.
	{Path: []string{"self", "position", "plane"}, Kind: "int",
		DocSummary: "Floor/plane index of our tile (Y/944; 0 ground, 1+ upper, 3 underground). From world.Self.Plane()."},
	{Path: []string{"self", "equipped", "all"}, Kind: "list<ItemView>",
		DocSummary: "All currently-wielded items as a real list (use .filter/.map/.find here)."},
	{Path: []string{"self", "equipped", "bonuses"}, Kind: "EquipmentBonuses",
		DocSummary: "Summed combat bonuses of all worn gear, recomputed live: .armour / .aim / .power / .magic / .prayer (the classic equipment-status panel)."},
	{Path: []string{"self", "equipped", "weapon"}, Kind: "EquipSlotView",
		DocSummary: "Weapon slot: .sprite_id (appearance id worn) / .is_empty / .slot. .id/.name/.def are best-effort (null today: no sprite->item map). See blockers."},
	{Path: []string{"self", "equipped", "shield"}, Kind: "EquipSlotView",
		DocSummary: "Shield slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "head"}, Kind: "EquipSlotView",
		DocSummary: "Head slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "hat"}, Kind: "EquipSlotView",
		DocSummary: "Hat/helmet slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "body"}, Kind: "EquipSlotView",
		DocSummary: "Body (chest armour) slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "legs"}, Kind: "EquipSlotView",
		DocSummary: "Legs slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "gloves"}, Kind: "EquipSlotView",
		DocSummary: "Gloves slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "boots"}, Kind: "EquipSlotView",
		DocSummary: "Boots slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "amulet"}, Kind: "EquipSlotView",
		DocSummary: "Amulet slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "cape"}, Kind: "EquipSlotView",
		DocSummary: "Cape slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "shirt"}, Kind: "EquipSlotView",
		DocSummary: "Shirt (torso undergarment) slot worn-equipment sprite; same shape as .weapon."},
	{Path: []string{"self", "equipped", "pants"}, Kind: "EquipSlotView",
		DocSummary: "Pants slot worn-equipment sprite; same shape as .weapon."},
}
