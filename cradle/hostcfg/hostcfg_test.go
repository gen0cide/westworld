package hostcfg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDecodeSingleHostYAML(t *testing.T) {
	src := `
name: stubbs
server: localhost:43594
mesa: localhost:7077
goal: "Finish tutorial island"
state: file
turn_timeout: 90s
settle: 400ms
`
	hosts, err := decodeDoc([]byte(src), false)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(hosts) != 1 {
		t.Fatalf("want 1 host, got %d", len(hosts))
	}
	h := hosts[0]
	if h.Name != "stubbs" || h.Mesa != "localhost:7077" || !h.Autonomous() {
		t.Fatalf("unexpected host: %+v", h)
	}
	if time.Duration(h.TurnTimeout) != 90*time.Second || time.Duration(h.Settle) != 400*time.Millisecond {
		t.Fatalf("durations not parsed: tt=%v settle=%v", time.Duration(h.TurnTimeout), time.Duration(h.Settle))
	}
	if err := h.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
	// The runtime.HostConfig mapping moved to package cradle (toHostConfig);
	// its assertions live in cradle/registry_test.go now.
}

func TestDecodeSingleHostJSON(t *testing.T) {
	src := `{"name":"bernard","mesa":"localhost:7077","goal":"explore","state":"memory"}`
	hosts, err := decodeDoc([]byte(src), true)
	if err != nil {
		t.Fatalf("decode json: %v", err)
	}
	h := hosts[0]
	if h.Name != "bernard" || h.State != StateMemory {
		t.Fatalf("unexpected: %+v", h)
	}
}

func TestDecodeSetWithDefaults(t *testing.T) {
	src := `
defaults:
  server: localhost:43594
  mesa: localhost:7077
  state: file
hosts:
  - name: alpha
    goal: "mine iron"
  - name: bravo
    goal: "fish"
    mesa: localhost:8088   # override the default mesa
`
	hosts, err := decodeDoc([]byte(src), false)
	if err != nil {
		t.Fatalf("decode set: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("want 2 hosts, got %d", len(hosts))
	}
	if hosts[0].Mesa != "localhost:7077" || hosts[0].Server != "localhost:43594" {
		t.Fatalf("defaults not merged into alpha: %+v", hosts[0])
	}
	if hosts[1].Mesa != "localhost:8088" {
		t.Fatalf("bravo did not override mesa: %+v", hosts[1])
	}
	if err := ValidateSet(hosts); err != nil {
		t.Fatalf("validate set: %v", err)
	}
}

func TestDecodeTemplate(t *testing.T) {
	src := `
defaults:
  server: localhost:43594
  mesa: localhost:7077
  state: memory
template:
  name_prefix: drone
  count: 200
  spec:
    director:
      routine: scenarios/walk.rt
      loop: true
`
	hosts, err := decodeDoc([]byte(src), false)
	if err != nil {
		t.Fatalf("decode template: %v", err)
	}
	if len(hosts) != 200 {
		t.Fatalf("want 200 drones, got %d", len(hosts))
	}
	if hosts[0].Name != "drone1" || hosts[199].Name != "drone200" {
		t.Fatalf("naming wrong: first=%q last=%q", hosts[0].Name, hosts[199].Name)
	}
	d := hosts[100]
	if d.Mesa != "localhost:7077" || d.State != StateMemory || d.Director.Routine != "scenarios/walk.rt" || !d.Director.Loop {
		t.Fatalf("template spec/defaults not applied: %+v", d)
	}
	if err := ValidateSet(hosts); err != nil {
		t.Fatalf("validate 200 drones: %v", err)
	}
}

func TestInlinePasswordRejected(t *testing.T) {
	src := `
name: leaky
mesa: localhost:7077
goal: "x"
password: hunter2
`
	_, err := decodeDoc([]byte(src), false)
	if err == nil {
		t.Fatal("expected an error for an inline password field")
	}
	if !strings.Contains(err.Error(), "password") {
		t.Fatalf("error should name the offending field: %v", err)
	}
}

func TestInlinePasswordRejectedEverywhere(t *testing.T) {
	cases := []struct {
		name string
		src  string
		json bool
	}{
		{"single yaml", "name: a\nmesa: m\ngoal: g\npassword: hunter2\n", false},
		{"single json", `{"name":"a","mesa":"m","goal":"g","password":"hunter2"}`, true},
		{"under defaults", "defaults:\n  password: hunter2\nhosts:\n  - name: a\n    mesa: m\n    goal: g\n", false},
		{"under hosts item", "hosts:\n  - name: a\n    mesa: m\n    goal: g\n    password: hunter2\n", false},
		{"under template spec", "template:\n  name_prefix: d\n  count: 1\n  spec:\n    mesa: m\n    goal: g\n    password: hunter2\n", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := decodeDoc([]byte(c.src), c.json); err == nil {
				t.Fatalf("inline password must be rejected (%s)", c.name)
			}
		})
	}
}

func TestRejectBadNames(t *testing.T) {
	for _, bad := range []string{"drone/1", "../evil", "a b", "name!", "", "a/b/c"} {
		if err := (Host{Name: bad, Mesa: "m"}).Validate(); err == nil {
			t.Fatalf("name %q must be rejected (path/URL hazard)", bad)
		}
	}
	for _, ok := range []string{"drone1", "delores", "host_7", "a.b-c"} {
		if err := (Host{Name: ok, Mesa: "m"}).Validate(); err != nil {
			t.Fatalf("name %q should be valid: %v", ok, err)
		}
	}
}

