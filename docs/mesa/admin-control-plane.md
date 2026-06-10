# mesa Admin control plane + `mesa-ctl`

> **STATUS: BUILT** (verified 2026-06-10 against branch HEAD `93d3cd1`). The v1
> design below shipped near-verbatim (PR #5, merge `1c2423e`) and has since grown
> three pieces the original plan deferred: **live `goal push`** (the shipped half
> of v2/v3), **`fleet gen`** (the hostcfg generator), and **`persona set`**
> (single-field dial edits, not in the original sketch). Code: the `Admin`
> service in `mesa/proto/mesa.proto:473` + handlers `mesa/mesad/admin.go` +
> `mesa/mesad/auth.go`, client `cmd/mesa-ctl`. Not built: live
> `PERSONA_REVISION` push (hosts log it — `runtime/runhost_bootstrap.go:574`),
> persona history/recook/hosts-roster — see **Open work** at the bottom.

## Problem it solves

mesad stores personas durably (Postgres `personas` table —
`mesa/mesad/ltm.go:200`), restores them on boot (`LoadPersonas` —
`mesa/mesad/server.go:215`), and `Register` (`server.go:171`) does the whole job
*live*: validate → render prose + system prompt → put in the in-memory registry
→ upsert to Postgres → authorize the host. Before the Admin service, `Register`
was only reachable from `main()` via the `-host host_id=persona.json` flag, so
any persona add/change/list/remove meant a mesad restart, and loading 200
personas meant 200 `-host` flags. The Admin service exposes that same path over
gRPC: a `mesa-ctl` put takes effect on the next `Act`/`Decide`/`Provision.Fetch`
immediately — no restart, no reconnect. (`-host` survives as a one-shot seed —
`cmd/mesad/main.go:73`.)

## The Admin service (`mesa/proto/mesa.proto:473`, handlers `mesa/mesad/admin.go`)

```proto
service Admin {
  rpc PutPersonas(stream PersonaUpsert) returns (BatchResult);
  rpc GetPersona(HostRef) returns (PersonaRecord);
  rpc ListPersonas(ListPersonasRequest) returns (PersonaList);
  rpc DeletePersona(HostRef) returns (AdminAck);
  rpc PushGoal(PushGoalRequest) returns (PushGoalResult);
}
```

- **Client-streaming `PutPersonas`** is the bulk primitive: 200 personas go up
  in one call with backpressure, each validated server-side, per-item results
  back (`ItemResult`/`BatchResult`). One bad persona does NOT fail the batch — a
  199-good / 1-bad import still lands the 199 (`admin.go:28`). A single put is a
  stream of one.
- Handlers are thin: `putOnePersona` (`admin.go:53`) unmarshals and calls
  `s.Register` — the existing validate/render/persist/authorize path.
  `GetPersona` returns JSON + the prose card the brain reads; `ListPersonas` is
  metadata-only (host_id, display name, updated_at) unless `with_json` is set,
  sorted by host_id. `DeletePersona` removes from the live registry, **revokes
  the host's derived auth token**, and deletes the Postgres row
  (`ltm.go:1078`); idempotent. `updated_at` is best-effort display metadata
  (`PersonaTimes` — `ltm.go:1086`) and never blocks a response.
- `PushGoal` is the live-push RPC — see **Goal push** below.

## Auth — operator tier, separate from per-host bearer (`mesa/mesad/auth.go`)

Host auth is a bearer-token → host_id map (`tokens` — `server.go:52`); a host
must never rewrite another host's identity. The Admin service authenticates with
a separate **operator credential**: every method under
`/westworld.mesa.v2.Admin/` is gated by `authenticateAdmin` inside the existing
`UnaryAuth`/`StreamAuth` interceptors (`auth.go:99`, `auth.go:123`).

- The credential is `$ADMIN_TOKEN` on mesad (`cmd/mesad/main.go:135` →
  `SetAdminToken`); `mesa-ctl` sends it as bearer metadata (`$ADMIN_TOKEN` or
  `-token` client-side).
- **Fail-closed:** an unset token disables the Admin API entirely — every Admin
  call returns `Unauthenticated` (`auth.go:25`). Constant-time compare.
- All other services keep host-token auth unchanged.

## `cmd/mesa-ctl`

```
mesa-ctl [-addr host:port] [-token <admin-token>] <command> [args]

  persona put <host_id> <file|->     register/replace one persona (file or stdin)
  persona import <dir|->             bulk: a directory of <host_id>.json, or NDJSON
                                     ({"host_id":..,"persona":{..}} per line) on stdin
  persona ls [--json]                list registered personas
  persona get <host_id> [--json]     facets + dials + prose; --json for raw JSON
  persona set <host_id> <field> <v>  edit one dial/field (validated server-side)
  persona rm <host_id>               remove a persona
  fleet gen [flags]                  emit a cradle hostcfg for the registered personas
  goal push <goal> [--match <glob>]  soft runtime goal override on running hosts
```

- **Bulk `import`** (the 200-persona path) accepts a **directory** of
  `<host_id>.json` (filename = host_id) or **NDJSON on stdin**
  (`cmd/mesa-ctl/main.go:144`). It opens one `PutPersonas` stream, prints the
  per-item report (`✓ Delores` / `✗ Teddy: validation: …`), and exits non-zero
  if any item failed. (The originally-sketched third format — a
  `{host_id: path}` manifest file — was never built; `collectImport` rejects
  plain files.)
- **`persona set`** (`cmd/mesa-ctl/set.go`) patches one field by its
  operator-facing name (the same names as the generated DB columns —
  `aggression`, `hexaco_h`, `cur_spatial`, `north_star`, ~33 fields): it fetches
  the persona, edits the JSON path, and re-uploads through `PutPersonas`, so the
  server re-validates (rejecting a bad band), re-renders the prose, updates the
  live registry, and persists.

## Storage (`mesa/mesad/ltm.go`)

- `personas (host_id PK, persona_json jsonb, updated_at)` (`ltm.go:200`), plus
  app-written `prose_card`/`cooked`/`created_at` and ~33 **`GENERATED ALWAYS
  ... STORED` projection columns** derived from `persona_json`
  (`ltm.go:236-331`) — every identity facet and dial is a first-class queryable
  SQL column, always in sync with the canonical JSON, zero app extraction.
- `UpsertPersona` (`ltm.go:1046`) is idempotent; bulk = N best-effort per-row
  upserts (deliberately not one all-or-nothing tx).
- No `persona_history` table — history/rollback is open (TODO `M-4`).

## Goal push — the live host-push slice that shipped

`Admin.PushGoal` (`server.go:670`) sets a **soft runtime goal override** on
every registered host matching an optional `filepath.Match` glob (empty = all),
then delivers a `GOAL_REVISION` `Directive` to the hosts currently holding an
open `Provision.Subscribe` stream. `mesa-ctl goal push <goal> [--match 'drone*']`
(`cmd/mesa-ctl/goal.go`) reports `matched` (goal set, including offline hosts)
vs `pushed` (delivered live).

Mechanics, all in `mesa/mesad/server.go`:

- **Subscriber registry** (`subs`, `server.go:77-79`): `Subscribe`
  (`server.go:609`) registers a buffered per-host directive channel and selects
  on it for the connection lifetime; a reconnect supersedes the stale stream. A
  slow subscriber whose buffer is full is logged + skipped, never blocking the
  fan-out.
- **Sticky across reconnects:** the pushed goal is written into the registry
  entry with `goalPushed=true` (copy-on-write — entries are immutable snapshots,
  so lock-free readers never race), so a (re)connecting `Subscribe` stream gets
  the standing operator override re-sent (`server.go:620`) and a later
  `Provision.Fetch` returns it in `Goals`. The persona/genesis baseline is NOT
  re-sent — only an operator push makes `goalPushed` true, so connect-time sends
  never clobber a richer genesis goal the host already holds.
- **Host side:** `subscribeDirectives` (`runtime/runhost_bootstrap.go:575`)
  applies `GOAL_REVISION` via `Host.SetLiveGoal` (`runtime/host.go:398`); the
  mesa director prefers the live goal over its genesis/persona goal each turn
  until it is replaced (`effectiveGoal` — `runtime/director_goals.go:27`). `PERSONA_REVISION` and
  `PEARL_REFRESH` directives are **logged only** today (recompile-on-revision is
  TODO `M-3` — now cheap, since the subscriber registry it was blocked on
  exists).

## End-to-end workflow (the up-to-200-host batch)

The driving use case: stand up a batch of up to 200 cradle hosts, each with its
own persona, with minimal manual wiring. The linchpin is one identity string
used everywhere:

> **`hostcfg name == RSC username == mesa host_id == persona key`**
> (the cradle daemon maps `Username: h.Name` and dials `Mesa: h.Mesa` —
> `cradle/registry.go:554-557`; mesa keys personas by the **authenticated**
> host_id — the request body is ignored, `server.go:592`; the host fetches its
> persona via `Provision.Fetch` by that id.)

Steps:

1. **Author/cook** N persona JSONs offline (`mesa/personacook`). *(exists, out
   of band)*
2. **Bulk-register** them in mesa: `mesa-ctl persona import <dir>` — host_ids
   from filenames. *(BUILT)*
3. **Generate the cradle hostcfg** matching those host_ids: `mesa-ctl fleet gen`
   (below). For prefix-named homogeneous batches the existing `template:` block
   already suffices (`cradle/hostcfg/examples/drones.hostcfg` expands
   drone[1..N]); the generator is for heterogeneous/named sets and guarantees
   the name↔host_id linkage by construction. *(BUILT)*
4. **Ensure RSC accounts exist** for the batch names on the OpenRSC server.
   *(PREREQUISITE — server-side, outside mesa/cradle; drone1..drone200 exist;
   arbitrary named batches need account creation — TODO `M-5`.)*
5. `cradle-server -config fleet.hostcfg` (`cmd/cradle-server`) launches the
   fleet; each host dials mesa as its name and `Provision.Fetch` returns its
   persona. *(exists)*

Ordering matters: a persona-driven host whose persona isn't registered gets
`NotFound` from `Provision.Fetch` (`server.go:596`). Register personas BEFORE
launch.

### `mesa-ctl fleet gen` (`cmd/mesa-ctl/fleet.go`)

Reads the registered persona set from mesa (`Admin.ListPersonas`) and emits a
cradle hostcfg on stdout whose host names == the registered host_ids, with
shared fields in a `defaults:` block. The output round-trips through
`hostcfg.Load`.

```
mesa-ctl fleet gen \
  --server localhost:43594 --mesa localhost:7077 \
  --password-env WESTWORLD_PASSWORD \
  [--goal "<shared goal>" | --goals goals.json] \
  [--match "drone*"] [--state memory] [--operator <name>] [--genesis=false] \
  > fleet.hostcfg
```

- Emits `defaults:` (server/mesa/password_env/state/operator/genesis) + a flat
  `hosts:` list of `{name: <host_id>}` entries — minimal per-host, names
  guaranteed to match mesa. (A `--template` output mode for uniform prefixes was
  sketched but never built; the flat list works for any names.)
- Goal placement: `--goal` (shared → defaults) or `--goals` (per-host JSON map →
  host entries); omit both to let goals ride from the persona/mesa. The
  generator only writes hostcfg fields — it never invents personas.
- This closes the loop: register → gen → run, with the naming contract enforced
  by construction rather than hand-sync.

## Deltas from the original v1 sketch

All small; the design otherwise shipped as written:

- `PersonaRecord` gained `name` (display name) and `prose` (the card the brain
  reads, returned on single get) — `mesa.proto:516`.
- The manifest import format was dropped (dir + NDJSON cover it).
- Admin auth is **fail-closed** (unset `$ADMIN_TOKEN` disables the API) rather
  than merely "checked".
- `persona set` + the generated projection columns were added beyond the plan;
  `goal push` and `fleet gen` shipped ahead of their phase.

## Open work

Tracked in [`docs/TODO.md`](../TODO.md) (the SSOT): **M-3** (v2 live
`PERSONA_REVISION` push + pearl/policy recompile-on-revision — unblocked by the
subscriber registry), **M-4** (v3 remainder: `mesa-ctl hosts` roster, `persona
recook`, history/rollback, `genesis <id>`), **M-5** (RSC account provisioning
for arbitrary batch names — external prerequisite), **M-6** (mTLS later,
`ListPersonas` paging when it bites).
