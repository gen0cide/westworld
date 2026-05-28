package world

import (
	"sync"
	"time"

	"github.com/gen0cide/westworld/event"
)

// World is the composition of all per-host state mirrors. It's the
// single object higher layers consult to ask "what does my agent
// believe about the world right now?"
type World struct {
	Self        *Self
	Inventory   *Inventory
	Npcs        *NpcsState
	Players     *PlayersState
	GroundItems *GroundItemsState
	Recent      *RecentEvents
	Trade       *TradeState
	Duel        *DuelState
	Bank        *BankState
}

// NewWorld returns a freshly-initialized World with all sub-mirrors
// ready (but empty until inbound packets populate them).
func NewWorld() *World {
	return &World{
		Self:        NewSelf(),
		Inventory:   NewInventory(),
		Npcs:        NewNpcsState(),
		Players:     NewPlayersState(),
		GroundItems: NewGroundItemsState(),
		Recent:      NewRecentEvents(),
		Trade:       NewTradeState(),
		Duel:        NewDuelState(),
		Bank:        NewBankState(),
	}
}

// GroundItemRecord is one ground item visible in the bot's local view.
// Keyed by absolute world (X, Y) — multiple ground items can share a
// tile, but for Phase 1 we assume one item per tile.
type GroundItemRecord struct {
	X, Y     int
	ItemID   int
	LastSeen time.Time
}

// GroundItemsState mirrors the ground items currently visible to the
// bot. Updated by world.Apply on inbound GroundItemEvent packets;
// emits/removes entries as items appear and disappear from view.
type GroundItemsState struct {
	mu sync.RWMutex
	m  map[[2]int]GroundItemRecord
}

func NewGroundItemsState() *GroundItemsState {
	return &GroundItemsState{m: map[[2]int]GroundItemRecord{}}
}

func (s *GroundItemsState) Add(x, y, itemID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[[2]int{x, y}] = GroundItemRecord{X: x, Y: y, ItemID: itemID, LastSeen: time.Now()}
}

func (s *GroundItemsState) Remove(x, y int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, [2]int{x, y})
}

// All returns a snapshot of every currently-visible ground item.
func (s *GroundItemsState) All() []GroundItemRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]GroundItemRecord, 0, len(s.m))
	for _, r := range s.m {
		out = append(out, r)
	}
	return out
}

// Near returns ground items within `radius` Chebyshev tiles of
// (cx, cy).
func (s *GroundItemsState) Near(cx, cy, radius int) []GroundItemRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []GroundItemRecord
	for _, r := range s.m {
		dx := r.X - cx
		if dx < 0 {
			dx = -dx
		}
		dy := r.Y - cy
		if dy < 0 {
			dy = -dy
		}
		if dx <= radius && dy <= radius {
			out = append(out, r)
		}
	}
	return out
}

// NpcRecord is what we last saw of one NPC: its server-assigned index,
// world position, and our last-seen timestamp.
type NpcRecord struct {
	Index    int
	X, Y     int
	TypeID   int
	LastSeen time.Time
}

// NpcsState tracks every NPC the bot has perceived, keyed by server
// index. Indices reshuffle when an NPC despawns/respawns; stale entries
// are pruned by Touch (we re-stamp every entry the server resends each
// tick).
type NpcsState struct {
	mu sync.RWMutex
	m  map[int]NpcRecord
}

func NewNpcsState() *NpcsState { return &NpcsState{m: map[int]NpcRecord{}} }

func (s *NpcsState) Set(rec NpcRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	rec.LastSeen = time.Now()
	s.m[rec.Index] = rec
}

func (s *NpcsState) Get(index int) (NpcRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.m[index]
	return r, ok
}

// All returns a snapshot of every known NPC record.
func (s *NpcsState) All() []NpcRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]NpcRecord, 0, len(s.m))
	for _, r := range s.m {
		out = append(out, r)
	}
	return out
}

