# 文档索引与维护说明

本目录包含 **立项级**、**需求可追溯**、**API 契约**、**架构**、**TDD/测试** 与 **GitHub 协作** 的完整文档。修改对外行为时，须同步更新 **API 文档** 与 **SRS**，并在 PR 中列出 `REQ-*`。

---

## 文档列表

### 立项与需求

| 文档 | 说明 |
|------|------|
| [PROJECT_INITIATION.md](./PROJECT_INITIATION.md) | **立项说明书**：背景、目标、范围、里程碑、风险、成功标准 |
| [SRS.md](./SRS.md) | **软件需求规格说明**：`REQ-F-*` / `REQ-NF-*` 与验收要点 |
| [SEARCH_API_V1.md](./SEARCH_API_V1.md) | **API 与产品说明 v1**：端点、鉴权、Markdown、截断、错误码 |

### 设计与工程

| 文档 | 说明 |
|------|------|
| [ARCHITECTURE.md](./ARCHITECTURE.md) | **概要设计**：逻辑组件、数据流、**技术选型**、境内驻留 |
| [DETAILED_DESIGN.md](./DETAILED_DESIGN.md) | **详细设计**：模块分层、内部接口、算法、数据与配置、测试要点 |
| [AGENT_MARKDOWN_PIPELINE.md](./AGENT_MARKDOWN_PIPELINE.md) | **合成层**：索引 / Tavily JSON / 未来 HTML 抓取 → Agent Markdown 的边界与演进顺序 |
| [ROUTER_FRAMEWORK_EVALUATION.md](./ROUTER_FRAMEWORK_EVALUATION.md) | **chi / gin / echo** 评估与选型记录 |
| [TESTING_AND_TDD.md](./TESTING_AND_TDD.md) | **TDD** 流程、测试分层、覆盖率、CI 策略 |
| [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md) | 分支保护、Actions、Secrets、Dependabot |
| [DEPLOY_LINUX_VM.md](./DEPLOY_LINUX_VM.md) | **Linux 虚机**：Docker Compose 部署、防火墙与安全组 |
| [ADMIN_UI_DESIGN.md](./ADMIN_UI_DESIGN.md) | **Admin 管理界面**（外网同端口）：方案、安全、路由与分期 |
| [ADMIN_ENABLE.md](./ADMIN_ENABLE.md) | **Admin 启用条件**：环境变量、Compose 注入与自检 |
| [SDLC.md](./SDLC.md) | 生命周期、发布与回滚、合并门禁 |
| [DOCUMENTATION_STANDARDS.md](./DOCUMENTATION_STANDARDS.md) | 文档维护规则与版本表 |

---

## 仓库根目录相关文件

| 文件 | 说明 |
|------|------|
| [../README.md](../README.md) | 项目总览与文档入口 |
| [../CONTRIBUTING.md](../CONTRIBUTING.md) | 贡献与 **TDD** 流程 |
| [../SECURITY.md](../SECURITY.md) | 漏洞报告方式 |

---

## CI 校验

`scripts/verify-docs.sh` 会检查上表中 **核心文档文件** 是否存在，防止误删。新增「必选」文档时，请同步更新该脚本中的列表。

---

## 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.3 | 2026-04-18 | 增加详细设计文档索引 |
| 0.4 | 2026-04-18 | 索引 [AGENT_MARKDOWN_PIPELINE.md](./AGENT_MARKDOWN_PIPELINE.md) |
| 0.5 | 2026-04-18 | AGENT_MARKDOWN_PIPELINE §2.1 补充可选 URL 摘录与 SRS REQ-F-012 |
| 0.2 | 2026-04-18 | 立项化扩展：SRS、TDD、GitHub、文档规范索引 |
| 0.1 | 2026-04-18 | 初稿 |
