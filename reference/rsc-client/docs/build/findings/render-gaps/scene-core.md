> _Verbatim render-completeness audit from the render-deob-completeness-audit workflow (2026-06-01). See ../../RENDER_DEOB_GAPS.md for the consolidated plan._

## Render-Completeness Audit — Scene (obf `lb`) + GameModel (obf `ca`)

### Methodology
Diffed every render-path method in `src/client/scene/Scene.java` and `src/client/scene/GameModel.java` against the clean post-fix decompile (`decompiled/normalized-clean/lb.java` and `ca.java`), with javap/javac cross-checks. Both deob files are unusually complete and heavily annotated; there are **no MISSING-METHODs, no SUMMARIZED/STUB bodies, and no MISSING-CLASSes** (Scanline already reconstructed). The gaps are (1) one genuine **BEHAVIORAL** transcription bug, (2) pervasive **OBF-DRIFT** in cross-class static/writer references that are render-critical, and (3) one cross-file **SIG-DRIFT** on the 6 frustum-AABB globals.

---

### Scene.java (obf `lb`) — 2309 lines vs clean 4805 (delta is stripped opaque-predicate/profiling/try-catch boilerplate, not lost logic)

**BEHAVIORAL — texture U/V transposed in the 128px writer calls (HIGH severity, real render bug)**
- Method `textureRasterScanlines`, lines **2210 (`wb.a`)** and **2212 (`gb.a`)**.
- The per-row coords are correctly defined (`texU = uBase + uRow*sx` = clean `var8*var28+var19`; `texV = uCol + vRow*sx` = clean `var22+var8*var29`; `texW = uStep + wRow*sx`).
- But in the two 128px branches the U and V arguments are swapped vs the clean base:
  - `wb.a` (clean lb.java:2058-2075): clean arg8 = texU(`sx*uRow+uBase`), arg9 = texV(`uCol+sx*vRow`); deob passes arg8=`texV`, arg9=`texU`. Swapped.
  - `gb.a` (clean lb.java:2127-2143): clean arg1 = texV, arg9 = texU; deob passes arg1=`texU`, arg9=`texV`. Swapped.
- The other three writers are **correct**: `cb.a` (2214/2223 vs clean 2435-2452 & 2190-2206), `jb.a` (2219 vs clean 2302-2318), `p.a` (2221 vs clean 2371-2387) all map A/B/C correctly.
- Effect: every 128px-wide, non-translucent textured face has its texture axes transposed (U↔V) — visibly wrong texturing on walls/large textures.
- Fix sketch: at line 2210 swap to `wb.a(vCol,10,0,0,this.raster, texW, s, texU, texV, px, ...)`; at line 2212 swap to `gb.a(texV, vBase,(byte)50, texW, s, sStep<<2, tex, px, texU, ...)`.

**OBF-DRIFT — render-critical cross-class statics & scanline writers left as raw obf names (HIGH; compile blocker, but out-of-file scope per task)**
- 6 frustum-AABB accumulators read/reset/translated as `da.K, m.j, oa.b, aa.f, nb.y, aa.b` (lines 538-543, 1587-1603) — undefined obf classes here.
- Sin/cos + scratch globals: `e.nb` (291-293), `ba.cc` (297-298 and 454-532), `k.e` (632-686), `k.o`/`da.bb`/`m.d`/`u.d` (2278-2286), `db.i` (286-287), `ib.a` (764-795).
- Scanline writers `wb.a/gb.a/jb.a/p.a/cb.a` (texture) and `ua.a/t.a/ia.a` (gradient) — raw obf statics on unrelated classes (2210-2266). Call-site arg expressions verified faithful; only the owners are un-deobbed.
- javac of Scene.java alone = 48 `cannot find symbol` errors, all on these obf owners (`ba, ib, aa, m, k, e, oa, nb, db, da, wb`, …). Fix sketch: route these to the real deob owner classes/fields once those classes are named (Surface, GameShell, SurfaceSprite, World, etc.) — same work the Mudclient pass is doing.

**Verified FAITHFUL (no gap):** constructor; addModel/removeModel/clearModels; sprite table; setMouseLoc; `setCameraOrientation` (eye-vector yaw→pitch→roll, inverse-angle storage) vs clean 582-628; `setBounds` vs clean; `expandFrustum` (rotation order + all 6 min/max polarities decoded from `~`-idiom) vs clean 1247-1309; `setLightFull`/`setLight`; `fillColour`/texture cache (`allocateTextures`/`defineTexture`/`prepareTexture` LRU/`buildTexturePixels` shade copies/`scrollTexture`); `initialisePolygon3d` (method293 normals + cull sign + screen AABB) and `initialisePolygon2d`; `polygonsQSort`; `polygonsIntersectSort`; `reorderRange`; `separatedOrInOrder` + `faceOrders` (both plane-band halves); `polygonsOverlap` (the irreducible 3-phase intersect, every interpolate/spanOrder/edgeOrder arg verified); `spanOrder`/`edgeOrder`/`interpolate` vs clean 3920/652/155; `render` (frustum build, project, visible-poly build incl. TRANSPARENT==12345678 skip at 1650, depth+overlap sort, draw loop); `drawSolidFace` + `clipNearCrossing` (near-plane clip, Gouraud vs flat shade, fog); `drawSpriteFace`; `generateScanlines`/`triEdges`/`quadEdges`/`walkScanEdge` (triangle path byte-verified vs clean 2722-2939); `rasterize` gradient-ramp build; `gradientRasterScanlines` (all 3 writer branches + transparent/wideBand gating verified vs clean 2542-2669); `flushToImage`.

