# Remote Client — Build Log (M1)

**Status:** M1 shipped — builds, vets, tests green.
**Branch:** `feat/remote-client`
**Scope of this log:** what M1 built, how to run it, the adversarial-review
fixes applied in the Finish phase, and the final gate results.

This is the factual record that closes the M1 build. The frozen cross-layer
contract is `50-impl-spec.md`; the north star is `00-overview.md`.

---

## 1. What M1 shipped — the four layers

A full-fledged remote RuneScape Classic client driven entirely from a browser:
right-click anything in the world, get the authentic RSC context menu, click an
option, and the bot does it — plus inventory, equipment, and chat. No native
window, no CGo. The server is `cradle`, already logged into a real RSC server as
a `runtime.Host`; it streams rendered PNG frames and exposes a JSON API.

### Layer 1 — screen-space hit-testing (`render/`)

- `render/hittest.go` — shared billboard projection helpers (`billboardCamera`,
  `projectBillboard`) extracted VERBATIM from `DrawEntitySprites`, with the
  `buildScene` camera defaults baked in (Zoom 750 / W 512 / H 334).
- `render/render.go` — `DrawEntitySprites` refactored to call those helpers;
  the rendered framebuffer is byte-identical (guarded by `render/fidelity_*_test.go`).
- `render/pick.go` — `render.Pick(land, f, v, px, py) []PickCandidate`: billboards
  (NPC / player / self) projected nearest-camera-first, then tile-grounded targets
  (dynamic scenery → static scenery → boundary → terrain) via `PickTile` +
  `facts.At`. `render.TargetKind` int enum + `PickCandidate`. Returns plane-LOCAL X/Y.
- `render/hittest_test.go` — projection round-trip, AABB math/containment,
  clip/zero-size rejects, ordering, identity completeness, renderer-parity golden.

### Layer 2 — menu + dispatch (`remoteclient/`)

- `remoteclient/target.go` — the wire types (`MenuTarget`, `MenuOption`,
  `Candidate`, `PickResponse`), the string `TargetKind`+`Kind*` consts, the
  internal `OptionID`+`Opt*` dispatch keys, and `kindToWire`.
- `remoteclient/menu.go` — `BuildMenu(kind, defs) (opts, ids)` (the pure,
  deterministic per-kind verb-list builder), `BuildCandidates` (pick → wire
  Candidate, folds plane → absolute Y), `InventoryMenu` (same builder for /state
  + the /pick inventory fallback).
- `remoteclient/dispatch.go` — `ActionHost`/`ExamineHost`/`WorldView` interfaces,
  `Dispatcher`, `Dispatch(ctx, t, optionId) (msg, lane, err)`, and (added this
  phase) the pure `ResolveLane(t, optionId) (lane, err)`. Maps `(kind, OptionID)`
  → the exact `runtime.Host` method per the §4 table (NPC verb-route, 1-based
  `InteractAt`, 0-based slots, boundary same-packet, examine = no packet).
- `remoteclient/menu_test.go`, `remoteclient/dispatch_test.go` — table-driven menu
  ordering + dispatch routing, stale-target/unknown-option rejection, and the
  `ResolveLane` purity + validation suites.

### Layer 3 — HTTP handlers (`cmd/cradle/`)

- `cmd/cradle/remoteclient.go` — `serveClient(...)`: mirrors `spectate`'s
  bootstrap, re-registers `/frame /pos /walk /shot /clip` verbatim, and adds
  `POST /pick`, `POST /act`, `GET /state`, `POST /chat`, `GET|POST /examine`.
  Owns the single `actCh` worker (walk = coalescing retarget; interactions =
  serialized single-action lane with a 1.5s soft timeout), the Follow goroutine
  (own cancellable ctx), and the server-session chat ring.
- `cmd/cradle/main.go` — `cfg.client`/`cfg.clientAddr` config fields + the
  `-client` / `-client-addr` flags + the post-`spectate` invocation block.
- `cmd/cradle/spectate.go` — **NOT modified** (the `-spectate` mode is untouched;
  `serveClient` reuses its package-main helpers `buildLiveView`,
  `newMotionTracker`, `waitForLivePosition`, `montage`).

### Layer 4 — browser SPA (`cmd/cradle/clientpage.go`)

- `const clientPage` — the raw-string HTML/CSS/JS SPA: viewport (left-click
  default-act, right-click `#menu` from `/pick`), sparse 30-cell inventory grid
  (left = default option, right = `options[]` from `/state`), equipment summary,
  chat box (`::`/`/` → command, `@name` → pm, else → say; appends `seq > lastSeq`).
  Pure page asset — no `net/http`, no Go logic.

---

## 2. How to run it

```
cradle -client -client-addr localhost:8090 \
       -username <bot> -facts <OpenRSC-root> [other login flags as for -spectate]
```

`-client` boots `serveClient` after login; open `http://localhost:8090/` in a
browser. The `-facts` root must point at the OpenRSC content tree (it supplies
the landscape + `models.orsc` + the facts defs). `-client-addr` defaults to
`localhost:8090`. The legacy `-spectate` mode is unchanged and still serves its
own page on its own address.

---

## 3. Fixes applied this phase (post-review Finish)

The menu/dispatch package was audited by a prior agent; this phase applied the
hit-testing-math + HTTP-handler review findings. Confirmed-by-reading fixes:

### CRITICAL — `/act` double-dispatch (the mandated fix)

`POST /act` called `disp.Dispatch(context.Background(), ...)` "just to resolve
the lane", but `Dispatch` EXECUTES the Host call while returning the lane. The
worker then re-called `Dispatch`, so every non-examine action fired TWICE — the
first on an unbounded `context.Background()` OUTSIDE the serialized worker (and
for Follow, the pre-call BLOCKED the HTTP goroutine for up to the 10s startup
timeout on an un-cancellable ctx).

