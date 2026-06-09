// Command orsc-ctl is a thin CLI over the OpenRSC Admin API. It speaks HTTP+JSON
// to the admin server, unwrapping the {ok,data,error} response envelope, and
// authenticates with the X-Admin-Token header.
//
// Commands self-register into a grouped registry, so each command group lives in
// its own file (server.go, world.go, player_*.go, health.go) with no edits here.
//
//	orsc-ctl health                                  # API + game-server liveness
//	orsc-ctl server stats                            # live tick/heap/queue stats
//	orsc-ctl server save-all                         # force-save every player
//	orsc-ctl world announce "Restart in 5 minutes"   # broadcast announcement
//	orsc-ctl player mute Delores -minutes 30         # (downstream group)
//	orsc-ctl -json server uptime                     # raw JSON output
//
// Auth: -token (default $OPENRSC_ADMIN_API_TOKEN). Base URL: -base (default
// $OPENRSC_ADMIN_API_BASE, else http://127.0.0.1:43099/admin/v1).
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"time"
)

const defaultBase = "http://127.0.0.1:43099/admin/v1"

// jsonOut toggles raw JSON output; honored by printResult. It is a package-level
// flag so command handlers can render via printResult without threading it.
var jsonOut *bool

// cmd is one registered subcommand. run receives the constructed Client and the
// args that follow "<group> <name>" (or just "<name>" for the top-level group).
type cmd struct {
	name  string
	usage string
	help  string
	run   func(c *Client, args []string) error
}

// registry maps a group name to its commands. Group "" is the top-level group
// (invoked as `orsc-ctl <name>`); other groups are `orsc-ctl <group> <name>`.
var registry = map[string][]cmd{}

// register adds a command to a group. Call it from an init() in each group file.
func register(group string, c cmd) {
	registry[group] = append(registry[group], c)
}

func main() {
	base := flag.String("base", envOr("OPENRSC_ADMIN_API_BASE", defaultBase), "admin API base URL (default $OPENRSC_ADMIN_API_BASE)")
	token := flag.String("token", os.Getenv("OPENRSC_ADMIN_API_TOKEN"), "admin token (default $OPENRSC_ADMIN_API_TOKEN)")
	jsonOut = flag.Bool("json", false, "emit raw JSON instead of pretty output")
	timeout := flag.Duration("timeout", 15*time.Second, "per-request timeout")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 || args[0] == "help" || args[0] == "-h" || args[0] == "--help" {
		usage()
		if len(args) == 0 {
			os.Exit(2)
		}
		return
	}

	group, name, rest := resolve(args)
	c := findCmd(group, name)
	if c == nil {
		fmt.Fprintf(os.Stderr, "orsc-ctl: unknown command %q\n\n", joinCmd(group, name))
		usage()
		os.Exit(2)
	}

	client := NewClient(*base, *token, *timeout)
	if err := c.run(client, rest); err != nil {
		fail(err)
	}
}

// resolve maps the raw args onto (group, name, rest). It first tries the args as
// a top-level command (group ""), then as a "<group> <name>" pair.
func resolve(args []string) (group, name string, rest []string) {
	// Top-level command: orsc-ctl <name> [args...]
	if findCmd("", args[0]) != nil {
		return "", args[0], args[1:]
	}
	// Grouped command: orsc-ctl <group> <name> [args...]
	if _, ok := registry[args[0]]; ok && len(args) >= 2 {
		return args[0], args[1], args[2:]
	}
	// Fall through with whatever we have so findCmd reports "unknown".
	if len(args) >= 2 {
		return args[0], args[1], args[2:]
	}
	return args[0], "", nil
}

// findCmd looks up a command by group + name.
func findCmd(group, name string) *cmd {
	for i := range registry[group] {
		if registry[group][i].name == name {
			return &registry[group][i]
		}
	}
	return nil
}

func joinCmd(group, name string) string {
	if group == "" {
		return name
	}
	return group + " " + name
}

// usage prints a grouped listing of every registered command and its usage line.
func usage() {
	fmt.Fprintf(os.Stderr, `orsc-ctl [-base URL] [-token TOKEN] [-json] [-timeout DUR] <command> [args]

A CLI for the OpenRSC Admin API.

`)
	groups := make([]string, 0, len(registry))
	for g := range registry {
		groups = append(groups, g)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i] == "" {
			return true // top-level first
		}
		if groups[j] == "" {
			return false
		}
		return groups[i] < groups[j]
	})
	for _, g := range groups {
		cmds := append([]cmd(nil), registry[g]...)
		sort.Slice(cmds, func(i, j int) bool { return cmds[i].name < cmds[j].name })
		if g == "" {
			fmt.Fprintln(os.Stderr, "Top-level:")
		} else {
			fmt.Fprintf(os.Stderr, "%s:\n", g)
		}
		for _, c := range cmds {
			fmt.Fprintf(os.Stderr, "  %-40s %s\n", joinCmd(g, c.usage), c.help)
		}
		fmt.Fprintln(os.Stderr)
	}
	fmt.Fprintf(os.Stderr, "Auth: $OPENRSC_ADMIN_API_TOKEN (or -token). Base: $OPENRSC_ADMIN_API_BASE (or -base, default %s).\n", defaultBase)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// parseFlags parses fs against args with POSIX-style interspersed handling so
// flags and positionals may appear in any order. Go's flag package stops at the
// first non-flag token, which silently drops flags written AFTER a positional
// (e.g. `mute Delores -minutes 5` would never parse -minutes). parseFlags pulls
// positionals out of the way and keeps parsing the remaining flags, then leaves
// the collected positionals (in original order) available via fs.Args(). After
// it returns, callers read positionals from fs.Args() exactly as before.
//
// A literal "--" terminates flag parsing in the standard way: everything after
// it is taken verbatim as a positional (so an argument may begin with a dash).
func parseFlags(fs *flag.FlagSet, args []string) error {
	var positionals []string
	for len(args) > 0 {
		// A "--" terminator: the remainder are all positionals, verbatim.
		if args[0] == "--" {
			positionals = append(positionals, args[1:]...)
			break
		}
		if err := fs.Parse(args); err != nil {
			return err
		}
		rest := fs.Args()
		if len(rest) == 0 {
			break
		}
		// Parse stopped on the first leftover token because it is a positional
		// (or a bare "--"). Take it as a positional and resume after it.
		positionals = append(positionals, rest[0])
		args = rest[1:]
	}
	// Re-parse a synthetic list of just the positionals so fs.Args() returns
	// them in order. Prepend "--" so a positional that begins with a dash (one
	// collected after a "--" terminator) is taken verbatim rather than being
	// re-interpreted as an unknown flag. No flags remain to consume, so this
	// never errors.
	return fs.Parse(append([]string{"--"}, positionals...))
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "orsc-ctl:", err)
	os.Exit(1)
}
