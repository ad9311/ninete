-- +goose Up
CREATE TABLE IF NOT EXISTS "tags" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_tags_user_lower_name"
ON "tags" ("user_id", lower("name"));

PRAGMA user_version = 5;

-- +goose Down
DROP INDEX IF EXISTS "uq_tags_user_lower_name";
DROP TABLE IF EXISTS "tags";

PRAGMA user_version = 4;
