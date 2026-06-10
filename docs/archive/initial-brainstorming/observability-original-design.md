> **ARCHIVED (initial brainstorming), 2026-06-10.** This design was superseded or manifested
> differently. `cmd/delos`, the `obs` package, and the `brain_calls` ledger were never built as
> named; the intent landed as the cradle daemon (`cradle/` + `cmd/cradle-server` web UI/API),
> the per-host `debughttp/` control plane, the `agent_thought` decision stream
> (`runtime/decisions.go`), and host→mesa telemetry (`runtime/telemetry.go` → the mesa
> `metrics` table). The living doc is `docs/observability.md` (rewritten same-commit).
> Still-open residue lives in `docs/TODO.md` (O-9 cost/token ledger, O-7 analytics/cost
> dashboard/perturbation toolkit). Kept verbatim below for the record.

# Observability — the technician tablets

## The metaphor

In Westworld, the technicians at Delos use tablets to observe hosts: their current narrative, their behavior matrix, conversation history, memory state, behavioral patterns. They can pause a host, edit attributes, replay scenes. The tablets are how the operators *understand* the system.

We want the same: a web UI served from `cmd/delos` that exposes any host's complete state, captures the chain-of-thought from every brain call, and supports population-scale analytics.

## The data plumbing (Phase 3+ groundwork)

The UI itself is Phase 5. The DATA the UI consumes must flow from Phase 3 onward, or we won't have it when we want it.

Three data streams:

### 1. Brain call ledger

