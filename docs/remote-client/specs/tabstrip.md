# Spec: Authentic side-panel tab strip (Chrome C3)

**Status:** ready to implement. **Scope:** frontend-only (`web/` only). **Backend:** none — no Go, no new endpoints. **Runs LAST** (after magic/prayer tab components, if any, exist).

## 0. Goal & summary

Replace the two ad-hoc text buttons in `web/src/components/SidePanel.tsx` with an
authentic RSC-style **single-row tab strip**. Per `100-rsc-ui-map.md` §2 the real
client draws six 32px sprite tab icons in one row at the top of the side panel
(`qc` = active tab, values 0–6). We **approximate** that single row with compact
icon-ish buttons styled from the RSC palette CSS vars in `web/src/styles.css`
(the UI-chrome sprite atlas — `spriteMedia`/`tg` — is **not** served to the browser;
see §1).

Tabs to render, **left→right**, in this order:

| Order | Tab | Component | Always? | RSC `qc` |
|------|-----|-----------|---------|----------|
| 1 | Inventory | `InventoryGrid` (+ keep `EquipmentPanel` summary? NO — see §4) | yes | 1 |
| 2 | Equipment | `EquipmentPanel` | yes | 2 (design) |
| 3 | Stats | `StatsPanel` | yes | 3 |
| 4 | Magic | `MagicPanel` | **only if file exists** | 5 |
| 5 | Prayer | `PrayerPanel` | **only if file exists** | 5 |
| 6 | Friends | — (disabled placeholder) | yes | 4 |

### Conditional Magic/Prayer — AUTHORITATIVE check (do this first)

```sh
ls web/src/components/MagicPanel.tsx web/src/components/PrayerPanel.tsx 2>/dev/null
```

As of writing this spec, **NEITHER exists** (`web/src/components/` contains only:
BankWindow, ChatPanel, EquipmentPanel, Hud, InventoryGrid, ItemSprite, SidePanel,
StatsPanel, Viewport). Therefore:

- **If `MagicPanel.tsx` exists:** add the Magic tab + import + body case below.
- **If `PrayerPanel.tsx` exists:** add the Prayer tab + import + body case.
- **If a file is absent:** OMIT that tab entirely (no import, no `TABS` entry, no
  body case). Do **not** create the component — that is another task's job.
- Verify the *exact* exported symbol & props of any magic/prayer component you wire
  in by opening the file; this spec assumes `export function MagicPanel({ state }: { state: GameState | null })`
  and `export function PrayerPanel({ state }: { state: GameState | null })`. If the
  real signature differs, pass the props it declares instead (the panels read from
  `state.self` / future `state.spells`; pass the whole `state` to be safe).

## 1. Backend — NONE

No struct, no handler, no Host method. Confirmed:

- All three always-on tabs render data already in `GameState`: `inventory`
  (`stateResponse.Inventory`), `equipment` (`stateResponse.Equipment`), and
  `self.skills` (`stateResponse.Self.skills`) — see `cmd/cradle/remoteclient.go:297`
  (`stateResponse`) and the `/state` handler that fills them (`remoteclient.go:983`).
- Tab switching is **pure client view state** (`useState`), exactly like the current
  `SidePanel.tsx` `tab` state. The original client's `qc` is also local UI state, not
  a server round-trip.
- **No sprite tab icons available.** `cmd/cradle/sprites.go:27` rejects every `kind`
  except `"item"` (`if kind != "item" { http.Error(..., "unsupported kind (only 'item' for now)", 400) }`).
  So `spriteMedia`/`tg` chrome cannot be fetched; we style buttons with CSS. (If a
  future task adds `kind=media`, the buttons can swap to `<img src="/sprite?kind=media&id=…">`
  — out of scope here.)

> If, while implementing, you discover a Host method is genuinely required (you will
> not — this is view-only state), STOP and note it rather than editing `world/` or
> `runtime/`.

## 2. Web types — NONE required

No additions to `web/src/types.ts`. The `Tab` union is local to `SidePanel.tsx` (§4).
`GameState` already carries everything (`web/src/types.ts:105`). Do not add fields.

## 3. Web api — NONE required

No additions to `web/src/api.ts`. Tab switching makes no network call.

## 4. Component: rewrite `web/src/components/SidePanel.tsx`

Replace the **entire file**. Key behavioral decisions:

1. **Separate Inventory and Equipment tabs.** Today both render under one "inv" tab.
   The C3 task lists "Inventory + Equipment + Stats are always present" as three
   distinct tabs, matching the real client (`qc=1` inventory, `qc=2` design/worn).
   So Inventory tab = `<InventoryGrid>` only; Equipment tab = `<EquipmentPanel>` only.
2. **Friends** is a disabled placeholder: a non-interactive tab button (greyed,
   `disabled`, `aria-disabled`) and, if somehow selected, a "Friends — coming soon"
   stub body. It is never the default and cannot become active (button `disabled`).
