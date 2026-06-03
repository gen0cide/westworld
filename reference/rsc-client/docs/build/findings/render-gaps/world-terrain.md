> _Verbatim render-completeness audit from the render-deob-completeness-audit workflow (2026-06-01). See ../../RENDER_DEOB_GAPS.md for the consolidated plan._

## RENDER-COMPLETENESS AUDIT — World (k), WorldEntity (w), GameCharacter (ta)

**Scope/method:** Diffed every method in the deob sources at `.../src/client/world/{World,WorldEntity,GameCharacter}.java` against the jar's clean post-fix decompile at `.../decompiled/normalized-clean/{k,w,ta}.java`, with the `loadMapData` "Couldn't-be-decompiled" method verified from raw bytecode and cross-checked against the rev-204 oracle `~/code/rsc-hacking/mudclient204/src/World.java`. Obf↔readable via NAMING.md (World/Scene swap applied: **k=World, lb=Scene, w=WorldEntity, ta=GameCharacter**).

**Headline:** World.java is a near-complete, high-fidelity transcription. Every clean `k.java` method is present (no MISSING-METHOD, no MISSING-CLASS, no STUB/elision) and the load-bearing render logic (heightmap, ramp, overlay/250-remap, walls/method422, roof cascade, scenery/addModels, loadMapData) is faithful. WorldEntity and GameCharacter are passive data holders with no self-render methods; both are complete. The defects are a small number of precise BEHAVIORAL/SIG-DRIFT items, the most important being a color-channel argument-order landmine and a wrong-class call on the .jm fallback.

---

### World.java (obf k) — method-by-method result

Verified FAITHFUL (clean evidence in parens):
- `getTerrainHeight` ← clean `g` (k.java:2068) — index `L[q][48*x+y]*3`, quadrant map ✓
- `getTerrainColour` ← `a(byte,…)` (k.java:2606) — `eb[q][y+48*x]` ✓
- `getWallNorthsouth`/`getWallEastwest`/`getWallDiagonal`/`getWallRoof`/`getTileDirection`/`getTileDecoration`/`getTileDecorationOr`/`getTileType` ← clean `e`(4172)/`a`(4026)/`c`(2678)/`d`(1788)/`b`(3898)/`b`(3987)/`d`(2664)/`d`(1858) — all index expressions, quadrant swaps and the getWallRoof `y*48+x` transposition match ✓
- `getElevation` ← clean `f` (k.java:2516) — both triangle branches, `aX/aY` flip, `/anInt585` ✓
- `route` ← clean `a(int[],…)` (k.java:2209) — BFS bitmasks (0x78/0x72/0x74/0x71, diagonals 0x7C/0x76/0x79/0x73), via-codes, backtrace ✓
- `setWallObjectAdjacency`/`clearWallObjectAdjacency`/`removeObject`/`removeObject2` ← clean `a`(2110)/`a`(602? no — 4100)/`a`(206…)/`a`(380) — directional bit sets/clears + neighbour propagation ✓
- `setTerrainAmbience`/`method425`/`method404` ← clean `a`(?)/`b`(1830)/`c`(2548) — cell split, 0x63/0x59/0x56/0x6C masks ✓
- `reset`/`setTiles`/`setTileDecoration` ← clean `b`(535)/`a`/`e` — 250→9/9/2 remap matches oracle setTiles ✓
- `unpackData` ← clean `a(int,boolean,byte[])` (k.java:508) — 24-bit header, verbatim/bzip2 ✓
- `method428` ← clean `a(…,byte,…)` (k.java:2191) — `+80000` sentinel raise of `ab[ax][ay]`/`ab[bx][by]` ✓
- `method422` ← clean `a(int,ca,…)` (k.java:4065) — 4-vertex wall quad, front/back colour, `lb.Tb==5⇒E=30000+id` ✓ (benign deviation, see B3)
- `method402` ← clean `a(int,byte,…)` (k.java:3941) — 0x7F7F7F/0xFEFEFF dimming, param swap documented ✓
- `addModels` ← clean `a(ca[],byte)` (k.java:1898) — diag 48001..59999 gate, w/h-by-dir, translate/orient/register, footprint-clear ✓
- `buildSection` ← clean `a(int,int,boolean,int,int)` (k.java:602–1786) — terrain vertices (type-4 flatten), gouraud floor tri/quad selection (decoType 4/5/2 cascade), 8×8 split, 4 wall orientations, roof-top levelling (label909), `method428` pass, `buildRoofs`, `+80000` strip ✓
- `buildRoofs` ← clean roof loop (k.java:1448–1716, label853) — 4 corner pts, ridge `+80000`, 16-unit `isRoofCorner` taper (magic 26431), 7-way emission cascade with all vertex tags ✓
- `loadMapData` ← clean bytecode-only `b` (k.java:2717–3897) — .hei RLE+prefix-sum (seed 64/35), .dat walls/diag(+12000)/roof/deco/dir RLE, .loc(+48000), .jm delta + 16-bit diag, IOException seed (plane0=-6/plane3=8). Verified against oracle World.java:252–438 ✓

