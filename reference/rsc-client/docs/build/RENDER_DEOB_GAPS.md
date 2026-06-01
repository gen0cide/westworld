# Deob render logic — completeness gaps & closing plan

Consolidated from four area audits (scene-core, surface-sprites, world-terrain, mudclient-render),
each a method-by-method diff of the deob render path against the jar's clean post-fix decompile in
`decompiled/normalized-clean/`. This file is a read-only inventory + plan; it edits no source.

## 1. Status (honest)

The render logic is **structurally complete but not yet a trustworthy oracle, and not silently
correct.** Across all four areas there are **0 MISSING-CLASS**, **1 MISSING-METHOD**
(`ClientStream.rgb` is referenced but undeclared), and **1 SUMMARIZED/STUB** (a doc-level note: the
static FontWidths half of merged obf class `n` lives in `Surface.GameFonts`, not in `Scanline.java`) —
so the big "not-yet-deobfuscated holes" are essentially closed: every clean method body has been
transcribed, including the bytecode-only `loadMapData`, `buildSection`/`buildRoofs`, and the
"Couldn't-be-decompiled" `flushEventQueue`. The remaining defects are **10 BEHAVIORAL divergences**
(silent render bugs: a texture U/V transpose, a BMP-width misroute, a wrong-class `.jm` reader, a
channel-order ramp landmine, plus five Mudclient mis-routes/dropped resets and a `~qc` guard error)
and **pervasive OBF-DRIFT** (raw obf field/method/owner names — ~48 `cannot find symbol` in Scene
alone, ~all 14 Mudclient render methods un-renamed) plus **3 SIG-DRIFT** mislabels (the systemic
World↔Scene alias inversion being the worst, because it makes the deob actively misleading as a
fidelity oracle). Bottom line: the deob is faithful enough to read alongside the clean decompile, but
shipping it as the render engine requires closing the 10 behavioral bugs and resolving the obf-drift
that overlaps the M2 347-pair compile gate.

## 2. Consolidated gap table

