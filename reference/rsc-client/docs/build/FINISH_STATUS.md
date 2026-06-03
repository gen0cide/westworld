# Finish status — audit snapshot

_Date: 2026-06-01. JDK 17 verified (openjdk 17.0.19, javac 17.0.19). Read-only audit; no commits._

## TL;DR

- **Compile state is MUCH worse than the "~100 residual errors" the task brief assumed.** That "~100" was just javac's default *display cap*. The true count is **2515 errors** (`-Xmaxerrs 3000`), of which **2504 are in `Mudclient.java`** and **2443 are `cannot find symbol`**.
- Root cause is concentrated and mechanical, not 2500 independent bugs: the assemble step left **obfuscated identifiers un-renamed** in `Mudclient.java`. Two dominant patterns account for the vast majority:
  1. The static string pool is **declared as `il`** (`private static final String[] il = new String[660]` at line 99) but **referenced 785× as `STRINGS[...]`** → **~754 errors** from one missing rename.
  2. **21 field declarations (lines 104–199) still use obfuscated TYPE tokens** (`da ja lb k ca ba ta qa wb ra sa`) that are not class names in the deob tree → those fields are untyped, which cascades into **most of the 2483 method-body errors** (`int cannot be dereferenced`, `array required but int found`, undefined `this.<field>` reads).
- **Milestone position: M0 (tree compiles) is UNMET — every other milestone is gated on it.** Server, assets, RSA override are all otherwise ready.
- **The fix is small in *kinds* even though it is large in *count*:** resolve ~11 obf class-tokens to their deob class names, rename `il`→`STRINGS` (or add an alias), then re-run and iterate on the residual member-name drift.

---

## 1. Git state

- **Current branch:** `feat/remote-client` (NOTE: the task brief said `deob/rsc-client-reference`; HEAD is actually on `feat/remote-client`. `deob/rsc-client-reference` exists as a separate local branch and is behind.)
- **Last 8 commits (committed work):**
  - `a3f3d82` docs(runtime): runtime prep READY — OpenRSC server live + boot assets staged
  - `123f9c5` docs(runtime): deob client runtime bring-up plan + milestone ladder
  - `28aa079` docs(rebase): render reconciliation — ours vs orsc vs Jagex (ground truth)
  - `82ff10e` feat(deob): subsystem packages compile clean (0 errors) — member-drift reconciliation
  - `4668b76` docs(rebase): upstream render-refactor analysis + rebase plan
  - `a8ab7d4` feat(render): port World.addModels — diagonal scenery objects/doors now build geometry
  - `33e7226` fix(deob): apply 5 net-handler divergences to Mudclient packet dispatch
  - `276b58a` feat(render): fix 7 Go render-fidelity bugs + add JAR oracle engine + render-gap audit

### Uncommitted working-tree changes — ownership split

**OURS (deob-client, in-progress — the assembled reconcile):**
- `reference/rsc-client/src/client/Mudclient.java` (modified) — the freshly assembled 11,650-line Mudclient; this is the file with the 2504 residual errors.

**NOT OURS — separate in-progress work by someone else (DO NOT TOUCH, do not commit):**
- `cmd/cradle/main.go` (modified)
- `proto/v235/buffer.go` (modified)
- `render/pick.go` (modified)
- `render/render.go` (modified)
- `cmd/cradle/clientpage.go` (untracked)
- `cmd/cradle/remoteclient.go` (untracked)
- `docs/remote-client/` (untracked dir)
- `remoteclient/` (untracked dir)
- `render/hittest.go`, `render/hittest_test.go` (untracked)

Only `Mudclient.java` is in our lane. Everything under `cmd/cradle`, `render/`, `proto/`, `remoteclient/`, `docs/remote-client/` is the Go cradle/render tree — out of scope, hands off.

---

## 2. Compile state (full-deob javac, JDK 17)

Command (corrected — `javac @-` does not read stdin; used a real arg-file):
```
find reference/rsc-client/src/client -name '*.java' > /tmp/deob-sources.txt   # 68 files
javac -Xmaxerrs 3000 -d /tmp/deob-full @/tmp/deob-sources.txt
```

**Result: 2515 errors, 7 warnings.** (Default `javac` caps the display at 100 — that is the "~100" in the brief; the true total is 2515.)

### Errors by file
| count | file |
|------:|------|
| 2504 | `Mudclient.java` |
| 4 | `scene/SurfaceSprite.java` |
| 3 | `util/Timer.java` |
| 3 | `shell/GameShell.java` |
| 2 | `net/SocketFactory.java` |
| 1 each | `world/World.java`, `util/JSBridge.java`, `util/Globals.java`, `scene/SpriteScaler.java`, `net/ProxySocketFactory.java`, `net/ClientStream.java` |

The ~11 non-Mudclient errors are back-references *into* Mudclient (e.g. `ze is not public in Mudclient`, `Jk is not public in Mudclient`, `x is not public in Component`) — they resolve as a side effect of fixing Mudclient's visibility/member names; they are not independent.

