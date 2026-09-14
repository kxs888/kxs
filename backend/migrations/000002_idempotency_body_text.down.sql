ALTER TABLE idempotency_keys
    ALTER COLUMN response_body TYPE JSONB USING COALESCE(response_body::jsonb, 'null'::jsonb);
