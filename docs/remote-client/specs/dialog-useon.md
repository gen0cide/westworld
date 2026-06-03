# Spec F3 + F4 — NPC dialog-option menu + use-item-on-target

Two related UX features for the cradle remote client, cloning the BANK template
(`/state` read block + dedicated POST endpoint + serialized `enqueueAction`
worker + web window/component). This is a SPEC: do not write source here, only
implement against it.

- **F3 — NPC dialog-option menu.** When the server presents an NPC conversation
  menu (a list of dialog choices), surface them as clickable options. Pick one
  → server reply.
- **F4 — Use-item-on-target.** Drag an inventory item onto a target (another
  inventory item, scenery, boundary, ground item, npc, player) → fire the
  matching `UseItemOn*` packet.

Both are read-mostly + one POST endpoint each, identical in shape to `/bank`.

---

## 0. Backend facts (cited — already implemented, do NOT modify)

### F3 data flow (dialog menu arrives → host picks an option)

- Server packet `event.NpcDialog{Options []string}` is decoded and ingested in
  `world/world.go:835-840`:
  ```go
  case event.NpcDialog:
      w.Recent.SetDialogOptions(e.Options)
  ```
  Event type: `event/events.go:326-332` (`NpcDialog{NpcIndex int; Options []string}`).
- The live menu is held in `world.RecentEvents`:
  - `world/recent.go:97-100` — `type DialogOptionsRecord struct { Options []string; At time.Time }`
  - `world/recent.go:226-235` — `func (r *RecentEvents) DialogOptions() *DialogOptionsRecord` (returns nil when no menu is open; returns a deep copy of `Options`)
  - `world/recent.go:251-254` — `func (r *RecentEvents) ClearDialogOptions()`
  - `world/recent.go:212-220` — `func (r *RecentEvents) DialogText() *DialogTextRecord` (the NPC speech-bubble line that usually precedes the menu; `DialogTextRecord{Text string; At time.Time}` at `world/recent.go:88-91`).
- `world.World.Recent` is the exported field (`world/world.go:19`), reached from the handler via `host.World().Recent`.
- Host method to PICK an option (the only call F3 needs):
  - `runtime/combat.go:218-224`
    ```go
    func (h *Host) ChooseDialogOption(ctx context.Context, index int) error
    ```
    `index` is 0-based into the same `Options` order the bot received. It sends
    the `CHOOSE_OPTION` packet; it does NOT clear the menu — the handler clears
    it via `host.World().Recent.ClearDialogOptions()` after a successful send
    (matching `runtime/actions_bank.go:122-127`, where bank.open does
    `ChooseDialogOption` then `ClearDialogOptions`).

> There is **no** `host.ChooseDialogOption`-style guard accessor; the menu's
> openness is read directly from `host.World().Recent.DialogOptions() != nil`,
> exactly as `runtime/actions_bank.go:116` reads it.

### F4 data flow (drag inventory item onto target → fire UseItemOn\*)

All six already exist on `*runtime.Host` AND are already on the
`remoteclient.ActionHost` interface (`remoteclient/dispatch.go:45-50`):

| Target kind        | Host method (file:line)                                                                 | signature                                            |
|--------------------|------------------------------------------------------------------------------------------|------------------------------------------------------|
| inventory item     | `runtime/boundary.go:48` `UseItemOnItem`                                                 | `(ctx, slot1, slot2 int) error`                      |
| scenery            | `runtime/boundary.go:56` `UseItemOnScenery`                                              | `(ctx, x, y, slot int) error`                        |
| boundary (wall)    | `runtime/boundary.go:34` `UseItemOnBoundary`                                             | `(ctx, x, y, direction, slot int) error`             |
| ground item        | `runtime/boundary.go:84` `UseItemOnGroundItem`                                           | `(ctx, x, y, groundItemID, slot int) error`          |
| npc                | `runtime/boundary.go:99` `UseItemOnNpc`                                                  | `(ctx, npcServerIndex, slot int) error`              |
| player             | `runtime/boundary.go:130` `UseItemOnPlayer`                                              | `(ctx, playerServerIndex, slot int) error`           |

- `slot`/`slot1`/`slot2` are 0-based inventory slot indices (no conversion).
- `x,y` are ABSOLUTE world coords (plane folded into Y) — this is exactly what
  `MenuTarget.X/Y` already carry on the wire (`remoteclient/target.go:55-58,
  71-80`), so a `MenuTarget` from `/pick` round-trips into these methods unchanged.
