# 软件开发生命周期与工程规范

适用于本仓库 **AI 联网搜索服务** 的研发与交付，与 [PROJECT_INITIATION.md](./PROJECT_INITIATION.md)、[SRS.md](./SRS.md)、[SEARCH_API_V1.md](./SEARCH_API_V1.md)、[ARCHITECTURE.md](./ARCHITECTURE.md)、[TESTING_AND_TDD.md](./TESTING_AND_TDD.md)、[GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md) 配套使用。

---

## 1. 阶段与产出物（门禁）

| 阶段 | 目标 | 主要产出物 | 进入下一阶段的最低条件 |
|------|------|------------|------------------------|
| 立项与范围 | 明确做/不做 | [PROJECT_INITIATION.md](./PROJECT_INITIATION.md) | 干系人对 MVP 达成一致 |
| 需求 | 可测试、可追溯 | [SRS.md](./SRS.md) 中的 `REQ-*`；契约细节见 [SEARCH_API_V1.md](./SEARCH_API_V1.md) | 无未决的契约级歧义（开放项记入 API 文档「待定」表） |
| 设计 | 可实现、可部署 | [ARCHITECTURE.md](./ARCHITECTURE.md)（含技术选型 §2.4）、[DETAILED_DESIGN.md](./DETAILED_DESIGN.md) | 概要/详细设计与 API、SRS 一致；技术决议表已填或显式「待决议」 |
| 实现 | 可运行的服务 | 代码、`.env.example`、密钥不入库 | **TDD**：新行为先有测试；通过 CI（见 §4） |
| 测试 | 风险可见 | [TESTING_AND_TDD.md](./TESTING_AND_TDD.md)；CI 报告 | 无阻塞级缺陷；`REQ-*` 有对应测试或已说明例外 |
| 发布 | 可回滚的上线 | 变更说明、版本号、回滚步骤 | 发布检查清单完成（见 §6） |
| 运维与变更 | 稳定与演进 | 监控与告警、事件记录；变更走评审 | 线上问题可追溯到版本与配置 |

---

## 2. 版本与兼容性

- **API 版本**：路径前缀 `/v1/`；**破坏性变更**须新 major（如 `/v2/`）或明确弃用期（本项目初期以 MVP 为主，变更优先反映在文档版本表）。
- **文档版本**：`SEARCH_API_V1.md` 文末「文档修订」表须随行为变化更新。
- **实现版本**：建议使用 **语义化版本**（SemVer）标记可部署制品；与 API 文档的「契约变更」在发布说明中对应。

---

## 3. 分支与代码评审

- **主分支**（如 `main`）：始终可发布；仅通过合并请求（Pull/Merge Request）合入。
- **功能分支**：`feature/<简短说明>`；**缺陷修复**：`fix/<简短说明>`。
- **合并要求**（最低）：
  - 至少 **一名** 其他成员 Code Review（单人项目时可自审 [`.github/PULL_REQUEST_TEMPLATE.md`](../.github/PULL_REQUEST_TEMPLATE.md) 清单）。
  - CI **必须通过**（文档校验 + `scripts/run-tests.sh`，见 [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md)）。
  - 涉及 **API 或行为** 的改动必须 **同步更新** `docs/SEARCH_API_V1.md`、`docs/SRS.md`（及受影响的 `docs/ARCHITECTURE.md`）。

---

## 4. 测试与质量门禁（TDD）

细则见 [TESTING_AND_TDD.md](./TESTING_AND_TDD.md)。

| 层级 | 内容 | 建议门禁 |
|------|------|----------|
| 单元测试 | 纯函数、判定逻辑（如「足够」阈值封装）、Markdown 拼装与截断 | CI 必跑；覆盖率阈值实现稳定后写入工作流 |
| 集成测试 | 带测试双倍的索引、假第三方 API | 合并前或单独 CI job |
| 契约/冒烟 | `POST /v1/search` 的 200 Markdown、401、400 等 | 与 SRS `REQ-F-*` 对齐 |

- **TDD**：默认 **红 → 绿 → 重构**；PR 模板中须勾选测试与文档项。
- **静态分析**：按技术栈启用 Linter/Formatter；依赖漏洞由 [Dependabot](../.github/dependabot.yml) 跟踪。
- **密钥**：API Key、云凭证 **禁止** 提交；使用环境变量或 GitHub Secrets，见 [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md)。

---

## 5. 安全与合规检查点

- **鉴权**：Bearer Key 校验失败返回 **401**；不在日志中输出完整 Key。
- **限流**：防爬与防滥用；与产品「禁止高并发爬站」一致。
- **驻留**：生产索引与存储在 **境内**；跨境第三方处理在隐私文档披露（见架构文档）。
- **依赖第三方搜索**：遵守各供应商 ToS；不在文档中承诺超出供应商能力的行为。

---

## 6. 发布与回滚

**发布前检查清单（示例）**

- [ ] `docs/SEARCH_API_V1.md` 与当前行为一致  
- [ ] 环境变量与密钥已在目标环境配置  
- [ ] 数据库/索引迁移（若有）已演练  
- [ ] 监控与告警（可用性、错误率、延迟）可用  
- [ ] 回滚方式明确（上一版本镜像/制品 + 配置）

**回滚**：优先 **恢复上一已知良好制品**；若涉及索引结构破坏级变更，须有备份恢复预案。

---

## 7. 缺陷与变更管理

- **缺陷**：记录复现步骤、期望/实际、环境、版本；修复后补充回归用例。
- **契约变更**：先改文档并评审，再改实现；或同时提交但须在 MR 中 **显式说明**。

---

## 8. 文档维护责任

| 变更类型 | 须更新的文档 |
|----------|----------------|
| 范围、里程碑、风险 | `PROJECT_INITIATION.md` |
| `REQ-*` 增删或验收标准 | `SRS.md` |
| 请求/响应语义、状态码、截断规则 | `SEARCH_API_V1.md` |
| 概要架构、技术选型决议、组件与驻留 | `ARCHITECTURE.md` |
| 模块、内部接口、算法、配置项 | `DETAILED_DESIGN.md` |
| 测试策略、TDD 门禁 | `TESTING_AND_TDD.md` |
| GitHub 工作流、分支规则 | `GITHUB_WORKFLOW.md` |
| 文档编写规则 | `DOCUMENTATION_STANDARDS.md` |
| 流程本身（分支策略、发布清单） | `SDLC.md` |

---

## 9. 文档修订

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿：阶段门禁、分支、测试、发布与安全检查点 |
| 0.2 | 2026-04-18 | 对齐立项/SRS/TDD/GitHub 文档与 CI 门禁 |
| 0.3 | 2026-04-18 | 设计阶段对齐概要/详细设计文档 |
