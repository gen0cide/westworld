# Layer 5 — Reconciled Per-File Implementation Spec (the cross-layer contract)

**Status:** Authoritative (M1 build)
**Reconciles:** `00-overview.md` (north star), `10-hit-testing.md` (Layer 1),
`20-menu-dispatch.md` (Layer 2), `30-http-api.md` (Layer 3), `40-browser-ui.md`
(Layer 4), pressure-tested against `90-backlog.md` (M2).

> **This document wins.** Where 10–40 disagree on a name, a type, a field, a
> file, or an endpoint, the resolution **here** is final and the implementer
> follows this doc. Section 1 lists every contradiction and its ruling so no one
> has to diff the layer docs. Sections 2–6 are the frozen contracts. Section 7 is
> the per-file plan (path, layer, model, exact signatures, one-paragraph spec).
> Section 8 is the build order + dependency graph. Every signature/struct/const
> cited was re-verified against the live source on `feat/remote-client` (file:line
> in-line); trust those, not the layer docs, where they differ.

---

## 0. Verified ground truth (re-checked for this spec)

These are the facts the contracts below are built on; all re-verified now.

| Fact | Source (verified) |
|---|---|
| `SetCamera(x, z, y, pitch, yaw, roll, distance int32)` — note arg names `x,z,y` | `render/camera.go:36` |
| `PickTile` and `DrawEntitySprites` BOTH call `SetCamera(localX, -elev, localZ, 912, rot*4, 0, zoom*2)` (i.e. `x=localX, z=-elev, y=localZ`) | `render/pick.go:32`, `render/render.go` (the §2.1 reconstruction) |
| `DrawEntitySprites` swaps `cam.CameraYaw,cam.CameraRoll` once at entry; `projectPoint` assumes already-swapped | `render/render.go:404-406`, `render/pick.go:35,82` |
| Billboard rect: `screenW=(worldW<<9)/camZ`, `screenH=(worldH<<9)/camZ`, blit at `(sx-screenW/2, feetY-screenH)` size `(screenW,screenH)`; reject `screenW<=0||screenH<=0`; painter sort `camZ` DESC; blit depth `camZ-spriteDepthBias` | `render/render.go:539-560` |
| `viewDist=9`, `terrainSize=160`, `clipNear=5`, `spriteDepthBias=64`, `playerBillboardW=145`, `playerBillboardH=220`, `groundItemPixelToWorld=2`, `npcBillboardSize(npcID)` | `render/scene.go:14-15`, `render/terrain.go:7`, `render/render.go:562,617`, `render/entitysprite.go:825,846-847` |
| `buildScene` defaults: `Zoom=750, W=512, H=334` (NOT PickTile's 600/512/336) | `render/render.go` (per 10-hit-testing §2.1) |
| `PickTile(land,v,px,py) -> (worldX,worldY int, ok bool)` returns **plane-LOCAL** coords | `render/pick.go:15-72` |
| `View{X,Y,Plane,Rotation,Zoom,W,H,AnimFrame,Entities,NoSelf,SelfHeading,SelfOff*,SelfStepPhase,SelfMoving,Self*appearance,BoundaryRemoved func,DynamicScenery,SceneryRemoved func,GroundItems}` | `render/render.go:27-119` |
| `Entity{X,Y,Kind,NpcID,Heading,EquipSprites,...,Index,OffX,OffZ,StepPhase,Moving}`; `EntityKind`={`EntityNPC`,`EntityPlayer`,`EntitySelf`} | `render/entity.go:6-65` |
| `GroundItemMarker{X,Y,ItemID}`, `DynamicSceneryItem{X,Y,ID,Direction}` | `render/render.go:113-126` |
| `buildLiveView` puts plane-LOCAL Y on every View entity (`Y-planeOffset`); self is server index 0 (skipped from `Players.All()` loop) | `cmd/cradle/spectate.go:184-249, 200-201` |
| `facts.At(x,y) []Placement{Kind,DefID,Name,X,Y,Direction,Extra}`, Kind∈`scenery/boundary/npc_spawn/ground_item`; accessors `SceneryDef/BoundaryDef/NpcDef/ItemDef(id)` return nil-safe ptrs | `facts/facts.go:53-107` |
| Defs: `SceneryDef.Command1/Command2/Name/Description` (`defs.go:8-18`); `BoundaryDef.Command1/Command2` + `BlocksMovement()` (`defs.go:28-56`); `NpcDef.Command1/Command2/Attackable/Name/Description` (`defs.go:62-74`); `ItemDef.Command/IsWearable/IsStackable/Name` (`defs.go:79-92`) | `facts/defs.go` |
| `BoundaryLoc.Direction` (0..3) == `InteractWithBoundary` 3rd arg | `facts/defs.go:103-108`, `runtime/boundary.go:19` |
| Host actions (all `func(ctx,…) error`): `Walk(x,y)` h.go:666; `AttackNpc(idx)` combat:21; `TalkToNpc(idx)` combat:514; `NpcCommand(idx)` boundary:116; `AttackPlayer(idx)` combat:204; `InitTradeRequest(idx)` trade:14; `InitDuelRequest(idx)` duel:14; `Follow(name,timeout)` follow:25; `PickUpItem(x,y,itemID)` items:27; `InteractAt(x,y,option)` boundary:71; `InteractWithBoundary(x,y,dir)` boundary:19; `ItemCommand(slot)` items:68; `DropItem(slot)` items:15; `EquipItem(slot)` magic:42; `UnequipItem(slot)` magic:47; `Say(msg)` h.go:1005; `Command(cmd)` h.go:1000; `PrivateMessage(to,msg)` social:12; `UseItemOn{Item,Scenery,Boundary,GroundItem,Npc,Player}` boundary:34-130 | verified via grep |
| Examines (NO packet, return `runtime.Examination{Kind,Name,Description,X,Y,Detail}`, with `.String()`): `ExamineNpc(idx)` :50, `ExaminePlayer(idx)` :198, `ExamineSelf()` :216, `ExamineGroundItem(x,y)` :138, `ExamineScenery(x,y)` :154, `ExamineBoundary(x,y)` :174, `ExamineInventorySlot(slot)` :118, `ExamineItem(itemID)` :88 | `runtime/examine.go` |
| World reads (lock-safe copies): `Self.Position()/Heading()/CombatLevel()/HP()/MaxHP()/Prayer()/MaxPrayer()/Fatigue()/QuestPoints()/SkillLevel(id)/SkillMax(id)/SkillXP(id)/EquipSprites() [12]int/HasEquip()/ActivePrayers()`; `Inventory.Slots() []InvSlot{ItemID,Amount,Wielded}`; `Npcs.All()/Get(idx)` `NpcRecord{Index,X,Y,TypeID,Heading,...}`; `Players.All()/Get(idx)` `PlayerRecord{Index,Name,X,Y,Heading,CombatLevel,...}`; `GroundItems.All()` `GroundItemRecord{X,Y,ItemID}`; `Scenery.All()` `SceneryRecord{X,Y,ID}` | `world/self.go`, `world/inventory.go:6,60`, `world/world.go:125,326-334,353,599-607,50`, `world/scenery.go:18` |
| `world.NumSkills=18`; `event.SkillName(event.SkillID)` lowercases (id 3=`hits`,5=`prayer`); `event.NumEquipSlots=12`, `event.EquipSlotName(slot)`, `event.EquipSlot*` consts | `world/self.go:44`, `event/events.go:69-93,192-234` |
| Chat sources: `Recent.Chat() *ChatRecord{Speaker,Message,At}`, `Recent.PM() *PMRecord{Sender,Message,At}`, `Recent.ServerMessages() []ServerMsgRecord{Message,Kind,At}` (ring cap 32), all single-slot except the server ring | `world/recent.go:50-99,104-200`, `ServerMsgRingCap=42→32` :39-42 |
| Existing `spectate(ctx,log,cfg,host,land,f)` owns: `walkCh chan [2]int` (depth-1 retarget), `motion *motionTracker`, `bundle`, `buildLiveView`, `renderOne`, `atoiOr`, the `mux`, `spectatePage`; invoked from `main.go:521` gated on `cfg.spectate` | `cmd/cradle/spectate.go:270-553`, `cmd/cradle/main.go:521` |
| `config{username,factsRoot,renderRotation,renderZoom,renderW,renderH,spectate bool,spectateAddr string,...}` | `cmd/cradle/main.go:41-72` |

---

## 1. Contradiction ledger — every cross-doc conflict, resolved

The five layer docs were written in parallel and disagree in seven places. Each
ruling below is final.

### C1 — `MenuTarget` field shape (10 vs 20 vs 30 vs 40)

- **10** (`render.PickCandidate`): `{Kind TargetKind(int); Plane; X,Y; Index; NpcID; ItemID; DefID; Direction; Dynamic; ScreenRect; CamZ}` — Layer-1 internal.
- **20**: `MenuTarget{Kind TargetKind(string); Index; DefID; ItemID; Slot; X; Y; Plane; Dir; Name}`.
- **30**: `MenuTarget{Kind string; Index; X; Y; Dir; ID; Slot}` — `id` collapses DefID/ItemID/TypeID; `Slot` not-omitempty (-1 sentinel).
- **40**: ref `{kind; index; defId; itemId; x; y; dir; slot}` — split id back out.

**RULING — the wire `MenuTarget` is doc 30's shape, with one addition (`name`).**
The JSON-on-the-wire `MenuTarget` is **flat, collapsed-id**:
`{kind string, index int, x int, y int, dir int, id int, slot int, name string}`.
`id` is the single polymorphic def/catalog id (scenery DefID | boundary DefID |
ground-item ItemID | inventory ItemID | npc TypeID-info). `slot` is `-1` when
N/A and is **always emitted** (slot 0 is real). `name` is carried for display +
the one name-keyed Host method (`Follow`) and for staleness messages.
- Doc 40's split `defId`/`itemId` is **rejected** (two fields that are never both
  set; collapsing to `id` is simpler and the UI treats `ref` as opaque anyway —
  doc 40 §5.2 itself says "treats `ref` as opaque", so the internal field names
  don't matter to it). **Doc 40 must be read as using `id`, not `defId/itemId`.**
- Doc 20's string `TargetKind` and doc 10's int `TargetKind` are **distinct
  types in distinct packages** and that is fine (see C3): Layer-1 `render` uses
  an int enum internally; the wire/Layer-2 kind is the **string** in §2 below.
- `render.PickCandidate` (doc 10) stays exactly as doc 10 specifies — it is the
  Layer-1 internal type, NOT the wire type. The handler (Layer 3) maps a
  `render.PickCandidate` → the wire `MenuTarget`, folding plane into absolute Y
  and collapsing `NpcID/DefID/ItemID` → `id`.

### C2 — `MenuOption` wire fields (20 vs 30 vs 40)

- **20**: `{id OptionID(string), label}`.
- **30**: `{id int, text, isDefault bool}`.
- **40**: `{id int, verb}`.

**RULING — `MenuOption` on the wire is `{id int, verb string}` + a separate
`defaultOption int` on the candidate.** Final shape:
- `MenuOption{ID int, Verb string}` (JSON `{"id":0,"verb":"Attack"}`).
- The default action is **`options[0]`** (entry 0 is always the left-click
  default, per overview §4 and doc 20 §2). We do **not** emit a per-option
  `isDefault` flag (doc 30) — ordering carries it, exactly as doc 40 and 20
  intend, which is less error-prone (exactly one default is structural, not a
  field that could be set twice). `verb` (doc 40) is chosen over `text`/`label`
  because the UI composes `verb + " " + label` (doc 40 §2.2/§5.5).
- The string `OptionID` (doc 20, `"attack"`,`"talk_to"`,…) is the **Go-internal
  dispatch key**, NOT on the wire. The wire `id` is the **index into that
  candidate's `options[]`**; Layer 2 maps `(kind, id)` → internal `OptionID` →
  Host call. So `id` is meaningful only paired with its `ref` (doc 30 §2.5).

### C3 — `OptionID` type: string enum vs int index (20 vs 30/40)

**RULING — both, at different layers.** Layer 2 keeps a Go `OptionID string`
enum as the dispatch contract (stable, label-independent — doc 20's key insight).
The wire uses the **int index** into `options[]`. Layer 2 builds the `options[]`
slice once per candidate and remembers, per `(kind, index)`, which `OptionID` it
is — concretely, `BuildMenu` returns both the wire `[]MenuOption` and an internal
`[]OptionID` parallel slice (or the dispatcher re-derives the option list from
the same builder given `(kind, defs)` and indexes it). **Decision: the dispatcher
re-builds the option list deterministically from the `ref` + defs and indexes by
`optionId`** — so the server is stateless between `/pick` and `/act` (no per-pick
session), matching doc 30's stateless-ref principle. This means `BuildMenu` (verb
list) is a **pure function of `(kind, defs-from-ref)`** so `/pick` and `/act`
compute the identical list.

### C4 — Layer-2 package name: `remoteui` vs `dispatch` (20 vs 00)

**RULING — package `remoteclient`, import path
`github.com/gen0cide/westworld/remoteclient`.** Doc 20 proposed `remoteui`; the
overview §4 wrote "(new package, e.g. remoteui/)" and "(e.g. dispatch package)".
`remoteclient` is chosen: it is the obvious sibling name to the
`cmd/cradle/remoteclient.go` handler file, reads correctly (it is the remote
client's domain logic), and avoids "ui" (this package has no UI — that is Layer
4's browser page). **All references to `remoteui` in doc 20 mean
`remoteclient`.**

### C5 — Hosting: evolve `-spectate` vs new `-client` flag (30 vs 40)

- **30** §0: do NOT fork — extend the existing `spectate(...)` func, register new
  routes on its `mux`, serve the SPA at `GET /client`, legacy page stays at `/`.
- **40** §0: a NEW parallel `-client` flag (`cfg.client`/`cfg.clientAddr`), serve
  SPA at `/`, "do not edit `spectate.go`'s page or handlers."

**RULING — new `-client` flag, new `serveClient(...)` function, SPA at `GET /`.**
Doc 40's separation wins, with doc 30's helper-reuse. Concretely:
- A new flag `-client` (`cfg.client bool`, `cfg.clientAddr string`, default
  `localhost:8090`) added beside `-spectate` in `main.go`, invoked from
  `main.go` right after the `cfg.spectate` block.
- A new top-level func **`serveClient(ctx, log, cfg, host, land, f)`** in
  `cmd/cradle/remoteclient.go`. It is a near-copy of `spectate`'s bootstrap
  (waitForLivePosition, OpenBundle, motionTracker, the action worker, the
  graceful-shutdown goroutine, ListenAndServe on `cfg.clientAddr`), reusing the
  **shared helpers** `buildLiveView`, `motionTracker`, `waitForLivePosition`,
  `renderOne`-equivalent, `atoiOr` — all of which already live in `spectate.go`
  in the same `package main`, so `serveClient` calls them directly with **no
  refactor and no edit to `spectate.go`**.
- The SPA is served at `GET /` (doc 40), with `/frame /pos /walk /shot /clip`
  re-registered on the client mux (they are tiny; `serveClient` re-declares them
  exactly as `spectate` does, OR — preferred to avoid duplication — factors the
  frame/pos/walk/shot/clip handler-registration into a shared helper
  `registerViewportRoutes(mux, deps)` in `remoteclient.go` that BOTH could use;
  but since doc 40 forbids editing `spectate.go`, `serveClient` simply
  re-registers them itself by copying the ~5 closures). The legacy `-spectate`
  mode is untouched and keeps `spectatePage` at its own `/`.
- **Why not doc 30 (`/client` on the spectate mux):** keeping the heavy SPA and
  its mutating endpoints (`/act`, `/chat`) entirely out of the read-only
  spectator process is the safer separation doc 40 argues for, and the duplicated
  bootstrap is ~30 lines. The cost of doc 30's approach (the spectator process
  gaining write endpoints) is worse than 30 lines of copy. Both modes still share
  one binary and all Go helpers.

> **Net effect on the file plan:** `cmd/cradle/spectate.go` is **NOT modified**
> (doc 40 §0). All new server code is in `cmd/cradle/remoteclient.go` +
> `cmd/cradle/clientpage.go`, plus a two-line flag add + a four-line invocation
> in `cmd/cradle/main.go`.

### C6 — `/state` JSON shape (30 vs 40)

Doc 30 and doc 40 specify materially different `/state` documents (30 has
`self.skills[18]`, `recentChat` with `at`; 40 has `inventory[30]` always-30,
`equipment` keyed by string slot name, `chat` with monotonic `seq` cursor,
`stats`).

**RULING — the frozen `/state` shape is §5 below**, which takes:
- doc 30's `self` block (rich: hp/prayer/fatigue/skills[18]) — the UI HUD reads a
  subset, the rest is forward-compatible and cheap.
- doc 40's **inventory item carrying its own `options[]` + `defaultOptionId`** (so
  the inventory right-click menu needs no `/pick` round-trip — doc 40 §2.4) and
  doc 30's field set per slot (`wielded/wearable/stackable/command/name/amount`).
  Inventory is emitted **occupied-slots-only** (doc 30) — the UI lays them out
  sparse into a 30-cell grid by `slot` (doc 40 renders 30 cells; it reads sparse
  fine).
