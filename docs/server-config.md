# Server config — westworld.conf

## What this is

`inc/westworld.conf` is the OpenRSC server config that westworld targets. It is
based on `preservation.conf` (OpenRSC's most authentic-content profile) with all
anti-abuse / network limits removed so a 500-host swarm from a single machine can
connect.

**Current posture: P2P / members world** (`member_world: true`). The deployed
config runs as a full members world so the entire RSC content surface (members
skills, NPCs, scenery, quests) is available to the live-test scenario catalog —
F2P-only restrictions were limiting what we could exercise. The conf header
documents this as a **pre-launch dev posture**: flip `member_world: false` to get
F2P geographic confinement when wiring production hosts (see "Members (P2P) vs
F2P posture" below).

> **History note.** An earlier revision of westworld.conf (and earlier revisions
> of this doc) ran F2P (`member_world: false`) for geographic confinement. That
> was reverted to P2P to widen the live-test catalog. If you read references to
> "F2P enforcement / port 43596" elsewhere, they predate this change.

## Where it lives, where it runs

The conf file lives in this repo at `inc/westworld.conf` because it's a westworld
project artifact — we author it, we maintain it, it's the **source-of-truth
backup**. The OpenRSC server lives at `~/Code/openrsc/`. The launcher resolves a
conf by name from `~/Code/openrsc/server/<name>.conf`:

```
cd ~/Code/openrsc/server
ant runserver -DconfFile=westworld          # resolves ~/Code/openrsc/server/westworld.conf
```

> **Drift warning (real, observed 2026-05-31).** The repo file and the deployed
> `~/Code/openrsc/server/westworld.conf` were maintained as **copies**, not a
> symlink, and they drifted (deployed had moved to P2P / port 43594 while the
> repo still said F2P / 43596). The repo file has since been re-synced from the
> deployed one. **Recommended fix: establish the symlink** the workflow always
> intended, so there is exactly one source of truth:
> ```
> ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf
> ```
> After symlinking, edit only the repo file; the deployed copy follows. (This is
> the OpenRSC steward's first infra task — see
> [`agents/openrsc-steward.md`](agents/openrsc-steward.md).)

Alternatively, the conf header documents a direct relative-path invocation
(`ant runserver -DconfFile=../../westworld/inc/westworld`); keep whichever form
you use consistent with that header.

## Members (P2P) vs F2P posture

OpenRSC gates the members map and content on the `member_world` flag. The P2P
gate locs in the server (e.g. the Falador west gate, the Karamja boats — handled
in `DoorAction.java` and friends) check `MEMBER_WORLD`:

- **`member_world: true` (current):** the gate checks pass, so the **full members
  map and content surface are reachable** — there is *no* geographic confinement.
  This is what the live-test catalog needs.
- **`member_world: false` (the production-F2P plan):** the same gate locs **refuse
  passage**, giving free geographic confinement to the F2P map area without any
  custom server code. Flip this when wiring production F2P hosts.

`can_feature_membs` tracks `member_world` (both `true` now) and enables members
custom features (e.g. smithing missile heads).

## Diff against preservation.conf

The full file is in `inc/westworld.conf`. Values below are verified against the
deployed conf and `~/Code/openrsc/server/preservation.conf` as of 2026-05-31.

### Network / anti-abuse limits — removed entirely

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `max_connections_per_ip` | 20 | 10000 | 500+ hosts from a single machine |
| `max_connections_per_second` | 20 | 10000 | mass swarm spinup |
| `max_packets_per_second` | 100 | 1000 | higher per-host packet rate |
| `max_logins_per_second` | 2 | 1000 | mass startup |
| `max_logins_per_server_per_tick` | 5 | 100 | same |
| `max_password_guesses_per_five_minutes` | 10 | 100000 | effectively disabled |
| `max_players` | 2000 | 5000 | room for 500 hosts + headroom |
| `max_players_per_ip` | 3 | 10000 | all hosts from localhost |
| `packet_limit` | 100 | 1000 | same |
| `connection_limit` | 10 | 10000 | same |
| `network_flood_ip_ban_minutes` | 5 | 0 | never IP-ban from this signal |
| `suspicious_player_ip_ban_minutes` | 60 | 0 | never IP-ban from this signal |
| `idle_timer` | 300000 (5min) | 0 | hosts may sit idle deliberately; never auto-kick |
| `enforce_custom_client_version` | true | false | our Go client identifies differently |

### Bot-friendly registration / identity

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `want_packet_register` | true | true | **critical** — allows account creation through the wire protocol, no website registration |
| `want_registration_limit` | true | false | hosts register en masse |
| `registration_limit_count` | 2 | 100000 | effectively disabled |
| `want_global_rules_agreement` | true | false | skip the rules-agreement screen for bot registration |
| `want_email` | true | false | hosts don't need email |

### Members (P2P) content posture

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `member_world` | true | true | **P2P** — full members world (kept from preservation; earlier westworld F2P value reverted) |
| `can_feature_membs` | true | true | members custom features enabled |
| `want_runecraft` | false | true | members skill — enabled for the live-test catalog |
| `want_gianne_badge` | true | true | gnome stronghold (Gianne dishes) |
| `want_blurberry_badge` | true | true | gnome stronghold (Blurberry bar) |

> Note: `member_world` / `can_feature_membs` / the gnome badges match
> preservation's defaults (preservation is itself a members world). The only
> *content* setting westworld flips on vs preservation is `want_runecraft`. The
> meaningful posture story is the **F2P→P2P reversion** described above, not a
> diff from preservation.

### Authentic content — preserved verbatim

| Setting | Value | Why |
|---|---|---|
| `combat_exp_rate` / `skilling_exp_rate` | 1 / 1 | authentic |
| `wilderness_boost` / `skull_boost` | 0 / 0 | authentic |
| `want_fatigue` | true | authentic; tests hosts' sleepword handling |
| `location_data` | 1 | authentic (incl. discontinued) |
| `want_fixed_broken_mechanics` | true | authentic |
| `npc_blocking` / `player_blocking` | 2 / 1 | authentic 2018 RSC |
| `npc_respawn_multiplier` | 1.0 | authentic |
| `aggro_range` | 1 | authentic |
| `show_unidentified_herb_names` | false | authentic — server doesn't leak herb identities |

### Operational overhead — reduced

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `want_pcap_logging` | true | false | reduce server overhead with 500 hosts |
| `monitor_online` | true | false | no external monitoring on dev server |
| `warn_excessive_captcha_failure` | true | false | not running mod alerts |
| `server_sided_word_filtering` | true | false | hosts may say things; filter is admin overhead |
| `npc_kill_messages` | true | false | less chat noise for hosts |
| `want_global_friend` | true | false | don't expose the admin friend list to hosts |
| `want_keyboard_shortcuts` | 1 | 0 | hosts don't use the keyboard |
| `want_discord_*` webhooks (all) | mostly true | false | no Discord integration on dev server (a couple were already false in preservation) |

### Cosmetic / identity

| Setting | preservation | westworld |
|---|---|---|
| `db_name` | preservation | westworld |
| `server_name` / `server_name_welcome` | RSC Preservation | Westworld |
| `welcome_text` | Join our Discord to stay in touch! | This world has been opened to its hosts. |

## Database

The conf names a database: `db_name: westworld`. OpenRSC will look for
`server/inc/sqlite/westworld.db` (sqlite mode) at startup. Initialize it once:

```bash
cd ~/Code/openrsc
make import-authentic-sqlite db=westworld
```

This pipes the schema from `server/database/sqlite/core.sqlite` into a new
`westworld.db` — schema only, no data.

## Database lifecycle

The westworld DB is throwaway research data. For experiments:
- Back up before major cohort experiments.
- Wipe between fundamentally different population designs.
- Don't worry about migrations — the schema is fixed by OpenRSC; we don't modify
  it server-side.

## Ports

| Port | Purpose |
|---|---|
| **43594** | TCP game port — westworld `cradle` hosts connect here (`-server localhost:43594`). |
| **43494** | WebSocket port (`ws_server_port`) — for browser/WS clients. |

> The repo default in `cmd/cradle/main.go` is still `localhost:43596` (the old
> preservation port); pass `-server localhost:43594` explicitly, which the
> scenario runner scripts already do.

## Open questions

- **Multiple worlds**: run multiple westworld instances (different cohort
  experiments on different servers)? Probably yes for larger experiments; each
  needs its own port/conf/db.
- **Production F2P flip**: when wiring production F2P hosts, flip
  `member_world: false` and re-verify the gate-loc confinement holds for the
  current OpenRSC build.
- **Public exposure**: if/when we open the world to observers, which limits do we
  re-enable? Probably none initially — observers are auth-gated by accounts we
  hand out.
- **Performance tuning**: at 500 hosts OpenRSC may strain; we'll learn what
  limits matter as we scale.

## Setup procedure (for the record)

```bash
# Once at project setup — symlink so the repo file is the single source of truth
ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf

# Once: initialize the westworld DB (schema only)
cd ~/Code/openrsc
make import-authentic-sqlite db=westworld

# Each launch
cd ~/Code/openrsc/server
ant runserver -DconfFile=westworld
```

Server listens on TCP **43594** (game) / **43494** (WebSocket). Westworld cradle
hosts connect to 43594.
