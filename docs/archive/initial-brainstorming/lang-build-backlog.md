> **ARCHIVED (initial brainstorming), 2026-06-10.** The surface this backlog tracked shipped —
> see `docs/lang/api.md`, `docs/lang/actions.md`, `docs/state.md`. Open crumbs were harvested
> into `docs/TODO.md` (examine verb = DSL-21; idle_ticks, GroundItem.quantity, research tails
> = MP-10). **The #121 content table below (bucket=21 not 52, def-655 = Gnome course, pottery
> wheel, ...) is the only copy anywhere — preserved verbatim, cited from docs/lessons-learned/.**
> Kept verbatim below for the record.

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
- **RESOLVED** Boundary direction-byte encoding — it's a **4-value enum 0–3**
  (0=N-edge of (x,y), 1=E-edge, 2=`\` diagonal, 3=`/` diagonal), distinct from
  scenery's 8-way compass. `(x,y,dir)` is the exact server lookup key
  (`getWallObjectWithDir` demands an exact dir match → null + suspicious-flag
  otherwise). `.at(x,y,dir)` must echo the same dir carried on opcode 91 /
  `BoundaryLocs.json`. Canonical address: wall between rows y-1/y → `(x,y,0)`;
  between cols x-1/x → `(x,y,1)` (pick the higher-x/higher-y tile). See
  protocol.md §1 boundary dir-byte table.
- **RESOLVED** Opcode 50 disambiguation — **not** shared in v235. Opcode 50 =
  `CAST_ON_NPC`, opcode **135** = `NPC_USE_ITEM` — two distinct opcodes, both
  length 4 (NOT length-disambiguated). The open item was a cradle bug, not a
  wire ambiguity: `outUseItemOnNpc` must be **135** (`magic.cast(spell,npc)`
  correctly uses 50). With the wrong literal, `use(item, npc)` silently fires a
  bogus `CAST_ON_NPC` with `spellID = inventory slot`. See protocol.md §3/§5.
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
  *Data-source RESOLVED:* read off opcode **206** (`event.PrayersActive`) — one
  byte per prayer slot (length = slot count = 14), slot `i` active iff byte==1.
- **BUILD** `magic` spellbook: `known`/`has_runes_for` (rename from `self.spells.*`).
  *Data-source RESOLVED — no packet exists; build entirely from data the cradle
  already has.* `magic.known` = `facts.Spells` filtered by `ReqLevel <= self
  magic level` (stat index 6). `has_runes_for(spell)` = inventory count ≥ each
  required rune count, **plus** an equipped elemental staff satisfies that
  element (fire rune 31→items 197/615/682, water 32→102/616/683, air
  33→101/617/684, earth 34→103/618/685). `can_cast` is the **level-only** static
  gate per api.md (rune/gear check stays in `has_runes_for`).
- **BUILD** `bank.is_open` + `bank.slots` (NotYetImplemented). *Unblocks bank-balance assertions.*
  *Data-source RESOLVED:* `bank.slots`/`used`/`max_size` ← opcode **42**
  (`event.BankOpened`, `storedSize`/`maxBankSize` clamped to 255) mutated by
  opcode **249** (`event.BankSlotUpdate`); `bank.is_open` is a **state machine**
  (true on 42, false on 203) — no in-packet boolean. **Caveat:** bank/inventory
  amounts use the generator's *smart* `writeUnsignedShortInt` (cradle decodes
  this) while the OpenRSC bundled GUI client reads `get32()` (fixed 4-byte) —
  they disagree for amounts ≤ 32767; verify against the live server.
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
- **BUILD (NEW gap) fatigue→sleep** — `self.fatigue` is exposed (read) but there
  is **no way to act on it**: `self.is_sleeping` is a stub
  (`accessors.go:109`) and there is no verb to use the sleeping bag / submit a
  sleep-word / wait-for-wake. Every legacy skilling loop (IdleRSC
  `getShouldSleep()`/`sleepHandler()`, AutoRune `GetSomeSleep`) gates on
  fatigue and then sleeps. Wire pieces already exist inbound: opcode 117
  (`InSleepScreen` — sleep-word captcha appeared), 194 (`InSleepwordIncorrect`),
  84 (`InStopSleep` — woke up). **Needs Alex's input on two design questions:**
  (a) the **sleep-word/captcha** — submitting the bag triggers a fatigue-captcha
  the bot must answer (OCR/brain or a server that disables it?); (b) is sleep a
  **body** faculty (a `sleep`/use-bag verb + wake detection) or a **mind-layer**
  policy (api.md §11 puts budgets in the mind layer, and sleep sits ambiguously
  between perception and policy)?
- **RESOLVED (CONTENT, not engine)** `use(key, boundary)` — rides the **same**
  opcode 161 (`USE_WITH_BOUNDARY`) as knife-on-web; the cradle already
  dispatches `*boundaryView` targets to `UseItemOnBoundary`. The server
  distinguishes by item+door, not a separate opcode. Author the scenario as
  `use(key, world.boundaries.at(x,y,dir))` targeting a real keyed door. Read
  result by **case-insensitive substring-match on `locked`** (and `unlock` for
  success) in `world.messages` — OpenRSC emits *"The door is locked"* /
  *"this door is locked"* (case varies), never the literal *"it's locked"*;
  correct key → *"you unlock the door"* → *"you go through the door"*.
- **RESOLVED (CONTENT, not engine)** firemaking — NOT use-item-on-item (91) and
  NOT use-on-scenery (115); lighting logs is **use-tinderbox-on-ground-item
  (opcode 53)**, which the cradle already speaks (`UseItemOnGroundItem`). The
  RSC mechanic is: **drop the logs first**, then `use(tinderbox, ground-item)`.
  Using the tinderbox on logs in the inventory returns *"I think you should put
  the logs down before you light them!"* (never lights). Use **plain Logs
  (item 14)** on the default OpenRSC config (`CUSTOM_FIREMAKING=false`): only
  id 14 lights via the authentic path; oak/willow (632–636) return *"Nothing
  interesting happens"* despite having FiremakingDef entries. The fix is routine
  sequencing — `drop(logs)` → locate the new `world.ground_items` entry →
  `use(tinderbox, that)` — not an engine change.
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
**#121 content table — VERIFIED against OpenRSC def data + server source:**

| Item / scenario | Correction |
|---|---|
| **bucket** | **= item 21** (empty wooden Bucket). NOT 52/53 — **52 = Silverlight, 53 = Broken shield**. Filled variants: Milk 22, Water 50. |
| **gold-ore** | **= item 152** (smeltable gold ore; smelts to gold bar 172). Confirmed three ways. Iron ore is the adjacent 151 — don't confuse. |
| **def-655** | **= LOG_GNOME_COURSE** ("log", cmd1 "balance on") — the **Gnome Agility Course** log balance, **NOT Al-Kharid**. |
| **pottery wheel** | **= scenery def-179** (POTTERY_WHEEL, cmd1 "WalkTo"). Live placements: **(227,524)/(228,524)** (Barbarian Village), **(22,573)** (Rimmington). (Pottery oven = def-178; spinning wheel = def-121.) Interaction is use-clay-on-wheel. |
| **Druidic-Ritual gate** | Have the test host (must be SuperMod) run `::completeallquests` in setup — it sets every quest stage to -1, which satisfies Herblaw's `getQuestStage(DRUIDIC_RITUAL) != -1` gate. Confirmed. |
| **"protect from melee" prayer** | **Does NOT exist in RSC.** The highest combat-protect prayer is **"Protect from missiles"** (slot 13, req lvl 40, ranged-only). Rename `combat-prayer-protect-from-melee` → `combat-prayer-protect-from-missiles` (and `recall_query` → "protect from missiles"); the scenario body **already activates the real "Protect from missiles"** prayer, so no logic change. |
| **iron-scimitar** | **IS a smith-menu item: item 83**, smithable at level 20 from **2 iron bars** (the "Scimitar (2 bars)" option). No repoint needed. **But the setup has a separate bug:** `smithing-iron-scimitar` uses `item 171 2` (171 = **steel** bar) — must be `item 170 2` (170 = **iron** bar). With steel bars the menu would smith a steel scimitar (needs lvl 35 > the scenario's setstat 30 → fails) and `inventory.find(83)` never resolves. Hammer (item 168) is correct. |

- **SKIP** jewellery-teleport (ring of dueling / amulet of glory — RS2 items, absent in RSC; already replaced with enchant).
- **NOTE** the bot-agent "CONTENT FIX: `cast_on_npc(rat,0)`" is **wrong** — the name-string form is verified live; `combat-magic-attack-on-rat` is flaky (cast timing), not mis-signatured.
- **NOTE** stale comment in `facts/prayers.go:6` claiming "OpenRSC adds Protect from Melee at 14" — NOT present in this OpenRSC build (Prayers.java ends at slot 13); correct or drop it.

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
