# Remote Client — Layer 3: the HTTP/JSON API

**Status:** Design (M1)
**Layer:** 3 of 4 (see `00-overview.md` §4)
**Implements against:** Layer 1 (`render.Pick`, `10-hit-testing.md`), Layer 2
(menu + dispatch, `20-menu-dispatch.md`), and the verified `runtime.Host` /
`world` / `facts` surface (§2 of the overview).
**Consumed by:** Layer 4 (the browser SPA, `40-browser-ui.md`).

This document is the **wire contract**. It specifies every endpoint, its method,
path, request schema, and response schema, with exact field names and JSON
types. The browser UI codes directly against these shapes; Layers 1 and 2 codes
against the Go structs these shapes serialize from. Where this doc and `10` /
`20` disagree on a field name, `50-impl-spec.md` reconciles — but the intent is
that they never disagree, so this file is deliberately exhaustive.

> Conventions used throughout
> - All bodies are `application/json; charset=utf-8` unless noted (frames are
>   `image/png`).
> - JSON field names are **lowerCamelCase**. Go structs use `json:"..."` tags to
>   match; do not rely on Go's default capitalised export names.
> - Integer world coordinates are **absolute** (plane folded in: `Y = localY +
>   plane*world.PlaneHeight`) unless a field is explicitly documented as
>   plane-local. `/pos` and refs carry `plane` separately so the client never has
>   to do the arithmetic.
> - Every handler is **thin**: parse request → call Layer 1/2 (or a `world`/
>   `runtime` accessor) → encode JSON. No game logic lives in the handler.
> - Read endpoints are safe to call concurrently from many HTTP goroutines
>   (every `world` accessor RLocks + copies). **Mutating** endpoints (`/walk`,
>   `/act`, `/chat`) enqueue onto the single action worker — see §7.

---

## 0. Hosting decision: evolve `-spectate`, do not fork

The full client is **served by the same `spectate(...)` function** in
`cmd/cradle/spectate.go`, extended with the new routes below. Rationale:

- `/frame`, `/pos`, `/walk`, `/shot`, `/clip`, `buildLiveView`, `motionTracker`,
  the serialized walk worker, and the model `bundle` are all already wired there
  and must be reused verbatim (overview §6).
- A second flag would duplicate all of that bootstrap (waiting for live
  position, opening `models.orsc`, the graceful-shutdown goroutine).

The existing `-spectate` flag keeps working unchanged: the new routes are
additive and the legacy `spectatePage` HTML is left as the `GET /` default. The
full-client SPA (Layer 4) is served at `GET /client` (a second HTML const), so
both UIs share one running server and one logged-in `Host`. New handler code
should live in a new file `cmd/cradle/remoteclient.go` and be registered onto
the same `mux` from inside `spectate(...)`; this keeps `spectate.go` focused and
satisfies the overview's "new file" note without forking the bootstrap.

The shared per-server objects the new handlers close over (all already
constructed in `spectate`): `host *runtime.Host`, `land *pathfind.Landscape`,
`f *facts.Facts`, `bundle` (models), `cfg`, `log`, the `motion *motionTracker`,
and the `walkCh` action queue (generalised in §7).

---

## 1. Endpoint map

| Method | Path        | Purpose                                   | Mutates? | Status |
|--------|-------------|-------------------------------------------|----------|--------|
| GET    | `/`         | legacy spectator HTML page                | no       | exists |
| GET    | `/client`   | full-client SPA HTML page                 | no       | NEW    |
| GET    | `/frame`    | one rendered PNG of the host's view       | no       | exists |
| GET    | `/pos`      | host tile + plane (HUD)                    | no       | exists |
| POST   | `/walk`     | screen click → tile → `host.Walk`         | yes      | exists\* |
| GET    | `/shot`     | save current frame to disk (agent aid)    | no\*\*   | exists |
| GET    | `/clip`     | save a 12-frame burst contact sheet       | no\*\*   | exists |
| POST   | `/pick`     | screen click → ordered menu candidates    | no       | NEW    |
| POST   | `/act`      | execute one menu option on one target     | yes      | NEW    |
| GET    | `/state`    | self + inventory + equipment + chat       | no       | NEW    |
| POST   | `/chat`     | send public / command / private message   | yes      | NEW    |
| GET    | `/examine`  | examine one ref (no packet)               | no       | NEW    |

\* `/walk` exists today as `GET` with query params. It stays as-is for the
legacy page; the new SPA may equally call it, OR route walks through `/act` with
a `terrain` target (§5). Both reach the same `walkCh`. Documented as-is in §3.

\*\* `/shot` and `/clip` write PNGs to `spectatorShotDir` (`/tmp/render_out/
spectator`) for the controlling agent to read; they do not mutate game state.

---

