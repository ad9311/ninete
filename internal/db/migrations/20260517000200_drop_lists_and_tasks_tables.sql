-- +goose Up
DROP TABLE IF EXISTS "tasks";
DROP INDEX IF EXISTS "uq_lists_user_lower_name";
DROP TABLE IF EXISTS "lists";

PRAGMA user_version = 21;

-- +goose Down
CREATE TABLE IF NOT EXISTS "lists" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX IF NOT EXISTS "uq_lists_user_lower_name"
ON "lists" ("user_id", lower("name"));

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

PRAGMA user_version = 20;