- doc 40's `equipment` as `{slot:<label>, name, sprite}` (string slot label) —
  but with doc 30's note that `itemId` is `0`/best-effort and `sprite` is the raw
  sprite id (since only sprite ids are mirrored). Final: `{slot string, sprite int, itemId int}`.
- chat: **doc 40's `seq` monotonic cursor** (the UI appends only `seq >
  lastSeq` — simpler than doc 30's `at`+kind+text dedupe), with doc 30's
  server-side append-only ring as the implementation. Field key is **`chat`**
  (doc 40), not `recentChat` (doc 30). Each line: `{seq int, kind string, who
  string, text string}` where `kind ∈ "public"|"npc"|"private"|"system"|"self"`
  (doc 40's vocabulary; `npc`=dialog text, `system`=server msg).

### C7 — render.PickCandidate `Plane` / `ScreenRect` / `Dynamic` extras (10 vs 20)

Doc 20's "Expected (to be finalised by Layer 1)" `PickCandidate` sketch omits
`Plane`, `ScreenRect`, `CamZ`, `Dynamic`, `NpcID`. **RULING — doc 10's
`PickCandidate` is authoritative and complete** (it is Layer 1's owned type).
Layer 2 (`remoteclient.BuildMenu`) consumes only the subset it needs (`Kind,
Index, NpcID, DefID, ItemID, X, Y, Plane, Direction, Dynamic`). Doc 20's sketch
is superseded by doc 10 §1.2.

---

## 2. Frozen contract A — kinds + the `MenuTarget` wire ref

There are **two** kind enums, deliberately, in two packages:

```go
// render package (Layer 1) — internal int enum, drives which identity fields the
// picker populates. From 10-hit-testing §1.1, UNCHANGED.
type render.TargetKind int
const ( render.TargetNPC; render.TargetPlayer; render.TargetSelf;
        render.TargetGroundItem; render.TargetScenery; render.TargetBoundary;
        render.TargetTerrain )
