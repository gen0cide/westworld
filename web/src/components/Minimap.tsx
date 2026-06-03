// Minimap (F1 — 100-rsc-ui-map §4.4): a round 152px <canvas> radar overlay in the
// top-right of #view. It plots the nearby-entity dots from /state.entities around a
// white self dot at the centre, and ROTATES with the camera so "up" is the camera's
// forward direction (mirrors the authentic drawUiTabMinimap rotate-by-yaw).
//
// Coordinate roles (see types.ts MinimapDot):
//   - dx,dy  RELATIVE world deltas (entity.X-self.X, entity.Y-self.Y) — plotted +
//            rotated here.
//   - x,y    absolute-X / plane-local-Y world tile — passed to walkTile() for
//            click-to-walk (host.Walk expects plane-local Y; the backend folds it).
//
// Interaction: left-click a dot -> walk to that thing's tile; left-click empty ->
// walk to the world tile under the cursor (inverse-rotate the click to a delta,
// add self). Right-click is suppressed (a verb menu needs a world-tile /pick which
// the current screen-pixel /pick can't provide — documented follow-up).

import { useEffect, useRef } from 'react'
import { walkTile } from '../api'
import { useUI } from '../ui'
import type { DotKind, Entities, MinimapDot } from '../types'

const SIZE = 152
const PAD = 6

// §4.4 dot palette. Literal hexes (canvas can't read CSS vars without
// getComputedStyle); scenery is pure 0x00FFFF for fidelity (the --rsc-cyan var is
// the lighter #c8ffff, used elsewhere in the UI).
const DOT_COLOR: Record<DotKind, string> = {
  npc: '#ffff00',
  player: '#ffffff',
  ground_item: '#ff0000',
  scenery: '#00ffff',
}
const SELF_COLOR = '#ffffff'

// RSC rotates the map so the world spins opposite the camera; the camera yaw rot is
// the 256-step compass (App ROT step = 4, range 0..255). A full turn = 2*PI.
function thetaFor(rot: number): number {
  return (rot / 256) * Math.PI * 2
}

export function Minimap(
  { entities, rot, self }: {
    entities: Entities | undefined
    rot: number
    self?: { x: number; y: number }
  },
) {
  const ref = useRef<HTMLCanvasElement | null>(null)
  const ui = useUI()
  const dots = entities?.dots ?? []
  const radius = entities?.radius ?? 16

  useEffect(() => {
    const cv = ref.current
    if (!cv) return
    const g = cv.getContext('2d')
    if (!g) return
    const c = SIZE / 2
    const scale = (SIZE / 2 - PAD) / radius
    const theta = thetaFor(rot)

    // dark circular field
    g.clearRect(0, 0, SIZE, SIZE)
    g.fillStyle = '#0a140a'
    g.beginPath()
    g.arc(c, c, SIZE / 2, 0, Math.PI * 2)
    g.fill()

    // a faint north notch (does NOT rotate — points at the panel's top edge).
    g.fillStyle = '#244a24'
    g.fillRect(c - 1, 1, 2, 4)

    // entity dots: world (dx,dy) -> canvas (right=+x, down=+y), then rotate by theta.
    for (const d of dots) {
      const px = c + (d.dx * Math.cos(theta) - d.dy * Math.sin(theta)) * scale
      const py = c + (d.dx * Math.sin(theta) + d.dy * Math.cos(theta)) * scale
      if (Math.hypot(px - c, py - c) > SIZE / 2 - 1) continue
      g.fillStyle = DOT_COLOR[d.kind] ?? '#fff'
      g.fillRect(Math.round(px) - 1, Math.round(py) - 1, 3, 3)
    }

    // self at the centre (slightly taller so it reads as the focus dot).
    g.fillStyle = SELF_COLOR
    g.fillRect(c - 1, c - 2, 3, 4)
  }, [dots, radius, rot])

  const onClick = (e: React.MouseEvent<HTMLCanvasElement>) => {
    const cv = ref.current
    if (!cv) return
    const r = cv.getBoundingClientRect()
    const mx = (e.clientX - r.left) * (SIZE / r.width)
    const my = (e.clientY - r.top) * (SIZE / r.height)
    const c = SIZE / 2
    const scale = (SIZE / 2 - PAD) / radius
    const theta = thetaFor(rot)

    // nearest dot first (left-click "walk to that thing's tile"), within ~6px.
    let best: MinimapDot | null = null
    let bestD = 6 * 6
    for (const d of dots) {
      const px = c + (d.dx * Math.cos(theta) - d.dy * Math.sin(theta)) * scale
      const py = c + (d.dx * Math.sin(theta) + d.dy * Math.cos(theta)) * scale
      const dd = (px - mx) ** 2 + (py - my) ** 2
      if (dd < bestD) {
        bestD = dd
        best = d
      }
    }
    if (best) {
      const tile = best
      void walkTile(tile.x, tile.y).then((res) => {
        if (res.message) ui.flash(res.message)
      })
      return
    }

    // empty space: inverse-rotate the click back to a world delta, add self -> walk.
    if (!self) {
      ui.flash('click a dot to walk')
      return
    }
    const rx = (mx - c) / scale
    const ry = (my - c) / scale
    const inv = -theta
    const dx = Math.round(rx * Math.cos(inv) - ry * Math.sin(inv))
    const dy = Math.round(rx * Math.sin(inv) + ry * Math.cos(inv))
    void walkTile(self.x + dx, self.y + dy).then((res) => {
      if (res.message) ui.flash(res.message)
    })
  }

  return (
    <canvas
      ref={ref}
      width={SIZE}
      height={SIZE}
      className="minimap"
      onClick={onClick}
      onContextMenu={(e) => e.preventDefault()}
      title="minimap"
    />
  )
}
