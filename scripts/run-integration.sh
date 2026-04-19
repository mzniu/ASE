#!/usr/bin/env bash
# Optional OpenSearch integration tests (same Go version as go.mod).
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
exec go test -tags=integration -count=1 ./...
