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
	Current      [18]int
	Max          [18]int
	Experience   [18]int
	QuestPoints  int
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
	ItemID   int
	Amount   int
	Wielded  bool
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

// WelcomeInfo: post-login welcome screen data.
type WelcomeInfo struct {
	base
	LastLoginIP      string
	DaysSinceLogin   int
	RecoveryDaysAgo  int
	UnreadMessages   int
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
type SleepScreenAppeared struct {
	base
	ImageBytes int // we don't decode the image; just note its size
}

func (SleepScreenAppeared) Kind() string { return "sleep_screen" }

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
	Index     int  // server-assigned player index
	X         int  // absolute world coord
	Y         int  // absolute world coord
	Sprite    int
	IsNew     bool // first time we've seen this player
	Removed   bool // player left view range
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
// update-types 3 (at NPC) and 4 (at player).
type OtherPlayerProjectile struct {
	base
	CasterIndex      int
	ProjectileID     int
	VictimNpcIndex   int
	VictimPlayerIndex int
}

func (OtherPlayerProjectile) Kind() string { return "other_player_projectile" }

// OtherPlayerAppearance: a player's appearance/identity was sent to
// us. From opcode 234 update-type 5. Phase 1.6 captures just the
// name + appearance ID; equipment/colors are consumed but not yet
// exposed as fields.
type OtherPlayerAppearance struct {
	base
	PlayerIndex  int
	Name         string
	AppearanceID int
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
