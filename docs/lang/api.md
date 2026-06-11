# The Routine API — frozen surface (v1, ratified 2026-05-29)

> **STATUS — re-verified by hand 2026-06-10 against HEAD `0bfa818` (branch
> `tidy/structure-and-docs`).** The contract (§1–§7, §9–§11) is **FROZEN (v1)**,
> ratified 2026-05-29, and unchanged. The body build-out that the freeze
> anticipated (tasks #115–120) has since **LANDED**: the §10 migration is
> executed — the namespaced verbs (`trade.request`, `bank.deposit`,
> `duel.stake`, `magic.cast`, `prayer.activate`, `combat.attack`, `shop.buy`)
> are live, view-dispatched callables (`runtime/namespace_actions.go` + the
> per-namespace verb tables in `runtime/dsl_actions.go`), and the old flat
> names are gone from `dsl/spec/actions.go` except the two §9 aliases.
> Every §8 status tag below was re-checked against `dsl/spec/*` +
> `runtime/views_*.go` + `runtime/actions_*.go` on the verify date.
>
> **Runtime: 1.0** — this contract is versioned with the Routine Runtime
> ([`versioning.md`](versioning.md) · [`CHANGELOG.md`](CHANGELOG.md)). A
> `.routine` targets this surface by declaring `runtime "1.0"`.
>
> **Two parts, like Godoc:** this hand-written doc is the *contract* — the
> model, the type vocabulary, the namespacing rules, the capability boundary,
> and one fully-worked reference section (`trade`) that fixes the depth bar.
> The *exhaustive* per-entry reference (§8) was DESIGNED to be **generated
> from `dsl/spec/*`** by a `go run ./cmd/specdoc` tool so it could never
> drift. **That generator was never built** — §8 is hand-maintained, and it
> drifted exactly as this header predicted (a dozen `(to build)` tags went
> stale within days of the body build-out). The 2026-06-10 pass re-verified
> §8 by hand; the durable fix is still the generator — **`DSL-17` in
> [`/docs/TODO.md`](../TODO.md)**, the single source of truth for open work.

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

> This is the depth every namespace gets. `bank`, `duel`, `shop`, `combat`,
> `magic`, `prayer`, `self`, `world`, `inventory` follow the same form in §8 —
> hand-maintained (re-verified 2026-06-10) pending the `cmd/specdoc` generator
> (TODO.md `DSL-17`).

## 8. The namespaces (full reference)

### `self` — Views

#### `self.position` → `Position`
Faculty: View. Current tile coordinates. Returns `{x: Int, y: Int, plane: Int}` — `plane` is the floor index derived from the wire Y (RSC stacks floors in Y-space at `world.PlaneHeight` intervals; `world.PlaneOf`). Every `.position` view (self, players, npcs, ground items, placements, boundaries) carries the same three fields (`positionView` in `runtime/views_self.go`). Never null. GUI: player sprite location on minimap and main view. (exists)

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
Faculty: View. Fatigue as a normalized percentage 0–100 (the wire value is 0–750; `Self.FatiguePercent()` scales it — the manual teaches `if self.fatigue > 80`). At 100 the host must sleep before gaining more xp. GUI: the fatigue bar. (exists)

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
Faculty: View. True while engaged with an NPC or player. **De-stubbed** — delegates to the combat faculty's engaged check (`combatView.engaged()` in `runtime/views_combat.go`): true when we have a resolvable combat target (wire-observed engagement, or our last-attacked NPC/player still alive and visible) or another player is firing at us. Clears the moment the engaged NPC is pruned on death. GUI: the combat overlay on screen. (exists)

#### `self.is_sleeping` → `Bool`
Faculty: View. True if the sleep screen is showing. **De-stubbed** — reads the world `SleepState` mirror set on `SEND_SLEEPSCREEN` (true) / `SEND_STOPSLEEP` (false). GUI: the sleep dialogue. (exists)

#### `self.skills` → `SkillsView`
Faculty: View. Access per-skill data. Supports the 18 RSC skills (attack, defense, strength, hits — alias hitpoints, ranged, prayer, magic, cooking, woodcutting, fletching, fishing, firemaking, crafting, smithing, mining, herblaw, agility, thieving; `skillIDs` in `runtime/views_self.go`). Each skill resolves to a `Skill` instance. GUI: the skills menu. (exists)

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

#### `self.equipped` → equipped-view (list surface + per-slot surface)
Faculty: View. Serves two surfaces on one value:

- **List surface** — `self.equipped` iterates / indexes the currently-wielded inventory items (`.all`, `.length`, `.first`, `.last`, `[N]`, and iteration). Each element is an `InvSlot` carrying `.id`, `.amount`, `.name` from facts. (`.filter`/`.map`/`.find` are list-method callables; route them through `self.equipped.all`, e.g. `self.equipped.all.filter(i => i.is_wearable)`.)
- **Per-slot surface** — `self.equipped.<slot>` resolves to the item actually worn in that slot. Slots: `weapon`, `shield`, `helmet` (alias `hat`), `head`, `body`, `legs`, `gloves`, `boots`, `amulet`, `cape`. Resolved from the host's own inventory (the Wielded slots, matched to a slot by the item def's WearSlot) — so they're **exact**. Each slot returns a worn-item view with `.name`, `.id`, `.def` (full ItemDef from facts), `.slot_name`, `.is_empty`.

Note: `head` / `body` / `legs` each span **two** underlying RSC body-animation layers (the metal/style picks which — e.g. plate-mail legs vs. a skirt, a medium vs. a large helm); the accessor returns whichever is worn. The exact-layer names are also exposed: `large_helmet`, `med_helmet`, `platebody`, `chainbody`, `platelegs`, `skirt`. (exists)

#### `self.equipped.bonuses` → equipment-bonuses-view
Faculty: View. The summed combat bonus of all worn gear — the "equipment status" the in-game equipment screen shows. Recomputed on demand from the currently-wielded items, so it always reflects what's equipped right now. Fields: `.armour`, `.weapon_aim` (alias `.aim`), `.weapon_power` (alias `.power`), `.magic`, `.prayer` (all `Int`). GUI: the equipped panel (character equipment screen) and its bonus totals. (exists)

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
- `Position` = `{x: Int, y: Int, plane: Int}` (plane derived from the wire Y; see `self.position` above)
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

- **`inventory.find_all(item: ItemRef)`** → `List<InvSlot>`. Faculty: View. Returns all slots matching the item as InvSlot instances (one per slot, even for stackables — so a stackable in one slot appears once, unstackables in multiple slots appear multiple times). Empty list if no match. `ItemRef` resolution same as `has`. (exists — `invFindAllCallable` in `runtime/views_inventory.go`)

- **`inventory.find_any(items: List<ItemRef> | ItemRef...)`** → `InvSlot?`. Faculty: View. Returns the first (lowest-indexed) inventory slot matching **any** of the supplied item ids/names, as an InvSlot instance, or `Null` if none match. Accepts a single list (`inventory.find_any([373, 372])`) or varargs (`inventory.find_any(373, "Lobster")`). "First matching" is by slot order, not argument order. Item refs that don't resolve are skipped (best-effort union); an empty arg set errors. Collapses gem/food/axe or-chains into one call. `ItemRef` resolution same as `has`. (exists; #117)

- **`inventory.slot_of(item: ItemRef)`** → `Int?`. Faculty: View. Returns the 0-based slot index of the first matching slot, or `Null` if not found. Range: 0–29. `ItemRef` resolution same as `has`. GUI: internal indexing used by the client; routines may need it for equipment-slot reasoning. (exists)

### `world` — Visible entities and messages (Views only)

The `world` namespace exposes the current state of entities visible to the player: NPCs, other players, ground items, static locations, and boundaries. All members are read-only Views with no Action verbs.

#### `world` — Views

##### Lists and entity views

- **`world.players`** → `List<Player>`. Faculty: View. All players visible within the view radius. Each element is a Player instance with `.index`, `.name`, `.position`, `.combat_level` (from the appearance packet; null until seen), `.relative_level` / `.threat` / `.threat_colour` (danger relative to the host), `.is_skulled`, `.equipment` (per-player worn-equipment view) plus direct per-slot accessors, `.hp_fraction` / `.health` (PVP target only), `.is_friend` (stub: always false), `.in_combat_with` (stub: always null). Returns empty list when alone. (exists)

- **`world.npcs`** → `List<Npc>`. Faculty: View. All NPCs visible within the view radius. Each element is an Npc instance with `.index`, `.type_id`, `.position`, `.name` (from facts, or "" if def not loaded), `.combat_level` (computed: (atk+str+def)/4 + hits/4 from def), `.relative_level` / `.threat` / `.threat_colour` (danger relative to the host), `.max_hp` (hits field from def), `.is_attackable` (from def), `.is_aggressive` (from def), `.hp_fraction` / `.health` (the engaged combat target only; null otherwise). Returns empty list when no NPCs nearby. (exists)

