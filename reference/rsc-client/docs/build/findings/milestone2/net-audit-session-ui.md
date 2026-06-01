> _Verbatim report from the deob-compile-and-net-depth workflow (2026-06-01)._

## Per-Opcode Fidelity Audit — INCOMING dispatch, "session-ui" opcode group

**Scope.** Server→client opcodes for: login/registration response, stats & XP, inventory, bank, shop, trade, duel, dialogue/options, chat & private messages, friends/ignore, settings/options, sleep/fatigue, system update, logout. These are split across **two** deob dispatchers: the master `handlePacket` (obf `b(int,byte,int)`, deob `Mudclient.java:4105`) and the social/login pre-dispatcher `onFriendUpdate` (obf `a(int,int,int)`, deob `Mudclient.java:5430`). Ground truth = clean Vineflower decompile `decompiled/normalized-clean/client.java`: master at `14213`, pre-dispatcher at `10011`. Opcode numbers cross-referenced against OpenRSC `Payload235Generator` (`put(OpcodeOut.X, N)`).

**Method.** I first nailed the buffer-read primitives (`net/Buffer.java`, `net/BitBuffer.java`) so the obf accessor letters map to exact widths, then diffed every branch read-for-read:

| obf call | deob method | width |
|---|---|---|
| `mg.a((byte)104)` | `getUnsignedByte` | 1B unsigned |
| `mg.a(false)` | `getSignedShort` | 2B signed |
| `mg.h(20869)` | `readRawByte` | 1B signed (raw) |
| `mg.f(255)` | `getUnsignedShort` | 2B unsigned |
| `mg.b(-129)` | `getInt` | **4B** |
| `mg.b(byte)` | `getSmartUnsigned` | 1–2B |
| `mg.c(103)` | `getSmartSigned` | 2/4B |
| `mg.c((byte)-44)` | `getString` | NUL-terminated |
| `mg.g(0)` | `getLong` | 8B |
| `mg.f(n,w)` / `i` / `j` / `k` | `readBits` / init / finish / `getBitPosition` | bit-level |

Confirmed against OpenRSC `writeInt`/`writeShort`/`writeByte`/`writeZeroQuotedString` per opcode (e.g. SEND_STATS writes XP as `writeInt` ⇒ deob `mg.b(-129)`=4B getInt is correct).

---

### Opcodes audited (session-ui)

**`handlePacket` (master):** 111, 53, 245, 252, 25, 156, 153, 83, 59, 92, 128, 97, 162, 101, 137, 15, 240, 206, 5, 42, 203, 33, 176, 225, 20, 6, 30, 249, 90, 123, 159, 253, 210, 172, 204, 89, 222, 114, 117, 244, 84, 194, 52, 213.
**`onFriendUpdate` (social/login):** 131, 4, 183, 189, 165, 149, 237, 109, 120, 51, 87 (+ default delegation to `handlePacket`).

All read the same fields, in the same order, at the same widths as the clean decompile, **except** the two divergences below. Spot details verified as MATCH (high-risk ones):
- **156** (`Mudclient.java:4566-4571`): 18×`a(byte)` current + 18×`a(byte)` base + 18×`b(-129)` XP + `a(byte)` QP — exact bounds & 4B XP, matches clean `14694` and OpenRSC SEND_STATS.
- **234 type-1 chat** (`4283-4301`): null-player gate reads **nothing** before bailing — identical to clean `14705` (`if (null==var103) break`); no desync on unknown player. Type-5 null path skips `f(255)+c×2+a(byte)` then `mg.w += 6 + n` byte-for-byte (`4312-4317` vs clean `14772-14781`).
- **101 / 20 shop-shaped lists**, **172 / 6 / 97 duel-stake** (id `f(255)` + amount `b(-129)`=4B): widths match clean and OpenRSC `writeShort`+`writeInt`.
- **131** (`5439-5441`): the authentic *duplicate* sender-string read when `flags&1` is preserved (clean `10027-10033`).
- **120** (`5549-5561`): `g(0)` 8-byte message-id de-dup ring matches clean `10144`.

**Naming note (not a divergence):** the deob's trade/duel/shop *comments* for opcodes 92/176, 20, 6/97, 30/162/137 use labels swapped vs OpenRSC's `OpcodeOut` names (e.g. OpenRSC calls 92 TRADE_WINDOW / 176 DUEL_WINDOW; the deob comments say the opposite). Irrelevant to fidelity: for every such opcode the deob writes the **same fields the jar writes**, so behaviour is identical regardless of the label.

