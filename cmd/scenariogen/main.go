// Command scenariogen maintains the scenario MANIFEST (scenarios.yaml) and
// validates it against the .routine corpus under examples/scenarios/.
//
// SOURCE OF TRUTH: the `.routine` files are authoritative. They are authored and
// edited by hand (the live-test "run-to-ground" campaign edits them directly), and
// the runner scripts execute them directly (`for f in examples/scenarios/*/*.routine`).
// scenarios.yaml is a *code-free manifest* that merely REFERENCES each file plus its
// metadata (category / hosts / admin / timeout). It contains NO routine bodies — that
// design (embedding code in YAML) is what historically drifted, because edits landed
// in the .routine files but not the YAML.
//
// Modes:
//
//	go run ./cmd/scenariogen            # VALIDATE: manifest ⇄ corpus are consistent
//	go run ./cmd/scenariogen -reindex   # RE-DERIVE the manifest from the corpus headers
//
// Validation (the default, CI-able — exits non-zero on any problem) checks both
// directions: every manifest entry points at a real .routine file whose name matches
// and that parses; and every .routine file in the corpus has a manifest entry (so a
// hand-added file without a row — the classic orphan — is caught). Full DSL parse
// errors are also surfaced by `cmd/parsecheck`; this tool reuses the same parser.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gen0cide/westworld/runtime"
)

// Scenario is one manifest entry: metadata + a reference to the .routine file.
// No routine body — the file is the source of truth.
type Scenario struct {
	ID       string `yaml:"id"`
	Category string `yaml:"category"`
	File     string `yaml:"file"`
	Hosts    string `yaml:"hosts,omitempty"`
	Admin    bool   `yaml:"admin,omitempty"`
	Timeout  int    `yaml:"timeout,omitempty"`
}

type catalog struct {
	Scenarios []Scenario `yaml:"scenarios"`
}

// snake converts kebab-case → snake_case for the routine name / filename. The
// loader (runtime.ParseRoutineFile) enforces filename ↔ routine-name match.
func snake(id string) string { return strings.ReplaceAll(id, "-", "_") }

const manifestHeader = `# Scenario manifest — one entry per .routine the runners execute.
#
# The .routine files under examples/scenarios/<category>/ are the SOURCE OF TRUTH;
# this file is a code-free index that references them (no routine bodies live here).
# Edit a scenario by editing its .routine file. Add a scenario by creating the file
# AND adding a row here (run` + " `go run ./cmd/scenariogen` " + `to verify, or
#` + " `-reindex` " + `to re-derive this file from the corpus headers).
#
# Field semantics:
#   id:       kebab-case unique id; the file is <id>.routine (kebab→snake).
#   category: skills | combat | quests | social | edges | movement (== the subdir).
#   file:     repo-relative path to the .routine file.
#   hosts:    which host(s) the runner needs ("any-drone", "Drone1-and-Drone2", …).
#   admin:    true if the scenario uses admin ::commands.
#   timeout:  per-scenario wall-clock budget in seconds.
`

func main() {
	var (
		yamlPath = flag.String("yaml", "cmd/scenariogen/scenarios.yaml", "Scenario manifest")
		root     = flag.String("root", "examples/scenarios", "Root of the .routine corpus")
		reindex  = flag.Bool("reindex", false, "Re-derive the manifest from the corpus headers and write it to -yaml")
	)
	flag.Parse()

	if *reindex {
		if err := doReindex(*root, *yamlPath); err != nil {
			fmt.Fprintf(os.Stderr, "reindex: %v\n", err)
			os.Exit(1)
		}
		return
	}
	if err := validate(*yamlPath, *root); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// corpusFiles returns every .routine path under root, sorted.
func corpusFiles(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && filepath.Ext(p) == ".routine" {
			files = append(files, p)
		}
		return nil
	})
	sort.Strings(files)
	return files, err
}

