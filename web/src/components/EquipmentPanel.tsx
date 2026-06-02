// Equipment summary: one row per worn slot (sprite > 0), showing the real worn
// item icon when the layer resolved to a wielded item (joined by worn appearance
// id in the /state builder), else a text fallback. Mirrors renderEquipment().

import type { EquipItem } from '../types'
import { ItemSprite } from './ItemSprite'

export function EquipmentPanel({ equipment }: { equipment: EquipItem[] }) {
  if (!equipment.length) {
    return (
      <div className="equip">
        <div className="row"><span className="slot">no equipment</span></div>
      </div>
    )
  }
  return (
    <div className="equip">
      {equipment.map((e, i) => (
        <div className="row" key={i}>
          <span className="slot">{e.slot}</span>
          <span className="equip-item">
            {e.itemId ? (
              <ItemSprite id={e.itemId} name={e.name} />
            ) : (
              <span className="stub">spr {e.sprite}</span>
            )}
            {e.name && <span className="equip-name">{e.name}</span>}
          </span>
        </div>
      ))}
    </div>
  )
}
