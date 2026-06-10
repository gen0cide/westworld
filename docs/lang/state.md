# State — The IF THIS Query Layer

> **STATUS — BUILT. Verified 2026-06-10 against HEAD `0bfa818`
> (branch `tidy/structure-and-docs`).** The query layer this doc once
> planned (~100 accessors) shipped. `dsl/spec/accessors.go` carries
> **114 spec'd paths**, 8 of them `NotYetImplemented` (the five
> `host.*` rows plus `self.prayer.active`, `self.spells.known`,
> `inventory.weight`). Implementations live in the `runtime/views_*.go`
> family — one file per root (`views_self` / `views_world` /
> `views_inventory` / `views_combat` / `views_trade` / `views_bank` /
> `views_duel` / `views_magic` / `views_prayer` / `views_shop`), plus
> `views_entities` / `views_records` / `views_scenery` (entity views,
> recent-event records, locs/boundaries — split out of `views_world.go`
> post-verify in c67df24) and the worn-equipment views in
> `runtime/equipment.go`. Every table below was re-checked against
> those files on the verify date. Open query-layer work lives in
> [`../TODO.md`](../TODO.md) — see DSL-2, DSL-5, DSL-12, DSL-13,
> DSL-20, DSL-22.

> **Host** — an autonomous AI actor. This doc enumerates every
> piece of state a host can read about itself and its world.

The query layer is the foundation everything else stacks on. Every
piece of state that's visible in the RSC client UI — and anything
a thinking human would derive from it — must be a query. Queries
are cheap, side-effect-free, and always reflect current state at
the moment of access.

If a routine can't ask the question "what is X right now?", that's
a missing query and we need to add it. If a routine needs to call
an action to learn something, that's a smell — actions cost
packets, blocking time, and risk failure. Queries are free.

## Reserved names

Eleven namespace roots are reserved — the canonical list is
`spec.ReservedRoots()` in `dsl/spec/spec_roots.go`, which the
validator, the runtime, and this doc all derive from. Ten are
wired as live views in `runtime/dsl_bridge.go` (the
`it.Reserved[...]` block); `host` is reserved **ahead of being
built** so no routine can bind the name. All views are read-only —
mutations happen via the action layer (verbs are documented in
[`actions.md`](actions.md); several live ON these roots as
namespaced callables, e.g. `bank.deposit`, `magic.cast`).

| Name | Returns | Backing file |
|---|---|---|
| `self` | First-person vitals, skills, equipment | `runtime/views_self.go` |
| `world` | Visible entities, recent-events buffer, static locs, dialog, boundaries, dynamic scenery | `runtime/views_world.go` |
| `inventory` | Inventory slots and predicates | `runtime/views_inventory.go` |
| `combat` | Engagement state + combat verbs | `runtime/views_combat.go` |
| `trade` | Trade-handshake state + verbs | `runtime/views_trade.go` |
| `bank` | Bank-UI state + verbs | `runtime/views_bank.go` |
| `duel` | Duel-handshake state + rules + verbs | `runtime/views_duel.go` |
| `magic` | Spell catalog + level/rune gates + `cast` | `runtime/views_magic.go` |
| `prayer` | Prayer catalog + active set + toggles | `runtime/views_prayer.go` |
| `shop` | Shop-UI state + buy/sell | `runtime/views_shop.go` |
| `host` | Persona / identity layer — **NOT BUILT** (see below) | spec rows only, all `NotYetImplemented` |

