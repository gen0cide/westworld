// Prayer tab — lists all 14 prayer slots with live active state.
// Left-click toggles; right-click opens an Activate/Deactivate context menu.

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
