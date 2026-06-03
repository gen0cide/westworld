> _Verbatim investigation report from the `rsc-client-build-options` workflow (2026-06-01). One of five parallel agents; see `../OPTIONS.md` for the synthesis._

## A2 — rscplus hook/reflection surface & re-pointing onto deob names

### Scope and method

I read `JClassPatcher.java` (5,858 lines) and `Reflection.java` (963 lines) in full, parsed every `node.name.equals(...)` dispatch, every `hook*Variable*` mirror, and every obfuscated-owner `FieldInsnNode`/`MethodInsnNode`/`TypeInsnNode` emitted by the instruction rewrites, then cross-referenced each obfuscated identifier against `docs/NAMING.md` and `docs/MUDCLIENT_SKELETON.md`. All counts below are mechanically extracted, not eyeballed.

---

### 1. Per-class patch routines (the `patch()` dispatch)

`patch(byte[])` reads the class with `ClassReader.SKIP_DEBUG`, dispatches on `node.name`, then always runs `patchGeneric`, and finally re-emits with `ClassWriter.COMPUTE_MAXS` (note: **COMPUTE_MAXS only, not COMPUTE_FRAMES** — relevant to re-pointing risk, see §4).

| Routine | Obf target (`node.name`) | NAMING name | What it hooks |
|---|---|---|---|
| `patchRenderer` | `ua` | **Surface** (`client.scene`) | `present(Image)` blit hook (also gated by `Replay.isSeeking`); `@ran@` color override (`RanOverrideEffect`); custom `@nnn@` color-parse rewrite in `drawString a(IILjava/lang/String;IIBI)V`; `getClearColor()` rewrite in `a(Z)V`. |
| `patchApplet` | `e` | **GameShell** (`client.shell`) | Rewrites `addMouseListener/addMouseMotionListener/addKeyListener` to `PUTSTATIC Game/MouseHandler/KeyboardHandler.listener_*`; FPS calc into `Game/Client.fps` (reads `e.Ib:I`); removes throwable crash + 5s shutdown sleep; `ScaledWindow.hookLoadingGraphics` → `PUTFIELD e.u:Ljava/awt/Graphics;`; `error_game_hook`; Java 9+ `isDisplayable`/`isRightMouseButton` fixes; Jagex-font toggle in `b(B)Z` (writes `PUTFIELD e.C` guard, calls `GameApplet.loadJfFonts`). |
| `patchMenu` | `qa` | **Panel** (`client.ui`) | `switchList(Object)` gate in `e(II)V`; register-panel "extra draw" hook in `a(B)V` reading `GETFIELD qa.U:[I`, calling `Game/Panel.draw_extra_hook`. |
| `patchData` | **`m`** | **SocketFactory** per NAMING — **CONFLICT** (see §3) | In `m.a([BBZ)V` appends `Item.patchItemNames`, `Item.patchItemCommands`, `WorldMapWindow.initScenery`, `WorldMapWindow.initBoundaries`. This is the game-config/data-decode method, not a socket factory. |
| `patchClient` | `client` | **Mudclient** | Largest routine (≈lines 1336–4876). ZZZ-sleep scaling; end-of-days array resize 50→200; URL-check removal; `initCreateExtraPanelsHook`; `handlePacket a(III)V` → `Replay.patchClient/saveEncOpcode/checkPoint`; logout/replay hooks (`Replay.closeReplayPlayback/disconnect_hook/paused`); combat-menu show/hide (`Renderer.combat_menu_shown`); chat-history clear skip; bank-drawn flag; XTEA key dump + `login_attempt_hook`; `gameInputHook`, `skipActionReportAbuseTabHook`, `processChatCommand`, `processPrivateCommand`, `mouse_action_hook`; **drawNPC / drawPlayer / drawItem** hooks reading `ta`/`client` fields; right-click bounds fixes; combat-style packet emit via `da.b`/`ja.c`. |
| `patchRandom` | `f` | **RecordLoader** (`client.data`) | In `f.a(ILtb;)V` replaces a validation path with `new java.util.Random().nextBytes(...)`. (NAMING `f`=RecordLoader; Reflection also loads `f` for `PUTRANDOM f.a(int,tb)`.) |
| `patchGameApplet` | `da` | **ClientStream** (`client.net`) | `cacheURLHook` in `a(Ljava/net/URL;ZZ)[B`; `disconnect_hook` in `a(Z)V`; `Replay.dumpRawInputStream` in `a([BIII)V`; `Replay.dumpRawOutputStream` in `run()` reading `da.Y:[B` and field `Q:Ljava/io/OutputStream;`. (Routine is mis-named "GameApplet" but targets the network stream class.) |
| `patchRendererHelper` | `lb` | **Scene** (`client.scene`) | "Renderbug" fix in `b(IZ)V` reading `lb.D`, `k.e:J` (World static long), `jb.o:I`, writing `lb.D:[J`; crash fix in `c(I)V`; camera `offset_height` add in `a(IIIIIIII)V` (field `o`); `getFogColor` rewrite in `c(I)V`; `Camera.postSetCamera()` after `setCamera a(IIIIIIII)V`. |
| `patchSoundHelper` | `sa` | **AudioChannel** (`client.audio`) | `getAndSetSoundGameContainer` in `a(Lc;Ljava/awt/Component;II)Lsa;` (keeps a stable Component across reloads). |
| `patchRightClickMenu` | `wb` | **MessageList** (`client.ui`) | Item-highlight rendering in `a(IIIIIZ)I` reading `wb.b:[Lt;`, `t.o/t.p:String`, `wb.t:Lba;`, `wb.i:I`, calling `ba.a(Ljava/lang/String;IIIZI)V`, `Renderer.getHighlightColour`, `Client.calcStringLength`; right-click bounds via `wb.D/wb.I` → `Renderer.setRightClickMenuBounds`. |
| `patchSoundPlayerJava` | `pb` | **SourceLinePlayer** (`client.audio`) | SFX-delay fix in `b(I)V` reading `pb.y:Ljavax/sound/sampled/AudioFormat;`, rewriting `SourceDataLine.open`. |
| `patchGeneric` | **all classes** | — | `u.a(IJ)V`→`Client.shadowSleep`; `PrintStream.println`→`Logger.Game`; every `ATHROW`→`Client.HandleException`; **and the entire field-mirror block** (`hook*Variable*`). |

