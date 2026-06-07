package hostcfg

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// setFile is the host-SET form of a document: shared defaults plus an explicit
// host list and/or a drone[1..N] template.
type setFile struct {
	Defaults *Host     `yaml:"defaults" json:"defaults"`
	Hosts    []Host    `yaml:"hosts" json:"hosts"`
	Template *Template `yaml:"template" json:"template"`
}

// Template generates Count hosts named <NamePrefix><n> (n from Start, default 1),
// each built from the set's defaults overlaid with Spec.
type Template struct {
	NamePrefix string `yaml:"name_prefix" json:"name_prefix"`
	Count      int    `yaml:"count" json:"count"`
	Start      int    `yaml:"start,omitempty" json:"start,omitempty"`
	Spec       Host   `yaml:"spec" json:"spec"`
}

// Expand flattens a set into concrete hosts: each explicit host and each
// templated drone is the defaults overlaid with its own fields.
func (sf setFile) Expand() ([]Host, error) {
	var base Host
	if sf.Defaults != nil {
		base = *sf.Defaults
	}
	var out []Host
	for _, h := range sf.Hosts {
		out = append(out, applyDefaults(base, h))
	}
	if sf.Template != nil {
		t := sf.Template
		if t.NamePrefix == "" {
			return nil, errors.New("template: name_prefix is required")
		}
		if t.Count <= 0 {
			return nil, fmt.Errorf("template: count must be > 0, got %d", t.Count)
		}
		start := t.Start
		if start == 0 {
			start = 1
		}
		for i := 0; i < t.Count; i++ {
			h := applyDefaults(base, t.Spec)
			h.Name = fmt.Sprintf("%s%d", t.NamePrefix, start+i)
			out = append(out, h)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("config set defines no hosts (need hosts: and/or template:)")
	}
	return out, nil
}

// applyDefaults overlays h's non-zero fields onto base (h wins).
func applyDefaults(base, h Host) Host {
	out := base
	if h.Name != "" {
		out.Name = h.Name
	}
	if h.Server != "" {
		out.Server = h.Server
	}
	if h.PasswordEnv != "" {
		out.PasswordEnv = h.PasswordEnv
	}
	if h.Mesa != "" {
		out.Mesa = h.Mesa
	}
	if h.Goal != "" {
		out.Goal = h.Goal
	}
	if h.Director.Routine != "" {
		out.Director.Routine = h.Director.Routine
	}
	if len(h.Director.Routines) > 0 {
		out.Director.Routines = h.Director.Routines
	}
	if h.Director.Loop {
		out.Director.Loop = true
	}
	if h.State != "" {
		out.State = h.State
	}
	if h.DataDir != "" {
		out.DataDir = h.DataDir
	}
	if h.Genesis != nil {
		out.Genesis = h.Genesis
	}
	if h.TurnTimeout != 0 {
		out.TurnTimeout = h.TurnTimeout
	}
	if h.Settle != 0 {
		out.Settle = h.Settle
	}
	if h.Supervision != "" {
		out.Supervision = h.Supervision
	}
	return out
}

// Load reads hosts from a file or a directory of *.hostcfg/.yaml/.yml/.json
// files. It does NOT validate cross-host invariants — call ValidateSet on the
// result before launching.
func Load(path string) ([]Host, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return loadDir(path)
	}
	return loadFile(path)
}

func loadDir(dir string) ([]Host, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() || !isConfigFile(e.Name()) {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names) // deterministic load order
	var all []Host
	for _, n := range names {
		hs, err := loadFile(filepath.Join(dir, n))
		if err != nil {
			return nil, err
		}
		all = append(all, hs...)
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no *.hostcfg/.yaml/.yml/.json files with hosts in %s", dir)
	}
	return all, nil
}

func loadFile(path string) ([]Host, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	hosts, err := decodeDoc(raw, isJSONFile(path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filepath.Base(path), err)
	}
	return hosts, nil
}

// decodeDoc parses one document's bytes into hosts, choosing the single-host or
// set form by inspecting its top-level keys. jsonFmt selects strict JSON parsing
// (otherwise strict YAML; JSON is a YAML subset so .yaml files may contain JSON).
func decodeDoc(raw []byte, jsonFmt bool) ([]Host, error) {
	set, err := isSetDoc(raw, jsonFmt)
	if err != nil {
		return nil, err
	}
	if set {
		var sf setFile
		if err := strictDecode(raw, jsonFmt, &sf); err != nil {
			return nil, err
		}
		return sf.Expand()
	}
	var h Host
	if err := strictDecode(raw, jsonFmt, &h); err != nil {
		return nil, err
	}
	if h.Name == "" {
		return nil, errors.New("empty or invalid host document (no name; use a hosts:/template: set for multiple hosts)")
	}
	return []Host{h}, nil
}

// isSetDoc reports whether a document is the host-SET form (has a top-level
// `hosts:` or `template:` key) vs. a single host.
func isSetDoc(raw []byte, jsonFmt bool) (bool, error) {
	var probe map[string]interface{}
	var err error
	if jsonFmt {
		err = json.Unmarshal(raw, &probe)
	} else {
		err = yaml.Unmarshal(raw, &probe)
	}
	if err != nil {
		return false, err
	}
	_, hasHosts := probe["hosts"]
	_, hasTemplate := probe["template"]
	return hasHosts || hasTemplate, nil
}

// strictDecode decodes with unknown-field rejection, which catches typos AND an
// inline `password:` key (a security guard — passwords come from env only).
func strictDecode(raw []byte, jsonFmt bool, v interface{}) error {
	if jsonFmt {
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		if err := dec.Decode(v); err != nil {
			return err
		}
		if dec.More() {
			return errors.New("trailing content after the first JSON document; one host doc per file (use a hosts: list or a directory)")
		}
		return nil
	}
	dec := yaml.NewDecoder(bytes.NewReader(raw))
	dec.KnownFields(true)
	if err := dec.Decode(v); err != nil {
		return err
	}
	// Reject a multi-document stream: a second Decode that isn't EOF means more
	// host docs follow that would be silently dropped.
	var extra interface{}
	switch err := dec.Decode(&extra); err {
	case io.EOF:
		return nil
	case nil:
		return errors.New("multiple YAML documents in one file are not supported; use a hosts: list or a directory")
	default:
		return fmt.Errorf("trailing content after first document: %w", err)
	}
}

func isConfigFile(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".hostcfg", ".yaml", ".yml", ".json":
		return true
	}
	return false
}

func isJSONFile(path string) bool { return strings.EqualFold(filepath.Ext(path), ".json") }
