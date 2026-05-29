# The Wire Protocol — Payload235 shadow (companion to api.md)

> Status: **companion**. This is the packet/protocol shadow of the Routine
> API (`api.md`) — the wire-level layer underneath every faculty. It exists
> for *our own sanity* while building and debugging the cradle, not for
> host-facing routine authors. A persona never sees an opcode; it sees
> `trade.offer([...])`. This doc is how we know `trade.offer` puts
> `[byte count] [short id, int amount]…` on opcode 46.
>
> **Opcode bytes are the cradle's own constants.** Outbound values are the
> bytes we *encode* (`proto/v235/opcodes.go`, the per-subsystem `action/*.go`
> `out…` consts); inbound values are the bytes our decode `switch` matches
> on (`proto/v235/inbound_opcodes.go`, `inbound.go`). All values are for
> **protocol v235** (mc234/235-era final RSC, what RSC+ speaks); OpenRSC's
> `Payload235Parser` / `Payload235Generator` are the upstream reference.
> Other revisions are out of scope.
>
> **Each section maps to an `api.md` namespace** (`self`/`world`,
> `inventory`, ambient verbs, `combat`, `magic`, `prayer`, `trade`, `bank`,
> `shop`, `duel`, plus the social/chat ambient verbs). Where `api.md` says
> "wire opcode 142," this is the section that says what bytes follow it.

---

## 0. Framing, crypto, login (the shared prelude)

Everything below this section rides on the same frame + ISAAC envelope. This
prelude is namespace-agnostic — it's the transport every faculty packet uses.

### Framing

After the RSA-encrypted login completes, all packets in both directions share
one frame structure. **The opcode byte (post-ISAAC) is encrypted via the ISAAC
stream cipher; the payload is not.**

**Outbound** (client → server), opcode `op`, payload `data` of length `N`,
`length = N + 1`:

