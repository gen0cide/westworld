> _Verbatim runtime bring-up research from the deob-runtime-bringup-design workflow (2026-06-01)._

## FINDINGS

### How the deob client BOOTS (what EXISTS)

The lifecycle is fully present in the tree and is **applet-superclass-but-standalone-capable**. `GameShell extends java.applet.Applet` (GameShell.java:74) but has a complete standalone path that bypasses all applet/AppletStub plumbing when `InputState.gameFrame != null`.

**The deob already HAS its own `main()`** — `Mudclient.main(String[])` at `Mudclient.java:544`. It is not just a stub; it is the real reconstructed entry. There is no need to write a new launcher; the ~10-line launcher the task asked about already exists inside Mudclient. Its flow:

1. `GameShell.audioQueue = BZip.instance` (Mudclient.java:549)
2. `BZip.nodeId = Integer.parseInt(args[0])` — args[0] = node/world id (int)
3. Select audio backend from `args[1]` ("live"/"rc"/"wip") → `LinkedQueue.audioFactory`
4. `new Mudclient()`; `client.isApplet = false`
5. Parse `args[2..]` for "members"/"veterans" flags
6. Call `client.a(false, "local.runescape.com", 32+prio, "classic", nodeId+7000, (byte)112, displayDepth, 512, 346)` (Mudclient.java:594-604) — this **is** `GameShell.startApplication(...)`.

**`GameShell.startApplication(...)`** (GameShell.java:576-612, obf `a(boolean,String,int,String,int,byte,int,int,int)`) is the standalone bring-up. Params: `(resizable, title, loaderArg, loaderName, port, doUpdate, storeFlag, width, height)`. It:
- sets `appletWidth/appletHeight`,
- `InputState.gameFrame = new GameFrame(this, 800, 600, title, resizable, false)` (creates + shows the AWT window — GameFrame.java:150-170, pure AWT, compiles as-is),
- reflectively disables focus traversal,
- `LinkedQueue.sharedInt = storeFlag`, `loadingStep = 1`,
- builds the `LoaderThread` (`ImageLoader.loaderThread = ImageLoader.imageWidthCarrier = new LoaderThread(loaderArg, loaderName, 0, true)`),
- if `doUpdate > 20`, runs `CacheUpdater.downloadAndVerifyCrcs(new URL("http","127.0.0.1",port,""), this, -91)` — an HTTP content-pack pull,
- `gameThread = new Thread(this); gameThread.start()` → runs `GameShell.run()`.

**`GameShell.run()`** (GameShell.java:272-413) is the main loop: one-time boot (`loadJagex()` loads `jagex.jag`→logo + 8 bitmap fonts via `Panel.loadFont`, then `startGame()`), then FPS-paced loop calling `handleInputs(119)` N times and `draw(false)` once per frame. `Mudclient` overrides `startGame`/`handleInputs`/`draw`/`onClosing`.

**`LoaderThread`** (the privileged I/O worker) is started by its own constructor (LoaderThread.java:468-471, daemon, priority 10). It opens cache files only when `isWin32` (Microsoft JVM) — **on a standard JDK 17 the cache-file-open block at LoaderThread.java:417-464 is skipped entirely** (`isWin32` from `java.vendor`), and most privileged opcodes (3,5,6,7,12,13,14,15,16,21) throw because of the `if (!isWin32) throw` guard at line 656. Only socket-open (1/22), thread-spawn (2), URL-open (4), and reflection (8/9) work without Win32. This matters: cache reads on a standard JVM go through `DataStore`/`CacheFile` constructed elsewhere (loadGameConfig at Mudclient.java:1150-1155), not through LoaderThread's per-index cache files.

### Applet vs Application

