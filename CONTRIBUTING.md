# 贡献指南

感谢参与 **AI 联网搜索服务** 的开发。本仓库以 **GitHub** 协作，并以 **TDD** 为默认开发方式。

---

## 1. 开始之前

请先阅读：

- [docs/PROJECT_INITIATION.md](docs/PROJECT_INITIATION.md) — 范围与目标  
- [docs/SEARCH_API_V1.md](docs/SEARCH_API_V1.md) — 对外 API 契约  
- [docs/SRS.md](docs/SRS.md) — 可追溯需求 `REQ-*`  
- [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) — 概要设计与 **技术选型**（§2.4）  
- [docs/DETAILED_DESIGN.md](docs/DETAILED_DESIGN.md) — **详细设计**（模块、内部接口、配置）  
- [docs/TESTING_AND_TDD.md](docs/TESTING_AND_TDD.md) — **TDD 与测试分层**  
- [docs/GITHUB_WORKFLOW.md](docs/GITHUB_WORKFLOW.md) — 分支保护与 CI  

---

## 2. 开发流程（TDD）

1. 从 `main` 拉取最新，创建 `feature/*` 或 `fix/*` 分支。  
2. **先写测试**（红）→ **最小实现**（绿）→ **重构**。  
3. 本地运行与 CI 相同的测试入口（见下节）。  
4. 提交 PR，填写模板中的 **测试** 与 **文档** 勾选项。  
5. 确保 **CI 全部通过** 后再请求合并。

涉及对外行为时：同步更新 `docs/SEARCH_API_V1.md` 与 `docs/SRS.md`，并在 PR 描述中列出 **REQ-* ID**。  
涉及内部模块、接口或配置项时：同步更新 `docs/DETAILED_DESIGN.md`；若变更技术栈或基础设施，更新 `docs/ARCHITECTURE.md` §2.4 决议表。

---

## 3. 本地命令

| 命令 | 说明 |
|------|------|
| `bash scripts/verify-docs.sh` | 校验关键文档存在（与 CI 一致） |
| `bash scripts/run-tests.sh` | 当前仓库为 Go：`gofmt` 检查、`go vet`、`go test ./...` |
| `make fmt` | `go fmt ./...`（改写文件） |
| `make check` | 与 CI 相同：`scripts/check-go.sh` |

实现落地后，推荐提供 **Makefile** 且包含 `test` 目标，作为唯一标准入口。

### Go 开发约定

- **模块路径**：将 `go.mod` 第一行改为你的仓库，例如 `go mod edit -module=github.com/ORG/ase`。  
- **提交前**：执行 `make check`（或 `make fmt` 后再 `make check`），确保能通过 GitHub Actions。  
- **风格**：以 **`gofmt`** 为准（本仓库 `.editorconfig` 对 `*.go` 使用 **Tab 缩进**）。  
- **可选静态分析**：安装 [golangci-lint](https://golangci-lint.run/) 后执行 `golangci-lint run`（配置见 [.golangci.yml](./.golangci.yml)）。  
- **应用入口**：`cmd/server`；业务代码放在 **`internal/`**，避免被外部模块引用（Go 可见性惯例）。

---

## 4. Code Review 关注点

- 是否有 **失败路径** 测试（401、400、截断、504/503 等按 SRS）。  
- 是否引入 **密钥** 或 **生产 URL** 进仓库。  
- 文档与行为是否 **一致**。  

---

## 5. 行为准则

保持讨论就事论事、尊重不同观点。若后续采用开源行为准则，将在本文件链接。
