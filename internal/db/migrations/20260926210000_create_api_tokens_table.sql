-- +goose Up
-- Personal access tokens: the credential a client other than the browser (the
-- local MCP server; docs/architecture.md, "Bearer tokens") sends as
-- "Authorization: Bearer". Only the SHA-256 of the token is stored — the
-- plaintext is shown once, at creation.
-- "prefix" is the token's first characters, kept so the settings page can tell
-- tokens apart without holding anything that authenticates.
--
-- "expires_at" is NULL for a token that never expires. "revoked_at" is set
-- rather than the row deleted, so a revoked token keeps answering 401 through
-- the same lookup instead of vanishing into "no such token".
CREATE TABLE IF NOT EXISTS "api_tokens" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name" TEXT NOT NULL,
  "token_hash" TEXT NOT NULL,
  "prefix" TEXT NOT NULL,
  "scope" TEXT NOT NULL CHECK ("scope" IN ('read', 'write')),
  "expires_at" INTEGER,
  "last_used_at" INTEGER,
  "revoked_at" INTEGER,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_api_tokens_token_hash"
ON "api_tokens" ("token_hash");

PRAGMA user_version = 35;

-- +goose Down
DROP INDEX IF EXISTS "uq_api_tokens_token_hash";
DROP TABLE IF EXISTS "api_tokens";

PRAGMA user_version = 34;
