# GO↔JAR render diff — first ground-truth comparison

Built by the `go-jar-render-diff` workflow (2026-06-01). The JAR is the **true render oracle**, and
this milestone makes it usable *without* the deob compiling. Harness:
`rscplus/dumprender/{PatchAndRender,DumpRenderer}.java`.

## The JAR oracle engine works — headless

We can drive the obfuscated `rsclassic` jar to render one controlled frame **headless** — no display
/ Xvfb, no network, no applet bootstrap, no GameData archive:

1. **Classloader + 3 ASM patches** (via `rscplus/lib/asm-5.0.4`): `k.b` (`loadMapData`) → `RETURN` (so
   injected grids survive `buildSection`); `ua.b` (minimap) → `RETURN`; the single `Math.random`
   invokestatic in `k` → `ldc 0.5` (⇒ per-vertex ambience 0 = deterministic seed-0 speckle).
2. Construct obf `ua(512,334,1,null)` (Surface — **`Component=null` skips all AWT**, allocates only the
   `int[] rb` pixel buffer) → `lb(ua,15000,15000,1000)` (Scene) → `k(lb,ua)` (World).
3. Synthesize the **5 wall-def static tables** the fixture's wall id needs (`v.a`/`client.Jk` front/back
   colour, `ib.d` height, `lb.Tb` visible-flag, `u.a` adjacency) — mirroring the Go `syntheticFacts`.
4. Inject the 7 terrain grids into World's obf `byte[][]` fields (verified obf↔deob: `L`=height,
   `eb`=colour, `f`=N/S wall, `P`=E/W wall, `s`=diagonal, `A`=roof, `R`=overlay, `mb`=direction).
5. `buildSection` = `k.a(0,122,true,0,0)` → **192 models**; `setBounds` + `setCameraOrientation`
   (mode arg **must be −12349** to derive the look vector); `render` = `lb.c(0)` → **81 225 lit pixels**.
6. Read `ua.rb` → PNG; walk `lb.Z[0..ab]` → `rscdump-faces/1` JSON.

This recipe (constructing the obf render pipeline directly by reflection) is reusable and is the basis
for the JAR side of the 3-way diff.

## First GO↔JAR result (single-tile-door fixture)

**Structural (geometry) diff — CLEAN MATCH.**
- `only-in-JAR = 0` — every face the oracle builds exists in the GO scene (GO is a strict superset).
- **Door quad SHARED, byte-identical**: centroid `(10304,−96,10368)`, 4 verts, fill `−15719`.
- `only-in-GO = 16256` — all terrain quads beyond the obf grid's smaller window extent (a harness
  window-size artifact), **0 wall/door/roof divergences**.
- → On the overlapping region, the Go render lib builds the **same geometry, vertex counts, fills, and
  transformed centroids** as the real client. The door-perception bug class is *not* present for the
  straight-wall door (the diagonal-door case is a separate, now-root-caused gap — see
  `RENDER_FIDELITY_FINDINGS.md`).

**Pixel diff — 98.65% differ, but it's harness drift, not a Go defect.**
- The divergence is **camera-framing + clear-screen fill + terrain palette** (the design's "rule-4"
  reconciliation items): the JAR harness leaves part of the frame uncleared and frames the camera
  differently. Where terrain *is* drawn, the green matches GO to **1 LSB** (`#084800` vs `#084900`) —
  the rasterizer colour math agrees; only *which* pixels get terrain vs clear-screen differs.

## What's left for a pixel-exact 3-way diff
- Reconcile the JAR harness camera derivation + `clipFar` (GO 7000 vs the jar's per-frame value) and the
  clear-screen fill + terrain palette for the synthetic grass id.
- Then the real payoff: once the **deob compiles**, stand up the DEOB render engine into the same
  harness for the full **GO ↔ DEOB ↔ JAR** comparison on real scenes — the definitive faithfulness check.

Artifacts: `testdata/rscdump/out/jar-door.{png,faces.json}` and `testdata/rscdump/out/gojar/`.
