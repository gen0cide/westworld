> _Verbatim runtime bring-up research from the deob-runtime-bringup-design workflow (2026-06-01)._

## FINDINGS

**Area: server-login-rsa — how the compiled deob CONNECTS + LOGS IN, and how to point it at a reachable server.**

### 1. The deob login path (what EXISTS)

The connect+login flow is fully present and traced end-to-end:

- **Connect.** `Mudclient.loginUser(int,String,String,boolean)` (`src/client/Mudclient.java:1797`) opens the socket via `this.createSocket(dummy, port, this.serverHost)` at line 1839 and wraps it in `new ClientStream(socket, this)`. The socket itself is produced by `StreamFactory.createSocketFactory(int port, String host)` (`src/client/util/StreamFactory.java:161`) → `new ProxySocketFactory()` with `factory.host/factory.port` set → `ProxySocketFactory.openSocket((byte)50)` (`src/client/net/ProxySocketFactory.java:193`), which falls back to `SocketFactory.openSocketDirect(false)` = `new Socket(host, port)` (`src/client/net/SocketFactory.java:246-250`) when no proxy is configured. Host/port are plain fields: `SocketFactory.host` (obf m.h) and `SocketFactory.port` (obf m.f).
- **RSA login block.** `loginUser` builds opcode-0 LOGIN: marker byte 10, four random 32-bit session-key words, username, padding, then `rsaBlock.rsaEncrypt(BitBuffer.RSA_MODULUS, -118, FontBuilder.rsaPublicExponent)` (call at line 1888). `Buffer.rsaEncrypt(BigInteger modulus, int dummy, BigInteger exponent)` (`src/client/net/Buffer.java:620`) does raw `plain.modPow(exponent, modulus)` then writes a 2-byte length prefix + ciphertext — no padding. This matches OpenRSC's `Crypto.decryptRSA` (raw `modPow`, `openrsc/server/src/.../Crypto.java:88-92`).
- **ISAAC + XTEA.** Same four session-key words seed XTEA over the plaintext tail (UID + password, `Buffer.teaDecrypt`/xteaEncrypt) and then `clientStream.seedIsaac(...)` (`ISAAC`, `src/client/net/ISAAC.java`) for all subsequent opcode scrambling. The deob's ISAAC and XTEA constants (golden-ratio `0x9E3779B9`, 32 rounds) match OpenRSC exactly.

### 2. THE KEY ISSUE — making the RSA modulus/exponent settable (what to ADD)

Both fields are already `public static` and **directly assignable** — no obf indirection like rscplus needed, because in the deob they are real Java fields, not bytecode-patched:

- **`client.net.BitBuffer.RSA_MODULUS`** (`src/client/net/BitBuffer.java:50`, obf `ja.K`) — currently baked to the Jagex 1024-bit key `ca950472...`.
- **`client.ui.FontBuilder.rsaPublicExponent`** (`src/client/ui/FontBuilder.java:92`, obf, = `new BigInteger("10001",16)` = 65537).

To point at a private server, before `loginUser` runs do:
```java
BitBuffer.RSA_MODULUS = new BigInteger("<server modulus, base 10 or 16>");
FontBuilder.rsaPublicExponent = new BigInteger("65537");
```
This is the deob equivalent of rscplus's `JConfig.SERVER_RSA_MODULUS`/`SERVER_RSA_EXPONENT` (rscplus needs `JClassPatcher`/`Client.modulus` bytecode injection because it loads the *obfuscated* jar; the deob is source, so a plain static write — or a tiny `setServerRsa(modulus, exponent)` setter — suffices). Recommend adding a one-line static setter for cleanliness, but it is not strictly required.

### 3. What server we can connect to, and the RSA mismatch (BLOCKER, but solvable)

- **Reachable server:** the same OpenRSC instance the cradle uses — world "openrsc" on **`localhost:43596`** (cradle default `cmd/cradle/main.go:77`; `openrsc/server/default.conf:21` documents the per-world port map, default.conf itself binds `43594`). It speaks the rev-234/235 protocol the deob targets (deob `ClientIOException.BUILD_REVISION = 234`; cradle/OpenRSC = 235; OpenRSC `enforce_custom_client_version: false` so the 1-rev gap is tolerated).
- **RSA key mismatch (the real blocker):** the deob's baked modulus is Jagex's 1024-bit `ca950472...`. The OpenRSC server generates its **own 512-bit** RSA keypair on first boot (`Crypto.generateRSAKeys`, `KeyPairGenerator.initialize(512)`, persisted to `openrsc/server/client.pem`/`server.pem`). The cradle has that server's modulus baked as decimal in `proto/v235/rsa.go DefaultServerRSA()` = `7634250561283973106419144827843935010165327069935723928109242614288318739395804201883596278169185387687268116837066108754542364007806573724086207095863517` (also in `docs/protocol.md:102`). **The deob MUST override `BitBuffer.RSA_MODULUS` with this exact value**, or the server's `modPow`-decrypt yields garbage and login fails. NOTE this modulus is environment-specific: if `client.pem`/`server.pem` are regenerated the value changes; read it live from the server log line `Crypto.getPublicModulus()` (`Server.java:498`) or from `client.pem`.
- **Login-block layout caveat (verify before declaring victory):** the cradle's `proto/v235/login.go BuildRSABlock` puts the **password (20 bytes) inside the RSA block** and the **username in the XTEA tail** (61-byte block, OpenRSC `clientVersion>204` path). The deob's classic `loginUser` does the **opposite**: username in the RSA block, password+UID in the XTEA tail. Both target the same OpenRSC server, so the server's `LoginPacketHandler` must accept the deob's classic ordering. This is the one substantive protocol-compatibility risk for the deob talking to OpenRSC and should be confirmed against `LoginPacketHandler.java` (the cradle proves the >204 layout works; the deob uses the rev-204-style layout — they may diverge). If they diverge, the deob's `loginUser` block construction (not the RSA key) is what needs adjusting.

