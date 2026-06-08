// Command cradle-ctl is a thin CLI over the cradle-server control plane. Every
// operation is an HTTP/JSON call against -server, so the CLI and the web UI use
// the identical API.
//
//	cradle-ctl list
//	cradle-ctl status drone7
//	cradle-ctl pause drone7
//	cradle-ctl state drone7
//	cradle-ctl eval drone7 'say("hello")'
//	cradle-ctl tail drone7          # live thoughts + chat
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/coder/websocket"
)

func main() {
	server := flag.String("server", "localhost:8099", "cradle-server control-plane host:port")
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	c := &client{host: *server, base: "http://" + *server}
	cmd, rest := args[0], args[1:]

	var err error
	switch cmd {
	case "list":
		err = c.list()
	case "status":
		err = c.showJSON(hostPath(rest, ""))
	case "pause", "resume", "stop":
		err = c.control(cmd, name(rest))
	case "logoff", "logout":
		// Graceful: stop now triggers a clean RSC logout (host.Close → LogoutGraceful)
		// before disconnecting, so the server saves + releases the session.
		err = c.control("stop", name(rest))
	case "state":
		err = c.showJSON(hostPath(rest, "/debug/state"))
	case "events":
		path := hostPath(rest, "/debug/events")
		if len(rest) > 1 {
			path += "?kind=" + rest[1]
		}
		err = c.showJSON(path)
	case "eval":
		if len(rest) < 2 {
			err = fmt.Errorf("usage: cradle-ctl eval <name> <dsl...>")
		} else {
			err = c.body("/api/hosts/"+rest[0]+"/debug/eval", strings.Join(rest[1:], " "))
		}
	case "script":
		if len(rest) < 2 {
			err = fmt.Errorf("usage: cradle-ctl script <name> <file.rt>")
		} else {
			err = c.scriptFile(rest[0], rest[1])
		}
	case "tail":
		err = c.tail(name(rest))
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "cradle-ctl:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `cradle-ctl [-server host:port] <command> [args]

  list                     all hosts (status, position, HP, goal)
  status <name>            one host's status snapshot
  pause|resume|stop <name> lifecycle control
  logoff <name>            graceful RSC logout (clean save) then disconnect
  state <name>             live world snapshot (position, vitals, inventory, npcs, chat)
  events <name> [kind]     recent recorded events (optionally filtered by kind)
  eval <name> <dsl...>     run one DSL line against the live host
  script <name> <file.rt>  run a routine file against the live host
  tail <name>              stream live thoughts + chat until Ctrl-C
`)
}

type client struct {
	host string // raw host:port (for ws)
	base string // http://host:port
}

func (c *client) get(path string) ([]byte, int, error) {
	resp, err := http.Get(c.base + path)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return b, resp.StatusCode, nil
}

func (c *client) post(path, body string) ([]byte, int, error) {
	resp, err := http.Post(c.base+path, "application/json", strings.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return b, resp.StatusCode, nil
}

// hostStatus mirrors cradle.HostStatus (decoupled via JSON).
type hostStatus struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Goal       string `json:"goal"`
	Autonomous bool   `json:"autonomous"`
	Restarts   int    `json:"restarts"`
	Err        string `json:"err"`
	Live       bool   `json:"live"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
	HP         int    `json:"hp"`
	MaxHP      int    `json:"max_hp"`
}

func (c *client) list() error {
	b, code, err := c.get("/api/hosts")
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return serverErr(b)
	}
	var hosts []hostStatus
	if err := json.Unmarshal(b, &hosts); err != nil {
		return err
	}
	fmt.Printf("%-18s %-11s %-11s %-9s %s\n", "NAME", "STATUS", "POS", "HP", "GOAL")
	for _, h := range hosts {
		pos, hp := "-", "-"
		if h.Live {
			pos = fmt.Sprintf("%d,%d", h.X, h.Y)
			hp = fmt.Sprintf("%d/%d", h.HP, h.MaxHP)
		}
		goal := h.Goal
		if !h.Autonomous && goal == "" {
			goal = "(scripted)"
		}
		if h.Err != "" {
			goal = "ERR: " + h.Err
		}
		fmt.Printf("%-18s %-11s %-11s %-9s %s\n", h.Name, h.Status, pos, hp, truncate(goal, 60))
	}
	return nil
}

func (c *client) control(action, hostName string) error {
	b, code, err := c.post("/api/hosts/"+hostName+"/"+action, "")
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return serverErr(b)
	}
	fmt.Printf("%s: %s ok\n", hostName, action)
	return nil
}

func (c *client) showJSON(path string) error {
	b, code, err := c.get(path)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return serverErr(b)
	}
	return printJSON(b)
}

func (c *client) body(path, payload string) error {
	b, code, err := c.post(path, payload)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return serverErr(b)
	}
	return printJSON(b)
}

func (c *client) scriptFile(hostName, file string) error {
	src, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return c.body("/api/hosts/"+hostName+"/debug/script", string(src))
}

func (c *client) tail(hostName string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	url := "ws://" + c.host + "/api/hosts/" + hostName + "/debug/ws"
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return fmt.Errorf("connect %s: %w", url, err)
	}
	defer conn.CloseNow()
	fmt.Printf("tailing %s (Ctrl-C to stop)\n", hostName)
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		var rec struct {
			At   time.Time       `json:"at"`
			Kind string          `json:"kind"`
			Data json.RawMessage `json:"data"`
		}
		if json.Unmarshal(data, &rec) != nil {
			continue
		}
		if line := formatEvent(rec.Kind, rec.Data); line != "" {
			fmt.Printf("%s  %s\n", rec.At.Format("15:04:05"), line)
		}
	}
}

// formatEvent renders the interesting live events (thoughts + chat); other kinds
// are dropped from the tail.
func formatEvent(kind string, data json.RawMessage) string {
	// encoding/json matches keys case-insensitively, so these Go field names match
	// both the lowercase-tagged AgentThought and the untagged chat events.
	switch kind {
	case "agent_thought":
		var t struct{ Reasoning, Perception string }
		json.Unmarshal(data, &t)
		if t.Reasoning != "" {
			return "🧠 " + t.Reasoning
		}
		return ""
	case "chat_received":
		var m struct{ Speaker, Message string }
		json.Unmarshal(data, &m)
		if m.Speaker != "" {
			return "💬 " + m.Speaker + ": " + m.Message
		}
		return "💬 " + m.Message
	case "other_player_chat":
		var m struct{ MessageText string }
		json.Unmarshal(data, &m)
		return "💬 " + m.MessageText
	case "system_message":
		var m struct{ Message string }
		json.Unmarshal(data, &m)
		return "⚙  " + m.Message
	case "private_message":
		var m struct{ Sender, Message string }
		json.Unmarshal(data, &m)
		return "✉  " + m.Sender + ": " + m.Message
	}
	return ""
}

// ---- small helpers --------------------------------------------------------

func name(rest []string) string {
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "cradle-ctl: missing <name>")
		os.Exit(2)
	}
	return rest[0]
}

func hostPath(rest []string, suffix string) string {
	return "/api/hosts/" + name(rest) + suffix
}

func serverErr(b []byte) error {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(b, &e) == nil && e.Error != "" {
		return fmt.Errorf("%s", e.Error)
	}
	return fmt.Errorf("%s", strings.TrimSpace(string(b)))
}

func printJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		fmt.Println(string(b))
		return nil
	}
	out, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(out))
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
