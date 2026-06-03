> _Verbatim investigation report from the `rsc-client-build-options` workflow (2026-06-01). One of five parallel agents; see `../OPTIONS.md` for the synthesis._

## A5 — Deob↔obf fidelity & verification strategy

**Scope.** This is a risk memo about one question: *is the deobfuscated client behaviorally 1:1 with `rsclassic-1091943135.jar`?* Every build option that uses the deob **as the actual running client** (G1's "re-point monkey-patching at our code") inherits this risk. I read the audit docs and the `net/` package (ISAAC, Buffer, BitBuffer, Packet, ClientStream) and cross-checked rscplus's obf references. Bottom line up front: **the deob is an excellent comprehension reference with credible coverage, but nothing in the current artifact proves runtime equivalence — there is zero executable verification today (no build, no test, no byte-diff, no replay). Treat "deob = real client" as UNVERIFIED until the cheap checks below pass.**

---

### 1. What the docs claim — and what that does / does not guarantee

**The confidence story the docs tell:**

- **Decompiler-fix provenance (strong).** The obfuscator injected bogus overlapping `RuntimeException` exception-table ranges whose handler is a lone `athrow`, breaking Vineflower on **238 methods across 41/71 files**. An ASM pass strips only exception-table entries whose handler is just `ATHROW` (table-only change → all classes still JVM-verify), cutting failures to **4 methods / 4 files** and recovering real source. This is a legitimate, low-risk transform and the most trustworthy claim in the set.
- **Shift normalization (lossless).** Junk-masked shifts (`x >> 1599703024` ≡ `x >> 16`, Java masks shift counts to 5 bits) reduced to canonical form. Mechanically sound — I verified the reasoning holds in ISAAC (`<< -402254995` → `<< 13`, etc.) and in Buffer's XTEA shifts.
- **Per-class deob + 3 oracles.** Stripped `client.vh` opaque predicate, profiling counters, catch-rethrow wrappers, anti-tamper guards, XOR string pools; matched against **OpenRSC Payload235** (rev-235 opcodes), **LeadingBot mudclient.java**, **mudclient204**.
- **Correctness comb-over (the key claim).** Re-audited *every* class against the **clean post-fix decompilation as ground truth**, fixing **~236 logic bugs** the first pass (built on the *defective* decompilation) had introduced.
- **World/Scene swap resolved.** `k`=World (terrain), `lb`=Scene (renderer) were originally backwards. I **verified this is fixed in source** (World.java = terrain methods, classname `World`; Scene.java = render methods, classname `Scene`).
- **Rendering full-expansion.** 100% method presence, zero summarized bodies (Surface 58/58, Scene 40/40, World 37/37, GameModel 38/38), with three hot spots fully transcribed and a WorldEntity normalX/normalY inversion fixed.

**What this genuinely buys you:** high confidence in *structural* completeness (every class, every method present) and in *human-reviewed* per-method correctness against a clean decompile. The comb-over is the right methodology — auditing against clean bytecode rather than the broken first pass is exactly what reduces semantic drift.

**What it does NOT guarantee (the honest gaps):**

1. **"236 bugs fixed" is a measure of effort, not residual correctness.** It tells you the first pass was *very* buggy (236 known errors from building on a defective decompile). It says nothing about how many of the same class of error *survive* unspotted. The comb-over is a manual LLM review, not a proof.
2. **No executable verification anywhere.** There is no build (the deob still references dead Microsoft J++ stubs — `com.ms.*`, DirectX/DirectSound — in 9 files incl. Mudclient.java), no unit test, no byte-for-byte protocol diff, no decompile-roundtrip-diff. **Zero runtime evidence.** Every correctness claim is review-based.
3. **Rendering "100% method presence" ≠ behavioral equivalence.** Presence is a structural property. A transcribed-but-subtly-wrong polygon sort or off-by-one in a raster span is exactly what method-presence checking cannot catch — and it's precisely G4's failure mode ("can't perceive a door").
4. **Deliberate semantic edits are baked in.** The deob *removes* anti-tamper guard branches and dead stores it asserts are dead. If any "always-false guard" is *not* always false at some call site, control flow changed. These judgments are individually plausible but unaudited in aggregate.
5. **Decoded constants are claims, not byte-checks.** The RSA modulus in BitBuffer (`ca950472...`), the ISAAC shift reductions, XOR string-pool keys, the XTEA delta — all are presented as decoded values in comments. None is shown to be byte-verified against the jar's constant pool.

---

### 2. 71-class coverage check — **PASS, with two naming caveats**

- **File count: confirmed 71.** `find … src -name '*.java'` returns exactly **71** (70 under `client/{audio,data,nativeapi,net,scene,shell,ui,util,world}` + `client/Mudclient.java`). Matches NAMING.md's 71 and the 71 rows of `audit_manifest.tsv`. **Every obf class has `deob_exists=1`.**
- **rscplus obf-name cross-check: every class rscplus touches is covered.** rscplus's `JClassPatcher` dispatches on `{ua, e, qa, m, client, f, da, lb, sa, wb, pb}` and `Reflection.java` loads `{b, ca, client, da, e, f, ja, k, kb, lb, na, qa, ta, tb, ua, wb}`. **The union of these is a strict subset of our 71-class manifest** — no obf class referenced by rscplus is missing or only partially covered. Mapping for the rscplus-critical ones:
  - `ua`=Surface, `e`=GameShell, `qa`=Panel, `client`=Mudclient, `da`=ClientStream, `lb`=Scene, `sa`=AudioChannel, `wb`=MessageList, `pb`=SourceLinePlayer, `b`=Packet, `ca`=GameModel, `ja`=BitBuffer, `k`=World, `kb`=InputState, `na`=StreamFactory, `ta`=GameCharacter, `tb`=Buffer. All present.
- **Caveat A — rscplus nicknames ≠ our role names (not a conflict).** rscplus calls obf `m`→`patchData` and `f`→`patchRandom`. Our manifest has `m`=SocketFactory and `f`=RecordLoader. I checked the patch bodies: `patchData` hooks `m.a([BBZ)V` — which our deob documents as `SocketFactory.initGameData(byte[],byte,boolean)` (one-time startup data load); `patchRandom` patches `f.a(ILtb;)V` to inject `java.util.Random` — which is the character-validation helper our deob found co-located in `RecordLoader`. **Same classes, same methods; only the human labels differ.** No coverage gap, but anyone re-pointing rscplus at the deob must map by *obf name + method descriptor*, never by rscplus's nickname.
- **Caveat B — stale prose in NAMING.md.** NAMING.md lines 18–19 still read "*still need to be swapped + cross-references reconciled*" for the World/Scene swap, contradicting the INDEX/commit which say it's done. The **code is correct** (verified above); the doc text is just lagging. Fix the prose so a future reader doesn't "re-swap" already-correct files.

---

### 3. Highest-behavioral-risk areas (risk-ranked), and how to verify each

#### RISK 1 — Protocol / net layer (CRITICAL: silent or total login failure)
The net code I read is the *highest-quality* part of the deob (clean, well-reasoned, oracle-cross-checked), yet it concentrates the most catastrophic failure modes because any single bit/byte error breaks the wire protocol and the server desyncs or rejects login. Specific hazards:

- **ISAAC opcode keystream (`ISAAC.java`, `Packet.finishPacket`/`isaacCommand`).** Outbound opcode = `getNextValue() + opcode`; inbound = `(value − getNextValue()) & 0xFF`. The `getNextValue()` underflow logic (`resultIndex-- == 0` → refill, `init()` arms cursor at 256) **must match exactly** or the two ISAAC streams desync after the very first packet and every subsequent opcode is garbage. The reduced shift constants (13/6/2/16, 11-2-8-16-10-4-8-9) are asserted, not byte-checked.
- **RSA login block (`Buffer.rsaEncrypt`, `BitBuffer.RSA_MODULUS`).** `modPow(exponent, modulus)` then `[2-byte len][cipher]`. **Concrete divergence found:** rscplus does **not** use the jar's baked-in modulus — `JConfig.SERVER_RSA_MODULUS` (a decimal key, per-world via `Settings.WORLD_RSA_PUB_KEYS`) *replaces* it at runtime, and the jar's value is the hex `ca950472…`. So (a) the deob's hardcoded modulus is a comment-claim, never byte-verified; and (b) any "deob-as-client" path must reproduce rscplus's **key-override mechanism**, or it will encrypt the login block to the wrong key and fail to authenticate against the worlds rscplus targets.
- **Buffer smart-ints / framing (`Buffer.getSmart*`, `putVariableLengthShort`, `Packet.finishPacket` length header).** The 1-vs-2-byte threshold (160 for packet length; 128/32768 for smart values; the `hi = 160 + len/256` split) is off-by-one-fragile. The deob also *rewrites* `~(x+y) < ~z`-style obfuscated comparisons into plain `<`; a single inverted comparison silently corrupts every variable-length read.
- **XTEA (`Buffer.teaDecrypt`).** The deob itself flags a prior CFR reconstruction that **dropped a leading `n7 +` term** — direct evidence this exact method had a real, plausible-looking bug. Good that it was caught; bad as a signal of how subtle these are.

**How to verify:** (a) **Decompile-diff** — recompile each net class to bytecode, decompile both it and the jar's class with the same tool, normalize, and diff method-by-method; zero semantic delta is the goal. (b) **Byte-compare protocol traffic** — drive the deob and the real jar (via rscplus) through the same login + a scripted session against the same server (or a captured replay), and diff the raw socket bytes. Identical ISAAC keystream and identical login block = the strongest possible evidence. (c) **Unit-test the primitives in isolation** — seed both ISAAC implementations with a fixed key and assert the first N keystream words match the real class via reflection; round-trip every `Buffer` put/get against the jar's `tb`.

#### RISK 2 — Rendering hot paths (HIGH: directly threatens G4's "faithful perception")
Surface/Scene/World/GameModel are where G4 lives. "100% method presence" was verified; *behavioral* correctness was not. The fixed WorldEntity normalX/normalY inversion and the three "previously summarized, now transcribed" hot spots (`polygonsOverlap`, `loadMapData`, `flushEventQueue`) are exactly the spots where a transcription slip produces a *visually plausible but wrong* result — which is the perception-bug class the colleague wants gone. World is terrain/collision (door-walkability!), so a `getWall*`/elevation/`route` error is a direct G4 hazard.

**How to verify:** (a) **Golden-frame diff** — render identical scene state (same map chunk + model list) through the deob Surface and through the real jar's `ua`, and pixel-diff the output buffers. (b) **Targeted collision/perception oracles** — for `World` walkability, assert the deob's wall/route output equals the jar's for a battery of tiles (this is also the direct cross-check for `westworld/render/`'s Go port). (c) Diff `westworld/render/` *and* the deob *against the same jar method* rather than against each other — three-way agreement is much stronger than two-way.