## 2. The `MenuTarget` ref wire format (central to `/pick` → `/act`)

`/pick` returns candidates and `/act` consumes one. To make `/pick` output feed
**directly** into `/act` input with no client-side reconstruction, a target is an
**explicit, self-describing JSON object** — not an opaque server token. The
browser stores the `ref` it received and POSTs it back verbatim. This keeps the
server stateless between `/pick` and `/act` (no per-pick session map to expire),
and keeps the wire human-readable for debugging.

### `MenuTarget` (the `ref` object)

```jsonc
{
  "kind":  "npc",   // enum, see below — selects which fields are meaningful
  "index": 412,     // server actor index (npc | player); 0 otherwise
  "x":     331,     // ABSOLUTE world X of the target tile / actor
  "y":     552,     // ABSOLUTE world Y (plane folded in)
  "dir":   0,       // boundary edge direction (boundary only); 0 otherwise
  "id":    47,      // def/catalog id: scenery DefID | boundary DefID |
                    //   ground-item ItemID | npc TypeID (info only); 0 otherwise
  "slot":  -1       // inventory slot 0..29 (inventory_item only); -1 otherwise
}
```

`kind` enum (matches `render.Pick` candidate kinds + the inventory pseudo-target):

| `kind`           | identity fields used   | maps to (Layer 1 / world)                  |
|------------------|------------------------|--------------------------------------------|
| `npc`            | `index` (+`x,y,id`)    | `world.NpcRecord.Index`, `TypeID`=`id`     |
| `player`         | `index` (+`x,y`)       | `world.PlayerRecord.Index`                 |
| `self`           | (none; `x,y`=self tile)| `world.Self`                               |
| `ground_item`    | `x,y,id`               | `world.GroundItemRecord{X,Y,ItemID=id}`    |
| `scenery`        | `x,y,id` (+`dir`)      | `facts.At(x,y)` scenery / dyn `world.Scenery` |
| `boundary`       | `x,y,dir,id`           | `facts.At(x,y)` boundary / `world.Boundaries` |
| `terrain`        | `x,y`                  | `render.PickTile` result (walk-here)       |
| `inventory_item` | `slot,id`              | `world.Inventory.Slots()[slot]`            |

Notes:
- `index` is the **stable server actor index** carried on `render.Entity.Index`
  / `world.NpcRecord.Index` / `world.PlayerRecord.Index`. It is the only field
  Layer 2 needs to call `AttackNpc`, `TalkToNpc`, `AttackPlayer`, `NpcCommand`,
  `ExamineNpc`, `ExaminePlayer`. `x,y,id` on actor refs are advisory (UI labels /
  staleness checks); the server re-resolves by `index` at `/act` time.
- `inventory_item` refs do **not** come from `/pick` (the viewport hit-tester
  only sees the 3D world). The browser builds them itself from `/state`
  inventory entries (`slot` + `itemId`). They share the `ref` schema so `/act`
  has one code path. The viewport never returns `kind:"inventory_item"`.
- Numeric zero / `-1` sentinels make the object trivially constructible in JS
  and round-trip-safe. Unknown extra keys are ignored by the decoder
  (`json.Decoder` default), so the client may attach UI-only fields (e.g. a
  cached `label`) without breaking `/act`.

### Go type

```go
// MenuTarget is the self-describing ref that /pick emits and /act consumes.
// It carries enough identity for Layer 2 to resolve the exact runtime.Host
// call without any server-side per-pick session state.
type MenuTarget struct {
    Kind  string `json:"kind"`            // see kind enum
    Index int    `json:"index,omitempty"` // server actor index (npc|player)
    X     int    `json:"x"`               // absolute world X
    Y     int    `json:"y"`               // absolute world Y (plane folded in)
    Dir   int    `json:"dir,omitempty"`   // boundary edge direction
    ID    int    `json:"id,omitempty"`    // scenery/boundary DefID | ItemID | TypeID
    Slot  int    `json:"slot"`            // inventory slot, -1 if N/A (not omitempty)
}
```

`Slot` is **not** `omitempty` (slot 0 is a real slot); it is always emitted, set
to `-1` for non-inventory targets so the JS round-trips it unchanged.

### `MenuOption` (one verb on a target)

```jsonc
{
  "id":        1,            // stable per-target option id, used by /act
  "text":      "Talk-to",    // the authentic RSC menu label to display
  "isDefault": true          // true for exactly one option = the left-click action
}
```

`id` is an index into the target's option list (Layer 2 owns the mapping
`{kind, optionId} → Host method`). It is **only meaningful paired with its
`ref`** — option id 1 means different verbs for an NPC vs. a scenery object.
Exactly one option per candidate has `isDefault: true` (RSC left-click =
top/default action; the overview §4 left-click rule). `Examine` is always
present as a non-default option (it sends no packet — `runtime/examine.go`).

