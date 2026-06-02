# Friends / Ignore (F2) — Implementation Spec

Template cloned: **Bank** (commit `ad9633c`). This is the one backlog feature
(`90-backlog.md` §0 / §4) where the read-side does **not** fully exist yet.

> **TL;DR of the gap:** outbound *add/remove friend* + *PM* already exist
> (`action/pm.go` opcodes 195/167/218, `runtime/social.go`). Everything else is
> MISSING: (a) no outbound **add/remove ignore** opcodes/Host methods, (b) no
> **inbound decode** of the friend-update (149) / ignore-list (109) packets, so
> (c) there is **no `world.SocialState` roster mirror** to surface in `/state`.
> The Go-backend-only part of this feature (UI + a *cosmetic* friends/ignore
> panel that only echoes what the user typed) IS buildable today; the
> **authentic online/offline roster + the ignore action** are NOT — they need
> new `world/`, `event/`, `proto/`, `action/`, `runtime/` work, which this spec
> documents precisely but does NOT implement (those dirs are off-limits to the
> implementer of this spec).

This spec is split into two clearly-labelled tracks:

- **TRACK A (buildable now, web + cmd/cradle only):** Friends/Ignore tab UI,
  add/remove/PM wiring through the EXISTING `host.AddFriend` / `host.RemoveFriend`
  / `host.PrivateMessage`. The friend roster shown is the local optimistic list
  the user built this session (names they added) — there is no server "online"
  truth yet, so each entry renders status "unknown".
