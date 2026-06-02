// Inventory: fixed 30-slot grid. Left-click fires the slot's default option;
// right-click opens the item menu using the inline options from /state (no
// extra round-trip). Mirrors the original renderInventory()/handlers.

import { act as apiAct } from '../api'
import { useUI } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { InvItem, MenuTarget } from '../types'

const INV_SLOTS = 30

function slotRef(it: InvItem): MenuTarget {
  return { kind: 'inventory_item', index: 0, x: 0, y: 0, dir: 0, id: it.itemId, slot: it.slot }
}

function qtyBadge(amount: number): string {
  return amount >= 1000 ? (amount / 1000).toFixed(1) + 'k' : String(amount)
}

export function InventoryGrid({ inventory }: { inventory: InvItem[] }) {
  const ui = useUI()
  const bySlot = new Map<number, InvItem>()
  for (const it of inventory) bySlot.set(it.slot, it)

  const onLeft = (it: InvItem) => {
    void apiAct(slotRef(it), it.defaultOptionId).then((r) => { if (r.message) ui.flash(r.message) })
  }
  const onRight = (e: React.MouseEvent, it: InvItem) => {
    e.preventDefault()
    ui.openMenu(e.clientX, e.clientY, [{ ref: slotRef(it), label: it.name, options: it.options }])
  }

  return (
    <div className="inv">
      {Array.from({ length: INV_SLOTS }, (_, i) => {
        const it = bySlot.get(i)
        if (!it) return <div key={i} className="cell" />
        return (
          <div
            key={i}
            className={'cell' + (it.wielded ? ' wield' : '')}
            title={it.name}
            onClick={() => onLeft(it)}
            onContextMenu={(e) => onRight(e, it)}
          >
            <ItemSprite id={it.itemId} name={it.name} />
            {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
          </div>
        )
      })}
    </div>
  )
}
