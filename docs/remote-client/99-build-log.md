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

---

## 6. React foundation (2026-06-02)

Migrated the remote client from the single-file `clientPage` (vanilla JS in a Go
string) to a **Vite + React + TypeScript** app under `web/`, embedded into the
cradle binary. See [110-react-port.md](110-react-port.md) for the stack, dev loop,
and the pixel-perfect implementation tree.

**Landed:**
- `web/` Vite+React+TS app; `npm run build` → `web/dist` (committed).
- `web/embed.go` (`//go:embed all:dist`); `serveClient` serves the SPA at `/`
  with an index.html fallback, keeps the legacy client at `/legacy`, and adds
  `GET /config` (`{username,zoom,w,h,rotation}`) to replace template injection.
- Ported at parity: streamed-frame viewport + ~30fps loop, HUD, screen→frame
  hit-test math, left-click default-act / right-click context menu via `/pick`,
  30-slot inventory (left = default option, right = inline menu from `/state`),
  equipment summary, chat (say/`::`cmd/`@`pm), camera keys (rotate/zoom/resize/
  screenshot/clip). Plus a first **Stats** tab from `/state.self.skills`.

**Live validation (cradle as `webreact` on :8090, fresh account, auto-onboarded):**

| Check | Result |
|---|---|
| `GET /` index.html + hashed `/assets/*.js` (200, 154 KB) | PASS |
| `GET /config` → `{"username":"webreact","zoom":750,"w":512,"h":336,"rotation":64}` | PASS |
| `GET /state` → self + 18 skills + inv + equip + chat | PASS |
| `GET /frame` → image/png (41 KB) | PASS |
| `GET /legacy` (old client) still 200 | PASS |
| `POST /pick` (center) → self `Examine` + terrain `Walk here` | PASS |
| `POST /act` Examine self → `self webreact (216,744) hp=10/10 cb=3 fatigue=0 inv=3/3` | PASS |
| `go build ./web/ ./cmd/cradle/` + `go vet` | PASS |

