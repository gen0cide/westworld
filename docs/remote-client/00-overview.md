# Remote Client — Overview & North Star

**Status:** In progress (M1 build)
**Started:** 2026-06-01
**Branch:** `feat/remote-client`
**Owner doc set:** `docs/remote-client/`

This is the anchor document for turning `cmd/cradle -spectate` (a passive
browser viewport that streams rendered PNG frames + click-to-walk) into a
**full-fledged remote RuneScape Classic client** — one where you can right-click
anything in the world, get an authentic context menu, and act on it, all from a
browser, with no native window and no CGo.

Every workflow agent working on this effort should read this file first.

---

## 1. North star

> Open a browser, see the live game, **right-click any NPC / player / item /
> object / wall / tile and get the real RSC menu**, click an option, and the bot
> does it. Manage inventory, read and send chat. Eventually: bank, shop, trade,
> duel, skills, minimap — the whole client.

The display is the browser. The server is `cradle`, already logged into a real
RSC server as a `runtime.Host`. We stream rendered frames and expose a JSON API
for interaction.

---

## 2. The key insight: the protocol layer is already done

`runtime.Host` exposes the **complete, tested wire-action surface**. "Interact
with anything" is therefore a **UI + routing** problem, not a protocol problem.
We do **not** need to reverse-engineer packet opcodes — `Host` is the source of
truth for what each interaction sends.

Verified `Host` action methods (file refs are approximate, verify before use):

| Domain | Methods |
|---|---|
| Movement | `Walk`, `WalkTo`, `WalkToOpts` (runtime/host.go) |
| Combat | `AttackNpc`, `AttackNpcRanged`, `AttackPlayer`, `SetCombatStyle`, `Retreat`, `ChooseDialogOption` (runtime/combat.go) |
| Talk | `TalkToNpc` (runtime/combat.go), `NpcCommand` (runtime/boundary.go) |
| Items (ground) | `PickUpItem` (runtime/items.go) |
| Items (inv) | `DropItem`, `ItemCommand` (runtime/items.go), `EquipItem`, `UnequipItem` (runtime/magic.go) |
| Use-on | `UseItemOnItem`, `UseItemOnScenery`, `UseItemOnBoundary`, `UseItemOnNpc`, `UseItemOnPlayer`, `UseItemOnGroundItem` (runtime/boundary.go) |
| Objects | `InteractAt(x,y,option)`, `InteractWithBoundary(x,y,dir)` (runtime/boundary.go) |
| Magic/Prayer | `CastOnSelf/Npc/Player/Land/Inventory`, `ActivatePrayer`, `DeactivatePrayer` (runtime/magic.go, prayer.go) |
| Trade | `InitTradeRequest`, `AcceptIncomingTrade`, `DeclineTrade`, `OfferTradeItems`, `ConfirmTrade`, `FinalizeTrade` (runtime/trade.go) |
| Duel | `InitDuelRequest`, `AcceptIncomingDuel`, `DeclineDuel`, `OfferDuelItems`, `SetDuelRules`, `AcceptDuelOffer`, `AcceptDuelConfirm` (runtime/duel.go) |
| Shop | `ShopBuy`, `ShopSell`, `ShopClose` (runtime/shop.go) |
| Bank | `BankDeposit`, `BankWithdraw`, `WithdrawAll/X`, `DepositAll`, `BankClose` (runtime/bank.go) |
| Social | `PrivateMessage`, `AddFriend`, `RemoveFriend` (runtime/social.go), `Follow` (runtime/follow.go) |
| Chat | `Say`, `Command` (runtime/host.go) |
| Examine (no packet) | `ExamineNpc/Item/Scenery/Boundary/GroundItem/Player/InventorySlot/Self` (runtime/examine.go) |

**Menu verbs** come from `facts` defs: `SceneryDef.Command1/Command2`,
`BoundaryDef.Command1/Command2`, `NpcDef.Command1/Command2`, `ItemDef.Command`
(facts/defs.go). `facts.At(x,y) []Placement` enumerates scenery/boundary/spawns
on a tile. `runtime/examine.go` centralizes name + examine text per target.

---

## 3. What's missing: screen-space picking

`render.PickTile(land, view, px, py)` (render/pick.go) maps a click → terrain
tile and powers click-to-walk today. There is **no entity/object hit-testing**.

But the billboard projection inside `render.DrawEntitySprites`
(render/render.go:403–556) already computes each NPC/player/ground-item's
on-screen rectangle — it just discards it after blitting. The reusable
`projectPoint(cam, X, Y, Z, cx, cy)` helper (render/pick.go:78) does the
world→screen transform.

**Picking strategy (the central new render capability — see `10-hit-testing.md`):**

- **Billboards (NPC / player / self):** project foot point with the camera, derive
  the screen AABB exactly as `DrawEntitySprites` does
  (`screenW=(worldW<<viewDist)/camZ`, `screenH=(worldH<<viewDist)/camZ`,
  rect = `(sx-screenW/2, feetY-screenH)..(sx+screenW/2, feetY)`), test the click,
  return candidates **nearest-camera-first** (painter order reversed).
- **Ground items:** same, using the composited icon's world size.
- **Tile-grounded targets (scenery / boundary / terrain):** use `PickTile` to get
  the tile under the cursor, then enumerate via `facts.At(x,y)` + dynamic
  `world.Scenery`/`world.Boundaries` for scenery & wall objects; terrain itself is
  the always-present, lowest-priority "Walk here" target.

