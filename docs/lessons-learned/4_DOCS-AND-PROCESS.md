# Status decays in days: the documentation-drift saga

*Every hand-maintained truth claim in `docs/` — the status matrix, the phase
ladder, the API tags, the version ledger — went false within six to ten days of
being written, each by a different mechanism. One of them predicted its own
death in its own header and died on schedule anyway. This chapter is the
autopsy, and the 2026-06-10 docs-grounding tidy is the worked example of the
fix.*

> Chapter skeleton: **the problem as experienced → the false leads → the
> determined fix → the durable rules → sources.** Read this before you write,
> edit, or *trust* any status banner, task ID, or "NOT STARTED" claim in this
> repo.

---

## The problem as experienced

This project moves fast — a subsystem goes from empty directory to live code in
under a week — and it is built almost entirely by LLM sessions that orient by
reading `docs/`. That combination turned stale status claims from a cosmetic
problem into an operational one: a session that reads "memory: EMPTY" designs a
second memory system; a session that reads "Pearl was REJECTED" deletes the
policy engine's reason to exist. By the time the docs-grounding audit ran
(2026-06-10, 219 docs, per-doc verdict records — `docs/_research/docs-audit/`),
the falsehoods had piled up in five distinct shapes.

**Shape 1: the status matrix froze in ten days, with its own convention
watching.** Commit `57f263f` (2026-05-31, literally titled "docs: audit for
accuracy") gave `docs/index.md` a subsystem status matrix — brain **STUB**,
memory **EMPTY**, persona **PLANNED**, mesa **EMPTY/PLANNED**, observability
**EMPTY/PLANNED** — and, sixty lines below it, an accuracy convention in bold:
*"If you implement something that a doc marks STUB/EMPTY/PLANNED, update that
banner + this matrix in the same change."* Then the build happened: `memory/`
landed 2026-06-06 (`3dd0371`, tiered Manager), mesa LTM + genesis 2026-06-07
(`0ef66e7`), persona/pearl alongside. Nobody updated the matrix — not out of
defiance, but because the engineers shipping `feat(memory)` commits were not
*in* `docs/index.md` when they shipped. By the audit, exactly ten days after
the matrix was written, **every one of those five rows was false**
(audit verdict: `docs/_research/docs-audit/inventory.md`, index.md row —
"structure G, matrix stale"). The convention was correct, prominent, verbatim
— and enforced by nothing.

**Shape 2: "NOT STARTED, zero commits" — falsifiable by one grep.** The
master phase ladder asserted, in two places:

> *"Phase 2.6 (knowledge ingestion) was never started."*
> *"### Phase 2.6 — Knowledge ingestion (RAG corpus) — ⏳ NOT STARTED …
> never begun (zero commits as of HEAD `fd0731c`)."*
> — `docs/phases.md` at HEAD `d42e7ce` (lines 94, 181–185, pre-tidy)

That text entered the doc on 2026-05-31 in commit `cce77b2` — a commit titled,
with no irony available to it at the time, **"git-grounded roadmap
refactor."** Three days *earlier*, commit `ccbc220` (2026-05-28) had shipped
with the title **"feat: Phase 2.5 close-out + Phase 2.6 Slice 1 + live-test
catalog"** and a body reading *"Phase 2.6 Slice 1 (rsc.wiki corpus) — shipped.
New package `cognition/corpus` …"*. A single
`git log --oneline --grep="2.6"` falsifies the claim in under a second. Worse:
the doc *pinned a HEAD* (`fd0731c`, 2026-05-30) — which post-dates `ccbc220` —
so the claim was false at the very commit it cited as evidence. The pin gave
the assertion the *costume* of verification with none of the substance. And
`phases.md` was edited again on 2026-06-07 (`607600b`) with the falsity
surviving untouched.

**Shape 3: api.md §8 drifted exactly as its own header predicted.** When the
Routine API was frozen (`38ef5a0`, 2026-05-29), the header made the failure
mode explicit: the exhaustive per-entry reference (§8) was designed to be
*"generated from `dsl/spec/*` … by a `go run ./cmd/specdoc`"* tool, because
*"hand-maintaining ~200 spells + ~90 skill accessors by hand would drift on
the first change."* The generator was never built. §8 was hand-maintained. The
body build-out began falsifying tags the same day the freeze landed (`3678044`
shop end-to-end, `ba8d043` fatigue→sleep — both 2026-05-29). By the audit, a
dozen `(to build)` tags were false — e.g. `combat.engaged` tagged *"stub,
always returns false … (to build — perception gap)"* (pre-tidy api.md:602)
while `runtime/views_combat.go:32-33` returned the real wire-observed
engagement — and §8 contradicted *itself*: the combat notes called
`combat.target`/`combat.engaged` "stubs pending task #8" while the entries
above them said `(exists)` (resolution recorded at the rewritten
`docs/lang/api.md:648`). Note that "task #8": a bare task number from a
colliding namespace, which is Shape 5. The doc knew precisely how it would
die, wrote it down, and died that way.

**Shape 4: semver decayed in six days without a test gate.** The Routine
Runtime versioning policy shipped 2026-05-31 (`898b3fc`): semantic versioning,
mandatory `runtime "X.Y"` targeting, a CHANGELOG, and — *planned, not built* —
a spec surface-snapshot test that would fail when the builtin/event/accessor
tables changed without a bump (pre-tidy `docs/lang/versioning.md` §4:
"*Planned:* a `dsl/spec` surface-snapshot test…"). The first unbumped,
un-changelogged surface change landed **six days later** (`57ab761`,
2026-06-06). By 2026-06-10 the builtin table had grown 54 → 67 and the event
table 31 → 34 against an unmoved `RuntimeVersion = "1.0.0"`, including a
*breaking* `converse`/`go_to` rework (`3c0ab33`, 2026-06-08 — MAJOR material)
— with the CHANGELOG still showing exactly one entry. Even the policy doc's
own "224 committed `.routine` files" count had drifted to 294/299. The
*mechanism* (loader version checks) was code and held perfectly; the
*discipline* (bump + entry per change) was prose and held for six days. That
asymmetry is the whole lesson.

**Shape 5 (the quiet one): banners stale in the *opposite* direction.** Most
doc rot overclaims. Ours also *underclaimed*, which for an LLM-driven repo is
the more dangerous direction. `docs/memory.md` carried the best banner in the
tree — *"STATUS: EMPTY PACKAGE / DESIGN ONLY (verified 2026-05-31). The
`memory/` directory contains **no Go files**"* — dated, specific, true to the
letter when written, and falsified six days later by `3dd0371`. `reveries.md`'s
banner ("EMPTY PACKAGE / DESIGN ONLY, verified 2026-05-31") was wrong **in both
directions at once** by audit time: the empty `reveries/` directory no longer
existed, *and* the design's pieces had shipped under other names — the mood
model as `limbic.Affect`, the jitter seed as `persona.ReverieSeed`, injection
as pearl `EffectInject` (inventory.md, reveries.md row). A single
built/unbuilt bit cannot represent "manifested differently," so it lies no
matter which way you set it. The extreme case was
`docs/_research/SESSION-STATE.md`, the designated resume-here handoff doc:
"Nothing is built yet" and "Pearl was REJECTED" — both *inverted* by the build
(inventory.md, SESSION-STATE row: "PS bordering inverted"). A session resuming
from it would have re-designed a shipped stack and treated a live subsystem as
rejected doctrine.

**And underneath all of it: the `#N` task-ID namespace collision.** Five
workstreams independently minted small task numbers: the canonical ladder
`#19`–`#123` (`docs/tasks.md`), the 2026-06 in-session tracker `#14`–`#36`
(the metacognition cluster `#27`/`#30a`/`#32`/`#33`), the plutonium movement
backlog `#1`–`#15`, the dsl-gap-analysis redesigns `#1`–`#11`, and the persona
worksheet `A/B/C/D-n` (ground rule now at `docs/tasks.md:13-31`). A bare
"#31" in a commit message meant three different things depending on the
reader's last-loaded context. The freeze-era in-session IDs `#96`–`#113` were
only partially back-filled into the ledger, and `docs/tasks.md:102-103`
carried outright **false checkmarks** — `#89`/`#90` ticked for surfaces
(`by_type`, `last_attacked_npc`) that never existed under those names (fixed
`46e16b7`, 2026-06-10).

## The false leads

**False lead #1: discipline as the enforcement mechanism.** The accuracy
convention failed in ten days; the semver policy failed in six; both were
written clearly by people who meant them. The pattern repeats too precisely to
blame the authors: a convention that requires a human (or an LLM session) to
*remember to look elsewhere* while shipping code loses to any session that
didn't read that doc this time. Discipline is real but it is a decay rate, not
a guarantee.

**False lead #2: the costume of verification.** Pinning "zero commits as of
HEAD `fd0731c`" *looks* git-grounded; so does a commit titled "git-grounded
roadmap refactor." Neither involved running the one-line grep that falsifies
the claim. Dates and hash pins make staleness *computable*; they do not make
the claim *true at the pin*. Verification is the grep, not the citation
format.

**False lead #3: predicting drift instead of preventing it.** api.md's header
is the purest specimen: a precise, correct forecast of the failure mode,
shipped *without the generator that was the entire mitigation*. Documenting a
failure mode and mitigating it feel similar at the keyboard. They are not.

**False lead #4: status as one global bit.** EMPTY/PLANNED/BUILT works for a
subsystem that ships as a unit. This project's designs ship *piecemeal under
other names* (reveries → limbic + persona + pearl fragments; memory.md's
design → `memory/` + `hostkv/` + `limbic/ledger` + mesa crons, all shaped
differently). A banner that can't say *where each fragment went* will be false
in both directions simultaneously, and the reader can't even tell which kind
of false.

**False lead #5: fixing status where you happen to be standing.** Status lived
in the banner, the index matrix, `state.md`, `phases.md`, and `tasks.md`
at once; each edit fixed the copy in front of the editor. Five copies of one
fact is four future lies (the same disease as the `westworld.conf`
copies-not-symlink story — see the config-drift chapter).

**False lead #6: letting every workstream count from #1.** Each tracker was
locally coherent; nobody decided to collide. Namespaces collide by default —
they have to be *partitioned* by an act of will, and until they were, the only
honest arbiter was the commit hash.

## The determined fix

The fix was the 2026-06-10 **docs-grounding audit + tidy** itself — this
chapter's own provenance, and the worked example of every rule below. Shape:
audit first (219 docs inventoried with per-doc verdicts;
`docs/_research/docs-audit/inventory.md`), the audit then **adversarially
re-checked** (26 verdicts re-litigated, 24 upheld, 2 downgraded — the audit
did not trust itself either; inventory.md "Recheck disposition"), then a
staged execution plan with operator STOP gates and measurable success criteria
stated as greps (`docs/_research/docs-audit/execution-plan.md`). The concrete
mechanics that came out of it:

1. **Banners carry a verify-date + commit hash, and falsity classes get
   named.** Every rewritten root doc now opens "verified against code
   2026-06-10, HEAD `0bfa818`" (e.g. `docs/index.md:3`, `docs/tasks.md:3`).
   Staleness becomes computable: a future reader diffs the pin against HEAD
   instead of arguing with prose.

2. **"NOT STARTED" claims must survive `git log --grep` before being
   asserted.** The Phase-2.6 falsity was deleted in the phases.md rewrite;
   `tasks.md` was recast as an explicitly **closed historical ledger** so it
   can no longer pretend to be a frontier (`docs/tasks.md:1-11`).

3. **One TODO SSOT with collision-free, namespaced IDs — and the commit hash
   wins.** `docs/TODO.md` absorbed every open item from every colliding
   tracker, tagging each with its source namespace and a collision-free ID
   (`MP-n`, `DSL-n`, `C-n`, `P-n`, …). The ground rule is written where the
   collisions lived: *"when an in-session tracker and this doc disagree on a
   number, the commit hash wins"* (`docs/tasks.md:18-21`, restated at the
   ID-numbering note, `docs/tasks.md:212-217`).

