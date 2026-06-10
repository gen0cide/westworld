# Brain — the LLM strategist

> **STATUS: BUILT — the live LLM lives in mesa, not in `brain/`** (verified 2026-06-10,
> branch HEAD `0bfa818`). The `brain/` package is the host-side seam: the `Strategist`
> interface (`Decide(ctx, Situation) -> *Decision`), the `Situation`/`Decision` types
> (`brain/brain.go`), and the deterministic `StubStrategist` (`brain/stub.go`) — the
> **offline fallback** when no mesa is wired. The real models sit behind three fixed
> per-tier seams in mesad (`mesa/mesad/server.go`: `actLLM`/`decideLLM`/`genesisLLM`),
> all built on the SDK-free Anthropic Messages client in `mesa/llm`. The host holds no
> API keys and makes no external calls — every LLM invocation is a mesa RPC. Verified
> live (tutorial-island completion via the `MesaDirector` Act loop). The original
> Phase-4 `AnthropicBrain` design (DecisionClass routing, `brain_calls` cost ledger,
> per-host budgets, earned `DSLSurface`) never manifested — it is preserved at
> `docs/archive/initial-brainstorming/brain-economics-design.md`. See
> `cognition-and-autonomy.md` for the surrounding loop.

## What "the brain" does

The brain is the LLM-driven layer of a host. It does NOT make every decision a host makes — most decisions are made by the deterministic routine interpreter, the reactive event handler, the pearl policy gate, and the cheap local routine-replay loop. The LLM comes in when:

