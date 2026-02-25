-- +goose Up
CREATE TABLE IF NOT EXISTS "lists" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_lists_user_lower_name"
ON "lists" ("user_id", lower("name"));

PRAGMA user_version = 9;

-- +goose Down
DROP INDEX IF EXISTS "uq_lists_user_lower_name";
DROP TABLE IF EXISTS "lists";

PRAGMA user_version = 8;
