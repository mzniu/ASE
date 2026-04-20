# AI 联网搜索服务（面向 Agent）

为开发者提供 **无状态 REST API**：使用用户级 API Key 调用 **`POST /v1/search`**，返回经整理的 **Markdown**，供 Agent / AI 工具作为联网检索上下文使用。

源码仓库：**[github.com/mzniu/ASE](https://github.com/mzniu/ASE)**（私有仓库需相应权限）。开发与合并请求须遵循 **TDD（测试驱动开发）** 与 [CONTRIBUTING.md](./CONTRIBUTING.md) 中的约定。

---

## 文档入口

| 文档 | 说明 |
|------|------|
| [立项说明书](docs/PROJECT_INITIATION.md) | 背景、目标、范围、里程碑、风险与成功标准 |
| [软件需求规格说明（SRS）](docs/SRS.md) | 可追溯需求 ID，与测试/实现的追踪关系 |
| [API 与产品说明 v1](docs/SEARCH_API_V1.md) | 对外契约：端点、鉴权、响应、截断与错误语义 |
| **项目主页** | 启动服务后访问根路径 **`GET /`**（嵌入 `internal/webcontent/index.html`）：项目介绍 + **通用 Agent Skill** 配置、API 摘要、可复制安装指令；机器可读 **`GET /api/info`**（JSON）。Skill 文件亦可从本服务 **`GET /skills/ase-search-api/*`** 下载（与 API 同源，无需访问 GitHub） |
| [测试与 TDD 规范](docs/TESTING_AND_TDD.md) | 红-绿-重构、测试分层、CI、合并门禁 |
| [GitHub 工作流](docs/GITHUB_WORKFLOW.md) | 分支保护、Actions、密钥、Dependabot |
| [架构概要](docs/ARCHITECTURE.md) | 概要设计：逻辑组件、**技术选型**、数据流、境内驻留 |
| [详细设计](docs/DETAILED_DESIGN.md) | 模块分层、内部接口、算法、配置与测试要点 |
| [Agent Markdown 合成层](docs/AGENT_MARKDOWN_PIPELINE.md) | 检索结果 → Markdown：Tavily/索引与「HTML 抓取」路径的边界 |
| [SDLC](docs/SDLC.md) | 生命周期、发布与回滚检查清单 |
| [文档编写规范](docs/DOCUMENTATION_STANDARDS.md) | 文档维护责任与版本表规则 |
| [文档索引](docs/README.md) | 全量文档列表与维护责任 |
| [Linux 虚机部署（Docker）](docs/DEPLOY_LINUX_VM.md) | SSH 登录后在目标机克隆仓库、`docker compose` 与防火墙要点 |
| [安全策略](SECURITY.md) | 漏洞报告渠道（发布前替换占位仓库名） |

---

## 技术栈（已决议）

- **Go**（版本以根目录 `go.mod` 为准；CI 使用 `go-version-file` 对齐）+ **OpenSearch 2.x**（境内集群）；HTTP 路由 **chi v5**。详见 [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) §2.4、[docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §2.1、[docs/ROUTER_FRAMEWORK_EVALUATION.md](./docs/ROUTER_FRAMEWORK_EVALUATION.md)。

### OpenSearch（索引优先，REQ-F-006）

同时配置 **`OPENSEARCH_URLS`**（逗号分隔，如 `https://search.example:9200`）与 **`OPENSEARCH_INDEX`** 后，服务使用 [opensearch-go](https://github.com/opensearch-project/opensearch-go) 对索引字段 **`title`**、**`body_text`** 做 `multi_match` 检索；可选 **`OPENSEARCH_USER`** / **`OPENSEARCH_PASSWORD`**（HTTP Basic）、**`OPENSEARCH_SEARCH_SIZE`**（默认 10）。未配置两项时回退为内存空索引（与此前行为一致）。映射与查询约定见 [docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §6.3。

### 端点一览（v1）

| 方法 | 路径 | 说明 |
|------|------|------|
| `POST` | `/v1/search` | 主业务：JSON 查询 → Markdown（需 Bearer） |
| `POST` | `/v1/documents` | 可选：向索引写入 `id` / `title` / `body_text`；未配置 OpenSearch 时 **501**（详见 [SEARCH_API_V1.md](./docs/SEARCH_API_V1.md)） |
| `GET` | `/` | **HTML 项目主页**（嵌入静态页），**无鉴权**、**不限流** |
| `GET` | `/api/info` | JSON 服务发现（服务名与常用路径），**无鉴权**、**不限流** |
| `GET` | `/health` | 探活 JSON，**无鉴权**、**不限流** |
| `GET` | `/metrics` | **Prometheus** 文本指标（如 `ase_search_orchestration_total`），**无鉴权**、**不限流** |

## 快速开始

**要求**：安装 **Go**（与 `go.mod` 中 `go` 行一致或更新），并确认 `go version` 可用。

若 **`go mod download` 访问 proxy.golang.org 超时**（常见于国内网络），可为当前用户设置一次：

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
```

```bash
# 当前 go.mod 模块名为 github.com/example/ase；若 fork 到自己的仓库可改为例如：
# go mod edit -module=github.com/mzniu/ASE && go mod tidy

make check          # gofmt 检查 + go vet + go test（与 CI 一致；CI 中 Go 版本与 go.mod 对齐）
# 或：go fmt ./... && go vet ./... && go test ./...
go run ./cmd/server
# 另开终端：需 DEV_API_KEY 时
# export DEV_API_KEY=dev-only
# curl -sS -X POST http://127.0.0.1:18080/v1/search -H "Content-Type: application/json" -H "Authorization: Bearer dev-only" -d '{"query":"hello"}'
```

环境变量见 [.env.example](./.env.example)。在仓库根目录放置 **`.env`**（勿提交）时，启动时会 **优先用 `.env` 覆盖同名环境变量**（`godotenv.Overload`），避免 Windows「用户/系统环境变量」里残留的 **`TAVILY_API_KEY=`** 或旧值导致 **`.env` 不生效**（若仅用 `Load`，已存在的键不会被文件覆盖）。请从仓库根目录执行 `go run ./cmd/server`，且 `.env` 建议保存为 **UTF-8 无 BOM**，以免首行键名异常。**生产环境须配置真实 API Key 校验策略**，勿依赖「未设置 `DEV_API_KEY` 则任意 Bearer 通过」的开发行为。

### 探活与指标

**`GET /health`** 返回 `{"status":"ok"}`（JSON），**不需要** Bearer，且 **不计入** 业务限流，便于 K8s / 负载均衡探针。

**`GET /metrics`** 供 Prometheus 抓取（搜索编排结果计数等），同样 **无鉴权**、**不在** `/v1` 限流组内。

### 可选：OpenSearch 集成测试

已配置 **`OPENSEARCH_URLS`** 等环境变量时，可在本地执行 **`bash scripts/run-integration.sh`**（`go test -tags=integration`，见 `internal/adapter/opensearch/integration_test.go`）。未配置时会 **跳过** 相关用例。

### 多搜索引擎（浏览器 + Tavily）

在请求 JSON 中指定 **`providers`**（如 `["baidu"]`、`["bing","google"]`、`["tavily"]`），或通过环境变量 **`SEARCH_DEFAULT_PROVIDERS`**（逗号分隔）设置默认列表。未配置时按 **baidu → bing → google → tavily** 取第一个已启用的引擎。注册名包括 **`stub`**（测试用）。

### 百度搜索（无头浏览器）

设置 **`BAIDU_BROWSER_ENABLED=true`** 且本机已安装 **Chrome / Chromium** 时，索引不足会通过 [chromedp](https://github.com/chromedp/chromedp) 打开百度桌面 SERP 并解析 `#content_left` 中的结果（见 `internal/adapter/baidubrowser`）。可选 **`CHROME_EXEC_PATH`**、**`BAIDU_BROWSER_MAX_RESULTS`**。**Bing / Google** 同理： **`BING_BROWSER_ENABLED`**、**`GOOGLE_BROWSER_ENABLED`**（见 `.env.example`）。请控制频率，并自行评估验证码与合规。

### Tavily（联网回落）

未启用百度浏览器且设置 **`TAVILY_API_KEY`** 时，索引不足会调用 [Tavily Search API](https://docs.tavily.com/documentation/api-reference/endpoint/search)（`POST /search`，`Authorization: Bearer …`，默认 **`TAVILY_MAX_RESULTS=10`**，并请求 **`include_raw_content`** 以尽量增加正文）。否则使用内置 **stub** 回落。查询经境外 Tavily 时须符合你的隐私/合规披露。

### 可选：对结果 URL 再抓取摘录（REQ-F-012）

默认 **关闭**。设置 **`PROVIDER_FETCH_RESULT_URLS=true`**（及可选的 **`PROVIDER_FETCH_MAX_URLS`**、**`FETCH_PER_URL_TIMEOUT_MS`**、**`FETCH_CONCURRENCY`**）后，在回落路径上对结果中的 **`http`/`https` URL** 做并发 GET，经 Readability 与 HTML→**Markdown** 写入响应中的 **「## 正文」**。详见 [docs/AGENT_MARKDOWN_PIPELINE.md](./docs/AGENT_MARKDOWN_PIPELINE.md)；对目标站点的抓取须自行评估 ToS 与合规。

### Docker 与本地 OpenSearch

- **Compose**：请使用 **Docker Compose V2**（命令为 **`docker compose`**，空格）。勿使用旧版 Python **`docker-compose` 1.x**；在新版 Docker Engine 上可能报错 **`KeyError: 'ContainerConfig'`**，见 [docs/DEPLOY_LINUX_VM.md](./docs/DEPLOY_LINUX_VM.md) 中「§9. 故障：KeyError: 'ContainerConfig'」。
- **Admin**：配置 `ADMIN_USERNAME`、`ADMIN_PASSWORD_BCRYPT`（或开发用 `ADMIN_PASSWORD`）、`ADMIN_SESSION_SECRET` 后，浏览器访问 **`/admin/`**（脱敏配置与 OpenSearch 索引只读）。见 [docs/ADMIN_UI_DESIGN.md](./docs/ADMIN_UI_DESIGN.md)。
- **镜像构建**：默认 **`Dockerfile`** 基于 **Debian slim + Chromium**，供 baidu/bing/google 等无头浏览器 Provider；更小、无浏览器的运行时见 **`Dockerfile.distroless`**（适合仅 stub/Tavily）。
- **一次编排（API + OpenSearch）**：在仓库根目录执行 **`docker compose up --build -d`**。会启动 **`opensearch`**（9200）与 **`ase`**（默认把宿主机 **18080** 映射到容器内 18080）；`ase` 已默认 **`shm_size: 2gb`**、**`CHROME_EXEC_PATH=/usr/bin/chromium`**，并默认启用 **baidu/bing/google** 三个浏览器 Provider；**`TAVILY_API_KEY`** 在 `.env` 中设置后即注册 **tavily**。未在请求 JSON 里写 **`providers`** 时，默认仅走 **`SEARCH_DEFAULT_PROVIDERS`（默认 `baidu`）**；**`DEV_API_KEY=dev-only`** 可在 `.env` 覆盖。
- **仅索引节点（旧用法）**：**`docker compose up -d opensearch`**，宿主机上 **`OPENSEARCH_URLS=http://localhost:9200`** 连接。索引映射与查询约定见 [docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §6.3。

编排启动后可探活并试搜（与 compose 中默认 Key 一致）：

```bash
curl -sS http://127.0.0.1:18080/health
curl -sS -X POST http://127.0.0.1:18080/v1/search -H "Content-Type: application/json" -H "Authorization: Bearer dev-only" -d '{"query":"hello"}'
```

#### OpenSearch 容器立刻退出（exit 1）时

1. **看日志**：`docker logs ase-opensearch-1`（容器名以 `docker compose ps` 为准）。  
   - 若出现 **`OPENSEARCH_INITIAL_ADMIN_PASSWORD`**：OpenSearch 2.12+ 要求该变量（仓库内 `docker-compose.yml` 已内置开发用密码）；请 **`git pull`** 后重新 `docker compose up`。  
   - 若出现 **`vm.max_map_count`** 或内存/bootstrap 相关报错，见下文。  
2. **提高 vm.max_map_count**（Docker Desktop 使用 WSL2 时，在 **PowerShell** 执行，发行版名可用 `wsl -l -v` 查看）：

```powershell
wsl -d docker-desktop -- sysctl -w vm.max_map_count=262144
```

若提示找不到 `docker-desktop`，可尝试 `docker-desktop-data` 或在 WSL 发行版内执行同一 `sysctl` 命令。可写入 `%USERPROFILE%\.wslconfig` 持久化（需查当前 Windows 文档）。  
3. **清空损坏的数据卷后重试**：`docker compose down -v`，再 `docker compose up --build -d`（**会删除** OpenSearch 卷内数据）。  
4. 仓库内 `docker-compose.yml` 已为 OpenSearch 配置 **`cap_add: IPC_LOCK`** 与 **`ulimits.memlock`**；若仍失败，请在 **Docker Desktop → Settings → Resources** 提高内存/磁盘配额。

---

## 许可证

本项目采用 **[MIT License](./LICENSE)**（Copyright © 2026 Mingzhu Niu）。使用或分发时请保留许可证全文及版权声明。