Defects found in World.java:

- **[BEHAVIORAL — HIGH] Ground colour ramp argument order is the obf packer order, not (R,G,B); doc comment is wrong.** `World.java:174–180` calls `ClientStream.rgb(...)` with args copied verbatim from clean `da.a(var0,(byte)-66,var2,var3)` (k.java:4244/4261/4278/4294, magic byte dropped). Clean `da.a` = `-1 - var0/8 - (var2/8)*1024 - (var3/8)*32` (da.java:288) and oracle `Scene.rgb(i,j,k)=-1-(i/8)*1024-(j/8)*32-k/8` (Scene.java:1016) ⇒ **`da.a(var0,_,var2,var3) == Scene.rgb(R=var2, G=var3, B=var0)`**, i.e. the surviving 3 args are **(B, R, G)**. The values only render correctly if `ClientStream.rgb` keeps the `(B,R,G)` order. The ctor Javadoc (`World.java:142–146`) asserts "`da.a(R,(byte)-66,B,G) == Scene.rgb(R,G,B)`" — **incorrect**; under a true `(R,G,B)` rgb the four bands have R/G/B swapped (e.g. band 1 `rgb(0,3i,144)` would give R=0,G=3i,B=144 vs oracle R=3i,G=144,B=0). Fix sketch: either implement `ClientStream.rgb(b,r,g)` with that exact body, or reorder all four calls to `(255-4i, 255-1.75i, 255-4i)` etc. and fix the comment.
- **[MISSING-METHOD / SIG-DRIFT — HIGH] `ClientStream.rgb` is not defined.** Deob `ClientStream.java` declares no `rgb` (grep: only profiling fields). All four ramp calls + the `Scene.Tb`/`Mudclient.Jk` table imports reference members that may not exist yet. Fix: add `static int rgb(int,int,int)` to ClientStream with the `da.a` body (param order per B1).
- **[BEHAVIORAL / wrong-class — HIGH] `.jm` fallback calls the wrong class.** `World.java:1382` calls `EntityDef.a(STR_MAPS_DIR + name + STR_JM, -19675, jm, 20736)`. Clean bytecode invokes `ta.a (String,I,[B,I)V` = **GameCharacter** (k.java loadMapData bytecode offset 643). Neither deob `EntityDef` nor clean `t.java` has an `(String,int,byte[],int)` overload — the readFully reader is `GameCharacter.readFromStore` (clean `ta.a`, ta.java:46). The World import comment (`World.java:22`) miscredits it to `EntityDef // obf t`. Fix: `GameCharacter.readFromStore(STR_MAPS_DIR + name + STR_JM, -19675, jm, 20736)`.
- **[OBF field-label — MED, not a render bug] `mapPack`/`memberMapPack` (obf m/gb) identities are inverted vs the rev-235 bytecode.** Clean `.dat` loads `gb` first then `m` (offsets 2e5/31a) and `.loc` from `gb` (offset 593). The oracle (World.java:314–316,370) loads `.dat` from mapPack-then-memberMapPack and `.loc` from mapPack. The deob (World.java:1338–1340,1368) uses readable `mapPack`-first / `.loc`-from-`mapPack` — **matching the oracle**, so rendered output is correct. But this means obf `gb`=mapPack(free) and obf `m`=memberMapPack, the reverse of the field-decl comments (`World.java:107–108`). Risk is only if Mudclient fills the readable fields by obf order. Fix: correct the `m`/`gb` field-identity comments (no logic change).
- **[BEHAVIORAL — INFO/benign] `method422` 4th-vertex tag passes `magic` not `magic ^ -14505`.** `World.java:751` passes `magic` to `model.e(...,magic)`; clean passes `var6 ^ -14505` (k.java:4084) = 95. Confirmed harmless: that 4th param is only used in a discarded anti-tamper expr in `GameModel.e` (ca.java:1158, `100/((-46-var4)/58)`, no div-by-zero for either value). No render effect; note for fidelity only.
- **[INFO] `.jm` 16-bit diagonal masking** (`World.java:1390`) uses `& 0xFF` on both bytes, matching the oracle (World.java:155); the raw bytecode `baload`s are sign-extended but the oracle confirms masking is intended. No action.

