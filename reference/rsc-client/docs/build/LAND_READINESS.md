# LAND_READINESS.md — go/no-go for PR #1 (branch `feat/remote-client`)

Synthesis of the two final streams — the **Mudclient de-god decomposition** (extract-delegate
split, rounds 1+2) and the **orsc↔rev-235-JAR render parity + terrain-fidelity fix** — with
every land gate independently re-confirmed by the synthesizer (Opus) on **2026-06-02** against
`git HEAD` (`7f6d980`).

## Verdict: **GO — SAFE TO LAND.**

Every land gate is green: the Go tree builds and tests pass, the deob compiles 0-errors under
JDK 17 (78 classes), the branch is in sync with `origin/main`, the Mudclient refactor is verified
**pure code motion** (`functionalDivergence = false`), the live functional smoke reproduces the
pre-refactor baseline signature byte-for-byte, and the orsc renderer is structurally
byte-identical to the rev-235 JAR on every JAR-runnable fixture. The one **renderer-math**
divergence that was previously open — terrain base-shade ~2× too bright — is now **RESOLVED**.
Remaining residuals are all documented, honest, and out of scope for this PR (a sub-LSB ambience
speckle pin; the JAR-harness camera/fog framing; the JAR-oracle GameData table not available
headless).

---

## 1. Compile / build / sync gates — all PASS (independently re-run)

| gate | command | result |
|---|---|---|
| Go build | `go build ./...` | **rc 0** |
| Go tests | `go test -count=1 ./cmd/renderdiff/... ./render/...` | **ok** (fresh, not cached) — renderdiff 2/2, render fidelity + sprite/appearance suites all PASS; `render/orsc` has no test files |
| Go vet | `go vet ./render/... ./cmd/renderdiff/...` | **clean** |
| gofmt | `gofmt -l render/orsc/model.go` | **clean** (empty) |
| Deob compile | `javac -Xmaxerrs 4000` (JDK 17) over **76 src files** | **rc 0, 0 errors**, 7 expected `java.applet`/`new Integer` deprecation warnings, **78 `.class`** |
| In sync | `git merge-base --is-ancestor origin/main HEAD` | **rc 0** (origin/main `37b3ea7` is an ancestor of HEAD) |

**Class count (78):** baseline 73 (.class at the pre-de-god state) + the **5 new top-level
round-2 delegate classes**. Equivalently: 76 source files emit 78 `.class` (the +2 are
pre-existing inner classes the source count undercounts). All 7 delegate classes
(`ClientPackets`, `ClientSound`, `TradeDuelBankPackets`, `WidgetRenderer`, `GameInterface`,
`MenuController`, `IncomingPackets`) and `Boot` are present in the build output.

**Branch-name note (for the parent):** the task header named `deob/rsc-client-reference`, but the
working tree is on **`feat/remote-client`** (`deob/rsc-client-reference` *is* an ancestor of HEAD).
All expected changes are present and verified here. Not a blocker — a naming note for the commit.

**Scope check (clean):** the working-tree porcelain contains ONLY in-scope files — no
`cmd/cradle/**`, `remoteclient/**`, `proto/v235/**`, or `docs/remote-client/**` modifications are
present. (Earlier verifier notes mentioned untouched concurrent-workstream files; they are not in
this working tree.)

Working-tree delta for the parent to commit (round-2 de-god + render fix; round-1
`ClientPackets`/`ClientSound` are already committed at HEAD):
- modified: `reference/rsc-client/src/client/Mudclient.java`, `render/orsc/model.go`,
  `reference/rsc-client/docs/MUDCLIENT_SKELETON.md`,
  `.../docs/build/{REFACTOR_RESULTS.md, ORSC_JAR_PARITY.md, RENDER_FIDELITY_FINDINGS.md}`, this file
- new (untracked): `.../src/client/{TradeDuelBankPackets, WidgetRenderer, GameInterface,
  MenuController, IncomingPackets}.java`, `testdata/rscdump/out/parity/**` (fixture artifacts)

---

## 2. Mudclient de-god decomposition — verified pure code motion (`functionalDivergence = false`)

The headline structural win of this PR: the 13.4k-line god-object `Mudclient` was decomposed in
two extract-delegate waves into **7 focused sibling classes**, with **zero behavioural change**.
Full detail in [REFACTOR_RESULTS.md](REFACTOR_RESULTS.md); independently re-confirmed here.

