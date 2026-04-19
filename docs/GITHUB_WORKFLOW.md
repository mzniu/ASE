# GitHub 工作流与仓库治理

代码托管于 **GitHub** 时，建议按下述方式配置，与 [SDLC.md](./SDLC.md)、[TESTING_AND_TDD.md](./TESTING_AND_TDD.md) 一致。

---

## 1. 分支模型

| 分支 | 用途 |
|------|------|
| `main` | 默认主分支；**受保护**；仅经 PR 合并 |
| `feature/*` | 新功能 |
| `fix/*` | 缺陷修复 |

可选：发布标签 `v*` 对应 [语义化版本](https://semver.org/lang/zh-CN/) 的发行版。

---

## 2. 分支保护规则（`main`）

在仓库 **Settings → Branches → Branch protection rule** 中为 `main` 启用（按需勾选）：

- **Require a pull request before merging**（需至少 1 个审批时由团队规模决定）。
- **Require status checks to pass before merging**：至少包含 CI 工作流中的 **verify-docs** 与 **test**（或合并后的单一工作流任务）。
- **Require conversation resolution before merging**（有审查意见时）。
- **Do not allow bypassing the above settings**（可选，管理员也遵守）。

单人维护时：至少保留 **CI 必须通过**，避免坏提交进主分支。

---

## 3. GitHub Actions

- 工作流文件： [.github/workflows/ci.yml](../.github/workflows/ci.yml)
- **Job 1**：文档完整性校验（`scripts/verify-docs.sh`）。
- **Job 2**：运行 `scripts/run-tests.sh`（按顺序尝试 Go / Node / Python 测试；无则提示尚未配置）。

实现语言确定后，保持 **CI 与本地** 均通过 `bash scripts/run-tests.sh` 或根目录 `make test`（`make test` 仅转调该脚本，**勿**在脚本内再调用 `make test`，以免递归）。

---

## 4. 密钥与配置

| 类型 | 存放位置 | 说明 |
|------|----------|------|
| 第三方搜索 API Key | GitHub **Secrets**（如 `BING_API_KEY`） | 仅用于 Actions 中的集成测试 **若** 使用沙箱；默认测试仍用双倍 |
| 部署凭据 | 环境级 Secrets / 云厂商 OIDC | 不在仓库明文 |
| 本地开发 | `.env`（**加入 `.gitignore`**） | 提供 `.env.example` 列出键名，无值 |

---

## 5. Dependabot

在 `.github/dependabot.yml` 中为所用生态启用依赖更新（示例：npm、gomod、pip），频率 **weekly** 即可；安全更新可配合 **Dependabot alerts**。

---

## 6. Issue / PR 模板

- Issue：[.github/ISSUE_TEMPLATE/](../.github/ISSUE_TEMPLATE/)
- PR：[.github/PULL_REQUEST_TEMPLATE.md](../.github/PULL_REQUEST_TEMPLATE.md)

填写 PR 时须勾选 **测试** 与 **文档** 相关项，见模板。

---

## 7. CODEOWNERS（可选）

团队扩大后，在根目录 `CODEOWNERS` 中为 `docs/`、`/.github/` 指定审查者，确保契约与流程文档变更有人审。

---

## 8. 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿 |
