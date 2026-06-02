# Brain — the LLM strategist

> **STATUS: STUB** (verified 2026-05-31 against `brain/`). The
> package ships a `Strategist` interface (`Decide(ctx, Situation)
> -> *Decision`) plus the `Situation` / `Decision` value types and a
> deterministic `StubStrategist`. **No LLM is called.** The stub is a
> trivial heuristic: first option when `Situation.Options` is
> non-empty, else "yes" for should/closed yes-no questions,
> "Lumbridge" for where-questions, a canned line for what/how/why,
> "ok" otherwise — with `Confidence` a fixed per-branch constant
> (see `brain/stub.go`). The `AnthropicBrain` struct, tiered
> model routing (Sonnet vs Haiku), per-class rate limiters, prompt
> caching, the `brain_calls` cost ledger, and per-host budgets
> described below are **ASPIRATIONAL / Phase 4** — none of it
> exists. The "Open questions" section lists FUTURE decisions, not
> settled design. The real types in code are `Situation`,
> `Decision`, `Strategist` (`brain/brain.go`) and `StubStrategist`
> (`brain/stub.go`).

## What "the brain" does

The brain is the LLM-driven layer of a host. It does NOT make every decision a host makes — most decisions are made by the deterministic routine interpreter and the reactive event handler. The brain comes in when:

1. A routine completes and the host needs to decide what to do next
2. A novel situation arises that the current routine doesn't cover
3. Someone sends a chat message that requires a thoughtful response (not a canned one)
4. It's time to write a new routine because no existing one fits
5. Periodic reflection on recent events
6. Persona-consistency self-checks (occasional)

When the brain is asked to decide, it returns one of:
- **RunRoutine**(name, args) — execute a named routine from the library
- **WriteRoutine**(source) — author a fresh DSL routine, then run it
- **DirectAction**(action) — do this one specific action now (rare; mostly for chat)
- **Idle**(duration) — do nothing for a while

(Phase 4) The brain is invoked by the agent-driver loop at goal boundaries — a routine returns / is interrupted / the host idles — plus the `cognitiveLoop` for background appraisal; see `architecture.md` / `layers.md`. (Design, not built.)

(Phase 4) The cached Persona chunk carries the consolidated persona schema (HEXACO + mood + voice; see `personas.md`) so the host's reasoning is in-character; mood biases tone.

### Tutorial-Island bootstrap mode (Phase-4 design)

At host birth there is a naive DIRECT bootstrap mode that SHORT-CIRCUITS mesa retrieval (no `DecisionRequest` assembly): the tutorial's in-game system messages + NPC dialogue ARE the context, and the model emits one action at a time through the disclosed survival-core surface. It is the cheapest, earliest, mesa-free validation milestone, and the on-ramp that seeds the host's earned vocabulary + first routines. See `_research/host-bootstrap-and-knowledge-gating.md` §5. (Design, not built.)

## Interface

```go
package brain

type Brain interface {
    Decide(ctx context.Context, req DecisionRequest) (DecisionResponse, error)
}

type DecisionRequest struct {
    DecisionClass DecisionClass    // strategic, tactical, chat, reactive, reflection, ...
    Persona       *persona.Persona
    World         world.View       // read-only snapshot
    RecentEvents  []event.Event
    Knowledge     []RagChunk       // pre-retrieved by cognition layer
    Relations     []Relationship   // nearby players' relational records, pre-fetched
    RoutineIndex  []RoutineRef     // names + descriptions of available routines
    UrgentEvent   *event.Event     // optional: what triggered this call
}

type DecisionResponse struct {
    Decision   Decision
    TokenUsage TokenUsage
    Latency    time.Duration
    Reasoning  string             // captured for chain-of-thought logging
}

type Decision interface{ isDecision() }
type RunRoutine struct { Name string; Args map[string]any }
type WriteRoutine struct { Source string; Description string }
type DirectAction struct { Action action.Action }
type Idle struct { Duration time.Duration }
```

Single concrete implementation: `AnthropicBrain`. Mock for tests: `MockBrain`. Both satisfy `Brain`.

## Tiered routing inside AnthropicBrain