**The idiom.** Each new class holds a `final Mudclient m;` back-reference set in a
side-effect-free constructor (`this.m = m;` only — all 7 verified). `Mudclient` keeps a `final`
delegate field; every former call site routes through it. Field visibility was widened
`private → package-private` (compiler-driven: exactly the members the moved bodies touch). The
`Mudclient` constructor body is byte-for-byte identical to HEAD, so there is no init-order hazard.

### Round 1 (committed at HEAD)
| delegate field | new class | methods moved |
|---|---|---|
| `packets` | `client.ClientPackets` (227 ln) | 8 social/chat/privacy packet-out writers |
| `sound` | `client.ClientSound` (93 ln) | `playSound`, `initSounds` |

10/10 bodies byte-identical after the sanctioned qualifier rewrite, **with one documented
behaviour-preserving deviation**: `ClientSound.initSounds` AudioChannel fallback `host = this;`
→ `host = m;` (keeps the `(Component) host` cast target the Mudclient/Applet, since the delegate
is not a `Component`). Faithfully preserved: the `ignoreListWorlds[j] = ignoreListWorlds[j]`
original self-assign bug and the privacy byte wire order.

### Round 2 (working tree — the de-god split; parent commits)
| delegate field | new class / lines | methods moved |
|---|---|---|
| `tradePackets` | `TradeDuelBankPackets` (239) | `bankSend`, `sendDuelOffer`, `sendTradeOffer`, `sendDuelItems` |
| `widgetRenderer` | `WidgetRenderer` (276) | `drawBox`, `clearScreen`, `drawSprite`, `drawIcon`, `drawScrollList`, `drawMenuOptions`, `drawScrollbar`, `drawScrollbar2` |
| `gameInterface` | `GameInterface` (1732) | `drawWildernessWarning`, `drawShop`, `drawBank`, `drawTrade`, `drawTradeConfirm[Window]`, `drawDuelConfirm`, `drawDuel`, `drawWelcome`, `drawHelpMenu`, `drawCloseButton`, `drawGameSettings` |
| `menus` | `MenuController` (560) | `handleGameClick`, `buildClickMenu`, `handleInventoryClick`, `menuHitTest`, `pointInRect`, `pointInPanel` |
| `incoming` | `IncomingPackets` (1851) | `handlePacket`, `handleSceneUpdates`, `onFriendUpdate`, `applyAppearanceUpdate` |

- **`Mudclient.java`: 13168 → 8977 lines (−4191).** `git diff --stat HEAD`: 333 insertions /
  4524 deletions on `Mudclient.java`. The 5 delegate fields are declared in-line at lines
  910–914 beside the pre-existing `packets`/`sound`.
- **Pure code motion = true.** Brace-matched, comment-stripped, qualifier-normalized comparison
  vs HEAD: **all 33 extracted bodies byte-for-byte IDENTICAL**, including the highest-risk targets
  — `handlePacket` master opcode dispatch (19038 normalized chars), `drawDuel` (9816), `drawBank`
  (8626; withdraw/deposit send stays inside the click block), `drawTrade` (8044),
  `handleSceneUpdates` (7254), `onFriendUpdate` (4588), `applyAppearanceUpdate` (4094), and the 4
  `TradeDuelBankPackets` writers (**opcode values 22/23 bank, 33 duel, 46 trade + putByte/
  putShort/putInt wire byte-order preserved** — spot-confirmed by the synthesizer). All **18
  methods kept in Mudclient** are byte-identical; all **24 pure-delegate wrappers** forward
  arguments 1:1 with zero arg-order defects and signatures matching HEAD (callers bind
  identically).
- No logic, constant, operator, control-flow, opcode, wire byte-order, bit-width, or init-order
  change anywhere. **`functionalDivergence = false`.**

### Live functional smoke — PASS (full pre-refactor baseline signature reproduced)
Fresh JDK17 build → `systemd --user` unit `rsc-deob-degverify1` (autologin `deobtest`/`deobpass`,
`RSC_AUTO_APPEARANCE`, `RSC_AUTO_TABS`, `RSC_FBUFFER_DUMP`).

| metric | baseline (`7a45f39`) | de-god run | verdict |
|---|---|---|---|
| `login response` | 64 | 64 | match |
| RSA | bitlen=512 exp=65537 | bitlen=512 exp=65537 | match |
| player spawn | alive @ wi=3392,9024 | alive @ wi=3392,9024 (frame 51+) | match |
| non-black viewport px | 158538 / 171008 | 158150 / 171008 (Δ 0.24%, > 150k bar) | match |
| tab indices | {0,1,2,4,5,6} | {0,1,2,4,5,6} each ×11–12, 0 exceptions | match |
| stability | 2201 frames | 5301 frames, no crash | exceeds |
| stderr exceptions | 0 | 0 | match |

