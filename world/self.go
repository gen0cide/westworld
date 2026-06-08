package world

import (
	"sync"

	"github.com/gen0cide/westworld/event"
)

// PlaneHeight is the vertical distance (in tiles) between stacked
// floors in the RSC world map. Floors are not a separate coordinate on
// the wire — they are encoded as a Y offset, so the floor (plane) of
// any tile is its Y divided by this height.
//
// Source: OpenRSC Formulae.getHeight = (int)(y / 944) and
// ActionSender.sendWorldInfo's distanceBetweenFloors = 944
// (SEND_WORLD_INFO). The client derives its render floor the same way.
const PlaneHeight = 944

// PlaneOf returns the floor/plane index for an absolute world Y. 0 is
// ground level, 1 is the first upper floor, etc. (RSC also uses plane 3
// for the underground / dungeon band.) Pure derivation — no packet.
func PlaneOf(y int) int {
	if y < 0 {
		return 0
	}
	return y / PlaneHeight
}

// Coord is an RSC world tile coordinate. RSC uses big-endian uint16
// for both x and y on the wire. We keep them as ints here for ergonomic
// arithmetic.
//
// Y carries the floor: tiles on upper floors have Y offset by a
// multiple of PlaneHeight. Use Plane() to recover the floor index.
type Coord struct {
	X int
	Y int
}

// Plane returns the floor index this coordinate sits on (Y / PlaneHeight).
func (c Coord) Plane() int { return PlaneOf(c.Y) }

// NumSkills matches OpenRSC's 18-skill catalog (Attack..Thieving).
const NumSkills = 18

// Self is the host's view of its own player. Read/write safe across
// goroutines.
type Self struct {
	mu sync.RWMutex

	position    Coord
	heading     int            // 8-way facing (own-coords sprite, 0=N..7=NW)
	skillLevels [NumSkills]int // current (boostable) level
	skillMax    [NumSkills]int // max (base) level
	skillXP     [NumSkills]int // total xp
	fatigue     int            // 0..750 scaled
	questPoints int

	// lastDeathAt records the tile where we died (captured by
	// Apply(event.Death) BEFORE the respawn position packet
	// overwrites self.position). Zero-value Coord{} until first
	// death. Routines reading `self.last_death_at` after `on death`
	// fires get the death spot, which is useful for "walk back to
	// where I died" recovery patterns.
	lastDeathAt Coord
	// deathCount increments each time we die. Useful for routines
	// that want to give up after N deaths.
	deathCount int

	// sleeping mirrors the sleep-screen state. Set true when the
	// server sends SEND_SLEEPSCREEN (v235 opcode 117) — the old
	// anti-bot fatigue captcha is up — and false when it sends
	// SEND_STOPSLEEP (v235 opcode 84, the wake packet). Backs the
	// self.is_sleeping accessor. While true, the only useful action
	// is answering the sleep word (the runtime auto-answers "asleep").
	sleeping bool

	// activePrayers mirrors SEND_PRAYERS_ACTIVE — one bool per
	// prayer slot, indexed 0..13 (Thick Skin, Burst of Strength,
	// Clarity of Thought, Rock Skin, Superhuman Strength, Improved
	// Reflexes, Rapid Restore, Rapid Heal, Protect Item, Steel
	// Skin, Ultimate Strength, Incredible Reflexes, Protect from
	// Magic, Protect from Missiles, Protect from Melee — 15
	// total in OpenRSC). Length matches whatever the server sends.
	activePrayers []bool

	// equipBySlot is the host's own per-slot worn-equipment SPRITE ids,
	// indexed by event.EquipSlot* (head, shirt, pants, shield, weapon,
	// hat, body, legs, gloves, boots, amulet, cape). These are appearance
	// SPRITE ids (itemDef.getAppearanceId() & 0xFF), NOT catalogue item
	// ids — identical in shape to PlayerRecord.EquipBySlot. A zero slot
	// means nothing worn there. hasEquip is true once a worn-equipment
	// block has been observed for self.
	//
	// NOTE: nothing decodes into this yet. Self's own appearance update
	// rides in the same opcode-234 type-5 record as other players', but
	// it is keyed by our player index, and the runtime does not yet track
	// which index is ours (the own-position packet, opcode 191, carries
	// no index). Until self-index identification lands, this stays zero
	// and self.equipped.<slot> reports 0. See blockers.
	equipBySlot [event.NumEquipSlots]int
	hasEquip    bool

	// The host's own appearance colour indices (hair, top, trouser/bottom,
	// skin), from the type-5 appearance update for our own player index
	// (SelfPlayerIndex == 0). They index the client clothing/skin colour
	// tables and dye our composite sprite layers in the render path.
	// hasColours is true once they have been observed for self.
	colourHair    int
	colourTop     int
	colourTrouser int
	colourSkin    int
	hasColours    bool
}

