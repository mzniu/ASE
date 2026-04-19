# 详细设计说明书

本文档在 [ARCHITECTURE.md](./ARCHITECTURE.md) 概要设计基础上，给出 **模块划分、内部接口、核心算法、数据与配置、错误与可观测性** 等实现级约定。对外 HTTP 契约仍以 [SEARCH_API_V1.md](./SEARCH_API_V1.md) 为准；需求追溯以 [SRS.md](./SRS.md) 的 `REQ-*` 为准。

| 文档版本 | 0.6 |
|----------|-----|
| 修订日期 | 2026-04-18 |

---

## 1. 引言

### 1.1 目的

为编码与 TDD 提供 **单一事实来源**：模块边界清晰、依赖可注入、行为可测。

### 1.2 范围

涵盖 **`POST /v1/search`** 同步路径及内部依赖；不含未立项的计费、审计、多租户控制台。

### 1.3 参考文档

| 文档 | 用途 |
|------|------|
| [SEARCH_API_V1.md](./SEARCH_API_V1.md) | 对外 API 与截断/错误语义 |
| [ARCHITECTURE.md](./ARCHITECTURE.md) | 逻辑组件与技术选型决议 |
| [SRS.md](./SRS.md) | 功能/非功能需求 ID |
| [TESTING_AND_TDD.md](./TESTING_AND_TDD.md) | 测试策略 |

### 1.4 命名约定

- 下文 **接口名** 为逻辑名；具体语言中的 `interface` / `Protocol` / `ABC` 名称可与项目惯例对齐，但 **职责不得拆分漂移**。  
- **配置键** 使用 `SCREAMING_SNAKE_CASE` 示例，实现时可映射为环境变量或配置文件。

---

## 2. 设计目标与约束

| 目标/约束 | 说明 |
|-----------|------|
| REQ-NF-001 无状态 | 不在进程内保存跨请求会话；所需上下文均来自请求与配置 |
| 同步响应 | 单请求在 **Deadline** 内完成；部分结果可截断后 **200** |
| 可替换实现 | 索引、Provider、Fetcher 均通过接口注入，便于 Fake/Stub |
| 密钥 | API Key、第三方 Key 仅来自配置/密钥管理，**禁止** 写入日志正文 |

### 2.1 已决议：Go + OpenSearch

| 项 | 约定 |
|----|------|
| 语言 | **Go**（≥ 1.22，以仓库 `go.mod` 为准） |
| HTTP | **chi v5**（[ROUTER_FRAMEWORK_EVALUATION.md](./ROUTER_FRAMEWORK_EVALUATION.md)） |
| 自建索引 | **OpenSearch 2.x**（境内集群）；HTTP 客户端使用 **`opensearch-project/opensearch-go`** |
| 相似度 | MVP **无向量字段**；`EnoughPolicy` 的「相似度」取 OpenSearch 返回的 **`_score` 经批次内 min-max 归一化** 到 \([0,1]\)（单次查询内可比即可） |
| 持久化 | **无 PostgreSQL**；文档与可检索正文以索引为主（见 §6.3） |

**建议目录结构**（可按实现微调，须在 README 说明）：

```
cmd/
  server/          # main，注入依赖、监听 HTTP
internal/
  handler/         # POST /v1/search，Problem Details
  orchestrator/    # SearchOrchestrator
  domain/          # EnoughPolicy, TruncationPolicy, MarkdownComposer（纯逻辑优先）
  port/            # 接口：IndexRepository, SearchProvider, PageFetcher
  adapter/
    opensearch/    # IndexRepository 实现
    provider/      # Baidu / Bing / Tavily 等
    fetch/         # PageFetcher
  config/
```

测试：`internal/...` 与 `*_test.go` 同包或 `package xxx_test`；契约测试可放在 `test/integration/`（目录名随仓库约定）。

---

## 3. 逻辑分层与模块

