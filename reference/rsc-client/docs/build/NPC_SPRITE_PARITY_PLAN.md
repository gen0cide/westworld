# NPC / Player 2D-Sprite Parity — Implementation Plan

## RESULTS (2026-06-03) — Phases 0–4 DONE, full 3-way 1:1 for a static humanoid

Implemented via the `npc-sprite-parity-impl` workflow (Opus implement+verify per phase) plus
one coordinator-caught fix. **Independently re-measured first-hand** (rebuilt all three legs
from committed source, re-rendered, PIL max-channel diff — not trusting the verify agents):

| entity | orsc↔DEOB / orsc↔JAR / DEOB↔JAR |
|---|---|
| **NPC rat** (serverId 19, 1 layer, raw colour) | **0 / 0 / 0** |
| **default player** (head1/body1/legs1, 3-layer dye/skin) | **0 / 0 / 0** |
| scenery regression (ladder, gate off) | 0 / 0 / 0 (no regression) |

DEOB↔JAR PNGs are byte-identical; orsc is pixel-identical. The orsc per-layer 16.16
`spriteClipping` port (`render/orsc/entityblit.go`) is byte-faithful (isolated single-layer
test = 0px vs the DEOB oracle), so the whole-canvas-NN scaler was the entire on-screen gap.

**Reproduction (the flags matter):** fixture `testdata/rscdump/hunt/door_diag_obj.json`,
`RSC_MESH_CACHE=/tmp/rsc-run/cache` on all three, `ORSC_FLAT_AMBIENCE=1` for orsc.
- Rat: `RSC_MESH_NPC=19 RSC_NPC_PHASE2=1`.
- Player: `RSC_MESH_PLAYER=1` (NO `RSC_NPC_PHASE2` — orsc's `phase2 = gate && RSC_NPC_PHASE2`
  draws the *rat* when set, so the player must run without it). Dir is the optional 2nd field
  (`RSC_MESH_PLAYER=<gate>[:<dir>[:<step>]]`, default 0), mirroring the NPC gate.

**A coordinator catch worth recording:** the Phase-4 verify agent reported the player at
0/0/0, but my independent reproduction measured **880px** — it had run the player with
`RSC_NPC_PHASE2` set (so it measured the rat-as-player). The real residual was a **pose
mismatch**: orsc parsed the player facing dir from the *leading* `RSC_MESH_PLAYER` field, so
`=1` rendered the dir-1 (turned) frame while the oracle treats the value as on/off and renders
dir-0; different frame → different per-layer trims → 880px. Fixed in `render/orsc/entityspec.go`
`playerGateDir()` (read dir from the 2nd field, default 0) — commit `a32c391`. Lesson restated:
re-measure subagent parity claims first-hand, with the exact documented invocation.

**Honest residuals / follow-ups (not blockers):**
- **Flipped facings (dir 5–7) and non-zero frames are NOT 3-way verified** — the DEOB/JAR
  oracle harness hardcodes the dir-0 standing frame, so it cannot render a flipped/animated
  entity. orsc's flip path is implemented but only structurally exercised. (Plan Phase 5.)
- **Harness ergonomics wart:** orsc renders the Phase-0 *solid debug billboard* (cyan) for an
  NPC gate unless `RSC_NPC_PHASE2` is set, and the rat/player flags are mutually exclusive.
  Making the real sprite the default + retiring the debug billboard is a clean-up follow-up.
- The `door_diag_obj` *base scenery* (no entity gate) still shows orsc↔DEOB 2691px — the
  PRE-EXISTING diagonal-door-leaf divergence (orsc builds a synthetic leaf the def-less oracle
  doesn't), confirmed identical at HEAD before this work. Clean scenery fixtures are 0px.

Commits (westworld `feat/remote-client`): `d996f11` p0, `46523ef` p0.5, `63f6bcf` p1,
`9de7ae9` p2, `ff3140e` p3, `f6f7d85` p4, `a32c391` player-dir fix. JAR oracle (rscplus branch
`deob/npc-parity-phase0`): `91da6b8` p0, `14a36ae` p1, `62a577e` p3.

---

Status of the plan below: **DELIVERED** (was: PLAN, no code). Produced 2026-06-03 by the `npc-sprite-parity-research`
workflow (5 parallel Opus research lenses → synthesize → adversarial critique → finalize).
This is the implementation reference for extending the 3-engine pixel-parity rig
(orsc / rev-235 JAR / readable DEOB) from static scenery meshes to **2D-sprite humanoid
entities** (players + NPCs). Companion: `ORSC_JAR_PARITY.md` (terrain + scenery, all 1:1).

## How a humanoid reaches the screen (all three legs share this)
1. Per tick the client queues each player/NPC as a **2-vertex billboard face** in the
   Scene's `view` GameModel via `Scene.addSprite`, with a MAGIC sprite id (player=`i+5000`,
   npc=`20000+i`) and a world-space size (player 145×220; npc = `entityIndexTableC[id]` ×
   `legacyMaskTable[id]`). World pos = `(currentX, -getElevation, currentY)`.
   (`Mudclient.java:6512`)
