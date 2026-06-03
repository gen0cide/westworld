# Building a functioning client from the deob — plan & options

This folder plans the **next phase** after the deobfuscation: turning our readable RSC client
(`reference/rsc-client/src/`, the deob of `rsclassic-1091943135.jar`, rev ~233–235) into something
that **builds and runs**, and wiring it to the surrounding tools (rscplus, the Go cradle, the Go
render lib). It was produced by the `rsc-client-build-options` investigation workflow (5 parallel
agents + a synthesis pass, 2026-06-01).

## Read in this order

| Doc | What it is |
|-----|-----------|
| **README.md** (this file) | The framing, the recommendation, and the milestone status. |
| [MILESTONE_1_RESULTS.md](MILESTONE_1_RESULTS.md) | **Measured numbers (2026-06-01).** The estimates below were replaced with real results: protocol proven, build green, compile blocker diagnosed. |
| [MILESTONE_2_RESULTS.md](MILESTONE_2_RESULTS.md) | **Current status.** Compile gate did NOT close (347 member-drift pairs, not ~7 — M1 under-measured); net-handler audit found 5 divergences (3 world-perception). Strategic pivot: GO↔JAR diff first. |
| [NET_HANDLER_AUDIT.md](NET_HANDLER_AUDIT.md) | Per-opcode incoming-packet audit: ~62/70 MATCH, 5 catalogued divergences (D1 missing scenery handler, D2 ground-item gate, D3 boundary desync, …) with exact jar line refs + fixes. |
| [OPTIONS.md](OPTIONS.md) | **The decision doc.** Options A–D with an at-a-glance matrix, per-option pros/cons, the recommendation, and open questions. |
| [RENDER_DIFF_DESIGN.md](RENDER_DIFF_DESIGN.md) | **Implementation-ready design** for the 3-way render diff: the `rscdump/1` state-dump schema, per-engine produce/consume mapping, determinism plan, and the phased build (first tool = a hand-authored single-tile dump). |
| [RENDER_DIFF_STATUS.md](RENDER_DIFF_STATUS.md) | **Build status.** Phase 0 + the diff harness are DONE (GO engine): `rscdump/1` schema, `render.RenderDump`, `cmd/renderdiff` (pixel + structural), self-tested — flags a removed door, zero false positives. The two Java engines are the gated next step. |
| [findings/01-deob-compilability.md](findings/01-deob-compilability.md) | Can the deob compile? Dependency surface + the real blocker. |
| [findings/02-rscplus-hook-surface.md](findings/02-rscplus-hook-surface.md) | Every rscplus hook/reflection target, cross-referenced to our `NAMING.md`; re-pointing effort. |
| [findings/03-rscplus-build-baseline.md](findings/03-rscplus-build-baseline.md) | Does rscplus build/run today? (Yes — verified.) Toolchain, deps, runtime/server config, headless gap. |
| [findings/04-replay-cradle-render.md](findings/04-replay-cradle-render.md) | rscplus replays, the Go cradle, and the Go render lib — how they relate to the deob. |
| [findings/05-deob-fidelity.md](findings/05-deob-fidelity.md) | Is the deob behaviorally 1:1? Risk memo + a tiered verification strategy. |

Background for all of the above lives one level up in [`../`](../) — start with
[`../INDEX.md`](../INDEX.md), [`../ARCHITECTURE.md`](../ARCHITECTURE.md), and the obf↔name bridge
[`../NAMING.md`](../NAMING.md).

## Status (Milestone 1 ran 2026-06-01 — measured, not estimated)

> Full detail + evidence in [MILESTONE_1_RESULTS.md](MILESTONE_1_RESULTS.md).

- ✅ **rscplus builds and runs the patch path** — `BUILD SUCCESSFUL` under JDK 17, 230 classes,
  10.28 MB jar; the ASM patcher rewrites **71/71** jar classes cleanly. **Option B is banked.**
- ✅ **Protocol/net is *proven* faithful** — **48/48** crypto/protocol constants byte-match the jar
  (incl. the 1024-bit RSA modulus); ISAAC/Buffer/XTEA/RSA are **byte-identical** to two independent
  oracles **and** to the jar's own obfuscated classes by reflection; the net **decompile-diff is an
  empty semantic delta**. Two previously-feared bugs (XTEA `n7+`, RSA prefix width) measured as
  **non-bugs**.
- ❌ **The deob does NOT compile (Milestone 2 correction).** M1's "~7 drift sites" was an
  under-measurement: the real blocker is **347 member-name-drift pairs** (1,264 errors) — members
  renamed in their owning class but still called by old names elsewhere (same drift class as
  Mudclient's ~1,535 sites). M2 landed real progress (49 types + 1,024 members widened, dead-deps
  removed, `LoaderThread` rewired, all 7 named drift sites fixed, `Scanline.java` reconstructed) but
  the deob still doesn't compile. See [MILESTONE_2_RESULTS.md](MILESTONE_2_RESULTS.md).