4. **Generated reference or no reference.** §8 was re-verified by hand exactly
   once — as part of the audit, with the labor priced in — and the durable fix
   is filed as **DSL-17**: build `cmd/specdoc` and regenerate §8 from
   `dsl/spec/*` "so it can never drift" (`docs/TODO.md:63`). The rewritten
   api.md header now tells the prophecy-came-true story in place, so no future
   maintainer re-trusts the hand tags (`docs/lang/api.md:22-29`).

5. **Version arrears acknowledged in-band, with the test gate as the real
   fix.** The CHANGELOG got a backfilled **"Unversioned arrears"** section —
   every unbumped change listed with the classification it *should* have
   received, newest first (`docs/lang/CHANGELOG.md:21-33`) — and both
   versioning docs wear an IN-ARREARS banner instead of pretending. The bump
   plus the **surface-snapshot test that fails on unbumped change** is
   **DSL-18** (`docs/TODO.md:64`): the policy doc planned that test in §4,
   didn't build it, and decayed in six days; it is no longer optional.

6. **Banners must name the absorbing packages.** The rewritten `reveries.md`
   replaces the dead EMPTY bit with a **"What already manifested, where"
   table** — each design fragment mapped to the code that absorbed it
   (`limbic.Affect` / `runtime/limbic.go` / `persona` `Trajectory.Mood`, …)
   with every pointer verified at HEAD (`docs/reveries.md:22-33`). The old
   `memory.md` design was archived **with a same-commit fresh replacement** at
   the same path describing the real tiers (`ff72002`, 2026-06-10 — the audit
   made the replacement *mandatory in the same commit* so the 7+ inbound links
   never 404'd; inventory.md, memory.md row). The SESSION-STATE handoffs get
   SUPERSEDED banners pointing at the live successor (execution-plan item 45).

7. **The OUTCOME-banner-on-spec pattern.** When a spec gets executed, the spec
   doesn't get deleted or silently rot — it gets an **OUTCOME banner**
   prepended, recording what actually shipped and *every deviation, with who
   decided it*. The exemplar is `docs/_research/runhost-extraction-spec.md:8`:
   *"**OUTCOME (shipped):** … **Two design changes vs. this spec, decided with
   the user:** (1) no mesa connection pooling … (2) the mesa address is
   per-host …"*. The spec stays honest history instead of becoming aspirational
   fiction; the deviations are exactly the part a future reader needs and
   exactly the part that otherwise evaporates. The tidy institutionalized the
   pattern for itself: execution-plan Stage 6 closes by appending a completion
   record of what shipped per stage, hashes, and deviations
   (`execution-plan.md`, Stage 6 item 4).

8. **Impl specs re-pin to the current tree before execution.** An
   implementation spec is a bag of `file:line` citations, and line numbers rot
   faster than any prose. The Phase-5b-1 spec opens with a **"TREE / CITATION
   CORRECTION (read first)"** block: the proposed spec had pinned everything
   to `00a63fa`, the working tree had moved to `c8e3bb3`, and *"all line
   numbers below are re-pinned against `c8e3bb3`. Where a number moved, I note
   it"* — plus a build/test-green baseline at the new pin
   (`docs/_research/phase-5b-1-impl.md:5-13`). And even that wasn't enough:
   the 5b-2/3 execution record notes the `perceiveShop` line anchors were off
   *again* at implementation time and re-pins them once more
   (`docs/_research/phase-5b-23-impl.md:765`). Re-pinning is not a one-time
   courtesy; it is the first step of executing any spec.

## The durable rules

1. **A status assertion without an enforcement mechanism is a future lie —
   generate it, gate it with a test, or date-pin it and expect decay.** The
   observed half-lives here: ten days (index matrix), six days (semver), six
   days (memory banner).
2. **Never write "NOT STARTED" without running `git log --grep` first** — the
   commit titles are the falsifier, and one of ours contained the phase number
   verbatim.
3. **A hash pin proves when you looked, not that you looked.** "Zero commits as
   of HEAD X" was false at HEAD X. Run the check; don't wear its costume.
4. **Every status banner carries a verify-date and a commit hash** so
   staleness is computable instead of arguable.
5. **When a design ships piecemeal, the banner must name the absorbing
   packages** — a single built/unbuilt bit on a fragmented design is wrong in
   both directions at once.
6. **Stale-toward-unbuilt is the dangerous direction in an LLM-driven repo:**
   a reader that believes "EMPTY" re-designs what exists. Audit underclaims as
   aggressively as overclaims.
7. **Generated reference or no reference** — hand-maintained per-entry tags
   drift on the first change, exactly as api.md's own header predicted.
8. **A version policy without a snapshot test that fails on unbumped surface
   change is a polite request**, and it decays within days.
9. **One task tracker, collision-free namespaced IDs; when trackers disagree,
   the commit hash wins.**
10. **Stamp executed specs with an OUTCOME banner recording the deviations** —
    a spec without its outcome becomes fiction the day the build diverges.
11. **Re-pin an impl spec's citations to the current tree before executing
    it** — and expect to re-pin again at the keyboard.
12. **Documenting a failure mode is not mitigating it.** If the header can
    predict the drift, the same commit can prevent it. Ship the generator.

## Sources

- `docs/_research/docs-audit/inventory.md` — the 219-doc audit: per-doc
  verdicts (index.md, phases.md, api.md, versioning/CHANGELOG, memory.md,
  reveries.md, tasks.md, SESSION-STATE rows) + the adversarial-recheck
  disposition.
- `docs/_research/docs-audit/execution-plan.md` — the staged tidy plan, STOP
  gates, grep-stated success criteria, Stage-4 chapter brief, Stage-6 OUTCOME
  record.
- `docs/_research/docs-audit/lessons-learned-drafts.md` — chapter outlines
  this file expands.
- Commits: `ccbc220` (2026-05-28, "Phase 2.6 Slice 1" — the falsifier),
  `cce77b2` (2026-05-31, the "git-grounded" refactor that asserted NOT
  STARTED), `57f263f` (2026-05-31, matrix + accuracy convention + the verified
  banners), `38ef5a0` (2026-05-29, api.md freeze + the prophecy), `898b3fc`
  (2026-05-31, semver policy), `57ab761` (2026-06-06, first unbumped surface
  change), `3c0ab33` (2026-06-08, unbumped breaking rework), `3dd0371`
  (2026-06-06, `memory/` is born — banner falsified), `607600b` (2026-06-07,
  phases.md edited, falsity survives), `46e16b7` (2026-06-10, false #89/#90
  checkmarks fixed), `ff72002` (2026-06-10, memory.md archive + same-commit
  replacement).
- Live artifacts of the fix: `docs/index.md:3` (pinned banner),
  `docs/tasks.md:13-31` (namespace ground rule + commit-hash-wins),
  `docs/TODO.md:63-64` (DSL-17 specdoc generator, DSL-18 bump + snapshot
  test), `docs/lang/CHANGELOG.md:22-33` (the arrears section),
  `docs/lang/versioning.md:1-15` (IN-ARREARS banner), `docs/reveries.md:22-33`
  (the manifestation map), `docs/lang/api.md:22-29,648` (prophecy +
  contradiction resolution).
- Pattern exemplars: `docs/_research/runhost-extraction-spec.md:8` (OUTCOME
  banner), `docs/_research/phase-5b-1-impl.md:5-13` +
  `docs/_research/phase-5b-23-impl.md:765` (re-pin to the current tree).