### 4. Booting the compiled deob (what EXISTS vs what to ADD)

- **Entry point.** `GameShell` (`src/client/shell/GameShell.java:74`) extends `java.applet.Applet`, is `Runnable`, and has a **standalone path** `startApplication(boolean,String,int,String,int,byte,int,int,int)` (line 576) that creates a `GameFrame` window, sets up the `LoaderThread`, and starts the game thread (`run()`, line 272: loads Jagex logo+fonts, then `startGame()`, then the FPS-paced loop calling `handleInputs(119)` and `draw(false)`).
- **MUST ADD a `main()`.** The skeleton lists a `static void main(String[])` for Mudclient (`MUDCLIENT_SKELETON.md` bootstrap row, originally ~line 36183), but it is the applet/browser variant. For a standalone run you need a `main()` that: (a) sets `BitBuffer.RSA_MODULUS` + `FontBuilder.rsaPublicExponent` to the OpenRSC values, (b) sets the world host/port (`serverHost`/`loginPort` = `localhost`/`43596`), (c) constructs the `Mudclient` and calls `startApplication(...)` (or directly drives `startGame`). rscplus's headless harness `rscplus/dumprender/PatchAndRender.java:70-102` shows the reflection-based construction pattern (build Surface→Scene→World, invoke render) if a headless boot is wanted instead of a window.
- **MUST PROVIDE a local cache.** `startApplication` optionally pulls a content pack via `CacheUpdater.downloadAndVerifyCrcs(new URL("http","127.0.0.1",port,""), ...)` (GameShell.java:602); the boot also loads `jagex.jag`/`fonts###.jag` (logo+fonts) and game archives (`integer.dat`, `string.dat`, models, maps, textures) via `StreamBase.loadResource`. Candidate caches exist in `rscplus/assets/` (content0..11, maps, sprites). This is a real dependency to wire up but the assets are available in-repo.

### 5. Environment blockers (CONCRETE)

- **JDK version: BLOCKER for the default toolchain.** The deob compiles cleanly under **`/usr/lib/jvm/java-17-openjdk`** (verified). The environment's default `javac` is **26.0.1**, where `java.applet` was *removed* (JDK 21+); `GameShell`/`Globals`/`JSBridge` import `java.applet.*` and will not compile or run on 21/26. Must pin JDK 17 (java.applet is present-but-deprecated there). There is **no build system** in `reference/rsc-client/` — just `src/`; you compile with `find src -name '*.java' | javac -d out @-`.
- **Mudclient compile state:** as the task warned, `src/client/Mudclient.java` is mid-reconciliation. I verified that under JDK 17 **all 100 compile errors are confined to Mudclient.java** (it still references unresolved obf short names `da`, `ja`, `lb`, `k`, `ca`, …); every other subsystem class compiles. I designed against the committed subsystem APIs and did not depend on Mudclient's compile state. Also noted: the in-progress Mudclient references `BitBuffer.RSA_MODULUS` (correct) but `FontBuilder.RSA_EXPONENT` and `ClientIOException.CLIENT_VERSION` and `this.loginPort` — names that don't match the committed `FontBuilder.rsaPublicExponent` / `ClientIOException.BUILD_REVISION` / `portA`/`portB`; those are the reconciliation's to fix, not mine.

### Minimal path to a successful login
1. Build with **JDK 17** (`/usr/lib/jvm/java-17-openjdk/bin/javac`), full `src/` tree, once Mudclient reconciliation lands.
2. Add a `main()` that overrides `BitBuffer.RSA_MODULUS = new BigInteger("7634250561...3517")` and `FontBuilder.rsaPublicExponent = new BigInteger("65537")`, sets host=`localhost` port=`43596`, and calls `startApplication(...)`.
3. Provide the boot cache (`jagex.jag`, `fonts###.jag`, game archives) — source from `rscplus/assets/`.
4. Start OpenRSC server (world `openrsc`, port 43596); read its live modulus from `client.pem`/startup log in case the keypair was regenerated.
5. Log in with an existing account; password via `WESTWORLD_PASSWORD` env (the cradle convention — `cmd/cradle/main.go:124`; default username `alex`). No in-deob account-creation needed — reuse the cradle's already-provisioned account on that server.
6. **Verify** the deob's classic RSA-block ordering (username-in-RSA / password-in-XTEA) is accepted by OpenRSC's `LoginPacketHandler`; if not, adjust `loginUser`'s block construction to the >204 layout shown in `proto/v235/login.go`.

Relevant absolute paths: `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/net/BitBuffer.java` (RSA_MODULUS), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/ui/FontBuilder.java` (rsaPublicExponent), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/Mudclient.java:1797-1905` (loginUser), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/util/StreamFactory.java:161` (createSocketFactory), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/net/SocketFactory.java:109,129,246` (host/port/openSocketDirect), `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/shell/GameShell.java:272,576` (run/startApplication), `/home/free/code/rsc-hacking/westworld/proto/v235/rsa.go` + `/home/free/code/rsc-hacking/westworld/proto/v235/login.go` (server modulus + login layout), `/home/free/code/rsc-hacking/openrsc/server/src/com/openrsc/server/net/rsc/Crypto.java` (512-bit keygen), `/home/free/code/rsc-hacking/openrsc/server/default.conf:21` (port map).