| Area | Class (obf) | Method | Gap type | Severity | Clean ref | One-line fix |
|---|---|---|---|---|---|---|
| scene-core | Scene (lb) | `textureRasterScanlines` | BEHAVIORAL | render-breaking | lb.java:2058-2075 / 2127-2143 | Swap texU/texV args in the `wb.a`@2210 and `gb.a`@2212 128px calls (U↔V transposed). |
| scene-core | Scene (lb) | render path (statics + writers) | OBF-DRIFT | fidelity (compile blocker) | — | Route ~20 raw obf owners (`e,ba,k,db,u,m,da,oa,aa,nb,ib` + writers `wb,gb,jb,p,cb,ua,t,ia`) to real deob classes; clears 48 `cannot find symbol`. |
| scene-core | Scene (lb) | frustum-AABB globals | SIG-DRIFT | fidelity | lb.java:888 | Unify the 6 globals' owner naming with GameModel (Scene uses raw `da.K/m.j/oa.b/aa.f/nb.y/aa.b`). |
| scene-core | GameModel (ca) | `project` | SIG-DRIFT | render-breaking-if-mismatched | ca.java:888 | Name the 6 frustum globals to match Scene (deob uses `ClientStream/SocketFactory/NameHash/BZip/DataStore.frustum*` vs Scene `da/m/oa/aa/nb`); must agree or cull reads uninit globals. |
| scene-core | GameModel (ca) | `applyRotation` | OBF-DRIFT | cosmetic | ca.java:910 | Rename straggler `pa.a`@910 → `ImageLoader.sine9`. |
| surface-sprites | ImageLoader (pa) | `loadBmpImage` | BEHAVIORAL | render-breaking | pa.java:21 (write); fb.java:35 / lb.java:3770 (read) | Write/read `World.surfaceWidth`, not fabricated `imageWidthCarrier.imageWidth` (LoaderThread); delete the carrier. |
| surface-sprites | Surface (ua) | `drawstring`/`GameFonts` | SIG-DRIFT | fidelity | ua.java:2160/2164 | Point `GameFonts.antialiased[]` at `SurfaceImageProducer.k[font]` and `GameFonts.data` at real `m.b`; local stand-in is stuck false → wrong drop-shadow/AA. |
| surface-sprites | Scanline (n) | (class) | SUMMARIZED/STUB | cosmetic | n.java:4,6,12,14 + `<clinit>` | Add doc note: obf `n` was Scanline+FontWidths merged; static half is in `Surface.GameFonts` (`n.a` byte-exact). No logic change. |
| world-terrain | World (k) | ctor (ground ramp) | BEHAVIORAL | render-breaking | da.java:288 / Scene.java:1016 | Ramp args are obf `(B,R,G)` order, not `(R,G,B)`; implement `ClientStream.rgb` as the `(B,R,G)` packer or reorder all four calls + fix the ctor Javadoc. |
| world-terrain | World (k) | ctor (ground ramp) | MISSING-METHOD | render-breaking | da.java:288 | Add `static int rgb(int b,int r,int g){ return -1 - b/8 - (r/8)*1024 - (g/8)*32; }` to `ClientStream` (undeclared). |
| world-terrain | World (k) | `loadMapData` (.jm) | BEHAVIORAL / wrong-class | render-breaking | k.java loadMapData bytecode +643; ta.java:46 | `World.java:1382` calls `EntityDef.a(...)` (no such overload) — should be `GameCharacter.readFromStore(...)` (clean `ta.a`). |
| world-terrain | World (k) | `mapPack`/`memberMapPack` fields | OBF-DRIFT (label) | cosmetic | World.java:107-108 | Field-identity comments inverted vs bytecode (`gb`=mapPack, `m`=memberMapPack); logic matches oracle, fix comments to avoid mis-wiring. |
| world-terrain | World (k) | `method422` | BEHAVIORAL (benign) | cosmetic | k.java:4084 | 4th-vertex tag passes `magic` not `magic ^ -14505`; harmless (used only in discarded anti-tamper expr). Note for fidelity. |
| world-terrain | WorldEntity (w) | `ERROR_SIGS` | BEHAVIORAL (info) | cosmetic | w.java:23 | z[] string-pool inputs shortened; affects only exception text, never render. |
| mudclient-render | Mudclient (client) | all render methods | SIG-DRIFT | fidelity (oracle-poisoning) | client.java:221/143; L7294-7332; L6930/L7090 | Swap `scene`↔`world` aliases: `scene` should be `Ek`(lb=Scene, owns camera+render), `world` should be `Hh`(k=World, owns terrain/elevation). |
| mudclient-render | Mudclient (client) | `drawActiveInterface` | BEHAVIORAL | render-breaking | client.java:1410 | `h((byte)127)` is `drawSocialDialog`; deob calls `drawDuelConfirm((byte)127)` (int overload). Call `drawSocialDialog((byte)127)`. |
| mudclient-render | Mudclient (client) | `drawActiveInterface` | BEHAVIORAL (structural) | render-breaking | client.java:1364-1434 | Clean fall-through chain flattened to exclusive if/else-if; self-guarding tail-default panels never reached. Model as first-match chain that still fires social-dialog default. |
| mudclient-render | Mudclient (client) | `drawChat` | BEHAVIORAL | render-breaking | client.java:5496-5498 | Restore `if (param != 2) this.ak = null;` before ground-item loop; else stale ground-item icons persist. |
| mudclient-render | Mudclient (client) | `drawUiTabMinimap` | BEHAVIORAL | render-breaking | client.java:4462-4464 | Restore `if (byteParam <= 119) this.bf = null;` (byte param was discarded); else stale health bars persist. |
| mudclient-render | Mudclient (client) | `drawInventoryTab` | BEHAVIORAL | render-breaking | client.java:5314/5320/5328/5336 | Four `~qc==-1` guards mis-coded as `qc==1`; change to `qc==0` (sub-tabs selectable only from wrong prior state). |
| mudclient-render | Mudclient (client) | all 14 render methods | OBF-DRIFT | fidelity (compile blocker) | — | ~50 obf fields in `drawGameFrame`, obf method calls (`world.a/c`,`scene.f`,`surface.a/b/e`), 163 `lb/ca/ta/ba/ua`-typed refs; `drawWorld` not even alias-renamed (raw `Ek/Hh/yd/zh`). |

## 3. Counts

### Per gap type (totals)

| Gap type | Count | Notes |
|---|---|---|
| MISSING-CLASS | **0** | (Scanline already reconstructed; no other missing render class) |
| MISSING-METHOD | **1** | `ClientStream.rgb` referenced but undeclared (world-terrain) |
| SUMMARIZED/STUB | **1** | doc-level: Scanline omits FontWidths half of merged `n` (cosmetic) |
| OBF-DRIFT | **5** | Scene statics+writers, GameModel `pa.a` straggler, World field-label, Mudclient all-methods, (Surface counted under SIG-DRIFT) |
| SIG-DRIFT | **3** | Scene frustum globals, GameModel frustum globals, Surface GameFonts stand-ins, Mudclient World↔Scene inversion → counted as the systemic Mudclient item + Scene/GameModel/Surface = 4 mislabels; the load-bearing one is Mudclient |
| BEHAVIORAL | **10** | Scene U/V (1), ImageLoader BMP width (1), World ramp order (1), World .jm wrong-class (1), World method422 tag (1, benign), WorldEntity ERROR_SIGS (1, info), Mudclient ×5 (drawActiveInterface routing + structural, drawChat, drawUiTabMinimap, drawInventoryTab) |
| **TOTAL gaps** | **~21 rows** | (writer U/V counted as one bug class / two sites) |

