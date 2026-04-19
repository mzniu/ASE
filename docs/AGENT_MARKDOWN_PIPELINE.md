# 从检索结果到 Agent Markdown（合成层说明）

本文档补齐 **「人类向 SERP / 网页」→「面向大模型的 Markdown」** 的设计边界，与 [DETAILED_DESIGN.md](./DETAILED_DESIGN.md) 中的 `MarkdownComposer` / 编排器一致。

| 文档版本 | 0.3 |
|----------|-----|
| 修订日期 | 2026-04-18 |

---

## 1. 我们到底在转换什么？

| 来源 | 原始形态 | 当前实现（MVP） | 后续可演进 |
|------|----------|-----------------|------------|
| **自建索引（OpenSearch）** | 已清洗的正文片段 / 字段 | `AgentMarkdownFromIndexHits`（内部 `MarkdownFromIndex`）：按「要点 + 片段」组织，**非**十条链接列表 | 引入去重、段落合并、引用块 |
| **Tavily 等 API** | **结构化 JSON**（`title` / `url` / `content`） | `AgentMarkdownFromProviderItems`（内部 `MarkdownFromProvider`）：映射为短摘要型 Markdown；**不解析 HTML SERP** | 可选 `include_raw_content`、多段合并、LLM 摘要 |
| **百度桌面 SERP（无头 Chrome）** | 渲染后 DOM | `internal/adapter/baidubrowser`：`chromedp` + `goquery` → `ProviderItem`（`BAIDU_BROWSER_ENABLED`） | 验证码/DOM 变更；低 QPS；合规自负 |
| **落地页 HTML（通用）** | 整页 HTML | **`PageFetcher`** 摘录路径（如 Tavily 结果 URL） | Readability 类抽取；与百度 SERP Provider 不同层 |

**结论**：  
- **没有遗漏「JSON → Markdown」的独立重型模块**：Tavily 返回的已是 **机器友好字段**，当前层做的是 **版式与语义编排**（标题层级、避免门户列表 UI）。  
- **百度 SERP**：通过 **chromedp** 得到 DOM 后解析为 `ProviderItem`，再进入同一合成入口。  
- **通用「任意网页 → 正文」**：仍以 **`PageFetcher` + 抽取** 演进（见 §3）。

---

## 2. 合成层在代码中的位置

- **规范化**：`port.Hit` / `port.ProviderItem`（URL、标题、摘要/正文片段）。  
- **面向 Agent 的 Markdown**：`internal/domain` 内  
  - `AgentMarkdownFromIndexHits`（索引命中）  
  - `AgentMarkdownFromProviderItems`（Provider 回落）  
  二者是 **REQ-F-005** 的显式入口；内部目前调用模板化实现，便于替换为更强策略而不改编排器签名。  
- **截断**：`TruncateToRunes`（REQ-F-009）。

编排器（`internal/orchestrator`）只做 **顺序**：索引 → 足够性 → 否则 Provider → **（可选）落地页抓取** → **合成 → 截断**，不在此写死 HTML 解析逻辑。

### 2.1 可选：对 Provider 结果中的 URL 做 HTTPS 摘录（已实现开关）

| 项 | 说明 |
|----|------|
| **动机** | Tavily 等 API 的 `content`/`snippet` 可能偏短；对结果中的 **`https` 落地页** 再拉取正文，可提升 Agent 可用信息量。 |
| **代价** | 同步延迟上升；目标站可能反爬、付费墙或返回非 HTML；须遵守 robots/站点 ToS 与出境披露。 |
| **默认** | **关闭**（`PROVIDER_FETCH_RESULT_URLS` 未设为真时），仅使用 API 返回字段。 |
| **开启后行为** | 在单请求 **Deadline** 内，按结果顺序对最多 **`PROVIDER_FETCH_MAX_URLS`** 个 **互异 `https://` URL** 顺序发起 GET；正文经 **轻量 HTML→纯文本**（非 Readability 级 DOM 分析），按 URL 与 `ProviderItem` 对齐后 **拼入「页面摘录」** 段落，再进入 `AgentMarkdownFromProviderItems`。 |
| **实现位置** | `internal/port.PageFetcher`、`internal/adapter/fetch`（`Noop` / `Simple`）、`internal/domain.EnrichProviderItemsWithFetch`。 |

**建议**：生产环境先 **小 `PROVIDER_FETCH_MAX_URLS`（如 1～2）** 与合理 **`FETCH_PER_URL_TIMEOUT_MS`**，并监控 P95 与上游错误率。

### 2.2 百度搜索（`BAIDU_BROWSER_ENABLED`）

| 项 | 说明 |
|----|------|
| **作用** | 作为 **SearchProvider** 实现（优先于 Tavily），在索引不足时用无头 Chrome 打开 `https://www.baidu.com/s?wd=…`，解析有机结果标题/链接/摘要。 |
| **依赖** | 本机 **Chrome/Chromium**；可选 **`CHROME_EXEC_PATH`**。 |
| **风险** | 验证码、反爬、DOM 变更；须遵守百度规则与低频使用。 |

---

## 3. 建议的后续实现顺序（与合规一致）

1. **`PageFetcher`**：受控并发、单域限速、仅允许 `https`、遵守 robots/ToS。  
2. **正文抽取**：HTML → 主文本（或轻量 Markdown），产出与 `port.ProviderItem` 可统一的中间结构。  
3. **合成策略（可选其一或组合）**  
   - **规则**：去重 URL、按分数排序、拼成「引用式」段落（仍避免蓝色链接列表 UI）。  
   - **模型**：对「中间文本」做一次短摘要（需注意延迟、成本、出境合规）。  
4. **索引写回**：把规范化文档写入 OpenSearch，减少对外部 API 的依赖。

---

## 4. 与隐私/合规的交叉

- 经 **Tavily** 或 **境外抓取** 的路径须在隐私说明中披露。  
- 「爬虫」若指高并发整站抓取，与本项目 **产品非目标** 冲突；仅允许 **受控、可追溯** 的 Fetch。

---

## 5. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿：明确 JSON 与 HTML 两条路径及 MVP 边界 |
| 0.2 | 2026-04-18 | §2.1：可选对 Provider 结果 URL 的 HTTPS 摘录与配置键 |
| 0.3 | 2026-04-18 | §2.2：百度 SERP（chromedp + goquery）；表格区分 SERP 与落地页抓取 |
