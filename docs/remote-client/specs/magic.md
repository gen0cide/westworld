# Spec: Magic Spellbook Tab (D2)

**Scope:** `cmd/cradle/remoteclient.go`, `web/src/` only. No edits to `world/`,
`runtime/`, or `reference/`.

**Task tree reference:** 110-react-port.md §D2.

---

## 1. Design decisions

### 1.1 Static catalog via `GET /spells` (NOT per-tick `/state`)

The spell catalog (`facts.Spells`, 96 entries) is read from an embedded XML at
`init()` and never changes at runtime. Including it in every `/state` response
would send ~10 KB of static data every 400 ms. Instead:

- `GET /spells` — **served once at page load**; returns the full catalog with all
  rune requirements. Response is immutable; the SPA caches it in module state.
- `/state` gains a small `magic` sub-object (≤ 200 bytes) that carries only the
  **per-tick runtime delta**: current magic level, boosted level, and per-spell
  boolean flags (`hasRunes`, `canCast`). The SPA merges those flags over the
  static catalog on each poll.

This mirrors the existing separation of static item defs (served via `/sprite`)
vs live inventory state (served via `/state`).

### 1.2 Cast endpoint: `POST /cast`

Self-targeted and no-target spells are the only ones that make sense from the
spellbook tab (the player cannot meaningfully click an NPC or a map tile from the
side panel). The spec therefore exposes a single endpoint:

```
POST /cast  { "spellId": N, "targetKind"?: "self"|"npc"|"player",
              "targetIndex"?: N }
```

- Omit `targetKind` (or `"self"`) → `host.CastOnSelf`.
- `"npc"` + `targetIndex` → `host.CastOnNpc`.
- `"player"` + `targetIndex` → `host.CastOnPlayer`.

`CastOnLand` and `CastOnInventory` are **not** wired here; they belong to
point-and-click interactions on the viewport/inventory, already handled by the
generic `/act` endpoint via the dispatcher.

---

## 2. Go backend additions (`cmd/cradle/remoteclient.go`)

All additions are in `cmd/cradle/remoteclient.go` or a new
`cmd/cradle/magic.go`. The recommended split: put the new wire-type structs in
`remoteclient.go` adjacent to the existing bank types; put the two route
handlers in a new file `cmd/cradle/magic.go` with a
`registerMagicRoutes(mux, host, f, enqueueAction)` function called once inside
`serveClient`.

### 2.1 New wire structs (add to `remoteclient.go`)

```go
// stateMagic is the per-tick magic block in /state. The full spell catalog
// is served separately by GET /spells.
type stateMagic struct {
    Level    int              `json:"level"`    // current/boosted magic level
    MaxLevel int              `json:"maxLevel"` // base magic level (skill max)
    Spells   []stateMagicFlag `json:"spells"`   // indexed by spell id
}

// stateMagicFlag carries the two runtime booleans for one spell.
// The SPA joins these to the static catalog by array index (= spell id).
type stateMagicFlag struct {
    ID       int  `json:"id"`
    CanCast  bool `json:"canCast"`  // req_level <= current magic level
    HasRunes bool `json:"hasRunes"` // inventory+staff satisfies rune cost
}

// castRequest is the body of POST /cast.
type castRequest struct {
    SpellID     int    `json:"spellId"`
    TargetKind  string `json:"targetKind,omitempty"`  // "self"|"npc"|"player"; empty = self
    TargetIndex int    `json:"targetIndex,omitempty"` // NPC or player server-index
}
```

### 2.2 Extend `stateResponse`

In the `stateResponse` struct add one field after `Bank`:

```go
Magic *stateMagic `json:"magic,omitempty"`
```

### 2.3 Populate `magic` block in the `/state` handler

Inside the existing `/state` handler, after the bank block, add:

