#!/usr/bin/env bash
# 在 ASE 仓库根目录的上一级执行时，请先 cd 到仓库根再运行；或任意位置：
#   bash scripts/deploy/restart-ase.sh
#
# 行为：git 拉取 → 先构建 ase 镜像（旧容器继续服务，中断时间短）→ 再 up --force-recreate 轮换容器。
# OpenSearch 不重建、默认保持运行（docker compose --no-deps）。
#
# 依赖：git、Docker Compose v2（docker compose）或 docker-compose。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

docker_compose() {
	if docker compose version >/dev/null 2>&1; then
		docker compose "$@"
	elif command -v docker-compose >/dev/null 2>&1; then
		docker-compose "$@"
	else
		echo "未检测到 Docker Compose。"
		exit 1
	fi
}

if ! command -v docker >/dev/null 2>&1 || ! docker info >/dev/null 2>&1; then
	echo "请先确保 Docker 可用且当前用户可访问守护进程。"
	exit 1
fi

if [[ ! -f docker-compose.yml ]]; then
	echo "未在仓库根目录找到 docker-compose.yml（当前：$ROOT）"
	exit 1
fi

if [[ -d .git ]]; then
	echo ">>> git pull（当前分支）"
	git pull --ff-only
else
	echo "提示：当前目录不是 git 仓库，跳过 git pull。"
fi

echo ">>> 构建 ase 镜像（构建期间旧容器仍运行）"
if [[ "${1:-}" == "--no-cache" ]]; then
	docker_compose build --no-cache ase
else
	docker_compose build ase
fi

echo ">>> 轮换 ase 容器（--force-recreate；中断仅发生在停旧启新）"
docker_compose up -d --no-deps --force-recreate ase

echo ""
docker_compose ps
echo ""
echo "探活："
curl -sS -m 5 "http://127.0.0.1:${ASE_HOST_PORT:-18080}/health" && echo "" || echo "（若失败请检查端口映射、ASE_HOST_PORT 与防火墙）"
