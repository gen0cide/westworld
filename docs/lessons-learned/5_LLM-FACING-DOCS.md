# LLM-facing docs: the manual is the program

*The only reader of our DSL manual is a language model that copies shapes. We taught
it a parse error as the canonical example, six reviewers missed it, and a thousand
`note()` calls went into a journal nobody could read back. Meanwhile a `Cache:true`
flag that silently did nothing quietly defeated the fleet cost model.*

---

## The problem as experienced

By 2026-06-09 the host had a real language — typed Results, event handlers,
watchers, `select`, a tiered memory store, ~114 view accessors — and the live
evidence said the LLM was playing with a fraction of it. A full inventory of 69
live LLM-authored routines (165 Act calls, pulled from `/tmp/ww-cradle.log`)
against the parser/validator/interp/spec surface found
(`docs/_research/dsl-manual-analysis-2026-06-10.md`):

- **Heavy use of a narrow slice:** `if` (564), f-strings (332), `wait` (286),
  `converse` (195), `go_to` (142), bare `.err != null` checks (194).
- **Zero use of the structural features:** on-handlers (one failed attempt, then
  abandoned forever), `select`, `try/recover`, `defer`, `require`, `proc`,
  `elif` — all 0/69.
- **The smoking gun:** 194 bare `.err` null-checks, **zero** `.err.code` typed
  branches. The 20-code error taxonomy — the engine's biggest structural advantage
  over the reference bot — was functionally dead.
- **`note()` called ~1000 times as a write-only memory.** The manual's own advice
  was "note in your journal which shops you've tried" (old `dslmanual.go:150`) —
  but `note()` has no read-back. Meanwhile `remember`/`recollect`/`forget`, a fully
  built tiered storage layer, sat at zero usage because no prose ever taught them.
- **~10 f-string parse-error retry loops** (multi-expression placeholders, escaped
  quotes). The model's documented coping strategy in the log: "removing all
  f-strings."
- Hosts were **losing with a better hand**: richer perception views than Plutonium,
  invisible to the author ("we never showed our cards" —
  `docs/_research/plutonium/dsl-gap-analysis.md`).

Then the comparative study read the manual itself and found the teaching was not
just incomplete but *wrong* in load-bearing places:

- **The manual taught `&&` as its canonical compound condition** (old
  `dslmanual.go:132`, the toll-decision example), and three spec docstrings did
  too — but the lexer had only the keywords `and`/`or`/`not` (`token.go:275-277`
  at the time). Every two-condition routine copied from the manual was a
  guaranteed hard parse error. **"Six prior reviewers' drafts missed this until
  one caught it"** (`dsl-gap-analysis.md:216`) — arguably the single
  highest-frequency authoring failure, taught *by* the manual.
- **The flagship mining example was a runtime error:** `m = search_map(...); if
  m.length > 0 { ... }` — `search_map` returns a Result; `m.length` is "result has
  no field" (`dsl-manual-analysis` §cut; confirmed against `interp/error.go:220-234`).
  The most-copied example modeled the bug the later Result section warned against.
- **The validator's own error message recommended syntax that fails validation:**
  the repeat-until message suggested `timeout 30s`, which itself dies with
  "unbound identifier 's'" (unit suffixes are legal only in `select`;
  negative-tested in the manual analysis; old `validator.go:332`).
- **The generated appendix lied by omission while claiming completeness:** its
  header declared "if it is NOT listed, it does not exist" (`dsl/spec/manual.go:19`)
  while `world.scenery`, `world.boundaries`, `world.dialog`, `inventory.capacity`,
  `bank.has/count` and every per-entity view field existed at runtime with zero
  spec rows (`dsl-gap-analysis.md:7`). The reference actively disowned real
  capabilities.
- **A half-truth taught assume-success:** the project line "typed Results beat
  string-matching" was only half true. Travel and item-missing are typed;
  world-interaction verbs (`interact_at`, `use`-on-scenery, `open_boundary`) are
  fire-and-forget `conn.Send` returning `Ok(null)` even against an empty tile
  (`docs/_research/plutonium/dsl-crossref.md:426`, verified at
  `action/use.go:38-50`). A manual implying interactions return typed failures
  teaches the LLM to assume success and loop on silent no-progress
  (`dsl-synthesis.md:66,143`).

