package main

// Top-level command (group ""): liveness check. Invoked as `orsc-ctl health`.

func init() {
	register("", cmd{
		name:  "health",
		usage: "health",
		help:  "basic API and game-server liveness",
		run: func(c *Client, args []string) error {
			res, err := c.get("/health", nil)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}
