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

if ! docker info >/dev/null 2>&1; then
	echo "无法连接 Docker 守护进程（常见原因：当前用户无权访问 /var/run/docker.sock）。"
	echo ""
	echo "推荐：把用户加入 docker 组后重新登录 SSH，再运行本脚本："
	echo "  sudo usermod -aG docker \"\$USER\""
	echo "  newgrp docker"
	echo ""
	echo "或临时用 root 执行："
	echo "  sudo bash scripts/deploy/vm-install.sh"
	exit 1
fi

# Compose V2：docker compose（需 docker-compose-plugin）。勿依赖旧版 Python docker-compose 1.x（Docker 24+ 下易 KeyError: ContainerConfig）。
docker_compose() {
	if docker compose version >/dev/null 2>&1; then
		docker compose "$@"
	else
		echo "未检测到 Docker Compose V2（docker compose）。"
		echo "Ubuntu/Debian：sudo apt-get update && sudo apt-get install -y docker-compose-plugin && docker compose version"
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