// parseHeader reads the leading `# Key: value` comment block a .routine carries
// (emitted historically by this generator) into a Scenario manifest row.
func parseHeader(path string) (Scenario, error) {
	f, err := os.Open(path)
	if err != nil {
		return Scenario{}, err
	}
	defer f.Close()

	s := Scenario{File: path, Category: filepath.Base(filepath.Dir(path))}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		switch {
		case strings.HasPrefix(line, "routine "):
			// Header block ends at the routine declaration.
			return s, sc.Err()
		case strings.HasPrefix(line, "# Scenario:"):
			s.ID = strings.TrimSpace(strings.TrimPrefix(line, "# Scenario:"))
		case strings.HasPrefix(line, "# Category:"):
			s.Category = strings.TrimSpace(strings.TrimPrefix(line, "# Category:"))
		case strings.HasPrefix(line, "# Hosts:"):
			s.Hosts = strings.TrimSpace(strings.TrimPrefix(line, "# Hosts:"))
		case strings.HasPrefix(line, "# Admin:"):
			s.Admin = true
		case strings.HasPrefix(line, "# Timeout:"):
			v := strings.TrimSpace(strings.TrimPrefix(line, "# Timeout:"))
			v = strings.TrimSuffix(v, "s")
			if n, err := strconv.Atoi(v); err == nil {
				s.Timeout = n
			}
		}
	}
	return s, sc.Err()
}

// doReindex scans the corpus and (re)writes the manifest. This is the bootstrap +
// resync tool; the normal workflow is hand-maintenance verified by `validate`.
func doReindex(root, yamlPath string) error {
	files, err := corpusFiles(root)
	if err != nil {
		return err
	}
	var cat catalog
	for _, f := range files {
		s, err := parseHeader(f)
		if err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
		if s.ID == "" {
			return fmt.Errorf("%s: missing '# Scenario:' header", f)
		}
		cat.Scenarios = append(cat.Scenarios, s)
	}
	sort.Slice(cat.Scenarios, func(i, j int) bool {
		if cat.Scenarios[i].Category != cat.Scenarios[j].Category {
			return cat.Scenarios[i].Category < cat.Scenarios[j].Category
		}
		return cat.Scenarios[i].ID < cat.Scenarios[j].ID
	})
	body, err := yaml.Marshal(cat)
	if err != nil {
		return err
	}
	out := manifestHeader + "\n" + string(body)
	if err := os.WriteFile(yamlPath, []byte(out), 0o644); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "reindexed %d scenarios from %s into %s\n", len(cat.Scenarios), root, yamlPath)
	return nil
}

// validate checks the manifest and corpus are consistent in BOTH directions and
// that every referenced file parses. Returns a non-nil error listing all problems.
func validate(yamlPath, root string) error {
	raw, err := os.ReadFile(yamlPath)
	if err != nil {
		return fmt.Errorf("read %s: %w", yamlPath, err)
	}
	var cat catalog
	if err := yaml.Unmarshal(raw, &cat); err != nil {
		return fmt.Errorf("parse %s: %w", yamlPath, err)
	}

	var problems []string
	referenced := make(map[string]bool, len(cat.Scenarios))
	seenID := make(map[string]bool, len(cat.Scenarios))

	for _, s := range cat.Scenarios {
		if seenID[s.ID] {
			problems = append(problems, fmt.Sprintf("duplicate manifest id %q", s.ID))
		}
		seenID[s.ID] = true

		want := filepath.Join(root, s.Category, snake(s.ID)+".routine")
		if s.File != want {
			problems = append(problems, fmt.Sprintf("%s: file %q does not match category/id convention (want %q)", s.ID, s.File, want))
		}
		referenced[s.File] = true

		if _, err := os.Stat(s.File); err != nil {
			problems = append(problems, fmt.Sprintf("%s: referenced file missing: %v", s.ID, err))
			continue
		}
		if _, err := runtime.ParseRoutineFile(s.File); err != nil {
			problems = append(problems, fmt.Sprintf("%s: %s does not parse: %v", s.ID, s.File, err))
		}
	}

	// Reverse check: every .routine in the corpus must have a manifest entry.
	files, err := corpusFiles(root)
	if err != nil {
		return fmt.Errorf("walk %s: %w", root, err)
	}
	for _, f := range files {
		if !referenced[f] {
			problems = append(problems, fmt.Sprintf("orphan: %s has no manifest entry (add a row or run -reindex)", f))
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return fmt.Errorf("scenario manifest validation FAILED (%d problem(s)):\n  - %s",
			len(problems), strings.Join(problems, "\n  - "))
	}
	fmt.Fprintf(os.Stderr, "OK: %d manifest entries ⇄ %d corpus files, all consistent and parsing\n", len(cat.Scenarios), len(files))
	return nil
}
