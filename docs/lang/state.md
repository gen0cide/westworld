# State — The IF THIS Query Layer

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

## Status

- **Implemented today**: ~15 accessors on `self`, `inventory`,
  `world`, `combat`. See `runtime/dsl_views.go` for what exists.
- **This doc**: the full plan — ~100 accessors organized by
  domain. **Most are not yet built.**
- **Implementation order**: vitals → skills → equipment →
  inventory → world entities → prayer → magic → recent-events
  buffers → combat target → bank/trade state.

## Reserved names

Four entities live in the routine's scope from startup. All are
read-only — mutations happen via the action layer.

| Name | Returns | Notes |
|---|---|---|
| `self` | First-person view of the host's vitals, skills, inventory references | The default for in-game state queries |
| `host` | Persona / identity layer | `host.persona.*`, `host.defaults.*`, and the parent in `extends host` chains |
| `world` | Visible NPCs, players, ground items, scenery; static facts via `world.locs.*` | Everything outside the host's body |
| `inventory` | Inventory slots and predicates | Separate top-level for ergonomics; same data as `self.inventory` |
| `combat` | Current combat state — target, engagement, recent damage | Read-only view of the combat subsystem |

Naming convention reminder (see [`syntax.md`](syntax.md)):

- Booleans prefix `is_` → `is_busy`, `is_in_combat`
- Maximums prefix `max_` → `max_hp`, `max_prayer`
- Counts no prefix → `free`, `length`, `level`
- Error codes SCREAMING_SNAKE → `PATH_BLOCKED`

## `self` — first-person host state

### Vitals (~12 fields)

| Accessor | Type | Notes |
|---|---|---|
| `self.position` | Position | `.x`, `.y`, `.plane` |
| `self.hp` | Int | Current hits (boosted) |
| `self.max_hp` | Int | Base hits skill level |
| `self.hp_fraction` | Float | `hp / max_hp`, 0.0–1.0 |
| `self.prayer` | Int | Current prayer points |
| `self.max_prayer` | Int | Base prayer skill level |
| `self.fatigue` | Int | 0–100 — RSC sleep mechanic threshold |
| `self.combat_level` | Int | Derived from melee skills |
| `self.quest_points` | Int | |
| `self.is_busy` | Bool | Currently performing an action |
| `self.is_in_combat` | Bool | Engaged with an NPC or player |
| `self.is_sleeping` | Bool | Sleep-screen up due to fatigue |

### Skills (18 × 5 fields ≈ 90 fields)

Each skill exposes the same shape:

| Field | Type | Notes |
|---|---|---|
| `level` | Int | Current (boostable) level |
| `max_level` | Int | Base level |
| `xp` | Int | Total experience |
| `xp_to_next_level` | Int | XP remaining until next level |
| `percent_to_next_level` | Float | 0.0–1.0 progress |

Skill names (canonical 18, matching the RSC catalog):

```
attack, defense, strength, hits, ranged, prayer, magic,
cooking, woodcutting, fletching, fishing, firemaking,
crafting, smithing, mining, herblaw, agility, thieving
```

Access:

```
self.skills.fishing.level                 # current
self.skills.fishing.xp_to_next_level
self.skills.attack.percent_to_next_level
```

Also a flat list for iteration:

```
self.skills.list()        # → [{name: "attack", level: 50, ...}, ...]
self.skills.list().filter(s => s.level >= 80)
```

### Equipment (worn slots, ~12 fields + derived)

```
self.equipped.weapon         # item-view or null
self.equipped.shield
self.equipped.head
self.equipped.body
self.equipped.legs
self.equipped.gloves
self.equipped.boots
self.equipped.cape
self.equipped.amulet
self.equipped.ring

self.equipped.style          # "accurate" | "aggressive" |
                             #   "defensive" | "controlled"
self.equipped.total_bonuses  # derived sum of all worn bonuses
```

Each non-null slot returns an item-view (see "Item-views" below).
`total_bonuses` exposes `.atk_stab`, `.atk_slash`, `.atk_crush`,
`.def_*`, `.str_bonus`, `.magic_bonus`, `.prayer_bonus`, etc.

### Prayer (active list + per-prayer state, ~14 × 4 = 56 fields)

