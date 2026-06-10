# 3 — Config and environment drift: the symlink that was never made (twice)

> **The rules, up front.** Copies drift — symlink or generate. A fix that exists
> only in a doc is **not applied** — verify the artifact, not the instruction.
> Every "current posture" claim carries a date and a source-of-truth pointer.
> And when the environment offers you an obvious state flag, check whether its
> *absence* is overloaded before keying logic on it — for tutorial state,
> position is the discriminator, not the cache key.

This chapter is about a one-line fix that was decided in May 2026, written into
three documents, re-affirmed in an agent charter — and, as of 2026-06-10, still
does not exist on disk. In between, the same pair of files drifted twice,
exactly as the never-executed fix predicted.

## The problem as experienced

westworld's server config has two homes. The repo copy,
`inc/westworld.conf`, is the version-controlled "source-of-truth backup"
(committed at bootstrap on 2026-05-27, `75beb49`). The deployed copy,
`~/Code/openrsc/server/westworld.conf`, is the one the server actually reads —
`ant runserver -DconfFile=westworld` resolves the conf by name from the server
directory (`docs/server-config.md`, "Where it lives, where it runs"). Two
files, one meaning, no mechanism keeping them equal.

**Drift #1 (discovered 2026-05-31).** The bootstrap-era repo conf said F2P on
port 43596 (`75beb49:inc/westworld.conf:29,45` — `server_port: 43596`,
`member_world: false`). Meanwhile the live server had been flipped to
P2P / members world on port 43594, because F2P-only restrictions were limiting
what the live-test scenario catalog could exercise (`docs/server-config.md:17-24`;
the original F2P decision and its reversal are both recorded in
`docs/questions-and-decisions.md:333-345`, "F2P out of gate — **REVERSED**").
Only the deployed copy was edited. The repo copy — the supposed source of
truth — silently became fiction.

The cost wasn't the conf line itself; it was the **fossil field** the stale
line left behind. Docs written against the repo copy taught F2P/43596 as
current reality. The incoming server-steward agent's very first chartered task
had to include a warning that "`docs/server-config.md` still describes the old
F2P / port-43596 posture; the live server is P2P / 43594"
(`cce77b2:docs/agents/openrsc-steward.md:127-128`). Code defaults fossilized too:
`cmd/legacy-cradle/main.go:81` *still* defaults to `localhost:43596`, and every
runner script has to override it (`docs/scenarios.md:371`;
`docs/server-config.md`, "Ports"). One drifted file radiated stale claims into
docs and binaries for weeks.

The 2026-05-31 handoff commit (`cce77b2`) cleaned this up: repo conf re-synced
from deployed, `docs/server-config.md` rewritten for P2P/43594, and — crucially
— a drift warning plus the correct fix added:

```bash
ln -sf ~/Code/westworld/inc/westworld.conf ~/Code/openrsc/server/westworld.conf
```

**Drift #2 (2026-06-09, nine days later).** The deployed conf was edited again
— `want_runecraft: false`, comment: "MATCH uranium/preservation — runecraft is
a custom (non-authentic) skill" — while the repo copy still says
`want_runecraft: true` (`inc/westworld.conf:222`). Verified during the
2026-06-10 docs audit and re-verified live while writing this chapter:
`~/Code/openrsc/server/westworld.conf` is still a plain file (`-rw-r--r--`,
mtime Jun 9 02:10), not a symlink, and `diff` against the repo copy shows
exactly that one line differing (`docs/server-config.md:42-61`, the
twice-observed drift warning).

## The false leads

1. **Copies plus discipline as the sync mechanism.** The implicit model was
   "we'll remember to edit both." Nine days separated the documented fix from
   the next drift. Discipline is not a mechanism; it's a countdown.

2. **Treating *documented* as *done*.** This is the heart of the chapter. The
   2026-05-31 fix was written into `docs/server-config.md` (drift warning +
   symlink command), into the steward charter as a first task, and later into
   `docs/TODO.md` — and never executed. Note the original charter wording:
   "Propose (don't yet apply, unless Alex says go) establishing the conf
   symlink" (`cce77b2:docs/agents/openrsc-steward.md:130-131`). The fix
   shipped with a hold on it, the hold was never lifted, and nobody noticed
   because the *documentation* of the fix made the problem feel handled. Three
   documents described the symlink; zero shells ran it.

3. **The drift warning as the fix.** A warning box documents the disease. It
   cures nothing. The warning was present, accurate, and totally ineffective
   at preventing drift #2.

