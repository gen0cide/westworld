> _Verbatim execution report from the `rsc-client-milestone1-numbers` workflow (2026-06-01). See `../../MILESTONE_1_RESULTS.md` for the synthesis._

# RSC Deob Subsystem Compile + Net Protocol Equivalence Report

**Worktree:** `/home/free/code/rsc-hacking/westworld/.claude/worktrees/wf_1b683d6f-41a-3`
**JDK:** verified `openjdk 17.0.19` (JAVA_HOME=/usr/lib/jvm/java-17-openjdk).
**Note on setup:** the worktree branch HEAD (`3a384ec`) predates `reference/`; the deob sources live on `main` (`5b78e9b`). I copied `reference/` from the main checkout into the worktree (isolated; does not touch main).

## Q1 — Subsystem compile (breadth)

I applied the dead-dep fixes, then compiled the whole subsystem at once (sourcepath = all 66 non-`Mudclient` files; the packages are mutually circular — `util↔net↔scene↔world↔ui↔data↔audio↔shell` — so no strict lowest-dependency-first topological order exists; an isolated per-package compile is impossible because every package both provides to and consumes from others).

**Headline:** With JDK-17 + dead-dep fixes applied, the FIRST hard error was the `_` parameter in `ui/Panel.java` (Java 9+ reserved word). After fixing that, the true error inventory is **1158 errors**, distributed across all packages — but the cause is structural, not obf-drift:

| category | count | meaning |
|---|---|---|
| `X is not public in client.Y` | 329 | deob split a flat default-package jar into `client.*` subpackages; 47 top-level types stayed package-private |
| `cannot find symbol` | 753 | almost entirely **cascades** from the unresolved not-public imports |
| Mudclient refs | 22 | out of scope (excluded) |

To isolate genuine drift from the visibility noise I ran a mechanical experiment: made all 49 top-level types `public`. Errors did **not** drop (1250) — they shifted to a *second* visibility layer (package-private **members** accessed cross-package). Final categorization post-widening:
- **78** package-private member (field/method/ctor) access errors
- **22** Mudclient (out of scope)
- **7** genuine signature/drift errors (`cannot be applied to given types`) at exactly 7 sites: `net/Packet.java:467`, `scene/GameModel.java:302/304/346`, `shell/GameShell.java:658`, `shell/LoaderThread.java:413`, `world/World.java:877`. Example (real bug): `CachePath.resolveFile(int,null,String,int)` called but declared `(int,String)`.
- rest = cascading `cannot find symbol`

**Conclusion:** the deob does NOT compile as subpackages, but the blocker is **architectural (the flat-jar→subpackage visibility split), not obf-name drift**. Genuine residual drift is small (~7 signature sites + a handful of unimported bare obf-aliases `mb`/`i`/`ga`/`w`/`Utility` in `net`). The `GameShell extends java.applet.Applet` base is fine on JDK 17 (deprecation **warning** only). The dead-dep `nativeapi` package compiles **clean standalone** after my fixes.

**Per-package error counts (baseline):** net 335, scene 195, audio 153, shell 121, world 104, util 100, data 88, ui 62, **nativeapi 0**.

## Q2 — Protocol equivalence (depth) — net layer

**Step 4 — Known-Answer Tests (all PASS), harness in `q2kat/`:**
- **ISAAC** (`net/ISAAC.java`): deob keystream == mudclient204 ISAAC == OpenRSC `orsc.net.Isaac` for both an all-zero seed and a random 8-int seed, 64 words each. Zero-seed first word `0x182600f3` is the published ISAAC zero-seed test vector. **PASS.**
- **Buffer/BitBuffer round-trips** (`net/Buffer.java`, `net/BitBuffer.java`): byte/short/int3/int/long/signed-short/smart-signed/varlen put↔get identity; `BitBuffer.readBits` cross-checked vs mudclient204 `Utility.getBitMask` across widths {1,2,3,5,7,8,11,13,16,19,24,31,32}. **PASS.**
- **XTEA** (`Buffer.teaDecrypt`): output byte-identical to OpenRSC `RSBuffer.a(...)` XTEA-decrypt on the same ciphertext. The prior-flagged dropped `n7 +` term is **present and correct** — a negative-control variant without it diverges. **PASS.**
- **RSA** (`Buffer.rsaEncrypt`): == `BigInteger.modPow(exponent, modulus)` + **2-byte** length prefix, using the jar's baked-in modulus `ca950472…` (1024-bit → 128-byte ciphertext, framed length 130). The 2-byte framing matches OpenRSC rev-235 `RSBuffer.encodeWithRSA` (`putShort(len)`); rev-204 used a 1-byte prefix, so this is the correct rev-235 behavior. **PASS.**