- ✅ **Net handlers measured-faithful** — the per-opcode audit found ~62/70 MATCH and **5 catalogued
  divergences** (3 world-perception: D1 *missing* scenery handler, D2 ground-item gate, D3 boundary
  1-byte desync), all with exact jar fixes. See [NET_HANDLER_AUDIT.md](NET_HANDLER_AUDIT.md).
- 🔬 **Rendering fidelity: under audit now** — the `render-fidelity-bug-hunt` workflow is hunting the
  Go render lib vs the deob spec using the [render-diff tool](RENDER_DIFF_STATUS.md).

**Next gate (revised after M2):** (1) apply the 5 net fixes (D1→D3→D2→D4+D5); (2) **pivot the render
diff to GO↔JAR** (rscplus dump hook) — it needs no deob compile and gives Go-vs-ground-truth sooner;
(3) resolve the 347 member-drift pairs (generated obf→name member map) to unblock the DEOB engine +
the replay byte-cursor harness + the eventual Mudclient reconciliation.

## The situation in one picture

```
   rscplus (Java, builds clean @ JDK 17)                 westworld cradle (Go)
   ┌───────────────────────────────────┐                 ┌──────────────────────────────┐
   │ Launcher → JClassLoader            │                 │ proto/v235 (ISAAC/RSA/opcodes)│
   │   └ JClassPatcher.patch(byte[])    │                 │ world.World  (its perception) │
   │       ~433 ASM hooks, by OBF name  │                 │ render/  (PNG of world.World) │
   │   └ Reflection (48F/74M/3C, OBF)   │   talks to       │   └ -spectate / SnapshotFrom… │
   │ loads  ▼                           │   ─────────────▶ │ talks directly to OpenRSC svr │
   │   assets/rsclassic-…jar  ◀── THE SAME BYTES ──▶  reference/rsc-client/src (our deob) │
   └───────────────────────────────────┘                 └──────────────────────────────┘
            obfuscated, frozen, signed                       Go ports cite our Scene.java/
                                                             World.java by file:line  (G4)
```

**The load-bearing facts** (full evidence in `findings/`):

1. **rscplus and our deob are two views of the *exact same bytes*.** rscplus ships the same jar we
   deobfuscated and patches it via ASM keyed to *obfuscated* names (`ua`=Surface, `client`=Mudclient,
   …). Our `NAMING.md` is the verified bridge between the two. → `findings/02`.
