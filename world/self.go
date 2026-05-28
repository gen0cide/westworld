package world

import "sync"

// Coord is an RSC world tile coordinate. RSC uses big-endian uint16
// for both x and y on the wire. We keep them as ints here for ergonomic
// arithmetic.
type Coord struct {
	X int
	Y int
}

// NumSkills matches OpenRSC's 18-skill catalog (Attack..Thieving).
const NumSkills = 18

// Self is the host's view of its own player. Read/write safe across
// goroutines.
type Self struct {
	mu sync.RWMutex

	position    Coord
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

	// activePrayers mirrors SEND_PRAYERS_ACTIVE — one bool per
	// prayer slot, indexed 0..13 (Thick Skin, Burst of Strength,
	// Clarity of Thought, Rock Skin, Superhuman Strength, Improved
	// Reflexes, Rapid Restore, Rapid Heal, Protect Item, Steel
	// Skin, Ultimate Strength, Incredible Reflexes, Protect from
	// Magic, Protect from Missiles, Protect from Melee — 15
	// total in OpenRSC). Length matches whatever the server sends.
	activePrayers []bool
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
func (s *Self) Fatigue() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.fatigue
}

// SetFatigue updates fatigue from an inbound packet.
func (s *Self) SetFatigue(v int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.fatigue = v
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