---

### DIVERGENCES

**DIV-1 — Opcode 137 (SEND_SHOP_CLOSE): wrong handler entirely — phantom byte read + wrong field.** SEMANTIC, MEDIUM.
- deob (`Mudclient.java:4742-4746`):
  ```java
  if (opcode == 137) { this.Mi = this.mg.h(20869) == 1; return; }
  ```
- clean (`client.java:15661-15663`, branch `~var1 == -138` ⇒ opcode 137):
  ```java
  if (~var1 == -138) { this.uk = false; return; }
  ```
- Why it matters: the jar's opcode-137 handler reads **zero** payload bytes and clears the shop/search-list-open flag `uk`. The deob instead **reads one byte** (`h(20869)`=`readRawByte`) and sets the *trade/duel accepted* flag `Mi`. SEND_SHOP_CLOSE has an empty payload, so the extra read consumes a byte that isn't there (off-by-one into the next packet on a multi-packet read), and the shop-close never clears `uk`. The deob body is a verbatim copy of opcode 15's body. `uk` gates the shop/list render path (read at deob `7026` / clean `1372`), so the UI close is broken.

**DIV-2 — Opcode 101 (shop/search-list open): extra `uk = false` clears the flag it just set.** SEMANTIC, LOW–MEDIUM.
- deob (`Mudclient.java:4702` sets `uk=true`, then **`4715` sets `uk=false`**) inside the same handler.
- clean (`client.java:15534` sets `uk=true` and **never clears it** in this branch).
- Why it matters: after a list-open packet the jar leaves `uk == true` (list shown); the deob leaves `uk == false`. Combined with DIV-1 (137 no longer clears `uk`), the entire `uk` open/close lifecycle is inverted: the deob author appears to have relocated 137's `uk=false` into 101 and overwritten 137 with opcode-15's logic. No buffer-width change here (the read sequence in 101 is otherwise byte-identical to clean), purely a wrong state write.

**MINOR-1 — Opcode 149 (SEND_FRIEND_UPDATE): `messageSlot` constant-folded `var1 ^ 87` → `0`.** COSMETIC/EQUIVALENT.
- clean (`client.java:10223`) passes `var1 ^ 87` as the `messageSlot` arg of `showServerMessage(...)` in exactly one place (the "<friend> has logged out", matched-by-current-name branch); the deob (`Mudclient.java:5474`) passes `0`. `messageSlot` indexes `messageHistoryMessage[messageSlot]` (`Mudclient.java:2130`); every other of the ~12 call sites passes 0. Given the pre-dispatcher's own tail invariant `if (a != 87) requestLogout(56)` (clean `10299`), the normal-dispatch caller passes `a == 87`, so `var1 ^ 87 == 0` and the fold is behaviourally exact. Listed only because correctness depends on that caller invariant (which lives in the net layer, proven faithful in Milestone-1).

---

## NUMBERS

- **Opcodes audited (this group): 55** — `handlePacket`: 111, 53, 245, 252, 25, 156, 153, 83, 59, 92, 128, 97, 162, 101, 137, 15, 240, 206, 5, 42, 203, 33, 176, 225, 20, 6, 30, 249, 90, 123, 159, 253, 210, 172, 204, 89, 222, 114, 117, 244, 84, 194, 52, 213 (44); `onFriendUpdate`: 131, 4, 183, 189, 165, 149, 237, 109, 120, 51, 87 (11).
- **MATCH: 52**
- **MINOR: 1** (opcode 149 — `var1^87`→`0` constant fold, equivalent under caller invariant)
- **DIVERGENCE: 2**
  - **Opcode 137 (SEND_SHOP_CLOSE)** — SEMANTIC, **MEDIUM**: reads a phantom byte (jar reads 0) and sets `Mi` instead of clearing `uk`; body is a wrong copy of opcode 15. Risk: 1-byte buffer desync on empty-payload close + shop/list never closes.
  - **Opcode 101 (shop/search-list open)** — SEMANTIC, **LOW–MEDIUM**: extra `uk = false` after `uk = true`; jar leaves `uk == true`. Inverts the `uk` open/close UI state lifecycle (no buffer-width change).

Both divergences are UI/session state, not world-perception (no scenery/boundary/entity coordinate impact). DIV-1 is the only one with a buffer-read-width consequence (one extra `readRawByte` on a zero-length packet).
