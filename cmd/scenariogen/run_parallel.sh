#!/usr/bin/env bash
# Parallel scenario sweep across the drone1..drone20 pool.
#
# Each drone is a dedicated worker that runs its slice of the catalog
# SERIALLY (one scenario fully finishes — including the cradle's
# graceful logout — before that drone starts its next). All 20 drones
# run concurrently. This is both ~15-20x faster than the sequential
# runner AND structurally avoids the login-code-4 race: a given
# account never re-logs within seconds of itself, since its prior
# scenario has fully logged out (LogoutGraceful) before the next.
#
# Multi-host scenarios (Hosts referencing bernard / Drone1 / Drone2)
# are skipped — they need a second coordinated host.
#
# Usage: ./run_parallel.sh [server] [num_drones]
#   server     default localhost:43594
#   num_drones default 20  (uses drone1..droneN)
#
# Results: /tmp/scenario_results.tsv  (scenario<TAB>outcome)

set -u
SERVER="${1:-localhost:43594}"
NDRONES="${2:-20}"
RESULTS=/tmp/scenario_results.tsv
WORKDIR=$(mktemp -d /tmp/sweep.XXXXXX)
: > "$RESULTS"

cd "$(dirname "$0")/../.." || exit 1   # repo root

if [ -f .local.env ]; then set -a; source .local.env; set +a; fi
if [ -z "${WESTWORLD_PASSWORD:-}" ]; then
  echo "no WESTWORLD_PASSWORD; aborting" >&2; exit 1
fi
go build -o /tmp/cradle ./cmd/cradle || { echo "cradle build failed" >&2; exit 1; }

# Collect the single-host scenarios (skip multi-host).
scenarios=()
for f in examples/scenarios/*/*.routine; do
  if grep -qE "^# Hosts:.*(bernard|Drone[12])" "$f" && ! grep -q "^# Hosts: any-drone$" "$f"; then
    label="$(basename "$(dirname "$f")")/$(basename "$f" .routine)"
    printf "%s\tSKIP\tmulti-host\n" "$label" >> "$RESULTS"
    continue
  fi
  scenarios+=("$f")
done
echo "single-host scenarios: ${#scenarios[@]}  | drones: $NDRONES  | server: $SERVER" >&2

# One worker per drone. Worker i takes scenarios at indices i, i+N, i+2N, ...
run_worker() {
  local widx="$1" drone="drone$1" out="$WORKDIR/d$1.tsv"
  : > "$out"
  local total="${#scenarios[@]}" i f label to dwell raw outcome
  for (( i=widx-1; i<total; i+=NDRONES )); do
    f="${scenarios[$i]}"
    label="$(basename "$(dirname "$f")")/$(basename "$f" .routine)"
    to=$(grep -oE "^# Timeout: [0-9]+" "$f" | head -1 | awk '{print $3}'); to=${to:-30}
    dwell=$((to + 18))   # +18s headroom for the graceful-logout cooldown
    # NOTE: facts (NpcDefs/GameObjectDef/etc.) MUST stay enabled — they
    # resolve type_id→NPC name and def_id→scenery name. With -facts "" every
    # `n.name == "..."` and `world.locs.search("tree")` returns nothing, which
    # silently aborts ~all NPC/scenery scenarios as "not in view / not nearby".
    # Run, with up to 2 retries on a transient code-4 ("already logged in")
    # login race — the prior account session can linger a few seconds after
    # logout, and a fresh relogin then bounces. A short backoff + retry
    # removes that noise so the sweep reflects real scenario outcomes.
    local attempt
    for attempt in 1 2 3; do
      raw=$(/tmp/cradle -username "$drone" -server "$SERVER" -routine "$f" -dwell "${dwell}s" -reset-on-exit 2>&1 \
              | grep -E "routine ended|run failed" | head -1)
      case "$raw" in
        *"kind=returned"*"PASS:"*) outcome="PASS";;
        *"kind=returned"*"FAIL:"*) outcome="FAIL: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
        *"kind=aborted"*)          outcome="ABORT: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
        *"kind=errored"*)          outcome="ERROR: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
        *"run failed"*)            outcome="LOGIN/SETUP FAIL: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
        *)                          outcome="??: $raw";;
      esac
      case "$outcome" in
        "LOGIN/SETUP FAIL:"*"code 4"*) sleep 4; continue;;  # transient — retry
        *) break;;
      esac
    done
    printf "%s\t%s\n" "$label" "$outcome" >> "$out"
    printf "[%s] %-55s %s\n" "$drone" "$label" "$outcome" >&2
  done
}

for (( w=1; w<=NDRONES; w++ )); do run_worker "$w" & done
wait

# Merge per-drone results, sort, append to the SKIP lines already there.
cat "$WORKDIR"/d*.tsv | sort >> "$RESULTS"
rm -rf "$WORKDIR"

echo >&2
echo "==================== SUMMARY ====================" >&2
awk -F'\t' '{o=$2; sub(/:.*/,"",o); c[o]++} END{for(k in c) printf "%-18s %d\n", k, c[k]}' "$RESULTS" | sort -k2 -rn >&2
echo "tsv: $RESULTS" >&2
