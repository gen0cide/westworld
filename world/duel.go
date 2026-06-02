package world

import (
	"sync"
	"time"
)

// DuelState mirrors the server's duel-window state for the host.
// Parallel to TradeState — RSC allows only one active duel at a time.
type DuelState struct {
	mu sync.RWMutex
	d  *DuelRecord
}

// DuelRecord is the snapshotted state of the active duel. nil if no
// duel is in progress.
//
// Duels follow the same two-screen handshake as trades — first an
// offer screen with stake items + rule toggles, then a confirm
// screen. The Phase field tracks UI state the same way.
type DuelRecord struct {
	// WithIndex is the server-side index of the opponent.
	WithIndex int
	// WithName is the opponent's username.
	WithName string
	// MyOffer / TheirOffer are the staked items each side has put
	// up. Replaced wholesale when either side updates.
	MyOffer    []TradeItem
	TheirOffer []TradeItem
	// First/Second accepted flags — same semantics as TradeRecord.
	MyFirstAccepted     bool
	TheirFirstAccepted  bool
	MySecondAccepted    bool
	TheirSecondAccepted bool
	// Rules are the duel rule toggles. Either side can change them
	// from the offer screen; the server unifies them and broadcasts
	// the new state to both. A change resets both first-accept flags.
	Rules DuelRules
	// Phase tracks the UI state: "request_sent", "open" (offer
	// screen), "confirm" (final review), "completed", "cancelled".
	Phase string
	// OpenedAt and UpdatedAt are wall-clock timestamps for debugging.
	OpenedAt  time.Time
	UpdatedAt time.Time
}

// DuelRules are the four pre-fight toggles. true = disallowed.
type DuelRules struct {
	DisallowRetreat bool
	DisallowMagic   bool
	DisallowPrayer  bool
	DisallowWeapons bool
}

// NewDuelState returns an empty duel state.
func NewDuelState() *DuelState { return &DuelState{} }

// Duel returns a snapshot of the current duel, or nil if no duel is
// active. Slices are deep-copied so callers can read without holding
// the lock.
func (s *DuelState) Duel() *DuelRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.d == nil {
		return nil
	}
	c := *s.d
	c.MyOffer = append([]TradeItem(nil), s.d.MyOffer...)
	c.TheirOffer = append([]TradeItem(nil), s.d.TheirOffer...)
	return &c
}

// IsActive returns true iff a duel is in a non-terminal phase.
func (s *DuelState) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.d == nil {
		return false
	}
	return s.d.Phase != "completed" && s.d.Phase != "cancelled"
}

// BeginRequest seeds the record after our routine sends a duel
// request. Phase advances to "open" on server confirmation.
func (s *DuelState) BeginRequest(withIndex int, withName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d = &DuelRecord{
		WithIndex: withIndex,
		WithName:  withName,
		Phase:     "request_sent",
		OpenedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

// MarkOpened transitions to the offer phase. Called when the server
// emits DuelOpened (SEND_DUEL_WINDOW).
func (s *DuelState) MarkOpened(withIndex int, withName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		s.d = &DuelRecord{OpenedAt: time.Now()}
	}
	s.d.WithIndex = withIndex
	if withName != "" {
		s.d.WithName = withName
	}
	s.d.Phase = "open"
	s.d.UpdatedAt = time.Now()
}

// SetMyOffer records the items WE are staking. Called after a
// successful outbound DUEL_OFFER_ITEM. Resets accepts (server rule).
func (s *DuelState) SetMyOffer(items []TradeItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.MyOffer = append([]TradeItem(nil), items...)
	s.d.MyFirstAccepted = false
	s.d.TheirFirstAccepted = false
	s.d.MySecondAccepted = false
	s.d.TheirSecondAccepted = false
	s.d.UpdatedAt = time.Now()
}

// SetTheirOffer records the items THEY are staking. Called on
// inbound DuelOtherOffer.
func (s *DuelState) SetTheirOffer(items []TradeItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.TheirOffer = append([]TradeItem(nil), items...)
	s.d.TheirFirstAccepted = false
	s.d.MyFirstAccepted = false
	s.d.UpdatedAt = time.Now()
}

// UpdateTheirOfferNoReset is for the confirm-window apply path where
// the server pushes the final items list but accepts must NOT be
// reset (we just transitioned to "confirm" precisely because both
// sides accepted).
func (s *DuelState) UpdateTheirOfferNoReset(items []TradeItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.TheirOffer = append([]TradeItem(nil), items...)
	s.d.UpdatedAt = time.Now()
}

// UpdateRulesNoReset is the no-reset variant for the confirm-window
// path.
func (s *DuelState) UpdateRulesNoReset(r DuelRules) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.Rules = r
	s.d.UpdatedAt = time.Now()
}

// SetRules records the rule toggles. Called on inbound DuelSettings
// (when the opponent changed them) OR after our routine sends
// DUEL_FIRST_SETTINGS_CHANGED. Resets first-accept flags.
func (s *DuelState) SetRules(r DuelRules) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.Rules = r
	s.d.MyFirstAccepted = false
	s.d.TheirFirstAccepted = false
	s.d.UpdatedAt = time.Now()
}

// MarkOtherFirstAccepted: opponent clicked Accept on the offer
// screen (inbound SEND_DUEL_OTHER_ACCEPTED). Advances to confirm once
// BOTH sides have first-accepted, regardless of arrival order (mirrors
// MarkMyFirstAccepted). The server also pushes SEND_DUEL_CONFIRMWINDOW
// (MarkConfirmShown) for this transition, but deriving it locally too
// keeps the phase correct if that push races or is missed — and matches
// the fix in world/trade.go, which has no server-pushed confirm event.
func (s *DuelState) MarkOtherFirstAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.TheirFirstAccepted = true
	if s.d.MyFirstAccepted {
		s.d.Phase = "confirm"
	}
	s.d.UpdatedAt = time.Now()
}

// MarkOtherSecondAccepted: opponent clicked Accept on the confirm
// screen. Inbound from SEND_DUEL_OTHER_ACCEPTED while phase=confirm.
func (s *DuelState) MarkOtherSecondAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.TheirSecondAccepted = true
	s.d.UpdatedAt = time.Now()
}

// MarkMyFirstAccepted is called after our routine sends
// DUEL_FIRST_ACCEPTED.
func (s *DuelState) MarkMyFirstAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.MyFirstAccepted = true
	if s.d.TheirFirstAccepted {
		s.d.Phase = "confirm"
	}
	s.d.UpdatedAt = time.Now()
}

// MarkConfirmShown is called when the server pushes
// SEND_DUEL_CONFIRMWINDOW — both sides have first-accepted and the
// confirm screen is up.
func (s *DuelState) MarkConfirmShown() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.Phase = "confirm"
	s.d.UpdatedAt = time.Now()
}

// MarkMySecondAccepted is called after our routine sends
// DUEL_SECOND_ACCEPTED on the confirm screen.
func (s *DuelState) MarkMySecondAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	s.d.MySecondAccepted = true
	s.d.UpdatedAt = time.Now()
}

// MarkClosed transitions to terminal state. completed=true => fight
// is about to start (or has started); false => declined/cancelled.
func (s *DuelState) MarkClosed(completed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d == nil {
		return
	}
	if completed {
		s.d.Phase = "completed"
	} else {
		s.d.Phase = "cancelled"
	}
	s.d.UpdatedAt = time.Now()
}

// Clear wipes any duel state.
func (s *DuelState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d = nil
}
