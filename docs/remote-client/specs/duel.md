# Spec: Duel Window (E4)

> Mechanical implementer's spec — copy-pasteable Go structs, handler logic,
> TS types, component sketches, and wiring. Do not implement until this spec
> is reviewed. Do NOT touch `reference/`, `world/`, or `runtime/`.

---

## 1. Backend: Host methods called (file:line)

All in `runtime/duel.go`:

| Method | Signature | Line |
|---|---|---|
| `InitDuelRequest` | `(ctx, serverIndex int) error` | :14 |
| `AcceptIncomingDuel` | `(ctx, fromIndex int) error` | :43 |
| `DeclineDuel` | `(ctx) error` | :54 |
| `OfferDuelItems` | `(ctx, items []world.TradeItem) error` | :61 |
| `SetDuelRules` | `(ctx, r world.DuelRules) error` | :76 |
| `AcceptDuelOffer` | `(ctx) error` — first-screen accept | :88 |
| `AcceptDuelConfirm` | `(ctx) error` — confirm-screen accept | :102 |

`world.TradeItem{ItemID int, Amount int}` (world/trade.go:49-51).
`world.DuelRules{DisallowRetreat, DisallowMagic, DisallowPrayer, DisallowWeapons bool}`
(world/duel.go:43-48).

World state is accessed via `host.World().Duel` (`*world.DuelState`):
- `Duel() *DuelRecord` — snapshot (nil if no active duel; world/duel.go:61)
- `IsActive() bool` — true iff phase not "completed"/"cancelled" (world/duel.go:74)

`DuelRecord` fields used by the handler:

```go
type DuelRecord struct {
    WithIndex           int          // opponent server index
    WithName            string       // opponent username
    MyOffer             []TradeItem
    TheirOffer          []TradeItem
    MyFirstAccepted     bool
    TheirFirstAccepted  bool
    MySecondAccepted    bool
    TheirSecondAccepted bool
    Rules               DuelRules
    Phase               string       // "request_sent"|"open"|"confirm"|"completed"|"cancelled"
}
```

---

## 2. Backend additions: `cmd/cradle/remoteclient.go`

### 2.1 New wire types — add to `remoteclient.go` alongside the bank block

```go
// stateDuelItem mirrors world.TradeItem for the duel stake grids.
type stateDuelItem struct {
    ItemID int    `json:"itemId"`
    Name   string `json:"name"`
    Amount int    `json:"amount"`
}

// stateDuelRules mirrors world.DuelRules (true = disallowed).
type stateDuelRules struct {
    DisallowRetreat bool `json:"disallowRetreat"`
    DisallowMagic   bool `json:"disallowMagic"`
    DisallowPrayer  bool `json:"disallowPrayer"`
    DisallowWeapons bool `json:"disallowWeapons"`
}

// stateDuel is present in /state only while a duel is active
// (phase == "open" or "confirm"). nil closes the <DuelWindow>.
type stateDuel struct {
    Phase               string          `json:"phase"`              // "open"|"confirm"
    WithName            string          `json:"withName"`
    MyOffer             []stateDuelItem `json:"myOffer"`
    TheirOffer          []stateDuelItem `json:"theirOffer"`
    Rules               stateDuelRules  `json:"rules"`
    MyFirstAccepted     bool            `json:"myFirstAccepted"`
    TheirFirstAccepted  bool            `json:"theirFirstAccepted"`
    MySecondAccepted    bool            `json:"mySecondAccepted"`
    TheirSecondAccepted bool            `json:"theirSecondAccepted"`
}

// duelRequest is the body of POST /duel.
// op: "stake" | "rules" | "accept1" | "accept2" | "decline"
//   stake  — replace our staked items; Items is the full new list.
//   rules  — update rule toggles; Rules is the full new set.
//   accept1 — first accept (offer screen); Items/Rules ignored.
//   accept2 — final accept (confirm screen); Items/Rules ignored.
//   decline — abort duel at any phase.
type duelRequest struct {
    Op    string          `json:"op"`
    Items []stateDuelItem `json:"items,omitempty"` // for op=stake
    Rules *stateDuelRules `json:"rules,omitempty"` // for op=rules
}
```

### 2.2 Extend `stateResponse`

