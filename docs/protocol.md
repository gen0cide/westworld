# RSC wire protocol (Payload235)

> **STATUS: BUILT — verified against code 2026-06-10, HEAD `0bfa818`.** This is the
> byte-level transport reference: framing, ISAAC, the RSA/XTEA login handshake,
> heartbeat, and disconnect semantics. The Go implementation lives in `proto/v235/`
> (framing, crypto, login, decoders) and `session/` (connection, handshake, I/O
> loops). The per-faculty opcode catalog — what bytes follow each opcode for every
> verb — is the companion [`lang/protocol.md`](lang/protocol.md). Open protocol
> work lives in [`TODO.md`](TODO.md) §1 (notably **MP-3** inbound opcode-validity
> gate + permanent ISAAC regression test, **MP-10** low-priority research tails,
> **MP-12** sleep/wire hardening).

Compiled from direct reading of the OpenRSC reference, with byte-level layouts verified against `RSCProtocolDecoder.java`, `RSCProtocolEncoderMain.java`, and `Payload235Parser.java`.

## Targeting

- **Client revision**: mc234/235 era (final RSC). This is what RSC+ speaks. OpenRSC's `Payload235Parser` / `Payload235Generator` are the reference.
- **Server endpoint**: `localhost:43594` (game port; see [server-config.md](server-config.md)). Both `cradle/hostcfg` (`DefaultServer`, `hostcfg.go:38`) and `cmd/host` (`main.go:69`) default to it. Only the legacy single-host debug binary (`cmd/legacy-cradle`) still defaults to the old preservation port 43596.

OpenRSC supports many revisions via the `PayloadXXX` family; we implement only 235. (The 10010 custom-wire migration question is **MP-5** in [`TODO.md`](TODO.md) — unratified, zero code.)

## Framing

After the RSA-encrypted login completes, all packets in both directions share the same frame structure. **The first byte (post-ISAAC) of the opcode is encrypted via the ISAAC stream cipher; the payload is not.**

Implementation: `proto/v235/frame.go` (`EncodeFrame` / `FrameDecoder`).

### Outbound (client → server) encoding

For a packet with opcode `op` and payload `data` of length `N`:

Let `length = N + 1` (opcode included).

**Case A — `length == 1`** (opcode only, no payload):
```
Wire: [length=1] [ENC(op)]
```

**Case B — `length < 160` and `length != 1`**:
The last byte of `data` is moved between the length byte and the opcode byte. This is OpenRSC's "tail-byte reordering" — `RSCProtocolEncoderMain.java:76` explicitly comments `// Strangely, the last byte of the Payload goes between length and encoded opcode`.
```
Wire: [length] [data[N-1]] [ENC(op)] [data[0..N-2]]
```

**Case C — `length >= 160`**:
Length spans two bytes. Layout is straightforward.
```
Wire: [(length/256) + 160] [length & 0xFF] [ENC(op)] [data[0..N-1]]
```

### Inbound (server → client) decoding

Symmetric to outbound. Read length byte; if `< 160`, that's the length. If `>= 160`, read second byte and compute `length = (firstByte - 160) * 256 + secondByte`. (Equivalent form in code: `256 * firstByte - (40960 - secondByte)`, see `RSCProtocolDecoder.java:154`.)

If `lengthLength == 1` and `length > 1`, the next byte is the tail-byte. Read it, then read `length - 1` more bytes; the first of those is `ENC(op)` and the rest plus the tail-byte (appended at end) form the payload.

### ISAAC desync mitigation

OpenRSC's decoder loops up to 256 ISAAC states trying to find an "isPossiblyValid" opcode (`RSCProtocolDecoder.java:180-212`) as a recoverable-desync safety net.

> **Westworld does NOT implement this loop.** Our `v235.FrameDecoder` (`proto/v235/frame.go`) advances the ISAAC stream exactly once per opcode and trusts sync. **Desync is fatal:** a wrong opcode byte propagates and the connection is torn down (the decode error surfaces via `conn.Err()` and `readLoop` returns — `session/conn.go:221-265`). We aim never to desync rather than recover; if a decode error appears in the logs, treat it as a real protocol bug to fix, not transient noise. Decoders also carry Tier-1 inline anomaly assertions (`proto/v235/anomaly.go`) that WARN with surrounding hex when a decoded field is implausible (item id > 5000, coord > 16384, …) so cascading off-by-N bugs surface immediately. See [lang/protocol-debug-strategy.md](lang/protocol-debug-strategy.md) for diagnosis; the permanent opcode-validity gate + server-reference ISAAC regression test is **MP-3** in [`TODO.md`](TODO.md).