And on the cost side, the same disease — a doc-shaped artifact (a flag) read as a
promise nobody verified: **H18**. Both distillation crons marked their analyzer
prefixes `Cache:true`, but the prefixes (~381–509 tokens on Haiku, ~417–556 on
Sonnet) were far below the model minimum cacheable lengths (4096 / 2048). **A
`cache_control` breakpoint below the model minimum is SILENTLY ignored — no error,
`cache_creation_input_tokens` stays 0.** Every consolidation call (one per active
host per 60s tick, up to 64 hosts) and every insight call (per host per 180s) paid
full input price forever; the "cheap × volume" cost model the 200-host go-live
rested on did not hold (`docs/cognition-audit-findings.md` H18).

## The false leads

**False lead #1: the appendix counts as teaching.** The shipped prompt was
hand-written prose + a spec-generated "COMPLETE API REFERENCE" appendix, and the
working theory was that listing a verb made it available. The usage tally
refuted this flatly: of the 12 verbs with >20 live uses, all 12 were taught with
worked prose examples; appendix-only surface got ~zero pickup. The lone
counterexample proves the rule — `bank.*` was used from the appendix alone *only
because the goal forced banking* (`dsl-manual-analysis`, action-surface
observations). The `hits.find(h => h.reach == "open")` and
`for r in scan_for("rock")` idioms were reproduced near-verbatim from the manual's
two worked examples. The LLM uses exactly and only what the examples model.

**False lead #2: review catches taught-wrong examples.** Six reviewer drafts read
the `&&` toll example and passed it. Reviewers (human and LLM alike) read examples
as *illustrative prose*; only a machine that parses them reads them as *code*. The
same blindness let the flagship example's `m.length` Result bug and the
`walk_to(FARX, FARY)` unbound-placeholder pseudo-example (fails bind-before-use if
copied) survive every pass.

