#!/usr/bin/env bash
# Run Wails dev with a working Go module proxy (always override a broken global GOPROXY).
set -euo pipefail
cd "$(dirname "$0")/.."

export GOPROXY="${GO_MOD_PROXY:-https://proxy.golang.org,direct}"
exec wails dev "$@"
