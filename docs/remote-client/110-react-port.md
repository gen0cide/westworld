# Remote Client → React Port + Pixel-Perfect UI Tree

Status doc for migrating the cradle remote client from the single-file vanilla-JS
page to a **Vite + React + TypeScript** app, then reimplementing the authentic
RuneScape Classic UI as close to pixel-perfect as practical.

Authoritative references:
- UI ground truth → [100-rsc-ui-map.md](100-rsc-ui-map.md) (layouts/colors/sprites)
- API contract → [30-http-api.md](30-http-api.md) + [50-impl-spec.md](50-impl-spec.md)
- M2 backend audit → [90-backlog.md](90-backlog.md)

---

## 0. Stack & layout (DECIDED)

- **Vite + React 18 + TypeScript**, source under `web/`.
- Built to `web/dist`, embedded into the cradle binary via `web/embed.go`
  (`//go:embed all:dist`) and served at `/` by `serveClient`
  (`cmd/cradle/remoteclient.go`). The legacy single-file client stays at
  `/legacy` for A/B comparison.
- Render defaults come from **`GET /config`** (`{username,zoom,w,h,rotation}`),
  replacing the old `fmt.Fprintf` template injection.

### Dev loop
```
cd web && npm install            # once
npm run dev                      # vite @ :5173, proxies API → :8090 (HMR)
# in another shell, run the bot+server:
cradle -server localhost:43594 -username <acct> -password <pw> \
       -facts <openrsc-root> -client -client-addr localhost:8090
```
Set `CRADLE_ADDR` to point the dev proxy elsewhere.

### Prod build (what the binary serves)
```
cd web && npm run build          # tsc -b && vite build → web/dist
go build -o /tmp/cradle ./cmd/cradle
```
`web/dist` is committed so `go build ./...` is green on a fresh checkout. **After
any frontend change, `npm run build` before rebuilding cradle**, or the binary
serves stale assets. (A future `go generate` could automate this.)

### File map (`web/src/`)
- `types.ts` — wire types mirroring the Go JSON structs.
- `api.ts` — typed client (`getConfig/getState/frameURL/walk/pick/act/sendChat/…`).
- `state.ts` — `useConfig()`, `useGameState()` (polls `/state`, dedupes chat by seq).
- `ui.tsx` — `UIProvider`/`useUI`: shared flash toast + right-click `ContextMenu`.
- `App.tsx` — camera state, anim/rotate timers, keyboard controls, layout.
- `components/` — `Viewport`, `Hud`, `ChatPanel`, `SidePanel`, `InventoryGrid`,
  `EquipmentPanel`, `StatsPanel`.

---

## 1. Backend is far ahead of the UI

The Go runtime/world/proto layers already implement the systems; the web client
just doesn't surface them yet. Most "features" below = **expose existing state
over HTTP + draw the panel**, not new protocol work.

| System | Backend (exists) | Web exposed |
|---|---|---|
| inventory / equipment / chat / examine | ✅ | ✅ |
| world pick + menu dispatch | ✅ | ✅ |
| skills/stats | ✅ `views_self.go` (`/state.self.skills`) | ⚠ data flows; basic tab only |
| bank | ✅ `world/bank.go`, `Host.BankDeposit/Withdraw/Close` | ❌ |
| shop | ✅ `world/shop.go`, `Host.ShopBuy/Sell/Close` | ❌ |
| trade | ✅ `world/trade.go` + handshake opcodes | ✅ live-verified full handshake (open→confirm→completed) |
| duel | ✅ `world/duel.go`, `runtime/actions_duel.go` | ✅ live-verified full handshake |
| magic / prayer | ✅ `actions_magic.go` / `actions_prayer.go` + views | ❌ |
| friends / ignore | ✅ roster decode (149/109) + ignore opcodes (132/241) now done | ✅ `<FriendsTab>` |

---

## 2. Implementation TREE

Legend: `[x]` done · `[~]` partial · `[ ]` todo.

### A. Foundation  `[x]`
- [x] A1 Vite+React+TS scaffold under `web/`; `npm run build` green.
- [x] A2 `web/embed.go` + serve `dist` at `/`, SPA fallback, `/legacy` kept.
- [x] A3 `GET /config` endpoint + `useConfig` seed.
- [x] A4 Port viewport frame-loop (`Viewport`), HUD, screen→frame math.
- [x] A5 Port pick→default-act (left) and pick→context-menu (right).
- [x] A6 Port inventory grid (left = default opt, right = inline menu).
- [x] A7 Port equipment summary + chat (say/cmd/pm) + camera keys.
- [x] A8 Live-validated on :8090 (config/state/frame/pick/act/legacy all 200).