3. Tab strip is **one row**; buttons are compact and icon-ish (a glyph + tiny label),
   sized so all rendered tabs fit the 261px panel (see styles §6: each tab ~`flex:1`,
   `min-width:0`, `font:9px`).
4. **Interaction:** left-click selects the tab. **Right-click** (`onContextMenu`) is
   suppressed (`e.preventDefault()`) on the strip so the browser menu never appears
   over it — the real client has no right-click verb on tab headers. (Item cells keep
   their own right-click menus inside the bodies.)
5. Default tab = `'inv'`.

### 4.1 Full file (copy-paste; then delete the Magic/Prayer lines if those files are absent)

```tsx
// Right-hand panel with an authentic-ish RSC tab strip (one row of compact
// icon-ish buttons; 100-rsc-ui-map §2). The real client uses spriteMedia tab
// icons — those aren't served to the browser (sprites.go only serves item
// icons), so we approximate with palette-styled buttons. Tab state is local UI
// state, mirroring the original client's `qc`. Magic/Prayer tabs are wired in
// only when their components exist (this task runs last).

import { useState } from 'react'
import type { GameState } from '../types'
import { EquipmentPanel } from './EquipmentPanel'
import { InventoryGrid } from './InventoryGrid'
import { StatsPanel } from './StatsPanel'
// CONDITIONAL — keep only if the file exists (see spec §0):
import { MagicPanel } from './MagicPanel'
import { PrayerPanel } from './PrayerPanel'

type Tab = 'inv' | 'equip' | 'stats' | 'magic' | 'prayer' | 'friends'

interface TabDef {
  id: Tab
  glyph: string // compact icon-ish stand-in for the spriteMedia tab icon
  label: string
  disabled?: boolean
}

// Ordered left→right. Magic/Prayer entries are CONDITIONAL: remove the line if
// the matching component file does not exist (spec §0). Order matches RSC qc.
const TABS: TabDef[] = [
  { id: 'inv', glyph: '\u{1F392}', label: 'Inv' },     // 🎒
  { id: 'equip', glyph: '\u{1F6E1}', label: 'Worn' },  // 🛡
  { id: 'stats', glyph: '\u{1F4CA}', label: 'Stats' }, // 📊
  { id: 'magic', glyph: '✨', label: 'Magic' },    // ✨  (remove if no MagicPanel)
  { id: 'prayer', glyph: '\u{1F64F}', label: 'Pray' }, // 🙏  (remove if no PrayerPanel)
  { id: 'friends', glyph: '\u{1F465}', label: 'Friends', disabled: true }, // 👥
]

export function SidePanel({ state }: { state: GameState | null }) {
  const [tab, setTab] = useState<Tab>('inv')
  return (
    <div id="side">
      <div className="tabstrip" role="tablist" onContextMenu={(e) => e.preventDefault()}>
        {TABS.map((t) => (
          <button
            key={t.id}
            role="tab"
            type="button"
            aria-selected={tab === t.id}
            aria-disabled={t.disabled || undefined}
            disabled={t.disabled}
            title={t.label}
            className={'tab' + (tab === t.id ? ' active' : '') + (t.disabled ? ' disabled' : '')}
            onClick={() => { if (!t.disabled) setTab(t.id) }}
          >
            <span className="tglyph" aria-hidden="true">{t.glyph}</span>
            <span className="tlabel">{t.label}</span>
          </button>
        ))}
      </div>
      <div className="tabbody">
        {tab === 'inv' && <InventoryGrid inventory={state?.inventory ?? []} />}
        {tab === 'equip' && <EquipmentPanel equipment={state?.equipment ?? []} />}
        {tab === 'stats' && <StatsPanel self={state?.self ?? null} />}
        {tab === 'magic' && <MagicPanel state={state} />}
        {tab === 'prayer' && <PrayerPanel state={state} />}
        {tab === 'friends' && <div className="stub-body">Friends — coming soon</div>}
      </div>
    </div>
  )
}
```

### 4.2 Deletion checklist when a component is absent

If `MagicPanel.tsx` does **not** exist, delete ALL of:
- the `import { MagicPanel } from './MagicPanel'` line,
- the `{ id: 'magic', ... }` entry in `TABS`,
- the `{tab === 'magic' && <MagicPanel state={state} />}` body line,
- (optional) `'magic'` from the `Tab` union — harmless to leave but cleaner to drop.

Same pattern for `PrayerPanel.tsx` / `'prayer'`. **Friends stays regardless.**

> The file as written imports MagicPanel/PrayerPanel unconditionally; with the
> current tree those imports DO NOT RESOLVE and `tsc`/vite build will FAIL. You MUST
> apply §4.2 before building when the files are absent. (At spec time: delete both
> magic and prayer lines.)

## 5. App wiring — NO CHANGE

`web/src/App.tsx:96` already renders `<SidePanel state={state} />` with the full
state object. No change needed — `SidePanel` keeps the same single `state` prop.

## 6. Styles: `web/src/styles.css`

