-- 幂等回放必须原样保存响应字节，不能走 jsonb 规范化。
ALTER TABLE idempotency_keys
    ALTER COLUMN response_body TYPE TEXT USING COALESCE(response_body::text, '');
