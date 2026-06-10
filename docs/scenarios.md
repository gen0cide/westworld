# The Scenario System

> **STATUS: BUILT** (verified 2026-06-10 against branch `tidy/structure-and-docs`,
> HEAD `0bfa818`). The `.routine` corpus (`examples/scenarios/`, **263 files** —
> the hand-maintained source of truth), the **code-free manifest**
> (`cmd/scenariogen/scenarios.yaml`, 263 rows referencing the files), the
> **validator** (`cmd/scenariogen`), and four runner scripts all exist and work.
> Two sibling catalogs (`scenarios_proposed.yaml`, `scenarios_bots_proposed.yaml`)
> are **proposals not yet merged** ([TODO.md](TODO.md) S-2), still in the old
> embedded-code schema (§3).
>
> **Source-of-truth model (refactored 2026-05-31 — read §2.1):** the `.routine`
> files are authoritative; `scenarios.yaml` only *references* them (no embedded
> code). `go run ./cmd/scenariogen` validates manifest ⇄ corpus consistency both
> ways; `-reindex` re-derives the manifest. The old embedded-code catalog is
> preserved at `scenarios.yaml.bak`.
>
> Where a doc string in the codebase disagrees with the code, this doc states
> what the **code actually does** and flags the drift. Open scenario work lives
> in [TODO.md](TODO.md) (S-1…S-5); this doc carries no backlog.

Scenarios are the project's primary **gap-finder and live-test harness** for the
[routine DSL](lang/README.md) and the runtime engine. This doc is the
authoritative guide: what a scenario is, the manifest schema, the lifecycle
(idea → author → index → run → triage), single- vs multi-host execution, and
a step-by-step recipe for adding and running one.