`patchTracer` (targets `client`/`lb`) is dev-only and commented out.

**Quantities (patcher side):**
- **Distinct obfuscated classes targeted by `node.name` dispatch: 11** (`ua, e, qa, m, client, f, da, lb, sa, wb, pb`).
- **Field-mirror hooks (`hookClassVariable` 129, `hookConditionalClassVariable` 10, `hookStaticVariable` 30, `hookStaticVariableClone` 2 = 171 calls)** referencing **166 distinct (owner,field) pairs across 24 obfuscated classes**: `client`(116), `e`(8), `lb`(7), `ua`(7), `ba`(3), then `ac, fa, ia, l, n`(2 each), and singletons `b, cb, d, db, fb, h, ja, ka, kb, m, pb, qb, s, u`.
- **Instruction-rewrite obfuscated field refs: 57 distinct (owner,field,desc)** across `client, da, e, jb, k, lb, pb, qa, t, ta, ua, wb`.
- **Instruction-rewrite obfuscated method refs: 7** — `ba.a(II[BI[B)V`, `ba.a(Ljava/lang/String;IIIZI)V`, `client.a(ZLjava/lang/String;…)V`, `da.b(I)V`, `da.b(II)V`, `ja.c(II)V`, `tb.b(II)V`.

Combined, the patcher names roughly **27 distinct obfuscated classes** and **~230 distinct obfuscated members**.

---

### 2. `Reflection.java` inventory

`Reflection.Load()` resolves everything **by `classLoader.loadClass("<obf>")`** (hardcoded obf class names), then:
- **Methods/constructors:** matched by **`method.toGenericString().equals(<full descriptor const>)`** — i.e. the full Java-source signature including modifiers, obf return type, obf owner, and obf param types (e.g. `"private final ta client.b(int,byte)"`). It walks superclasses for the `da`/`ja` families. This is **not** a structural/descriptor scan — it is an exact string compare against the obfuscated names baked into the constants.
- **Fields:** mostly **hardcoded `getDeclaredField("<obf>")`** (e.g. `client.getDeclaredField("Jh")`, all 15 `ta` fields, all 19 `qa` menu fields); a few via `field.getName().equals("<obf>")` loops (`b.f`, `b.d`, `tb.w`, `tb.F`, `kb.a`, `k.m`, `k.I`).