---

### GameModel.java (obf `ca`) — 1419 lines vs clean 2410

**SIG-DRIFT (cross-file) — frustum statics named inconsistently with Scene (HIGH; correctness-of-linkage)**
- `project`, lines **1220-1222**: the 6 frustum globals are read as `ClientStream.frustumNearZ, SocketFactory.frustumFarZ, NameHash.frustumMinX, BZip.frustumMaxX, DataStore.frustumMinY, BZip.frustumMaxY`.
- These are the **same 6 globals** Scene writes as `da.K, m.j, oa.b, aa.f, nb.y, aa.b`. Verified mapping (from clean accept-test lb.java:888 ⇄ ca.java:888): `da.K`=NearZ↔boundZ1, `m.j`=FarZ↔boundZ2, `oa.b`=MinX↔boundX1, `aa.f`=MaxX↔boundX2, `nb.y`=MinY↔boundY1, `aa.b`=MaxY↔boundY2 (note `aa` carries both MaxX=aa.f and MaxY=aa.b, so `BZip`=class `aa` — internally consistent there).
- The two files must agree on these owner names or the frustum cull silently reads uninitialised globals. The reject-condition polarity itself is **correct** (all 6 terms verified as exact inversions of the clean accept test).
- Fix sketch: pick one owner naming for the 6 globals and use it in both files (Scene currently uses raw `da/m/oa/aa/nb`; GameModel uses readable-but-wrong-class `ClientStream/...`).

**OBF-DRIFT (minor, render path):** `applyRotation` uses `pa.a`/`ImageLoader.sine9` mix (line 910 raw `pa.a`, 918-934 `ImageLoader.sine9`); `project` uses `ImageLoader.sine11` (deob) for what clean calls `pa.j`. Consistent within the method but the `pa.a` at 910 is an un-renamed straggler. Fix: rename `pa.a`→`ImageLoader.sine9`.

**Verified FAITHFUL (no gap):** all ctors incl. `GameModel(byte[],…)` and `GameModel(String)` base-64 decoders; `allocate`; `clear`/`reduce`/`vertexAt`/`createVertex`/`createFace`; `split`/`merge`/`copyLighting`/`copy`; `setLight`(×2)/`setLightDirection`/`setVertexAmbience`; `rotate`/`orient`/`translate`/`place`/`copyPosition`/`determineTransformKind`; `translateVertices` (the deliberate baseZ→Y / baseY→Z axis swap is faithful — verified vs clean call site ca.java:1608 `d(0,xb,Sb,r)`); `applyRotation`/`applyScale`/`applyShear`; `computeBounds`; `light` (divisor + face & per-vertex intensity formulas vs clean 1280-1403); `relight` (cross-product normals, the `Z*65535` vs `X/Y*65536` asymmetry is faithful vs clean 1639); `apply` (state-1/2 paths); `project` (yaw/roll/pitch rotation math + perspective divide, byte-verified vs clean 882-983); `commit`; `textureId`/`readBase64`/string codec.

---

## GAPS

**Scene.java (obf lb)**
- BEHAVIORAL: **1** (U/V transposed in `wb.a`@2210 and `gb.a`@2212 — count as one bug class, two sites). HIGHEST SEVERITY.
- OBF-DRIFT: render-critical raw-obf references at **~20 distinct owners** (`e,ba,k,db,u,m,da,oa,aa,nb,ib` + writers `wb,gb,jb,p,cb,ua,t,ia`); **48** javac `cannot find symbol` errors in this file alone.
- MISSING-METHOD / MISSING-CLASS / SUMMARIZED-STUB: **0**.

**GameModel.java (obf ca)**
- SIG-DRIFT (cross-file): **1** (the 6 frustum globals named `ClientStream/SocketFactory/NameHash/BZip/DataStore.*` here vs `da/m/oa/aa/nb` in Scene — must unify). HIGH severity for linkage.
- OBF-DRIFT: **1** straggler (`pa.a`@910 not renamed to `ImageLoader.sine9`). LOW.
- BEHAVIORAL / MISSING-METHOD / MISSING-CLASS / SUMMARIZED-STUB: **0**.

**Highest-severity items (ranked):**
1. **Scene.textureRasterScanlines U/V swap** (lines 2210, 2212) — only logic-level rendering correctness bug found; corrupts all 128px non-translucent textures. Fix = swap the texU/texV args in the wb.a and gb.a calls.
2. **Frustum-static cross-file naming** (Scene `da.K/m.j/oa.b/aa.f/nb.y/aa.b` ⇄ GameModel `ClientStream/SocketFactory/NameHash/BZip/DataStore.frustum*`) — unify owners or the cull reads the wrong globals.
3. **Render-critical OBF-DRIFT in Scene** (scanline writers + scattered statics) — non-logic but blocks compile/link; resolve by pointing at the real deob owner classes (the in-flight Mudclient renaming covers most).

Files: `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/scene/Scene.java`, `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/scene/GameModel.java`. Clean oracles: `/home/free/code/rsc-hacking/westworld/reference/rsc-client/decompiled/normalized-clean/lb.java`, `.../ca.java`.
