// Equipment summary: one row per worn slot (sprite > 0). Mirrors
// renderEquipment().

import type { EquipItem } from '../types'

export function EquipmentPanel({ equipment }: { equipment: EquipItem[] }) {
  if (!equipment.length) {
    return <div className="equip"><div className="row"><span className="slot">no equipment</span></div></div>
  }
  return (
    <div className="equip">
      {equipment.map((e, i) => (
        <div className="row" key={i}>
          <span className="slot">{e.slot}</span>
          <span>{e.itemId ? `item ${e.itemId}` : `spr ${e.sprite}`}</span>
        </div>
      ))}
    </div>
  )
}
