package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"
)

// Group "" (top-level): `orsc-ctl reset` — the fleet reset sweep (TODO O-6).
// Resets each named account to fresh-character state: every skill back to its
// fresh-account level (all 1s except Hits 10 — the createPlayer vector from
// the OpenRSC server source), inventory wiped, bank wiped, fatigue 0, then
// exactly -coins coins (catalog id 10) granted.
//
// The admin API can only address ONLINE players (it resolves usernames via
// the live world), so for each offline account the sweep opens a throwaway
// game session first — v235 login with the shared password from
// $WESTWORLD_PASSWORD (configurable via -password-env; the value is never
// printed) — waits for the server to register the player, runs the admin-API
// sequence, verifies by reading skills/inventory/bank/fatigue back through
// the API, then logs out. Accounts that are already online (e.g. under a live
// cradle host) are mutated in place without a second login; note their
// in-memory host state will be stale afterwards.
//
// The sweep is idempotent — every step sets absolute state (wipe-then-grant,
// level patches), so running it twice converges to the same result.
//
//	orsc-ctl reset -hosts drone51..drone150            # the full sweep
//	orsc-ctl reset -hosts drone51..drone60,drone99     # ranges + names mix
//	orsc-ctl reset -hosts drone51..drone150 -dry-run   # print plan, no I/O
//	orsc-ctl reset -probe drone150                     # ONE sacrificial host
//
// Operator checklist before a live run:
//  1. $OPENRSC_ADMIN_API_TOKEN (or -token) — the admin API auth token.
//  2. $WESTWORLD_PASSWORD — the shared drone account password.
//  3. Stop cradle hosts for the target range (recommended), or accept that
//     live hosts get reset under their own feet.
//  4. Validate end-to-end with `-probe drone150` (logs into ONE sacrificial
//     account) before sweeping the whole range. Probe is never the default.
func init() {
	register("", cmd{
		name:  "reset",
		usage: "reset -hosts A..B[,C] [-coins N] [-server H:P] [-password-env VAR] [-ramp DUR] [-dry-run | -probe NAME]",
		help:  "reset accounts to fresh-character state (stats 1s/Hits 10, inv+bank wiped, fatigue 0, N coins)",
		run:   runReset,
	})
}

// resetOpts carries the parsed flags for one sweep.
type resetOpts struct {
	hosts       []string
	coins       int
	server      string
	passwordEnv string
	ramp        time.Duration
	loginWait   time.Duration
	dryRun      bool
	probe       string
}

func runReset(c *Client, args []string) error {
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	hostsSpec := fs.String("hosts", "", "host spec: comma list of names and prefixN..prefixM ranges (e.g. drone51..drone150)")
	coins := fs.Int("coins", 100, "coins to grant after the wipe (catalog id 10)")
	server := fs.String("server", envOr("WESTWORLD_SERVER", "localhost:43594"), "game server host:port for session logins (default $WESTWORLD_SERVER)")
	passwordEnv := fs.String("password-env", "WESTWORLD_PASSWORD", "env var holding the shared account password (value never printed)")
	ramp := fs.Duration("ramp", 750*time.Millisecond, "delay between hosts")
	loginWait := fs.Duration("login-wait", 10*time.Second, "how long to wait for a fresh login to appear in the admin API")
	dryRun := fs.Bool("dry-run", false, "print the would-be sequence for the FIRST host and exit without connecting to anything")
	probe := fs.String("probe", "", "run the FULL sequence against exactly ONE sacrificial account (e.g. drone150); mutually exclusive with -hosts")
	if err := parseFlags(fs, args); err != nil {
		return err
	}
	if rest := fs.Args(); len(rest) > 0 && *hostsSpec == "" && *probe == "" {
		*hostsSpec = rest[0] // allow `orsc-ctl reset drone51..drone150`
	}

	o := resetOpts{
		coins: *coins, server: *server, passwordEnv: *passwordEnv,
		ramp: *ramp, loginWait: *loginWait, dryRun: *dryRun, probe: *probe,
	}
	if o.coins < 0 {
		return fmt.Errorf("reset: -coins must be >= 0")
	}

	switch {
	case o.probe != "" && *hostsSpec != "":
		return fmt.Errorf("reset: -probe and -hosts are mutually exclusive (probe targets exactly one sacrificial account)")
	case o.probe != "":
		o.hosts = []string{o.probe}
	case *hostsSpec != "":
		hosts, err := parseHosts(*hostsSpec)
		if err != nil {
			return fmt.Errorf("reset: %w", err)
		}
		o.hosts = hosts
	default:
		return fmt.Errorf("usage: orsc-ctl reset -hosts drone51..drone150 [-coins 100] (or -dry-run, or -probe NAME)")
	}

	if o.dryRun {
		printDryRun(o)
		return nil
	}

	if os.Getenv(o.passwordEnv) == "" {
		fmt.Fprintf(os.Stderr, "reset: warning: $%s is empty — hosts that need a session login will FAIL\n", o.passwordEnv)
	}

	results := make([]hostResult, 0, len(o.hosts))
	failed := 0
	for i, host := range o.hosts {
		res := resetOneHost(c, host, o)
		results = append(results, res)
		if res.Status != "PASS" {
			failed++
		}
		printHostResult(res)
		if i < len(o.hosts)-1 {
			time.Sleep(o.ramp)
		}
	}

	if *jsonOut {
		if err := printJSON(results); err != nil {
			return err
		}
	} else {
		fmt.Printf("\nreset: %d/%d PASS, %d FAIL\n", len(results)-failed, len(results), failed)
	}
	if failed > 0 {
		return fmt.Errorf("reset: %d of %d hosts failed", failed, len(results))
	}
	return nil
}

