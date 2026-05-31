# RSC wire protocol (Payload235)

Implementation-ready reference. Compiled from direct reading of the OpenRSC reference, with byte-level layouts verified against `RSCProtocolDecoder.java`, `RSCProtocolEncoderMain.java`, and `Payload235Parser.java`.

## Targeting

- **Client revision**: mc234/235 era (final RSC). This is what RSC+ speaks. OpenRSC's `Payload235Parser` / `Payload235Generator` are the reference.
- **Server endpoint**: `localhost:43596` (game port; see [server-config.md](server-config.md))

OpenRSC supports many revisions via the `PayloadXXX` family; we implement only 235. Other revisions are out of scope for v1.

## Framing

After the RSA-encrypted login completes, all packets in both directions share the same frame structure. **The first byte (post-ISAAC) of the opcode is encrypted via the ISAAC stream cipher; the payload is not.**

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

> **Westworld does NOT implement this loop.** Our `v235.FrameDecoder` (`proto/v235/frame.go`) advances the ISAAC stream exactly once per opcode and trusts sync. **Desync is fatal:** a wrong opcode byte propagates and the connection is torn down (the decode error surfaces via `conn.Err()` and `readLoop` returns). We aim never to desync rather than recover; if a decode error appears in the logs, treat it as a real protocol bug to fix, not transient noise. See [lang/protocol-debug-strategy.md](lang/protocol-debug-strategy.md) for diagnosis.

## ISAAC cipher

**Standard Bob Jenkins ISAAC** (1996, original spec). Two instances per connection: `inCipher` (decrypt inbound opcodes) and `outCipher` (encrypt outbound opcodes). Both seeded from the same 4×int (16-byte) key block extracted from the RSA-encrypted login payload.

**Opcode XOR**:
- Encode: `enc_op = (op + outCipher.nextInt()) & 0xFF`
- Decode: `op = (enc_op - inCipher.nextInt()) & 0xFF`

Only the opcode byte is touched. Payload is plain.

**Sources**: `login/ISAACCipher.java`, `net/rsc/ISAACContainer.java:15-21`.

## Login sequence

### Phase 0 — TCP connect + session ID

Client opens TCP connection. Server (after a brief ~100ms delay, per `RSCSessionIdSender.java`) sends 4 or 8 bytes of session ID, raw (no length/opcode wrapper).

**v1 assumption**: 8-byte (long) session ID. Verify when testing.

### Phase 1 — Login packet (opcode 0)

Sent **before ISAAC is active**, so framing is unusual for this one packet. Length is 2 bytes (big-endian), no tail-byte gymnastics, opcode is plain (`0`).

```
Wire: [length_high] [length_low] [opcode=0] [version_block] [rsa_block_length] [RSA-encrypted_block] [xtea_block_length] [XTEA-encrypted_block]
```

The RSA block contains (after server decrypts with private key):

```
[1 byte]   checksum = 10
[16 bytes] 4 ints (big-endian) — XTEA / ISAAC keys
[20 bytes] password, null-padded, spaces→underscores
[24 bytes] 6 nonces
```

The XTEA block (encrypted with the 4 keys from the RSA block) contains:

```
[1 byte]   limit30
[24 bytes] 6 more nonces
[?]        username, null-terminated
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

Success codes have `(code & 0x40) != 0`. Failure codes have it clear.

Source: `util/rsc/LoginResponse.java`.

### Phase 3 — ISAAC seeding

Both client and server initialize their ISAAC ciphers with the 4 keys from the RSA login block. From this point on, all framed packets use ISAAC-encoded opcodes.

## Opcodes we need for Phase 0

Verified directly against `Payload235Parser.java`:

**Outbound (client → server):**

| Opcode | Name | Payload |
|---|---|---|
| 0 | LOGIN | RSA + XTEA blocks (see above) |
| 67 | HEARTBEAT | (empty) |
| 187 | WALK_TO_POINT | `[short:x] [short:y]` plus optional waypoint pairs |
| 102 | LOGOUT | (empty) |
| 31 | CONFIRM_LOGOUT | (empty; response to server-initiated logout) |

**Inbound (server → client):**

| Opcode | Name | Payload |
|---|---|---|
| 165 | SEND_LOGOUT | (empty; logout confirmation) |
| 191 | SEND_PLAYER_COORDS | bitpacked mob update with own position |
| 234 | SEND_UPDATE_PLAYERS | bitpacked appearance updates for visible players |

For Phase 0 we **must** decode opcode 191 (position update) and 165 (logout ack). Other inbound opcodes can be unknown for now — they're logged and discarded.

## Data type encodings

All multi-byte numbers are big-endian.

| Type | Bytes |
|---|---|
| byte | 1 (unsigned via `& 0xFF`) |
| short | 2 |
| int | 4 |
| long | 8 |
| string | variable, null-terminated (`\0`) or newline-terminated (`\n` = `0x0A`) |

For bitpacked payloads (inbound 191, 234), bit-level read access is needed. RSC's bit layout: most-significant-bit first within each byte.

## Heartbeat cadence

Client sends opcode 67 (`OutHeartbeat`, empty payload) periodically. Without heartbeats the server may consider the connection dead and disconnect.

**In westworld** `runtime/host.go` `heartbeatLoop` sends one every `HeartbeatInterval` (default **5s**). **If a heartbeat write fails the loop logs a warning and returns — it is not retried, and the connection is not re-established.** A dropped connection terminates the host (there is no auto-reconnect).

## Secrets and environment

- **Account password** comes from the `WESTWORLD_PASSWORD` env var (or the `-password` flag; the flag wins). Never embed it in code, docs, or `ps`-visible args. `cmd/cradle/main.go` reads it and exits if unset.
- **ISAAC/XTEA keys** are four random `uint32`s generated fresh per login (`proto/v235/login.go`), committed inside the RSA block, and used to seed both ISAAC directions on a successful response.
- **RSA modulus/exponent** are the hardcoded `DefaultServerRSA()` constants (see [RSA key](#phase-1--login-packet-opcode-0) above).

The full connection + login walkthrough, the typed-event model, and the world-state mirror are documented in [world-state.md](world-state.md).

## What Phase 0 ignores

- Bit-level decoding of opcodes 191/234 — we just need to extract our OWN position from 191; everything else can be a TODO
- All other inbound opcodes (chat, NPCs appearing, items, etc.) — log and discard
- Outbound packet rate limiting (we don't spam in Phase 0)
- ISAAC desync recovery (we trust sync; if we drift, we fail loudly and fix the underlying issue)

These are filled in during Phase 1.

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
