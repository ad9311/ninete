-- +goose Up
CREATE TABLE IF NOT EXISTS "macro_goals" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "kcal"       INTEGER NOT NULL DEFAULT 0,
  "protein_g"  INTEGER NOT NULL DEFAULT 0,
  "carbs_g"    INTEGER NOT NULL DEFAULT 0,
  "fat_g"      INTEGER NOT NULL DEFAULT 0,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
CREATE UNIQUE INDEX "index_macro_goals_on_user_id" ON "macro_goals" ("user_id");
PRAGMA user_version = 12;

-- +goose Down
DROP TABLE IF EXISTS "macro_goals";
PRAGMA user_version = 11;