**False lead #3: patch the manual incrementally.** Two same-day patches
(`937f707`, `aef6858`, both 2026-06-08 — "teach the planner scan_for, f-string
rules, search_map .val"; "dslmanual NPC-dialog…") each taught the
diagnosis-of-the-day. Neither was preceded by an inventory of what the full
surface was versus what was taught versus what was used, so the structural holes —
on-handlers unlearnable from prose (top-level-only is a parser rule no example
showed), the dead error taxonomy, zero entity fields taught anywhere in the
shipped prompt — persisted through every patch.

**False lead #4: trust the toolchain's own voice.** The validator message
recommending `timeout 30s` is the purest specimen: an error message is the doc the
model reads at its moment of maximum attention (a retry), and it taught a form
that re-fails. Same family: the `equipment_changed` docstring claimed bare
handlers work (arity is enforced exactly), and the appendix rendered a phantom
`host.name` accessor with no runtime root behind it (`b42a52b` commit message).

**False lead #5: the earned-vocabulary design.** An early design direction had the
grammar free but the symbol table *earned* — capabilities disclosed progressively.
It was never built, and the evidence ran exactly opposite: full disclosure with
worked examples is what made hosts competent, and *withheld* prose (not withheld
capability) was already producing the observed incompetence
(`docs/_research/docs-audit/execution-plan.md` decision A3).

**False lead #6: over-claiming as harmless marketing.** "Typed Results" as a
headline win survived until the cross-reference checked which verbs actually
produce the codes. Shipping the runtime ack later without fixing the manual first
would have left the LLM assuming success in the interim — the synthesis's explicit
sequencing note (`dsl-synthesis.md:50`).

**False lead #7: `Cache:true` means cached.** The flag is a request, not a
contract. The SDK-free client emitted it with no min-length check, the API
silently dropped it, and nothing in telemetry surfaced the zero-hit prefix — the
defeat showed up only as a bill (`cognition-audit-findings.md` H18).

## The determined fix

Three commits, in dependency order — meet the model, then rewrite the teaching,
then stop lying about the cache.

**1. Meet the model's prior where it's cheap (`0214a0b`, 2026-06-09).** The lexer
now accepts `&&`/`||`/prefix-`!` as aliases for `and`/`or`/`not` — "Postel's law
for an LLM-authored language — C-family training emits them constantly and a
single `&&` previously killed the whole routine at parse time" (commit message).
Canonical spelling stays `and`/`or`/`not`; single `&`/`|` remain errors; gapless
keyword-bang (`wait!`) still rejects. Pinned by `dsl/parser/alias_ops_test.go`.
The cheaper alternative (purge `&&` from the docs and teach keywords) was
explicitly rejected as weaker: the model's prior is overwhelmingly `&&`, and the
prior always wins eventually (`dsl-gap-analysis.md:61`).

**2. Ground-up manual rewrite, machine-validated (`b42a52b`, 2026-06-10).** Not a
patch — a full-surface inventory first (every construct/verb/view: taught-status ×
used-live status, the tables in `dsl-manual-analysis-2026-06-10.md`), then a
rewrite whose every claim was verified at file:line and whose **17/17 prose code
blocks parse+validate against the real parser+validator**, plus 5 inline coverage
files exercising every taught name and negative tests proving each prohibition is
real (the validation harness section of the analysis). The structural choices all
derive from the evidence:

- **Worked examples are the highest-leverage real estate** (prose presence
  predicts usage), so every task cluster got exactly one validate-clean idiom in a
  PLAYBOOK, replicating the `search_map` treatment — the one thoroughly-taught
  read surface and, not coincidentally, the most-adopted.
- **FILE SHAPE shown as a complete validated file first**, because top-level-only
  handlers is unlearnable from prose alone — `on` inside a routine body is a parse
  error, the root cause of zero live handler usage.
- **The f-string idiom canonized from what the retry loops converged on:** assign
  to a local first, then interpolate. The model had already discovered the fix
  across ~10 retries; the manual's job was to make the discovered idiom the
  starting point, not the destination.
- **`note()` rerouted:** narration only. State that needs read-back goes through
  `remember`/`recollect`, taught as a pair with a cross-routine worked example —
  the biggest teach-to-fix win, attacking two documented live failures
  (re-authoring identical failed routines, re-checking spent shops).
- **Honesty sections:** ENGINE ALREADY HANDLES (auto-doors, fatigue/HP reflexes —
  so routines stop reimplementing them badly), ANTI-PATTERNS (every documented
  live failure with its replacement), and the fire-and-forget note — a successful
  return means the click was *sent*; verify skilling outcomes via
  `world.messages.contains(...)`, branch `r.err.code` only where codes genuinely
  exist (travel, item-missing). The half-truth, retired by saying the truth.
- **Companion truth-fixes shipped in the same commit**, because the docs *around*
  the manual teach too: the validator message now suggests `timeout 30` (plain
  seconds), the `equipment_changed` docstring no longer claims bare handlers work,
  and phantom `host.name` is marked NotYetImplemented pointing at the live
  `self.name`.

**3. The cache as a contract, not a vibe.** Two sides:

- **The manual is a deploy artifact.** It ships as one `Cache:true` system block
  (`mesa/mesad/act.go:31`); its const carries the warning "Keep it stable — every
  edit invalidates the cache" (`mesa/mesad/dslmanual.go:8`). The rewrite replaced
  the prose *wholesale in one commit* and priced it explicitly as a one-time
  prompt-cache invalidation — one cost event, not a drip of invalidations from
  incremental patches (`b42a52b` commit message; `dsl-manual-analysis`
  integration notes).
- **H18 resolved by honesty, not by pretending (`03018a1`, 2026-06-08):** the
  below-minimum cron prefixes dropped `Cache:true` entirely and are priced as full
  input, with the mechanism documented at the call sites — "the analyzer prefix is
  ~390 tokens, FAR below Haiku's 4096-token minimum … a cache_control breakpoint
  here is SILENTLY ignored by Anthropic (never written, never read —
  cache_creation/read stay 0)" (`mesa/mesad/cron.go:448-450,506`,
  `mesa/mesad/cron_insight.go:571-572,640-641`) and pinned by a test so a future
  refactor can't silently re-add the pretend flag
  (`mesa/mesad/cron_insight_test.go:652`). The alternative (pad the prefix past
  the minimum with the shared manual, Act-style) remains available when the cost
  math favors it; what's forbidden is the flag that does nothing.

