import { chooseDialog } from '../api'
import { useUI } from '../ui'
import type { Dialog } from '../types'

/**
 * NpcDialog renders the authentic RSC NPC option menu: an optional speech line
 * followed by a left-click list of choices, anchored over the lower viewport
 * (not a modal — it does not block the world). Clicking an option posts to
 * POST /dialog; the server reply (or the handler clearing the menu) makes the
 * next /state poll drop `state.dialog`, which unmounts this component.
 */
export function NpcDialog({ dialog }: { dialog: Dialog }) {
  const ui = useUI()
  const pick = (i: number) => {
    void chooseDialog(i).then((r) => {
      if (r.message) ui.flash(r.message)
    })
  }
  return (
    <div className="dialogwin" role="menu">
      {dialog.npcText && <div className="dialogtext">{dialog.npcText}</div>}
      <ul className="dialogopts">
        {dialog.options.map((opt, i) => (
          <li key={i} className="dialogopt" role="menuitem" onClick={() => pick(i)}>
            {opt}
          </li>
        ))}
      </ul>
    </div>
  )
}