```go
// magic block: per-tick flags merged over the static catalog.
var magicBlock *stateMagic
{
    level    := host.World().Self.SkillLevel(6) // magicSkillID = 6
    maxLevel := host.World().Self.SkillMax(6)

    // Build per-spell has-runes map from inventory (same logic as
    // spellHasRunesStaffCallable in runtime/views_magic.go:292-324).
    invSlots := host.World().Inventory.Slots()
    counts := make(map[int]int, len(invSlots))
    staffCovered := make(map[int]bool)
    for _, sl := range invSlots {
        counts[sl.ItemID] += sl.Amount
        if sl.Wielded {
            // elementalStaffItems: Fire(31)→{197,615,682}, Water(32)→{102,616,683},
            // Air(33)→{101,617,684}, Earth(34)→{103,618,685}
            for runeID, staves := range magicStaffItems {
                for _, staveID := range staves {
                    if sl.ItemID == staveID {
                        staffCovered[runeID] = true
                    }
                }
            }
        }
    }
    flags := make([]stateMagicFlag, len(facts.Spells))
    for i, sp := range facts.Spells {
        canCast := sp.ReqLevel <= level
        hasRunes := true
        for _, r := range sp.Runes {
            if staffCovered[r.ItemID] {
                continue
            }
            if counts[r.ItemID] < r.Count {
                hasRunes = false
                break
            }
        }
        flags[i] = stateMagicFlag{ID: sp.ID, CanCast: canCast, HasRunes: hasRunes}
    }
    magicBlock = &stateMagic{Level: level, MaxLevel: maxLevel, Spells: flags}
}
```

Then include `Magic: magicBlock` in the `resp := stateResponse{...}` literal.

**Note:** `magicStaffItems` is a package-level map `map[int][]int` in
`cmd/cradle/magic.go` (copy the same mapping from
`runtime/views_magic.go:79-84`). It is duplicated deliberately — the spec
constraint forbids importing from `runtime/`.

### 2.4 `GET /spells` handler (in `cmd/cradle/magic.go`)

```go
// registerMagicRoutes adds GET /spells and POST /cast to mux.
func registerMagicRoutes(
    mux *http.ServeMux,
    host *runtime.Host,
    f *facts.Facts,
    enqueueAction func(func(context.Context) (string, error)) (string, error),
) {
    // GET /spells — static catalog, served once. The response is fully
    // determined by the embedded XML (facts.Spells); no world state is read.
    // Cache-Control: immutable (facts never change per process lifetime).
    type runeEntry struct {
        ItemID int `json:"itemId"`
        Count  int `json:"count"`
    }
    type spellEntry struct {
        ID          int        `json:"id"`
        Name        string     `json:"name"`
        ReqLevel    int        `json:"reqLevel"`
        Type        int        `json:"type"`
        Exp         int        `json:"exp"`
        Description string     `json:"description"`
        Members     bool       `json:"members"`
        Evil        bool       `json:"evil"`
        Runes       []runeEntry `json:"runes"`
    }
    // Build once at registration time; response bytes are immutable.
    catalog := make([]spellEntry, len(facts.Spells))
    for i, sp := range facts.Spells {
        runes := make([]runeEntry, len(sp.Runes))
        for j, r := range sp.Runes {
            runes[j] = runeEntry{ItemID: r.ItemID, Count: r.Count}
        }
        catalog[i] = spellEntry{
            ID: sp.ID, Name: sp.Name, ReqLevel: sp.ReqLevel,
            Type: int(sp.Type), Exp: sp.ExpReward,
            Description: sp.Description, Members: sp.Members,
            Evil: sp.Evil, Runes: runes,
        }
    }
    catalogJSON, _ := json.Marshal(catalog)

    mux.HandleFunc("/spells", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("Cache-Control", "public, max-age=3600, immutable")
        _, _ = w.Write(catalogJSON)
    })

    // POST /cast — funnels through the serialised action worker.
    mux.HandleFunc("/cast", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
            return
        }
        var req castRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
            return
        }
        if facts.SpellByID(req.SpellID) == nil {
            http.Error(w, "unknown spellId", http.StatusBadRequest)
            return
        }
        w.Header().Set("Content-Type", "application/json; charset=utf-8")
        w.Header().Set("Cache-Control", "no-store")

        spellID := req.SpellID
        targetKind := req.TargetKind
        targetIndex := req.TargetIndex

        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            switch targetKind {
            case "npc":
                return fmt.Sprintf("Cast spell %d on NPC %d", spellID, targetIndex),
                    host.CastOnNpc(wctx, targetIndex, spellID)
            case "player":
                return fmt.Sprintf("Cast spell %d on player %d", spellID, targetIndex),
                    host.CastOnPlayer(wctx, targetIndex, spellID)
            default: // "" or "self"
                return fmt.Sprintf("Cast spell %d on self", spellID),
                    host.CastOnSelf(wctx, spellID)
            }
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
    })
}
```

Call `registerMagicRoutes(mux, host, f, enqueueAction)` inside `serveClient`,
directly after the `registerSpriteRoutes(mux, log)` line.

### 2.5 Host methods called (exact signatures with file:line)

