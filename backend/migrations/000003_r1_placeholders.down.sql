DROP INDEX IF EXISTS idempotency_keys_expires_at_idx;
ALTER TABLE idempotency_keys DROP COLUMN IF EXISTS expires_at;
ALTER TABLE users DROP COLUMN IF EXISTS permissions;
