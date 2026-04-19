#!/usr/bin/env bash
# Go：gofmt 检查、vet、test（与 CI / Makefile check 一致）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

if [[ ! -f go.mod ]]; then
	echo "No go.mod; skip check-go.sh"
	exit 0
fi

out=$(find . -name '*.go' -not -path './vendor/*' -exec gofmt -l {} + 2>/dev/null || true)
if [[ -n "${out}" ]]; then
	echo "gofmt needed on:" >&2
	echo "${out}" >&2
	echo "Run: go fmt ./..." >&2
	exit 1
fi

echo "Running: go vet ./..."
go vet ./...

echo "Running: go test ./..."
go test ./...
