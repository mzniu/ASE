# AI 联网搜索服务（面向 Agent）

为开发者提供 **无状态 REST API**：使用用户级 API Key 调用 **`POST /v1/search`**，返回经整理的 **Markdown**，供 Agent / AI 工具作为联网检索上下文使用。

本仓库代码将托管于 **GitHub**；开发与合并请求须遵循 **TDD（测试驱动开发）** 与 [CONTRIBUTING.md](./CONTRIBUTING.md) 中的约定。

---

## 文档入口

| 文档 | 说明 |
|------|------|
| [立项说明书](docs/PROJECT_INITIATION.md) | 背景、目标、范围、里程碑、风险与成功标准 |
| [软件需求规格说明（SRS）](docs/SRS.md) | 可追溯需求 ID，与测试/实现的追踪关系 |
| [API 与产品说明 v1](docs/SEARCH_API_V1.md) | 对外契约：端点、鉴权、响应、截断与错误语义 |
| [测试与 TDD 规范](docs/TESTING_AND_TDD.md) | 红-绿-重构、测试分层、CI、合并门禁 |
| [GitHub 工作流](docs/GITHUB_WORKFLOW.md) | 分支保护、Actions、密钥、Dependabot |
| [架构概要](docs/ARCHITECTURE.md) | 概要设计：逻辑组件、**技术选型**、数据流、境内驻留 |
| [详细设计](docs/DETAILED_DESIGN.md) | 模块分层、内部接口、算法、配置与测试要点 |
| [Agent Markdown 合成层](docs/AGENT_MARKDOWN_PIPELINE.md) | 检索结果 → Markdown：Tavily/索引与「HTML 抓取」路径的边界 |
| [SDLC](docs/SDLC.md) | 生命周期、发布与回滚检查清单 |
| [文档编写规范](docs/DOCUMENTATION_STANDARDS.md) | 文档维护责任与版本表规则 |
| [文档索引](docs/README.md) | 全量文档列表与维护责任 |
| [安全策略](SECURITY.md) | 漏洞报告渠道（发布前替换占位仓库名） |

---

## 技术栈（已决议）

