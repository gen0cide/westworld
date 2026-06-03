# Trade Window — Implementation Spec (E3)

> Spec only — do NOT implement source code from this document alone; a second
> logged-in bot is required to exercise the live two-party handshake.

---

## 1. Backend context

### 1.1 Inbound opcodes (proto/v235/inbound_opcodes.go:33-39)

| Opcode | Const | Meaning |
|---|---|---|
| 92 | `InTradeWindow` | Server asks us to open the trade offer window |
| 97 | `InTradeOtherItems` | Other side's current offer items |
| 15 | `InTradeAccepted` | Server confirms BOTH first-screen accepts → advance to confirm |
| 20 | `InTradeOpenConfirm` | Server opens the second (confirm) screen |
| 162 | `InTradeOtherAccepted` | Other player clicked first-screen Accept |
| 128 | `InTradeClose` | Trade cancelled or completed |

### 1.2 World state mirror (world/trade.go)

`host.World().Trade` is `*world.TradeState`. The only public accessor the HTTP
layer needs:

```go
// world/trade.go:61
func (s *TradeState) Trade() *TradeRecord   // snapshot; nil if no trade active

// world/trade.go:75
func (s *TradeState) IsActive() bool        // true when phase != completed/cancelled
```

`TradeRecord` fields (world/trade.go:19-46):

```go
type TradeRecord struct {
    WithIndex           int         // other player's server index
    WithName            string      // other player's display name
    MyOffer             []TradeItem // {ItemID, Amount}
    TheirOffer          []TradeItem
    MyFirstAccepted     bool
    TheirFirstAccepted  bool
    MySecondAccepted    bool
    TheirSecondAccepted bool
    Phase               string      // "request_sent"|"open"|"confirm"|"completed"|"cancelled"
    OpenedAt            time.Time
    UpdatedAt           time.Time
}
```

### 1.3 Host methods (runtime/trade.go)

| Method | File:line | Sends |
|---|---|---|
| `h.OfferTradeItems(ctx, []world.TradeItem)` | runtime/trade.go:63 | Offer items on screen 1; updates `world.Trade.SetMyOffer` |
| `h.ConfirmTrade(ctx)` | runtime/trade.go:89 | Accept button on screen 1 (first accept); idempotent |
| `h.FinalizeTrade(ctx)` | runtime/trade.go:105 | Accept button on screen 2 (confirm); requires screen-1 accepted first |
| `h.DeclineTrade(ctx)` | runtime/trade.go:55 | Decline/abort in any phase; marks trade cancelled immediately |

**Note:** There is no `RemoveTradeItems` Host method — the RSC protocol replaces
the whole offer atomically. To remove items the client sends a new `OfferTradeItems`
with the revised list (including zero quantity for removed items, or simply
omitting them). The `offer` op therefore doubles as add AND remove.

**Note:** `InitTradeRequest` / `AcceptIncomingTrade` (runtime/trade.go:14,43) are
triggered by right-clicking a player in the world viewport via the existing
`/act` dispatch path — they are NOT new `/trade` ops. The `/trade` endpoint only
handles post-open window interactions.

---

## 2. `/state` trade block

### 2.1 Go types — add to `cmd/cradle/remoteclient.go`

Add these type declarations near the `stateBank` / `stateBankSlot` block
(after line 319 of the current file):

```go
// stateTrade is present in /state only while a trade is active (phase is
// "open" or "confirm"). nil otherwise.
type stateTrade struct {
    Phase               string           `json:"phase"`            // "open"|"confirm"
    PartnerName         string           `json:"partnerName"`
    MyOffer             []stateTradeItem `json:"myOffer"`
    TheirOffer          []stateTradeItem `json:"theirOffer"`
    MyFirstAccepted     bool             `json:"myFirstAccepted"`
    TheirFirstAccepted  bool             `json:"theirFirstAccepted"`
    MySecondAccepted    bool             `json:"mySecondAccepted"`
    TheirSecondAccepted bool             `json:"theirSecondAccepted"`
}

type stateTradeItem struct {
    ItemID int    `json:"itemId"`
    Name   string `json:"name"`   // resolved via facts.ItemDef; "" if unknown
    Amount int    `json:"amount"`
}
```