```
self.prayer.active         # list of currently-active prayers
self.prayer.available      # list of prayers we have the level for
self.prayer.thick_skin.is_active        # per-name booleans
self.prayer.thick_skin.level_req
self.prayer.thick_skin.drain_rate
```

The 14 RSC prayers: `thick_skin`, `burst_of_strength`,
`clarity_of_thought`, `rock_skin`, `superhuman_strength`,
`improved_reflexes`, `rapid_restore`, `rapid_heal`,
`protect_items`, `steel_skin`, `ultimate_strength`,
`incredible_reflexes`, `paralyze_monster`, `protect_from_missiles`.

### Magic (spells + selected, ~50 spells × 4 fields)

```
self.spells.known            # list of spells you can cast
self.spells.selected         # currently-targeted spell or null
self.spells.wind_strike.level_req
self.spells.wind_strike.runes_required   # → [{name: "mind", count: 1}, ...]
self.spells.wind_strike.can_cast         # bool: have runes + level
```

### Wielded (convenience accessor)

```
self.wielded                 # item-view of currently-wielded weapon
                             # null if unarmed
                             # equivalent to self.equipped.weapon
```

## `host` — persona / identity layer

This is the read surface for the host's identity, separate from
moment-by-moment game state.

| Accessor | Notes |
|---|---|
| `host.name` | The host's chosen RSC username |
| `host.persona.shyness` | 0–1 trait values; full set in `docs/personas.md` |
| `host.persona.ambition` | |
| `host.persona.chattiness` | |
| `host.persona.aggression` | |
| `host.mood` | `{calm, anxious, excited, bored, frustrated}` weights, sums to 1.0 |
| `host.motivation.north_star` | String — persona's long-term goal |
| `host.motivation.current_focus` | String — current sub-goal |
| `host.defaults.<event>` | Parent handler reference for `extends host` |

Most routines barely touch `host.*`. It exists primarily so
persona-tier defaults and traits are introspectable from
expressive routines (e.g. a brave host's combat routine might
ignore HP thresholds that a timid host's wouldn't).

## `world` — what's visible

### Visible entities (lists)

```
world.players          # list of player-views in view radius
world.npcs             # list of npc-views in view radius
world.ground_items     # list of ground-item-views in view radius
world.scenery          # list of scenery-views in view radius
```

Each list supports the standard collection methods:

```
world.npcs.length
world.npcs.nearest_by_name("Goblin")           # → npc-view or null
world.npcs.nearest(self.position)              # → closest npc-view
world.npcs.filter(n => n.combat_level <= 30)
world.npcs.contains(target)
```

### Static world knowledge (facts-derived)

```
world.locs.banks
world.locs.altars
world.locs.furnaces
world.locs.anvils
world.locs.fishing_spots
world.locs.trees
world.locs.rocks
world.locs.doors
world.locs.ladders
```

Each returns a queryable list backed by `facts.NearestByName`:

```
world.locs.banks.nearest(self.position)        # → placement or null
world.locs.fishing_spots.within(50)            # list of placements ≤ 50 tiles
```

Static knowledge is computed from OpenRSC config — no runtime
cost beyond a map lookup.

### Recent-events buffer (read-only)

Short ring buffer of recent transient events the host has
observed. Useful for routines that want to react to "did someone
just say X?" without a dedicated handler.

```
world.last_chat            # → {speaker, message, kind} or null
world.last_pm              # → {sender, message} or null
world.last_damage          # → {amount, source, when}
world.last_server_message  # → {text, when}
```

## `inventory` — what I'm carrying

| Accessor | Returns | Notes |
|---|---|---|
| `inventory.free` | Int | Empty slot count |
| `inventory.weight` | Int | Sum across all slots |
| `inventory.is_full` | Bool | |
| `inventory.slots` | List | All slots, each with `.item_id`, `.amount`, `.is_wielded`, `.is_stackable`, `.name` |
| `inventory.has(item)` | Bool | Item ID, item-view, or name |
| `inventory.count(item)` | Int | |
| `inventory.find(item)` | Slot-view or null | First matching slot |
| `inventory.slot_of(item)` | Int or null | First matching slot index |

## `combat` — current engagement

| Accessor | Type |
|---|---|
| `combat.is_engaged` | Bool |
| `combat.target` | Entity (NPC or player) or null |
| `combat.style` | "accurate" / "aggressive" / "defensive" / "controlled" |
| `combat.target.hp_fraction` | Float, 0–1, if observable |
| `combat.last_damage_dealt` | Int |
| `combat.last_damage_taken` | Int |