The `trade` / `bank` / `duel` / `dialog` views are also reachable
as `world.trade` etc. (back-compat; `worldView.Get` routes them).
`shop` is top-level only — `worldView.Get` has no "shop" case
(the routing note in `views_shop.go`'s header is stale on this).

Naming convention reminder (see [`syntax.md`](syntax.md)):

- Booleans prefix `is_` → `is_busy`, `is_in_combat`
- Maximums prefix `max_` → `max_hp`, `max_prayer`
- Counts no prefix → `free`, `length`, `level`
- Error codes SCREAMING_SNAKE → `PATH_BLOCKED`

## Collection methods

Entity lists are `interp.List` values; the interpreter synthesizes
their methods in `dsl/interp/interp.go` (`evalMember`):

```
world.npcs.length
world.npcs.filter(n => n.is_attackable)        # → list
world.npcs.map(n => n.name)                    # → list
world.npcs.find(n => n.name.lower == "cook")   # → first match or null
world.npcs.nearest(self.position)              # → closest by Chebyshev, or null
world.npcs.first / .last                       # → element or null
world.npcs.random                              # → seeded anti-sameness pick, or null
```

There is **no** `nearest_by_name` and **no** `.contains` on lists
(the old design had both) — compose `.filter(...)` (or `.find`) with
`.nearest(pos)` instead. `.nearest` accepts anything with `.x`/`.y`
(a position, another entity) or two Int args, and skips elements
without coordinates.

## `self` — first-person host state

### Identity + vitals

| Accessor | Type | Notes |
|---|---|---|
| `self.name` | String | The username the runtime logged in with (alias `.username`) |
| `self.position` | Position | `.x`, `.y`, `.plane` (plane = wire-Y / 944: 0 ground, 1+ upstairs, 3 underground) |
| `self.hp` | Int | Current hits (boostable) |
| `self.max_hp` | Int | Base hits skill level |
| `self.hp_fraction` | Float | `hp / max_hp`, 0.0–1.0 |
| `self.prayer` | Int | Current prayer points |
| `self.max_prayer` | Int | Base prayer skill level |
| `self.fatigue` | Int | Normalized 0–100 percent (the raw wire value is 0–750) |
| `self.combat_level` | Int | Derived from melee skills |
| `self.quest_points` | Int | |
| `self.is_busy` | Bool | **STUB — always false** (needs a per-host current-action registry) |
| `self.is_in_combat` | Bool | Real — delegates to `combat.engaged` |
| `self.is_sleeping` | Bool | Real — mirrors SEND_SLEEPSCREEN/SEND_STOPSLEEP; the cradle auto-answers the captcha so this is usually a brief flicker |
| `self.last_death_at` | Position? | Death tile — captured before the respawn packet overwrites position, so reading it in an `on death` handler gives where you died. Null if never died |
| `self.death_count` | Int | |
| `self.wielded` | ItemView? | First Wielded inventory slot; null if unarmed |

### Skills

Each skill exposes the same shape (`skillView` in
`runtime/views_self.go`):

| Field | Type | Notes |
|---|---|---|
| `level` | Int | Current (boostable) level |
| `max_level` | Int | Base level |
| `xp` | Int | Total experience |
| `xp_to_next_level` | Int | XP remaining until next level (canonical RSC XP table, precomputed) |
| `percent_to_next_level` | Float | 0.0–1.0 progress |
| `name` / `id` | String / Int | The skill's own name + catalog index |

Skill names (canonical 18, matching the RSC catalog; `hitpoints` is
accepted as an alias for `hits`):

```
attack, defense, strength, hits, ranged, prayer, magic,
cooking, woodcutting, fletching, fishing, firemaking,
crafting, smithing, mining, herblaw, agility, thieving
```

```
self.skills.fishing.level
self.skills.fishing.xp_to_next_level
self.skills.attack.percent_to_next_level
```

A flat `self.skills.list()` iterator was planned but never built —
access is by name only.

### Equipment

`self.equipped` serves **two surfaces** on one value
(`equippedView`, `runtime/views_self.go`):

1. **List surface** — the currently-wielded inventory items:
   `.all` / `.length` / `.first` / `.last` / `[N]` / `for ... in`.
   `.filter`/`.map`/`.find` only exist on real lists, so route them
   through `.all`: `self.equipped.all.filter(i => i.is_wearable)`.
2. **Per-slot surface** — `self.equipped.weapon` / `.shield` /
   `.head` (aliases `hat`, `helmet`) / `.body` (alias `torso`) /
   `.legs` / `.gloves` (alias `hands`) / `.boots` (alias `feet`) /
   `.amulet` (alias `neck`) / `.cape` (alias `back`).

Per-slot accessors resolve from the host's **own inventory** (the
Wielded slots, matched to the slot by the item def's WearSlot —
`Host.selfWornGroup`, `runtime/equipment.go`), so they're exact.
Each returns an equip-slot view exposing `.name`, `.id`, `.def`
(full ItemDef from facts), `.slot`, `.slot_name`, `.is_empty`, and
`.sprite_id` (the raw worn-appearance sprite from the opcode-234
appearance mirror).

