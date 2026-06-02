// Shared UI layer: a flash toast and a right-click context menu, exposed via a
// small context so any component can raise a menu or a status flash without
// prop-drilling.
//
// The menu is built from generic rows ({text, run}); openMenu() adapts /pick
// candidates (run = act on the server), and openActions() takes arbitrary
// client-side handlers (e.g. bank deposit/withdraw quantities). Both share one
// menu surface and the global close-triggers in App.

import {
  createContext, useCallback, useContext, useMemo, useRef, useState,
  type ReactNode,
} from 'react'
import { act as apiAct } from './api'
import type { Candidate } from './types'

export interface MenuItem { text: string; run: () => void }
export interface MenuSection { header?: string; items: MenuItem[] }
interface MenuState { x: number; y: number; sections: MenuSection[] }

interface UI {
  flash: (msg: string) => void
  openMenu: (x: number, y: number, candidates: Candidate[]) => void
  openActions: (x: number, y: number, sections: MenuSection[]) => void
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

  const closeMenu = useCallback(() => setMenu(null), [])

  const openActions = useCallback((x: number, y: number, sections: MenuSection[]) => {
    if (!sections.some((s) => s.items.length)) return
    setMenu({ x, y, sections })
  }, [])

  const openMenu = useCallback((x: number, y: number, candidates: Candidate[]) => {
    if (!candidates.length) return
    const multi = candidates.length > 1
    const sections: MenuSection[] = candidates.map((cand) => ({
      header: multi ? cand.label : undefined,
      items: cand.options.map((opt) => ({
        text: multi ? opt.verb : `${opt.verb} ${cand.label}`,
        run: () => { void apiAct(cand.ref, opt.id).then((r) => { if (r.message) flash(r.message) }) },
      })),
    }))
    setMenu({ x, y, sections })
  }, [flash])

  const api = useMemo<UI>(
    () => ({ flash, openMenu, openActions, closeMenu }),
    [flash, openMenu, openActions, closeMenu],
  )

  return (
    <UICtx.Provider value={api}>
      {children}
      <div className={'flash' + (showFlash ? ' show' : '')}>{flashMsg}</div>
      {menu && <ContextMenu menu={menu} onClose={closeMenu} />}
    </UICtx.Provider>
  )
}

function ContextMenu({ menu, onClose }: { menu: MenuState; onClose: () => void }) {
  const rows = menu.sections.reduce((n, s) => n + s.items.length + (s.header ? 1 : 0), 0)
  const left = Math.min(menu.x, window.innerWidth - 160)
  const top = Math.min(menu.y, window.innerHeight - 24 * (rows + 1))
  return (
    <div className="menu" style={{ left, top }} onContextMenu={(e) => e.preventDefault()}>
      {menu.sections.map((sec, si) => (
        <div key={si}>
          {sec.header && <div className="hdr">{sec.header}</div>}
          {sec.items.map((it, ii) => (
            <div
              key={ii}
              className="opt"
              onClick={(e) => { e.stopPropagation(); onClose(); it.run() }}
            >
              {it.text}
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}
