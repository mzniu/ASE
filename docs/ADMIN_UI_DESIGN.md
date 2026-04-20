# Admin 管理界面方案设计（外网 18080）

## 1. 目标与边界

| 目标 | 说明 |
|------|------|
| **访问入口** | 与 ASE 服务**同一对外端口**（如 **`18080`**），通过浏览器访问管理功能。 |
| **鉴权** | **必须先登录**才能进入管理区；与面向 Agent 的 **`Bearer` API Key** 区分（可共用密钥体系或独立管理员账号）。 |
| **配置可见** | 展示**当前生效的运行时配置摘要**（脱敏：不展示 `TAVILY_*`、`AUTH_*` 等敏感值，仅「已配置 / 未配置」或掩码）。 |
| **索引可见** | 展示 OpenSearch 中**索引列表、文档量、mapping 只读预览**；**不**把 OpenSearch 原生端口（9200）暴露到公网。 |

| 非目标（首期可不做） | 说明 |
|----------------------|------|
| 替代 OpenSearch Dashboards 的全部能力 | 首期以**只读巡检**为主；复杂查询可用后续嵌入或跳转（内网）。 |
| 在浏览器直连 OpenSearch | **禁止**；所有 OS 访问经 ASE **服务端代理**并带权限校验。 |
| 修改索引 mapping / 删除索引 | 若要做，需单独评审（误操作风险高），建议二期 + 二次确认。 |

---

## 2. 安全原则（外网必守）

1. **OpenSearch 不对公网暴露**：当前 `docker-compose` 中 OS 为**开发向**（如 `plugins.security.disabled`），9200 仅应内网/本机可达。
2. **管理面与业务面同端口时**：攻击面叠加，建议：
   - **路径隔离**：如仅 `/admin/*` 为管理静态资源与 API；
   - **生产**优先：**IP 允许列表**、**VPN/Zero Trust**、或在**前置负载均衡**上对 `/admin` 做独立访问策略；
   - **TLS**：公网入口建议由云 LB / Nginx 终结 HTTPS，后端可仍 HTTP（内网）。
3. **密钥**：管理员口令或 Session 密钥使用**环境变量**注入，**不入库、不写仓库**。

---

## 3. 架构选项对比

### 方案 A（推荐）：在 ASE 进程内扩展路由（同端口 18080）

- **做法**：在现有 Go HTTP 服务上增加：
  - `GET /admin`（或 `/admin/`）→ 嵌入或托管单页应用（SPA）静态资源；
  - `POST /admin/api/login` → 签发 **HttpOnly Cookie** 的 Session 或 **短期 JWT**；
  - `GET /admin/api/...` → 返回脱敏配置、代理调用 OpenSearch 只读 API（使用现有 `opensearch-go` 或新增轻量 HTTP 客户端）。
- **优点**：单一二进制、无额外容器、端口天然就是 18080；与现有 `chi` 路由一致。
- **缺点**：Go 服务职责增加，需规范静态资源嵌入与前端构建流程。

### 方案 B：独立 `admin` 侧车容器 + 反向代理统一到 18080

- **做法**：新增轻量 Admin 前端/后端容器，前面加 **Nginx/Caddy**：`location / { proxy ASE }`，`location /admin { proxy admin }`。
- **优点**：前后端技术栈自由；ASE 仓库可保持「纯 API」。
- **缺点**：多组件、Compose 复杂；需维护代理与证书。

### 方案 C：仅内网 OpenSearch Dashboards，外网只做「跳转说明」

- **做法**：外网 18080 仅提供**文档链接**与**健康状态**；Dashboards 仅 VPN 访问。
- **优点**：实现成本最低、暴露面最小。
- **缺点**：不满足「外网 18080 直接打开管理界面」的强需求。

**结论**：若必须**外网同端口**，优先 **方案 A**；若团队希望前后端分离迭代快，可选 **方案 B**。

---

## 4. 推荐落地形态（方案 A 细化）

### 4.1 URL 规划

