package world

import (
	"sync"
	"time"
)

// RecentEvents is a small per-host buffer of the most-recent
// transient events (chat, PMs, damage, server messages, NPC dialog
// text). Populated by world.Apply from event-bus events; read by
// DSL routines via `world.last_chat` / `last_pm` / etc.
//
// Each slot holds the SINGLE most-recent event of that kind, not
// a list — routines that need "did anything just happen?" semantics
// look at the timestamp and decide whether the value is fresh.
// Routines that need a history-of-N use an `on` handler instead.
//
// Future work: extend to a small ring buffer per kind if a routine
// genuinely needs "the last 5 chats" rather than "the very latest."
// For Phase 2.5 single-slot-per-kind is enough.
type RecentEvents struct {
	mu             sync.RWMutex
	chat           *ChatRecord
	pm             *PMRecord
	damage         *DamageRecord
	serverMsg      *ServerMsgRecord
	dialogText     *DialogTextRecord
	dialogOptions  *DialogOptionsRecord

	// serverMsgRing is the bounded history of the last N server
	// messages (oldest-first), backing world.messages -> List<Message>
	// and the `on message` event (#119). The single-value serverMsg
	// above is kept as the newest entry for world.last_server_message
	// backward-compat — it equals serverMsgRing[len-1] once any message
	// has arrived. A nil/empty ring means no server message this session.
	serverMsgRing []ServerMsgRecord
}

// ServerMsgRingCap bounds the server-message history kept for
// world.messages. Old entries fall off the front once the ring is
// full; routines that need a longer tail should snapshot earlier.
const ServerMsgRingCap = 32

// NewRecentEvents constructs an empty buffer.
func NewRecentEvents() *RecentEvents { return &RecentEvents{} }

// ChatRecord is the most recent public chat we observed (from
// ChatReceived). Speaker / Message are server-supplied;
// At is the wall-clock time we recorded it.
type ChatRecord struct {
	Speaker string
	Message string
	At      time.Time
}

// PMRecord is the most recent incoming private message
// (PrivateMessage).
type PMRecord struct {
	Sender  string
	Message string
	At      time.Time
}

// DamageRecord is the most recent hit we took (DamageTaken).
// Source is the attacking entity's name (or empty if unknown).
type DamageRecord struct {
	Amount int
	Source string
	At     time.Time
}

// ServerMsgRecord is one player-facing message we observed. Most are
// gray in-game server messages ("You can't go through this door."),
// but the same ring also collects dialog-box / NPC-speech text so
// world.messages is a complete log of everything the server told the
// player. Kind distinguishes the source ("server" | "dialog"); it
// defaults to "server" for the legacy SetServerMessage path.
type ServerMsgRecord struct {
	Message string
	Kind    string
	At      time.Time
}

// DialogTextRecord is the most recent NPC speech-bubble text we
// observed. The accompanying option list (when the NPC presents
// choices) lives on a separate event; this buffer holds the
// scrolling text only.
type DialogTextRecord struct {
	Text string
	At   time.Time
}

// DialogOptionsRecord is the latest NPC dialog menu — the
// list of strings the server asked us to pick from. Cleared
// (set nil) once a routine answers (or by an explicit reset).
// At is when the menu was offered.
type DialogOptionsRecord struct {
	Options []string
	At      time.Time
}

// Chat returns the most recent ChatRecord, or nil if no chat has
// arrived this session.
func (r *RecentEvents) Chat() *ChatRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.chat == nil {
		return nil
	}
	c := *r.chat
	return &c
}

// SetChat records a new chat. Replaces any prior chat record.
func (r *RecentEvents) SetChat(speaker, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.chat = &ChatRecord{Speaker: speaker, Message: message, At: time.Now()}
}

// PM returns the most recent PMRecord.
func (r *RecentEvents) PM() *PMRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.pm == nil {
		return nil
	}
	c := *r.pm
	return &c
}