// hostResult is the per-host outcome of the sweep.
type hostResult struct {
	Host          string  `json:"host"`
	Status        string  `json:"status"` // PASS | FAIL
	AlreadyOnline bool    `json:"alreadyOnline"`
	SessionOpened bool    `json:"sessionOpened"`
	SkillsReset   int     `json:"skillsReset"`
	Coins         int     `json:"coins"`
	Detail        string  `json:"detail,omitempty"`
	Seconds       float64 `json:"seconds"`
}

func failResult(r hostResult, started time.Time, format string, a ...any) hostResult {
	r.Status = "FAIL"
	r.Detail = fmt.Sprintf(format, a...)
	r.Seconds = time.Since(started).Seconds()
	return r
}

// resetOneHost runs the full sequence for one account: presence probe →
// optional session login → fresh-vector skill patch → fatigue 0 → inventory
// wipe → bank wipe → coin grant → verification read-backs → logout.
func resetOneHost(c *Client, host string, o resetOpts) hostResult {
	started := time.Now()
	r := hostResult{Host: host, Coins: o.coins}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	p := "/players/" + url.PathEscape(host)

	// 1. Presence probe. "Player is not online" is the expected offline shape;
	// any other error (bad token, API down) fails the host immediately.
	_, err := c.get(p, nil)
	switch {
	case err == nil:
		r.AlreadyOnline = true
	case isNotOnlineErr(err):
		// Offline — open a throwaway session to bring the player into the world.
		password := os.Getenv(o.passwordEnv)
		if password == "" {
			return failResult(r, started, "offline and $%s is empty — cannot log in", o.passwordEnv)
		}
		sess, err := openSession(ctx, o.server, host, password, sessionLogger())
		if err != nil {
			return failResult(r, started, "session login: %v", err)
		}
		defer sess.close(ctx)
		r.SessionOpened = true
		if err := waitOnline(ctx, c, p, o.loginWait); err != nil {
			return failResult(r, started, "%v", err)
		}
	default:
		return failResult(r, started, "presence probe: %v", err)
	}

	// 2. Read the live skill list and patch every skill to its fresh level.
	skills, err := getSkills(c, p)
	if err != nil {
		return failResult(r, started, "read skills: %v", err)
	}
	body, err := freshSkillsBody(skills)
	if err != nil {
		return failResult(r, started, "%v", err)
	}
	if _, err := c.patch(p+"/skills", body); err != nil {
		return failResult(r, started, "patch skills: %v", err)
	}
	r.SkillsReset = len(skills)

	// 3. Fatigue back to the fresh-account value.
	if _, err := c.patch(p+"/fatigue", map[string]any{"fatigue": 0}); err != nil {
		return failResult(r, started, "patch fatigue: %v", err)
	}

	// 4. Wipe inventory and bank, then grant exactly the coin stake. Wipe
	// before grant keeps the sweep idempotent: re-running never accumulates.
	if _, err := c.del(p+"/inventory", nil); err != nil {
		return failResult(r, started, "wipe inventory: %v", err)
	}
	if _, err := c.del(p+"/bank", nil); err != nil {
		return failResult(r, started, "wipe bank: %v", err)
	}
	if o.coins > 0 {
		grant := map[string]any{"catalogId": coinsCatalogID, "amount": o.coins, "noted": false}
		raw, err := c.post(p+"/inventory/items", grant)
		if err != nil {
			return failResult(r, started, "grant coins: %v", err)
		}
		// The grant uses "existing inventory add semantics" (openapi.json) —
		// a 200 can still carry added:false (e.g. player not world-registered).
		var res struct {
			Added *bool `json:"added"`
		}
		if err := json.Unmarshal(raw, &res); err == nil && res.Added != nil && !*res.Added {
			return failResult(r, started, "grant coins: server reported added=false")
		}
	}

	// 5. Verify by reading everything back through the admin API. Mutations go
	// through the server's AdminActionQueue and apply on a GAME TICK (~650ms),
	// so a read fired straight after the last POST can race the queue (the
	// drone150 probe failed exactly this way: coins granted, read too early).
	// Poll each check briefly instead of trusting one read.
	verifyDeadline := time.Now().Add(5 * time.Second)
	var verr error
	for {
		verr = nil
		skills, err = getSkills(c, p)
		if err != nil {
			verr = fmt.Errorf("verify read skills: %v", err)
		} else if err := verifySkills(skills); err != nil {
			verr = err
		} else if inv, err := getItems(c, p+"/inventory", "inventory"); err != nil {
			verr = fmt.Errorf("verify read inventory: %v", err)
		} else if o.coins > 0 {
			verr = verifyInventory(inv, o.coins)
		} else if len(inv) != 0 {
			verr = fmt.Errorf("verify inventory: %d stacks remain, want 0", len(inv))
		}
		if verr == nil || time.Now().After(verifyDeadline) {
			break
		}
		time.Sleep(650 * time.Millisecond)
	}
	if verr != nil {
		return failResult(r, started, "%v (after retries)", verr)
	}
	bank, err := getItems(c, p+"/bank", "bank")
	if err != nil {
		return failResult(r, started, "verify read bank: %v", err)
	}
	if err := verifyBank(bank); err != nil {
		return failResult(r, started, "%v", err)
	}
	if fatigue, err := getFatigue(c, p); err != nil {
		return failResult(r, started, "verify read fatigue: %v", err)
	} else if fatigue != 0 {
		return failResult(r, started, "verify fatigue: %d, want 0", fatigue)
	}

	r.Status = "PASS"
	r.Seconds = time.Since(started).Seconds()
	return r
}