- **`length == 1`** (opcode only): `[length=1] [ENC(op)]`
- **`length < 160` and `!= 1`**: tail-byte reordering — the last payload byte
  is moved between the length byte and the opcode:
  `[length] [data[N-1]] [ENC(op)] [data[0..N-2]]`
  (OpenRSC `RSCProtocolEncoderMain.java:76`: *"Strangely, the last byte of the
  Payload goes between length and encoded opcode"*.)
- **`length >= 160`**: two-byte length:
  `[(length/256) + 160] [length & 0xFF] [ENC(op)] [data[0..N-1]]`

**Inbound** (server → client) is symmetric. Read length byte; if `< 160` that's
the length, else read a second byte and compute
`length = (firstByte - 160) * 256 + secondByte`
(`RSCProtocolDecoder.java:154`). On a one-byte length with `length > 1`, the
next byte is the tail-byte: read it, read `length - 1` more, first of those is
`ENC(op)`, the rest plus the tail-byte (appended) form the payload.

**ISAAC desync mitigation.** The decoder loops up to 256 ISAAC states looking
for an `isPossiblyValid` opcode (`RSCProtocolDecoder.java:180-212`) — a
recoverable-desync safety net. We aim to never desync; the loop is a fallback.

### ISAAC cipher

Standard Bob Jenkins ISAAC (1996). Two instances per connection: `inCipher`
(decrypt inbound opcodes), `outCipher` (encrypt outbound opcodes), both seeded
from the same 4×int (16-byte) key block from the RSA login payload. Only the
opcode byte is touched:

- Encode: `enc_op = (op + outCipher.nextInt()) & 0xFF`
- Decode: `op = (enc_op - inCipher.nextInt()) & 0xFF`

Sources: `proto/v235/isaac.go`; upstream `login/ISAACCipher.java`,
`net/rsc/ISAACContainer.java:15-21`.

### Login sequence

- **Phase 0 — session ID.** Server sends (after ~100ms,
  `RSCSessionIdSender.java`) 4 or 8 raw bytes of session ID, no wrapper.
  v1 assumption: 8-byte long; verify on test.
- **Phase 1 — login (opcode 0).** Sent *before* ISAAC is active, so framing
  is unusual: 2-byte big-endian length, plain opcode `0`, no tail-byte.
  ```
  [length_hi] [length_lo] [opcode=0] [version_block]
  [rsa_block_len] [RSA block] [xtea_block_len] [XTEA block]
  ```
  RSA block (server-decrypted): `[1] checksum=10` · `[16] 4 ints = XTEA/ISAAC
  keys` · `[20] password, null-padded, spaces→underscores` · `[24] 6 nonces`.
  XTEA block (encrypted with those 4 keys): `[1] limit30` · `[24] 6 more
  nonces` · `[?] username, null-terminated`. We target client version 235
  (the RSA+XTEA path), not the 93–177 hash path.
- **Phase 2 — login response.** Server sends **1 raw byte** (no framing):

  | Code | Meaning |
  |---|---|
  | 0 | LOGIN_UNSUCCESSFUL |
  | 1 | RECONNECT_SUCCESSFUL |
  | 3 | INVALID_CREDENTIALS |
  | 4 | ACCOUNT_LOGGEDIN |
  | 7 | LOGIN_ATTEMPTS_EXCEEDED |
  | 64+ | LOGIN_SUCCESSFUL (varies by group; 86 = admin/group 1, 64 = regular/group 10) |

  Success ⇔ `(code & 0x40) != 0` (`v235.LoginSuccessful`).
- **Phase 3 — ISAAC seeding.** Both ends init ISAAC from the 4 RSA keys. From
  here on, all framed packets use ISAAC-encoded opcodes.

### Data type encodings

All multi-byte numbers big-endian.

| Type | Bytes |
|---|---|
| byte | 1 (unsigned via `& 0xFF`) |
| short | 2 |
| int | 4 |
| long | 8 |
| string | variable; null-terminated, zero-quoted, or zero-padded |
| smart08_16 | 1 byte if `< 128`, else 2 (`Packet.getSmart08_16`) |
| unsignedShortInt-smart | OpenRSC "smart" amount encoding (`readUnsignedShortIntSmart`) |

Bitpacked payloads (191, 79, 234) read MSB-first within each byte.

### Heartbeat

Opcode **67**, empty payload, periodic (~5s, RSC-traditional; verify by
observation). Without it the server may drop the connection as dead.

---

## 1. `self` / `world` — position, NPCs, ground items, boundaries

These back the `self` View (`position`, `hp`, `fatigue`, …) and the `world`
View lists (`npcs`, `players`, `ground_items`, `boundaries`, `messages`).

### Own + nearby players — opcode 191 (`InSendPlayerCoords`)

Bitpacked per-tick update. Always carries own position; nearby players follow.
Bit layout (client ≥ 177, includes v235):

```
[11 bits]  own X
[13 bits]  own Y
[4 bits]   own sprite (facing/anim, 0-15)
[8 bits]   localPlayersCount

for each local player:
  [1 bit]  needsUpdate
  if set:
    [1 bit]  updateType (0=movement, 1=removal-or-sprite)
    if 0: [3 bits] direction (0-7)
    if 1: [2 bits] subType — 3=remove player; else sprite update w/ 4 bits
          if subType != 3: [4 bits] new sprite

for each NEW player entering view (until packet ends):
  [11 bits] player index
  [5 bits]  offset X (rel to own X, mod-32 mask)
  [5 bits]  offset Y
  [4 bits]  sprite
  [1 bit]   isKnownPlayer flag (omitted on some client versions)
```

The new-player loop end vs. trailing garbage is ambiguous from the packet
alone; we read to byte-granular buffer exhaustion (fine for Phase 1.5).

### NPC movement / spawn — opcode 79 (`InNpcCoords`)

Bitpacked, ported from `GameStateUpdater.updateNpcs`:

```
[8 bits]  localNpcCount

for each local NPC (server's order):
  [1 bit]  needsUpdate
  if set:
    [1 bit]  discriminator (0=movement, 1=not-moving)
    if 0: [3 bits] sprite (direction 0-7)
    if 1: peek next 2 bits —
            == 3 (0b11): REMOVE_NPC (consume 2 bits)
            else: high 2 bits of a 4-bit sprite; consume 2 more (sprite changed)

for each NEW NPC entering view (until packet ends):
  [12 bits] npc index
  [5 bits]  offset X (signed, mod-64 wrapped)
  [5 bits]  offset Y (signed, mod-64 wrapped)
  [4 bits]  sprite
  [10 bits] npc_type_id (joins NpcDef.ID)
```

Sprite-only updates are consumed but suppressed (noise reduction); movement
and removal are emitted.

### Ground items — opcode 99 (`InGroundItemHandler`)

Per `GameStateUpdater.java:1128-1175`. Multiple entries may pack into one
frame. Per entry: a `0xFF` first byte is a removal sentinel (id follows;
format is inconsistent and currently treated as unknown). Otherwise the first
two bytes form a uint16 item ID where the high bit (`0x8000`) marks
disappearance; then `[byte offsetX] [byte offsetY]` (signed, player-relative).

```
itemID = (first << 8) | low
disappear = itemID & 0x8000
if disappear: itemID &= 0x7FFF
[byte] offsetX (int8)
[byte] offsetY (int8)
```

### Boundary state — opcode 91 (`InBoundaryHandler`, SEND_BOUNDARY_HANDLER)

Dynamic boundary updates (opened doors, cut webs). Each record:
`[short id] [byte offsetX] [byte offsetY] [byte dir]`, coordinates
player-relative. `id == 0xFFFF` (→ -1) means the boundary was removed.

### Misc world / self inbound

| Opcode | Const | Payload |
|---|---|---|
| 33 | `InExperience` | single skill XP changed |
| 114 | `InFatigue` | fatigue value changed |
| 156 | `InStats` | full stats dump |
| 159 | `InStat` | one stat: level + max + xp |
| 83 | `InDeath` | we died |
| 84 | `InStopSleep` | sleep ended |
| 117 | `InSleepScreen` | sleep-word captcha appeared |
| 194 | `InSleepwordIncorrect` | last sleep-word was wrong |
| 182 | `InWelcomeInfo` | post-login welcome screen |
| 52 | `InSystemUpdate` | system-update countdown |
| 165 | `InSendLogout` | server confirms clean logout |

---

## 2. `inventory` — full dump, slot updates, ground pickup/drop

Backs `inventory.items`/`find`/`count` and the ambient `pick_up`, `drop`,
`bury`/`eat` verbs.

### Inbound

| Opcode | Const | Payload |
|---|---|---|
| 53 | `InInventory` | full inventory dump (per-slot id + amount; amount present per stackability, `0xFFFF` sentinel) |
| 90 | `InInventorySlotUpdate` | single slot changed |
| 123 | `InInventoryRemoveItem` | item removed |

Amount encoding follows item stackability — the decoder is handed an
`isStackable(itemID)` predicate; stackable items carry an amount field,
non-stackable do not. (Whether `facts.ItemStackability` is the single source
of truth is an open question below.)

### Outbound — ambient item verbs (`action/interaction.go`, `equip.go`)

| Opcode | Const | Payload | Verb |
|---|---|---|---|
| 247 | `outGroundItemTake` (GROUND_ITEM_TAKE) | `[short x] [short y] [short itemId]` | `pick_up` |
| 246 | `outItemDrop` (ITEM_DROP) | `[short inventorySlot]` | `drop` |
| 90 | `outItemCommand` (ITEM_COMMAND) | `[short slot]` — default right-click action (server picks command1: Bury/Eat/Drink/Bake/…) | `bury`/`eat`/… |
| 165 | `outPlayerFollow` (PLAYER_FOLLOW) | `[short serverIndex]` | (follow) |
| 169 | `outItemEquip` | `[short slot]` | `equip` |
| 170 | `outItemUnequip` | `[short slot]` | (unequip) |

> Note opcode **90 and 165 reuse**: 90 is both inbound slot-update and outbound
> item-command; 165 is both inbound logout-ack and outbound player-follow.
> Disambiguated by direction (decode `switch` vs. encode const).

---

## 3. Ambient verbs — movement, use-on-X, dialog, examine

### Movement (`action/walk.go`)

| Opcode | Const | Payload |
|---|---|---|
| 187 | `OutWalkToPoint` (WALK_TO_POINT) | `[short x] [short y]` + optional waypoint pairs |
| 16 | `OutWalkToEntity` (WALK_TO_ENTITY) | same shape; hints "approach to interact, don't land on the tile" |

Waypoint cap: roughly `count - 25` corners of deltas. Walk-search distance
override: authentic 15, Phase-1.7 testing uses 30.

### Use-with-X (`action/use.go`)

Each target class has its own opcode (the client picks at click time):

| Opcode | Const | Payload |
|---|---|---|
| 91 | `outUseItemOnItem` (ITEM_USE_ITEM) | `[short slot1] [short slot2]` |
| 53 | `outUseItemOnGroundItem` (GROUND_ITEM_USE_ITEM) | `[short x] [short y] [short groundItemID] [short slotID]` |
| 115 | `outUseItemOnScenery` (USE_ITEM_ON_SCENERY) | `[short x] [short y] [short slotID]` (no direction byte) |
| 161 | `outUseItemOnBoundary` (USE_WITH_BOUNDARY) | `[short x] [short y] [byte direction] [short slotID]` |
| 50 | `outUseItemOnNpc` (NPC_USE_ITEM) | `[short serverIndex] [short slotID]` |
| 113 | `outUseItemOnPlayer` (PLAYER_USE_ITEM) | `[short serverIndex] [short slotID]` |
| 136 | `outObjectCommand` (OBJECT_COMMAND) | primary click on scenery (opt 1) |
| 79 | `outObjectCommand2` (OBJECT_COMMAND2) | secondary click on scenery (opt 2) |

> Opcode **50 reuse**: `outUseItemOnNpc` (NPC_USE_ITEM) shares the byte with
> `outCastOnNpc` (CAST_ON_NPC, §5) — disambiguated by length/context. Opcode
> **53** is both this use-on-ground-item and inbound full-inventory; **79** is
> both OBJECT_COMMAND2 and inbound NpcCoords.

### NPC clicks (`action/combat.go`)

| Opcode | Const | Payload |
|---|---|---|
| 153 | `outNpcTalkTo` | `[short serverIndex]` — `talk_to` |
| 202 | `outNpcCommand` | `[short serverIndex]` — right-click (talk/attack/use by default cmd) |

### Dialog (`action/dialog.go`; inbound 222 / 245)

| Opcode | Const / dir | Payload |
|---|---|---|
| 222 | `InNpcDialogText` (BOX, in) | `[zero-quoted string] text` |
| 245 | `InNpcDialogOptions` (OPTIONS_MENU_OPEN, in) | `[byte num_options]` then per option `[zero-quoted string]` |
| 116 | `outDialogChoice` (QUESTION_DIALOG_ANSWER, out) | `[byte optionIndex]` |

### Boundary interact (`action/boundary.go`)

| Opcode | Const | Payload |
|---|---|---|
| 14 | `outInteractWithBoundary` | `[short x] [short y] [byte direction]` — primary |
| 127 | `outInteractWithBoundary2` | same — right-click alternate |

### Examine

No dedicated outbound opcode is confirmed in `action/*.go` yet (see open
questions — may reuse ItemCommand/object-command with a command variant).

---

## 4. `combat` — attack, style; inbound damage / projectiles

### Outbound (`action/combat.go`)

| Opcode | Const | Payload |
|---|---|---|
| 190 | `outNpcAttack` | `[short serverIndex]` |
| 171 | `outPlayerAttack` | `[short serverIndex]` (PvP) |
| 29 | `outCombatStyle` | `[byte style]` — 0=controlled, 1=aggressive, 2=accurate, 3=defensive |

### Inbound combat — carried inside opcode 234 (`InSendUpdatePlayers`)

The update-players packet (§7 trade has no claim here — this is per-player
update records) carries combat events as `updateType` records:

```
[byte updateType]
  2 = damage taken:   [byte dmg] [byte curHits] [byte maxHits]
  3 = projectile→NPC:    [short projID] [short victimNpcIndex]
  4 = projectile→player: [short projID] [short victimPlayerIndex]
```

`dmg`/`curHits`/`maxHits` are each a single byte (0–255). Damage is player-only
in this packet (no inbound NPC-damage record). NPC health on `Npc.health` (the
0–1 target health bar in `api.md`) derives from the bitpacked NPC update, not
from a damage event here.

---

## 5. `magic` — cast variants

Each cast target class has its own opcode keyed on what's targeted
(`action/magic.go`). Spell ID is u16:

| Opcode | Const | Payload |
|---|---|---|
| 137 | `outCastOnSelf` | `[u16 spellID]` |
| 50 | `outCastOnNpc` | `[u16 npcIndex] [u16 spellID]` |
| 229 | `outCastOnPlayer` | `[u16 playerIndex] [u16 spellID]` |
| 158 | `outCastOnLand` | `[u16 x] [u16 y] [u16 spellID]` |
| 249 | `outCastOnGround` | `[u16 x] [u16 y] [u16 itemID] [u16 spellID]` |
| 99 | `outCastOnScenery` | `[u16 x] [u16 y] [u16 spellID]` |
| 180 | `outCastOnBoundary` | `[u16 x] [u16 y] [u8 dir] [u16 spellID]` |
| 4 | `outCastOnInventory` | `[u16 slot] [u16 spellID]` — disambiguated from FORGOT_PASSWORD by login state |

> Opcode reuse: **50** CAST_ON_NPC vs. NPC_USE_ITEM (§3); **249** CAST_ON_GROUND
> vs. inbound bank-update (§8); **99** CAST_ON_SCENERY vs. inbound
> ground-item-handler (§1); **4** CAST_ON_INVENTORY vs. FORGOT_PASSWORD
> (login-state disambiguated). These are resolved by direction + login state.

---

## 6. `prayer` — activate / deactivate

`action/prayer.go`. Prayer index 0..13 (Thick Skin, Burst of Strength, …):

| Opcode | Const | Payload |
|---|---|---|
| 60 | `outPrayerActivated` | `[byte prayerID]` |
| 254 | `outPrayerDeactivated` | `[byte prayerID]` |
| 206 | `InPrayersActive` (in) | `[N bytes]`, one per prayer slot, 0/1 |

---

## 7. `trade` — two-screen handshake

The fully-worked exemplar in `api.md`. Two distinct accepts (offer screen,
then confirm screen). Item lists are `[byte count]` then per-item
`[short catalogID] [int amount]`; up to 255 items wire-wise (12-item authentic
limit is enforced elsewhere). `action/trade.go`.

### Outbound

| Opcode | Const | Payload | `api.md` |
|---|---|---|---|
| 142 | `outInitTradeRequest` (PLAYER_INIT_TRADE_REQUEST) | `[short serverIndex]` | `trade.request` (also re-sends to accept an incoming request) |
| 46 | `outAddTradeItems` (PLAYER_ADDED_ITEMS_TO_TRADE_OFFER) | `[byte count]` + per-item `[short id] [int amount]` | `trade.offer` |
| 55 | `outAcceptTradeOffer` (PLAYER_ACCEPTED_INIT_TRADE_REQUEST) | (empty) — first accept, offer screen | `trade.accept` |
| 104 | `outAcceptTradeConfirm` | (empty) — second accept, confirm screen | `trade.confirm` |
| 230 | `outDeclineTrade` | (empty) | `trade.decline` |

### Inbound

| Opcode | Const | Notes |
|---|---|---|
| 92 | `InTradeWindow` | SEND_TRADE_WINDOW — opens the OFFER screen |
| 97 | `InTradeOtherItems` | `[byte count]` + per-item `[short id] [int amount]` — opponent's offer |
| 162 | `InTradeOtherAccepted` | other player clicked accept (offer screen) |
| 15 | `InTradeAccepted` | accept echo / reset |
| 20 | `InTradeOpenConfirm` | SEND_TRADE_OPEN_CONFIRM — FINAL review screen; `[zero-quoted name]` then opponent items, then own items (each `[byte count]` + `[short id] [int amount]`) |
| 128 | `InTradeClose` | trade cancelled |

---

## 8. `bank` — open dump, slot update, deposit/withdraw

`action/bank.go`. The bank UI must already be open server-side
(`player.accessingBank()`), set by talking to a banker NPC; otherwise
deposit/withdraw are rejected as "suspicious."

### Outbound

| Opcode | Const | Payload |
|---|---|---|
| 23 | `outBankDeposit` | `[short catalogID] [int amount] [int magicNumber=0]` (magicNumber ignored by server) |
| 22 | `outBankWithdraw` | same shape as deposit |
| 212 | `outBankClose` | (empty) — releases the accessingBank flag |

### Inbound

| Opcode | Const | Payload |
|---|---|---|
| 42 | `InBankOpen` | SEND_BANK_OPEN — full slot dump |
| 249 | `InBankUpdate` | SEND_BANK_UPDATE — `[byte slot] [short catalogID] [smartUnsignedShortInt amount]` |
| 203 | `InBankClose` | bank window closed |

---

## 9. `shop` — open / close

`shop.buy`/`sell`/`stock`/`price` in `api.md`. Inbound opcodes are defined;
the decoders and outbound buy/sell payloads are not yet implemented in the
cradle (carried in open questions, including `stockSensitivity` /
`isGeneralStore` semantics).

| Opcode | Const | Notes |
|---|---|---|
| 101 | `InShopOpen` | shop window opened with inventory |
| 137 | `InShopClose` | shop window closed |

---

## 10. `duel` — two-screen handshake + rule toggles

Mirrors `trade` (offer screen → confirm screen) plus four pre-fight rule
toggles. Item lists same shape as trade. `action/duel.go`.

### Outbound

| Opcode | Const | Payload | `api.md` |
|---|---|---|---|
| 103 | `outInitDuelRequest` | `[short serverIndex]` | `duel.request` |
| 33 | `outDuelOfferItem` (DUEL_OFFER_ITEM) | `[byte count]` + per-item `[short id] [int amount]` | `duel.stake` |
| 8 | `outDuelSettingsChanged` (DUEL_FIRST_SETTINGS_CHANGED) | `[4 × byte]` disallow retreat/magic/prayer/weapons (resets both first-accepts) | `duel.set_rules` |
| 176 | `outDuelFirstAccepted` | (empty) — first accept, offer screen | `duel.accept` |
| 77 | `outDuelSecondAccepted` | (empty) — final accept, confirm screen | `duel.confirm` |
| 197 | `outDuelDeclined` | (empty) — abort at any phase | `duel.decline` |

### Inbound

| Opcode | Const | Notes |
|---|---|---|
| 176 | `InDuelWindow` | SEND_DUEL_WINDOW — offer screen opened (note: same byte as outbound first-accept) |
| 6 | `InDuelItems` | SEND_DUEL_OPPONENTS_ITEMS — opp's stake |
| 30 | `InDuelSettings` | SEND_DUEL_SETTINGS — unified rule toggles broadcast back |
| 210 | `InDuelAccepted` | SEND_DUEL_ACCEPTED — our accept echoed back (client treats as no-op) |
| 253 | `InDuelOtherAccepted` | SEND_DUEL_OTHER_ACCEPTED — opp clicked accept |
| 172 | `InDuelConfirmWindow` | SEND_DUEL_CONFIRMWINDOW — final review screen |
| 225 | `InDuelClose` | SEND_DUEL_CLOSE — duel cancelled |

---

## 11. Social / chat ambient verbs (`say`, `whisper`, `add_friend`)

`action/chat.go`, `action/pm.go`. Chat bodies are RSC-compressed (a fixed
Huffman-like prefix code; `proto/v235/stringencryption.go`,
`EncipherRSCString`).

### Outbound

| Opcode | Const | Payload |
|---|---|---|
| 216 | `outChatMessage` (CHAT_MESSAGE) | RSC-compressed message (`getEncryptedString`) — `say` |
| 218 | `outPrivateMessage` (SOCIAL_SEND_PRIVATE_MESSAGE) | `[zero-padded recipient] [smart08_16 char-count] [RSC-compressed bytes]` — `whisper` |
| 195 | `outAddFriend` (SOCIAL_ADD_FRIEND) | `[string name]` — `add_friend` |
| 167 | `outRemoveFriend` (SOCIAL_REMOVE_FRIEND) | `[string name]` |

> PM gotcha: the smart char-count prefix is load-bearing — without it the
> server reads the first ciphertext byte as the length and decompresses bogus
> characters. Recipient must have us mutually friended or the server silently
> drops the packet.

### Inbound

| Opcode | Const | Notes |
|---|---|---|
| 131 | `InServerMessage` | chat / server message / quest message (trade & duel *requests* arrive as message-type substrings here today — fragile, see open questions) |
| 120 | `InPrivateMessage` | PM from another player (carries a sender icon/rank sprite) |

### Public chat / quest chat — inside opcode 234

`updateType` records in the update-players packet (§4 also rides 234) carry
chat: `0` = action bubble (`[short bubbleID]`); `1`/`6`/`7` = public / quest /
muted chat (`[smart char-count]` + RSC-compressed body); `5` = appearance /
identity (variable, see open questions).

---

## 12. Admin / test (control plane, fenced — `command()`)

`action/command.go`. Outside `api.md`'s GUI-equivalence rule (the fenced
admin/test layer).

