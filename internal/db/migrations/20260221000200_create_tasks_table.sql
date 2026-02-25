-- +goose Up
CREATE TABLE IF NOT EXISTS "tasks" (
  "id"          INTEGER PRIMARY KEY NOT NULL,
  "list_id"     INTEGER NOT NULL REFERENCES "lists"("id") ON DELETE CASCADE,
  "user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "description" TEXT NOT NULL,
  "priority"    INTEGER NOT NULL DEFAULT 1,
  "done"        INTEGER NOT NULL DEFAULT 0,
  "created_at"  INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at"  INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

PRAGMA user_version = 10;

-- +goose Down
DROP TABLE IF EXISTS "tasks";

PRAGMA user_version = 9;
