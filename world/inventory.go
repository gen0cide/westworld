package world

import "sync"

// InvSlot is one slot's contents.
type InvSlot struct {
	ItemID  int
	Amount  int
	Wielded bool
}

// Inventory tracks the host's inventory as the server reports it.
// The total slot count is 30 in RSC; we store up to that.
type Inventory struct {
	mu    sync.RWMutex
	slots []InvSlot // dynamically sized; can be 0..30
}

// NewInventory returns an empty inventory.
func NewInventory() *Inventory { return &Inventory{} }

// Replace replaces the entire inventory contents (used after a full
// snapshot from the server).
func (i *Inventory) Replace(slots []InvSlot) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.slots = append(i.slots[:0], slots...)
}

// Set updates one slot. If item is nil, the slot is cleared (set to
// zero values). Extends the slice if needed.
func (i *Inventory) Set(slot int, item *InvSlot) {
	i.mu.Lock()
	defer i.mu.Unlock()
	for len(i.slots) <= slot {
		i.slots = append(i.slots, InvSlot{})
	}
	if item == nil {
		i.slots[slot] = InvSlot{}
	} else {
		i.slots[slot] = *item
	}
}

// RemoveSlot deletes the item at slot and shifts every later slot down
// one, matching the server's ArrayList.remove(index) (RSC inventories
// have no holes). A burst of RemoveSlot(0) — e.g. ::wipeinv, which loops
// remove(get(0)) — therefore empties the whole list, where repeated
// Set(0, nil) would idempotently blank only slot 0 and strand the rest.
func (i *Inventory) RemoveSlot(slot int) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if slot < 0 || slot >= len(i.slots) {
		return
	}
	i.slots = append(i.slots[:slot], i.slots[slot+1:]...)
}

// Slots returns a snapshot of all slot contents.
func (i *Inventory) Slots() []InvSlot {
	i.mu.RLock()
	defer i.mu.RUnlock()
	out := make([]InvSlot, len(i.slots))
	copy(out, i.slots)
	return out
}

// Count returns the total quantity of items with the given ID across
// all slots.
func (i *Inventory) Count(itemID int) int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	total := 0
	for _, s := range i.slots {
		if s.ItemID == itemID {
			total += s.Amount
		}
	}
	return total
}

// FreeSlots returns the number of empty slots remaining (out of 30).
func (i *Inventory) FreeSlots() int {
	i.mu.RLock()
	defer i.mu.RUnlock()
	used := 0
	for _, s := range i.slots {
		if s.ItemID > 0 {
			used++
		}
	}
	return 30 - used
}

// Has returns true if at least one slot contains the given item.
func (i *Inventory) Has(itemID int) bool {
	return i.Count(itemID) > 0
}
