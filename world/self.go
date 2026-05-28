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
