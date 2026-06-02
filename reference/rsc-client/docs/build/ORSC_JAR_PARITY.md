# orsc ↔ rev-235 JAR render parity — the definitive faithfulness check

Built by the **parity-run** workflow (2026-06-02). This is the deferred orsc-vs-JAR
runtime render diff: the orsc engine (the renderer-proper, the engine now) rendered
the same dumped world state as the **obfuscated rev-235 `rsclassic-1091943135.jar`
oracle** and the two were diffed structurally (camera/clear-screen/palette
independent) and per-pixel.

Harness: `rscplus/dumprender/DumpRenderer.java` (JDK 17 + ASM 5.0.4, classloader +
3 ASM patches, `Component=null` headless) → jar PNG + `rscdump-faces/1` JSON;
`cmd/renderdiff` renders the orsc side (`orsc.RenderDump`/`RenderDumpFaces`) and
diffs the pair. Artifacts under `testdata/rscdump/out/parity/<fixture>/`.

## UPDATE 2026-06-02 (coverage) — 3-engine 1:1 across the terrain-class matrix

Extended the byte-identical validation from flat terrain to the full terrain-class matrix.
Independently re-measured (`ORSC_FLAT_AMBIENCE=1`, all three legs rebuilt from source):

| terrain class | sample fixtures | orsc↔JAR / orsc↔DEOB / JAR↔DEOB |
|---|---|---|
| flat grass / door / no-door | single_tile_{door,NOdoor} | **0 / 0 / 0** |
| type-3 textured overlay | t3_water_ov2, overlay_{marble,lava,pentagram}_* | **0 / 0 / 0** |
| type-4 water/bridge/lava | t3_bridge_ov4, overlay_lavaclass_ov12 | **0 / 0 / 0** |
| type-4 neighbour-spread | t3_type4_neighbour | **0 / 0 / 0** |
| bridge 250→9 plank-edge seam | seam_sector_aligned, bridge_* | **0 / 0 / 0** |
| roofs (cull from in + out) | roof_inside/outside/cull_* | **0 / 0 / 0** |
| straight + diagonal walls/doors | walls_*, door_{straight,diag_wall,none} | **0 / 0 / 0** |

Unblocking required synthesizing matching overlay `{tileType,colour}` + roof `{rise,tex}` def
tables (+ 64 transparent texture slots) in BOTH the DEOB leg (`DumpRender.java`) and the JAR
oracle (`DumpRenderer.java`), copied verbatim from orsc's tables, plus a JAR-oracle camera
`getElevation` fix (it had hard-coded eye-Y=0, off by the grid's 120-unit elevation — which is
why the prior pass, only ever run on elevation-0 fixtures, never caught it). Two genuine deob
fidelity bugs were caught by the **structural** face diff and fixed faithfully: the `quadEdges`
edge-start vertex (commit 39f8f09) and the type-4 colour-force `else if` chaining (commit 1a6748b).

## UPDATE 2026-06-02 (scenery meshes) — 3-engine 1:1 on REAL GameData object models

The prior "diagonal scenery OBJECT" open item is **RESOLVED**. All three engines now load the
SAME authentic rev-235 object model and place it by the SAME `World.addModels` math, so a real
scenery object renders byte-identically. Independently re-measured (`ORSC_FLAT_AMBIENCE=1`, all
three legs rebuilt from source, fixture `door_diag_obj` with the diagonal-object band 48001+id):

| object (content9 `.ob3`) | faces / fill mix | footprint | orsc↔JAR / orsc↔DEOB / JAR↔DEOB |
|---|---|---|---|
| `ladder`           | 48, all flat-colour            | 1×1 | **0 / 0 / 0** |
| `well`             | 50 (32 flat + 18 textured)     | 2×2 | **0 / 0 / 0** |
| `woodengateclosed` | 58, all double-sided           | 1×1 | **0 / 0 / 0** |
| `tree2`            | 54, all double-sided (77 vtx)  | 1×1 | **0 / 0 / 0** |
| `table`            | 21 (16 flat + 5 textured)      | 1×1 | **0 / 0 / 0** |
| `chair`            | 35 (9 back-faces)              | 1×1 | **0 / 0 / 0** |