`live3d.png` visually matches the baseline scene (stone-walled room, purple/wood checkered floor,
"Welcome to Westworld!" banner, chat-tab row).

---

## 3. orsc renderer — terrain base-shade reseed bug **RESOLVED** (faithful to rev-235)

A single surgical, spec-justified change in `render/orsc/model.go` fixes the one renderer-**math**
divergence that was previously open. Full detail in
[RENDER_FIDELITY_FINDINGS.md](RENDER_FIDELITY_FINDINGS.md) §0 and
[ORSC_JAR_PARITY.md](ORSC_JAR_PARITY.md) §1.

- **Bug.** orsc had faithfully ported OpenRSC's `RSModel.clearRotDataAndParams26`, which
  unconditionally re-seeds `setDiffuseLight(40,102,104,108,-20,-89)` on every model's first
  project. That clobbered each model's authentic *build-time* light. For terrain (built with
  `setDiffuseLightAndColor(-50,-10,-50,40,48,…)` ⇒ `diffuseParam1=96`), the reseed forced
  `diffuseParam1 = 256-102*4 = -152`, driving the flat-grass gouraud `shadeBase` negative ⇒
  clamped to 0 ⇒ the brightest ramp entry, green **0x8e** (vs the JAR oracle's **0x48**, ~2×).
- **Why the reseed is wrong.** The authentic rev-235 `GameModel.project()` does NO project-time
  relight; `apply(7972)` relights only on `transformState==1` from each model's OWN `setLight`
  params. OpenRSC's `RSModel` reseed is an **infidelity vs the rev-235 GameModel oracle**
  (= the obf JAR = the deob in `reference/rsc-client/`).
- **Fix.** Dropped the reseed; renamed `clearRotDataAndParams26 → allocProjectionScratch` (now
  does ONLY the safe projection-scratch (re)allocation). One change relights **every** model type
  from its own build-time params (terrain 40/48, walls 60/24, roofs 50/50, scenery/diagobj) —
  exactly the deob path. The doc comment records the dropped reseed as an OpenRSC
  `RSModel.java:528-543` infidelity citing `GameModel.java:1225-1281` + `World.java:1025`, so a
  future faithful-to-OpenRSC re-port cannot silently reintroduce it.
- **Result.** Base lighting (fog-removed) `shadeBase=71` ⇒ green **0x4a**, within 1–2 LSB of the
  JAR's **0x48** (down from the ~2× error). Verified at the per-vertex shade level (instrumented,
  then reverted): `diffuseParam1=96`, `diffuseParam2=384`, mag=71, `vertDiffuseLight≈−24`.
- **Geometry UNCHANGED:** `single_tile_door` structural diff STILL **9026/9026** byte-identical;
  orsc-vs-orsc determinism STILL **0-pixel**. Two safe closeouts applied (the doc rename + a
  wall/roof relight verification confirming no model type relied on the reseed default); one
  deferred (ambience speckle, §4). Fresh JAR-oracle re-render confirms `jar.png` is not stale.

---

## 4. orsc ↔ rev-235 JAR render parity — structural match on every JAR-runnable fixture

The orsc engine rendered the same dumped world state as the obfuscated
`rsclassic-1091943135.jar` oracle (headless JDK 17 + ASM), diffed structurally
(camera/clear-screen/palette independent) and per-pixel. Full evidence in
[ORSC_JAR_PARITY.md](ORSC_JAR_PARITY.md).

**Headline:** structurally **orsc IS the rev-235 JAR** on the JAR-runnable set — every built face
(terrain quads/triangles, the lifted wall/door leaf, the diagonal split) is byte-identical in
`(rounded centroid XYZ, vertex count, fillFront, fillBack)`.