```go
type stateResponse struct {
    Self      stateSelf        `json:"self"`
    Inventory []stateInvItem   `json:"inventory"`
    Equipment []stateEquipItem `json:"equipment"`
    Chat      []chatEntry      `json:"chat"`
    Bank      *stateBank       `json:"bank,omitempty"`
    Duel      *stateDuel       `json:"duel,omitempty"` // ADD THIS
}
```

### 2.3 Populate `duelBlock` in the `/state` handler

Insert this block immediately after the `bankBlock` population (around line 963
in the current file):

```go
// duel block: present only while a duel is in the "open" or "confirm" phase.
var duelBlock *stateDuel
if rec := host.World().Duel.Duel(); rec != nil &&
    (rec.Phase == "open" || rec.Phase == "confirm") {

    toItems := func(src []world.TradeItem) []stateDuelItem {
        out := make([]stateDuelItem, 0, len(src))
        for _, it := range src {
            name := ""
            if f != nil {
                if def := f.ItemDef(it.ItemID); def != nil {
                    name = def.Name
                }
            }
            out = append(out, stateDuelItem{ItemID: it.ItemID, Name: name, Amount: it.Amount})
        }
        return out
    }

    duelBlock = &stateDuel{
        Phase:               rec.Phase,
        WithName:            rec.WithName,
        MyOffer:             toItems(rec.MyOffer),
        TheirOffer:          toItems(rec.TheirOffer),
        Rules:               stateDuelRules(rec.Rules),
        MyFirstAccepted:     rec.MyFirstAccepted,
        TheirFirstAccepted:  rec.TheirFirstAccepted,
        MySecondAccepted:    rec.MySecondAccepted,
        TheirSecondAccepted: rec.TheirSecondAccepted,
    }
}
```

Note: `stateDuelRules(rec.Rules)` works because both structs have the same four
`bool` fields in the same order; a direct cast is safe.

Then in the `stateResponse` literal:

```go
resp := stateResponse{
    // ... existing fields ...
    Bank: bankBlock,
    Duel: duelBlock, // ADD
}
```

### 2.4 `POST /duel` handler

Register immediately after `POST /bank`:

```go
// POST /duel — stake items, set rules, accept (two-stage), or decline.
// Duel uses the same serialized action worker as /bank and /act.
mux.HandleFunc("/duel", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req duelRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")

    // Guard: all ops except "decline" require an active duel.
    if req.Op != "decline" && !host.World().Duel.IsActive() {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "no active duel"})
        return
    }

    switch req.Op {
    case "stake":
        // Convert stateDuelItem → world.TradeItem.
        items := make([]world.TradeItem, 0, len(req.Items))
        for _, it := range req.Items {
            if it.Amount <= 0 {
                _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "item amount must be > 0"})
                return
            }
            items = append(items, world.TradeItem{ItemID: it.ItemID, Amount: it.Amount})
        }
        captured := items
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return fmt.Sprintf("Stake %d item type(s)", len(captured)),
                host.OfferDuelItems(wctx, captured)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "rules":
        if req.Rules == nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "rules object required for op=rules"})
            return
        }
        r2 := world.DuelRules(*req.Rules)
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Set duel rules", host.SetDuelRules(wctx, r2)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "accept1":
        rec := host.World().Duel.Duel()
        if rec == nil || rec.Phase != "open" {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "duel not in offer phase"})
            return
        }
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Accept duel offer", host.AcceptDuelOffer(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "accept2":
        rec := host.World().Duel.Duel()
        if rec == nil || rec.Phase != "confirm" {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "duel not in confirm phase"})
            return
        }
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Confirm duel", host.AcceptDuelConfirm(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "decline":
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Decline duel", host.DeclineDuel(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    default:
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "unknown op: must be stake|rules|accept1|accept2|decline"})
    }
})
```

Note: `world.DuelRules(*req.Rules)` is a direct struct cast — safe because
`stateDuelRules` and `world.DuelRules` have the same four `bool` fields in the
same order and both are in-package plain structs.

---

## 3. Frontend: `web/src/types.ts`

Add alongside the `Bank`/`BankSlot` block:

