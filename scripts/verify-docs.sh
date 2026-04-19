#!/usr/bin/env bash
# 校验立项与工程文档完整性（CI 与本地均可运行）
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

REQUIRED=(
  "README.md"
  "CONTRIBUTING.md"
  "SECURITY.md"
  "docs/README.md"
  "docs/PROJECT_INITIATION.md"
  "docs/SRS.md"
  "docs/SEARCH_API_V1.md"
  "docs/ARCHITECTURE.md"
  "docs/DETAILED_DESIGN.md"
  "docs/AGENT_MARKDOWN_PIPELINE.md"
  "docs/ROUTER_FRAMEWORK_EVALUATION.md"
  "docs/SDLC.md"
  "docs/TESTING_AND_TDD.md"
  "docs/GITHUB_WORKFLOW.md"
  "docs/DOCUMENTATION_STANDARDS.md"
)

missing=0
for f in "${REQUIRED[@]}"; do
  if [[ ! -f "$f" ]]; then
    echo "ERROR: missing required file: $f" >&2
    missing=1
  fi
done

if [[ "$missing" -ne 0 ]]; then
  exit 1
fi

echo "OK: all required documentation files present."
