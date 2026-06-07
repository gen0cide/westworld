# host ↔ mesa — gRPC protocol spec (v2, game-native)

> The wire contract, distilled around the GAME LOOP. The host's currency is
> **DSL + game state**: it ships game SITUATIONS up and receives game MOVES
> (DSL) down. mesa's primary output is DSL — routines to run, routines to author,
> single actions, and the policy that selects among them — NOT abstract
> "decisions". Each isolated host owns its OWN connection and authenticates as
> its own `host_id`; the cradle is a lifecycle manager, not in the data path.
>
> Domain mapping: `Move`/`WriteRoutine`⇿`brain` RunRoutine/WriteRoutine/Idle sum
> type; `Decision`⇿`brain.Decision`; `Episode.relation`⇿`limbic.Entry` delta;
> `Directive{pearl_refresh}`⇿`pearl.Table`. Mirrors `mesa/client` Go types.

```proto
syntax = "proto3";
package westworld.mesa.v2;
option go_package = "github.com/gen0cide/westworld/mesa/proto;mesapb";

message HostRef { string host_id = 1; string persona_version = 2; }

// ───────── Game — the agent interface (the core) ─────────
service Game {
  rpc Act(Situation)  returns (Move);      // "what do I do now?" → a DSL move
  rpc Decide(Choice)  returns (Decision);  // narrow in-routine option pick (pearl-miss)
}

message Situation {                  // game state the host already knows
  HostRef host = 1;
  string  goal = 2;
  string  trigger = 3;              // routine_exhausted | novel | chat | low_hp | ...
  World   world = 4;
  repeated string recent = 5;       // recent salient events
  Affect  affect = 6;
  map<string,string> hints = 7;     // counterparty, last_outcome, ...
}
message World {
  int32 x = 1; int32 y = 2; string region = 3;
  int32 hp_cur = 4; int32 hp_max = 5; int32 combat_level = 6; double fatigue = 7;
  int32 inv_free = 8; repeated string inv_summary = 9;
  repeated string nearby_npcs = 10; repeated string nearby_players = 11;
}
message Move {                       // mesa's answer is DSL, not a choice string
  MoveKind kind = 1;
  string reasoning = 2;
  string routine_path = 3;          // RUN_ROUTINE: a library routine
  string routine_name = 4;          // WRITE_ROUTINE: name of the authored routine
  string dsl_source   = 5;          // WRITE_ROUTINE: the authored DSL (host parses+gates+runs)
  bool   quarantined  = 6;          // WRITE_ROUTINE: run once/gated; do NOT auto-promote to library
  repeated string args = 7;         // RUN/WRITE routine args
  string verb = 8;                  // DIRECT_ACTION: a DSL action name
  repeated string action_args = 9;
  int32  idle_seconds = 10;         // IDLE
}
enum MoveKind { RUN_ROUTINE = 0; WRITE_ROUTINE = 1; DIRECT_ACTION = 2; IDLE = 3; }

message Choice   { HostRef host = 1; string question = 2; repeated string options = 3; Affect affect = 4; }
message Decision { string choice = 1; string reasoning = 2; double confidence = 3;
                   string cache_key = 4; int64 cache_ttl_seconds = 5; }

// ───────── Knowledge — game recall (RAG, game-typed) ─────────
service Knowledge {
  rpc Recall(Query) returns (KnowledgeSet);   // deadline ~200ms
}
message Query        { HostRef host = 1; string text = 2; QueryKind kind = 3; int32 top_k = 4; }
enum QueryKind       { ANY = 0; PROCEDURAL = 1; ENTITY = 2; EPISODIC = 3; SOCIAL = 4; }
message KnowledgeSet { repeated KnowledgeItem items = 1; }
message KnowledgeItem{ QueryKind kind = 1; string text = 2; string dsl = 3;   // dsl set for procedural how-to
                       string provenance = 4; double score = 5; }

// ───────── Journal — game memory write (Mirror; async) ─────────
service Journal {
  rpc Remember(stream Episode) returns (RememberAck); // client-stream, fire-and-forget
}
message Episode {
  HostRef host = 1;
  string  idempotency_key = 2;      // host_id + event-hash + occurred_at (UNIQUE)
  string  kind = 3;                 // kill|death|trade|scam|quest_step|discovery|social|kv
  string  text = 4;
  double  importance = 5;           // local hint; mesa re-scores
  int64   occurred_at_unix = 6;
  RelationDelta relation = 7;       // optional social/trust delta (lossless, never deleted)
  map<string,string> tags = 8;
}
message RelationDelta { string name=1; double d_alpha=2; double d_beta=3; int32 d_encounters=4;
                        double total_value_traded=5; repeated string add_tags=6; }
message RememberAck   { int64 accepted = 1; int64 deduped = 2; }

// ───────── Provision — compiled game behavior pushed down (per-host stream) ─────────
service Provision {
  rpc Subscribe(SubscribeRequest) returns (stream Directive);
}
message SubscribeRequest { HostRef host = 1; int64 last_applied_id = 2; }
message Directive {
  int64 id = 1;                     // monotonic per host; ignore id <= last_applied
  DirectiveKind kind = 2;
  bytes payload = 3;                // typed by kind
}
enum DirectiveKind {
  DIRECTIVE_UNSPECIFIED = 0;
  ROUTINE_UPSERT      = 1;          // payload = a DSL routine for the library
  PEARL_REFRESH       = 2;          // payload = compiled selection policy (pearl.Table)
  PERSONA_REVISION    = 3;          // payload = compiled Cornerstone
  GOAL_REVISION       = 4;
  TRUST_DECAY         = 5;
  REVERIE_REBASELINE  = 6;
}

message Affect { double stress = 1; double confidence = 2; double valence = 3; }
```

