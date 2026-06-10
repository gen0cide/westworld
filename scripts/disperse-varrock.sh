#!/usr/bin/env bash
# disperse-varrock.sh — re-teleport each drone onto a DISTINCT, pre-validated
# Varrock path tile, so the fleet spreads across the city's cement paths instead
# of piling on the town point (122,509) and dispersing slowly by wandering.
#
# The tile file (default /tmp/varrock-paths.txt, "x y" per line) is produced by:
#   varrock-tiles -minx 108 -maxx 144 -miny 486 -maxy 531 -paths -n 200 -o /tmp/varrock-paths.txt
# Every tile in it is overlay==1 (cement path), walkable, and reachable from the
# town centre — so no drone lands in an object, on grass, or boxed off.
#
# drone<i> gets the i-th tile (1-based). If there are more drones than tiles, it
# wraps (a few share a tile). Teleport uses the COMMAND opcode, NOT chat.
#
#   scripts/disperse-varrock.sh                 # drone1..drone200 → tiles
#   COUNT=200 BATCH=25 DELAY=3 scripts/disperse-varrock.sh
#   TILES=/tmp/varrock-paths.txt scripts/disperse-varrock.sh

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
COUNT="${COUNT:-200}"
START="${START:-1}"
BATCH="${BATCH:-25}"
DELAY="${DELAY:-3}"
CRADLE="${CRADLE:-localhost:8099}"
TILES="${TILES:-/tmp/varrock-paths.txt}"

if [[ ! -f "${TILES}" ]]; then
    echo "disperse: tile file ${TILES} not found; generate it with varrock-tiles -paths -o ${TILES}" >&2
    exit 2
fi

# Read tiles into arrays.
mapfile -t TILELINES < "${TILES}"
NTILES=${#TILELINES[@]}
if (( NTILES == 0 )); then echo "disperse: ${TILES} is empty" >&2; exit 2; fi
echo "disperse: ${NTILES} path tiles loaded from ${TILES}"

BIN="$(mktemp -d)"; trap 'rm -rf "${BIN}"' EXIT
( cd "${REPO_ROOT}" && go build -o "${BIN}/cradle-ctl" ./cmd/cradle-ctl )

END=$(( START + COUNT - 1 ))
echo "disperse: teleporting drone${START}..drone${END} onto distinct path tiles (batch ${BATCH}, ${DELAY}s apart)"
n=0
for (( i=START; i<=END; i++ )); do
    line="${TILELINES[$(( (i-START) % NTILES ))]}"
    x="${line%% *}"; y="${line##* }"
    "${BIN}/cradle-ctl" -server "${CRADLE}" eval "drone${i}" "command(\"teleport ${x} ${y}\")" >/dev/null \
        && echo "  → drone${i} → (${x},${y})" \
        || echo "  ✗ drone${i} eval failed" >&2
    n=$(( n + 1 ))
    if (( n % BATCH == 0 )) && (( i < END )); then sleep "${DELAY}"; fi
done
echo "disperse: done — ${COUNT} drones placed on Varrock path tiles."
