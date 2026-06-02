// HUD overlay: host coords, combat/hp, camera state, control legend. Mirrors
// the original drawHud() text block.

import type { Self } from '../types'
import type { Camera4 } from './Viewport'

export function Hud({ self, camera }: { self: Self | null; camera: Camera4 }) {
  const x = self?.x ?? 0, y = self?.y ?? 0, plane = self?.plane ?? 0
  const cl = self?.combatLevel ?? 0, hp = self?.hp ?? 0, maxHp = self?.maxHp ?? 0
  return (
    <div className="hud">
      {`host (${x}, ${y})  plane ${plane}\n`}
      {`combat ${cl}  hp ${hp}/${maxHp}\n`}
      {`rot ${camera.rot}  zoom ${camera.zoom}  ${camera.w}x${camera.h}\n`}
      {'left-click act \xB7 right-click menu \xB7 < >  rotate \xB7 +− zoom'}
    </div>
  )
}
