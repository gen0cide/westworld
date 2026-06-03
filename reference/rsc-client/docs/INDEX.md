# RSC Client Deobfuscation — documentation index

All documentation for the deobfuscation of `rscplus/assets/rsclassic-1091943135.jar`
(**RuneScape Classic rev ~233–235, a Microsoft J++ build**) into ~50k lines of readable Java.
Deobfuscated source lives one level up in `../src/`. Shipped as
[gen0cide/westworld PR #1](https://github.com/gen0cide/westworld/pull/1).

> **The deob now compiles + runs.** It builds 0 errors / **78 classes** under JDK 17 (**76 source
> files**, incl. the **7 extract-delegate classes** the Mudclient de-god split produced), boots,
> autologins against a live local OpenRSC server, holds a stable multi-minute session, and renders
> the live in-game 3D world.
> → **[build/RUN.md](build/RUN.md)** — reproducible build + run recipe (JDK 17, server,
> content host, `client.Boot`, the headless bring-up hooks, the `systemd-run` launch, frame
> capture).
> → **[build/FUNCTIONING_CLIENT.md](build/FUNCTIONING_CLIENT.md)** — achievement summary:
> the M0→live-3D milestone ladder, screenshots, the class of found-by-running bugs, the
> commit trail, and an honest WHAT WORKS / WHAT'S NEXT.
> → **[build/LAND_READINESS.md](build/LAND_READINESS.md)** — go/no-go for PR #1: the
> compile/build/sync gates, the Mudclient de-god decomposition (pure-motion verified,
> `functionalDivergence=false`), the terrain base-shade fix, the orsc↔JAR parity verdict per
> E-row, residual gaps, and next steps.
> → **[build/REFACTOR_RESULTS.md](build/REFACTOR_RESULTS.md)** — the Mudclient de-god split in two
> waves (round 1: `ClientPackets` + `ClientSound`; round 2: `TradeDuelBankPackets`,
> `WidgetRenderer`, `GameInterface`, `MenuController`, `IncomingPackets`): what moved, the line
> delta (13416 → 8977), the byte-identical pure-code-motion audit, and the live smoke test.
> → **[build/RENDER_FIDELITY_FINDINGS.md](build/RENDER_FIDELITY_FINDINGS.md)** §0 — the orsc
> terrain base-shade reseed fix: dropping OpenRSC's `RSModel` project-time light reseed to match
> the rev-235 `GameModel` (grass green 0x8e → 0x4a ≈ the JAR's 0x48), geometry unchanged.
> → **[build/ORSC_JAR_PARITY.md](build/ORSC_JAR_PARITY.md)** — the orsc-vs-rev-235-JAR runtime
> render parity: byte-identical face sets on every JAR-runnable fixture, all 5 E-rows settled,
> the terrain base-shade divergence now RESOLVED, the GameData-table blockers, and the residual
> harness camera/fog pixel drift.

## Read in this order

| Doc | What it is |
|-----|-----------|
| [build/README.md](build/README.md) | **Next phase → start here for building a runnable client.** Plan + options (A–D) for turning this deob into something that builds/runs and wiring it to rscplus + the Go cradle/render lib. Includes a recommendation and a Week-1 milestone. |
| [ARCHITECTURE.md](ARCHITECTURE.md) | **Start here for understanding the client.** How it works: entry/lifecycle, every subsystem (render, net, world, scene, data, audio, native), and the one-frame data flow. |
| [NAMING.md](NAMING.md) | The verified obfuscated→readable map for all 71 classes, with package, role, oracle evidence, and the full inheritance hierarchy. |
| [MUDCLIENT_SKELETON.md](MUDCLIENT_SKELETON.md) | Structural index of the main `client`→`Mudclient` class: all 484 fields + 113 methods grouped by function (bootstrap/login/mainloop/packetin/packetout/scene/ui/input/util), plus the de-god extract-delegate split into 7 sibling classes (`ClientPackets`, `ClientSound`, `TradeDuelBankPackets`, `WidgetRenderer`, `GameInterface`, `MenuController`, `IncomingPackets`). |
| [VINEFLOWER_FIX.md](VINEFLOWER_FIX.md) | Root-cause + fix for the decompiler failure (bogus overlapping exception-table ranges). Repro/build/run commands. Cut failures 238→4. |
| [WORKSPACE_README.md](WORKSPACE_README.md) | The reproducible pipeline (decompile → strip → normalize → deobfuscate) and workspace layout. |
| [NAMING_DRAFT.md](NAMING_DRAFT.md) | The initial naming hypothesis (kept for provenance; superseded by NAMING.md). |
| [audit_manifest.tsv](audit_manifest.tsv) | Per-class audit manifest: obf name, deob name, package, count of methods recovered by the decompiler fix. |
| [audit/rendering_clean_methods.txt](audit/rendering_clean_methods.txt) | Ground-truth method inventory for the rendering classes (used to verify 100% method presence). |

## What we did (summary)

**Goal.** Turn a heavily-obfuscated, Jagex-signed RSC client jar into human-readable Java —
optimizing for **readability and correctness**, not recompilation (it intentionally still references
long-dead Microsoft J++ stubs, so it's a comprehension reference).

**The jar.** RSC revision ~233–235, a Microsoft J++ build: 71 obfuscated classes using
`com.ms.directX`/DirectSound/Win32 for optional native acceleration over the pure-Java/AWT path.

**Pipeline.**
1. **Decompiled** with Vineflower (primary) + CFR (cross-reference).
2. **Fixed the decompiler.** The obfuscator injects dozens of bogus, partially-overlapping
   `RuntimeException` exception-table ranges whose handler is a dead lone `athrow`; these make
   Vineflower's CFG decomposition throw `parsing failure!` and bail on **238 methods across 41 of 71
   files**. Not fixable by upgrading or any CLI flag. We built an ASM pass that strips exception-table
   entries whose handler is just `ATHROW` (only the table changes → all classes still JVM-verify),
   cutting failures to **4 methods / 4 files** and recovering real source for 238 methods.
3. **Shift-normalized** the junk-masked bit-shifts (`x >> 1599703024` ≡ `x >> 16`; Java masks shift
   counts to 5 bits) — lossless.
4. **Deobfuscated per-class** with parallel multi-agent workflows (mixed Opus/Sonnet): stripped the
   obfuscation (always-false `client.vh` opaque predicate, per-method profiling counters, catch-rethrow
   wrappers, anti-tamper constant guards, XOR string pools), renamed every symbol, added doc comments,
   and matched logic to three oracles — **OpenRSC Payload235** (exact rev-235 protocol opcodes),
   **LeadingBot mudclient.java**, and **mudclient204**.
5. **Correctness comb-over (all-Opus).** Re-audited every class against the clean (post-fix)
   decompilation as ground truth — fixing **~236 logic bugs** that the first pass (built on the
   defective decompilation) had introduced, and correcting a class-identity swap in the map
   (`k`=World/terrain, `lb`=Scene/renderer were backwards).
6. **Rendering full-expansion.** The renderer is the highest-value code, so it got a dedicated pass +
   independent verification: **100% method presence, zero summarized bodies** — Surface 58/58,
   Scene 40/40, World 37/37, GameModel 38/38, and all sprite/raster/font helpers. Fully transcribed the
   three previously-summarized hot spots (Scene's convex-overlap `polygonsOverlap`, World's
   `loadMapData`, SurfaceSprite's `flushEventQueue`) and fixed a WorldEntity normalX/normalY inversion.

**Result.** 71 classes / ~50k lines of named, doc-commented Java (incl. the assembled `Mudclient.java`,
11.5k lines), organized into packages, with the docs above. Tooling from the decompiler fix lives at
`~/code/rsc-hacking/deob/tools/{strip-obf-exceptions.jar, vineflower-fixed.jar}` and is reusable for any
Jagex-obfuscated jar.