Note: `head`, `body`, and `legs` each span **two** underlying RSC
body-animation layers (the metal/style picks which — e.g. plate-mail
legs vs. a skirt, a medium vs. a large helm), and the accessor
returns whichever is worn. The exact-layer names are also available
directly: `large_helmet`, `med_helmet`, `platebody`, `chainbody`,
`platelegs`, `skirt`.

```
self.equipped.bonuses        # summed combat bonus of worn gear
                             #   ("equipment status")
```

`bonuses` is recomputed from the currently-wielded items and exposes
`.armour`, `.weapon_aim` (alias `.aim`), `.weapon_power`
(alias `.power`), `.magic`, `.prayer`.

### Prayers

Per-prayer-name accessors (`self.prayer.thick_skin.is_active`) were
never built. The real surface is the back-compat `self.prayers.*`
root plus the promoted top-level `prayer.*` root
(`runtime/views_prayer.go`):

```
self.prayers.active          # list of currently-active slot indices (Int)
self.prayers.count           # how many are on
self.prayers.is_active(N)    # Bool for one slot
self.prayers.book            # list of PrayerDef views
self.prayers.by_id(N)        # PrayerDef or null
self.prayers.by_name("Thick skin")   # PrayerDef or null

prayer.active(slot|name)     # Bool — Int slot or String name (case-insensitive)
prayer.active_list           # list<Int> (from opcode 206)
prayer.activate(p) / prayer.deactivate(p)   # action verbs
```

A PrayerDef exposes `.id`, `.name`, `.req_level`, `.drain_rate`,
`.description`. Slot indices follow the RSC prayer-book order
(0 Thick Skin … 13 Protect from Missiles, 14 Protect from Melee —
members). A "prayers I have the level for" convenience
(`self.prayer.available`) is not built — see TODO.md DSL-20.

### Magic

Per-spell-name accessors (`self.spells.wind_strike.*`), a
`.selected` spell, and `.castable` were never built. The real
surface (`runtime/views_magic.go`):

```
self.spells.count / .book               # full catalog of SpellDef views
self.spells.known                       # spells within BASE magic level
self.spells.by_id(N) / .by_name("Wind strike")
self.spells.has_runes_for(spell)        # inventory-only rune check

magic.level / magic.max_level           # current (boosted) / base Magic level
magic.known                             # spells within CURRENT (boosted) level
magic.can_cast(spell)                   # Bool — level-only gate, no rune check
magic.has_runes_for(spell)              # Bool — an equipped elemental staff
                                        #   satisfies its element's rune
magic.cast(spell, target?)              # the one polymorphic cast verb
```

A SpellDef exposes `.id`, `.name`, `.req_level`, `.type`, `.exp`,
`.description`, `.members`, `.evil`, `.runes` (list of
`[item_id, count]` pairs). Spell refs are Int id, String name
(case-insensitive), or a SpellDef view.

## `host` — persona / identity layer

**DESIGN — NOT BUILT.** The `host` root is reserved in
`spec.ReservedRoots()` *ahead* of being wired (so no authored
routine can bind `host = ...` that would break the day the view
lands), and its five accessor rows in `dsl/spec/accessors.go` are
all `NotYetImplemented`. `runtime/dsl_bridge.go` registers no view
for it; reading `host.*` fails at runtime today. The validator also
still rejects `extends host` / `super()` handlers — that whole
reflex tier is TODO.md **DSL-13** (and `host.idle_ticks` is
**DSL-12**).