2. `Scene.render` projects that billboard with the **same 3D camera** as terrain/scenery and
   depth-sorts it **with** the geometry (an NPC behind a wall is occluded). Per face it
   computes `w=(spriteW<<viewDist)/vz`, `h=(spriteH<<viewDist)/vz`, `x=vx-w/2`,
   `y=baseY+vy-h`, `scale=(256<<viewDist)/vz`, `skew=0`. (`Scene.java:1704,1841`)
3. The magic id dispatches to `drawPlayer`/`drawNpc` (`SurfaceSprite.spriteClipping` 8-arg),
   which loop **~12 body-part layers** `Tg[walkAnim][layer]` and blit each via
   `Surface.spriteClipping` (10-arg) — a scaled, skewed, two-tint transparent blit
   (`r==g==b → ×dye`; `r==255&&g==b → ×skin`; else copy). (`Mudclient.java:5655` player /
   `:5845` npc — note the method names are swapped vs their obf labels.)

## Feasibility verdict (honest, probe-backed)
- **(A) DEOB↔JAR on-screen 1:1 — SOLID, low risk.** Same content1 JAG sprites, same
  `parseSprite`, same 10-arg `spriteClipping`, same projection. No new algorithm.
- **(B) orsc↔DEOB composite-canvas 1:1 — SOLID, small.** Body-part sprite **pixels are
  byte-identical** across content1 JAG and OpenRSC `Authentic_Sprites.orsc`, and the
  sprite-selection index chain coincides (`spriteOffsets==animationNumbers`, `charColour`/
  `hasF` match — re-verified by probe). orsc already has a faithful compositor
  (`render/entitysprite.go`).
- **(C) FULL orsc on-screen 1:1 — FEASIBLE but MODERATE-TO-HARD, and required.** Probe-
  confirmed root cause: at the fixture camera the rat billboard downscales ~3× (346→~118px).
  orsc composites all layers at **integer** trim offsets then whole-canvas integer-NN scales
  (and discards skew/flip); DEOB/JAR apply each layer's trim in **16.16 fixed-point at blit
  time** with fractional carry. Under heavy downscale these sample different source columns
  at nearly every output column, and per-layer sub-pixel trim can't be recovered post-
  composite. So orsc on-screen 1:1 needs: (1) register **N raw layers per entity** (bypass
  pre-compositing), (2) a faithful 16.16 `spriteClipping`/`transparentSpritePlot` port in
  `drawEntity`, (3) restore flip/skew. Spans `render/` (decode) ↔ `render/orsc/` (blit).
  Everything else (decode, recolour rule, projection formula, palette tables, sprite index,
  billboard size) is already byte/formula-faithful.

## Single-source strategy
Point **all three** legs at the authentic **content1 JAG** (`/tmp/rsc-run/cache/content1_*`).
The Go side already has the archive reader (`assets/jag.go`) + a single-sprite decoder
(`render/orsc/textures_content11.go parseContent11Sprite`); only a multi-frame extension is
needed. Removes the dead `/Users/flint/...Authentic_Sprites.orsc` path; one canonical diff
source; zero algorithmic cost (pixels identical). Keep `Authentic_Sprites.orsc` as a
regression cross-check only.

## First test entity
**NPC Rat, content0 serverId=19**, single standing frame, dir=0 (south, un-flipped), step=0.
- ONE non-empty layer (`Tg` grid `[123,-1,-1,…]`) → collapses the 12-layer loop to one blit.
- animID=123; `spriteOffsets[123]=837==animationNumbers()[123]`; `charColour=4805259` (a raw
  24-bit value, **not** a dye marker 1/2/3 → dye path is identity); `hasF=false` (no +15).
- dir=0 → walkAnim=0, step=0, no flip; frame=`sf[0]+0=0`.
- Billboard 346×136 = content0 `entityIndexTableC[19]`/`legacyMaskTable[19]` == OpenRSC
  NpcDefs id=19 camera1/camera2 (probed) → identical rect in all three.
- **Off-by-one (must encode):** content0 grid stores the RAW animID; the DEOB/JAR NPC path
  reads `equippedItem[bodyPart]-1`, so the harness must synthesize `equippedItem[0]=animID+1`
  (=124). The PLAYER path reads the grid directly with **no** −1.

Second entity: default-human **player** (head1/body1/legs1 = animIDs 0/1/2, charColour markers
1/2/3) to exercise the 3-layer stack + dye/skin. Third: a flipped heading (dir 5–7).

## New tools to build
1. **Multi-frame content1 sprite decoder** (orsc) — extend `parseContent11Sprite` to a
   frame-block loop (~40 LOC); re-route `entitysprite.go decodeEntitySprite` off
   `Authentic_Sprites.orsc` to content1.