---

## 3. Existing endpoints (documented as-is)

These are already implemented in `cmd/cradle/spectate.go`; the SPA reuses them.
Documented here for completeness — do not change their behaviour.

### `GET /frame`

One freshly-rendered PNG of the host's current view.

Query params (all optional; defaults from `cfg`):

| param  | type | meaning                                    | default            |
|--------|------|--------------------------------------------|--------------------|
| `rot`  | int  | RSC camera yaw 0..255 (masked `&0xff`)     | `cfg.renderRotation` |
| `zoom` | int  | camera distance basis (`dist=zoom*2`)      | `cfg.renderZoom`   |
| `w`    | int  | output width px                            | `cfg.renderW`      |
| `h`    | int  | output height px                           | `cfg.renderH`      |
| `anim` | int  | model-swap animation frame (fire flicker)  | `0`                |
| `t`    | any  | cache-buster (ignored server-side)         | —                  |

Response: `200` `Content-Type: image/png`, `Cache-Control: no-store`. Body is the
PNG. `503` if the host position has not loaded (`x<=0 && y<=0`); `500` on render
error.

The browser must use the same `w,h` it requested when interpreting `/pick`
clicks (the picker rebuilds the camera from `rot,zoom,w,h` — §4).

### `GET /pos`

Host tile + plane for the HUD.

Response `200`:
```json
{ "x": 331, "y": 552, "plane": 0 }
```
`x,y` absolute; `plane` = `pos.Plane()`. `Cache-Control: no-store`.

### `POST /walk` (currently `GET` with query)

Screen click → tile → `host.Walk`. Today implemented as `GET /walk?px&py&rot&
zoom&w&h`. Query params:

| param | type | meaning                                  |
|-------|------|------------------------------------------|
| `px`  | int  | click X in frame pixels (post-letterbox) |
| `py`  | int  | click Y in frame pixels                  |
| `rot`,`zoom`,`w`,`h` | int | the camera the page rendered with |

Behaviour: builds a plane-local `render.View`, `render.PickTile` → nearest tile,
adds the plane offset back, and **retargets** the single-slot `walkCh` (drains
any pending target, then non-blocking sends the new one — RSC walk overrides the
previous, so spamming clicks just retargets). Returns immediately; the worker
calls `host.Walk` asynchronously.

Response `200`:
```json
{ "x": 333, "y": 551 }
```
(absolute tile chosen). `400` missing `px`/`py`; `503` position not loaded;
`204` no tile under the click.

> SPA note: the SPA may keep calling `/walk` for plain left-clicks on terrain,
> OR funnel everything through `/act` with a `terrain` target. Both hit `walkCh`.
> `/walk` is retained unchanged for the legacy page regardless.

### `GET /shot` and `GET /clip`

Agent-facing capture helpers (hotkeys `p` / `c`). Same camera query params as
`/frame`. `/shot` writes `shot_HHMMSS.png` + `shot.png`; `/clip` writes a
12-frame (~2.6s) burst tiled into `clip_sheet_HHMMSS.png` + `clip_sheet.png`,
under `spectatorShotDir`. Responses:

```json
{ "saved": "/tmp/render_out/spectator/shot_142530.png" }
```
```json
{ "saved": "/tmp/render_out/spectator/clip_sheet_142530.png", "frames": 12 }
```

Not used by the SPA UI directly; retained for the controlling agent's review.

---

## 4. `POST /pick` — screen click → ordered menu candidates

Maps a right-click (or default left-click probe) at frame pixel `(px,py)` to a
depth-ordered list of interactable targets, each with its authentic option list
and examine text. **Read-only** — sends no packet, mutates nothing. Calls
`render.Pick` (Layer 1) then Layer 2's menu builder per candidate.

### Request body

```jsonc
{
  "px":   256,   // click X in frame pixels (after undoing object-fit letterbox)
  "py":   170,   // click Y in frame pixels
  "rot":  64,    // camera yaw the frame was rendered with (0..255)
  "zoom": 600,   // camera zoom the frame was rendered with
  "w":    512,   // frame width  px (MUST match the rendered frame)
  "h":    336    // frame height px (MUST match the rendered frame)
}
```

`rot,zoom,w,h` are **required** and must equal what `/frame` was last called
with, because `render.Pick` rebuilds the exact same camera to forward-project
(identical to how `PickTile` works today — `render/pick.go`). A mismatch yields
wrong candidates, not an error. `px,py` are post-letterbox frame pixels (the SPA
undoes `object-fit:contain` scaling exactly as the legacy page does).

### Response `200`