`render.Pick(...)` returns a single depth-ordered list of candidates, each
carrying enough identity (Kind, server Index / DefID / ItemID, world X/Y) to
build a menu and dispatch an action.

### Verified render view-model (render/render.go, render/entity.go)

- `View{X,Y,Plane,Rotation,Zoom,W,H,AnimFrame, Entities[], Self*..., GroundItems[], DynamicScenery[], BoundaryRemoved/SceneryRemoved funcs}`
- `Entity{X,Y,Kind,NpcID,Heading,EquipSprites,...,Index,OffX,OffZ,StepPhase,Moving}` — `.Index` IS the server NPC/player index.
- `GroundItemMarker{X,Y,ItemID}`, `DynamicSceneryItem{X,Y,ID,Direction}`
- Camera: `SetCamera(localX,-elev,localZ, pitch=912, yaw=rotation*4, roll=0, dist=zoom*2)`; RenderTo swaps `CameraYaw`/`CameraRoll` before projecting (replicate in any picker).

---

## 4. Architecture (4 layers)

```
   Browser (SPA: viewport + right-click menu + inventory + equipment + chat)
        │  HTTP/JSON (poll frame + state; POST pick / menu-action / chat)
        ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Layer 3  HTTP API           cmd/cradle/remoteclient.go (new)   │
   │   GET  /frame /pos /state   POST /pick /act /chat ...          │
   └──────────────────────────────────────────────────────────────┘
        │ calls
        ▼
   ┌──────────────────────────────────────────────────────────────┐
   │ Layer 2  Menu + dispatch    (new package, e.g. remoteui/)      │
   │   target taxonomy → option list (authentic verbs)              │
   │   {target, option} → exact runtime.Host method call            │
   └──────────────────────────────────────────────────────────────┘
        │ uses                              │ uses
        ▼                                   ▼
   ┌─────────────────────────┐   ┌──────────────────────────────────┐
   │ Layer 1  render.Pick     │   │  runtime.Host (DONE — §2)         │
   │  screen (px,py) →        │   │  + facts defs + examine (§2)      │
   │  depth-ordered targets   │   └──────────────────────────────────┘
   └─────────────────────────┘
```

Data flow for a right-click:

1. Browser POSTs `/pick {px,py,rot,zoom,w,h}`.
2. Handler calls `render.Pick` → ordered candidates; Layer 2 builds each
   candidate's authentic option list + examine text → returns menu JSON.
3. Browser renders the context menu at the cursor.
4. User clicks an option → browser POSTs `/act {targetRef, option}`.
5. Layer 2 maps `{target, option}` → the exact `Host` method + args, invokes it
   (serialized through a worker like the existing `/walk` queue).

Left-click = the **top** menu option (RSC default-action semantics).

---

## 5. Milestones

### M1 (this build) — world interaction + inventory + chat

- **Layer 1** `render.Pick` hit-testing (+ unit tests on the projection math).
- **Layer 2** menu model + dispatch for: NPC, player, self, ground item, scenery
  object, wall/boundary, terrain tile, **inventory item**.
- **Layer 3** endpoints: `/pick`, `/act`, `/state` (inventory + equipment + stats +
  recent chat), `/chat` (send public / command / PM), plus existing
  `/frame /pos /walk /shot /clip`.
- **Browser UI**: viewport with right-click context menu + left-click default
  action; inventory grid (right-click: item command / Use / Drop / Examine,
  drag-to-use-on optional); equipment summary; chat box with scrollback + input.
- **Gate**: `go build ./...`, `go vet ./...`, `go test ./render/... ./runtime/...`.

### M2 (documented backlog — see `90-backlog.md`)

Bank window, shop window, trade window, duel window, skills/stats tab, minimap,
friends/ignore lists, dialog-option (NPC chat) UI, settings, use-item-on-target
drag UX, multi-bot tabs. Architecture must keep these cheap to add (the dispatch
table + `/state` are designed to extend).

---

## 6. Conventions & constraints

- Pure stdlib `net/http`; no new heavy deps. Browser is the only client.
- Reuse `buildLiveView` / `motionTracker` from `cmd/cradle/spectate.go`; do **not**
  break the existing `-spectate` mode.
- All actions serialized through a single worker goroutine (RSC actions override
  each other; the existing `/walk` queue is the pattern).
- World reads are lock-safe (every accessor RLocks + copies) — safe from HTTP
  goroutines.
- Do **not** touch the uncommitted `reference/rsc-client/**` deob edits; this work
  is Go (`render/`, `runtime/`, `cmd/cradle/`, new package) + `docs/remote-client/`.
- New flag: `-client` (full client) alongside `-spectate`, sharing helpers — or
  evolve `-spectate`; the spec doc decides and documents the choice.

---

## 7. Doc set (filled in by the build workflow)

- `00-overview.md` — this file (north star).
- `10-hit-testing.md` — Layer 1 design (render.Pick API, projection, tests).
- `20-menu-dispatch.md` — Layer 2 design (target taxonomy, option lists, dispatch table).
- `30-http-api.md` — Layer 3 design (endpoints, JSON request/response schemas).
- `40-browser-ui.md` — SPA design (components, interaction model, polling).
- `50-impl-spec.md` — reconciled per-file implementation spec (the cross-layer contract).
- `90-backlog.md` — M2+ forward plan.
- `99-build-log.md` — what the workflow actually built + verification results.
