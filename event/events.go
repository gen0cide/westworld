package event

import "time"

// Event is implemented by every concrete event type. Kind returns a
// stable string identifier used for routing and logging.
type Event interface {
	Kind() string
	Time() time.Time
}

// base is embedded in every concrete event for the timestamp.
type base struct {
	at time.Time
}

func (b base) Time() time.Time { return b.at }
func newBase() base            { return base{at: time.Now()} }

// MessageType corresponds to OpenRSC's MessageType / RSChatMessage type
// codes — broadly the "channel" a chat message belongs to.
type MessageType int

const (
	MessageGame  MessageType = 0
	MessageQuest MessageType = 1
	MessageChat  MessageType = 2
	// More types exist in OpenRSC; add as needed.
)

// ChatReceived: another player said something publicly within view.
type ChatReceived struct {
	base
	Type    MessageType
	Speaker string // empty if system message
	Message string
	Color   string // optional, empty if none
}

func (ChatReceived) Kind() string { return "chat_received" }

func NewChatReceived(typ MessageType, speaker, message, color string) ChatReceived {
	return ChatReceived{base: newBase(), Type: typ, Speaker: speaker, Message: message, Color: color}
}

// PrivateMessage: a whisper from another player.
type PrivateMessage struct {
	base
	Sender     string
	FormerName string
	Message    string
	IconSprite int
	World      int
}

func (PrivateMessage) Kind() string { return "private_message" }

// SystemMessage: server-originated text (no player sender).
type SystemMessage struct {
	base
	Type    MessageType
	Message string
	Color   string
}

func (SystemMessage) Kind() string { return "system_message" }

// RoutineNote: the host's own in-character note() narration from a running
// routine — her richest, most "alive" output. Published to the bus so the
// cradle UI (and any subscriber) can stream it as a narration feed.
type RoutineNote struct {
	base
	Text string
}

func (RoutineNote) Kind() string { return "routine_note" }

// WhisperReceived: an operator thought injected over the debug control plane
// (debughttp POST /whisper) — the mentor channel without walking. The runtime
// queues it on the Host and voices it into the next director situation;
// published to the bus so feeds (cradle UI, JSONL log) show the injection the
// moment it lands.
type WhisperReceived struct {
	base
	Text    string
	Urgency string // low|normal|high
}

func (WhisperReceived) Kind() string { return "whisper" }

// NewWhisperReceived stamps a whisper event (the runtime publishes these
// outside the decoder, so it needs the package-internal base timestamp).
func NewWhisperReceived(text, urgency string) WhisperReceived {
	return WhisperReceived{base: newBase(), Text: text, Urgency: urgency}
}

// SkillID maps RSC skill positions used in stat packets.
type SkillID int

const (
	SkillAttack     SkillID = 0
	SkillDefence    SkillID = 1
	SkillStrength   SkillID = 2
	SkillHits       SkillID = 3
	SkillRanged     SkillID = 4
	SkillPrayer     SkillID = 5
	SkillMagic      SkillID = 6
	SkillCooking    SkillID = 7
	SkillWoodcut    SkillID = 8
	SkillFletching  SkillID = 9
	SkillFishing    SkillID = 10
	SkillFiremaking SkillID = 11
	SkillCrafting   SkillID = 12
	SkillSmithing   SkillID = 13
	SkillMining     SkillID = 14
	SkillHerblaw    SkillID = 15
	SkillAgility    SkillID = 16
	SkillThieving   SkillID = 17
)

// SkillName returns a human label for a skill ID.
func SkillName(s SkillID) string {
	switch s {
	case SkillAttack:
		return "attack"
	case SkillDefence:
		return "defence"
	case SkillStrength:
		return "strength"
	case SkillHits:
		return "hits"
	case SkillRanged:
		return "ranged"
	case SkillPrayer:
		return "prayer"
	case SkillMagic:
		return "magic"
	case SkillCooking:
		return "cooking"
	case SkillWoodcut:
		return "woodcutting"
	case SkillFletching:
		return "fletching"
	case SkillFishing:
		return "fishing"
	case SkillFiremaking:
		return "firemaking"
	case SkillCrafting:
		return "crafting"
	case SkillSmithing:
		return "smithing"
	case SkillMining:
		return "mining"
	case SkillHerblaw:
		return "herblaw"
	case SkillAgility:
		return "agility"
	case SkillThieving:
		return "thieving"
	}
	return "unknown"
}

// StatUpdate: one skill's current/max/experience changed.
type StatUpdate struct {
	base
	Skill      SkillID
	Current    int
	Max        int
	Experience int
}

func (StatUpdate) Kind() string { return "stat_update" }

// StatsSnapshot: full stats dump (typically after login).
type StatsSnapshot struct {
	base
	Current     [18]int
	Max         [18]int
	Experience  [18]int
	QuestPoints int
}