```jsonc
{
  "candidates": [
    {
      "ref": {                       // a MenuTarget (§2) — POST verbatim to /act
        "kind": "npc", "index": 412, "x": 331, "y": 550,
        "dir": 0, "id": 47, "slot": -1
      },
      "kind":    "npc",              // == ref.kind, duplicated for convenient UI switch
      "label":   "Goblin",          // display name (facts def name / player name / item name)
      "examine": "An ugly green monster.", // examine text (runtime/examine.go Description)
      "detail":  "atk=6 str=6 def=6 hp=10, aggressive", // optional Examination.Detail; "" if none
      "dist":    3,                  // chebyshev tile distance self→target (UI ordering aid)
      "options": [
        { "id": 0, "text": "Attack",   "isDefault": true  },
        { "id": 1, "text": "Talk-to",  "isDefault": false },
        { "id": 2, "text": "Examine",  "isDefault": false }
      ]
    },
    {
      "ref": { "kind": "scenery", "index": 0, "x": 330, "y": 551, "dir": 0, "id": 38, "slot": -1 },
      "kind": "scenery", "label": "Tree", "examine": "A tree.", "detail": "",
      "dist": 4,
      "options": [
        { "id": 0, "text": "Chop",    "isDefault": true },
        { "id": 1, "text": "Examine", "isDefault": false }
      ]
    },
    {
      "ref": { "kind": "terrain", "index": 0, "x": 332, "y": 552, "dir": 0, "id": 0, "slot": -1 },
      "kind": "terrain", "label": "Walk here", "examine": "", "detail": "",
      "dist": 0,
      "options": [ { "id": 0, "text": "Walk here", "isDefault": true } ]
    }
  ]
}
```

Ordering: **nearest-camera-first** (painter order reversed — overview §3), so
`candidates[0]` is the top of the right-click menu and its default option is the
left-click action. The always-present `terrain` "Walk here" candidate is **last**
(lowest priority) so it never shadows an actor/object under the cursor but is
always available as a fallback. `candidates` is never null; an empty world click
still returns at least the `terrain` candidate (or `[]` only if no tile
projects, e.g. clicking sky — the SPA treats `[]` as "do nothing").

Field types:

| field                | type    | notes |
|----------------------|---------|-------|
| `candidates`         | array   | depth-ordered, nearest-first; may be empty |
| `candidates[].ref`   | object  | a `MenuTarget` (§2) |
| `candidates[].kind`  | string  | mirror of `ref.kind` (UI convenience) |
| `candidates[].label` | string  | display name; never null (falls back to `"Object"`/`"Item N"`) |
| `candidates[].examine` | string| `Examination.Description`; `""` if unknown |
| `candidates[].detail`  | string| `Examination.Detail`; `""` if none |
| `candidates[].dist`    | int   | chebyshev distance self→`(ref.x,ref.y)` |
| `candidates[].options` | array | ≥1 `MenuOption`; exactly one `isDefault:true` |

Errors: `503` host position not loaded; `400` malformed JSON / missing
`px,py,w,h`; `405` non-POST.

> The SPA renders the menu by listing `options[].text` for `candidates[0]` first,
> then a separator, then deeper candidates (RSC stacks overlapping targets). A
> plain left-click skips the menu and immediately `/act`s `candidates[0]` with
> its `isDefault` option.

---

## 5. `POST /act` — execute one option on one target

Takes a `ref` (from `/pick`, or self-built for inventory) plus the chosen
`optionId`, resolves it through Layer 2's dispatch table to the exact
`runtime.Host` method, and **enqueues** it on the action worker (§7). Returns as
soon as the action is queued/started — it does not block on the full
walk-then-interact round trip (those can take seconds).

### Request body

```jsonc
{
  "ref": {                        // a MenuTarget exactly as /pick returned it
    "kind": "npc", "index": 412, "x": 331, "y": 550,
    "dir": 0, "id": 47, "slot": -1
  },
  "optionId": 1                   // which MenuOption.id on that target (Talk-to)
}
```

### Response `200`

```json
{ "ok": true, "message": "Talk-to Goblin" }
```

| field     | type   | notes |
|-----------|--------|-------|
| `ok`      | bool   | `true` = the action was resolved + enqueued |
| `message` | string | short human-readable echo of what was dispatched (UI toast / log) |

On a resolvable-but-rejected request (e.g. NPC no longer visible, slot empty),
return `200` with `{"ok": false, "message": "<reason>"}` rather than an HTTP
error — the UI surfaces it as a toast and keeps going. Reserve HTTP errors for
malformed input:

- `400` malformed JSON, unknown `ref.kind`, or `optionId` out of range for that
  kind.
- `405` non-POST.

### Dispatch (Layer 2 owns the table; shown here for the contract)

`/act` itself does **no** game logic — it calls one Layer 2 function,
`dispatch.Do(ctx, host, ref, optionId) (message string, err error)`. The table
(detailed in `20-menu-dispatch.md`) maps `{kind, optionId}` → `Host` call. The
verified `Host` signatures the table targets (all `func(ctx, ...) error` unless
noted):

