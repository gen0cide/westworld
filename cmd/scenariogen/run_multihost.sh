#!/usr/bin/env bash
# Run the multi-host scenarios by pairing a primary drone with a
# partner host that runs a complementary responder routine (or just
# idles online). The single-host parallel runner skips these because
# they need a coordinated second party (trade/duel acceptor, summon
# target, follow target, PvP opponent).
#
# Coordination: launch the PARTNER first (it logs in, and responder
# routines wait ~8s for the primary to summon them, then enter their
# listen loop). Sleep a beat, then run the PRIMARY (which summons the
# partner via admin "::<name> to" and drives the interaction). We
# never kill the partner — it exits on its own (graceful logout +
# reset-on-exit), which keeps its account cleanly released so the
# NEXT pairing that reuses it logs in without a code-4 race.
#
# Usage: ./run_multihost.sh [server]
# Results appended to /tmp/scenario_results.tsv (MH rows).

set -u
SERVER="${1:-localhost:43594}"
RESULTS=/tmp/scenario_results.tsv
cd "$(dirname "$0")/../.." || exit 1   # repo root
SC=examples/scenarios

if [ -f .local.env ]; then set -a; source .local.env; set +a; fi
[ -n "${WESTWORLD_PASSWORD:-}" ] || { echo "no WESTWORLD_PASSWORD" >&2; exit 1; }
go build -o /tmp/cradle ./cmd/legacy-cradle || { echo "cradle build failed" >&2; exit 1; }

# facts MUST stay enabled (default openrsc root): they resolve type_id→NPC
# name and def_id→scenery name. -facts "" silently breaks every name-based
# find() (n.name / world.locs.search), aborting NPC/scenery scenarios.
R() { /tmp/cradle -server "$SERVER" -reset-on-exit "$@"; }

# run_pair <scenario_file> <primary_user> <partner_user> <partner_routine|"">
run_pair() {
  local sc="$1" primary="$2" partner="$3" proutine="$4"
  local label; label="MH/$(basename "$sc" .routine)"
  local plog; plog=$(mktemp /tmp/mh_partner.XXXX)

  # Partner up first (background). Responder routines self-wait for
  # the summon; idle partners just dwell online.
  if [ -n "$proutine" ]; then
    R -username "$partner" -routine "$proutine" -dwell 75s >"$plog" 2>&1 &
  else
    R -username "$partner" -dwell 45s >"$plog" 2>&1 &
  fi
  local ppid=$!
  sleep 5   # partner: login + ready + (responder) enter listen loop

  # Primary drives the interaction.
  local raw
  raw=$(R -username "$primary" -routine "$sc" -dwell 45s 2>&1 | grep -E "routine ended|run failed" | head -1)
  local outcome
  case "$raw" in
    *"kind=returned"*"PASS:"*) outcome="PASS";;
    *"kind=returned"*)         outcome="OK: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
    *"kind=aborted"*)          outcome="ABORT: $(echo "$raw" | sed -E 's/.*value="([^"]+)".*/\1/')";;
    *"kind=errored"*)          outcome="ERROR: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
    *"run failed"*)            outcome="LOGIN/SETUP FAIL: $(echo "$raw" | sed -E 's/.*err="([^"]+)".*/\1/')";;
    *)                          outcome="??: $raw";;
  esac
  printf "%s\t%s\n" "$label" "$outcome" | tee -a "$RESULTS"

  # Let the partner finish + log out gracefully (clean account release
  # for the next pairing). Don't kill it.
  wait "$ppid" 2>/dev/null
  # Surface the partner's terminal note for diagnosis.
  echo "    partner($partner): $(grep -oE 'routine ended.*' "$plog" | head -1 | sed -E 's/.*value=([^ ]+).*/\1/')" >&2
  rm -f "$plog"
}

echo "=== multi-host scenarios ===" >&2
run_pair "$SC/social/summon_bernard_to_self.routine"      drone3 bernard ""
run_pair "$SC/social/private_message_bernard.routine"     drone3 bernard ""
run_pair "$SC/social/social_follow_bernard.routine"       drone3 bernard ""
run_pair "$SC/social/social_trade_initiate_bernard.routine" drone3 bernard examples/routines/bernard_respond_trade.routine
run_pair "$SC/social/social_trade_offer_decline.routine"  drone3 bernard examples/routines/bernard_respond_trade.routine
run_pair "$SC/social/social_duel_initiate_bernard.routine" drone3 bernard examples/routines/bernard_respond_duel.routine
run_pair "$SC/combat/two_drone_attack_each_other.routine" drone1 drone2 ""

echo "=== multi-host done ===" >&2
