-- +goose Up
CREATE TABLE "mood_entries" (
  "id"         INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "mood"       TEXT    NOT NULL DEFAULT '',
  "notes"      TEXT    NOT NULL DEFAULT '',
  "logged_at"  INTEGER NOT NULL DEFAULT (unixepoch()),
  "created_at" INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at" INTEGER NOT NULL DEFAULT (unixepoch())
);

PRAGMA user_version = 19;

-- +goose Down
DROP TABLE IF EXISTS "mood_entries";

PRAGMA user_version = 18;