The intended read surface, kept here as the live spec:

| Accessor | Notes |
|---|---|
| `host.name` | The host's RSC username (use `self.name`, which is live) |
| `host.persona.*` | Trait values; full set in `docs/personas.md` |
| `host.mood` | Emotional-state weights summing to 1.0 |
| `host.motivation.north_star` | String — persona's long-term goal |
| `host.motivation.current_focus` | String — current sub-goal |
| `host.defaults.<event>` | Parent handler reference for `extends host` |

Most routines will barely touch `host.*`. It exists primarily so
persona-tier defaults and traits are introspectable from
expressive routines (e.g. a brave host's combat routine might
ignore HP thresholds that a timid host's wouldn't).

## `world` — what's visible

### Visible entities

```
world.players          # player-views — NOTE: includes the host's own record
world.npcs             # npc-views in view radius
world.ground_items     # ground-item-views in view radius
world.scenery          # DYNAMIC scenery (see below)
```

`world.players` and `world.npcs` return plain lists (all collection
methods above). `world.ground_items` and `world.scenery` are richer
views that also iterate/index like lists:

```
world.ground_items.by_id(N)            # nearest visible item with id N, or null
world.ground_items.by_id(N, radius=R)  # same, within R tiles
world.ground_items.nearest             # closest to self (bare field), or null
world.ground_items.nearest(pos)        # closest to an explicit position
world.ground_items.most_valuable       # highest base-value visible item, or null
world.ground_items.all / .length / [N]

world.scenery.by_id(97)                # nearest visible fire (def 97)
world.scenery.by_id(97, radius=2)
world.scenery.nearest / .all / .length
```

`world.scenery` mirrors SEND_SCENERY_HANDLER (opcode 48) — the
**only** place runtime-spawned scenery surfaces. A fire you just lit
never appears in the static facts map, so `world.locs.search("fire")`
finds only pre-placed fires; use `world.scenery` for one you created.

### Boundaries and dialog

```
world.boundaries.at(x, y, dir)         # boundary view (synthesized even if
                                       #   facts doesn't know the tile)
world.boundaries.near(radius)          # nearest-first openable doors/gates
world.boundaries.is_open(x, y, dir)    # observed-open via SEND_BOUNDARY_HANDLER
world.boundaries.dynamic               # all dynamic overrides, for debugging

world.dialog.is_open                   # NPC option menu up?
world.dialog.options                   # list<String>
world.dialog.find_option("yes")        # 1-based index (pairs with answer()), 0 if absent
world.dialog.clear()                   # reset after answering
```

### Static world knowledge (facts-derived)

```
world.locs.banks         world.locs.trees
world.locs.altars        world.locs.rocks
world.locs.furnaces      world.locs.doors
world.locs.anvils        world.locs.ladders
world.locs.fishing_spots world.locs.shops
world.locs.scenery       world.locs.spawn_points
world.locs.search("furnace")           # generic substring search
```

Each returns a lazily-resolved loc-list (`locListView`):

```
world.locs.banks.nearest(self.position)   # → placement or null (radius-64
                                          #   scan via facts.NearestByName)
world.locs.fishing_spots.within(50)       # → list of placements ≤ 50 tiles
                                          #   of self (optional 2nd arg: center)
world.locs.banks.names                    # → matching def names
```

`.length` counts matching def *names*, not placements — prefer
`.nearest()`/`.within()`. Name search prefers scenery over NPC defs
(discovered live 2026-05-28: "bank" matched both the Bank booth and
the Banker NPC, and routines walked to the NPC spawn instead of the
counter). Static knowledge comes from OpenRSC config loaded at
startup — no runtime cost beyond the scan.

### Map perception (world-map gazetteer)

