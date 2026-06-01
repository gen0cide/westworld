# Layer 2 — Menu Model & Action Dispatch

**Status:** Design (M1)
**Depends on:** `00-overview.md` (north star), `10-hit-testing.md` (Layer 1 `render.Pick`)
**Consumed by:** `30-http-api.md` (Layer 3 endpoints), `40-browser-ui.md`

This document specifies **Layer 2**: the bridge between a screen-space pick and a
`runtime.Host` action. It defines

1. the **target taxonomy** (what you can right-click),
2. per-target **ordered menu option lists** with authentic RSC verbs,
3. a **serializable `MenuTarget`** that round-trips browser ↔ server losslessly,
4. the **dispatch table** mapping `{targetKind, optionID}` → an exact `Host`
   method + argument expression, and
5. the **package layout** and how it consumes `render.Pick` candidates.

Every `runtime.Host` signature cited below was verified against the live source
on the `feat/remote-client` branch (file:line in each row). Layer 2 is pure
routing — it sends **no packets of its own**; every action goes through `Host`
(which owns the wire) or `runtime/examine.go` (which reads state, no packet).

> **The single most important convention to internalise:** RSC's wire protocol
> (and therefore every `Host` method) is **0-based** for inventory slots, NPC
> dialog options, and spell ids, but **1-based** for scenery interact options
> (`InteractAt` `option` is 1 = primary / 2 = secondary). Boundary `direction`
> is an enum (0..3), not an ordinal. See §6 for the full table — getting this
> wrong sends a valid-looking packet to the wrong handler.

---

## 1. Target taxonomy

A **target** is a concrete thing in (or attached to) the world the user can act
on. Nine kinds, in two families:

### World targets (produced by `render.Pick` from a screen click)

| Kind | Identity | Source of candidate |
|---|---|---|
| `npc` | server `Index` (+ `NpcID` type, X, Y) | billboard AABB hit-test (`render.Entity{Kind:EntityNPC}`) |
| `player` | server `Index` (+ `Name`, X, Y) | billboard AABB hit-test (`render.Entity{Kind:EntityPlayer}`) |
| `self` | the host itself | the always-present centre billboard (`EntitySelf`) |
| `ground_item` | `ItemID` + tile X, Y | composited-icon AABB hit-test (`render.GroundItemMarker`) |
| `scenery` | `DefID` + tile X, Y (+ Direction) | `PickTile` → `facts.At(x,y)` scenery + live `world.DynamicScenery` |
| `boundary` | `DefID` + tile X, Y + **Direction (0..3)** | `PickTile` → `facts.At(x,y)` boundary + live `world.Boundaries` |
| `terrain` | tile X, Y, Plane | `PickTile` — always present, lowest priority |

### UI targets (produced by the browser panels, NOT `render.Pick`)

| Kind | Identity | Source |
|---|---|---|
| `inventory_item` | inventory **slot** (0-based) + `ItemID` | inventory grid panel (`/state` → `world.Inventory.Slots()`) |
| `equipment_item` | the **inventory slot** of the wielded item (0-based) + `ItemID` | equipment summary panel |

> **Critical equipment fact.** RSC has no separate "equipment store" you can
> address. A worn item is just an inventory slot with `InvSlot.Wielded == true`
> (`world/inventory.go:6`). The wire **carries no item-id-by-equip-slot** — the
> per-slot data in the appearance update is *sprite ids*, not catalogue ids
> (`world/world.go:376-386`, `PlayerRecord.EquipBySlot`). Therefore
> `EquipItem(slot)` / `UnequipItem(slot)` (`runtime/magic.go:42,47`) both take an
> **inventory slot index**, and an `equipment_item` target is keyed by the
> inventory slot of the wielded item — never by the 12-way `event.EquipSlot*`
> index. The equipment panel exists only to *display* worn items grouped by body
> slot; acting on one resolves back to its inventory slot.

`self` is split out from `player` deliberately: you can't Trade-with / Attack
yourself, and Examine-self has its own accessor (`ExamineSelf`).

---

## 2. Ordered menu option lists (authentic verbs)

RSC right-click menus are **ordered**, and the **top entry is the left-click
default action** (clicking without right-clicking fires entry 0). Verbs come
from the `facts` defs; `Examine` is always last; `Cancel` is implicit (the
browser closes the menu, no server call). Option ordering below mirrors the
authentic mudclient `menu*` arrays.

