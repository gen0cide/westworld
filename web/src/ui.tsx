// Shared UI layer: a flash toast and a right-click context menu, exposed via a
// small context so any component (viewport, inventory, …) can raise a menu or a
// status flash without prop-drilling. Mirrors the original client's openMenu /
// flash helpers.

import {
  createContext, useCallback, useContext, useMemo, useRef, useState,
  type ReactNode,
} from 'react'
import { act as apiAct } from './api'
import type { Candidate } from './types'

interface MenuState { x: number; y: number; candidates: Candidate[] }

interface UI {
  flash: (msg: string) => void
  openMenu: (x: number, y: number, candidates: Candidate[]) => void
  closeMenu: () => void
}

const UICtx = createContext<UI | null>(null)

export function useUI(): UI {
  const ctx = useContext(UICtx)
  if (!ctx) throw new Error('useUI must be used within <UIProvider>')
  return ctx
}

export function UIProvider({ children }: { children: ReactNode }) {
  const [flashMsg, setFlashMsg] = useState('')
  const [showFlash, setShowFlash] = useState(false)
  const [menu, setMenu] = useState<MenuState | null>(null)
  const flashTimer = useRef<number | undefined>(undefined)

  const flash = useCallback((msg: string) => {
    setFlashMsg(msg)
    setShowFlash(true)
    window.clearTimeout(flashTimer.current)
    flashTimer.current = window.setTimeout(() => setShowFlash(false), 1500)
  }, [])

  const openMenu = useCallback((x: number, y: number, candidates: Candidate[]) => {
    if (!candidates.length) return
    setMenu({ x, y, candidates })
  }, [])
  const closeMenu = useCallback(() => setMenu(null), [])

  const api = useMemo<UI>(() => ({ flash, openMenu, closeMenu }), [flash, openMenu, closeMenu])

  return (
    <UICtx.Provider value={api}>
      {children}
      <div className={'flash' + (showFlash ? ' show' : '')}>{flashMsg}</div>
      {menu && (
        <ContextMenu
          menu={menu}
          onClose={closeMenu}
          onFlash={flash}
        />
      )}
    </UICtx.Provider>
  )
}

function ContextMenu(
  { menu, onClose, onFlash }: {
    menu: MenuState
    onClose: () => void
    onFlash: (m: string) => void
  },
) {
  const multi = menu.candidates.length > 1
  // Clamp roughly to the viewport; the menu is small so an estimate is fine.
  const left = Math.min(menu.x, window.innerWidth - 160)
  const top = Math.min(menu.y, window.innerHeight - 24 * (menu.candidates.length + 1))

  const run = async (cand: Candidate, optionId: number) => {
    onClose()
    const r = await apiAct(cand.ref, optionId)
    if (r.message) onFlash(r.message)
  }

  return (
    <div
      className="menu"
      style={{ left, top }}
      onContextMenu={(e) => e.preventDefault()}
    >
      {menu.candidates.map((cand, ci) => (
        <div key={ci}>
          {multi && <div className="hdr">{cand.label}</div>}
          {cand.options.map((opt) => (
            <div
              key={opt.id}
              className="opt"
              onClick={(e) => { e.stopPropagation(); void run(cand, opt.id) }}
            >
              {multi ? opt.verb : `${opt.verb} ${cand.label}`}
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