// waitOnline polls the presence endpoint until the freshly logged-in player
// is visible to the admin API (login → world registration is async).
func waitOnline(ctx context.Context, c *Client, playerPath string, wait time.Duration) error {
	deadline := time.Now().Add(wait)
	for {
		_, err := c.get(playerPath, nil)
		if err == nil {
			return nil
		}
		if !isNotOnlineErr(err) {
			return fmt.Errorf("waiting for login: %w", err)
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("logged in but not visible to the admin API after %s", wait)
		}
		select {
		case <-time.After(300 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// isNotOnlineErr matches the admin API's "Player is not online: <name>" error
// (AdminRouter.playerByName), the expected shape for an offline account.
func isNotOnlineErr(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "not online")
}

// getSkills fetches and decodes GET {playerPath}/skills.
func getSkills(c *Client, playerPath string) ([]skillInfo, error) {
	raw, err := c.get(playerPath+"/skills", nil)
	if err != nil {
		return nil, err
	}
	var out struct {
		Skills []skillInfo `json:"skills"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode skills: %w", err)
	}
	return out.Skills, nil
}

// getItems fetches an inventory/bank payload and decodes its item array.
// The API nests the array under the container object — {"inventory":
// {"items": [...]}} / {"bank": {"items": [...]}} (verified against the
// live openapi.json + a mid-session player read; the flat shape this
// originally expected made every verify see zero items).
func getItems(c *Client, path, container string) ([]itemEntry, error) {
	raw, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	src := out
	if v, ok := out[container]; ok {
		var inner map[string]json.RawMessage
		if err := json.Unmarshal(v, &inner); err == nil {
			src = inner
		}
	}
	var items []itemEntry
	if v, ok := src["items"]; ok {
		if err := json.Unmarshal(v, &items); err != nil {
			return nil, fmt.Errorf("decode %s items: %w", path, err)
		}
	}
	return items, nil
}

// getFatigue reads the fatigue field off the player snapshot.
func getFatigue(c *Client, playerPath string) (int, error) {
	raw, err := c.get(playerPath, nil)
	if err != nil {
		return 0, err
	}
	var out struct {
		Fatigue int `json:"fatigue"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return 0, fmt.Errorf("decode player: %w", err)
	}
	return out.Fatigue, nil
}

// sessionLogger keeps the session package's logging at warn+ so a 100-host
// sweep doesn't drown the report (and login codes still surface in errors).
func sessionLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

// printDryRun renders the would-be sequence for the FIRST host only. No
// connection of any kind is made — this is a pure print.
func printDryRun(o resetOpts) {
	host := o.hosts[0]
	fmt.Printf("reset DRY RUN — no connections made. Sequence for %q (1 of %d hosts):\n\n", host, len(o.hosts))
	for i, step := range resetPlan(host, o.coins, authenticSkills) {
		switch step.Kind {
		case "session":
			fmt.Printf("%2d. [session] %-7s %s\n", i+1, step.Method, step.Body)
		default:
			fmt.Printf("%2d. %-6s %s\n", i+1, step.Method, step.Path)
			if step.Body != "" {
				fmt.Printf("      %s\n", step.Body)
			}
		}
	}
	fmt.Printf("\nThen: %s between hosts; same sequence for each of the %d hosts.\n", o.ramp, len(o.hosts))
	fmt.Printf("Auth: admin API token via -token/$OPENRSC_ADMIN_API_TOKEN; account password via $%s (never printed).\n", o.passwordEnv)
}

// printHostResult emits one aligned report line per host as the sweep runs.
func printHostResult(r hostResult) {
	if *jsonOut {
		return // the final JSON array carries everything
	}
	state := "was-online"
	if r.SessionOpened {
		state = "logged-in"
	}
	if r.Status == "PASS" {
		fmt.Printf("%-12s PASS  (%s, %d skills reset, %d coins, bank empty, %.1fs)\n",
			r.Host, state, r.SkillsReset, r.Coins, r.Seconds)
		return
	}
	fmt.Printf("%-12s FAIL  %s (%.1fs)\n", r.Host, r.Detail, r.Seconds)
}
