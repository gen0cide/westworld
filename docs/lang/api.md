# The Routine API — frozen surface (v1, ratified 2026-05-29)

> Status: **FROZEN (v1)** — ratified 2026-05-29. The design contract for the
> host-facing DSL; cognition / brain / persona build on it. Amend only by
> deliberate decision (and record it), never by drift. The body build-out
> (tasks #115–120) brings the implementation up to this surface.
>
> **Two parts, like Godoc:** this hand-written doc is the *contract* — the
> model, the type vocabulary, the namespacing rules, the capability boundary,
> and one fully-worked reference section (`trade`) that fixes the depth bar.
> The *exhaustive* per-entry reference (every field/verb/spell with its types,
> args, returns, errors) is **generated from `dsl/spec/*`** so it can never
> drift from the code — the spec structs gain `Args[]{name,type}`, `Returns`,
> `Errors[]`, `Faculty`, and `GuiEquivalent` fields, and a `go run
> ./cmd/specdoc` emits the reference in this exact format. (Hand-maintaining
> ~200 spells + ~90 skill accessors by hand would drift on the first change.)

---

## 1. Faculties — and the View vs Action line

Every host-facing primitive is exactly one of three kinds.

- **View (Perception).** A pure read of the local world-mirror. No packet, no
  side effect, synchronous, returns a value (or `Null`). Plain fields
  (`self.hp`) and parameterized getters (`inventory.find(373)`) are both Views.
- **Action.** Sends a packet — *doing*. Out-of-world side effect, resolves
  **asynchronously** (consequences land later as mirror updates + events),
  returns a **`Result`** reporting whether the *request* was accepted.
- **Reaction (Event).** An `on <event>(...)` / `select` hook that fires when the
  mirror changes.

**A View is not a getter-Action.** The line is *side-effect + async + Result*.
`trade.their_offer` (View — read the screen) vs `trade.accept()` (Action — click
Accept, returns `Result`, can fail). They share the `trade` root because they're
one subsystem; they differ in faculty. **If it can fail or change the world →
Action (returns `Result`); if it only reports current state → View.**

## 2. Value types

Scalars: `Int`, `Float`, `Bool`, `String`, `Null`. Composite: `List<T>`.

`Result<T>` — every Action returns this. Either `Ok<T>` (here `T` is almost
always `Null`) or `Err{ code: ErrorCode, reason: String }`. `ErrorCode` is a
fixed enum (`dsl/interp/error.go`): `PATH_BLOCKED`, `NO_SUCH_ITEM`,
`NOT_LOGGED_IN`, `BANK_NOT_OPEN`, `TRADE_NOT_ACTIVE`, `DIALOG_NOT_OPEN`,
`NOT_IMPLEMENTED`, … Routines branch on `res.err`/`res.err.code` or use the
bang form to abort.

**Definition vs. Instance.** A *Definition* is the static catalog entry — one
per game id, immutable, from the **facts registry** (loaded from the OpenRSC
defs). An *Instance* is a live occurrence in the world *right now*, and it
**carries its `.def`** plus its live state. "A Rune 2-handed Sword" is an
`ItemDef`; "the two of them in inventory slots 3 and 7" are two `InvSlot`
instances whose `.def` is that same `ItemDef`. Perception returns instances;
the catalog/resolution returns defs.

**Defs** (static; one per id; from the facts registry):

| def | fields (all `View`) |
| --- | --- |
| `ItemDef` | `id: Int`, `name: String`, `stackable: Bool`, `wearable: Bool`, `members: Bool`, `edible: Bool`, … |
| `NpcDef` | `id: Int`, `name: String`, `combat_level: Int`, `max_hp: Int`, `attackable: Bool`, `aggressive: Bool`, `command1/2: String` |
| `LocDef` | `id: Int`, `name: String`, `command1: String`, `command2: String` |
| `SpellDef` | `id: Int`, `name: String`, `level: Int`, `runes: List<[Int,Int]>` |
| `PrayerDef` | `id: Int`, `name: String`, `level: Int`, `drain: Int` |

**Instances** (live; from the world mirror; each has `.def` + live fields):

| instance | fields (all `View`) |
| --- | --- |
| `Position` | `x: Int`, `y: Int`, `plane: Int` |
| `Skill` | `level: Int` (effective/current), `max_level: Int` (base), `xp: Int`, `xp_to_next: Int`, `percent_to_next: Float` |
| `InvSlot` | `idx: Int` (0–29), `def: ItemDef`, `quantity: Int` |
| `Npc` | `server_index: Int`, `def: NpcDef`, `position: Position`, `health: Float` (0–1; **your target only**) |
| `Player` | `index: Int`, `name: String`, `position: Position` |
| `Loc` | `def: LocDef`, `position: Position` |
| `GroundItem` | `def: ItemDef`, `position: Position`, `quantity: Int` |
| `Boundary` | `def: LocDef?`, `position: Position`, `direction: Int` |
| `Message` | `text: String`, `kind: String`, `at: Time` |

Static attributes go through `.def` (`slot.def.name`, `npc.def.combat_level`),
live state directly (`slot.quantity`, `npc.position`, `npc.health`). A `find`
returns instances: `inventory.find(ItemRef) -> InvSlot?` (first match),
`inventory.find_all(ItemRef) -> List<InvSlot>` (every slot — non-stackables
occupy several), `inventory.count(ItemRef) -> Int` (total quantity).

**Refs — how you *name* a thing to a verb/view:**
`ItemRef = Int(id) | String(name) | ItemDef | InvSlot`;
`NpcRef = Int(server_index) | Npc`; `SpellRef = Int(id) | String(name) | SpellDef`;
`PlayerRef = Int(index) | String(name) | Player`;
`Target = Npc | Player | Loc | GroundItem | Boundary | Position`.
A `String` ref resolves through the facts registry with **conservative,
deterministic** normalization (case/space/punct) + high-confidence typo
tolerance only — anything ambiguous returns `Null`. Slang ("r2h", "addy") and
ambiguity route to the cognition-plane `resolve()` (§5), never the frozen body.

## 3. The capability boundary (principle #1)

- A **View** may exist only if a human can *see* that on screen. (So
  `Npc.health` is the 0–1 health bar of your current target — never raw hidden
  hitpoints; an off-screen NPC's stats are not Views.)
- An **Action** may exist only if a human can do it through the client.
- **Reactions** fire only on things a player would notice.

This boundary is what makes the freeze possible: the 2001 client is fixed, so
new game content flows through *data*, never new API.

## 4. Body vs Mind

The API is *faculties*, not *behavior*. Auto-eat thresholds, retreat rules,
kill/time budgets, prayer-flicking, autocasting are **decisions** — a player has
no button for them. They are **mind** (cognition/persona), composed above this
API from `when`/`select`/loops. They are not verbs here. "Fat" = complete
faculties; **lean in policy.**

## 5. Control plane — the non-GUI primitives (outside §3 by design)

GUI-equivalence governs the **body** (game I/O). A routine also has its own
faculties — flow, logging, thinking, test-setup — that no GUI player has. These
are a deliberately **separate, fenced layer**, never confused with body verbs:

- **Flow / timing (language constructs):** `if/elif/else`, `while`, `for`,
  `return <v>`, `abort <msg>`, `wait(sec: Float)`, `wait_until(pred: fn->Bool,
  timeoutSec: Float) -> Bool`, `repeat { … } until <cond> timeout <Float sec>`,
  `when <cond> { … }`, `select { on <event> {…} timeout <Ns> {…} }`,
  `defer { … }`, `try { … } recover err { … }`.
- **Introspection:** `note(msg: String) -> Null` — emit to the host log
  (observability only; not seen in-world).
- **Recognition / fuzzy resolution (mind-access, *learnable* — NOT GUI-equivalent):**
  `resolve(text: String, kind?: String) -> List<Match{ def, kind, score }>` —
  ranked candidates, best first. Pipeline: the **host's learned-alias store**
  (persisted; the host *grows its own lingo* as it plays) → conservative
  fuzzy/token match vs canonical names → brain (LLM) fallback, which on success
  **writes the resolved alias back** (so `"r2h" → "Rune 2-handed Sword"` is a
  cheap table hit forever after). This is the host's *recognition* faculty — the
  analogue of a player instantly knowing what a "r2h" is. Definitions are looked
  up exactly (facts); only the *name→canonical* recognition is fuzzy, and ids
  are never invented by the LLM (they come from the resolved `def` or from a
  perceived instance). We ship **no** curated slang table — each host learns and
  seeds its own.
- **Cognition bridge (mind-access — NOT GUI-equivalent, fenced):**
  `recall(query: String, top?: Int) -> List<Chunk>` (knowledge corpus),
  `ask_brain(prompt: String) -> String`, `contemplate_reality(…)`. These are how
  the routine *thinks*; they hit the knowledge/brain layer, not the game.
- **Admin / test (fenced, test-harness only):** `command(cmd: String) ->
  Result<Null>` — `setstat`, `item`, `teleport`, etc. Never appears in
  persona-authored behavior.

Answer to "can we add those in?": **yes** — but in this fenced control-plane
section, explicitly outside the GUI-equivalence rule, so the body surface stays
honest and freezable.

## 6. Naming & namespacing rules

1. **Namespace by subsystem; Views and Actions share the root**
   (`trade.accept()` + `trade.their_offer`). Roots: `self`, `world`,
   `inventory`, `combat`, `magic`, `prayer`, `trade`, `bank`, `duel`, `shop`.
2. **Top-level for the ambient body** — verbs owned by no subsystem (movement,
   generic item interaction, chat, NPC clicks): they read like player-speech.
3. **Parameterize, never enumerate** — one `magic.cast(spell, target)`, one
   `on xp_gain(skill)`.
4. **One obvious way** — each capability once; the alias list (§9) is the only
   permitted second-names.

## 7. Reference-entry format (the depth bar) + the `trade` exemplar

Every entry documents: **signature** (args with types) → **return type**;
**faculty**; **semantics**; **errors** (Actions); **nullability** (Views);
**GUI equivalent**. The generated reference (§Status) emits exactly this.

### `trade` — Views

- **`trade.is_active`** → `Bool`. Faculty: View. True while a trade window is
  open. GUI: the trade window is on screen.
- **`trade.phase`** → `String` ∈ {`"none"`, `"offer"`, `"confirm"`}. Which
  screen is showing.
- **`trade.my_offer`** → `List<Item>`. Items you have put up (empty list, never
  null). GUI: your side of the offer.
- **`trade.their_offer`** → `List<Item>`. The other party's items.
- **`trade.accepted`** → `Bool`. You clicked Accept on the **offer** screen.
- **`trade.they_accepted`** → `Bool`. They did.
- **`trade.both_accepted`** → `Bool`. Both did → the confirm screen opens.
- **`trade.confirmed`** / **`trade.they_confirmed`** / **`trade.both_confirmed`**
  → `Bool`. Same three, for the **confirm** screen.

### `trade` — Actions (all → `Result<Null>`)

- **`trade.request(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Walks
  adjacent, then sends a trade request to `p` (wire opcode 142). In RSC this one
  action both *initiates* a trade and *accepts an incoming* one — clicking the
  player serves both, and mutual requests open the window (so there is no
  separate "accept request" verb). Errors: `NO_SUCH_ITEM`-class if `p` is not a
  visible player. GUI: right-click player → "Trade with".
- **`trade.offer(items: List<[Int, Int]>)`** → `Result<Null>`. Sets/replaces
  your offer; each element is `[item_id, amount]`. Per the server rule this
  **resets both parties' offer-accepts**. Errors: `TRADE_NOT_ACTIVE`;
  `NO_SUCH_ITEM` if you don't hold an offered item. GUI: drag items into the
  window.
- **`trade.accept()`** → `Result<Null>`. Clicks "Accept" on the **offer** screen
  (screen 1). Idempotent; re-fires automatically if an offer change reset your
  accept. Errors: `TRADE_NOT_ACTIVE`. GUI: the first Accept button.
- **`trade.confirm()`** → `Result<Null>`. Clicks "Accept" on the **confirm**
  screen (screen 2), completing your half. The confirm screen only opens after
  `trade.both_accepted`. Errors: `TRADE_NOT_ACTIVE`; `NOT_IMPLEMENTED`/not-ready
  if called before the offer screen is accepted. GUI: the second Accept button.
- **`trade.decline()`** → `Result<Null>`. Declines/closes. Errors:
  `TRADE_NOT_ACTIVE`. GUI: close the window.

### `trade` — Events

- **`on trade_opened(other: Player)`** — window opened with `other`.
- **`on trade_other_offer(items: List<Item>)`** — they changed their offer.
- **`on trade_other_accepted()`** — they accepted the offer screen.
- **`on trade_closed(completed: Bool)`** — closed; `completed` iff goods moved.

> This is the depth every namespace gets (generated). `bank`, `duel`, `shop`,
> `combat`, `magic`, `prayer`, `self`, `world`, `inventory` follow the same form;
> their summaries are in §8 pending generation.

## 8. The namespaces (full reference)

### `self` — Views

#### `self.position()` → `Position`
Faculty: View. Current tile coordinates. Returns `{x: Int, y: Int}` — no plane field (RSC single-plane; plane tracking is stage 2). Never null. GUI: player sprite location on minimap and main view. (exists)

#### `self.name` → `String`
Faculty: View. The host's RSC username (from host opts.Username). Never null. Also aliased as `self.username`. GUI: name above player sprite in-game. (exists)

#### `self.hp` → `Int`
Faculty: View. Current hitpoints (boostable via potions). Range 0 to max_hp. GUI: health bar. (exists)

#### `self.max_hp` → `Int`
Faculty: View. Base hitpoints (current hits skill level). GUI: the max label on health bar. (exists)

#### `self.hp_fraction` → `Float`
Faculty: View. Ratio hp / max_hp. Range 0.0–1.0 (clamped). Never null. Returns 0 if max_hp ≤ 0. (exists)

#### `self.prayer` → `Int`
Faculty: View. Current prayer points. Shorthand for prayer skill's current level. GUI: prayer bar. (exists)

#### `self.max_prayer` → `Int`
Faculty: View. Base prayer points (prayer skill max level). GUI: prayer bar max label. (exists)

#### `self.fatigue` → `Int`
Faculty: View. Fatigue meter (0–750 scaled; RSC sleep mechanic). At 100 the sleep screen appears and the host must rest. GUI: fatigue bar if visible (varies by client version). (exists)

#### `self.combat_level` → `Int`
Faculty: View. Derived combat level (computed from attack/defense/magic/ranged/prayer). Never null. (exists)

#### `self.quest_points` → `Int`
Faculty: View. Total quest points earned. Range 0+. Never null. (exists)

#### `self.death_count` → `Int`
Faculty: View. Number of times died in this session. Incremented by Apply(event.Death). Range 0+. Never null. (exists)

#### `self.last_death_at` → `Position?`
Faculty: View. Tile where the host last died. Captured at Apply(event.Death) before respawn packet. Returns `{x: Int, y: Int}` or null if never died. Routines read this after `on death` fires to walk back to death spot. Nullability: null until first death. (exists)

#### `self.is_busy` → `Bool`
Faculty: View. True while performing an action (moving, attacking, eating, etc.). **Stub; always returns false.** Requires per-host action registry in stage 2. GUI: would reflect if the player is mid-action. (to build — perception gap)

#### `self.is_in_combat` → `Bool`
Faculty: View. True while engaged with an NPC or player. **Stub; always returns false.** Requires combat-target observation. GUI: would reflect active combat engagement. (to build — perception gap)

#### `self.is_sleeping` → `Bool`
Faculty: View. True if the sleep screen is showing (fatigue >= 100). **Stub; always returns false.** Requires SleepScreen state field. GUI: the sleep dialogue. (to build — perception gap)

#### `self.skills` → `SkillsView`
Faculty: View. Access per-skill data. Supports 18 skill names (attack, defense, strength, hits, hitpoints, ranged, prayer, magic, cooking, woodcutting, fletching, fishing, firemaking, crafting, smithing, mining, herblaw, agility, thieving). Each skill resolves to a `Skill` instance. GUI: the skills menu. (exists)

##### `self.skills.<name>` → `Skill`
One skill's live state, where `<name>` is one of the 18 skill names. Never null; invalid names return nil (evaluates to falsy). Returns a Skill object with fields:
- `.level` → `Int` (current/boostable level, 0–127 or higher if boosted)
- `.max_level` → `Int` (base level, 1–99)
- `.xp` → `Int` (total experience, 0–13034431 at level 99)
- `.xp_to_next_level` → `Int` (experience gap to reach next level; 0 if already at 99)
- `.percent_to_next_level` → `Float` (progress 0.0–1.0 toward the next level)
- `.name` → `String` (the skill name as requested, e.g., "attack")
- `.id` → `Int` (index 0–17 in the RSC catalog)

GUI: the detailed skill display in the skills menu. (exists)

#### `self.wielded` → `InvSlot?`
Faculty: View. Currently-wielded weapon item, or null if unarmed. Returns the first `InvSlot` marked wielded. Nullability: null if no item is equipped. GUI: the weapon visible on the player sprite. (exists)

#### `self.equipped` → `List<InvSlot>`
Faculty: View. All currently-wielded items as a list (empty list if nothing equipped). Per-slot accessors (.weapon, .shield, .head, etc.) require decoding the equipment-by-slot packet (not yet wired) — until then routines filter the list themselves. Never null. Returns `List<InvSlot>` where each slot carries `.id`, `.amount`, and `.name` from facts. GUI: the equipped panel (character equipment screen). (exists)

#### `self.prayers` → `PrayersView`
Faculty: View. Access active prayers and the static prayer catalog. Supports:
- `.active` → `List<Int>` (list of active prayer slot indices, empty if none)
- `.count` → `Int` (number of active prayers)
- `.is_active(slot_index)` → `Bool` (whether one prayer slot is on)
- `.book` → `List<PrayerDef>` (the static prayer catalog, 0–14 slots, immutable)
- `.by_id(id)` → `PrayerDef?` (look up a prayer by slot index, null if invalid)
- `.by_name(name)` → `PrayerDef?` (look up a prayer by name, case-insensitive, null if not found)

Each `PrayerDef` carries `.id` / `.name` / `.req_level` / `.drain_rate` / `.description` (all static from facts). GUI: the prayer book interface. (exists)

#### `self.spells` → `SpellsView`
Faculty: View. Access known spells and the static spellbook catalog. Supports:
- `.book` → `List<SpellDef>` (complete spell catalog from facts)
- `.known` → `List<SpellDef>` (spells whose req_level we have the magic level for, live-computed)
- `.count` → `Int` (total spell catalog size)
- `.by_id(id)` → `SpellDef?` (look up a spell by id, null if invalid)
- `.by_name(name)` → `SpellDef?` (look up a spell by name, case-insensitive, null if not found)
- `.has_runes_for(spell_id_or_name)` → `Bool` (true iff inventory holds all required runes)

Each `SpellDef` carries `.id` / `.name` / `.req_level` / `.type` / `.exp` (experience reward) / `.description` / `.members` / `.evil` / `.runes` (list of [item_id, count] pairs, immutable from facts). GUI: the spellbook interface. (exists)

---

Notes on §2 types:
- `Position` = `{x: Int, y: Int}` (plane not yet implemented; RSC is single-plane. Stage 2 work will add plane tracking.)
- `Skill` = `{level, max_level, xp, xp_to_next_level, percent_to_next_level, name, id}` (all Int except percent_to_next_level Float)
- `InvSlot` = instance with fields `.id` (item id), `.amount` (quantity), `.name` (from facts), `.is_wielded` (bool)
- `PrayerDef` = static catalog entry from facts `{id, name, req_level, drain_rate, description}`
- `SpellDef` = static catalog entry from facts `{id, name, req_level, type, exp, description, members, evil, runes}`

### `inventory` — Views

The `inventory` namespace provides Views into the host's item container state. All accessors are pure reads reflecting the live world mirror (updated by the inbound packet loop).

- **`inventory.free`** → `Int`. Faculty: View. Number of empty inventory slots remaining. Range: 0–30 (RSC standard 30-slot inventory). GUI: visible in the inventory UI as "Free slots: N". (exists)

- **`inventory.used`** → `Int`. Faculty: View. Number of occupied slots. Computed as 30 - `.free`. Range: 0–30. (exists)

- **`inventory.capacity`** → `Int`. Faculty: View. Total slot capacity (always 30 in RSC). (exists)

- **`inventory.is_full`** → `Bool`. Faculty: View. True iff `.free == 0`. (exists)

- **`inventory.slots`** → `List<InvSlot>`. Faculty: View. Every slot's current content as an InvSlot instance. Each InvSlot carries `.id` (item id), `.amount` (quantity), `.name` (human-readable name from facts), `.is_stackable` (bool), `.is_wearable` (bool), `.is_members_only` (bool), `.command` (default right-click verb). Order matches server slot order. Empty list if inventory is empty. GUI: the entire inventory window. (exists)

- **`inventory.has(item: ItemRef)`** → `Bool`. Faculty: View. True iff at least one slot contains the item (by id, name, or ItemDef). `ItemRef = Int(id) | String(name) | ItemDef | InvSlot`. String refs resolve via facts (deterministic exact match first, substring fallback, else returns false; ambiguous names return false). GUI: equivalent to eyeballing the inventory. (exists)

- **`inventory.count(item: ItemRef)`** → `Int`. Faculty: View. Total quantity of the named item across all slots. Returns 0 if not present. Accounts for stacking — a single slot of 42 Fire Runes counts as 42. `ItemRef` resolution same as `has`. GUI: if you were to manually sum up all instances of an item. (exists)

- **`inventory.find(item: ItemRef)`** → `InvSlot?`. Faculty: View. Returns the first (lowest-indexed) slot matching the item, as an InvSlot instance carrying `.id`, `.amount`, `.name`, `.is_stackable`, `.is_wearable`, `.is_members_only`, `.command`. Returns `Null` if no match found. `ItemRef` resolution same as `has`. GUI: clicking an item to select it (shows the first occurrence). (exists)

- **`inventory.find_all(item: ItemRef)`** → `List<InvSlot>`. Faculty: View. Returns all slots matching the item as InvSlot instances (one per slot, even for stackables — so a stackable in one slot appears once, unstackables in multiple slots appear multiple times). Empty list if no match. `ItemRef` resolution same as `has`. (to build — perception gap)

- **`inventory.slot_of(item: ItemRef)`** → `Int?`. Faculty: View. Returns the 0-based slot index of the first matching slot, or `Null` if not found. Range: 0–29. `ItemRef` resolution same as `has`. GUI: internal indexing used by the client; routines may need it for equipment-slot reasoning. (exists)

### `world` — Visible entities and messages (Views only)

The `world` namespace exposes the current state of entities visible to the player: NPCs, other players, ground items, static locations, and boundaries. All members are read-only Views with no Action verbs.

#### `world` — Views

##### Lists and entity views

- **`world.players`** → `List<Player>`. Faculty: View. All players visible within the view radius. Each element is a Player instance with `.index`, `.name`, `.position`, `.is_friend` (stub: always false), `.in_combat_with` (stub: always null). Returns empty list when alone. (exists)

- **`world.npcs`** → `List<Npc>`. Faculty: View. All NPCs visible within the view radius. Each element is an Npc instance with `.index`, `.type_id`, `.position`, `.name` (from facts, or "" if def not loaded), `.combat_level` (computed: (atk+str+def)/4 + hits/4 from def), `.max_hp` (hits field from def), `.is_attackable` (from def), `.is_aggressive` (from def), `.hp_fraction` (stub: null until combat-target tracking lands, to build — perception gap). Returns empty list when no NPCs nearby. (exists)

- **`world.ground_items`** → Iterable<GroundItem>. Faculty: View. All ground items visible within the view radius. Also callable as a list (e.g., `for gi in world.ground_items { ... }`). Each element carries `.item_id`/`.id`, `.position`, `.name`, `.is_mine` (stub: always false, to build). Supports `.by_id(item_id, radius?: Int)` → GroundItem | Null (nearest matching item, optionally within radius), `.nearest` → GroundItem | Null (closest ground item by Chebyshev distance), `.all` → List<GroundItem>, `.length` → Int (count of visible items). (exists)

- **`world.locs`** → View. Faculty: View. Static, facts-derived location queries. Supports the following named categories (each returns a searchable list):
  - `.banks` → Loc list. Scenery/NPCs matching "bank".
  - `.altars` → Loc list. Matching "altar".
  - `.furnaces` → Loc list. Matching "furnace".
  - `.anvils` → Loc list. Matching "anvil".
  - `.fishing_spots` → Loc list. Matching "fishing".
  - `.trees` → Loc list. Matching "tree".
  - `.rocks` → Loc list. Matching "rock".
  - `.doors` → Loc list. Boundaries matching "door".
  - `.ladders` → Loc list. Boundaries matching "ladder".
  - `.shops` → Loc list. Matching "shop".
  - `.scenery` → Loc list. Every scenery def in facts.
  - `.spawn_points` → Loc list. Every NPC spawn point.
  - `.search(needle: String)` → Loc list. Generic substring search across scenery/boundary/NPC names.
  
  Each Loc list supports:
  - `.nearest(x?: Int, y?: Int)` → Placement | Null. Closest matching location within a 64-tile radius. Uses Chebyshev distance. Defaults center to `self.position` if no args.
  - `.within(radius: Int, x?: Int, y?: Int)` → List<Placement>. All matching placements within Chebyshev radius (default center: `self.position`).
  - `.names` → List<String>. The names of all matching defs.
  - `.length` → Int. Approximate count (count of matching def kinds, not actual placements).
  
  A Placement carries `.x`, `.y`, `.position`, `.name`, `.kind` ("scenery" | "npc_spawn" | "boundary"), `.def_id`, `.direction` (for boundaries; 0 for scenery). (exists)

- **`world.boundaries`** → View. Faculty: View. Dynamic boundary state (doors opened, webs cut, etc.). Supports:
  - `.at(x: Int, y: Int, dir?: Int)` → Boundary | Null. View for a specific boundary tile + direction (dir defaults to 0). Returns a synthetic view even if facts doesn't know the def. A Boundary carries `.x`, `.y`, `.direction`/`.dir`, `.position`, `.name` (from facts, or "door" if unknown).
  - `.is_open(x: Int, y: Int, dir?: Int)` → Bool. True iff a SEND_BOUNDARY_HANDLER marked this tile/dir as removed. False for unknown tiles (safe default).
  - `.dynamic` → List<[Int, Int, Int, Int]>. Each element is [x, y, dir, boundary_id] for all currently-removed boundaries (useful for debugging/persistence).
  (exists)

##### Recent-event snapshots

These are single-value buffers; each holds the most-recent event of its kind observed this session, or Null if never fired.

- **`world.last_chat`** → ChatRecord | Null. Faculty: View. Most recent public chat message. Fields: `.speaker` (String), `.message` (String), `.at` (Time). Method: `.contains(needle: String)` → Bool (case-insensitive substring match). (exists)

- **`world.last_pm`** → PMRecord | Null. Faculty: View. Most recent private message. Fields: `.sender`, `.message`, `.at`. Method: `.contains(needle)` → Bool. (exists)

- **`world.last_damage`** → DamageRecord | Null. Faculty: View. Most recent damage taken. Fields: `.amount` (Int), `.source` (String, empty if admin-issued; no attacker tracking yet), `.at` (Time). (exists)

- **`world.last_server_message`** → ServerMsgRecord | Null. Faculty: View. Most recent server-side system message. Fields: `.message` (String), `.at` (Time). Method: `.contains(needle)` → Bool. (exists)

- **`world.last_dialog_text`** → DialogTextRecord | Null. Faculty: View. Most recent NPC speech-bubble text (not the same as a dialog menu option). Fields: `.text` (String), `.at` (Time). (exists)

##### Dialog menu state

- **`world.dialog`** → View. Faculty: View. Current NPC dialog menu state (or Null-safe empty when no menu open). Supports:
  - `.is_open` → Bool. True iff a dialog menu is currently presented.
  - `.options` → List<String>. The menu option texts (empty list if none / menu not open).
  - `.find_option(substring: String)` → Int. 0-based index of the first option whose text contains (case-insensitive) the given substring, or -1 if no match. Substring matching is forgiving — quest text like "Yes, I'd like to help." can be matched with just "Yes".
  - `.clear()` → Null. Reset the cached options after resolving a menu (the server doesn't reliably signal menu close; callers must signal intent).
  (exists)

##### Trade/duel state (from flat world.trade / world.duel)

- **`world.trade`** → View. Faculty: View. Trade window state and offer screens. (Note: top-level `trade.*` is the frozen namespace per §6; these are still accessible via `world.trade.*` for backwards compatibility, but new code prefers `trade.*`.) Supports: `.is_active`, `.phase` (values: "request_sent", "open", "confirm", "completed", "cancelled", "none"), `.my_offer`, `.their_offer`, `.my_first_accepted`, `.their_first_accepted`, `.my_second_accepted`, `.their_second_accepted`, `.both_first_accepted`, `.both_second_accepted`, plus `.with` (opponent name), `.with_index` (opponent index). CORRECTION: The actual field names use "first"/"second" to distinguish offer-screen vs. confirm-screen accepts (not "accepted"/"confirmed" as in the exemplar contract at §7). (exists)

- **`world.duel`** → View. Faculty: View. Duel window state. Same shape as `world.trade` (using my_first_accepted / my_second_accepted / both_first_accepted / both_second_accepted) plus rule toggles: `.disallow_retreat`, `.disallow_magic`, `.disallow_prayer`, `.disallow_weapons` (all Bool). (exists)

##### Bank state

- **`world.bank`** → View. Faculty: View. Bank window state. (Note: top-level `bank.*` is the frozen namespace; these are still accessible here for backwards compatibility.) Supports: `.is_open` (Bool), `.max_size` (Int), `.used` (Int), `.free` (Int), `.slots` (List<[item_id, amount]>), `.has(item_id)` / `.count(item_id)` → Int (total quantity of item in bank). (exists)

---

#### `world` — Entity-view fields (per-instance)

These fields live on individual entity instances returned from the lists above.

##### Player instance `.` fields

- `.index` → Int. The player's server-side index (0–2047 typical). (exists)
- `.name` → String. Username. (exists)
- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y}`. (exists)
- `.is_friend` → Bool. Stub: always false until friend-list tracking lands. (to build — perception gap)
- `.in_combat_with` → Player | Null. Stub: always null until per-player combat target tracking lands. (to build — perception gap)

##### Npc instance `.` fields

- `.index` → Int. Server-side NPC index. (exists)
- `.type_id` → Int. NPC type ID (key into facts.NpcDefs). (exists)
- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y}`. (exists)
- `.name` → String. Loaded from facts NpcDef; returns "" if def not yet loaded. (exists)
- `.combat_level` → Int | Null. Computed from def: (attack + strength + defense) / 4 + hits / 4. Null if def not loaded. (exists)
- `.max_hp` → Int | Null. The "hits" field from the NPC def; null if def not loaded. (exists)
- `.is_attackable` → Bool. From NPC def; false if def not loaded. (exists)
- `.is_aggressive` → Bool. From NPC def; false if def not loaded. (exists)
- `.hp_fraction` → Float | Null. Stub: always null. Only your current combat target's health bar is visible in-game; tracking all NPC HP would require a perception feature not yet implemented. (to build — perception gap)
- `.in_combat_with` → Player | Npc | Null. Stub: always null. Would track which entity this NPC is fighting. (to build — perception gap)

##### GroundItem instance `.` fields

CORRECTION: GroundItem does NOT carry a `.quantity` field. The underlying GroundItemRecord contains only {X, Y, ItemID, LastSeen} and no quantity tracking.

- `.item_id` / `.id` → Int. Item ID. (exists)
- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y}`. (exists)
- `.name` → String. Item name from facts. (exists)
- `.is_mine` → Bool. Stub: always false until the host tracks 3-minute loot-ownership windows. Routines picking up loot should check this, but currently fall back to "try and see" with `pick_up()`'s `SERVER_REJECTED` error. (to build — perception gap)

##### Placement instance `.` fields (from locs queries)

- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y}`. (exists)
- `.name` → String. Scenery/boundary/NPC def name. (exists)
- `.kind` → String. "scenery" | "npc_spawn" | "boundary". (exists)
- `.def_id` → Int. Facts registry ID. (exists)
- `.direction` → Int. For boundaries, the direction; 0 for scenery. (exists)

##### Boundary instance `.` fields (from world.boundaries.at)

- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y}`. (exists)
- `.direction` / `.dir` → Int. Direction (0–3). (exists)
- `.name` → String. Boundary def name from facts, or "door" if unknown. (exists)

