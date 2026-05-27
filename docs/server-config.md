# Server config — westworld.conf

## What this is

`inc/westworld.conf` is the OpenRSC server config that westworld targets. It's based on `preservation.conf` (the most authentic content settings) with all anti-abuse / network limits removed and `member_world: false` for F2P geographic confinement.

## Where it lives, where it runs

The conf file lives in this repo at `inc/westworld.conf` because it's a westworld project artifact — we author it, we maintain it, it's tied to westworld's needs.

The OpenRSC server lives at `~/Code/openrsc/`. To run it with our config, the conf file is referenced by name to the OpenRSC launcher:

```
cd ~/Code/openrsc/server
ant runserver -DconfFile=westworld
```

The launcher expects the conf file in `~/Code/openrsc/server/<name>.conf`. So our workflow is one of:

1. **Symlink**: `ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf`. Edits to our file propagate.
2. **Copy on change**: copy the file before each server launch. Brittle but explicit.
3. **Build script**: a small make target or shell script that ensures the symlink exists and launches OpenRSC pointed at it.

Recommendation: symlink at setup time, never touch again. Document it in the project README.

## Diff against preservation.conf

The full file is in `inc/westworld.conf`. The category-level diff:

### Limits — removed entirely

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `max_connections_per_ip` | 3 | 10000 | 500+ bots from a single machine |
| `max_connections_per_second` | 20 | 10000 | mass swarm spinup |
| `max_packets_per_second` | 100 | 1000 | active swarm runs at higher per-host packet rate |
| `max_logins_per_second` | 2 | 1000 | mass startup |
| `max_logins_per_server_per_tick` | 5 | 100 | same |
| `max_password_guesses_per_five_minutes` | 10 | 100000 | effectively disabled |
| `max_players` | 2000 | 5000 | room for 500 bots + headroom |
| `max_players_per_ip` | 3 | 10000 | all bots from localhost |
| `network_flood_ip_ban_minutes` | 5 | 0 | no flood bans |
| `suspicious_player_ip_ban_minutes` | 60 | 0 | no suspicion bans |
| `want_registration_limit` | true | false | bots register en masse |
| `registration_limit_count` | 2 | 100000 | effectively disabled |
| `packet_limit` | 100 | 1000 | same |
| `connection_limit` | 10 | 10000 | same |
| `is_localhost_restricted` | false | false | (preservation tried to enforce; we want loopback) |
| `enforce_custom_client_version` | true | false | our Go client identifies differently |

### F2P enforcement

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `member_world` | true | false | F2P; gate locs in OpenRSC source enforce geographic confinement |
| `can_feature_membs` | true | false | also F2P |

The P2P gate locs in OpenRSC's `DoorAction.java` (the Falador west gate, Karamja boats, etc.) check `MEMBER_WORLD` and refuse passage when false. This gives us free geographic confinement to the F2P map area.

### Authentic content — preserved verbatim

| Setting | Value | Why |
|---|---|---|
| `combat_exp_rate` | 1 | authentic |
| `skilling_exp_rate` | 1 | authentic |
| `wilderness_boost` | 0 | authentic |
| `skull_boost` | 0 | authentic |
| `want_fatigue` | true | authentic; tests bots' sleepword handling |
| `location_data` | 1 | authentic (incl discontinued) |
| `want_fixed_broken_mechanics` | true | authentic |
| `npc_blocking` | 2 | authentic 2018 RSC |
| `player_blocking` | 1 | authentic 2018 RSC |
| `npc_respawn_multiplier` | 1.0 | authentic |
| `pidless_catching` | true | (preservation default) |

### Operational overhead — reduced

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `want_pcap_logging` | true | false | reduce server overhead with 500 hosts |
| `monitor_online` | true | false | no external monitoring on dev server |
| `idle_timer` | 300000 (5min) | 0 | bots may sit idle deliberately; never auto-kick |
| `warn_excessive_captcha_failure` | true | false | not running mod alerts |
| `server_sided_word_filtering` | true | false | bots may say things; filter is admin overhead |
| All Discord webhooks | true | false | no external integrations on dev server |

### Bot-friendly server packet registration

| Setting | preservation | westworld | Why |
|---|---|---|---|
| `want_packet_register` | true | true | (already on in preservation) **critical**: allows bot account creation through the wire protocol, no website registration |

## Database

The conf names a database: `db_name: westworld`. This means OpenRSC will look for `server/inc/sqlite/westworld.db` (sqlite mode) when starting. We need to initialize it:

```bash
cd ~/Code/openrsc
make import-authentic-sqlite db=westworld
```

This pipes the schema from `server/database/sqlite/core.sqlite` into a new `westworld.db`. No data, just the schema.

## Database lifecycle

The westworld DB is throwaway-research-data. For experiments:
- Backup before major cohort experiments
- Wipe between fundamentally different population designs
- Don't worry about migrations — schema is fixed by OpenRSC; we don't modify it server-side

## Open questions

- **Multiple worlds**: should we eventually run multiple westworld instances (different cohort experiments on different servers)? Probably yes for larger experiments. Each would need its own port/conf.
- **Public exposure**: if/when we open the world to other observers, what additional limits do we re-enable? Probably none initially — observers are auth-gated by accounts we hand out.
- **Performance tuning**: at 500 hosts, OpenRSC may strain. We'll learn what limits matter as we scale.

## Setup procedure (for the record)

```bash
# Once at project setup
ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf

# Once: initialize the westworld DB
cd ~/Code/openrsc
make import-authentic-sqlite db=westworld

# Each launch
cd ~/Code/openrsc/server
ant runserver -DconfFile=westworld
```

Server listens on TCP 43596. Westworld cradle hosts connect there.