```typescript
export interface DuelItem {
  itemId: number
  name: string
  amount: number
}

export interface DuelRules {
  disallowRetreat: boolean
  disallowMagic: boolean
  disallowPrayer: boolean
  disallowWeapons: boolean
}

export interface Duel {
  phase: 'open' | 'confirm'
  withName: string
  myOffer: DuelItem[]
  theirOffer: DuelItem[]
  rules: DuelRules
  myFirstAccepted: boolean
  theirFirstAccepted: boolean
  mySecondAccepted: boolean
  theirSecondAccepted: boolean
}
```

Extend `GameState`:

```typescript
export interface GameState {
  self: Self
  inventory: InvItem[]
  equipment: EquipItem[]
  chat: ChatEntry[]
  bank?: Bank
  duel?: Duel  // ADD: present only while duel window is open
}
```

---

## 4. Frontend: `web/src/api.ts`

Add alongside `bankAction`:

```typescript
export type DuelOp = 'stake' | 'rules' | 'accept1' | 'accept2' | 'decline'

export interface DuelStakeItem { itemId: number; amount: number }

export interface DuelRulesPayload {
  disallowRetreat: boolean
  disallowMagic: boolean
  disallowPrayer: boolean
  disallowWeapons: boolean
}

/** POST /duel */
export async function duelAction(
  op: DuelOp,
  items?: DuelStakeItem[],
  rules?: DuelRulesPayload,
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/duel', { op, items, rules })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

---

## 5. Frontend: `web/src/components/DuelWindow.tsx`

Create `web/src/components/DuelWindow.tsx`. Full implementation sketch:

```typescript
// DuelWindow — modal shown when state.duel is present (phase "open" or
// "confirm"). Two-stage handshake matching RSC's authentic UI (§4.8).
//
// Offer screen (phase="open"):
//   Left grid  — our staked items (from inventory; left-click = add 1;
//                right-click = qty menu; items already staked can be
//                clicked in the My Offer grid to remove).
//   Right grid — opponent's staked items (read-only).
//   Rules row  — four toggles (retreat/magic/prayer/weapons). Either side
//                can flip; the server unifies and broadcasts the result.
//   Accept row — "Accept Offer" button (disabled if already accepted;
//                highlights if TheirFirstAccepted is true to prompt the user).
//   Decline    — "Decline" button (always enabled).
//
// Confirm screen (phase="confirm"):
//   Two read-only grids (locked items) + read-only rules — no changes allowed.
//   "Confirm Fight" button + "Decline" button.

import { duelAction } from '../api'
import { useUI, type MenuSection } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { Duel, DuelItem, DuelRules, InvItem } from '../types'

function qtyBadge(a: number): string {
  return a >= 1000 ? (a / 1000).toFixed(1) + 'k' : String(a)
}

const PRESETS = [1, 5, 10]
const RULE_LABELS: [keyof DuelRules, string][] = [
  ['disallowRetreat',  'No Retreat'],
  ['disallowMagic',    'No Magic'],
  ['disallowPrayer',   'No Prayer'],
  ['disallowWeapons',  'No Weapons'],
]

interface Props {
  duel: Duel
  inventory: InvItem[]
}

