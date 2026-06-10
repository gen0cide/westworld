# Server config — westworld.conf

> **STATUS: verified against code 2026-06-10, branch `tidy/structure-and-docs`,
> HEAD `0bfa818`.** Conf values below were re-checked against
> `inc/westworld.conf`, the deployed `~/Code/openrsc/server/westworld.conf`,
> and `~/Code/openrsc/server/preservation.conf` on that date. Open work is
> tracked in [`TODO.md`](TODO.md) (the SSOT — IDs like `O-8`/`O-12` below refer
> to its items); this doc carries no backlog of its own.

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

> **Drift warning (real, now observed TWICE).** The repo file and the deployed
> `~/Code/openrsc/server/westworld.conf` are maintained as **copies**, not a
> symlink, and they keep drifting:
>
> 1. **2026-05-31** — deployed had moved to P2P / port 43594 while the repo
>    still said F2P / 43596. Repo re-synced from deployed.
> 2. **2026-06-09** — deployed flipped `want_runecraft: false` (comment:
>    "MATCH uranium/preservation — runecraft is a custom (non-authentic)
>    skill") while the repo still says `true`. As of 2026-06-10 the deployed
>    file is still a plain file (not a symlink) and `diff` shows exactly that
>    one line differing. Which side is correct is an open reconciliation call.
>
> The fix — establish the symlink the workflow always intended, then reconcile
> `want_runecraft` and update this doc's diff table — is tracked as **O-12** in
> [`TODO.md`](TODO.md) (§6), and is the OpenRSC steward's standing first task
> (see [`agents/openrsc-steward.md`](agents/openrsc-steward.md)):
> ```
> ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf
> ```
> After symlinking, edit only the repo file; the deployed copy follows.

Alternatively, the conf header documents a direct relative-path invocation
(`ant runserver -DconfFile=../../westworld/inc/westworld`); keep whichever form
you use consistent with that header.

## Members (P2P) vs F2P posture

OpenRSC gates the members map and content on the `member_world` flag. The P2P
gate locs in the server (e.g. the Falador west gate, the Karamja boats — handled
in `DoorAction.java` and friends; 23 plugin files reference `MEMBER_WORLD`, 21 of them authentic-content)
behave as follows:

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
repo conf and `~/Code/openrsc/server/preservation.conf` as of 2026-06-10. The
deployed conf matches the repo conf except for the one `want_runecraft` drift
line noted above (O-12).

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
| `want_runecraft` | false | true (repo) / **false (deployed, drifted 2026-06-09)** | members skill — repo enables it for the live-test catalog; deployed reverted to match uranium/preservation (runecraft is a custom, non-authentic skill). Reconciliation = O-12. |
| `want_gianne_badge` | true | true | gnome stronghold (Gianne dishes) |
| `want_blurberry_badge` | true | true | gnome stronghold (Blurberry bar) |

> Note: `member_world` / `can_feature_membs` / the gnome badges match
> preservation's defaults (preservation is itself a members world). The only
> *content* setting the repo conf flips on vs preservation is `want_runecraft`
> — and that is exactly the line currently in drift. The meaningful posture
> story is the **F2P→P2P reversion** described above, not a diff from
> preservation.

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
| `want_discord_*` webhooks (all 8) | mostly true | false | no Discord integration on dev server (a couple were already false in preservation) |

### Cosmetic / identity

| Setting | preservation | westworld |
|---|---|---|
| `db_name` | preservation | westworld |
| `server_name` / `server_name_welcome` | RSC Preservation | Westworld |
| `welcome_text` | Join our Discord to stay in touch! | This world has been opened to its hosts. |

## Database

The conf names a database: `db_name: westworld`. **The DB is MySQL/MariaDB, not
the old SQLite file.** The backend is selected in
`~/Code/openrsc/server/connections.conf` (`db_type: mysql`,
`db_host: localhost:3306`); the credentials live there too — **never print
`db_pass`**, read it into a shell var if you need it. Initialize the DB once:

```bash
cd ~/Code/openrsc
make create-mariadb db=westworld            # CREATE DATABASE westworld
make import-authentic-mariadb db=westworld  # pipe server/database/mysql/core.sql — schema only, no data
```

(An earlier revision of this doc described the `import-authentic-sqlite` /
`server/inc/sqlite/westworld.db` path; that target still exists in the OpenRSC
Makefile but is not what the deployed server uses.)

For what a fresh host's rows actually look like across the MySQL tables — and
the spawn-recipe for minting new tutorial hosts straight from SQL — see
[`tutorial-host-snapshot.md`](tutorial-host-snapshot.md).

## Database lifecycle

The westworld DB is throwaway research data. For experiments:
- Back up before major cohort experiments (the OpenRSC `Makefile` has
  `mysqldump`-based backup targets).
- Wipe between fundamentally different population designs.
- Don't worry about migrations — the schema is fixed by OpenRSC; we don't modify
  it server-side.

## Ports

| Port | Purpose |
|---|---|
| **43594** | TCP game port (`server_port`) — westworld hosts connect here. |
| **43494** | WebSocket port (`ws_server_port`) — for browser/WS clients. |

> Default-address status across the binaries: the cradle daemon
> (`cmd/cradle-server`, via `cradle/hostcfg/hostcfg.go:38`
> `DefaultServer = "localhost:43594"`) and `cmd/host`
> (`cmd/host/main.go:69`) already default to **43594**. Only the legacy
> single-host debug binary still defaults to the old preservation port:
> `cmd/legacy-cradle/main.go:81` says `localhost:43596` — pass
> `-server localhost:43594` explicitly there (`scripts/dev-launch.sh` and the
> `cmd/scenariogen/run_*.sh` runners already do).

## Open work

Server/ops decisions (multiple worlds per experiment, the production F2P flip +
gate-loc re-verify, public-exposure limit re-enable, perf at 500 hosts) live in
[`TODO.md`](TODO.md) §6 as **O-8**; the conf-drift symlink + `want_runecraft`
reconciliation is **O-12**; westworld.conf ↔ uranium.conf content alignment (the
pcap/golden-corpus prerequisite) is **MP-6** (§1).

## Setup procedure (for the record)

```bash
# Once at project setup — symlink so the repo file is the single source of truth
# (chartered but NOT yet executed — see O-12)
ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf

# Once: create + initialize the westworld DB (MySQL, schema only)
cd ~/Code/openrsc
make create-mariadb db=westworld
make import-authentic-mariadb db=westworld

# Each launch
cd ~/Code/openrsc/server
ant runserver -DconfFile=westworld
```

Server listens on TCP **43594** (game) / **43494** (WebSocket). Westworld
hosts connect to 43594.
