// Typed client for the cradle JSON API. Every path is origin-relative so it
// works both under `vite dev` (proxied to :8090) and when embedded in the
// cradle binary (served at /).

import type {
  ActResponse, ClientConfig, GameState, MenuTarget, PickResponse, SpellDef,
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

/** Walk to an absolute-X / plane-local-Y world tile (used by the minimap §4.4).
 *  Posts a synthetic terrain target through /act: ResolveLane does NOT
 *  identity-validate KindTerrain, BuildMenu yields a single "Walk here" option at
 *  index 0, and dispatchTerrain calls host.Walk(ctx, x, y) directly. So optionId 0
 *  == "Walk here". x is absolute world X; y is plane-local (entity.Y - plane*944),
 *  the space host.Walk expects. */
export async function walkTile(x: number, y: number): Promise<ActResponse> {
  const ref: MenuTarget = { kind: 'terrain', x, y, slot: -1 }
  return act(ref, 0)
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

/** POST /prayer — activate (on:true) or deactivate (on:false) one prayer slot. */
export async function prayerAction(id: number, on: boolean): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/prayer', { id, on })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

export type TradeOp = 'offer' | 'accept' | 'finalize' | 'decline'

export interface TradeItemInput {
  itemId: number
  amount: number
}

/** POST /trade — send a trade window action. */
export async function tradeAction(
  op: TradeOp,
  items?: TradeItemInput[],
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/trade', { op, items })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

export type DuelOp = 'stake' | 'rules' | 'accept1' | 'accept2' | 'decline'

export interface DuelStakeItem { itemId: number; amount: number }

export interface DuelRulesPayload {
  disallowRetreat: boolean
  disallowMagic: boolean
  disallowPrayer: boolean
  disallowWeapons: boolean
}

/** POST /duel — stake items, set rules, accept (two-stage), or decline. */
export async function duelAction(
  op: DuelOp,
  items?: DuelStakeItem[],
  rules?: DuelRulesPayload,
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/duel', { op, items, rules })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

export type ShopOp = 'buy' | 'sell' | 'close'

/** POST /shop — buy/sell a quantity of a catalogue item, or close the window. */
export async function shopAction(op: ShopOp, itemId = 0, amount = 0): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/shop', { op, itemId, amount })
  } catch (e) {
    return { ok: false, message: String(e) }
  }
}

// Module-level cache: fetched once, never re-fetched.
let _spellCatalog: SpellDef[] | null = null

/** GET /spells — static; cached after first fetch. */
export async function getSpells(): Promise<SpellDef[]> {
  if (_spellCatalog) return _spellCatalog
  _spellCatalog = await (await fetch('/spells')).json() as SpellDef[]
  return _spellCatalog
}

export type CastTargetKind = 'self' | 'npc' | 'player'

/** POST /cast — routes through the serialised action worker server-side. */
export async function castSpell(
  spellId: number,
  targetKind: CastTargetKind = 'self',
  targetIndex = 0,
): Promise<ActResponse> {
  try {
    return await postJSON<ActResponse>('/cast', { spellId, targetKind, targetIndex })
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
