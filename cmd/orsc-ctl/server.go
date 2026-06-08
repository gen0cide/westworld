package main

// Group "server": server-wide stats and operations.

func init() {
	register("server", cmd{
		name:  "stats",
		usage: "stats",
		help:  "live tick timing, player/NPC counts, heap, admin queue depth",
		run: func(c *Client, args []string) error {
			res, err := c.get("/server/stats", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("server", cmd{
		name:  "uptime",
		usage: "uptime",
		help:  "server uptime in nanoseconds and milliseconds",
		run: func(c *Client, args []string) error {
			res, err := c.get("/server/uptime", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("server", cmd{
		name:  "save-all",
		usage: "save-all",
		help:  "queue a forced save for every online player",
		run: func(c *Client, args []string) error {
			res, err := c.post("/server/save-all", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("server", cmd{
		name:  "ip-counts",
		usage: "ip-counts",
		help:  "map of current IP address to online player count",
		run: func(c *Client, args []string) error {
			res, err := c.get("/server/ip-counts", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	register("server", cmd{
		name:  "player-counts",
		usage: "player-counts",
		help:  "online player totals and unique IP counts",
		run: func(c *Client, args []string) error {
			res, err := c.get("/server/player-counts", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}