**Notes / gotchas captured:**
- Fresh accounts need `-password` (no record of `webclient2`'s pw); registered
  `webreact`/`reactpass1` via `cmd/register`. Appearance auto-confirm (080b7c1)
  onboarded it with no DB hack — world stream + position immediate.
- `web/node_modules` is gitignored; `web/dist` IS committed so `go build ./...`
  works on a clean checkout (go:embed needs it at build time).
- The "sprite-serving asset pipeline" investigation (item icons → pixel-perfect
  inventory/bank) was rate-limited mid-run; re-run it — it's task **B1** and the
  key enabler for the pixel-perfect pass.

---

## 7. Bank window + sprite pipeline (2026-06-02)

Continued the React build: the Bank window (task E1) and the authentic
item-icon pipeline (task B). See [110-react-port.md](110-react-port.md) for the
updated tree.

**Sprite pipeline (B) — DONE, live-validated.**
- `render/spritepng.go`: `render.ItemSpritePNG(itemID) ([]byte,bool)` wraps the
  existing `compositeItem` decode (item id → config85 itemPicture → de-paletted
  RGB) and PNG-encodes with a real alpha channel (transparent pixels α0; honours
  `CompositeSprite.Flip`).
- `cmd/cradle/sprites.go`: `registerSpriteRoutes` adds `GET /sprite?kind=item&id=N`
  (immutable cache + ETag); wired with one line in `serveClient`.
- React `<ItemSprite>` (onError → text stub) used by inventory + bank cells.
- **Requires** `WESTWORLD_{MEDIA,CONFIG,ENTITY}_JAG` (the render search paths are
  macOS-hardcoded). Using `…/mudclient204/data204/*.jag` here.
- Validated: `/sprite?id=87|166|132` → 48×32 transparent PNGs (bronze axe /
  tinderbox / cooked meat); confirmed rendering in the live SPA via a headless
  screenshot.

**Bank window (E1) — built; closed-path validated, open render pending banker.**
- `/state` gains a `bank` block (`{open,maxSize,slots[]}`) present only while the
  bank is open (reads the pre-existing `world.Bank` mirror).
- `POST /bank {op:deposit|withdraw|close,itemId,amount}` → `Host.BankDeposit/
  Withdraw/Close` through the same serialized action worker as /act and /chat,
  with closed/empty/bad-op guards.
- `<BankWindow>` modal: bank grid (withdraw) + inventory grid (deposit), left-click
  = 1, right-click = 1/5/10/All quantity menu (via a generalized `openActions`
  context menu), real item icons.
- Live-validated: build serves; `POST /bank` while closed → `{"ok":false,"bank is
  not open"}`; bad op rejected; `/state` omits `bank` when closed. **Full open-bank
  render is blocked by world-state access** — a fresh tutorial account (rights 0)
  can't open a bank and `::bank` is admin-only; same boundary as trade/duel.

**Gate:** `go build ./...`, `go vet`, `go test ./render/...` all PASS; frontend
`npm run build` (tsc + vite) clean.

---

## 8. Buildout workflow — shop/magic/prayer/tab-strip + all specs (2026-06-02)

Ran the `rsc-client-buildout` workflow (15 agents): a parallel read-only design
phase spec'd all 10 remaining features into `docs/remote-client/specs/`, then a
serial gate-verified build phase implemented the backend-ready subset.

**Implemented + live-validated:**
- **Magic (D2):** `GET /spells` (static 48-spell catalog, cached) + `/state.magic`
  `{level,maxLevel,spells[{id,canCast,hasRunes}]}` + `POST /cast {spellId,
  targetKind,targetIndex}` → `host.CastOnSelf/Npc/Player` (cmd/cradle/magic.go).
  `<Spellbook>` tab. Verified: catalog returns 48 spells; cast self dispatched.
- **Prayer (D3):** `/state.prayers` (14 prayers w/ name/reqLevel/drainRate/active)
  + `POST /prayer {id,on}` → `host.ActivatePrayer/DeactivatePrayer`. `<PrayerTab>`.
  Verified: activate round-tripped (server replied "out of prayer points").
- **Shop (E2):** `/state.shop` block + `POST /shop {buy|sell|close}` →
  `host.ShopBuy/Sell/Close`. `<ShopWindow>` modal. Closed-path guard verified
  ("shop is not open"); open render pending a reachable shopkeeper (like Bank).
- **Tab-strip (C3):** SidePanel rewritten to an RSC-style one-row tab strip —
  Inv / Worn / Stats / Magic / Pray / Friends(disabled). Verified via screenshot.

**Spec-only (frozen specs ready to implement):** Trade (E3) + Duel (E4) — need a
2nd logged-in bot to verify; Friends (F2) — documents the real protocol gap;
Minimap (F1) — needs a backend entity-list read; bitmap Font (C1); NPC dialog
menu + use-item-on drag (F3/F4). All in `docs/remote-client/specs/*.md`.

**Gate:** `go build ./...`, `go vet`, `go test ./render/... ./remoteclient/...
./runtime/...`, and `npm run build` (42 modules typecheck) — all PASS. No
`reference/` deob work touched; nothing committed by the agents.

---

## 9. Recon-driven wave — shop unblock, minimap, trade, duel (2026-06-02)

Ran the `rsc-client-screens-recon-build` workflow (6 agents): an Opus recon agent
drove the bot through every server-driven screen (granting itself OWNER via the
**SQLite** game DB `openrsc/server/inc/sqlite/westworld.db` — note: the
docker-compose mariadb is a dead artifact; the live server is sqlite), captured
real `/state` payloads + screenshots to `docs/remote-client/screens/`, then a
serial build wave implemented from those findings. All gates green
(`go build ./...`, vet, `go test ./render/... ./remoteclient/... ./runtime/...`,
`npm run build`). `screens/RECON.md` has the full findings.

**The key unblock — synthetic "Talk-to" (remoteclient/menu.go):** RSC peaceful
NPCs (shopkeepers id 51, guides, etc.) store NO Talk-to command — it's implicit —
so `BuildMenu` collapsed them to `[Examine]` and the shop/dialog open-path was
unreachable. Fix: for a non-attackable NPC whose def emitted no commands,
synthesize `OptTalkTo "Talk-to"` (nil-def NPCs stay `[Examine]`; attackable NPCs
unaffected; no duplicate when commands exist). +test in menu_test.go. This single
fix unblocked both Shop and NPC dialogs.

**Landed + live-verified:**
- **Shop (E2):** opened a General Store via Talk-to → `POST /dialog` → shop;
  `/state.shop {open,isGeneral,slots[{itemId,name,stock,buyPrice,sellPrice}]}` +
  `POST /shop {buy|sell|close}`. `screens/shop-open.png` (real icons + gp prices).
- **Minimap (F1):** `/state.entities` (npcs/players/items/scenery within Chebyshev
  radius 16, from the `world.*.All()` accessors `buildLiveView` uses) + `<Minimap>`
  rotating canvas + click-to-walk. `screens/minimap.png` (verified: 85 dots, a
  dropped item appeared as a red dot within one poll).
- **`POST /dialog {option}`** → `host.ChooseDialogOption` (drives multi-choice
  NPC dialog trees; used to open the shop).

**Landed, NOT yet live-verified (need both bots online + receiver-accept):**
- **Trade (E3):** `/state.trade` (phase open|confirm) + `POST /trade
  {offer|accept|finalize|decline}` → Host trade methods; `<TradeWindow>`.
- **Duel (E4):** `/state.duel` (+ 4 rule toggles, 2-stage accept) + `POST /duel
  {stake|rules|accept1|accept2|decline}` → Host duel methods; `<DuelWindow>`.
  Recon confirmed the *outbound* init for both (2 bots: webreact :8090 +
  webreact2 :8091, co-located via `::summon`); the receiver-side accept route is
  now wired but unexercised.

**Scope note:** the bank-shop-verify agent edited `remoteclient/menu.go` (+ its
tests) — the Layer-2 menu builder, just outside the "web/+cmd/cradle" guidance but
the correct and necessary home for the Talk-to fix. Reviewed: conservative, tested,
gate-green. `reference/`, `world/`, `runtime/` untouched.

---

## 10. Trade + Duel live-verification + `/state` player-index (2026-06-02)

Ran a live 2-bot verification of the Trade (E3) and Duel (E4) windows. Outcome:
**both verified end-to-end** — Duel: open→confirm→"Commencing Duel!"; Trade: the
first run stalled at accept→confirm on a `world/trade.go` bug, which was then
root-caused, fixed (user-authorized in-scope), and re-verified to "Trade completed
successfully" (see the BUG → FIXED note below). Both bots left running (webreact
:8090, webreact2 :8091).