The remaining structural guard is the spec architecture: the appendix is
*generated* from `dsl/spec` and consistency-tested (`dsl/spec/consistency_test.go`),
so the reference half cannot drift from the engine; the analysis recommends
graduating the prose-block extraction harness into a unit test beside it so the
*prose* half can't drift either (harvested to `docs/TODO.md`).

## The durable rules

1. **Prose presence predicts usage.** The LLM uses exactly and only what worked
   examples model; an appendix row is availability, not teaching. Spend your token
   budget on validated worked examples, one per task cluster.
2. **Every example in an LLM-facing doc must be machine-validated.** Six reviewers
   read `&&` and passed it; the parser failed it instantly. Reviews read examples
   as prose — only a harness reads them as code.
3. **An error message is documentation — validate its suggested syntax like any
   other example.** It is read at the moment of maximum model attention.
4. **Meet the model's prior where it's cheap; canonize the idioms it converges on
   under failure.** Accepting `&&` removed the highest-frequency parse trap with
   one lexer change; the assign-first f-string idiom was the model's own
   discovered fix, promoted to teaching.
5. **Half-truths teach assume-success.** If some verbs are fire-and-forget, say so
   and teach the verification idiom — or the model loops on silent no-progress.
6. **Never describe a read-back workflow for a write-only tool.** "Note which
   shops you've tried" turned a journal into a black hole; route state through the
   store that has a read verb.
7. **A cache that can silently not engage needs a gauge.** Assert prefix length
   against the model minimum and surface `cache_creation/read_input_tokens` in
   telemetry — a silently-ignored `cache_control` shows up only on the bill.
8. **Treat the cached prefix as a deploy artifact.** Every edit is a fleet-wide
   cost event; batch changes into one stable text and price the invalidation
   deliberately.
9. **Generated reference or no reference.** A header that says "if it is NOT
   listed, it does not exist" is a promise only a generator plus a consistency
   test can keep.

## Sources

- `docs/_research/dsl-manual-analysis-2026-06-10.md` — the full-surface inventory
  (taught × used-live tables), the 69-routine usage tally, the f-string retry /
  assign-first finding, the `note()`-vs-`remember` routing, the 17/17 + negative-
  test validation record, the prompt-cache integration notes.
- `docs/_research/plutonium/dsl-gap-analysis.md` — the `&&` finding (§headline,
  item 1, raw note :216 "six prior reviewers' drafts missed this"), the
  spec-vs-runtime drift inventory, the lexer-alias recommendation (#1).
- `docs/_research/plutonium/dsl-synthesis.md` (:25,:50,:66,:143) and
  `dsl-crossref.md` (:420,:426) — the fire-and-forget correction and the
  "don't let the half-truth teach assume-success" resolution.
- `docs/cognition-audit-findings.md` — H18 (silent below-minimum `cache_control`),
  resolution header (all findings fixed 2026-06-08).
- Commits: `0214a0b` (2026-06-09, `&&`/`||`/`!` lexer aliases + docstring fixes),
  `b42a52b` (2026-06-10, ground-up manual rewrite + validator-message /
  `equipment_changed` / `host.name` truth-fixes), `03018a1` (2026-06-08, H18
  cache honesty), `937f707`/`aef6858` (2026-06-08, the incremental-patch era).
- Code: `mesa/mesad/dslmanual.go:8` (cache-stability contract),
  `mesa/mesad/act.go:31` (the one cached block), `dsl/validator/validator.go:332`
  (the corrected message), `dsl/parser/alias_ops_test.go`,
  `dsl/spec/manual.go:19` + `dsl/spec/consistency_test.go` (the completeness
  promise and its enforcement), `mesa/mesad/cron.go:448-450,506` +
  `mesa/mesad/cron_insight.go:571-572,640-641` + `cron_insight_test.go:652`
  (the H18 fix and its pin).
