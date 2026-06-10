# The Wire Protocol — Payload235 shadow (companion to api.md)

> **STATUS: CURRENT — verified 2026-06-10 against branch
> `tidy/structure-and-docs`, HEAD `0bfa818`.** Every opcode below is BUILT —
> encoded in `action/*.go` or decoded in `proto/v235/` — unless a row says
> otherwise. Remaining wire research is in the open-questions ledger at the
> bottom; the *actionable* items there are tracked in
> [`docs/TODO.md`](../TODO.md) (MP-10 research tails, MP-12 sleep/wire
> hardening).

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

- **Phase 0 — session ID.** Server sends (after a brief delay —
  `SESSION_ID_SENDER_TIMER`, default 640ms; `RSCSessionIdSender.java`) a
  raw session ID, no length/opcode wrapper. Built behaviour
  (`consumeOptionalSessionID`, `session/conn.go:114-138`): wait up to
  `PreLoginReadDelay` (default 800ms) and read-and-discard **4 bytes** if
  they arrive; a timeout (our login packet beat the timer, or the server
  suppressed the ID) is also fine — v235 login never uses it. The
  800ms-vs-640ms race hardening is `TODO.md` MP-12.
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

Bitpacked payloads (191, 79) read MSB-first within each byte. (234 is NOT
bitpacked — it's `[short count]` + byte-structured records, §4/§11.)

### Heartbeat

Opcode **67**, empty payload, periodic. The runtime sends one every 5s by
default (`runtime/host.go:438`, `heartbeatLoop`) — live-validated. Without
it the server drops the connection as dead.

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
alone; the decoder reads to byte-granular buffer exhaustion
(`proto/v235/playercoords.go`).

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

Per `GameStateUpdater.java:1128-1175` + `Payload235Generator`. The packet is
a **SEQUENCE of entries, read until the payload is exhausted** — the server
streams **every in-view ground item** per update. Two entry shapes:

```
[0xFF] [offsetX] [offsetY]            3 bytes — out-of-range CLEAR: remove
                                      whatever item is at (x,y). NO id is sent.
[idHi] [idLo] [offsetX] [offsetY]     4 bytes — ADD id at (x,y); if the id
                                      carries the 0x8000 bit it is an
                                      in-range REMOVAL of id & 0x7FFF.
```

`offsetX`/`offsetY` are signed int8, relative to the player's **current**
tile (`item.x - player.x`). Decoder: `decodeGroundItem`
(`proto/v235/inbound.go`) → one `event.GroundItemUpdates` batch;
`world.Apply` resolves offsets against `Self.Position()` and mutates
`world.GroundItems`. Pinned by `proto/v235/ground_item_test.go`.

> **Why multi-entry decode is load-bearing.** The old decoder read only the
> FIRST entry per frame, silently hiding every item the server happened to
> list later — e.g. the free Bronze Pickaxe (id 156) on the Barbarian-Village
> table, which fed an entire epistemic wild-goose chase
> (`docs/world-knowledge-and-learning.md`). The regression test's decisive
> assertion is exactly "the pickaxe decodes even as the second entry."

### Boundary state — opcode 91 (`InBoundaryHandler`, SEND_BOUNDARY_HANDLER)

Dynamic boundary updates (opened doors, cut webs). Each record:
`[short id] [byte offsetX] [byte offsetY] [byte dir]`, coordinates
player-relative. The `id` field is the **DoorDef index** (0-indexed, same id
space as `BoundaryLocs.json`; e.g. 0=Wall, 23=Door, 24=web), so `.at()` can
populate `.name` from `facts.BoundaryDef(id)`.

> **Removal sentinel caveat.** A generic decode convention treats `id == 0xFFFF`
> (→ -1) as "boundary removed", but **the authentic v235 server never emits a
> per-record boundary removal** — it just overwrites in place
> (`GameStateUpdater.updateWallObjects`: the non-custom removal branch is
> commented out; only the custom-client branch emits `60000`). Our decoder's
> `0xFFFF` → -1 branch (`decodeBoundaryUpdates`) is defensive and unreachable
> on an authentic v235 link (`_research/plutonium/findings.md`). For our
> 235-only target, treat opcode-91 records as **add/overwrite only**. The way
> stale boundary state actually leaves the world model is the **opcode-211
> bulk region clear** (below) — door-open state itself rides opcode 91 as an
> overwrite to the door's new def id.

#### Boundary direction byte — the canonical 4-value enum (0–3)

Boundaries (server `GameObjectType.BOUNDARY`, type byte 1) carry a **4-value
direction enum, 0–3** — this is distinct from the 8-way (0/2/4/6) compass used
for *scenery*. The `(x, y, dir)` triple is the literal server lookup key
(`getWallObjectWithDir` requires an **exact** dir match). Server collision
(`World.registerBoundary`) and client (`World.applyWallToCollisionFlags`) are
byte-for-byte identical:

| dir | Geometry | Blocks | Collision flag on (x,y) | Derived on neighbour |
|---|---|---|---|---|
| **0** | wall on the **NORTH edge** of tile (x,y) — segment runs E–W | N↔S | `WALL_NORTH` (1) | `(x, y-1) \|= WALL_SOUTH` (4) |
| **1** | wall on the **EAST edge** of tile (x,y) — segment runs N–S | E↔W | `WALL_EAST` (2) | `(x-1, y) \|= WALL_WEST` (8) |
| **2** | **`\` diagonal** (NW↔SE), drawn (x,z)→(x+1,z+1) | corner | `FULL_BLOCK_A` (16) | — (single tile) |
| **3** | **`/` diagonal** (SW↔NE), drawn (x+1,z)→(x,z+1) | corner | `FULL_BLOCK_B` (32) | — (single tile) |

(`FULL_BLOCK_C` = 64 is the scenery full-block, never used by boundaries.
Boundaries only set collision when `DoorDef.doorType == 1`; doorType 0 is
decorative/no-clip.)

**No N/S-of-tile-above ambiguity.** There is exactly one canonical address for
any boundary: a wall between rows `y-1`/`y` is **always** `(x, y, 0)` (never
"south edge of the tile above" — there is no south dir value); a wall between
columns `x-1`/`x` is **always** `(x, y, 1)`. The server derives the `WALL_SOUTH`
/ `WALL_WEST` flag on the lower neighbour automatically — it is internal
collision bookkeeping, not an addressable boundary. When deriving a boundary
from a target tile, pick the **higher-x / higher-y** tile of the pair.

**`.at(x,y,dir)` must echo the server's exact dir.** No canonicalization, no
N/S aliasing — pass the same `(x,y,dir)` carried on opcode 91 / present in
`BoundaryLocs.json`. A wrong dir → `getWallObjectWithDir` returns null →
`setSuspiciousPlayer(true,…)` and the action silently fails. (Empirical: the
web at `(208,547)` is DoorDef 24, dir **2** — a `\` diagonal — so
`use(knife, world.boundaries.at(208,547,2))` is exactly right. Of 967 locs in
`BoundaryLocs.json`: dir 0→425, dir 1→404, dir 2→73, dir 3→65.)

### Scenery state — opcode 48 (`InSceneryHandler`, SEND_SCENERY_HANDLER)

Dynamic scenery (GameObject) updates: lit fires, depleted rocks, cut/regrown
trees. Same record shape as the boundary handler **minus the direction
byte** — read records to payload exhaustion:

```
[short id] [byte offsetX] [byte offsetY]     offsets signed int8, player-relative
```

`id == 60000` (`world.SceneryRemoveSentinel`) means "the scenery at this tile
is gone" — `GameStateUpdater.updateGameObjects` emits a
`GameObjectLoc(60000, …)` for every object leaving view/removed, and *unlike*
boundaries this sentinel IS authentic for scenery. Decoder:
`decodeSceneryUpdates` (`proto/v235/inbound.go`) → `event.SceneryUpdates`;
`world.Apply` resolves to absolute tiles and feeds `world.Scenery`
(`world/scenery.go`), which keeps a separate `removed`-tile set so the
renderer can suppress the static baseline object (a mined rock must not pop
back). This stream is the ONLY place runtime-spawned scenery exists — a fire
lit by firemaking (def 97) never appears in the static landscape data.

### Bulk region clear — opcode 211 (`InRemoveWorldEntity`, SEND_REMOVE_WORLD_ENTITY)

The server splits entity removal into two channels by distance. Within ±128
tiles, removals ride the in-view per-entity packets (scenery 48 / ground-item
99 sentinels above). **Beyond ±128 tiles** the per-entity offsets can't be
expressed in a signed byte, so the server batches the locations and flushes
them via `sendClearLocations` → opcode 211 — its **only eviction channel for
far-away entities**. Without decoding it, every door/rock/ground-item the
host walks >128 tiles from stays in the world model forever (the original
Plutonium-study gap, `_research/plutonium/findings.md`).

```
repeat to payload exhaustion:
  [short offsetX] [short offsetY]     SIGNED 16-bit, player-relative —
                                      NOT int8 like the per-entity packets
```

Each point names an **8×8 region** (`abs >> 3`), not an exact tile.
`world.Apply` sweeps all three dynamic stores per region —
`Boundaries.RemoveRegion`, `Scenery.RemoveRegion`,
`GroundItems.RemoveRegion` (`world/world.go:1142-1156`) — returning those
tiles to the static baseline until fresh state streams in on re-approach.
The per-tick send order is position → 48 → 91 → ground items → 211, so
`Self.Position()` is current when the offsets resolve. Decoder pinned by
`proto/v235/ground_item_test.go` (`TestDecodeRemoveWorldEntity`).

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
| 194 | `InSleepwordIncorrect` | last sleep-word was wrong — **constant only, no decode case yet** (lands as `UnknownPacket`; the latent infinite-resleep trap, `TODO.md` MP-12) |
| 182 | `InWelcomeInfo` | post-login welcome screen |
| 52 | `InSystemUpdate` | system-update countdown — **constant only, no decode case yet** (lands as `UnknownPacket`) |
| 165 | `InSendLogout` | server confirms clean logout |

### Sleep round-trip — in 117 → **out 45** → in 84 (or in 194)

SEND_SLEEPSCREEN (117) carries the raw captcha image bytes; we don't OCR it.
The answer is **opcode 45, `outSleepWord` (SLEEPWORD_ENTERED)** —
`action/sleep.go`, per `Payload235Parser.java:697-702`:

```
[byte]                sleepDelay   — we send 0 (only the 233-compat client uses it)
[zero-padded string]  sleepWord    — the typed word
```

On this OpenRSC server the word is hardcoded `"asleep"`
(`CaptchaGenerator.java:79-80` when prerendered sleepwords aren't loaded), so
the runtime auto-answers the moment 117 arrives (`runtime/frame.go:130-140`,
`const sleepWord = "asleep"` — correctness depends on that server config; the
guard/log hardening is `TODO.md` MP-12). Success → SEND_STOPSLEEP (84, empty);
the fatigue reset arrives separately as a `FatigueUpdate`. Opcode **244**
(SEND_SLEEP_FATIGUE) has no constant at all yet — also MP-12.

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
non-stackable do not. The single source of truth is
`facts.ItemDef(id).IsStackable`, wired as the `Host.isStackableItem`
closure (`runtime/host.go:516`) and handed to `v235.DecodeInbound` at
`runtime/frame.go:161`.

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

Waypoint cap: 25 corners of deltas (`action.MaxWalkCorners` — the client
caps at `count - 25`). Walk-click range: authentic 15
(`VIEW_DISTANCE*8-1`); the cradle uses 30 (`action.MaxClickRange`,
`action/walk.go:23` — a deliberate testing override, documented in place).

### Use-with-X (`action/use.go`)

Each target class has its own opcode (the client picks at click time):

| Opcode | Const | Payload |
|---|---|---|
| 91 | `outUseItemOnItem` (ITEM_USE_ITEM) | `[short slot1] [short slot2]` |
| 53 | `outUseItemOnGroundItem` (GROUND_ITEM_USE_ITEM) | `[short x] [short y] [short groundItemID] [short slotID]` |
| 115 | `outUseItemOnScenery` (USE_ITEM_ON_SCENERY) | `[short x] [short y] [short slotID]` (no direction byte) |
| 161 | `outUseItemOnBoundary` (USE_WITH_BOUNDARY) | `[short x] [short y] [byte direction] [short slotID]` |
| 135 | `outUseItemOnNpc` (NPC_USE_ITEM) | `[short serverIndex] [short slotID]` |
| 113 | `outUseItemOnPlayer` (PLAYER_USE_ITEM) | `[short serverIndex] [short slotID]` |
| 136 | `outObjectCommand` (OBJECT_COMMAND) | primary click on scenery (opt 1) |
| 79 | `outObjectCommand2` (OBJECT_COMMAND2) | secondary click on scenery (opt 2) |

> **NPC_USE_ITEM is opcode 135, NOT 50** (v235). Opcode 50 is `CAST_ON_NPC`
> (§5) and opcode **135** is `NPC_USE_ITEM` — two **distinct** opcodes, *not* a
> shared byte disambiguated by length (both are length 4). The opcode→meaning
> map is per-revision; in *no* revision is byte 50 overloaded as both. (v235:
> 50/135; v203: 50/135; v198: 49/160; v196: 185/22 — always split.) The cradle
> const `outUseItemOnNpc` must be **135**; `magic.cast(spell, npc)` (opcode 50)
> and `use(item, npc)` (opcode 135) are byte-identical 4-byte payloads
> distinguished *only* by the opcode, so a wrong literal silently turns
> `use(item, npc)` into a bogus `CAST_ON_NPC` (slot read as spellID).
>
> Opcode **53** is both this use-on-ground-item and inbound full-inventory;
> **79** is both OBJECT_COMMAND2 and inbound NpcCoords.

### NPC clicks (`action/combat.go`)

| Opcode | Const | Payload |
|---|---|---|
| 153 | `outNpcTalkTo` | `[short serverIndex]` — `talk_to` |
| 202 | `outNpcCommand` | `[short serverIndex]` — right-click (talk/attack/use by default cmd) |

### Dialog (`action/dialog.go`; inbound 222 / 89 / 245)

| Opcode | Const / dir | Payload |
|---|---|---|
| 222 | `InNpcDialogText` (SEND_BOX, in) | `[zero-quoted string] text` — big dialog box |
| 89 | `InNpcDialogBox` (SEND_BOX2, in) | identical single zero-quoted string — small dialog box; both decode to the same `NpcDialogText` event (`Payload235Generator.java:135-143` writes the same `MessageBoxStruct.message` for both) |
| 245 | `InNpcDialogOptions` (OPTIONS_MENU_OPEN, in) | `[byte num_options]` then per option `[zero-quoted string]` |
| 116 | `outDialogChoice` (QUESTION_DIALOG_ANSWER, out) | `[byte optionIndex]` |

### Boundary interact (`action/boundary.go`)

All four boundary verbs carry `[short x] [short y] [byte direction]` (+ a
trailing field for use/cast) and the server resolves via
`getWallObjectWithDir(Point(x,y), dir)` — the dir must match exactly (see §1
boundary dir-byte table). The "open"/no-item verbs are 14/127; item-on-boundary
is **161** (§3 use-with-X); cast-on-boundary is **180** (§5).

| Opcode | Const | Payload |
|---|---|---|
| 14 | `outInteractWithBoundary` | `[short x] [short y] [byte direction]` — left-click (command1), `player.click=0` |
| 127 | `outInteractWithBoundary2` | same — right-click (command2), `player.click=1` |

> **Locked doors are CONTENT, not a missing opcode.** Key-on-door and
> knife-on-web both ride the **same** opcode 161 (`USE_WITH_BOUNDARY`) — the
> server distinguishes by the item+door, not a separate opcode. The unlock
> messages: correct key → *"you unlock the door"* → *"you go through the door"*;
> wrong/no key on a keyed door (via 161) → *"Nothing interesting happens"*.
> Clicking **Open** (14/127) on a keyed door → *"The door is locked"* (case and
> exact wording vary: *"this door is locked"*, *"The door is locked."*, etc. —
> the literal *"it's locked"* is never emitted). A robust "is it locked?" read
> should **case-insensitively substring-match `locked`** (and `unlock` for
> success) in `world.messages`.

### Examine

**No wire opcode is involved.** Examine text lives in the static defs the
client already has (item/scenery/boundary defs — `facts/`), so an examine
verb is def-side only; nothing crosses the wire. (`runtime/examine.go` is the
location-gazetteer summary, not an examine verb; the DSL verb itself is
unbuilt — `TODO.md` DSL-21.)

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

`dmg`/`curHits`/`maxHits` are each a single byte (0–255). Damage in opcode 234
is **player-only** (no inbound NPC-damage record here). **NPC** current/max HP
comes from opcode **104** (`SEND_UPDATE_NPC`) — see below.

### Inbound NPC events — opcode 104 (`InUpdateNpc`, SEND_UPDATE_NPC)

The per-NPC update packet, decoded by `DecodeUpdateNpcs`. **This is the source
of NPC HP** (the field opcode 234 never carries) — the prerequisite for an
honest `Npc.health` / target-HP perception:

```
[short count]
for each record:
  [short npcIndex]
  [byte  type]
    1 = NPC chat:        [short recipientIndex(signed, -1=none)] [smart charCount] [RSC-compressed body]
    2 = hits update:     [byte dmg] [byte curHits] [byte maxHits]   ← the NPC's OWN current/max HP
    3 = projectile→NPC:    [short projType] [short victimNpcIndex]
    4 = projectile→player: [short projType] [short victimPlayerIndex]
    5 = skull/PK flag:   [byte skullType]            (consumed, not surfaced)
    6 = wield update:    [byte wield] [byte wield2]   (consumed)
    7 = action bubble:   [short bubbleId]             (consumed)
```

Type-2 → `event.NpcDamage{NpcIndex, Damage, CurHits, MaxHits}`. Note: this
arrives only when an NPC takes damage / its hits change, so target-HP
perception is event-driven, not a steady poll.

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

> Opcode reuse: **249** CAST_ON_GROUND vs. inbound bank-update (§8); **99**
> CAST_ON_SCENERY vs. inbound ground-item-handler (§1); **4** CAST_ON_INVENTORY
> vs. FORGOT_PASSWORD (login-state disambiguated). These are resolved by
> direction + login state.
>
> **Opcode 50 is NOT reused.** `outCastOnNpc` = 50 (CAST_ON_NPC) and
> `outUseItemOnNpc` = **135** (NPC_USE_ITEM, §3) are distinct opcodes in v235,
> not a shared byte. CAST_ON_NPC payload is `[u16 npcIndex] [u16 spellID]`;
> NPC_USE_ITEM is `[u16 serverIndex] [u16 slotID]` — identical 4-byte shapes,
> told apart purely by the opcode.

---

## 6. `prayer` — activate / deactivate

`action/prayer.go`. Prayer index 0..13 (Thick Skin, Burst of Strength, …):

| Opcode | Const | Payload |
|---|---|---|
| 60 | `outPrayerActivated` | `[byte prayerID]` |
| 254 | `outPrayerDeactivated` | `[byte prayerID]` |
| 206 | `InPrayersActive` (in) | `[N bytes]`, one byte per prayer slot, value 0/1 |

> **Opcode 206 (`SEND_PRAYERS_ACTIVE`) layout.** No count prefix — the payload
> length **is** the slot count (`N` = 14 in stock RSC, slots 0–13, matching
> `facts.Prayers`). Slot `i` is active iff byte `i == 1`. Sent on any prayer
> toggle and on a full resync. `prayer.active(slot)` indexes this list; the
> active-list = indices where the byte is set. (Server-side this is the only
> source of active-prayer state; the active set is fully on the wire — no
> client-side computation needed. Cradle: `event.PrayersActive{Active []bool}`.)

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
| 15 | `InTradeAccepted` | accept echo / reset — **constant only, no decode case yet** (lands as `UnknownPacket`; the open 15-vs-162 question below stands) |
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
| 42 | `InBankOpen` | SEND_BANK_OPEN — `[byte storedSize] [byte maxBankSize]` then per slot `[short catalogID] [unsignedShortInt amount]` |
| 249 | `InBankUpdate` | SEND_BANK_UPDATE — `[byte slot] [short catalogID] [unsignedShortInt amount]` |
| 203 | `InBankClose` | bank window closed (empty payload) |

> **Bank contents + open-state are fully on the wire.** `bank.slots` ← the
> per-slot `[id, amount]` list from opcode 42, mutated by opcode-249 deltas;
> `bank.used` = `storedSize`; `bank.max_size` = `maxBankSize` (both clamped to
> 255 — a >255-slot custom bank is unrepresentable on this wire). There is no
> in-packet "is bank open" boolean: **`bank.is_open` is a state machine** — set
> true on opcode 42 (open), false on opcode 203 (close), mirroring the client's
> `showDialogBank` flag.
>
> **Amount-width caveat (verify on live capture).** The v235 generator writes
> bank/inventory amounts with `writeUnsignedShortInt` — a **smart** width: 2
> bytes if `amount ≤ 32767`, else 4 bytes with the high bit set. The cradle
> mirrors this (`readUnsignedShortIntSmart`). **However**, the OpenRSC bundled
> GUI client reads the same field with `get32()` (unconditional 4-byte). So the
> generator and the bundled client **disagree** for amounts ≤ 32767. The cradle
> follows the **generator** (the wire authority for v235), but sanity-check
> against whatever server you actually connect to. Same caveat applies to
> inventory amounts (opcode 53/90).

---

## 9. `shop` — open / buy / sell / close

`shop.buy`/`sell`/`stock`/`price` in `api.md`. A shop is opened by interacting
with a shopkeeper NPC; the shop UI must be open server-side
(`player.getShop()` non-null) or the buy/sell packet is flagged suspicious.

### Outbound (`action/shop.go`)

| Opcode | Const | Payload |
|---|---|---|
| 236 | `outShopBuy` (SHOP_BUY) | `[short catalogID] [short shopStock] [short amount]` |
| 221 | `outShopSell` (SHOP_SELL) | `[short catalogID] [short shopStock] [short amount]` |
| 166 | `outShopClose` (SHOP_CLOSE) | (empty) — releases the server-side shop ref |

> `shopStock` is the quantity the client *believes* the shop holds for that
> item — a price-sync sanity hint the server clamps against its real stock.
> Pass `world.Shop.Stock(catalogID)`; `0` is acceptable (server clamps anyway).

### Inbound

| Opcode | Const | Payload |
|---|---|---|
| 101 | `InShopOpen` | SEND_SHOP_OPEN (see layout below) |
| 137 | `InShopClose` | shop window closed |

**Opcode 101 (`SEND_SHOP_OPEN`) layout:**

```
[byte] shopItemCount     — number of catalogue entries
[byte] isGeneralStore    — 1 = general store, 0 = specialty
[byte] sellPriceMod      — unsigned, base sell percentage
[byte] buyPriceMod       — unsigned, base buy percentage
[byte] stockSensitivity  — unsigned, price-drift multiplier
for each entry:
  [short catalogID]
  [short stock]          — current stock
  [short baseStock]      — baseline stock (reference for stock-sensitive pricing; NOT a gp value)
```

The gp price is derived client-side from the item-def base price via the
pricing formula; the runtime layer recomputes it (`world.ShopState.BuyPrice`).
`shop.is_open` is a state machine over opcode 101 (open) / 137 (close), same
pattern as the bank.

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
muted chat (`[smart char-count]` + RSC-compressed body — the leading icon
byte is written when `updateType != 6`, i.e. for types 1 and 7 but NEVER
quest chat; reading it unconditionally garbles quest-chat messages); `5` =
appearance / identity (full v235 layout decoded —
`proto/v235/updateplayers.go` case 5: appearanceID, name×2, worn-sprite
block, 4 colour bytes, combatLevel, skullType).

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
| 33 | DuelOfferItem | 38 | Command | 45 | SleepWord |
| 46 | AddTradeItems | 50 | CastOnNpc | 53 | UseItemOnGroundItem |
| 55 | AcceptTradeOffer | 60 | PrayerActivated | 67 | Heartbeat |
| 77 | DuelSecondAccepted | 79 | ObjectCommand2 | 90 | ItemCommand |
| 91 | UseItemOnItem | 99 | CastOnScenery | 102 | Logout |
| 103 | InitDuelRequest | 104 | AcceptTradeConfirm | 113 | UseItemOnPlayer |
| 115 | UseItemOnScenery | 116 | DialogChoice | 127 | InteractWithBoundary2 |
| 135 | UseItemOnNpc | 136 | ObjectCommand | 137 | CastOnSelf |
| 142 | InitTradeRequest | 153 | NpcTalkTo | 158 | CastOnLand |
| 161 | UseItemOnBoundary | 165 | PlayerFollow | 166 | ShopClose |
| 167 | RemoveFriend | 169 | ItemEquip | 170 | ItemUnequip |
| 171 | PlayerAttack | 176 | DuelFirstAccepted | 180 | CastOnBoundary |
| 187 | WalkToPoint | 190 | NpcAttack | 195 | AddFriend |
| 197 | DuelDeclined | 202 | NpcCommand | 212 | BankClose |
| 216 | ChatMessage | 218 | PrivateMessage | 221 | ShopSell |
| 229 | CastOnPlayer | 230 | DeclineTrade | 236 | ShopBuy |
| 246 | ItemDrop | 247 | GroundItemTake | 249 | CastOnGround |
| 254 | PrayerDeactivated | | | | |

## Inbound opcode quick-index

| Op | Const | Op | Const | Op | Const |
|---|---|---|---|---|---|
| 6 | DuelItems | 15 | TradeAccepted | 20 | TradeOpenConfirm |
| 30 | DuelSettings | 33 | Experience | 42 | BankOpen |
| 48 | SceneryHandler | 52 | SystemUpdate | 53 | Inventory |
| 79 | NpcCoords | 83 | Death | 84 | StopSleep |
| 89 | NpcDialogBox | 90 | InventorySlotUpdate | 91 | BoundaryHandler |
| 92 | TradeWindow | 97 | TradeOtherItems | 99 | GroundItemHandler |
| 101 | ShopOpen | 104 | UpdateNpc | 114 | Fatigue |
| 117 | SleepScreen | 120 | PrivateMessage | 123 | InventoryRemoveItem |
| 128 | TradeClose | 131 | ServerMessage | 137 | ShopClose |
| 156 | Stats | 159 | Stat | 162 | TradeOtherAccepted |
| 165 | SendLogout | 172 | DuelConfirmWindow | 176 | DuelWindow |
| 182 | WelcomeInfo | 191 | SendPlayerCoords | 194 | SleepwordIncorrect |
| 203 | BankClose | 206 | PrayersActive | 210 | DuelAccepted |
| 211 | RemoveWorldEntity | 222 | NpcDialogText | 225 | DuelClose |
| 234 | SendUpdatePlayers | 245 | NpcDialogOptions | 249 | BankUpdate |
| 253 | DuelOtherAccepted | | | | |

---

## Open questions / unverified

> Still-open items below are research tails, not blockers. The actionable
> subset is tracked in [`docs/TODO.md`](../TODO.md): **MP-10** (low-pri
> protocol research tails — it cites this ledger directly) and **MP-12**
> (sleep/wire hardening: decode 194, decide 244). Struck items carry their
> resolution inline.

- ~~[wire-protocol-interactions] Bitpack encoding details for PlayerCoords and UpdatePlayers (field order, bit widths)~~ **RESOLVED** — both fully decoded: 191's bit layout is in §1 (`proto/v235/playercoords.go`); 234 is byte-structured records, all eight update types decoded (`proto/v235/updateplayers.go`, pinned by `updateplayers_test.go`).
- [wire-protocol-interactions] NPC Dialog color tag format in opcode 222 (hex vs named vs other)
- [wire-protocol-interactions] Appearance Screen customization payload format (opcode 59 → opcode 235 response)
- [wire-protocol-interactions] Private Message icon sprite ID mapping to player rank/status
- ~~[wire-protocol-interactions] Boundary interaction direction byte encoding (north/east/south/west values)~~ **RESOLVED** — 4-value enum 0–3 (0=N-edge, 1=E-edge, 2=`\` diag, 3=`/` diag); see §1 boundary dir-byte table.
- [wire-protocol-interactions] Exact XTEA key derivation from ISAAC keys; confirm first 4 uint32s used
- ~~[wire-protocol-interactions] Opcode disambiguation when same byte appears in multiple contexts (e.g., 50: CAST_ON_NPC vs NPC_USE_ITEM by length/context)~~ **RESOLVED** — not shared in v235: 50=CAST_ON_NPC, 135=NPC_USE_ITEM (distinct opcodes, not length-disambiguated); see §3/§5.
- [wire-protocol-mapping] NPC index correlation across despawn/respawn cycles
- ~~[wire-protocol-mapping] Full appearance block layout (outfit, colors, animations, equipment) for opcode 234 type 5~~ **RESOLVED** — decoded in `updateplayers.go` case 5: `[short appearanceID] [zqstring name]×2 [byte N + N worn-sprite bytes] [4 colour bytes: hair/top/trouser/skin] [byte combatLevel] [byte skullType]` (per `GameStateUpdater.issuePlayerAppearanceUpdatePacket`).
- ~~[wire-protocol-mapping] Projectile visual fields for opcode 234 types 3 and 4~~ **RESOLVED** — `[short projectileType] [short victimIndex]` → `event.OtherPlayerProjectile` (`updateplayers.go` cases 3/4; §4).
- [wire-protocol-mapping] Smart encoding variance cases (non-standard length headers)
- [wire-protocol-mapping] ISAAC cipher sync edge cases post-login
- ~~[wire-protocol-mapping] Boundary direction field: 4-way vs 8-way compass encoding~~ **RESOLVED** — 4-way (0–3) for boundaries; the 8-way (0/2/4/6) compass applies only to scenery. See §1.
- ~~[wire-protocol-mapping] Server-side stackable item definition mechanism~~ **RESOLVED** — the item def's stackable flag from the defs data; surfaced as `facts.ItemDef(id).IsStackable` (`facts/defs.go:121`).
- ~~[wire-protocol-research] ItemUseOnGroundItem exact field order and wire alignment~~ **RESOLVED** — `[short x] [short y] [short groundItemID] [short slot]` per Payload235Parser GROUND_ITEM_USE_ITEM (`action/use.go:97-122`).
- ~~[wire-protocol-research] Confirm opcode 249 is CAST_ON_GROUND_ITEM in server payload parser~~ **RESOLVED** — `outCastOnGround = 249`, written from Payload235Parser (`action/magic.go:17`).
- ~~[wire-protocol-research] Verify facts.ItemStackability is the single source of truth for amount encoding in inventory packets~~ **RESOLVED** — `Host.isStackableItem` (`runtime/host.go:516`, backed by `facts.ItemDef.IsStackable`) is the only amount-width authority, handed to `DecodeInbound` at `runtime/frame.go:161`; a nil callback degrades to all-unstackable (safe under-read).
- [wire-protocol-research] Player position sprite encoding: confirm if sprite is 4 bits or if subType bits alias with sprite bits across versions
- ~~[wire-protocol-research] Ground item opcode 99 payload structure: confirm 0xFF + id vs other removal markers~~ **RESOLVED** — `0xFF` carries NO id: it is a 3-byte coord-only clear `[255][x][y]`; in-range removal is the `0x8000` high bit on the id. See §1; pinned by `proto/v235/ground_item_test.go`.
- [wire-protocol-research] Duel request detection via substring parsing in type-7 SEND_SERVER_MESSAGE is fragile; confirm server message format stability
- [wire-protocol-research] Trade/duel requests via SEND_SERVER_MESSAGE (type 6 and type 7) should ideally have dedicated opcodes to avoid parsing fragility
- ~~[wire-protocol-documentation] Ground item removal packet format (opcode 99 0xFF sentinel behavior)~~ **RESOLVED** — see §1 ground items: `[255][x][y]` out-of-range clear (no id) + `0x8000` in-range removal, multi-entry to payload exhaustion.
- ~~[wire-protocol-documentation] Exact smart encoding threshold for inventory amounts in Payload235Generator~~ **RESOLVED** — the high bit of the first byte: clear → 2-byte u16 (values ≤ 32767); set → 4-byte int masked `& 0x7FFFFFFF` (`readUnsignedShortIntSmart`, `proto/v235/strings.go:46`).
- [wire-protocol-documentation] Bank preset opcodes (91, 92) wire format and availability by world/client
- [wire-protocol-documentation] Duel combat transition signal (no explicit packet, inferred from state)
- [wire-protocol-documentation] Client version item ID upper limits and ItemDef.maxId enforcement
- ~~[wire-protocol] Examine interaction opcode and payload structure — whether it uses a dedicated opcode or reuses ItemCommand (90) with a special command variant.~~ **RESOLVED (moot)** — examine is def-side; no wire opcode exists or is needed (§3 Examine). The DSL verb itself is open as `TODO.md` DSL-21.
- ~~[wire-protocol] Boundary interact (opcodes 14, 127) outbound payload structure in cradle — assumed to match TargetObjectStruct but not yet verified in action/*.go.~~ **RESOLVED** — `[short x] [short y] [byte direction]` (`sendBoundaryPacket`, `action/boundary.go:47-49`).
- ~~[wire-protocol] GroundItemHandler (opcode 99) multi-entry packet semantics — does server ever pack multiple add/remove records in a single frame?~~ **RESOLVED** — yes, routinely: the server streams every in-view ground item per update; the decoder reads to payload exhaustion (§1; the hidden-pickaxe regression).
- [wire-protocol] ItemCommand custom-client amount field semantics — does amount mean 'perform N times' or 'set resulting stack to N'?
- [wire-protocol] Equipment-tab fallback logic in ItemOnObjectStruct.itemID — exact conditions for client/server using itemID-based lookup vs. slotID.
- ~~[wire-protocol] Exact opcode number for GROUND_ITEM_TAKE on outbound (action/interaction.go) — confirmed as 247 in server but not yet read from cradle source.~~ **RESOLVED** — `outGroundItemTake = 247` (`action/interaction.go:17`).
- [wire-protocol] Option 30 behavior: Is it a legacy RSC quirk or explicit cancel code?
- [wire-protocol] Duel/trade completion: Should success detection rely on inventory deltas or explicit packet?
- [wire-protocol] Menu truncation: Should server reject >5 options or silently truncate?
- [wire-protocol] NPC busy timeout: Is 20-second multi-timeout hardcoded or config-driven?
- [combat-wire-protocol] Damage capping: byte field (0–255) in opcode 234 type 2 cannot represent damages >255; does server split large damages?
- ~~[combat-wire-protocol] NPC health bars: no inbound damage event for NPCs in opcode 234 type 2 (player-only); how are NPC bars displayed?~~ **RESOLVED** — NPC HP rides opcode **104** (SEND_UPDATE_NPC) type 2 → `event.NpcDamage` (§4; `proto/v235/updatenpcs.go`).
- [combat-wire-protocol] Retro vs. custom client differences: do opcode or field layouts change between client versions (retro 38–39, authentic 61–204, custom)?
- [combat-wire-protocol] Combat event sequencing: is damage update (opcode 234 type 2) sent before/after/during health sprite change on opcode 79 or 234?
- [combat-wire-protocol] Update-type 7 authenticity: GameStateUpdater marks type 7 as 'not authentic' but generates it; is this OpenRSC-only or incompatible with authentic client?
- ~~[trade-wire-subsystem] Are trade payload fields (item IDs, amounts) ISAAC-encrypted before wire transmission or handled raw by the action layer?~~ **RESOLVED** — only the opcode byte is ISAAC-touched; ALL payloads ride plaintext (§0 framing; `proto/v235/isaac.go`).
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
- ~~[wire-protocol-mapping] What do boundary direction values (0-3) map to (north/south/east/west)? Is it consistent with NPC/object spawning?~~ **RESOLVED** — 0=N-edge of (x,y), 1=E-edge of (x,y), 2=`\` diagonal, 3=`/` diagonal. NOT the same as scenery's 8-way spawn dir. See §1.
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
