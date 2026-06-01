# Milestone 2 — measured results (compile gate + net-handler depth)

Ran 2026-06-01 (`deob-compile-and-net-depth` workflow). Two outcomes: one hard
correction (the compile gate did **not** close) and one high-value result (the net surface is now
*measured-faithful at the handler layer*, with 5 catalogued bugs). Raw reports:
[findings/milestone2/](findings/milestone2/) and the full audit in
[NET_HANDLER_AUDIT.md](NET_HANDLER_AUDIT.md).

## 1. Compile gate — FAIL (and M1 under-measured)

Milestone 1 estimated the deob was ~7 drift sites from compiling. **That was wrong**, and we now know
why: M1 only widened top-level *types* and never added the cross-package *imports*, so a large amount
of drift stayed hidden inside the `cannot find symbol` cascade M1 dismissed as visibility noise.

The real picture, measured by a full subsystem compile (JDK 17, Mudclient excluded):

| Metric | Value |
|---|---|
| Subsystems compile | **FAIL** |
| Residual errors | **1,264** subsystem-internal (1,281 total − 17 Mudclient-dependent) |
| Root cause | **347 distinct (owner-type, member) member-name-drift pairs** (1,191 `cannot find symbol`) |
| Per package | scene 376, net 293, world 170, audio 135, ui 116, shell 90, util 76, data 32 |

**The drift is pervasive cross-class member-name drift:** members were renamed in their *owning* deob
class but are still called by their old obf/short names *elsewhere* — e.g. `GameModel.e(...)` (41×),
`ISAAC.emptyString(...)` (22×), `Surface.getLineHeight/getStringWidth/drawString`, the
`BZip→DecodeBuffer` 36-field state cluster. This is the **same drift class** as `Mudclient.java`'s
~1,535 obf call sites — i.e. the deob was renamed per-class without propagating renames across
references. Resolving the 347 pairs is real per-member reconciliation against ground truth (the clean
decompile + the per-member `// obf:` map), not a mechanical sweep.

**What did land (real progress, uncommitted in the working tree):** 4 dead-Microsoft-dep files
deleted; `JSBridge` stubbed; `LoaderThread` rewired from reflection to typed calls (drifted method
names fixed); **all 7 named drift sites fixed** against jar ground truth (the `CacheFile.getUnsignedShort`
arg reorder, `CachePath.resolveFile/initialize/resolveOrCreateCacheFile` renames, `JSBridge.call`
arity, `Surface.blackScreen`, the `StreamFactory` package+arg drift); a **missing `Scanline.java`**
(obf `n`) reconstructed from the clean decompile; **49 types + 1,024 members widened** (package
structure preserved); ~97 cross-package imports added. `Mudclient.java` untouched.

### Strategic implication

The **DEOB render engine is now gated behind this 347-pair reconciliation** — so the first real
**GO↔DEOB** render diff is further out than Milestone 1 implied.

**Pivot:** the **JAR oracle does not need the deob to compile.** rscplus builds clean and drives the
obf jar directly, so a **GO↔JAR** diff — Go vs the *true* ground truth — is reachable *now* via a
rscplus dump hook (RENDER_DIFF_DESIGN §3), without waiting on the deob compile. GO↔JAR is the more
valuable diff anyway (the jar is the authoritative oracle; the deob is a transcription of it). So the
render-diff effort should prioritize the **JAR engine over the DEOB engine**.

## 2. Net-handler depth audit — PASS-with-findings (full detail: [NET_HANDLER_AUDIT.md](NET_HANDLER_AUDIT.md))