Render-breaking subset (must fix before shipping the render engine): **9** —
Scene U/V transpose; ImageLoader BMP width; World ramp channel order; World `.jm` wrong-class;
`ClientStream.rgb` missing; Mudclient drawActiveInterface (routing) + drawActiveInterface (structural)
+ drawChat reset + drawUiTabMinimap reset + drawInventoryTab `~qc` guards. (GameModel frustum-global
naming is render-breaking only if the two files disagree at link time — treat as a hard prerequisite.)

### Per area / per class

| Area | Class | MISS-CLASS | MISS-METHOD | SUMM/STUB | OBF-DRIFT | SIG-DRIFT | BEHAVIORAL |
|---|---|---|---|---|---|---|---|
| scene-core | Scene (lb) | 0 | 0 | 0 | 1 (~20 owners) | 1 | 1 (U/V) |
| scene-core | GameModel (ca) | 0 | 0 | 0 | 1 (`pa.a`) | 1 | 0 |
| surface-sprites | Surface (ua) | 0 | 0 | 0 | 0 | 1 (GameFonts) | 0 |
| surface-sprites | SurfaceSprite (ba) | 0 | 0 | 0 | 0 | 0 | 0 |
| surface-sprites | SpriteDecoder (ea) | 0 | 0 | 0 | 0 | 0 | 0 |
| surface-sprites | SpriteScaler (ia) | 0 | 0 | 0 | 0 | 0 | 0 |
| surface-sprites | ImageLoader (pa) | 0 | 0 | 0 | 0 | 0 | 1 (BMP width) |
| surface-sprites | SurfaceImageProducer (fb) | 0 | 0 | 0 | 0 | 0 | 0 |
| surface-sprites | Scanline (n) | 0 | 0 | 1 | 0 | 0 | 0 |
| world-terrain | World (k) | 0 | 1 (`rgb`) | 0 | 1 (m/gb label) | 1 (.jm class) | 2 (ramp + benign tag) |
| world-terrain | WorldEntity (w) | 0 | 0 | 0 | 0 | 0 | 0 (1 info) |
| world-terrain | GameCharacter (ta) | 0 | 0 | 0 | 0 | 0 | 0 |
| mudclient-render | Mudclient (client) | 0 | 0 | 0 | 1 (all methods) | 1 (World↔Scene) | 5 |

**True "not-yet-deobfuscated" holes:** none structural. 1 MISSING-METHOD (`ClientStream.rgb`,
trivially mechanical), 1 doc-level SUMMARIZED (Scanline/FontWidths, no logic owed). Every other clean
method body is present and transcribed — including the hardest ones (`loadMapData`,
`buildSection`/`buildRoofs`, `flushEventQueue`), which were verified line-for-line against bytecode and
the rev-204 oracle.

**Highest-priority (silent render bugs):** the 10 BEHAVIORAL rows — these compile clean (or nearly)
yet render wrong, so they will not be caught by the compile gate and must be hunted explicitly against
the clean decompile.

## 4. Prioritized closing plan (execution order)

### Tier A — Mechanical (call-site rename / point at a known declaration)
Low risk, high leverage; most also retire `cannot find symbol` errors and so **advance the M2 347-pair
compile gate**.

1. **World `.jm` reader** — change `EntityDef.a(...)`@World:1382 → `GameCharacter.readFromStore(...)`.
   (Also a compile fix: `EntityDef` has no such overload.) — overlaps M2.
2. **GameModel `pa.a`@910 straggler** → `ImageLoader.sine9`. — overlaps M2.
3. **Scene render-path OBF-DRIFT** — route the ~20 raw obf owners (`e,ba,k,db,u,m,da,oa,aa,nb,ib` +
   writers `wb,gb,jb,p,cb,ua,t,ia`) to their real deob classes. Clears all 48 Scene
   `cannot find symbol`. — directly the M2 347-pair work.
4. **Frustum-global naming unification** (Scene `da/m/oa/aa/nb` ⇄ GameModel
   `ClientStream/SocketFactory/NameHash/BZip/DataStore.frustum*`) — pick one owner naming, apply to
   both files. — overlaps M2; correctness-of-linkage prerequisite.
5. **Surface `GameFonts` stand-ins** — point `antialiased[]`→`SurfaceImageProducer.k[font]`,
   `data`→`m.b`. — partly M2 (wiring).
6. **Label/comment fixes (no logic):** World `m`/`gb` field-identity comments (World:107-108); add
   Scanline merged-class doc note; correct the World ctor ramp Javadoc once #11 is settled.

### Tier B — Careful transcription (a real body / declaration owed from the clean decompile)
7. **Add `ClientStream.rgb`** — `static int rgb(int b,int r,int g){ return -1 - b/8 - (r/8)*1024 -
   (g/8)*32; }` (param order is `(B,R,G)`; verify against da.java:288 / Scene.java:1016). Pairs with
   behavioral fix #11. — overlaps M2 (unblocks World ramp call sites).
