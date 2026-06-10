package runtime

import (
	"fmt"
	"strings"
	"time"

	"github.com/gen0cide/westworld/dsl/interp"
	"github.com/gen0cide/westworld/world"
)

// ---------- recent-events views ----------

// chatRecordView wraps a world.ChatRecord so DSL routines can
// read .speaker / .message / .at.
type chatRecordView struct{ r *world.ChatRecord }

func (v *chatRecordView) Kind() string    { return "chat_record" }
func (v *chatRecordView) Display() string { return v.r.Speaker + ": " + v.r.Message }
func (v *chatRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "speaker":
		return interp.String(v.r.Speaker), true
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

type pmRecordView struct{ r *world.PMRecord }

func (v *pmRecordView) Kind() string    { return "pm_record" }
func (v *pmRecordView) Display() string { return "PM from " + v.r.Sender + ": " + v.r.Message }
func (v *pmRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "sender":
		return interp.String(v.r.Sender), true
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

type damageRecordView struct{ r *world.DamageRecord }

func (v *damageRecordView) Kind() string    { return "damage_record" }
func (v *damageRecordView) Display() string { return "took " + intDisp(v.r.Amount) + " damage" }
func (v *damageRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "amount":
		return interp.Int(int64(v.r.Amount)), true
	case "source":
		return interp.String(v.r.Source), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	}
	return nil, false
}

type serverMsgRecordView struct{ r *world.ServerMsgRecord }

func (v *serverMsgRecordView) Kind() string    { return "server_msg_record" }
func (v *serverMsgRecordView) Display() string { return v.r.Message }
func (v *serverMsgRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "message":
		return interp.String(v.r.Message), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Message}, true
	}
	return nil, false
}

// messageView is a Message INSTANCE (api.md §2/§8): one entry in the
// world.messages log. Fields: .text (String), .kind (String), .at
// (Time, formatted). Method: .contains(needle) — case-insensitive
// substring match. The Message type unifies what was the single
// last_server_message record into a list element.
type messageView struct {
	text string
	kind string
	at   time.Time
}

func (m *messageView) Kind() string    { return "message" }
func (m *messageView) Display() string { return m.text }
func (m *messageView) Get(field string) (interp.Value, bool) {
	switch field {
	case "text", "message":
		return interp.String(m.text), true
	case "kind":
		return interp.String(m.kind), true
	case "at":
		return interp.String(m.at.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: m.text}, true
	}
	return nil, false
}

type dialogTextRecordView struct{ r *world.DialogTextRecord }

func (v *dialogTextRecordView) Kind() string    { return "dialog_text_record" }
func (v *dialogTextRecordView) Display() string { return v.r.Text }
func (v *dialogTextRecordView) Get(field string) (interp.Value, bool) {
	switch field {
	case "text":
		return interp.String(v.r.Text), true
	case "at":
		return interp.String(v.r.At.Format("15:04:05")), true
	case "contains":
		return substringCallable{haystack: v.r.Text}, true
	}
	return nil, false
}

// substringCallable backs the `.contains(needle)` method on
// message record views. Case-insensitive by default — routines
// branch on prose without caring about server capitalization.
// Returns Bool. Used as: world.last_server_message.contains("locked").
type substringCallable struct{ haystack string }

func (c substringCallable) Kind() string    { return "builtin" }
func (c substringCallable) Display() string { return "<contains>" }
func (c substringCallable) Call(args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("contains takes 1 arg (needle), got %d", len(args))
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, fmt.Errorf("contains: needle must be String, got %s", args[0].Kind())
	}
	return interp.Bool(strings.Contains(strings.ToLower(c.haystack), strings.ToLower(string(needle)))), nil
}