This is the per-opcode handler diff M1 named as the gate to extend the proven net result from
primitives up to the packet handlers. ~70 incoming opcode/handler units diffed (deob `Mudclient.java`
vs the jar's clean decompile, cross-checked against OpenRSC Payload235 + LeadingBot/mudclient204):

- **~62 MATCH** (byte-faithful), **4 MINOR** (cosmetic / behaviourally-equivalent), **5 DIVERGENCE**.
- Verdict: **the net surface upgrades from M1's "HIGH — proven (primitives)" to "HIGH — proven
  primitives + measured-faithful handlers, with 5 catalogued, fixable divergences."** It is no longer
  an estimate — but the handler layer is *not yet* an empty semantic delta like the primitives are.

### The 5 divergences (all are deob transcription bugs with exact jar fixes — `[WP]` = world-perception)

| ID | Sev | Opcode | Bug | Fix |
|----|-----|--------|-----|-----|
| **D1** `[WP]` | **CRITICAL** | 48 SCENERY | **Entire handler MISSING** from `Mudclient.handlePacket` (jar encodes it as `if(-49==~var1)`, so a `==48` grep missed it). Scenery (trees/signs/fences/door-objects) never appears or updates mid-session. | Transcribe jar `client.java:14510-14659` between the 99 and 111 branches |
| **D2** `[WP]` | HIGH | 91 GROUND-ITEM | Add gate `if(!placed)` instead of `if(itemId != 65535)` → re-sent items vanish; `65535` sentinel spawns a phantom item | `Mudclient.java:4433` → `if(itemId != 65535)` |
| **D3** `[WP]` | HIGH | 99 BOUNDARY | Spurious `mg.w--` un-reads the `0xFF` removal marker → **1-byte desync corrupts every subsequent wall/door add/remove in the packet** (the door-perception class) | delete `mg.w--` at `Mudclient.java:4188` |
| **D4** | MED | 137 SHOP-CLOSE | Wrong body (copy of opcode 15): reads 1 phantom byte (over-read) + sets wrong flag; shop never closes | `Mudclient.java:4742-4746` → `uk=false; return;` |
| **D5** | LOW-MED | 234.6 / 101 | Self-speech reads a string before the null-check (desyncs the 234 packet); opcode-101 extra `uk=false` inverts the shop-list lifecycle | guard the read; delete extra `uk=false` |

**Two non-bugs worth recording:** the jar binds **99=boundary, 91=ground-item** (the *reverse* of
OpenRSC's enum labels) — the deob correctly follows the jar; only doc comments carry OpenRSC labels.
And many trade/duel/shop comment labels are swapped vs OpenRSC names while the code writes the correct
fields.

**Relevance caveat:** these are deob *client* handler bugs. The westworld **cradle uses its own Go
`proto/v235`** (bit-for-bit with OpenRSC), so the live cradle is *not directly* affected — but (a)
they must be fixed for the deob to be a faithful client / replay-diff oracle, and (b) D1/D2/D3 are
worth spot-checking in the Go `world.Apply` path, since they are exactly the scenery/boundary/item
state the renderer draws.

### What still needs end-to-end replay/live traffic to fully close

The 5 are found by static decompile-diff; D3/D4/D5(a) are *cursor-desync* bugs whose end-to-end blast
radius needs a **replay byte-cursor diff** (drive fixed-deob `handlePacket` and jar `client.b(...)`
over an identical captured stream; assert identical post-packet offsets + world-state). The narrow
branches (234.6, 91's two modes) must be fuzzed/chosen in the corpus. The **outgoing** opcode bodies
were not field-diffed (incoming only). All of this is gated behind the compile fix.

## 3. Updated next steps

1. **Apply the 5 net fixes** to `Mudclient.java` (D1→D3→D2→D4+D5b→D5a) — high value, exact jar refs;
   D1/D2/D3 are the world-perception fixes any deob-as-client / replay-diff depends on.
2. **Resolve the 347 member-drift pairs** to close the compile gate — the prerequisite for the DEOB
   render engine, the replay byte-cursor harness, and ultimately the Mudclient reconciliation (same
   drift class). Tractable but substantial; best done as a generated obf→name member map applied to
   call sites.
3. **Pivot the render diff to GO↔JAR first** (rscplus dump hook) — it doesn't need the deob to
   compile and gives Go-vs-ground-truth sooner. DEOB engine follows once (2) lands.
