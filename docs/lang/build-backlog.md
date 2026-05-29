# Body-layer build backlog (centralized)

Consolidates every scattered gap note — `scenarios_proposed.yaml` &
`scenarios_bots_proposed.yaml` trailing blocks, `protocol.md` open questions,
`writing-routines.md` §10, the `api.md` §8 `(to-build)` tags, and task #113 —
into one burn-down list. **Goal:** freeze `api.md`, then close these so the
body API is complete + stable before cognition/brain/persona.

Status legend: **DONE** (shipped this cycle) · **BUILD** (a real body faculty
we're missing) · **RESEARCH** (wire/behavior to verify first) · **REFACTOR**
(rename to the frozen surface) · **MIND-OUT** (a policy, belongs to
cognition/persona — explicitly *not* a body faculty) · **CONTENT** (scenario
data, not engine).

---

## DONE this cycle (closed — do not re-open)
- `pickpocket`/`npc_command` verb (NPC command1). · dialog `answer` 1-based +
  clear-between-menus (chained menus). · inventory mirror: opcode 123 =
  remove-and-shift (`wipeinv`). · trade split `confirm`/`finalize`. ·
  `interact_at(view)` (resolvePoint). · facts-on (name resolution). · setstat
  level-first + `setcurrentstats`. · cooking off the quest-gated range.

## 0. FREEZE + namespacing REFACTOR — gates everything
- **REFACTOR** Ratify `api.md` → freeze, then migrate flat→namespaced per the
  §10 map: `trade.*`/`bank.*`/`duel.*` promoted to roots; `magic.cast`,
  `prayer.activate`, `combat.attack`; `world.last_server_message →
  world.messages`. Touches the 88+195 scenarios + routines + guide. Mechanical
  but large; do it once, scripted. **Blocks** clean authoring of everything below.
- **REFACTOR** Def/Instance view types (`InvSlot{idx,def:ItemDef,quantity}`,
  `Npc{def:NpcDef,…}`, `Loc{def:LocDef,…}`, `GroundItem{def:ItemDef,…}`).

## 1. PROTOCOL / wire (`protocol.md`)
- **RESEARCH→BUILD** Decode the combat/health fields carried inside opcode 234
  (`InSendUpdatePlayers`). **This is the prerequisite** for the §3 combat
  perception faculties (`combat.target`, `Npc.health`, `in_combat`) — they're
  stubs *because* 234 isn't fully decoded.
- **RESEARCH** Bitpack field order/widths for PlayerCoords / UpdatePlayers
  (affects player/npc perception fidelity).
- **RESEARCH** Boundary direction-byte encoding (N/E/S/W) — needed for robust
  `world.boundaries.at` / door interaction.
- **RESEARCH** Opcode 50 disambiguation (`CAST_ON_NPC` vs `NPC_USE_ITEM` by
  length/context).
- **RESEARCH** (lower) XTEA/ISAAC key derivation confirm; dialog color-tag
  format (222); appearance payload (59→235); PM icon→rank mapping. (protocol.md
  completeness; not body-API-blocking.)

## 2. SESSION
- **DONE** LogoutGraceful, `-reset-on-exit`, code-4 login retry. (No open body work.)

## 3. WORLD / PERCEPTION (accessors a GUI player has, we stub/lack)
- **BUILD** `combat.target` (Npc|Player|Null), `combat.engaged`/`self.is_in_combat`,
  `Npc.health` (0–1 of *your target*) — currently stubs; perception gaps. *Dep: §1 opcode-234.*
- **BUILD** `self.combat_style` getter (set is write-only). *Unblocks combat-style xp tests.*
- **BUILD** `self.equipped` per-slot accessor (only `.wielded`/list now). *Unblocks "is a bow equipped".*
- **BUILD** `prayer.active(slot)`/active list (NotYetImplemented). *Unblocks prayer tests.*
- **BUILD** `magic` spellbook: `known`/`has_runes_for` (rename from `self.spells.*`).
- **BUILD** `bank.is_open` + `bank.slots` (NotYetImplemented). *Unblocks bank-balance assertions.*
- **BUILD** `world.shop` view (`is_open`/`stock`/`price`). *Unblocks shop tests (with §4 verbs).*
- **BUILD** `world.ground_items.nearest(pos)` / value-sorted (only `by_id`). *Unblocks loot-most-valuable.*
- **BUILD** `self.position.plane` (floor; RSC encodes upper floors as a coord offset). *Unblocks ladder/floor tests.*
- **BUILD** `inventory.find_any([ids])` multi-id find (convenience; collapses gem-loop or-chains).

## 4. ACTION (verbs a GUI player can do, we lack)
- **BUILD** `shop.buy(ItemRef,qty)` / `shop.sell(ItemRef,qty)` — InterfaceShopHandler-backed. *Highest-value money-making unblock.*
- **BUILD** `bank.deposit_all(keep?)` — **wire-only**, `Host.DepositAll` already exists. (Cheapest win.)
- **BUILD** `bank.withdraw_all` / `withdraw_x` preset.
- **BUILD** `combat.retreat()`/`flee` — break-melee (verify GUI-doable = walk away).
- **BUILD** `examine(target)` — output is a `Message`; the canonical perception self-test.
- **BUILD** `bury(item)` alias of `use_inventory_default`.
- **RESEARCH** `use(key, boundary)` key-on-door opcode (only knife-on-web confirmed) + door-locked message read (via `world.messages`).
- **RESEARCH** firemaking `use(tinderbox, logs)` path (returned "Nothing interesting"; §10 fire-cooking gap).
- *(Not verbs: `autocast`, `high_alch`/`low_alch` → use `magic.cast(spell,target)`; counted loop → hand-roll.)*

## 5. EVENTS (parameterize + add)
- **BUILD** `on xp_gain(skill)` (parameterize — replaces per-skill variants).
- **BUILD** `on message(pattern)` (message-pane reactions; pairs with `world.messages`).
- **BUILD** `npc_killed`/`target_died` (replace `combat.target==null` polling; *dep: §1*).
- `on item_gained`/`item_lost`, `bank_opened` — exist; fold into the parameterized scheme.

## 6. COGNITION (fenced control-plane, evolvable)
- **BUILD** `resolve(text, kind?) -> List<Match>` + the host learned-alias store
  (fuzzy → brain → write-back). Subsumes `spell_by_name`. *The recognition faculty.*

## 7. MIND-OUT (do NOT build — cognition/persona policies)
- auto-eat / HP-threshold config · timed prayer-flick · ammo-low watcher ·
  session/kill-count/time-box budget · autocast loop. These compose above the
  body API from `when`/`select`/loops.

## 8. CONTENT / scenario fixes (data, not engine)
- **CONTENT** `combat-prayer-protect-from-melee` — misnomer (no melee-protect prayer in RSC); repoint to a real prayer.
- **CONTENT** `smithing-iron-scimitar` — scimitar isn't a smith-menu item; repoint to a real iron weapon.
- **CONTENT** verify ids: gold-ore `152`, bucket `52`/`53`, agility def-`655` name; Druidic-Ritual gate via `completeallquests`.
- **RESEARCH** pottery-wheel def name/coords (unresolved by name search).
- **SKIP** jewellery-teleport (ring of dueling / amulet of glory — RS2 items, absent in RSC; already replaced with enchant).
- **NOTE** the bot-agent "CONTENT FIX: `cast_on_npc(rat,0)`" is **wrong** — the name-string form is verified live; `combat-magic-attack-on-rat` is flaky (cast timing), not mis-signatured.

## 9. Cross-doc reconciliation (other `docs/` TODOs folded in)
- **`primitives-backlog.md`** — mostly shipped (Tiers 1–2). Reconcile: `repeat_until`
  is now **DONE** (the doc still says DEFERRED #85); its **Tier-3 "defer" items
  (Prayer / Magic / Shop APIs) are now PROMOTED** — the live-test scenarios are
  the routines that need them (→ §3/§4 above). `host.idle_ticks` (AutoRune `%IdleC`)
  — **BUILD, low-pri** self/meta accessor (ticks since last meaningful action).
  Its principle "convenience verbs live in routine-land, not the DSL" *confirms*
  our §7 MIND-OUT line.
- **Jitter / per-reverie variance** (`primitives-backlog.md` §Jitter) — a **Phase-4
  persona-runtime** layer (wait-jitter, npc-pick randomization, walk-corner skew,
  injected invisibly per seed). Not body burn-down; flagged here so it isn't lost.
- **`#93` `super()` / per-handler `extends host`** — DSL construct, **deferred to
  Phase 4** (persona tier); not a body faculty, leave parked.
- **REPL extensions (`#54`)**, **observability "pending chat queue"** — polish,
  non-blocking; parked.
- **`personas.md` / `reveries.md`** — the deferred Phase-3 design session (already
  a memory); out of this body-layer burn-down.

## 10. §8-audit spec/impl drift (reconcile during the #115 refactor)
The per-namespace audit (api.md §8 generation) verified the surface against
source and flagged concrete drift to fix while refactoring:
- **inventory:** runtime `itemViewVal` exposes flattened `{id, amount, name,
  is_stackable, is_wearable, is_members_only, command}` — must become
  `InvSlot{idx, def: ItemDef, quantity}` per §2. `find_all` not yet implemented.
- **world:** `GroundItem.quantity` **missing** in impl (build it); `world` entity
  stubs `hp_fraction`/`is_friend`/`is_mine`/`in_combat_with` (perception gaps).
- **trade:** phase-enum values + acceptance field names **drift** from the frozen
  spec — align during the namespacing refactor.
- **combat:** `style` is write-only (build the read getter); `target`/`engaged`
  stubs (dep #116); `last_npc`/`last_player` already implemented.
- **magic:** five `cast_on_self/npc/player/land/item` exist → unify to
  `magic.cast(spell, target)` (spec'd `NotYetImplemented`); `magic.can_cast`
  View missing; `magic.level`/`magic.spells` need a `magicView` + promotion from
  `self.spells.*`; events `on spell_cast/hit/resist` absent.
- **self:** `position.plane` not impl; `equipped` is a List, not per-slot
  (`.weapon`/`.shield` need an equipment-by-slot decoder); `prayers`/`spells`
  fully implemented under `self.*` → promote to `prayer.*`/`magic.*`.