- `UseItemOnScenery/Boundary/GroundItem/Npc/Player` pathfind-then-fire
  (`walkAndAct`); `UseItemOnItem` fires immediately. All are blocking and belong
  on the serialized action worker, like every other interaction.

> The existing M1 inventory `OptUse` is a no-op placeholder
> (`remoteclient/dispatch.go:458-460`, returns `"Use-on is M2"`, `LaneSync`).
> F4 does NOT change `dispatch.go` or the `/act` lane machinery — it adds a
> dedicated `/useon` endpoint, because use-on needs a SECOND (dragged) target
> that the single-`MenuTarget` `/act` contract cannot express. Touching
> `remoteclient/` is allowed (it is not `reference/`, `world/`, or `runtime/`),
> but is unnecessary; keep the change inside `cmd/cradle/` + `web/`.

`render.Pick` already hit-tests npc/player/self/ground-item/scenery/boundary, so
the browser can obtain a drop-target `MenuTarget` via the existing `/pick`
endpoint (`cmd/cradle/remoteclient.go:683-769`). F4 reuses `/pick` to resolve
the drop coordinates into a typed target; it adds no new picking code.

---

## 1. F3 — NPC dialog menu

### 1.1 Go: `/state` additions (`cmd/cradle/remoteclient.go`)

Add a new response sub-struct and an optional pointer field on `stateResponse`,
exactly mirroring the `Bank *stateBank` pattern (window present only while open).

In the wire-types block (next to `stateBank`, ~`cmd/cradle/remoteclient.go:307-319`):
```go
// stateDialog mirrors world.RecentEvents.DialogOptions() for the SPA. Present
// only while the server has an NPC option menu open. NpcText is the most recent
// speech-bubble line (may be empty); Options are the choices in server order
// (option index == array index == the arg to host.ChooseDialogOption).
type stateDialog struct {
    Open    bool     `json:"open"`
    NpcText string   `json:"npcText"`
    Options []string `json:"options"`
}
```

On `stateResponse` (after the `Bank *stateBank` field, `cmd/cradle/remoteclient.go:304`):
```go
    // Dialog is present only while an NPC option menu is open.
    Dialog *stateDialog `json:"dialog,omitempty"`
```

In the `/state` handler, after the bank block (`cmd/cradle/remoteclient.go:962-981`)
and before building `resp` (`:983`):
```go
// dialog block: present only while an NPC option menu is open.
var dialogBlock *stateDialog
if rec := host.World().Recent.DialogOptions(); rec != nil && len(rec.Options) > 0 {
    npcText := ""
    if dt := host.World().Recent.DialogText(); dt != nil {
        npcText = dt.Text
    }
    opts := make([]string, len(rec.Options))
    copy(opts, rec.Options)
    dialogBlock = &stateDialog{Open: true, NpcText: npcText, Options: opts}
}
```
Then add `Dialog: dialogBlock,` to the `stateResponse{...}` literal
(`cmd/cradle/remoteclient.go:983-989`).

### 1.2 Go: `POST /dialog` handler (`cmd/cradle/remoteclient.go`)

Add a request type next to `bankRequest` (`cmd/cradle/remoteclient.go:323-327`):
```go
// dialogRequest is the body of POST /dialog: choose option Index (0-based, into
// the options[] last seen in /state) of the open NPC menu.
type dialogRequest struct {
    Index int `json:"index"`
}
```

Register the handler immediately after the `/bank` handler block
(`cmd/cradle/remoteclient.go:1041`), cloning the bank guard + `enqueueAction` shape:
```go
// POST /dialog — choose one NPC dialog option (spec F3). Serialized through the
// same action worker as /act and /bank.
mux.HandleFunc("/dialog", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req dialogRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")

    // Validate against the live menu BEFORE enqueuing so a stale click (menu
    // already gone, or index out of range) returns {ok:false} without a packet.
    rec := host.World().Recent.DialogOptions()
    if rec == nil || len(rec.Options) == 0 {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "no dialog menu open"})
        return
    }
    if req.Index < 0 || req.Index >= len(rec.Options) {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false,
            Message: fmt.Sprintf("option %d out of range (%d options)", req.Index, len(rec.Options))})
        return
    }
    idx := req.Index
    label := rec.Options[idx]
    msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
        if err := host.ChooseDialogOption(wctx, idx); err != nil {
            return "", err
        }
        // The server does not always tell us the menu closed; clear it so the
        // next /state poll hides the window until a new menu arrives (mirrors
        // runtime/actions_bank.go:122-127).
        host.World().Recent.ClearDialogOptions()
        return "Chose: " + label, nil
    })
    if runErr != nil {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
        return
    }
    _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
})
```