// SetPM records a new incoming private message.
func (r *RecentEvents) SetPM(sender, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.pm = &PMRecord{Sender: sender, Message: message, At: time.Now()}
}

// Damage returns the most recent DamageRecord.
func (r *RecentEvents) Damage() *DamageRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.damage == nil {
		return nil
	}
	c := *r.damage
	return &c
}

// SetDamage records a new hit.
func (r *RecentEvents) SetDamage(amount int, source string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.damage = &DamageRecord{Amount: amount, Source: source, At: time.Now()}
}

// ServerMessage returns the most recent server message.
func (r *RecentEvents) ServerMessage() *ServerMsgRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.serverMsg == nil {
		return nil
	}
	c := *r.serverMsg
	return &c
}

// SetServerMessage records a new server-origin (opcode 131) message.
// Equivalent to SetMessage("server", message) — kept as the named
// entry point most callers use.
func (r *RecentEvents) SetServerMessage(message string) {
	r.SetMessage("server", message)
}

// SetMessage records a new player-facing message of the given kind
// ("server" | "dialog"). Appends to the bounded ring (dropping the
// oldest once ServerMsgRingCap is exceeded) and updates the single-
// value latest slot for world.last_server_message. Feeding dialog /
// NPC-speech text here keeps world.messages a complete log of
// everything the server told the player, so a failing routine can be
// diagnosed from the message history regardless of which opcode
// carried the text.
func (r *RecentEvents) SetMessage(kind, message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rec := ServerMsgRecord{Message: message, Kind: kind, At: time.Now()}
	r.serverMsg = &rec
	r.serverMsgRing = append(r.serverMsgRing, rec)
	if len(r.serverMsgRing) > ServerMsgRingCap {
		// Drop oldest. Copy down rather than reslice so the backing
		// array doesn't grow unbounded over a long session.
		n := copy(r.serverMsgRing, r.serverMsgRing[len(r.serverMsgRing)-ServerMsgRingCap:])
		r.serverMsgRing = r.serverMsgRing[:n]
	}
}

// ServerMessages returns a copy of the bounded server-message ring,
// oldest-first. Empty (len 0) if no server message has arrived this
// session. Backs world.messages -> List<Message>.
func (r *RecentEvents) ServerMessages() []ServerMsgRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.serverMsgRing) == 0 {
		return nil
	}
	out := make([]ServerMsgRecord, len(r.serverMsgRing))
	copy(out, r.serverMsgRing)
	return out
}

// DialogText returns the most recent NPC dialog speech bubble.
func (r *RecentEvents) DialogText() *DialogTextRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.dialogText == nil {
		return nil
	}
	c := *r.dialogText
	return &c
}

// DialogOptions returns the most recent dialog options menu, or
// nil if no menu is currently presented. Routines branch on this
// to decide between answer(N) (after a find_option(text) lookup)
// vs ignoring an unexpected menu.
func (r *RecentEvents) DialogOptions() *DialogOptionsRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.dialogOptions == nil {
		return nil
	}
	c := *r.dialogOptions
	c.Options = append([]string(nil), r.dialogOptions.Options...)
	return &c
}

// SetDialogOptions records the menu the server just presented.
// Replaces any prior menu (a new menu while one is open means
// the old one was implicitly resolved — common in branching
// quest dialogs).
func (r *RecentEvents) SetDialogOptions(options []string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dialogOptions = &DialogOptionsRecord{Options: append([]string(nil), options...), At: time.Now()}
}

// ClearDialogOptions wipes the current menu — routines call this
// after answer() to keep the buffer accurate. The server doesn't
// always tell us when a menu closes, so explicit reset is part
// of the routine's responsibility.
func (r *RecentEvents) ClearDialogOptions() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dialogOptions = nil
}

// SetDialogText records new NPC speech bubble text.
func (r *RecentEvents) SetDialogText(text string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dialogText = &DialogTextRecord{Text: text, At: time.Now()}
}