## ISAAC cipher

**Standard Bob Jenkins ISAAC** (1996, original spec). Two instances per connection: `inCipher` (decrypt inbound opcodes) and `outCipher` (encrypt outbound opcodes). Both seeded from the same 4×int (16-byte) key block extracted from the RSA-encrypted login payload.

**Opcode XOR**:
- Encode: `enc_op = (op + outCipher.nextInt()) & 0xFF`
- Decode: `op = (enc_op - inCipher.nextInt()) & 0xFF`

Only the opcode byte is touched. Payload is plain.

**Sources**: `login/ISAACCipher.java`, `net/rsc/ISAACContainer.java:15-21`. Go port: `proto/v235/isaac.go` (256-word state, `SetKeys` + 2-pass init, `EncodeOpcode`/`DecodeOpcode`).

## Login sequence

### Phase 0 — TCP connect + session ID

Client opens TCP connection. Server, after a brief delay (its `SESSION_ID_SENDER_TIMER` default is 640ms; `RSCSessionIdSender.java`), sends a session ID raw — no length/opcode wrapper.

**Built behaviour** (`session/conn.go:114-138`, `consumeOptionalSessionID`): after dialing, the client waits up to `PreLoginReadDelay` (default **800ms**) for **4 bytes** of session ID and discards them. If the deadline expires with no bytes (our login packet beat the server's timer), that is also fine — the session ID is never used for v235 login. Hardening the 800ms-vs-640ms race is part of **MP-12** in [`TODO.md`](TODO.md).

### Phase 1 — Login packet (opcode 0)

Sent **before ISAAC is active**, so the opcode is plain (`0`). Built in `proto/v235/login.go` (`LoginPayload.EncodeLoginFrame`), driven by `session/handshake.go` (`Conn.Login`). The payload body, per `LoginPacketHandler.java:103-200` (the `clientVersion > 204` path):

```
[1 byte]  info        — 0 = new login, 1 = reconnect
[1 byte]  info2       — 0 for modern clients
[1 byte]  info3       — high 8 bits of clientVersion if > 65536, else 0
[2 bytes] clientVersion (big-endian short; we send 235)
[2 bytes] rsaLength,  then [rsaLength bytes]  RSA-encrypted block
[2 bytes] xteaLength, then [xteaLength bytes] XTEA-encrypted block
```

The RSA plaintext block is exactly **61 bytes** (`BuildRSABlock`, per `LoginPacketHandler.java:152-173`):

```
byte 0       checksum = 10 (authentic-client signature)
bytes 1-16   4 uint32s big-endian — the ISAAC/XTEA keys
bytes 17-36  password — UTF-8, spaces→underscores, right-SPACE-padded to 20
byte 37      unused (server never reads it; we write 0xFF)
bytes 38-57  5 uint32s big-endian — nonces[0..4]
bytes 58-60  nonces[5], low 24 bits only
```

The XTEA plaintext block (`BuildXTEABlock`, per `LoginPacketHandler.java:178-192`), encrypted in place with the same 4 keys (`proto/v235/xtea.go`):

```
byte 0       limit30 (we send 30)
bytes 1-24   6 uint32s big-endian — nonces[6..11]
bytes 25..   username (UTF-8), zero-padded to the next 8-byte boundary
             (XTEA operates on 64-bit blocks)
```

**Note for implementation**: there are *two* distinct login-layout cases in `LoginPacketHandler.java` based on `clientVersion`. We target client version 235, which uses the RSA+XTEA blocks above. The simpler "version 93-177" hash-based path is *not* what we implement.

**RSA key**: server's public key — modulus and exponent. On the server side these live in `server/client.pem` / `server/server.pem`, generated on first startup.

**In westworld** the key is **hardcoded** in `proto/v235/rsa.go` `DefaultServerRSA()` (there is no runtime fetch or `.pem` parse):
- Exponent: `65537`
- Modulus: `7634250561283973106419144827843935010165327069935723928109242614288318739395804201883596278169185387687268116837066108754542364007806573724086207095863517`

OpenRSC uses **raw RSA** — no PKCS#1 padding, just `ciphertext = plaintext^e mod n` (`Crypto.java:88-92`); `rsa.go` matches Java's `BigInteger` byte layout (prepends a `0x00` sign byte when the high bit is set).

> **Verification:** these values were captured from the running OpenRSC server and are stable across restarts. If the server is reset and regenerates its key, this constant **must be updated** in `proto/v235/rsa.go` or every login fails. Never run against a server with a mismatched key.

### Phase 2 — Login response

Server sends **1 raw byte** (no framing, ISAAC not yet active):

| Code | Meaning |
|---|---|
| 0 | LOGIN_UNSUCCESSFUL |
| 1 | RECONNECT_SUCCESSFUL |
| 3 | INVALID_CREDENTIALS |
| 4 | ACCOUNT_LOGGEDIN |
| 7 | LOGIN_ATTEMPTS_EXCEEDED |
| 64+ | LOGIN_SUCCESSFUL (varies by group_id; 86 = admin/group 1; 64 = regular/group 10) |

Success codes have `(code & 0x40) != 0`. Failure codes have it clear. (`v235.LoginSuccessful`, `proto/v235/opcodes.go`.)

Source: `util/rsc/LoginResponse.java`.

### Phase 3 — ISAAC seeding

Both client and server initialize their ISAAC ciphers with the 4 keys from the RSA login block. From this point on, all framed packets use ISAAC-encoded opcodes. (`Conn.SetIsaacKeys`, `session/conn.go:153-157` — also flips the frame decoder from `PlainDecode` to ISAAC decode.)

## Opcode coverage

> **Historical note.** This doc originally carried a minimal "opcodes we need for
> Phase 0" list — login/heartbeat/walk/logout outbound; 191/165/234 inbound, with
> "extract own position from 191, TODO everything else". That phase is long over.
> Opcode 191 is fully bit-decoded (`DecodePlayerCoords`, `proto/v235/playercoords.go`),
> as is 234 (`DecodeUpdatePlayers`, `proto/v235/updateplayers.go`).

The opcode SSOT is the code: **outbound** = `proto/v235/opcodes.go` plus the per-subsystem `out…` consts in `action/*.go`; **inbound** = `proto/v235/inbound_opcodes.go`, with the decode switch in `proto/v235/inbound.go`. The full per-namespace catalog (what bytes follow each opcode, per faculty) is [`lang/protocol.md`](lang/protocol.md).

Entries proven since the last revision, worth byte-level notes here:

| Opcode | Dir | Name | Payload |
|---|---|---|---|
| 48 | in | SEND_SCENERY_HANDLER | repeated `[short id] [byte offX] [byte offY]` until exhausted; offsets are signed, player-relative; `id == 60000` is the remove sentinel (`decodeSceneryUpdates`, `inbound.go:676`) |
| 89 | in | SEND_BOX2 | single zero-quoted string — byte-identical payload to 222 (SEND_BOX); both decode to `NpcDialogText` (`inbound.go:61-68`) |
| 211 | in | SEND_REMOVE_WORLD_ENTITY | repeated `[short offX] [short offY]` (SIGNED shorts, player-relative) until exhausted; each point names an 8×8 region (`abs>>3`) whose dynamic boundary/scenery/ground-item entries must be swept — the server's **only** eviction channel for far-away entities (`decodeRemoveWorldEntity`, `inbound.go:568`) |
| 45 | out | SLEEPWORD_ENTERED | `[ubyte sleepDelay]` (we send 0) `[zero-padded string sleepWord]` — answers the SEND_SLEEPSCREEN captcha (`action/sleep.go:35`) |

**Ground items (opcode 99, SEND_GROUND_ITEM_HANDLER)** are a SEQUENCE of entries, read until the payload is exhausted (`decodeGroundItem`, `inbound.go:514`; proven by `proto/v235/ground_item_test.go` — reading only the first entry hid the Barbarian-Village pickaxe):

```
[0xFF] [x] [y]            3 bytes — out-of-range CLEAR: remove the item at (x,y); no id sent
[idHi] [idLo] [x] [y]     4 bytes — ADD id at (x,y); if id carries the 0x8000 bit,
                          it is an in-range REMOVAL of id & 0x7FFF
```

`x`,`y` are signed int8 offsets from the player's current tile.

## Data type encodings

All multi-byte numbers are big-endian. (`proto/v235/buffer.go`.)

| Type | Bytes |
|---|---|
| byte | 1 (unsigned via `& 0xFF`) |
| short | 2 |
| int | 4 |
| long | 8 |
| string | variable, null-terminated (`\0`) or newline-terminated (`\n` = `0x0A`); most server-emitted text is **zero-quoted** — leading `0x00`, UTF-8 content, trailing `0x00` (`proto/v235/strings.go`) |

For bitpacked payloads (inbound 191, 234, and 79 NPC coords — `proto/v235/npccoords.go`), bit-level read access is needed. RSC's bit layout: most-significant-bit first within each byte.

## Heartbeat cadence

Client sends opcode 67 (`OutHeartbeat`, empty payload) periodically. Without heartbeats the server may consider the connection dead and disconnect.

**In westworld** `runtime/host.go` `heartbeatLoop` (`host.go:1038`) sends one every `HeartbeatInterval` (default **5s**, live-validated). **If a heartbeat write fails the loop logs a warning and returns — it is not retried, and the connection is not re-established.**

## Disconnect and restart semantics

The host process itself never auto-reconnects: a dead connection ends the run. Two nuances matter:

- **Server EOF is a restartable failure, not a clean exit.** `readLoop` classifies a remote hang-up (idle-kick, server bounce) as `ErrServerClosed` (`session/conn.go:251-258`); treating it as a nil-error clean exit once made a multi-day soak silently bleed the fleet to zero.
- **Whether a dead host comes back is the supervisor's call.** Under `cmd/cradle-server`, the registry (`cradle/registry.go:211`, `supervise`) applies the per-host `supervision` policy from hosts.yaml (`cradle/hostcfg`): `restart` relaunches with exponential backoff (1s base, doubling to a 30s cap; a relaunch is a fresh login and a fresh genesis), while `hold` — the default — keeps the crashed host visible for inspection instead of restart-looping it. A standalone `cmd/host` process just exits.

## Secrets and environment

- **Account password** comes from the `WESTWORLD_PASSWORD` env var (or the `-password` flag; the flag wins). Never embed it in code, docs, or `ps`-visible args. `cmd/host/main.go:91-95` reads it and exits if unset. Under the cradle daemon, each host in hosts.yaml names its env var via `password_env` (default `WESTWORLD_PASSWORD`; `cradle/hostcfg/hostcfg.go`), resolved once at launch by `cmd/cradle-server`.
- **ISAAC/XTEA keys** are four random `uint32`s generated fresh per login (`proto/v235/login.go`), committed inside the RSA block, and used to seed both ISAAC directions on a successful response.
- **RSA modulus/exponent** are the hardcoded `DefaultServerRSA()` constants (see [RSA key](#phase-1--login-packet-opcode-0) above).

The full connection + login walkthrough, the typed-event model, and the world-state mirror are documented in [world-state.md](world-state.md).

## References

- `net/RSCProtocolEncoderMain.java:55-100` — outbound framing, tail-byte handling
- `net/RSCProtocolDecoder.java:147-238` — inbound framing, ISAAC desync mitigation
- `net/rsc/parsers/impl/Payload235Parser.java:27-42` — opcode → enum mapping
- `net/rsc/generators/impl/Payload235Generator.java:25-93` — outbound opcode map
- `net/rsc/LoginPacketHandler.java:103-205` — login flow
- `login/ISAACCipher.java` — ISAAC implementation
- `login/LoginRequest.java:42-160` — login request lifecycle
- `util/rsc/LoginResponse.java` — response code definitions
- `net/RSCSessionIdSender.java` — session ID format