- **`world.ground_items`** → Iterable<GroundItem>. Faculty: View. All ground items visible within the view radius. Also callable as a list (e.g., `for gi in world.ground_items { ... }`). Each element carries `.item_id`/`.id`, `.position`, `.name`, `.is_mine` (stub: always false, to build). Supports `.by_id(item_id, radius?: Int)` → GroundItem | Null (nearest matching item, optionally within radius), `.nearest` → GroundItem | Null (closest ground item to `self.position` by Chebyshev distance) which is **also callable** as `.nearest(pos)` → GroundItem | Null to recenter the search on an explicit position (a position-like value with `.x`/`.y`, or two Int `x, y` args; #117), `.most_valuable` → GroundItem | Null (the visible item with the highest `facts.ItemDef` base value — enables loot-most-valuable; #117), `.all` → List<GroundItem>, `.length` → Int (count of visible items). (exists)

- **`world.scenery`** → Iterable<Placement>. Faculty: View. DYNAMIC scenery (GameObject) state streamed by the server via `SEND_SCENERY_HANDLER` (opcode 48) — the live, in-view scenery including runtime-spawned objects that the static `world.locs` data does NOT contain (a fire you just lit (def 97), a depleted/regrown rock, etc.). Each element is a Placement (kind="scenery", with `.x`/`.y`/`.position`/`.name`/`.def_id`), so results drop straight into `use(item, scenery)` → USE_ITEM_ON_SCENERY (opcode 115). Supports `.by_id(def_id, radius?: Int)` → Placement | Null (nearest matching scenery to `self.position`, optionally within radius; use `radius=0` for the exact tile), `.nearest` → Placement | Null (closest scenery to `self.position`), `.all` → List<Placement>, `.length` → Int. Use this to find a fire you just lit: `world.scenery.by_id(97, radius=0)` (NOT `world.ground_items.by_id(97)` — a fire is a GameObject, not a ground item — and NOT `world.locs.search("fire")` — that only knows pre-placed fires). (exists)

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

- **`world.messages`** → List<Message>. Faculty: View. The bounded server-message log, **oldest-first** (last N entries; N is `world.ServerMsgRingCap`). Not a single-value buffer — unlike the `last_*` accessors above, this accumulates a history. Each Message carries `.text` (String, also aliased `.message`), `.kind` (String, "server"), `.at` (Time), and the method `.contains(needle)` → Bool. The list itself supports `.length`, `.first`, `.last`, iteration, etc. Fed by the same server-message source as `on message` / `on server_message`. (#119, exists)

- **`world.last_dialog_text`** → DialogTextRecord | Null. Faculty: View. Most recent NPC speech-bubble text (not the same as a dialog menu option). Fields: `.text` (String), `.at` (Time). (exists)

##### Dialog menu state

- **`world.dialog`** → View. Faculty: View. Current NPC dialog menu state (or Null-safe empty when no menu open). Supports:
  - `.is_open` → Bool. True iff a dialog menu is currently presented.
  - `.options` → List<String>. The menu option texts (empty list if none / menu not open).
  - `.find_option(substring: String)` → Int. **1-based** index of the first option whose text contains (case-insensitive) the given substring, or **0** if no match — matches the top-level `find_option` + `answer()`, so `answer(world.dialog.find_option("Yes"))` is correct. Substring matching is forgiving — quest text like "Yes, I'd like to help." can be matched with just "Yes".
  - `.clear()` → Null. Reset the cached options after resolving a menu (the server doesn't reliably signal menu close; callers must signal intent).
  (exists)

##### Trade/duel state (from flat world.trade / world.duel)

- **`world.trade`** → View. Faculty: View. Trade window state and offer screens. (Note: top-level `trade.*` is the frozen namespace per §6; these are still accessible via `world.trade.*` for backwards compatibility, but new code prefers `trade.*`.) Supports: `.is_active`, `.phase` (values: "request_sent", "open", "confirm", "completed", "cancelled", "none"), `.my_offer`, `.their_offer`, plus `.with` (opponent name), `.with_index` (opponent index), and the accept flags under BOTH namings — the frozen §7 names (`.accepted`, `.they_accepted`, `.both_accepted`, `.confirmed`, `.they_confirmed`, `.both_confirmed`) and the legacy `my/their/both_first_accepted` / `_second_accepted` aliases (`runtime/views_trade.go` serves each pair from one case). (exists)

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
- `.position` → Position. Shorthand for `{x, y, plane}` (shared `positionView`; see `self.position`). (exists)
- `.combat_level` → Int | Null. From the appearance packet's combat-level bytes; null until the appearance has been seen. (exists)
- `.relative_level` → Int | Null. The player's combat level minus the host's (positive = they out-level us). Null until the appearance is seen. (exists)
- `.threat` → String | Null. How dangerous this player is *relative to the host* — see the threat-band note below. Null until the appearance is seen. (exists)
- `.threat_colour` → String | Null. The `@col@` tag the client would paint over the level for that threat. Null until the appearance is seen. (exists)
- `.is_skulled` → Bool. True if the player shows a wilderness skull (`SkullType != 0`). (exists)
- `.equipment` → equipment-view. The player's worn equipment, resolved from their appearance packet. Per-slot accessors `.helmet` / `.weapon` / `.shield` / `.body` / `.legs` / `.gloves` / `.boots` / `.amulet` / `.cape` / `.head` (also exposed directly on the player instance) plus `.all` (List) and `.length`. Each slot is a worn-item view: `.name`, `.id`, `.def`, `.is_empty`, `.ambiguous`, `.candidates`. Resolved from the appearance value (item AppearanceID & 0xFF). Same-metal melee weapons share a worn appearance, so a weapon slot can be `.ambiguous` (`.id` null, multiple `.candidates`) — faithful to what's visually distinguishable; helmets/armour resolve exactly. (exists)
- `.hp_fraction` → Float | Null. Current/max hitpoints ratio of a PVP opponent, from the appearance-damage update; null for an unfought player. (exists)
- `.health` → Int | Null. Current hitpoints of a PVP opponent; null for an unfought player. (exists)
- `.is_friend` → Bool. Stub: always false until friend-list tracking lands. (to build — perception gap)
- `.in_combat_with` → Player | Null. Stub: always null until per-player combat target tracking lands. (to build — perception gap)

##### Npc instance `.` fields

- `.index` → Int. Server-side NPC index. (exists)
- `.type_id` → Int. NPC type ID (key into facts.NpcDefs). (exists)
- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y, plane}` (shared `positionView`; see `self.position`). (exists)
- `.name` → String. Loaded from facts NpcDef; returns "" if def not yet loaded. (exists)
- `.combat_level` → Int | Null. Computed from def: (attack + strength + defense) / 4 + hits / 4. Null if def not loaded. (exists)
- `.relative_level` → Int | Null. This NPC's combat level minus the host's (positive = it out-levels us). Null if def not loaded. (exists)
- `.threat` → String | Null. How dangerous this NPC is *relative to the host* — see the threat-band note below. Null if def not loaded. (exists)
- `.threat_colour` → String | Null. The `@col@` tag the client would paint over the level for that threat. Null if def not loaded. (exists)
- `.max_hp` → Int | Null. The "hits" field from the NPC def; null if def not loaded. (exists)
- `.is_attackable` → Bool. From NPC def; false if def not loaded. (exists)
- `.is_aggressive` → Bool. From NPC def; false if def not loaded. (exists)
- `.hp_fraction` → Float | Null. Current/max hitpoints ratio of the engaged combat target, from the opcode-104 health update; null for an NPC we've never fought (only the engaged target's health bar is on the wire). (exists)
- `.health` → Int | Null. Current hitpoints of the engaged combat target; null otherwise (same gate as `.hp_fraction`). (exists)
- `.in_combat_with` → Player | Npc | Null. Stub: always null. Would track which entity this NPC is fighting. (to build — perception gap)

**Threat-band note** (`.threat` / `.threat_colour` on both Player and Npc instances): these express how dangerous an entity is *relative to the host* — the authentic RSC cue the client paints as the level number's colour (server `Formulae.getLvlDiffColour`; "darker red = more dangerous"). A host can't read a UI colour, so we surface the concept. Bands from most to least dangerous: `deadly` (`@red@`), `very dangerous` (`@or3@`), `dangerous` (`@or2@`), `risky` (`@or1@`), `even` (`@whi@`), `favourable` (`@gr1@`), `easy` (`@gr2@`), `very easy` (`@gr3@`), `trivial` (`@gre@`).

##### GroundItem instance `.` fields

CORRECTION: GroundItem does NOT carry a `.quantity` field. The underlying GroundItemRecord contains only {X, Y, ItemID, LastSeen} and no quantity tracking.

- `.item_id` / `.id` → Int. Item ID. (exists)
- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y, plane}` (shared `positionView`; see `self.position`). (exists)
- `.name` → String. Item name from facts. (exists)
- `.is_mine` → Bool. Stub: always false until the host tracks 3-minute loot-ownership windows. Routines picking up loot should check this, but currently fall back to "try and see" with `pick_up()`'s `SERVER_REJECTED` error. (to build — perception gap)

##### Placement instance `.` fields (from locs queries)

- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y, plane}` (shared `positionView`; see `self.position`). (exists)
- `.name` → String. Scenery/boundary/NPC def name. (exists)
- `.kind` → String. "scenery" | "npc_spawn" | "boundary". (exists)
- `.def_id` → Int. Facts registry ID. (exists)
- `.direction` → Int. For boundaries, the direction; 0 for scenery. (exists)

##### Boundary instance `.` fields (from world.boundaries.at)

- `.x`, `.y` → Int. Tile coordinates. (exists)
- `.position` → Position. Shorthand for `{x, y, plane}` (shared `positionView`; see `self.position`). (exists)
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

- **`on message(text: String)`** — A new server (system) message arrived. Fires alongside `on server_message` and is fed by the same source, but is backed by the bounded `world.messages` ring (last N entries, oldest-first). Filter in-body with `text.contains(needle)`, or read the accumulated log via `world.messages`. (#119, exists)

- **`on xp_gain(skill: String, amount: Int)`** — A skill's experience increased. `skill` is the lowercase skill name (`"attack"`, `"fishing"`, …); `amount` is the positive xp delta (not the new total). Filter on the skill name in-body (`if skill == "fishing" { ... }`). Synthesized by diffing the per-skill xp mirror across the server's stat/xp packets, so it fires once per real gain regardless of which packet carried it. Replaces any per-skill event variants. (#119, exists)

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

2. **Trade/Duel accept field naming — RESOLVED.** The discrepancy this item recorded (runtime served only `my/their/both_first_accepted` / `_second_accepted`, not the §7 exemplar names) was closed by the §10 rename: `runtime/views_trade.go` + `runtime/views_duel.go` now serve the frozen names (`.accepted` / `.they_accepted` / `.both_accepted` / `.confirmed` / `.they_confirmed` / `.both_confirmed`) as canonical, with the first/second names kept as back-compat aliases on the same switch cases.

3. **Bank fields expanded**: The provided schema referenced `world.bank.is_open` and `.slots` generically. The actual implementation adds `.max_size`, `.used`, `.free` fields and the methods `.has(item_id)` / `.count(item_id)` (aliases).

4. **world.trade.phase values**: The provided schema (§7) lists `{offer, confirm}` as valid phase values. The actual implementation uses `{request_sent, open, confirm, completed, cancelled, none}` (more granular).

5. **Duel-related events**: The provided schema did not document duel events (duel_opened, duel_other_offer, etc.); they exist in runtime/dsl_events.go and are documented in the events list above.

#### Tag notes

- **(exists)** — Implemented and tested in `runtime/views_*.go` (per-namespace view files) + `runtime/dsl_events.go`.
- **(to build — perception gap)** — Specified here as the *target* the API should reach, but a perception-layer feature must land first. Currently stubbed (always null/false/safe default).

Newer events not catalogued above also exist in `dsl/spec/events.go` (`item_gained`, `damage_taken`, `level_up`, `equipment_changed`, `player_equipment_changed`, `player_level_changed`); the per-event reference lives in [`events.md`](events.md).

### `combat` — Views

- **`combat.target`** → `Npc|Player|Null`. Faculty: View. Current engagement target for the player. Resolution order (`combatView.target()` in `runtime/views_combat.go`): (1) the wire-observed engagement on our own player record (`EngagedNpcIndex` / `EngagedPlayerIndex`, landed by the opcode-104/234 decoders — gated on `EngagedAt` so a never-engaged record doesn't read as "engaging index 0"); (2) melee fallback to our last-attack intent (`host.lastAttackedNpcIndex` / `lastAttackedPlayerIndex`) while that entity is still in view — authentic v235 melee carries no projectile, so intent is the only client-side signal. Returns `Null` when nothing is engaged, or once the target dies (the `on target_died` / `on npc_killed` edge fires) or leaves view. Resolves both NPC and player targets. GUI: the name/health-bar overlay on screen (NPC combat or PVP duel). Tag: **(exists)**.

- **`combat.engaged`** → `Bool`. Faculty: View. True while the player is in active combat: we have a resolvable target (per `combat.target` above), or another player is firing projectiles at us (their `EngagedPlayerIndex` == our index). False when idle, traveling, or between targets. Nullability: never null. GUI: the combat overlay. Tag: **(exists)** — `self.is_in_combat` delegates to this same check.

- **`combat.style`** → `String`. Faculty: View. Read-side mirror of the most-recently-applied melee xp-split (#117) — write-through from `combat.set_style`. RSC sends no confirmation echo, so this reflects *intent*; defaults to the server start state `"controlled"` before any `set_style`. Tag: **(exists)**.

- **`combat.last_npc`** → `Npc|Null`. Faculty: View. The most recently attacked NPC, resolved live from `world.npcs` by stored index. Null if no NPC has been attacked this session, or if the previous target has since left view or died. Routines use this to detect target loss and retarget or flee. GUI: the name/health-bar that *was* on screen. Nullability: `Null` on target loss or never attacked. Tag: **(exists)**. Stored in `host.lastAttackedNpcIndex` (set on `attack()` dispatch, stays set across entity leaving view).

- **`combat.last_player`** → `Player|Null`. Faculty: View. The most recently attacked player, resolved live from the player roster. Null if no player has been attacked, or if the previous PVP opponent has logged out / left the area. Common in duels to track engagement for post-combat loot or honor tracking. GUI: the opponent's name/health-bar in a duel overlay. Nullability: `Null` on opponent loss. Tag: **(exists)**. Stored in `host.lastAttackedPlayerIndex` (set on `attack()` dispatch).

### `combat` — Actions (all → `Result<Null>`)

- **`combat.attack(target: Npc|Player|Int)`** → `Result<Null>`. Faculty: Action. Initiates combat with the target (NPC, player, etc.). Accepts an `Npc` view, `Player` view, or server index `Int`. Pathfinds from `self.position` to an adjacent tile, then sends the attack opcode (190 for NPC, 171 for player). Stores the target index in `host.lastAttackedNpcIndex` or `host.lastAttackedPlayerIndex` for `combat.last_npc`/`combat.last_player` queries. Errors: `OUT_OF_RANGE` if the target is too far to reach; `PATH_BLOCKED` if pathfinding fails; `SERVER_REJECTED` for other server rejections (target dead, not attackable, etc.). GUI: clicking an NPC or player's sprite to attack. Tag: **(exists)**. Note: spec name in `/dsl/spec/actions.go` is `"attack"`, and it functions as an alias for `combat.attack()`; within the `combat` namespace, it is promoted to `combat.attack()`.

- **`combat.set_style(style: String|Int)`** → `Result<Null>`. Faculty: Action. Changes the melee xp-split mode. Accepts a string (`"controlled"`, `"aggressive"`, `"accurate"`, `"defensive"`) or an int (0–3 with the same mapping: 0=controlled, 1=aggressive, 2=accurate, 3=defensive). Takes effect on the next attack tick; RSC does not send a confirmation packet, so read `combat.style` (intent mirror) or observe `self.skills.<style>.xp` deltas to confirm application. Errors: none on the DSL side (the server silently applies the change or returns `SERVER_REJECTED`). GUI: clicking the combat-style toggle button to cycle modes. Tag: **(exists)** — `combat.set_style` is the only name; the old flat `set_combat_style` was removed from `dsl/spec/actions.go` in the §10 migration.

- **`combat.retreat(wait_rounds?: Bool)`** → `Result<Null>`. Faculty: Action. Breaks melee by walking one tile away (#117). RSC forbids fleeing until the opponent has landed 3 hits ("the first 3 rounds of combat"); `wait_rounds` (default true) waits out that anti-kite window before moving. Errors: `RETREAT_TOO_EARLY` (typed code in `dsl/interp/error.go`) if attempted inside the window with `wait_rounds=false`. GUI: walking away mid-fight. Tag: **(exists)** — `dslRetreat` in `runtime/actions_combat.go`.

- **`combat.retreat_to(x: Int, y: Int, wait_rounds?: Bool)`** → `Result<Null>`. Faculty: Action. Same break-combat semantics, then flees to a specific safe tile once the retreat is allowed. Tag: **(exists)** — `dslRetreatTo`.

### `combat` — Events

- **`on combat_started(target: Npc|Player)`** — Combat engaged; `target` is the entity being attacked. (to build — not in `dsl/spec/events.go`; poll `combat.engaged` / `combat.target` instead.)

- **`on combat_ended()`** — Player disengaged from combat (target died, player fled, or server ended the bout). (to build — not in `dsl/spec/events.go`.)

- **`on target_changed(from: Npc|Player|Null, to: Npc|Player|Null)`** — Explicit combat retarget during an active bout. (to build — not in `dsl/spec/events.go`.)

- **`on target_died(target: Npc)`** — The engaged NPC combat target (`combat.target` / `combat.last_npc`) died — its opcode-104 current-hitpoints reading transitioned from >0 to 0. `target` is the dead Npc view (or `Null` if it already left view). Fires once per kill via the alive→dead edge detector in the host (no double-fire on repeated 0-hits packets). `combat.target` clears to `Null` after the kill. (#119, exists)

- **`on npc_killed(target: Npc)`** — Alias of `on target_died`: fires on the same death edge, with the same `target` arg. Provided so routines can read either phrasing; both handlers fire if both are present. (#119, exists)

---

#### Notes on the Combat Model

**Perception vs. Server State:**
- The full combat perception surface is **live**: `combat.target` and `combat.engaged` read the wire-observed engagement (opcode-104/234 decoders) with a melee intent fallback; `combat.style` is the write-through intent mirror; `last_npc`/`last_player` track last-attack intent across the target leaving view. (An earlier revision of this section called target/engaged "stubs pending task #8" while the entries above said (exists) — that contradiction is resolved: they are real, see `runtime/views_combat.go`.)
- Authentic v235 melee has **no projectile signal** — a pure defender (attacked first, never retaliated) registers via the "someone is firing at me" check only for ranged/magic; for melee, take damage events (`on damage_taken`) as the cue.

**Capability Boundary:**
- All Views are GUI-equivalent: a player sees the target bar, the style toggle, and the names of recent opponents.
- `attack()` and `set_style()` mirror direct client actions; `retreat()`/`retreat_to()` mirror walking away mid-fight (including the server's 3-round rule).
- Combat rules (toggle retreat in duels, auto-retaliate, multi-target priority) are **mind layer** — not exposed here; use `when`/`select` to build them.

**Migration from the Flat API (executed):**

| old (flat) | frozen | status |
| --- | --- | --- |
| `attack(t)` | `combat.attack(t)` | done — `attack` kept as the §9 alias |
| (none) | `combat.engaged` | exists |
| (none) | `combat.target` | exists |
| `set_combat_style(...)` | `combat.set_style(...)` | done — flat name removed |
| (none) | `combat.last_npc` | exists |
| (none) | `combat.last_player` | exists |
| (none) | `combat.retreat()` / `combat.retreat_to(x,y)` | exists |
| (none) | `combat.style` | exists (read-side intent mirror) |

### `magic` — Views + Actions

#### `magic` — Views

- **`magic.level`** → `Int`. Faculty: View. Current Magic skill level (effective/boosted — `SkillLevel(6)`). Range: 1–99. GUI: the Magic skill level in the stats panel. Tag: (exists — `runtime/views_magic.go`).

- **`magic.max_level`** → `Int`. Faculty: View. Base Magic skill level (unboostable max). Range: 1–99. GUI: the Magic stat level before boosts. Tag: (exists).

- **`magic.can_cast(spell: SpellRef)`** → `Bool`. Faculty: View. True iff you have the required Magic level to cast the spell (gates on the *current/boosted* level; not a rune/gear check). SpellRef accepts Int(spellbook_id) | String(name) | SpellDef. Returns false if spell unknown or level insufficient. Null-safe: unknown spell returns false, never Null. GUI: the spell book UI highlights castable spells; hovering shows req level. Tag: (exists — `spellCanCastCallable`).

- **`magic.known`** → `List<SpellDef>`. Faculty: View. Spells whose `req_level` ≤ the *current/boosted* Magic level (the promoted `magic.spells.known` below keys off the base level instead). Tag: (exists).

- **`magic.has_runes_for(spell: SpellRef)`** → `Bool`. Faculty: View. Root-level rune check that **honors equipped elemental staves** (an air staff satisfies the air-rune requirement, etc. — `elementalStaffItems` in `runtime/views_magic.go`); the promoted `magic.spells.has_runes_for` is a plain inventory-only check. Tag: (exists).

- **`magic.spells`** → `SpellsView`. Faculty: View. Namespace for spell catalog accessors, promoted from `self.spells` (which remains for back-compat). Tag: (exists).
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

  **Tag:** (exists — `dslMagicCast` in `runtime/actions_magic.go`). The unified polymorphic verb shipped: target shape selects the opcode (omitted/Null → self; Npc view → combat cast; Player view → PvP; item view or named `spell=` → inventory-item cast; `[x, y]` list or any value with `.x`/`.y` → tile-target). The flat `cast_on_self/npc/player/land/item` names were **removed from the DSL surface** in the §10 migration — their bodies survive only as Go backing functions `magic.cast` dispatches to. `cast(spell, target?)` remains as the sanctioned §9 alias.

#### `magic` — Events

- **`on spell_cast(caster: Player, spell: SpellDef, target?: Target)`** — You (caster=self) or another player cast a spell. `target` is Null for self-targeted, or the target entity (Npc, Player, Loc, Position, GroundItem, Boundary). Fires on both sides (caster's client + spectators). GUI: the spell animation plays in-world. Tag: (to build — spell message interception pending; currently requires parsing world.last_server_message or game chat for "You cast X").

- **`on spell_hit(caster: Player, spell: SpellDef, damage?: Int)`** — Caster's spell landed (hit). `damage` is Int (hitsplat value) for offensive spells, Null for non-damage (buffs, teleports, etc.). Fires on the target's client and the caster's. GUI: hitplat appears on the target. Tag: (to build — awaits damage-event packet decoder).

- **`on spell_resist(caster: Player, spell: SpellDef)`** — Spell was resisted / dodged (magic defense success). Caster and target see this. GUI: the "Resist" message appears. Tag: (to build).

---

#### Migration map (flat → frozen magic) — executed

| old (flat) | frozen | why |
| --- | --- | --- |
| `cast_on_self(spell_id)` | `magic.cast(spell)` | Self-targeted cast; spell by id or name. |
| `cast_on_npc(npc, spell_id)` | `magic.cast(spell, target=npc)` | NPC combat-cast; overload dispatch by target type. |
| `cast_on_player(player, spell_id)` | `magic.cast(spell, target=player)` | PvP cast. |
| `cast_on_land(x, y, spell_id)` | `magic.cast(spell, target={x,y})` or `magic.cast(spell, target=Position{x,y})` | Tile-targeted AOE. |
| `cast_on_item(item, spell_id)` | `magic.cast(item, spell=spell_id)` | Inventory-item cast (alchemy, enchanting). Spell passed via named arg to disambiguate from positional item. |
| `self.spells.*` | `magic.spells.*` | Promoted to root namespace per §6 (subsystem naming). |
| — | `magic.level` | Alias for `self.skills.magic.level` (skill index 6) + convenience + future extensibility (e.g., per-book level variations). |
| — | `magic.max_level` | Alias for `self.skills.magic.max_level` (skill index 6). |
| — | `magic.can_cast(spell)` | Pure-read check: "can I cast this?" (exists). |

### `prayer` — Views + Actions

#### `prayer` — Views

All backed by `prayerView` in `runtime/views_prayer.go` (catalog/count reads fall through to the back-compat `self.prayers` view).

- **`prayer.active(slot: Int | name: String)`** → `Bool`. Faculty: View. Whether one prayer is currently on; accepts an Int slot index or a case-insensitive prayer name. Bounds-checked; never null. GUI: lit prayer buttons in the prayer book. (exists — note this is a *callable* on the prayer root; under back-compat `self.prayers.active` the same word reads as the active list.)
- **`prayer.active_list`** → `List<Int>`. Faculty: View. The currently-active prayer slot indices (decoded from opcode 206). Empty list if none. (exists)
- **`prayer.count`** → `Int`. Number of active prayers. (exists)
- **`prayer.is_active(slot)`** → `Bool`. Per-slot check (same as `prayer.active(slot)`). (exists)
- **`prayer.book`** → `List<PrayerDef>`. Faculty: View. The static prayer catalog (from facts). Never null. GUI: the prayer book interface. (exists)
- **`prayer.by_id(id)`** → `PrayerDef | Null`. Faculty: View. Look up a prayer by slot index (0–14). Null if out of range. (exists)
- **`prayer.by_name(name)`** → `PrayerDef | Null`. Faculty: View. Look up a prayer by name (case-insensitive, from facts). Null if no match. (exists)

There is no `prayer.points` — current prayer points stay on the self root as `self.prayer` / `self.max_prayer`.

Each `PrayerDef` carries `.id` / `.name` / `.req_level` / `.drain_rate` / `.description` (all static from facts).

#### `prayer` — Actions (all → `Result<Null>`)

- **`prayer.activate(prayer: Int | String)`** → `Result<Null>`. Faculty: Action. Activate a prayer by slot index (0–14) or name (`resolvePrayerID` accepts both). The server silently rejects if prayer level is too low or prayer points are zero; check the result or poll `prayer.active(N)` to confirm. Errors: `SERVER_REJECTED` (level requirement not met, no prayer points, etc.). GUI: click the prayer icon in the prayer panel. Tag: (exists — namespaced via `prayerVerbs`; the flat `activate_prayer` was removed in the §10 migration).
- **`prayer.deactivate(prayer: Int | String)`** → `Result<Null>`. Faculty: Action. Deactivate (toggle off) a prayer. Errors: `SERVER_REJECTED` (rarely; shouldn't fail). GUI: click the active prayer icon to toggle off. Tag: (exists — flat `deactivate_prayer` removed).

#### `prayer` — Notes

- **Namespacing (§6):** `self.prayers.*` promoted to top-level `prayer.*` per the contract; `self.prayers.*` remains for back-compat.
- **No events** in this namespace (acknowledged; not in `dsl/spec/events.go`).
- **No perception gap:** all player-visible prayer state (active list, slot activity, catalog) is surfaced; points live on `self.prayer`.

### `trade` — Frozen API surface

> **§10 executed.** The top-level `trade` root is live (`runtime/views_trade.go`
> + `tradeVerbs` in `runtime/dsl_actions.go`): frozen view names are canonical
> (with the `first/second_accepted` names kept as aliases), the five frozen
> action verbs are view-dispatched, and the old flat verbs
> (`trade_request`, `accept_trade`, `offer_trade`, `confirm_trade`,
> `finalize_trade`, `decline_trade`) are **gone from `dsl/spec/actions.go`**.

#### `trade` — Views

- **`trade.is_active`** → `Bool`. Faculty: View. True while a trade window is open (phase ≠ "completed" and ≠ "cancelled"). GUI: the trade window is visible on screen. Tag: (exists).

- **`trade.phase`** → `String` ∈ {`"none"`, `"request_sent"`, `"open"`, `"confirm"`, `"completed"`, `"cancelled"`}. Faculty: View. Which screen is showing. `"none"` when no trade is active; `"request_sent"` after initiating a request (before symmetric acceptance); `"open"` on the offer screen (both sides accepted request); `"confirm"` on the final review screen; `"completed"` or `"cancelled"` on terminal. GUI: none (internal state). Tag: (exists).

- **`trade.with`** → `String | Null`. Faculty: View. The opponent's username, or null if no trade active. GUI: opponent display. Tag: (exists).

- **`trade.with_index`** → `Int`. Faculty: View. The opponent's server-side player index, or 0 if no trade active. GUI: none. Tag: (exists).

- **`trade.my_offer`** → `List<[Int, Int]>`. Faculty: View. Items you have put up for trade; each element is `[item_id: Int, quantity: Int]`. Empty list (never null) if no trade or you haven't offered yet. GUI: your side of the offer panel. Tag: (exists).

- **`trade.their_offer`** → `List<[Int, Int]>`. Faculty: View. The opponent's offered items, same shape. Empty list if no trade or they haven't offered yet. GUI: their side of the offer panel. Tag: (exists).

- **`trade.accepted`** → `Bool`. Faculty: View. True iff you clicked "Accept" on the **offer** screen (screen 1). Resets to false if either party changes their offer. GUI: first Accept button visual state. Tag: (exists; alias `my_first_accepted`).

- **`trade.they_accepted`** → `Bool`. Faculty: View. True iff the opponent clicked "Accept" on the **offer** screen. Resets on offer change. GUI: opponent's first Accept state. Tag: (exists; alias `their_first_accepted`).

- **`trade.both_accepted`** → `Bool`. Faculty: View. True iff both parties accepted the offer screen → the confirm screen appears. GUI: both-accepted indicator. Tag: (exists; alias `both_first_accepted`).

- **`trade.confirmed`** → `Bool`. Faculty: View. True iff you clicked "Accept" on the **confirm** screen (screen 2). Only meaningful after `both_accepted`. GUI: second Accept button visual state. Tag: (exists; alias `my_second_accepted`).

- **`trade.they_confirmed`** → `Bool`. Faculty: View. True iff the opponent clicked "Accept" on the **confirm** screen. Only meaningful after both parties' first accept. GUI: opponent's confirm state. Tag: (exists; alias `their_second_accepted`).

- **`trade.both_confirmed`** → `Bool`. Faculty: View. True iff both parties confirmed the final screen → trade executes. GUI: trade-complete indicator. Tag: (exists; alias `both_second_accepted`).

#### `trade` — Actions (all → `Result<Null>`)

- **`trade.request(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent to `p`, then sends a trade request (wire opcode 142). Dual-purpose: initiates a new trade request AND accepts an incoming one — in RSC both sides must request each other (symmetric handshake) for the window to open, so this one verb absorbed the old `trade_request` + `accept_trade` pair (same packet). String names resolve via world.Players deterministically. Errors: `TARGET_OUT_OF_VIEW` if `p` (player view, server-index Int, or player-name String) is not visible in the world. GUI: right-click player → "Trade with". Tag: (exists — `dslTradeRequest`).

- **`trade.offer(items: List<[Int, Int]>)`** → `Result<Null>`. Faculty: Action. Sets/replaces your offered items; each element is `[item_id: Int, amount: Int]`. Per server rule, any offer change resets both parties' accept flags on the offer screen. Errors: `TRADE_NOT_ACTIVE` if no trade is open; `NO_SUCH_ITEM` if you don't hold an offered item. GUI: drag items into your offer panel. Tag: (exists — `dslOfferTrade`).

- **`trade.accept()`** → `Result<Null>`. Faculty: Action. Clicks "Accept" on the **offer** screen (screen 1). Both sides must accept before the confirm screen appears. Idempotent: re-calling after you've already accepted is a no-op, and it re-fires automatically if an offer change reset your acceptance. Errors: `TRADE_NOT_ACTIVE` if no trade is active. GUI: the first "Accept" button. Tag: (exists — `dslConfirmTrade`, the old offer-screen accept body).

- **`trade.confirm()`** → `Result<Null>`. Faculty: Action. Clicks "Accept" on the **confirm** screen (screen 2), completing your half of the trade. The confirm screen only appears after `trade.both_accepted`. Idempotent on extra calls. Errors: `TRADE_NOT_ACTIVE` if no trade is active; returns error if called before the offer screen is accepted (check `.accepted` first). GUI: the second "Accept" button. Tag: (exists — `dslFinalizeTrade`).

- **`trade.decline()`** → `Result<Null>`. Faculty: Action. Declines/closes the trade window (can be called at any phase). Marks the trade phase → "cancelled". Errors: `TRADE_NOT_ACTIVE` if no trade is active. GUI: close button / cancel. Tag: (exists — `dslDeclineTrade`).

#### `trade` — Events

- **`on trade_request(from: String)`** — Another player initiated a trade request with you. `from` is the requester's player name (String); call `trade.request(from)` to accept the symmetric handshake. OpenRSC's notification packet carries name only, not server index; lookup via world.Players. Tag: (exists — spec'd in `dsl/spec/events.go` and wired in `runtime/dsl_events.go` from `event.TradeRequestReceived`).

- **`on trade_opened(other_index: Int)`** — Both players have accepted the initial trade request; the offer screen is now open. `other_index` is the opponent's server-side player index. Use `trade.my_offer`, `trade.their_offer`, and `trade.with` / `trade.with_index` to read state. GUI: the offer panel appears. Tag: (exists; the contract's `other: Player` param shape remains an open rename — today it is the Int index).

- **`on trade_other_offer(items: List<[Int, Int]>)`** — The opponent updated their offered items. `items` is the new list of `[item_id, amount]` pairs. Mirrors `trade.their_offer`. Both parties' offer-screen accept flags reset. GUI: opponent's offer side updates. Tag: (exists).

- **`on trade_other_accepted()`** — The opponent clicked "Accept" on the current screen (offer or confirm). No parameter. Allows reacting before both-sides acceptance. GUI: opponent's Accept button visual state changes. Tag: (exists).

- **`on trade_closed(completed: Bool)`** — The trade ended. `completed` is true iff both sides confirmed the final screen (goods exchanged); false on decline/cancel. Server protocol close packet has no completion bit; implementation infers from world.Trade.Phase == "completed" after Apply. GUI: trade window closes. Tag: (exists).

- **`on trade_confirm_shown()`** — Server pushed the final confirm-window — both sides have first-accepted and we're on the final review screen. No parameter. GUI: confirm panel appears. Tag: (exists).

#### Capability boundary & implementation notes

- All Views report screen-visible state or cached world-mirror state (opponent name, phase, accepted flags, offer lists). The `.with_index` is a resolved lookup convenience matching Npc-resolution rules in §2.
- The `.my_offer` / `.their_offer` are lists of items each side has proposed; the GUI player sees them by dragging to the window. Format is List<[Int, Int]> (item_id, quantity pairs), not abstract List<Item>.
- The twin-accept flow (offer screen, then confirm screen) is the frozen, immutable RSC server protocol — the API surfaces it directly as separate `trade.accept()` / `trade.confirm()` verbs and separate offer/confirm accept flags.
- Trade requires adjacency to initiate (opcode 142); the action automatically pathfinds to the target player's last-known position.
- The old `trade_request` and `accept_trade` verbs were dual-purpose (same opcode 142), so the frozen API consolidated them to `trade.request(p: PlayerRef)`.
- Phase transitions: "none" (idle) → "request_sent" (after initiating) → "open" (both accepted request) → "confirm" (both first-accepted) → "completed" or "cancelled" (terminal).

#### Frozen nomenclature mapping (§10) — executed

| Old (flat API, removed) | Frozen namespaced API | Reason |
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

- **Views** (all exist): `trade.is_active`, `trade.phase`, `trade.with`, `trade.with_index`, `trade.my_offer`, `trade.their_offer`, and the six accept flags under their frozen names (`accepted` / `they_accepted` / `both_accepted` / `confirmed` / `they_confirmed` / `both_confirmed`) with the `first/second_accepted` back-compat aliases.
- **Actions** (all exist, namespaced only): `trade.request()`, `trade.offer()`, `trade.accept()`, `trade.confirm()`, `trade.decline()`.
- **Events** (all exist): `trade_request`, `trade_opened` (Int index param; the contract's `other: Player` shape is an open rename), `trade_other_offer`, `trade_other_accepted`, `trade_closed`, `trade_confirm_shown`.

### `bank` — Views + Actions

#### `bank` — Views

- **`bank.is_open`** → `Bool`. Faculty: View. True while a bank window is open. GUI: the bank window is on screen. (exists)
- **`bank.max_size`** → `Int`. Bank slot capacity. Returns 0 if bank not open. GUI: visible from the window chrome. (exists)
- **`bank.used`** → `Int`. Number of occupied bank slots. Returns 0 if bank not open. GUI: inferred from slot display. (exists)
- **`bank.free`** → `Int`. Available slots (max_size - used). Returns 0 if bank not open. GUI: free visual space. (exists)
- **`bank.slots`** → `List<[Int, Int]>`. Bank contents; each element is `[item_id, amount]`. Empty list if bank not open (never null). GUI: the itemized inventory pane. (exists)
- **`bank.count(item: ItemRef)`** → `Int`. Total quantity of an item in the bank (0 if absent). Alias: `bank.has()`. GUI: calculated from visible slots. (exists)

#### `bank` — Actions (all → `Result<Null>`)

- **`bank.open(banker: NpcRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent to `banker`, opens the dialog, **finds and selects the bank-access option itself** (matches the banker menu against `bankAccessNeedles` — "access my bank" / "bank account" / … — then answers and polls for the opcode-42 bank window; `dslOpenBank` in `runtime/actions_bank.go`). No manual `answer()` step needed. Errors: `TARGET_OUT_OF_VIEW` if banker is not visible; `PATH_BLOCKED` if unreachable; `BANK_NOT_OPEN` if no bank-access option is in the menu or the window never opens within the poll budget. GUI: walk to banker, right-click → "Talk To", pick the bank option. (exists — namespaced; flat `open_bank` removed)
- **`bank.deposit(item: ItemRef, amount: Int)`** → `Result<Null>`. Faculty: Action. Deposits `amount` of the item from inventory into the open bank. Errors: `NO_SUCH_ITEM` if inventory does not hold the item; `SERVER_REJECTED` for server rejections (including not-open). GUI: drag item into bank window. (exists)
- **`bank.withdraw(item: ItemRef, amount: Int)`** → `Result<Null>`. Faculty: Action. Withdraws `amount` of the item from the open bank into inventory. Errors: `NO_SUCH_ITEM` if bank does not hold the item; `INVENTORY_FULL` if inventory cannot accept the item (hard limit: 30 slots); `SERVER_REJECTED` for other server rejections. GUI: drag item from bank window. (exists)
- **`bank.deposit_all(keep?: List<ItemRef>)`** → `Result<Null>`. Faculty: Action. Deposits every inventory item except those in the optional `keep` list. Gates client-side on the bank being open. Errors: `BANK_NOT_OPEN`; `NO_SUCH_ITEM` if a keep ref doesn't resolve; individual item deposits may fail per `deposit()` rules (no early exit — continues on item-level errors, logs warnings). GUI: emulated via repeated drag (a routine convenience). (exists — `dslDepositAll`, registered in `bankVerbs`; #117/#118)
- **`bank.withdraw_all(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Withdraws the bank's whole stack of the item. Errors: `BANK_NOT_OPEN`; per-`withdraw()` rules otherwise. (exists — `dslWithdrawAll`; #117/#118)
- **`bank.withdraw_x(item: ItemRef, amount: Int)`** → `Result<Null>`. Faculty: Action. Withdraw-X convenience. Errors: `BANK_NOT_OPEN`; per-`withdraw()` rules otherwise. (exists — `dslWithdrawX`; #117/#118)
- **`bank.close()`** → `Result<Null>`. Faculty: Action. Closes the bank window. Clears the server's `accessingBank` flag. Errors: None (idempotent; succeeds even if not open). GUI: close-button or move away. (exists — namespaced; flat `close_bank` removed)

#### `bank` — Events

- **`on bank_opened(max_size: Int)`** — Bank window opened (talk_to banker → select bank from dialog, or right-click banker → "Bank"). `max_size` is the slot capacity (typically 48 for free players, 54 for members). `bank.slots` has the contents. (exists)
- **`on bank_slot_update(slot: Int, item_id: Int, amount: Int)`** — A bank slot was updated. `amount=0` means the slot was emptied. `bank` state reflects the change before the handler fires. (exists)
- **`on bank_closed()`** — Bank window closed. `bank.is_open` is false from this point. (exists)

**Tags & Notes:**
- All Views: report the mirror state (no round-trip to server).
- All Actions: send packets, return `Result<Null>` on immediate acceptance (async consequences via mirror updates + events).
- **Naming:** the §10 promotion is done — `bank.open/deposit/withdraw/deposit_all/withdraw_all/withdraw_x/close` are the only names (`bankVerbs` in `runtime/dsl_actions.go`); the flat `open_bank`/`deposit`/`withdraw`/`close_bank` rows are gone from `dsl/spec/actions.go`.
- **Error code coverage:** the bulk verbs (`deposit_all`/`withdraw_all`/`withdraw_x`) and `bank.open` gate client-side and return `BANK_NOT_OPEN` exactly. Plain `deposit`/`withdraw` classify server errors through `wrapServerErr` (`runtime/dsl_actions.go`): typed `*DoorLockedError` → `DOOR_LOCKED` first, then string-match ("too far"/"out of range" → `OUT_OF_RANGE`, "path"/"stalled" → `PATH_BLOCKED`, "inventory full" → `INVENTORY_FULL`, "canceled"/"deadline exceeded" → `INTERRUPTED`, "timeout" → `ACTION_TIMEOUT`, else `SERVER_REJECTED`).
- **Perception gap:** None — all views (is_open, max_size, used, free, slots, count) are GUI-visible and fully implemented in `runtime/views_bank.go`.
- **No server secrets:** All views report only what the player sees on the bank window UI.

### `duel` — Views + Actions

#### `duel` — Views

All backed by `duelView` in `runtime/views_duel.go`; same first/second back-compat aliases as trade.

- **`duel.is_active`** → `Bool`. Faculty: View. True while a duel window is open. GUI: the duel window is on screen. (exists)
- **`duel.phase`** → `String` ∈ {`"none"`, `"request_sent"`, `"open"`, `"confirm"`, `"completed"`, `"cancelled"`}. Which phase the duel is in. (exists)
- **`duel.my_offer`** → `List<[Int, Int]>`. Items you have staked; each element is `[item_id, amount]` (empty list, never null). Same field name as trade — there is no separate `my_stake` spelling. GUI: your side of the offer. (exists)
- **`duel.their_offer`** → `List<[Int, Int]>`. The opponent's staked items. (exists)
- **`duel.accepted`** → `Bool`. You clicked Accept on the **offer** screen. (exists; alias `my_first_accepted`)
- **`duel.they_accepted`** → `Bool`. They did. (exists; alias `their_first_accepted`)
- **`duel.both_accepted`** → `Bool`. Both did → the confirm screen opens. (exists; alias `both_first_accepted`)
- **`duel.confirmed`** → `Bool`. You clicked Accept on the **confirm** screen. (exists; alias `my_second_accepted`)
- **`duel.they_confirmed`** → `Bool`. They did on the confirm screen. (exists; alias `their_second_accepted`)
- **`duel.both_confirmed`** → `Bool`. Both confirmed the fight. (exists; alias `both_second_accepted`)
- **`duel.with`** → `String | Null`. The opponent's username (null if inactive). The contract's `duel.opponent` rename never shipped — `with` / `with_index` match trade. (exists)
- **`duel.with_index`** → `Int`. The opponent's server-side index (0 if inactive). (exists)
- **`duel.disallow_retreat`** → `Bool`. Retreat rule is active. (exists)
- **`duel.disallow_magic`** → `Bool`. Magic rule is active. (exists)
- **`duel.disallow_prayer`** → `Bool`. Prayer rule is active. (exists)
- **`duel.disallow_weapons`** → `Bool`. Weapons rule is active. (exists)

#### `duel` — Actions (all → `Result<Null>`)

- **`duel.request(p: PlayerRef)`** → `Result<Null>`. Faculty: Action. Walks adjacent, then sends a duel request to `p` (wire opcode 103). In RSC this one action both *initiates* a duel and *accepts an incoming* one — clicking the player serves both, and mutual requests open the window (so there is no separate "accept request" verb). Errors: `OUT_OF_RANGE` if `p` is not a visible or reachable player. GUI: right-click player → "Challenge". (exists — `dslDuelRequest`; flat `duel_request`/`accept_duel` removed)
- **`duel.stake(items: List<[Int, Int]>)`** → `Result<Null>`. Sets/replaces your stake; each element is `[item_id, amount]`. Per the server rule this **resets both parties' offer-accepts**. Errors: `TRADE_NOT_ACTIVE` if no duel is active; `NO_SUCH_ITEM` if you don't hold a staked item. GUI: drag items into the window. (exists — `dslOfferDuel`)
- **`duel.set_rules(retreat?: Bool, magic?: Bool, prayer?: Bool, weapons?: Bool)`** → `Result<Null>`. Sets the four rule toggles — named args, or one positional 4-element Bool list `[retreat, magic, prayer, weapons]`; pass `true` to disallow. Either side can change rules from the offer screen; the server unifies them and broadcasts the result. A change resets both first-accept flags. Errors: `TRADE_NOT_ACTIVE`. GUI: toggle buttons on the offer screen. (exists — `dslSetDuelRules`)
- **`duel.accept()`** → `Result<Null>`. Clicks "Accept" on the **offer** screen (screen 1). Idempotent; re-fires automatically if a stake change reset your accept. Errors: `TRADE_NOT_ACTIVE`. GUI: the first Accept button. (exists — `dslAcceptDuelOffer`)
- **`duel.confirm()`** → `Result<Null>`. Clicks "Accept" on the **confirm** screen (screen 2), completing your half. The confirm screen only opens after `duel.both_accepted`. Errors: `TRADE_NOT_ACTIVE`; `NOT_IMPLEMENTED`/not-ready if called before the offer screen is accepted. GUI: the second Accept button. (exists — `dslAcceptDuelConfirm`)
- **`duel.decline()`** → `Result<Null>`. Declines/closes the duel at any phase. Errors: `TRADE_NOT_ACTIVE`. GUI: close the window. (exists — `dslDeclineDuel`)

#### `duel` — Events

All spec'd in `dsl/spec/events.go` and wired in `runtime/dsl_events.go`. (An earlier revision listed hypothetical names — `duel_requested`, `duel_stake_updated`, `duel_rules_changed` — tagged "to build"; the real catalog below shipped instead, mirroring the trade event names.)

- **`on duel_request_incoming(from: String)`** — duel request received from `from` (their username; name-only, like `trade_request`). (exists)
- **`on duel_opened(other_index: Int)`** — window opened after both sides sent a request. (exists)
- **`on duel_other_offer(items: List<[Int, Int]>)`** — opponent changed their stake. (exists)
- **`on duel_settings_update(disallow_retreat: Bool, disallow_magic: Bool, disallow_prayer: Bool, disallow_weapons: Bool)`** — the rules changed. (exists)
- **`on duel_other_accepted()`** — opponent accepted the offer screen. (exists)
- **`on duel_confirm_shown()`** — the confirm screen is now shown. (exists)
- **`on duel_closed(completed: Bool)`** — closed; `completed` iff the duel began (both sides confirmed). (exists)

### `shop` — Views + Actions

**BUILT.** The `shop` namespace mirrors the bank/trade/duel pattern and is live end-to-end: the `shop` root is registered (`it.Reserved["shop"]` in `runtime/dsl_bridge.go`), `shopView` + the `stock`/`price` callables live in `runtime/views_shop.go`, the action bodies in `runtime/actions_shop.go`, the `world.Shop` mirror (`ShopState`) is fed by the inbound shop-open/close decoders (opcodes 101/137), and `SHOP_NOT_OPEN` is a typed code in `dsl/interp/error.go`. Only the events remain open.

#### `shop` — Views

- **`shop.is_open`** → `Bool`. Faculty: View. True while a shop window is open. GUI: the shop window visible on screen. (exists)
- **`shop.is_general`** → `Bool`. Faculty: View. True if the open shop is a general store. False when closed. (exists)
- **`shop.slots`** → `List<[Int, Int, Int]>`. Faculty: View. The shop catalogue; each element is `[item_id, stock, base_stock]`. Empty list when closed. GUI: the shop item grid. (exists)
- **`shop.stock(item: ItemRef)`** → `Int`. Faculty: View. Quantity of the item currently in stock (0 if not stocked / shop closed). GUI: the quantity shown in the shop list. (exists)
- **`shop.price(item: ItemRef)`** → `Int`. Faculty: View. The unit **buy** price in gp. The shop packet carries no prices — the view resolves the item's catalogue base price from `facts.ItemDef` and applies the authentic client price formula (`world.ShopState.BuyPrice`). Returns 0 if closed/unstocked/unknown def. GUI: the unit price on hover/click. (exists)

#### `shop` — Actions (all → `Result<Null>`)

View-dispatched member calls on the shop root (per the bank/trade pattern); there are no flat `shop_*` builtins in `dsl/spec/actions.go`.

- **`shop.buy(item: ItemRef, qty: Int)`** → `Result<Null>`. Faculty: Action. Buys `qty` of the item from the open shop. Errors: `SHOP_NOT_OPEN` (gated client-side before any packet); `SERVER_REJECTED` (out of stock, member-only item for a free player, insufficient coins). GUI: click item → choose qty → confirm. (exists — `dslShopBuy`)
- **`shop.sell(item: ItemRef, qty: Int)`** → `Result<Null>`. Faculty: Action. Sells `qty` of the item to the open shop. Errors: `SHOP_NOT_OPEN`; `NO_SUCH_ITEM` if inventory does not hold it (checked client-side); `SERVER_REJECTED` (shop won't buy it). GUI: right-click item → sell/sell-all → confirm. (exists — `dslShopSell`)
- **`shop.close()`** → `Result<Null>`. Faculty: Action. Closes the shop window. GUI: close button or ESC. (exists — `dslShopClose`)

#### `shop` — Events

Not yet in `dsl/spec/events.go` — poll `shop.is_open` / `shop.slots` meanwhile.

- **`on shop_opened(npc: Npc)`** — Shop window opened via NPC interaction. (to build)
- **`on shop_stock_update(item_id: Int, new_stock: Int)`** — A shop slot's stock changed (purchase, sale, or restock). (to build)
- **`on shop_closed()`** — Shop window closed. (to build)

#### `shop` — Design notes

- **No separate `shop.open()` action:** shop opening is via NPC interaction (right-click → Trade or talk_to + dialog), mirroring bank.
- **No per-NPC shop routing:** single-shop assumption matches the bank/trade/duel pattern (one `world.Shop` record at a time).
- **Sell-price derivation:** the server enforces sell pricing; the API exposes buy price only.
- **`world.shop.*` does not resolve** — unlike bank/trade/duel there is no back-compat path under the `world` root; the shop surface is top-level `shop.*` only.

### Ambient verbs — top-level Actions and Primitives

**Ambient verbs** are top-level Actions and Primitives owned by no subsystem root — they model the core interaction loops of a player: movement, generic item interaction, NPC engagement, scenery manipulation, combat, social chat, and session control. They read like player-speech rather than API subdomains. All are documented per §7 depth bar (signature, faculty, semantics, errors, nullability, GUI equivalent, tags).

#### Movement & Navigation

- **`walk_to(x: Int, y: Int, attempt_open_doors?: Bool)`** → `Result<Null>`. Faculty: Action. Walk to the absolute-coordinate tile (x, y); blocks until arrival or failure. Optional `attempt_open_doors` (default true) controls whether doors are auto-opened on the path or the walk fails (strict mode for scouts). Errors: `PATH_BLOCKED` (no path exists after all obstacles); `DOOR_LOCKED` (a door on the path couldn't be opened; reason includes server prose like "you need a key"); `OUT_OF_RANGE` (destination too far for one walk cycle). GUI: click the minimap or right-click a tile and walk. (exists)
- **`walk_path(corners: List<[Int, Int]>)`** → `Result<Null>`. Faculty: Action. Pre-planned multi-corner walk; each element is `[x, y]`. Max 25 corners per packet; for longer routes chunk via repeated calls. Server interpolates each segment. Useful for cognitive routes the routine planned itself. Errors: `PATH_BLOCKED` (packet send fails). GUI: used by the cognitive layer for pre-computed routes; no direct GUI equivalent. (exists)
- **`is_reachable(x: Int, y: Int)`** → `Bool` OR **`is_reachable(target: View)`** → `Bool`. Faculty: View (pure; no packet). Chebyshev pathfinding from self.position to (x, y). Bounded by 96-tile FOV grid; cross-region checks use sequential is_reachable along a planned chain. Target can be any view with .x/.y (Player, Npc, Loc, GroundItem, Position, etc.) or named (x=X, y=Y). Returns true iff reachable, false if not. GUI: none (used by pathfinding logic). (exists)
- **`go_to(place: String | x: Int, y: Int)`** → `Result<Null>`. Faculty: Action. Walk to a named gazetteer place (resolved through the world-map facts) or to explicit coordinates. Errors: `NO_SUCH_ITEM` if the name doesn't resolve / no map data; walk errors as per `walk_to`. (exists — `dslGoTo` in `runtime/actions_movement.go`)

#### Interaction: Items & Inventory

- **`pick_up(ground_item: GroundItem)`** → `Result<GroundItem>`. Faculty: Action. Walk adjacent and pick up a ground item; returns the picked-up item-view on success so routines can do `got = pick_up!(item); say(got.name)`. Errors: `OUT_OF_RANGE` (too far to reach); `INVENTORY_FULL` (no slot available); `TARGET_OUT_OF_VIEW` (item no longer visible). GUI: right-click ground item → "Take". (exists)
- **`drop(item: ItemRef, amount?: Int)`** → `Result<Null>`. Faculty: Action. Drop the item from inventory to the ground; `amount` controls the count dropped (defaults to full stack). ItemRef can be an InvSlot, Int(slot_index), String(name), or ItemDef. Errors: `NO_SUCH_ITEM` (not in inventory); `OUT_OF_RANGE` (can't drop in current position). GUI: right-click inventory item → "Drop". (exists)
- **`eat(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Use an item from inventory (Eat/Drink/Bury — server decides by item def). Alias of `use_inventory_default(item)`. ItemRef as in `drop()`. Errors: `NO_SUCH_ITEM` (not in inventory); `EAT_IN_COMBAT` (typed code — RSC rejects item actions while fighting with "You can't do that whilst you are fighting"; retreat first or branch). GUI: right-click food item → "Eat", or right-click herbs → "Clean", or right-click bones → "Bury". (exists)
- **`equip(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Move an inventory item into its worn slot (armor, weapons, etc.). Server enforces level requirements and slot conflicts. ItemRef: InvSlot, Int(slot_index 0–29), String(name), or ItemDef. Errors: `NO_SUCH_ITEM` (not in inventory); `SERVER_REJECTED` (level requirement not met, slot conflict, or item not wearable). GUI: right-click inventory item → "Equip", or drag to equipment slot. (exists)
- **`unequip(item: ItemRef)`** → `Result<Null>`. Faculty: Action. Return a wielded item to inventory. ItemRef: InvSlot or String(name). Errors: `NO_SUCH_ITEM` (not wielded); `INVENTORY_FULL` (inventory is full). GUI: click the equipped item in equipment panel → "Remove", or right-click → "Take off". (exists)

#### Interaction: NPCs & Scenery

- **`talk_to(npc: NpcRef)`** → `Result<Null>`. Faculty: Action. Walk adjacent and open the NPC's dialog. NpcRef: Npc view (from world.npcs), Int(server_index). Errors: `OUT_OF_RANGE` (too far); `TARGET_OUT_OF_VIEW` (NPC no longer visible); `PATH_BLOCKED` (can't reach adjacent tile). GUI: right-click NPC → "Talk". (exists)
- **`answer(option: Int)`** → `Result<Null>`. Faculty: Action. Pick a dialog-menu option by **1-based** index (pairs with `find_option` / `world.dialog.find_option`). Errors: `DIALOG_NOT_OPEN`-class failures if no menu is showing. GUI: click the dialog option. (exists)
- **`converse(npc: NpcRef)`** → `Result<Null>`. Faculty: Action. Talk and **listen**: advances the NPC's dialog, stopping at each real choice so the routine can read what was said and then `answer(find_option("..."))`. Takes the NPC only — it never picks options for you. (exists — `dslConverse`)
- **`pickpocket(npc: NpcRef)`** → `Result<Null>`. Faculty: Action. Walk adjacent and fire the NPC's primary action command (command1) — e.g., "Pickpocket" on a Man or Robed Gnome. One attempt per call; loop for several attempts. `pickpocket` is the canonical frozen-contract name; the old `npc_command` surface name is **removed** per §10 (the Go handler body is still `dslNpcCommand`). NpcRef: Npc view or Int(server_index). Errors: `OUT_OF_RANGE`; `TARGET_OUT_OF_VIEW`; `PATH_BLOCKED`. GUI: right-click thievable NPC → "Pickpocket". (exists)
- **`interact_at(target: Position | Loc, option?: Int)`** → `Result<Null>`. Faculty: Action. Primary (option=1, default) or secondary (option=2) click on a scenery tile. The verb depends on the Location def — "Chop" on a tree, "Mine" on a rock, "Climb-Up" on a ladder, etc. Target: a Loc from world.locs or a Position {x, y}. option: 1 (primary) or 2 (secondary); defaults to 1. Errors: `OUT_OF_RANGE` (too far to reach); `TARGET_OUT_OF_VIEW` (tile no longer visible); `PATH_BLOCKED` (can't reach interact distance). GUI: right-click scenery → command1 or command2 from the location's def. (exists)
- **`use(item: ItemRef, target: Target)`** → `Result<Null>`. Faculty: Action. Use one item on a target. Target kind picks the opcode: boundary (key on door), GroundItem/InvSlot (chisel on gem, paint on stone), Loc/scenery (log on fire), NPC (potion on NPC), etc. ItemRef as in `drop()`. Target: Npc, Player, Loc, GroundItem, Boundary, Position, or named position {x, y}. Errors: `NO_SUCH_ITEM` (item not in inventory); `OUT_OF_RANGE` (target too far); `TARGET_OUT_OF_VIEW` (target no longer visible); `SERVER_REJECTED` (server rejects the interaction). GUI: drag inventory item onto target; or right-click item then click target. (exists)
- **`open_boundary(boundary: Boundary)`** → `Result<Null>`. Faculty: Action. Default click on a boundary (door, gate, fence, web, etc.). Action depends on location def — "Open", "Close", "Climb-Up", "Cut-Through". Boundary view from world.locs/boundaries or position-derived. Errors: `OUT_OF_RANGE` (too far to reach); `TARGET_OUT_OF_VIEW` (boundary no longer visible); `DOOR_LOCKED` (door is locked; reason includes server prose like "this door requires a key" or "members only"). GUI: right-click door/boundary → first option from the def's command1. (exists)

#### Combat

- **`attack(target: Target)`** → `Result<Null>`. Faculty: Action. Initiate combat with an NPC or player — the sanctioned §9 flat alias of `combat.attack`. Target: Npc, Player, or Int(server_index). Errors: `OUT_OF_RANGE` (too far); `TARGET_OUT_OF_VIEW` (no longer visible); `TARGET_DEAD` (already dead). GUI: right-click NPC/player → "Attack", or left-click with attack weapon equipped. (exists)
- **`attack_ranged(target: NpcRef)`** → `Result<Null>`. Faculty: Action. Fires from the current tile with **no melee pre-walk** — for safespot ranging (stand within bow range of a barriered target, then loop). Errors as `attack`. (exists — `dslAttackRanged` / `Host.AttackNpcRanged`)

#### Magic & Spells

The flat `cast_on_self/npc/player/land/item` family was **removed from the DSL surface** when the unified `magic.cast(spell, target?)` shipped (§10 executed — their Go bodies survive only as backing functions). Use `magic.cast(...)` (see the `magic` namespace above) or its sanctioned §9 flat alias `cast(spell, target?)`.

#### Prayer

The flat `activate_prayer` / `deactivate_prayer` verbs were **removed from the DSL surface** in the §10 migration. Use `prayer.activate(prayer)` / `prayer.deactivate(prayer)` (see the `prayer` namespace above) — both accept a slot index or a prayer name.

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
- **Map perception (world-map gazetteer)** — the host can orient itself on the world map via the control-plane verbs `where_am_i()`, `where_is(name)`, and `bearing_to(x, y)` (named places + typed points-of-interest), and `look_around` leads with location. The WorldOracle-backed reach explainers `search_map(name)`, `reachable(target)`, `survey_map()` inform (open/gated/blocked + what's needed) and never auto-route; `scan_for(type)` enumerates nearby scenery of a kind (rocks/trees/…) as a distance-sorted list. All live on the control plane; see [`actions.md`](actions.md) for signatures. (exists)
- **`nearest_npc(name?: String)`** → `Npc | Null`. Faculty: View (pure). The closest visible NPC, optionally filtered by name. (exists)
- **`find_option(needle: String)`** → `Int`. Faculty: View (pure). Returns the 1-based index of the first dialog option containing `needle` (case-insensitive substring), or 0 if none match. Used with `answer(find_option("Yes"))` after `wait_for_dialog()`. Returns Int. GUI: none. (exists)
- **`wait_for_dialog(timeout_seconds?: Int)`** → `Bool`. Faculty: Action (primitive-like, yields). Block until an NPC dialog menu opens, or until timeout_seconds elapses (default 5s). Polls every 200ms. Returns Bool: true if a menu landed, false on timeout. GUI: none (internal timing utility replacing brittle `wait N; if world.dialog.is_open` patterns). (exists)
- **Combat style** — the flat `set_combat_style` verb was removed in the §10 migration; use `combat.set_style(style)` (and read back `combat.style`). See the `combat` namespace above.

#### Spatial Primitives (for `bounds` blocks)

- **`box(x1: Int, y1: Int, x2: Int, y2: Int)`** → `Shape`. Faculty: Primitive (returns a Shape object for use inside `bounds <shape> { ... }`). Axis-aligned rectangle; inclusive edges; arg order normalized (min/max computed internally). Used in bounding spatial reactors (e.g., `bounds box(120, 640, 130, 660) { on item_appeared(...) { ... } }`). Returns Shape. GUI: none. (exists)
- **`circle(cx: Int, cy: Int, radius: Int)`** → `Shape`. Faculty: Primitive. Chebyshev-radius disk (RSC uses 8-way movement; circles look square). Used inside `bounds`. Returns Shape. GUI: none. (exists)
- **`near(radius?: Int)`** → `Shape`. Faculty: Primitive. Disk centered on self.position at routine start (radius default depends on context). Used inside `bounds near(8) { ... }` for "react to events within 8 tiles of where I logged in". Returns Shape. GUI: none. (exists)

#### Ambient verbs — notes

- **Top-level:** all of these verbs are truly ambient — subsystem-free, player-speech-idiom, owned by no root (not self.*, not world.*, not inventory.*, etc.).
- **Canonical names:** per the frozen contract (§10), ambient verbs "keep their names — already faculty-shaped and subsystem-free." `pickpocket` is canonical (frozen); `npc_command` was dropped, not aliased.
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
- **`format(template: String, args...)`** → `String`. Faculty: Primitive. Positional `{}` string templating, consumed left-to-right; `{{`/`}}` escape literal braces; args render exactly as f-string interpolation renders them. Returns a plain `String` (no `.val`/`.err`); no bang. Literal templates are arity-checked at validation time; a dynamic-template mismatch aborts with `FORMAT_MISMATCH`. GUI: none. (exists)

#### Cognition bridge (mind-access — NOT GUI-equivalent, fenced)

These hit the knowledge/brain/memory layer, not the game. In production (`runtime/runhost.go`, formerly `runhost_bootstrap.go`) `Host.Strategist` / `Host.Retriever` are the real mesa-backed brain + recall (`mesaclient.AsStrategist` / `AsRetriever`); `brain.StubStrategist` / `cognition.StubClient` remain the deterministic defaults for tests and offline runs.

- **`recall(query: String, top?: Int)`** → `List<Chunk>`. Faculty: MemoryStdlib. Knowledge/memory search — prefers the host's knowledge corpus (`Host.Corpus`) and falls back to the Retriever. Tag: (exists — mesa-backed in production).
- **`relation_with(name: String)`** → relational record. Faculty: MemoryStdlib. Look up the host's relationship with another entity. Tag: (exists).
- **`contemplate_reality(question?: String)`** → decision. Faculty: LLMStdlib. Open-ended LLM decision via `Host.Strategist`. Tag: (exists — mesa-backed in production).
- **`evaluate(situation)`** → `Float`. Faculty: LLMStdlib. 0–1 confidence assessment. Tag: (exists).
- **`decide(options, context?)`** → option. Faculty: LLMStdlib. Pick one option (Strategist verdicts are memoized in the host's bounded decision cache). Tag: (exists).
- **`remember(key: String, value)`** / **`recollect(key)`** / **`forget(key)`**. Faculty: MemoryStdlib. The durable host-memory verbs over the tiered `memory` package (local store + write-back journal to mesa). Tag: (exists — `dslRemember`/`dslRecollect`/`dslForget`).
- **`resolve(text: String, kind?: String)`** → `List<Match{ def, kind, score }>`. Faculty: recognition / fuzzy resolution (mind-access, learnable). Ranked candidates, best first. Pipeline: the host's learned-alias store → conservative fuzzy/token match vs canonical names → brain (LLM) fallback (which on success writes the resolved alias back). Definitions and ids come from the facts catalog — never invented by the LLM. `kind` restricts the search to one catalog: "item", "npc", "loc", "spell", "prayer". Each Match also reports which pipeline `.stage` produced it ("alias"/"fuzzy"/…). Tag: (exists — `runtime/actions_resolve.go` + `Host.Resolver`, exactly the §5 pipeline).
- **`resolve_one(text: String, kind?: String)`** → `Match | Null`. The common-case sugar: best match or Null. Tag: (exists).
- **`ask_brain(prompt: String)`** → `String`. Faculty: LLMStdlib. Open-ended brain-access decision (like `contemplate_reality` but explicitly brain-access). Tag: (to build — not in spec; routines use `contemplate_reality` instead).
- **`exec(prompt: String)`** → `Result<Null>`. Faculty: LLMStdlib. LLM fragment executor. Tag: (to build — spec'd `NotYetImplemented`).
- **`improvise(prompt: String)`** → `Result<Null>`. Faculty: LLMStdlib. One-shot LLM fragment. Tag: (to build — spec'd `NotYetImplemented`).
- **`reflect_now()`** → `Result<Null>`. Faculty: LLMStdlib. Reflection pass. Tag: (to build — spec'd `NotYetImplemented`).
- **`wait_for_chat(timeout?: Float, from?: String)`** → `Result<…>`. Faculty: MemoryStdlib. Event waiter for incoming chat. Tag: (to build — spec'd `NotYetImplemented`).
- **`observe(target, ticks: Int)`** → `Result<…>`. Faculty: MemoryStdlib. Entity watcher. Tag: (to build — spec'd `NotYetImplemented`).

Spec entries marked `NotYetImplemented` (also `mine`/`fish`/`chop`/`cook` among the ambient verbs) validate fine but execute as a `NOT_IMPLEMENTED`-failing stub (`makeStub` in `runtime/dsl_actions.go`) — the routine sees a typed `Err`, never a crash.

#### Persona reads (identity state, no bang)

- **`mood()`** → emotional state. Faculty: PersonaRead. Tag: (to build — spec'd `NotYetImplemented`).
- **`motivation()`** → goal state. Faculty: PersonaRead. Tag: (to build — spec'd `NotYetImplemented`).

#### Admin / test (fenced, test-harness only)

- **`command(cmd: String)`** → `Result<Null>`. Faculty: PrimaryAction (test-harness only). Admin commands (`setstat`, `item`, `teleport`, etc.). Never appears in persona-authored behavior. Tag: (exists). (Also listed under ambient verbs as the player-facing surface; here noted as the fenced admin escape hatch.)

#### `control` — notes

- This entire plane is **explicitly outside the GUI-equivalence rule (§3)** so the body surface stays honest and freezable.
- Flow constructs are grammar, not callables — they have no `dsl/spec/actions.go` entry; they live in the parser/validator.
- The brain/memory verbs are real in production (mesa-backed Strategist/Retriever/memory manager); the deterministic stubs remain the test/offline defaults. The remaining `(to build)` entries are spec'd `NotYetImplemented`.

## 9. Sanctioned aliases (the only second-names)

- `attack(t)` → `combat.attack(t)`
- `cast(spell, target?)` → `magic.cast(...)`

## 10. Migration map (flat → frozen)

> **Executed.** This map is now the historical rename ledger (the namespace
> verb tables in `runtime/dsl_actions.go` cite it). The flat left-column
> names are gone from `dsl/spec/actions.go`, except the two sanctioned §9
> aliases (`attack`, `cast`) and the ambient verbs that kept their names.

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

**Duel** — the old `confirm_duel()` was a single toggle for both screens (the
same bug already split out of trade). The freeze split it to match trade:

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
