// DuelWindow — modal shown when state.duel is present (phase "open" or
// "confirm"). Two-stage handshake matching RSC's authentic UI.
//
// Offer screen (phase="open"):
//   Left grid  — our staked items (left-click = remove all; right-click = qty menu).
//   Right grid — opponent's staked items (read-only).
//   Inventory  — source items to stake (left-click = add 1; right-click = qty menu).
//   Rules row  — four toggles (retreat/magic/prayer/weapons).
//   Accept row — "Accept Offer" button (disabled if already accepted).
//   Decline    — "Decline" button (always enabled).
//
// Confirm screen (phase="confirm"):
//   Two read-only grids + read-only rules — no changes allowed.
//   "Confirm Fight" button + "Decline" button.

import { duelAction, type DuelStakeItem } from '../api'
import { useUI, type MenuSection } from '../ui'
import { ItemSprite } from './ItemSprite'
import type { Duel, DuelItem, DuelRules, InvItem } from '../types'

function qtyBadge(a: number): string {
  return a >= 1000 ? (a / 1000).toFixed(1) + 'k' : String(a)
}

const PRESETS = [1, 5, 10]

const RULE_LABELS: [keyof DuelRules, string][] = [
  ['disallowRetreat', 'No Retreat'],
  ['disallowMagic', 'No Magic'],
  ['disallowPrayer', 'No Prayer'],
  ['disallowWeapons', 'No Weapons'],
]

interface Props {
  duel: Duel
  inventory: InvItem[]
}

export function DuelWindow({ duel, inventory }: Props) {
  const ui = useUI()

  // Merge a delta into the current staked list and return the full new list.
  const mergeStake = (
    current: DuelItem[],
    itemId: number,
    delta: number,
  ): DuelStakeItem[] => {
    const map = new Map(current.map((i) => [i.itemId, i.amount]))
    const cur = map.get(itemId) ?? 0
    const next = cur + delta
    if (next <= 0) map.delete(itemId)
    else map.set(itemId, next)
    return Array.from(map.entries()).map(([id, amount]) => ({ itemId: id, amount }))
  }

  const sendStake = (newItems: DuelStakeItem[]) => {
    void duelAction('stake', newItems).then((r) => { if (r.message) ui.flash(r.message) })
  }

  const stakeItem = (itemId: number, amount: number) => {
    sendStake(mergeStake(duel.myOffer, itemId, amount))
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

  const unstakeMenu = (e: React.MouseEvent, item: DuelItem) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: item.name || `item ${item.itemId}`,
      items: [
        ...PRESETS.filter((n) => n < item.amount).map((n) => ({
          text: `Remove ${n}`,
          run: () => {
            const updated = duel.myOffer
              .map((i) => i.itemId === item.itemId ? { ...i, amount: i.amount - n } : i)
              .filter((i) => i.amount > 0)
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
          <button className="duel-decline-btn" onClick={decline}>Decline X</button>
        </div>

        {/* Stake grids */}
        <div className="duelgrids">

          {/* My stake */}
          <div className="duelside">
            <div className="duelside-hdr">
              Your Stake{duel.myFirstAccepted && !isConfirm ? ' (accepted)' : ''}
            </div>
            <div className="dualgrid">
              {duel.myOffer.length === 0 && (
                <div className="empty">nothing staked</div>
              )}
              {duel.myOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell"
                  title={`${it.name || `item ${it.itemId}`} x${it.amount} — click to remove`}
                  onClick={() => !isConfirm && unstakeItem(it.itemId)}
                  onContextMenu={(e) => !isConfirm && unstakeMenu(e, it)}
                >
                  <ItemSprite id={it.itemId} name={it.name} />
                  {it.amount > 1 && <div className="qty">{qtyBadge(it.amount)}</div>}
                </div>
              ))}
            </div>

            {/* Inventory source — hidden on confirm screen */}
            {!isConfirm && (
              <>
                <div className="duelside-hdr">Inventory (click to stake)</div>
                <div className="dualgrid">
                  {inventory.length === 0 && <div className="empty">inventory empty</div>}
                  {inventory.map((it) => (
                    <div
                      key={it.slot}
                      className="cell"
                      title={`${it.name} x${it.amount} — click to stake 1`}
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

          {/* Their stake (read-only) */}
          <div className="duelside">
            <div className="duelside-hdr">
              {duel.withName}&apos;s Stake{duel.theirFirstAccepted && !isConfirm ? ' (accepted)' : ''}
            </div>
            <div className="dualgrid">
              {duel.theirOffer.length === 0 && (
                <div className="empty">nothing staked</div>
              )}
              {duel.theirOffer.map((it) => (
                <div
                  key={it.itemId}
                  className="cell duel-readonly"
                  title={`${it.name || `item ${it.itemId}`} x${it.amount}`}
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
              className={`duel-accept-btn${duel.theirFirstAccepted && !duel.myFirstAccepted ? ' prompt' : ''}`}
              onClick={accept1}
              disabled={duel.myFirstAccepted}
              title={duel.myFirstAccepted ? 'Waiting for opponent...' : 'Accept Offer'}
            >
              {duel.myFirstAccepted ? 'Waiting...' : 'Accept Offer'}
              {duel.theirFirstAccepted && !duel.myFirstAccepted && ' ← they accepted'}
            </button>
          ) : (
            <button
              className={`duel-accept-btn${duel.theirSecondAccepted && !duel.mySecondAccepted ? ' prompt' : ''}`}
              onClick={accept2}
              disabled={duel.mySecondAccepted}
              title={duel.mySecondAccepted ? 'Waiting for opponent...' : 'Confirm Fight'}
            >
              {duel.mySecondAccepted ? 'Waiting...' : 'Confirm Fight'}
              {duel.theirSecondAccepted && !duel.mySecondAccepted && ' ← they confirmed'}
            </button>
          )}
          <button className="duel-decline-btn" onClick={decline}>Decline</button>
        </div>

      </div>{/* /duelwin */}
    </div>
  )
}
