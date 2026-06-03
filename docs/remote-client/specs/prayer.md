# Prayer Tab — Implementation Spec

**Task tree ref:** D3 in `docs/remote-client/110-react-port.md`
**Scope:** expose the existing prayer state + activate/deactivate actions over
HTTP, add a Prayer tab to SidePanel, wire types and api, add styles.

---

## 1. Backend — existing Host methods (read-only, do NOT modify)

| Method | File:line | Notes |
|---|---|---|
| `host.ActivatePrayer(ctx, prayerID int) error` | `runtime/prayer.go:12` | wraps `action.ActivatePrayer`; sends opcode 60 |
| `host.DeactivatePrayer(ctx, prayerID int) error` | `runtime/prayer.go:17` | sends opcode 254 |
| `host.World().Self.ActivePrayers() []bool` | `world/self.go:300` | copy of the active-prayer bitmap; `[i]` true = slot i is on |
| `host.World().Self.PrayerActive(idx int) bool` | `world/self.go:290` | single-slot check |
| `host.World().Self.Prayer() int` | `world/self.go:374` | current prayer points = `SkillLevel(5)` |
| `host.World().Self.MaxPrayer() int` | `world/self.go:375` | base prayer level = `SkillMax(5)` |
| `facts.Prayers` (slice) | `facts/prayers.go:14` | static catalog; 14 entries, ID 0–13 |
| `facts.PrayerDef` struct fields | `facts/prayers.go:4` | `ID int`, `Name string`, `ReqLevel int`, `DrainRate int`, `Description string` |

---

## 2. Backend additions — `cmd/cradle/remoteclient.go`

### 2a. New wire types (add after `stateBankSlot`, before `bankRequest`)

```go
// statePrayer is one entry in the /state prayer list.
// Active reflects the live server-confirmed bitmap (opcode 206).
// ReqLevel, DrainRate, and Description come from the static facts catalog.
type statePrayer struct {
    ID          int    `json:"id"`
    Name        string `json:"name"`
    ReqLevel    int    `json:"reqLevel"`
    DrainRate   int    `json:"drainRate"`
    Description string `json:"description"`
    Active      bool   `json:"active"`
}

// prayerRequest is the body of POST /prayer.
// On is true to activate, false to deactivate.
type prayerRequest struct {
    ID int  `json:"id"`
    On bool `json:"on"`
}
```

### 2b. Extend `stateResponse` (add `Prayers` field)

Current (line 304):
```go
Bank *stateBank `json:"bank,omitempty"`
```
Add immediately after:
```go
Prayers []statePrayer `json:"prayers"`
```

`Prayers` is **always present** (not omitempty) — it is the full 14-entry catalog
with live `active` flags. The SPA reads it on every poll to reflect drain-based
deactivations the server sends.

### 2c. Populate prayers in the `/state` handler

Add this block in the `GET /state` handler, after the `bankBlock` block and
before the `resp := stateResponse{...}` literal:

```go
// prayer list: full static catalog + live active flags from the world mirror.
activePrayers := host.World().Self.ActivePrayers()
prayerList := make([]statePrayer, len(facts.Prayers))
for i, def := range facts.Prayers {
    active := false
    if i < len(activePrayers) {
        active = activePrayers[i]
    }
    prayerList[i] = statePrayer{
        ID:          def.ID,
        Name:        def.Name,
        ReqLevel:    def.ReqLevel,
        DrainRate:   def.DrainRate,
        Description: def.Description,
        Active:      active,
    }
}
```

Then update the `resp` literal to include:
```go
Prayers:   prayerList,
```

### 2d. New POST /prayer handler

Add after the `/bank` handler (around line 1041), following the same
`enqueueAction` pattern:

