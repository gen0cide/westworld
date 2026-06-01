> _Verbatim report from the deob-compile-and-net-depth workflow (2026-06-01)._

# RSC Deob Subsystem Compile Report (JDK 17)

## (a) PASS / FAIL

**FAIL** — the subsystems do **not** yet compile clean. Authoritative subsystem-only compile (JDK 17, `sourcepath` = all `client/**` except `Mudclient.java`, output `/tmp/deob-build`):

- **Residual: 1281 errors**, of which **17 are Mudclient-dependent (out of scope)** → **1264 subsystem-internal errors**, **7 warnings** (java.applet deprecation, as expected).
- Started at 4424 errors after type-widening; the dead-dep fixes, member widening, import reconciliation, and the named drift fixes drove the count down, then **uncovered** the true blocker.

**Residual is dominated by pervasive cross-class member-name drift** (1191 `cannot find symbol`; **347 distinct (owner-type, member) pairs**), NOT visibility. Members were renamed in their owning deob class but still referenced by old obf/short names elsewhere — e.g. `GameModel.e(...)` (41×), `ISAAC.emptyString(...)` (22×), `Surface.getLineHeight/getStringWidth/drawString(...)`, `StreamBase.and(...)`, `EntityDef.getCount(...)`, `NameTable.lookupName(...)`, and the BZip→`DecodeBuffer` BZip2-state cluster (`s.input`/`s.origPtr`/`s.limit`… 36 fields). Other residual: 35 `int→char[]`, 17 `drawSprite` overload mismatches, 14 `does not override`, plus a few `cannot be applied` signature sites.

This contradicts Milestone-1's prediction of "~78 member + 7 drift." **M1 under-measured**: it widened only top-level types and never added the cross-package imports, so this drift stayed hidden inside the `cannot find symbol` cascade that M1 dismissed as noise. Resolving the remaining 347 pairs is a genuine per-method ground-truth reconciliation (re-deobfuscation), beyond the mechanical scope. Residual per package: scene 376, net 293, world 170, audio 135, ui 116, shell 90, util 76, data 32.

## (b) Ordered change list

**Files deleted (4)** — dead Microsoft/Netscape deps:
- `nativeapi/DirectSoundPlayer.java`, `nativeapi/DirectDrawModes.java`, `nativeapi/Win32MouseCallback.java` (all `com.ms.*`), `audio/AudioOutput.java` (only DirectSoundPlayer implemented it; confirmed no other referents).

**File added (1)**: `scene/Scanline.java` — `Scanline` (obf `n`) was referenced by `Scene.java` (`new Scanline()`, `Scanline[]`, `sl.d/k/e/l`) but never emitted as a file. Reconstructed faithful to clean decompile `n.java` instance fields (`int d,k,e,l` = startX/endX/startS/endS).

**Stubs / param fixes:**
- `util/JSBridge.java` — dropped `import netscape.javascript.JSObject`; `call(...)` now no-op returning `null`.
- `ui/Panel.java` — reserved-word param `byte _` → `byte unusedGuard` (Java 9+).

**Visibility split — WIDENED, package structure preserved (collapse avoided):**
- **49 top-level types** widened package-private → `public` (line-start `class/interface/enum`, incl. 2 in non-UTF8 "binary" files DataStore/FontWidths done byte-faithfully). 0 package-private top-level types remain.
- **1024 members** widened: 1019 default-access fields/methods/ctors → `public` (brace-depth-1 declarations only, never locals; verified no false positives in method bodies), plus **5 genuinely-private** cross-package-accessed fields (`GameModel.{lightAmbience,unpickable,transformState}`, `AudioChannel.{bufferSizeFrames,requestedBufferFrames}`).
- **~97 cross-package imports added** across ~25 files (compiler-driven; e.g. `client.net.StreamBase` into LinkedQueue/FilterNode, `client.util.ErrorHandler` into net classes, static-access qualifiers like `InputState`/`Packet`/`StreamFactory`). This is what surfaced the hidden member drift.

