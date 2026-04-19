# 在 Linux 虚机上部署 ASE（Docker Compose）

面向已有一台可 SSH 的 Linux 主机（如 Azure VM、阿里云 ECS），用本仓库的 **`docker-compose.yml`** 一次拉起 **OpenSearch + ASE**。

## 1. 准备

- 虚机建议 **≥ 2 vCPU、≥ 4 GB RAM**（OpenSearch JVM 约 256 MB，留余量给系统与其它进程）。
- 开放云安全组 / NSG **入站**：**22**（SSH）、**18080**（ASE HTTP，若仅内网访问可限制来源 IP）。
- 本机安装 **Git**、**Docker Engine** 与 **Docker Compose V2 插件**（**必须**能执行 `docker compose`，勿依赖旧版独立命令 `docker-compose`）：
  - **安装**：见下文 **§1.1**（若直接 `apt install docker-compose-plugin` 报 *Unable to locate package*，说明未配置 Docker 官方源，或改用 **手动安装插件二进制**）。
  - 若只装了 Python 版 **`docker-compose` 1.29.x**，在新版 Docker Engine（如 24+）上执行 `up` 可能报错 **`KeyError: 'ContainerConfig'`**，请改用 Compose V2 或卸载旧包后仅保留插件（见下文 §9）。
- 安装 Docker 后，**把登录用户加入 `docker` 组**，否则会出现 `Permission denied` 访问套接字：
  ```bash
  sudo usermod -aG docker "$USER"
  ```
  然后 **退出 SSH 再登录**，或执行 `newgrp docker`，再运行 `docker info` 确认无报错。

### 1.1 安装 Compose V2（`Unable to locate package docker-compose-plugin` 时）

发行版自带的 `apt` **不一定**包含 `docker-compose-plugin`，需任选其一。

**方式 A — 已用 Docker 官方仓库装过 `docker-ce` 的**：直接再装插件即可：

```bash
sudo apt-get update
sudo apt-get install -y docker-compose-plugin
docker compose version
```

**方式 B — 尚未添加 Docker 官方 APT 源**：按 Docker 文档为 **Ubuntu** 或 **Debian** 添加 `download.docker.com` 源后再安装 `docker-compose-plugin`（与安装 `docker-ce` 相同流程）。官方索引：<https://docs.docker.com/engine/install/>（选你的发行版，跟「Set up Docker’s apt repository」步骤）。

**方式 C — 仅手动安装 Compose 插件（不依赖 apt 包名）**：从 GitHub 发布页下载二进制到 Docker **CLI 插件目录**（需已安装 `docker` 命令）：

```bash
# 版本号可到 https://github.com/docker/compose/releases 取最新 v2.x
COMPOSE_VER=v2.29.7
ARCH=$(uname -m)
case "$ARCH" in x86_64) DARCH=x86_64 ;; aarch64|arm64) DARCH=aarch64 ;; *) echo "unsupported: $ARCH"; exit 1 ;; esac
sudo mkdir -p /usr/local/lib/docker/cli-plugins
sudo curl -fsSL "https://github.com/docker/compose/releases/download/${COMPOSE_VER}/docker-compose-linux-${DARCH}" \
  -o /usr/local/lib/docker/cli-plugins/docker-compose
sudo chmod +x /usr/local/lib/docker/cli-plugins/docker-compose
docker compose version
```

若 `docker compose version` 仍失败，可改为当前用户目录插件（再执行 `docker compose version`）：

```bash
mkdir -p ~/.docker/cli-plugins
cp /usr/local/lib/docker/cli-plugins/docker-compose ~/.docker/cli-plugins/ 2>/dev/null || true
# 或直接把上面 curl 目标改为 ~/.docker/cli-plugins/docker-compose 后 chmod +x
```

## 2. 获取代码

**私有仓库**需任选其一：

- HTTPS + Personal Access Token：  
  `git clone https://<TOKEN>@github.com/mzniu/ASE.git`