**Classes loaded (16):** `e, na, ca, client, da, b, ja, tb, f, lb, ua, ta, qa, wb, kb, k`.
**Constructors (3):** `qa(ua,int)`, `da(Socket,e)`, `tb(int)`.
**Fields (48 declared static `Field` slots; ~42 distinct `getDeclaredField` names + 7 `getName().equals`):** on `client` (`Jh, Uh, hk, sb`[superclass], `hg, Qg, zg, Lf, sh, Ki, sk, bc, Ub, cl, Fe`), `ta` (`C, b, c, s, i, K, u, B, G, z, h, J`), `qa` (`w, J, tb, kb, B, j, ob, O, sb, U, g, D, vb, Y, k, yb, eb, pb`), `tb` (`w, F`), `b` (`f, d`), `kb` (`a`), `k` (`m, I`), `ua` (`i`), `da` (`Y`).
**Methods (74):** spread over `e`(3), `na`(1), `ca`(2), `client`(20), `da`(3), `b`(4), `ja`(12), `f`(1), `lb`(1), `ua`(13), `qa`(11), `wb`(1).

How it finds them: **purely by hardcoded obfuscated name + exact descriptor string.** There is no fuzzy matching — change any obf name or any param type and the match silently returns `null` (the code tolerates null fields via `if (x != null) x.setAccessible(true)`, so failures are **silent**, not loud).

---

### 3. Obf → deob coverage table & gaps

NAMING.md verifies **all 71 obfuscated classes**. Every obfuscated class rscplus depends on therefore has a readable name. Class-level coverage:

| obf | NAMING deob name | Used by rscplus as | In NAMING? |
|---|---|---|---|
| `client` | Mudclient | patchClient + 116 field mirrors + reflection | ✅ |
| `e` | GameShell | patchApplet, reflection | ✅ |
| `ua` | Surface | patchRenderer, reflection | ✅ |
| `qa` | Panel | patchMenu, reflection (19 fields) | ✅ |
| `da` | ClientStream | patchGameApplet, reflection | ✅ |
| `lb` | **Scene** | patchRendererHelper, Camera mirrors | ✅ (post-correction) |
| `k` | **World** | `k.e:J` renderbug, reflection `m/I` | ✅ (post-correction) |
| `wb` | MessageList | patchRightClickMenu, reflection | ✅ |
| `sa` | AudioChannel | patchSoundHelper | ✅ |
| `pb` | SourceLinePlayer | patchSoundPlayerJava | ✅ |
| `f` | RecordLoader | patchRandom, reflection PUTRANDOM | ✅ |
| `m` | **SocketFactory** | **patchData (`m.a([BBZ)V` data decode)** | ⚠️ **CONFLICT** |
| `na` | StreamFactory | reflection LOAD_DATA | ✅ |
| `ca` | GameModel | reflection (rotate/setLight) | ✅ |
| `ta` | GameCharacter | drawNPC/drawPlayer rewrites, reflection (15 fields) | ✅ |
| `b` | Packet | reflection (newPacket/sendPacket/…/formatText) | ✅ |
| `ja` | BitBuffer | reflection (12 buffer methods), `ja.c(II)V` | ✅ |
| `tb` | Buffer | reflection (ctor + offset/array), `tb.b(II)V` | ✅ |
| `kb` | InputState | reflection `gameReference = kb.a` | ✅ |
| `ba` | SurfaceSprite | mirrors + `ba.a(...)` rewrites | ✅ |
| `t` | EntityDef | `t.o/t.p:String` (item-highlight) | ✅ |
| `jb` | DownloadWorker | `jb.o:I` (renderbug) | ✅ |
| `u` | StringCodec | `u.a(IJ)V`→shadowSleep | ✅ |
| `h` | TextEncoder | `h.c:[I` itemSpriteMasks | ✅ |
| `n` | FontWidths | `n.a/n.g` mirrors | ✅ |
| `l` | Globals | `l.a/l.c` mirrors | ✅ |
| `d` | CacheFile | `d.l` maxRetries | ✅ |
| `s` | FontBuilder | `s.c` exponent | ✅ |
| `cb` | CacheUpdater | `cb.c` friends_formerly | ✅ |
| `ac` | DecodeBuffer | `ac.x/ac.z` mirrors | ✅ |
| `fa` | ClientIOException | `fa.d` version, `fa.e` stackable | ✅ |
| `ia` | SpriteScaler | `ia.g/ia.a` ignores | ✅ |
| `db` | LinkedQueue | `db.g` ignores_count | ✅ |
| `ka` | IntHolder | `ka.c` item_members | ✅ |
| `fb` | SurfaceImageProducer | `fb.k` gameFontStates | ✅ |
| `qb` | GameFrame | `qb.k` gameFontData | ✅ |
| `ja`(K)/`s`(c) | RSA modulus/exponent | ✅ |

**Class coverage: 36/36 obf classes rscplus touches are present in NAMING.** No missing class.

