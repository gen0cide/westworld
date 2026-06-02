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