Fix: added a pure, side-effect-free `(*Dispatcher).ResolveLane(t, optionId)
(Lane, error)` to `remoteclient/dispatch.go`. It re-derives `ids[optionId]` →
`OptionID` → `Lane` and re-validates the option range (`ErrUnknownOption`) and
volatile identity (`requireNpcVisible`/`requirePlayerVisible`/`requireSlotItem`
→ `ErrStaleTarget`) with NO Host call. The `/act` handler now routes via
`ResolveLane`, then calls `Dispatch` EXACTLY ONCE inside the chosen lane:

- LaneSync (examine / reserved Use): one synchronous `Dispatch` on `r.Context()`
  (no packet).
- LaneFollow: `startFollow` only (the single Follow runs in the follow goroutine
  on its own cancellable ctx); `Dispatch` is NOT called.
- LaneWalk / LaneAction: `Dispatch` runs exactly once inside `enqueueWalk` /
  `enqueueAction` with the worker ctx.

`ErrUnknownOption` → 400 and `ErrStaleTarget` → `{ok:false}` are preserved.
A new `TestResolveLanePure` (a mock `ActionHost` asserted to record zero calls +
a `failExamine` `ExamineHost` that `t.Fatal`s on any call) proves `ResolveLane`
fires no packet across every kind/option; `TestResolveLaneValidation` proves it
still reports `ErrUnknownOption`/`ErrStaleTarget` without side effects.

### MAJOR — `enqueueAction` detached-goroutine busy fallback

When `actCh` was full, the handler ran the action on a fire-and-forget goroutine
on a fresh 30s ctx OUTSIDE the single worker, defeating serialization (two
interaction packets could be in flight at once). Fix: removed the detached
goroutine. The action lane now drain-then-sends like walk — a new interaction
supersedes the queued-but-unstarted one (RSC "latest action wins") and EVERY
real packet stays on the single `actCh` worker.

### MAJOR — `chatRing.ingestRecent` server-message index drift

Server messages were de-duped by an absolute slice index (`lastSysIdx`) into
`world.Recent.ServerMessages()`, which is a BOUNDED ring (`ServerMsgRingCap=32`)
that drops its oldest entries and copies the tail down once full — so indices
shift and, after >32 messages, new system/NPC-dialog lines would stop reaching
`/state`. Fix: dedupe by timestamp (`lastSysMsgAt time.Time`, append only
`m.At.After(lastSysMsgAt)`), the same approach `Chat()`/`PM()` already use.
Index-shift-immune.

### MINOR (fixed because trivial / frozen-contract drift)

- `enqueueWalk` cannibalizing a queued interaction: the shared `drainPending`
  helper now signals a drained interaction's `done` channel with a "superseded"
  result, so a click-to-walk that supersedes a queued Attack no longer leaves
  that handler hanging until its soft timeout reporting a false "(running)".
- `pickRequest.Slot` (frozen spec §6.1 `PickRequest{...; Slot int}`): added the
  `Slot int` field, defaulted to -1 server-side, and wired the inventory-slot
  fallback branch in `/pick` (builds the candidate from `world.Inventory` +
  `InventoryMenu` when `Slot>=0`). The browser `pick()` helper now sends
  `slot:-1` on screen picks so omitted-key clients still take the screen path.
- `render/pick.go` doc comment corrected: with `land==nil` Pick returns an EMPTY
  list (billboard projection also needs the terrain grid + host elevation), not
  "billboards only".

---

## 4. Review outcome — findings deferred

These review findings were declined this phase (with reason); none is a
correctness bug in M1:

- **pick.go:211** — `PickTile` uses looser camera defaults (600/336) than the
  billboard path (750/334). Unreachable in production: the `/pick` handler always
  supplies W,H>0 and defaults Zoom from `cfg.renderZoom`. Latent inconsistency
  only; the 10-hit-testing doc accepts nearest-tile as robust. Left as-is.
- **pick.go:384** — dynamic-scenery `Direction` is always 0 because the live path
  never populates it. Pre-existing data-source limitation, not a Pick bug;
  scenery dispatch uses `InteractAt(x,y,option)`, not Direction (Dir is advisory
  per spec §2). No render change.
- **/state equipment `itemId:0` / no `name`** — by-design for M1 (matches frozen
  §5 `{slot, sprite, itemId}`, itemId best-effort 0; M2 enriches). Equipment shows
  sprite numbers in M1 as expected.
- **/state inventory `DefaultOptionID` hardcoded 0** — structurally correct
  (`BuildMenu` always places the left-click default at index 0 per §3). No
  behavioral bug; left as a documented invariant.
- **Action soft-timeout dropped error** — a slow action that exceeds 1.5s returns
  ok:true "(running)" and its eventual error is not surfaced. Matches §6.4 intent
  for M1 interactive use; not changed.
- **/chat cosmetic `::` self-echo** — the self-echo shows the bare typed text; the
  busy-fallback concern it noted is resolved by the enqueueAction fix above. The
  optional server-side `::`/`/` stripping for echo symmetry is cosmetic; not done.

---

## 5. Final gate results

Run from the module root (`github.com/gen0cide/westworld`):

| Gate | Command | Result |
|---|---|---|
| build | `go build ./...` | PASS |
| vet | `go vet ./...` | PASS |
| test | `go test ./render/... ./remoteclient/... ./runtime/...` | PASS |

Render fidelity tests (`render/fidelity_*_test.go`) pass — rendering stays
byte-identical after the `DrawEntitySprites` projection-helper refactor. The new
`remoteclient` `ResolveLane` purity/validation suites pass. No test was weakened
or deleted to make the gate green.
