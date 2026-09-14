# HTTP 包装白名单（S0）

全局中间件**最多 4 层**；国密开启时允许第 5 层。顺序固定：

| 层 | 名称 | 范围 |
| --- | --- | --- |
| 1 | Recover | 全局 |
| 2 | RequestID + Trace | 全局 |
| 3 | Timeout + BodyLimit + 安全头 / CORS | 全局（合并为**一个**包装，不算 3 层） |
| 4 | AccessLog | 全局；禁止记录 JWT / Authorization / 请求体 |
| 5 | 可选国密 SM4 | **仅** `SM_CRYPTO_ENABLED=true`，挂在 `/api/v1`。`/healthz` `/readyz` `/metrics` `/api/v1/stream` **永不加密** |

## 禁止做成全局中间件

以下必须按**路由局部** `Group().Use(...)` 或在 **service 显式调用**：

- Auth / RBAC
- 幂等（`Idempotency-Key`，示例仅 `POST /api/v1/ping-writes`）
- 限流
- 审计（`audit.Record`，禁止「所有 POST 自动记」）
- 团队隔离

本文件是审查合并版框架提示词在本仓的落地约束。不要再拼贴已废弃的 SKELETON / EXTRA / REALTIME 提示词。
