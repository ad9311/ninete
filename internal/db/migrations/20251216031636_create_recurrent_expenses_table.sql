-- +goose Up
CREATE TABLE IF NOT EXISTS "recurrent_expenses" (
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

PRAGMA user_version = 5;

-- +goose Down
DROP TABLE IF EXISTS "recurrent_expenses";

PRAGMA user_version = 4;
