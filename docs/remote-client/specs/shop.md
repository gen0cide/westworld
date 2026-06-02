# Shop Window — Implementation Spec (E2)

Clone the bank feature (ad9633c) exactly. Only files under `web/` and
`cmd/cradle/` change. `world/` and `runtime/` are read-only; all needed
Host methods already exist.

---

## 1. Backend state fields exposed to the SPA

`world.ShopState.Shop()` (world/shop.go:67) returns `*world.ShopRecord` (or nil
when closed).  The fields the SPA needs:

| ShopRecord field  | Type      | Meaning |
|-------------------|-----------|---------|
| `IsGeneral`       | bool      | true = general store (buys anything player carries) |
| `SellPriceMod`    | int       | % the shop pays player (e.g. 30 → 30% of basePrice) |
| `BuyPriceMod`     | int       | % player pays shop (e.g. 130 → 130% of basePrice) |
| `PriceMultiplier` | int       | stock-drift sensitivity (0 = fixed price) |
| `Slots`           | []ShopSlot| catalogue entries |

Each `world.ShopSlot`:

| field       | type | meaning |
|-------------|------|---------|
| `ItemID`    | int  | catalogue/def id |
| `Stock`     | int  | current quantity available to buy (0 = out of stock) |
| `BaseStock` | int  | baseline quantity; used for price drift |

The displayed prices are computed per-slot in the Go handler (not in the SPA)
using `world.ShopState.BuyPrice(itemID, basePrice)` and
`world.ShopState.SellPrice(itemID, basePrice)` (world/shop.go:162 and 183).
`basePrice` comes from `facts.Facts.ItemDef(itemID).BasePrice` (facts/defs.go:91).

---

## 2. Go additions to `cmd/cradle/remoteclient.go`

### 2a. New wire structs (add after `stateBankSlot` / `bankRequest` block, ~line 318)

```go
// stateShop mirrors world.ShopRecord for the SPA. Present only while the
// shop window is open (nil in /state when closed).
type stateShop struct {
    Open      bool            `json:"open"`
    IsGeneral bool            `json:"isGeneral"`
    Slots     []stateShopSlot `json:"slots"`
}

type stateShopSlot struct {
    ItemID    int    `json:"itemId"`
    Name      string `json:"name"`
    Stock     int    `json:"stock"`
    BuyPrice  int    `json:"buyPrice"`  // gp player pays, 0 if basePrice unknown
    SellPrice int    `json:"sellPrice"` // gp shop pays player, 0 if basePrice unknown
}

// shopRequest is the body of POST /shop.
// op: "buy" | "sell" | "close"
// For buy/sell: ItemID is the catalogue id, Amount the quantity.
type shopRequest struct {
    Op     string `json:"op"`
    ItemID int    `json:"itemId"`
    Amount int    `json:"amount"`
}
```

### 2b. Add `Shop *stateShop` field to `stateResponse` (after the Bank field, ~line 304)

```go
// Shop is present only while the shop window is open.
Shop *stateShop `json:"shop,omitempty"`
```

### 2c. Populate the shop block in the GET /state handler (after the bank block, ~line 962)

Insert immediately after the `bankBlock` assignment block and before
`resp := stateResponse{...}`:

```go
// shop block: present only while the shop window is open.
var shopBlock *stateShop
if rec := host.World().Shop.Shop(); rec != nil {
    sslots := make([]stateShopSlot, 0, len(rec.Slots))
    for _, ss := range rec.Slots {
        name := ""
        buyPrice := 0
        sellPrice := 0
        if f != nil {
            if def := f.ItemDef(ss.ItemID); def != nil {
                name = def.Name
                buyPrice = host.World().Shop.BuyPrice(ss.ItemID, def.BasePrice)
                sellPrice = host.World().Shop.SellPrice(ss.ItemID, def.BasePrice)
            }
        }
        sslots = append(sslots, stateShopSlot{
            ItemID:    ss.ItemID,
            Name:      name,
            Stock:     ss.Stock,
            BuyPrice:  buyPrice,
            SellPrice: sellPrice,
        })
    }
    shopBlock = &stateShop{Open: true, IsGeneral: rec.IsGeneral, Slots: sslots}
}
```

Add `Shop: shopBlock` to the `stateResponse` struct literal.

### 2d. POST /shop handler (add after the POST /bank handler, ~line 1041)