```

```go
// remoteclient package (Layer 2/3 wire) — string kind, the JSON contract.
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
)
```

The 1:1 name map (`render.TargetNPC` → `"npc"`, etc.) lives in a helper
`kindToWire(render.TargetKind) TargetKind` in `remoteclient`. `inventory_item`
has **no** render counterpart (built by the UI/handler from `/state`).
`equipment_item` from doc 20 is **dropped for M1** — its only verb is `Remove`,
which is `inventory_item`'s `remove` on the same underlying slot (doc 20 §1's
equipment fact). The equipment panel acts via an `inventory_item` ref. (Reserved
for M2; adding it is a new const + rows, contract-1 clean.)

**The wire `MenuTarget` (frozen — C1):**

```go
// MenuTarget is the self-describing ref /pick emits and /act consumes. Flat,
// value-typed, round-trips losslessly. The server is stateless between /pick and
// /act: it re-resolves volatile identity (Index, Slot) against the live world at
// /act time. Unknown extra JSON keys are ignored (the UI may attach a cached
// label without breaking /act).
type MenuTarget struct {
    Kind  TargetKind `json:"kind"`            // "npc"|"player"|"self"|"ground_item"|"scenery"|"boundary"|"terrain"|"inventory_item"
    Index int        `json:"index,omitempty"` // server actor index (npc|player); 0 otherwise
    X     int        `json:"x"`               // ABSOLUTE world X (plane folded in for tile targets)
    Y     int        `json:"y"`               // ABSOLUTE world Y (plane folded in)
    Dir   int        `json:"dir,omitempty"`   // boundary edge dir 0..3 (boundary only)
    ID    int        `json:"id,omitempty"`    // scenery DefID | boundary DefID | ground/inv ItemID | npc TypeID(info)
    Slot  int        `json:"slot"`            // inventory slot 0..29; -1 if N/A (NOT omitempty)
    Name  string     `json:"name,omitempty"`  // display name; the Follow method's key; staleness messages
}
```

Per-kind field population (the only fields read at dispatch):

| kind | dispatch reads | advisory (labels/staleness) |
|---|---|---|
| `npc` | `Index` | `X,Y,ID(=TypeID),Name` |
| `player` | `Index`, `Name` (Follow) | `X,Y` |
| `self` | `X,Y` | — |
| `ground_item` | `X,Y,ID(=ItemID)` | `Name` |
| `scenery` | `X,Y` (+`ID` for the verb list) | `Dir,Name` |
| `boundary` | `X,Y,Dir` | `ID,Name` |
| `terrain` | `X,Y` | — |
| `inventory_item` | `Slot` (re-validated vs `ID`) | `ID(=ItemID),Name` |

> **Y is ABSOLUTE on the wire** (plane folded in). `render.PickCandidate.Y` is
> **plane-local**; the Layer-3 handler adds `Plane*world.PlaneHeight` when it maps
> a candidate → `MenuTarget`, exactly as `/walk` does (`spectate.go:380`). At
> dispatch the absolute `Y` flows straight into the Host method (every Host action
> takes absolute coords).

**Boundary `Dir` (0..3)** is the `createModel` edge encoding (View doc comment
`render/render.go:90`), == `facts.At`'s `Placement.Direction` for a boundary, ==
`InteractWithBoundary`'s 3rd arg. No remapping.

---

## 3. Frozen contract B — the menu + option wire types

```go
// MenuOption is one verb row. id is the index into THIS candidate's options[]
// (dispatch-table-relative to kind, NOT a global enum). verb is the authentic
// RSC label. options[0] is ALWAYS the left-click default (no isDefault flag).
type MenuOption struct {
    ID   int    `json:"id"`
    Verb string `json:"verb"`
}

