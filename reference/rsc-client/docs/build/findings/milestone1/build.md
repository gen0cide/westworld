> _Verbatim execution report from the `rsc-client-milestone1-numbers` workflow (2026-06-01). See `../../MILESTONE_1_RESULTS.md` for the synthesis._

## rscplus Stage 0 / Option B — Build & Sanity Report

### 1. JDK Verification

```
$ export JAVA_HOME=/usr/lib/jvm/java-17-openjdk && export PATH="$JAVA_HOME/bin:$PATH"
$ java -version
openjdk version "17.0.19" 2026-04-21
OpenJDK Runtime Environment (build 17.0.19+10)
OpenJDK 64-Bit Server VM (build 17.0.19+10, mixed mode, sharing)
```

JDK 17.0.19 confirmed active. Default java on PATH is JDK 26; override required for every invocation.

### 2. Ant Build: `clean dist`

Invoked via launcher class:
```
java -cp ".../apache-ant-1.10.5/lib/*" org.apache.tools.ant.launch.Launcher \
  -Dant.home=.../apache-ant-1.10.5 clean dist
```

Full Ant output (abbreviated):
```
clean:
   [delete] Deleting directory .../rscplus/build/main
   [delete] Deleting directory .../rscplus/bin

compile:
    [mkdir] Created dir: .../rscplus/build/main
    [javac] Compiling 82 source files to .../rscplus/build/main
    [javac] <66 deprecation/removal warnings — all @removal for Float(float),
             Integer(int), Boolean(boolean), Double(double) box constructors,
             and java.applet.Applet/AppletStub/AppletContext>
    [javac] 66 warnings

dist:
    [mkdir] Created dir: .../rscplus/bin
      [jar] Building jar: .../rscplus/dist/rscplus.jar

BUILD SUCCESSFUL
Total time: 2 seconds
```

### 3. Artifact Confirmation

**dist/rscplus.jar** exists:
```
-rw-r--r-- 1 free free 10284002 Jun  1 18:55 dist/rscplus.jar
Size: 10,284,002 bytes  (~9.81 MiB)
```

**META-INF/MANIFEST.MF:**
```
Manifest-Version: 1.0
Ant-Version: Apache Ant 1.10.5
Created-By: 17.0.19+10 (Arch Linux)
Main-Class: Client.Launcher
```

**build/main .class count:** 230 class files compiled.

### 4. Compile Result

- Errors: **0**
- Warnings: **66** (all `[removal]` deprecation warnings — box constructor variants `new Float(f)`, `new Integer(i)`, `new Boolean(b)`, `new Double(d)` deprecated for removal in Java 17; `java.applet.Applet`/`AppletStub`/`AppletContext` likewise deprecated for removal. None of these break compilation on JDK 17.)
- Source files compiled: **82**

### 5. Runtime Patch Probe

The probe directly instantiates `JClassPatcher.getInstance()` (clean singleton, no GUI needed), pre-seeds four `HashMap<String,Boolean>` entries in `Settings` that `patch()` would otherwise NPE on via Boolean unboxing of null (`DISASSEMBLE`, `SAVE_LOGININFO`, `PATCH_GENDER`, `FIX_SFX_DELAY` — all keyed by `Settings.currentProfile` = `"custom"`), then reads every `.class` entry from `assets/rsclassic-1091943135.jar` and calls `patcher.patch(byte[])`.

Selected output (patching all 71 classes):
```
[STARTUP][INFO] Patching GameApplet (da.class)
[STARTUP][INFO] Patching random (f.class)
[STARTUP][INFO] Patching right-click menu (wb.class)
[STARTUP][INFO] Patching sound helper (sa.class)
[STARTUP][INFO] Patching renderer helper (lb.class)
[STARTUP][INFO] patching setCamera()
[STARTUP][INFO] Patching data (m.class)
[STARTUP][INFO] Patching applet (e.class)
[STARTUP][INFO] Patching font loading...
[STARTUP][INFO] Patching client (client.class)
[STARTUP][INFO] Patching renderer (ua.class)
[STARTUP][INFO] Patching java sound player (pb.class)
[STARTUP][INFO] Patching menu (qa.class)
...
Result: 71 patched cleanly / 71 total; 0 failures
```

Every class in the vanilla jar processed without exception. `patchGeneric()` runs on all 71; named-class patches fire for: `ua` (renderer), `e` (applet), `qa` (menu), `m` (data), `client`, `f` (random), `da` (game applet), `lb` (renderer helper), `sa` (sound helper), `wb` (right-click menu), `pb` (java sound player).

**javap verification of the 6 named targets** (confirm presence in jar):

| obf name | declared type |
|---|---|
| `ua` | `class ua implements ImageProducer, ImageObserver` |
| `client` | `public final class client extends e` |
| `lb` | `final class lb` |
| `e` | `public class e extends java.applet.Applet implements Runnable, MouseListener...` |
| `qa` | `final class qa` |
| `da` | `final class da extends b implements Runnable` |

All 6 confirmed present and parseable with `javap`.

### 6. What Running the Client Still Requires

From `build.xml` targets `run`/`distrun` and `Launcher.java`:

- **Display / windowing system.** `Client.Launcher extends JFrame`; it calls `setVisible(true)` at line 330. `Game.java:53` extends `JFrame implements AppletStub`. `Launcher.java:2014` calls `GraphicsEnvironment.getLocalGraphicsEnvironment()`. No `-Djava.awt.headless=true` guard. A live `$DISPLAY` (X11) or Wayland with XWayland is required.
- **World `.ini` file(s).** `Settings.Dir.WORLDS` = `<config_dir>/worlds/`. On startup, `Settings.initializeWorlds()` (line 3236) scans `Dir.WORLDS` for `*.ini` files and validates each one. Zero world files triggers a warning; an invalid `.ini` causes a modal error dialog and possibly aborts. At minimum one valid world `.ini` must exist before launch (hostname, port, RSA key).
- **`config.ini`.** `Settings.initDir()` looks for `config.ini` next to the JAR (portable mode) or in `~/.config/rscplus/` (XDG). If absent, a first-run dialog is presented (requires display).
- **Replay directory.** Created automatically at `<config_dir>/replay/` — no manual setup needed.
- **`initWorlds.lock` file protocol.** `Launcher` uses a lock file at `<config_dir>/initWorlds.lock` to coordinate world download threads; this is managed automatically.
- **`-Xmx1500m` JVM heap** (specified in the `run` target's `<jvmarg>`). The `distrun` target does not pass this — it just does `java -jar dist/rscplus.jar` from `bin/`.

In summary: to actually launch, you need `$DISPLAY`, at least one `worlds/*.ini` pointing at a reachable RSC server, and optionally `config.ini` (or accept the first-run wizard).

---

## NUMBERS

| Metric | Value |
|---|---|
| Build | **PASS** (BUILD SUCCESSFUL, 2 seconds) |
| .class count in build/main | **230** |
| dist/rscplus.jar bytes | **10,284,002** (~9.81 MiB) |
| Compile errors | **0** |
| Compile warnings | **66** (all @removal deprecation, JDK-17 harmless) |
| Patch probe | **71 / 71 classes patched cleanly; 0 failures** |
| Named targets verified | ua, client, lb, e, qa, da — all present in vanilla jar |