Notation: `Command1` / `Command2` mean "include only if the def field is
non-empty"; **bold** marks the left-click default (always entry 0).

### `npc`
Authentic NPC menu order (mudclient: command1 → command2 → talk → attack →
examine, de-duplicating talk/attack already named by a command):

1. **`Command1`** if `NpcDef.Command1 != ""` (e.g. `Talk-to`, `Pickpocket`)
2. `Command2` if `NpcDef.Command2 != ""`
3. `Attack` if `NpcDef.Attackable` **and** not already named by Command1/2
4. `Talk-to` if no Command1/2 supplied it and the NPC is talkable (fallback)
5. `Examine`

For the overwhelmingly common cases this reduces to: a shopkeeper/quest NPC →
`[Talk-to, Examine]` (Talk-to is Command1, default); a monster →
`[Attack, Examine]` (Attack is Command1, default); a thievable NPC →
`[Pickpocket, Attack, Examine]`. `NpcDef.Command1/Command2` at
`facts/defs.go:65-66`.

### `player`
Fixed standard menu (no per-def verbs for players):

1. **`Trade with`**
2. `Follow`
3. `Duel with`
4. `Attack` (server rejects outside PvP zones — still offered)
5. `Examine`

> `Follow` has a caveat — see §5 dispatch note (it takes a *name* and blocks).

### `self`
1. **`Walk here`** (centre tile) — or omit and let `terrain` handle it
2. `Examine` (→ `ExamineSelf`)

In practice `self` rarely surfaces its own menu; clicking the host's own
billboard usually falls through to the `terrain` target under it.

### `ground_item`
1. **`Pick up`**
2. `Examine`

(`Use` of an inventory item *onto* a ground item is a drag/cast interaction, not
a context-menu entry — deferred to M2 use-on UX; the dispatch verb exists,
`UseItemOnGroundItem`.)

### `scenery`
1. **`Command1`** if `SceneryDef.Command1 != ""` (e.g. `Chop`, `Mine`, `Open`, `Search`)
2. `Command2` if `SceneryDef.Command2 != ""`
3. `Examine`

`SceneryDef.Command1/Command2` at `facts/defs.go:12-13`. Note `Command2` is
*often literally* `"Examine"` in the def data — Layer 2 must **de-duplicate**:
if `Command2 == "Examine"` (case-insensitive) drop it and rely on the synthetic
Examine entry, so the menu never shows `Examine` twice.

### `boundary` (wall / door / gate / fence)
1. **`Command1`** if `BoundaryDef.Command1 != ""` (e.g. `Open` for a door)
2. `Command2` if `BoundaryDef.Command2 != ""` (often `Close`/`Examine`)
3. `Examine`

`BoundaryDef.Command1/Command2` at `facts/defs.go:32-33`. Same `Examine`
de-dup rule as scenery. A plain wall has `Command1 == "WalkTo"` and no real
interaction — its menu collapses to `[Examine]` (and the tile under it still
offers `Walk here` via the `terrain` target).

### `terrain`
1. **`Walk here`**
2. *(Cancel is implicit)*

The lowest-priority, always-present target. `Walk here` reuses the existing
`/walk` path (`render.PickTile` → `host.Walk`).