```
┌─────────────────────────────────────────────────────────┐
│  Transport（HTTP）                                       │
│  路由、鉴权中间件、限流、请求解析、响应头、Problem Details │
└───────────────────────────┬─────────────────────────────┘
                            │
┌───────────────────────────▼─────────────────────────────┐
│  Application：SearchOrchestrator                         │
│  编排：索引优先 → 回落 → 合成 → 截断                     │
└───────────┬───────────────────────────────┬─────────────┘
            │                               │
┌───────────▼──────────┐         ┌──────────▼────────────┐
│ Domain / Policy      │         │ Ports（接口）           │
│ EnoughPolicy         │         │ IndexRepository       │
│ RateLimitPolicy      │         │ SearchProvider          │
│ TruncationPolicy     │         │ PageFetcher             │
│ MarkdownComposer     │         │ (可选) IndexWriter      │
└──────────────────────┘         └──────────┬──────────────┘
                                            │
                                 ┌──────────▼──────────────┐
                                 │ Adapters（基础设施实现） │
                                 │ OpenSearch / HTTP / …   │
                                 └─────────────────────────┘
```

| 模块 | 职责 | 主要 REQ |
|------|------|----------|
| Transport | `POST /v1/search` 唯一入口；401/400/429；成功 `text/markdown` | REQ-F-001～005、011 |
| SearchOrchestrator | 串联检索、回落、合成、截断；传播 `request_id` 与 Deadline | REQ-F-006～010 |
| EnoughPolicy | 根据命中条数、文本长度、相似度与配置判断是否「足够」 | REQ-F-007 |
| MarkdownComposer | 将检索/抓取结果转为 **非门户列表式** Markdown | REQ-F-005 |
| TruncationPolicy | 在时限或字符预算内硬截断，附 **截断提示** 文案常量 | REQ-F-009 |
| IndexRepository | 自建索引读（及可选异步写） | REQ-F-006、008 |
| SearchProvider | 第三方搜索 API；可多实现 + 优先级/熔断 | REQ-F-008 |
| PageFetcher | 受控 HTTP 获取正文；并发与单域限速 | REQ-F-008、合规 |
| RateLimiter | 每 Key / 全局限流 | REQ-F-011 |

---

## 4. 核心流程

### 4.1 主路径（伪代码）

```
function HandleSearch(ctx, HTTPRequest):
    request_id ← 生成或从网关注入
    若 RateLimiter 拒绝: return 429 + Problem Details

    api_key ← 自 Authorization Bearer 解析
    若无效: return 401

    body ← 解析 JSON
    若缺 query 或类型非法: return 400

    deadline ← 自配置的单请求最长处理时间
    ctx ← 带 deadline 的上下文

    hits ← IndexRepository.Search(ctx, query)
    若 EnoughPolicy.Satisfied(hits, query):
        md ← MarkdownComposer.FromIndexHits(hits)
        md ← TruncationPolicy.Apply(md, ctx)
        return 200, text/markdown, md

    // 回落
    ext ← SearchProvider.Search(ctx, query)   // 可链式多 Provider
    pages ← PageFetcher.FetchBatch(ctx, ext.URLs, 限流参数)
    md ← MarkdownComposer.FromExternal(ext, pages)
    可选: 异步 IndexWriter.Enqueue(规范化文档)  // 不在同步路径阻塞

    若 md 为空或仅空白:
        return 504 或 503 + Problem Details  // REQ-F-010

    md ← TruncationPolicy.Apply(md, ctx)
    return 200, text/markdown, md
```

### 4.2 「足够」判定（EnoughPolicy）

输入：`query`、`hits`（来自索引的候选列表，每项含 `score`、`text_len`、可选 `vector_similarity` 等）。

建议 **可配置布尔表达式**（实现可选用简单规则引擎或代码分支），默认逻辑示例：

- `len(hits) >= MIN_HIT_COUNT` **且**  
- `sum(h.text_len for h in hits) >= MIN_TOTAL_TEXT_LEN` **且**  
- 若启用向量：`max(h.similarity for h in hits) >= MIN_SIMILARITY`

任一条件不满足 → **不足**，走回落。阈值来自配置（见 §8）。

**单元测试**：边界值（刚好等于阈值、差 1）；REQ-F-007。

### 4.3 截断（TruncationPolicy）

- 输入：完整 Markdown 字符串、`deadline` 剩余时间、**最大输出字符数** `MAX_RESPONSE_CHARS`（或 UTF-8 字节数，须文档化取哪一种）。  
- 若超时或超长：**截断到预算内**，并在末尾追加 **固定模板** 短句（例如：「（以下内容因长度限制已截断）」—— 最终措辞以《响应风格指南》为准）。  
- **仍有任意非空可用片段** → **200**；**完全为空** → 按 §4.4 失败路径。

### 4.4 完全失败（与 API 对齐）

