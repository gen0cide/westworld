package world

import (
	"sync"
	"time"
)

// BankState mirrors the bot's view of the bank window. Open is true
// while the window is up; Slots is the full bank contents (one entry
// per occupied slot).
type BankState struct {
	mu sync.RWMutex
	b  *BankRecord
}

// BankRecord is the snapshot of the open bank.
type BankRecord struct {
	MaxSize   int
	Slots     []BankSlot
	OpenedAt  time.Time
	UpdatedAt time.Time
}

// BankSlot is one occupied bank slot.
type BankSlot struct {
	ItemID int
	Amount int
}

// NewBankState returns an empty bank state.
func NewBankState() *BankState { return &BankState{} }

// Bank returns a snapshot of the current bank, or nil if no bank
// window is open.
func (s *BankState) Bank() *BankRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.b == nil {
		return nil
	}
	c := *s.b
	c.Slots = append([]BankSlot(nil), s.b.Slots...)
	return &c
}

// IsOpen returns true iff the bank window is currently open.
func (s *BankState) IsOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.b != nil
}

// Open seeds a fresh bank record from a SEND_BANK_OPEN packet.
func (s *BankState) Open(maxSize int, items []BankSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	s.b = &BankRecord{
		MaxSize:   maxSize,
		Slots:     append([]BankSlot(nil), items...),
		OpenedAt:  now,
		UpdatedAt: now,
	}
}

// UpdateSlot applies a SEND_BANK_UPDATE packet. amount=0 means the
// slot was emptied; otherwise the slot is set/created in-place.
//
// OpenRSC sends updates by slot INDEX (not item ID), so we mutate by
// index. If the bank is closed when this fires, the update is
// ignored (defensive — shouldn't happen).
func (s *BankState) UpdateSlot(slot, itemID, amount int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.b == nil {
		return
	}
	s.b.UpdatedAt = time.Now()
	if amount == 0 {
		// Removed — drop the slot if it exists.
		if slot >= 0 && slot < len(s.b.Slots) {
			s.b.Slots = append(s.b.Slots[:slot], s.b.Slots[slot+1:]...)
		}
		return
	}
	// Grow if necessary.
	for len(s.b.Slots) <= slot {
		s.b.Slots = append(s.b.Slots, BankSlot{})
	}
	s.b.Slots[slot] = BankSlot{ItemID: itemID, Amount: amount}
}

// Close wipes the bank state.
func (s *BankState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.b = nil
}

// Has returns the total amount of itemID held across all slots.
// Returns 0 if the bank is closed or the item isn't present.
func (s *BankState) Has(itemID int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.b == nil {
		return 0
	}
	total := 0
	for _, sl := range s.b.Slots {
		if sl.ItemID == itemID {
			total += sl.Amount
		}
	}
	return total
}