### 2.2 Add `Trade` field to `stateResponse`

```go
// In the stateResponse struct (currently at line 297), add after Bank:
Trade *stateTrade `json:"trade,omitempty"`
```

### 2.3 Populate in the `/state` handler

Add the following block inside `mux.HandleFunc("/state", ...)` after the
`bankBlock` section (around line 981), before building `resp`:

```go
// trade block: present only while trade phase is "open" or "confirm".
var tradeBlock *stateTrade
if rec := host.World().Trade.Trade(); rec != nil &&
    (rec.Phase == "open" || rec.Phase == "confirm") {
    myItems := make([]stateTradeItem, 0, len(rec.MyOffer))
    for _, ti := range rec.MyOffer {
        name := ""
        if f != nil {
            if def := f.ItemDef(ti.ItemID); def != nil {
                name = def.Name
            }
        }
        myItems = append(myItems, stateTradeItem{
            ItemID: ti.ItemID, Name: name, Amount: ti.Amount,
        })
    }
    theirItems := make([]stateTradeItem, 0, len(rec.TheirOffer))
    for _, ti := range rec.TheirOffer {
        name := ""
        if f != nil {
            if def := f.ItemDef(ti.ItemID); def != nil {
                name = def.Name
            }
        }
        theirItems = append(theirItems, stateTradeItem{
            ItemID: ti.ItemID, Name: name, Amount: ti.Amount,
        })
    }
    tradeBlock = &stateTrade{
        Phase:               rec.Phase,
        PartnerName:         rec.WithName,
        MyOffer:             myItems,
        TheirOffer:          theirItems,
        MyFirstAccepted:     rec.MyFirstAccepted,
        TheirFirstAccepted:  rec.TheirFirstAccepted,
        MySecondAccepted:    rec.MySecondAccepted,
        TheirSecondAccepted: rec.TheirSecondAccepted,
    }
}
```

Update `resp` construction to include `Trade: tradeBlock`.

---

## 3. `POST /trade` endpoint

### 3.1 Wire types — add to `cmd/cradle/remoteclient.go`

```go
// tradeRequest is the body of POST /trade.
// Op values:
//   offer         — replace my entire offer; Items is the new list (may be empty to clear).
//   accept        — click Accept on the offer screen (phase "open" → ConfirmTrade).
//   finalize      — click Accept on the confirm screen (phase "confirm" → FinalizeTrade).
//   decline       — abort the trade in any phase (→ DeclineTrade).
type tradeRequest struct {
    Op    string           `json:"op"`
    Items []tradeItemInput `json:"items,omitempty"` // required for op=offer
}

type tradeItemInput struct {
    ItemID int `json:"itemId"`
    Amount int `json:"amount"`
}
```

### 3.2 Handler — register inside `serveClient` after the `/bank` handler

```go
// POST /trade — interact with the active trade window (110-react §E3).
// All ops run through the serialised action worker.
mux.HandleFunc("/trade", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req tradeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")

    // Guard: trade must be active for all ops except decline.
    if req.Op != "decline" && !host.World().Trade.IsActive() {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "no active trade"})
        return
    }

    switch req.Op {
    case "offer":
        if req.Items == nil {
            req.Items = []tradeItemInput{} // treat missing as clear
        }
        items := make([]world.TradeItem, 0, len(req.Items))
        for _, it := range req.Items {
            if it.Amount <= 0 {
                _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "amount must be > 0"})
                return
            }
            items = append(items, world.TradeItem{ItemID: it.ItemID, Amount: it.Amount})
        }
        op := items // capture
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return fmt.Sprintf("Offer %d item type(s)", len(op)),
                host.OfferTradeItems(wctx, op)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "accept":
        rec := host.World().Trade.Trade()
        if rec == nil || rec.Phase != "open" {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "trade not in offer phase"})
            return
        }
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Accept offer", host.ConfirmTrade(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "finalize":
        rec := host.World().Trade.Trade()
        if rec == nil || rec.Phase != "confirm" {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "trade not in confirm phase"})
            return
        }
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Finalize trade", host.FinalizeTrade(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    case "decline":
        msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
            return "Decline trade", host.DeclineTrade(wctx)
        })
        if runErr != nil {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
            return
        }
        _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})

    default:
        _ = json.NewEncoder(w).Encode(actResponse{
            OK:      false,
            Message: "unknown op: must be offer|accept|finalize|decline",
        })
    }
})
```

