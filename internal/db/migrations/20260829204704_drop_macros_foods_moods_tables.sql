-- +goose Up
-- The only irreversible step of Phase 8 (docs/spa-migration.md §5). Macros,
-- foods and moods were dropped from the app in Phase 0B; these tables and the
-- old mood taggings survived only so the data was not lost out from under
-- anyone before this migration was written. Safe now: the owner confirmed the
-- data is no longer needed.
DELETE FROM "taggings" WHERE "taggable_type" = 'mood_entry';
DROP TABLE "foods";
DROP TABLE "macro_entries";
DROP TABLE "macro_goals";
DROP TABLE "mood_entries";

PRAGMA user_version = 31;

-- +goose Down
-- Recreates the tables empty. This is not a restore: the data dropped above
-- does not come back.
CREATE TABLE "foods" (
  "id"              INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"         INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"            TEXT    NOT NULL DEFAULT '',
  "kcal"            REAL    NOT NULL DEFAULT 0,
  "protein_g"       REAL    NOT NULL DEFAULT 0,
  "carbs_g"         REAL    NOT NULL DEFAULT 0,
  "fat_g"           REAL    NOT NULL DEFAULT 0,
  "created_at"      INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at"      INTEGER NOT NULL DEFAULT (unixepoch()),
  "fiber_g"         REAL NOT NULL DEFAULT 0,
  "sodium_g"        REAL NOT NULL DEFAULT 0,
  "saturated_fat_g" REAL NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX "uq_foods_user_lower_name" ON "foods" ("user_id", lower("name"));

CREATE TABLE "macro_entries" (
  "id"              INTEGER PRIMARY KEY NOT NULL,
  "user_id"         INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"            TEXT NOT NULL,
  "kcal"            REAL NOT NULL DEFAULT 0,
  "protein_g"       REAL NOT NULL DEFAULT 0,
  "carbs_g"         REAL NOT NULL DEFAULT 0,
  "fat_g"           REAL NOT NULL DEFAULT 0,
  "date"            INTEGER NOT NULL,
  "created_at"      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at"      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "meal_type"       TEXT NOT NULL DEFAULT 'other',
  "fiber_g"         REAL NOT NULL DEFAULT 0,
  "sodium_g"        REAL NOT NULL DEFAULT 0,
  "saturated_fat_g" REAL NOT NULL DEFAULT 0
);
CREATE INDEX "idx_macro_entries_user_date" ON "macro_entries" ("user_id", "date");

CREATE TABLE "macro_goals" (
  "id"              INTEGER PRIMARY KEY NOT NULL,
  "user_id"         INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "kcal"            REAL NOT NULL DEFAULT 0,
  "protein_g"       REAL NOT NULL DEFAULT 0,
  "carbs_g"         REAL NOT NULL DEFAULT 0,
  "fat_g"           REAL NOT NULL DEFAULT 0,
  "created_at"      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at"      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "fiber_g"         REAL NOT NULL DEFAULT 0,
  "sodium_g"        REAL NOT NULL DEFAULT 0,
  "saturated_fat_g" REAL NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX "index_macro_goals_on_user_id" ON "macro_goals" ("user_id");

CREATE TABLE "mood_entries" (
  "id"         INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "mood"       TEXT    NOT NULL DEFAULT '',
  "notes"      TEXT    NOT NULL DEFAULT '',
  "logged_at"  INTEGER NOT NULL DEFAULT (unixepoch()),
  "created_at" INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at" INTEGER NOT NULL DEFAULT (unixepoch())
);
CREATE INDEX "idx_mood_entries_user_logged_at" ON "mood_entries" ("user_id", "logged_at");

PRAGMA user_version = 30;
