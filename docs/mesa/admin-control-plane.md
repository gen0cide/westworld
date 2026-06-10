# mesa Admin control plane + `mesa-ctl`

**Status:** design / plan (no code yet)
**Decision:** gRPC `Admin` service + a `mesa-ctl` client. Ship persona CRUD + **bulk
load** first; live host-push and niceties are follow-ons.

## Problem

mesa already stores personas durably (Postgres `personas` table — `ltm.go:110`),
restores them on boot (`LoadPersonas` — `server.go:137`), and `Register`
(`server.go:100`) already does the whole job *live*: validate → render prose +
system prompt → put in the in-memory registry → upsert to Postgres → authorize
the host. So persistence is ~done and the `-host host_id=persona.json` flag is
really just a **first-run seed**.

The gap is operational: **there is no way to add/change/list/remove a persona
without restarting mesad**, and **no bulk path at all**. `Register` is only
called from `main()` via `-host`. Today, loading 200 personas means either 200
`-host` flags + restart, or a raw Postgres bulk insert + restart. Neither is a
real mechanism.

## Goal

An operator can manage personas against a *running* mesad — one at a time or 200
at once — with no restart, each persona validated at ingest, via a `mesa-ctl`
CLI. Storage stays in Postgres (already the system of record).

## Why no restart is "free"

`Register` already updates the in-memory registry (`s.reg`) *and* Postgres in one
call (`server.go:100-116`). Exposing `Register` over RPC means a `mesa-ctl put`
takes effect on the next `Act`/`Decide`/`Provision.Fetch` immediately — no
restart, no reconnect. A connected host picks up its new persona on its next
`Provision.Fetch`; pushing it *live* to an already-running host is phase 2 (see
below) over the existing `Provision.Subscribe` stream.

## Design

### 1. proto — a new operator-only `Admin` service (`mesa/proto/mesa.proto`)

```proto
service Admin {
  // Bulk upsert: client streams personas, server validates + registers each,
  // returns a per-item report. One bad persona does NOT fail the batch.
  rpc PutPersonas(stream PersonaUpsert) returns (BatchResult);
  rpc GetPersona(HostRef) returns (PersonaRecord);
  rpc ListPersonas(ListPersonasRequest) returns (PersonaList);
  rpc DeletePersona(HostRef) returns (AdminAck);
}

message PersonaUpsert { string host_id = 1; bytes persona_json = 2; }
message ItemResult    { string host_id = 1; bool ok = 2; string error = 3; }
message BatchResult   { int32 ok = 1; int32 failed = 2; repeated ItemResult items = 3; }
message PersonaRecord { string host_id = 1; bytes persona_json = 2; string updated_at = 3; }
message PersonaList   { repeated PersonaRecord personas = 1; } // json omitted unless requested
message ListPersonasRequest { bool with_json = 1; }
```

- **Client-streaming `PutPersonas`** is the bulk primitive: 200 personas go up in
  one call with backpressure, each validated server-side, per-item results back.
  Single-put is just a stream of one.
- Handlers are thin: `PutPersonas` loops `s.Register(item.host_id, persona)` —
  reusing the existing validate/render/persist/authorize path — collecting
  failures instead of aborting. `Get/List` read the registry / `personas` table;
  `Delete` removes from `s.reg` + `DELETE FROM personas`.

### 2. Auth — operator tier, separate from per-host bearer

Today auth is a per-host bearer-token map (`tokens` — `server.go:48`); a host must
not be able to rewrite another host's identity. Add an **operator credential**
checked for `Admin/*` methods in the existing `UnaryAuth`/`StreamAuth`
interceptors:

- v1: an `ADMIN_TOKEN` env (mesad reads it; `mesa-ctl` sends it as metadata).
- later: mTLS / per-operator tokens.

Admin methods are gated on the operator token; all other services keep host-token
auth unchanged.

### 3. `cmd/mesa-ctl` (mirrors `cmd/cradle-ctl`)

```
mesa-ctl -addr :7077 persona put <host_id> <file|->     # single (stream of 1)
mesa-ctl persona import <dir|manifest|->                # BULK
mesa-ctl persona ls [--json]
mesa-ctl persona get <host_id>
mesa-ctl persona rm <host_id>
```

**Bulk `import` input formats** (the 200-persona path):
- a **directory** of `<host_id>.json` (filename = host_id), or
- a **manifest** `{ "host_id": "path.json", ... }`, or
- **NDJSON** on stdin: one `{"host_id":..., "persona":{...}}` per line.

`import` opens a `PutPersonas` stream, sends each, and prints the per-item report
(`✓ Delores`, `✗ Teddy: validation: …`), exiting non-zero if any failed.

### 4. Storage (already there; small adds)

- `personas (host_id PK, persona_json jsonb, updated_at)` — reuse as-is;
  `UpsertPersona` (`ltm.go:407`) is already idempotent.
