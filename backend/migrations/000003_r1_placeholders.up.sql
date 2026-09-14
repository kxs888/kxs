-- B7：种子用户权限占位（无业务 CRUD 实现）。
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS permissions TEXT[] NOT NULL DEFAULT ARRAY['patient.view', 'task.create', 'report.view']::text[];

-- C6：幂等键 expires_at 默认 24h；过期后同 key 可重新占用，未过期异 body 不得覆盖。
ALTER TABLE idempotency_keys
    ADD COLUMN IF NOT EXISTS expires_at TIMESTAMPTZ NOT NULL DEFAULT now() + interval '24 hours';

CREATE INDEX IF NOT EXISTS idempotency_keys_expires_at_idx ON idempotency_keys (expires_at);
