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
          <span>Trading with: <strong>{trade.partnerName || '...'}</strong></span>
          <button onClick={() => void tradeAction('decline')}>Decline X</button>
        </div>

        <div className="tradecols">
          {/* Left: my offer */}
          <div className="tradecol">
            <div className="tradecolhdr">
              Your offer {trade.myFirstAccepted && !isConfirmPhase ? '(accepted)' : ''}
            </div>
            <div className="tradegrid">
              {trade.myOffer.length === 0 && <div className="empty">nothing</div>}
              {trade.myOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell"
                  title={`${it.name || '#' + it.itemId} x${it.amount} - left-click to remove 1`}
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
              {trade.partnerName || 'Their'} offer {trade.theirFirstAccepted && !isConfirmPhase ? '(accepted)' : ''}
            </div>
            <div className="tradegrid">
              {trade.theirOffer.length === 0 && <div className="empty">nothing</div>}
              {trade.theirOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell readonly"
                  title={`${it.name || '#' + it.itemId} x${it.amount}`}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Inventory strip - hidden in confirm phase (offers locked) */}
        {!isConfirmPhase && (
          <>
            <div className="bsec">Your inventory (click to add to offer)</div>
            <div className="bankgrid">
              {inventory.length === 0 && <div className="empty">inventory empty</div>}
              {inventory.map((it) => (
                <div
                  key={it.slot}
                  className="cell"
                  title={`${it.name} x${it.amount} - click to add 1`}
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
            {myAccepted ? '(waiting...)' : acceptLabel}
          </button>
          {theyAccepted && !myAccepted && (
            <span className="tradeready">{trade.partnerName} accepted!</span>
          )}
        </div>
      </div>
    </div>
  )
}
