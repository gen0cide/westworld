# RSC wire protocol

## What we're targeting

The client mudclient revision is roughly 234/235 — the final RSC era. This is what RSC+ speaks. OpenRSC implements multiple revisions via versioned parser/generator pairs in `server/src/com/openrsc/server/net/rsc/{parsers,generators}/impl/`:

```
Payload38   ← mc38  (very early 2001)
Payload69   ← mc69
Payload115  ← mc115
Payload140  ← mc140 (RSC v3, ~2003)
Payload177  ← mc177
Payload196/198/199/201/202/203  ← mc196-203 (final RSC era)
Payload235  ← mc235  (what RSC+ speaks; what we target)
PayloadCustom ← OpenRSC's own extensions on top of 235
```

**Westworld V1 implements only Payload235.** Other revisions are out of scope until there's a research reason to support them (a "retro" cohort, perhaps).

## What's "the protocol" practically

Two reference files in the OpenRSC source tree:

- `server/src/com/openrsc/server/net/rsc/parsers/impl/Payload235Parser.java`
- `server/src/com/openrsc/server/net/rsc/generators/impl/Payload235Generator.java`

The parser handles inbound (client → server) packets. The generator handles outbound (server → client) packets. We need to **invert** both:

- We need to *encode* what OpenRSC's Payload235Parser *decodes*
- We need to *decode* what OpenRSC's Payload235Generator *encodes*

Every Go protocol function should have a doc comment citing the Java source file + line range it was derived from, plus the commit hash of OpenRSC we ported against. This makes future debugging tractable when something diverges.

## Framing and crypto

RSC packets are length-prefixed and have a 1-byte opcode at the head. Specific framing details:

- Inbound (server → client): 1-byte length OR 2-byte length depending on size; opcode byte; payload
- Outbound (client → server): same framing rules, mirrored
- Login is special: uses an RSA-encrypted credential block

**RSA**: server publishes modulus + exponent (65537 always). Client encrypts the login block — username, password, session info, nonces — with the server pubkey before transmission. Server decrypts with its private key.

We already know the OpenRSC server's RSA modulus from earlier session work:

```
modulus: 7634250561283973106419144827843935010165327069935723928109242614288318739395804201883596278169185387687268116837066108754542364007806573724086207095863517
exponent: 65537
```

These persist across server restarts in `server/{client,server}.pem` files. They're stable for our dev environment. In production we'd read them dynamically from a config file or initial server handshake message.

## Critical packets for Phase 0

The minimum-viable wire surface for "log in, walk, log out":

| Direction | Opcode (approx, verify against Payload235) | Purpose |
|---|---|---|
| Outbound | 32 (varies) | Session-init / handshake request |
| Inbound | session ID response | 8 bytes of session ID |
| Outbound | 0 (login) | RSA-encrypted login block |
| Inbound | 1-byte response code | Login result (64 = success, 3 = invalid creds, etc.) |
| Inbound | mob update | Server tells us our new world position |
| Outbound | walk-to packet | Send target coords; server validates path |
| Inbound | mob update (response) | Server confirms new position |
| Outbound | logout request | Initiate clean logout |
| Inbound | logout ack | Server confirms |

Opcode numbers in the table above are placeholders — they must be verified against `Payload235Parser.java` constants before implementation. Don't trust this table verbatim.

## Login response codes (already known)

From OpenRSC's `LoginResponse.java`:

```
0  = LOGIN_UNSUCCESSFUL
1  = RECONNECT_SUCCESSFUL
3  = INVALID_CREDENTIALS
4  = ACCOUNT_LOGGEDIN
7  = LOGIN_ATTEMPTS_EXCEEDED
64+ = LOGIN_SUCCESSFUL (one per group_id; e.g., 64 = regular player; 89 = admin)
```

Successful login responses are an array indexed by group_id. Our admin account `alex` (group_id 1) gets response code 86 on successful login (from `LOGIN_SUCCESSFUL[1] = 86`).

## Anti-replay protection

OpenRSC's `logins` table has a UNIQUE INDEX on the nonce column. Every successful login inserts a row with the cryptographic nonces extracted from the RSA-encrypted block. Repeated identical nonces (e.g., from a client that doesn't regenerate them) trigger SQLException → login fails with INVALID_CREDENTIALS.

For us: every Go-client login attempt must generate fresh random nonces. RNG: `crypto/rand` — never `math/rand`.

## Walking

The walk-to packet sends a target X,Y. The server validates walkability and either:
- Sends a path-confirmation mob update for our character moving along the path
- Sends a rejection message ("you can't go there") if blocked (e.g., the F2P gate at Falador west)

Our state mirror in `world` updates our coords based on inbound mob updates, not based on optimistic local prediction. This makes the bot's belief about its position always match the server's belief.

## Heartbeat

The OpenRSC server expects clients to send periodic packets — at minimum a no-op ping every ~10s, possibly tied to a specific opcode. Without this, the server may consider the connection dead. We must implement this from Phase 0 or the bot disconnects after ~30s of idle.

The exact heartbeat mechanism needs to be confirmed against the Java reference — possibly there's no explicit heartbeat and any movement/chat packet counts as keep-alive. To verify when we implement Phase 0.

## What we're not implementing in Phase 0

- Inventory tracking (Phase 1)
- Combat (Phase 1)
- Banking (Phase 1)
- Chat send/receive (Phase 1)
- Spell casting (Phase 1)
- NPC dialog (Phase 1)
- Trading (later)

All Phase 1 work is "more opcodes" — the framing/crypto/heartbeat infrastructure is what Phase 0 establishes.

## Open questions

- Does mc234/235 framing differ in subtle ways from earlier revisions (mc204) in the same parser file?
- Are there RSC+-specific extensions we need to mirror? RSC+ adds some QoL features client-side; do any require server-cooperation packets we haven't seen?
- What's the exact heartbeat mechanism?

These will be resolved during Phase 0 implementation against the Java reference + live testing against the OpenRSC server.
