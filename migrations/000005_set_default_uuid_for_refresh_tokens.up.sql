CREATE EXTENSION IF NOT EXISTS pgcrypto;

ALTER TABLE refresh_tokens
  ALTER COLUMN uuid SET DEFAULT gen_random_uuid();