- 或配置 **SSH Deploy Key** 后：  
  `git clone git@github.com:mzniu/ASE.git`

```bash
cd /opt   # 或你的目录
sudo mkdir -p /opt/ase && sudo chown "$USER:$USER" /opt/ase
cd /opt/ase
git clone <你的仓库 URL> .
```

## 3. 环境变量

```bash
cp .env.example .env
nano .env   # 或 vim
```

至少确认：

- **`DEV_API_KEY`** 或 **`AUTH_VALID_API_KEYS`**（生产勿用弱口令）。
- 若使用本机 Compose 内的 OpenSearch：**`OPENSEARCH_URLS`**、**`OPENSEARCH_INDEX`** 在 compose 里已由环境注入容器，宿主机上的 `.env` 主要给 **直接 `go run`** 用；**仅 Docker 部署**时可在 compose 的 `ase.environment` 中覆盖，或保持 compose 默认。

Compose 内 ASE 已设置 `OPENSEARCH_URLS=http://opensearch:9200`，一般**无需**在 `.env` 再写宿主机 `localhost:9200`。

## 4. 内核参数（OpenSearch）

若 OpenSearch 容器反复退出，在 **宿主机**执行（需 root）：

```bash
sudo sysctl -w vm.max_map_count=262144
echo "vm.max_map_count=262144" | sudo tee /etc/sysctl.d/99-opensearch.conf
sudo sysctl --system
```

## 5. 防火墙（示例：ufw）

```bash
sudo ufw allow 22/tcp
sudo ufw allow 18080/tcp
sudo ufw status
```

## 6. 安装并启动

在仓库根目录：

```bash
bash scripts/deploy/vm-install.sh
```

若出现 **`unknown flag: --build`** 或 **`docker: 'compose' is not a docker command`**，说明未安装 Compose V2 插件，请按 **§1.1** 安装后再运行 `vm-install.sh`。

若出现 **`PermissionError: ... docker.sock`** 或 **`Permission denied`**（访问 Docker 套接字），见上文「把用户加入 docker 组」，或使用 **`sudo bash scripts/deploy/vm-install.sh`** 临时部署。

或手动：

```bash
docker compose up --build -d
docker compose ps
curl -sS http://127.0.0.1:18080/health
```

## 7. 对外访问

- 确认 **`HTTP_ADDR`**：默认 **`:18080`** 即监听所有接口；若只监听回环，外网无法访问。
- 浏览器：`http://<公网IP>:18080/`（主页）、`/health` 探活。

## 8. 更新版本

```bash
cd /opt/ase
git pull
docker compose up --build -d
```

## 9. 故障：`KeyError: 'ContainerConfig'`（旧版 `docker-compose`）

**现象**：使用 **`docker-compose`**（Compose **V1**，如 1.29.2）在 **`docker compose up`** / 重建容器时崩溃，栈中出现 `container.image_config['ContainerConfig']`。

**原因**：Compose V1 与较新的 Docker Engine 镜像元数据不兼容。

**处理**（任选其一，推荐前两条）：

1. 安装 **Compose V2 插件**（`apt` 找不到包时见 **§1.1**），之后一律使用 **`docker compose`**（中间有空格）：
   ```bash
   docker compose version
   ```
2. 在仓库根目录用 **`docker compose`** 重试（不要用 `docker-compose`）：
   ```bash
   docker compose up --build -d
   ```
3. 可选：移除旧的独立包，避免误用（包名因发行版而异，请先 `dpkg -l | grep -i compose` 再卸载）。

仓库内 **`scripts/deploy/vm-install.sh`**、**`scripts/deploy/restart-ase.sh`** 仅调用 **`docker compose`**，不再回退到 V1。

---

| 文档 | 说明 |
|------|------|
| [SEARCH_API_V1.md](./SEARCH_API_V1.md) | API 契约 |
| [README.md](../README.md) | 环境变量与 OpenSearch 排障 |
