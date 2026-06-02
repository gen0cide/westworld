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
| trade | ✅ `world/trade.go` + handshake opcodes | ❌ |
| duel | ✅ `world/duel.go`, `runtime/actions_duel.go` | ❌ |
| magic / prayer | ✅ `actions_magic.go` / `actions_prayer.go` + views | ❌ |
| friends / ignore | ⚠ partial — **the one real protocol gap** (see 90-backlog §4) | ❌ |

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
- [x] B1 `GET /sprite?kind=item&id=N` → transparent PNG. `render.ItemSpritePNG`
      (render/spritepng.go) wraps the existing `compositeItem` decode (item id →
      `config85.jag` itemPicture → de-paletted RGB + alpha). Handler in
      `cmd/cradle/sprites.go` (`registerSpriteRoutes`), one line in serveClient.
      **Requires the jag env vars** (`WESTWORLD_{MEDIA,CONFIG,ENTITY}_JAG`) —
      currently `…/mudclient204/data204/{media58,config85,entity24}.jag`. Search
      paths in render are macOS-hardcoded; deriving them from `factsRoot` is a
      nice-to-have so the vars aren't mandatory.
- [x] B2 Immutable cache headers + ETag (`item-N`). Icons are static.
- [x] B3 React `<ItemSprite id name>` (onError → text stub); used by inventory +
      bank cells. Validated live: bronze axe/tinderbox/cooked-meat render as real
      48×32 icons in the SPA.
- [ ] B4 UI chrome sprites (tab strip, borders, buttons) — optional; CSS may
      suffice. `kind=npc|player` sprite serving is feasible later (compositeNPC/
      compositePlayer) but not wired.

### C. Pixel-perfect chrome  `[~]`
- [ ] C1 Authentic font: bake `FontBuilder` glyph metrics into a webfont/bitmap, or
      tune CSS to match (size table in 100-rsc-ui-map §3). **Biggest remaining gap.**
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
- [ ] D4 Equipment as authentic worn-slot layout (needs item ids/sprites, B-tier).

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
- [~] E2 **Shop** — BUILT (closed-path validated; open render pending a reachable
      shopkeeper, like E1). `/state.shop` block + `POST /shop {op:buy|sell|close}` →
      `Host.ShopBuy/Sell/Close`; `<ShopWindow>` modal (stock grid = buy w/ price,
      inventory grid = sell), real icons.
- [ ] E3 **Trade** — SPEC written → `specs/trade.md` (stateTrade + POST /trade →
      `Host.OfferTradeItems/ConfirmTrade/FinalizeTrade/DeclineTrade`; two-grid
      `<TradeWindow>`). Build needs a 2nd logged-in bot to verify.
- [ ] E4 **Duel** — SPEC written → `specs/duel.md` (stateDuel + rules + 2-stage
      accept; `<DuelWindow>`). Build needs a 2nd bot.
- [ ] E5 Window open/close lifecycle: `/state` should signal which window is open
      so the SPA can show/hide modals (add `window: {kind, ...}` to state).

### F. New ground  `[ ]`  (all SPEC'd under `specs/`)
- [ ] F1 **Minimap** — SPEC → `specs/minimap.md` (needs a new backend entity-list
      read; then a client-side rotating `<Minimap>` canvas, §4.4).
- [ ] F2 **Friends / ignore** — SPEC → `specs/friends.md`. ⚠ confirmed the real
      protocol gap (90-backlog §4): documents the missing outbound opcodes/Host
      methods + the `<FriendsTab>` UI. Needs backend protocol work first.
- [ ] F3 NPC dialog-option menu — SPEC → `specs/dialog-useon.md` (§5a).
- [ ] F4 Use-item-on-target drag UX — SPEC → `specs/dialog-useon.md`; routes to the
      `UseItemOn*` Host methods already on the dispatch interface (§5b).
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
