-- +goose Up
CREATE TABLE "macro_templates_new" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"     INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"        TEXT    NOT NULL DEFAULT '',
  "kcal"        REAL    NOT NULL DEFAULT 0,
  "protein_g"   REAL    NOT NULL DEFAULT 0,
  "carbs_g"     REAL    NOT NULL DEFAULT 0,
  "fat_g"       REAL    NOT NULL DEFAULT 0,
  "amount"      REAL    NOT NULL DEFAULT 0,
  "amount_unit" TEXT    NOT NULL DEFAULT 'g',
  "created_at"  INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at"  INTEGER NOT NULL DEFAULT (unixepoch())
);
INSERT INTO "macro_templates_new" ("id", "user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "amount", "amount_unit", "created_at", "updated_at")
SELECT "id", "user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "amount_g", 'g', "created_at", "updated_at"
FROM "macro_templates";
DROP TABLE "macro_templates";
ALTER TABLE "macro_templates_new" RENAME TO "macro_templates";
PRAGMA user_version = 17;

-- +goose Down
CREATE TABLE "macro_templates_old" (
  "id"         INTEGER PRIMARY KEY AUTOINCREMENT,
  "user_id"    INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "name"       TEXT    NOT NULL DEFAULT '',
  "kcal"       REAL    NOT NULL DEFAULT 0,
  "protein_g"  REAL    NOT NULL DEFAULT 0,
  "carbs_g"    REAL    NOT NULL DEFAULT 0,
  "fat_g"      REAL    NOT NULL DEFAULT 0,
  "amount_g"   REAL    NOT NULL DEFAULT 0,
  "created_at" INTEGER NOT NULL DEFAULT (unixepoch()),
  "updated_at" INTEGER NOT NULL DEFAULT (unixepoch())
);
INSERT INTO "macro_templates_old" ("id", "user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "amount_g", "created_at", "updated_at")
SELECT "id", "user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g", "amount", "created_at", "updated_at"
FROM "macro_templates";
DROP TABLE "macro_templates";
ALTER TABLE "macro_templates_old" RENAME TO "macro_templates";
PRAGMA user_version = 16;
