-- +goose Up
ALTER TABLE "macro_entries" ADD COLUMN "fiber_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "macro_entries" ADD COLUMN "sodium_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "macro_entries" ADD COLUMN "saturated_fat_g" REAL NOT NULL DEFAULT 0;

ALTER TABLE "foods" ADD COLUMN "fiber_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "foods" ADD COLUMN "sodium_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "foods" ADD COLUMN "saturated_fat_g" REAL NOT NULL DEFAULT 0;

PRAGMA user_version = 22;

-- +goose Down
CREATE TABLE "macro_entries_new" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT NOT NULL,
  "kcal"       REAL NOT NULL DEFAULT 0,
  "protein_g"  REAL NOT NULL DEFAULT 0,
  "carbs_g"    REAL NOT NULL DEFAULT 0,
  "fat_g"      REAL NOT NULL DEFAULT 0,
  "date"       INTEGER NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "meal_type"  TEXT NOT NULL DEFAULT 'other'
);
INSERT INTO "macro_entries_new" SELECT "id","user_id","name","kcal","protein_g","carbs_g","fat_g","date","created_at","updated_at","meal_type" FROM "macro_entries";
DROP TABLE "macro_entries";
ALTER TABLE "macro_entries_new" RENAME TO "macro_entries";

CREATE TABLE "foods_new" (
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
INSERT INTO "foods_new" SELECT "id","user_id","name","kcal","protein_g","carbs_g","fat_g","created_at","updated_at" FROM "foods";
DROP TABLE "foods";
ALTER TABLE "foods_new" RENAME TO "foods";
CREATE UNIQUE INDEX IF NOT EXISTS "uq_foods_user_lower_name" ON "foods" ("user_id", lower("name"));

PRAGMA user_version = 21;
