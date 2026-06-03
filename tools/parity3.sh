#!/usr/bin/env bash
# parity3.sh — render the SAME rscdump fixture through all three render-parity
# legs (orsc / DEOB / JAR) with identical env, then print the three pairwise
# max-channel diffs (orsc<->DEOB, orsc<->JAR, DEOB<->JAR). Goal: 0/0/0 each.
#
# All gate envs (RSC_MESH_NPC / RSC_MESH_PLAYER / RSC_MESH_ITEM / RSC_NPC_*) are
# inherited from the caller's environment so a pose is selected by exporting them
# before invoking. RSC_MESH_CACHE defaults to /tmp/rsc-run/cache; ORSC_FLAT_AMBIENCE
# is forced on for the orsc leg only (matches the JAR/DEOB ambience RNG pin).
#
# usage:  RSC_MESH_NPC=19 tools/parity3.sh [FIXTURE] [OUTDIR]
set -u
WW=/home/free/code/rsc-hacking/westworld
RSCPLUS=/home/free/code/rsc-hacking/rscplus
JAR=$RSCPLUS/assets/rsclassic-1091943135.jar
# The rev-235 client needs java.applet.Applet (removed in JDK 21+); the box default
# java is JDK 26, so the DEOB/JAR legs MUST run on JDK 17.
JAVA17="${JAVA17:-/usr/lib/jvm/java-17-openjdk/bin/java}"
FIXTURE="${1:-$WW/testdata/rscdump/hunt/door_diag_obj.json}"
OUT="${2:-/tmp/parity}"
CACHE="${RSC_MESH_CACHE:-/tmp/rsc-run/cache}"
mkdir -p "$OUT"
export RSC_MESH_CACHE="$CACHE"

echo "== fixture=$FIXTURE  cache=$CACHE  gates: NPC=${RSC_MESH_NPC:-} PLAYER=${RSC_MESH_PLAYER:-} ITEM=${RSC_MESH_ITEM:-} PHASE2=${RSC_NPC_PHASE2:-} DEBUG=${RSC_NPC_DEBUG_BILLBOARD:-}"

# orsc leg
( cd "$WW" && ORSC_FLAT_AMBIENCE=1 go run ./cmd/meshrender -fixture "$FIXTURE" -out "$OUT/orsc.png" -cache "$CACHE" ) || { echo "orsc render FAILED"; exit 1; }
# DEOB leg
( cd "$WW" && "$JAVA17" -Djava.awt.headless=true -cp /tmp/deob-run client.DumpRender "$FIXTURE" "$OUT" deob ) || { echo "DEOB render FAILED"; exit 1; }
# JAR leg
( cd "$RSCPLUS/dumprender" && "$JAVA17" -Djava.awt.headless=true \
    -cp ".:$RSCPLUS/lib/asm-5.0.4.jar:$RSCPLUS/lib/asm-tree-5.0.4.jar:$RSCPLUS/lib/asm-util-5.0.4.jar:$RSCPLUS/lib/json-20201115.jar" \
    DumpRenderer "$JAR" "$FIXTURE" "$OUT" jar ) || { echo "JAR render FAILED"; exit 1; }

echo "== diffs (goal 0/0/0 each):"
python3 "$WW/tools/pngdiff.py" "$OUT/orsc.png" "$OUT/deob.png" --label "orsc<->DEOB"
python3 "$WW/tools/pngdiff.py" "$OUT/orsc.png" "$OUT/jar.png"  --label "orsc<->JAR "
python3 "$WW/tools/pngdiff.py" "$OUT/deob.png" "$OUT/jar.png"  --label "DEOB<->JAR "