### B. Asset pipeline — **pixel-perfect enabler**  `[x]`  (item icons live)
- [x] B1 `GET /sprite?kind=item&id=N` → transparent PNG via `render.ItemSpritePNG`
      (render/spritepng.go) + `cmd/cradle/sprites.go` (`registerSpriteRoutes`).
      **Sprite source (updated 2026-06-02 by the concurrent effort):** resolves
      from OpenRSC `Authentic_Sprites.orsc` at `spriteItem + ItemDef.AppearanceID`
      (`render/itemsprite.go`, archive loaded by `render/sprites.go::sprites()`),
      located via env var **`WESTWORLD_SPRITES_ORSC`** — the old
      `WESTWORLD_{MEDIA,CONFIG,ENTITY}_JAG` vars are **obsolete**. On this box:
      `openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc`.
- [x] B2 Immutable cache headers + ETag (`item-N`). Icons are static.
- [x] B3 React `<ItemSprite id name>` (onError → text stub); used by inventory +
      bank cells. Validated live: bronze axe/tinderbox/cooked-meat render as real
      48×32 icons in the SPA.
- [x] B4 tab-strip icons — DONE 2026-06-02. `render.MediaSpritePNG` + GET
      `/sprite?kind=media&id=N` serve the authentic `spriteMedia+0` closed-tab-strip
      sprite (197×32, black colour-keyed → transparent); `SidePanel` fetches it once
      and CSS-crops a 32px icon cell per tab (`object-fit:none` + per-tab
      `object-position`). Live-verified in-browser (all six icons render crisp).
      Worn→wrench and Pray→map are the only non-1:1 cells (rev 235 has no
      worn/prayer/friends tab). Borders/buttons remain CSS; `kind=npc|player`
      sprite serving (compositeNPC/compositePlayer) is feasible later but not wired.