```go
// POST /prayer — activate or deactivate one prayer slot (D3).
// Body: {id: 0..13, on: bool}
// Funneled through the serialized action worker; guards: id range check,
// prayer points check (on:true only). Server may silently reject (level
// below req or 0 prayer points) — caller should re-read /state.
mux.HandleFunc("/prayer", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req prayerRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    if req.ID < 0 || req.ID > 13 {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "id must be 0..13"})
        return
    }
    // Guard: activating when prayer points are 0 is a no-op on the server.
    // Warn but still send — the user may want to queue for when they sip a potion.
    if req.On && host.World().Self.Prayer() == 0 {
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "prayer points are 0"})
        return
    }
    id, on := req.ID, req.On
    msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
        if on {
            verb := "Activate"
            return fmt.Sprintf("%s prayer %d", verb, id), host.ActivatePrayer(wctx, id)
        }
        return fmt.Sprintf("Deactivate prayer %d", id), host.DeactivatePrayer(wctx, id)
    })
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")
    if runErr != nil {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
        return
    }
    _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
})
```

---

## 3. Web types — `web/src/types.ts`

Add after the `Bank`/`BankSlot` block, before `GameState`:

```ts
export interface Prayer {
  id: number
  name: string
  reqLevel: number
  drainRate: number
  description: string
  active: boolean
}
```

Extend `GameState`:
```ts
prayers: Prayer[]  // always present; 14 entries, active flag live
```

---

## 4. Web API — `web/src/api.ts`

Add after `bankAction`:

```ts
/** POST /prayer — activate (on:true) or deactivate (on:false) one prayer slot. */
export async function prayerAction(id: number, on: boolean): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/prayer', { id, on })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

Add `Prayer` to the import in `api.ts` if it ends up needed there; types are
imported in the component only.

---

## 5. Component — `web/src/components/PrayerTab.tsx` (new file)

Props: `{ prayers: Prayer[], prayerPts: number, maxPrayer: number }`

Behaviour:
- **Left-click** a prayer row: toggle. If `active`, call `prayerAction(id, false)`;
  if `!active`, call `prayerAction(id, true)`. Flash the actResponse message on
  completion.
- **Right-click** a prayer row: open a `ui.openActions` context menu with two
  items: "Activate" and "Deactivate" (same calls as left-click). This mirrors the
  bank/inventory right-click pattern.
- **Greyed-out / disabled appearance** for prayers where
  `maxPrayer < prayer.reqLevel` (player's base prayer level is too low). A
  disabled prayer can still be clicked; the server rejects it silently.

```tsx
// web/src/components/PrayerTab.tsx
import { prayerAction } from '../api'
import { useUI, type MenuSection } from '../ui'
import type { Prayer } from '../types'