---

### WorldEntity.java (obf w) — result

Clean `w.java` is a passive struct (17 instance fields + 4 static + ctor) with 3 incidental static utils and the z()/z() decoders. **All present and faithful** in the deob:
- Fields l/r/k/s/e/m/j/h/q/u/o/i/t/b/c/f/p → semantic names (normalY/normalX/normalZ/normalDot/minX/maxX/minZ/maxZ/minDepth/maxDepth/model/faceIndex/sortDepth/objectId/active/slotIndex/prevSortIndex); statics n/d/a/g ✓. Ctor sets p=-1, f=0 ✓.
- `trimAndValidateString` ← `a(CharSequence,byte)` (w.java:26) ✓; `computeCrc32` ← `a(int,int,byte[],int)` (w.java:127) — init -1, `wb.q[(b^crc)&0xFF]^(crc>>>8)`, `~crc` ✓; `readSignedShort` ← `a(byte[],int,int)` (w.java:159) — `256*ub(hi)+ub(lo)`, `>=32768⇒-65536` ✓.
- **No normals/bounds/sort-key computation exists in `w`** — those fields are WRITTEN by Scene (lb); so there is **no missing render method here**. (The prompt's "WorldEntity normals/bounds/sort-key" logic lives in scene/Scene.java, outside this file's scope.)

Defect:
- **[INFO — error-path only] z[] string-pool inputs were shortened.** Deob `ERROR_SIGS` (WorldEntity.java:425–436) feeds `decodeStep1("S8B")` etc., dropping the leading obfuscation char present in clean `z(z("\u0011S8B"))` (w.java:23) and truncating `"\b\b\u0016\u0006"`→`"\b\b"`. Affects only exception-message text, never render output.

---

### GameCharacter.java (obf ta) — render-relevant result

Clean `ta.java` is a passive entity struct (equip `m[12]`, `F[10]`, waypoints `k[10]`, name/message strings, anim/combat ints) with **no self-render methods**; rendering is done by Mudclient/Scene reading these fields. Deob is complete:
- `readFromStore` ← clean `a(String,int,byte[],int)` (ta.java:46) — DataStore.openStream + readFully + EOF-swallow ✓ (this is the real `.jm` reader miscalled by World — see World defect B3).
- `popcount` ← clean `a(int,byte)` (ta.java:67) — parallel bit-count ✓; ctor field inits ✓; z()/z() decoders ✓.
- No MISSING-METHOD / STUB / render gap.

---

## GAPS

Per-class counts by type:

| Class | MISSING-METHOD | MISSING-CLASS | STUB | OBF-DRIFT | SIG-DRIFT | BEHAVIORAL |
|---|---|---|---|---|---|---|
| World (k) | 1 (`ClientStream.rgb` undefined) | 0 | 0 | 1 (m/gb field-label, non-render) | 1 (.jm→EntityDef vs GameCharacter) | 2 (ramp arg-order; +benign method422 tag) |
| WorldEntity (w) | 0 | 0 | 0 | 0 | 0 | 0 (1 INFO: z[] pool, error-path) |
| GameCharacter (ta) | 0 | 0 | 0 | 0 | 0 | 0 |

Highest-severity items (fix order):
1. **HIGH — World ramp channel order (`World.java:174–180`).** Args are obf `(B,R,G)` order; the ctor doc claims `(R,G,B)`. Wrong unless `ClientStream.rgb` is the `(B,R,G)` `da.a` packer. Mis-implementing rgb swaps Green/Blue on all terrain → globally wrong ground colour. Evidence: da.java:288 vs Scene.java:1016.
2. **HIGH — `ClientStream.rgb` does not exist (`World.java:174`).** Must be added as `static int rgb(int b,int r,int g){ return -1 - b/8 - (r/8)*1024 - (g/8)*32; }`.
3. **HIGH — `.jm` reader wrong class (`World.java:1382`).** `EntityDef.a(...)` has no such overload; clean calls `ta.a` = `GameCharacter.readFromStore`. Breaks raw-`.jm` map loading (and won't compile against EntityDef). 
4. **MED — m/gb field-identity comments inverted (`World.java:107–108`).** Logic matches the oracle (correct render); only the obf-mapping comments are backwards — fix to avoid Mudclient mis-wiring the pack byte arrays.

No MISSING-CLASS in this area (Scanline was outside it). No summarized/elided method bodies — every clean `k.java`/`w.java`/`ta.java` method is fully transcribed; `loadMapData` and `buildSection`/`buildRoofs` (the large, error-prone, bytecode-only/obfuscated methods) were verified line-for-line against bytecode and the rev-204 oracle and are faithful.