Every brain call writes a row to mesa's `brain_calls` table with:
- `bot_id`, `called_at`
- `decision_class` (Strategic / Tactical / Chat / etc.)
- `model` (e.g., "claude-sonnet-4-6")
- `prompt` (the full constructed prompt)
- `response` (the model's response)
- `input_tokens`, `output_tokens`, `cost_usd`, `latency_ms`
- `reasoning` (if the model produced reasoning content, captured separately)

This is the chain-of-thought archive. Replayable. Filterable. Aggregatable.

### 2. Event stream

Every event published on a host's event bus is also written to a structured log (via the `obs` package). Local cradle logs go to stdout (structured JSON via slog). Subscribers (delos UI) can stream live events from a specific host via a WebSocket or SSE endpoint exposed by cradle.

The event log includes:
- Inbound packet decoded → typed event
- Outbound action sent
- State mirror updates
- Reverie injections
- Routine step transitions

### 3. State snapshots

On admin request, a host can serialize its current state to JSON:
- Persona
- Current world state (position, inventory, stats, fatigue, poison, nearby entities)
- Current routine (name, step, local variables)
- Working memory ring buffer
- Recent reflections (from mesa cache)
- Pending chat queue

This is the "freeze a moment in time" view. Useful for debugging or for the research record.

## The delos UI (Phase 5)

Web app served by `cmd/delos`. Single-page-ish, probably HTMX or a thin JS framework — not a heavy React build. Sections:

### Host roster
```
[Hosts: 247 active / 247 total]
[Filter: cohort=lumbridge_regular  status=active  location=varrock]

alex          Lumbridge       fishing      level 12  ●● [tap]
ruth          Varrock square  idle         level 8   ●●● [tap]
jimbob        Edgeville bank  banking      level 22  ● [tap]
...
```

Status dots: green = active, yellow = idle, red = struggling/crashed.

### Tap view (per-host)
```
┌─────────────────────────────────────────────────────────────────┐
│ ruth  |  Varrock square (122, 506)  |  level 8 |  cohort: varrock_loiterer
├─────────────────────────────────────────────────────────────────┤
│ PERSONA                                                          │
│ Ruth, 14yo girl from Lumbridge, shy, likes cooking and chatting │
│ with cats. North star: "Make a few friends I see often."        │
├─────────────────────────────────────────────────────────────────┤
│ CURRENT ROUTINE                                                  │
│ social_loiter_varrock (step 3 of 8)                              │
│ Local vars: idle_count=2, last_chat_target=jimbob               │
├─────────────────────────────────────────────────────────────────┤
│ NEARBY                                                           │
│ jimbob (friend, 14 trades), alex (acquaintance, 2 chats),       │
│ randomdude_42 (stranger)                                         │
├─────────────────────────────────────────────────────────────────┤
│ RECENT EVENTS                                                    │
│ [12s ago] ChatReceived from jimbob: "any good fish here?"       │
│ [4s ago]  ChatQueued: "i think the swamp has shrimp"             │
│ [2s ago]  WalkAction to (124, 506)                               │
├─────────────────────────────────────────────────────────────────┤
│ BRAIN STREAM (live)        ▶ [tap to subscribe]                  │
│ ...                                                              │
├─────────────────────────────────────────────────────────────────┤
│ [Pause]  [Force reflection]  [Inject event]  [Edit persona]     │
└─────────────────────────────────────────────────────────────────┘
```

The "Brain stream" section is the live tap — when an admin clicks subscribe, the UI receives a stream of brain calls as they happen: the prompt going in, the response coming out, the parsed decision. Indistinguishable from watching a host think.

### Cohort analytics view
```
COHORT: lumbridge_regular  (47 hosts)

Avg lifetime XP gain (last 24h):    34,820
Avg trades / host / day:             3.2
Avg chat sends / host / day:         42
Cost / host / day:                   $0.42
Most-used routines:                  fish_swamp (used 1200×), train_combat_chickens (980×)

Compared to varrock_loiterer cohort:
  +28% chat   -41% xp gain   -22% trade volume   -14% cost
```

These are the kinds of queries that *only* work because mesa is centralized.

### Brain cost dashboard
```
TOTAL BRAIN COST (last 24h):  $186.42
  Strategic (Sonnet):          $94.20  (51%)
  ScriptGen (Sonnet):          $42.80  (23%)
  Tactical (Haiku):            $18.00  (10%)
  Chat (Haiku):                $14.20  (8%)
  ImportanceScore (Haiku):      $7.50  (4%)
  Reflection (Sonnet):          $6.40  (3%)
  PersonaAudit (Sonnet):        $3.32  (2%)

[Per-host cost distribution chart]
[Top 10 most-expensive hosts]
```

Per Alex's observation: cost transparency is foundational. We need to know exactly where the money is going to make rational scaling decisions.

## Admin actions

The UI exposes admin actions per-host (auth-gated by admin token):

- **Pause / resume**: stop a host's loop without disconnecting from the server
- **Force reflection**: trigger an immediate reflection-generation call
- **Inject event**: simulate an inbound event for testing reactions
- **Edit persona**: update persona JSON in mesa (host re-reads on next strategist call)
- **Restart**: kill and respawn the host
- **Snapshot**: dump current state to a research record
- **Adjust cohort**: move a host to a different cohort (rare, for experiments)

## Privacy considerations (relevant for any future open-sourcing)

If we ever publish the swarm transcripts as a research dataset, the captured brain prompts and responses contain agent "thoughts" that could be considered private. Even though no real humans are involved, the prompts may contain content about other agents that — once published — exists in a "social" form.

For now: it's all our data. We make decisions about publication later. The system records everything.

## What we don't need

- Real-time graphs of every metric. Quarterly dashboards are fine.
- Mobile-optimized UI. This is a desktop research tool.
- Multi-tenancy. One operator (Alex + collaborators) owns the population.

## Phases

| Phase | Observability work |
|---|---|
| 0-2 | Structured logs only; no UI |
| 3 | Brain call ledger writes; event log writes; mesa schema for the ledger |
| 4 | Admin API on cradle for state snapshots + event stream subscription |
| 5 | The web UI itself; cohort analytics; cost dashboard |

## Open questions

- Frontend tech: HTMX vs Vue vs raw HTML+JS? Probably whatever's easiest for a research tool. Will decide in Phase 5.
- Real-time vs polling: brain stream and event tap want real-time; rest can poll every few seconds.
- Storage of historic prompts: prompts can be large. Keep them in Postgres TEXT columns indefinitely? Probably yes for v1; archive to object storage if needed later.
