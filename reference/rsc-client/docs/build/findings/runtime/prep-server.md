> _Server bring-up report (2026-06-01)._

# OpenRSC Server Bringup Report

## BUILD STATUS: SUCCESS
- Compiled with **JDK 17** (`/usr/lib/jvm/java-17-openjdk`). Server source targets `javac source/target 1.8`; compiles clean on JDK 17 (deprecation warnings only). Do **not** use JDK 26 (system default) or 20+ — `source 8` was dropped after 17.
- The bundled `inc/ant` is incomplete (no `lib/`), so I used the vendored **Apache Ant 1.10.5** at `/home/free/code/rsc-hacking/openrsc/Portable_Windows/apache-ant-1.10.5`.
- All Java deps are vendored in `server/lib/` — **no network needed**.
- Artifacts: `/home/free/code/rsc-hacking/openrsc/server/core.jar` (38 MB) + `plugins.jar` (2.5 MB). Build log: `/tmp/rsc-run/server-build.log`.

## RUN STATUS: LIVE on 127.0.0.1:43594 (TCP game) + :43494 (WS)
- Started in background (PID 318415, wrapper `/tmp/rsc-run/start-server.sh`, log `/tmp/rsc-run/server-run.log`). Log says "Game world is now online on TCP port 43594!"; `ss` confirms LISTEN; raw TCP connect to 43594 succeeds.
- Config used: **`westworld.conf`** (the canonical westworld profile per `westworld/docs/server-config.md`), copied to `/home/free/code/rsc-hacking/openrsc/server/westworld.conf`. Port **43594**, db `westworld`, `enforce_custom_client_version: false`, `custom_sprites: false`.
- **Port caveat:** the cradle's default `-server localhost:43596` is STALE. The live game port is **43594**. Pass `-server localhost:43594` (the scenario scripts already do this per the doc).

## DATABASE: SQLite, no MySQL needed
- `DatabaseType.DEFAULT = SQLITE`; `db_type` unset → SQLite at `server/inc/sqlite/<db_name>.db`. MySQL/MariaDB is not installed and not required.
- I initialized `server/inc/sqlite/westworld.db` schema-only (0 players) from `server/database/sqlite/core.sqlite` via `sqlite3`.

## RSA PUBLIC KEY — the value the deob must set `BitBuffer.RSA_MODULUS` to
Keys are **FRESH**, generated on first boot (`Crypto.generateRSAKeys` writes `client.pem`/`server.pem` in the server CWD only if absent; 512-bit; persist across restarts). They do **NOT** match the stale `DefaultServerRSA` in `/home/free/code/rsc-hacking/westworld/proto/v235/rsa.go` — that file must be updated (or read from `client.pem`).
- exponent (decimal): `65537`
- modulus (decimal): `8470727801174954902989859055344934434282083179399207801708507751976321325965228952554034824402302678046886295251980280826867546707365065713308009848924031`
- modulus (hex): `a1bc0e158a70c86b8bccd31aeac822bae08c2fd25681cfd8ecb6fa85562b08945b8fe1157433a6126a39797e39d5af7dd19a912514841962878850dc2c00677f` (512-bit)
- Verified two independent ways: server startup log "RSA modulus:" line AND ASN.1-decode of `/home/free/code/rsc-hacking/openrsc/server/client.pem` — they match. Private key: `/home/free/code/rsc-hacking/openrsc/server/server.pem`.
- WARNING: deleting the `.pem` files makes the next boot generate a new key (modulus changes).

## ACCOUNT STORY: register over the wire (no seed account)
- `westworld.db` has **zero players** — there is no pre-seeded `alex`. `want_packet_register: true` (and `want_email`/`want_global_rules_agreement: false`) means accounts are created **in-protocol** (register opcode → `CharacterCreateRequest`, `LoginPacketHandler` ~line 738), no website/email. Logins from `127.0.0.1` bypass rate/IP limits. Password = cradle `-password` flag or `WESTWORLD_PASSWORD`. So the deob's first login path is **register-then-login** (or insert a `players` row directly).

## LOGIN-FORMAT NOTE: >204 layout, matches the cradle
- Server takes the `clientVersion > 204` branch (`LoginPacketHandler.java:145-196`): password in the RSA block, username in the XTEA block, ISAAC keys + nonces in RSA — exactly the layout in `/home/free/code/rsc-hacking/westworld/proto/v235/login.go`. Raw RSA (`BigInteger.modPow`, no PKCS#1 padding), matching the cradle's `m^e mod n`.
- Version gate (`LoginRequest.java:211`): for `clientVersion <= 10000` (e.g. 235), it rejects only if `WANT_CUSTOM_SPRITES` — which is **false** here. So a version-235 deob client is accepted.

Full details written to `/tmp/rsc-run/server-info.txt`.

## SERVER STATUS: READY