4. **Companion-doc rot riding along.** While the conf drifted, the doc
   describing it rotted independently: `docs/server-config.md` kept describing
   a SQLite database the server stopped using on 2026-05-29 (the MySQL
   cutover; `docs/agents/openrsc-steward.md:95-96`,
   `docs/tutorial-host-snapshot.md:7-10`). Fixed only in the 2026-06-10 audit
   rewrite (`docs/server-config.md`, "Database"). Drift clusters: the moment
   one "current state" claim goes stale, audit its neighbors.

## The determined fix

The fix has not changed since 2026-05-31; what changed is the insistence that
it be **applied, not re-documented**:

- **One source of truth.** Establish the symlink so the repo file is the only
  file: `ln -sf ~/Code/westworld/inc/westworld.conf
  ~/Code/openrsc/server/westworld.conf`. After that, edit only the repo file;
  the deployed copy follows. Tracked as **O-12** (`docs/TODO.md:147`) and the
  steward's standing first task, now with an execution checklist instead of a
  "propose" hold: reconcile, symlink, confirm the server still starts on
  43594, report (`docs/agents/openrsc-steward.md:149-166`).
- **Reconciliation is a decision, not a merge.** Because the drift lived,
  both sides now carry legitimate intent (repo: runecraft for the live-test
  catalog; deployed: match uranium/preservation authenticity). Which side
  wins is an explicit open call routed to Alex (`docs/server-config.md`,
  "Members (P2P) content posture" table; O-12). The longer a drift survives,
  the more it converts from a sync bug into a policy question.
- **Interim rule until the symlink exists:** change the **repo** file so the
  change is version-controlled, re-copy to the deployed path, restart
  (`docs/agents/openrsc-steward.md:80-88`).
- **Self-dating posture claims.** The rewritten doc carries a dated STATUS
  banner ("verified against code 2026-06-10 ... re-checked against
  `inc/westworld.conf`, the deployed conf, and `preservation.conf`") and a
  history note that retro-dates fossils elsewhere: "If you read references to
  'F2P enforcement / port 43596' elsewhere, they predate this change"
  (`docs/server-config.md:25-28`). Stale references now identify themselves.
- **Audit the artifact, not the instruction.** The 2026-06-09 re-drift was
  caught by `ls -la` (is it a symlink?) and `diff` (what differs?), not by
  reading any doc. That is the verification procedure; the docs are just where
  its results get recorded.

## The tutorial cache-key ambiguity

A smaller environment-drift cousin, recorded while capturing the fresh-host DB
snapshot from `stubbs2` (player id 227) on 2026-06-06
(`docs/tutorial-host-snapshot.md`).

The obvious way to ask "has this host done the tutorial?" is the server's own
flag: `player.getCache().getInt("tutorial")`. It lies by omission. The cache
key is created as the host progresses and **removed by the boatman on
completion**, and `getInt` defaults to 0 when the key is absent — so "never
started" and "completed" are **indistinguishable in `player_cache`**: no key
either way (`docs/tutorial-host-snapshot.md:15-19`).

**Position is the real discriminator.** A fresh host sits at `(216, 744)`, the
tutorial-island start room; a graduate is at the Lumbridge spawn,
`~(120, 648)` (`docs/tutorial-host-snapshot.md:13-14`). Anything minting
tutorial hosts by SQL clone must honor both halves: leave `player_cache`
without a `tutorial` key *and* set position `(216,744)`
(`docs/tutorial-host-snapshot.md:69-72`). Key detection logic on the flag
alone and you will confidently misclassify hosts at both ends of the pipeline.

The general shape: **absence is overloaded**. A missing key meant two opposite
things, and only one observable (position) was actually monotonic with the
state we cared about. Before keying logic on any environment flag, verify it
against *both* end states.

## The durable rules

1. **Copies drift; symlink or generate.** Two files with one meaning and no
   mechanism is a scheduled outage. Nine days, in our case.
2. **A fix that exists only in a doc is not applied.** "Chartered" is the most
   dangerous status a fix can have — it feels done. Verify the artifact
   (`ls -la`, `diff`), never the instruction. If a fix ships with a hold
   ("propose, don't apply"), track the hold as the open item.
3. **Every "current posture" claim needs a date and a source-of-truth
   pointer**, and posture flips leave a history note that retro-dates the
   fossils ("references to X elsewhere predate this change").
4. **Reconcile drifts as decisions.** Once both copies have accumulated
   intent, name who decides which side is right; don't silently pick one.
5. **Drift clusters.** One stale "current state" claim (F2P/43596) predicts
   stale neighbors (the SQLite section). Audit the blast radius, not the line.
6. **Absence is overloaded.** Never key state detection on a flag that is
   absent in two different states — find the discriminator (here: position)
   and verify it against both end states before trusting it.
