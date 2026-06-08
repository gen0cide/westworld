package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	mesapb "github.com/gen0cide/westworld/mesa/proto"
)

// editField maps an operator-facing field name (the same names as the generated
// DB columns) to its path inside persona_json and whether it's numeric.
type editField struct {
	path  []string
	float bool
}

// editable is the set of persona fields `persona set` can patch — the dials +
// identity facets, by their column names. Trait dials target the .band (the
// authored word; the sampler derives the numeric mu). Validation of the value
// happens server-side via persona.Validate on the re-upload.
var editable = map[string]editField{
	"hexaco_h":          {[]string{"cornerstone", "hexaco", "H", "band"}, false},
	"hexaco_e":          {[]string{"cornerstone", "hexaco", "E", "band"}, false},
	"hexaco_x":          {[]string{"cornerstone", "hexaco", "X", "band"}, false},
	"hexaco_a":          {[]string{"cornerstone", "hexaco", "A", "band"}, false},
	"hexaco_c":          {[]string{"cornerstone", "hexaco", "C", "band"}, false},
	"hexaco_o":          {[]string{"cornerstone", "hexaco", "O", "band"}, false},
	"north_star_value":  {[]string{"cornerstone", "values", "north_star_value"}, false},
	"secondary_value":   {[]string{"cornerstone", "values", "secondary_value"}, false},
	"patience":          {[]string{"cornerstone", "prefs", "patience", "band"}, false},
	"loss_aversion":     {[]string{"cornerstone", "prefs", "loss_aversion", "band"}, false},
	"aggression":        {[]string{"cornerstone", "prefs", "aggression", "band"}, false},
	"decisiveness":      {[]string{"cornerstone", "prefs", "decisiveness", "band"}, false},
	"tenacity":          {[]string{"cornerstone", "prefs", "tenacity", "band"}, false},
	"bulk_apperception": {[]string{"cornerstone", "prefs", "bulk_apperception", "band"}, false},
	"self_preservation": {[]string{"cornerstone", "prefs", "self_preservation", "band"}, false},
	"coop_type":         {[]string{"cornerstone", "prefs", "coop_type"}, false},
	"risk_economic":     {[]string{"cornerstone", "prefs", "risk", "economic"}, false},
	"risk_bodily":       {[]string{"cornerstone", "prefs", "risk", "bodily"}, false},
	"risk_social":       {[]string{"cornerstone", "prefs", "risk", "social"}, false},
	"attention_level":   {[]string{"cornerstone", "prefs", "attention", "level"}, false},
	"cur_social":        {[]string{"cornerstone", "prefs", "curiosity", "social"}, true},
	"cur_spatial":       {[]string{"cornerstone", "prefs", "curiosity", "spatial"}, true},
	"cur_skill":         {[]string{"cornerstone", "prefs", "curiosity", "skill"}, true},
	"cur_economic":      {[]string{"cornerstone", "prefs", "curiosity", "economic"}, true},
	"cur_risk":          {[]string{"cornerstone", "prefs", "curiosity", "risk"}, true},
	"north_star_theme":  {[]string{"cornerstone", "identity", "north_star", "theme"}, false},
	"north_star":        {[]string{"cornerstone", "identity", "north_star", "statement"}, false},
	"voice_formality":   {[]string{"cornerstone", "identity", "voice", "formality"}, false},
	"voice_typo_feel":   {[]string{"cornerstone", "identity", "voice", "typo_feel"}, false},
	"voice_register":    {[]string{"cornerstone", "identity", "voice", "register"}, false},
	"archetype":         {[]string{"cornerstone", "generation_meta", "archetype"}, false},
	"cohort_id":         {[]string{"cornerstone", "generation_meta", "cohort_id"}, false},
}

// setCmd implements `mesa-ctl persona set <host_id> <field> <value>`: fetch the
// persona, patch one field in its JSON, and re-upload it through PutPersonas so
// the server validates (rejecting a bad band), re-renders the prose, updates the
// live registry, and persists — after which the generated column reflects it.
func setCmd(ctx context.Context, c mesapb.AdminClient, args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: mesa-ctl persona set <host_id> <field> <value>\n  fields: %s", editableList())
	}
	host, field, value := args[0], args[1], args[2]
	key := strings.ToLower(strings.ReplaceAll(field, ".", "_"))
	ef, ok := editable[key]
	if !ok {
		return fmt.Errorf("unknown field %q; editable fields:\n  %s", field, editableList())
	}

	rec, err := c.GetPersona(ctx, &mesapb.HostRef{HostId: host})
	if err != nil {
		return err
	}
	var m map[string]any
	if err := json.Unmarshal(rec.PersonaJson, &m); err != nil {
		return fmt.Errorf("decode persona: %w", err)
	}

	var v any = value
	if ef.float {
		f, perr := strconv.ParseFloat(value, 64)
		if perr != nil {
			return fmt.Errorf("%s expects a number, got %q", key, value)
		}
		v = f
	}
	setPath(m, ef.path, v)

	patched, err := json.Marshal(m)
	if err != nil {
		return err
	}
	res, err := stream(ctx, c, []*mesapb.PersonaUpsert{{HostId: host, PersonaJson: patched}})
	if err != nil {
		return err
	}
	for _, it := range res.Items {
		if !it.Ok {
			return fmt.Errorf("%s rejected: %s", host, it.Error)
		}
	}
	fmt.Printf("✓ %s  %s = %v\n", host, key, v)
	return nil
}

// setPath sets v at the nested map path, creating/replacing non-object
// intermediates as it descends (a null/absent dial becomes a fresh object).
func setPath(m map[string]any, path []string, v any) {
	cur := m
	for i, k := range path {
		if i == len(path)-1 {
			cur[k] = v
			return
		}
		next, _ := cur[k].(map[string]any)
		if next == nil {
			next = map[string]any{}
			cur[k] = next
		}
		cur = next
	}
}

// getPath reads the value at a nested map path as a display string ("" if absent).
func getPath(m map[string]any, path ...string) string {
	var cur any = m
	for _, k := range path {
		mm, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = mm[k]
	}
	switch x := cur.(type) {
	case nil:
		return ""
	case string:
		return x
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	default:
		return fmt.Sprint(cur)
	}
}

func editableList() string {
	names := make([]string, 0, len(editable))
	for k := range editable {
		names = append(names, k)
	}
	sort.Strings(names)
	return strings.Join(names, ", ")
}
