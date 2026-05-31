package spec

import (
	"fmt"
	"strconv"
	"strings"
)

// RuntimeVersion is the semantic version of the **Routine Runtime** — the
// .routine language surface (the builtin / event / accessor tables in this
// package), the interpreter semantics, and the frozen host API
// (docs/lang/api.md). Bump it on every surface/semantics change per the policy
// in docs/lang/versioning.md, and record the change in docs/lang/CHANGELOG.md:
//
//   MAJOR — breaking: remove/rename/retype a builtin·event·accessor, change
//           evaluation semantics (truthiness, scope, eval order), remove syntax.
//           Scripts targeting an older major may break.
//   MINOR — additive & backward-compatible: new builtin·event·accessor, new
//           optional param, new syntax, relaxed cap. Older-minor scripts run.
//   PATCH — fixes only: bug/perf/diagnostic, no surface or behavior change for
//           correct scripts.
//
// Anchor: api.md was frozen as "v1" (commit 38ef5a0), so the runtime starts at
// 1.0.0.
const RuntimeVersion = "1.0.0"

// Version is a parsed semver. Patch is optional in a script's target.
type Version struct{ Major, Minor, Patch int }

func (v Version) String() string { return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch) }

// ParseVersion parses "X", "X.Y", or "X.Y.Z" into a Version; missing trailing
// components default to 0.
func ParseVersion(s string) (Version, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Version{}, fmt.Errorf("empty version")
	}
	parts := strings.SplitN(s, ".", 3)
	var v Version
	dst := []*int{&v.Major, &v.Minor, &v.Patch}
	for i, p := range parts {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil || n < 0 {
			return Version{}, fmt.Errorf("invalid version %q: component %q is not a non-negative integer", s, p)
		}
		*dst[i] = n
	}
	return v, nil
}

// Runtime returns the parsed current RuntimeVersion.
func Runtime() Version {
	v, _ := ParseVersion(RuntimeVersion)
	return v
}

// CheckTarget reports whether a script declaring `runtime "<target>"` can run on
// the current RuntimeVersion. A target is compatible iff it shares the runtime's
// MAJOR and its MINOR is <= the runtime's MINOR (the runtime is same-or-newer
// within the major). PATCH never affects compatibility. Returns a descriptive
// error when the target is malformed or incompatible.
func CheckTarget(target string) error {
	want, err := ParseVersion(target)
	if err != nil {
		return fmt.Errorf("runtime target %q: %w", target, err)
	}
	cur := Runtime()
	if want.Major != cur.Major {
		return fmt.Errorf("runtime target %q is incompatible with runtime %s: different MAJOR version (a major bump is breaking — see docs/lang/CHANGELOG.md)", target, RuntimeVersion)
	}
	if want.Minor > cur.Minor {
		return fmt.Errorf("runtime target %q needs a newer runtime than %s: it requires minor >= %d", target, RuntimeVersion, want.Minor)
	}
	return nil
}
