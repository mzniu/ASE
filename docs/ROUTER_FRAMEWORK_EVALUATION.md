# HTTP 路由框架评估：chi、gin、echo

面向 **`POST /v1/search`** 为主的同步 API，在 **Go 1.22+** 下对三个常见路由库做对比，便于记录决策依据；**当前仓库已选用 chi v5**（见 [ARCHITECTURE.md](./ARCHITECTURE.md) §2.4）。

| 维度 | **chi** | **gin** | **echo** |
|------|---------|---------|----------|
| 与 `net/http` 关系 | **原生**：`http.Handler` / `http.HandlerFunc`，路由即中间件链 | 自定义 **`gin.Context`**，与标准 `ResponseWriter` 有封装 | 自定义 **`echo.Context`**，类似 gin |
| 学习曲线 | 低；与官方 `ServeMux` 思维接近 | 中；需熟悉 gin 的 binding、错误写法 | 中 |
| 中间件与可测性 | 标准 `httptest` + 直接测 `Handler`，**TDD 友好** | 常用 `httptest` + `gin.CreateTestContext` 或跑全引擎 | 介于两者之间 |
| 生态与中文资料 | 足够；略少于 gin | **国内资料与示例最多** | 中等 |
| 性能 | 足够本服务场景；差异多在微优化 | 略宣传「快」，对本项目瓶颈（OpenSearch/外网）通常 **非主矛盾** | 与 gin 同量级感知 |
| 依赖与锁定 | 轻量、接口稳定 | 功能多、依赖面略大 | 中等 |
| 风险 | 大型团队若强依赖 gin 生态，选 chi 会有迁移成本 | API 与 std 有距离，长期略偏「框架绑定」 | 社区体量小于 gin |

---

## 结论与推荐

1. **本仓库采用：chi v5**  
   - **理由**：与 **`net/http` 一致**，handler 易做 **纯函数式测试**；中间件模型清晰，适合 **鉴权、限流、RequestID** 等与 SRS 对齐的横切能力；对「单端点 + 少量中间件」的 MVP **无冗余抽象**。

2. **若团队已有强 gin 经验、且希望大量复用国内示例代码**：可选 **gin**，成本主要在测试与 `Context` 适配，功能上完全可行。

3. **echo**：适合喜欢其 API 风格的团队；与 gin 相比无压倒性优势，可作为第三选项。

---

## 迁移提示（若日后从 chi 换 gin/echo）

- 保持 **`internal/handler`** 内业务与路由注册分离**：将「解析 JSON、鉴权、写 Problem Details」留在可单测的函数中，路由层仅负责挂载路径。  
- 更换框架时 **只改 `cmd/server` 与中间件装配**，`Search` 核心逻辑应尽量只依赖 `http.ResponseWriter` + `*http.Request`。

---

## 修订记录

| 版本 | 日期 | 说明 |
|------|------|------|
| 0.1 | 2026-04-18 | 初稿；决议采用 **chi v5** |