func (StatsSnapshot) Kind() string { return "stats_snapshot" }

// ExperienceGain: a single skill gained XP. Not always paired with
// a StatUpdate (server may emit one or the other).
type ExperienceGain struct {
	base
	Skill SkillID
	XP    int
}

func (ExperienceGain) Kind() string { return "experience_gain" }

// FatigueUpdate: server says fatigue value changed.
type FatigueUpdate struct {
	base
	// Value is the server's "scaled" fatigue value 0..750 (or whatever
	// the server normalizes to). The actual fatigue percent is approx
	// (Value / 7.5) %. Don't rely on this being a clean percent.
	Value int
}

func (FatigueUpdate) Kind() string { return "fatigue_update" }

// InventoryItem represents one item slot's content. ItemID can match
// OpenRSC's ItemId catalog. Wielded is true if the high bit was set on
// the wire (the item is currently equipped).
type InventoryItem struct {
	ItemID  int
	Amount  int
	Wielded bool
}

// Equipment slot indices, matching OpenRSC's AppearanceId.SLOT_*
// constants. The appearance update (opcode 234 / 104 type-5) sends a
// worn-sprite byte per slot in this exact order, so the wire index ==
// the slot constant.
//
// IMPORTANT: these are RSC body-ANIMATION layers, not human equip slots —
// each "head/body/legs" concept has TWO layers (the metal armour type
// picks which), verified against ItemDefs.json wearSlot:
//
//	0 EquipSlotHead  = LARGE/full helmets        }
//	5 EquipSlotHat   = MEDIUM helmets            } both render on the head
//	1 EquipSlotShirt = plate-mail BODY           }
//	6 EquipSlotBody  = chain-mail / leather body } both render on the torso
//	2 EquipSlotPants = plate-mail LEGS           }
//	7 EquipSlotLegs  = skirts                     } both render on the legs
//
// A player can only wear one of each pair, so the runtime's per-slot
// accessors (helmet/body/legs) check BOTH layers and return whichever is
// worn (see runtime equipSlotGroups). EquipSlotName collapses each pair to
// one label.
const (
	EquipSlotHead   = 0 // large / full helmet
	EquipSlotShirt  = 1 // plate-mail body
	EquipSlotPants  = 2 // plate-mail legs
	EquipSlotShield = 3
	EquipSlotWeapon = 4
	EquipSlotHat    = 5 // medium helmet
	EquipSlotBody   = 6 // chain-mail / leather body
	EquipSlotLegs   = 7 // skirt (leg armour)
	EquipSlotGloves = 8
	EquipSlotBoots  = 9
	EquipSlotAmulet = 10
	EquipSlotCape   = 11
	// NumEquipSlots is the worn-items array length OpenRSC sends in the
	// authentic v235 appearance update (Player.wornItems is int[12]).
	NumEquipSlots = 12
)

// EquipSlotName returns a human label for an equip slot index.
func EquipSlotName(slot int) string {
	switch slot {
	case EquipSlotHead, EquipSlotHat:
		return "head" // large helm (0) or medium helm (5)
	case EquipSlotShirt, EquipSlotBody:
		return "body" // platebody (1) or chain/leather (6)
	case EquipSlotPants, EquipSlotLegs:
		return "legs" // platelegs (2) or skirt (7)
	case EquipSlotShield:
		return "shield"
	case EquipSlotWeapon:
		return "weapon"
	case EquipSlotGloves:
		return "gloves"
	case EquipSlotBoots:
		return "boots"
	case EquipSlotAmulet:
		return "amulet"
	case EquipSlotCape:
		return "cape"
	}
	return "unknown"
}

// InventorySnapshot: full inventory delivery (typically after login).
type InventorySnapshot struct {
	base
	Items []InventoryItem
}

func (InventorySnapshot) Kind() string { return "inventory_snapshot" }

// InventorySlotUpdate: one inventory slot changed.
type InventorySlotUpdate struct {
	base
	Slot int
	Item *InventoryItem // nil means slot was cleared
}

func (InventorySlotUpdate) Kind() string { return "inventory_slot_update" }

// InventoryRemoveSlot: the item at Slot was removed and every later
// slot shifts down one (the server's ArrayList.remove(index) semantics —
// RSC inventories have no holes). Distinct from InventorySlotUpdate with
// a nil Item, which only blanks a single slot: a burst of RemoveSlot(0)
// (e.g. ::wipeinv removing get(0) repeatedly) empties the whole list,
// whereas repeated blank-slot-0 would leave items in slots 1..n.
type InventoryRemoveSlot struct {
	base
	Slot int
}

func (InventoryRemoveSlot) Kind() string { return "inventory_remove_slot" }

