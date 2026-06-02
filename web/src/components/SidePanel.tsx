// Right-hand panel with an authentic RSC tab strip (one row of compact icon
// buttons; 100-rsc-ui-map §2). The icons are the real spriteMedia tab graphics:
// the closed tab strip (media 0 = spriteMedia+0) bakes all six 32px icons in a
// row, served via GET /sprite?kind=media&id=0 and CSS-cropped per tab (one fetch,
// browser-cached). Tab state is local UI state, mirroring the original `qc`.
//
// RSC rev 235 only has SIX authentic tab icons: settings/wrench, appearance/face,
// magic/spellbook, stats/chart, map/globe, inventory/backpack. Our React tabs
// don't line up 1:1, so two are best-effort: Worn → the wrench (gear) icon and
// Pray → the map/globe icon — neither has a dedicated icon in this rev (prayer
// shares the magic tab; there is no worn/friends tab). See ICON cell offsets.

import { useState } from 'react'
import type { GameState } from '../types'
import { EquipmentPanel } from './EquipmentPanel'
import { InventoryGrid } from './InventoryGrid'
import { StatsPanel } from './StatsPanel'
import { Spellbook } from './Spellbook'
import { PrayerTab } from './PrayerTab'
import { FriendsTab } from './FriendsTab'

type Tab = 'inv' | 'equip' | 'stats' | 'magic' | 'prayer' | 'friends'

interface TabDef {
  id: Tab
  cell: number // index 0-5 of this tab's icon within the closed media-0 strip
  label: string
  disabled?: boolean
}

// Each icon is a 32px cell in the closed tab strip (media 0). Left→right the
// strip holds: 0=settings/wrench, 1=appearance/face, 2=magic/spellbook,
// 3=stats/chart, 4=map/globe, 5=inventory/backpack. Tabs ordered to match RSC qc.
const TABS: TabDef[] = [
  { id: 'inv',     cell: 5, label: 'Inv' },     // backpack (exact)
  { id: 'equip',   cell: 0, label: 'Worn' },    // wrench/gear (closest authentic)
  { id: 'stats',   cell: 3, label: 'Stats' },   // bar-chart (exact)
  { id: 'magic',   cell: 2, label: 'Magic' },   // spellbook (exact)
  { id: 'prayer',  cell: 4, label: 'Pray' },    // map/globe (no prayer icon in rev)
  { id: 'friends', cell: 1, label: 'Friends' }, // smiley face (appearance)
]

// Pixel offset of icon cell N inside the 197x32 media-0 strip (32px + 1px gap).
const ICON_X = (cell: number) => cell * 33

export function SidePanel({ state }: { state: GameState | null }) {
  const [tab, setTab] = useState<Tab>('inv')
  return (
    <div id="side">
      <div className="tabstrip" role="tablist" onContextMenu={(e) => e.preventDefault()}>
        {TABS.map((t) => (
          <button
            key={t.id}
            role="tab"
            type="button"
            aria-selected={tab === t.id}
            aria-disabled={t.disabled || undefined}
            disabled={t.disabled}
            title={t.label}
            className={'tab' + (tab === t.id ? ' active' : '') + (t.disabled ? ' disabled' : '')}
            onClick={() => { if (!t.disabled) setTab(t.id) }}
          >
            <img
              className="ticon"
              src="/sprite?kind=media&id=0"
              alt={t.label}
              draggable={false}
              style={{ objectPosition: `-${ICON_X(t.cell)}px 0` }}
            />
            <span className="tlabel">{t.label}</span>
          </button>
        ))}
      </div>
      <div className="tabbody">
        {tab === 'inv'     && <InventoryGrid inventory={state?.inventory ?? []} />}
        {tab === 'equip'   && <EquipmentPanel equipment={state?.equipment ?? []} />}
        {tab === 'stats'   && <StatsPanel self={state?.self ?? null} />}
        {tab === 'magic'   && (
          state?.magic
            ? <Spellbook magic={state.magic} />
            : <div className="spellhdr">Loading…</div>
        )}
        {tab === 'prayer'  && (
          <PrayerTab
            prayers={state?.prayers ?? []}
            prayerPts={state?.self?.prayer ?? 0}
            maxPrayer={state?.self?.maxPrayer ?? 1}
          />
        )}
        {tab === 'friends' && <FriendsTab social={state?.social ?? { friends: [], ignores: [] }} />}
      </div>
    </div>
  )
}
