# RuneScape Classic client — deobfuscated reference

Human-readable reverse-engineering of `rscplus/assets/rsclassic-1091943135.jar`.

## What this jar is
- **RSC client, revision ~233–235**, a **Microsoft J++ build**: 71 obfuscated classes (default
  package, signed `Created-By: Jagex`). Uses `com.ms.directX` / DirectSound / `com.ms.win32.User32`
  / `netscape.javascript` for optional native acceleration — these degrade to pure-Java AWT.
- Entry: `e` (GameShell, `extends Applet`) → `client` (the main `mudclient` class).
- Goal here is **readability, not recompilation.** (Recompiling would need long-dead Microsoft J++
  stub classes plus fixes to ~760 decompiler bugs; not worth it for a comprehension reference.)

## Layout
```
decompiled/
  vineflower/   raw Vineflower 1.10.1 output (best base; obfuscated names)
  cfr/          raw CFR 0.152 output (cross-reference; differs where VF is wrong)
  normalized/   vineflower + lossless shift-normalization pass  ← deobfuscation base
src/            deobfuscated, renamed, commented Java, organized into packages:
                client, client.{shell,scene,net,world,data,ui,audio,nativeapi,util}
docs/
  NAMING_DRAFT.md   initial class-naming hypothesis (inheritance + anchors)
  NAMING.md         verified obf→name→package map for all 71 classes
  ARCHITECTURE.md   how the client works (engine/render/net/world/data + native paths)
```

## Pipeline (reproducible)
0. **Strip obfuscated exception tables** (the key fix). The obfuscator injects dozens of bogus,
   partially-overlapping `RuntimeException` ranges whose handler is a dead lone `athrow`; these
   make Vineflower's CFG decomposition throw `parsing failure!` (`DomHelper.parseGraph`) and bail
   with `// $VF: Couldn't be decompiled` on **238 methods / 41 files**. `deob/tools/strip-obf-exceptions.jar`
   (ASM) drops any try-catch whose handler's first opcode is `ATHROW` — only the exception table
   changes, so all 71 classes still pass JVM verification. After stripping: **4 method failures / 4 files**.
1. Decompile: `java -jar ~/code/rsc-hacking/deob/tools/vineflower-fixed.jar <jar> <out>` (does step 0 +
   decompile in one shot). Plain `vineflower.jar` / `cfr.jar` remain as cross-reference. The deob waves
   below ran on the *un-stripped* base + CFR fallback; the clean base is in `decompiled/vineflower-clean/`
   + `decompiled/normalized-clean/` (only `ba`/`ib`/`k`/`client` retain 1 residual each — genuine
   decompiler structuring limits, not obfuscation).
2. **Shift normalization** (`/tmp/normalize_shifts.py`): the obfuscator masks shift amounts with junk
   (`x >> 1599703024`); Java uses only the low 5 bits, so this is rewritten to `x >> 16` — lossless,
   206 sites fixed. Output → `decompiled/normalized/`.
3. **Per-class deobfuscation** (parallel agent workflow): for each class, strip the obfuscation
   (opaque predicate `client.vh`, profiling counters, per-method exception wrappers, anti-tamper
   guards, junk bit-masks), rename every symbol, add doc comments, match logic to the oracles below.

## Oracles used for naming
- `~/code/rsc-hacking/openrsc/.../Payload235{Parser,Generator}.java` — **exact rev-235 protocol** (opcodes).
- `~/code/rsc-hacking/rscdump.com-runescape-classic-dump/LeadingBot/mudclient.java` — readable deob client (9474 lines).
- `~/code/rsc-hacking/mudclient204/src/` — rev-204 named client (structural anchor).

## Status
- [x] Decompiled (Vineflower + CFR), shift-normalized base.
- [x] Workspace + class-naming draft.
- [ ] `NAMING.md` (all 71) + core-class rewrite — **wave 1 in progress** (workflow `rsc-deob-wave1`).
- [ ] Wave 2: giants (`client`, `ua`=Surface, `lb`=World, `k`=Scene) + remaining utility/data classes + `ARCHITECTURE.md`.
