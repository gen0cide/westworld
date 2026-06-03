> _Verbatim investigation report from the `rsc-client-build-options` workflow (2026-06-01). One of five parallel agents; see `../OPTIONS.md` for the synthesis._

## A1 — Deob compilability & dependency surface

### Executive verdict

**Can it be made to compile? Yes, in principle — but not by stubbing a handful of dead dependencies. The external Microsoft/Netscape dependency surface is trivially small and fully fallback-covered; the real blocker is that the deob tree is internally name-inconsistent and was never intended to compile.** It is a *comprehension reference*, exactly as the docs state (ARCHITECTURE.md:109, INDEX.md:24-25).

There are two independent problems, and they are very different sizes:

1. **Dead external deps (com.ms.* / netscape.javascript)** — confined to **4 files**, every one of which is an *optional* native path with an in-tree pure-Java/AWT twin already present. Removing/stubbing this is **LOW effort**.
2. **Internal symbol inconsistency** — the standalone subsystem classes (Surface, Scene, World, ClientStream, Panel…) were renamed to human-readable identifiers, but `Mudclient.java` (the 11,559-line assembled mega-class, the linchpin) still calls them by their **obfuscated short names** and still declares its own fields/locals with **obfuscated types and names**. Example: Surface declares `blitSprite()/drawString()/drawWorld()` but Mudclient calls `surface.a(...)` **322 times**; Mudclient locals are typed `ta npc`, `ba surface`, `GameModel m = ...b(-2)`. There are **~1,535 obfuscated-style `x.ab(` method calls in Mudclient alone** with no matching declared method in the renamed classes. This is the genuine cost and it is **HIGH effort**.

No build file (`build.xml`/`pom.xml`/`gradle`) and no `.class` artifacts exist anywhere under `reference/rsc-client/` — confirming it has never been compiled.

**Bottom line for the colleague's goals:** For **G1/G2** (fork rscplus, re-point its ASM/reflection hooks at our code), making the *whole deob* a recompilable client is a large, multi-week refactor. But that is almost certainly **the wrong target**. rscplus already runs against the obfuscated jar; the deob's highest leverage is as the **1:1 oracle for the Go render lib (G4)** and as a **named map for writing/validating rscplus hooks** — uses that need *reading*, not *compiling*. See "Recommendation" at the end.

---

### Dependency table

Every non-JDK/non-AWT reference in the tree (`grep` of all 71 files):