1. **A routine completes and the host needs to decide what to do next** — the `MesaDirector` Act loop (`runtime/mesa_director.go`, `Next` at :572): each conductor turn it snapshots the host's state into a `mesaclient.Situation` and asks `mesa.Act` "what do I do now?". The `HybridDirector` (`runtime/hybrid_director.go`) sits in front: learned routines replay from the local library with **no** LLM call; only a novel situation escalates to Act.
2. **A novel situation arises that the current routine doesn't cover** — same Act escalation, plus the reactive tier's interrupt path: `mesa.ExtractDialog` (`runtime/reactive.go:474`) turns a latched dialog window into knowledge-ledger claims + one classified intent the host uses *deterministically* to decide whether to interrupt.
3. **Someone sends a chat message that requires a thoughtful response** — `mesa.Chat`, called from the social-reflex goroutine (`runtime/runhost_bootstrap.go:370`), off the Act loop. One cheap call, persona-grounded, 80-char-capped reply, silence on error (the reflex never wedges). Ask-mode is the same seam used proactively: the host asks a goal-blocking question it genuinely can't answer (no bluffing).
4. **It's time to write a new routine because no existing one fits** — Act returns a `WRITE_ROUTINE` move with authored DSL; mesa validates it before it ever reaches the host (see below).
5. **Periodic reflection on recent events** — runs mesa-side as the distillation crons (`mesa/mesad/cron.go`, `cron_insight.go`): Tier-1 batched claim extraction (the bulk) + the rare Tier-2 insight pass, concurrency-capped so they never starve Act/Chat/Decide. The host is not involved.
6. **Session genesis** — one heavy call at login (`mesa.Genesis`, invoked from `runtime/runhost_bootstrap.go:199`): mesa gathers the host's persona + episodes + relationships + standing goal and compiles this session's goal, mood baseline, and keyword attention ladder. The host runs cheap on the compiled output all session. (This replaced the old doc's "persona-consistency self-checks" — identity maintenance happens at session boundaries, not on a timer.)
7. **In-routine choices** — the DSL verbs `decide(options, context?)`, `contemplate_reality(question)`, `evaluate(situation)` (`runtime/actions_ambient.go:1285`) route through `Host.Strategist`. With mesa linked that is `mesaclient.AsStrategist` (`mesa/client/adapters.go:19`) → `mesa.Decide`.
8. **Operator console** — `mesa.AnalysisInterpret` (`runtime/analysis.go:224`) classifies a flat operator directive (analysis mode) into command / answer / hypothetical. Deliberately NOT in persona.

When asked for a move, the planner returns one of (manifested as `mesapb.MoveKind`, `mesa/proto/mesa.pb.go`):
- **RUN_ROUTINE** — execute a named routine from the library
- **WRITE_ROUTINE** — author a fresh DSL routine, then run it
- **DIRECT_ACTION** — do this one specific verb now
- **IDLE** — do nothing for a while

## The host-side seam: `brain/`

`brain.Strategist` is the only LLM-shaped interface the host knows: `Decide(ctx, Situation) (*Decision, error)` where `Situation` = question + optional `cognition.Bundle` + optional enumerated options, and `Decision` = choice + reasoning + confidence (`brain/brain.go`). `brain` imports cognition only; it does NOT import runtime, dsl, or mesa.

Wiring: `runtime.New` defaults `Host.Strategist` to `&brain.StubStrategist{}` (`runtime/host.go:469`); when a mesa link exists, bootstrap swaps in `mesaclient.AsStrategist(mc, username)` (`runtime/runhost_bootstrap.go:109`). The stub is deterministic and zero-I/O — Options-first, then question-prefix heuristics — so routine tests and offline hosts get stable answers with no key.

Two layers make `decide()` cheap before any RPC fires (`dslDecide`, `runtime/actions_ambient.go`):
- **Pearl first refusal**: the host's compiled policy may answer locally (`Pearl.TryDecide`) or hand back a persona-biased option ordering for the LLM on a miss (`runtime/host.go:108`).
- **Decision cache**: a bounded LRU memoizes Strategist verdicts for repeated pearl-miss decisions in materially-the-same state (`runtime/host.go:238`). Pearl hits are never cached — already free and authoritative.

## The mesa-side reality: three model tiers

There is no `DecisionClass`-keyed router. mesad holds three fixed `*llm.Client` seams, one per cost tier, set by `mesa/cmd/mesad` flags when `ANTHROPIC_API_KEY` is present (`mesa/cmd/mesad/main.go:53-85`):

| Seam | Default model (flag) | What runs on it |
|---|---|---|
| `actLLM` | `claude-sonnet-4-6` (`-act-model`) | `Act` / DSL authoring (high volume; the DSL manual is a prompt-cached prefix), `ExtractDialog` nuance escalation, Tier-2 insight cron |
| `decideLLM` | `claude-haiku-4-5-20251001` (`-decide-model`) | `Decide` option-picks, `Chat` (+ask mode), `AnalysisInterpret`, `ExtractDialog` base tier, Tier-1 consolidation cron |
| `genesisLLM` | `claude-opus-4-8` (`-genesis-model`) | `Genesis` only — rare, history-rich login compile (`mesa/mesad/genesis.go`) |

`mesa/llm.DefaultModel` is `claude-opus-4-8` (used when a client is constructed with an empty model id; `mesa/llm/anthropic.go:20`). Tier discipline is enforced at the call sites: the crons never route to `genesisLLM` — "Opus on bulk is the cardinal cost violation" (`mesa/mesad/cron.go:112`, `cron_insight.go:95-108`) — and `ExtractDialog` escalates at most to the Sonnet-class Act tier, never Opus (`mesa/mesad/act.go:537`).

**Degradation**: with `ANTHROPIC_API_KEY` unset all three seams are nil; the LLM RPCs return `Unavailable` and the host degrades to local behavior (persona provisioning still works; `decide()` falls back to the stub when no mesa is wired at all).

**Prompt caching** is real but simpler than the old five-chunk design: `llm.SystemBlock{Cache: true}` marks an ephemeral `cache_control` breakpoint (`mesa/llm/anthropic.go:52`). `act()` caches the shared, static DSL manual — the hand-written manual + the generated `spec.APIReference()` (`mesa/mesad/dslmanual.go:388`) — while the per-host persona card rides uncached behind it (`mesa/mesad/act.go:30-33`).

## Validation: author → validate → re-prompt

The manifested descendant of the old "response parsing" design, strictly stronger because it runs mesa-side, **before** the move round-trips to the host (`mesa/mesad/act.go`, `maxActAttempts = 3`):

- mesa parses the model's raw output into a `Move` and validates it: authored DSL must round-trip the real `dsl/parser` + `validator`; literal args are statically checked against the world name-set catalog so a hallucinated item/NPC/place name is caught (`mesa/mesad/catalog.go` — degrades to a no-op when `-facts` is unset).
- A rejected move re-prompts the model with the exact rejection ("YOUR PREVIOUS MOVE WAS REJECTED … fix exactly this problem").
- Exhausted retries → `IDLE` 3s rather than shipping a broken move.
- Unrecognised `goal_op` values are reset to a plain progress report and logged.

## What never manifested (archived)

The original Phase-4 body of this doc is preserved verbatim — with shipped-counterpart pointers — at `docs/archive/initial-brainstorming/brain-economics-design.md`:

- **`AnthropicBrain`** + `DecisionRequest`/`DecisionClass` model routing + per-class rate limiters → replaced by the three fixed mesad seams above.
- **`brain_calls` cost ledger + per-host budgets** — nothing tracks tokens or dollars today; open as **O-9** in `docs/TODO.md`.
- **Five-chunk cache plan** → replaced by the two-block cached-manual + persona assembly in `act()`.
- **`DSLGrammar`/`DSLSurface` earned vocabulary** (progressive disclosure, G2) — never built; the shipped manual teaches the full surface, and the DSL-manual analysis showed prose presence predicts usage. Whether to resurrect it as a cohort experiment is an open operator call — **C-12** in `docs/TODO.md`.
- **`MockBrain`** — never existed; tests use `StubStrategist` host-side and the test-only completer overrides mesa-side (`consolidateLLMOverride`/`insightLLMOverride`, `mesa/mesad/server.go`).
- **Tutorial-island DIRECT bootstrap mode** — never built as designed; the host completed tutorial island through the standard mesa Act loop.

Backlog: `docs/TODO.md` is the SSOT — the brain-relevant residue is **O-9** (LLM cost/token ledger), **C-12** (progressive disclosure / earned vocabulary, operator call), **C-16** (the surviving design opens from this doc's old "Open questions" section: persona-drift audit cadence, cost-cap behavior, rate-limit tier fallback), and **O-7** (cost dashboard re-scope).