**Step 5 — Direct jar-vs-deob reflection (NOT blocked — stronger than expected):**
The prior-findings `client.vh` concern was real (`o(int[])`, `o.b(int)`, `tb.a(...)` all open with `getstatic client.vh:Z`), but on JDK 17 the `client`/Mudclient `<clinit>` **completes cleanly** (no applet/graphics needed). Via URLClassLoader on the vanilla jar:
- obf ISAAC `o` keystream == deob ISAAC: **MATCH** (`405143795, 806046349, …`).
- obf Buffer `tb` XTEA `a(byte,int,int[],int)` == deob: **MATCH** (passing dummy `87`).
- obf Buffer `tb` RSA `a(BigInteger,int,BigInteger)` == deob: **MATCH** (passing dummy `-94` to avoid the obf's `-98/((d-6)/52)` div-by-zero anti-tamper; framed len 130 both).

**Step 6 — Tier-2 decompile-diff on net (vineflower), workspace in `diffwork/`:**
Compiled the real `net/{ISAAC,Buffer,BitBuffer}.java` (with minimal stubs for the unimported obf-aliases `mb/i/ga/w/Utility/StreamBase` + `InputState/DownloadWorker` — themselves the net drift inventory), decompiled both the deob `.class` and the jar's `o/tb/ja.class`. After normalizing the obf noise (shift constants mod 32, opaque `client.vh` branches, `h++/j++/m++` profiling counters, complement comparisons `-1==~x`, dummy params, RuntimeException rethrow wrappers), the **semantic delta is empty** for: ISAAC `getNextValue`+`isaac` refill, Buffer `teaDecrypt` (incl. the `n7 +` term), and Buffer `rsaEncrypt` (`modPow(exp,mod)`+`putShort`+`putBytes`).

## Edits made (reproducible)

Deob source edits under `reference/rsc-client/src/client/`:
1. **Deleted** 4 dead Microsoft-dep files: `nativeapi/DirectSoundPlayer.java`, `nativeapi/DirectDrawModes.java`, `nativeapi/Win32MouseCallback.java`, `audio/AudioOutput.java` (the latter only implemented by DirectSoundPlayer).
2. **`util/JSBridge.java`**: dropped `import netscape.javascript.JSObject`; `call(...)` now a no-op returning `null`.
3. **`shell/LoaderThread.java`**: removed imports of deleted classes; replaced `DirectDrawModes directDraw` / `Win32MouseCallback win32MouseCallback` fields (deleted) and retyped `displayModeSetter`→`DisplayModeSetter`, `robotCursor`→`RobotCursor`; rewired all `Class.forName("ha"/"j"/"rb").newInstance()` + reflective `getMethod(...).invoke(...)` to direct typed calls, fixing drifted method names: `enter`→`enterFullscreen(Frame,int,int,int,int)`, `exit`→`exitFullscreen()`, `listmodes`→`listAvailableModes()`, `movemouse`→`moveMouse(int,int)`, `showcursor`→`setCursorVisibility(Component,boolean)`, `setcustomcursor`→`setCustomCursor(...)`; dropped the dead Win32/DirectDraw branches. (My rewiring is clean — 0 residual errors reference the deleted classes. The 37 remaining LoaderThread errors are the subpackage-visibility issue + the pre-existing `CachePath.resolveFile` signature drift at line 413, untouched.)
4. **`nativeapi/DisplayModeSetter.java`**: widened `final class`→`public final class` so cross-package `LoaderThread` can reference it.
5. **`ui/Panel.java`**: renamed the reserved-word parameter `byte _`→`byte unusedGuard` (Java 9+ fix).
6. **Experimental (measurement only):** widened all 49 package-private top-level types to `public` to isolate genuine drift from visibility noise — this is currently applied in the worktree; it's a measurement transform, not a recommended fix (the right fix is collapsing the subpackages or widening members too).

Harness files I created (not deob edits): `q2kat/` (KAT + jar-reflection harnesses, compiled+passing) and `diffwork/` (decompile-diff workspace).

## NUMBERS

- **Q1 = 0/8 subsystem packages compile clean** (whole-subsystem compile = **1158 errors** in the deob-as-authored state, +`_`-keyword fix). Root cause = flat-jar→subpackage **visibility split** (329 not-public + ~753 cascades), NOT obf-drift. Out-of-scope Mudclient refs = 22. After mechanically widening top-level types to public: 1250 errors → genuine non-visibility, non-Mudclient drift narrows to **7 signature sites** + a few unimported bare obf-aliases. The fixed dead-dep `nativeapi` package compiles **clean standalone**.
- **Q2 per-primitive:** ISAAC **PASS**, Buffer round-trip **PASS**, BitBuffer.readBits **PASS**, XTEA **PASS** (n7+ term verified), RSA **PASS** (2-byte frame, jar modulus).
- **Q2 jar-reflection:** **NOT blocked** — `client.vh` `<clinit>` completes on JDK 17; obf `o` ISAAC, obf `tb` XTEA, obf `tb` RSA all **MATCH** the deob byte-for-byte.
- **Q2 net decompile-diff:** **empty semantic delta** (after obf normalization) for ISAAC getNextValue/isaac/init, Buffer teaDecrypt, Buffer rsaEncrypt.