- **Go**（版本以根目录 `go.mod` 为准；CI 使用 `go-version-file` 对齐）+ **OpenSearch 2.x**（境内集群）；HTTP 路由 **chi v5**。详见 [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) §2.4、[docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §2.1、[docs/ROUTER_FRAMEWORK_EVALUATION.md](./docs/ROUTER_FRAMEWORK_EVALUATION.md)。

### OpenSearch（索引优先，REQ-F-006）

同时配置 **`OPENSEARCH_URLS`**（逗号分隔，如 `https://search.example:9200`）与 **`OPENSEARCH_INDEX`** 后，服务使用 [opensearch-go](https://github.com/opensearch-project/opensearch-go) 对索引字段 **`title`**、**`body_text`** 做 `multi_match` 检索；可选 **`OPENSEARCH_USER`** / **`OPENSEARCH_PASSWORD`**（HTTP Basic）、**`OPENSEARCH_SEARCH_SIZE`**（默认 10）。未配置两项时回退为内存空索引（与此前行为一致）。映射与查询约定见 [docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §6.3。配置 OpenSearch 后可用 **`POST /v1/documents`** 写入文档（见 [SEARCH_API_V1.md](./docs/SEARCH_API_V1.md)）；**`GET /metrics`** 暴露 Prometheus 指标（如 `ase_search_orchestration_total`）。

## 快速开始

**要求**：安装 **Go**（与 `go.mod` 中 `go` 行一致或更新），并确认 `go version` 可用。

若 **`go mod download` 访问 proxy.golang.org 超时**（常见于国内网络），可为当前用户设置一次：

```powershell
go env -w GOPROXY=https://goproxy.cn,direct
```

```bash
# 将模块路径换成你的 GitHub 仓库（示例）
# go mod edit -module=github.com/ORG/ase

make check          # gofmt 检查 + go vet + go test（与 CI 一致）
# 或：go fmt ./... && go vet ./... && go test ./...
go run ./cmd/server
# 另开终端：需 DEV_API_KEY 时
# export DEV_API_KEY=dev-only
# curl -sS -X POST http://127.0.0.1:18080/v1/search -H "Content-Type: application/json" -H "Authorization: Bearer dev-only" -d '{"query":"hello"}'
```

环境变量见 [.env.example](./.env.example)。在仓库根目录放置 **`.env`**（勿提交）时，启动时会 **优先用 `.env` 覆盖同名环境变量**（`godotenv.Overload`），避免 Windows「用户/系统环境变量」里残留的 **`TAVILY_API_KEY=`** 或旧值导致 **`.env` 不生效**（若仅用 `Load`，已存在的键不会被文件覆盖）。请从仓库根目录执行 `go run ./cmd/server`，且 `.env` 建议保存为 **UTF-8 无 BOM**，以免首行键名异常。**生产环境须配置真实 API Key 校验策略**，勿依赖「未设置 `DEV_API_KEY` 则任意 Bearer 通过」的开发行为。

### 探活

**`GET /health`** 返回 `{"status":"ok"}`（JSON），**不需要** Bearer，且 **不计入** 业务限流，便于 K8s / 负载均衡探针。

### 多搜索引擎（浏览器 + Tavily）

在请求 JSON 中指定 **`providers`**（如 `["baidu"]`、`["bing","google"]`、`["tavily"]`），或通过环境变量 **`SEARCH_DEFAULT_PROVIDERS`**（逗号分隔）设置默认列表。未配置时按 **baidu → bing → google → tavily** 取第一个已启用的引擎。注册名包括 **`stub`**（测试用）。

### 百度搜索（无头浏览器）

设置 **`BAIDU_BROWSER_ENABLED=true`** 且本机已安装 **Chrome / Chromium** 时，索引不足会通过 [chromedp](https://github.com/chromedp/chromedp) 打开百度桌面 SERP 并解析 `#content_left` 中的结果（见 `internal/adapter/baidubrowser`）。可选 **`CHROME_EXEC_PATH`**、**`BAIDU_BROWSER_MAX_RESULTS`**。**Bing / Google** 同理： **`BING_BROWSER_ENABLED`**、**`GOOGLE_BROWSER_ENABLED`**（见 `.env.example`）。请控制频率，并自行评估验证码与合规。

### Tavily（联网回落）

未启用百度浏览器且设置 **`TAVILY_API_KEY`** 时，索引不足会调用 [Tavily Search API](https://docs.tavily.com/documentation/api-reference/endpoint/search)（`POST /search`，`Authorization: Bearer …`，默认 **`TAVILY_MAX_RESULTS=10`**，并请求 **`include_raw_content`** 以尽量增加正文）。否则使用内置 **stub** 回落。查询经境外 Tavily 时须符合你的隐私/合规披露。

### 可选：对结果 URL 再抓取摘录（REQ-F-012）

默认 **关闭**。设置 **`PROVIDER_FETCH_RESULT_URLS=true`**（及可选的 **`PROVIDER_FETCH_MAX_URLS`**、**`FETCH_PER_URL_TIMEOUT_MS`**、**`FETCH_CONCURRENCY`**）后，在回落路径上对结果中的 **`http`/`https` URL** 做并发 GET，经 Readability 与 HTML→**Markdown** 写入响应中的 **「## 正文」**。详见 [docs/AGENT_MARKDOWN_PIPELINE.md](./docs/AGENT_MARKDOWN_PIPELINE.md)；对目标站点的抓取须自行评估 ToS 与合规。

### Docker 与本地 OpenSearch

- **镜像构建**：仓库根目录 **`Dockerfile`**（多阶段构建，产物为单二进制；需联网依赖时使用 **distroless base** 镜像以携带 CA）。
- **本地索引**：**`docker compose up -d opensearch`**（见 **`docker-compose.yml`**）可启动单节点 OpenSearch（**开发用**，已关闭安全插件）。配置示例：`OPENSEARCH_URLS=http://localhost:9200`、`OPENSEARCH_INDEX=ase_documents`。索引映射与查询约定见 [docs/DETAILED_DESIGN.md](./docs/DETAILED_DESIGN.md) §6.3。

---

## 许可证

待定（在仓库根目录添加 `LICENSE` 后更新本段）。