| `ref.kind`       | default option → call                                   | other options |
|------------------|---------------------------------------------------------|---------------|
| `npc`            | `AttackNpc(ctx, ref.Index)` *or* `TalkToNpc` (def order)| `NpcCommand(ctx, ref.Index)`, `ExamineNpc` (no packet) |
| `player`         | `AttackPlayer(ctx, ref.Index)`                          | `Follow`, `PrivateMessage`/`AddFriend` (need name from `world.Players.Get`), `ExaminePlayer` |
| `self`           | (no default action; menu of) `ExamineSelf`              | future: prayers/equipment |
| `ground_item`    | `PickUpItem(ctx, ref.X, ref.Y, ref.ID)`                 | `ExamineGroundItem(ref.X, ref.Y)` |
| `scenery`        | `InteractAt(ctx, ref.X, ref.Y, optionId)`               | `ExamineScenery(ref.X, ref.Y)` |
| `boundary`       | `InteractWithBoundary(ctx, ref.X, ref.Y, ref.Dir)`      | `ExamineBoundary(ref.X, ref.Y)` |
| `terrain`        | `Walk(ctx, ref.X, ref.Y)` (via `walkCh`)                | — |
| `inventory_item` | `ItemCommand(ctx, ref.Slot)` (the def's verb)           | `EquipItem`/`UnequipItem(ctx, ref.Slot)`, `DropItem(ctx, ref.Slot)`, `ExamineInventorySlot(ref.Slot)` |

`scenery`'s `optionId` maps to the `InteractAt` `option` arg (RSC scenery has
Command1=option 0 / Command2=option 1; `SceneryDef.Command1/Command2`). The
`Examine*` calls return an `Examination` (no packet) — `/act` puts its `String()`
into `message` and returns `ok:true` without touching `walkCh`.

> Use-item-on-target (`UseItemOnNpc`, `UseItemOnScenery`, …) is an M2 drag-UX
> (backlog §). When added, `/act` gains an optional `"using": {slot, itemId}`
> field on the request; the dispatch table branches to the `UseItemOn*` family
> when present. The `ref`/`optionId` shape does not change.

---

## 6. `GET /state` — self + inventory + equipment + chat

The SPA's poll for everything that is not the rendered frame: HUD stats, the
inventory grid, the equipment summary, and the chat scrollback. **Read-only**,
fully lock-safe (each piece is a `world` accessor that RLocks + copies). One call
returns one consistent snapshot.

### Response `200`

```jsonc
{
  "self": {
    "x": 331, "y": 552, "plane": 0,
    "heading": 4,
    "combatLevel": 31,
    "hp": 18, "maxHp": 20,
    "prayer": 12, "maxPrayer": 13,
    "fatigue": 4210,            // raw 0..100% scaled units as world.Self.Fatigue reports
    "questPoints": 9,
    "skills": [                 // 18 entries, RSC catalog order (index == skill id)
      { "id": 0,  "name": "attack",      "level": 20, "max": 20, "xp": 4470 },
      { "id": 1,  "name": "defence",     "level": 18, "max": 18, "xp": 3110 },
      { "id": 2,  "name": "strength",    "level": 22, "max": 22, "xp": 6210 },
      { "id": 3,  "name": "hits",        "level": 18, "max": 20, "xp": 4360 }
      /* … through id 17 "thieving"; always all 18, even if 1/1/0 */
    ]
  },
  "inventory": [                // present slots only, in slot order; max 30
    {
      "slot":     0,
      "itemId":   77,
      "name":     "Bronze sword",
      "amount":   1,
      "wielded":  false,
      "wearable": true,         // ItemDef.IsWearable
      "stackable": false,       // ItemDef.IsStackable
      "command":  "Wield"       // ItemDef.Command ("" if none) — the inv default verb
    },
    { "slot": 1, "itemId": 10, "name": "Coins", "amount": 250,
      "wielded": false, "wearable": false, "stackable": true, "command": "" }
  ],
  "equipment": [                // worn items summary, by equip slot; present slots only
    { "slot": 4, "name": "Weapon", "itemId": 0, "sprite": 13 },
    { "slot": 6, "name": "Body",   "itemId": 0, "sprite": 21 }
  ],
  "recentChat": [               // newest-LAST (chronological); bounded (see below)
    { "kind": "chat",   "from": "Zezima", "text": "hi", "at": 1717261530 },
    { "kind": "pm",     "from": "Bob",    "text": "trade?", "at": 1717261535 },
    { "kind": "server", "from": "",       "text": "You can't reach that.", "at": 1717261540 },
    { "kind": "self",   "from": "bernard","text": "selling lobsters", "at": 1717261541 }
  ]
}
```

