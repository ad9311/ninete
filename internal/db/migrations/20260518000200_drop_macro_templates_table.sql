-- +goose Up
INSERT OR IGNORE INTO "foods"
  ("user_id", "name", "kcal", "protein_g", "carbs_g", "fat_g",
   "fiber_g", "sodium_g", "saturated_fat_g", "created_at", "updated_at")
SELECT "user_id",
       "name",
       CASE WHEN "amount" > 0 THEN "kcal"      * 100.0 / "amount" ELSE "kcal"      END,
       CASE WHEN "amount" > 0 THEN "protein_g" * 100.0 / "amount" ELSE "protein_g" END,
       CASE WHEN "amount" > 0 THEN "carbs_g"   * 100.0 / "amount" ELSE "carbs_g"   END,
       CASE WHEN "amount" > 0 THEN "fat_g"     * 100.0 / "amount" ELSE "fat_g"     END,
       0,
       0,
       0,
       "created_at",
       "updated_at"
FROM "macro_templates";

DROP TABLE "macro_templates";

PRAGMA user_version = 24;

-- +goose Down
CREATE TABLE "macro_templates" (
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

PRAGMA user_version = 23;