| Method | File:line | Signature |
|--------|-----------|-----------|
| `host.CastOnSelf` | `runtime/magic.go:11` | `func (h *Host) CastOnSelf(ctx context.Context, spellID int) error` |
| `host.CastOnNpc` | `runtime/magic.go:18` | `func (h *Host) CastOnNpc(ctx context.Context, npcServerIndex, spellID int) error` |
| `host.CastOnPlayer` | `runtime/magic.go:24` | `func (h *Host) CastOnPlayer(ctx context.Context, playerServerIndex, spellID int) error` |
| `host.World().Self.SkillLevel(6)` | `world/self.go:162` | Returns current (boosted) magic level |
| `host.World().Self.SkillMax(6)` | `world/self.go:172` | Returns base magic level |
| `host.World().Inventory.Slots()` | world package | Returns `[]world.InvSlot{ItemID, Amount, Wielded, …}` |

Facts package reads (no host call, pure statics):

| Symbol | File:line |
|--------|-----------|
| `facts.Spells` | `facts/spells.go:43` — `[]SpellDef` |
| `facts.SpellByID(id)` | `facts/spells.go:108` — `*SpellDef` |
| `facts.SpellDef.{ID,Name,ReqLevel,Type,ExpReward,Description,Members,Evil,Runes}` | `facts/spells.go:12-19` |
| `facts.Rune.{ItemID,Count}` | `facts/spells.go:37-40` |

---

## 3. Web types additions (`web/src/types.ts`)

Append to `types.ts`:

```ts
/** One entry from GET /spells — static catalog. */
export interface SpellDef {
  id: number
  name: string
  reqLevel: number
  type: number   // 1=self 2=offensive 3=curse 4=inventory 5=teleother 6=summon
  exp: number
  description: string
  members: boolean
  evil: boolean
  runes: Array<{ itemId: number; count: number }>
}

/** Per-tick magic flags from /state.magic (indexed by spell id). */
export interface MagicFlag {
  id: number
  canCast: boolean
  hasRunes: boolean
}

export interface MagicState {
  level: number
  maxLevel: number
  spells: MagicFlag[]  // parallel to the static catalog; index = spell id
}

// Extend GameState:
// Add to the GameState interface: `magic?: MagicState`
```

Also add `magic?: MagicState` to the `GameState` interface block.

---

## 4. Web API additions (`web/src/api.ts`)

```ts
import type { SpellDef, ActResponse } from './types'

// Module-level cache: fetched once, never re-fetched.
let _spellCatalog: SpellDef[] | null = null

/** GET /spells — static; cached after first fetch. */
export async function getSpells(): Promise<SpellDef[]> {
  if (_spellCatalog) return _spellCatalog
  _spellCatalog = await (await fetch('/spells')).json() as SpellDef[]
  return _spellCatalog
}

export type CastTargetKind = 'self' | 'npc' | 'player'

/** POST /cast — routes through the serialised action worker server-side. */
export async function castSpell(
  spellId: number,
  targetKind: CastTargetKind = 'self',
  targetIndex = 0,
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/cast', { spellId, targetKind, targetIndex })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

---

## 5. Component: `<Spellbook>` (`web/src/components/Spellbook.tsx`)

### Props

```ts
interface SpellbookProps {
  magic: MagicState          // from state.magic (per-tick flags)
  spells: SpellDef[]         // full catalog from getSpells() (static)
}
```

### Behaviour

- Rendered inside `SidePanel` when the `'magic'` tab is active.
- The component calls `getSpells()` in a `useEffect` on mount (once; cached).
- Displays every spell in catalog order (index = spell id). Each row:
  - **Left-click**: call `castSpell(spell.id)` → `ui.flash(result.message)`.
    Only self-targeted spells (type === 1) or offensive (type === 2) with no
    target index make practical sense from the panel; the same endpoint accepts
    all; the client sends without a target (defaults to self-cast for all).
    Implementer note: offensive spells with no target will be rejected or
    no-op server-side (the server validates them via the underlying protocol);
    the tab is still useful for teleport/heal/enchant-via-panel.
  - **Right-click**: `ui.openActions` with a single section containing:
    - `"Cast on self"` (always present; calls `castSpell(spell.id, 'self')`)
    - `"Cast on NPC…"` (present when type===2 or type===3) — placeholder; no
      NPC picker is wired yet; show but call `ui.flash('select NPC from map')`.
    - `"Examine"` — shows `spell.description` via `ui.flash`.

### Color coding (per 100-rsc-ui-map.md §4.3)

| Condition | Color |
|-----------|-------|
| `canCast && hasRunes` | `--rsc-yellow` (`#ffff00`) — have runes |
| `canCast && !hasRunes` | `--rsc-white` (`#ffffff`) — level ok, missing rune(s) |
| `!canCast` | `--rsc-gray` (`#808080`) — level too low |