export function DuelWindow({ duel, inventory }: Props) {
  const ui = useUI()

  // ---- offer screen helpers -------------------------------------------

  const stakeItem = (itemId: number, amount: number) => {
    // Merge with the current staked list: if the item is already staked,
    // add to its amount; otherwise append. Always send the full list.
    const existing = [...duel.myOffer]
    const idx = existing.findIndex((i) => i.itemId === itemId)
    if (idx >= 0) {
      existing[idx] = { ...existing[idx], amount: existing[idx].amount + amount }
    } else {
      existing.push({ itemId, amount })
    }
    void duelAction('stake', existing).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const unstakeItem = (itemId: number) => {
    const updated = duel.myOffer.filter((i) => i.itemId !== itemId)
    void duelAction('stake', updated).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const stakeMenu = (
    e: React.MouseEvent,
    itemId: number,
    have: number,
    name: string,
  ) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: name,
      items: [
        ...PRESETS.map((n) => ({
          text: `Stake ${n}`,
          run: () => stakeItem(itemId, Math.min(n, have)),
        })),
        { text: `Stake All (${have})`, run: () => stakeItem(itemId, have) },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const unstakeMenu = (
    e: React.MouseEvent,
    item: DuelItem,
  ) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: item.name || `item ${item.itemId}`,
      items: [
        ...PRESETS.filter((n) => n < item.amount).map((n) => ({
          text: `Remove ${n}`,
          run: () => {
            const updated = duel.myOffer.map((i) =>
              i.itemId === item.itemId ? { ...i, amount: i.amount - n } : i,
            ).filter((i) => i.amount > 0)
            void duelAction('stake', updated)
          },
        })),
        { text: 'Remove All', run: () => unstakeItem(item.itemId) },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const toggleRule = (key: keyof DuelRules) => {
    const updated = { ...duel.rules, [key]: !duel.rules[key] }
    void duelAction('rules', undefined, updated).then((r) => {
      if (r.message) ui.flash(r.message)
    })
  }

  const accept1 = () => {
    void duelAction('accept1').then((r) => { if (r.message) ui.flash(r.message) })
  }
  const accept2 = () => {
    void duelAction('accept2').then((r) => { if (r.message) ui.flash(r.message) })
  }
  const decline = () => {
    void duelAction('decline').then((r) => { if (r.message) ui.flash(r.message) })
  }

  // ---- render ---------------------------------------------------------

  const isConfirm = duel.phase === 'confirm'

  return (
    <div className="modal-backdrop">
      <div className="duelwin">

        {/* Header */}
        <div className="duelhdr">
          <span>
            {isConfirm
              ? `Duel vs ${duel.withName} — Confirm`
              : `Duel vs ${duel.withName}`}
          </span>
          <button className="duel-decline-btn" onClick={decline}>Decline ✕</button>
        </div>

        {/* Stake grids */}
        <div className="duelgrids">

          <div className="duelside">
            <div className="duelside-hdr">Your Stake</div>
            {/* My staked items (left-click removes; right-click qty-remove) */}
            <div className="dualgrid">
              {duel.myOffer.length === 0 && (
                <div className="empty">nothing staked</div>
              )}
              {duel.myOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell"
                  title={`${it.name || `item ${it.itemId}`} ×${it.amount} — click to remove`}
                  onClick={() => !isConfirm && unstakeItem(it.itemId)}
                  onContextMenu={(e) => !isConfirm && unstakeMenu(e, it)}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>

            {/* Inventory (source of items to stake) — hidden on confirm screen */}
            {!isConfirm && (
              <>
                <div className="duelside-hdr">From Inventory (click to stake)</div>
                <div className="dualgrid">
                  {inventory.length === 0 && <div className="empty">inventory empty</div>}
                  {inventory.map((it) => (
                    <div
                      key={it.slot}
                      className="cell"
                      title={`${it.name} ×${it.amount} — click to stake 1`}
                      onClick={() => stakeItem(it.itemId, 1)}
                      onContextMenu={(e) => stakeMenu(e, it.itemId, it.amount, it.name)}
                    >
                      <ItemSprite id={it.itemId} name={it.name} />
                      {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                    </div>
                  ))}
                </div>
              </>
            )}
          </div>

          <div className="duelside">
            <div className="duelside-hdr">Their Stake ({duel.withName})</div>
            <div className="dualgrid">
              {duel.theirOffer.length === 0 && (
                <div className="empty">nothing staked</div>
              )}
              {duel.theirOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell duel-readonly"
                  title={`${it.name || `item ${it.itemId}`} ×${it.amount}`}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>
          </div>

        </div>{/* /duelgrids */}

        {/* Rules row */}
        <div className="duelrules">
          <span className="duelrules-lbl">Rules:</span>
          {RULE_LABELS.map(([key, label]) => (
            <button
              key={key}
              className={`rule-toggle${duel.rules[key] ? ' active' : ''}`}
              onClick={() => !isConfirm && toggleRule(key)}
              title={duel.rules[key] ? `${label} (enforced)` : `${label} (allowed)`}
              disabled={isConfirm}
            >
              {label}
            </button>
          ))}
        </div>

        {/* Accept / confirm row */}
        <div className="duelactions">
          {!isConfirm ? (
            <button
              className={`duel-accept-btn${duel.theirFirstAccepted ? ' prompt' : ''}`}
              onClick={accept1}
              disabled={duel.myFirstAccepted}
              title={duel.myFirstAccepted ? 'Waiting for opponent…' : 'Accept Offer'}
            >
              {duel.myFirstAccepted ? 'Waiting…' : 'Accept Offer'}
              {duel.theirFirstAccepted && !duel.myFirstAccepted && ' ← they accepted'}
            </button>
          ) : (
            <button
              className={`duel-accept-btn${duel.theirSecondAccepted ? ' prompt' : ''}`}
              onClick={accept2}
              disabled={duel.mySecondAccepted}
              title={duel.mySecondAccepted ? 'Waiting for opponent…' : 'Confirm Fight'}
            >
              {duel.mySecondAccepted ? 'Waiting…' : 'Confirm Fight'}
              {duel.theirSecondAccepted && !duel.mySecondAccepted && ' ← they confirmed'}
            </button>
          )}
          <button className="duel-decline-btn" onClick={decline}>Decline</button>
        </div>

      </div>{/* /duelwin */}
    </div>
  )
}
```

---

## 6. Frontend: CSS additions (`web/src/styles.css`)

Add after `.bankgrid` rules (the comment `/* ---- modal windows … ---- */`
already names duel as a future window):

```css
/* ---- duel window ---- */
.duelwin {
  background: #0a140a;
  border: 2px solid #1a3d1a;
  box-shadow: 0 0 0 2px #000, 4px 4px 0 #000;
  width: 540px;
  max-height: 92vh;
  overflow-y: auto;
  padding: 0 0 8px;
}
.duelhdr {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #1a3d1a;
  color: #cfc;
  padding: 5px 8px;
  font-weight: bold;
}
.duelhdr button, .duelactions button {
  background: #06100a;
  border: 1px solid #000;
  font: 12px monospace;
  padding: 2px 8px;
  cursor: pointer;
}
.duelgrids {
  display: flex;
  gap: 8px;
  padding: 4px 8px;
}
.duelside {
  flex: 1;
  min-width: 0;
}
.duelside-hdr {
  color: #6c6;
  padding: 4px 0 2px;
  border-bottom: 1px solid #1a3d1a;
  font-size: 11px;
  margin-bottom: 2px;
}
.dualgrid {
  display: grid;
  grid-template-columns: repeat(4, 49px);
  gap: 1px;
  padding: 2px 0;
  min-height: 36px;
}
.dualgrid .empty { color: #9a9; padding: 4px; grid-column: 1 / -1; font-size: 11px; }
.duel-readonly { opacity: 0.75; cursor: default; }
/* Rules toggles */
.duelrules {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-top: 1px solid #1a3d1a;
  flex-wrap: wrap;
}
.duelrules-lbl { color: #6c6; font-size: 11px; margin-right: 4px; }
.rule-toggle {
  background: #06100a;
  color: #9a9;
  border: 1px solid #1a3d1a;
  font: 11px monospace;
  padding: 2px 6px;
  cursor: pointer;
}
.rule-toggle.active {
  background: #8b0000;
  color: #fcc;
  border-color: #f00;
}
.rule-toggle:disabled { opacity: 0.5; cursor: default; }
/* Accept / decline row */
.duelactions {
  display: flex;
  gap: 8px;
  padding: 6px 8px 0;
  border-top: 1px solid #1a3d1a;
}
.duel-accept-btn {
  background: #1a3d1a;
  color: #cfc;
  border: 1px solid #3a6d3a;
  font: 12px monospace;
  padding: 3px 12px;
  cursor: pointer;
}
.duel-accept-btn.prompt {
  /* Opponent has clicked; prompt ours */
  border-color: var(--rsc-yellow);
  color: var(--rsc-yellow);
}
.duel-accept-btn:disabled { opacity: 0.55; cursor: default; }
.duel-decline-btn { color: #f99; }
.duel-decline-btn:hover { background: #a00; color: #000; }
```

---

## 7. Frontend: App.tsx wiring

In `web/src/App.tsx`, add:

```typescript
import { DuelWindow } from './components/DuelWindow'
```

In the JSX return, alongside the bank conditional:

```tsx
{state?.duel && (state.duel.phase === 'open' || state.duel.phase === 'confirm') && (
  <DuelWindow duel={state.duel} inventory={state.inventory} />
)}
```

The `/state` poll (400 ms cadence via `useGameState`) is sufficient for duel
responsiveness; no SSE needed for the spec. If two-party latency becomes
noticeable, G1 (SSE push) can be applied later as a drop-in.

---

## 8. Interaction semantics

### Offer screen (`phase="open"`)

| UI gesture | Action |
|---|---|
| Left-click inventory item | `duelAction('stake', mergedList)` — adds 1 to stake |
| Right-click inventory item | qty menu → `duelAction('stake', mergedList)` — adds N |
| Left-click "My Stake" cell | `duelAction('stake', filteredList)` — removes all of that item |
| Right-click "My Stake" cell | qty menu → partial removal |
| Click rule toggle | `duelAction('rules', undefined, updatedRules)` |
| Click "Accept Offer" | `duelAction('accept1')` — button disabled after click; "Waiting…" |
| Click "Decline" | `duelAction('decline')` — closes window |

Rule toggle resets both first-accept flags server-side; `TheirFirstAccepted`
snaps back to false on the next `/state` poll.

### Confirm screen (`phase="confirm"`)

All grids are read-only. Only "Confirm Fight" and "Decline" are interactive.
"Confirm Fight" is disabled after click. The `theirSecondAccepted` flag
highlighting mirrors the offer screen's `theirFirstAccepted` prompt.

### Transition rules

- Window appears: `/state` emits `duel != null` with phase "open".
- Transition to confirm: `/state` changes phase to "confirm". `<DuelWindow>`
  re-renders automatically (same component, new props).
- Duel ends (fight or cancel): `/state` `duel` becomes null; `<DuelWindow>`
  unmounts. No explicit "close" packet from the browser needed for the happy
  path; "Decline" sends `op=decline` explicitly.

---

## 9. Test plan

1. **Unit — Go handler**  
   - POST `/duel` with no active duel and `op=stake` → `{ok:false, message:"no active duel"}`.  
   - POST `/duel` with `op=rules` and `rules=null` → `{ok:false, message:"rules object required…"}`.  
   - POST `/duel` with `op=accept1` when `phase="confirm"` → `{ok:false, message:"duel not in offer phase"}`.  
   - POST `/duel` with `op=accept2` when `phase="open"` → `{ok:false, message:"duel not in confirm phase"}`.  
   - POST `/duel` with unknown `op` → `{ok:false, message:"unknown op: must be stake|rules|accept1|accept2|decline"}`.

2. **Integration — /state**  
   - While no duel active: `/state` → `duel` field absent (omitempty).  
   - With `world.Duel.MarkOpened(…)` called in a test harness (or via a second
     bot): `/state` → `duel.phase="open"`, `duel.withName` matches, item
     arrays present (may be empty), rules all false.

3. **UI — manual two-bot test**  
   - Bot A initiates duel with Bot B (via /act → "Duel" option on player).  
   - Both /state responses show `duel.phase="open"`.  
   - Bot A browser: stake 1 item → item appears in My Stake grid.  
   - Bot B browser: confirm their offer shows Bot A's staked item.  
   - Both click "Accept Offer" → phase flips to "confirm" in the next poll.  
   - Both click "Confirm Fight" → `duel` disappears from `/state`; fight begins.  
   - Decline path: click "Decline" at any phase → `duel` disappears on next poll.

4. **UI — rule toggles**  
   - Click "No Magic" toggle → it becomes red (`active`); next `/state` poll
     shows `rules.disallowMagic=true`.  
   - Click it again → toggle deactivates; next poll shows `false`.  
   - After toggling, both `myFirstAccepted` and `theirFirstAccepted` become
     false (verify via `/state`); "Accept Offer" re-enables.

5. **TypeScript compilation gate**  
   - `npm --prefix web run build` must remain green after adding the new types,
     api function, component, and App.tsx import/JSX.

---

## 10. What is NOT in this spec

- Opening a duel from the game world. That path already works through the
  existing `/pick` → `/act` → `Duel` menu option dispatch (dispatcher layer 2
  calls `host.InitDuelRequest`). No new endpoint needed to initiate a duel.
- Accepting an incoming duel request. That is triggered by the opponent's
  client; the server emits `SEND_DUEL_WINDOW` which flips world state to
  `phase="open"`. The browser user sees the `<DuelWindow>` automatically on
  the next `/state` poll. The `AcceptIncomingDuel` Host method is not called
  from the HTTP layer — it is the DSL path used by autonomous routines.
- SSE/WebSocket push (G1) — the 400 ms poll is sufficient to spec this
  feature; lower latency is a later optimization.
