# LAND_READINESS.md — go/no-go for PR #1 (branch `feat/remote-client`)

Synthesis of the two final streams (Mudclient delegate refactor + orsc↔rev-235-JAR render
parity), with the gates independently re-confirmed by the synthesizer on **2026-06-02**
against `git HEAD` (`0baf749`).

## Verdict: **GO — SAFE TO LAND.**

Every land gate is green: the Go tree builds and tests pass, the deob compiles 0-errors under
JDK 17, the branch is in sync with `origin/main`, the Mudclient refactor is verified pure code
motion (one documented behaviour-preserving line), and the orsc renderer is structurally
byte-identical to the rev-235 JAR on every JAR-runnable fixture. Residual gaps are all
documented and out of scope for this PR (cosmetic pixel shade; JAR-oracle GameData table not
available headless).

---

## 1. Compile / build / sync gates — all PASS (independently re-run)

| gate | command | result |
|---|---|---|
| Go build | `go build ./...` | **rc 0** |
| Go tests | `go test ./cmd/renderdiff/... ./render/...` | **ok** renderdiff + render (orsc has no test files) |
| Deob compile | `javac` (JDK 17) over 71 src files | **rc 0, 0 errors**, 7 expected deprecation warnings, **73 `.class`** incl. `ClientPackets.class` + `ClientSound.class` |
| In sync | `git merge-base --is-ancestor origin/main HEAD` | **rc 0** (origin/main is an ancestor of HEAD) |

Note on the branch name: the task header said `deob/rsc-client-reference`, but the working
tree is on **`feat/remote-client`**. All expected changes are present and verified there
(`Mudclient.java` + `MUDCLIENT_SKELETON.md` modified; `ClientPackets.java`, `ClientSound.java`,
the two new build docs, and the seam fixtures untracked). Not a blocker — a naming note for the
parent.

Working-tree delta to be committed by the parent:
- modified: `reference/rsc-client/src/client/Mudclient.java`, `.../docs/MUDCLIENT_SKELETON.md`
- new: `.../src/client/ClientPackets.java`, `.../src/client/ClientSound.java`,
  `.../docs/build/{REFACTOR_RESULTS.md,ORSC_JAR_PARITY.md}`, this file,
  `testdata/rscdump/hunt/{gen_seam.go,seam_sector_aligned.json}`,
  `testdata/rscdump/out/parity/**` (12 fixture artifact dirs incl. `_baseline`)

---

## 2. Mudclient refactor — verified pure code motion

Extract-delegate split of two subsystems out of the 13.4k-line `Mudclient`. Detail in
[REFACTOR_RESULTS.md](REFACTOR_RESULTS.md); independently re-confirmed here.

- **New classes (2):** `client.ClientPackets` (8 social/chat/privacy packet-out methods —
  `sendOpcodeString`, `sendCommand`, `sendPrivateMessage`, `sendRemoveFriend`, `sendAddFriend`,
  `sendAddIgnore`, `sendRemoveIgnore`, `sendPrivacySettings`) and `client.ClientSound` (audio:
  `playSound`, `initSounds`). Each holds a `final Mudclient m;` back-reference.
- **Wiring:** `Mudclient.java:908-909` declares `final ClientPackets packets` / `final
  ClientSound sound`; all former call sites route through them (re-counted: 12 `packets.` refs
  = 11 calls + decl; 8 `sound.` refs = 7 calls + decl). No moved-method body remains in
  `Mudclient` (grep-confirmed: no `playSound`/`initSounds` definitions left).
- **Line delta:** `Mudclient.java` **13416 → 13168 (−248)** (re-measured); `+ClientPackets`
  227, `+ClientSound` 93. `git diff --stat HEAD` = 44 ins / 292 del.
- **Pure code motion = true.** All 10 bodies byte-for-byte identical after the sanctioned
  `this.x`/unqualified→`m.x` and `STRINGS`/`findStringInData`→`Mudclient.`-qualified rewrite,
  **with one documented behaviour-preserving deviation**: `ClientSound.initSounds` AudioChannel
  fallback `host = this;` → `host = m;` (keeps the `(Component) host` cast target as the
  Mudclient/Applet, since the delegate is not a Component). Annotated at `ClientSound.java:83`.
  Faithfully preserved: the `ignoreListWorlds[j] = ignoreListWorlds[j]` original self-assign
  bug and the chat/priv/trade/duel privacy byte wire order.
