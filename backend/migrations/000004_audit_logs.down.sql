DROP INDEX IF EXISTS audit_logs_trace_id_idx;
DROP INDEX IF EXISTS audit_logs_request_id_idx;
DROP INDEX IF EXISTS audit_logs_action_idx;

ALTER TABLE audit_logs DROP COLUMN IF EXISTS trace_id;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS user_agent;
ALTER TABLE audit_logs DROP COLUMN IF EXISTS outcome;

-- 表本体由 000001 创建；完整 DROP TABLE 见 000001 down。