- 回落链路全部失败、或合成结果为空、且在 deadline 内无法得到任何非空 Markdown → **504**（网关/上游超时语义）或 **503**（依赖不可用）；**Content-Type** 为 Problem Details，**非** `text/markdown` 成功体。

### 4.5 「人类向 SERP / HTML」→ Agent Markdown（实现边界）

本项目的 **REQ-F-005** 目标是 **面向大模型的可读 Markdown**，而非复刻门户「十条蓝色链接」式 UI。实现上分两条路径，边界如下。

| 来源 | 原始形态 | MVP 实现 | 说明 |
|------|----------|----------|------|
| 自建索引 | 已清洗的正文片段等 | `AgentMarkdownFromIndexHits` → 内部 `MarkdownFromIndex` | 索引侧已承担清洗；合成层做版式与语义编排 |
| Tavily 等 API | **结构化 JSON**（`title` / `url` / `content` 等） | `AgentMarkdownFromProviderItems` → 内部 `MarkdownFromProvider` | **不**把第三方响应当作 HTML SERP 解析；字段映射后即进入合成器 |
| Provider 结果中的 **https 落地 URL**（可选） | 整页 HTML（GET） | **`PROVIDER_FETCH_RESULT_URLS=true`** 时：`port.PageFetcher`（`internal/adapter/fetch.Simple`）顺序拉取最多 N 个 URL，**轻量 HTML→纯文本**，`EnrichProviderItemsWithFetch` 拼入摘要后再合成 | REQ-F-012；仅 **https**；默认 **关闭**；非 Readability 级抽取，见 [AGENT_MARKDOWN_PIPELINE.md](./AGENT_MARKDOWN_PIPELINE.md) §2.1 |
| 浏览器爬虫 / 整页 HTML | DOM、人类搜索结果页 | 与上一行共用 **Fetch + 轻量摘录**；更广义的 **Readability / 多页融合** | 仍可作为演进：更强抽取、每主机限速、`FromExternal` 全文融合 |

编排器（§3 `SearchOrchestrator`）当前顺序为：索引检索 → 足够性 → 否则 **SearchProvider** → **（可选）按配置对结果 URL 做 `FetchPlainText` 与摘录合并** → **合成（`AgentMarkdown*`）** → 截断。

---

## 5. 内部接口（Ports）

以下为 **逻辑契约**；具体方法签名随语言调整。

### 5.1 IndexRepository

| 方法 | 说明 |
|------|------|
| `Search(ctx, query) -> []Hit` | 经 OpenSearch **search** API；`Hit` 含 `id`、`body`/`snippet`、`score`（`_score`）、`similarity`（由 `_score` 归一化填充，见 §6.3） |
| （可选）`Upsert(ctx, documents)` | 索引写回；可与异步队列配合，**不在**同步 `search` 路径阻塞 |

### 5.2 SearchProvider

| 方法 | 说明 |
|------|------|
| `Search(ctx, query) -> ProviderResult` | `ProviderResult` 含 `[]ResultItem`（`url`、`title`、`snippet`）与 `error` |

实现：`internal/adapter/baidubrowser`（百度桌面 SERP + chromedp）、`TavilyProvider` 等；**测试**使用 `FakeProvider` 或 stub 固定返回。

### 5.3 PageFetcher

Go 实现：`internal/port.PageFetcher`，方法 **`FetchPlainText(ctx, urls, limit) -> []FetchedPage`**（按顺序取前 `limit` 个互异 **https** URL，顺序 GET，尊重 `ctx` 时限）。适配器 **`internal/adapter/fetch`**：`Noop`（默认）、`Simple`（`net/http` + 轻量 `HTMLToPlain`）。

| 方法（逻辑名） | 说明 |
|------|------|
| `FetchPlainText` | 受控 GET；单 URL 超时见 `FETCH_PER_URL_TIMEOUT_MS` |
| （演进）`FetchBatch(ctx, urls, opts)` | 若将来引入全局并发/每主机限速，可扩展 opts 或包装 `Simple` |

### 5.4 MarkdownComposer

逻辑上的 **Markdown 合成器**；在 Go 实现中，MVP 对应 `internal/domain` 的 **`AgentMarkdownFromIndexHits` / `AgentMarkdownFromProviderItems`**（REQ-F-005 的稳定入口），内部再调用模板化 `MarkdownFromIndex` / `MarkdownFromProvider`，便于日后替换策略而不改编排器。详见 [AGENT_MARKDOWN_PIPELINE.md](./AGENT_MARKDOWN_PIPELINE.md)。

