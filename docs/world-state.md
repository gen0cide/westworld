# World-State Mirror, Event Model & Connection

> **STATUS: BUILT — verified against code 2026-06-10, branch `tidy/structure-and-docs`,
> HEAD `0bfa818`.** The `world/`, `event/`, `proto/v235/`, `session/`, and
> `runtime/host.go` layers described here are working code, driving live hosts
> against OpenRSC. This doc is the architectural map for anyone touching
> server/protocol/state code. Where a field or path is reserved / stubbed, it
> is flagged inline. Open work is tracked in [`TODO.md`](TODO.md) (the SSOT —
> IDs like `MP-1`/`DSL-20` below refer to its items); this doc carries no
> backlog of its own.

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

### Where a host connects

Every entry point talks to a **local OpenRSC server**; two default ports are
in play and they are easy to confuse:

| Caller | Default server | Source |
|---|---|---|
| `cmd/host` (`-server` flag) | `localhost:43594` | `cmd/host/main.go:69` |
| `cmd/cradle-server` (per-host `server:` in `*.hostcfg`) | `localhost:43594` | `cradle/hostcfg/hostcfg.go:38` (`DefaultServer`) |
| `cmd/legacy-cradle` (`-server` flag) | `localhost:43596` | `cmd/legacy-cradle/main.go:81` |
| scenariogen swarm scripts | `localhost:43594` | `cmd/scenariogen/run_*.sh` |

All point at the same OpenRSC instance; pass the address explicitly to be
sure. See [server-config.md](server-config.md) for the server side.

### Where the password comes from

The account password is **never embedded** in code, docs, or config files.
`cmd/host` and `cmd/legacy-cradle` read it, in priority order: the `-password`
flag, then the **`WESTWORLD_PASSWORD` env var**; missing → the process exits
with an error. `cmd/cradle-server` never takes a password at all: each host's
`*.hostcfg` names the env var that holds it (`password_env`, default
`WESTWORLD_PASSWORD` — `cradle/hostcfg/hostcfg.go`), resolved at launch; an
inline `password:` key is **rejected** by the strict parser. `.local.env`
holds the value for local runs and the scenariogen scripts hard-fail if it is
unset. Refer to it only as the `WESTWORLD_PASSWORD` env var.

### The handshake, phase by phase

`Host.Connect` calls `session.Dial` then `conn.Login`, then `conn.Start`.