### Rune cost display

Below the spell name, render the rune list as a row of small text badges
`"<count>×<itemName>"`. Rune item names are resolved from the static catalog
(the implementer should maintain a small hardcoded map of the ~10 rune item ids
to short names — e.g., `{33: 'Air', 34: 'Earth', 31: 'Fire', 32: 'Water',
35: 'Mind', 36: 'Chaos', 38: 'Death', 39: 'Blood', 40: 'Nature', 41: 'Law',
42: 'Cosmic', 43: 'Body', 44: 'Soul', 619: 'Fire(∞)', 620: 'Water(∞)',
621: 'Air(∞)', 622: 'Earth(∞)'}`) rather than calling `/sprite` for each rune
(the `/sprite` pipeline only covers item-picture sprites; rune icons are text
stubs here unless the implementer also wires rune sprites later).

### Component skeleton

```tsx
import { useEffect, useState } from 'react'
import { castSpell, getSpells } from '../api'
import { useUI, type MenuSection } from '../ui'
import type { MagicState, SpellDef } from '../types'

// Short rune-name lookup: rune item id → short display label.
const RUNE_NAMES: Record<number, string> = {
  31: 'Fire', 32: 'Water', 33: 'Air', 34: 'Earth',
  35: 'Mind', 36: 'Chaos', 38: 'Death', 39: 'Blood',
  40: 'Nature', 41: 'Law', 42: 'Cosmic', 43: 'Body', 44: 'Soul',
}

export function Spellbook({ magic }: { magic: MagicState }) {
  const ui = useUI()
  const [catalog, setCatalog] = useState<SpellDef[]>([])

  useEffect(() => {
    void getSpells().then(setCatalog)
  }, [])

  // Build a flag lookup by spell id.
  const flagMap = new Map(magic.spells.map((f) => [f.id, f]))

  const cast = (spellId: number) => {
    void castSpell(spellId).then((r) => ui.flash(r.message ?? (r.ok ? 'ok' : 'failed')))
  }

  const spellMenu = (e: React.MouseEvent, sp: SpellDef) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: sp.name,
      items: [
        { text: 'Cast on self', run: () => cast(sp.id) },
        ...(sp.type === 2 || sp.type === 3
          ? [{ text: 'Cast on NPC…', run: () => ui.flash('select NPC from map') }]
          : []),
        { text: 'Examine', run: () => ui.flash(sp.description) },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  return (
    <div className="spellbook">
      <div className="spellhdr">
        Magic — level {magic.level}/{magic.maxLevel}
      </div>
      {catalog.map((sp) => {
        const flag = flagMap.get(sp.id)
        const canCast  = flag?.canCast  ?? false
        const hasRunes = flag?.hasRunes ?? false
        const colorClass = !canCast ? 'sp-low' : hasRunes ? 'sp-ready' : 'sp-norune'
        const runeText = sp.runes
          .map((r) => `${r.count}×${RUNE_NAMES[r.itemId] ?? `#${r.itemId}`}`)
          .join(' ')
        return (
          <div
            key={sp.id}
            className={`spellrow ${colorClass}`}
            title={sp.description}
            onClick={() => canCast && cast(sp.id)}
            onContextMenu={(e) => spellMenu(e, sp)}
          >
            <span className="sp-lvl">Lv{sp.reqLevel}</span>
            <span className="sp-name">{sp.name}</span>
            {runeText && <span className="sp-runes">{runeText}</span>}
          </div>
        )
      })}
    </div>
  )
}
```

---

## 6. `SidePanel` changes (`web/src/components/SidePanel.tsx`)

```tsx
// Change Tab type:
type Tab = 'inv' | 'stats' | 'magic'

// Add import:
import { Spellbook } from './Spellbook'

// Add tab button after Stats button:
<button className={tab === 'magic' ? 'active' : ''} onClick={() => setTab('magic')}>Magic</button>

// Add to tabbody switch:
// Inside the ternary, extend to a three-way:
{tab === 'inv' ? (
  <>
    <EquipmentPanel equipment={state?.equipment ?? []} />
    <InventoryGrid inventory={state?.inventory ?? []} />
  </>
) : tab === 'stats' ? (
  <StatsPanel self={state?.self ?? null} />
) : (
  state?.magic
    ? <Spellbook magic={state.magic} />
    : <div className="spellhdr">Loading…</div>
)}
```