```go
// POST /shop — buy/sell/close on the open shop window (110-react §E2).
mux.HandleFunc("/shop", func(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req shopRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
        return
    }
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.Header().Set("Cache-Control", "no-store")
    switch req.Op {
    case "buy", "sell":
        if !host.World().Shop.IsOpen() {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "shop is not open"})
            return
        }
        if req.Amount <= 0 {
            _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "amount must be > 0"})
            return
        }
    case "close":
    default:
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "unknown op: must be buy|sell|close"})
        return
    }
    op, id, amount := req.Op, req.ItemID, req.Amount
    msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
        switch op {
        case "buy":
            return fmt.Sprintf("Buy %d of item %d", amount, id), host.ShopBuy(wctx, id, amount)
        case "sell":
            return fmt.Sprintf("Sell %d of item %d", amount, id), host.ShopSell(wctx, id, amount)
        case "close":
            return "Close shop", host.ShopClose(wctx)
        }
        return "", fmt.Errorf("unknown shop op %q", op)
    })
    if runErr != nil {
        _ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
        return
    }
    _ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
})
```

### 2e. Host method signatures called (no changes to these files)

| Method | File:line | Signature |
|--------|-----------|-----------|
| `host.World().Shop.Shop()` | world/shop.go:67 | `func (s *ShopState) Shop() *ShopRecord` |
| `host.World().Shop.IsOpen()` | world/shop.go:79 | `func (s *ShopState) IsOpen() bool` |
| `host.World().Shop.BuyPrice(itemID, basePrice int)` | world/shop.go:162 | `func (s *ShopState) BuyPrice(itemID, basePrice int) int` |
| `host.World().Shop.SellPrice(itemID, basePrice int)` | world/shop.go:183 | `func (s *ShopState) SellPrice(itemID, basePrice int) int` |
| `host.ShopBuy(ctx, catalogID, amount int)` | runtime/shop.go:23 | `func (h *Host) ShopBuy(ctx context.Context, catalogID, amount int) error` |
| `host.ShopSell(ctx, catalogID, amount int)` | runtime/shop.go:29 | `func (h *Host) ShopSell(ctx context.Context, catalogID, amount int) error` |
| `host.ShopClose(ctx)` | runtime/shop.go:35 | `func (h *Host) ShopClose(ctx context.Context) error` |

---

## 3. Web layer

### 3a. `web/src/types.ts` additions

Add after the `Bank` / `BankSlot` block and update `GameState`:

```typescript
export interface ShopSlot {
  itemId: number
  name: string
  stock: number
  buyPrice: number   // gp the player pays; 0 if unknown
  sellPrice: number  // gp the shop pays the player; 0 if unknown
}

export interface Shop {
  open: boolean
  isGeneral: boolean
  slots: ShopSlot[]
}
```

In `GameState`, add:

```typescript
shop?: Shop  // present only while the shop window is open
```

### 3b. `web/src/api.ts` addition

Add after `bankAction`:

```typescript
export type ShopOp = 'buy' | 'sell' | 'close'

/** POST /shop — buy/sell a quantity of a catalogue item, or close the window. */
export async function shopAction(op: ShopOp, itemId = 0, amount = 0): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/shop', { op, itemId, amount })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

---

## 4. `web/src/components/ShopWindow.tsx` (new file)

Modelled exactly on `BankWindow.tsx`. Props: `shop: Shop` + `inventory: InvItem[]`.

Layout:
- `modal-backdrop` > `shopwin` div
- Header row (`.shophdr`): shop title ("General Store" if `isGeneral`, else "Shop") + Close button
- Section label `.bsec` "Shop Stock" + `.shopgrid` — one cell per `shop.slots` entry.
  Each cell: `<ItemSprite id name/>` + price badge (`.price`, yellow) + stock badge
  (`.qty`, standard). Left-click = buy 1 (calls `shopAction('buy', itemId, 1)`).
  Right-click = quantity menu: Buy 1 / Buy 5 / Buy 10 / Buy All (stock) via `ui.openActions`.
  Out-of-stock cells (`stock === 0`) get `className="cell dimmed"` and only show the
  sell path (or are disabled for buy; see below).
- Section label `.bsec` "Your Inventory" + `.shopgrid` — one cell per inventory item.
  Left-click = sell 1 (`shopAction('sell', it.itemId, 1)`).
  Right-click = quantity menu: Sell 1 / Sell 5 / Sell 10 / Sell All (amount).
  For general stores, show all inventory items (stock may be 0 in shop.slots for
  player-brought items). For specialty stores, show only items that appear in
  `shop.slots` (i.e. `shop.slots.some(s => s.itemId === it.itemId)`).

```tsx
// web/src/components/ShopWindow.tsx
import { shopAction } from '../api'
import { useUI, type MenuSection } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { Shop, InvItem } from '../types'