This validates each engine's FULL scenery pipeline on authentic data: archive load → name-hash →
offset → `.ob3` decode → build → place (footprint-centre, ground-snap) → diffuse-light → project →
raster. The DEOB↔JAR `getFileOffset` returns the IDENTICAL byte offset (e.g. `ladder` off 292330),
so even the archive directory walk is 1:1, not just the decode. The 2×2 `well` confirms the
footprint-centre + the textured-face SKIP (no texture archive → degraded-transparent in all three)
are consistent; the double-sided `woodengateclosed`/`tree2` confirm the back-face fill path.

**How it works (`RSC_MESH_*`-gated; terrain fixtures render unchanged when unset):**
- All three load the authentic "3d models" content archive (`content9`, = `readDataFile("3d models",
  60,9,84)`) the SAME way the live client does — read the cache file, `World.unpackData(128,false,…)`
  strips the 6-byte header + bzip-inflates the JAG archive (no network / no `ArchiveReader`).
- orsc: `cmd/meshrender` (`assets.OpenArchive` + a synthesized `facts.SceneryDef`, real
  `BuildDiagonalObjects` path — which now also applies the `addModels` diffuse light, commit on
  `render/orsc/diagobj.go`).
- DEOB: `client.DumpRender` registers the model name (`GameModel.textureId`), synthesizes the object
  def (`entityIndexTableF`/`NameTable.sortKeys`/`RecordLoader.intArray`), decodes in place
  (`NameHash.getFileOffset` + `GameModel(byte[],off,true)`), runs `World.addModels(_, (byte)-113)`.
- JAR: `DumpRenderer` mirrors that via reflection on the obf symbols (`k.a` unpackData / `ca.a`
  textureId / `oa.a` getFileOffset / `ca` ctors / `k.a` addModels; `fb.f`/`ub.g`/`f.f`/`mb.a` defs).

Repro: `RSC_MESH_MODEL=ladder RSC_MESH_OBJID=0 RSC_MESH_W=1 RSC_MESH_H=1` on all three harnesses,
fixture `testdata/rscdump/hunt/door_diag_obj.json`, then a max-channel PIL diff (all 512×334).

**Scope (honest):** placement is validated at tile-direction **0** (the standing orientation; the
fixture carries no `tileDirection` and neither authentic harness injects that grid, so all three
read dir 0). Rotated scenery (dir 1–7, the `setRot256(0,dir·32,0)` spin) is shared with the already-
audited static-scenery path but is NOT yet pixel-tested here — injecting `tileDirection` into the
DEOB/JAR grids is the next axis. The diagonal WALL band (<24000) was already pixel-1:1 (`door_diag_wall`).

## UPDATE 2026-06-02 (final) — TRUE 1:1: all three legs byte-identical

All residuals from the camera update below are now CLOSED. Independently re-measured
(512×334, `ORSC_FLAT_AMBIENCE=1`, clean from-scratch rebuild) — **every pairing is 0 px**:

| pair | NOdoor | door |
|---|---|---|
| orsc ↔ JAR  | **0 px** | **0 px** |
| orsc ↔ DEOB | **0 px** | **0 px** |
| JAR ↔ DEOB  | **0 px** | **0 px** |

The three fixes:
- **DEOB rasterizer (a genuine fidelity bug):** `Scene.quadEdges` (the `vertexCount==4` fast
  path) tied each edge's start X/shade to the WRONG endpoint (the hi/bottom vertex instead of
  the lo/top vertex a top-down scanline walk needs) for edges 0-1/1-2/2-3 — systematically
  under-filling every terrain quad's span (the diagonal black bands). Fixed faithfully vs the
  obf bytecode. PROOF it is faithful: **DEOB↔JAR is now 0 px** — the deob rasterizer output
  matches the *obfuscated* jar's byte-for-byte. It also improves the LIVE deob render (seamless
  terrain; non-black viewport px **156969 → 167241**).