**Import note:** The handler body uses `world.TradeItem` — `world` is already
imported in `remoteclient.go`. No new imports needed.

---

## 4. TypeScript types — add to `web/src/types.ts`

```typescript
export interface TradeItem {
  itemId: number
  name: string
  amount: number
}

export interface Trade {
  phase: 'open' | 'confirm'
  partnerName: string
  myOffer: TradeItem[]
  theirOffer: TradeItem[]
  myFirstAccepted: boolean
  theirFirstAccepted: boolean
  mySecondAccepted: boolean
  theirSecondAccepted: boolean
}
```

Add `trade?: Trade` to `GameState`:

```typescript
export interface GameState {
  self: Self
  inventory: InvItem[]
  equipment: EquipItem[]
  chat: ChatEntry[]
  bank?: Bank
  trade?: Trade   // ← add
}
```

---

## 5. TypeScript API client — add to `web/src/api.ts`

```typescript
export type TradeOp = 'offer' | 'accept' | 'finalize' | 'decline'

export interface TradeItemInput {
  itemId: number
  amount: number
}

/** POST /trade — send a trade window action. */
export async function tradeAction(
  op: TradeOp,
  items?: TradeItemInput[],
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/trade', { op, items })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

---

## 6. React component — `web/src/components/TradeWindow.tsx`

### 6.1 Props

```typescript
interface TradeWindowProps {
  trade: Trade        // from state.trade (non-null when shown)
  inventory: InvItem[]
}
```

### 6.2 Structure

The component is a modal (same `.modal-backdrop` wrapper as `BankWindow`).
It has two columns side by side ("Your offer" / "Their offer") plus an
inventory strip below so the user can pick items to add.

**Layout (inner div class `.tradewin`):**

```
┌─────────────────────────────────────────────────────────────┐
│  Trading with: <partnerName>          [Decline ✕]           │  .tradehdr
├─────────────────────┬───────────────────────────────────────┤
│  Your offer         │  Their offer                          │  .tradecolhdr ×2
│  [item grid 4-wide] │  [item grid 4-wide read-only]         │  .tradegrid ×2
├─────────────────────┴───────────────────────────────────────┤
│  ── offer phase ──                                          │
│  Your inventory (click to add to offer)                     │  .bsec (reuse)
│  [inventory strip 8-wide, left-click=add 1, rclick=qty]     │  .bankgrid (reuse)
├─────────────────────────────────────────────────────────────┤
│  [Accept offer]  (yellow if theirFirstAccepted, else grey)  │  .tradebtn
│  phase=confirm: both offers locked; [Confirm trade] button  │
└─────────────────────────────────────────────────────────────┘
```

### 6.3 Full component code

```typescript
// TradeWindow.tsx
// Two-party trade modal. Shown when state.trade.phase is "open" or "confirm".
// Left grid = my offer (left-click removes item; right-click qty to remove).
// Right grid = their offer (read-only display).
// Bottom strip = inventory items available to add to my offer.

import { tradeAction, type TradeItemInput } from '../api'
import { useUI, type MenuSection } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { Trade, InvItem } from '../types'

function qtyBadge(a: number): string { return a >= 1000 ? (a / 1000).toFixed(1) + 'k' : String(a) }

const PRESETS = [1, 5, 10]

