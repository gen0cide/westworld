package spec_test

import (
	"sort"
	"strings"
	"testing"

	"github.com/gen0cide/westworld/dsl/spec"
)

// These tests pin internal invariants of the spec tables. They run
// fast and exist purely to catch "added an entry, forgot to fill a
// field" mistakes at build time.

// ----- spec.Actions self-consistency -----

func TestActionNamesAreUnique(t *testing.T) {
	seen := map[string]int{}
	for _, a := range spec.Actions {
		if seen[a.Name] > 0 {
			t.Errorf("duplicate action name: %q", a.Name)
		}
		seen[a.Name]++
	}
}

func TestActionNamesAreSnakeCase(t *testing.T) {
	for _, a := range spec.Actions {
		if a.Name == "" {
			t.Errorf("action has empty name (Kind=%s)", a.Kind)
			continue
		}
		for _, r := range a.Name {
			if r != '_' && (r < 'a' || r > 'z') && (r < '0' || r > '9') {
				t.Errorf("action %q has non-snake_case character %q", a.Name, r)
				break
			}
		}
		if a.Name[0] < 'a' || a.Name[0] > 'z' {
			t.Errorf("action %q must start with a lowercase letter", a.Name)
		}
		if strings.HasSuffix(a.Name, "!") {
			t.Errorf("action %q must not include the bang suffix; bang is implicit from BangEligible()", a.Name)
		}
	}
}

func TestActionArityIsCoherent(t *testing.T) {
	for _, a := range spec.Actions {
		if a.MinArgs < 0 {
			t.Errorf("action %q: MinArgs < 0 (%d)", a.Name, a.MinArgs)
		}
		if a.MaxArgs >= 0 && a.MaxArgs < a.MinArgs {
			t.Errorf("action %q: MaxArgs (%d) < MinArgs (%d)", a.Name, a.MaxArgs, a.MinArgs)
		}
		if a.MaxArgs >= 0 && len(a.Params) > a.MaxArgs {
			t.Errorf("action %q: Params (%d) exceeds MaxArgs (%d)", a.Name, len(a.Params), a.MaxArgs)
		}
	}
}

func TestActionParamKindsAlignWithParams(t *testing.T) {
	valid := map[string]bool{
		spec.CatalogNone:       true,
		spec.CatalogPlaceOrPOI: true,
		spec.CatalogItem:       true,
		spec.CatalogNPC:        true,
	}
	for _, a := range spec.Actions {
		if a.ParamKinds == nil {
			continue // nil = all params uncatalogued; fine
		}
		// ParamKinds may be a PREFIX of Params (trailing uncatalogued params
		// can be omitted — ParamKind treats out-of-range as CatalogNone), but
		// it must never be LONGER than Params.
		if len(a.ParamKinds) > len(a.Params) {
			t.Errorf("action %q: ParamKinds len (%d) exceeds Params len (%d)", a.Name, len(a.ParamKinds), len(a.Params))
		}
		for i, k := range a.ParamKinds {
			if !valid[k] {
				t.Errorf("action %q: ParamKinds[%d] = %q is not a known catalog kind", a.Name, i, k)
			}
		}
	}
}

func TestActionDocSummaryNonEmpty(t *testing.T) {
	for _, a := range spec.Actions {
		if a.DocSummary == "" {
			t.Errorf("action %q: empty DocSummary", a.Name)
		}
	}
}

func TestActionBangEligibilityMatchesKind(t *testing.T) {
	for _, a := range spec.Actions {
		want := false
		switch a.Kind {
		case spec.PrimaryAction, spec.LLMStdlib, spec.MemoryStdlib:
			want = true
		}
		if got := a.BangEligible(); got != want {
			t.Errorf("action %q (Kind=%s): BangEligible()=%v, want %v", a.Name, a.Kind, got, want)
		}
	}
}

func TestStripBang(t *testing.T) {
	cases := map[string]struct {
		base    string
		hadBang bool
	}{
		"eat":     {"eat", false},
		"eat!":    {"eat", true},
		"walk_to": {"walk_to", false},
		"!":       {"!", false}, // single char — too short to strip
		"":        {"", false},
	}
	for input, want := range cases {
		base, hadBang := spec.StripBang(input)
		if base != want.base || hadBang != want.hadBang {
			t.Errorf("StripBang(%q) = (%q, %v), want (%q, %v)", input, base, hadBang, want.base, want.hadBang)
		}
	}
}

