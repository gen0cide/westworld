package spec

import (
	"fmt"
	"testing"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		in            string
		maj, min, pat int
		ok            bool
	}{
		{"1", 1, 0, 0, true},
		{"1.2", 1, 2, 0, true},
		{"1.2.3", 1, 2, 3, true},
		{" 2.0 ", 2, 0, 0, true},
		{"", 0, 0, 0, false},
		{"x", 0, 0, 0, false},
		{"1.x", 0, 0, 0, false},
		{"-1", 0, 0, 0, false},
		{"1..2", 0, 0, 0, false},
	}
	for _, c := range cases {
		v, err := ParseVersion(c.in)
		if (err == nil) != c.ok {
			t.Errorf("ParseVersion(%q): ok=%v, err=%v", c.in, c.ok, err)
			continue
		}
		if c.ok && (v.Major != c.maj || v.Minor != c.min || v.Patch != c.pat) {
			t.Errorf("ParseVersion(%q) = %v, want %d.%d.%d", c.in, v, c.maj, c.min, c.pat)
		}
	}
}

func TestRuntimeVersionWellFormed(t *testing.T) {
	if _, err := ParseVersion(RuntimeVersion); err != nil {
		t.Fatalf("RuntimeVersion %q is not a valid semver: %v", RuntimeVersion, err)
	}
}

func TestCheckTarget(t *testing.T) {
	cur := Runtime()
	sameMinor := fmt.Sprintf("%d.%d", cur.Major, cur.Minor)
	majorOnly := fmt.Sprintf("%d", cur.Major)
	newerMinor := fmt.Sprintf("%d.%d", cur.Major, cur.Minor+1)
	newerMajor := fmt.Sprintf("%d.0", cur.Major+1)
	olderMajor := fmt.Sprintf("%d.0", cur.Major-1) // may be "0.0" or "-1.0" — both incompatible

	compatible := []string{RuntimeVersion, sameMinor, majorOnly}
	for _, target := range compatible {
		if err := CheckTarget(target); err != nil {
			t.Errorf("CheckTarget(%q) should be compatible with %s, got: %v", target, RuntimeVersion, err)
		}
	}
	incompatible := []string{newerMinor, newerMajor, olderMajor, "garbage"}
	for _, target := range incompatible {
		if err := CheckTarget(target); err == nil {
			t.Errorf("CheckTarget(%q) should be INCOMPATIBLE with %s, got nil", target, RuntimeVersion)
		}
	}
}
