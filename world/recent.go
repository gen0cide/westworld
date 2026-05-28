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
	mu         sync.RWMutex
	chat       *ChatRecord
	pm         *PMRecord
	damage     *DamageRecord
	serverMsg  *ServerMsgRecord
	dialogText *DialogTextRecord
}

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

// ServerMsgRecord is the most recent server-origin chat (the gray
// in-game messages: "You can't go through this door.").
type ServerMsgRecord struct {
	Message string
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

// SetServerMessage records a new server-origin message.
func (r *RecentEvents) SetServerMessage(message string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.serverMsg = &ServerMsgRecord{Message: message, At: time.Now()}
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

// SetDialogText records new NPC speech bubble text.
func (r *RecentEvents) SetDialogText(text string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.dialogText = &DialogTextRecord{Text: text, At: time.Now()}
}
