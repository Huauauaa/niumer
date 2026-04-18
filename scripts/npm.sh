#!/usr/bin/env bash
# Run npm for frontend with a clean environment. Use this if plain `npm` exits
# with no output (often caused by a broken ~/.npmrc).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
NPM_BIN="$(command -v npm)"
NODE_BIN="$(command -v node)"
NODE_DIR="$(dirname "$NODE_BIN")"
NPM_HOME="${NPM_HOME:-${TMPDIR:-/tmp}/niumer-npm-home}"
mkdir -p "$NPM_HOME"
exec env -i \
  PATH="$NODE_DIR:/usr/bin:/bin:/usr/sbin" \
  HOME="$NPM_HOME" \
  NPM_CONFIG_USERCONFIG=/dev/null \
  "$NPM_BIN" "$@" --prefix "$ROOT/frontend"
