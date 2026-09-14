# ADR 0001 — CGA S0 技术栈

状态：已采纳（切片 0）

## 决策

单体 Go 服务，`APP_ROLE=api|worker|all`。HTTP 用 Gin；数据用 PostgreSQL + pgx + golang-migrate（纯 SQL）。日志 slog JSON 分渠道（app / access / audit）。追踪 OTel，endpoint 空则为 noop。校验 go-playground/validator。密码 argon2id，访问令牌 JWT 短 TTL。

## 不采用

- chi（主控指定 Gin）、SQLite、GORM、GraphQL、Kafka、Helm、微服务拆分
- 超过 4 层的全局 HTTP 中间件（国密开启时允许第 5 层，且仅挂 `/api/v1`）
- 全局 Auth / RBAC / 幂等 / 限流 / 审计 / 团队隔离中间件
- 用轮询冒充实时；SSE 推送敏感全文
- 把国密密钥写入数据库或仓库

## 后果

切片 0 只提供可运行框架与占位接口。未做领域返回空列表或 `COMMON_NOT_IMPLEMENTED`，禁止假装 CRUD。
