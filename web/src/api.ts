// Typed client for the cradle JSON API. Every path is origin-relative so it
// works both under `vite dev` (proxied to :8090) and when embedded in the
// cradle binary (served at /).

import type {
  ActResponse, ClientConfig, GameState, MenuTarget, PickResponse,
} from './types'

export interface Camera {
  rot: number
  zoom: number
  w: number
  h: number
  anim?: number
}

async function postJSON<T>(path: string, body: unknown): Promise<T> {
  const r = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  })
  return r.json() as Promise<T>
}

export async function getConfig(): Promise<ClientConfig> {
  return (await fetch('/config')).json()
}

export async function getState(): Promise<GameState> {
  return (await fetch('/state')).json()
}

/** URL for one rendered frame. Use as an <img> src; the cache-buster forces a
 *  fresh decode each tick. */
export function frameURL(cam: Camera): string {
  const q = new URLSearchParams({
    rot: String(cam.rot), zoom: String(cam.zoom),
    w: String(cam.w), h: String(cam.h),
    anim: String(cam.anim ?? 0), t: String(performance.now()),
  })
  return `/frame?${q}`
}

export function walk(px: number, py: number, cam: Camera): void {
  const q = new URLSearchParams({
    px: String(px), py: String(py),
    rot: String(cam.rot), zoom: String(cam.zoom),
    w: String(cam.w), h: String(cam.h),
  })
  void fetch(`/walk?${q}`)
}

export async function pick(
  px: number, py: number, cam: Camera, slot = -1,
): Promise<PickResponse> {
  try {
    return await postJSON<PickResponse>('/pick', {
      px, py, rot: cam.rot, zoom: cam.zoom, w: cam.w, h: cam.h, slot,
    })
  } catch {
    return { candidates: [] }
  }
}

export async function act(ref: MenuTarget, optionId: number): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/act', { ref, optionId })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

export type ChatKind = 'say' | 'command' | 'pm'

export async function sendChat(
  kind: ChatKind, text: string, to?: string,
): Promise<ActResponse> {
  const body: Record<string, string> = { kind, text }
  if (kind === 'pm' && to) body.to = to
  return postJSON<ActResponse>('/chat', body)
}

export type BankOp = 'deposit' | 'withdraw' | 'close'

/** POST /bank — deposit/withdraw a quantity of an item, or close the window. */
export async function bankAction(op: BankOp, itemId = 0, amount = 0): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/bank', { op, itemId, amount })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

export function shot(cam: Camera): Promise<Response> {
  return fetch(`/shot?${new URLSearchParams({
    rot: String(cam.rot), zoom: String(cam.zoom), w: String(cam.w), h: String(cam.h),
  })}`)
}

export function clip(cam: Camera): Promise<Response> {
  return fetch(`/clip?${new URLSearchParams({
    rot: String(cam.rot), zoom: String(cam.zoom), w: String(cam.w), h: String(cam.h),
  })}`)
}