// NewSelf returns a Self with zero values. Caller should update from
// inbound packets after login.
func NewSelf() *Self { return &Self{} }

// Position returns the current believed position.
func (s *Self) Position() Coord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.position
}

// SetPosition updates the position.
func (s *Self) SetPosition(c Coord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.position = c
}

// Heading returns the host's last-reported 8-way facing (0=N..7=NW), the
// own-position sprite the renderer adds to the camera term to choose which side
// of the local-player sprite to draw.
func (s *Self) Heading() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.heading
}

// SetHeading records the host's 8-way facing (masked to 0..7).
func (s *Self) SetHeading(h int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.heading = h & 7
}

// Plane returns the floor index of our current position (0 = ground,
// 1+ = upper floors, 3 = the underground band). Derived from the wire
// Y — RSC stacks floors vertically in Y-space at PlaneHeight (944)
// intervals; there is no separate floor field on any packet. Feeds the
// post-rename self.position.plane accessor.
func (s *Self) Plane() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return PlaneOf(s.position.Y)
}

// SkillLevel returns the current (boostable) level of a skill.
func (s *Self) SkillLevel(id int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id < 0 || id >= NumSkills {
		return 0
	}
	return s.skillLevels[id]
}

// SkillMax returns the max (base) level of a skill.
func (s *Self) SkillMax(id int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id < 0 || id >= NumSkills {
		return 0
	}
	return s.skillMax[id]
}

// SkillXP returns the total experience for a skill.
func (s *Self) SkillXP(id int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if id < 0 || id >= NumSkills {
		return 0
	}
	return s.skillXP[id]
}

// SetSkill updates one skill's current/max/xp values.
func (s *Self) SetSkill(id, cur, mx, xp int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if id < 0 || id >= NumSkills {
		return
	}
	s.skillLevels[id] = cur
	s.skillMax[id] = mx
	s.skillXP[id] = xp
}

// SetAllSkills replaces all 18 skills at once (after the full-stats packet).
func (s *Self) SetAllSkills(cur, mx, xp [NumSkills]int, questPoints int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.skillLevels = cur
	s.skillMax = mx
	s.skillXP = xp
	s.questPoints = questPoints
}

// Fatigue returns the current fatigue value (server-scaled, 0..750).
// This is the RAW source of truth; the wire protocol and policy/decision
// consumers depend on this scale. Human/LLM-facing readers should use
// FatiguePercent instead.
func (s *Self) Fatigue() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fatigue
}

// FatiguePercent returns fatigue normalized to 0..100 (rounded) for
// human/LLM-facing display. RSC stores fatigue raw as 0..750; readers the
// brain or an operator sees (the self.fatigue DSL accessor, the examine
// string, the debug repl, the dashboard snapshot) report this percentage so
// "if self.fatigue > 80" means ~80%, not ~600/750.
func (s *Self) FatiguePercent() int {
	s.mu.RLock()
	raw := s.fatigue
	s.mu.RUnlock()
	return (raw*100 + 375) / 750 // round(raw*100/750)
}

// SetFatigue updates fatigue from an inbound packet.
func (s *Self) SetFatigue(v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fatigue = v
}

// IsSleeping reports whether the fatigue sleep-screen is currently up.
// True between SEND_SLEEPSCREEN (opcode 117) and SEND_STOPSLEEP (opcode
// 84). Backs the self.is_sleeping DSL accessor.
func (s *Self) IsSleeping() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sleeping
}

// SetSleeping updates the sleep-screen state. Called by the runtime on
// SEND_SLEEPSCREEN (true) and SEND_STOPSLEEP (false).
func (s *Self) SetSleeping(v bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sleeping = v
}

// QuestPoints returns the host's current QP total.
func (s *Self) QuestPoints() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.questPoints
}

// HP is a convenience accessor for the Hits skill's current value.
func (s *Self) HP() int { return s.SkillLevel(3) }

