// Friends / Ignore tab. Add/remove friends (host.AddFriend/RemoveFriend),
// PM a friend (reuses /chat kind:pm), add/remove ignores. Online status comes
// from state.social.friends[].online, which drives the status dot + row tooltip.
// Received PMs already stream into <ChatPanel> (kind:private), so the PM
// composer here is send-only.

import { useState } from 'react'
import { socialAction, sendChat } from '../api'
import { useUI, type MenuSection } from '../ui'
import type { Social } from '../types'

const IGNORE_SUPPORTED = true

export function FriendsTab({ social }: { social: Social }) {
  const ui = useUI()
  const [addName, setAddName] = useState('')
  const [ignoreName, setIgnoreName] = useState('')
  const [pmTarget, setPmTarget] = useState('')
  const [pmText, setPmText] = useState('')

  const run = (p: Promise<{ ok: boolean; message?: string }>) =>
    void p.then((r) => { if (r.message) ui.flash(r.message) })

  const addFriend = () => {
    const n = addName.trim()
    if (!n) return
    run(socialAction('add_friend', n)); setAddName('')
  }

  const addIgnore = () => {
    const n = ignoreName.trim()
    if (!n) return
    run(socialAction('add_ignore', n)); setIgnoreName('')
  }

  const rowMenu = (e: React.MouseEvent, name: string) => {
    e.preventDefault()
    const sections: MenuSection[] = [{
      header: name,
      items: [
        { text: 'Message', run: () => setPmTarget(name) },
        { text: 'Remove friend', run: () => run(socialAction('remove_friend', name)) },
      ],
    }]
    ui.openActions(e.clientX, e.clientY, sections)
  }

  const sendPm = () => {
    const t = pmTarget.trim(), m = pmText.trim()
    if (!t || !m) { ui.flash('pick a friend + type a message'); return }
    run(sendChat('pm', m, t)); setPmText('')
  }

  return (
    <div className="friends">
      <div className="head">Friends</div>
      <div className="friendlist">
        {social.friends.length === 0 && <div className="empty">no friends yet</div>}
        {social.friends.map((f) => (
          <div
            key={f.name}
            className={'frow' + (f.name === pmTarget ? ' sel' : '')}
            onClick={() => setPmTarget(f.name)}
            onContextMenu={(e) => rowMenu(e, f.name)}
            title={f.online ? `online${f.world ? ' (' + f.world + ')' : ''}` : 'offline'}
          >
            <span className={'dot ' + (f.online ? 'on' : 'off')} />
            <span className="nm">{f.name}</span>
          </div>
        ))}
      </div>

      <div className="addrow">
        <input
          value={addName}
          placeholder="add friend…"
          onChange={(e) => setAddName(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') addFriend() }}
        />
        <button onClick={addFriend}>Add</button>
      </div>

      <div className="head">Message {pmTarget && <span className="nm">→ {pmTarget}</span>}</div>
      <div className="addrow">
        <input
          value={pmText}
          placeholder={pmTarget ? `message ${pmTarget}…` : 'select a friend first'}
          onChange={(e) => setPmText(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') sendPm() }}
        />
        <button onClick={sendPm} disabled={!pmTarget}>Send</button>
      </div>

      <div className="head">Ignore</div>
      <div className="friendlist">
        {social.ignores.length === 0 && <div className="empty">no one ignored</div>}
        {social.ignores.map((n) => (
          <div key={n} className="frow">
            <span className="nm">{n}</span>
            <button
              className="mini"
              disabled={!IGNORE_SUPPORTED}
              title="unignore"
              onClick={() => run(socialAction('remove_ignore', n))}
            >Unignore</button>
          </div>
        ))}
      </div>
      <div className="addrow">
        <input
          value={ignoreName}
          disabled={!IGNORE_SUPPORTED}
          placeholder="add ignore…"
          onChange={(e) => setIgnoreName(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') addIgnore() }}
        />
        <button disabled={!IGNORE_SUPPORTED} onClick={addIgnore}>Ignore</button>
      </div>
    </div>
  )
}
