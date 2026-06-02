// Right-hand panel with a small tab bar. Inventory tab shows equipment summary
// + the 30-slot grid; Stats tab shows skills. The pixel-perfect pass will grow
// this into the authentic RSC tab strip (inventory/stats/magic/prayer/…).

import { useState } from 'react'
import type { GameState } from '../types'
import { EquipmentPanel } from './EquipmentPanel'
import { InventoryGrid } from './InventoryGrid'
import { StatsPanel } from './StatsPanel'

type Tab = 'inv' | 'stats'

export function SidePanel({ state }: { state: GameState | null }) {
  const [tab, setTab] = useState<Tab>('inv')
  return (
    <div id="side">
      <div className="tabbar">
        <button className={tab === 'inv' ? 'active' : ''} onClick={() => setTab('inv')}>Inventory</button>
        <button className={tab === 'stats' ? 'active' : ''} onClick={() => setTab('stats')}>Stats</button>
      </div>
      <div className="tabbody">
        {tab === 'inv' ? (
          <>
            <EquipmentPanel equipment={state?.equipment ?? []} />
            <InventoryGrid inventory={state?.inventory ?? []} />
          </>
        ) : (
          <StatsPanel self={state?.self ?? null} />
        )}
      </div>
    </div>
  )
}