---

#### `world` — Events

The world namespace emits events for visibility changes and location updates:

- **`on item_appeared(item_id: Int, x: Int, y: Int)`** — A ground item entered visibility. Coords are absolute (resolved from relative offsets at packet arrival). Routines can filter by item_id; the full visible set is queryable via `world.ground_items`.

- **`on item_disappeared(item_id: Int, x: Int, y: Int)`** — A ground item left visibility (picked up or we walked away). Same coord handling.

- **`on coords_changed(x: Int, y: Int)`** — Our position updated (fires on every coord packet). Useful for polling-free pathfinding and tile-entry logic.

- **`on boundary_changed(x: Int, y: Int, dir: Int, id: Int)`** — A dynamic boundary changed (door opened, web cut, etc.). id=-1 means the boundary was removed. `world.boundaries.is_open(x, y, dir)` reflects the change before the handler fires.

- **`on bank_opened(max_size: Int)`** — The bank window opened. max_size is the slot capacity. `world.bank.slots` has the current contents.

- **`on bank_slot_update(slot: Int, item_id: Int, amount: Int)`** — A bank slot changed. amount=0 means the slot was emptied. `world.bank` reflects the change before the handler fires.

- **`on bank_closed()`** — The bank window closed. `world.bank.is_open` is now false.

