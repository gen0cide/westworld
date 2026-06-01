> _Verbatim runtime bring-up research from the deob-runtime-bringup-design workflow (2026-06-01)._

## FINDINGS

### What the deob loads at startup, and from where

**Two-phase boot (exists in deob).** `GameShell.run()` (GameShell.java:272) does a one-time boot: `loadJagex()` (GameShell.java:619) loads the `jagex.jag` boot archive (logo.tga) + 8 bitmap fonts (`h11p`..`h24b` `.jf`), then calls `startGame()` → Mudclient's `init`/`loadGameConfig` group (Mudclient.java:1122+), which loads all game data, then `drawLoginScreen`. The standalone entry `GameShell.startApplication(...)` (GameShell.java:576) already exists and wires up `GameFrame` + `LoaderThread` + game thread — no entry class is missing; only a ~10-line `main()` must be ADDED.

**The asset model is HTTP "content packs", NOT a Jagex disk cache.** This is the single most important finding. There are two code paths and on a standard JVM only the HTTP one is live:

1. **On-disk Jagex cache (`main_file_cache.dat2/.idx*`) — DEAD on a standard JDK.** `LoaderThread` only opens those `CacheFile`s inside `if (this.isWin32)` (LoaderThread.java:417). `isWin32` is set true only when `java.vendor` contains "microsoft" (LoaderThread.java:360). On OpenJDK it is false, so `dataFile`/`indexFile255` are never opened. Consequently `CacheUpdater.downloadAndVerifyCrcs` finds `ImageLoader.imageWidthCarrier.dataFile == null` (CacheUpdater.java:239) and leaves `SocketFactory.globalArchive == null`. The `ArchiveReader`/`DataStore`/`CacheFile`/`CachePath` 520-byte-sector machinery is therefore irrelevant for bring-up — we do NOT need a Jagex disk cache.

2. **HTTP content download — the live path.** Every game-data load goes through `StreamBase.loadResource(-101, name, pct, contentIndex)` (StreamBase.java:224). It: (a) returns `ClientRuntimeException.byteRowScratch[contentIndex]` if cached; (b) since `globalArchive==null`, skips the archive; (c) builds the URL **`BASE_URL + "content" + contentIndex + "_" + Long.toHexString(Buffer._junkArray[contentIndex])`** and downloads it (3 retries) via `ClientStream.downloadFile` → `DownloadWorker`; (d) verifies CRC against `Buffer._junkArray[contentIndex]`; (e) decompresses via `World.unpackData`. `BASE_URL` is set from the applet code-base / `startApplication` host (CacheUpdater.java:183).

**The CRC table comes from `contentcrcs`.** `CacheUpdater.downloadAndVerifyCrcs` (CacheUpdater.java:173, called from both `startApplet` and `startApplication`) GETs `BASE_URL + "contentcrcs" + hex(millis)`, reads 12 big-endian ints into `Buffer._junkArray[0..11]`, then a 13th (trailing checksum) and verifies via `Buffer.verifyCrc`.

### The local asset source IS available and matches the deob

`/home/free/code/rsc-hacking/rscplus/assets/content/` contains exactly `content0_229aa476` … `content11_7d5437c5` plus `contentcrcs`. I hexdumped `contentcrcs`: it is 13 big-endian 4-byte ints (`229aa476 1c9fa8c3 2fdddb3c … 7d5437c5 61a8182d 00000000 0000`) — the 12 content CRCs match the filename hex suffixes one-for-one (e.g. `content0_229aa476` ↔ first CRC `0x229aa476`). **This is precisely the rev-235 content-pack layout and CRC scheme the deob's `loadResource`/`downloadAndVerifyCrcs` expect.** These are the rev-233-235 (J++ re-release) assets, the same revision as the deob.

**Content-file index → archive mapping** (the 3rd arg to the `fetchAsset` = `Mudclient.a(String,int,int,int)` wrapper, which calls `loadResource`):
- `0` = **Configuration/GameData** (`drawOptionsTab(false)`, Mudclient.java:9155 `a(STRINGS[225],10,0,78)`) → `SocketFactory.initGameData` allocates ALL item/NPC/entity/animation defs. **Required for login screen.**
- `1` = entity sprites (`a.dat`, people), `2` = members entity sprites (`f.dat`)
- `4` = `map`, `5` = `members map`, `6` = `land`(scape), `7` = `members landscape`
- `8` = **2d graphics** (`index.dat` + `inv1/inv2/bubble/…` UI sprites — needed to draw login/UI)
- `9` = **3d models** (`.ob2`/`.ob3`)
- `10` = `Sounds` (members only), `11` = `Textures`

`.jag` archives inside each content file are unpacked by `EntityDef.extractArchiveEntry` (EntityDef.java:285): name-hash `hash*61+(ch-' ')` over the upper-cased filename, BZip2-decompressed.

