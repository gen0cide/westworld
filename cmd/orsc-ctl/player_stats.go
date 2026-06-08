package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// Group "player": stat / skill / quest WRITE commands. Each maps to a single
// admin endpoint; body fields are taken from the spec examples
// (.data.paths[...].requestBody.content["application/json"].example) and exposed
// as subcommand-local flags. Path params come from positional args.

func init() {
	// restore <username> -> POST /players/{username}/restore
	// Body schema (example {"allSkills":true}): restores current levels to base
	// levels, defaulting to all skills.
	register("player", cmd{
		name:  "restore",
		usage: "restore <username> [-all-skills]",
		help:  "restore current stat levels to base levels (defaults to all skills)",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player restore", flag.ContinueOnError)
			allSkills := fs.Bool("all-skills", true, "restore all skills")
			if err := fs.Parse(args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player restore <username> [-all-skills]")
			}
			username := rest[0]

			path := "/players/" + url.PathEscape(username) + "/restore"
			body := map[string]any{"allSkills": *allSkills}
			res, err := c.post(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// skills-set <username> -> PATCH /players/{username}/skills
	// Body schema (example {"skills":[{"skill":"hits","currentLevel":10,
	// "baseLevel":10}]}): bulk patch. Each -skill entry is
	// "name[:current[:base[:xp]]]" and may be repeated; -raw passes a JSON body
	// straight through for full control.
	register("player", cmd{
		name:  "skills-set",
		usage: "skills-set <username> -skill name[:current[:base[:xp]]] [-skill ...] | -raw JSON",
		help:  "bulk-patch skills (levels/xp); repeat -skill per skill",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player skills-set", flag.ContinueOnError)
			var skills skillList
			fs.Var(&skills, "skill", "skill patch \"name[:current[:base[:xp]]]\" (repeatable)")
			raw := fs.String("raw", "", "raw JSON request body (overrides -skill)")
			if err := fs.Parse(args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player skills-set <username> -skill name[:current[:base[:xp]]] [-skill ...] | -raw JSON")
			}
			username := rest[0]
			path := "/players/" + url.PathEscape(username) + "/skills"

			if *raw != "" {
				var body json.RawMessage
				if err := json.Unmarshal([]byte(*raw), &body); err != nil {
					return fmt.Errorf("-raw is not valid JSON: %w", err)
				}
				res, err := c.patch(path, body)
				if err != nil {
					return err
				}
				return printResult(res)
			}
			if len(skills) == 0 {
				return fmt.Errorf("usage: orsc-ctl player skills-set <username> -skill name[:current[:base[:xp]]] [-skill ...] | -raw JSON")
			}
			body := map[string]any{"skills": []map[string]any(skills)}
			res, err := c.patch(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// skill-set <username> <skill> -> PATCH /players/{username}/skills/{skill}
	// Body schema (example {"currentLevel":10,"baseLevel":10,"experience":1154}):
	// patches a single named skill. At least one field is required.
	register("player", cmd{
		name:  "skill-set",
		usage: "skill-set <username> <skill> [-current-level N] [-base-level N] [-experience N]",
		help:  "patch one skill's current/base level and/or experience",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player skill-set", flag.ContinueOnError)
			current := fs.Int("current-level", 0, "current (boosted) level")
			base := fs.Int("base-level", 0, "base level")
			experience := fs.Int("experience", 0, "experience points")
			if err := fs.Parse(args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 2 {
				return fmt.Errorf("usage: orsc-ctl player skill-set <username> <skill> [-current-level N] [-base-level N] [-experience N]")
			}
			username := rest[0]
			skill := rest[1]

			body := map[string]any{}
			setFlagInt(fs, "current-level", "currentLevel", *current, body)
			setFlagInt(fs, "base-level", "baseLevel", *base, body)
			setFlagInt(fs, "experience", "experience", *experience, body)
			if len(body) == 0 {
				return fmt.Errorf("player skill-set: set at least one of -current-level, -base-level, -experience")
			}

			path := "/players/" + url.PathEscape(username) + "/skills/" + url.PathEscape(skill)
			res, err := c.patch(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// fatigue <username> -> PATCH /players/{username}/fatigue
	// Body schema (example {"fatigue":0}): sets the player's fatigue value.
	// fatigue is required.
	register("player", cmd{
		name:  "fatigue",
		usage: "fatigue <username> -fatigue N",
		help:  "set the player's fatigue value",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player fatigue", flag.ContinueOnError)
			fatigue := fs.Int("fatigue", 0, "fatigue value (required)")
			if err := fs.Parse(args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 1 {
				return fmt.Errorf("usage: orsc-ctl player fatigue <username> -fatigue N")
			}
			username := rest[0]
			if !flagSet(fs, "fatigue") {
				return fmt.Errorf("player fatigue: -fatigue is required")
			}

			path := "/players/" + url.PathEscape(username) + "/fatigue"
			body := map[string]any{"fatigue": *fatigue}
			res, err := c.patch(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})

	// quest-set <username> <questId> -> PATCH /players/{username}/quests/{questId}
	// Body schema (example {"stage":1}): set a quest stage, or complete the quest
	// via existing completion semantics. -stage is required unless -complete.
	register("player", cmd{
		name:  "quest-set",
		usage: "quest-set <username> <questId> -stage N | -complete",
		help:  "set a quest stage (or -complete the quest)",
		run: func(c *Client, args []string) error {
			fs := flag.NewFlagSet("player quest-set", flag.ContinueOnError)
			stage := fs.Int("stage", 0, "quest stage to set")
			complete := fs.Bool("complete", false, "complete the quest")
			if err := fs.Parse(args); err != nil {
				return err
			}
			rest := fs.Args()
			if len(rest) < 2 {
				return fmt.Errorf("usage: orsc-ctl player quest-set <username> <questId> -stage N | -complete")
			}
			username := rest[0]
			questID := rest[1]

			body := map[string]any{}
			if flagSet(fs, "stage") {
				body["stage"] = *stage
			}
			if *complete {
				body["complete"] = true
			}
			if len(body) == 0 {
				return fmt.Errorf("player quest-set: set -stage N or -complete")
			}

			path := "/players/" + url.PathEscape(username) + "/quests/" + url.PathEscape(questID)
			res, err := c.patch(path, body)
			if err != nil {
				return err
			}
			return printResult(res)
		},
	})
}

// skillList collects repeatable -skill entries for the bulk skills patch. Each
// value is "name[:current[:base[:xp]]]" where the trailing numeric fields are
// optional; only the fields actually given are sent.
type skillList []map[string]any

func (s *skillList) String() string {
	parts := make([]string, 0, len(*s))
	for _, m := range *s {
		if name, ok := m["skill"].(string); ok {
			parts = append(parts, name)
		}
	}
	return strings.Join(parts, ",")
}

func (s *skillList) Set(v string) error {
	fields := strings.Split(v, ":")
	name := strings.TrimSpace(fields[0])
	if name == "" {
		return fmt.Errorf("skill patch needs a skill name: %q", v)
	}
	m := map[string]any{"skill": name}
	keys := []string{"currentLevel", "baseLevel", "experience"}
	for i, key := range keys {
		idx := i + 1
		if idx >= len(fields) || strings.TrimSpace(fields[idx]) == "" {
			continue
		}
		n, err := strconv.Atoi(strings.TrimSpace(fields[idx]))
		if err != nil {
			return fmt.Errorf("skill patch %q: %s must be an integer: %w", v, key, err)
		}
		m[key] = n
	}
	*s = append(*s, m)
	return nil
}

// flagSet reports whether the named flag was explicitly provided on the command
// line (as opposed to left at its default).
func flagSet(fs *flag.FlagSet, name string) bool {
	found := false
	fs.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

// setFlagInt writes val into body[jsonKey] only if the flag named flagName was
// explicitly set, so unset optional fields are omitted from the request body.
func setFlagInt(fs *flag.FlagSet, flagName, jsonKey string, val int, body map[string]any) {
	if flagSet(fs, flagName) {
		body[jsonKey] = val
	}
}