Related docs: [lang/writing-routines.md](lang/writing-routines.md) (how to write
the DSL inside a scenario body + the engine gaps you'll hit),
[lang-scenario-candidates.md](archive/initial-brainstorming/lang-scenario-candidates.md) (corpus-mined
candidate ideas), [lang/api.md](lang/api.md) (the full DSL builtin/accessor
surface), [lang/syntax.md](lang/syntax.md) (grammar).

---

## 1. What a scenario *is*

A **scenario** is a scripted, usually short (seconds to a few minutes) live-test
that drives one (or two) bot hosts against the **live OpenRSC server** and
checks an assertable outcome. Each scenario is one hand-authored, deterministic
`.routine` file — the same DSL the production hosts run — indexed by a metadata
row in the `scenarios.yaml` manifest (§2.1).

Scenarios serve three overlapping purposes, in priority order:

1. **Gap-finders (primary).** A scenario is deliberately written to exercise a
   behaviour end-to-end against the real engine. When it fails, it surfaces a
   concrete gap: a DSL verb isn't wired, an `on <event>` never fires, a
   `world.*` accessor returns the wrong value, the state mirror lags a server
   push, or an admin command behaves differently than assumed. The failure
   message (plus the captured server messages) tells you *which layer* broke.
2. **Regression tests.** Once a gap is closed, the scenario flips to PASS and
   guards against re-introduction. Re-running the full sweep before/after a
   change is the cheapest way to spot a regression.
3. **A behaviour corpus.** The 263 `.routine` files double as worked
   examples of "how do I make a bot do X" for every skill, common quest starts,
   combat, trading, and movement.

The assertable outcome is a **DSL boolean expression** (the trailing
PASS/FAIL block in the file; `pass:` in the legacy schema, §3) — typically a
delta on XP, an inventory count, a skill level, position, trade/duel state, or a
substring of `world.last_server_message`. If the expression is true the routine
`return`s `"PASS: <id>"`; otherwise it `abort`s `"FAIL: <id> (predicate was
false)"`. Scenarios may also `abort` earlier inside the body with a specific
diagnostic ("lobster never appeared in inventory").

### Where scenarios sit in the stack

```
 .routine corpus  ◄──validates──►  cmd/scenariogen  ──►  scenarios.yaml
 examples/scenarios/<cat>/*.routine  (Go validator)      (code-free manifest)
 [SOURCE OF TRUTH, hand-maintained]   (-reindex re-derives the manifest)
                                                          │
              run_*.sh  ──►  legacy-cradle -routine ──► dsl/interp (interpreter)
                                                          │              │
                                                    runtime.Host ── live OpenRSC server
                                                          │
                                  "routine ended" log  ◄──┘  (kind + value/err)
                                                          │
                                  run_*.sh parses ──► PASS/FAIL/ABORT/ERROR/SKIP
```

The scenario system therefore *integrates* every other westworld layer: the DSL
parser/interpreter (`dsl/`), the runtime host + state mirror (`runtime/`), the
wire protocol (`proto/`), and the static facts (OpenRSC defs). A scenario
failing is a real signal somewhere in that chain.

---

## 2. The three catalogs

All three live in `cmd/scenariogen/`. **`scenarios.yaml` is now a code-free
manifest** (§2.1) referencing the `.routine` files. The two proposal files still
use the **old embedded-code schema** (a `scenarios:` list with `body`/`setup`/`pass`
fields) and must be converted to `.routine` files + manifest rows when merged.

| File | Entries | Status | Notes |
|---|---|---|---|
| `scenarios.yaml` | **263** | **MANIFEST** (§2.1) — code-free index referencing the 263 `.routine` files; validated by `go run ./cmd/scenariogen`, re-derivable via `-reindex`. | The reference index; the `.routine` files are the source of truth. |
| `scenarios_proposed.yaml` | 47 | PROPOSED — herblaw / skillups / combat / quest scenarios verified against OpenRSC defs but **not merged** ([TODO.md](TODO.md) S-2). | Convert to `.routine` files + manifest rows after review (§11). Has a trailing comment block listing scenarios blocked on a not-yet-existing verb + members/ambiguity flags. |
| `scenarios_bots_proposed.yaml` | 60 | PROPOSED — bot-archetype loops (power-gather, production, combat, thieving, banking/travel) modeled on classic RSC scripts (SBot/APOS/AutoRune/idlescript). Every id/coord/recipe/xp verified against OpenRSC defs (2026-05-29). **Not merged** (S-2). | Every id carries a `bot-` prefix to avoid collisions. Has a "NEEDS VERB" backlog comment block at the end of the file. |

> The "verified 2026-05-29" notes in the proposal files are **manual**
> annotations against `ItemDefs.json` / `SpellDef.xml` / `GameObjectDef.xml` /
> `SceneryLocs.json`. There is no automated check that those ids stay valid if
> the OpenRSC defs drift — re-verify by hand when merging.

### 2.1 The `.routine` files are the source of truth; the catalog is a code-free manifest

**Refactored 2026-05-31.** Originally `scenarios.yaml` *embedded the routine code*
and `scenariogen` emitted the `.routine` files from it. That drifted badly: the
"run-to-ground" campaign edited the generated files **directly** and never touched
the catalog —

| commit | `scenarios.yaml` | `.routine` files |
|---|---|---|
| `18ac18b` (r1+r2 merge) | 0 | 48 |
| `fd0731c` (r3-combat) | 0 | 10 |
| `40bea3a` / `272ca58` / `5117845` (r3 ranged/magic/gather) | 0 | 3 / 3 / 2 |

— so the catalog regenerated stale *pre-r3* versions (regenerating
`bot-autofighter-ranged-shortbow` produced the old "goblin @ (206,497)" scenario,
not the committed "JailRanger @ (285,659)" one — an 82-line divergence), and the
`bot_autofighter_ranged_jail_safespot.routine` file `40bea3a` added had no catalog
entry at all (the 196-files-vs-195-entries gap).

**The fix — the current model:**

- The **`.routine` files under `examples/scenarios/` are the source of truth** —
  authored and edited by hand; the runner scripts execute them directly
  (`for f in examples/scenarios/*/*.routine`).
- **`scenarios.yaml` is now a code-free manifest** — one row per scenario (`id`,
  `category`, `file`, `hosts`, `admin`, `timeout`) that *references* the file. **No
  routine bodies live in YAML**, so the embedding-drift class is eliminated.
- **`cmd/scenariogen` is now a validator**, not a body generator:
  - `go run ./cmd/scenariogen` — validates the manifest ⇄ corpus **both ways**:
    every entry points at a real file whose name matches and that parses, **and**
    every `.routine` file has a manifest row (so a hand-added file with no row — the
    old orphan — now fails the check). Exits non-zero on any problem; CI-able.
  - `go run ./cmd/scenariogen -reindex` — re-derives the manifest from the corpus
    headers (the bootstrap / resync tool).
- The old embedded-code catalog is preserved at `cmd/scenariogen/scenarios.yaml.bak`.

**Workflow:**
- **Edit** a scenario → edit its `.routine` file. (No regeneration; nothing to clobber.)
- **Add** a scenario → create the `.routine` file *and* add a manifest row (or run
  `-reindex`), then `go run ./cmd/scenariogen` to verify.
- Full DSL parse errors are also surfaced by `go run ./cmd/parsecheck`.

`scenarios.yaml` breaks down by `category` (= corpus subdirectory) as:

| Category | Count | What it covers |
|---|---|---|
| `skills` | 128 | One+ minimal exercise per RSC skill (woodcut, mining, cooking, magic, prayer, herblaw, …). |
| `combat` | 36 | Attacking NPCs, eating under fire, retreat, ranged/magic, multi-kill, two-drone PvP. |
| `edges` | 36 | Edge cases / fragile paths (empty-bucket fill, off-by-one dialog, mirror gaps). |
| `quests` | 25 | Quest *starts* (talk-to-NPC + first dialog branch); full walkthroughs are follow-up work. |
| `movement` | 20 | Walking, pathing, teleport, region-load races. |
| `social` | 18 | Chat, friends, PMs, trade, duel, follow, summon. Most multi-host scenarios live here (14 of 15). |

(There is no formal enum of categories in code — `category` is a free-form
string; the validator only enforces that it equals the corpus subdirectory
(`examples/scenarios/<category>/<id>.routine`, `cmd/scenariogen/main.go`
`validate()`). The six above are the ones in use, listed in the manifest's
header comment.)

---

## 3. The legacy embedded-code schema (field by field)

> **LEGACY (read §2.1 first).** This `body`/`setup`/`pass` embedded-code schema is
> **no longer how `scenarios.yaml` works** — that file is now a code-free manifest
> and `cmd/scenariogen` is a validator. This schema survives only in the two
> **proposal files** and in `scenarios.yaml.bak`, and documents how the retired
> generator built `.routine` files (useful when converting a proposal: render it
> mentally, hand-write the `.routine`, add a manifest row). The `Scenario` struct
> below describes the **old** generator, preserved in `scenarios.yaml.bak`'s shape.

Each list item under `scenarios:` (in a *proposal* file or the `.bak`) is one
legacy `Scenario`. The canonical worked example:

```yaml
- id: heal-via-food
  category: skills
  hosts: any-drone
  admin: true
  recall_query: lobster food eating heal hp
  precondition:
    - self.hp > 0
  setup:
    - item 373 1
  body:
    - 'command(f"damage {self.name} 9")'
    - wait 1
    - lobster = inventory.find(373)
    - 'if lobster == null { abort "lobster never appeared in inventory" }'
    - eat(lobster)
    - ok = wait_until(_ => self.hp > 1, 5)
    - 'if not ok { abort f"eat had no effect after 5s — still at hp={self.hp}" }'
  pass: self.hp > 5
  timeout: 15
```

| Field | Type | Required | Meaning & generator behaviour |
|---|---|---|---|
| `id` | string | yes | kebab-case, unique. Becomes the **filename** (`<id>.routine`) and, kebab→snake, the **routine name** (`heal_via_food`). A loader validator enforces *filename-stem == routine-name*, and DSL identifiers can't contain hyphens — that's why the conversion exists. |
| `category` | string | yes | Output subdirectory under `examples/scenarios/`. Use one of the six in §2. |
| `hosts` | string | yes | Free-form orchestration hint. `any-drone` = single host, drone-agnostic. `any-drone-and-bernard` / `Drone1-and-Drone2` = multi-host pair (see §6). The runner scripts grep this header to decide whether to skip a scenario or pair two cradles. |
| `admin` | bool | no | `true` if the routine uses admin `::commands` (`item`, `damage`, `setstat`, `teleport`, `summon`, `wipeinv`, …). Requires an admin account (Drone1-20 + Bernard + Delores + Arnold are admin in the OpenRSC DB clone). Emits a `# Admin: required` header line; the generator does **not** otherwise gate on it — a non-admin host just gets "You are not a moderator" at runtime. |
| `recall_query` | string | no | A search query run against the rsc.wiki corpus **at generation time** (only if `-corpus` is set, which it is by default — see §4). The top 2 hits are stamped into the file header as `# Wiki grounding:` citations. Pure documentation; has zero runtime effect. Empty disables. |
| `precondition` | list of expr | no | Boolean DSL **expressions** (not statements). Emitted into a `require { … }` block at the top of the routine. If any is false at entry, the routine aborts before the body runs. Use for "must be alive", "must have inventory space", etc. |
| `setup` | list of string | no | Admin command strings **without** the `::` prefix (e.g. `item 373 1`, `teleport 135 663`, `summon bernard`). The generator wraps each as `command("<setup>")` followed by `wait 1.5`. **Special case:** a bare `wipeinv` is rewritten as `command(f"wipeinv {self.name}")` + `wait_until(_ => inventory.used == 0, 5)` — see §5. |
| `body` | list of string | no | Raw DSL lines, emitted verbatim at 4-space indent. A single list item may be a multi-line YAML block scalar (`\|-`) holding a `select{}` / `for{}` / `repeat…until` / `if{}` construct; the generator re-indents each line to body level. This is where the actual test logic lives. |
| `pass` | string (expr) | no | A boolean DSL expression. If non-empty, the generator appends `if <pass> { return "PASS: <id>" }` then `abort "FAIL: <id> (predicate was false)"`. **Leave empty** when the body already returns/aborts explicitly (e.g. trivial-PASS social scenarios verified via the partner's log). |
| `timeout` | int (seconds) | no | Human-facing header annotation **and** the runner derives `cradle -dwell` from it. Defaults to 30s in the runners if absent. Budget = setup waits + action waits + RNG/server variance + margin. |
| `notes` | string | no | Free-form annotation baked into the header as `# Notes:`. Convention: cite ground-truth sources (`ItemDefs.json` ids, `SpellDef.xml` names, coordinates) and explain *why* the scenario is interesting or fragile. |

### What the generator emitted

For the example above, the retired generator wrote
`examples/scenarios/skills/heal_via_food.routine` (the file is committed and
still looks exactly like this — it is the shape to hand-write when converting):

```
# Scenario: heal-via-food
# Category: skills
# Hosts: any-drone
# Admin: required
# Timeout: 15s
#
# Wiki grounding (query: lobster food eating heal hp):
#   - Manual:Cooking (2003) § Cooking Items — https://classic.runescape.wiki/...
#   - Manual:Cooking & Fishing (2002) § Cooking and fishing guide — https://...

runtime "1.0"
routine heal_via_food() {
    require {
        self.hp > 0
    }

    command("item 373 1")
    wait 1.5

    command(f"damage {self.name} 9")
    wait 1
    lobster = inventory.find(373)
    if lobster == null { abort "lobster never appeared in inventory" }
    eat(lobster)
    ok = wait_until(_ => self.hp > 1, 5)
    if not ok { abort f"eat had no effect after 5s — still at hp={self.hp}" }

    if self.hp > 5 {
        return "PASS: heal-via-food"
    }
    abort "FAIL: heal-via-food (predicate was false)"
}
```

> **The `runtime "1.0"` directive is mandatory** for disk-loaded routines —
> `runtime.ParseRoutineFile` rejects a file without it (`dsl_bridge.go`; see
> [lang/versioning.md](lang/versioning.md)), so the validator fails too. The
> retired generator predates it (the line was added corpus-wide afterwards);
> always include it when hand-converting a proposal.

> **Doc-drift note.** The legacy schema comment block (now at the top of
> `scenarios.yaml.bak`) says setup commands get a `wait 0.5` between them. The
> retired generator actually emitted **`wait 1.5`** (admin commands take effect
> via a server round-trip + push packet, and 0.5s let `find()`/`search()` race a
> not-yet-synced mirror), and that is what the committed `.routine` files
> contain. Trust the **1.5s** when converting a proposal.

---

## 4. The validator: `cmd/scenariogen`

`scenarios.yaml` is now a **code-free manifest** (§2.1) and `cmd/scenariogen` is a
**validator**, not a body generator — it writes no scenario bodies and recompiles
nothing.

```bash
go run ./cmd/scenariogen            # VALIDATE manifest ⇄ corpus (CI-able; non-zero on any problem)
go run ./cmd/scenariogen -reindex   # RE-DERIVE the manifest from the corpus headers
```

Flags (defaults shown):

| Flag | Default | Purpose |
|---|---|---|
| `-yaml` | `cmd/scenariogen/scenarios.yaml` | The manifest to validate / (with `-reindex`) write. |
| `-root` | `examples/scenarios` | Root of the `.routine` corpus. |
| `-reindex` | `false` | Re-derive the manifest from the corpus `# header` lines and write it to `-yaml` (bootstrap / resync). |

**Validation** checks both directions and exits non-zero on any failure: every
manifest entry points at a real `.routine` whose filename matches its `id` and that
parses (`runtime.ParseRoutineFile`), **and** every `.routine` in the corpus has a
manifest row — catching a hand-added file with no row (the orphan that motivated
this refactor). Full DSL parse errors across the corpus are also surfaced by
`go run ./cmd/parsecheck`.

> **Removed in the 2026-05-31 refactor:** the old `-out` / `-id` / `-dry` flags and
> the `-corpus` rsc.wiki grounding — the tool no longer writes bodies, so there is
> nothing to ground at "generation time." Wiki-grounding citations already stamped
> into some `.routine` headers remain as static comments. `apply_fixes.py` (a legacy
> batch find/replace tool for the old embedded-code `scenarios.yaml`) is likewise
> obsolete under the manifest model.

---

## 5. Self-cleaning setup and the `wipeinv` special case

> The self-cleaning **pattern** below is real and lives directly in the `.routine`
> files now. The "the generator renders `setup:` as …" mechanics are **legacy**
> (the generator no longer renders bodies, §4) — but the emitted shape is exactly
> what the committed `.routine` files contain, so this still explains what you'll
> see and should write.

Sweeps run many scenarios back-to-back, often on the **same drone**, sometimes
without a clean logout between them. So scenarios are designed to be
**self-cleaning**: their setup restores known state before the body runs, rather
than trusting the previous scenario to have tidied up.

The `wipeinv` special case is the linchpin. The OpenRSC admin command
`::wipeinv` **requires an explicit player name** — a bare `::wipeinv` is a silent
no-op that does *not* empty the bag. Because the generator renders setup lines as
plain `command("<literal>")` (no f-string interpolation, so a bare `wipeinv`
could never carry the name), it special-cases it:

```
# YAML:  setup: [ wipeinv ]
# emits:
    command(f"wipeinv {self.name}")
    wait_until(_ => inventory.used == 0, 5)
```

The name-bearing f-string actually clears the bag, and the `wait_until` blocks
until the mirror confirms an empty inventory — so subsequent item grants land
into a clean bag regardless of residual inventory from a prior scenario. **Always
put `wipeinv` first in `setup:`**, before any `item …` grants.

> This is why authors must never write a bare `command("wipeinv")` in a `body:`
> line — it would be a no-op. In `setup:` use the bare `wipeinv` token (the
> generator wraps it); in `body:` write `command(f"wipeinv {self.name}")`
> yourself. See [lang/writing-routines.md §4](lang/writing-routines.md).

---

## 6. Execution models: single-host vs multi-host

### How one scenario runs

A scenario is run by `cradle -routine <file>` (`cmd/legacy-cradle` — the
single-connection CLI the runner scripts build to `/tmp/cradle`; not
`cmd/cradle-server`, the fleet daemon). Cradle logs the host in, loads the DSL
interpreter + state mirror + facts, runs the routine, and logs:

```
routine ended  kind=returned  value="PASS: heal-via-food"  err=""
```

`kind` is the `dsl/interp` `ResultKind`: `completed` (fell off the end),
`returned` (`return value`), `aborted` (`abort reason`), `errored` (runtime/type
error), or `canceled` (context timeout). The runner scripts map these to
outcomes (§7).

Relevant cradle flags for scenarios (defaults shown):

| Flag | Default | Purpose |
|---|---|---|
| `-routine <path>` | — | Parse + run this `.routine` against the live host. |
| `-username <name>` | `alex` | RSC account. Admin scenarios need an admin account. |
| `-server <host:port>` | `localhost:43596` | OpenRSC server. **The runner scripts override this to `localhost:43594`.** |
| `-password` | — | Or set `WESTWORLD_PASSWORD` env (see security note below). |
| `-dwell <dur>` | `5s` | **Doubles as the routine's timeout** — `RunRoutine` runs under `context.WithTimeout(ctx, dwell)`. The runners set `dwell = timeout + headroom`. (The flag help says "no -dwell needed", which is true for the *idle*-after-routine sense; the runners still pass it because it bounds the run.) |
| `-facts <openrsc-root>` | `/Users/flint/Code/openrsc` | Static OpenRSC defs that resolve `type_id → NPC name` and `def_id → scenery name`. **On by default.** With `-facts ""`, every `n.name == "…"` and `world.locs.search("tree")` returns nothing and name-based scenarios abort as "not in view / nearby". |
| `-reset-on-exit` | `false` | ADMIN/TEST ONLY. Before logout, runs `wipeinv <name>` / `heal` / `recharge` / `teleport 120 649` so the next scenario on this drone starts clean. The modern parallel runners pass this; never pass it for production hosts. |

### Single-host (`any-drone`)

One cradle process. The scenario is pure DSL against one live connection. Setup
commands run in-process (`wait 1.5` between each). Cleanup is via the
self-cleaning `setup:` blocks (§5) and/or `-reset-on-exit`. **The old approach**
— a separate cradle invocation running `scenario_reset.routine` between every
scenario — is **deprecated**: it caused a login storm (one login per scenario)
that broke 54+ scenarios in an early sweep.

> `examples/routines/scenario_reset.routine` still exists but is a **retired
> pattern**. Don't use it as a model — it calls a bare `command("wipeinv")`,
> which is the no-op described in §5. The in-process `-reset-on-exit` flag
> replaced it.

### Multi-host (`any-drone-and-bernard`, `Drone1-and-Drone2`)

Two cradle processes — a **primary** and a **partner** — for interactions where
one bot must wait for another's action: trade, duel, follow, summon, PvP. The
15 multi-host scenarios in `scenarios.yaml` (14 `any-drone-and-bernard`, all in
`social`; 1 `Drone1-and-Drone2` in `combat`) are skipped by the single-host
runners. `run_multihost.sh` currently hard-codes **7** of them as `run_pair`
lines; the other 8 (the duel-rules / trade-screen assertion scenarios) have no
pairing line yet and so never run in a sweep — add a `run_pair` line to cover
one.

`run_multihost.sh` orchestrates the pairing:

1. Launch the **partner** in the background. If it has a *responder routine*
   (e.g. `bernard_respond_trade.routine`), `-dwell 75s`; if it just idles online
   as a summon/duel target, `-dwell 45s`. Partner runs with `-reset-on-exit`.
2. `sleep 5` — partner login + (responder) entry into its listen loop.
3. Launch the **primary** with the scenario, `-dwell 45s`.
4. Wait for the primary to finish, then **reap the partner gracefully** (it
   returns on its own and logs out cleanly — never killed — so its account is
   released for the next pairing without a "code 4 / already logged in" race).

**Responder pattern** (the partner side). The partner waits on game events with
`select { on <event> … timeout … }` and reacts. From
`examples/routines/bernard_respond_trade.routine` (Bernard is *not* admin, so he
can't teleport — the primary summons him):

```
routine bernard_respond_trade() {
    wait 8                                  # let the summon land
    select {
        on trade_request(from) { res = trade.request(from); … }   # accept by re-requesting
        timeout 60s { return "no_request" }
    }
    select { on trade_opened(other) { … } timeout 10s { return "no_open" } }
    trade.offer([])                          # Bernard offers nothing
    trade.accept()                           # first accept (offer screen)
    select { on trade_other_accepted() { … } timeout 15s { return "no_first_accept" } }
    trade.confirm()                          # second accept (confirm screen)
    select { on trade_closed(completed) { … } timeout 15s { return "no_close" } }
}
```

> Trade API recap (a historical confusion): `trade.request(x)` initiates *or*
> accepts a request; `trade.accept()` is the **first** accept (offer screen);
> `trade.confirm()` is the **second** accept (confirm screen); `trade.offer([])`
> sets your side; events `trade_opened` / `trade_other_offer` /
> `trade_other_accepted` / `trade_closed` drive the reactor.

**Primary pattern** (the scenario side), from
`examples/scenarios/social/social_trade_initiate_bernard.routine`:

```
routine social_trade_initiate_bernard() {
    command("teleport 122 658")
    wait 1.5
    command("summon bernard")
    wait 1.5
    command("item 18 5")
    wait 1.5

    wait_until(_ => world.players.find(p => p.name.lower == "bernard") != null, 8)
    b = world.players.find(p => p.name.lower == "bernard")
    if b == null { abort "bernard not summoned" }
    trade.request(b)
    ok = wait_until(_ => world.trade.is_active, 10)
    if not ok { abort "trade window never opened" }

    if world.trade.is_active {
        return "PASS: social-trade-initiate-bernard"
    }
    abort "FAIL: social-trade-initiate-bernard (predicate was false)"
}
```

The current multi-host pairings are hard-coded in `run_multihost.sh` (drone3 +
bernard for trade/duel/follow/PM/summon; drone1 + drone2 for mutual attack).
Adding a new multi-host scenario means adding a `run_pair …` line there.

---

## 7. The runners (four of them) and the result format

All four live in `cmd/scenariogen/`, source `WESTWORLD_PASSWORD` from
`.local.env`, build cradle (`go build -o /tmp/cradle ./cmd/legacy-cradle`), and
write tab-separated outcomes to `/tmp/scenario_results.tsv`. They differ in
concurrency model and facts default.

| Script | Model | Facts | `-reset-on-exit` | Use when |
|---|---|---|---|---|
| `run_one_per_drone.sh` | **1:1 fan-out** — scenario *i* → drone *(i mod 220)+1*, all concurrent, each drone runs exactly one test then logs out. | **ON** (default) | yes | Fastest full sweep; wall-clock ≈ slowest single test. Also stress-tests the MySQL concurrent-login path. **Preferred for a clean full sweep.** |
| `run_parallel.sh` | **20 workers** — drones 1-20, each runs its slice serially; ~15-20× faster than sequential and avoids back-to-back relogin races. | **ON** (default) | yes | Full sweep when you don't have 220 accounts free / want bounded load. |
| `run_all.sh` | **Sequential** — every single-host scenario on one drone (default drone3), one at a time. | **OFF** (`-facts ""`) | no | Legacy / simplest. **Warning:** passes `-facts ""`, which silently breaks every name-based `find()`/`search()`. Prefer the parallel runners; if you must use this, it's only sound for scenarios that don't resolve entities by name. |
| `run_multihost.sh` | **Paired** — the 7 hard-coded multi-host pairings (§6; 8 more multi-host files have no pairing line yet). | ON (default) | yes (partner **and** primary) | Trade / duel / follow / summon / PvP. Skipped by the single-host runners. |

> **Facts inconsistency (real, in the code).** `run_all.sh` passes `-facts ""`
> while `run_parallel.sh` / `run_one_per_drone.sh` / `run_multihost.sh` all leave
> facts enabled and carry an explicit comment that facts *must* stay on. A
> name-resolving scenario can therefore PASS under the parallel runners and
> ABORT under `run_all.sh`. This is a known footgun, not intended behaviour —
> use a parallel runner for anything that looks up NPCs/scenery by name.

**Multi-host detection.** The single-host runners skip a file if its
`# Hosts:` header matches `bernard|Drone[12]` and is not exactly
`# Hosts: any-drone` — so a malformed/missing header could mis-classify a file.

**Result format.** The runners grep the cradle log for the first `routine ended`
or `run failed` line and map it:

| Outcome | Source | Meaning |
|---|---|---|
| `PASS` | `kind=returned … "PASS:"` | `pass:` expression was true (or body returned a PASS string). |
| `FAIL: <value>` | `kind=returned … "FAIL:"` | `pass:` expression was false. |
| `ABORT: <value>` | `kind=aborted` | Body called `abort <reason>` (precondition or in-body diagnostic). |
| `ERROR: <err>` | `kind=errored` | Runtime/type error: unknown builtin, type mismatch, nil deref in DSL. |
| `LOGIN/SETUP FAIL` | `run failed` | Couldn't connect/login or setup failed before the routine ran. The parallel runners retry up to 3× on a transient login "code 4". |
| `SKIP` | (single-host runner) | The file is multi-host. |

On any non-PASS, the runners dump the **last ~20 server messages** the bot
received (the cradle logs each player-facing message as a `server msg` line in
`runtime/host.go`) — these are the single most useful triage signal ("You can't
light a fire here", "You need a Druidic Ritual", "You can't retreat for another
3 rounds").

> The result parsing depends on the exact strings `routine ended` /
> `run failed` and the `kind=…` / `value="…"` / `err="…"` shape. If cradle's
> logging format changes, the sweeps silently mis-classify — keep that contract
> stable.

---

## 8. Lifecycle: idea → author → index → run → triage

```
 ① idea          ② author .routine   ③ index+validate   ④ run            ⑤ triage
 corpus mining   copy the closest    add a manifest      run_one_per_     PASS  → done / regression guard
 TODO.md gaps    file, edit header   row (or -reindex);  drone.sh         FAIL  → predicate false → engine or content?
 a live bug      + body by hand      go run              (or _parallel/   ABORT → in-body diagnostic → read the reason
                                     ./cmd/scenariogen    _multihost)     ERROR → DSL runtime bug → fix body/builtin
                                                                          server msgs explain WHY
```

### ① Idea / corpus

Scenario ideas come from three places:

- **Gaps.** Read [TODO.md](TODO.md) — the open-work SSOT — for open DSL/engine
  gaps ([tasks.md](tasks.md) is the *closed* historical ledger; don't mine it
  for open work). Write a scenario that exercises the gap concretely; it should
  FAIL today and PASS once the gap is built.
- **Corpus mining.** Analyse classic RSC bot scripts for canonical loops not yet
  covered. [lang-scenario-candidates.md](archive/initial-brainstorming/lang-scenario-candidates.md) is a
  worked list of ~22 such candidates grouped by theme (fatigue/sleep,
  bot-to-bot trade/duel, combat/health, shop, loot, inventory, banking, prayer,
  menu fragility) with a gap → scenario cross-reference table.
- **Regression.** A bug report becomes a scenario that reproduces it.

### ② Author the `.routine`

Copy the closest existing file under `examples/scenarios/<category>/` and adapt
it by hand — header comment block, `require`, setup `command(...)` lines, body,
PASS/FAIL tail (§9 is the full recipe; §3 shows the canonical shape).

### ③ Index + validate

Add a manifest row to `cmd/scenariogen/scenarios.yaml` (or run
`go run ./cmd/scenariogen -reindex` to derive it from the file's header), then
`go run ./cmd/scenariogen` — it must print
`OK: N manifest entries ⇄ N corpus files`. A missing row, a mismatched
filename, or a parse error all fail the check.

### ④ Run

Test the one file in isolation first, then sweep (§9 steps 3-4).

### ⑤ Triage PASS/FAIL into engine vs content gaps

This is the payoff. For each non-PASS, decide which layer is responsible:

- **Engine gap** (fix in Go): the failure is a missing/wrong verb, an event that
  never fired, a `world.*` accessor returning stale/null, or a mirror that lags
  a server push. The server message often confirms the action *was* valid but
  the DSL couldn't observe/express it. Route to [TODO.md](TODO.md) (the open-work
  SSOT — DSL gaps live in its DSL-\* section, e.g. DSL-19 event-vocabulary).
- **Content gap** (fix in the `.routine`): the scenario assumed a wrong item id,
  wrong coordinate, a members-only recipe on F2P, an unmet quest gate, or too
  short a timeout. The server message says "you can't do that here / you need
  X". Fix the scenario file.

The live-test campaign tracks this triage at scale (last recorded ~57/88 PASS on
an early sweep; the corpus has since grown to 263 — the open-ended
run-to-ground campaign is [TODO.md](TODO.md) S-1). Scenarios are gap-finders, so
a not-yet-100% pass rate is expected and informative — each FAIL is a TODO item,
not necessarily a broken test.

---

## 9. How a new agent adds + runs a scenario (step by step)

**Prereqs:** repo root is `/Users/flint/Code/westworld`; an admin account
(drone1-20 / bernard / delores / arnold); the OpenRSC server reachable (the
runners default to `localhost:43594`); `WESTWORLD_PASSWORD` available — it's read
from `.local.env` automatically (refer to it only as the `WESTWORLD_PASSWORD`
env var; never print it).

**Step 1 — Author the `.routine` file.** Copy the closest existing file under
`examples/scenarios/<category>/` to
`examples/scenarios/<category>/<your_scenario_id>.routine` (snake_case; the
routine name **must** equal the filename stem — `runtime.ParseRoutineFile`
enforces it) and edit it by hand:

- **Header comment block** (parsed by `-reindex` and grepped by the runners):
  `# Scenario: <kebab-case-id>`, `# Category: <one of the six>`,
  `# Hosts: any-drone` (exact string, or the multi-host forms — the runners
  pattern-match this line, §7), `# Admin: required` (only if you use
  `::commands`), `# Timeout: <N>s` (the runners derive `-dwell` = N + headroom
  from this line; default 30 if absent).
- **`runtime "1.0"`** directive before the `routine` line — mandatory for
  disk-loaded files; the validator rejects a file without it (§3).
- **`require { … }`** block of boolean preconditions (aborts before the body).
- **Setup**: `command("<admin cmd>")` + `wait 1.5` pairs; put
  `command(f"wipeinv {self.name}")` + `wait_until(_ => inventory.used == 0, 5)`
  first (§5).
- **Body**: the test logic. **Poll after teleport/summon** with `wait_until`
  before binding entities — the mirror loads asynchronously.
- **PASS/FAIL tail**: `if <predicate> { return "PASS: <id>" }` then
  `abort "FAIL: <id> (predicate was false)"` — or return/abort explicitly in
  the body.
- Cite ground truth (`ItemDefs.json` ids, `SpellDef.xml` names, coords) in a
  `# Notes:` header line; say why the scenario is interesting or fragile.

**Step 2 — Index + validate:**

```bash
# add a row to cmd/scenariogen/scenarios.yaml by hand (id/category/file/hosts/admin/timeout)
# …or derive it from your file's header block:
go run ./cmd/scenariogen -reindex
go run ./cmd/scenariogen            # must print "OK: N manifest entries ⇄ N corpus files"
```

The validator fails on a missing row (orphan file), a row whose `file` doesn't
match `examples/scenarios/<category>/<snake(id)>.routine`, a duplicate id, or a
file that doesn't parse.

**Step 3 — Run it in isolation** (against your test server, with facts ON):

```bash
go build -o /tmp/cradle ./cmd/legacy-cradle
set -a; source .local.env; set +a          # loads WESTWORLD_PASSWORD
/tmp/cradle -username drone3 -server localhost:43594 \
  -routine examples/scenarios/<cat>/your_scenario_id.routine \
  -dwell 60s -reset-on-exit -v 2>&1 | tail -60
```

Look for the `routine ended` line. If `kind=aborted`/`errored`/`returned …
FAIL`, read the `value`/`err` and the `server msg` lines, then iterate on the
file directly — there is nothing to regenerate. (For a multi-host scenario,
launch the partner first per §6, or add a `run_pair` line and use
`run_multihost.sh`.)

**Step 4 — Run a sweep** to confirm you didn't break neighbours:

```bash
./cmd/scenariogen/run_one_per_drone.sh localhost:43594   # 1:1 fan-out, facts ON
# or, bounded load:
./cmd/scenariogen/run_parallel.sh localhost:43594 20
# results: /tmp/scenario_results.tsv
```

Snapshot a baseline to diff later:
`./cmd/scenariogen/run_parallel.sh 2>&1 | tee /tmp/sweep_$(date +%s).log`.

**Step 5 — Commit** (only when asked):

```bash
git add cmd/scenariogen/scenarios.yaml \
        examples/scenarios/<cat>/your_scenario_id.routine
git commit -m "Add scenario: your-scenario-id — <one-line gap/behaviour>"
```

(Always commit the `.routine` **and** its manifest row together so the corpus
and the index stay in sync — `go run ./cmd/scenariogen` is the pre-commit
check.)

---

## 10. Troubleshooting

| Symptom | Likely cause / fix |
|---|---|
| ABORT before the body (`require` failed) | A precondition is false (e.g. `self.hp > 0` but the bot is dead). Check account state on the server; add a setup command to establish it. |
| ABORT "X not in view / nearby" or a name lookup returns null | Facts are off — you ran `run_all.sh` (`-facts ""`) or passed `-facts ""`. Use a parallel runner or pass `-facts /Users/flint/Code/openrsc`. |
| ABORT right after teleport/summon | Post-teleport race: the entity mirror loads async. Wrap the bind in `wait_until(_ => world.npcs.find(...) != null, 8)` before using it. |
| FAIL (predicate false) with no obvious reason | The scenario ran but `pass:` was false. Re-check setup/body logic; read the dumped `server msg` lines. |
| ERROR (`kind=errored`) | A DSL runtime/type error: unknown builtin, type mismatch, hyphen in an identifier. Check body syntax against [lang/api.md](lang/api.md). |
| Times out (`kind=canceled`) | `timeout` too short or an infinite wait in the body. Increase `timeout` ~50% and audit `repeat…until`/`wait_until` bounds. |
| "You are not a moderator" in server msgs | An `admin: true` scenario ran on a non-admin account. Use drone1-20 / bernard / delores / arnold. |
| Item grant lands in a dirty bag | Put bare `wipeinv` first in `setup:` (§5) so the generated `wait_until(inventory.used == 0)` clears it before grants. |
| `LOGIN/SETUP FAIL … code 4` | Transient "already logged in" relogin race; the parallel runners auto-retry 3× with backoff. If persistent, the prior session hasn't released — wait or use a different drone. |
| Multi-host hangs | Partner didn't enter its listen loop before the primary acted. Increase the `sleep 5` / partner `-dwell` in `run_multihost.sh`. |
| Scenario unexpectedly SKIP'd | Its `# Hosts:` header matched the multi-host pattern. Set `hosts: any-drone` for single-host. |

---

## 11. Maintenance notes

- **Corpus ↔ manifest sync.** The 263 `.routine` files under
  `examples/scenarios/` are the **hand-maintained source of truth** (§2.1) —
  edit them directly. Keep `scenarios.yaml` in sync (adjust the row, or run
  `go run ./cmd/scenariogen -reindex`), verify with `go run ./cmd/scenariogen`,
  and commit both together. **Never** treat the files as generated artifacts —
  that was the pre-2026-05-31 model, and it's the drift class the refactor
  eliminated.
- **Def drift.** Scenarios pin item ids, spell names, object def ids, and
  coordinates against the OpenRSC defs. There is no automated re-verification; if
  the server's defs change, scenarios fail with "you can't do that" server
  messages. Re-grep the offending scenario's `# Notes:` and confirm the id
  against the live `ItemDefs.json` / `SpellDef.xml` / `GameObjectDef.xml` /
  `SceneryLocs.json`. The automated ids/coords CI check is
  [TODO.md](TODO.md) S-5.
- **Retiring scenarios.** When a gap is permanently closed and a scenario is pure
  redundancy, delete the `.routine` file **and** remove its manifest row — the
  validator fails on either half left behind (missing file / orphan). More
  often, keep it as a regression guard.
- **Merging proposals** ([TODO.md](TODO.md) S-2). To pull in
  `scenarios_proposed.yaml` / `scenarios_bots_proposed.yaml`, hand-convert each
  entry: render the legacy schema mentally (§3), write the `.routine` file, add
  a manifest row, validate. Re-verify ids against the live defs first; mind the
  members-only and needs-verb caveats in the proposal files' trailing comment
  blocks.