## Data flows

### Act — the agent step (the core: state up, DSL move down)
```
conductor: GoalDirector.Next (local) can't choose an intent for the current goal
  → Game.Act(Situation{goal, trigger:"routine_exhausted", world, recent, affect})  ── gRPC, maturity-dialed
       mesa: Recall (RAG) + memory + persona → LLM → choose a MOVE:
         RUN_ROUTINE{path,args}      → host runs a library routine
         WRITE_ROUTINE{name,dsl,QUARANTINED} → host PARSES + VALIDATES + runs it
                                         under the pearl Gate (every action vetoable),
                                         LOGS it, and does NOT add it to the durable
                                         library (promotion = separate reviewed step)
         DIRECT_ACTION{verb,args}    → host runs one gated action
         IDLE{seconds}               → host waits
  → host executes the move via the existing conductor/interpreter
DEGRADE: ErrOffline/timeout → local GoalDirector fallback (idle/wander). Never blocks.
```

### Decide — narrow in-routine choice
```
routine: decide(["fight","flee"])  → pearl.TryDecide MISS
  → Game.Decide(Choice{question,options,affect})  → Decision{choice, cache_key}
  → host folds cache_key→choice into its local decision cache (next time local)
DEGRADE: timeout → pearl bias / safe default (decide! form).
```

### Recall — game knowledge
```
routine/Act needs to know "how do I smith a dagger" / "who is player X"
  → Knowledge.Recall(Query{text, kind})  →  KnowledgeSet
       PROCEDURAL items may carry `dsl` (a runnable snippet/routine)
DEGRADE: empty set; caller proceeds.
```

### Remember — game episodes + social (async, never on the hot path)
```
Limbic (host) batches game events: kills/deaths/trades/scams/quest-steps + trust deltas
  → Journal.Remember(stream Episode)   fire-and-forget, idempotency_key dedups
DEGRADE: mesa down → host buffers in the hostkv write-back journal, drains on reconnect.
```

### Provision — compiled game behavior down (per-host stream)
```
host (login): Provision.Subscribe(host_id, last_applied_id)   ── its own stream
mesa crons push, per host:
   ROUTINE_UPSERT(dsl)   ← mesa authored/improved a routine for this host's library
   PEARL_REFRESH(Table)  ← recompiled which-routine-when selection policy
   PERSONA_REVISION / GOAL_REVISION / TRUST_DECAY / REVERIE_REBASELINE
host applies strictly-increasing id; PEARL_REFRESH → atomic.Pointer[pearl.Engine] swap.
```

## The throughline
The host's behavior IS `{routine library (author-written + mesa-provisioned)} +
{selection policy (mesa-compiled pearl.Table)} + {escape hatch to Game.Act for
novel situations}`. **DSL is the lingua franca.** Act (synchronous "write me a
script now") and Provision/ROUTINE_UPSERT (async "here's your improved library")
are the same currency at two timescales. ALL mesa-authored DSL — inline (Act) or
provisioned — runs through the local parse/validate + pearl Gate; inline authored
routines are additionally **quarantined** (run-once, logged, no auto-promotion).

## Semantics
| Capability | RPC | Style | Deadline | On failure |
|---|---|---|---|---|
| Act | Game.Act | unary | maturity-dialed | local Director fallback |
| Decide | Game.Decide | unary | maturity-dialed | pearl bias / default |
| Recall | Knowledge.Recall | unary | ~200ms | empty set, proceed |
| Remember | Journal.Remember | client-stream | loose | hostkv journal, drain on reconnect |
| Provision | Provision.Subscribe | server-stream (1/host) | push | reconnect + resume from last_applied_id |

Phase A: these are Go structs behind `mesaclient.Client` (StubClient default).
Phase B: generate from this proto (tracked `mesa/proto/`); swap the `mesaclient`
impl to a gRPC stub — the `Client` interface + host seams are unchanged.
```