**The one real conflict — `m`:**
- rscplus's `patch()` line 75 dispatches `node.name.equals("m") → patchData`, and `patchData` hooks **`m.a([BBZ)V`** (a byte-array decode that, on completion, populates item names/commands and world-map scenery/boundaries). That is a **game-data/config decode method**.
- NAMING.md row `m` = **SocketFactory** (`abstract Socket a(byte)`, `new Socket(host,port)`). A SocketFactory is not the data-decode class, and `a([BBZ)V` is not `a(byte):Socket`.
- **Conclusion:** either (a) NAMING.md has `m` mislabeled (the genuine data/config-decode class is `m`, and SocketFactory is some other letter), or (b) the deob assigned SocketFactory to `m` correctly and rscplus is decoding a method on a class the map filed elsewhere. **This must be resolved before any re-point** — it is the single class-level cross-reference that does not line up between the patcher and our map. (Cheap to settle: disassemble `m.a([BBZ)V` from the jar and compare to the deob `SocketFactory.java` vs the data classes `nb/ob/t/EntityDef`.)

**Doc inconsistency to fix (not a coverage gap):** `MUDCLIENT_SKELETON.md` lists `Hh=scene` typed `k (Scene)` and `Ek=world` typed `lb (World)` — the **pre-correction (swapped)** identities. NAMING.md's correction (and the patcher's actual usage: `lb` is the 3D render-helper holding `[J D` depth buffers + `Camera` distances; `k.e` is a static `long` world-tick marker) confirm **`lb=Scene`, `k=World`**. SKELETON.md should be reconciled so a re-point doesn't pick the wrong target for camera/scene hooks.

**Field-name verification depth:** NAMING covers classes; individual obf field letters (`client.Jh`, `ta.C`, `qa.w`, etc.) are *not* enumerated in NAMING but the load-bearing ones are documented in `MUDCLIENT_SKELETON.md` (e.g. `Jh=clientStream`, `mg=incomingPacket`, `li=surface`, `Hh/Ek` scene/world, `wi=localPlayer`, `Tb/rg` npc/player arrays, panel `qa` instances). The ~116 `client` field mirrors are **the highest-risk part of any re-point**: each is a single obfuscated letter+descriptor, and only a subset are individually named in SKELETON. A faithful re-point needs a complete `client` field map (484 fields), which neither doc fully provides.

---

### 4. Re-pointing approaches — effort & risk

Baseline reality: the deob is a **comprehension reference, not a compilable build** (it still references dead Microsoft J++ stub classes). Approaches (b) and (c) presuppose first making the deob compile — that is itself a substantial prerequisite, separate from the hook re-point.

#### (a) Keep ASM patching, retarget to named classes/fields

Change every `node.name.equals("ua")` → `"client/scene/Surface"`, every `hookClassVariable(... "client", "Jh", "Lda;" ...)` → `(... "client/Mudclient", "clientStream", "Lclient/net/ClientStream;" ...)`, and every emitted `FieldInsnNode/MethodInsnNode` owner+name+desc to the named equivalents.

- **Effort: HIGH but purely mechanical.** ~27 class names + ~166 mirror tuples + 57 field-rewrite tuples + 7 method-rewrite tuples + all the descriptor constants in Reflection ≈ **~300 string edits**, each requiring the obf→deob field/method name *and* the rewritten type descriptor (`Lda;`→`Lclient/net/ClientStream;`, `[Lta;`→`[Lclient/world/GameCharacter;`). The field-name half does not exist yet at full fidelity (see §3) — you must first produce a complete per-field obf→name table for `client` and `ta`/`qa`.
- **Risks:**
  - **Descriptor mismatches (HIGH):** `hook*Variable` matches on `field.desc.equals(desc)` — if a single descriptor is wrong the hook silently no-ops (Reflection silently nulls). Re-deriving descriptors for 230 members is the main error surface.
  - **Behavioral drift (HIGH):** the patcher's `find*` loops match on bytecode shape (specific `LDC 334.0`, `SIPUSH 510`, instruction adjacency, `getPrevious().getPrevious()`). Recompiling the deob with a modern javac produces **different bytecode** (constant folding, branch ordering, `StringBuilder` vs concat, frame layout) — every shape-matching hook (`patchRenderer`, the ZZZ/right-click-bounds rewrites, `patchRightClickMenu`'s "+26 nodes" walk) can break even with correct names. These are the fragile ones.
  - **COMPUTE_FRAMES (MEDIUM):** the writer uses `COMPUTE_MAXS` only. Vanilla classes are JDK-1.1/J++ era (no stack-map frames), so this works. A **recompiled JDK-8+ deob has stack-map frames**, and ASM tree transforms that change control flow without `COMPUTE_FRAMES` will produce `VerifyError`. Re-pointing onto recompiled classes forces `COMPUTE_FRAMES` (and a `ClassWriter` that can resolve common-superclasses of the named hierarchy) — a real change, not just renames.
