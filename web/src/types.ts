// Wire types — mirror the Go JSON contracts in cmd/cradle/remoteclient.go and
// the remoteclient package (target.go). Keep field names in sync with the
// `json:"..."` tags on the server structs.

export type TargetKind =
  | 'npc' | 'player' | 'self' | 'ground_item'
  | 'scenery' | 'boundary' | 'terrain' | 'inventory_item'

/** MenuTarget — a stable reference to something pickable (target.go). */
export interface MenuTarget {
  kind: TargetKind
  index?: number
  x: number
  y: number
  dir?: number
  id?: number
  slot: number
  name?: string
}

/** One menu verb for a candidate; id indexes the candidate's options[]. */
export interface MenuOption {
  id: number
  verb: string
}

/** A right-click candidate: a target plus its ordered verbs. */
export interface Candidate {
  ref: MenuTarget
  label: string
  options: MenuOption[]
}

export interface PickResponse {
  candidates: Candidate[]
}

export interface ActResponse {
  ok: boolean
  message?: string
}

export interface Skill {
  id: number
  name: string
  level: number
  max: number
  xp: number
}

export interface Self {
  x: number
  y: number
  plane: number
  heading: number
  combatLevel: number
  hp: number
  maxHp: number
  prayer: number
  maxPrayer: number
  fatigue: number
  questPoints: number
  skills: Skill[]
}

export interface InvItem {
  slot: number
  itemId: number
  name: string
  amount: number
  wielded: boolean
  wearable: boolean
  stackable: boolean
  command: string
  defaultOptionId: number
  options: MenuOption[]
}

export interface EquipItem {
  slot: string
  sprite: number
  itemId: number
}

export interface ChatEntry {
  seq: number
  kind: string // public | npc | private | system | self
  who: string
  text: string
}

export interface GameState {
  self: Self
  inventory: InvItem[]
  equipment: EquipItem[]
  chat: ChatEntry[]
}

/** GET /config — render defaults injected by the server at boot. */
export interface ClientConfig {
  username: string
  zoom: number
  w: number
  h: number
  rotation: number
}
