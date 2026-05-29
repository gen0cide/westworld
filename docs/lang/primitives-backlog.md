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

## Tier 1 — Blocking for broad script support ✓ (mostly shipped)

The single most common code paths in real bots hit these.

| Surface | Status | Wire / notes |
|---|---|---|
| `use(item, scenery)` | ✓ #75 | opcode 115, polymorphic dispatch lives in `runtime/dsl_actions.go` |
| `use(item, ground_item)` | ✓ #76 | opcode 53 |
| `use(item, npc)` / `use(item, player)` | ✓ #77 | NPC_USE_ITEM + PLAYER_USE_ITEM opcodes |
| `interact_at(x, y, command?)` | ✓ #78 | opcode 136 (OBJECT_COMMAND) / 79 (OBJECT_COMMAND2) |
| `distance_to(target)` | ✓ #79 | Chebyshev distance accessor |
| `in_region(x1, y1, x2, y2)` | ✓ #79 | rectangle containment |
| `world.last_*.contains(text)` | ✓ #80 | substring match on the recent-events ring |
| `world.dialog.options` | ✓ #81 | option text list; `find_option(text)` for content-based selection |
| `npc.is_in_combat` / `npc.is_talking` | ✓ #82 (decoder sweep) | live state available on entity view |
| `host.idle_ticks` | TODO | AutoRune's `%IdleC` — ticks since last meaningful action. No routine has needed it yet. |

## Tier 2 — High value, secondary blockers ✓ (mostly shipped)

| Surface | Status | Wire / notes |
|---|---|---|
| `walk_path([(x,y), ...])` | ✓ #83 | pre-planned multi-corner walk |
| `world.npcs.by_type(id).random()` | ✓ #89 | typed NPC selection + jitter-friendly random |
| `repeat { … } until <cond> timeout <secs>` | ✓ #85/#96 | shipped (note: timeout is a bare number of seconds, not `Ns`) |
| `world.ground_items.by_id(id, radius=N)` | ✓ #88 | nearest-by-type with optional radius |
| `dialog.is_open` / `dialog.reset()` / `wait_for_dialog(timeout)` | ✓ #86 | explicit quest-menu state |
| `is_reachable(x, y)` | ✓ #84 | pre-flight pathfinder check |
| `event.item_gained(id, count)` | ✓ #87 | inventory growth event; replaces poll loops |
| `last_attacked_npc` / `last_attacked_player` accessor | ✓ #90 | tracked through combat resolution |

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
