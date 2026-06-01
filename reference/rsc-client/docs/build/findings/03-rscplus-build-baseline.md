> _Verbatim investigation report from the `rsc-client-build-options` workflow (2026-06-01). One of five parallel agents; see `../OPTIONS.md` for the synthesis._

## A3 — rscplus build & run baseline

This is the control baseline: **rscplus as-is compiles and packages cleanly**, and its only hard runtime dependency for an offline/replay run is a JDK that still ships `java.applet`. Full findings below.

---

### 1. Build requirements

| Item | Value |
|---|---|
| Build tool | **Apache Ant** (`build.xml`, default target `run`). Targets: `compile`, `dist`, `distrun`, `run`, `test`/`test-compile`, `clean`, `doc`, `setversion`, `format-source`. |
| Source layout | `src/` (main, 82 `.java` files across `Client`, `Game`, `Replay/{common,game,scraper}`, `Client/Extensions`, `Client/FlatLaf`) + `test/` (JUnit). |
| Main class | `Client.Launcher` (set in build.xml `jar.main-class` and the dist manifest). |
| Output | `build/main` (classes, 230 produced) → `dist/rscplus.jar` (fat jar, ~10 MB, all `lib/*.jar` shaded in via `zipgroupfileset`, minus junit/hamcrest). |
| README-stated JDK | "jdk 1.7 or 1.8"; `.classpath` pins `JavaSE-1.8`. The Eclipse `.project`/`.settings` target Java 8. |
| **Actual JDK ceiling (key finding)** | The code imports `java.applet.Applet` / `AppletStub` / `AppletContext` (`Launcher.java`, `Game.java`, `Client.java`, `ScaledWindow.java`). `java.applet` was **removed in JDK 21+**. Verified here: present in JDK 17 (`javap java.applet.Applet` succeeds), **absent in JDK 26** (`class not found`). So the buildable/runnable range is effectively **JDK 8–17** (8 ideal per the client's own warning in `Launcher.main`, which logs "rsc wasn't designed for Java version N… for best results use version 8" when `javaVersion >= 9`). |
| No-error compile | Only deprecation warnings (`new Float/Integer/Double`, `Applet`, `newInstance`) — **66 warnings, 0 errors**. |

The compile is also a Microsoft-J++-jar caveat-free zone for *rscplus itself*: rscplus is ordinary Java that *wraps* the obfuscated jar at runtime; it does not compile against the obfuscated classes (it reaches them via ASM bytecode patching + reflection on obfuscated names like `ua`, `client`, `qa`).

---

### 2. Exact dependency list (`lib/`)

All are bundled in the repo (no network fetch needed). On the **main compile/runtime classpath** (build.xml includes `lib/*.jar`, excludes junit + hamcrest):

- `asm-5.0.4.jar`, `asm-tree-5.0.4.jar`, `asm-util-5.0.4.jar` — bytecode patching engine (`JClassPatcher`, ~5858 lines).
- `commons-compress-1.18.jar` — replay/archive handling.
- `json-20201115.jar` — config/world JSON, server-extension data.
- `jsoup-1.16.1.jar` — HTML scraping (hiscores/population).
- `jansi-1.18.jar` — colorized console (`Logger`).
- `jinput-2.0.9.jar` + `lib/jinput-natives/` (`.dll`/`.so`/`.jnilib`) — gamepad/joystick input; natives extracted at startup via `Launcher.extractJInputNatives()`.
- `jna-5.13.0.jar` + `jna-platform-5.13.0.jar` — Windows GDI32 DPI scaling (`Launcher.getScaleFactor()`) and OS integration.
- `flatlaf-3.1.1.jar` — Swing look-and-feel for the config/launcher windows.
- `activation.jar` — JavaBeans Activation (JDK 9+ removed it from the JRE, so it's bundled).
- `orange-extensions-1.3.0.jar` — compile-time stub for `com.apple.eawt.Application` (macOS dock menu in `Launcher`).
- `coreapi-2.0.9.jar` — present in `lib/` (transitive/aux; not in `.classpath`, picked up by build.xml's `lib/*.jar` glob).

**Test-only** (separate path): `junit-4.12.jar`, `hamcrest-core-1.3.jar`.

**Tooling**: `tools/google-java-format-1.7-all-deps.jar` (the `format-source` target). `binaries/build-scripts/*.sh` are packr/AppImage/installer builders (Linux AppImage, macOS dmg, Win32/64) — not needed to build or run from source.

---

### 3. Compile attempt result (this environment)

**Tooling present:** `java`/`javac` **26.0.1** on PATH (too new — see below), plus `/usr/lib/jvm/java-17-openjdk` (17.0.19). **`ant` is NOT installed** (`which ant` → not found; not in pacman). A usable Ant exists only as a bundled copy under the openrsc tree: `/home/free/code/rsc-hacking/openrsc/Portable_Windows/apache-ant-1.10.5` (its `bin/ant` shell wrapper is non-executable on this fs, so I invoked it via its launcher class). `gradle` is present but unused by this project; no `mvn`. No Linux JDK 8 anywhere (the only JDK 8 on disk is a **Windows** `zulu8…win_x64` build under openrsc — `javac.exe`, won't run on Linux).

**What I ran** (Ant 1.10.5 via its launcher jar, forced onto **JDK 17**):
```
java -cp <ant-1.10.5/lib/*> org.apache.tools.ant.launch.Launcher -Dant.home=<ant> clean compile
```
**Result: `BUILD SUCCESSFUL`** — 82 source files → 230 `.class` files, 66 deprecation warnings, **zero errors**. The `dist` target also succeeded, producing `/home/free/code/rsc-hacking/rscplus/dist/rscplus.jar` (10,284,002 bytes, runnable manifest `Main-Class: Client.Launcher`).

**Failure mode to flag:** building or running with the **default `java` (JDK 26)** will fail because `java.applet` no longer exists. A fork must either pin JDK ≤17 (install a Linux JDK 8/11/17 — none of 8/11 is currently installed; only 17 and 26 are) **or** refactor away the `java.applet.*` types (replace `Applet`/`AppletStub`/`AppletContext` with the `java.awt.Panel`/manual stub the deob notes already describe as the pure-AWT degrade path).

---

### 4. Runtime / server config

**Launch chain:** `Launcher.main(args)` (ignores `args` entirely — no CLI parsing) → static init (server extensions, world files, settings, Swing windows, tray icon) → `Launcher.init()` (Swing progress window) → `run()` → `JConfig.create(world)` → `new JClassLoader(...).fetch("/assets/rsclassic-1091943135.jar")` (reads the obfuscated jar from inside its own jar/classpath, ASM-patches every class via `JClassPatcher.patch()`) → `loadClass(config.getJarClass())` (obf entry class `client`, from `initial_class=client.class`) → instantiates it as an `Applet` and `Game.start()`.

**Where server config lives:**
- **Per-world `.ini` files** in the config dir, parsed by `Settings.initWorlds()` → `World` objects → `WORLD_URLS`/`WORLD_PORTS`/`WORLD_RSA_PUB_KEYS`/`WORLD_RSA_EXPONENTS`/`WORLD_SERVER_TYPES`/`WORLD_SERVER_EXTENSION` maps. **Config dir on Linux:** `$XDG_CONFIG_HOME/RSCPlus` or `~/.config/RSCPlus` (worlds in its `Worlds/` subdir). **There are no world files shipped in the repo** — on first run `WORLDS_TO_DISPLAY=0`, `noWorldsConfigured=true`; the user must add a world (URL, port, server type, RSA modulus/exponent) through the GUI, or a subscription downloads them.
- **`JConfig`** turns the selected world into the applet's `jav_config.ws`-style params: `m_data["codebase"]="http://<worldURL>/"`, RSA exponent/modulus copied from the world, default port `43594` (`Replay.DEFAULT_PORT`). `JConfig` ships hard-coded defaults (RSA modulus, `VERSION=124`/`SUBVERSION=2`, `cachesubdir=rsclassic`) that are overwritten per-world.
- **Cache content** is bundled in `assets/content/` (`content0…content11`), `assets/map/`, `assets/bank/`, fonts, etc. — so the client does not need a cache server.

**Does it need a live server?** For real play, **yes** — it makes a raw TCP socket to `WORLD_URL:WORLD_PORT` and does RSA login. **But it also has a fully offline path:** the **Replay system**. `JConfig.changeWorld()` detects the special replay world and sets `codebase=http://127.0.0.1/`; `Replay.initializeReplayPlayback()` spins up `ReplayServer` (a `ServerSocketChannel` on loopback) that **replays recorded server packets to the vanilla client** — no external server, no real network. This is the offline/"drone scenario" hook for G3: replays are captured as `version.bin`/`keys.bin`/`in.bin.gz`/`out.bin.gz`/`metadata.bin` in a `replayDirectory`, and `ReplayServer` feeds them back deterministically.

---

### 5. Headless / scripted-launch situation (relevant to G3 drones)

**There is no headless or CLI path today.** Findings:

- **No `args` handling:** `Launcher.main(String[] args)` never reads `args`. The only external knobs are JVM system properties read at startup — `usingBinary`, `binaryVersion`, `downloadWorlds` (all for packaged-binary/world-subscription flavoring), plus standard `os.*`/`user.home`/`APPDATA`/`XDG_CONFIG_HOME`. **None selects a world, starts a replay, or chooses an account.**
- **No headless support:** zero hits for `java.awt.headless`, `GraphicsEnvironment.isHeadless`, or `-Djava.awt.headless`. `Launcher extends JFrame`, calls `setVisible(true)` unconditionally in `init()`, instantiates `ScaledWindow`, `ConfigWindow`, `WorldMapWindow`, `QueueWindow`, a tray icon, and Swing dialogs (update prompts, world-validation warnings) on the main path. It is a hard **Swing/AWT GUI app**; running it requires a display (or `Xvfb`/virtual framebuffer).
- **Replay is GUI/event-driven, not scriptable:** `Replay.initializeReplayPlayback()` is reached only through in-game UI — `Client.java:1391` (the `Renderer.replayOption` menu choice) and `QueueWindow`/`ReplayQueue` (the queue table + buttons wired in `ConfigWindow`/`Renderer`). There is no programmatic "start replay X and exit" entry point.

**Implication for the cradle (G3/G2):** to drive many drones, a fork will need to *add* a scripted entry point — e.g. parse `args`/system-properties in `Launcher.main` to (a) auto-select a world or a replay directory and (b) call `Replay.initializeReplayPlayback()` directly — and run under a virtual display (Xvfb) since the AWT/Swing dependency is unavoidable without a deeper rewrite. The replay-driven `ReplayServer` loopback design is the right substrate (no live server needed), but the launch automation around it does not yet exist.

---

### Bottom line

- **Builds cleanly today** via Ant under **JDK 8–17** (verified `BUILD SUCCESSFUL` on JDK 17, both `compile` and `dist`). All deps are vendored in `lib/`. The dist jar is at `/home/free/code/rsc-hacking/rscplus/dist/rscplus.jar`.
- **This environment is missing two things** to reproduce a stock build: **Ant is not installed** (only a non-executable bundled copy exists under openrsc), and **the default JDK is 26**, which breaks `java.applet`. Install Ant + a JDK ≤17 (only 17 and 26 are present; no 8/11).
- **Running** needs a world `.ini` (URL/port/RSA) in `~/.config/RSCPlus/Worlds/` for live play, or a replay directory for the offline `ReplayServer` path. **No live server is strictly required** thanks to replays.
- **No headless/CLI/scripted launch exists** — it's a pure Swing/AWT GUI with no `args` parsing; the cradle will need an added programmatic launch hook plus a virtual display.

Key files: `/home/free/code/rsc-hacking/rscplus/build.xml`, `/home/free/code/rsc-hacking/rscplus/.classpath`, `/home/free/code/rsc-hacking/rscplus/src/Client/Launcher.java` (main, lines 1056+), `/home/free/code/rsc-hacking/rscplus/src/Client/JConfig.java`, `/home/free/code/rsc-hacking/rscplus/src/Client/JClassLoader.java`, `/home/free/code/rsc-hacking/rscplus/src/Client/JClassPatcher.java`, `/home/free/code/rsc-hacking/rscplus/src/Client/Settings.java` (worlds: 3244+, config dir: 2935+), `/home/free/code/rsc-hacking/rscplus/src/Game/Replay.java`, `/home/free/code/rsc-hacking/rscplus/src/Game/ReplayServer.java`.