**LoaderThread rewiring** (reflection → direct typed calls; Win32 branches dropped):
- Removed imports of deleted `DirectDrawModes`/`Win32MouseCallback`; removed fields `directDraw`/`win32MouseCallback`; retyped `displayModeSetter` `Object`→`DisplayModeSetter` (obf `ha`) and `robotCursor` `Object`→`RobotCursor` (obf `j`, in `client.shell`).
- Constructor: `Class.forName("ha"/"j").newInstance()` → `new DisplayModeSetter()` / `new RobotCursor()`; dropped dead `forName("rb")` (DirectSoundPlayer) and the `new DirectDrawModes()`/`new Win32MouseCallback()` Win32 branches.
- dispatch(): reflective `getMethod(...).invoke(...)` → direct calls, fixing drifted names — `exit`→`exitFullscreen()`, `enter`→`enterFullscreen(frame,w,h,x,y)`, `listmodes`→`listAvailableModes()`, `movemouse`→`moveMouse(x,y)`, `showcursor`→`setCursorVisibility(comp,vis)`, `setcustomcursor`→`setCustomCursor(comp,pixels,w,h,hotspot)`. **0 residual LoaderThread errors reference any deleted class.**

**The 7 genuine drift sites — all FIXED (verified 0 errors remain at each), each against jar/clean-decompile ground truth:**
1. `net/Packet.java` (telemetry `outgoingTelemetry.write/seek`) — **already resolved** by member widening (the methods exist on the resolved type; no edit needed).
2/3/4. `scene/GameModel.java` 302/304/346 — `CacheFile.getUnsignedShort(offset,(byte)7,data)` → `getUnsignedShort(data, offset)`. Ground truth: `CacheFile.getUnsignedShort(byte[] buffer, int offset)` — obf `a(int,byte,byte[])` was reordered and its dead `byte` dropped (per CacheFile.java's own doc).
5. `shell/GameShell.java` 658 — `JSBridge.call(STR[59],(byte)..,applet)` → `JSBridge.call(STR[59], applet)` (both sites). Ground truth: deob `call(String, Applet)` stripped the dead anti-tamper `byte` (per JSBridge.java doc).
6. `shell/LoaderThread.java` (orig 413) — `CachePath.resolveFile(idx,null,"random.dat",0)` (4-arg) → `resolveOrCreateCacheFile(idx,null,"random.dat",0)`; `CachePath.getCacheFile(2,...)` → `resolveFile(2,...)`; `CachePath.init(...)` → `initialize(...)`. Ground truth: CachePath.java declares `initialize(int,byte,String)`=obf `r.a(int,byte,String)`, `resolveFile(int,String)`=obf `r.a(int,String)`, `resolveOrCreateCacheFile(int,String,String,int)`=obf `r.a(int,String,String,int)` — caller used wrong names for all three.
7. `world/World.java` 880 (orig 877) — `surface.a(true) // blackScreen` → `surface.blackScreen(true)`. Ground truth: `Surface.blackScreen(boolean)` is the only boolean method matching the inline comment.

Plus the related net-alias fix: `shell/LoaderThread.java` import `client.net.StreamFactory` → `client.util.StreamFactory` (StreamFactory lives in `util`; the `net` import was wrong); and `StreamFactory.createSocketFactory(4718, port, host)` → `createSocketFactory(port, host)` (deob stripped the dead `4718` magicKey first arg).

## (c) Mudclient.java

**NOT modified.** Verified byte-identical to HEAD (`git diff --quiet HEAD` → clean). It was temporarily moved out of the source tree only during compilation (so the `sourcepath` excludes it per spec) and restored afterward; it is present and untouched in the working tree.

## (d) git diff --stat (uncommitted; HEAD still 5b78e9b)

`git diff --stat HEAD -- reference/rsc-client/src/client`: **69 files changed, 1219 insertions(+), 1894 deletions(-)** (4 deleted + 65 modified) plus **1 new untracked file** `scene/Scanline.java` (42 lines) = 70 changed paths. (Several `data`/`net` files are non-UTF8 and shown by git as `Bin`; their `public ` widenings are byte-faithful.) Nothing committed.

## NUMBERS

- subsystems compile: **FAIL**
- residual error count: **1281** (subsystem-only compile; **1264** excluding the 17 out-of-scope Mudclient-dependent refs) — 7 warnings
- types widened: **49** (package-private top-level types → public; 0 remain)
- members widened: **1024** (1019 default-access fields/methods/ctors + 5 private cross-package fields)
- drift sites fixed: **7/7** named sites resolved (+ LoaderThread StreamFactory package/arg drift); **347** distinct member-name-drift pairs remain unresolved (genuine re-deobfuscation, out of mechanical scope)
- files deleted: **4** (+1 added: Scanline.java)
