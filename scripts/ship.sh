#!/bin/sh
# The ONLY way we deploy. Two-person project, one deploy target: this machine.
#
# Contract (docs/development-workflow.md):
#   - main is the single shippable truth; deploys are a function of main ONLY.
#   - This script refuses to ship anything but a clean checkout of origin/main.
#   - All binaries build together from one SHA (no mixed-build fleets) and the
#     SHA is stamped into them (-X main.build) and into /tmp/ww-bin/DEPLOYED.
#
# Usage: scripts/ship.sh [hostcfg]        (default /tmp/society.hostcfg)
#   SKIP_TESTS=1 scripts/ship.sh          (emergency redeploy; gates already green)
set -e
cd "$(dirname "$0")/.."

CFG="${1:-/tmp/society.hostcfg}"
[ -f "$CFG" ] || { echo "ship: hostcfg not found: $CFG" >&2; exit 1; }
[ -z "$(git status --porcelain)" ] || { echo "ship: working tree dirty — commit or stash first" >&2; exit 1; }

git checkout -q main
git pull --ff-only -q origin main   # refuses divergence: local main == origin/main or we stop
SHA=$(git rev-parse --short HEAD)
echo "ship: building $SHA"

go build ./...
./scripts/vet.sh
if [ -z "$SKIP_TESTS" ]; then go test ./... -count=1 >/dev/null; echo "ship: tests green"; fi

mkdir -p /tmp/ww-bin
for b in cradle-server mesad cradle-ctl mesa-ctl; do
  go build -ldflags "-X main.build=$SHA" -o "/tmp/ww-bin/$b.new" "./cmd/$b"
done

echo "ship: stopping fleet"
pkill -f mesad-supervise 2>/dev/null || true
pkill -x mesad           2>/dev/null || true
pkill -f '/tmp/ww-bin/cradle-server' 2>/dev/null || true
sleep 3
pgrep -x mesad >/dev/null && { pkill -9 -x mesad; sleep 1; }   # graceful drain can wedge (2026-06-11)

for b in cradle-server mesad cradle-ctl mesa-ctl; do mv "/tmp/ww-bin/$b.new" "/tmp/ww-bin/$b"; done
echo "$SHA $(date '+%F %T')" > /tmp/ww-bin/DEPLOYED

for f in /tmp/ww-mesad.log /tmp/ww-cradle.log; do [ -f "$f" ] && mv "$f" "$f.prev"; done

set -a; . ./.local.env; set +a
nohup /tmp/ww-bin/mesad-supervise.sh >/dev/null 2>&1 &
sleep 4
nc -z localhost 7077 || { echo "ship: mesad failed to come up" >&2; exit 1; }
nohup /tmp/ww-bin/cradle-server -config "$CFG" -ramp 2s >> /tmp/ww-cradle.log 2>&1 &
sleep 10

HOSTS=$(curl -s --max-time 8 localhost:8099/api/hosts | python3 -c 'import json,sys; print(len(json.load(sys.stdin)))' 2>/dev/null || echo 0)
echo "ship: deployed $SHA — $HOSTS hosts registering (ramp in progress), build stamps in startup logs"
