-- +goose Up
CREATE TABLE IF NOT EXISTS "macro_entries" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT NOT NULL,
  "kcal"       INTEGER NOT NULL DEFAULT 0,
  "protein_g"  INTEGER NOT NULL DEFAULT 0,
  "carbs_g"    INTEGER NOT NULL DEFAULT 0,
  "fat_g"      INTEGER NOT NULL DEFAULT 0,
  "date"       INTEGER NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
PRAGMA user_version = 11;

-- +goose Down
DROP TABLE IF EXISTS "macro_entries";
PRAGMA user_version = 10;
