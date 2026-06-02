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

### B. Asset pipeline — **pixel-perfect enabler**  `[ ]`  ⚠ do early
- [ ] B1 `GET /sprite?kind=item|media|entity&id=N` → PNG, decoded from the RSC
      paletted sprite atlas the Go render layer already loads. (Re-run the
      rate-limited "sprite pipeline" investigation: find item-id→sprite-id map +
      where pixels are decoded in `render/`/`scene/`; design the encoder.)
- [ ] B2 Item-icon cache (id→PNG, ETag/immutable cache headers).
- [ ] B3 React `<ItemSprite id>` component; swap inventory text-stub → real icon.
- [ ] B4 UI chrome sprites (tab strip, borders, buttons) — optional; CSS may suffice.

### C. Pixel-perfect chrome  `[ ]`
- [ ] C1 Authentic font: bake `FontBuilder` glyph metrics into a webfont/bitmap, or
      tune CSS to match (size table in 100-rsc-ui-map §3).
- [ ] C2 RSC color palette as CSS vars (§3); recolor menu/chat/panels.
- [ ] C3 512×334 fixed-aspect game area + authentic right-tab strip & chat tabs.
- [ ] C4 Right-click menu styled to `drawMenuOptions` (gray box, white border,
      yellow hover/title).

### D. Read panels (cheap; data already in `/state`)  `[~]`
- [~] D1 Stats/skills tab (have a basic one; make pixel-perfect per §4.2).
- [ ] D2 Magic spellbook tab — needs `/state` spell list + `Host` cast wiring +
      `views_magic.go` exposure.
- [ ] D3 Prayer tab — `/state` prayer list + toggle action.
- [ ] D4 Equipment as authentic worn-slot layout (needs item ids/sprites, B-tier).

### E. Windows (server-driven; expose state + 2 actions each)  `[ ]`
- [ ] E1 **Bank** — add bank block to `/state` (or `/bank`) when open
      (`world/bank.go`), `POST /bank {op:deposit|withdraw,id,amount}` →
      `Host.BankDeposit/Withdraw/Close`; `<BankWindow>` grid (§4.5).
- [ ] E2 **Shop** — `world/shop.go` → `/shop` state + `POST /shop {buy|sell}` →
      `Host.ShopBuy/Sell`; `<ShopWindow>` w/ prices (§4.6).
- [ ] E3 **Trade** — `world/trade.go` state + offer/accept actions; live two-party
      `<TradeWindow>` (§4.7). Needs a second logged-in bot to test.
- [ ] E4 **Duel** — `world/duel.go` + `runtime/actions_duel.go`; `<DuelWindow>`
      w/ rule toggles (§4.8).
- [ ] E5 Window open/close lifecycle: `/state` should signal which window is open
      so the SPA can show/hide modals (add `window: {kind, ...}` to state).

### F. New ground  `[ ]`
- [ ] F1 **Minimap** — render top-down from world data (npc/player/item/scenery
      dots, §4.4). Either a `/minimap` PNG endpoint or draw client-side from a
      `/state` entity list (decide; client-side keeps it crisp + interactive).
- [ ] F2 **Friends / ignore** — ⚠ the real capability gap (90-backlog §4): confirm
      `Host` add/remove/list exist; add if missing, then `<FriendsTab>`.
- [ ] F3 NPC dialog-option (chat-menu) UI — when the server sends a menu of dialog
      choices, surface clickable options (90-backlog §5a).
- [ ] F4 Use-item-on-target drag UX (inv item → world/inv target), backed by the
      `UseItemOn*` Host methods already on the dispatch interface (90-backlog §5b).
- [ ] F5 Multi-bot tabs (90-backlog §6) — later.

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
