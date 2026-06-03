package world

import (
	"strings"
	"sync"
)

// Friend is one entry of the host's friend roster, mirrored from the server's
// SEND_FRIEND_UPDATE (opcode 149) packets. Online=false means offline (or not
// yet reported); World is the friend's server/world name when online.
type Friend struct {
	Name       string
	FormerName string
	World      string
	Online     bool
}

// SocialState mirrors the server-authoritative friend + ignore rosters for the
// host. The server sends a SEND_FRIEND_UPDATE (149) burst at login and one per
// status change, and a full SEND_IGNORE_LIST (109) whenever the ignore list
// changes — so the roster here is kept in sync by World.Apply. Removing a friend
// produces NO server packet, so the HTTP layer calls RemoveFriend after a
// successful outbound remove.
type SocialState struct {
	mu          sync.RWMutex
	friends     map[string]*Friend // key: socialKey(name)
	friendOrder []string           // display order (original-case names)
	ignores     []string
}

// NewSocialState returns an empty roster.
func NewSocialState() *SocialState {
	return &SocialState{friends: map[string]*Friend{}}
}

func socialKey(s string) string { return strings.ToLower(strings.TrimSpace(s)) }

// ApplyFriendUpdate inserts or updates one friend from a 149 packet. When the
// packet is a rename (rename=true, formerName set), the entry previously keyed
// by formerName is re-keyed to the new name in place.
func (s *SocialState) ApplyFriendUpdate(name, formerName, world string, online, rename bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := socialKey(name)
	if k == "" {
		return
	}
	if rename && formerName != "" {
		if fk := socialKey(formerName); fk != k {
			s.dropLocked(fk)
		}
	}
	f, ok := s.friends[k]
	if !ok {
		f = &Friend{}
		s.friends[k] = f
		s.friendOrder = append(s.friendOrder, name)
	}
	f.Name = name
	f.FormerName = formerName
	f.World = world
	f.Online = online
}

// RemoveFriend drops a friend from the local roster (the server sends no packet
// on remove). No-op if absent.
func (s *SocialState) RemoveFriend(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dropLocked(socialKey(name))
}

// dropLocked removes key k from the map + order slice. Caller holds s.mu.
func (s *SocialState) dropLocked(k string) {
	if k == "" {
		return
	}
	if _, ok := s.friends[k]; !ok {
		return
	}
	delete(s.friends, k)
	out := s.friendOrder[:0]
	for _, n := range s.friendOrder {
		if socialKey(n) != k {
			out = append(out, n)
		}
	}
	s.friendOrder = out
}

// SetIgnores replaces the ignore list from a 109 bulk packet.
func (s *SocialState) SetIgnores(names []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ignores = append([]string(nil), names...)
}

// RemoveIgnore drops a name from the local ignore list. The server re-sends the
// full ignore list (109) on ADD but NOT on remove — its SOCIAL_REMOVE_IGNORE
// handler is silent — so the HTTP layer mirrors the removal here after a
// successful outbound remove (matching the friend-remove path). Case-insensitive.
func (s *SocialState) RemoveIgnore(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := socialKey(name)
	out := s.ignores[:0]
	for _, n := range s.ignores {
		if socialKey(n) != k {
			out = append(out, n)
		}
	}
	s.ignores = out
}

// Friends returns a display-ordered snapshot of the friend roster.
func (s *SocialState) Friends() []Friend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Friend, 0, len(s.friendOrder))
	for _, n := range s.friendOrder {
		if f := s.friends[socialKey(n)]; f != nil {
			out = append(out, *f)
		}
	}
	return out
}

// Ignores returns a snapshot of the ignore list (display names).
func (s *SocialState) Ignores() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.ignores...)
}
