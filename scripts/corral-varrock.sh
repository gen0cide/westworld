#!/usr/bin/env bash
# corral-varrock.sh — get the drone fleet into Varrock and keep them roaming
# inside the gates, with NO restart of mesad or the cradle.
#
# Two levers, both live:
#   1. A SOFT goal push (mesa-ctl goal push, --match 'drone*') sets every drone's
#      runtime goal to "roam within Varrock, don't pass the gates". It's a bias,
#      not a fence — there is no boundary arbiter; the director just prefers it.
#   2. A 1:1 teleport: command("goto Varrock") on each host via cradle-ctl eval,
#      in batches (default 10) spaced out (default 30s) so they fan out as they
#      land instead of stacking on one square. NOTE: it's the COMMAND opcode
#      (the "::" admin command), NOT chat — say("::goto…") sends literal chat
#      text and does nothing; the command() DSL verb sends opcode 38.
#
# DE-RISK FIRST: run the smoke mode (one drone) and confirm goto actually
# teleports on this OpenRSC server before sweeping all 200:
#   scripts/corral-varrock.sh --smoke
#
# Then the full run:
#   scripts/corral-varrock.sh                 # goal push + teleport drone1..drone200
#   scripts/corral-varrock.sh --no-goal       # teleport only (don't touch goals)
#   COUNT=50 BATCH=5 DELAY=20 scripts/corral-varrock.sh
#
# Env / flags:
#   COUNT   (200)  how many drones (drone<START>..drone<START+COUNT-1>)
#   START   (1)    first drone index
#   BATCH   (10)   teleports per batch
#   DELAY   (30)   seconds to wait between batches
#   CRADLE  (localhost:8099)  cradle-server control plane
#   MESAD   (localhost:7077)  mesad gRPC (for the goal push)
#   --smoke        teleport ONLY drone<START>, skip the goal push (verification)
#   --no-goal      skip the goal push; teleport only
#   --goal-only    push the goal; do NOT teleport

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${REPO_ROOT}/.local.env"

COUNT="${COUNT:-200}"
START="${START:-1}"
BATCH="${BATCH:-10}"
DELAY="${DELAY:-30}"
CRADLE="${CRADLE:-localhost:8099}"
MESAD="${MESAD:-localhost:7077}"

SMOKE=0; NO_GOAL=0; GOAL_ONLY=0
for arg in "$@"; do
    case "$arg" in
        --smoke)     SMOKE=1 ;;
        --no-goal)   NO_GOAL=1 ;;
        --goal-only) GOAL_ONLY=1 ;;
        *) echo "corral-varrock: unknown flag $arg" >&2; exit 2 ;;
    esac
done

# The soft Varrock-roam disposition. A goal instruction only — never a veto.
VARROCK_GOAL="You are in the city of Varrock. Roam and explore freely within it — wander the streets, the central square, the marketplace, and the shops, and keep moving so you don't cluster with others. Do NOT leave Varrock: stay inside the city and never pass through its gates out into the wilderness or countryside. If you reach a gate or the edge of the city, turn back inward and keep exploring Varrock."

# Source secrets (ADMIN_TOKEN for the goal push) if present.
if [[ -f "${ENV_FILE}" ]]; then
    set -a; source "${ENV_FILE}"; set +a
fi

# Build the two CLIs once into a temp dir (always current with the tree).
BIN="$(mktemp -d)"
trap 'rm -rf "${BIN}"' EXIT
echo "corral: building mesa-ctl + cradle-ctl…"
( cd "${REPO_ROOT}" && go build -o "${BIN}/mesa-ctl" ./cmd/mesa-ctl && go build -o "${BIN}/cradle-ctl" ./cmd/cradle-ctl )

push_goal() {
    if [[ -z "${ADMIN_TOKEN:-}" ]]; then
        echo "corral: ADMIN_TOKEN not set (need it for the goal push); put it in ${ENV_FILE}" >&2
        exit 2
    fi
    echo "corral: pushing soft Varrock-roam goal to drone* …"
    "${BIN}/mesa-ctl" -addr "${MESAD}" goal push "${VARROCK_GOAL}" --match 'drone*'
}

teleport() {
    local i="$1"
    # command("goto Varrock") → COMMAND opcode (the "::" admin command), self-tele
    # to the varrock town point (122,509). NOT say() — that's chat and is ignored.
    "${BIN}/cradle-ctl" -server "${CRADLE}" eval "drone${i}" 'command("goto Varrock")' \
        && echo "  → drone${i} goto Varrock" \
        || echo "  ✗ drone${i} eval failed (is it running?)" >&2
}

# --- smoke: one drone, no goal change, verify ::goto works -------------------
if [[ "${SMOKE}" -eq 1 ]]; then
    echo "corral: SMOKE — teleporting only drone${START}; verify it lands in Varrock"
    teleport "${START}"
    echo "corral: check with  cradle-ctl -server ${CRADLE} state drone${START}  (expect Varrock coords, ~120,500)"
    exit 0
fi

# --- goal push (unless suppressed) ------------------------------------------
if [[ "${NO_GOAL}" -eq 0 ]]; then
    push_goal
fi
if [[ "${GOAL_ONLY}" -eq 1 ]]; then
    echo "corral: --goal-only, done."
    exit 0
fi

# --- batched teleport sweep -------------------------------------------------
END=$(( START + COUNT - 1 ))
echo "corral: teleporting drone${START}..drone${END} in batches of ${BATCH}, ${DELAY}s apart"
n=0
for (( i=START; i<=END; i++ )); do
    teleport "${i}"
    n=$(( n + 1 ))
    if (( n % BATCH == 0 )) && (( i < END )); then
        echo "corral: batch done (${n}/${COUNT}); waiting ${DELAY}s for fan-out…"
        sleep "${DELAY}"
    fi
done
echo "corral: all ${COUNT} teleport commands sent."
