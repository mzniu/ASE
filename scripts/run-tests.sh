#!/usr/bin/env bash
# 统一测试入口：与本地、CI 行为一致。实现落地后在此固定命令。
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

# 注意：不要在此调用 `make test`，否则与根目录 Makefile 的 `test` 目标（若指向本脚本）形成死循环。

if [[ -f go.mod ]]; then
  bash "${ROOT}/scripts/check-go.sh"
  exit 0
fi

if [[ -f package.json ]] && grep -q '"test"' package.json; then
  echo "Running: npm test"
  if [[ -f package-lock.json ]] || [[ -f npm-shrinkwrap.json ]]; then
    npm ci
  else
    npm install
  fi
  npm test
  exit 0
fi

if [[ -f pyproject.toml ]] || [[ -f pytest.ini ]] || [[ -f setup.cfg ]]; then
  echo "Running: pytest"
  if command -v pytest >/dev/null 2>&1; then
    pytest
  else
    pip install -q pytest
    pytest
  fi
  exit 0
fi

echo "::notice title=Tests::No automated test runner detected yet. Add a Makefile target 'test' or standard project files (go.mod, package.json with test script, pytest) when implementation starts."
exit 0