// LastDeathAt returns the tile where we last died, or zero-value
// Coord{} if never. Snapshotted by RecordDeath before the respawn
// position packet overwrites the live position.
func (s *Self) LastDeathAt() Coord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastDeathAt
}

// DeathCount returns the number of times we've died this session.
func (s *Self) DeathCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.deathCount
}

// RecordDeath captures the current position as the death spot and
// increments the death counter. Called by world.Apply on event.Death,
// BEFORE the subsequent respawn position packet lands.
func (s *Self) RecordDeath() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastDeathAt = s.position
	s.deathCount++
}

// SetActivePrayers replaces the active-prayer bitmap. Called on
// inbound SEND_PRAYERS_ACTIVE (opcode 206).
func (s *Self) SetActivePrayers(active []bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.activePrayers = append([]bool(nil), active...)
}

// PrayerActive returns true iff prayer slot idx is currently on.
// Out-of-range indices return false (safe default).
func (s *Self) PrayerActive(idx int) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if idx < 0 || idx >= len(s.activePrayers) {
		return false
	}
	return s.activePrayers[idx]
}

// ActivePrayers returns a copy of the active-prayer bitmap.
func (s *Self) ActivePrayers() []bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]bool(nil), s.activePrayers...)
}

// EquipSpriteAt returns the worn-equipment SPRITE id in the given equip
// slot (event.EquipSlot*), or 0 if nothing is worn there / the slot is
// out of range. These are appearance sprite ids, not catalogue item
// ids. Feeds the self.equipped.<slot> accessor.
func (s *Self) EquipSpriteAt(slot int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if slot < 0 || slot >= event.NumEquipSlots {
		return 0
	}
	return s.equipBySlot[slot]
}

// HasEquip reports whether a worn-equipment block has been observed for
// self. Until self-index identification lands this is always false (see
// the equipBySlot field doc).
func (s *Self) HasEquip() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.hasEquip
}

// SetWornEquipment records the host's own per-slot worn-equipment sprite
// ids (indexed by event.EquipSlot*). Additive setter for the future
// self-appearance landing path; not called by world.Apply yet because
// the runtime cannot yet identify which player index is ours.
func (s *Self) SetWornEquipment(worn [event.NumEquipSlots]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.equipBySlot = worn
	s.hasEquip = true
}

// EquipSprites returns a copy of the host's full per-slot worn-equipment
// sprite-id array (indexed by event.EquipSlot*). All-zero until self's
// appearance update lands.
func (s *Self) EquipSprites() [event.NumEquipSlots]int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.equipBySlot
}

// SetAppearanceColours records the host's own appearance colour indices
// (hair, top, trouser/bottom, skin — the wire order). Called from
// world.Apply when the type-5 appearance update for our own player index
// arrives.
func (s *Self) SetAppearanceColours(hair, top, trouser, skin int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.colourHair = hair
	s.colourTop = top
	s.colourTrouser = trouser
	s.colourSkin = skin
	s.hasColours = true
}

// AppearanceColours returns the host's own (hair, top, trouser, skin)
// colour indices and whether they have been observed yet.
func (s *Self) AppearanceColours() (hair, top, trouser, skin int, ok bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.colourHair, s.colourTop, s.colourTrouser, s.colourSkin, s.hasColours
}

// MaxHP is a convenience accessor for the Hits skill's max value.
func (s *Self) MaxHP(...int) int { return s.SkillMax(3) }

// Prayer / MaxPrayer convenience accessors for skill index 5.
func (s *Self) Prayer() int    { return s.SkillLevel(5) }
func (s *Self) MaxPrayer() int { return s.SkillMax(5) }

// CombatLevel computes a rough combat level from current skills.
// The exact formula varies by mechanic; this is a heuristic.
func (s *Self) CombatLevel() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	atk := s.skillMax[0]
	def := s.skillMax[1]
	str := s.skillMax[2]
	hp := s.skillMax[3]
	ranged := s.skillMax[4]
	pray := s.skillMax[5]
	magic := s.skillMax[6]
	base := def + hp + pray/2
	melee := (atk + str) * 13 / 10
	rangedC := ranged * 13 / 8
	magicC := magic * 13 / 8
	best := melee
	if rangedC > best {
		best = rangedC
	}
	if magicC > best {
		best = magicC
	}
	return (base + best) / 4
}