2. **rscplus already builds and runs.** Verified `BUILD SUCCESSFUL` + a runnable 10 MB
   `dist/rscplus.jar` under **JDK 17** (it imports `java.applet`, removed in JDK 21+; this box's
   default `java` is JDK 26, and **Ant isn't installed**). → `findings/03`.
3. **The deob is *not* compilable today — and the blocker isn't the dead Microsoft deps.** Those are
   4 files with in-tree pure-Java/AWT twins (a day's work). The real cost is that **`Mudclient.java`
   (11,559 lines) is only half-renamed**: it still calls the renamed subsystems by obfuscated method
   names (`surface.a(...)` ×322; ~1,535 obf call sites total) and declares obf-typed fields
   (`ta npc`, `ba surface`). Reconciling that is the multi-week mountain. → `findings/01`.
4. **The cradle — which actually drives the testing goals — never touches rscplus or the jar.** It's a
   standalone **Go** client (`proto/v235` + `world.World`), and the Go `render/` lib is a line-by-line
   port of our `Scene.java`/`World.java`. For the render audit (G4) the deob's value is as a **textual
   spec**, needing *zero* compilation. → `findings/04`.
5. **"Deob = real client" is unverified.** There is no build, test, or byte-diff yet; the "236 bugs
   fixed" figure measures how buggy the first pass was, not what survives. Any option that *runs* the
   deob inherits this risk (login/protocol is the catastrophic surface). → `findings/05`.

## The goals we're serving

From the colleague's brief:

- **G1** — Fork rscplus; be able to **build and run** it; ideally re-point its patching/reflection at *our* code.
- **G2** — RSCPlus keeps functioning as normal (replay, config, hooks, overlays).
- **G3** — Use rscplus **replays** to show scenarios to the Claude-Code-driven cradle tester.
- **G4** — Audit the Go `render/` lib against our named client so it's a faithful 1:1 (fixes door/bridge/roof perception bugs).

## Recommendation (see [OPTIONS.md](OPTIONS.md) for the full case)

**Adopt Option D — a staged hybrid — with Option B as a permanent backbone.**

- **Option B** (fork + build rscplus *unchanged* against the obf jar; use the deob as a comprehension
  reference + the obf↔name bridge for the render audit) already delivers G1's practical meaning, all
  of G2, the G3 substrate, and G4's textual audit — at **LOW effort, LOW risk**. There's no reason to
  ever not have it running.
- **Option A** (the colleague's literal ideal: re-point rscplus's ~433 hooks + reflection at a
  recompiled deob) is real but is the **highest-cost, highest-risk** path. Its worst risks —
  bytecode-shape fragility, `COMPUTE_FRAMES`, silent descriptor mismatches, unverified fidelity —
  **can't be retired by effort, only by measurement.** And its main beneficiary (the cradle) doesn't
  use rscplus at all. So we *measure first*, mechanize the mapping, and only commit to A if the gates
  justify it. Option D is the safe road to A — and fails cheaply (with B still running) if A proves
  too fragile.

## Milestone 1 — DONE (2026-06-01). Results → [MILESTONE_1_RESULTS.md](MILESTONE_1_RESULTS.md)

The goal was to **bank a working client** and **answer two yes/no questions** before paying for
anything expensive. All three tasks ran (via the `rsc-client-milestone1-numbers` workflow):

- [x] **Stand up rscplus (B / Stage 0).** ✅ `BUILD SUCCESSFUL` under JDK 17 (via the bundled Ant
  launcher; Ant isn't installed as a binary). `dist/rscplus.jar` = 10.28 MB, `Main-Class:
  Client.Launcher`, 0 errors. Patch-probe: **71/71 jar classes patched cleanly.** Running still needs a
  display + a world `.ini` (or a replay dir for the offline `ReplayServer`). → `findings/milestone1/build.md`.
- [x] **Constant audit (Tier 0).** ✅ **48/48 constants byte-faithful, 0 mismatches** — RSA modulus
  (1024-bit, verified statically *and* by live clinit), ISAAC/XTEA shifts, the `0x9E3779B9` delta,
  framing thresholds, and 6 XOR string-pool round-trips. Located the RSA **exponent** in obf `s.c`
  (a decoy field). Confirmed the **RSA-key-override** (`JConfig.SERVER_RSA_MODULUS` →
  `ja.K`, `s.c`). → `findings/milestone1/constants.md`.
- [x] **"Does the deob compile?"** ⚠️ **0/8 packages — but the blocker is a mechanical
  flat-jar→subpackage *visibility split*, not obf drift.** 329 "not public" + ~753 cascades; after
  widening types to public, genuine drift = **7 signature sites**. `nativeapi` compiles clean after
  the dead-dep fixes. *Bonus:* protocol equivalence ran too — ISAAC/Buffer/XTEA/RSA all PASS vs two
  oracles **and** vs the jar by reflection, and the net **decompile-diff is empty**. →
  `findings/milestone1/compile-protocol.md`.

> **Reproducible artifacts:** the worktree at `westworld/.claude/worktrees/wf_1b683d6f-41a-3` holds the
> applied dead-dep fixes + the `q2kat/` (equivalence harness) and `diffwork/` (decompile-diff)
> workspaces. Every edit is enumerated in `findings/milestone1/compile-protocol.md` for replay into the
> main tree. (Worktrees may be GC'd — treat the report as the source of truth.)

**Revised gate (post-measurement):** the original "do subsystems link?" gate came back *architectural,
not semantic* — so the cheapest unblock is to **resolve the subpackage-visibility split** (collapse
toward the flat layout or widen members) + fix the **7 drift sites**, then run a **per-opcode
net-handler decompile-diff**. Defer the render-diff harness (design is ready in
[RENDER_DIFF_DESIGN.md](RENDER_DIFF_DESIGN.md)) and the Mudclient reconciliation until the deob compiles.

<details><summary>Original Week-1 plan (kept for provenance)</summary>

**Gate:** if the subsystems compile and the constants match → proceed to Stage 2 (mechanize the
obf→name+descriptor map and *generate* the rscplus remap from `NAMING.md` instead of ~300 hand-edits)
and Stage 3 (primitive equivalence tests via rscplus's `JClassLoader`). If not → we've spent <1 week
to learn the re-point is premature, and B is still serving every practical goal.

</details>

## Doc-fix action items (found during the investigation)

These are small correctness bugs in the *existing* deob docs that could mislead the work above. Fix
before anyone acts on them:

- [ ] **`../NAMING.md` lines 18–19** still say the World/Scene swap "needs to be done" — but the code
  is already corrected (commit `960310c`: `lb`=Scene, `k`=World). Update the prose so no one
  re-swaps correct files. → `findings/05` Caveat B, `findings/02` §3.
- [ ] **`../MUDCLIENT_SKELETON.md`** still has the **pre-swap** identities (`Hh=scene→k`,
  `Ek=world→lb`). Reconcile to `lb`=Scene, `k`=World. → `findings/02` §3.
- [ ] Record that rscplus's patch-routine **nicknames are misleading**: `patchData` operates on
  `m`=SocketFactory (`m.a([BBZ)V` = data/config decode) and `patchRandom` on `f`=RecordLoader. Any
  re-point must bind by **obf name + descriptor**, never by the nickname. → `findings/02` §3,
  `findings/05` Caveat A.
- [ ] Re-validate the Go `render/` line citations against the *current* `World.java`/`Scene.java`
  (post-swap) so they didn't pin to pre-swap line numbers. → `findings/04` §3.

## Reproduce / iterate

The workflow script that produced these findings is saved in the session at
`…/workflows/scripts/rsc-client-build-options-wf_cf9379aa-640.js` (re-runnable with `resumeFromRunId`
for cached results). Edit it to add investigation dimensions and re-run.
