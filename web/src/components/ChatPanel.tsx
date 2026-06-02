// Chat log + input. Classifies raw input (::cmd | /cmd → command, @name → pm,
// else say) and POSTs /chat. Auto-pins to the newest line when already at the
// bottom. Mirrors the original sendChat()/appendChat().

import { useEffect, useRef } from 'react'
import { sendChat, type ChatKind } from '../api'
import { useUI } from '../ui'
import type { ChatEntry } from '../types'

const KIND_CLASS: Record<string, string> = {
  npc: 'npc', private: 'pm', system: 'sys', self: 'you',
}

export function ChatPanel({ chat }: { chat: ChatEntry[] }) {
  const logRef = useRef<HTMLDivElement | null>(null)
  const atBottom = useRef(true)
  const ui = useUI()

  useEffect(() => {
    const el = logRef.current
    if (el && atBottom.current) el.scrollTop = el.scrollHeight
  }, [chat])

  const onScroll = () => {
    const el = logRef.current
    if (!el) return
    atBottom.current = el.scrollTop + el.clientHeight >= el.scrollHeight - 4
  }

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Escape') { e.currentTarget.blur(); return }
    if (e.key !== 'Enter') return
    const raw = e.currentTarget.value.trim()
    e.currentTarget.value = ''
    if (!raw) return
    let kind: ChatKind = 'say', to = '', body = raw
    if (raw.startsWith('::') || raw.startsWith('/')) {
      kind = 'command'
      body = (raw.startsWith('::') ? raw.slice(2) : raw.slice(1)).trim()
      if (!body) { ui.flash('empty command'); return }
    } else if (raw.startsWith('@')) {
      kind = 'pm'
      const sp = raw.indexOf(' ')
      if (sp < 0) { ui.flash('usage: @name message'); return }
      to = raw.slice(1, sp).trim()
      body = raw.slice(sp + 1).trim()
      if (!to || !body) { ui.flash('usage: @name message'); return }
    }
    sendChat(kind, body, to)
      .then((r) => { if (!r.ok && r.message) ui.flash(r.message) })
      .catch((err) => ui.flash('chat: ' + String(err)))
  }

  return (
    <div id="chat">
      <div className="chatlog" ref={logRef} onScroll={onScroll}>
        {chat.map((line) => (
          <div key={line.seq} className={KIND_CLASS[line.kind] ?? ''}>
            {(line.who ? `[${line.who}] ` : '') + line.text}
          </div>
        ))}
      </div>
      <input
        className="chatinput"
        autoComplete="off"
        spellCheck={false}
        placeholder="say&#8230;  ::cmd  @name pm"
        onKeyDown={onKeyDown}
      />
    </div>
  )
}
