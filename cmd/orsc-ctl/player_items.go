package main

import (
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Group "player" (inventory/bank/cache WRITE side): mutate a player's carried
// items, banked items, and live cache entries.
//
// Item endpoints (inventory + bank, add + remove) share the same body shape,
// per the spec example {"noted":false,"amount":1,"catalogId":10}: an item is
// identified by catalogId OR itemId (at least one required), with an amount and
// a noted flag. The add/remove verbs map to POST/DELETE on the same path.
//
// Cache set is a PATCH carrying {"type":"int","value":1} where type is optional
// (int|long|boolean|string) and value is the stored payload. Clears/deletes
// carry no body.

func init() {
	register("player", cmd{
		name:  "inv-add",
		usage: "inv-add <username> (-catalog-id N | -item-id N) [-amount N] [-noted]",
		help:  "add an item to a player's inventory",
		run: func(c *Client, args []string) error {
			username, body, err := itemArgs("player inv-add", args)
			if err != nil {
				return err
			}
			res, err := c.post("/players/"+url.PathEscape(username)+"/inventory/items", body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "inv-remove",
		usage: "inv-remove <username> (-catalog-id N | -item-id N) [-amount N] [-noted]",
		help:  "remove an item from a player's inventory (DELETE with body)",
		run: func(c *Client, args []string) error {
			username, body, err := itemArgs("player inv-remove", args)
			if err != nil {
				return err
			}
			res, err := c.del("/players/"+url.PathEscape(username)+"/inventory/items", body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "inv-clear",
		usage: "inv-clear <username>",
		help:  "remove every item from a player's inventory",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player inv-clear", args)
			if err != nil {
				return err
			}
			res, err := c.del("/players/"+url.PathEscape(username)+"/inventory", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "bank-add",
		usage: "bank-add <username> (-catalog-id N | -item-id N) [-amount N] [-noted]",
		help:  "add an item to a player's bank",
		run: func(c *Client, args []string) error {
			username, body, err := itemArgs("player bank-add", args)
			if err != nil {
				return err
			}
			res, err := c.post("/players/"+url.PathEscape(username)+"/bank/items", body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "bank-remove",
		usage: "bank-remove <username> (-catalog-id N | -item-id N) [-amount N] [-noted]",
		help:  "remove an item from a player's bank (DELETE with body)",
		run: func(c *Client, args []string) error {
			username, body, err := itemArgs("player bank-remove", args)
			if err != nil {
				return err
			}
			res, err := c.del("/players/"+url.PathEscape(username)+"/bank/items", body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "bank-clear",
		usage: "bank-clear <username>",
		help:  "remove every item from a player's bank",
		run: func(c *Client, args []string) error {
			username, err := usernameArg("player bank-clear", args)
			if err != nil {
				return err
			}
			res, err := c.del("/players/"+url.PathEscape(username)+"/bank", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "cache-set",
		usage: "cache-set <username> <key> -value V [-type int|long|boolean|string]",
		help:  "set a live cache entry on a player",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player cache-set", flag.ContinueOnError)
			value := fs.String("value", "", "cache value to store (required)")
			typ := fs.String("type", "", "value type: int, long, boolean, or string (optional)")
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			valueSet := false
			fs.Visit(func(f *flag.Flag) {
				if f.Name == "value" {
					valueSet = true
				}
			})
			rest := fs.Args()
			if len(rest) < 2 {
				return fmt.Errorf("usage: orsc-ctl player cache-set <username> <key> -value V [-type int|long|boolean|string]")
			}
			username, key := rest[0], rest[1]
			if !valueSet {
				return fmt.Errorf("player cache-set: -value is required")
			}

			body := map[string]any{"value": cacheValue(*typ, *value)}
			if t := strings.TrimSpace(*typ); t != "" {
				if !validCacheType(t) {
					return fmt.Errorf("player cache-set: -type must be one of int, long, boolean, string (got %q)", t)
				}
				body["type"] = t
			}

			path := "/players/" + url.PathEscape(username) + "/cache/" + url.PathEscape(key)
			res, err := c.patch(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("player", cmd{
		name:  "cache-del",
		usage: "cache-del <username> <key>",
		help:  "delete a live cache entry from a player",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player cache-del", flag.ContinueOnError)
			if err := parseFlags(fs, args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 2 {
				return fmt.Errorf("usage: orsc-ctl player cache-del <username> <key>")
			}
			username, key := rest[0], rest[1]
			path := "/players/" + url.PathEscape(username) + "/cache/" + url.PathEscape(key)
			res, err := c.del(path, nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}

// itemArgs parses the shared inventory/bank item-mutation body. The item is
// identified by -catalog-id OR -item-id (at least one, matching the spec note
// "Use catalogId or itemId"); -amount defaults to 1 and -noted defaults false.
// Only flags the caller actually set are sent, so add/remove honor the server's
// existing semantics for the chosen identifier.
func itemArgs(name string, args []string) (username string, body map[string]any, err error) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	catalogID := fs.Int("catalog-id", -1, "catalog (item) id to add/remove")
	itemID := fs.Int("item-id", -1, "item id to add/remove (alternative to -catalog-id)")
	amount := fs.Int("amount", 1, "item amount/quantity")
	noted := fs.Bool("noted", false, "treat the item as noted")
	if err = parseFlags(fs, args); err != nil {
		return "", nil, err
	}

	var catalogSet, itemSet bool
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "catalog-id":
			catalogSet = true
		case "item-id":
			itemSet = true
		}
	})

	rest := fs.Args()
	if len(rest) < 1 {
		return "", nil, fmt.Errorf("usage: orsc-ctl %s <username> (-catalog-id N | -item-id N) [-amount N] [-noted]", name)
	}
	username = rest[0]

	if !catalogSet && !itemSet {
		return "", nil, fmt.Errorf("%s: one of -catalog-id or -item-id is required", name)
	}

	body = map[string]any{
		"amount": *amount,
		"noted":  *noted,
	}
	if catalogSet {
		body["catalogId"] = *catalogID
	}
	if itemSet {
		body["itemId"] = *itemID
	}
	return username, body, nil
}

// validCacheType reports whether t is one of the cache value types the spec
// enumerates for cache-set.
func validCacheType(t string) bool {
	switch t {
	case "int", "long", "boolean", "string":
		return true
	}
	return false
}

// cacheValue coerces the -value string into the JSON shape implied by -type so
// the body matches the spec example ({"type":"int","value":1}): numeric for
// int/long, a real bool for boolean, otherwise (and on parse failure) a string.
func cacheValue(typ, value string) any {
	switch strings.TrimSpace(typ) {
	case "int", "long":
		if n, err := strconv.ParseInt(value, 10, 64); err == nil {
			return n
		}
	case "boolean":
		if b, err := strconv.ParseBool(value); err == nil {
			return b
		}
	}
	return value
}
