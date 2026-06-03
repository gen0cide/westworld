# NPC / Player 2D-Sprite Parity — Implementation Plan

## RESULTS — full 3-way 1:1 across the WHOLE 2D-sprite surface (static + flip + anim + F-frame + ground items)

Phases 0–4 (2026-06-03): static humanoid 1:1. Phase 5 + ground items + ergonomics (2026-06-03,
the completion push): flipped facings, animation steps, the +15 F-frame, and dropped ground items,
all 3-way 1:1. Implemented via a research→critique cascade then sequential implement→verify workflows
(Opus) plus two coordinator-caught fixes. **Every headline re-measured first-hand** (rebuilt all three
legs from committed source, re-rendered, max-channel diff via `tools/parity3.sh` — not trusting the
verify agents):

| entity / pose | orsc↔DEOB / orsc↔JAR / DEOB↔JAR |
|---|---|
| **NPC rat** dir 0 (serverId 19, 1 layer, raw colour) | **0 / 0 / 0** |
| **NPC rat** flip dir 5/6/7, turned 1/2/3, anim step 1/2 | **0 / 0 / 0** |
| **default player** dir 0 (head1/body1/legs1, 3-layer dye/skin) | **0 / 0 / 0** |
| **default player** flip dir 5/6/7 | **0 / 0 / 0** |
| **NPC skeleton** dir 5/6/7 (serverId 45, weapon+body, **+15 F-frame**) | **0 / 0 / 0** |
| **NPC goblin** dir 5 (serverId 4, gobweap+body, +15 F-frame) | **0 / 0 / 0** |
| **ground item** Logs (id 14, no mask) + Iron scimitar (id 82, grey-tint mask) | **0 / 0 / 0** |
| walk-wrap (`19:0:3` ≡ `19:0:1`, sf[3]=sf[1]) | byte-equal per leg |
| regressions (rat/player dir 0, items 14/82, ladder scenery) | 0 / 0 / 0 |

DEOB↔JAR PNGs are byte-identical; orsc is pixel-identical. The orsc per-layer 16.16
`spriteClipping` port (`render/orsc/entityblit.go`) is byte-faithful (isolated single-layer
test = 0px vs the DEOB oracle), so the whole-canvas-NN scaler was the entire static on-screen gap.

**The completion push (C→B→A):**
- **C — ergonomics cleanup** (`096c65a`): real composited sprite is the **default** for an NPC gate;
  `RSC_NPC_PHASE2` retired to an accepted no-op; cyan debug billboard behind `RSC_NPC_DEBUG_BILLBOARD=1`;
  NPC+player gates independent (both can co-place).
- **B — ground items** (`7acc1ca` DEOB, `14293a8` orsc; rscplus `3fb9ca6` JAR): a dropped item is a 2D
  billboard via `Scene.addSprite(40000+id, …, 96, 64, …)`. orsc draws it through the **raw 1-layer 16.16**
  path (`AddEntityLayers`), recolouring via `transparentSpritePlot` (`Dye=pictureMask, Skin=0`) — **not** the
  48×32 `compositeItem`/`recolorItemPixel` path (whose blueMask branch the authentic rev-235 draw never uses).
  Icons decode from **content8** ("2d graphics", numbered `objects{N}.dat` sheets, 30/sheet). content8≡
  `Authentic_Sprites` byte-verified for item 14; picker AABB reconciled to the literal 96×64.
- **A — flip / anim / +15 F-frame** (`a854c88` orsc, `168d82f` DEOB, `6811524` orsc fix; rscplus `a63227b` JAR):
  both Java oracles learned a **computed** frame `= sf[step&3] + 3·col` (`sf={0,1,2,1}`, col = `walkAnim` with
  5/6/7→3/2/1+flip), the per-dir `Tg[walkAnim]` layer order, the `+15` F-frame, the f.dat sub-block decode, and a
  def-driven billboard (sprites/colours/`camera1×camera2`) keyed by the gate id. orsc's flip path already
  existed; the env `:step` (raw walk-cycle index) now resolves through `npcWalkModel={0,1,2,1}` into `StepPhase`.

**Reproduction (the flags matter):** fixture `testdata/rscdump/hunt/door_diag_obj.json`,
`RSC_MESH_CACHE=/tmp/rsc-run/cache` on all three, `ORSC_FLAT_AMBIENCE=1` for orsc.
- Rat: `RSC_MESH_NPC=19` (the real composited rat is now the **default** gate behaviour — no
  `RSC_NPC_PHASE2` needed). `RSC_NPC_PHASE2` is a **retired, accepted no-op alias**: it is never
  read as a gate but is tolerated without error, so the muscle-memory `RSC_MESH_NPC=19
  RSC_NPC_PHASE2=1` still renders the real rat byte-identically.
- Player: `RSC_MESH_PLAYER=1`. Dir is the optional 2nd field
  (`RSC_MESH_PLAYER=<gate>[:<dir>[:<step>]]`, default 0), mirroring the NPC gate. The NPC and
  player gates are now **independent** — setting both places both entities (they share the host
  centre tile and may overdraw).
- Placement-sanity opt-in: `RSC_NPC_DEBUG_BILLBOARD=1` (any non-empty value) makes an engaged NPC
  gate draw the **solid cyan debug billboard** instead of the real sprite, isolating the
  projection (screen w/h/x/y) from sprite decode/composite. This is a retained diagnostic, no
  longer the default.

