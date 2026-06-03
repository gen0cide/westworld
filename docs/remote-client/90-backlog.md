# Remote Client — M2+ Forward Plan & M1 Extensibility Pressure-Test

**Status:** Planning (written during M1 build)
**Reads:** `00-overview.md` §5 (milestone framing), §2 (Host method table), §3 (render
view-model). Layer docs `10/20/30/40/50` are **not yet written** as of this doc; this
backlog is therefore grounded directly in the verified source (file:line refs below),
and the "M1 contracts" it pressure-tests are the contracts the overview *prescribes*
for those docs. When 10–50 land, re-confirm the four contract invariants in §0 against
their concrete shapes — they are written to be satisfiable, and §8 flags the few
decisions that would break them if M1 gets them wrong.

---

## 0. The four M1 contracts every M2 feature rides on

The whole point of this doc: M2 must be **additive**, never a refactor. That holds iff
M1 freezes these four contracts with the right shape. Each §1–§5 feature is checked
against all four; §8 collects the decisions M1 must get right *now*.

1. **`MenuTarget` is a tagged union (sum type), not a fixed struct.** A target carries
   `Kind` + the identity its kind needs (`Index` for npc/player, `DefID`+`X`+`Y`+`Direction`
   for scenery/boundary, `ItemID`+slot for inventory, `X`+`Y`+`ItemID` for ground item,
   `X`+`Y` for tile). Adding a `Kind` (e.g. `bank_slot`, `shop_slot`, `trade_slot`,
   `spell`, `prayer`, `friend`, `dialog_option`, `minimap`) must require **zero changes
   to existing targets**. → New `Kind` const + new identity fields (proto3-style: only
   read the fields your Kind defines).

2. **The dispatch table is `map[ (Kind, Option) ] -> func(target, args) error`** (or a
   `switch` with the same shape). Each entry calls exactly one `Host` method. Adding a
   feature = adding rows, never editing rows. The serialization worker (§7) is shared by
   all rows.

3. **`/state` is one flat JSON document of *optional, nil-when-closed* sub-objects**, each
   mirroring exactly one `world.*Record` snapshot. `bank`, `shop`, `trade`, `duel`,
   `dialog` are **absent / null** until their window opens. Adding a window = adding one
   top-level key. The browser renders a panel iff its key is non-null. → §6 fixes the
   exact M1 envelope so M2 only appends keys.

4. **Every action is `POST /act {target, option, args}` through ONE worker goroutine.**
   New endpoints are the exception (only when payload shape differs hard, e.g. `/chat`,
   `/offer`); the default is "another `(Kind,Option)` row + reuse `/act`". The worker is
   the existing `/walk` channel-queue pattern (`spectate.go:293-307`), generalized to a
   `func() error` queue.

**Verdict up front:** all twelve deferred features satisfy contracts 1–4 with **zero
refactor**, *provided* M1 ships the five decisions in §8. The world read-layer is already
far more complete than M1 needs — every M2 window has a `*Record` snapshot accessor today
(verified below). The only true *missing capability* in the whole backlog is the
**friends/ignore roster read-side** (write-side exists; the roster packets aren't decoded
— `views_world.go:365`), which is a `world/` + `event/` decode task, not a UI-layer one.

---

## 1. Trade / window-style interactions (bank, shop, trade, duel)

These four are one architectural family: a modal window backed by a `world.*Record`
snapshot + a small set of `Host` action methods. They are the strongest evidence the
architecture extends, because **the read-side and write-side already exist in full** —
M2 is purely Layer-2 dispatch rows + Layer-3 `/state` keys + Layer-4 panels.

### 1a. Bank window

- **Host methods (overview §2 "Bank"):** `BankDeposit(ctx, catalogID, amount)`,
  `BankWithdraw(ctx, catalogID, amount)`, `WithdrawAll(ctx, catalogID)`,
  `WithdrawX(ctx, catalogID, amount)`, `DepositAll(ctx, keepIDs map[int]struct{})`,
  `BankClose(ctx)` — all in `runtime/bank.go:12-59`.
- **World state (read-side, exists):** `world.BankState.Bank() *BankRecord`
  (`world/bank.go:35`) → `BankRecord{MaxSize, Slots []BankSlot{ItemID,Amount}, Opened/UpdatedAt}`
  (`world/bank.go:17-29`). Lock-safe snapshot copy. `nil` when closed — drives contract 3.
