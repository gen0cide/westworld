// Bank window — shown (as a modal over the viewport) only while the server has
// the bank open (state.bank.open). Top grid = banked items (withdraw), bottom =
// your inventory (deposit). Left-click moves 1; right-click opens a quantity
// menu (1/5/10/All). After any action the /state poll refreshes both grids.
//
// Icons are name-stubs for now; they swap to authentic <ItemSprite> once the
// GET /sprite pipeline lands (110-react-port §B).

import { bankAction } from '../api'
import { useUI, type MenuSection } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { Bank, InvItem } from '../types'

function qtyBadge(a: number): string { return a >= 1000 ? (a / 1000).toFixed(1) + 'k' : String(a) }

const PRESETS = [1, 5, 10]

export function BankWindow({ bank, inventory }: { bank: Bank; inventory: InvItem[] }) {
  const ui = useUI()

  const move = (op: 'deposit' | 'withdraw', itemId: number, amount: number) => {
    void bankAction(op, itemId, amount).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const quantityMenu = (
    e: React.MouseEvent, op: 'deposit' | 'withdraw', itemId: number, have: number, name: string,
  ) => {
    e.preventDefault()
    const verb = op === 'deposit' ? 'Deposit' : 'Withdraw'
    const items = [
      ...PRESETS.map((n) => ({ text: `${verb} ${n}`, run: () => move(op, itemId, n) })),
      { text: `${verb} All (${have})`, run: () => move(op, itemId, have) },
    ]
    const sections: MenuSection[] = [{ header: name, items }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  return (
    <div className="modal-backdrop">
      <div className="bankwin">
        <div className="bankhdr">
          <span>Bank of Westworld &#8212; {bank.slots.length}/{bank.maxSize || '?'}</span>
          <button onClick={() => bankAction('close')}>Close ✕</button>
        </div>

        <div className="bsec">Bank</div>
        <div className="bankgrid">
          {bank.slots.length === 0 && <div className="empty">bank is empty</div>}
          {bank.slots.map((s) => (
            <div
              key={s.slot}
              className="cell"
              title={`${s.name || 'item ' + s.itemId} ×${s.amount}`}
              onClick={() => move('withdraw', s.itemId, 1)}
              onContextMenu={(e) => quantityMenu(e, 'withdraw', s.itemId, s.amount, s.name || `item ${s.itemId}`)}
            >
              <ItemSprite id={s.itemId} name={s.name} />
              {s.amount > 1 && <div className="qty">{qtyBadge(s.amount)}</div>}
            </div>
          ))}
        </div>

        <div className="bsec">Inventory</div>
        <div className="bankgrid">
          {inventory.length === 0 && <div className="empty">inventory is empty</div>}
          {inventory.map((it) => (
            <div
              key={it.slot}
              className="cell"
              title={`${it.name} ×${it.amount}`}
              onClick={() => move('deposit', it.itemId, 1)}
              onContextMenu={(e) => quantityMenu(e, 'deposit', it.itemId, it.amount, it.name)}
            >
              <ItemSprite id={it.itemId} name={it.name} />
              {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
