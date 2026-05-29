package world

import (
	"sync"
	"time"
)

// ShopState mirrors the bot's view of the shop window — the same
// open/snapshot/close pattern as BankState/TradeState. One shop slot
// at a time (RSC, like the bank, only ever has one shop window open).
//
// The mirror is populated from the inbound SEND_SHOP_OPEN packet
// (opcode 101). RSC has no per-slot shop-update packet: the server
// re-sends the FULL shop list on every stock change, so Open() is the
// single mutation entry point (re-Open replaces the snapshot). Close()
// fires on SEND_SHOP_CLOSE (opcode 137).
type ShopState struct {
	mu sync.RWMutex
	s  *ShopRecord
}

// ShopRecord is the snapshot of the open shop. Field names mirror the
// SEND_SHOP_OPEN wire layout decoded in proto/v235 (see decodeShopOpen)
// and the OpenRSC ShopStruct (outgoing).
type ShopRecord struct {
	// IsGeneral is true for a general store (buys/sells anything the
	// player carries); false for a specialty shop (fixed catalogue).
	// From the wire `isGeneralStore` byte.
	IsGeneral bool
	// SellPriceMod / BuyPriceMod are the shop's base price-percentage
	// modifiers (RSC `sellModifier` / `buyModifier`). The displayed
	// unit price is basePrice * (mod + stockOffset) / 100. BuyPriceMod
	// is the percentage the player PAYS when buying; SellPriceMod is
	// the (lower) percentage the player RECEIVES when selling.
	SellPriceMod int
	BuyPriceMod  int
	// PriceMultiplier (RSC `stockSensitivity`) scales how far the
	// per-item price drifts from the modifier as stock deviates from
	// the base amount. 0 on a fixed-price shop.
	PriceMultiplier int
	// Slots is the shop's catalogue, one entry per stocked item.
	Slots     []ShopSlot
	OpenedAt  time.Time
	UpdatedAt time.Time
}

// ShopSlot is one item in the shop catalogue.
type ShopSlot struct {
	// ItemID is the catalogue (def) id.
	ItemID int
	// Stock is the quantity currently available to buy (0 = out of
	// stock; a general store may show 0-stock player items it will
	// buy).
	Stock int
	// BaseStock is the shop's baseline stock for this item (RSC
	// `baseAmount`, client `shopItemPrice`). Used as the reference
	// point for the stock-sensitive price drift, NOT a gp value.
	BaseStock int
}

// NewShopState returns an empty shop state.
func NewShopState() *ShopState { return &ShopState{} }

// Shop returns a snapshot of the current shop, or nil if no shop
// window is open. The returned struct is safe to read without the
// lock (slices are freshly copied).
func (s *ShopState) Shop() *ShopRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.s == nil {
		return nil
	}
	c := *s.s
	c.Slots = append([]ShopSlot(nil), s.s.Slots...)
	return &c
}

// IsOpen returns true iff a shop window is currently open.
func (s *ShopState) IsOpen() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.s != nil
}

// Open seeds (or replaces) the shop record from a SEND_SHOP_OPEN
// packet. RSC re-sends the whole list on every stock change, so a
// re-Open while a shop is already up is the normal stock-update path
// (it replaces the snapshot wholesale).
func (s *ShopState) Open(isGeneral bool, sellMod, buyMod, priceMult int, slots []ShopSlot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	openedAt := now
	if s.s != nil {
		// Preserve the original open time across stock-update re-sends.
		openedAt = s.s.OpenedAt
	}
	s.s = &ShopRecord{
		IsGeneral:       isGeneral,
		SellPriceMod:    sellMod,
		BuyPriceMod:     buyMod,
		PriceMultiplier: priceMult,
		Slots:           append([]ShopSlot(nil), slots...),
		OpenedAt:        openedAt,
		UpdatedAt:       now,
	}
}

// Close wipes the shop state (SEND_SHOP_CLOSE, or we sent shop.close()).
func (s *ShopState) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = nil
}

// Stock returns the quantity of itemID currently in stock. Returns 0
// if the shop is closed or the item isn't stocked. General stores may
// list 0-stock player items they'll buy.
func (s *ShopState) Stock(itemID int) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.s == nil {
		return 0
	}
	total := 0
	for _, sl := range s.s.Slots {
		if sl.ItemID == itemID {
			total += sl.Stock
		}
	}
	return total
}

// slot returns a pointer-by-value copy of the slot for itemID and
// whether it exists. Caller holds no lock requirement (RLock taken
// here). Used by the price accessor.
func (s *ShopState) slot(itemID int) (ShopSlot, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.s == nil {
		return ShopSlot{}, false
	}
	for _, sl := range s.s.Slots {
		if sl.ItemID == itemID {
			return sl, true
		}
	}
	return ShopSlot{}, false
}

// BuyPrice computes the unit BUY price (gp) the client would display
// for itemID, given the item's catalogue base price. This mirrors the
// authentic client formula (GenUtil.computeItemCost, single unit):
//
//	offset  = clamp(priceMultiplier * (baseStock - currentStock), -100, 100)
//	scaling = max(buyPriceMod + offset, 10)
//	price   = basePrice * scaling / 100
//
// basePrice comes from the caller (runtime resolves it via
// facts.ItemDef.BasePrice). Returns 0 if the shop is closed or the
// item isn't stocked.
func (s *ShopState) BuyPrice(itemID, basePrice int) int {
	sl, ok := s.slot(itemID)
	if !ok {
		return 0
	}
	s.mu.RLock()
	buyMod := 0
	priceMult := 0
	if s.s != nil {
		buyMod = s.s.BuyPriceMod
		priceMult = s.s.PriceMultiplier
	}
	s.mu.RUnlock()
	return unitShopPrice(basePrice, buyMod, priceMult, sl.BaseStock, sl.Stock)
}

// SellPrice computes the unit SELL price (gp) — what the shop pays the
// player for one unit. Same formula as BuyPrice but using the shop's
// sellPriceMod (lower than buyPriceMod). Per api.md the public API only
// exposes buy price, but the sell price is cheap to derive and useful
// for routines deciding whether a sale is worth it.
func (s *ShopState) SellPrice(itemID, basePrice int) int {
	sl, ok := s.slot(itemID)
	if !ok {
		return 0
	}
	s.mu.RLock()
	sellMod := 0
	priceMult := 0
	if s.s != nil {
		sellMod = s.s.SellPriceMod
		priceMult = s.s.PriceMultiplier
	}
	s.mu.RUnlock()
	return unitShopPrice(basePrice, sellMod, priceMult, sl.BaseStock, sl.Stock)
}

// unitShopPrice is the shared single-unit price kernel — see BuyPrice.
func unitShopPrice(basePrice, priceMod, priceMult, baseStock, currentStock int) int {
	offset := priceMult * (baseStock - currentStock)
	if offset < -100 {
		offset = -100
	} else if offset > 100 {
		offset = 100
	}
	scaling := priceMod + offset
	if scaling < 10 {
		scaling = 10
	}
	return basePrice * scaling / 100
}