The current `.tabbar` rules (lines ~117–132) target the old two-button bar. **Replace
the `.tabbar` / `.tabbar button` / `.tabbar button.active` block** (keep `.tabbody`)
with the strip styles below. Reuse the existing palette vars (`--rsc-yellow`,
`--rsc-cyan`, `--panel-bg`, `--panel-border`, etc.).

```css
/* ---- side-panel tab strip (one row; 100-rsc-ui-map §2). Approximates the
 * spriteMedia tab icons with compact palette-styled buttons. ---- */
.tabstrip {
  display: flex;
  border-bottom: 2px solid var(--panel-border);
  background: var(--panel-inset);
}
.tabstrip .tab {
  flex: 1 1 0;
  min-width: 0;                 /* let all tabs fit the 261px panel */
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1px;
  background: var(--panel-bg);
  color: #6c6;
  border: 0;
  border-right: 1px solid var(--panel-border);
  padding: 3px 0 2px;
  cursor: pointer;
  line-height: 1;
}
.tabstrip .tab:last-child { border-right: 0; }
.tabstrip .tab .tglyph { font-size: 14px; }
.tabstrip .tab .tlabel { font: 9px monospace; letter-spacing: -0.3px; }
.tabstrip .tab:hover:not(.disabled):not(.active) { background: #0f1f0f; color: var(--rsc-cyan); }
.tabstrip .tab.active {
  background: #0a0;
  color: #000;
  font-weight: bold;
  box-shadow: inset 0 -2px 0 var(--rsc-yellow);
}
.tabstrip .tab.disabled { color: #3a4a3a; cursor: default; }

.stub-body { padding: 8px; color: #6c6; font-size: 12px; }
```

> If `.tabbody` already exists (it does, ~line 132) keep it; only the `.tabbar*`
> rules are replaced. Do not leave dead `.tabbar` selectors.

### 6.1 (Optional, cheap) fixed 512×334-aspect game view

The task offers this as optional. The viewport currently letterboxes via
`object-fit: contain` on `#view img` (`styles.css:43`), and the grid column for the
view is `1fr` (`#app { grid-template-columns: 1fr 261px }`, line 32). To pin the
**rendered frame** to the authentic 512:334 ≈ 1.533 aspect without touching layout
logic, add to `#view`:

```css
#view { aspect-ratio: 512 / 334; }   /* optional C3 stretch goal */
```

Caveat: `#view` is a grid area sized by `1fr`/`1fr 168px`; `aspect-ratio` on a grid
item with a stretched height may be ignored. **Only ship this if a quick visual check
shows the frame keeping the 512×334 shape; otherwise omit it** — it is explicitly
optional and must not regress the existing fill behavior. Do not restructure `#app`.

## 7. Test plan

Build + manual smoke (the cradle server is already live on :8090 — do **not** restart it):

```sh
npm --prefix web run build      # must compile clean (tsc + vite). FAILS if a
                                # MagicPanel/PrayerPanel import is left in §4.1
                                # while the file is absent — fix per §4.2.
go build ./...                  # re-embeds web/dist into the cradle binary
```

Then, in the browser at the live client (or `npm --prefix web run dev`):

1. **Strip renders one row.** Top of the right panel shows a single row of tabs:
   Inv, Worn, Stats, [Magic, Pray — only if those components exist], Friends. No
   wrapping; all fit within the 261px panel.
2. **Default = Inventory.** On load the Inv tab is active (green/yellow underline)
   and the 30-slot grid shows, with authentic item icons via `<ItemSprite>`.
3. **Switching works.** Click Worn → `EquipmentPanel` rows. Click Stats →
   `StatsPanel` (combat/qp/fatigue header + skill rows). Click Magic/Pray (if
   present) → their panels. Active styling moves with the selection.
4. **Friends disabled.** The Friends tab is greyed, shows a `not-allowed`/default
   cursor, does not change the active tab when clicked, and never shows its body.
5. **No browser context menu** when right-clicking the tab strip; item cells inside
   the Inventory body STILL open their right-click verb menu (regression check —
   `InventoryGrid` `onContextMenu` unchanged).
6. **Palette fidelity:** active tab uses RSC green/yellow; idle uses dim green;
   hover uses cyan. Colors come from the `--rsc-*` / `--panel-*` vars (no hardcoded
   new colors beyond the existing scheme).
7. (If §6.1 shipped) the rendered frame keeps the 512×334 aspect and is not
   stretched; if it regresses fill behavior, revert §6.1 only.

## 8. Files touched (all under `web/` — hard constraint satisfied)

- `web/src/components/SidePanel.tsx` — full rewrite (§4).
- `web/src/styles.css` — replace `.tabbar*` block with `.tabstrip*` + `.stub-body`;
  optional `#view { aspect-ratio }` (§6).
- (Conditional, only if they already exist) imports of `MagicPanel`/`PrayerPanel`.

No changes to `cmd/cradle/`, `world/`, `runtime/`, `reference/`, `types.ts`, `api.ts`,
or `App.tsx`.