| Opcode | Const | Payload |
|---|---|---|
| 38 | `outCommand` (COMMAND) | `[string command]` — e.g. `heal`, `teleport 120 504`, `spawnnpc 184` |

---

## Outbound opcode quick-index

| Op | Const | Op | Const | Op | Const |
|---|---|---|---|---|---|
| 0 | OutLogin | 4 | CastOnInventory | 8 | DuelSettingsChanged |
| 14 | InteractWithBoundary | 16 | WalkToEntity | 22 | BankWithdraw |
| 23 | BankDeposit | 29 | CombatStyle | 31 | ConfirmLogout |
| 33 | DuelOfferItem | 38 | Command | 46 | AddTradeItems |
| 50 | UseItemOnNpc / CastOnNpc | 55 | AcceptTradeOffer | 60 | PrayerActivated |
| 67 | Heartbeat | 77 | DuelSecondAccepted | 79 | ObjectCommand2 |
| 90 | ItemCommand | 91 | UseItemOnItem | 99 | CastOnScenery |
| 102 | Logout | 103 | InitDuelRequest | 104 | AcceptTradeConfirm |
| 113 | UseItemOnPlayer | 115 | UseItemOnScenery | 116 | DialogChoice |
| 127 | InteractWithBoundary2 | 136 | ObjectCommand | 137 | CastOnSelf |
| 142 | InitTradeRequest | 153 | NpcTalkTo | 158 | CastOnLand |
| 161 | UseItemOnBoundary | 165 | PlayerFollow | 167 | RemoveFriend |
| 169 | ItemEquip | 170 | ItemUnequip | 171 | PlayerAttack |
| 176 | DuelFirstAccepted | 180 | CastOnBoundary | 187 | WalkToPoint |
| 190 | NpcAttack | 195 | AddFriend | 197 | DuelDeclined |
| 202 | NpcCommand | 212 | BankClose | 216 | ChatMessage |
| 218 | PrivateMessage | 229 | CastOnPlayer | 230 | DeclineTrade |
| 246 | ItemDrop | 247 | GroundItemTake | 249 | CastOnGround |
| 254 | PrayerDeactivated | | | | |