The host carries a world-map gazetteer of named places and typed
points-of-interest, so it can orient itself the way a player reading
the map would. The query-facing entry points are control-plane
verbs documented in [`actions.md`](actions.md): `where_am_i()` (name
the current region), `where_is(name)` (locate a named place),
`bearing_to(x, y)` (compass direction + distance to a tile), plus
the deliberate map-study verbs `search_map` / `reachable` /
`survey_map` (`runtime/actions_worldmap.go` — each charges in-world
"study the map" seconds). The `look_around` action leads with
location. They live on the control plane, not on `world` — there is
**no** `position.region_name` field.

### Recent-events buffer (read-only)

The single most-recent of each transient kind, or null if never
observed this session (`world/recent.go`):

```
world.last_chat            # → {speaker, message, at, contains(s)} or null
world.last_pm              # → {sender, message, at, contains(s)} or null
world.last_damage          # → {amount, source, at} or null
world.last_server_message  # → {message, at, contains(s)} or null
world.last_dialog_text     # → {text, at, contains(s)} — NPC speech bubbles
world.messages             # → bounded server-message LOG, oldest-first
```

`world.messages` is a real ring (cap 32 — `ServerMsgRingCap`,
`world/recent.go`); each Message carries `.text` (alias `.message`),
`.kind`, `.at`, and `.contains(needle)` (case-insensitive). The
`on message(text)` event fires per new entry. `.at` fields format as
`"15:04:05"` strings.

**Incoming only.** These records are written by `world.World.Apply`
on inbound events (`ChatReceived` / `OtherPlayerChat` /
`PrivateMessage` — `world/world.go`); own `say(...)` calls are
outgoing packets and never touch the buffer. Same for `last_pm`
(outgoing `tell(...)` does not record). This is intentional:
routines react to others, not their own utterances. To confirm an
outgoing message was sent, use the action's return value (`say`
returns `{val: null}` on success, an error otherwise).

## `inventory` — what I'm carrying

| Accessor | Returns | Notes |
|---|---|---|
| `inventory.free` | Int | Empty slot count |
| `inventory.used` | Int | `30 - free` |
| `inventory.capacity` | Int | 30 (RSC convention) |
| `inventory.is_full` | Bool | |
| `inventory.slots` | List | InvSlot item-views (fields below) |
| `inventory.has(item)` | Bool | ItemRef = id, name, or item-view |
| `inventory.count(item)` | Int | Total amount across slots |
| `inventory.find(item)` | InvSlot or null | First matching slot |
| `inventory.find_all(item)` | List | Every matching slot (non-stackables occupy several) |
| `inventory.find_any([a, b, ...])` | InvSlot or null | First slot matching ANY ref — collapses gem/food/axe or-chains; also varargs |
| `inventory.slot_of(item)` | Int or null | First matching slot index |

`inventory.weight` is spec'd but `NotYetImplemented`.

## `combat` — current engagement

| Accessor | Type | Notes |
|---|---|---|
| `combat.engaged` | Bool | We have a resolvable target, OR someone is firing at us. (The name is `engaged`, not the planned `is_engaged`.) |
| `combat.target` | Npc/Player view or null | Wire-observed engagement on our own slot, falling back to last-attack intent (authentic v235 melee carries no projectile signal). Clears to null when the engaged NPC is observed at 0 hp |
| `combat.style` | String | "controlled"/"aggressive"/"accurate"/"defensive" — write-through mirror of `combat.set_style` (RSC sends no echo); defaults "controlled" |
| `combat.last_npc` | Npc view or null | Most-recently-attacked NPC, resolved live; null once it dies/leaves view |
| `combat.last_player` | Player view or null | Same for players (duels) |

Read target health on the returned view: `combat.target.hp_fraction`
/ `.health`. The planned `last_damage_dealt`/`last_damage_taken`
accessors were never built — use `world.last_damage` and the target's
health fields. Verbs (`combat.attack` / `set_style` / `retreat` /
`retreat_to`) are in [`actions.md`](actions.md).