func TestByNameRoundtrip(t *testing.T) {
	for _, a := range spec.Actions {
		got, ok := spec.ByName(a.Name)
		if !ok {
			t.Errorf("ByName(%q) returned ok=false", a.Name)
			continue
		}
		if got.Name != a.Name {
			t.Errorf("ByName(%q) returned spec with Name=%q", a.Name, got.Name)
		}
	}
	if _, ok := spec.ByName("nonexistent_thing_xyz"); ok {
		t.Error("ByName(nonexistent) should return ok=false")
	}
}

// ----- spec.Events self-consistency -----

func TestEventNamesAreUnique(t *testing.T) {
	seen := map[string]int{}
	for _, e := range spec.Events {
		if seen[e.Name] > 0 {
			t.Errorf("duplicate event name: %q", e.Name)
		}
		seen[e.Name]++
	}
}

func TestEventNamesAreSnakeCase(t *testing.T) {
	for _, e := range spec.Events {
		if e.Name == "" {
			t.Errorf("event has empty name")
			continue
		}
		for _, r := range e.Name {
			if r != '_' && (r < 'a' || r > 'z') {
				t.Errorf("event %q has non-snake_case character %q", e.Name, r)
				break
			}
		}
	}
}

func TestEventDocSummaryNonEmpty(t *testing.T) {
	for _, e := range spec.Events {
		if e.DocSummary == "" {
			t.Errorf("event %q: empty DocSummary", e.Name)
		}
	}
}

func TestEventByNameRoundtrip(t *testing.T) {
	for _, e := range spec.Events {
		got, ok := spec.EventByName(e.Name)
		if !ok || got.Name != e.Name {
			t.Errorf("EventByName(%q) failed: got=%v ok=%v", e.Name, got, ok)
		}
	}
}

// ----- spec.Accessors self-consistency -----

func TestAccessorPathsAreUnique(t *testing.T) {
	seen := map[string]int{}
	for _, a := range spec.Accessors {
		key := strings.Join(a.Path, ".")
		if seen[key] > 0 {
			t.Errorf("duplicate accessor path: %s", key)
		}
		seen[key]++
	}
}

func TestAccessorsRootReservedName(t *testing.T) {
	allowedRoots := map[string]bool{
		"self":      true,
		"host":      true,
		"world":     true,
		"inventory": true,
		"combat":    true,
		"bank":      true,
		"trade":     true,
		"duel":      true,
		"magic":     true,
		"prayer":    true,
		"shop":      true,
	}
	for _, a := range spec.Accessors {
		if len(a.Path) == 0 {
			t.Errorf("accessor with empty path: %+v", a)
			continue
		}
		root := a.Path[0]
		if !allowedRoots[root] {
			t.Errorf("accessor path %s starts with non-reserved root %q", strings.Join(a.Path, "."), root)
		}
	}
}

func TestAccessorDocsNonEmpty(t *testing.T) {
	for _, a := range spec.Accessors {
		if a.DocSummary == "" {
			t.Errorf("accessor %s: empty DocSummary", strings.Join(a.Path, "."))
		}
		if a.Kind == "" {
			t.Errorf("accessor %s: empty Kind", strings.Join(a.Path, "."))
		}
	}
}

// ----- summary reporting -----

func TestSpecSummary(t *testing.T) {
	// Not a real test — just emits a summary in verbose mode so we
	// can see the spec's growth over time.
	kindCounts := map[spec.ActionKind]int{}
	notYet := 0
	for _, a := range spec.Actions {
		kindCounts[a.Kind]++
		if a.NotYetImplemented {
			notYet++
		}
	}
	t.Logf("Actions: %d total (%d not yet implemented)", len(spec.Actions), notYet)
	kinds := []spec.ActionKind{}
	for k := range kindCounts {
		kinds = append(kinds, k)
	}
	sort.Slice(kinds, func(i, j int) bool { return int(kinds[i]) < int(kinds[j]) })
	for _, k := range kinds {
		t.Logf("  %s: %d", k, kindCounts[k])
	}
	t.Logf("Events: %d", len(spec.Events))
	t.Logf("Accessors: %d", len(spec.Accessors))
}