| Dependency | Where (files) | Classification | Fix |
|---|---|---|---|
| `com.ms.directX.{DirectSound,DSBufferDesc,DSCursors,WaveFormatEx}` | `nativeapi/DirectSoundPlayer.java` | **(a) Optional native, has pure-Java fallback** | Fallback is `audio/SourceLinePlayer.java` (obf `pb`, extends `AudioChannel`) — full `javax.sound.sampled.SourceDataLine` path (`initOutput`, `openLine`, `writeSamples`, `getBufferedFrames`). `DirectSoundPlayer` is a non-functional skeleton anyway (its write/close methods were stripped from the jar — see DirectSoundPlayer.java:22-27). **Delete the file**; drop the `AudioOutput` marker. LOW. |
| `com.ms.directX.{DirectDraw,DDSurfaceDesc,IEnumModesCallback}`, `com.ms.awt.WComponentPeer`, `com.ms.com.IUnknown`, `com.ms.win32.User32` | `nativeapi/DirectDrawModes.java` | **(a) Optional native, has pure-Java fallback** | Fallback is `nativeapi/DisplayModeSetter.java` (obf `ha`) — pure AWT `GraphicsDevice.setFullScreenWindow/setDisplayMode/getDisplayModes`, same `enterFullscreen/exitFullscreen/listAvailableModes` surface. **Delete `DirectDrawModes`**; route the loader's fullscreen path to `DisplayModeSetter`. LOW. |
| `com.ms.dll.{Callback,Root}`, `com.ms.awt.WComponentPeer`, `com.ms.win32.User32` | `nativeapi/Win32MouseCallback.java` | **(a) Optional native, has pure-Java fallback** | Fallback is `shell/RobotCursor.java` (obf `j`) — pure AWT `Robot.mouseMove` + transparent `createCustomCursor` for hide/show. Covers cursor-warp + hide/show + custom cursor. **Delete `Win32MouseCallback`**; route cursor ops to `RobotCursor`. LOW. |
| `netscape.javascript.JSObject` | `util/JSBridge.java` | **(b) Applet/browser plumbing** | One method: `JSObject.getWindow(applet).call(name, null)` — navigates the host page on logout/crash (called from `GameShell.navigateAway`). A standalone app has no browser. **Stub `JSBridge.call(...)` to return null** (or no-op). LOW. |
| `java.applet.Applet/AppletContext` | `shell/GameShell.java`, `util/JSBridge.java`, Mudclient | **(b) Applet/browser plumbing** | `GameShell extends Applet`. The whole applet-context delegation (`getParameter/getCodeBase/getDocumentBase/getAppletContext`) is already written to **fall through to the standalone `GameFrame` path** when `InputState.gameFrame != null` (GameShell.java:216-247, 538-545, 626-634, 776-822). Keep `extends Applet` (JDK 8 still has it) or re-base on a `Panel`/`Canvas`. LOW–MED. |
| `java.lang.reflect.*`, `Class.forName("ha"/"j"/"rb")` | `shell/LoaderThread.java`, `util/ListNode.java`, `net/ProxySocketFactory.java` | **(c) Self-referential reflection — minor** | `LoaderThread` loads `DisplayModeSetter`/`RobotCursor`/`DirectSoundPlayer` **by obfuscated name** (`Class.forName("ha")`, `("j")`, `("rb")`) and invokes obf method names (`"enter"`, `"movemouse"`, `"showcursor"`, `"setcustomcursor"`, `"listmodes"`). After renaming, these strings won't resolve — but they're in `try{}catch` blocks that degrade gracefully. Replace reflective calls with direct typed calls to the renamed classes. LOW–MED. |
| `java.awt.Robot` | `shell/RobotCursor.java` | **JDK, fine** | Standard AWT; works headful. No action. |

That is the **complete** external surface. `grep` for `native`/JNI found only the English word in comments and `getSystemEventQueue`/`getPeer` (legit AWT). No JNI declarations exist.

---

### Entry-point / lifecycle situation

Per ARCHITECTURE.md:9-21, the lifecycle is well-defined and **present in the tree**:

- **`GameShell` (obf `e`)** — `src/client/shell/GameShell.java`, `extends java.applet.Applet implements Runnable, MouseListener, MouseMotionListener, KeyListener`. Owns the game `Thread` and the main loop `run()` (GameShell.java:257-399: boot → FPS pacing → `handleInputs()` ×N → `draw()`). Has the **standalone launch path already**: `startApplication(...)` (GameShell.java:562-594) builds a `GameFrame`, sets `InputState.gameFrame`, creates the `LoaderThread`, optionally runs a `CacheUpdater`, then `new Thread(this).start()`.
- **`GameFrame` (obf `qb`)** — `src/client/shell/GameFrame.java`, `extends java.awt.Frame`. The standalone host window; pure AWT, no external deps. Compiles as-is (uses deprecated `show()/resize()` — JDK 8 ok).
- **`LoaderThread` (obf `c`)** — `src/client/shell/LoaderThread.java`, the privileged I/O + OS-bridge worker (sockets, threads, cache files, fullscreen, cursor, clipboard) via a `ListNode` work-queue. Pure JDK except the by-name reflection above.
- **`Mudclient` (obf `client`)** — `src/client/Mudclient.java:74` `public class Mudclient extends GameShell`. The game itself; overrides `startGame()/handleInputs()/draw()/onClosing()`.