## `bank` (when bank UI is open)

| Accessor | Notes |
|---|---|
| `bank.is_open` | Bool — state machine over opcodes 42/203 |
| `bank.slots` | List of `[item_id, amount]` **pairs** (not named views) |
| `bank.used` / `bank.max_size` / `bank.free` | Occupied / capacity / empty slots (all 0 when closed) |
| `bank.has(id)` / `bank.count(id)` | Int total of that item id — **Int-only today**, names don't resolve (TODO.md DSL-22) |

Verbs: `open` / `deposit` / `withdraw` / `deposit_all` /
`withdraw_all` / `withdraw_x` / `close` — see actions.md.

## `trade` (when trading)

| Accessor | Notes |
|---|---|
| `trade.is_active` | Bool |
| `trade.phase` | "none" / "request_sent" / "open" / "confirm" / "completed" / "cancelled" |
| `trade.with` | Opponent name (String) or null — **not** the planned `trade.opponent` |
| `trade.with_index` | Opponent server index |
| `trade.my_offer` / `trade.their_offer` | Lists of `[item_id, amount]` pairs — **not** the planned `mine`/`theirs` |
| `trade.accepted` / `trade.they_accepted` / `trade.both_accepted` | Offer-screen (screen 1) accepts |
| `trade.confirmed` / `trade.they_confirmed` / `trade.both_confirmed` | Confirm-screen (screen 2) accepts |

The `my_first_accepted` / `their_second_accepted` etc. spellings are
kept as back-compat aliases. Verbs: `request` / `offer` / `accept` /
`confirm` / `decline`.

## `duel` (when dueling)

Same shape as `trade` (`is_active` / `phase` / `with` / `with_index`
/ `my_offer` / `their_offer` / the accepted+confirmed families),
plus the four rule toggles: `duel.disallow_retreat` /
`.disallow_magic` / `.disallow_prayer` / `.disallow_weapons`.
Verbs: `request` / `set_rules` / `stake` / `accept` / `confirm` /
`decline`.

## `shop` (when a shop window is open)

| Accessor | Notes |
|---|---|
| `shop.is_open` | Bool |
| `shop.is_general` | Bool — general store? |
| `shop.slots` | List of `[item_id, stock, base_stock]` triples |
| `shop.stock(item)` | Int units in stock (0 if absent/closed) |
| `shop.price(item)` | Int unit buy price in gp |

Verbs: `buy` / `sell` / `close`.

## Entity-views

These are the shapes returned by collection accessors and entity
lookups. Each is a snapshot — fields don't auto-update; re-query
to refresh. Per-entity-view fields are documented **here, not as
`spec.Accessors` rows** (they aren't rooted at a fixed top-level
path) — `dsl/spec/accessors.go` delegates to this section, and the
view implementations are the canonical source:
`runtime/views_entities.go` (player / npc / ground-item),
`runtime/views_scenery.go` (placement / boundary), and
`runtime/equipment.go` (worn-slot views + threat bands). New fields
land there + here. (Adding spec rows for these is TODO.md DSL-2.)

### `player-view`

```
.index            # server-assigned int
.name             # string
.x, .y            # absolute world coords
.position         # Position
.combat_level     # from the appearance packet; null until seen
.relative_level   # combat_level − mine (null until seen)
.threat           # string — danger relative to me (null until seen)
.threat_colour    # the @col@ tag for that threat (null until seen)
.is_skulled       # bool — wilderness skull on the player
.hp_fraction      # 0–1 — null unless a damage update has landed (PVP)
.health           # int current hits, same gate
.equipment        # per-player equipment view (see below)
.is_friend        # bool — STUB, always false (friend-list mirror pending, DSL-20)
.in_combat_with   # STUB, always null (per-player target tracking pending, DSL-20)
```