`SidePanel` receives `state: GameState | null` (unchanged prop type); `state.magic`
is already present since `GameState` was extended in step 3.

---

## 7. CSS additions (`web/src/styles.css`)

Add after the `.stats` block:

```css
/* ---- spellbook (magic tab) ---- */
.spellbook { padding: 4px 6px; font-size: 12px; }
.spellhdr  { color: var(--rsc-green); padding: 2px 0 4px;
             border-bottom: 1px solid var(--panel-border);
             margin-bottom: 4px; font-weight: bold; }
.spellrow  {
  display: flex; flex-direction: column; gap: 1px;
  padding: 3px 4px; border-bottom: 1px solid #111;
  cursor: pointer;
}
.spellrow:hover { background: #0d200d; }
.sp-ready  { color: var(--rsc-yellow); }
.sp-norune { color: var(--rsc-white); }
.sp-low    { color: var(--rsc-gray); cursor: default; }
.sp-lvl    { font-size: 10px; opacity: .7; margin-right: 4px; }
.sp-name   { font-weight: bold; }
.sp-runes  { font-size: 10px; color: #9a9; margin-top: 1px; }
```

---

## 8. `App.tsx` changes

None required. The spellbook data flows through `SidePanel` which already
receives `state`. No new modal window (the spellbook lives inside the side
panel, not in a `modal-backdrop` overlay).

---

## 9. `state.ts` impact

None. `useGameState` polls `/state` and returns the full `GameState` object.
Once `GameState.magic` is added to `types.ts`, the `magic` field is
automatically available on every poll with no changes to `state.ts`.

---

## 10. Test plan

### 10.1 Backend (`GET /spells`)

- `curl http://localhost:8090/spells | python3 -m json.tool` — should return a
  JSON array of 96 objects each with fields `id, name, reqLevel, type, exp,
  description, members, evil, runes`.
- Re-request: confirm `Cache-Control: public, max-age=3600, immutable` header is
  present.
- Spell 0 should be `{"id":0,"name":"Wind strike","reqLevel":1,"type":2,...}`.

### 10.2 Backend (`/state` magic block)

- `curl http://localhost:8090/state | python3 -m json.tool | grep -A5 magic`
  — should contain `"magic": {"level": N, "maxLevel": N, "spells": [...]}`.
- `spells` array length must equal 96.
- `canCast` for spell 0 (Wind Strike, reqLevel 1) should be `true` for any
  Magic-leveled character.
- `hasRunes` for Wind Strike (`runes: [{itemId:35,count:1},{itemId:33,count:1}]`)
  should be `false` when the character holds no runes, `true` when they hold at
  least 1 Mind and 1 Air rune.

### 10.3 Backend (`POST /cast`)

- `curl -X POST -d '{"spellId":0}' http://localhost:8090/cast` — returns
  `{"ok":true,"message":"Cast spell 0 on self"}` (or server-error if the spell
  is not castable in this context; `ok:false` is acceptable).
- `curl -X POST -d '{"spellId":9999}' http://localhost:8090/cast` — returns
  HTTP 400 `unknown spellId`.
- `curl -X POST -d '{"spellId":0,"targetKind":"npc","targetIndex":3}'
  http://localhost:8090/cast` — dispatches `CastOnNpc(ctx, 3, 0)` on the action
  worker; returns `ok:true` (server may reject with a protocol error if NPC 3
  is not visible, which surfaces as `ok:false` with the error message).

### 10.4 UI smoke test

1. Open `http://localhost:8090/` in a browser (or run `npm run dev` at
   `web/` against a live cradle on `:8090`).
2. A **Magic** tab button appears in the side panel's tab bar.
3. Click **Magic**: a scrollable list of spells renders with `Lv1 Wind strike`
   at the top. Level-too-low spells are gray; available-but-no-runes are white;
   ready-to-cast are yellow.
4. Left-click a yellow spell: `ui.flash` shows the server's response message.
5. Right-click any spell: a context menu appears with `Cast on self` / (for
   attack spells) `Cast on NPC…` / `Examine`. `Examine` flashes the spell
   description.
6. Switch to Inventory and back to Magic: the spell list is still rendered
   without a second `/spells` fetch (catalog cached in module state).
7. Open the browser Network panel and verify `/spells` is fetched once with
   status 200 and served on the next switch with a 200 from cache (not
   re-fetched).
