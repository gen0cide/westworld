// Spellbook tab — renders the magic spellbook inside the SidePanel.
// Static catalog is fetched once from GET /spells (cached in api.ts module state).
// Per-tick runtime flags (canCast, hasRunes) come from /state.magic.

import { useEffect, useState } from 'react'
import { castSpell, getSpells } from '../api'
import { useUI } from '../ui'
import type { MenuSection } from '../ui'
import type { MagicState, SpellDef } from '../types'

// Short rune-name lookup: rune item id → short display label.
const RUNE_NAMES: Record<number, string> = {
  31: 'Fire', 32: 'Water', 33: 'Air', 34: 'Earth',
  35: 'Mind', 36: 'Chaos', 38: 'Death', 39: 'Blood',
  40: 'Nature', 41: 'Law', 42: 'Cosmic', 43: 'Body', 44: 'Soul',
  619: 'Fire(∞)', 620: 'Water(∞)', 621: 'Air(∞)', 622: 'Earth(∞)',
}

export function Spellbook({ magic }: { magic: MagicState }) {
  const ui = useUI()
  const [catalog, setCatalog] = useState<SpellDef[]>([])

  useEffect(() => {
    void getSpells().then(setCatalog)
  }, [])

  // Build a flag lookup by spell id.
  const flagMap = new Map(magic.spells.map((f) => [f.id, f]))

  const cast = (spellId: number) => {
    void castSpell(spellId).then((r) => ui.flash(r.message ?? (r.ok ? 'ok' : 'failed')))
  }

  const spellMenu = (e: React.MouseEvent, sp: SpellDef) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: sp.name,
      items: [
        { text: 'Cast on self', run: () => cast(sp.id) },
        ...(sp.type === 2 || sp.type === 3
          ? [{ text: 'Cast on NPC…', run: () => ui.flash('select NPC from map') }]
          : []),
        { text: 'Examine', run: () => ui.flash(sp.description) },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  return (
    <div className="spellbook">
      <div className="spellhdr">
        Magic — level {magic.level}/{magic.maxLevel}
      </div>
      {catalog.map((sp) => {
        const flag = flagMap.get(sp.id)
        const canCast  = flag?.canCast  ?? false
        const hasRunes = flag?.hasRunes ?? false
        const colorClass = !canCast ? 'sp-low' : hasRunes ? 'sp-ready' : 'sp-norune'
        const runeText = sp.runes
          .map((r) => `${r.count}×${RUNE_NAMES[r.itemId] ?? `#${r.itemId}`}`)
          .join(' ')
        return (
          <div
            key={sp.id}
            className={`spellrow ${colorClass}`}
            title={sp.description}
            onClick={() => canCast && cast(sp.id)}
            onContextMenu={(e) => spellMenu(e, sp)}
          >
            <span className="sp-lvl">Lv{sp.reqLevel}</span>
            <span className="sp-name">{sp.name}</span>
            {runeText && <span className="sp-runes">{runeText}</span>}
          </div>
        )
      })}
    </div>
  )
}
