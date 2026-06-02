// Tiny drag context for F4 use-item-on-target (spec §2.5). Holds the in-flight
// drag source (an InvItem) in a ref so the Viewport — a different subtree from
// the inventory side panel — can read it on drop. Kept separate from useUI so
// neither concern leaks into the other.

import { createContext, useContext, useRef, type ReactNode } from 'react'
import type { InvItem } from './types'

interface DragCtx {
  begin: (it: InvItem) => void
  take: () => InvItem | null // returns + clears the current source
}

const Ctx = createContext<DragCtx | null>(null)

export function useDrag(): DragCtx {
  const c = useContext(Ctx)
  if (!c) throw new Error('useDrag must be used within <DragProvider>')
  return c
}

export function DragProvider({ children }: { children: ReactNode }) {
  const src = useRef<InvItem | null>(null)
  const api: DragCtx = {
    begin: (it) => { src.current = it },
    take: () => { const s = src.current; src.current = null; return s },
  }
  return <Ctx.Provider value={api}>{children}</Ctx.Provider>
}