8. **drawActiveInterface structural chain** — re-model the clean fall-through priority ladder
   (client.java:1364-1434) as a first-match chain that still invokes the self-guarding tail-default
   panels (incl. social-dialog) when no flag is set. Needs careful reading, not a one-liner.

### Tier C — Behavioral fixes (silent render bugs; verify each against the clean decompile)
These are the highest-priority correctness items; none is caught by the compile gate.
9.  **Scene `textureRasterScanlines` U/V transpose** — swap texU/texV in `wb.a`@2210 and `gb.a`@2212.
10. **ImageLoader BMP width** — write/read `World.surfaceWidth`; delete `imageWidthCarrier`.
11. **World ground-ramp channel order** — fix arg order (or implement rgb per #7); correct Javadoc.
12. **Mudclient `drawActiveInterface` social-dialog route** — `drawDuelConfirm((byte)127)` →
    `drawSocialDialog((byte)127)`. — touches Mudclient (gated, see §5).
13. **Mudclient `drawChat` reset** — restore `if (param != 2) this.ak = null;`. — Mudclient (gated).
14. **Mudclient `drawUiTabMinimap` reset** — restore `if (byteParam <= 119) this.bf = null;` and stop
    discarding the byte param. — Mudclient (gated).
15. **Mudclient `drawInventoryTab` guards** — four `qc == 1` → `qc == 0`. — Mudclient (gated).
16. **(Optional, fidelity-only)** World `method422` tag `magic`→`magic ^ -14505`; WorldEntity
    `ERROR_SIGS` string-pool inputs. Benign; cosmetic.

### Tier D — Mudclient OBF-DRIFT (the systemic mislabel)
17. **World↔Scene alias inversion** — swap `scene`↔`world` across all Mudclient render methods so
    `scene`→`Ek`(lb=Scene), `world`→`Hh`(k=World); then alias-rename `drawWorld`'s raw `Ek/Hh/yd/zh`.
    Semantically inert today but required so the deob is a trustworthy oracle. Large diff, all in
    Mudclient (gated, see §5). Resolves the swap flagged in NAMING.md L16-19.

### Overlap with M2 347-pair compile gate
Tier-A items 1-5 and Tier-B item 7 each retire `cannot find symbol`/member-drift pairs, so closing them
**doubles as compile-gate progress**. The Scene OBF-DRIFT (A.3) and frustum unification (A.4) are the
densest overlap. The 10 BEHAVIORAL fixes (Tier C) mostly do **not** touch the compile gate (they
compile and run, just wrong) — they must be driven by the fidelity bug-hunt, not the build.

### Mudclient-gating
Items 12-15 and 17 (all `Mudclient.java` edits) are **gated behind the in-flight net-fix edit** to the
same file — do not touch Mudclient until that writer lands (see §5).

## 5. Sequencing

This closing pass **must run after** the two concurrent workflows finish, to avoid clobbering a moving
target and a moving oracle:

1. **After the net-fix (Mudclient writer).** A separate workflow is editing `Mudclient.java`
   (network/handler path). All Mudclient render fixes here (Tier-C items 12-15, Tier-D item 17) write
   the *same file*. Closing them earlier would either be overwritten by the net-fix or force a manual
   merge. Wait for the net-fix edit to land, then re-read `Mudclient.java` before applying any of these.
2. **After the render-fidelity bug-hunt (deob-as-oracle reader).** The behavioral findings (Tiers B/C)
   are derived by *reading* the deob against the clean decompile; a concurrent fidelity hunt may surface
   additional behavioral divergences or revise line numbers. Closing fixes before it finishes risks
   acting on a stale oracle. Let it complete, fold any new findings into this table, then execute.

Recommended order once both are done: Tier A (mechanical, advances compile gate) → Tier B (owed bodies)
→ Tier C (behavioral, verify each) → Tier D (Mudclient alias swap, last and largest). Re-run the build
after Tier A/B to confirm `cannot find symbol` count drops, and re-diff the touched methods against the
clean decompile after Tier C.

---

*Sources: `src/client/scene/{Scene,GameModel,Surface,SurfaceSprite,SpriteDecoder,SpriteScaler,ImageLoader,SurfaceImageProducer,Scanline}.java`,
`src/client/world/{World,WorldEntity,GameCharacter}.java`, `src/client/Mudclient.java`; clean oracles
`decompiled/normalized-clean/{lb,ca,ua,ba,ea,ia,pa,fb,n,k,w,ta,client}.java`; map `docs/NAMING.md`;
compile-gate context `docs/build/MILESTONE_2_RESULTS.md` (347-pair / 1,191 `cannot find symbol`).*
