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
	Boundaries  *DynamicBoundaries
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
		Boundaries:  NewDynamicBoundaries(),
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
//
// The hits / engagement fields exist so the #117 Npc.health accessor
// has somewhere to land. Note on provenance: opcode 234
// (InSendUpdatePlayers) only carries NPC data indirectly — when a
// player fires a projectile AT an NPC (type-3), the victim NPC index
// is recorded here as IncomingFromPlayerIndex, giving "this NPC is
// being attacked by player X". An NPC's own current/max hitpoints ride
// in the separate SEND_UPDATE_NPC packet (opcode 104, type-2 damage),
// which is not yet decoded; CurHits/MaxHits are reserved for that
// follow-up and stay HasHits=false until it lands.
type NpcRecord struct {
	Index    int
	X, Y     int
	TypeID   int
	LastSeen time.Time

	// --- combat state ---

	// CurHits / MaxHits / LastDamage: the NPC's health as cur/max
	// hitpoints. Reserved for the opcode-104 SEND_UPDATE_NPC type-2
	// decoder (not part of opcode 234). HasHits gates reads.
	CurHits      int
	MaxHits      int
	LastDamage   int
	HasHits      bool
	LastDamageAt time.Time

	// IncomingFromPlayerIndex: the index of a player who most recently
	// fired a projectile at this NPC (opcode 234 type-3). -1 / zero
	// time means "no observed incoming attack". Lets the engine answer
	// "is this NPC being fought, and by whom".
	IncomingFromPlayerIndex int
	IncomingProjectileID    int
	IncomingAt              time.Time
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

// Set records an NPC's position/type (the per-tick opcode-79 update).
// It PRESERVES any combat state already accumulated for that index
// from other packets (damage / incoming-attack updates), since the
// position update carries none of it and would otherwise clobber it.
func (s *NpcsState) Set(rec NpcRecord) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if prev, ok := s.m[rec.Index]; ok {
		// Carry forward combat fields the position update doesn't set.
		rec.CurHits = prev.CurHits
		rec.MaxHits = prev.MaxHits
		rec.LastDamage = prev.LastDamage
		rec.HasHits = prev.HasHits
		rec.LastDamageAt = prev.LastDamageAt
		rec.IncomingFromPlayerIndex = prev.IncomingFromPlayerIndex
		rec.IncomingProjectileID = prev.IncomingProjectileID
		rec.IncomingAt = prev.IncomingAt
	}
	rec.LastSeen = time.Now()
	s.m[rec.Index] = rec
}

// SetHits records an NPC's current/max hitpoints (and the damage that
// produced them). Reserved for the opcode-104 SEND_UPDATE_NPC type-2
// decoder; creates the record if the NPC isn't known yet.
func (s *NpcsState) SetHits(index, damage, curHits, maxHits int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.LastDamage = damage
	r.CurHits = curHits
	r.MaxHits = maxHits
	r.HasHits = true
	r.LastDamageAt = time.Now()
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[index] = r
}