### Errors by kind
| count | kind |
|------:|------|
| 2443 | `cannot find symbol` |
| 33 | `int cannot be dereferenced` |
| 10 | `array required, but int found` |
| 5 | `method does not override or implement a method from a supertype` |
| 4 | `incompatible types: int cannot be converted to boolean` |
| 3 | `... is not public in Mudclient/Component` |
| 3 | `incompatible types: String → int` |
| 3 | `incompatible types: lossy int → byte` |
| 3 | `incompatible types: boolean → int` |
| ~5 | method-applicability (`drawScrollList`, `drawScrollbarN`, `loadEntitySprites`) |

### `cannot find symbol` broken down (2443)
- **2319 variable**, **68 class**, **56 method.**

### Line distribution within Mudclient.java
- **Field-decl region (lines 104–199): 21 errors** — all are the obf-TYPE-token field declarations.
- **Method-body region (>200): 2483 errors** — overwhelmingly downstream of (a) the unresolved field types and (b) the `il`/`STRINGS` rename.
- (min line 104, max line 11548 — spread across the whole file, as expected for field-type cascades.)

---

## Categorized work-list for the Compile phase

### A. ONE rename clears ~754 errors — the string pool
- Line 99 declares: `private static final String[] il = new String[660]; // obf: il — STRINGS: XOR-encrypted string pool`
- The body references it **785×** as `STRINGS[...]` (top missing variable = `STRINGS`, 754 hits across 584 distinct lines).
- **Fix:** rename the field `il` → `STRINGS` (or declare `STRINGS` as an alias of `il`). This is the single highest-leverage fix in the file.

### B. ~11 obf class-tokens → deob class names (clears the 21 field-decl errors + most method-body cascades)
The field declarations at lines 104–199 still carry obfuscated type tokens. Mapping (confirmed from `/tmp/mud/mud_map.tsv` role annotations + the deob class set in `src/client`):

| obf token | deob class | role (from field comment) |
|-----------|-----------|---------------------------|
| `da` | `ClientStream` | clientStream (outgoing packet stream) |
| `ja` | `BitBuffer` | incomingPacket (inbound bit-buffer) |
| `lb` | `World` *(per deob field comment)* | world — **see inversion note below** |
| `k`  | `Scene` *(per deob field comment)* | scene — **see inversion note below** |
| `ca` | `GameModel` | wall/npc model arrays (`ca[]`) |
| `ba` | `Surface` | surface (2D blitter) |
| `ta` | `GameCharacter` | players/npcs (knownPlayers, npcsCache, localPlayer, …) |
| `qa` | `Panel` | panelQuest / panelGameAlt / panelDuel / panelLogin |
| `wb` | `MessageList` | friendsList / chatList (also reused as menu builder) |
| `ra` | `AudioMixer` | soundMixer |
| `sa` | `AudioChannel` | soundChannel |

**⚠ World↔Scene (Ek/Hh) inversion — load-bearing, documented in `RENDER_DEOB_GAPS.md` (rows at lines 43–49):** the deob field comments label `Ek`(lb)=World and `Hh`(k)=Scene, but **ground-truth `client.java` has them inverted: `lb`=Scene (owns camera+render) and `k`=World (owns terrain/elevation).** Resolve every render-method member call **by the receiver's DECLARED TYPE, not its name** — `scene.*` calls must route to Scene's API and `world.*` to World's. Do not trust the `scene`/`world` aliases at face value when wiring members.

The 68 missing-class errors and the bulk of the 2319 missing-variable errors (`this.<field>` reads that fail because the field's declared type didn't resolve) collapse once these declarations type-check.