#### RISK 3 — Obfuscation-stripping changed control flow (MEDIUM-HIGH: subtle, broad)
The deob removed `client.vh` opaque predicates, profiling counters, catch-rethrow wrappers, **and anti-tamper guard branches/dead stores** on the *assertion* they're inert. The net files contain many such judgments: `flushPacket`'s `guardCode != -6924` branch, `verifyCrc`'s `dummy != -422797528`, `teaDecrypt`'s `dummy != 87`, ClientStream's `(!closing) != expectOpen`. Each is individually reasonable, but they are removed *manually* and unaudited in aggregate; one guard that is not actually always-false at one call site = changed behavior. Profiling-counter removal is safe; guard-branch removal is the real exposure. (Lower-priority but real: the deob *keeps* dead stores "for fidelity" in some places and *removes* them in others — inconsistent, though harmless.)

**How to verify:** This is the cleanest win for **decompile-diff against the clean (post-fix) decompilation**, since the comb-over already used that as ground truth — re-running it mechanically (not by eye) catches any guard whose removal altered the CFG. Also: search the jar for every callsite of each guarded method and confirm the passed constant matches the "always-X" assumption.

#### RISK 4 — Decoded constants are unverified (MEDIUM)
RSA modulus, ISAAC/XTEA shift reductions, XOR key tables, the `0x9E3779B9` golden-ratio constants. All asserted in comments.