func TestRejectBothRoutineModes(t *testing.T) {
	h := Host{Name: "a", Director: Director{Routine: "x.rt", Routines: []string{"y.rt"}}}
	if err := h.Validate(); err == nil || !strings.Contains(err.Error(), "not both") {
		t.Fatalf("expected mutually-exclusive director error, got %v", err)
	}
}

func TestRejectMultiDocStream(t *testing.T) {
	src := "name: a\nmesa: m\n---\nname: b\nmesa: m\n"
	if _, err := decodeDoc([]byte(src), false); err == nil || !strings.Contains(err.Error(), "multiple") {
		t.Fatalf("multi-doc stream must be rejected (silent drop), got %v", err)
	}
}

func TestRejectEmptyDoc(t *testing.T) {
	if _, err := decodeDoc([]byte("{}"), false); err == nil {
		t.Fatal("empty {} document must be rejected, not yield a phantom nameless host")
	}
}

func TestValidateErrors(t *testing.T) {
	cases := []struct {
		name string
		h    Host
		want string
	}{
		{"no name", Host{Goal: "x", Mesa: "m"}, "name is required"},
		{"goal no mesa", Host{Name: "a", Goal: "x"}, "requires a mesa address"},
		{"nothing to do", Host{Name: "a"}, "needs a goal"},
		{"bad state", Host{Name: "a", Goal: "x", Mesa: "m", State: "ram"}, "state must be"},
		{"bad supervision", Host{Name: "a", Goal: "x", Mesa: "m", Supervision: "maybe"}, "supervision must be"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := c.h.Validate()
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("want error containing %q, got %v", c.want, err)
			}
		})
	}
}

func TestValidateSetDuplicateName(t *testing.T) {
	hosts := []Host{
		{Name: "dup", Goal: "x", Mesa: "m"},
		{Name: "dup", Goal: "y", Mesa: "m"},
	}
	err := ValidateSet(hosts)
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("want duplicate-name error, got %v", err)
	}
}

func TestMesaOnlyIsAutonomous(t *testing.T) {
	// A mesa-connected host with no goal and no routine is valid: it runs the
	// genesis-driven autonomous loop (goal compiled at login). This is how Delores
	// runs.
	h := Host{Name: "delores", Mesa: "localhost:7077"}
	if err := h.Validate(); err != nil {
		t.Fatalf("mesa-only host should be valid: %v", err)
	}
	if !h.Autonomous() {
		t.Fatal("mesa-only host should be autonomous")
	}
}

func TestScriptedHostNeedsNoMesa(t *testing.T) {
	h := Host{Name: "scripted", Director: Director{Routine: "wander.rt", Loop: true}}
	if err := h.Validate(); err != nil {
		t.Fatalf("offline scripted host should be valid: %v", err)
	}
	if h.Autonomous() {
		t.Fatal("scripted host is not autonomous")
	}
}

func TestResolvePassword(t *testing.T) {
	const env = "WW_TEST_PW_HOSTCFG"
	h := Host{Name: "a", PasswordEnv: env}
	t.Setenv(env, "")
	if _, err := h.ResolvePassword(); err == nil {
		t.Fatal("empty env should error")
	}
	t.Setenv(env, "topsecret")
	pw, err := h.ResolvePassword()
	if err != nil || pw != "topsecret" {
		t.Fatalf("resolve: pw=%q err=%v", pw, err)
	}
	// Default env name when unset.
	if (Host{Name: "b"}).PasswordEnvName() != DefaultPasswordEnv {
		t.Fatal("default password env wrong")
	}
}

func TestSupervisionDefaultHold(t *testing.T) {
	if (Host{}).SupervisionPolicy() != SuperviseHold {
		t.Fatal("supervision should default to hold")
	}
	if (Host{Supervision: SuperviseRestart}).SupervisionPolicy() != SuperviseRestart {
		t.Fatal("explicit supervision not honored")
	}
}

func TestExamplesParse(t *testing.T) {
	entries, err := os.ReadDir("examples")
	if err != nil {
		t.Skipf("no examples dir: %v", err)
	}
	for _, e := range entries {
		if e.IsDir() || !isConfigFile(e.Name()) {
			continue
		}
		hosts, err := Load(filepath.Join("examples", e.Name()))
		if err != nil {
			t.Fatalf("%s: load: %v", e.Name(), err)
		}
		if err := ValidateSet(hosts); err != nil {
			t.Fatalf("%s: validate: %v", e.Name(), err)
		}
	}
}

func TestLoadDir(t *testing.T) {
	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("a.hostcfg", "name: a\nmesa: m\ngoal: g\n")
	write("b.hostcfg", "name: b\nmesa: m\ngoal: g\n")
	write("ignore.txt", "name: nope\n") // non-config extension skipped
	hosts, err := Load(dir)
	if err != nil {
		t.Fatalf("load dir: %v", err)
	}
	if len(hosts) != 2 {
		t.Fatalf("want 2 hosts from dir, got %d", len(hosts))
	}
	if err := ValidateSet(hosts); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
