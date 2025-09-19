-- +goose Up
CREATE TABLE "refresh_tokens" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "token_hash" BLOB NOT NULL UNIQUE CHECK (length("token_hash") = 32),
  "issued_at" INTEGER NOT NULL,
  "expires_at" INTEGER NOT NULL
);

PRAGMA user_version = 2;

-- +goose Down
DROP TABLE IF EXISTS "refresh_tokens";

PRAGMA user_version = 1;