**How to verify:** Cheap and high-value — extract the jar's constant pool / `<clinit>` and assert each decoded constant matches. The XOR string-pool decoders are even self-checking: run the deob's `decodeXor`/`xorDecode` and confirm output equals the claimed plaintext.

---

### 4. Recommended verification strategy (cheapest first; gate trust on each)

Any option that uses the deob as the real client should clear these **in order**, treating each as a gate:

1. **Tier 0 — Constant audit (hours, no build needed).** Dump the jar's constant pools (`javap -c -p` per class) and assert every decoded constant in the deob matches: RSA modulus, ISAAC/XTEA shifts and deltas, XOR key tables, framing thresholds (160/128/32768), the 5000-byte buffer sizes. Run the deob's own XOR decoders to self-check the string pools. Catches RISK 4 and the RSA-key divergence immediately.

2. **Tier 1 — Get it to compile (the prerequisite for everything real).** Stub the dead J++ `com.ms.*`/DirectX paths down to the documented pure-Java/AWT fallback (9 files reference them, incl. Mudclient.java). Until this compiles, **no runtime claim is testable** — this is the single biggest blocker between "comprehension reference" and "trustable client." Failure to compile cleanly is itself a fidelity signal (missing/renamed members).

3. **Tier 2 — Decompile-roundtrip diff (the workhorse).** Compile the deob, decompile both it and the jar with the *same* Vineflower-fixed tool, normalize naming, and diff per method. Prioritize: all of `net/`, then `world/`+`scene/`+`Surface`. **Any non-cosmetic delta is a bug or an intentional edit that must be justified.** This is the broadest, most mechanical catch for RISKs 1 and 3 and removes reliance on "236 bugs fixed" as a trust proxy.