// SetIncomingAttack records that a player fired a projectile at this
// NPC (opcode 234 type-3) — "this NPC is being attacked by player
// fromPlayerIndex". Creates the record if the NPC isn't known yet.
func (s *NpcsState) SetIncomingAttack(npcIndex, fromPlayerIndex, projectileID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[npcIndex]
	r.Index = npcIndex
	r.IncomingFromPlayerIndex = fromPlayerIndex
	r.IncomingProjectileID = projectileID
	r.IncomingAt = time.Now()
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[npcIndex] = r
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
//
// The Combat* / hits / engagement fields are decoded from inbound
// opcode 234 (InSendUpdatePlayers) and exist so the #117 combat
// perception accessors (combat.target / in_combat / target health)
// have real data to read. They are populated lazily — a record only
// carries combat data once the relevant 234 sub-update arrives for
// that player; until then the zero value + the Has* flags signal
// "unknown".
type PlayerRecord struct {
	Index    int
	Name     string // may be empty until we get an appearance update
	X, Y     int
	LastSeen time.Time

	// --- combat state (opcode 234) ---

	// CombatLevel + SkullType come from the type-5 appearance update's
	// two trailing bytes. SkullType is the per-player combat flag:
	// 0 = no skull, 1 = skulled / PK-flagged (recent player-vs-player
	// combat in the wilderness). HasAppearanceCombat is true once those
	// bytes have been seen for this player.
	CombatLevel         int
	SkullType           int
	HasAppearanceCombat bool

	// EquipBySlot is the per-slot worn-equipment SPRITE ids decoded from
	// the type-5 appearance update's equipment block, indexed by
	// event.EquipSlot* (head, shirt, pants, shield, weapon, hat, body,
	// legs, gloves, boots, amulet, cape). NOTE: these are appearance
	// SPRITE ids (itemDef.getAppearanceId() & 0xFF), NOT catalogue item
	// ids — the wire carries no item-id-by-slot. A zero slot means
	// nothing worn there. HasEquip is true once the block was decoded.
	// This data exists for self too: the server's appearance update for
	// our own player rides in the same opcode-234 type-5 record.
	EquipBySlot [event.NumEquipSlots]int
	HasEquip    bool

	// CurHits / MaxHits / LastDamage come from the type-2 damage
	// update ([byte damage][byte curHits][byte maxHits]). This is the
	// engaged target's health exactly as the wire encodes it (current
	// vs max hitpoints, not a fraction). HasHits is true once a damage
	// update has been observed; LastDamageAt stamps the most recent one
	// so callers can decide whether the health reading is still fresh.
	CurHits      int
	MaxHits      int
	LastDamage   int
	HasHits      bool
	LastDamageAt time.Time

	// Engagement (who-is-fighting-whom) comes from the type-3/4
	// projectile updates: this player fired ProjectileID at the entity
	// identified by EngagedNpcIndex (type 3) or EngagedPlayerIndex
	// (type 4). The unused side is -1. EngagedAt stamps the firing so
	// stale engagements can be aged out. (Melee combat in the authentic
	// v235 protocol surfaces only as type-2 damage on both
	// participants, not as a projectile, so an empty engagement does
	// NOT imply "not fighting".)
	EngagedNpcIndex    int
	EngagedPlayerIndex int
	ProjectileID       int
	EngagedAt          time.Time
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

// SetAppearanceCombat records the combat-state bytes from a type-5
// appearance update (opcode 234): the player's combat level and skull
// type (0=none, 1=skulled / PK-flagged). Preserves position/name like
// the other PlayersState setters.
func (s *PlayersState) SetAppearanceCombat(index, combatLevel, skullType int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.CombatLevel = combatLevel
	r.SkullType = skullType
	r.HasAppearanceCombat = true
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[index] = r
}

// SetWornEquipment records the per-slot worn-equipment sprite ids from
// a type-5 appearance update (opcode 234). worn is indexed by
// event.EquipSlot*; count is how many slot bytes the wire carried.
// Preserves position/name/combat like the other PlayersState setters.
func (s *PlayersState) SetWornEquipment(index int, worn [event.NumEquipSlots]int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.EquipBySlot = worn
	r.HasEquip = true
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[index] = r
}

// SetHits records a player's current/max hitpoints (and the damage
// that produced them) from a type-2 damage update (opcode 234). This
// is the engaged target's health as the wire encodes it — cur/max
// hitpoints, not a fraction.
func (s *PlayersState) SetHits(index, damage, curHits, maxHits int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.LastDamage = damage
	r.CurHits = curHits
	r.MaxHits = maxHits
	r.HasHits = true
	r.LastDamageAt = time.Now()
	if r.LastSeen.IsZero() {
		r.LastSeen = time.Now()
	}
	s.m[index] = r
}

// SetEngagement records that this player fired a projectile at a
// target (opcode 234 type-3/4). Exactly one of npcIndex / playerIndex
// is the real victim; pass -1 for the unused side.
func (s *PlayersState) SetEngagement(index, projectileID, npcIndex, playerIndex int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.m[index]
	r.Index = index
	r.ProjectileID = projectileID
	r.EngagedNpcIndex = npcIndex
	r.EngagedPlayerIndex = playerIndex
	r.EngagedAt = time.Now()
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
	case event.InventoryRemoveSlot:
		w.Inventory.RemoveSlot(e.Slot)
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
	case event.NpcDamage:
		// Opcode 104 type-2: the NPC's OWN current/max hitpoints. This
		// un-stubs Npc.health for ANY visible NPC (not just a projectile
		// victim). SetHits preserves position/type already accumulated
		// for this index.
		w.Npcs.SetHits(e.NpcIndex, e.Damage, e.CurHits, e.MaxHits)
		w.Recent.SetDamage(e.Damage, "")
		return true
	case event.NpcChat:
		// Opcode 104 type-1: an NPC spoke within view. Surface it via the
		// recent-dialog channel so routines watching NPC speech can read
		// it (the speaker name resolves from the npc def elsewhere).
		w.Recent.SetDialogText(e.MessageText)
		return true
	case event.NpcProjectile:
		// Opcode 104 type-3/4 (custom-client only on stock OpenRSC): an
		// NPC fired at a target. Record the incoming attack on a victim
		// NPC so "is this NPC being fought" stays answerable; player
		// victims are tracked on the player record's incoming side once
		// that field exists.
		if e.VictimIsNpc {
			w.Npcs.SetIncomingAttack(e.VictimNpcIndex, -1, e.ProjectileID)
		}
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
		// Land the combat-state bytes (combat level + skull/PK flag)
		// when the decoder recovered them.
		if e.HasCombat {
			w.Players.SetAppearanceCombat(e.PlayerIndex, e.CombatLevel, e.SkullType)
		}
		// Land the per-slot worn-equipment sprites (feeds the post-rename
		// self.equipped.<slot> / player.equipped.<slot> accessors). These
		// are sprite ids, not item ids — see PlayerRecord.EquipBySlot.
		if e.HasWorn {
			w.Players.SetWornEquipment(e.PlayerIndex, e.WornSprites)
		}
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
	case event.PrayersActive:
		w.Self.SetActivePrayers(e.Active)
		return true
	case event.BoundaryUpdates:
		// Offsets are player-relative at delivery; resolve to
		// absolute tiles using the player's CURRENT position.
		pos := w.Self.Position()
		for _, d := range e.Updates {
			ax := pos.X + d.OffsetX
			ay := pos.Y + d.OffsetY
			w.Boundaries.Set(ax, ay, d.Dir, d.ID)
		}
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
		// Land the damaged player's current/max hitpoints on the
		// mirror — this is the engaged target's health as the wire
		// encodes it (cur/max), feeding the #117 target-health view.
		w.Players.SetHits(e.PlayerIndex, e.Damage, e.CurHits, e.MaxHits)
		return true
	case event.OtherPlayerProjectile:
		// A player fired a projectile (opcode 234 type-3 at an NPC,
		// type-4 at a player) — this is the wire's who-is-fighting-whom
		// signal. Record the engagement on the caster, and mirror the
		// incoming attack onto the victim NPC so "is this NPC being
		// fought" is answerable.
		npcVictim := -1
		playerVictim := -1
		if e.VictimIsNpc {
			npcVictim = e.VictimNpcIndex
		} else {
			playerVictim = e.VictimPlayerIndex
		}
		w.Players.SetEngagement(e.CasterIndex, e.ProjectileID, npcVictim, playerVictim)
		if npcVictim >= 0 {
			w.Npcs.SetIncomingAttack(npcVictim, e.CasterIndex, e.ProjectileID)
		}
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
