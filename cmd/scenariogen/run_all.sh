#!/usr/bin/env bash
# Sequentially run every single-host scenario in examples/scenarios/
# against the named drone (default Drone3). Captures the routine
# outcome and writes a summary to stdout + /tmp/scenario_results.tsv.
#
# Multi-host scenarios (any with `Hosts: any-drone-and-bernard` or
# `Drone1-and-Drone2`) are skipped — running them sequentially
# would deadlock waiting for the other host.
#
# Usage: ./run_all.sh [username] [server]

set -u
USER="${1:-drone3}"
SERVER="${2:-localhost:43594}"
RESULTS=/tmp/scenario_results.tsv
: > "$RESULTS"

# Load WESTWORLD_PASSWORD from .local.env
if [ -f .local.env ]; then set -a; source .local.env; set +a; fi
if [ -z "${WESTWORLD_PASSWORD:-}" ]; then
  echo "no WESTWORLD_PASSWORD; aborting" >&2
  exit 1
fi

# Ensure cradle is built
go build -o /tmp/cradle ./cmd/legacy-cradle || { echo "cradle build failed" >&2; exit 1; }

total=0; pass=0; fail=0; abort=0; err=0; skip=0
printf "%-65s %s\n" "SCENARIO" "OUTCOME"
printf "%-65s %s\n" "$(printf '=%.0s' {1..65})" "======="

for f in examples/scenarios/*/*.routine; do
  total=$((total + 1))
  name=$(basename "$f" .routine)
  cat=$(basename "$(dirname "$f")")
  label="$cat/$name"

  if grep -qE "^# Hosts:.*(bernard|Drone[12])" "$f" && ! grep -q "^# Hosts: any-drone$" "$f"; then
    printf "%-65s %s\n" "$label" "SKIP (multi-host)"
    printf "%s\tSKIP\tmulti-host\n" "$label" >> "$RESULTS"
    skip=$((skip + 1))
    continue
  fi

  # NOTE: a previous version of this script ran a reset routine
  # via a fresh cradle invocation between scenarios. That caused
  # login thrash on the server and broke 54+ scenarios in sweep #6.
  # The right approach is in-process reset via admin commands at
  # the top of each scenario's setup (wipeinv / heal / recharge) —
  # easier and avoids the login storm. See task #110 follow-up.

  to=$(grep -oE "^# Timeout: [0-9]+" "$f" | head -1 | awk '{print $3}')
  to=${to:-30}
  dwell=$((to + 5))

  # Run scenario; capture the FULL output so a non-PASS outcome can be
  # explained with the server messages the bot actually received.
  full=$(mktemp /tmp/run_all.XXXXXX)
  /tmp/cradle -username "$USER" -server "$SERVER" -routine "$f" -dwell "${dwell}s" -facts "" > "$full" 2>&1
  raw=$(grep -E "routine ended|run failed" "$full" | head -1)

  case "$raw" in
    *"kind=returned"*"PASS:"*) outcome="PASS"; pass=$((pass + 1));;
    *"kind=returned"*"FAIL:"*) outcome="FAIL: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')"; fail=$((fail + 1));;
    *"kind=aborted"*)          outcome="ABORT: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')"; abort=$((abort + 1));;
    *"kind=errored"*)          outcome="ERROR: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')"; err=$((err + 1));;
    *"run failed"*)            outcome="LOGIN/SETUP FAIL: $raw"; err=$((err + 1));;
    *)                          outcome="??: $raw"; err=$((err + 1));;
  esac

  printf "%-65s %s\n" "$label" "$outcome"
  printf "%s\t%s\n" "$label" "$outcome" >> "$RESULTS"

  # On any non-PASS, dump the player-facing server messages so the
  # sweep log says WHY it failed (the cradle logs each as a `server
  # msg` line — runtime/host.go). PASS lines stay terse.
  if [ "$outcome" != "PASS" ]; then
    msgs=$(grep -F '"server msg"' "$full" | sed -E 's/.*msg="server msg" //' | tail -20)
    if [ -n "$msgs" ]; then
      while IFS= read -r line; do printf "    ↳ %s\n" "$line"; done <<< "$msgs"
    else
      printf "    ↳ (no server messages captured)\n"
    fi
  fi
  rm -f "$full"
done

echo
echo "==================== SUMMARY ===================="
echo "total: $total  pass: $pass  fail: $fail  abort: $abort  err: $err  skip: $skip"
echo "tsv: $RESULTS"