## `bank` (when bank UI is open)

| Accessor | Notes |
|---|---|
| `bank.is_open` | Bool — bank interface currently shown |
| `bank.slots` | List of bank items with `.item_id`, `.amount`, `.name` |
| `bank.has(item)` | Bool |
| `bank.count(item)` | Int |

## `trade` (when trading)

| Accessor | Notes |
|---|---|
| `trade.is_active` | |
| `trade.opponent` | Player-view of the trading partner |
| `trade.mine` | List of items I've offered |
| `trade.theirs` | List of items they've offered |
| `trade.both_accepted` | Bool — second-stage confirm state |

## Entity-views

These are the shapes returned by collection accessors and entity
lookups. Each is a snapshot — fields don't auto-update; re-query
to refresh.

### `player-view`

```
.index            # server-assigned int
.name             # string
.x, .y            # absolute world coords
.position         # Position struct
.combat_level     # int (if observable from appearance)
.is_friend        # bool — host has added them as a friend
.in_combat_with   # entity (npc or player) or null
```

### `npc-view`

```
.index            # server-assigned int
.type_id          # int — index into facts.NpcDefs
.name             # string — looked up from facts
.x, .y, .position
.combat_level     # from facts
.max_hp           # from facts
.hp_fraction      # 0–1 if we're attacking them; null otherwise
.is_attackable    # from facts
.in_combat_with   # entity or null
```

### `ground-item-view`

```
.item_id
.name             # looked up from facts
.x, .y, .position
.amount           # for stackable items
.is_mine          # bool — true if owned by us (3-min ownership window)
```

### `item-view` (in inventory, bank, or equipped)

```
.item_id
.name             # looked up from facts
.amount
.is_stackable
.is_wielded       # for inventory items only
.def              # full ItemDef from facts (heal_amount, weight, etc.)
```

### `placement` (from `world.locs.*`)

```
.kind             # "scenery" | "boundary" | "npc_spawn"
.def_id
.name
.x, .y, .position
.direction        # 0–7 if boundary
```

### `position`

```
.x
.y
.plane            # int — 0=ground, 1=upstairs, etc. (or use floor?)
.region_name      # derived from facts loc lookups (e.g. "lumbridge")
```

## Performance notes

All queries are cheap and side-effect-free. The cost model:

- **Field access** (`.x`, `.hp`) — single map / struct lookup.
  O(1), nanoseconds.
- **Collection iteration** (`world.npcs`) — snapshot the
  underlying mirror's slice. O(n) where n is visible entities,
  usually <50.
- **Static facts lookup** (`world.locs.banks.nearest(p)`) — full
  scan over named placements. O(m) where m is named matches,
  typically small. Cached one-time at startup.
- **Recent buffer queries** (`world.last_chat`) — O(1) read of a
  small ring buffer.

There's no query that hits the network, an LLM, or even the disk.
Routines can read state freely.

## Open questions for tomorrow

- **`self.skills.list()` shape** — list of records vs map keyed
  by name vs both? Probably both: `.list()` returns a list, and
  named access (`.fishing`) works directly.
- **Region naming convention** — Do we surface region as a string
  (`"lumbridge"`), an enum, or a typed `Region` struct with
  bounds? Probably a string for now, upgrade later.
- **`spells.known` filtering** — Should it filter by "have runes
  to cast" or just "have the magic level"? I'd say magic level
  only, and add `.spells.castable` for "have runes + level".
- **Wielded vs equipped.weapon** — keep both as aliases, or
  collapse to one? `wielded` is the RSC vernacular; keep it.

## Building this

Implementation lives in `runtime/dsl_views.go`. Each entity view
is a Go struct implementing `interp.Getter` (member access) and
sometimes `interp.Indexer` (index access) or `interp.Callable`
(method-style access via `Getter` returning a callable).

The 100-field menu is big but mostly mechanical — each field is
a switch arm in a `Get(field string)` method. Estimated effort:
2-3 days of focused work, mostly typing.

The biggest sub-task is the static facts integration for
`world.locs.*` and the per-NPC/per-item enrichment (combat_level
from facts, heal_amount on food items, etc.). That logic already
exists in `facts/` and `runtime/examine.go`; the work is wiring
it through the view structs.