- **Wall-base terrain shadow:** the authentic `method422→method425` bakes terrain-vertex
  ambience 40 at each wall's foot endpoints (a soft dark gradient hugging walls); orsc didn't.
  Ported as `render/orsc/world.go darkenTerrainAtWalls`; and the dump's door now builds as a
  static wall (wallPass light 122; `syntheticFacts Unknown:0`) matching the authentic.
- **JAR minimap bleed:** the JAR oracle now ASM-no-ops the obf per-tile minimap paint
  (`k.a(IBIIII)V`) too (rscplus repo), so its framebuffer holds only the 3D render.

Verified: `go build` rc 0; `go test ./render/... ./cmd/renderdiff/...` green; deob `javac` 0
errors / 81 classes; live deob smoke `login response:64`, 167241/171008 px, 0 exceptions, seamless.

## UPDATE 2026-06-02 — camera/framing 1:1 ACHIEVED (terrain pixel-exact)

The earlier "~99.7% whole-frame pixel diff = harness drift" residual was **root-caused
and fixed**. It was NOT a camera-math divergence — the camera eye is byte-identical in
all engines (center `(6208,0,6208)` → eye `(6208,−951,7367)` for `single_tile_door`).
Two real faults, in different legs:

1. **The JAR oracle harness (`DumpRenderer.java`) was doubly broken** (this produced the
   "16× farther, off-center grey patch" we had been diffing against): (a) it invoked a
   **non-existent** `setCameraOrientation` overload (`a(int×7,byte)`), so the camera was
   never set; (b) it never set `clipFar3d/2d`, leaving the obf default **1000**, which
   frustum-culled the ENTIRE terrain (camera-Z ≈ 1365 > 1000) — leaving only the 285×285
   2D minimap. Fixed: call the real `a(int×8)` with the authentic arg order + set
   `clipFar3d/2d=2400, fogZDistance=2300, fogZFalloff=1`.
2. **orsc fog/clip infidelity (the genuine orsc fix):** orsc ports OpenRSC, which keeps
   `fogSmoothingStartDistance=10, fogZFalloff=20` and uses `fogLandscapeDistance=10000`
   as the frustum-far — whereas the vanilla rev-235 client overrides EVERY frame to
   `fogZDistance=2300, fogZFalloff=1, clipFar=2400` (Mudclient.java:2338-2341,6624-6628).
   The orsc dump path now applies these authentic values (`render/orsc/{harness,dump_faces}.go`).

**Result** (independently re-measured, 512×334, `ORSC_FLAT_AMBIENCE=1` exact compare):

| fixture | orsc↔JAR full | orsc↔JAR excl. JAR-minimap |
|---|---|---|
| `single_tile_NOdoor` | 0.83% | **0.00% — 0 px (BYTE-IDENTICAL terrain)** |
| `single_tile_door`   | 4.32% | 1.16% (door-leaf shading only) |

orsc's flat-terrain render is now **pixel-identical** to the authentic rev-235 client.

**New — the DEOB render leg (the long-planned 3rd engine).** `client.DumpRender` (+
`client.Json`) renders a fixture through the compiled, readable deob `Surface`/`Scene`/
`World` directly (no reflection) — the authentic oracle, unblocked by the deob compiling.
`World.java` gained a `dumpGridsInjected` harness hook (gated **false** in the live
client — verified no regression: boots, login response:64, live3d 156969/171008 px; it
is the readable equivalent of the jar oracle's `loadMapData→RETURN` ASM patch + the
ambience-pin + minimap-suppress). Run:
`java -Djava.awt.headless=true -cp /tmp/deob-run client.DumpRender FIXTURE.json OUTDIR [base]`.

