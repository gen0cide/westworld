// Top-level app: owns camera state + the animation/rotation timers, wires the
// keyboard controls (rotate/zoom/resize/screenshot/clip), and lays out the
// viewport, chat, and side panel. Mirrors the original single-file client's
// control scheme and three-loop structure.

import { useEffect, useRef, useState } from 'react'
import { clip, shot } from './api'
import { useConfig, useGameState } from './state'
import { useUI } from './ui'
import { Viewport, type Camera4 } from './components/Viewport'
import { Hud } from './components/Hud'
import { Minimap } from './components/Minimap'
import { ChatPanel } from './components/ChatPanel'
import { SidePanel } from './components/SidePanel'
import { BankWindow } from './components/BankWindow'
import { ShopWindow } from './components/ShopWindow'
import { TradeWindow } from './components/TradeWindow'
import { DuelWindow } from './components/DuelWindow'
import { NpcDialog } from './components/NpcDialog'

const ROT = 4
const DEFAULTS: Camera4 = { rot: 64, zoom: 1000, w: 512, h: 334 }

export function App() {
  const cfg = useConfig()
  const ui = useUI()
  const { state, chat } = useGameState(400)

  const [camera, setCamera] = useState<Camera4>(DEFAULTS)
  const animRef = useRef(0)
  const keys = useRef<Record<string, boolean>>({})
  const initedFromCfg = useRef(false)
  const cameraRef = useRef(camera)
  cameraRef.current = camera

  // One-time camera seed from /config.
  useEffect(() => {
    if (cfg && !initedFromCfg.current) {
      initedFromCfg.current = true
      setCamera((c) => ({ ...c, zoom: cfg.zoom, w: cfg.w, h: cfg.h, rot: cfg.rotation & 0xff }))
    }
  }, [cfg])

  // anim counter (drives model-swap frames in /frame).
  useEffect(() => {
    const id = window.setInterval(() => { animRef.current = (animRef.current + 1) & 0x3fffffff }, 160)
    return () => window.clearInterval(id)
  }, [])

  // held-arrow rotation.
  useEffect(() => {
    const id = window.setInterval(() => {
      if (keys.current['ArrowLeft']) setCamera((c) => ({ ...c, rot: (c.rot - ROT + 256) & 255 }))
      if (keys.current['ArrowRight']) setCamera((c) => ({ ...c, rot: (c.rot + ROT) & 255 }))
    }, 33)
    return () => window.clearInterval(id)
  }, [])

  // keyboard: zoom/resize/screenshot/clip, suppressed while typing in chat.
  useEffect(() => {
    const HANDLED = ['ArrowLeft', 'ArrowRight', '+', '=', '-', '_', '[', ']', 'p', 'c']
    const down = (e: KeyboardEvent) => {
      if (document.activeElement instanceof HTMLInputElement) return
      if (!HANDLED.includes(e.key)) return
      e.preventDefault()
      keys.current[e.key] = true
      if (e.key === 'p') { ui.flash('shot…'); void shot(cameraRef.current).then(() => ui.flash('shot saved')); return }
      if (e.key === 'c') { ui.flash('clip… walk now (~3s)'); void clip(cameraRef.current).then(() => ui.flash('clip saved')); return }
      setCamera((c) => {
        if (e.key === '+' || e.key === '=') return { ...c, zoom: Math.max(c.zoom - 150, 250) }
        if (e.key === '-' || e.key === '_') return { ...c, zoom: Math.min(c.zoom + 150, 4000) }
        if (e.key === '[') return { ...c, w: Math.max(c.w - 128, 256), h: Math.max(c.h - 84, 168) }
        if (e.key === ']') return { ...c, w: Math.min(c.w + 128, 1280), h: Math.min(c.h + 84, 840) }
        return c
      })
    }
    const up = (e: KeyboardEvent) => { keys.current[e.key] = false }
    document.addEventListener('keydown', down)
    document.addEventListener('keyup', up)
    return () => { document.removeEventListener('keydown', down); document.removeEventListener('keyup', up) }
  }, [ui])

  // global menu-close triggers (click anywhere, scroll, Escape).
  useEffect(() => {
    const close = () => ui.closeMenu()
    const esc = (e: KeyboardEvent) => { if (e.key === 'Escape') ui.closeMenu() }
    document.addEventListener('click', close, true)
    document.addEventListener('scroll', close, true)
    document.addEventListener('keydown', esc)
    return () => {
      document.removeEventListener('click', close, true)
      document.removeEventListener('scroll', close, true)
      document.removeEventListener('keydown', esc)
    }
  }, [ui])

  return (
    <div id="app">
      <Viewport
        camera={camera}
        animRef={animRef}
        hud={
          <>
            <Hud self={state?.self ?? null} camera={camera} />
            <Minimap
              entities={state?.entities}
              rot={camera.rot}
              self={state?.self ? { x: state.self.x, y: state.self.y } : undefined}
            />
          </>
        }
      />
      <ChatPanel chat={chat} />
      <SidePanel state={state} />
      {state?.bank?.open && (
        <BankWindow bank={state.bank} inventory={state.inventory} />
      )}
      {state?.shop?.open && (
        <ShopWindow shop={state.shop} inventory={state.inventory} />
      )}
      {state?.trade && (state.trade.phase === 'open' || state.trade.phase === 'confirm') && (
        <TradeWindow trade={state.trade} inventory={state.inventory} />
      )}
      {state?.duel && (state.duel.phase === 'open' || state.duel.phase === 'confirm') && (
        <DuelWindow duel={state.duel} inventory={state.inventory} />
      )}
      {state?.dialog?.open && <NpcDialog dialog={state.dialog} />}
    </div>
  )
}