// Candidate is one entry in /pick's depth-ordered list. ref round-trips to /act.
type Candidate struct {
    Ref     MenuTarget   `json:"ref"`
    Kind    TargetKind   `json:"kind"`     // == Ref.Kind, duplicated for UI switch convenience
    Label   string       `json:"label"`    // display name (Examination.Name / "Object" / "Item N")
    Examine string       `json:"examine"`  // Examination.Description ("" if none)
    Detail  string       `json:"detail"`   // Examination.Detail ("" if none)
    Dist    int          `json:"dist"`     // chebyshev self->(ref.x,ref.y); UI ordering aid
    Options []MenuOption `json:"options"`  // >=1; options[0] = default
}

// PickResponse is /pick's body.
type PickResponse struct {
    Candidates []Candidate `json:"candidates"` // depth-ordered nearest-first; terrain last; never null
}
```

Internal (NOT on the wire) — Layer 2's stable dispatch key:

```go
type OptionID string
const (
    OptCommand1 OptionID = "command1"   // scenery/boundary/npc primary def verb
    OptCommand2 OptionID = "command2"
    OptAttack   OptionID = "attack"
    OptTalkTo   OptionID = "talk_to"
    OptTrade    OptionID = "trade"
    OptFollow   OptionID = "follow"
    OptDuel     OptionID = "duel"
    OptPickup   OptionID = "pickup"
    OptCommand  OptionID = "command"     // inventory default verb (Eat/Drink/Bury)
    OptWield    OptionID = "wield"
    OptRemove   OptionID = "remove"
    OptDrop     OptionID = "drop"
    OptUse      OptionID = "use"         // M2 placeholder (no-op in M1)
    OptWalkHere OptionID = "walk_here"
    OptExamine  OptionID = "examine"
)
```

**The verb-list builder is a pure function of `(kind, defs)`** so `/pick` and
`/act` produce the identical option ordering (C3). Per-kind ordered lists
(authentic RSC order, from doc 20 §2, `Examine` last, de-dup rules applied):

| kind | ordered `(OptionID, Verb)` list |
|---|---|
| `npc` | `command1` if `NpcDef.Command1!=""`; `command2` if `Command2!=""`; `attack` if `Attackable` and not already named; then `examine`. (monster→`[Attack,Examine]`; shop NPC→`[Talk-to,Examine]`) |
| `player` | `[trade:"Trade with", follow:"Follow", duel:"Duel with", attack:"Attack", examine:"Examine"]` (fixed) |
| `self` | `[examine:"Examine"]` (walk falls through to terrain) |
| `ground_item` | `[pickup:"Pick up", examine:"Examine"]` |
| `scenery` | `command1` if `SceneryDef.Command1!=""`; `command2` if `Command2!=""` AND not case-insensitively `"Examine"`; then `examine` |
| `boundary` | `command1` if `BoundaryDef.Command1!=""` and not `"WalkTo"`; `command2` if `Command2!=""` and not `"Examine"`; then `examine` |
| `terrain` | `[walk_here:"Walk here"]` |
| `inventory_item` | `command` if `ItemDef.Command!=""` and not Wield/Wear; `wield:"Wield"` if `IsWearable && !Wielded`; `remove:"Remove"` if `Wielded`; `drop:"Drop"`; `examine:"Examine"`. (fold a Wield/Wear `Command` into the `wield` entry, not both) |

`Label` is `Examination.Name` (which the builder resolves via the matching
`Examine*` call) with fallbacks `"Object"`/`"Item N"`/player name. `Examine` text
is pre-resolved into `Candidate.Examine`/`Detail` (free — no packet) so the UI
needs no `/examine` round-trip for the menu (doc 20 §3, doc 30 §4).

---

## 4. Frozen contract C — the dispatch table (concrete Go mapping)

`Dispatch(ctx, target MenuTarget, optionId int) (message string, err error)`
re-derives the option list for `target.Kind` (from defs via `target.ID`), maps
`options[optionId].OptionID` → the Host call. The complete `(kind, OptionID) →
call` table (verified signatures, §0):

```
npc:
  command1 -> verb-route on NpcDef.Command1:
                "Attack" -> AttackNpc(ctx, t.Index)
                "Talk-to"-> TalkToNpc(ctx, t.Index)
                else     -> NpcCommand(ctx, t.Index)        // Pickpocket etc.
  command2 -> same verb-route on NpcDef.Command2
  attack   -> AttackNpc(ctx, t.Index)
  talk_to  -> TalkToNpc(ctx, t.Index)
  examine  -> ExamineNpc(t.Index)                            // no packet; msg=String()