| 方法 | 说明 |
|------|------|
| `FromIndexHits(hits) -> string` | 非固定模板，但禁止「十条蓝色链接列表」式编排 |
| `FromExternal(providerResult, pages) -> string` | 融合摘要与正文，可含文末参考区；依赖 **PageFetcher** 与 HTML→文本抽取，**当前仓库未实现** |

---

## 6. 数据设计（概念模型）

### 6.1 索引文档（示例字段）

实际字段名以选型引擎映射为准；下列用于评审与测试夹具。

| 字段 | 类型 | 说明 |
|------|------|------|
| `doc_id` | string | 稳定 ID（URL 哈希或 UUID） |
| `url` | string | 来源 |
| `title` | string | 可选 |
| `body_text` | string | 清洗后正文（或分段存储） |
| `indexed_at` | datetime | 可选 |
| `embedding` | float[] | 若采用向量检索 |

### 6.2 Hit（内存结构）

| 字段 | 说明 |
|------|------|
| `score` | OpenSearch 返回的原始 **`_score`**（BM25 等） |
| `text_len` | 参与合成的文本长度（UTF-8 字符或字节数，实现须统一并在配置中说明） |
| `similarity` | MVP：**非向量**；由本批次 `score` **min-max 归一化** 到 \([0,1]\)，供 `EnoughPolicy` 与 `MIN_SIMILARITY` 比较 |

### 6.3 OpenSearch 索引与查询约定

- **索引名称**：配置项 `OPENSEARCH_INDEX`（如 `ase_documents`）；环境可分 `dev` / `prod` 前缀。  
- **连接**：`OPENSEARCH_URLS`（逗号分隔节点 URL）、`OPENSEARCH_USER` / `OPENSEARCH_PASSWORD` 或等价签名方式（以托管方要求为准）；**TLS** 与证书校验在生产强制开启。  
- **映射（概念）**：至少包含 `url`（keyword）、`title`（text）、`body_text`（text，主要检索字段）、`indexed_at`（date，可选）。分析器：中文场景选用 **ik_smart / ik_max_word** 等须在 **索引创建时** 固定并文档化（依赖 OpenSearch 插件是否可用）。  
- **查询**：MVP 使用 **`multi_match` 或 `simple_query_string`** 对 `title`+`body_text` 检索；`size` 与 `min_score` 可配置，用于控制候选条数与噪声。  
- **归一化**：对当次响应的 hits 收集 `_score`，令 `similarity_i = (score_i - min) / (max - min)`；若 `max == min`，则置 `similarity_i = 1`（或按产品约定常数），并在单元测试中锁定。

---

## 7. HTTP 层细节

| 项 | 约定 |
|----|------|
| 路径 | 仅注册 `POST /v1/search`；其余返回 **404** 或 **405**（须全仓库一致） |
| 成功 `Content-Type` | `text/markdown; charset=utf-8` |
| `request_id` | 响应头建议：`X-Request-Id`（可选，便于排障） |
| Problem Details | `Content-Type: application/problem+json`；`type` URI 建议可区分 401/400/429/503/504 |

---

## 8. 配置项（示例清单）

实现时集中在一处加载（如 `config.yaml` + 环境变量覆盖）。**默认值**在首次实现后由压测填入 [SEARCH_API_V1.md](./SEARCH_API_V1.md) 待定表。

