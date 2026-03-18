-- +goose Up
ALTER TABLE "macro_entries" ADD COLUMN "meal_type" TEXT NOT NULL DEFAULT 'other';
PRAGMA user_version = 14;

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
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
INSERT INTO "macro_entries_new" SELECT "id","user_id","name","kcal","protein_g","carbs_g","fat_g","date","created_at","updated_at" FROM "macro_entries";
DROP TABLE "macro_entries";
ALTER TABLE "macro_entries_new" RENAME TO "macro_entries";
PRAGMA user_version = 13;
