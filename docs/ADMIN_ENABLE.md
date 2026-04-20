# Admin 管理页（`/admin/`）启用条件

以下**全部满足**时，ASE 会注册完整 Admin（登录、配置与索引 API），并输出日志：`admin UI enabled: GET /admin/`。

| 条件 | 环境变量 | 要求 |
|------|----------|------|
| 1. 用户名 | `ADMIN_USERNAME` | 非空 |
| 2. 密码（二选一） | `ADMIN_PASSWORD_BCRYPT` | **推荐**：bcrypt 哈希 |
|  | `ADMIN_PASSWORD` | 仅开发：明文；若已配置 `ADMIN_PASSWORD_BCRYPT` 则忽略此项 |
| 3. 会话密钥 | `ADMIN_SESSION_SECRET` | 非空且长度 **≥ 16** |

**可选**：`ADMIN_SESSION_TTL_SECONDS`（默认 `86400`）。

未满足时：仍会响应 **`/admin/`**，但返回 **503** 与说明页（不再 **404**）。

---

## Docker Compose 部署注意

仅把变量写在宿主机 **`.env`** 中**不够**：必须在 **`docker-compose.yml` 的 `ase.environment`** 中传入容器（本仓库已增加 `ADMIN_*` 的 `${VAR:-}` 注入）。

修改 `.env` 后执行：

```bash
docker compose up -d --build ase
```

---

## 检查是否已启用

```bash
curl -sS http://127.0.0.1:18080/api/info | jq .links.admin
```

若返回 `"/admin/"` 且访问 `/admin/` 为登录页，即已启用。

---

详见 [ADMIN_UI_DESIGN.md](./ADMIN_UI_DESIGN.md)。