2. **Per-layer 16.16 `spriteClipping` blit** (orsc, Milestone C) — faithful Go port of
   `Surface.spriteClipping`(10-arg)+`transparentSpritePlot`, replacing the whole-canvas
   `drawEntity` + the pre-compositing `AddEntity` contract (N raw layers per entity), restore
   flip/skew. ~200 LOC, cross-package. **The one real chunk.**
3. **Shared static-entity spec** — drive off the existing rscdump `Entities[]` field
   (`internal/rscdump/schema.go`; orsc's `addViewEntities` already reads it; the Java legs
   parse the same JSON) so the spec is identical across legs with no env drift. Alt env gate
   `RSC_MESH_NPC=<serverId>[:<dir>:<step>]`.
4. **DEOB `engineArraySize` fix** (harness-side, do NOT edit `SocketFactory`) — reflection-set
   `StreamFactory.engineArraySize = CacheUpdater.contentNames.length` after `initGameData`
   (the src lost the write into a local at `SocketFactory.java:596`), else `loadEntitySprites`
   loops 0 times. Or hand-seed only the rat's 27-slot block.
5. **Projection-rect replicator** (all legs) — the draft's "read back projected w/h/x/y from
   `addSprite`" is **impossible** (`Scene.baseX/baseY/viewDistance` are private; the rect is a
   method-local). Instead: register the billboard, run the engine's projection, then
   reflection-read the projected vertex (`model.vertexViewX/Y/projectVertexZ`) + the private
   projection fields and recompute the 5-line rect; feed THAT rect to the per-layer
   `spriteClipping` (the entity face must NOT dispatch through `SurfaceSprite`, whose
   `.client` is null headless). orsc's `endScene` does this natively once Milestone C lands.

## Phased implementation (each phase gated on a measured diff)
- **Phase 0 — placement sanity:** blit ONE solid-colour billboard at the host tile in all
  three; PNG-diff the screen **rect** is pixel-identical. Isolates projection from decode.
  Also proves the Java reflection-read projection reproduces the engine's own rect.
- **Phase 0.5 — index assertion (already proven):** lock `spriteOffsets[123]==837`,
  `charColour`/`hasF`, content0 grid, and the `equippedItem` round-trip as a regression guard.
- **Phase 1 — DEOB↔JAR rat on-screen 1:1:** decode the rat frame-0 block from content1 into
  the Surface slots at 837, synthesize `equippedItem[0]=124`, blit at the Phase-0 rect. **=0.**
- **Phase 2 — orsc composite-canvas 1:1 (Milestone B):** switch orsc to content1, suppress
  demo entities, place the rat via `Entities[]`, byte-compare the UNSCALED composite canvas
  vs a DEOB-reconstructed canvas. **=0.** Isolates the scaler as the only remaining gap.
- **Phase 3 — DEOB↔JAR player + dye/skin:** 3-layer stack, grey→dye + skin recolour, per-layer
  trim. **NOTE** player dye arrays are `ClientStream.sharedIntArrayT`/`SocketFactory.itemSpriteIndex`/
  `Surface.unusedIntsAb`/`ChatCipher.unusedE` (NOT the NPC `ei/Dg/Wh`) — verify those equal
  orsc's tables first. **=0.**
- **Phase 4 — orsc on-screen 1:1 (Milestone C):** port the per-layer 16.16 blit (tool #2);
  first size the single-layer-rat residual at ~3× downscale, then close it; re-run Phases 1/3
  with orsc in the diff. Gate: orsc↔DEOB **=0** (or a documented, understood residual).
- **Phase 5 (stretch):** flipped heading (orsc must restore flip), F-frame (+15; needs a
  `hasF=true` entity), per-part hand/shield offsets.

## Effort
DEOB↔JAR subset (Phases 0/0.5/1/3): modest, low-risk — every primitive proven present.
orsc composite-canvas (Phase 2): small. Full 3-way on-screen incl. the orsc per-layer scaler
(Phase 4): the one moderate-to-hard effort. Overall ≈ a week of focused implementation,
front-loaded by the cheap Phase-0/0.5 de-risking.

## Key gotchas (all probe-confirmed this session)
- NPC `equippedItem=animID+1` synthesis vs PLAYER grid `=animID` (off-by-one).
- DEOB `engineArraySize` never assigned → harness reflection-set or hand-seed.
- Projection rect is private/local → register+project+reflection-read+recompute.
- JAR plain-Java `parseSprite` discards `fullHeight` (trim branch early-returns when
  `spriteHeightFull==0`) and is single-frame → extend both.
- Obf field collision: `Surface.spriteWidthFull=Eb` vs `Scene.spriteHeight=Eb` → scope
  reflection to the `ua`-declared field on the `ua` instance.
- Player dye arrays ≠ NPC dye arrays (provenance differs; verify before the player compare).