### C. Pixel-perfect chrome  `[~]`
- [x] C1 **Authentic font — DONE + live-verified.** Self-hosted Helvetica-metric
      webfont (a 95-glyph subset of Liberation Sans, OFL, ~13KB plain+bold under
      `web/public/fonts/`, `//go:embed`'d); whole stylesheet switched to a
      `--rsc-font` stack (no `monospace`); `-webkit-font-smoothing:none` + the
      authentic 1px black shadow. Proportional text confirmed in the SPA.
- [x] C2 RSC color palette as CSS vars (§3); chat recolored (self=cyan, npc=yellow,
      pm=blue, system=dim). Menu/panels partly themed.
- [~] C3 Inventory is now a 5-wide grid sized to the native 48×32 icon (RSC
      layout); still TODO: 512×334 fixed-aspect game area + authentic sprite
      tab-strip (currently text "Inventory/Stats" buttons) + chat tab bar.
- [ ] C4 Right-click menu styled to `drawMenuOptions` (gray box, white border,
      yellow hover/title) — current menu is close but not exact.

### D. Read panels (cheap; data already in `/state`)  `[x]`
- [~] D1 Stats/skills tab (have a basic one; make pixel-perfect per §4.2).
- [x] D2 **Magic spellbook tab** — DONE. `GET /spells` serves the static 48-spell
      catalog (cached); `/state.magic` carries `{level,maxLevel,spells[{id,canCast,
      hasRunes}]}`; `POST /cast {spellId,targetKind,targetIndex}` →
      `host.CastOnSelf/Npc/Player` (cmd/cradle/magic.go). `<Spellbook>` tab with
      RSC color coding. Validated live (catalog + flags + cast dispatch).
- [x] D3 **Prayer tab** — DONE. `/state.prayers[{id,name,reqLevel,drainRate,active}]`
      (14 prayers); `POST /prayer {id,on}` → `host.ActivatePrayer/DeactivatePrayer`.
      `<PrayerTab>` with toggle + right-click menu. Validated live (the activate
      round-trip surfaced the server's "out of prayer points" message).
- [x] D4 **Equipment worn icons — DONE + live-verified.** `/state.equipment` now
      joins each worn layer to its wielded inventory item by appearance
      (`ItemDef.AppearanceID & 0xFF` == the EquipSprites value), so `<EquipmentPanel>`
      shows the real item icon (via `/sprite`) + name; text fallback for unresolved
      layers (e.g. the default body). Verified: worn iron platemail/legs/helm render
      their authentic icons. (From the render-fidelity audit, which otherwise found
      NPC/player/scenery sprite resolution correct.)

### E. Windows (server-driven; expose state + 2 actions each)  `[ ]`
- [~] E1 **Bank** — DONE except a live open-bank render. `/state` now emits a
      `bank` block (`{open,maxSize,slots[]}`) only while the window is open;
      `POST /bank {op:deposit|withdraw|close,itemId,amount}` →
      `Host.BankDeposit/Withdraw/Close` via the serialized worker; `<BankWindow>`
      modal (bank grid = withdraw, inventory grid = deposit, right-click = 1/5/10/
      All quantity menu, real icons). Closed-path + guards validated live. **Open
      render pending a reachable banker** — the fresh tutorial bot (rights 0) can't
      open a bank and `::bank` is admin-only; same world-state boundary as E3/E4.
      To test: drive a bot to a banker NPC, or grant admin + `::bank`/`::fillbank`.
- [x] E2 **Shop** — DONE + live-verified (`screens/shop-open.png`). `/state.shop`
      block + `POST /shop {buy|sell|close}` → `Host.ShopBuy/Sell/Close`;
      `<ShopWindow>` (stock grid w/ gp prices, inventory grid = sell), real icons.
      Unblocked by the **Talk-to fix** (see note below) + `POST /dialog`.
- [x] E4-parity E3 **Trade** — live-verified FULL handshake: open→confirm→completed
      (`screens/trade-open.png`, `screens/trade-confirm.png`). `/state.trade` (phase
      open|confirm) + `POST /trade {offer|accept|finalize|decline}` →
      `Host.OfferTradeItems/ConfirmTrade/FinalizeTrade/DeclineTrade`; two-grid
      `<TradeWindow>`. **Verified (2 bots, 2026-06-02):** mutual "Trade with" `/act`
      opened the window on BOTH sides at phase `open`; an offer of bronze Axe
      (itemId 87) from bot1 propagated correctly — `bot1.myOffer == bot2.theirOffer`,
      each side's `partnerName` naming the OTHER player; after both POST
      `/trade {op:accept}` BOTH advanced to `phase=confirm`
      (`myFirstAccepted=true`/`theirFirstAccepted=true`, button → "Confirm trade");
      after both POST `/trade {op:finalize}` the trade block cleared and the server
      emitted **"Trade completed successfully"**.
      **Root-cause fix (2026-06-02, `world/`):** the first run stalled at `open` —
      the `event.TradeConfirmShown` handler (`world/world.go`) was self-defeating: it
      called `SetTheirOffer` (which RESETS both accept flags) then `MarkMyFirstAccepted`
      (which can only advance while `TheirFirstAccepted` is set, just wiped). Trade
      had no `MarkConfirmShown`/`UpdateTheirOfferNoReset` (the working duel path does).
      Fixed by adding both methods to `world/trade.go` and rewriting the handler to
      mirror duel (no-reset offer apply + direct phase set); also made
      `MarkOtherFirstAccepted` advance the phase symmetrically (trade + duel) as a
      defensive net. Gate green; live re-test reached completed.
- [x] E4 **Duel** — live-verified FULL handshake: open→confirm→accept2→"Commencing
      Duel!" (`screens/duel-open.png`, `screens/duel-confirm.png`). `/state.duel`
      (+ 4 rule toggles, 2-stage accept) + `POST /duel
      {stake|rules|accept1|accept2|decline}` →
      `Host.OfferDuelItems/SetDuelRules/AcceptDuelOffer/AcceptDuelConfirm`;
      `<DuelWindow>`. **Verified (2 bots, 2026-06-02):** mutual "Duel with" `/act`
      opened the window on both at phase `open`; a `disallowMagic=true` rule toggle
      propagated to BOTH bots' `/state`; accept1 on both advanced to `phase=confirm`
      (`myFirstAccepted=true`/`theirFirstAccepted=true`); accept2 on both cleared the
      duel block and the server emitted "Commencing Duel!". The DUEL path does NOT
      have the trade accept-flag bug.
- [ ] E5 Window open/close lifecycle: `/state` should signal which window is open
      so the SPA can show/hide modals (add `window: {kind, ...}` to state).

### F. New ground  `[~]`
- [x] F1 **Minimap** — DONE + live-verified (`screens/minimap.png`). `/state.entities`
      (nearby npcs/players/items/scenery within Chebyshev radius 16, from the same
      `world.*.All()` accessors the `/pick` path uses) + `<Minimap>` rotating canvas
      (dots per §4.4) + click-to-walk via `walkTile`. v1 gaps: friend (green) dots
      (server stub), static scenery, right-click verb menu, compass sprite.
      **NEW (2026-06-02):** player dots in `/state.entities.dots` now carry an
      `index` field (`*int`, the server's GLOBAL player index as seen by THAT
      client; omitted/nil for npc/item/scenery dots). This is what made
      deterministic player-targeting possible without pixel-picking — the Trade/Duel
      `/act` ref index is read straight from the firing bot's own `/state`. **Two
      bugs fixed live to make this work** (`cmd/cradle/remoteclient.go`): (a) the
      dots loop skipped self by `pl.Index == 0`, which wrongly dropped a legit
      other-player at global index 0 (e.g. the first account to log in after a
      server restart) — now skips self by NAME (`strings.EqualFold(pl.Name,
      cfg.username)`); (b) `Index` was a plain `int` with `omitempty`, so a player
      at index 0 had the field silently dropped from JSON — changed to `*int`.
      **Index is NOT a symmetric per-pair id:** from bot1's view Webreact2 was index
      0; from bot2's view Webreact was index 1. The act ref index MUST come from the
      firing bot's own `/state`, never the partner's.
      Known follow-up: the spectator 3D-render loop (`spectate.go:210`) still has the
      same `pl.Index == 0` skip (cosmetic; drops a real index-0 player from the
      rendered viewport composite). Not load-bearing for trade/duel (those read the
      SPA modal's `/state.entities.dots`, now fixed); left unfixed because the clean
      fix needs `cfg.username` threaded through `buildLiveView` (6 call sites).
- [x] F2 **Friends / ignore — DONE + live-verified; the protocol gap is CLOSED.**
      Implemented Track A + the previously-missing Track B: decode inbound
      SEND_FRIEND_UPDATE (149) + SEND_IGNORE_LIST (109) (`proto/v235`), a
      `world.SocialState` roster mirror, outbound add/remove ignore (132/241) +
      `Host.AddIgnore/RemoveIgnore`, `/state.social` + `POST /social`, and the
      `<FriendsTab>`. Verified with 2 bots: mutual-friend → real `online:true`,
      `world:Westworld`; 149 login-burst auto-restores friends; ignore 109 round-trips
      (`ignores:['Ignoreme']`); "Staff may not be added to ignore list" surfaces.
      (Friend- and ignore-REMOVE are mirrored locally — the server re-sends the list
      only on add/login.)
- [x] F3 **NPC dialog-option menu — DONE + live-verified.** `/state.dialog`
      ({open,npcText,options}) from `world.Recent.DialogOptions()`; `POST /dialog`
      hardened (live-menu validation + clear-after); `<NpcDialog>` lower-viewport box.
      Verified: Talk-to a shopkeeper → "Can I help you at all?" + clickable options.
- [x] F4 **Use-item-on-target drag UX — DONE + backend-verified.** `POST /useon`
      {slot,target} routes `target.kind` to the six `host.UseItemOn*` methods
      (re-validating source slot + npc/player visibility); web `<DragProvider>` +
      draggable inventory cells (item-on-item) + Viewport world drop (reuses
      `screenToFrame` + `/pick`). Verified: item-on-item fired "Use item on item";
      empty source → "source slot empty"; bad npc → "npc no longer visible".
- [ ] F5 Multi-bot tabs (90-backlog §6) — later.

> **Pixel-perfect font** (C1) is SPEC'd → `specs/font.md`. All `specs/*.md` are
> frozen, copy-pasteable implementation specs produced by the buildout workflow.

### G. Cross-cutting  `[ ]`
- [ ] G1 Replace 400ms `/state` poll with SSE/WebSocket push for windows/chat that
      need low latency (trade/duel). Polling is fine for inventory/stats.
- [ ] G2 `go generate` (or Makefile) to run `npm run build` before `go build`.
- [ ] G3 Component/contract tests; keep `web` typecheck in the gate.

---

## 3. Suggested order

1. **B (sprite pipeline)** → real item icons unlock pixel-perfect inventory/bank/shop.
2. **C + D1** → authentic chrome + stats, mostly CSS/data already present.
3. **E1 Bank, E2 Shop** → the windows the user asked for; backend complete.
4. **D2/D3 magic/prayer**, then **E3/E4 trade/duel** (need a 2nd bot).
5. **F minimap, friends, dialog menus, use-on**.

Each E/F item is a clean fan-out unit (state endpoint + action endpoint +
component) — good candidates for a Sonnet implementation pass against a frozen
spec, with Opus designing the state/window lifecycle (E5/G1) first.