4. **Tier 3 — Primitive unit tests via reflection.** Use rscplus's `JClassLoader` to load the real obf classes and assert the deob matches them on fixed inputs: ISAAC keystream from a seed, every `Buffer`/`BitBuffer` put/get round-trip, XTEA decrypt of a known block, `World` wall/route for a tile battery, RSA block for a fixed plaintext+key. Fast, deterministic, pins the protocol and collision primitives.

5. **Tier 4 — Replay/traffic diff (the integration proof, and it doubles as G3).** Drive the deob and the real jar (via rscplus) through the **same rscplus replay**, and byte-diff: (a) raw socket traffic (login block + ISAAC keystream + packet framing), (b) golden render frames. Identical bytes/pixels across a real session is the strongest evidence of 1:1. This is also exactly the cradle-tester pipeline (G3), so it's reusable infrastructure, not throwaway QA.

6. **Tier 5 — Three-way render reconciliation (closes G4).** Diff `westworld/render/` (Go), the deob (Java), and the real jar method-for-method on identical scene/terrain state. Three-way agreement is the bar for calling the Go lib a faithful 1:1 and for killing perception bugs like the unwalkable door.

**Practical guidance for G1 specifically:** when re-pointing rscplus's hooks at the deob, bind strictly by **obf name + method descriptor** (rscplus's `patchData`/`patchRandom` nicknames are misleading — they operate on `m`=SocketFactory and `f`=RecordLoader), and reproduce rscplus's **runtime RSA-key override** (`JConfig.SERVER_RSA_MODULUS`) rather than the jar's baked-in `ca950472…` modulus, or login will fail against rscplus's target worlds.

**Honest verdict.** Coverage is genuinely complete (71/71, every rscplus reference covered) and the net code reads as correct and well-oracled. But confidence today rests entirely on human review of a hand-stripped decompile, with **no executable check of any kind**. The probability of at least one surviving semantic divergence in ~50k lines of manually-deobfuscated, guard-stripped code is high enough that **no deob-as-client option should be trusted before at least Tier 0–3 pass**, and G3/G4 should not be trusted before Tier 4–5. The good news: the cheapest tiers (constant audit, decompile-diff, primitive tests) are high-coverage and the replay infrastructure needed for the strongest proof is the same infrastructure the project is already building.

Key files: `/home/free/code/rsc-hacking/westworld/reference/rsc-client/src/client/net/{ISAAC,Buffer,BitBuffer,Packet,ClientStream}.java`; docs at `/home/free/code/rsc-hacking/westworld/reference/rsc-client/docs/{INDEX.md,NAMING.md,audit_manifest.tsv,audit/rendering_clean_methods.txt}`; rscplus refs at `/home/free/code/rsc-hacking/rscplus/src/Client/JClassPatcher.java`, `/home/free/code/rsc-hacking/rscplus/src/Game/Reflection.java`, `/home/free/code/rsc-hacking/rscplus/src/Client/JConfig.java`.