| 路径 | 用途 |
|------|------|
| `/` | 现有项目主页（不变） |
| `/v1/*` | 现有 API（不变） |
| `/admin/` | Admin SPA（登录页 + 控制台） |
| `/admin/api/login` | 登录（JSON body：用户名/密码或单字段 `admin_token`） |
| `/admin/api/logout` | 登出 |
| `/admin/api/session` | 当前会话是否有效 |
| `/admin/api/config` | 脱敏配置快照（JSON） |
| `/admin/api/indices` | 索引列表（代理 `_cat/indices` 或 `_cluster/state` 精简） |
| `/admin/api/indices/{name}` | 单个索引 `_count`、`_mapping`（只读） |

静态资源通过 `embed` 或 `http.FileServer` 挂载在 `/admin/` 下，**前端路由**用 `history` 模式时需服务端对 `/admin/*` 回退到 `index.html`（常见 SPA 约定）。

### 4.2 鉴权设计

| 方式 | 说明 |
|------|------|
| **环境变量** | 例如 `ADMIN_USERNAME` + `ADMIN_PASSWORD_HASH`（bcrypt）或 `ADMIN_TOKEN`（随机长串，仅比对）。 |
| **Session** | 登录成功后 Set-Cookie：`Secure; HttpOnly; SameSite=Lax`；服务端 Session 存内存或 Redis（多副本时需 Redis）。 |
| **与 API Key 关系** | **`/v1/*` 仍用 `Authorization: Bearer`**；**/admin** 用 Cookie 或独立 `X-Admin-Token` header**，避免与普通调用方密钥混用。 |

未配置管理员凭证时：**不注册 `/admin` 路由**或返回 **503 + 说明**（避免默认弱口令上线）。

### 4.3「配置」页展示内容（建议）

- 监听地址、`SEARCH_DEFAULT_PROVIDERS`、各 `*_BROWSER_ENABLED`、`PROVIDER_FETCH_RESULT_URLS`、OpenSearch「已连接 / 索引名」、Tavily「已配置 API Key：是/否」（不显示值）。
- **禁止**：完整 `DEV_API_KEY`、`AUTH_VALID_API_KEYS`、`TAVILY_API_KEY` 原文。

### 4.4「索引」页展示内容（建议）

- 调用 OpenSearch：`GET /_cat/indices?format=json` 或 `GET /_cluster/health` + 索引列表。
- 每索引：`docs.count`、store size、`health`。
- 详情：`/_mapping` 只读、`/_count`、`/_search?size=0` 聚合可选（二期）。

---

## 5. 实现阶段建议

| 阶段 | 内容 |
|------|------|
| **MVP** | 环境变量开关管理员凭证；`/admin` 极简页面（或 Server 渲染表格）；登录后展示脱敏配置 + 索引列表；所有 OS 请求服务端代理。 |
| **二期** | 索引内样例文档只读预览、mapping 折叠展示、基础集群健康图；审计日志（谁何时登录）。 |
| **三期** | 可选 OIDC、只读/运维角色分离；若需改索引设置，需强二次确认与审计。 |

---

## 6. 与现有 ASE 代码的关系

- **配置读取**：复用 `internal/config` 的 `Config`，增加 `Admin*` 字段与脱敏序列化函数。
- **OpenSearch**：复用或扩展 `internal/adapter/opensearch`，新增**只读**管理用方法（禁止把原始 `OPENSEARCH_URLS` 返回给浏览器）。
- **路由**：`cmd/server/main.go` 中 `r.Route("/admin", ...)`，并对 `/admin/api/*` 套 **中间件** 校验 Session。

---

## 7. 部署与运维注意

- **防火墙**：仅开放 **18080** 时，确认云安全组未误开放 **9200**。
- **变更管理员密码**：改环境变量 + 滚动重启 ASE 容器即可。
- **压力**：管理 API 应**单独限流**（避免扫库）；索引列表请求频率限制。

---

## 8. 文档维护

本文件为**方案设计**；实现后应在 `README.md` 或 `DEPLOY_LINUX_VM.md` 中增加「管理员界面」小节（访问路径、环境变量、安全提示）。