- **Verdict:** lowest *conceptual* change, but you inherit both the descriptor-fidelity problem and the bytecode-shape-fragility problem. Best if you keep wrapping the **obf jar** and only relabel for readability (then names are cosmetic and shapes are unchanged — LOW risk); HIGH risk if you point at a recompiled deob.

#### (b) Replace ASM with direct source hooks in the recompiled deob

Delete the field mirrors/instruction rewrites and instead call `Game.*` hooks directly from named methods (`Surface.present()` calls `Renderer.present(...)`; `Mudclient.handlePacket()` calls `Replay.patchClient()`; expose `clientStream` etc. as fields Game can read).

- **Effort: VERY HIGH.** Requires (1) making the deob compile (resolve the J++ stubs, pick the pure-AWT path) and (2) hand-placing ~70+ hook callsites that the patcher injects at precise bytecode offsets — including the shape-matched ones (ZZZ scaling, right-click bounds, font toggle, combat-menu show/hide) that have no clean "method boundary" to hook from source. Many patcher hooks insert *mid-method* at a matched instruction; replicating them in source means understanding and re-deriving the surrounding logic.
- **Risks:** **Behavioral drift becomes a correctness problem you own** — you're now editing game logic, and any divergence from vanilla breaks G2 (rscplus features) and G3 (replay fidelity). Replay capture/playback (`Replay.*` hooks in `patchClient`/`patchGameApplet`) is especially sensitive: the raw input/output stream dumps and opcode checkpoints must byte-match. Eliminates COMPUTE_FRAMES/descriptor risk, but trades it for "is my hand-written deob behaviorally identical to the obf jar?" which is much harder to prove.
- **Verdict:** cleanest end state, highest effort and highest behavioral risk. Only worth it once the deob is provably equivalent.

#### (c) Replace Reflection.java with typed access against the compiled deob

Swap the 48 `Field`/74 `Method`/3 `Constructor` reflective handles for direct typed access (`mud.clientStream`, `character.currentHits`, `new Panel(surface, n)`).

- **Effort: MEDIUM** (smaller and more localized than the patcher) — but **gated on the deob compiling** and on visibility: most target fields/methods are `private final` in the obf classes. Direct typed access needs them made package-visible/public, or accessors added — i.e. you *do* edit the deob, you just edit declarations instead of bodies.
- **Risks:** **LOWEST.** Compile-time type checking replaces silent-null reflection — descriptor mismatches become compiler errors instead of runtime no-ops, which is strictly safer. No COMPUTE_FRAMES concern (no bytecode transform). Main risk is the visibility edits subtly changing semantics (unlikely) and the same `m`/`na` and Scene/World labeling questions surfacing as type errors (which is good — they'd be caught).
- **Verdict:** the most attractive first step *if* the deob compiles, and a good forcing function to validate the obf→deob field/method map.

#### Cross-cutting

- **G1/G2 first-move recommendation:** the lowest-risk path to "build and run rscplus" is to **fork and build it unchanged against the existing obf jar** (Ant/JDK-8, libs in `lib/`) — no re-point needed for it to function. Re-pointing onto the deob is a *separate, optional* enhancement that only pays off once the deob is recompilable.
- **Prerequisite for (b)/(c):** a complete, verified **per-field obf→name table for `client` (484 fields) and `ta`/`qa`** with descriptors. Today only a load-bearing subset is named (SKELETON.md). This table is the gating artifact for any re-point and the place to also resolve the `m` conflict and the Scene/World label reconciliation.
- **Biggest single risk across all approaches:** the patcher's reliance on **bytecode-shape matching** (constants, instruction adjacency, fixed node-walk distances). These survive only because the obf jar's bytecode is fixed. Any recompiled deob invalidates them regardless of how correct the names are — so approach (a)-onto-deob and (b) both must budget for re-deriving every shape-matched hook, while (a)-onto-obf-jar (rename-only) and (c) avoid that entirely.

**Files referenced:** `/home/free/code/rsc-hacking/rscplus/src/Client/JClassPatcher.java`, `/home/free/code/rsc-hacking/rscplus/src/Game/Reflection.java`, `/home/free/code/rsc-hacking/westworld/reference/rsc-client/docs/NAMING.md`, `/home/free/code/rsc-hacking/westworld/reference/rsc-client/docs/MUDCLIENT_SKELETON.md`.
