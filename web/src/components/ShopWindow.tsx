// Shop window — shown (as a modal over the viewport) only while the server has
// the shop open (state.shop.open). Top grid = shop stock (buy), bottom = your
// inventory (sell). Left-click moves 1; right-click opens a quantity menu
// (1/5/10/All). After any action the /state poll refreshes both grids.

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
