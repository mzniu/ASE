# 文档编写与维护规范

保证「文档充分」且 **与代码/测试同步**，避免口头约定漂移。

---

## 1. 文档体系

| 类型 | 路径 | 维护时机 |
|------|------|----------|
| 立项 | [PROJECT_INITIATION.md](./PROJECT_INITIATION.md) | 范围或里程碑变化 |
| 需求（可追溯） | [SRS.md](./SRS.md) | 增删 `REQ-*` 或验收标准变化 |
| API 契约 | [SEARCH_API_V1.md](./SEARCH_API_V1.md) | 任何对外行为变化 |
| 概要架构与选型 | [ARCHITECTURE.md](./ARCHITECTURE.md) | 组件、数据流、驻留、§2.4 技术决议 |
| 详细设计 | [DETAILED_DESIGN.md](./DETAILED_DESIGN.md) | 模块、内部接口、算法、配置项变化 |
| TDD / 测试 | [TESTING_AND_TDD.md](./TESTING_AND_TDD.md) | 测试策略、CI 门禁变化 |
| GitHub | [GITHUB_WORKFLOW.md](./GITHUB_WORKFLOW.md) | 分支规则、工作流文件变化 |
| SDLC | [SDLC.md](./SDLC.md) | 发布流程、角色分工变化 |

---

## 2. 与变更的绑定规则

1. **行为变更**：须同时更新 **SEARCH_API_V1.md**（若对外）与 **SRS.md**（对应 `REQ-*`），并在 PR 描述中列出需求 ID。
2. **仅文档**：可在同一 PR 中修正笔误；若修正改变语义，须按「行为变更」处理。
3. **版本表**：`SEARCH_API_V1.md`、`PROJECT_INITIATION.md`、`SRS.md` 等文末 **修订记录** 须追加一行（版本号或日期 + 简述）。

---

## 3. 风格

- 使用 **中文** 叙述需求与流程；HTTP 方法、Header、路径、代码标识符保持 **英文**。
- 表格优先于长段落列举；避免重复粘贴大段代码——用引用指向 `SEARCH_API_V1.md`。
- Mermaid 图仅在渲染环境支持时使用；关键信息在正文中 **文字可复述**。

---

## 4. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿 |
| 0.2 | 2026-04-18 | 增加概要/详细设计文档条目 |
