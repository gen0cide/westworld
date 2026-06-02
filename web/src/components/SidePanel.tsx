// Right-hand panel with an authentic-ish RSC tab strip (one row of compact
// icon-ish buttons; 100-rsc-ui-map §2). The real client uses spriteMedia tab
// icons — those aren't served to the browser (sprites.go only serves item
// icons), so we approximate with palette-styled buttons. Tab state is local UI
// state, mirroring the original client's `qc`.

import { useState } from 'react'
import type { GameState } from '../types'
import { EquipmentPanel } from './EquipmentPanel'
import { InventoryGrid } from './InventoryGrid'
import { StatsPanel } from './StatsPanel'
import { Spellbook } from './Spellbook'
import { PrayerTab } from './PrayerTab'

type Tab = 'inv' | 'equip' | 'stats' | 'magic' | 'prayer' | 'friends'

interface TabDef {
  id: Tab
  glyph: string // compact icon-ish stand-in for the spriteMedia tab icon
  label: string
  disabled?: boolean
}

// Ordered left→right. Order matches RSC qc.
const TABS: TabDef[] = [
  { id: 'inv',     glyph: '\u{1F392}', label: 'Inv' },     // 🎒
  { id: 'equip',   glyph: '\u{1F6E1}', label: 'Worn' },    // 🛡
  { id: 'stats',   glyph: '\u{1F4CA}', label: 'Stats' },   // 📊
  { id: 'magic',   glyph: '✨',    label: 'Magic' },   // ✨
  { id: 'prayer',  glyph: '\u{1F64F}', label: 'Pray' },    // 🙏
  { id: 'friends', glyph: '\u{1F465}', label: 'Friends', disabled: true }, // 👥
]

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
            <span className="tglyph" aria-hidden="true">{t.glyph}</span>
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
        {tab === 'friends' && <div className="stub-body">Friends — coming soon</div>}
      </div>
    </div>
  )
}