// ItemGained: synthetic event fired when the host's inventory
// gains a net positive count of an item id. Derived by diffing
// pre/post InventorySnapshot or InventorySlotUpdate. Routines
// subscribe via `on item_gained(item_id, count) { ... }` instead
// of polling inventory.count() in a loop.
//
// "Net positive" means an actual addition, not an equip-shuffle or
// stack-into-existing. A 1-coin add to a 100-coin stack fires
// item_gained(10, 1), not (10, 101).
type ItemGained struct {
	base
	ItemID int
	Count  int // positive delta in this event
}

func (ItemGained) Kind() string { return "item_gained" }

// WelcomeInfo: post-login welcome screen data.
type WelcomeInfo struct {
	base
	LastLoginIP     string
	DaysSinceLogin  int
	RecoveryDaysAgo int
	UnreadMessages  int
}

func (WelcomeInfo) Kind() string { return "welcome_info" }

// LogoutConfirm: server acknowledges our logout request.
type LogoutConfirm struct {
	base
}

func (LogoutConfirm) Kind() string { return "logout_confirm" }

// Death: we died.
type Death struct {
	base
}

func (Death) Kind() string { return "death" }

// GroundItemEvent: an item entered or left visibility. Coords are
// relative to the player at delivery time.
type GroundItemEvent struct {
	base
	ItemID    int
	OffsetX   int
	OffsetY   int
	Disappear bool
}

func (GroundItemEvent) Kind() string {
	return "ground_item"
}

// NpcDialog: server asks us to choose from a list of dialog options.
type NpcDialog struct {
	base
	Options []string
}

func (NpcDialog) Kind() string { return "npc_dialog" }

// NpcDialogText: NPC speech bubble text.
type NpcDialogText struct {
	base
	Text string
}

func (NpcDialogText) Kind() string { return "npc_dialog_text" }

// SleepScreenAppeared: the fatigue captcha is being shown.
//
// Wire: SEND_SLEEPSCREEN (v235 opcode 117), payload is the raw captcha
// image bytes (OpenRSC Payload235Generator.java:397-400 — writeBytes of
// ss.image). We don't OCR the image: on this server no prerendered
// sleepword archive is loaded (conf/server/data/sleepwords/ is absent),
// so SleepHandler.process has knowCorrectWord=false and accepts ANY
// answer. The runtime answers with SendSleepWord("asleep") — but only
// AFTER the sleep-fatigue drain (SleepFatigueUpdate) reaches 0:
// answering immediately wakes the player before any per-tick drain has
// happened, committing the UNDRAINED value back (the soak-retro "sleep
// is a behavioral no-op" bug). See runtime/frame.go.
type SleepScreenAppeared struct {
	base
	ImageBytes int // we don't decode the image; just note its size
}

func (SleepScreenAppeared) Kind() string { return "sleep_screen" }

// SleepFatigueUpdate: the server's per-tick report of the PROVISIONAL
// fatigue value draining while we are on the sleep screen.
//
// Wire: SEND_SLEEP_FATIGUE (v235 opcode 244), same [short scaled
// 0..750] body as SEND_FATIGUE (Payload235Generator.java shares one
// case for both). While asleep the server drains
// Player.sleepStateFatigue once per game tick (bed −42000, sleeping
// bag −8400, of MAX_FATIGUE 150000 — Player.startSleepEvent) and
// reports it here; the value only COMMITS to real fatigue when the
// sleepword answer wakes us successfully (Player.handleWakeup:
// fatigue = sleepStateFatigue). Deliberately NOT applied to
// world.Self fatigue — an unexpected wake keeps the old value.
type SleepFatigueUpdate struct {
	base
	Value int
}

func (SleepFatigueUpdate) Kind() string { return "sleep_fatigue_update" }

// SleepwordIncorrect: our last SLEEPWORD_ENTERED answer was rejected.
//
// Wire: SEND_SLEEPWORD_INCORRECT (v235 opcode 194), no payload
// (NoPayloadStruct). The server keeps us asleep and re-sends a fresh
// SEND_SLEEPSCREEN after incorrect-tries seconds (SleepHandler.java),
// so the runtime's answer flow retries on that next screen instead of
// trapping asleep forever. Only possible when the server has a
// prerendered sleepword archive loaded with KNOWN words — the current
// westworld.conf setup accepts any answer, so seeing this event means
// the server config changed under us.
type SleepwordIncorrect struct {
	base
}

func (SleepwordIncorrect) Kind() string { return "sleepword_incorrect" }

// SleepEnded: the server woke us up (correct sleep word, or the server
// otherwise ended the sleep). Wire: SEND_STOPSLEEP (v235 opcode 84), no
// payload (OpenRSC Payload235Generator.java:127 — break with no body).
// On a successful wake the drained fatigue commits server-side
// (SleepHandler → sendWakeUp(true) → handleWakeup) and the new value
// rides in on a separate SEND_FATIGUE packet handled as FatigueUpdate.
type SleepEnded struct {
	base
}

