# Development workflow — the git & ship contract

> Two developers (Alex + Claude), one deploy target (this machine), one remote
> (GitHub, backup + history — not a gate). Agreed 2026-06-11.

## The loop

```
main (local) ←→ origin/main (GitHub)      the single shippable truth
  └── one short-lived feature branch per batch
        └── many small UNSIGNED commits    the archaeology log
```

1. **Branch per batch** — cut from current main. Name it for the batch
   (`fix/mesad-oom-incident`, `feat/cradle-ui-v2`).
2. **Commit early, often, unsigned.** `commit.gpgsign false` is set
   repo-locally, so this is automatic. Small commits with what+why messages
   are the record of how we accomplished things — never squash. Committing
   must never block on a human: signing happens only at ship time.
3. **Ship** (requires Alex present for one Touch ID):
   - bulk-sign the batch: `git rebase --force-rebase --gpg-sign main`
   - `git checkout main && git merge --no-ff -S <branch>` — the merge commit
     marks the batch boundary; `git log --first-parent main` is the release log
   - `git push origin main` — **the only push**; the merge carries every
     branch commit to GitHub, so branch refs are never pushed (exception: a
     branch living for days may be pushed as an off-machine checkpoint, then
     its remote ref deleted after merge)
   - delete the local branch
4. **Deploy = `scripts/ship.sh`** — the only way binaries reach the fleet.
   It refuses a dirty tree, refuses a main that isn't a fast-forward of
   origin/main, gates (build + vet + full tests), builds ALL binaries from
   one SHA with `-X main.build=<sha>` stamps, swaps atomically, relaunches,
   and records the SHA in `/tmp/ww-bin/DEPLOYED`. "What's running" is the
   `build=` field in each daemon's startup log; "is it latest" is comparing
   that to `origin/main`.

## Why these specific choices

- **Unsigned commits + bulk-sign at ship**: 1Password requires Touch ID and
  auto-locks; per-commit signing blocks autonomous work on human presence.
  Signing the backlog in one rebase right before merge gives the same signed
  history on GitHub with one human moment per batch.
- **`--no-ff` merges**: batch boundaries survive in history; detail commits
  sit one level below.
- **No branch pushes**: GitHub holds main; merge commits carry the full
  graph. Less remote state to manage, identical history.
- **Deploys only from main via ship.sh**: the fleet can never run a mix of
  build vintages or a branch that didn't merge — both have bitten us.

## Secrets

`./.local.env` (gitignored) holds launch secrets, sourced by ship.sh.
Never print or commit it. Commit hygiene: grep staged diffs for secrets
before shipping.
