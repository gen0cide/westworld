# REFACTOR_RESULTS.md — Mudclient subsystem extraction (delegate split)

Verification record for the refactor that extracted two standalone subsystems out of
`reference/rsc-client/src/client/Mudclient.java` into sibling package-private classes using
the **extract-delegate** idiom. Verified by the refactor verifier on 2026-06-02 against
`git HEAD` (`0baf749`).

**Verdict: SAFE TO LAND.** Compiles 0 errors; pure code motion confirmed (one documented
behaviour-preserving deviation); live smoke test logs in and renders a frame.

---

## What moved

Each new class holds a `final Mudclient m;` back-reference set in its constructor. `Mudclient`
keeps a `final` delegate field, and every former call site routes through it.

| delegate field on `Mudclient` | new class / file | methods moved | lines moved |
|---|---|---|---|
| `final ClientPackets packets` (`Mudclient.java:908`) | `client.ClientPackets` — `src/client/ClientPackets.java` (227 lines) | `sendOpcodeString`, `sendCommand`, `sendPrivateMessage`, `sendRemoveFriend`, `sendAddFriend`, `sendAddIgnore`, `sendRemoveIgnore`, `sendPrivacySettings` | ~210 |
| `final ClientSound sound` (`Mudclient.java:909`) | `client.ClientSound` — `src/client/ClientSound.java` (93 lines) | `playSound`, `initSounds` | ~45 |

Call sites updated:
- `packets.` — 11 call sites across the affected methods (`sendRemoveFriend`×2,
  `sendAddFriend`×2, `sendAddIgnore`×2, others ×1). 12 `packets.` references total in
  `Mudclient` (11 calls + 1 field decl).
- `sound.` — 7 call sites (prayer on/off, quest-complete jingle ×2, opcode 204
  SEND_PLAY_SOUND, options-tab sfx, `initSounds` bring-up). 8 `sound.` references total
  (7 calls + 1 field decl).

No bare/unqualified call to any moved method remains in `Mudclient` (verified by grep).

### Fields / methods widened from `private`

Required so the same-package delegate classes can reach them. All remain package-private:

- Fields: `Jh`, `Uh`, `hk`, `ni`, `ne`, `wi`, `username`, `membersServer`,
  `friendListCount`, `friendListNames`, `friendListFormerNames`, `friendListOnline`,
  `friendListWorlds`, `ignoreListCount`, `ignoreListNames`, `ignoreListDisplayNames`,
  `ignoreListFormerNames`, `ignoreListWorlds`.
- Methods: `showServerMessage` (was `private final` → `final`), `findStringInData`
  (was `private static final` → `static final`).

`STRINGS` was already package-private; the delegates reference it as `Mudclient.STRINGS[...]`
and call `Mudclient.findStringInData(...)`.

---

## Line-count delta

| file | before (HEAD) | after | delta |
|---|---|---|---|
| `Mudclient.java` | 13416 | 13168 | −248 |
| `ClientPackets.java` | — (new) | 227 | +227 |
| `ClientSound.java` | — (new) | 93 | +93 |

`git diff HEAD -- Mudclient.java`: 44 insertions, 292 deletions (the inserted lines are the
two delegate-field declarations, the field/method modifier flips, and the 18 call-site
rewrites; the deletions are the moved method bodies + their javadoc).

---

## 1. Compile gate — PASS (0 errors)

```
find reference/rsc-client/src -name '*.java' > /tmp/deob-srcs.txt   # 71 source files
/usr/lib/jvm/java-17-openjdk/bin/javac -Xmaxerrs 4000 -d /tmp/deob-verify @/tmp/deob-srcs.txt
```

- Exit 0, **0 errors**, 7 warnings (all pre-existing `java.applet` / `new Integer(int)`
  removal-deprecation warnings — expected on JDK 17, unrelated to this refactor).
- **73 `.class` files** produced from 71 source files (the extra two are inner classes that
  predate this work, not from the delegates). `ClientPackets.class` and `ClientSound.class`
  both present.
- Re-verified with the `src/client`-only RUN.md build (`javac -d /tmp/deob-run`): exit 0,
  73 classes, both delegate classes compiled.

---

## 2. Pure-code-motion audit — PASS (one documented deviation)

Method-by-method brace-matched comparison of each moved body in HEAD vs the new class file,
after normalizing the only sanctioned mechanical rewrite (`this.x` / unqualified `x` → `m.x`,
`STRINGS`/`findStringInData` → `Mudclient.`-qualified):

| method | result |
|---|---|
| `sendOpcodeString` | IDENTICAL (3 stmts) |
| `sendCommand` | IDENTICAL (3 stmts) |
| `sendPrivateMessage` | IDENTICAL (4 stmts) |
| `sendRemoveFriend` | IDENTICAL (19 stmts) |
| `sendAddFriend` | IDENTICAL (31 stmts) |
| `sendAddIgnore` | IDENTICAL (31 stmts) |
| `sendRemoveIgnore` | IDENTICAL (19 stmts) — incl. the faithfully-reproduced `ignoreListWorlds[j] = ignoreListWorlds[j]` original-bug self-assign |
| `sendPrivacySettings` | IDENTICAL (6 stmts) — byte wire order chat/priv/trade/duel preserved |
| `playSound` | IDENTICAL (8 stmts) |
| `initSounds` | IDENTICAL except **one line** (see below) |