- **Phase 0 — dial + optional session ID** (`session.Dial`, `consumeOptionalSessionID`).
  TCP dial (5s default). Then a deadline-bounded read of up to 4 bytes for an
  *optional* session ID, waited on for `PreLoginReadDelay` (default **800ms**;
  the server's `SESSION_ID_SENDER_TIMER` is ~640ms). If bytes arrive they are
  **discarded** (v235 login does not need them); a read timeout is **not** an
  error — we just proceed. A premature EOF *is* fatal. (Hardening the
  800ms-vs-640ms race is TODO `MP-12`.)

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

### Heartbeat, termination & restart

`Host.Run` spawns `heartbeatLoop`, sending `OutHeartbeat` (opcode 67, empty)
every `HeartbeatInterval` (default **5s**). **If a heartbeat write fails the
loop logs a warning and returns — it is not retried.** The main loop ends when
`ctx` is cancelled or `conn.Recv()` closes (remote EOF or a decode/IO error
surfaced via `conn.Err()`).

The **Host itself never reconnects** — a dropped connection terminates
`Host.Run`. What happens next depends on the runner:

- A server-initiated close (idle-kick, server bounce, network drop) surfaces
  as **`session.ErrServerClosed`** (`session/conn.go`), deliberately distinct
  from a local `Close()`: it makes `Host.Run` return an *error*, not a clean
  stop. (Before this distinction, a multi-day soak silently bled the fleet to
  zero as idle-kicked hosts were marked "stopped".)
- Under **`cmd/cradle-server`**, the per-host supervisor
  (`cradle/registry.go` `supervise`) applies the host's `supervision:` policy:
  `restart` relaunches on error with exponential backoff (1s doubling, capped
  30s); `hold` (the default) keeps the dead host for inspection. A relaunch is
  a fresh login and a **fresh `World`**.
- Under **`cmd/host`**, the process simply exits.

Clean logout has its own subtlety — the server sends no logout-confirm and
refuses logout while in combat; see `LogoutGraceful` in `runtime/host.go`.

---

## 2. Event model

**Code:** `event/events.go` + `event/agent.go` (the ~60 typed events),
`event/bus.go` (pub/sub), `proto/v235/inbound.go` + the multi-event decoders.
**Driver:** `runtime/host.go` `handleFrame`.

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
  right NPC index. The host snapshots the order *before* `Apply` mutates it,
  and compares the server's local-NPC count against the snapshot — a mismatch
  logs a loud desync warning (WARN-only today; a per-tick coherent rebuild is
  TODO `MP-1`). Opcode 191 / 79 also need the player's current position to
  resolve relative offsets to absolute tiles, so the host passes
  `Self.Position()` in.

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
them by diffing world state across an `Apply`, because the wire carries totals
and re-sent snapshots, not deltas/edges:

| Event | Derived how (in `host.go`) | DSL hook |
|---|---|---|
| `ItemGained` | inventory count diff before/after an inventory packet | `on item_gained` |
| `XPGain` | per-skill xp diff before/after an xp-bearing packet (raw packets carry the new total) | `on xp_gain` |
| `LevelUp` | per-skill *max* (base) level diff across stat updates | `on level_up` |
| `TargetDied` | watches `NpcDamage` on the engaged NPC index; fires on the alive→0-hits edge | `on npc_killed` / `on target_died` |
| `EquipmentChanged` | own worn/wielded set diff across inventory updates, one event per changed slot | `on equipment_changed` |
| `PlayerLevelChanged` | another player's appearance packet re-sends periodically; combat level diffed against the last-seen record | `on player_level_changed` |
| `PlayerEquipmentChanged` | same appearance diff, per worn slot that actually changed | `on player_equipment_changed` |

The `Player*Changed` pair is diffed inside the opcode-234 multi-event path
(snapshot before `Apply`, compare, publish edges); the rest are diffed in the
single-event path. These give DSL routines clean "this just happened" hooks
instead of forcing them to re-derive edges from re-sent state.

The bus also carries **host-internal events** that never touch the wire: an
`event.AgentThought` (`"agent_thought"`, `event/agent.go`) is published per
cognition turn so observers can replay the host's reasoning alongside its
perception.

### The bus (`event/bus.go`)

In-memory typed pub/sub. `Subscribe(kind, buffer)` returns a buffered receive
channel; `"*"` subscribes to everything. `Publish` is **non-blocking and
lossy**: if a subscriber's channel is full the event is **dropped for that
subscriber** (no head-of-line blocking). Routines must drain promptly or accept
drops. `Bus.Close()` closes all subscriber channels (a late `Subscribe` on a
closed bus returns a pre-closed channel rather than panicking).

> **Ordering guarantee:** in `handleFrame`, `world.Apply(ev)` runs **before**
> `bus.Publish(ev)`. So any subscriber that reacts to an event sees a world
> state already consistent with it (e.g. an `on npc_killed` handler can read the
> NPC's final state). Multi-event frames apply+publish each event in wire order.

### Who subscribes (observability is BUILT)

The bus is no longer just the DSL's wakeup channel — it is the spine of the
observability substrate (the intent behind the never-built `cmd/delos`/`obs/`
design shipped as the cradle daemon + debug control plane; see
[observability.md](observability.md)):

- **`debughttp/`** — the per-host HTTP control plane. Subscribes `"*"` and
  records every event to an in-memory ring + a JSONL file; serves `GET /ws`
  (live event stream), `GET /events?since=N&kind=K`, `GET /state`, and
  `POST /eval` / `POST /script` for driving the host. Mounted by `cmd/host`
  (`-debug-addr`), `cmd/legacy-cradle` (`-debug-http`), and the cradle daemon,
  which spins one per live host and proxies it under
  `/api/hosts/{name}/debug/` (`cradle/api.go`).
- **Four in-process subscriber goroutines**: `Host.Run` launches the limbic
  loop (`runtime/limbic.go` — affect + trust ledger + the perception/knowledge
  ledgers via `perceptionHandle`), the episodic-memory journal
  (`runtime/memory_journal.go`), and telemetry (`runtime/telemetry.go` —
  counters reported to mesa); the `RunHost` bootstrap adds the decision log
  (`runtime/decisions.go` — `agent_thought` → `decisions.jsonl`).
- **DSL routine subscriptions** (`on <kind>` handlers) and the host's own
  `slog` logging.

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
    Players     *PlayersState      // other players in view (+ dynamic self-index)
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
| `Self` | position + plane (`Coord.Plane()` = Y/944), heading, 18 skills (cur/max/xp), fatigue, quest pts, sleep flag, active prayers, own worn-sprite/colours, death tile + count | `OwnPositionUpdate`(191), `StatUpdate`(159)/`StatsSnapshot`(156)/`ExperienceGain`(33), `FatigueUpdate`(114), `Sleep*`(117/84), `PrayersActive`(206), `Death`(83), own slot of `OtherPlayerAppearance`(234) |
| `Inventory` | per-slot item id / amount / wielded | `InventorySnapshot`(53), `InventorySlotUpdate`(90), `InventoryRemoveSlot`(123) |
| `Npcs` | per-index position, type, heading, **ordered local list** for opcode 79; live combat fields (cur/max hits via 104 type-2, incoming-attack tracking) | `NpcNearby`(79), `NpcDamage`/`NpcChat`/`NpcProjectile`(104), incoming projectiles via `OtherPlayerProjectile`(234) |
| `Players` | per-index position, name, heading, combat level/skull, worn sprites, colours, hits, engagement; **self-index identification by username** | `NearbyPlayerEvent`(191), `OtherPlayerAppearance`/`Damage`/`Projectile`(234) |
| `GroundItems` | visible items keyed by absolute (x,y) | `GroundItemUpdates`(99 — the whole in-view batch per packet), region sweep via `RemoveWorldEntities`(211) |
| `Recent` | most-recent chat, PM, damage, server message, dialog text/options + a messages ring | `ChatReceived`/`SystemMessage`(131), `PrivateMessage`(120), `NpcChat`(104), `NpcDialogText`(222/89), `NpcDialog`(245) |
| `Trade` | offer/confirm/closed phase machine + both offers + accept flags | `TradeOpened`(92), `TradeOtherOffer`(97), `TradeOtherAccepted`(162), `TradeConfirmShown`(20), `TradeClosed`(128) |
| `Duel` | same shape as Trade + rule toggles | `DuelOpened`(176), `DuelOtherOffer`(6), `DuelSettingsUpdate`(30), `DuelOtherAccepted`(253/210), `DuelConfirmShown`(172), `DuelClosed`(225) |
| `Bank` | open flag, max size, slot contents | `BankOpened`(42), `BankSlotUpdate`(249), `BankClosed`(203) |
| `Shop` | catalogue, price modifiers, general-store flag | `ShopOpened`(101), `ShopClosed`(137) |
| `Boundaries` | dynamic doors/walls at absolute tiles | `BoundaryUpdates`(91), region sweep via `RemoveWorldEntities`(211) |
| `Scenery` | dynamic GameObjects at absolute tiles (id 60000 = remove sentinel) | `SceneryUpdates`(48), region sweep via `RemoveWorldEntities`(211) |

### `World.Apply(ev)` — the one mutation point

`Apply` (`world/world.go`) is a single type switch over `event.Event`. Each case
routes to the relevant sub-mirror's setter and returns `true` if the world
meaningfully changed (the host uses that only for a debug log). **This is the
only place that mutates world state** — decoders and the bus never write. A few
cases do more than a straight write:

- **Relative→absolute resolution.** Ground items (99), boundaries (91), scenery
  (48), and region clears (211) arrive as offsets relative to the player *at
  delivery time*, so `Apply` reads `Self.Position()` and resolves to absolute
  tiles before storing. NPC movement (79) is a one-tile **relative** step
  (`MoveBy`) applied to the NPC's *own* stored position — basing it on the
  host's position caused the historical "phantom-crowd" bug.
- **Bulk region eviction.** `RemoveWorldEntities` (opcode 211,
  `SEND_REMOVE_WORLD_ENTITY`) is the server's only eviction channel for
  far-away dynamic entities: each player-relative point names an 8×8 region
  (`abs>>3`), and `Apply` sweeps **all three** dynamic stores (`Boundaries`,
  `Scenery`, `GroundItems`) via their `RemoveRegion` methods.
- **Combat-field preservation.** An NPC's per-tick position update carries no
  health/engagement data, so `NpcsState.Set`/`Move` deliberately **carry
  forward** accumulated combat fields rather than clobbering them.
- **State-machine inference.** Trade/duel `*Closed` packets carry no
  "completed" bit, so `Apply` infers success from the record's accept flags
  (see [the phase machines](#tradeduel-phase-machines) below).
- **Self-index identification (NOT a hardcoded 0).** The server does **not**
  always list our own player at index 0 — our index is whatever slot we
  occupy, and it shifts with who else is in view. The old hardcoded
  `SelfPlayerIndex==0` read mis-attributed *other players'* appearance and
  combat onto `Self` whenever anyone was nearby (the bug that drew us wearing
  a bystander's kit). Today `PlayersState` is told our username at startup
  (`SetSelfName`, `runtime/host.go:498`); the appearance update that **names
  us** pins our real server index (`SetName` → `selfIndex`), and `Apply`
  mirrors appearance bits onto `Self` only for `e.PlayerIndex ==
  Players.SelfIndex()`. The `SelfPlayerIndex = 0` constant survives only as
  the pre-detection default (correct while alone) — see `world/world.go`
  around the const for the full caveat.

### Trade/duel phase machines

`TradeState` / `DuelState` are HOT-path, cognition-free protocol mirrors: every
transition is a real wire event applied synchronously inside `world.Apply`. Both
use the same literal `Phase` strings (`world/trade.go`, `world/duel.go`):
`request_sent` → `open` (offer screen) → `confirm` (final review) →
`completed` | `cancelled`, with `Clear()` resetting to idle for the next
exchange. Load-bearing details, all verified in code:

- **Offer changes un-accept.** Any offer update (`SetMyOffer` /
  `SetTheirOffer`) clears the first-accept flags on **both** sides (server
  rule) — a routine must re-accept after the other side touches their offer.
  For duels, `DuelSettingsUpdate` (`SetRules`) likewise resets both
  first-accepts: a settings change invalidates prior consent.
- **`confirm` is server-pushed, not requested.** Reaching the confirm screen
  (`TradeConfirmShown` op 20 / `DuelConfirmShown` op 172) implies both sides
  first-accepted: for a trade, `Apply` marks our own first-accept as implied
  (`MarkMyFirstAccepted` — the server transitioned, so our echo is moot); for
  a duel, `MarkConfirmShown` forces the phase outright **without** resetting
  accepts. A duel's `DuelOtherAccepted` (op 253/210) is disambiguated by the
  current phase: `open` → first-accept, `confirm` → second-accept.
- **Close-completion inference.** `SEND_TRADE_CLOSE`(128) /
  `SEND_DUEL_CLOSE`(225) carry no completion bit. `Apply` recovers the
  outcome from the accept flags: if we reached `MySecondAccepted` the close
  means *completed* (for a duel, "completed" means **the fight has started** —
  a real control handoff to a combat routine), otherwise `cancelled`.
- **Request events are name-bearing but asymmetric.** `TradeRequestReceived`
  arrives as a typed opcode-131 message with the requester's name structured;
  `DuelRequestReceived` parses the name out of the message text body.

Mermaid state diagrams for both machines (every edge labeled with its
event → setter) live in
[`_research/reference/diagrams/state-machines.md`](_research/reference/diagrams/state-machines.md)
(machines 4–5).

### Thread safety

Every sub-mirror carries its own `sync.RWMutex`; all accessors lock
appropriately, so reads and writes are goroutine-safe. The net loop writes via
`Apply` on the read goroutine's call path while DSL routines (other goroutines)
read concurrently — that is the whole point of the locks. **Read-side accessors
return snapshots/copies** (e.g. `Npcs.All()`, `Players.FindByName`,
`Inventory.Slots()`), never internal references, so a caller can't observe a
half-applied mutation. There is no cross-mirror transaction: each event's
`Apply` is the unit of consistency, and within one event multiple sub-mirror
writes are not atomic relative to a concurrent reader of a *different* mirror
(a per-tick coherent snapshot is TODO `MP-1`).

### Read-side API (what routines query)

Routines never see events unless they explicitly subscribe; mostly they read the
mirror. Examples (Go accessor → conceptual DSL surface):

- `world.Self.Position()`, `.Plane()`, `.SkillLevel(id)`, `.Fatigue()`, `.IsSleeping()`
- `world.Inventory.Slots()`, `.Count(itemID)`
- `world.Players.FindByName(name)`, `.All()`, `.Self()` (our own record at the
  dynamically-identified self index)
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
- **Worn equipment is sprites on the wire — resolved via a reverse index.**
  `OtherPlayerAppearance` / `PlayerRecord.EquipBySlot` carry appearance
  **sprite** ids (low byte of `getAppearanceId()`), not catalogue item ids.
  The old "no sprite→item lookup" gap is **closed**: `facts/worn.go` builds a
  `[wearSlot][appearanceID&0xFF] → []*ItemDef` reverse index at load
  (`buildWornIndex`) and `Facts.ResolveWorn(slot, lowByte)` returns the
  candidate items. Most slots resolve to one item; same-metal melee weapons
  return several candidates — which is faithful (a human can't tell a bronze
  short sword from a bronze long sword on someone's back either).
- **Removal aging is partial.** Dynamic entities (boundaries / scenery /
  ground items) are evicted in bulk by the opcode-211 region clears, and
  NPC/player records are removed on explicit wire removals (opcode-79
  `REMOVE_NPC`, player-removed flag). But there is still **no TTL sweep** for
  a record whose removal packet was missed — it lingers until overwritten
  (TODO `MP-13`).
- **Ground items carry no quantity.** `GroundItemRecord` is keyed by tile with
  one item id and no `Amount` field; stacked drops and multi-item tiles
  collapse (TODO `DSL-20`).
- **No reconnect / no mirror persistence.** The `World` lives entirely in
  memory for one connection; when the connection dies, the host run (and its
  mirror) is gone. Under `cmd/cradle-server` the supervisor relaunches the
  host with a **fresh, empty `World`** (§1) — what survives a relaunch is the
  *durable* per-host state (hostkv, the memory journal), never the mirror.

Open items above and the rest of the wire/mirror backlog live in
[`TODO.md`](TODO.md) §1 (`MP-*`) and §2 (`DSL-*`).