### 1.3 Web types (`web/src/types.ts`)
```ts
export interface Dialog {
  open: boolean
  npcText: string
  options: string[]
}
```
Add to `GameState` (after `bank?: Bank`, `web/src/types.ts:110`):
```ts
  dialog?: Dialog // present only while an NPC option menu is open
```

### 1.4 Web api (`web/src/api.ts`)
After `bankAction` (`web/src/api.ts:93`):
```ts
/** POST /dialog — pick option `index` (0-based) of the open NPC menu. */
export async function chooseDialog(index: number): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/dialog', { index })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

### 1.5 Web component `web/src/components/DialogMenu.tsx` (NEW)
Props: `{ dialog: Dialog }`. The classic RSC dialog menu is a left-click list of
choices (no right-click menu). Left-click → `chooseDialog(index)`; flash the
result. The `/state` poll (400 ms) clears `state.dialog` after the server reply,
which unmounts the component.
```tsx
import { chooseDialog } from '../api'
import { useUI } from '../ui'
import type { Dialog } from '../types'

export function DialogMenu({ dialog }: { dialog: Dialog }) {
  const ui = useUI()
  const pick = (i: number) => {
    void chooseDialog(i).then((r) => { if (r.message) ui.flash(r.message) })
  }
  return (
    <div className="dialogwin">
      {dialog.npcText && <div className="dialogtext">{dialog.npcText}</div>}
      <ul className="dialogopts">
        {dialog.options.map((opt, i) => (
          <li key={i} className="dialogopt" onClick={() => pick(i)}>{opt}</li>
        ))}
      </ul>
    </div>
  )
}
```

### 1.6 App wiring (`web/src/App.tsx`)
Import `DialogMenu`, and render it overlaid on the viewport (NOT a modal backdrop
— the authentic dialog box sits at the bottom of the game view and does not block
the world). Add after the bank block (`web/src/App.tsx:97-99`):
```tsx
{state?.dialog?.open && <DialogMenu dialog={state.dialog} />}
```

### 1.7 Styles (`web/src/styles.css`)
Append (RSC dialog box is a translucent panel anchored over the lower viewport;
options are yellow, highlight to white on hover — `docs/remote-client/100-rsc-ui-map.md` §3 palette via the existing CSS vars):
```css
/* ---- NPC dialog option menu (F3) ---- */
.dialogwin {
  position: absolute;
  left: 50%;
  bottom: 12px;
  transform: translateX(-50%);
  min-width: 320px;
  max-width: 80%;
  background: rgba(10, 20, 10, .92);
  border: 2px solid var(--panel-border);
  box-shadow: 0 0 0 2px #000;
  padding: 6px 10px;
  z-index: 45;            /* above viewport, below modal-backdrop (40) windows? */
  font-size: 14px;
}
.dialogtext { color: var(--rsc-cyan); padding: 2px 0 6px; white-space: pre-wrap; }
.dialogopts { list-style: none; margin: 0; padding: 0; }
.dialogopt { color: var(--rsc-yellow); padding: 2px 4px; cursor: pointer; }
.dialogopt:hover { color: var(--rsc-white); background: #143014; }
```
(`z-index: 45` puts the dialog above the `45`-free range; bank/shop modals use
`modal-backdrop` z-index `40` with their content centered — the dialog never
co-exists with the bank window, so ordering is moot, but 45 keeps it above the
plain viewport HUD.)

---

## 2. F4 — Use-item-on-target (drag UX)

### 2.1 Interaction model (web only — no new backend picking)

1. **Drag start** — in `InventoryGrid` (`web/src/components/InventoryGrid.tsx`),
   make each occupied cell `draggable`. On `onDragStart`, stash the source
   `InvItem` (slot + itemId + name) in a small shared "drag" context (new tiny
   provider, see §2.5). Set `e.dataTransfer.effectAllowed = 'move'`.
2. **Drop on another inventory cell** — `InventoryGrid` cells also accept the
   drop: `onDragOver` → `e.preventDefault()`; `onDrop` → if the drag source slot
   != this slot, call `useOn({kind:'item', slot1: src.slot, slot2: dst.slot})`.
   This is the only target resolvable without a screen pick.
3. **Drop on the world (viewport)** — `Viewport` (`web/src/components/Viewport.tsx`)
   accepts the drop on its `<img>`. On `onDrop`, compute the click pixel
   (`e.clientX/Y` minus the image's bounding rect, scaled to the render `w/h` —
   the SAME math the existing left-click-to-walk / right-click-to-pick path uses
   in `Viewport`), then call the existing `pick(px, py, cam)` to get candidates,
   take `candidates[0]` (nearest, depth-ordered — `remoteclient/target.go:109-114`),
   and route its `ref.kind` to `useOn` (§2.3). If `candidates` is empty or only
   terrain, flash `"nothing to use that on"` and abort (RSC drops the item-use on
   bare ground silently; we no-op).

> Rationale for reusing `/pick`: the drop target's typed `MenuTarget`
> (kind + index + absolute x,y + dir + id) is exactly what `/pick` already emits,
> and it carries everything the `/useon` handler needs. No new hit-testing.

### 2.2 Go: `POST /useon` handler (`cmd/cradle/remoteclient.go`)

Request type next to `bankRequest`/`dialogRequest`:
```go
// useOnRequest is the body of POST /useon: use inventory slot Slot on a target.
// Target is the SAME MenuTarget /pick emits; the handler routes on Target.Kind to
// the matching host.UseItemOn* method. For Kind=="inventory_item" the target item
// is identified by Target.Slot (the second slot); for everything else the world
// coords / index in Target locate the drop target.
type useOnRequest struct {
    Slot   int                     `json:"slot"`   // dragged source inventory slot (0-based)
    Target remoteclient.MenuTarget `json:"target"` // drop target (from /pick, or hand-built for item-on-item)
}
```

Register after `/dialog` (`cmd/cradle/remoteclient.go`):
```go
// POST /useon — use a dragged inventory slot on a target (spec F4). Routes on
// Target.Kind to the matching host.UseItemOn* method via the serialized worker.
mux.HandleFunc("/useon", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req useOnRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")

    // Re-validate the SOURCE slot against the live inventory (it may have shifted
    // since the drag started; RemoveSlot compacts holes — same hazard
    // remoteclient.requireSlotItem guards, dispatch.go:489-501).
    slots := host.World().Inventory.Slots()
    if req.Slot < 0 || req.Slot >= len(slots) || slots[req.Slot].ItemID == 0 {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "source slot empty"})
        return
    }
    t := req.Target
    src := req.Slot

    // Re-validate volatile target identity for the kinds that need it (npc/player
    // visibility, target inventory slot occupancy), mirroring the dispatcher's
    // staleness checks; reject before touching the wire.
    switch t.Kind {
    case remoteclient.KindNPC:
        if _, ok := host.World().Npcs.Get(t.Index); !ok {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "npc no longer visible"})
            return
        }
    case remoteclient.KindPlayer:
        if _, ok := host.World().Players.Get(t.Index); !ok {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "player no longer visible"})
            return
        }
    case remoteclient.KindInventoryItem:
        if t.Slot < 0 || t.Slot >= len(slots) || slots[t.Slot].ItemID == 0 || t.Slot == src {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "invalid target item slot"})
            return
        }
    case remoteclient.KindScenery, remoteclient.KindBoundary, remoteclient.KindGroundItem:
        // coords carried in Target; nothing volatile to re-validate here.
    default:
        _ = json.NewEncoder(w).Encode(actResponse{OK: false,
            Message: fmt.Sprintf("cannot use an item on %q", t.Kind)})
        return
    }

    msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
        switch t.Kind {
        case remoteclient.KindInventoryItem:
            return "Use item on item", host.UseItemOnItem(wctx, src, t.Slot)
        case remoteclient.KindScenery:
            return "Use item on scenery", host.UseItemOnScenery(wctx, t.X, t.Y, src)
        case remoteclient.KindBoundary:
            return "Use item on boundary", host.UseItemOnBoundary(wctx, t.X, t.Y, t.Dir, src)
        case remoteclient.KindGroundItem:
            return "Use item on ground item", host.UseItemOnGroundItem(wctx, t.X, t.Y, t.ID, src)
        case remoteclient.KindNPC:
            return "Use item on npc", host.UseItemOnNpc(wctx, t.Index, src)
        case remoteclient.KindPlayer:
            return "Use item on player", host.UseItemOnPlayer(wctx, t.Index, src)
        }
        return "", fmt.Errorf("unhandled use-on kind %q", t.Kind)
    })
    if runErr != nil {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
        return
    }
    _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
})
```

> `host.World().Npcs.Get` / `Players.Get` are the SAME accessors the existing
> `worldViewAdapter` uses (`cmd/cradle/remoteclient.go:231-239`). `Inventory.Slots()`
> is used at `cmd/cradle/remoteclient.go:707, 906`. No new Host method is required.

### 2.3 Web api (`web/src/api.ts`)
After `chooseDialog`:
```ts
import type { MenuTarget } from './types' // already imported