player:
  trade    -> InitTradeRequest(ctx, t.Index)
  follow   -> FOLLOW LANE (own goroutine): Follow(ctx, t.Name, followStartupTimeout)
  duel     -> InitDuelRequest(ctx, t.Index)
  attack   -> AttackPlayer(ctx, t.Index)
  examine  -> ExaminePlayer(t.Index)                         // no packet
self:
  examine  -> ExamineSelf()                                  // no packet
ground_item:
  pickup   -> PickUpItem(ctx, t.X, t.Y, t.ID)
  examine  -> ExamineGroundItem(t.X, t.Y)                    // no packet
scenery:        // option arg is 1-BASED
  command1 -> InteractAt(ctx, t.X, t.Y, 1)
  command2 -> InteractAt(ctx, t.X, t.Y, 2)
  examine  -> ExamineScenery(t.X, t.Y)                       // no packet
boundary:       // both commands = same packet (server toggles open/close)
  command1 -> InteractWithBoundary(ctx, t.X, t.Y, t.Dir)
  command2 -> InteractWithBoundary(ctx, t.X, t.Y, t.Dir)
  examine  -> ExamineBoundary(t.X, t.Y)                      // no packet
terrain:
  walk_here-> Walk(ctx, t.X, t.Y)                            // via the walk-retarget lane
inventory_item: // slot is 0-BASED; re-validate Slots()[t.Slot].ItemID==t.ID first
  command  -> ItemCommand(ctx, t.Slot)
  wield    -> EquipItem(ctx, t.Slot)
  remove   -> UnequipItem(ctx, t.Slot)
  drop     -> DropItem(ctx, t.Slot)
  use      -> NO-OP in M1 (return msg "Use-on is M2", ok:true)  // verb reserved
  examine  -> ExamineInventorySlot(t.Slot)                   // no packet