func (SleepEnded) Kind() string { return "sleep_ended" }

// UnknownPacket: a frame whose opcode we don't yet decode. Captured
// for observability (the technician tablets can show these to the
// admin).
type UnknownPacket struct {
	base
	Opcode      byte
	PayloadSize int
}

func (UnknownPacket) Kind() string { return "unknown_packet" }

// OwnPositionUpdate: server says our player is at this absolute
// position. Derived from the bitpacked SendPlayerCoords packet
// (opcode 191).
type OwnPositionUpdate struct {
	base
	X      int
	Y      int
	Sprite int // 0-15, our facing direction / animation index
}

func (OwnPositionUpdate) Kind() string { return "own_position" }

// NearbyPlayerEvent: a player entered or moved within visibility.
// Computed from the bitpacked SendPlayerCoords packet by resolving
// offset deltas against our own position.
type NearbyPlayerEvent struct {
	base
	Index   int // server-assigned player index
	X       int // absolute world coord
	Y       int // absolute world coord
	Sprite  int
	IsNew   bool // first time we've seen this player
	Removed bool // player left view range
}

func (NearbyPlayerEvent) Kind() string { return "nearby_player" }

// OtherPlayerChat: a player visible to us said something. Decoded
// from inbound opcode 234 (UpdatePlayers), update-type 1/6/7.
// MessageText may be empty if the RSC compressed-string decoder
// couldn't recover the body (Phase 1.6 limitation); MessageRaw is
// always the raw compressed bytes for later analysis.
type OtherPlayerChat struct {
	base
	PlayerIndex int
	Icon        int    // chat icon byte (e.g., admin crown = 2 or 3)
	ChatKind    string // "public", "quest", "muted"
	MessageText string // decoded text (best effort)
	MessageRaw  []byte // raw compressed bytes
}

func (OtherPlayerChat) Kind() string { return "other_player_chat" }

// OtherPlayerDamage: a player took damage. Decoded from opcode 234
// update-type 2.
type OtherPlayerDamage struct {
	base
	PlayerIndex int
	Damage      int
	CurHits     int
	MaxHits     int
}

func (OtherPlayerDamage) Kind() string { return "other_player_damage" }

// OtherPlayerProjectile: a player fired a projectile. From opcode 234
// update-types 3 (at NPC) and 4 (at player). This is the wire's
// who-is-fighting-whom signal for ranged/magic combat.
//
// VictimIsNpc disambiguates which victim field is authoritative
// (server indices start at 0, so a zero VictimNpcIndex/VictimPlayerIndex
// is a valid target — callers must branch on VictimIsNpc, not on a
// zero check).
type OtherPlayerProjectile struct {
	base
	CasterIndex       int
	ProjectileID      int
	VictimNpcIndex    int
	VictimPlayerIndex int
	VictimIsNpc       bool // true: VictimNpcIndex valid (type 3); false: VictimPlayerIndex valid (type 4)
}

func (OtherPlayerProjectile) Kind() string { return "other_player_projectile" }

// OtherPlayerAppearance: a player's appearance/identity was sent to
// us. From opcode 234 update-type 5. Captures the name + appearance
// ID plus the two trailing combat-state bytes the v235 wire carries
// (combatLevel + skullType). Equipment/colors are consumed but not
// yet exposed as fields.
//
// Wire layout of the two combat bytes (GameStateUpdater.java:1010-1011,
// the isUsing233CompatibleClient path): after the worn-item sprites
// and the 4 colour bytes, the server appends
//
//	[byte] combatLevel  (the player's combat level, 3..~123)
//	[byte] skullType    (0 = no skull, 1 = skulled / PK-flagged —
//	                      engaged in recent player-vs-player combat)
//
// HasCombat is false when the decoder couldn't reach those bytes
// (truncated/abbreviated record); consumers must not treat a zero
// CombatLevel as authoritative unless HasCombat is true.
type OtherPlayerAppearance struct {
	base
	PlayerIndex  int
	Name         string
	AppearanceID int
	CombatLevel  int  // trailing combat-level byte (0 if !HasCombat)
	SkullType    int  // 0 = no skull, 1 = skulled / PK-flagged
	HasCombat    bool // the two trailing combat bytes were decoded

	// The four appearance colour bytes, in the wire order the server
	// writes them (mudclient.java:5429-5432): hair, top, trouser
	// (bottom), skin. Hair/top/trouser are INDICES into the client
	// clothing-colour tables; skin is an index into the skin table.
	// They drive the per-layer dye when the renderer composites this
	// player's real appearance. HasColours is false when the decoder
	// couldn't reach those bytes (truncated/abbreviated record).
	HairColour    int
	TopColour     int
	TrouserColour int
	SkinColour    int
	HasColours    bool

	// WornSprites is the per-slot worn-equipment sprite IDs decoded from
	// the equipment block (one byte per slot, indexed by EquipSlot*).
	// The wire carries SPRITE / appearance IDs (the low byte of
	// AppearanceId, i.e. itemDef.getAppearanceId() & 0xFF), NOT catalogue
	// item IDs — there is no item-id-by-slot in this packet, so mapping a
	// sprite back to an item requires a sprite→item lookup we don't have
	// yet (mapping gap). A zero value means "nothing worn in that slot".
	// HasWorn is true once the equipment block was decoded.
	WornSprites [NumEquipSlots]int
	WornCount   int // number of slot bytes the wire actually carried
	HasWorn     bool
}

