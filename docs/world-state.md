# World-State Mirror, Event Model & Connection

> **Status: IMPLEMENTED.** The `world/`, `event/`, `proto/v235/`, `session/`,
> and `runtime/host.go` layers described here are working code, driving the
> live cradle against OpenRSC. This doc is the architectural map for anyone
> touching server/protocol/state code. Where a field or path is reserved /
> stubbed, it is flagged inline.

This is the layer that turns a TCP byte stream from an OpenRSC server into a
queryable snapshot of "what does my agent believe about the world right now?"
Three pieces, in dataflow order:

1. **[Connection + login](#1-connection--login-handshake)** (`session/`, `proto/v235/login.go`) — dial, handshake, ISAAC, framed duplex I/O.
2. **[Event model](#2-event-model)** (`proto/v235/`, `event/`) — raw frame → typed event; the decode↔state bridge and the pub/sub bus.
3. **[World-state mirror](#3-world-state-mirror)** (`world/`) — `world.Apply(event)` mutates the per-host snapshot; routines read it back.

The net loop in `runtime/host.go` (`handleFrame`) stitches them:

```
conn.Recv() → v235.Decode*(frame) → event.Event → world.Apply(ev) → bus.Publish(ev)
   (bytes)        (proto/v235)         (event)      (mutate mirror)   (notify subscribers)
                                                            ↓
                                              DSL routines read world.* / subscribe bus
```

Related docs: [protocol.md](protocol.md) (wire framing/ISAAC/login bytes),
[lang/protocol.md](lang/protocol.md) (faculty→opcode map),
[server-config.md](server-config.md) (the OpenRSC server it talks to),
[architecture.md](architecture.md) (where this sits in the whole system).

---

## 1. Connection + login handshake

**Code:** `session/conn.go`, `session/handshake.go`, `proto/v235/login.go`,
`proto/v235/rsa.go`, `proto/v235/xtea.go`, `proto/v235/isaac.go`.
**Driver:** `runtime/host.go` `Host.Connect`.

### Where the cradle connects

The cradle connects to a **local OpenRSC server**. There are two ports in play
and they are easy to confuse:

| Caller | Default server | Source |
|---|---|---|
| `cmd/cradle` (`-server` flag) | `localhost:43596` | `cmd/cradle/main.go:77` |
| scenariogen swarm scripts | `localhost:43594` | `cmd/scenariogen/run_*.sh` |

Both point at the same OpenRSC instance; pass `-server` explicitly to be sure.
See [server-config.md](server-config.md) for the server side.

### Where the password comes from

The account password is **never embedded** in code or docs. The cradle reads it,
in priority order: the `-password` flag, then the **`WESTWORLD_PASSWORD` env var**
(`cmd/cradle/main.go:123-129`); missing → the process exits with an error.
`.local.env` holds it for local runs and the scenariogen scripts hard-fail if
it is unset. Refer to it only as the `WESTWORLD_PASSWORD` env var.

### The handshake, phase by phase

`Host.Connect` calls `session.Dial` then `conn.Login`, then `conn.Start`.

- **Phase 0 — dial + optional session ID** (`session.Dial`, `consumeOptionalSessionID`).
  TCP dial (5s default). Then a deadline-bounded read of up to 4 bytes for an
  *optional* session ID, waited on for `PreLoginReadDelay` (default **800ms**;
  the server's `SESSION_ID_SENDER_TIMER` is ~640ms). If bytes arrive they are
  **discarded** (v235 login does not need them); a read timeout is **not** an
  error — we just proceed. A premature EOF *is* fatal.

- **Phase 1 — login packet** (`conn.Login` → `LoginPayload.EncodeLoginFrame`).
  Sent **plain** (opcode 0, `OutLogin`), before ISAAC is active. The payload
  carries an **RSA-encrypted block** (checksum=10, four ISAAC/XTEA keys, the
  20-byte space-padded password with spaces→underscores, and nonces) and an
  **XTEA-encrypted block** (a limit byte, six more nonces, the username
  null-padded to an 8-byte boundary). The four ISAAC keys are random per login
  and are the commitment that seeds the cipher. See `proto/v235/login.go`
  (`BuildRSABlock`, `BuildXTEABlock`) for exact byte layout.

- **Phase 2 — response** (`conn.Login` reads 1 raw byte, 10s deadline).
  `v235.LoginSuccessful(code)` is `(code & 0x40) != 0`. Anything else is a
  rejection (codes: 0 unsuccessful, 3 invalid creds, 4 already-logged-in, 7
  attempts-exceeded) and `Login` returns an error.

- **Phase 3 — ISAAC seeding** (`conn.SetIsaacKeys`).
  On success both ciphers are seeded from the four keys we committed to in the
  RSA block, and the frame decoder swaps from `PlainDecode` to ISAAC opcode
  decode. From here every framed opcode byte is ISAAC-encrypted.

After login, `conn.Start()` launches the read + write goroutines (`readLoop`,
`writeLoop`), turning the `Conn` into a duplex channel pair: `conn.Send(opcode,
payload)` out, `conn.Recv()` in.

### RSA server key

The server's public key (`modulus` + `exponent 65537`) is **hardcoded** in
`proto/v235/rsa.go` `DefaultServerRSA()` — captured from the running OpenRSC
server (which persists it to `server/{client,server}.pem`). OpenRSC uses **raw
RSA** (no PKCS#1 padding; `plaintext^e mod n`). If the server is reset and
regenerates its key, this constant must be updated, or login fails. (There is no
runtime fetch/verify; the constant is the source of truth.)

### Heartbeat + termination

`Host.Run` spawns `heartbeatLoop`, sending `OutHeartbeat` (opcode 67, empty)
every `HeartbeatInterval` (default **5s**). **If a heartbeat write fails the
loop logs a warning and returns — it is not retried.** The main loop ends when
`ctx` is cancelled or `conn.Recv()` closes (remote EOF or a decode/IO error
surfaced via `conn.Err()`). There is **no automatic reconnect**: a dropped
connection terminates the host. (Clean logout has its own subtlety — the server
sends no logout-confirm and refuses logout while in combat; see `LogoutGraceful`
in `runtime/host.go`.)

---

## 2. Event model

**Code:** `event/events.go` (the ~40 typed events), `event/bus.go` (pub/sub),
`proto/v235/inbound.go` + the multi-event decoders. **Driver:**
`runtime/host.go` `handleFrame`.

Typed events are the **decode↔state bridge**. A decoder's only job is to turn a
frame's bytes into one or more concrete `event.Event` values; it never touches
world state. Everything downstream (the world mirror, DSL `on`-handlers,
observers) consumes events, not bytes. This keeps the wire format isolated in
one package.

### The Event interface

Every event embeds `base` (a timestamp) and implements:

```go
type Event interface {
    Kind() string    // stable routing/logging id, e.g. "npc_damage"
    Time() time.Time
}
```

`Kind()` is the bus routing key and the DSL `on <kind>` binding name.

### Decode pipeline: single-event vs. multi-event

`handleFrame` splits inbound frames two ways:

- **Multi-event opcodes** are special-cased because one frame yields *many*
  events. Each has a dedicated decoder returning a slice:

  | Opcode | Const | Decoder | Emits |
  |---|---|---|---|
  | 191 | `InSendPlayerCoords` | `DecodePlayerCoords` | one `OwnPositionUpdate` + N `NearbyPlayerEvent` |
  | 234 | `InSendUpdatePlayers` | `DecodeUpdatePlayers` | per-record: `OtherPlayerAppearance`, `OtherPlayerChat`, `OtherPlayerDamage`, `OtherPlayerProjectile`, `PlayerActionBubble` |
  | 104 | `InUpdateNpc` | `DecodeUpdateNpcs` | per-record: `NpcNearby`, `NpcDamage`, `NpcChat`, `NpcProjectile` |
  | 79 | `InNpcCoords` | `DecodeNpcCoords` | per-record: `NpcNearby` (new / move / remove) |

  Opcode 79's decoder is **stateful**: it takes the current ordered local-NPC
  index list (`world.Npcs.Order()`) so positional update slots map back to the
  right NPC index. The host snapshots the order *before* `Apply` mutates it.
  Opcode 191 / 79 also need the player's current position to resolve relative
  offsets to absolute tiles, so the host passes `Self.Position()` in.

- **Single-event opcodes** fall through to `v235.DecodeInbound(frame,
  isStackable)`, a big opcode switch returning one event (or
  `event.UnknownPacket` for opcodes not yet decoded, or `nil`). This is the
  single chokepoint for all message-bearing opcodes (131 server/chat, 120 PM,
  222/89 dialog, 245 menu), which is why `logServerMessage` taps here.

A handful of opcodes are handled inline in `handleFrame` rather than via a
decoder (e.g. 117 `InSleepScreen` — the host also auto-answers the sleep-word
captcha "asleep" — and 84 `InStopSleep`).

### Synthetic (host-derived) events

Some events are **not** decoded from any single packet — the host synthesizes
them by diffing world state across an `Apply`, because the wire carries totals,
not deltas:

| Event | Derived how (in `host.go`) |
|---|---|
| `ItemGained` | inventory count diff before/after an inventory packet |
| `XPGain` | per-skill xp diff before/after an xp-bearing packet (raw packets carry the new total) |
| `TargetDied` | watches `NpcDamage` on the engaged NPC index; fires on the alive→0-hits edge |

These give DSL routines clean "this just happened" hooks
(`on item_gained`, `on xp_gain`, `on npc_killed`).

### The bus (`event/bus.go`)

In-memory typed pub/sub. `Subscribe(kind, buffer)` returns a buffered receive
channel; `"*"` subscribes to everything. `Publish` is **non-blocking and
lossy**: if a subscriber's channel is full the event is **dropped for that
subscriber** (no head-of-line blocking). Routines must drain promptly or accept
drops. `Bus.Close()` closes all subscriber channels.

> **Ordering guarantee:** in `handleFrame`, `world.Apply(ev)` runs **before**
> `bus.Publish(ev)`. So any subscriber that reacts to an event sees a world
> state already consistent with it (e.g. an `on npc_killed` handler can read the
> NPC's final state). Multi-event frames apply+publish each event in wire order.

### Observability is NOT here yet

The bus is the intended subscription point for an observability substrate (event
log, technician UI). **As of now no such subscriber exists** — `obs/` is empty
and `cmd/delos` does not exist. See [observability.md](observability.md) (a
Phase 3+ *roadmap*, not a spec). Today the only "observer" is the host's own
`slog` logging and the DSL routine subscriptions.

---

## 3. World-state mirror

**Code:** `world/world.go` (`World`, `Apply`, and the `Npcs`/`Players`/
`GroundItems` mirrors), plus one file per sub-mirror (`self.go`, `inventory.go`,
`trade.go`, `duel.go`, `bank.go`, `shop.go`, `recent.go`, `boundaries.go`,
`scenery.go`).

`World` is the composition of every per-host state mirror — the single object
higher layers consult. One `*World` per host, created by `NewWorld()` (all
sub-mirrors start empty until packets populate them).

```go
type World struct {
    Self        *Self              // our own player
    Inventory   *Inventory         // our 30-slot inventory
    Npcs        *NpcsState         // NPCs in view (+ ordered list for opcode 79)
    Players     *PlayersState      // other players in view (index 0 == us)
    GroundItems *GroundItemsState  // visible ground items, keyed by tile
    Recent      *RecentEvents      // ring of recent chat/PM/damage/server-msg/dialog
    Trade       *TradeState        // trade-window state machine
    Duel        *DuelState         // duel-window state machine
    Bank        *BankState         // bank screen contents
    Shop        *ShopState         // shop catalogue
    Boundaries  *DynamicBoundaries // dynamic doors/walls (absolute tiles)
    Scenery     *DynamicScenery    // dynamic GameObjects (fires, depleted rocks)
}
```

### Sub-mirror status

| Sub-mirror | Tracks | Fed by (event → opcode) |
|---|---|---|
| `Self` | position, heading, 18 skills (cur/max/xp), fatigue, quest pts, sleep flag, active prayers, own worn-sprite/colours, death tile + count | `OwnPositionUpdate`(191), `StatUpdate`(159)/`StatsSnapshot`(156)/`ExperienceGain`(33), `FatigueUpdate`(114), `Sleep*`(117/84), `PrayersActive`(206), `Death`(83), own slot of `OtherPlayerAppearance`(234) |
| `Inventory` | per-slot item id / amount / wielded | `InventorySnapshot`(53), `InventorySlotUpdate`(90), `InventoryRemoveSlot`(123) |
| `Npcs` | per-index position, type, heading, **ordered local list** for opcode 79; reserved combat fields | `NpcNearby`(79), `NpcDamage`(104), `NpcProjectile`(104) |
| `Players` | per-index position, name, heading, combat level/skull, worn sprites, colours, hits, engagement | `NearbyPlayerEvent`(191), `OtherPlayerAppearance`/`Damage`/`Projectile`(234) |
| `GroundItems` | visible items keyed by absolute (x,y) | `GroundItemEvent`(99) |
| `Recent` | most-recent chat, PM, damage, server message, dialog text/options + a messages ring | `ChatReceived`/`SystemMessage`(131), `PrivateMessage`(120), `NpcChat`(104), `NpcDialogText`(222/89), `NpcDialog`(245) |
| `Trade` | offer/confirm/closed phase machine + both offers + accept flags | `TradeOpened`(92), `TradeOtherOffer`(97), `TradeOtherAccepted`(162), `TradeConfirmShown`(20), `TradeClosed`(128) |
| `Duel` | same shape as Trade + rule toggles | `DuelOpened`(176), `DuelOtherOffer`(6), `DuelSettingsUpdate`(30), `DuelOtherAccepted`(253/210), `DuelConfirmShown`(172), `DuelClosed`(225) |
| `Bank` | open flag, max size, slot contents | `BankOpened`(42), `BankSlotUpdate`(249), `BankClosed`(203) |
| `Shop` | catalogue, price modifiers, general-store flag | `ShopOpened`(101), `ShopClosed`(137) |
| `Boundaries` | dynamic doors/walls at absolute tiles | `BoundaryUpdates`(91) |
| `Scenery` | dynamic GameObjects at absolute tiles | `SceneryUpdates`(48) |

### `World.Apply(ev)` — the one mutation point

`Apply` (`world/world.go`) is a single type switch over `event.Event`. Each case
routes to the relevant sub-mirror's setter and returns `true` if the world
meaningfully changed (the host uses that only for a debug log). **This is the
only place that mutates world state** — decoders and the bus never write. A few
cases do more than a straight write:

- **Relative→absolute resolution.** Ground items (99), boundaries (91), scenery
  (48) arrive as offsets relative to the player *at delivery time*, so `Apply`
  reads `Self.Position()` and resolves to absolute tiles before storing. NPC
  movement (79) is a one-tile **relative** step (`MoveBy`) applied to the NPC's
  *own* stored position — basing it on the host's position caused the historical
  "phantom-crowd" bug.
- **Combat-field preservation.** An NPC's per-tick position update carries no
  health/engagement data, so `NpcsState.Set`/`Move` deliberately **carry
  forward** accumulated combat fields rather than clobbering them.
- **State-machine inference.** Trade/duel `*Closed` packets carry no
  "completed" bit, so `Apply` infers success from the record's accept flags
  (e.g. we reached our second accept ⇒ treat the close as a completed trade).
- **Self is player index 0.** OpenRSC always lists our own player at index 0
  (`SelfPlayerIndex`), so its appearance / damage updates ride in the same
  opcode-234 stream and `Apply` mirrors the relevant bits onto `Self`.

### Thread safety

Every sub-mirror carries its own `sync.RWMutex`; all accessors lock
appropriately, so reads and writes are goroutine-safe. The net loop writes via
`Apply` on the read goroutine's call path while DSL routines (other goroutines)
read concurrently — that is the whole point of the locks. **Read-side accessors
return snapshots/copies** (e.g. `Npcs.All()`, `Players.FindByName`,
`Inventory.Slots()`), never internal references, so a caller can't observe a
half-applied mutation. There is no cross-mirror transaction: each event's
`Apply` is the unit of consistency, and within one event multiple sub-mirror
writes are not atomic relative to a concurrent reader of a *different* mirror.

### Read-side API (what routines query)

Routines never see events unless they explicitly subscribe; mostly they read the
mirror. Examples (Go accessor → conceptual DSL surface):

- `world.Self.Position()`, `.SkillLevel(id)`, `.Fatigue()`, `.IsSleeping()`
- `world.Inventory.Slots()`, `.Count(itemID)`
- `world.Players.FindByName(name)`, `.All()`, `.Self()` (our index-0 record)
- `world.Npcs.All()`, `.Get(index)`, `.Order()` (the opcode-79 mirror)
- `world.GroundItems.Near(cx, cy, radius)`
- `world.Recent.ServerMessage()`, `.Chat()`, `.DialogOptions()`
- `world.Trade.Trade()` / `world.Bank.IsOpen()` / `world.Shop` for the screens

See [lang/api.md](lang/api.md) / [lang/protocol.md](lang/protocol.md) for the
DSL-facing names.

### Known limitations

- **No continuous poll.** The mirror is purely event-driven. State only changes
  when the server sends a packet — e.g. an NPC's `LastSeen` ages but its
  position does not until the server re-sends it. Walk-progress loops poll
  `Self.Position()` precisely because there is no tick callback.
- **Sprites, not item ids.** Worn-equipment in `OtherPlayerAppearance` /
  `PlayerRecord.EquipBySlot` are appearance **sprite** ids (low byte of
  `getAppearanceId()`), not catalogue item ids — there is no sprite→item lookup
  yet (mapping gap).
- **No removal aging.** Records are removed only when the server explicitly says
  so (opcode-79 `REMOVE_NPC`, player-removed flag). There is no TTL sweep, so a
  record whose removal packet was missed lingers until overwritten.
- **No reconnect / no persistence.** The mirror lives entirely in memory for one
  connection; on disconnect the host (and its `World`) is gone.