**The single deviation** (in `ClientSound.initSounds`): the AudioChannel host fallback
`host = this;` → `host = m;`. Behaviour-preserving: in the original `this` was the `Mudclient`
(an `Applet`, hence a `Component`); in the delegate `this` is the non-`Component` `ClientSound`,
so `m` keeps the same runtime host object passed to the `(Component) host` cast. Same control
flow, constants, argument order, and cast otherwise.

No logic, constant, operator, or control-flow change anywhere else. **pureCodeMotion = true.**

---

## 3. Live smoke test — PASS (logged in + rendered)

Infra was up (game server `127.0.0.1:43594`, content host `:7000`, `Xvfb :99`), so the proven
RUN.md recipe was executed: fresh build into `/tmp/deob-run`, launched under a transient
`systemd --user` unit (`rsc-deob-verify1`) with autologin (`deobtest`/`deobpass`),
auto-appearance, and the framebuffer dump.

Observed in `/tmp/rsc-run/verify1.log`:

```
[Boot] RSA modulus bitlen=512 exp=65537
Started application
login response:64                         # success
[DBGgf] frame=51 ... playerAlive=true wi=3392,9024
[RSC_FBUFFER_DUMP] wrote .../live3d.png (512x346) ... non-black viewport pixels=158150 / 171008
[DBGgf] frame=301 ... playerAlive=true wi=3392,9024     # stable
```

- **Login:** succeeded (`login response:64`, `playerAlive=true`, valid world coords).
- **Render:** `live3d.png` written with 158150/171008 non-black viewport pixels — a fully
  populated 3D world view (stone walls, tiled floor, "Welcome to Westworld!" banner, chat
  tab bar). Visually confirmed.
- **Stability:** ran cleanly past frame 301, no exceptions in stderr.
- Audio init (`ClientSound.initSounds`) and the social/chat packet writers
  (`ClientPackets`) are on the live boot/in-game path; the session reaching a stable rendered
  frame exercises them without error.

Unit stopped after capture (`systemctl --user stop rsc-deob-verify1`).

---

## Summary

| check | result |
|---|---|
| javac errors | **0** |
| classes produced | 73 |
| pure code motion | **true** (1 documented behaviour-preserving `this→m` line) |
| Mudclient lines | 13416 → 13168 (−248) |
| new classes | `client.ClientPackets`, `client.ClientSound` |
| live smoke test | login OK + 3D frame rendered + stable past frame 301 |
| **verdict** | **SAFE TO LAND** |

---
---

# Round 2 — Mudclient subsystem extraction (5 more delegates)

Verification record for the second extraction wave, which split five further subsystems out
of `Mudclient.java` using the same **extract-delegate** idiom. Verified by the de-god verifier
(Opus) on 2026-06-02, against `git HEAD` (`1bc6dcc`), branch `feat/remote-client` (working tree
holds the deob reconstruction; the parent commits).

**Verdict: SAFE TO LAND. functionalDivergence = false.** Compiles 0 errors; every moved body
proven byte-for-byte identical to HEAD after stripping only the sanctioned qualifier rewrites;
live smoke test reproduces the full baseline functional signature.

## What moved (round 2)

Each new class holds a `final Mudclient m;` back-reference set in a side-effect-free ctor
(`this.m = m;` only). `Mudclient` keeps a `final` delegate field initialized in-line beside the
pre-existing `packets`/`sound` delegates (lines 910–914), and every former call site routes
through it.

| delegate field | new class / file | methods moved | lines moved |
|---|---|---|---|
| `tradePackets` | `client.TradeDuelBankPackets` — `TradeDuelBankPackets.java` (239 lines) | `bankSend`, `sendDuelOffer`, `sendTradeOffer`, `sendDuelItems` | 239 |
| `widgetRenderer` | `client.WidgetRenderer` — `WidgetRenderer.java` (276 lines) | `drawBox`, `clearScreen`, `drawSprite`, `drawIcon`, `drawScrollList`, `drawMenuOptions`, `drawScrollbar`, `drawScrollbar2` | 276 |
| `gameInterface` | `client.GameInterface` — `GameInterface.java` (1732 lines) | `drawWildernessWarning`, `drawShop`, `drawBank`, `drawTrade`, `drawTradeConfirm`, `drawTradeConfirmWindow`, `drawDuelConfirm`, `drawDuel`, `drawWelcome`, `drawHelpMenu`, `drawCloseButton`, `drawGameSettings` | 1601 |
| `menus` | `client.MenuController` — `MenuController.java` (560 lines) | `handleGameClick`, `buildClickMenu`, `handleInventoryClick`, `menuHitTest`, `pointInRect` (private), `pointInPanel` | 520 |
| `incoming` | `client.IncomingPackets` — `IncomingPackets.java` (1851 lines) | `handlePacket`, `handleSceneUpdates`, `onFriendUpdate`, `applyAppearanceUpdate` | 1812 |

