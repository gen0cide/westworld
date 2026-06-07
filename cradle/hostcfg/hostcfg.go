// Package hostcfg defines the *.hostcfg configuration surface for the cradle and
// loads it from YAML or JSON into runtime.HostConfig values.
//
// A *.hostcfg document is EITHER a single host (top-level host fields) or a host
// SET (a `defaults:` block plus a `hosts:` list and/or a `template:` for a
// drone[1..N] batch). The cradle can be pointed at a single file, one big
// multi-host file, or a directory of *.hostcfg files.
//
// SECURITY: a host's password is NEVER written inline. The file names the
// environment variable that holds it (`password_env`, default WESTWORLD_PASSWORD);
// the secret is resolved at launch via ResolvePassword and never parsed, logged,
// or marshalled. An inline `password:` key is rejected (strict unknown-field
// parsing).
package hostcfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gen0cide/westworld/runtime"
	"gopkg.in/yaml.v3"
)

// nameRe constrains a host name to a safe charset: it becomes a filename
// (<name>.db, aliases.json, the JSONL log), a data-dir leaf, a URL path segment
// (/api/hosts/<name>/...), and the mesa host key. A '/' or '..' would traverse
// paths or make the host's debug surface unroutable.
var nameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// Defaults applied when a field is omitted.
const (
	DefaultServer      = "localhost:43594"
	DefaultPasswordEnv = "WESTWORLD_PASSWORD"
)

// State store modes (per-host local KV).
const (
	StateFile   = "file"   // durable bbolt file (production cradle mode)
	StateMemory = "memory" // ephemeral in-memory store (no durable state, no mesa memory mirror)
)

// Supervision policies on host exit/crash.
const (
	SuperviseRestart = "restart" // auto-restart with backoff (soak tests)
	SuperviseHold    = "hold"    // leave dead + flagged for inspection (debug default)
)

// Host is one host's authored configuration. Zero-value fields fall back to
// package defaults (server) or are merged from a set's `defaults:` block.
type Host struct {
	Name        string   `yaml:"name" json:"name"`                                     // identity; must be unique in the loaded set
	Server      string   `yaml:"server,omitempty" json:"server,omitempty"`             // OpenRSC host:port (default DefaultServer)
	PasswordEnv string   `yaml:"password_env,omitempty" json:"password_env,omitempty"` // env var holding the secret (default WESTWORLD_PASSWORD)
	Mesa        string   `yaml:"mesa,omitempty" json:"mesa,omitempty"`                 // per-host mesa instance host:port; "" => offline
	Goal        string   `yaml:"goal,omitempty" json:"goal,omitempty"`                 // autonomous north-star (requires mesa)
	Director    Director `yaml:"director,omitempty" json:"director,omitempty"`         // scripted routine(s) when not autonomous
	State       string   `yaml:"state,omitempty" json:"state,omitempty"`               // file | memory (default file)
	DataDir     string   `yaml:"data_dir,omitempty" json:"data_dir,omitempty"`         // override store/alias dir; "" => ~/.westworld/hosts/<name>
	Genesis     *bool    `yaml:"genesis,omitempty" json:"genesis,omitempty"`           // run login Genesis when mesa-linked (default true)
	TurnTimeout Duration `yaml:"turn_timeout,omitempty" json:"turn_timeout,omitempty"`
	Settle      Duration `yaml:"settle,omitempty" json:"settle,omitempty"`
	Supervision string   `yaml:"supervision,omitempty" json:"supervision,omitempty"` // restart | hold (default hold)
}

// Director is the fixed-routine spec used when the host is not running the
// autonomous Act planner. (The interactive REPL is intentionally not exposed —
// a daemon host has no stdin.)
type Director struct {
	Routine  string   `yaml:"routine,omitempty" json:"routine,omitempty"`   // single routine file (looped with Loop)
	Routines []string `yaml:"routines,omitempty" json:"routines,omitempty"` // ordered routine files, run as a sequence
	Loop     bool     `yaml:"loop,omitempty" json:"loop,omitempty"`         // with Routine, repeat forever
}

// PasswordEnvName is the env var this host resolves its password from.
func (h Host) PasswordEnvName() string {
	if h.PasswordEnv != "" {
		return h.PasswordEnv
	}
	return DefaultPasswordEnv
}

// ResolvePassword reads the host's password from its environment variable. The
// value is never stored on the Host or logged. An empty value is an error so a
// misconfigured host fails fast rather than attempting a bad login.
func (h Host) ResolvePassword() (string, error) {
	name := h.PasswordEnvName()
	v := os.Getenv(name)
	if v == "" {
		return "", fmt.Errorf("host %q: password env %q is empty", h.Name, name)
	}
	return v, nil
}

