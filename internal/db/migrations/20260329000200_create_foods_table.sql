-- +goose Up
CREATE TABLE "foods" (
  "id"         INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT    NOT NULL DEFAULT '',
  "kcal"       REAL    NOT NULL DEFAULT 0,
  "protein_g"  REAL    NOT NULL DEFAULT 0,
  "carbs_g"    REAL    NOT NULL DEFAULT 0,
  "fat_g"      REAL    NOT NULL DEFAULT 0,
  "created_at" INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at" INTEGER NOT NULL DEFAULT (unixepoch())
);

PRAGMA user_version = 18;

-- +goose Down
DROP TABLE IF EXISTS "foods";

PRAGMA user_version = 17;