export function PrayerTab({
  prayers,
  prayerPts,
  maxPrayer,
}: {
  prayers: Prayer[]
  prayerPts: number
  maxPrayer: number
}) {
  const ui = useUI()

  const toggle = (p: Prayer) => {
    void prayerAction(p.id, !p.active).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const contextMenu = (e: React.MouseEvent, p: Prayer) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: p.name,
      items: [
        { text: 'Activate',   run: () => { void prayerAction(p.id, true).then((r)  => { if (r.message) ui.flash(r.message) }) } },
        { text: 'Deactivate', run: () => { void prayerAction(p.id, false).then((r) => { if (r.message) ui.flash(r.message) }) } },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  return (
    <div className="prayer-tab">
      <div className="prayer-pts">
        Prayer: {prayerPts}/{maxPrayer}
      </div>
      <ul className="prayer-list">
        {prayers.map((p) => {
          const tooLow = maxPrayer < p.reqLevel
          return (
            <li
              key={p.id}
              className={'prayer-row' + (p.active ? ' active' : '') + (tooLow ? ' disabled' : '')}
              title={`${p.description} (drain: ${p.drainRate} pts/min, req: ${p.reqLevel})`}
              onClick={() => toggle(p)}
              onContextMenu={(e) => contextMenu(e, p)}
            >
              <span className="prayer-indicator">{p.active ? '●' : '○'}</span>
              <span className="prayer-name">{p.name}</span>
              <span className="prayer-level">Lv{p.reqLevel}</span>
            </li>
          )
        })}
      </ul>
    </div>
  )
}
```

---

## 6. SidePanel wiring — `web/src/components/SidePanel.tsx`

### Type union
```ts
type Tab = 'inv' | 'stats' | 'prayer'
```

### New import
```ts
import { PrayerTab } from './PrayerTab'
```

### Tab bar button (add after the Stats button)
```tsx
<button className={tab === 'prayer' ? 'active' : ''} onClick={() => setTab('prayer')}>Prayer</button>
```

### Tab body case (add inside the `tabbody` conditional)
```tsx
{tab === 'prayer' && (
  <PrayerTab
    prayers={state?.prayers ?? []}
    prayerPts={state?.self?.prayer ?? 0}
    maxPrayer={state?.self?.maxPrayer ?? 1}
  />
)}
```

No changes needed in `App.tsx` — prayer is a SidePanel tab, not a modal.

---

## 7. Styles — `web/src/styles.css`

Add after the `.stats` block (around line 165):

```css
/* ---- prayer tab ---- */
.prayer-tab { padding: 4px 6px; font-size: 12px; }
.prayer-pts {
  color: var(--rsc-cyan);
  padding: 2px 0 6px;
  border-bottom: 1px solid #1a3d1a;
  margin-bottom: 4px;
}
.prayer-list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.prayer-row {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 3px 2px;
  cursor: pointer;
  border-bottom: 1px solid #0e200e;
  color: var(--rsc-white);
}
.prayer-row:hover { background: #0f1f0f; }
.prayer-row.active { color: var(--rsc-green); }
.prayer-row.disabled { color: var(--rsc-gray); }
.prayer-indicator { font-size: 10px; flex-shrink: 0; }
.prayer-name { flex: 1; }
.prayer-level { color: var(--rsc-yellow); flex-shrink: 0; }
```

Color semantics match `100-rsc-ui-map §4.3`:
- active → `@gre@` = `var(--rsc-green)` (`#00ff00`)
- too-low → `@bla@` → mapped to gray (`var(--rsc-gray)`) since we don't have black-on-dark
- available (unlocked, inactive) → white (`var(--rsc-white)`)

---

## 8. API contract summary

| Endpoint | Method | Request | Response | Notes |
|---|---|---|---|---|
| `/state` | GET | — | `{…, prayers: Prayer[]}` | always 14 entries; `active` live |
| `/prayer` | POST | `{id:0–13, on:bool}` | `{ok:bool, message?:string}` | enqueueAction; guard 0-pts |

`Prayer` wire shape (mirrors `statePrayer` Go struct):
```
{ id, name, reqLevel, drainRate, description, active }
```

---

## 9. Test plan

1. **Backend unit check** — `go build ./cmd/cradle` must compile green after the
   additions (the new types + handler are straightforward; no new imports needed
   beyond what is already in scope: `context`, `fmt`, `encoding/json`, `facts`).

2. **`GET /state` prayer block** — `curl localhost:8090/state | jq '.prayers'`
   must return a 14-element array with `id` 0–13, `name`, `reqLevel`, `drainRate`,
   `description`, and `active:false` when idle.

3. **Activate via `POST /prayer`**:
   ```
   curl -X POST localhost:8090/prayer -d '{"id":0,"on":true}'
   # expect {"ok":true,"message":"Activate prayer 0"}
   curl localhost:8090/state | jq '.prayers[0].active'
   # expect: true  (after server confirms opcode 206)
   ```

4. **Deactivate**:
   ```
   curl -X POST localhost:8090/prayer -d '{"id":0,"on":false}'
   curl localhost:8090/state | jq '.prayers[0].active'
   # expect: false
   ```

5. **Guard — 0 prayer points**: while `self.prayer == 0`, POST with `on:true`
   must return `{ok:false, message:"prayer points are 0"}`.

6. **Guard — bad id**: `{"id":99,"on":true}` → `{ok:false, message:"id must be 0..13"}`.

7. **SPA Prayer tab** — in the browser, click "Prayer" in the tab bar:
   - 14 rows appear, each with name and Lv requirement.
   - Prayer points header shows current/max.
   - Left-click a row → `active` indicator flips on next `/state` poll (~400ms).
   - Right-click a row → context menu with Activate/Deactivate items.
   - Prayers with `reqLevel > maxPrayer` render in gray.

8. **No regression** — Inventory, Stats, bank modal, and chat still work
   normally after the SidePanel `Tab` type expansion.