Running standalone does **NOT** need AppletStub plumbing. `startApplication` sets `InputState.gameFrame`, and every applet-context method (`getParameter`, `getCodeBase`, `getDocumentBase`, `getAppletContext`, `getGraphics`, `getSize`, `createImage`, `isDisplayable`) short-circuits to the `GameFrame` (or returns null) when `gameFrame != null` (GameShell.java:230-239, 552-559, 645-652, 799-845, 1039-1047). **rscplus's `Game extends JFrame implements AppletStub` plumbing (Game.java:53, setStub/getCodeBase/getParameter) is needed ONLY because the obfuscated jar is a pure Applet with no standalone path.** The deob has its own `startApplication`, so we replicate **none** of the AppletStub layer — the cleanest standalone entry is just `Mudclient.main()`. Keeping `extends Applet` is fine on JDK 17 (java.applet.Applet still exists, deprecated).

How rscplus boots the obf jar (for contrast, Launcher.java:432-457): `JConfig.create(world)` → `JClassLoader.fetch("/assets/rsclassic-1091943135.jar")` → `loadClass(config.getJarClass())` → `(Applet) client.newInstance()` → `game.setApplet(applet)` (calls `applet.setStub(this)`) → `game.start()` → later `applet.init(); applet.start()`. That is the *applet* lifecycle; the deob sidesteps it.

### What we must ADD / what's BROKEN (blockers)

1. **NAMING-DRIFT BLOCKER (HIGH, the real one).** `Mudclient.main()`/`init()` reference `GameShell.audioQueue`, `GameShell.fontMetrics`, and `super.startLoaderThread(...)` (Mudclient.java:549,626,630,643). **None of these exist on the committed `GameShell.java`** — its standalone entry is named `startApplication(...)`/`startApplet(...)` and it has no `audioQueue`/`fontMetrics` static fields. So `main()` calls `client.a(false, "local.runescape.com", ...)` (the obf name for `startApplication`) while GameShell declares `startApplication`. This is exactly the half-renamed Mudclient↔subsystem mismatch documented in `docs/build/findings/01-deob-compilability.md` (~1,535 obf-name call sites). **Until Mudclient is reconciled against GameShell's renamed members, the boot path does not link.** This is the dominant cost and is being worked separately (the Mudclient reconcile workflow). Our entry design should target GameShell's *renamed* surface: `new Mudclient(); startApplication(false, title, loaderArg, "classic", port, doUpdate, storeFlag, 512, 346)`.

2. **RSA KEY MISMATCH BLOCKER (must ADD an override).** The deob bakes the real **Jagex** 1024-bit public modulus `ca950472ae9765185b...464f4057` (`BitBuffer.RSA_MODULUS`, BitBuffer.java:50-57) with exponent 65537 (`FontBuilder.rsaPublicExponent`, FontBuilder.java:92). The login handshake encrypts with it (`Mudclient.java:1888: rsaBlock.rsaEncrypt(BitBuffer.RSA_MODULUS, -118, FontBuilder.RSA_EXPONENT)`). **The OpenRSC server the cradle targets uses a *different* key** — modulus `7634250561283973106419...095863517` (decimal), exp 65537 (`proto/v235/rsa.go:26`, `DefaultServerRSA`). So a deob client run against the local OpenRSC server will send a credential block the server cannot decrypt → login fails. **We must ADD an RSA-modulus setter / override** (set `BitBuffer.RSA_MODULUS` to the OpenRSC modulus before login), mirroring how rscplus does it via `JConfig.SERVER_RSA_MODULUS`/`SERVER_RSA_EXPONENT` (JConfig.java:39-41, overwritten per-world from `Settings.WORLD_RSA_PUB_KEYS`). The deob has no config indirection — the key is a hardcoded static, so the override is a one-line static assignment we inject before `startGame`/login.

3. **HOST/PORT must be ADDED/overridden.** In standalone mode `getCodeBase()` returns null, so `loadGameConfig` (Mudclient.java:1194-1209) cannot derive the host from the code base. The `main()` path passes `"local.runescape.com"` as title (not host) and `nodeId+7000` as the update port. The actual game host/port the deob uses is computed in `loadGameConfig` from `serverHost`/`portA`/`portB` (e.g. `serverHost="local.runescape.com"`, `portB=43594`). **We must point these at the OpenRSC server: host `localhost`, port `43594`** (default world, per `openrsc/server/default.conf:21`; cabbage=43595, openrsc=43596, dev=43599). Cleanest is to set `serverHost`/`portB` directly or override the codebase-host logic, since there is no JConfig-style world table in the deob.

