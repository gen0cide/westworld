package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/gen0cide/westworld/cradle/hostcfg"
	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// fleetOpts are the cradle-side fields the generator writes; the host SET comes
// from mesa's persona registry, so names line up by construction.
type fleetOpts struct {
	Server      string
	Mesa        string
	PasswordEnv string
	State       string
	Operator    string
	Goal        string            // shared goal → defaults block
	Goals       map[string]string // per-host goal → host entry (overrides shared)
	Genesis     *bool             // nil => leave unset (cradle default)
}

// genFleet builds a cradle hostcfg (YAML) for the given host_ids: a `defaults:`
// block plus one `hosts:` entry per id (name == host_id), so the cradle fleet
// lines up with the mesa persona registry by construction. The output round-trips
// through hostcfg.Load.
func genFleet(hostIDs []string, o fleetOpts) ([]byte, error) {
	if len(hostIDs) == 0 {
		return nil, fmt.Errorf("no host_ids to generate (no personas matched)")
	}
	defaults := hostcfg.Host{
		Server:      o.Server,
		Mesa:        o.Mesa,
		PasswordEnv: o.PasswordEnv,
		State:       o.State,
		Operator:    o.Operator,
		Genesis:     o.Genesis,
	}
	if len(o.Goals) == 0 {
		defaults.Goal = o.Goal // shared (or empty → goal rides from the persona/mesa)
	}
	hosts := make([]hostcfg.Host, 0, len(hostIDs))
	for _, id := range hostIDs {
		h := hostcfg.Host{Name: id}
		if g, ok := o.Goals[id]; ok {
			h.Goal = g
		}
		hosts = append(hosts, h)
	}
	doc := struct {
		Defaults hostcfg.Host   `yaml:"defaults"`
		Hosts    []hostcfg.Host `yaml:"hosts"`
	}{Defaults: defaults, Hosts: hosts}
	return yaml.Marshal(doc)
}

// fleetCmd implements `mesa-ctl fleet gen`: read the registered persona set from
// mesa and emit a matching cradle hostcfg on stdout.
func fleetCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) == 0 || args[0] != "gen" {
		return fmt.Errorf("usage: mesa-ctl fleet gen [flags]  (the only subcommand)")
	}
	fs := flag.NewFlagSet("fleet gen", flag.ContinueOnError)
	server := fs.String("server", "localhost:43594", "OpenRSC server host:port (defaults block)")
	mesa := fs.String("mesa", "localhost:7077", "mesad host:port (defaults block)")
	pwEnv := fs.String("password-env", "WESTWORLD_PASSWORD", "env var holding the shared account password")
	state := fs.String("state", "", "file|memory (omit for the cradle default)")
	operator := fs.String("operator", "", "in-game operator username for the ANALYSIS trigger")
	goal := fs.String("goal", "", "shared autonomous goal for every host")
	goalsFile := fs.String("goals", "", "path to a JSON map {host_id: goal} for per-host goals")
	match := fs.String("match", "", "only include host_ids matching this glob (e.g. 'drone*')")
	genesis := fs.String("genesis", "", "true|false to set genesis in defaults (omit to leave unset)")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	list, err := c.ListPersonas(ctx, &mesapb.ListPersonasRequest{})
	if err != nil {
		return err
	}
	var ids []string
	for _, p := range list.Personas {
		if *match != "" {
			ok, err := filepath.Match(*match, p.HostId)
			if err != nil {
				return fmt.Errorf("bad --match pattern: %w", err)
			}
			if !ok {
				continue
			}
		}
		ids = append(ids, p.HostId)
	}

	var goals map[string]string
	if *goalsFile != "" {
		raw, err := os.ReadFile(*goalsFile)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(raw, &goals); err != nil {
			return fmt.Errorf("goals file: %w", err)
		}
	}

	var gen *bool
	switch *genesis {
	case "true":
		t := true
		gen = &t
	case "false":
		f := false
		gen = &f
	case "":
		// leave unset
	default:
		return fmt.Errorf("--genesis must be true or false")
	}

	out, err := genFleet(ids, fleetOpts{
		Server: *server, Mesa: *mesa, PasswordEnv: *pwEnv, State: *state,
		Operator: *operator, Goal: *goal, Goals: goals, Genesis: gen,
	})
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(out)
	return err
}