/** POST /useon — use dragged inventory `slot` on `target` (a MenuTarget). */
export async function useOn(slot: number, target: MenuTarget): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/useon', { slot, target })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```
(`MenuTarget` is already exported from `web/src/types.ts:10-19`.)

### 2.4 Web helper for item-on-item target
Build the inventory-item `MenuTarget` the same way `InventoryGrid.slotRef` does
(`web/src/components/InventoryGrid.tsx:12-14`):
```ts
const itemTarget = (dst: InvItem): MenuTarget =>
  ({ kind: 'inventory_item', index: 0, x: 0, y: 0, dir: 0, id: dst.itemId, slot: dst.slot })
```

### 2.5 Web drag context `web/src/drag.tsx` (NEW, tiny)
A minimal provider holding the in-flight drag source so `Viewport` (a different
subtree from the side panel) can read it on drop. Keeps `useUI` untouched.
```tsx
import { createContext, useContext, useRef, type ReactNode } from 'react'
import type { InvItem } from './types'

interface DragCtx {
  begin: (it: InvItem) => void
  take: () => InvItem | null   // returns + clears the current source
}
const Ctx = createContext<DragCtx | null>(null)
export function useDrag(): DragCtx {
  const c = useContext(Ctx)
  if (!c) throw new Error('useDrag must be used within <DragProvider>')
  return c
}
export function DragProvider({ children }: { children: ReactNode }) {
  const src = useRef<InvItem | null>(null)
  const api: DragCtx = {
    begin: (it) => { src.current = it },
    take: () => { const s = src.current; src.current = null; return s },
  }
  return <Ctx.Provider value={api}>{children}</Ctx.Provider>
}
```
Wrap `<App/>` content with `<DragProvider>` where `<UIProvider>` is mounted
(inspect `web/src/main.tsx`; nest `DragProvider` inside `UIProvider`).