`.<slot>` accessors (`.helmet`/`.weapon`/`.shield`/`.body`/`.legs`/
`.gloves`/`.boots`/`.amulet`/`.cape`/`.head` + the aliases listed
under self.equipped) are exposed directly on the player-view and on
`.equipment`. They resolve from the player's appearance packet. Each
slot is a worn-item view: `.name`, `.id`, `.def`, `.slot`,
`.slot_name`, `.is_empty`, `.ambiguous`, `.candidates`. Same-metal
melee weapons share a worn appearance, so a weapon slot can be
`.ambiguous` (`.id` null, multiple `.candidates`) — faithful to
what's visually distinguishable. Helmets and armour resolve exactly.
`.equipment` also exposes `.all` (list) and `.length`.

### `npc-view`

```
.index            # server-assigned int
.type_id          # int — index into facts.NpcDefs
.name             # ALWAYS a string ("" if the def hasn't loaded —
                  #   filter/find lambdas need no null-guards)
.x, .y, .position
.def              # NpcDef catalog view (id/name/description/combat_level/
                  #   max_hp/attackable/aggressive/command1/command2)
.combat_level     # (atk+str+def)/4 + hits/4, from facts
.relative_level   # combat_level − mine
.threat           # string — danger relative to me
.threat_colour    # the @col@ tag for that threat
.max_hp           # from facts
.hp_fraction      # 0–1 — null unless we've fought it (only the engaged
                  #   target gets health updates on the wire)
.health           # int current hits, same gate
.is_attackable    # from facts
.is_aggressive    # from facts
.in_combat_with   # STUB, always null
```

**`threat` / `threat_colour`** (on both player- and npc-views)
express how dangerous an entity is *relative to the host* — the
authentic RSC cue the client paints as the level's colour (server
`Formulae.getLvlDiffColour`; "darker red = more dangerous"). The
host can't read a UI colour, so we surface the concept
(`threatBand`, `runtime/equipment.go`). The bands, from most to
least dangerous: `deadly`, `very dangerous`, `dangerous`, `risky`,
`even`, `favourable`, `easy`, `very easy`, `trivial`.

### `ground-item-view`

```
.id               # item id (alias .item_id)
.name             # looked up from facts
.x, .y, .position
.def              # ItemDef catalog view
.is_mine          # bool — STUB, always false (the 3-min ownership window
                  #   is not tracked, DSL-20; pick_up falls through to
                  #   "try and see" with SERVER_REJECTED)
```

There is **no** `.amount`/`.quantity` — the world mirror's ground-item
record carries no stack size (TODO.md DSL-20).

### `item-view` / InvSlot (in inventory, or via self.wielded)

Frozen instance shape (api.md §2): `.idx` (slot index, null when
unknown), `.quantity`, `.def` (ItemDef). Back-compat flat fields are
retained: `.id`, `.amount`, `.name`, `.is_stackable`, `.is_wearable`,
`.is_wielded`, `.is_members_only`, `.command`. (Note: the flat id
field is `.id` — inventory slots have no `.item_id` alias.)

An **ItemDef** (via `.def` on inventory items, ground items, and
exact worn slots) exposes `.id`, `.name`, `.description`,
`.command`, `.stackable`/`.is_stackable`, `.wearable`/`.is_wearable`,
`.members`/`.is_members_only`, `.tradable`/`.is_tradable`,
`.edible`/`.is_edible` (command == Eat/Drink heuristic).

### `placement` (from `world.locs.*` and `world.scenery`)

```
.kind             # "scenery" | "boundary" | "npc_spawn"
.def_id
.name
.x, .y, .position
.direction        # 0–7 if boundary
```

### `boundary-view` (from `world.boundaries.*`)

```
.x, .y, .position
.direction        # alias .dir
.name             # facts lookup at the tile; defaults "door"
.door_type        # BoundaryDef.DoorType (0 when def unknown)
.is_openable      # doors/doorframes; unknown def assumed openable
.blocks_when_closed  # closeable doors (DoorType==1) — open before crossing
```