func (OtherPlayerAppearance) Kind() string { return "other_player_appearance" }

// PlayerActionBubble: a player's action bubble showed a sprite. From
// opcode 234 update-type 0.
type PlayerActionBubble struct {
	base
	PlayerIndex int
	BubbleID    int
}

func (PlayerActionBubble) Kind() string { return "player_action_bubble" }

// NpcNearby: an NPC is visible within our view. From inbound opcode
// 79's bitpacked NPC update.
type NpcNearby struct {
	base
	Index  int // server-assigned NPC index (stable within a tick set)
	X      int // absolute world coord (IsNew only — player-relative offset)
	Y      int // absolute world coord (IsNew only)
	DX, DY int // RELATIVE one-tile move delta (movement updates only). opcode-79
	// movement encodes a DIRECTION, not a coord, so a moving NPC must be shifted
	// from its OWN stored position; basing it on the host's position teleported
	// every moving NPC onto the host (the phantom-crowd bug).
	Sprite  int  // facing direction / animation
	TypeID  int  // joins to facts.NpcDef.ID for name lookup
	IsNew   bool // first time we've seen this NPC
	Removed bool // server dropped this NPC from view (despawn/death/left range)
}

func (NpcNearby) Kind() string { return "npc_nearby" }

// NpcDamage: an NPC took damage. Decoded from inbound opcode 104
// (SEND_UPDATE_NPC) update-type 2 — the NPC's OWN current/max
// hitpoints as the wire encodes them (cur/max, not a fraction).
//
// Wire layout of the type-2 record (GameStateUpdater.java:529-536):
//
//	[short] npcIndex
//	[byte]  2          (update type)
//	[byte]  damage     (the hit just applied)
//	[byte]  curHits    (NPC's current hitpoints AFTER the hit)
//	[byte]  maxHits    (NPC's max hitpoints)
//
// Unlike opcode 234 (which only surfaces an NPC's health indirectly,
// as a projectile victim), opcode 104 carries every visible NPC's own
// health whenever it changes. This un-stubs Npc.health for ANY npc.
type NpcDamage struct {
	base
	NpcIndex int
	Damage   int
	CurHits  int
	MaxHits  int
}

func (NpcDamage) Kind() string { return "npc_damage" }

// NpcChat: an NPC said something within view. Decoded from inbound
// opcode 104 (SEND_UPDATE_NPC) update-type 1. MessageText is the
// decoded (RSC-decompressed) body. RecipientIndex is the player the
// NPC is addressing (-1 if broadcast / none).
type NpcChat struct {
	base
	NpcIndex       int
	RecipientIndex int
	MessageText    string
	MessageRaw     []byte
}

func (NpcChat) Kind() string { return "npc_chat" }

// NpcProjectile: an NPC fired a projectile at a target. Decoded from
// inbound opcode 104 update-types 3 (at NPC) and 4 (at player). Only
// sent to custom clients by OpenRSC; decoded here for completeness so
// the record stream stays aligned if a server ever emits it. Mirrors
// OtherPlayerProjectile.
type NpcProjectile struct {
	base
	CasterNpcIndex    int
	ProjectileID      int
	VictimNpcIndex    int
	VictimPlayerIndex int
	VictimIsNpc       bool
}

func (NpcProjectile) Kind() string { return "npc_projectile" }

// TradeRequestReceived: another player has initiated a trade with us.
// We can accept (by re-sending InitTradeRequest pointing back at them
// — RSC trades use a symmetric "both sides request each other"
// handshake; the server opens the window only when both ends
// match) or decline.
//
// The OpenRSC server signals this via opcode 131 SEND_SERVER_MESSAGE
// with MessageType=TRADE(6), sender=<requester>, text="" (for
// client version > 204). The packet carries no server-index, so we
// only know the requester's NAME — callers needing an index resolve
// via world.Players.
type TradeRequestReceived struct {
	base
	FromPlayerIndex int    // -1 if unknown; OpenRSC's notification has only the name
	FromPlayerName  string // always set
}

