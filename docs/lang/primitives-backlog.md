# Primitives backlog (from AutoRune + IdleRSC script surveys, 2026-05-28)

This is the buy/build list distilled from surveying two existing
RSC botting frameworks — AutoRune (Alex's earlier work) and
IdleRSC (which hosts three back-compat APIs: sbot, apos, idlescript).
The DSL's job is to make these patterns natural without re-creating
every Java method 1:1; many "verbs" in the old frameworks collapse
into convenience wrappers over a small number of well-shaped
primitives.

The list is the surviving design intent if context is lost. **Tier
ordering reflects what the bulk of real scripts actually call** —
not theoretical completeness. Skill-specific surfaces (prayer,
magic, fatigue tricks) defer until a routine needs them.

## Design principles surfacing from the survey

1. **`use(item, target)` is the keystone.** Most "skill" verbs are
   wrappers — cooking is `use(raw, fire)`, smelting is `use(ore,
   furnace)`, lockpicking is `use(key, door)`, fletching is
   `use(log, knife)`. We've wired the boundary + inventory-on-
   inventory variants; scenery / ground-item / npc / player still
   needed.

2. **Convenience verbs live in routine-land, not the DSL.** A
   `cook(item)` helper is a routine snippet, not a builtin. The
   DSL stays small; the standard library of idioms grows.

3. **Anti-ban is not our concern, but naturalistic variance is.**
   The user owns the world (no detection risk). However hosts
   should *feel* different from each other and from session to
   session. The runtime silently injects per-reverie variance:
   "give me a banker" returns a *random* banker each time, "wait
   2 seconds" carries small jitter, `walk_to` picks slightly
   varied corner-paths. These are RUNTIME concerns invisible to
   the host script — set per reverie (persona slot) and
   deterministic per seed. Hosts don't write `random_jitter()`;
   it just happens. See [jitter design](#jitter-design) below.

4. **State observed via accessors, state changed via verbs.** Old
   frameworks mix the two (`GetX()`, `setBankItem()`). We keep
   the bright line: `self.position.x` is a read; `walk_to(...)`
   is a write. Bang variants (`!`) carry the abort-on-failure
   semantics, which the old frameworks lacked entirely.

5. **Polymorphic targets, not v1-style typed verbs.** sbot has
   `AttackNPC` / `AttackPlayer` / `MagicNPC`; we have `attack(target)`
   that dispatches on view type. Same for `use`. Routines stay
   readable; the runtime resolves the opcode.

## Tier 1 — Blocking for broad script support

The single most common code paths in real bots hit these. We can't
ship a meaningful library of example routines without them.

| Surface | Status | Wire / notes |
|---|---|---|
| `use(item, scenery)` | partially built | opcode 115, `UseItemOnScenery` exists; need DSL dispatch + `*sceneryView` |
| `use(item, ground_item)` | TODO | opcode 53; `*groundItemView` already exists, just wire the dispatch arm |
| `use(item, npc)` / `use(item, player)` | TODO | opcodes need investigation; thieving, trade-prep, item-give patterns |
| `interact_at(x, y, command?)` | TODO | far-range click-on-object — cook on the fire 1+ tiles away. Opcode 136 (OBJECT_COMMAND) / 79 (OBJECT_COMMAND2) |
| `distance_to(target)` | TODO | Chebyshev distance accessor. Cheap utility, both surveys flag. |
| `in_region(x1, y1, x2, y2)` | TODO | "am I inside this rectangle?" — AutoRune's `GoToIfCoordsIn` |
| `world.last_server_message.contains("locked")` | TODO | substring match on the recent-events ring; routines branch on prose without re-implementing string scan |
| `world.dialog.options` | TODO | list of strings; `answer(N)` is blind index today — surface the option text so routines pick by content via `find_option(text)` |
| `npc.is_in_combat` / `npc.is_talking` | TODO | live state, not just type/level. Combat-loop reactor needs these to detect "target busy with another player". |
| `host.idle_ticks` | TODO | AutoRune's `%IdleC` — ticks since last meaningful action. Anti-stuck logic + retry windowing. |

## Tier 2 — High value, secondary blockers

| Surface | Status | Wire / notes |
|---|---|---|
| `walk_path([(x,y), ...])` | TODO | pre-planned multi-corner walk; bypasses the 96×96 pathfinder grid for long routes (Lumbridge → Ardougne) |
| `world.npcs.by_type(id).random()` | TODO | random NPC selection — but see jitter section: this might be the *default* runtime behavior |
| `repeat_until(predicate, timeout=Ns)` | TODO | retry block with timeout; banker-busy "Please wait" loops |
| `world.ground_items.by_id(id, radius=N)` | TODO | nearest ground item of type, with optional radius filter |
| `dialog.is_open` / `dialog.reset()` / `wait_for_dialog(timeout)` | TODO | explicit quest-menu state, vs implicit answer() timing |
| `is_reachable(x, y)` | TODO | sbot's `CanReach` — "is there a path?" before walking |
| `event.item_gained(id, count)` | TODO | fires when inventory grows; replaces poll-loops in fishing/woodcutting routines |
| `last_attacked_npc` / `last_attacked_player` accessor | TODO | sbot's `SetVarLastPlayerID` — track the last opponent for flee/duel logic |

## Tier 3 — Defer until first routine needs them

| Surface | Why later |
|---|---|
| Prayer API (enable/disable/check) | No routine uses prayer yet; ~58 accessors in our spec table marked NotYetImplemented |
| Spell API (cast on self/npc/item/ground) | Magic skill is Phase 5+ territory |
| Fatigue trick / explicit fatigue state | Mining/woodcutting at high level; we're not there |
| Shop API (buy/sell/query) | Not blocking any current scenario |
| World hop | RSC has one world; n/a until swarm phase |
| Paint / overlay | delos territory, not host territory |
| Server-side `info` / `position` / health commands | Already covered by `command(...)` |

## Jitter design

Goal: hosts that don't write any "be human-like" code still
*behave* differently from each other and across sessions, in
ways routine authors don't need to think about.

Knobs the runtime silently varies (per reverie, deterministic
per seed):

- **NPC / scenery selection.** When a routine asks for "a banker"
  via `world.npcs.by_type(banker_id)`, runtime returns a slightly
  shuffled list — most often the nearest, occasionally the second-
  nearest. Reverie A always picks the closest; reverie B prefers
  the leftmost; reverie C wanders to a far one once in a while.
- **Wait jitter.** `wait 2` resolves to ~1.8-2.3s with bounds
  shaped by persona (twitchy = tight jitter, distractible = wide).
- **Walk corner-paths.** When pathfind returns multiple equal-cost
  routes, runtime picks one weighted by persona (some hosts cut
  corners, some hug walls).
- **Action ordering when serialized.** If a routine queues `eat`
  then `attack`, runtime can introduce a brief settling pause
  scaled by persona.

What the runtime does NOT silently change (because correctness
matters): target identity when the routine explicitly named one,
inventory slot resolution, error codes, return values.

Reverie config (stored alongside the host's persona slot, future
Phase 4 work):

```yaml
reverie: maeve
seed: 0x4D414556  # MAEV in ASCII; deterministic
jitter:
  wait_ms_pct: 0.15        # wait N → uniform N±15%
  npc_pick_top_n: 3        # randomize among 3 nearest matches
  scenery_pick_top_n: 2
  path_corner_skew: 0.2    # bias toward straight-line vs hugging
  action_pause_ms_max: 80  # settling pause after action chain
```

Hosts feel like individuals not because their scripts say so but
because the runtime executes the same script with different
seeds. This is also why the SAME routine running on bernard vs
delores can produce visibly different behavior without either
script being aware.

**Why this matters for the survey backlog:** several requested
surfaces (`random_npc`, `move_to_randomly`, `wait_with_variance`)
are *already* handled by this design — they don't need to be
explicit DSL verbs. The DSL stays small; the runtime carries the
naturalism.

## Out of scope (deliberate)

- **Real anti-cheat / anti-ban.** We own the world. Detection
  isn't a risk. The variance work is for *believability* — making
  500 hosts feel like a town, not 500 copies of one bot.
- **Mouse / pixel-level emulation.** We're at the packet layer.
  No mouse-movement randomization, no UI-click coordinates.
- **Per-tick rate limiting beyond cap budget.** The cap budget
  is correctness-coupled (runaway loops). Naturalism is a
  separate layer.
