> _Verbatim render-completeness audit from the render-deob-completeness-audit workflow (2026-06-01). See ../../RENDER_DEOB_GAPS.md for the consolidated plan._

# READ-ONLY Render-Completeness Audit — Mudclient render orchestration

**Scope:** the render-path methods of `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/Mudclient.java` (obf `client`), diffed method-by-method against the clean post-fix decompile `/home/free/code/rsc-hacking/westworld/reference/rsc-client/decompiled/normalized-clean/client.java`.

**Obf↔deob render-method map (verified from `// obf:` comments + clean line numbers):**

| deob method | obf | clean | role |
|---|---|---|---|
| `drawGameFrame` @7192 | `f(int)` | L6916 | per-tick 3D scene rebuild + render + blit |
| `drawWorld` @5901 | `s(int)` | L10318 | picked-model hit-test → context-menu builder |
| `drawActiveInterface` @7099 | `I(int)` | L1335 | per-tick panel dispatcher + in-world frame driver |
| `drawChat` @9213 | `l(int)` | L5407 | post-render overlays (damage splats / ground-item icons / health bars) |
| `drawUiTabMinimap` @2489 | `a(boolean,byte)` | L4450 | rotated minimap sprite + scenery/item/NPC/player dots + click-to-walk |
| `drawUiTabMagic` @2615 | `b(boolean,byte)` | L4696 | spell/prayer tab |
| `drawMinimap` @9509 | `k(int)` | L4631 | compass-ring + fatigue bar + HUD-strip borders |
| `drawInventoryTab` @9575 | `D(int)` | L5307 | tab-strip hover→`qc` state machine (no pixels) |
| `drawProgressBar` @11270 | `y(int)` | L5093 | 3-pass login/loading world-preview render |
| `drawGameSettings` @9652 | `b(int,boolean)` | L4696* | settings/privacy tab |
| `drawCharDesign` @9927 | `t(int)` | L17076 | char-design 3D preview |
| `drawSocialDialog` @10255 | `h(byte)` | L13519 | social name-entry dialog |
| `drawWelcome` @9272 | `j(int)` | L9308 | welcome/last-login box |
| `clearScreen` @11424 | `q(byte)` | L5567 | screen reset |

---

## Findings by class/method

### Mudclient — render orchestration

**OBF-DRIFT (pervasive, not behavioral):** Every render method retains obf-named instance fields and obf single-letter method calls. `drawGameFrame` alone uses ~50 distinct obf fields (`Wd, Oi, Mg, yj, qe, nc, wi, Zh, zf, xh, Ug, tg, Nc, le, ug, Si, si, pj, kg, K, eh, Eb, Be, oc, …`) and obf method calls `world.a/c`, `scene.a/f`, `surface.a/b/e`, `panelGame.a/b/c`. Field/local types are still obf class short-names (163 `lb/ca/ta/ba/ua` references in the file; e.g. `ca[] models`, `ta player`). The bodies are logically transcribed but ~90% un-renamed at the symbol level. Same for `drawWorld` (uses raw `Ek/Hh/yd/zh` — not even the `world`/`scene` aliases the other methods use → internal inconsistency).

**HIGHEST-SEVERITY — SIG-DRIFT / mislabel: World↔Scene name inversion (oracle-poisoning).**
- Deob field aliases (Mudclient L411/L429): `scene` = obf `Hh` (type `k` = **World**); `world` = obf `Ek` (type `lb` = **Scene**).
- Ground truth (clean L221 `private lb Ek;`, L143 `private k Hh;`; verified by usage): **`Ek` (lb=Scene)** owns the camera fields `Mb/X/P/G` and the renderer entry points `a(...)` (addModel), `c(-113)` (render), `a((byte)67,…)` (clearModels) — clean L7073/L7294-7297/L7332; **`Hh` (k=World)** owns terrain `bb[][]`, elevation projection `f(x,z,…)`, loaded-flag `.Z` — clean L6930/L6952/L7090.
- Consequence: in deob `drawGameFrame` the camera is set via `this.scene.Mb=3000` (L7365) and the scene is rendered via `this.world.c(-113)` (L7395) / `this.world.a(20000+i,…)` (L7310). The compiled semantics are preserved (the alias only renames the same obf field), so this is **not** a runtime bug — but as a render-fidelity *oracle* it is actively misleading: it asserts World owns the camera and renders the scene, and Scene owns terrain elevation. This is the unresolved swap flagged in NAMING.md L16-19. Fix sketch: swap the two aliases (`scene`↔`world`) across the whole render path so `scene`→`Ek`(lb) and `world`→`Hh`(k); then `scene.Mb/X/P/G`, `scene.c(-113)`, `scene.a(model,…)` and `world.f(...)` (elevation) read correctly.