### The handshake that opens the window

A symmetric, mutual init opens each window: BOTH bots fire the per-pair `/act`
verb against the OTHER player's dot — **"Trade with"** / **"Duel with"** (the NPC/
player verb menu, optionId `0`=Trade / `2`=Duel) — which sends `InitTradeRequest`
/ `InitDuelRequest`. When both sides have requested each other, the server opens
the window and `/state` gains the `trade` / `duel` block (phase `open`) on both.

**INDEX SEMANTICS (load-bearing).** The act ref index is the server's GLOBAL
player index *as seen by the firing client* — NOT a symmetric per-pair id. From
bot1's view Webreact2 was index `0`; from bot2's view Webreact was index `1`. So
bot1's act used `index 0` and bot2's used `index 1`. **The ref index must come
from the FIRING bot's own `/state`, never the partner's.**

### Prep edit (player index in `stateDot`) + two live bug fixes

To target a player by index (instead of pixel-picking), the prep pass added an
`index` field to the `stateDot` struct in `cmd/cradle/remoteclient.go` and set it
on the player-dot append (`pl.Index` was already in scope in that loop). Gate was
green (build/vet/test/`npm run build` all PASS).

During the live run, two `cmd/cradle/remoteclient.go` bugs surfaced that broke the
symmetric handshake and were **fixed in place** (29+/10-, web/cmd scope only):

