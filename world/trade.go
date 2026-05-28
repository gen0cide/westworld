package world

import (
	"sync"
	"time"
)

// TradeState mirrors the server's trade-window state for the host.
// One trade slot — RSC allows only one active trade at a time. All
// fields are protected by mu; readers should snapshot via the
// Trade() accessor below.
type TradeState struct {
	mu sync.RWMutex
	t  *TradeRecord
}

// TradeRecord is the snapshotted state of the active trade. nil if
// no trade is in progress.
type TradeRecord struct {
	// WithIndex is the server-side index of the other player.
	WithIndex int
	// WithName is the other player's username (resolved via
	// world.Players when the trade opens).
	WithName string
	// MyOffer / TheirOffer are the items each side has put up
	// for trade in the first-screen offer phase. Replaced wholesale
	// when either side updates their offer.
	MyOffer    []TradeItem
	TheirOffer []TradeItem
	// FirstAccepted flags fire when either side clicks the first
	// "Accept" button (the one on the item-offer screen). Both
	// must accept to proceed to the confirm screen.
	MyFirstAccepted    bool
	TheirFirstAccepted bool
	// SecondAccepted flags fire on the final confirm screen
	// click. Both must accept for the trade to complete.
	MySecondAccepted    bool
	TheirSecondAccepted bool
	// Phase tracks the UI state: "request_sent", "open" (offer
	// screen), "confirm" (final review), "completed", "cancelled".
	Phase string
	// OpenedAt and UpdatedAt are wall-clock timestamps for
	// debugging — routines can stale-check against time.Now().
	OpenedAt  time.Time
	UpdatedAt time.Time
}

// TradeItem is a single stake — item id + quantity.
type TradeItem struct {
	ItemID int
	Amount int
}

// NewTradeState returns an empty trade state.
func NewTradeState() *TradeState { return &TradeState{} }

// Trade returns a snapshot of the current trade, or nil if no
// trade is active. The returned struct is safe to read without
// holding the lock (it's a deep-ish copy — the slice headers are
// fresh; routines don't mutate them).
func (s *TradeState) Trade() *TradeRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.t == nil {
		return nil
	}
	c := *s.t
	c.MyOffer = append([]TradeItem(nil), s.t.MyOffer...)
	c.TheirOffer = append([]TradeItem(nil), s.t.TheirOffer...)
	return &c
}

// IsActive returns true iff a trade is currently in any phase
// other than completed/cancelled.
func (s *TradeState) IsActive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.t == nil {
		return false
	}
	return s.t.Phase != "completed" && s.t.Phase != "cancelled"
}

// BeginRequest seeds a new trade record after our routine sent
// a trade request. The phase advances to "open" when the server
// confirms via TradeOpened.
func (s *TradeState) BeginRequest(withIndex int, withName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.t = &TradeRecord{
		WithIndex: withIndex,
		WithName:  withName,
		Phase:     "request_sent",
		OpenedAt:  time.Now(),
		UpdatedAt: time.Now(),
	}
}

// MarkOpened transitions to the offer phase. Called when the
// server emits TradeOpened. Resolves WithName from world.Players
// if not already set.
func (s *TradeState) MarkOpened(withIndex int, withName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		s.t = &TradeRecord{OpenedAt: time.Now()}
	}
	s.t.WithIndex = withIndex
	if withName != "" {
		s.t.WithName = withName
	}
	s.t.Phase = "open"
	s.t.UpdatedAt = time.Now()
}

// SetMyOffer records the items WE are offering. Called after a
// successful outbound OfferTradeItems.
func (s *TradeState) SetMyOffer(items []TradeItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	s.t.MyOffer = append([]TradeItem(nil), items...)
	// Updating my offer un-accepts both sides (server rule).
	s.t.MyFirstAccepted = false
	s.t.TheirFirstAccepted = false
	s.t.MySecondAccepted = false
	s.t.TheirSecondAccepted = false
	s.t.UpdatedAt = time.Now()
}

// SetTheirOffer records the items THEY are offering. Called on
// inbound TradeOtherOffer event.
func (s *TradeState) SetTheirOffer(items []TradeItem) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	s.t.TheirOffer = append([]TradeItem(nil), items...)
	s.t.TheirFirstAccepted = false
	s.t.MyFirstAccepted = false
	s.t.UpdatedAt = time.Now()
}

// MarkOtherFirstAccepted is set when the server emits
// TradeOtherAccepted in the offer phase.
func (s *TradeState) MarkOtherFirstAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	s.t.TheirFirstAccepted = true
	s.t.UpdatedAt = time.Now()
}

// MarkMyFirstAccepted is called after our routine sends the
// first ConfirmTrade.
func (s *TradeState) MarkMyFirstAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	s.t.MyFirstAccepted = true
	if s.t.TheirFirstAccepted {
		s.t.Phase = "confirm"
	}
	s.t.UpdatedAt = time.Now()
}

// MarkMySecondAccepted is called after our routine sends the
// second ConfirmTrade on the final confirm screen.
func (s *TradeState) MarkMySecondAccepted() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	s.t.MySecondAccepted = true
	s.t.UpdatedAt = time.Now()
}

// MarkClosed transitions to terminal state. completed=true
// indicates a successful trade; completed=false means
// cancellation (decline or abort).
func (s *TradeState) MarkClosed(completed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.t == nil {
		return
	}
	if completed {
		s.t.Phase = "completed"
	} else {
		s.t.Phase = "cancelled"
	}
	s.t.UpdatedAt = time.Now()
}

// Clear wipes any trade state — used after both sides leave the
// terminal phase so the next trade starts cleanly.
func (s *TradeState) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.t = nil
}