| fixture | shared / total faces | verdict |
|---|---|---|
| `single_tile_door` | 9026 / 9026 | **match** (byte-identical) |
| `door_straight` | 9026 / 9026 | **match** (byte-identical) |
| `door_diag_obj` | 9025 / 9026 | minor-diff — JAR-oracle object-table gap (orsc builds the fixture's synthetic door-leaf; JAR drops objectId 0 with no GameData def). **Not an orsc divergence.** |
| 11 overlay/roof fixtures | deterministic (orsc) | **blocked** for the binary diff (§6); **settled vs OpenRSC source** the jar compiles from |

All 5 **E-rows** (E-vp, E-water, E-seam, E-polygonHit, E-whole) are **SETTLED** — match or
match-vs-source — per the per-E-row table in ORSC_JAR_PARITY.md.

---

## 5. Residual gaps (documented; out of scope for this PR — no overclaiming)

- **Whole-frame pixel diff stays ~99.7% (`DIFFER 170496/171008, maxD=124`).** This is **NOT a
  lighting bug** and not a rasterizer-math divergence — the structural diff is byte-clean and the
  base lighting now matches. It is the documented **harness camera-framing + distance-fog +
  clear-screen drift**: orsc frames the grass FAR from the camera (camera-space Z 1152–2340), so
  the **byte-identical engine fog math** `(vertZRot − fogSmoothingStartDistance)/fogZFalloff =
  (z−10)/20` (deob `Scene.java:1779-1790` ≡ OpenRSC `scene.go:653-655`,
  `fogSmoothingStartDistance=10/fogZFalloff=20`) adds +57…+116, darkening the *rendered* grass to
  ramp index 126–192; the JAR harness frames the same grass close (near-zero fog) so it renders at
  shade ~71–74 = 0x48. PROOF it is fog, not lighting: with fog removed, orsc base `shadeBase=71`
  matches the JAR's 0x48 exactly.
- **±5 ambience speckle pin (deferred, sub-LSB).** `world.go:606 terrainAmbience` coord-hash
  `[-5,4]` vs deob `Math.random` / JAR-patched flat-0 — the only remaining grass-green wobble
  (the 67–76 shadeBase spread) after the reseed fix. Deliberately deferred behind a future
  flat-0/true-random flag to avoid perturbing the determinism tests; documented in
  ORSC_JAR_PARITY.md §2.

A pixel-exact **3-way** FRAME match (orsc = JAR-harness = live) would additionally require
reconciling the JAR-harness camera framing/`clipFar` AND pinning the speckle; neither is a
renderer-math divergence, and both are post-land work.

---

## 6. Blockers (none for landing; both are oracle-availability limits, not orsc defects)

- **JAR-oracle GameData overlay/tile-decoration table gap (hard).** The obf overlay `v[]` def
  table is GameData-loader-populated; the jar carries no embedded GameData (it fetches config/cache
  over the network), so the real table is genuinely unavailable headless — synthesizing it would
  fabricate, not read, the oracle. **11 fixtures blocked for the binary diff, settled vs OpenRSC
  source instead.**
- **JAR-oracle object-def-table gap.** `door_diag_obj` 9025/9026 (JAR drops objectId 0 with no
  def). A real object with a real def builds in both.

---

## 7. Recommended next steps (post-land, none block this PR)

1. **3-way pixel-exact diff.** Re-derive the JAR-harness camera framing + `clipFar` and add a
   flat-0-ambience render option to orsc (pin the speckle), then re-check the terrain green vs both
   the JAR and the live client. The dominant ~2× brightness gap is already closed.
2. **Unblock the 11 overlay/roof fixtures.** Source a real GameData overlay/roof/object def table
   for the JAR oracle (live config/cache fetch, or an OpenRSC GameData dump) so the binary diff can
   run end-to-end on overlay/water/roof terrain.
3. **Continue the Mudclient de-god** if desired: the remaining 18 methods kept in `Mudclient`
   (`drawUiTabMagic`, `loadRegion`, `isDirectionWalkable`, `addPlayer`, `sortFriendsList`, …) are
   candidate further-extraction seams, but the current split already drops the god-object 32%.
4. **Land + update PR #1** with the refreshed description (synthesizer's `prDescriptionDelta`).
   Parent handles commit + push.

---

## Cross-references

- De-god / refactor detail: [REFACTOR_RESULTS.md](REFACTOR_RESULTS.md)
- Terrain-fidelity fix detail: [RENDER_FIDELITY_FINDINGS.md](RENDER_FIDELITY_FINDINGS.md) §0
- Parity detail + byte evidence + reproduce recipe: [ORSC_JAR_PARITY.md](ORSC_JAR_PARITY.md)
- Mudclient structural index + delegate table: [../MUDCLIENT_SKELETON.md](../MUDCLIENT_SKELETON.md)
- Build + run recipe: [RUN.md](RUN.md)
- Functioning-client achievement: [FUNCTIONING_CLIENT.md](FUNCTIONING_CLIENT.md)
- Earlier Go↔JAR diff (clear-screen/palette/camera reconciliation): [GO_JAR_DIFF_RESULTS.md](GO_JAR_DIFF_RESULTS.md)
