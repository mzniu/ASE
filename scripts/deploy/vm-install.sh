#!/usr/bin/env bash
# 在 Linux 虚机上执行（已 SSH 登录后）：于仓库根目录运行，拉起 OpenSearch + ASE。
# 用法：bash scripts/deploy/vm-install.sh
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$ROOT"

if ! command -v docker >/dev/null 2>&1; then
	echo "未检测到 docker，请先安装 Docker Engine：https://docs.docker.com/engine/install/"
	exit 1
fi

# Compose V2：docker compose（需 docker-compose-plugin）；旧环境可能是 docker-compose 独立命令
docker_compose() {
	if docker compose version >/dev/null 2>&1; then
		docker compose "$@"
	elif command -v docker-compose >/dev/null 2>&1; then
		docker-compose "$@"
	else
		echo "未检测到 Docker Compose。"
		echo "Ubuntu/Debian 推荐：sudo apt-get update && sudo apt-get install -y docker-compose-plugin"
		echo "然后确认：docker compose version"
		exit 1
	fi
}

if [[ ! -f .env ]]; then
	echo "提示：未找到 .env。可直接使用 compose 内默认变量启动；若需自定义请 cp .env.example .env 后编辑。"
fi

echo "构建并启动（Docker Compose）..."
docker_compose up --build -d

echo ""
docker_compose ps
echo ""
echo "本机探活："
curl -sS -m 5 "http://127.0.0.1:18080/health" && echo "" || echo "（若失败请检查 HTTP_ADDR、防火墙与云安全组）"