**BEHAVIORAL — `drawActiveInterface` (`I`) mis-routes the social-dialog slot.** Clean L1410 calls `this.h((byte)127)` = the **byte** overload `h(byte)` = `drawSocialDialog`. Deob L7131 calls `this.drawDuelConfirm((byte) 127)` = the **int** overload `h(int)` = `drawDuelConfirm`. When a social name-entry dialog is open (`Bj != 0`) the deob renders the duel-confirm panel instead of the social dialog. (The deob's own header note L7135 even lists "h(127)=drawDuelConfirm not drawSocialDialog" as a "fix" — it is inverted.) Fix: call `drawSocialDialog((byte)127)`.

**BEHAVIORAL/structural — `drawActiveInterface` flattens the clean fall-through priority chain.** Clean `I(int)` (L1364-1434) is a sequence of independent `if(flag){ panel(); break label250; }` blocks with structural fall-through, so the tail panels `drawSocialDialog h(127)`, `drawReportNameEntry d(false)`, `drawTradeConfirmWindow N(-54)`, `drawShop M(-89)`, `drawWildernessWarning H(120)` are reached as **fall-through defaults** when no higher-priority flag fired (each self-guards internally — e.g. `H(int)` L6785 and `h(byte)` L13522 no-op on their own state). The deob converted this into one exclusive `if/else-if` ladder firing exactly one branch and never reaching the unconditional tail calls (and only sets `inWorld` on the `Bj==0` branch). The deob's equivalence claim (L7138) is only partly valid (mutually-exclusive flags), but combined with the wrong `h(127)` routing above it drops/mis-fires the social-dialog default. Fix: model the clean ladder as a first-match chain that, when no panel flag is set, still invokes the self-guarding `drawSocialDialog((byte)127)`.

**BEHAVIORAL — `drawChat` (`l`) drops the ground-item-icon reset.** Clean L5496-5498:
```
if (var25 != 2) { this.ak = (int[])null; }
```
where `var25 == param`. The deob `drawChat` (L9213-9263) omits this entirely (jumps straight from the damage-splat loop to the ground-item loop). `ak[]` is the ground-item-icon id array; without the reset stale ground-item icons persist across frames. Fix: insert `if (param != 2) this.ak = null;` after the damage-splat loop (the `mk++` profiling counter at clean L5500 is correctly omitted). (All five damage-splat overlap conditions and the health-bar/ground-item loops verified correct.)

**BEHAVIORAL — `drawUiTabMinimap` (`a(boolean,byte)`) drops the health-bar reset.** Clean L4462-4464:
```
if (var2 <= 119) { this.bf = (int[])null; }
```
The deob renamed the byte param to `unused` and dropped the clause (body jumps from box setup L2496 straight to the zoom calc L2501). `bf[]` is the entity-health-bar array consumed by `drawChat` L9259-9261; without the reset health bars never clear. Fix: keep the byte param meaningful and restore `if (<byteParam> <= 119) this.bf = null;`. (Rest of the method — dot loops, friend coloring, center marker, compass sprite, click-to-walk unprojection — verified complete and correct.)

**BEHAVIORAL — `drawInventoryTab` (`D`) mis-deobfuscates four `~qc==-1` guards.** `~qc == -1` ⟺ `qc == 0`, but the deob wrote `qc == 1` in four "enter sub-tab" hit-tests:
- clean L5314 (`qc==2` entry) `~this.qc == -1` → deob L5580 `qc == 1` (should be `qc == 0`)
- clean L5320 (`qc==3` entry) → deob L5585 `qc == 1` (should be `qc == 0`)
- clean L5328 (`qc==5` entry) → deob L5591 `qc == 1` (should be `qc == 0`)
- clean L5336 (`qc==6` entry) → deob L5597 `qc == 1` (should be `qc == 0`)

(The `qc==0` entries at clean L5309/L5324 and all the `qc!=0`/`~qc!=-1` re-select cases were transcribed correctly; threshold ranges match.) These four sub-tabs can only be selected from the wrong prior tab-state. Fix: change those four `qc == 1` guards to `qc == 0`.

**Verified faithful (no behavioral gap; OBF-DRIFT only):** `drawGameFrame` body (door/roof layer hide, region-label latches, all five entity-placement loops — players/projectiles/npcs/ground-items/scenery, ambient sparkle, both camera branches, world render `c(-113)`, the four HUD/countdown/wilderness/friends overlays, panel ticks, blit) matches clean `f(int)` L6916-7340 constant-for-constant. `drawWorld` menu-builder matches clean `s(int)` (ground-click write `rf=Qg+Hh.q[slot]`, `Cg=zg+Hh.E[slot]` matches clean L10690-10691). `drawMinimap` (`k`) matches clean compass-seam thresholds. `drawProgressBar` (`y`) 3-pass render + orange-bar loops + tile-eviction loop complete. `drawCharDesign`, `drawWelcome`, `drawSocialDialog`, `drawGameSettings` are full bodies (84-204 lines), reconstructed from clean per their FIX-vs-old notes — no stubs remain.

**MISSING-CLASS:** none beyond `scene/Scanline.java`, which is already reconstructed/present. All scene (`GameModel, ImageLoader, Scanline, Scene, SpriteDecoder, SpriteScaler, SurfaceImageProducer, Surface, SurfaceSprite`) and world (`GameCharacter, WorldEntity, World`) deob files exist.

**MISSING-METHOD / SUMMARIZED-STUB:** none in the render path. Prior stubs (`drawChat` UnsupportedOperationException, `drawTrade`/`drawTradeConfirmWindow`/`drawDuel` stubs, the in-world section of `drawActiveInterface` that was commented out) have all been reconstructed (confirmed: no `TODO`/`UnsupportedOperation`/`throw new` in the L5900-11600 render range).

---

## GAPS

**Per-class counts (Mudclient render path):**

| Gap type | Count | Methods |
|---|---|---|
| MISSING-CLASS | 0 | (Scanline already reconstructed) |
| MISSING-METHOD | 0 | — |
| SUMMARIZED/STUB | 0 | (all prior stubs reconstructed) |
| SIG-DRIFT (mislabel) | 1 systemic | World↔Scene alias inversion (`scene`=Hh/World, `world`=Ek/Scene) across **all** render methods |
| OBF-DRIFT | ~all 14 render methods | obf field names (~50 in drawGameFrame), obf method calls (`world.a/c`,`scene.f`,`surface.a/b/e`), obf-typed locals/fields (163 `lb/ca/ta/ba/ua` refs); `drawWorld` not even alias-renamed (raw `Ek/Hh/yd/zh`) |
| BEHAVIORAL | 5 | drawActiveInterface ×2, drawChat ×1, drawUiTabMinimap ×1, drawInventoryTab ×1 |

**Highest-severity items (in priority order):**

1. **[SIG-DRIFT, systemic — oracle-poisoning] World↔Scene name inversion.** Deob `scene`→`Ek`(lb=Scene)?? No — `scene`→`Hh`(k=World), `world`→`Ek`(lb=Scene). Camera (`Mb/X/P/G`), `addModel`, `render c(-113)` live on the Scene (`Ek`) but are written as `this.scene.…`/`this.world.…` inversely. Semantically inert but makes the deob a misleading render oracle. Fix: swap the `scene`/`world` aliases across the render path. Evidence: clean L221/L143 (types), L7294-7332 (camera+render on `Ek`), L6930/L7090 (`.Z`/elevation on `Hh`); deob L411/L429 (aliases), L7365/L7395.

2. **[BEHAVIORAL] drawActiveInterface routes the social-dialog slot to drawDuelConfirm.** Clean `h((byte)127)`=`drawSocialDialog` (L1410); deob `drawDuelConfirm((byte)127)` (L7131). Fix: call `drawSocialDialog`.

3. **[BEHAVIORAL] drawChat drops `if (param != 2) this.ak = null;`** (clean L5496-5498) → stale ground-item icons persist. Fix: restore the reset before the ground-item loop (deob ~L9238).

4. **[BEHAVIORAL] drawUiTabMinimap drops `if (byteParam <= 119) this.bf = null;`** (clean L4462-4464) → stale entity health bars persist. Fix: restore the reset and stop discarding the byte param (deob ~L2496).

5. **[BEHAVIORAL] drawInventoryTab mis-codes four `~qc==-1` guards as `qc==1` instead of `qc==0`** (clean L5314/5320/5328/5336 → deob L5580/5585/5591/5597) → four inventory sub-tabs selectable only from the wrong prior state. Fix: change those four `qc == 1` to `qc == 0`.

6. **[BEHAVIORAL/structural] drawActiveInterface flattens the clean fall-through priority chain** into an exclusive if/else-if ladder (clean L1364-1434), so the self-guarding tail-default panels are never reached as fall-through. Lower runtime impact than #2 (panel flags are largely mutually exclusive) but compounds #2. Fix: model as a first-match chain that still invokes the self-guarding social-dialog default when no flag is set.

Relevant files: subject `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/Mudclient.java`; oracle `/home/free/code/rsc-hacking/westworld/reference/rsc-client/decompiled/normalized-clean/client.java`; map `/home/free/code/rsc-hacking/westworld/reference/rsc-client/docs/NAMING.md`.