- **`on chat_received(speaker: String, message: String)`** — Another player said something publicly. The speaker name is empty string (safe fallback) if the appearance packet hasn't arrived yet (rare).

- **`on private_message(sender: String, message: String)`** — Incoming PM from a friend.

- **`on server_message(text: String)`** — System message from the server (login success, "You must be carrying..." etc.). Mirrors `world.last_server_message`.

- **`on death()`** — We died. Fires once when the server confirms HP=0, **before** respawn. `self.hp` is reset to max immediately after, so this is the only reliable hook for death reactions.

- **`on trade_opened(other: Int)`** — Trade window opened with player at given server index. (exists)

- **`on trade_other_offer(items: List<[Int, Int]>)`** — The other party changed their offer. Each item is [item_id, amount]. (exists)

- **`on trade_other_accepted()`** — They accepted the offer screen. (exists)

- **`on trade_confirm_shown()`** — The confirm screen is now shown (both parties accepted the offer). (exists)

- **`on trade_closed(completed: Bool)`** — Trade window closed. completed=true iff the trade was finalized (items moved). (exists)

- **`on duel_opened(other: Int)`** — Duel window opened with player at given server index. (exists)

- **`on duel_other_offer(items: List<[Int, Int]>)`** — The opponent staked items. Each item is [item_id, amount]. (exists)

- **`on duel_settings_update(disallow_retreat: Bool, disallow_magic: Bool, disallow_prayer: Bool, disallow_weapons: Bool)`** — Duel rules changed. (exists)

- **`on duel_other_accepted()`** — The opponent accepted the offer screen. (exists)

- **`on duel_confirm_shown()`** — The confirm screen is now shown. (exists)

- **`on duel_closed(completed: Bool)`** — Duel window closed. completed=true iff the duel was finalized. (exists)

---

#### Type summary: Definition vs. Instance

| Category | Definition | Instance | Location |
|----------|-----------|----------|----------|
| Item | `ItemDef` (facts) | `InvSlot`, `GroundItem` | inventory, ground_items |
| Npc | `NpcDef` (facts) | `Npc` | world.npcs (carry `.def`) |
| Loc | `LocDef` (facts) | `Placement` | locs queries (facts-backed) |
| Player | — | `Player` | world.players |
| Boundary | `BoundaryDef` (facts) | `Boundary` | world.boundaries.at() |
| Spell | `SpellDef` (facts) | — | self.spells.book / by_id / by_name |
| Prayer | `PrayerDef` (facts) | — | self.prayers.book / by_id / by_name |

All instances carry a `.def` (or look it up lazily from facts) and live state (`.position`, `.quantity`/`.amount`, `.hp_fraction`, etc.).

---

#### Corrections from source verification

1. **GroundItem.quantity removed**: The provided schema lists `.quantity` on GroundItem instances, but the actual runtime/world.go GroundItemRecord has no quantity field. Instances carry only {ItemID, X, Y, LastSeen}. This field should be removed.

2. **Trade/Duel accept field naming**: The provided schema names trade accepts as `.accepted` / `.they_accepted` / `.both_accepted` and `.confirmed` / `.they_confirmed` / `.both_confirmed` (matching §7 exemplar). The actual runtime uses `.my_first_accepted` / `.their_first_accepted` / `.both_first_accepted` (first = offer screen) and `.my_second_accepted` / `.their_second_accepted` / `.both_second_accepted` (second = confirm screen). This is a CRITICAL discrepancy requiring the exemplar at §7 to be updated, OR the implementation to be renamed to match the contract.

3. **Bank fields expanded**: The provided schema referenced `world.bank.is_open` and `.slots` generically. The actual implementation adds `.max_size`, `.used`, `.free` fields and the methods `.has(item_id)` / `.count(item_id)` (aliases).

4. **world.trade.phase values**: The provided schema (§7) lists `{offer, confirm}` as valid phase values. The actual implementation uses `{request_sent, open, confirm, completed, cancelled, none}` (more granular).

5. **Duel-related events**: The provided schema does not document duel events (duel_opened, duel_other_offer, etc.), but they exist in runtime/dsl_events.go and are surfaced via world.duel. These should be added to the events section.

#### Tag notes

- **(exists)** — Implemented and tested in runtime/dsl_views.go + runtime/dsl_events.go.
- **(to build — perception gap)** — Specified here as the *target* the API should reach, but a perception-layer feature must land first. Currently stubbed (always null/false/safe default).

### `combat` — Views