| 配置键 | 含义 | 备注 |
|--------|------|------|
| `AUTH_VALID_API_KEYS` 或密钥存储后端 | 校验 Bearer | MVP 可用静态列表；后续换 DB |
| `REQUEST_DEADLINE_MS` | 单请求服务端最长处理时间 | 与截断策略联动 |
| `MAX_RESPONSE_CHARS` | Markdown 最大字符数 | REQ-F-009 |
| `MIN_HIT_COUNT` / `MIN_TOTAL_TEXT_LEN` / `MIN_SIMILARITY` | EnoughPolicy | REQ-F-007 |
| `RATE_LIMIT_PER_KEY_RPS` / `RATE_LIMIT_GLOBAL_RPS` | 限流 | REQ-F-011 |
| `FETCH_MAX_CONCURRENCY` / `FETCH_PER_HOST_INTERVAL_MS` | 防爬 | 合规 |
| `SEARCH_PROVIDER_CHAIN` | Provider 顺序 | 如 `baidu,bing` |
| `OPENSEARCH_URLS` | OpenSearch 节点 | 逗号分隔 |
| `OPENSEARCH_INDEX` | 索引名 | 与 §6.3 一致 |
| `OPENSEARCH_USER` / `OPENSEARCH_PASSWORD` | 基本认证或其它机制 | 按集群要求 |
| `OPENSEARCH_SEARCH_SIZE` | `/_search` 的 `size`（候选条数上限） | 默认 `10`；实现见 `internal/adapter/opensearch` |
| 各 `*_API_KEY` / `*_ENDPOINT` | 第三方搜索 | 环境变量注入 |
| `PROVIDER_FETCH_RESULT_URLS` | 是否对 Provider 结果中的 **https** URL 再抓取摘录 | 默认 `false`；`true` / `1` / `yes` / `on` 为开启（REQ-F-012） |
| `PROVIDER_FETCH_MAX_URLS` | 每请求最多抓取的结果条数（顺序、去重后） | 默认 `2` |
| `FETCH_PER_URL_TIMEOUT_MS` | 单次 GET 超时（毫秒） | 默认 `8000`；与 `REQUEST_DEADLINE_MS` 叠加 |

---

## 9. 错误与日志

| HTTP | 场景 | 日志级别 |
|------|------|----------|
| 400 | JSON/query 非法 | WARN，附 `request_id` |
| 401 | Key 无效 | WARN，**不**记录完整 Key |
| 429 | 限流 | WARN |
| 503 | Provider 全不可用 | ERROR，附错误类型聚合 |
| 504 | 无可用片段且超时 | WARN/ERROR |
| 500 | 未捕获异常 | ERROR，附 stack 或错误码 |

结构化字段建议：`timestamp`, `level`, `request_id`, `path`, `status`, `latency_ms`, `error_code`（内部枚举）。

---

## 10. 安全设计要点

- **REQ-NF-004**：日志中对 `Authorization` 仅记录「已提供/未提供」或 Key 的 **哈希前缀**。  
- **输入**：`query` 最大长度校验，防极端 payload（与 SRS REQ-F-002 一致）。  
- **出站**：HTTPS 调第三方；TLS 版本与证书校验按运行时默认安全策略。

---

## 11. 测试设计要点（与 SRS 对齐）

| REQ ID | 测试类型 | 要点 |
|--------|----------|------|
| REQ-F-003～004 | 契约 | 401/200、`Content-Type` |
| REQ-F-007 | 单元 | EnoughPolicy 边界 |
| REQ-F-006/008 | 集成 | Fake Index / Fake Provider 调用次数 |
| REQ-F-009 | 单元/集成 | 小 `MAX_RESPONSE_CHARS` 或短 deadline → 200 + 截断提示 |
| REQ-F-010 | 集成 | Provider+Fetcher 全失败 → 非 200 Markdown |
| REQ-F-011 | 单元/集成 | 超限 429 |
| REQ-F-012 | 单元/集成 | 关闭时无额外 GET；开启时编排器调用 `PageFetcher` 且合成含摘录 |

---

## 12. 与概要设计、技术选型的衔接

- **已决议**：**Go + OpenSearch**，客户端 **`opensearch-go`**；HTTP 路由 **chi v5**。  
- **相似度**：MVP **不使用向量字段**；`Hit.similarity` 由 **`_score` 批次内归一化** 得到（§6.2、§6.3）。若后续启用 OpenSearch kNN 或外接向量库，须更新索引映射、`EnoughPolicy` 与本节。  
- **OpenSearch 小版本**：以境内托管实例为准；升级前在预发跑集成测试与索引兼容性检查。

---

## 13. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿：模块、流程、接口、数据与配置 |
| 0.2 | 2026-04-18 | 对齐 **Go + OpenSearch**：目录结构、§6.3 索引与 `_score` 归一化、配置项 |
| 0.3 | 2026-04-18 | 固定 **chi v5**；与已初始化 `cmd/server` 对齐 |
| 0.4 | 2026-04-18 | §4.5：JSON 与 HTML 两条合成路径边界；§5.4 对齐 `AgentMarkdown*` 实现入口 |
| 0.5 | 2026-04-18 | REQ-F-012：可选 Provider 结果 URL 抓取；§5.3 `PageFetcher` 与 §8 配置对齐代码 |
| 0.6 | 2026-04-18 | §8：`OPENSEARCH_SEARCH_SIZE`；OpenSearch 读路径实现于 `internal/adapter/opensearch` |
