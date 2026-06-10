# The pathing saga: one root cause wearing five costumes

*Walk-through-walls, mishandled doors, "boxed upstairs," and the fatigue deadlock —
a patch-and-revert cycle that consumed most of the project's life and collapsed, in one
ground-truth study, to a single subsystem that never read the data we already had.*

Read this before touching `pathfind/`, `runtime/traverse.go`, `runtime/pathing.go`,
`runtime/obstacles.go`, or `worldmap/`. Everything below is cited; re-verify the cites,
don't re-trust the prose.

---

## The problem as experienced

The timeline is short and dense. The repo was born 2026-05-27 (`8aa2a18`). The
client-side pathfinder landed the same day (`772434d`) — and from its very first
commit the collision grid treated openable doors as walkable (the comment shipped
with it: *"Openable doors (Unknown=1) are treated as walkable — the bot resolves
them"*, pre-fix `pathfind/grid.go:72-73`, see `git show b24012c~1:pathfind/grid.go`)
and hard-forced plane 0 (`if plane != 0 { ... plane = 0 }`, same file `:76-80`).
Door auto-opening was bolted onto `walk_to` the next day (`69b0ba0`, 2026-05-28).
The definitive fix landed 2026-06-09 (`b24012c` + `6a329fd`). Thirteen calendar
days — but in this repo's compressed cadence that was the project's entire middle
life, and movement was its haunted subsystem the whole way. The symptoms were varied
enough to be filed as separate bugs:

- Hosts **walked straight at a wall** instead of routing eight tiles around it, then
  stood there grinding against it (`plutonium/README.md:50`).
- Hosts **walked through walls** — visibly clipping geometry the server clearly
  modeled, especially upstairs and in dungeons (`plutonium/README.md:48`).
- **Doors were mishandled in every direction at once**: closed doors were planned
  *through* as if open; a door the host just opened still read as a wall; locked doors
  were re-planned through forever; and a scenery *gate* (the cabbage-field/Lumbridge
  farm kind, def 60) was invisible to the door machinery entirely — the live
  "penned at the cabbage field" bug (`plutonium/movement-design.md:381`).
- The flagship live failure: **Delores boxed in an upstairs room at 100% fatigue**,
  XP frozen, the stall machinery flailing — she could see the bed, could not path to
  it, and nothing in her ever *decided* to sleep anyway
  (`docs/_research/SESSION-STATE-2026-06-09-forage-and-metacognition.md` §4).
- Long soaks quietly **bled the fleet to zero**: a routine server idle-kick is a clean
  EOF, which stored no error, so the supervisor never restarted the host
  (`plutonium/README.md:64`, surprise #7).

Each symptom got its own diagnosis-of-the-day, its own patch, and frequently its own
revert. The repo accumulated scar tissue: a door-blocking change that was reverted
when it broke something else (`plutonium/movement-design.md:263` — "the REVERTED
door-BLOCKING was correct to revert ONLY because it blocked openable doors as hard
walls" with no open-flow to resolve them); a multi-tile scenery footprint that only
stamped one tile (root cause 1, `movement-design.md:5-9` — the old
`TestFindPathToHans` literally depended on planning through Lumbridge castle's
phantom door half, per the `b24012c` commit message); and a no-path fallback whose
comment — "the server paths us up to the door" — was simply false
(`plutonium/README.md:50`). Movement bugs became "mysterious live behavior": the
class of bug you stop filing because you no longer believe the fix will hold.

## The false leads

**False lead #1: "are we even speaking the right protocol?"** The wire was the first
suspect — exotic symptoms suggest corrupted decode. The comparative study
byte-verified the entire authentic-v235 stack: tail-byte framing matched
`RSCProtocolEncoderMain.java`, the ISAAC keystream was proven bit-identical to the
server's `ISAACCipher.java` by a differential test, the RSA modulus matched the
server pem bit-for-bit, the login layout matched `LoginPacketHandler.java`
(`plutonium/README.md:38`). Protocol was never close to the problem. Fear retired
(`README.md:58`, surprise #1).

**False lead #2: "our perception layer is really messed up."** Also no. Wall objects
are edge-keyed, scenery tile-keyed, mirroring the server's type-1/type-0 split;
decode layouts are correct and version-appropriate; `PlaneOf(Y) = Y/944` matches the
server's `Formulae.getHeight` (`plutonium/README.md:41`). The genuinely missing piece
in perception was narrow: opcode 211 (bulk region-clear) was undecoded — and on the
authentic dialect it is the **sole** eviction channel for far-away boundaries, so
stale doors and scenery accumulated forever once you walked away and came back. Real,
but secondary (`README.md:41,63`).

**False lead #3: patch the symptom where it shows.** The door-block revert is the
canonical case: closed doors were once made to block, something else broke (the
auto-open machinery couldn't then reach them), and the change was reverted — leaving
closed openable doors blocking in **no** layer, with traversal relying on
plan-through + a 5-second stall heuristic + a 3×3 static-facts scan to *infer* a door
might be in the way (`plutonium/movement-design.md:22-25,64`). Each patch at the
symptom layer relocated the bug instead of fixing it.

**False lead #4: fallbacks that turn failure into silence.** When BFS returned
`ErrNoPath`, `WalkToOpts` fired a **raw far-target server Walk** — and the server's
`Path.addStep` straight-lines and *stops at the first wall* (`plutonium/README.md:50`,
old `host.go:1188`). `walkAndAct` swallowed `ErrNoPath` and fired the action from
wherever the host stood ("fire-action-anyway turns hard failures into silent
no-ops", `movement-design.md:396`) — worse, an action packet with no successful walk
plants a **latent** server-side `WalkToAction` that can fire later when the host
wanders near the object, a phantom-action hazard (`movement-design.md:222`). Every
one of these converted a diagnosable planning failure into a silent no-op, and that
is the engine of the thrash: the grid's bugs never surfaced as grid bugs, they
surfaced as ghosts.

**False lead #5: treat the working reference bot as a byte-oracle.** Plutonium (the
mature reference bot at `/Users/flint/Code/plutonium`) handles doors and planes
correctly, so the temptation was to copy it literally. Two traps were caught live.
First, Plutonium's dialect carries a direction byte on scenery packets that **v235
does not** — `Payload235Generator.java:646-653` writes only `[short id, byte x,
byte y]` (`movement-design.md:14`); porting Plutonium's read would have injected a
desync. Second, and worse: a prior *adversarially verified* finding claimed our
5-bit new-entity coordinate offsets were a bug to be widened to 6 bits "to match
Plutonium." The server source proves the opposite — `GameStateUpdater.java:273`
writes `forAuthentic ? 5 : 6` with the literal comment *"only have 5 bits in the
rsc235 protocol"* (`plutonium/README.md:29`). The "fix" would have corrupted the
entire NPC/player stream. Even the study's own first framing ("different servers")
was wrong and had to be corrected in place: same OpenRSC codebase, dialect
negotiated per-connection from the client's declared version
(`plutonium/README.md:23`, correction block). **The reference bot is an oracle for
algorithms, never for bytes; the server source is the only wire authority.**

**False lead #6: teach the DSL surface to compensate.** Richer selectors
(`reachable=` filtering, path-cost "nearest") and manual prose about reachability
were considered before the grid was honest. The synthesis called this out directly:
the DSL's reachability/nearest/walk surface and the runtime's collision grid are
**the same bug seen from two altitudes**, and the dependency runs one way — grid
first, DSL second; fix the surface first and "you don't fix the bug — you relocate
it: the selector now confidently returns 'reachable' targets that aren't"
(`plutonium/dsl-synthesis.md:3,11,164`). The flagship "ours-better" rating on
auto-door `walk_to` was *aspirational* — architecturally right, sitting on a broken
foundation (`dsl-synthesis.md:42-43`).

## The determined fix

The turn came from a deliberate methodology change, ordered by Alex on 2026-06-09
after pausing the metacognition work (`SESSION-STATE-2026-06-09…md:8`): stop
patching, run a ground-truth comparative study — the Plutonium analysis, 44
read-only agents over four references (our tree, Plutonium, the OpenRSC server
source as the correctness oracle, and the authentic client lineage
Client_Base/RSCPlus), 17 dimensions, 78 adversarially verified findings
(`plutonium/README.md:3`) — write **one** design
(`plutonium/movement-design.md`), and implement it **once**, in dependency order.

The study's headline: every pathing/door/upstairs symptom collapses to **ONE root
cause** — `pathfind/grid.go` built a collision grid that was **static-only,
plane-0-locked, and dynamic-blind** (`plutonium/README.md:15,43`) — with three
coupled defects:

- **(A) Dynamic-blind.** `BuildGrid` imported only static `facts` and never consulted
  `world.DynamicBoundaries`/`DynamicScenery` — which we *already decoded and stored*
  from opcodes 91/48. A closed door you could see was invisible to BFS; a door you
  just opened still read as a wall (`README.md:46`). The expensive bug was a missing
  wire-up, not a missing capability (`README.md:61`, surprise #4).
- **(B) Openable doors counted as walkable.** The blocking predicate excluded all
  openable doors (`Unknown==1`), but the server blocks closed `DoorType==1` doors
  regardless — **440 of 967** boundary instances. BFS under-blocked them 100%, so the
  door-open branch never fired: BFS always "found" a path straight through
  (`README.md:47`).
- **(C) Plane-0 lock.** `BuildGrid` hard-forced plane 0 and all four callers passed
  literal 0. For an upstairs/dungeon Y the sector math returned nil → an empty,
  all-walkable wall grid → "walks through walls upstairs" / "boxed upstairs"
  (`README.md:48`). The worldmap oracle was plane-0-only too
  (`movement-design.md:99`), so upstairs hosts also lost all reachability filtering.

The raw-Walk fallback (false lead #4) was the amplifier that made all three look
like ghosts.

The design then nailed the wire facts the fix had to honor — facts that live nowhere
else and that any future regression hunt needs:

- **A door "opening" is an opcode-91 id-OVERWRITE, never a removal**, on this wire:
  the authentic server never sends in-range boundary removals; open = ADD of the
  open def (doorframe id 11) at the same `(x,y,dir)`, re-close 3000ms later = another
  ADD of the closed id (`GameStateUpdater.java:1190-1232`, via
  `movement-design.md:24,217`). You confirm a door opened by watching the boundary's
  ID swap (or your own position advancing through it — mandatory for
  `doDoor(replaceID=-1)`, which streams *nothing*), never by timing.
- **v235 scenery packets carry no direction byte** (the Plutonium trap above);
  direction must be re-derived from the static loc at the tile
  (`movement-design.md:14`).
- **A door is a BOUNDARY — an edge between two tiles — not an object standing at a
  tile.** Alex's abstraction correction, recorded verbatim:
  *"`interact_at` acts on the OBJECT at the tile (e.g. the bed), NOT a door. A door
  is a boundary (edge between tiles), opened via `open_boundary(view)` or by `go_to`
  pathing through. Do NOT make interact_at 'open doors' — wrong abstraction."*
  (`SESSION-STATE-2026-06-09…md:53`). Doors and scenery gates are two distinct
  server dispatch channels (opcode 14 OpBound vs 136 OpLoc) and must be classified,
  not guessed (`movement-design.md:247,256`).
- **Approach rects are server-exact**: `Mob.atObject` accepts only specific stand
  rects per boundary direction, with **no** adjacency slack — an out-of-rect open
  click is silently dropped (`plutonium/findings.md:9-15`). Pathing must route into
  the rect the server will accept, not "near the door." Implemented as
  `approachRect`/`boundaryApproachRect` in `runtime/boundary.go:13-36`, mirroring
  `GameObject.getObjectBoundary`.
- **Confirm traversal by STATE, never by packet or timing**: wall-door = boundary id
  swap to a non-`DoorType==1` def; scenery gate = def swap (60→59); ladder = `Y/944`
  band change; walk = position delta (`movement-design.md:103` step 4; now the
  documented contract at `runtime/traverse.go:93-105`).

The implementation executed the design's steps 1–10 in dependency order, in two
commit waves on 2026-06-09:

1. **Protocol & store hygiene first** (`b24012c`): decode opcode 211
   (`InRemoveWorldEntity`, now `proto/v235/inbound.go:59,560`) so the live overlay
   can never be fed stale state; `RemoveRegion` on all three world stores; a
   wire-order application test.
2. **Static grid correctness** (`b24012c`): the type-2 scenery footprint loop —
   multi-tile gates had unmodeled FREE tiles; the per-direction edge mapping was
   verified correct against all three references, *only the loop was missing*
   (`movement-design.md:9`) — plus TileDef terrain blocking and the
   `stampSceneryFootprint` helper (`pathfind/grid.go:243-256`).
3. **Live overlay rewrite** (`b24012c`): closed `DoorType==1` doors block (the
   un-revert, done right this time because BFS can now route *to* them and the new
   flow opens them), scenery def-swap re-derivation via the static loc's direction,
   authentic swap semantics. The design's own deployment warning:
   *"DO NOT deploy step 3 to live without step 5 — closed doors blocking without the
   new open flow re-pens her"* (`movement-design.md:165`).
4. **Obstacle classifier** (`b24012c`): `findTraversableNear` + door/gate/ladder
   def tables — classify the obstacle at the exact blocked edge, don't infer it from
   stalls; on ambiguity try the wall channel then the object channel, exactly like
   Plutonium's `_process_doors` (`movement-design.md:103`).
5. **Traversal flow rewrite** (`b24012c`): plan → walk → classify → open via the
   obstacle's own channel → **confirm-by-state** → replan; **the raw-Walk fallback
   deleted outright** — "There is deliberately NO blind direct-walk fallback"
   (`runtime/traverse.go:102-105`); authentic walkAndAct (opcode-16 walks + action
   back-to-back).
6. **Refusal classifier + BlockedEdge ledger + typed errors** (`6a329fd`):
   locked/toll/quest/member server prose is classified
   (`runtime/obstacles.go:38-58`) and becomes a remembered blocked edge with a
   precondition and a 5-minute TTL (`obstacles.go:60-75`), consulted by every future
   plan — and `DoorLockedError` is unwrapped with `errors.As`, **not** a bare type
   assertion, which silently breaks the moment anyone wraps the error
   (`movement-design.md:34`; the fix carries its own warning comment in
   `runtime/dsl_actions.go`).
7. **Plane transitions** (`6a329fd`): ladder/stair macro + cross-plane `go_to` —
   planes stack in Y at 944-tile bands with no walkable connection between bands.
8. **Per-plane worldmap oracle** (`6a329fd`): the reachability oracle covers all
   four planes with gate-aware conditional edges (`worldmap/worldmap.go:12,321`,
   `worldmap/reachable.go:5-12`).
9. **Corridor `go_to`** (`6a329fd`): plan over the oracle, greedy stepping demoted
   to same-plane fallback.
10. **Test suite** (`b24012c`/`6a329fd`): `pathfind/overlay_test.go` (rewritten to
    authentic id-overwrite semantics — the old version pinned the revert),
    `pathfind/plane_test.go`, door-cycle and footprint tests, per the four-tier test
    plan in `movement-design.md:212`.

Two coupled liveness fixes shipped alongside (`0214a0b`, same day), because the study
showed the deadlock was a conjunction: a **proactive fatigue→sleep arbiter**
(`runtime/detour.go:273-279`) — the sleepword captcha was already auto-answered, but
nothing ever *decided* to sleep, and even if it had, the broken grid couldn't path to
the bed (`plutonium/README.md:65-67`) — and **clean-EOF supervisor restart**
(`session.ErrServerClosed` sentinel + pinned regression test
`session/conn_eof_test.go:39`, `TestServerEOFStoresErrServerClosed`), so an idle-kick
restarts the host instead of silently spinning no-op turns against a dead socket.

One correctness note the study preserved: our door *failure* handling was already
**better than the mature reference** — we capture the server's prose failure message
and return a typed error; Plutonium fails silently and its scripts hand-roll
reactions (`plutonium/README.md:62`, surprise #5). The design's explicit instruction:
preserve it, don't regress it. The prose itself is a trap: refusal strings vary
across the server's plugins ("The door is locked", quest variants, "You must pay a
toll of 10 gold coins to pass", "You need to talk to the border guard"), so
classification is case-insensitive substring families verified against the plugin
sources, not exact matches (`runtime/obstacles.go:9-58`,
`movement-design.md:105-107`).

**Coda — the same lesson at another layer, one day earlier.** On 2026-06-08, the #31
perception-reachability feature shipped (`a270236`): `scan_for` drops and
`describeArea` annotates entities in unreachable "negative space." It used **raw
collision components** to decide reachability — which is door-blind: a closed door
is a collision wall, so the room behind *any* openable door read as a different
component, and perception confidently marked door-connected rooms unreachable
("she can't come into the room with the open one",
`SESSION-STATE-2026-06-09…md:37-43`). A brand-new feature, built carefully, inherited
the grid's dishonesty the moment it consumed the grid. The door/gate-aware oracle
primitives (`worldmap/reachable.go`) are what perception must consult. This is false
lead #6 happening *in vivo*: any layer built on a lying truth source amplifies the
lie.

Outcome: the walk-through-walls, door, and boxed-upstairs symptom families closed
together — because they were always one bug.

## The durable rules

1. **The server source is the wire authority. A working peer bot is an oracle for
   algorithms and idioms, never for byte layouts.** When one field decodes weird,
   read the server's generator/parser — don't pattern-match a different dialect
   (`plutonium/README.md:27-29`).
2. **Never let a failure become a silent no-op.** Blind fallbacks ("walk anyway,"
   "send the action anyway," swallow `ErrNoPath`) don't add robustness; they convert
   diagnosable failures into mysterious live behavior — and on this wire a blind
   action even plants a latent server-side action that fires later. Delete the
   fallback, surface the typed error, let the layer above replan
   (`runtime/traverse.go:102-105`, `movement-design.md:222,396`).
3. **Fix at the truth-source layer first.** Grid before traversal, traversal before
   DSL surface, surface before pedagogy. Fixing a layer above the broken one
   relocates the bug and teaches the LLM confident lies
   (`dsl-synthesis.md:11,164`).
4. **When symptoms multiply, suspect convergence.** Weeks of "separate" wall, door,
   and plane bugs were one subsystem. Decompose by *mechanism* — the study split one
   symptom into three coupled defects — not by where the symptom was observed
   (`plutonium/README.md:43-50,60`).
5. **If the data is already decoded and stored, the bug is a missing wire-up.** Go
   looking for the consumer that never reads it before designing new capability
   (`README.md:61`).
6. **Confirm by state, never by timing.** Door = id overwrite (never removal,
   in-range), gate = def swap, ladder = plane change, walk = position delta.
   Anything timing-based will lie under load (`movement-design.md:103,217`).
7. **Model the world's real ontology.** A door is an edge between two tiles, not an
   object at a tile; doors and gates are different server dispatch channels; approach
   geometry must use the server's exact accepted stand rects, which have zero
   adjacency slack (`SESSION-STATE-2026-06-09…md:53`, `findings.md:9-15`).
8. **What the world refuses, remember.** Locked/toll/quest refusals become a
   blocked-edge ledger consulted by every future plan — re-planning through the same
   locked door forever is an epistemic bug, not a pathing bug
   (`runtime/obstacles.go:60-75`).
9. **Adversarial verification must include the correctness oracle.** A "verified"
   finding nearly corrupted the entity stream (the 5-bit→6-bit "fix"); it was caught
   only because verification re-derived the claim from server source instead of
   trusting the prior agent (`README.md:29,59`).
10. **Stop patching when you're reverting.** A patch-revert cycle on the same
    subsystem is the signal to stand up a ground-truth study and implement one
    design, once, in dependency order (`movement-design.md` §steps; commits
    `b24012c`/`6a329fd`).
11. **Ids and def names are traps — ground-truth them before navigation or
    interaction depends on them.** The verified #121 content table (bucket = item 21,
    NOT 52 — 52 is Silverlight; def-655 = the *Gnome Agility Course* log balance,
    not Al-Kharid; gold ore 152 adjacent to iron 151) is preserved verbatim in
    `docs/archive/initial-brainstorming/lang-build-backlog.md` §8 — the same
    discipline as rule 1, applied to content data.

## Sources

**Primary corpus** (the four-reference study, 2026-06-09):

- `docs/_research/plutonium/README.md` — the three verdicts, the byte-oracle caveat
  (and its own in-place correction), the 5-bit near-miss, the surprises list, the
  15-item backlog.
- `docs/_research/plutonium/movement-design.md` — the 8 evidence-pinned root causes,
  the 10-case obstacle taxonomy, the two-layer collision model, implementation
  steps 1–10, the four-tier test plan.
- `docs/_research/plutonium/findings.md` — all 78 verified findings with file:line in
  three repos; the server ground-truth corpus (`Mob.atObject` rects, ISAAC
  differential, opcode 211/91 semantics).
- `docs/_research/plutonium/dimensions.md`, `dsl-synthesis.md` — the 17 comparison
  dimensions; the layer-attribution doctrine ("the same bug at two altitudes").
- `docs/_research/SESSION-STATE-2026-06-09-forage-and-metacognition.md` — the live
  deadlock decomposition (§4), the #31 door-blind regression (§3), Alex's
  door-is-a-boundary correction (line 53).

**Commits** (all in this branch's history; dates verified):

- `772434d` (2026-05-27) — pathfinder born with defects B and C already in place.
- `69b0ba0` (2026-05-28) — `walk_to` auto-opens doors (on the broken grid).
- `a270236` (2026-06-08) — #31 perception reachability (shipped door-blind; the coda).
- `b24012c` (2026-06-09) — movement design steps 1–5.
- `6a329fd` (2026-06-09) — steps 6–10.
- `0214a0b` (2026-06-09) — fatigue→sleep arbiter + server-EOF supervisor restart.

**Code** (current tree):

- `pathfind/grid.go` — `BuildGrid` with live overlay + plane auto-derivation
  (`:107-111`), `stampSceneryFootprint` (`:243-256`).
- `runtime/traverse.go` — the classify→open→confirm→replan flow and the
  no-blind-fallback contract (`:90-105`).
- `runtime/obstacles.go` — refusal classifier + `BlockedEdge` ledger.
- `runtime/boundary.go` — server-exact `approachRect` (`:13-36`).
- `worldmap/worldmap.go`, `worldmap/reachable.go` — all-four-planes, gate-aware
  oracle.
- `proto/v235/inbound.go` — opcode 211 decode (`:59`, `:560`).
- `runtime/detour.go:273-279` — `fatigueArbiter`.
- `session/conn_eof_test.go:39` — the pinned EOF-restart regression test.

**A note on stale line numbers:** the study and design docs cite `runtime/host.go`
line numbers (e.g. the raw-Walk fallback at `host.go:1188`) from the pre-fix tree.
On 2026-06-10, `279298d` (RT-1) split `host.go` into `host/frame/traverse/session`;
the traversal flow now lives in `runtime/traverse.go`. When chasing an old cite,
resolve it against the commit the doc was written at (`git show b24012c~1:…`), not
the current tree.
