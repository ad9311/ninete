-- +goose Up
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
INSERT INTO "macro_entries_new" SELECT * FROM "macro_entries";
DROP TABLE "macro_entries";
ALTER TABLE "macro_entries_new" RENAME TO "macro_entries";

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
INSERT INTO "macro_goals_new" SELECT * FROM "macro_goals";
DROP TABLE "macro_goals";
ALTER TABLE "macro_goals_new" RENAME TO "macro_goals";
CREATE UNIQUE INDEX "index_macro_goals_on_user_id" ON "macro_goals" ("user_id");

PRAGMA user_version = 13;

-- +goose Down
CREATE TABLE "macro_entries_old" (
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
INSERT INTO "macro_entries_old" SELECT * FROM "macro_entries";
DROP TABLE "macro_entries";
ALTER TABLE "macro_entries_old" RENAME TO "macro_entries";

CREATE TABLE "macro_goals_old" (
  "id"         INTEGER PRIMARY KEY NOT NULL,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "kcal"       INTEGER NOT NULL DEFAULT 0,
  "protein_g"  INTEGER NOT NULL DEFAULT 0,
  "carbs_g"    INTEGER NOT NULL DEFAULT 0,
  "fat_g"      INTEGER NOT NULL DEFAULT 0,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);
INSERT INTO "macro_goals_old" SELECT * FROM "macro_goals";
DROP TABLE "macro_goals";
ALTER TABLE "macro_goals_old" RENAME TO "macro_goals";
CREATE UNIQUE INDEX "index_macro_goals_on_user_id" ON "macro_goals" ("user_id");

PRAGMA user_version = 12;