- Bulk = N upserts; fine per-row for v1, optionally one tx for atomicity.
- *(optional)* `persona_history (host_id, persona_json, updated_at)` append-on-
  upsert so `mesa-ctl persona get --history` / rollback is possible later. The
  persona docs already anticipate revisions.

## Phasing

- **v1 (this slice):** `Admin` service (PutPersonas/Get/List/Delete) + `ADMIN_TOKEN`
  auth + `mesa-ctl` with single put, bulk `import`, ls/get/rm. Kills the restart
  and gives the 200-persona path. Host picks up changes on next `Fetch`.
- **v2 — live host push:** on upsert, send a `PERSONA_REVISION` `Directive` over
  the existing `Provision.Subscribe` stream to the host if connected, so a running
  host re-runs `CompilePolicy` without reconnecting. Needs a server-side registry
  of active subscriber channels (Subscribe currently just sends goals once then
  blocks — `server.go:426`).
- **v3 — operator niceties:** `mesa-ctl hosts` (list connected/registered),
  `persona recook <id>` (trigger the personacook prose cook), `goal push`,
  persona history/rollback, `genesis <id>`.

## Migration / compatibility

- Keep `-host` as a one-shot seed (or fold seeding into `mesa-ctl persona
  import`). No breaking change.
- Bulk-loaded personas are assumed already authored/cooked offline (the
  personacook output); `import` is the ingest of pre-cooked persona JSON. Cook is
  out of band (v3 `recook` can trigger it server-side later).

## End-to-end workflow (the up-to-200-host test)

The driving use case: stand up a batch of up to 200 cradle hosts, each with its own
persona, with minimal manual wiring. The linchpin is one identity string used
everywhere:

> **`hostcfg name == RSC username == mesa host_id == persona key`**
> (cradle sets `Username: h.Name` and dials `Mesa: h.Mesa` — hostcfg.go:186-189;
> mesa keys personas by the authenticated host_id; the host fetches its persona via
> `Provision.Fetch` by that id.)

Steps:
1. **Author/cook** N persona JSONs offline (personacook). *(exists, out of band)*
2. **Bulk-register** them in mesa: `mesa-ctl persona import <dir>` — host_ids from
   filenames/manifest. *(NEW — Admin service)*
3. **Generate the cradle hostcfg** matching those host_ids: `mesa-ctl fleet gen`
   (below). For prefix-named homogeneous batches the existing `template:` block
   already suffices (drones.hostcfg expands drone[1..200]); the generator is for
   heterogeneous/named sets and guarantees the name↔host_id linkage. *(NEW)*
4. **Ensure RSC accounts exist** for the batch names on the OpenRSC server.
   *(PREREQUISITE — server-side, outside mesa/cradle; drone1..drone200 already
   exist. New named batches need account creation — see open questions.)*
5. `cradle-server -config fleet.hostcfg` → launches the fleet; each host dials mesa
   as its name and `Provision.Fetch` returns its persona. *(exists)*

Ordering matters: a persona-driven host whose persona isn't registered gets
`NotFound` from `Provision.Fetch` (server.go:413). Register personas BEFORE launch.

### `mesa-ctl fleet gen` (the smooth hostcfg generator)

Reads the registered persona set from mesa (`Admin.ListPersonas`) and emits a
cradle hostcfg whose host names == the registered host_ids, with shared fields in
`defaults:`:

```
mesa-ctl fleet gen \
  --server localhost:43594 --mesa localhost:7077 \
  --password-env WESTWORLD_PASSWORD \
  [--goal "<shared goal>" | --goals goals.json] \
  [--match "drone*"] [--state memory] [--genesis=false] \
  > fleet.hostcfg
```

- Emits a `defaults:` block (server/mesa/password_env/state/...) + a `hosts:` list
  of `{name: <host_id>, goal: <...>}` — minimal per-host, names guaranteed to match
  mesa. (Could also emit a `template:` when names share a prefix.)
- Goal placement: `--goal` (shared) or `--goals` (per-host map); omit to let goals
  ride from the persona/mesa. The generator only writes hostcfg fields — it does
  not invent personas.
- This closes the loop: register → gen → run, with the naming contract enforced by
  construction rather than hand-sync.

## Open questions

1. `ADMIN_TOKEN` env (simple) vs mTLS now? (Lean: env now.)
2. Bulk upsert atomic (one tx, all-or-nothing) vs per-item best-effort (recommended,
   so one bad persona of 200 doesn't block the rest)?
3. Does `List` ever need to page (200+ personas)? Probably yes eventually; add a
   cursor when it bites.
4. **RSC account provisioning** for batch names is a hard external prerequisite
   (`Provision.Fetch` and login both key on the name). drone1..drone200 exist
   already; arbitrary named batches need accounts created OpenRSC-side. Is there
   tooling, or is `mesa-ctl`/`cradle` expected to grow an account-creation step?
   (Out of scope for the Admin service itself, but part of the full workflow.)
5. `fleet gen` output shape: a flat `hosts:` list (works for any names) vs a
   `template:` (only when names share a prefix). Default to `hosts:`; offer
   `--template` when uniform.
