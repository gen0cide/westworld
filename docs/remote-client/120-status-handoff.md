# Remote Client — Status & Session Handoff

**Last updated:** 2026-06-02. **Branch:** `deob/rsc-client-reference` (shared with a
concurrent deob effort that owns `reference/` + parts of `render/`/`runtime/` —
**never edit `reference/`**). This doc is the single entry point for continuing
the cradle remote-client work in a fresh session.

The goal: a **full-fledged, pixel-perfect-ish React reimplementation of the
RuneScape Classic UI**, driven by a live OpenRSC bot (`cmd/cradle -client`).

---

## 1. Where we are

The cradle bot exposes an HTTP/JSON API + frame stream; a **Vite + React + TS**
SPA (in `web/`, `go:embed`'d into the binary) renders the world and every panel.
The Go backend (`world/`, `runtime/`) already implements the game systems — most
features = *expose existing state over HTTP + draw the panel*.

### Feature status

| Feature | State |
|---|---|
| Viewport (streamed frame, pick→act, right-click menus) | ✅ live |
| Inventory / Equipment / Chat (say/`::`cmd/`@`pm) | ✅ live |
| **Item sprites** (`GET /sprite`, real RSC icons) | ✅ live |
| Stats/skills · Magic spellbook (`/spells`,`/cast`) · Prayer (`/prayer`) tabs | ✅ live |
| RSC tab-strip (Inv/Worn/Stats/Magic/Pray/Friends) | ✅ live |
| **Bank** (`/state.bank` + `/bank`) | ✅ live-verified (`screens/bank-open.png`) |
| **Shop** (`/state.shop` + `/shop`, via Talk-to→`/dialog`) | ✅ live-verified (`screens/shop-open.png`) |
| **Minimap** (`/state.entities` + `<Minimap>` canvas) | ✅ live-verified (`screens/minimap.png`) |
| **Trade** (`/state.trade` + `/trade` + `<TradeWindow>`) | ⚠ BUILT, not live-verified (needs 2-bot accept handshake) |
| **Duel** (`/state.duel` + `/duel` + `<DuelWindow>`) | ⚠ BUILT, not live-verified (same) |
| NPC dialog | ⚠ partial: `POST /dialog` + Talk-to land; `<NpcDialog>` option UI pending |
| Friends/ignore | ❌ spec only — **the real protocol gap** (`specs/friends.md`) |
| Pixel-perfect font (C1) · sprite tab icons (B4) | ❌ spec/CSS-approx only |

---

## 2. Architecture & directories (the 4 layers)

- **Layer 1 — hit-testing:** `render/pick.go`, `render/hittest.go`.
- **Layer 2 — menu/dispatch:** `remoteclient/` — `menu.go` (`BuildMenu`; holds the
  synthetic **Talk-to** verb for command-less NPCs), `target.go`, `dispatch.go`.
- **Layer 3 — HTTP API:** `cmd/cradle/remoteclient.go` (`serveClient`: `/state`,
  all action endpoints, the one serialized `enqueueAction` worker), plus
  `sprites.go` (`/sprite`), `magic.go` (`/spells`,`/cast`), `clientpage.go`
  (legacy client at `/legacy`), `spectate.go`, `main.go`.
- **Layer 4 — SPA:** `web/` — Vite+React+TS. `web/embed.go` (`//go:embed all:dist`).
  - `src/api.ts` (typed client), `src/types.ts` (wire types), `src/state.ts`
    (`useConfig`/`useGameState`), `src/ui.tsx` (`UIProvider`/`useUI`: flash +
    `openMenu`/`openActions` context menu), `src/App.tsx`, `src/styles.css`.
  - `src/components/`: `Viewport, Hud, ChatPanel, SidePanel, InventoryGrid,
    EquipmentPanel, StatsPanel, Spellbook, PrayerTab, BankWindow, ShopWindow,
    TradeWindow, DuelWindow, Minimap, ItemSprite`.
- **Backend (do NOT edit — methods already exist):** `world/{bank,shop,trade,duel}.go`
  (state mirrors), `runtime/{bank,shop,trade,duel,magic,prayer}.go` + `actions_*.go`
  + `views_*.go` (Host methods). `runtime/` is also touched by the concurrent deob
  agent — leave it alone; if a Host method is missing, note it, don't add it.

### Sprite resolution (changed 2026-06-02)
Item icons resolve from OpenRSC's **`Authentic_Sprites.orsc`** at
`spriteItem + ItemDef.AppearanceID` (`render/itemsprite.go`, archive loaded by
`render/sprites.go::sprites()`). The path comes from env var
**`WESTWORLD_SPRITES_ORSC`** (the older `WESTWORLD_{MEDIA,CONFIG,ENTITY}_JAG`
vars are **obsolete**). On this box:
`/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc`.

### HTTP surface
`GET /` (SPA) · `/legacy` · `/config` · `/pos` · `/frame` · `/state` · `/spells` ·
`/sprite?kind=item&id=N` · `/examine` · `/shot` · `/clip`.
`POST /walk /act /pick /chat /bank /shop /cast /prayer /trade /duel /dialog`.
`/state` carries `self, inventory, equipment, chat, magic, prayers, entities`
always, plus `bank|shop|trade|duel` only while that window is open.

---

## 3. How to run / build

```bash
# build
go build -o /tmp/cradle ./cmd/cradle/

# run a bot + serve the client (needs the sprites archive for item icons)
WESTWORLD_SPRITES_ORSC=/home/free/code/rsc-hacking/openrsc/Client_Base/Cache/video/Authentic_Sprites.orsc \
  /tmp/cradle -server localhost:43594 -username webreact -password reactpass1 \
  -facts /home/free/code/rsc-hacking/openrsc -client -client-addr localhost:8090
# (nohup ... & to detach; poll GET /pos until x!=0; open http://localhost:8090/)

# frontend: edit web/src/**, then BEFORE go build:
npm --prefix web run build        # tsc + vite -> web/dist (go:embed'd)
# dev HMR: cd web && npm run dev   # vite :5173 proxies API -> :8090

# register a new account:
go run ./cmd/register -server localhost:43594 -username <u> -password <p>
```

**Gate (must stay green):** `go build ./...` · `go vet ./cmd/cradle/ ./render/
./remoteclient/` · `go test ./render/... ./remoteclient/... ./runtime/...` ·
`npm --prefix web run build`. **Always `npm run build` before `go build`** or the
binary serves stale assets (`web/dist` is committed).

---

## 4. Game-access recipe (to live-test windows)

The OpenRSC server (`:43594`) DB is **SQLite**: `openrsc/server/inc/sqlite/westworld.db`
(the `docker-compose.yml` mariadb is a dead artifact). Admin (`::`) commands need
the account in group OWNER(0)/ADMIN(1)/MOD(3).

```bash
# grant OWNER (do it while LOGGED OUT — a live session overwrites DB on save):
#  1) kill the bot (SIGTERM -> graceful logout)   2) then:
sqlite3 openrsc/server/inc/sqlite/westworld.db \
  "UPDATE players SET group_id=0 WHERE username='Webreact';"
#  3) relaunch -> logs in as OWNER (also bypasses the tutorial-island action gate)
```
Useful commands (via `POST /chat {"kind":"command","text":"bank"}`): `::bank`,
`::fillbank`, `::teleport varrock`, `::spawnnpc 51` (shopkeeper), `::summon <Name>`.
**`webreact` and `webreact2` are already OWNER.** A 2nd bot for trade/duel:
`webreact2`/`reactpass2` on `-client-addr localhost:8091`. **Gotcha:** server login
**code 4 = "already logged in"** — wait ~60s after an abrupt kill before relaunching
the same account.

---

## 5. Subagent / workflow approach (how we work here)

Use the **Workflow tool** (the user opts in with the word "workflow"). Pattern that
has worked:
- **Design/recon (parallel or single, Opus):** read-only or bot-driving; each writes
  to its OWN file (`docs/remote-client/specs/<feat>.md` or `screens/`). Recon (Opus)
  drives the live bot through screens and captures real `/state` + screenshots — its
  findings drive the build wave.
- **Build (Sonnet, from frozen specs):** implement one feature per agent. **Serialize
  any agents that edit the SHARED files** (`cmd/cradle/remoteclient.go`,
  `web/src/{types.ts,api.ts,App.tsx,styles.css}`) — parallel edits corrupt them. New
  component files are conflict-free. Each agent runs the gate and **self-reverts** if
  it can't go green (so the tree stays green for the next).
- **Verify (Sonnet):** full gate + live CDP screenshots.

**Standing permissions/constraints for subagents:**
- **Sandbox bypass is granted** — set Bash `dangerouslyDisableSandbox: true` for
  network/port/docker/CDP commands (the user authorized this).
- ONLY edit `web/`, `cmd/cradle/`, `docs/remote-client/` (and `remoteclient/menu.go`
  when a menu change is genuinely needed). **Never `reference/`; never `world/` or
  `runtime/`.** Never `git commit` (the main session reviews + commits).

### Screenshots (headless)
`chromium --headless=new --screenshot=out.png --virtual-time-budget=6000 URL` works
for the default view. To click a side-panel tab / open a window, use the **CDP
recipe** (sandbox OFF): `nohup chromium --headless=new --no-sandbox
--remote-debugging-port=9222 --user-data-dir=/tmp/cdp-X URL &`, then a Node ESM
script (global `fetch`+`WebSocket`) → `/json` → `Page.navigate` → `Runtime.evaluate`
to `.click()` the `button[title=...]` → `Page.captureScreenshot`. (In-sandbox the
debug port fails with exit 144.)

---

## 6. Remaining work (recommended order)

1. **Live-verify Trade + Duel (E3/E4):** drive the two OWNER bots through a real
   accept handshake (bot A offers via `/trade`/`/duel` → bot B accepts) until the
   window reaches phase `open`/`confirm`; screenshot both sides. Components +
   endpoints already exist — this is verification + any payload fixes.
2. **NPC dialog UI (F3):** add a `/state` dialog-options block (the data is in
   `world.recent` `DialogOptionsRecord`) + an `<NpcDialog>` choice list posting to
   the existing `POST /dialog`. (`specs/dialog-useon.md`)
3. **Use-item-on-target drag (F4):** drag inv item → target, routing to the existing
   `UseItemOn*` Host methods. (`specs/dialog-useon.md`)
4. **Friends/ignore (F2):** the real protocol gap — needs new outbound opcodes/Host
   methods first (`specs/friends.md`), then `<FriendsTab>`.
5. **Pixel-perfect chrome:** bitmap font (C1, `specs/font.md`), sprite tab-strip
   icons (B4 — would need `/sprite?kind=media`), equipment worn-slot icons (D4).
6. **Minimap polish:** friend (green) dots, static scenery, right-click verb menu,
   compass.
7. **Cross-cutting:** SSE/WebSocket push for low-latency windows (G1); a
   `go generate` that runs `npm run build` before `go build` (G2).

---

## 7. Doc index (`docs/remote-client/`)

`00-overview` · `10-hit-testing` · `20-menu-dispatch` · `30-http-api` ·
`40-browser-ui` · `50-impl-spec` (frozen cross-layer contract) · `90-backlog` ·
`99-build-log` (chronological record — read §6–§9 for the React era) ·
`100-rsc-ui-map` (authentic layouts/colors/sprites) ·
**`110-react-port` (the live TASK TREE with statuses)** ·
`120-status-handoff` (this file) · `specs/*.md` (frozen per-feature impl specs) ·
`screens/*.png` + `screens/RECON.md` (live captures + the access recon).
