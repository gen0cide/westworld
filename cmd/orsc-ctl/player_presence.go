package main

import (
	"flag"
	"fmt"
	"net/url"
	"strings"
)

// Group "player": moderation, presence, and movement actions against a single
// online player (identified by username path param) plus a couple of fleet-wide
// summon/return verbs. Request body field names follow the OpenRSC Admin API
// spec examples (loose {"type":"object"} schemas — the examples are canonical):
//
//	alert        {"message": "..."}                       message required
//	kick         {"reason": "..."}                        reason optional
//	mute         {"shadow":false,"minutes":30,"notify":true}
//	global-mute  {"shadow":false,"minutes":30,"notify":true}
//	teleport     {"x":122,"y":503,"showBubble":true}      x,y required
//	summon-to    {"toPlayer":"alex"} or {"x":..,"y":..}
//	summon-all   {"x":122,"y":503,"limit":25} or {"toPlayer":..}
//
// jail/release/unmute/global-unmute/return/return-all carry no body.

func init() {
	// alert <username> <message...> — message box + quest-channel admin message.
	// The text may be given as trailing positional args or via -message.
	register("player", cmd{
		name:  "alert",
		usage: "alert <username> <message> (or -message \"...\")",
		help:  "send an alert/message box to a player (message required)",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player alert", flag.ContinueOnError)
			message := fs.String("message", "", "alert text (else trailing positional args)")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player alert <username> <message>  (or -message \"...\")")
			}
			username := rest[0]
			msg := *message
			if msg == "" {
				msg = strings.Join(rest[1:], " ")
			}
			if strings.TrimSpace(msg) == "" {
				return fmt.Errorf("player alert: message is required (positional or -message)")
			}
			path := "/players/" + url.PathEscape(username) + "/alert"
			res, err := c.post(path, map[string]any{"message": msg})
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// kick <username> [-reason "..."] — force unregister with optional reason.
	register("player", cmd{
		name:  "kick",
		usage: "kick <username> [-reason \"...\"]",
		help:  "force-disconnect a player (optional reason)",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player kick", flag.ContinueOnError)
			reason := fs.String("reason", "Administrative action", "kick reason")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player kick <username> [-reason \"...\"]")
			}
			path := "/players/" + url.PathEscape(rest[0]) + "/kick"
			res, err := c.post(path, map[string]any{"reason": *reason})
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// jail <username> — no body.
	register("player", cmd{
		name:  "jail",
		usage: "jail <username>",
		help:  "send a player to jail",
		run: func(c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: orsc-ctl player jail <username>")
			}
			path := "/players/" + url.PathEscape(args[0]) + "/jail"
			res, err := c.post(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// release <username> — no body.
	register("player", cmd{
		name:  "release",
		usage: "release <username>",
		help:  "release a player from jail",
		run: func(c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: orsc-ctl player release <username>")
			}
			path := "/players/" + url.PathEscape(args[0]) + "/release"
			res, err := c.post(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// mute <username> [-minutes N] [-shadow] [-notify] — minutes=-1 is permanent.
	register("player", cmd{
		name:  "mute",
		usage: "mute <username> [-minutes N] [-shadow] [-notify]",
		help:  "mute a player; minutes=-1 is permanent",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player mute", flag.ContinueOnError)
			minutes := fs.Int("minutes", 30, "mute duration in minutes (-1 = permanent)")
			shadow := fs.Bool("shadow", false, "shadow mute")
			notify := fs.Bool("notify", true, "notify the player")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player mute <username> [-minutes N] [-shadow] [-notify]")
			}
			path := "/players/" + url.PathEscape(rest[0]) + "/mute"
			body := map[string]any{
				"minutes": *minutes,
				"shadow":  *shadow,
				"notify":  *notify,
			}
			res, err := c.post(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// unmute <username> — DELETE, no body.
	register("player", cmd{
		name:  "unmute",
		usage: "unmute <username>",
		help:  "clear a player's normal mute",
		run: func(c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: orsc-ctl player unmute <username>")
			}
			path := "/players/" + url.PathEscape(args[0]) + "/mute"
			res, err := c.del(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// global-mute <username> [-minutes N] [-shadow] [-notify] — minutes=-1 permanent.
	register("player", cmd{
		name:  "global-mute",
		usage: "global-mute <username> [-minutes N] [-shadow] [-notify]",
		help:  "global-mute a player; minutes=-1 is permanent",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player global-mute", flag.ContinueOnError)
			minutes := fs.Int("minutes", 30, "global mute duration in minutes (-1 = permanent)")
			shadow := fs.Bool("shadow", false, "shadow mute")
			notify := fs.Bool("notify", true, "notify the player")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player global-mute <username> [-minutes N] [-shadow] [-notify]")
			}
			path := "/players/" + url.PathEscape(rest[0]) + "/global-mute"
			body := map[string]any{
				"minutes": *minutes,
				"shadow":  *shadow,
				"notify":  *notify,
			}
			res, err := c.post(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// global-unmute <username> — DELETE, no body.
	register("player", cmd{
		name:  "global-unmute",
		usage: "global-unmute <username>",
		help:  "clear a player's global mute",
		run: func(c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: orsc-ctl player global-unmute <username>")
			}
			path := "/players/" + url.PathEscape(args[0]) + "/global-mute"
			res, err := c.del(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// teleport <username> -x N -y N [-bubble] — x,y required.
	register("player", cmd{
		name:  "teleport",
		usage: "teleport <username> -x N -y N [-bubble]",
		help:  "teleport a player to x/y (x and y required)",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player teleport", flag.ContinueOnError)
			x := fs.Int("x", 0, "destination x coordinate (required)")
			y := fs.Int("y", 0, "destination y coordinate (required)")
			bubble := fs.Bool("bubble", true, "show the teleport bubble")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player teleport <username> -x N -y N [-bubble]")
			}
			if !flagSet(fs, "x") || !flagSet(fs, "y") {
				return fmt.Errorf("player teleport: -x and -y are required")
			}
			path := "/players/" + url.PathEscape(rest[0]) + "/teleport"
			body := map[string]any{
				"x":          *x,
				"y":          *y,
				"showBubble": *bubble,
			}
			res, err := c.post(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// summon <username> [-to-player NAME | -x N -y N] — summon to a player or x/y.
	register("player", cmd{
		name:  "summon",
		usage: "summon <username> [-to-player NAME | -x N -y N]",
		help:  "summon a player to another player or to x/y",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player summon", flag.ContinueOnError)
			toPlayer := fs.String("to-player", "", "summon target: an online player's username")
			x := fs.Int("x", 0, "destination x coordinate")
			y := fs.Int("y", 0, "destination y coordinate")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player summon <username> [-to-player NAME | -x N -y N]")
			}
			if flagSet(fs, "x") != flagSet(fs, "y") {
				return fmt.Errorf("player summon: -x and -y must be given together")
			}
			body := map[string]any{}
			switch {
			case *toPlayer != "":
				body["toPlayer"] = *toPlayer
			case flagSet(fs, "x") && flagSet(fs, "y"):
				body["x"] = *x
				body["y"] = *y
			default:
				return fmt.Errorf("player summon: provide -to-player NAME or -x N -y N")
			}
			path := "/players/" + url.PathEscape(rest[0]) + "/summon-to"
			res, err := c.post(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// return <username> — return-from-summon, no body.
	register("player", cmd{
		name:  "return",
		usage: "return <username>",
		help:  "return a player from a summon to their prior location",
		run: func(c *Client, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("usage: orsc-ctl player return <username>")
			}
			path := "/players/" + url.PathEscape(args[0]) + "/return-from-summon"
			res, err := c.post(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// summon-all [-to-player NAME | -x N -y N] [-limit N] — fleet-wide summon.
	register("player", cmd{
		name:  "summon-all",
		usage: "summon-all [-to-player NAME | -x N -y N] [-limit N]",
		help:  "summon all online players to another player or to x/y",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player summon-all", flag.ContinueOnError)
			toPlayer := fs.String("to-player", "", "summon target: an online player's username")
			x := fs.Int("x", 0, "destination x coordinate")
			y := fs.Int("y", 0, "destination y coordinate")
			limit := fs.Int("limit", 0, "cap how many players are moved (0 = no cap)")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			if flagSet(fs, "x") != flagSet(fs, "y") {
				return fmt.Errorf("player summon-all: -x and -y must be given together")
			}
			body := map[string]any{}
			switch {
			case *toPlayer != "":
				body["toPlayer"] = *toPlayer
			case flagSet(fs, "x") && flagSet(fs, "y"):
				body["x"] = *x
				body["y"] = *y
			default:
				return fmt.Errorf("player summon-all: provide -to-player NAME or -x N -y N")
			}
			if flagSet(fs, "limit") {
				body["limit"] = *limit
			}
			res, err := c.post("/players/summon-all", body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// return-all — return every summoned player, no body.
	register("player", cmd{
		name:  "return-all",
		usage: "return-all",
		help:  "return all summoned players to their prior locations",
		run: func(c *Client, args []string) error {
			res, err := c.post("/players/return-all", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}