```

**Three execution lanes** (doc 20 §5, doc 30 §9), all in the `actCh` worker model
(§6 below):
1. **Walk-retarget lane** — `terrain`/`walk_here`: fire-and-forget, drain-then-send
   on a depth-1 channel (the existing `walkCh` semantics). Responds immediately.
2. **Single-action lane** — every interaction EXCEPT walk + follow: enqueue on the
   serialized worker, wait on a `done` chan with a ~1.5s soft timeout (on timeout
   return `{"ok":true,"message":"<verb> (running)"}`).
3. **Follow lane** — `Follow` runs on its own cancellable goroutine (it blocks
   until ctx cancel — `follow.go:25`). Stored as the active follow; any new
   single-action/walk lane action or `/act follow` again cancels it.
4. **Examine** — never enqueued; resolved synchronously (no packet); `message` =
   `Examination.String()`, `ok:true`.

**0-based vs 1-based cheat sheet (doc 20 §6):** inventory slot 0-based;
`InteractAt` option **1-based** (command1→1, command2→2 — the ONLY 1-based
ordinal); boundary dir is an enum 0..3 (not an ordinal); actor `Index` opaque.

**Testability interface (doc 20 §7):** `Dispatcher` accepts a narrow
`ActionHost interface` (+ `ExamineHost interface`) listing exactly the verified
methods above, satisfied by `*runtime.Host`, so `dispatch_test.go` mocks routing.

---

## 5. Frozen contract D — `/state` JSON (the §C6 resolution)

```jsonc
{
  "self": {
    "x": 331, "y": 552, "plane": 0, "heading": 4,
    "combatLevel": 31,
    "hp": 18, "maxHp": 20,
    "prayer": 12, "maxPrayer": 13,
    "fatigue": 4210, "questPoints": 9,
    "skills": [ { "id": 0, "name": "attack", "level": 20, "max": 20, "xp": 4470 } /* …18 total, id==index, name=event.SkillName(id) */ ]
  },
  "inventory": [                       // OCCUPIED slots only, slot order
    {
      "slot": 0, "itemId": 77, "name": "Bronze sword", "amount": 1,
      "wielded": false, "wearable": true, "stackable": false, "command": "Wield",
      "defaultOptionId": 0,            // index into options[] of the left-click verb
      "options": [ {"id":0,"verb":"Wield"}, {"id":1,"verb":"Drop"}, {"id":2,"verb":"Examine"} ]
    }
  ],
  "equipment": [ { "slot": "Weapon", "sprite": 13, "itemId": 0 } ],   // sprite>0 only
  "chat": [                            // chronological (oldest-first), capped ~64
    { "seq": 1201, "kind": "public",  "who": "Zezima", "text": "hi" },
    { "seq": 1202, "kind": "npc",     "who": "Man",    "text": "Hello there!" },
    { "seq": 1203, "kind": "private", "who": "Bob",    "text": "trade?" },
    { "seq": 1204, "kind": "system",  "who": "",       "text": "You can't reach that." },
    { "seq": 1205, "kind": "self",    "who": "bernard","text": "selling lobsters" }
  ]
}
```

Field sources (all lock-safe `world` accessors, §0):
- `self.*`: `Self.Position()/Heading()/CombatLevel()/HP()/MaxHP()/Prayer()/MaxPrayer()/Fatigue()/QuestPoints()`; `skills[i]` = `{SkillLevel(i),SkillMax(i),SkillXP(i),event.SkillName(event.SkillID(i))}` for `i in 0..NumSkills-1`.
- `inventory[]`: occupied `Inventory.Slots()` + `facts.ItemDef`. `options`/`defaultOptionId` come from the **same `BuildMenu(KindInventoryItem, …)`** used by `/pick` (so the inventory right-click menu needs no server round-trip — doc 40 §2.4). Name fallback `"item N"`.
- `equipment[]`: `Self.EquipSprites() [12]int`; emit `slot=event.EquipSlotName(i), sprite=v, itemId=0` for each `v>0`. (M2 enriches `itemId`; shape is forward-compatible.)
- `chat[]`: the server-session **chat ring** (cap ~64) the handler maintains in `serveClient`'s closure. It is grown from (a) outgoing `/chat` sends (kind `"self"`), and (b) on each `/state`, a de-duped tail of `Recent.Chat()` (kind `"public"`, who=Speaker), `Recent.PM()` (kind `"private"`, who=Sender), newest `Recent.ServerMessages()` (kind `"system"` or `"npc"` per the record's own `Kind` field: `"dialog"`→`"npc"`, else `"system"`). `seq` is a monotonic counter the ring assigns; the UI appends only `seq > lastSeq`.

`Cache-Control: no-store`. `503` only while position is unloaded (login);
otherwise `200`. **M2 adds top-level nullable keys** (`bank/shop/trade/duel/
dialog/spells/prayers/social/pendingTarget`) — never edits these (backlog §0
contract 3); the M1 shape is the floor.

---

## 6. Frozen contract E — endpoints, server entry, shared symbols

### 6.1 Endpoint map (the M1 routes `serveClient` registers on its mux)

| Method | Path | Body / query | Response | Mutates |
|---|---|---|---|---|
| GET | `/` | — | `clientPage` HTML | no |
| GET | `/frame` | `?rot&zoom&w&h&anim&t` | `image/png` | no |
| GET | `/pos` | — | `{x,y,plane}` | no |
| GET | `/walk` | `?px&py&rot&zoom&w&h` | `{x,y}` | yes (walk lane) |
| GET | `/shot`,`/clip` | `?rot&zoom&w&h` | `{saved,…}` | no |
| POST | `/pick` | `PickRequest` | `PickResponse` | no |
| POST | `/act` | `ActRequest` | `ActResponse` | yes |
| GET | `/state` | — | §5 doc | no |
| POST | `/chat` | `ChatRequest` | `{ok,message?}` | yes |
| GET/POST | `/examine` | `?ref=<urlenc json>` or `{ref}` | `Examination` JSON | no |

`/frame /pos /walk /shot /clip` behave EXACTLY as in `spectate.go` (re-registered
verbatim by `serveClient`). New request/response Go types (in
`cmd/cradle/remoteclient.go`):

```go
type PickRequest struct { PX, PY, Rot, Zoom, W, H int; Slot int } // Slot for inv-fallback; -1 default
type ActRequest  struct { Ref remoteclient.MenuTarget `json:"ref"`; OptionID int `json:"optionId"` }
type ActResponse struct { OK bool `json:"ok"`; Message string `json:"message,omitempty"` }
type ChatRequest struct { Kind, Text, To string } // kind: "say"|"command"|"pm"
```

`/pick` rebuilds the plane-local `render.View` from the host's CURRENT position +
the request camera (identical to the `/walk` handler), calls
`render.Pick(land, f, v, px, py)`, maps each `render.PickCandidate` →
`remoteclient.Candidate` (folding `Plane*PlaneHeight` into absolute `ref.Y`,
collapsing ids → `ref.ID`, running `BuildMenu`). `/act` decodes the ref,
re-validates volatile identity, calls `remoteclient.Dispatch`. `/examine`
dispatches `ref.kind` → the matching `Examine*` and returns the lowercased
`Examination`.

### 6.2 Server entry func (the symbol `main.go` calls)

```go
// serveClient serves the full remote-client SPA + JSON API. Mirrors spectate()'s
// bootstrap (waitForLivePosition, OpenBundle, motionTracker, action worker,
// graceful shutdown) but adds the M1 client routes and serves clientPage at "/".
// Reuses the package-main helpers buildLiveView, motionTracker, waitForLivePosition,
// atoiOr. Blocks until ctx is cancelled.
func serveClient(ctx context.Context, log *slog.Logger, cfg config,
    host *runtime.Host, land *pathfind.Landscape, f *facts.Facts) error