**What an application `main()` would need:** instantiate `Mudclient`, then call the standalone entry it already inherits — conceptually `new Mudclient().startApplication(resizable, title, loaderArg, loaderName, port, doUpdate, storeFlag, 512, 384)`. `startApplication` constructs the `GameFrame`, wires `LoaderThread`, kicks the game thread. **The lifecycle plumbing is complete and self-contained** — no missing entry class. (In the real obf jar a tiny outer launcher set these args; that launcher is not in the 71-class tree, but it's ~10 lines to write.) The only true gaps are (i) the dead-dep stubs above and (ii) the symbol-consistency problem below.

---

### Internal-consistency spot-check (the real blocker)

The tree is **not** internally consistent at the symbol level. The deobfuscation renamed *standalone* classes thoroughly but did **not** propagate those renames into `Mudclient.java`, and several caller/callee signatures drifted. Concrete evidence:

1. **Mudclient calls subsystems by obf method names.** `surface.a(` ×322, `surface.b(` ×58, `surface.c(` ×48, `clientStream.b(` ×70, `world.a(` ×30, `f.e(` ×79, `mg.a(` ×73, `objectModels[i].b(-2)`, etc. — **~1,535 short obf-name calls in Mudclient**. But `Surface.java` declares `blitSprite/drawString/drawWorld/drawMinimapSprite/...` (readable). `surface.a(...)` matches **no** declared Surface method. Same for `World`, `ClientStream`, `GameModel`, `Panel`.
2. **Mudclient still uses obf types and field names.** Locals like `ta npc = this.Ff[n]`, `ba surface`, fields `mg/zh/Jh/Af/te/We/rg/Tb`. `surface` is even declared `private ba surface;` (Mudclient.java:416) — type `ba` (= `SurfaceSprite`) — so it's a half-rename: the variable is renamed, its type is not.
3. **Caller/callee signature drift even inside the renamed half.** `LoaderThread.dispatch()` calls:
   - `this.win32MouseCallback.moveMouse(23529, y, x)` and `.showCursor(-4, comp, visible)` (LoaderThread.java:751,774) — but `Win32MouseCallback` declares `setCursorPosition(int,int,int)` and `installHook(int,Component,boolean)`. **`moveMouse`/`showCursor` don't exist on that class.**
   - `this.directDraw.restoreDisplayMode((Frame) node.obj, 0)` (LoaderThread.java:707) — but `DirectDrawModes` declares `exitFullscreen(Frame)` (no 2-arg `restoreDisplayMode`).
   - `this.directDraw.listModes((byte) -100)` (LoaderThread.java:843) — but `DirectDrawModes.listModes()` takes **no args**.

   These are exactly the cases where the renamer rewrote one side of a call but the docstrings even admit the other side was meant to be reached by reflection/obf-name. Several are inside `try/catch` so they'd "work" by failing silently, but they won't *compile*.

So even after deleting all 4 com.ms/netscape files, the tree fails to compile primarily inside `Mudclient.java` (every `surface.a(...)`, `world.a(...)` is an "unknown method/symbol") and at the `LoaderThread → nativeapi` seams. The standalone subsystem files (scene/, world/, net/, data/, audio/ engine core) are individually much closer to internally consistent — the docs' "100% method presence" claim for the renderer is about *presence*, not *call-site name agreement with Mudclient*.

---

### Ranked, concrete task list (to make it a pure-Java client)

Ordered by dependency / leverage. Effort tiers reflect that the **dead deps are cheap and the name reconciliation is the mountain.**

1. **[LOW] Stub `util/JSBridge.java`** — drop the `netscape.javascript` import; make `call(String, Applet)` return `null`. Removes 1 of 4 external files.
2. **[LOW] Delete `nativeapi/DirectSoundPlayer.java`** and the `audio/AudioOutput` marker interface (its only implementor). Audio is fully served by `audio/SourceLinePlayer.java`. Removes 1 external file.
3. **[LOW] Delete `nativeapi/DirectDrawModes.java` and `nativeapi/Win32MouseCallback.java`.** Removes the last 2 external files — **the entire com.ms/netscape surface is now gone.**
4. **[LOW–MED] Rewire `shell/LoaderThread.java` to the AWT fallbacks directly.** Replace the `isDirectDraw`/`isWin32` Win32 branches and the `Class.forName("ha"/"j"/"rb")` reflective calls with direct typed calls to `DisplayModeSetter` (fullscreen) and `RobotCursor` (cursor/move). Fixes the `moveMouse`/`showCursor`/`restoreDisplayMode(Frame,int)`/`listModes(byte)` signature mismatches noted above. ~6 call sites.
5. **[LOW–MED] Decide the `GameShell` base.** Keep `extends java.applet.Applet` (JDK 8 compiles it) for minimal churn, OR re-base on `java.awt.Panel`/`Canvas` and delete the applet-context delegation. The standalone path already exists; this is mostly deletion.
6. **[MED] Make the standalone subsystem files compile in isolation.** scene/, world/, net/, data/, util/, ui/, audio/ engine core. Resolve intra-package signature drifts (like the `LoaderThread`→nativeapi ones) and any remaining obf-name residue. Iterate `javac` per package, lowest-dependency-first (util → net/data → world/scene → ui/audio). This is bounded and tractable because these files were the focus of the rename/verify passes.
7. **[HIGH — the real cost] Reconcile `Mudclient.java` against the renamed subsystems.** ~1,535 obf-name method calls + obf field types/names (`ta`, `ba`, `Ff`, `mg`, `zh`…) must be mapped to the readable declarations. `docs/NAMING.md` (the verified obf↔name map for all 71 classes) is the key, but it's class-level; **a method-/field-level obf↔name map does not appear to exist in the docs** and would have to be derived (the per-class files carry `// obf: X` comments on members, so it's extractable but laborious). Until this is done, the assembled client does not link even though every callee exists somewhere. Plan for this to dominate the schedule.
8. **[LOW] Write a `main()` launcher + `build.xml`/Gradle.** ~10-line launcher calling the existing `startApplication(...)`; a JDK-8 build script. Trivial once 1–7 land.