### Minimal asset set
- **(a) Login screen:** boot fonts/logo (`jagex.jag` + 8 `.jf` — present in `rscplus/assets/jf/`, but note the deob's `loadJagex` reads them via `StreamBase.loadResource`/`readDataFile` as a `jagex.jag` archive, so they must be packaged as that archive or served as content), `contentcrcs`, content index **0** (GameData) and **8** (2d graphics). That reaches `drawLoginScreen`.
- **(b) Render a scene:** additionally content **9** (models), **11** (textures), **6**/**4** (landscape+map), **1** (entity sprites). All present in `rscplus/assets/content/`.

### What must be ADDED (the deob lacks these)
1. **A `main()` launcher** calling `new Mudclient().startApplication(...)` (or the `startApplet` path). ~10 lines. The skeleton's `main` (Mudclient.java:36183 region) reads applet-style mode args; a standalone shim is simplest.
2. **A content HTTP server** at `BASE_URL` serving `contentcrcs` and `contentN_<hex>` from `rscplus/assets/content/`. Trivial: a static file server rooted there. `BASE_URL` must be set to that server (via the code-base in `startApplet`, or the `startApplication` host:port at GameShell.java:602 which is hard-coded to `http://127.0.0.1:<port>`). **No download CONTENT server beyond a static file host is needed**, and no Jagex disk cache.
3. **An RSA modulus override.** `Mudclient` login (Mudclient.java:1888) RSA-encrypts the login block with `BitBuffer.RSA_MODULUS` (BitBuffer.java:50, the baked Jagex 1024-bit key `ca950472…`) and `FontBuilder.RSA_EXPONENT`. To talk to the local OpenRSC server this MUST be replaced with the server's key.
4. **Host/port override.** `loadGameConfig` (Mudclient.java:1194-1209) derives `serverHost` from the applet code-base host and port `43594`/`40000+nodeId`. The local OpenRSC/cradle server listens on **`localhost:43596`** (cmd/cradle/main.go:77). A standalone launcher must force host=`127.0.0.1`, port=`43596`.

### Server / credential availability
- **Login protocol MATCHES.** The deob's login (Mudclient.java:1882-1903) is the rev-235 **RSA-block (marker 10, 4 ISAAC/session keys, username, padding) + XTEA-encrypted tail (limit30, 24-byte UID record, password)** scheme, then `seedIsaac`. This is exactly what OpenRSC's Payload235 server and `westworld/proto/v235/login.go` (BuildRSABlock 61 bytes / BuildXTEABlock) implement. So the wire format is compatible — only the RSA key differs.
- **RSA key — the one real credential blocker, but resolvable.** OpenRSC `Crypto.java` (openrsc/server/.../rsc/Crypto.java:97) **generates a 512-bit keypair** into `client.pem`/`server.pem` on first run (it is NOT a fixed Jagex key). `westworld/proto/v235/rsa.go:26` records the running server's modulus as `7634250561283973106419144827843935010165327069935723928109242614288318739395804201883596278169185387687268116837066108754542364007806573724086207095863517` with exponent `65537`. The deob must bake THIS modulus (not the Jagex `ca950472…`). NOTE: this is 512-bit; the deob block + the proto note both use raw RSA (no padding), and the 61-byte plaintext fits. If the server's `.pem` is regenerated this value changes (rsa.go warns of this) — so the override should read the live server key.
- **A server + account exist.** OpenRSC server (`localhost:43596`) and the cradle already log in (cradle default `-server localhost:43596 -username … WESTWORLD_PASSWORD`). The same host/account work for the deob once RSA + host/port are overridden.

### rscplus precedent (for the overrides)
rscplus does exactly these overrides against the obf jar: `JConfig.SERVER_RSA_MODULUS`/`SERVER_RSA_EXPONENT` (rscplus/src/Client/JConfig.java:39-41, baked Jagex `8919358150…` by default but overridable per-world) and `Settings.WORLD_RSA_PUB_KEYS`/`WORLD_PORTS`/`WORLD_URLS` per world. For the compiled deob we replicate this by setting `BitBuffer.RSA_MODULUS`/`FontBuilder.RSA_EXPONENT` and the host/port directly (we own the source — no ASM needed).

### Blockers (honest)
- **BLOCKER (soft):** The whole client only runs once `Mudclient.java` compiles — that reconciliation is the separate in-progress workflow (findings/01 estimates it as the dominant cost: ~1,535 obf-name call sites). The boot/asset design here is sound and ready, but cannot be exercised until Mudclient links.
- **BLOCKER (must-fix-by-edit):** `BitBuffer.RSA_MODULUS` is the Jagex key; without overriding it to the OpenRSC server's `7634250561…` modulus, the server cannot decrypt the login block (login fails silently). This is a one-line source edit (or a setter we ADD).
- **Packaging detail:** the boot fonts/logo are loaded via `loadJagex`→`StreamBase.loadResource`/`readDataFile` as a `jagex.jag` archive; the bare `.jf` files in `rscplus/assets/jf/` must be packaged into the archive form the deob reads, or that boot path adapted. Minor but real.
- **No blocker on assets:** `rscplus/assets/content/` IS the matching rev-235 content+CRC set; serve it statically and the deob's HTTP content loads succeed without any Jagex disk cache.

Relevant files: `reference/rsc-client/src/client/net/StreamBase.java:224` (loadResource), `reference/rsc-client/src/client/data/CacheUpdater.java:173` (contentcrcs), `reference/rsc-client/src/client/shell/LoaderThread.java:417` (Win32-gated disk cache), `reference/rsc-client/src/client/net/BitBuffer.java:50` (baked RSA modulus), `reference/rsc-client/src/client/Mudclient.java:1188-1209` (host/port) and `:1882-1903` (login block), `reference/rsc-client/src/client/data/EntityDef.java:285` (.jag format); assets `rscplus/assets/content/contentcrcs` + `content0..11`; server `openrsc/server/src/com/openrsc/server/net/rsc/Crypto.java:97`, `westworld/proto/v235/rsa.go:26` (server modulus), `westworld/cmd/cradle/main.go:77` (localhost:43596).