func (TradeRequestReceived) Kind() string { return "trade_request" }

// DuelRequestReceived: another player wants to duel us. OpenRSC
// notifies via SEND_SERVER_MESSAGE with MessageType=INVENTORY(7)
// and text containing "wishes to duel with you" — sender is
// embedded in the text body. The decoder parses the name out.
type DuelRequestReceived struct {
	base
	FromPlayerName string
}

func (DuelRequestReceived) Kind() string { return "duel_request" }

// TradeOpened: both players accepted; the trade window is now active.
// Contains our pending items and the other player's pending items.
type TradeOpened struct {
	base
	OtherPlayerIndex int
	OpponentName     string
	MyItems          []InventoryItem
	OpponentItems    []InventoryItem
}

func (TradeOpened) Kind() string { return "trade_opened" }

// TradeOtherOffer: opponent updated their offered items list. Fires
// every time they add/remove. Used to keep world.Trade.TheirOffer
// in sync. Replaces TheirOffer wholesale; this is server-canonical.
type TradeOtherOffer struct {
	base
	Items []InventoryItem
}

func (TradeOtherOffer) Kind() string { return "trade_other_offer" }

// TradeClosed: trade was cancelled (by either side).
type TradeClosed struct {
	base
	// Completed is true if the trade went all the way through
	// (both sides confirmed twice and items exchanged). False
	// for any cancellation path (decline, abort during offer,
	// abort during confirm).
	Completed bool
}

func (TradeClosed) Kind() string { return "trade_closed" }

// TradeOtherAccepted: the other player clicked "Accept" in the trade
// window.
type TradeOtherAccepted struct {
	base
}

func (TradeOtherAccepted) Kind() string { return "trade_other_accepted" }

// --- Duel events ---
//
// Duels share the two-screen handshake shape with trades plus a set
// of rule toggles (retreat/magic/prayer/weapons). One event per
// distinct server packet so DSL handlers can pick which to react to.

// DuelOpened: both players accepted; the duel offer window is now
// active. Payload from SEND_DUEL_WINDOW (opcode 176 inbound) — just
// the opponent's server-side player index.
type DuelOpened struct {
	base
	OtherPlayerIndex int
}

func (DuelOpened) Kind() string { return "duel_opened" }

// DuelOtherOffer: opponent updated their staked items. Payload from
// SEND_DUEL_OPPONENTS_ITEMS (opcode 6 inbound).
type DuelOtherOffer struct {
	base
	Items []InventoryItem
}

func (DuelOtherOffer) Kind() string { return "duel_other_offer" }

// DuelSettingsUpdate: opponent (or we, echoed) changed the rule
// toggles. Payload from SEND_DUEL_SETTINGS (opcode 30 inbound).
type DuelSettingsUpdate struct {
	base
	DisallowRetreat bool
	DisallowMagic   bool
	DisallowPrayer  bool
	DisallowWeapons bool
}

func (DuelSettingsUpdate) Kind() string { return "duel_settings_update" }

// DuelOtherAccepted: opponent clicked Accept on the current screen
// (offer or confirm). Payload from SEND_DUEL_OTHER_ACCEPTED (opcode
// 253) or SEND_DUEL_ACCEPTED (opcode 210) — both carry an accepted
// byte that the decoder normalizes.
type DuelOtherAccepted struct {
	base
}

func (DuelOtherAccepted) Kind() string { return "duel_other_accepted" }

// DuelConfirmShown: server pushed the final confirm-window
// (SEND_DUEL_CONFIRMWINDOW, opcode 172) — both sides have
// first-accepted and we're now on the final review screen.
type DuelConfirmShown struct {
	base
	OpponentName    string
	OpponentItems   []InventoryItem
	MyItems         []InventoryItem
	DisallowRetreat bool
	DisallowMagic   bool
	DisallowPrayer  bool
	DisallowWeapons bool
}

func (DuelConfirmShown) Kind() string { return "duel_confirm_shown" }

// DuelClosed: duel ended. Completed=true means both confirmed and
// the fight has started; false on any cancel/decline path
// (SEND_DUEL_CLOSE, opcode 225).
type DuelClosed struct {
	base
	Completed bool
}

func (DuelClosed) Kind() string { return "duel_closed" }

// TradeConfirmShown: server pushed the final confirm-window
// (SEND_TRADE_OPEN_CONFIRM, opcode 20). Same shape as
// DuelConfirmShown — opponent name + both item lists.
type TradeConfirmShown struct {
	base
	OpponentName  string
	OpponentItems []InventoryItem
	MyItems       []InventoryItem
}

func (TradeConfirmShown) Kind() string { return "trade_confirm_shown" }

// --- Bank events ---
//
// BankOpened: server pushed the full bank snapshot in response to
// the player talking to a banker. Maxsize is the bank's slot
// capacity (varies by membership / quest unlocks).
type BankOpened struct {
	base
	MaxSize int
	Items   []InventoryItem
}