// SupervisionPolicy is the host's restart policy, defaulting to hold (keep a dead
// host visible for inspection rather than restart-looping a crash).
func (h Host) SupervisionPolicy() string {
	if h.Supervision == "" {
		return SuperviseHold
	}
	return h.Supervision
}

// Autonomous reports whether this host runs the mesa Act planner rather than a
// fixed-routine director: either an explicit goal, or mesa-connected with no fixed
// routine (the goal is compiled at login by genesis + the persona north-star).
func (h Host) Autonomous() bool {
	if strings.TrimSpace(h.Goal) != "" {
		return true
	}
	if h.Director.Routine != "" || len(h.Director.Routines) > 0 {
		return false
	}
	return h.Mesa != ""
}

// Validate checks one host in isolation. ValidateSet additionally enforces
// cross-host name uniqueness.
func (h Host) Validate() error {
	if strings.TrimSpace(h.Name) == "" {
		return errors.New("host: name is required")
	}
	if !nameRe.MatchString(h.Name) {
		return fmt.Errorf("host %q: name must match %s (it becomes a filename, URL segment, and DB key)", h.Name, nameRe.String())
	}
	if h.Director.Routine != "" && len(h.Director.Routines) > 0 {
		return fmt.Errorf("host %q: set director.routine OR director.routines, not both", h.Name)
	}
	hasGoal := strings.TrimSpace(h.Goal) != ""
	hasDirector := h.Director.Routine != "" || len(h.Director.Routines) > 0
	hasMesa := h.Mesa != ""
	switch {
	case hasGoal && !hasMesa:
		return fmt.Errorf("host %q: a goal requires a mesa address (autonomous play needs mesa)", h.Name)
	case !hasGoal && !hasDirector && !hasMesa:
		return fmt.Errorf("host %q: needs a goal (autonomous), director.routine/routines (scripted), or a mesa address (genesis-driven autonomy)", h.Name)
	}
	if h.State != "" && h.State != StateFile && h.State != StateMemory {
		return fmt.Errorf("host %q: state must be %q or %q, got %q", h.Name, StateFile, StateMemory, h.State)
	}
	if h.Supervision != "" && h.Supervision != SuperviseRestart && h.Supervision != SuperviseHold {
		return fmt.Errorf("host %q: supervision must be %q or %q, got %q", h.Name, SuperviseRestart, SuperviseHold, h.Supervision)
	}
	return nil
}

// ValidateSet validates every host and enforces unique names across the set
// (username drives the data-dir, bbolt file, alias store, and mesa host key — a
// collision corrupts shared state and fights over one RSC session).
func ValidateSet(hosts []Host) error {
	if len(hosts) == 0 {
		return errors.New("no hosts loaded")
	}
	seen := make(map[string]bool, len(hosts))
	for _, h := range hosts {
		if err := h.Validate(); err != nil {
			return err
		}
		if seen[h.Name] {
			return fmt.Errorf("duplicate host name %q (names must be unique across the loaded set)", h.Name)
		}
		seen[h.Name] = true
	}
	return nil
}

// ToHostConfig maps an authored host into a runtime.HostConfig, given the already
// resolved password. Logger and Debug are left nil — the cradle wires the
// process child-logger and (mounted) debug surface itself.
func (h Host) ToHostConfig(password string) runtime.HostConfig {
	genesis := true
	if h.Genesis != nil {
		genesis = *h.Genesis
	}
	server := h.Server
	if server == "" {
		server = DefaultServer
	}
	return runtime.HostConfig{
		Server:   server,
		Username: h.Name,
		Password: password,
		Goal:     h.Goal,
		Mesa:     h.Mesa,
		Director: runtime.DirectorSpec{
			Routine:  h.Director.Routine,
			Routines: h.Director.Routines,
			Loop:     h.Director.Loop,
		},
		DataDir:     h.DataDir,
		Fresh:       h.State == StateMemory, // memory => ephemeral store + no mesa memory mirror (drone mode)
		Genesis:     genesis,
		TurnTimeout: time.Duration(h.TurnTimeout),
		Settle:      time.Duration(h.Settle),
	}
}

// Duration is a time.Duration that (un)marshals as a Go duration string ("2m",
// "500ms") in both YAML and JSON, instead of an opaque nanosecond integer.
type Duration time.Duration

func parseDuration(s string) (Duration, error) {
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return Duration(d), nil
}

func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	var s string
	if err := node.Decode(&s); err != nil {
		return err
	}
	v, err := parseDuration(s)
	if err != nil {
		return err
	}
	*d = v
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) { return time.Duration(d).String(), nil }

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	v, err := parseDuration(s)
	if err != nil {
		return err
	}
	*d = v
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) { return json.Marshal(time.Duration(d).String()) }
