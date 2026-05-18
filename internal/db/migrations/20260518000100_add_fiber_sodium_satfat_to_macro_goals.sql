-- +goose Up
ALTER TABLE "macro_goals" ADD COLUMN "fiber_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "macro_goals" ADD COLUMN "sodium_g" REAL NOT NULL DEFAULT 0;
ALTER TABLE "macro_goals" ADD COLUMN "saturated_fat_g" REAL NOT NULL DEFAULT 0;

PRAGMA user_version = 23;

-- +goose Down
CREATE TABLE "macro_goals_new" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "kcal"       REAL NOT NULL DEFAULT 0,
  "protein_g"  REAL NOT NULL DEFAULT 0,
  "carbs_g"    REAL NOT NULL DEFAULT 0,
  "fat_g"      REAL NOT NULL DEFAULT 0,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
INSERT INTO "macro_goals_new" SELECT "id","user_id","kcal","protein_g","carbs_g","fat_g","created_at","updated_at" FROM "macro_goals";
DROP TABLE "macro_goals";
ALTER TABLE "macro_goals_new" RENAME TO "macro_goals";
CREATE UNIQUE INDEX IF NOT EXISTS "index_macro_goals_on_user_id" ON "macro_goals" ("user_id");

PRAGMA user_version = 22;