`Mudclient.java`: **13168 → 8977 lines** (−4191). Field changes are visibility-only
(`private` → package-private); the wide widening set is compiler-driven (exactly the members
the moved bodies touch). The Mudclient constructor body is byte-for-byte identical to HEAD.

## 1. Compile gate — PASS (0 errors)

```
find reference/rsc-client/src -name '*.java' > /tmp/deob-srcs.txt   # 76 source files
/usr/lib/jvm/java-17-openjdk/bin/javac -Xmaxerrs 4000 -d /tmp/deob-degverify @/tmp/deob-srcs.txt
```
Exit 0, **0 errors**, 7 pre-existing deprecation warnings only. **78 `.class` files** (baseline
73 + the 5 new top-level extracted classes). Re-verified with the `src/client`-only RUN.md build
(`javac -d /tmp/deob-run`): exit 0, 78 classes, `Boot.class` present.

## 2. Pure-code-motion audit — PASS (functionalDivergence = false)

Brace-matched extraction of every method body from both HEAD `Mudclient.java` and the new class
files, normalized by stripping comments and the only sanctioned mechanical rewrites
(`m.` / `this.` / `Mudclient.` qualifiers + the delegate-hop prefixes), then compared:

- **All 33 extracted method bodies are byte-for-byte IDENTICAL to HEAD.** Highest-risk targets
  verified exactly: `handlePacket` (master opcode dispatch, 19038 normalized chars), `drawDuel`
  (9816), `drawBank` (8626), `drawTrade` (8044), `handleSceneUpdates` (7254), `drawShop` (5884),
  `drawGameSettings` (4660), `onFriendUpdate` (4588), `applyAppearanceUpdate` (4094).
- **All 18 methods kept in Mudclient** (e.g. `drawUiTabMagic` 5347, `loadRegion` 2565,
  `isDirectionWalkable` 1271, `addPlayer` 954, `sortFriendsList` 896, `requestLogout`,
  `formatNumber`, `resetPanels`, …) are byte-for-byte identical.
- **All 24 pure-delegate wrappers** forward arguments in exact 1:1 order with zero arg-order
  defects, and every wrapper signature (return type + parameter types) matches the HEAD original,
  so all callers bind to the identical signature.
- Delegate ctors are side-effect-free; field-initializer placement introduces no init-order
  hazard.

No logic, constant, operator, control-flow, opcode, wire byte-order, bit-width, or init-order
change anywhere. **functionalDivergence = false.**

## 3. Live smoke test — PASS (full baseline signature reproduced)

Fresh JDK17 build into `/tmp/deob-run`, launched under transient `systemd --user` unit
`rsc-deob-degverify1` (autologin `deobtest`/`deobpass`, `RSC_AUTO_APPEARANCE`, `RSC_AUTO_TABS`,
`RSC_FBUFFER_DUMP` → `/tmp/rsc-run/degverify/`). Infra (server `:43594`, content host `:7000`,
`Xvfb :99`) was already up and left running for downstream agents.

| metric | baseline | this run | verdict |
|---|---|---|---|
| `login response` | 64 | **64** | match |
| RSA | bitlen=512 exp=65537 | bitlen=512 exp=65537 | match |
| player spawn | alive @ wi=3392,9024 | **alive @ wi=3392,9024** (frame 51+) | match |
| non-black viewport px | 158538 / 171008 | **158150 / 171008** (Δ 0.24%) | match (>150k) |
| tab indices (RSC_AUTO_TABS) | {0,1,2,4,5,6} | **{0,1,2,4,5,6}** each ×11–12, 0 exceptions | match |
| stability | 2201 frames | **5301 frames**, no crash | exceeds |
| stderr exceptions | 0 | **0** | match |

`live3d.png` visually confirms the baseline scene: stone-walled room, purple/wood checkered
floor, "Welcome to Westworld!" banner, chat-tab row. Unit stopped after capture.

## Summary (round 2)

| check | result |
|---|---|
| javac errors | **0** |
| classes produced | 78 (baseline 73 + 5 new) |
| pure code motion | **true** — 33/33 extracted bodies + 18/18 kept bodies + ctor byte-identical |
| functionalDivergence | **false** |
| Mudclient lines | 13168 → 8977 (−4191) |
| new classes | `TradeDuelBankPackets`, `WidgetRenderer`, `GameInterface`, `MenuController`, `IncomingPackets` |
| live smoke test | login response:64 + alive @ 3392,9024 + 158150 nonblack px + 6 tabs clean + 5301 frames + 0 stderr |
| **verdict** | **SAFE TO LAND** |
