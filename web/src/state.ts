// Polling hooks for the cradle API: client config (once) and game state (loop).

import { useEffect, useRef, useState } from 'react'
import { getConfig, getState } from './api'
import type { ChatEntry, ClientConfig, GameState } from './types'

/** Fetch GET /config once at boot. */
export function useConfig(): ClientConfig | null {
  const [cfg, setCfg] = useState<ClientConfig | null>(null)
  useEffect(() => {
    let alive = true
    getConfig().then((c) => { if (alive) setCfg(c) }).catch(() => {})
    return () => { alive = false }
  }, [])
  return cfg
}

const MAX_CHAT = 200

/** Poll GET /state on an interval. Chat is the bounded server ring (last ~32);
 *  we de-dupe by seq and accumulate up to MAX_CHAT lines client-side. */
export function useGameState(intervalMs = 400): {
  state: GameState | null
  chat: ChatEntry[]
} {
  const [state, setState] = useState<GameState | null>(null)
  const [chat, setChat] = useState<ChatEntry[]>([])
  const lastSeq = useRef(-1)
  const busy = useRef(false)

  useEffect(() => {
    let alive = true
    const tick = async () => {
      if (busy.current) return
      busy.current = true
      try {
        const s = await getState()
        if (!alive) return
        setState(s)
        const fresh = (s.chat ?? []).filter((c) => c.seq > lastSeq.current)
        if (fresh.length) {
          lastSeq.current = fresh[fresh.length - 1].seq
          setChat((prev) => {
            const next = [...prev, ...fresh]
            return next.length > MAX_CHAT ? next.slice(next.length - MAX_CHAT) : next
          })
        }
      } catch { /* transient; next tick retries */ } finally {
        busy.current = false
      }
    }
    void tick()
    const id = window.setInterval(tick, intervalMs)
    return () => { alive = false; window.clearInterval(id) }
  }, [intervalMs])

  return { state, chat }
}
