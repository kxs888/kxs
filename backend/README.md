# CGA 后端（切片 0）

独立 Go module。本切片只做**框架可运行**，不做完整 CGA CRUD。

统一提示词落地：`docs/http-wrappers.md` + `docs/adr/0001-stack.md`（审查合并版）。不要再拼贴已废弃的 SKELETON / EXTRA / REALTIME。

## 如何 make run

```bash
cd backend
cp .env.example .env   # 按需修改
make run               # docker compose 起 Postgres + API；无 Docker 时用本地 DATABASE_URL
```

默认 `http://localhost:8080`。演示用户：`BOOTSTRAP_USERNAME` / `BOOTSTRAP_PASSWORD`（见 `.env.example`）。

```bash
curl -s localhost:8080/healthz
curl -s localhost:8080/readyz
TOKEN=$(curl -s -X POST localhost:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"demo","password":"demo-pass-change-me"}' | jq -r .data.access_token)
curl -s localhost:8080/api/v1/me -H "Authorization: Bearer $TOKEN"
curl -s -X POST localhost:8080/api/v1/ping-writes \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-key-001' \
  -d '{"message":"hello-s0"}'
```

`make test` 运行 `go test ./...`。`SM_CRYPTO_ENABLED` 默认 `false`，密钥只走环境变量。

成功信封含 `meta.request_id` / `meta.trace_id`；失败另有 `error.trace_id`。登录成功/失败与 ping-writes 成功均由 service 显式 `audit.Record`（detail 不含病历/JWT）。

## S0 验收 MUST

| 项 | 要求 |
| --- | --- |
| **A** 探活 | `GET /healthz` 200 信封；`GET /readyz` ping DB，失败 `COMMON_UNAVAILABLE` |
| **B** 认证占位 | `POST /api/v1/auth/login` argon2id + JWT access 短 TTL；`GET /api/v1/me`；Auth **非**全局中间件 |
| **C** 信封 | `respond.OK` / `respond.Err`；`internal/errcode` 的 `COMMON_` / `AUTH_` / … |
| **D** 幂等与审计 | `POST /api/v1/ping-writes` + `Idempotency-Key` 双提交回放且只落一行；`audit.Record` 显式调用，禁止全局「所有 POST 自动记」 |
| **F1** 国密可选 | 默认关可跑通；开时第 5 层仅 `/api/v1`；healthz/readyz/metrics/stream 永不加密；单测开/关 |

未做领域：`GET /api/v1/tasks` 空列表。无患者主档/评估/知情同意/离线/HIS 接口。

OpenAPI：`api/openapi.yaml`。
