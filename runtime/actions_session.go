package runtime

import (
	"context"

	"github.com/gen0cide/westworld/dsl/interp"
)

// Session & admin action handler bodies (command / logout). Registered
// in the central actionHandlers table in dsl_actions.go.

// dslCommand sends a server admin command via the dedicated command
// opcode (38), NOT public chat. The DSL passes the command WITHOUT
// the leading "::" — that prefix is the in-game UI convention; on
// the wire it's a distinct opcode. Common commands: tele <x> <y>,
// summon <name>, blink, invisible. Requires admin permissions on
// the host's account; non-admins get rejected by the server.
func dslCommand(ctx context.Context, h *Host, args []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if len(args) != 1 {
		return nil, errf("command takes 1 argument (cmd), got %d", len(args))
	}
	cmd := args[0].Display()
	if err := h.Command(ctx, cmd); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}

func dslLogout(ctx context.Context, h *Host, _ []interp.Value, _ map[string]interp.Value) (interp.Value, error) {
	if err := h.Logout(ctx); err != nil {
		return wrapServerErr(err), nil
	}
	return interp.Ok(interp.Null{}), nil
}