4. **main() needs args.** `args[0]`=nodeid (int, e.g. 0), `args[1]`=mode ("live"), `args[2..]`= optional "members"/"veterans". A trivial wrapper `main()` or invoking the existing one with `{"0","live"}` suffices once (1) is fixed.

### Server / cache / credentials availability in this environment

- **Server: AVAILABLE.** OpenRSC server is at `/home/free/code/rsc-hacking/openrsc`; default world port **43594** (`server/default.conf:21`), websocket 43494. The cradle defaults to `localhost:43596` (the "openrsc" world). Either works as a login target; pick one and match its RSA key.
- **RSA: AVAILABLE but must be wired.** OpenRSC modulus is captured in `proto/v235/rsa.go:26` (decimal) — use that to override `BitBuffer.RSA_MODULUS`. No `.pem` is checked into `openrsc/server` here (the key is captured in Go, not a file we can read at runtime).
- **Cache/assets: CANDIDATE AVAILABLE.** rscplus bundles `content0..11`, maps, sprites under `/home/free/code/rsc-hacking/rscplus/assets/` — a candidate local cache to feed `DataStore`/`CacheFile` (loadGameConfig builds `Packet.archiveStore = new DataStore(loaderThread.cacheFile, 24, 0)`). The deob's `loadJagex` boot archives (`jagex.jag`, `fonts###.jag`) and content packs must be present locally; `startApplication`'s `CacheUpdater` HTTP pull (to `127.0.0.1:port`) expects a content server — for a local run we should instead point `DataStore` at the bundled assets and pass `doUpdate <= 20` so the HTTP update is skipped (GameShell.java:598).
- **Credentials: AVAILABLE.** The cradle uses a username + `WESTWORLD_PASSWORD` env var against the OpenRSC server (`cmd/cradle/main.go:6-7`); the same OpenRSC account works for the deob login.

### Bottom line for the entry-launcher area
The boot/run machinery EXISTS and is self-contained (`Mudclient.main` → `GameShell.startApplication` → `GameFrame` + `LoaderThread` + `gameThread`/`run()` loop); no AppletStub layer is needed for standalone. We must ADD only: (a) an RSA-modulus override to the OpenRSC key, (b) host/port pinned to the local OpenRSC server (`localhost:43594`), (c) a local-cache wiring (bundled rscplus assets) with the HTTP CacheUpdater disabled (`doUpdate<=20`). The single hard prerequisite outside our control is **blocker #1: the Mudclient↔GameShell name reconciliation** (`audioQueue`/`fontMetrics`/`startLoaderThread` vs `startApplication`), without which the entry path does not compile/link — that is the separately-running Mudclient workflow's job.

Relevant files: `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/GameShell.java` (run:272, startApplication:576, startApplet:714, closeProgram:245, loadJagex:619), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/GameFrame.java` (ctor:150), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/LoaderThread.java` (ctor:332, dispatch:542, isWin32 guard:656), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/InputState.java` (gameFrame:52), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/Mudclient.java` (main:544, init:619, loadGameConfig:1125, RSA login:1888), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/net/BitBuffer.java:50` (baked Jagex RSA modulus), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/ui/FontBuilder.java:92` (exponent), `/home/free/code/rsc-hacking/westworld/proto/v235/rsa.go:26` (OpenRSC server modulus), `/home/free/code/rsc-hacking/openrsc/server/default.conf:21` (port 43594), `/home/free/code/rsc-hacking/rscplus/src/Client/{Launcher.java:432-457 (obf-jar applet boot), JConfig.java:39-41 (RSA override pattern)}`, `/home/free/code/rsc-hacking/rscplus/src/Game/Game.java:53,66,76 (AppletStub layer we do NOT need)`, `/home/free/code/rsc-hacking/rscplus/assets/` (candidate cache).