### 2.6 InventoryGrid changes (`web/src/components/InventoryGrid.tsx`)
- `import { useDrag } from '../drag'` and `import { useOn } from '../api'`.
- Each occupied cell gains:
  ```tsx
  draggable
  onDragStart={() => drag.begin(it)}
  onDragOver={(e) => e.preventDefault()}
  onDrop={(e) => {
    e.preventDefault()
    const src = drag.take()
    if (src && src.slot !== it.slot) {
      void useOn(src.slot, itemTarget(it)).then((r) => { if (r.message) ui.flash(r.message) })
    }
  }}
  ```
- `<ItemSprite>` already sets `draggable={false}` on the `<img>`
  (`web/src/components/ItemSprite.tsx:19`), so the DRAG fires from the cell, not
  the icon — leave ItemSprite as-is.

### 2.7 Viewport drop target (`web/src/components/Viewport.tsx`)
- `import { useDrag } from '../drag'`, `import { useOn, pick } from '../api'`.
- On the viewport `<img>` (or its wrapper that already owns the click handlers):
  ```tsx
  onDragOver={(e) => e.preventDefault()}
  onDrop={async (e) => {
    e.preventDefault()
    const src = drag.take()
    if (!src) return
    const { px, py } = pixelFromEvent(e)   // REUSE the existing click→pixel math in Viewport
    const res = await pick(px, py, camera)
    const top = res.candidates.find((c) => c.ref.kind !== 'terrain')
    if (!top) { ui.flash('nothing to use that on'); return }
    const r = await useOn(src.slot, top.ref)
    if (r.message) ui.flash(r.message)
  }}
  ```
  `pixelFromEvent` is whatever Viewport already uses to convert a mouse event
  into render-space `px,py` for `/pick` / `/walk`; reuse it verbatim (do not
  re-derive). `camera` is the prop Viewport already receives.

