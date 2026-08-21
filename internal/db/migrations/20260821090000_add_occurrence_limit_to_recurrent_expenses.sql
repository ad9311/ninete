-- +goose Up
-- "occurrence_limit" caps how many expenses a recurrent expense may generate.
-- 0 means unlimited, which is what every existing row keeps. Once
-- "occurrence_count" reaches the limit the row is archived and the cron job
-- stops copying it until the owner unarchives it by hand.
ALTER TABLE "recurrent_expenses" ADD COLUMN "occurrence_limit" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "recurrent_expenses" ADD COLUMN "occurrence_count" INTEGER NOT NULL DEFAULT 0;
ALTER TABLE "recurrent_expenses" ADD COLUMN "archived_at" INTEGER;

PRAGMA user_version = 30;

-- +goose Down
CREATE TABLE "recurrent_expenses_new" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "category_id" INTEGER NOT NULL REFERENCES "categories"("id") ON DELETE CASCADE,
  "description" TEXT NOT NULL,
  "amount" INTEGER NOT NULL,
  "period" INTEGER NOT NULL DEFAULT 1,
  "last_copy_created_at" INTEGER,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);
INSERT INTO "recurrent_expenses_new" SELECT "id","user_id","category_id","description","amount","period","last_copy_created_at","created_at","updated_at" FROM "recurrent_expenses";
DROP TABLE "recurrent_expenses";
ALTER TABLE "recurrent_expenses_new" RENAME TO "recurrent_expenses";
CREATE INDEX IF NOT EXISTS "idx_recurrent_expenses_user_id" ON "recurrent_expenses" ("user_id");
CREATE INDEX IF NOT EXISTS "idx_recurrent_expenses_category_id" ON "recurrent_expenses" ("category_id");

PRAGMA user_version = 29;