### Field reference

`self` (from `world.Self`):

| field          | type | source |
|----------------|------|--------|
| `x`,`y`        | int  | `Self.Position()` (absolute) |
| `plane`        | int  | `pos.Plane()` |
| `heading`      | int  | `Self.Heading()` (0..7) |
| `combatLevel`  | int  | `Self.CombatLevel()` |
| `hp`,`maxHp`   | int  | `Self.HP()` / `Self.MaxHP()` (skill id 3) |
| `prayer`,`maxPrayer` | int | `Self.Prayer()` / `Self.MaxPrayer()` (skill id 5) |
| `fatigue`      | int  | `Self.Fatigue()` |
| `questPoints`  | int  | `Self.QuestPoints()` |
| `skills`       | array| 18 entries; `level`=`SkillLevel(id)`, `max`=`SkillMax(id)`, `xp`=`SkillXP(id)`, `name`=`event.SkillName(id)` (lowercase) |

`inventory` (from `world.Inventory.Slots()` + `facts.ItemDef`):

| field      | type   | source |
|------------|--------|--------|
| `slot`     | int    | index in `Slots()` (0-based) |
| `itemId`   | int    | `InvSlot.ItemID` |
| `amount`   | int    | `InvSlot.Amount` |
| `wielded`  | bool   | `InvSlot.Wielded` |
| `name`     | string | `ItemDef.Name` (`"item N"` fallback if def missing) |
| `wearable` | bool   | `ItemDef.IsWearable` |
| `stackable`| bool   | `ItemDef.IsStackable` |
| `command`  | string | `ItemDef.Command` (the inv-item default verb; `""` if none) |

Only **occupied** slots are emitted (skip `ItemID == 0`); the UI renders a 30-cell
grid and places each entry at its `slot`. This keeps the payload small and lets
the grid stay sparse.

`equipment` (from `world.Self.EquipSprites()` + `event.EquipSlotName`): a summary
of worn gear. Today the server-mirrored data is the per-slot **sprite id**
(`Self.EquipSprites()[slot]`, value `>0` = worn), not the catalog item id — so
`itemId` is `0`/best-effort and `name` is the **slot label**
(`event.EquipSlotName(slot)`, e.g. "Weapon", "Body") with `sprite` carrying the
raw sprite id. Only slots with `sprite > 0` are emitted. (M2 may enrich this once
worn-item catalog ids are mirrored; the field set is forward-compatible — add
`itemId` resolution without changing shape.)

| field    | type   | source |
|----------|--------|--------|
| `slot`   | int    | equip slot 0..11 (`event.EquipSlot*`) |
| `name`   | string | `event.EquipSlotName(slot)` |
| `itemId` | int    | `0` for now (no catalog mirror yet) |
| `sprite` | int    | `Self.EquipSprites()[slot]` (>0) |

`recentChat`: a merged, **chronological (newest-last)** view of the host's
transient message buffer (`world.Recent`, `world/recent.go`). RSC has no single
chat log, so `/state` assembles one from the available `RecentEvents` slots:

- `kind: "chat"` from `Recent.Chat()` → `{from: Speaker, text: Message}`.
- `kind: "pm"` from `Recent.PM()` → `{from: Sender, text: Message}`.
- `kind: "server"` AND `kind: "dialog"` from `Recent.ServerMessages()` (the
  bounded ring, `ServerMsgRingCap == 32`) → `{from: "", text: Message}`, using
  the record's own `Kind` field to pick `"server"` vs `"dialog"`.
- `kind: "self"` — messages the SPA sent via `/chat`; see below.

| field  | type   | notes |
|--------|--------|-------|
| `kind` | string | `"chat" \| "pm" \| "server" \| "dialog" \| "self"` |
| `from` | string | speaker/sender; `""` for server messages |
| `text` | string | the message |
| `at`   | int    | unix seconds (`record.At.Unix()`); for client sort/dedupe |

Implementation note on the merge: the single-slot `Chat()` / `PM()` accessors
hold only the **latest** of their kind (see `RecentEvents` doc — it is one slot
per kind, not a list). To give the SPA real scrollback, the remote-client server
keeps its **own** small append-only chat ring (cap ~64) that it grows from two
sources: (1) the bot's **outgoing** `/chat` sends (tagged `"self"`), and (2) a
de-duped tail of `Recent` — on each `/state` it appends any `Chat()`/`PM()`/
newest-`ServerMessages()` entries whose `At` is newer than the last one it
already captured (dedupe by `at`+`kind`+`text`). This ring lives next to
`motion`/`walkCh` in `spectate`'s closure (server-session scoped). The wire shape
above does not change regardless of how deep the buffer is.