### `inventory_item`
1. **`Command`** if `ItemDef.Command != ""` (e.g. `Eat`, `Drink`, `Bury`) — this is the left-click default
2. `Wield` / `Wear` if `ItemDef.IsWearable` **and** the slot is not already `Wielded`
3. `Remove` if the slot **is** `Wielded` (i.e. it's worn; lets you unequip from the inventory view too)
4. `Use` (use-on; M2 drag UX — verb present, target picked later)
5. `Drop`
6. `Examine`

`ItemDef.Command` / `IsWearable` at `facts/defs.go:83,87`; `InvSlot.Wielded` at
`world/inventory.go:9`. If `ItemDef.Command` is itself `"Wield"`/`"Wear"`
(common for armour/weapons), fold it into entry 2 rather than emitting both —
the def's Command **is** the wield verb for wearables. So an un-worn sword
becomes `[Wield, Use, Drop, Examine]` with `Wield` as the default; food becomes
`[Eat, Use, Drop, Examine]`.

### `equipment_item` (worn slot in the equipment panel)
1. **`Remove`** (→ `UnequipItem` on the underlying inventory slot)
2. `Examine`

---

## 3. Serializable `MenuTarget`

One flat struct, JSON-tagged, that round-trips browser ↔ server **losslessly**
and carries enough identity to dispatch without a second pick. Every field is a
value type (ints/strings) — no pointers, no server handles — so the browser can
hold it opaquely between `/pick` and `/act`.

```go
// MenuTarget identifies one actionable thing, in a form that survives a
// browser ↔ server JSON round-trip. The server emits it from a pick (world
// targets) or a panel (inventory/equipment targets); the browser echoes it back
// verbatim in /act. No field references live server state — Index/Slot are
// re-validated at dispatch time against the current world mirror.
type MenuTarget struct {
    Kind TargetKind `json:"kind"` // see TargetKind constants

    // World identity. Which subset is populated depends on Kind:
    //   npc/player          -> Index (server actor index)
    //   ground_item         -> ItemID + X,Y
    //   scenery             -> DefID + X,Y (+ Dir for completeness)
    //   boundary            -> DefID + X,Y + Dir (0..3, see below)
    //   terrain/self        -> X,Y,Plane
    //   inventory_item      -> Slot (0-based) + ItemID
    //   equipment_item      -> Slot (0-based inventory slot of the worn item) + ItemID
    Index  int `json:"index,omitempty"`  // server NPC/player index
    DefID  int `json:"defId,omitempty"`  // scenery/boundary def id (facts)
    ItemID int `json:"itemId,omitempty"` // ground/inventory/equipment item catalogue id
    Slot   int `json:"slot,omitempty"`   // 0-based inventory slot (inv/equip)

    X     int `json:"x,omitempty"`
    Y     int `json:"y,omitempty"` // ABSOLUTE world Y (plane offset already applied)
    Plane int `json:"plane,omitempty"`
    Dir   int `json:"dir,omitempty"` // boundary direction 0..3 (see below)

    // Name is carried for display + the one Host method keyed by name (Follow).
    Name string `json:"name,omitempty"`
}

type TargetKind string

const (
    KindNPC           TargetKind = "npc"
    KindPlayer        TargetKind = "player"
    KindSelf          TargetKind = "self"
    KindGroundItem    TargetKind = "ground_item"
    KindScenery       TargetKind = "scenery"
    KindBoundary      TargetKind = "boundary"
    KindTerrain       TargetKind = "terrain"
    KindInventoryItem TargetKind = "inventory_item"
    KindEquipmentItem TargetKind = "equipment_item"
)
```

A menu sent to the browser is the target plus its options:

```go
// MenuOption is one row in the context menu.
type MenuOption struct {
    ID    OptionID `json:"id"`    // stable enum, dispatch keys on it (NOT the label)
    Label string   `json:"label"` // authentic verb shown to the user
}

// Menu is what /pick returns per candidate (depth-ordered list of these).
type Menu struct {
    Target  MenuTarget   `json:"target"`
    Options []MenuOption `json:"options"` // Options[0] is the left-click default
    Examine string       `json:"examine"` // pre-resolved examine text (no extra round-trip)
}
```

`OptionID` is a stable string enum (`"attack"`, `"talk_to"`, `"command1"`,
`"command2"`, `"pickup"`, `"wield"`, `"remove"`, `"use"`, `"drop"`,
`"walk_here"`, `"trade"`, `"follow"`, `"duel"`, `"examine"`, …). **Dispatch keys
on `(Kind, OptionID)`, never on the human label** — labels are localised verbs
from def data and vary per target; the OptionID is the contract.

> **Why pre-resolve `Examine` into the menu.** `runtime/examine.go` sends no
> packet (it reads facts + world mirrors, `examine.go:11-19`), so resolving it
> at pick time is free and saves a round-trip. `Examine` dispatch is then a pure
> client-side menu draw — see §4.

### Boundary `Dir` convention (0..3) — verified

The boundary direction is the authentic `createModel` edge encoding
(`render/render.go:83-90`, matching `facts.BoundaryLoc.Direction`
`facts/defs.go:107`):

```
0 = edge (x,y)..(x+1,y)     east-west edge   (.orsc VerticalWall byte)
1 = edge (x,y)..(x,y+1)     north-south edge (.orsc HorizontalWall byte)
2 = edge (x,y)..(x+1,y+1)   '\' diagonal
3 = edge (x+1,y)..(x,y+1)   '/' diagonal
```

`MenuTarget.Dir` is exactly the value `facts.At(x,y)` reports as
`Placement.Direction` for a `boundary` placement, fed straight into
`host.InteractWithBoundary(ctx, x, y, dir)` (`runtime/boundary.go:19`). No
remapping.

---

## 4. The dispatch table

`(Kind, OptionID)` → the exact `Host` call. `t` is the `MenuTarget`; `ctx` is a
per-action `context.WithTimeout(parent, 30s)` (mirroring the spectate `/walk`
worker, `cmd/cradle/spectate.go:300`). All signatures verified at the cited
file:line.

### NPC (`KindNPC`) — `t.Index` is the server NPC index

| OptionID | Host call | Signature ref |
|---|---|---|
| `command1` | dispatch by verb: `Attack`→`host.AttackNpc(ctx, t.Index)`; `Talk-to`→`host.TalkToNpc(ctx, t.Index)`; else (`Pickpocket`/other)→`host.NpcCommand(ctx, t.Index)` | combat.go:21 / combat.go:514 / boundary.go:116 |
| `command2` | same verb dispatch as `command1` | — |
| `attack` | `host.AttackNpc(ctx, t.Index)` | combat.go:21 |
| `talk_to` | `host.TalkToNpc(ctx, t.Index)` | combat.go:514 |
| `examine` | `host.ExamineNpc(t.Index)` (no packet; text already in menu) | examine.go:50 |

> `command1`/`command2` are *labels from `NpcDef`*; Layer 2 resolves the label
> to the right method. The three primitive NPC verbs the wire distinguishes are
> Attack (`AttackNpc`), Talk-to (`TalkToNpc`), and "primary command"
> (`NpcCommand`, opcode NPC_COMMAND → server command1, e.g. pickpocket,
> `boundary.go:111-125`). A label that isn't "Attack"/"Talk-to" routes to
> `NpcCommand`. All three pathfind adjacent first.

### Player (`KindPlayer`) — `t.Index` is the server player index, `t.Name` the name

| OptionID | Host call | Signature ref |
|---|---|---|
| `trade` | `host.InitTradeRequest(ctx, t.Index)` | trade.go:14 |
| `follow` | `host.Follow(ctx, t.Name, startupTimeout)` — **special, see note** | follow.go:25 |
| `duel` | `host.InitDuelRequest(ctx, t.Index)` | duel.go:14 |
| `attack` | `host.AttackPlayer(ctx, t.Index)` | combat.go:204 |
| `examine` | `host.ExaminePlayer(t.Index)` (no packet) | examine.go:198 |

> **`Follow` is the one method keyed by name, and it blocks.**
> `Host.Follow(ctx, targetName string, startupTimeout time.Duration)` waits for
> the target's appearance to learn the index, then **loops until ctx is
> cancelled** re-sending the follow packet (follow.go:25-92). It must NOT run on
> the shared single-action worker (it would wedge the queue). Layer 2 runs
> `Follow` on its **own cancellable goroutine**, stored as the "active follow";
> a new movement/action or a `/act follow stop` cancels it. This is why
> `MenuTarget.Name` exists. (Index→name comes from the pick's
> `world.PlayerRecord.Name`, `world/world.go:355`.)

### Self (`KindSelf`)

| OptionID | Host call | Signature ref |
|---|---|---|
| `walk_here` | `host.Walk(ctx, t.X, t.Y)` (or queue on the walk worker) | host.go:666 |
| `examine` | `host.ExamineSelf()` (no packet) | examine.go:216 |

### Ground item (`KindGroundItem`) — `t.ItemID` + `t.X,t.Y`

| OptionID | Host call | Signature ref |
|---|---|---|
| `pickup` | `host.PickUpItem(ctx, t.X, t.Y, t.ItemID)` | items.go:27 |
| `use` (M2) | `host.UseItemOnGroundItem(ctx, t.X, t.Y, t.ItemID, srcSlot)` | boundary.go:84 |
| `examine` | `host.ExamineGroundItem(t.X, t.Y)` (no packet) | examine.go:138 |

### Scenery (`KindScenery`) — `t.X,t.Y`; **`option` is 1-based**

| OptionID | Host call | Signature ref |
|---|---|---|
| `command1` | `host.InteractAt(ctx, t.X, t.Y, 1)` | boundary.go:71 |
| `command2` | `host.InteractAt(ctx, t.X, t.Y, 2)` | boundary.go:71 |
| `use` (M2) | `host.UseItemOnScenery(ctx, t.X, t.Y, srcSlot)` | boundary.go:56 |
| `examine` | `host.ExamineScenery(t.X, t.Y)` (no packet) | examine.go:154 |

> `InteractAt`'s `option` argument is **1 = primary, 2 = secondary**
> (`boundary.go:66-79`; opcode 136 / 79). This is the single 1-based interact
> ordinal in the whole table. `command1` → `1`, `command2` → `2`.

### Boundary (`KindBoundary`) — `t.X,t.Y` + `t.Dir` (0..3)

| OptionID | Host call | Signature ref |
|---|---|---|
| `command1` | `host.InteractWithBoundary(ctx, t.X, t.Y, t.Dir)` | boundary.go:19 |
| `command2` | `host.InteractWithBoundary(ctx, t.X, t.Y, t.Dir)` — same packet (the boundary has one default interaction; `command2` if present is usually `Close`, the same wire op) | boundary.go:19 |
| `use` (M2) | `host.UseItemOnBoundary(ctx, t.X, t.Y, t.Dir, srcSlot)` | boundary.go:34 |
| `examine` | `host.ExamineBoundary(t.X, t.Y)` (no packet) | examine.go:174 |

> Boundary interaction carries **no option ordinal** on the wire — there is one
> `InteractWithBoundary(x,y,direction)` packet (open/close a door is the same
> toggle op, the server flips state). So both `command1` and `command2` map to
> the same call; the labels differ for UX only. `t.Dir` is passed through
> unchanged (§3).

### Terrain (`KindTerrain`) — `t.X,t.Y`

| OptionID | Host call | Signature ref |
|---|---|---|
| `walk_here` | `host.Walk(ctx, t.X, t.Y)` (reuse the `/walk` worker queue) | host.go:666 |

### Inventory item (`KindInventoryItem`) — `t.Slot` 0-based + `t.ItemID`

| OptionID | Host call | Signature ref |
|---|---|---|
| `command` | `host.ItemCommand(ctx, t.Slot)` (Eat/Drink/Bury — server-side default action) | items.go:68 |
| `wield` | `host.EquipItem(ctx, t.Slot)` | magic.go:42 |
| `remove` | `host.UnequipItem(ctx, t.Slot)` (when the slot is already `Wielded`) | magic.go:47 |
| `use` (M2) | `host.UseItemOnItem(ctx, t.Slot, dstSlot)` (or `UseItemOn{Npc,Player,Scenery,Boundary,GroundItem}` if dragged onto a world target) | boundary.go:48 / 56 / 34 / 84 / 99 / 130 |
| `drop` | `host.DropItem(ctx, t.Slot)` | items.go:15 |
| `examine` | `host.ExamineInventorySlot(t.Slot)` (no packet) | examine.go:118 |
| `cast` (M2) | `host.CastOnInventory(ctx, t.Slot, spellID)` (enchant/convert) | magic.go:37 |

> **Slot is 0-based** end to end: `world.Inventory.Slots()` returns a slice
> indexed from 0 (`world/inventory.go:60`), and every item Host method
> (`ItemCommand`, `DropItem`, `EquipItem`, `UnequipItem`) takes that same 0-based
> slot. No +1/−1 conversion anywhere. Validate `t.Slot` against the **current**
> `Slots()` at dispatch (the inventory may have shifted since the pick — RSC
> `RemoveSlot` compacts holes, `world/inventory.go:50-57`); if `Slots()[t.Slot]`
> no longer holds `t.ItemID`, re-resolve the slot by `ItemID` or reject the
> stale action.

### Equipment item (`KindEquipmentItem`) — `t.Slot` = inventory slot of the worn item

| OptionID | Host call | Signature ref |
|---|---|---|
| `remove` | `host.UnequipItem(ctx, t.Slot)` | magic.go:47 |
| `examine` | `host.ExamineInventorySlot(t.Slot)` (no packet) | examine.go:118 |

> An `equipment_item` is identified by the **inventory slot** of the wielded
> item (the slot where `InvSlot.Wielded == true`), per §1's equipment fact. The
> equipment panel groups worn slots by body part for display, but `Remove`
> resolves to that inventory slot and calls `UnequipItem(slot)` — the exact same
> call `inventory_item`'s `remove` uses. There is no equip-slot-indexed wire op.

---

## 5. Build-time vs dispatch-time responsibilities

**Menu builder** (`render.Pick` candidate or panel slot → `Menu`): pure, no
packets. For each candidate:
- look up the def (`facts.SceneryDef/BoundaryDef/NpcDef/ItemDef`, all return
  `nil`-safe pointers, `facts/facts.go:53-63`),
- emit the ordered `[]MenuOption` per §2 (de-dup `Examine`, fold Wield-Command),
- resolve `Examine` text via the matching `runtime/examine.go` accessor and
  stash it in `Menu.Examine`.

**Dispatcher** (`(MenuTarget, OptionID)` → effect): the §4 table. Two execution
lanes:
- **Single-action lane** — every action *except* Follow goes through ONE
  serialized worker goroutine (the `walkCh` pattern, `spectate.go:293-307`),
  because RSC actions override each other (a new attack cancels a pending walk).
  Re-target semantics: drain any queued action, enqueue the latest
  (non-blocking), exactly as `/walk` does (`spectate.go:382-389`).
- **Follow lane** — `Follow` runs on its own cancellable goroutine (it blocks
  until ctx cancel, §4 player note). Starting any single-action lane action, or
  an explicit follow-stop, cancels the active follow.
- **Examine lane** — never enqueued; resolved synchronously at pick time (no
  packet), so the browser already has the text. An `examine` `/act` (if the UI
  re-requests it) just returns the cached `Examination`.

Dispatch always re-validates volatile identity against the live world mirror
before firing (NPC/player still visible at `Index`? inventory `Slot` still holds
`ItemID`?) and returns a typed "stale target" error rather than firing a
mis-aimed packet.

---

## 6. Convention reference (0-based vs 1-based) — the cheat sheet

| Quantity | Base | Where | Source |
|---|---|---|---|
| Inventory slot (`ItemCommand/Drop/Equip/Unequip/UseItemOnItem/CastOnInventory`) | **0-based** | dispatch `t.Slot` | inventory.go:60, items.go:15/68, magic.go:37/42/47 |
| Scenery interact `option` (`InteractAt`) | **1-based** (1=primary, 2=secondary) | dispatch | boundary.go:66-79 |
| Boundary `direction` (`InteractWithBoundary`) | enum **0..3** (not an ordinal) | dispatch `t.Dir` | boundary.go:19, render.go:83-90 |
| NPC dialog option (`ChooseDialogOption`) | **0-based** | M2 dialog UI | combat.go:218-224 |
| Spell id (`CastOn*`) | **0-based** spellbook index | M2 magic UI | magic.go:9-13 |
| Server actor `Index` (npc/player) | opaque server handle | dispatch `t.Index` | world.go:126, 354 |
| `MenuOption` list index | 0-based; **entry 0 = left-click default** | menu UX | §2 |

---

## 7. Package layout & `render.Pick` consumption

New package: **`github.com/gen0cide/westworld/remoteui`** (Layer 2). Pure
domain logic, no `net/http` (that's Layer 3 in `cmd/cradle`), no rendering
(that's `render`). Depends on `runtime`, `facts`, `world`, `render`, `action`
(for `action.CombatStyle` if combat-style options land in M2).

```
remoteui/
  target.go     // MenuTarget, TargetKind, OptionID, MenuOption, Menu (§3)
  menu.go       // BuildMenu(...) — candidate/slot -> Menu (§2, §5 builder)
  dispatch.go   // Dispatcher{host} + Dispatch(ctx, MenuTarget, OptionID) (§4)
  worker.go     // single-action lane (walkCh pattern) + follow lane (§5)
  menu_test.go  // table-driven: each def shape -> expected ordered options
  dispatch_test.go // (Kind,OptionID) -> assert correct host method invoked (host iface mock)
```

### Consuming `render.Pick`

`render.Pick` (Layer 1, designed in `10-hit-testing.md`) returns a
**depth-ordered** `[]PickCandidate`, nearest-camera-first, each carrying enough
identity to build a `MenuTarget`. Expected (to be finalised by Layer 1) shape:

```go
// render.PickCandidate (Layer 1) — one hit under the cursor.
type PickCandidate struct {
    Kind  PickKind // npc | player | self | ground_item | scenery | boundary | terrain
    Index int      // server actor index (npc/player)
    NpcID int      // sprite/type id (npc)
    DefID int      // scenery/boundary def id
    ItemID int     // ground item id
    X, Y, Plane int
    Dir   int      // boundary edge 0..3
    Name  string   // player name (from world.PlayerRecord)
}
```

Layer 2's entry point maps each candidate → a `Menu`:

```go
// BuildMenus turns the depth-ordered pick candidates into per-candidate menus,
// preserving order (candidates[0] is topmost / nearest). The browser shows a
// stacked context menu; the FIRST option of the FIRST candidate is the
// left-click default for the whole click.
func (b *MenuBuilder) BuildMenus(cands []render.PickCandidate) []Menu
```

The `PickKind` → `TargetKind` mapping is 1:1 by name. `terrain` is always the
last (lowest-priority) candidate Layer 1 appends, so `Walk here` is the
universal fallback option. Inventory/equipment menus do **not** come from
`render.Pick` — Layer 3 builds those `MenuTarget`s directly from
`world.Inventory.Slots()` (`world/inventory.go:60`) + the equipment panel and
passes them to `BuildMenu` / `Dispatch` the same way.

### `Host` surface this package depends on (interface for testability)

`Dispatcher` should accept a narrow interface (not the concrete `*runtime.Host`)
so `dispatch_test.go` can assert routing with a mock:

```go
type ActionHost interface {
    // movement
    Walk(ctx context.Context, x, y int) error
    // npc
    AttackNpc(ctx context.Context, serverIndex int) error
    TalkToNpc(ctx context.Context, serverIndex int) error
    NpcCommand(ctx context.Context, npcServerIndex int) error
    // player
    InitTradeRequest(ctx context.Context, serverIndex int) error
    InitDuelRequest(ctx context.Context, serverIndex int) error
    AttackPlayer(ctx context.Context, serverIndex int) error
    Follow(ctx context.Context, targetName string, startupTimeout time.Duration) error
    // ground items
    PickUpItem(ctx context.Context, x, y, itemID int) error
    // scenery / boundary
    InteractAt(ctx context.Context, x, y, option int) error
    InteractWithBoundary(ctx context.Context, x, y, direction int) error
    // inventory / equipment
    ItemCommand(ctx context.Context, slot int) error
    DropItem(ctx context.Context, slot int) error
    EquipItem(ctx context.Context, slotIndex int) error
    UnequipItem(ctx context.Context, slotIndex int) error
    // M2 use-on (verbs reserved; not wired into M1 menus)
    UseItemOnItem(ctx context.Context, slot1, slot2 int) error
    UseItemOnScenery(ctx context.Context, x, y, slot int) error
    UseItemOnBoundary(ctx context.Context, x, y, direction, slot int) error
    UseItemOnGroundItem(ctx context.Context, x, y, groundItemID, slot int) error
    UseItemOnNpc(ctx context.Context, npcServerIndex, slot int) error
    UseItemOnPlayer(ctx context.Context, playerServerIndex, slot int) error
}
// *runtime.Host satisfies ActionHost (every method verified, §4).

// Examine reads go through a second tiny interface (no packets):
type ExamineHost interface {
    ExamineNpc(serverIndex int) runtime.Examination
    ExaminePlayer(serverIndex int) runtime.Examination
    ExamineSelf() runtime.Examination
    ExamineGroundItem(x, y int) runtime.Examination
    ExamineScenery(x, y int) runtime.Examination
    ExamineBoundary(x, y int) runtime.Examination
    ExamineInventorySlot(slot int) runtime.Examination
    ExamineItem(itemID int) runtime.Examination
}
```

This keeps Layer 2 decoupled from `runtime`'s internals while pinning the exact
verified signatures it routes to.

---

## 8. Open questions deferred to M2 (do not block M1)

- **Use-item-on-target drag UX** — verbs (`UseItemOn*`) are in the dispatch
  table but no menu emits them in M1; M1 inventory `Use` is a no-op placeholder.
- **Combat-style submenu** — `SetCombatStyle(ctx, action.CombatStyle)`
  (combat.go:230) is a settings toggle, not a per-target action; lives in the
  settings tab (M2).
- **Spell/prayer casting menus** — `CastOn*` / `ActivatePrayer` exist; the magic
  & prayer panels are M2 (`90-backlog.md`).
- **Dialog-option menu** — `ChooseDialogOption(ctx, index)` (0-based,
  combat.go:218) drives the NPC-chat option UI, M2.
- **Trade/duel/shop/bank windows** — the `Init*` methods open them; the windows
  themselves are M2.