### 2.8 Styles (`web/src/styles.css`)
Optional drop-highlight (the inventory `.cell` already has a hover cursor):
```css
.cell.dropok { outline: 1px dashed var(--rsc-yellow); }
#view.dropok { outline: 2px dashed var(--rsc-yellow); outline-offset: -2px; }
```
(Toggling these classes on `onDragEnter`/`onDragLeave` is a nicety, not required
for function.)

---

## 3. Files touched (allowed dirs only)

- `cmd/cradle/remoteclient.go` — add `stateDialog`, `dialogRequest`,
  `useOnRequest` types; `Dialog` field on `stateResponse` + its `/state` block;
  `/dialog` and `/useon` handlers. (No new `cmd/cradle/<feature>.go` file is
  needed; everything lives in `serveClient`'s mux beside `/bank`. A
  `cmd/cradle/dialoguseon.go` register-func split is acceptable but only if the
  mux + closures are passed in — `enqueueAction`, `host`, `f` are closure-locals,
  so inline is simpler and matches the bank precedent.)
- `web/src/types.ts` — `Dialog` interface + `GameState.dialog?`.
- `web/src/api.ts` — `chooseDialog`, `useOn`.
- `web/src/components/DialogMenu.tsx` — NEW (F3 component).
- `web/src/drag.tsx` — NEW (F4 drag context).
- `web/src/components/InventoryGrid.tsx` — drag source + inventory-cell drop.
- `web/src/components/Viewport.tsx` — world drop target (reuses existing pixel math).
- `web/src/App.tsx` — render `<DialogMenu>`.
- `web/src/main.tsx` — wrap with `<DragProvider>`.
- `web/src/styles.css` — `.dialogwin`/`.dialogtext`/`.dialogopts`/`.dialogopt`
  + `.dropok` rules.

NOT touched: `reference/`, `world/`, `runtime/`, `remoteclient/`.

Build: `npm --prefix web run build` then `go build ./...` (do not restart the
running cradle on :8090).

---

## 4. Test plan

### F3 (dialog menu)
1. **Backend unit-ish (curl).** With a session where an NPC menu is open
   (e.g. talk to a banker / quest NPC so `event.NpcDialog` arrives), `GET /state`
   → assert `dialog.open == true`, `dialog.options.length > 0`, `dialog.npcText`
   matches the speech bubble.
2. `POST /dialog {"index":0}` → `{ok:true, message:"Chose: <opt0>"}`. Then the
   next `GET /state` shows `dialog` absent (handler cleared the menu) until the
   server sends the follow-up menu.
3. **Out-of-range / stale.** `POST /dialog {"index":99}` → `{ok:false,
   "...out of range..."}`. `POST /dialog` with no menu open → `{ok:false,
   "no dialog menu open"}`. No packet is sent in either case.
4. **UI.** Trigger an NPC conversation in the SPA; the `DialogMenu` appears at
   the bottom of the viewport with the speech line + clickable options. Click an
   option → flash shows "Chose: …" and the box disappears (or swaps to the next
   menu) within one `/state` poll (~400 ms).

### F4 (use-on)
5. **Item-on-item (no pick).** Drag a usable item (e.g. tinderbox) onto logs in
   the inventory grid → `POST /useon {slot:<src>, target:{kind:"inventory_item",
   slot:<dst>,...}}` fires `host.UseItemOnItem(src,dst)`; flash "Use item on item".
6. **Item-on-world.** Drag an item onto an NPC / scenery / ground item / door in
   the viewport → SPA calls `/pick` at the drop pixel, takes the top non-terrain
   candidate, then `/useon`; assert the right `UseItemOn{Npc,Scenery,GroundItem,
   Boundary}` ran (server log line from `runtime/boundary.go:35/57/85/100`).
7. **Drop on bare ground.** Drag onto an empty tile → `/pick` returns only
   terrain → flash "nothing to use that on", no `/useon`, no packet.
8. **Stale source.** Start a drag, let inventory shift (item consumed) before
   drop → `/useon` returns `{ok:false,"source slot empty"}`.
9. **Stale npc/player target.** Drag onto an NPC that walks out of view before
   drop resolves → `{ok:false,"npc no longer visible"}`.
10. **Serialization.** Rapidly drop two use-ons; confirm they run one-at-a-time
    on the action worker (no two packets in flight), same as `/act`/`/bank`.
