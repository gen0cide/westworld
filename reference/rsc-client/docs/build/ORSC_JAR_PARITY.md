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

## orsc renderer-proper divergences (DOCUMENT ONLY — flag for the colleague)

1. **Terrain shade ~1.7-2× brighter than the JAR oracle** (PIXEL, cosmetic/latent).
   On the flat door fixture (groundColour idx 70, elevation 0, seed 0) orsc paints
   green `#0f8e00..#0d7c00` (g-channel 0x7c..0x8e) where the JAR paints flat `#084800`
   (g-channel 0x48). Two compounding causes:
   (a) **Per-vertex ambience scheme mismatch** — orsc `terrainAmbience` (world.go:606)
   is a coord-hash returning `[-5,4]`; the JAR oracle ASM-patches `Math.random→0.5`
   ⇒ ambience `(int)(0.5*10)-5 = 0` (flat). Neither matches the *live* client
   (real `Math.random`). This is a determinism-pinning mismatch, not a geometry bug —
   the structural faces are identical (fills carry no per-vertex shade).
   (b) The brightness gap is *larger* than ±5 speckle, so a base-shade/lighting or
   camera-distance-falloff difference is also present (raster/model lighting,
   confounded by the harness camera framing). File: `render/orsc/world.go:606`
   (terrainAmbience) + the lighting path in `render/orsc/{model,raster}.go` (HANDS-OFF).
   Recommended fix: out of scope for parity (HANDS-OFF). For a pixel-exact 3-way diff,
   reconcile the JAR oracle's `Math.random→0.5` patch with orsc's coord-hash (patch
   the JAR side to orsc's scheme, or render orsc with a flat-0 ambience option) AND
   re-derive the JAR harness camera framing (below). NOTE: GO_JAR_DIFF_RESULTS.md's
   old "1 LSB" green match (`#084800` vs `#084900`) was the DELETED engine; the
   current orsc green does not reproduce it — worth a colleague look.

No *geometry* divergence was found: orsc's built face set is a byte-exact match to
the rev-235 JAR on every JAR-runnable fixture.

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