func (BankOpened) Kind() string { return "bank_opened" }

// BankSlotUpdate: a single bank slot was updated (deposit, withdraw,
// or server adjustment).
type BankSlotUpdate struct {
	base
	Slot   int
	ItemID int
	Amount int
}

func (BankSlotUpdate) Kind() string { return "bank_slot_update" }

// BankClosed: the bank window closed (player walked away or
// explicitly closed it).
type BankClosed struct {
	base
}

func (BankClosed) Kind() string { return "bank_closed" }

// --- Shop events ---
//
// ShopOpened: server pushed the full shop catalogue (SEND_SHOP_OPEN,
// opcode 101) in response to the player opening a shop via NPC
// interaction. RSC has no per-slot shop-update packet — the server
// re-sends the whole list on every stock change, so a fresh
// ShopOpened with the same shop IS the stock-update path. The world
// layer detects per-item stock deltas against its prior snapshot if a
// shop_stock_update event is wanted (see api.md §8 shop events).
type ShopOpened struct {
	base
	// IsGeneral is true for a general store (buys anything).
	IsGeneral bool
	// SellPriceMod / BuyPriceMod / PriceMultiplier are the shop's
	// price-percentage modifiers (RSC sellModifier / buyModifier /
	// stockSensitivity). See world.ShopRecord for the price formula.
	SellPriceMod    int
	BuyPriceMod     int
	PriceMultiplier int
	// Items is the catalogue, one entry per stocked item.
	Items []ShopItem
}

func (ShopOpened) Kind() string { return "shop_opened" }

// ShopItem is one shop-catalogue entry decoded from SEND_SHOP_OPEN.
// Stock is the buyable quantity; BaseStock is the shop's baseline
// stock (the reference point for stock-sensitive pricing, not a gp
// value).
type ShopItem struct {
	ItemID    int
	Stock     int
	BaseStock int
}

// ShopClosed: the shop window closed (SEND_SHOP_CLOSE, opcode 137, or
// the player walked away / sent shop.close()).
type ShopClosed struct {
	base
}

func (ShopClosed) Kind() string { return "shop_closed" }

// PrayersActive: full snapshot of which prayers are currently
// active. The server pushes this on every prayer toggle and at
// login. Active[i] is true iff prayer slot i is on. Slot count is
// 14 in stock RSC but the packet length is what the server sent —
// we don't enforce a fixed length.
type PrayersActive struct {
	base
	Active []bool
}

func (PrayersActive) Kind() string { return "prayers_active" }

// BoundaryDelta is one entry from SEND_BOUNDARY_HANDLER. id=-1 (or
// 0xFFFF as a u16) marks the boundary as removed (door opened, web
// cut). Otherwise id is the new boundary def ID at (X, Y, Dir).
// Coords are relative to the player at packet delivery; the
// world.Apply path resolves them to absolute tiles.
type BoundaryDelta struct {
	ID      int // -1 = removed; otherwise boundary def ID
	OffsetX int
	OffsetY int
	Dir     int
}

// BoundaryUpdates: a batch of dynamic boundary changes within
// view. Each delta is one (x, y, dir) tile/direction tuple.
type BoundaryUpdates struct {
	base
	Updates []BoundaryDelta
}

func (BoundaryUpdates) Kind() string { return "boundary_updates" }

// SceneryDelta is one entry from SEND_SCENERY_HANDLER (opcode 48).
// id == 60000 (SceneryRemoveSentinel) marks the scenery at this tile
// as removed (object left view, fire burned out, rock depleted).
// Otherwise id is the scenery def ID now occupying (X, Y). Coords are
// player-relative at packet delivery; world.Apply resolves them to
// absolute tiles.
type SceneryDelta struct {
	ID      int // 60000 = removed; otherwise scenery def ID
	OffsetX int
	OffsetY int
}

// SceneryUpdates: a batch of dynamic scenery (GameObject) changes
// within view. The initial region-load packet enumerates all visible
// scenery; subsequent packets carry single adds/removes (e.g. a lit
// fire appearing at the player's tile).
type SceneryUpdates struct {
	base
	Updates []SceneryDelta
}

func (SceneryUpdates) Kind() string { return "scenery_updates" }

// GroundItemDelta is one entry in a SEND_GROUND_ITEM (opcode 99) batch — an
// item appearing on, or being cleared from, a tile relative to the player.
// Disappear=true means the server removed the item at this tile (either the
// in-range 0x8000-id removal or the out-of-range 255 clear).
type GroundItemDelta struct {
	ItemID    int // 0 for the out-of-range (255) clear, where no id is sent
	OffsetX   int
	OffsetY   int
	Disappear bool
}

