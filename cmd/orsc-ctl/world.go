package main

import (
	"flag"
	"fmt"
	"strings"
)

// Group "world": broadcast messages to every online player. Both endpoints take
// a JSON body of {"message": "<text>"}. The text is accepted as positional args
// (joined with spaces) or via -message.

func init() {
	register("world", cmd{
		name:  "announce",
		usage: "announce <message>",
		help:  "broadcast an announcement chat message to all online players",
		run: func(c *Client, args []string) error {
			msg, err := messageArg("world announce", args)
			if err != nil {
				return err
			}
			res, err := c.post("/world/announcement", map[string]any{"message": msg})
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("world", cmd{
		name:  "system-message",
		usage: "system-message <message>",
		help:  "broadcast a system-style message/message box to all online players",
		run: func(c *Client, args []string) error {
			msg, err := messageArg("world system-message", args)
			if err != nil {
				return err
			}
			res, err := c.post("/world/system-message", map[string]any{"message": msg})
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}

// messageArg resolves the broadcast text from a -message flag or, failing that,
// the remaining positional args joined with spaces.
func messageArg(name string, args []string) (string, error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	msg := fs.String("message", "", "message text (else positional args)")
	if err := fs.Parse(args); err != nil {
		return "", err
	}
	if *msg == "" {
		*msg = strings.Join(fs.Args(), " ")
	}
	if strings.TrimSpace(*msg) == "" {
		return "", fmt.Errorf("usage: orsc-ctl %s <message>  (or -message \"...\")", name)
	}
	return *msg, nil
}
