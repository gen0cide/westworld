# Fresh tutorial-island host — DB snapshot

Captured from `stubbs2` (player id 227) on 2026-06-06: exactly what a
brand-new character looks like across the MySQL `westworld` DB, so new
hosts can be spawned straight into the tutorial.

> **DB connection** is MySQL now (not the old SQLite file). Creds live in
> `~/Code/openrsc/server/connections.conf` (`db_type: mysql`, host
> `localhost:3306`, user `root`); the db name is `westworld` (from
> `inc/westworld.conf`). Never print `db_pass` — read it into a shell var.

## What distinguishes a fresh tutorial host
- **Position `(216, 744)`** — the tutorial-island start room. (A *finished*
  host is at the Lumbridge spawn ~`(120, 648)`.)
- **No `tutorial` key in `player_cache`.** The stage is
  `player.getCache().getInt("tutorial")`, which defaults to **0** when the
  key is absent — so "never started" and "completed" both have no key; the
  *position* is the real discriminator. The key gets created as the host
  progresses and removed by the boatman on completion.
- Stats all level 1 except **hits = 10** (`experience.hits = 4000` is just
  the level-10 base, not a bonus). `combat = 3`, `skill_total = 28`.

## Tables that hold a fresh host's artifacts (player id = N)
Only **8** tables; `logins` is just history (skip for a template).

| table | key col | rows | contents |
|---|---|---|---|
| `players` | `id` | 1 | core row (below) |
| `curstats` | `playerID` | 1 | all skills 1, hits 10 |
| `maxstats` | `playerID` | 1 | all skills 1, hits 10 |
| `experience` | `playerID` | 1 | all 0 except `hits = 4000` |
| `capped_experience` | `playerID` | 1 | all NULL |
| `invitems` | `playerID` | 3 | 3 slots → `itemstatuses` instances |
| `itemstatuses` | `itemID` | 3 | the actual items (below) |
| `player_cache` | `playerID` | 9 | misc flags; **no `tutorial` key** |
| `logins` | `playerID` | 1 | login history — not part of the template |

### `players` core values (secrets omitted)
`group_id=10`, `combat=3`, `skill_total=28`, `x=216`, `y=744`, `fatigue=0`,
`combatstyle=1`, `male=1`, `quest_points=0`, `online=0`, appearance:
`haircolour=2, topcolour=11, trousercolour=11, skincolour=0, headsprite=1,
bodysprite=2`. (Plus `pass`/`salt`/`email`/`creation_ip`/`login_ip` — the
credential/PII columns; set these per host, don't copy verbatim.)

### Starter inventory
`invitems` rows point at per-instance `itemstatuses` (not catalogue ids):

| slot | itemstatuses.itemID (instance) | catalogID | item | amount | durability |
|---|---|---|---|---|---|
| 0 | 8309 | 87 | bronze Axe | 1 | 100 |
| 1 | 8310 | 166 | tinderbox | 1 | 100 |
| 2 | 8311 | 132 | cookedmeat | 1 | 100 |

> Note the two-level model: `invitems(playerID, itemID, slot)` →
> `itemstatuses(itemID, catalogID, amount, noted, wielded, durability)`.
> Worn gear has `itemstatuses.wielded=1`; the appearance packet exposes
> only `catalogID.appearanceID & 0xFF` per slot (see the equipment
> perception work in [[westworld-tutorial-harness]]).

### `player_cache` (9 entries, none tutorial-related)
`kitten_events, last_spell_cast, preferredLanguage, total_played,
gnomeball_npc, gnomeball_goals, setting_block_global, kitten_hunger,
kitten_loneliness`.

## Spawning a new tutorial host
- **Easiest:** register a new account via the client (`want_packet_register:
  true` in westworld.conf) — the server creates exactly the rows above,
  fresh at `(216,744)` with the starter items and no `tutorial` key.
- **By SQL clone:** copy stubbs2's rows into the 7 template tables with a
  new `players.id` + unique `username`/`pass`/`salt`, leaving `player_cache`
  without a `tutorial` key and position at `(216,744)`. Re-point `invitems`
  at fresh `itemstatuses` instance ids.