**A coordinator catch worth recording:** the Phase-4 verify agent reported the player at
0/0/0, but my independent reproduction measured **880px** — it had run the player with
`RSC_NPC_PHASE2` set (so it measured the rat-as-player). The real residual was a **pose
mismatch**: orsc parsed the player facing dir from the *leading* `RSC_MESH_PLAYER` field, so
`=1` rendered the dir-1 (turned) frame while the oracle treats the value as on/off and renders
dir-0; different frame → different per-layer trims → 880px. Fixed in `render/orsc/entityspec.go`
`playerGateDir()` (read dir from the 2nd field, default 0) — commit `a32c391`. Lesson restated:
re-measure subagent parity claims first-hand, with the exact documented invocation.

**Two coordinator catches worth recording (the completion push):**
1. The Phase-5 verify agent reported the skeleton +15 F-frame at 0/0/0 in some framings but the full
   3-way gate measured **365px** (orsc↔oracles); my first-hand re-measure confirmed the residual. Root
   cause (the fix agent corrected my own hypothesized diagnosis when told to *verify, not assume*): orsc's
   content1 decoder only read the 15-frame **body** block, so the +15 flipped F-frame (frame 24, in the
   `<name>f.dat` 9-frame sub-block at slots 18–26) never decoded and the weapon layer was silently dropped
   — orsc drew body-only. Fixed by assembling the full 27-slot block (15 body + 3 `a.dat` + 9 `f.dat`),
   mirroring `loadEntitySprites` — `6811524`. A before/after proof showed post-fix orsc differs from pre-fix
   by *exactly* the 365 weapon pixels.
2. The F-frame was once thought "blocked on data not in the repo." It was not: the real NPC def table is
   local (orsc `NpcDefs.json`; Java legs `content0`), `camera1×camera2` = billboard W×H (proven by rat
   id19 = 346×136), skeleton id45 = `sprites[134,133]`/216×234/colours0. No fetch needed.

**Honest residuals / follow-ups (not blockers):**
- ~~**Flipped facings (dir 5–7) and non-zero frames are NOT 3-way verified.**~~ **RESOLVED** (Phase 5): rat
  flip/turned/anim, player flip, and the skeleton/goblin +15 F-frame are all full 3-way 0/0/0. Drive via
  `RSC_MESH_NPC=<id>:<dir>[:<step>]` / `RSC_MESH_PLAYER=<gate>:<dir>[:<step>]`; `:step` is the raw walk-cycle
  index resolved through `sf={0,1,2,1}` on all three legs (so `:step 3 ≡ :step 1`).
- **GAP B — NPC flipped hand/shield per-part offsets** (`Mudclient.java:5888-5910`, bodyPart 3/4 dx/dy when
  `entityFlags!=1`): unported in orsc. No entity in the verified matrix exercises it (the hasF beasts —
  skeleton/goblin/zombie — carry their weapon in slot 0 and take the clean +15 branch). Documented residual.
- **orsc per-layer FullW architectural nuance:** orsc's `spriteClippingLayer` derives each layer's U-step from
  *that layer's own* FullW vs a shared projected `w`, whereas authentic uses a per-item base (frame-0) FullW.
  These coincide whenever an NPC's layers share one full-canvas width (true for every tested entity: skeleton/
  goblin both 108) — so all tested poses are 0/0/0 — but would diverge for a hypothetical NPC whose layers had
  genuinely different full-canvas widths. Not exercised by any current gate; documented.
- **Ground items — blueMask + multi-drop stacks:** only blue==0 items (14, 82) are verified. The authentic
  ground/inventory draw uses grey-tint(colour1)+skin(colour2) only — never blueMask — so blueMask (potion-class)
  items are an untested residual (orsc's `recolorItemPixel` blueMask branch is off the billboard path). Multi-drop
  Y-stacks (`Le[i]≠0`) are untested (single drop only).
- **Player animation beyond dir-flip:** the player base-frame divisor is `sectorAlloc[serverId]` vs the NPC's
  fixed `/6`; the rig sidesteps it by driving `:step` directly (the divisor never enters), so player walk-cycle
  *via movingStep* is not exercised. Documented.
- ~~**Harness ergonomics wart:** orsc renders the Phase-0 *solid debug billboard* (cyan) for an
  NPC gate unless `RSC_NPC_PHASE2` is set, and the rat/player flags are mutually exclusive.
  Making the real sprite the default + retiring the debug billboard is a clean-up follow-up.~~
  **RESOLVED:** `RSC_MESH_NPC=<id>` now draws the real composited rat by default; `RSC_NPC_PHASE2`
  is a retired accepted-no-op alias; the NPC/player gates are independent (both can be set); and
  the cyan debug billboard moved behind the explicit `RSC_NPC_DEBUG_BILLBOARD=1` opt-in.
- The `door_diag_obj` *base scenery* (no entity gate) still shows orsc↔DEOB 2691px — the
  PRE-EXISTING diagonal-door-leaf divergence (orsc builds a synthetic leaf the def-less oracle
  doesn't), confirmed identical at HEAD before this work. Clean scenery fixtures are 0px.

Commits — static humanoid (westworld `feat/remote-client`): `d996f11` p0, `46523ef` p0.5, `63f6bcf` p1,
`9de7ae9` p2, `ff3140e` p3, `f6f7d85` p4, `a32c391` player-dir fix. JAR oracle (rscplus branch
`deob/npc-parity-phase0`): `91da6b8` p0, `14a36ae` p1, `62a577e` p3.
Completion push (westworld): `5fe159b` parity tooling, `096c65a` C ergonomics, `7acc1ca`+`14293a8` B ground items,
`a854c88`+`168d82f` A flip/anim/F-frame, `6811524` A F-frame fix. JAR oracle (rscplus): `3fb9ca6` B, `a63227b` A.
Verification harness: `tools/parity3.sh` (3-leg renderer, JDK17-pinned) + `tools/pngdiff.py` (max-channel diff).

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