// PlayerRecord is what we last saw of one nearby player.
type PlayerRecord struct {
	Index    int
	Name     string // may be empty until we get an appearance update
	X, Y     int
	LastSeen time.Time
}

type PlayersState struct {
	mu sync.RWMutex
	m  map[int]PlayerRecord
}

func NewPlayersState() *PlayersState { return &PlayersState{m: map[int]PlayerRecord{}} }

func (s *PlayersState) SetPosition(index, x, y int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.X = x
	r.Y = y
	r.LastSeen = time.Now()
	s.m[index] = r
}

func (s *PlayersState) SetName(index int, name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.Name = name
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[index] = r
}

func (s *PlayersState) Remove(index int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, index)
}

func (s *PlayersState) Get(index int) (PlayerRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	r, ok := s.m[index]
	return r, ok
}

// All returns a snapshot of every known player record.
func (s *PlayersState) All() []PlayerRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]PlayerRecord, 0, len(s.m))
	for _, r := range s.m {
		out = append(out, r)
	}
	return out
}

// FindByName returns the most-recently-seen player record whose name
// case-insensitively matches `name`. Returns (zero, false) if not found.
func (s *PlayersState) FindByName(name string) (PlayerRecord, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var best PlayerRecord
	var found bool
	for _, r := range s.m {
		if equalFold(r.Name, name) {
			if !found || r.LastSeen.After(best.LastSeen) {
				best = r
				found = true
			}
		}
	}
	return best, found
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// Apply updates the world state based on an inbound event. Returns
// true if the world was meaningfully changed (caller may use this for
// metrics or change-detection).
//
// Called by the runtime layer for every decoded inbound packet, BEFORE
// the event is published to subscribers. Subscribers therefore see a
// world state already consistent with the event they're being told
// about.
func (w *World) Apply(ev event.Event) bool {
	switch e := ev.(type) {
	case event.StatUpdate:
		w.Self.SetSkill(int(e.Skill), e.Current, e.Max, e.Experience)
		return true
	case event.StatsSnapshot:
		w.Self.SetAllSkills(snapshotSkills(e.Current), snapshotSkills(e.Max), snapshotSkills(e.Experience), e.QuestPoints)
		return true
	case event.ExperienceGain:
		// Only XP changed; preserve current/max.
		s := int(e.Skill)
		w.Self.SetSkill(s, w.Self.SkillLevel(s), w.Self.SkillMax(s), e.XP)
		return true
	case event.FatigueUpdate:
		w.Self.SetFatigue(e.Value)
		return true
	case event.InventorySnapshot:
		slots := make([]InvSlot, len(e.Items))
		for i, it := range e.Items {
			slots[i] = InvSlot{ItemID: it.ItemID, Amount: it.Amount, Wielded: it.Wielded}
		}
		w.Inventory.Replace(slots)
		return true
	case event.InventorySlotUpdate:
		if e.Item == nil {
			w.Inventory.Set(e.Slot, nil)
		} else {
			w.Inventory.Set(e.Slot, &InvSlot{
				ItemID:  e.Item.ItemID,
				Amount:  e.Item.Amount,
				Wielded: e.Item.Wielded,
			})
		}
		return true
	case event.OwnPositionUpdate:
		w.Self.SetPosition(Coord{X: e.X, Y: e.Y})
		return true
	case event.Death:
		// Snapshot the current position as the death spot BEFORE
		// the respawn position packet (which arrives in the same
		// tick) overwrites self.position. Routines reading
		// self.last_death_at after `on death` fires see the death
		// tile, useful for "walk back to where I died" recovery.
		w.Self.RecordDeath()
		return true
	case event.NpcNearby:
		w.Npcs.Set(NpcRecord{Index: e.Index, X: e.X, Y: e.Y, TypeID: e.TypeID})
		return true
	case event.NearbyPlayerEvent:
		if e.Removed {
			w.Players.Remove(e.Index)
		} else {
			w.Players.SetPosition(e.Index, e.X, e.Y)
		}
		return true
	case event.OtherPlayerAppearance:
		w.Players.SetName(e.PlayerIndex, e.Name)
		return true
	case event.ChatReceived:
		w.Recent.SetChat(e.Speaker, e.Message)
		return true
	case event.OtherPlayerChat:
		// Resolve speaker name via the player index; the
		// appearance event preceding chat seeds Players.
		// If unresolved (rare — appearance precedes chat in
		// normal flow), fall back to a placeholder so routines
		// reading world.last_chat.speaker still get something
		// stable.
		speaker := ""
		if rec, ok := w.Players.Get(e.PlayerIndex); ok {
			speaker = rec.Name
		}
		w.Recent.SetChat(speaker, e.MessageText)
		return true
	case event.PrivateMessage:
		w.Recent.SetPM(e.Sender, e.Message)
		return true
	case event.SystemMessage:
		w.Recent.SetServerMessage(e.Message)
		return true
	case event.NpcDialogText:
		w.Recent.SetDialogText(e.Text)
		return true
	case event.NpcDialog:
		// Server presented a dialog options menu. Stored in
		// Recent.DialogOptions until a routine calls answer(N) +
		// ClearDialogOptions(), or until a new menu replaces it.
		w.Recent.SetDialogOptions(e.Options)
		return true
	case event.TradeOpened:
		// Trade window is now active. Resolve the other side's
		// name from PlayersState if we know it (we usually do —
		// trade initiation requires visibility).
		name := ""
		if rec, ok := w.Players.Get(e.OtherPlayerIndex); ok {
			name = rec.Name
		}
		w.Trade.MarkOpened(e.OtherPlayerIndex, name)
		return true
	case event.TradeOtherAccepted:
		w.Trade.MarkOtherFirstAccepted()
		return true
	case event.TradeClosed:
		// The protocol's SEND_TRADE_CLOSE has no completion bit, so
		// the decoder always sets Completed=false. Recover the true
		// outcome from the trade record's accept flags: if WE
		// reached MySecondAccepted before the close arrived, the
		// server was about to (or just did) move items — count that
		// as success. Otherwise it's a cancellation.
		completed := e.Completed
		if !completed {
			if rec := w.Trade.Trade(); rec != nil && rec.MySecondAccepted && rec.TheirSecondAccepted {
				completed = true
			} else if rec != nil && rec.MySecondAccepted {
				// We confirmed twice but never observed their
				// second accept — server still completed it after
				// our final click (the other side's accept might
				// not have surfaced as an opcode in this version).
				// Trust our own state.
				completed = true
			}
		}
		w.Trade.MarkClosed(completed)
		return true
	case event.TradeOtherOffer:
		items := make([]TradeItem, len(e.Items))
		for i, it := range e.Items {
			items[i] = TradeItem{ItemID: it.ItemID, Amount: it.Amount}
		}
		w.Trade.SetTheirOffer(items)
		return true
	case event.TradeConfirmShown:
		// Server moved both sides to the final review screen. Update
		// items to the canonical view + transition phase.
		items := make([]TradeItem, len(e.OpponentItems))
		for i, it := range e.OpponentItems {
			items[i] = TradeItem{ItemID: it.ItemID, Amount: it.Amount}
		}
		w.Trade.SetTheirOffer(items)
		// Force phase to "confirm" — both sides MUST have
		// first-accepted to reach this packet, so MarkMyFirstAccepted
		// is implied (server transitioned).
		if rec := w.Trade.Trade(); rec != nil && rec.Phase != "confirm" {
			w.Trade.MarkMyFirstAccepted()
		}
		return true
	case event.DuelOpened:
		name := ""
		if rec, ok := w.Players.Get(e.OtherPlayerIndex); ok {
			name = rec.Name
		}
		w.Duel.MarkOpened(e.OtherPlayerIndex, name)
		return true
	case event.DuelOtherOffer:
		items := make([]TradeItem, len(e.Items))
		for i, it := range e.Items {
			items[i] = TradeItem{ItemID: it.ItemID, Amount: it.Amount}
		}
		w.Duel.SetTheirOffer(items)
		return true
	case event.DuelSettingsUpdate:
		w.Duel.SetRules(DuelRules{
			DisallowRetreat: e.DisallowRetreat,
			DisallowMagic:   e.DisallowMagic,
			DisallowPrayer:  e.DisallowPrayer,
			DisallowWeapons: e.DisallowWeapons,
		})
		return true
	case event.DuelOtherAccepted:
		// Differentiate by current phase: offer → first-accept,
		// confirm → second-accept.
		if d := w.Duel.Duel(); d != nil && d.Phase == "confirm" {
			w.Duel.MarkOtherSecondAccepted()
		} else {
			w.Duel.MarkOtherFirstAccepted()
		}
		return true
	case event.DuelConfirmShown:
		// Server pushed the final review screen — both sides
		// first-accepted. Move state to "confirm" and update items/
		// rules to the server-canonical view. Critically, do NOT
		// reset accepts — we ARE here because both already accepted.
		items := make([]TradeItem, len(e.OpponentItems))
		for i, it := range e.OpponentItems {
			items[i] = TradeItem{ItemID: it.ItemID, Amount: it.Amount}
		}
		w.Duel.UpdateTheirOfferNoReset(items)
		w.Duel.UpdateRulesNoReset(DuelRules{
			DisallowRetreat: e.DisallowRetreat,
			DisallowMagic:   e.DisallowMagic,
			DisallowPrayer:  e.DisallowPrayer,
			DisallowWeapons: e.DisallowWeapons,
		})
		w.Duel.MarkConfirmShown()
		return true
	case event.BankOpened:
		slots := make([]BankSlot, len(e.Items))
		for i, it := range e.Items {
			slots[i] = BankSlot{ItemID: it.ItemID, Amount: it.Amount}
		}
		w.Bank.Open(e.MaxSize, slots)
		return true
	case event.BankSlotUpdate:
		w.Bank.UpdateSlot(e.Slot, e.ItemID, e.Amount)
		return true
	case event.BankClosed:
		w.Bank.Close()
		return true
	case event.DuelClosed:
		// SEND_DUEL_CLOSE (opcode 225) has no completion bit; infer
		// from the duel record. If we reached MySecondAccepted, the
		// server is closing the offer window because the fight is
		// commencing (not because someone cancelled).
		completed := e.Completed
		if !completed {
			if rec := w.Duel.Duel(); rec != nil && rec.MySecondAccepted {
				completed = true
			}
		}
		w.Duel.MarkClosed(completed)
		return true
	case event.OtherPlayerDamage:
		// Damage to ANY player gets recorded if it's us. The host's
		// own player index is tracked via OwnPositionUpdate; for now
		// record every damage we see and let routines filter by
		// source. Future: cross-check with our own player index.
		w.Recent.SetDamage(e.Damage, "")
		return true
	case event.GroundItemEvent:
		// GroundItemEvent offsets are relative to the player at the
		// time of packet delivery. We need the player's CURRENT
		// position to compute the absolute tile.
		pos := w.Self.Position()
		x := pos.X + e.OffsetX
		y := pos.Y + e.OffsetY
		if e.Disappear {
			w.GroundItems.Remove(x, y)
		} else {
			w.GroundItems.Add(x, y, e.ItemID)
		}
		return true
	}
	return false
}

// snapshotSkills converts a slice-like [18]int from event types to
// the same shape stored on Self.
func snapshotSkills(in [18]int) [NumSkills]int {
	var out [NumSkills]int
	for i := 0; i < NumSkills && i < len(in); i++ {
		out[i] = in[i]
	}
	return out
}
