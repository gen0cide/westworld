# Current state (read this first on context refresh)

Last refreshed: 2026-05-28, after DSL steps 1–5 (interpreter shipped, runs against mocked builtins).

This doc captures where the bot actually is so a fresh-context Claude
can pick up productively without re-deriving everything from the
codebase + chat. It's deliberately frank about what's verified live
vs. built-but-untested vs. designed-but-not-built — Phase 2 is far
from done.

## Two-line summary

The bot (host = single OpenRSC connection + world-state mirror +
event bus) can navigate the world like a real client, kill mobs in a
loop, eat food, examine its surroundings, exchange PMs, and open
doors. The custom `.routine` DSL is *started* — lexer + AST + parser
skeleton compile and pass tests, but statement parsing, expression
parsing, validation, and interpretation are all still ahead.

## What's actually verified live against OpenRSC

These have been run end-to-end and confirmed by server logs:

- **Login / logout cycle** for accounts `alex` (admin) and `delores`
  (bot, combat lvl 34, atk/str/def all 30).
- **Walking via pathfinder**: multi-corner walk packets. Verified
  with a 25+ tile walk around Lumbridge castle walls. Direct port
  of `mudclient.walkToArea`'s wire format.
- **Open boundary (door)**: pathfinder routes adjacent to the door,
  sends InteractWithBoundary opcode 14, server confirms `onOpBound`
  trigger.
- **Goblin combat loop**: AttackNpc → server walks delores to mob →
  combat resolves → `onKillNpc` server log → ground item drops
  (bones) → PickUpItem with `onTakeObj` confirmation → ItemCommand
  (Bury) with `Bones.onOpInv: Bury` confirmation. Ran 3 kill cycles
  in a single 5-min session.
- **Server-side player follow** (opcode 165): delores trailing
  alex around Lumbridge.
- **PM exchange** with `alex`: AddFriend → PrivateMessage → alex
  echoes back the exact text. PM smart08_16 length prefix bug
  found and fixed during testing.
- **Inbound chat decode**: opcode 234 update-players, opcode 79 NPC
  coords, opcode 131 server messages, opcode 120 PM with body
  decompression.

## Built but NOT verified live

These compile + have unit tests but have never been run
end-to-end in-game. Each one is a tripwire — assume bugs until proven
otherwise.

- **TalkToNpc on a static NPC** (Hans / shopkeeper / banker). One
  attempt walked partway; never reached the NPC because RSC NPCs
  wander.
- **ChooseDialogOption** (opcode 116). Outbound packet built; never
  fired against a real `Default.onTalkNpc` dialog tree.
- **AttackPlayer** (opcode 171). Never tested.
- **Trade handshake** end-to-end. Init/Accept/Decline/Offer/Confirm
  primitives exist; the full two-stage confirm dance is untested.
- **Bank deposit / withdraw / close**. Primitives exist; bot has
  never visited a banker NPC.
- **DropItem** (opcode 246). Never tested in-game.
- **AutoEat** watcher. Code runs in background but delores hasn't
  taken enough damage from goblins to trigger an eat.
- **Death / respawn**. Combat loop has a death watcher but delores
  hasn't died.

## Not yet built