## Inbound opcode quick-index

| Op | Const | Op | Const | Op | Const |
|---|---|---|---|---|---|
| 6 | DuelItems | 15 | TradeAccepted | 20 | TradeOpenConfirm |
| 30 | DuelSettings | 33 | Experience | 42 | BankOpen |
| 52 | SystemUpdate | 53 | Inventory | 79 | NpcCoords |
| 83 | Death | 84 | StopSleep | 90 | InventorySlotUpdate |
| 91 | BoundaryHandler | 92 | TradeWindow | 97 | TradeOtherItems |
| 99 | GroundItemHandler | 101 | ShopOpen | 114 | Fatigue |
| 117 | SleepScreen | 120 | PrivateMessage | 123 | InventoryRemoveItem |
| 128 | TradeClose | 131 | ServerMessage | 137 | ShopClose |
| 156 | Stats | 159 | Stat | 162 | TradeOtherAccepted |
| 165 | SendLogout | 172 | DuelConfirmWindow | 176 | DuelWindow |
| 182 | WelcomeInfo | 191 | SendPlayerCoords | 194 | SleepwordIncorrect |
| 203 | BankClose | 206 | PrayersActive | 210 | DuelAccepted |
| 222 | NpcDialogText | 225 | DuelClose | 234 | SendUpdatePlayers |
| 245 | NpcDialogOptions | 249 | BankUpdate | 253 | DuelOtherAccepted |

