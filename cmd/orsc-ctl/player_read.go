package main

import (
	"fmt"
	"net/url"
)

// Group "player" (read-only): inspect a player's account, location, skills,
// inventory, bank, quests, and persistent cache. Every command but `list` takes
// a <username> positional, URL-escaped into the path. None of these endpoints
// take a request body, so no body flags are exposed.

func init() {
	register("player", cmd{
		name:  "list",
		usage: "list",
		help:  "list all known players",
		run: func(c *Client, args []string) error {
			res, err := c.get("/players", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "get",
		usage: "get <username>",
		help:  "fetch a single player's account record",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player get", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username), nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "location",
		usage: "location <username>",
		help:  "fetch a player's current world location",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player location", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/location", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "skills",
		usage: "skills <username>",
		help:  "fetch a player's skill levels and experience",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player skills", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/skills", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "inventory",
		usage: "inventory <username>",
		help:  "fetch a player's inventory contents",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player inventory", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/inventory", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "bank",
		usage: "bank <username>",
		help:  "fetch a player's bank contents",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player bank", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/bank", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "quests",
		usage: "quests <username>",
		help:  "fetch a player's quest progress",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player quests", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/quests", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "cache",
		usage: "cache <username>",
		help:  "fetch a player's full persistent cache (key/value store)",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player cache", args)
			if err != nil {
				return err
			}
			res, err := c.get("/players/"+url.PathEscape(username)+"/cache", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "cache-get",
		usage: "cache-get <username> <key>",
		help:  "fetch a single key from a player's persistent cache",
		run: func(c *Client, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("usage: orsc-ctl player cache-get <username> <key>")
			}
			username := args[0]
			key := args[1]
			path := "/players/" + url.PathEscape(username) + "/cache/" + url.PathEscape(key)
			res, err := c.get(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}

// usernameArg pulls the required <username> positional from args, returning a
// clear usage error when it is missing.
func usernameArg(name string, args []string) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("usage: orsc-ctl %s <username>", name)
	}
	return args[0], nil
}