export function TradeWindow({ trade, inventory }: { trade: Trade; inventory: InvItem[] }) {
  const ui = useUI()

  // Build the new offer by patching the current myOffer list.
  const patchOffer = (
    items: { itemId: number; amount: number }[],
    delta: { itemId: number; amount: number },
    add: boolean,
  ): TradeItemInput[] => {
    const map = new Map(items.map((i) => [i.itemId, i.amount]))
    const cur = map.get(delta.itemId) ?? 0
    const next = add ? cur + delta.amount : Math.max(0, cur - delta.amount)
    if (next <= 0) map.delete(delta.itemId)
    else map.set(delta.itemId, next)
    return Array.from(map.entries()).map(([itemId, amount]) => ({ itemId, amount }))
  }

  const sendOffer = (newItems: TradeItemInput[]) => {
    void tradeAction('offer', newItems).then((r) => { if (r.message) ui.flash(r.message) })
  }

  // Add from inventory.
  const addItem = (itemId: number, amount: number) => {
    sendOffer(patchOffer(trade.myOffer, { itemId, amount }, true))
  }

  // Remove from my offer.
  const removeItem = (itemId: number, amount: number) => {
    sendOffer(patchOffer(trade.myOffer, { itemId, amount }, false))
  }

  const addQtyMenu = (e: React.MouseEvent, itemId: number, have: number, name: string) => {
    e.preventDefault()
    const items = [
      ...PRESETS.map((n) => ({ text: `Add ${n}`, run: () => addItem(itemId, n) })),
      { text: `Add All (${have})`, run: () => addItem(itemId, have) },
    ]
    const sections: MenuSection[] = [{ header: `Add: ${name}`, items }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const removeQtyMenu = (e: React.MouseEvent, itemId: number, have: number, name: string) => {
    e.preventDefault()
    const items = [
      ...PRESETS.filter((n) => n <= have).map((n) => ({ text: `Remove ${n}`, run: () => removeItem(itemId, n) })),
      { text: `Remove All (${have})`, run: () => removeItem(itemId, have) },
    ]
    const sections: MenuSection[] = [{ header: `Remove: ${name}`, items }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const isConfirmPhase = trade.phase === 'confirm'

  // First-accept button: disabled once we've accepted, highlighted when they've accepted.
  const myAccepted = isConfirmPhase ? trade.mySecondAccepted : trade.myFirstAccepted
  const theyAccepted = isConfirmPhase ? trade.theirSecondAccepted : trade.theirFirstAccepted

  const handleAccept = () => {
    if (isConfirmPhase) {
      void tradeAction('finalize').then((r) => { if (r.message) ui.flash(r.message) })
    } else {
      void tradeAction('accept').then((r) => { if (r.message) ui.flash(r.message) })
    }
  }

  const acceptLabel = isConfirmPhase ? 'Confirm trade' : 'Accept offer'

  return (
    <div className="modal-backdrop">
      <div className="tradewin">
        <div className="tradehdr">
          <span>Trading with: <strong>{trade.partnerName || '…'}</strong></span>
          <button onClick={() => void tradeAction('decline')}>Decline ✕</button>
        </div>

        <div className="tradecols">
          {/* Left: my offer */}
          <div className="tradecol">
            <div className="tradecolhdr">
              Your offer {trade.myFirstAccepted && !isConfirmPhase ? '✓' : ''}
            </div>
            <div className="tradegrid">
              {trade.myOffer.length === 0 && <div className="empty">nothing</div>}
              {trade.myOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell"
                  title={`${it.name || '#' + it.itemId} ×${it.amount} — left-click to remove 1`}
                  onClick={() => removeItem(it.itemId, 1)}
                  onContextMenu={(e) =>
                    removeQtyMenu(e, it.itemId, it.amount, it.name || `#${it.itemId}`)
                  }
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>
          </div>

          {/* Right: their offer (read-only) */}
          <div className="tradecol">
            <div className="tradecolhdr">
              {trade.partnerName || 'Their'} offer {trade.theirFirstAccepted && !isConfirmPhase ? '✓' : ''}
            </div>
            <div className="tradegrid">
              {trade.theirOffer.length === 0 && <div className="empty">nothing</div>}
              {trade.theirOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell readonly"
                  title={`${it.name || '#' + it.itemId} ×${it.amount}`}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Inventory strip — hidden in confirm phase (offers locked) */}
        {!isConfirmPhase && (
          <>
            <div className="bsec">Your inventory (click to add to offer)</div>
            <div className="bankgrid">
              {inventory.length === 0 && <div className="empty">inventory empty</div>}
              {inventory.map((it) => (
                <div
                  key={it.slot}
                  className="cell"
                  title={`${it.name} ×${it.amount} — click to add 1`}
                  onClick={() => addItem(it.itemId, 1)}
                  onContextMenu={(e) => addQtyMenu(e, it.itemId, it.amount, it.name)}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>
          </>
        )}

        {/* Accept / Confirm button row */}
        <div className="tradebtns">
          <button
            className={'tradebtn' + (theyAccepted ? ' partner-ready' : '') + (myAccepted ? ' accepted' : '')}
            disabled={myAccepted}
            onClick={handleAccept}
          >
            {myAccepted ? '(waiting…)' : acceptLabel}
          </button>
          {theyAccepted && !myAccepted && (
            <span className="tradeready">{trade.partnerName} accepted!</span>
          )}
        </div>
      </div>
    </div>
  )
}
```

---

## 7. App.tsx wiring

In `web/src/App.tsx`:

1. Add import:

   ```typescript
   import { TradeWindow } from './components/TradeWindow'
   ```

2. In the JSX return, after the `BankWindow` conditional:

   ```tsx
   {state?.trade && (state.trade.phase === 'open' || state.trade.phase === 'confirm') && (
     <TradeWindow trade={state.trade} inventory={state.inventory} />
   )}
   ```

The trade window and bank window are mutually exclusive by server semantics
(the server only opens one window at a time), but the JSX guard is explicit
about the phase to be safe.

---

## 8. CSS additions — append to `web/src/styles.css`

```css
/* ---- trade window ---- */
.tradewin {
  background: #0a140a;
  border: 2px solid #1a3d1a;
  box-shadow: 0 0 0 2px #000, 4px 4px 0 #000;
  width: 530px;
  max-height: 88vh;
  overflow-y: auto;
  padding: 0 0 8px;
}
.tradehdr {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #1a3d1a;
  color: #cfc;
  padding: 5px 8px;
  font-weight: bold;
}
.tradehdr button {
  background: #06100a;
  color: #f99;
  border: 1px solid #000;
  font: 12px monospace;
  padding: 2px 8px;
  cursor: pointer;
}
.tradehdr button:hover { background: #a00; color: #000; }

.tradecols {
  display: flex;
  gap: 1px;
  border-bottom: 1px solid #1a3d1a;
}
.tradecol { flex: 1; }
.tradecolhdr {
  color: #6c6;
  padding: 4px 8px 2px;
  border-bottom: 1px solid #1a3d1a;
  font-size: 11px;
}
/* 4-wide grid for offer items */
.tradegrid {
  display: grid;
  grid-template-columns: repeat(4, 49px);
  gap: 1px;
  padding: 4px 8px;
  min-height: 40px;
}
.tradegrid .empty { color: #9a9; padding: 4px; grid-column: 1 / -1; font-size: 11px; }
.cell.readonly { cursor: default; }

.tradebtns {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 8px 0;
}
.tradebtn {
  background: #06100a;
  color: #cfc;
  border: 1px solid #1a3d1a;
  font: 12px monospace;
  padding: 4px 14px;
  cursor: pointer;
}
.tradebtn:hover:not(:disabled) { background: #0a0; color: #000; }
.tradebtn:disabled { color: #6c6; cursor: default; }
.tradebtn.partner-ready { border-color: var(--rsc-yellow); color: var(--rsc-yellow); }
.tradebtn.accepted { border-color: #6c6; }
.tradeready { color: var(--rsc-yellow); font-size: 11px; }
```

---

## 9. How the two-phase flow maps to user interactions

| Phase | state.trade.phase | User action | Component does | op sent |
|---|---|---|---|---|
| Offer screen | `open` | Left-click inventory item | add 1 of that item to myOffer | `offer` with updated list |
| Offer screen | `open` | Right-click inventory item | qty context menu (1/5/10/All) | `offer` |
| Offer screen | `open` | Left-click my offer item | remove 1 from myOffer | `offer` |
| Offer screen | `open` | Right-click my offer item | qty-remove context menu | `offer` |
| Offer screen | `open` | Click "Accept offer" | ConfirmTrade (screen 1 accept) | `accept` |
| Confirm screen | `confirm` | Click "Confirm trade" | FinalizeTrade (screen 2 accept) | `finalize` |
| Any | `open`/`confirm` | Click "Decline ✕" | DeclineTrade | `decline` |

**State machine note:** When the user calls `offer` after having accepted
(`myFirstAccepted=true`), the server resets both accept flags. The world mirror
reflects this via `SetMyOffer` → clears all four accepted bools. The `/state`
poll will reflect the reset within 400ms; no special SPA logic needed — just
re-render from the new state.

---

## 10. Test plan

All steps assume two accounts: **A** (cradle-controlled, browser client open)
and **B** (a second logged-in bot or manual client).

1. **Guard — no active trade:**
   POST `/trade {"op":"offer","items":[]}` with no trade in progress →
   expect `{ok:false, message:"no active trade"}`.

2. **Open trade window:**
   Stand A adjacent to B. Right-click B in the viewport → "Trade" option →
   `/act` → `InitTradeRequest`. B accepts (via its own client or DSL script).
   Server emits `InTradeWindow` (opcode 92) to A → `world.Trade.Phase="open"`.
   Next `/state` poll → `state.trade.phase="open"` present; `<TradeWindow>` appears.

3. **Offer items:**
   Left-click a stackable item in the inventory strip → calls `/trade {"op":"offer","items":[{"itemId":...,"amount":1}]}`.
   Verify `/state` next poll shows the item in `state.trade.myOffer`.

4. **Partner offer visible:**
   B offers items from its side. Server emits `InTradeOtherItems` (opcode 97) →
   `world.Trade.TheirOffer` updated. Next `/state` poll shows items in `state.trade.theirOffer`.

5. **First accept:**
   A clicks "Accept offer" → `/trade {"op":"accept"}` → `ConfirmTrade` (runtime/trade.go:89).
   Verify `/state` → `myFirstAccepted:true`. B also accepts → server emits
   `InTradeAccepted` (opcode 15) + `InTradeOpenConfirm` (opcode 20) →
   `world.Trade.Phase="confirm"`. Inventory strip disappears; "Confirm trade" button appears.

6. **Finalize:**
   A clicks "Confirm trade" → `/trade {"op":"finalize"}` → `FinalizeTrade` (runtime/trade.go:105).
   B also finalizes → trade completes. Server emits `InTradeClose` (opcode 128) →
   `world.Trade.Phase="completed"` → `state.trade` drops to nil → `<TradeWindow>` unmounts.

7. **Decline:**
   Repeat step 2. A clicks "Decline ✕" → `/trade {"op":"decline"}` → `DeclineTrade`
   (runtime/trade.go:55). Server emits `InTradeClose` → phase=cancelled →
   `<TradeWindow>` unmounts. Verify via `/state` that `trade` is absent.

8. **Bad op:**
   POST `/trade {"op":"bogus"}` → `{ok:false, message:"unknown op: ..."}`.

9. **Wrong phase guard:**
   While `phase="open"`, POST `/trade {"op":"finalize"}` →
   `{ok:false, message:"trade not in confirm phase"}`.
