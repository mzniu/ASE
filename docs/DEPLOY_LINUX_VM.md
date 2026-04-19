# 在 Linux 虚机上部署 ASE（Docker Compose）

面向已有一台可 SSH 的 Linux 主机（如 Azure VM、阿里云 ECS），用本仓库的 **`docker-compose.yml`** 一次拉起 **OpenSearch + ASE**。

## 1. 准备

- 虚机建议 **≥ 2 vCPU、≥ 4 GB RAM**（OpenSearch JVM 约 256 MB，留余量给系统与其它进程）。
- 开放云安全组 / NSG **入站**：**22**（SSH）、**18080**（ASE HTTP，若仅内网访问可限制来源 IP）。
- 本机安装 **Git**、**Docker Engine** 与 **Docker Compose V2**（`docker compose version` 可用）。

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

---

| 文档 | 说明 |
|------|------|
| [SEARCH_API_V1.md](./SEARCH_API_V1.md) | API 契约 |
| [README.md](../README.md) | 环境变量与 OpenSearch 排障 |
