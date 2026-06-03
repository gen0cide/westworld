// Viewport: the streamed 3D frame plus screen interaction. Drives the ~30fps
// <img> swap loop (mirrors the original tick()), converts screen → frame coords
// undoing object-fit:contain letterboxing, and routes left-click → default act,
// right-click → context menu via /pick.

import { useEffect, useRef } from 'react'
import { frameURL, pick, act as apiAct, useOn, walk, type Camera } from '../api'
import { useUI } from '../ui'
import { useDrag } from '../drag'

export interface Camera4 { rot: number; zoom: number; w: number; h: number }

export function Viewport(
  { camera, animRef, hud }: {
    camera: Camera4
    animRef: React.MutableRefObject<number>
    hud: React.ReactNode
  },
) {
  const imgRef = useRef<HTMLImageElement | null>(null)
  const camRef = useRef<Camera>({ ...camera, anim: 0 })
  const busy = useRef(false)
  const ui = useUI()
  const drag = useDrag()

  // Keep the loop's camera snapshot current without re-creating the interval.
  useEffect(() => { camRef.current = { ...camera, anim: animRef.current } }, [camera, animRef])

  useEffect(() => {
    const id = window.setInterval(() => {
      if (busy.current) return
      busy.current = true
      const cam = { ...camRef.current, anim: animRef.current }
      const next = new Image()
      next.onload = () => { if (imgRef.current) imgRef.current.src = next.src; busy.current = false }
      next.onerror = () => { busy.current = false }
      next.src = frameURL(cam)
    }, 33)
    return () => window.clearInterval(id)
  }, [animRef])

  // screen → frame pixel, undoing letterbox of object-fit:contain.
  const screenToFrame = (e: React.MouseEvent): { px: number; py: number } | null => {
    const img = imgRef.current
    if (!img) return null
    const r = img.getBoundingClientRect()
    const { w, h } = camera
    const scale = Math.min(r.width / w, r.height / h)
    if (scale <= 0) return null
    const px = Math.round((e.clientX - (r.left + (r.width - w * scale) / 2)) / scale)
    const py = Math.round((e.clientY - (r.top + (r.height - h * scale) / 2)) / scale)
    if (px < 0 || py < 0 || px >= w || py >= h) return null
    return { px, py }
  }

  const onClick = async (e: React.MouseEvent) => {
    const fp = screenToFrame(e)
    if (!fp) return
    const cam: Camera = { ...camera }
    const j = await pick(fp.px, fp.py, cam)
    const cand = j.candidates[0]
    const opt = cand?.options[0]
    if (cand && opt) {
      const r = await apiAct(cand.ref, opt.id)
      if (r.message) ui.flash(r.message)
    } else {
      walk(fp.px, fp.py, cam) // terrain fallback
    }
  }

  const onContextMenu = async (e: React.MouseEvent) => {
    e.preventDefault()
    const fp = screenToFrame(e)
    if (!fp) return
    const sx = e.clientX, sy = e.clientY
    const j = await pick(fp.px, fp.py, { ...camera })
    ui.openMenu(sx, sy, j.candidates)
  }

  // Drop a dragged inventory item onto the world (spec §2.7). Reuse the same
  // screenToFrame pixel math the click/pick path uses, /pick at that pixel,
  // take the top non-terrain candidate, and useOn its ref.
  const onDrop = async (e: React.DragEvent) => {
    e.preventDefault()
    const src = drag.take()
    if (!src) return
    const fp = screenToFrame(e)
    if (!fp) { ui.flash('nothing to use that on'); return }
    const cam: Camera = { ...camera }
    const j = await pick(fp.px, fp.py, cam)
    const top = j.candidates.find((c) => c.ref.kind !== 'terrain')
    if (!top) { ui.flash('nothing to use that on'); return }
    const r = await useOn(src.slot, top.ref)
    if (r.message) ui.flash(r.message)
  }

  return (
    <div id="view">
      <img
        ref={imgRef}
        draggable={false}
        onClick={onClick}
        onContextMenu={onContextMenu}
        onDragOver={(e) => e.preventDefault()}
        onDrop={onDrop}
      />
      {hud}
    </div>
  )
}
