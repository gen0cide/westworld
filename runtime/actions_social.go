package runtime

import (
	"context"
	"strings"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Social & chat action handler bodies (say / whisper / add_friend /
// find_option). Registered in the central actionHandlers table in
// dsl_actions.go.

// ---------- social ----------

func dslSay(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("say takes 1 argument (message), got %d", len(args))
	}
	msg := args[0].Display()
	if err := h.Say(ctx, msg); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslWhisper(ctx context.Context, h *Host, args []interp.Value, named map[string]interp.Value) (interp.Value, error) {
	var to, msg string
	if v, ok := named["to"]; ok {
		to = v.Display()
	}
	if v, ok := named["message"]; ok {
		msg = v.Display()
	}
	// Positional fallback: whisper(target, message).
	if to == "" && len(args) >= 1 {
		to = args[0].Display()
	}
	if msg == "" && len(args) >= 2 {
		msg = args[1].Display()
	}
	if to == "" || msg == "" {
		return nil, errf("whisper requires target and message")
	}
	if err := h.PrivateMessage(ctx, to, msg); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslAddFriend wraps Host.AddFriend — adds a player to the friend
// list so PMs can flow. Returns Result{Null} on success.
func dslAddFriend(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("add_friend takes 1 arg (name), got %d", len(args))
	}
	name, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("add_friend: name must be String, got %s", args[0].Kind())
	}
	if err := h.AddFriend(ctx, string(name)); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

// dslFindOption returns the 1-based index of the first dialog option
// whose text contains `needle` (case-insensitive substring), or 0
// if no match. Returns 0 (not -1) so routines can write
// `answer(find_option("Yes"))` and have the server respond
// gracefully when nothing matches.
//
// Dialog options must have been surfaced via world.dialog.options
// (set by the NPC dialog menu handler); call wait_for_dialog first.
func dslFindOption(_ context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("find_option takes 1 arg (needle text), got %d", len(args))
	}
	needle, ok := args[0].(interp.String)
	if !ok {
		return nil, errf("find_option: needle must be String, got %s", args[0].Kind())
	}
	rec := h.world.Recent.DialogOptions()
	if rec == nil {
		return interp.Int(0), nil
	}
	lower := strings.ToLower(string(needle))
	for i, opt := range rec.Options {
		if strings.Contains(strings.ToLower(opt), lower) {
			return interp.Int(int64(i + 1)), nil
		}
	}
	return interp.Int(0), nil
}
