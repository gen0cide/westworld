#!/usr/bin/env bash
# 1-DRONE-PER-TEST parallel sweep. Each single-host scenario gets its OWN
# dedicated drone (scenario[i] -> drone#(i+1)); all run concurrently, each
# drone runs exactly ONE test then logs out. With ~186 single-host scenarios
# and 220 drone accounts this is a true 1:1 fan-out — zero back-to-back reuse,
# wall-clock ≈ the slowest single test (not a serial slice). Also battle-tests
# the MySQL concurrent-login path (the reason we migrated off SQLite).
#
# Launches are lightly staggered (LOGIN_STAGGER) so 186 logins don't hit the
# server in one burst; since per-test dwell (>=48s) >> total stagger (~22s),
# essentially all tests still overlap.
#
# Multi-host scenarios (Hosts referencing bernard / Drone1 / Drone2) need a
# second coordinated host, so they're SKIP'd here (can't be 1 drone / 1 test).
#
# Usage: ./run_one_per_drone.sh [server] [login_stagger_seconds]
#   server  default localhost:43594
#   stagger default 0.12
# Results: /tmp/scenario_results.tsv  (scenario<TAB>outcome)

set -u
SERVER="${1:-localhost:43594}"
STAGGER="${2:-0.12}"
MAXDRONES=220
RESULTS=/tmp/scenario_results.tsv
WORKDIR=$(mktemp -d /tmp/sweep1.XXXXXX)
: > "$RESULTS"

cd "$(dirname "$0")/../.." || exit 1   # repo root
if [ -f .local.env ]; then set -a; source .local.env; set +a; fi
if [ -z "${WESTWORLD_PASSWORD:-}" ]; then echo "no WESTWORLD_PASSWORD; aborting" >&2; exit 1; fi
go build -o /tmp/cradle ./cmd/cradle || { echo "cradle build failed" >&2; exit 1; }

# Single-host scenarios only (skip multi-host pairs).
scenarios=()
for f in examples/scenarios/*/*.routine; do
  if grep -qE "^# Hosts:.*(bernard|Drone[12])" "$f" && ! grep -q "^# Hosts: any-drone$" "$f"; then
    label="$(basename "$(dirname "$f")")/$(basename "$f" .routine)"
    printf "%s\tSKIP\tmulti-host\n" "$label" >> "$RESULTS"
    continue
  fi
  scenarios+=("$f")
done
N=${#scenarios[@]}
[ "$N" -gt "$MAXDRONES" ] && echo "WARNING: $N scenarios > $MAXDRONES drones; last $((N-MAXDRONES)) share via round-robin." >&2
echo "single-host scenarios: $N | one drone each (cap $MAXDRONES) | stagger ${STAGGER}s | server $SERVER" >&2

run_one() {
  local idx="$1"
  local dnum=$(( (idx % MAXDRONES) + 1 )); local drone="drone$dnum"
  local f="${scenarios[$idx]}" out="$WORKDIR/s$idx.tsv" full="$WORKDIR/s$idx.cradle.log"
  local label to dwell raw outcome attempt
  label="$(basename "$(dirname "$f")")/$(basename "$f" .routine)"
  to=$(grep -oE "^# Timeout: [0-9]+" "$f" | head -1 | awk '{print $3}'); to=${to:-30}
  dwell=$((to + 18))
  for attempt in 1 2 3; do
    /tmp/cradle -username "$drone" -server "$SERVER" -routine "$f" -dwell "${dwell}s" -reset-on-exit > "$full" 2>&1
    raw=$(grep -E "routine ended|run failed" "$full" | head -1)
    case "$raw" in
      *"kind=returned"*"PASS:"*) outcome="PASS";;
      *"kind=returned"*"FAIL:"*) outcome="FAIL: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
      *"kind=aborted"*)          outcome="ABORT: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
      *"kind=errored"*)          outcome="ERROR: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
      *"run failed"*)            outcome="LOGIN/SETUP FAIL: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
      *)                          outcome="??: $raw";;
    esac
    case "$outcome" in
      "LOGIN/SETUP FAIL:"*"code 4"*) sleep 4; continue;;
      *) break;;
    esac
  done
  printf "%s\t%s\n" "$label" "$outcome" >> "$out"
  printf "[%s] %-55s %s\n" "$drone" "$label" "$outcome" >&2
  if [ "$outcome" != "PASS" ]; then
    local msgs; msgs=$(grep -F '"server msg"' "$full" | sed -E 's/.*msg="server msg" //' | tail -20)
    if [ -n "$msgs" ]; then
      while IFS= read -r line; do printf "[%s]     ↳ %s\n" "$drone" "$line" >&2; done <<< "$msgs"
    fi
  fi
}

for (( i=0; i<N; i++ )); do run_one "$i" & sleep "$STAGGER"; done
wait

cat "$WORKDIR"/s*.tsv 2>/dev/null | sort >> "$RESULTS"
rm -rf "$WORKDIR"
echo >&2; echo "==================== SUMMARY ====================" >&2
awk -F'\t' '{o=$2; sub(/:.*/,"",o); c[o]++} END{for(k in c) printf "%-18s %d\n", k, c[k]}' "$RESULTS" | sort -k2 -rn >&2
echo "tsv: $RESULTS" >&2
