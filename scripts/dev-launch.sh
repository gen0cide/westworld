#!/usr/bin/env bash
# Developer launcher — sources .local.env (gitignored) for secrets,
# then execs cradle. Use this instead of inlining WESTWORLD_PASSWORD
# in shell history or commit messages.
#
# Setup: create .local.env in the repo root with
#   export WESTWORLD_PASSWORD=...
# (any other dev-only env vars also go here).
#
# Usage:
#   ./scripts/dev-launch.sh -repl -username delores -server localhost:43594
#   ./scripts/dev-launch.sh -username delores -routine examples/routines/say_hello.routine
#
# Anything after the script name is forwarded as cradle's args.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ENV_FILE="${REPO_ROOT}/.local.env"

if [[ ! -f "${ENV_FILE}" ]]; then
    echo "dev-launch: missing ${ENV_FILE}" >&2
    echo "  create it with at minimum:" >&2
    echo "    export WESTWORLD_PASSWORD=<value>" >&2
    exit 2
fi

# Source the env file with `set -a` so plain `KEY=value` lines
# (without `export`) still become exported env vars for the
# cradle subprocess. `set +a` restores normal behavior.
set -a
# shellcheck disable=SC1090
source "${ENV_FILE}"
set +a

if [[ -z "${WESTWORLD_PASSWORD:-}" ]]; then
    echo "dev-launch: .local.env did not export WESTWORLD_PASSWORD" >&2
    exit 2
fi

# Build cradle if the binary is missing or stale relative to sources.
CRADLE_BIN="${REPO_ROOT}/cradle"
if [[ ! -x "${CRADLE_BIN}" ]] || [[ -n "$(find "${REPO_ROOT}" -name '*.go' -newer "${CRADLE_BIN}" -not -path '*/local/*' -not -path '*/.git/*' 2>/dev/null | head -1)" ]]; then
    echo "[dev-launch] (re)building cradle..."
    (cd "${REPO_ROOT}" && go build -o cradle ./cmd/cradle)
fi

exec "${CRADLE_BIN}" "$@"