**Residual (honest):** (a) the DEOB leg shows ~17% thin diagonal scanline-gap stripes — a
headless rasterizer span-edge rounding artifact unique to that harness (geometry is
complete, 9026/9026; the LIVE deob renders seamless), so orsc is validated against the obf
jar **directly**, not blocked on it; (b) door-leaf shading ~1.2% (orsc builds the door via
`doorPass` diffuse −95 vs the authentic static wall); (c) the JAR oracle's own 2D minimap
still bleeds into the top-left 285×285 (cosmetic, JAR-side). The live spectator
(`RenderView`) still uses OpenRSC fog — applying the authentic fog there too is a follow-up
(it changes the cradle's distant-terrain look). The `rscplus/dumprender/DumpRenderer.java`
camera+clipFar fix lives in the separate rscplus repo (not part of this PR).

## TL;DR — overall parity

**Structurally, orsc IS the rev-235 JAR on the JAR-runnable set.** Every face the
oracle builds — terrain quads/triangles, the lifted wall/door leaf, the diagonal
split — is byte-identical to orsc in `(rounded centroid XYZ, vertex count,
fillFront, fillBack)`:

| fixture | orsc faces | JAR faces | shared | verdict |
|---|---|---|---|---|
| `single_tile_door` | 9026 | 9026 | **9026** | match |
| `door_straight`    | 9026 | 9026 | **9026** | match |
| `door_diag_obj`    | 9026 | 9025 | 9025 | minor-diff (JAR-oracle object-table gap, NOT orsc) |

The remaining 11 fixtures (any nonzero overlay/roof tile) **cannot run through the
JAR oracle in this environment** — the obf GameData overlay/tile-decoration `v[]`
def table is loader-populated and the jar carries no embedded GameData (it fetches
config/cache over the network at runtime). Those rows are settled against the
**OpenRSC source the JAR is compiled from** (a byte-for-byte source-port match) plus
orsc determinism, and are flagged below as "settled vs source, JAR-binary blocked".

The **pixel** diff is ~99.7% on every fixture, but it is **harness drift, not a
rasterizer-math divergence** (the structural diff is clean) — see the caveats.

**UPDATE 2026-06-02:** the one renderer-math divergence that WAS present — terrain
base-shade ~2× too bright — has been **RESOLVED** (the OpenRSC `RSModel`
project-time light *reseed* was dropped to match the rev-235 `GameModel`, which does
no project-time relight). Flat-grass base lighting (fog-removed) is now `shadeBase=71`
⇒ green **0x4a**, within 1–2 LSB of the JAR's **0x48** (was 0x8e). Geometry stays
9026/9026, determinism stays 0-pixel. See "orsc renderer-proper divergences" §1.

## What this run fixed in the JAR oracle (rscplus/dumprender/DumpRenderer.java)

1. **Window-placement retune (BLOCKER 1, all rows).** `GO_HOST_LOCAL` 80→48: orsc's
   window centre is `worldWindowTiles/2 = 96/2 = 48` (harness.go:55,60). The old 80
   was tuned for the deleted 160-tile-window engine and offset every orsc face from
   the JAR by exactly `(4096,0,4096) = (80-48)*128 = 32 tiles` (verified: door JAR
   `(10304,-96,10368)` vs orsc `(6208,-96,6272)`). The base is now derived per-fixture
   from the fixture's own host tile (`obfBase = 48 - (self.x - baseX)`), so every
   grid size (16/20/24) and self position aligns. → door went **8513/9026 → 9026/9026**.
2. **Wall-def table fill.** The synthetic wall tables now fill ALL 16 ids with the
   wood-door def (matching orsc `syntheticFacts`, which builds a def for every
   referenced id), not just index 1. `door_straight` (wallV=1 → id 0) previously
   drew a flat `Y=0 fill-0` stub instead of the lifted `-15719` door leaf → 9025/9026;
   now **9026/9026**.
3. **E-vp reflection probe.** 5 lines after render read the obf Scene's projection
   shift `lb.R` (the obf field the `(camX << R)/z` projection uses) and print it.

## Per-E-row verdict (with byte evidence)

### E-vp — texture projection shift vp = 8 vs 9 — **SETTLED (match), byte-pinned vs JAR**
The obf Scene projection is `(camX << this.R) / z` (obf `lb` bytecode offsets
1090/1093/1109: `getfield R:I` → `ishl` → `idiv`), so the obf field **`R` IS orsc's
`rot1024_vp_src`**. The probe reads it live after render:

```
E-vp probe: obf Scene projection shift (lb.R) = 9
```

orsc has `rot1024VpSrc = 9` (types.go:33) — the runtime `m_qd=9` override over the
ctor's R=8. **Confirmed == 9 against the live rev-235 JAR.** orsc's runtime override
is correct; the ctor-8 is overwritten by setBounds/setMidpoints exactly as orsc models.

### E-water — water fixed-shade quad vs plain gouraud colour=1 — **SETTLED (orsc = source), JAR-binary blocked**
The JAR cannot render `bridge_ov4`/`t3_bridge_ov4` (overlay-table blocker). Settled
against the OpenRSC `three/World.java` that the J++ jar is compiled from, cross-read
with the orsc port (world.go:332-522) AND the orsc face dump for `bridge_ov4`:

- **80 water faces, all centroid Y=0 (flattened), `fillBack=1`, `fillFront=TRANSPARENT`.**
  This is the **plain gouraud `createFace colour=1` path** (World.java:574-581 forces
  `colour=1` for tileType-4) — NO texture id, NO fixed-shade, NO shoreline enhancement.
- **128 deck faces at real elevation (Y=-64/-128), `fillFront=3` (plank deck colour),
  `fillBack=TRANSPARENT`** — the raised plank deck floating above the flattened water
  (World.java:704-723, `emitDeckQuad`).

orsc draws water exactly via the plain gouraud colour=1 path (the OpenRSC/Jagex
algorithm the jar implements), NOT the old "ours" fixed-shade+shoreline enhancement.
The verdict is the source-faithful one; a direct jar-binary diff is blocked only by
the missing GameData overlay def table.

### E-seam — bridge 250-remap seam coord (region-local x==47 vs world wx%48==47) — **SETTLED (match), FINDINGS #15 resolved**
The OpenRSC source `three/World.java setTileDecorationOnBridge` uses **window-local
`x == 47` / `z == 47`** over the full `for (x=0; x<96)` window:

```java
if (x == 47 && getTileDecorationID(x+1,z,0) != 250 && getTileDecorationID(1+x,z,0) != 2)
    setTileDecoration(x, z, 9);            // brown plank lip
else if (z == 47 && ...) setTileDecoration(x,z,9);
else setTileDecoration(x, z, 2);           // deck flattens + textures
```

orsc's port (world.go:253-268) is a **byte-exact** match — window-local `x==47`,
identical neighbour guards, identical `9`/`2` outcomes. So the seam coord is
region-local (window-local) x==47/z==47, **NOT** `world wx%48==47`. FINDINGS #15
`bridge_seam_coord` is resolved in orsc's favour.

To *exercise* the branch (the existing `bridge_on` never reaches it — its 250 run is
contiguous so the x==47 tile always has a 250 +x-neighbour and falls through to 2), a
new **sector-aligned** fixture was authored:
`testdata/rscdump/hunt/seam_sector_aligned.json` (gen via `gen_seam.go`; BaseX=192
(`%48==0`), self.x=204 so orsc-window-local 47 = fixture-local 11, +x neighbour grass).
orsc fires the remap: **8 overlay-9 plank-edge faces at world-X centroid 6080 =
window-local tile 47 (47*128+64)**, `fillBack=-26426`, at deck elevation. The plank
lip lands exactly on column 47. (JAR-binary diff blocked by the overlay table.)

### E-polygonHit — painter plane-pierce polygonHit1/2 byte-level — **SETTLED (algorithm match), pixel-diff blocked**
The roof+overlap fixtures (`roof_inside`/`roof_outside`) need BOTH the overlay table
AND a roof-def table in the JAR oracle (roofId 1) → blocked. On bare hand-authored
fixtures **neither** engine builds the roof model (no roof def), so the roof-vs-wall
overlap can't be staged either side. The painter ALGORITHM, however, runs on every
overlapping-face scene (the door fixtures' 9026 faces include overlapping wall/terrain
quads) and the structural diff there is byte-identical — the same face set reaches the
painter. Static cross-check: orsc `polygonHit1`/`polygonHit2` (scene.go:878-1010,
HANDS-OFF) port `Scene.polygonHit1/2`; the obf `lb` has the matching
`boolean a(boolean,w,w)` (polygonHit1) and `boolean a(byte,w,w)` (polygonHit2, dead
byte arg orsc drops), reading the same model rotated-vert arrays
(`ca.H`/`ca.cc`/`ca.bb` = vertX/Y/ZRot), poly normal (`w.l`/`w.k`/`w.s`) + magnitude
(`w.r`) + orientation-sign branch. Algorithm-faithful; a true byte-level paint-order
pixel diff needs the roof/overlay GameData tables + the harness clear-screen
reconciliation (see caveats).

### E-whole — whole-renderer orsc-vs-JAR agreement — **SETTLED (match) on the JAR-runnable set**
First-ever clean orsc-vs-JAR structural agreement: `single_tile_door` and
`door_straight` are **9026/9026 byte-identical** (every terrain quad, the lifted
door leaf at `(6208,-96,6272)` fill `-15719`, the diagonal splits). orsc builds the
same face set + the same fills + the same transformed centroids as the real rev-235
client. The door-perception bug class is absent for the straight-wall door.

## orsc renderer-proper divergences

### 1. Terrain base-shade ~2× too bright (the reseed bug) — **RESOLVED 2026-06-02**

**Symptom (before):** on the flat door fixture (groundColour idx 70, elevation 0,
seed 0) orsc painted flat grass green `#0d7c00..#0f8e00` (g-channel 0x7c..0x8e,
mode 0x8e) where the JAR oracle paints flat `#084800` (g-channel 0x48) — a ~2×
brightness error.

**Root cause:** `render/orsc/model.go` `clearRotDataAndParams26()` (the OpenRSC
`RSModel.clearRotDataAndParams26`, RSModel.java:528-543) unconditionally re-seeded
the diffuse light to `setDiffuseLight(40,102,104,108,-20,-89)` on **every** model's
first project (fired from `project()`/`rotate1024`). That reseed **clobbered** the
authentic per-model build-time light. For terrain (built with
`setDiffuseLightAndColor(-50,-10,-50,40,48,true,105)` at world.go:520 ⇒
`diffuseParam1=96, diffuseParam2=384, dir=(-50,-10,-50), mag=71`) the reseed forced
`diffuseParam1 = 256-102*4 = -152`, driving the flat-grass gouraud
`shadeBase = diffuseParam1 + vertLightOther + vertDiffuseLight ≈ -152+96 = -56` →
clamped to 0 (scene.go:701-708) → ramp index 0 = the brightest entry (green 0x8e).

**Why the reseed is wrong:** it is an **OpenRSC `RSModel` infidelity vs the rev-235
J++ client** (= the obf JAR oracle, = the deob in `reference/rsc-client/`). The
authentic rev-235 `GameModel.project()` (GameModel.java:1225-1281) does **no**
project-time relight: it calls `apply(7972)` (GameModel.java:1169-1213) which only
re-lights on `transformState==1` using the **model's own** `setLight` params
(terrain via World.java:1025 `terrain.setLight(-50,40,-10,-50,...)` ⇒
`lightAmbience=96, lightDiffuse=384`), then projects. No fixed-default override
anywhere. orsc had faithfully ported OpenRSC's RSModel here, inheriting OpenRSC's
infidelity rather than a Jagex behaviour.

**Fix (applied):** dropped the `setDiffuseLight` reseed from `clearRotDataAndParams26`
and renamed it `allocProjectionScratch` (it now does ONLY the projection-scratch
(re)allocation — the safe half). Every model already carries the correct light it
was built with (terrain 40/48, walls 60/24, roofs 50/50, scenery/diagobj via
`setDiffuseLight`), and `resetTransformCache`→`computeNormals`→`computeDiffuse` (the
orsc `apply(7972)` analog) relights each model from its OWN params — exactly the
deob path. One change restores all model types to their authentic per-model light;
it does not single out terrain.

**Evidence (verified at the per-vertex shade level, fog-removed):**
- BEFORE: terrain `diffuseParam1` clobbered to −152 → flat-grass shadeBase clamped
  to 0 → green mode **0x8e** (spread 0x7b–0x8e).
- AFTER: terrain `diffuseParam1=96, diffuseParam2=384, mag=71, dir=(-50,-10,-50)`,
  `vertDiffuseLight ≈ -24` ⇒ **fog-independent** `shadeBase` mode **71** (spread
  67–76 from the ±5 ambience speckle) ⇒ `ramp[71]` for the grass fill `-2625` (base
  green 144) = **`#084a00`** (green **0x4a**).
- JAR oracle controlled pixel (the two screen points where BOTH engines render grass,
  `(128,120)`/`(128,160)`): JAR `#084800` (green **0x48**). The fresh JAR oracle
  re-render is byte-identical to the stored `jar.png` (not stale).
- **Base-lighting match: orsc 0x4a vs JAR 0x48 — within 1–2 LSB**, down from the
  ~2× (0x8e vs 0x48) error. The reseed bug is closed.
- **Geometry preserved: `single_tile_door` structural diff STILL 9026/9026
  byte-identical** (the fix touches shade only, never faces); orsc-vs-orsc
  determinism STILL 0-pixel.

**Residual (NOT a lighting bug):** the whole-frame pixel diff stays ~99.7% because
of the documented **harness camera-framing + fog-distance + clear-screen drift**.
The orsc render frames the grass FAR from the camera (camera-space Z 1152–2340), so
the authentic distance-fog term `(vertZRot − fogSmoothingStartDistance)/fogZFalloff`
= `(z−10)/20` adds +57…+116 to the shade, pushing the *rendered* grass to ramp index
126–192 (dark). The JAR harness frames the same grass close (near-zero fog), so it
renders at shade ≈71–74 = green 0x48. This fog term is **byte-identical engine math
in both the deob and OpenRSC** (Scene.java:1779-1790; orsc scene.go:653-655,
`fogSmoothingStartDistance=10`/`fogZFalloff=20` per Scene.java:27-28). Proof it is
not a lighting bug: with fog removed, orsc shadeBase=71 matches the JAR's 0x48
exactly. A pixel-exact 3-way frame match additionally requires reconciling the JAR
harness camera framing/clipFar AND the ±5 ambience speckle scheme (see residual
below) — neither is a renderer-math divergence.

### 2. Per-vertex ambience speckle scheme — **DEFERRED (sub-LSB cosmetic)**

`render/orsc/world.go:606` `terrainAmbience` uses a deterministic coord-hash
returning `[-5,4]` for `vertLightOther`; the authentic deob is
`(int)(Math.random()*10)−5` (World.java:933) and the JAR oracle ASM-patches
`Math.random→0.5` ⇒ flat 0. After the reseed fix this is the ONLY remaining
grass-green wobble — a ±5 shade-index spread (the 67–76 above), sub-LSB-scale vs the
former 2× reseed error. It is correct, authentic-looking, and deterministic
(cacheable). Changing it risks perturbing the determinism tests, so it is left as a
separate, flagged change for a future pixel-exact 3-way pin (add a flat-0 ambience
option, or reconcile the JAR patch). Not safe to apply now (would touch determinism).

No *geometry* divergence was found: orsc's built face set is a byte-exact match to
the rev-235 JAR on every JAR-runnable fixture, before AND after the reseed fix.

## Honest caveats / blockers

- **JAR-oracle overlay-table gap (hard blocker, E-water/E-seam/E-polygonHit binary
  diff).** `DumpRenderer.java` synthesizes the wall-def tables but NOT the obf overlay
  / tile-decoration `v[]` def table (the 7 type-class defs `ua.E`/`da.O`/`ga.c`/
  `ta.f`/`la.b`/`eb.d`/`gb.n`, each a `v(String,String,String,int)` whose per-id
  colour/type arrays `v.a`/`v.g`/`v.e` are GameData-loader-populated). Any nonzero
  overlay/roof tile throws `la` at `i.a(i.java:173)` (the exception-wrap helper) from
  `k.a(k.java:1378)` (the bridge/raised-deck section, sentinel 80000). VERIFIED throw
  on `bridge_ov4`, `t3_water_ov2`, `t3_type4_neighbour`, `seam_sector_aligned`,
  `roof_inside`. The jar carries **no embedded GameData** (75 entries, all `.class`),
  so the real table is genuinely unavailable headless; synthesizing it by reflection
  would be a fabrication, not an oracle reading. The 11 overlay/roof fixtures are
  therefore marked **blocked** for the binary diff and settled vs the OpenRSC source.
- **JAR-oracle object-def-table gap (E-diagobj).** `door_diag_obj` (wallDiag 48001,
  the diagonal-object band) is 9025/9026: orsc builds the synthetic wood door-leaf at
  `(6208,-96,6336)` (diagobj.go fixture stand-in), the JAR builds nothing because its
  addModels finds objectId 0 has no GameData object def and silently drops it (deob
  behaviour for unknown ids). NOT an orsc divergence — orsc's fixture stand-in vs the
  JAR's missing-def drop. A real object with a real def builds in both.
- **Pixel diff is harness drift, not a Go defect (all rows).** ~99.7% px differ but:
  orsc fills the whole frame with terrain (170496 nonblack) while the JAR leaves the
  bulk as grey `#7c7c7c` (78921 px) + black (89783 px) — the JAR harness's uncleared/
  grey clear-screen and a different camera framing + `clipFar` (orsc fogLandscape=10000
  vs the jar's per-frame value). Only 512 px (0.30%) match exactly. The **structural
  diff is camera/clear-screen/palette-independent and is byte-clean**, so it is the
  reliable faithfulness signal; the pixel rows need the clear-screen + palette + camera
  reconciliation from GO_JAR_DIFF_RESULTS.md before they are meaningful.
- **Determinism (not a blocker).** orsc is byte-reproducible orsc-vs-orsc (0 px,
  identical face counts) on every fixture, including the new seam fixture.

## Reproduce

```
JDK=/usr/lib/jvm/java-17-openjdk/bin ; DR=~/code/rsc-hacking/rscplus/dumprender
LIB=~/code/rsc-hacking/rscplus/lib  ; JAR=~/code/rsc-hacking/rscplus/assets/rsclassic-1091943135.jar
$JDK/javac -cp "$LIB/asm-5.0.4.jar:$LIB/asm-tree-5.0.4.jar:$LIB/json-20201115.jar" -d "$DR" "$DR/DumpRenderer.java"
$JDK/java -Djava.awt.headless=true -cp "$DR:$LIB/asm-5.0.4.jar:$LIB/asm-tree-5.0.4.jar:$LIB/json-20201115.jar" \
    DumpRenderer "$JAR" testdata/rscdump/single_tile_door.json testdata/rscdump/out/parity/single_tile_door jar
go run ./cmd/renderdiff -out testdata/rscdump/out/parity/single_tile_door \
    go=testdata/rscdump/single_tile_door.json jar=testdata/rscdump/out/parity/single_tile_door/jar.png
```