- **`combat.target`** → `Npc|Player|Null`. Faculty: View. Current engagement target for the player. Returns the actively-attacked NPC or player (held in the host's combat state), or `Null` if not currently attacking anything. GUI: the name/health-bar overlay on screen (NPC combat or PVP duel). Nullability: `Null` when no target is engaged. Tag: **(to build — perception gap)**. Current: stub, always returns `Null`.

- **`combat.engaged`** → `Bool`. Faculty: View. True while the player is in active combat (trading blows with an opponent). False when idle, traveling, or between targets. GUI: the auto-retaliate loop and combat-end side effects (stop running, clear target). Nullability: never null; defaults to `false` until combat tracking lands. Tag: **(to build — perception gap)**. Current: stub, always returns `false`.

- **`combat.last_npc`** → `Npc|Null`. Faculty: View. The most recently attacked NPC, resolved live from `world.npcs` by stored index. Null if no NPC has been attacked this session, or if the previous target has since left view or died. Routines use this to detect target loss and retarget or flee. GUI: the name/health-bar that *was* on screen. Nullability: `Null` on target loss or never attacked. Tag: **(exists)**. Stored in `host.lastAttackedNpcIndex` (set on `attack()` dispatch, stays set across entity leaving view).

- **`combat.last_player`** → `Player|Null`. Faculty: View. The most recently attacked player, resolved live from the player roster. Null if no player has been attacked, or if the previous PVP opponent has logged out / left the area. Common in duels to track engagement for post-combat loot or honor tracking. GUI: the opponent's name/health-bar in a duel overlay. Nullability: `Null` on opponent loss. Tag: **(exists)**. Stored in `host.lastAttackedPlayerIndex` (set on `attack()` dispatch).

### `combat` — Actions (all → `Result<Null>`)

- **`combat.attack(target: Npc|Player|Int)`** → `Result<Null>`. Faculty: Action. Initiates combat with the target (NPC, player, etc.). Accepts an `Npc` view, `Player` view, or server index `Int`. Pathfinds from `self.position` to an adjacent tile, then sends the attack opcode (190 for NPC, 171 for player). Stores the target index in `host.lastAttackedNpcIndex` or `host.lastAttackedPlayerIndex` for `combat.last_npc`/`combat.last_player` queries. Errors: `OUT_OF_RANGE` if the target is too far to reach; `PATH_BLOCKED` if pathfinding fails; `SERVER_REJECTED` for other server rejections (target dead, not attackable, etc.). GUI: clicking an NPC or player's sprite to attack. Tag: **(exists)**. Note: spec name in `/dsl/spec/actions.go` is `"attack"`, and it functions as an alias for `combat.attack()`; within the `combat` namespace, it is promoted to `combat.attack()`.

- **`combat.set_style(style: String|Int)`** → `Result<Null>`. Faculty: Action. Changes the melee xp-split mode. Accepts a string (`"controlled"`, `"aggressive"`, `"accurate"`, `"defensive"`) or an int (0–3 with the same mapping: 0=controlled, 1=aggressive, 2=accurate, 3=defensive). Takes effect on the next attack tick; RSC does not send a confirmation packet, so observe `self.skills.<style>.xp` deltas to confirm application. Errors: none on the DSL side (the server silently applies the change or returns `SERVER_REJECTED`). GUI: clicking the combat-style toggle button to cycle modes. Tag: **(exists)**. Sends opcode 29 to the server; dispatches to `action.SetCombatStyle()`. Note: spec name is `"set_combat_style"` in `/dsl/spec/actions.go`; in the `combat` namespace, the frozen contract uses `combat.set_style()` as the canonical form.

- **`combat.retreat()`** → `Result<Null>`. Faculty: Action. Stops attacking and breaks combat. Used to disengage from the current target without attacking again. Returns `Ok` even if not in combat (idempotent). Errors: none documented — always succeeds (or `SERVER_REJECTED` on connection loss). GUI: walking away or clicking outside the combat area. Tag: **(to build)** — not yet wired in the action handler map. Future: will send an opcode (likely write-out a retreat sequence or silence the auto-retaliate).

### `combat` — Events

- **`on combat_started(target: Npc|Player)`** — Combat engaged; `target` is the entity being attacked. Fires when the first attack packet is sent (may fire repeatedly if target is swapped). Future work (step 8) will wire the actual combat-loop state to avoid duplicate fires.

- **`on combat_ended()`** — Player disengaged from combat (target died, player fled, or server ended the bout). Fires when `combat.engaged` transitions to false.

- **`on target_changed(from: Npc|Player|Null, to: Npc|Player|Null)`** — Explicit combat retarget during an active bout. Fired when `combat.target` changes while engaged. Allows retaliation policies and multi-target loops.

- **`on target_died(target: Npc)`** — The NPC being attacked died in combat. Fired on receiving a death packet; the target's health bar clears and `combat.target` becomes `Null` but `combat.engaged` may stay true for the client's retaliation tick.

---

#### Notes on the Combat Model

**Perception vs. Server State:**
- `combat.target` and `combat.engaged` are **stubs pending real combat-loop integration** (task #8). Until then, they return `Null`/`false`. The rest of the combat surface (`attack()`, `set_style()`, `last_npc`/`last_player`) works as specified.
- `combat.style` is a **write-only view** — set via `set_style()`; there is no read-side view yet. The XP deltas on the skill being trained are the only client-side confirmation. (To build: a read-side `combat.style` View that mirrors the most-recently-applied style.)
- `last_npc` and `last_player` are the only **stable perception** of recent combat — routines rely on these to detect target loss and pivot.

**Capability Boundary:**
- All Views are GUI-equivalent: a player sees the target bar, the style toggle, and the names of recent opponents.
- `attack()` and `set_style()` mirror direct client actions; `retreat()` mirrors leaving combat.
- Combat rules (toggle retreat in duels, auto-retaliate, multi-target priority) are **mind layer** — not exposed here; use `when`/`select` to build them.

**Migration from the Flat API:**

| today (flat) | frozen | why |
| --- | --- | --- |
| `attack(t)` | `combat.attack(t)` | namespaced; same semantics |
| (none) | `combat.engaged` | new perception field (to build — stub) |
| (none) | `combat.target` | new perception field (to build — stub) |
| `set_combat_style(...)` | `combat.set_style(...)` | namespaced, simplified arg shape |
| (none) | `combat.last_npc` | new perception field (exists) |
| (none) | `combat.last_player` | new perception field (exists) |
| (none) | `combat.retreat()` | new action (to build) |
| (none) | `combat.style` | read-side View (to build — not yet implemented) |

### `magic` — Views + Actions

#### `magic` — Views

- **`magic.level`** → `Int`. Faculty: View. Current Magic skill level (effective/current). Range: 1–99 (member-gated quest point soft cap). Null after logout. GUI: the Magic skill level in the stats panel. Tag: (exists, rename from `self.skills.magic.level`).

- **`magic.max_level`** → `Int`. Faculty: View. Base Magic skill level (unboostable max). Range: 1–99. GUI: the Magic stat level before boosts. Tag: (exists, rename from `self.skills.magic.max_level`).

- **`magic.can_cast(spell: SpellRef)`** → `Bool`. Faculty: View. True iff you have the required Magic level to cast the spell and pass any other static prerequisites (not dynamic rune/gear checks). SpellRef accepts Int(spellbook_id) | String(name) | SpellDef. Returns false if spell unknown or level insufficient. Null-safe: unknown spell returns false, never Null. GUI: the spell book UI highlights castable spells; hovering shows req level. Tag: (to build — perception gap; currently impossible to distinguish castable from un-castable without trying).

- **`magic.spells`** → `SpellsView`. Faculty: View. Namespace for spell catalog accessors. Tag: (exists, rename from `self.spells`).
  - **`magic.spells.book`** → `List<SpellDef>`. The full spellbook catalog (all spells in index order, regardless of level). GUI: the complete magic menu (greyed-out if not castable). Tag: (exists).
  - **`magic.spells.known`** → `List<SpellDef>`. Spells whose `req_level <= self.magic.level`. Live-updated per skill changes. GUI: the active (non-greyed) spells in the magic menu. Tag: (exists).
  - **`magic.spells.by_id(spell_id: Int)`** → `SpellDef | Null`. Lookup by spellbook index (0-based). Returns Null if id is out of range or unknown. Tag: (exists).
  - **`magic.spells.by_name(name: String)`** → `SpellDef | Null`. Lookup by canonical name (case-insensitive, e.g. "Fire Bolt"). Returns Null if not found. Tag: (exists).
  - **`magic.spells.has_runes_for(spell: SpellRef)`** → `Bool`. True iff inventory holds enough runes to cast the spell (checks all required runes from the spell's def). SpellRef accepts Int(id) | String(name) | SpellDef. Returns false if spell unknown or runes insufficient; never Null. Tag: (exists).
  - **`magic.spells.count`** → `Int`. Total spellbook size (constant per game build). Tag: (exists).

**SpellDef instance fields** (static properties of a spell, from facts registry):
- `id: Int` — spellbook index (0-based).
- `name: String` — canonical spell name ("Fire Bolt", "Teleport to House").
- `req_level: Int` — Magic skill required to cast.
- `type: Int` — spell category (1=self-targeted, 2=offensive, 3=debuff, 4=inventory-item, 5=teleother, 6=summon).
- `exp: Int` — XP awarded per cast (in tenths, matching OpenRSC convention).
- `description: String` — lore/flavor text.
- `members: Bool` — members-only flag.
- `evil: Bool` — legacy alignment flag (RSC era; not used in live gameplay).
- `runes: List<[Int, Int]>` — required runes as `[[item_id, count], ...]` pairs.

#### `magic` — Actions (all → `Result<Null>`)

- **`magic.cast(spell: SpellRef, target?: Target)`** → `Result<Null>`. Faculty: Action. Cast a spell, optionally on a target. Unifies the old `cast_on_self`, `cast_on_npc`, `cast_on_player`, `cast_on_land`, `cast_on_item` variants into one polymorphic verb.

  **Overload resolution (target type determines opcode):**
  - `target` omitted or `Null` — self-targeted cast (heal, teleport, buffs).
  - `target: Npc` — offensive/combat cast on NPC (Wind Strike, curse, etc.). Pathfinds adjacent if needed.
  - `target: Player` — PvP cast on player (server rejects outside legal duel/PvP zones). Pathfinds adjacent.
  - `target: Position` or `[Int, Int]` — tile-targeted AOE cast (Telekinetic Grab to ground item, AOE spells). Must be within range (exact range per spell; 15–25 tiles typical).
  - `target: GroundItem` — alchemy/item-conversion on a ground item (if the spell supports it; most don't).
  - `target: Loc` — rarely used (enchanting altars, if implemented); server rejects invalid targets.
  - **Ambiguous inventory-item cast (enchanting jewelry, high alchemy):** Pass the item-to-enchant as the positional arg; the spell via named `spell` or positional (overload resolution picks the spell by type). Example: `magic.cast(slot=7, spell="Enchant Amulet")` or `magic.cast("Gold Amulet", spell=12)`.

  **Semantics:**
  - Checks spell level and rune requirements (fails if insufficient).
  - Sends the appropriate opcode (non-targeted, on-NPC, on-player, on-tile, on-item per target shape).
  - Non-blocking (server confirmation may arrive later, reflected in world state + message events).
  - XP awarded server-side; routines see updated `self.xp` post-cast.

  **Errors:** 
  - `NO_SUCH_ITEM` if spell is unknown (name doesn't resolve).
  - `OUT_OF_RANGE` if target is too far (position-based checks; tile range limit).
  - `TARGET_OUT_OF_VIEW` if target NPC/Player has left the visible world since lookup.
  - `INVENTORY_EMPTY` if required runes are missing.
  - `SERVER_REJECTED` if server sends an error (e.g., "You need a target" for combat spells cast without target, "That spell is members-only").
  - `NOT_LOGGED_IN` if the session ended.
  - `INTERRUPTED` if the action was cancelled mid-flight (rare; ctx.Cancel).

  **GUI equivalent:** Click spell in spellbook, then (for targeted spells) click or right-click the target, or (for self-targeted) cast directly.

  **Tag:** (to build; stub exists at `cast_on_self`, etc.; unification + overload dispatch pending). Temporary fallback: use `cast_on_self(spell_id)`, `cast_on_npc(npc, spell_id)`, `cast_on_player(player, spell_id)`, `cast_on_land(x, y, spell_id)`, `cast_on_item(item, spell_id)` until the unified version lands.

#### `magic` — Events

- **`on spell_cast(caster: Player, spell: SpellDef, target?: Target)`** — You (caster=self) or another player cast a spell. `target` is Null for self-targeted, or the target entity (Npc, Player, Loc, Position, GroundItem, Boundary). Fires on both sides (caster's client + spectators). GUI: the spell animation plays in-world. Tag: (to build — spell message interception pending; currently requires parsing world.last_server_message or game chat for "You cast X").

- **`on spell_hit(caster: Player, spell: SpellDef, damage?: Int)`** — Caster's spell landed (hit). `damage` is Int (hitsplat value) for offensive spells, Null for non-damage (buffs, teleports, etc.). Fires on the target's client and the caster's. GUI: hitplat appears on the target. Tag: (to build — awaits damage-event packet decoder).

- **`on spell_resist(caster: Player, spell: SpellDef)`** — Spell was resisted / dodged (magic defense success). Caster and target see this. GUI: the "Resist" message appears. Tag: (to build).

---

#### Migration map (flat → frozen magic)

| today (flat) | frozen | why |
| --- | --- | --- |
| `cast_on_self(spell_id)` | `magic.cast(spell)` | Self-targeted cast; spell by id or name. |
| `cast_on_npc(npc, spell_id)` | `magic.cast(spell, target=npc)` | NPC combat-cast; overload dispatch by target type. |
| `cast_on_player(player, spell_id)` | `magic.cast(spell, target=player)` | PvP cast. |
| `cast_on_land(x, y, spell_id)` | `magic.cast(spell, target={x,y})` or `magic.cast(spell, target=Position{x,y})` | Tile-targeted AOE. |
| `cast_on_item(item, spell_id)` | `magic.cast(item, spell=spell_id)` | Inventory-item cast (alchemy, enchanting). Spell passed via named arg to disambiguate from positional item. |
| `self.spells.*` | `magic.spells.*` | Promoted to root namespace per §6 (subsystem naming). |
| — | `magic.level` | Alias for `self.skills.magic.level` (skill index 6) + convenience + future extensibility (e.g., per-book level variations). |
| — | `magic.max_level` | Alias for `self.skills.magic.max_level` (skill index 6). |
| — | `magic.can_cast(spell)` | Pure-read check: "can I cast this?" (stub not yet a public View; to build). |

### `prayer` — Views + Actions

#### `prayer` — Views

- **`prayer.points`** → `Int`. Faculty: View. Current prayer points (shorthand for `self.prayer`). Range 0+. Never null. GUI: the prayer bar. (exists)
- **`prayer.active(slot)`** → `Bool`. Faculty: View. Whether the prayer at the given slot index is currently on. Bounds-checked; never null. GUI: lit prayer buttons in the prayer book. (exists)
- **`prayer.book`** → `List<PrayerDef>`. Faculty: View. The static prayer catalog (from facts). Never null. GUI: the prayer book interface. (exists)
- **`prayer.by_id(id)`** → `PrayerDef | Null`. Faculty: View. Look up a prayer by slot index (0–14). Null if out of range. (exists)
- **`prayer.by_name(name)`** → `PrayerDef | Null`. Faculty: View. Look up a prayer by name (case-insensitive, from facts). Null if no match. (exists)

Each `PrayerDef` carries `.id` / `.name` / `.req_level` / `.drain_rate` / `.description` (all static from facts).

#### `prayer` — Actions (all → `Result<Null>`)

- **`prayer.activate(prayer_index: Int)`** → `Result<Null>`. Faculty: Action. Activate a prayer slot (0–13). The server silently rejects if prayer level is too low or prayer points are zero; check the result or poll `prayer.active(N)` to confirm. Errors: `SERVER_REJECTED` (level requirement not met, no prayer points, etc.). GUI: click the prayer icon in the prayer panel. Tag: (rename from `activate_prayer`; exists).
- **`prayer.deactivate(prayer_index: Int)`** → `Result<Null>`. Faculty: Action. Deactivate (toggle off) a prayer slot. Errors: `SERVER_REJECTED` (rarely; shouldn't fail). GUI: click the active prayer icon to toggle off. Tag: (rename from `deactivate_prayer`; exists).

#### `prayer` — Notes

- **Namespacing (§6):** Promotes `self.prayers.*` → top-level `prayer.*` per the contract.
- **Migration:** `activate_prayer(name)` → `prayer.activate(prayer_index)`; `deactivate_prayer(name)` → `prayer.deactivate(prayer_index)`.
- **No events** in this namespace (acknowledged; future "to build").
- **No perception gap:** all player-visible prayer state (points, active list, slot activity, catalog) is surfaced.

### `trade` — Frozen API surface

#### `trade` — Views

- **`trade.is_active`** → `Bool`. Faculty: View. True while a trade window is open (phase ≠ "completed" and ≠ "cancelled"). GUI: the trade window is visible on screen. Tag: (exists).

- **`trade.phase`** → `String` ∈ {`"none"`, `"request_sent"`, `"open"`, `"confirm"`, `"completed"`, `"cancelled"`}. Faculty: View. Which screen is showing. `"none"` when no trade is active; `"request_sent"` after initiating a request (before symmetric acceptance); `"open"` on the offer screen (both sides accepted request); `"confirm"` on the final review screen; `"completed"` or `"cancelled"` on terminal. GUI: none (internal state). Tag: (exists).

- **`trade.with`** → `String | Null`. Faculty: View. The opponent's username, or null if no trade active. GUI: opponent display. Tag: (exists).

- **`trade.with_index`** → `Int`. Faculty: View. The opponent's server-side player index, or 0 if no trade active. GUI: none. Tag: (exists).

- **`trade.my_offer`** → `List<[Int, Int]>`. Faculty: View. Items you have put up for trade; each element is `[item_id: Int, quantity: Int]`. Empty list (never null) if no trade or you haven't offered yet. GUI: your side of the offer panel. Tag: (exists).

- **`trade.their_offer`** → `List<[Int, Int]>`. Faculty: View. The opponent's offered items, same shape. Empty list if no trade or they haven't offered yet. GUI: their side of the offer panel. Tag: (exists).

- **`trade.my_first_accepted`** → `Bool`. Faculty: View. True iff you clicked "Accept" on the **offer** screen (screen 1). Resets to false if either party changes their offer. GUI: first Accept button visual state. Tag: (rename to `trade.accepted`; exists).

- **`trade.their_first_accepted`** → `Bool`. Faculty: View. True iff the opponent clicked "Accept" on the **offer** screen. Resets on offer change. GUI: opponent's first Accept state. Tag: (rename to `trade.they_accepted`; exists).

- **`trade.both_first_accepted`** → `Bool`. Faculty: View. True iff both parties accepted the offer screen → the confirm screen appears. Shorthand for `.my_first_accepted && .their_first_accepted`. GUI: both-accepted indicator. Tag: (rename to `trade.both_accepted`; exists).

- **`trade.my_second_accepted`** → `Bool`. Faculty: View. True iff you clicked "Accept" on the **confirm** screen (screen 2). Only meaningful after `both_first_accepted`. GUI: second Accept button visual state. Tag: (rename to `trade.confirmed`; exists).

- **`trade.their_second_accepted`** → `Bool`. Faculty: View. True iff the opponent clicked "Accept" on the **confirm** screen. Only meaningful after both parties' first accept. GUI: opponent's confirm state. Tag: (rename to `trade.they_confirmed`; exists).

- **`trade.both_second_accepted`** → `Bool`. Faculty: View. True iff both parties confirmed the final screen → trade executes. Shorthand for `.my_second_accepted && .their_second_accepted`. GUI: trade-complete indicator. Tag: (rename to `trade.both_confirmed`; exists).

#### `trade` — Actions (all → `Result<Null>`)

- **`trade_request(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent to `p`, then sends a trade request (wire opcode 142). Dual-purpose: initiates a new trade request AND accepts an incoming one — in RSC both sides must request each other (symmetric handshake) for the window to open. Errors: `TARGET_OUT_OF_VIEW` if `p` (player view, server-index Int, or player-name String) is not visible in the world. GUI: right-click player → "Trade with". Tag: (rename to `trade.request()` in frozen API; merges old `trade_request` + `accept_trade`; exists).

- **`accept_trade(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Accepts an incoming trade by re-sending the trade-request packet to the original requester `p` (player view, name String, or server-index Int). String names resolve via world.Players deterministically. Symmetric RSC handshake: both sides must request each other for the window to open. Errors: `TARGET_OUT_OF_VIEW` if `p` is not visible. GUI: responding to incoming "Trade with" right-click. Tag: (rename to `trade.request()`; merges with `trade_request` in frozen API; exists).

- **`offer_trade(items: List<[Int, Int]>)`** → `Result<Null>`. Faculty: Action. Sets/replaces your offered items; each element is `[item_id: Int, amount: Int]`. Per server rule, any offer change resets both parties' accept flags on the offer screen (`.my_first_accepted` and `.their_first_accepted` become false). Errors: `TRADE_NOT_ACTIVE` if no trade is open; `NO_SUCH_ITEM` if you don't hold an offered item. GUI: drag items into your offer panel. Tag: (rename to `trade.offer()` in frozen API; exists).

- **`confirm_trade()`** → `Result<Null>`. Faculty: Action. Clicks "Accept" on the **offer** screen (screen 1). Both sides must accept before the confirm screen appears. Idempotent: re-calling after you've already accepted is a no-op, and it re-fires automatically if an offer change reset your acceptance. Errors: `TRADE_NOT_ACTIVE` if no trade is active. GUI: the first "Accept" button. Tag: (rename to `trade.accept()` in frozen API; exists).

- **`finalize_trade()`** → `Result<Null>`. Faculty: Action. Clicks "Accept" on the **confirm** screen (screen 2), completing your half of the trade. The confirm screen only appears after `trade.both_first_accepted`. Idempotent on extra calls. Errors: `TRADE_NOT_ACTIVE` if no trade is active; returns error if called before the offer screen is accepted (check `.my_first_accepted` first). GUI: the second "Accept" button. Tag: (rename to `trade.confirm()` in frozen API; exists).

- **`decline_trade()`** → `Result<Null>`. Faculty: Action. Declines/closes the trade window (can be called at any phase). Marks the trade phase → "cancelled". Errors: `TRADE_NOT_ACTIVE` if no trade is active. GUI: close button / cancel. Tag: (rename to `trade.decline()` in frozen API; exists).

#### `trade` — Events

- **`on trade_request(from: String)`** — Another player initiated a trade request with you. `from` is the requester's player name (String). Call `accept_trade(from)` to accept the symmetric handshake. OpenRSC's notification packet carries name only, not server index; lookup via world.Players. Tag: (rename to `on trade_request` in frozen event name; currently spec'd in dsl/spec/events.go but runtime translator conditionally wires; note: currently runtime support may be incomplete).

- **`on trade_opened(other_index: Int)`** — Both players have accepted the initial trade request; the offer screen is now open. `other_index` is the opponent's server-side player index. Use `trade.my_offer`, `trade.their_offer`, and `trade.with` / `trade.with_index` to read state. GUI: the offer panel appears. Tag: (rename from `other_index: Int` → `other: Player` in frozen API; currently returns Int; exists).

- **`on trade_other_offer(items: List<[Int, Int]>)`** — The opponent updated their offered items. `items` is the new list of `[item_id, amount]` pairs. Mirrors `trade.their_offer`. Both parties' offer-screen accept flags reset. GUI: opponent's offer side updates. Tag: (exists).

- **`on trade_other_accepted()`** — The opponent clicked "Accept" on the current screen (offer or confirm). No parameter. Allows reacting before both-sides acceptance. GUI: opponent's Accept button visual state changes. Tag: (exists).

- **`on trade_closed(completed: Bool)`** — The trade ended. `completed` is true iff both sides confirmed the final screen (goods exchanged); false on decline/cancel. Server protocol close packet has no completion bit; implementation infers from world.Trade.Phase == "completed" after Apply. GUI: trade window closes. Tag: (exists).

- **`on trade_confirm_shown()`** — Server pushed the final confirm-window — both sides have first-accepted and we're on the final review screen. No parameter. GUI: confirm panel appears. Tag: (exists).

#### Capability boundary & implementation notes

- All Views report screen-visible state or cached world-mirror state (opponent name, phase, accepted flags, offer lists). The `.with_index` is a resolved lookup convenience matching Npc-resolution rules in §2.
- The `.my_offer` / `.their_offer` are lists of items each side has proposed; the GUI player sees them by dragging to the window. Format is List<[Int, Int]> (item_id, quantity pairs), not abstract List<Item>.
- The twin-accept flow (offer screen, then confirm screen) is the frozen, immutable RSC server protocol — the API surfaces it directly as separate `.confirm_trade()` / `.finalize_trade()` verbs and separate first/second accept flags.
- Trade requires adjacency to initiate (opcode 142); the action automatically pathfinds to the target player's last-known position.
- The `trade_request` and `accept_trade` old verbs are dual-purpose (same opcode 142), so the frozen API consolidates them to `trade.request(p: PlayerRef)`.
- Phase transitions: "none" (idle) → "request_sent" (after initiating) → "open" (both accepted request) → "confirm" (both first-accepted) → "completed" or "cancelled" (terminal).

#### Frozen nomenclature mapping (§10)

| Today (flat old API) | Frozen namespaced API | Reason |
| --- | --- | --- |
| `trade_request(p)` + `accept_trade(p)` | `trade.request(p)` | Both send opcode 142 (symmetric RSC handshake); one player action, one verb. |
| `offer_trade([...])` | `trade.offer([...])` | Promote to namespace. |
| `confirm_trade()` | `trade.accept()` | First accept-click (offer screen). Rename clarifies intent (was "confirm" — confusing with final confirm). |
| `finalize_trade()` | `trade.confirm()` | Second accept-click (confirm screen). |
| `decline_trade()` | `trade.decline()` | Promote to namespace. |
| `world.trade.is_active` / `phase` / `my_offer` / `their_offer` | `trade.is_active` / `phase` / `my_offer` / `their_offer` | Promote from world.trade to root. |
| `world.trade.my_first_accepted` / `their_first_accepted` / `both_first_accepted` | `trade.accepted` / `they_accepted` / `both_accepted` | Offer-screen flags; drop "first_" prefix for clarity. |
| `world.trade.my_second_accepted` / `their_second_accepted` / `both_second_accepted` | `trade.confirmed` / `they_confirmed` / `both_confirmed` | Confirm-screen flags; rename "second_" → "confirmed". |

#### Tag summary

- **Views**: `trade.is_active` (exists), `trade.phase` (exists), `trade.with` (exists), `trade.with_index` (exists), `trade.my_offer` (exists), `trade.their_offer` (exists), `trade.my_first_accepted` (rename to `trade.accepted`), `trade.their_first_accepted` (rename to `trade.they_accepted`), `trade.both_first_accepted` (rename to `trade.both_accepted`), `trade.my_second_accepted` (rename to `trade.confirmed`), `trade.their_second_accepted` (rename to `trade.they_confirmed`), `trade.both_second_accepted` (rename to `trade.both_confirmed`).
- **Actions**: `trade_request` (rename to `trade.request()`, merges with `accept_trade`), `accept_trade` (merged into `trade.request()`), `offer_trade` (rename to `trade.offer()`), `confirm_trade` (rename to `trade.accept()`), `finalize_trade` (rename to `trade.confirm()`), `decline_trade` (rename to `trade.decline()`).
- **Events**: `trade_request` (exists, spec'd but runtime translator may need completion check), `trade_opened` (exists; frozen API: param rename to `other: Player`), `trade_other_offer` (exists), `trade_other_accepted` (exists), `trade_closed` (exists), `trade_confirm_shown` (exists).

### `bank` — Views + Actions

#### `bank` — Views

- **`bank.is_open`** → `Bool`. Faculty: View. True while a bank window is open. GUI: the bank window is on screen. (exists)
- **`bank.max_size`** → `Int`. Bank slot capacity. Returns 0 if bank not open. GUI: visible from the window chrome. (exists)
- **`bank.used`** → `Int`. Number of occupied bank slots. Returns 0 if bank not open. GUI: inferred from slot display. (exists)
- **`bank.free`** → `Int`. Available slots (max_size - used). Returns 0 if bank not open. GUI: free visual space. (exists)
- **`bank.slots`** → `List<[Int, Int]>`. Bank contents; each element is `[item_id, amount]`. Empty list if bank not open (never null). GUI: the itemized inventory pane. (exists)
- **`bank.count(item: ItemRef)`** → `Int`. Total quantity of an item in the bank (0 if absent). Alias: `bank.has()`. GUI: calculated from visible slots. (exists)

#### `bank` — Actions (all → `Result<Null>`)

- **`bank.open(banker: NpcRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent to `banker`, then opens the bank UI via dialog tree (equivalent to talk_to). Does NOT advance the dialog to the bank screen — DSL must call `answer()` to select the bank option. Errors: `NO_SUCH_ITEM`-class if `banker` does not resolve to a visible NPC; `TARGET_OUT_OF_VIEW` if banker is not visible; `PATH_BLOCKED` if unreachable. GUI: walk to banker, right-click → "Talk To". (rename from `open_bank`)
- **`bank.deposit(item: ItemRef, amount: Int)`** → `Result<Null>`. Faculty: Action. Deposits `amount` of the item from inventory into the open bank. Errors: `BANK_NOT_OPEN` (bank window must be open server-side via talk_to to a banker); `NO_SUCH_ITEM` if inventory does not hold the item or insufficient quantity; `SERVER_REJECTED` for other server rejections. GUI: drag item into bank window. (exists)
- **`bank.withdraw(item: ItemRef, amount: Int)`** → `Result<Null>`. Faculty: Action. Withdraws `amount` of the item from the open bank into inventory. Errors: `BANK_NOT_OPEN`; `NO_SUCH_ITEM` if bank does not hold the item; `INVENTORY_FULL` if inventory cannot accept the item (hard limit: 30 slots); `SERVER_REJECTED` for other server rejections. GUI: drag item from bank window. (exists)
- **`bank.deposit_all(keep?: List<Int>)`** → `Result<Null>`. Faculty: Action. Deposits every inventory item except those in the optional `keep` list (item IDs). Sleeps 700ms between deposits to allow the mirror to sync. Errors: `BANK_NOT_OPEN`; individual item deposits may fail per `deposit()` rules (no early exit — continues on item-level errors, logs warnings). GUI: emulated via repeated drag (not a single player action; a routine convenience). (to build — implemented in runtime/bank.go Host.DepositAll but not registered in dsl/spec/actions.go)
- **`bank.close()`** → `Result<Null>`. Faculty: Action. Closes the bank window. Clears the server's `accessingBank` flag. Errors: None (idempotent; succeeds even if not open). GUI: close-button or move away. (rename from `close_bank`)

#### `bank` — Events

- **`on bank_opened(max_size: Int)`** — Bank window opened (talk_to banker → select bank from dialog, or right-click banker → "Bank"). `max_size` is the slot capacity (typically 48 for free players, 54 for members). `bank.slots` has the contents. (exists)
- **`on bank_slot_update(slot: Int, item_id: Int, amount: Int)`** — A bank slot was updated. `amount=0` means the slot was emptied. `bank` state reflects the change before the handler fires. (exists)
- **`on bank_closed()`** — Bank window closed. `bank.is_open` is false from this point. (exists)

**Tags & Notes:**
- All Views: report the mirror state (no round-trip to server).
- All Actions: send packets, return `Result<Null>` on immediate acceptance (async consequences via mirror updates + events).
- **Naming:** `open_bank` (action) → `bank.open(banker)` per namespacing rule §6.1; `close_bank` → `bank.close()`. Current entries are `open_bank`, `deposit`, `withdraw`, `close_bank` in dsl/spec/actions.go — promote to the `bank.*` namespace.
- **Unwired:** `bank.deposit_all(keep)` is fully implemented in runtime/bank.go as Host.DepositAll (iterates inventory slots, deposits non-keyed items with 700ms sleeps, logs warnings on per-item failures, returns nil on completion) but NOT registered in dsl/spec/actions.go — tag as (to build). Wire by adding an ActionSpec entry and dslDepositAll wrapper in runtime/dsl_actions.go.
- **Error code coverage:** wrapServerErr in runtime/dsl_actions.go maps string-match patterns; "inventory full" → `INVENTORY_FULL`, "path" / "stalled" → `PATH_BLOCKED`, "timeout" → `ACTION_TIMEOUT`, else `SERVER_REJECTED`. BANK_NOT_OPEN error code exists in dsl/interp/error.go but is not explicitly checked in wrapServerErr — bank actions currently rely on Host.BankDeposit/Withdraw returning formatted-string errors, which the string-match classifier doesn't distinguish from generic failures. Future: Host.BankDeposit/Withdraw should return a typed error (e.g., *BankNotOpenError) so wrapServerErr can return the exact code.
- **Perception gap:** None — all views (is_open, max_size, used, free, slots, count) are GUI-visible and fully implemented in runtime/dsl_views.go bankView.Get().
- **No server secrets:** All views report only what the player sees on the bank window UI.

### `duel` — Views + Actions

#### `duel` — Views

- **`duel.is_active`** → `Bool`. Faculty: View. True while a duel window is open. GUI: the duel window is on screen. (exists)
- **`duel.phase`** → `String` ∈ {`"none"`, `"request_sent"`, `"open"`, `"confirm"`, `"completed"`, `"cancelled"`}. Which phase the duel is in. (exists)
- **`duel.my_stake`** → `List<[Int, Int]>`. Items you have put up; each element is `[item_id, amount]` (empty list, never null). GUI: your side of the offer. (exists)
- **`duel.their_stake`** → `List<[Int, Int]>`. The opponent's items. (exists)
- **`duel.accepted`** → `Bool`. You clicked Accept on the **offer** screen. (rename from `my_first_accepted`)
- **`duel.they_accepted`** → `Bool`. They did. (rename from `their_first_accepted`)
- **`duel.both_accepted`** → `Bool`. Both did → the confirm screen opens. (rename from `both_first_accepted`)
- **`duel.confirmed`** → `Bool`. You clicked Accept on the **confirm** screen. (rename from `my_second_accepted`)
- **`duel.they_confirmed`** → `Bool`. They did on the confirm screen. (rename from `their_second_accepted`)
- **`duel.both_confirmed`** → `Bool`. Both confirmed the fight. (rename from `both_second_accepted`)
- **`duel.opponent`** → `String`. The opponent's username (empty if inactive). (rename from `with`)
- **`duel.opponent_index`** → `Int`. The opponent's server-side index (0 if inactive). (rename from `with_index`)
- **`duel.disallow_retreat`** → `Bool`. Retreat rule is active. (exists)
- **`duel.disallow_magic`** → `Bool`. Magic rule is active. (exists)
- **`duel.disallow_prayer`** → `Bool`. Prayer rule is active. (exists)
- **`duel.disallow_weapons`** → `Bool`. Weapons rule is active. (exists)

#### `duel` — Actions (all → `Result<Null>`)

- **`duel.request(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent, then sends a duel request to `p` (wire opcode 103). In RSC this one action both *initiates* a duel and *accepts an incoming* one — clicking the player serves both, and mutual requests open the window (so there is no separate "accept request" verb). Errors: `OUT_OF_RANGE` if `p` is not a visible or reachable player. GUI: right-click player → "Challenge". (rename from `duel_request` and `accept_duel`)
- **`duel.stake(items: List<[Int, Int]>)`** → `Result<Null>`. Sets/replaces your stake; each element is `[item_id, amount]`. Per the server rule this **resets both parties' offer-accepts**. Errors: `TRADE_NOT_ACTIVE` if no duel is active; `NO_SUCH_ITEM` if you don't hold a staked item. GUI: drag items into the window. (rename from `offer_duel`)
- **`duel.set_rules(retreat?: Bool, magic?: Bool, prayer?: Bool, weapons?: Bool)`** → `Result<Null>`. Sets the four rule toggles; pass `true` to disallow, `false` or omit to allow. Either side can change rules from the offer screen; the server unifies them and broadcasts the result. A change resets both first-accept flags. Errors: `TRADE_NOT_ACTIVE`. GUI: toggle buttons on the offer screen. (exists)
- **`duel.accept()`** → `Result<Null>`. Clicks "Accept" on the **offer** screen (screen 1). Idempotent; re-fires automatically if a stake change reset your accept. Errors: `TRADE_NOT_ACTIVE`. GUI: the first Accept button. (rename from `accept_duel_offer`)
- **`duel.confirm()`** → `Result<Null>`. Clicks "Accept" on the **confirm** screen (screen 2), completing your half. The confirm screen only opens after `duel.both_accepted`. Errors: `TRADE_NOT_ACTIVE`; `NOT_IMPLEMENTED`/not-ready if called before the offer screen is accepted. GUI: the second Accept button. (rename from `accept_duel_confirm`)
- **`duel.decline()`** → `Result<Null>`. Declines/closes the duel at any phase. Errors: `TRADE_NOT_ACTIVE`. GUI: close the window. (rename from `decline_duel`)

#### `duel` — Events

- **`on duel_requested(opponent: String)`** — duel request received from `opponent` (their username). (to build)
- **`on duel_opened(opponent: Player)`** — window opened with `opponent` after both sides sent a request. (to build)
- **`on duel_stake_updated(items: List<[Int, Int]>)`** — opponent changed their stake. (to build)
- **`on duel_rules_changed(disallow_retreat: Bool, disallow_magic: Bool, disallow_prayer: Bool, disallow_weapons: Bool)`** — opponent changed the rules. (to build)
- **`on duel_other_accepted()`** — opponent accepted the offer screen. (to build)
- **`on duel_closed(completed: Bool)`** — closed; `completed` iff the duel began (both sides confirmed). (to build)

### `shop` — Views + Actions (to build)

The `shop` namespace mirrors the bank/trade/duel pattern. It is a **specification-only** surface: only the inbound opcodes (101 InShopOpen, 137 InShopClose in `proto/v235/inbound_opcodes.go`) exist in the runtime today; all Views, Actions, and Events below are tagged `(to build)` pending implementation.

#### `shop` — Views

- **`shop.is_open`** → `Bool`. Faculty: View. True while a shop window is open. GUI: the shop window visible on screen. (to build)
- **`shop.stock(item: ItemRef)`** → `Int`. Faculty: View. Quantity of the item currently in stock (0 if not stocked). GUI: the quantity shown in the shop list. (to build)
- **`shop.price(item: ItemRef)`** → `Int`. Faculty: View. The unit buy price of the item. Sell price is derived as 40% of buy price (RSC rule); the API exposes buy price only. GUI: the unit price on hover/click. (to build)

#### `shop` — Actions (all → `Result<Null>`)

- **`shop.buy(item: ItemRef, qty: Int)`** → `Result<Null>`. Faculty: Action. Buys `qty` of the item from the open shop. Errors: `SHOP_NOT_OPEN` (must be added to error.go); `SERVER_REJECTED` (out of stock, member-only item for a free player, insufficient coins). GUI: click item → choose qty → confirm. (to build)
- **`shop.sell(item: ItemRef, qty: Int)`** → `Result<Null>`. Faculty: Action. Sells `qty` of the item to the open shop. Errors: `SHOP_NOT_OPEN`; `NO_SUCH_ITEM` if inventory does not hold it; `SERVER_REJECTED` (shop won't buy it). GUI: right-click item → sell/sell-all → confirm. (to build)
- **`shop.close()`** → `Result<Null>`. Faculty: Action. Closes the shop window. GUI: close button or ESC. (to build)

#### `shop` — Events

- **`on shop_opened(npc: Npc)`** — Shop window opened via NPC interaction (right-click → Trade or talk_to + dialog). (to build)
- **`on shop_stock_update(item_id: Int, new_stock: Int)`** — A shop slot's stock changed (purchase, sale, or restock). (to build)
- **`on shop_closed()`** — Shop window closed. (to build)

#### `shop` — Design notes

- **No separate `shop.open()` action:** shop opening is via NPC interaction (right-click → Trade or talk_to + dialog), mirroring bank.
- **No per-NPC shop routing:** single-shop assumption matches the bank/trade/duel pattern.
- **Sell-price derivation:** RSC's strict 40%-of-buy rule is encoded server-side; the API exposes buy price only.
- **Restock events:** OpenRSC restocks on fixed intervals; `shop_stock_update` allows dynamic reactions.
- **Implementation roadmap:** add `SHOP_NOT_OPEN` to `dsl/interp/error.go`; add accessor/action/event specs to `dsl/spec/*`; add `ShopState` to `world/world.go`; add `shopView` to `runtime/dsl_views.go`; add `dslShopBuy`/`dslShopSell`/`dslShopClose` to `runtime/dsl_actions.go`; wire shop event translation in `runtime/dsl_events.go`; decode opcodes 101/137 in the inbound handler.

### Ambient verbs — top-level Actions and Primitives

**Ambient verbs** are top-level Actions and Primitives owned by no subsystem root — they model the core interaction loops of a player: movement, generic item interaction, NPC engagement, scenery manipulation, combat, social chat, and session control. They read like player-speech rather than API subdomains. All are documented per §7 depth bar (signature, faculty, semantics, errors, nullability, GUI equivalent, tags).

#### Movement & Navigation

- **`walk_to(x: Int, y: Int, attempt_open_doors?: Bool)`** → `Result<Null>`. Faculty: Action. Walk to the absolute-coordinate tile (x, y); blocks until arrival or failure. Optional `attempt_open_doors` (default true) controls whether doors are auto-opened on the path or the walk fails (strict mode for scouts). Errors: `PATH_BLOCKED` (no path exists after all obstacles); `DOOR_LOCKED` (a door on the path couldn't be opened; reason includes server prose like "you need a key"); `OUT_OF_RANGE` (destination too far for one walk cycle). GUI: click the minimap or right-click a tile and walk. (exists)
- **`walk_path(corners: List<[Int, Int]>)`** → `Result<Null>`. Faculty: Action. Pre-planned multi-corner walk; each element is `[x, y]`. Max 25 corners per packet; for longer routes chunk via repeated calls. Server interpolates each segment. Useful for cognitive routes the routine planned itself. Errors: `PATH_BLOCKED` (packet send fails). GUI: used by the cognitive layer for pre-computed routes; no direct GUI equivalent. (exists)
- **`is_reachable(x: Int, y: Int)`** → `Bool` OR **`is_reachable(target: View)`** → `Bool`. Faculty: View (pure; no packet). Chebyshev pathfinding from self.position to (x, y). Bounded by 96-tile FOV grid; cross-region checks use sequential is_reachable along a planned chain. Target can be any view with .x/.y (Player, Npc, Loc, GroundItem, Position, etc.) or named (x=X, y=Y). Returns true iff reachable, false if not. GUI: none (used by pathfinding logic). (exists)

#### Interaction: Items & Inventory

- **`pick_up(ground_item: GroundItem)`** → `Result<GroundItem>`. Faculty: Action. Walk adjacent and pick up a ground item; returns the picked-up item-view on success so routines can do `got = pick_up!(item); say(got.name)`. Errors: `OUT_OF_RANGE` (too far to reach); `INVENTORY_FULL` (no slot available); `TARGET_OUT_OF_VIEW` (item no longer visible). GUI: right-click ground item → "Take". (exists)
- **`drop(item: ItemRef, amount?: Int)`** → `Result<Null>`. Faculty: Action. Drop the item from inventory to the ground; `amount` controls the count dropped (defaults to full stack). ItemRef can be an InvSlot, Int(slot_index), String(name), or ItemDef. Errors: `NO_SUCH_ITEM` (not in inventory); `OUT_OF_RANGE` (can't drop in current position). GUI: right-click inventory item → "Drop". (exists)
- **`eat(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Use an item from inventory (Eat/Drink/Bury — server decides by item def). Alias of `use_inventory_default(item)`. ItemRef as in `drop()`. Errors: `NO_SUCH_ITEM` (not in inventory). GUI: right-click food item → "Eat", or right-click herbs → "Clean", or right-click bones → "Bury". (exists)
- **`equip(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Move an inventory item into its worn slot (armor, weapons, etc.). Server enforces level requirements and slot conflicts. ItemRef: InvSlot, Int(slot_index 0–29), String(name), or ItemDef. Errors: `NO_SUCH_ITEM` (not in inventory); `SERVER_REJECTED` (level requirement not met, slot conflict, or item not wearable). GUI: right-click inventory item → "Equip", or drag to equipment slot. (exists)
- **`unequip(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Return a wielded item to inventory. ItemRef: InvSlot or String(name). Errors: `NO_SUCH_ITEM` (not wielded); `INVENTORY_FULL` (inventory is full). GUI: click the equipped item in equipment panel → "Remove", or right-click → "Take off". (exists)

#### Interaction: NPCs & Scenery

- **`talk_to(npc: NpcRef)`** → `Result<Null>`. Faculty: Action. Walk adjacent and open the NPC's dialog. NpcRef: Npc view (from world.npcs), Int(server_index). Errors: `OUT_OF_RANGE` (too far); `TARGET_OUT_OF_VIEW` (NPC no longer visible); `PATH_BLOCKED` (can't reach adjacent tile). GUI: right-click NPC → "Talk". (exists)
- **`pickpocket(npc: NpcRef)`** → `Result<Null>`. Faculty: Action. Walk adjacent and fire the NPC's primary action command (command1) — e.g., "Pickpocket" on a Man or Robed Gnome. One attempt per call; loop for several attempts. Functionally identical to `npc_command(npc)`; `pickpocket` is the canonical frozen-contract name. NpcRef: Npc view or Int(server_index). Errors: `OUT_OF_RANGE`; `TARGET_OUT_OF_VIEW`; `PATH_BLOCKED`. GUI: right-click thievable NPC → "Pickpocket". (exists, canonical from `npc_command`)
- **`interact_at(target: Position | Loc, option?: Int)`** → `Result<Null>`. Faculty: Action. Primary (option=1, default) or secondary (option=2) click on a scenery tile. The verb depends on the Location def — "Chop" on a tree, "Mine" on a rock, "Climb-Up" on a ladder, etc. Target: a Loc from world.locs or a Position {x, y}. option: 1 (primary) or 2 (secondary); defaults to 1. Errors: `OUT_OF_RANGE` (too far to reach); `TARGET_OUT_OF_VIEW` (tile no longer visible); `PATH_BLOCKED` (can't reach interact distance). GUI: right-click scenery → command1 or command2 from the location's def. (exists)
- **`use(item: ItemRef, target: Target)`** → `Result<Null>`. Faculty: Action. Use one item on a target. Target kind picks the opcode: boundary (key on door), GroundItem/InvSlot (chisel on gem, paint on stone), Loc/scenery (log on fire), NPC (potion on NPC), etc. ItemRef as in `drop()`. Target: Npc, Player, Loc, GroundItem, Boundary, Position, or named position {x, y}. Errors: `NO_SUCH_ITEM` (item not in inventory); `OUT_OF_RANGE` (target too far); `TARGET_OUT_OF_VIEW` (target no longer visible); `SERVER_REJECTED` (server rejects the interaction). GUI: drag inventory item onto target; or right-click item then click target. (exists)
- **`open_boundary(boundary: Boundary)`** → `Result<Null>`. Faculty: Action. Default click on a boundary (door, gate, fence, web, etc.). Action depends on location def — "Open", "Close", "Climb-Up", "Cut-Through". Boundary view from world.locs/boundaries or position-derived. Errors: `OUT_OF_RANGE` (too far to reach); `TARGET_OUT_OF_VIEW` (boundary no longer visible); `DOOR_LOCKED` (door is locked; reason includes server prose like "this door requires a key" or "members only"). GUI: right-click door/boundary → first option from the def's command1. (exists)

#### Combat

- **`attack(target: Target)`** → `Result<Null>`. Faculty: Action. Initiate combat with an NPC or player. Target: Npc, Player, or Int(server_index). Errors: `OUT_OF_RANGE` (too far); `TARGET_OUT_OF_VIEW` (no longer visible); `TARGET_DEAD` (already dead). GUI: right-click NPC/player → "Attack", or left-click with attack weapon equipped. (exists)

#### Magic & Spells

- **`cast_on_self(spell_id: Int)`** → `Result<Null>`. Faculty: Action. Cast a non-targeted spell (heal, teleport, etc.). spell_id is the spellbook index (0–...). Errors: `NO_SUCH_ITEM` (spell not in spellbook); `SERVER_REJECTED` (insufficient runes or level). GUI: left-click spell in spellbook → cast on self. (exists)
- **`cast_on_npc(npc: NpcRef, spell_id: Int)`** → `Result<Null>`. Faculty: Action. Combat-cast on an NPC. NpcRef: Npc view or Int(server_index). Errors: `OUT_OF_RANGE`; `TARGET_OUT_OF_VIEW`; `NO_SUCH_ITEM` (spell not available); `SERVER_REJECTED` (insufficient runes/level/target restrictions). GUI: left-click spell in spellbook → click NPC. (exists)
- **`cast_on_player(player: PlayerRef, spell_id: Int)`** → `Result<Null>`. Faculty: Action. Cast on another player (PvP-only — server rejects outside legal zones). PlayerRef: Player view or String(name). Errors: `OUT_OF_RANGE`; `TARGET_OUT_OF_VIEW`; `SERVER_REJECTED` (not in PvP zone, spell banned, etc.). GUI: left-click spell → click player. (exists)
- **`cast_on_land(x: Int, y: Int, spell_id: Int)`** → `Result<Null>`. Faculty: Action. Tile-targeted cast (AOE spells like Earthquake, Blizzard). Errors: `OUT_OF_RANGE` (tile too far); `SERVER_REJECTED` (spell not allowed on that tile or insufficient runes/level). GUI: left-click AOE spell → click ground tile. (exists)
- **`cast_on_item(item: ItemRef, spell_id: Int)`** → `Result<Null>`. Faculty: Action. Cast on an inventory item (enchanting jewelry, alching, etc.). ItemRef as in `drop()`. Errors: `NO_SUCH_ITEM` (item not in inventory); `SERVER_REJECTED` (item not enchantable/alchable or insufficient runes/level). GUI: left-click spell (e.g., Alchemy) → click inventory item. (exists)

#### Prayer

- **`activate_prayer(prayer_index: Int)`** → `Result<Null>`. Faculty: Action. Activate a prayer slot (0–13, per the 14-prayer layout). Server silently rejects if prayer level too low or prayer points == 0; check result or poll `self.prayers.is_active(N)` after to confirm. Errors: `SERVER_REJECTED` (level requirement not met, no prayer points, etc.). GUI: click prayer icon in the prayer panel. (exists)
- **`deactivate_prayer(prayer_index: Int)`** → `Result<Null>`. Faculty: Action. Deactivate (toggle off) a prayer slot. Errors: `SERVER_REJECTED` (rarely; shouldn't fail). GUI: click active prayer icon to toggle off. (exists)

#### Social & Chat

- **`say(message: String)`** → `Result<Null>`. Faculty: Action. Public chat message (broadcast to all players in view). Message is the raw text (no length validation here; server enforces max 80 chars or similar). Errors: `SERVER_REJECTED` (message too long, contains illegal chars, etc.; rate-limited if spamming). GUI: type in chat box → press Enter. (exists)
- **`whisper(to: String, message: String)`** → `Result<Null>`. Faculty: Action. Private message to another player. `to` is the target player's username (String). Message: the text. Errors: `NO_SUCH_ITEM` (player not visible or doesn't exist; server-dependent); `SERVER_REJECTED` (recipient has PMs blocked, doesn't exist, or message violates rules). GUI: right-click player → "Message", or /msg command. (exists)
- **`add_friend(name: String)`** → `Result<Null>`. Faculty: Action. Add a player to the friend list (required before PMs can be sent or received on some servers). `name`: player username. Errors: `SERVER_REJECTED` (player doesn't exist, already a friend, friend list full, etc.). GUI: click "Add Friend" option after right-clicking a player. (exists)
- **`follow(target: PlayerRef)`** → `Result<Null>`. Faculty: Action. Server-side follow of a player (opcode 165). Target: Player view or String(name). Once called, self automatically pathfollows the target until the target logs out or follow is cancelled (cancellation not yet surfaced). Errors: `OUT_OF_RANGE` (target too far); `TARGET_OUT_OF_VIEW` (no longer visible); `NO_SUCH_ITEM` (player not found). GUI: right-click player → "Follow". (exists)

#### Session & Admin

- **`logout()`** → `Result<Null>`. Faculty: Action. Initiate logout (disconnects from server). Errors: `NOT_LOGGED_IN` (already logged out); `SERVER_REJECTED` (can't log out at this moment, e.g., in combat). GUI: click the Logout button or log out via menu. (exists)
- **`command(cmd: String)`** → `Result<Null>`. Faculty: Action. Send an admin command without the leading `::` (e.g., `command("tele 103 532")` sends `::tele 103 532`). Requires admin permissions on the account. Common commands: `tele <x> <y>`, `summon <name>`, `blink`, `invisible`. Errors: `SERVER_REJECTED` (not an admin, command syntax error, or command not available on this server). GUI: type `::command` in chat and press Enter. (exists)

#### Utilities & Helpers

- **`distance_to(target: View)`** → `Int` OR **`distance_to(x: Int, y: Int)`** → `Int`. Faculty: View (pure). Chebyshev distance from self.position to target (max of |dx|, |dy|), matching RSC's walk cost. Target: any view with .x/.y (Player, Npc, Loc, GroundItem, Position, Boundary), or named (x=X, y=Y). Returns Int (tiles; 0 if on same tile). GUI: none (used internally for proximity checks). (exists)
- **`distance_to_xy(x: Int, y: Int)`** → `Int`. Faculty: View (pure). Positional shorthand: `distance_to_xy(304, 542)` == `distance_to(x=304, y=542)`. Returns Int. GUI: none. (exists)
- **`in_region(x1: Int, y1: Int, x2: Int, y2: Int)`** → `Bool`. Faculty: View (pure). Returns true iff self.position is inside the axis-aligned rectangle (x1, y1)..(x2, y2) inclusive. Arg order normalized (min/max computed internally). Used by area-restricted routines. Returns Bool. GUI: none. (exists)
- **`find_option(needle: String)`** → `Int`. Faculty: View (pure). Returns the 1-based index of the first dialog option containing `needle` (case-insensitive substring), or 0 if none match. Used with `answer(find_option("Yes"))` after `wait_for_dialog()`. Returns Int. GUI: none. (exists)
- **`wait_for_dialog(timeout_seconds?: Int)`** → `Bool`. Faculty: Action (primitive-like, yields). Block until an NPC dialog menu opens, or until timeout_seconds elapses (default 5s). Polls every 200ms. Returns Bool: true if a menu landed, false on timeout. GUI: none (internal timing utility replacing brittle `wait N; if world.dialog.is_open` patterns). (exists)
- **`set_combat_style(style: String | Int)`** → `Result<Null>`. Faculty: Action. Change melee XP-split mode. Accepts "controlled" (even split), "aggressive" (all Strength), "accurate" (all Attack), "defensive" (all Defense), or Int 0–3 with the same mapping. Errors: `SERVER_REJECTED` (invalid style). GUI: click combat styles in the equipment panel. (exists)

#### Spatial Primitives (for `bounds` blocks)

- **`box(x1: Int, y1: Int, x2: Int, y2: Int)`** → `Shape`. Faculty: Primitive (returns a Shape object for use inside `bounds <shape> { ... }`). Axis-aligned rectangle; inclusive edges; arg order normalized (min/max computed internally). Used in bounding spatial reactors (e.g., `bounds box(120, 640, 130, 660) { on item_appeared(...) { ... } }`). Returns Shape. GUI: none. (exists)
- **`circle(cx: Int, cy: Int, radius: Int)`** → `Shape`. Faculty: Primitive. Chebyshev-radius disk (RSC uses 8-way movement; circles look square). Used inside `bounds`. Returns Shape. GUI: none. (exists)
- **`near(radius?: Int)`** → `Shape`. Faculty: Primitive. Disk centered on self.position at routine start (radius default depends on context). Used inside `bounds near(8) { ... }` for "react to events within 8 tiles of where I logged in". Returns Shape. GUI: none. (exists)

#### Ambient verbs — notes

- **Top-level:** all of these verbs are truly ambient — subsystem-free, player-speech-idiom, owned by no root (not self.*, not world.*, not inventory.*, etc.).
- **Canonical names:** per the frozen contract (§10), ambient verbs "keep their names — already faculty-shaped and subsystem-free." `pickpocket` is canonical (frozen); `npc_command` is a legacy alias.
- **Faculty-shaped:** Actions return Result<T> (most T=Null; pick_up returns GroundItem); Primitives return scalar/Shape directly with no Result wrapper.
- **Capability boundary:** Views (is_reachable, distance_to, distance_to_xy, in_region, find_option, wait_for_dialog) expose only what a human sees on screen; Actions map to client UI clicks and keypresses; Primitives (box, circle, near) are purely local utilities.

### `control` — The control plane (non-GUI primitives, outside §3 by design)

GUI-equivalence (§3) governs the **body** (game I/O). A routine also has its own faculties — flow, logging, thinking, test-setup — that no GUI player has. These are a deliberately **separate, fenced layer**, never confused with body verbs. They are documented here for completeness but are not subject to the capability boundary.

#### Flow / timing (language constructs)

These are grammar-level constructs (parsed in `dsl/parser/*`, validated in `dsl/validator/*`), not entries in `dsl/spec/actions.go`:

- `if / elif / else`, `while`, `for`, `return <v>`, `abort <msg>`.
- `when <cond> { … }` — fire a block when a condition becomes true.
- `select { on <event> {…} timeout <Ns> {…} }` — wait on one of several events with a timeout.
- `repeat { … } until <cond> timeout <Float sec>` — bounded loop.
- `defer { … }` — run on scope exit.
- `try { … } recover err { … }` — catch a bang-propagated error.

#### Timing primitives (Primitive faculty)

- **`wait(seconds: Float)`** → `Null`. Faculty: Primitive. Sleep for the given duration. GUI: none. (exists)
- **`wait_until(pred: fn -> Bool, timeout_seconds?: Float)`** → `Bool`. Faculty: Primitive. Block until the predicate returns true or the timeout elapses; returns true iff the predicate fired. GUI: none. (exists)

#### Introspection (Primitive faculty)

- **`note(msg: String)`** → `Null`. Faculty: Primitive. Emit a line to the host log (observability only; not seen in-world). GUI: none. (exists)

#### Cognition bridge (mind-access — NOT GUI-equivalent, fenced)

These hit the knowledge/brain/memory layer, not the game. Several are implemented as deterministic stubs pending Phase 3–4 LLM + memory integration.

- **`recall(query: String, top?: Int)`** → `List<Chunk>`. Faculty: MemoryStdlib. Vector/knowledge-corpus search. Tag: (exists — canned stub in Phase 3).
- **`relation_with(name: String)`** → relational record. Faculty: MemoryStdlib. Look up the host's relationship with another entity. Tag: (exists — canned stub).
- **`contemplate_reality(question?: String)`** → decision. Faculty: LLMStdlib. Open-ended LLM decision. Tag: (exists — canned in Phase 3).
- **`evaluate(situation)`** → `Float`. Faculty: LLMStdlib. 0–1 confidence assessment. Tag: (exists — canned stub).
- **`decide(options, context?)`** → option. Faculty: LLMStdlib. Pick one option. Tag: (exists — canned stub).
- **`resolve(text: String, kind?: String)`** → `List<Match{ def, kind, score }>`. Faculty: recognition / fuzzy resolution (mind-access, learnable). Ranked candidates, best first. Pipeline: the host's learned-alias store → conservative fuzzy/token match vs canonical names → brain (LLM) fallback (which on success writes the resolved alias back). Definitions are looked up exactly (facts); only the name→canonical recognition is fuzzy, and ids are never invented by the LLM. Tag: (to build — not in `dsl/spec/actions.go` or actionHandlers; requires host alias store design first).
- **`ask_brain(prompt: String)`** → `String`. Faculty: LLMStdlib. Open-ended brain-access decision (like `contemplate_reality` but explicitly brain-access). Tag: (to build — not in spec; routines use `contemplate_reality` instead for now).
- **`exec(prompt: String)`** → `Result<Null>`. Faculty: LLMStdlib. LLM fragment executor. Tag: (to build).
- **`improvise(prompt: String)`** → `Result<Null>`. Faculty: LLMStdlib. One-shot LLM fragment. Tag: (to build).
- **`reflect_now()`** → `Result<Null>`. Faculty: LLMStdlib. Reflection pass. Tag: (to build).
- **`wait_for_chat(timeout?: Float, from?: String)`** → `Result<…>`. Faculty: MemoryStdlib. Event waiter for incoming chat. Tag: (to build).
- **`observe(target, ticks: Int)`** → `Result<…>`. Faculty: MemoryStdlib. Entity watcher. Tag: (to build).

#### Persona reads (identity state, no bang)

- **`mood()`** → emotional state. Faculty: PersonaRead. Tag: (to build).
- **`motivation()`** → goal state. Faculty: PersonaRead. Tag: (to build).

#### Admin / test (fenced, test-harness only)

- **`command(cmd: String)`** → `Result<Null>`. Faculty: PrimaryAction (test-harness only). Admin commands (`setstat`, `item`, `teleport`, etc.). Never appears in persona-authored behavior. Tag: (exists). (Also listed under ambient verbs as the player-facing surface; here noted as the fenced admin escape hatch.)

#### `control` — notes

- This entire plane is **explicitly outside the GUI-equivalence rule (§3)** so the body surface stays honest and freezable.
- Flow constructs are grammar, not callables — they have no `dsl/spec/actions.go` entry; they live in the parser/validator.
- The cognition/persona stubs return deterministic values today; real implementations land in Phase 3 (memory mesa) and Phase 4 (real brain).

## 9. Sanctioned aliases (the only second-names)

- `attack(t)` → `combat.attack(t)`
- `cast(spell, target?)` → `magic.cast(...)`

## 10. Migration map (flat → frozen)

**Trade** — note the two distinct screens, and that `request` absorbs both old
request/accept-request verbs (same packet):

| today (flat) | frozen | why |
| --- | --- | --- |
| `trade_request(p)` **and** `accept_trade(p)` | `trade.request(p)` | Both send opcode 142. Clicking a player to trade *is* both initiating and accepting an incoming request — one player action, one verb. |
| `offer_trade([...])` | `trade.offer([...])` | — |
| `confirm_trade()` | `trade.accept()` | This was the **offer-screen** accept (screen 1). The old name "confirm" was the very confusion we're removing. |
| `finalize_trade()` | `trade.confirm()` | The **confirm-screen** accept (screen 2). |
| `decline_trade()` | `trade.decline()` | — |
| `world.trade.{is_active,phase,my_offer,their_offer}` | `trade.{…}` | promoted onto the `trade` root |
| `world.trade.{my,their,both}_first_accepted` | `trade.{accepted,they_accepted,both_accepted}` | offer screen |
| `world.trade.{my,their,both}_second_accepted` | `trade.{confirmed,they_confirmed,both_confirmed}` | confirm screen |

**Duel** — currently `confirm_duel()` is a single toggle for both screens (the
same bug we already split out of trade). The freeze splits it to match trade:

| today (flat) | frozen | why |
| --- | --- | --- |
| `duel_request(p)` **and** `accept_duel(p)` | `duel.request(p)` | same packet, same reasoning as trade |
| `set_duel_rules(...)` | `duel.set_rules(...)` | — |
| (stake offer) | `duel.stake([[id,qty]])` | — |
| `accept_duel_offer()` | `duel.accept()` | offer screen (screen 1) |
| `accept_duel_confirm()` | `duel.confirm()` | confirm screen (screen 2) |
| `decline_duel()` | `duel.decline()` | — |

**Other:**

| today | frozen |
| --- | --- |
| `open_bank(b)` / `deposit` / `withdraw` | `bank.open(b)` / `bank.deposit(...)` / `bank.withdraw(...)` (+ new `bank.deposit_all`, `bank.close`) |
| `cast_on_self/npc/item(spell, t)` | `magic.cast(spell, t?)` |
| `activate_prayer(name)` | `prayer.activate(name)` |
| `attack(t)` | `combat.attack(t)` (alias `attack` kept) |
| `npc_command(npc)` / `pickpocket(npc)` | `pickpocket(npc)` (one canonical) |
| `world.last_server_message` | `world.messages: List<Message>` + `on message(pattern)` |

Ambient verbs (`walk_to`, `use`, `interact_at`, `examine`, `eat`, `drop`,
`bury`, `equip`, `pick_up`, `say`, `whisper`, `wait`, `wait_until`, `note`,
`talk_to`) keep their names — already faculty-shaped and subsystem-free.

## 11. Explicitly out of scope (do NOT add)

- Behavioral policy (auto-eat, retreat rules, budgets, autocast, prayer-flick) →
  mind layer.
- Anything a GUI player can't do or see (server-secret state, raw NPC hp,
  teleport without a spell).
- Per-variant enumerations where a parameter suffices.