Different decision classes get different models. The routing is pure Go logic — no LLM picks which LLM to call.

| DecisionClass | Model | Why |
|---|---|---|
| `Strategic` (long-horizon goal updates, persona-defining choices) | Sonnet 4.6 | Better reasoning + persona consistency |
| `ScriptGen` (writing fresh DSL routines) | Sonnet 4.6 | Code generation benefits from larger model |
| `RoutineSelect` (which existing routine?) | Haiku 4.5 | Selection from list, easy |
| `Tactical` (mid-routine "what next") | Haiku 4.5 | With RAG context, Haiku is plenty |
| `ChatReply` (responding to other players) | Haiku 4.5 | Casual RSC chat is easy |
| `Reactive` (urgent: HP low, flee/fight) | Haiku 4.5 | Speed matters more than depth |
| `ImportanceScore` (rating event memorability) | Haiku 4.5 | Tiny prompts, batchable |
| `TrustGrade` (rating a relational event's cooperative/defective valence + severity w∈[0.3,8]) | Haiku 4.5 | Tiny, batchable — like `ImportanceScore` |
| `Reflection` (consolidating recent events) | Sonnet 4.6 | Self-reflection benefits from larger model |
| `PersonaAudit` (occasional self-check) | Sonnet 4.6 | Rare, quality-critical |

```go
type AnthropicBrain struct {
    client *anthropic.Client
    models map[DecisionClass]string
    
    // Per-class rate limiters and cost tracking
    limiters map[DecisionClass]*rate.Limiter
    
    cognition cognition.Client  // for retrieval before LLM call
}

func (b *AnthropicBrain) Decide(ctx context.Context, req DecisionRequest) (DecisionResponse, error) {
    model := b.models[req.DecisionClass]
    prompt := buildPrompt(req, model)
    if err := b.limiters[req.DecisionClass].Wait(ctx); err != nil {
        return DecisionResponse{}, err
    }
    return b.call(ctx, model, prompt)
}
```

`TrustGrade` (Phase-4) is invoked on the mesa relational-update path (see `mesa.md` / `_research/social-graph-and-trust-ledger.md`): it judges whether a relational event was cooperation or defection given context, returns `{cooperative bool, w float}`, and writes a `brain_calls` row like any other call.

## Prompt structure

Five cacheable chunks plus one volatile chunk per call. The cache TTL (5 min default; 1 hour for slow-changing data) is critical for keeping per-call input cost low.

| Chunk | Cache strategy | Approximate tokens |
|---|---|---|
| Persona (name, traits, north star, vocabulary, quirks) | Cached 1h | 500 |
| `DSLGrammar` (static syntax skeleton — NO verb names) | Cached forever (global) | ~400 |
| `DSLSurface` (per-host EARNED symbol table) | Cached per-host, invalidated on graduation (K candidates not cached) | ~400 |
| Routine index (private + public names + descriptions) | Cached 5min, refreshed on routine save | 300 |
| Recent journal summary | Cached 5min | 300 |
| Volatile: current world state + retrieved knowledge + nearby relations + urgent event | Not cached | 800 |

Total per Sonnet call ~3000 tokens input, ~300 tokens output. With aggressive caching, effective input cost drops 5-10x vs uncached.

The two routine-authoring chunks are split along the line **the grammar is free; the symbol table is earned** (Phase-4 design, per `_research/host-bootstrap-and-knowledge-gating.md` §2):

- **`DSLGrammar`** — the static SYNTAX SKELETON with NO verb names: how to *shape* a routine (the `on` / `when` / `select` / `defer` / `try` constructs, the bang `!` operator + the `Result`/`Error` model, lambdas, f-strings, control flow), the meta-rules (handlers can't yield, actions forbidden in `proc`/`require`), and 1-2 worked examples using only survival-core verbs. It is *grammar, not vocabulary*, so it is legitimately cached forever and global across all hosts.
- **`DSLSurface`** — the per-host EARNED symbol table, assembled fresh at decision time as `survival_core ∪ host.vocabulary ∪ retrieved_candidates(goal, K)`: the static survival-core allowlist, the host's earned vocabulary (the mesa `bot_vocabulary` table — verbs/accessors the host has used successfully), and a small budget of K goal-retrieved candidate verbs (top-K from the embedded DSL-spec corpus). Cached PER-HOST (not globally), invalidated when the host's vocabulary graduates; the K candidates are goal-volatile and uncached. Assembled by `cognition.PrepareDecision` (slice 5).

The grammar is free; the symbol table is earned — advanced automation must EMERGE from learning, not be front-loaded from the full API (`research-goals.md` §1a). A host only ever sees verbs that are in `survival_core`, ones it has used successfully, or a small budget of goal-relevant candidates; this is what makes scripting competence a learned, observable trajectory (the falling punt rate). The per-call token budget is unchanged; only the cache *key* for the surface moves from global to per-host.

## Response parsing

The LLM is instructed to respond in a structured format. Anthropic tool-use is the cleanest path:

```json
[
  {"name": "run_routine", "input": {"name": "fish_at_swamp", "args": {}}},
  {"name": "write_routine", "input": {"source": "...", "description": "..."}},
  {"name": "say_chat", "input": {"message": "hey"}},
  {"name": "idle", "input": {"seconds": 30}}
]
```

The brain's response parser converts these tool calls into typed `Decision` values. Validation:
- `run_routine`: routine must exist in the host's library (private + public)
- `write_routine`: source must parse cleanly in the DSL (round-trip through the lexer/parser before accepting); description required. Additionally (Phase-4, per `_research/host-bootstrap-and-knowledge-gating.md` §2.3) the generated source must use ONLY the host's currently-disclosed `DSLSurface` — a real spec verb the host has not yet earned is rejected with a proposed `ERR_VERB_NOT_LEARNED` (vs `ERR_UNKNOWN_ACTION` for a verb not in the spec at all). This typed rejection is also how the host learns the real interface incrementally: it corrects on the next pass from the typed error, the same teach-by-typed-error principle the validator already embodies
- `say_chat`: message goes to the chat queue, not sent immediately (chat is async)
- `idle`: duration capped at 5 minutes

Invalid responses get one retry with a "your previous response was invalid because..." prompt. Second failure logs and returns `Idle{30s}` as a safe fallback.

## Cost accounting

Every brain call writes a row to mesa's `brain_calls` table: prompt, response, model, tokens, cost (computed from model pricing × tokens), latency. This is BOTH the chain-of-thought capture for observability AND the cost ledger for budgeting.

Aggregate queries answer:
- Cost per host per day (and which hosts are expensive)
- Cost per cohort (and which cohort is most cost-efficient per "interesting event")
- Cost per decision class (where is our budget going?)
- Token usage trends over time

## Per-host budgeting

Each host has a soft budget (default: $1/day in Haiku-equivalent). When approaching the limit:
- Tactical calls drop in priority (more reliance on routine continuation, less re-strategizing)
- Reflection cadence relaxes (every 30 min → every 2 hours)
- Persona audits skip

This is a graceful degradation, not a hard cutoff. A host that's having an interesting day (lots of social interaction, novel situations) gets more budget; a host fishing for 8 hours straight gets less.

## Mock brain for tests

```go
type MockBrain struct {
    Responses []DecisionResponse  // canned, returned in order
}

func (m *MockBrain) Decide(ctx context.Context, req DecisionRequest) (DecisionResponse, error) {
    if len(m.Responses) == 0 { return DecisionResponse{Decision: Idle{30 * time.Second}}, nil }
    r := m.Responses[0]
    m.Responses = m.Responses[1:]
    return r, nil
}
```

Tests construct scenarios by pre-loading responses. No Anthropic key needed in CI.

## Open questions

- **Prompt caching strategy**: which chunks deserve the 1h cache TTL vs default 5min? Probably persona + DSL grammar + tool defs go 1h; everything else 5min.
- **Persona audit frequency**: how often does an agent self-check for persona drift? Probably weekly. Cheap and rare.
- **Cost overrun behavior**: hard caps or soft degradation only? Soft for v1, hard caps available as admin override.
- **Multi-model fallback**: if Sonnet is rate-limited, fall back to Haiku for strategic calls? Probably yes — better to make a degraded decision than no decision.