**Effort tier summary:** Tasks 1–5 (kill all dead deps, fix the native seam, settle the entry point): **~1–2 days, LOW**. Tasks 6–8 (compile the subsystems + assemble): **MED, ~1 week**. Task 7 (Mudclient ↔ subsystem name reconciliation): **HIGH, the bulk of the project — multiple weeks**, because Mudclient is half-renamed by design.

---

### Recommendation against the colleague's goals

- The deob's **external dependency surface is a non-issue** — 4 files, all with working in-tree pure-Java/AWT fallbacks already written (`SourceLinePlayer`, `DisplayModeSetter`, `RobotCursor`) plus a 1-line `JSBridge` stub. Anyone claiming "it can't run because of DirectX/Netscape" is wrong; that part is a day's work.
- The **real cost is internal name reconciliation inside `Mudclient.java`** (~1,535 obf-name call sites + obf field types), which the project deliberately left half-done because the goal was readability, not recompilation.
- **For G1/G2 (re-point rscplus at our code):** do **not** try to compile the whole deob and have rscplus load it. rscplus's value is that it patches the *real, working, signed* obf jar by obf name (`JClassPatcher` dispatches on `ua/e/qa/...`; `Reflection.java` binds obf fields/methods). Our deob's `// obf:` annotations + `NAMING.md` are the perfect **dictionary to author/verify those hooks**, but the patch target should stay the obf jar.
- **For G4 (Go render lib 1:1 audit):** this is where the deob shines and **needs zero compilation**. `scene/Surface.java`, `scene/Scene.java`, `scene/GameModel.java`, `world/World.java`, `world/WorldEntity.java` are the fully-expanded, verified rasterizer/terrain/visibility code (ARCHITECTURE.md:25-45) — read them directly as the oracle for the Go port (e.g. `World.route`, `Scene.polygonsOverlap`, projection/scanline fill). The half-renamed `Mudclient` does not impede this at all.

Relevant files: external-dep files `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/nativeapi/{DirectDrawModes,DirectSoundPlayer,Win32MouseCallback}.java` and `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/util/JSBridge.java`; fallbacks `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/audio/SourceLinePlayer.java`, `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/nativeapi/DisplayModeSetter.java`, `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/RobotCursor.java`; lifecycle `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/{GameShell,GameFrame,LoaderThread,InputState}.java`; the reconciliation blocker `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/Mudclient.java` (11,559 lines); render oracle `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/scene/Surface.java` and siblings.