function qtyBadge(a: number): string { return a >= 1000 ? (a / 1000).toFixed(1) + 'k' : String(a) }
function gpBadge(p: number): string { return p > 0 ? (p >= 1000 ? (p / 1000).toFixed(1) + 'k' : String(p)) + 'gp' : '?' }

const PRESETS = [1, 5, 10]

export function ShopWindow({ shop, inventory }: { shop: Shop; inventory: InvItem[] }) {
  const ui = useUI()

  const buy = (itemId: number, amount: number) => {
    void shopAction('buy', itemId, amount).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const sell = (itemId: number, amount: number) => {
    void shopAction('sell', itemId, amount).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const buyMenu = (e: React.MouseEvent, itemId: number, stock: number, name: string) => {
    e.preventDefault()
    if (stock === 0) return  // nothing to buy
    const items = [
      ...PRESETS.map((n) => ({ text: `Buy ${n}`, run: () => buy(itemId, n) })),
      { text: `Buy All (${stock})`, run: () => buy(itemId, stock) },
    ]
    const sections: MenuSection[] = [{ header: name, items }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const sellMenu = (e: React.MouseEvent, itemId: number, have: number, name: string) => {
    e.preventDefault()
    const items = [
      ...PRESETS.map((n) => ({ text: `Sell ${n}`, run: () => sell(itemId, n) })),
      { text: `Sell All (${have})`, run: () => sell(itemId, have) },
    ]
    const sections: MenuSection[] = [{ header: name, items }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  // For specialty shops, only inventory items stocked by the shop are sellable.
  const sellableInv = shop.isGeneral
    ? inventory
    : inventory.filter((it) => shop.slots.some((s) => s.itemId === it.itemId))

  return (
    <div className="modal-backdrop">
      <div className="shopwin">
        <div className="shophdr">
          <span>{shop.isGeneral ? 'General Store' : 'Shop'}</span>
          <button onClick={() => shopAction('close')}>Close ✕</button>
        </div>

        <div className="bsec">Shop Stock</div>
        <div className="shopgrid">
          {shop.slots.length === 0 && <div className="empty">shop is empty</div>}
          {shop.slots.map((s) => (
            <div
              key={s.itemId}
              className={'cell' + (s.stock === 0 ? ' dimmed' : '')}
              title={`${s.name || 'item ' + s.itemId} — Buy: ${gpBadge(s.buyPrice)} (${s.stock} in stock)`}
              onClick={() => s.stock > 0 && buy(s.itemId, 1)}
              onContextMenu={(e) => buyMenu(e, s.itemId, s.stock, s.name || `item ${s.itemId}`)}
            >
              <ItemSprite id={s.itemId} name={s.name} />
              <div className="price">{gpBadge(s.buyPrice)}</div>
              {s.stock > 1 && <div className="qty">{qtyBadge(s.stock)}</div>}
            </div>
          ))}
        </div>

        <div className="bsec">Your Inventory</div>
        <div className="shopgrid">
          {sellableInv.length === 0 && <div className="empty">nothing to sell here</div>}
          {sellableInv.map((it) => {
            const shopSlot = shop.slots.find((s) => s.itemId === it.itemId)
            const sp = shopSlot?.sellPrice ?? 0
            return (
              <div
                key={it.slot}
                className="cell"
                title={`${it.name} — Sell: ${gpBadge(sp)} (you have ${it.amount})`}
                onClick={() => sell(it.itemId, 1)}
                onContextMenu={(e) => sellMenu(e, it.itemId, it.amount, it.name)}
              >
                <ItemSprite id={it.itemId} name={it.name} />
                <div className="price">{gpBadge(sp)}</div>
                {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}
```

---

## 5. `web/src/App.tsx` wiring

Add import at the top alongside BankWindow:

```typescript
import { ShopWindow } from './components/ShopWindow'
```

In the JSX return block, add after the `BankWindow` conditional:

```tsx
{state?.shop?.open && (
  <ShopWindow shop={state.shop} inventory={state.inventory} />
)}
```

The full JSX tail becomes:

```tsx
      {state?.bank?.open && (
        <BankWindow bank={state.bank} inventory={state.inventory} />
      )}
      {state?.shop?.open && (
        <ShopWindow shop={state.shop} inventory={state.inventory} />
      )}
    </div>
  )
```

---

## 6. `web/src/styles.css` additions

Append to the end of the file (inside the modal windows section, after `.bankgrid .empty`).

```css
/* ---- shop window ---- */
.shopwin {
  background: #0a140a;
  border: 2px solid #1a3d1a;
  box-shadow: 0 0 0 2px #000, 4px 4px 0 #000;
  width: 444px;
  max-height: 88vh;
  overflow-y: auto;
  padding: 0 0 8px;
}
.shophdr {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: #1a3d1a;
  color: #cfc;
  padding: 5px 8px;
  font-weight: bold;
}
.shophdr button {
  background: #06100a;
  color: #f99;
  border: 1px solid #000;
  font: 12px monospace;
  padding: 2px 8px;
  cursor: pointer;
}
.shophdr button:hover { background: #a00; color: #000; }
/* shopgrid reuses bankgrid column layout (8-wide, 49px cells) */
.shopgrid {
  display: grid;
  grid-template-columns: repeat(8, 49px);
  gap: 1px;
  padding: 4px 8px;
}
.shopgrid .empty { color: #9a9; padding: 4px; grid-column: 1 / -1; }
/* price badge: yellow gp label below the icon, inside the cell */
.cell .price {
  position: absolute;
  left: 0;
  right: 0;
  bottom: 0;
  color: var(--rsc-yellow);
  font-size: 8px;
  text-align: center;
  text-shadow: 0 0 2px #000;
  pointer-events: none;
}
/* dimmed = out-of-stock shop slot */
.cell.dimmed { opacity: 0.4; cursor: default; }
```

Note: `.cell` and `.qty` are already defined in the shared block; `.price` is new.
The `.bsec` section label class is shared with the bank window — no new rule needed.

---

## 7. Test plan

1. **Go build check**: `go build ./cmd/cradle/` compiles without error after adding
   the four struct types and the `/shop` handler.

2. **Closed shop — /state poll**: With no shop open, `GET /state` must return JSON
   with `"shop":null` (field absent due to `omitempty`) — verify with `curl`.

3. **Open shop — /state fields**: Open a shop in-game on the bot. Poll `/state`;
   confirm `shop.open == true`, `shop.slots` is non-empty, each slot has `itemId`,
   `name` (non-empty for known items), `stock`, `buyPrice > 0`, `sellPrice > 0`.
   Confirm `isGeneral` matches the shop type.

4. **Buy — POST /shop**: Send `{"op":"buy","itemId":<id>,"amount":1}` while the
   shop is open; expect `{"ok":true,"message":"Buy 1 of item <id>"}`. Confirm the
   bot's inventory gains the item and the shop re-sends stock (world mirror updates
   on the next SEND_SHOP_OPEN packet — no optimistic update needed).

5. **Sell — POST /shop**: Send `{"op":"sell","itemId":<id>,"amount":1}` while
   carrying the item; expect `{"ok":true}`. Confirm inventory decreases.

6. **Close — POST /shop**: Send `{"op":"close"}`; expect `{"ok":true}`. Next
   `/state` poll must return `shop: null`. (ShopClose eagerly clears the world
   mirror at runtime/shop.go:35 so this is immediate.)

7. **Guard — shop not open**: Send a buy/sell while shop is nil; expect
   `{"ok":false,"message":"shop is not open"}`.

8. **Guard — bad amount**: `{"op":"buy","itemId":1,"amount":0}` → `{"ok":false}`.

9. **Guard — unknown op**: `{"op":"steal"}` → `{"ok":false}`.

10. **SPA smoke test**: Open the shop in-game, reload the SPA. The ShopWindow
    modal must appear. Shop Stock grid shows items with yellow gp prices. Out-of-
    stock slots are dimmed (`opacity 0.4`). Left-click a stocked item triggers a
    buy-1. Right-click shows Buy 1/5/10/All menu. Inventory grid shows sellable
    items; left-click sells 1, right-click shows Sell menu. Close button dismisses
    the modal (next /state poll confirms shop is gone). For a general store, all
    carried items appear in the inventory sell panel. For a specialty store, only
    items listed in shop.slots appear.