// GroundItemUpdates is the batch of ground-item changes from ONE opcode-99
// packet. The server streams EVERY in-view ground item per update, so the
// decoder must read all entries — the prior single-entry decode hid items
// (e.g. the free Barbarian-Village pickaxe) behind whatever came first.
type GroundItemUpdates struct {
	base
	Updates []GroundItemDelta
}

func (GroundItemUpdates) Kind() string { return "ground_item_updates" }

// RemovePoint is one player-relative point from SEND_REMOVE_WORLD_ENTITY
// (opcode 211). The absolute tile (self+offset) names an 8x8 REGION
// (abs>>3) whose dynamic boundary/scenery/ground-item entries the server
// has cleaned up.
type RemovePoint struct {
	OffsetX int
	OffsetY int
}

// RemoveWorldEntities is the opcode-211 bulk region clear — the server's
// only eviction channel for far-away dynamic entities. world.Apply sweeps
// every store entry whose (x>>3, y>>3) matches a point's region.
type RemoveWorldEntities struct {
	base
	Points []RemovePoint
}

func (RemoveWorldEntities) Kind() string { return "remove_world_entities" }

// ===== #119 synthetic events =====
//
// XPGain and TargetDied are SYNTHESIZED in runtime/frame.go by diffing
// world state across an Apply (the same pattern as ItemGained). The
// raw wire packets carry totals/health, not deltas/death-edges — the
// host computes the edge and publishes these so DSL routines get a
// clean "this just happened" hook.

// XPGain: a skill's total experience just increased by Amount. Unlike
// the raw ExperienceGain / StatUpdate packets (which carry the NEW
// TOTAL xp, not the delta), this synthetic event carries the positive
// delta plus the running total, derived by diffing the per-skill xp
// mirror before/after Apply. Powers the `on xp_gain(skill, amount)`
// DSL event. SkillName(Skill) gives the lowercase skill name the DSL
// handler binds.
type XPGain struct {
	base
	Skill  SkillID
	Amount int // positive xp delta (new_total - old_total)
	Total  int // new total xp for the skill
}

func (XPGain) Kind() string { return "xp_gain" }

// TargetDied: the player's engaged combat target (the NPC most
// recently attacked, i.e. combat.target / combat.last_npc) just
// transitioned to dead — its opcode-104 (SEND_UPDATE_NPC type-2)
// current-hitpoints reading went from >0 to 0. Synthesized in
// runtime/frame.go by watching NpcDamage on the engaged index.
// Powers the `on npc_killed(target)` and `on target_died(target)`
// DSL events. NpcIndex is the dead target's server index (resolve to
// a view via world.Npcs); TypeID joins to facts.NpcDef for the name.
type TargetDied struct {
	base
	NpcIndex int
	TypeID   int
}

func (TargetDied) Kind() string { return "target_died" }

// LevelUp: one of the host's OWN skills just gained a base level — its
// max (base) level increased. Synthesized in runtime/frame.go by diffing
// per-skill max levels across stat updates (the same edge pattern as
// XPGain). Powers `on level_up(skill, new_level)`. Skill is the catalog
// id (use event.SkillName); NewLevel is the new base level.
type LevelUp struct {
	base
	Skill    SkillID
	NewLevel int
}

func (LevelUp) Kind() string { return "level_up" }

// EquipmentChanged: the host's OWN worn/wielded set changed (an item was
// equipped or removed). Synthesized by diffing the inventory's Wielded
// flags across inventory updates. Powers `on equipment_changed()`; the
// handler reads self.equipped / self.wielded for the new state.
type EquipmentChanged struct {
	base
	Slot string // human slot that changed: helmet/body/legs/weapon/...
}

func (EquipmentChanged) Kind() string { return "equipment_changed" }

// PlayerEquipmentChanged: another visible player's worn equipment
// changed (their appearance packet's per-slot sprites differ from what
// we last saw). Synthesized in runtime/frame.go by diffing
// PlayerRecord.EquipBySlot on each appearance update. Powers
// `on player_equipment_changed(player)`. PlayerIndex resolves to a
// world.players view; Name is the player's name.
type PlayerEquipmentChanged struct {
	base
	PlayerIndex int
	Name        string
	Slot        string // human slot that changed: helmet/body/legs/weapon/...
}

func (PlayerEquipmentChanged) Kind() string { return "player_equipment_changed" }

// PlayerLevelChanged: another visible player's combat level changed (the
// appearance packet's combat-level byte differs from what we last saw).
// Synthesized by diffing PlayerRecord.CombatLevel. Powers
// `on player_level_changed(player, new_level)`.
type PlayerLevelChanged struct {
	base
	PlayerIndex int
	Name        string
	OldLevel    int
	NewLevel    int
}

func (PlayerLevelChanged) Kind() string { return "player_level_changed" }
