-- +goose Up
-- A static monthly budget amount per category, in cents to match
-- "expenses"."amount". Nothing time-derived is stored: the comparison against a
-- date range is computed per render.
CREATE TABLE IF NOT EXISTS "expense_budgets" (
  "id" INTEGER PRIMARY KEY NOT NULL,
  "user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
  "category_id" INTEGER NOT NULL REFERENCES "categories"("id") ON DELETE CASCADE,
  "amount" INTEGER NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

-- Backs the ON CONFLICT upsert and doubles as the per-user lookup index.
CREATE UNIQUE INDEX IF NOT EXISTS "idx_expense_budgets_user_category"
ON "expense_budgets" ("user_id", "category_id");

PRAGMA user_version = 29;

-- +goose Down
DROP TABLE IF EXISTS "expense_budgets";

PRAGMA user_version = 28;