- **Duels** (PvP gentleman's duels) — different opcode/UI than
  trades.
- **All 18 RSC skills**: cooking, woodcutting, fletching, mining,
  smithing, fishing, firemaking, crafting, magic, prayer, ranged,
  herblaw, agility, thieving. Each has its own trigger opcode and
  state machine. Nothing wired.
- **Combat style toggle** (3 melee styles).
- **Magic spells** (cast-on-target opcodes exist in the enum but
  outbound action not built).
- **Dynamic boundary updates** — cut webs / opened doors changing
  state at runtime aren't reflected in the pathfinder grid. The
  grid is static-only; the inbound packets that signal boundary
  state change aren't decoded.
- **Inventory amount decoder bug**: `decodeInventory` reads
  `uint32` amount for every slot regardless of whether the item is
  stackable. Non-stackable slots get garbage amounts (you'll see
  `bronze Axe x10879108` in inventory listings). Needs to consult
  `facts.ItemDef.IsStackable`.

## The DSL — exact status

Per `docs/dsl.md`, the routine runtime has 9 implementation steps.
Through 2026-05-28 we've shipped steps 1–5.

| Step | Status |
|---|---|
| 1. Lexer + parser skeleton | ✓ |
| 2. Real statement parser (if/while/for/return/abort/wait/require/assign) | ✓ |
| 3. Expression parser with precedence climbing | ✓ |
| 4. Static validator | ✓ |
| 5. AST interpreter (locals, control flow, procs, member/index, builtins) | ✓ |
| 6. Action channel + Host bridge | pending |
| 7. Resource caps (op budget, wall clock, recursion, memory) | pending |
| 8. Event handler dispatch + two-tier scope | pending |
| 9. Conformance suite + delos observability hooks | pending |

Step 5 ships a working interpreter that runs `.routine` files end
to end against mock builtins + entity protocols. Step 6 is the
bridge into the real Host so a routine can actually drive the bot.

The DSL is the unblock for everything else. Without it, every "test
mining" / "test trading" requires Go code + recompile. With it,
you write `mine_iron.routine` and `cradle -routine ...`. Resist the
temptation to wire skill primitives before DSL is usable.

Conventions established for the DSL:

- File extension is **`.routine`**, not `.ws` (chose with alex; `.ws`
  was historic WordStar / WSF Windows Script File).
- Packages live under `dsl/`: `dsl/token/`, `dsl/lex/`, `dsl/ast/`,
  `dsl/parser/`, `dsl/validator/`, `dsl/interp/`.
- Hand-written recursive-descent — no parser generator, no regex
  in the lexer.
- F-string lexer alternates between literal-fragment mode and
  placeholder mode using two flags (`inFString`, `inPlaceholder`).
- AST nodes implement `Node` (has `Pos()` + private marker) plus
  `Stmt`/`Expr` sub-interfaces for type-safety in the parser.
- Control flow in the interpreter is propagated via panic with
  sentinel types (`returnSignal`, `breakSignal`, etc.) recovered
  at the top of each block — keeps every recursive call site
  clean of plumbing.
- Truthiness, equality, and numeric promotion follow Python-style
  rules per dsl.md.
- Reserved names `self` / `world` / `inventory` / `combat` are
  bound from `Interpreter.Reserved` at routine startup; entities
  expose attribute access via the `Getter` interface and indexing
  via `Indexer`.

## Repo / package layout

```
westworld/
├── action/        — outbound packet helpers (one file per concern)
├── cmd/cradle/    — single-bot CLI driver
├── docs/          — design docs (architecture, brain, dsl, etc.)
├── dsl/
│   ├── token/     — token kinds + Position + Token
│   ├── lex/       — lexer (state-machine, hand-written)
│   ├── ast/       — AST node types (Node/Stmt/Expr interfaces)
│   ├── parser/    — recursive-descent parser (currently: stub bodies)
│   ├── validator/ — (not yet)
│   └── interp/    — (not yet)
├── event/         — typed event types + event.Bus pub/sub
├── facts/         — static OpenRSC defs + locs (loaded once per process)
├── pathfind/      — BFS, grid, sector loader, multi-corner walk encoding
├── proto/v235/    — wire format: framing, opcodes, ISAAC, RSC compression
├── runtime/       — Host (per-bot stateful object); concerns split across
│                    follow.go, combat.go, items.go, boundary.go, social.go,
│                    bank.go, trade.go, combat_loop.go, auto_eat.go,
│                    examine.go, pathing.go, host.go
├── session/       — TCP/ISAAC session wrapper
└── world/         — per-host world-state mirror (Self, Inventory, Npcs,
                     Players, GroundItems)
```

## Important runtime concepts

### What I called "reactor patterns"

I used "reactor" as a sloppy name for **long-running goroutines that
subscribe to the event bus and drive Host actions in response**.
There are two of them today:

- `runtime.CombatLoop` (in `combat_loop.go`) — kill → wait for drop
  signal → loot → bury, looping until kill cap reached or ctx cancels
- `runtime.AutoEat` (in `auto_eat.go`) — background HP-threshold
  watcher; eats food when HP fraction drops below the configured
  threshold

The term is mine, not from the design docs. In the DSL these would
both be routines with their own `on` handlers. Once the DSL is live,
these go away — they're temporary Go-coded versions of what should
be `.routine` files.

### "Host" = one bot

`runtime.Host` is the per-bot object. Owns: TCP session, world
mirror, event bus, facts pointer (shared across hosts in a swarm),
pathfind landscape archive (also shared), logger.

### Walk packet quirks

`mudclient.walkToArea` sends `firstStep` + signed-byte deltas of
direction-change corners. `Path.addStep` server-side interpolates
between corners — only the *final* tile of each addStep gets a
`player-blocking` check. Sending tile-by-tile waypoints (one per
inter-corner tile) makes every intermediate get the player-blocking
check, which truncates the walk if any other player is on a tile.
Always send corner-compressed.

### Server-side action handler pitfalls observed

- `GroundItemTake` calls `player.resetPath()` if the item isn't in
  view — sending walk+take in one burst nukes the walk. Two-phase
  required: walk first, poll until within view radius, then take.
- `GameObjectWallAction` (boundaries) only sets a WalkToAction, never
  initiates walking. The bot must send a Walk packet to the boundary's
  tile before the interact packet.
- NPC pathing handlers `setFollowing` + per-tick `walkToEntity` only
  emit 1-step paths that the WalkingQueue silently drops if the next
  tile isn't adjacent — so any NPC further than 1 tile needs an
  explicit precursor walk.
- `Path.MAXIMUM_SIZE = 50` total tiles. Multi-corner walks longer
  than this get silently truncated.

## Security note

The OpenRSC test password leaked twice into committed files
(once as a default flag value in `cmd/cradle/main.go`, once into
an earlier version of this doc). Both leaks were scrubbed via
`git filter-repo --replace-text` + `--replace-message` and
force-pushed to origin.

The bot reads its password from the `WESTWORLD_PASSWORD` env var
at runtime. **Never** embed the literal value anywhere — code,
docs, commit messages, log lines, none of it. When discussing
the incident, refer to it as "the OpenRSC test password" and
nothing more.

Rotate the OpenRSC-side password if the value has any reuse
risk elsewhere — even after force-push, GitHub caches the old
commit blobs for some time.

Before committing: `git diff --cached | grep -i <known-secret>`.

## Open design questions for next session

From `docs/dsl.md` "Open language-design questions":

- Range syntax `..` vs `..=` — currently `..` per dsl.md.
- F-string syntax `f"hi {name}"` vs shell-style — Python f-strings
  chosen.
- Truthiness rules — Python-style "many falsey" chosen, codify in
  interpreter step.
- Equality — value equality only chosen.

Nothing blocks step 2 (statement parser) on these.

## Quick commands for next session

```bash
# Build everything
cd ~/Code/westworld && go build ./...

# Run all tests
cd ~/Code/westworld && go test ./...

# Run the bot with a single action
export WESTWORLD_PASSWORD=...
cd ~/Code/westworld && /tmp/cradle -username delores -look-around -dwell 5s

# Server logs (server-side trigger plugins log to here)
tail -f /Users/flint/Code/openrsc/server/logs/rsc_preservation_1.log

# Latest pcap per player (for wire-level debugging)
ls -lt "/Users/flint/Code/openrsc/server/logs/pcaps/RSC Preservation/<name>/2026-05/" | head

# Decode a pcap
gunzip /tmp/x.pcap.gz && tcpdump -r /tmp/x.pcap -A -x -nn | head
```

## How to pick up cleanly

1. `git pull` on `~/Code/westworld`.
2. `go test ./...` to confirm everything still passes (dsl/ should
   show four green packages: lex, parser, validator, interp).
3. The interpreter at `dsl/interp/` runs `.routine` files end-to-end
   but only against **mocked** builtins/entities. Step 6 wires it to
   the real Host so a routine can actually drive an OpenRSC bot.
4. Step 6 plan: create `runtime/dsl_bridge.go` exposing
   `Host.RegisterDSLBuiltins(it)` that adds Callables for every
   action in `dsl.md` "Built-in actions" + primitive `wait`. Also
   need `runtime.SelfView` / `runtime.WorldView` / `runtime.InventoryView`
   types implementing `interp.Getter` so the routine can read
   `self.hp`, `inventory.free`, `world.npcs`, etc.
5. After step 6 is up, add `cradle -routine path.routine` to
   `cmd/cradle/main.go` so we can run real `.routine` files at the
   command line.
6. Resist the temptation to wire skill primitives before steps 6+8
   are done. The DSL exists so skill testing becomes "write a
   routine," not "write Go + recompile."

**Reminder about secrets:** the OpenRSC test password leaked twice
already (once as a default flag, once in this very state.md). Read
the value from `WESTWORLD_PASSWORD` env var only; do NOT embed it
in any file, doc, or commit message — not even when documenting
incidents. Grep the staged diff for the literal before committing.
