# 软件需求规格说明（SRS）

本文档对 **AI 联网搜索服务** MVP 需求给出 **唯一标识**，便于测试用例、实现与变更的 **可追溯性**。详细 HTTP 字段与示例以 [SEARCH_API_V1.md](./SEARCH_API_V1.md) 为准。

| 文档版本 | 0.7 |
|----------|-----|
| 修订日期 | 2026-04-18 |

---

## 1. 引言

### 1.1 目的

定义可验证的需求，支撑 **TDD**：每个 `REQ-*` 至少对应一类测试（单元、集成或契约），或在「不适用」栏说明理由。

### 1.2 参考文档

- [SEARCH_API_V1.md](./SEARCH_API_V1.md)
- [ARCHITECTURE.md](./ARCHITECTURE.md)
- [DETAILED_DESIGN.md](./DETAILED_DESIGN.md)
- [PROJECT_INITIATION.md](./PROJECT_INITIATION.md)

---

## 2. 功能需求

| ID | 描述 | 验收要点（测试需覆盖） |
|----|------|------------------------|
| REQ-F-001 | 系统提供 `POST /v1/search` | 路径与方法正确；非法方法返回 **405** 或项目约定的一致行为 |
| REQ-F-002 | 请求体为 JSON，且包含必填字段 `query`（字符串） | 缺字段或类型错误 → **400**；错误体可机读 |
| REQ-F-003 | 使用 `Authorization: Bearer <api_key>` 鉴权 | 缺失/无效 Key → **401** |
| REQ-F-004 | 成功时返回 **200**，`Content-Type` 含 `text/markdown` 与 `charset=utf-8` | 契约测试断言头与 body 为 Markdown 文本 |
| REQ-F-005 | 响应正文为 **一段 Markdown**，不采用传统「搜索结果列表 UI」形态 | 快照或语义断言（由《响应风格指南》细化时加强） |
| REQ-F-006 | 当自建索引对当前查询满足「足够」条件时，优先使用索引生成响应 | 集成测试：索引预置数据时 **不** 调用（或调用次数为 0）外部 Provider 双倍 |
| REQ-F-007 | 「足够」由可配置阈值判定，至少包含：命中条数、文本长度、相似度 | 单元测试覆盖边界值；配置变更可改变判定结果 |
| REQ-F-008 | 当索引不足时，回落到第三方搜索 API 与/或受控网页获取，再整理为 Markdown | 集成测试：索引空时走 Provider/抓取双倍 |
| REQ-F-009 | 当处理接近时限或输出预算时，允许 **硬截断** 正文，仍返回 **200** 与 **部分 Markdown** | 单元或集成测试：模拟短超时/小预算，断言 200 且正文含截断说明（措辞见风格指南） |
| REQ-F-010 | 若在时限内 **无法** 产生任何可用 Markdown 片段，返回 **504** 或 **503**（与 API 文档一致），错误体非 Markdown 成功路径 | 集成测试注入「全失败」依赖 |
| REQ-F-011 | 对请求实施限流，防止滥用与高并发爬站 | 超限 → **429**；测试可用假时钟或独立限流配置 |
| REQ-F-012 | （可选，服务端配置）在第三方搜索回落后，可对结果中的 **`https` 落地 URL** 做 **受控、限量** 的 GET，将 **纯文本摘录** 并入 Markdown；**默认关闭** | 配置关闭时行为与未实现该能力时一致；开启时集成/单元测试：`FetchPlainText` 与 `EnrichProviderItemsWithFetch` 双倍；见 [AGENT_MARKDOWN_PIPELINE.md](./AGENT_MARKDOWN_PIPELINE.md) §2.1 |

---

## 3. 非功能需求

| ID | 描述 | 验收要点 |
|----|------|----------|
| REQ-NF-001 | **无状态**：单请求不依赖服务端会话状态 | 负载均衡下任意实例行为一致；测试中不依赖内存会话 |
| REQ-NF-002 | **数据驻留**：自建存储与自建索引位于 **境内** 指定区域 | 部署文档与 IaC 变量可审计；架构文档一致 |
| REQ-NF-003 | **隐私披露**：产品文档说明经境外第三方搜索 API 可能产生的数据出境 | 对外隐私/数据处理说明存在且与架构一致 |
| REQ-NF-004 | **安全**：日志不输出完整 API Key；密钥来自环境变量或密钥管理 | 静态检查或审查清单 |
| REQ-NF-005 | **可维护性**：核心逻辑具备自动化测试；主分支 CI 必须通过 | 见 [TESTING_AND_TDD.md](./TESTING_AND_TDD.md) |
| REQ-NF-006 | **可观测性（建议）**：请求 ID、结构化日志、基础指标；**`GET /health`** 探活（无鉴权、不限流） | `internal/handler/health_test.go`；`cmd/server` 注册路由 |

---

## 4. 需求追踪矩阵（维护方式）

| REQ ID | 测试套件（示例占位） | 实现模块（示例占位） | 备注 |
|--------|----------------------|----------------------|------|
| REQ-F-001～005 | `internal/handler/search_test.go` | `internal/handler`, `cmd/server` | 契约与 Problem Details |
| REQ-F-006～008 | `internal/orchestrator/service_test.go`, `internal/adapter/tavily/*_test.go`, `internal/adapter/opensearch/*_test.go`, `internal/adapter/baidubrowser/*_test.go`, `internal/adapter/bingbrowser/*_test.go`, `internal/adapter/googlebrowser/*_test.go` | `internal/orchestrator`, `internal/adapter/tavily`, `internal/adapter/stubprovider`, `internal/adapter/opensearch`, `internal/adapter/baidubrowser`, `internal/adapter/bingbrowser`, `internal/adapter/googlebrowser` | 索引优先 / 回落；多 Provider 注册表（`providers` / `SEARCH_DEFAULT_PROVIDERS`）；浏览器：百度 / Bing / Google（`*_BROWSER_ENABLED`）与 Tavily 并存 |
| REQ-F-012 | `internal/orchestrator/service_test.go`, `internal/adapter/fetch/*_test.go`, `internal/domain/enrich_test.go` | `internal/adapter/fetch`, `internal/domain`, `internal/orchestrator` | 默认关；`PROVIDER_FETCH_RESULT_URLS=true` 时启用 |
| REQ-F-007 | `internal/domain/*_test.go` | `internal/domain` | Enough / 归一化 |
| REQ-F-009 | `internal/domain/truncate_test.go` | `internal/domain` | 截断 |
| REQ-F-011 | 集成/手测 | `internal/middleware/ratelimit.go` | 限流中间件 |
| REQ-NF-002 | `integration` + 部署检查 | `infra` | 待部署阶段 |

**规则**：合并影响某 `REQ-*` 的 PR 时，须更新本表或测试清单，并在 PR 描述中列出涉及的需求 ID。

---

## 5. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿，与 API v1 对齐 |
| 0.2 | 2026-04-18 | 更新追踪矩阵：对应 `internal/` 包与测试路径 |
| 0.3 | 2026-04-18 | 回落路径增加 Tavily 适配器与测试引用 |
| 0.4 | 2026-04-18 | REQ-F-012：可选 Provider 结果 URL 的受控 HTTPS 摘录 |
| 0.5 | 2026-04-18 | 追踪矩阵：OpenSearch 读路径 `internal/adapter/opensearch` |
| 0.6 | 2026-04-18 | 追踪矩阵：`internal/adapter/baidubrowser`（百度 SERP，优先于 Tavily） |
| 0.7 | 2026-04-18 | 追踪矩阵：Bing/Google 浏览器、`GET /health`；REQ-NF-006 探活 |