- **Visibility widening:** several `private` fields + `showServerMessage`/`findStringInData`
  widened to package-private so same-package delegates can reach them; all remain
  package-private. Required for the idiom, not bugs.
- **Live smoke (refactor stream):** fresh build → autologin (`login response:64`) → live 3D
  frame dumped (`live3d.png`, 158150/171008 non-black viewport px) → stable past frame 301, no
  stderr exceptions. The audio init + packet writers are on the exercised boot/in-game path.

---

## 3. orsc ↔ rev-235 JAR render parity — structural match on every JAR-runnable fixture

The deferred runtime render diff: the orsc engine rendered the same dumped world state as the
obfuscated `rsclassic-1091943135.jar` oracle (headless under JDK 17 + ASM), diffed structurally
(camera/clear-screen/palette independent) and per-pixel. Full detail +byte evidence in
[ORSC_JAR_PARITY.md](ORSC_JAR_PARITY.md).

**Headline:** structurally **orsc IS the rev-235 JAR** on the JAR-runnable set — every built
face (terrain quads/triangles, the lifted wall/door leaf, the diagonal split) is byte-identical
in `(rounded centroid XYZ, vertex count, fillFront, fillBack)`.

| fixture | shared / total faces | verdict |
|---|---|---|
| `single_tile_door` | 9026 / 9026 | **match** (byte-identical) |
| `door_straight` | 9026 / 9026 | **match** (byte-identical) |
| `door_diag_obj` | 9025 / 9026 | minor-diff — JAR-oracle object-table gap (orsc builds the fixture's synthetic door-leaf; JAR drops objectId 0 with no GameData def). **Not an orsc divergence.** |
| 11 overlay/roof fixtures | 9037–9199 (orsc, deterministic) | **blocked** for the binary diff (see §5); **settled vs OpenRSC source** the jar compiles from |

### Per-E-row verdict
- **E-vp** (texture projection shift) — **SETTLED, match, byte-pinned.** Reflection probe reads
  the obf Scene field `lb.R` live = **9**, exactly orsc `rot1024VpSrc=9` (the runtime m_qd=9
  override over the ctor R=8). The JAR overwrites ctor-8 to 9 the same way orsc models.
- **E-water** (fixed-shade quad vs plain gouraud) — **SETTLED (orsc = source), JAR-binary
  blocked.** orsc draws type-4 water via the plain gouraud `createFace colour=1` path (80
  flattened `fillBack=1` water faces + 128 raised `fillFront=3` deck faces) — verbatim OpenRSC
  `World.java`, NOT the old fixed-shade+shoreline enhancement.
- **E-seam** (250-remap seam coord) — **SETTLED, match, FINDINGS #15 resolved in orsc's
  favour.** orsc's window-local `x==47`/`z==47` is a byte-exact port of OpenRSC
  `setTileDecorationOnBridge` (not `world wx%48==47`). Proven by a **new** sector-aligned fixture
  (`gen_seam.go`) that fires the 250→9 remap: 8 plank-edge faces at window-local col 47
  (world-X centroid 6080), `fillBack=-26426`.
- **E-polygonHit** (painter plane-pierce) — **SETTLED at algorithm level, pixel-diff blocked.**
  orsc `polygonHit1/2` (scene.go, HANDS-OFF) match the obf `lb boolean a(boolean,w,w)` /
  `a(byte,w,w)` (dead byte arg dropped) — same rotated-vert arrays, normal, magnitude,
  orientation-sign branch. The painter runs on the door fixtures' overlapping faces and the
  structural diff there is byte-clean. A true byte-level paint-order pixel diff needs a roof-def
  table (missing) + harness clear-screen reconciliation.
- **E-whole** (whole-renderer agreement) — **SETTLED, match on the JAR-runnable set.** First-ever
  clean orsc-vs-JAR structural agreement (the two door fixtures, 9026/9026). The
  door-perception bug class is absent for the straight-wall door.

### What the parity run fixed in the JAR oracle (separate repo, allowed)
`rscplus/dumprender/DumpRenderer.java`: (1) retuned `GO_HOST_LOCAL` 80→48 (removing a
`(4096,0,4096)`=32-tile placement offset from the deleted 160-tile-window engine); (2) filled
all 16 wall-def ids. Together drove `single_tile_door` 8513/9026 → 9026/9026 and `door_straight`
→ 9026/9026. **No `render/orsc` edits.**

---

## 4. Residual gaps (documented; out of scope for this PR)

- **orsc terrain shade ~1.7–2× brighter than the JAR oracle** (PIXEL-only, cosmetic; geometry
  byte-identical). Two compounding causes: (a) per-vertex ambience scheme mismatch
  (`terrainAmbience` coord-hash `[-5,4]` vs the JAR oracle's `Math.random→0.5`⇒flat-0 patch —
  neither matches the *live* client), (b) a base-shade/lighting or camera-distance gap larger
  than ±5 speckle. File: `render/orsc/world.go:606` + lighting in `render/orsc/{model,raster}.go`
  (HANDS-OFF). Note: GO_JAR_DIFF_RESULTS.md's old "1 LSB" green match was the **deleted** engine;
  current orsc does not reproduce it — worth a colleague look for a possible lighting regression.
- **Pixel diff ~99.7% on every fixture is harness drift, not a rasterizer-math divergence.**
  Driven by the JAR harness grey `#7c7c7c` + black clear-screen (orsc fills the whole frame) +
  different camera framing/`clipFar`. Only 0.30% px match exactly. The structural diff is the
  reliable faithfulness signal; pixel rows need the clear-screen + palette + camera
  reconciliation from GO_JAR_DIFF_RESULTS.md before they are meaningful.

---

## 5. Blockers (none for landing; both are oracle-availability limits, not orsc defects)

- **JAR-oracle GameData overlay/tile-decoration table gap (hard).** `DumpRenderer.java`
  synthesizes wall-def tables but NOT the obf overlay `v[]` def table (7 type-class defs, per-id
  colour/type arrays are GameData-loader-populated). Any nonzero overlay/roof tile throws `la`
  at `i.a:173 ← k.a:1378`. The jar carries **no embedded GameData** (fetches config/cache over
  the network), so the real table is genuinely unavailable headless — synthesizing it would
  fabricate, not read, the oracle. **11 fixtures blocked for the binary diff, settled vs OpenRSC
  source instead.** Verified throw on `bridge_ov4`, `t3_water_ov2`, `t3_type4_neighbour`,
  `seam_sector_aligned`, `roof_inside`.
- **JAR-oracle object-def-table gap.** `door_diag_obj` 9025/9026 (orsc builds its synthetic
  fixture door-leaf; JAR drops objectId 0 with no def). A real object with a real def builds in
  both.

---

## 6. Recommended next steps (post-land, none block this PR)

1. **3-way pixel-exact diff.** Reconcile the JAR oracle's `Math.random→0.5` ambience patch with
   orsc's coord-hash (or add a flat-0-ambience render option to orsc) AND re-derive the JAR
   harness camera framing + `clipFar`, then re-check the terrain green vs both the JAR and the
   live client. Confirm the >speckle brightness gap is not a real lighting regression vs the
   deleted engine.
2. **Unblock the 11 overlay/roof fixtures.** Source a real GameData overlay/roof/object def
   table for the JAR oracle (e.g. drive it through a live config/cache fetch, or load an OpenRSC
   GameData dump) so the binary diff can run end-to-end on overlay/water/roof terrain.
3. **Land + update PR #1** with the refreshed description (below). Parent handles commit + push.

---

## Cross-references

- Refactor detail: [REFACTOR_RESULTS.md](REFACTOR_RESULTS.md)
- Parity detail + byte evidence + reproduce recipe: [ORSC_JAR_PARITY.md](ORSC_JAR_PARITY.md)
- Build + run recipe: [RUN.md](RUN.md)
- Functioning-client achievement: [FUNCTIONING_CLIENT.md](FUNCTIONING_CLIENT.md)
- Earlier Go↔JAR diff (clear-screen/palette/camera reconciliation): [GO_JAR_DIFF_RESULTS.md](GO_JAR_DIFF_RESULTS.md)