`Cache-Control: no-store`. The SPA polls this on a cadence (§8). Errors: `503`
position not loaded (during login); otherwise always `200`.

---

## 7. `POST /chat` — send public / command / private

Sends a chat message through `runtime.Host`. **Mutating** (it writes to the wire)
→ routed through the action worker (§9). Also appended to the server's chat ring
as a `"self"` entry so it shows up in the next `/state`'s `recentChat`.

### Request body

```jsonc
{
  "kind": "say",            // "say" | "command" | "pm"
  "text": "selling lobs",   // the message (for command: WITHOUT a leading "::")
  "to":   ""                // PM recipient name; required iff kind=="pm"
}
```

| field  | type   | required | notes |
|--------|--------|----------|-------|
| `kind` | string | yes      | `"say"` → `Host.Say`; `"command"` → `Host.Command`; `"pm"` → `Host.PrivateMessage` |
| `text` | string | yes      | non-empty; the message body |
| `to`   | string | iff `pm` | recipient player name |

Mapping to verified `Host` methods:

- `say`     → `host.Say(ctx, text)` (runtime/host.go:1005)
- `command` → `host.Command(ctx, text)` (runtime/host.go:1000) — strip a leading
  `::` if the user typed one (RSC commands are entered as `::cmd`; `Command` wants
  the bare command).
- `pm`      → `host.PrivateMessage(ctx, to, text)` (runtime/social.go:12)

### Response `200`

```json
{ "ok": true }
```

`200` `{"ok": false, "message": "..."}` for a rejected-but-valid request (e.g.
`pm` with empty `to`, or send failed). `400` malformed JSON / unknown `kind` /
empty `text`. `405` non-POST.

---

## 8. `GET /examine` — examine one ref (optional convenience)

Returns the full `Examination` for a target without going through `/pick`. Useful
for hover tooltips and inventory long-press, and as the canonical source the SPA
can call when it only has a `ref`. Sends no packet (`runtime/examine.go`).

### Request

`ref` is passed as a **JSON object in a query param** (URL-encoded), since this
is a `GET`:

```
GET /examine?ref=%7B%22kind%22%3A%22npc%22%2C%22index%22%3A412%7D
```

i.e. `ref` = url-encoded `{"kind":"npc","index":412}` (the same `MenuTarget`
shape; only the fields needed to resolve are required). A `POST /examine` with a
`{ "ref": {...} }` body is also accepted for clients that prefer not to URL-encode
JSON; both decode the same `MenuTarget`.

### Response `200` — the `Examination` JSON (`runtime.Examination`)

```jsonc
{
  "kind":        "npc",                       // Examination.Kind
  "name":        "Goblin",                    // Examination.Name
  "description": "An ugly green monster.",    // Examination.Description
  "x":           331,                         // absolute
  "y":           550,
  "detail":      "atk=6 str=6 def=6 hp=10, aggressive" // Examination.Detail
}
```

Field names are the lowercased `Examination` struct fields. Dispatch by
`ref.kind` to the matching `Examine*` method:

| `ref.kind`       | call |
|------------------|------|
| `npc`            | `ExamineNpc(ref.Index)` |
| `player`         | `ExaminePlayer(ref.Index)` |
| `self`           | `ExamineSelf()` |
| `ground_item`    | `ExamineGroundItem(ref.X, ref.Y)` |
| `scenery`        | `ExamineScenery(ref.X, ref.Y)` |
| `boundary`       | `ExamineBoundary(ref.X, ref.Y)` |
| `inventory_item` | `ExamineInventorySlot(ref.Slot)` |
| `terrain`        | (none) → `{"kind":"terrain"}` empty Examination |

`400` malformed/missing `ref`; otherwise `200` (the `Examine*` methods return a
populated-with-`Detail` zero value rather than erroring on a missing target).

---

## 9. Action serialization — the single worker

RSC actions **override each other** (a new walk cancels the prior; an attack
replaces a pending interact). Every mutating endpoint (`/walk`, `/act`, `/chat`)
must therefore funnel through **one** worker goroutine, exactly like the existing
`walkCh` pattern in `spectate.go` (overview §6). Generalise it:

```go
// actReq is one queued world action; the worker runs them one at a time so
// RSC's "latest action wins" semantics hold (a new pick/walk supersedes a
// pending one). Mirrors the existing walkCh retarget pattern.
type actReq struct {
    run  func(ctx context.Context) (string, error) // closes over host + resolved args
    done chan actResult                             // nil for fire-and-forget (e.g. retargeting walk)
}
type actResult struct { msg string; err error }

actCh := make(chan actReq, 1) // depth-1: retarget by drain-then-send (like walkCh)
go func() {
    for {
        select {
        case <-ctx.Done():
            return
        case req := <-actCh:
            wctx, cancel := context.WithTimeout(ctx, 30*time.Second)
            msg, err := req.run(wctx)
            cancel()
            if req.done != nil { req.done <- actResult{msg, err} }
        }
    }
}()
```