- **`/state` field:** `"bank": null | {maxSize, slots:[{itemID,amount,name}]}` (name resolved
  via `facts.ItemDef`). New top-level key only.
- **Endpoint/UI:** no new endpoint — bank deposit/withdraw are `(Kind=bank_slot|inv_slot,
  Option=Deposit/Withdraw/Withdraw-X/Withdraw-All)` dispatch rows through `/act`. New panel:
  two-column bank-grid + inventory-grid (reuse M1's inventory grid component), close button →
  `(bank, Close)`.
- **Contracts:** 1 ✓ (`bank_slot` Kind, identity = catalogID); 2 ✓ (5 rows); 3 ✓ (`bank` key);
  4 ✓ (`/act`). **No blocker.**

### 1b. Shop window

- **Host methods (§2 "Shop"):** `ShopBuy(ctx, catalogID, amount)`, `ShopSell(ctx, catalogID,
  amount)`, `ShopClose(ctx)` (`runtime/shop.go:23-37`).
- **World state (exists):** `world.ShopState.Shop() *ShopRecord` (`world/shop.go:67`) →
  `ShopRecord{IsGeneral, SellPriceMod, BuyPriceMod, PriceMultiplier, Slots []ShopSlot{ItemID,
  Stock,BaseStock}, ...}` (`world/shop.go:25-61`). Price helpers already computed server-side:
  `BuyPrice(itemID, basePrice)` / `SellPrice(...)` (`world/shop.go:162,183`) — the panel shows
  authentic prices with **no client math**. `IsOpen()` (`shop.go:79`) gates the panel.
- **`/state` field:** `"shop": null | {isGeneral, slots:[{itemID,name,stock,buyPrice,sellPrice}]}`
  (call `BuyPrice`/`SellPrice` server-side when building `/state`).
- **Endpoint/UI:** `(shop_slot, Buy)` / `(inv_slot, Sell)` rows via `/act`; panel = shop grid +
  inv grid + close. No new endpoint.
- **Contracts:** all ✓. **No blocker.**

### 1c. Trade window

- **Host methods (§2 "Trade"):** `InitTradeRequest(ctx, serverIndex)`, `AcceptIncomingTrade(ctx,
  fromIndex)`, `DeclineTrade`, `OfferTradeItems(ctx, []world.TradeItem)`, `ConfirmTrade`,
  `FinalizeTrade` (`runtime/trade.go:14-105`).
- **World state (exists):** `world.TradeState.Trade() *TradeRecord` (`world/trade.go:61`) →
  `TradeRecord{WithIndex, WithName, MyOffer/TheirOffer []TradeItem, My/TheirFirstAccepted,
  My/TheirSecondAccepted, Phase ("request_sent"|"open"|"confirm"|"completed"|"cancelled"), ...}`
  (`world/trade.go:19-47`). `IsActive()` gates the panel; `Phase` drives which buttons enable.
- **`/state` field:** `"trade": null | {withName, phase, myOffer:[...], theirOffer:[...],
  myAccept, theirAccept, ...}`.
- **Endpoints/UI:** `InitTradeRequest` is a **world-target** action — it rides the player
  right-click menu: `(player, Trade)` row → `host.InitTradeRequest(ctx, target.Index)`. The
  *offer* mutates a multi-item basket, so it needs a small dedicated payload: `POST /offer
  {kind:"trade", items:[{itemID,amount}]}` → `OfferTradeItems`. Accept/Confirm/Decline are
  `(trade, Accept|Confirm|Decline)` rows on `/act`. Panel = two offer grids + my-inventory grid
  (click to stage) + Accept/Decline.
- **Contracts:** 1 ✓ (`Trade` is a normal player-menu verb; `trade_slot` for the basket); 2 ✓;
  3 ✓; 4 — the **basket offer is the one place a bespoke endpoint is justified** (`/offer`), and
  M1's §8.2 decision (allow a small family of typed-payload POSTs beside `/act`) covers it. The
  same `/offer {kind:"duel",...}` serves §1d. **No blocker** *if* §8.2 holds.

### 1d. Duel window

- **Host methods (§2 "Duel"):** `InitDuelRequest`, `AcceptIncomingDuel`, `DeclineDuel`,
  `OfferDuelItems(ctx, []world.TradeItem)`, `SetDuelRules(ctx, world.DuelRules)`,
  `AcceptDuelOffer`, `AcceptDuelConfirm` (`runtime/duel.go:14-102`).
- **World state (exists):** `world.DuelState.Duel() *DuelRecord` (`world/duel.go:61`) →
  `DuelRecord{WithName, MyOffer/TheirOffer, accept flags, Rules DuelRules{DisallowRetreat/
  Magic/Prayer/Weapons}, Phase, ...}` (`world/duel.go:21-55`).
- **`/state` field:** `"duel": null | {withName, phase, myOffer, theirOffer, rules:{...}, accepts}`.
- **Endpoints/UI:** `(player, Duel)` → `InitDuelRequest(target.Index)`; basket via the shared
  `/offer {kind:"duel",...}`; **rule toggles** are a 4-checkbox sub-panel → `(duel, SetRules)`
  row carrying the four booleans in `args` → `SetDuelRules(world.DuelRules{...})`; accept buttons
  are `(duel, AcceptOffer|AcceptConfirm|Decline)` rows.
- **Contracts:** all ✓ — the rules toggle is the one case where `/act`'s `args` map earns its
  keep (contract 1's "args" field). **No blocker** *if* §8.3 (typed `args` map on `/act`) holds.

---

## 2. Skills / stats tab, magic tab, prayer tab

Three read-mostly tabs sharing one pattern: a `world.Self` / `facts` snapshot rendered as a
list, where a few entries are *clickable actions*.

- **Skills/stats:** **read-only.** Backed by `world.Self.SkillLevel(id)`, `SkillMax(id)`,
  `SkillXP(id)`, `SetAllSkills`/`NumSkills`, `QuestPoints()`, `CombatLevel()`, `Fatigue()`,
  `HP()/MaxHP()` (`world/self.go:162-379`) and the `selfView/skillsView/skillView` shapes
  (`runtime/views_self.go:166-311`). No Host action, no packet. `/state` field: `"skills":
  [{id,name,cur,max,xp}], "questPoints", "combatLevel", "fatigue"`. Pure additive panel.
- **Magic tab:** read `world.Self`/`facts.SpellDef` via `magicView`/`spellsView`/`spellDefView`
  (`runtime/views_magic.go`), `spellCanCastCallable`/`spellHasRunesCallable` for enable/disable.
  Actions (overview §2 "Magic/Prayer"): `CastOnSelf/Npc/Player/Land/Inventory`. Critically, casting
  is a **two-step world interaction**, not a tab-internal click: select spell in tab → it becomes a
  "pending cast" cursor → next world/inventory pick dispatches `CastOnNpc(spellID, target.Index)`
  etc. This is the **same deferred-target state machine as use-item-on** (§5) — build them on one
  mechanism. `/state` field: `"spells":[{id,name,level,canCast}], "pendingCast": null|spellID`.
- **Prayer tab:** `prayerView`/`prayersView`/`prayerActiveCallable` (`runtime/views_prayer.go`),
  `world.Self.ActivePrayers()`/`PrayerActive(idx)` (`world/self.go:282-310`). Actions:
  `ActivatePrayer`/`DeactivatePrayer` — pure toggles, `(prayer, Toggle)` rows on `/act`.
  `/state` field: `"prayers":[{id,name,level,active}]`.
- **Contracts:** 1 ✓ (`spell`/`prayer` Kinds; skills carries no target); 2 ✓; 3 ✓ (three keys);
  4 ✓. **No blocker** — but see §8.5: magic's pending-cast is a hard dependency on the shared
  deferred-target mechanism, so M1 should reserve `pendingTarget` in the picker contract even if
  M1 only uses it for use-item-on.

---

## 3. Minimap

- **Host methods:** none new — left-click-to-walk on the minimap reuses M1's `/walk` (the worker
  already exists, `spectate.go:293`). Right-click a minimap dot = the same player/npc menu as the
  3D view, so it reuses `/act`.
- **World state (exists, no new decode):** everything a minimap dot needs is already perceived —
  `world.Npcs.All()` (`NpcRecord{Index,X,Y,TypeID,Heading}`), `world.Players.All()`
  (`PlayerRecord{Index,Name,X,Y,CombatLevel,SkullType}`), `world.GroundItems.All()`,
  `world.Self.Position()`/`Heading()`. The compass is `view.Rotation` (already in the View model,
  `render/render.go:30`). Plane/landscape walkability for the map backdrop is `pathfind.Landscape`
  (already loaded in `spectate.go:271`).
- **`/state` (or a dedicated `/minimap`):** the dot list is large and changes every frame, so it
  belongs on the **frame-rate poll**, not the slower `/state` poll. Cleanest: a `GET /minimap`
  returning `{self:{x,y,heading}, rotation, npcs:[{x,y,index,hostile?}], players:[{x,y,index,name}],
  items:[{x,y}]}` polled at the same cadence as `/frame`. This is the **one new GET that is genuinely
  separate from `/state`** — justified because its update frequency differs by ~10x (contract 3 is
  about *windows*, which are slow; the minimap is fast like `/frame`).
- **UI:** a fixed-corner canvas; render dots from the poll; click → `/walk` (left) or `/pick`+menu
  (right). Optionally render the map backdrop server-side as a small PNG via a `render.RenderMinimap`
  (future render capability) to avoid shipping landscape to the browser.
- **Contracts:** 1 ✓ (`minimap` is just a coordinate source feeding existing `tile`/`player`/`npc`
  targets); 2 ✓ (reuses rows); 3 — minimap intentionally bypasses `/state` for cadence reasons (a
  documented exception, not a violation); 4 ✓. **No blocker.** §8.4: M1 must NOT assume `/state` is
  the *only* polled JSON doc — leave room for a second fast poll.

---

## 4. Friends / ignore lists  ⚠ THE ONE REAL CAPABILITY GAP

- **Host methods (§2 "Social"):** `AddFriend(ctx, name)`, `RemoveFriend(ctx, name)`,
  `PrivateMessage(ctx, recipient, message)` (`runtime/social.go:12-25`). Write-side is **complete**.
- **World state: MISSING.** There is **no decoded friend/ignore roster** in `world/`. The code
  itself flags this: `views_world.go:365` — "we send AddFriend ... (currently we send AddFriend but
  don't decode the friend-list packets)". So the browser can *add/remove/PM* but cannot *display the
  list*, online status, or unread PMs as a roster.
- **What M2 needs:** a `world/social.go` `SocialState{Friends []FriendRecord{Name,World,Online},
  Ignores []string}` fed by a new `event/` decoder for the friend-list / friend-status /
  ignore-list server packets, plus `Host.Social() *SocialRecord` snapshot accessor. PM *history*
  can ride the existing `RecentEvents.PM()` (`world/recent.go:122`) ring — incoming PMs are already
  captured (`PMRecord{Sender,Message,At}`), they're just not aggregated per-friend.
- **`/state` field (after the decode lands):** `"social": {friends:[{name,online}], ignores:[name],
  pms:[{sender,message,at}]}`. Until then the panel can be **send-only** (add/remove/PM via `/act`
  rows + a PM compose box), which is shippable without the decode.
- **Endpoint/UI:** `(player, Add-Friend|Remove-Friend|Send-Message)` rows reuse `/act`; PM compose
  reuses `/chat` with a `recipient` field (M1's `/chat` already carries public/command/PM modes per
  overview §5). Roster panel needs the decode.
- **Contracts:** 1 ✓ (`friend` Kind, identity = name); 2 ✓; 3 ✓ (`social` key); 4 ✓. **No M1
  architectural blocker** — the gap is a *data-decode* task in `world/`+`event/`, entirely outside
  the four UI-layer contracts. **Flag for M2 scoping, not for M1 design.** Recommend: M1's `/chat`
  must accept a `recipient`/`mode` param (it already plans to per §5 overview) so PM works day-one.

---

## 5. NPC dialog-option (chat-menu) UI + use-item-on drag UX

### 5a. NPC dialog-option menu

- **Host method (§2 "Combat/Talk"):** `ChooseDialogOption(ctx, index)` (`runtime/combat.go:222`),
  reached after `TalkToNpc` (`runtime/combat.go`).
- **World state (exists, end-to-end):** `world.RecentEvents.DialogOptions() *DialogOptionsRecord`
  (`world/recent.go:226`) → `{Options []string, At}`; `DialogText() *DialogTextRecord` for the NPC
  speech bubble (`world/recent.go:212`); `SetDialogOptions`/`ClearDialogOptions` already manage the
  lifecycle (`recent.go:241,251`). The whole talk→options→answer loop the DSL routines use is live.
- **`/state` field:** `"dialog": null | {text, options:[string]}` (null when no menu open). Nulls per
  contract 3.
- **Endpoint/UI:** picking option N is `(dialog_option, Choose)` with `args:{index:N}` → wait —
  cleaner: a thin `POST /dialog {index}` is fine too, but the `(Kind=dialog_option, index)` row keeps
  it on `/act` and within contract 2. After answering, the browser optimistically expects the server
  to push a new `dialog` (or null) on the next `/state`. UI: a modal text box + clickable option list,
  exactly the in-game widget.
- **Contracts:** 1 ✓ (`dialog_option` Kind, identity = index); 2 ✓; 3 ✓; 4 ✓. **No blocker.** The
  read-side maturity here is the second-strongest extensibility proof after §1.

### 5b. Use-item-on-target drag UX

- **Host methods (§2 "Use-on"):** `UseItemOnItem`, `UseItemOnScenery`, `UseItemOnBoundary`,
  `UseItemOnNpc`, `UseItemOnPlayer`, `UseItemOnGroundItem` (`runtime/boundary.go`).
- **The mechanism — a deferred-target state machine** (this is the architecturally important one):
  RSC "use X on Y" is two picks. M1 lists this as *optional* drag UX, but the **deferred-target
  state must be a first-class part of the picker contract** because **magic casting (§2) reuses it
  exactly**. Design: a server-side (or client-side) `pendingUse {sourceKind:"inv_slot"|"spell",
  sourceID}` set when the user picks "Use" on an inventory item (or selects a spell). The *next*
  world/inventory pick, instead of opening a menu, dispatches `UseItemOnNpc(sourceSlot, target.Index)`
  / `CastOnNpc(spellID, target.Index)` / etc. based on the second target's Kind.
- **`/state` / endpoint:** `"pendingTarget": null | {kind:"use_item"|"cast_spell", id}` so the UI can
  draw the "use-with" cursor and any client can observe the pending state. The second pick is a normal
  `/pick` that the dispatch layer routes to a `UseItemOn*`/`CastOn*` based on `pendingTarget`.
- **Contracts:** 1 — needs the picker/`MenuTarget` to *carry a pending-source mode*, see §8.5; 2 ✓
  (six `UseItemOn*` + five `CastOn*` rows keyed by `(pendingSourceKind, secondTargetKind)`); 3 ✓
  (`pendingTarget` key); 4 ✓. **Potential blocker → §8.5: M1 must reserve `pendingTarget` even if it
  ships only inventory's `Use`.**

---

## 6. Multi-bot tabs

- **Host methods:** none — this is a *fan-out* of the entire existing surface across N `*runtime.Host`
  instances. `cmd/cradle` already constructs one host; multi-bot = a `map[botID]*Host` + a session
  selector.
- **World state:** each `Host` owns its own `world.World` (lock-safe). No sharing concerns.
- **`/state` / endpoints:** the cheapest forward-compatible move is to **scope every endpoint by a
  `bot` query param / path segment from day one**: `/state?bot=ID`, `/frame?bot=ID`,
  `POST /act?bot=ID`. In M1 there's one bot so `bot` defaults to the sole host — but **baking the
  param into the M1 endpoint signatures costs nothing and avoids a path-scheme refactor in M2.** The
  per-bot worker goroutine (§7) is keyed by `bot`.
- **UI:** a tab strip; each tab is an independent SPA instance bound to a `bot` id.
- **Contracts:** 1/2 ✓ (targets/dispatch are per-host already); 3 ✓ (each `/state` is one host's doc);
  4 — the worker must become **one queue per bot** (trivial: a `map[botID]chan func()`). **Soft
  blocker → §8.1: M1 should route `/walk`+`/act` through a worker registry keyed by an (M1-implicit)
  bot id, and accept an optional `bot` param, so M2 multi-bot is `len(map)>1` instead of a rewrite.**

---

## 7. The shared serialization worker (cross-cutting)

Every action in §1–§6 funnels through **one worker per bot**, generalizing the M1 `/walk`
channel-queue (`spectate.go:293-307`). M1 ships it as `walkCh chan [2]int`; M2 wants
`actCh chan func() error` (a closure that calls the right `Host` method). **Recommendation for
M1 (see §8.1):** ship the worker as a `func() error` queue *now*, with `/walk` enqueuing
`func() error { return host.Walk(...) }`. That single decision makes every M2 action a
zero-infrastructure addition. The retarget semantics (drop-pending, enqueue-latest,
`spectate.go:382-389`) stay for movement; window actions should *queue* rather than drop (a buy
then a sell must both run) — so the queue needs a small "coalesce vs serialize" flag per enqueue.

---

## 8. M1 DECISIONS THAT MUST BE RIGHT NOW (or M2 pays a refactor)

These are the only things that can't be deferred. Each is cheap in M1 and expensive later.

1. **Worker = `chan func() error`, not `chan [2]int`; routed via a registry keyed by bot id.**
   (Enables §1–§6 actions + §6 multi-bot with no infra change.) M1 has one bot, but the closure
   queue + registry shape must exist. *Blocks: every M2 action + multi-bot.*

2. **Allow a small, documented family of typed-payload POSTs beside `/act`** — at minimum
   `/offer {kind, items}` and `/chat {mode, recipient, message}`. Don't force *every* interaction
   through `/act`'s generic shape; basket offers and chat have genuinely different payloads.
   *Blocks: trade/duel offers (§1c/1d), PM (§4).*

3. **`/act` request carries an `args` object (free-form map), not just `{target, option}`.**
   Duel rules (4 booleans), Withdraw-X / Buy-N amounts, dialog-option index, and Cast-spell id all
   need a payload alongside the target+option. *Blocks: bank/shop quantities (§1a/1b), duel rules
   (§1d), dialog index (§5a).*

4. **`/state` is one doc of *nullable window sub-objects*, AND the architecture must permit a
   second fast poll (`/minimap`) separate from `/state`.** Don't hardwire "all JSON state comes
   from `/state`": windows poll slow, minimap polls fast. *Blocks: minimap cadence (§3); keeps
   window additions to one key each (§1).*

5. **The picker / `MenuTarget` contract must reserve a `pendingTarget` (deferred-source) mode**
   — even if M1 only uses it for inventory `Use`. Magic casting (§2) and use-item-on (§5b) are the
   *same* two-pick machine; if M1 hardcodes "pick → menu → act" with no pending-source state, M2
   magic+use-on require touching the core pick flow. Reserve `pendingTarget: null|{kind,id}` in
   `/state` and a "second pick resolves the pending source" branch in the dispatch entry point.
   *Blocks: magic casting (§2), use-item-on (§5b).*

**Everything else is purely additive** (new `Kind` consts, new dispatch rows, new `/state` keys,
new panels) and needs no M1 foresight beyond the five above.

---

## 9. Summary matrix

| M2 feature | Host methods (exist?) | World read-state (exists?) | New `/state` key | New endpoint? | M1 blocker |
|---|---|---|---|---|---|
| Bank window | ✓ `runtime/bank.go` | ✓ `BankRecord` | `bank` | no (`/act`) | none |
| Shop window | ✓ `runtime/shop.go` | ✓ `ShopRecord` (+prices) | `shop` | no | none |
| Trade window | ✓ `runtime/trade.go` | ✓ `TradeRecord` | `trade` | `/offer` | §8.2 |
| Duel window | ✓ `runtime/duel.go` | ✓ `DuelRecord` | `duel` | `/offer` | §8.2, §8.3 |
| Skills/stats tab | n/a (read) | ✓ `world.Self` | `skills` | no | none |
| Magic tab | ✓ `runtime/magic.go` | ✓ `views_magic` | `spells`,`pendingCast` | no | §8.5 |
| Prayer tab | ✓ `runtime/prayer.go` | ✓ `Self.ActivePrayers` | `prayers` | no | none |
| Minimap | reuses `/walk`,`/pick` | ✓ Npcs/Players/Items/Self | (own `/minimap`) | `/minimap` | §8.4 |
| Friends/ignore | ✓ write `runtime/social.go` | ⚠ **roster NOT decoded** | `social` | no (`/act`,`/chat`) | none UI; **data gap** |
| Dialog-option UI | ✓ `ChooseDialogOption` | ✓ `DialogOptionsRecord` | `dialog` | no | none |
| Use-item-on drag | ✓ `UseItemOn*` ×6 | ✓ inventory | `pendingTarget` | no | §8.5 |
| Multi-bot tabs | n/a (fan-out) | ✓ per-Host world | (per-bot `/state`) | `?bot=` param | §8.1 |

**Bottom line:** the protocol *and* the world read-layer are essentially M2-complete today.
Eleven of twelve features are pure UI-layer additions satisfying contracts 1–4. The lone data
gap (friends/ignore roster decode, §4) is outside the UI contracts and doesn't affect M1 design.
Ship the five §8 decisions in M1 and every M2 window/tab is "add rows + a key + a panel."
