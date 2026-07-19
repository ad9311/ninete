-- +goose Up
CREATE TABLE IF NOT EXISTS "expense_category_mappings" (
  "id" INTEGER PRIMARY KEY NOT NULL,
	"user_id" INTEGER NOT NULL REFERENCES "users"("id") ON DELETE CASCADE,
	"category_id" INTEGER NOT NULL REFERENCES "categories"("id") ON DELETE CASCADE,
	"description_key" TEXT NOT NULL,
  "created_at" INTEGER NOT NULL DEFAULT (strftime('%s','now')),
  "updated_at" INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE UNIQUE INDEX "idx_expense_category_mappings_user_desc"
  ON "expense_category_mappings" ("user_id", "description_key");

PRAGMA user_version = 25;

-- +goose Down
DROP TABLE IF EXISTS "expense_category_mappings";

PRAGMA user_version = 24;