---

## Open questions / unverified

- [wire-protocol-interactions] Bitpack encoding details for PlayerCoords and UpdatePlayers (field order, bit widths)
- [wire-protocol-interactions] NPC Dialog color tag format in opcode 222 (hex vs named vs other)
- [wire-protocol-interactions] Appearance Screen customization payload format (opcode 59 → opcode 235 response)
- [wire-protocol-interactions] Private Message icon sprite ID mapping to player rank/status
- [wire-protocol-interactions] Boundary interaction direction byte encoding (north/east/south/west values)
- [wire-protocol-interactions] Exact XTEA key derivation from ISAAC keys; confirm first 4 uint32s used
- [wire-protocol-interactions] Opcode disambiguation when same byte appears in multiple contexts (e.g., 50: CAST_ON_NPC vs NPC_USE_ITEM by length/context)
- [wire-protocol-mapping] NPC index correlation across despawn/respawn cycles
- [wire-protocol-mapping] Full appearance block layout (outfit, colors, animations, equipment) for opcode 234 type 5
- [wire-protocol-mapping] Projectile visual fields for opcode 234 types 3 and 4
- [wire-protocol-mapping] Smart encoding variance cases (non-standard length headers)
- [wire-protocol-mapping] ISAAC cipher sync edge cases post-login
- [wire-protocol-mapping] Boundary direction field: 4-way vs 8-way compass encoding
- [wire-protocol-mapping] Server-side stackable item definition mechanism
- [wire-protocol-research] ItemUseOnGroundItem exact field order and wire alignment
- [wire-protocol-research] Confirm opcode 249 is CAST_ON_GROUND_ITEM in server payload parser
- [wire-protocol-research] Verify facts.ItemStackability is the single source of truth for amount encoding in inventory packets
- [wire-protocol-research] Player position sprite encoding: confirm if sprite is 4 bits or if subType bits alias with sprite bits across versions
- [wire-protocol-research] Ground item opcode 99 payload structure: confirm 0xFF + id vs other removal markers
- [wire-protocol-research] Duel request detection via substring parsing in type-7 SEND_SERVER_MESSAGE is fragile; confirm server message format stability
- [wire-protocol-research] Trade/duel requests via SEND_SERVER_MESSAGE (type 6 and type 7) should ideally have dedicated opcodes to avoid parsing fragility
- [wire-protocol-documentation] Ground item removal packet format (opcode 99 0xFF sentinel behavior)
- [wire-protocol-documentation] Exact smart encoding threshold for inventory amounts in Payload235Generator
- [wire-protocol-documentation] Bank preset opcodes (91, 92) wire format and availability by world/client
- [wire-protocol-documentation] Duel combat transition signal (no explicit packet, inferred from state)
- [wire-protocol-documentation] Client version item ID upper limits and ItemDef.maxId enforcement
- [wire-protocol] Examine interaction opcode and payload structure — whether it uses a dedicated opcode or reuses ItemCommand (90) with a special command variant.
- [wire-protocol] Boundary interact (opcodes 14, 127) outbound payload structure in cradle — assumed to match TargetObjectStruct but not yet verified in action/*.go.
- [wire-protocol] GroundItemHandler (opcode 99) multi-entry packet semantics — does server ever pack multiple add/remove records in a single frame?
- [wire-protocol] ItemCommand custom-client amount field semantics — does amount mean 'perform N times' or 'set resulting stack to N'?
- [wire-protocol] Equipment-tab fallback logic in ItemOnObjectStruct.itemID — exact conditions for client/server using itemID-based lookup vs. slotID.
- [wire-protocol] Exact opcode number for GROUND_ITEM_TAKE on outbound (action/interaction.go) — confirmed as 247 in server but not yet read from cradle source.
- [wire-protocol] Option 30 behavior: Is it a legacy RSC quirk or explicit cancel code?
- [wire-protocol] Duel/trade completion: Should success detection rely on inventory deltas or explicit packet?
- [wire-protocol] Menu truncation: Should server reject >5 options or silently truncate?
- [wire-protocol] NPC busy timeout: Is 20-second multi-timeout hardcoded or config-driven?
- [combat-wire-protocol] Damage capping: byte field (0–255) in opcode 234 type 2 cannot represent damages >255; does server split large damages?
- [combat-wire-protocol] NPC health bars: no inbound damage event for NPCs in opcode 234 type 2 (player-only); how are NPC bars displayed?
- [combat-wire-protocol] Retro vs. custom client differences: do opcode or field layouts change between client versions (retro 38–39, authentic 61–204, custom)?
- [combat-wire-protocol] Combat event sequencing: is damage update (opcode 234 type 2) sent before/after/during health sprite change on opcode 79 or 234?
- [combat-wire-protocol] Update-type 7 authenticity: GameStateUpdater marks type 7 as 'not authentic' but generates it; is this OpenRSC-only or incompatible with authentic client?
- [trade-wire-subsystem] Are trade payload fields (item IDs, amounts) ISAAC-encrypted before wire transmission or handled raw by the action layer?
- [trade-wire-subsystem] Should the cradle decoder honor the noted flag from inbound trade packets, or is noted=false always correct for the mirror?
- [trade-wire-subsystem] What does the tradeAccepted byte value signify beyond 0/1, and why is the comparison on line 191 of PlayerTradeHandler checking == 1 specifically?
- [trade-wire-subsystem] Is the order of items in the confirm screen (opponent first, then own) guaranteed, or does the server vary this?
- [wire-protocol] Trade noted items: Are noted fields in the struct actually serialized in new client versions, or are they legacy/zero-always?
- [wire-protocol] TradeAcceptStruct encoding: Why are opcodes 15 and 162 both used for the same [byte accepted] field? Does 15 only fire on reset?
- [wire-protocol] DuelAcceptStruct encoding: Why is opcode 210 (our own accept echo) sent at all if it's treated as a no-op by the client?
- [wire-protocol] ISAAC cipher alignment: Any field-boundary padding or alignment quirks in the encrypted frames for trade/duel item lists?
- [wire-protocol] Item count caps: Are the 12-item trade limit and 8-item duel limit RSC-authentic, or OpenRSC safeguards?
- [wire-protocol-bank] Does the server ever send SEND_BANK_UPDATE with the same slot index twice in sequence (e.g., deposit creates new slot, then updates it again)? Behavior suggests single update per action, but untested edge case.
- [wire-protocol-bank] When bank size exceeds 255 slots (member world with custom content), how does the wire represent itemsStoredSize and maxBankSize? Code shows truncation to 255; is this a known client limitation or does a larger custom client exist?
- [wire-protocol-bank] The authentic client's magicNumber in deposit/withdraw — has its original purpose (if any) been documented in RSC-era changelogs, or is it pure dead code?
- [wire-protocol-mapping] What values exist for the stockSensitivity byte in SEND_SHOP_OPEN beyond 0 (flat) and 1 (stock-adjusted)?
- [wire-protocol-mapping] Is ground item ID globally unique per world or scoped to tile? How does server disambiguate multiple stacks of same item on one tile?
- [wire-protocol-mapping] What do boundary direction values (0-3) map to (north/south/east/west)? Is it consistent with NPC/object spawning?
- [wire-protocol-mapping] Where is the RSC Huffman tree defined or is there a canonical reference for string compression?
- [wire-protocol-mapping] Are there explicit transition packets between trade/duel offer-screen and confirm-screen or does client render based on broadcasted accept flags?
- [wire-protocol-mapping] Does scenery state persistence (cut webs, opened doors) survive server restart or is it session-scoped?
- [wire-protocol-mapping] What constraint does isGeneralStore = 0 vs 1 impose (allowed items, sell-to rules)?
- [wire-protocol-mapping] How are opcode conflicts (4, 8, 197, 247) disambiguated permanently post-login, or is this a quirk of the auth handshake only?
- [wire-protocol] Spell ID ordinal vs. spell-def ID routing
- [wire-protocol] Prayer points auto-restore mechanism
- [wire-protocol] Long message paging strategy
- [wire-protocol] Chat range visibility enforcement
- [wire-protocol] PvP spell cast-outside-zones rejection behavior
- [wire-protocol] Prayer mutual exclusion extensibility
- [wire-protocol] Enchantment spell temporary vs. permanent effects
