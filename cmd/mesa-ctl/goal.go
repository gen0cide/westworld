package main

import (
	"context"
	"fmt"
	"strings"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// goalCmd implements `mesa-ctl goal push <goal> [--match <glob>]`: set a SOFT
// runtime goal override on the matching, currently-running hosts via mesad's
// Provision.Subscribe push stream — no restart, no recompile. The director
// prefers the pushed goal over its genesis/persona goal until it's replaced.
func goalCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: mesa-ctl goal push <goal> [--match <glob>]")
	}
	switch args[0] {
	case "push":
		return goalPush(ctx, c, args[1:])
	default:
		return fmt.Errorf("unknown goal subcommand %q (want: push)", args[0])
	}
}

func goalPush(ctx context.Context, c mesapb.AdminClient, args []string) error {
	var parts []string
	match := ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--match", "-match":
			if i+1 >= len(args) {
				return fmt.Errorf("--match needs a glob (e.g. 'drone*')")
			}
			i++
			match = args[i]
		default:
			parts = append(parts, args[i])
		}
	}
	goal := strings.TrimSpace(strings.Join(parts, " "))
	if goal == "" {
		return fmt.Errorf("usage: mesa-ctl goal push <goal> [--match <glob>]")
	}

	res, err := c.PushGoal(ctx, &mesapb.PushGoalRequest{Goal: goal, Match: match})
	if err != nil {
		return err
	}
	scope := "all hosts"
	if match != "" {
		scope = "match " + match
	}
	fmt.Printf("goal pushed (%s): set on %d, delivered live to %d\n", scope, res.Matched, res.Pushed)
	if n := len(res.HostIds); n > 0 && n <= 20 {
		fmt.Printf("  → %s\n", strings.Join(res.HostIds, ", "))
	}
	fmt.Printf("  goal: %s\n", goal)
	return nil
}