```

Invoked from `main.go` after the `cfg.spectate` block:
```go
if cfg.client {
    if err := serveClient(rootCtx, log, cfg, host, loadedLandscape, loadedFacts); err != nil {
        log.Warn("client failed", "err", err)
    }
}
```

### 6.3 Shared symbols the layers rely on (frozen names)

| Symbol | Kind | Owner file | Used by |
|---|---|---|---|
| `clientPage` | `const string` (raw-string HTML; `fmt.Fprintf` verbs `%s`=username, `%d`=zoom, `%d`=w, `%d`=h, in that order; literal `%`→`%%`) | `cmd/cradle/clientpage.go` | `remoteclient.go` `GET /` handler |
| `serveClient` | func (§6.2) | `cmd/cradle/remoteclient.go` | `cmd/cradle/main.go` |
| `cfg.client bool`, `cfg.clientAddr string` | config fields + flags `-client`/`-client-addr` (default `localhost:8090`) | `cmd/cradle/main.go` | flag block + invocation |
| `github.com/gen0cide/westworld/remoteclient` | package (Layer 2) | `remoteclient/*.go` | `remoteclient.go` handler |
| `remoteclient.MenuTarget`, `.TargetKind`+`Kind*` consts, `.MenuOption`, `.Candidate`, `.PickResponse` | wire types (§2,§3) | `remoteclient/target.go` | handler + dispatch + (the UI mirrors them in JS) |
| `remoteclient.BuildMenu(kind, …) (label, examine, detail string, opts []MenuOption)` | pure verb-list builder (§3) | `remoteclient/menu.go` | `/pick`, `/state` inventory, `/act` |
| `remoteclient.BuildCandidates(host, f, cands []render.PickCandidate, selfX, selfY, plane int) []Candidate` | candidate→menu mapper | `remoteclient/menu.go` | `/pick` |
| `remoteclient.NewDispatcher(host ActionHost, examine ExamineHost) *Dispatcher` + `(*Dispatcher).Dispatch(ctx, MenuTarget, optionId) (string, error)` | dispatch (§4) | `remoteclient/dispatch.go` | `/act` |
| `render.Pick`, `render.PickCandidate`, `render.TargetKind`+consts | Layer-1 API (doc 10) | `render/pick.go` | `remoteclient` + `/pick` |
| `render.billboardCamera`, `render.projectBillboard` | shared projection (doc 10 §5; unexported) | `render/hittest.go` | `render.Pick` + `DrawEntitySprites` |

### 6.4 The action worker (generalised, doc 30 §9 + backlog §8.1)

`serveClient` runs **one `actCh chan actReq`** worker (replacing the bare
`walkCh chan [2]int`), depth-1:
```go
type actReq struct {
    run      func(ctx context.Context) (string, error)
    done     chan actResult // nil = fire-and-forget retarget (walk)
    coalesce bool           // true = drain-then-send (walk); false = serialize
}
type actResult struct{ msg string; err error }
```
Walk uses `coalesce:true, done:nil` (preserves spam-click retarget). Interactions
use `coalesce:false` + a `done` chan with a 1.5s soft timeout. This is the
backlog §8.1 `func() error` queue (M1 has one bot; the closure-queue shape is the
forward-compatible decision). **Follow** does NOT use `actCh` — own goroutine.

---

## 7. FILE PLAN

Model column: **opus** = correctness-critical core (the projection math + the
dispatch routing — a wrong number or a mis-routed packet is a silent, hard-to-
catch bug). **sonnet** = mechanical (HTTP glue that parses→calls→encodes, and the
browser page that follows doc 40 field-by-field).

### Layer 1 — `render.Pick` (opus)

1. **`render/hittest.go`** — CREATE, opus.
   `func billboardCamera(land *pathfind.Landscape, v *View) (cam Camera, baseX, baseY, cx, cy int, heights [][]int32)`;
   `func projectBillboard(cam Camera, cx, cy int, heights [][]int32, lx, ly int, ox, oz int32, worldW, worldH int) (rect [4]int, sx, feetY, camZ int32, ok bool)`.
   Extract the foot-projection closure (`render.go:415-447`) and the
   `screenW/screenH` rect block (`render.go:539-554`) VERBATIM into these two
   pure helpers; bake in the `buildScene` defaults (Zoom 750 / W 512 / H 334, NOT
   PickTile's 600/336); swap yaw/roll EXACTLY once inside `billboardCamera`
   (callers pass `SetCamera`'s un-swapped output). The blit depth bias
   (`spriteDepthBias`) stays in `DrawEntitySprites`, not here.

2. **`render/render.go`** — MODIFY, opus.
   Rewrite `DrawEntitySprites`'s inline `project` closure + per-item
   `screenW/screenH` computation to call `billboardCamera`/`projectBillboard`.
   Keep the far-first painter sort (`camZ` DESC) and the `camZ-spriteDepthBias`
   blit depth unchanged. The rendered framebuffer must stay **byte-identical**
   (guarded by `render/fidelity_*_test.go`).

3. **`render/pick.go`** — MODIFY, opus.
   Add `TargetKind` int enum + consts; `PickCandidate` struct (doc 10 §1.2,
   complete: `Kind,Plane,X,Y,Index,NpcID,ItemID,DefID,Direction,Dynamic,ScreenRect,CamZ`);
   `func Pick(land *pathfind.Landscape, f *facts.Facts, v View, px, py int) []PickCandidate`.
   Billboards via the shared helper (skip `composite*==nil`), sorted
   nearest-camera-first (stable); then tile targets via `PickTile` +
   `facts.At(tx, ty+Plane*planeHeightTiles)` + the View's `DynamicScenery` /
   `BoundaryRemoved` / `SceneryRemoved`, in fixed priority dynamic-scenery →
   static-scenery → boundary → terrain. `f` may be nil (billboards+terrain only);
   `land` may be nil (billboards only). Returns plane-LOCAL X/Y.

4. **`render/hittest_test.go`** — CREATE, opus.
   The doc 10 §6 suite: projection round-trip (`TestProjectBillboardMatchesProjectPoint`),
   AABB math + containment (`TestBillboardAABB`, `TestBillboardAABBContainment`),
   clip/zero-size rejects, ordering (`TestPickOrderingNearestFirst`,
   `TestPickStableTieOrder`, `TestPickTerrainAlwaysLast`), identity completeness
   (NPC/ground-item/boundary/scenery, the last two with a hand-built `*facts.Facts`
   fixture), and the renderer-parity golden `TestDrawEntitySpritesAABBUnchanged`.
   Pure integer math — no archives, no PNGs.

### Layer 2 — `remoteclient` package (opus for dispatch+menu; types are small)

5. **`remoteclient/target.go`** — CREATE, opus.
   `TargetKind` string + `Kind*` consts (§2); `MenuTarget` (§2); `MenuOption`,
   `Candidate`, `PickResponse` (§3); `OptionID` string + `Opt*` consts (§3);
   `kindToWire(render.TargetKind) TargetKind`. Pure types + the kind map; no
   logic. (opus only because the field tags/sentinels are the cross-layer
   contract and must match §2/§3 exactly.)

6. **`remoteclient/menu.go`** — CREATE, opus.
   `BuildMenu(kind TargetKind, defs … ) (label, examine, detail string, opts []MenuOption, ids []OptionID)` — the pure per-kind verb-list builder (§3 table), de-dup `Examine`, fold Wield-`Command`, `Examine` last. `BuildCandidates(ex ExamineHost, f *facts.Facts, cands []render.PickCandidate, selfX, selfY, plane int) []Candidate` — maps each `render.PickCandidate` → `Candidate` (fold plane→absolute Y on the ref, collapse `NpcID/DefID/ItemID`→`ref.ID`, set `ref.Name`/`Label`/`Examine`/`Detail` via the right `Examine*`, compute `Dist`, attach `BuildMenu` options). Also exposes `InventoryMenu(slot int, s world.InvSlot, def *facts.ItemDef) (Candidate)` so `/state` builds inventory `options[]`/`defaultOptionId` from the SAME builder. The builder ordering must be deterministic (so `/act` re-derives the identical `ids[]` and indexes `optionId` into it).

7. **`remoteclient/dispatch.go`** — CREATE, opus.
   `ActionHost` + `ExamineHost` interfaces (the exact verified method set, §0/§4);
   `Dispatcher{host ActionHost, ex ExamineHost}`; `NewDispatcher(...)`;
   `Dispatch(ctx, t MenuTarget, optionId int) (message string, err error)` — re-derives the option list for `t.Kind` (via `BuildMenu` + `t.ID`), maps `ids[optionId]` → the §4 Host call (verb-route NPC command1/2; 1-based `InteractAt`; 0-based slots; boundary same-packet; examine returns `String()` no-packet). Returns a typed stale-target sentinel (`ErrStaleTarget`) when re-validation fails, so the handler emits `{ok:false}` not an HTTP error. `Follow` and walk are flagged to their lanes by the CALLER (handler owns the worker); `Dispatch` returns a `Lane`/`Follow` hint or the handler inspects the resolved `OptionID`. **Decision: `Dispatch` returns `(message string, lane Lane, err error)`** where `Lane ∈ {LaneAction, LaneWalk, LaneFollow, LaneSync}` so the handler routes to the right worker lane without re-deriving.

8. **`remoteclient/menu_test.go`** — CREATE, opus.
   Table-driven: each def shape (monster/shop-NPC/thievable; tree/door/wall;
   food/weapon/worn item) → expected ordered `[]MenuOption` + `[]OptionID`,
   asserting de-dup and Wield-fold.

9. **`remoteclient/dispatch_test.go`** — CREATE, opus.
   `(kind, optionId)` → assert the correct `ActionHost` method + args via a mock
   ActionHost; cover the NPC verb-route branches, 1-based `InteractAt`, 0-based
   slot re-validation, boundary same-packet, examine-no-packet, and the `Lane`
   returned for walk/follow/action/sync.

### Layer 3 — HTTP handlers (sonnet)

10. **`cmd/cradle/remoteclient.go`** — CREATE, sonnet.
    `serveClient(...)` (§6.2): copy `spectate`'s bootstrap, re-register
    `/frame /pos /walk /shot /clip` (copy the closures), add `/pick /act /state
    /chat /examine`, run the §6.4 `actCh` worker + the chat ring (§5) + the active-
    follow cancel slot. Request/response types `PickRequest/ActRequest/ActResponse/
    ChatRequest` (§6.1). Each handler is THIN: decode → call `render.Pick` /
    `remoteclient.BuildCandidates` / `remoteclient.Dispatch` / world accessors →
    encode. `GET /` serves `clientPage` via
    `fmt.Fprintf(w, clientPage, cfg.username, cfg.renderZoom, cfg.renderW, cfg.renderH)`.
    All JSON `Cache-Control: no-store`. Mirror `spectate.go`'s style/comment tone.

11. **`cmd/cradle/main.go`** — MODIFY, sonnet.
    Add `cfg.client bool` + `cfg.clientAddr string` to `config`; two `flag.*Var`
    lines (`-client`, `-client-addr` default `localhost:8090`); the four-line
    invocation block after the `cfg.spectate` block (§6.2).

### Layer 4 — Browser SPA (sonnet)

12. **`cmd/cradle/clientpage.go`** — CREATE, sonnet.
    `const clientPage = ` raw-string HTML+CSS+JS implementing doc 40 §1–§5
    field-for-field. ONLY the page asset — no `net/http`, no handlers, no Go
    logic (doc 40 §4 file-boundary rule). `fmt.Fprintf` verb order
    `%s,%d,%d,%d` (username,zoom,w,h); literal `%`→`%%`. JS consumes the §2/§3/§5
    shapes: viewport (left-click default-act, right-click `#menu` from
    `/pick`), inventory grid (left=defaultOptionId, right=`options[]` from
    `/state`), equipment summary, chat box (`::`/`/`→command, `@name`→pm,
    else→say; append `seq>lastSeq`; ignore camera keys while chat focused). Reuse
    `spectatePage`'s frame loop + `screenToFrame` un-letterbox math verbatim.

### Doc

13. **`docs/remote-client/99-build-log.md`** — CREATE, sonnet (at the end).
    Record what was built + the M1 gate results (`go build/vet/test`).

---

## 8. BUILD ORDER & DEPENDENCY GRAPH

```
render/hittest.go ─┐
                   ├─> render/render.go (DrawEntitySprites calls the helper; fidelity tests guard)
render/pick.go ────┘   └─> render/hittest_test.go  ........... [Layer 1 GATE: go test ./render/...]
        │
        ▼  (render.Pick + render.PickCandidate exist)
remoteclient/target.go
        ▼
remoteclient/menu.go ──> remoteclient/menu_test.go
        ▼
remoteclient/dispatch.go ──> remoteclient/dispatch_test.go  .. [Layer 2 GATE: go test ./remoteclient/...]
        │
        ▼  (BuildCandidates + Dispatch + wire types exist)
cmd/cradle/clientpage.go   (independent of Go API; can be written in parallel — it only
        │                   needs the FROZEN §2/§3/§5 JSON shapes from THIS doc)
        ▼
cmd/cradle/remoteclient.go ──> cmd/cradle/main.go (flag + invoke)
        │                                          [M1 GATE: go build ./... ; go vet ./... ;
        ▼                                           go test ./render/... ./runtime/... ./remoteclient/...]
docs/remote-client/99-build-log.md
```

**Strict ordering rules:**
- Layer 1 (files 1–4) first and complete — `remoteclient` imports `render.Pick`/
  `render.PickCandidate`, which don't exist until file 3. File 2's refactor must
  keep `render/fidelity_*_test.go` byte-identical BEFORE proceeding.
- `remoteclient/target.go` (file 5) before `menu.go`/`dispatch.go` (6/7) — they
  use its types. `menu.go` before `dispatch.go` (dispatch re-derives via
  `BuildMenu`).
- `clientpage.go` (file 12) depends ONLY on this doc's frozen JSON (§2/§3/§5), not
  on any Go symbol, so it can be authored in parallel with Layers 1–2. It is
  wired in only by `remoteclient.go` (file 10) referencing the `clientPage`
  const.
- `cmd/cradle/remoteclient.go` (file 10) last among code — it imports
  `remoteclient`, `render`, references `clientPage`, and reuses `spectate.go`'s
  package-main helpers. `main.go` (file 11) wires the flag.
- **Do NOT modify**: `cmd/cradle/spectate.go` (doc 40 §0), `render/camera.go`,
  any `world/`/`runtime/`/`facts/` file, anything under `reference/rsc-client/**`.

**Cross-layer invariants that must hold at the M1 gate (round-trip proofs):**
- `/pick` emits a `ref` (`MenuTarget`) that `/act` consumes verbatim with no
  client reconstruction (C1 — flat collapsed-id, `slot:-1` sentinel always
  present).
- `options[optionId].verb` shown by the UI and the Host call fired by `Dispatch`
  agree because BOTH index the SAME deterministic `BuildMenu(kind, defs)` list
  (C3).
- `render.PickCandidate.Y` (plane-local) → `ref.Y` (absolute, `+Plane*PlaneHeight`)
  → Host call (absolute) — the single conversion lives in `BuildCandidates`
  (handler side), mirroring `spectate.go:380`.
- `/state` inventory `options[]` are built by the same `BuildMenu` as `/pick`'s,
  so the inventory right-click menu and a hypothetical world-pick of the same item
  would offer identical verbs.