### `position`

```
.x
.y
.plane            # int — wire-Y / 944: 0 ground, 1+ upstairs, 3 underground
```

No `.region_name` — region naming went to the gazetteer verbs
(`where_am_i()`), not a position field.

## Performance notes

All queries are cheap and side-effect-free. The cost model:

- **Field access** (`.x`, `.hp`) — single map / struct lookup.
  O(1), nanoseconds.
- **Collection iteration** (`world.npcs`) — snapshot the
  underlying mirror's slice. O(n) where n is visible entities,
  usually <50.
- **Static facts lookup** (`world.locs.banks.nearest(p)`) — a
  radius-64 `facts.NearestByName` scan per matching def name.
  O(m), typically small; the facts data loads once at startup.
- **Recent buffer queries** (`world.last_chat`) — O(1) read of a
  small ring buffer.

There's no query that hits the network, an LLM, or even the disk.
Routines can read state freely. (The map-*study* verbs `search_map`
/ `reachable` / `survey_map` are the deliberate exception: they're
actions that charge in-world seconds — see actions.md.)

## Adding a query

Add a row in `dsl/spec/accessors.go` **and** a switch arm in the
relevant view's `Get(field string)` method in `runtime/views_*.go`
— the spec is the documentation + discovery + consistency source;
the views are the behavior. `dsl/spec/consistency_test.go` enforces
unique paths, reserved roots, and non-empty docs; the REPL's
`.accessors` meta-command (`runtime/repl.go`) lists every row with
its stub status. Tests for new fields go beside the view file
(`views_combat_test.go`, `views_self_equip_test.go`, ... are the
precedents).

## History — how the build plan landed

The original version of this doc was the build plan: a ~100-field
menu tracked as tasks #56–#65. All ten rows shipped (Phase 2.5
build-out, then the §10 namespace promotion, the #115 surface
freeze, and the #117/#118/#119 perception + bulk-verb + event
batches). Names drifted from the plan as the surface froze —
planned → shipped:

| Planned | Shipped |
|---|---|
| `combat.is_engaged` | `combat.engaged` |
| `combat.last_damage_dealt/taken` | `world.last_damage` + `target.health`/`.hp_fraction` |
| `trade.opponent` / `.mine` / `.theirs` | `trade.with` / `.my_offer` / `.their_offer` |
| `self.prayer.<name>.is_active/.level_req/.drain_rate` | `prayer.active(slot\|name)` + `prayers.by_name(s).req_level/.drain_rate` |
| `self.spells.<name>.*` / `.selected` / `.castable` | `magic.known` / `can_cast` / `has_runes_for` + `spells.by_name` |
| `list.nearest_by_name(s)` / `list.contains(x)` | `.filter(...)` / `.find(...)` composed with `.nearest(pos)` |
| `self.skills.list()` | never built — named access only |
| `position.region_name` | never built — gazetteer verbs instead |
| `inventory.weight` | never built (`NotYetImplemented` spec row) |
| "aliases are forbidden" | relaxed in practice — back-compat aliases exist (`hitpoints`, equipment-slot synonyms, `.aim`/`.power`, `.message`/`.text`); new accessors should still pick one canonical name |

The plan's four "open questions" all got de-facto answers in that
table (no `.list()`; gazetteer over region strings; `known` =
level-only with a separate rune check; `wielded` kept as the RSC
vernacular alongside `equipped.weapon`).

Remaining query-layer work is tracked in [`../TODO.md`](../TODO.md)
— notably DSL-2 (spec rows for the undocumented roots + entity-view
field tables), DSL-5 (`.distance`/`.is_reachable` view fields),
DSL-12 (`host.idle_ticks`), DSL-13 (the `host.*` reflex tier),
DSL-20 (de-stub `is_friend`/`is_mine`/`in_combat_with`, ground-item
quantity, `self.prayer.available`), and DSL-22 (name resolution in
`bank.has`/`count`).