- **`/walk` and `terrain`/walk-style `/act`** keep the *fire-and-forget retarget*
  behaviour: drain any pending req, non-blocking send the new one, respond
  immediately (`done == nil`). This preserves today's "spam-click to retarget"
  feel.
- **`/act` (interactions)** and **`/chat`** want a result for the `message` /
  `ok` field. They send with a `done` channel and wait on it **with a short
  timeout** (e.g. select on `done` vs. a ~1.5s timer) so the HTTP response is
  snappy even when the underlying `Walk`-then-interact takes seconds: on timeout
  the handler returns `{"ok": true, "message": "<verb> (running)"}` while the
  worker keeps going. The actual long-running work (pathing + interact) lives in
  the `Host` method and the worker, not the handler.
- The existing `walkCh` may be kept as-is and a separate `actCh` added, OR
  unified into one `actCh`; `50-impl-spec.md` picks one. Either way there is a
  single point of serialization per concern, and walk retargeting semantics are
  preserved. Reads (`/pos`, `/state`, `/pick`, `/examine`, `/frame`) never touch
  the worker — they read lock-safe `world` copies directly from the HTTP
  goroutine.

---

## 10. Polling cadence & cache headers

The browser is the only client; everything is poll-based over plain HTTP (no
websockets — stdlib-only constraint). Recommended cadences (the SPA owns the
timers; these are guidance for `40-browser-ui.md`):

| Endpoint  | Cadence                          | Rationale |
|-----------|----------------------------------|-----------|
| `/frame`  | ~30 fps **gated** (next request only after the prior image `onload`, as the legacy page does) — effective ~15-30 fps | smooth glide without piling up requests |
| `/state`  | every 500 ms                     | inventory/stats/chat change at human/tick rate |
| `/pos`    | every 500 ms (legacy HUD)        | the SPA may drop this in favour of `/state.self` |
| `/pick`   | on right-click / left-click only | user-driven, never polled |
| `/act`    | on menu-option click only        | user-driven |
| `/chat`   | on send only                     | user-driven |
| `/examine`| on hover (debounced ~200 ms)     | optional tooltips |

**Cache headers (server side):**

- All dynamic JSON + the PNG frame: `Cache-Control: no-store` (already set for
  `/pos` and `/frame` today; apply the same to `/state`, `/pick`, `/act`,
  `/chat`, `/examine`). These are point-in-time game state — never cache.
- The SPA HTML (`GET /client`) and any static assets: served inline as a Go
  string const (like `spectatePage`), so caching is moot; send
  `Cache-Control: no-cache` so a reload always re-fetches during development.
- Clients should still append a cache-buster (`&t=<perf.now()>`) to `/frame`
  (the legacy page does) to defeat any intermediary; `/pick`/`/act`/`/chat`/
  `/state` are POST or `no-store` GET and need none.

`Content-Type`:
- `/frame`,`/shot` image: `image/png`.
- everything else: `application/json; charset=utf-8`.

There is no auth and no CORS layer — the server binds `cfg.spectateAddr`
(localhost) and the single SPA is same-origin. Do not add auth in M1 (out of
scope; the overview keeps the surface stdlib-only and local).

---

## 11. Cross-layer field-name contract (summary)

The three identity primitives every layer agrees on, so `/pick` → `/act` →
`Host` is a pure pass-through:

1. **`MenuTarget`** (`{kind,index,x,y,dir,id,slot}`) — emitted by `render.Pick`'s
   candidate (Layer 1 supplies `kind`/coords/ids; the `slot` field is Layer-3/UI
   only). Carried verbatim through `/pick` → SPA → `/act` → Layer-2 dispatch.
2. **`MenuOption`** (`{id,text,isDefault}`) — built by Layer 2 per candidate from
   the `facts` def commands + `Examine`. `id` is dispatch-table-relative to
   `kind`.
3. **`Examination`** (`{kind,name,description,x,y,detail}`) — `runtime.Examination`
   verbatim, surfaced by `/pick` (as `label`/`examine`/`detail`) and `/examine`.

`render.Pick` (Layer 1) must therefore return candidates carrying at minimum:
`Kind string`, `Index int`, `X,Y int` (absolute), `Dir int`, `ID int` — i.e. the
non-`slot` subset of `MenuTarget` — plus the `Examination` for `label`/`examine`.
`10-hit-testing.md` defines that Go return type; this doc fixes its JSON
projection. Layer 2 (`20-menu-dispatch.md`) defines the per-kind `{optionId →
Host call}` table; this doc fixes its JSON projection (`options[]`) and the
`/act` request that feeds it.