- **TRACK B (BLOCKED — backend gap, do NOT build under the spec's constraints):**
  the authentic online/offline roster (requires decoding inbound 149/109) and
  the ignore add/remove actions (requires outbound 132/241 + Host methods).
  Section 7 specifies EXACTLY what is missing with `file:line` so a backend owner
  can unblock it; once they do, Track A's `status: 'unknown'` cells light up with
  zero web changes beyond reading the new `/state.social` fields.

---

## 0. Backend audit (what exists vs. what is missing)

### Exists (outbound + Host, verified)

| Capability | Host method | file:line | action fn | outbound opcode |
|---|---|---|---|---|
| Add friend | `(*Host).AddFriend(ctx, name string) error` | `runtime/social.go:18` | `action.AddFriend` `action/pm.go:57` | 195 `SOCIAL_ADD_FRIEND` |
| Remove friend | `(*Host).RemoveFriend(ctx, name string) error` | `runtime/social.go:23` | `action.RemoveFriend` `action/pm.go:67` | 167 `SOCIAL_REMOVE_FRIEND` |
| Private message | `(*Host).PrivateMessage(ctx, recipient, message string) error` | `runtime/social.go:12` | `action.PrivateMessage` `action/pm.go:37` | 218 `SOCIAL_SEND_PRIVATE_MESSAGE` |

PM is **already plumbed end-to-end through the remote client** via `POST /chat`
with `kind:"pm"` (`cmd/cradle/remoteclient.go:1083`, `host.PrivateMessage`). The
web `sendChat('pm', body, to)` path is `web/src/api.ts:76`. **Do not add a new PM
endpoint** — Track A's "PM" button reuses `sendChat('pm', …)`.

Incoming PMs are already decoded (opcode 120, `proto/v235/inbound.go:293` →
`event.PrivateMessage` `event/events.go:47`) and mirrored as the latest PM
(`world.RecentEvents.SetPM` `world/recent.go:133`); they already appear in the
chat ring as `kind:"private"` lines (`cmd/cradle/remoteclient.go` chat ring via
`ring.ingestRecent`). So **received PMs already render in `<ChatPanel>`** today.

### MISSING (the real protocol gap)

1. **No outbound add/remove ignore.** `action/pm.go:19-23` declares only
   `outPrivateMessage 218 / outAddFriend 195 / outRemoveFriend 167`. The reference
   client sends **132 `SOCIAL_ADD_IGNORE`** (`reference/.../Mudclient.java:4818`,
   `4854`) and **241 `SOCIAL_REMOVE_IGNORE`** (`Mudclient.java:4862`, `4883`).
   There is no `action.AddIgnore` / `action.RemoveIgnore` and no
   `(*Host).AddIgnore` / `(*Host).RemoveIgnore` (grep: zero hits in `action/`,
   `runtime/`).

2. **No inbound roster decode.** `proto/v235/inbound_opcodes.go` defines NO
   friend/ignore opcodes, and `proto/v235/inbound.go`'s `DecodeInbound` switch
   (`inbound.go:24`) has no case for them. The reference handles:
   - **149 `SEND_FRIEND_UPDATE`** — one packet per friend; carries `name`,
     `formerName`, a flags byte (bit0 = match-by-former/rename, bit2 = online),
     and (if online) `onlineWorld` string. Sent as a burst at login and on every
     status change. (`reference/.../Mudclient.java:6794-6840`).
   - **109 `SEND_IGNORE_LIST`** — bulk: `count` byte, then `count` × 4 strings
     (current, name2, world/title, formerName). (`Mudclient.java:6874-6881`).
   - **237 `SEND_IGNORE_LIST_RENAME`** — incremental ignore rename/append.
     (`Mudclient.java:6842-6872`). (Optional for v1.)

3. **No `world` mirror.** There is no `world.SocialState` / friends roster
   record. `world/world.go:13` `World` struct has `Recent`, `Bank`, etc., but no
   `Social`. The DSL `is_friend` view is hard-stubbed to `false` precisely
   because of this — see the comment at `runtime/views_world.go:363-369`:
   > *"Stubs until the host tracks friends list… is_friend will derive from a
   > friends-list mirror once we decode the friend-list packets (currently we
   > send AddFriend outbound but don't mirror server-side state)."*

**Conclusion:** Track A ships with the methods above. Track B's roster +
ignore is the only backlog item that is genuinely backend-blocked
(`90-backlog.md` §0: *"The only true missing capability in the whole backlog is
the friends/ignore roster read-side… a `world/` + `event/` decode task"*).

---

## TRACK A — buildable now (web/ + cmd/cradle/ only)

### A1. `cmd/cradle/remoteclient.go` — Go struct + handler additions

#### A1.1 `/state` types (add near the bank types, after `stateBankSlot` at line 319)

Track A has no server-side roster yet, so `/state.social` reflects an
**optimistic local roster** maintained in the handler closure (names the user
added/removed this session). This is intentionally a thin echo until Track B
lands; the JSON shape is forward-compatible so Track B only fills in `online`
and `world`.

```go
// stateSocial mirrors the (future) world.SocialState roster for the SPA. Until
// the inbound 149/109 packets are decoded (specs/friends.md §7), Online is
// always false and World is "" — the roster is the optimistic local list the
// user built this session via /social. The SPA renders <FriendsTab> from this.
type stateSocial struct {
	Friends []stateFriend `json:"friends"`
	Ignores []string      `json:"ignores"`
}

type stateFriend struct {
	Name   string `json:"name"`
	Online bool   `json:"online"` // always false in Track A (no roster mirror yet)
	World  string `json:"world"`  // "" in Track A
}
```

Add one field to `stateResponse` (after `Bank *stateBank` at line 304):

```go
	// Social is always present (never nil) — unlike bank/shop it is not a
	// window-lifecycle object; the Friends tab is always available. In Track A
	// it is the optimistic local roster; Track B fills it from world.SocialState.
	Social stateSocial `json:"social"`
```

#### A1.2 Optimistic local roster (declare inside `serveClient`, beside `ring := &chatRing{}` at line 510)

```go
	// social is the optimistic local friend/ignore roster surfaced at
	// /state.social. Track A has no server roster mirror (specs/friends.md §7),
	// so we track what the user added/removed this session under socialMu.
	// Track B replaces reads here with host.World().Social.* once it exists.
	var (
		socialMu      sync.Mutex
		localFriends  = map[string]bool{} // key: normalized name -> present
		localIgnores  = map[string]bool{}
		friendOrder   []string            // insertion order, display-stable
		ignoreOrder   []string
	)
	normName := func(s string) string { return strings.ToLower(strings.TrimSpace(s)) }
	addLocal := func(set map[string]bool, order *[]string, name string) {
		k := normName(name)
		if k == "" || set[k] {
			return
		}
		set[k] = true
		*order = append(*order, name) // keep original-case display name
	}
	delLocal := func(set map[string]bool, order *[]string, name string) {
		k := normName(name)
		if !set[k] {
			return
		}
		delete(set, k)
		out := (*order)[:0]
		for _, n := range *order {
			if normName(n) != k {
				out = append(out, n)
			}
		}
		*order = out
	}
```

> `sync` and `strings` are already imported in this file (used by the action
> worker and `/chat`); no new imports.

#### A1.3 Populate `/state.social` (in the `/state` handler, build alongside `bankBlock` ~line 962)

```go
	// social block — optimistic local roster (Track A). Always present.
	socialMu.Lock()
	friends := make([]stateFriend, 0, len(friendOrder))
	for _, n := range friendOrder {
		friends = append(friends, stateFriend{Name: n, Online: false, World: ""})
	}
	ignores := append([]string(nil), ignoreOrder...)
	socialMu.Unlock()
	socialBlock := stateSocial{Friends: friends, Ignores: ignores}
```

Then add `Social: socialBlock,` to the `stateResponse{…}` literal at line 983.

#### A1.4 `POST /social` handler (new route; register beside `/bank` ~line 995)

```go
// socialRequest is the body of POST /social. op selects the action; Name is the
// player username for all ops. add_ignore/remove_ignore are accepted by the
// HTTP layer but currently return ok:false "not yet supported" because the
// outbound ignore opcodes + Host methods do not exist (specs/friends.md §7);
// the SPA disables those buttons. add_friend/remove_friend funnel through the
// existing host.AddFriend/RemoveFriend and update the optimistic local roster.
type socialRequest struct {
	Op   string `json:"op"`   // add_friend | remove_friend | add_ignore | remove_ignore
	Name string `json:"name"`
}
```

```go
	// POST /social — add/remove friend (and, once Track B lands, ignore).
	// Routed through the same serialized action worker as /act, /bank, /chat.
	mux.HandleFunc("/social", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req socialRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		name := strings.TrimSpace(req.Name)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		if name == "" {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: "name required"})
			return
		}
		switch req.Op {
		case "add_ignore", "remove_ignore":
			// BLOCKED: no outbound ignore opcode / Host method yet.
			// See specs/friends.md §7. SPA keeps these buttons disabled.
			_ = json.NewEncoder(w).Encode(actResponse{
				OK: false, Message: "ignore is not yet supported (backend gap)",
			})
			return
		case "add_friend", "remove_friend":
		default:
			_ = json.NewEncoder(w).Encode(actResponse{
				OK: false, Message: "unknown op: must be add_friend|remove_friend|add_ignore|remove_ignore",
			})
			return
		}
		op := req.Op
		msg, runErr := enqueueAction(func(wctx context.Context) (string, error) {
			switch op {
			case "add_friend":
				return "Add friend " + name, host.AddFriend(wctx, name)
			case "remove_friend":
				return "Remove friend " + name, host.RemoveFriend(wctx, name)
			}
			return "", fmt.Errorf("unknown social op %q", op)
		})
		if runErr != nil {
			_ = json.NewEncoder(w).Encode(actResponse{OK: false, Message: runErr.Error()})
			return
		}
		// Mirror into the optimistic local roster so /state reflects it before
		// any (future) server roster packet arrives.
		socialMu.Lock()
		switch op {
		case "add_friend":
			addLocal(localFriends, &friendOrder, name)
		case "remove_friend":
			delLocal(localFriends, &friendOrder, name)
		}
		socialMu.Unlock()
		_ = json.NewEncoder(w).Encode(actResponse{OK: true, Message: msg})
	})
```

> Reuses existing helpers: `enqueueAction` (`remoteclient.go:444`), `actResponse`
> (search the file — same struct `/bank` and `/chat` encode), `host.AddFriend` /
> `host.RemoveFriend` (`runtime/social.go:18,23`). No new imports.

### A2. Web types — `web/src/types.ts`

Add after `Bank` (line 103):

```ts
export interface Friend {
  name: string
  online: boolean // false in Track A until the roster decode lands
  world: string   // '' in Track A
}

export interface Social {
  friends: Friend[]
  ignores: string[]
}
```

Add to `GameState` (after `bank?: Bank` at line 110):

```ts
  social: Social // always present; friends/ignores rosters
```

### A3. Web API — `web/src/api.ts`

Add after `bankAction` (line 93):

```ts
export type SocialOp = 'add_friend' | 'remove_friend' | 'add_ignore' | 'remove_ignore'

/** POST /social — add/remove a friend or ignore by name. add_ignore /
 *  remove_ignore return ok:false until the backend ignore opcodes land
 *  (specs/friends.md §7); the UI disables those controls. */
export async function socialAction(op: SocialOp, name: string): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/social', { op, name })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}
```

(PM is NOT added here — Track A's PM button calls the existing
`sendChat('pm', text, to)` from `web/src/api.ts:76`.)

### A4. Component — `web/src/components/FriendsTab.tsx` (new)

Props: `{ social: Social }`. Lives inside `<SidePanel>` as a third tab (A5).
Interaction model mirrors the bank window's left-click=primary,
right-click=context-menu pattern (`BankWindow.tsx`) and uses the shared
`useUI()` flash + `openActions` (`web/src/ui.tsx:24,51`).

- **Friends list:** each row shows the name and a status dot. Online dot is
  green (`--rsc-green`), offline/unknown is gray (`--rsc-gray`). In Track A all
  are gray (`online:false`); the component reads `f.online` so it lights up for
  free once Track B fills it.
  - **Left-click a friend row** → start a PM: focus an inline composer prefilled
    with that name (or, simplest: call `ui.flash` + open the context menu). For
    v1 use the context menu for everything to match Bank's single-interaction
    surface.
  - **Right-click a friend row** → `ui.openActions` menu, header = name, items:
    - `Message` → set composer target = name, focus the PM input.
    - `Remove friend` → `socialAction('remove_friend', name)`.
- **Add-friend input:** a small text input + "Add" button → on Enter/click,
  `socialAction('add_friend', value)` then clear. Echoes server `message` via
  `ui.flash`.
- **PM composer:** target name (read-only chip showing the selected friend) + a
  message input. On Enter → `sendChat('pm', msg, target)`; show flash if
  `!ok`. Received PMs already stream into `<ChatPanel>` (kind `private`), so the
  composer is send-only.
- **Ignore list:** render `social.ignores` as plain name rows. Add-ignore input
  + per-row "Unignore" button are present but **disabled** (greyed) with a
  `title="ignore not yet supported (backend gap)"`; clicking calls
  `socialAction('add_ignore'|'remove_ignore', name)` which returns the
  backend-gap message — but keep them disabled so the UI is honest. When Track B
  lands, flip `disabled={!IGNORE_SUPPORTED}` to `false`.

Reference implementation skeleton (copy-pasteable):

```tsx
// Friends / Ignore tab. Add/remove friends (host.AddFriend/RemoveFriend),
// PM a friend (reuses /chat kind:pm). Online status comes from state.social
// (false everywhere until the roster decode lands — specs/friends.md §7).
// Ignore add/remove is rendered disabled until the backend gap is closed.

import { useState } from 'react'
import { socialAction, sendChat } from '../api'
import { useUI, type MenuSection } from '../ui'
import type { Social } from '../types'

const IGNORE_SUPPORTED = false // flip to true once outbound 132/241 + Host land

export function FriendsTab({ social }: { social: Social }) {
  const ui = useUI()
  const [addName, setAddName] = useState('')
  const [pmTarget, setPmTarget] = useState('')
  const [pmText, setPmText] = useState('')

  const run = (p: Promise<{ ok: boolean; message?: string }>) =>
    void p.then((r) => { if (r.message) ui.flash(r.message) })

  const addFriend = () => {
    const n = addName.trim()
    if (!n) return
    run(socialAction('add_friend', n)); setAddName('')
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
              title={IGNORE_SUPPORTED ? 'unignore' : 'ignore not yet supported (backend gap)'}
              onClick={() => run(socialAction('remove_ignore', n))}
            >Unignore</button>
          </div>
        ))}
      </div>
      <div className="addrow">
        <input disabled={!IGNORE_SUPPORTED} placeholder="add ignore… (backend gap)" />
        <button disabled title="ignore not yet supported (backend gap)">Ignore</button>
      </div>
    </div>
  )
}
```

### A5. `web/src/components/SidePanel.tsx` — add the Friends tab

Change the `Tab` union and add a third button + body branch:

```tsx
type Tab = 'inv' | 'stats' | 'friends'
```

In the `tabbar`, add after the Stats button (line 19):

```tsx
        <button className={tab === 'friends' ? 'active' : ''} onClick={() => setTab('friends')}>Friends</button>
```

In the `tabbody` ternary (lines 22-29), convert to handle three tabs, e.g.:

```tsx
        {tab === 'inv' && (<><EquipmentPanel equipment={state?.equipment ?? []} /><InventoryGrid inventory={state?.inventory ?? []} /></>)}
        {tab === 'stats' && <StatsPanel self={state?.self ?? null} />}
        {tab === 'friends' && <FriendsTab social={state?.social ?? { friends: [], ignores: [] }} />}
```

Add the import: `import { FriendsTab } from './FriendsTab'`.

> No `App.tsx` change is required — `FriendsTab` lives inside `<SidePanel>`,
> which `App.tsx:96` already renders. (Contrast with Bank, which is a modal in
> `App.tsx:97`; Friends is a side-panel tab, always available.)

### A6. Styles — `web/src/styles.css`

Append (after the context-menu block at line 225). Palette vars
(`--rsc-green`, `--rsc-gray`, etc.) are defined at the top of the file
(lines 7-13); the `.head`/`.empty` classes already exist for stats/bank reuse.

```css
/* ---- friends / ignore tab ---- */
.friends { padding: 4px 6px; font-size: 12px; }
.friends .friendlist { display: flex; flex-direction: column; gap: 1px; margin-bottom: 6px; max-height: 30vh; overflow-y: auto; }
.friends .frow {
  display: flex; align-items: center; gap: 6px;
  padding: 2px 4px; cursor: pointer; color: #cfc;
}
.friends .frow:hover { background: #123012; }
.friends .frow.sel { background: #1a3d1a; }
.friends .frow .nm { flex: 1; }
.friends .frow .dot { width: 8px; height: 8px; border-radius: 50%; flex: 0 0 auto; }
.friends .frow .dot.on { background: var(--rsc-green); }
.friends .frow .dot.off { background: var(--rsc-gray); }
.friends .empty { color: #9a9; padding: 4px; }
.friends .addrow { display: flex; gap: 4px; margin: 2px 0 8px; }
.friends .addrow input {
  flex: 1; background: #06100a; color: #cfc; border: 1px solid #1a3d1a;
  font: 12px monospace; padding: 2px 4px;
}
.friends .addrow input:disabled { color: #566; }
.friends .addrow button, .friends .frow .mini {
  background: #1a3d1a; color: #cfc; border: 1px solid #000;
  font: 12px monospace; padding: 2px 8px; cursor: pointer;
}
.friends .addrow button:hover:not(:disabled) { background: #0a0; color: #000; }
.friends button:disabled { opacity: .45; cursor: not-allowed; }
.friends .frow .mini { padding: 1px 6px; }
```

---

## TRACK B — backend gap (BLOCKED; spec only, do NOT implement here)

> **These edits live in `proto/`, `event/`, `world/`, `action/`, `runtime/` —
> all OFF-LIMITS to the implementer of this spec.** This section is the
> hand-off so a backend owner can unblock the authentic roster + ignore. When
> done, Track A's web code lights up with the change described in B5.

### B1. Outbound ignore opcodes + action fns — `action/pm.go`

Add the opcode constants (mirror the friend block at `action/pm.go:19-23`):

```go
	outAddIgnore    byte = 132 // SOCIAL_ADD_IGNORE    (reference Mudclient.java:4854)
	outRemoveIgnore byte = 241 // SOCIAL_REMOVE_IGNORE (reference Mudclient.java:4883)
```

Add `AddIgnore` / `RemoveIgnore` modeled on `AddFriend` (`action/pm.go:57`):
zero-padded username payload (`buf.WriteZeroPaddedString(name)`), then
`conn.Send(outAddIgnore, buf.Bytes())`. (The reference also sends ADD_IGNORE for
a *temporary* in-session ignore vs a persisted one via a flag, but v1 can send
the persisted form only.)

### B2. Host methods — `runtime/social.go`

Add beside `AddFriend` (`runtime/social.go:18`):

```go
func (h *Host) AddIgnore(ctx context.Context, name string) error {
	return action.AddIgnore(ctx, h.conn, name)
}
func (h *Host) RemoveIgnore(ctx context.Context, name string) error {
	return action.RemoveIgnore(ctx, h.conn, name)
}
```

### B3. Inbound decode — `proto/v235/`

1. Add opcodes to `proto/v235/inbound_opcodes.go`:
   ```go
   InFriendUpdate byte = 149 // SEND_FRIEND_UPDATE — one friend's status/rename
   InIgnoreList   byte = 109 // SEND_IGNORE_LIST   — full ignore list (bulk)
   // optional: InIgnoreListRename byte = 237
   ```
2. Add `case InFriendUpdate:` / `case InIgnoreList:` to the `DecodeInbound`
   switch (`proto/v235/inbound.go:24`) → new decoders.

   **`decodeFriendUpdate` (opcode 149)** — wire format from
   `reference/.../Mudclient.java:6794-6840`:
   - `name`        = `ReadZeroQuotedString()` (current display name)
   - `formerName`  = `ReadZeroQuotedString()`
   - `flags`       = `ReadUnsignedByte()` (bit0 = match-by-former/rename; bit2 = online)
   - if `flags & 4` (online): `world` = `ReadZeroQuotedString()`
   Emit a new `event.FriendUpdate{Name, FormerName, Online: flags&4!=0, World, Rename: flags&1!=0}`.

   **`decodeIgnoreList` (opcode 109)** — `reference/.../Mudclient.java:6874-6881`:
   - `count` = `ReadUnsignedByte()`
   - loop `count`× read 4 zero-quoted strings; the **first** (`current name`) is
     the one to display. Emit `event.IgnoreList{Names: []string}`.

   (The string reader is `(*Buffer).ReadZeroQuotedString`, already used by
   `decodePrivateMessage` at `proto/v235/inbound.go:295`.)

### B4. Events + world mirror — `event/events.go`, `world/`

1. `event/events.go` (beside `PrivateMessage` at `event/events.go:47`):
   ```go
   type FriendUpdate struct { base; Name, FormerName, World string; Online, Rename bool }
   func (FriendUpdate) Kind() string { return "friend_update" }
   type IgnoreList struct { base; Names []string }
   func (IgnoreList) Kind() string { return "ignore_list" }
   ```
2. New `world/social.go`: `type SocialState struct{ mu sync.RWMutex; friends map[string]Friend; friendOrder []string; ignores []string }`
   with `Friend{Name, FormerName, World string; Online bool}`, methods
   `ApplyFriendUpdate(e event.FriendUpdate)` (insert/update/rename in place,
   matching the reference's match-by-name / match-by-former logic at
   `Mudclient.java:6802-6840`), `SetIgnores([]string)`, and a lock-safe
   `Friends() []Friend` + `Ignores() []string` snapshot accessor (mirror the
   `world.BankState.Bank()` snapshot pattern at `world/bank.go:35`).
3. Add `Social *SocialState` to the `World` struct (`world/world.go:13`),
   construct it in `NewWorld` (beside `Bank: NewBankState()` at
   `world/world.go:40`), and wire the two new events into `World.Apply`
   (beside the `event.PrivateMessage` case at `world/world.go:822`):
   ```go
   case event.FriendUpdate: w.Social.ApplyFriendUpdate(e)
   case event.IgnoreList:    w.Social.SetIgnores(e.Names)
   ```
4. (Nice-to-have) flip `runtime/views_world.go:368` `is_friend` to read
   `w.Social` instead of the hard-coded `false`.

### B5. Cradle + web flip once B1–B4 land

- `cmd/cradle/remoteclient.go`: in `/state`, replace the optimistic-roster build
  (A1.3) with reads from `host.World().Social.Friends()` /
  `.Ignores()` — set `Online`/`World` from the snapshot. Replace the
  `add_ignore`/`remove_ignore` backend-gap branch in `/social` (A1.4) with
  `enqueueAction(... host.AddIgnore / host.RemoveIgnore ...)`. The optimistic
  `localFriends`/`socialMu` scaffolding can stay as a pre-ack overlay or be
  deleted.
- `web/`: set `IGNORE_SUPPORTED = true` in `FriendsTab.tsx` and enable the
  disabled inputs. **No other web change** — `f.online`/`f.world` already drive
  the status dot.

---

## Test plan

Build: `npm --prefix web run build && go build ./cmd/cradle` (the running
server on :8090 is **not** restarted by this spec's implementer).

### Track A (buildable now)

1. **Typecheck/build:** `npm --prefix web run build` passes (no TS errors on the
   new `Social`/`Friend` types); `go build ./...` passes (new structs/handler).
2. **`/state` shape:** `curl -s localhost:8090/state | jq .social` →
   `{ "friends": [], "ignores": [] }` initially (always present, never null).
3. **Add friend:** `curl -s -XPOST localhost:8090/social -d '{"op":"add_friend","name":"zezima"}'`
   → `{"ok":true,...}`; next `/state` shows `social.friends[0].name == "zezima"`,
   `online:false`. Confirm the outbound 195 packet is sent (server logs /
   `host.AddFriend` ran).
4. **Remove friend:** `... '{"op":"remove_friend","name":"zezima"}'` → roster
   empties on next `/state`.
5. **Ignore is honestly blocked:** `... '{"op":"add_ignore","name":"x"}'` →
   `{"ok":false,"message":"ignore is not yet supported (backend gap)"}`.
6. **UI:** open the SPA, click the new **Friends** tab. Add a name → row appears
   with a gray dot. Right-click → menu shows `Message` / `Remove friend`. Select
   a friend, type in the PM composer, Send → the line you sent appears as a
   `self`/`you` chat line (existing `/chat` ring echo at `remoteclient.go:1096`);
   an incoming PM from that player appears as a `private` (yellow) chat line.
   Ignore inputs are visibly disabled with the backend-gap tooltip.
7. **Validation:** empty name → `{"ok":false,"message":"name required"}`; bad op
   → unknown-op message; GET `/social` → 405.

### Track B (only after the backend gap is closed)

8. Log in with a real friend online → a 149 burst arrives at login;
   `/state.social.friends[].online` is `true` and the dot turns green; the world
   string (if any) shows in the row tooltip.
9. A friend logging in/out flips `online` within one `/state` poll (400 ms).
10. Server ignore list (opcode 109) populates `/state.social.ignores`;
    add/remove ignore round-trips through `host.AddIgnore`/`RemoveIgnore` (132/241).
11. `is_friend` DSL view (`runtime/views_world.go:368`) returns `true` for a
    rostered friend.

---

## File-touch summary (Track A only — the only part this spec authorizes)

| File | Change |
|---|---|
| `cmd/cradle/remoteclient.go` | `stateSocial`/`stateFriend` types; `Social` field on `stateResponse`; optimistic-roster vars in `serveClient`; populate `social` in `/state`; new `POST /social` handler + `socialRequest`. |
| `web/src/types.ts` | `Friend`, `Social` interfaces; `social: Social` on `GameState`. |
| `web/src/api.ts` | `SocialOp` type + `socialAction()`. |
| `web/src/components/FriendsTab.tsx` | **new** component. |
| `web/src/components/SidePanel.tsx` | third `friends` tab (button + body branch + import). |
| `web/src/styles.css` | `.friends .*` block. |

Off-limits (Track B owner only): `proto/v235/`, `event/`, `world/`, `action/`,
`runtime/` per the changes in §B1-B4.
