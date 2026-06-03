# Remote-client screen recon (RECON wave)

Date: 2026-06-02. Driver bot: `webreact` (player id 4) on `http://localhost:8090`.
Second bot: `webreact2` (player id 16) on `http://localhost:8091`.
Both granted **Group.OWNER (group_id=0)** so `::` commands work.

This doc is the ground truth for the NEXT wave: which server-driven screens are
reachable, the REAL `/state` payload shape for each, and exactly what differs
from `docs/remote-client/specs/*.md` and the current `web/src/components/*`.

------------------------------------------------------------------------------
## How admin was obtained (access method)

The OpenRSC server is NOT in docker (the `openrsc/docker-compose.yml` mariadb is
a deployment artifact and is NOT running — `db_type: sqlite` in
`openrsc/server/connections.conf`). The live game DB is a **SQLite file**:

    /home/free/code/rsc-hacking/openrsc/server/inc/sqlite/westworld.db

(confirmed via the server JVM's open fd: pid → `inc/sqlite/westworld.db`).
Schema: `players.group_id` (int), groups from
`server/.../entity/player/Group.java`: OWNER=0, ADMIN=1, SUPER_MOD=2, MOD=3,
USER=10. `Player.isMod()` is true for OWNER/ADMIN/SUPER_MOD/MOD, and most
`::` commands gate on `isMod()` (e.g. `Moderator.canUse()` returns
`player.isMod()`); `::fillbank`/`::spawnnpc` are in `Admins.java` (owner/admin).

CRITICAL persistence gotcha: a running server holds the player in memory and
**periodically saves it back to the DB, overwriting any UPDATE made while the
account is logged in.** So a naive `UPDATE players SET group_id=0` got reverted.

Working procedure:
1. `kill -TERM` the cradle bot → graceful logout → server writes the player's
   final state (group_id=10) to the DB.
2. `sqlite3 westworld.db "UPDATE players SET group_id=0 WHERE username='Webreact';"`
   (use `-cmd ".timeout 8000"` — journal_mode is `delete`, so the write briefly
   contends with the server). Verify it sticks (no session owns the row now).
3. Relaunch cradle → login reads group_id=0 from DB → bot is OWNER in memory.
   The DB row then stays 0 because memory now agrees.

After relogin `::bank` opened the bank → admin confirmed. NOTE: tutorial island
gates many actions behind `!isMod()` (`Player.java:3283`); OWNER bypasses this,
so commands work even though the bot is still on tutorial island at login.

Cradle launch (jag env vars give real item icons):

    WESTWORLD_MEDIA_JAG=.../media58.jag WESTWORLD_CONFIG_JAG=.../config85.jag \
    WESTWORLD_ENTITY_JAG=.../entity24.jag /tmp/cradle -server localhost:43594 \
    -username webreact -password reactpass1 -facts .../openrsc -client \
    -client-addr localhost:8090   (nohup ... &)

------------------------------------------------------------------------------
## Two-bot setup (trade/duel)

- Registered: `go run ./cmd/register -server localhost:43594 -username webreact2 -password reactpass2` → player id 16.
- Granted OWNER **before first login** (clean — no session to overwrite it).
- Launched a 2nd cradle on `:8091` with the same jag env vars.
- Co-located: from bot1 (OWNER) `::summon Webreact2` teleports bot2 onto bot1's
  tile. Both ended at Varrock (122,509) after `::teleport varrock` on bot1.

------------------------------------------------------------------------------
## THE CORE FINDING — what `/state` exposes vs what each screen needs

`stateResponse` (cmd/cradle/remoteclient.go:297-312) has ONLY these windowed
blocks: `bank` (omitempty), `shop` (omitempty), `magic`, `prayers`.
**There is NO `trade` and NO `duel` field, and no structured `dialog` field.**
The `/state` handler (remoteclient.go:1037-1152) only builds `bankBlock`,
`shopBlock`, `magicBlock`, `prayerList`.

So the reachability of each server-driven window is gated by TWO independent
things: (a) can the bot trigger the server to open it, and (b) does `/state`
surface it. Summary table:

| screen      | backend record/Host | /state block | HTTP endpoint   | React component        | reachable now |
|-------------|---------------------|--------------|-----------------|------------------------|---------------|
| bank        | world.BankState     | YES (bank)   | POST /bank      | BankWindow.tsx (wired) | **YES**       |
| shop        | world.ShopState     | YES (shop)   | POST /shop      | ShopWindow.tsx (wired) | NO (open path)|
| trade       | world.TradeState + runtime/trade.go | NO | NONE      | none                   | NO            |
| duel        | world.DuelState + runtime/duel.go   | NO | NONE      | none                   | NO            |
| npc-dialog  | world.recent DialogText/DialogOptions | flattened→chat | NONE | none           | partial (text)|

------------------------------------------------------------------------------
## BANK — reachable, fully working

How reached: `::bank` (opens) then `::fillbank` (fills). NOTE: `::fillbank`
fills the bank record but the open window doesn't auto-refresh — close+reopen
(`POST /bank {op:close}` then `::bank` again) to see the 192 filled slots.

Real `/state.bank` (trimmed):

    {"open":true,"maxSize":192,"slotCount":192,"slots":[
      {"slot":0,"itemId":0,"name":"Iron Mace","amount":50},
      {"slot":1,"itemId":1,"name":"Iron Short Sword","amount":50},
      {"slot":2,"itemId":2,"name":"Iron Kite Shield","amount":50},
      {"slot":3,"itemId":3,"name":"Iron Square Shield","amount":50},
      {"slot":4,"itemId":4,"name":"Wooden Shield","amount":50},
      {"slot":5,"itemId":5,"name":"Medium Iron Helmet","amount":50}, ... ]}

Screenshot: docs/remote-client/screens/bank.png — renders correctly: titled
"Bank of Westworld — 192/192", Close button, real item sprites, side tabs.
Impl notes: matches the current BankWindow.tsx + spec. The only rough edge is
the fillbank-doesn't-refresh-open-window behaviour, which is a server quirk, not
a client bug.

------------------------------------------------------------------------------
## SHOP — backend + UI fully built, but NOT openable through the HTTP menu

Backend is complete: `/state` builds `shopBlock` from `host.World().Shop.Shop()`
(remoteclient.go:1059-1082), `POST /shop` (buy/sell/close) exists, ShopWindow.tsx
is built and wired in App.tsx (`state?.shop?.open && <ShopWindow .../>`).
types.ts has `Shop`/`ShopSlot`. The wire record is `world.ShopRecord`
(world/shop.go:25-59): IsGeneral, SellPriceMod, BuyPriceMod, PriceMultiplier,
Slots[]{ItemID,Stock,BaseStock}; the /state slot adds Name + computed
BuyPrice/SellPrice from `def.BasePrice`.

BLOCKER (the open path): an RSC shop opens ONLY when the player **Talk-to**s a
shopkeeper. The cradle right-click menu for an NPC (remoteclient menu.go
BuildMenu, KindNPC case, lines 66-83) emits ONLY the NPC's facts
`Command1`/`Command2` + a synthetic Attack (if attackable) + Examine. In
OpenRSC's `NpcDefs.json`, shopkeepers (e.g. id 51 "Shopkeeper") have
`command`/`command2` = "" — the standard "Talk-to" interaction is IMPLICIT in
RSC and is NOT stored in the def. So the menu collapses to **[Examine] only**,
and there is no `OptTalkTo` entry to fire. `host.TalkToNpc` and the dispatcher's
`OptTalkTo → TalkToNpc` mapping (dispatch.go:312-313) both EXIST but are
unreachable because BuildMenu never adds Talk-to.

Verified: spawned a Shopkeeper (`::spawnnpc 51`), `/pick` returned it with
options `['Examine']`. A Banker (id 95, command="Bank") `/pick`'d with
`['Bank','Collect','Examine']` and `POST /act optionId=0` fired "Bank" with
`{ok:true}` — proving the NPC-interaction→window pipeline works WHEN a verb
exists. So shop will work the moment a synthetic "Talk-to" verb is added.

FIX FOR NEXT WAVE: in `remoteclient/menu.go` BuildMenu KindNPC, add a synthetic
`OptTalkTo "Talk-to"` for NPCs that are NOT attackable and have no Command that
already means talk (this matches authentic RSC, which always offers Talk-to on
peaceful NPCs). That single change makes shop AND npc-dialog reachable.

stateSample (the shape the open shop WILL produce — derived from shop.go +
the /state builder, since it can't be opened to dump live):

    {"open":true,"isGeneral":true,"slots":[
      {"itemId":<def id>,"name":"<def name>","stock":<qty>,
       "buyPrice":<gp player pays>,"sellPrice":<gp shop pays>}, ... ]}

Screenshot: not captured (window cannot be opened). See viewport.png/magic-tab.png
for the side panel + scene.

------------------------------------------------------------------------------
## TRADE — backend half-wired, NO /state block, NO endpoints, NO component

How driven: from bot1 `/pick` bot2 (player, index 1) → options
`['Trade with','Follow','Duel with','Attack','Examine']`. `POST /act optionId=0`
("Trade with") returned `{ok:true,"message":"Trade with"}`; bot1 chat shows
"Sending trade request"; bot2 log shows `TRADE REQUEST received`. So the OUTBOUND
init works (runtime/trade.go `InitTradeRequest` via dispatchPlayer OptTrade).

BLOCKERS:
1. NO `/state.trade` block — `/state` never serializes `host.World().Trade.Trade()`,
   so the SPA can't see the trade at all (confirmed null on both bots).
2. NO HTTP endpoint to ACCEPT/offer/confirm. The Host methods all exist
   (`runtime/trade.go`: InitTradeRequest, AcceptIncomingTrade, DeclineTrade,
   OfferTradeItems, ConfirmTrade, FinalizeTrade) but nothing routes to
   AcceptIncomingTrade etc. So bot2 can't accept and the window never opens.
3. NO TradeWindow.tsx, no `Trade` type in types.ts.

Ground-truth record to mirror — `world.TradeRecord` (world/trade.go:19-46):
WithIndex, WithName, MyOffer[]/TheirOffer[] ({ItemID,Amount}), MyFirstAccepted,
TheirFirstAccepted, MySecondAccepted, TheirSecondAccepted,
Phase ("request_sent"|"open"|"confirm"|"completed"|"cancelled"), OpenedAt,
UpdatedAt. See docs/remote-client/specs/trade.md for the intended UI.

FIX FOR NEXT WAVE: add a `trade` block to stateResponse + /state builder (mirror
TradeRecord like the bank block); add POST /trade {op: request|accept|decline|
offer|confirm} routing to the existing Host methods through enqueueAction; build
TradeWindow.tsx (two-screen offer→confirm per the spec).

------------------------------------------------------------------------------
## DUEL — identical situation to trade

How driven: from bot1 `/pick` bot2 → `POST /act optionId=2` ("Duel with")
returned `{ok:true,"message":"Duel with"}`, chat "Sending duel request"
(runtime/duel.go `InitDuelRequest` via dispatchPlayer OptDuel).

BLOCKERS: same three as trade — no `/state.duel`, no HTTP endpoint to
accept/offer/set-rules/confirm (Host methods exist: InitDuelRequest,
AcceptIncomingDuel, DeclineDuel, OfferDuelItems, SetDuelRules, AcceptDuelOffer,
AcceptDuelConfirm), no DuelWindow.tsx, no `Duel` type.

Ground-truth record — `world.DuelRecord` (world/duel.go:21-53): same fields as
TradeRecord PLUS `Rules DuelRules{DisallowRetreat,DisallowMagic,DisallowPrayer,
DisallowWeapons}` (true = disallowed). The duel UI is trade + a 4-toggle rules
panel; see docs/remote-client/specs/duel.md.

------------------------------------------------------------------------------
## NPC-DIALOG — text reaches chat, options + window do NOT

How observed: at login the tutorial guide auto-spoke. The server sends it as a
`kind=dialog` message; cradle's chat ingest (remoteclient.go:206 — `if m.Kind
== "dialog"`) folds it into the chat ring as a `kind:"npc"` entry.

Real `/state.chat` dialog entry:

    {"seq":3,"kind":"npc","who":"",
     "text":"@gre@Welcome to the Westworld tutorial.% %Most actions are performed
     with the mouse. ... Try left clicking on one of the guides to talk to her."}

(The `@gre@`/`%` RSC color+newline codes are passed through raw.)

BLOCKERS:
1. The dialog OPTION MENU (the list of player choices) is tracked in
   `world.recent` (`DialogOptionsRecord{Options []string, At}`,
   world/recent.go:93-100) but is NOT surfaced in `/state` at all — there is no
   dialog block and no way to ANSWER an option (no endpoint).
2. Initiating a conversation is blocked by the SAME missing "Talk-to" verb as
   shop: peaceful NPCs (guides, quest NPCs like "Wilough" id 781) `/pick` with
   `['Examine']` only, so the bot can't even start a dialog through the menu.
3. No NpcDialog component and no `dialog` type.

stateSample (the dialog text as it ACTUALLY appears today — inside chat, not a
dedicated block):  see the chat entry above.

FIX FOR NEXT WAVE: (a) synthetic "Talk-to" verb (shared with shop fix);
(b) add a `dialog` block to /state mirroring DialogText + DialogOptions;
(c) POST /dialog {optionIndex} to answer; (d) NpcDialog component (speech bubble
+ option buttons). Spec: docs/remote-client/specs/dialog-useon.md.

------------------------------------------------------------------------------
## Minimap data source (nearby npcs / players / ground-items / scenery)

The minimap-relevant entities are enumerated by `buildLiveView` in
**cmd/cradle/spectate.go:183-251** (the SAME view `/frame` and `/pick` use). It
pulls four live world accessors:

- **NPCs**: `host.World().Npcs.All()` → `[]world.NpcRecord` —
  world/world.go:334 (record type world/world.go:125). Fields used for the
  minimap: X, Y, TypeID, Heading, Index (spectate.go:189-198).
- **Players**: `host.World().Players.All()` → `[]world.PlayerRecord` —
  world/world.go:607 (record type world/world.go:353). Index 0 is self and is
  skipped (spectate.go:199-215); fields: Index, X, Y, Heading (+ equip/colours).
- **Ground items**: `host.World().GroundItems.All()` → `[]world.GroundItemRecord`
  — world/world.go:81 (record type world/world.go:50). Fields: X, Y, ItemID
  (spectate.go:243-248).
- **Dynamic scenery**: `host.World().Scenery.All()` → `[]world.SceneryRecord`
  — world/scenery.go:88 (record type world/scenery.go:18). Fields: X, Y, ID
  (spectate.go:237-242). Removed scenery/boundaries via
  `world.Scenery.IsRemoved` / `world.Boundaries.IsRemoved` (spectate.go:235-236).
- **Boundaries (walls)**: `host.World().Boundaries.All()` →
  `map[BoundaryKey]int` — world/boundaries.go:66.

`/pick` (remoteclient.go:758-844) builds that same live view, runs
`render.Pick` (render/pick.go:199; billboards at pick.go:242, ground items at
pick.go:312, dynamic scenery at pick.go:388), then `remoteclient.BuildCandidates`
(remoteclient/menu.go:209) maps hits to wire candidates with absolute coords.

For a minimap feature: read those four `.All()` accessors directly (they are
already plane-aware and lock-safe), translate each entity's (X, Y) relative to
`host.World().Self.Position()`, and color by kind. No new world plumbing needed;
the data is exactly what `buildLiveView` already collects.

------------------------------------------------------------------------------
## Blockers (precise)

- **Shop / npc-dialog initiation**: `remoteclient/menu.go` BuildMenu (KindNPC,
  lines 66-83) never synthesizes a "Talk-to" verb, and OpenRSC NpcDefs store no
  Talk-to command. `host.TalkToNpc` + dispatcher OptTalkTo exist but are
  unreachable. → add synthetic Talk-to for peaceful NPCs. (Recon-only this phase;
  no source edited.)
- **Trade / duel windows**: `/state` has no trade/duel block; no
  /trade or /duel HTTP endpoints; no TradeWindow/DuelWindow components or types.
  Only the OUTBOUND request fires (InitTradeRequest / InitDuelRequest). The
  receiving side (AcceptIncoming*) is implemented in runtime/ but unrouted, so
  neither window can ever open. Both bots confirmed `trade:null` / `duel:null`.
- Could NOT capture live shop/trade/duel `/state` JSON because the windows can't
  be opened through the HTTP surface; the ground-truth shapes above are taken
  from the world/*.go records the implementers must mirror.

------------------------------------------------------------------------------
## Screenshots saved (docs/remote-client/screens/)

- bank.png — bank window open over the scene, 192 real item icons, side tabs.
- magic-tab.png — Varrock scene (bot1 + bot2 + spawned Banker) with the Magic
  spellbook tab active; chat log shows the trade/duel/summon commands.
- viewport.png — plain game viewport (no window open) for reference.