### C. Residual member-name drift (the genuine reconcile remainder, ~50–60 after A+B)
Once fields resolve, real obf→deob member drift surfaces. Top missing names still expected:
- **Missing methods (56):** `a` (12 — misrouted; e.g. `ISAAC.a(...)` should be `Surface.rgb2long`), `normaliseName` (4), `drawMinimapEntity` (4), `resetMenuState` (2), plus singletons: `sendChatMessage`, `initShopPanel`, `hasInventoryItems`, `hasInboundData`, `getInventoryCount`, `drawTeleportBubble`, `drawPlayer`, `drawNpc`, `drawItem`, `clampXp`.
- **Misrouted calls flagged as applicability errors:** `drawScrollList` (2), `drawScrollbarN` (1), `loadEntitySprites` (1) — wrong arg counts/types → resolve against the real signatures.
- **Method-named missing variables** that are actually un-renamed fields: `panelQuest` (52), `GameData` (23), `mouseClickButton` (22), `cameraAngle` (17), `spriteBaseInventory` (16), `fatalLoadError` (16), `mouseLastButton` (15), `socialDialogMode`/`panelMagic`/`isMembersWorld`/`inputDialogType` (13 each), and a tail of obf short-name fields (`ua`,`fa`,`db`,`f`,`n`,`l`,`ia`,`uc`,…). Resolve via **`/tmp/mud/mud_map.tsv`** (Mudclient's own obf→deob field/method map, 535 rows).

### D. Structural / type errors (the ~50 non-symbol errors — mostly symptoms of A+B)
- `int cannot be dereferenced` (33) at L5347/5355/5558–5590/6049–6059/7682/7714/7975 — a field currently typed `int` (unresolved) being used as an object; clears when its real object type is restored.
- `array required, but int found` (10) at L3462–3528 — same root: an array field left as `int`.
- `int↔boolean` (4) at L2192–2194, L3611 — obf comparison idioms (`x != 0` flattened); restore boolean semantics.
- `method does not override...` (5) at L1042/1124/1655 + 2 — applet/AWT override signatures drifted (the `x is not public in Component` at L5472/5819/6941 + `String→int` are the same three call sites — `Component.x`/layout drift).
- `lossy int→byte` (3), `String→int` (3) — narrow once the surrounding types resolve.

**Suggested order:** A (STRINGS rename) → B (field types + World/Scene inversion) → re-run javac → C (member drift via `mud_map.tsv` + `OBF_MEMBER_MAP.tsv` for subsystem receivers) → D (mop up the structural stragglers). Expect the count to drop from 2515 to low-hundreds after A+B, then to zero through C+D.

### Resolution maps available
- `reference/rsc-client/docs/build/OBF_MEMBER_MAP.tsv` (48 KB) — subsystem-class members; use for calls on subsystem receivers (surface/world/scene/clientStream/…), keyed by (deob_class, obf_member, arg_count).
- `/tmp/mud/mud_map.tsv` (34 KB, 535 rows) — Mudclient's OWN obf→deob fields/methods; use for `this.<obf>` references in Mudclient method bodies.
- Ground truth: `reference/rsc-client/decompiled/normalized-clean/client.java` (obf class `client`).
- `reference/rsc-client/docs/build/RENDER_DEOB_GAPS.md` — the 5 render bugs + the World/Scene inversion.

---

## 3. Milestone position (M0..M5, per RUNTIME_BRINGUP_PLAN.md)

**We are at M0 — "Tree compiles to bytecode" — and M0 is UNMET.** Every milestone above is explicitly gated on M0.

| Milestone | State | Notes |
|-----------|-------|-------|
| **M0 — tree compiles** | ❌ **UNMET — the immediate gate** | 2515 javac errors, all confined to `Mudclient.java` (+11 back-refs). Subsystem packages otherwise compiled clean as of commit `82ff10e`. |
| M1 — boots + opens window | ⛔ gated on M0 | Boot path exists: `Mudclient.main()` (L544) → `GameShell.startApplication` → GameFrame + LoaderThread. Xvfb ready (`/usr/bin/Xvfb`). JDK 17 ready. No AppletStub layer needed. |
| M2 — title screen + cache | ⛔ gated on M0 | Reachable with server down; content staged (below). |
| M3 — connects + logs in | ⛔ gated on M0 + server | Server up; RSA override documented (BitBuffer.RSA_MODULUS + FontBuilder.rsaPublicExponent, both public static); in-protocol register-then-login; 127.0.0.1 bypasses rate limits. |
| M4 — live world frame | ⛔ gated on M3 | Watch render drift sites (`scene/GameModel.java:302/304/346`) + the World/Scene inversion. |
| M5 — playable | ⛔ gated on M4 | Input loop + opcode-handler correctness unmeasured. |

**Immediate gate:** finish the `Mudclient.java` member/type reconcile to reach a clean compile (M0). Nothing downstream can be exercised until then.

---

## 4. Server / assets readiness

- **OpenRSC server: UP.** TCP connect to `127.0.0.1:43594` succeeds (OPEN). Java server processes (ant-launched) are running. Logs at `/tmp/rsc-run/server-run2.log`. Restart if needed: `/tmp/rsc-run/start-server.sh` (background), poll the port.
- **Content cache: STAGED.** `/tmp/rsc-run/cache` holds `content0..content11` packs + `contentcrcs` + `jagex.jag` (content index 3, the boot logo/fonts archive). content9 (models), content11 (textures), content6/4 (landscape+map), content1 (entity sprites) all present.
- **Content host:** `/tmp/rsc-run/content-host.py` serves the cache on **port 43594 by default** (same port as the game server — start it on a different port, e.g. `python3 content-host.py 8000`, OR pre-stage the cache and pass `doUpdate<=20` to skip the HTTP CacheUpdater entirely, which is the simpler path). `start-content-host.sh` wrapper present.
- **RSA override (ready, apply before login):** `BitBuffer.RSA_MODULUS = new BigInteger("8470727801174954902989859055344934434282083179399207801708507751976321325965228952554034824402302678046886295251980280826867546707365065713308009848924031")` and `FontBuilder.rsaPublicExponent = BigInteger.valueOf(65537)`.
- **Display:** Xvfb at `/usr/bin/Xvfb` (e.g. `Xvfb :99 -screen 0 1024x768x24`).

**Net:** runtime infra (server, assets, RSA, display) is READY. The sole blocker on the whole ladder is the M0 compile of `Mudclient.java`.