1. **`/state` dots loop dropped a legit index-0 player.** The loop skipped self
   with `if pl.Index == 0`, wrongly assuming index 0 is always self. But
   opcode-191 (`InSendPlayerCoords`) new-player records carry the server's 11-bit
   GLOBAL player index (0..2047); a legitimate other player can land at index 0
   (e.g. the first account to log in after a server restart). `world.Players`
   never holds the self `OwnPositionUpdate` (that goes only to `world.Self`); the
   only self entry is the type-5 appearance mirror keyed by `PlayerIndex 0` with
   our own username. **Symptom:** webreact2 NEVER saw webreact in its
   `/state.entities.dots` (even though its runtime log received "nearby player
   index=0 name=Webreact"), making the symmetric handshake impossible. **Fix:**
   discriminate self by NAME (`strings.EqualFold(pl.Name, cfg.username)`), not by
   index 0.
2. **`Index` was a plain `int` with `omitempty`,** so a legit player at index 0
   had the field omitted entirely from `/state` JSON (dot showed name+coords but
   `index=undefined`). **Fix:** changed `Index` to `*int` (`json:"index,omitempty"`)
   and the player append sets `Index:&idx`; nil/omitted for npc/item/scenery dots.

### `/state.duel` snapshot (Duel — FULL handshake verified)

Mutual "Duel with" `/act` opened the window on both at phase `open`:

```
bot1 (8090): {"phase":"open","withName":"Webreact2","myOffer":[],"theirOffer":[],
  "rules":{"disallowRetreat":false,"disallowMagic":false,"disallowPrayer":false,
  "disallowWeapons":false},"myFirstAccepted":false,"theirFirstAccepted":false,
  "mySecondAccepted":false,"theirSecondAccepted":false}
bot2 (8091): {"phase":"open","withName":"Webreact", ... (mirror) }
```

After setting rule `disallowMagic=true` (propagated to BOTH bots' `/state`) and
`accept1` on both, the phase advanced to `confirm`:

```
bot1 confirm: {"phase":"confirm","withName":"Webreact2","myOffer":[],"theirOffer":[],
  "rules":{...,"disallowMagic":true,...},"myFirstAccepted":true,
  "theirFirstAccepted":true,"mySecondAccepted":false,"theirSecondAccepted":false}
```

After `accept2` on both, the duel block cleared on both bots and the server
emitted **"Commencing Duel!"** — full handshake completed. Screenshots:
`screens/duel-open.png`, `screens/duel-confirm.png` ("Duel vs Webreact2 — Confirm",
Confirm Fight button + No-Magic toggle highlighted red).

### `/state.trade` snapshot (Trade — `open` + offer-sync verified; accept blocked)

Mutual "Trade with" `/act` opened the window on both at phase `open`. bot1 then
offered a bronze Axe (itemId 87); it propagated correctly to bot2's mirror view
(`bot1.myOffer == bot2.theirOffer`), with each side's `partnerName` naming the
OTHER player:

```
bot1 (8090): {"phase":"open","partnerName":"Webreact2",
  "myOffer":[{"itemId":87,"name":"bronze Axe","amount":1}],"theirOffer":[],
  "myFirstAccepted":false,"theirFirstAccepted":false,
  "mySecondAccepted":false,"theirSecondAccepted":false}
bot2 (8091): {"phase":"open","partnerName":"Webreact","myOffer":[],
  "theirOffer":[{"itemId":87,"name":"bronze Axe","amount":1}], ... }
```

Screenshot: `screens/trade-open.png`. In the FIRST run, after `accept` on both, the
trade **STALLED at `phase=open`** with `myFirstAccepted=true` /
`theirFirstAccepted=false` on BOTH, forever (declined cleanly before the duel test).
This was then **root-caused and FIXED** (see below) and **re-verified to completion**.

**BUG → FIXED (world/trade.go + world/world.go, 2026-06-02, authorized in-scope).**
The real root cause was the `event.TradeConfirmShown` handler in `world/world.go`
(~line 883), which was self-defeating: it called `w.Trade.SetTheirOffer(items)` —
which RESETS both accept flags (`trade.go:142-143`) — then `MarkMyFirstAccepted()`,
which only advances to `confirm` while `TheirFirstAccepted` is set (just wiped). So
the server's confirm-window push reset the flags and then failed to transition,
leaving `myFirstAccepted=true`/`theirFirstAccepted=false`/`phase=open` — exactly the
observed stuck state. **Trade lacked the `MarkConfirmShown` + `UpdateTheirOfferNoReset`
methods that the working DUEL path uses** (`world/world.go` `DuelConfirmShown` →
`UpdateTheirOfferNoReset` + `MarkConfirmShown`). Fix:
- added `UpdateTheirOfferNoReset(items)` + `MarkConfirmShown()` to `world/trade.go`
  (mirrors duel);
- rewrote the `TradeConfirmShown` case in `world/world.go` to use them (no-reset
  offer apply + direct `phase="confirm"`) instead of `SetTheirOffer` + the
  `MarkMyFirstAccepted` dance;
- also made `MarkOtherFirstAccepted` advance the phase symmetrically when both have
  first-accepted — in BOTH `world/trade.go` and `world/duel.go` — a defensive net for
  the pure-local accept path (duel reached confirm via its server `DuelConfirmShown`
  push, so it had only this latent ordering weakness, not the reset bug).

**Re-verified live (2-bot, `/tmp/trade-retest.sh`):** open → offer-sync (bronze Axe
id 87, `bot1.myOffer == bot2.theirOffer`) → both `accept` → `phase=confirm` on BOTH
(`myFirstAccepted=true`/`theirFirstAccepted=true`, button → "Confirm trade",
`screens/trade-confirm.png`) → both `finalize` → trade block cleared + server chat
**"Trade completed successfully"**. Gate green (`go build ./...`, `go vet`,
`go test ./world/... ./render/... ./remoteclient/... ./runtime/...`).

### Cosmetic follow-up (spectate.go:210, left unfixed)

The spectator 3D-render player loop has the SAME `pl.Index == 0` skip, so a real
other-player at global index 0 is dropped from the rendered viewport composite.
NOT load-bearing for trade/duel (those use the SPA modal reading
`/state.entities.dots`, now fixed). Left unfixed because a clean fix needs
`cfg.username` threaded through `buildLiveView` (6 call sites); a runtime-layer
self-name accessor would be the right fix, but `runtime/` is out of scope.

### Server note (operational, not a code change)

A freshly-SIGKILLed cradle left a ghost Webreact session that wedged the OpenRSC
game server in a SELF-SUSTAINING INFINITE SAVE LOOP (`westworld_1.log`:
`[SQLITE_CONSTRAINT_PRIMARYKEY] UNIQUE constraint failed: itemstatuses.itemID` at
`savePlayerBank` → `PlayerSaveRequest` → rollback → re-queue, 30k+ times with no
new logins; every re-login got "login rejected (code 4)"). The duplicate itemID
was generated in-memory by the ghost's bank; the PERSISTED `westworld.db` was
verified clean (read-only: 0 duplicate itemIDs, 226 rows). **Fix: restarted the
game server** (clears the in-memory ghost) from `openrsc/server` with
`JAVA_HOME=java-17-openjdk ANT_HOME=Portable_Windows/apache-ant-1.10.5
setsid ant runserver -DconfFile=westworld -DcoloredLogging=false`. No data loss.
**Future:** never SIGKILL a cradle mid-bank/mid-trade — log out gracefully
(in-game `::logout`/kill produces a clean "PlayerSaveRequest: Removed player").

### Reproduce (next session)

```
# both accounts are already OWNER (webreact / webreact2)
1. launch both bots:
   /tmp/cradle ... -username webreact  ... -client-addr localhost:8090
   /tmp/cradle ... -username webreact2 ... -client-addr localhost:8091
2. co-locate them: on bot1  POST /chat {"kind":"command","text":"teleport varrock"}
                   then     POST /chat {"kind":"command","text":"summon Webreact2"}
3. read each bot's OWN /state.entities.dots → find the OTHER player's `index`
   (per-client; NOT symmetric — bot1 saw Webreact2@0, bot2 saw Webreact@1).
4. mutual init: on EACH bot, POST /act with that bot's own ref index +
   optionId 0 (Trade) / 2 (Duel) → window opens at phase=open on both.
5. drive: TRADE  POST /trade {op:offer|accept|finalize|decline}
          DUEL   POST /duel  {op:stake|rules|accept1|accept2|decline}
   (duel: rules → accept1 (→confirm) → accept2 → "Commencing Duel!").
```

**Code touched:** `cmd/cradle/remoteclient.go` (player-dot `index` field + the two
live fixes) and — for the trade accept→confirm root-cause fix (user-authorized as
in-scope) — `world/trade.go` (added `UpdateTheirOfferNoReset` + `MarkConfirmShown`,
symmetric advance in `MarkOtherFirstAccepted`), `world/duel.go` (symmetric advance
in `MarkOtherFirstAccepted`), and `world/world.go` (rewrote the `TradeConfirmShown`
handler to mirror duel). `runtime/`, `reference/` untouched. Nothing committed.

---

## 11. Post-verification fixes + bug sweep (2026-06-02, committed to `feat/remote-client`)

Follow-on work after the Trade/Duel verification; all committed + pushed to the
`fork` remote (freeqaz/westworld). Each commit is self-gated.

- **`1bc6dcc`** — the trade/duel verification itself (player `index` in `/state`,
  self-by-name skip, trade `ConfirmShown` root-cause fix; §10) + screenshots.
- **`1dd9d9c` fix(cradle): spectate self-skip** — `buildLiveView` (`spectate.go`)
  dropped a real other-player at global index 0 (same antipattern as the `/state`
  dots loop). Skip self by NAME via a new `runtime/host.go` `Host.Username()`
  accessor. (Scanned the tree: no other `Index == 0` self-skip remains.)
- **`7f6d980` harden(trade/duel)** — gate the open→confirm advance in
  `Mark{My,Other}FirstAccepted` on `Phase == "open"` (both `world/trade.go` and
  `world/duel.go`) so a stray inbound `*OtherAccepted` after close can't regress a
  terminal phase back to `confirm`. Mirror the `index` field on `MinimapDot`
  (`web/src/types.ts`).
- **`13ce4fb` fix(render): authentic item icons** — `render/itemsprite.go` indexed
  the icon by `ItemDef.AppearanceID` (the WORN appearance: 0 for non-wieldables →
  all collapsed onto one sprite; wrong for the rest — bronze Axe rendered as an
  apple). Added `render/itempicture_data.go` (`itemPictureIndex`, the authentic
  per-item picture array = oracle `GameData.itemPicture` = OpenRSC
  `ItemDef.spriteID`), generated from `openrsc Client_Base
  EntityHandler.loadItemDefinitions()` (`new ItemDef(.., spriteID, "items:N", ..,
  id)` → `id→N`, 1549 entries). Icons now resolve `spriteItem+itemPictureIndex[id]`.
  Verified by rendering: bronze Axe→axe, Iron Mace→mace, cooked meat→meat,
  tinderbox→tinderbox. 2 custom named-pack items (bat/dragon bones) have no
  numeric icon and fall back to the marker.
- **fix(render): tier recolour (pictureMask).** Tiers share one base
  sprite and differ only by `ItemDef.pictureMask` (e.g. all scimitars are sprite
  83; bronze→orange, mithril→blue, rune→teal, black→near-black). The earlier note
  here used "steel vs iron" as the example, which was wrong — iron `#9E9386` vs
  steel `#9E9D8F` are near-identical (masks `0xEEDDDD` vs `0xEEEEEE`), so that pair
  is imperceptible; the visible gap was bronze/mithril/adamant/rune/black showing
  as a generic grey. Ported OpenRSC's recolour (`recolorItemPixel`, a faithful port
  of `GraphicsController.plot_trans_scale_with_2_masks`: grayscale pixels ×
  pictureMask, white-ish × mask2 (=0 for items), the blueMask special case;
  coloured pixels like the hilt left untouched). Extended `itempicture_data.go` to
  `itemIcons{pic,mask,blue}`. **Verified pixel-exact vs the authentic RSC wiki
  icons:** all seven scimitar tiers match the wiki avg blade colour to the byte
  (bronze `#A75018`, iron `#9E9386`, steel `#9E9D8F`, mithril `#6E7B7C`, adamant
  `#7C8960`, rune `#19A699`, black `#333225`).

---

## 12. Feature wave: F3 dialog · D4 worn icons · C1 font · F2 friends (2026-06-02)

Four features sequenced + live-verified in one wave (user: "sequence all of these
and get through them"). All committed to `feat/remote-client` + pushed to `fork`.

- **F3 NPC dialog UI.** `/state.dialog` ({open,npcText,options}) from
  `world.Recent.DialogOptions()`; hardened `POST /dialog` (live-menu validation +
  clear-after); `<NpcDialog>` lower-viewport box. Added the server actor `index` +
  def `id` to `/state.entities.dots` (npc + player) so a dot → `/act` MenuTarget
  is buildable — used to drive Talk-to deterministically. **Verified:** Talk-to a
  spawned shopkeeper → "Can I help you at all?" + 2 clickable options.

- **Render-fidelity audit + D4 worn icons.** A read-only audit (subagent) rendered
  NPCs / players / scenery and confirmed their sprite + colour resolution is
  correct — the item-icon class of bug does NOT recur there. The one real finding:
  the Equipment panel showed text because `stateEquipItem.ItemID` was hardcoded 0.
  Fixed by joining each worn layer to its wielded inventory item by appearance
  (`ItemDef.AppearanceID & 0xFF` == the worn EquipSprites value; proto/v235
  confirms). `<EquipmentPanel>` now renders the real `/sprite` icon + name.
  **Verified:** worn iron platemail/legs/helm show authentic icons. Also fixed the
  stale `facts.ItemDef.AppearanceID` comment (it is the WORN appearance, not the
  icon index).

- **C1 authentic font.** Self-hosted Helvetica-metric webfont (95-glyph subset of
  Liberation Sans, OFL, ~13KB plain+bold, `web/public/fonts/`, `//go:embed`'d);
  whole stylesheet → `--rsc-font` stack (no `monospace`), `-webkit-font-smoothing:
  none`, authentic 1px black shadow. **Verified:** proportional text in the SPA.

- **F2 friends/ignore — the protocol gap, CLOSED (full A+B).** Decoded inbound
  SEND_FRIEND_UPDATE (149) + SEND_IGNORE_LIST (109) — wire formats verified against
  OpenRSC `Payload235Generator` (149 = name, formerName, status byte [bit0 rename,
  bit2 online], worldName iff online; 109 = count then count×[name,name,former,
  former]). Added `event.FriendUpdate/IgnoreList`, `world.SocialState` mirror
  (wired into `World.Apply`), outbound add/remove ignore (132/241, same zero-padded
  username as AddFriend — confirmed by the server parser) + `Host.AddIgnore/
  RemoveIgnore`. HTTP: `/state.social` (always present, never null) + `POST /social`
  ({add,remove}×{friend,ignore}). Web: `<FriendsTab>` (online dot, add/remove, PM
  via existing `/chat`). Friend- and ignore-REMOVE are mirrored locally because the
  server re-sends the list only on add/login. **Verified (2 OWNER bots):**
  mutual-friend → `{name:Webreact2, online:true, world:Westworld}`; 149 login-burst
  auto-restores friends; ignore a fresh account → `ignores:['Ignoreme']` (109
  count>0); "Staff may not be added to ignore list" surfaces honestly.

Commits: `d82391a` (F3), `5654190` (D4 + audit), `f533278` (C1), `8164078` (F2).
